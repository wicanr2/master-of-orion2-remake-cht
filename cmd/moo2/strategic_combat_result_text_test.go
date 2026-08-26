package main

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestStrategicCombatRefusalsComeFromCatalog(t *testing.T) {
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		texts := []string{
			bombardmentRefusalText(lang, shell.GroundBombardResult{Reason: shell.BombardNoCombatShip}),
			invasionRefusalText(lang, shell.GroundInvasionResult{Reason: shell.GroundNoTroops}),
			mindControlRefusalText(lang, shell.GroundInvasionResult{Reason: shell.GroundRequiresTelepathy}),
			monsterCombatRefusalText(lang, shell.MonsterCombatNoBlueprint),
		}
		for _, got := range texts {
			if got == "" || strings.Contains(got, "galaxy.") {
				t.Fatalf("lang=%v unresolved refusal: %q", lang, got)
			}
		}
	}
}
