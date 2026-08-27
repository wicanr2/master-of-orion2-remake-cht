package main

import (
	"image/color"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestTurnSummaryCatalogAndFormatContracts(t *testing.T) {
	wants := map[string]int{
		"turnsummary.report.stardate": 1, "turnsummary.report.industry_research": 2,
		"turnsummary.report.food_tax": 2, "turnsummary.report.treasury": 2,
		"turnsummary.report.fiscal_crisis": 4, "turnsummary.report.research_complete": 0,
		"turnsummary.report.event": 1, "turnsummary.build.completed": 2,
		"turnsummary.build.refit_completed": 2, "turnsummary.build.refit_cancelled": 2,
		"turnsummary.build.artificial_planet": 1, "turnsummary.build.unknown": 0,
		"turnsummary.transition.galaxy": 0,
	}
	for key, want := range wants {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			got := uiText(lang, key)
			if got == "" || got == key {
				t.Errorf("回合摘要缺少外部文案：%s (%v)", key, lang)
			}
			if count := strings.Count(got, "%"); count != want {
				t.Errorf("%s (%v) 有 %d 個格式參數，預期 %d", key, lang, count, want)
			}
		}
	}
}

func TestTurnSummaryBuildNoticesAreTypedAndLocalized(t *testing.T) {
	notices := []shell.BuildNotice{
		{Kind: shell.BuildNoticeCompleted, ColonyIndex: 1, Name: "研究實驗室"},
		{Kind: shell.BuildNoticeRefitCompleted, ColonyIndex: 0, Name: "先驅號"},
		{Kind: shell.BuildNoticeRefitCancelled, ColonyIndex: 2, Name: "守望號"},
		{Kind: shell.BuildNoticeArtificialPlanet, ColonyIndex: 0, Name: "新世界"},
	}
	for _, notice := range notices {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			got := turnSummaryBuildNoticeText(lang, notice)
			if got == "" || strings.Contains(got, "%!") {
				t.Errorf("typed 建造通知格式化失敗：%+v (%v) => %q", notice, lang, got)
			}
		}
	}
	if got := turnSummaryBuildNoticeText(i18n.English, notices[0]); !strings.Contains(got, "Research Laboratory") {
		t.Fatalf("英文完工通知未翻譯建築名稱：%q", got)
	}
}

func TestTurnSummaryWorstCaseMessagesStayAboveCloseButton(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	messages := make([]turnSummaryMessage, 0, 20)
	for i := 0; i < 20; i++ {
		messages = append(messages, turnSummaryMessage{
			text: "◆ 財政危機與安塔蘭突襲同時發生，這是一條刻意加長的摘要訊息",
			size: 14, col: color.RGBA{240, 130, 100, 255},
		})
	}
	extras := turnSummaryDynamicExtras(fnt, messages)
	if len(extras) != turnSummaryDynamicRect.maxLines() {
		t.Fatalf("最壞情境應收束到 %d 列，got %d", turnSummaryDynamicRect.maxLines(), len(extras))
	}
	for _, extra := range extras {
		_, h := fnt.Measure(extra.text, extra.size)
		if extra.x < float64(turnSummaryDynamicRect.contentX()) ||
			extra.x+extra.maxW > float64(turnSummaryDynamicRect.x+turnSummaryDynamicRect.w) ||
			extra.y < float64(turnSummaryDynamicRect.y) ||
			extra.y+h > float64(turnSummaryDynamicRect.y+turnSummaryDynamicRect.h) {
			t.Errorf("回合摘要列越出安全框：%+v，字高 %.1f", extra, h)
		}
		if extra.y+h >= 324 {
			t.Errorf("回合摘要列侵入 CLOSE 按鈕：%+v", extra)
		}
	}
	if !strings.HasSuffix(extras[len(extras)-1].text, "…") {
		t.Fatalf("溢位末列應有省略號：%q", extras[len(extras)-1].text)
	}
}

func TestTurnSummarySourceHasNoEmbeddedFixedText(t *testing.T) {
	raw, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	start := strings.Index(src, "func (b *sceneBuilder) turnSummary()")
	end := strings.Index(src, "// researchAreaOrder")
	if start < 0 || end <= start {
		t.Fatal("找不到回合摘要來源切片")
	}
	slice := src[start:end]
	if strings.Contains(slice, ".tr(") {
		t.Fatal("turnSummary 不得再用 tr 內嵌雙語文案")
	}
	for _, fixed := range []string{"Stardate %d report", "財政危機：出售", "完成一項研究"} {
		if strings.Contains(slice, `"`+fixed) {
			t.Errorf("turnSummary 仍內嵌固定玩家文案 %q", fixed)
		}
	}
}
