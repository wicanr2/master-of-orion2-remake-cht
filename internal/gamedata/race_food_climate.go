package gamedata

// RaceFoodClimate 回傳水生人口計算食物時使用的氣候；不改變行星本身氣候。
// GAME_MANUAL.pdf p.23：Tundra→Terran、Swamp→Ocean、Terran→Gaia。
func RaceFoodClimate(climate PlanetClimate, aquatic bool) PlanetClimate {
	if !aquatic {
		return climate
	}
	switch climate {
	case TUNDRA:
		return TERRAN
	case SWAMP:
		return OCEAN
	case TERRAN:
		return GAIA
	default:
		return climate
	}
}
