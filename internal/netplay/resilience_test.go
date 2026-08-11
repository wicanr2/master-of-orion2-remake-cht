package netplay

import (
	"errors"
	"net"
	"sync"
	"testing"
	"time"
)

func TestAuthenticatedTLSSessionCanResumeSamePlayer(t *testing.T) {
	opts := LobbyOptions{AuthToken: "抽樣測試密碼", EnableTLS: true}
	lb, err := HostWithOptions("127.0.0.1:0", "主機", 77, opts)
	if err != nil {
		t.Skipf("環境不能開 TCP/TLS：%v", err)
	}
	defer lb.Close()
	lb.SetMaxPlayers(2)

	wrongDone := make(chan error, 1)
	go func() {
		_, _, _, _, err := JoinWithOptions(lb.Addr(), "錯誤密碼", 2*time.Second,
			JoinOptions{LobbyOptions: LobbyOptions{AuthToken: "錯誤", EnableTLS: true}})
		wrongDone <- err
	}()
	if _, err := lb.AcceptOne(2 * time.Second); err == nil {
		t.Fatal("錯誤密碼不應通過大廳驗證")
	}
	if err := <-wrongDone; err == nil {
		t.Fatal("客戶端應看到身份驗證失敗")
	}

	joined := make(chan struct {
		c      net.Conn
		id     int
		roster Roster
		token  string
		err    error
	}, 1)
	go func() {
		c, id, roster, token, err := JoinWithOptions(lb.Addr(), "客戶端", 2*time.Second,
			JoinOptions{LobbyOptions: opts})
		joined <- struct {
			c      net.Conn
			id     int
			roster Roster
			token  string
			err    error
		}{c, id, roster, token, err}
	}()
	p, err := lb.AcceptOne(2 * time.Second)
	if err != nil {
		t.Fatalf("正確密碼加入失敗：%v", err)
	}
	result := <-joined
	if result.err != nil {
		t.Fatalf("客戶端握手失敗：%v", result.err)
	}
	if p.ID != 1 || result.id != 1 || result.token == "" || result.roster.Seed != 77 {
		t.Fatalf("第一次身份／名冊不對：p=%+v result=%+v", p, result)
	}
	_ = result.c.Close()

	lb.SetReconnectOnly()
	rejoined := make(chan struct {
		id    int
		token string
		err   error
	}, 1)
	go func() {
		c, id, _, token, err := JoinWithOptions(lb.Addr(), "不應改名", 2*time.Second,
			JoinOptions{LobbyOptions: opts, ResumeToken: result.token})
		if c != nil {
			_ = c.Close()
		}
		rejoined <- struct {
			id    int
			token string
			err   error
		}{id, token, err}
	}()
	p2, err := lb.AcceptOne(2 * time.Second)
	if err != nil {
		t.Fatalf("resume 加入失敗：%v", err)
	}
	r := <-rejoined
	if p2.ID != 1 || r.id != 1 || r.token != result.token || r.err != nil {
		t.Fatalf("resume 沒保留身份：p=%+v result=%+v", p2, r)
	}
}

func TestSessionHeartbeatTimeoutAfterReconnectGrace(t *testing.T) {
	a, b := net.Pipe()
	defer b.Close()
	s := NewSessionWithOptions(0, false, map[int]net.Conn{1: a}, SessionOptions{
		HeartbeatInterval: 10 * time.Millisecond,
		PeerTimeout:       25 * time.Millisecond,
		ReconnectGrace:    45 * time.Millisecond,
		ReconnectInterval: 5 * time.Millisecond,
		WriteTimeout:      10 * time.Millisecond,
	})
	defer s.Close()
	waitFor(t, "心跳逾時錯誤", func() bool { return errors.Is(s.Err(), ErrPeerTimeout) })
	if s.Peers() != 0 {
		t.Fatalf("逾時後 peer 應被移出，得到 %d", s.Peers())
	}
}

func TestSessionReconnectCallbackReplacesPeer(t *testing.T) {
	a, oldRemote := net.Pipe()
	defer oldRemote.Close()
	newRemote := make(chan net.Conn, 1)
	var once sync.Once
	s := NewSessionWithOptions(0, false, map[int]net.Conn{1: a}, SessionOptions{
		HeartbeatInterval: 10 * time.Millisecond,
		PeerTimeout:       100 * time.Millisecond,
		ReconnectGrace:    250 * time.Millisecond,
		ReconnectInterval: 10 * time.Millisecond,
		WriteTimeout:      25 * time.Millisecond,
		Reconnect: func(peerID int) (net.Conn, error) {
			if peerID != 1 {
				t.Fatalf("callback 收到錯誤 peer id %d", peerID)
			}
			var c, remote net.Conn
			once.Do(func() {
				c, remote = net.Pipe()
				newRemote <- remote
			})
			if c == nil {
				return nil, errors.New("只允許一次抽樣重連")
			}
			return c, nil
		},
	})
	defer s.Close()
	_ = oldRemote.Close()

	var remote net.Conn
	select {
	case remote = <-newRemote:
	case <-time.After(2 * time.Second):
		t.Fatal("重連 callback 沒有被呼叫")
	}
	defer remote.Close()
	go func() {
		for {
			var m Message
			if err := ReadFrame(remote, &m); err != nil {
				return
			}
			if m.Kind == KindPing {
				_ = WriteFrame(remote, Message{Kind: KindPong})
			}
		}
	}()

	waitFor(t, "新 peer 建立", func() bool { return s.Peers() == 1 })
	go func() { _ = WriteFrame(remote, Message{Kind: KindChat, Player: 1, Text: "重連成功"}) }()
	got := pollUntil(t, s, 1)
	if got[0].Kind != KindChat || got[0].Text != "重連成功" {
		t.Fatalf("重連後訊息不對：%+v", got[0])
	}
}
