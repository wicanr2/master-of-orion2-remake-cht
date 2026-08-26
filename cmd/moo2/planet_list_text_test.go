package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestPlanetListCatalogAndTextBounds(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	keys := []string{"planet.list.status.colony", "planet.list.status.outpost", "planet.list.empty"}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		for _, key := range keys {
			text := uiText(lang, key)
			if text == "" || text == key {
				t.Fatalf("lang=%v key=%s unresolved: %q", lang, key, text)
			}
		}
		checkClippedTextFits(t, fnt, planetListColumnRect(0, 4, true), uiText(lang, "planet.list.status.colony"), 9)
		checkClippedTextFits(t, fnt, planetListColumnRect(0, 4, true), uiText(lang, "planet.list.status.outpost"), 9)
		checkClippedTextFits(t, fnt, planetListEmptyTextRect(), uiText(lang, "planet.list.empty"), 12)
	}
	checkClippedTextFits(t, fnt, planetListColumnRect(0, 0, false), "AN EXTREMELY LONG PLANET NAME", 12)
	checkClippedTextFits(t, fnt, planetListMessageTextRect(), "Fleet en route to AN EXTREMELY LONG SYSTEM NAME (999 turns).", 9)
}

func TestPlanetListRectsStayInsideRowsAndAboveActions(t *testing.T) {
	for row := 0; row < planetListRows; row++ {
		top := int(planetListRowY[row]) - 25
		bottom := top + 50
		for column := 0; column < 5; column++ {
			for _, secondary := range []bool{false, true} {
				r := planetListColumnRect(row, column, secondary)
				if r.y < top || r.y+r.h > bottom {
					t.Fatalf("row=%d column=%d secondary=%v rect=%+v outside %d..%d", row, column, secondary, r, top, bottom)
				}
			}
		}
	}
	if r := planetListMessageTextRect(); r.y+r.h > 386 {
		t.Fatalf("message rect intrudes into colony button: %+v", r)
	}
}

func TestPlanetListSourceHasNoInlineTranslation(t *testing.T) {
	raw, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "func (b *sceneBuilder) planets()")
	end := strings.Index(source[start:], "// planetListAction")
	if start < 0 || end < 0 {
		t.Fatal("planets source slice not found")
	}
	if strings.Contains(source[start:start+end], ".tr(") {
		t.Fatal("planets() must not embed translated player text")
	}
}
