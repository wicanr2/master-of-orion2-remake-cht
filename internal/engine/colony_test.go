package engine

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestRunColonyTurn(t *testing.T) {
	cs := ColonyState{
		Population: 10, PopMax: 20,
		Farmers: 4, Workers: 4, Scientists: 2,
		FoodPerFarmer: 3, IndustryPerWorker: 5, ResearchPerScientist: 4,
		PlanetSize: gamedata.MEDIUM_PLANET, // 容忍 = 2*(2+1) = 6
	}
	got := RunColonyTurn(cs)

	// 食物:4*3=12,消耗 10,盈餘 2
	if got.Food != 12 || got.FoodConsumed != 10 || got.FoodSurplus != 2 || got.Starving {
		t.Errorf("食物錯誤:%+v", got)
	}
	// 工業:4*5=20;污染:eighths=8,產污=20,清理=(20-6)/2=7,淨=13
	if got.GrossIndustry != 20 || got.PollutingProduction != 20 ||
		got.PollutionCleanupCost != 7 || got.NetIndustry != 13 {
		t.Errorf("工業/污染錯誤:%+v", got)
	}
	// 研究:2*4=8
	if got.Research != 8 {
		t.Errorf("研究 = %d,預期 8", got.Research)
	}
	// 成長:base=sqrt(2000*10*10/20)=sqrt(10000)=100,bonus=0 → 100
	if got.PopGrowth != 100 {
		t.Errorf("成長 = %d,預期 100", got.PopGrowth)
	}
}

func TestRunColonyTurnPollutionProcessor(t *testing.T) {
	// 污染處理器:eighths 8→4,產污減半 → 清理成本降。
	cs := ColonyState{
		Population: 10, PopMax: 20, Workers: 4, IndustryPerWorker: 5,
		PlanetSize: gamedata.MEDIUM_PLANET, PollutionProcessor: true,
	}
	got := RunColonyTurn(cs)
	// gross=20,eighths=4,產污=20*4/8=10,清理=(10-6)/2=2,淨=18
	if got.PollutingProduction != 10 || got.PollutionCleanupCost != 2 || got.NetIndustry != 18 {
		t.Errorf("處理器污染錯誤:%+v", got)
	}
}

func TestRunColonyTurnStarving(t *testing.T) {
	// 食物不足 → 饑荒,成長為 0 並標 Starving。
	cs := ColonyState{
		Population: 10, PopMax: 20, Farmers: 1, FoodPerFarmer: 3,
		PlanetSize: gamedata.MEDIUM_PLANET,
	}
	got := RunColonyTurn(cs)
	if got.FoodSurplus != -7 || !got.Starving || got.PopGrowth != 0 {
		t.Errorf("饑荒處理錯誤:%+v", got)
	}
}

func TestRunColonyTurnTolerantRace(t *testing.T) {
	// Tolerant 種族:清理成本 0,淨工業 = 毛工業。
	cs := ColonyState{
		Population: 5, PopMax: 20, Workers: 6, IndustryPerWorker: 5,
		PlanetSize: gamedata.TINY_PLANET, TolerantRace: true,
	}
	got := RunColonyTurn(cs)
	if got.PollutionCleanupCost != 0 || got.NetIndustry != 30 {
		t.Errorf("Tolerant 污染錯誤:%+v", got)
	}
}

// TestRunColonyTurnFlatFood 驗證 FlatFood(殖民地整體固定食物加成,如水耕農場 p.99 +2)
// 與農夫數無關——1 農夫與 5 農夫的固定加成都應是同一個值,不像 per-farmer 欄位會隨人數放大。
func TestRunColonyTurnFlatFood(t *testing.T) {
	base := ColonyState{
		Population: 5, PopMax: 20, FoodPerFarmer: 3,
		PlanetSize: gamedata.MEDIUM_PLANET, FlatFood: 2,
	}
	small := base
	small.Farmers = 1 // 1*3+2=5
	big := base
	big.Farmers = 5 // 5*3+2=17

	gotSmall := RunColonyTurn(small)
	if gotSmall.Food != 5 {
		t.Errorf("1 農夫 + FlatFood(2) = %d,預期 5", gotSmall.Food)
	}
	gotBig := RunColonyTurn(big)
	if gotBig.Food != 17 {
		t.Errorf("5 農夫 + FlatFood(2) = %d,預期 17", gotBig.Food)
	}
}

// TestRunColonyTurnFlatIndustry 驗證 FlatIndustry 在污染縮減之前併入毛工業(手冊:固定產能
// 一樣算殖民地產能、一樣產生污染),且淨工業因此連帶反映固定值的貢獻。
func TestRunColonyTurnFlatIndustry(t *testing.T) {
	cs := ColonyState{
		Population: 10, PopMax: 20, Workers: 4, IndustryPerWorker: 5,
		PlanetSize:   gamedata.MEDIUM_PLANET, // 容忍 = 6
		FlatIndustry: 10,
	}
	got := RunColonyTurn(cs)
	// gross = 4*5+10 = 30,eighths=8,產污=30,清理=(30-6)/2=12,淨=18
	if got.GrossIndustry != 30 {
		t.Errorf("毛工業(含 FlatIndustry) = %d,預期 30", got.GrossIndustry)
	}
	if got.NetIndustry != 18 {
		t.Errorf("淨工業 = %d,預期 18", got.NetIndustry)
	}
}

// TestRunColonyTurnFlatResearch 驗證 FlatResearch(如研究實驗室 p.94 固定 +5)直接加總到研究
// 產出,與科學家數無關。
func TestRunColonyTurnFlatResearch(t *testing.T) {
	cs := ColonyState{
		Population: 5, PopMax: 20, Scientists: 2, ResearchPerScientist: 4,
		PlanetSize: gamedata.MEDIUM_PLANET, FlatResearch: 5,
	}
	got := RunColonyTurn(cs)
	if got.Research != 13 { // 2*4+5
		t.Errorf("研究(含 FlatResearch) = %d,預期 13", got.Research)
	}
}

// TestRunColonyTurnPopMaxBonusRaisesGrowthCeiling 驗證生態圈(p.99「星球人口上限 +2」)
// 對 PopMax 的直接加成會提高成長公式的上限參數,使原本「已滿(popAgg>=popMax)」的殖民地
// 重新出現成長空間。
func TestRunColonyTurnPopMaxBonusRaisesGrowthCeiling(t *testing.T) {
	full := ColonyState{
		Population: 20, PopMax: 20, Farmers: 20, FoodPerFarmer: 1,
		PlanetSize: gamedata.MEDIUM_PLANET,
	}
	got := RunColonyTurn(full)
	if got.PopGrowth != 0 {
		t.Fatalf("已滿殖民地(20/20)成長應為 0,實得 %d", got.PopGrowth)
	}

	// 生態圈使 PopMax 22(直接疊加,見 shell.applyBuildingEffect),同殖民地重新有成長空間。
	withBiosphere := full
	withBiosphere.PopMax = 22
	got2 := RunColonyTurn(withBiosphere)
	if got2.PopGrowth <= 0 {
		t.Errorf("生態圈提高 PopMax 後(20/22)應恢復成長,實得 %d", got2.PopGrowth)
	}
}

// TestRunColonyTurnFlatGrowthUntilPopMax 驗證複製中心(p.99「+0.1 單位/回合,直到達人口上限
// 為止」)的 FlatGrowth:未滿人口上限時併入成長,達到上限後不再套用(即使欄位仍非 0)。
func TestRunColonyTurnFlatGrowthUntilPopMax(t *testing.T) {
	notFull := ColonyState{
		Population: 10, PopMax: 20, FlatGrowth: 30,
		Farmers: 10, FoodPerFarmer: 1, // 食物打平消耗,避免饑荒把成長歸零,隔離 FlatGrowth 單一變數
		PlanetSize: gamedata.MEDIUM_PLANET,
	}
	got := RunColonyTurn(notFull)
	// base=sqrt(2000*10*10/20)=100,+FlatGrowth 30 = 130
	if got.PopGrowth != 130 {
		t.Errorf("未滿上限時 FlatGrowth 應併入成長:%d,預期 130", got.PopGrowth)
	}

	full := ColonyState{
		Population: 20, PopMax: 20, FlatGrowth: 30,
		Farmers: 20, FoodPerFarmer: 1,
		PlanetSize: gamedata.MEDIUM_PLANET,
	}
	got2 := RunColonyTurn(full)
	if got2.PopGrowth != 0 {
		t.Errorf("已達人口上限時 FlatGrowth 不應再套用:%d,預期 0", got2.PopGrowth)
	}
}

func TestRunColonyTurnMorale(t *testing.T) {
	// 士氣 +30%(3 格笑臉)→ 食物/工業/研究皆 ×1.3。
	cs := ColonyState{
		Population: 10, PopMax: 20, Farmers: 4, Workers: 2, Scientists: 2,
		FoodPerFarmer: 5, IndustryPerWorker: 10, ResearchPerScientist: 5,
		PlanetSize: gamedata.TINY_PLANET, TolerantRace: true, MoralePercent: 30,
	}
	got := RunColonyTurn(cs)
	if got.Food != 26 { // 20*130/100
		t.Errorf("士氣調整食物 = %d,預期 26", got.Food)
	}
	if got.GrossIndustry != 26 { // 20*1.3
		t.Errorf("士氣調整工業 = %d,預期 26", got.GrossIndustry)
	}
	if got.Research != 13 { // 10*1.3
		t.Errorf("士氣調整研究 = %d,預期 13", got.Research)
	}
}
