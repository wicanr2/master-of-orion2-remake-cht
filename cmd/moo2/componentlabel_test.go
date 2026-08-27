package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// hasHanIn 回報字串裡有沒有漢字(重用 lang_gap_test.go 的判斷)。
func hasHanIn(s string) bool { return hasHan(s) }

// TestComponentLabelEnglishHasNoHan 英文模式下**每一個**元件名都不能還是中文。
//
// ⚠ 這條測試是從截圖抓出來的:艦艇設計畫面英文模式顯示「Weapon: 雷射」「Armor: 無裝甲」
// ——元件表存的是中文,而且它同時是查表 key,不能直接換成英文。
// 英文顯示名改成**從 `UnlockTech` 推導原版科技名**,這條測試盯著推導有沒有漏。
func TestComponentLabelEnglishHasNoHan(t *testing.T) {
	sets := map[string][]shell.Component{
		"武器": shell.WeaponOptions,
		"裝甲": shell.ArmorOptions,
		"護盾": shell.ShieldOptions,
		"特殊": shell.SpecialOptions,
	}
	for what, opts := range sets {
		for _, c := range opts {
			got := componentLabel(i18n.English, c)
			if hasHanIn(got) {
				t.Errorf("%s「%s」的英文顯示名還是中文:%q(UnlockTech=%d)", what, c.Name, got, c.UnlockTech)
			}
		}
	}
}

func TestComponentNoTechLabelsComeFromExternalCatalog(t *testing.T) {
	for _, key := range componentNoTechTextKey {
		for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("%s 缺少語言 %v 的外部文案", key, lang)
			}
		}
	}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		got := componentLabel(lang, shell.Component{Name: "未登記規則鍵"})
		if want := uiText(lang, "component.label.unknown"); got != want {
			t.Errorf("未知元件 fallback=%q，want %q", got, want)
		}
	}
}

func TestComponentTechLabelsComeFromExternalTechCatalog(t *testing.T) {
	sets := [][]shell.Component{shell.WeaponOptions, shell.ArmorOptions, shell.ShieldOptions, shell.SpecialOptions}
	for _, opts := range sets {
		for _, c := range opts {
			if c.UnlockTech == gamedata.TECH_NONE {
				continue
			}
			en := gamedata.TechnologyName(c.UnlockTech)
			if !techCatalog(i18n.Traditional).Has(en) {
				t.Errorf("%q 的原版科技鍵 %q 在 tech.json 缺少繁中譯文", c.Name, en)
			}
			for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
				want := techCatalog(lang).Translate(en)
				if got := componentLabel(lang, c); got != want {
					t.Errorf("%q 語言 %v=%q，want 外部科技目錄 %q", c.Name, lang, got, want)
				}
			}
		}
	}
}

func TestComponentLabelSourceDoesNotEmbedNoTechDisplayText(t *testing.T) {
	src, err := os.ReadFile("componentlabel.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"No Armor", "No Shield", "Unarmed", "Battle Computer", "Regeneration", "return c.Name"} {
		if strings.Contains(string(src), forbidden) {
			t.Errorf("componentlabel.go 仍內嵌玩家文案 %q", forbidden)
		}
	}
}

// TestComponentNoTechChineseLabelsRemainCompatible 確認外部繁中目錄仍呈現既有無科技名稱；
// 規則鍵本身不會因顯示層外部化而修改。
func TestComponentNoTechChineseLabelsRemainCompatible(t *testing.T) {
	for _, opts := range [][]shell.Component{shell.WeaponOptions, shell.ArmorOptions, shell.ShieldOptions, shell.SpecialOptions} {
		for _, c := range opts {
			if c.UnlockTech != gamedata.TECH_NONE {
				continue
			}
			if got := componentLabel(i18n.Traditional, c); got != c.Name {
				t.Errorf("中文模式應原樣回傳:%q → %q", c.Name, got)
			}
		}
	}
}
