package main

import "testing"

func TestDiplomacyProposalGridFits640x480(t *testing.T) {
	d := &diplomacyScreen{}
	for i := 0; i < 9; i++ {
		x, y, w, h := d.optRect(i)
		if x < 0 || y < 0 || x+w > moo2ScreenW || y+h > 365 {
			t.Fatalf("第 %d 個外交提案按鈕超出畫面配置: (%d,%d,%d,%d)", i, x, y, w, h)
		}
		for j := 0; j < i; j++ {
			px, py, pw, ph := d.optRect(j)
			if x < px+pw && px < x+w && y < py+ph && py < y+h {
				t.Fatalf("外交提案按鈕 %d 與 %d 重疊", i, j)
			}
		}
	}
}

func TestDiplomacyBreakButtonsStayAboveEndAudience(t *testing.T) {
	d := &diplomacyScreen{backRect: [4]int{250, 430, 140, 34}}
	for i := 0; i < 4; i++ {
		x, y, w, h := d.breakRect(i)
		if x < 0 || x+w > moo2ScreenW || y < 0 || y+h >= d.backRect[1] {
			t.Fatalf("第 %d 個終止按鈕與結束對談按鈕衝突: (%d,%d,%d,%d)", i, x, y, w, h)
		}
	}
}
