package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/netplay"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestNetworkGamePlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"network.transition.shared_setup", "network.transition.galaxy", "network.player.fallback",
		"network.error.parse_start", "network.error.incomplete_start", "network.error.missing_session",
		"network.error.not_enough_players", "network.error.send_commands", "network.error.connection",
		"network.error.commands_before_table", "network.error.ready_before_table",
		"network.error.turn_data_missing", "network.error.replay_player",
		"network.error.fingerprint_missing", "network.error.send_fingerprint",
		"network.error.turn_state_missing", "network.error.stopped", "network.error.title",
		"network.error.return_hint",
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("正式網路回合缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		if got := strings.Count(uiText(lang, "network.error.replay_player"), "%"); got != 2 {
			t.Errorf("replay_player (%v) 應有兩個格式參數，實得 %d", lang, got)
		}
	}
}

func TestNetworkWaitErrorTextFitsOriginalPanel(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		checkClippedTextFits(t, fnt, networkWaitErrorTitleTextRect(), uiText(lang, "network.error.title"), 13)
		detail := fmt.Sprintf(uiText(lang, "network.error.connection"), "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		checkClippedTextFits(t, fnt, networkWaitErrorDetailTextRect(), detail, 11)
		checkClippedTextFits(t, fnt, networkWaitErrorHintTextRect(), uiText(lang, "network.error.return_hint"), 11)
	}
}

func TestNetworkWaitHandleChatKeepsFormalRendererLog(t *testing.T) {
	b := &sceneBuilder{
		lang:          i18n.Traditional,
		networkTurn:   &networkTurnState{table: netplay.NewTable(2, 7)},
		networkRoster: netplay.Roster{Players: []netplay.Player{{ID: 0, Name: "A"}, {ID: 1, Name: "B"}}},
	}
	s := &networkWaitScreen{b: b, visual: &netNextTurnScreen{b: b, names: []string{"A", "B"}}}
	s.handle(netplay.Message{Kind: netplay.KindChat, Player: 1, Text: "hello"})
	lines := s.visual.chat.Lines()
	if len(lines) != 1 || lines[0].Speaker != 1 || lines[0].Text != "hello" {
		t.Fatalf("正式等待畫面應保留聊天 speaker／text，實得 %+v", lines)
	}
	if s.b.networkTurn.table == nil || s.b.networkTurn.table.Turn() != 7 {
		t.Fatal("處理聊天不得消耗或替換鎖步回合表")
	}
}

func TestNetworkGameSourceHasNoEmbeddedPlayerTextOrDirectDraw(t *testing.T) {
	raw, err := os.ReadFile("networkgame.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, forbidden := range []string{".tr(", ".fnt.Draw(", ".fnt.DrawCentered("} {
		if strings.Contains(src, forbidden) {
			t.Errorf("networkgame.go 不得出現 %s", forbidden)
		}
	}
	for _, value := range []string{"網路共同開局設定", "Network match stopped", "點擊返回多人設定"} {
		if strings.Contains(src, `"`+value+`"`) {
			t.Errorf("networkgame.go 仍內嵌玩家文案 %q", value)
		}
	}
}
