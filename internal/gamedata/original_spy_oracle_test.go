package gamedata

import "testing"

func TestOriginalSpyEmpireBonuses(t *testing.T) {
	attack, defense := OriginalSpyEmpireBonuses(45, 20, 1, 18, 12, -10)
	if attack != 93 || defense != 77 {
		t.Fatalf("兩張原版計分表=(%d,%d), want (93,77)", attack, defense)
	}
}
