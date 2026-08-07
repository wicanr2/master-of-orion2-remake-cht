package main

import "testing"

// choosenetplyrs_test.go:名冊畫面的版面護欄。
//
// 這張畫面的視窗**尺寸隨人數變**(中段面板每位玩家重複一次),所以測的是算式,
// 不是「常數等於某個數」。

// x = (0x280 − 資產27.寬)/2、總高 = 資產28.高×列數 − 1 + 資產27.高 + 資產29.高、
// y = (0x1E0 − 總高)/2。
func TestChooseNetPlayersWindowFollowsTheOriginalFormula(t *testing.T) {
	for _, c := range []struct{ rows, wantY int }{
		{4, (moo2ScreenH - (36*4 - 1 + 81 + 38)) / 2},
		{8, (moo2ScreenH - (36*8 - 1 + 81 + 38)) / 2},
		{1, (moo2ScreenH - (36*1 - 1 + 81 + 38)) / 2},
	} {
		x, y := cnpWindow(c.rows)
		if want := (moo2ScreenW - cnpBannerW) / 2; x != want {
			t.Errorf("%d 人:x 應為 %d,實得 %d", c.rows, want, x)
		}
		if y != c.wantY {
			t.Errorf("%d 人:y 應為 %d,實得 %d", c.rows, c.wantY, y)
		}
	}
	// 人數越多視窗越高 → y 越小。這是「尺寸隨資料變」這件事本身的護欄。
	_, y4 := cnpWindow(4)
	_, y8 := cnpWindow(8)
	if y8 >= y4 {
		t.Errorf("8 人的視窗應該比 4 人高(y 更小):y4=%d y8=%d", y4, y8)
	}
}

// 每列點擊區:x1=+0x6A、x2=+0x1B3、y1=+(i×0x24+0x40)、高 0x1D。
func TestChooseNetPlayersRowRectMatchesTheImmediates(t *testing.T) {
	winX, winY := cnpWindow(4)
	for i := 0; i < 4; i++ {
		x, y, w, h := cnpRowRect(winX, winY, i)
		if x != winX+0x6A {
			t.Errorf("第 %d 列 x 應為 winX+106,實得 %d", i, x-winX)
		}
		if w != 0x1B3-0x6A {
			t.Errorf("第 %d 列寬應為 435−106 = %d,實得 %d", i, 0x1B3-0x6A, w)
		}
		if y != winY+i*0x24+0x40 {
			t.Errorf("第 %d 列 y 應為 winY+i×36+64,實得 %d", i, y-winY)
		}
		if h != 0x1D {
			t.Errorf("第 %d 列高應為 29,實得 %d", i, h)
		}
	}
	// 列距要等於中段面板的高——原版就是一片一片疊上去的,對不上代表其中一個抄錯。
	_, y0, _, _ := cnpRowRect(winX, winY, 0)
	_, y1, _, _ := cnpRowRect(winX, winY, 1)
	if y1-y0 != cnpRowH {
		t.Errorf("列距 %d 應等於中段面板高 %d", y1-y0, cnpRowH)
	}
}

// 八列(上限)時整個視窗仍要留在畫面內。
func TestChooseNetPlayersFitsOnScreenAtMaxRows(t *testing.T) {
	winX, winY := cnpWindow(cnpMaxRows)
	if winY < 0 {
		t.Errorf("8 列時 y 是負的(%d)——視窗超出畫面上緣", winY)
	}
	bottom := winY + cnpBannerH + cnpMaxRows*cnpRowH + cnpFootH
	if bottom > moo2ScreenH {
		t.Errorf("8 列時底緣 %d 超出畫面高 %d", bottom, moo2ScreenH)
	}
	if winX+cnpBannerW > moo2ScreenW {
		t.Errorf("右緣 %d 超出畫面寬 %d", winX+cnpBannerW, moo2ScreenW)
	}
}

// 大廳狀態兩行字要在底框**外面**、而且 8 人時仍留在畫面內。
//
// 這條是截圖抓出來的:第一版把字畫進底框裡,第一行壓在金屬圓角上、第二行掉出視窗。
func TestChooseNetPlayersInfoLinesSitBelowTheWindowAndStayOnScreen(t *testing.T) {
	for rows := 1; rows <= cnpMaxRows; rows++ {
		_, winY := cnpWindow(rows)
		bottom := winY + cnpBannerH + rows*cnpRowH + cnpFootH
		y1, y2 := cnpInfoBaselines(winY, rows)
		if y1 <= bottom {
			t.Errorf("%d 人:第一行基線 %d 落在視窗內(底緣 %d)——會壓到底框", rows, y1, bottom)
		}
		if y2 <= y1 {
			t.Errorf("%d 人:第二行 %d 應在第一行 %d 下方", rows, y2, y1)
		}
		if y2 > moo2ScreenH-3 {
			t.Errorf("%d 人:第二行基線 %d 太靠近畫面下緣 %d", rows, y2, moo2ScreenH)
		}
	}
}
