package gamedata

import "testing"

// TestCombatRangeLevel 對照手冊 Range Penalty 表的「Regular (sq)」列:
// Range 0/1/2/3/4/5/6/7/8 對應 Regular sq 0/1-3/4-6/7-9/10-12/13-15/16-18/19-21/22-24。
func TestCombatRangeLevel(t *testing.T) {
	cases := []struct{ sq, want int }{
		{0, 0},
		{1, 1}, {3, 1},
		{4, 2}, {6, 2},
		{7, 3}, {9, 3},
		{10, 4}, {12, 4},
		{13, 5}, {15, 5},
		{16, 6}, {18, 6},
		{19, 7}, {21, 7},
		{22, 8}, {24, 8},
		{30, 8}, // 超過表列範圍,夾限在 8
	}
	for _, c := range cases {
		if got := CombatRangeLevel(c.sq); got != c.want {
			t.Errorf("CombatRangeLevel(%d) = %d,預期 %d", c.sq, got, c.want)
		}
	}
}

// TestCombatRangeLevelPointDefense 對照手冊 PD (sq) 列(僅偶數 level 有值):
// 0→level0、1-3→level2、4-6→level4、7-9→level6、10-12→level8。
func TestCombatRangeLevelPointDefense(t *testing.T) {
	cases := []struct{ sq, want int }{
		{0, 0},
		{1, 2}, {3, 2},
		{4, 4}, {6, 4},
		{7, 6}, {9, 6},
		{10, 8}, {12, 8},
		{20, 8}, // 超過表列範圍,夾限在 8
	}
	for _, c := range cases {
		if got := CombatRangeLevelPointDefense(c.sq); got != c.want {
			t.Errorf("CombatRangeLevelPointDefense(%d) = %d,預期 %d", c.sq, got, c.want)
		}
	}
}

// TestCombatRangeLevelHeavy 對照手冊 Hv (sq) 列:
// 0-3→0、4-9→1、10-15→2、16-21→3、22-27→4、28-33→5、34-39→6、40-45→7、46-51→8。
func TestCombatRangeLevelHeavy(t *testing.T) {
	cases := []struct{ sq, want int }{
		{0, 0}, {3, 0},
		{4, 1}, {9, 1},
		{10, 2}, {15, 2},
		{16, 3}, {21, 3},
		{22, 4}, {27, 4},
		{28, 5}, {33, 5},
		{34, 6}, {39, 6},
		{40, 7}, {45, 7},
		{46, 8}, {51, 8},
	}
	for _, c := range cases {
		if got := CombatRangeLevelHeavy(c.sq); got != c.want {
			t.Errorf("CombatRangeLevelHeavy(%d) = %d,預期 %d", c.sq, got, c.want)
		}
	}
}

// TestCombatRangeLevelPenalty 對照手冊 Range Penalty 表:
// level 0/1/2/3/4/5/6/7/8 → penalty 0/0/10/20/30/40/55/70/85。
func TestCombatRangeLevelPenalty(t *testing.T) {
	cases := []struct{ level, want int }{
		{0, 0}, {1, 0}, {2, 10}, {3, 20}, {4, 30},
		{5, 40}, {6, 55}, {7, 70}, {8, 85},
		{-1, 0}, {9, 85}, // 越界夾限
	}
	for _, c := range cases {
		if got := CombatRangeLevelPenalty(c.level); got != c.want {
			t.Errorf("CombatRangeLevelPenalty(%d) = %d,預期 %d", c.level, got, c.want)
		}
	}
}

// TestCombatRangeLevelPenaltyDoubled:Classic Fusion Beam / Plasma Cannon /
// Mauler Device 的內建 2x range to-hit penalty(手冊原文:「This mod doubles
// the calculated penalty.」)。
func TestCombatRangeLevelPenaltyDoubled(t *testing.T) {
	if got := CombatRangeLevelPenaltyDoubled(4, false); got != 30 {
		t.Errorf("CombatRangeLevelPenaltyDoubled(4,false) = %d,預期 30", got)
	}
	if got := CombatRangeLevelPenaltyDoubled(4, true); got != 60 {
		t.Errorf("CombatRangeLevelPenaltyDoubled(4,true) = %d,預期 60", got)
	}
}

// TestCombatHitThreshold 對照手冊 Damage Potential 段落的兩個 Examples:
//  1. Death Ray at 23 sq(range level 8,penalty 85,pdBonus 0)
//     手冊:「B = (100 - (40 - -85)) = (100 - 95) = 5」→ hit_threshold = 95(封頂)。
//  2. 同一武器距離改為 11 sq(range level 4,penalty 30,pdBonus 0)
//     手冊:「B = (100 - (40 - -30)) = (100 - 70) = 30」→ hit_threshold = 70。
func TestCombatHitThreshold(t *testing.T) {
	lv8 := CombatRangeLevel(23) // = 8
	if got := CombatHitThreshold(CombatRangeLevelPenalty(lv8), 0); got != 95 {
		t.Errorf("CombatHitThreshold(23sq) = %d,預期 95(手冊 Example 1)", got)
	}
	lv4 := CombatRangeLevel(11) // = 4
	if got := CombatHitThreshold(CombatRangeLevelPenalty(lv4), 0); got != 70 {
		t.Errorf("CombatHitThreshold(11sq) = %d,預期 70(手冊 Example 2)", got)
	}
	// PD_bonus 會降低 threshold(手冊公式:40 + range_penalty - PD_bonus)。
	if got := CombatHitThreshold(85, 50); got != 75 {
		t.Errorf("CombatHitThreshold(85,50) = %d,預期 75", got)
	}
	// 未達 95 上限時不封頂。
	if got := CombatHitThreshold(10, 0); got != 50 {
		t.Errorf("CombatHitThreshold(10,0) = %d,預期 50", got)
	}
}

// TestCombatClassicToHit 對照手冊 Classic Chance to Hit Formula 的三個分支。
func TestCombatClassicToHit(t *testing.T) {
	// [1] random(100) > 95 一律命中,無視其他數值。
	if !CombatClassicToHit(96, -999, 999) {
		t.Errorf("CombatClassicToHit(96,-999,999) 應為 true(幸運一擊)")
	}
	// [2] BA+CO-AF-BD >= 99 一律命中。
	if !CombatClassicToHit(1, 99, 999) {
		t.Errorf("CombatClassicToHit(1,99,999) 應為 true(netAttack>=99)")
	}
	// [3a] roll+netAttack >= hit_threshold。
	if !CombatClassicToHit(50, 20, 70) {
		t.Errorf("CombatClassicToHit(50,20,70) 應為 true(50+20>=70)")
	}
	if CombatClassicToHit(49, 20, 70) {
		t.Errorf("CombatClassicToHit(49,20,70) 應為 false(49+20<70)")
	}
}

// TestCombatAlternativeToHit 對照手冊「1.50 Alternative Chance to Hit
// Formula (Optional)」的三個分支。
func TestCombatAlternativeToHit(t *testing.T) {
	// [1] random(100) > 95 一律命中。
	if !CombatAlternativeToHit(96, -999, 999, 0) {
		t.Errorf("CombatAlternativeToHit(96,...) 應為 true(幸運一擊)")
	}
	// [2] netAttack-range_penalty+PD_bonus >= 99 一律命中。
	if !CombatAlternativeToHit(1, 100, 1, 0) { // 100-1+0=99
		t.Errorf("CombatAlternativeToHit(1,100,1,0) 應為 true(adjusted>=99)")
	}
	// [3] adjusted + roll >= 40。
	if !CombatAlternativeToHit(10, 40, 20, 10) { // adjusted=40-20+10=30;30+10=40
		t.Errorf("CombatAlternativeToHit(10,40,20,10) 應為 true(30+10>=40)")
	}
	if CombatAlternativeToHit(9, 40, 20, 10) { // 30+9=39<40
		t.Errorf("CombatAlternativeToHit(9,40,20,10) 應為 false(30+9<40)")
	}
}

// 飛彈 Beam Defense 公式(CombatMissileSpeed 對應概念、MissileBonus 表、
// MissileBeamDefense)已在 missile_test.go 涵蓋(MissileSpeed /
// MissileWarheadBonus / MissileBeamDefense),此處不重複測試。

// TestCombatFighterSpeed 對照手冊公式:
// Speed = BaseSpeed of Fighter + 2*(FTLlevel-1) + TDBonus(4)。
func TestCombatFighterSpeed(t *testing.T) {
	// Interceptor(BaseSpeed 10)@ FTLlevel 1: 10 + 0 + 4 = 14。
	if got := CombatFighterSpeed(CombatFighterBaseSpeedInterceptor, 1); got != 14 {
		t.Errorf("CombatFighterSpeed(Interceptor,1) = %d,預期 14", got)
	}
	// Assault Shuttle(BaseSpeed 6)@ FTLlevel 3: 6 + 4 + 4 = 14。
	if got := CombatFighterSpeed(CombatFighterBaseSpeedAssaultShuttle, 3); got != 14 {
		t.Errorf("CombatFighterSpeed(AssaultShuttle,3) = %d,預期 14", got)
	}
	// Bomber / Heavy Fighter(BaseSpeed 8)@ FTLlevel 0: 8 - 2 + 4 = 10。
	if got := CombatFighterSpeed(CombatFighterBaseSpeedBomber, 0); got != 10 {
		t.Errorf("CombatFighterSpeed(Bomber,0) = %d,預期 10", got)
	}
	if got := CombatFighterSpeed(CombatFighterBaseSpeedHeavyFighter, 0); got != 10 {
		t.Errorf("CombatFighterSpeed(HeavyFighter,0) = %d,預期 10", got)
	}
}

// TestCombatFighterBeamDefense 對照手冊公式:
// 5 * Speed + RacialShipDefenseBonus + FighterPilotBonus + HelmsmanBonus。
func TestCombatFighterBeamDefense(t *testing.T) {
	speed := CombatFighterSpeed(CombatFighterBaseSpeedInterceptor, 1) // 14
	if got := CombatFighterBeamDefense(speed, 0, 0, 0); got != 70 {   // 5*14
		t.Errorf("CombatFighterBeamDefense(14,0,0,0) = %d,預期 70", got)
	}
	if got := CombatFighterBeamDefense(speed, 25, 10, 5); got != 110 { // 70+25+10+5
		t.Errorf("CombatFighterBeamDefense(14,25,10,5) = %d,預期 110", got)
	}
}
