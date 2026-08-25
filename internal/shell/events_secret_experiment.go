package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func (s *GameSession) completePlayerSecretExperiment() (gamedata.ResearchTopic, bool) {
	next, topic, completed := engine.ForceCompleteResearchTopic(s.Player)
	s.Player = next
	if !completed {
		return topic, false
	}
	applyResearchRaceTrait(&s.Player, topic, s.RaceCreative(), s.RaceUncreative(), s.researchRandForTurn().Intn)
	applyResearchTopicGrantCallbacks(&s.Player, topic)
	s.UpdatePlayerShipDesignsAfterTech()
	return topic, true
}

func (s *GameSession) completeAISecretExperiment(aiIndex int) (gamedata.ResearchTopic, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return 0, false
	}
	a := &s.AIPlayers[aiIndex]
	next, topic, completed := engine.ForceCompleteResearchTopic(a.Player)
	a.Player = next
	if !completed {
		return topic, false
	}
	applyResearchRaceTrait(&a.Player, topic,
		aiRaceHasTrait(*a, gamedata.TRAIT_CREATIVE),
		aiRaceHasTrait(*a, gamedata.TRAIT_UNCREATIVE), s.researchRandForTurn().Intn)
	applyResearchTopicGrantCallbacks(&a.Player, topic)
	s.updateAIShipDesignsAfterTech(aiIndex)
	return topic, true
}
