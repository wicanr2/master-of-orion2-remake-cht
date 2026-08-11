package shell

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

// network.go：多人對局需要的快照、席位與指令重播邊界。
//
// 這裡不依賴 netplay。規則層只知道「一批玩家指令」與「同一個快照」；
// TCP 如何傳輸由 cmd/moo2 與 internal/netplay 負責，避免邏輯核心反向依賴傳輸層。

// MarshalSnapshot 將目前完整可重建的對局狀態序列化，供網路共同開局與回合基準使用。
func (s *GameSession) MarshalSnapshot() ([]byte, error) {
	if s == nil {
		return nil, fmt.Errorf("shell: 不能序列化 nil 對局")
	}
	b, err := json.Marshal(s.snapshot())
	if err != nil {
		return nil, fmt.Errorf("shell: 序列化對局快照: %w", err)
	}
	return b, nil
}

// NetworkStateHash 回傳適合跨玩家比對的狀態指紋。
//
// SelectedStar、SelectedFleet 與 ShowRelocationLines 是純 UI 選取／顯示偏好；
// 每位玩家本來就可以看不同的星系或關掉線條，不能因為畫面焦點不同就把公平的
// 鎖步對局判成分歧。它們仍保留在一般存檔與 StateHash 中，只有網路共識比對時
// 以固定值正規化。
func (s *GameSession) NetworkStateHash() string {
	if s == nil {
		return ""
	}
	// 網路對局會讓每台機器載入自己的真人席位：頂層 Player/Colonies 是
	// ActiveSeat 的活資料，而同一份資料也保存在 Seats[ActiveSeat]。若直接
	// 雜湊 snapshot，主機(席位 0)與客戶端(席位 1)即使剛套用同一份共同開局，
	// 也會因為本地觀看席位不同而被誤判分岔。複製 session 後固定載入席位 0，
	// 只做共識用的正規化，不改動呼叫端實際正在操作的席位。
	consensus := *s
	if s.HotseatEnabled() {
		consensus.Seats = append([]seat(nil), s.Seats...)
		if err := consensus.SetActiveSeat(0); err != nil {
			return ""
		}
	}
	snap := consensus.snapshot()
	snap.SelectedStar = -1
	snap.SelectedFleet = 0
	snap.ShowRelocationLines = false
	for i := range snap.Seats {
		snap.Seats[i].SelectedStar = -1
		snap.Seats[i].SelectedFleet = 0
	}
	b, err := json.Marshal(snap)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// RestoreSnapshot 從 MarshalSnapshot 的資料重建一局對局。
// 版本不相容時失敗即關閉，不能讓兩個版本各自猜著繼續。
func RestoreSnapshot(data []byte) (*GameSession, error) {
	var snap sessionSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("shell: 解析對局快照: %w", err)
	}
	if snap.Version != saveFormatVersion {
		return nil, fmt.Errorf("shell: 對局快照版本 %d 不相容(需 %d)", snap.Version, saveFormatVersion)
	}
	return snap.restore(), nil
}

// SetCommandRecorder 設定本回合玩家操作的記錄出口。傳 nil 可停用。
func (s *GameSession) SetCommandRecorder(fn func(PlayerCommand)) {
	if s == nil {
		return
	}
	s.commandRecorder = fn
}

// recordPlayerCommand 是各個玩家狀態入口共用的記錄點。遠端重播與世界結算
// 會暫停它，避免把重播出的操作再次送回網路。
func (s *GameSession) recordPlayerCommand(c PlayerCommand) {
	if s == nil || s.commandRecorder == nil || s.commandReplayDepth > 0 {
		return
	}
	s.commandRecorder(c)
}

// SetupNetworkWithNames 把已生成的 AI 帝國按名冊順序轉成網路真人席位。
// 第 0 席是主機；其餘席位依 names[1:] 接管 AI。未接管的 AI 仍由主機規則
// 層在每回合自動決策。回傳實際建立的席位數。
func (s *GameSession) SetupNetworkWithNames(names []string) int {
	if len(names) <= 1 {
		s.Seats, s.ActiveSeat = nil, 0
		return 1
	}
	if len(names) > MaxHotseatSeats {
		names = names[:MaxHotseatSeats]
	}
	indices := make([]int, 0, len(names)-1)
	for i := 0; i < len(names)-1 && i < len(s.AIPlayers); i++ {
		indices = append(indices, i)
	}
	got := s.SetupHotseatWithAIIndices(indices)
	for i := 0; i < got && i < len(names); i++ {
		s.Seats[i].PlayerName = names[i]
	}
	s.ActiveSeat = 0
	s.loadSeat(s.Seats[0])
	return got
}

// SetActiveSeat 切換目前被規則層視為「本方」的席位。網路重播會在每一批
// 指令套用前切席位，完成後再切回呼叫端指定的席位。
func (s *GameSession) SetActiveSeat(i int) error {
	if s == nil || len(s.Seats) == 0 {
		if i == 0 {
			return nil
		}
		return fmt.Errorf("shell: 對局沒有席位 %d", i)
	}
	if i < 0 || i >= len(s.Seats) {
		return fmt.Errorf("shell: 席位 %d 超出範圍(共 %d 席)", i, len(s.Seats))
	}
	if s.ActiveSeat >= 0 && s.ActiveSeat < len(s.Seats) {
		s.Seats[s.ActiveSeat] = s.saveSeat()
	}
	s.ActiveSeat = i
	s.loadSeat(s.Seats[i])
	return nil
}

// ApplyPlayerCommandsForSeat 將一批指令套用到指定席位，並在結束後還原原本
// 的作用中席位。指令未知或快照席位不合法時回傳錯誤，呼叫端應停止網路對局。
func (s *GameSession) ApplyPlayerCommandsForSeat(seatIndex int, cmds []PlayerCommand) error {
	if s == nil || len(s.Seats) == 0 {
		return fmt.Errorf("shell: 對局沒有可重播的席位")
	}
	if seatIndex < 0 || seatIndex >= len(s.Seats) {
		return fmt.Errorf("shell: 指令指定不存在的席位 %d", seatIndex)
	}
	current := s.ActiveSeat
	if current < 0 || current >= len(s.Seats) {
		current = 0
	}
	if current != seatIndex {
		if err := s.SetActiveSeat(seatIndex); err != nil {
			return err
		}
	}
	err := s.ApplyPlayerCommands(cmds)
	// 先保存被重播席位，再把原本的席位載回來；否則下一個玩家的指令
	// 會覆蓋剛剛重播的結果。
	s.Seats[seatIndex] = s.saveSeat()
	if current != seatIndex {
		if err2 := s.SetActiveSeat(current); err == nil {
			err = err2
		}
	}
	return err
}
