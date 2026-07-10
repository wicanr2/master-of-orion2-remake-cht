package main

// topicname.go:研究主題名的中文化 helper(顯示層)。
//
// 架構:internal/shell 不 import i18n,故 shell.ResearchTopicName 回英文名
// (= gamedata.TopicEnglishName,也就是 tech.tsv 的 i18n key);由此 cmd 層 helper
// 經 tech.tsv catalog 翻成中文。用一份 per-lang 惰性載入的共用 catalog,讓 galaxy /
// 回合摘要 / 研究抉擇三個顯示端都能翻,不必各自 os.Open tech.tsv、也不必到處傳 catalog。

import (
	"os"
	"sync"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
)

var (
	techCatOnce sync.Once
	techCatZh   *i18n.Catalog // 繁中 tech.tsv catalog(惰性載入一次)
)

func techCatalog(lang i18n.Lang) *i18n.Catalog {
	techCatOnce.Do(func() {
		techCatZh = i18n.New(i18n.Traditional)
		if f, err := os.Open("assets/i18n/tech.tsv"); err == nil {
			_, _ = techCatZh.LoadTSV(f)
			f.Close()
		}
	})
	// 依當前語言回傳:英文模式直接回一個 English catalog(Translate 回原字串)。
	if lang == i18n.English {
		return i18n.New(i18n.English)
	}
	return techCatZh
}

// topicNameZh 回傳研究主題的顯示名:繁中模式查 tech.tsv 得中文,英文/查無回英文名。
func topicNameZh(lang i18n.Lang, t gamedata.ResearchTopic) string {
	en := gamedata.TopicEnglishName(t)
	return techCatalog(lang).Translate(en)
}
