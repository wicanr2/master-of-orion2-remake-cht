package gamedata

import "testing"

// TestShipHullSpace 對照 GAME_MANUAL.pdf p.121 的 Class/Space 表逐列核對。
func TestShipHullSpace(t *testing.T) {
	cases := []struct {
		name  string
		class CombatShipClass
		want  int
	}{
		{"Frigate", SHIP_FRIGATE, 25},
		{"Destroyer", SHIP_DESTROYER, 60},
		{"Cruiser", SHIP_CRUISER, 120},
		{"Battleship", SHIP_BATTLESHIP, 250},
		{"Titan", SHIP_TITAN, 500},
		{"DoomStar", SHIP_DOOMSTAR, 1200},
	}
	for _, c := range cases {
		if got := ShipHullSpace(c.class); got != c.want {
			t.Errorf("ShipHullSpace(%s)=%d, want %d", c.name, got, c.want)
		}
	}
}

// TestShipHullSpaceOutOfRange 邊界:非法 class 回 0,不 panic。
func TestShipHullSpaceOutOfRange(t *testing.T) {
	if got := ShipHullSpace(-1); got != 0 {
		t.Errorf("ShipHullSpace(-1)=%d, want 0", got)
	}
	if got := ShipHullSpace(CombatShipClass(6)); got != 0 {
		t.Errorf("ShipHullSpace(6)=%d, want 0", got)
	}
}

// TestWeaponSpaceByName 核對手冊 p.124 確認值(Size 欄)。
func TestWeaponSpaceByName(t *testing.T) {
	cases := []struct {
		name string
		want int
	}{
		{"雷射", 10},
		{"質量投射器", 10},
		{"中子爆破槍", 10},
		{"核融合光束", 10},
		{"高斯砲", 10},
		{"相位砲", 10},
		{"電漿砲", 25},
		{"死光", 30},
	}
	for _, c := range cases {
		if got := WeaponSpaceByName[c.name]; got != c.want {
			t.Errorf("WeaponSpaceByName[%s]=%d, want %d", c.name, got, c.want)
		}
	}
}

// TestSpecialSpace 估計模型的行為驗證(非手冊精確值,見 SpecialSpaceEstimatePercent 註解)。
func TestSpecialSpace(t *testing.T) {
	if got := SpecialSpace(1200, true); got != 60 { // Doom Star 1200 * 5% = 60
		t.Errorf("SpecialSpace(1200,true)=%d, want 60", got)
	}
	if got := SpecialSpace(25, true); got != 1 { // Frigate 25 * 5% = 1.25 → 1(捨去,最少 1 格)
		t.Errorf("SpecialSpace(25,true)=%d, want 1", got)
	}
	if got := SpecialSpace(1200, false); got != 0 {
		t.Errorf("SpecialSpace(1200,false)=%d, want 0 (未選裝)", got)
	}
	if got := SpecialSpace(0, true); got != 0 {
		t.Errorf("SpecialSpace(0,true)=%d, want 0 (無艦體空間)", got)
	}
}

// TestShipDesignFitsSmallVsLargeHull 小艦體塞大元件超格、大艦體容納的邊界情境
// (以手冊確認的 Frigate=25 / Doom Star=1200 空間,套死光 Size=30 為例)。
func TestShipDesignFitsSmallVsLargeHull(t *testing.T) {
	deathRaySpace := WeaponSpaceByName["死光"]    // 30
	frigateSpace := ShipHullSpace(SHIP_FRIGATE) // 25
	if deathRaySpace <= frigateSpace {
		t.Fatalf("前提不成立:死光(%d)應該比護衛艦空間(%d)大,才能驗證超格情境", deathRaySpace, frigateSpace)
	}
	if frigateSpace-deathRaySpace >= 0 {
		t.Errorf("護衛艦裝死光應該超格(space used %d > hull %d)", deathRaySpace, frigateSpace)
	}
	doomStarSpace := ShipHullSpace(SHIP_DOOMSTAR) // 1200
	if deathRaySpace > doomStarSpace {
		t.Errorf("末日之星裝死光不應超格(space used %d, hull %d)", deathRaySpace, doomStarSpace)
	}
}
