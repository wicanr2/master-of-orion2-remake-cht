package main

import (
	"image/color"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func councilTitleTextRect() textSafeRect {
	return textSafeRect{x: 40, y: 12, w: 560, h: 38, insetX: 4, insetY: 2}
}

func councilSummaryTextRect(y, h int) textSafeRect {
	return textSafeRect{x: 40, y: y, w: 560, h: h, insetX: 4, insetY: 2, lineH: 20}
}

func councilVoteTextRect(index int) textSafeRect {
	return textSafeRect{x: 80 + index*165, y: 370, w: 150, h: 40, insetX: 5, insetY: 3}
}

func councilDecisionTextRect(index int) textSafeRect {
	return textSafeRect{x: 120, y: 402 + index*30, w: 400, h: 26, insetX: 5, insetY: 2}
}

func councilRowTextRect(row int) textSafeRect {
	return textSafeRect{x: 70, y: 116 + row*24, w: 500, h: 22, insetX: 4, insetY: 1}
}

func councilCentered(fnt *uifont.Font, rect textSafeRect, text string, size float64, col color.RGBA) []extraText {
	return rect.centeredExtras(fnt, text, size, col)
}
