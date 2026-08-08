package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// TestNameTranslatorRoundTrip 名稱池存英文,而**中文模式的輸出必須與改動前完全相同**。
//
// 這條測的是那個前提:池子裡每一條英文都翻得出中文(沒有漏 key),
// 而且翻出來的不是英文原字串(那代表查表落空,`Translate` 會原樣回傳)。
//
// ⚠ 真正的證據是畫廊 34 張逐位元比對(第 84 項(名稱池雙語化)驗過:0 張不同)。
// 這條測試是**便宜的前哨**——譯表少一行就會在這裡紅,不必等跑完整個畫廊。
func TestNameTranslatorRoundTrip(t *testing.T) {
	installNameTranslator(i18n.Traditional)
	tr := shell.NameTranslatorForTest()
	if tr == nil {
		t.Fatal("中文模式應該裝上翻譯器")
	}
	pools := map[string][]string{
		"星名": shell.StarNamePoolForTest(),
		"艦名": shell.ShipNamePoolForTest(),
	}
	for what, pool := range pools {
		if len(pool) == 0 {
			t.Fatalf("%s池是空的", what)
		}
		untranslated := 0
		for _, en := range pool {
			if tr(en) == en {
				untranslated++
			}
		}
		if untranslated > 0 {
			t.Errorf("%s池有 %d 條查不到譯文(共 %d 條)——譯表少了 key?", what, untranslated, len(pool))
		}
	}
}

// TestNameTranslatorEnglishIsIdentity 英文模式**不裝**翻譯器:名稱維持英文原文。
//
// 這一條擋的是「英文模式也去查表」這種寫法——那會讓譯表裡剛好有的 key 被換成中文,
// 而且是隨譯表內容漂移的隨機行為。
func TestNameTranslatorEnglishIsIdentity(t *testing.T) {
	installNameTranslator(i18n.English)
	if shell.NameTranslatorForTest() != nil {
		t.Error("英文模式不該裝翻譯器(nil = 恆等 = 英文原文)")
	}
	installNameTranslator(i18n.Traditional) // 還原,免得影響同一輪的其他測試
}
