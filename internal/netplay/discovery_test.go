package netplay

import (
	"encoding/json"
	"net"
	"strconv"
	"testing"
	"time"
)

// discovery_test.go:區網探索的護欄。
//
// ⚠ 全部走 **127.0.0.1**,不走真的廣播位址——測試不該把封包送到辦公室網路上。

// 找一個空著的 UDP 埠(讓作業系統挑,再把它讓出來)。
func freeUDPPort(t *testing.T) int {
	t.Helper()
	c, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("拿不到空埠:%v", err)
	}
	port := c.LocalAddr().(*net.UDPAddr).Port
	c.Close()
	return port
}

func loopback(port int) string {
	return net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
}

// 主機廣播 → 客戶端收得到,而且 TCP 位址的**主機部分換成封包來源**。
func TestAnnounceIsPickedUpByDiscover(t *testing.T) {
	port := freeUDPPort(t)
	// 先開聽的那一端,否則第一份廣播會落空。
	conn, err := listenDiscovery(loopback(port))
	if err != nil {
		t.Fatalf("開監聽失敗:%v", err)
	}
	defer conn.Close()

	// 廣播裡故意寫一個「錯的」主機部分(0.0.0.0)——主機常常不知道自己對外是哪個位址。
	a, err := Announce(Game{Name: "ORION", Addr: "0.0.0.0:24501", Players: 2, Max: 8},
		loopback(port), 50*time.Millisecond)
	if err != nil {
		t.Fatalf("開廣播失敗:%v", err)
	}
	defer a.Close()

	games, err := collect(conn, 700*time.Millisecond)
	if err != nil {
		t.Fatalf("收集失敗:%v", err)
	}
	if len(games) != 1 {
		t.Fatalf("應收到 1 場對局,實得 %d 場:%+v", len(games), games)
	}
	g := games[0]
	if g.Name != "ORION" || g.Players != 2 || g.Max != 8 {
		t.Errorf("內容不對:%+v", g)
	}
	if host, _, _ := net.SplitHostPort(g.Addr); host != "127.0.0.1" {
		t.Errorf("主機部分應換成封包來源 127.0.0.1,實得 %q", g.Addr)
	}
	if _, port, _ := net.SplitHostPort(g.Addr); port != "24501" {
		t.Errorf("埠應沿用封包裡的 24501,實得 %q", g.Addr)
	}
}

// 同一場對局廣播很多次,清單裡只能有一筆。
func TestDiscoverDedupesRepeatedBeacons(t *testing.T) {
	port := freeUDPPort(t)
	conn, err := listenDiscovery(loopback(port))
	if err != nil {
		t.Fatalf("開監聽失敗:%v", err)
	}
	defer conn.Close()
	a, err := Announce(Game{Name: "A", Addr: "0.0.0.0:24501"}, loopback(port), 30*time.Millisecond)
	if err != nil {
		t.Fatalf("開廣播失敗:%v", err)
	}
	defer a.Close()
	games, _ := collect(conn, 500*time.Millisecond)
	if len(games) != 1 {
		t.Fatalf("重複廣播應去重成 1 筆,實得 %d 筆", len(games))
	}
}

// 不是我們的 UDP 封包一律略過——24502 沒有被指派,不等於沒有人用。
func TestDiscoverIgnoresForeignPackets(t *testing.T) {
	port := freeUDPPort(t)
	conn, err := listenDiscovery(loopback(port))
	if err != nil {
		t.Fatalf("開監聽失敗:%v", err)
	}
	defer conn.Close()

	send := func(payload []byte) {
		c, err := net.Dial("udp4", loopback(port))
		if err != nil {
			t.Fatalf("送測試封包失敗:%v", err)
		}
		c.Write(payload)
		c.Close()
	}
	send([]byte("這不是 JSON"))
	wrong, _ := json.Marshal(beacon{Magic: "別的程式", Game: Game{Name: "X", Addr: "1.2.3.4:1"}})
	send(wrong)
	noAddr, _ := json.Marshal(beacon{Magic: beaconMagic, Game: Game{Name: "缺位址"}})
	send(noAddr)

	// 正對照:同一輪裡送一份**合法**的,證明過濾器不是把全部都丟掉。
	ok, _ := json.Marshal(beacon{Magic: beaconMagic, Game: Game{Name: "OK", Addr: "0.0.0.0:24501"}})
	send(ok)

	games, _ := collect(conn, 400*time.Millisecond)
	if len(games) != 1 {
		t.Fatalf("只有那份合法的該留下,實得 %d 筆:%+v", len(games), games)
	}
	if games[0].Name != "OK" {
		t.Errorf("留下的應該是 OK,實得 %q", games[0].Name)
	}
}

// Browser 不阻塞:開著收,隨時讀快照。UI 是單執行緒的,阻塞兩秒等於畫面凍住兩秒。
func TestBrowserCollectsInTheBackground(t *testing.T) {
	port := freeUDPPort(t)
	br, err := Browse(loopback(port))
	if err != nil {
		t.Fatalf("開 Browser 失敗:%v", err)
	}
	defer br.Close()
	if len(br.Games()) != 0 {
		t.Fatal("剛開始不該有對局")
	}
	a, err := Announce(Game{Name: "LAN", Addr: "0.0.0.0:24501"}, loopback(port), 30*time.Millisecond)
	if err != nil {
		t.Fatalf("開廣播失敗:%v", err)
	}
	defer a.Close()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(br.Games()) > 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("兩秒內 Browser 一場都沒收到")
}

// 清單依名稱排序,不依到達順序——到達順序每次都不一樣,而順序決定玩家點到哪一場。
func TestGamesAreSortedByName(t *testing.T) {
	got := sortGames(map[string]Game{
		"c:1": {Name: "ZULU", Addr: "c:1"},
		"a:1": {Name: "ALPHA", Addr: "a:1"},
		"b:1": {Name: "MIKE", Addr: "b:1"},
	})
	want := []string{"ALPHA", "MIKE", "ZULU"}
	for i, w := range want {
		if got[i].Name != w {
			t.Errorf("第 %d 筆應為 %s,實得 %s", i, w, got[i].Name)
		}
	}
}

// 名稱上限 8:真值取自 Change_MP_Game_Name_ @ 0xF5777 的 edx = 8。
func TestLongNamesAreTruncatedToTheOriginalLimit(t *testing.T) {
	if GameNameMax != 8 {
		t.Fatalf("名稱上限應為 8(原版 Change_MP_Game_Name_ 的 edx),實得 %d", GameNameMax)
	}
	port := freeUDPPort(t)
	conn, err := listenDiscovery(loopback(port))
	if err != nil {
		t.Fatalf("開監聽失敗:%v", err)
	}
	defer conn.Close()
	a, err := Announce(Game{Name: "ABCDEFGHIJKL", Addr: "0.0.0.0:24501"}, loopback(port), 30*time.Millisecond)
	if err != nil {
		t.Fatalf("開廣播失敗:%v", err)
	}
	defer a.Close()
	games, _ := collect(conn, 400*time.Millisecond)
	if len(games) != 1 {
		t.Fatalf("應收到 1 筆,實得 %d", len(games))
	}
	if games[0].Name != "ABCDEFGH" {
		t.Errorf("名稱應截到 8 字元,實得 %q", games[0].Name)
	}
}
