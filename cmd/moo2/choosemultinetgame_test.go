package main

import "testing"

// choosemultinetgame_test.go:對局清單的版面護欄。釘算式與立即數,不釘「畫得好不好看」。

// x = (0x280 − 479)/2、y = ((0x1E0 − 384) − 0x51)/2 + 0x25。
//
// 那個 −0x51 剛好等於標題帶的高,但這張畫面**沒有畫標題帶**——它是版面上的讓位。
// 這條測試把算式重算一次,免得日後有人「順手把那個看起來多餘的 −81 拿掉」。
func TestChooseMultiNetGameWindowMatchesTheOriginalFormula(t *testing.T) {
	x, y := cmngWindow()
	if want := (moo2ScreenW - cmngPanelW) / 2; x != want {
		t.Errorf("x 應為 %d,實得 %d", want, x)
	}
	if want := ((moo2ScreenH-cmngPanelH)-cmngYBias)/2 + cmngYOffset; y != want {
		t.Errorf("y 應為 %d,實得 %d", want, y)
	}
	// 代進真實尺寸的答案(算錯的話上面兩條會一起錯,所以另外釘一次結果)。
	if x != 80 || y != 44 {
		t.Errorf("代入 479×384 後應為 (80,44),實得 (%d,%d)", x, y)
	}
	if y+cmngPanelH > moo2ScreenH {
		t.Errorf("面板底緣 %d 超出畫面高 %d", y+cmngPanelH, moo2ScreenH)
	}
}

// 每列熱區 362×22、列距 27、起點 +64,10 列都要在面板內。
func TestChooseMultiNetGameRowsMatchTheImmediates(t *testing.T) {
	winX, winY := cmngWindow()
	for i := 0; i < cmngMaxRows; i++ {
		x, y, w, h := cmngRowRect(winX, winY, i)
		if x != winX+0x26 {
			t.Errorf("第 %d 列 x 應為 winX+38,實得 winX+%d", i, x-winX)
		}
		if w != 0x190-0x26 {
			t.Errorf("第 %d 列寬應為 400−38 = %d,實得 %d", i, 0x190-0x26, w)
		}
		if h != 0x16 {
			t.Errorf("第 %d 列高應為 22,實得 %d", i, h)
		}
		if y != winY+0x40+i*0x1B {
			t.Errorf("第 %d 列 y 應為 winY+64+i×27,實得 winY+%d", i, y-winY)
		}
		if y+h > winY+cmngPanelH {
			t.Errorf("第 %d 列底緣 %d 超出面板底緣 %d", i, y+h, winY+cmngPanelH)
		}
		if x+w > winX+cmngPanelW {
			t.Errorf("第 %d 列右緣超出面板", i)
		}
	}
	// 列距要大於列高,否則兩列的熱區會重疊——點第 1 列會選到第 2 列。
	if cmngRowStep <= cmngRowH {
		t.Errorf("列距 %d 應大於列高 %d,否則熱區重疊", cmngRowStep, cmngRowH)
	}
}

// 文字在 22 px 的列裡**垂直置中**(原版 (0x16 − 字高)/2),而且比熱區起點低 3 px。
//
// ⚠ 這裡釘的是**上緣**不是基線:uifont/ebiten text/v2 的 y 是行框上緣。
// 第一版多加了一個字高當基線,整欄字掉到下一列——截圖上選取框在第一列、字在第二列。
func TestChooseMultiNetGameTextIsCentredInTheRow(t *testing.T) {
	_, winY := cmngWindow()
	const glyphH = 13
	for i := 0; i < 3; i++ {
		got := cmngTextTop(winY, i, glyphH)
		want := winY + cmngTextFirstY + i*cmngRowStep + (cmngRowH-glyphH)/2
		if got != want {
			t.Errorf("第 %d 列文字上緣應為 %d,實得 %d", i, want, got)
		}
	}
	// 文字起點比熱區起點低 3 px,那是原版兩個立即數的差(0x43 − 0x40)。
	if cmngTextFirstY-cmngRowFirst != 3 {
		t.Errorf("文字起點應比熱區起點低 3 px,實得 %d", cmngTextFirstY-cmngRowFirst)
	}
	// 不論字高多少,整個字框都要留在那一列裡——這條才是「置中」真正要保證的事。
	for _, gh := range []int{9, 13, 18, 22} {
		top := cmngTextTop(winY, 0, gh)
		rowTop := winY + cmngRowFirst
		if top < rowTop || top+gh > rowTop+3+cmngRowH {
			t.Errorf("字高 %d 時文字 (%d..%d) 跑出列 (%d..%d)",
				gh, top, top+gh, rowTop, rowTop+cmngRowH)
		}
	}
	// 第 i 列的字絕不能落進第 i+1 列的熱區——那正是第一版的症狀。
	winX, _ := cmngWindow()
	for i := 0; i < cmngMaxRows-1; i++ {
		top := cmngTextTop(winY, i, glyphH)
		_, nextY, _, _ := cmngRowRect(winX, winY, i+1)
		if top+glyphH > nextY {
			t.Errorf("第 %d 列的字底緣 %d 掉進第 %d 列(起點 %d)", i, top+glyphH, i+1, nextY)
		}
	}
}

// 底下那顆鈕的位置是 Add_Button_Field_ 的立即數,要落在面板內。
func TestChooseMultiNetGameButtonSitsInsideThePanel(t *testing.T) {
	winX, winY := cmngWindow()
	bx, by := winX+cmngBtnX, winY+cmngBtnY
	if cmngBtnX != 0xBF || cmngBtnY != 0x158 {
		t.Errorf("按鈕位移應為 (0xBF,0x158),實得 (0x%X,0x%X)", cmngBtnX, cmngBtnY)
	}
	if bx < winX || bx > winX+cmngPanelW || by < winY || by > winY+cmngPanelH {
		t.Errorf("按鈕 (%d,%d) 不在面板內", bx-winX, by-winY)
	}
	// 按鈕要在最後一列下面,否則會壓到清單。
	_, lastY, _, lastH := cmngRowRect(winX, winY, cmngMaxRows-1)
	if by < lastY+lastH {
		t.Errorf("按鈕 y=%d 壓到最後一列(底緣 %d)", by, lastY+lastH)
	}
}
