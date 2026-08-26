package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestOfficerCatalogAndTextBounds(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	keys := []string{
		"officer.transition.screen", "officer.transition.galaxy", "officer.marker.selected", "officer.marker.candidate",
		"officer.target.colony", "officer.target.no_colony", "officer.target.ship", "officer.target.no_ship",
		"officer.roster.mercenary", "officer.roster.hired", "officer.roster.assigned_colony",
		"officer.roster.assigned_ship", "officer.roster.unassigned", "officer.roster.empty",
		"officer.message.hired", "officer.message.colony_assigned", "officer.message.ship_assigned",
	}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		for _, key := range keys {
			if got := uiText(lang, key); got == "" || got == key {
				t.Fatalf("lang=%v key=%s unresolved: %q", lang, key, got)
			}
		}
		checkClippedTextFits(t, fnt, officerTargetTextRect(), officerText(lang, "officer.target.ship", 999, strings.Repeat("超長艦名", 12)), officerDynamicFont)
		checkClippedTextFits(t, fnt, officerMessageTextRect(), officerText(lang, "officer.message.colony_assigned", strings.Repeat("超長領袖名", 8), 999), officerDynamicFont)
		checkClippedTextFits(t, fnt, officerNameTextRect(0), uiText(lang, "officer.marker.selected")+strings.Repeat("超長領袖名", 8), 12)
		checkClippedTextFits(t, fnt, officerSkillTextRect(0), officerText(lang, "officer.roster.mercenary", "EXTREMELY LONG SKILL NAME", 99, 999999), officerDynamicFont)
		checkClippedTextFits(t, fnt, officerAssignmentTextRect(0), officerText(lang, "officer.roster.assigned_colony", 999), officerDynamicFont)
		checkClippedTextFits(t, fnt, officerEmptyTextRect(), uiText(lang, "officer.roster.empty"), 11)
	}
}

func TestOfficerTextRectsStayInsideRowsAndAboveButtons(t *testing.T) {
	for row, center := range officerRowCenters() {
		top, bottom := int(center)-54, int(center)+54
		for _, r := range []textSafeRect{officerNameTextRect(row), officerSkillTextRect(row), officerAssignmentTextRect(row)} {
			if r.y < top || r.y+r.h > bottom {
				t.Fatalf("row=%d rect=%+v outside %d..%d", row, r, top, bottom)
			}
			if r.y+r.h > officerButtonY {
				t.Fatalf("row=%d rect=%+v intrudes into buttons", row, r)
			}
		}
	}
	if officerTargetTextRect().y+officerTargetTextRect().h > officerMessageTextRect().y {
		t.Fatal("target and message rectangles overlap")
	}
	if officerMessageTextRect().y+officerMessageTextRect().h > officerAssignmentTextRect(0).y {
		t.Fatal("message rectangle overlaps first-row assignment status")
	}
	if officerEmptyTextRect().x+officerEmptyTextRect().w > officerTargetTextRect().x {
		t.Fatal("empty-roster hint intrudes into the right-side status panel")
	}
}

func TestOfficerSourceHasNoInlineTranslation(t *testing.T) {
	raw, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "func (b *sceneBuilder) officer()")
	end := strings.Index(source[start:], "// info 建原版科技總覽畫面")
	if start < 0 || end < 0 {
		t.Fatal("officer source slice not found")
	}
	slice := source[start : start+end]
	if strings.Contains(slice, ".tr(") {
		t.Fatal("officer() must not embed translated player text")
	}
	if strings.Contains(slice, "▶ ") || strings.Contains(slice, "◆ ") {
		t.Fatal("officer markers must come from JSON")
	}
}
