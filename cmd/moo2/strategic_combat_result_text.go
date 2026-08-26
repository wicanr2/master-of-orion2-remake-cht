package main

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func typedRefusalText(lang i18n.Lang, prefix, code string) string {
	if code == "" {
		return ""
	}
	key := prefix + code
	text := uiText(lang, key)
	if text == key {
		return uiText(lang, prefix+"unknown")
	}
	return text
}

func bombardmentRefusalText(lang i18n.Lang, result shell.GroundBombardResult) string {
	return typedRefusalText(lang, "galaxy.bombard.refusal.", result.Reason.String())
}

func invasionRefusalText(lang i18n.Lang, result shell.GroundInvasionResult) string {
	return typedRefusalText(lang, "galaxy.invasion.refusal.", result.Reason.String())
}

func mindControlRefusalText(lang i18n.Lang, result shell.GroundInvasionResult) string {
	return typedRefusalText(lang, "galaxy.mind_control.refusal.", result.Reason.String())
}

func monsterCombatRefusalText(lang i18n.Lang, code shell.MonsterCombatRefusalCode) string {
	return typedRefusalText(lang, "galaxy.monster_combat.refusal.", code.String())
}
