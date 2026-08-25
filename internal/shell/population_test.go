package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestPopulationGrowthPointBoundaryAndSnapshot(t *testing.T) {
	s := NewDemoSession()
	s.PlayerColonies[0].Population = 8
	s.PlayerColonies[0].Farmers = 4
	s.PlayerColonies[0].Workers = 2
	s.PlayerColonies[0].Scientists = 2
	s.popAccum = make([]int, len(s.PlayerColonies))
	s.popAccum[0] = gamedata.PopulationGrowthPointsPerUnit - 1
	s.LastPlayerOutput = engine.EmpireOutput{Colonies: []engine.ColonyOutput{{PopGrowth: 0}}}

	s.advancePopulation()
	if got := s.PlayerColonies[0].Population; got != 8 {
		t.Fatalf("999 點不應增加人口，實得 %d", got)
	}
	if got := s.snapshot().restore().popAccum[0]; got != gamedata.PopulationGrowthPointsPerUnit-1 {
		t.Fatalf("快照還原不得縮放成長點數，實得 %d", got)
	}

	s.LastPlayerOutput.Colonies[0].PopGrowth = 1
	s.advancePopulation()
	if got := s.PlayerColonies[0].Population; got != 9 {
		t.Fatalf("1,000 點應恰增加一人口，實得 %d", got)
	}
	if got := s.popAccum[0]; got != 0 {
		t.Fatalf("兌換一人口後餘數應為 0，實得 %d", got)
	}
}

func TestPopulationGrowthCreditsExactSourceGroup(t *testing.T) {
	s := NewDemoSession()
	c := &s.PlayerColonies[0]
	c.Population, c.Farmers, c.Workers, c.Scientists = 2, 1, 1, 0
	c.PopMax = 20
	c.OwnerRaceProfileKnown, c.OwnerRaceSlotKnown, c.OwnerRaceSlot = true, true, 0
	c.PopulationGroups = []engine.PopulationGroup{
		{RaceSlot: 0, RaceSlotKnown: true, Farmers: 1, ProfileKnown: true},
		{RaceSlot: 1, RaceSlotKnown: true, Workers: 1,
			GrowthPoints: gamedata.PopulationGrowthPointsPerUnit - 1, ProfileKnown: true},
	}
	var rates [gamedata.PopulationRaceSlots]int
	rates[1] = 1
	s.LastPlayerOutput = engine.EmpireOutput{Colonies: []engine.ColonyOutput{{
		PopulationGroupGrowth: rates, PopulationGroupGrowthCount: 2,
	}}}

	s.advancePopulation()
	if c.Population != 3 || c.PopulationGroups[0].Farmers != 1 ||
		c.PopulationGroups[1].Workers+c.PopulationGroups[1].Farmers != 2 || c.PopulationGroups[1].GrowthPoints != 0 {
		t.Fatalf("新增人口必須保留來源 slot 1，got colony=%+v groups=%+v", *c, c.PopulationGroups)
	}
}

func TestNegativeGrowthRemovesNonFarmerAndKeepsLastColonist(t *testing.T) {
	s := NewDemoSession()
	c := &s.PlayerColonies[0]
	c.Population, c.Farmers, c.Workers, c.Scientists = 2, 1, 1, 0
	c.OwnerRaceProfileKnown, c.OwnerRaceSlotKnown, c.OwnerRaceSlot = true, true, 0
	c.PopulationGroups = []engine.PopulationGroup{{RaceSlot: 0, RaceSlotKnown: true,
		Farmers: 1, Workers: 1, ProfileKnown: true}}
	var rates [gamedata.PopulationRaceSlots]int
	rates[0] = -1
	s.LastPlayerOutput = engine.EmpireOutput{Colonies: []engine.ColonyOutput{{
		PopulationGroupGrowth: rates, PopulationGroupGrowthCount: 1,
	}}}
	s.advancePopulation()
	if c.Population != 1 || c.Farmers != 1 || c.Workers != 0 || c.PopulationGroups[0].GrowthPoints != 999 {
		t.Fatalf("負池應先刪非農夫並加回 1000，got colony=%+v", *c)
	}
	s.LastPlayerOutput.Colonies[0].PopulationGroupGrowth[0] = -1000
	s.advancePopulation()
	if c.Population != 1 || c.PopulationGroups[0].GrowthPoints != 0 {
		t.Fatalf("保護槽最後一人不得死亡且負債歸零，got colony=%+v", *c)
	}
}

func TestNegativeGrowthPrisonerWritebackAndAlienExtinction(t *testing.T) {
	s := NewDemoSession()
	c := &s.PlayerColonies[0]
	c.Population, c.Farmers, c.Workers, c.Scientists = 2, 1, 1, 0
	c.UnassimilatedPop, c.UnassimilatedWorkers = 1, 1
	c.OwnerRaceProfileKnown, c.OwnerRaceSlotKnown, c.OwnerRaceSlot = true, true, 0
	c.PopulationGroups = []engine.PopulationGroup{
		{RaceSlot: 0, RaceSlotKnown: true, Farmers: 1, ProfileKnown: true},
		{RaceSlot: 1, RaceSlotKnown: true, Workers: 1, PrisonerWorkers: 1, ProfileKnown: true},
	}
	var rates [gamedata.PopulationRaceSlots]int
	rates[1] = -1
	s.LastPlayerOutput = engine.EmpireOutput{Colonies: []engine.ColonyOutput{{
		PopulationGroupGrowth: rates, PopulationGroupGrowthCount: 2,
	}}}
	s.advancePopulation()
	if c.Population != 1 || c.Workers != 0 || c.UnassimilatedPop != 0 || c.UnassimilatedWorkers != 0 ||
		c.PopulationGroups[1].Workers != 0 || c.PopulationGroups[1].PrisonerWorkers != 0 {
		t.Fatalf("外族 prisoner 死亡必須同步所有計數，got colony=%+v groups=%+v", *c, c.PopulationGroups)
	}
}

func TestAINegativeGrowthUsesSamePopulationRule(t *testing.T) {
	a := AIOpponent{Colonies: []engine.ColonyState{{
		Population: 2, Farmers: 1, Workers: 1, PopMax: 20,
		OwnerRaceProfileKnown: true, OwnerRaceSlotKnown: true, OwnerRaceSlot: 0,
		PopulationGroups: []engine.PopulationGroup{{RaceSlot: 0, RaceSlotKnown: true,
			Farmers: 1, Workers: 1, ProfileKnown: true}},
	}}}
	var rates [gamedata.PopulationRaceSlots]int
	rates[0] = -1
	advanceAIColonyPopulation(&a, engine.EmpireOutput{Colonies: []engine.ColonyOutput{{
		PopulationGroupGrowth: rates, PopulationGroupGrowthCount: 1,
	}}}, 1, newRandStream(7))
	if a.Colonies[0].Population != 1 || a.Colonies[0].Workers != 0 || a.Colonies[0].PopulationGroups[0].GrowthPoints != 999 {
		t.Fatalf("AI 應共用負成長刪人口規則，got %+v", a.Colonies[0])
	}
}

func TestPopulationNegativePassRunsBeforePositivePass(t *testing.T) {
	s := NewDemoSession()
	c := &s.PlayerColonies[0]
	c.Population, c.PopMax, c.Farmers, c.Workers, c.Scientists = 2, 2, 1, 1, 0
	c.OwnerRaceProfileKnown, c.OwnerRaceSlotKnown, c.OwnerRaceSlot = true, true, 0
	c.PopulationGroups = []engine.PopulationGroup{
		{RaceSlot: 0, RaceSlotKnown: true, Farmers: 1,
			GrowthPoints: gamedata.PopulationGrowthPointsPerUnit - 1, ProfileKnown: true},
		{RaceSlot: 1, RaceSlotKnown: true, Workers: 1, ProfileKnown: true},
	}
	var rates [gamedata.PopulationRaceSlots]int
	rates[0], rates[1] = 1, -1
	s.LastPlayerOutput = engine.EmpireOutput{Colonies: []engine.ColonyOutput{{
		PopulationGroupGrowth: rates, PopulationGroupGrowthCount: 2,
	}}}
	s.advancePopulation()
	if c.Population != 2 || populationGroupUnits(c.PopulationGroups[0]) != 2 || populationGroupUnits(c.PopulationGroups[1]) != 0 {
		t.Fatalf("應先刪 slot1 再讓 owner slot0 使用空出的容量，got colony=%+v groups=%+v", *c, c.PopulationGroups)
	}
}

func TestPositivePopulationSlotOrderOwnerFirstAndConsumesFinalRandomOne(t *testing.T) {
	rng := newRandStream(17)
	order := positivePopulationSlotOrder(2, 4, rng)
	if len(order) != 4 || order[0] != 2 {
		t.Fatalf("owner slot 2 必須固定第一，got %v", order)
	}
	seen := [4]bool{}
	for _, slot := range order {
		if slot < 0 || slot >= len(seen) || seen[slot] {
			t.Fatalf("洗牌必須是 0..3 的排列，got %v", order)
		}
		seen[slot] = true
	}
	if rng.Draws() != 3 { // 其餘三格，含最後一格 Random(1)。
		t.Fatalf("sub_FE9F5 形式應消耗 3 次亂數，got %d", rng.Draws())
	}
}

// TestPopulationGrowthWriteback 驗證殖民地人口會隨回合成長並回寫 Population,且不超過 PopMax。
func TestPopulationGrowthWriteback(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true // 隔離瘟疫/隕石(會扣人口,干擾精確斷言)
	if len(s.PlayerColonies) == 0 {
		t.Fatal("需至少一個殖民地")
	}
	startPop := s.PlayerColonies[0].Population
	startFarmers := s.PlayerColonies[0].Farmers
	startWorkers := s.PlayerColonies[0].Workers
	startScientists := s.PlayerColonies[0].Scientists
	popMax := s.PlayerColonies[0].PopMax

	// 跑足夠回合讓成長累加跨過門檻。
	for i := 0; i < 30; i++ {
		s.EndTurn()
	}

	endPop := s.PlayerColonies[0].Population
	if endPop <= startPop {
		t.Fatalf("30 回合後人口應成長:起始 %d → %d", startPop, endPop)
	}
	if endPop > popMax {
		t.Fatalf("人口 %d 超過上限 %d", endPop, popMax)
	}
	end := s.PlayerColonies[0]
	// 新人口一定要被指派職務。新增者會先試工人；若那會造成食物赤字，則依
	// assignNewColonist 的保守原版近似改派農夫，不能把「工人必增」當作不變量。
	assigned := end.Farmers + end.Workers + end.Scientists
	if assigned != end.Population {
		t.Fatalf("人口與職務數必須同步:人口 %d，農／工／科 %d/%d/%d",
			end.Population, end.Farmers, end.Workers, end.Scientists)
	}
	if end.Farmers <= startFarmers && end.Workers <= startWorkers && end.Scientists <= startScientists {
		t.Fatalf("人口成長後至少一種職務必須增加:農 %d→%d、工 %d→%d、科 %d→%d",
			startFarmers, end.Farmers, startWorkers, end.Workers, startScientists, end.Scientists)
	}
	t.Logf("殖民地0 人口 %d→%d(上限 %d),農／工／科 %d/%d/%d→%d/%d/%d",
		startPop, endPop, popMax, startFarmers, startWorkers, startScientists,
		end.Farmers, end.Workers, end.Scientists)
}

// TestPopulationCappedAtMax 驗證人口成長受 PopMax 硬上限。
func TestPopulationCappedAtMax(t *testing.T) {
	s := NewDemoSession()
	// 把第一殖民地逼近上限,跑很多回合,確認不越界。
	s.PlayerColonies[0].Population = s.PlayerColonies[0].PopMax - 1
	for i := 0; i < 200; i++ {
		s.EndTurn()
	}
	if s.PlayerColonies[0].Population > s.PlayerColonies[0].PopMax {
		t.Fatalf("人口 %d 越過上限 %d", s.PlayerColonies[0].Population, s.PlayerColonies[0].PopMax)
	}
}
