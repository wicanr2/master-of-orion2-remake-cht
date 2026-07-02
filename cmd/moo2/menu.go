package main

import (
	"image/color"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// 主選單六按鈕:座標取自 openorion2 mainmenu.cpp(x=415, w=153),英文烘在背景圖。
var menuOverlays = []labelRect{
	{415, 172, 153, 23, "Continue", 0},
	{415, 195, 153, 22, "Load Game", 0},
	{415, 217, 153, 23, "New Game", 0},
	{415, 240, 153, 22, "Multi Player", 0},
	{415, 262, 153, 23, "Hall of Fame", 0},
	{415, 285, 153, 22, "Quit Game", 0},
}

// runMenu 渲染主選單(中/英)。
func runMenu(dirs []string, lang i18n.Lang, fnt *uifont.Font, tsvPath, shot string, frames int) error {
	return runOverlay(dirs, "mainmenu.lbx", 21, lang, fnt, tsvPath, menuOverlays,
		color.RGBA{104, 224, 96, 255}, 15,
		"Master of Orion II — 主選單 (cht)", shot, frames)
}
