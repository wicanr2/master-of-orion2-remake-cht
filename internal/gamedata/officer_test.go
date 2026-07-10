package gamedata

import "testing"

func TestLeaderExpLevel(t *testing.T) {
	cases := []struct{ exp, want int }{
		{0, 0}, {59, 0}, {60, 1}, {149, 1}, {150, 2}, {299, 2},
		{300, 3}, {499, 3}, {500, 4}, {10000, 4},
	}
	for _, c := range cases {
		if got := LeaderExpLevel(c.exp); got != c.want {
			t.Errorf("LeaderExpLevel(%d) = %d,預期 %d", c.exp, got, c.want)
		}
	}
}

func TestLeaderSkillBonus(t *testing.T) {
	const (
		assassin       = 0x00 // common code0 base 2
		diplomat       = 0x02 // common code2 base 10
		megawealth     = 0x04 // common code4 base 10(不隨等級倍增)
		engineer       = 0x10 // captain code0 base 2
		navigator      = 0x14 // captain,專屬值表
		environmentali = 0x20 // admin code0 base -10
	)
	cases := []struct {
		name                     string
		id, tier, expLevel, want int
	}{
		{"tier0 無技能", assassin, 0, 4, 0},
		{"暗殺 tier1 exp0", assassin, 1, 0, 2},          // 2*(0+1)
		{"暗殺 tier1 exp4", assassin, 1, 4, 10},         // 2*5
		{"暗殺 tier2 exp4 進階+50%", assassin, 2, 4, 15},  // 2*5=10 +5
		{"外交 tier1 exp1", diplomat, 1, 1, 20},         // 10*2
		{"外交 tier2 exp1", diplomat, 2, 1, 30},         // 20 +10
		{"鉅富 tier1 exp0 不倍增", megawealth, 1, 0, 10},   // base 10,不 *(exp+1)
		{"鉅富 tier2 exp3 僅+50%", megawealth, 2, 3, 15}, // 10 +5
		{"工程師 tier1 exp0", engineer, 1, 0, 2},         // captain base 2
		{"環保 tier1 exp0 負值", environmentali, 1, 0, -10},
		{"領航 tier1 exp0", navigator, 1, 0, 1}, // navigatorSkillValues[0][0]
		{"領航 tier2 exp4", navigator, 2, 4, 4}, // navigatorSkillValues[1][4]
		{"領航 tier1 exp2", navigator, 1, 2, 2}, // navigatorSkillValues[0][2]
	}
	for _, c := range cases {
		if got := LeaderSkillBonus(c.id, c.tier, c.expLevel); got != c.want {
			t.Errorf("%s: LeaderSkillBonus(%#x,%d,%d) = %d,預期 %d",
				c.name, c.id, c.tier, c.expLevel, got, c.want)
		}
	}
}

func TestLeaderSkillTier(t *testing.T) {
	// bit 位置 = skillnum*2,每技能佔 2 bit(tier 0-3)。
	// commonSkills: code0(Assassin)=tier1(0b01)、code6(Researcher)=tier2(0b10) → bits = 0b10_...(code6) | 0b01(code0)
	commonSkills := uint32(1) | uint32(2)<<(2*6) // Assassin tier1, Researcher tier2
	specialSkills := uint32(3) << (2 * 3)        // code3 tier3(進階上限,captain=Helmsman/admin=Instructor)

	cases := []struct {
		name       string
		skillID    int
		leaderType int
		want       int
	}{
		{"common Assassin tier1(與 leaderType 無關)", int(SKILL_ASSASSIN), LeaderTypeAdmin, 1},
		{"common Researcher tier2", int(SKILL_RESEARCHER), LeaderTypeCaptain, 2},
		{"common 未設技能 tier0", int(SKILL_TRADER), LeaderTypeAdmin, 0},
		{"captain Helmsman tier3,leaderType=Captain 才生效", int(SKILL_HELMSMAN), LeaderTypeCaptain, 3},
		{"captain Helmsman,leaderType=Admin 視為 0(specialSkills 不讀 captain 表)", int(SKILL_HELMSMAN), LeaderTypeAdmin, 0},
		{"admin Instructor tier3,leaderType=Admin 才生效", int(SKILL_INSTRUCTOR), LeaderTypeAdmin, 3},
		{"admin Instructor,leaderType=Captain 視為 0", int(SKILL_INSTRUCTOR), LeaderTypeCaptain, 0},
	}
	for _, c := range cases {
		if got := LeaderSkillTier(c.skillID, c.leaderType, commonSkills, specialSkills); got != c.want {
			t.Errorf("%s: LeaderSkillTier(%#x,%d,...) = %d,預期 %d",
				c.name, c.skillID, c.leaderType, got, c.want)
		}
	}
}

func TestLeaderMaintenanceCost(t *testing.T) {
	cases := []struct {
		name          string
		hireCost      int
		hasMegawealth bool
		want          int
	}{
		{"Megawealth 免費", 500, true, 0},
		{"無 Megawealth,ceil(299/100)=3", 299, false, 3},
		{"無 Megawealth,整除 200/100=2", 200, false, 2},
		{"無 Megawealth,下限1(hireCost=0)", 0, false, 1},
	}
	for _, c := range cases {
		if got := LeaderMaintenanceCost(c.hireCost, c.hasMegawealth); got != c.want {
			t.Errorf("%s: LeaderMaintenanceCost(%d,%v) = %d,預期 %d",
				c.name, c.hireCost, c.hasMegawealth, got, c.want)
		}
	}
}

func TestLeaderHireModifier(t *testing.T) {
	cases := []struct {
		name    string
		bonuses []int
		want    int
	}{
		{"無 Famous 領袖,修正0", nil, 0},
		{"單一 Famous,取該負值", []int{-60}, -60},
		{"多個 Famous,取最負(效果最強)", []int{-60, -300, -120}, -300},
		{"混入 tier0 的0(不影響 MIN)", []int{0, -60, 0}, -60},
	}
	for _, c := range cases {
		if got := LeaderHireModifier(c.bonuses); got != c.want {
			t.Errorf("%s: LeaderHireModifier(%v) = %d,預期 %d", c.name, c.bonuses, got, c.want)
		}
	}
}
