package engine

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func originalAIJobTestColony() ColonyState {
	return ColonyState{
		Population: 5, Farmers: 2, Workers: 2, Scientists: 1,
		FoodPerFarmer: 2, IndustryPerWorker: 3, ResearchPerScientist: 3,
		OwnerRaceProfileKnown: true, OwnerRaceSlotKnown: true, OwnerRaceSlot: 0,
		PlanetGravity: gamedata.NORMAL_G, RaceGravityKnown: true, RaceGravity: gamedata.NORMAL_G,
		PopulationGroups: []PopulationGroup{{
			RaceSlot: 0, RaceSlotKnown: true, ProfileKnown: true, Gravity: gamedata.NORMAL_G,
			Farmers: 2, Workers: 2, Scientists: 1,
		}},
	}
}

func TestApplyOriginalAIUnblockadedJobsConsumesNormalCandidates(t *testing.T) {
	c := originalAIJobTestColony()
	got, ok := ApplyOriginalAIUnblockadedJobs(PlayerState{}, []ColonyState{c}, OriginalAIJobContext{
		Personality: ai.PersonalityErratic, ColonyFoodHalf: []int{4}, ColonyFoodHalfKnown: []bool{true},
	})
	if !ok {
		t.Fatal("完整 population groups 與 +0xDD 應可走原版未封鎖路徑")
	}
	if co := RunColonyTurn(got[0]); co.FoodSurplusHalf < 0 {
		t.Fatalf("sub_D6AD4 下游 pass 應在職務平衡後補足農夫：colony=%+v output=%+v", got[0], co)
	}
	if got[0].Workers+got[0].Scientists == 0 {
		t.Fatalf("餵飽後仍應保留工業／研究人口：%+v", got[0])
	}
	if !PopulationGroupsComplete(got[0]) || got[0].PopulationGroups[0].Farmers != got[0].Farmers {
		t.Fatalf("總職務與逐 race 群組必須同步：%+v", got[0].PopulationGroups)
	}
}

func TestApplyOriginalAIJobsAdditionalFarmersCoverEveryUntransportedColony(t *testing.T) {
	home := originalAIJobTestColony()
	outpost := originalAIJobTestColony()
	outpost.Population, outpost.Farmers, outpost.Workers, outpost.Scientists = 1, 0, 1, 0
	outpost.PopulationGroups[0].Farmers = 0
	outpost.PopulationGroups[0].Workers = 1
	outpost.PopulationGroups[0].Scientists = 0
	got, ok := ApplyOriginalAIJobs(PlayerState{}, []ColonyState{home, outpost}, OriginalAIJobContext{
		ColonyFoodHalf: []int{4, 4}, ColonyFoodHalfKnown: []bool{true, true}, ColonyBlockaded: []bool{false, false},
	})
	if !ok {
		t.Fatal("完整未封鎖輸入不應 fallback")
	}
	for i := range got {
		if co := RunColonyTurn(got[i]); co.FoodSurplusHalf < 0 {
			t.Fatalf("零貨運艦時不得讓未封鎖殖民地留在饑荒：i=%d colony=%+v output=%+v", i, got[i], co)
		}
	}
}

func TestApplyOriginalAIJobsUsesAvailableFoodTransport(t *testing.T) {
	surplus := originalAIJobTestColony()
	surplus.Population, surplus.Farmers, surplus.Workers, surplus.Scientists = 1, 1, 0, 0
	surplus.FoodPerFarmer = 5
	surplus.PopulationGroups[0].Farmers = 1
	surplus.PopulationGroups[0].Workers = 0
	surplus.PopulationGroups[0].Scientists = 0
	deficit := originalAIJobTestColony()
	deficit.Population, deficit.Farmers, deficit.Workers, deficit.Scientists = 1, 0, 1, 0
	deficit.PopulationGroups[0].Farmers = 0
	deficit.PopulationGroups[0].Workers = 1
	deficit.PopulationGroups[0].Scientists = 0
	got, pressure, ok := ApplyOriginalAIJobsWithTransport(
		PlayerState{ActiveFreighters: 2}, []ColonyState{surplus, deficit}, OriginalAIJobContext{
			ColonyFoodHalf: []int{5, 5}, ColonyFoodHalfKnown: []bool{true, true},
			ColonyBlockaded: []bool{false, false},
		})
	if !ok || pressure {
		t.Fatalf("充足運力不應回退或觸發造艦壓力：ok=%v pressure=%v", ok, pressure)
	}
	if got[1].Farmers != 0 || got[1].Workers != 1 {
		t.Fatalf("可跨殖民地運糧時不應強迫缺糧殖民地本地自給：%+v", got[1])
	}
}

func TestApplyOriginalAIUnblockadedJobsPreservesSpecialAndPrisoner(t *testing.T) {
	c := originalAIJobTestColony()
	c.Population = 7
	c.Farmers = 3
	c.Workers = 3
	c.Scientists = 1
	c.PopulationGroups[0].PrisonerFarmers = 1
	c.PopulationGroups = append(c.PopulationGroups, PopulationGroup{
		RaceSlot: gamedata.NativeColonistSlot, RaceSlotKnown: true, ProfileKnown: true,
		Farmers: 1, Workers: 1, Gravity: gamedata.NORMAL_G,
	})
	got, ok := ApplyOriginalAIUnblockadedJobs(PlayerState{}, []ColonyState{c}, OriginalAIJobContext{
		Personality: ai.PersonalityHonorable, LateTech: true,
		ColonyFoodHalf: []int{4}, ColonyFoodHalfKnown: []bool{true},
	})
	if !ok {
		t.Fatal("完整輸入不應 fallback")
	}
	native := got[0].PopulationGroups[1]
	if native.Farmers != 1 || native.Workers != 1 || native.Scientists != 0 {
		t.Fatalf("Natives 前置區不得被改職：%+v", native)
	}
	g := got[0].PopulationGroups[0]
	if g.PrisonerFarmers+g.PrisonerWorkers+g.PrisonerScientists != 1 {
		t.Fatalf("改職不得遺失 PRISONER：%+v", g)
	}
}

func TestApplyOriginalAIUnblockadedJobsRejectsUnknownTypedInput(t *testing.T) {
	valid := originalAIJobTestColony()
	invalid := originalAIJobTestColony()
	invalid.PopulationGroups = nil
	before := valid.PopulationGroups[0]
	if _, ok := ApplyOriginalAIUnblockadedJobs(PlayerState{}, []ColonyState{valid, invalid}, OriginalAIJobContext{
		ColonyFoodHalf: []int{4, 4}, ColonyFoodHalfKnown: []bool{true, true},
	}); ok {
		t.Fatal("缺逐種族資料不得冒稱原版職務路徑")
	}
	if valid.PopulationGroups[0] != before {
		t.Fatalf("後段驗證失敗不得部分修改呼叫端：before=%+v after=%+v", before, valid.PopulationGroups[0])
	}
}

func TestApplyOriginalAIJobsBlockadedFeedsThenUsesWorkers(t *testing.T) {
	c := originalAIJobTestColony()
	got, ok := ApplyOriginalAIJobs(PlayerState{}, []ColonyState{c}, OriginalAIJobContext{
		Personality: ai.PersonalityErratic, ColonyFoodHalf: []int{4},
		ColonyFoodHalfKnown: []bool{true}, ColonyBlockaded: []bool{true},
	})
	if !ok {
		t.Fatal("完整封鎖殖民地輸入不應 fallback")
	}
	co := RunColonyTurn(got[0])
	if co.FoodHalf < co.FoodConsumedHalf {
		t.Fatalf("可耕作封鎖殖民地應先補足食物：colony=%+v output=%+v", got[0], co)
	}
	if got[0].Scientists != 0 || got[0].Farmers+got[0].Workers != got[0].Population {
		t.Fatalf("糧食足夠後其餘一般人口應為工人：%+v", got[0])
	}
}

func TestApplyOriginalAIJobsBlockadedNoFarmingUsesEventBranch(t *testing.T) {
	c := originalAIJobTestColony()
	ctx := OriginalAIJobContext{ColonyFoodHalf: []int{0}, ColonyFoodHalfKnown: []bool{true}, ColonyBlockaded: []bool{true}}
	got, ok := ApplyOriginalAIJobs(PlayerState{}, []ColonyState{c}, ctx)
	if !ok || got[0].Workers != c.Population || got[0].Farmers != 0 || got[0].Scientists != 0 {
		t.Fatalf("無農業且事件 filter 未命中應全轉工人：ok=%v colony=%+v", ok, got)
	}
	c.ResearchDiverted = true
	got, ok = ApplyOriginalAIJobs(PlayerState{}, []ColonyState{c}, ctx)
	if !ok || got[0].Scientists != c.Population || got[0].Farmers != 0 || got[0].Workers != 0 {
		t.Fatalf("無農業且事件 filter 命中應全轉科學家：ok=%v colony=%+v", ok, got)
	}
}
