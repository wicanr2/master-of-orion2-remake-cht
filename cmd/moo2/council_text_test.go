package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestCouncilTextCatalogAndGeometry(t *testing.T) {
	keys := []string{
		"council.title", "council.victory.decided", "council.victory.player", "council.victory.enemy",
		"council.pending.winner", "council.pending.explanation", "council.pending.accept", "council.pending.reject",
		"council.vote.prompt", "council.vote.abstain", "council.status.not_convened", "council.status.requirements",
		"council.status.convened", "council.status.no_majority", "council.breakdown.header",
		"council.breakdown.candidate", "council.breakdown.abstains", "council.breakdown.votes_for",
		"council.breakdown.row", "council.breakdown.explanation",
	}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		for _, key := range keys {
			if got := uiText(lang, key); got == "" || got == key {
				t.Fatalf("lang=%v key=%s unresolved: %q", lang, key, got)
			}
		}
	}
	for i := 0; i < 3; i++ {
		r := councilVoteTextRect(i)
		if r.x != 80+i*165 || r.y != 370 || r.w != 150 || r.h != 40 {
			t.Fatalf("vote rect %d = %+v", i, r)
		}
	}
	for i := 0; i < 2; i++ {
		r := councilDecisionTextRect(i)
		if r.x != 120 || r.y != 402+i*30 || r.w != 400 || r.h != 26 {
			t.Fatalf("decision rect %d = %+v", i, r)
		}
	}
	if strings.Count(uiText(i18n.English, "council.breakdown.header"), "%") != 5 {
		t.Fatal("breakdown header format parameter count changed")
	}
	_ = fmt.Sprintf(uiText(i18n.English, "council.breakdown.header"), 1, "A", "B", 7, 10)
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		checkClippedTextFits(t, fnt, councilDecisionTextRect(0), uiText(lang, "council.pending.accept"), 12)
		checkClippedTextFits(t, fnt, councilDecisionTextRect(1), uiText(lang, "council.pending.reject"), 12)
		checkClippedTextFits(t, fnt, councilVoteTextRect(2), uiText(lang, "council.vote.abstain"), 16)
		header := fmt.Sprintf(uiText(lang, "council.breakdown.header"), 99, "LONG CANDIDATE ALPHA", "LONG CANDIDATE BETA", 999, 999)
		checkClippedTextFits(t, fnt, councilSummaryTextRect(82, 30), header, 13)
	}
}

func TestCouncilSourceHasNoInlineTranslation(t *testing.T) {
	source, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(string(source), "func (b *sceneBuilder) council()")
	end := strings.Index(string(source)[start:], "// ============ NEW GAME")
	if start < 0 || end < 0 {
		t.Fatal("council source slice not found")
	}
	if strings.Contains(string(source)[start:start+end], ".tr(") {
		t.Fatal("council() must not embed translated player text")
	}
}
