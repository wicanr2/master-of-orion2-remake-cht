package main

import (
	"fmt"
	"strings"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// aiRaidNoticeText 將型別化 AI 突襲結果套入外部文案；規則層不保存成品句子。
func aiRaidNoticeText(lang i18n.Lang, report *shell.AIRaidReport) string {
	if report == nil {
		return ""
	}
	aiName, starName := report.AIName, report.StarName
	if lang == i18n.English {
		if report.AINameEN != "" {
			aiName = report.AINameEN
		}
		if report.StarNameEN != "" {
			starName = report.StarNameEN
		}
	}
	if aiName == "" {
		aiName = uiText(lang, "common.unknown_empire")
	}
	if starName == "" {
		starName = uiText(lang, "common.unknown")
	}
	if report.Repelled {
		return fmt.Sprintf(uiText(lang, "raid.notice.repelled"), aiName, starName, report.FleetLost)
	}
	parts := []string{fmt.Sprintf(uiText(lang, "raid.notice.breakthrough"),
		aiName, starName, report.PopLost, report.BCLost)}
	if report.Building != "" {
		parts = append(parts, fmt.Sprintf(uiText(lang, "raid.notice.building_destroyed"),
			colonyBuildingLabel(lang, report.Building)))
	}
	if report.FleetLost > 0 {
		parts = append(parts, fmt.Sprintf(uiText(lang, "raid.notice.attacker_attrition"), report.FleetLost))
	}
	return strings.Join(parts, uiText(lang, "raid.notice.detail_separator"))
}
