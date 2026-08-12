package main

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestConfirmMessageTextSafeRectStaysAboveButtons(t *testing.T) {
	r := confirmMessageTextRect()
	if r.x < confirmBoxX || r.y < confirmBoxY || r.x+r.w > confirmBoxX+confirmBoxW || r.y+r.h > confirmBoxY+confirmBoxH {
		t.Fatalf("確認訊息框 (%d,%d,%d,%d) 超出確認框", r.x, r.y, r.w, r.h)
	}
	if r.y+r.h >= confirmYesY {
		t.Fatalf("確認訊息底部 %d 壓到按鈕 y=%d", r.y+r.h, confirmYesY)
	}
}

func TestConfirmMessageTextHasFixedHeightOverflowPolicy(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	text := strings.Repeat("這是一段很長的確認說明，英文版也可能比原始按鈕文字長很多。", 12)
	r := confirmMessageTextRect()
	lines := r.lines(fnt, text, confirmTextSize)
	if len(lines) != r.maxLines() {
		t.Fatalf("行數 = %d，預期剛好收束到 %d", len(lines), r.maxLines())
	}
	if !strings.HasSuffix(lines[len(lines)-1], "…") {
		t.Fatalf("最後一行沒有省略號：%q", lines[len(lines)-1])
	}
	for _, line := range lines {
		if width, _ := fnt.Measure(line, confirmTextSize); width > r.contentWidth() {
			t.Fatalf("%q 寬 %.0f 超出 %.0f", line, width, r.contentWidth())
		}
	}
	extras := r.centeredExtras(fnt, text, confirmTextSize, infoBodyCol)
	for _, e := range extras {
		if e.align != 1 || e.x != float64(r.x)+float64(r.w)/2 || e.y < float64(r.y) || e.y > float64(r.y+r.h) {
			t.Fatalf("置中行不在確認安全框內：%+v", e)
		}
	}
}
