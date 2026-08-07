package netplay

import (
	"errors"
	"io"
	"net"
	"sync"
)

// session.go:對局期間的**訊息幫浦**——把已連上的 net.Conn 變成「可以每幀問一次有沒有新訊息」。
//
// ============ 這一層先前是空的 ============
//
// 大廳(lobby.go)負責把大家連起來、發名冊,然後就結束了;回合表(lockstep.go)是純資料結構,
// 自己不碰任何 socket。中間**沒有東西在跑**:連線建立之後沒有任何 goroutine 在讀它。
//
// 所以聊天列的資料模型(chat.go)、輸入與繪製(cmd/moo2/netnextturn.go)雖然都做完了,
// 打出去的字**從來沒有離開過本機**——`sendChat` 只把它加進自己的記錄。這一檔補的就是那一段。
//
// ============ 為什麼要 Poll 而不是回呼 ============
//
// ebiten 的 `Update` 是單一 goroutine 每幀跑一次。讀 socket 一定要另一條 goroutine
// (不然畫面會卡在 `Read` 上),但**規則狀態不能被那條 goroutine 直接改**——那是資料競爭,
// 而且會讓「同樣的指令序列算出同樣結果」這個鎖步前提失效(改動時機取決於封包什麼時候到)。
//
// 所以中間隔一個佇列:讀取 goroutine 只管往佇列塞,`Update` 每幀 `Poll()` 一次把它清空,
// **所有狀態變更都發生在 Update 這一條線上**。順序由佇列保證,與封包抵達的時間點無關。
//
// ============ 星狀拓樸與轉發(與原版的差異,誠實標註)============
//
// 原版 `Send_Chat_Msg_` @ 0xDD3B8 是**逐一走過玩家陣列直接發給每一個人**——IPX 廣播式的
// 對等網路,誰都能直接送到誰。
//
// remake 走 TCP:大廳是主機 listen、客戶端 dial,客戶端之間**沒有連線**。
// 所以要讓 A 說的話 B 聽得到,主機必須轉發。`Session` 的 `relay` 就是這件事,
// 主機端開 true、客戶端開 false。
//
// 這是**移植決策不是還原**:換成 TCP 星狀拓樸的必然結果,不是原版的行為。
// (要做成對等網路就得讓每個客戶端互相 dial,連線數 N²、而且家用 NAT 後面根本連不上。)

// ErrSessionClosed 是對已關閉的 Session 送訊。
var ErrSessionClosed = errors.New("netplay: 連線已關閉")

// Session 是一局網路對戰期間的連線集合與收訊佇列。
//
// 零值不可用,請用 NewSession。可以從多條 goroutine 呼叫。
type Session struct {
	me    int
	relay bool

	mu     sync.Mutex
	conns  map[int]net.Conn
	inbox  []Message
	err    error
	closed bool

	wg sync.WaitGroup
}

// NewSession 開一個訊息幫浦。
//
//   - me 是本方玩家編號(送出去的訊息會蓋上它,不信任呼叫端自己填的 Player)。
//   - conns 是「玩家編號 → 已連上的連線」。主機端放所有客戶端;客戶端只放主機那一條
//     (鍵用主機的玩家編號)。
//   - relay=true 時,收到的訊息會轉發給**其他**連線(主機端用;見檔頭「星狀拓樸」)。
//
// 每條連線各起一個讀取 goroutine,Close 時一起收掉。
func NewSession(me int, relay bool, conns map[int]net.Conn) *Session {
	s := &Session{me: me, relay: relay, conns: map[int]net.Conn{}}
	for id, c := range conns {
		if c == nil {
			continue
		}
		s.conns[id] = c
	}
	for id, c := range s.conns {
		s.wg.Add(1)
		go s.readLoop(id, c)
	}
	return s
}

// Me 回傳本方玩家編號。
func (s *Session) Me() int { return s.me }

// Peers 回傳目前還連著的對手編號數。
func (s *Session) Peers() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.conns)
}

// readLoop 是單條連線的讀取迴圈。
//
// 讀到 EOF 當成「對面正常離線」——**不記成錯誤**。真正的錯誤(壞封包、連線異常)才記,
// 因為那兩件事對玩家的意思完全不同:一個是「他走了」,一個是「這局可能已經不同步了」。
func (s *Session) readLoop(from int, c net.Conn) {
	defer s.wg.Done()
	for {
		var m Message
		err := ReadFrame(c, &m)
		if err != nil {
			s.dropConn(from)
			if !errors.Is(err, io.EOF) && !s.isClosed() {
				s.setErr(err)
			}
			return
		}
		// 發話者的認定分兩種情況,不能一律蓋掉:
		//
		//   - **主機**(relay=true)收到的每一則都是**第一手**——那條連線就是那個玩家。
		//     封包裡自報的 Player 不採信,以連線為準(否則客戶端可以冒名說話)。
		//   - **客戶端**(relay=false)只有一條連線(對主機),收到的訊息**可能是主機轉發**
		//     別人說的。這時候蓋成「從主機來的」會把原始發話者抹掉,所有人的話都變成主機說的。
		//
		// 客戶端因此信任主機標的編號。這是星狀拓樸的必然:主機本來就轉發所有流量,
		// 它要造假的話蓋不蓋 Player 都擋不住。
		if s.relay {
			m.Player = from
		}
		s.push(m)
		if s.relay {
			s.forward(from, m)
		}
	}
}

// forward 把收到的訊息轉給除了來源以外的所有連線(只有主機會做)。
//
// 轉發失敗不中斷:一條線斷掉不該讓其他人的聊天跟著停。斷掉的那條由它自己的
// readLoop 收拾。
func (s *Session) forward(from int, m Message) {
	s.mu.Lock()
	targets := make([]net.Conn, 0, len(s.conns))
	for id, c := range s.conns {
		if id != from {
			targets = append(targets, c)
		}
	}
	s.mu.Unlock()
	for _, c := range targets {
		_ = WriteFrame(c, m)
	}
}

// Send 把一則訊息送給所有連線。Player 一律蓋成本方編號。
//
// **不會**把它加進自己的收訊佇列:本機的顯示由呼叫端自己處理(聊天列就是這樣——
// 打完 Enter 立刻看到自己那一行,不必等封包繞一圈回來)。
func (s *Session) Send(m Message) error {
	m.Player = s.me
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return ErrSessionClosed
	}
	targets := make([]net.Conn, 0, len(s.conns))
	for _, c := range s.conns {
		targets = append(targets, c)
	}
	s.mu.Unlock()

	var firstErr error
	for _, c := range targets {
		if err := WriteFrame(c, m); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// SendChat 是聊天列用的捷徑。空字串(或只有空白)不送——原版按 Enter 沒打字也是什麼都不做。
func (s *Session) SendChat(text string) error {
	m := ChatMessage(s.me, text)
	if m.Text == "" {
		return nil
	}
	return s.Send(m)
}

// Poll 取出並清空目前收到的訊息(不阻塞)。給每幀呼叫。
//
// 回傳的順序就是收到的順序。沒有新訊息時回 nil。
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

// Err 回傳第一個非「對面正常離線」的錯誤(沒有就回 nil)。
func (s *Session) Err() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.err
}

// Close 關掉所有連線並等讀取 goroutine 結束。可重複呼叫。
func (s *Session) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	conns := make([]net.Conn, 0, len(s.conns))
	for _, c := range s.conns {
		conns = append(conns, c)
	}
	s.conns = map[int]net.Conn{}
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

func (s *Session) dropConn(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.conns, id)
}

func (s *Session) setErr(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.err == nil {
		s.err = err
	}
}

func (s *Session) isClosed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}
