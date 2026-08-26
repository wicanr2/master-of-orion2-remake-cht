package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func loadSaveTextKeys() []string {
	return []string{
		"loadsave.transition.menu", "loadsave.transition.galaxy",
		"loadsave.slot.empty", "loadsave.slot.autosave_empty", "loadsave.slot.unnamed_empire",
		"loadsave.slot.numbered", "loadsave.slot.autosave", "loadsave.slot.stardate",
		"loadsave.message.empty_slot", "loadsave.message.autosave_readonly",
		"loadsave.message.pick_save_slot", "loadsave.message.no_session",
		"loadsave.message.write_failed", "loadsave.message.pick_load_slot", "loadsave.message.load_failed",
		"loadsave.button.load", "loadsave.button.save", "loadsave.button.cancel",
	}
}

func TestLoadSavePlayerTextComesFromExternalCatalog(t *testing.T) {
	for _, key := range loadSaveTextKeys() {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("存讀檔彈窗缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestEnglishGameFontFallbackIsNonNil(t *testing.T) {
	fnt, err := loadFont("", i18n.English)
	if err != nil {
		t.Fatal(err)
	}
	if fnt == nil {
		t.Fatal("英文動態玩家文案不得因未指定外部向量字型而取得 nil 字型")
	}
}

func TestLoadSaveSafeRectsStayInsideOwners(t *testing.T) {
	s := &loadGameScreen{winX: 182, winY: 52, winW: 276, winH: 376}
	for i := 0; i < 10; i++ {
		x, y, w, h := s.slotRect(i)
		for name, r := range map[string]textSafeRect{
			"名稱": s.slotTitleTextRect(i), "星曆": s.slotStardateTextRect(i), "時間": s.slotTimeTextRect(i),
		} {
			if r.x < x || r.y < y || r.x+r.w > x+w || r.y+r.h > y+h {
				t.Fatalf("第 %d 格%s安全框超出槽位：%+v owner=(%d,%d,%d,%d)", i, name, r, x, y, w, h)
			}
		}
		if s.slotTitleTextRect(i).x+s.slotTitleTextRect(i).w >= s.winX+loadSlotIconX {
			t.Fatalf("第 %d 格名稱侵入對局圖示欄", i)
		}
		if s.slotTimeTextRect(i).x+s.slotTimeTextRect(i).w >= s.winX+loadSlotIconX {
			t.Fatalf("第 %d 格時間侵入對局圖示欄", i)
		}
	}
	lx, ly, lw, lh := s.loadRect()
	cx, cy, cw, ch := s.cancelRect()
	for _, pair := range []struct {
		name       string
		r          textSafeRect
		x, y, w, h int
	}{
		{"動作", s.actionTextRect(), lx, ly, lw, lh},
		{"取消", s.cancelTextRect(), cx, cy, cw, ch},
	} {
		if 2*pair.r.x+pair.r.w != 2*pair.x+pair.w || 2*pair.r.y+pair.r.h != 2*pair.y+pair.h {
			t.Fatalf("%s按鈕文字框與可見外框／熱區中心不同", pair.name)
		}
	}
	msg := s.messageTextRect()
	if msg.y < 0 || msg.y+msg.h > moo2ScreenH {
		t.Fatalf("狀態訊息框超出畫面：%+v", msg)
	}
}

func TestLoadSaveLongestExternalTextIsWidthBounded(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	s := &loadGameScreen{winX: 182, winY: 52, winW: 276, winH: 376}
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		values := []struct {
			r textSafeRect
			v string
			s float64
		}{
			{s.slotTitleTextRect(0), fmt.Sprintf(uiText(lang, "loadsave.slot.numbered"), 10, strings.Repeat("銀河帝國ABCDEFGHIJKLMNOPQRSTUVWXYZ", 8)), 12},
			{s.slotStardateTextRect(0), fmt.Sprintf(uiText(lang, "loadsave.slot.stardate"), "999999999999"), 10},
			{s.slotTimeTextRect(0), "99/12/31 23:59", 7},
			{s.messageTextRect(), uiText(lang, "loadsave.message.autosave_readonly"), 12},
		}
		for _, tc := range values {
			clipped := tc.r.clipped(fnt, tc.v, tc.s)
			w, _ := fnt.Measure(clipped, tc.s)
			if w > tc.r.contentWidth() {
				t.Fatalf("%q 裁切後寬 %.1f 超過安全框 %.1f", tc.v, w, tc.r.contentWidth())
			}
		}
	}
}

func TestLoadSaveRealAssets(t *testing.T) {
	dir := os.Getenv("MOO2_LOADSAVE_TEST")
	if dir == "" {
		t.Skip("未設 MOO2_LOADSAVE_TEST，跳過正版 GAME.LBX 測試")
	}
	res, err := assets.NewResolver(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []int{loadWinBGAsset, loadWinBtnAsset, loadWinCanAsset, loadWinIconSing, loadWinIconHot} {
		im, _ := loadWinImage(res, id, id != loadWinBGAsset)
		if im == nil {
			t.Errorf("GAME.LBX#%d 未能以 MAINMENU.LBX#21 調色盤解碼", id)
		}
	}
}

func TestLoadSaveSourceHasNoEmbeddedPlayerTextOrDirectDraw(t *testing.T) {
	raw, err := os.ReadFile("loadgame.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, forbidden := range []string{".tr(", ".fnt.Draw(", ".fnt.DrawCentered("} {
		if strings.Contains(src, forbidden) {
			t.Errorf("loadgame.go 不得出現 %s", forbidden)
		}
	}
	for _, value := range []string{"That slot is empty.", "Unnamed Empire", "Stardate %s", "載入", "儲存", "取消", "主選單", "星系主畫面"} {
		if strings.Contains(src, `"`+value+`"`) {
			t.Errorf("loadgame.go 仍內嵌玩家文案 %q", value)
		}
	}
}
