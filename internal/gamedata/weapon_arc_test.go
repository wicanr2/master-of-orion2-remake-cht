package gamedata

import "testing"

func TestWeaponArcCostPercent(t *testing.T) {
	cases := []struct {
		arc  WeaponArc
		want int
	}{
		{ARC_FWD, 0},
		{ARC_BACK, 0},
		{ARC_FWD_EXT, 25},
		{ARC_BACK_EXT, 25},
		{ARC_360, 50},
		{ARC_MONSTER_360, 50},
	}
	for _, tc := range cases {
		if got := WeaponArcCostPercent(tc.arc); got != tc.want {
			t.Errorf("WeaponArcCostPercent(%d)=%d, want %d", tc.arc, got, tc.want)
		}
	}
}

func TestWeaponArcAdjustedValue(t *testing.T) {
	if got := WeaponArcAdjustedValue(10, ARC_FWD_EXT); got != 12 {
		t.Fatalf("10 格的 Fwd Ext 應為 12,得到 %d", got)
	}
	if got := WeaponArcAdjustedValue(10, ARC_360); got != 15 {
		t.Fatalf("10 格的 360 度應為 15,得到 %d", got)
	}
	if got := WeaponArcAdjustedValue(25, ARC_FWD_EXT); got != 31 {
		t.Fatalf("25 格的 Fwd Ext 應為 31(25+6),得到 %d", got)
	}
	if got := WeaponArcAdjustedValue(0, ARC_360); got != 0 {
		t.Fatalf("零基礎值不應被火線角變成正數,得到 %d", got)
	}
}

func TestCombatFacingForVectorMatchesOriginalSixteenDirections(t *testing.T) {
	cases := []struct {
		dx, dy int
		want   int
	}{
		{1, 0, 0},   // 右
		{1, -1, 2},  // 右上
		{0, -1, 4},  // 上
		{-1, -1, 6}, // 左上
		{-1, 0, 8},  // 左
		{-1, 1, 10}, // 左下
		{0, 1, 12},  // 下
		{1, 1, 14},  // 右下
	}
	for _, tc := range cases {
		if got := CombatFacingForVector(tc.dx, tc.dy); got != tc.want {
			t.Errorf("CombatFacingForVector(%d,%d)=%d, want %d", tc.dx, tc.dy, got, tc.want)
		}
	}
}

func TestRelativeBearingMaskPreservesOriginalOverlappingBoundaries(t *testing.T) {
	cases := []struct {
		angle int
		want  int
	}{
		{0, 3}, {60, 7}, {90, 6}, {120, 14}, {180, 12}, {240, 14}, {270, 6}, {300, 7}, {359, 3},
	}
	for _, tc := range cases {
		if got := RelativeBearingMask(tc.angle); got != tc.want {
			t.Errorf("RelativeBearingMask(%d)=%d, want %d", tc.angle, got, tc.want)
		}
	}
	if got := RelativeBearingMaskForVector(1, 0, 8); got != 12 {
		t.Fatalf("朝向左但目標在右側應是相對後方遮罩 12,得到 %d", got)
	}
}

func TestWeaponArcAllowsRelativeBearing(t *testing.T) {
	forward := RelativeBearingMaskForVector(1, 0, 0)
	back := RelativeBearingMaskForVector(-1, 0, 0)
	if !WeaponArcAllowsRelativeBearing(ARC_FWD, forward) {
		t.Fatal("前向射界應允許正前方")
	}
	if WeaponArcAllowsRelativeBearing(ARC_FWD, back) {
		t.Fatal("前向射界不應允許正後方")
	}
	if !WeaponArcAllowsRelativeBearing(ARC_360, back) || !WeaponArcAllowsRelativeBearing(ARC_MONSTER_360, back) {
		t.Fatal("兩種全向 raw arc 都應允許後方")
	}
}
