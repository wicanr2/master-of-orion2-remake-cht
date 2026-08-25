package shell

import (
	"reflect"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestOriginalEarthquakeDamageFormula(t *testing.T) {
	tests := []struct {
		population, buildings, roll3, roll2, want int
	}{
		{1, 0, 1, 1, 1},
		{10, 0, 1, 1, 2},
		{10, 5, 3, 2, 7},
		{20, 10, 3, 2, 15},
	}
	for _, tc := range tests {
		if got := originalEarthquakeDamage(tc.population, tc.buildings, tc.roll3, tc.roll2); got != tc.want {
			t.Fatalf("originalEarthquakeDamage(%d,%d,%d,%d)=%d，want %d",
				tc.population, tc.buildings, tc.roll3, tc.roll2, got, tc.want)
		}
	}
}

func TestPickEarthquakeColonyUsesReservoirOrder(t *testing.T) {
	colonies := make([]engine.ColonyState, 3)
	wants := []int{1, 2, 3}
	returns := []int{0, 1, 0}
	calls := make([]int, 0, 3)
	pick, ok := pickEarthquakeColony(colonies, func(n int) int {
		calls = append(calls, n)
		return returns[len(calls)-1]
	})
	if !ok || pick != 2 {
		t.Fatalf("第三候選應取代第一候選：pick=%d ok=%v", pick, ok)
	}
	if !reflect.DeepEqual(calls, wants) {
		t.Fatalf("reservoir 應依序擲 Random(1),Random(2),Random(3)：got %v", calls)
	}
}

func TestResolveEarthquakeReusesStrategicBuildingDamage(t *testing.T) {
	colony := engine.ColonyState{Population: 1, Farmers: 1}
	buildings := map[string]bool{"自動工廠": true}
	result := resolveEarthquakeDamage(&colony, buildings, 0, 0, engine.PlayerState{}, false, newRandStream(7))
	if len(result.DestroyedBuildingIDs) != 1 || result.DestroyedBuildingIDs[0] != 7 {
		t.Fatalf("唯一一般建築應由共用 sub_DCEBD 模型摧毀：%+v", result)
	}
	removeDestroyedRawBuildings(buildings, result.DestroyedBuildingIDs)
	if buildings["自動工廠"] || colony.Population != 1 {
		t.Fatalf("建築應刪除且最後人口應保留：buildings=%v colony=%+v", buildings, colony)
	}
}

func TestPlayerEarthquakeRemovesDestroyedBuildingEffect(t *testing.T) {
	s := NewDemoSession()
	s.eventRand = newRandStream(9)
	s.PlayerColonies[0] = engine.ColonyState{Population: 1, Farmers: 1, IndustryPerWorker: 1, FlatIndustry: 5}
	s.ColonyBuildings[0] = map[string]bool{"自動工廠": true}
	s.PlayerColonyMarines = []int{0}
	s.PlayerColonyTanks = []int{0}
	impact, ok := s.applyPlayerEarthquake()
	if !ok || impact.BuildingsDestroyed != 1 || impact.ColonyDestroyed {
		t.Fatalf("唯一一般建築應吸收地震傷害：impact=%+v ok=%v", impact, ok)
	}
	if s.ColonyBuildings[0]["自動工廠"] || s.PlayerColonies[0].IndustryPerWorker != 0 ||
		s.PlayerColonies[0].FlatIndustry != 0 {
		t.Fatalf("建築旗標與 typed 產出效果必須一起移除：colony=%+v buildings=%v",
			s.PlayerColonies[0], s.ColonyBuildings[0])
	}
}

func TestPlayerEarthquakeCanDestroyLastColonyAndParallelSlots(t *testing.T) {
	s := NewDemoSession()
	s.eventRand = newRandStream(11)
	s.PlayerColonies[0] = engine.ColonyState{Population: 1, Farmers: 1}
	s.ColonyBuildings[0] = map[string]bool{}
	s.PlayerColonyMarines = []int{0}
	s.PlayerColonyTanks = []int{0}
	s.MarineBarracksAge = []int{0}
	s.ArmorBarracksAge = []int{0}
	s.AutoBuild = []bool{false}
	s.RepeatBuild = []ColonyBuild{{}}
	s.LastBuilt = []string{""}

	impact, ok := s.applyPlayerEarthquake()
	if !ok || !impact.ColonyDestroyed || impact.PopulationLost != 1 {
		t.Fatalf("最後人口應依 100 點尾端規則摧毀殖民地：impact=%+v ok=%v", impact, ok)
	}
	lengths := []int{len(s.PlayerColonies), len(s.Builds), len(s.BuildQueue), len(s.AutoBuild),
		len(s.RepeatBuild), len(s.LastBuilt), len(s.ColonyBuildings), len(s.PlayerColonyMarines),
		len(s.PlayerColonyTanks), len(s.PlayerColonyStars), len(s.PlayerColonyPlanets)}
	for _, got := range lengths {
		if got != 0 {
			t.Fatalf("殖民地摧毀後所有平行陣列都應移除一槽：lengths=%v", lengths)
		}
	}
}

func TestAIEarthquakeWritesBackAndRemovesDestroyedColony(t *testing.T) {
	s := NewDemoSession()
	s.eventRand = newRandStream(13)
	a := &s.AIPlayers[0]
	a.Colonies = []engine.ColonyState{{Population: 1, Farmers: 1}}
	a.ColonyBuildings = []map[string]bool{{}}
	a.ColonyStars = []int{1}
	a.ColonyPlanets = []int{-1}
	a.ColonyMarines, a.ColonyTanks = nil, nil
	a.MarineBarracksAge, a.ArmorBarracksAge = nil, nil

	impact, ok := s.applyAIEarthquake(0)
	if !ok || !impact.ColonyDestroyed || len(a.Colonies) != 0 || len(a.ColonyStars) != 0 ||
		len(a.ColonyBuildings) != 0 || len(a.ColonyMarines) != 0 || len(a.ColonyTanks) != 0 {
		t.Fatalf("AI 殖民地與平行欄位未完整回寫：impact=%+v ai=%+v ok=%v", impact, *a, ok)
	}
}

func TestHotseatEarthquakeTargetsInactiveSeat(t *testing.T) {
	s := NewDemoSession()
	if got := s.SetupHotseatWithAIIndices([]int{0}); got != 2 {
		t.Fatalf("需要兩個熱座席位，got %d", got)
	}
	s.eventRand = newRandStream(17)
	activePopulation := s.PlayerColonies[0].Population
	s.Seats[1].PlayerColonies[0] = engine.ColonyState{Population: 1, Farmers: 1}
	s.Seats[1].ColonyBuildings[0] = map[string]bool{}
	s.Seats[1].PlayerColonyMarines = []int{0}
	s.Seats[1].PlayerColonyTanks = []int{0}

	ev := *gamedata.RandomEventByID(7)
	result, ok := s.applyRandomEventLocalizedToTarget(ev, eventEmpireTarget{kind: eventEmpireSeat, index: 1})
	if !ok || result.Message == "" || result.MessageEN == "" {
		t.Fatalf("非目前席位地震應產生同一次雙語結果：result=%+v ok=%v", result, ok)
	}
	if len(s.Seats[1].PlayerColonies) != 0 {
		t.Fatalf("地震結果應寫回非目前席位：%+v", s.Seats[1].PlayerColonies)
	}
	if s.PlayerColonies[0].Population != activePopulation {
		t.Fatalf("結算後應恢復目前席位，人口 %d → %d", activePopulation, s.PlayerColonies[0].Population)
	}
}
