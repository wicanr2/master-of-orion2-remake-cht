package main

import (
	"testing"

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

// TestComponentLabelChineseUnchanged 中文模式必須原樣回傳 —— 它同時是查表 key,
// 動了就會有比對失效。
func TestComponentLabelChineseUnchanged(t *testing.T) {
	for _, c := range shell.SpecialOptions {
		if got := componentLabel(i18n.Traditional, c); got != c.Name {
			t.Errorf("中文模式應原樣回傳:%q → %q", c.Name, got)
		}
	}
}
