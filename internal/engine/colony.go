package engine

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// 每人口單位每回合消耗 1 食物(MOO2 標準)。
const foodPerPopulation = 1

// colonyGravityPenaltyPercent 回傳本殖民地目前生效的重力懲罰百分點(0 或負值,GAME_MANUAL.pdf
// p.58)。行星重力產生器(NormalizeGravity=true,p.104「正常化至 Normal-G,消除 Low-G/Heavy-G
// 負面效果」)一律強制歸零,不論 PlanetGravity 是什麼。
//
// 種族 Low-G/High-G 重力天賦尚未在 ColonyState 建模(見該欄位註解),固定以 gamedata.NORMAL_G
// 當「種族重力」基準傳給 GravityPenaltyPercent——這代表懲罰值只反映「行星重力」單一因子,
// 是 remake 建模簡化,非手冊逐字依據,待補種族天賦欄位後需回頭校正。
func colonyGravityPenaltyPercent(cs ColonyState) int {
	if cs.NormalizeGravity {
		return 0
	}
	return gamedata.GravityPenaltyPercent(cs.PlanetGravity, gamedata.NORMAL_G)
}

// colonyFood 計算農業產出(經士氣+重力調整)、消耗與盈餘。FlatFood(水耕農場/地底農場等殖民地
// 整體固定食物加成)與人數無關,故加在士氣/重力調整之外——手冊沒有明講固定產出是否吃士氣或
// 重力加成,這裡採「不吃」的保守假設(士氣/重力是勞動效率調整,固定加成是設施自動產出,概念上
// 分開),屬 remake 建模選擇而非手冊逐字依據,若未來找到反證再調整。
//
// 士氣與重力套用順序:兩者都是單純百分比調整(手冊沒有描述兩者相乘或有先後依存關係),
// 先加總成單一百分點(MoralePercent + 重力懲罰)再套一次 GravityAdjustedProduction,不分兩次
// 各自除法——避免兩次連續整數除法各自捨去造成的複合誤差(例如 100 先乘 0.75 捨去、再乘 1.1
// 捨去,結果會因套用順序不同而不同,但手冊沒有給任何「先重力後士氣」或反過來的根據)。這也與
// ColonyState 既有慣例一致:多個百分比/固定加成先加總,再套一次公式(GrowthBonusSum、
// IncomeBonusPercent 皆是同一模式)。
func colonyFood(cs ColonyState) (food, consumed, surplus int) {
	pct := cs.MoralePercent + colonyGravityPenaltyPercent(cs)
	food = gamedata.GravityAdjustedProduction(cs.Farmers*cs.FoodPerFarmer, pct) + cs.FlatFood
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
	// 工業與研究同樣經士氣+重力調整(手冊:每格士氣 ±10%、重力 -25%/-50% 套用於
	// 食物/工業/研究三者,p.58/p.63)。FlatIndustry/FlatResearch(殖民地整體固定加成,見
	// ColonyState 欄位註解)與調整後的 per-worker 產出分開相加,採與 colonyFood/FlatFood
	// 同款「固定加成不吃士氣/重力」假設。士氣與重力合併成單一百分點再套一次公式的理由見
	// colonyFood 註解(避免兩次連續整數除法的複合誤差)。
	// FlatIndustry 在污染縮減之前併入 gross(依手冊,固定產能也算「殖民地產能」,一樣會產生
	// 污染,見下方 colonyPollution 以 gross 全額計算 pollutingProd/cleanupCost)。
	pct := cs.MoralePercent + colonyGravityPenaltyPercent(cs)
	gross := gamedata.GravityAdjustedProduction(cs.Workers*cs.IndustryPerWorker, pct) + cs.FlatIndustry
	pollutingProd, cleanupCost, netIndustry := colonyPollution(cs, gross)
	research := gamedata.GravityAdjustedProduction(cs.Scientists*cs.ResearchPerScientist, pct) + cs.FlatResearch
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
