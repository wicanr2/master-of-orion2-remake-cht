package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestDiplomacyAudienceFixedTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"diplomacy.audience.fallback_enemy", "diplomacy.audience.opening",
		"diplomacy.audience.title", "diplomacy.audience.emissary",
		"diplomacy.audience.agreements", "diplomacy.audience.option.peace",
		"diplomacy.audience.option.trade", "diplomacy.audience.option.research",
		"diplomacy.audience.option.nonaggression", "diplomacy.audience.option.alliance",
		"diplomacy.audience.option.threat", "diplomacy.audience.option.tribute_5",
		"diplomacy.audience.option.tribute_10", "diplomacy.audience.option.gift_cash",
		"diplomacy.audience.option.special_food", "diplomacy.audience.option.special_research",
		"diplomacy.audience.option.gift_tech", "diplomacy.audience.option.gift_star",
		"diplomacy.audience.break.trade", "diplomacy.audience.break.research",
		"diplomacy.audience.break.formal", "diplomacy.audience.break.tribute",
		"diplomacy.audience.break.special", "diplomacy.audience.button.end",
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("外交對談缺少外部文案 %s (%v)", key, lang)
			}
		}
	}
	for _, key := range []string{
		"diplomacy.audience.opening", "diplomacy.audience.emissary",
		"diplomacy.audience.agreements",
	} {
		for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
			if got := strings.Count(uiText(lang, key), "%s"); got != 1 {
				t.Errorf("%s (%v) 的格式參數數量=%d，預期 1", key, lang, got)
			}
		}
	}
}

func TestDiplomacyAudienceButtonTextRectsMatchHitRectsAndFit(t *testing.T) {
	d := &diplomacyScreen{}
	fnt := uifont.LoadBitmapTC()
	for i, key := range []string{
		"diplomacy.audience.option.peace", "diplomacy.audience.option.trade",
		"diplomacy.audience.option.research", "diplomacy.audience.option.nonaggression",
		"diplomacy.audience.option.alliance", "diplomacy.audience.option.threat",
		"diplomacy.audience.option.tribute_5", "diplomacy.audience.option.tribute_10",
		"diplomacy.audience.option.gift_cash", "diplomacy.audience.option.special_food",
		"diplomacy.audience.option.special_research", "diplomacy.audience.option.gift_tech",
		"diplomacy.audience.option.gift_star",
	} {
		x, y, w, h := d.optRect(i)
		r := d.optTextRect(i)
		if r.x != x || r.y != y || r.w != w || r.h != h {
			t.Fatalf("提議 %d 的文字框未由熱區推導：rect=%+v hit=%d,%d,%d,%d", i, r, x, y, w, h)
		}
		for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
			checkClippedTextFits(t, fnt, r, uiText(lang, key), 15)
		}
	}
	for i, key := range []string{
		"diplomacy.audience.break.trade", "diplomacy.audience.break.research",
		"diplomacy.audience.break.formal", "diplomacy.audience.break.tribute",
		"diplomacy.audience.break.special",
	} {
		x, y, w, h := d.breakRect(i)
		r := d.breakTextRect(i)
		if r.x != x || r.y != y || r.w != w || r.h != h {
			t.Fatalf("解約 %d 的文字框未由熱區推導：rect=%+v hit=%d,%d,%d,%d", i, r, x, y, w, h)
		}
		if x < 0 || y < 0 || x+w > moo2ScreenW || y+h > moo2ScreenH {
			t.Fatalf("解約 %d 超出 640×480 畫布：%d,%d,%d,%d", i, x, y, w, h)
		}
		_, optionY, _, optionH := d.optRect(12)
		if y < optionY+optionH {
			t.Fatalf("解約列與最後一列提議重疊：breakY=%d optionBottom=%d", y, optionY+optionH)
		}
		for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
			checkClippedTextFits(t, fnt, r, uiText(lang, key), 11)
		}
	}
}

func TestDiplomacyAudienceLegacyFixedTextIsNotEmbedded(t *testing.T) {
	raw, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, text := range []string{
		"外交對談", "提議和平", "Propose Peace",
		"特殊貿易：食物換現金", "Special Trade: Food for Credits",
		"終止特殊貿易", "End Special Trade", "結束對談", "END AUDIENCE",
	} {
		if strings.Contains(source, `"`+text+`"`) {
			t.Errorf("interactive.go 仍內嵌外交對談固定文案 %q", text)
		}
	}
}
