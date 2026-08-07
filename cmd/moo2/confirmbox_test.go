package main

import "testing"

// confirmbox_test.go:確認框的版面護欄。
//
// 這些數字全部來自 `Confirmation_Box_` @ 0x77658 / `Draw_Confirm_Box_` @ 0x778E4 的立即數
// (見 confirmbox.go 檔頭)。釘住它們的理由:版面常數最容易在「順手調一下比較好看」時被改掉,
// 而改掉之後畫面看起來仍然「正常」——只是不再是原版的位置。

// 兩顆鈕的貼圖點與熱區:sub_12B7E1 貼在 (235,302)/(345,302);
// sub_11438B 的熱區是 235..286 / 345..396 × 302..323。
func TestConfirmButtonGeometryMatchesDisassembly(t *testing.T) {
	cases := []struct {
		name                 string
		x, y, wantX2, wantY2 int
		gotW, gotH           int
	}{
		{"Y", confirmYesX, confirmYesY, 286, 323, confirmBtnW, confirmBtnH},
		{"N", confirmNoX, confirmNoY, 396, 323, confirmBtnW, confirmBtnH},
	}
	for _, c := range cases {
		if got := c.x + c.gotW; got != c.wantX2 {
			t.Errorf("%s 鈕右緣應為 %d(反組譯 ebx),實得 %d", c.name, c.wantX2, got)
		}
		if got := c.y + c.gotH; got != c.wantY2 {
			t.Errorf("%s 鈕下緣應為 %d(反組譯 ecx),實得 %d", c.name, c.wantY2, got)
		}
	}
	if confirmYesY != confirmNoY {
		t.Errorf("兩顆鈕在原版是同一列(edx 都是 12Eh),實得 %d / %d", confirmYesY, confirmNoY)
	}
}

// 文字塊的中心要落在底框中心上:框在 (161,117) 寬 313 → 中心 317.5;
// 文字左緣 204 + 折行寬 224 / 2 = 316。差 1.5px 是原版自己的,不是這裡算錯。
func TestConfirmTextBlockIsCenteredOnTheBox(t *testing.T) {
	boxCenter := float64(confirmBoxX) + 313.0/2
	textCenter := float64(confirmTextX) + float64(confirmTextW)/2
	if d := boxCenter - textCenter; d < -2 || d > 2 {
		t.Errorf("文字塊中心 %.1f 與底框中心 %.1f 差太多(%.1f)", textCenter, boxCenter, d)
	}
}

// 折行:中文沒有空白可斷,逐字累積;'\n' 強制換行。
func TestWrapToWidthBreaksOnNewlineWithoutFont(t *testing.T) {
	// fnt 為 nil 時不折行(量不到寬度),整段原樣回傳——不能因此 panic 或吞掉文字。
	got := wrapToWidth(nil, "一二三\n四五六", 12, 40)
	if len(got) != 1 || got[0] != "一二三\n四五六" {
		t.Errorf("沒有字型時應原樣回傳,實得 %q", got)
	}
}
