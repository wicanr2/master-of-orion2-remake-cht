package main

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
)

func shipDesignText(lang i18n.Lang, key string, args ...any) string {
	template := uiText(lang, key)
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}

func shipDesignHullCostRect(index int) textSafeRect {
	// 原版可點艦體槽高度不一致，但 remake 的中文字形固定高 16px；相鄰價格欄
	// 使用六條連續 16px 安全列，中心與原槽中心誤差最多 1px，且不互相侵入。
	return textSafeRect{x: dsHullX1 + 8, y: 54 + index*16, w: 66, h: 16, insetX: 3}
}

func shipDesignComponentNameRect(row int) textSafeRect {
	return textSafeRect{x: 300, y: 54 + row*22, w: 166, h: 22, insetX: 5, insetY: 2}
}

func shipDesignComponentEffectRect(row int) textSafeRect {
	return textSafeRect{x: 466, y: 54 + row*22, w: 166, h: 22, insetX: 4, insetY: 2}
}

func shipDesignArcTextRect() textSafeRect {
	return textSafeRect{x: 300, y: 148, w: 332, h: 18, insetX: 5, insetY: 1}
}
func shipDesignAmmoTextRect() textSafeRect {
	return textSafeRect{x: 300, y: 167, w: 332, h: 18, insetX: 5, insetY: 1}
}
func shipDesignTotalTextRect() textSafeRect {
	return textSafeRect{x: 300, y: 214, w: 332, h: 22, insetX: 5, insetY: 2}
}
func shipDesignTreasuryTextRect() textSafeRect {
	return textSafeRect{x: 19, y: 441, w: 122, h: 25, insetX: 5, insetY: 2}
}
func shipDesignStatusTextRect() textSafeRect {
	return textSafeRect{x: 300, y: 238, w: 332, h: 36, insetX: 5, insetY: 2, lineH: 16}
}
func shipDesignSpaceHeaderRect() textSafeRect {
	return textSafeRect{x: 300, y: 277, w: 332, h: 18, insetX: 5, insetY: 1}
}

func shipDesignSpaceRowRect(index int) textSafeRect {
	return textSafeRect{x: 300 + (index/3)*166, y: 296 + (index%3)*18, w: 166, h: 18, insetX: 5, insetY: 1}
}

func shipDesignModHeaderRect() textSafeRect {
	return textSafeRect{x: 300, y: 350, w: 332, h: 16, insetX: 5}
}

func shipDesignModTextRect(index int) textSafeRect {
	r := designModChipRect(index)
	return textSafeRect{x: int(r.x), y: int(r.y), w: int(r.w), h: 16, insetX: 4}
}

func shipDesignControlTextRect(rect [4]int) textSafeRect {
	return textSafeRect{x: rect[0], y: rect[1], w: rect[2], h: rect[3], insetX: 2, insetY: 1}
}
