package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
)

func TestOriginalAIHumanTargetScoreConsumesTypedVerticalInputs(t *testing.T) {
	race := -1
	for index := range Races {
		candidate := AIOpponent{RaceIndex: index}
		if originalAIRelationGovernment(candidate) != 0 {
			race = index
			break
		}
	}
	if race < 0 {
		t.Fatal("測試資料找不到非 government raw 0 的原版種族")
	}
	s := &GameSession{
		Difficulty: 2, Turn: 20, RaceIndex: 0,
		Player:         engine.PlayerState{},
		PlayerColonies: []engine.ColonyState{{Population: 20, PopMax: 40}},
		Fleets:         []Fleet{{Ships: []Ship{originalPowerTestShip()}}},
		AIPlayers: []AIOpponent{{
			RaceIndex: race, Personality: 1,
			PopulationRaceSlot: 1, PopulationRaceSlotKnown: true,
			OriginalRaw28: 1, OriginalRaw28Known: true,
			OriginalHumanIncidentKnown: true,
			OriginalRelationRaw:        -20, OriginalRelationKnown: true,
			Colonies: []engine.ColonyState{{Population: 30, PopMax: 40}},
			Ships:    []Ship{originalPowerTestShip()},
		}},
	}
	result, ratio, ok := s.originalAIHumanTargetScore(0, func(n int) int { return n })
	if !ok || ratio <= 0 || result.ActionLimit <= 0 {
		t.Fatalf("typed score=%+v ratio=%d ok=%v", result, ratio, ok)
	}
}

func TestOriginalAIHumanTargetScoreFailsClosedForUnknownIncident(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) == 0 {
		t.Fatal("demo 應有 AI")
	}
	s.AIPlayers[0].OriginalHumanIncidentKnown = false
	if _, _, ok := s.originalAIHumanTargetScore(0, func(n int) int { return n }); ok {
		t.Fatal("GAM／舊 JSON 缺 +0x71F／+0x6CF 時不得產生部分 score")
	}
}
