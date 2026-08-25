package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

const originalAIDeclinedLeaderCooldown = 30

func (s *GameSession) advanceOfficerCooldowns() {
	for id, turns := range s.OfficerCooldowns {
		turns--
		if turns <= 0 {
			delete(s.OfficerCooldowns, id)
			continue
		}
		s.OfficerCooldowns[id] = turns
	}
}

func aiWantsLeader(ai *AIOpponent, leader Leader) bool {
	if ai == nil {
		return false
	}
	preferred := []gamedata.LeaderSkills{
		gamedata.SKILL_ASSASSIN, gamedata.SKILL_FAMOUS, gamedata.SKILL_MEGAWEALTH,
		gamedata.SKILL_OPERATIONS, gamedata.SKILL_SPYMASTER, gamedata.SKILL_TELEPATH,
		gamedata.SKILL_HELMSMAN, gamedata.SKILL_WEAPONRY,
	}
	for _, skill := range preferred {
		if leaderSkillTier(leader, int(skill)) > 0 {
			return true
		}
	}
	// 原版 +0x59D 的完整 raw 名稱仍未知；有可運作研究帝國是目前 typed gate。
	if leaderSkillTier(leader, int(gamedata.SKILL_RESEARCHER)) > 0 && len(ai.Colonies) > 0 {
		return true
	}
	return leaderSkillTier(leader, int(gamedata.SKILL_TRADER)) > 0 &&
		aiRaceHasTrait(*ai, gamedata.TRAIT_REPULSIVE)
}

func aiLeaderHireCost(ai *AIOpponent, leader Leader) int {
	return gamedata.LeaderHireCost(5, leaderDisplayLevelToExpLevel(leader.Level),
		leaderFamousHireModifier(ai.Leaders))
}

func (s *GameSession) processAILeaderOffer(aiIndex int) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return
	}
	ai := &s.AIPlayers[aiIndex]
	if ai.LeaderOffer == nil {
		return
	}
	leader := *ai.LeaderOffer
	cost := aiLeaderHireCost(ai, leader)
	accept := aiWantsLeader(ai, leader) && !leaderSlotsFullFor(ai.Leaders, leader.Ship) && ai.Player.BC > cost+50
	if accept {
		ai.Player.BC -= cost
		leader.RawStatus = 0
		leader.RawETA = 0
		leader.RawPlayerIndex = aiIndex + 1
		ai.Leaders = append(ai.Leaders, leader)
	} else {
		if s.OfficerCooldowns == nil {
			s.OfficerCooldowns = make(map[int]int)
		}
		s.OfficerCooldowns[leader.ID] = originalAIDeclinedLeaderCooldown
	}
	ai.LeaderOffer = nil
}

func (s *GameSession) advanceAILeaderOffers() {
	s.advanceOfficerCooldowns()
	// 原版 Do_AI_Leaders 從高玩家槽往低槽處理。
	for i := len(s.AIPlayers) - 1; i >= 0; i-- {
		s.processAILeaderOffer(i)
		s.assignAILeaders(i)
	}
}

func (s *GameSession) generateAILeaderOffers() {
	// Random_Officer_Check 的槽順序是玩家 0 之後依序 1..N-1。
	for i := range s.AIPlayers {
		ai := &s.AIPlayers[i]
		charismatic := aiRaceHasTrait(*ai, gamedata.TRAIT_CHARISMATIC)
		repulsive := aiRaceHasTrait(*ai, gamedata.TRAIT_REPULSIVE)
		chance := officerRecruitChanceFor(s.Turn, ai.LeaderLastOfferTurn, ai.Leaders, charismatic, repulsive)
		if chance <= 0 || s.officerRandForTurn().Intn(100) >= chance {
			continue
		}
		candidate, ok := s.pickOfficerCandidate(ai.Leaders, charismatic, repulsive)
		if !ok {
			continue
		}
		copyCandidate := candidate
		ai.LeaderOffer = &copyCandidate
		ai.LeaderLastOfferTurn = s.Turn
	}
}

func ensureAIColonyLeaderSlots(ai *AIOpponent) {
	if len(ai.ColonyLeaderNames) < len(ai.Colonies) {
		ai.ColonyLeaderNames = append(ai.ColonyLeaderNames,
			make([]string, len(ai.Colonies)-len(ai.ColonyLeaderNames))...)
	} else if len(ai.ColonyLeaderNames) > len(ai.Colonies) {
		ai.ColonyLeaderNames = ai.ColonyLeaderNames[:len(ai.Colonies)]
	}
}

func aiShipAssignmentScore(ship Ship) int {
	if ship.ProductionCost > 0 {
		return ship.ProductionCost
	}
	return ShipCost(ship.Class)
}

func aiColonyAssignmentScore(colonyIndex int, ai *AIOpponent) int {
	colony := ai.Colonies[colonyIndex]
	return colony.Population*100 + colony.IndustryPerWorker*10 + colony.ResearchPerScientist*10
}

func (s *GameSession) assignAILeaders(aiIndex int) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return
	}
	ai := &s.AIPlayers[aiIndex]
	ensureAIColonyLeaderSlots(ai)
	for li := range ai.Leaders {
		leader := &ai.Leaders[li]
		if leader.RawStatus == originalLeaderLimboStatus || leader.RawStatus == 1 {
			continue
		}
		if leader.Ship {
			best := -1
			for si := range ai.Ships {
				if ai.Ships[si].OfficerName != "" {
					continue
				}
				if best < 0 || aiShipAssignmentScore(ai.Ships[si]) > aiShipAssignmentScore(ai.Ships[best]) {
					best = si
				}
			}
			if best >= 0 {
				ai.Ships[best].OfficerName = leader.Name
				ai.Ships[best].OfficerID = leader.ID
				leader.RawStatus = 1
				leader.RawLocation = best
			}
			continue
		}
		best := -1
		for ci := range ai.Colonies {
			if ai.ColonyLeaderNames[ci] != "" {
				continue
			}
			if best < 0 || aiColonyAssignmentScore(ci, ai) > aiColonyAssignmentScore(best, ai) {
				best = ci
			}
		}
		if best >= 0 {
			ai.ColonyLeaderNames[best] = leader.Name
			applyColonyLeaderBonusDelta(&ai.Colonies[best], colonyLeaderBonusFor(*leader), 1)
			leader.RawStatus = 1
			leader.RawLocation = best
		}
	}
}
