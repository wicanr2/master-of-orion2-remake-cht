package main

import (
	"regexp"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// lang_coverage_test.go:英文模式的**資料面**護欄。
//
// 這裡不驗畫面(那要跑 -lang en 的過場截圖廊,見 docs/HONEST-STATUS.md)。
// 這裡只擋一件事:**新增一族 / 一個旗色時漏填英文欄**。漏了不會編譯失敗、
// 中文模式也完全正常,只有英文模式會突然出現一格空白或一段中文——
// 而那要跑截圖才看得到。

var cjk = regexp.MustCompile(`[\p{Han}]`)

// TestRaceEntriesHaveEnglish:種族清單只保存穩定鍵；英文複數名與單數形容詞來自外部 catalog。
func TestRaceEntriesHaveEnglish(t *testing.T) {
	for i, e := range raceSelectList {
		name := raceSelectEntryText(i18n.English, e, "name")
		adj := raceSelectEntryText(i18n.English, e, "adjective")
		if name == "" || adj == "" || name == e.key || adj == e.key {
			t.Errorf("raceSelectList[%d] %q：外部英文名稱／形容詞缺漏", i, e.key)
		}
		if cjk.MatchString(name) || cjk.MatchString(adj) {
			t.Errorf("raceSelectList[%d] %q：外部英文欄含中文(name=%q adjective=%q)", i, e.key, name, adj)
		}
	}
}

// TestFlagColorsHaveLocalizedNames:旗色資料只保留穩定鍵，玩家文案由 ui.json 提供。
func TestFlagColorsHaveLocalizedNames(t *testing.T) {
	for i, fc := range shell.FlagColors {
		key := "nameflag.color." + fc.Key
		if got := uiText(i18n.English, key); got == key || got == "" || cjk.MatchString(got) {
			t.Errorf("shell.FlagColors[%d] %q 的英文外部文案無效:%q", i, fc.Key, got)
		}
		if got := uiText(i18n.Traditional, key); got == key || got == "" {
			t.Errorf("shell.FlagColors[%d] %q 的中文外部文案無效:%q", i, fc.Key, got)
		}
	}
}
