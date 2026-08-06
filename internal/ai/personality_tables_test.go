package ai

import "testing"

// TestPersonalityTablesMatchOriginalDump 驗證 14 張表與反組譯 dump 出來的 word 陣列逐格相同。
// 值取自 Orion2.exe module 13/27 的 `_personality_*`(每張表 7 個有效值)。
func TestPersonalityTablesMatchOriginalDump(t *testing.T) {
	cases := []struct {
		name string
		got  [personalityTableLen]int
		want [7]int
	}{
		{"relation_modifiers", personalityRelationModifiers, [7]int{-50, -20, -20, 0, 20, 30, -70}},
		{"threat_modifiers", personalityThreatModifiers, [7]int{40, 30, 20, 5, 0, -10, 50}},
		{"max_threats", personalityMaxThreats, [7]int{1, 2, 3, 3, 4, 5, 2}},
		{"sneak_attacks_modifiers", personalitySneakAttackModifiers, [7]int{-10, -50, -30, -20, 0, 20, 0}},
		{"sneak_attack_restrictions", personalitySneakAttackRestrictions, [7]int{-10, -50, -30, 0, 50, 15, -30}},
		{"npc_peace_modifiers", personalityNPCPeaceModifiers, [7]int{-10, -50, -30, 0, 50, 15, -30}},
		{"winning_peace_proposal", personalityWinningPeaceProposalChance, [7]int{5, 5, 10, 30, 40, 50, 1}},
		{"threat_demand_chance", personalityThreatDemandChance, [7]int{0, 30, 80, 20, 10, 0, 80}},
		{"demand_modifiers", personalityDemandModifiers, [7]int{10, 50, 30, 0, -50, -20, -40}},
		{"expansion_chance", personalityExpansionChance, [7]int{80, 100, 80, 70, 90, 30, 90}},
		{"losing_ground_chance", personalityLosingGroundChance, [7]int{60, 100, 70, 50, 30, 0, 100}},
		{"reward_proposal_chance", personalityRewardProposalChance, [7]int{0, 5, 5, 10, 15, 20, 0}},
		{"peace_duration", personalityPeaceDuration, [7]int{5, 10, 20, 5, 50, 40, 5}},
		{"break_peace_treaty", personalityBreakPeaceTreatyChance, [7]int{7, 10, 1, 1, 0, 0, 5}},
	}
	for _, c := range cases {
		for i := 0; i < 7; i++ {
			if c.got[i] != c.want[i] {
				t.Errorf("%s[%d] = %d,原版 dump 是 %d", c.name, i, c.got[i], c.want[i])
			}
		}
	}
}

// TestPersonalityTableSemanticsAreSelfConsistent 驗證表的語意與性格名字相符。
// 這是「7 欄真的對應 Personality 0..6」的獨立佐證——若欄位順序錯了,下面每一條都會破。
func TestPersonalityTableSemanticsAreSelfConsistent(t *testing.T) {
	// 和平主義最友善、排外與失信最不友善。
	if PersonalityRelationModifier(PersonalityPacifist) <= PersonalityRelationModifier(PersonalityXenophobic) {
		t.Error("和平主義的基礎關係應優於排外")
	}
	if PersonalityRelationModifier(PersonalityDishonored) >= PersonalityRelationModifier(PersonalityXenophobic) {
		t.Error("失信(外交狀態)的關係應是全表最差")
	}
	// 冷酷無情最會擴張,和平主義最不擴張。
	if PersonalityExpansionChance(PersonalityRuthless) != 100 {
		t.Errorf("冷酷無情的擴張積極度應為 100,got %d", PersonalityExpansionChance(PersonalityRuthless))
	}
	for _, p := range []Personality{PersonalityXenophobic, PersonalityRuthless, PersonalityAggressive,
		PersonalityErratic, PersonalityHonorable, PersonalityDishonored} {
		if PersonalityExpansionChance(PersonalityPacifist) >= PersonalityExpansionChance(p) {
			t.Errorf("和平主義的擴張積極度應低於 %v", p)
		}
	}
	// 重信譽與和平主義從不毀約。
	for _, p := range []Personality{PersonalityHonorable, PersonalityPacifist} {
		if PersonalityBreakPeaceTreatyChance(p) != 0 {
			t.Errorf("%v 不應毀約,got %d%%", p, PersonalityBreakPeaceTreatyChance(p))
		}
	}
	// 重信譽維持和約最久。
	for _, p := range []Personality{PersonalityXenophobic, PersonalityRuthless, PersonalityAggressive,
		PersonalityErratic, PersonalityPacifist, PersonalityDishonored} {
		if PersonalityPeaceDuration(PersonalityHonorable) <= PersonalityPeaceDuration(p) {
			t.Errorf("重信譽的和約維持回合數應長於 %v", p)
		}
	}
}

// TestPersonalityQueriesClampUnknown 驗證未知性格值退回中性的「反覆無常」,不會越界。
func TestPersonalityQueriesClampUnknown(t *testing.T) {
	for _, p := range []Personality{-1, 7, 99} {
		if got, want := PersonalityRelationModifier(p), PersonalityRelationModifier(PersonalityErratic); got != want {
			t.Errorf("未知性格 %d 應退回反覆無常的值 %d,got %d", p, want, got)
		}
	}
}

// TestProfileForPersonalityDiffers 驗證不同性格會推導出不同的經濟傾向
// (否則性格接線就白做了——三個 AI 還是同一套行為)。
func TestProfileForPersonalityDiffers(t *testing.T) {
	seen := map[string]bool{}
	for p := PersonalityXenophobic; p <= PersonalityDishonored; p++ {
		seen[ProfileForPersonality(p).Name] = true
	}
	if len(seen) < 3 {
		t.Errorf("7 種性格應映射出至少 3 種不同的經濟傾向,got %d 種:%v", len(seen), seen)
	}
	if ProfileForPersonality(PersonalityPacifist).Name == ProfileForPersonality(PersonalityRuthless).Name {
		t.Error("和平主義與冷酷無情不應共用同一套經濟傾向")
	}
}
