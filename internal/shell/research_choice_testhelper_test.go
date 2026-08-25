package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"

func runResearchForTest(s *GameSession) (engine.PlayerState, bool) {
	return engine.RunResearchPhase(s.Player, 1)
}
