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

// 飛彈 Beam Defense(手冊 p117-120):以明列公式 Speed=12+2*(ftl-1)+4、BeamDefense=5*Speed+bonus 驗證
// (手冊表格 Speed 欄與公式有 +4 落差,見 missile.go 檔頭矛盾註解;此處測公式)。
func TestMissileBeamDefense(t *testing.T) {
	if got := MissileSpeed(1); got != 16 { // 12+0+4
		t.Errorf("MissileSpeed(1) = %d,預期 16", got)
	}
	if got := MissileSpeed(4); got != 22 { // 12+6+4
		t.Errorf("MissileSpeed(4) = %d,預期 22", got)
	}
	// ftl1 Nuclear:5*16+(-10)=70;ftl4 Zeon:5*22+70=180
	if got := MissileBeamDefense(MissileFTLNuclear, MissileWarheadNuclear); got != 70 {
		t.Errorf("MissileBeamDefense(1,Nuclear) = %d,預期 70", got)
	}
	if got := MissileBeamDefense(MissileFTLAntiMatter, MissileWarheadZeon); got != 180 {
		t.Errorf("MissileBeamDefense(4,Zeon) = %d,預期 180", got)
	}
}
