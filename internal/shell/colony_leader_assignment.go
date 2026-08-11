package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// ColonyLeaderNames 是 PlayerColonies 的平行指派表；空字串代表該殖民地
// 沒有殖民地領袖。用名稱是為了兼容舊存檔與既有 OfficerName 識別方式，
// 真正的 Leader 技能仍從 s.Leaders 取完整 Skills。
//
// 領袖指派不直接塞進 engine.ColonyState 的「不可逆累加」流程，而是由本檔
// 先撤銷舊領袖的精確欄位增量，再套用新領袖；因此換任／解除任命不會留下
// 看不見的科研、收入或士氣殘值。

type colonyLeaderBonus struct {
	flatResearch, income, morale, growth int
	food, industry, research, pollution  int
}

func colonyLeaderBonusFor(l Leader) colonyLeaderBonus {
	bySkill := map[int][]int{}
	expLevel := leaderDisplayLevelToExpLevel(l.Level)
	for _, sk := range leaderSkills(l) {
		if b := gamedata.LeaderSkillBonus(sk.ID, sk.Tier, expLevel); b != 0 {
			bySkill[sk.ID] = append(bySkill[sk.ID], b)
		}
	}
	var out colonyLeaderBonus
	for id, values := range bySkill {
		bonus := gamedata.LeaderSkillCombine(id, values)
		switch id {
		case int(gamedata.SKILL_RESEARCHER):
			out.flatResearch += bonus
		case int(gamedata.SKILL_TRADER), int(gamedata.SKILL_FINANCIAL_LEADER):
			out.income += bonus
		case int(gamedata.SKILL_SPIRITUAL_LEADER):
			out.morale += bonus
		case int(gamedata.SKILL_MEDICINE):
			out.growth += bonus
		case int(gamedata.SKILL_FARMING_LEADER):
			out.food += bonus
		case int(gamedata.SKILL_LABOR_LEADER):
			out.industry += bonus
		case int(gamedata.SKILL_SCIENCE_LEADER):
			out.research += bonus
		case int(gamedata.SKILL_ENVIRONMENTALIST):
			out.pollution -= bonus
		}
	}
	return out
}

func applyColonyLeaderBonusDelta(c *engine.ColonyState, b colonyLeaderBonus, sign int) {
	c.FlatResearch += sign * b.flatResearch
	c.IncomeBonusPercent += sign * b.income
	c.MoralePercent += sign * b.morale
	c.GrowthBonusSum += sign * b.growth
	c.FoodBonusPercent += sign * b.food
	c.IndustryBonusPercent += sign * b.industry
	c.ResearchBonusPercent += sign * b.research
	c.PollutionReductionPercent += sign * b.pollution
}

func (s *GameSession) ensureColonyLeaderSlots() {
	for len(s.ColonyLeaderNames) < len(s.PlayerColonies) {
		s.ColonyLeaderNames = append(s.ColonyLeaderNames, "")
	}
	if len(s.ColonyLeaderNames) > len(s.PlayerColonies) {
		s.ColonyLeaderNames = s.ColonyLeaderNames[:len(s.PlayerColonies)]
	}
}

// AssignLeaderToColony 將一名非艦艇領袖指派到指定殖民地。領袖會先從
// 其他殖民地或艦艇解除，保證一人一職；同一殖民地原領袖的增量也會撤回。
func (s *GameSession) AssignLeaderToColony(colonyIdx, leaderIdx int) bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdAssignColonyLeader, Args: []int{colonyIdx, leaderIdx}})
	if colonyIdx < 0 || colonyIdx >= len(s.PlayerColonies) || leaderIdx < 0 || leaderIdx >= len(s.Leaders) {
		return false
	}
	leader := s.Leaders[leaderIdx]
	if leader.Ship || leader.Name == "" {
		return false
	}
	s.ensureColonyLeaderSlots()
	oldName := s.ColonyLeaderNames[colonyIdx]
	if oldName == leader.Name {
		// 修復人工編輯 JSON 造成的雙重指派；重按本身不重複套用加成。
		s.clearOfficerAssignment(leader.Name)
		return true
	}
	for i, name := range s.ColonyLeaderNames {
		if name == leader.Name && i != colonyIdx {
			if old, ok := s.leaderByName(name); ok {
				applyColonyLeaderBonusDelta(&s.PlayerColonies[i], colonyLeaderBonusFor(old), -1)
			}
			s.ColonyLeaderNames[i] = ""
		}
	}
	if oldName != "" {
		if old, ok := s.leaderByName(oldName); ok {
			applyColonyLeaderBonusDelta(&s.PlayerColonies[colonyIdx], colonyLeaderBonusFor(old), -1)
		}
	}
	s.ColonyLeaderNames[colonyIdx] = leader.Name
	// 領袖不能同時是艦艇軍官；這個呼叫對殖民地領袖通常是 no-op，
	// 但能修復人工編輯 JSON 造成的雙重指派。
	s.clearOfficerAssignment(leader.Name)
	applyColonyLeaderBonusDelta(&s.PlayerColonies[colonyIdx], colonyLeaderBonusFor(leader), 1)
	return true
}

// UnassignLeaderFromColony 解除指定殖民地領袖並撤回其已套用的欄位加成。
func (s *GameSession) UnassignLeaderFromColony(colonyIdx int) bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdUnassignColonyLeader, Args: []int{colonyIdx}})
	if colonyIdx < 0 || colonyIdx >= len(s.PlayerColonies) {
		return false
	}
	s.ensureColonyLeaderSlots()
	name := s.ColonyLeaderNames[colonyIdx]
	if name == "" {
		return false
	}
	if leader, ok := s.leaderByName(name); ok {
		applyColonyLeaderBonusDelta(&s.PlayerColonies[colonyIdx], colonyLeaderBonusFor(leader), -1)
	}
	s.ColonyLeaderNames[colonyIdx] = ""
	return true
}

// ColonyLeaderFor 回傳指定殖民地目前指派的領袖。
func (s *GameSession) ColonyLeaderFor(colonyIdx int) (Leader, bool) {
	if colonyIdx < 0 || colonyIdx >= len(s.PlayerColonies) {
		return Leader{}, false
	}
	s.ensureColonyLeaderSlots()
	return s.leaderByName(s.ColonyLeaderNames[colonyIdx])
}

// AssignedColonyForLeader 回傳領袖目前服務的殖民地索引。
func (s *GameSession) AssignedColonyForLeader(name string) (int, bool) {
	if name == "" {
		return -1, false
	}
	s.ensureColonyLeaderSlots()
	for i, assigned := range s.ColonyLeaderNames {
		if assigned == name {
			return i, true
		}
	}
	return -1, false
}

// DismissColonyLeader 將殖民地領袖從人才庫移除，並先解除其殖民地任職。
func (s *GameSession) DismissColonyLeader(name string) bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdDismissColonyLeader, Text: name})
	idx := -1
	for i, l := range s.Leaders {
		if l.Name == name {
			if l.Ship {
				return false
			}
			idx = i
			break
		}
	}
	if idx < 0 {
		return false
	}
	s.ensureColonyLeaderSlots()
	for i, assigned := range s.ColonyLeaderNames {
		if assigned == name {
			s.UnassignLeaderFromColony(i)
		}
	}
	s.Leaders = append(s.Leaders[:idx], s.Leaders[idx+1:]...)
	return true
}

func (s *GameSession) leaderByName(name string) (Leader, bool) {
	if name == "" {
		return Leader{}, false
	}
	for _, l := range s.Leaders {
		if l.Name == name {
			return l, true
		}
	}
	return Leader{}, false
}
