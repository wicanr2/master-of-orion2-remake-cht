package gamedata

import "testing"

// 手冊 p.137:只有 Megawealth 與 Researcher 可累加,其餘取最佳者。
func TestOnlyMegawealthAndResearcherAreCumulative(t *testing.T) {
	if !LeaderSkillCumulative(int(SKILL_MEGAWEALTH)) {
		t.Error("Megawealth 應可累加")
	}
	if !LeaderSkillCumulative(int(SKILL_RESEARCHER)) {
		t.Error("Researcher 應可累加")
	}
	for _, id := range []LeaderSkills{
		SKILL_TRADER, SKILL_FINANCIAL_LEADER, SKILL_SPIRITUAL_LEADER,
		SKILL_MEDICINE, SKILL_INSTRUCTOR, SKILL_COMMANDO, SKILL_NAVIGATOR,
	} {
		if LeaderSkillCumulative(int(id)) {
			t.Errorf("技能 %d 不該可累加(手冊只列了 Megawealth 與 Researcher)", id)
		}
	}
}

// 累加型相加、其餘取最佳——這是 remake 先前做錯的那條。
func TestLeaderSkillCombineFollowsTheApplicabilityRule(t *testing.T) {
	// Researcher 累加:兩個 +5 = +10。
	if got := LeaderSkillCombine(int(SKILL_RESEARCHER), []int{5, 5}); got != 10 {
		t.Errorf("Researcher 兩份 +5 應累加成 +10,得到 %d", got)
	}
	// Trader 取最佳:+5 與 +12 → +12,不是 +17。
	if got := LeaderSkillCombine(int(SKILL_TRADER), []int{5, 12}); got != 12 {
		t.Errorf("Trader 應取最佳的 +12(不是相加的 17),得到 %d", got)
	}
	// 順序不影響結果。
	if a, b := LeaderSkillCombine(int(SKILL_TRADER), []int{12, 5}),
		LeaderSkillCombine(int(SKILL_TRADER), []int{5, 12}); a != b {
		t.Errorf("取最佳不該受順序影響:%d vs %d", a, b)
	}
	// 空清單回 0。
	if got := LeaderSkillCombine(int(SKILL_TRADER), nil); got != 0 {
		t.Errorf("沒有領袖時應回 0,得到 %d", got)
	}
}

// 負加成(環保官 −10%)要取**絕對值最大**的那個——取數值最大會挑到最弱的。
func TestLeaderSkillCombinePicksStrongestNegativeBonus(t *testing.T) {
	got := LeaderSkillCombine(int(SKILL_ENVIRONMENTALIST), []int{-10, -30})
	if got != -30 {
		t.Errorf("負加成應取最強的 −30(絕對值最大),得到 %d ——"+
			"取數值最大會變成挑最弱的那個領袖", got)
	}
}

// admin 那一列的基礎值 + 單位:對照表本身。
// Instructor 是這一列唯二的「固定點數」之一,而且正好對上手冊那句
// 「Boosts the number of experience points earned each turn」。
func TestAdminSkillBaseValues(t *testing.T) {
	for _, tc := range []struct {
		id   LeaderSkills
		want int
		name string
	}{
		{SKILL_ENVIRONMENTALIST, -10, "環保官"},
		{SKILL_FARMING_LEADER, 10, "農業官"},
		{SKILL_FINANCIAL_LEADER, 10, "財務官"},
		{SKILL_INSTRUCTOR, 1, "教官"},
		{SKILL_LABOR_LEADER, 10, "勞工官"},
		{SKILL_MEDICINE, 10, "醫官"},
		{SKILL_SCIENCE_LEADER, 10, "科學官"},
		{SKILL_SPIRITUAL_LEADER, 5, "心靈導師"},
		{SKILL_TACTICS, 2, "戰術官"},
	} {
		// tier 1、expLevel 0 → base × (0+1) = base。
		if got := LeaderSkillBonus(int(tc.id), 1, 0); got != tc.want {
			t.Errorf("%s 的基礎加成應是 %d,得到 %d", tc.name, tc.want, got)
		}
	}
}

// 原版自己就沒實作 Tactics——remake 不做它與原版一致,不是缺口。
func TestTacticsIsUnimplementedInTheOriginalToo(t *testing.T) {
	if !LeaderSkillTacticsIsUnimplementedInTheOriginal {
		t.Error("手冊 Tactics 那條的最後一句明寫 This skill is not implemented")
	}
}
