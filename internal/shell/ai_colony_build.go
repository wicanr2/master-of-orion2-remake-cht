package shell

import (
	"math"

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

// originalAIPrimaryPopulationCapacity 對映 Compute_AI_Data_ cache+1：sub_E0C1D 以主要
// 人口種族、星球大小／氣候重建容量，再疊 Advanced City Planning +5 與 Biospheres +2。
// ColonyState.PopMax 是 owner 口徑且可能含歷史烘入值，混合人口時不能直接代用。
func originalAIPrimaryPopulationCapacity(colony engine.ColonyState, built map[string]bool, known map[gamedata.Technology]bool) (int, bool) {
	slot, ok := originalAIPrimaryPopulationSlot(colony)
	if !ok || !colony.OwnerRaceProfileKnown || colony.PlanetSize < gamedata.TINY_PLANET ||
		colony.PlanetSize > gamedata.HUGE_PLANET || colony.Climate < gamedata.TOXIC || colony.Climate > gamedata.GAIA {
		return 0, false
	}
	aquatic, tolerant, subterranean := colony.Aquatic, colony.TolerantRace, colony.Subterranean
	if slot != colony.OwnerRaceSlot {
		found := false
		for _, group := range colony.PopulationGroups {
			if group.RaceSlotKnown && group.ProfileKnown && group.RaceSlot == slot {
				aquatic, tolerant, subterranean = group.Aquatic, group.Tolerant, group.Subterranean
				found = true
				break
			}
		}
		if !found {
			return 0, false
		}
	}
	capacity := gamedata.PlanetBasePopMax(colony.PlanetSize, colony.Climate) +
		gamedata.RacePopulationCapacityDelta(colony.PlanetSize, colony.Climate, aquatic, tolerant, subterranean)
	if known[gamedata.TECH_ADVANCED_CITY_PLANNING] {
		capacity += 5
	}
	if built["生態圈"] {
		capacity += 2
	}
	return capacity, true
}

type originalAIBuildScoreContext struct {
	lateTech                      bool
	priorityGate                  bool
	aquatic                       bool
	empireFoodBalanceHalf         int
	colonyFoodHalf                int
	colonyFoodHalfKnown           bool
	pollutionCleanupCost          int
	ownerLowGravity               bool
	ownerHighGravity              bool
	primaryPopCapacity            int
	primaryPopCapKnown            bool
	netIndustry                   int
	raceGrowthPercent             int
	government                    gamedata.MoraleGovernmentType
	treasuryBefore                int
	netBC                         int
	strategicPressureContextKnown bool
	reachTreatyNear               int
	reachNoPolicyNear             int
	reachWarNear                  int
	reachExtended                 int
	incomingOtherFleetETA9        bool
	hostileAlienPopulation        bool
	armorBarracksBuilt            bool
	marineBarracksBuilt           bool
	commandPointsSupply           int
	usedCommandPoints             int
}

// originalAIStrategicPressureContext 是 raw 2／8／22／23／24／26／27／28／40／41／42／47
// 共用的 session-wide 暫態輸入；不進存檔。
// score 公式本身可精確測試，而跨帝國航程由 GameSession 在候選建立前一次投影。
type originalAIStrategicPressureContext struct {
	known                  bool
	reachTreatyNear        int
	reachNoPolicyNear      int
	reachWarNear           int
	reachExtended          int
	incomingOtherFleetETA9 bool
	hostileAlienPopulation bool
}

// originalAIFuelRangeParsecs 對映 sub_10034D 與原始表 0x17FFDE..0x18001：
// tech 167/51/98/194/184 分別寫 4/6/9/12/255。沒有任何已知 fuel application 時
// 原版寫 0；remake 不以「一般都有 Standard Fuel Cells」掩蓋不完整舊存檔。
func originalAIFuelRangeParsecs(ps engine.PlayerState) (int, bool) {
	known := knownTechnologyApplications(ps)
	switch {
	case known[gamedata.TECH_THORIUM_FUEL_CELLS]:
		return 255, true
	case known[gamedata.TECH_URRIDIUM_FUEL_CELLS]:
		return 12, true
	case known[gamedata.TECH_IRIDIUM_FUEL_CELLS]:
		return 9, true
	case known[gamedata.TECH_DEUTERIUM_FUEL_CELLS]:
		return 6, true
	case known[gamedata.TECH_STANDARD_FUEL_CELLS]:
		return 4, true
	}
	return 0, false
}

func (s *GameSession) originalAIStarDistanceParsecs(a, b int) float64 {
	if a < 0 || b < 0 || a >= len(s.Stars) || b >= len(s.Stars) {
		return math.Inf(1)
	}
	w, h := gamedata.GalaxyParsecSpan(s.GalaxySizeClass())
	dx := (s.Stars[a].X - s.Stars[b].X) * w
	dy := (s.Stars[a].Y - s.Stars[b].Y) * h
	return math.Hypot(dx, dy)
}

func anyOriginalAIColonyInRange(s *GameSession, target int, stars []int, limit float64) bool {
	for _, star := range stars {
		if s.originalAIStarDistanceParsecs(target, star) <= limit {
			return true
		}
	}
	return false
}

func (s *GameSession) originalAIPolicyBetween(ownerAI, sourceSlot int) (gamedata.ForeignPolicy, bool, bool) {
	if ownerAI < 0 || ownerAI >= len(s.AIPlayers) {
		return gamedata.DIPLO_NONE, false, false
	}
	playerSlot := -1
	for _, colony := range s.PlayerColonies {
		if colony.OwnerRaceSlotKnown {
			playerSlot = colony.OwnerRaceSlot
			break
		}
	}
	if sourceSlot == playerSlot && playerSlot >= 0 {
		return s.AIPlayers[ownerAI].Treaty.FormalPolicy, len(s.PlayerColonies) > 0, true
	}
	for i := range s.AIPlayers {
		other := s.AIPlayers[i]
		if !other.PopulationRaceSlotKnown || other.PopulationRaceSlot != sourceSlot {
			continue
		}
		if i == ownerAI {
			return gamedata.DIPLO_NONE, true, true
		}
		if ownerAI >= len(s.AIPolicies) || i >= len(s.AIPolicies[ownerAI]) {
			return gamedata.DIPLO_NONE, false, false
		}
		return s.AIPolicies[ownerAI][i], len(other.Colonies) > 0, true
	}
	return gamedata.DIPLO_NONE, false, false
}

// originalAIStrategicPressureContext 投影 sub_D3A68／sub_D3BA0、Compute_AI_Data_ cache+5 與
// sub_CFF02 的 session-wide 輸入。任何 player-slot／外交矩陣缺口都令 known=false，
// 讓 caller 回到明示 fallback，避免把舊存檔零值冒充原版精確狀態。
func (s *GameSession) originalAIStrategicPressureContext(aiIndex, colonyIndex int) originalAIStrategicPressureContext {
	ctx := originalAIStrategicPressureContext{}
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) || colonyIndex < 0 ||
		colonyIndex >= len(s.AIPlayers[aiIndex].Colonies) || colonyIndex >= len(s.AIPlayers[aiIndex].ColonyStars) {
		return ctx
	}
	owner := &s.AIPlayers[aiIndex]
	if !owner.PopulationRaceSlotKnown {
		return ctx
	}
	target := owner.ColonyStars[colonyIndex]
	if target < 0 || target >= len(s.Stars) {
		return ctx
	}

	classify := func(stars []int, ps engine.PlayerState, policy gamedata.ForeignPolicy) bool {
		rangeParsecs, ok := originalAIFuelRangeParsecs(ps)
		if !ok {
			return false
		}
		r := float64(rangeParsecs)
		if anyOriginalAIColonyInRange(s, target, stars, r) {
			switch {
			case policy >= gamedata.DIPLO_LIMITED_WAR:
				ctx.reachWarNear++
			case policy == gamedata.DIPLO_NONE:
				ctx.reachNoPolicyNear++
			default:
				ctx.reachTreatyNear++
			}
		} else if anyOriginalAIColonyInRange(s, target, stars, 1.5*r) {
			ctx.reachExtended++
		}
		return true
	}

	if len(s.PlayerColonies) > 0 {
		if !classify(s.PlayerColonyStars, s.Player, owner.Treaty.FormalPolicy) {
			return originalAIStrategicPressureContext{}
		}
	}
	for i := range s.AIPlayers {
		if i == aiIndex || len(s.AIPlayers[i].Colonies) == 0 {
			continue
		}
		if aiIndex >= len(s.AIPolicies) || i >= len(s.AIPolicies[aiIndex]) {
			return originalAIStrategicPressureContext{}
		}
		if !classify(s.AIPlayers[i].ColonyStars, s.AIPlayers[i].Player, s.AIPolicies[aiIndex][i]) {
			return originalAIStrategicPressureContext{}
		}
	}

	for i := range s.Fleets {
		if s.Fleets[i].DestStar == target && s.Fleets[i].ETA == 9 {
			ctx.incomingOtherFleetETA9 = true
		}
	}
	for i := range s.AIPlayers {
		if i != aiIndex && s.AIPlayers[i].FleetDestStar == target && s.AIPlayers[i].FleetETA == 9 {
			ctx.incomingOtherFleetETA9 = true
		}
	}

	colony := owner.Colonies[colonyIndex]
	if !engine.PopulationGroupsComplete(colony) {
		return originalAIStrategicPressureContext{}
	}
	for _, group := range colony.PopulationGroups {
		if group.RaceSlot < 0 || group.RaceSlot >= 8 || group.RaceSlot == owner.PopulationRaceSlot ||
			group.Farmers+group.Workers+group.Scientists <= 0 {
			continue
		}
		policy, active, ok := s.originalAIPolicyBetween(aiIndex, group.RaceSlot)
		if !ok {
			return originalAIStrategicPressureContext{}
		}
		if !active {
			continue
		}
		if policy >= gamedata.DIPLO_LIMITED_WAR {
			ctx.hostileAlienPopulation = true
		}
	}
	ctx.known = true
	return ctx
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
	ruthless := 0
	if personality == ai.PersonalityRuthless {
		ruthless = 1
	}
	switch rawID {
	case 1, 14: // raw 1／14 跳表直接進 0xD0417 的共同零分尾端
		return 0, true
	case 26, 27, 42, 47: // 0xD04B3..0xD0549：四種固定殖民地防禦
		if !ctx.strategicPressureContextKnown {
			return 0, false
		}
		if ctx.priorityGate && !ctx.incomingOtherFleetETA9 {
			return 0, true
		}
		incoming := 0
		if ctx.incomingOtherFleetETA9 {
			incoming = 1
		}
		score := 10*incoming + 4*ctx.reachTreatyNear + 8*ctx.reachNoPolicyNear +
			16*ctx.reachWarNear + 4*ctx.reachExtended
		if score != 0 {
			score += ruthless
		}
		return score + originalAIBudgetFactor(ctx.treasuryBefore, ctx.netBC), true
	case 8, 40, 41: // 0xD01C6..0xD02BA：三層軌道基地
		if !ctx.strategicPressureContextKnown {
			return 0, false
		}
		if ctx.priorityGate && !ctx.incomingOtherFleetETA9 {
			return 0, true
		}
		incoming := 0
		if ctx.incomingOtherFleetETA9 {
			incoming = 1
		}
		score := 10*incoming + 3*ctx.reachTreatyNear + 6*ctx.reachNoPolicyNear +
			12*ctx.reachWarNear + 3*ctx.reachExtended
		if rawID == 40 {
			score = 10*incoming + 4*ctx.reachTreatyNear + 8*ctx.reachNoPolicyNear +
				16*ctx.reachWarNear + 3*ctx.reachExtended
		}
		if deficit := ctx.usedCommandPoints + 1 - ctx.commandPointsSupply; deficit > 0 {
			score += deficit
		}
		if score != 0 {
			score += ruthless
		}
		return score + originalAIBudgetFactor(ctx.treasuryBefore, ctx.netBC), true
	case 23, 24, 28: // 0xD03B5..0xD04B2：三種 Planetary Shield
		if !ctx.strategicPressureContextKnown {
			return 0, false
		}
		if ctx.priorityGate && !ctx.incomingOtherFleetETA9 {
			return 0, true
		}
		incoming := 0
		if ctx.incomingOtherFleetETA9 {
			incoming = 1
		}
		score := 10*incoming + ctx.reachTreatyNear + 4*ctx.reachNoPolicyNear +
			12*ctx.reachWarNear + 2*ctx.reachExtended
		if rawID == 28 {
			score = 10*incoming + 4*ctx.reachTreatyNear + 8*ctx.reachNoPolicyNear +
				12*ctx.reachWarNear + 4*ctx.reachExtended
		}
		if score != 0 {
			score += ruthless
		}
		if rawID == 28 && colony.Climate == gamedata.RADIATED {
			score += 2 * pacifist
		}
		return score + originalAIBudgetFactor(ctx.treasuryBefore, ctx.netBC), true
	case 2, 22: // 0xD02BF..0xD03B2：Armor Barracks／Marine Barracks
		if !ctx.strategicPressureContextKnown {
			return 0, false
		}
		budgetFactor := originalAIBudgetFactor(ctx.treasuryBefore, ctx.netBC)
		minimumPopulation := 3
		if rawID == 22 {
			minimumPopulation = 2
		}
		if colony.Population < minimumPopulation && budgetFactor == 0 {
			return 0, true
		}
		incoming := 0
		if ctx.incomingOtherFleetETA9 {
			incoming = 1
		}
		hostileAlien := 0
		if ctx.hostileAlienPopulation {
			hostileAlien = 1
		}
		if rawID == 2 {
			score := 2*incoming + ctx.reachTreatyNear + ctx.reachNoPolicyNear +
				3*ctx.reachWarNear + ctx.reachExtended
			if score != 0 {
				score += ruthless
			}
			if !ctx.marineBarracksBuilt && int(ctx.government)/2 <= 1 {
				score += 6
			}
			return score + hostileAlien, true
		}
		score := 5*incoming + ctx.reachTreatyNear + 3*ctx.reachNoPolicyNear +
			6*ctx.reachWarNear + 2*ctx.reachExtended
		if score != 0 {
			score += ruthless
		}
		if !ctx.armorBarracksBuilt && int(ctx.government)/2 <= 1 {
			score += 12
		}
		return score + 3*hostileAlien, true
	case 29, 39: // 0xD089C..0xD08C8／0xD09D0..0xD09ED：Stock Exchange／Spaceport
		if !ctx.primaryPopCapKnown {
			return 0, false
		}
		minimumPopulation := 5
		if rawID == 39 {
			minimumPopulation = 3
		}
		if ctx.priorityGate || colony.Population < minimumPopulation {
			return 0, true
		}
		return (colony.Population + ctx.primaryPopCapacity + honorable) / 3, true
	case 33: // 0xD08EE..0xD0913 → 0xD0991：Recyclotron
		if !ctx.primaryPopCapKnown {
			return 0, false
		}
		tolerant, known := originalAIPrimaryPopulationTolerant(colony)
		if !known {
			return 0, false
		}
		nonTolerant := 1
		if tolerant {
			nonTolerant = 0
		}
		return (2*colony.Population+ctx.primaryPopCapacity)/3 + 2*(nonTolerant+pacifist+honorable), true
	case 38: // 0xD099A..0xD09CB：Space Academy
		rawNetIndustry := int(int16(ctx.netIndustry))
		if ctx.priorityGate || rawNetIndustry < 17 && colony.Population < 5 &&
			originalAIBudgetFactor(ctx.treasuryBefore, ctx.netBC) == 0 {
			return 0, true
		}
		delta := rawNetIndustry - 15
		// sub_134C92 以 unsigned compare 處理輸入；signed 負差值回 65535，
		// 隨後被 Colony_Building_Score_ 的共同尾端夾成 1000。
		if delta < 0 {
			return 1000, true
		}
		return originalAIIntegerSqrt(delta), true
	case 25: // 0xD0844..0xD089A：Planetary Gravity Generator
		// 原版先判 High-G；雙 trait 的髒資料不得被 Low-G 分支覆蓋。
		if ctx.ownerHighGravity {
			if colony.PlanetGravity == gamedata.LOW_G {
				return 3 + pacifist, true
			}
			return 0, true
		}
		if ctx.ownerLowGravity {
			switch colony.PlanetGravity {
			case gamedata.NORMAL_G:
				return 3 + pacifist, true
			case gamedata.HEAVY_G:
				return 6 + pacifist, true
			default:
				return 0, true
			}
		}
		switch colony.PlanetGravity {
		case gamedata.LOW_G:
			return 3 + pacifist, true
		case gamedata.HEAVY_G:
			return 6 + pacifist, true
		default:
			return 0, true
		}
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

func chooseAIColonyBuilding(a *AIOpponent, colony int, empireOut engine.EmpireOutput, difficulty, turn int, pressure ...originalAIStrategicPressureContext) (ColonyBuild, bool) {
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
		netIndustry:           out.NetIndustry,
		pollutionCleanupCost:  out.PollutionCleanupCost,
		ownerLowGravity:       aiRaceHasTrait(*a, gamedata.TRAIT_LOW_G),
		ownerHighGravity:      aiRaceHasTrait(*a, gamedata.TRAIT_HIGH_G),
		armorBarracksBuilt:    built["裝甲營房"],
		marineBarracksBuilt:   built["海軍陸戰隊營"],
		commandPointsSupply:   empireOut.Player.CommandPointsSupply,
		usedCommandPoints:     empireOut.Player.UsedCommandPoints,
	}
	if len(pressure) > 0 {
		ctx.strategicPressureContextKnown = pressure[0].known
		ctx.reachTreatyNear = pressure[0].reachTreatyNear
		ctx.reachNoPolicyNear = pressure[0].reachNoPolicyNear
		ctx.reachWarNear = pressure[0].reachWarNear
		ctx.reachExtended = pressure[0].reachExtended
		ctx.incomingOtherFleetETA9 = pressure[0].incomingOtherFleetETA9
		ctx.hostileAlienPopulation = pressure[0].hostileAlienPopulation
	}
	ctx.colonyFoodHalf, ctx.colonyFoodHalfKnown = originalAIColonyFoodHalf(a.Colonies[colony], built, known)
	ctx.primaryPopCapacity, ctx.primaryPopCapKnown = originalAIPrimaryPopulationCapacity(a.Colonies[colony], built, known)
	maxScore := 1 // raw Assign_Colony_New_Building_ 也把最大分數下限夾到 1。
	for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
		if built[b.NameZH] || aiOrbitalBaseSuperseded(b.NameZH, built) {
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

// aiOrbitalBaseSuperseded 阻止高階軌道基地已完工後又候選低階基地。原版三者是取代鏈，
// 不是可重複建造的三個獨立建築；這也安全處理舊存檔中殘留的低階產品。
func aiOrbitalBaseSuperseded(name string, built map[string]bool) bool {
	switch name {
	case "星基":
		return built["戰鬥站"] || built["星辰要塞"]
	case "戰鬥站":
		return built["星辰要塞"]
	default:
		return false
	}
}

// applyAICompletedOrbitalBase 維持 Star Base → Battlestation → Star Fortress 的單槽取代鏈。
func applyAICompletedOrbitalBase(name string, built map[string]bool) {
	switch name {
	case "戰鬥站":
		delete(built, "星基")
	case "星辰要塞":
		delete(built, "星基")
		delete(built, "戰鬥站")
	}
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

// applyAICompletedPlanetaryShield 維持三面護盾的取代關係，並把手冊可見的
// Radiated→Barren 效果同步到 AI colony 與全局 planet。逐發減傷由建築 map 的既有
// 軌道轟炸 consumer 讀取，不在 ColonyState 重複保存數值。
func (s *GameSession) applyAICompletedPlanetaryShield(aiIndex, colony int, name string, built map[string]bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) || colony < 0 || colony >= len(s.AIPlayers[aiIndex].Colonies) {
		return
	}
	switch name {
	case gamedata.BuildingPlanetaryFluxShield:
		delete(built, gamedata.BuildingPlanetaryRadiationShield)
	case gamedata.BuildingPlanetaryBarrierShield:
		delete(built, gamedata.BuildingPlanetaryRadiationShield)
		delete(built, gamedata.BuildingPlanetaryFluxShield)
	case gamedata.BuildingPlanetaryRadiationShield:
	default:
		return
	}
	a := &s.AIPlayers[aiIndex]
	c := &a.Colonies[colony]
	if c.Climate != gamedata.RADIATED {
		return
	}
	applyClimateChangeToColony(c, gamedata.BARREN,
		aiRaceHasTrait(*a, gamedata.TRAIT_AQUATIC), aiRaceHasTrait(*a, gamedata.TRAIT_TOLERANT))
	syncPlanetClimate(s.aiColonyPlanet(aiIndex, colony), gamedata.BARREN)
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
			build, ok = chooseAIColonyBuilding(a, i, out, s.Difficulty, s.Turn,
				s.originalAIStrategicPressureContext(aiIndex, i))
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
			if !a.ColonyBuildings[i][build.Name] && !aiOrbitalBaseSuperseded(build.Name, a.ColonyBuildings[i]) {
				s.applyAICompletedPlanetaryShield(aiIndex, i, build.Name, a.ColonyBuildings[i])
				applyAICompletedOrbitalBase(build.Name, a.ColonyBuildings[i])
				a.ColonyBuildings[i][build.Name] = true
				applyAICompletedBuilding(&a.Colonies[i], build.Name)
			}
		}
		delete(a.ColonyBuilds, key)
	}
	return shipProduction
}
