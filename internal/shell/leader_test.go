package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestApplyLeaderColonyBonuses_Researcher 科學家(admin,SKILL_RESEARCHER,common code6 base5)
// 套進殖民地 FlatResearch。Level=5→expLevel=4(leaderDisplayLevelToExpLevel),Tier=1:
// bonus = 5*(4+1) = 25(非百分比技能,Researcher 是固定研究點數,見 leaderSkillIDByName 註解)。
func TestApplyLeaderColonyBonuses_Researcher(t *testing.T) {
	leaders := []Leader{{Name: "馮·諾伊曼", Skill: "科學家", Level: 5, Ship: false, Tier: 1}}
	colony := engine.ColonyState{FlatResearch: 10}

	applyLeaderColonyBonuses(leaders, &colony)

	want := 10 + 25
	if colony.FlatResearch != want {
		t.Errorf("FlatResearch = %d, want %d", colony.FlatResearch, want)
	}
}

// TestApplyLeaderColonyBonuses_Trader 貿易家(admin,SKILL_TRADER,common code9 base10,%技能)
// 套進 IncomeBonusPercent。Level=4→expLevel=3,Tier=1:bonus = 10*(3+1) = 40。
func TestApplyLeaderColonyBonuses_Trader(t *testing.T) {
	leaders := []Leader{{Name: "洛克斐勒", Skill: "貿易家", Level: 4, Ship: false, Tier: 1}}
	colony := engine.ColonyState{IncomeBonusPercent: 50} // 模擬已建太空港+50%

	applyLeaderColonyBonuses(leaders, &colony)

	want := 50 + 40
	if colony.IncomeBonusPercent != want {
		t.Errorf("IncomeBonusPercent = %d, want %d", colony.IncomeBonusPercent, want)
	}
}

// TestApplyLeaderColonyBonuses_ShipOfficerSkipped 艦艇軍官(Ship=true)不影響殖民地,即使技能
// 名稱查得到 id(工程師→SKILL_ENGINEER)也不加總——ship officer 的效果走 applyLeaderShipBonuses
// (目前艦艇/戰鬥迴圈尚未整合,見 ship.go 的 With Officer 系列函式)。
func TestApplyLeaderColonyBonuses_ShipOfficerSkipped(t *testing.T) {
	leaders := []Leader{{Name: "圖靈", Skill: "工程師", Level: 3, Ship: true, Tier: 1}}
	colony := engine.ColonyState{FlatResearch: 5, IncomeBonusPercent: 5}

	applyLeaderColonyBonuses(leaders, &colony)

	if colony.FlatResearch != 5 || colony.IncomeBonusPercent != 5 {
		t.Errorf("colony changed for Ship=true leader: FlatResearch=%d IncomeBonusPercent=%d, want unchanged (5,5)",
			colony.FlatResearch, colony.IncomeBonusPercent)
	}
}

// TestApplyLeaderColonyBonuses_UnmappedSkillSkipped 「指揮官」沒有可對應的技能 id
// (leaderSkillIDByName 刻意不收錄,見該表註解),不應套用任何加成、也不應 panic。
func TestApplyLeaderColonyBonuses_UnmappedSkillSkipped(t *testing.T) {
	leaders := []Leader{{Name: "漢尼拔", Skill: "指揮官", Level: 6, Ship: false, Tier: 1}} // 假設誤設為殖民地領袖也要安全略過
	colony := engine.ColonyState{FlatResearch: 7}

	applyLeaderColonyBonuses(leaders, &colony)

	if colony.FlatResearch != 7 {
		t.Errorf("FlatResearch = %d, want unchanged 7", colony.FlatResearch)
	}
}

// TestApplyLeaderColonyBonuses_NoLeadersNoop 空領袖清單不影響殖民地(回歸既有零值行為)。
func TestApplyLeaderColonyBonuses_NoLeadersNoop(t *testing.T) {
	colony := engine.ColonyState{FlatResearch: 3, IncomeBonusPercent: 8}
	applyLeaderColonyBonuses(nil, &colony)

	if colony.FlatResearch != 3 || colony.IncomeBonusPercent != 8 {
		t.Errorf("colony changed with nil leaders: FlatResearch=%d IncomeBonusPercent=%d", colony.FlatResearch, colony.IncomeBonusPercent)
	}
}

// TestLeaderDisplayLevelToExpLevel 驗證 Level(1..5 顯示等級,demoLeaders 目前有超出上限的6)
// 到 gamedata.LeaderSkillBonus 用的 expLevel(0..4)換算,見函式註解的 Level-1 夾值規則。
func TestLeaderDisplayLevelToExpLevel(t *testing.T) {
	cases := []struct{ level, want int }{
		{1, 0}, {3, 2}, {5, 4}, {6, 4}, {0, 0}, {-1, 0},
	}
	for _, c := range cases {
		if got := leaderDisplayLevelToExpLevel(c.level); got != c.want {
			t.Errorf("leaderDisplayLevelToExpLevel(%d) = %d, want %d", c.level, got, c.want)
		}
	}
}

// 舊路徑(只有中文標籤的 demo 領袖)仍要能解出技能 id。
//
// 2026-08-08(第 45 項(領袖技能))改寫:原本這支釘的是「指揮官刻意未收錄」,而那個斷言本身是錯的
// ——「指揮官」是 SKILL_COMMANDO 的譯名,只是它的消費端在地面戰而不是殖民地經濟。
// 現在標籤表在 gamedata(27 個技能全收),「有沒有效果」改由 applyLeaderColonyBonuses
// 的 switch 決定,兩件事分開。
func TestLeaderSkillLabelFallbackResolvesIDs(t *testing.T) {
	cases := []struct {
		name string
		want int
	}{
		{"科學家", int(gamedata.SKILL_RESEARCHER)},
		{"貿易家", int(gamedata.SKILL_TRADER)},
		{"工程師", int(gamedata.SKILL_ENGINEER)},
		{"指揮官", int(gamedata.SKILL_COMMANDO)},
	}
	for _, c := range cases {
		l := Leader{Name: "某人", Skill: c.name, Level: 3, Tier: 1}
		if got := leaderSkillTier(l, c.want); got != 1 {
			t.Errorf("標籤 %q 應解出技能 %#x 的階 1,得到 %d", c.name, c.want, got)
		}
	}
	// 查不到的標籤誠實回 nil,不臆造技能。
	if got := leaderSkills(Leader{Name: "某人", Skill: "不存在的技能", Tier: 1}); got != nil {
		t.Errorf("無法對應的標籤應回 nil,得到 %+v", got)
	}
}

// 帶了 Skills 的領袖不再看標籤——標籤是翻譯過的,不能當識別鍵。
//
// 這一條釘的是一個會完全靜默的失效:英文模式下 `Skill` 存的是 "Scientist",
// 舊的中文標籤比對一個都不成立,**所有領袖加成同時消失**而畫面上毫無異狀。
func TestSkillIDsWinOverTheDisplayLabel(t *testing.T) {
	l := Leader{
		Name:  "Von Neumann",
		Skill: "Scientist", // 英文模式的顯示標籤,查不到中文表
		Level: 3,
		Skills: []LeaderSkill{
			{ID: int(gamedata.SKILL_RESEARCHER), Tier: 1},
		},
	}
	if got := leaderSkillTier(l, int(gamedata.SKILL_RESEARCHER)); got != 1 {
		t.Errorf("有 Skills 時應直接讀 id,得到階 %d", got)
	}
	var c engine.ColonyState
	applyLeaderColonyBonuses([]Leader{l}, &c)
	if c.FlatResearch <= 0 {
		t.Errorf("英文標籤的科學家仍應加研究點數,得到 %d", c.FlatResearch)
	}
}

// 一位領袖可以同時有好幾項技能——HERODATA 的技能欄位是每技能 2 bit,不是一人一技。
func TestOneLeaderCanCarrySeveralSkills(t *testing.T) {
	l := Leader{
		Name:  "多才",
		Skill: "農業官",
		Level: 3,
		Skills: []LeaderSkill{
			{ID: int(gamedata.SKILL_FARMING_LEADER), Tier: 1},
			{ID: int(gamedata.SKILL_LABOR_LEADER), Tier: 1},
			{ID: int(gamedata.SKILL_RESEARCHER), Tier: 1},
		},
	}
	var c engine.ColonyState
	applyLeaderColonyBonuses([]Leader{l}, &c)
	if c.FoodBonusPercent <= 0 || c.IndustryBonusPercent <= 0 || c.FlatResearch <= 0 {
		t.Errorf("三項技能應各自生效,得到 食物%%=%d 工業%%=%d 研究=%d",
			c.FoodBonusPercent, c.IndustryBonusPercent, c.FlatResearch)
	}
	// 正對照:沒帶的技能不會憑空出現。
	if c.MoralePercent != 0 {
		t.Errorf("沒有心靈導師技能,士氣不該變動,得到 %d", c.MoralePercent)
	}
}

// 進階技能(tier 2)要比一般階強 50%——Tier 寫死 1 的時候這個差異一次都沒發生過。
func TestAdvancedTierIsStrongerThanNormal(t *testing.T) {
	mk := func(tier int) engine.ColonyState {
		var c engine.ColonyState
		applyLeaderColonyBonuses([]Leader{{
			Name: "某人", Level: 3,
			Skills: []LeaderSkill{{ID: int(gamedata.SKILL_FARMING_LEADER), Tier: tier}},
		}}, &c)
		return c
	}
	normal, advanced := mk(1), mk(2)
	if advanced.FoodBonusPercent <= normal.FoodBonusPercent {
		t.Errorf("進階階應比一般階強:一般 %d%%、進階 %d%%",
			normal.FoodBonusPercent, advanced.FoodBonusPercent)
	}
	if want := normal.FoodBonusPercent + normal.FoodBonusPercent/2; advanced.FoodBonusPercent != want {
		t.Errorf("進階階應是一般階 +50%%(%d%%),得到 %d%%", want, advanced.FoodBonusPercent)
	}
}
