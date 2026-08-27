package main

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func eventReportEmpireName(lang i18n.Lang, session *shell.GameSession, kind string, index int, fallback string) string {
	if kind == "ai" && session != nil && index >= 0 && index < len(session.AIPlayers) {
		race := session.AIPlayers[index].RaceIndex
		if race >= 0 && race < gamedata.OrigRaceCount {
			if lang == i18n.English {
				return gamedata.OrigRaceEnglishNames[race]
			}
			return gamedata.OrigRaceChineseNames[race]
		}
	}
	if kind == "player" && session != nil && session.PlayerName != "" {
		return session.PlayerName
	}
	if kind == "seat" && session != nil {
		if name := session.SeatName(index); name != "" {
			return name
		}
	}
	if fallback != "" {
		return fallback
	}
	if kind == "player" {
		return uiText(lang, "common.you")
	}
	return uiText(lang, "common.unknown_empire")
}

// eventReportMessageText 是事件畫面、回合摘要與 INFO 摘要的共同訊息入口。
// 尚未型別化的事件保留舊 Message／MessageEN 相容路徑。
func eventReportMessageText(lang i18n.Lang, session *shell.GameSession, report *shell.EventReport) string {
	if report == nil {
		return ""
	}
	if report.EventID == 34 && (report.TargetKind != "" || report.SecondaryTargetKind != "" ||
		report.TargetName != "" || report.SecondaryTargetName != "") {
		from := eventReportEmpireName(lang, session, report.TargetKind, report.TargetIndex, report.TargetName)
		to := eventReportEmpireName(lang, session, report.SecondaryTargetKind, report.SecondaryTargetIndex,
			report.SecondaryTargetName)
		return fmt.Sprintf(uiText(lang, "event.status.surrender"), from, to)
	}
	if lang == i18n.English && report.MessageEN != "" {
		return report.MessageEN
	}
	return report.Message
}
