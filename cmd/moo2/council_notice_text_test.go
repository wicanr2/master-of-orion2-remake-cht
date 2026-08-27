package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestCouncilNoticeCatalogCoverage(t *testing.T) {
	keys := []string{"common.you", "council.notice.insufficient", "council.notice.vote_requested",
		"council.notice.player_elected", "council.notice.enemy_elected", "council.notice.no_majority",
		"council.notice.unknown"}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		for _, key := range keys {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("語系 %v 缺少議會通知外部文案 %q", lang, key)
			}
		}
	}
}

func TestCouncilNoticeTextCoversTypedKinds(t *testing.T) {
	s := shell.NewDemoSession()
	base := shell.CouncilNotice{Meeting: 3, CandidateIdx: [2]int{-1, 0},
		CandidateName: [2]string{"你", "AI（人類）"}, Votes: [2]int{7, 5}, TotalVotes: 12}
	cases := []struct {
		kind shell.CouncilNoticeKind
		zh   string
		en   string
	}{
		{shell.CouncilNoticeInsufficientCandidates, "候選人不足", "fewer than two candidates"},
		{shell.CouncilNoticeVoteRequested, "請投票", "vote for"},
		{shell.CouncilNoticePlayerElected, "當選銀河領袖", "you were elected"},
		{shell.CouncilNoticeEnemyElectedPending, "接受或拒絕", "accept or reject"},
		{shell.CouncilNoticeNoMajority, "皆未達", "fell short"},
	}
	for _, tc := range cases {
		n := base
		n.Kind = tc.kind
		if tc.kind == shell.CouncilNoticeEnemyElectedPending {
			n.WinnerSlot = 1
		}
		zh := councilNoticeText(i18n.Traditional, s, &n)
		en := councilNoticeText(i18n.English, s, &n)
		if !strings.Contains(zh, tc.zh) || !strings.Contains(en, tc.en) {
			t.Errorf("kind=%v typed notice 格式錯誤：zh=%q en=%q", tc.kind, zh, en)
		}
		if strings.Contains(en, "人類") || strings.Contains(en, "候選") {
			t.Errorf("英文議會通知洩漏繁中名稱／句型：%q", en)
		}
	}
	twoAI := base
	twoAI.Kind = shell.CouncilNoticeEnemyElectedPending
	twoAI.CandidateIdx = [2]int{20, 21}
	twoAI.CandidateName = [2]string{"Wrong Candidate", "Correct Candidate"}
	twoAI.WinnerSlot = 1
	twoAI.Votes = [2]int{3, 9}
	if got := councilNoticeText(i18n.English, nil, &twoAI); !strings.Contains(got, "Correct Candidate") || !strings.Contains(got, "9/12") {
		t.Errorf("兩名 AI 候選時未依 WinnerSlot 顯示當選者：%q", got)
	}
	if councilNoticeText(i18n.Traditional, s, nil) != "" {
		t.Error("nil 議會通知應回空字串")
	}
	unknown := base
	unknown.Kind = 255
	if !strings.Contains(councilNoticeText(i18n.English, s, &unknown), "unavailable") {
		t.Error("未知議會通知種類沒有安全 fallback")
	}
}

func TestCouncilRuleSourceHasNoEmbeddedNoticeSentences(t *testing.T) {
	raw, err := os.ReadFile("../../internal/shell/council.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, forbidden := range []string{"LastCouncil ", "候選人不足,流會", "請投票給 %s", "票當選銀河領袖", "皆未達2/3"} {
		if strings.Contains(src, forbidden) {
			t.Errorf("議會規則層仍保存成品通知 %q", forbidden)
		}
	}
}
