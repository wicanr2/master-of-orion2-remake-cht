package main

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// TestTruncateToWidth 守住殖民地「已建:…」建築清單溢出「建造」欄 cell 的修正
// (使用者 2026-07-11 回饋:點陣字 12px 下限把小字撐大、長清單撞出畫面右緣)。
func TestTruncateToWidth(t *testing.T) {
	fnt := uifont.LoadBitmapTC() // 純點陣 12px,無需外部字型檔
	full := "已建:星基、海軍陸戰隊營"
	if w, _ := fnt.Measure(full, 10); w <= 110 {
		t.Fatalf("前提不成立:原字串量寬 %.0f 應 >110(建造欄寬)", w)
	}
	got := truncateToWidth(fnt, full, 10, 110)
	if w, _ := fnt.Measure(got, 10); w > 110 {
		t.Fatalf("截斷後仍超出欄寬:%.0f > 110(%q)", w, got)
	}
	if !strings.HasSuffix(got, "…") {
		t.Fatalf("截斷後應以省略號結尾,got %q", got)
	}
	if short := "已建:星基"; truncateToWidth(fnt, short, 10, 110) != short {
		t.Fatalf("未超寬的短字串不應被截斷")
	}
}
