package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestFleetRosterCatalogAndTextBounds(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	keys := []string{"fleet.roster.unknown_location", "fleet.roster.header.active", "fleet.roster.header.inactive", "fleet.roster.transit", "fleet.roster.split", "fleet.roster.damage", "fleet.transition.design", "fleet.transition.officers", "fleet.transition.fleet", "fleet.transition.galaxy"}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		for _, key := range keys {
			if got := uiText(lang, key); got == "" || got == key {
				t.Fatalf("lang=%v key=%s unresolved: %q", lang, key, got)
			}
		}
		header := fmt.Sprintf(uiText(lang, "fleet.roster.header.active"), 999, "AN EXTREMELY LONG SYSTEM NAME", 999)
		header += fmt.Sprintf(uiText(lang, "fleet.roster.transit"), "ANOTHER EXTREMELY LONG SYSTEM", 999)
		checkClippedTextFits(t, fnt, fleetHeaderTextRect(312), header, fleetRosterFont)
		checkClippedTextFits(t, fnt, fleetSplitTextRect(332), fmt.Sprintf(uiText(lang, "fleet.roster.split"), 999), fleetRosterFont)
		checkClippedTextFits(t, fnt, fleetShipNameTextRect(350), uiText(lang, "fleet.selection.selected")+strings.Repeat("超長艦艇名稱", 8), fleetRosterFont)
		checkClippedTextFits(t, fnt, fleetShipClassTextRect(350), "EXTREMELY LONG SHIP CLASS", fleetRosterFont)
		checkClippedTextFits(t, fnt, fleetShipDamageTextRect(350), fmt.Sprintf(uiText(lang, "fleet.roster.damage"), 100), fleetRosterFont)
	}
}

func TestFleetRosterColumnsDoNotOverlap(t *testing.T) {
	rects := []textSafeRect{fleetShipNameTextRect(350), fleetShipClassTextRect(350), fleetShipDamageTextRect(350)}
	for i := 1; i < len(rects); i++ {
		if rects[i-1].x+rects[i-1].w > rects[i].x {
			t.Fatalf("columns overlap: %+v %+v", rects[i-1], rects[i])
		}
	}
	for _, r := range append(rects, fleetHeaderTextRect(312), fleetSplitTextRect(332)) {
		if r.x < fleetRosterX || r.x+r.w > fleetRosterX+fleetRosterW {
			t.Fatalf("rect leaves roster: %+v", r)
		}
	}
}

func TestFleetSourceHasNoInlineTranslation(t *testing.T) {
	raw, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "func (b *sceneBuilder) fleet()")
	end := strings.Index(source[start:], "// 艦艇設計畫面的原版座標")
	if start < 0 || end < 0 {
		t.Fatal("fleet source slice not found")
	}
	if strings.Contains(source[start:start+end], ".tr(") {
		t.Fatal("fleet() must not embed translated player text")
	}
}
