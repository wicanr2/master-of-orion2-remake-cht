package main

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// componentlabel.go:元件名的英文顯示名(第 85 項(元件名英文))。
//
// ============ 問題 ============
//
// `shell.Component.Name` 是**中文**,而且它同時是**查表 key**(元件比對、存檔、
// `specialDeviceByName` 對照表全都用它)。所以不能像星名/艦名那樣把資料換成英文
// ——那會動到十幾處比對。
//
// ============ 作法:從原版科技名推導,不手寫第二份表 ============
//
// 每個元件都帶著 `UnlockTech`,而 `gamedata.TechnologyNames` 是**原版執行檔的科技名表**。
// 所以英文顯示名 = 那個科技的原版英文名,一行推導,不必再維護一份 77 列的對照
// ——**兩份表遲早會漂移**,這是本專案反覆踩過的形狀。
//
// 只有 `UnlockTech == 0` 的幾項要手寫(那幾項在原版沒有對應科技:三個「無」是 UI 佔位,
// 戰鬥電腦在原版是獨立槽、重生程序是種族特性)。

// componentNoTechEnglish 是沒有解鎖科技的元件的英文名。
//
// ⚠ 這張表**只該有這五筆**。新增元件時如果又要往這裡加,先問「它的 UnlockTech 為什麼是 0」
// ——多半是漏填,而不是真的沒有對應科技。
var componentNoTechEnglish = map[string]string{
	"無":    "None",
	"無裝甲":  "No Armor",
	"無護盾":  "No Shield",
	"無武裝":  "Unarmed",
	"戰鬥電腦": "Battle Computer",
	"重生程序": "Regeneration",
}

// componentLabel 回傳元件在指定語言下的顯示名。中文模式直接回 `Name`(它本來就是中文)。
func componentLabel(lang i18n.Lang, c shell.Component) string {
	if lang != i18n.English {
		return c.Name
	}
	if en, ok := componentNoTechEnglish[c.Name]; ok {
		return en
	}
	if c.UnlockTech != gamedata.TECH_NONE {
		return gamedata.TechnologyName(c.UnlockTech)
	}
	return c.Name // 沒有科技也不在上表:退回中文(至少看得到東西),並由測試盯著
}
