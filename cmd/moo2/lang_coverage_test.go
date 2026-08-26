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

// TestRaceEntriesHaveEnglish:種族清單每一列都要有英文複數名與單數形容詞形。
// enAdj 用在「Human Empire」這種場合,不是自動去尾 s(Alkari / Bulrathi /
// Mrrshan / Sakkra 本來就沒有 s)。
func TestRaceEntriesHaveEnglish(t *testing.T) {
	for i, e := range raceSelectList {
		if e.en == "" || e.enAdj == "" {
			t.Errorf("raceSelectList[%d] %q:英文欄未填(en=%q enAdj=%q)", i, e.zh, e.en, e.enAdj)
		}
		if cjk.MatchString(e.en) || cjk.MatchString(e.enAdj) {
			t.Errorf("raceSelectList[%d] %q:英文欄裡有中文(en=%q enAdj=%q)", i, e.zh, e.en, e.enAdj)
		}
	}
}

// TestRacesHaveEnglishDesc:能力說明的英文版。種族選擇畫面肖像下方那一行會用到。
func TestRacesHaveEnglishDesc(t *testing.T) {
	for i, r := range shell.Races {
		if r.EnName == "" || r.EnDesc == "" {
			t.Errorf("shell.Races[%d] %q:EnName/EnDesc 未填", i, r.Name)
		}
		if cjk.MatchString(r.EnDesc) {
			t.Errorf("shell.Races[%d] %q:EnDesc 裡有中文:%q", i, r.Name, r.EnDesc)
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
