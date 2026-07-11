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
