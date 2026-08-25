package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestOriginalCometCreationFormula(t *testing.T) {
	if got := originalCometStrength(0, 1); got != 110 {
		t.Fatalf("最低耐久=%d，want 110", got)
	}
	if got := originalCometStrength(4, 5); got != 190 {
		t.Fatalf("最高耐久=%d，want 190", got)
	}
	if got := originalCometCountdown(0, 5); got != 15 {
		t.Fatalf("簡單難度最長倒數=%d，want 15", got)
	}
	if got := originalCometCountdown(4, 1); got != 7 {
		t.Fatalf("最高難度最短倒數=%d，want 7", got)
	}
}

func TestOriginalCometImpactDamage(t *testing.T) {
	tests := []struct{ pop, buildings, a, b, want int }{{1, 0, 1, 1, 1}, {10, 5, 1, 1, 3}, {10, 5, 3, 3, 9}}
	for _, tc := range tests {
		if got := originalCometImpactDamage(tc.pop, tc.buildings, tc.a, tc.b); got != tc.want {
			t.Fatalf("impact(%d,%d,%d,%d)=%d，want %d", tc.pop, tc.buildings, tc.a, tc.b, got, tc.want)
		}
	}
}

func TestCometInterceptionIncludesAllOwnersButOnlyStationaryShips(t *testing.T) {
	s := NewDemoSession()
	star := s.PlayerColonyStarIndex(0)
	s.Fleets = []Fleet{
		{AtStar: star, ETA: 0, Ships: []Ship{{Class: "巡洋艦"}}},
		{AtStar: star, ETA: 2, Ships: []Ship{{Class: "末日之星"}}},
	}
	s.AIPlayers[0].FleetStar, s.AIPlayers[0].FleetETA = star, 0
	s.AIPlayers[0].Ships = []Ship{{Class: "戰艦"}}
	want := int(shipSizeClass("巡洋艦")) + 1 + int(shipSizeClass("戰艦")) + 1
	if got := s.cometInterceptionStrength(star); got != want {
		t.Fatalf("停泊的玩家與 AI 艦都應貢獻，航行艦不計：got %d want %d", got, want)
	}
}

func TestCometCanBeFullyInterceptedWithoutColonyDamage(t *testing.T) {
	s := NewDemoSession()
	planet, star := s.ColonyPlanetIndex(0), s.PlayerColonyStarIndex(0)
	s.Fleets = []Fleet{{AtStar: star, ETA: 0, Ships: []Ship{{Class: "末日之星"}}}}
	before := s.PlayerColonies[0]
	e := PersistentEvent{Kind: PersistentComet, PlanetIndex: planet, StarIndex: star,
		Countdown: 3, Strength: 1, InitialStrength: 100}
	done, message, messageEN := s.stepComet(&e)
	if !done || message == "" || messageEN == "" {
		t.Fatalf("彗星應被完全攔截：done=%v zh=%q en=%q", done, message, messageEN)
	}
	if s.PlayerColonies[0].Population != before.Population {
		t.Fatal("成功攔截不得傷害殖民地")
	}
}

func TestCometImpactUsesStrategicDamageInsteadOfAutomaticDestruction(t *testing.T) {
	s := NewDemoSession()
	planet, star := s.ColonyPlanetIndex(0), s.PlayerColonyStarIndex(0)
	s.Fleets = []Fleet{NewFleet(star)}
	s.PlayerColonies[0] = engine.ColonyState{Population: 1, Farmers: 1, Climate: gamedata.BARREN,
		IndustryPerWorker: 1, FlatIndustry: 5}
	s.ColonyBuildings[0] = map[string]bool{"自動工廠": true}
	e := PersistentEvent{Kind: PersistentComet, PlanetIndex: planet, StarIndex: star,
		Countdown: 1, Strength: 1000, InitialStrength: 1000}
	done, _, _ := s.stepComet(&e)
	if !done || len(s.PlayerColonies) != 1 || s.PlayerColonies[0].Population != 1 {
		t.Fatalf("唯一建築應吸收撞擊，不得固定全滅：colonies=%+v", s.PlayerColonies)
	}
	if s.ColonyBuildings[0]["自動工廠"] || s.PlayerColonies[0].FlatIndustry != 0 {
		t.Fatalf("被毀建築及效果應同步移除：buildings=%v colony=%+v", s.ColonyBuildings[0], s.PlayerColonies[0])
	}
}

func TestCometImpactWritesBackInactiveHotseat(t *testing.T) {
	s := NewDemoSession()
	if s.SetupHotseatWithAIIndices([]int{0}) != 2 {
		t.Fatal("需要兩席熱座")
	}
	targetPlanet := s.Seats[1].PlayerColonyPlanets[0]
	targetStar := s.Seats[1].PlayerColonyStars[0]
	s.Seats[1].PlayerColonies[0] = engine.ColonyState{Population: 1, Farmers: 1, Climate: gamedata.BARREN}
	s.Seats[1].ColonyBuildings[0] = map[string]bool{}
	s.Seats[1].PlayerColonyMarines = []int{0}
	s.Seats[1].PlayerColonyTanks = []int{0}
	activePop := s.PlayerColonies[0].Population
	e := PersistentEvent{Kind: PersistentComet, PlanetIndex: targetPlanet, StarIndex: targetStar,
		Countdown: 1, Strength: 1000, InitialStrength: 1000}
	done, _, _ := s.stepComet(&e)
	if !done || len(s.Seats[1].PlayerColonies) != 0 {
		t.Fatalf("非目前席位應承受撞擊並同步刪除：%+v", s.Seats[1].PlayerColonies)
	}
	if s.PlayerColonies[0].Population != activePop {
		t.Fatal("結算後必須恢復目前席位")
	}
}

func TestCometImpactWritesBackAIColony(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	planet, star := a.ColonyPlanets[0], a.ColonyStars[0]
	a.Colonies[0] = engine.ColonyState{Population: 1, Farmers: 1, Climate: gamedata.BARREN}
	a.ColonyBuildings[0] = map[string]bool{}
	a.ColonyMarines, a.ColonyTanks = []int{0}, []int{0}
	a.FleetETA = 2 // 航行中的 AI 艦不參與攔截。
	e := PersistentEvent{Kind: PersistentComet, PlanetIndex: planet, StarIndex: star,
		Countdown: 1, Strength: 1000, InitialStrength: 1000}
	done, _, _ := s.stepComet(&e)
	if !done || len(a.Colonies) != 0 || len(a.ColonyPlanets) != 0 || len(a.ColonyMarines) != 0 || len(a.ColonyTanks) != 0 {
		t.Fatalf("AI 彗星撞擊應同步殖民地平行陣列：ai=%+v", *a)
	}
}

func TestCometRecordSurvivesSaveLoad(t *testing.T) {
	s := NewDemoSession()
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentComet, PlanetIndex: 7, StarIndex: 3,
		Turns: 2, Countdown: 9, Strength: 123, InitialStrength: 170}}
	restored := s.snapshot().restore()
	if len(restored.PersistentEvents) != 1 {
		t.Fatal("彗星 record 未保存")
	}
	e := restored.PersistentEvents[0]
	if e.Kind != PersistentComet || e.PlanetIndex != 7 || e.StarIndex != 3 || e.Countdown != 9 || e.Strength != 123 || e.InitialStrength != 170 {
		t.Fatalf("彗星 record 往返失真：%+v", e)
	}
}

func TestCometIsInImplementedEventPool(t *testing.T) {
	ev := gamedata.RandomEventByID(2)
	if ev == nil || !ev.Implemented {
		t.Fatalf("事件 2 應進正常事件池：%+v", ev)
	}
	if gamedata.OriginalEventMinimumTurn(2) != 200 {
		t.Fatal("事件 2 最早 elapsed turn 必須是 200")
	}
}
