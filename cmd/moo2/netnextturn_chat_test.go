package main

import (
	"net"
	"testing"
	"time"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/netplay"
)

// 這一組測的是「畫面 ↔ 幫浦」的接線,不是幫浦本身(那在 internal/netplay/session_test.go)。
//
// 螢幕物件刻意用零值 sceneBuilder 之外的方式建:`pumpChat`/`sendChat` 都不碰 b,
// 所以不需要真的 LBX 資產就測得動(`b.netNextTurn()` 需要資產,測試環境沒有)。

// 打完 Enter 立刻看到自己那一行,而且那一行**真的送出去了**。
//
// 這是這一項的核心:先前 sendChat 只加進本機記錄,字從來沒離開過這台機器。
func TestSendChatWritesToTheWireAndShowsLocally(t *testing.T) {
	a, b := net.Pipe()
	me := netplay.NewSession(1, false, map[int]net.Conn{0: a})
	defer me.Close()
	peer := netplay.NewSession(0, true, map[int]net.Conn{1: b})
	defer peer.Close()

	s := &netNextTurnScreen{me: 1}
	s.attach(me)
	s.typing = "我還在算研究"
	s.sendChat()

	if s.typing != "" {
		t.Errorf("送出後輸入行應清空,得到 %q", s.typing)
	}
	lines := s.chat.Lines()
	if len(lines) != 1 || lines[0].Text != "我還在算研究" || lines[0].Speaker != 1 {
		t.Fatalf("本機記錄應立刻多一行(不等封包繞回來),得到 %+v", lines)
	}

	var got []netplay.Message
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && len(got) == 0 {
		got = append(got, peer.Poll()...)
		time.Sleep(time.Millisecond)
	}
	if len(got) != 1 {
		t.Fatalf("對面應收到 1 則,得到 %d 則", len(got))
	}
	if got[0].Text != "我還在算研究" {
		t.Errorf("對面收到的內文不對:%q", got[0].Text)
	}
}

// 對面說的話會在下一幀進到記錄裡。
func TestPumpChatDrainsIncomingIntoTheLog(t *testing.T) {
	a, b := net.Pipe()
	me := netplay.NewSession(0, true, map[int]net.Conn{1: a})
	defer me.Close()
	peer := netplay.NewSession(1, false, map[int]net.Conn{0: b})
	defer peer.Close()

	s := &netNextTurnScreen{me: 0, names: []string{"我", "對手"}}
	s.attach(me)
	if err := peer.SendChat("換你了"); err != nil {
		t.Fatalf("送訊失敗:%v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && s.chat.Len() == 0 {
		s.pumpChat()
		time.Sleep(time.Millisecond)
	}
	lines := s.chat.Lines()
	if len(lines) != 1 {
		t.Fatalf("應收進 1 行,得到 %d 行", len(lines))
	}
	if lines[0].Text != "換你了" || lines[0].Speaker != 1 {
		t.Errorf("收進來的那一行不對:%+v", lines[0])
	}
}

// 非聊天的訊息不進聊天記錄——回合表的推進在別處,在這裡順手動它會讓同一件事有兩個入口。
func TestPumpChatIgnoresNonChatMessages(t *testing.T) {
	a, b := net.Pipe()
	me := netplay.NewSession(0, true, map[int]net.Conn{1: a})
	defer me.Close()
	peer := netplay.NewSession(1, false, map[int]net.Conn{0: b})
	defer peer.Close()

	s := &netNextTurnScreen{me: 0}
	s.attach(me)
	if err := peer.Send(netplay.Message{Kind: netplay.KindTurnDone, Turn: 3}); err != nil {
		t.Fatalf("送訊失敗:%v", err)
	}
	// 再送一則聊天當「柵欄」:它收到了就代表前面那則也已經流過 pumpChat。
	if err := peer.SendChat("柵欄"); err != nil {
		t.Fatalf("送訊失敗:%v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && s.chat.Len() == 0 {
		s.pumpChat()
		time.Sleep(time.Millisecond)
	}
	lines := s.chat.Lines()
	if len(lines) != 1 || lines[0].Text != "柵欄" {
		t.Errorf("只有聊天該進記錄,得到 %+v", lines)
	}
}

// 沒有幫浦(單機、截圖廊)時一切照舊:打字、送出、記錄都能動,只是話不出去。
//
// 截圖廊那張 30_netwait.png 走的就是這條路徑——這條測試守的是它不會因為接了幫浦而爆掉。
func TestChatStillWorksWithoutASession(t *testing.T) {
	s := &netNextTurnScreen{me: 0}
	s.pumpChat() // 不該 panic
	s.typing = "自己跟自己說話"
	s.sendChat()
	if s.chat.Len() != 1 {
		t.Errorf("沒有連線時本機記錄仍該多一行,得到 %d 行", s.chat.Len())
	}
}

// 線斷了不打斷輸入:字已經在自己的記錄裡,斷線由 Session.Err 那條線回報。
func TestSendChatSurvivesABrokenConnection(t *testing.T) {
	a, b := net.Pipe()
	me := netplay.NewSession(1, false, map[int]net.Conn{0: a})
	b.Close()
	defer me.Close()

	s := &netNextTurnScreen{me: 1}
	s.attach(me)
	s.typing = "還在嗎"
	s.sendChat() // 不該 panic,也不該把字吞掉

	if s.chat.Len() != 1 {
		t.Errorf("送不出去時本機記錄仍該保留那一行,得到 %d 行", s.chat.Len())
	}
}
