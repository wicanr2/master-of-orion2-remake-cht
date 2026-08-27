package gamedata

// OriginalNPCGovernmentScores 是 word_180CCC @ IDA linear EA 0x180CCC。
// 原始 16 bytes SHA-256：d045c3a754e49617cce57a4cdbfcd3a1cdf54b955ea21ea36be5b418e6f33f3c。
var OriginalNPCGovernmentScores = [8]int{-50, -20, -20, 0, 20, 30, -70, 0}
var originalNPCIncidentGovernmentThresholds = [8]int{40, 30, 20, 5, 0, -10, 50, 0}

type OriginalNPCIncidentMemoryInput struct {
	PendingReason       int
	PendingMagnitude    int
	Memory              int
	Government          int
	ProtectedAgreement  bool
	DemocracyMemoryFlag bool
}

type OriginalNPCIncidentMemoryResult struct {
	PendingReason    int
	PendingMagnitude int
	Memory           int
}

// OriginalNPCIncidentMemoryStep 對應 sub_252D5 的單一 ordered pair。
// +0x64F 是待處理 reason、+0x65F 是 Change_Relations_ 保存的最強變動幅度，
// +0x71F 是重複事件記憶；roll100 必須回傳 1..100。
func OriginalNPCIncidentMemoryStep(in OriginalNPCIncidentMemoryInput, roll100 func() int) (OriginalNPCIncidentMemoryResult, bool) {
	out := OriginalNPCIncidentMemoryResult{PendingReason: in.PendingReason, PendingMagnitude: in.PendingMagnitude, Memory: in.Memory}
	if in.PendingReason == 0 || in.PendingMagnitude == 0 || in.PendingMagnitude == 10000 {
		return out, true
	}
	if in.Government < 0 || in.Government >= len(originalNPCIncidentGovernmentThresholds) ||
		in.Memory < 0 || in.Memory > 127 || roll100 == nil {
		return out, false
	}
	roll := roll100()
	if roll < 1 || roll > 100 {
		return out, false
	}
	if in.PendingMagnitude > 0 {
		if roll >= 2*in.PendingMagnitude {
			out.PendingReason = 0
		} else if out.Memory > 0 {
			out.Memory--
		}
		return out, true
	}

	increment := 1
	if in.ProtectedAgreement {
		increment = 2
	}
	threshold := originalNPCIncidentGovernmentThresholds[in.Government] - 2*in.PendingMagnitude
	if in.Government == 4 && in.DemocracyMemoryFlag {
		threshold = 50 - 2*in.PendingMagnitude
	}
	if roll < threshold || in.ProtectedAgreement {
		out.Memory += increment
		if out.Memory > 127 {
			out.Memory = 127
		}
	} else {
		out.PendingReason = 0
	}
	return out, true
}

func OriginalNPCGovernmentScore(rawGovernment int) (int, bool) {
	if rawGovernment < 0 || rawGovernment >= len(OriginalNPCGovernmentScores) {
		return 0, false
	}
	return OriginalNPCGovernmentScores[rawGovernment], true
}

type OriginalNPCTreatyInput struct {
	Difficulty          int
	CurrentRaw          int
	ReputationRaw       int
	TreatyBiasRaw       int
	AgreementBiasRaw    int
	Policy              ForeignPolicy
	TradeActive         bool
	ResearchActive      bool
	OuterGovernment     int
	InnerGovernment     int
	HumanWarCount       int
	NonHumanWarCount    int
	ThirdPartyBonus     int
	TributeBlocked      bool
	OuterStrength       int
	InnerStrength       int
	OuterThirdPartyWars int
}

type OriginalNPCTreatyResult struct {
	Policy           ForeignPolicy
	TradeActive      bool
	ResearchActive   bool
	TributeMode      int
	RelationDelta    int
	TreatyBiasRaw    int
	AgreementBiasRaw int
	Processed        bool
}

// OriginalNPCTreatyNegotiation 實作 sub_2552D 中單一有接觸、存活 AI ordered
// pair 的已閉合決策。roll(n) 必須回傳 1..n；非法輸出失敗即關閉。
func OriginalNPCTreatyNegotiation(in OriginalNPCTreatyInput,
	roll func(int) int) (out OriginalNPCTreatyResult, ok bool) {
	out = OriginalNPCTreatyResult{
		Policy: in.Policy, TradeActive: in.TradeActive, ResearchActive: in.ResearchActive,
		TreatyBiasRaw: in.TreatyBiasRaw, AgreementBiasRaw: in.AgreementBiasRaw,
	}
	if in.Difficulty < 0 || in.Difficulty > 6 || in.CurrentRaw < -100 || in.CurrentRaw > 100 ||
		in.ReputationRaw < -128 || in.ReputationRaw > 127 || in.Policy < DIPLO_NONE ||
		in.Policy > DIPLO_WAR || in.HumanWarCount < 0 || in.NonHumanWarCount < 0 ||
		in.OuterThirdPartyWars < 0 || roll == nil {
		return out, false
	}
	checked := func(n int) (int, bool) {
		v := roll(n)
		return v, v >= 1 && v <= n
	}
	frequency := 250 - 40*in.Difficulty
	gate, ok := checked(frequency)
	if !ok {
		return out, false
	}
	if gate != 1 {
		return out, true
	}
	out.Processed = true
	degrade := func(v int) int { return v - 30 }
	defer func() {
		out.TreatyBiasRaw = degrade(out.TreatyBiasRaw)
		out.AgreementBiasRaw = degrade(out.AgreementBiasRaw)
	}()
	if in.Policy >= DIPLO_LIMITED_WAR || in.TreatyBiasRaw < -30 {
		return out, true
	}
	outerGov, outerOK := OriginalNPCGovernmentScore(in.OuterGovernment)
	innerGov, innerOK := OriginalNPCGovernmentScore(in.InnerGovernment)
	if !outerOK || !innerOK {
		return out, false
	}
	base := in.ReputationRaw + in.CurrentRaw
	if in.Policy == DIPLO_NON_AGGRESSION {
		base += 20
	}
	if in.Policy == DIPLO_ALLIANCE {
		base += 40
	}
	if in.TradeActive {
		base += 20 // 0x25642 與 0x25666 對同一 +0x62F 各加 10。
	}
	base += in.ThirdPartyBonus

	treatyRoll, ok := checked(100)
	if !ok {
		return out, false
	}
	treatyScore := base + in.TreatyBiasRaw + treatyRoll + outerGov + innerGov + 50*in.HumanWarCount
	if treatyScore >= 200 && in.Policy == DIPLO_NON_AGGRESSION && in.NonHumanWarCount == 0 {
		out.Policy = DIPLO_ALLIANCE
	} else if treatyScore >= 100 && in.Policy != DIPLO_NON_AGGRESSION && in.Policy != DIPLO_ALLIANCE {
		out.Policy = DIPLO_NON_AGGRESSION
	}

	agreementRoll, ok := checked(100)
	if !ok {
		return out, false
	}
	agreementScore := base + in.AgreementBiasRaw + agreementRoll + innerGov
	if agreementScore > 110 && !in.ResearchActive {
		out.ResearchActive = true
	} else if agreementScore > 80 && !in.TradeActive {
		out.TradeActive = true
	}

	tributeRoll, ok := checked(100)
	if !ok {
		return out, false
	}
	tributeScore := base + in.TreatyBiasRaw + tributeRoll + outerGov + innerGov
	ratio := 100 * (in.OuterStrength + 1) / (in.InnerStrength + 1)
	if ratio > 800 {
		ratio = 800
	}
	for n := 0; n < in.OuterThirdPartyWars; n++ {
		ratio /= 2
	}
	if tributeScore > 150 && !in.TributeBlocked && ratio < 100 &&
		in.OuterStrength < in.InnerStrength {
		chance, valid := checked(20)
		if !valid {
			return out, false
		}
		if chance <= in.Difficulty+1 {
			out.TributeMode = 2
			delta, valid := checked(3)
			if !valid {
				return out, false
			}
			out.RelationDelta = delta + 3
		}
	}
	return out, true
}

// OriginalNPCNegotiationBiasRecovery 對應 sub_4DAB2：每回合 +10，正值夾回 0。
func OriginalNPCNegotiationBiasRecovery(raw int) int {
	raw += 10
	if raw > 0 {
		return 0
	}
	return raw
}
