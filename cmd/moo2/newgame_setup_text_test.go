package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestNewGameSetupDynamicTextComesFromCatalog(t *testing.T) {
	keys := []string{"newgame.setting.galaxy_size", "newgame.setting.empires", "newgame.transition.setup", "newgame.transition.main_menu"}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		for _, key := range keys {
			if got := uiText(lang, key); got == "" || got == key {
				t.Fatalf("lang=%v key=%s unresolved: %q", lang, key, got)
			}
		}
		_ = fmt.Sprintf(uiText(lang, "newgame.setting.galaxy_size"), "HUGE GALAXY", 999)
		_ = fmt.Sprintf(uiText(lang, "newgame.setting.empires"), 99)
	}
}

func TestNewGameSetupValueRectsFitLongestLabels(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		b := &sceneBuilder{lang: lang, newGameDiff: 4, newGameSize: 3, newGameAge: 2,
			newGameEmpires: 8, newGameTech: 2}
		for _, setting := range ngSettings {
			x, y, w, h := ngStripRect(setting)
			r := ngStripTextRect(setting)
			if 2*r.x+r.w != 2*x+w || 2*r.y+r.h != 2*y+h {
				t.Fatalf("setting %s text/hit center mismatch: rect=%+v hit=%d,%d,%d,%d", setting.act, r, x, y, w, h)
			}
			label := setting.label(b)
			fits := false
			for _, size := range []float64{12, 11, 10} {
				clipped := r.clipped(fnt, label, size)
				_, measuredH := fnt.Measure(clipped, size)
				if measuredH <= float64(r.h-2*r.insetY) {
					fits = true
					break
				}
			}
			if !fits {
				t.Fatalf("lang=%v setting %s cannot fit label %q", lang, setting.act, label)
			}
		}
	}
}

func TestNewGameSetupSourceHasNoInlineTranslation(t *testing.T) {
	raw, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "var ngSettings")
	end := strings.Index(source[start:], "// newGamePics")
	if start < 0 || end < 0 {
		t.Fatal("new game setup source slice not found")
	}
	if strings.Contains(source[start:start+end], ".tr(") {
		t.Fatal("new game setup slice must not embed translated player text")
	}
}
