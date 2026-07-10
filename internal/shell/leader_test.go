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
	leaders := []Leader{{"馮·諾伊曼", "科學家", 5, false, 1}}
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
	leaders := []Leader{{"洛克斐勒", "貿易家", 4, false, 1}}
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
	leaders := []Leader{{"圖靈", "工程師", 3, true, 1}}
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
	leaders := []Leader{{"漢尼拔", "指揮官", 6, false, 1}} // 假設誤設為殖民地領袖也要安全略過
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

// TestLeaderSkillIDByNameMapping 回歸鎖:確保三個已收錄的中文標籤仍對應正確的 gamedata 技能 id,
// 且「指揮官」仍刻意未收錄(見表格註解——沒有唯一對應、不臆測)。
func TestLeaderSkillIDByNameMapping(t *testing.T) {
	cases := []struct {
		name string
		want int
	}{
		{"科學家", int(gamedata.SKILL_RESEARCHER)},
		{"貿易家", int(gamedata.SKILL_TRADER)},
		{"工程師", int(gamedata.SKILL_ENGINEER)},
	}
	for _, c := range cases {
		got, ok := leaderSkillIDByName[c.name]
		if !ok {
			t.Errorf("leaderSkillIDByName[%q] 找不到,預期 %d", c.name, c.want)
			continue
		}
		if got != c.want {
			t.Errorf("leaderSkillIDByName[%q] = %d, want %d", c.name, got, c.want)
		}
	}
	if _, ok := leaderSkillIDByName["指揮官"]; ok {
		t.Errorf(`leaderSkillIDByName["指揮官"] 不應收錄(無唯一對應技能),但找到了`)
	}
}
