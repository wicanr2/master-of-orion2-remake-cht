package gamedata

import "testing"

// TestDamageDissipationPenalty 對照手冊 Damage Potential 表的「Penalty」列(colspan 已核對)。
func TestDamageDissipationPenalty(t *testing.T) {
	cases := []struct{ level, want int }{
		{0, 0}, {1, 0}, {2, 10}, {3, 20}, {4, 30}, {5, 40}, {6, 50}, {7, 60}, {8, 65},
		{9, 65}, // 超出表列範圍,夾限在 level 8
		{-1, 0}, // 夾限在 level 0
	}
	for _, c := range cases {
		if got := DamageDissipationPenalty(c.level); got != c.want {
			t.Errorf("DamageDissipationPenalty(%d) = %d,預期 %d", c.level, got, c.want)
		}
	}
}

// TestDamageMountAdjustedValue 逐字複現手冊「Hv, PD, HEF & Ordnance」小節「a 50-100 beam with
// dissipation Range_penalty of 30」的 5 組共 10 個數字。
func TestDamageMountAdjustedValue(t *testing.T) {
	const rp = 30
	cases := []struct {
		name                        string
		base, hv, hef, pd, rangePts int
		want                        int
	}{
		{"Hv min", 50, DamageMountBonusHeavy, 0, 0, rp, 60},
		{"Hv max", 100, DamageMountBonusHeavy, 0, 0, rp, 120},
		{"Hv+HEF min", 50, DamageMountBonusHeavy, DamageMountBonusHEF, 0, rp, 85},
		{"Hv+HEF max", 100, DamageMountBonusHeavy, DamageMountBonusHEF, 0, rp, 170},
		{"PD min", 50, 0, 0, DamageMountPenaltyPointDefense, rp, 10},
		{"PD max", 100, 0, 0, DamageMountPenaltyPointDefense, rp, 20},
		{"PD 2x min(夾為1)", 50, 0, 0, DamageMountPenaltyPointDefense, rp * 2, 1},
		{"PD 2x max(夾為1)", 100, 0, 0, DamageMountPenaltyPointDefense, rp * 2, 1},
		{"PD+HEF 2x min", 50, 0, DamageMountBonusHEF, DamageMountPenaltyPointDefense, rp * 2, 20},
		{"PD+HEF 2x max", 100, 0, DamageMountBonusHEF, DamageMountPenaltyPointDefense, rp * 2, 40},
	}
	for _, c := range cases {
		if got := DamageMountAdjustedValue(c.base, c.hv, c.hef, c.pd, c.rangePts); got != c.want {
			t.Errorf("%s: DamageMountAdjustedValue(%d,%d,%d,%d,%d) = %d,預期 %d",
				c.name, c.base, c.hv, c.hef, c.pd, c.rangePts, got, c.want)
		}
	}
}

// TestDamageApplyDissipation 對照手冊「Reduced by Range」範例表(Phasor / Mauler / Death Ray
// 三個 base,逐 level 完整核對)。Laser Cannon 範例的 level7 儲存格(手冊印「1-1」)與本公式
// 算出的 1-2 對不上,詳見 damage.go 的 DamageApplyDissipation 註解,不在此強行斷言。
func TestDamageApplyDissipation(t *testing.T) {
	type pair struct{ min, max int }
	// Phasor base 5-20,手冊逐 level(0..8)給出:5-20,5-20,5-18,4-16,4-14,3-12,3-10,2-8,2-7
	phasorWant := []pair{{5, 20}, {5, 20}, {5, 18}, {4, 16}, {4, 14}, {3, 12}, {3, 10}, {2, 8}, {2, 7}}
	for level, want := range phasorWant {
		gotMin, gotMax := DamageApplyDissipation(5, 20, level)
		if gotMin != want.min || gotMax != want.max {
			t.Errorf("Phasor level%d: DamageApplyDissipation(5,20,%d) = %d-%d,預期 %d-%d",
				level, level, gotMin, gotMax, want.min, want.max)
		}
	}
	// Mauler base 100-100(min=max),手冊逐 level 給出:100,100,90,80,70,60,50,40,35
	maulerWant := []int{100, 100, 90, 80, 70, 60, 50, 40, 35}
	for level, want := range maulerWant {
		gotMin, gotMax := DamageApplyDissipation(100, 100, level)
		if gotMin != want || gotMax != want {
			t.Errorf("Mauler level%d: DamageApplyDissipation(100,100,%d) = %d-%d,預期 %d-%d",
				level, level, gotMin, gotMax, want, want)
		}
	}
	// Death Ray base 50-100,手冊逐 level 給出:50-100,50-100,45-90,40-80,35-70,30-60,25-50,20-40,18-35
	deathRayWant := []pair{{50, 100}, {50, 100}, {45, 90}, {40, 80}, {35, 70}, {30, 60}, {25, 50}, {20, 40}, {18, 35}}
	for level, want := range deathRayWant {
		gotMin, gotMax := DamageApplyDissipation(50, 100, level)
		if gotMin != want.min || gotMax != want.max {
			t.Errorf("Death Ray level%d: DamageApplyDissipation(50,100,%d) = %d-%d,預期 %d-%d",
				level, level, gotMin, gotMax, want.min, want.max)
		}
	}
}

// TestDamageApplyDissipationLaserKnownMismatch 記錄手冊 Laser Cannon 範例表 level7 那格與
// 公式算出值不同,做為已知落差的證據(不是本檔的斷言目標),避免日後誤以為沒注意到這件事。
func TestDamageApplyDissipationLaserKnownMismatch(t *testing.T) {
	_, gotMax := DamageApplyDissipation(1, 4, 7) // 手冊印 1-1(max=1),公式算出 1-2(max=2)
	if gotMax != 2 {
		t.Fatalf("預期本公式在 Laser level7 算出 max=2(與手冊該格的 1-1 不同,見 damage.go 註解),實際卻是 %d,表示公式本身變了,需要重新比對其餘 30+ 個已驗證數字", gotMax)
	}
}

// TestDamageForHit 逐字複現手冊「Different Min-Max Damage」小節兩個 worked example
// (Death Ray,roll=85、netAttack=10 為共同輸入)。
func TestDamageForHit(t *testing.T) {
	// Example 1: range 23 sq → level8,accuracy range_penalty=85(combat.go 的 to-hit 表,
	// 與本檔 dissipation 表不同),hitThreshold = min(40+85-0,95) = 95。
	rangePenalty1 := CombatRangeLevelPenalty(CombatRangeLevel(23))
	if rangePenalty1 != 85 {
		t.Fatalf("前提有誤:CombatRangeLevelPenalty(CombatRangeLevel(23)) = %d,預期 85", rangePenalty1)
	}
	hitThreshold1 := CombatHitThreshold(rangePenalty1, 0)
	if hitThreshold1 != 95 {
		t.Fatalf("前提有誤:CombatHitThreshold(85,0) = %d,預期 95", hitThreshold1)
	}
	minDmg1, maxDmg1 := DamageApplyDissipation(50, 100, CombatRangeLevel(23)) // Death Ray 在 level8
	if got := DamageForHit(minDmg1, maxDmg1, 85, 10, hitThreshold1); got != 18 {
		t.Errorf("Example1(23 sq): DamageForHit(%d,%d,85,10,%d) = %d,預期 18(=min_dmg)",
			minDmg1, maxDmg1, hitThreshold1, got)
	}

	// Example 2: range 11 sq → level4,accuracy range_penalty=30,hitThreshold=min(40+30-0,95)=70。
	rangePenalty2 := CombatRangeLevelPenalty(CombatRangeLevel(11))
	if rangePenalty2 != 30 {
		t.Fatalf("前提有誤:CombatRangeLevelPenalty(CombatRangeLevel(11)) = %d,預期 30", rangePenalty2)
	}
	hitThreshold2 := CombatHitThreshold(rangePenalty2, 0)
	if hitThreshold2 != 70 {
		t.Fatalf("前提有誤:CombatHitThreshold(30,0) = %d,預期 70", hitThreshold2)
	}
	minDmg2, maxDmg2 := DamageApplyDissipation(50, 100, CombatRangeLevel(11)) // Death Ray 在 level4
	if got := DamageForHit(minDmg2, maxDmg2, 85, 10, hitThreshold2); got != 65 {
		t.Errorf("Example2(11 sq): DamageForHit(%d,%d,85,10,%d) = %d,預期 65(=min_dmg+⅚*(max-min+1))",
			minDmg2, maxDmg2, hitThreshold2, got)
	}
}

// TestDamageForHitMaxDamageBranches 對照手冊 [1][2] 兩個「必為 max dmg」分支。
func TestDamageForHitMaxDamageBranches(t *testing.T) {
	if got := DamageForHit(10, 40, 96, 0, 50); got != 40 {
		t.Errorf("random(100)>95 應直接給 max_dmg,got %d", got)
	}
	if got := DamageForHit(10, 40, 1, 99, 50); got != 40 {
		t.Errorf("netAttack>=99 應直接給 max_dmg,got %d", got)
	}
}

// TestDamageShieldCapacity 對照手冊 Class I/III/V/VII/X Shield 條目的「absorbing up to N times
// the ship's size」倍率(N = 5*等級數字)。
func TestDamageShieldCapacity(t *testing.T) {
	cases := []struct {
		name       string
		reduction  int
		shipSize   int
		wantFactor int // 「N times ship's size」的 N
	}{
		{"Class I", DamageShieldReductionClassI, 4, 5},
		{"Class III", DamageShieldReductionClassIII, 4, 15},
		{"Class V", DamageShieldReductionClassV, 4, 25},
		{"Class VII", DamageShieldReductionClassVII, 4, 35},
		{"Class X", DamageShieldReductionClassX, 4, 50},
	}
	for _, c := range cases {
		want := c.wantFactor * c.shipSize
		if got := DamageShieldCapacity(c.reduction, c.shipSize); got != want {
			t.Errorf("%s: DamageShieldCapacity(%d,%d) = %d,預期 %d(=%d*size)",
				c.name, c.reduction, c.shipSize, got, want, c.wantFactor)
		}
	}
}

// TestDamageAfterShield 對照手冊各護盾等級「each attack reduced by N points of damage」,以及
// Hard Shields 的額外 -3 與「immunity to shield-piercing weapons」規則。
func TestDamageAfterShield(t *testing.T) {
	cases := []struct {
		name                       string
		dmg, shieldReduction       int
		hardShield, shieldPiercing bool
		want                       int
	}{
		{"Class VII 護盾", 20, DamageShieldReductionClassVII, false, false, 13},
		{"護盾打不穿,夾在0", 5, DamageShieldReductionClassX, false, false, 0},
		{"Shield Piercing 無視護盾", 20, DamageShieldReductionClassX, false, true, 20},
		{"Hard Shields 使 SP 失效", 20, DamageShieldReductionClassX, true, true, 20 - DamageShieldReductionClassX - DamageHardShieldBonus},
		{"Hard Shields 疊加(無 SP)", 20, DamageShieldReductionClassI, true, false, 20 - DamageShieldReductionClassI - DamageHardShieldBonus},
	}
	for _, c := range cases {
		if got := DamageAfterShield(c.dmg, c.shieldReduction, c.hardShield, c.shieldPiercing); got != c.want {
			t.Errorf("%s: DamageAfterShield(%d,%d,%v,%v) = %d,預期 %d",
				c.name, c.dmg, c.shieldReduction, c.hardShield, c.shieldPiercing, got, c.want)
		}
	}
}

// TestDamageApplyArmor 對照手冊「AP: ... All of the damage done passes through as if there
// were no armor at all.」以及 Heavy Armor / Xentronium Armor 使 AP 失效的規則。
func TestDamageApplyArmor(t *testing.T) {
	// 一般武器:先扣裝甲,溢出才傷結構。
	toArmor, toStructure, remain := DamageApplyArmor(30, 20, false, false)
	if toArmor != 20 || toStructure != 10 || remain != 0 {
		t.Errorf("一般武器溢出裝甲: got (%d,%d,%d),預期 (20,10,0)", toArmor, toStructure, remain)
	}
	toArmor, toStructure, remain = DamageApplyArmor(10, 20, false, false)
	if toArmor != 10 || toStructure != 0 || remain != 10 {
		t.Errorf("一般武器未打穿裝甲: got (%d,%d,%d),預期 (10,0,10)", toArmor, toStructure, remain)
	}
	// AP mod:全部繞過裝甲直接打結構。
	toArmor, toStructure, remain = DamageApplyArmor(30, 20, true, false)
	if toArmor != 0 || toStructure != 30 || remain != 20 {
		t.Errorf("AP mod 繞過裝甲: got (%d,%d,%d),預期 (0,30,20)", toArmor, toStructure, remain)
	}
	// 目標有 Heavy Armor / Xentronium Armor:AP 失效,退回一般規則。
	toArmor, toStructure, remain = DamageApplyArmor(30, 20, true, true)
	if toArmor != 20 || toStructure != 10 || remain != 0 {
		t.Errorf("AP 被 Heavy/Xentronium Armor 抵銷: got (%d,%d,%d),預期 (20,10,0)", toArmor, toStructure, remain)
	}
}

// TestDamageSphericalShipRollCount 對照手冊「Thus a frigate gets one roll, a destroyer two
// rolls, etc.」
func TestDamageSphericalShipRollCount(t *testing.T) {
	cases := []struct {
		class CombatShipClass
		want  int
	}{
		{SHIP_FRIGATE, 1},
		{SHIP_DESTROYER, 2},
		{SHIP_CRUISER, 3},
		{SHIP_BATTLESHIP, 4},
		{SHIP_TITAN, 5},
		{SHIP_DOOMSTAR, 6},
	}
	for _, c := range cases {
		if got := DamageSphericalShipRollCount(c.class); got != c.want {
			t.Errorf("DamageSphericalShipRollCount(%v) = %d,預期 %d", c.class, got, c.want)
		}
	}
}

// TestDamageSphericalFlyerDestroyed 對照手冊「DESTRUCTION IF: aggD * 25 / hit points >=
// random(100)」。
func TestDamageSphericalFlyerDestroyed(t *testing.T) {
	// aggD=10, hitPoints=20 → 10*25/20=12(整數除法),roll<=12 才摧毀。
	if !DamageSphericalFlyerDestroyed(10, 20, 12) {
		t.Errorf("roll=12 應摧毀(12>=12)")
	}
	if DamageSphericalFlyerDestroyed(10, 20, 13) {
		t.Errorf("roll=13 不應摧毀(12<13)")
	}
}

// TestDamageEngineExplosionPotential 對照手冊「5 times the maximum engine hit points ...
// tripled if the ship has a Quantum Detonator onboard.」
func TestDamageEngineExplosionPotential(t *testing.T) {
	if got := DamageEngineExplosionPotential(40, false); got != 200 {
		t.Errorf("DamageEngineExplosionPotential(40,false) = %d,預期 200", got)
	}
	if got := DamageEngineExplosionPotential(40, true); got != 600 {
		t.Errorf("DamageEngineExplosionPotential(40,true) = %d,預期 600", got)
	}
}
