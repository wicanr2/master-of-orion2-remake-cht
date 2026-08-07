package netplay

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

// lobby.go:**開局前的連線大廳**——主機聽、客戶端連、雙方拿到同一份玩家名冊。
//
// 這是連線流程裡唯一「傳輸層做得完」的部分:名冊本身不需要知道遊戲規則。
// 名冊到齊之後才輪到規則層(同一顆種子開局、之後走 lockstep.go 的鎖步)。
//
// ============ 為什麼名冊要由主機廣播,而不是各自累積 ============
//
// 每個人各自累積「我看到誰連進來了」會得到**順序不同**的名冊,而玩家編號就是名冊索引——
// 編號不同,鎖步排序(lockstep.go 的 Commands())就會在不同機器上排出不同順序。
// 主機是唯一的權威:它決定編號,其他人照抄。
//
// ⚠ 誠實留白:沒有斷線重連、沒有心跳、沒有加密。這是區網對戰的最小可用形狀,
// 不是能上公網的實作。真要上公網還需要 NAT 穿透與抗惡意端的驗證,那是另一件事。

// Player 是名冊上的一位玩家。
type Player struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Roster 是主機廣播的玩家名冊。
type Roster struct {
	Players []Player `json:"players"`
	// Seed 是這一局的起始種子。**由主機決定並廣播**——兩邊各自抽種子的話,
	// 銀河會長得不一樣,那是連鎖步都還沒開始就已經分岔。
	Seed int64 `json:"seed"`
}

// KindRoster 是主機廣播名冊的訊息種類。
const KindRoster MsgKind = "roster"

// rosterMessage 把名冊塞進 Message 的一則(用 Detail 帶 JSON 太醜,另外定一個型別)。
type rosterMessage struct {
	Kind    MsgKind  `json:"kind"`
	Players []Player `json:"players"`
	Seed    int64    `json:"seed"`
	// You 是收訊端自己的玩家編號——由主機指派,客戶端不自己決定。
	You int `json:"you"`
}

// Lobby 是主機端的大廳。
type Lobby struct {
	mu      sync.Mutex
	ln      net.Listener
	seed    int64
	players []Player
	conns   map[int]net.Conn
}

// Host 開一個大廳並開始聽。addr 用 "host:port";port 給 0 讓系統挑(測試用)。
//
// hostName 是主機玩家自己的名字——他是 0 號,永遠在名冊第一位。
func Host(addr, hostName string, seed int64) (*Lobby, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("netplay: 聽不起來(%s):%w", addr, err)
	}
	return &Lobby{
		ln: ln, seed: seed,
		players: []Player{{ID: 0, Name: hostName}},
		conns:   map[int]net.Conn{},
	}, nil
}

// Addr 回傳實際監聽的位址(port 給 0 時用它拿到系統挑的那個)。
func (l *Lobby) Addr() string { return l.ln.Addr().String() }

// Close 關掉大廳與所有連線。
func (l *Lobby) Close() error {
	l.mu.Lock()
	for _, c := range l.conns {
		c.Close()
	}
	l.conns = map[int]net.Conn{}
	l.mu.Unlock()
	return l.ln.Close()
}

// Roster 回傳目前的名冊(複本)。
func (l *Lobby) Roster() Roster {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := Roster{Seed: l.seed, Players: append([]Player(nil), l.players...)}
	sort.Slice(out.Players, func(i, j int) bool { return out.Players[i].ID < out.Players[j].ID })
	return out
}

// AcceptOne 收一位玩家:等對方連上、讀他的 hello、指派編號、把新名冊廣播給所有人。
//
// 逐位收而不是背景 goroutine 一直收:UI 端要能一位一位顯示「誰進來了」,
// 而且逐位收讓錯誤處理留在呼叫端,不必在 goroutine 之間傳錯誤。
func (l *Lobby) AcceptOne(timeout time.Duration) (Player, error) {
	if timeout > 0 {
		if tl, ok := l.ln.(*net.TCPListener); ok {
			_ = tl.SetDeadline(time.Now().Add(timeout))
		}
	}
	c, err := l.ln.Accept()
	if err != nil {
		return Player{}, fmt.Errorf("netplay: 接受連線失敗:%w", err)
	}
	var hello Message
	if err := ReadFrame(c, &hello); err != nil {
		c.Close()
		return Player{}, fmt.Errorf("netplay: 對方沒送出 hello:%w", err)
	}
	if hello.Kind != KindHello {
		c.Close()
		return Player{}, fmt.Errorf("netplay: 第一則應是 hello,收到 %q", hello.Kind)
	}

	l.mu.Lock()
	p := Player{ID: len(l.players), Name: hello.Name}
	l.players = append(l.players, p)
	l.conns[p.ID] = c
	l.mu.Unlock()

	if err := l.broadcastRoster(); err != nil {
		return p, err
	}
	return p, nil
}

// broadcastRoster 把目前名冊送給每一位已連線的客戶端。
func (l *Lobby) broadcastRoster() error {
	r := l.Roster()
	l.mu.Lock()
	conns := make(map[int]net.Conn, len(l.conns))
	for id, c := range l.conns {
		conns[id] = c
	}
	l.mu.Unlock()
	for id, c := range conns {
		if err := WriteFrame(c, rosterMessage{Kind: KindRoster, Players: r.Players, Seed: r.Seed, You: id}); err != nil {
			return fmt.Errorf("netplay: 廣播名冊給玩家 %d 失敗:%w", id, err)
		}
	}
	return nil
}

// Conn 回傳某位客戶端的連線(對局開始後由鎖步層使用)。
func (l *Lobby) Conn(id int) net.Conn {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.conns[id]
}

// --- 客戶端 ---

// Join 連上主機、自報名字,並讀回主機指派的名冊。
//
// 回傳連線本身(對局開始後由鎖步層使用)、自己的玩家編號、與名冊。
func Join(addr, name string, timeout time.Duration) (net.Conn, int, Roster, error) {
	c, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, -1, Roster{}, fmt.Errorf("netplay: 連不上 %s:%w", addr, err)
	}
	if err := WriteFrame(c, Message{Kind: KindHello, Name: name}); err != nil {
		c.Close()
		return nil, -1, Roster{}, err
	}
	if timeout > 0 {
		_ = c.SetReadDeadline(time.Now().Add(timeout))
	}
	var r rosterMessage
	if err := ReadFrame(c, &r); err != nil {
		c.Close()
		return nil, -1, Roster{}, fmt.Errorf("netplay: 沒收到名冊:%w", err)
	}
	_ = c.SetReadDeadline(time.Time{})
	if r.Kind != KindRoster {
		c.Close()
		return nil, -1, Roster{}, fmt.Errorf("netplay: 應收到名冊,收到 %q", r.Kind)
	}
	return c, r.You, Roster{Players: r.Players, Seed: r.Seed}, nil
}
