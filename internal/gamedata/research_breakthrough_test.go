package gamedata

import "testing"

func TestResearchBreakthroughChanceBoundaries(t *testing.T) {
	for _, tc := range []struct {
		cost, progress, want int
	}{
		{400, 400, 0},
		{400, 401, 1},
		{400, 440, 10},
		{400, 799, 99},
		{400, 800, 100},
		{0, 100, 0},
	} {
		if got := ResearchBreakthroughChance(tc.cost, tc.progress); got != tc.want {
			t.Errorf("cost=%d progress=%d: got %d want %d", tc.cost, tc.progress, got, tc.want)
		}
	}
}

func TestResearchBreakthroughSucceededUsesOneToHundredInclusive(t *testing.T) {
	if !ResearchBreakthroughSucceeded(10, 10) || ResearchBreakthroughSucceeded(10, 11) {
		t.Fatal("突破應以 roll <= chance 判定")
	}
	if ResearchBreakthroughSucceeded(100, 0) || ResearchBreakthroughSucceeded(100, 101) {
		t.Fatal("原版 random(100) 範圍外的擲骰必須失敗即關閉")
	}
}
