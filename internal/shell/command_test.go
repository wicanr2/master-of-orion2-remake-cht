package shell

import (
	"encoding/json"
	"sort"
	"strings"
	"testing"
)

// command_test.go:玩家指令層的護欄。

// 不認得的指令名必須**報錯**,不能靜默忽略——鎖步裡「跳過一條」等於兩邊從此不同步,
// 而且不會有人發現。
func TestUnknownCommandIsAnError(t *testing.T) {
	s := NewDemoSession()
	err := s.ApplyPlayerCommand(PlayerCommand{Name: "沒有這條指令"})
	if err == nil {
		t.Fatal("不認得的指令應該報錯")
	}
	if !strings.Contains(err.Error(), "沒有這條指令") {
		t.Errorf("錯誤訊息應該講出是哪一條,實得:%v", err)
	}
	// 一批指令中途出錯要停下來並指出是第幾條。
	err = s.ApplyPlayerCommands([]PlayerCommand{
		{Name: CmdCycleTaxRate}, {Name: "壞掉的"}, {Name: CmdCycleTaxRate},
	})
	if err == nil || !strings.Contains(err.Error(), "第 1 條") {
		t.Errorf("應指出是第 1 條出錯,實得:%v", err)
	}
}

// PlayerCommandNames 必須與 ApplyPlayerCommand 的 switch 一致:
// 表上有的都認得,認得的都在表上。這是「網路對戰支援到哪」的唯一答案,不能漂。
func TestEveryListedCommandIsHandled(t *testing.T) {
	names := PlayerCommandNames()
	if !sort.StringsAreSorted(names) {
		t.Error("指令清單應該是排序過的(比對與顯示都靠它)")
	}
	seen := map[string]bool{}
	for _, n := range names {
		if seen[n] {
			t.Errorf("指令 %q 在清單裡重複了", n)
		}
		seen[n] = true
		s := NewDemoSession()
		if err := s.ApplyPlayerCommand(PlayerCommand{Name: n}); err != nil {
			t.Errorf("清單上的 %q 卻沒被 ApplyPlayerCommand 處理:%v", n, err)
		}
	}
}

// 指令要能過 JSON 往返——它會走在網路上。
func TestCommandSurvivesJSON(t *testing.T) {
	want := PlayerCommand{Name: CmdEnqueueBuild, Args: []int{0, 60}, Text: "住宅"}
	b, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got PlayerCommand
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Name != want.Name || got.Text != want.Text || len(got.Args) != 2 || got.Args[1] != 60 {
		t.Errorf("往返後對不上:%+v", got)
	}
}

// 走指令層與直接呼叫方法必須得到**一模一樣的狀態**——指令層只是轉接,不該有自己的規則。
func TestCommandPathMatchesDirectCall(t *testing.T) {
	cases := []struct {
		name   string
		cmd    PlayerCommand
		direct func(*GameSession)
	}{
		{"派艦隊", PlayerCommand{Name: CmdSendFleet, Args: []int{3}},
			func(s *GameSession) { s.SendFleet(3) }},
		{"排建造", PlayerCommand{Name: CmdEnqueueBuild, Args: []int{0, 60}, Text: "住宅"},
			func(s *GameSession) { s.EnqueueBuild(0, "住宅", 60) }},
		{"設集結點", PlayerCommand{Name: CmdSetRelocation, Args: []int{0, 4}},
			func(s *GameSession) { s.SetColonyRelocation(0, 4) }},
		{"稅率", PlayerCommand{Name: CmdCycleTaxRate},
			func(s *GameSession) { s.CycleTaxRate() }},
		{"職務轉移", PlayerCommand{Name: CmdShiftJob, Args: []int{0}, Text: "農夫>工人"},
			func(s *GameSession) { s.ShiftColonyJob(0, "農夫", "工人") }},
		{"拓殖", PlayerCommand{Name: CmdColonizeStar, Args: []int{0}},
			func(s *GameSession) { s.ColonizeStar(0) }},
		{"結束回合", PlayerCommand{Name: CmdEndTurn},
			func(s *GameSession) { s.EndTurn() }},
	}
	for _, c := range cases {
		viaCmd, viaDirect := NewDemoSession(), NewDemoSession()
		if err := viaCmd.ApplyPlayerCommand(c.cmd); err != nil {
			t.Fatalf("%s:%v", c.name, err)
		}
		c.direct(viaDirect)
		if a, b := viaCmd.StateHash(), viaDirect.StateHash(); a != b {
			t.Errorf("%s:走指令層與直接呼叫的結果不同(%s vs %s)", c.name, a[:12], b[:12])
		}
	}
}

// 職務轉移的字串參數要拆得對(分隔符 '>' 不會出現在職務名裡)。
func TestShiftJobArgs(t *testing.T) {
	for _, c := range []struct{ in, from, to string }{
		{"農夫>工人", "農夫", "工人"},
		{"科學家>農夫", "科學家", "農夫"},
		{"沒有分隔符", "沒有分隔符", ""},
	} {
		from, to := shiftJobArgs(c.in)
		if from != c.from || to != c.to {
			t.Errorf("%q 應拆成 (%q,%q),實得 (%q,%q)", c.in, c.from, c.to, from, to)
		}
	}
}

// 參數不足時走預設值、不 panic——送出端有 bug 不該讓整局中止(見 arg 的註解)。
func TestMissingArgsDoNotPanic(t *testing.T) {
	s := NewDemoSession()
	for _, n := range PlayerCommandNames() {
		if err := s.ApplyPlayerCommand(PlayerCommand{Name: n}); err != nil {
			t.Errorf("%q 少了參數卻報錯:%v", n, err)
		}
	}
}

// 同一批指令套兩次到兩份同種子的 session,結果必須一致——這是鎖步的核心前提,
// 也是 internal/netplay 端到端測試依賴的性質。
func TestSameCommandsSameState(t *testing.T) {
	cmds := []PlayerCommand{
		{Name: CmdCycleTaxRate},
		{Name: CmdEnqueueBuild, Args: []int{0, 60}, Text: "住宅"},
		{Name: CmdSendFleet, Args: []int{2}},
		{Name: CmdEndTurn},
		{Name: CmdSetRelocation, Args: []int{0, 5}},
		{Name: CmdEndTurn},
	}
	a, b := NewDemoSession(), NewDemoSession()
	if err := a.ApplyPlayerCommands(cmds); err != nil {
		t.Fatal(err)
	}
	if err := b.ApplyPlayerCommands(cmds); err != nil {
		t.Fatal(err)
	}
	if a.StateHash() != b.StateHash() {
		t.Error("同一批指令跑出不同狀態")
	}
	// 正對照:指令真的有效果,否則這支測試等於只驗了「兩個空 session 一樣」。
	if a.StateHash() == NewDemoSession().StateHash() {
		t.Error("測試前提不成立:整批指令對狀態毫無影響")
	}
}
