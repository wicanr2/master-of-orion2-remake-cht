package main

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
)

const (
	menuToggleHitX = 12
	menuToggleHitW = 220
	menuToggleHitH = 22
	menuLanguageY  = 428
	menuVersionY   = 450
)

func mainMenuText(lang i18n.Lang, key string, args ...any) string {
	template := uiText(lang, key)
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}

func mainMenuToggleOwnerRect(y int) textSafeRect {
	return textSafeRect{x: menuToggleHitX, y: y, w: menuToggleHitW, h: menuToggleHitH}
}

func mainMenuToggleTextRect(y int) textSafeRect {
	return textSafeRect{x: menuToggleHitX, y: y, w: menuToggleHitW, h: menuToggleHitH, insetX: 8, insetY: 3}
}
