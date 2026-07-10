package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"

func runResearchForTest(s *GameSession) (engine.PlayerState, bool) {
	return engine.RunResearchPhase(s.Player, 0) // 進度已灌滿,0 新點即可觸發完成判定
}
