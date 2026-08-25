package shell

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// EmpireSurrender 對應原版 player+0xE72 的延後投降關係。原版 setter 先建立
// GNN 34，下一個 sub_E4DC9 consumer 才搬資產，因此不能直接在新聞函式內修改帝國。
type EmpireSurrender struct {
	SurrenderAI   int    `json:"surrenderAI"`
	ReceiverKind  string `json:"receiverKind"`
	ReceiverIndex int    `json:"receiverIndex"`
}

func (s *GameSession) hasPendingSurrender(aiIndex int) bool {
	for _, pending := range s.PendingSurrenders {
		if pending.SurrenderAI == aiIndex {
			return true
		}
	}
	return false
}

func (s *GameSession) queueEmpireSurrender(aiIndex int, receiver eventEmpireTarget) bool {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) || len(s.AIPlayers[aiIndex].Colonies) == 0 ||
		!receiver.alive || (receiver.kind == eventEmpireAI && receiver.index == aiIndex) ||
		s.hasPendingSurrender(aiIndex) {
		return false
	}
	pending := EmpireSurrender{SurrenderAI: aiIndex, ReceiverKind: receiver.kind.String(), ReceiverIndex: receiver.index}
	s.PendingSurrenders = append(s.PendingSurrenders, pending)
	surrenderer := eventEmpireTarget{kind: eventEmpireAI, index: aiIndex, alive: true}
	from, to := s.eventEmpireTargetName(surrenderer), s.eventEmpireTargetName(receiver)
	report := s.statusTargetReport(34, surrenderer,
		fmt.Sprintf("%s 已向 %s 無條件投降；帝國接收程序即將開始", from, to),
		fmt.Sprintf("%s has surrendered unconditionally to %s; imperial absorption will now begin.", from, to))
	report.TargetName = from
	report.SecondaryTargetKind = receiver.kind.String()
	report.SecondaryTargetIndex = receiver.index
	report.SecondaryTargetName = to
	s.queueStatusBroadcast(report)
	return true
}

func surrenderPower(a AIOpponent) int {
	power := maxInt(0, a.FleetStrength) + maxInt(0, a.Player.BC)/10
	for _, colony := range a.Colonies {
		power += maxInt(0, colony.Population) * 4
	}
	return maxInt(1, power)
}

func (s *GameSession) playerSurrenderPower() int {
	power := maxInt(0, s.Player.BC) / 10
	power += len(s.AllShips()) * 10
	for _, colony := range s.PlayerColonies {
		power += maxInt(0, colony.Population) * 4
	}
	return maxInt(1, power)
}

// advanceAISurrenders 的接收資產契約已由 IDA 閉合；自動觸發仍是明示近似，
// 因 raw player+0x717 與 sub_27A3D 的完整評分欄位尚未定名。
func (s *GameSession) advanceAISurrenders() {
	if !s.EnableAIVsAI {
		return
	}
	s.ensureAIAIState()
	for loser := range s.AIPlayers {
		if len(s.AIPlayers[loser].Colonies) == 0 || s.hasPendingSurrender(loser) {
			continue
		}
		loserPower := surrenderPower(s.AIPlayers[loser])
		best := eventEmpireTarget{}
		bestPower := 0
		if s.AIPlayers[loser].Treaty.FormalPolicy >= gamedata.DIPLO_LIMITED_WAR {
			best = eventEmpireTarget{kind: eventEmpirePlayer, index: 0, alive: len(s.PlayerColonies) > 0}
			if s.HotseatEnabled() {
				best.kind, best.index = eventEmpireSeat, s.ActiveSeat
			}
			bestPower = s.playerSurrenderPower()
		}
		for receiver := range s.AIPlayers {
			if receiver == loser || len(s.AIPlayers[receiver].Colonies) == 0 ||
				loser >= len(s.AIWars) || receiver >= len(s.AIWars[loser]) || !s.AIWars[loser][receiver] {
				continue
			}
			power := surrenderPower(s.AIPlayers[receiver])
			if power > bestPower {
				best = eventEmpireTarget{kind: eventEmpireAI, index: receiver, alive: true}
				bestPower = power
			}
		}
		// 保守的「顯著弱勢」gate；避免近似評分讓勢均力敵的 AI 無故退出。
		if best.alive && bestPower >= loserPower*4 {
			s.queueEmpireSurrender(loser, best)
		}
	}
}

func surrenderReceiverTarget(p EmpireSurrender) (eventEmpireTarget, bool) {
	kind, ok := eventEmpireKindFromString(p.ReceiverKind)
	if !ok {
		return eventEmpireTarget{}, false
	}
	return eventEmpireTarget{kind: kind, index: p.ReceiverIndex}, true
}

func eventEmpireKindFromString(value string) (eventEmpireKind, bool) {
	for _, kind := range []eventEmpireKind{eventEmpirePlayer, eventEmpireSeat, eventEmpireAI} {
		if kind.String() == value {
			return kind, true
		}
	}
	return eventEmpirePlayer, false
}

func mergeSurrenderTechnology(receiver, loser *engine.PlayerState) {
	if receiver == nil || loser == nil {
		return
	}
	// sub_E4B5F 明確只掃 0x53（83）個 application，不擴成整張 212-entry 表。
	for raw := 0; raw < 0x53; raw++ {
		tech := gamedata.Technology(raw)
		topic, ok := gamedata.OrigTechTopic(tech)
		if ok && playerStateKnowsTech(*loser, topic, tech) && !playerStateKnowsTech(*receiver, topic, tech) {
			grantTechnologyApplication(receiver, topic, tech)
		}
	}
}

func unassignSurrenderedLeaders(leaders []Leader, receiverSlot int) []Leader {
	out := append([]Leader(nil), leaders...)
	for i := range out {
		out[i].RawETA = 0
		out[i].RawLocation = -1
		out[i].RawPlayerIndex = receiverSlot
	}
	return out
}

func (s *GameSession) resolvePendingSurrenders() {
	pending := append([]EmpireSurrender(nil), s.PendingSurrenders...)
	s.PendingSurrenders = nil
	for _, item := range pending {
		receiver, ok := surrenderReceiverTarget(item)
		if !ok || item.SurrenderAI < 0 || item.SurrenderAI >= len(s.AIPlayers) ||
			len(s.AIPlayers[item.SurrenderAI].Colonies) == 0 {
			continue
		}
		receiver.alive = s.eventEmpireTargetAlive(receiver)
		if !receiver.alive || (receiver.kind == eventEmpireAI && receiver.index == item.SurrenderAI) {
			continue
		}
		s.resolveEmpireSurrender(item.SurrenderAI, receiver)
	}
}

func (s *GameSession) eventEmpireTargetAlive(target eventEmpireTarget) bool {
	switch target.kind {
	case eventEmpirePlayer:
		return !s.HotseatEnabled() && len(s.PlayerColonies) > 0
	case eventEmpireSeat:
		return target.index >= 0 && target.index < len(s.Seats) && len(s.Seats[target.index].PlayerColonies) > 0
	case eventEmpireAI:
		return target.index >= 0 && target.index < len(s.AIPlayers) && len(s.AIPlayers[target.index].Colonies) > 0
	default:
		return false
	}
}

func (s *GameSession) resolveEmpireSurrender(loserIndex int, receiver eventEmpireTarget) {
	loser := &s.AIPlayers[loserIndex]
	switch receiver.kind {
	case eventEmpireAI:
		s.transferSurrenderToAI(loserIndex, receiver.index)
	case eventEmpirePlayer:
		s.transferSurrenderToPlayer(loserIndex)
	case eventEmpireSeat:
		if receiver.index == s.ActiveSeat {
			s.transferSurrenderToPlayer(loserIndex)
		} else {
			active := s.ActiveSeat
			if active < 0 || active >= len(s.Seats) || receiver.index < 0 || receiver.index >= len(s.Seats) {
				return
			}
			s.Seats[active] = s.saveSeat()
			s.loadSeat(s.Seats[receiver.index])
			s.transferSurrenderToPlayer(loserIndex)
			s.Seats[receiver.index] = s.saveSeat()
			s.loadSeat(s.Seats[active])
		}
	}
	loser.Ships = nil
	loser.ShipDesigns = nil
	loser.ShipBuildProgress = 0
	loser.FleetStrength, loser.FleetInvestPool = 0, 0
	loser.FleetETA, loser.FleetDestStar, loser.FleetStar = 0, 0, 0
	loser.FleetTargetAI, loser.FleetTargetAISet = -1, false
	loser.Colonies, loser.ColonyStars, loser.ColonyPlanets, loser.ColonyBuildings = nil, nil, nil, nil
	loser.ColonyMarines, loser.ColonyTanks = nil, nil
	loser.MarineBarracksAge, loser.ArmorBarracksAge = nil, nil
	loser.ColonyLeaderNames, loser.Leaders = nil, nil
	loser.LeaderOffer = nil
	loser.Player.BC, loser.Player.ActiveFreighters = 0, 0
	loser.OwnedStars, loser.Spies, loser.DefensiveAgents = 0, 0, 0
	loser.Treaty = TreatyState{}
	s.clearSurrenderDiplomacy(loserIndex, receiver)
	s.markSurrenderedEmpireInactive(loserIndex)
}

// 投降有自己的事件 34；不可讓同一個殖民地歸零轉換又被事件 29 當成軍事滅亡。
func (s *GameSession) markSurrenderedEmpireInactive(aiIndex int) {
	targets := s.eventEmpireTargets()
	if len(s.StatusBroadcast.EmpireAlive) != len(targets) {
		return
	}
	for i, target := range targets {
		if target.kind == eventEmpireAI && target.index == aiIndex {
			s.StatusBroadcast.EmpireAlive[i] = false
			return
		}
	}
}

func (s *GameSession) transferSurrenderToAI(loserIndex, receiverIndex int) {
	loser, receiver := &s.AIPlayers[loserIndex], &s.AIPlayers[receiverIndex]
	ensureAIGroundForceSlots(loser)
	ensureAIGroundForceSlots(receiver)
	for i, colony := range loser.Colonies {
		receiver.Colonies = append(receiver.Colonies, colony)
		receiver.ColonyStars = append(receiver.ColonyStars, parallelInt(loser.ColonyStars, i, -1))
		receiver.ColonyPlanets = append(receiver.ColonyPlanets, parallelInt(loser.ColonyPlanets, i, -1))
		receiver.ColonyBuildings = append(receiver.ColonyBuildings, cloneBuildings(parallelBuildings(loser.ColonyBuildings, i)))
		receiver.ColonyMarines = append(receiver.ColonyMarines, parallelInt(loser.ColonyMarines, i, 0))
		receiver.ColonyTanks = append(receiver.ColonyTanks, parallelInt(loser.ColonyTanks, i, 0))
		receiver.MarineBarracksAge = append(receiver.MarineBarracksAge, parallelInt(loser.MarineBarracksAge, i, 0))
		receiver.ArmorBarracksAge = append(receiver.ArmorBarracksAge, parallelInt(loser.ArmorBarracksAge, i, 0))
	}
	mergeSurrenderTechnology(&receiver.Player, &loser.Player)
	receiver.Player.BC += loser.Player.BC
	receiver.Player.ActiveFreighters += loser.Player.ActiveFreighters
	receiver.Leaders = append(receiver.Leaders, unassignSurrenderedLeaders(loser.Leaders, receiver.PopulationRaceSlot)...)
	receiver.ColonyLeaderNames = append(receiver.ColonyLeaderNames, make([]string, len(loser.Colonies))...)
	receiver.OwnedStars = len(receiver.ColonyStars)
	s.syncAIRaceEngineFields(receiver)
	s.updateAIShipDesignsAfterTech(receiverIndex)
}

func (s *GameSession) transferSurrenderToPlayer(loserIndex int) {
	loser := &s.AIPlayers[loserIndex]
	ensureAIGroundForceSlots(loser)
	for i, colony := range loser.Colonies {
		star := parallelInt(loser.ColonyStars, i, -1)
		s.appendPlayerColony(colony, star, parallelInt(loser.ColonyPlanets, i, -1))
		idx := len(s.PlayerColonies) - 1
		s.ColonyBuildings[idx] = cloneBuildings(parallelBuildings(loser.ColonyBuildings, i))
		s.PlayerColonyMarines[idx] = parallelInt(loser.ColonyMarines, i, 0)
		s.PlayerColonyTanks[idx] = parallelInt(loser.ColonyTanks, i, 0)
		s.MarineBarracksAge[idx] = parallelInt(loser.MarineBarracksAge, i, 0)
		s.ArmorBarracksAge[idx] = parallelInt(loser.ArmorBarracksAge, i, 0)
		if star >= 0 && star < len(s.Stars) {
			s.Stars[star].Owner = 1
		}
	}
	s.ensureBuildQueue()
	mergeSurrenderTechnology(&s.Player, &loser.Player)
	s.Player.BC += loser.Player.BC
	s.Player.ActiveFreighters += loser.Player.ActiveFreighters
	s.Leaders = append(s.Leaders, unassignSurrenderedLeaders(loser.Leaders, 0)...)
	s.prepPlayerDerived()
	s.UpdatePlayerShipDesignsAfterTech()
}

func parallelInt(values []int, index, fallback int) int {
	if index >= 0 && index < len(values) {
		return values[index]
	}
	return fallback
}

func parallelBuildings(values []map[string]bool, index int) map[string]bool {
	if index >= 0 && index < len(values) {
		return values[index]
	}
	return nil
}

func (s *GameSession) clearSurrenderDiplomacy(loser int, receiver eventEmpireTarget) {
	s.ensureAIRelations()
	s.ensureAIAIState()
	for i := range s.AIPlayers {
		if loser < len(s.AIRelations) && i < len(s.AIRelations[loser]) {
			s.AIRelations[loser][i], s.AIRelations[i][loser] = 0, 0
			s.AIWars[loser][i], s.AIWars[i][loser] = false, false
			s.AIPolicies[loser][i], s.AIPolicies[i][loser] = gamedata.DIPLO_NONE, gamedata.DIPLO_NONE
			s.AITrade[loser][i], s.AITrade[i][loser] = false, false
			s.AIResearch[loser][i], s.AIResearch[i][loser] = false, false
		}
	}
	if loser < len(s.PlayerSpies) {
		if receiver.kind == eventEmpireAI && receiver.index >= 0 && receiver.index < len(s.PlayerSpies) {
			s.PlayerSpies[receiver.index] = minSurrenderInt(63, s.PlayerSpies[receiver.index]+s.PlayerSpies[loser])
		}
		s.PlayerSpies[loser] = 0
	}
}

func minSurrenderInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
