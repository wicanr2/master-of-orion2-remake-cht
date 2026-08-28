package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func originalTargetPopulationColony(slot int) engine.ColonyState {
	return engine.ColonyState{
		Population: 1, Farmers: 1, OwnerRaceSlot: slot, OwnerRaceSlotKnown: true, OwnerRaceProfileKnown: true,
		PopulationGroups: []engine.PopulationGroup{{
			RaceSlot: slot, RaceSlotKnown: true, ProfileKnown: true, Farmers: 1,
		}},
	}
}

func TestOriginalAIHumanGovernmentZeroReachabilityCountsRangeAndSharedColony(t *testing.T) {
	s := &GameSession{
		Stars:             []Star{{X: 0.10, Y: 0.10, Wormhole: -1}, {X: 0.11, Y: 0.10, Wormhole: -1}},
		PlayerColonies:    []engine.ColonyState{originalTargetPopulationColony(0)},
		PlayerColonyStars: []int{0},
		AIPlayers: []AIOpponent{{
			PopulationRaceSlot: 1, PopulationRaceSlotKnown: true,
			Player: engine.PlayerState{GrantedTechs: map[gamedata.Technology]bool{
				gamedata.TECH_STANDARD_FUEL_CELLS: true,
			}},
			Colonies:    []engine.ColonyState{originalTargetPopulationColony(1)},
			ColonyStars: []int{1},
		}},
	}
	if got, ok := s.originalAIHumanGovernmentZeroReachability(0); !ok || got != 1 {
		t.Fatalf("range score=%d ok=%v, want 1/true", got, ok)
	}
	s.PlayerColonies[0].Population = 2
	s.PlayerColonies[0].Farmers = 2
	s.PlayerColonies[0].PopulationGroups = append(s.PlayerColonies[0].PopulationGroups,
		engine.PopulationGroup{RaceSlot: 1, RaceSlotKnown: true, ProfileKnown: true, Farmers: 1})
	if got, ok := s.originalAIHumanGovernmentZeroReachability(0); !ok || got != 5 {
		t.Fatalf("shared-colony score=%d ok=%v, want 5/true", got, ok)
	}
}

func TestOriginalAIHumanGovernmentZeroReachabilityUsesVisitedWormholePartner(t *testing.T) {
	s := &GameSession{
		Stars: []Star{
			{X: 0.10, Y: 0.10, Wormhole: 1},
			{X: 0.90, Y: 0.90, Wormhole: 0},
			{X: 0.89, Y: 0.90, Wormhole: -1},
		},
		PlayerColonies:    []engine.ColonyState{originalTargetPopulationColony(0)},
		PlayerColonyStars: []int{0},
		AIPlayers: []AIOpponent{{
			PopulationRaceSlot: 1, PopulationRaceSlotKnown: true,
			Player: engine.PlayerState{GrantedTechs: map[gamedata.Technology]bool{
				gamedata.TECH_STANDARD_FUEL_CELLS: true,
			}},
			Colonies: []engine.ColonyState{originalTargetPopulationColony(1)}, ColonyStars: []int{2},
			ExploredStars: []bool{false, true, true}, ExploredStarsKnown: true,
		}},
	}
	if got, ok := s.originalAIHumanGovernmentZeroReachability(0); !ok || got != 1 {
		t.Fatalf("visited wormhole score=%d ok=%v, want 1/true", got, ok)
	}
	s.AIPlayers[0].ExploredStars[1] = false
	if got, ok := s.originalAIHumanGovernmentZeroReachability(0); !ok || got != 0 {
		t.Fatalf("unvisited wormhole score=%d ok=%v, want 0/true", got, ok)
	}
	s.AIPlayers[0].ExploredStarsKnown = false
	if _, ok := s.originalAIHumanGovernmentZeroReachability(0); ok {
		t.Fatal("舊存檔缺逐帝國造訪歷史時不得猜測蟲洞可達")
	}
}

func TestOriginalAIExploredStarsSurviveSnapshot(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers[0].ExploredStars[3] = true
	got := s.snapshot().restore().AIPlayers[0]
	if !got.ExploredStarsKnown || len(got.ExploredStars) != len(s.Stars) || !got.ExploredStars[3] {
		t.Fatalf("snapshot lost explored stars: known=%v stars=%v", got.ExploredStarsKnown, got.ExploredStars)
	}
}

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
