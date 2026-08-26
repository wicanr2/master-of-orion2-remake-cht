package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestHiScoreSafeRectsStayInsideContentAndAwayFromContinue(t *testing.T) {
	if hsPanelX < 0 || hsPanelY < 0 || hsPanelX+hsPanelW > 640 || hsPanelY+hsPanelH > 480 {
		t.Fatalf("得分面板超出 640x480：(%v,%v,%v,%v)", hsPanelX, hsPanelY, hsPanelW, hsPanelH)
	}
	for _, r := range []textSafeRect{hiScoreTitleTextRect(), hiScoreSummaryTextRect()} {
		if r.x < 136 || r.y < 122 || r.x+r.w > 504 || r.y+r.h > 414 {
			t.Fatalf("標題／結果安全框超出面板：%+v", r)
		}
	}
	for i := 0; i < 32; i++ {
		isTotal := i == 31
		y, ok := hiScoreRowY(i, isTotal)
		if !ok {
			continue
		}
		label := hiScoreLabelTextRect(int(y), 18)
		value := hiScoreValueTextRect(int(y), 18)
		for _, r := range []textSafeRect{label, value} {
			if r.x < 136 || r.y < 122 || r.x+r.w > 504 || r.y+r.h > hsContinueY {
				t.Fatalf("分數列安全框撞面板／繼續按鈕：%+v", r)
			}
		}
	}
}

func TestHiScoreColumnsBoundLongBilingualLabelsAndValues(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	longLabel := strings.Repeat("殲滅種族與銀河議會勝利的長英文說明 ", 8)
	for _, size := range []float64{14} {
		r := hiScoreLabelTextRect(hsScoreTop, int(size)+4)
		label := r.clipped(fnt, longLabel, size)
		width, _ := fnt.Measure(label, size)
		if width > r.contentWidth() {
			t.Fatalf("標籤寬 %.1f 超出 %.1f：%q", width, r.contentWidth(), label)
		}
		valueRect := hiScoreValueTextRect(hsScoreTop, int(size)+4)
		value := valueRect.clipped(fnt, fmt.Sprintf("%d", 2147483647), size)
		valueWidth, _ := fnt.Measure(value, size)
		if valueWidth > valueRect.contentWidth() {
			t.Fatalf("數值寬 %.1f 超出 %.1f：%q", valueWidth, valueRect.contentWidth(), value)
		}
	}
}

func TestHiScoreRowsHaveExplicitOverflowBoundary(t *testing.T) {
	lastY, ok := hiScoreRowY(8, true)
	if !ok || lastY+24 > float64(hsScoreBottom) || lastY+24 > float64(hsContinueY) {
		t.Fatalf("總分列未在繼續按鈕前收束：y=%.0f bottom=%d continue=%d", lastY, hsScoreBottom, hsContinueY)
	}
	if _, ok := hiScoreRowY(9, false); ok {
		t.Fatalf("超量第 10 列仍被允許繪製")
	}
}
