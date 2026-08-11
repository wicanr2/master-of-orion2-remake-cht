package shell

import (
	"fmt"

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
	EventID   int    // gamedata.RandomEvent.ID
	Name      string // 事件名(中文)
	NameEN    string // 事件名(英文顯示)
	Good      bool   // 原版 _event_good_array
	Message   string // GNN 風格播報文字(已填入殖民地名/數字)
	MessageEN string // 同一結算結果的英文播報文字
}

// eventResult 是事件結算的雙語顯示結果。規則效果只在同一個 switch 裡執行一次，
// 兩種語言共用同一組已落地的數字，避免翻譯分支重算亂數或效果。
type eventResult struct {
	Message   string
	MessageEN string
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
		s.eventRand = newRandStream(s.EventSeed*2654435761 + 1)
	}
	if s.eventRand.Float64() >= eventChancePerTurn {
		return
	}
	pool := randomEventPoolForSession(s)
	if len(pool) == 0 {
		return
	}
	// 抽到「當下沒有適用對象」的事件(例如沒有艦隊時抽到艦船爆炸)就重抽,
	// 最多試幾輪;全部不適用就這回合沒事件,不硬湊一個空事件出來。
	for try := 0; try < 8; try++ {
		ev := pool[s.eventRand.Intn(len(pool))]
		if result, ok := s.applyRandomEventLocalized(ev); ok {
			s.LastEvent = result.Message
			s.LastEventReport = &EventReport{
				EventID: ev.ID, Name: ev.Name, NameEN: eventNameEN(ev.ID), Good: ev.Good,
				Message: result.Message, MessageEN: result.MessageEN,
			}
			return
		}
	}
}

// randomEventPoolForSession 把種族的幸運效果放在抽樣入口，而不是在每個事件
// switch 裡散落特判。這樣所有已實作的災害都會被排除，好事件的相對機率也會
// 提升；沒有好事件時才保留一般池，避免自訂測試局卡死。
func randomEventPoolForSession(s *GameSession) []gamedata.RandomEvent {
	pool := gamedata.ImplementedRandomEvents()
	if s == nil || !s.RaceLucky() {
		return pool
	}
	good := make([]gamedata.RandomEvent, 0, len(pool))
	for _, ev := range pool {
		if ev.Good {
			good = append(good, ev)
		}
	}
	if len(good) > 0 {
		return good
	}
	return pool
}

// applyRandomEvent 結算一個事件。回傳 ok=false 表示當下沒有適用對象(呼叫端會重抽)。
// 保留這個純中文包裝給既有 shell 測試與內部呼叫；畫面路徑使用下方雙語結果。
func (s *GameSession) applyRandomEvent(ev gamedata.RandomEvent) (string, bool) {
	result, ok := s.applyRandomEventLocalized(ev)
	return result.Message, ok
}

func (s *GameSession) applyRandomEventLocalized(ev gamedata.RandomEvent) (eventResult, bool) {
	switch ev.ID {
	case 0: // 古代遺骸科技:考古隊在古船殘骸裡挖到科技
		gain := 100 + s.Turn*2
		s.Player.ResearchProgress += gain
		return eventResult{
			Message:   fmt.Sprintf("考古隊在一艘古代異星船艦的殘骸中解出失落的科技,研究進度 +%d RP", gain),
			MessageEN: fmt.Sprintf("Archaeologists recovered lost technology from the wreckage of an ancient alien ship. Research progress +%d RP.", gain),
		}, true

	case 1: // 氣候改善:行星軸心偏移,氣候往 Terran 方向推進一級
		i, ok := s.pickColony()
		if !ok {
			return eventResult{}, false
		}
		targets := gamedata.TerraformNextClimateOptions(s.PlayerColonies[i].Climate)
		if len(targets) == 0 {
			return eventResult{}, false // 已是 Terran/Gaia 或該氣候不可改造
		}
		before := s.PlayerColonies[i].Climate
		s.applyClimateChange(i, targets[0])
		return eventResult{
			Message: fmt.Sprintf("%s 的行星軸心突然偏移,環境由 %s 改善為 %s,農民歡聲雷動",
				s.colonyLabel(i), climateDisplayName(before), climateDisplayName(targets[0])),
			MessageEN: fmt.Sprintf("The climate on the affected colony shifted from %s to %s. Its farmers celebrate.",
				climateDisplayNameEN(before), climateDisplayNameEN(targets[0])),
		}, true

	case 3: // 電腦病毒:研究中心電腦全面感染
		if s.Player.ResearchProgress <= 0 {
			return eventResult{}, false
		}
		loss := s.Player.ResearchProgress / 3
		if loss < 1 {
			loss = s.Player.ResearchProgress
		}
		s.Player.ResearchProgress -= loss
		return eventResult{
			Message:   fmt.Sprintf("研究中心的電腦遭病毒全面感染,損失 %d 點研究進度", loss),
			MessageEN: fmt.Sprintf("A computer virus infected the research network. The empire lost %d research progress.", loss),
		}, true

	case 4: // 外交暗殺:被抓到策劃暗殺,關係惡化
		j, ok := s.pickAI()
		if !ok {
			return eventResult{}, false
		}
		s.adjustAIRelation(j, -12)
		return eventResult{
			Message:   fmt.Sprintf("一名外交官被查獲策劃暗殺%s高層,兩國關係惡化", s.AIPlayers[j].Name),
			MessageEN: "A diplomat was caught plotting an assassination against a rival empire's leadership. Relations have worsened.",
		}, true

	case 5: // 外交聯姻:關係改善
		j, ok := s.pickAI()
		if !ok {
			return eventResult{}, false
		}
		s.adjustAIRelation(j, +12)
		return eventResult{
			Message:   fmt.Sprintf("我方大使與%s軍方要員締結聯姻,兩國關係大幅改善", s.AIPlayers[j].Name),
			MessageEN: "Our ambassador married into a rival empire's military leadership. Relations have greatly improved.",
		}, true

	case 6: // 富商捐獻
		gain := 50 + s.Turn
		s.Player.BC += gain
		return eventResult{
			Message:   fmt.Sprintf("一名富商向帝國捐獻了 %d BC", gain),
			MessageEN: fmt.Sprintf("A wealthy merchant donated %d BC to the empire.", gain),
		}, true

	case 7: // 地震:人口與建築受損
		i, ok := s.pickColony()
		if !ok || s.PlayerColonies[i].Population <= 1 {
			return eventResult{}, false
		}
		lost := s.losePopulation(i, 1+s.eventRand.Intn(2))
		bldg := s.destroyRandomBuilding(i)
		if bldg != "" {
			return eventResult{
				Message: fmt.Sprintf("%s 發生劇烈地震,%d 百萬居民罹難,%s 毀於一旦",
					s.colonyLabel(i), lost, bldg),
				MessageEN: fmt.Sprintf("A violent earthquake struck a colony. %d million residents died and a building was destroyed.", lost),
			}, true
		}
		return eventResult{
			Message:   fmt.Sprintf("%s 發生劇烈地震,%d 百萬居民罹難", s.colonyLabel(i), lost),
			MessageEN: fmt.Sprintf("A violent earthquake struck a colony. %d million residents died.", lost),
		}, true

	case 8: // 艦船爆炸:一艘軍艦離奇爆炸
		// **打的是整個帝國**不是玩家目前選中的那一支艦隊——事件不看玩家的操作焦點。
		if s.ShipCount() <= 1 {
			return eventResult{}, false // 只剩一艘就不炸,避免玩家瞬間失去全部艦隊
		}
		impact, ok := s.resolveStrategicShipExplosion()
		if !ok {
			return eventResult{}, false
		}
		collateral := ""
		if len(impact.Collateral) > 0 {
			collateral = fmt.Sprintf(",衝擊波使 %d 艘倖存艦受損", len(impact.Collateral))
		}
		return eventResult{
			Message: fmt.Sprintf("軍艦「%s」離奇爆炸%s,調查仍在進行中", impact.Lost.Name, collateral),
			MessageEN: fmt.Sprintf("The warship \"%s\" exploded mysteriously%s. The investigation continues.", impact.Lost.Name, func() string {
				if len(impact.Collateral) == 0 {
					return ""
				}
				return fmt.Sprintf("; the blast damaged %d surviving ships", len(impact.Collateral))
			}()),
		}, true

	case 10: // 工業意外:輻射污染,氣候惡化 + 人口損失
		i, ok := s.pickColony()
		if !ok || s.PlayerColonies[i].Population <= 1 {
			return eventResult{}, false
		}
		lost := s.losePopulation(i, 2)
		return eventResult{
			Message: fmt.Sprintf("%s 發生重大工業事故,輻射擴散全球,%d 百萬平民喪生",
				s.colonyLabel(i), lost),
			MessageEN: fmt.Sprintf("A major industrial accident spread radiation across a colony. %d million civilians died.", lost),
		}, true

	case 11: // 礦產枯竭:礦產等級下降一級
		i, from, to, ok := s.shiftColonyMineral(-1)
		if !ok {
			return eventResult{}, false
		}
		return eventResult{
			Message: fmt.Sprintf("礦工長期超量開採,%s 的礦產等級由「%s」降為「%s」",
				s.colonyLabel(i), mineralDisplayName(from), mineralDisplayName(to)),
			MessageEN: fmt.Sprintf("Long-term overmining depleted a colony's mineral deposits from %s to %s.",
				mineralDisplayNameEN(from), mineralDisplayNameEN(to)),
		}, true

	case 12: // 礦產發現:礦產等級上升一級
		i, from, to, ok := s.shiftColonyMineral(+1)
		if !ok {
			return eventResult{}, false
		}
		return eventResult{
			Message: fmt.Sprintf("勘探隊在 %s 發現前所未知的礦脈,礦產等級由「%s」升為「%s」",
				s.colonyLabel(i), mineralDisplayName(from), mineralDisplayName(to)),
			MessageEN: fmt.Sprintf("Explorers found a previously unknown mineral vein. A colony's deposits improved from %s to %s.",
				mineralDisplayNameEN(from), mineralDisplayNameEN(to)),
		}, true

	case 13: // 艦船叛變:一艘艦倒戈給某個 AI
		// 同上:叛變的是帝國裡的任何一艘,不限玩家目前選中的艦隊。
		if s.ShipCount() <= 1 {
			return eventResult{}, false
		}
		j, ok := s.pickAI()
		if !ok {
			return eventResult{}, false
		}
		lost, ok := s.removeShipGlobal(s.eventRand.Intn(s.ShipCount()))
		if !ok {
			return eventResult{}, false
		}
		s.AIPlayers[j].FleetStrength += 10
		return eventResult{
			Message:   fmt.Sprintf("軍艦「%s」發生叛變,投奔%s", lost.Name, s.AIPlayers[j].Name),
			MessageEN: "A warship mutinied and defected to a rival empire.",
		}, true

	case 15: // 海盜劫掠:國庫被偷
		if s.Player.BC <= 0 {
			return eventResult{}, false
		}
		loss := 40
		if loss > s.Player.BC {
			loss = s.Player.BC
		}
		s.Player.BC -= loss
		return eventResult{
			Message:   fmt.Sprintf("海盜突襲得手,自國庫竊走 %d BC", loss),
			MessageEN: fmt.Sprintf("Pirates raided the treasury and stole %d BC.", loss),
		}, true

	case 16: // 瘟疫:人口大量死亡
		i, ok := s.pickColony()
		if !ok || s.PlayerColonies[i].Population <= 1 {
			return eventResult{}, false
		}
		lost := s.losePopulation(i, 2)
		return eventResult{
			Message: fmt.Sprintf("%s 爆發瘟疫,已有 %d 百萬居民死亡,全境進入隔離",
				s.colonyLabel(i), lost),
			MessageEN: fmt.Sprintf("A plague broke out in a colony. %d million residents died and the colony entered quarantine.", lost),
		}, true

	case 17: // 人口暴增:成長率倍增(remake 以一次性 +2 人口表達)
		i, ok := s.pickColony()
		if !ok {
			return eventResult{}, false
		}
		c := &s.PlayerColonies[i]
		if c.Population >= c.PopMax {
			return eventResult{}, false
		}
		gain := 2
		if c.Population+gain > c.PopMax {
			gain = c.PopMax - c.Population
		}
		c.Population += gain
		for n := 0; n < gain; n++ {
			s.assignNewColonist(i)
		}
		return eventResult{
			Message:   fmt.Sprintf("%s 出現人口暴增,新生兒數量翻倍,人口 +%d", s.colonyLabel(i), gain),
			MessageEN: fmt.Sprintf("A population boom doubled the number of births at a colony. Population +%d.", gain),
		}, true

	case 18: // 秘密實驗:高層機密實驗撞出新科技
		gain := 80 + s.Turn
		s.Player.ResearchProgress += gain
		return eventResult{
			Message:   fmt.Sprintf("科學家在高層機密實驗中意外撞見新技術的關鍵,研究進度 +%d RP", gain),
			MessageEN: fmt.Sprintf("Scientists found a breakthrough during a classified experiment. Research progress +%d RP.", gain),
		}, true

	// --- 太空怪獸入侵(19-23)。每種怪獸的**最早回合**由手冊 p.180-181 逐條給定,
	//     照抄不臆改;怪獸實體與戰鬥見 monster.go。太空鰻在 remake 沒有獨立實體
	//     (手冊:牠只封鎖不攻擊),用九頭蛇以外最接近「純擋路」的變形蟲代打並標明。---
	case 19: // 太空變形蟲(手冊:≥100 回合)
		message, ok := s.spawnInvadingMonster(gamedata.MonsterAmoeba, 100)
		return wrapEventResult(s, ev, message, ok)
	case 20: // 太空水晶(手冊:≥200 回合)
		message, ok := s.spawnInvadingMonster(gamedata.MonsterCrystal, 200)
		return wrapEventResult(s, ev, message, ok)
	case 21: // 太空巨龍(手冊:≥300 回合)
		message, ok := s.spawnInvadingMonster(gamedata.MonsterDragon, 300)
		return wrapEventResult(s, ev, message, ok)
	case 22: // 太空鰻(手冊:≥150 回合)。
		// ⚠ remake 沒有太空鰻的獨立實體(手冊 p.180:牠「never attack colonies or outposts」、
		// 只封鎖系統,且 30 回合後會分裂)。目前用一頭盤據星系的怪獸近似「封鎖」那一半,
		// 分裂與「不攻擊」的差異尚未建模——標明,不假裝已經忠實。
		message, ok := s.spawnInvadingMonster(gamedata.MonsterAmoeba, 150)
		return wrapEventResult(s, ev, message, ok)
	case 23: // 太空九頭蛇(手冊:≥250 回合)
		message, ok := s.spawnInvadingMonster(gamedata.MonsterHydra, 250)
		return wrapEventResult(s, ev, message, ok)

	// --- 持續型事件(24-26、28),見 events_persistent.go ---
	case 24: // 超新星(手冊:≥200 回合,倒數 6-14)
		message, ok := s.startSupernova()
		return wrapEventResult(s, ev, message, ok)
	case 25: // 時空異象:整個星系凍結
		message, ok := s.startStasis()
		return wrapEventResult(s, ev, message, ok)
	case 26: // 超空間獸:航行中的艦隊有機率損失艦艇
		message, ok := s.startWarpBeast()
		return wrapEventResult(s, ev, message, ok)
	case 28: // 蟲洞:航行中的艦隊瞬間抵達
		message, ok := s.applyWormhole()
		return wrapEventResult(s, ev, message, ok)
	}
	return eventResult{}, false
}

// wrapEventResult 保留尚未有逐句英文模板的持續事件/怪獸事件可用英文播報。
// 不把中文內文硬塞進英文畫面；精確模板仍可依原版 EVENTMSG 後續補上。
func wrapEventResult(s *GameSession, ev gamedata.RandomEvent, message string, ok bool) (eventResult, bool) {
	if !ok {
		return eventResult{}, false
	}
	messageEN := fmt.Sprintf("A %s event has been reported.", eventNameEN(ev.ID))
	// 持續事件與怪獸入侵的英文播報包含實際目標與倒數等上下文。
	switch ev.ID {
	case 19, 20, 21, 22, 23:
		if n := len(s.Monsters); n > 0 {
			m := s.Monsters[n-1]
			messageEN = fmt.Sprintf("A %s invaded the galaxy and now occupies the %s system; clear it before the system can be entered.",
				eventNameEN(ev.ID), s.starNameEN(m.StarIndex))
		}
	case 24:
		if n := len(s.PersistentEvents); n > 0 {
			e := s.PersistentEvents[n-1]
			messageEN = fmt.Sprintf("The star in the %s system became unstable; it will go supernova in %d turns unless research reaches %d RP.",
				s.starNameEN(e.StarIndex), e.Countdown, e.ResearchNeeded)
		}
	case 25:
		if n := len(s.PersistentEvents); n > 0 {
			messageEN = fmt.Sprintf("A space-time anomaly froze the %s system; its colonies produce nothing and do not grow.",
				s.starNameEN(s.PersistentEvents[n-1].StarIndex))
		}
	case 26:
		messageEN = "A warp beast began roaming the space lanes; ships in transit may be dragged into another dimension."
	case 28:
		messageEN = "A wormhole appeared along the fleet's route and shortened the journey to one turn."
	}
	return eventResult{
		Message:   message,
		MessageEN: messageEN,
	}, true
}

func eventNameEN(id int) string {
	names := map[int]string{
		0: "Ancient Technology", 1: "Climate Improvement", 2: "Comet Strike", 3: "Computer Virus",
		4: "Diplomatic Assassination", 5: "Diplomatic Marriage", 6: "Merchant Donation", 7: "Earthquake",
		8: "Ship Explosion", 9: "Hyperspace Turbulence", 10: "Industrial Accident", 11: "Mineral Depletion",
		12: "Mineral Discovery", 13: "Ship Mutiny", 14: "Pirate Activity", 15: "Pirate Raid",
		16: "Plague", 17: "Population Boom", 18: "Secret Experiment", 19: "Space Amoeba",
		20: "Space Crystal", 21: "Space Dragon", 22: "Space Eel", 23: "Space Hydra",
		24: "Supernova", 25: "Space-Time Anomaly", 26: "Warp Beast", 27: "Warp Funnel",
		28: "Wormhole", 29: "Empire Destroyed", 30: "Empire Ascendant", 31: "Ranking Bulletin",
		32: "Orion Discovered", 33: "Antaran Defeat", 34: "Empire Surrender", 35: "Rebel Assimilation",
	}
	if name, ok := names[id]; ok {
		return name
	}
	return fmt.Sprintf("Event #%d", id)
}

func climateDisplayNameEN(c gamedata.PlanetClimate) string {
	names := [...]string{"Radiated", "Toxic", "Barren", "Desert", "Tundra", "Ocean", "Swamp", "Arid", "Terran", "Gaia"}
	if c < 0 || int(c) >= len(names) {
		return "Unknown"
	}
	return names[c]
}

func mineralDisplayNameEN(m gamedata.PlanetMinerals) string {
	names := [...]string{"Ultra Poor", "Poor", "Average", "Rich", "Ultra Rich"}
	if m < 0 || int(m) >= len(names) {
		return "Average"
	}
	return names[m]
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
	// 取**該殖民地座落行星**的名字,不是「那顆星的代表行星」——同星系可以有多個殖民地,
	// 用後者會讓兩個殖民地同名。
	if n := s.ColonyName(i); n != "" {
		return n
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
		p := s.ColonyPlanet(i)
		if p == nil {
			continue
		}
		m := int(p.MineralID) + delta
		if m < int(gamedata.ULTRA_POOR) || m > int(gamedata.ULTRA_RICH) {
			continue
		}
		cand = append(cand, i)
	}
	if len(cand) == 0 {
		return 0, 0, 0, false
	}
	i := cand[s.eventRand.Intn(len(cand))]
	p := s.ColonyPlanet(i)
	if p == nil {
		return 0, 0, 0, false
	}
	from = p.MineralID
	to = gamedata.PlanetMinerals(int(from) + delta)
	p.MineralID = to
	p.Mineral = mineralDisplayName(to)
	// 重力也跟著礦產(密度)變——原版 _gravity_table 就是 [mineral][size]。
	p.GravityID = gamedata.PlanetGravityFor(to, p.SizeID)
	p.Gravity = gravityDisplayName(p.GravityID)
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
		s.eventRand = newRandStream(s.EventSeed*2654435761 + 1)
	}
}
