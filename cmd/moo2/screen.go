package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// screen.go:全 cmd/moo2 共用的畫面介面與切換指令。
//
// ⚠ 這兩個型別原本住在 `play.go`(2026-08-08 第 81 項(淘汰簡約殼)刪掉的那個簡約殼)。
// 它們不是那個殼的東西——`interactive.go` 與十來個畫面檔都在用,只是當初先寫在那裡。
// 拆出來之後 `play.go` 才刪得乾淨。

// transition 是畫面切換指令:切到 next,或 quit。
type transition struct {
	next screen
	quit bool
}

// screen 是一個可互動畫面。
type screen interface {
	update(in shell.InputState) *transition
	draw(dst *ebiten.Image, font *uifont.Font)
}
