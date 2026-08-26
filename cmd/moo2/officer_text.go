package main

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
)

const officerDynamicFont = 9

func officerText(lang i18n.Lang, key string, args ...any) string {
	template := uiText(lang, key)
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}

func officerTargetTextRect() textSafeRect {
	return textSafeRect{x: 296, y: 36, w: 331, h: 18, insetX: 4, insetY: 1}
}

func officerMessageTextRect() textSafeRect {
	return textSafeRect{x: 296, y: 55, w: 331, h: 34, insetX: 4, insetY: 1, lineH: 16}
}

func officerNameTextRect(row int) textSafeRect {
	center := int(officerRowCenters()[row])
	return textSafeRect{x: 90, y: center - 33, w: 205, h: 24, insetX: 5, insetY: 2}
}

func officerSkillTextRect(row int) textSafeRect {
	center := int(officerRowCenters()[row])
	return textSafeRect{x: 90, y: center - 5, w: 205, h: 24, insetX: 5, insetY: 2}
}

func officerAssignmentTextRect(row int) textSafeRect {
	center := int(officerRowCenters()[row])
	return textSafeRect{x: 326, y: center + 1, w: 282, h: 22, insetX: 4, insetY: 2}
}

func officerEmptyTextRect() textSafeRect {
	return textSafeRect{x: 38, y: 55, w: 252, h: 62, insetX: 6, insetY: 3, lineH: 24}
}
