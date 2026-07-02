package engine

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// 每人口單位每回合消耗 1 食物(MOO2 標準)。
const foodPerPopulation = 1

// colonyFood 計算農業產出、消耗與盈餘。
func colonyFood(cs ColonyState) (food, consumed, surplus int) {
	food = cs.Farmers * cs.FoodPerFarmer
	consumed = cs.Population * foodPerPopulation
	return food, consumed, food - consumed
}

// colonyPollution 依毛工業產出計算污染清理成本與淨工業。
// 順序(對照 production.go 註解建議):eighths → 產污產能 → 清理成本 → 淨工業。
func colonyPollution(cs ColonyState, grossIndustry int) (pollutingProd, cleanupCost, netIndustry int) {
	tolerance := gamedata.PollutionTolerance(cs.PlanetSize)
	eighths := gamedata.PollutionEighths(cs.PollutionProcessor, cs.AtmosphericRenewer, cs.CoreWasteDump)
	pollutingProd = gamedata.PollutionPollutingProduction(grossIndustry, eighths)
	cleanupCost = gamedata.PollutionCleanupCost(pollutingProd, tolerance, cs.TolerantRace)
	return pollutingProd, cleanupCost, grossIndustry - cleanupCost
}

// colonyGrowth 依人口成長公式計算本回合成長。住房獎金 h 於 Housing 配置時併入。
// 饑荒(食物盈餘 < 0)時不套用成長公式(成長須有食物支撐),回 0 並由呼叫端標 Starving。
func colonyGrowth(cs ColonyState, foodSurplus, netIndustry int) int {
	if foodSurplus < 0 {
		return 0 // 饑荒:成長公式不適用(饑荒減員屬另一機制,待移植)
	}
	base := gamedata.ColonyBaseGrowth(cs.Population, cs.Population, cs.PopMax)
	bonus := cs.GrowthBonusSum
	if cs.Housing {
		bonus += gamedata.ColonyHousingBonus(netIndustry, cs.Population)
	}
	return gamedata.ColonyGrowth(base, bonus)
}

// RunColonyTurn 執行一個殖民地的一回合經濟結算,依 MOO2 順序:
// 食物 → 工業 → 污染(縮減淨工業)→ 研究 → 人口成長。
func RunColonyTurn(cs ColonyState) ColonyOutput {
	food, consumed, surplus := colonyFood(cs)
	gross := cs.Workers * cs.IndustryPerWorker
	pollutingProd, cleanupCost, netIndustry := colonyPollution(cs, gross)
	research := cs.Scientists * cs.ResearchPerScientist
	growth := colonyGrowth(cs, surplus, netIndustry)

	return ColonyOutput{
		Food:                 food,
		FoodConsumed:         consumed,
		FoodSurplus:          surplus,
		Starving:             surplus < 0,
		GrossIndustry:        gross,
		PollutingProduction:  pollutingProd,
		PollutionCleanupCost: cleanupCost,
		NetIndustry:          netIndustry,
		Research:             research,
		PopGrowth:            growth,
	}
}
