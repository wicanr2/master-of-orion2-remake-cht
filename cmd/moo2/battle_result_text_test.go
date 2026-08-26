package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestBattleResultCatalogAndTextBounds(t *testing.T) {
	keys := []string{"battle.result.victory", "battle.result.defeat", "battle.result.against",
		"battle.result.start", "battle.result.round", "battle.result.losses",
		"battle.result.enemy.antaran_home_defense", "battle.result.enemy.unknown"}
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		for _, key := range keys {
			if got := uiText(lang, key); got == "" || got == key {
				t.Fatalf("lang=%v key=%s unresolved: %q", lang, key, got)
			}
		}
		against := fmt.Sprintf(uiText(lang, "battle.result.against"), "AN EXTREMELY LONG ENEMY EMPIRE AND FLEET NAME")
		checkClippedTextFits(t, fnt, battleResultTextRect(0), against, 15)
		round := battleRoundText(lang, shell.BattleRoundResult{Round: 99, EnemyDestroyed: 999, PlayerDestroyed: 999})
		checkClippedTextFits(t, fnt, battleResultLogTextRect(0), round, 12)
		losses := fmt.Sprintf(uiText(lang, "battle.result.losses"), 999, 999)
		checkClippedTextFits(t, fnt, battleResultLossTextRect(6), losses, 13)
	}
}

func TestBattleResultSourceHasNoInlineTranslation(t *testing.T) {
	raw, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "func (b *sceneBuilder) battleResult()")
	end := strings.Index(source[start:], "// council 建")
	if start < 0 || end < 0 {
		t.Fatal("battleResult source slice not found")
	}
	if strings.Contains(source[start:start+end], ".tr(") {
		t.Fatal("battleResult() must not embed translated player text")
	}
}

func TestBattleResultSpecialEnemyIsLocalizedFromTypedKind(t *testing.T) {
	result := &shell.BattleResult{EnemyKind: shell.BattleEnemyAntaran}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		if got := battleResultEnemyText(lang, result); got == "" || strings.Contains(got, "battle.result") {
			t.Fatalf("lang=%v special enemy unresolved: %q", lang, got)
		}
	}
}
