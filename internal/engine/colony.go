package engine

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// 每人口單位每回合消耗 1 食物(MOO2 標準)。
const foodPerPopulation = 1

// colonyFood 計算農業產出(經士氣調整)、消耗與盈餘。FlatFood(水耕農場/地底農場等殖民地
// 整體固定食物加成)與人數無關,故加在士氣調整之外——手冊沒有明講固定產出是否吃士氣加成,
// 這裡採「不吃」的保守假設(士氣是勞動效率調整,固定加成是設施自動產出,概念上分開),
// 屬 remake 建模選擇而非手冊逐字依據,若未來找到反證再調整。
func colonyFood(cs ColonyState) (food, consumed, surplus int) {
	food = gamedata.MoraleProductionOutput(cs.Farmers*cs.FoodPerFarmer, cs.MoralePercent) + cs.FlatFood
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
//
// FlatGrowth(複製中心 p.99)只在「尚未達人口上限」時併入,對齊手冊「直到達星球人口上限為止」
// ——一旦 Population>=PopMax,連同 base 一起归零,不會讓固定成長點數在封頂後繼續無意義累積
// (避免之後 PopMax 因生態圈等建築提高時,尘封已久的虛增點數瞬間兌現成不合理的暴增人口)。
func colonyGrowth(cs ColonyState, foodSurplus, netIndustry int) int {
	if foodSurplus < 0 {
		return 0 // 饑荒:成長公式不適用(饑荒減員屬另一機制,待移植)
	}
	base := gamedata.ColonyBaseGrowth(cs.Population, cs.Population, cs.PopMax)
	bonus := cs.GrowthBonusSum
	if cs.Housing {
		bonus += gamedata.ColonyHousingBonus(netIndustry, cs.Population)
	}
	growth := gamedata.ColonyGrowth(base, bonus)
	if cs.Population < cs.PopMax {
		growth += cs.FlatGrowth
	}
	return growth
}

// RunColonyTurn 執行一個殖民地的一回合經濟結算,依 MOO2 順序:
// 食物 → 工業 → 污染(縮減淨工業)→ 研究 → 人口成長。
func RunColonyTurn(cs ColonyState) ColonyOutput {
	food, consumed, surplus := colonyFood(cs)
	// 工業與研究同樣經士氣調整(手冊:每格士氣 ±10% 套用於食物/工業/研究/收入)。FlatIndustry/
	// FlatResearch(殖民地整體固定加成,見 ColonyState 欄位註解)與士氣調整後的 per-worker
	// 產出分開相加,採與 colonyFood/FlatFood 同款「固定加成不吃士氣」假設。
	// FlatIndustry 在污染縮減之前併入 gross(依手冊,固定產能也算「殖民地產能」,一樣會產生
	// 污染,見下方 colonyPollution 以 gross 全額計算 pollutingProd/cleanupCost)。
	gross := gamedata.MoraleProductionOutput(cs.Workers*cs.IndustryPerWorker, cs.MoralePercent) + cs.FlatIndustry
	pollutingProd, cleanupCost, netIndustry := colonyPollution(cs, gross)
	research := gamedata.MoraleProductionOutput(cs.Scientists*cs.ResearchPerScientist, cs.MoralePercent) + cs.FlatResearch
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
