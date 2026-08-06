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
