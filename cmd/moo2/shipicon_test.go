package main

// shipicon_test.go:星圖艦隊圖示 + 旗色索引的護欄。
//
// 這兩者綁在一起:艦隊圖示的資產編號是 `205 + 旗色×4`,所以**旗色的順序本身就是資料**。
// 順序錯了不會有任何錯誤訊息——顏色名還是對的,只是畫面上的小圖換成別人的顏色。

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// TestShipIconAssetMatchesOriginal:資產 = 205 + 旗色×4 + 縮放(`Get_Ship_Icon_Pict_Seg_`)。
// 順帶釘住「旗色只有 8 種、資產落在 205..236」——多算一格會踩到 237 起的中立/怪獸那組。
func TestShipIconAssetMatchesOriginal(t *testing.T) {
	if len(shell.FlagColors) != 8 {
		t.Fatalf("旗色應為 8 種(原版 BUFFER0 每 4 張一組共 32 張),實得 %d", len(shell.FlagColors))
	}
	for i := range shell.FlagColors {
		want := 205 + i*4
		if got := shipIconAsset(i); got != want {
			t.Errorf("旗色 %d 的資產 = %d,want %d", i, got, want)
		}
		if got := shipIconAsset(i); got < 205 || got > 236 {
			t.Errorf("旗色 %d 的資產 %d 超出帝國那組 205..236", i, got)
		}
	}
	// 越界夾回第 0 色:上游有 bug 時畫面不該整個不見。
	for _, bad := range []int{-1, 8, 99} {
		if got := shipIconAsset(bad); got != 205 {
			t.Errorf("旗色 %d(越界)應夾回 205,實得 %d", bad, got)
		}
	}
}

// TestFlagColorsMatchOriginalOrder:旗色順序**就是原版的旗色索引**,不能重排。
//
// 排錯了選紅色會開出白色的艦隊(艦隊圖示 = 205 + 旗色×4),而且中文模式看不出來——
// 顏色名還是對的,只有畫面上的小圖不對。2026-08-07 修過一次:先前是
// 紅/黃/綠/**藍/白/紫/橙/棕**,後五個全錯位。
//
// 兩個獨立來源:① BUFFER0.LBX 205/209/…/233 渲染出來量代表色;
// ② openorion2 `src/gfx.h` 的 `FONT_COLOR_PLAYER_*`(RED/YELLOW/GREEN/SILVER/BLUE/BROWN/PURPLE/ORANGE)。
func TestFlagColorsMatchOriginalOrder(t *testing.T) {
	want := []string{"Red", "Yellow", "Green", "Silver", "Blue", "Brown", "Purple", "Orange"}
	if len(shell.FlagColors) != len(want) {
		t.Fatalf("旗色數 %d,want %d", len(shell.FlagColors), len(want))
	}
	for i, w := range want {
		if got := shell.FlagColors[i].EnName; got != w {
			t.Errorf("第 %d 色 = %q,want %q(原版旗色索引順序)", i, got, w)
		}
	}
	// 量到的代表色:銀是灰的(R≈G≈B)、藍偏亮。這兩條擋的是「名字對了但 RGB 抄舊值」。
	if s := shell.FlagColors[3]; !(abs8(s.R, s.G) < 20 && abs8(s.G, s.B) < 20) {
		t.Errorf("第 3 色(銀)RGB (%d,%d,%d) 不像灰色", s.R, s.G, s.B)
	}
	if bl := shell.FlagColors[4]; !(bl.B > bl.R && bl.B > bl.G) {
		t.Errorf("第 4 色(藍)RGB (%d,%d,%d) 的藍不是最強", bl.R, bl.G, bl.B)
	}
}

func abs8(a, b uint8) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}
