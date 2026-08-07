package netplay

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"strings"
	"sync"
	"testing"
)

// netplay_test.go:傳輸層與鎖步的護欄。全部不需要真的開一局遊戲,也不需要真的網路
// (`net.Pipe()` 就夠)——只有一支刻意走真的 TCP,驗「換成真 socket 也一樣」。

func TestFrameRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	want := Message{Kind: KindTurnDone, Player: 2, Turn: 7, StateHash: "deadbeef",
		Commands: []Command{{Name: "send_fleet", Args: []int{0, 5}}, {Name: "build", Text: "住宅"}}}
	if err := WriteFrame(&buf, want); err != nil {
		t.Fatalf("寫入失敗:%v", err)
	}
	var got Message
	if err := ReadFrame(&buf, &got); err != nil {
		t.Fatalf("讀取失敗:%v", err)
	}
	if got.Kind != want.Kind || got.Player != want.Player || got.Turn != want.Turn ||
		got.StateHash != want.StateHash || len(got.Commands) != 2 ||
		got.Commands[0].Name != "send_fleet" || got.Commands[0].Args[1] != 5 ||
		got.Commands[1].Text != "住宅" {
		t.Errorf("往返之後對不上:\n want %+v\n got  %+v", want, got)
	}
}

// 兩則訊息黏在同一個位元組流裡也要能各自讀出來——這正是長度前綴存在的理由。
func TestTwoFramesInOneStream(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteFrame(&buf, Message{Kind: KindHello, Player: 0, Name: "甲"}); err != nil {
		t.Fatal(err)
	}
	if err := WriteFrame(&buf, Message{Kind: KindHello, Player: 1, Name: "乙"}); err != nil {
		t.Fatal(err)
	}
	for i, want := range []string{"甲", "乙"} {
		var m Message
		if err := ReadFrame(&buf, &m); err != nil {
			t.Fatalf("第 %d 則讀取失敗:%v", i, err)
		}
		if m.Name != want {
			t.Errorf("第 %d 則應是 %q,實得 %q", i, want, m.Name)
		}
	}
}

// 讀完就是 EOF,而且要原樣回傳 io.EOF——呼叫端靠它分辨「對面關了」與「壞掉了」。
func TestReadFrameReturnsEOFWhenStreamEnds(t *testing.T) {
	var buf bytes.Buffer
	var m Message
	if err := ReadFrame(&buf, &m); err != io.EOF {
		t.Errorf("空的流應回 io.EOF,實得 %v", err)
	}
}

// 壞掉(或惡意)的長度前綴不該讓我們去配置一塊天文數字的記憶體。
func TestOversizedLengthPrefixIsRejected(t *testing.T) {
	var buf bytes.Buffer
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], MaxFrameBytes+1)
	buf.Write(hdr[:])
	var m Message
	if err := ReadFrame(&buf, &m); err != ErrFrameTooLarge {
		t.Errorf("超長前綴應被擋下,實得 %v", err)
	}
}

// 走真的 TCP:兩邊各送一則,各收一則。
func TestOverRealTCP(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skipf("這個環境開不了 socket:%v", err)
	}
	defer ln.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	var srvErr error
	go func() {
		defer wg.Done()
		c, err := ln.Accept()
		if err != nil {
			srvErr = err
			return
		}
		defer c.Close()
		var m Message
		if err := ReadFrame(c, &m); err != nil {
			srvErr = err
			return
		}
		srvErr = WriteFrame(c, Message{Kind: KindTurnDone, Player: 1, Turn: m.Turn, StateHash: m.StateHash})
	}()

	c, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatalf("連不上:%v", err)
	}
	defer c.Close()
	if err := WriteFrame(c, Message{Kind: KindTurnDone, Player: 0, Turn: 12, StateHash: "abc123"}); err != nil {
		t.Fatalf("送出失敗:%v", err)
	}
	var reply Message
	if err := ReadFrame(c, &reply); err != nil {
		t.Fatalf("收回覆失敗:%v", err)
	}
	wg.Wait()
	if srvErr != nil {
		t.Fatalf("伺服端出錯:%v", srvErr)
	}
	if reply.Turn != 12 || reply.StateHash != "abc123" || reply.Player != 1 {
		t.Errorf("回覆內容不對:%+v", reply)
	}
}

// --- 鎖步回合表 ---

func TestTableReadyOnlyWhenEveryoneArrives(t *testing.T) {
	tb := NewTable(3, 5)
	if tb.Ready() {
		t.Error("一個都還沒到就說到齊了")
	}
	for p := 0; p < 2; p++ {
		if err := tb.Add(Message{Kind: KindTurnDone, Player: p, Turn: 5}); err != nil {
			t.Fatal(err)
		}
	}
	if tb.Ready() {
		t.Error("三人局只到兩個不該算到齊")
	}
	if miss := tb.Missing(); len(miss) != 1 || miss[0] != 2 {
		t.Errorf("還沒到的應是 [2],實得 %v", miss)
	}
	if err := tb.Add(Message{Kind: KindTurnDone, Player: 2, Turn: 5}); err != nil {
		t.Fatal(err)
	}
	if !tb.Ready() {
		t.Error("全員到齊卻沒說到齊")
	}
}

// 指令要**依玩家編號**排序,不是依到達順序——否則「誰的封包先到」會影響結果,
// 那正是鎖步最典型的分岔來源。
func TestCommandsAreOrderedByPlayerNotArrival(t *testing.T) {
	tb := NewTable(3, 1)
	// 故意倒著送。
	tb.Add(Message{Kind: KindTurnDone, Player: 2, Turn: 1, Commands: []Command{{Name: "c"}}})
	tb.Add(Message{Kind: KindTurnDone, Player: 0, Turn: 1, Commands: []Command{{Name: "a"}}})
	tb.Add(Message{Kind: KindTurnDone, Player: 1, Turn: 1, Commands: []Command{{Name: "b"}}})
	got := tb.Commands()
	want := []string{"a", "b", "c"}
	if len(got) != 3 {
		t.Fatalf("應有 3 條指令,實得 %d", len(got))
	}
	for i := range want {
		if got[i].Name != want[i] {
			t.Errorf("第 %d 條應是 %q,實得 %q(指令沒有依玩家編號排序)", i, want[i], got[i].Name)
		}
	}
}

func TestTableRejectsWrongTurnAndDuplicates(t *testing.T) {
	tb := NewTable(2, 5)
	if err := tb.Add(Message{Kind: KindHello, Player: 0}); err == nil {
		t.Error("回合表不該收 hello")
	}
	if err := tb.Add(Message{Kind: KindTurnDone, Player: 0, Turn: 6}); err == nil {
		t.Error("回合對不上應該報錯(有一邊漏了東西)")
	}
	if err := tb.Add(Message{Kind: KindTurnDone, Player: 0, Turn: 5}); err != nil {
		t.Fatal(err)
	}
	if err := tb.Add(Message{Kind: KindTurnDone, Player: 0, Turn: 5}); err == nil {
		t.Error("同一個玩家送兩次應該報錯")
	}
}

// 指紋不一致要抓得出來,而且要指出是第幾回合、哪兩個玩家。
func TestDesyncIsDetectedWithTurnAndPlayers(t *testing.T) {
	tb := NewTable(2, 9)
	tb.Add(Message{Kind: KindTurnDone, Player: 0, Turn: 9, StateHash: "aaaaaaaa1111"})
	if tb.Desync() != "" {
		t.Error("還沒到齊就不該下分岔結論")
	}
	tb.Add(Message{Kind: KindTurnDone, Player: 1, Turn: 9, StateHash: "bbbbbbbb2222"})
	msg := tb.Desync()
	if msg == "" {
		t.Fatal("兩邊指紋不同卻沒抓出來")
	}
	if !strings.Contains(msg, "9") || !strings.Contains(msg, "aaaaaaaa") || !strings.Contains(msg, "bbbbbbbb") {
		t.Errorf("分岔訊息要講清楚回合與雙方指紋,實得:%s", msg)
	}
}

func TestNoDesyncWhenHashesMatch(t *testing.T) {
	tb := NewTable(2, 9)
	tb.Add(Message{Kind: KindTurnDone, Player: 0, Turn: 9, StateHash: "same"})
	tb.Add(Message{Kind: KindTurnDone, Player: 1, Turn: 9, StateHash: "same"})
	if msg := tb.Desync(); msg != "" {
		t.Errorf("指紋一樣卻報了分岔:%s", msg)
	}
}

// 零值 Table 不該假裝到齊——沒設定人數就說 Ready 會讓對局在只有一個人的情況下推進。
func TestZeroTableIsNeverReady(t *testing.T) {
	var tb Table
	if tb.Ready() {
		t.Error("零值回合表不該說到齊")
	}
	if err := tb.Add(Message{Kind: KindTurnDone, Player: 0, Turn: 0}); err != nil {
		t.Fatalf("零值表應該收得下第 0 回合:%v", err)
	}
	if tb.Ready() {
		t.Error("players 為 0 時永遠不該到齊")
	}
}
