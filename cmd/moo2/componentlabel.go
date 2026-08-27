package main

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// componentlabel.go:元件規則鍵與雙語顯示名的隔離層(第 85 項(元件名英文))。
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
// 所以顯示名 = 那個科技的原版英文名再交給外部 tech.json 選語言,不必再維護一份 77 列的對照
// ——**兩份表遲早會漂移**,這是本專案反覆踩過的形狀。
//
// 只有 `UnlockTech == 0` 的幾項需要額外顯示路由；中英文顯示值都留在 ui.json。

// componentNoTechTextKey 將既有規則鍵路由到外部玩家文案；map 的值是語意鍵，不是顯示文字。
var componentNoTechTextKey = map[string]string{
	"無":    "component.label.none",
	"無裝甲":  "component.label.no_armor",
	"無護盾":  "component.label.no_shield",
	"無武裝":  "component.label.unarmed",
	"戰鬥電腦": "component.label.battle_computer",
	"重生程序": "component.label.regeneration",
}

// componentLabel 回傳元件在指定語言下的顯示名；規則鍵不得直接充當未知值的玩家文案。
func componentLabel(lang i18n.Lang, c shell.Component) string {
	if key, ok := componentNoTechTextKey[c.Name]; ok {
		return uiText(lang, key)
	}
	if c.UnlockTech != gamedata.TECH_NONE {
		return techCatalog(lang).Translate(gamedata.TechnologyName(c.UnlockTech))
	}
	return uiText(lang, "component.label.unknown")
}
