package netplay_test

// lockstep_shell_test.go:把傳輸層與規則層接起來的端到端測試。
//
// 放在**外部測試套件**(`netplay_test`)是刻意的:`internal/netplay` 本身不相依
// `internal/shell`(傳輸層不該知道規則),但「兩層合起來真的能鎖步」這件事必須有人驗。
// 外部測試套件正是放這種測試的地方——它同時 import 兩邊,而不會讓生產程式碼耦合。

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/netplay"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// toPlayerCommands 把線上的 netplay.Command 轉成規則層的 shell.PlayerCommand。
//
// 兩邊的欄位形狀刻意一樣,但**型別分開**:傳輸層不該 import 規則層(見 frame.go 檔頭),
// 所以轉換發生在同時認識兩層的地方——正式對局裡是 cmd/moo2(組裝端),這裡是測試。
func toPlayerCommands(cmds []netplay.Command) []shell.PlayerCommand {
	out := make([]shell.PlayerCommand, 0, len(cmds))
	for _, c := range cmds {
		out = append(out, shell.PlayerCommand{Name: c.Name, Args: c.Args, Text: c.Text})
	}
	return out
}

// TestTwoPeersStayInSyncOverAPipe:兩個對等端各自跑一份 GameSession,
// 每回合互相送出「我這回合下了什麼指令 + 我現在的狀態指紋」,收齊之後**依玩家編號**
// 套用同一批指令再結束回合。每一回合的指紋都必須相同。
//
// 這是鎖步的完整迴圈,只是把 socket 換成 net.Pipe():少了真網路的延遲與分片,
// 但協定邏輯一模一樣(真 socket 的框架讀寫由 TestOverRealTCP 顧)。
func TestTwoPeersStayInSyncOverAPipe(t *testing.T) {
	c0, c1 := net.Pipe()
	defer c0.Close()
	defer c1.Close()

	// 兩邊都從同一顆種子開局——鎖步的前提。
	sessions := [2]*shell.GameSession{shell.NewDemoSession(), shell.NewDemoSession()}
	if sessions[0].StateHash() != sessions[1].StateHash() {
		t.Fatal("測試前提不成立:兩邊開場就不一樣")
	}

	// 每個玩家這一回合要下的指令(決定性:只看回合數)。
	cmdsFor := func(player, turn int) []netplay.Command {
		switch {
		case turn == 2 && player == 0:
			return []netplay.Command{{Name: shell.CmdEnqueueBuild, Args: []int{0, 60}, Text: "住宅"}}
		case turn == 4 && player == 1:
			return []netplay.Command{{Name: shell.CmdSetRelocation, Args: []int{0, 3}}}
		case turn == 6 && player == 0:
			return []netplay.Command{{Name: shell.CmdSendFleet, Args: []int{1}}}
		case turn == 8 && player == 1:
			return []netplay.Command{{Name: shell.CmdCycleTaxRate}}
		case turn == 10 && player == 0:
			return []netplay.Command{{Name: shell.CmdShiftJob, Args: []int{0}, Text: "農夫>工人"}}
		}
		return nil
	}

	const turns = 25
	var wg sync.WaitGroup
	errs := make([]error, 2)
	hashes := [2][]string{}
	var mu sync.Mutex

	peer := func(id int, conn net.Conn) {
		defer wg.Done()
		s := sessions[id]
		for turn := 1; turn <= turns; turn++ {
			mine := netplay.Message{
				Kind: netplay.KindTurnDone, Player: id, Turn: turn,
				StateHash: s.StateHash(), Commands: cmdsFor(id, turn),
			}
			// 兩邊同時寫、同時讀:net.Pipe 是同步的,所以寫必須在自己的 goroutine 裡,
			// 不然兩邊都卡在 Write 上等對方讀。
			werr := make(chan error, 1)
			go func() { werr <- netplay.WriteFrame(conn, mine) }()

			var theirs netplay.Message
			if err := netplay.ReadFrame(conn, &theirs); err != nil && err != io.EOF {
				errs[id] = err
				return
			}
			if err := <-werr; err != nil {
				errs[id] = err
				return
			}

			tb := netplay.NewTable(2, turn)
			if err := tb.Add(mine); err != nil {
				errs[id] = err
				return
			}
			if err := tb.Add(theirs); err != nil {
				errs[id] = err
				return
			}
			if !tb.Ready() {
				t.Errorf("玩家 %d 第 %d 回合沒收齊", id, turn)
				return
			}
			if d := tb.Desync(); d != "" {
				errs[id] = errDesync(d)
				return
			}
			if err := s.ApplyPlayerCommands(toPlayerCommands(tb.Commands())); err != nil {
				errs[id] = err // 不認得的指令 = 兩邊版本不同,停下來
				return
			}
			s.EndTurn()

			mu.Lock()
			hashes[id] = append(hashes[id], s.StateHash())
			mu.Unlock()
		}
	}

	wg.Add(2)
	go peer(0, c0)
	go peer(1, c1)
	wg.Wait()

	for id, err := range errs {
		if err != nil {
			t.Fatalf("玩家 %d 出錯:%v", id, err)
		}
	}
	if len(hashes[0]) != turns || len(hashes[1]) != turns {
		t.Fatalf("回合數不對:%d / %d", len(hashes[0]), len(hashes[1]))
	}
	for i := range hashes[0] {
		if hashes[0][i] != hashes[1][i] {
			t.Fatalf("第 %d 回合分岔:%s vs %s", i+1, hashes[0][i][:12], hashes[1][i][:12])
		}
	}
	// 正對照:指令真的有被套用,否則這支測試等於只驗了「兩個什麼都不做的 session 一樣」。
	if len(sessions[0].BuildQueueFor(0)) == 0 && sessions[0].ColonyRelocation(0) == shell.ColonyRelocationNone {
		t.Error("測試前提不成立:整場下來沒有任何指令產生效果")
	}
}

type errDesync string

func (e errDesync) Error() string { return string(e) }

// TestUnknownCommandOverTheWireIsRejected:對面送來一條這邊不認得的指令時,
// 必須**停下來**而不是跳過。
//
// 跳過在鎖步裡是最糟的處理:一邊套用了、另一邊沒有,而且沒有人會知道——
// 幾十回合之後才以「你的畫面跟我不一樣」爆出來。
func TestUnknownCommandOverTheWireIsRejected(t *testing.T) {
	s := shell.NewDemoSession()
	err := s.ApplyPlayerCommands(toPlayerCommands([]netplay.Command{
		{Name: shell.CmdCycleTaxRate},
		{Name: "未來版本才有的指令"},
	}))
	if err == nil {
		t.Fatal("不認得的指令應該讓整批停下來")
	}
}

// 傳輸層的指令名就是規則層那份清單——兩邊漂掉的話,網路對戰會出現
// 「這個操作在單機做得到、連線就不會同步過去」的靜默缺口。
func TestEveryPlayerCommandCanTravel(t *testing.T) {
	s := shell.NewDemoSession()
	for _, name := range shell.PlayerCommandNames() {
		var buf bytes.Buffer
		if err := netplay.WriteFrame(&buf, netplay.Message{
			Kind: netplay.KindTurnDone, Player: 0, Turn: 1,
			Commands: []netplay.Command{{Name: name}},
		}); err != nil {
			t.Fatalf("%q 送不出去:%v", name, err)
		}
		var m netplay.Message
		if err := netplay.ReadFrame(&buf, &m); err != nil {
			t.Fatalf("%q 讀不回來:%v", name, err)
		}
		if err := s.ApplyPlayerCommands(toPlayerCommands(m.Commands)); err != nil {
			t.Errorf("%q 過了線之後規則層不認得:%v", name, err)
		}
	}
}
