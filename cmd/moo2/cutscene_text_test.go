package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestCutscenePlayerTextIsExternalized(t *testing.T) {
	src, err := os.ReadFile("cutscene.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	for _, forbidden := range []string{".tr(", ".DrawCentered(", "點擊跳過", "click to skip"} {
		if strings.Contains(text, forbidden) {
			t.Errorf("cutscene.go 不得保留非原版提示或內嵌玩家文案：%q", forbidden)
		}
	}
	for _, key := range []string{"cutscene.transition.main_menu", "cutscene.transition.hall_of_fame"} {
		for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("缺少外部過場轉場文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestCutsceneSkipAcceptsMouseOrAnyKey(t *testing.T) {
	for name, input := range map[string]shell.InputState{
		"滑鼠": {ClickReleased: true},
		"鍵盤": {AnyKeyPressed: true},
	} {
		if !cutsceneSkipRequested(input) {
			t.Errorf("%s 輸入應可跳過原版過場", name)
		}
	}
	if cutsceneSkipRequested(shell.InputState{}) {
		t.Error("空輸入不得跳過過場")
	}
}
