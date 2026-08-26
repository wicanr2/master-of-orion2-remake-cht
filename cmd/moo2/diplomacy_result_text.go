package main

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func diplomacyResultText(lang i18n.Lang, result shell.DiplomacyResult) string {
	if result.Code == "" {
		return ""
	}
	key := "diplomacy.response." + string(result.Code)
	template := uiText(lang, key)
	if template == "" || template == key {
		template = uiText(lang, "diplomacy.response.unknown")
	}
	switch result.Code {
	case shell.DiploResultCashInsufficient:
		return fmt.Sprintf(template, result.Enemy, result.Available, result.Amount)
	case shell.DiploResultCashAccepted:
		return fmt.Sprintf(template, result.Enemy, result.Amount)
	case shell.DiploResultTechAccepted, shell.DiploResultStarAccepted:
		return fmt.Sprintf(template, result.Enemy, result.Detail)
	default:
		return fmt.Sprintf(template, result.Enemy)
	}
}
