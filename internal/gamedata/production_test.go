package gamedata

import "testing"

func TestProdConstants(t *testing.T) {
	// GAME_MANUAL.pdf 直接給出的常數,防止之後被誤改。
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"ProdWorkerMinimum", ProdWorkerMinimum, 1},
		{"ProdAutomatedFactoryPerWorkerBonus", ProdAutomatedFactoryPerWorkerBonus, 1},
		{"ProdAutomatedFactoryFlatBonus", ProdAutomatedFactoryFlatBonus, 5},
		{"ProdDeepCoreMinePerWorkerBonus", ProdDeepCoreMinePerWorkerBonus, 3},
		{"ProdDeepCoreMineFlatBonus", ProdDeepCoreMineFlatBonus, 15},
		{"ProdRecyclotronPerPopulation", ProdRecyclotronPerPopulation, 1},
		{"ProdMicroliteConstructionPerWorkerBonus", ProdMicroliteConstructionPerWorkerBonus, 1},
		{"ProdAlienUncooperativeNumerator", ProdAlienUncooperativeNumerator, 3},
		{"ProdAlienUncooperativeDenominator", ProdAlienUncooperativeDenominator, 4},
		{"PollutionNanoDisassemblersMultiplier", PollutionNanoDisassemblersMultiplier, 2},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d,預期 %d", c.name, c.got, c.want)
		}
	}
}

func TestProdWorkerOutput(t *testing.T) {
	cases := []struct{ base, want int }{
		{0, 1},  // 低於下限 → 1
		{1, 1},  // 剛好下限
		{3, 3},  // 高於下限 → 原值
		{-5, 1}, // 負值 → 下限
	}
	for _, c := range cases {
		if got := ProdWorkerOutput(c.base); got != c.want {
			t.Errorf("ProdWorkerOutput(%d) = %d,預期 %d", c.base, got, c.want)
		}
	}
}

func TestProdRoboticFactoryBonus(t *testing.T) {
	// GAME_MANUAL.pdf: +5 Ultra Poor, +8 Poor, +10 Abundant, +15 Rich, +20 Ultra Rich。
	cases := []struct{ minerals, want int }{
		{int(ULTRA_POOR), 5},
		{int(POOR), 8},
		{int(ABUNDANT), 10},
		{int(RICH), 15},
		{int(ULTRA_RICH), 20},
		{-1, 0}, // 超出範圍
		{5, 0},  // 超出範圍
	}
	for _, c := range cases {
		if got := ProdRoboticFactoryBonus(c.minerals); got != c.want {
			t.Errorf("ProdRoboticFactoryBonus(%d) = %d,預期 %d", c.minerals, got, c.want)
		}
	}
}

func TestProdAlienWorkerOutput(t *testing.T) {
	// base * 3/4,向下取整。
	cases := []struct{ base, want int }{
		{4, 3},  // 4*3/4=3
		{10, 7}, // 10*3/4=7.5→7
		{0, 0},
		{8, 6}, // 8*3/4=6
	}
	for _, c := range cases {
		if got := ProdAlienWorkerOutput(c.base); got != c.want {
			t.Errorf("ProdAlienWorkerOutput(%d) = %d,預期 %d", c.base, got, c.want)
		}
	}
}

func TestPollutionTolerance(t *testing.T) {
	// tolerance = 2*(size+1);手冊範例:medium(size class 3)→6。
	cases := []struct {
		size PlanetSize
		want int
	}{
		{TINY_PLANET, 2},
		{SMALL_PLANET, 4},
		{MEDIUM_PLANET, 6}, // 手冊範例
		{LARGE_PLANET, 8},
		{HUGE_PLANET, 10},
	}
	for _, c := range cases {
		if got := PollutionTolerance(c.size); got != c.want {
			t.Errorf("PollutionTolerance(%v) = %d,預期 %d", c.size, got, c.want)
		}
	}
}

func TestPollutionToleranceWithNanoDisassemblers(t *testing.T) {
	if got := PollutionToleranceWithNanoDisassemblers(MEDIUM_PLANET); got != 12 { // 6*2
		t.Errorf("PollutionToleranceWithNanoDisassemblers(MEDIUM) = %d,預期 12", got)
	}
	if got := PollutionToleranceWithNanoDisassemblers(HUGE_PLANET); got != 20 { // 10*2
		t.Errorf("PollutionToleranceWithNanoDisassemblers(HUGE) = %d,預期 20", got)
	}
}

func TestPollutionCleanupCost(t *testing.T) {
	// 超出容忍值的一半用於清理污染(向下取整);tolerantRace 一律 0。
	cases := []struct {
		production, tolerance int
		tolerantRace          bool
		want                  int
	}{
		{10, 6, false, 2}, // 超出4,一半=2
		{9, 6, false, 1},  // 超出3,一半=1(向下取整)
		{6, 6, false, 0},  // 剛好等於容忍值→0
		{3, 6, false, 0},  // 未超出→0
		{100, 6, true, 0}, // Tolerant 種族不受影響
		{0, 0, false, 0},
	}
	for _, c := range cases {
		if got := PollutionCleanupCost(c.production, c.tolerance, c.tolerantRace); got != c.want {
			t.Errorf("PollutionCleanupCost(%d,%d,%v) = %d,預期 %d",
				c.production, c.tolerance, c.tolerantRace, got, c.want)
		}
	}
}

func TestPollutionEighths(t *testing.T) {
	cases := []struct {
		processor, renewer, coreWasteDump bool
		want                              int
	}{
		{false, false, false, 8}, // 無建築,全部產能致污染
		{true, false, false, 4},  // 只有污染處理器 → 1/2
		{false, true, false, 2},  // 只有大氣更新器 → 1/4
		{true, true, false, 1},   // 兩者皆有 → 1/8(手冊直接給的組合值)
		{false, false, true, 0},  // 核心廢料場 → 全消除
		{true, true, true, 0},    // 核心廢料場取代前兩者,依然 0
	}
	for _, c := range cases {
		if got := PollutionEighths(c.processor, c.renewer, c.coreWasteDump); got != c.want {
			t.Errorf("PollutionEighths(%v,%v,%v) = %d,預期 %d",
				c.processor, c.renewer, c.coreWasteDump, got, c.want)
		}
	}
}

func TestPollutionPollutingProduction(t *testing.T) {
	cases := []struct{ production, eighths, want int }{
		{16, 8, 16},
		{16, 4, 8},
		{16, 2, 4},
		{16, 1, 2},
		{16, 0, 0},
		{17, 4, 8}, // 17*4/8=8.5→8(向下取整)
	}
	for _, c := range cases {
		if got := PollutionPollutingProduction(c.production, c.eighths); got != c.want {
			t.Errorf("PollutionPollutingProduction(%d,%d) = %d,預期 %d",
				c.production, c.eighths, got, c.want)
		}
	}
}
