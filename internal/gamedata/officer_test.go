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
