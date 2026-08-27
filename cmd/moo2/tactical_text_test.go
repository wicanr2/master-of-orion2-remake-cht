package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
)

func TestTacticalCoreAndWeaponTextComesFromCatalog(t *testing.T) {
	keys := []string{
		"tactical.title", "tactical.log.initial",
		"tactical.weapon.mode.ready", "tactical.weapon.mode.standby", "tactical.weapon.mode.off",
		"tactical.weapon.mode.ready_log", "tactical.weapon.mode.standby_log", "tactical.weapon.mode.off_log",
		"tactical.weapon.mode_changed", "tactical.weapon.ammo_unlimited",
		"tactical.weapon.description", "tactical.weapon.fallback_name",
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("格子戰術缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
	for key, want := range map[string]int{
		"tactical.weapon.mode_changed": 2,
		"tactical.weapon.description":  7,
	} {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := strings.Count(uiText(lang, key), "%"); got != want {
				t.Errorf("%s (%v) 有 %d 個格式參數，預期 %d", key, lang, got, want)
			}
		}
	}
}

func TestTacticalWeaponFixedTextIsNotEmbedded(t *testing.T) {
	raw, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, fixed := range []string{"Weapon slot %d", "standby once", "彈藥%s；改造%s"} {
		if strings.Contains(src, `"`+fixed) {
			t.Errorf("interactive.go 仍內嵌格子戰術固定文案 %q", fixed)
		}
	}
}
