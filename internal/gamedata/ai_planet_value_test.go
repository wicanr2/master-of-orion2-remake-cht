package gamedata

import "testing"

func baseInput() AIPlanetValueInput {
	return AIPlanetValueInput{
		Habitable: true, MaxPop: 12, Minerals: ABUNDANT,
		Climate: TERRAN, Gravity: NORMAL_G, FoodBase: 2,
	}
}

// TestAIPlanetValueRanksGoodWorldsHigher 驗證估值把好行星排在前面。
// 這是這個函式唯一真正要做對的事——分數的絕對值不具意義,排序才是。
func TestAIPlanetValueRanksGoodWorldsHigher(t *testing.T) {
	terran := baseInput()
	radiated := baseInput()
	radiated.Climate = RADIATED
	radiated.FoodBase = 0
	if AIPlanetValue(terran, AIObjectiveBalancedLow) <= AIPlanetValue(radiated, AIObjectiveBalancedLow) {
		t.Error("類地星的估值應高於放射星")
	}

	rich, poor := baseInput(), baseInput()
	rich.Minerals, poor.Minerals = ULTRA_RICH, ULTRA_POOR
	if AIPlanetValue(rich, AIObjectiveBalancedLow) <= AIPlanetValue(poor, AIObjectiveBalancedLow) {
		t.Error("富礦星的估值應高於貧礦星")
	}

	big, small := baseInput(), baseInput()
	big.MaxPop, small.MaxPop = 20, 5
	if AIPlanetValue(big, AIObjectiveBalancedLow) <= AIPlanetValue(small, AIObjectiveBalancedLow) {
		t.Error("人口上限高的行星估值應較高")
	}
}

// TestAIPlanetValueUninhabitableIsZero 驗證不可住的行星一律 0 分。
func TestAIPlanetValueUninhabitableIsZero(t *testing.T) {
	gas := baseInput()
	gas.Habitable = false // 氣態巨星/小行星帶(原版 planet.type != HABITABLE)
	if got := AIPlanetValue(gas, AIObjectiveBalancedLow); got != 0 {
		t.Errorf("非宜居行星應為 0 分,got %d", got)
	}
	nopop := baseInput()
	nopop.MaxPop = 0
	if got := AIPlanetValue(nopop, AIObjectiveBalancedLow); got != 0 {
		t.Errorf("人口上限 0 的行星應為 0 分,got %d", got)
	}
}

// TestAIPlanetValueGravityPenalty 驗證重力天賦與行星重力的四種搭配(原版四個分支)。
func TestAIPlanetValueGravityPenalty(t *testing.T) {
	normalRace := baseInput()
	heavy := baseInput()
	heavy.Gravity = HEAVY_G
	if AIPlanetValue(heavy, AIObjectiveBalancedLow) >= AIPlanetValue(normalRace, AIObjectiveBalancedLow) {
		t.Error("一般種族看高重力行星應打折")
	}
	heavyRace := heavy
	heavyRace.RaceHeavyG = true
	if AIPlanetValue(heavyRace, AIObjectiveBalancedLow) <= AIPlanetValue(heavy, AIObjectiveBalancedLow) {
		t.Error("高重力種族看高重力行星不該打折")
	}
	// Low-G 種族在常重力行星反而要打折(原版分支:gravity==NORMAL 且種族 Low-G)。
	lowGRaceOnNormal := baseInput()
	lowGRaceOnNormal.RaceLowG = true
	if AIPlanetValue(lowGRaceOnNormal, AIObjectiveBalancedLow) >= AIPlanetValue(normalRace, AIObjectiveBalancedLow) {
		t.Error("低重力種族看常重力行星應打折")
	}
}

// TestAIPlanetValueObjectiveShiftsPreference 驗證目標傾向會改變偏好方向:
// 重礦產的 AI 相對更看重富礦星,重人口的 AI 相對更看重高人口星。
func TestAIPlanetValueObjectiveShiftsPreference(t *testing.T) {
	richBarren := AIPlanetValueInput{Habitable: true, MaxPop: 12,
		Minerals: ULTRA_RICH, Climate: BARREN, Gravity: NORMAL_G}
	poorTerran := AIPlanetValueInput{Habitable: true, MaxPop: 20,
		Minerals: ULTRA_POOR, Climate: TERRAN, Gravity: NORMAL_G, FoodBase: 2}

	mineralRatio := float64(AIPlanetValue(richBarren, AIObjectiveMineral)) /
		float64(AIPlanetValue(poorTerran, AIObjectiveMineral))
	popRatio := float64(AIPlanetValue(richBarren, AIObjectivePopulation)) /
		float64(AIPlanetValue(poorTerran, AIObjectivePopulation))
	if mineralRatio <= popRatio {
		t.Errorf("重礦產 AI 對「富礦 vs 貧礦類地」的相對偏好應強於重人口 AI:%.3f vs %.3f",
			mineralRatio, popRatio)
	}
}

// TestClimateMaintenanceModifiersMatchOriginal 驗證氣候維護表與原版 dump 逐格相同。
// 值反直覺(Barren 0、Desert 25),兩次獨立 dump 一致,照抄不臆改——這個測試就是防止
// 未來有人「覺得不合理」而順手改掉。
func TestClimateMaintenanceModifiersMatchOriginal(t *testing.T) {
	want := [10]int{50, 25, 0, 25, 0, 0, 0, 0, 0, 0}
	for c := TOXIC; c <= GAIA; c++ {
		if got := ClimateMaintenanceModifier(c); got != want[c] {
			t.Errorf("氣候 %d 的維護修正應為 %d(原版 _climate_maintenance_modifiers),got %d",
				c, want[c], got)
		}
	}
}

// TestAIPlanetValueClampedTo16Bit 驗證輸出夾在 16-bit(原版存的是 int16 陣列)。
func TestAIPlanetValueClampedTo16Bit(t *testing.T) {
	huge := AIPlanetValueInput{Habitable: true, MaxPop: 100000,
		Minerals: ULTRA_RICH, Climate: GAIA, Gravity: NORMAL_G, FoodBase: 3, Special: 5}
	if got := AIPlanetValue(huge, AIObjectivePopulation); got != 65535 {
		t.Errorf("極端輸入應夾在 65535,got %d", got)
	}
}

// TestAIProximityValueFavoursNearby 驗證鄰近價值隨距離遞減(原版 120/distance)。
func TestAIProximityValueFavoursNearby(t *testing.T) {
	near := AIProximityValue([]int{2})
	far := AIProximityValue([]int{20})
	if near <= far {
		t.Errorf("近的星鄰近價值應較高:距離2=%d 距離20=%d", near, far)
	}
	// 多顆我方星應該累加。
	if two := AIProximityValue([]int{4, 4}); two <= AIProximityValue([]int{4}) {
		t.Error("多顆我方鄰星的鄰近價值應累加")
	}
	// 距離 0 不應除以零。
	if AIProximityValue([]int{0}) != AIProximityOwnWeight {
		t.Error("距離 0 應視為 1,回傳完整權重")
	}
}

// TestAIContextualEnemyPenalty 驗證敵方鄰居會壓低估值(原版 /(n+2))。
// 這是「避開敵方勢力範圍」的來源。
func TestAIContextualEnemyPenalty(t *testing.T) {
	clean := AIContextualInput{Base: 1000, Size: MEDIUM_PLANET}
	oneEnemy := clean
	oneEnemy.NeighborEnemyN = 1
	twoEnemies := clean
	twoEnemies.NeighborEnemyN = 2

	c, o, w := AIContextualPlanetValue(clean), AIContextualPlanetValue(oneEnemy), AIContextualPlanetValue(twoEnemies)
	if !(c > o && o > w) {
		t.Errorf("敵方鄰居越多估值應越低:無敵 %d > 一敵 %d > 兩敵 %d", c, o, w)
	}
	if o != 1000/3 {
		t.Errorf("一個敵方鄰居應除以 (1+2):want %d got %d", 1000/3, o)
	}
}

// TestAIContextualEmptyNeighborBonus 驗證「鄰近有無主星」會加分(整區開發的潛力)。
func TestAIContextualEmptyNeighborBonus(t *testing.T) {
	alone := AIContextualInput{Base: 500, Size: MEDIUM_PLANET}
	withEmpty := alone
	withEmpty.NeighborEmpty = 800
	if AIContextualPlanetValue(withEmpty) <= AIContextualPlanetValue(alone) {
		t.Error("鄰近有可殖民的無主星應加分")
	}
	if got, want := AIContextualPlanetValue(withEmpty), 500+800/8; got != want {
		t.Errorf("無主鄰居加分應為總和/8:want %d got %d", want, got)
	}
}

// TestAIContextualOwnNeighborScalesBySize 驗證「鄰近已有我方殖民地」時改成依行星大小縮放
// (原版:同系第二顆殖民地的邊際價值較低,大星才值得)。
func TestAIContextualOwnNeighborScalesBySize(t *testing.T) {
	base := AIContextualInput{Base: 1000, NeighborOwnN: 1, NeighborOwn: 500}
	small, huge := base, base
	small.Size, huge.Size = SMALL_PLANET, HUGE_PLANET
	if AIContextualPlanetValue(huge) <= AIContextualPlanetValue(small) {
		t.Error("鄰近已有我方殖民地時,大星的邊際價值應高於小星")
	}
	// 已有殖民地的星走另一條路徑:加鄰居價值而非縮放。
	colonized := base
	colonized.Colonized = true
	colonized.Size = SMALL_PLANET
	if AIContextualPlanetValue(colonized) <= AIContextualPlanetValue(small) {
		t.Error("已殖民的星應走「加鄰居價值」路徑,不套用大小縮放折扣")
	}
}

// --- Enemy_Colony_Worth_To_Player_ @ 0xD8D11 ---

// 權重恆和 6(原版最後 idiv 6),且各檔的分派與組語逐條相同。
func TestAIEnemyTargetWeightsSumToSix(t *testing.T) {
	cases := []struct {
		policy      AIForeignPolicy
		owner, self int
	}{
		{DiploNone, 5, 1},
		{DiploNonAggression, 5, 1},
		{DiploAlliance, 5, 1},
		{DiploPeace, 5, 1},
		{DiploLimitedWar, 4, 2},
		{DiploWar, 5, 1}, // ⚠ 原版對 5 沒有專屬分支,落在 default——不是漏寫
		{DiploTotalWar, 6, 0},
		{AIForeignPolicy(7), 5, 1},
	}
	for _, c := range cases {
		o, s := aiEnemyTargetWeights(c.policy)
		if o != c.owner || s != c.self {
			t.Errorf("policy %d 權重 = (%d,%d),want (%d,%d)", c.policy, o, s, c.owner, c.self)
		}
		if o+s != AIEnemyTargetWeightSum {
			t.Errorf("policy %d 權重和 = %d,want %d", c.policy, o+s, AIEnemyTargetWeightSum)
		}
	}
}

// 核心語意:目標估值偏向「主人的估值」——打他最痛的地方,不是搶我最想要的地方。
func TestAIEnemyColonyValueFavoursVictimValuation(t *testing.T) {
	const ownerVal, selfVal = 600, 60
	peace := AIEnemyColonyValue(ownerVal, selfVal, DiploPeace, false)
	limited := AIEnemyColonyValue(ownerVal, selfVal, DiploLimitedWar, false)
	total := AIEnemyColonyValue(ownerVal, selfVal, DiploTotalWar, false)

	// 全面戰爭那一檔完全不看自己想不想要(6:0)。
	if total != ownerVal {
		t.Errorf("全面戰爭應完全採用主人估值:%d,want %d", total, ownerVal)
	}
	// 有限戰爭給自己的估值多一點權重 → 在「主人估值 > 自己估值」時總分較低。
	if limited >= peace {
		t.Errorf("有限戰爭(4:2)應低於和平(5:1):%d vs %d", limited, peace)
	}
	// 三檔都應落在兩個估值之間。
	for _, v := range []int{peace, limited, total} {
		if v < selfVal || v > ownerVal {
			t.Errorf("估值 %d 落在 [%d, %d] 之外", v, selfVal, ownerVal)
		}
	}
}

// shiftToSelf(原版玩家結構偏移 0x8B8 的種族旗標)把權重往自己的估值挪一格。
func TestAIEnemyColonyValueShiftToSelf(t *testing.T) {
	const ownerVal, selfVal = 600, 60
	base := AIEnemyColonyValue(ownerVal, selfVal, DiploPeace, false)
	shifted := AIEnemyColonyValue(ownerVal, selfVal, DiploPeace, true)
	if shifted >= base {
		t.Errorf("往自己的估值挪一格後應變低(自己估值較小):%d vs %d", shifted, base)
	}
	// 挪到底(6:0 → 5:1)不會讓權重變負。
	if got := AIEnemyColonyValue(ownerVal, selfVal, DiploTotalWar, true); got >= ownerVal {
		t.Errorf("全面戰爭挪一格後應低於純主人估值:%d", got)
	}
}
