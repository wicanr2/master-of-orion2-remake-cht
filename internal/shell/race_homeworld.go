package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// raceFoodClimate 回傳種族實際用來查食物產出的氣候。
// 手冊 p.23 明定水生族把 Tundra→Terran、Swamp→Ocean、Terran→Gaia;
// 其餘氣候維持原值。這是「食物產出」對映,不改 Planet 上供玩家查看的原始氣候。
func raceFoodClimate(climate gamedata.PlanetClimate, aquatic bool) gamedata.PlanetClimate {
	if !aquatic {
		return climate
	}
	switch climate {
	case gamedata.TUNDRA:
		return gamedata.TERRAN
	case gamedata.SWAMP:
		return gamedata.OCEAN
	case gamedata.TERRAN:
		return gamedata.GAIA
	default:
		return climate
	}
}

// racePopulationClimate 回傳種族實際用來查人口上限的氣候。
// 環境耐受族的手冊規則是「所有環境視為 Terran」；水生族則使用自己的水生對映。
// 兩者同時存在時,環境耐受對人口上限的規則優先,因為它明確涵蓋所有環境。
func racePopulationClimate(climate gamedata.PlanetClimate, aquatic, tolerant bool) gamedata.PlanetClimate {
	if tolerant {
		return gamedata.TERRAN
	}
	return raceFoodClimate(climate, aquatic)
}

// racePopulationMax 集中處理建立殖民地時的種族人口上限修正。
// Subterranean 的 +2×星球大小是手冊 p.24 的明文；其餘環境值由 gamedata 查表。
func racePopulationMax(size gamedata.PlanetSize, climate gamedata.PlanetClimate, aquatic, tolerant, subterranean bool) int {
	max := gamedata.PlanetBasePopMax(size, racePopulationClimate(climate, aquatic, tolerant))
	if subterranean {
		max += 2 * (int(size) + 1)
	}
	return max
}

// applyHomeworldRaceTraits 把只屬於母星的種族能力寫入殖民地與行星資料。
// r 的數值型 picks 已由呼叫端先套到殖民地,但母星環境能力需要改寫「基礎行星值」,
// 因此這裡以 r 的三個產出加成重新組合,避免在重複套用時把種族加成漏掉。
func applyHomeworldRaceTraits(c *engine.ColonyState, p *Planet, r Race, has func(gamedata.RaceTrait) bool) {
	if c == nil {
		return
	}
	aquatic := has(gamedata.TRAIT_AQUATIC)
	tolerant := has(gamedata.TRAIT_TOLERANT)
	subterranean := has(gamedata.TRAIT_SUBTERRANEAN)
	lithovore := has(gamedata.TRAIT_LITHOVORE)
	cybernetic := has(gamedata.TRAIT_CYBERNETIC)

	size := gamedata.MEDIUM_PLANET
	if has(gamedata.TRAIT_LARGE_HOMEWORLD) {
		size = gamedata.LARGE_PLANET
	}
	mineral := gamedata.ABUNDANT
	switch {
	case has(gamedata.TRAIT_RICH_HOMEWORLD):
		mineral = gamedata.RICH
	case has(gamedata.TRAIT_POOR_HOMEWORLD):
		mineral = gamedata.POOR
	}
	// 母星生成器已把玩家母星固定成 Terran；水生特性在食物／人口公式中把 Terran
	// 視同 Gaia,不在這裡把畫面上的母星原始氣候改成 Ocean。這保留原版母星行星資料的
	// 單一來源,也讓「水生對映」只作用於規則而不竄改玩家看到的環境名稱。
	climate := gamedata.TERRAN

	c.PlanetSize = size
	c.MineralRichness = mineral
	c.Climate = climate
	c.Aquatic = aquatic
	c.Subterranean = subterranean
	c.TolerantRace = tolerant
	c.Lithovore = lithovore
	c.Cybernetic = cybernetic
	c.FoodPerFarmer = gamedata.ClimateFoodPerFarmer(raceFoodClimate(climate, aquatic)) + r.FoodBonus
	c.IndustryPerWorker = gamedata.MineralIndustryPerWorker(mineral) + r.IndBonus
	c.ResearchPerScientist = gamedata.ResearchPerScientistNorm + r.ResBonus
	if has(gamedata.TRAIT_ARTIFACTS_HOMEWORLD) {
		// 手冊 p.24:母星每名科學家額外 +2,不是把一般星球 Ancient Artifacts
		// 的「每人固定 5」誤套成整個種族的科研基準。
		c.ResearchPerScientist += 2
	}
	c.PopMax = racePopulationMax(size, climate, aquatic, tolerant, subterranean)
	if c.PopMax < c.Population {
		c.PopMax = c.Population
	}

	if p == nil {
		return
	}
	p.ClimateID = climate
	p.SizeID = size
	p.MineralID = mineral
	p.Climate = climateDisplayName(climate)
	p.Size = sizeDisplayName(size)
	p.Mineral = mineralDisplayName(mineral)
}
