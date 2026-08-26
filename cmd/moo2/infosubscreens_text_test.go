package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestInfoSubscreenPlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"info.no_game_data", "info.history.metric", "info.history.insufficient_turns",
		"info.tech.researching", "info.tech.completed_count", "info.races.ai_relations",
		"info.summary.treasury_value", "info.reference.categories", "info.reference.howto",
		"info.reference.footer",
	}
	keys = append(keys, infoTabTextKeys[:]...)
	for i := 1; i <= 14; i++ {
		keys = append(keys, fmt.Sprintf("info.reference.category.%02d", i))
	}
	for i := 1; i <= 10; i++ {
		keys = append(keys, fmt.Sprintf("info.reference.howto.%02d", i))
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("INFO 缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestInfoSubscreenStaticTextFitsSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		for _, key := range infoTabTextKeys {
			checkClippedTextFits(t, fnt, infoTitleTextRect(), uiText(lang, key), 15)
		}
		checkClippedTextFits(t, fnt, infoCenteredTextRect(int(infoPanelY)+26, 20),
			fmt.Sprintf(uiText(lang, "info.history.metric"), historyMetricLabel(lang, 0)), 11)
		checkClippedTextFits(t, fnt, infoContentTextRect(int(infoPanelY)+44, 18),
			fmt.Sprintf(uiText(lang, "info.tech.researching"), "Trans Dimensional", 999999), 12)
		checkClippedTextFits(t, fnt, infoSummaryValueTextRect(int(infoPanelY)+63),
			fmt.Sprintf(uiText(lang, "info.summary.treasury_value"), 999999, -99999), 11)
		for i, value := range infoTextList(lang, "info.reference.category.", 14) {
			checkClippedTextFits(t, fnt, infoReferenceTextRect(0, int(infoPanelY)+64+i*17), "• "+value, 10)
		}
		for i, value := range infoTextList(lang, "info.reference.howto.", 10) {
			checkClippedTextFits(t, fnt, infoReferenceTextRect(1, int(infoPanelY)+64+i*17), "• "+value, 10)
		}
	}
}

func TestInfoSubscreenSourceHasNoEmbeddedPlayerSentences(t *testing.T) {
	raw, err := os.ReadFile("infosubscreens.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	if strings.Contains(src, ".tr(") {
		t.Fatal("infosubscreens.go 不得再用 tr 內嵌中英文玩家文案")
	}
	for _, value := range []string{"HISTORY GRAPH", "歷史曲線圖", "No game data yet", "尚無對局資料", "Full rules:", "詳細規則見"} {
		if strings.Contains(src, `"`+value+`"`) {
			t.Errorf("infosubscreens.go 仍內嵌玩家文案 %q", value)
		}
	}
}
