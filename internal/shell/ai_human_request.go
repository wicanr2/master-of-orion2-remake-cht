package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// applyOriginalAIHumanRequestAccept 對映 sub_1AEB5(human, AI) 的四種 payload。
func (s *GameSession) applyOriginalAIHumanRequestAccept(aiIndex int, action gamedata.OriginalHumanDiplomaticAction) bool {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return false
	}
	a := &s.AIPlayers[aiIndex]
	switch action.Kind {
	case gamedata.OriginalHumanDiplomaticActionDirect:
		if action.DirectTier < 1 || action.DirectTier > 2 {
			return false
		}
		a.OriginalHumanDirectRequestTier = action.DirectTier
		return true
	case gamedata.OriginalHumanDiplomaticActionTechnology:
		tech := gamedata.Technology(action.Technology)
		topic, ok := gamedata.OrigTechTopic(tech)
		if !ok || !playerStateKnowsTech(s.Player, topic, tech) {
			return false
		}
		grantTechnologyApplication(&a.Player, topic, tech)
		return true
	case gamedata.OriginalHumanDiplomaticActionCredits:
		if action.Credits <= 0 || action.Credits > 32000 {
			return false
		}
		a.Player.BC += action.Credits
		s.Player.BC -= action.Credits
		if s.Player.BC < 1 {
			s.Player.BC = 0
		}
		return true
	case gamedata.OriginalHumanDiplomaticActionColony:
		return s.transferOriginalHumanColonyToAI(aiIndex, action.Colony)
	}
	return false
}

// transferOriginalHumanColonyToAI 對映 sub_E4AB3 的玩家可見所有權轉移；候選 producer
// 已排除首都與唯一殖民星，typed 平行陣列必須一起搬移。
func (s *GameSession) transferOriginalHumanColonyToAI(aiIndex, starIdx int) bool {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) || starIdx < 0 || starIdx >= len(s.Stars) || len(s.PlayerColonies) <= 1 {
		return false
	}
	ci := -1
	for i := range s.PlayerColonies {
		if s.PlayerColonyStarIndex(i) == starIdx {
			ci = i
			break
		}
	}
	if ci < 0 || s.PlayerCapitolPlanetKnown && s.ColonyPlanetIndex(ci) == s.PlayerCapitolPlanet {
		return false
	}
	a := &s.AIPlayers[aiIndex]
	for _, owned := range a.ColonyStars {
		if owned == starIdx {
			return false
		}
	}
	colony, planet := s.PlayerColonies[ci], s.ColonyPlanetIndex(ci)
	var buildings map[string]bool
	if ci < len(s.ColonyBuildings) {
		buildings = cloneBuildings(s.ColonyBuildings[ci])
	}
	marines, tanks, marineAge, armorAge := 0, 0, 0, 0
	if ci < len(s.PlayerColonyMarines) {
		marines = s.PlayerColonyMarines[ci]
	}
	if ci < len(s.PlayerColonyTanks) {
		tanks = s.PlayerColonyTanks[ci]
	}
	if ci < len(s.MarineBarracksAge) {
		marineAge = s.MarineBarracksAge[ci]
	}
	if ci < len(s.ArmorBarracksAge) {
		armorAge = s.ArmorBarracksAge[ci]
	}
	s.removePlayerColony(ci)
	ensureAIGroundForceSlots(a)
	a.Colonies = append(a.Colonies, colony)
	a.ColonyStars = append(a.ColonyStars, starIdx)
	a.ColonyPlanets = append(a.ColonyPlanets, planet)
	a.ColonyBuildings = append(a.ColonyBuildings, buildings)
	a.ColonyMarines = append(a.ColonyMarines, marines)
	a.ColonyTanks = append(a.ColonyTanks, tanks)
	a.MarineBarracksAge = append(a.MarineBarracksAge, marineAge)
	a.ArmorBarracksAge = append(a.ArmorBarracksAge, armorAge)
	s.Stars[starIdx].Owner = 2
	a.OwnedStars++
	return true
}

// AcceptOriginalAIHumanDiplomaticRequest 執行 reason 105／106 的接受 callback。
func (s *GameSession) AcceptOriginalAIHumanDiplomaticRequest(aiIndex int) bool {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return false
	}
	a := &s.AIPlayers[aiIndex]
	r := a.OriginalHumanDiplomaticRequest
	if r == nil || (r.ReasonCode != 105 && r.ReasonCode != 106) || !s.applyOriginalAIHumanRequestAccept(aiIndex, r.Action) {
		return false
	}
	a.WantsAudience, a.AudienceReason, a.OriginalHumanDiplomaticRequest = false, "", nil
	return true
}

// AcknowledgeOriginalAIHumanDiplomaticNotice 清除 reason 124；sub_1AFA6 證實它不進二選一。
func (s *GameSession) AcknowledgeOriginalAIHumanDiplomaticNotice(aiIndex int) bool {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return false
	}
	a := &s.AIPlayers[aiIndex]
	if a.OriginalHumanDiplomaticRequest == nil || a.OriginalHumanDiplomaticRequest.ReasonCode != 124 {
		return false
	}
	a.WantsAudience, a.AudienceReason, a.OriginalHumanDiplomaticRequest = false, "", nil
	return true
}
