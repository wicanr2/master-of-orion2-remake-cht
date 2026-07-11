// Package uifont 用 ebiten text/v2 渲染文字,支援 CJK。
//
// 設計依 mom playbook(docs/kickoff/08 §2):
//   - [HARD] 字型定案前先確認 Go 的 opentype/sfnt 解析得動(某些 CFF/.ttc 會失敗)。
//     本套件的 Load/LoadCollection 若解析失敗會回 error,即為該檢查。
//   - text/v2 以向量字在目標像素尺寸直接 rasterize(已是清晰結果),取代 mom 手動
//     rasterize+supersample 的做法;MOO2 原生 640×480,文字以原尺寸繪製即銳利。
package uifont

import (
	"image/color"
	"math"

	bitmapfont "github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2"
	text "github.com/hajimehoshi/ebiten/v2/text/v2"
	xfont "golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

// Font 是一份已解析字型,依尺寸快取 face。
//
// 兩種模式:
//   - 向量模式(src != nil):opentype + text/v2,依請求 size 直接 rasterize(見上方套件註解)。
//   - 點陣模式(bm != nil):bitmapfont/v4 FaceTC 繁中點陣字,原生 bmPx(12px)銳利,
//     其他尺寸用整數最近鄰縮放(見 docs/tech/pixel-font-decision.md)。
type Font struct {
	src   *opentype.Font
	faces map[float64]text.Face

	bm   text.Face // 點陣模式:非 nil 時 Draw/DrawCentered/Measure 走點陣分支
	bmPx float64   // 點陣字原生 px(FaceTC = 12)
}

// LoadBitmapTC 建立以 bitmapfont FaceTC(繁中點陣,12px)為源的 Font。
// 點陣字內嵌於套件,無需外部字型檔。放大用整數最近鄰(保銳利,見 pixel-font-decision.md)。
func LoadBitmapTC() *Font {
	return &Font{bm: text.NewGoXFace(bitmapfont.FaceTC), bmPx: 12}
}

// bitmap 回報這個 Font 是否為點陣模式。
func (f *Font) bitmap() bool { return f.bm != nil }

// scaleFor 依請求 logical 字高回傳整數縮放倍率(點陣只在原生 px 銳利,故整數倍)。
func (f *Font) scaleFor(size float64) int {
	s := int(math.Round(size / f.bmPx))
	if s < 1 {
		s = 1
	}
	return s
}

// Load 解析單一 TTF/OTF 字型(opentype.Parse)。解析失敗即回 error(字型不相容檢查)。
func Load(data []byte) (*Font, error) {
	f, err := opentype.Parse(data)
	if err != nil {
		return nil, err
	}
	return &Font{src: f, faces: map[float64]text.Face{}}, nil
}

// LoadCollection 從 .ttc 字型集合取第 index 個字型(opentype.ParseCollection)。
func LoadCollection(data []byte, index int) (*Font, error) {
	coll, err := opentype.ParseCollection(data)
	if err != nil {
		return nil, err
	}
	f, err := coll.Font(index)
	if err != nil {
		return nil, err
	}
	return &Font{src: f, faces: map[float64]text.Face{}}, nil
}

// Face 取得(並快取)指定像素高度的 face。
func (f *Font) Face(size float64) text.Face {
	if fc, ok := f.faces[size]; ok {
		return fc
	}
	otf, err := opentype.NewFace(f.src, &opentype.FaceOptions{
		Size: size, DPI: 72, Hinting: xfont.HintingFull,
	})
	if err != nil {
		// NewFace 對已解析字型幾乎不會失敗;真失敗則回傳一個無法用的空 face 交由呼叫端察覺。
		return nil
	}
	fc := text.NewGoXFace(otf)
	f.faces[size] = fc
	return fc
}

// Draw 在 (x,y) 以左上為基準畫一段文字。
func (f *Font) Draw(dst *ebiten.Image, s string, x, y, size float64, c color.Color) {
	if f.bitmap() {
		sc := float64(f.scaleFor(size))
		op := &text.DrawOptions{}
		op.GeoM.Scale(sc, sc)
		op.GeoM.Translate(x, y)
		op.ColorScale.ScaleWithColor(c)
		op.Filter = ebiten.FilterNearest
		text.Draw(dst, s, f.bm, op)
		return
	}
	face := f.Face(size)
	if face == nil {
		return
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, face, op)
}

// DrawCentered 以 (cx,cy) 為中心水平+垂直置中畫一段文字(用 text/v2 對齊,免手算)。
func (f *Font) DrawCentered(dst *ebiten.Image, s string, cx, cy, size float64, c color.Color) {
	if f.bitmap() {
		sc := float64(f.scaleFor(size))
		op := &text.DrawOptions{}
		op.LayoutOptions.PrimaryAlign = text.AlignCenter
		op.LayoutOptions.SecondaryAlign = text.AlignCenter
		op.GeoM.Scale(sc, sc)
		op.GeoM.Translate(cx, cy)
		op.Filter = ebiten.FilterNearest
		op.ColorScale.ScaleWithColor(c)
		text.Draw(dst, s, f.bm, op)
		return
	}
	face := f.Face(size)
	if face == nil {
		return
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(cx, cy)
	op.LayoutOptions.PrimaryAlign = text.AlignCenter
	op.LayoutOptions.SecondaryAlign = text.AlignCenter
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, face, op)
}

// Measure 回傳字串在指定尺寸下的寬高(供置中/換行計算,對應 mom「量寬也要支援 CJK」)。
func (f *Font) Measure(s string, size float64) (w, h float64) {
	if f.bitmap() {
		sc := float64(f.scaleFor(size))
		w, h := text.Measure(s, f.bm, 0)
		return w * sc, h * sc
	}
	face := f.Face(size)
	if face == nil {
		return 0, 0
	}
	return text.Measure(s, face, 0)
}
