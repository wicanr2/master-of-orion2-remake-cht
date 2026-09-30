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

func TestInfoStanceLabelUsesStoredCodeAndCurrentLanguage(t *testing.T) {
	for _, tc := range []struct {
		stored, code string
	}{
		{"宣戰", "war"},
		{"敵視", "hostile"},
		{"中立", "neutral"},
		{"提議貿易", "trade"},
		{"提議結盟", "alliance"},
		{"", "unknown"},
	} {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			key := "info.races.stance." + tc.code
			if got, want := infoStanceLabel(lang, tc.stored), uiText(lang, key); got != want || got == key {
				t.Errorf("存檔態勢 %q、語系 %v：%q，預期 %q", tc.stored, lang, got, want)
			}
		}
	}
}

func TestInfoStanceLabelSurvivesChangedTranslation(t *testing.T) {
	// 先初始化共用 catalog，再暫時換成譯文已更新的版本；舊存檔仍寫「宣戰」。
	uiText(i18n.Traditional, "info.races.stance.war")
	original := uiCatalogZH
	updated := i18n.New(i18n.Traditional)
	if _, err := updated.LoadJSON(strings.NewReader(`[
		{"key":"info.races.stance.war","english":"At War","value":"交戰狀態"},
		{"key":"info.races.stance.unknown","english":"Unknown","value":"未知"}
	]`)); err != nil {
		t.Fatal(err)
	}
	uiCatalogZH = updated
	t.Cleanup(func() { uiCatalogZH = original })
	for _, tc := range []struct {
		lang i18n.Lang
		want string
	}{
		{i18n.English, "At War"},
		{i18n.Traditional, "交戰狀態"},
	} {
		if got := infoStanceLabel(tc.lang, "宣戰"); got != tc.want {
			t.Errorf("舊存檔在語系 %v 顯示 %q，預期 %q", tc.lang, got, tc.want)
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
