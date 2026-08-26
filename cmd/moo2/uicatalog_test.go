package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
)

func TestNetInfoPlayerTextComesFromExternalJSON(t *testing.T) {
	cases := []struct {
		key, zh, en string
	}{
		{"netinfo.caption.waiting_for_joiners", "等待其他玩家加入", "Waiting for players to join"},
		{"netinfo.caption.joining", "加入對局中", "Joining game"},
		{"netinfo.caption.wait_race_info", "等待種族資料", "Waiting for race info"},
		{"netinfo.caption.initializing", "初始化連線", "Initializing network"},
		{"netinfo.caption.sending_data", "傳送資料", "Sending data"},
		{"netinfo.caption.generating_map", "產生星圖", "Generating map"},
		{"netinfo.caption.getting_data", "接收資料", "Getting data"},
		{"netinfo.title.waiting_for_joiners", "加入網路遊戲設定", "JOIN NETWORK GAME SETUP"},
		{"netinfo.title.joining", "等待加入遊戲", "WAITING TO JOIN GAME"},
		{"netinfo.title.wait_race_info", "接收種族設定", "RECEIVING RACE SETUPS"},
		{"netinfo.title.initializing", "初始化網路", "INITIALIZING NETWORK"},
		{"netinfo.title.sending_data", "傳送遊戲資料", "SENDING GAME DATA"},
		{"netinfo.title.generating_map", "產生星圖", "GENERATING MAP"},
		{"netinfo.title.getting_data", "接收遊戲資料", "RECEIVING GAME DATA"},
		{"netinfo.label.status", "狀態", "STATUS"},
		{"netinfo.button.start", "開始連線對局", "START NET GAME"},
	}
	for _, tc := range cases {
		if got := uiText(i18n.Traditional, tc.key); got != tc.zh {
			t.Errorf("uiText(zh,%q)=%q，預期 %q", tc.key, got, tc.zh)
		}
		if got := uiText(i18n.English, tc.key); got != tc.en {
			t.Errorf("uiText(en,%q)=%q，預期 %q", tc.key, got, tc.en)
		}
	}
}

func TestNetInfoSourceHasNoEmbeddedPlayerSentences(t *testing.T) {
	raw, err := os.ReadFile("netinfo.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	if strings.Contains(src, ".tr(") {
		t.Fatal("netinfo.go 不得再用 tr 內嵌中英文玩家文案")
	}
	for _, text := range []string{
		"等待其他玩家加入", "Waiting for players to join",
		"開始連線對局", "START NET GAME",
	} {
		if strings.Contains(src, `"`+text+`"`) {
			t.Errorf("netinfo.go 仍內嵌玩家文案 %q", text)
		}
	}
}
