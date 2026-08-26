package main

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func planetTypeText(lang i18n.Lang, kind gamedata.PlanetType) string {
	return uiText(lang, fmt.Sprintf("planet.type.%d", kind))
}

func colonizationRefusalText(lang i18n.Lang, result shell.ColonizationResult) string {
	key := "galaxy.colonization.refusal." + result.Reason.String()
	switch result.Reason {
	case shell.ColonizeRequiresOutpost:
		return fmt.Sprintf(uiText(lang, key), planetTypeText(lang, result.PlanetType))
	case shell.ColonizeMonsterBlocked:
		return fmt.Sprintf(uiText(lang, key), uiText(lang, fmt.Sprintf("monster.name.%d", result.MonsterKind)))
	case "":
		return ""
	default:
		text := uiText(lang, key)
		if text == key {
			return uiText(lang, "galaxy.colonization.refusal.unknown")
		}
		return text
	}
}

func outpostRefusalText(lang i18n.Lang, result shell.OutpostResult) string {
	key := "galaxy.outpost.refusal." + result.Reason.String()
	if result.Reason == shell.OutpostMonsterBlocked {
		return fmt.Sprintf(uiText(lang, key), uiText(lang, fmt.Sprintf("monster.name.%d", result.MonsterKind)))
	}
	if result.Reason == "" {
		return ""
	}
	text := uiText(lang, key)
	if text == key {
		return uiText(lang, "galaxy.outpost.refusal.unknown")
	}
	return text
}
