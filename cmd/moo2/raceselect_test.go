package main

import "testing"

// 種族選擇的中文按鈕與 remake 能力資訊框都必須有固定的內框；這支擋住
// 「再把說明文字塞回肖像下方空白」而重現越過機械飾板的問題。
func TestRaceSelectTextPanelsRemainInsideTheirRects(t *testing.T) {
	s := &raceSelectScreen{}
	for i := range raceSelectList {
		x, y, w, h := s.rowRect(i)
		tr := s.rowTextRect(i)
		tx, ty := tr.contentX(), tr.contentY()
		tw, th := tr.w-2*tr.insetX, tr.h-2*tr.insetY
		if tw <= 0 || th <= 0 {
			t.Fatalf("第 %d 個種族按鈕沒有可用的中文字內框：(%d,%d,%d,%d)", i, x, y, w, h)
		}
		if x < 0 || y < 0 || x+w > moo2ScreenW || y+h > moo2ScreenH {
			t.Fatalf("第 %d 個種族按鈕超出畫布：(%d,%d,%d,%d)", i, x, y, w, h)
		}
		if tx < x+rsButtonFaceInset || ty < y+rsButtonFaceInset || tx+tw > x+w-rsButtonFaceInset || ty+th > y+h-rsButtonFaceInset {
			t.Fatalf("第 %d 個種族譯文內框越過按鈕面：button=(%d,%d,%d,%d), text=(%d,%d,%d,%d)",
				i, x, y, w, h, tx, ty, tw, th)
		}
		if tx+tw/2 != x+w/2 || ty+th/2 != y+h/2 {
			t.Fatalf("第 %d 個種族譯文沒有保持按鈕整數像素中心：button=(%d,%d,%d,%d), text=(%d,%d,%d,%d)",
				i, x, y, w, h, tx, ty, tw, th)
		}
	}

	cx, cy, cw, _ := s.cancelRect()
	if rsInfoX <= cx+cw || rsInfoY < cy {
		t.Fatalf("能力資訊框與取消鈕重疊：資訊=(%d,%d,%d,%d),取消=(%d,%d,%d)",
			rsInfoX, rsInfoY, rsInfoW, rsInfoH, cx, cy, cw)
	}
	if rsInfoX+rsInfoW > rsPortX+rsPortW || rsInfoY+rsInfoH > 438 {
		t.Fatalf("能力資訊框越過肖像下方安全帶：資訊=(%d,%d,%d,%d)", rsInfoX, rsInfoY, rsInfoW, rsInfoH)
	}
	if raceSelectInfoDescriptionTextRect().y+raceSelectInfoDescriptionTextRect().h > rsInfoY+rsInfoH {
		t.Fatalf("種族說明會越過資訊框：infoY=%d infoH=%d", rsInfoY, rsInfoH)
	}
}
