package main

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestInfoTextSafeRectsStayInsideNativePanel(t *testing.T) {
	x, y, w, h := infoPanelBounds()
	inside := func(name string, r textSafeRect) {
		t.Helper()
		if r.x < x || r.y < y || r.x+r.w > x+w || r.y+r.h > y+h {
			t.Fatalf("%s 安全框 (%d,%d,%d,%d) 超出 INFO 面板 (%d,%d,%d,%d)",
				name, r.x, r.y, r.w, r.h, x, y, w, h)
		}
	}

	inside("分頁標題", infoTitleTextRect())
	inside("分頁說明", infoCenteredTextRect(y+48, 24))
	inside("科技摘要", infoContentTextRect(y+44, 14))
	for column := 0; column < 2; column++ {
		inside("科技主題欄", infoTechTopicTextRect(column, y+88, 12))
	}
	for column := 0; column < 5; column++ {
		inside("種族統計欄", infoRaceStatTextRect(column, y+44, 12))
	}
	inside("種族關係", infoRaceRelationTextRect(y+236))
	inside("回合摘要標籤", infoSummaryLabelTextRect(y+44))
	inside("回合摘要值", infoSummaryValueTextRect(y+44))
	for column := 0; column < 2; column++ {
		inside("參考資料欄", infoReferenceTextRect(column, y+64))
	}
	inside("圖表最大值", infoHistoryMaxTextRect())
	inside("圖表零點", infoHistoryZeroTextRect())
	inside("圖表起始回合", infoHistoryTurnTextRect(false))
	inside("圖表結束回合", infoHistoryTurnTextRect(true))
	for i := 0; i < infoHistoryLegendSlots; i++ {
		inside("圖例", infoHistoryLegendTextRect(i))
	}
}

func TestInfoTextSafeRectsBoundLongBilingualData(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	cases := []struct {
		name string
		r    textSafeRect
		text string
		size float64
	}{
		{"英文分頁標題", infoTitleTextRect(), "RELATIONS BETWEEN ALL KNOWN EMPIRES AND THEIR CURRENT TREATIES", 15},
		{"研究主題", infoTechTopicTextRect(1, 132, 12), "✓ " + strings.Repeat("超長科技名稱 ", 16), 10},
		{"帝國名", infoRaceStatTextRect(0, 108, 12), strings.Repeat("Very Long Custom Empire ", 8), 11},
		{"關係摘要", infoRaceRelationTextRect(280), strings.Repeat("[Bulrathi alliance] ", 24), 10},
		{"事件摘要", infoTextRect(int(infoPanelX)+24, 305, int(infoPanelW)-48, 12), strings.Repeat("重大事件與特殊貿易狀態 ", 18), 10},
		{"參考資料", infoReferenceTextRect(1, 108), strings.Repeat("A long reference entry ", 16), 10},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lines := tc.r.lines(fnt, tc.text, tc.size)
			if len(lines) == 0 || len(lines) > tc.r.maxLines() {
				t.Fatalf("行數 = %d，安全框上限 = %d", len(lines), tc.r.maxLines())
			}
			for _, line := range lines {
				if width, _ := fnt.Measure(line, tc.size); width > tc.r.contentWidth() {
					t.Fatalf("%q 寬 %.0f 超出 %.0f", line, width, tc.r.contentWidth())
				}
			}
		})
	}
}
