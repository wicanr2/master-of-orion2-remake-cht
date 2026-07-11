package gamedata

import "testing"

// weapon_mods_test.go 驗證武器改造(mod)佔格/成本/命中/傷害公式,逐一對照
// weapon_mods.go 各常數註解引用的 GAME_MANUAL.pdf p.115-118 原文數字。

func TestWeaponSpaceWithMods_NoModsIsNeutral(t *testing.T) {
	if got := WeaponSpaceWithMods(10, nil); got != 10 {
		t.Errorf("無 mod 應中性回傳基礎值,got %d want 10", got)
	}
}

func TestWeaponSpaceWithMods_HeavyMountDoublesSpace(t *testing.T) {
	// 手冊:「This modification increases the size and cost of the weapon by 100%」。
	if got := WeaponSpaceWithMods(10, []WeaponModCode{ModHeavyMount}); got != 20 {
		t.Errorf("HV 應 +100%%(佔格倍增),got %d want 20", got)
	}
}

func TestWeaponSpaceWithMods_PointDefenseHalvesSpace(t *testing.T) {
	// 手冊:「This modification decreases the size and cost of the weapon by half (50%)」。
	if got := WeaponSpaceWithMods(10, []WeaponModCode{ModPointDefense}); got != 5 {
		t.Errorf("PD 應 -50%%(佔格減半),got %d want 5", got)
	}
}

func TestWeaponSpaceWithMods_AutoFireFlatBonus(t *testing.T) {
	// 手冊:「increases the size and cost of the weapon by 50」(固定值,非百分比)。
	if got := WeaponSpaceWithMods(10, []WeaponModCode{ModAutoFire}); got != 60 {
		t.Errorf("AF 應固定 +50,got %d want 60", got)
	}
}

func TestWeaponSpaceWithMods_ArmorPiercingAndShieldPiercing(t *testing.T) {
	if got := WeaponSpaceWithMods(10, []WeaponModCode{ModArmorPiercing}); got != 15 {
		t.Errorf("AP 應 +50%%,got %d want 15", got)
	}
	if got := WeaponSpaceWithMods(10, []WeaponModCode{ModShieldPiercing}); got != 15 {
		t.Errorf("SP 應 +50%%,got %d want 15", got)
	}
}

func TestWeaponSpaceWithMods_EnvelopingAndNoRangeDissipation(t *testing.T) {
	if got := WeaponSpaceWithMods(10, []WeaponModCode{ModEnveloping}); got != 20 {
		t.Errorf("ENV 應 +100%%,got %d want 20", got)
	}
	if got := WeaponSpaceWithMods(10, []WeaponModCode{ModNoRangeDissipation}); got != 12 {
		t.Errorf("NR 應 +25%%(10*1.25=12.5→整數除法12),got %d want 12", got)
	}
}

func TestWeaponSpaceWithMods_Stacking(t *testing.T) {
	// AP(+50%) + CO(+50%) 加總一次套用 = +100%,再加 AF 固定 +50。
	got := WeaponSpaceWithMods(20, []WeaponModCode{ModArmorPiercing, ModContinuousFire, ModAutoFire})
	want := 20 + 20*100/100 + WeaponModAutoFireFlatSpaceCost // 20+20+50=90
	if got != want {
		t.Errorf("mod 疊加,got %d want %d", got, want)
	}
}

func TestWeaponCostWithMods_SharesSpaceFormula(t *testing.T) {
	// 手冊:「adds to the size AND cost」,同一套公式套用在成本上。
	if got := WeaponCostWithMods(40, []WeaponModCode{ModHeavyMount}); got != 80 {
		t.Errorf("成本應與佔格用同一套公式,got %d want 80", got)
	}
}

func TestWeaponModNetAttackBonus(t *testing.T) {
	cases := []struct {
		mods []WeaponModCode
		want int
	}{
		{nil, 0},
		{[]WeaponModCode{ModContinuousFire}, 25},
		{[]WeaponModCode{ModAutoFire}, -20},
		{[]WeaponModCode{ModContinuousFire, ModAutoFire}, 5},
	}
	for _, c := range cases {
		if got := WeaponModNetAttackBonus(c.mods); got != c.want {
			t.Errorf("WeaponModNetAttackBonus(%v)=%d want %d", c.mods, got, c.want)
		}
	}
}

func TestWeaponModPDBonus(t *testing.T) {
	if got := WeaponModPDBonus(nil); got != 0 {
		t.Errorf("無 PD 應回 0,got %d", got)
	}
	if got := WeaponModPDBonus([]WeaponModCode{ModPointDefense}); got != 25 {
		t.Errorf("PD 應回 25,got %d", got)
	}
}

func TestWeaponModDamageBonuses(t *testing.T) {
	hv, pd := WeaponModDamageBonuses([]WeaponModCode{ModHeavyMount})
	if hv != DamageMountBonusHeavy || pd != 0 {
		t.Errorf("HV 應給 hvBonus=%d pdPenalty=0,got hv=%d pd=%d", DamageMountBonusHeavy, hv, pd)
	}
	hv, pd = WeaponModDamageBonuses([]WeaponModCode{ModPointDefense})
	if hv != 0 || pd != DamageMountPenaltyPointDefense {
		t.Errorf("PD 應給 hvBonus=0 pdPenalty=%d,got hv=%d pd=%d", DamageMountPenaltyPointDefense, hv, pd)
	}
	hv, pd = WeaponModDamageBonuses(nil)
	if hv != 0 || pd != 0 {
		t.Errorf("無 mod 應中性(0,0),got hv=%d pd=%d", hv, pd)
	}
}

// 端到端:HV 武器實際傷害應為基礎值的 150%(透過既有 DamageMountAdjustedValue)。
func TestWeaponModDamageBonuses_EndToEndHeavyMount(t *testing.T) {
	hv, pd := WeaponModDamageBonuses([]WeaponModCode{ModHeavyMount})
	got := DamageMountAdjustedValue(100, hv, 0, pd, 0)
	if got != 150 {
		t.Errorf("HV 武器 100 基礎傷害應變 150(150%%),got %d", got)
	}
}

func TestWeaponModDamageBonuses_EndToEndPointDefense(t *testing.T) {
	hv, pd := WeaponModDamageBonuses([]WeaponModCode{ModPointDefense})
	got := DamageMountAdjustedValue(100, hv, 0, pd, 0)
	if got != 50 {
		t.Errorf("PD 武器 100 基礎傷害應減半,got %d", got)
	}
}

func TestWeaponModEnvelopingMultiply(t *testing.T) {
	if got := WeaponModEnvelopingMultiply(10, nil); got != 10 {
		t.Errorf("無 ENV 應中性,got %d", got)
	}
	if got := WeaponModEnvelopingMultiply(10, []WeaponModCode{ModEnveloping}); got != 40 {
		t.Errorf("ENV 應四倍,got %d want 40", got)
	}
}

func TestWeaponModArmorPiercingAndShieldPiercing(t *testing.T) {
	if WeaponModArmorPiercing(nil) {
		t.Error("無 AP 應回 false")
	}
	if !WeaponModArmorPiercing([]WeaponModCode{ModArmorPiercing}) {
		t.Error("有 AP 應回 true")
	}
	if WeaponModShieldPiercing(nil) {
		t.Error("無 SP 應回 false")
	}
	if !WeaponModShieldPiercing([]WeaponModCode{ModShieldPiercing}) {
		t.Error("有 SP 應回 true")
	}
}

func TestCombatRangeLevelForBeamMods(t *testing.T) {
	sq := 12 // Regular level = ceil(12/3) = 4
	regular := CombatRangeLevel(sq)
	heavy := CombatRangeLevelHeavy(sq)
	pd := CombatRangeLevelPointDefense(sq)
	if got := CombatRangeLevelForBeamMods(sq, nil); got != regular {
		t.Errorf("無 mod 應用一般表,got %d want %d", got, regular)
	}
	if got := CombatRangeLevelForBeamMods(sq, []WeaponModCode{ModHeavyMount}); got != heavy {
		t.Errorf("HV 應用 Heavy 表,got %d want %d", got, heavy)
	}
	if got := CombatRangeLevelForBeamMods(sq, []WeaponModCode{ModPointDefense}); got != pd {
		t.Errorf("PD 應用 PointDefense 表,got %d want %d", got, pd)
	}
	if heavy == regular || pd == regular {
		t.Fatal("測試前提:所選 squares 需讓三張表算出不同 level,否則測不出差異(換個 squares 值)")
	}
}

func TestWeaponModSpaceCostPercent_LookupTable(t *testing.T) {
	cases := map[WeaponModCode]int{
		ModHeavyMount:         100,
		ModPointDefense:       -50,
		ModContinuousFire:     50,
		ModArmorPiercing:      50,
		ModEnveloping:         100,
		ModNoRangeDissipation: 25,
		ModShieldPiercing:     50,
	}
	for mod, want := range cases {
		got, ok := WeaponModSpaceCostPercent(mod)
		if !ok || got != want {
			t.Errorf("WeaponModSpaceCostPercent(%s)=%d,%v want %d,true", mod, got, ok, want)
		}
	}
	if _, ok := WeaponModSpaceCostPercent(ModAutoFire); ok {
		t.Error("AF 是固定值不是百分比,WeaponModSpaceCostPercent 應回 ok=false")
	}
}
