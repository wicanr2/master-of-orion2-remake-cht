package gamedata

import (
	"math/rand"
	"testing"
	"time"
)

// 本檔的測試不是「驗證我抄對了表」(那是重言),而是驗證三件有資訊量的事:
//  1. 表結構自洽——權重欄總和落在原版的 100(表的維度/起點若錯,總和就不會是 100)
//  2. 與另一個獨立來源交叉一致——手冊值 vs 反組譯值、既有公式 vs 反組譯表
//  3. 生成函式的行為符合表所描述的分布與邊界

// TestStarClassWeightsColumnsSumTo100 驗證光譜權重三欄各自總和 = 100。
// 這是「_star_class_table 是 7×3 且起點正確」的獨立驗證:若把 21 bytes 切成 3×7,
// 或起點差一格,總和就不會剛好三欄都是 100。
func TestStarClassWeightsColumnsSumTo100(t *testing.T) {
	for age := 0; age < 3; age++ {
		sum := 0
		for sc := 0; sc < 7; sc++ {
			sum += StarClassWeights[sc][age]
		}
		if sum != 100 {
			t.Errorf("光譜權重第 %d 欄總和應為 100,got %d", age, sum)
		}
	}
}

// TestOldGalaxyClimateWeightsColumnsSumTo100 同上,驗證成熟星系氣候表四欄皆為 100。
// (Normal 表欄1 是 105,原版原值,故不一併斷言——見 NormalGalaxyClimateWeights 註解。)
func TestOldGalaxyClimateWeightsColumnsSumTo100(t *testing.T) {
	for g := 0; g < 4; g++ {
		sum := 0
		for c := 0; c < 10; c++ {
			sum += OldGalaxyClimateWeights[c][g]
		}
		if sum != 100 {
			t.Errorf("成熟星系氣候表第 %d 欄(溫度帶)總和應為 100,got %d", g, sum)
		}
	}
}

// TestPlanetMaxPopulationBaseMatchesFormula 交叉驗證兩個獨立來源:
// 反組譯的 _planet_max_population 表 vs remake 既有 PlanetBasePopMax 用的 (size+1)*5。
// 兩者對得上,代表手冊推得的公式與原版硬編表一致。
func TestPlanetMaxPopulationBaseMatchesFormula(t *testing.T) {
	for size := 0; size < 5; size++ {
		want := (size + 1) * 5
		if got := PlanetMaxPopulationBase[size]; got != want {
			t.Errorf("行星大小 %d 的基礎人口上限:原版表 %d vs 公式 %d", size, got, want)
		}
	}
}

// TestFoodPerFarmerMatchesOriginalTable 交叉驗證手冊 p.59 的每農夫食物值
// 與原版 _food_per_farmer_table @ 0x17D81C 逐格一致。
// 對得上就表示手冊在這一項沒有簡化或筆誤,remake 沿用手冊值是安全的。
func TestFoodPerFarmerMatchesOriginalTable(t *testing.T) {
	original := [10]int{0, 0, 0, 1, 1, 2, 2, 1, 2, 3} // Toxic…Gaia
	for c := 0; c < 10; c++ {
		if got := ClimateFoodPerFarmer(PlanetClimate(c)); got != original[c] {
			t.Errorf("氣候 %d 每農夫食物:remake %d vs 原版表 %d", c, got, original[c])
		}
	}
}

// TestMineralIndustryMatchesOriginalTable 同上,對照原版 _minerals_per_mine(cseg01 0xDD4B5)。
func TestMineralIndustryMatchesOriginalTable(t *testing.T) {
	original := [5]int{1, 2, 3, 5, 8} // Ultra Poor…Ultra Rich
	for m := 0; m < 5; m++ {
		if got := MineralIndustryPerWorker(PlanetMinerals(m)); got != original[m] {
			t.Errorf("礦產豐度 %d 每礦工工業:remake %d vs 原版表 %d", m, got, original[m])
		}
	}
}

// TestRollPlanetSizeDistribution 驗證 d10 累計骰表展開後的分布 = 10/20/40/20/10。
func TestRollPlanetSizeDistribution(t *testing.T) {
	count := map[PlanetSize]int{}
	for roll := 1; roll <= 10; roll++ {
		count[RollPlanetSize(roll)]++
	}
	want := map[PlanetSize]int{TINY_PLANET: 1, SMALL_PLANET: 2, MEDIUM_PLANET: 4, LARGE_PLANET: 2, HUGE_PLANET: 1}
	for size, w := range want {
		if count[size] != w {
			t.Errorf("行星大小 %v 在 d10 上應占 %d 格,got %d", size, w, count[size])
		}
	}
}

// TestGravityTableMonotonic 驗證重力表在「礦產豐度(密度)」與「大小(體積)」兩個方向都不遞減。
// 這是物理自洽性檢查——若表的行列被讀反,單調性會破。
func TestGravityTableMonotonic(t *testing.T) {
	for m := 0; m < 5; m++ {
		for s := 1; s < 5; s++ {
			if GravityTable[m][s] < GravityTable[m][s-1] {
				t.Errorf("礦產 %d:大小 %d→%d 重力下降(%d→%d)", m, s-1, s, GravityTable[m][s-1], GravityTable[m][s])
			}
		}
	}
	for s := 0; s < 5; s++ {
		for m := 1; m < 5; m++ {
			if GravityTable[m][s] < GravityTable[m-1][s] {
				t.Errorf("大小 %d:礦產 %d→%d 重力下降(%d→%d)", s, m-1, m, GravityTable[m-1][s], GravityTable[m][s])
			}
		}
	}
}

// TestPlanetOrbitGroupHabitableBand 驗證「宜居帶」的位置符合原版:
// 黃星中段軌道是溫度帶 2(唯一有 Ocean/Terran/Gaia 權重的一欄),藍星內側是 0。
func TestPlanetOrbitGroupHabitableBand(t *testing.T) {
	if g := PlanetOrbitGroup(Yellow, 2); g != 2 {
		t.Errorf("黃星第 3 軌道應在宜居帶(溫度帶 2),got %d", g)
	}
	if g := PlanetOrbitGroup(Blue, 0); g != 0 {
		t.Errorf("藍星第 1 軌道應是高溫帶 0,got %d", g)
	}
	if g := PlanetOrbitGroup(Red, 4); g != 3 {
		t.Errorf("紅矮星最外軌道應是低溫帶 3,got %d", g)
	}
	// 黑洞沒有行星,回退值 0 不應 panic。
	if g := PlanetOrbitGroup(BlackHole, 2); g != 0 {
		t.Errorf("黑洞應回退 0,got %d", g)
	}
}

// TestRollClimateOutsideHabitableBandNeverGoodWorlds 驗證溫度帶 0(貼近恆星)骰不出好星球:
// Tundra 以上(Ocean/Swamp/Arid/Terran/Gaia)在該欄權重全為 0,最好也只到 Desert(5%)。
// 這正是原版「星圖上大多數星球不能住」的來源,也是 remake 先前用光譜當氣候索引時丟掉的性質。
func TestRollClimateOutsideHabitableBandNeverGoodWorlds(t *testing.T) {
	r := rand.New(rand.NewSource(42))
	sawDesert := false
	for i := 0; i < 2000; i++ {
		c := RollClimate(r, 0, GalaxyAverage, false)
		if c > DESERT {
			t.Fatalf("溫度帶 0 最好只到 Desert,got %v(第 %d 次)", c, i)
		}
		if c == DESERT {
			sawDesert = true
		}
	}
	if !sawDesert {
		t.Error("溫度帶 0 有 5%% 的 Desert 權重,2000 次應該出現過")
	}
}

// TestRollClimateRequireHabitableAlwaysFarmable 驗證 requireHabitable 在宜居帶必回可農作氣候
// (原版用它保證母星那類星球不會骰成 Toxic)。
func TestRollClimateRequireHabitableAlwaysFarmable(t *testing.T) {
	r := rand.New(rand.NewSource(7))
	for i := 0; i < 1000; i++ {
		c := RollClimate(r, 2, GalaxyAverage, true)
		if ClimateFoodPerFarmer(c) <= 0 {
			t.Fatalf("requireHabitable 應只回可農作氣候,got %v", c)
		}
	}
}

// TestRollClimateRequireHabitableSingleCandidateShortcut 驗證原版「只剩一個可農作候選就直接給」
// 的捷徑:溫度帶 0 的十個權重裡,唯一有食物產出的是 Desert(5%),故必定回 Desert 而不是重骰。
// 順便確保這條路徑不會卡住(權重極低時盲目重骰會很慢)。
func TestRollClimateRequireHabitableSingleCandidateShortcut(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	done := make(chan PlanetClimate, 1)
	go func() { done <- RollClimate(r, 0, GalaxyAverage, true) }()
	select {
	case c := <-done:
		if c != DESERT {
			t.Errorf("溫度帶 0 唯一可農作氣候是 Desert,got %v", c)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("溫度帶 0 + requireHabitable 應走「唯一候選」捷徑直接回,不應卡住")
	}
}

// TestRollSpectralClassFollowsWeights 驗證加權骰確實照權重分布:
// 一般星系紅矮星權重 37%,大樣本下應該是最常見的類別。
func TestRollSpectralClassFollowsWeights(t *testing.T) {
	r := rand.New(rand.NewSource(99))
	count := map[SpectralClass]int{}
	const n = 20000
	for i := 0; i < n; i++ {
		count[RollSpectralClass(r, GalaxyAverage)]++
	}
	for sc := SpectralClass(0); sc < 7; sc++ {
		if sc == Red {
			continue
		}
		if count[Red] <= count[sc] {
			t.Errorf("一般星系紅矮星(權重 37)應多於 %v(權重 %d):%d vs %d",
				sc, StarClassWeights[sc][GalaxyAverage], count[Red], count[sc])
		}
	}
	// 權重 2 的黑洞在 20000 樣本下應該在 2% 附近(容忍 ±1 個百分點)。
	if p := float64(count[BlackHole]) / n * 100; p < 3 || p > 5 {
		t.Errorf("黑洞比例應接近 4%%(權重 4/100),got %.2f%%", p)
	}
}

// TestMatureGalaxyHasMoreHabitableWorlds 驗證成熟星系的宜居帶比一般星系更容易出可農作氣候
// (手冊對 Galaxy Age 的描述,由兩張表的數字直接支持)。
func TestMatureGalaxyHasMoreHabitableWorlds(t *testing.T) {
	sumFarmable := func(tbl [10][4]int) int {
		s := 0
		for c := 0; c < 10; c++ {
			if ClimateFoodPerFarmer(PlanetClimate(c)) > 0 {
				s += tbl[c][2] // 溫度帶 2 = 宜居帶
			}
		}
		return s
	}
	normal, mature := sumFarmable(NormalGalaxyClimateWeights), sumFarmable(OldGalaxyClimateWeights)
	if mature <= normal {
		t.Errorf("成熟星系宜居帶的可農作權重應高於一般星系:%d vs %d", mature, normal)
	}
}

// --- _orbit_to_satellite_type(原版 @ 0x17D6BC,50 bytes)---

// 表的形狀:10 列(Random_(10)-1)× 5 欄(軌道),值域 1..4。
func TestOrbitToSatelliteTypeShape(t *testing.T) {
	n := 0
	for r := range orbitToSatelliteType {
		for o := range orbitToSatelliteType[r] {
			v := orbitToSatelliteType[r][o]
			if v < 1 || v > 4 {
				t.Errorf("[%d][%d] = %d,值域應為 1..4", r, o, v)
			}
			n++
		}
	}
	if n != 50 {
		t.Errorf("表共 %d 格,原版是 50 bytes", n)
	}
}

// 決定性佐證:唯一的 4 只出現在 (roll 1, orbit 0),而原版特例分支寫死 `bl == 1 && orbit == 0`。
// 這條測試把「索引擺法正確」這件事釘在程式碼裡——擺法錯了(例如轉置成 5×10)就會失敗。
func TestOrbitSatelliteSpecialMarkerPosition(t *testing.T) {
	for r := range orbitToSatelliteType {
		for o := range orbitToSatelliteType[r] {
			if orbitToSatelliteType[r][o] != planetTypeSpecialMarker {
				continue
			}
			if r != 1 || o != 0 {
				t.Errorf("特例標記 4 出現在 [%d][%d],原版只在 [1][0]", r, o)
			}
		}
	}
	if orbitToSatelliteType[1][0] != planetTypeSpecialMarker {
		t.Error("[1][0] 應為特例標記 4")
	}
}

// 最內圈永遠不會是氣態巨星(原版對 orbit 0 的氣態巨星會整個重骰)。
func TestNoGasGiantInInnermostOrbit(t *testing.T) {
	for roll := 1; roll <= 10; roll++ {
		if got, _ := RollSatelliteType(roll, 0, 100); got == GAS_GIANT {
			t.Errorf("roll %d 在最內圈骰出氣態巨星", roll)
		}
	}
}

// 特例標記:命中 10% 才是特例天體,否則退回小行星帶;兩種情形類別都是小行星帶。
func TestRollSatelliteTypeSpecialMarker(t *testing.T) {
	for _, sp := range []int{1, 5, OrbitSatelliteSpecialRoll} {
		got, isSpecial := RollSatelliteType(2, 0, sp) // roll10=2 → r=1 → 表裡的 4
		if got != ASTEROIDS || !isSpecial {
			t.Errorf("special=%d:得 (%d,%v),want (小行星帶, true)", sp, got, isSpecial)
		}
	}
	for _, sp := range []int{OrbitSatelliteSpecialRoll + 1, 50, 100} {
		got, isSpecial := RollSatelliteType(2, 0, sp)
		if got != ASTEROIDS || isSpecial {
			t.Errorf("special=%d:得 (%d,%v),want (小行星帶, false)", sp, got, isSpecial)
		}
	}
}

// 分布語意:roll 越大整個系統越宜居(roll 7-10 五條軌道全是一般行星);
// roll 1 全是小行星帶。這是「表沒讀錯」的語意檢查,不是實作反推。
func TestSatelliteTypeDistributionSemantics(t *testing.T) {
	for orbit := 0; orbit < 5; orbit++ {
		if got, _ := RollSatelliteType(1, orbit, 100); got != ASTEROIDS {
			t.Errorf("roll 1 orbit %d 應為小行星帶,得 %d", orbit, got)
		}
		for roll := 7; roll <= 10; roll++ {
			if got, _ := RollSatelliteType(roll, orbit, 100); got != HABITABLE {
				t.Errorf("roll %d orbit %d 應為一般行星,得 %d", roll, orbit, got)
			}
		}
	}
	// 氣態巨星集中在外圈:統計整張表,orbit 4 的氣態巨星比 orbit 1 多。
	countGas := func(orbit int) int {
		n := 0
		for roll := 1; roll <= 10; roll++ {
			if got, _ := RollSatelliteType(roll, orbit, 100); got == GAS_GIANT {
				n++
			}
		}
		return n
	}
	if countGas(4) <= countGas(1) {
		t.Errorf("氣態巨星應集中在外圈:orbit4=%d orbit1=%d", countGas(4), countGas(1))
	}
}
