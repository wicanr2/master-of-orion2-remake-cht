package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestGalaxyStarPanelFixedTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"galaxy.status.tax", "galaxy.star_panel.button.close", "galaxy.star_panel.monster",
		"galaxy.star_panel.special", "galaxy.star_panel.environment.climate_size",
		"galaxy.star_panel.environment.gravity_minerals", "galaxy.star_panel.marines.fleet",
		"galaxy.star_panel.marines.garrison", "galaxy.star_panel.crew", "galaxy.star_panel.crew_next",
		"galaxy.star_panel.status.transit", "galaxy.star_panel.status.arrived",
		"galaxy.star_panel.button.load_marines", "galaxy.star_panel.button.bombard",
		"galaxy.star_panel.button.invade", "galaxy.star_panel.button.mind_control",
		"galaxy.star_panel.button.attack", "galaxy.star_panel.button.colonize",
		"galaxy.star_panel.button.outpost", "galaxy.star_panel.button.target",
		"galaxy.star_panel.button.dispatch", "galaxy.star_panel.error.no_ftl",
		"galaxy.star_panel.result.marines_loaded", "galaxy.star_panel.error.no_marines",
		"galaxy.star_panel.result.mind_control", "galaxy.star_panel.result.colonized",
		"galaxy.star_panel.result.outpost",
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Fatalf("缺少星圖側欄外部文案 %s (%v)", key, lang)
			}
		}
	}
}

func TestGalaxyStarPanelTemplatesFormatAndFit(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	cases := []struct {
		key  string
		args []any
		r    textSafeRect
		size float64
	}{
		{"galaxy.star_panel.environment.climate_size", []any{"TERRAN", "HUGE"}, starPanelEnvironmentTextRect(353), 11},
		{"galaxy.star_panel.environment.gravity_minerals", []any{"NORMAL-G", "ULTRA RICH"}, starPanelEnvironmentTextRect(369), 11},
		{"galaxy.star_panel.marines.garrison", []any{999, 999}, starPanelMarineTextRect(), 11},
		{"galaxy.star_panel.status.transit", []any{999}, starPanelButtonTextRect(402), 11},
		{"galaxy.star_panel.button.attack", []any{"ANCIENT GUARDIAN"}, starPanelButtonTextRect(402), 12},
		{"galaxy.star_panel.button.target", []any{"▶ BUILD OUTPOST", "A VERY LONG PLANET NAME"}, starPanelButtonTextRect(402), 12},
	}
	for _, tc := range cases {
		for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
			text := fmt.Sprintf(uiText(lang, tc.key), tc.args...)
			if strings.Contains(text, "%!") {
				t.Fatalf("星圖側欄模板格式錯誤 %s (%v)：%q", tc.key, lang, text)
			}
			checkClippedTextFits(t, fnt, tc.r, text, tc.size)
		}
	}
}

func TestGalaxyStarPanelBottomRowIsInsideVisibleFrame(t *testing.T) {
	bottom := starPanelButtonTextRect(446)
	if bottom.y+bottom.h != starPanelY+starPanelH {
		t.Fatalf("最下列底緣=%d，面板底緣=%d；文字必須完整落在框內", bottom.y+bottom.h, starPanelY+starPanelH)
	}
	if bottom.y+bottom.h > moo2ScreenH {
		t.Fatalf("最下列超出 640×480 畫布：%+v", bottom)
	}
}

func TestGalaxyStarPanelSliceHasNoEmbeddedFixedText(t *testing.T) {
	raw, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	start := strings.Index(source, "// starPanelTextLayoutStart")
	end := strings.Index(source, "// starPanelTextLayoutEnd")
	if start < 0 || end <= start {
		t.Fatal("找不到星圖側欄守門標記")
	}
	slice := source[start:end]
	if strings.Contains(slice, ".tr(") {
		t.Fatal("星圖側欄不得再以 tr 內嵌固定中英文文案")
	}
	for _, phrase := range []string{"氣候 %s", "艦隊陸戰隊", "▶ 軌道轟炸", "▶ 派遣艦隊至此星"} {
		if strings.Contains(slice, `"`+phrase) {
			t.Errorf("星圖側欄仍內嵌固定文案 %q", phrase)
		}
	}
}
