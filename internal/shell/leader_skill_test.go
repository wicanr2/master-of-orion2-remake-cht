package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 兩個貿易家不該加兩份——手冊只有 Megawealth 與 Researcher 可累加。
// 這是 remake 先前的實際行為,第 101 項修掉。
func TestTwoTradersDoNotStack(t *testing.T) {
	one := engine.ColonyState{}
	applyLeaderColonyBonuses([]Leader{{Name: "甲", Skill: "貿易家", Level: 3, Ship: false, Tier: 1}}, &one)

	two := engine.ColonyState{}
	applyLeaderColonyBonuses([]Leader{
		{Name: "甲", Skill: "貿易家", Level: 3, Ship: false, Tier: 1},
		{Name: "乙", Skill: "貿易家", Level: 3, Ship: false, Tier: 1},
	}, &two)

	if two.IncomeBonusPercent != one.IncomeBonusPercent {
		t.Errorf("兩個同階貿易家應與一個相同(取最佳者):一個 %d%%、兩個 %d%%",
			one.IncomeBonusPercent, two.IncomeBonusPercent)
	}
	// 取的是**最強**的那個:高階領袖在場時要用高階的值。
	mixed := engine.ColonyState{}
	applyLeaderColonyBonuses([]Leader{
		{Name: "甲", Skill: "貿易家", Level: 1, Ship: false, Tier: 1},
		{Name: "乙", Skill: "貿易家", Level: 5, Ship: false, Tier: 1},
	}, &mixed)
	best := engine.ColonyState{}
	applyLeaderColonyBonuses([]Leader{{Name: "乙", Skill: "貿易家", Level: 5, Ship: false, Tier: 1}}, &best)
	if mixed.IncomeBonusPercent != best.IncomeBonusPercent {
		t.Errorf("應取最強那位的 %d%%,得到 %d%%",
			best.IncomeBonusPercent, mixed.IncomeBonusPercent)
	}
}

// 正對照:科學家**是**累加型,兩個要加兩份。
// 少了這條,「一律取最佳」也會讓上面那支通過。
func TestTwoResearchersDoStack(t *testing.T) {
	one := engine.ColonyState{}
	applyLeaderColonyBonuses([]Leader{{Name: "甲", Skill: "科學家", Level: 3, Ship: false, Tier: 1}}, &one)
	two := engine.ColonyState{}
	applyLeaderColonyBonuses([]Leader{
		{Name: "甲", Skill: "科學家", Level: 3, Ship: false, Tier: 1},
		{Name: "乙", Skill: "科學家", Level: 3, Ship: false, Tier: 1},
	}, &two)
	if two.FlatResearch != one.FlatResearch*2 {
		t.Errorf("兩個同階科學家應累加(手冊明列):一個 %d、兩個 %d(期望 %d)",
			one.FlatResearch, two.FlatResearch, one.FlatResearch*2)
	}
}

// 四個新接的 admin 技能各自落在正確的欄位。
func TestAdminLeaderSkillsLandInTheRightFields(t *testing.T) {
	for _, tc := range []struct {
		skill string
		field func(engine.ColonyState) int
		name  string
	}{
		{"財務官", func(c engine.ColonyState) int { return c.IncomeBonusPercent }, "收入%"},
		{"心靈導師", func(c engine.ColonyState) int { return c.MoralePercent }, "士氣"},
		{"醫官", func(c engine.ColonyState) int { return c.GrowthBonusSum }, "成長"},
	} {
		var c engine.ColonyState
		applyLeaderColonyBonuses([]Leader{{Name: "某人", Skill: tc.skill, Level: 3, Ship: false, Tier: 1}}, &c)
		if tc.field(c) <= 0 {
			t.Errorf("%s 應加到「%s」欄位,得到 %d", tc.skill, tc.name, tc.field(c))
		}
	}
	// 教官**不**動殖民地欄位——它是帝國層的艦員經驗。
	var c engine.ColonyState
	applyLeaderColonyBonuses([]Leader{{Name: "教頭", Skill: "教官", Level: 5, Ship: false, Tier: 1}}, &c)
	if c.IncomeBonusPercent != 0 || c.MoralePercent != 0 || c.GrowthBonusSum != 0 || c.FlatResearch != 0 {
		t.Errorf("教官不該加到任何殖民地欄位,得到 %+v", c)
	}
}

// 教官讓艦員每回合多拿經驗(手冊「all ship crews in your empire」)。
func TestInstructorBoostsCrewExperience(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.Leaders = nil
	base := s.Fleet().Ships[0].CrewXP
	s.advanceCrewExperience()
	plain := s.Fleet().Ships[0].CrewXP - base

	s.Leaders = []Leader{{Name: "教頭", Skill: "教官", Level: 3, Ship: false, Tier: 1}}
	want := gamedata.LeaderSkillBonus(int(gamedata.SKILL_INSTRUCTOR), 1,
		leaderDisplayLevelToExpLevel(3))
	if want <= 0 {
		t.Fatal("測試前提:教官的加成應為正")
	}
	base2 := s.Fleet().Ships[0].CrewXP
	s.advanceCrewExperience()
	if got := s.Fleet().Ships[0].CrewXP - base2; got != plain+want {
		t.Errorf("有教官時每回合應多 %d 點經驗:沒教官 +%d、有教官 +%d", want, plain, got)
	}
}

// 教官也不累加——兩位取最強的那位。
func TestTwoInstructorsDoNotStack(t *testing.T) {
	one := leaderInstructorXPBonus([]Leader{{Name: "甲", Skill: "教官", Level: 5, Ship: false, Tier: 1}})
	two := leaderInstructorXPBonus([]Leader{
		{Name: "甲", Skill: "教官", Level: 5, Ship: false, Tier: 1},
		{Name: "乙", Skill: "教官", Level: 5, Ship: false, Tier: 1},
	})
	if one <= 0 {
		t.Fatal("測試前提:一位教官的加成應為正")
	}
	if two != one {
		t.Errorf("兩位教官應取最強的那位(%d),得到 %d", one, two)
	}
}

// 三個分項百分比技能各自落在自己的欄位(第 102 項)。
func TestPerCategoryLeaderSkillsLandInTheirOwnFields(t *testing.T) {
	for _, tc := range []struct {
		skill string
		field func(engine.ColonyState) int
		other []func(engine.ColonyState) int
		name  string
	}{
		{"農業官", func(c engine.ColonyState) int { return c.FoodBonusPercent },
			[]func(engine.ColonyState) int{
				func(c engine.ColonyState) int { return c.IndustryBonusPercent },
				func(c engine.ColonyState) int { return c.ResearchBonusPercent },
			}, "食物%"},
		{"勞工官", func(c engine.ColonyState) int { return c.IndustryBonusPercent },
			[]func(engine.ColonyState) int{
				func(c engine.ColonyState) int { return c.FoodBonusPercent },
				func(c engine.ColonyState) int { return c.ResearchBonusPercent },
			}, "工業%"},
		{"科學官", func(c engine.ColonyState) int { return c.ResearchBonusPercent },
			[]func(engine.ColonyState) int{
				func(c engine.ColonyState) int { return c.FoodBonusPercent },
				func(c engine.ColonyState) int { return c.IndustryBonusPercent },
			}, "研究%"},
	} {
		var c engine.ColonyState
		applyLeaderColonyBonuses([]Leader{{Name: "某人", Skill: tc.skill, Level: 3, Ship: false, Tier: 1}}, &c)
		if tc.field(c) <= 0 {
			t.Errorf("%s 應加到「%s」,得到 %d", tc.skill, tc.name, tc.field(c))
		}
		for i, f := range tc.other {
			if f(c) != 0 {
				t.Errorf("%s 不該動到另外兩項(第 %d 個),得到 %d", tc.skill, i, f(c))
			}
		}
	}
}

// 科學官(百分比)與科學家(固定點數)是**不同的兩個技能**,落在不同欄位。
// 這條防的是「名字像就混用」——一個是 %+d%%、一個是 %+d。
func TestScienceLeaderAndResearcherAreDifferentSkills(t *testing.T) {
	var a engine.ColonyState
	applyLeaderColonyBonuses([]Leader{{Name: "甲", Skill: "科學官", Level: 3, Ship: false, Tier: 1}}, &a)
	var b engine.ColonyState
	applyLeaderColonyBonuses([]Leader{{Name: "乙", Skill: "科學家", Level: 3, Ship: false, Tier: 1}}, &b)
	if a.ResearchBonusPercent <= 0 || a.FlatResearch != 0 {
		t.Errorf("科學官應只加百分比:%%=%d、固定=%d", a.ResearchBonusPercent, a.FlatResearch)
	}
	if b.FlatResearch <= 0 || b.ResearchBonusPercent != 0 {
		t.Errorf("科學家應只加固定點數:固定=%d、%%=%d", b.FlatResearch, b.ResearchBonusPercent)
	}
}
