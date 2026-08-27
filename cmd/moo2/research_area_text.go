package main

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
)

func researchAreaText(lang i18n.Lang, key string, args ...any) string {
	template := uiText(lang, key)
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}

func researchAreaTopicTextRect(hit hitRegion) textSafeRect {
	return textSafeRect{x: hit.x + 6, y: hit.y + 26, w: hit.w - 12, h: 28, insetX: 2, insetY: 2}
}
