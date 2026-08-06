package shell

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// discovery.go:抵達星系時的一次性發現(原版 `Do_System_Discoveries_At_Star_` @ 0xE9927)。
//
// 原版把五種特殊物產做成「艦隊第一次進到那個星系」就結算的發現事件,而不是殖民之後才生效:
//
//	太空殘骸   → 國庫 +50 BC
//	海盜藏寶   → 國庫 +100 BC
//	失散殖民地 → 就地生出一個殖民地,人口 = min(該行星人口上限, 3)
//	受困英雄   → 免費得到一名領袖
//	遠古文物   → 白送一項「現在就能研究」的科技
//
// 數字的來源見 gamedata/planet_special.go 的逐條註解(手冊只給定性描述,金額與人口上限
// 是從反組譯的指令讀出來的)。
//
// remake 的落差與取捨:
//   - 原版的觸發旗標存在 `Star.special`(結算後覆寫成訊息碼),remake 沒有 Star.special 這個
//     欄位,改用 Planet.SpecialSeen 記「已結算過」——效果一樣是不重複觸發,存法是 remake 的。
//   - 原版失散殖民地的人口上限用 `Colony_Race_Pop_Limit_`(含種族/科技修正);remake 用
//     gamedata.PlanetBasePopMax(既有的「大小×氣候」推導),再與 3 取小。
//   - 原版遠古文物送的是「狀態 == RSTATE_READY」的研究主題;remake 的研究狀態模型只有
//     CompletedTopics,故改送 researchQueue() 裡第一個未完成的主題(仍是「一項、且是現在
//     真的排得到的」,選法是 remake 的)。

// SystemDiscovery 是一次星系發現的結果(供回合摘要/報告畫面顯示)。
type SystemDiscovery struct {
	StarIndex int                  // 觸發的星索引
	StarName  string               // 星名(顯示用)
	Special   gamedata.PlanetSpecial // 觸發的特殊物產
	Name      string               // 特殊物產中文名
	Message   string               // 已填好數字/名稱的敘述
	BCGained  int                  // 一次性入袋的 BC(太空殘骸/海盜藏寶)
	ColonyIdx int                  // 失散殖民地建成後的殖民地索引;-1 = 無
	LeaderGot string               // 免費領袖的名字;空 = 無
	TechGot   string               // 白送的科技主題名;空 = 無
}

// discoverSystemSpecials 結算玩家艦隊剛抵達的這顆星的一次性發現。
// 沒有可觸發的特殊物產(或已經結算過)回 nil,呼叫端不需要顯示任何東西。
func (s *GameSession) discoverSystemSpecials(starIdx int) *SystemDiscovery {
	if starIdx < 0 || starIdx >= len(s.Planets) || starIdx >= len(s.Stars) {
		return nil
	}
	p := &s.Planets[starIdx]
	if p.SpecialSeen || p.NoPlanet {
		return nil
	}
	sp := p.SpecialID
	if !gamedata.SpecialIsSystemDiscovery(sp) {
		return nil
	}
	p.SpecialSeen = true // 原版是把 Star.special 覆寫成訊息碼,效果相同:同一星系只觸發一次

	d := &SystemDiscovery{
		StarIndex: starIdx,
		StarName:  s.Stars[starIdx].Name,
		Special:   sp,
		Name:      gamedata.PlanetSpecialName(sp),
		ColonyIdx: -1,
	}

	switch {
	case gamedata.SpecialDiscoveryBC(sp) > 0:
		d.BCGained = gamedata.SpecialDiscoveryBC(sp)
		s.Player.BC += d.BCGained
		d.Message = fmt.Sprintf("勘查小隊在 %s 星系裡找到%s,變賣所得 %d BC 已入國庫。",
			d.StarName, d.Name, d.BCGained)

	case gamedata.SpecialFoundsSplinterColony(sp):
		if idx, pop, ok := s.foundSplinterColony(starIdx); ok {
			d.ColonyIdx = idx
			d.Message = fmt.Sprintf("%s 星系裡有一支與帝國失聯多年的同胞殖民地,%d 單位人口重歸旗下。",
				d.StarName, pop)
		} else {
			// 該星已有歸屬或行星資料不可殖民 → 只留敘述,不硬塞一個殖民地。
			d.Message = fmt.Sprintf("%s 星系傳來失散同胞的訊號,但當地已無法重建殖民地。", d.StarName)
		}

	case gamedata.SpecialGrantsFreeLeader(sp):
		if ld, ok := s.grantMaroonedLeader(); ok {
			d.LeaderGot = ld.Name
			d.Message = fmt.Sprintf("一名受困在 %s 星系的傭兵領袖獲救,%s 為報答而加入帝國麾下。",
				d.StarName, ld.Name)
		} else {
			d.Message = fmt.Sprintf("%s 星系裡救出一名受困的傭兵領袖,但帝國領袖席位已滿,對方另謀高就。", d.StarName)
		}

	case gamedata.SpecialGrantsFreeTech(sp):
		if name, ok := s.grantArtifactTech(); ok {
			d.TechGot = name
			d.Message = fmt.Sprintf("%s 星系的遠古文物解析完成,帝國直接掌握了「%s」。", d.StarName, name)
		} else {
			d.Message = fmt.Sprintf("%s 星系發現遠古文物,但帝國已無可立即解析的研究主題。", d.StarName)
		}

	default:
		return nil
	}
	return d
}

// foundSplinterColony 在 starIdx 就地生出失散殖民地,回傳殖民地索引與起始人口。
// 該星已有歸屬、或行星資料不可殖民時回 ok=false(不改任何狀態)。
func (s *GameSession) foundSplinterColony(starIdx int) (idx, pop int, ok bool) {
	if starIdx < 0 || starIdx >= len(s.Stars) || s.Stars[starIdx].Owner != 0 {
		return 0, 0, false
	}
	foodBonus, indBonus, resBonus := 0, 0, 0
	if s.RaceIndex >= 0 && s.RaceIndex < len(Races) {
		r := Races[s.RaceIndex]
		foodBonus, indBonus, resBonus = r.FoodBonus, r.IndBonus, r.ResBonus
	}
	colony, cok, _ := s.newColonyFromStar(starIdx, s.Government, foodBonus, indBonus, resBonus)
	if !cok {
		return 0, 0, false
	}
	// 人口 = min(該行星人口上限, 3)(原版 Colony_Race_Pop_Limit_ 後接 cmp al,3 的夾擠)。
	pop = colony.PopMax
	if pop > gamedata.SplinterColonyMaxPopulation {
		pop = gamedata.SplinterColonyMaxPopulation
	}
	if pop < 1 {
		pop = 1
	}
	colony.Population = pop
	colony.Farmers = pop // 原版這批人預設是農夫(可耕的星),與殖民船新殖民地同一個預設
	s.appendPlayerColony(colony, starIdx)
	s.Stars[starIdx].Owner = 1
	return len(s.PlayerColonies) - 1, pop, true
}

// grantMaroonedLeader 免費招入一名受困領袖(不扣 BC),回傳該領袖。
// 對應類別(殖民地/艦艇)的四席已滿、或候選池已用盡時回 ok=false。
func (s *GameSession) grantMaroonedLeader() (Leader, bool) {
	cands := s.mercCandidates()
	for i := s.MercOfferedIdx; i < len(cands); i++ {
		ld := cands[i]
		if s.leaderSlotsFull(ld.Ship) {
			continue
		}
		s.MercOfferedIdx = i + 1
		s.Leaders = append(s.Leaders, ld)
		// 只套新入列這一名的殖民地加成(applyLeaderColonyBonuses 是累加,見 HireMerc 同款註解)。
		if len(s.PlayerColonies) > 0 && !ld.Ship {
			applyLeaderColonyBonuses([]Leader{ld}, &s.PlayerColonies[0])
		}
		return ld, true
	}
	return Leader{}, false
}

// grantArtifactTech 白送一項研究主題(標記為已完成),回傳其顯示名。
// 沒有可送的主題(全都研究完了)回 ok=false。
func (s *GameSession) grantArtifactTech() (string, bool) {
	if s.Player.CompletedTopics == nil {
		s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{}
	}
	for _, t := range researchQueue() {
		if s.Player.CompletedTopics[t] {
			continue
		}
		s.Player.CompletedTopics[t] = true
		s.advanceResearch() // 目前主題若正好是它,順勢推進到下一個,避免研究卡在已完成的主題
		return gamedata.TopicEnglishName(t), true
	}
	return "", false
}

// appendPlayerColony 把一筆殖民地接到 PlayerColonies 與全部平行陣列上(padding 模式與
// ColonizeStar 相同——那邊是 inline 展開的,這裡抽成函式給 discovery 共用)。
func (s *GameSession) appendPlayerColony(colony engine.ColonyState, starIdx int) {
	s.PlayerColonies = append(s.PlayerColonies, colony)
	s.Builds = append(s.Builds, ColonyBuild{})
	for len(s.ColonyBuildings) < len(s.PlayerColonies) {
		s.ColonyBuildings = append(s.ColonyBuildings, nil)
	}
	for len(s.PlayerColonyMarines) < len(s.PlayerColonies) {
		s.PlayerColonyMarines = append(s.PlayerColonyMarines, 0)
	}
	for len(s.MarineBarracksAge) < len(s.PlayerColonies) {
		s.MarineBarracksAge = append(s.MarineBarracksAge, 0)
	}
	for len(s.PlayerColonyTanks) < len(s.PlayerColonies) {
		s.PlayerColonyTanks = append(s.PlayerColonyTanks, 0)
	}
	for len(s.ArmorBarracksAge) < len(s.PlayerColonies) {
		s.ArmorBarracksAge = append(s.ArmorBarracksAge, 0)
	}
	for len(s.popAccum) < len(s.PlayerColonies) {
		s.popAccum = append(s.popAccum, 0)
	}
	for len(s.PlayerColonyStars) < len(s.PlayerColonies)-1 {
		s.PlayerColonyStars = append(s.PlayerColonyStars, -1)
	}
	s.PlayerColonyStars = append(s.PlayerColonyStars, starIdx)
}
