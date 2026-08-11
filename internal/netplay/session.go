package netplay

import (
	"errors"
	"net"
	"strings"
	"sync"
	"time"
)

// session.go：對局期間的訊息幫浦、心跳與可選重連。
//
// Update goroutine 只透過 Poll 改遊戲狀態；socket goroutine 只把訊息放進 inbox。
// 心跳 Ping/Pong 在這層消化，不會污染鎖步回合表。斷線後 Session 會保留玩家
// 身份一段 grace 期間；客戶端可用 Reconnect callback 重新加入，主機則可由
// Lobby.SetReconnectHandler 呼叫 ReplacePeer。NAT 穿透仍由外部 relay 或 UPnP
// 負責，本檔不宣稱能穿過家用路由器。

// ErrSessionClosed 是對已關閉的 Session 送訊。
var ErrSessionClosed = errors.New("netplay: 連線已關閉")

// ErrPeerTimeout 表示心跳在允許的重連寬限期後仍沒有恢復。
var ErrPeerTimeout = errors.New("netplay: 對手心跳逾時，重連寬限期已用盡")

// SessionOptions 控制心跳、寫入逾時與重連。
type SessionOptions struct {
	HeartbeatInterval time.Duration
	PeerTimeout       time.Duration
	ReconnectGrace    time.Duration
	ReconnectInterval time.Duration
	WriteTimeout      time.Duration

	// Reconnect 只在客戶端使用；回傳的新連線必須已完成 lobby challenge，
	// 並仍代表同一個 peerID。主機通常留 nil，改由 Lobby 的 callback 替換。
	Reconnect func(peerID int) (net.Conn, error)
}

// DefaultSessionOptions 是公開網路可用的保守預設：5 秒心跳、15 秒偵測、
// 30 秒重連寬限。單元測試可用 NewSessionWithOptions 注入更短的時間。
func DefaultSessionOptions() SessionOptions {
	return SessionOptions{
		HeartbeatInterval: 5 * time.Second,
		PeerTimeout:       15 * time.Second,
		ReconnectGrace:    30 * time.Second,
		ReconnectInterval: 2 * time.Second,
		WriteTimeout:      3 * time.Second,
	}
}

func normalizeSessionOptions(opts SessionOptions) SessionOptions {
	def := DefaultSessionOptions()
	if opts.HeartbeatInterval == 0 {
		opts.HeartbeatInterval = def.HeartbeatInterval
	} else if opts.HeartbeatInterval < 0 {
		opts.HeartbeatInterval = 0
	}
	if opts.PeerTimeout == 0 {
		opts.PeerTimeout = def.PeerTimeout
	} else if opts.PeerTimeout < 0 {
		opts.PeerTimeout = 0
	}
	if opts.ReconnectGrace == 0 {
		opts.ReconnectGrace = def.ReconnectGrace
	} else if opts.ReconnectGrace < 0 {
		opts.ReconnectGrace = 0
	}
	if opts.ReconnectInterval == 0 {
		opts.ReconnectInterval = def.ReconnectInterval
	} else if opts.ReconnectInterval < 0 {
		opts.ReconnectInterval = 0
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = def.WriteTimeout
	} else if opts.WriteTimeout < 0 {
		opts.WriteTimeout = 0
	}
	return opts
}

type sessionPeer struct {
	conn     net.Conn
	writeMu  sync.Mutex
	lastSeen time.Time
}

// Session 是一局網路對戰期間的連線集合與收訊佇列。
// 零值不可用，請用 NewSession 或 NewSessionWithOptions。
type Session struct {
	me    int
	relay bool
	opts  SessionOptions

	mu           sync.Mutex
	peers        map[int]*sessionPeer
	inbox        []Message
	err          error
	closed       bool
	reconnecting map[int]bool
	done         chan struct{}

	wg sync.WaitGroup
}

// NewSession 使用 DefaultSessionOptions 開一個訊息幫浦。
func NewSession(me int, relay bool, conns map[int]net.Conn) *Session {
	return NewSessionWithOptions(me, relay, conns, DefaultSessionOptions())
}

// NewSessionWithOptions 開一個帶心跳與重連策略的訊息幫浦。
//
//   - me 是本方玩家編號；送出的訊息會蓋上它。
//   - conns 是玩家編號 → 已完成 lobby 握手的連線。主機放所有客戶端，
//     客戶端只放主機那一條。
//   - relay=true 時，收到的非內部訊息會轉發給其他連線。
func NewSessionWithOptions(me int, relay bool, conns map[int]net.Conn, opts SessionOptions) *Session {
	opts = normalizeSessionOptions(opts)
	now := time.Now()
	s := &Session{
		me: me, relay: relay, opts: opts,
		peers:        map[int]*sessionPeer{},
		reconnecting: map[int]bool{},
		done:         make(chan struct{}),
	}
	for id, c := range conns {
		if c == nil {
			continue
		}
		s.peers[id] = &sessionPeer{conn: c, lastSeen: now}
	}
	for id, p := range s.peers {
		s.wg.Add(1)
		go s.readLoop(id, p)
	}
	if opts.HeartbeatInterval > 0 {
		s.wg.Add(1)
		go s.heartbeatLoop()
	}
	return s
}

// Me 回傳本方玩家編號。
func (s *Session) Me() int { return s.me }

// Peers 回傳目前仍有有效連線的對手編號數；重連寬限期間會暫時下降。
func (s *Session) Peers() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.peers)
}

func (s *Session) readLoop(from int, p *sessionPeer) {
	defer s.wg.Done()
	for {
		var m Message
		if err := ReadFrame(p.conn, &m); err != nil {
			s.disconnect(from, p, err)
			return
		}
		if !s.touch(from, p) {
			return // 這是被 ReplacePeer 淘汰的舊讀取迴圈。
		}
		switch m.Kind {
		case KindPing:
			if err := s.writePeer(from, p, Message{Kind: KindPong, Player: s.me}); err != nil {
				s.disconnect(from, p, err)
				return
			}
			continue
		case KindPong:
			continue
		}

		// 主機收到第一手訊息時，以 socket 對應的玩家編號為準；客戶端收到
		// 主機轉發時保留主機寫入的原始 Player。
		if s.relay {
			m.Player = from
		}
		s.push(m)
		if s.relay {
			s.forward(from, m)
		}
	}
}

func (s *Session) heartbeatLoop() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.opts.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			for id, p := range s.peerSnapshot() {
				if s.opts.PeerTimeout > 0 && now.Sub(s.lastSeen(id, p)) > s.opts.PeerTimeout {
					s.disconnect(id, p, ErrPeerTimeout)
					continue
				}
				if err := s.writePeer(id, p, Message{Kind: KindPing, Player: s.me}); err != nil {
					s.disconnect(id, p, err)
				}
			}
		case <-s.done:
			return
		}
	}
}

func (s *Session) peerSnapshot() map[int]*sessionPeer {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[int]*sessionPeer, len(s.peers))
	for id, p := range s.peers {
		out[id] = p
	}
	return out
}

func (s *Session) lastSeen(id int, p *sessionPeer) time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	if current := s.peers[id]; current == p {
		return current.lastSeen
	}
	return time.Now()
}

func (s *Session) touch(id int, p *sessionPeer) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.peers[id] != p {
		return false
	}
	p.lastSeen = time.Now()
	return true
}

func (s *Session) writePeer(id int, p *sessionPeer, m Message) error {
	p.writeMu.Lock()
	defer p.writeMu.Unlock()
	s.mu.Lock()
	closed := s.closed || s.peers[id] != p
	s.mu.Unlock()
	if closed {
		return ErrSessionClosed
	}
	if s.opts.WriteTimeout > 0 {
		_ = p.conn.SetWriteDeadline(time.Now().Add(s.opts.WriteTimeout))
		defer p.conn.SetWriteDeadline(time.Time{})
	}
	return WriteFrame(p.conn, m)
}

// disconnect 移除舊 peer，然後保留一段時間給外部 lobby 或 client callback 恢復。
func (s *Session) disconnect(id int, p *sessionPeer, cause error) {
	s.mu.Lock()
	if s.closed || s.peers[id] != p {
		s.mu.Unlock()
		return
	}
	delete(s.peers, id)
	if s.reconnecting[id] {
		s.mu.Unlock()
		_ = p.conn.Close()
		return
	}
	s.reconnecting[id] = true
	s.mu.Unlock()
	_ = p.conn.Close()

	// 封包格式錯誤不是暫時斷線，不能以重連掩蓋協定／版本不相容。
	if protocolError(cause) {
		s.mu.Lock()
		delete(s.reconnecting, id)
		s.mu.Unlock()
		s.setErr(cause)
		return
	}
	s.wg.Add(1)
	go s.reconnectPeer(id)
}

func protocolError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrFrameTooLarge) || strings.Contains(err.Error(), "反序列化失敗")
}

func (s *Session) reconnectPeer(id int) {
	defer s.wg.Done()
	deadline := time.Now().Add(s.opts.ReconnectGrace)
	for {
		if s.isClosed() {
			return
		}
		if s.opts.Reconnect != nil {
			c, err := s.opts.Reconnect(id)
			if err == nil && c != nil {
				if err := s.ReplacePeer(id, c); err == nil {
					return
				}
				_ = c.Close()
			}
		}
		if s.opts.ReconnectGrace <= 0 || time.Now().After(deadline) {
			s.finishReconnect(id)
			return
		}
		wait := s.opts.ReconnectInterval
		if wait <= 0 {
			wait = 50 * time.Millisecond
		}
		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
		case <-s.done:
			if !timer.Stop() {
				<-timer.C
			}
			return
		}
	}
}

func (s *Session) finishReconnect(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.reconnecting, id)
	if !s.closed && s.peers[id] == nil && s.err == nil {
		s.err = ErrPeerTimeout
	}
}

// ReplacePeer 以新的、已完成身份握手的連線取代斷線 peer，並保留原玩家編號。
// 主機的 Lobby.SetReconnectHandler 應在接受 resume token 後呼叫它。
func (s *Session) ReplacePeer(id int, c net.Conn) error {
	if c == nil {
		return errors.New("netplay: 不可用 nil 連線替換 peer")
	}
	p := &sessionPeer{conn: c, lastSeen: time.Now()}
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		_ = c.Close()
		return ErrSessionClosed
	}
	old := s.peers[id]
	s.peers[id] = p
	delete(s.reconnecting, id)
	s.mu.Unlock()
	if old != nil {
		_ = old.conn.Close()
	}
	s.wg.Add(1)
	go s.readLoop(id, p)
	return nil
}

// forward 把收到的訊息轉給除了來源以外的所有連線。
func (s *Session) forward(from int, m Message) {
	for id, p := range s.peerSnapshot() {
		if id == from {
			continue
		}
		if err := s.writePeer(id, p, m); err != nil {
			s.disconnect(id, p, err)
		}
	}
}

// Send 把一則訊息送給所有連線。Player 一律蓋成本方編號。
func (s *Session) Send(m Message) error {
	m.Player = s.me
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrSessionClosed
	}
	s.mu.Unlock()

	var firstErr error
	for id, p := range s.peerSnapshot() {
		if err := s.writePeer(id, p, m); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			s.disconnect(id, p, err)
		}
	}
	return firstErr
}

// SendChat 是聊天列用的捷徑。空字串不送。
func (s *Session) SendChat(text string) error {
	m := ChatMessage(s.me, text)
	if m.Text == "" {
		return nil
	}
	return s.Send(m)
}

// Poll 取出並清空目前收到的訊息，不阻塞。
func (s *Session) Poll() []Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.inbox) == 0 {
		return nil
	}
	out := s.inbox
	s.inbox = nil
	return out
}

// Err 回傳第一個非暫時重連失敗的錯誤。
func (s *Session) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

// Close 關掉所有連線並等讀取／心跳／重連 goroutine 結束。可重複呼叫。
func (s *Session) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	close(s.done)
	conns := make([]net.Conn, 0, len(s.peers))
	for _, p := range s.peers {
		conns = append(conns, p.conn)
	}
	s.peers = map[int]*sessionPeer{}
	s.reconnecting = map[int]bool{}
	s.mu.Unlock()

	var firstErr error
	for _, c := range conns {
		if err := c.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	s.wg.Wait()
	return firstErr
}

func (s *Session) push(m Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.inbox = append(s.inbox, m)
}

func (s *Session) setErr(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err == nil && !s.closed {
		s.err = err
	}
}

func (s *Session) isClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}
