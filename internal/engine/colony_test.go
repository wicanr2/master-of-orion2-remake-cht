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
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, // 容忍 = 2*(2+1) = 6
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
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, PollutionProcessor: true,
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
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
	}
	got := RunColonyTurn(cs)
	if got.FoodSurplus != -7 || !got.Starving || got.PopGrowth != 0 {
		t.Errorf("饑荒處理錯誤:%+v", got)
	}
}

func TestRunColonyTurnLithovoreNeedsNoFood(t *testing.T) {
	// 手冊規則:食岩族不需要食物。即使沒有農夫,也不饑荒且仍可正常成長。
	cs := ColonyState{
		Population: 10, PopMax: 20, Workers: 6, Scientists: 4,
		FoodPerFarmer: 3, IndustryPerWorker: 5, ResearchPerScientist: 4,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
		Lithovore: true,
	}
	got := RunColonyTurn(cs)
	if got.Food != 0 || got.FoodConsumed != 0 || got.FoodSurplus != 0 || got.Starving {
		t.Fatalf("食岩族應免食物消耗與饑荒,得到 %+v", got)
	}
	if got.PopGrowth == 0 {
		t.Fatal("食岩族在沒有食物時仍應能依正常人口公式成長")
	}
}

// TestRunColonyTurnCyberneticUsesHalfLedger 驗證手冊所述「每人口半食物、另以半生產力補足」
// 不是把 0.5 直接截成 0，而是落在可持久追蹤的半單位帳本。整數欄位只作既有 UI 相容值。
func TestRunColonyTurnCyberneticUsesHalfLedger(t *testing.T) {
	cs := ColonyState{
		Population: 5, PopMax: 20, Farmers: 5, Workers: 5,
		FoodPerFarmer: 2, IndustryPerWorker: 5,
		PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G,
		MineralRichness: gamedata.ABUNDANT, TolerantRace: true, Cybernetic: true,
	}
	got := RunColonyTurn(cs)
	// 食物產出 10 = 20 半單位；半機械人口消耗 5 半單位，餘 15 半單位。
	if got.FoodHalf != 20 || got.FoodConsumedHalf != 5 || got.FoodSurplusHalf != 15 {
		t.Fatalf("半機械食物帳本錯誤:%+v", got)
	}
	if got.FoodConsumed != 2 || got.FoodSurplus != 7 || got.Starving {
		t.Errorf("整數相容食物欄位錯誤:%+v", got)
	}
	// 毛/淨工業 25；扣掉 5 個半單位後是 45 半單位 = 22.5。
	if got.IndustryConsumedHalf != 5 || got.NetIndustryHalf != 45 || got.NetIndustry != 22 {
		t.Errorf("半機械生產帳本錯誤:%+v", got)
	}
}

func TestRunColonyTurnTolerantRace(t *testing.T) {
	// Tolerant 種族:清理成本 0,淨工業 = 毛工業。
	cs := ColonyState{
		Population: 5, PopMax: 20, Workers: 6, IndustryPerWorker: 5,
		PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, TolerantRace: true,
	}
	got := RunColonyTurn(cs)
	if got.PollutionCleanupCost != 0 || got.NetIndustry != 30 {
		t.Errorf("Tolerant 污染錯誤:%+v", got)
	}
}

func TestRunColonyTurnUsesExactPrisonerJobs(t *testing.T) {
	base := ColonyState{
		Population: 3, PopMax: 10, Farmers: 1, Workers: 1, Scientists: 1,
		FoodPerFarmer: 4, IndustryPerWorker: 4, ResearchPerScientist: 4,
		UnassimilatedPop: 1, PlanetSize: gamedata.MEDIUM_PLANET,
		PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
	}
	farmer := base
	farmer.UnassimilatedFarmers = 1
	worker := base
	worker.UnassimilatedWorkers = 1
	if got := RunColonyTurn(farmer); got.Food != 3 || got.Research != 4 {
		t.Fatalf("農夫 prisoner 只應降低食物：%+v", got)
	}
	if got := RunColonyTurn(worker); got.Food != 4 || got.GrossIndustry != 3 || got.Research != 4 {
		t.Fatalf("工人 prisoner 只應降低工業：%+v", got)
	}
}

func TestRunColonyTurnUsesKnownRaceGravity(t *testing.T) {
	base := ColonyState{
		Population: 2, PopMax: 10, Farmers: 2, FoodPerFarmer: 4,
		PlanetGravity: gamedata.HEAVY_G, PlanetSize: gamedata.MEDIUM_PLANET,
		MineralRichness: gamedata.ABUNDANT,
	}
	unknown := RunColonyTurn(base)
	if unknown.Food != 4 { // 舊狀態未知 race gravity → Normal-G 在 Heavy-G 上 -50%。
		t.Fatalf("未知種族重力 fallback 食物 = %d，want 4", unknown.Food)
	}
	high := base
	high.RaceGravity, high.RaceGravityKnown = gamedata.HEAVY_G, true
	if got := RunColonyTurn(high).Food; got != 8 {
		t.Fatalf("High-G 種族在 Heavy-G 食物 = %d，want 8", got)
	}
	low := base
	low.PlanetGravity = gamedata.NORMAL_G
	low.RaceGravity, low.RaceGravityKnown = gamedata.LOW_G, true
	if got := RunColonyTurn(low).Food; got != 6 {
		t.Fatalf("Low-G 種族在 Normal-G 食物 = %d，want 6", got)
	}
}

func TestRunColonyTurnGravityGeneratorOverridesRaceMismatch(t *testing.T) {
	cs := ColonyState{
		Population: 1, PopMax: 10, Workers: 1, IndustryPerWorker: 8,
		PlanetGravity: gamedata.HEAVY_G, RaceGravity: gamedata.LOW_G, RaceGravityKnown: true,
		NormalizeGravity: true, PlanetSize: gamedata.MEDIUM_PLANET, MineralRichness: gamedata.ABUNDANT,
	}
	if got := RunColonyTurn(cs).GrossIndustry; got != 8 {
		t.Fatalf("重力產生器應消除 mismatch，毛工業 = %d，want 8", got)
	}
}

func TestRunColonyTurnMixedRaceUsesEachGroupsGravity(t *testing.T) {
	cs := ColonyState{
		Population: 2, PopMax: 10, Workers: 2, IndustryPerWorker: 4,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.HEAVY_G,
		OwnerRaceProfileKnown: true, OwnerRaceSlot: 0, OwnerRaceSlotKnown: true,
		PopulationGroups: []PopulationGroup{
			{RaceSlot: 0, RaceSlotKnown: true, Workers: 1, Gravity: gamedata.LOW_G, ProfileKnown: true},
			{RaceSlot: 1, RaceSlotKnown: true, Workers: 1, Gravity: gamedata.HEAVY_G, ProfileKnown: true},
		},
	}
	out := RunColonyTurn(cs)
	if out.GrossIndustry != 6 { // Low-G:4×50%=2；High-G:4×100%=4。
		t.Fatalf("混居逐群重力毛工業 = %d，want 6", out.GrossIndustry)
	}
	cs.NormalizeGravity = true
	if got := RunColonyTurn(cs).GrossIndustry; got != 8 {
		t.Fatalf("重力產生器應同時解除兩群懲罰，got %d want 8", got)
	}
}

func TestRunColonyTurnMixedRaceKeepsScientistTrait(t *testing.T) {
	cs := ColonyState{
		Population: 2, PopMax: 10, Scientists: 2, ResearchPerScientist: 3,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G,
		OwnerRaceProfileKnown: true, OwnerRaceSlot: 0, OwnerRaceSlotKnown: true,
		PopulationGroups: []PopulationGroup{
			{RaceSlot: 0, RaceSlotKnown: true, Scientists: 1, Gravity: gamedata.NORMAL_G, ProfileKnown: true},
			{RaceSlot: 1, RaceSlotKnown: true, Scientists: 1, ResearchBonus: 2, Gravity: gamedata.NORMAL_G, ProfileKnown: true},
		},
	}
	if got := RunColonyTurn(cs).Research; got != 8 {
		t.Fatalf("一般 3 + 席隆 5 應為 8 RP，got %d", got)
	}
}

func TestRunColonyTurnSpecialGroupsAndPerSlotGrowth(t *testing.T) {
	cs := ColonyState{
		Population: 4, PopMax: 20, Farmers: 2, Workers: 1, Scientists: 1,
		FoodPerFarmer: 3, IndustryPerWorker: 2, ResearchPerScientist: 1,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.HEAVY_G,
		OwnerRaceProfileKnown: true, OwnerRaceSlot: 0, OwnerRaceSlotKnown: true,
		PopulationGroups: []PopulationGroup{
			{RaceSlot: 0, RaceSlotKnown: true, Farmers: 1, Gravity: gamedata.HEAVY_G, ProfileKnown: true},
			{RaceSlot: gamedata.AndroidColonistSlot, RaceSlotKnown: true, Workers: 1, Scientists: 1,
				FoodBonus: 6, IndustryBonus: 3, ResearchBonus: 3, GravityImmune: true,
				Gravity: gamedata.NORMAL_G, GrowthBonusPercent: 50, ProfileKnown: true},
			{RaceSlot: gamedata.NativeColonistSlot, RaceSlotKnown: true, Farmers: 1,
				FoodBonus: 4, GravityImmune: true, Gravity: gamedata.NORMAL_G, ProfileKnown: true},
		},
	}
	out := RunColonyTurn(cs)
	if out.Food != 10 || out.GrossIndustry != 5 || out.Research != 4 {
		t.Fatalf("特殊人口產出 = 食物/工業/研究 %d/%d/%d，want 10/5/4", out.Food, out.GrossIndustry, out.Research)
	}
	if out.PopulationGroupGrowthCount != 3 {
		t.Fatalf("逐槽成長數 = %d，want 3", out.PopulationGroupGrowthCount)
	}
	if out.PopulationGroupGrowth[0] <= 0 || out.PopulationGroupGrowth[1] != 0 || out.PopulationGroupGrowth[2] != 0 {
		t.Fatalf("一般 slot 應自然成長，Android／Natives 不進正成長 pass，got %v", out.PopulationGroupGrowth)
	}
}

func TestPopulationConsumptionAndShortageOriginalPriority(t *testing.T) {
	cs := ColonyState{
		Population: 4, PopMax: 20, Workers: 4,
		OwnerRaceProfileKnown: true, OwnerRaceSlot: 0, OwnerRaceSlotKnown: true,
		PopulationGroups: []PopulationGroup{
			{RaceSlot: 0, RaceSlotKnown: true, Workers: 1, ProfileKnown: true},
			{RaceSlot: 1, RaceSlotKnown: true, Workers: 1, Cybernetic: true, ProfileKnown: true},
			{RaceSlot: gamedata.AndroidColonistSlot, RaceSlotKnown: true, Workers: 1, ProfileKnown: true},
			{RaceSlot: gamedata.NativeColonistSlot, RaceSlotKnown: true, Workers: 1, ProfileKnown: true},
		},
	}
	consumption := colonyPopulationConsumption(cs)
	if consumption.foodTotal() != 5 || consumption.industryTotal() != 3 {
		t.Fatalf("逐族半單位消耗 = food %d industry %d，want 5/3", consumption.foodTotal(), consumption.industryTotal())
	}
	// 2 半食物先供 owner；2 半工業必須先完整供 Android，外族 Cybernetic 仍短缺食物與工業。
	rates := populationShortageGrowth(cs, consumption, 2, 2)
	if rates[0] != 0 || rates[1] != -50 || rates[gamedata.AndroidColonistSlot] != 0 ||
		rates[gamedata.NativeColonistSlot] != -50 {
		t.Fatalf("原版短缺優先序 signed rates = %v，want slot0/1/8/9 = 0/-50/0/-50", rates)
	}
}

func TestPopulationConsumptionCyberneticPrecedesLithovore(t *testing.T) {
	cs := ColonyState{
		Population: 1, PopMax: 10, Farmers: 1,
		OwnerRaceProfileKnown: true, OwnerRaceSlot: 0, OwnerRaceSlotKnown: true,
		PopulationGroups: []PopulationGroup{{RaceSlot: 0, RaceSlotKnown: true, Farmers: 1,
			Cybernetic: true, Lithovore: true, ProfileKnown: true}},
	}
	got := colonyPopulationConsumption(cs)
	if got.foodTotal() != 1 || got.industryTotal() != 1 {
		t.Fatalf("Cybernetic+Lithovore 依原版分支應為 1 半食物 + 1 半工業，got %d/%d",
			got.foodTotal(), got.industryTotal())
	}
}

func TestRunColonyTurnSignedGrowthOffsetsPositiveGrowth(t *testing.T) {
	cs := ColonyState{
		Population: 2, PopMax: 20, Farmers: 2, FoodPerFarmer: 0,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G,
		OwnerRaceProfileKnown: true, OwnerRaceSlot: 0, OwnerRaceSlotKnown: true,
		PopulationGroups: []PopulationGroup{{RaceSlot: 0, RaceSlotKnown: true, Farmers: 2,
			Gravity: gamedata.NORMAL_G, ProfileKnown: true}},
	}
	out := RunColonyTurn(cs)
	base := gamedata.ColonyGrowth(gamedata.ColonyBaseGrowth(2, 2, 20), 0)
	want := base - 100 // 4 個未滿足半食物 × -25。
	if out.PopulationGroupGrowthCount != 1 || out.PopulationGroupGrowth[0] != want || out.PopGrowth != want {
		t.Fatalf("signed 成長應為正成長 %d + 短缺 -100 = %d，got %+v", base, want, out)
	}
}

func TestRunColonyTurnOldSavePrisonersFallbackToRatio(t *testing.T) {
	cs := ColonyState{
		Population: 4, PopMax: 10, Workers: 4, IndustryPerWorker: 4,
		UnassimilatedPop: 2, PlanetSize: gamedata.MEDIUM_PLANET,
		PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
	}
	if got := RunColonyTurn(cs).GrossIndustry; got != 14 {
		t.Fatalf("舊存檔比例 fallback 毛工業 = %d，want 14", got)
	}
}

// TestRunColonyTurnFlatFood 驗證 FlatFood(殖民地整體固定食物加成,如水耕農場 p.99 +2)
// 與農夫數無關——1 農夫與 5 農夫的固定加成都應是同一個值,不像 per-farmer 欄位會隨人數放大。
func TestRunColonyTurnFlatFood(t *testing.T) {
	base := ColonyState{
		Population: 5, PopMax: 20, FoodPerFarmer: 3,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, FlatFood: 2,
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
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, // 容忍 = 6
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
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, FlatResearch: 5,
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
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
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
		Population: 10, PopMax: 20, FlatGrowth: gamedata.CloningCenterGrowthPoints,
		Farmers: 10, FoodPerFarmer: 1, // 食物打平消耗,避免饑荒把成長歸零,隔離 FlatGrowth 單一變數
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
	}
	got := RunColonyTurn(notFull)
	// base=sqrt(2000*10*10/20)=100,+FlatGrowth 100 = 200
	if got.PopGrowth != 200 {
		t.Errorf("未滿上限時 FlatGrowth 應併入成長:%d,預期 200", got.PopGrowth)
	}

	full := ColonyState{
		Population: 20, PopMax: 20, FlatGrowth: gamedata.CloningCenterGrowthPoints,
		Farmers: 20, FoodPerFarmer: 1,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
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
		PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, TolerantRace: true, MoralePercent: 30,
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

func TestRunColonyTurnGovernmentJobBonuses(t *testing.T) {
	base := ColonyState{Population: 6, PopMax: 20, Farmers: 2, Workers: 2, Scientists: 2,
		FoodPerFarmer: 10, IndustryPerWorker: 10, ResearchPerScientist: 10,
		PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G,
		MineralRichness: gamedata.ABUNDANT, TolerantRace: true}
	unity := base
	unity.GovernmentFoodBonusPercent, unity.GovernmentIndustryBonusPercent = 50, 50
	got := RunColonyTurn(unity)
	if got.Food != 30 || got.GrossIndustry != 30 || got.Research != 20 {
		t.Fatalf("統一只應提升食物／工業：food=%d industry=%d research=%d", got.Food, got.GrossIndustry, got.Research)
	}
	federation := base
	federation.GovernmentResearchBonusPercent = 75
	got = RunColonyTurn(federation)
	if got.Food != 20 || got.GrossIndustry != 20 || got.Research != 35 {
		t.Fatalf("聯邦只應提升研究：food=%d industry=%d research=%d", got.Food, got.GrossIndustry, got.Research)
	}
}

// TestRunColonyTurnGravityHeavyPenalty 驗證 Heavy-G 行星(GAME_MANUAL.pdf p.58:「All three
// types of production are reduced by 50%」)對食物/工業/研究三種 per-worker 產出都打對折,
// 固定加成（此測試未設 Flat*，略）不受影響；未提供種族重力時採 NORMAL_G 相容 fallback。
// (見 colonyGravityPenaltyPercent 註解),故 HEAVY_G 行星懲罰 = gravityPenaltyTable[NORMAL_G][HEAVY_G] = -50。
func TestRunColonyTurnGravityHeavyPenalty(t *testing.T) {
	cs := ColonyState{
		Population: 10, PopMax: 20, Farmers: 4, Workers: 4, Scientists: 2,
		FoodPerFarmer: 5, IndustryPerWorker: 10, ResearchPerScientist: 5,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.HEAVY_G, MineralRichness: gamedata.ABUNDANT,
	}
	got := RunColonyTurn(cs)
	if got.Food != 10 { // 4*5=20,×50% = 10
		t.Errorf("Heavy-G 食物 = %d,預期 10", got.Food)
	}
	if got.GrossIndustry != 20 { // 4*10=40,×50% = 20
		t.Errorf("Heavy-G 毛工業 = %d,預期 20", got.GrossIndustry)
	}
	if got.Research != 5 { // 2*5=10,×50% = 5
		t.Errorf("Heavy-G 研究 = %d,預期 5", got.Research)
	}
}

// TestRunColonyTurnGravityNormalizeGravityCancelsPenalty 驗證行星重力產生器
// (NormalizeGravity=true,p.104「正常化至 Normal-G,消除 Low-G/Heavy-G 負面效果」)接線後
// 真的會讓同一顆 Heavy-G 行星的重力懲罰歸零——這是本輪把「無效旗標」接成「有效旗標」的
// 直接證明:與 TestRunColonyTurnGravityHeavyPenalty 同一組 cs,只差 NormalizeGravity。
func TestRunColonyTurnGravityNormalizeGravityCancelsPenalty(t *testing.T) {
	cs := ColonyState{
		Population: 10, PopMax: 20, Farmers: 4, Workers: 4, Scientists: 2,
		FoodPerFarmer: 5, IndustryPerWorker: 10, ResearchPerScientist: 5,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.HEAVY_G, MineralRichness: gamedata.ABUNDANT,
		NormalizeGravity: true,
	}
	got := RunColonyTurn(cs)
	if got.Food != 20 { // 無懲罰:4*5=20
		t.Errorf("重力產生器後食物 = %d,預期 20(懲罰應歸零)", got.Food)
	}
	if got.GrossIndustry != 40 { // 無懲罰:4*10=40
		t.Errorf("重力產生器後毛工業 = %d,預期 40(懲罰應歸零)", got.GrossIndustry)
	}
	if got.Research != 10 { // 無懲罰:2*5=10
		t.Errorf("重力產生器後研究 = %d,預期 10(懲罰應歸零)", got.Research)
	}
}

// TestRunColonyTurnGravityNormalGNoPenalty 驗證 Normal-G 行星(手冊:「Production rates on
// these planets are unaffected by gravity」)不受重力調整,產出與零值 pct 相同。
func TestRunColonyTurnGravityNormalGNoPenalty(t *testing.T) {
	cs := ColonyState{
		Population: 10, PopMax: 20, Farmers: 4, Workers: 4, Scientists: 2,
		FoodPerFarmer: 5, IndustryPerWorker: 10, ResearchPerScientist: 5,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
	}
	got := RunColonyTurn(cs)
	if got.Food != 20 || got.GrossIndustry != 40 || got.Research != 10 {
		t.Errorf("Normal-G 不應受重力影響:%+v", got)
	}
}

// TestRunColonyTurnGravityAndMoraleCombinedPercent 驗證士氣與重力先加總成單一百分點再套一次
// 公式(colonyFood/RunColonyTurn 註解說明的順序選擇):Low-G(-25%)+ 士氣 +10% = 合併 -15%,
// 而非兩次連續除法(20*75/100=15,15*110/100=16,與合併版 20*85/100=17 不同)——用這組數字
// 確認 remake 採「先加總、再套一次」而非「連續套用兩次」。
func TestRunColonyTurnGravityAndMoraleCombinedPercent(t *testing.T) {
	cs := ColonyState{
		Population: 10, PopMax: 20, Workers: 4, IndustryPerWorker: 5,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.LOW_G, MineralRichness: gamedata.ABUNDANT,
		MoralePercent: 10, TolerantRace: true, // Tolerant 避開污染清理,單純看毛工業
	}
	got := RunColonyTurn(cs)
	// base=4*5=20,合併百分點=10+(-25)=-15 → 20*85/100=17(無條件捨去)
	if got.GrossIndustry != 17 {
		t.Errorf("士氣+重力合併毛工業 = %d,預期 17(合併 -15%% 一次套用,非連續兩次除法)", got.GrossIndustry)
	}
}
