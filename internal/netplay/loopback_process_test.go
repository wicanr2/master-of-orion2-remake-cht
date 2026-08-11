package netplay

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// loopback_process_test.go：用兩個實際 test process 抽樣走過 TCP 大廳與
// 對局訊息幫浦。它不是完整遊戲測試，但能抓到「大廳收得到人、開局後 socket
// 卻沒有讀寫」這類同一程序單元測試看不出的接線錯誤。

const (
	loopbackRole  = "MOO2_NETPLAY_ROLE"
	loopbackAddr  = "MOO2_NETPLAY_ADDR"
	loopbackReady = "MOO2_NETPLAY_READY"
)

func TestNetplayProcessHelper(t *testing.T) {
	role := os.Getenv(loopbackRole)
	if role == "" {
		return
	}
	addr := os.Getenv(loopbackAddr)
	if addr == "" {
		t.Fatal("helper 缺少 TCP 位址")
	}
	switch role {
	case "host":
		runLoopbackHost(t, addr, os.Getenv(loopbackReady))
	case "client":
		runLoopbackClient(t, addr)
	default:
		t.Fatalf("未知 helper 角色 %q", role)
	}
}

func TestTwoProcessTCPHandshakeAndTurnMessage(t *testing.T) {
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := probe.Addr().String()
	_ = probe.Close()

	dir := t.TempDir()
	ready := filepath.Join(dir, "host.ready")
	host := startLoopbackHelper(t, "host", addr, ready)
	waitForPath(t, ready, 5*time.Second)
	client := startLoopbackHelper(t, "client", addr, "")

	waitLoopbackProcess(t, host, 8*time.Second)
	waitLoopbackProcess(t, client, 8*time.Second)
}

func startLoopbackHelper(t *testing.T, role, addr, ready string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run", "^TestNetplayProcessHelper$", "-test.v")
	cmd.Env = append(os.Environ(), loopbackRole+"="+role, loopbackAddr+"="+addr, loopbackReady+"="+ready)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatalf("啟動 %s helper：%v", role, err)
	}
	return cmd
}

func waitForPath(t *testing.T, path string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("TCP 主機 helper 沒有在期限內準備好")
}

func waitLoopbackProcess(t *testing.T, cmd *exec.Cmd, timeout time.Duration) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("loopback helper 失敗：%v", err)
		}
	case <-time.After(timeout):
		_ = cmd.Process.Kill()
		t.Fatalf("loopback helper 超時：%s", cmd.Args[0])
	}
}

func runLoopbackHost(t *testing.T, addr, readyPath string) {
	lobby, err := Host(addr, "host", 42)
	if err != nil {
		t.Fatal(err)
	}
	defer lobby.Close()
	if readyPath == "" {
		t.Fatal("host helper 缺少 ready 路徑")
	}
	if err := os.WriteFile(readyPath, []byte(lobby.Addr()), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := lobby.AcceptOne(5 * time.Second); err != nil {
		t.Fatal(err)
	}
	sess := NewSession(0, true, lobby.Connections())
	defer sess.Close()
	if err := sess.Send(Message{Kind: KindTurnDone, Turn: 3, Commands: []Command{{Name: "probe"}}}); err != nil {
		t.Fatal(err)
	}
	waitForSessionMessage(t, sess, func(m Message) bool {
		return m.Kind == KindTurnReady && m.Turn == 3 && m.Player == 1
	})
}

func runLoopbackClient(t *testing.T, addr string) {
	deadline := time.Now().Add(5 * time.Second)
	var conn net.Conn
	var err error
	for time.Now().Before(deadline) {
		conn, _, _, err = Join(addr, "client", 300*time.Millisecond)
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil || conn == nil {
		t.Fatalf("client helper 無法加入：%v", err)
	}
	sess := NewSession(1, false, map[int]net.Conn{0: conn})
	defer sess.Close()
	waitForSessionMessage(t, sess, func(m Message) bool {
		return m.Kind == KindTurnDone && m.Turn == 3 && len(m.Commands) == 1
	})
	if err := sess.Send(Message{Kind: KindTurnReady, Turn: 3, StateHash: "loopback-ok"}); err != nil {
		t.Fatal(fmt.Errorf("client helper 回傳 ready：%w", err))
	}
}

func waitForSessionMessage(t *testing.T, sess *Session, match func(Message) bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		for _, m := range sess.Poll() {
			if match(m) {
				return
			}
		}
		if err := sess.Err(); err != nil {
			t.Fatal(err)
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("等待 TCP 對局訊息逾時")
}
