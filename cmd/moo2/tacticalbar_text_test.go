package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestTacticalBarPlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"tactical.mode.scan_on", "tactical.mode.board_on", "tactical.mode.normal",
		"tactical.options.combat_return_unavailable", "tactical.retreat.summary",
		"tactical.auto.no_target", "tactical.auto.stopped_suffix", "tactical.auto.round_cap",
		"tactical.scan.summary", "tactical.board.select_ship", "tactical.board.unavailable",
		"tactical.board.no_marines", "tactical.board.captured", "tactical.board.victory_suffix",
		"tactical.board.repelled", "tactical.mode.scan_hint", "tactical.mode.board_hint",
	}
	for _, button := range barButtonsCHT {
		keys = append(keys, button.textKey)
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("格子戰術控制列缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestTacticalBarButtonTextFitsOriginalPlates(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		for _, button := range barButtonsCHT {
			label := uiText(lang, button.textKey)
			width, height := fnt.Measure(label, 12)
			if width > barButtonPlateW-4 || height > barButtonPlateH {
				t.Errorf("%s (%v) 尺寸 %.1fx%.1f 超出控制列文字安全框 %dx%d",
					button.action, lang, width, height, barButtonPlateW-4, barButtonPlateH)
			}
			if !pointInRect(button.cx, button.cy, button.cx-27, button.cy-9, 54, 18) {
				t.Errorf("%s 的文字中心未落在熱區中心", button.action)
			}
		}
	}
}

func TestTacticalBarFormatContractsMatch(t *testing.T) {
	wants := map[string]int{
		"tactical.retreat.summary":  1,
		"tactical.scan.summary":     10,
		"tactical.board.no_marines": 1,
		"tactical.board.captured":   2,
		"tactical.board.repelled":   4,
	}
	for key, want := range wants {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := strings.Count(uiText(lang, key), "%"); got != want {
				t.Errorf("%s (%v) 有 %d 個格式參數，預期 %d", key, lang, got, want)
			}
		}
	}
}

func TestTacticalBarSourceHasNoEmbeddedPlayerSentences(t *testing.T) {
	raw, err := os.ReadFile("tacticalbar.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	if strings.Contains(src, ".tr(") {
		t.Fatal("tacticalbar.go 不得再用 tr 內嵌中英文玩家文案")
	}
	for _, value := range []string{
		"Scan mode: click an enemy ship", "掃描模式：點一艘敵艦",
		"The full original settings screen" + " is not built yet", "完整的原版設定畫面" + "尚未完成",
	} {
		if strings.Contains(src, `"`+value) {
			t.Errorf("tacticalbar.go 仍內嵌玩家文案 %q", value)
		}
	}
}
