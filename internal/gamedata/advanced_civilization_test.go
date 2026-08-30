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
