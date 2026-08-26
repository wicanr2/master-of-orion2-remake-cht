package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestColonyScreenPlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"common.unknown", "common.unknown_build",
		"colony.transition.summary", "colony.transition.officers", "colony.transition.colony",
		"colony.button.leaders", "colony.button.return", "colony.button.change", "colony.button.buy",
		"colony.title.default", "colony.assimilation", "colony.population",
		"colony.output.food", "colony.output.industry", "colony.output.research", "colony.output.morale",
		"colony.buildings.title", "colony.buildings.none", "colony.buildings.separator",
		"colony.job.farmers", "colony.job.workers", "colony.job.scientists",
		"colony.job.count", "colony.job.hint", "colony.planet.minerals", "colony.planet.gravity",
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("殖民地主畫面缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestColonyScreenExternalTextFitsSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	buttons := []struct {
		x, y, w, h int
		key        string
		size       float64
	}{
		{colBtnX, colLeadersY, colBtnW, colLeadersH, "colony.button.leaders", 12},
		{colBtnX, colReturnY, colBtnW, colReturnH, "colony.button.return", 12},
		{colChangeX, colChangeY, colChangeW, colChangeH, "colony.button.change", 11},
		{colBuyX, colChangeY, colBuyW, colChangeH, "colony.button.buy", 11},
	}
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		for _, button := range buttons {
			r := colonyButtonTextRect(button.x, button.y, button.w, button.h)
			checkClippedTextFits(t, fnt, r, uiText(lang, button.key), button.size)
			if 2*r.x+r.w != 2*button.x+button.w || 2*r.y+r.h != 2*button.y+button.h {
				t.Errorf("%s 文字框與按鈕熱區中心不同", button.key)
			}
		}
		checkClippedTextFits(t, fnt, colonyLeftTitleTextRect(), "ABCDEFGHIJKLMNOPQRSTUVWXYZ", 13)
		checkClippedTextFits(t, fnt, colonyAssimilationTextRect(),
			fmt.Sprintf(uiText(lang, "colony.assimilation"), 999, 999), 10)
		checkClippedTextFits(t, fnt, colonyPopulationTextRect(),
			fmt.Sprintf(uiText(lang, "colony.population"), 999, 999), 11)
		for row, key := range []string{"colony.job.farmers", "colony.job.workers", "colony.job.scientists"} {
			checkClippedTextFits(t, fnt, colonyJobTextRect(row, 0), uiText(lang, key), 12)
			checkClippedTextFits(t, fnt, colonyJobTextRect(row, 1), fmt.Sprintf(uiText(lang, "colony.job.count"), 999), 12)
			checkClippedTextFits(t, fnt, colonyJobTextRect(row, 2), uiText(lang, "colony.job.hint"), 10)
		}
	}
}

func TestColonyScreenSourceHasNoEmbeddedPlayerSentencesOrDirectDraw(t *testing.T) {
	raw, err := os.ReadFile("colonyscreen.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, forbidden := range []string{".tr(", ".fnt.Draw(", ".fnt.DrawCentered("} {
		if strings.Contains(src, forbidden) {
			t.Errorf("colonyscreen.go 不得出現 %s", forbidden)
		}
	}
	for _, value := range []string{"LEADERS", "領袖", "Unassimilated", "未同化", "Farmers", "農夫"} {
		if strings.Contains(src, `"`+value+`"`) {
			t.Errorf("colonyscreen.go 仍內嵌玩家文案 %q", value)
		}
	}
}
