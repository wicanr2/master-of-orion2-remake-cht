package gamedata

import "testing"

// TestSatelliteBeamSpaceWithArc 驗證 arc-cost 佔格公式(比照 WeaponSpaceWithMods:
// base + base*pct/100,整數運算)在手冊錨點上的算法。
func TestSatelliteBeamSpaceWithArc(t *testing.T) {
	cases := []struct {
		name           string
		baseSpace, pct int
		want           int
	}{
		{"雷射 base10 + 1.3 衛星 arc25%", 10, 25, 12},   // 10 + 10*25/100 = 10+2
		{"雷射 base10 + 1.5 衛星 arc33%", 10, 33, 13},   // 10 + 10*33/100 = 10+3(330/100=3)
		{"雷射 base10 + 1.5 地面砲台 arc50%", 10, 50, 15}, // 10 + 10*50/100 = 10+5
		{"雷射 base10 + 1.3 地面砲台 arc0%", 10, 0, 10},   // 無懲罰,佔格不變
		{"電漿砲 base25 + 1.5 衛星 arc33%", 25, 33, 33},  // 25 + 25*33/100 = 25+8(825/100=8)
	}
	for _, c := range cases {
		if got := SatelliteBeamSpaceWithArc(c.baseSpace, c.pct); got != c.want {
			t.Errorf("%s: SatelliteBeamSpaceWithArc(%d, %d) = %d,want %d", c.name, c.baseSpace, c.pct, got, c.want)
		}
	}
}

// TestSatelliteBeamSpaceWithArc_MinimumOne 驗證邊界:即使 baseSpace<=0,結果最少夾在 1
// (比照 WeaponSpaceWithMods 的既有慣例,避免 0/負值佔格造成除零或無限塞武器)。
func TestSatelliteBeamSpaceWithArc_MinimumOne(t *testing.T) {
	if got := SatelliteBeamSpaceWithArc(0, 25); got != 1 {
		t.Errorf("SatelliteBeamSpaceWithArc(0, 25) = %d,want 1(下限保護)", got)
	}
}

// TestSatelliteWeaponFitCount 驗證 fit 計算對手冊錨點(飛彈基地 300 space、地面砲台
// 450 space)的算法,以及 spaceBudget/perWeaponSpace 非正值的邊界處理。
func TestSatelliteWeaponFitCount(t *testing.T) {
	cases := []struct {
		name                        string
		spaceBudget, perWeaponSpace int
		want                        int
	}{
		{"飛彈基地 300 space / 核飛彈 10 space(不吃 arc)", MissileBaseSpace, 10, 30},
		{"地面砲台 450 space / 1.3 雷射(arc0%,perBeam10)", GroundBatterySpace, 10, 45},
		{"地面砲台 450 space / 1.5 雷射(arc50%,perBeam15)", GroundBatterySpace, 15, 30},
		{"星基 250 space / 1.3 雷射(arc25%,perBeam12)", StarBaseSpace, 12, 20},
		{"戰鬥站 500 space / 1.3 雷射(arc25%,perBeam12)", BattlestationSpace, 12, 41},
		{"星辰要塞 1200 space / 1.3 雷射(arc25%,perBeam12)", StarFortressSpace, 12, 100},
		{"perWeaponSpace 非正值回 0", 250, 0, 0},
		{"spaceBudget 非正值回 0", 0, 12, 0},
	}
	for _, c := range cases {
		if got := SatelliteWeaponFitCount(c.spaceBudget, c.perWeaponSpace); got != c.want {
			t.Errorf("%s: SatelliteWeaponFitCount(%d, %d) = %d,want %d", c.name, c.spaceBudget, c.perWeaponSpace, got, c.want)
		}
	}
}

// TestSatelliteStrengthScaleCalibration 驗證校準除數 20 在雷射參考點下,重現既有
// 星基/戰鬥站 tier(4/8),星辰要塞算出 20(非近似 19,見 SatelliteStrengthScale 註解的
// 誠實落差說明——本測試如實斷言計算結果,不假造成 19)。
func TestSatelliteStrengthScaleCalibration(t *testing.T) {
	const laserValue = 4  // WeaponOptions「雷射」Value
	const laserSpace = 10 // WeaponSpaceByName["雷射"]

	compute := func(hullSpace, arcPct int) int {
		perBeam := SatelliteBeamSpaceWithArc(laserSpace, arcPct)
		fit := SatelliteWeaponFitCount(hullSpace, perBeam)
		return fit * laserValue / SatelliteStrengthScale
	}

	if got := compute(StarBaseSpace, 25); got != 4 {
		t.Errorf("星基(1.3 arc25%%)atk = %d,want 4(重現既有 tier)", got)
	}
	if got := compute(BattlestationSpace, 25); got != 8 {
		t.Errorf("戰鬥站(1.3 arc25%%)atk = %d,want 8(重現既有 tier)", got)
	}
	if got := compute(StarFortressSpace, 25); got != 20 {
		t.Errorf("星辰要塞(1.3 arc25%%)atk = %d,want 20(非近似 19,見 SatelliteStrengthScale 註解)", got)
	}

	// 版本差異:1.5 arc-cost 較高(33% > 25%)→ perBeam 較大 → fit 較少 → atk 應較低。
	if got13, got15 := compute(StarBaseSpace, 25), compute(StarBaseSpace, 33); got15 >= got13 {
		t.Errorf("星基:1.5 atk(%d)應 < 1.3 atk(%d)(arc-cost 較高、防禦略弱)", got15, got13)
	}
	if got13, got15 := compute(BattlestationSpace, 25), compute(BattlestationSpace, 33); got15 >= got13 {
		t.Errorf("戰鬥站:1.5 atk(%d)應 < 1.3 atk(%d)", got15, got13)
	}
	if got13, got15 := compute(StarFortressSpace, 25), compute(StarFortressSpace, 33); got15 >= got13 {
		t.Errorf("星辰要塞:1.5 atk(%d)應 < 1.3 atk(%d)", got15, got13)
	}
}
