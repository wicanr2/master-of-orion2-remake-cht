package engine

import "testing"

// 再生反應爐 p.81 的兩句話各是一個獨立斷言,兩個都要驗。
//
//	① each unit of population generates 1 industrial production, regardless of its assigned job
//	② This increased production does not count toward the planetary pollution level
//
// ①靠「淨工業剛好多了 Population」驗;②靠「污染相關的三個數字一個都沒動」驗。
// 少了②,把它接成 FlatIndustry 也會讓①通過——而那正是接錯的地方。
func TestRecyclotronAddsPopulationProductionWithoutPollution(t *testing.T) {
	base := ColonyState{
		Population: 8, Workers: 3, PopMax: 10,
		IndustryPerWorker: 3, PlanetSize: 2,
	}
	off := RunColonyTurn(base)

	on := base
	on.Recyclotron = true
	got := RunColonyTurn(on)

	// ① 每單位人口 +1,不分職業 —— 是 Population(8)而不是 Workers(3)。
	if diff := got.NetIndustry - off.NetIndustry; diff != base.Population {
		t.Errorf("淨工業應多 %d(人口數),實際多 %d;若多的是 %d 就是接成「每工人」了",
			base.Population, diff, base.Workers)
	}
	// ② 污染三個數字一個都不能動。
	if got.PollutingProduction != off.PollutingProduction {
		t.Errorf("污染產能不該變:%d → %d", off.PollutingProduction, got.PollutingProduction)
	}
	if got.PollutionCleanupCost != off.PollutionCleanupCost {
		t.Errorf("清污成本不該變:%d → %d", off.PollutionCleanupCost, got.PollutionCleanupCost)
	}
}

// 正對照:同樣多出來的產能若接成 FlatIndustry,污染一定會跟著變
// ——這條證明上面那個「污染沒變」不是因為這個殖民地本來就不會污染。
func TestFlatIndustryDoesRaisePollutionSoTheRecyclotronCheckHasTeeth(t *testing.T) {
	base := ColonyState{
		Population: 8, Workers: 3, PopMax: 10,
		IndustryPerWorker: 3, PlanetSize: 2,
	}
	off := RunColonyTurn(base)
	flat := base
	flat.FlatIndustry += base.Population
	got := RunColonyTurn(flat)
	if got.PollutingProduction == off.PollutingProduction {
		t.Fatal("這個殖民地本來就不產生污染,上一支測試等於沒驗到——換一組參數")
	}
}

// 毛工業要把再生的那份算進去(玩家看到的產能總數),只有污染那一段不算。
func TestRecyclotronCountsTowardGrossIndustry(t *testing.T) {
	base := ColonyState{Population: 6, Workers: 2, PopMax: 10, IndustryPerWorker: 2, PlanetSize: 2}
	off := RunColonyTurn(base)
	on := base
	on.Recyclotron = true
	got := RunColonyTurn(on)
	if diff := got.GrossIndustry - off.GrossIndustry; diff != base.Population {
		t.Errorf("毛工業應多 %d,實際多 %d", base.Population, diff)
	}
}
