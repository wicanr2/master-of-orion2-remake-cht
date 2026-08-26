package main

import (
	"fmt"
	"strings"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func treatySummaryPartText(lang i18n.Lang, part shell.TreatySummaryPart) string {
	key := "diplomacy.summary." + string(part.Kind)
	template := uiText(lang, key)
	if template == "" || template == key {
		template = uiText(lang, "diplomacy.summary.unknown")
	}
	switch part.Kind {
	case shell.TreatySummaryPayingTribute, shell.TreatySummaryReceivingTribute:
		return fmt.Sprintf(template, part.Percent)
	case shell.TreatySummaryTrade, shell.TreatySummaryResearch:
		return fmt.Sprintf(template, part.Turns, part.Value)
	case shell.TreatySummarySpecialFood, shell.TreatySummarySpecialResearch, shell.TreatySummarySpecialUnknown:
		return fmt.Sprintf(template, part.Turns)
	default:
		return template
	}
}

func treatySummaryText(lang i18n.Lang, state shell.TreatyState) string {
	parts := shell.TreatySummaryParts(state)
	texts := make([]string, 0, len(parts))
	for _, part := range parts {
		texts = append(texts, treatySummaryPartText(lang, part))
	}
	return strings.Join(texts, uiText(lang, "list.separator"))
}
