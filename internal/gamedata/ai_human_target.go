package gamedata

// OriginalHumanTargetPersonalityScore 對應 word_181080 的七個 signed word。
var OriginalHumanTargetPersonalityScore = [...]int{-10, -5, -3, 0, 20, 20, -10}

// originalHumanTargetIncidentDivisor 對應 word_180CF0 的七個 signed word。
// 原版亦把同一張表用作 personality 最大威脅數；sub_544A1 在此重用為
// AI→真人重複事件記憶的負分除數。
var originalHumanTargetIncidentDivisor = [...]int{1, 2, 3, 3, 4, 5, 2}

// OriginalHumanTargetScoreInput 是 sub_544A1 已閉合前半段的 typed 輸入。
// 預先計算欄位仍須由各自原版 producer 提供；Known=false 時 composer 失敗即關閉。
type OriginalHumanTargetScoreInput struct {
	RelationRaw, Personality, DiplomatBonus, Difficulty, FormalPolicy int
	Government, Raw28, SourcePopulation, TargetPopulationCapacity     int
	ForceWarRaw, FoodDeficitTurns, PowerRatio                         int
	GovernmentOneTargetValue                                          int
	GovernmentOneTargetExists                                         bool
	GovernmentZeroReachabilityKnown                                   bool
	GovernmentZeroReachability                                        int
	IncidentMemory, IncidentReason                                    int
	TreatyGrievance, TreatyVictimRaw, SourceRaw                       int
	TreatyVictimKnown                                                 bool
	PopulationDominance, PopulationTrend                              int
	TargetRaw1DFIs3, TargetCharismatic, TargetBetrayedHonorable       bool
}

type OriginalHumanTargetScoreResult struct {
	Score, WorstTerm, ReasonCode, ActionLimit int
	ForcedType2                               bool
}

// OriginalHumanTargetScore 對應 sub_544A1 @ 0x544A1..0x54A27 已閉合的分數順序。
// roll 採 1..n；政府與食物分支的亂數不可延後到 Outcome。
func OriginalHumanTargetScore(in OriginalHumanTargetScoreInput,
	roll func(int) int) (OriginalHumanTargetScoreResult, bool) {
	if roll == nil || in.RelationRaw < -128 || in.RelationRaw > 127 ||
		in.Personality < 0 || in.Personality >= len(OriginalHumanTargetPersonalityScore) ||
		in.DiplomatBonus < 0 || in.Difficulty < 0 || in.Difficulty > 6 ||
		in.FormalPolicy < 0 || in.FormalPolicy > 6 || in.Government < 0 || in.Government > 7 ||
		in.Raw28 < 0 || in.Raw28 > 255 || in.ForceWarRaw < 0 || in.ForceWarRaw > 255 {
		return OriginalHumanTargetScoreResult{}, false
	}
	relationTerm := in.RelationRaw / 10
	if in.RelationRaw >= 50 {
		relationTerm = in.RelationRaw / 5
	}
	out := OriginalHumanTargetScoreResult{Score: relationTerm, WorstTerm: relationTerm, ReasonCode: 67, ActionLimit: 200}
	addTerm := func(term, reason int) {
		if term < out.WorstTerm {
			out.WorstTerm, out.ReasonCode = term, reason
		}
		out.Score += term
	}
	override := func(reason, limit int) {
		out.Score, out.WorstTerm, out.ReasonCode, out.ForcedType2 = -150, -150, reason, true
		if limit != 0 {
			out.ActionLimit = limit
		}
	}
	incident, incidentReason, ok := OriginalHumanTargetIncidentScore(in.IncidentMemory, in.IncidentReason, in.Personality)
	if !ok {
		return OriginalHumanTargetScoreResult{}, false
	}
	if incident != 0 {
		addTerm(incident, incidentReason)
	}
	forced, ok := OriginalHumanTargetForcedPopulation(in.ForceWarRaw, in.Raw28,
		in.SourcePopulation, in.TargetPopulationCapacity)
	if !ok {
		return OriginalHumanTargetScoreResult{}, false
	}
	if forced {
		override(114, 0)
	}
	if in.Government == 3 {
		r := roll(200)
		trigger, valid := OriginalHumanTargetGovernmentThree(in.Government, in.Difficulty, r)
		if !valid {
			return OriginalHumanTargetScoreResult{}, false
		}
		if trigger {
			override(109, 100)
		}
	}
	foodRoll := roll(100)
	food, valid := OriginalHumanTargetFoodDeficit(in.FoodDeficitTurns, foodRoll)
	if !valid {
		return OriginalHumanTargetScoreResult{}, false
	}
	if food {
		override(119, 0)
	}
	govOne, valid := OriginalHumanTargetGovernmentOnePressure(in.Government, in.PowerRatio,
		in.GovernmentOneTargetValue, in.GovernmentOneTargetExists)
	if !valid {
		return OriginalHumanTargetScoreResult{}, false
	}
	if govOne != 0 {
		addTerm(govOne, 115)
	}
	if in.Government == 0 {
		if !in.GovernmentZeroReachabilityKnown {
			return OriginalHumanTargetScoreResult{}, false
		}
		r := roll(400)
		trigger, valid := OriginalHumanTargetGovernmentZeroExpansion(in.Government,
			in.GovernmentZeroReachability, r)
		if !valid {
			return OriginalHumanTargetScoreResult{}, false
		}
		if trigger {
			override(121, 0)
		}
	}
	grievance, grievanceReason, valid := OriginalHumanTargetTreatyGrievance(in.TreatyGrievance,
		in.TreatyVictimKnown, in.TreatyVictimRaw, in.SourceRaw)
	if !valid {
		return OriginalHumanTargetScoreResult{}, false
	}
	if grievance != 0 {
		addTerm(grievance, grievanceReason)
	}
	pressure, limit, active, valid := OriginalHumanTargetPowerPressure(in.PowerRatio, in.Government)
	if !valid {
		return OriginalHumanTargetScoreResult{}, false
	}
	if active {
		addTerm(pressure, 112)
		out.ActionLimit = limit
	}
	if in.PopulationDominance != 0 {
		addTerm(in.PopulationDominance, 178)
		out.ActionLimit = 150
	}
	if in.PopulationTrend != 0 {
		addTerm(in.PopulationTrend, 117)
	}
	if in.FormalPolicy == int(DIPLO_NON_AGGRESSION) {
		out.Score += 10
	} else if in.FormalPolicy == int(DIPLO_ALLIANCE) {
		out.Score += 20
	}
	personality := OriginalHumanTargetPersonalityScore[in.Personality]
	if in.Personality == 4 && in.TargetBetrayedHonorable {
		personality = OriginalHumanTargetPersonalityScore[6]
	}
	out.Score += personality
	if in.TargetRaw1DFIs3 {
		out.Score += 5
	}
	if in.TargetCharismatic {
		out.Score += 10
	}
	out.Score += in.DiplomatBonus + 15 - in.Difficulty
	return out, true
}

// OriginalHumanTargetIncidentScore 對應 sub_544A1 @ 0x54524..0x5457A。
// memory 是 AI→真人方向 +0x71F 的 signed byte，rememberedReason 是 +0x6CF；
// 原版只讓 1..9 的 +0x64F reason 複製進 +0x6CF，因此超出範圍時失敗即關閉。
// reasonCode 是玩家可見訊息選擇器使用的 rememberedReason+70。
func OriginalHumanTargetIncidentScore(memory, rememberedReason, personality int) (score, reasonCode int, ok bool) {
	if memory < -128 || memory > 127 || rememberedReason < 0 || rememberedReason > 9 ||
		personality < 0 || personality >= len(originalHumanTargetIncidentDivisor) {
		return 0, 0, false
	}
	if memory <= 0 || rememberedReason == 0 {
		return 0, 0, true
	}
	divisor := originalHumanTargetIncidentDivisor[personality]
	return -10 * memory / divisor, rememberedReason + 70, true
}

// OriginalHumanTargetPopulationDominance 對應 sub_544A1 @ 0x547DE..0x54889。
// populations 只包含仍有效且未被淘汰的帝國；source 是其中的來源 AI 索引。
// 原版在僅餘一國，或少於三國且來源人口嚴格高於其他存活帝國時，加上 -10。
func OriginalHumanTargetPopulationDominance(populations []int, source int) (int, bool) {
	if source < 0 || source >= len(populations) || len(populations) == 0 {
		return 0, false
	}
	for _, population := range populations {
		if population < 0 || population > 32767 {
			return 0, false
		}
	}
	if len(populations) == 1 {
		return -10, true
	}
	if len(populations) >= 3 {
		return 0, true
	}
	otherMax := -1
	for index, population := range populations {
		if index != source && population > otherMax {
			otherMax = population
		}
	}
	if populations[source] > otherMax {
		return -10, true
	}
	return 0, true
}

// OriginalHumanTargetPopulationTrend 對應 sub_544A1 @ 0x5488B..0x54949。
// 原版自第 100 個相對回合起，比較雙方目前與 40 格前的 +0xB9B 人口歷史；
// 只有真人成長量大於來源 AI 時，加入 (sourceGrowth-targetGrowth)/2。
func OriginalHumanTargetPopulationTrend(relativeTurn, sourceNow, source40, targetNow, target40 int) (int, bool) {
	if relativeTurn < 0 || sourceNow < 0 || sourceNow > 255 || source40 < 0 || source40 > 255 ||
		targetNow < 0 || targetNow > 255 || target40 < 0 || target40 > 255 {
		return 0, false
	}
	if relativeTurn < 100 {
		return 0, true
	}
	sourceGrowth := sourceNow - source40
	targetGrowth := targetNow - target40
	if targetGrowth <= sourceGrowth {
		return 0, true
	}
	return (sourceGrowth - targetGrowth) / 2, true
}

// OriginalHumanTargetPowerPressure 對應 sub_544A1 @ 0x54768..0x547DC。
// powerRatio 直接採 sub_500CF／OriginalNPCPowerRatio 的 0..800 結果。
// 一般正人口路徑在 ratio>=300 且來源政府 raw !=5 時加入 -ratio/40，並把
// 後續行動上限改成 150；active=false 表示原版未進入此分支。
func OriginalHumanTargetPowerPressure(powerRatio, sourceGovernment int) (score, actionLimit int, active, ok bool) {
	if powerRatio < 0 || powerRatio > 800 || sourceGovernment < 0 || sourceGovernment > 7 {
		return 0, 0, false, false
	}
	if powerRatio < 300 || sourceGovernment == 5 {
		return 0, 0, false, true
	}
	return -powerRatio / 40, 150, true, true
}

// OriginalHumanTargetForcedPopulation 對應 0x54593..0x545CE。sourceTypeRaw 是
// player+0x28；raw 2 無條件觸發，否則來源總人口嚴格大於真人殖民容量一半才觸發。
func OriginalHumanTargetForcedPopulation(forceWarRaw int, sourceTypeRaw int,
	sourcePopulation, targetPopulationCapacity int) (bool, bool) {
	if forceWarRaw < 0 || forceWarRaw > 255 || sourceTypeRaw < 0 || sourceTypeRaw > 255 ||
		sourcePopulation < 0 || sourcePopulation > 32767 || targetPopulationCapacity < 0 || targetPopulationCapacity > 32767 {
		return false, false
	}
	return forceWarRaw == 1 && (sourceTypeRaw == 2 || sourcePopulation > targetPopulationCapacity/2), true
}

// OriginalHumanTargetGovernmentThree 對應 0x545CE..0x54614；roll200 採 1..200。
// 觸發時 score 覆寫 -150、原因 109、行動上限 100。
func OriginalHumanTargetGovernmentThree(government, difficulty, roll200 int) (bool, bool) {
	if government < 0 || government > 7 || difficulty < 0 || difficulty > 6 || roll200 < 1 || roll200 > 200 {
		return false, false
	}
	return government == 3 && roll200 <= difficulty+1, true
}

// OriginalHumanTargetFoodDeficit 對應 0x54614..0x5464C。原版條件是
// Random_(100) < signed +0x7EC，而非 <=。
func OriginalHumanTargetFoodDeficit(foodDeficitTurns, roll100 int) (bool, bool) {
	if foodDeficitTurns < -32768 || foodDeficitTurns > 32767 || roll100 < 1 || roll100 > 100 {
		return false, false
	}
	return roll100 < foodDeficitTurns, true
}

// OriginalHumanTargetGovernmentOnePressure 對應 0x5464C..0x546C4。
// targetValue 是 source→target 的 dword +0x857，targetExists 對應 +0x837!=-1。
func OriginalHumanTargetGovernmentOnePressure(government, powerRatio, targetValue int,
	targetExists bool) (int, bool) {
	if government < 0 || government > 7 || powerRatio < 0 || powerRatio > 800 || targetValue < 0 {
		return 0, false
	}
	if government != 1 || powerRatio < 100 || targetValue < 200 || !targetExists {
		return 0, true
	}
	return -targetValue / 20, true
}

// OriginalHumanTargetGovernmentZeroExpansion 對應 0x546C4..0x5470C。
// reachableColonyScore 是 sub_DCB47 的非負總數；Random_(400)<=總數時覆寫 -150。
func OriginalHumanTargetGovernmentZeroExpansion(government, reachableColonyScore, roll400 int) (bool, bool) {
	if government < 0 || government > 7 || reachableColonyScore < 0 || roll400 < 1 || roll400 > 400 {
		return false, false
	}
	return government == 0 && roll400 <= reachableColonyScore, true
}

// OriginalHumanTargetTreatyGrievance 對應 sub_544A1 @ 0x5470C..0x54768。
// grievanceRaw 是 signed byte +0x7EE；victimRaw 是 +0x7F6。只有 writer 已知時才可
// 產生原因碼，victim==source 使用 177，否則 176。
func OriginalHumanTargetTreatyGrievance(grievanceRaw int, victimKnown bool,
	victimRaw, sourceRaw int) (score, reasonCode int, ok bool) {
	if grievanceRaw < -128 || grievanceRaw > 127 || victimRaw < -128 || victimRaw > 127 ||
		sourceRaw < -128 || sourceRaw > 127 {
		return 0, 0, false
	}
	score = 3 * grievanceRaw / 5
	if !victimKnown {
		return score, 0, true
	}
	if victimRaw == sourceRaw {
		return score, 177, true
	}
	return score, 176, true
}

// OriginalHumanTargetOutcomeInput 是 sub_544A1 尾端已閉合的決策輸入。
// Score 是上游所有方向關係、事件、性格與領袖修正合成後的 signed 分數。
type OriginalHumanTargetOutcomeInput struct {
	Score                     int
	ContactTurns              int
	PowerRatio                int
	Difficulty                int
	DiplomaticActionAvailable bool
	ForcedType2               bool
	SourceStrongest           bool
	GlobalEscalation          bool
	SourceRepulsive           bool
	TargetRepulsive           bool
}

// OriginalHumanTargetBaseScore 閉合 sub_544A1 可直接表示的基礎項；尚未 typed 的
// directional incident memory 由 additional 明確傳入，不在此函式猜值。
func OriginalHumanTargetBaseScore(relationRaw, personality, diplomatBonus, difficulty,
	formalPolicy, additional int, targetCharismatic, targetBetrayedHonorable bool) (int, bool) {
	if relationRaw < -128 || relationRaw > 127 || personality < 0 || personality >= len(OriginalHumanTargetPersonalityScore) ||
		diplomatBonus < 0 || difficulty < 0 || difficulty > 6 || formalPolicy < 0 || formalPolicy > 6 {
		return 0, false
	}
	relationScore := relationRaw / 10
	if relationRaw >= 50 {
		relationScore = relationRaw / 5
	}
	personalityScore := OriginalHumanTargetPersonalityScore[personality]
	if personality == 4 && targetBetrayedHonorable {
		personalityScore = OriginalHumanTargetPersonalityScore[6]
	}
	score := relationScore + personalityScore + diplomatBonus + 15 - difficulty + additional
	if formalPolicy == int(DIPLO_NON_AGGRESSION) {
		score += 10
	} else if formalPolicy == int(DIPLO_ALLIANCE) {
		score += 20
	}
	if targetCharismatic {
		score += 10
	}
	return score, true
}

func originalHumanTargetISqrt(value uint32) uint32 {
	var result uint32
	bit := uint32(1) << 30
	for bit > value {
		bit >>= 2
	}
	for bit != 0 {
		if value >= result+bit {
			value -= result + bit
			result = (result >> 1) + bit
		} else {
			result >>= 1
		}
		bit >>= 2
	}
	return result
}

// OriginalHumanTargetThreshold 對應 0x54A27..0x54A87。
func OriginalHumanTargetThreshold(score, contactTurns int) (int, bool) {
	if score < -32768 || score > 32767 || contactTurns < 0 || contactTurns > 250 {
		return 0, false
	}
	if score < 0 {
		return contactTurns * -score, true
	}
	v := uint32(int32(score))
	cube := v*v*v + 5 // 原版 32-bit unsigned 算術。
	root := originalHumanTargetISqrt(cube)
	if root == 0 {
		return 0, false
	}
	return contactTurns / int(root), true
}

// OriginalHumanTargetOutcome 對應 0x54AC9..0x54CBB。roll 採專案 1-based 契約；
// 函式保留原版 RNG 消耗順序，即使後續 gate 失敗也不延後已發生的擲骰。
func OriginalHumanTargetOutcome(in OriginalHumanTargetOutcomeInput, roll func(int) int) (int, bool) {
	if in.ContactTurns < 0 || in.ContactTurns > 250 || in.PowerRatio < 0 ||
		in.Difficulty < 0 || in.Difficulty > 6 || roll == nil {
		return 0, false
	}
	threshold, ok := OriginalHumanTargetThreshold(in.Score, in.ContactTurns)
	if !ok {
		return 0, false
	}
	r3 := roll(3)
	if r3 < 1 || r3 > 3 {
		return 0, false
	}
	actionCount := in.PowerRatio/40 + (r3 - 1) - 1
	if actionCount < 1 {
		actionCount = 1
	}
	if !in.DiplomaticActionAvailable {
		actionCount = 0
	}
	r100 := roll(100)
	if r100 < 1 || r100 > 100 {
		return 0, false
	}
	if r100-1 > threshold || in.ContactTurns < 10 {
		return 0, true
	}
	r16 := roll(16)
	if r16 < 1 || r16 > 16 {
		return 0, false
	}
	if r16-1+in.Difficulty >= 16 || actionCount <= 0 || in.ForcedType2 {
		return 2, true
	}
	// 原版在檢查 strongest/global/actionCount 前無條件消耗 Random_(4)。
	r4a := roll(4)
	if r4a < 1 || r4a > 4 {
		return 0, false
	}
	if in.SourceStrongest && in.GlobalEscalation && actionCount > 3 {
		return 4, true
	}
	if in.SourceRepulsive || in.TargetRepulsive {
		return 0, true
	}
	r4b := roll(4)
	if r4b < 1 || r4b > 4 {
		return 0, false
	}
	if r4b-1 > int(originalHumanTargetISqrt(uint32(actionCount+4))) {
		return 3, true
	}
	return 1, true
}
