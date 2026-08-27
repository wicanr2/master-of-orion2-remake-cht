package main

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
)

var colonySummaryRowY = [...]int{47, 78, 109, 140, 171, 202, 233, 264, 295}

func colonySummaryText(lang i18n.Lang, key string, args ...any) string {
	template := uiText(lang, key)
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}

func colonySummaryColumnRect(row, column int) textSafeRect {
	x := [...]int{18, 104, 236, 376}
	w := [...]int{78, 118, 128, 128}
	return textSafeRect{x: x[column], y: colonySummaryRowY[row] - 15, w: w[column], h: 30, insetX: 3, insetY: 2}
}

func colonySummaryBuildRect(row int) textSafeRect {
	return textSafeRect{x: 510, y: colonySummaryRowY[row] - 15, w: 120, h: 30, insetX: 3, insetY: 2}
}

func colonySummaryEmpireRect(row int) textSafeRect {
	return textSafeRect{x: 518, y: 352 + row*17, w: 108, h: 17, insetX: 3}
}

func colonySummaryPlanetInfoRect(row int) textSafeRect {
	return textSafeRect{x: 10, y: 352 + row*17, w: 82, h: 17, insetX: 4}
}

func colonySummaryProductionRect(row int) textSafeRect {
	return textSafeRect{x: 102, y: 352 + row*17, w: 269, h: 17, insetX: 5}
}

func planetEnvironmentLabel(lang i18n.Lang, category string, index int) string {
	return uiText(lang, fmt.Sprintf("planet.environment.%s.%d", category, index))
}

func climateName(lang i18n.Lang, value gamedata.PlanetClimate) string {
	if int(value) < 0 || int(value) > 9 {
		return uiText(lang, "planet.environment.unknown")
	}
	return planetEnvironmentLabel(lang, "climate", int(value))
}

func gravityName(lang i18n.Lang, value gamedata.PlanetGravity) string {
	index := map[gamedata.PlanetGravity]int{gamedata.LOW_G: 0, gamedata.NORMAL_G: 1, gamedata.HEAVY_G: 2}
	if i, ok := index[value]; ok {
		return planetEnvironmentLabel(lang, "gravity", i)
	}
	return uiText(lang, "planet.environment.unknown")
}

func mineralsName(lang i18n.Lang, value gamedata.PlanetMinerals) string {
	if int(value) < 0 || int(value) > 4 {
		return uiText(lang, "planet.environment.unknown")
	}
	return planetEnvironmentLabel(lang, "minerals", int(value))
}

func planetSizeName(lang i18n.Lang, value gamedata.PlanetSize) string {
	if int(value) < 0 || int(value) > 4 {
		return uiText(lang, "planet.environment.unknown")
	}
	return planetEnvironmentLabel(lang, "size", int(value))
}
