package gamedata

import "testing"

// TestMissileSpecialDefensiveSystems 核對手冊 p123 三種特殊防禦裝置的固定機率。
func TestMissileSpecialDefensiveSystems(t *testing.T) {
	if MissileLightningFieldDestroyChance != 50 {
		t.Errorf("MissileLightningFieldDestroyChance = %d,預期 50", MissileLightningFieldDestroyChance)
	}
	if MissileCloakingDeviceMissChance != 50 {
		t.Errorf("MissileCloakingDeviceMissChance = %d,預期 50", MissileCloakingDeviceMissChance)
	}
	if MissileDisplacementDeviceMissChance != 30 {
		t.Errorf("MissileDisplacementDeviceMissChance = %d,預期 30", MissileDisplacementDeviceMissChance)
	}
	if MissileDefaultHitChance != 100 {
		t.Errorf("MissileDefaultHitChance = %d,預期 100", MissileDefaultHitChance)
	}
}

// TestMissileJamChance 核對手冊 p123 Missile Evasion 範例:
// Wide Area Jammer 艦隊加成(+70)+ Stabilizer(+25)+ 種族懲罰(-20)+ 一般艦員(+7)
// + 統帥加成一半(10/2=5)= 87;攻擊方 Tachyon Scanner 已知加成 20;飛彈具 ECCM 減半。
// P = [(70+25-20+7+(10/2))-20] / 2 = 33%
func TestMissileJamChance(t *testing.T) {
	defenderEvasionBonus := MissileJammerWideAreaFleet +
		MissileInertialStabilizer +
		MissileShipDefenseRacialBonus[0] + // -20 檔
		MissileCrewRegular +
		MissileHelmsmanEvasionBonus(10)
	if defenderEvasionBonus != 87 {
		t.Fatalf("defenderEvasionBonus = %d,預期 87", defenderEvasionBonus)
	}

	const attackerScannerBonus = 20 // 手冊範例:Tachyon Scanner 已知加成
	if got := MissileJamChance(defenderEvasionBonus, attackerScannerBonus, true); got != 33 {
		t.Errorf("MissileJamChance(87,20,ECCM) = %d,預期 33", got)
	}
	// 不含 ECCM 時不減半。
	if got := MissileJamChance(defenderEvasionBonus, attackerScannerBonus, false); got != 67 {
		t.Errorf("MissileJamChance(87,20,無ECCM) = %d,預期 67", got)
	}
}

func TestMissileHelmsmanEvasionBonus(t *testing.T) {
	if got := MissileHelmsmanEvasionBonus(10); got != 5 {
		t.Errorf("MissileHelmsmanEvasionBonus(10) = %d,預期 5", got)
	}
}

// TestMissileAMRRangeIndex 核對手冊 p125 AMR 格→Range 對照表。
func TestMissileAMRRangeIndex(t *testing.T) {
	cases := []struct {
		sq   int
		want int
	}{
		{0, 1}, {1, 1}, {2, 1}, // 0-2 → Range1
		{3, 2}, {4, 2}, {5, 2}, // 3-5 → Range2
		{6, 3}, {7, 3}, {8, 3}, // 6-8 → Range3
		{9, 4}, {10, 4}, {11, 4}, // 9-11 → Range4
		{12, 5}, {13, 5}, {14, 5}, // 12-14 → Range5
		{15, 6}, // 15-17 → Range6(AMR 最大射程只到 15)
	}
	for _, c := range cases {
		if got := MissileAMRRangeIndex(c.sq); got != c.want {
			t.Errorf("MissileAMRRangeIndex(%d) = %d,預期 %d", c.sq, got, c.want)
		}
	}
}

// TestMissileAMRChanceToHit 核對手冊 p125 核對表:Range 0-6 → 65/61/58/55/51/48/45(%)。
func TestMissileAMRChanceToHit(t *testing.T) {
	want := []int{65, 61, 58, 55, 51, 48, 45}
	for rangeIndex, exp := range want {
		if got := MissileAMRChanceToHit(rangeIndex); got != exp {
			t.Errorf("MissileAMRChanceToHit(%d) = %d,預期 %d", rangeIndex, got, exp)
		}
	}
}

// TestMissileAMREndToEnd 端到端核對:格距離 0-2 落在 Range1,命中率應為 61%
// (手冊原文:「within 0-2 squares of the ship's center, AMR fire has a 61% chance of success」)。
func TestMissileAMREndToEnd(t *testing.T) {
	for sq := 0; sq <= 2; sq++ {
		idx := MissileAMRRangeIndex(sq)
		if got := MissileAMRChanceToHit(idx); got != 61 {
			t.Errorf("sq=%d: MissileAMRChanceToHit(MissileAMRRangeIndex(%d))=%d,預期 61", sq, sq, got)
		}
	}
}

// ★ 飛彈速度:**逐項對上手冊那張附表**。
//
// 這支測試 2026-08-07 改寫。先前它測的是手冊的「明列公式」(含無條件 +4),
// 並在註解裡寫「表格與公式有落差,此處測公式」——等於把一個未解的矛盾釘成了規格。
// `Missile_Speed_` @ 0x3CD21 顯示那個 +4 是 **Fast 改造**的旗標 0x10 才加的,
// 所以**表才是一般飛彈的值**,公式是改造後的值。
func TestMissileSpeedMatchesTheManualTable(t *testing.T) {
	// 手冊附表 Speed 欄,FTL 0..6。
	want := []int{10, 12, 14, 16, 18, 20, 22}
	for ftl, w := range want {
		if got := MissileSpeed(ftl); got != w {
			t.Errorf("MissileSpeed(%d) = %d,手冊附表是 %d", ftl, got, w)
		}
	}
	// 裝了 Fast 改造才是手冊那條「明列公式」的值。
	for ftl, w := range want {
		if got := MissileSpeedFast(ftl); got != w+4 {
			t.Errorf("MissileSpeedFast(%d) = %d,預期 %d(表值 +4)", ftl, got, w+4)
		}
	}
}

// 基礎速度依武器類型分檔(原版 Missile_Speed_ 的分支表),而且有兩檔**不加 FTL 項**。
func TestMissileBaseSpeedMatchesTheOriginalBranches(t *testing.T) {
	for _, c := range []struct {
		kind        int
		boosted     bool
		wantBase    int
		wantWithFTL bool
	}{
		{0x0E, false, 12, true}, {0x11, false, 12, true},
		{0x12, false, 20, false}, {0x13, false, 20, false},
		{0x1C, false, 6, true}, {0x1C, true, 10, true},
		{0x1D, false, 8, true}, {0x1D, true, 12, true},
		{0x1E, false, 8, true}, {0x1E, true, 12, true},
		{0x1F, false, 10, true}, {0x1F, true, 14, true},
		{0x28, false, 24, false},
		{0x20, false, 0, true}, // 分支表沒有的類型
	} {
		base, withFTL := MissileBaseSpeed(c.kind, c.boosted)
		if base != c.wantBase || withFTL != c.wantWithFTL {
			t.Errorf("類型 0x%X(boosted=%v):得 (%d,%v),預期 (%d,%v)",
				c.kind, c.boosted, base, withFTL, c.wantBase, c.wantWithFTL)
		}
	}
	// 那兩檔**不隨驅動等級變**——這是原版 `xor ecx, ecx` 的意思,很容易漏抄。
	for _, kind := range []int{0x12, 0x28} {
		if MissileSpeedOf(kind, 0, false, false) != MissileSpeedOf(kind, 6, false, false) {
			t.Errorf("類型 0x%X 的速度不該隨 FTL 等級變", kind)
		}
	}
	// 標準檔(0x0E)則要隨 FTL 變,而且與 MissileSpeed 一致——正對照。
	for ftl := 0; ftl <= 6; ftl++ {
		if got, want := MissileSpeedOf(0x0E, ftl, false, false), MissileSpeed(ftl); got != want {
			t.Errorf("類型 0x0E 在 FTL %d 應等於 MissileSpeed(%d)=%d,實得 %d", ftl, ftl, want, got)
		}
	}
	// Fast 是最後才加的,而且對「不加 FTL 項」的檔一樣有效。
	if got := MissileSpeedOf(0x28, 3, true, false); got != 24+4 {
		t.Errorf("類型 0x28 + Fast 應為 28,實得 %d", got)
	}
}

// Beam Defense = 5×Speed + 彈頭加成,用**表值**算。
func TestMissileBeamDefense(t *testing.T) {
	// ftl1 Nuclear:5*12+(-10)=50;ftl4 Zeon:5*18+70=160
	if got := MissileBeamDefense(MissileFTLNuclear, MissileWarheadNuclear); got != 50 {
		t.Errorf("MissileBeamDefense(1,Nuclear) = %d,預期 50", got)
	}
	if got := MissileBeamDefense(MissileFTLAntiMatter, MissileWarheadZeon); got != 160 {
		t.Errorf("MissileBeamDefense(4,Zeon) = %d,預期 160", got)
	}
	// Fast 版本正好高 20(5×4)——先前所有飛彈都是這個值,等於預設有改造。
	if got, want := MissileBeamDefenseFast(MissileFTLNuclear, MissileWarheadNuclear), 70; got != want {
		t.Errorf("MissileBeamDefenseFast(1,Nuclear) = %d,預期 %d", got, want)
	}
	if MissileBeamDefenseFast(MissileFTLNuclear, MissileWarheadNuclear)-
		MissileBeamDefense(MissileFTLNuclear, MissileWarheadNuclear) != 5*MissileFastBonus {
		t.Error("Fast 版本與一般版本應正好差 5×4 = 20")
	}
}
