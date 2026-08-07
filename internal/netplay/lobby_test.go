package netplay

import (
	"net"
	"testing"
	"time"
)

// lobby_test.go:大廳的護欄。全部走真的 TCP loopback——大廳的重點就是「連得上、
// 而且每個人拿到的名冊一樣」,用假的 io.Pipe 驗不到 Accept/Dial 那一段。

func TestLobbyAssignsIDsAndBroadcastsTheSameRoster(t *testing.T) {
	lb, err := Host("127.0.0.1:0", "主機玩家", 4242)
	if err != nil {
		t.Skipf("這個環境開不了 socket:%v", err)
	}
	defer lb.Close()

	type joined struct {
		id int
		r  Roster
	}
	results := make(chan joined, 2)
	errs := make(chan error, 2)
	for _, name := range []string{"甲", "乙"} {
		go func(name string) {
			c, id, r, err := Join(lb.Addr(), name, 3*time.Second)
			if err != nil {
				errs <- err
				return
			}
			defer c.Close()
			results <- joined{id, r}
		}(name)
	}

	for i := 0; i < 2; i++ {
		if _, err := lb.AcceptOne(3 * time.Second); err != nil {
			t.Fatalf("第 %d 位加入失敗:%v", i+1, err)
		}
	}

	seenIDs := map[int]bool{}
	for i := 0; i < 2; i++ {
		select {
		case j := <-results:
			if seenIDs[j.id] {
				t.Errorf("玩家編號 %d 被指派了兩次——編號就是鎖步的排序鍵,重複會直接分岔", j.id)
			}
			seenIDs[j.id] = true
			if j.id == 0 {
				t.Error("0 號是主機,客戶端不該拿到 0")
			}
			if j.r.Seed != 4242 {
				t.Errorf("種子應由主機廣播(4242),實得 %d", j.r.Seed)
			}
		case err := <-errs:
			t.Fatalf("加入失敗:%v", err)
		case <-time.After(3 * time.Second):
			t.Fatal("等不到加入結果")
		}
	}

	r := lb.Roster()
	if len(r.Players) != 3 {
		t.Fatalf("名冊應有 3 人(主機 + 兩位),實得 %d", len(r.Players))
	}
	if r.Players[0].ID != 0 || r.Players[0].Name != "主機玩家" {
		t.Errorf("0 號應是主機,實得 %+v", r.Players[0])
	}
	// 名冊必須依編號排序:編號就是鎖步的排序鍵,名冊順序漂了 UI 顯示也會跟著漂。
	for i := 1; i < len(r.Players); i++ {
		if r.Players[i-1].ID >= r.Players[i].ID {
			t.Errorf("名冊沒有依編號排序:%+v", r.Players)
			break
		}
	}
}

// 第一則不是 hello 就要被擋下來——那代表對面不是這個遊戲(或版本不同)。
func TestLobbyRejectsNonHelloFirstMessage(t *testing.T) {
	lb, err := Host("127.0.0.1:0", "主機", 1)
	if err != nil {
		t.Skipf("這個環境開不了 socket:%v", err)
	}
	defer lb.Close()

	go func() {
		c, err := dialRaw(lb.Addr())
		if err != nil {
			return
		}
		defer c.Close()
		_ = WriteFrame(c, Message{Kind: KindTurnDone, Turn: 1})
	}()

	if _, err := lb.AcceptOne(3 * time.Second); err == nil {
		t.Error("第一則不是 hello 應該被擋下來")
	}
}

// 連不上的位址要快快失敗並講清楚,不要卡住 UI。
func TestJoinFailsFastOnBadAddress(t *testing.T) {
	start := time.Now()
	_, _, _, err := Join("127.0.0.1:1", "甲", 500*time.Millisecond)
	if err == nil {
		t.Fatal("連到不存在的位址應該失敗")
	}
	if d := time.Since(start); d > 3*time.Second {
		t.Errorf("失敗花了 %v,逾時沒有生效", d)
	}
}

// dialRaw 是測試用的裸連線(不送 hello)。
func dialRaw(addr string) (net.Conn, error) { return net.DialTimeout("tcp", addr, 3*time.Second) }
