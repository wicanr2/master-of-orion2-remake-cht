package main

import (
	"sync"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// nametranslator.go:把原版的英文專有名詞(星名 / 艦名)翻成當前語言。
//
// ============ 為什麼名稱池存英文 ============
//
// 那兩個池子(829 條星名、672 條艦名)先前存的是**中文**。中文模式當然沒問題,
// 而**英文模式的星圖上會出現中文星名**——那不是「漏翻」,是資料本身就是中文,
// 沒有任何 `tr()` 補得回來。這是英文模式最大的一塊缺口(1,501 條,佔 internal/ 中文
// 字面值的六成)。
//
// 改成英文原文之後,中文由這裡翻。**翻譯發生在取名當下**(銀河生成、造艦),
// 不是顯示當下——理由見 shell.GameSession.localName 的註解。
//
// ⚠ 這個翻譯器只吃**專有名詞**兩張表(starname-random / shipname),不吃一般 UI 譯表。
// 混在一起的話,像 "Wolf"(艦名)這種與 UI 詞彙撞字的 key 會被別張表搶走。

var (
	nameCatOnce sync.Once
	nameCat     *i18n.Catalog
)

// installNameTranslator 依語言把翻譯器裝進 internal/shell。
//
// 英文模式**不裝**(nil = 恆等 = 英文原文),這樣連查表都省了,也不會因為譯表裡
// 有某個 key 就把英文換成中文。
func installNameTranslator(lang i18n.Lang) {
	if lang == i18n.English {
		shell.SetNameTranslator(nil)
		return
	}
	shell.SetNameTranslator(func(en string) string { return properNounCatalog().Translate(en) })
}

// properNounCatalog 載入專有名詞譯表(只有這兩張)。載不到就回空 catalog
// ——`Translate` 對查不到的 key 回傳原字串,也就是退回英文,不會變成空白。
func properNounCatalog() *i18n.Catalog {
	nameCatOnce.Do(func() {
		nameCat = i18n.New(i18n.Traditional)
		for _, f := range []string{"starname-random.json", "shipname.json"} {
			if fh, err := OpenI18NJSON(f); err == nil {
				_, _ = nameCat.LoadJSON(fh)
				fh.Close()
			}
		}
	})
	return nameCat
}
