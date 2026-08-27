package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestSurrenderEventTextUsesTypedTargets(t *testing.T) {
	s := shell.NewDemoSession()
	if len(s.AIPlayers) < 2 {
		t.Fatal("測試需要兩個 AI 帝國")
	}
	report := &shell.EventReport{EventID: 34, TargetKind: "ai", TargetIndex: 0, TargetName: "繁中投降者",
		SecondaryTargetKind: "ai", SecondaryTargetIndex: 1, SecondaryTargetName: "繁中接收者"}
	zh := eventReportMessageText(i18n.Traditional, s, report)
	en := eventReportMessageText(i18n.English, s, report)
	wantZH := gamedata.OrigRaceChineseNames[s.AIPlayers[0].RaceIndex]
	wantEN := gamedata.OrigRaceEnglishNames[s.AIPlayers[0].RaceIndex]
	if !strings.Contains(zh, wantZH) || !strings.Contains(zh, "無條件投降") {
		t.Errorf("AI→AI 繁中 typed 通知錯誤：%q", zh)
	}
	if !strings.Contains(en, wantEN) || !strings.Contains(en, "surrendered unconditionally") ||
		strings.Contains(en, "繁中") || strings.Contains(en, wantZH) {
		t.Errorf("AI→AI 英文 typed 通知洩漏繁中或缺資料：%q", en)
	}

	s.PlayerName = "Commander"
	report.SecondaryTargetKind, report.SecondaryTargetIndex, report.SecondaryTargetName = "player", 0, "玩家"
	if got := eventReportMessageText(i18n.English, s, report); !strings.Contains(got, "Commander") {
		t.Errorf("AI→玩家通知未使用玩家名稱：%q", got)
	}
	if s.SetupHotseat(2) == 2 {
		report.SecondaryTargetKind, report.SecondaryTargetIndex, report.SecondaryTargetName = "seat", 1, ""
		if got, want := eventReportMessageText(i18n.Traditional, s, report), s.SeatName(1); !strings.Contains(got, want) {
			t.Errorf("AI→熱座通知未使用席位名稱 %q：%q", want, got)
		}
	}
}

func TestSurrenderEventTextFallbacks(t *testing.T) {
	unknown := &shell.EventReport{EventID: 34, TargetKind: "ai", TargetIndex: 99,
		SecondaryTargetKind: "ai", SecondaryTargetIndex: 98}
	if got := eventReportMessageText(i18n.English, nil, unknown); strings.Count(got, "Unknown Empire") != 2 {
		t.Errorf("非法雙 target 沒有外部安全 fallback：%q", got)
	}
	legacy := &shell.EventReport{EventID: 34, Message: "舊通知", MessageEN: "Legacy notice"}
	if got := eventReportMessageText(i18n.English, nil, legacy); got != "Legacy notice" {
		t.Errorf("舊存檔事件 34 未走 MessageEN fallback：%q", got)
	}
	if eventReportMessageText(i18n.Traditional, nil, nil) != "" {
		t.Error("nil event report 應回空字串")
	}
}

func TestSurrenderNoticeCatalogAndSourceBoundary(t *testing.T) {
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		if got := uiText(lang, "event.status.surrender"); got == "" || got == "event.status.surrender" {
			t.Errorf("語系 %v 缺少事件 34 外部模板", lang)
		}
	}
	raw, err := os.ReadFile("../../internal/shell/empire_surrender.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"無條件投降", "surrendered unconditionally", "imperial absorption"} {
		if strings.Contains(string(raw), forbidden) {
			t.Errorf("投降規則層仍內嵌玩家通知 %q", forbidden)
		}
	}
	for _, file := range []string{"eventscreen.go", "interactive.go", "infosubscreens.go"} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(src), "LastEventReport.MessageEN") {
			t.Errorf("%s 仍繞過共同事件文字入口", file)
		}
	}
}
