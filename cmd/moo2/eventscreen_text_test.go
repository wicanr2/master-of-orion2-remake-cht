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

func TestEventScreenPlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"event.transition.summary", "event.header.gnn", "event.header.survey",
		"event.tag.alert", "event.tag.good_news", "event.tag.discovery",
		"event.title.format", "event.button.continue", "event.discovery.bc",
		"event.discovery.splinter.success", "event.discovery.splinter.failed",
		"event.discovery.leader.success", "event.discovery.leader.full",
		"event.discovery.tech.success", "event.discovery.tech.none", "event.discovery.tech.separator",
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("事件快報缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestDiscoveryRuleSourceHasNoEmbeddedPlayerSentences(t *testing.T) {
	raw, err := os.ReadFile("../../internal/shell/discovery.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, forbidden := range []string{"勘查小隊", "失散同胞", "受困的傭兵", "遠古文物解析", "survey team found", "stranded mercenary", "ancient artifacts in"} {
		if strings.Contains(strings.ToLower(src), strings.ToLower(forbidden)) {
			t.Errorf("discovery.go 仍內嵌玩家句子 %q", forbidden)
		}
	}
}

func TestEventScreenExternalTextFitsSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		checkClippedTextFits(t, fnt, eventHeaderTextRect(), uiText(lang, "event.header.gnn"), 13)
		title := fmt.Sprintf(uiText(lang, "event.title.format"), uiText(lang, "event.tag.good_news"),
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		for _, original := range []bool{false, true} {
			checkClippedTextFits(t, fnt, eventTitleTextRect(original), title, 15)
			for _, line := range eventBodyTextRect(original).lines(fnt,
				"這是一段刻意很長的事件敘述，用來驗證多行中文字不會穿過報告底板、插圖或繼續按鈕。", 13) {
				checkClippedTextFits(t, fnt, textSafeRect{w: eventBodyTextRect(original).w, h: 20}, line, 13)
			}
		}
		checkClippedTextFits(t, fnt, eventButtonTextRect(), uiText(lang, "event.button.continue"), 13)
	}
	r := eventButtonTextRect()
	if 2*r.x+r.w != 2*270+100 || 2*r.y+r.h != 2*372+24 {
		t.Fatal("事件按鈕文字框與可見按鈕中心不同")
	}
}

func TestEventArtworkAssetIDBounds(t *testing.T) {
	for _, id := range []int{0, 17, 35} {
		assetID, ok := eventArtworkAssetID(id, 38)
		if !ok || assetID != id+2 {
			t.Fatalf("合法事件 %d 未映射到 EVENTS.LBX#%d", id, id+2)
		}
	}
	for _, tc := range []struct{ id, count int }{{-1, 38}, {36, 38}, {35, 37}, {0, 1}} {
		if _, ok := eventArtworkAssetID(tc.id, tc.count); ok {
			t.Errorf("非法事件資產未被拒絕：id=%d count=%d", tc.id, tc.count)
		}
	}
}

func TestEventScreenAcceptsClickOutsideVisibleButton(t *testing.T) {
	called := ""
	s := &overlayScreen{
		hits: []hitRegion{{270, 372, 100, 24, "ok"}}, clickAnywhereAction: "ok",
		onAction: func(action string) *origTransition { called = action; return &origTransition{} },
	}
	if tr := s.update(shell.InputState{MouseX: 20, MouseY: 20, ClickReleased: true}); tr == nil || called != "ok" {
		t.Fatal("事件畫面應依原版接受可見按鈕外的全畫面點擊")
	}
}

func TestEventArtworkRealArchive(t *testing.T) {
	dir := os.Getenv("MOO2_EVENTS_TEST")
	if dir == "" {
		t.Skip("未設 MOO2_EVENTS_TEST，跳過正版 EVENTS.LBX 測試")
	}
	res, err := assets.NewResolver(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []int{0, 35} {
		im := loadEventArtwork(res, id)
		if im == nil || im.Bounds().Dx() != 157 || im.Bounds().Dy() != 125 {
			t.Fatalf("EVENTS.LBX 事件 %d 未解成 157×125", id)
		}
	}
	if im := loadEventArtwork(res, -1); im != nil {
		t.Fatal("畫廊固定 EventID=-1 不得誤索引原版插圖")
	}
}

func TestEventScreenSourceHasNoEmbeddedPlayerSentencesOrDirectDraw(t *testing.T) {
	raw, err := os.ReadFile("eventscreen.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, forbidden := range []string{".tr(", ".fnt.Draw(", ".fnt.DrawCentered("} {
		if strings.Contains(src, forbidden) {
			t.Errorf("eventscreen.go 不得出現 %s", forbidden)
		}
	}
	for _, value := range []string{"GOOD NEWS", "喜訊", "IMPERIAL SURVEY REPORT", "帝國勘查回報", "CONTINUE", "繼續"} {
		if strings.Contains(src, `"`+value+`"`) {
			t.Errorf("eventscreen.go 仍內嵌玩家文案 %q", value)
		}
	}
}
