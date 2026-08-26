package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// aiColonyBuildKey 以星系索引保存佇列；舊存檔若缺 ColonyStars，使用不會和合法星系
// 索引衝突的負值，待對映恢復後自然建立正式 key。
func aiColonyBuildKey(a *AIOpponent, colony int) int {
	if colony >= 0 && colony < len(a.ColonyStars) && a.ColonyStars[colony] >= 0 {
		return a.ColonyStars[colony]
	}
	return -colony - 1
}

// aiBuildingScore 是原版 Colony_Building_Score_ 的 typed 轉接層。已閉合 case 先走
// originalAIExactBuildingScore；其餘 case 的完整欄位語意尚未全部映射，才使用 remake 已有的
// 殖民地輸出與建築分類。加權抽選、難度濾門、候選範圍與逐殖民地產品資料形狀則直接依
// IDA 證據實作。
func aiOriginalLateTechReached(ps engine.PlayerState) bool {
	if gamedata.IsHyperAdvancedTopic(ps.ResearchTopic) {
		return true
	}
	for topic, completed := range ps.CompletedTopics {
		if completed && gamedata.IsHyperAdvancedTopic(topic) {
			return true
		}
	}
	for topic, level := range ps.HyperAdvancedLevels {
		if level > 0 && gamedata.IsHyperAdvancedTopic(topic) {
			return true
		}
	}
	return false
}

// aiOriginalPriorityBuildingGate 對映 Colony_Building_Score_ @ 0xD010D..0xD019A。
// 原版科技狀態表基址為 player+0x117，建築旗標基址為 colony+0x136；這三組科技／建築
// offset 已逐項對回受版控 enum，不以名稱相似度猜測。
func aiOriginalPriorityBuildingGate(colony engine.ColonyState, built map[string]bool, known map[gamedata.Technology]bool, government gamedata.MoraleGovernmentType) bool {
	if colony.MineralRichness <= gamedata.ABUNDANT &&
		known[gamedata.TECH_AUTOMATED_FACTORIES] && !built["自動工廠"] {
		return true
	}
	// raw 政府碼 0..3 的 signed /2 皆 <=1：Feudal／Confederation／Dictatorship／Imperium。
	if int(government)/2 > 1 {
		return false
	}
	return known[gamedata.TECH_MARINE_BARRACKS] && !built["海軍陸戰隊營"] ||
		known[gamedata.TECH_ARMOR_BARRACKS] && !built["裝甲營房"]
}

func originalAIPrimaryPopulationSlot(colony engine.ColonyState) (int, bool) {
	if !engine.PopulationGroupsComplete(colony) || !colony.OwnerRaceSlotKnown ||
		colony.OwnerRaceSlot < 0 || colony.OwnerRaceSlot >= 8 {
		return 0, false
	}
	// sub_D2A08 只計 packed colonist 低 nibble 中 0..7 的 player slot；8 以上的
	// Android／特殊槽不參與這個 AI cache 欄位。
	var counts [8]int
	total := 0
	for _, group := range colony.PopulationGroups {
		if group.RaceSlot < 0 || group.RaceSlot >= len(counts) {
			continue
		}
		n := group.Farmers + group.Workers + group.Scientists
		counts[group.RaceSlot] += n
		total += n
	}
	if total <= 0 {
		return 0, false
	}
	// 0xD2A45..0xD2A7C：同數時保留較高 slot；嚴格過半直接採該 slot。
	dominant := len(counts) - 1
	for slot := dominant - 1; slot >= 0; slot-- {
		if counts[dominant] < counts[slot] {
			dominant = slot
		}
	}
	if total < 2*counts[dominant] {
		return dominant, true
	}
	owner := colony.OwnerRaceSlot
	// 0xD2A7E..0xD2AA2 保留原版不尋常的 fallback：owner 嚴格超過三分之一時
	// 採 owner；否則只有 slot 0 非空且與 owner 數量不同時採 slot 0。
	if total < 3*counts[owner] {
		return owner, true
	}
	if counts[0] > 0 && counts[0] != counts[owner] {
		return 0, true
	}
	return owner, true
}

// originalAIFoodBuildingPopulationGate 對映 Compute_AI_Data_ 的 cache+2：只有主要人口與
// owner 同為 Lithovore 時為 false；profile 不完整時 known=false。
func originalAIFoodBuildingPopulationGate(colony engine.ColonyState) (eligible, known bool) {
	slot, known := originalAIPrimaryPopulationSlot(colony)
	if !known || !colony.OwnerRaceProfileKnown {
		return false, false
	}
	primaryLithovore := colony.Lithovore
	if slot != colony.OwnerRaceSlot {
		found := false
		for _, group := range colony.PopulationGroups {
			if group.RaceSlotKnown && group.ProfileKnown && group.RaceSlot == slot {
				primaryLithovore, found = group.Lithovore, true
				break
			}
		}
		if !found {
			return false, false
		}
	}
	return !(primaryLithovore && colony.Lithovore), true
}

// originalAIPrimaryPopulationTolerant 對映 Compute_AI_Data_ cache+4。cache+3 經 memset
// 與全部全域 xref 複核後沒有寫入端；因此污染建築讀到的 var_4 精確等於主要人口非 Tolerant。
func originalAIPrimaryPopulationTolerant(colony engine.ColonyState) (bool, bool) {
	slot, known := originalAIPrimaryPopulationSlot(colony)
	if !known || !colony.OwnerRaceProfileKnown {
		return false, false
	}
	if slot == colony.OwnerRaceSlot {
		return colony.TolerantRace, true
	}
	for _, group := range colony.PopulationGroups {
		if group.RaceSlotKnown && group.ProfileKnown && group.RaceSlot == slot {
			return group.Tolerant, true
		}
	}
	return false, false
}

type originalAIBuildScoreContext struct {
	lateTech              bool
	priorityGate          bool
	aquatic               bool
	empireFoodBalanceHalf int
	colonyFoodHalf        int
	colonyFoodHalfKnown   bool
	pollutionCleanupCost  int
	raceGrowthPercent     int
	government            gamedata.MoraleGovernmentType
	treasuryBefore        int
	netBC                 int
}

// originalAIColonyFoodHalf 對映 sub_DE03E 的 owner-independent +0xDD 快取：氣候基值先
// 轉成半單位；零基值且已知 Biomorphic Fungi 時改成 2，Weather Controller／Astro
// University 再分別加 4／2。Farming、Aquatic
// 與原住民等修飾不屬此欄，不能從已烘入這些效果的 FoodPerFarmer 直接倍增。
func originalAIColonyFoodHalf(colony engine.ColonyState, built map[string]bool, known map[gamedata.Technology]bool) (int, bool) {
	if colony.Climate < gamedata.TOXIC || colony.Climate > gamedata.GAIA {
		return 0, false
	}
	foodHalf := 2 * gamedata.ClimateFoodPerFarmer(colony.Climate)
	if foodHalf == 0 && known[gamedata.TECH_BIOMORPHIC_FUNGI] {
		foodHalf = 2
	}
	if built["氣候控制器"] {
		foodHalf += 4
	}
	if built["太空大學"] {
		foodHalf += 2
	}
	return foodHalf, true
}

// originalAIBudgetFactor 對映 Colony_Building_Score_ @ 0xD009B..0xD0142。
// sub_134C92 是 unsigned 32-bit 整數平方根，不是亂數；原版 +0xB2 是 signed word，
// 因此先以 int16 保留其儲存契約，再用 Go 的整數除法重現朝零截斷。
func originalAIBudgetFactor(treasuryBefore, netBC int) int {
	if treasuryBefore < 1500 {
		return 0
	}
	q := int(int16(netBC)) / 64
	// 負商轉成 uint32 後命中 sub_134C92 的高值捷徑並回傳 65535，隨後被夾成 10。
	if q < 0 {
		return 10
	}
	root := originalAIIntegerSqrt(q)
	if root > 10 {
		return 10
	}
	return root
}

func originalAIIntegerSqrt(n int) int {
	if n <= 0 {
		return 0
	}
	root := 0
	for (root+1)*(root+1) <= n {
		root++
	}
	return root
}

func originalAIExactBuildingScore(b gamedata.Building, colony engine.ColonyState, personality ai.Personality, ctx originalAIBuildScoreContext) (int, bool) {
	rawID, ok := gamedata.OriginalBuildingIDForName(b.NameZH)
	if !ok {
		return 0, false
	}
	honorable := 0
	if personality == ai.PersonalityHonorable {
		honorable = 1
	}
	erratic := 0
	if personality == ai.PersonalityErratic {
		erratic = 1
	}
	pacifist := 0
	if personality == ai.PersonalityPacifist {
		pacifist = 1
	}
	switch rawID {
	case 5, 13, 32: // 0xD074B..0xD077F：三棟污染處理建築
		tolerant, known := originalAIPrimaryPopulationTolerant(colony)
		if !known {
			return 0, false
		}
		if tolerant || rawID == 13 && ctx.priorityGate || ctx.pollutionCleanupCost <= 5 {
			return 0, true
		}
		if ctx.pollutionCleanupCost <= 10 {
			return pacifist, true
		}
		return originalAIIntegerSqrt(ctx.pollutionCleanupCost) + pacifist, true
	case 21, 43, 46: // 0xD07E9／0xD09F2／0xD0AB9：三棟食物建築
		eligible, known := originalAIFoodBuildingPopulationGate(colony)
		if !known || !ctx.colonyFoodHalfKnown {
			return 0, false
		}
		if !eligible || rawID != 21 && ctx.priorityGate {
			return 0, true
		}
		if rawID == 46 {
			if ctx.colonyFoodHalf <= 0 {
				return 0, true
			}
			score := 5
			if ctx.empireFoodBalanceHalf < 0 {
				score = 10
			}
			return score + 2*pacifist, true
		}
		bases := [4]int{12, 11, 10, 6}
		personalityBonus := 4 * pacifist
		if rawID == 43 {
			bases = [4]int{13, 12, 10, 7}
			personalityBonus = 3 * pacifist
		}
		bucket := ctx.colonyFoodHalf
		if bucket < 0 || bucket > 2 {
			bucket = 3
		}
		score := bases[bucket] + personalityBonus
		if ctx.empireFoodBalanceHalf < 0 {
			score -= ctx.empireFoodBalanceHalf
		}
		return score, true
	case 37: // 0xD0956..0xD0995：Soil Enrichment
		if ctx.priorityGate || colony.FoodPerFarmer <= 0 {
			return 0, true
		}
		eligible, known := originalAIFoodBuildingPopulationGate(colony)
		if !known {
			return 0, false
		}
		if !eligible {
			return 0, true
		}
		score := 3 + 2*pacifist
		if ctx.empireFoodBalanceHalf < 0 {
			score += 2
		}
		return score, true
	case 44: // 0xD0A53..0xD0AB4：Terraforming
		if ctx.priorityGate {
			return 0, true
		}
		base := 0
		switch colony.Climate {
		case gamedata.BARREN:
			base = 2
		case gamedata.DESERT, gamedata.ARID:
			base = 1
		case gamedata.TUNDRA:
			if ctx.aquatic {
				base = 1
			}
		case gamedata.OCEAN:
			if !ctx.aquatic {
				base = 4
			}
		case gamedata.SWAMP:
			if !ctx.aquatic {
				base = 6
			}
		}
		return base + 3*pacifist + originalAIBudgetFactor(ctx.treasuryBefore, ctx.netBC), true
	case 17: // 0xD07C2..0xD07C5 → 0xD0414：Gaia Transformation
		return originalAIBudgetFactor(ctx.treasuryBefore, ctx.netBC) + pacifist, true
	case 10: // 0xD071C..0xD0734：Cloning Center
		if ctx.priorityGate {
			return 0, true
		}
		score := originalAIBudgetFactor(ctx.treasuryBefore, ctx.netBC) / 2
		if ctx.raceGrowthPercent < 0 {
			score += pacifist
		}
		return score, true
	case 20, 31: // 0xD0616..0xD069E：Holo Simulator／Pleasure Dome
		// raw government / 2 == 3 只對應 Unification／Galactic Unification。
		if int(ctx.government)/2 == 3 {
			return 0, true
		}
		budgetFactor := originalAIBudgetFactor(ctx.treasuryBefore, ctx.netBC)
		if colony.Population < 3 && (budgetFactor == 0 || colony.Population < 2) {
			return 0, true
		}
		if rawID == 20 {
			return 10, true
		}
		return 16, true
	case 6: // 0xD06AD..0xD06CB：Autolab
		if ctx.priorityGate || ctx.lateTech {
			return 0, true
		}
		return 11 + 4*erratic, true
	case 19: // 0xD07CA..0xD07E4：Galactic Cybernet
		if ctx.priorityGate || ctx.lateTech {
			return 0, true
		}
		return 11, true
	case 30: // 0xD08CD..0xD08E9：Planetary Supercomputer
		if ctx.priorityGate || ctx.lateTech {
			return 0, true
		}
		return 8 + 3*erratic, true
	case 35: // 0xD0925..0xD0942：Research Laboratory
		if ctx.priorityGate || ctx.lateTech {
			return 0, true
		}
		return 5 + 2*erratic, true
	case 4: // 0xD06A3：Astro University
		return 5, true
	case 7: // 0xD06D0：Automated Factory
		return colony.Population + 13 + 2*honorable, true
	case 12: // 0xD0739：Deep Core Mine
		return colony.Population + 12 + 4*honorable, true
	case 15: // 0xD0784..0xD0792：Biospheres
		if ctx.priorityGate {
			return 0, true
		}
		return 18 + pacifist, true
	case 16: // 0xD0797..0xD07BD：Food Replicators
		eligible, known := originalAIFoodBuildingPopulationGate(colony)
		if !known {
			return 0, false
		}
		if !eligible {
			return 0, true
		}
		if ctx.empireFoodBalanceHalf < 0 {
			return 8 + pacifist, true
		}
		return 4 + pacifist, true
	case 34: // 0xD0918：Robotic Factory
		return 12 + 2*honorable, true
	case 36: // 0xD0947：Robo Mining Plant
		return colony.Population + 5 + 2*honorable, true
	default:
		return 0, false
	}
}

func aiBuildingScore(b gamedata.Building, colony engine.ColonyState, out engine.ColonyOutput, personality ai.Personality, ctx originalAIBuildScoreContext) int {
	if score, exact := originalAIExactBuildingScore(b, colony, personality, ctx); exact {
		return score
	}
	score := 20
	switch b.Category {
	case gamedata.CategoryProduction, gamedata.CategoryEnvironment:
		score += 80
	case gamedata.CategoryFood:
		score += 45
		if out.FoodSurplusHalf <= 0 {
			score += 80
		}
	case gamedata.CategoryResearch:
		score += 55
	case gamedata.CategoryDefense, gamedata.CategorySatellite, gamedata.CategoryMilitary:
		score += 40
	case gamedata.CategoryHousing:
		if colony.Population >= colony.PopMax-2 {
			score += 70
		}
	case gamedata.CategoryTrade, gamedata.CategoryMorale, gamedata.CategorySociety:
		score += 30
	}
	if score > 1000 {
		return 1000
	}
	return score
}

func chooseAIColonyBuilding(a *AIOpponent, colony int, empireOut engine.EmpireOutput, difficulty, turn int) (ColonyBuild, bool) {
	if colony < 0 || colony >= len(a.Colonies) {
		return ColonyBuild{}, false
	}
	if colony >= len(empireOut.Colonies) {
		return ColonyBuild{}, false
	}
	out := empireOut.Colonies[colony]
	built := map[string]bool(nil)
	if colony < len(a.ColonyBuildings) {
		built = a.ColonyBuildings[colony]
	}
	type candidate struct {
		build ColonyBuild
		score int
	}
	var candidates []candidate
	lateTech := aiOriginalLateTechReached(a.Player)
	government := effectiveAIGovernment(a)
	known := knownTechnologyApplications(a.Player)
	priorityGate := aiOriginalPriorityBuildingGate(a.Colonies[colony], built, known, government)
	ctx := originalAIBuildScoreContext{
		lateTech: lateTech, priorityGate: priorityGate,
		aquatic:               aiRaceHasTrait(*a, gamedata.TRAIT_AQUATIC),
		empireFoodBalanceHalf: empireOut.TotalFoodHalf,
		raceGrowthPercent:     aiColonistProductionProfile(*a).growth,
		government:            government,
		treasuryBefore:        empireOut.Player.BC - empireOut.NetBC,
		netBC:                 empireOut.NetBC,
		pollutionCleanupCost:  out.PollutionCleanupCost,
	}
	ctx.colonyFoodHalf, ctx.colonyFoodHalfKnown = originalAIColonyFoodHalf(a.Colonies[colony], built, known)
	maxScore := 1 // raw Assign_Colony_New_Building_ 也把最大分數下限夾到 1。
	for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
		if built[b.NameZH] {
			continue
		}
		score := aiBuildingScore(b, a.Colonies[colony], out, a.Personality, ctx)
		if score > maxScore {
			maxScore = score
		}
		candidates = append(candidates, candidate{ColonyBuild{Name: b.NameZH, Cost: b.ProductionCost}, score})
	}
	for _, action := range gamedata.AvailableSpecialActions(a.Player.CompletedTopics) {
		// 目前只閉合 raw 17／37／44；其餘 Special 的原版 AI 分數與可建 gate 尚未完成。
		applicable := false
		switch action.NameZH {
		case gamedata.GaiaTransformationActionName:
			applicable = gamedata.GaiaTransformationCanApply(a.Colonies[colony].Climate)
		case gamedata.TerraformActionName:
			applicable = len(gamedata.TerraformNextClimateOptions(a.Colonies[colony].Climate)) > 0
		case gamedata.SoilEnrichmentActionName:
			applicable = gamedata.TerraformSoilEnrichmentWorks(a.Colonies[colony].Climate)
		}
		if !applicable {
			continue
		}
		proxy := gamedata.Building{NameZH: action.NameZH, NameEN: action.NameEN}
		score, exact := originalAIExactBuildingScore(proxy, a.Colonies[colony], a.Personality, ctx)
		if !exact {
			continue
		}
		if score > maxScore {
			maxScore = score
		}
		candidates = append(candidates, candidate{ColonyBuild{Name: action.NameZH, Cost: action.ProductionCost}, score})
	}
	if len(candidates) == 0 {
		return ColonyBuild{}, false
	}
	// 0xD0BC5..0xD0BD9：(6-difficulty)*score < maxScore 的候選歸零。
	if difficulty < 0 {
		difficulty = 0
	}
	if difficulty > 4 {
		difficulty = 4
	}
	total := 0
	for i := range candidates {
		if (6-difficulty)*candidates[i].score < maxScore {
			candidates[i].score = 0
		}
		total += candidates[i].score
	}
	if total <= 0 {
		return ColonyBuild{}, false
	}
	// 原版呼叫共用 PRNG；remake 尚未逐位元重現該全局亂數流，故以可存檔狀態導出的
	// 決定性取樣作明示近似，避免讀檔後換產品。
	key := aiColonyBuildKey(a, colony)
	pick := int((uint64(turn+1)*0x9E3779B185EBCA87 + uint64(key+257)*0xC2B2AE3D27D4EB4F) % uint64(total))
	for _, c := range candidates {
		if pick < c.score {
			return c.build, true
		}
		pick -= c.score
	}
	return candidates[len(candidates)-1].build, true
}

// applyAICompletedBuilding 把已完成建築接到 AI 殖民地的主要經濟欄位。建築 map 仍是
// 防禦、維護與駐軍等其他消費端的權威；這裡只處理 ColonyState 內的累積型產出欄位。
func applyAICompletedBuilding(c *engine.ColonyState, name string) {
	switch name {
	case "自動工廠":
		c.IndustryPerWorker++
		c.FlatIndustry += 5
	case "研究實驗室":
		c.FlatResearch += 5
	case "太空港":
		c.IncomeBonusPercent += 50
	case "機器人採礦廠":
		c.IndustryPerWorker += 2
		c.FlatIndustry += 10
	case "深層核心礦場":
		c.IndustryPerWorker += 3
		c.FlatIndustry += 15
	case "污染處理器":
		c.PollutionProcessor = true
	case "大氣更新器":
		c.AtmosphericRenewer = true
	case "核心廢料場":
		c.CoreWasteDump = true
	case "行星超級電腦":
		c.FlatResearch += 10
	case "銀河網路中心":
		c.FlatResearch += 15
	case "水耕農場":
		c.FlatFood += 2
	case "地底農場":
		c.FlatFood += 4
	case "氣候控制器":
		c.FoodPerFarmer += 2
	case "行星證券交易所":
		c.IncomeBonusPercent += 100
	case "太空大學":
		c.FoodPerFarmer++
		c.IndustryPerWorker++
		c.ResearchPerScientist++
	case "生態圈":
		c.PopMax += 2
	case "複製中心":
		c.FlatGrowth += gamedata.CloningCenterGrowthPoints
	case "自動實驗室":
		c.FlatResearch += 30
	case "食物複製機":
		c.FoodReplicators = true
	case "再生反應爐":
		c.Recyclotron = true
	case "行星重力產生器":
		c.NormalizeGravity = true
	case "機器人工廠":
		c.FlatIndustry += gamedata.ProdRoboticFactoryBonus(int(c.MineralRichness))
	}
}

// applyAICompletedSpecial 套用已閉合的 AI 一次性產品；成功時回傳 true，呼叫端不得把它
// 寫入 ColonyBuildings。未知 Special 維持 false，避免以玩家近似路徑冒充原版 AI。
func (s *GameSession) applyAICompletedSpecial(aiIndex, colony int, name string) bool {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) || colony < 0 || colony >= len(s.AIPlayers[aiIndex].Colonies) {
		return false
	}
	a := &s.AIPlayers[aiIndex]
	c := &a.Colonies[colony]
	target := c.Climate
	switch name {
	case gamedata.GaiaTransformationActionName:
		if gamedata.GaiaTransformationCanApply(c.Climate) {
			target = gamedata.GaiaTransformationResultClimate
		}
	case gamedata.TerraformActionName:
		options := gamedata.TerraformNextClimateOptions(c.Climate)
		if len(options) > 0 {
			target = options[0]
		}
	case gamedata.SoilEnrichmentActionName:
		if gamedata.TerraformSoilEnrichmentWorks(c.Climate) {
			c.FoodPerFarmer += gamedata.TerraformSoilEnrichmentFoodBonusPerFarmer
		}
		return true
	default:
		return false
	}
	if target == c.Climate {
		return true
	}
	applyClimateChangeToColony(c, target,
		aiRaceHasTrait(*a, gamedata.TRAIT_AQUATIC), aiRaceHasTrait(*a, gamedata.TRAIT_TOLERANT))
	syncPlanetClimate(s.aiColonyPlanet(aiIndex, colony), target)
	return true
}

// advanceAIColonyBuilds 逐殖民地消費本回合淨工業，回傳沒有可建建築之殖民地的產能，
// 供尚待進一步 RE 的艦艇產品轉接層使用。同一份產能不會同時蓋建築又造艦。
func (s *GameSession) advanceAIColonyBuilds(aiIndex int, out engine.EmpireOutput) int {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return 0
	}
	a := &s.AIPlayers[aiIndex]
	if a.ColonyBuilds == nil {
		a.ColonyBuilds = make(map[int]ColonyBuild)
	}
	shipProduction := 0
	for i := range a.Colonies {
		if i >= len(out.Colonies) || out.Colonies[i].NetIndustry <= 0 {
			continue
		}
		key := aiColonyBuildKey(a, i)
		build := a.ColonyBuilds[key]
		if build.Name == "" {
			var ok bool
			build, ok = chooseAIColonyBuilding(a, i, out, s.Difficulty, s.Turn)
			if !ok {
				shipProduction += out.Colonies[i].NetIndustry
				continue
			}
		}
		build.Progress += out.Colonies[i].NetIndustry
		if build.Progress < build.Cost {
			a.ColonyBuilds[key] = build
			continue
		}
		if !s.applyAICompletedSpecial(aiIndex, i, build.Name) {
			for len(a.ColonyBuildings) <= i {
				a.ColonyBuildings = append(a.ColonyBuildings, nil)
			}
			if a.ColonyBuildings[i] == nil {
				a.ColonyBuildings[i] = make(map[string]bool)
			}
			if !a.ColonyBuildings[i][build.Name] {
				a.ColonyBuildings[i][build.Name] = true
				applyAICompletedBuilding(&a.Colonies[i], build.Name)
			}
		}
		delete(a.ColonyBuilds, key)
	}
	return shipProduction
}
