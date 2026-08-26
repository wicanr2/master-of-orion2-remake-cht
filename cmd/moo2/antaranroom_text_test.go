package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestAntaranRoomPlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"antaran.room.title",
		"antaran.room.subtitle",
		"antaran.room.defense",
		"antaran.room.player_power",
		"antaran.room.odds.low",
		"antaran.room.odds.ready",
		"antaran.room.button.assault",
		"antaran.room.button.retreat",
		"antaran.room.block.prefix",
		"antaran.room.block.game_over",
		"antaran.room.block.events_disabled",
		"antaran.room.block.no_portal",
		"antaran.room.block.no_fleet",
		"antaran.room.transition.fleet",
		"antaran.room.transition.battle",
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("王座廳缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestAntaranRoomFormatsAndTextFitSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		defense := fmt.Sprintf(uiText(lang, "antaran.room.defense"), 99, 999999)
		player := fmt.Sprintf(uiText(lang, "antaran.room.player_power"), 999999)
		for _, value := range []string{defense, player} {
			if strings.Contains(value, "%!") {
				t.Errorf("王座廳格式模板參數不相容：%q", value)
			}
		}
		checkClippedTextFits(t, fnt, antaranRoomTitleTextRect(), uiText(lang, "antaran.room.title"), 22)
		checkClippedTextFits(t, fnt, antaranRoomSubtitleTextRect(), uiText(lang, "antaran.room.subtitle"), 14)
		checkClippedTextFits(t, fnt, antaranRoomDefenseTextRect(), defense, 14)
		checkClippedTextFits(t, fnt, antaranRoomPlayerPowerTextRect(), player, 14)
		checkClippedTextFits(t, fnt, antaranRoomOddsTextRect(), uiText(lang, "antaran.room.odds.low"), 12)
		checkClippedTextFits(t, fnt, antaranRoomOddsTextRect(), uiText(lang, "antaran.room.odds.ready"), 12)
		checkClippedTextFits(t, fnt, antaranRoomButtonTextRect(96, 396, 190, 44),
			uiText(lang, "antaran.room.button.assault"), 16)
		checkClippedTextFits(t, fnt, antaranRoomButtonTextRect(354, 396, 190, 44),
			uiText(lang, "antaran.room.button.retreat"), 16)
		for _, button := range []struct {
			rect textSafeRect
			key  string
		}{
			{antaranRoomButtonTextRect(96, 396, 190, 44), "antaran.room.button.assault"},
			{antaranRoomButtonTextRect(354, 396, 190, 44), "antaran.room.button.retreat"},
		} {
			label := uiText(lang, button.key)
			if got := button.rect.clipped(fnt, label, 16); got != label {
				t.Errorf("王座廳按鈕標籤不得以省略號代替完整操作：%q → %q", label, got)
			}
		}
		for _, reason := range []string{
			"antaran.room.block.game_over",
			"antaran.room.block.events_disabled",
			"antaran.room.block.no_portal",
			"antaran.room.block.no_fleet",
		} {
			checkClippedTextFits(t, fnt, antaranRoomBlockTextRect(),
				uiText(lang, "antaran.room.block.prefix")+uiText(lang, reason), 12)
		}
	}
}

func TestAntaranRoomButtonTextCentersMatchHitRects(t *testing.T) {
	for _, tc := range []struct{ x, y, w, h int }{{96, 396, 190, 44}, {354, 396, 190, 44}} {
		r := antaranRoomButtonTextRect(tc.x, tc.y, tc.w, tc.h)
		if 2*r.x+r.w != 2*tc.x+tc.w || 2*r.y+r.h != 2*tc.y+tc.h {
			t.Errorf("王座廳按鈕文字框與熱區中心不一致：%+v", r)
		}
	}
}

func TestAntaranRoomSourcesContainNoEmbeddedPlayerSentences(t *testing.T) {
	for _, path := range []string{"antaranroom.go", "../../internal/shell/antaran_victory.go"} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		src := string(raw)
		if path == "antaranroom.go" && strings.Contains(src, ".tr(") {
			t.Fatal("antaranroom.go 不得再以 tr 內嵌雙語玩家文案")
		}
		for _, value := range []string{
			"THE ANTARAN THRONE ROOM", "安塔蘭王座廳", "Cannot launch:",
			"無法發動：", "本局關閉了安塔蘭攻擊", "LAUNCH THE FINAL ASSAULT",
		} {
			if strings.Contains(src, `"`+value) {
				t.Errorf("%s 仍內嵌王座廳玩家文案 %q", path, value)
			}
		}
	}
}
