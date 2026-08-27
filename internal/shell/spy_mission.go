package shell

// spy_mission.go:玩家對每個 AI 對手的間諜任務選擇。
//
// 原作的種族關係畫面有 Espionage、Sabotage、Hide 三個任務選項。原始反組譯已確認
// SABOTAGE 的命中門檻與建築清除呼叫鏈；remake 因此開放三種任務。原版三顆按鈕的
// 左右順序仍未知，UI 使用明確標籤的循環控制，不把未解出的座標語意冒充已證實。
// 舊存檔或外部資料若帶入未支援的值，一律退回 STEAL，保持結算失敗即關閉。

// SpyMission 是派往單一 AI 對手的間諜任務。
type SpyMission uint8

const (
	// SpyMissionSteal 是 Espionage：嘗試偷取我方尚未擁有的科技。
	SpyMissionSteal SpyMission = iota
	// SpyMissionSabotage 依原版 `0x1014A4`／`0x10130A` 執行破壞建築。
	SpyMissionSabotage
	// SpyMissionHide 是 HIDE：保持隱蔽，同時參與 Spy vs Spy。
	SpyMissionHide
)

// SpyMissionSupported 回傳 remake 目前有完整結算規則的任務。
func SpyMissionSupported(m SpyMission) bool {
	return m == SpyMissionSteal || m == SpyMissionSabotage || m == SpyMissionHide
}

// normalizedSpyMission 對不支援或未知的值採取保守退回。
func normalizedSpyMission(m SpyMission) SpyMission {
	if !SpyMissionSupported(m) {
		return SpyMissionSteal
	}
	return m
}

// ensurePlayerSpyMissions 讓任務陣列與 AIPlayers 平行。舊存檔沒有此欄位時，
// 零值 STEAL 正好是原本「所有間諜都偷科技」的相容預設。
func (s *GameSession) ensurePlayerSpyMissions() {
	for len(s.PlayerSpyMissions) < len(s.AIPlayers) {
		s.PlayerSpyMissions = append(s.PlayerSpyMissions, SpyMissionSteal)
	}
}

// SpyMissionFor 取得派往 targetIdx 的任務。越界與未支援值都採 STEAL。
func (s *GameSession) SpyMissionFor(targetIdx int) SpyMission {
	if targetIdx < 0 || targetIdx >= len(s.AIPlayers) {
		return SpyMissionSteal
	}
	s.ensurePlayerSpyMissions()
	return normalizedSpyMission(s.PlayerSpyMissions[targetIdx])
}

// SetSpyMission 設定派往 targetIdx 的任務。
//
// 三種任務都已有可追溯的最小結算規則；未知任務值仍拒絕寫入。
func (s *GameSession) SetSpyMission(targetIdx int, mission SpyMission) bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdSetSpyMission, Args: []int{targetIdx, int(mission)}})
	if targetIdx < 0 || targetIdx >= len(s.AIPlayers) || !SpyMissionSupported(mission) {
		return false
	}
	s.ensurePlayerSpyMissions()
	s.PlayerSpyMissions[targetIdx] = mission
	return true
}

// CycleSpyMission 在 STEAL → SABOTAGE → HIDE → STEAL 之間切換，供種族關係畫面使用。
func (s *GameSession) CycleSpyMission(targetIdx int) (SpyMission, bool) {
	current := s.SpyMissionFor(targetIdx)
	next := SpyMissionSteal
	switch current {
	case SpyMissionSteal:
		next = SpyMissionSabotage
	case SpyMissionSabotage:
		next = SpyMissionHide
	}
	if !s.SetSpyMission(targetIdx, next) {
		return current, false
	}
	return next, true
}
