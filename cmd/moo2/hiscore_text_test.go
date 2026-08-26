package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestHiScorePlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"hiscore.transition.summary", "hiscore.header.hall_of_fame", "hiscore.title.won", "hiscore.title.lost",
		"hiscore.reason.extermination", "hiscore.reason.council", "hiscore.reason.antaran", "hiscore.reason.game_over",
		"hiscore.summary.won", "hiscore.summary.lost", "hiscore.button.continue",
	}
	for _, line := range shell.NewDemoSession().ScoreLines() {
		keys = append(keys, line.TextKey)
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("最終得分缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestHiScoreExternalTextFitsSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		checkClippedTextFits(t, fnt, hiScoreTitleTextRect(), uiText(lang, "hiscore.title.lost"), 22)
		summary := fmt.Sprintf(uiText(lang, "hiscore.summary.lost"), 9999, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		checkClippedTextFits(t, fnt, hiScoreSummaryTextRect(), summary, 13)
		for i, line := range shell.NewDemoSession().ScoreLines() {
			y, ok := hiScoreRowY(i, i == 8)
			if !ok {
				t.Fatalf("第 %d 列沒有安全版面", i)
			}
			size := 14.0
			checkClippedTextFits(t, fnt, hiScoreLabelTextRect(int(y), int(size)+4), uiText(lang, line.TextKey), size)
			checkClippedTextFits(t, fnt, hiScoreValueTextRect(int(y), int(size)+4), "2147483647", size)
		}
		checkClippedTextFits(t, fnt, textSafeRect{x: hsContinueX, y: hsContinueY, w: hsContinueW, h: hsContinueH, insetX: 5}, uiText(lang, "hiscore.button.continue"), 12)
	}
}

func TestHiScoreAcceptsClickAnywhere(t *testing.T) {
	called := ""
	s := &overlayScreen{clickAnywhereAction: "ok", onAction: func(action string) *origTransition {
		called = action
		return &origTransition{}
	}}
	if tr := s.update(shell.InputState{MouseX: 20, MouseY: 20, ClickReleased: true}); tr == nil || called != "ok" {
		t.Fatal("最終得分應依原版接受全畫面點擊")
	}
}

func TestHiScoreRealBackground(t *testing.T) {
	dir := os.Getenv("MOO2_SCORE_TEST")
	if dir == "" {
		t.Skip("未設 MOO2_SCORE_TEST，跳過正版 SCORE.LBX 測試")
	}
	res, err := assets.NewResolver(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !hiScoreBackgroundAvailable(res) {
		t.Fatal("SCORE.LBX#0 未通過 640×480／調色盤驗證")
	}
}

func TestHiScoreSourceHasNoEmbeddedPlayerTextOrDirectDraw(t *testing.T) {
	raw, err := os.ReadFile("hiscore.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, forbidden := range []string{".tr(", ".fnt.Draw(", ".fnt.DrawCentered("} {
		if strings.Contains(src, forbidden) {
			t.Errorf("hiscore.go 不得出現 %s", forbidden)
		}
	}
	for _, value := range []string{"銀河霸主", "帝國殞落", "殲滅所有對手", "總分", "CONTINUE"} {
		if strings.Contains(src, `"`+value+`"`) {
			t.Errorf("hiscore.go 仍內嵌玩家文案 %q", value)
		}
	}
}
