package shell

import "testing"

// TestCombatSpriteForClass 驗證艦級→CMBTSHP 色塊內 sprite 索引對照(見 docs/tech/cmbtshp-ship-sprites.md)。
func TestCombatSpriteForClass(t *testing.T) {
	cases := []struct {
		class string
		want  int
	}{
		{"巡防艦", 3},
		{"驅逐艦", 12},
		{"巡洋艦", 20},
		{"戰艦", 28},
		{"泰坦", 36},
		{"末日之星", 43},
		{"偵察艦", 3},    // default(小艦)
		{"殖民船", 3},    // default(小艦)
		{"不存在的艦級", 3}, // default
	}
	for _, c := range cases {
		if got := CombatSpriteForClass(c.class); got != c.want {
			t.Errorf("CombatSpriteForClass(%q) = %d, want %d", c.class, got, c.want)
		}
	}
}

// TestCombatSpriteForStrength 驗證戰力值反推近似艦級的邊界值(shipStrength:巡防2/驅逐4/巡洋8/戰艦16/泰坦32/末日64)。
func TestCombatSpriteForStrength(t *testing.T) {
	cases := []struct {
		st   int
		want int
	}{
		{1, 3},
		{2, 3},
		{4, 12},
		{8, 20},
		{16, 28},
		{32, 36},
		{64, 43},
		{128, 43},
	}
	for _, c := range cases {
		if got := CombatSpriteForStrength(c.st); got != c.want {
			t.Errorf("CombatSpriteForStrength(%d) = %d, want %d", c.st, got, c.want)
		}
	}
}

func TestCMBTSHPSpriteIndexUsesRawColorAndPicture(t *testing.T) {
	cases := []struct {
		color, picture int
		want           int
	}{
		{0, 0, 0},
		{0, 43, 43},
		{1, 0, 45},
		{3, 17, 152},
		{7, 43, 358},
	}
	for _, c := range cases {
		got, ok := CMBTSHPSpriteIndex(c.color, c.picture)
		if !ok || got != c.want {
			t.Errorf("CMBTSHPSpriteIndex(%d,%d) = (%d,%v), want (%d,true)",
				c.color, c.picture, got, ok, c.want)
		}
	}
	for _, c := range [][2]int{{-1, 0}, {8, 0}, {0, -1}, {0, 44}} {
		if got, ok := CMBTSHPSpriteIndex(c[0], c[1]); ok || got != 0 {
			t.Errorf("CMBTSHPSpriteIndex(%d,%d) = (%d,%v), want (0,false)", c[0], c[1], got, ok)
		}
	}
}

func TestCombatSpriteForShipPreservesRawZeroPicture(t *testing.T) {
	ship := Ship{Class: "末日之星", CombatPicture: 0, CombatPictureKnown: true}
	if got := CombatSpriteForShip(ship, 2); got != 90 {
		t.Fatalf("raw picture=0 仍是合法 CMBTSHP 索引,got %d,want 90", got)
	}
	ship.CombatPicture = 44 // monster.lbx sentinel,回到艦級 fallback
	if got := CombatSpriteForShip(ship, 2); got != 2*45+43 {
		t.Fatalf("picture=44 應避開 palette／monster sentinel fallback,got %d", got)
	}
}

func TestCMBTSHPFrameForHeadingStaysWithinTwentyFrames(t *testing.T) {
	for heading := -32; heading < 32; heading++ {
		frame := CMBTSHPFrameForHeading(heading)
		if frame < 0 || frame >= CMBTSHPFrameCount {
			t.Fatalf("heading=%d 產生越界 CMBTSHP frame=%d", heading, frame)
		}
	}
	if CMBTSHPFrameForHeading(0) != 0 || CMBTSHPFrameForHeading(16) != 0 {
		t.Fatalf("heading 0/16 應繞回同一幀")
	}
}

func TestCMBTSHPFrameAtTickIsDeterministicAndStops(t *testing.T) {
	base := CMBTSHPFrameForHeading(3)
	want := []int{base, (base + 1) % CMBTSHPFrameCount, (base + 2) % CMBTSHPFrameCount, (base + 1) % CMBTSHPFrameCount}
	for phase, expected := range want {
		got := CMBTSHPFrameAtTick(3, phase*CMBTSHPFrameHoldTicks, true)
		if got != expected {
			t.Fatalf("phase=%d CMBTSHP frame=%d,want %d", phase, got, expected)
		}
	}
	if got := CMBTSHPFrameAtTick(3, CMBTSHPMotionDurationTicks, true); got != base {
		t.Fatalf("動畫結束後應回到朝向幀=%d,got %d", base, got)
	}
	if got := CMBTSHPFrameAtTick(3, 8, false); got != base {
		t.Fatalf("靜止艦不應因 timer 自行旋轉,got %d,want %d", got, base)
	}
}
