package gamedata

import "testing"

func TestAdvancedCivilizationExtraPlanetQuota(t *testing.T) {
	for _, tc := range []struct {
		stars, players, want int
	}{{20, 4, 2}, {21, 4, 2}, {40, 8, 2}, {4, 8, 0}, {20, 0, 0}} {
		if got := AdvancedCivilizationExtraPlanetQuota(tc.stars, tc.players); got != tc.want {
			t.Errorf("AdvancedCivilizationExtraPlanetQuota(%d,%d)=%d，預期 %d",
				tc.stars, tc.players, got, tc.want)
		}
	}
}

func TestAdvancedCivilizationStartingBC(t *testing.T) {
	for _, tc := range []struct {
		raw, want int
	}{{-1, 100}, {0, 200}, {1, 300}, {2, 400}} {
		if got := AdvancedCivilizationStartingBC(tc.raw); got != tc.want {
			t.Errorf("AdvancedCivilizationStartingBC(%d)=%d，預期 %d", tc.raw, got, tc.want)
		}
	}
}

func TestAdvancedCivilizationChooseRoundRobinAndConflict(t *testing.T) {
	candidates := [][]AdvancedCivilizationCandidate{
		{{Planet: 7, Distance: 9, Worth: 90}, {Planet: 8, Distance: 9, Worth: 80}},
		{{Planet: 7, Distance: 9, Worth: 100}, {Planet: 9, Distance: 10, Worth: 70}},
	}
	got := AdvancedCivilizationChoose(candidates, []int{0, 1}, 2, 0, 90)
	if len(got[0]) != 2 || got[0][0] != 7 || got[0][1] != 8 {
		t.Fatalf("玩家0 round-robin 結果=%v，預期 [7 8]", got[0])
	}
	if len(got[1]) != 0 {
		t.Fatalf("玩家1 的共用行星應被玩家0占用，距離10亦應被擋，得到 %v", got[1])
	}
	got = AdvancedCivilizationChoose(candidates, []int{1, 0}, 1, 1, 100)
	if len(got[1]) != 1 || got[1][0] != 7 || len(got[0]) != 1 || got[0][0] != 8 {
		t.Fatalf("隨機化順序應改變衝突歸屬，得到 %v", got)
	}
}
