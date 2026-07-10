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

// 技能 id 完整列舉已在 enums.go 生成(LeaderSkills 型別 + SKILL_ASSASSIN..SKILL_TACTICS,對照
// openorion2/src/gamestate.h:602-631,見 enums_test.go TestEnumSpotValues 的回歸抽查),此檔不重
// 複定義,直接引用 int(SKILL_MEGAWEALTH)/int(SKILL_NAVIGATOR) 供下面 LeaderSkillBonus 特例判斷。
const (
	skillMegawealth = int(SKILL_MEGAWEALTH)
	skillNavigator  = int(SKILL_NAVIGATOR)
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

// 領袖類型(gamestate.h:32-33 LEADER_TYPE_CAPTAIN/LEADER_TYPE_ADMIN)。save.Leader.Type 直接是
// 這個值(0/1),對照 openorion2 存檔欄位語意。
const (
	LeaderTypeCaptain = 0
	LeaderTypeAdmin   = 1
)

// LeaderSkillTier 依技能 id 對應的技能位元組(commonSkills 或 specialSkills,依 leaderType 決定
// specialSkills 是否生效)解出該技能的階(0 無/1 一般/2 進階)。對照 Leader::hasSkill
// (gamestate.cpp:631-662):
//   - common 技能(id 落在 COMMON_SKILLS_TYPE 範圍)一律讀 commonSkills,與 leaderType 無關。
//   - captain 技能只在 leaderType==LeaderTypeCaptain 時讀 specialSkills,否則視為 0(該領袖不是
//     艦艇軍官,不可能有這個技能)。
//   - admin 技能只在 leaderType==LeaderTypeAdmin 時讀 specialSkills,否則視為 0。
//   - 每個技能佔 2 bit(tier 0-2,3 保留未用),bit 位置 = skillnum*2。
//
// 供讀存檔的真實 Leader(save.Leader.CommonSkills/SpecialSkills)算出真實技能階,取代 demo 資料
// 手動指定 tier 的權宜作法。
func LeaderSkillTier(skillID, leaderType int, commonSkills, specialSkills uint32) int {
	skillType := (skillID & 0x30) >> 4
	code := skillID & 0x0f

	var bits uint32
	var max int

	switch skillType {
	case 0: // COMMON_SKILLS_TYPE
		bits = commonSkills
		max = len(baseSkillValues[0])
	case 1: // CAPTAIN_SKILLS_TYPE
		max = len(baseSkillValues[1])
		if leaderType == LeaderTypeCaptain {
			bits = specialSkills
		}
	case 2: // ADMIN_SKILLS_TYPE
		max = len(baseSkillValues[2])
		if leaderType == LeaderTypeAdmin {
			bits = specialSkills
		}
	default:
		return 0
	}

	if code >= max {
		return 0
	}
	return int((bits >> uint(2*code)) & 0x3)
}

// LeaderMaintenanceCost 領袖維護費(GameState::leaderMaintenanceCost,gamestate.cpp:2428-2441):
// 有 Megawealth 技能(任一階)者維護免費;否則 = ceil(hireCost/100),下限 1。
// hireCost 由 LeaderHireCost 算好傳入。不移植 LEADER_ID_LOKNAR(=65,固定角色 Loknar 的硬編免費
// 特例)——那是特定英雄 ID 的例外,不是技能規則,remake 目前也沒有這個具名英雄。
func LeaderMaintenanceCost(hireCost int, hasMegawealth bool) int {
	if hasMegawealth {
		return 0
	}
	ret := (hireCost + 99) / 100
	if ret < 1 {
		return 1
	}
	return ret
}

// LeaderHireModifier 依帝國內「已受雇領袖」的 Famous 技能加成算出雇用費修正(統一套用給該玩家
// 之後所有雇用報價)。對照 GameState::leaderHireModifier(gamestate.cpp:2407-2426):
//
//	「The bonus is not cumulative, only the leader with the highest effect counts」
//
// Famous 的 skillBonus 恆為負值(base=-60,見 baseSkillValues[0][3]),原碼用 MIN 取「效果最強」
// (最負)的那個,起始值 0(無 Famous 領袖時修正為 0,不打折也不加價)。famousBonuses 由呼叫端
// 對每個「已受雇」(isEmployed:Idle/Working/Unassigned)的領袖,以
// LeaderSkillBonus(int(SKILL_FAMOUS), tier, expLevel) 算好整理成陣列傳入(tier=0 的領袖不必放
// 進來,結果一樣是 0,放不放不影響 MIN)。
func LeaderHireModifier(famousBonuses []int) int {
	ret := 0
	for _, b := range famousBonuses {
		if b < ret {
			ret = b
		}
	}
	return ret
}
