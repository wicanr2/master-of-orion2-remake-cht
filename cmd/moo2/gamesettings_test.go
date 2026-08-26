package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestGameSettingsExternalTextAndSafeRects(t *testing.T) {
	s := &gameSettingsScreen{x: (moo2ScreenW - gameSettingsW) / 2, y: (moo2ScreenH - gameSettingsH) / 2}
	for i, key := range gameSettingsKeys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("SETTINGS 缺少外部文案：%s (%v)", key, lang)
			}
		}
		x, y, w, h := s.rowRect(i)
		if x < s.x || y < s.y || x+w > s.x+gameSettingsW || y+h > s.y+gameSettingsH {
			t.Errorf("第 %d 列超出 SETTINGS 背景", i)
		}
	}
}

func TestGameSettingsToggleAndAcceptPersist(t *testing.T) {
	b := &sceneBuilder{session: shell.NewDemoSession()}
	s := &gameSettingsScreen{b: b, settings: b.session.EffectiveGameSettings()}
	before := s.settings.ShowRelocationLines
	s.toggleOption(7)
	if s.settings.ShowRelocationLines == before {
		t.Fatal("第八列應切換遷移線")
	}
	b.session.ApplyGameSettings(s.settings)
	if b.session.ShowRelocationLines != s.settings.ShowRelocationLines {
		t.Fatal("相容消費端未同步")
	}
}

func TestGameSettingsSourceHasNoEmbeddedPlayerText(t *testing.T) {
	raw, err := os.ReadFile("gamesettings.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, value := range []string{"GAME SETTINGS", "End Of Turn Summary", "遊戲設定", "接受"} {
		if strings.Contains(src, `"`+value+`"`) {
			t.Errorf("gamesettings.go 內嵌玩家文案 %q", value)
		}
	}
}
