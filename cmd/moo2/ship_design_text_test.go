package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestShipDesignCatalogAndTextBounds(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	keys := []string{"shipdesign.title", "shipdesign.button.clear", "shipdesign.button.cancel", "shipdesign.button.build", "shipdesign.transition.screen", "shipdesign.transition.fleet", "shipdesign.message.no_space", "shipdesign.message.no_treasury", "shipdesign.component.weapon", "shipdesign.component.line", "shipdesign.arc", "shipdesign.ammo.variable", "shipdesign.total", "shipdesign.total.unknown", "shipdesign.unlocked", "shipdesign.space.row", "shipdesign.mods.available"}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		for _, key := range keys {
			if got := uiText(lang, key); got == "" || got == key {
				t.Fatalf("lang=%v key=%s unresolved: %q", lang, key, got)
			}
		}
		for i := 0; i < 6; i++ {
			checkClippedTextFits(t, fnt, shipDesignHullCostRect(i), shipDesignText(lang, "shipdesign.hull.cost", 999999), 8)
		}
		checkClippedTextFits(t, fnt, shipDesignComponentNameRect(0), shipDesignText(lang, "shipdesign.component.line", uiText(lang, "shipdesign.component.weapon"), strings.Repeat("EXTREMELY LONG WEAPON", 4)), 9)
		checkClippedTextFits(t, fnt, shipDesignArcTextRect(), shipDesignText(lang, "shipdesign.arc", "360 DEGREE", 999), 9)
		checkClippedTextFits(t, fnt, shipDesignTotalTextRect(), shipDesignText(lang, "shipdesign.total", strings.Repeat("超長艦體名", 4), 999999, 8, 8, 99), 9)
		checkClippedTextFits(t, fnt, shipDesignStatusTextRect(), shipDesignText(lang, "shipdesign.message.no_space", strings.Repeat("超長艦體名", 6)), 9)
		checkClippedTextFits(t, fnt, shipDesignSpaceRowRect(5), shipDesignText(lang, "shipdesign.space.row", "DOOM STAR", 999999, 999999), 8)
		checkClippedTextFits(t, fnt, shipDesignModHeaderRect(), uiText(lang, "shipdesign.mods.available"), 8)
		checkClippedTextFits(t, fnt, shipDesignModTextRect(7), strings.Repeat("No Range Dissipation", 3), 8)
	}
}

func TestShipDesignTextRectsDoNotEnterModArea(t *testing.T) {
	if shipDesignStatusTextRect().y+shipDesignStatusTextRect().h > shipDesignSpaceHeaderRect().y {
		t.Fatal("status overlaps space header")
	}
	if shipDesignSpaceHeaderRect().y+shipDesignSpaceHeaderRect().h > shipDesignSpaceRowRect(0).y {
		t.Fatal("space header overlaps rows")
	}
	for i := 0; i < 6; i++ {
		if i > 0 && shipDesignHullCostRect(i-1).y+shipDesignHullCostRect(i-1).h > shipDesignHullCostRect(i).y {
			t.Fatalf("hull cost rows %d/%d overlap", i-1, i)
		}
		if r := shipDesignSpaceRowRect(i); r.y+r.h > shipDesignModHeaderRect().y {
			t.Fatalf("space row %d enters mod area: %+v", i, r)
		}
	}
	for i := 0; i < 8; i++ {
		if r := shipDesignModTextRect(i); r.y+r.h > 431 {
			t.Fatalf("mod %d crosses panel: %+v", i, r)
		}
	}
}

func TestShipDesignSourceHasNoInlineTranslation(t *testing.T) {
	raw, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "func (b *sceneBuilder) shipDesign()")
	end := strings.Index(source[start:], "// officer 建原版軍官列表畫面")
	if start < 0 || end < 0 {
		t.Fatal("shipDesign source slice not found")
	}
	if strings.Contains(source[start:start+end], ".tr(") {
		t.Fatal("shipDesign() must not embed translated player text")
	}
	for _, forbidden := range []string{"\"Ship Design\"", "\"Frigate\"", "\"Destroyer\"", "\"Cruiser\"", "\"Battleship\"", "\"Titan\"", "\"Doom Star\"", "\"Clear\"", "\"Cancel\"", "\"Build\""} {
		if strings.Contains(source[start:start+end], forbidden) {
			t.Errorf("shipDesign() 仍內嵌玩家顯示文字 %s", forbidden)
		}
	}
}
