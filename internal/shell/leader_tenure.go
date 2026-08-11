package shell

// leader_tenure.go:原版領袖解除後的閒置記數與 active ETA。
//
// 這裡的「任期門檻」只指目前已由 IDA 靜態證實的 raw status=4 路徑：
// `Deassign_Officer @ 0x934CF` 每回合遞增 raw +0x37，signed 值達 30 後呼叫
// `Check_Officer_Fields @ 0x933F2`，後者把欄位清回未任職狀態。它不是把所有
// 活躍領袖硬設成 30 回合任期。原版同一條迴圈對 status=1 的領袖遞減 +0x37；
// ETA 從 1 變 0 且 location(+0x35)==1 時會進 Colony_Calculation。後一個重算
// 回呼的完整 remake 資料流仍保留為待 oracle，不把它誤寫成「到期就解雇」。

const (
	originalLeaderLimboStatus    = 4
	originalLeaderLimboThreshold = 30
	originalLeaderActiveStatus   = 1
)

// advanceActiveLeaderETA 對應 Deassign_Officer @ 0x934CF 的 status=1 分支：
// `+0x37 > 0` 時先遞減；若原值為 1 且 location(+0x35)==1，原版隨即呼叫
// Colony_Calculation。這裡保留 raw ETA 變更，並以 remake 近似 callback 重整
// 已指派殖民地的領袖衍生欄位／士氣；回傳「剛歸零且在殖民地」的人數供抽樣測試。
type leaderETAExpiry struct {
	Leader Leader
}

func advanceActiveLeaderETAWithRecords(leaders []Leader) []leaderETAExpiry {
	var expired []leaderETAExpiry
	for i := range leaders {
		leader := &leaders[i]
		if leader.RawStatus != originalLeaderActiveStatus || leader.RawETA <= 0 {
			continue
		}
		leader.RawETA--
		if leader.RawETA == 0 && leader.RawLocation == 1 {
			expired = append(expired, leaderETAExpiry{Leader: *leader})
		}
	}
	return expired
}

func advanceActiveLeaderETA(leaders []Leader) int {
	return len(advanceActiveLeaderETAWithRecords(leaders))
}

// applyLeaderETACallback 是原版 Colony_Calculation 的 remake 消費端。
//
// IDA 已確認 raw `sub_E2AB1 @ 0xE2AB1` 會掃描六個 `+0x48..+0x52` 槽，對符合
// empire／active 條件的資料依序呼叫 `sub_E1D59`、`sub_DF8F0`，最後呼叫
// `sub_E2710` 重算帝國衍生欄位；這些 raw 設計／艦隊／殖民地欄位沒有可安全對回
// remake 的一對一結構。故以「撤銷目前領袖增量 → 依同一份 Leader 重新套用 →
// 重算所有殖民地士氣」作為完整玩家可感知消費端的近似。領袖仍保留在職與殖民地指派，
// 不把 ETA=0 誤解成解雇。回傳實際刷新的殖民地數。
func (s *GameSession) applyLeaderETACallback(leader Leader) int {
	if s == nil || leader.Name == "" {
		return 0
	}
	s.ensureColonyLeaderSlots()
	refreshed := 0
	bonus := colonyLeaderBonusFor(leader)
	for i, assigned := range s.ColonyLeaderNames {
		if assigned != leader.Name || i >= len(s.PlayerColonies) {
			continue
		}
		applyColonyLeaderBonusDelta(&s.PlayerColonies[i], bonus, -1)
		applyColonyLeaderBonusDelta(&s.PlayerColonies[i], bonus, 1)
		s.recalcColonyMorale(i)
		refreshed++
	}
	// sub_E2710 會重建帝國層衍生欄位；remake 的安全對應是把所有殖民地的
	// 士氣快取一併刷新，避免 callback 發生在非目前選中殖民地時留下 stale 值。
	if leader.RawLocation == 1 {
		s.recalcAllColonyMorale()
	}
	return refreshed
}

// advanceLeaderLimbo 推進玩家側 raw status=4 領袖的閒置計數，回傳本回合
// 因達門檻而從 Leader Pool 清除的人數。RawStatus=0 的 demo／舊 JSON 領袖
// 不會被誤套這條原版 raw 路徑。
func (s *GameSession) advanceLeaderLimbo() int {
	// 原版領袖表是全域 67 筆；remake 依擁有者拆在玩家／AI slice，兩邊都要
	// 推進 active ETA。玩家側 location=1 的到期項目觸發 Colony_Calculation 近似
	// callback；AI 沒有殖民地領袖指派平行表，仍只保留 raw ETA 遞減。
	for _, expired := range advanceActiveLeaderETAWithRecords(s.Leaders) {
		s.applyLeaderETACallback(expired.Leader)
	}
	for i := range s.AIPlayers {
		advanceActiveLeaderETA(s.AIPlayers[i].Leaders)
	}
	released := 0
	for i := len(s.Leaders) - 1; i >= 0; i-- {
		leader := &s.Leaders[i]
		if leader.RawStatus != originalLeaderLimboStatus {
			continue
		}
		leader.RawETA++
		if leader.RawETA < originalLeaderLimboThreshold {
			continue
		}

		name := leader.Name
		// 不呼叫公開的 Unassign/Dismiss API：那些是玩家操作，會記錄一筆
		// PlayerCommand；這裡是世界回合自動清理。先反向撤銷殖民地領袖增量，
		// 再清除艦艇指派，避免「人不見了、效果還留著」。
		s.ensureColonyLeaderSlots()
		for colonyIdx, assigned := range s.ColonyLeaderNames {
			if assigned != name {
				continue
			}
			applyColonyLeaderBonusDelta(&s.PlayerColonies[colonyIdx], colonyLeaderBonusFor(*leader), -1)
			s.ColonyLeaderNames[colonyIdx] = ""
		}
		s.clearOfficerAssignment(name)
		copy(s.Leaders[i:], s.Leaders[i+1:])
		s.Leaders = s.Leaders[:len(s.Leaders)-1]
		released++
	}
	return released
}
