package main

import (
	"sync"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
)

var (
	uiCatalogOnce sync.Once
	uiCatalogZH   *i18n.Catalog
)

// uiText 以穩定語意鍵讀取自繪 remake 畫面的外部中英文文案。原版資料表仍走各自的
// per-source Translate；這個 catalog 只承載原版美術替換字與 remake 新增控制項。
func uiText(lang i18n.Lang, key string) string {
	uiCatalogOnce.Do(func() {
		uiCatalogZH = i18n.New(i18n.Traditional)
		if f, err := OpenI18NJSON("ui.json"); err == nil {
			_, _ = uiCatalogZH.LoadJSON(f)
			f.Close()
		}
	})
	return uiCatalogZH.TextFor(lang, key)
}
