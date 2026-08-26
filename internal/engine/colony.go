package engine

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// floorHalfToWhole 是 UI/舊 API 的相容轉換。Go 對負數除法朝 0 截斷，饑荒的 -0.5
// 若直接除 2 會變成 0、錯誤地看成不饑荒；這裡改成數學上的向下取整，保留負號。
func floorHalfToWhole(half int) int {
	whole := half / 2
	if half < 0 && half%2 != 0 {
		whole--
	}
	return whole
}

func floorQuarterToWhole(quarters int) int {
	whole := quarters / 4
	if quarters < 0 && quarters%4 != 0 {
		whole--
	}
	return whole
}

// colonyGravityPenaltyPercent 回傳本殖民地目前生效的重力懲罰百分點(0 或負值,GAME_MANUAL.pdf
// p.58)。行星重力產生器(NormalizeGravity=true,p.104「正常化至 Normal-G,消除 Low-G/Heavy-G
// 負面效果」)一律強制歸零,不論 PlanetGravity 是什麼。
//
// RaceGravityKnown 時使用所有者種族的 Low-G／Normal-G／High-G；舊 JSON 未知時才回退
// Normal-G。typed population group 完整時，實際產出改由 groupGravityPenaltyPercent
// 逐 player slot 判定；本函式只負責舊 JSON／群組不完整時的 owner fallback。
func colonyGravityPenaltyPercent(cs ColonyState) int {
	if cs.NormalizeGravity {
		return 0
	}
	raceGravity := gamedata.NORMAL_G
	if cs.RaceGravityKnown {
		raceGravity = cs.RaceGravity
	}
	return gamedata.GravityPenaltyPercent(cs.PlanetGravity, raceGravity)
}

func populationGroupsValid(cs ColonyState) bool {
	if len(cs.PopulationGroups) == 0 || len(cs.PopulationGroups) > gamedata.PopulationRaceSlots || !cs.OwnerRaceProfileKnown {
		return false
	}
	f, w, s := 0, 0, 0
	var seen [gamedata.PopulationRaceSlots]bool
	for _, g := range cs.PopulationGroups {
		if !g.RaceSlotKnown || g.RaceSlot < 0 || g.RaceSlot >= gamedata.PopulationRaceSlots || seen[g.RaceSlot] ||
			!g.ProfileKnown || g.Farmers < 0 || g.Workers < 0 || g.Scientists < 0 ||
			g.PrisonerFarmers < 0 || g.PrisonerFarmers > g.Farmers ||
			g.PrisonerWorkers < 0 || g.PrisonerWorkers > g.Workers ||
			g.PrisonerScientists < 0 || g.PrisonerScientists > g.Scientists {
			return false
		}
		seen[g.RaceSlot] = true
		f += g.Farmers
		w += g.Workers
		s += g.Scientists
	}
	return f == cs.Farmers && w == cs.Workers && s == cs.Scientists && f+w+s == cs.Population
}

type populationConsumption struct {
	foodOwner, foodAlien, foodPrisoner, foodNatives                 int
	industryOwner, industryAndroid, industryAlien, industryPrisoner int
	foodBySlot, industryBySlot                                      [gamedata.PopulationRaceSlots]int
}

func (p populationConsumption) foodTotal() int {
	return p.foodOwner + p.foodAlien + p.foodPrisoner + p.foodNatives
}

func (p populationConsumption) industryTotal() int {
	return p.industryOwner + p.industryAndroid + p.industryAlien + p.industryPrisoner
}

func groupPopulation(g PopulationGroup) int { return g.Farmers + g.Workers + g.Scientists }

func groupPrisoners(g PopulationGroup) int {
	return g.PrisonerFarmers + g.PrisonerWorkers + g.PrisonerScientists
}

// colonyPopulationConsumption 重建原版 colony+0xFC..+0x103 與逐 slot 的
// +0x104/+0x10C 半單位需求。分類與優先順序見 docs/spec/population-consumption-and-negative-growth.md。
func colonyPopulationConsumption(cs ColonyState) populationConsumption {
	var out populationConsumption
	if !populationGroupsValid(cs) {
		if !cs.Lithovore {
			if cs.Cybernetic {
				out.foodOwner = cs.Population
			} else {
				out.foodOwner = cs.Population * 2
			}
		}
		if cs.Cybernetic {
			out.industryOwner = cs.Population
		}
		if cs.OwnerRaceSlotKnown && cs.OwnerRaceSlot >= 0 && cs.OwnerRaceSlot < gamedata.PopulationRaceSlots {
			out.foodBySlot[cs.OwnerRaceSlot] = out.foodOwner
			out.industryBySlot[cs.OwnerRaceSlot] = out.industryOwner
		}
		return out
	}
	for _, g := range cs.PopulationGroups {
		pop, prisoners := groupPopulation(g), groupPrisoners(g)
		cooperative := pop - prisoners
		foodPerPop, industryPerPop := 2, 0
		switch g.RaceSlot {
		case gamedata.AndroidColonistSlot:
			foodPerPop, industryPerPop = 0, 2
		case gamedata.NativeColonistSlot:
			foodPerPop, industryPerPop = 2, 0
		default:
			if g.Cybernetic { // 原版先查 +0x8B0；即使資料同時帶 Lithovore 仍以 1 為準。
				foodPerPop, industryPerPop = 1, 1
			} else if g.Lithovore {
				foodPerPop = 0
			}
		}
		foodDemand, industryDemand := pop*foodPerPop, pop*industryPerPop
		out.foodBySlot[g.RaceSlot], out.industryBySlot[g.RaceSlot] = foodDemand, industryDemand
		switch g.RaceSlot {
		case gamedata.AndroidColonistSlot:
			out.industryAndroid += industryDemand
		case gamedata.NativeColonistSlot:
			out.foodNatives += foodDemand
		default:
			if cs.OwnerRaceSlotKnown && g.RaceSlot == cs.OwnerRaceSlot {
				out.foodOwner += foodDemand
				out.industryOwner += industryDemand
			} else {
				out.foodAlien += cooperative * foodPerPop
				out.foodPrisoner += prisoners * foodPerPop
				out.industryAlien += cooperative * industryPerPop
				out.industryPrisoner += prisoners * industryPerPop
			}
		}
	}
	return out
}

func groupGravityPenaltyPercent(cs ColonyState, g PopulationGroup) int {
	if cs.NormalizeGravity || g.GravityImmune {
		return 0
	}
	return gamedata.GravityPenaltyPercent(cs.PlanetGravity, g.Gravity)
}

func groupFoodPerFarmer(cs ColonyState, g PopulationGroup) int {
	ownerAquatic := gamedata.ClimateFoodPerFarmer(gamedata.RaceFoodClimate(cs.Climate, cs.Aquatic)) -
		gamedata.ClimateFoodPerFarmer(cs.Climate)
	groupAquatic := gamedata.ClimateFoodPerFarmer(gamedata.RaceFoodClimate(cs.Climate, g.Aquatic)) -
		gamedata.ClimateFoodPerFarmer(cs.Climate)
	return cs.FoodPerFarmer - cs.OwnerFoodBonus - ownerAquatic + g.FoodBonus + groupAquatic
}

func groupedJobOutput(cs ColonyState, job int) int {
	total := 0
	for _, g := range cs.PopulationGroups {
		units, prisoners, perUnit, workerFloor := 0, 0, 0, false
		switch job {
		case 0:
			units, prisoners, perUnit = g.Farmers, g.PrisonerFarmers, groupFoodPerFarmer(cs, g)
		case 1:
			units, prisoners = g.Workers, g.PrisonerWorkers
			perUnit = cs.IndustryPerWorker - cs.OwnerIndustryBonus + g.IndustryBonus + cs.IndustryPerWorkerBonus
			workerFloor = true
		case 2:
			units, prisoners = g.Scientists, g.PrisonerScientists
			perUnit = cs.ResearchPerScientist - cs.OwnerResearchBonus + g.ResearchBonus
		}
		base := gamedata.UncooperativeJobOutputExact(units, perUnit, prisoners, workerFloor)
		bonus := cs.MoralePercent + groupGravityPenaltyPercent(cs, g)
		switch job {
		case 0:
			bonus += cs.FoodBonusPercent + cs.GovernmentFoodBonusPercent
		case 1:
			bonus += cs.IndustryBonusPercent + cs.GovernmentIndustryBonusPercent
		case 2:
			bonus += cs.ResearchBonusPercent + cs.GovernmentResearchBonusPercent
		}
		total += gamedata.GravityAdjustedProduction(base, bonus)
	}
	return total
}

func colonyPrisonersByJob(cs ColonyState) (farmers, workers, scientists int) {
	if cs.UnassimilatedFarmers >= 0 && cs.UnassimilatedWorkers >= 0 && cs.UnassimilatedScientists >= 0 &&
		cs.UnassimilatedFarmers+cs.UnassimilatedWorkers+cs.UnassimilatedScientists == cs.UnassimilatedPop {
		return cs.UnassimilatedFarmers, cs.UnassimilatedWorkers, cs.UnassimilatedScientists
	}
	return gamedata.UncooperativeAlienUnits(cs.Farmers, cs.UnassimilatedPop, cs.Population),
		gamedata.UncooperativeAlienUnits(cs.Workers, cs.UnassimilatedPop, cs.Population),
		gamedata.UncooperativeAlienUnits(cs.Scientists, cs.UnassimilatedPop, cs.Population)
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
func colonyFood(cs ColonyState, consumption populationConsumption) (food, consumed, surplus, foodHalf, consumedHalf, surplusHalf int) {
	// FoodBonusPercent(農業官)與士氣/重力合併成單一百分點再套一次公式,
	// 理由同上面那段註解:避免多次連續整數除法的複合誤差。
	pct := cs.MoralePercent + colonyGravityPenaltyPercent(cs) + cs.FoodBonusPercent + cs.GovernmentFoodBonusPercent
	prisonerFarmers, _, _ := colonyPrisonersByJob(cs)
	if populationGroupsValid(cs) {
		food = groupedJobOutput(cs, 0) + cs.FlatFood
	} else {
		food = gamedata.GravityAdjustedProduction(
			gamedata.UncooperativeJobOutputExact(cs.Farmers, cs.FoodPerFarmer, prisonerFarmers, false),
			pct) + cs.FlatFood
	}
	food += floorQuarterToWhole(cs.Farmers * cs.AIDifficultyFoodQuarters)
	consumedHalf = consumption.foodTotal()
	foodHalf = food * 2
	surplusHalf = foodHalf - consumedHalf
	consumed = floorHalfToWhole(consumedHalf)
	surplus = floorHalfToWhole(surplusHalf)
	return food, consumed, surplus, foodHalf, consumedHalf, surplusHalf
}

// colonyPollution 依毛工業產出計算污染清理成本與淨工業。
// 順序(對照 production.go 註解建議):eighths → 產污產能 → **環保官** → 清理成本 → 淨工業。
//
// 環保官夾在「查表」與「扣容忍值」之間:它降的是「會產生污染的產能」(手冊逐字用語,
// 見 gamedata.PollutionReducedByPercent),與建築的八分之幾同一個量、同一條相乘鏈。
// 放到 grossIndustry 那一側會變成減產能——那是手冊那句話的反面。
func colonyPollution(cs ColonyState, grossIndustry int) (pollutingProd, cleanupCost, netIndustry int) {
	tolerance := gamedata.PollutionTolerance(cs.PlanetSize)
	if cs.NanoDisassemblers {
		tolerance = gamedata.PollutionToleranceWithNanoDisassemblers(cs.PlanetSize)
	}
	eighths := gamedata.PollutionEighths(cs.PollutionProcessor, cs.AtmosphericRenewer, cs.CoreWasteDump)
	pollutingProd = gamedata.PollutionPollutingProduction(grossIndustry, eighths)
	pollutingProd = gamedata.PollutionReducedByPercent(pollutingProd, cs.PollutionReductionPercent)
	cleanupCost = gamedata.PollutionCleanupCost(pollutingProd, tolerance, cs.TolerantRace)
	return pollutingProd, cleanupCost, grossIndustry - cleanupCost
}

// colonyGrowth 依逐 slot 人口成長公式計算本回合 signed 成長。住房獎金 h 於 Housing 配置時
// 併入；食物／工業短缺的負項由 populationShortageGrowth 先建立，再與正常正成長相加。
// 只有沒有 typed groups 的舊 JSON fallback 在饑荒時維持歷史安全行為（回 0）。
//
// FlatGrowth(複製中心 p.99)只在「尚未達人口上限」時併入,對齊手冊「直到達星球人口上限為止」
// ——一旦 Population>=PopMax,連同 base 一起归零,不會讓固定成長點數在封頂後繼續無意義累積
// (避免之後 PopMax 因生態圈等建築提高時,尘封已久的虛增點數瞬間兌現成不合理的暴增人口)。
func groupPopulationLimit(cs ColonyState, g PopulationGroup) int {
	ownerDelta := gamedata.RacePopulationCapacityDelta(cs.PlanetSize, cs.Climate,
		cs.Aquatic, cs.TolerantRace, cs.Subterranean)
	groupDelta := gamedata.RacePopulationCapacityDelta(cs.PlanetSize, cs.Climate,
		g.Aquatic, g.Tolerant, g.Subterranean)
	limit := cs.PopMax - ownerDelta + groupDelta
	if limit < 0 {
		return 0
	}
	if limit > gamedata.MaxPopulation {
		return gamedata.MaxPopulation
	}
	return limit
}

func consumeDemand(supply *int, demand *int, quantum int) {
	for *supply >= quantum && *demand >= quantum {
		*supply -= quantum
		*demand -= quantum
	}
}

func populationShortageGrowth(cs ColonyState, consumption populationConsumption, foodHalf, industryHalf int) [gamedata.PopulationRaceSlots]int {
	var rates [gamedata.PopulationRaceSlots]int
	if !populationGroupsValid(cs) {
		return rates
	}
	if foodHalf < 0 {
		foodHalf = 0
	}
	if industryHalf < 0 {
		industryHalf = 0
	}
	foodOwner, foodAlien := consumption.foodOwner, consumption.foodAlien
	foodPrisoner, foodNatives := consumption.foodPrisoner, consumption.foodNatives
	consumeDemand(&foodHalf, &foodOwner, 1)
	consumeDemand(&foodHalf, &foodAlien, 1)
	consumeDemand(&foodHalf, &foodPrisoner, 1)
	consumeDemand(&foodHalf, &foodNatives, 1)

	industryOwner, industryAndroid := consumption.industryOwner, consumption.industryAndroid
	industryAlien, industryPrisoner := consumption.industryAlien, consumption.industryPrisoner
	consumeDemand(&industryHalf, &industryOwner, 1)
	consumeDemand(&industryHalf, &industryAndroid, 2)
	consumeDemand(&industryHalf, &industryAlien, 1)
	consumeDemand(&industryHalf, &industryPrisoner, 1)

	if cs.OwnerRaceSlotKnown && cs.OwnerRaceSlot >= 0 && cs.OwnerRaceSlot < gamedata.PopulationRaceSlots {
		rates[cs.OwnerRaceSlot] = -25 * (foodOwner + industryOwner)
	}
	rates[gamedata.AndroidColonistSlot] = -500 * industryAndroid
	rates[gamedata.NativeColonistSlot] = -25 * foodNatives
	originalAlienDemand := consumption.foodAlien + consumption.foodPrisoner +
		consumption.industryAlien + consumption.industryPrisoner
	remainingAlienDemand := foodAlien + foodPrisoner + industryAlien + industryPrisoner
	if originalAlienDemand > 0 && remainingAlienDemand > 0 {
		pool := -25 * remainingAlienDemand
		for slot := 0; slot < gamedata.AndroidColonistSlot; slot++ {
			if cs.OwnerRaceSlotKnown && slot == cs.OwnerRaceSlot {
				continue
			}
			demand := consumption.foodBySlot[slot] + consumption.industryBySlot[slot]
			rates[slot] += pool * demand / originalAlienDemand
		}
	}
	return rates
}

func colonyGrowth(cs ColonyState, foodSurplus, netIndustry int, shortage [gamedata.PopulationRaceSlots]int) (int, [gamedata.PopulationRaceSlots]int, int) {
	var rates [gamedata.PopulationRaceSlots]int
	if populationGroupsValid(cs) {
		total := 0
		housing := 0
		if cs.Housing {
			housing = gamedata.ColonyHousingBonus(netIndustry, cs.Population)
		}
		for i, g := range cs.PopulationGroups {
			groupPop := groupPopulation(g)
			rates[i] = shortage[g.RaceSlot]
			if g.RaceSlot < gamedata.AndroidColonistSlot { // 原版正成長 pass 只走一般 player slots。
				base := gamedata.ColonyBaseGrowth(groupPop, cs.Population, groupPopulationLimit(cs, g))
				rates[i] += gamedata.ColonyGrowth(base, cs.GrowthBonusSum+g.GrowthBonusPercent+housing)
			}
			total += rates[i]
		}
		if cs.Population < cs.PopMax && cs.FlatGrowth != 0 {
			pick := 0
			for i, g := range cs.PopulationGroups {
				if cs.OwnerRaceSlotKnown && g.RaceSlotKnown && g.RaceSlot == cs.OwnerRaceSlot {
					pick = i
					break
				}
			}
			rates[pick] += cs.FlatGrowth
			total += cs.FlatGrowth
		}
		return total, rates, len(cs.PopulationGroups)
	}
	if foodSurplus < 0 {
		return 0, rates, 0 // 舊 JSON fallback 沒有逐 slot 消耗 profile，保留既有安全行為。
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
	return growth, rates, 0
}

// RunColonyTurn 執行一個殖民地的一回合經濟結算,依 MOO2 順序:
// 食物 → 工業 → 污染(縮減淨工業)→ 研究 → 人口成長。
func RunColonyTurn(cs ColonyState) ColonyOutput {
	consumption := colonyPopulationConsumption(cs)
	food, consumed, surplus, foodHalf, consumedHalf, surplusHalf := colonyFood(cs, consumption)
	// 工業與研究同樣經士氣+重力調整(手冊:每格士氣 ±10%、重力 -25%/-50% 套用於
	// 食物/工業/研究三者,p.58/p.63)。FlatIndustry/FlatResearch(殖民地整體固定加成,見
	// ColonyState 欄位註解)與調整後的 per-worker 產出分開相加,採與 colonyFood/FlatFood
	// 同款「固定加成不吃士氣/重力」假設。士氣與重力合併成單一百分點再套一次公式的理由見
	// colonyFood 註解(避免兩次連續整數除法的複合誤差)。
	// FlatIndustry 在污染縮減之前併入 gross(依手冊,固定產能也算「殖民地產能」,一樣會產生
	// 污染,見下方 colonyPollution 以 gross 全額計算 pollutingProd/cleanupCost)。
	pct := cs.MoralePercent + colonyGravityPenaltyPercent(cs)
	_, prisonerWorkers, prisonerScientists := colonyPrisonersByJob(cs)
	// 工業與研究各自再吃自己那一項的百分比(勞工官 / 科學官)——士氣是三項一起動,
	// 這兩個各管各的,所以不能併進上面那個共用的 pct。
	// 未整合外星人只產出 3/4(手冊,見 gamedata 的 UncooperativeJobOutput)。
	// 工業這一項要套「每工人至少 1 產能」的下限——3/4 會把礦產最差的 1 壓成 0。
	gross := 0
	if populationGroupsValid(cs) {
		gross = groupedJobOutput(cs, 1) + cs.FlatIndustry
	} else {
		gross = gamedata.GravityAdjustedProduction(
			gamedata.UncooperativeJobOutputExact(cs.Workers, cs.IndustryPerWorker+cs.IndustryPerWorkerBonus,
				prisonerWorkers, true),
			pct+cs.IndustryBonusPercent+cs.GovernmentIndustryBonusPercent) + cs.FlatIndustry
	}
	gross += floorQuarterToWhole(cs.Workers * cs.AIDifficultyIndustryQuarters)
	pollutingProd, cleanupCost, netIndustry := colonyPollution(cs, gross)
	// 再生反應爐(p.81)加在**污染縮減之後**:手冊明說這份產能不計入污染。
	// 每單位人口 +1,不分職業——所以是 Population 而不是 Workers。
	//
	// ⚠ 這一行的位置就是它的正確性。搬到 gross 那一行去(接成 FlatIndustry)語意會變成
	// 「這份產能也會污染」,那正好是手冊否定的那句。
	recycled := 0
	if cs.Recyclotron {
		recycled = cs.Population
		netIndustry += recycled
	}
	// Cybernetic 的另一半消耗是生產力。手冊只給「half production unit」；原版存檔的
	// industry_consumption_* 也以半單位保存。這裡在污染清理與 Recyclotron 之後扣除，
	// 因為兩者是殖民地產出／回收產出的獨立來源；「扣除點」是強推論，不宣稱為手冊逐字規則。
	industryConsumedHalf := consumption.industryTotal()
	availableIndustryHalf := netIndustry * 2
	netIndustryHalf := netIndustry*2 - industryConsumedHalf
	netIndustry = floorHalfToWhole(netIndustryHalf)
	// 食物複製機(p.85)在這裡:產能已經扣完污染、成長還沒算。
	// 「as needed」= 只補缺口,所以盈餘為正時什麼都不做——那條漏掉會變成印鈔機
	// (換滿食物 → 餘糧出售換 BC),見 gamedata/food_replicators.go。
	// 這裡刻意用半單位入口，保留 Cybernetic 奇數人口的半食物缺口；舊版只換
	// 完整食物會把 deficitHalf/2 朝 0 截斷，造成一個半單位永遠無法被複製機補上。
	replicatedHalf, replicatorProductionHalf := 0, 0
	if cs.FoodReplicators && surplusHalf < 0 {
		replicatedHalf, replicatorProductionHalf = gamedata.FoodReplicatorConvertHalf(-surplusHalf, netIndustryHalf)
		foodHalf += replicatedHalf
		surplusHalf += replicatedHalf
		netIndustryHalf -= replicatorProductionHalf
		availableIndustryHalf -= replicatorProductionHalf
		if availableIndustryHalf < 0 {
			availableIndustryHalf = 0
		}
		food = floorHalfToWhole(foodHalf)
		surplus = floorHalfToWhole(surplusHalf)
		netIndustry = floorHalfToWhole(netIndustryHalf)
	}
	research := 0
	if populationGroupsValid(cs) {
		research = groupedJobOutput(cs, 2) + cs.FlatResearch
	} else {
		research = gamedata.GravityAdjustedProduction(
			gamedata.UncooperativeJobOutputExact(cs.Scientists, cs.ResearchPerScientist, prisonerScientists, false),
			pct+cs.ResearchBonusPercent+cs.GovernmentResearchBonusPercent) + cs.FlatResearch
	}
	research += floorQuarterToWhole(cs.Scientists * cs.AIDifficultyResearchQuarters)
	shortageGrowth := populationShortageGrowth(cs, consumption, foodHalf, availableIndustryHalf)
	growth, groupGrowth, groupGrowthCount := colonyGrowth(cs, surplus, netIndustry, shortageGrowth)

	return ColonyOutput{
		Food:                       food,
		FoodConsumed:               consumed,
		FoodSurplus:                surplus,
		FoodHalf:                   foodHalf,
		FoodConsumedHalf:           consumedHalf,
		FoodSurplusHalf:            surplusHalf,
		IndustryConsumedHalf:       industryConsumedHalf,
		NetIndustryHalf:            netIndustryHalf,
		Starving:                   surplusHalf < 0,
		FoodReplicated:             replicatedHalf / 2,
		FoodReplicatedHalf:         replicatedHalf,
		FoodReplicatorCostHalfBC:   replicatedHalf * gamedata.FoodReplicatorBCHalfPerHalfFood,
		GrossIndustry:              gross + recycled,
		PollutingProduction:        pollutingProd,
		PollutionCleanupCost:       cleanupCost,
		NetIndustry:                netIndustry,
		Research:                   research,
		PopGrowth:                  growth,
		PopulationGroupGrowth:      groupGrowth,
		PopulationGroupGrowthCount: groupGrowthCount,
		Cybernetic:                 cs.Cybernetic,
	}
}
