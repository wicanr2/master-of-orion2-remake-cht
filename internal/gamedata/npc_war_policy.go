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

// OriginalNPCSpecialWarCandidateInput 是 sub_25DF1 三條 type-1 特殊宣戰分支
// 共用的已證實輸入。CurrentRelationRaw 對應方向性 player+0x617；
// FoodDeficitTurns 對應 player+0x7EC，會由每回合帝國食物結餘更新。
type OriginalNPCSpecialWarCandidateInput struct {
	Difficulty         int
	Government         int
	PowerRatio         int
	Cooldown           int
	TargetIsRotating   bool
	CurrentRelationRaw int
	FoodDeficitTurns   int
	ForcedWarRaw       int
}

// OriginalNPCForcedWarCandidate 對應 sub_25DF1 的第一條 reason 113 分支。
// player+0x60E 必須恰為 1；此分支不消耗亂數。
func OriginalNPCForcedWarCandidate(in OriginalNPCSpecialWarCandidateInput) (bool, bool) {
	if !validOriginalNPCSpecialWarInput(in) || in.ForcedWarRaw < 0 || in.ForcedWarRaw > 255 {
		return false, false
	}
	return in.ForcedWarRaw == 1 && in.PowerRatio >= 100 && in.Cooldown <= 0, true
}

// OriginalNPCGovernmentWarCandidate 對應 sub_25DF1 的 reason 20：只有 raw
// government 3、國力比例至少 100，且 Random(30*(difficulty²+1)) == 1。
func OriginalNPCGovernmentWarCandidate(in OriginalNPCSpecialWarCandidateInput, roll func(int) int) (bool, bool) {
	if !validOriginalNPCSpecialWarInput(in) || roll == nil {
		return false, false
	}
	if in.Government != 3 || in.PowerRatio < 100 || in.Cooldown > 0 {
		return false, true
	}
	span := 30 * (in.Difficulty*in.Difficulty + 1)
	random := roll(span)
	if random < 1 || random > span {
		return false, false
	}
	return random == 1, true
}

// OriginalNPCHostilityWarCandidate 對應 sub_25DF1 的 reason 68。原版先用
// (-relation-5)/(2*difficulty+1) 算門檻，再於輪值目標呼叫 Random(100)。
func OriginalNPCHostilityWarCandidate(in OriginalNPCSpecialWarCandidateInput, roll func(int) int) (bool, bool) {
	if !validOriginalNPCSpecialWarInput(in) || roll == nil {
		return false, false
	}
	if !in.TargetIsRotating || in.PowerRatio < 100 || in.Cooldown > 0 {
		return false, true
	}
	threshold := (-in.CurrentRelationRaw - 5) / (2*in.Difficulty + 1)
	random := roll(100)
	if random < 1 || random > 100 {
		return false, false
	}
	return random <= threshold, true
}

// OriginalNPCFoodDeficitWarCandidate 對應 sub_25DF1 的 reason 113 機率分支：
// Random(100) < player+0x7EC。roll 採本專案既有 1-based 契約，因此等價為
// roll(100) <= FoodDeficitTurns；即使 streak 為 0，原版仍會消耗一次亂數。
func OriginalNPCFoodDeficitWarCandidate(in OriginalNPCSpecialWarCandidateInput, roll func(int) int) (bool, bool) {
	if !validOriginalNPCSpecialWarInput(in) || roll == nil {
		return false, false
	}
	random := roll(100)
	if random < 1 || random > 100 {
		return false, false
	}
	return in.PowerRatio >= 100 && in.Cooldown <= 0 && random <= in.FoodDeficitTurns, true
}

func validOriginalNPCSpecialWarInput(in OriginalNPCSpecialWarCandidateInput) bool {
	return in.Difficulty >= 0 && in.Difficulty <= 6 && in.Government >= 0 && in.Government <= 7 &&
		in.PowerRatio >= 0 && in.CurrentRelationRaw >= -128 && in.CurrentRelationRaw <= 127 &&
		in.FoodDeficitTurns >= -32768 && in.FoodDeficitTurns <= 32767
}

// OriginalNPCFoodDeficitTurns 對應 sub_4DAB2：player+0xB0 為負時遞增
// player+0x7EC，否則歸零。原欄位為 signed word，inc word 依 16-bit
// 二補數語意從 32767 回繞為 -32768。
func OriginalNPCFoodDeficitTurns(current, foodBalance int) (int, bool) {
	if current < -32768 || current > 32767 {
		return 0, false
	}
	if foodBalance >= 0 {
		return 0, true
	}
	return int(int16(uint16(int16(current)) + 1)), true
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
