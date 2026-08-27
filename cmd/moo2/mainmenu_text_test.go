package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestMainMenuExternalTextAndBounds(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	keys := []string{
		"mainmenu.toggle.language", "mainmenu.toggle.rules", "mainmenu.transition.main_menu",
		"mainmenu.transition.new_game", "mainmenu.transition.galaxy", "mainmenu.transition.hiscore",
	}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		for _, key := range keys {
			if got := uiText(lang, key); got == "" || got == key {
				t.Fatalf("lang=%v key=%s unresolved: %q", lang, key, got)
			}
		}
		checkClippedTextFits(t, fnt, mainMenuToggleTextRect(menuLanguageY),
			mainMenuText(lang, "mainmenu.toggle.language"), 11)
		checkClippedTextFits(t, fnt, mainMenuToggleTextRect(menuVersionY),
			mainMenuText(lang, "mainmenu.toggle.rules", "1.5 COMMUNITY PATCH"), 11)
		for _, tc := range []struct {
			r     textSafeRect
			value string
		}{
			{mainMenuToggleTextRect(menuLanguageY), mainMenuText(lang, "mainmenu.toggle.language")},
			{mainMenuToggleTextRect(menuVersionY), mainMenuText(lang, "mainmenu.toggle.rules", "1.5")},
		} {
			if got := tc.r.clipped(fnt, tc.value, 11); got != tc.value {
				t.Fatalf("normal menu value was truncated: %q -> %q", tc.value, got)
			}
		}
	}
}

func TestMainMenuToggleTextRectsFollowHitRows(t *testing.T) {
	for _, y := range []int{menuLanguageY, menuVersionY} {
		owner := mainMenuToggleOwnerRect(y)
		text := mainMenuToggleTextRect(y)
		if text.x < owner.x || text.y < owner.y || text.x+text.w > owner.x+owner.w || text.y+text.h > owner.y+owner.h {
			t.Fatalf("text rect outside owner: text=%+v owner=%+v", text, owner)
		}
		if 2*text.x+text.w != 2*owner.x+owner.w || 2*text.y+text.h != 2*owner.y+owner.h {
			t.Fatalf("text rect and owner have different centers: text=%+v owner=%+v", text, owner)
		}
	}
	if mainMenuToggleOwnerRect(menuLanguageY).y+menuToggleHitH > mainMenuToggleOwnerRect(menuVersionY).y {
		t.Fatal("toggle rows overlap")
	}
}

func TestMainMenuSourceHasNoInlineTranslation(t *testing.T) {
	raw, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "func (b *sceneBuilder) menu()")
	if start < 0 {
		t.Fatal("menu source slice not found")
	}
	end := strings.Index(source[start:], "// tr 是")
	if end < 0 {
		t.Fatal("menu source slice end not found")
	}
	slice := source[start : start+end]
	for _, forbidden := range []string{".tr(", `"語言 繁體中文`, `"Language: English`, `"規則版本`, `"Rules %s`} {
		if strings.Contains(slice, forbidden) {
			t.Fatalf("menu() still embeds player text %q", forbidden)
		}
	}
}
