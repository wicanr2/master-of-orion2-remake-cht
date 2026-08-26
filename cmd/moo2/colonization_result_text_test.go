package main

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestColonizationAndOutpostResultTextComesFromCatalog(t *testing.T) {
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		plain := colonizationRefusalText(lang, shell.ColonizationResult{Reason: shell.ColonizeNoColonyShip})
		if plain == "" || strings.Contains(plain, "galaxy.") {
			t.Fatalf("lang=%v plain refusal unresolved: %q", lang, plain)
		}
		dynamic := colonizationRefusalText(lang, shell.ColonizationResult{
			Reason: shell.ColonizeRequiresOutpost, PlanetType: gamedata.GAS_GIANT,
		})
		if dynamic == "" || strings.Contains(dynamic, "%s") || strings.Contains(dynamic, "galaxy.") {
			t.Fatalf("lang=%v dynamic refusal unresolved: %q", lang, dynamic)
		}
		monster := outpostRefusalText(lang, shell.OutpostResult{
			Reason: shell.OutpostMonsterBlocked, MonsterKind: gamedata.MonsterCrystal,
		})
		if monster == "" || strings.Contains(monster, "%s") || strings.Contains(monster, "galaxy.") {
			t.Fatalf("lang=%v monster refusal unresolved: %q", lang, monster)
		}
	}
}
