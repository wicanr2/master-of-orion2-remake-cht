package netplay

import (
	"encoding/json"
	"errors"
	"net"
	"sort"
	"strconv"
	"sync"
	"time"
)

// discovery.go:**區網對局探索**——讓「加入對局」不必先知道對方的 IP。
//
// ============ 為什麼要做這一層(不是為了好看)============
//
// 原版走 IPX,而 IPX 自帶廣播式的服務公告:`Choose_Multi_Net_Game` 那張畫面之所以能
// 「列出區網上有哪些對局」,是協定本身給的能力,不是遊戲自己做的。
//
// remake 走 TCP,TCP **沒有**這個能力。照抄畫面而不補這一層,會得到一張永遠空的清單
// ——那不是還原,是把畫面畫出來而已。所以這裡用 UDP 廣播補回等價的能力:
//
//	主機:每秒往 255.255.255.255:24502 送一份「我在這裡」(名字、TCP 位址、人數)
//	客戶端:聽 24502,在一個時間窗內收集,依 TCP 位址去重
//
// ⚠ **這是移植決策,不是還原。** 原版沒有這段程式碼,因為它不需要。標在這裡,
// 免得日後被當成「從反組譯抄來的」。
//
// ============ 誠實留白 ============
//
// UDP 廣播只跨得過同一個廣播網域——換句話說「同一個區網」。跨網段要打對方位址,
// 那要文字輸入框(還沒做)。原版的 IPX 也是同一個限制,所以這不是退步。
//
// 沒有簽章、沒有加密:區網上任何人都能廣播一個假的對局。這與大廳本身的立場一致
// (見 lobby.go)——區網對戰的最低限度,不是給公開網路用的。

// DiscoveryPort 是探索用的 UDP 埠。
//
// 與大廳的 TCP 24501 相鄰但不同:同一個埠號不能同時給兩種協定用得清楚,
// 而相鄰的號碼在防火牆規則裡好寫成一段。
const DiscoveryPort = 24502

// GameNameMax 是對局名稱的長度上限。
//
// 真值:原版 `Change_MP_Game_Name_` @ 0xF5777 呼叫輸入常式時 `edx = 8`。
// 同一段還會把新名字與既有對局逐一 strcmp,撞名就跳訊息——也就是**名稱要唯一**。
const GameNameMax = 8

// Game 是探索到的一場對局。
type Game struct {
	Name    string `json:"name"`
	Addr    string `json:"addr"` // 大廳的 TCP 位址(host:port)
	Players int    `json:"players"`
	Max     int    `json:"max"`
}

// beacon 是廣播出去的封包內容。欄位與 Game 同形,另加一個標記避免把別的程式的
// UDP 封包當成對局(24502 沒有被指派,但「沒被指派」不等於「沒有人用」)。
type beacon struct {
	Magic string `json:"moo2"`
	Game
}

const beaconMagic = "moo2-lobby-1"

// Announcer 是主機端的廣播器。
type Announcer struct {
	conn  *net.UDPConn
	dst   *net.UDPAddr
	stop  chan struct{}
	once  sync.Once
	mu    sync.Mutex
	game  Game
	every time.Duration
}

// Announce 開始廣播一場對局。
//
// dst 留空時用 255.255.255.255:24502(區網廣播);測試會傳 127.0.0.1:埠 以免真的廣播
// ——**測試不該把封包送到辦公室網路上**。
func Announce(g Game, dst string, every time.Duration) (*Announcer, error) {
	if every <= 0 {
		every = time.Second
	}
	if dst == "" {
		dst = net.JoinHostPort("255.255.255.255", strconv.Itoa(DiscoveryPort))
	}
	addr, err := net.ResolveUDPAddr("udp4", dst)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, err
	}
	a := &Announcer{conn: conn, dst: addr, stop: make(chan struct{}), game: g, every: every}
	go a.loop()
	return a, nil
}

// Update 換掉廣播內容(人數會變)。
func (a *Announcer) Update(g Game) {
	a.mu.Lock()
	a.game = g
	a.mu.Unlock()
}

// Close 停止廣播。可重複呼叫。
func (a *Announcer) Close() error {
	a.once.Do(func() { close(a.stop) })
	return a.conn.Close()
}

func (a *Announcer) loop() {
	t := time.NewTicker(a.every)
	defer t.Stop()
	a.send() // 先送一份,不然第一個週期內加入的人什麼都看不到
	for {
		select {
		case <-a.stop:
			return
		case <-t.C:
			a.send()
		}
	}
}

func (a *Announcer) send() {
	a.mu.Lock()
	g := a.game
	a.mu.Unlock()
	b, err := json.Marshal(beacon{Magic: beaconMagic, Game: g})
	if err != nil {
		return
	}
	// 送失敗不處理:區網上暫時送不出去是常態(介面剛換、暫時沒有路由),
	// 而下一秒就會再送一次。這裡回報錯誤只會變成沒有人讀的日誌。
	_, _ = a.conn.WriteToUDP(b, a.dst)
}

// Discover 在 window 這段時間內收集區網上的對局,依 TCP 位址去重。
//
// bind 留空時聽 0.0.0.0:24502;測試傳 127.0.0.1:0 之類的位址以避開真實埠。
// 回傳的清單依名稱排序——**不依到達順序**:到達順序每次都不一樣,而清單順序會決定
// 玩家點到哪一場,那種「每次跑起來不一樣」的介面很難用也很難測。
func Discover(bind string, window time.Duration) ([]Game, error) {
	conn, err := listenDiscovery(bind)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return collect(conn, window)
}

// listenDiscovery 開探索用的 UDP socket。
func listenDiscovery(bind string) (*net.UDPConn, error) {
	if bind == "" {
		bind = net.JoinHostPort("0.0.0.0", strconv.Itoa(DiscoveryPort))
	}
	addr, err := net.ResolveUDPAddr("udp4", bind)
	if err != nil {
		return nil, err
	}
	return net.ListenUDP("udp4", addr)
}

// collect 讀到 window 為止,回傳去重排序後的清單。
func collect(conn *net.UDPConn, window time.Duration) ([]Game, error) {
	if window <= 0 {
		window = 2 * time.Second
	}
	deadline := time.Now().Add(window)
	if err := conn.SetReadDeadline(deadline); err != nil {
		return nil, err
	}
	seen := map[string]Game{}
	buf := make([]byte, 2048)
	for {
		n, from, err := conn.ReadFromUDP(buf)
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				break // 時間窗到了,正常結束
			}
			return sortGames(seen), err
		}
		var b beacon
		if json.Unmarshal(buf[:n], &b) != nil || b.Magic != beaconMagic || b.Addr == "" {
			continue // 不是我們的封包
		}
		// 廣播來源的 IP 比封包裡寫的可信:主機可能不知道自己對外是哪個位址
		//(多網卡、容器內、NAT)。只有埠沿用封包裡的。
		if _, port, err := net.SplitHostPort(b.Addr); err == nil && from != nil {
			b.Addr = net.JoinHostPort(from.IP.String(), port)
		}
		if len(b.Name) > GameNameMax {
			b.Name = b.Name[:GameNameMax]
		}
		seen[b.Addr] = b.Game
	}
	return sortGames(seen), nil
}

func sortGames(m map[string]Game) []Game {
	out := make([]Game, 0, len(m))
	for _, g := range m {
		out = append(out, g)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].Addr < out[j].Addr
	})
	return out
}

// Browser 是「一直聽著」的版本:UI 開著清單時每幀讀一次目前看到的對局。
//
// 為什麼不直接用 Discover:Discover 會**阻塞**一個時間窗,而 UI 是單執行緒的,
// 阻塞兩秒等於畫面凍住兩秒(同 lobby.AcceptOne 的處境,見 choosenetplyrs.go)。
type Browser struct {
	conn *net.UDPConn
	mu   sync.Mutex
	seen map[string]Game
	stop chan struct{}
	once sync.Once
}

// Browse 開始在背景收集。
func Browse(bind string) (*Browser, error) {
	conn, err := listenDiscovery(bind)
	if err != nil {
		return nil, err
	}
	b := &Browser{conn: conn, seen: map[string]Game{}, stop: make(chan struct{})}
	go b.loop()
	return b, nil
}

// Games 回傳目前看到的對局(排序後的快照)。
func (b *Browser) Games() []Game {
	b.mu.Lock()
	defer b.mu.Unlock()
	return sortGames(b.seen)
}

// Close 停止收集。可重複呼叫。
func (b *Browser) Close() error {
	b.once.Do(func() { close(b.stop) })
	return b.conn.Close()
}

func (b *Browser) loop() {
	buf := make([]byte, 2048)
	for {
		select {
		case <-b.stop:
			return
		default:
		}
		// 逐次設短讀取期限,好讓 Close 之後這個 goroutine 能收掉
		//(沒有期限的 ReadFromUDP 會一直卡著,直到有封包來)。
		_ = b.conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
		n, from, err := b.conn.ReadFromUDP(buf)
		if err != nil {
			var ne net.Error
			if errors.As(err, &ne) && ne.Timeout() {
				continue
			}
			return // socket 關了
		}
		var bc beacon
		if json.Unmarshal(buf[:n], &bc) != nil || bc.Magic != beaconMagic || bc.Addr == "" {
			continue
		}
		if _, port, err := net.SplitHostPort(bc.Addr); err == nil && from != nil {
			bc.Addr = net.JoinHostPort(from.IP.String(), port)
		}
		if len(bc.Name) > GameNameMax {
			bc.Name = bc.Name[:GameNameMax]
		}
		b.mu.Lock()
		b.seen[bc.Addr] = bc.Game
		b.mu.Unlock()
	}
}
