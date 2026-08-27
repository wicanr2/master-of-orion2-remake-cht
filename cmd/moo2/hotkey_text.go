package main

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
)

func hotkeyText(lang i18n.Lang, key string, args ...any) string {
	template := uiText(lang, key)
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}

func clampTextRectToStarViewport(r textSafeRect) textSafeRect {
	if r.x < starVX0 {
		r.x = starVX0
	}
	if r.y < starVY0 {
		r.y = starVY0
	}
	if r.x+r.w > starVX1+1 {
		r.x = starVX1 + 1 - r.w
	}
	if r.y+r.h > starVY1+1 {
		r.y = starVY1 + 1 - r.h
	}
	return r
}

func measureHintTextRect(mx, my int) textSafeRect {
	const width, height, gap = 210, 22, 12
	x, y := mx+gap, my-height-gap/2
	if x+width > starVX1+1 {
		x = mx - gap - width
	}
	if y < starVY0 {
		y = my + gap
	}
	return clampTextRectToStarViewport(textSafeRect{x: x, y: y, w: width, h: height, insetX: 5, insetY: 2})
}

func measureDistanceTextRect(fx, fy, tx, ty int) textSafeRect {
	const width, height = 96, 20
	cx, cy := (fx+tx)/2, (fy+ty)/2-8
	return clampTextRectToStarViewport(textSafeRect{
		x: cx - width/2, y: cy - height/2, w: width, h: height, insetX: 4, insetY: 2,
	})
}

func quickSaveFlashTextRect() textSafeRect {
	return textSafeRect{x: 240, y: 396, w: starVX1 + 1 - 240, h: 22, insetX: 6, insetY: 2}
}
