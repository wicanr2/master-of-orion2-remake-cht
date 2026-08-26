package main

import (
	"fmt"
	"image/color"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func battleResultEnemyText(lang i18n.Lang, result *shell.BattleResult) string {
	if result != nil && result.EnemyKind == shell.BattleEnemyAntaran {
		return uiText(lang, "battle.result.enemy.antaran_home_defense")
	}
	if result == nil || result.Enemy == "" {
		return uiText(lang, "battle.result.enemy.unknown")
	}
	return result.Enemy
}

func battleResultTextRect(row int) textSafeRect {
	return textSafeRect{x: 40, y: 44 + row*26, w: 560, h: 24, insetX: 3, insetY: 1}
}

func battleResultLogTextRect(row int) textSafeRect {
	return textSafeRect{x: 40, y: 122 + row*24, w: 560, h: 22, insetX: 3, insetY: 1}
}

func battleResultLossTextRect(logCount int) textSafeRect {
	y := 126 + min(logCount, 6)*24
	return textSafeRect{x: 40, y: y, w: 560, h: 26, insetX: 3, insetY: 1}
}

func battleResultExtras(fnt *uifont.Font, rect textSafeRect, text string, size float64, col color.RGBA) []extraText {
	return rect.leftExtras(fnt, text, size, col)
}

func battleRoundText(lang i18n.Lang, round shell.BattleRoundResult) string {
	return fmt.Sprintf(uiText(lang, "battle.result.round"), round.Round, round.EnemyDestroyed, round.PlayerDestroyed)
}
