package ai

// 原版 AI 性格行為表(14 張)。
//
// 來源:反組譯 Orion2.exe 的除錯符號表(2026-08-06),module 13 與 27 的 `_personality_*`
// 資料表,以 16-bit word 陣列儲存,**每張表 7 個有效值**(其後補零)。
// 7 這個數字與 remake 既有的 Personality 列舉(來自 patch1.5 內附 AIRACES.CFG 的
// `race_personality` 0-6)完全對應,兩個獨立來源互證。
//
// 語意逐項自洽,不是靠位址猜的:
//
//	relation_modifiers = [-50, -20, -20, 0, 20, 30, -70]
//	                      排外  冷酷  好戰 反覆 重信譽 和平  失信
//	→ 排外最不友善、和平主義最友善、失信(外交狀態)最差,完全符合各性格的名字。
//	expansion_chance   = [80, 100, 80, 70, 90, 30, 90] → 冷酷最會擴張、和平主義最不擴張。
//	break_peace_treaty = [7, 10, 1, 1, 0, 0, 5]        → 重信譽與和平主義從不毀約。
//
// remake 先前的 AI 關係演化用的是與性格無關的固定數字(勝負 ±10/−20、軍力差 /20),
// 三個 AI 對手除了名字以外行為完全一樣。這些表把「性格」變成真的有差別的東西。
//
// ⚠ 表的**用法**(原版在哪個判斷點怎麼套用這些數字)還沒有逐條反編確認,
// remake 目前是「取用數字、用自己的判斷點套」——數值有硬證,套用方式是 remake 的設計。
// 下面每個查詢函式都標明了這一點。

// personalityTableLen 是每張表的有效欄數(= Personality 的種類數)。
const personalityTableLen = 7

// 以下每張表的索引順序都是 Personality 列舉:
// 0 排外 / 1 冷酷無情 / 2 好戰 / 3 反覆無常 / 4 重信譽 / 5 和平主義 / 6 失信
var (
	// personalityRelationModifiers 是各性格的基礎關係傾向(原版 `_personality_relation_modifiers`
	// @ 0x180CCC)。正值 = 天生比較友善。
	personalityRelationModifiers = [personalityTableLen]int{-50, -20, -20, 0, 20, 30, -70}

	// personalityThreatModifiers 是發出威脅的傾向(`_personality_threat_modifiers` @ 0x180CDC)。
	personalityThreatModifiers = [personalityTableLen]int{40, 30, 20, 5, 0, -10, 50}

	// personalityMaxThreats 是同時能維持的威脅目標數上限(`_personality_max_threats` @ 0x180CF0)。
	personalityMaxThreats = [personalityTableLen]int{1, 2, 3, 3, 4, 5, 2}

	// personalitySneakAttackModifiers 是偷襲傾向(`_personality_sneak_attacks_modifiers` @ 0x180D04)。
	personalitySneakAttackModifiers = [personalityTableLen]int{-10, -50, -30, -20, 0, 20, 0}

	// personalitySneakAttackRestrictions 是偷襲的限制條件修正
	// (`_personality_sneak_attack_restictions` @ 0x180D18,原版拼字如此)。
	personalitySneakAttackRestrictions = [personalityTableLen]int{-10, -50, -30, 0, 50, 15, -30}

	// personalityNPCPeaceModifiers 是接受和平提案的傾向(`_personality_npc_peace_modifiers` @ 0x180D2C)。
	personalityNPCPeaceModifiers = [personalityTableLen]int{-10, -50, -30, 0, 50, 15, -30}

	// personalityWinningPeaceProposalChance 是「佔上風時仍主動提和」的機率
	// (`_base_winning_peace_proposal_chance` @ 0x180D40)。
	personalityWinningPeaceProposalChance = [personalityTableLen]int{5, 5, 10, 30, 40, 50, 1}

	// personalityThreatDemandChance 是提出要脅性要求的機率(`_threat_demand_chance` @ 0x180D54)。
	personalityThreatDemandChance = [personalityTableLen]int{0, 30, 80, 20, 10, 0, 80}

	// personalityDemandModifiers 是要求的強硬程度修正(`_personality_demand_modifiers` @ 0x180D68)。
	personalityDemandModifiers = [personalityTableLen]int{10, 50, 30, 0, -50, -20, -40}

	// personalityExpansionChance 是擴張(拓殖新星)的積極度(`_personality_expansion_chance` @ 0x180D7C)。
	personalityExpansionChance = [personalityTableLen]int{80, 100, 80, 70, 90, 30, 90}

	// personalityLosingGroundChance 是「劣勢時的反應強度」(`_personality_losing_ground_chance` @ 0x180D90)。
	personalityLosingGroundChance = [personalityTableLen]int{60, 100, 70, 50, 30, 0, 100}

	// personalityRewardProposalChance 是主動給好處以修補關係的機率
	// (`_personality_reward_proposal_chance` @ 0x180DA4)。
	personalityRewardProposalChance = [personalityTableLen]int{0, 5, 5, 10, 15, 20, 0}

	// personalityPeaceDuration 是締結和約後願意維持的回合數(`_personality_peace_duration` @ 0x18105C)。
	personalityPeaceDuration = [personalityTableLen]int{5, 10, 20, 5, 50, 40, 5}

	// personalityBreakPeaceTreatyChance 是撕毀和約的機率(`_personality_break_peace_treaty_chance`
	// @ 0x181070)。重信譽(4)與和平主義(5)是 0——從不毀約。
	personalityBreakPeaceTreatyChance = [personalityTableLen]int{7, 10, 1, 1, 0, 0, 5}
)

// clampPersonality 把性格代碼夾進表的範圍(未知值退回「反覆無常」這個中性性格)。
func clampPersonality(p Personality) int {
	i := int(p)
	if i < 0 || i >= personalityTableLen {
		return int(PersonalityErratic)
	}
	return i
}

// --- 查詢 API(數值有硬證;「在哪個判斷點套用」是 remake 的設計,見檔頭)---

// PersonalityRelationModifier 回傳該性格的基礎關係傾向。
func PersonalityRelationModifier(p Personality) int {
	return personalityRelationModifiers[clampPersonality(p)]
}

// PersonalityThreatModifier 回傳發出威脅的傾向修正。
func PersonalityThreatModifier(p Personality) int {
	return personalityThreatModifiers[clampPersonality(p)]
}

// PersonalityMaxThreats 回傳同時維持的威脅目標數上限。
func PersonalityMaxThreats(p Personality) int {
	return personalityMaxThreats[clampPersonality(p)]
}

// PersonalitySneakAttackModifier 回傳偷襲傾向修正。
func PersonalitySneakAttackModifier(p Personality) int {
	return personalitySneakAttackModifiers[clampPersonality(p)]
}

// PersonalitySneakAttackRestriction 回傳偷襲限制條件修正。
func PersonalitySneakAttackRestriction(p Personality) int {
	return personalitySneakAttackRestrictions[clampPersonality(p)]
}

// PersonalityPeaceModifier 回傳接受和平提案的傾向修正。
func PersonalityPeaceModifier(p Personality) int {
	return personalityNPCPeaceModifiers[clampPersonality(p)]
}

// PersonalityWinningPeaceProposalChance 回傳「佔上風時仍主動提和」的機率(%)。
func PersonalityWinningPeaceProposalChance(p Personality) int {
	return personalityWinningPeaceProposalChance[clampPersonality(p)]
}

// PersonalityThreatDemandChance 回傳提出要脅性要求的機率(%)。
func PersonalityThreatDemandChance(p Personality) int {
	return personalityThreatDemandChance[clampPersonality(p)]
}

// PersonalityDemandModifier 回傳要求強硬程度的修正。
func PersonalityDemandModifier(p Personality) int {
	return personalityDemandModifiers[clampPersonality(p)]
}

// PersonalityExpansionChance 回傳拓殖新星的積極度(%)。
func PersonalityExpansionChance(p Personality) int {
	return personalityExpansionChance[clampPersonality(p)]
}

// PersonalityLosingGroundChance 回傳劣勢時的反應強度(%)。
func PersonalityLosingGroundChance(p Personality) int {
	return personalityLosingGroundChance[clampPersonality(p)]
}

// PersonalityRewardProposalChance 回傳主動給好處修補關係的機率(%)。
func PersonalityRewardProposalChance(p Personality) int {
	return personalityRewardProposalChance[clampPersonality(p)]
}

// PersonalityPeaceDuration 回傳締結和約後願意維持的回合數。
func PersonalityPeaceDuration(p Personality) int {
	return personalityPeaceDuration[clampPersonality(p)]
}

// PersonalityBreakPeaceTreatyChance 回傳撕毀和約的機率(%)。
func PersonalityBreakPeaceTreatyChance(p Personality) int {
	return personalityBreakPeaceTreatyChance[clampPersonality(p)]
}

// ProfileForPersonality 依性格挑一個經濟傾向 Profile。
//
// 原版的「經濟傾向」是另一個維度(AI objective:Militarist/Technologist/Expansionist…),
// 與性格分開;remake 目前只有 Profile 一層,這裡用擴張積極度與威脅傾向把性格映射過去,
// 讓三個 AI 對手不再共用同一套經濟行為。映射本身是 remake 的設計,不是原版對照。
func ProfileForPersonality(p Personality) Profile {
	switch p {
	case PersonalityPacifist:
		return ProfileScientific // 擴張 30(全表最低)、威脅 -10 → 內政科研路線
	case PersonalityRuthless, PersonalityXenophobic:
		return ProfileAggressive // 擴張 100/80 且威脅傾向高
	case PersonalityAggressive:
		return ProfileAggressive
	case PersonalityHonorable:
		return ProfileExpansionist // 擴張 90 但威脅 0:穩健開拓
	default:
		return ProfileBalanced
	}
}
