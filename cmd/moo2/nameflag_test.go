package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestNameFlagPlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"nameflag.default_empire_name",
		"nameflag.title.name_empire",
		"nameflag.hint.keyboard_edit",
		"nameflag.title.banner_color",
		"nameflag.label.banner",
		"nameflag.button.back",
		"nameflag.button.start",
	}
	for _, fc := range shell.FlagColors {
		keys = append(keys, "nameflag.color."+fc.Key)
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("%s 缺少語言 %v 的外部文案", key, lang)
			}
		}
	}
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		label := fmt.Sprintf(uiText(lang, "nameflag.label.banner"), uiText(lang, "nameflag.color.silver"))
		if label == "" || strings.Contains(label, "%s") {
			t.Errorf("語言 %v 的旗色格式未正確套用：%q", lang, label)
		}
	}
}

func TestNameFlagSourceHasNoEmbeddedPlayerText(t *testing.T) {
	raw, err := os.ReadFile("nameflag.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	if strings.Contains(src, ".tr(") {
		t.Fatal("nameflag.go 不得再用 tr 內嵌雙語玩家文案")
	}
	for _, value := range []string{"NAME YOUR EMPIRE", "為你的帝國命名", "START GAME", "開始遊戲", "Galactic Empire", "銀河帝國"} {
		if strings.Contains(src, `"`+value+`"`) {
			t.Errorf("nameflag.go 仍內嵌玩家文案 %q", value)
		}
	}
}
