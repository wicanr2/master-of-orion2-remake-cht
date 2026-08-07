package lbx

import (
	"image/color"
	"testing"
)

// image_accum_test.go:delta 幀累積的護欄。
//
// 這一條是「面板整張消失」那個 bug 的直接測試:逐幀獨立上色 vs 累積上色,
// 對 delta 動畫會得到完全不同的結果,而錯的那個在畫面上是「什麼都沒有」。

// 造一張 2×1 的兩幀動畫:第 0 幀寫滿,第 1 幀只改右邊那顆。
func deltaImage() *Image {
	f0 := &Frame{W: 2, H: 1, Index: []uint8{1, 2}, Written: []bool{true, true}}
	f1 := &Frame{W: 2, H: 1, Index: []uint8{0, 3}, Written: []bool{false, true}}
	return &Image{Width: 2, Height: 1, Frames: []*Frame{f0, f1}}
}

func TestAccumulatedUpToKeepsPixelsTheLaterFrameDidNotTouch(t *testing.T) {
	pal := &Palette{}
	pal[1] = color.RGBA{R: 10, G: 10, B: 10, A: 255}
	pal[2] = color.RGBA{R: 20, G: 20, B: 20, A: 255}
	pal[3] = color.RGBA{R: 30, G: 30, B: 30, A: 255}
	im := deltaImage()

	// 直接上色第 1 幀:左邊那顆是透明的(第 1 幀沒寫)——這正是 bug 的樣子。
	solo := im.Frames[1].ToRGBA(pal, false)
	if solo.Pix[3] != 0 {
		t.Fatal("測試前提不成立:第 1 幀的左像素本來就該是透明的(delta 幀)")
	}

	// 累積到第 1 幀:左邊保留第 0 幀的值,右邊換成第 1 幀的。
	acc := im.AccumulatedUpToRGBA(pal, 1, false)
	if acc.Pix[3] == 0 {
		t.Error("累積後左像素仍是透明——delta 幀沒有疊上去")
	}
	if acc.Pix[0] != 10 {
		t.Errorf("左像素應保留第 0 幀的顏色 10,實得 %d", acc.Pix[0])
	}
	if acc.Pix[4] != 30 {
		t.Errorf("右像素應換成第 1 幀的顏色 30,實得 %d", acc.Pix[4])
	}
}

// n 夾在有效區間內,而且 n=0 等同直接上色第 0 幀。
func TestAccumulatedUpToClampsAndMatchesFrameZero(t *testing.T) {
	pal := &Palette{}
	pal[1] = color.RGBA{R: 10, A: 255}
	pal[2] = color.RGBA{R: 20, A: 255}
	pal[3] = color.RGBA{R: 30, A: 255}
	im := deltaImage()

	zero := im.AccumulatedUpToRGBA(pal, 0, false)
	solo := im.Frames[0].ToRGBA(pal, false)
	for i := range solo.Pix {
		if zero.Pix[i] != solo.Pix[i] {
			t.Fatalf("n=0 應等同直接上色第 0 幀,第 %d 個位元組不同", i)
		}
	}
	last := im.AccumulatedUpToRGBA(pal, 99, false)
	one := im.AccumulatedUpToRGBA(pal, 1, false)
	for i := range one.Pix {
		if last.Pix[i] != one.Pix[i] {
			t.Fatalf("n 超出範圍應夾到最後一幀,第 %d 個位元組不同", i)
		}
	}
}

// keyColor 時 index 0 要透明——與 Frame.ToRGBA 同語意,否則兩條路徑會畫出不同的邊緣。
func TestAccumulatedUpToHonoursKeyColor(t *testing.T) {
	pal := &Palette{}
	pal[0] = color.RGBA{R: 99, G: 99, B: 99, A: 255}
	pal[1] = color.RGBA{R: 10, A: 255}
	f0 := &Frame{W: 2, H: 1, Index: []uint8{0, 1}, Written: []bool{true, true}}
	im := &Image{Width: 2, Height: 1, Frames: []*Frame{f0}}
	acc := im.AccumulatedUpToRGBA(pal, 0, true)
	if acc.Pix[3] != 0 {
		t.Error("keyColor 時 index 0 應為透明")
	}
	if acc.Pix[7] == 0 {
		t.Error("非 index 0 的像素不該被 keyColor 影響")
	}
}
