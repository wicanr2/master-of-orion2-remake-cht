package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestNetNextTurnPlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"netwait.title", "netwait.status.no_game", "netwait.status.waiting", "netwait.status.ready",
		"netwait.player.you_suffix", "netwait.turn", "netwait.fingerprint", "netwait.desync.title",
		"netwait.chat.gnn_prefix", "netwait.chat.player_prefix", "netwait.chat.unknown_player",
		"netwait.chat.caret", "netwait.demo.player", "netwait.demo.gnn",
		"netwait.demo.other_player", "netwait.demo.local_player",
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("Net_Next_Turn 缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
	for _, key := range []string{"netwait.turn", "netwait.fingerprint", "netwait.chat.player_prefix", "netwait.chat.unknown_player"} {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := strings.Count(uiText(lang, key), "%"); got != 1 {
				t.Errorf("%s (%v) 應有一個格式參數，實得 %d", key, lang, got)
			}
		}
	}
}

func TestNetNextTurnExternalTextFitsSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		checkClippedTextFits(t, fnt, nntBannerTitleTextRect(), uiText(lang, "netwait.title"), 14)
		checkClippedTextFits(t, fnt, nntPlayerNameTextRect(0), "ABCDEFGHIJKLMNOPQRSTUVWXYZ（你）", 13)
		checkClippedTextFits(t, fnt, nntPlayerStatusTextRect(0), uiText(lang, "netwait.status.waiting"), 13)
		checkClippedTextFits(t, fnt, nntTurnTextRect(), fmt.Sprintf(uiText(lang, "netwait.turn"), 9999), 14)
		checkClippedTextFits(t, fnt, nntFingerprintTextRect(), fmt.Sprintf(uiText(lang, "netwait.fingerprint"), "01234567"), 12)
		checkClippedTextFits(t, fnt, nntDesyncTitleTextRect(), uiText(lang, "netwait.desync.title"), 13)
		checkClippedTextFits(t, fnt, nntDesyncDetailTextRect(), "player 8: 0123456789abcdef", 11)
		checkClippedTextFits(t, fnt, nntInputTextRect(), "（ABCDEFGHIJKLMNOPQRSTUVWXYZ）  1234567890＿", 11)
	}
}

func TestNetNextTurnSourceHasNoEmbeddedPlayerText(t *testing.T) {
	raw, err := os.ReadFile("netnextturn.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	if strings.Contains(src, ".tr(") {
		t.Fatal("netnextturn.go 不得再用 tr 內嵌中英文玩家文案")
	}
	for _, value := range []string{"等待其他玩家", "waiting…", "狀態指紋：", "One more turn."} {
		if strings.Contains(src, `"`+value+`"`) {
			t.Errorf("netnextturn.go 仍內嵌玩家文案 %q", value)
		}
	}
}
