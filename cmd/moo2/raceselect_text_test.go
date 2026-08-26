package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestRaceSelectPlayerTextComesFromExternalCatalog(t *testing.T) {
	staticKeys := []string{
		"race.select.title",
		"race.select.cancel",
		"race.select.transition.new_game",
		"race.select.empire_name",
	}
	for _, key := range staticKeys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("種族選擇缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
	for _, entry := range raceSelectList {
		for _, field := range []string{"name", "adjective", "description"} {
			for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
				if got := raceSelectEntryText(lang, entry, field); got == "" || strings.Contains(got, entry.key+"."+field) {
					t.Errorf("種族 %s 缺少 %s 外部文案 (%v)", entry.key, field, lang)
				}
			}
		}
	}
}

func TestRaceSelectFormatsAndTextFitSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	s := &raceSelectScreen{}
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		for i, entry := range raceSelectList {
			checkClippedTextFits(t, fnt, s.rowTextRect(i), raceSelectEntryText(lang, entry, "name"), 14)
			checkClippedTextFits(t, fnt, raceSelectInfoNameTextRect(), raceSelectEntryText(lang, entry, "name"), 12)
			description := raceSelectEntryText(lang, entry, "description")
			checkClippedTextFits(t, fnt, raceSelectInfoDescriptionTextRect(), description, 8)
			if got := raceSelectInfoDescriptionTextRect().clipped(fnt, description, 8); got != description {
				t.Errorf("種族能力摘要必須完整顯示：%s %q → %q", entry.key, description, got)
			}
			empire := fmt.Sprintf(uiText(lang, "race.select.empire_name"), raceSelectEntryText(lang, entry, "adjective"))
			if strings.Contains(empire, "%!") {
				t.Errorf("種族帝國名稱格式參數不相容：%q", empire)
			}
		}
		checkClippedTextFits(t, fnt, raceSelectTitleTextRect(), uiText(lang, "race.select.title"), 16)
		checkClippedTextFits(t, fnt, s.cancelTextRect(), uiText(lang, "race.select.cancel"), 14)
		if got := s.cancelTextRect().clipped(fnt, uiText(lang, "race.select.cancel"), 14); got != uiText(lang, "race.select.cancel") {
			t.Errorf("CANCEL adapter 必須完整顯示：%q", got)
		}
	}
}

func TestRaceSelectSourcesContainNoEmbeddedPlayerText(t *testing.T) {
	raw, err := os.ReadFile("raceselect.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	if strings.Contains(src, ".tr(") {
		t.Fatal("raceselect.go 不得再以 tr 內嵌雙語玩家文案")
	}
	for _, value := range []string{
		"選擇你的種族", "自行分配種族點數", "Spend your own race picks", "Human Empire",
		"阿爾卡里", "Darloks", "CANCEL",
	} {
		if strings.Contains(src, `"`+value) {
			t.Errorf("raceselect.go 仍內嵌玩家文案 %q", value)
		}
	}
}
