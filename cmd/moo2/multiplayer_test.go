package main

import (
	"testing"
	"time"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/netplay"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func mpButtonForTest(t *testing.T, act string) mpButton {
	t.Helper()
	for _, btn := range mpButtons {
		if btn.act == act {
			return btn
		}
	}
	t.Fatalf("找不到多人按鈕 %q", act)
	return mpButton{}
}

// TCP 重製的 NETWORK / JOIN 路徑已存在時，畫面不可以仍把它們誤畫成未實作。
func TestMultiplayerTCPEntryIsEnabledAndSelected(t *testing.T) {
	s := &multiplayerScreen{b: &sceneBuilder{}, mode: mpHotseat, humans: 2}
	if !s.implemented("network") || !s.implemented("join") {
		t.Fatal("TCP NETWORK 與 JOIN GAME 必須標為已實作")
	}
	if s.frameFor("network") == mpFrameDisabled {
		t.Fatal("NETWORK 不可畫成停用")
	}
	btn := mpButtonForTest(t, "network")
	x, y, w, h := s.btnRect(btn)
	s.update(shell.InputState{MouseX: x + w/2, MouseY: y + h/2, ClickReleased: true})
	if s.mode != mpNetwork {
		t.Fatalf("點 NETWORK 應切進 TCP 模式，得到 %d", s.mode)
	}
	if got := s.frameFor("network"); got != mpFrameSelected {
		t.Fatalf("NETWORK 選中後應使用 selected 幀，得到 %d", got)
	}
	if got := s.frameFor("join"); got == mpFrameDisabled {
		t.Fatal("NETWORK 模式的 JOIN GAME 不可畫成停用")
	}
}

// 已淘汰的數據機／序列線／TEN 維持停用，也不應在畫成灰色後仍偷偷處理點擊。
func TestMultiplayerLegacyTransportIsDisabledAndInert(t *testing.T) {
	s := &multiplayerScreen{b: &sceneBuilder{}, mode: mpHotseat, humans: 2}
	for _, act := range []string{"modem", "nullmodem", "comm", "ten"} {
		if s.implemented(act) {
			t.Errorf("%s 應維持停用", act)
		}
	}
	btn := mpButtonForTest(t, "modem")
	x, y, w, h := s.btnRect(btn)
	s.update(shell.InputState{MouseX: x + w/2, MouseY: y + h/2, ClickReleased: true})
	if s.mode != mpHotseat || s.msg != "" {
		t.Fatalf("停用的 MODEM 不該處理點擊: mode=%d msg=%q", s.mode, s.msg)
	}
}

// 完整包沒有 MULTIGM.LBX 時仍須維持原版已知面板尺寸與熱區，不能把可用的
// TCP 入口塌到左上角。
func TestMultiplayerMissingAssetsUsesCenteredFallbackGeometry(t *testing.T) {
	s := newMultiplayerScreen(&sceneBuilder{res: emptyResolverForMultiplayerTest(t), lang: i18n.Traditional})
	if s.panW != 482 || s.panH != 335 || s.panX != 79 || s.panY != 72 {
		t.Fatalf("缺資產備援面板幾何錯誤：%dx%d @ (%d,%d)", s.panW, s.panH, s.panX, s.panY)
	}
	network := mpButtonForTest(t, "network")
	x, y, w, h := s.btnRect(network)
	if x != 138 || y != 163 || w != 154 || h != 26 {
		t.Fatalf("缺資產備援 NETWORK 熱區錯誤：(%d,%d,%d,%d)", x, y, w, h)
	}
	s.update(shell.InputState{MouseX: x + w/2, MouseY: y + h/2, ClickReleased: true})
	if s.mode != mpNetwork {
		t.Fatal("缺資產備援仍須能點選 TCP NETWORK")
	}
}

func emptyResolverForMultiplayerTest(t *testing.T) *assets.Resolver {
	t.Helper()
	r, err := assets.NewResolver()
	if err != nil {
		t.Fatalf("建立空資產解析器：%v", err)
	}
	return r
}

func waitForMultiplayerTest(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("等待%s逾時", what)
}

// 從多人設定的 NETWORK 鈕開始，穿過 START NEW GAME 的名稱彈窗，再由客戶端走實際
// TCP 加入邏輯；接著驗共同快照與第一個空指令回合的 turn_done → turn_ready 鎖步。
// 這條不用原版美術，故可在隔離的 loopback 容器穩定重現，但沒有跳過 UI callback 或
// 直接偽造網路訊息。
func TestMultiplayerTCPHostJoinAndFirstTurn(t *testing.T) {
	host := &sceneBuilder{
		res:     emptyResolverForMultiplayerTest(t),
		lang:    i18n.Traditional,
		session: shell.NewDemoSession(),
	}
	defer host.closeNetwork()

	menu := newMultiplayerScreen(host)
	network := mpButtonForTest(t, "network")
	nx, ny, nw, nh := menu.btnRect(network)
	menu.update(shell.InputState{MouseX: nx + nw/2, MouseY: ny + nh/2, ClickReleased: true})
	if menu.mode != mpNetwork {
		t.Fatalf("NETWORK 鈕沒有切換 TCP 模式，得到 %d", menu.mode)
	}
	start := mpButtonForTest(t, "start")
	sx, sy, sw, sh := menu.btnRect(start)
	startTransition := menu.update(shell.InputState{MouseX: sx + sw/2, MouseY: sy + sh/2, ClickReleased: true})
	if startTransition == nil {
		t.Fatal("START NEW GAME 應開對局名稱彈窗")
	}
	nameBox, ok := startTransition.next.(*inputBoxScreen)
	if !ok {
		t.Fatalf("START NEW GAME 應先進 inputBoxScreen，得到 %T", startTransition.next)
	}
	nameBox.text = []rune("LOOPBACK")
	hostTransition := nameBox.accept()
	if hostTransition == nil || host.netLobby == nil {
		t.Fatal("確認對局名稱後應建立 TCP 大廳")
	}

	client := &sceneBuilder{res: emptyResolverForMultiplayerTest(t), lang: i18n.Traditional}
	defer client.closeNetwork()
	clientScreen, err := client.joinNetGame(netplay.Game{Name: "LOOPBACK", Addr: netLobbyDialAddr})
	if err != nil {
		t.Fatalf("客戶端加入主機大廳：%v", err)
	}
	clientLobby, ok := clientScreen.(*chooseNetPlayersScreen)
	if !ok {
		t.Fatalf("加入後應進玩家名冊，得到 %T", clientScreen)
	}
	waitForMultiplayerTest(t, "主機名冊收錄客戶端", func() bool {
		return len(host.netLobby.Roster().Players) == 2
	})
	if client.netMe != 1 {
		t.Fatalf("客戶端應取得席位 1，得到 %d", client.netMe)
	}

	// 主機完成共同開局後，客戶端在名冊畫面的正常 update 輪詢 GameStart。
	host.finishNetworkHostSetup()
	waitForMultiplayerTest(t, "客戶端套用共同開局", func() bool {
		clientLobby.update(shell.InputState{})
		return client.session != nil && !client.networkPending
	})
	if host.session.NetworkStateHash() != client.session.NetworkStateHash() {
		t.Fatal("共同開局快照後主機與客戶端狀態指紋不同")
	}
	if host.networkTurn == nil || client.networkTurn == nil {
		t.Fatal("共同開局後雙方都應進入第一回合鎖步狀態")
	}

	turn := host.session.Turn
	hostWaitTransition := host.submitNetworkTurn()
	clientWaitTransition := client.submitNetworkTurn()
	hostWait, ok := hostWaitTransition.next.(*networkWaitScreen)
	if !ok {
		t.Fatalf("主機提交回合後應進等待畫面，得到 %T", hostWaitTransition.next)
	}
	clientWait, ok := clientWaitTransition.next.(*networkWaitScreen)
	if !ok {
		t.Fatalf("客戶端提交回合後應進等待畫面，得到 %T", clientWaitTransition.next)
	}
	waitForMultiplayerTest(t, "第一回合完成 lockstep", func() bool {
		hostWait.update(shell.InputState{})
		clientWait.update(shell.InputState{})
		return host.session.Turn == turn+1 && client.session.Turn == turn+1
	})
	if host.networkError != "" || client.networkError != "" {
		t.Fatalf("第一回合不應斷線或不同步：host=%q client=%q", host.networkError, client.networkError)
	}
	if host.session.NetworkStateHash() != client.session.NetworkStateHash() {
		t.Fatal("第一回合結算後主機與客戶端狀態指紋不同")
	}
}
