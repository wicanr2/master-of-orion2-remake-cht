package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestInputBoxPlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"inputbox.title.game_name", "inputbox.title.host_address",
		"inputbox.button.accept", "inputbox.hint.accept_cancel",
		"network.host.default_name",
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("輸入框缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestInputBoxTextFitsSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	const x, y = inboxDefaultX, inboxDefaultY
	fx, fy, fw, fh := inboxFieldRect(x, y)
	inputRect := inboxInputTextRect(fx, fy, fw, fh)
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		checkClippedTextFits(t, fnt, inboxTitleTextRect(x, y), uiText(lang, "inputbox.title.host_address"), 15)
		checkClippedTextFits(t, fnt, inboxOKTextRect(x, y), uiText(lang, "inputbox.button.accept"), 12)
		checkClippedTextFits(t, fnt, inboxHintTextRect(x, y), uiText(lang, "inputbox.hint.accept_cancel"), 11)
		visible := inboxVisibleInputText(fnt, "ABCDEFGHIJKLMNOPQRSTUVWXYZ超長主機位址輸入內容", 14, inputRect)
		w, h := fnt.Measure(visible, 14)
		if w > inputRect.contentWidth() || h > float64(inputRect.h-2*inputRect.insetY) {
			t.Errorf("輸入內容加游標 %.1fx%.1f 超出安全框 %.1fx%d", w, h,
				inputRect.contentWidth(), inputRect.h-2*inputRect.insetY)
		}
	}
	bx, by, bw, bh := inboxOKRect(x, y)
	buttonText := inboxOKTextRect(x, y)
	if 2*buttonText.x+buttonText.w != 2*bx+bw || 2*buttonText.y+buttonText.h != 2*by+bh {
		t.Error("ACCEPT 文字框與 98×28 熱區中心不一致")
	}
	hint := inboxHintTextRect(x, y)
	if hint.y < fy+fh {
		t.Errorf("鍵盤提示 y=%d 回侵輸入欄底緣 %d", hint.y, fy+fh)
	}
	if hint.y+hint.h > by {
		t.Errorf("鍵盤提示底緣 %d 侵入按鈕頂緣 %d", hint.y+hint.h, by)
	}
}

func TestInputBoxSourcesHaveNoEmbeddedPlayerSentencesOrDirectDraw(t *testing.T) {
	for _, file := range []string{"inputbox.go", "choosenetplyrs.go", "choosemultinetgame.go"} {
		raw, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		src := string(raw)
		if file == "inputbox.go" {
			for _, forbidden := range []string{".tr(", ".fnt.Draw(", ".fnt.DrawCentered("} {
				if strings.Contains(src, forbidden) {
					t.Errorf("%s 不得出現 %s", file, forbidden)
				}
			}
		}
		for _, value := range []string{
			"Enter host address", "輸入主機位址", "Game name", "對局名稱",
			"Enter = accept", "Enter 確定", "主機玩家",
		} {
			if strings.Contains(src, `"`+value) {
				t.Errorf("%s 仍內嵌輸入框玩家文案 %q", file, value)
			}
		}
	}
}
