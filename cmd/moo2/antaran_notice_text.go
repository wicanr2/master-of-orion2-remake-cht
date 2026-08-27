package main

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// antaranNoticeText 將型別化安塔蘭結果套入外部玩家文案；規則層不保存語言句子。
func antaranNoticeText(lang i18n.Lang, notice *shell.AntaranNotice) string {
	if notice == nil {
		return ""
	}
	star := notice.StarName
	if lang == i18n.English && notice.StarNameEN != "" {
		star = notice.StarNameEN
	}
	switch notice.Kind {
	case shell.AntaranNoticeLaunched:
		return fmt.Sprintf(uiText(lang, "antaran.notice.launched"), star, notice.ETA)
	case shell.AntaranNoticeAIEngaged:
		return fmt.Sprintf(uiText(lang, "antaran.notice.ai_engaged"), star)
	case shell.AntaranNoticeUndefended:
		return fmt.Sprintf(uiText(lang, "antaran.notice.undefended"), star)
	case shell.AntaranNoticeBattle:
		key := "antaran.notice.battle.not_repelled"
		if notice.Repelled {
			key = "antaran.notice.battle.repelled"
		}
		return fmt.Sprintf(uiText(lang, key), star, notice.ShipsLost)
	default:
		return uiText(lang, "antaran.notice.unknown")
	}
}
