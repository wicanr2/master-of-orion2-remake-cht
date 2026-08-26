package main

import (
	"os"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func gameMenuTextKeys() []string {
	return []string{
		"gamemenu.button.save", "gamemenu.button.load", "gamemenu.button.new",
		"gamemenu.button.quit", "gamemenu.button.settings", "gamemenu.button.return",
		"gamemenu.label.music", "gamemenu.label.sound_fx",
		"gamemenu.message.no_saves", "gamemenu.transition.galaxy",
		"gamemenu.transition.new_game", "gamemenu.transition.main_menu",
	}
}

func TestGameMenuPlayerTextComesFromExternalCatalog(t *testing.T) {
	for _, key := range gameMenuTextKeys() {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("遊戲內選單缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestSliderVolumeAt(t *testing.T) {
	const x, w = 100, 155
	for _, tc := range []struct {
		name  string
		mouse int
		want  float64
	}{
		{name: "left edge mutes", mouse: 100, want: 0},
		{name: "right edge full", mouse: 254, want: 1},
		{name: "middle", mouse: 177, want: 77.0 / 154.0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sliderVolumeAt(tc.mouse, x, w); got != tc.want {
				t.Fatalf("sliderVolumeAt(%d, %d, %d) = %v, want %v", tc.mouse, x, w, got, tc.want)
			}
		})
	}
}

func TestGameMenuTextSafeRectsStayInsideOwners(t *testing.T) {
	s := &gameMenuScreen{winX: gameMenuWinX, winY: gameMenuWinY, winW: 276, winH: 376,
		btnImg: make([]*ebiten.Image, len(gameMenuButtons))}
	for i := range gameMenuButtons {
		x, y, w, h := s.btnRect(i)
		r := s.buttonTextRect(i)
		if r.x < x || r.y < y || r.x+r.w > x+w || r.y+r.h > y+h {
			t.Fatalf("按鈕 %d 文字安全框超出按鈕：%+v owner=(%d,%d,%d,%d)", i, r, x, y, w, h)
		}
		if 2*r.x+r.w != 2*x+w || 2*r.y+r.h != 2*y+h {
			t.Fatalf("按鈕 %d 文字與可見／點擊框中心不同", i)
		}
	}
	for i := 0; i < 2; i++ {
		r := s.volumeLabelTextRect(i)
		if r.x < s.winX || r.y < s.winY || r.x+r.w > s.winX+s.winW || r.y+r.h > s.winY+s.winH {
			t.Fatalf("音量標籤 %d 超出視窗：%+v", i, r)
		}
	}
	for name, r := range map[string]textSafeRect{"訊息": s.messageTextRect()} {
		if r.x < 0 || r.y < 0 || r.x+r.w > moo2ScreenW || r.y+r.h > moo2ScreenH {
			t.Fatalf("%s 文字安全框超出畫面：%+v", name, r)
		}
	}
}

func TestGameMenuLongestExternalTextIsWidthBounded(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	s := &gameMenuScreen{winX: gameMenuWinX, winY: gameMenuWinY, winW: 276, winH: 376,
		btnImg: make([]*ebiten.Image, len(gameMenuButtons))}
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		values := []struct {
			r    textSafeRect
			v    string
			size float64
		}{
			{s.buttonTextRect(0), uiText(lang, "gamemenu.button.save"), 12},
			{s.volumeLabelTextRect(1), uiText(lang, "gamemenu.label.sound_fx"), 18},
			{s.messageTextRect(), uiText(lang, "gamemenu.message.no_saves"), 12},
		}
		for _, tc := range values {
			w, _ := fnt.Measure(tc.r.clipped(fnt, tc.v, tc.size), tc.size)
			if w > tc.r.contentWidth() {
				t.Fatalf("%q 裁切後寬 %.1f 超過安全框 %.1f", tc.v, w, tc.r.contentWidth())
			}
		}
	}
}

func TestGameMenuRealAssets(t *testing.T) {
	dir := os.Getenv("MOO2_GAMEMENU_TEST")
	if dir == "" {
		t.Skip("未設 MOO2_GAMEMENU_TEST，跳過正版 GAME.LBX 測試")
	}
	res, err := assets.NewResolver(dir)
	if err != nil {
		t.Fatal(err)
	}
	b := &sceneBuilder{res: res}
	for _, id := range []int{0, 1, 2, 3, 4, 5, 6, 7} {
		im, _ := b.gameMenuImage(id, id != 0)
		if im == nil {
			t.Errorf("GAME.LBX#%d 未能以 BUFFER0.LBX#0 調色盤解碼", id)
		}
	}
}

func TestGameMenuSourceHasNoEmbeddedPlayerTextOrDirectDraw(t *testing.T) {
	raw, err := os.ReadFile("gamemenu.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, forbidden := range []string{".tr(", ".fnt.Draw(", ".fnt.DrawCentered("} {
		if strings.Contains(src, forbidden) {
			t.Errorf("gamemenu.go 不得出現 %s", forbidden)
		}
	}
	for _, value := range []string{"SAVE GAME", "儲存遊戲", "音樂", "遷移連線：", "No saved games yet", "主選單"} {
		if strings.Contains(src, `"`+value+`"`) {
			t.Errorf("gamemenu.go 仍內嵌玩家文案 %q", value)
		}
	}
}
