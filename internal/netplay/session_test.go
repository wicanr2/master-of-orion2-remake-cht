package netplay

import (
	"net"
	"testing"
	"time"
)

// waitFor 等到 cond 成立(或逾時)。
//
// 讀取在另一條 goroutine,所以送出去之後不能立刻斷言——但也不該 sleep 一個固定秒數
// (慢機器上會偽紅、快機器上白等)。輪詢到成立為止。
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("等 %s 逾時", what)
}

// pollUntil 收滿 n 則訊息(或逾時)。
func pollUntil(t *testing.T, s *Session, n int) []Message {
	t.Helper()
	var got []Message
	waitFor(t, "收到訊息", func() bool {
		got = append(got, s.Poll()...)
		return len(got) >= n
	})
	return got
}

// 兩台機器之間,聊天真的會走過線。
//
// 這是這一項存在的理由:資料模型(chat.go)與畫面(netnextturn.go)先前都做完了,
// 但中間沒有任何東西在跑——打出去的字從來沒離開過本機。
func TestChatCrossesTheWire(t *testing.T) {
	a, b := net.Pipe()
	host := NewSession(0, true, map[int]net.Conn{1: a})
	defer host.Close()
	client := NewSession(1, false, map[int]net.Conn{0: b})
	defer client.Close()

	if err := client.SendChat("在了嗎"); err != nil {
		t.Fatalf("送訊失敗:%v", err)
	}
	got := pollUntil(t, host, 1)
	if got[0].Kind != KindChat {
		t.Errorf("種類應為 %q,得到 %q", KindChat, got[0].Kind)
	}
	if got[0].Text != "在了嗎" {
		t.Errorf("內文應為「在了嗎」,得到 %q", got[0].Text)
	}
	if got[0].Player != 1 {
		t.Errorf("發話者應為 1(以連線為準,不採信封包自報),得到 %d", got[0].Player)
	}
}

// 主機收訊時**以連線為準**,客戶端自報的 Player 蓋掉——否則客戶端可以冒名說話。
func TestHostAttributesBySocketNotByClaim(t *testing.T) {
	a, b := net.Pipe()
	host := NewSession(0, true, map[int]net.Conn{2: a})
	defer host.Close()
	defer b.Close()

	// 直接寫一則自報 Player=0(冒充主機)的訊息。
	go func() { _ = WriteFrame(b, Message{Kind: KindChat, Player: 0, Text: "我是主機"}) }()

	got := pollUntil(t, host, 1)
	if got[0].Player != 2 {
		t.Errorf("應以連線編號 2 為準(冒名的 0 要被蓋掉),得到 %d", got[0].Player)
	}
}

// 主機把 A 說的話轉給 B——星狀拓樸下客戶端之間沒有連線,不轉發就聽不到彼此。
//
// 而且轉發時**要保留原始發話者**:客戶端若把「從主機來的」一律當成主機說的,
// 所有人的話都會顯示成主機說的。
func TestHostRelaysBetweenClientsKeepingTheSpeaker(t *testing.T) {
	a1, c1 := net.Pipe()
	a2, c2 := net.Pipe()
	host := NewSession(0, true, map[int]net.Conn{1: a1, 2: a2})
	defer host.Close()
	alice := NewSession(1, false, map[int]net.Conn{0: c1})
	defer alice.Close()
	bob := NewSession(2, false, map[int]net.Conn{0: c2})
	defer bob.Close()

	if err := alice.SendChat("誰要打安塔蘭"); err != nil {
		t.Fatalf("送訊失敗:%v", err)
	}

	// 主機自己收到。
	if got := pollUntil(t, host, 1); got[0].Player != 1 {
		t.Errorf("主機看到的發話者應為 1,得到 %d", got[0].Player)
	}
	// Bob 透過轉發收到,而且發話者仍是 Alice(1),不是主機(0)。
	got := pollUntil(t, bob, 1)
	if got[0].Text != "誰要打安塔蘭" {
		t.Errorf("Bob 收到的內文不對:%q", got[0].Text)
	}
	if got[0].Player != 1 {
		t.Errorf("轉發後發話者應仍為 1(Alice),得到 %d——被蓋成主機編號的話,所有人的話都會顯示成主機說的", got[0].Player)
	}
}

// 正對照:訊息**不會**被轉回發話者自己。
//
// 少了這條,「轉發給所有連線」的實作也會讓上面那條通過,而 Alice 會看到自己的話出現兩次
// (本機一次 + 繞回來一次)。
func TestRelayDoesNotEchoBackToSender(t *testing.T) {
	a1, c1 := net.Pipe()
	a2, c2 := net.Pipe()
	host := NewSession(0, true, map[int]net.Conn{1: a1, 2: a2})
	defer host.Close()
	alice := NewSession(1, false, map[int]net.Conn{0: c1})
	defer alice.Close()
	bob := NewSession(2, false, map[int]net.Conn{0: c2})
	defer bob.Close()

	if err := alice.SendChat("測試"); err != nil {
		t.Fatalf("送訊失敗:%v", err)
	}
	pollUntil(t, bob, 1) // 等轉發真的跑完,再檢查 Alice 那一端

	if got := alice.Poll(); len(got) != 0 {
		t.Errorf("發話者不該收到自己的話繞回來,得到 %+v", got)
	}
}

// 空訊息不送——原版 `Chat_Box_Input_Loop_` 也是先 `cmp byte_1AAC54, 0 / jz` 擋掉。
func TestSendChatIgnoresEmptyInput(t *testing.T) {
	a, b := net.Pipe()
	host := NewSession(0, true, map[int]net.Conn{1: a})
	defer host.Close()
	client := NewSession(1, false, map[int]net.Conn{0: b})
	defer client.Close()

	for _, empty := range []string{"", "   ", "\n\t"} {
		if err := client.SendChat(empty); err != nil {
			t.Fatalf("空訊息不該回錯誤,得到 %v", err)
		}
	}
	// net.Pipe 是無緩衝的:真的有寫出去,上面的 SendChat 會**卡住**而不是回 nil。
	// 能走到這裡本身就證明沒送。再確認一次收訊端是空的。
	time.Sleep(20 * time.Millisecond)
	if got := host.Poll(); len(got) != 0 {
		t.Errorf("空訊息不該送出去,主機卻收到 %+v", got)
	}
}

// 對面正常離線(EOF)**不記成錯誤**——「他走了」和「這局可能已經不同步了」
// 對玩家的意思完全不同,混在一起就沒辦法據此決定要不要停下來。
func TestPeerLeavingIsNotAnError(t *testing.T) {
	a, b := net.Pipe()
	host := NewSession(0, true, map[int]net.Conn{1: a})
	defer host.Close()

	if host.Peers() != 1 {
		t.Fatalf("開場應有 1 條連線,得到 %d", host.Peers())
	}
	b.Close()

	waitFor(t, "連線被移除", func() bool { return host.Peers() == 0 })
	if err := host.Err(); err != nil {
		t.Errorf("對面正常離線不該記成錯誤,得到 %v", err)
	}
}

// 關掉之後再送訊回明確的錯誤,而不是靜默成功。
func TestSendAfterCloseFails(t *testing.T) {
	a, b := net.Pipe()
	defer b.Close()
	s := NewSession(0, true, map[int]net.Conn{1: a})
	if err := s.Close(); err != nil {
		t.Fatalf("關閉失敗:%v", err)
	}
	if err := s.Close(); err != nil {
		t.Errorf("重複關閉應該是安全的,得到 %v", err)
	}
	if err := s.SendChat("還在嗎"); err != ErrSessionClosed {
		t.Errorf("關閉後送訊應回 ErrSessionClosed,得到 %v", err)
	}
}

// Poll 取完就清空,不會把同一則重複交出去。
func TestPollDrains(t *testing.T) {
	a, b := net.Pipe()
	host := NewSession(0, true, map[int]net.Conn{1: a})
	defer host.Close()
	client := NewSession(1, false, map[int]net.Conn{0: b})
	defer client.Close()

	for _, s := range []string{"一", "二", "三"} {
		if err := client.SendChat(s); err != nil {
			t.Fatalf("送訊失敗:%v", err)
		}
	}
	got := pollUntil(t, host, 3)
	if len(got) != 3 {
		t.Fatalf("應收到 3 則,得到 %d", len(got))
	}
	// 順序就是送出的順序(佇列保證,與封包抵達時間無關)。
	for i, want := range []string{"一", "二", "三"} {
		if got[i].Text != want {
			t.Errorf("第 %d 則應為 %q,得到 %q", i+1, want, got[i].Text)
		}
	}
	if again := host.Poll(); len(again) != 0 {
		t.Errorf("Poll 過的訊息不該再出現一次,得到 %+v", again)
	}
}

// 沒有連線的 Session(單機、或對手都走光了)Send 不回錯誤也不 panic。
//
// 畫面那一端會無條件呼叫 Send,這條守的是「單機開一張網路等待畫面」不會炸。
func TestSendWithNoPeersIsHarmless(t *testing.T) {
	s := NewSession(0, false, nil)
	defer s.Close()
	if err := s.SendChat("有人嗎"); err != nil {
		t.Errorf("沒有對手時送訊不該回錯誤,得到 %v", err)
	}
	if got := s.Poll(); len(got) != 0 {
		t.Errorf("不該憑空收到訊息,得到 %+v", got)
	}
}
