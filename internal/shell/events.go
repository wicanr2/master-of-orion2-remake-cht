package shell

import (
	"fmt"
	"math/rand"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 隨機事件:對齊原版 36 種事件表(gamedata.RandomEvents)。
//
// 換掉的舊實作:remake 先前是 6 種自編事件(經濟繁榮/太空海盜/富礦脈/瘟疫/科學突破/隕石),
// 名稱與效果都是自己想的。原版的事件清單、好壞旗標與訊息文字都在二進位與 EVENTMSG.LBX 裡,
// 沒有理由繼續用自編版本(見 gamedata/events.go 的來源說明)。
//
// 只從 `gamedata.ImplementedRandomEvents()` 抽:那些是 remake 已有子系統可以忠實結算的。
// 需要新機制的(太空怪獸、超新星、曲速漏斗…)照樣列在 gamedata 表裡,但不會被抽中——
// 寧可事件種類少一點,也不要抽到一個「跳出訊息卻什麼都沒發生」的假事件。
//
// 訊息文字比照原版的 GNN 新聞快報語氣(原文見 EVENTMSG.LBX 資產 8+id*4);
// remake 用自己的中文文案而非直譯,因為原版訊息帶 \x80..\x92 這一整套佔位符,
// 要逐一還原成 remake 的資料來源才能填,那是獨立的中文化工項。

// EventReport 是一則已發生的事件,供 UI(事件畫面/回合摘要)顯示。
type EventReport struct {
	EventID int    // gamedata.RandomEvent.ID
	Name    string // 事件名(中文)
	Good    bool   // 原版 _event_good_array
	Message string // GNN 風格播報文字(已填入殖民地名/數字)
}

// eventChancePerTurn 是每回合觸發事件的機率。
// 原版的觸發節奏在 Determine_Event_ 裡(還沒逐條反編),先沿用 remake 既有的 30%,
// 標明這是 remake 值而非原版值,免得日後被當成考據結果。
const eventChancePerTurn = 0.30

// advanceEvents 擲一次隨機事件並結算。結果同時寫入 LastEvent(既有的回合摘要文字)
// 與 LastEventReport(事件畫面用的結構化資料)。
func (s *GameSession) advanceEvents() {
	s.LastEvent = ""
	s.LastEventReport = nil
	if s.DisableEvents {
		return
	}
	if s.eventRand == nil {
		s.eventRand = rand.New(rand.NewSource(s.EventSeed*2654435761 + 1))
	}
	if s.eventRand.Float64() >= eventChancePerTurn {
		return
	}
	pool := gamedata.ImplementedRandomEvents()
	if len(pool) == 0 {
		return
	}
	// 抽到「當下沒有適用對象」的事件(例如沒有艦隊時抽到艦船爆炸)就重抽,
	// 最多試幾輪;全部不適用就這回合沒事件,不硬湊一個空事件出來。
	for try := 0; try < 8; try++ {
		ev := pool[s.eventRand.Intn(len(pool))]
		if msg, ok := s.applyRandomEvent(ev); ok {
			s.LastEvent = msg
			s.LastEventReport = &EventReport{EventID: ev.ID, Name: ev.Name, Good: ev.Good, Message: msg}
			return
		}
	}
}

// applyRandomEvent 結算一個事件。回傳 ok=false 表示當下沒有適用對象(呼叫端會重抽)。
func (s *GameSession) applyRandomEvent(ev gamedata.RandomEvent) (string, bool) {
	switch ev.ID {
	case 0: // 古代遺骸科技:考古隊在古船殘骸裡挖到科技
		gain := 100 + s.Turn*2
		s.Player.ResearchProgress += gain
		return fmt.Sprintf("考古隊在一艘古代異星船艦的殘骸中解出失落的科技,研究進度 +%d RP", gain), true

	case 1: // 氣候改善:行星軸心偏移,氣候往 Terran 方向推進一級
		i, ok := s.pickColony()
		if !ok {
			return "", false
		}
		targets := gamedata.TerraformNextClimateOptions(s.PlayerColonies[i].Climate)
		if len(targets) == 0 {
			return "", false // 已是 Terran/Gaia 或該氣候不可改造
		}
		before := s.PlayerColonies[i].Climate
		s.applyClimateChange(i, targets[0])
		return fmt.Sprintf("%s 的行星軸心突然偏移,環境由 %s 改善為 %s,農民歡聲雷動",
			s.colonyLabel(i), climateDisplayName(before), climateDisplayName(targets[0])), true

	case 3: // 電腦病毒:研究中心電腦全面感染
		if s.Player.ResearchProgress <= 0 {
			return "", false
		}
		loss := s.Player.ResearchProgress / 3
		if loss < 1 {
			loss = s.Player.ResearchProgress
		}
		s.Player.ResearchProgress -= loss
		return fmt.Sprintf("研究中心的電腦遭病毒全面感染,損失 %d 點研究進度", loss), true

	case 4: // 外交暗殺:被抓到策劃暗殺,關係惡化
		j, ok := s.pickAI()
		if !ok {
			return "", false
		}
		s.adjustAIRelation(j, -12)
		return fmt.Sprintf("一名外交官被查獲策劃暗殺%s高層,兩國關係惡化", s.AIPlayers[j].Name), true

	case 5: // 外交聯姻:關係改善
		j, ok := s.pickAI()
		if !ok {
			return "", false
		}
		s.adjustAIRelation(j, +12)
		return fmt.Sprintf("我方大使與%s軍方要員締結聯姻,兩國關係大幅改善", s.AIPlayers[j].Name), true

	case 6: // 富商捐獻
		gain := 50 + s.Turn
		s.Player.BC += gain
		return fmt.Sprintf("一名富商向帝國捐獻了 %d BC", gain), true

	case 7: // 地震:人口與建築受損
		i, ok := s.pickColony()
		if !ok || s.PlayerColonies[i].Population <= 1 {
			return "", false
		}
		lost := s.losePopulation(i, 1+s.eventRand.Intn(2))
		bldg := s.destroyRandomBuilding(i)
		if bldg != "" {
			return fmt.Sprintf("%s 發生劇烈地震,%d 百萬居民罹難,%s 毀於一旦",
				s.colonyLabel(i), lost, bldg), true
		}
		return fmt.Sprintf("%s 發生劇烈地震,%d 百萬居民罹難", s.colonyLabel(i), lost), true

	case 8: // 艦船爆炸:一艘軍艦離奇爆炸
		// **打的是整個帝國**不是玩家目前選中的那一支艦隊——事件不看玩家的操作焦點。
		if s.ShipCount() <= 1 {
			return "", false // 只剩一艘就不炸,避免玩家瞬間失去全部艦隊
		}
		lost, ok := s.removeShipGlobal(s.eventRand.Intn(s.ShipCount()))
		if !ok {
			return "", false
		}
		return fmt.Sprintf("軍艦「%s」離奇爆炸,調查仍在進行中", lost.Name), true

	case 10: // 工業意外:輻射污染,氣候惡化 + 人口損失
		i, ok := s.pickColony()
		if !ok || s.PlayerColonies[i].Population <= 1 {
			return "", false
		}
		lost := s.losePopulation(i, 2)
		return fmt.Sprintf("%s 發生重大工業事故,輻射擴散全球,%d 百萬平民喪生",
			s.colonyLabel(i), lost), true

	case 11: // 礦產枯竭:礦產等級下降一級
		i, from, to, ok := s.shiftColonyMineral(-1)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("礦工長期超量開採,%s 的礦產等級由「%s」降為「%s」",
			s.colonyLabel(i), mineralDisplayName(from), mineralDisplayName(to)), true

	case 12: // 礦產發現:礦產等級上升一級
		i, from, to, ok := s.shiftColonyMineral(+1)
		if !ok {
			return "", false
		}
		return fmt.Sprintf("勘探隊在 %s 發現前所未知的礦脈,礦產等級由「%s」升為「%s」",
			s.colonyLabel(i), mineralDisplayName(from), mineralDisplayName(to)), true

	case 13: // 艦船叛變:一艘艦倒戈給某個 AI
		// 同上:叛變的是帝國裡的任何一艘,不限玩家目前選中的艦隊。
		if s.ShipCount() <= 1 {
			return "", false
		}
		j, ok := s.pickAI()
		if !ok {
			return "", false
		}
		lost, ok := s.removeShipGlobal(s.eventRand.Intn(s.ShipCount()))
		if !ok {
			return "", false
		}
		s.AIPlayers[j].FleetStrength += 10
		return fmt.Sprintf("軍艦「%s」發生叛變,投奔%s", lost.Name, s.AIPlayers[j].Name), true

	case 15: // 海盜劫掠:國庫被偷
		if s.Player.BC <= 0 {
			return "", false
		}
		loss := 40
		if loss > s.Player.BC {
			loss = s.Player.BC
		}
		s.Player.BC -= loss
		return fmt.Sprintf("海盜突襲得手,自國庫竊走 %d BC", loss), true

	case 16: // 瘟疫:人口大量死亡
		i, ok := s.pickColony()
		if !ok || s.PlayerColonies[i].Population <= 1 {
			return "", false
		}
		lost := s.losePopulation(i, 2)
		return fmt.Sprintf("%s 爆發瘟疫,已有 %d 百萬居民死亡,全境進入隔離",
			s.colonyLabel(i), lost), true

	case 17: // 人口暴增:成長率倍增(remake 以一次性 +2 人口表達)
		i, ok := s.pickColony()
		if !ok {
			return "", false
		}
		c := &s.PlayerColonies[i]
		if c.Population >= c.PopMax {
			return "", false
		}
		gain := 2
		if c.Population+gain > c.PopMax {
			gain = c.PopMax - c.Population
		}
		c.Population += gain
		for n := 0; n < gain; n++ {
			s.assignNewColonist(i)
		}
		return fmt.Sprintf("%s 出現人口暴增,新生兒數量翻倍,人口 +%d", s.colonyLabel(i), gain), true

	case 18: // 秘密實驗:高層機密實驗撞出新科技
		gain := 80 + s.Turn
		s.Player.ResearchProgress += gain
		return fmt.Sprintf("科學家在高層機密實驗中意外撞見新技術的關鍵,研究進度 +%d RP", gain), true

	// --- 太空怪獸入侵(19-23)。每種怪獸的**最早回合**由手冊 p.180-181 逐條給定,
	//     照抄不臆改;怪獸實體與戰鬥見 monster.go。太空鰻在 remake 沒有獨立實體
	//     (手冊:牠只封鎖不攻擊),用九頭蛇以外最接近「純擋路」的變形蟲代打並標明。---
	case 19: // 太空變形蟲(手冊:≥100 回合)
		return s.spawnInvadingMonster(gamedata.MonsterAmoeba, 100)
	case 20: // 太空水晶(手冊:≥200 回合)
		return s.spawnInvadingMonster(gamedata.MonsterCrystal, 200)
	case 21: // 太空巨龍(手冊:≥300 回合)
		return s.spawnInvadingMonster(gamedata.MonsterDragon, 300)
	case 22: // 太空鰻(手冊:≥150 回合)。
		// ⚠ remake 沒有太空鰻的獨立實體(手冊 p.180:牠「never attack colonies or outposts」、
		// 只封鎖系統,且 30 回合後會分裂)。目前用一頭盤據星系的怪獸近似「封鎖」那一半,
		// 分裂與「不攻擊」的差異尚未建模——標明,不假裝已經忠實。
		return s.spawnInvadingMonster(gamedata.MonsterAmoeba, 150)
	case 23: // 太空九頭蛇(手冊:≥250 回合)
		return s.spawnInvadingMonster(gamedata.MonsterHydra, 250)

	// --- 持續型事件(24-26、28),見 events_persistent.go ---
	case 24: // 超新星(手冊:≥200 回合,倒數 6-14)
		return s.startSupernova()
	case 25: // 時空異象:整個星系凍結
		return s.startStasis()
	case 26: // 超空間獸:航行中的艦隊有機率損失艦艇
		return s.startWarpBeast()
	case 28: // 蟲洞:航行中的艦隊瞬間抵達
		return s.applyWormhole()
	}
	return "", false
}

// --- 事件用的小工具(都不改變既有公開行為,只是把重複的挑選/扣減收斂起來)---

// pickColony 隨機挑一個玩家殖民地。無殖民地時回 ok=false。
func (s *GameSession) pickColony() (int, bool) {
	if len(s.PlayerColonies) == 0 {
		return 0, false
	}
	return s.eventRand.Intn(len(s.PlayerColonies)), true
}

// pickAI 隨機挑一個 AI 對手。無對手時回 ok=false。
func (s *GameSession) pickAI() (int, bool) {
	if len(s.AIPlayers) == 0 {
		return 0, false
	}
	return s.eventRand.Intn(len(s.AIPlayers)), true
}

// colonyLabel 回傳殖民地的顯示名(優先用所在行星名,取不到才用序號)。
func (s *GameSession) colonyLabel(i int) string {
	if star := s.PlayerColonyStarIndex(i); star >= 0 && star < len(s.Planets) {
		if n := s.Planets[star].Name; n != "" {
			return n
		}
	}
	return fmt.Sprintf("殖民地 %d", i+1)
}

// losePopulation 從殖民地 i 扣掉最多 n 人口(至少留 1),由人數最多的職務扣起。
// 回傳實際扣掉的數量。
func (s *GameSession) losePopulation(i, n int) int {
	c := &s.PlayerColonies[i]
	lost := 0
	for ; n > 0 && c.Population > 1; n-- {
		c.Population--
		lost++
		switch {
		case c.Workers >= c.Farmers && c.Workers >= c.Scientists && c.Workers > 0:
			c.Workers--
		case c.Farmers >= c.Scientists && c.Farmers > 0:
			c.Farmers--
		case c.Scientists > 0:
			c.Scientists--
		}
	}
	return lost
}

// destroyRandomBuilding 隨機摧毀殖民地 i 的一棟建築,回傳建築名(沒有可摧毀的回空字串)。
// 首都不摧毀(原版 Pick_Random_Colony_No_Capitol_ 同樣把首都排除在事件目標外)。
func (s *GameSession) destroyRandomBuilding(i int) string {
	if i < 0 || i >= len(s.ColonyBuildings) || len(s.ColonyBuildings[i]) == 0 {
		return ""
	}
	names := make([]string, 0, len(s.ColonyBuildings[i]))
	for n := range s.ColonyBuildings[i] {
		if n == "首都" {
			continue
		}
		names = append(names, n)
	}
	if len(names) == 0 {
		return ""
	}
	// map 迭代順序不穩定,排序後再抽才能讓同一個 seed 重現同樣結果(事件必須可重現,
	// 見 events_test.go 的可重現性測試)。
	sortStrings(names)
	pick := names[s.eventRand.Intn(len(names))]
	delete(s.ColonyBuildings[i], pick)
	s.recalcColonyMorale(i)
	return pick
}

// shiftColonyMineral 把某個殖民地的礦產等級往 delta 方向移一級(±1),
// 回傳殖民地索引與前後等級。沒有可移動的殖民地(都已到上下限)時回 ok=false。
//
// 礦產等級存在 Planets[star].MineralID(星圖行星資料),殖民地的每礦工工業
// (IndustryPerWorker)由它推導,故兩者要一起更新才不會又出現「面板說豐富、產出卻沒變」。
func (s *GameSession) shiftColonyMineral(delta int) (idx int, from, to gamedata.PlanetMinerals, ok bool) {
	cand := make([]int, 0, len(s.PlayerColonies))
	for i := range s.PlayerColonies {
		star := s.PlayerColonyStarIndex(i)
		if star < 0 || star >= len(s.Planets) {
			continue
		}
		m := int(s.Planets[star].MineralID) + delta
		if m < int(gamedata.ULTRA_POOR) || m > int(gamedata.ULTRA_RICH) {
			continue
		}
		cand = append(cand, i)
	}
	if len(cand) == 0 {
		return 0, 0, 0, false
	}
	i := cand[s.eventRand.Intn(len(cand))]
	star := s.PlayerColonyStarIndex(i)
	from = s.Planets[star].MineralID
	to = gamedata.PlanetMinerals(int(from) + delta)
	s.Planets[star].MineralID = to
	s.Planets[star].Mineral = mineralDisplayName(to)
	// 重力也跟著礦產(密度)變——原版 _gravity_table 就是 [mineral][size]。
	s.Planets[star].GravityID = gamedata.PlanetGravityFor(to, s.Planets[star].SizeID)
	s.Planets[star].Gravity = gravityDisplayName(s.Planets[star].GravityID)
	s.PlayerColonies[i].IndustryPerWorker = gamedata.MineralIndustryPerWorker(to)
	return i, from, to, true
}

// adjustAIRelation 調整玩家與 AI j 的關係分數(夾在既有的 -40..40 尺度內)。
func (s *GameSession) adjustAIRelation(j, delta int) {
	if j < 0 || j >= len(s.AIPlayers) {
		return
	}
	r := s.AIPlayers[j].Relation + delta
	if r > 40 {
		r = 40
	}
	if r < -40 {
		r = -40
	}
	s.AIPlayers[j].Relation = r
}

// 避免為了一次排序把 sort 匯入撒進整個檔案(events.go 只有這裡需要)。
func sortStrings(a []string) {
	for i := 1; i < len(a); i++ {
		for j := i; j > 0 && a[j] < a[j-1]; j-- {
			a[j], a[j-1] = a[j-1], a[j]
		}
	}
}

// 編譯期確認 engine 仍被使用(applyClimateChange 等在 session.go,這裡只是型別參照)。
var _ = engine.ColonyState{}

// eventRandForTest 確保 eventRand 已初始化(測試直接呼叫事件工具函式時用)。
func (s *GameSession) eventRandForTest() {
	if s.eventRand == nil {
		s.eventRand = rand.New(rand.NewSource(s.EventSeed*2654435761 + 1))
	}
}
