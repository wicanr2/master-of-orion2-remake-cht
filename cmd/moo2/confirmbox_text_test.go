package main

import (
	"os"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestConfirmButtonTextComesFromExternalCatalog(t *testing.T) {
	for _, key := range []string{"confirm.button.yes", "confirm.button.no"} {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("確認框缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestConfirmButtonTextFitsAndSharesHitRectCenter(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	buttons := []struct {
		x, y int
		key  string
	}{
		{confirmYesX, confirmYesY, "confirm.button.yes"},
		{confirmNoX, confirmNoY, "confirm.button.no"},
	}
	for _, button := range buttons {
		r := confirmButtonTextRect(button.x, button.y)
		if 2*r.x+r.w != 2*button.x+confirmBtnW || 2*r.y+r.h != 2*button.y+confirmBtnH {
			t.Errorf("確認框按鈕文字框與熱區中心不一致：%s", button.key)
		}
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			checkClippedTextFits(t, fnt, r, uiText(lang, button.key), 10)
		}
	}
}

func TestConfirmFallbackDrawsButtonsInBothLanguages(t *testing.T) {
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		if !confirmNeedsButtonLabel(lang, false) {
			t.Errorf("%v 缺資產時未啟用外部按鈕標籤", lang)
		}
		b := &sceneBuilder{lang: lang, fnt: uifont.LoadBitmapTC()}
		s := &confirmScreen{b: b, msg: uiText(lang, "relocation.confirm.monster")}
		dst := ebiten.NewImage(640, 480)
		s.draw(dst)
	}
	if confirmNeedsButtonLabel(i18n.English, true) {
		t.Error("英文正版資產存在時不應覆蓋原版烘字")
	}
	if !confirmNeedsButtonLabel(i18n.Traditional, true) {
		t.Error("繁中正版資產存在時仍應擦底疊字")
	}
}

func TestConfirmSourceHasNoEmbeddedButtonTextOrDirectDraw(t *testing.T) {
	raw, err := os.ReadFile("confirmbox.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, forbidden := range []string{".fnt.Draw(", ".fnt.DrawCentered("} {
		if strings.Contains(source, forbidden) {
			t.Errorf("confirmbox.go 不得出現 %s", forbidden)
		}
	}
	for _, value := range []string{"是", "否", "YES", "NO"} {
		if strings.Contains(source, `"`+value+`"`) {
			t.Errorf("confirmbox.go 仍內嵌固定玩家文案 %q", value)
		}
	}
}
