package main

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
)

// tacticalText 是格子戰術玩家文案的唯一入口。Go 只保留語意鍵與動態參數；
// 固定句子與格式模板一律由外部 UI 文案目錄提供。
func tacticalText(lang i18n.Lang, key string, args ...any) string {
	value := uiText(lang, key)
	if len(args) == 0 {
		return value
	}
	return fmt.Sprintf(value, args...)
}
