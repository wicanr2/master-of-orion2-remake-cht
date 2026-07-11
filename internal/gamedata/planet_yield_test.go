package gamedata

import "testing"

// TestClimateFoodPerFarmerManualValues 抽樣核對 GAME_MANUAL.pdf p.58-59「Climate」小節逐條目
// 「Base Food per Unit」原文數字(見 planet_yield.go 表註解逐頁引用)。
func TestClimateFoodPerFarmerManualValues(t *testing.T) {
	cases := []struct {
		climate PlanetClimate
		want    int
	}{
		{TOXIC, 0},    // p.58:"Farming is impossible."
		{RADIATED, 0}, // p.58:"Natural farming is impossible"
		{BARREN, 0},   // p.59:"no potential for natural farming"
		{DESERT, 1},   // p.59
		{TUNDRA, 1},   // p.59
		{OCEAN, 2},    // p.59
		{SWAMP, 2},    // p.59
		{ARID, 1},     // p.59
		{TERRAN, 2},   // p.59
		{GAIA, 3},     // p.59
	}
	for _, c := range cases {
		if got := ClimateFoodPerFarmer(c.climate); got != c.want {
			t.Errorf("ClimateFoodPerFarmer(%v)=%d,want %d", c.climate, got, c.want)
		}
	}
}

// TestClimateFoodPerFarmerOutOfRange 邊界:超出 0-9 範圍回 0,不 panic。
func TestClimateFoodPerFarmerOutOfRange(t *testing.T) {
	if got := ClimateFoodPerFarmer(-1); got != 0 {
		t.Errorf("ClimateFoodPerFarmer(-1)=%d,want 0", got)
	}
	if got := ClimateFoodPerFarmer(PlanetClimate(10)); got != 0 {
		t.Errorf("ClimateFoodPerFarmer(10)=%d,want 0", got)
	}
}

// TestMineralIndustryPerWorkerManualValues 抽樣核對 GAME_MANUAL.pdf p.56-58「Mineral Richness」
// 小節逐條目「Base Industry per Unit」原文數字,並與既有 PlanetBaseProduction(formulas.go,
// 交叉驗證自 openorion2 mineralProductionTable)保持一致(本函式只是型別包裝)。
func TestMineralIndustryPerWorkerManualValues(t *testing.T) {
	cases := []struct {
		mineral PlanetMinerals
		want    int
	}{
		{ULTRA_POOR, 1}, // p.57
		{POOR, 2},       // p.57
		{ABUNDANT, 3},   // p.57
		{RICH, 5},       // p.57
		{ULTRA_RICH, 8}, // p.58
	}
	for _, c := range cases {
		if got := MineralIndustryPerWorker(c.mineral); got != c.want {
			t.Errorf("MineralIndustryPerWorker(%v)=%d,want %d", c.mineral, got, c.want)
		}
		if got, want := MineralIndustryPerWorker(c.mineral), PlanetBaseProduction(int(c.mineral)); got != want {
			t.Errorf("MineralIndustryPerWorker(%v)=%d 應與 PlanetBaseProduction 一致(同一張表),got %d want %d", c.mineral, got, got, want)
		}
	}
}

// TestGravityPenaltyPercentManualBaseline 核對手冊 p.58「Gravity」小節原文(以無重力天賦的
// 一般種族為基準,對應 gravityPenaltyTable 的 NORMAL_G 種族列):
// Low-G 行星 -25%、Normal-G 行星 0%、Heavy-G 行星 -50%。
func TestGravityPenaltyPercentManualBaseline(t *testing.T) {
	cases := []struct {
		planetGravity PlanetGravity
		want          int
	}{
		{LOW_G, -25},   // p.58:"decrease the output of farmers, scientists, and workers by 25%"
		{NORMAL_G, 0},  // p.58:"Production rates on these planets are unaffected by gravity."
		{HEAVY_G, -50}, // p.58:"All three types of production are reduced by 50%."
	}
	for _, c := range cases {
		if got := GravityPenaltyPercent(c.planetGravity, NORMAL_G); got != c.want {
			t.Errorf("GravityPenaltyPercent(%v, NORMAL_G)=%d,want %d", c.planetGravity, got, c.want)
		}
	}
}

// TestGravityPenaltyPercentSameAffinityNoPenalty 種族天賦與行星重力一致時無懲罰(如 Low-G
// 種族住 Low-G 行星),來源 openorion2 gravityPenalties 對角線皆為 0,手冊本節未反駁。
func TestGravityPenaltyPercentSameAffinityNoPenalty(t *testing.T) {
	for _, g := range []PlanetGravity{LOW_G, NORMAL_G, HEAVY_G} {
		if got := GravityPenaltyPercent(g, g); got != 0 {
			t.Errorf("GravityPenaltyPercent(%v,%v)=%d,want 0(天賦與行星一致應無懲罰)", g, g, got)
		}
	}
}

// TestGravityPenaltyPercentOutOfRange 邊界:超出範圍回 0(保守預設,不 panic)。
func TestGravityPenaltyPercentOutOfRange(t *testing.T) {
	if got := GravityPenaltyPercent(PlanetGravity(-1), NORMAL_G); got != 0 {
		t.Errorf("GravityPenaltyPercent(-1,NORMAL_G)=%d,want 0", got)
	}
	if got := GravityPenaltyPercent(NORMAL_G, PlanetGravity(99)); got != 0 {
		t.Errorf("GravityPenaltyPercent(NORMAL_G,99)=%d,want 0", got)
	}
}

// TestGravityAdjustedProduction 核對套用公式與 MoraleProductionOutput 同型
// (base*(100+penaltyPercent)/100,無條件捨去)。
func TestGravityAdjustedProduction(t *testing.T) {
	cases := []struct{ base, penalty, want int }{
		{100, 0, 100},
		{100, -25, 75},
		{100, -50, 50},
		{7, -25, 5}, // 7*75/100=5.25 → 無條件捨去 5
	}
	for _, c := range cases {
		if got := GravityAdjustedProduction(c.base, c.penalty); got != c.want {
			t.Errorf("GravityAdjustedProduction(%d,%d)=%d,want %d", c.base, c.penalty, got, c.want)
		}
	}
}

// TestPlanetBasePopMaxManualRanges 核對 GAME_MANUAL.pdf p.55-56「Size」小節逐段給出的人口容量
// 範圍(climateFactor 取 25 與 100 兩端代入,見 PlanetBasePopMax 註解逐項推導)。
func TestPlanetBasePopMaxManualRanges(t *testing.T) {
	cases := []struct {
		size         PlanetSize
		worstClimate PlanetClimate // climateFactor=25 的氣候(如 DESERT/TUNDRA/OCEAN 等常見值)
		bestClimate  PlanetClimate // GAIA,climateFactor=100
		wantWorst    int
		wantBest     int
	}{
		{TINY_PLANET, DESERT, GAIA, 1, 5},    // 手冊「1–5」
		{SMALL_PLANET, DESERT, GAIA, 3, 10},  // 手冊「3–10」
		{MEDIUM_PLANET, DESERT, GAIA, 4, 15}, // 手冊「4–15」
		{LARGE_PLANET, DESERT, GAIA, 5, 20},  // 手冊「5–20」
		{HUGE_PLANET, DESERT, GAIA, 6, 25},   // 手冊「6–25」
	}
	for _, c := range cases {
		if got := PlanetBasePopMax(c.size, c.worstClimate); got != c.wantWorst {
			t.Errorf("PlanetBasePopMax(%v,%v)=%d,want %d(手冊下限)", c.size, c.worstClimate, got, c.wantWorst)
		}
		if got := PlanetBasePopMax(c.size, c.bestClimate); got != c.wantBest {
			t.Errorf("PlanetBasePopMax(%v,%v)=%d,want %d(手冊上限,Gaia)", c.size, c.bestClimate, got, c.wantBest)
		}
	}
}
