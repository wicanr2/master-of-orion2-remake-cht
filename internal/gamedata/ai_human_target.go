package gamedata

// OriginalHumanTargetPersonalityScore 對應 word_181080 的七個 signed word。
var OriginalHumanTargetPersonalityScore = [...]int{-10, -5, -3, 0, 20, 20, -10}

// originalHumanTargetIncidentDivisor 對應 word_180CF0 的七個 signed word。
// 原版亦把同一張表用作 personality 最大威脅數；sub_544A1 在此重用為
// AI→真人重複事件記憶的負分除數。
var originalHumanTargetIncidentDivisor = [...]int{1, 2, 3, 3, 4, 5, 2}

// OriginalHumanTargetIncidentScore 對應 sub_544A1 @ 0x54524..0x5457A。
// memory 是 AI→真人方向 +0x71F 的 signed byte，rememberedReason 是 +0x6CF；
// 原版只讓 1..9 的 +0x64F reason 複製進 +0x6CF，因此超出範圍時失敗即關閉。
// reasonCode 是玩家可見訊息選擇器使用的 rememberedReason+70。
func OriginalHumanTargetIncidentScore(memory, rememberedReason, personality int) (score, reasonCode int, ok bool) {
	if memory < 0 || memory > 127 || rememberedReason < 0 || rememberedReason > 9 ||
		personality < 0 || personality >= len(originalHumanTargetIncidentDivisor) {
		return 0, 0, false
	}
	if memory == 0 || rememberedReason == 0 {
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
