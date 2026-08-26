package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// textSafeRect 是動態文字唯一可使用的邏輯安全框。它和美術／熱區的外框分離，
// 但由外框量得；因此不會因為某一段譯文變長就跨進相鄰欄、下一列或按鈕。
//
// insetX/Y 是保留給邊框的內縮，lineH 是多行文字的固定行距。單行文字可省略
// lineH；多行框的高度不足一行時寧可不畫，也不能讓字越界。
type textSafeRect struct {
	x, y, w, h     int
	insetX, insetY int
	lineH          int
}

func (r textSafeRect) contentWidth() float64 {
	w := r.w - 2*r.insetX
	if w < 0 {
		return 0
	}
	return float64(w)
}

func (r textSafeRect) contentX() int { return r.x + r.insetX }
func (r textSafeRect) contentY() int { return r.y + r.insetY }

func (r textSafeRect) maxLines() int {
	h := r.h - 2*r.insetY
	if r.w <= 0 || h <= 0 {
		return 0
	}
	if r.lineH <= 0 {
		return 1
	}
	return h / r.lineH
}

func (r textSafeRect) clipped(fnt *uifont.Font, text string, size float64) string {
	if r.contentWidth() <= 0 {
		return ""
	}
	return truncateToWidth(fnt, text, size, r.contentWidth())
}

// lines 先依安全欄寬折行，再依安全高度截掉多出的行。最後一行以省略號明示內容
// 被收束，而不是靜默壓到下一個控制項。
func (r textSafeRect) lines(fnt *uifont.Font, text string, size float64) []string {
	maxLines := r.maxLines()
	if maxLines == 0 || r.contentWidth() <= 0 || text == "" {
		return nil
	}
	lines := wrapToWidth(fnt, text, size, r.contentWidth())
	if len(lines) > maxLines {
		lines = append([]string(nil), lines[:maxLines]...)
		lines[maxLines-1] = r.clipped(fnt, lines[maxLines-1]+"…", size)
	}
	for i := range lines {
		lines[i] = r.clipped(fnt, lines[i], size)
	}
	return lines
}

// leftExtras 回傳已受雙軸限制的 overlay 動態文字。不要在 postDraw 裡自行組
// extraText，否則只有限寬、沒有行數與高度的舊問題會再次出現。
func (r textSafeRect) leftExtras(fnt *uifont.Font, text string, size float64, col color.RGBA) []extraText {
	if fnt == nil {
		return nil
	}
	lines := r.lines(fnt, text, size)
	out := make([]extraText, 0, len(lines))
	lineH := r.lineH
	if lineH <= 0 {
		lineH = r.h - 2*r.insetY
	}
	for i, line := range lines {
		out = append(out, extraText{
			x: float64(r.contentX()), y: float64(r.contentY() + i*lineH),
			size: size, text: line, col: col, maxW: r.contentWidth(),
		})
	}
	return out
}

// centeredExtras 是多行置中動態文字的安全版本。每一行都先由 lines 依相同的欄寬與
// 高度政策收束，再以整塊內容區的中心等距排列；不可先 Wrap 後自行用 y+i*lineH 直畫。
func (r textSafeRect) centeredExtras(fnt *uifont.Font, text string, size float64, col color.RGBA) []extraText {
	if fnt == nil {
		return nil
	}
	lines := r.lines(fnt, text, size)
	if len(lines) == 0 {
		return nil
	}
	lineH := r.lineH
	if lineH <= 0 {
		lineH = r.h - 2*r.insetY
	}
	contentH := r.h - 2*r.insetY
	cy := float64(r.contentY()) + float64(contentH)/2
	firstY := cy - float64(len(lines)-1)*float64(lineH)/2
	out := make([]extraText, 0, len(lines))
	for i, line := range lines {
		out = append(out, extraText{
			x: float64(r.x) + float64(r.w)/2, y: firstY + float64(i*lineH),
			size: size, text: line, col: col, align: 1, maxW: r.contentWidth(),
		})
	}
	return out
}

func (r textSafeRect) drawLeft(dst *ebiten.Image, fnt *uifont.Font, text string, size float64, col color.Color) {
	if fnt == nil {
		return
	}
	lineH := r.lineH
	if lineH <= 0 {
		lineH = r.h - 2*r.insetY
	}
	for i, line := range r.lines(fnt, text, size) {
		fnt.Draw(dst, line, float64(r.contentX()), float64(r.contentY()+i*lineH), size, col)
	}
}

func (r textSafeRect) drawRight(dst *ebiten.Image, fnt *uifont.Font, text string, size float64, col color.Color) {
	if fnt == nil || r.maxLines() == 0 {
		return
	}
	text = r.clipped(fnt, text, size)
	w, _ := fnt.Measure(text, size)
	fnt.Draw(dst, text, float64(r.x+r.w-r.insetX)-w, float64(r.contentY()), size, col)
}

func (r textSafeRect) drawCentered(dst *ebiten.Image, fnt *uifont.Font, text string, size float64, col color.Color) {
	if fnt == nil || r.maxLines() == 0 {
		return
	}
	fnt.DrawCentered(dst, r.clipped(fnt, text, size),
		float64(r.x)+float64(r.w)/2, float64(r.y)+float64(r.h)/2, size, col)
}

// drawCenteredLines 是給自繪對話框使用的多行置中版本。它和 overlay 的 extras 共用
// 完全相同的量測／截斷規則，避免 modal 畫面重建一套沒有高度上限的折行流程。
func (r textSafeRect) drawCenteredLines(dst *ebiten.Image, fnt *uifont.Font, text string, size float64, col color.Color) {
	for _, e := range r.centeredExtras(fnt, text, size, color.RGBAModel.Convert(col).(color.RGBA)) {
		fnt.DrawCentered(dst, e.text, e.x, e.y, e.size, e.col)
	}
}

// centeredExtraTextInSafeRect 是 overlay 動態按鈕的同款中心計算。外框／熱區與
// 文字安全框可以有不同內縮，但兩者的整數像素中心必須一致。
func centeredExtraTextInSafeRect(r textSafeRect, size float64, text string, col color.RGBA) extraText {
	return extraText{
		x: float64(r.x) + float64(r.w)/2, y: float64(r.y) + float64(r.h)/2,
		size: size, text: text, col: col, align: 1, maxW: r.contentWidth(),
	}
}
