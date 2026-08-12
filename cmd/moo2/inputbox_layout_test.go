package main

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestInputBoxTextSafeRectMatchesNativeGeometry(t *testing.T) {
	fx, fy, fw, fh := inboxFieldRect(inboxDefaultX, inboxDefaultY)
	r := inboxInputTextRect(fx, fy, fw, fh)
	if r.x < fx || r.y < fy || r.x+r.w > fx+fw || r.y+r.h > fy+fh {
		t.Fatalf("輸入文字安全框 %+v 超出輸入欄 (%d,%d,%d,%d)", r, fx, fy, fw, fh)
	}
	if r.contentX() != fx+6 || r.contentWidth() != float64(fw-18) {
		t.Fatalf("輸入文字安全欄未保留既有內縮：x=%d width=%.0f", r.contentX(), r.contentWidth())
	}
	bx, by, bw, bh := inboxOKRect(inboxDefaultX, inboxDefaultY)
	if by+bh > inboxDefaultY+inboxBoxH || bx+bw > inboxDefaultX+inboxBoxW {
		t.Fatalf("OK 按鈕超出輸入框：(%d,%d,%d,%d)", bx, by, bw, bh)
	}
}

func TestInputBoxCaretReservesMeasuredWidthForLongBilingualText(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	fx, fy, fw, fh := inboxFieldRect(inboxDefaultX, inboxDefaultY)
	r := inboxInputTextRect(fx, fy, fw, fh)
	for _, text := range []string{
		strings.Repeat("這是一段很長的繁體中文名稱", 12),
		strings.Repeat("A very long English multiplayer game name ", 12),
	} {
		visible := inboxVisibleInputText(fnt, text, 14, r)
		width, _ := fnt.Measure(visible, 14)
		if width > r.contentWidth() {
			t.Fatalf("輸入內容含游標後寬 %.1f 超出安全寬 %.1f：%q", width, r.contentWidth(), visible)
		}
		if !strings.HasSuffix(visible, "_") {
			t.Fatalf("截斷後遺失游標：%q", visible)
		}
	}
}
