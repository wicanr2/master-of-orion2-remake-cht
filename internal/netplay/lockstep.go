package netplay

import (
	"fmt"
	"sort"
)

// lockstep.go:鎖步回合表——「全部到齊才推進」的那一半。
//
// 原版 `Net_Next_Turn_` 的骨架是逐玩家的「這回合結束了嗎」旗標陣列 + 等待迴圈
// (見 frame.go 檔頭的反組譯)。這一檔是同一件事的資料結構版本,而且多做一件原版沒有的事:
// **逐回合比對狀態指紋**。
//
// 為什麼要比指紋:鎖步的前提是「同樣的指令序列在每台機器上算出同樣的結果」。
// 那個前提一旦破了(浮點差異、map 迭代順序、沒被存進快照的欄位…),兩邊會**安靜地**
// 漂開,幾十回合之後才以「你的畫面跟我不一樣」的形式爆出來——那時候已經回推不了是哪一步歪的。
// 每回合比一次,分岔的那一回合就是問題發生的那一回合。
//
// (規則層的決定性本身由 `internal/shell/determinism_test.go` 那組閘門守著;
//  這裡守的是「跑起來之後真的還一致」。)

// Table 收集一個回合裡各玩家的 TurnDone。
//
// 零值可用(players 為 0 時 Ready 恆為 false——沒設定人數就不會誤以為到齊了)。
type Table struct {
	players int
	turn    int
	got     map[int]Message
	ready   map[int]Message
}

// NewTable 開一張回合表:players 是這一局的人類玩家數,turn 是目前回合。
func NewTable(players, turn int) *Table {
	return &Table{players: players, turn: turn, got: map[int]Message{}, ready: map[int]Message{}}
}

// Turn 回傳目前回合。
func (t *Table) Turn() int { return t.turn }

// Add 收下一則 TurnDone。
//
// 回傳錯誤的三種情況都值得停下來看,而不是默默忽略:
//   - 不是 TurnDone
//   - 回合對不上(對面比我快一回合 → 一定有一邊漏了東西)
//   - 同一個玩家送了兩次(重送或身分被冒用)
func (t *Table) Add(m Message) error {
	if m.Kind != KindTurnDone {
		return fmt.Errorf("netplay: 回合表只收 turn_done,收到 %q", m.Kind)
	}
	if m.Turn != t.turn {
		return fmt.Errorf("netplay: 回合對不上(表在第 %d 回合,封包是第 %d 回合)", t.turn, m.Turn)
	}
	if t.players > 0 && (m.Player < 0 || m.Player >= t.players) {
		return fmt.Errorf("netplay: turn_done 的玩家 %d 不在名冊內", m.Player)
	}
	if t.got == nil {
		t.got = map[int]Message{}
	}
	if _, dup := t.got[m.Player]; dup {
		return fmt.Errorf("netplay: 玩家 %d 這一回合送了兩次", m.Player)
	}
	t.got[m.Player] = m
	return nil
}

// CommandsFor 回傳指定玩家這一回合的指令副本。鎖步重播必須逐玩家切換
// 作用中席位，不能只拿一個攤平後的清單，否則指令會套到錯的帝國。
func (t *Table) CommandsFor(player int) []Command {
	m, ok := t.got[player]
	if !ok {
		return nil
	}
	return append([]Command(nil), m.Commands...)
}

// AddReady 收下一則「全員指令已在本機重播完成」的狀態指紋。
// 這是 TurnDone 之後的第二階段；只有所有玩家都算出同一個 hash 才能推進世界回合。
func (t *Table) AddReady(m Message) error {
	if m.Kind != KindTurnReady {
		return fmt.Errorf("netplay: ready 表只收 turn_ready,收到 %q", m.Kind)
	}
	if m.Turn != t.turn {
		return fmt.Errorf("netplay: ready 回合對不上(表在第 %d 回合,封包是第 %d 回合)", t.turn, m.Turn)
	}
	if m.Player < 0 || m.Player >= t.players {
		return fmt.Errorf("netplay: turn_ready 的玩家 %d 不在名冊內", m.Player)
	}
	if t.ready == nil {
		t.ready = map[int]Message{}
	}
	if _, dup := t.ready[m.Player]; dup {
		return fmt.Errorf("netplay: 玩家 %d 這一回合送了兩次 turn_ready", m.Player)
	}
	t.ready[m.Player] = m
	return nil
}

// ReadyComplete 回傳所有玩家是否都完成第二階段重播。
func (t *Table) ReadyComplete() bool { return t.players > 0 && len(t.ready) == t.players }

// ReadyDesync 比對第二階段重播後的狀態指紋。回傳空字串代表一致或尚未到齊。
func (t *Table) ReadyDesync() string {
	if !t.ReadyComplete() {
		return ""
	}
	players := make([]int, 0, len(t.ready))
	for p := range t.ready {
		players = append(players, p)
	}
	sort.Ints(players)
	ref, refPlayer := "", -1
	for _, p := range players {
		h := t.ready[p].StateHash
		if h == "" {
			return fmt.Sprintf("第 %d 回合玩家 %d 沒有回報重播後狀態指紋", t.turn, p)
		}
		if ref == "" {
			ref, refPlayer = h, p
			continue
		}
		if h != ref {
			return fmt.Sprintf("第 %d 回合重播後狀態分岔:玩家 %d 是 %s,玩家 %d 是 %s",
				t.turn, refPlayer, short(ref), p, short(h))
		}
	}
	return ""
}

// Ready 回傳是不是全員到齊。
func (t *Table) Ready() bool { return t.players > 0 && len(t.got) == t.players }

// Missing 回傳還沒送到的玩家編號(升冪)——UI 的「等待 X 玩家」就靠它。
func (t *Table) Missing() []int {
	var out []int
	for p := 0; p < t.players; p++ {
		if _, ok := t.got[p]; !ok {
			out = append(out, p)
		}
	}
	return out
}

// Commands 回傳這一回合要套用的所有指令,**依玩家編號排序**。
//
// ⚠ 排序不是為了好看:鎖步要求每台機器**以同樣的順序**套用同樣的指令。
// 依到達順序套用會讓「誰的封包先到」影響結果,那正是 lock-step 最典型的分岔來源。
func (t *Table) Commands() []Command {
	players := make([]int, 0, len(t.got))
	for p := range t.got {
		players = append(players, p)
	}
	sort.Ints(players)
	var out []Command
	for _, p := range players {
		out = append(out, t.got[p].Commands...)
	}
	return out
}

// Desync 檢查各方回報的狀態指紋是否一致。
//
// 回傳空字串 = 一致(或還沒到齊、或沒有人帶指紋)。不一致時回傳一段人看得懂的說明,
// 呼叫端應該**停止對局**並顯示它——鎖步一旦分岔,繼續玩只會讓兩邊差得更遠。
func (t *Table) Desync() string {
	if !t.Ready() {
		return ""
	}
	players := make([]int, 0, len(t.got))
	for p := range t.got {
		players = append(players, p)
	}
	sort.Ints(players)

	ref, refPlayer := "", -1
	for _, p := range players {
		h := t.got[p].StateHash
		if h == "" {
			continue // 沒帶指紋的不參與比對(舊版本或測試),但也不當成「一致」
		}
		if ref == "" {
			ref, refPlayer = h, p
			continue
		}
		if h != ref {
			return fmt.Sprintf("第 %d 回合狀態分岔:玩家 %d 是 %s,玩家 %d 是 %s",
				t.turn, refPlayer, short(ref), p, short(h))
		}
	}
	return ""
}

// short 取指紋前 8 個字元(給人看的)。
func short(h string) string {
	if len(h) <= 8 {
		return h
	}
	return h[:8]
}
