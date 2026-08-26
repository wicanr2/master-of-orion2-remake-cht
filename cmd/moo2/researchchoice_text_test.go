package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestResearchChoicePlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"research.choice.title",
		"research.choice.topic_summary",
		"research.choice.transition.summary",
		"research.choice.transition.galaxy",
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("研究應用選擇缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		if got := strings.Count(uiText(lang, "research.choice.topic_summary"), "%s"); got != 1 {
			t.Errorf("research.choice.topic_summary (%v) 有 %d 個 %%s，預期 1", lang, got)
		}
	}
}

func TestResearchChoiceTextFitsSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	s := &researchChoiceScreen{}
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		checkClippedTextFits(t, fnt, researchChoiceTitleTextRect(), uiText(lang, "research.choice.title"), 18)
		summary := fmt.Sprintf(uiText(lang, "research.choice.topic_summary"),
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ超長研究主題名稱")
		checkClippedTextFits(t, fnt, researchChoiceTopicTextRect(), summary, 12)
		for i := 0; i < 4; i++ {
			r := s.rowTextRect(i)
			checkClippedTextFits(t, fnt, r, "ABCDEFGHIJKLMNOPQRSTUVWXYZ超長科技應用名稱", 16)
			x, y, w, h := s.rowRect(i)
			if 2*r.x+r.w != 2*x+w || 2*r.y+r.h != 2*y+h {
				t.Errorf("第 %d 列文字框與點擊框中心不同", i)
			}
		}
	}
}

func TestResearchChoiceSourceHasNoEmbeddedPlayerSentencesOrDirectDraw(t *testing.T) {
	raw, err := os.ReadFile("researchchoice.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, forbidden := range []string{".tr(", ".fnt.Draw(", ".fnt.DrawCentered("} {
		if strings.Contains(src, forbidden) {
			t.Errorf("researchchoice.go 不得出現 %s", forbidden)
		}
	}
	for _, value := range []string{
		"CHOOSE RESEARCH APPLICATION", "選擇研究應用",
		"the selected technology is granted at breakthrough", "突破後取得所選科技",
	} {
		if strings.Contains(src, `"`+value) {
			t.Errorf("researchchoice.go 仍內嵌玩家文案 %q", value)
		}
	}
}
