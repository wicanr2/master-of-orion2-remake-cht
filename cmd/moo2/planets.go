package main

import (
	"image/color"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// 行星列表畫面標籤覆蓋。座標多取自 openorion2 galaxy.cpp PlanetsListView::initWidgets;
// 欄位標題(烘在 header 列 18,15,397,16)與兩個區塊標題為量測估值。
var planetsOverlays = []labelRect{
	// 欄位標題列
	{18, 14, 78, 18, "Planet", 0},
	{96, 14, 80, 18, "Climate", 0},
	{178, 14, 80, 18, "Gravity", 0},
	{258, 14, 90, 18, "Minerals", 0},
	{350, 14, 64, 18, "Size", 0},
	// 排序區(區塊標題依實際可見位置,非 widget 點擊區;加高以完全蓋住英文)
	{435, 178, 196, 22, "Sort Priority", 0},
	{441, 200, 60, 24, "Climate", 0},
	{501, 200, 66, 24, "Minerals", 0},
	{567, 200, 57, 24, "Size", 0},
	// 篩選區
	{430, 242, 202, 22, "Display Restrictions", 13},
	{441, 266, 183, 21, "No Enemy Presence", 0},
	{441, 289, 183, 22, "Normal Gravity", 0},
	{441, 312, 183, 22, "Non-Hostile Environment", 0},
	{441, 335, 183, 22, "Mineral Abundance", 0},
	{441, 358, 183, 20, "Planets In Range", 0},
	// 動作按鈕
	{454, 386, 156, 23, "Send Colony Ship", 0},
	{454, 413, 157, 25, "Send Outpost Ship", 0},
	{454, 440, 157, 23, "Return", 0},
}

// runPlanets 渲染行星列表畫面(中/英)。
func runPlanets(dirs []string, lang i18n.Lang, fnt *uifont.Font, tsvPath, shot string, frames int) error {
	return runOverlay(dirs, "plntsum.lbx", 0, lang, fnt, tsvPath, planetsOverlays,
		color.RGBA{206, 218, 240, 255}, 14,
		"Master of Orion II — 行星列表 (cht)", shot, frames)
}
