package shell

// GameSettings 是原版對局內 SETTINGS 分頁的 13 個玩家偏好。
// Version 用來區分「舊 JSON 沒有設定區」與「玩家真的把全部開關關閉」。
type GameSettings struct {
	Version                    int  `json:"version"`
	EndOfTurnSummary           bool `json:"endOfTurnSummary"`
	EndOfTurnWait              bool `json:"endOfTurnWait"`
	EnemyMoves                 bool `json:"enemyMoves"`
	ExpandingHelp              bool `json:"expandingHelp"`
	AutoSelectShips            bool `json:"autoSelectShips"`
	Animations                 bool `json:"animations"`
	AutoSelectColony           bool `json:"autoSelectColony"`
	ShowRelocationLines        bool `json:"showRelocationLines"`
	ShowGNNReport              bool `json:"showGNNReport"`
	AutoDeleteTradeGoodHousing bool `json:"autoDeleteTradeGoodHousing"`
	AutoSaveGame               bool `json:"autoSaveGame"`
	ShowOnlySeriousTurnSummary bool `json:"showOnlySeriousTurnSummary"`
	ShipInitiative             bool `json:"shipInitiative"`
}

const gameSettingsVersion = 1

// DefaultGameSettings 精確對應 Orion2.exe sub_127E1 @ 0x127F4..0x1284F。
func DefaultGameSettings() GameSettings {
	return GameSettings{
		Version:          gameSettingsVersion,
		EndOfTurnSummary: true, EndOfTurnWait: true,
		AutoSelectShips: true, Animations: true,
		ShowRelocationLines: true, ShowGNNReport: true,
		AutoSaveGame: true,
	}
}

// EffectiveGameSettings 為舊存檔補原版預設，並同步歷史相容欄位。
func (s *GameSession) EffectiveGameSettings() GameSettings {
	if s == nil {
		return DefaultGameSettings()
	}
	settings := s.GameSettings
	if settings.Version == 0 {
		settings = DefaultGameSettings()
		settings.ShowRelocationLines = s.ShowRelocationLines
	}
	return settings
}

// ApplyGameSettings 套用設定；遷移線同時更新既有消費端相容欄位。
func (s *GameSession) ApplyGameSettings(settings GameSettings) {
	if s == nil {
		return
	}
	settings.Version = gameSettingsVersion
	s.GameSettings = settings
	s.ShowRelocationLines = settings.ShowRelocationLines
}

// HasSeriousTurnSummaryReport 判斷本回合是否包含應觸發「只顯示重要摘要」的玩家結果。
// 飢荒、叛亂與破產處分由原版 help 直接列舉；兩種敵襲是已有 typed 結果的威脅擴充。
func (s *GameSession) HasSeriousTurnSummaryReport() bool {
	if s == nil {
		return false
	}
	for _, colony := range s.LastPlayerOutput.Colonies {
		if colony.Starving {
			return true
		}
	}
	return len(s.LastRebellions) > 0 || len(s.LastBankruptcy) > 0 ||
		s.LastAntaranNotice != nil || s.LastRaidReport != nil
}
