package gamedata

// 軍官(Leader)數值公式,移植自 openorion2 gamestate.cpp:607-701(SA1 盤點標記為唯讀正確可複用)。
// 技能 id 編碼(gamestate.h):bit4-5 = 類型(0 common / 1 captain / 2 admin),bit0-3 = 技能碼。

// 經驗門檻(leaderExpThresholds):跨過門檻升一級,最高 4 級(MAX_LEADER_LEVELS=5,級 0..4)。
var leaderExpThresholds = []int{60, 150, 300, 500}

// LeaderExpLevel 依累積經驗回傳等級(0..4)。對照 Leader::expLevel。
//
//	exp <60→0,60..149→1,150..299→2,300..499→3,≥500→4
func LeaderExpLevel(experience int) int {
	for i, t := range leaderExpThresholds {
		if experience < t {
			return i
		}
	}
	return len(leaderExpThresholds)
}

// 各技能類型的基礎技能值 baseSkillValues[type][code](gamestate.cpp:75)。
// row0 common(10)、row1 captain(8)、row2 admin(9)。
var baseSkillValues = [3][]int{
	{2, 2, 10, -60, 10, 2, 5, 2, 2, 10},
	{2, 5, 5, 5, 1, 5, 2, 5},
	{-10, 10, 10, 1, 10, 10, 10, 5, 2},
}

// 領航(Navigator)技能特殊值 navigatorSkillValues[tier>1?1:0][expLevel]。
var navigatorSkillValues = [2][]int{
	{1, 1, 2, 2, 3},
	{1, 1, 3, 3, 4},
}

// 有特殊進程規則的技能 id(gamestate.h LeaderSkills enum)。
const (
	skillMegawealth = 0x04 // common,加成不隨等級倍增
	skillNavigator  = 0x14 // captain,用專屬值表
)

// LeaderSkillBonus 依技能 id、技能階(tier:0 無/1 一般/2 進階)、經驗等級回傳該技能加成。
// 對照 Leader::skillBonus:
//   - tier 0 → 0
//   - Navigator 用專屬值表(不隨下列通則)
//   - 一般技能:base = baseSkillValues[type][code];除 Megawealth 外乘以 (expLevel+1);tier>1 再 +50%
func LeaderSkillBonus(skillID, tier, expLevel int) int {
	if tier <= 0 {
		return 0
	}
	if skillID == skillNavigator {
		idx := 0
		if tier > 1 {
			idx = 1
		}
		return navigatorSkillValues[idx][expLevel]
	}
	skillType := (skillID & 0x30) >> 4
	code := skillID & 0x0f
	ret := baseSkillValues[skillType][code]
	if skillID != skillMegawealth {
		ret *= expLevel + 1
	}
	if tier > 1 {
		ret += ret / 2 // 進階技能 +50%
	}
	return ret
}
