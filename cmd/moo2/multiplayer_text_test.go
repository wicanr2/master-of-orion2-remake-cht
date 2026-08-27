package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestMultiplayerSetupPlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"multiplayer.setup.title",
		"multiplayer.setup.button.hotseat_count",
		"multiplayer.setup.note.hotseat",
		"multiplayer.setup.note.network",
		"multiplayer.setup.message.legacy_transport",
		"multiplayer.setup.message.choose_supported",
		"multiplayer.setup.message.no_saves",
		"multiplayer.setup.message.load_failed",
		"multiplayer.setup.message.comm_legacy",
		"multiplayer.setup.message.ten_closed",
		"multiplayer.setup.error.host",
		"multiplayer.setup.error.join",
		"multiplayer.setup.transition.new_game",
		"multiplayer.setup.transition.main_menu",
	}
	for _, button := range mpButtons {
		keys = append(keys, button.textKey)
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("多人主設定缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestMultiplayerSetupTextFitsSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	s := &multiplayerScreen{panX: 79, panY: 72, panW: 482, panH: 335}
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		checkFull := func(name string, rect textSafeRect, value string, size float64) {
			t.Helper()
			checkClippedTextFits(t, fnt, rect, value, size)
			if got := rect.clipped(fnt, value, size); got != value {
				t.Errorf("%s 不得被裁切：%q → %q", name, value, got)
			}
		}
		checkFull("標題", s.titleTextRect(), uiText(lang, "multiplayer.setup.title"), 16)
		for _, button := range mpButtons {
			value := uiText(lang, button.textKey)
			if button.act == "hotseat" {
				value = fmt.Sprintf(uiText(lang, "multiplayer.setup.button.hotseat_count"), maxHotseatSeats)
			}
			checkFull(button.act, s.buttonTextRect(button), value, 13)
		}
		for _, item := range []struct {
			name  string
			key   string
			value string
		}{
			{"熱座說明", "multiplayer.setup.note.hotseat", fmt.Sprintf(uiText(lang, "multiplayer.setup.note.hotseat"), maxHotseatSeats)},
			{"網路說明", "multiplayer.setup.note.network", uiText(lang, "multiplayer.setup.note.network")},
		} {
			checkFull(item.name, s.noteTextRect(), item.value, 10)
		}
		for _, key := range []string{
			"multiplayer.setup.message.legacy_transport",
			"multiplayer.setup.message.choose_supported",
			"multiplayer.setup.message.no_saves",
			"multiplayer.setup.message.load_failed",
			"multiplayer.setup.message.comm_legacy",
			"multiplayer.setup.message.ten_closed",
		} {
			checkFull(key, s.messageTextRect(), uiText(lang, key), 10)
		}
		for _, key := range []string{"multiplayer.setup.error.host", "multiplayer.setup.error.join"} {
			value := fmt.Sprintf(uiText(lang, key), "connection refused")
			if strings.Contains(value, "%!") {
				t.Errorf("格式參數不相容：%s = %q", key, value)
			}
			checkFull(key, s.messageTextRect(), value, 10)
		}
	}
}

func TestMultiplayerSetupNotesStayInsideMainPanel(t *testing.T) {
	s := &multiplayerScreen{panX: 79, panY: 72, panW: 482, panH: 335}
	for name, r := range map[string]textSafeRect{"note": s.noteTextRect(), "message": s.messageTextRect()} {
		if r.x < s.panX || r.y < s.panY || r.x+r.w > s.panX+s.panW || r.y+r.h > s.panY+s.panH {
			t.Fatalf("%s 未包含於多人主面板：rect=%+v panel=(%d,%d,%d,%d)", name, r, s.panX, s.panY, s.panW, s.panH)
		}
	}
}

func TestMultiplayerSetupButtonTextSharesHitRectCenter(t *testing.T) {
	s := &multiplayerScreen{panX: 79, panY: 72, panW: 482, panH: 335}
	for _, button := range mpButtons {
		x, y, w, h := s.btnRect(button)
		r := s.buttonTextRect(button)
		if 2*r.x+r.w != 2*x+w || 2*r.y+r.h != 2*y+h {
			t.Errorf("%s 文字框與熱區中心不一致：text=%+v hit=(%d,%d,%d,%d)", button.act, r, x, y, w, h)
		}
	}
}

func TestMultiplayerSetupSourceHasNoEmbeddedPlayerSentences(t *testing.T) {
	raw, err := os.ReadFile("multiplayer.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	if strings.Contains(source, ".tr(") {
		t.Fatal("multiplayer.go 不得再以 tr 內嵌雙語玩家文案")
	}
	for _, value := range []string{
		"多人遊戲設定", "開始新遊戲", "TEN 連線服務", "No saved games yet.",
		"Could not host:", "TCP network:",
	} {
		if strings.Contains(source, `"`+value) {
			t.Errorf("multiplayer.go 仍內嵌玩家文案 %q", value)
		}
	}
}
