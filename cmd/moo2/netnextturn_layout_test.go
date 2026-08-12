package main

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/netplay"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestNetNextTurnTextSafeRectsStayInsideNativePanels(t *testing.T) {
	inside := func(name string, r textSafeRect, x, y, w, h int) {
		t.Helper()
		if r.x < x || r.y < y || r.x+r.w > x+w || r.y+r.h > y+h {
			t.Fatalf("%s 安全框 (%d,%d,%d,%d) 超出面板 (%d,%d,%d,%d)",
				name, r.x, r.y, r.w, r.h, x, y, w, h)
		}
	}

	inside("標題", nntBannerTitleTextRect(), nntX, nntBannerY, nntPanelW, 48)
	for row := 0; ; row++ {
		y := nntRowFirst + row*nntRowStep
		if y > nntBotY-nntRowStep {
			break
		}
		inside("玩家名稱", nntPlayerNameTextRect(row), nntX, nntMidY, nntPanelW, 179)
		inside("玩家狀態", nntPlayerStatusTextRect(row), nntX, nntMidY, nntPanelW, 179)
	}
	inside("回合", nntTurnTextRect(), nntX, nntMidY, nntPanelW, 179)
	inside("指紋", nntFingerprintTextRect(), nntX, nntMidY, nntPanelW, 179)
	for row := 0; row < netplay.ChatLogMax; row++ {
		r := nntChatTextRect(row)
		inside("聊天", r, nntX, nntBotY, nntPanelW, 221)
		if r.y+r.h > nntInputY {
			t.Fatalf("第 %d 行聊天底部 %d 壓到輸入列 %d", row, r.y+r.h, nntInputY)
		}
	}
	inside("分岔標題", nntDesyncTitleTextRect(), nntX+16, nntBotY+137, 598, 44)
	inside("分岔詳情", nntDesyncDetailTextRect(), nntX+16, nntBotY+137, 598, 44)
	inside("聊天輸入", nntInputTextRect(), nntX+16, nntInputY, 598, nntInputH)
}

func TestNetNextTurnTextSafeRectsClipNamesChatAndInput(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	cases := []struct {
		name string
		r    textSafeRect
		text string
		size float64
	}{
		{"玩家名稱", nntPlayerNameTextRect(0), strings.Repeat("VeryWideName", 12), 13},
		{"玩家狀態", nntPlayerStatusTextRect(0), strings.Repeat("等待其他玩家完成同步 ", 10), 13},
		{"聊天", nntChatTextRect(0), strings.Repeat("W", netplay.ChatTextMax) + " (長前綴)", 11},
		{"輸入列", nntInputTextRect(), strings.Repeat("正在輸入非常長的多人對話 ", 14), 11},
		{"分岔", nntDesyncDetailTextRect(), strings.Repeat("0123456789abcdef", 12), 11},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lines := tc.r.lines(fnt, tc.text, tc.size)
			if len(lines) != 1 {
				t.Fatalf("單行安全框得到 %d 行", len(lines))
			}
			if width, _ := fnt.Measure(lines[0], tc.size); width > tc.r.contentWidth() {
				t.Fatalf("%q 寬 %.0f 超出 %.0f", lines[0], width, tc.r.contentWidth())
			}
		})
	}
}
