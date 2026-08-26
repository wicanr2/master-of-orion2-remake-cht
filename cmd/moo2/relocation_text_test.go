package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestRelocationPlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"relocation.refusal.invalid_star",
		"relocation.refusal.black_hole_origin",
		"relocation.refusal.black_hole_target",
		"relocation.refusal.unexplored_origin",
		"relocation.refusal.unexplored_target",
		"relocation.refusal.monster_origin",
		"relocation.refusal.no_colony",
		"relocation.refusal.write_failed",
		"relocation.confirm.monster",
		"relocation.fallback.system",
		"relocation.fallback.monster",
		"relocation.prompt.star_panel_target",
		"relocation.prompt.origin_set",
		"relocation.prompt.retarget_all",
		"relocation.prompt.choose_origin",
		"relocation.result.no_existing",
		"relocation.result.retargeted_count",
		"relocation.result.cleared",
		"relocation.result.set",
		"relocation.result.none_to_clear",
		"relocation.result.cleared_count",
		"relocation.button.set",
		"relocation.button.current",
		"relocation.button.retarget_all",
		"relocation.button.clear_all",
		"fleet.antares.entry",
		"relocation.transition.galaxy",
		"relocation.transition.fleet",
		"list.separator",
	}
	for monster := 1; monster <= 6; monster++ {
		keys = append(keys, fmt.Sprintf("monster.name.%d", monster))
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("集結點流程缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestRelocationFormatsAndButtonsFitSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		formatted := []string{
			fmt.Sprintf(uiText(lang, "relocation.confirm.monster"), "A Very Long Star System Name", uiText(lang, "monster.name.2")),
			fmt.Sprintf(uiText(lang, "relocation.refusal.monster_origin"), uiText(lang, "monster.name.2")),
			fmt.Sprintf(uiText(lang, "relocation.refusal.no_colony"), "A Very Long Star System Name"),
			fmt.Sprintf(uiText(lang, "relocation.result.retargeted_count"), 99),
			fmt.Sprintf(uiText(lang, "relocation.result.cleared_count"), 99),
		}
		for _, value := range formatted {
			if strings.Contains(value, "%!") {
				t.Errorf("集結點格式模板參數不相容：%q", value)
			}
		}
		checkClippedTextFits(t, fnt, starPanelButtonTextRect(424),
			fmt.Sprintf(uiText(lang, "relocation.button.current"), "A Very Long Star System Name"), 12)
		checkClippedTextFits(t, fnt, relocationRetargetAllTextRect(), uiText(lang, "relocation.button.retarget_all"), 11)
		checkClippedTextFits(t, fnt, relocationClearAllTextRect(), uiText(lang, "relocation.button.clear_all"), 11)
		checkClippedTextFits(t, fnt, fleetAntaranEntryTextRect(), uiText(lang, "fleet.antares.entry"), 13)
	}
	if fleetAntaranEntryTextRect().y+fleetAntaranEntryTextRect().h > relocationRetargetAllTextRect().y {
		t.Fatal("安塔蘭入口提示不得壓到下方集結點入口")
	}

	for _, tc := range []struct {
		text textSafeRect
		x, y int
	}{
		{relocationRetargetAllTextRect(), 20, 412},
		{relocationClearAllTextRect(), 168, 412},
	} {
		if 2*tc.text.x+tc.text.w != 2*tc.x+140 || 2*tc.text.y+tc.text.h != 2*tc.y+18 {
			t.Errorf("集結點 adapter 按鈕文字框與 140×18 熱區中心不一致：%+v", tc.text)
		}
	}
}

func TestRelocationSourcesContainNoEmbeddedPlayerSentences(t *testing.T) {
	for _, path := range []string{"relocation.go", "../../internal/shell/relocation.go"} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		src := string(raw)
		if path == "relocation.go" && strings.Contains(src, ".tr(") {
			t.Fatal("cmd/moo2/relocation.go 不得再以 tr 內嵌雙語玩家文案")
		}
		for _, value := range []string{
			"No colony has a rally point", "沒有任何殖民地設過集結點",
			"Relocation set", "集結點已設定", "黑洞不能當遷移起點",
			"Ships sent there will be attacked", "送過去的艦艇會遭到攻擊",
		} {
			if strings.Contains(src, `"`+value) {
				t.Errorf("%s 仍內嵌集結點玩家文案 %q", path, value)
			}
		}
	}
}
