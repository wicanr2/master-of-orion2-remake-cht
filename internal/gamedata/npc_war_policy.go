package gamedata

// OriginalNPCWarCandidateInput 是 sub_25DF1 的一般宣戰候選（reason 23）輸入。
// 國力仍由 remake 的 FleetStrength 投影；門檻、政策加權與亂數順序依 1.31 executable。
type OriginalNPCWarCandidateInput struct {
	Difficulty           int
	Government           int
	Policy               ForeignPolicy
	TradeActive          bool
	ResearchActive       bool
	TributeMode          int
	SourceStrength       int
	TargetStrength       int
	SourceThirdPartyWars int
	Cooldown             int
	TargetIsRotating     bool
	TargetAtWarWithAI    bool
}

// OriginalNPCPowerRatio 對應 sub_500CF 的比例與第三方戰爭折半規則。
func OriginalNPCPowerRatio(source, target, sourceThirdPartyWars int) (int, bool) {
	if source < 0 || target < 0 || sourceThirdPartyWars < 0 {
		return 0, false
	}
	ratio := 100 * (source + 1) / (target + 1)
	if ratio > 800 {
		ratio = 800
	}
	for i := 0; i < sourceThirdPartyWars; i++ {
		ratio /= 2
	}
	return ratio, true
}

// OriginalNPCGenericWarCandidate 實作 sub_25DF1 的 reason 23 分支。
// roll(200) 必須回傳 1..200；非法輸入或亂數輸出失敗即關閉。
func OriginalNPCGenericWarCandidate(in OriginalNPCWarCandidateInput, roll func(int) int) (bool, bool) {
	if in.Difficulty < 0 || in.Difficulty > 6 || in.Policy < DIPLO_NONE || in.Policy > DIPLO_WAR ||
		in.SourceStrength < 0 || in.TargetStrength < 0 || in.SourceThirdPartyWars < 0 || roll == nil {
		return false, false
	}
	if in.Cooldown > 0 || !in.TargetIsRotating || in.Policy >= DIPLO_LIMITED_WAR ||
		(in.Difficulty >= 3 && in.TargetAtWarWithAI) {
		return false, true
	}
	government, ok := OriginalNPCGovernmentScore(in.Government)
	if !ok {
		return false, false
	}
	random := roll(200)
	if random < 1 || random > 200 {
		return false, false
	}
	score := 3*government + random + 125 + 25*in.Difficulty
	switch in.Policy {
	case DIPLO_NON_AGGRESSION:
		score += 50
	case DIPLO_ALLIANCE:
		score += 100
	case DIPLO_PEACE:
		score += 200
	}
	if in.TradeActive {
		score += 50
	}
	if in.ResearchActive {
		score += 50
	}
	switch in.TributeMode {
	case 1:
		score += 50
	case 2:
		score += 100
	}
	ratio, valid := OriginalNPCPowerRatio(in.SourceStrength, in.TargetStrength, in.SourceThirdPartyWars)
	if !valid {
		return false, false
	}
	return ratio >= score, true
}

// OriginalNPCCeasefireThreshold 是 sub_2670A 的 AI↔AI 無人類第三方戰爭門檻。
func OriginalNPCCeasefireThreshold(difficulty, humanWarCount int) (int, bool) {
	if difficulty < 0 || difficulty > 6 || humanWarCount < 0 {
		return 0, false
	}
	return 90 - 15*difficulty - 20*humanWarCount, true
}
