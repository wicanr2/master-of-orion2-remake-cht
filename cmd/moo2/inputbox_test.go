package main

import "testing"

// inputbox_test.go:文字輸入彈窗的護欄。版面釘立即數,行為釘「不會爆」與長度上限。

// 版面全部相對彈窗左上角,而且要留在 288×151 的框裡。
func TestInputBoxLayoutMatchesTheImmediates(t *testing.T) {
	const x, y = inboxDefaultX, inboxDefaultY
	fx, fy, fw, fh := inboxFieldRect(x, y)
	if fx != x+0x22 || fy != y+0x36 || fh != 0x1A {
		t.Errorf("輸入欄應為 (x+34, y+54, 高 26),實得 (x+%d, y+%d, 高 %d)", fx-x, fy-y, fh)
	}
	if fw != inboxBoxW-0x36 {
		t.Errorf("輸入欄寬應為 288−54 = %d,實得 %d", inboxBoxW-0x36, fw)
	}
	bx, by, bw, bh := inboxOKRect(x, y)
	if bx != x+0x60 || by != y+0x64 {
		t.Errorf("OK 鈕應為 (x+96, y+100),實得 (x+%d, y+%d)", bx-x, by-y)
	}
	if bw != 98 || bh != 28 {
		t.Errorf("OK 鈕尺寸應為資產 1 的 98×28,實得 %d×%d", bw, bh)
	}
	// 兩者都要在彈窗框內。
	if fx < x || fx+fw > x+inboxBoxW || fy < y || fy+fh > y+inboxBoxH {
		t.Error("輸入欄跑出彈窗框")
	}
	if bx < x || bx+bw > x+inboxBoxW || by < y || by+bh > y+inboxBoxH {
		t.Error("OK 鈕跑出彈窗框")
	}
	// 輸入欄與 OK 鈕不能重疊——原版是上下排的。
	if fy+fh > by {
		t.Errorf("輸入欄底緣 %d 壓到 OK 鈕頂緣 %d", fy+fh, by)
	}
	// 整個彈窗要在畫面內。
	if x < 0 || y < 0 || x+inboxBoxW > moo2ScreenW || y+inboxBoxH > moo2ScreenH {
		t.Errorf("彈窗 (%d,%d,%d,%d) 超出 640×480", x, y, inboxBoxW, inboxBoxH)
	}
}

// 左右邊距**不對稱**(34 / 20)——那是兩個獨立的立即數,不是抄錯。
//
// 這條存在的理由是防「順手把它改成對稱」:看起來像 bug 的東西不一定是 bug。
func TestInputBoxFieldMarginsAreDeliberatelyAsymmetric(t *testing.T) {
	fx, _, fw, _ := inboxFieldRect(0, 0)
	left := fx
	right := inboxBoxW - (fx + fw)
	if left != 34 {
		t.Errorf("左邊距應為 34(0x22),實得 %d", left)
	}
	if right != 20 {
		t.Errorf("右邊距應為 20(288−34−234),實得 %d", right)
	}
	if left == right {
		t.Error("兩邊變成一樣了——那代表有人把原版的兩個立即數之一改掉了")
	}
}

// 長度上限:呼叫端給的值優先,但一律夾在 205(原版 0xCD)以內。
func TestInputBoxClampsMaxLength(t *testing.T) {
	for _, c := range []struct{ in, want int }{
		{8, 8}, {1, 1}, {205, 205},
		{0, inboxMaxLenCap}, {-3, inboxMaxLenCap}, {9999, inboxMaxLenCap},
	} {
		if got := inboxClampMaxLen(c.in); got != c.want {
			t.Errorf("上限 %d 應夾成 %d,實得 %d", c.in, c.want, got)
		}
	}
}

// 打字不能超過上限;而且超過之後**不是截斷成亂碼**,是直接不收。
func TestInputBoxStopsAcceptingAtTheLimit(t *testing.T) {
	s := &inputBoxScreen{max: 8}
	s.typeRunes([]rune("ORIONIAN"))
	if s.Text() != "ORIONIAN" {
		t.Fatalf("8 字元應該剛好收得下,實得 %q", s.Text())
	}
	s.typeRunes([]rune("XYZ"))
	if s.Text() != "ORIONIAN" {
		t.Errorf("超過上限後不該再收,實得 %q", s.Text())
	}
}

// 控制字元不進緩衝區——Enter / Tab 之類的會被 AppendInputChars 送進來。
func TestInputBoxDropsControlCharacters(t *testing.T) {
	s := &inputBoxScreen{max: 20}
	s.typeRunes([]rune{'a', '\n', 'b', '\t', 'c', 0x7f})
	if s.Text() != "abc" {
		t.Errorf("控制字元應被丟掉,實得 %q", s.Text())
	}
}

// 空字串按退格不能 panic——這是最容易漏的一條。
func TestInputBoxBackspaceOnEmptyIsSafe(t *testing.T) {
	s := &inputBoxScreen{max: 8}
	s.backspace()
	s.backspace()
	if s.Text() != "" {
		t.Errorf("空字串退格後仍該是空的,實得 %q", s.Text())
	}
	s.typeRunes([]rune("ab"))
	s.backspace()
	if s.Text() != "a" {
		t.Errorf("退格應刪掉最後一個字元,實得 %q", s.Text())
	}
}

// 中文一個字是一個 rune,不是三個 byte——上限 8 要能打 8 個中文字。
func TestInputBoxCountsRunesNotBytes(t *testing.T) {
	s := &inputBoxScreen{max: 8}
	s.typeRunes([]rune("銀河霸主二代對局"))
	if got := []rune(s.Text()); len(got) != 8 {
		t.Errorf("8 個中文字應該剛好收得下,實得 %d 個:%q", len(got), s.Text())
	}
}

// 初值超過上限時要先截掉,而不是讓緩衝區一開始就違規。
func TestInputBoxTruncatesInitialValue(t *testing.T) {
	b := &sceneBuilder{}
	s := b.inputBox(nil, "inputbox.title.game_name", "ABCDEFGHIJKL", 8, nil)
	if s.Text() != "ABCDEFGH" {
		t.Errorf("初值應截到 8 字元,實得 %q", s.Text())
	}
}

// accept 會把前後空白去掉——位址欄尤其容易貼到多餘的空白。
func TestInputBoxTrimsOnAccept(t *testing.T) {
	var got string
	b := &sceneBuilder{}
	s := b.inputBox(nil, "inputbox.title.host_address", "  192.168.1.20:24501  ", 45,
		func(v string) *origTransition { got = v; return nil })
	s.accept()
	if got != "192.168.1.20:24501" {
		t.Errorf("accept 應去掉前後空白,實得 %q", got)
	}
}
