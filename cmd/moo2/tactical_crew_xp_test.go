package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestCompactEnemyCasualtiesCountsDestroyedButNotCapturedHullClasses(t *testing.T) {
	tactical := &tacticalScreen{enemy: []shell.CombatShip{
		{Name: "存活艦", HP: 1, SizeClass: gamedata.CombatShipClass(5)},
		{Name: "擊沉巡洋艦", HP: 0, SizeClass: gamedata.CombatShipClass(2)},
		{Name: "俘獲戰艦", HP: 0, Captured: true, SizeClass: gamedata.CombatShipClass(3)},
	}}
	tactical.compactEnemyCasualties()
	if len(tactical.enemy) != 1 || tactical.enemy[0].Name != "存活艦" {
		t.Fatalf("應只留下存活艦，got %+v", tactical.enemy)
	}
	if got := tactical.destroyedEnemyHullClassSum; got != 3 {
		t.Fatalf("只應累加被擊沉巡洋艦的 1-based class 3，got %d", got)
	}
	// 已移除的 casualty 不會在下一次壓縮重複計分。
	tactical.compactEnemyCasualties()
	if got := tactical.destroyedEnemyHullClassSum; got != 3 {
		t.Fatalf("重複壓縮不應重複累加，got %d", got)
	}
}
