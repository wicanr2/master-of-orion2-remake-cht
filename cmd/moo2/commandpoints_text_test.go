package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func commandPointsTextKeys() []string {
	return []string{
		"commandpoints.title", "commandpoints.label.starting", "commandpoints.label.orbital_bases",
		"commandpoints.label.total", "commandpoints.label.used", "commandpoints.label.net",
		"commandpoints.penalty", "commandpoints.close",
	}
}

func TestCommandPointsPlayerTextComesFromExternalCatalog(t *testing.T) {
	for _, key := range commandPointsTextKeys() {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("指揮點數視窗缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestCommandPointsSafeRectsStayInsidePanel(t *testing.T) {
	for name, r := range map[string]textSafeRect{
		"標題": cpTitleTextRect(), "列標": cpLabelTextRect(cpPanelY + 56),
		"數值": cpValueTextRect(cpPanelY + 56), "懲罰": cpPenaltyTextRect(cpPanelY + 202),
		"關閉": cpCloseTextRect(),
	} {
		if r.x < cpPanelX || r.y < cpPanelY || r.x+r.w > cpPanelX+cpPanelW || r.y+r.h > cpPanelY+cpPanelH {
			t.Fatalf("%s 安全框超出面板：%+v", name, r)
		}
	}
	if cpLabelTextRect(cpPanelY+56).x+cpLabelTextRect(cpPanelY+56).w > cpValueTextRect(cpPanelY+56).x {
		t.Fatal("列標與數值安全框重疊")
	}
}

func TestCommandPointsLongestExternalTextIsWidthBounded(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		values := []struct {
			r    textSafeRect
			v    string
			size float64
		}{
			{cpTitleTextRect(), uiText(lang, "commandpoints.title"), 15},
			{cpLabelTextRect(cpPanelY + 56), uiText(lang, "commandpoints.label.starting"), 12},
			{cpPenaltyTextRect(cpPanelY + 202), fmt.Sprintf(uiText(lang, "commandpoints.penalty"), 9999, 99990), 11},
			{cpCloseTextRect(), uiText(lang, "commandpoints.close"), 11},
		}
		for _, tc := range values {
			w, _ := fnt.Measure(tc.r.clipped(fnt, tc.v, tc.size), tc.size)
			if w > tc.r.contentWidth() {
				t.Fatalf("%q 裁切後超出安全框", tc.v)
			}
		}
	}
}

func TestCommandPointsSourceHasNoEmbeddedPlayerTextOrDirectDraw(t *testing.T) {
	raw, err := os.ReadFile("commandpoints.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, forbidden := range []string{".tr(", ".fnt.Draw(", ".fnt.DrawCentered("} {
		if strings.Contains(src, forbidden) {
			t.Errorf("commandpoints.go 不得出現 %s", forbidden)
		}
	}
	if !strings.Contains(src, "gamedata.IncomeCommandOverflowCostPerPoint") {
		t.Error("超額懲罰文案必須與引擎共用每點 BC 契約")
	}
	for _, value := range []string{"COMMAND POINTS", "Starting Command Points", "軌道基地提供", "已使用", "點一下關閉"} {
		if strings.Contains(src, `"`+value+`"`) {
			t.Errorf("commandpoints.go 仍內嵌玩家文案 %q", value)
		}
	}
}
