package shell

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

type eventEmpireKind uint8

const (
	eventEmpirePlayer eventEmpireKind = iota
	eventEmpireSeat
	eventEmpireAI
)

func (k eventEmpireKind) String() string {
	switch k {
	case eventEmpireSeat:
		return "seat"
	case eventEmpireAI:
		return "ai"
	default:
		return "player"
	}
}

type eventEmpireTarget struct {
	kind       eventEmpireKind
	index      int
	population int
	alive      bool
	lucky      bool
}

func (s *GameSession) eventEmpireTargetAtStar(star int) (eventEmpireTarget, bool) {
	if colonyStarsContain(s.PlayerColonyStars, star) {
		if s.HotseatEnabled() {
			return eventEmpireTarget{kind: eventEmpireSeat, index: s.ActiveSeat, alive: true}, true
		}
		return eventEmpireTarget{kind: eventEmpirePlayer, index: 0, alive: true}, true
	}
	if s.HotseatEnabled() {
		for i := range s.Seats {
			if i != s.ActiveSeat && colonyStarsContain(s.Seats[i].PlayerColonyStars, star) {
				return eventEmpireTarget{kind: eventEmpireSeat, index: i, alive: true}, true
			}
		}
	}
	for i := range s.AIPlayers {
		if colonyStarsContain(s.AIPlayers[i].ColonyStars, star) {
			return eventEmpireTarget{kind: eventEmpireAI, index: i, alive: true}, true
		}
	}
	return eventEmpireTarget{}, false
}

func colonyPopulationTotal(colonies []engine.ColonyState) int {
	total := 0
	for i := range colonies {
		if colonies[i].Population > 0 {
			total += colonies[i].Population
		}
	}
	return total
}

func seatHasTrait(v seat, trait gamedata.RaceTrait) bool {
	if v.RaceIndex < 0 || v.RaceIndex >= len(Races) {
		return trait >= gamedata.TRAIT_LOW_G && trait <= gamedata.TRAIT_POOR_HOMEWORLD &&
			v.CustomRaceTraits&(uint32(1)<<uint(trait)) != 0
	}
	return gamedata.OrigRaceHasTrait(Races[v.RaceIndex].OrigIdx, trait)
}

// eventEmpireTargets 把 remake 的「目前玩家＋熱座快照＋AI」投影成原版 player[] 槽。
// 熱座目前席位先同步回快照，避免人口仍停在上次換人時。
func (s *GameSession) eventEmpireTargets() []eventEmpireTarget {
	if s.HotseatEnabled() {
		if s.ActiveSeat >= 0 && s.ActiveSeat < len(s.Seats) {
			s.Seats[s.ActiveSeat] = s.saveSeat()
		}
		out := make([]eventEmpireTarget, 0, len(s.Seats)+len(s.AIPlayers))
		for i := range s.Seats {
			pop := colonyPopulationTotal(s.Seats[i].PlayerColonies)
			out = append(out, eventEmpireTarget{kind: eventEmpireSeat, index: i,
				population: pop, alive: pop > 0, lucky: seatHasTrait(s.Seats[i], gamedata.TRAIT_LUCKY)})
		}
		for i := range s.AIPlayers {
			pop := colonyPopulationTotal(s.AIPlayers[i].Colonies)
			out = append(out, eventEmpireTarget{kind: eventEmpireAI, index: i,
				population: pop, alive: pop > 0, lucky: aiRaceHasTrait(s.AIPlayers[i], gamedata.TRAIT_LUCKY)})
		}
		return out
	}
	pop := colonyPopulationTotal(s.PlayerColonies)
	out := []eventEmpireTarget{{kind: eventEmpirePlayer, index: 0, population: pop,
		alive: pop > 0, lucky: s.RaceLucky()}}
	for i := range s.AIPlayers {
		aiPop := colonyPopulationTotal(s.AIPlayers[i].Colonies)
		out = append(out, eventEmpireTarget{kind: eventEmpireAI, index: i,
			population: aiPop, alive: aiPop > 0, lucky: aiRaceHasTrait(s.AIPlayers[i], gamedata.TRAIT_LUCKY)})
	}
	return out
}

func (s *GameSession) luckyCounter(target eventEmpireTarget) int {
	switch target.kind {
	case eventEmpireSeat:
		if target.index >= 0 && target.index < len(s.Seats) {
			return s.Seats[target.index].LuckyEventCounter
		}
	case eventEmpireAI:
		if target.index >= 0 && target.index < len(s.AIPlayers) {
			return s.AIPlayers[target.index].LuckyEventCounter
		}
	default:
		return s.LuckyEventCounter
	}
	return 0
}

func (s *GameSession) setLuckyCounter(target eventEmpireTarget, value int) {
	switch target.kind {
	case eventEmpireSeat:
		if target.index >= 0 && target.index < len(s.Seats) {
			s.Seats[target.index].LuckyEventCounter = value
			if target.index == s.ActiveSeat {
				s.LuckyEventCounter = value
			}
		}
	case eventEmpireAI:
		if target.index >= 0 && target.index < len(s.AIPlayers) {
			s.AIPlayers[target.index].LuckyEventCounter = value
		}
	default:
		s.LuckyEventCounter = value
	}
}

// advanceAllLuckyEventCounters 對應 sub_245C4 後接 sub_24511：先全部累加，再按槽序
// 擲骰，第一個成功槽停止。成功在 50 回合閘門前仍清零。
func (s *GameSession) advanceAllLuckyEventCounters(elapsed int) (eventEmpireTarget, bool) {
	targets := s.eventEmpireTargets()
	for _, target := range targets {
		if target.alive && target.lucky {
			s.setLuckyCounter(target, s.luckyCounter(target)+1)
		}
	}
	divisor := gamedata.OriginalLuckyEventDivisor(true, true)
	for _, target := range targets {
		if !target.alive || !target.lucky {
			continue
		}
		roll := s.eventRand.Intn(1000) + 1
		if !gamedata.OriginalLuckyEventRollSucceeds(s.luckyCounter(target), divisor, roll) {
			continue
		}
		s.setLuckyCounter(target, 0)
		return target, elapsed >= 50
	}
	return eventEmpireTarget{}, false
}

func (s *GameSession) chooseEventEmpireTarget(ev gamedata.RandomEvent, forced eventEmpireTarget, luckyForced bool) (eventEmpireTarget, bool) {
	if luckyForced {
		return forced, true
	}
	// 原版事件 9 是不帶帝國目標的全銀河 record；事件 24 由專用 helper 決定目標，兩者
	// 都不走 sub_22D57。事件 24 尚未保存完整專用目標 record，先維持目前載入帝國的效果。
	if ev.ID == 9 || ev.ID == 24 {
		if s.HotseatEnabled() {
			return eventEmpireTarget{kind: eventEmpireSeat, index: s.ActiveSeat, alive: true}, true
		}
		return eventEmpireTarget{kind: eventEmpirePlayer, index: 0, alive: true}, true
	}
	target, ok := s.chooseEventEmpireTargetByWeight(ev)
	if !ok || (!ev.Good && target.lucky) {
		return eventEmpireTarget{}, false
	}
	return target, true
}

// chooseEventEmpireTargetByWeight 只重建 sub_22D57 的人口權重抽選，不套建立事件外圈的
// Lucky 壞事件取消。事件 26 已 active 後每回合重抽下一個帝國時直接呼叫 sub_22D57，
// 因此不能再次把 Lucky 帝國當成「候選事件建立失敗」。
func (s *GameSession) chooseEventEmpireTargetByWeight(ev gamedata.RandomEvent) (eventEmpireTarget, bool) {
	targets := s.eventEmpireTargets()
	populations := make([]int, len(targets))
	eligible := make([]bool, len(targets))
	for i := range targets {
		populations[i], eligible[i] = targets[i].population, targets[i].alive
	}
	indices, weights := gamedata.OriginalEventVictimWeights(populations, eligible, ev.Good)
	if len(indices) == 0 {
		return eventEmpireTarget{}, false
	}
	_, normalized, ok := gamedata.OriginalEventWeightedChoice(weights, 0)
	if !ok {
		return eventEmpireTarget{}, false
	}
	total := 0
	for _, weight := range normalized {
		if weight > 0 {
			total += weight
		}
	}
	s.eventRandForTest()
	choice, _, ok := gamedata.OriginalEventWeightedChoice(weights, s.eventRand.Intn(total))
	if !ok || choice < 0 || choice >= len(indices) || indices[choice] < 0 || indices[choice] >= len(targets) {
		return eventEmpireTarget{}, false
	}
	return targets[indices[choice]], true
}

func (s *GameSession) applyRandomEventLocalizedToTarget(ev gamedata.RandomEvent, target eventEmpireTarget) (eventResult, bool) {
	// 事件 26 的 record 必須保存原版通用受害帝國槽。若先切換熱座再走一般玩家
	// dispatcher，會把非目前席位錯記為 player 0，下一回合便攻擊錯帝國。
	if ev.ID == 26 {
		message, ok := s.startWarpBeast(target)
		return wrapEventResult(s, ev, message, ok)
	}
	switch target.kind {
	case eventEmpireAI:
		return s.applyRandomEventLocalizedToAI(ev, target.index)
	case eventEmpireSeat:
		if target.index < 0 || target.index >= len(s.Seats) {
			return eventResult{}, false
		}
		active := s.ActiveSeat
		if active >= 0 && active < len(s.Seats) {
			s.Seats[active] = s.saveSeat()
		}
		if target.index != active {
			s.loadSeat(s.Seats[target.index])
		}
		result, ok := s.applyRandomEventLocalized(ev)
		s.Seats[target.index] = s.saveSeat()
		if target.index != active && active >= 0 && active < len(s.Seats) {
			s.loadSeat(s.Seats[active])
		}
		return result, ok
	default:
		return s.applyRandomEventLocalized(ev)
	}
}

// applyRandomEventLocalizedToAI 只接 AIOpponent 已有一對一欄位的事件。需要殖民地
// 平行陣列、艦隊轉移或外交雙邊 record 的事件仍回 false，讓候選誠實失敗。
func (s *GameSession) applyRandomEventLocalizedToAI(ev gamedata.RandomEvent, aiIndex int) (eventResult, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return eventResult{}, false
	}
	a := &s.AIPlayers[aiIndex]
	name := stripAILabel(a.Name)
	switch ev.ID {
	case 4, 5:
		return s.applyDiplomaticIncident(ev.ID, eventEmpireTarget{kind: eventEmpireAI, index: aiIndex})
	case 0:
		applications := originalAncientTechApplications(a.Player, s.ancientTechEmpireStates())
		names := grantAncientTechApplications(&a.Player, applications)
		if len(names) == 0 {
			return eventResult{}, false
		}
		s.updateAIShipDesignsAfterTech(aiIndex)
		return eventResult{Message: fmt.Sprintf("%s 的考古隊從古代異星船艦殘骸復原科技：%s", name, ancientTechNames(names)),
			MessageEN: fmt.Sprintf("Archaeologists of %s recovered technology from an ancient alien wreck: %s.", name, ancientTechNames(names))}, true
	case 1:
		_, before, ok := s.applyAIClimateEvent(aiIndex)
		if !ok {
			return eventResult{}, false
		}
		return eventResult{Message: fmt.Sprintf("%s 的一座殖民地氣候由 %s 改善為 %s", name,
			climateDisplayName(before), climateDisplayName(gamedata.TERRAN)),
			MessageEN: fmt.Sprintf("The climate of a %s colony shifted from %s to %s.", name,
				climateDisplayNameEN(before), climateDisplayNameEN(gamedata.TERRAN))}, true
	case 2:
		message, ok := s.startAIComet(aiIndex)
		if !ok {
			return eventResult{}, false
		}
		return eventResult{Message: message,
			MessageEN: fmt.Sprintf("A comet is approaching a colony of %s. Ships stationed in the system have begun interception.", name)}, true
	case 14:
		message, ok := s.startAIPirateActivity(aiIndex)
		if !ok {
			return eventResult{}, false
		}
		return eventResult{Message: message,
			MessageEN: fmt.Sprintf("Pirate activity erupted in a system controlled by %s.", name)}, true
	case 3:
		if a.Player.ResearchProgress < 10 {
			return eventResult{}, false
		}
		loss, ok := gamedata.OriginalComputerVirusLoss(a.Player.ResearchProgress, s.eventRand.Intn(50)+1)
		if !ok {
			return eventResult{}, false
		}
		a.Player.ResearchProgress -= loss
		return eventResult{Message: fmt.Sprintf("%s 的研究網路感染電腦病毒，損失 %d RP", name, loss),
			MessageEN: fmt.Sprintf("A computer virus cost %s %d research progress.", name, loss)}, true
	case 6:
		gain, ok := gamedata.OriginalMerchantDonation(s.Turn - 1)
		if !ok {
			return eventResult{}, false
		}
		a.Player.BC += gain
		return eventResult{Message: fmt.Sprintf("一名富商向%s捐獻了 %d BC", name, gain),
			MessageEN: fmt.Sprintf("A wealthy merchant donated %d BC to %s.", gain, name)}, true
	case 7:
		impact, ok := s.applyAIEarthquake(aiIndex)
		if !ok {
			return eventResult{}, false
		}
		return eventResult{Message: fmt.Sprintf("%s 的一座殖民地發生劇烈地震，%d 百萬居民罹難，%d 棟建築毀損",
			name, impact.PopulationLost, impact.BuildingsDestroyed),
			MessageEN: fmt.Sprintf("A violent earthquake struck a %s colony. %d million residents died and %d buildings were destroyed.",
				name, impact.PopulationLost, impact.BuildingsDestroyed)}, true
	case 8:
		impact, ok := s.resolveAIShipExplosion(aiIndex)
		if !ok {
			return eventResult{}, false
		}
		officerZH, officerEN := "", ""
		if impact.OfficerName != "" {
			officerZH = fmt.Sprintf("，艦長 %s 亦不幸罹難", impact.OfficerName)
			officerEN = fmt.Sprintf("; Captain %s was also killed", impact.OfficerName)
		}
		return eventResult{Message: fmt.Sprintf("%s 的軍艦「%s」離奇爆炸%s", name, impact.Lost.Name, officerZH),
			MessageEN: fmt.Sprintf("The %s warship \"%s\" exploded mysteriously%s.", name, impact.Lost.Name, officerEN)}, true
	case 13:
		return eventResult{Message: fmt.Sprintf("艦隊司令部收到一則關於%s艦艇叛變的未獲證實通報", name),
			MessageEN: fmt.Sprintf("Fleet Command received an unconfirmed report of a mutiny aboard a %s ship.", name)}, true
	case 10:
		impact, ok := s.applyAIIndustrialAccident(aiIndex)
		if !ok {
			return eventResult{}, false
		}
		return eventResult{Message: fmt.Sprintf("%s 的一座殖民地發生重大工業事故，造成 %d 百萬居民、%d 支陸戰隊與 %d 支裝甲部隊傷亡，%d 棟建築毀損",
			name, impact.PopulationLost, impact.MarinesLost, impact.TanksLost, impact.BuildingsDestroyed),
			MessageEN: fmt.Sprintf("A major industrial accident struck a %s colony: %d million residents, %d marines, and %d armor units were lost; %d buildings were destroyed.",
				name, impact.PopulationLost, impact.MarinesLost, impact.TanksLost, impact.BuildingsDestroyed)}, true
	case 11, 12:
		_, from, to, ok := s.applyAIMineralEvent(aiIndex, ev.ID)
		if !ok {
			return eventResult{}, false
		}
		if ev.ID == 11 {
			return eventResult{Message: fmt.Sprintf("%s 的一座殖民地因過度開採，礦產由「%s」降為「%s」", name,
				mineralDisplayName(from), mineralDisplayName(to)),
				MessageEN: fmt.Sprintf("Overmining depleted a %s colony from %s to %s.", name,
					mineralDisplayNameEN(from), mineralDisplayNameEN(to))}, true
		}
		return eventResult{Message: fmt.Sprintf("%s 的一座殖民地發現新礦脈，礦產由「%s」升為「%s」", name,
			mineralDisplayName(from), mineralDisplayName(to)),
			MessageEN: fmt.Sprintf("A new mineral vein improved a %s colony from %s to %s.", name,
				mineralDisplayNameEN(from), mineralDisplayNameEN(to))}, true
	case 15:
		if a.Player.BC < 100 {
			return eventResult{}, false
		}
		loss, ok := gamedata.OriginalPirateRaidLoss(a.Player.BC, s.eventRand.Intn(21)+1)
		if !ok {
			return eventResult{}, false
		}
		a.Player.BC -= loss
		return eventResult{Message: fmt.Sprintf("海盜自%s的國庫竊走 %d BC", name, loss),
			MessageEN: fmt.Sprintf("Pirates stole %d BC from %s.", loss, name)}, true
	case 16:
		_, need, ok := s.startAIPlague(aiIndex)
		if !ok {
			return eventResult{}, false
		}
		return eventResult{Message: fmt.Sprintf("%s 的一座殖民地爆發瘟疫，需要累積 %d RP 研發疫苗", name, need),
			MessageEN: fmt.Sprintf("A plague struck a %s colony; its researchers need %d RP for a cure.", name, need)}, true
	case 17:
		if _, ok := s.startAIPopulationBoom(aiIndex); !ok {
			return eventResult{}, false
		}
		return eventResult{Message: fmt.Sprintf("%s 的一座殖民地出現人口暴增，出生率將在數回合內提高", name),
			MessageEN: fmt.Sprintf("A population boom raised the birth rate at a %s colony for several turns.", name)}, true
	case 18:
		topic, completed := s.completeAISecretExperiment(aiIndex)
		if !completed {
			return eventResult{Message: fmt.Sprintf("%s 的秘密實驗結束，但目前沒有可完成的研究領域", name),
				MessageEN: fmt.Sprintf("The secret experiment of %s ended without an active research field.", name)}, true
		}
		topicName := ResearchTopicName(topic)
		return eventResult{Message: fmt.Sprintf("%s 的秘密實驗取得突破，立即完成研究領域：%s", name, topicName),
			MessageEN: fmt.Sprintf("A secret experiment of %s completed the research field: %s.", name, topicName)}, true
	case 19, 20, 21, 22, 23:
		kind, ok := eventMonsterKind(ev.ID)
		if !ok {
			return eventResult{}, false
		}
		message, ok := s.spawnInvadingMonsterForAI(kind, aiIndex)
		if !ok {
			return eventResult{}, false
		}
		return eventResult{Message: message,
			MessageEN: fmt.Sprintf("A %s invaded a colony system of %s.", eventNameEN(ev.ID), name)}, true
	case 25:
		message, ok := s.startAIStasis(aiIndex)
		if !ok {
			return eventResult{}, false
		}
		return eventResult{Message: message,
			MessageEN: fmt.Sprintf("A space-time anomaly froze a colony system of %s.", name)}, true
	case 26:
		message, ok := s.startWarpBeast(eventEmpireTarget{kind: eventEmpireAI, index: aiIndex, alive: true})
		return wrapEventResult(s, ev, message, ok)
	case 27:
		if a.FleetStrength <= 0 || s.hasPersistentEvent(PersistentWarpFunnel, -1) {
			return eventResult{}, false
		}
		s.PersistentEvents = append(s.PersistentEvents, PersistentEvent{
			Kind: PersistentWarpFunnel, StarIndex: -1, Turns: -1,
		})
		return eventResult{
			Message:   fmt.Sprintf("%s 的一支艦隊被曲速漏斗困在超空間中", name),
			MessageEN: fmt.Sprintf("A fleet of %s was trapped in hyperspace by a warp funnel.", name),
		}, true
	case 28:
		message, ok := s.applyAIWormhole(aiIndex)
		if !ok {
			return eventResult{}, false
		}
		return eventResult{Message: message,
			MessageEN: fmt.Sprintf("A wormhole carried a fleet of %s directly to its destination.", name)}, true
	}
	return eventResult{}, false
}

func (s *GameSession) eventEmpireTargetName(target eventEmpireTarget) string {
	switch target.kind {
	case eventEmpireSeat:
		return s.SeatName(target.index)
	case eventEmpireAI:
		if target.index >= 0 && target.index < len(s.AIPlayers) {
			return stripAILabel(s.AIPlayers[target.index].Name)
		}
	default:
		if s.PlayerName != "" {
			return s.PlayerName
		}
	}
	return "未知帝國"
}

func cloneEventReport(report *EventReport) *EventReport {
	if report == nil {
		return nil
	}
	copy := *report
	return &copy
}

func (s *GameSession) clearEventReportsForAllSeats() {
	s.LastEvent = ""
	s.LastEventReport = nil
	for i := range s.Seats {
		s.Seats[i].LastEvent = ""
		s.Seats[i].LastEventReport = nil
	}
}

func (s *GameSession) broadcastEventReport(report *EventReport) {
	for i := range s.Seats {
		s.Seats[i].LastEvent = report.Message
		s.Seats[i].LastEventReport = cloneEventReport(report)
	}
}
