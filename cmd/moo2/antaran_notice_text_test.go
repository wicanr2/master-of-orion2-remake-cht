package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestAntaranNoticeCatalogCoverage(t *testing.T) {
	keys := []string{
		"antaran.notice.launched", "antaran.notice.ai_engaged", "antaran.notice.undefended",
		"antaran.notice.battle.repelled", "antaran.notice.battle.not_repelled", "antaran.notice.unknown",
	}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		for _, key := range keys {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("語系 %v 缺少安塔蘭外部文案 %q", lang, key)
			}
		}
	}
}

func TestAntaranNoticeTextUsesTypedResultsAndCurrentLanguage(t *testing.T) {
	cases := []struct {
		notice shell.AntaranNotice
		wantZH string
		wantEN string
	}{
		{shell.AntaranNotice{Kind: shell.AntaranNoticeLaunched, StarName: "測試星", StarNameEN: "Test", ETA: 3}, "3 回合", "about 3 turns"},
		{shell.AntaranNotice{Kind: shell.AntaranNoticeAIEngaged, StarName: "測試星", StarNameEN: "Test"}, "當地守軍", "local defenders"},
		{shell.AntaranNotice{Kind: shell.AntaranNoticeUndefended, StarName: "測試星", StarNameEN: "Test"}, "未設防", "undefended"},
		{shell.AntaranNotice{Kind: shell.AntaranNoticeBattle, StarName: "測試星", StarNameEN: "Test", ShipsLost: 2, Repelled: true}, "已擊退", "was repelled"},
		{shell.AntaranNotice{Kind: shell.AntaranNoticeBattle, StarName: "測試星", StarNameEN: "Test", ShipsLost: 4}, "仍佔優勢", "still has the advantage"},
	}
	for _, tc := range cases {
		zh := antaranNoticeText(i18n.Traditional, &tc.notice)
		en := antaranNoticeText(i18n.English, &tc.notice)
		if !strings.Contains(zh, tc.wantZH) || !strings.Contains(en, tc.wantEN) {
			t.Errorf("安塔蘭通知分支錯誤：zh=%q en=%q", zh, en)
		}
		if !strings.Contains(en, "Test") || strings.Contains(en, "測試星") {
			t.Errorf("英文通知未使用英文星名：%q", en)
		}
	}
	if got := antaranNoticeText(i18n.Traditional, nil); got != "" {
		t.Errorf("nil 通知應回空字串：%q", got)
	}
}

func TestAntaranRuleSourceHasNoEmbeddedPlayerSentences(t *testing.T) {
	raw, err := os.ReadFile("../../internal/shell/antaran_invasion.go")
	if err != nil {
		t.Fatal(err)
	}
	src := strings.ToLower(string(raw))
	for _, forbidden := range []string{"安塔蘭艦隊已朝", "安塔蘭艦隊抵達", "我方損失", "an antaran fleet", "battle with the antarans"} {
		if strings.Contains(src, strings.ToLower(forbidden)) {
			t.Errorf("規則層仍內嵌安塔蘭玩家句子 %q", forbidden)
		}
	}
}
