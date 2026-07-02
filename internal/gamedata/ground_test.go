package gamedata

import "testing"

// TestGroundMarineBarracksCap 手冊 p.77:「half the current population... or half the base
// maximum population... whichever is less」。
func TestGroundMarineBarracksCap(t *testing.T) {
	cases := []struct {
		pop, planetMax int
		warlord        bool
		want           int
	}{
		{pop: 10, planetMax: 20, warlord: false, want: 5}, // min(10/2=5, 20/2=10) = 5
		{pop: 20, planetMax: 10, warlord: false, want: 5}, // min(10, 5) = 5
		{pop: 7, planetMax: 20, warlord: false, want: 3},  // 7/2=3(整數除法捨去)
		{pop: 10, planetMax: 20, warlord: true, want: 10}, // 5 * 2(Warlord)
		{pop: 0, planetMax: 20, warlord: false, want: 0},
	}
	for _, c := range cases {
		if got := GroundMarineBarracksCap(c.pop, c.planetMax, c.warlord); got != c.want {
			t.Errorf("GroundMarineBarracksCap(%d,%d,%v) = %d,預期 %d", c.pop, c.planetMax, c.warlord, got, c.want)
		}
	}
}

// TestGroundArmorBarracksCap 手冊 p.79:「one-quarter the current population... or a quarter
// of the base maximum population... whichever is less」。
func TestGroundArmorBarracksCap(t *testing.T) {
	cases := []struct {
		pop, planetMax int
		warlord        bool
		want           int
	}{
		{pop: 40, planetMax: 100, warlord: false, want: 10}, // min(40/4=10, 100/4=25) = 10
		{pop: 100, planetMax: 40, warlord: false, want: 10}, // min(25, 10) = 10
		{pop: 9, planetMax: 100, warlord: false, want: 2},   // 9/4=2(整數除法捨去)
		{pop: 40, planetMax: 100, warlord: true, want: 20},  // 10 * 2(Warlord)
	}
	for _, c := range cases {
		if got := GroundArmorBarracksCap(c.pop, c.planetMax, c.warlord); got != c.want {
			t.Errorf("GroundArmorBarracksCap(%d,%d,%v) = %d,預期 %d", c.pop, c.planetMax, c.warlord, got, c.want)
		}
	}
}

// TestGroundMarineBarracksUnits 手冊 p.77:初始 4 個單位,之後每 5 回合 +1,受上限夾住。
func TestGroundMarineBarracksUnits(t *testing.T) {
	// pop=20, planetMax=20 → cap = min(10,10) = 10
	cases := map[int]int{
		0:  4, // 剛建成
		4:  4, // 不滿 5 回合不增加
		5:  5, // 滿 5 回合 +1
		9:  5,
		10: 6,
		29: 9,  // 4 + 29/5 = 4+5 = 9
		30: 10, // 4 + 30/5 = 10,已達上限
		99: 10, // 上限夾住
	}
	for turns, want := range cases {
		if got := GroundMarineBarracksUnits(turns, 20, 20, false); got != want {
			t.Errorf("GroundMarineBarracksUnits(%d,20,20,false) = %d,預期 %d", turns, got, want)
		}
	}
}

// TestGroundArmorBarracksUnits 手冊 p.79:初始 2 個戰車營,之後每 5 回合 +1,受上限夾住。
func TestGroundArmorBarracksUnits(t *testing.T) {
	// pop=40, planetMax=40 → cap = min(10,10) = 10
	cases := map[int]int{
		0:  2,
		4:  2,
		5:  3,
		39: 9,  // 2 + 39/5 = 2+7 = 9
		40: 10, // 2 + 40/5 = 10,已達上限
		99: 10,
	}
	for turns, want := range cases {
		if got := GroundArmorBarracksUnits(turns, 40, 40, false); got != want {
			t.Errorf("GroundArmorBarracksUnits(%d,40,40,false) = %d,預期 %d", turns, got, want)
		}
	}
}

// TestGroundMarineHitsToKill 手冊 p.129 Planet Hits(Marine 基礎 1 hit,modified by Heavy-G,
// Powered Armor)+ p.24(High-G +1 hit)+ p.80(Powered Armor +1 hit)。
func TestGroundMarineHitsToKill(t *testing.T) {
	cases := []struct {
		highG, poweredArmor bool
		want                int
	}{
		{false, false, 1},
		{true, false, 2},
		{false, true, 2},
		{true, true, 3},
	}
	for _, c := range cases {
		if got := GroundMarineHitsToKill(c.highG, c.poweredArmor); got != c.want {
			t.Errorf("GroundMarineHitsToKill(%v,%v) = %d,預期 %d", c.highG, c.poweredArmor, got, c.want)
		}
	}
}

// TestGroundTankHitsToKill 手冊 p.129 Planet Hits(Tank 基礎 2 hits,modified by Heavy-G)。
func TestGroundTankHitsToKill(t *testing.T) {
	if got := GroundTankHitsToKill(false); got != 2 {
		t.Errorf("GroundTankHitsToKill(false) = %d,預期 2", got)
	}
	if got := GroundTankHitsToKill(true); got != 3 {
		t.Errorf("GroundTankHitsToKill(true) = %d,預期 3", got)
	}
}

// TestGroundBattleoidHitsToKill 手冊 p.81:「take 3 hits to destroy」,固定值不受其他修飾。
func TestGroundBattleoidHitsToKill(t *testing.T) {
	if GroundBattleoidHitsToKill != 3 {
		t.Errorf("GroundBattleoidHitsToKill = %d,預期 3", GroundBattleoidHitsToKill)
	}
}

// TestGroundArmorTechBonus 逐條核對手冊 p.90-92、p.114 的地面部隊戰力加成。
func TestGroundArmorTechBonus(t *testing.T) {
	cases := map[Technology]int{
		TECH_TITANIUM_ARMOR:   0, // 手冊未列基礎裝甲的地面加成
		TECH_TRITANIUM_ARMOR:  10,
		TECH_ZORTRIUM_ARMOR:   15,
		TECH_NEUTRONIUM_ARMOR: 20,
		TECH_ADAMANTIUM_ARMOR: 25,
		TECH_XENTRONIUM_ARMOR: 30,
		TECH_HEAVY_ARMOR:      0, // 艦用裝甲,手冊未提供地面加成
	}
	for tech, want := range cases {
		if got := GroundArmorTechBonus(tech); got != want {
			t.Errorf("GroundArmorTechBonus(%d) = %d,預期 %d", tech, got, want)
		}
	}
}

// TestGroundEquipmentTechBonus 核對手冊 p.80、p.108、p.109 的地面裝備戰鬥評等加成。
func TestGroundEquipmentTechBonus(t *testing.T) {
	cases := map[Technology]int{
		TECH_POWERED_ARMOR:    10,
		TECH_ANTIGRAV_HARNESS: 10,
		TECH_PERSONAL_SHIELD:  20,
		TECH_BATTLEOIDS:       0, // 見 GroundBattleoidCombatBonus,不走這個表
	}
	for tech, want := range cases {
		if got := GroundEquipmentTechBonus(tech); got != want {
			t.Errorf("GroundEquipmentTechBonus(%d) = %d,預期 %d", tech, got, want)
		}
	}
}

// TestGroundBattleoidCombatBonus 手冊 p.81:「a ground combat rating 10 higher than a tank」。
func TestGroundBattleoidCombatBonus(t *testing.T) {
	if GroundBattleoidCombatBonus != 10 {
		t.Errorf("GroundBattleoidCombatBonus = %d,預期 10", GroundBattleoidCombatBonus)
	}
}

// TestGroundRaceCombatBonus 手冊 p.15-16:Bulrathi +10 / Gnolam -10,其他種族 0。
func TestGroundRaceCombatBonus(t *testing.T) {
	if got := GroundRaceCombatBonus(GroundRaceBulrathi); got != 10 {
		t.Errorf("GroundRaceCombatBonus(Bulrathi) = %d,預期 10", got)
	}
	if got := GroundRaceCombatBonus(GroundRaceGnolam); got != -10 {
		t.Errorf("GroundRaceCombatBonus(Gnolam) = %d,預期 -10", got)
	}
	if got := GroundRaceCombatBonus(GroundRaceOther); got != 0 {
		t.Errorf("GroundRaceCombatBonus(Other) = %d,預期 0", got)
	}
}

// TestGroundApplyLowGPenalty 手冊 p.24:「Low-G troops suffer a 10% penalty during ground
// combat.」
func TestGroundApplyLowGPenalty(t *testing.T) {
	cases := map[int]int{
		100: 90,
		50:  45,
		10:  9,
		0:   0,
		// 7*10/100 在整數運算中先乘後除:7*10=70,70/100=0(整數除法捨去),故 7-0=7。
		7: 7,
	}
	for strength, want := range cases {
		if got := GroundApplyLowGPenalty(strength); got != want {
			t.Errorf("GroundApplyLowGPenalty(%d) = %d,預期 %d", strength, got, want)
		}
	}
}

// TestGroundSubterraneanBonus 手冊 p.24:「+10 ground combat bonus when defending their
// colonies」,僅防守時生效。
func TestGroundSubterraneanBonus(t *testing.T) {
	if got := GroundSubterraneanBonus(true); got != 10 {
		t.Errorf("GroundSubterraneanBonus(true) = %d,預期 10", got)
	}
	if got := GroundSubterraneanBonus(false); got != 0 {
		t.Errorf("GroundSubterraneanBonus(false) = %d,預期 0", got)
	}
}

// TestGroundBombHitsFromDamage 手冊 MANUAL_150.html p.129:「This damage is divided by 100
// to get the displayed number... The maximum number of bomb hits for the fleet in orbit is
// 320.」
func TestGroundBombHitsFromDamage(t *testing.T) {
	cases := map[int]int{
		0:     0,
		99:    0,
		100:   1,
		250:   2,
		32000: 320, // 剛好達上限
		32099: 320, // 未滿下一整百,仍是 320
		40000: 320, // 超過上限,夾住
		-50:   0,   // 防呆
	}
	for damage, want := range cases {
		if got := GroundBombHitsFromDamage(damage); got != want {
			t.Errorf("GroundBombHitsFromDamage(%d) = %d,預期 %d", damage, got, want)
		}
	}
}

// TestGroundMaxBombHitsPerFleet 手冊 MANUAL_150.html p.129 上限常數本身。
func TestGroundMaxBombHitsPerFleet(t *testing.T) {
	if GroundMaxBombHitsPerFleet != 320 {
		t.Errorf("GroundMaxBombHitsPerFleet = %d,預期 320", GroundMaxBombHitsPerFleet)
	}
	if GroundPlanetMissileEvasionPercent != 7 {
		t.Errorf("GroundPlanetMissileEvasionPercent = %d,預期 7", GroundPlanetMissileEvasionPercent)
	}
}

// TestGroundPlanetTotalHits 手冊 MANUAL_150.html p.129 Planet Hits 表逐項加總,手算對照:
// 5 棟建築(5)+ 有儲存生產(1)+ 3 個整數人口(3)+ 1 個人口零頭(1)
// + 4 個 Marine(每個 1 hit = 4)+ 2 個 Tank(每個 2 hits = 4)
// = 5+1+3+1+4+4 = 18
func TestGroundPlanetTotalHits(t *testing.T) {
	got := GroundPlanetTotalHits(5, true, 3, 1, 4, GroundMarineHitsToKill(false, false), 2, GroundTankHitsToKill(false))
	if got != 18 {
		t.Errorf("GroundPlanetTotalHits(...) = %d,預期 18", got)
	}

	// 不含儲存生產、High-G 種族 Marine(2 hits each)+ Battleoid(3 hits each)
	got2 := GroundPlanetTotalHits(0, false, 0, 0, 2, GroundMarineHitsToKill(true, false), 1, GroundBattleoidHitsToKill)
	// 2*2 + 1*3 = 7
	if got2 != 7 {
		t.Errorf("GroundPlanetTotalHits(...) = %d,預期 7", got2)
	}
}
