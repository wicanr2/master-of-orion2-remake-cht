package gamedata

// OriginalHumanTargetPersonalityScore 對應 word_181080 的七個 signed word。
var OriginalHumanTargetPersonalityScore = [...]int{-10, -5, -3, 0, 20, 20, -10}

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
	formalPolicy, additional int, targetCharismatic bool) (int, bool) {
	if relationRaw < -128 || relationRaw > 127 || personality < 0 || personality >= len(OriginalHumanTargetPersonalityScore) ||
		diplomatBonus < 0 || difficulty < 0 || difficulty > 6 || formalPolicy < 0 || formalPolicy > 6 {
		return 0, false
	}
	relationScore := relationRaw / 10
	if relationRaw >= 50 {
		relationScore = relationRaw / 5
	}
	score := relationScore + OriginalHumanTargetPersonalityScore[personality] + diplomatBonus + 15 - difficulty + additional
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
