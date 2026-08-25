package gamedata

import "testing"

func TestOriginalColonyCombatHitsMatchesIDAFormula(t *testing.T) {
	if got := OriginalColonyCombatHits(8, 4, 2, []int{7, 22}); got != 94 {
		t.Fatalf("人口 8 + 士兵 4 + 戰車 2 + 兩棟一般建築應為 94，got %d", got)
	}
}

func TestOriginalColonyCombatHitsExcludesSeparateOrbitalCombatants(t *testing.T) {
	if got := OriginalColonyCombatHits(8, 4, 2, []int{8, 40, 41}); got != 14 {
		t.Fatalf("三種軌道設施不得重複計入殖民地本體，got %d", got)
	}
}

func TestOriginalColonyCombatHitsRejectsInvalidAndDuplicateInputs(t *testing.T) {
	if got := OriginalColonyCombatHits(-8, -4, -2, []int{7, 7, -1, 49}); got != 40 {
		t.Fatalf("負值歸零、重複與範圍外 ID 忽略後應只剩一棟建築 40，got %d", got)
	}
}
