package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestColonySummaryCatalogAndTextBounds(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		keys := []string{"colony.summary.row.name", "colony.summary.build.empty", "colony.summary.build.progress", "colony.summary.build.backlog", "colony.summary.build.built", "colony.summary.empire.treasury", "colony.summary.planet.climate", "colony.summary.production.food", "colony.summary.production.starving", "colony.summary.transition.galaxy", "colony.summary.transition.colony", "colony.summary.transition.summary"}
		for _, key := range keys {
			if got := uiText(lang, key); got == "" || got == key {
				t.Fatalf("lang=%v key=%s unresolved: %q", lang, key, got)
			}
		}
		for i := 0; i < 10; i++ {
			if got := planetEnvironmentLabel(lang, "climate", i); strings.HasPrefix(got, "planet.environment") {
				t.Fatalf("missing climate %d", i)
			}
		}
		checkClippedTextFits(t, fnt, colonySummaryColumnRect(0, 0), colonySummaryText(lang, "colony.summary.row.name", 999), 11)
		checkClippedTextFits(t, fnt, colonySummaryBuildRect(0), colonySummaryText(lang, "colony.summary.build.with_built",
			strings.Repeat("EXTREMELY LONG BUILD ITEM", 4), colonySummaryText(lang, "colony.summary.build.built", strings.Repeat("超長建築名", 12))), 8)
		checkClippedTextFits(t, fnt, colonySummaryEmpireRect(0), colonySummaryText(lang, "colony.summary.empire.treasury", 999999), 8)
		checkClippedTextFits(t, fnt, colonySummaryPlanetInfoRect(2), colonySummaryText(lang, "colony.summary.planet.minerals", mineralsName(lang, gamedata.ULTRA_RICH)), 7)
		checkClippedTextFits(t, fnt, colonySummaryProductionRect(0), colonySummaryText(lang, "colony.summary.production.food", 999999, 999999, 999999), 9)
	}
}

func TestColonySummarySourceHasNoInlineTranslationOrDirectFontDraw(t *testing.T) {
	raw, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "func (b *sceneBuilder) colonySummary()")
	end := strings.Index(source[start:], "// races 建原版種族關係畫面")
	if start < 0 || end < 0 {
		t.Fatal("colonySummary source slice not found")
	}
	slice := source[start : start+end]
	if strings.Contains(slice, ".tr(") {
		t.Fatal("colonySummary() must not embed translated player text")
	}
	if strings.Contains(slice, ".font.Draw(") || strings.Contains(slice, "s.font.Draw(") {
		t.Fatal("colonySummary() must use textSafeRect for postDraw text")
	}
}

func TestColonySummaryRectsStayInsideCellsAndPanels(t *testing.T) {
	for row := range colonySummaryRowY {
		top, bottom := colonySummaryRowY[row]-15, colonySummaryRowY[row]+15
		for col := 0; col < 4; col++ {
			r := colonySummaryColumnRect(row, col)
			if r.y < top || r.y+r.h > bottom {
				t.Fatalf("row=%d col=%d rect=%+v", row, col, r)
			}
		}
		build := colonySummaryBuildRect(row)
		if build.y < top || build.y+build.h > bottom {
			t.Fatalf("row=%d build rect outside: %+v", row, build)
		}
	}
	for row := 0; row < 5; row++ {
		if r := colonySummaryEmpireRect(row); r.y+r.h > 438 {
			t.Fatalf("empire row=%d outside panel: %+v", row, r)
		}
	}
	for row := 0; row < 5; row++ {
		if r := colonySummaryPlanetInfoRect(row); r.y+r.h > 438 {
			t.Fatalf("planet row=%d outside panel: %+v", row, r)
		}
	}
	for row := 0; row < 5; row++ {
		if r := colonySummaryProductionRect(row); r.y+r.h > 438 {
			t.Fatalf("production row=%d outside panel: %+v", row, r)
		}
	}
}
