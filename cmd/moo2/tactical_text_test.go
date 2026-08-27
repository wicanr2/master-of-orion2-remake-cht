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
		"tactical.log.ship_done", "tactical.log.no_other_ready", "tactical.log.ship_waits",
		"tactical.log.already_finished", "tactical.log.waiting_ship",
		"tactical.log.move_insufficient", "tactical.log.moved", "tactical.log.no_player_ready",
		"tactical.log.no_active_mount", "tactical.log.phase_cloaked", "tactical.log.in_stasis",
		"tactical.log.ammo_depleted", "tactical.log.outside_arc", "tactical.log.out_of_range",
		"tactical.log.action_damage", "tactical.log.multi_mount_damage",
		"tactical.log.round_volley", "tactical.log.round_complete", "tactical.log.fighter_suffix",
		"tactical.log.victory", "tactical.log.defeat", "tactical.label.player_ship",
		"tactical.transition.result",
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
		"tactical.log.ship_done":          1,
		"tactical.log.ship_waits":         1,
		"tactical.log.already_finished":   1,
		"tactical.log.move_insufficient":  3,
		"tactical.log.moved":              4,
		"tactical.log.phase_cloaked":      1,
		"tactical.log.in_stasis":          1,
		"tactical.log.action_damage":      2,
		"tactical.log.multi_mount_damage": 1,
		"tactical.log.round_volley":       4,
		"tactical.log.round_complete":     2,
		"tactical.log.fighter_suffix":     2,
		"tactical.weapon.mode_changed":    2,
		"tactical.weapon.description":     7,
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
	start := strings.Index(src, "func newTacticalScreenForShips")
	end := strings.Index(src, "// battleResult 顯示")
	if start < 0 || end <= start {
		t.Fatal("找不到格子戰術來源切片邊界")
	}
	if strings.Contains(src[start:end], ".tr(") {
		t.Fatal("格子戰術來源切片不得再用 tr 內嵌玩家文案")
	}
	for _, fixed := range []string{"Weapon slot %d", "standby once", "彈藥%s；改造%s"} {
		if strings.Contains(src, `"`+fixed) {
			t.Errorf("interactive.go 仍內嵌格子戰術固定文案 %q", fixed)
		}
	}
}
