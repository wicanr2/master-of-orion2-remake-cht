package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// officer_assignment.go:逐艦軍官指派。
//
// 原版 save.Ship 有 int16 Officer 欄位(未指派為負值),直接索引固定 67 筆
// GameState::_leaders[]。remake 以 OfficerID 保存同一個來源序號，並保留名稱作舊 JSON
// 與人工編輯的回退；空字串=未指派。改派時仍會全帝國去重。
//
// 這個檔案只處理「誰在哪艘船」及技能查詢;戰鬥、航行、修復等消費端各自呼叫
// shipOfficerSkillBonus,避免把「帝國有軍官」錯當成「這艘船有軍官」。

// AssignOfficerToShip 把玩家的艦艇軍官 leaderIndex 指派到指定艦隊的指定艦艇。
//
// 一位軍官同時只能服務一艘船。若目標船已有其他軍官,會替換目標船上的指派;
// 若該軍官原本在別艘船,會先解除舊指派。殖民地領袖不能被指派到艦艇。
func (s *GameSession) AssignOfficerToShip(fleetIndex, shipIndex, leaderIndex int) bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdAssignShipOfficer, Args: []int{fleetIndex, shipIndex, leaderIndex}})
	if fleetIndex < 0 || fleetIndex >= len(s.Fleets) || shipIndex < 0 || shipIndex >= len(s.Fleets[fleetIndex].Ships) {
		return false
	}
	if leaderIndex < 0 || leaderIndex >= len(s.Leaders) || !s.Leaders[leaderIndex].Ship || s.Leaders[leaderIndex].Name == "" {
		return false
	}

	leader := &s.Leaders[leaderIndex]
	name := leader.Name
	// 先把這位軍官從所有艦艇移除,避免舊存檔或人工編輯 JSON 造成一人多艦。
	s.unassignColonyLeaderByName(name)
	s.clearOfficerAssignment(name)
	s.Fleets[fleetIndex].Ships[shipIndex].OfficerName = name
	s.Fleets[fleetIndex].Ships[shipIndex].OfficerID = leader.ID
	return true
}

// UnassignOfficerFromShip 解除指定艦艇上的軍官。回傳是否真的有解除。
func (s *GameSession) UnassignOfficerFromShip(fleetIndex, shipIndex int) bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdUnassignShipOfficer, Args: []int{fleetIndex, shipIndex}})
	if fleetIndex < 0 || fleetIndex >= len(s.Fleets) || shipIndex < 0 || shipIndex >= len(s.Fleets[fleetIndex].Ships) {
		return false
	}
	sh := &s.Fleets[fleetIndex].Ships[shipIndex]
	if sh.OfficerName == "" {
		return false
	}
	sh.OfficerName = ""
	sh.OfficerID = 0
	return true
}

// ReturnShipOfficerToPool 把指定艦艇軍官送回軍官人才庫。
//
// remake 沒有原版「所在星系／旅行中的軍官」欄位，因此「在人才庫」以沒有
// OfficerName 指派表示。這正是手冊 p.134 所說的 POOL 動作在目前資料模型
// 中能安全表達的最小語意；不虛構旅行回合。
func (s *GameSession) ReturnShipOfficerToPool(name string) bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdReturnShipOfficer, Text: name})
	if name == "" {
		return false
	}
	for _, l := range s.Leaders {
		if l.Name == name && l.Ship {
			s.clearOfficerAssignment(name)
			return true
		}
	}
	return false
}

// DismissShipOfficer 解雇一名已雇用的艦艇軍官。
//
// 只接受 Ship=true 的領袖。殖民地領袖的經濟加成目前是套用到殖民地狀態
// 的累加值，尚沒有可證實的逐殖民地任職欄位與反向回收流程；拒絕在這裡
// 靜默刪除殖民地領袖，避免留下「人不見了、加成還在」的壞狀態。
func (s *GameSession) DismissShipOfficer(name string) bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdDismissShipOfficer, Text: name})
	if name == "" {
		return false
	}
	idx := -1
	for i, l := range s.Leaders {
		if l.Name == name {
			if !l.Ship {
				return false
			}
			idx = i
			break
		}
	}
	if idx < 0 {
		return false
	}
	s.clearOfficerAssignment(name)
	copy(s.Leaders[idx:], s.Leaders[idx+1:])
	s.Leaders = s.Leaders[:len(s.Leaders)-1]
	return true
}

func (s *GameSession) clearOfficerAssignment(name string) {
	if name == "" {
		return
	}
	for fi := range s.Fleets {
		for si := range s.Fleets[fi].Ships {
			if s.Fleets[fi].Ships[si].OfficerName == name {
				s.Fleets[fi].Ships[si].OfficerName = ""
				s.Fleets[fi].Ships[si].OfficerID = 0
			}
		}
	}
}

func (s *GameSession) unassignColonyLeaderByName(name string) {
	if name == "" {
		return
	}
	if _, ok := s.leaderByName(name); !ok {
		return
	}
	s.ensureColonyLeaderSlots()
	for i, assigned := range s.ColonyLeaderNames {
		if assigned == name {
			s.UnassignLeaderFromColony(i)
		}
	}
}

// OfficerForShip 回傳這艘船上已指派且仍在 Leader Pool 的軍官。
// 舊存檔若留下已解雇軍官名稱,會被視為未生效,不臆造技能。
func (s *GameSession) OfficerForShip(sh Ship) (Leader, bool) {
	if sh.OfficerName == "" {
		return Leader{}, false
	}
	// 新格式先以來源 ID 尋找，並以名稱交叉確認，避免人工修改 JSON 後把另一位
	// 同 ID 的舊資料誤接上。舊格式沒有 officer_id 時再走名稱回退。
	for _, l := range s.Leaders {
		if l.Ship && l.ID == sh.OfficerID && l.Name == sh.OfficerName {
			return l, true
		}
	}
	for _, l := range s.Leaders {
		if l.Name == sh.OfficerName && l.Ship {
			return l, true
		}
	}
	return Leader{}, false
}

// AssignedShipForOfficer 回傳軍官目前服務的艦隊與艦艇索引;找不到時回 -1,-1。
func (s *GameSession) AssignedShipForOfficer(name string) (fleetIndex, shipIndex int, ok bool) {
	if name == "" {
		return -1, -1, false
	}
	for fi := range s.Fleets {
		for si := range s.Fleets[fi].Ships {
			if s.Fleets[fi].Ships[si].OfficerName == name {
				return fi, si, true
			}
		}
	}
	return -1, -1, false
}

// shipOfficerSkillBonus 回傳「這艘船上的軍官」某技能值;未指派、殖民地領袖、
// 已解雇軍官或沒有該技能皆為 0。
func (s *GameSession) shipOfficerSkillBonus(sh Ship, skill gamedata.LeaderSkills) int {
	l, ok := s.OfficerForShip(sh)
	if !ok {
		return 0
	}
	tier := leaderSkillTier(l, int(skill))
	if tier <= 0 {
		return 0
	}
	return gamedata.LeaderSkillBonus(int(skill), tier, leaderDisplayLevelToExpLevel(l.Level))
}

func (s *GameSession) shipOfficerHasSkill(sh Ship, skill gamedata.LeaderSkills) bool {
	return s.shipOfficerSkillBonus(sh, skill) != 0
}

func (s *GameSession) shipOfficerMissileEvasionBonus(sh Ship) int {
	return gamedata.MissileHelmsmanEvasionBonus(s.shipOfficerSkillBonus(sh, gamedata.SKILL_HELMSMAN))
}

// fighterPilotBonusForCombat 回傳目前參戰艦隊中最高的戰機飛行員加成。
//
// 原版 `Fighter_Pilot_Bonus @0x35EAE` 逐一掃描同一帝國的戰鬥艦記錄，排除無效記錄後，
// 對每艘有軍官的艦取 `Get_Fleet_Engineer_Bonus`，最後保留最大值；該 helper 的
// `5*(level+1)`／`15*(level+1)/2` 分支與 `LeaderSkillBonus(SKILL_FIGHTER_PILOT)` 的
// 普通／進階值相符。重製的戰鬥艦隊就是目前選定的參戰集合，因此以該集合取 max，
// 不把未參戰殖民地領袖或其他艦隊的技能誤算進來。
func (s *GameSession) fighterPilotBonusForCombat() int {
	if s == nil {
		return 0
	}
	best := 0
	for _, sh := range s.Fleet().Ships {
		if bonus := s.shipOfficerSkillBonus(sh, gamedata.SKILL_FIGHTER_PILOT); bonus > best {
			best = bonus
		}
	}
	return best
}

// fleetHasAssignedSkill 判定目前參戰／航行艦隊是否有指定技能的軍官。
// 這是實際規則消費端使用的查詢,與保留給舊測試／相容的全域領袖查詢分開。
func (s *GameSession) fleetHasAssignedSkill(fleetIndex int, skill gamedata.LeaderSkills) bool {
	if fleetIndex < 0 || fleetIndex >= len(s.Fleets) {
		return false
	}
	for _, sh := range s.Fleets[fleetIndex].Ships {
		if s.shipOfficerHasSkill(sh, skill) {
			return true
		}
	}
	return false
}

func (s *GameSession) selectedFleetHasAssignedSkill(skill gamedata.LeaderSkills) bool {
	return s.fleetHasAssignedSkill(s.SelectedFleet, skill)
}

func (s *GameSession) assignedEngineerTier(fleetIndex int) int {
	if fleetIndex < 0 || fleetIndex >= len(s.Fleets) {
		return 0
	}
	best := 0
	for _, sh := range s.Fleets[fleetIndex].Ships {
		l, ok := s.OfficerForShip(sh)
		if !ok {
			continue
		}
		if t := leaderSkillTier(l, int(gamedata.SKILL_ENGINEER)); t > best {
			best = t
		}
	}
	return best
}
