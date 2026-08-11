package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// hover.go：共用的滑鼠 hover 外框。原版按鈕的浮雕底圖仍保留，hover 只加一層
// 明亮外框，避免覆蓋原版美術或把按鈕重畫成另一套風格。

func pointInRect(mx, my, x, y, w, h int) bool {
	return mx >= x && mx < x+w && my >= y && my < y+h
}

func drawHoverBorder(dst *ebiten.Image, x, y, w, h float32, hovered bool) {
	if !hovered {
		return
	}
	vector.StrokeRect(dst, x-2, y-2, w+4, h+4, 2, color.RGBA{255, 232, 120, 255}, false)
}
