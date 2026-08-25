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
// 原版每次仍從 0..28 抽最多五個候選；尚未實作或當下不適用的 ID 會消耗候選後失敗，
// 不先縮成 implemented pool，否則會改變其餘事件的相對機率。只有效果成功落地才播報。
//
// 訊息文字比照原版的 GNN 新聞快報語氣(原文見 EVENTMSG.LBX 資產 8+id*4);
// remake 用自己的中文文案而非直譯,因為原版訊息帶 \x80..\x92 這一整套佔位符,
// 要逐一還原成 remake 的資料來源才能填,那是獨立的中文化工項。

// EventReport 是一則已發生的事件,供 UI(事件畫面/回合摘要)顯示。
type EventReport struct {
	EventID     int    // gamedata.RandomEvent.ID
	Name        string // 事件名(中文)
	NameEN      string // 事件名(英文顯示)
	Good        bool   // 原版 _event_good_array
	Message     string // GNN 風格播報文字(已填入殖民地名/數字)
	MessageEN   string // 同一結算結果的英文播報文字
	TargetKind  string // player／seat／ai；空字串表示舊存檔或全局特殊事件
	TargetIndex int    // seat／AI 的執行期索引；player 為 0
	TargetName  string // 事件發生帝國的顯示名
	// SecondaryTarget* 供事件 34 保存接收帝國；其他事件零值不改既有 JSON。
	SecondaryTargetKind  string `json:"secondaryTargetKind,omitempty"`
	SecondaryTargetIndex int    `json:"secondaryTargetIndex,omitempty"`
	SecondaryTargetName  string `json:"secondaryTargetName,omitempty"`
}

// eventResult 是事件結算的雙語顯示結果。規則效果只在同一個 switch 裡執行一次，
// 兩種語言共用同一組已落地的數字，避免翻譯分支重算亂數或效果。
type eventResult struct {
	Message   string
	MessageEN string
}

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
	elapsed := s.Turn - 1
	if elapsed < 0 {
		elapsed = 0
	}
	s.clearEventReportsForAllSeats()
	luckyTarget, luckyForced := s.advanceAllLuckyEventCounters(elapsed)
	generalDue := false
	if !luckyForced && elapsed >= 50 && elapsed >= s.EventLastTurn {
		threshold, attempts, ok := gamedata.OriginalEventScheduleThreshold(
			elapsed-s.EventLastTurn, s.EventAttemptCounter, s.Difficulty)
		if ok {
			s.EventAttemptCounter = attempts
			generalDue = gamedata.OriginalEventScheduleRollSucceeds(
				threshold, s.eventRand.Intn(512)+1)
		}
	}
	if !luckyForced && !generalDue {
		return
	}
	// 原版 sub_23563 與事件 31 分支位於 Determine_Event_ 尾端；用 defer 保證
	// 成功事件與五候選全失敗兩條返回路徑都在效果亂數消費後才檢查狀態新聞。
	defer s.advanceStatusGrowthAndRanking()
	// 原版每次直接從 29 個 ID 抽候選，最多五次。未實作或不適用的事件也消耗
	// 候選，不可先縮成可用池，否則會改變剩餘事件的相對機率與實際播報率。
	for try := 0; try < 5; try++ {
		ev := gamedata.RandomEventByID(s.eventRand.Intn(29))
		if !eventCandidateAllowed(s, ev, luckyForced, elapsed) {
			continue
		}
		target, ok := s.chooseEventEmpireTarget(*ev, luckyTarget, luckyForced)
		if !ok {
			continue
		}
		if result, ok := s.applyRandomEventLocalizedToTarget(*ev, target); ok {
			// 事件 24 的 sub_23A5F 直接選全銀河殖民星，不走一般帝國 target；
			// 報告要回填實際星系 owner，不能沿用為了進入全局 consumer 的目前玩家代理。
			if ev.ID == 24 && len(s.PersistentEvents) > 0 {
				if actual, found := s.eventEmpireTargetAtStar(s.PersistentEvents[len(s.PersistentEvents)-1].StarIndex); found {
					target = actual
				}
			}
			s.LastEvent = result.Message
			report := &EventReport{
				EventID: ev.ID, Name: ev.Name, NameEN: eventNameEN(ev.ID), Good: ev.Good,
				Message: result.Message, MessageEN: result.MessageEN,
				TargetKind: target.kind.String(), TargetIndex: target.index,
				TargetName: s.eventEmpireTargetName(target),
			}
			s.LastEventReport = report
			s.broadcastEventReport(report)
			s.EventLastTurn = elapsed
			s.EventAttemptCounter = 0
			return
		}
	}
}

func eventCandidateAllowed(s *GameSession, ev *gamedata.RandomEvent, luckyForced bool, elapsed int) bool {
	if s == nil || ev == nil || ev.Broadcast || !ev.Implemented ||
		elapsed < gamedata.OriginalEventMinimumTurn(ev.ID) {
		return false
	}
	// sub_2230A 對五個怪獸事件的最早日期分支都另呼叫 sub_233FA；亂流存在時
	// 候選直接失敗。1.31 靜態證據見 random-event-monsters-audit-20260825.md。
	if ev.ID >= 19 && ev.ID <= 23 && s.hasPersistentEvent(PersistentHyperspaceFlux, -1) {
		return false
	}
	return ev.Good || (s.Difficulty != 0 && !luckyForced)
}

// advanceLuckyEventCounter 對應 sub_245C4/sub_24511。標準 remake 尚未分開暴露
// Random Events 與 Antaran Attacks 設定，因此正常對局使用兩者皆開的原版除數 8。
// 原版在第 50 回合前擲骰成功仍會先清零，故回傳值才套回合閘門。
func (s *GameSession) advanceLuckyEventCounter(roll1Based int) bool {
	if s == nil || !s.RaceLucky() {
		return false
	}
	s.LuckyEventCounter++
	divisor := gamedata.OriginalLuckyEventDivisor(true, true)
	if !gamedata.OriginalLuckyEventRollSucceeds(s.LuckyEventCounter, divisor, roll1Based) {
		return false
	}
	s.LuckyEventCounter = 0
	return s.Turn-1 >= 50
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
		applications := originalAncientTechApplications(s.Player, s.ancientTechEmpireStates())
		names := grantAncientTechApplications(&s.Player, applications)
		if len(names) == 0 {
			return eventResult{}, false
		}
		s.UpdatePlayerShipDesignsAfterTech()
		return eventResult{
			Message:   fmt.Sprintf("考古隊從古代異星船艦殘骸中復原科技：%s", ancientTechNames(names)),
			MessageEN: fmt.Sprintf("Archaeologists recovered technology from an ancient alien wreck: %s.", ancientTechNames(names)),
		}, true

	case 1: // 氣候改善:原版直接把 climate 0..7 的目標行星改成 Terran
		i, before, ok := s.applyPlayerClimateEvent()
		if !ok {
			return eventResult{}, false
		}
		return eventResult{
			Message: fmt.Sprintf("%s 的行星軸心突然偏移,環境由 %s 改善為 %s,農民歡聲雷動",
				s.colonyLabel(i), climateDisplayName(before), climateDisplayName(gamedata.TERRAN)),
			MessageEN: fmt.Sprintf("The climate on the affected colony shifted from %s to %s. Its farmers celebrate.",
				climateDisplayNameEN(before), climateDisplayNameEN(gamedata.TERRAN)),
		}, true

	case 2: // 彗星來襲：跨回合由目標星系停泊艦艇攔截，倒數歸零才撞擊
		message, ok := s.startPlayerComet()
		if !ok {
			return eventResult{}, false
		}
		return eventResult{Message: message,
			MessageEN: "A comet is approaching one of the empire's colonies. Ships stationed in the system have begun interception."}, true

	case 3: // 電腦病毒:研究中心電腦全面感染
		if s.Player.ResearchProgress < 10 {
			return eventResult{}, false
		}
		loss, ok := gamedata.OriginalComputerVirusLoss(s.Player.ResearchProgress, s.eventRand.Intn(50)+1)
		if !ok {
			return eventResult{}, false
		}
		s.Player.ResearchProgress -= loss
		return eventResult{
			Message:   fmt.Sprintf("研究中心的電腦遭病毒全面感染,損失 %d 點研究進度", loss),
			MessageEN: fmt.Sprintf("A computer virus infected the research network. The empire lost %d research progress.", loss),
		}, true

	case 4, 5: // 外交暗殺／聯姻：只在交戰帝國間成立，並走原版 Change_Relations_ 的事件切片
		return s.applyDiplomaticIncident(ev.ID, eventEmpireTarget{kind: eventEmpirePlayer, index: 0})

	case 6: // 富商捐獻
		gain, ok := gamedata.OriginalMerchantDonation(s.Turn - 1)
		if !ok {
			return eventResult{}, false
		}
		s.Player.BC += gain
		return eventResult{
			Message:   fmt.Sprintf("一名富商向帝國捐獻了 %d BC", gain),
			MessageEN: fmt.Sprintf("A wealthy merchant donated %d BC to the empire.", gain),
		}, true

	case 7: // 地震:人口與建築受損
		impact, ok := s.applyPlayerEarthquake()
		if !ok {
			return eventResult{}, false
		}
		return eventResult{
			Message: fmt.Sprintf("%s 發生劇烈地震，%d 百萬居民罹難，%d 棟建築毀損",
				impact.ColonyName, impact.PopulationLost, impact.BuildingsDestroyed),
			MessageEN: fmt.Sprintf("A violent earthquake struck %s. %d million residents died and %d buildings were destroyed.",
				impact.ColonyName, impact.PopulationLost, impact.BuildingsDestroyed),
		}, true

	case 8: // 艦船爆炸:一艘軍艦離奇爆炸
		impact, ok := s.resolvePlayerShipExplosion()
		if !ok {
			return eventResult{}, false
		}
		officerZH, officerEN := "", ""
		if impact.OfficerName != "" {
			officerZH = fmt.Sprintf("，艦長 %s 亦不幸罹難", impact.OfficerName)
			officerEN = fmt.Sprintf("; Captain %s was also killed", impact.OfficerName)
		}
		return eventResult{
			Message:   fmt.Sprintf("軍艦「%s」離奇爆炸%s，調查仍在進行中", impact.Lost.Name, officerZH),
			MessageEN: fmt.Sprintf("The warship \"%s\" exploded mysteriously%s. The investigation continues.", impact.Lost.Name, officerEN),
		}, true

	case 9: // 超空間亂流：全銀河非跨維度艦隊暫停航行
		message, ok := s.startHyperspaceFlux()
		if !ok {
			return eventResult{}, false
		}
		return eventResult{Message: message,
			MessageEN: "A hyperspace flux swept across the galaxy; interstellar travel has stalled for non-trans-dimensional fleets."}, true

	case 10: // 工業事故：特殊人口／駐軍傷害後，再結算一點一般殖民地傷害
		impact, ok := s.applyPlayerIndustrialAccident()
		if !ok {
			return eventResult{}, false
		}
		return eventResult{
			Message: fmt.Sprintf("%s 發生重大工業事故，造成 %d 百萬居民、%d 支陸戰隊與 %d 支裝甲部隊傷亡，%d 棟建築毀損",
				impact.ColonyName, impact.PopulationLost, impact.MarinesLost, impact.TanksLost, impact.BuildingsDestroyed),
			MessageEN: fmt.Sprintf("A major industrial accident struck %s: %d million residents, %d marines, and %d armor units were lost; %d buildings were destroyed.",
				impact.ColonyName, impact.PopulationLost, impact.MarinesLost, impact.TanksLost, impact.BuildingsDestroyed),
		}, true

	case 11: // 礦產枯竭:原版只挑 Ultra Rich，並降一級
		i, from, to, ok := s.applyPlayerMineralEvent(11)
		if !ok {
			return eventResult{}, false
		}
		return eventResult{
			Message: fmt.Sprintf("礦工長期超量開採,%s 的礦產等級由「%s」降為「%s」",
				s.colonyLabel(i), mineralDisplayName(from), mineralDisplayName(to)),
			MessageEN: fmt.Sprintf("Long-term overmining depleted a colony's mineral deposits from %s to %s.",
				mineralDisplayNameEN(from), mineralDisplayNameEN(to)),
		}, true

	case 12: // 礦產發現:原版上升兩級，上限 Ultra Rich
		i, from, to, ok := s.applyPlayerMineralEvent(12)
		if !ok {
			return eventResult{}, false
		}
		return eventResult{
			Message: fmt.Sprintf("勘探隊在 %s 發現前所未知的礦脈,礦產等級由「%s」升為「%s」",
				s.colonyLabel(i), mineralDisplayName(from), mineralDisplayName(to)),
			MessageEN: fmt.Sprintf("Explorers found a previously unknown mineral vein. A colony's deposits improved from %s to %s.",
				mineralDisplayNameEN(from), mineralDisplayNameEN(to)),
		}, true

	case 13: // 艦艇叛變：1.31 建立端留下未初始化欄位，consumer 沒有任何移交效果
		return eventResult{
			Message:   "艦隊司令部收到一則未獲證實的艦艇叛變通報",
			MessageEN: "Fleet Command received an unconfirmed report of a ship mutiny.",
		}, true

	case 14: // 海盜活動：跨回合破壞星系內帝國的運輸船，由停泊艦隊共同清剿
		message, ok := s.startPlayerPirateActivity()
		if !ok {
			return eventResult{}, false
		}
		return eventResult{Message: message,
			MessageEN: "Pirate activity erupted in one of the empire's systems. Stationed ships have begun suppression."}, true

	case 15: // 海盜劫掠:國庫被偷
		if s.Player.BC < 100 {
			return eventResult{}, false
		}
		loss, ok := gamedata.OriginalPirateRaidLoss(s.Player.BC, s.eventRand.Intn(21)+1)
		if !ok {
			return eventResult{}, false
		}
		s.Player.BC -= loss
		return eventResult{
			Message:   fmt.Sprintf("海盜突襲得手,自國庫竊走 %d BC", loss),
			MessageEN: fmt.Sprintf("Pirates raided the treasury and stole %d BC.", loss),
		}, true

	case 16: // 瘟疫:持續負成長，研究完成後解除
		i, need, ok := s.startPlayerPlague()
		if !ok {
			return eventResult{}, false
		}
		return eventResult{
			Message: fmt.Sprintf("%s 爆發瘟疫，人口成長急遽惡化；研究團隊需要累積 %d RP 研發疫苗",
				s.colonyLabel(i), need),
			MessageEN: fmt.Sprintf("A plague struck a colony. Its researchers need %d RP to develop a cure.", need),
		}, true

	case 17: // 人口暴增:持續型成長加成，而非一次性加人口
		i, ok := s.startPlayerPopulationBoom()
		if !ok {
			return eventResult{}, false
		}
		return eventResult{
			Message:   fmt.Sprintf("%s 出現人口暴增，新生兒數量翻倍；成長潮將持續數回合", s.colonyLabel(i)),
			MessageEN: "A population boom doubled the number of births at a colony; the growth surge will continue for several turns.",
		}, true

	case 18: // 秘密實驗:立即完成目前研究 field，然後清空 field 與累積 RP
		topic, completed := s.completePlayerSecretExperiment()
		if !completed {
			return eventResult{
				Message:   "秘密實驗結束，但目前沒有可完成的研究領域",
				MessageEN: "The secret experiment ended, but there was no active research field to complete.",
			}, true
		}
		name := ResearchTopicName(topic)
		return eventResult{
			Message:   fmt.Sprintf("科學家在秘密實驗中取得突破，立即完成研究領域：%s", name),
			MessageEN: fmt.Sprintf("A secret experiment completed the research field: %s.", name),
		}, true

	// --- 太空怪獸入侵(19-23)。IDA sub_2230A/sub_23BEC 證實最早回合與受害帝國
	//     殖民星 reservoir sampling；怪獸實體與戰鬥見 monster.go。---
	case 19: // 太空變形蟲(手冊:≥100 回合)
		message, ok := s.spawnInvadingMonster(gamedata.MonsterAmoeba, 100)
		return wrapEventResult(s, ev, message, ok)
	case 20: // 太空水晶(手冊:≥200 回合)
		message, ok := s.spawnInvadingMonster(gamedata.MonsterCrystal, 200)
		return wrapEventResult(s, ev, message, ok)
	case 21: // 太空巨龍(手冊:≥300 回合)
		message, ok := s.spawnInvadingMonster(gamedata.MonsterDragon, 300)
		return wrapEventResult(s, ev, message, ok)
	case 22: // 太空鰻(≥150 回合，原版 type 13 獨立 loader)
		message, ok := s.spawnInvadingMonster(gamedata.MonsterEel, 150)
		return wrapEventResult(s, ev, message, ok)
	case 23: // 太空九頭蛇(手冊:≥250 回合)
		message, ok := s.spawnInvadingMonster(gamedata.MonsterHydra, 250)
		return wrapEventResult(s, ev, message, ok)

	// --- 持續型事件(24-28),見 events_persistent.go ---
	case 24: // 超新星（1.31：elapsed≥200，倒數 Random(5)+10-difficulty）
		message, ok := s.startSupernova()
		return wrapEventResult(s, ev, message, ok)
	case 25: // 時空異象:整個星系凍結
		message, ok := s.startStasis()
		return wrapEventResult(s, ev, message, ok)
	case 26: // 超空間獸:航行中的艦隊有機率損失艦艇
		message, ok := s.startWarpBeast(eventEmpireTarget{kind: eventEmpirePlayer, index: 0, alive: true})
		return wrapEventResult(s, ev, message, ok)
	case 27: // 曲速漏斗：1.31 是帶目標艦索引的報告型 persistent event
		message, ok := s.startWarpFunnel()
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
	case 9:
		messageEN = "A hyperspace flux swept across the galaxy; non-trans-dimensional fleets can no longer travel until it dissipates."
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
	case 27:
		messageEN = "A warp funnel trapped one fleet in hyperspace. Its crews are attempting to break free."
	case 28:
		messageEN = "A wormhole appeared along the fleet's route and carried it immediately to its destination."
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
			engine.RemovePopulationGroupUnit(c, gamedata.WORKER)
			c.Workers--
		case c.Farmers >= c.Scientists && c.Farmers > 0:
			engine.RemovePopulationGroupUnit(c, gamedata.FARMER)
			c.Farmers--
		case c.Scientists > 0:
			engine.RemovePopulationGroupUnit(c, gamedata.SCIENTIST)
			c.Scientists--
		}
	}
	return lost
}

// destroyRandomBuilding 隨機摧毀殖民地 i 的一棟建築,回傳建築名(沒有可摧毀的回空字串)。
// 首都不屬一般 ColonyBuildings 槽；若舊匯入資料帶有同名 key，也不把帝國固有狀態當成
// 可隨機摧毀建築。`sub_23DFE` 已證實是事件殖民地 filter，不再引用它作 No-Capitol 證據。
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
