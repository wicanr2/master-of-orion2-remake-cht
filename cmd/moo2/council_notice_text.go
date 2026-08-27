package main

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func councilNoticeEmpireName(lang i18n.Lang, session *shell.GameSession, idx int, fallback string) string {
	if idx == -1 {
		return uiText(lang, "common.you")
	}
	if session != nil && idx >= 0 && idx < len(session.AIPlayers) {
		race := session.AIPlayers[idx].RaceIndex
		if race >= 0 && race < gamedata.OrigRaceCount {
			if lang == i18n.English {
				return gamedata.OrigRaceEnglishNames[race]
			}
			return gamedata.OrigRaceChineseNames[race]
		}
	}
	if fallback != "" {
		return fallback
	}
	return uiText(lang, "common.unknown_empire")
}

// councilNoticeText 將型別化議會結果套入外部文案；規則層不保存成品句子。
func councilNoticeText(lang i18n.Lang, session *shell.GameSession, notice *shell.CouncilNotice) string {
	if notice == nil {
		return ""
	}
	a := councilNoticeEmpireName(lang, session, notice.CandidateIdx[0], notice.CandidateName[0])
	b := councilNoticeEmpireName(lang, session, notice.CandidateIdx[1], notice.CandidateName[1])
	switch notice.Kind {
	case shell.CouncilNoticeInsufficientCandidates:
		return fmt.Sprintf(uiText(lang, "council.notice.insufficient"), notice.Meeting)
	case shell.CouncilNoticeVoteRequested:
		return fmt.Sprintf(uiText(lang, "council.notice.vote_requested"), notice.Meeting, a, b)
	case shell.CouncilNoticePlayerElected:
		winner := notice.WinnerSlot
		if winner < 0 || winner > 1 {
			winner = 0
		}
		votes := notice.Votes[winner]
		return fmt.Sprintf(uiText(lang, "council.notice.player_elected"), notice.Meeting, votes, notice.TotalVotes)
	case shell.CouncilNoticeEnemyElectedPending:
		winnerSlot := notice.WinnerSlot
		if winnerSlot < 0 || winnerSlot > 1 {
			winnerSlot = 0
		}
		winner := [2]string{a, b}[winnerSlot]
		votes := notice.Votes[winnerSlot]
		return fmt.Sprintf(uiText(lang, "council.notice.enemy_elected"), notice.Meeting, winner, votes, notice.TotalVotes)
	case shell.CouncilNoticeNoMajority:
		return fmt.Sprintf(uiText(lang, "council.notice.no_majority"), notice.Meeting,
			a, notice.Votes[0], b, notice.Votes[1])
	default:
		return uiText(lang, "council.notice.unknown")
	}
}
