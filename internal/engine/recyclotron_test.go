package engine

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

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
		PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
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
		PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
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

// 食物複製機:饑荒時用產能換食物,而且只補到剛好不餓。
func TestFoodReplicatorsCoverTheDeficitButNeverCreateSurplus(t *testing.T) {
	// 人口 10 吃掉 10,農夫 2 × 每人 1 = 2 → 缺 8。
	base := ColonyState{
		Population: 10, PopMax: 12, Farmers: 2, Workers: 5,
		FoodPerFarmer: 1, IndustryPerWorker: 10, PlanetSize: 2,
		PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
	}
	off := RunColonyTurn(base)
	if off.FoodSurplus != -8 {
		t.Fatalf("測試前提:應缺 8 食物,得到 %d", off.FoodSurplus)
	}
	if off.FoodReplicated != 0 {
		t.Fatalf("沒有這棟建築時不該換算,得到 %d", off.FoodReplicated)
	}

	on := base
	on.FoodReplicators = true
	got := RunColonyTurn(on)

	if got.FoodSurplus != 0 {
		t.Errorf("複製機應剛好補平缺口(盈餘 0),得到 %d ——正數代表沒守住「as needed」", got.FoodSurplus)
	}
	if got.Starving {
		t.Error("補平之後不該還算饑荒")
	}
	if got.FoodReplicated != 8 {
		t.Errorf("應換出 8 單位食物,得到 %d", got.FoodReplicated)
	}
	if diff := off.NetIndustry - got.NetIndustry; diff != 16 {
		t.Errorf("換 8 食物應花 16 產能(2:1),實際少了 %d", diff)
	}
}

// 產能不夠時只補得了一部分,而且不會把產能扣成負的。
func TestFoodReplicatorsClampToAvailableProduction(t *testing.T) {
	base := ColonyState{
		Population: 10, PopMax: 12, Farmers: 1, Workers: 1,
		FoodPerFarmer: 1, IndustryPerWorker: 3, PlanetSize: 2,
		PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
		FoodReplicators: true,
	}
	got := RunColonyTurn(base)
	if got.NetIndustry < 0 {
		t.Errorf("淨產能不該被扣成負數,得到 %d", got.NetIndustry)
	}
	if got.FoodSurplus >= 0 {
		t.Errorf("產能不足時應仍然是赤字,得到盈餘 %d", got.FoodSurplus)
	}
	if got.FoodReplicated*2 > got.FoodReplicated*2+got.NetIndustry {
		t.Error("換算花掉的產能超出可用量")
	}
}

// 食物有盈餘時完全不動——否則會變成「換滿食物 → 餘糧出售換 BC」的印鈔機。
func TestFoodReplicatorsIdleWhenThereIsSurplus(t *testing.T) {
	base := ColonyState{
		Population: 4, PopMax: 12, Farmers: 6, Workers: 4,
		FoodPerFarmer: 3, IndustryPerWorker: 5, PlanetSize: 2,
		PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
	}
	off := RunColonyTurn(base)
	if off.FoodSurplus <= 0 {
		t.Fatalf("測試前提:應該有食物盈餘,得到 %d", off.FoodSurplus)
	}
	on := base
	on.FoodReplicators = true
	got := RunColonyTurn(on)
	if got.FoodReplicated != 0 {
		t.Errorf("有盈餘時不該換算,得到 %d", got.FoodReplicated)
	}
	if got.NetIndustry != off.NetIndustry || got.FoodSurplus != off.FoodSurplus {
		t.Errorf("有盈餘時應完全不動:產能 %d→%d、盈餘 %d→%d",
			off.NetIndustry, got.NetIndustry, off.FoodSurplus, got.FoodSurplus)
	}
}

// Cybernetic 奇數人口會留下半食物缺口；複製機應能補半單位，而不是先截成零。
func TestFoodReplicatorsCoverCyberneticHalfFood(t *testing.T) {
	cs := ColonyState{
		Population: 3, PopMax: 10, Farmers: 1, Workers: 1,
		FoodPerFarmer: 1, IndustryPerWorker: 4, PlanetSize: 2,
		PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
		Cybernetic: true, FoodReplicators: true,
	}
	got := RunColonyTurn(cs)
	if got.FoodSurplusHalf != 0 || got.Starving {
		t.Fatalf("半食物複製後應剛好補平,got %+v", got)
	}
	if got.FoodReplicatedHalf != 1 || got.FoodReplicated != 0 {
		t.Fatalf("應換出 1 半食物而非虛構 1 完整食物,got %+v", got)
	}
	if got.FoodReplicatorCostHalfBC != gamedata.FoodReplicatorBCHalfPerHalfFood {
		t.Fatalf("半食物成本應為 1 半 BC,got %d", got.FoodReplicatorCostHalfBC)
	}
}
