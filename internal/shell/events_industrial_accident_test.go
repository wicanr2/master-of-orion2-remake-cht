package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestOriginalIndustrialAccidentHitsAllowsZero(t *testing.T) {
	tests := []struct{ population, a, b, want int }{
		{1, 1, 1, 0}, {5, 1, 1, 1}, {10, 3, 3, 6},
	}
	for _, tc := range tests {
		if got := originalIndustrialAccidentHits(tc.population, tc.a, tc.b); got != tc.want {
			t.Fatalf("hits(%d,%d,%d)=%d，want %d", tc.population, tc.a, tc.b, got, tc.want)
		}
	}
}

func TestIndustrialAccidentClimateEligibility(t *testing.T) {
	for _, climate := range []gamedata.PlanetClimate{gamedata.TOXIC, gamedata.RADIATED} {
		if _, ok := originalIndustrialAccidentColony([]engine.ColonyState{{Climate: climate}}, newRandStream(1)); ok {
			t.Fatalf("%v 不得成為工業事故目標", climate)
		}
	}
	for climate := gamedata.BARREN; climate <= gamedata.GAIA; climate++ {
		if i, ok := originalIndustrialAccidentColony([]engine.ColonyState{{Climate: climate}}, newRandStream(1)); !ok || i != 0 {
			t.Fatalf("%v 應可成為工業事故目標：i=%d ok=%v", climate, i, ok)
		}
	}
}

func TestIndustrialSpecialHitsDoNotKillAndroid(t *testing.T) {
	c := engine.ColonyState{
		Population: 2, Workers: 2, OwnerRaceProfileKnown: true,
		PopulationGroups: []engine.PopulationGroup{{RaceSlot: gamedata.AndroidColonistSlot,
			RaceSlotKnown: true, Workers: 2, ProfileKnown: true}},
	}
	pop, marines, tanks := resolveIndustrialSpecialHits(&c, 0, 0, 5, newRandStream(3))
	if pop != 0 || marines != 0 || tanks != 0 || c.Population != 2 || c.Workers != 2 {
		t.Fatalf("全 Android 殖民地應浪費特殊命中：colony=%+v loss=%d/%d/%d", c, pop, marines, tanks)
	}
}

func TestIndustrialSpecialHitsExcludeAndroidInMixedColony(t *testing.T) {
	c := engine.ColonyState{
		Population: 2, Farmers: 1, Workers: 1, OwnerRaceProfileKnown: true,
		PopulationGroups: []engine.PopulationGroup{
			{RaceSlot: 0, RaceSlotKnown: true, Farmers: 1, ProfileKnown: true},
			{RaceSlot: gamedata.AndroidColonistSlot, RaceSlotKnown: true, Workers: 1, ProfileKnown: true},
		},
	}
	pop, _, _ := resolveIndustrialSpecialHits(&c, 0, 0, 1, newRandStream(5))
	if pop != 1 || c.Population != 1 || c.Farmers != 0 || c.Workers != 1 ||
		populationGroupUnits(c.PopulationGroups[1]) != 1 {
		t.Fatalf("混居特殊命中只能殺非 Android：colony=%+v popLost=%d", c, pop)
	}
}

func TestIndustrialAccidentTolerantPlayerIsIneligible(t *testing.T) {
	s := NewDemoSession()
	s.RaceIndex = -1
	s.CustomRaceTraits = uint32(1) << uint(gamedata.TRAIT_TOLERANT)
	if _, ok := s.applyPlayerIndustrialAccident(); ok {
		t.Fatal("Tolerant 玩家不得結算工業事故")
	}
}

func TestIndustrialAccidentZeroSpecialHitsStillAppliesRegularDamage(t *testing.T) {
	c := engine.ColonyState{Population: 1, Farmers: 1, Climate: gamedata.BARREN}
	buildings := map[string]bool{"自動工廠": true}
	regular, specialAndRegularPop, _, _ := resolveIndustrialAccident(&c, buildings, 0, 0,
		engine.PlayerState{}, false, newRandStream(7))
	if len(regular.DestroyedBuildingIDs) != 1 || specialAndRegularPop != 0 || c.Population != 1 {
		t.Fatalf("零次特殊命中後仍應由固定一般傷害摧毀唯一建築：result=%+v colony=%+v", regular, c)
	}
}
