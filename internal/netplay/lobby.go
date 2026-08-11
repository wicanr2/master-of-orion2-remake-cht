package netplay

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"

	"crypto/tls"
)

// ErrLobbyFull 表示此大廳已達呼叫端設定的玩家上限；接受迴圈可繼續等待下一位。
var ErrLobbyFull = errors.New("netplay: 大廳玩家已滿")

// ErrLobbyClosed 表示大廳已關閉，不應再重試 AcceptOne。
var ErrLobbyClosed = errors.New("netplay: 大廳已關閉")

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
// 可選強化由 HostWithOptions / JoinWithOptions 開啟：每次連線 challenge、共享
// 密碼 proof、TLS 1.3 與 resume token。NAT 穿透仍不在本套件內，公網部署仍需
// 外部 relay 或 UPnP；這裡不把區網 discovery 冒稱成 NAT 解法。

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
	// ResumeToken 只填給 You 對應的連線；不會把其他玩家的 token 廣播出去。
	ResumeToken string `json:"resumeToken,omitempty"`
}

// Lobby 是主機端的大廳。
type Lobby struct {
	mu            sync.Mutex
	ln            net.Listener
	setDeadline   func(time.Time) error
	seed          int64
	players       []Player
	conns         map[int]net.Conn
	opts          LobbyOptions
	resumeTokens  map[int]string
	resumeOwners  map[string]int
	reconnectOnly bool
	onReconnect   func(Player, net.Conn)
	maxPlayers    int
	closed        bool
}

// Host 開一個大廳並開始聽。addr 用 "host:port";port 給 0 讓系統挑(測試用)。
//
// hostName 是主機玩家自己的名字——他是 0 號,永遠在名冊第一位。
func Host(addr, hostName string, seed int64) (*Lobby, error) {
	return HostWithOptions(addr, hostName, seed, LobbyOptions{})
}

// HostWithOptions 開一個帶可選 challenge、TLS 與重連 token 的大廳。
func HostWithOptions(addr, hostName string, seed int64, opts LobbyOptions) (*Lobby, error) {
	base, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("netplay: 聽不起來(%s):%w", addr, err)
	}
	setDeadline := func(deadline time.Time) error { return nil }
	if tl, ok := base.(*net.TCPListener); ok {
		setDeadline = tl.SetDeadline
	}
	ln := net.Listener(base)
	if opts.EnableTLS {
		cfg, err := serverTLSConfig(opts)
		if err != nil {
			_ = base.Close()
			return nil, err
		}
		ln = tls.NewListener(base, cfg)
	}
	return &Lobby{
		ln: ln, setDeadline: setDeadline, seed: seed,
		players:      []Player{{ID: 0, Name: hostName}},
		conns:        map[int]net.Conn{},
		opts:         opts,
		resumeTokens: map[int]string{},
		resumeOwners: map[string]int{},
	}, nil
}

// Addr 回傳實際監聽的位址(port 給 0 時用它拿到系統挑的那個)。
func (l *Lobby) Addr() string { return l.ln.Addr().String() }

// Close 關掉大廳與所有連線。
func (l *Lobby) Close() error {
	l.mu.Lock()
	l.closed = true
	for _, c := range l.conns {
		c.Close()
	}
	l.conns = map[int]net.Conn{}
	l.mu.Unlock()
	return l.ln.Close()
}

// StopAccepting 關掉主機的 listen socket，但保留已加入玩家的連線。
// 開局後要停止收新玩家、又要把既有連線交給 Session 使用時，不能呼叫 Close。
//
// 舊呼叫端仍可用，但它會同時關掉重連入口；新的網路開局流程應改用
// SetReconnectOnly，讓既有玩家可以用 resume token 回來。
func (l *Lobby) StopAccepting() error {
	l.mu.Lock()
	l.closed = true
	l.mu.Unlock()
	return l.ln.Close()
}

// SetMaxPlayers 設定新玩家上限；重連不消耗新名額。零值表示不限制。
func (l *Lobby) SetMaxPlayers(max int) {
	l.mu.Lock()
	l.maxPlayers = max
	l.mu.Unlock()
}

// SetReconnectOnly 停止新玩家加入，但保留 listen socket 接受既有玩家重連。
func (l *Lobby) SetReconnectOnly() {
	l.mu.Lock()
	l.reconnectOnly = true
	l.mu.Unlock()
}

// ReconnectOnly 回傳大廳是否已進入只允許重連的狀態。
func (l *Lobby) ReconnectOnly() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.reconnectOnly
}

// SetReconnectHandler 設定主機在驗證 resume token 後替換 Session 連線的回呼。
// 回呼不在 Lobby 鎖內執行；它可以安全呼叫 Session.ReplacePeer。
func (l *Lobby) SetReconnectHandler(handler func(Player, net.Conn)) {
	l.mu.Lock()
	l.onReconnect = handler
	l.mu.Unlock()
}

// ResumeToken 回傳指定玩家的重連 token。token 不會出現在 Roster，也不應寫進日誌。
func (l *Lobby) ResumeToken(id int) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.resumeTokens[id]
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
	l.mu.Lock()
	closed := l.closed
	l.mu.Unlock()
	if closed {
		return Player{}, ErrLobbyClosed
	}
	if timeout > 0 && l.setDeadline != nil {
		_ = l.setDeadline(time.Now().Add(timeout))
	}
	if l.setDeadline != nil {
		defer l.setDeadline(time.Time{})
	}
	c, err := l.ln.Accept()
	if err != nil {
		return Player{}, fmt.Errorf("netplay: 接受連線失敗:%w", err)
	}
	challenge, err := newChallenge()
	if err != nil {
		_ = c.Close()
		return Player{}, err
	}
	if err := WriteFrame(c, Message{Kind: KindHelloChallenge, Protocol: ProtocolVersion, Challenge: challenge}); err != nil {
		_ = c.Close()
		return Player{}, fmt.Errorf("netplay: 傳送 hello challenge：%w", err)
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
	if hello.Protocol != ProtocolVersion || !verifyAuth(l.opts.AuthToken, challenge, hello.Auth) {
		c.Close()
		return Player{}, fmt.Errorf("netplay: 協定版本或身份驗證失敗")
	}

	var (
		p         Player
		old       net.Conn
		reconnect bool
		handler   func(Player, net.Conn)
		resume    string
	)
	if hello.ResumeToken == "" {
		var tokenErr error
		resume, tokenErr = newResumeToken()
		if tokenErr != nil {
			_ = c.Close()
			return Player{}, tokenErr
		}
	}

	l.mu.Lock()
	if hello.ResumeToken != "" {
		id, ok := l.resumeOwners[hello.ResumeToken]
		if !ok || l.resumeTokens[id] != hello.ResumeToken {
			l.mu.Unlock()
			_ = c.Close()
			return Player{}, fmt.Errorf("netplay: 無效或過期的 resume token")
		}
		p = l.players[id]
		old = l.conns[id]
		l.conns[id] = c
		reconnect = true
		resume = hello.ResumeToken
		handler = l.onReconnect
	} else {
		if l.maxPlayers > 0 && len(l.players) >= l.maxPlayers {
			l.mu.Unlock()
			_ = c.Close()
			return Player{}, ErrLobbyFull
		}
		if l.reconnectOnly {
			l.mu.Unlock()
			_ = c.Close()
			return Player{}, fmt.Errorf("netplay: 對局已開始，只接受既有玩家重連")
		}
		p = Player{ID: len(l.players), Name: hello.Name}
		l.players = append(l.players, p)
		l.conns[p.ID] = c
		l.resumeTokens[p.ID] = resume
		l.resumeOwners[resume] = p.ID
	}
	l.mu.Unlock()
	if old != nil {
		_ = old.Close()
	}

	if reconnect {
		if err := l.writeRoster(p.ID, c, resume); err != nil {
			return p, err
		}
		if handler != nil {
			handler(p, c)
		}
	} else if err := l.broadcastRoster(); err != nil {
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
		if err := l.writeRosterWithRoster(id, c, r, l.ResumeToken(id)); err != nil {
			return fmt.Errorf("netplay: 廣播名冊給玩家 %d 失敗:%w", id, err)
		}
	}
	return nil
}

func (l *Lobby) writeRoster(id int, c net.Conn, resume string) error {
	return l.writeRosterWithRoster(id, c, l.Roster(), resume)
}

func (l *Lobby) writeRosterWithRoster(id int, c net.Conn, r Roster, resume string) error {
	return WriteFrame(c, rosterMessage{
		Kind: KindRoster, Players: r.Players, Seed: r.Seed,
		You: id, ResumeToken: resume,
	})
}

// Conn 回傳某位客戶端的連線(對局開始後由鎖步層使用)。
func (l *Lobby) Conn(id int) net.Conn {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.conns[id]
}

// Connections 回傳目前已加入客戶端連線的複本，供對局 Session 接管。
// 呼叫端不應關閉其中的連線；生命週期由 Lobby.Close 或 Session.Close 負責。
func (l *Lobby) Connections() map[int]net.Conn {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make(map[int]net.Conn, len(l.conns))
	for id, c := range l.conns {
		out[id] = c
	}
	return out
}

// Broadcast 把一則已由主機決定的訊息送給所有已加入的客戶端。
// 它用於共同開局快照；對局開始後的一般訊息交給 Session。
func (l *Lobby) Broadcast(m Message) error {
	l.mu.Lock()
	conns := make(map[int]net.Conn, len(l.conns))
	for id, c := range l.conns {
		conns[id] = c
	}
	l.mu.Unlock()
	for id, c := range conns {
		if err := WriteFrame(c, m); err != nil {
			return fmt.Errorf("netplay: 廣播訊息給玩家 %d 失敗:%w", id, err)
		}
	}
	return nil
}

// --- 客戶端 ---

// Join 連上主機、自報名字,並讀回主機指派的名冊。
//
// 回傳連線本身(對局開始後由鎖步層使用)、自己的玩家編號、與名冊。
func Join(addr, name string, timeout time.Duration) (net.Conn, int, Roster, error) {
	c, id, r, _, err := JoinWithOptions(addr, name, timeout, JoinOptions{})
	return c, id, r, err
}

// JoinWithOptions 連上主機並回傳該玩家的 resume token。token 應只放在記憶體中，
// 或由呼叫端以平台安全儲存保存；不要寫進公開 log。
func JoinWithOptions(addr, name string, timeout time.Duration, opts JoinOptions) (net.Conn, int, Roster, string, error) {
	c, err := dialLobby(addr, timeout, opts.LobbyOptions)
	if err != nil {
		return nil, -1, Roster{}, "", fmt.Errorf("netplay: 連不上 %s:%w", addr, err)
	}
	if timeout > 0 {
		_ = c.SetDeadline(time.Now().Add(timeout))
	}
	var challenge Message
	if err := ReadFrame(c, &challenge); err != nil {
		_ = c.Close()
		return nil, -1, Roster{}, "", fmt.Errorf("netplay: 沒收到 hello challenge:%w", err)
	}
	if challenge.Kind != KindHelloChallenge || challenge.Protocol != ProtocolVersion || challenge.Challenge == "" {
		_ = c.Close()
		return nil, -1, Roster{}, "", fmt.Errorf("netplay: 主機 hello challenge 無效")
	}
	if err := WriteFrame(c, Message{
		Kind: KindHello, Protocol: ProtocolVersion, Name: name,
		Auth: authProof(opts.AuthToken, challenge.Challenge), ResumeToken: opts.ResumeToken,
	}); err != nil {
		c.Close()
		return nil, -1, Roster{}, "", err
	}
	var r rosterMessage
	if err := ReadFrame(c, &r); err != nil {
		c.Close()
		return nil, -1, Roster{}, "", fmt.Errorf("netplay: 沒收到名冊:%w", err)
	}
	_ = c.SetDeadline(time.Time{})
	if r.Kind != KindRoster {
		c.Close()
		return nil, -1, Roster{}, "", fmt.Errorf("netplay: 應收到名冊,收到 %q", r.Kind)
	}
	if r.ResumeToken == "" {
		return nil, -1, Roster{}, "", fmt.Errorf("netplay: 主機沒有回傳 resume token")
	}
	return c, r.You, Roster{Players: r.Players, Seed: r.Seed}, r.ResumeToken, nil
}

func dialLobby(addr string, timeout time.Duration, opts LobbyOptions) (net.Conn, error) {
	raw, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	if !opts.EnableTLS {
		return raw, nil
	}
	c := tls.Client(raw, clientTLSConfig(opts))
	if timeout > 0 {
		_ = c.SetDeadline(time.Now().Add(timeout))
	}
	if err := c.Handshake(); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("netplay: TLS handshake：%w", err)
	}
	if timeout > 0 {
		_ = c.SetDeadline(time.Time{})
	}
	return c, nil
}
