package main

import (
	"image/color"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestTextSafeRectLimitsWidthAndHeight(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	r := textSafeRect{x: 10, y: 20, w: 74, h: 25, insetX: 3, insetY: 1, lineH: 11}
	lines := r.lines(fnt, "這是一段很長的動態說明，不能跨進下一列按鈕", 10)
	if len(lines) != 2 {
		t.Fatalf("安全框最多兩行，得到 %d 行：%q", len(lines), lines)
	}
	for _, line := range lines {
		if w, _ := fnt.Measure(line, 10); w > r.contentWidth() {
			t.Fatalf("%q 寬 %.0f 超出安全寬 %.0f", line, w, r.contentWidth())
		}
	}
	if !strings.HasSuffix(lines[len(lines)-1], "…") {
		t.Fatalf("被高度截斷的最後一行應明示省略號，得到 %q", lines[len(lines)-1])
	}
}

func TestTextSafeRectCenterUsesItsOwnPanel(t *testing.T) {
	r := textSafeRect{x: 105, y: 209, w: 100, h: 20, insetX: 5, insetY: 2}
	if got, want := r.x+r.w/2, 155; got != want {
		t.Fatalf("文字中心 x=%d，want %d", got, want)
	}
	if got, want := r.y+r.h/2, 219; got != want {
		t.Fatalf("文字中心 y=%d，want %d", got, want)
	}
	if got := r.contentWidth(); got != 90 {
		t.Fatalf("內框寬 %.0f，want 90", got)
	}
}

func TestCenteredExtraTextUsesSafeInsetWithoutMovingPanelCenter(t *testing.T) {
	r := textSafeRect{x: 120, y: 126, w: 68, h: 20, insetX: 3, insetY: 2}
	e := centeredExtraTextInSafeRect(r, 10, "增派 0", color.RGBA{})
	if e.x != 154 || e.y != 136 {
		t.Fatalf("按鈕文字中心 = (%.0f,%.0f)，want (154,136)", e.x, e.y)
	}
	if e.maxW != 62 {
		t.Fatalf("按鈕文字安全欄寬 = %.0f，want 62", e.maxW)
	}
}
