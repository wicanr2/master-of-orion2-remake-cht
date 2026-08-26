package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestSub23DA0SelectorExcludesCompletedCapitol(t *testing.T) {
	colonies := []engine.ColonyState{{Population: 8}, {Population: 4}}
	buildings := []map[string]bool{{CapitolBuildName: true}, {"Automated Factory": true}}
	eligible := func(i int) bool { return !colonyHasCapitol(buildings, i) }
	for seed := int64(1); seed <= 20; seed++ {
		rng := newRandStream(seed)
		got, ok := pickEarthquakeColony(colonies, eligible, rng.Intn)
		if !ok || got != 1 {
			t.Fatalf("只有非 Capitol 殖民地 1 合格，seed=%d 得 (%d,%v)", seed, got, ok)
		}
	}
	if _, ok := pickEarthquakeColony(colonies,
		func(int) bool { return false }, newRandStream(1).Intn); ok {
		t.Fatal("所有殖民地都有 Capitol 時不得捏造事件目標")
	}
}

func TestIndustrialAndMineralDepletionExcludeCapitolButDiscoveryDoesNot(t *testing.T) {
	colonies := []engine.ColonyState{
		{Population: 8, Climate: gamedata.TERRAN},
		{Population: 5, Climate: gamedata.DESERT},
	}
	buildings := []map[string]bool{{CapitolBuildName: true}, nil}
	if got, ok := originalIndustrialAccidentColony(colonies, buildings, newRandStream(3)); !ok || got != 1 {
		t.Fatalf("工業事故只可選非 Capitol 殖民地 1，得到 (%d,%v)", got, ok)
	}
	planets := []Planet{{MineralID: gamedata.ULTRA_RICH}, {MineralID: gamedata.ULTRA_RICH}}
	planetAt := func(i int) *Planet { return &planets[i] }
	if got, ok := originalMineralEventColony(colonies, buildings, planetAt, 11, newRandStream(4)); !ok || got != 1 {
		t.Fatalf("礦產枯竭只可選非 Capitol 殖民地 1，得到 (%d,%v)", got, ok)
	}
	planets[0].MineralID = gamedata.POOR
	planets[1].MineralID = gamedata.ULTRA_RICH
	if got, ok := originalMineralEventColony(colonies, buildings, planetAt, 12, newRandStream(5)); !ok || got != 0 {
		t.Fatalf("礦產發現使用另一 helper，應可選 Capitol 殖民地 0，得到 (%d,%v)", got, ok)
	}
}

func TestPlagueAndSupernovaUseTypedCapitolBuildingState(t *testing.T) {
	s := &GameSession{}
	s.eventRand = newRandStream(7)
	s.PlayerColonies = []engine.ColonyState{{Population: 8}, {Population: 4}}
	s.Planets = make([]Planet, 2)
	s.PlayerColonyPlanets = []int{0, 1}
	s.PlayerColonyStars = []int{0, 1}
	s.ColonyBuildings = []map[string]bool{{CapitolBuildName: true}, nil}
	if got, ok := s.plagueTargetEligible(s.PlayerColonies, s.ColonyBuildings, s.ColonyPlanetIndex); !ok || got != 1 {
		t.Fatalf("瘟疫只可選非 Capitol 殖民地 1，得到 (%d,%v)", got, ok)
	}
	if s.galaxyHasEventColonyWithoutCapitolAtStar(0) {
		t.Fatal("只有 Capitol 的星系不得成為超新星候選")
	}
	if !s.galaxyHasEventColonyWithoutCapitolAtStar(1) {
		t.Fatal("有 active 非 Capitol 殖民地的星系應可成為超新星候選")
	}
}

func TestSupernovaCapitolFilterReadsNonCurrentHotseat(t *testing.T) {
	s := &GameSession{ActiveSeat: 0, Seats: []seat{
		{},
		{
			PlayerColonies:      []engine.ColonyState{{Population: 6}},
			PlayerColonyStars:   []int{3},
			ColonyBuildings:     []map[string]bool{nil},
			PlayerColonyPlanets: []int{8},
		},
	}}
	if !s.galaxyHasEventColonyWithoutCapitolAtStar(3) {
		t.Fatal("非目前熱座席位的無 Capitol 殖民地必須進入全銀河超新星候選")
	}
	s.Seats[1].ColonyBuildings[0] = map[string]bool{CapitolBuildName: true}
	if s.galaxyHasEventColonyWithoutCapitolAtStar(3) {
		t.Fatal("非目前熱座席位只有 Capitol 時不得使該星合格")
	}
}
