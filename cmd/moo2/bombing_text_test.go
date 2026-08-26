package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func bombingTextKeys() []string {
	return []string{
		"bombing.transition.galaxy", "bombing.error.no_session",
		"bombing.title.default", "bombing.title.colony",
		"bombing.report.salvo", "bombing.report.buildings",
		"bombing.report.population", "bombing.report.population_bio",
		"bombing.report.retaliated", "bombing.report.no_retaliation",
		"bombing.button.continue",
	}
}

func TestBombingPlayerTextComesFromExternalCatalog(t *testing.T) {
	for _, key := range bombingTextKeys() {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("軌道轟炸缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestBombingExternalTextFitsSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		checkClippedTextFits(t, fnt, bombingTitleTextRect(),
			fmt.Sprintf(uiText(lang, "bombing.title.colony"), strings.Repeat("ABCDEFGHIJKLMNOPQRSTUVWXYZ銀河殖民地", 6)), 18)
		lines := []string{
			fmt.Sprintf(uiText(lang, "bombing.report.salvo"), 2147483647, 2147483647),
			fmt.Sprintf(uiText(lang, "bombing.report.buildings"), 2147483647, 2147483647),
			fmt.Sprintf(uiText(lang, "bombing.report.population_bio"), 2147483647, 2147483647, 2147483647),
			fmt.Sprintf(uiText(lang, "bombing.report.retaliated"), 2147483647),
		}
		for row, value := range lines {
			checkClippedTextFits(t, fnt, bombingReportTextRect(row), value, 14)
		}
		s := &bombingScreen{}
		checkClippedTextFits(t, fnt, s.continueTextRect(), uiText(lang, "bombing.button.continue"), 14)
	}
}

func TestBombingSafeRectsStayInsideOwners(t *testing.T) {
	for row := 0; row < 4; row++ {
		r := bombingReportTextRect(row)
		if r.x < 0 || r.x+r.w > moo2ScreenW || r.y < 60 || r.y+r.h > 178 {
			t.Fatalf("轟炸戰報第 %d 列超出報告面板：%+v", row, r)
		}
	}
	s := &bombingScreen{}
	x, y, w, h := s.contRect()
	r := s.continueTextRect()
	if 2*r.x+r.w != 2*x+w || 2*r.y+r.h != 2*y+h {
		t.Fatal("轟炸按鈕文字框與可見外框／熱區中心不同")
	}
}

func TestBombingSourceHasNoEmbeddedPlayerTextOrDirectDraw(t *testing.T) {
	raw, err := os.ReadFile("bombing.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, forbidden := range []string{".tr(", ".fnt.Draw(", ".fnt.DrawCentered("} {
		if strings.Contains(src, forbidden) {
			t.Errorf("bombing.go 不得出現 %s", forbidden)
		}
	}
	for _, value := range []string{"ORBITAL BOMBARDMENT", "Salvo damage", "軌道轟炸", "繼續", "星系主畫面", "無對局"} {
		if strings.Contains(src, `"`+value+`"`) {
			t.Errorf("bombing.go 仍內嵌玩家文案 %q", value)
		}
	}
}
