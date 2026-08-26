package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestDiplomacyAudienceFixedTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"diplomacy.audience.fallback_enemy", "diplomacy.audience.opening",
		"diplomacy.audience.title", "diplomacy.audience.emissary",
		"diplomacy.audience.agreements", "diplomacy.audience.option.peace",
		"diplomacy.audience.option.trade", "diplomacy.audience.option.research",
		"diplomacy.audience.option.nonaggression", "diplomacy.audience.option.alliance",
		"diplomacy.audience.option.threat", "diplomacy.audience.option.tribute_5",
		"diplomacy.audience.option.tribute_10", "diplomacy.audience.option.gift_cash",
		"diplomacy.audience.option.special_food", "diplomacy.audience.option.special_research",
		"diplomacy.audience.option.gift_tech", "diplomacy.audience.option.gift_star",
		"diplomacy.audience.break.trade", "diplomacy.audience.break.research",
		"diplomacy.audience.break.formal", "diplomacy.audience.break.tribute",
		"diplomacy.audience.break.special", "diplomacy.audience.button.end",
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("外交對談缺少外部文案 %s (%v)", key, lang)
			}
		}
	}
	for _, key := range []string{
		"diplomacy.audience.opening", "diplomacy.audience.emissary",
		"diplomacy.audience.agreements",
	} {
		for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
			if got := strings.Count(uiText(lang, key), "%s"); got != 1 {
				t.Errorf("%s (%v) 的格式參數數量=%d，預期 1", key, lang, got)
			}
		}
	}
}

func TestDiplomacyAudienceButtonTextRectsMatchHitRectsAndFit(t *testing.T) {
	d := &diplomacyScreen{}
	fnt := uifont.LoadBitmapTC()
	for i, key := range []string{
		"diplomacy.audience.option.peace", "diplomacy.audience.option.trade",
		"diplomacy.audience.option.research", "diplomacy.audience.option.nonaggression",
		"diplomacy.audience.option.alliance", "diplomacy.audience.option.threat",
		"diplomacy.audience.option.tribute_5", "diplomacy.audience.option.tribute_10",
		"diplomacy.audience.option.gift_cash", "diplomacy.audience.option.special_food",
		"diplomacy.audience.option.special_research", "diplomacy.audience.option.gift_tech",
		"diplomacy.audience.option.gift_star",
	} {
		x, y, w, h := d.optRect(i)
		r := d.optTextRect(i)
		if r.x != x || r.y != y || r.w != w || r.h != h {
			t.Fatalf("提議 %d 的文字框未由熱區推導：rect=%+v hit=%d,%d,%d,%d", i, r, x, y, w, h)
		}
		for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
			checkClippedTextFits(t, fnt, r, uiText(lang, key), 15)
		}
	}
	for i, key := range []string{
		"diplomacy.audience.break.trade", "diplomacy.audience.break.research",
		"diplomacy.audience.break.formal", "diplomacy.audience.break.tribute",
		"diplomacy.audience.break.special",
	} {
		x, y, w, h := d.breakRect(i)
		r := d.breakTextRect(i)
		if r.x != x || r.y != y || r.w != w || r.h != h {
			t.Fatalf("解約 %d 的文字框未由熱區推導：rect=%+v hit=%d,%d,%d,%d", i, r, x, y, w, h)
		}
		if x < 0 || y < 0 || x+w > moo2ScreenW || y+h > moo2ScreenH {
			t.Fatalf("解約 %d 超出 640×480 畫布：%d,%d,%d,%d", i, x, y, w, h)
		}
		_, optionY, _, optionH := d.optRect(12)
		if y < optionY+optionH {
			t.Fatalf("解約列與最後一列提議重疊：breakY=%d optionBottom=%d", y, optionY+optionH)
		}
		for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
			checkClippedTextFits(t, fnt, r, uiText(lang, key), 11)
		}
	}
}

func TestDiplomacyAudienceLegacyFixedTextIsNotEmbedded(t *testing.T) {
	raw, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, text := range []string{
		"外交對談", "提議和平", "Propose Peace",
		"特殊貿易：食物換現金", "Special Trade: Food for Credits",
		"終止特殊貿易", "End Special Trade", "結束對談", "END AUDIENCE",
	} {
		if strings.Contains(source, `"`+text+`"`) {
			t.Errorf("interactive.go 仍內嵌外交對談固定文案 %q", text)
		}
	}
}

func TestDiplomacyResultCodesHaveExternalBilingualTemplates(t *testing.T) {
	codes := []shell.DiplomacyResultCode{
		shell.DiploResultFormalExists, shell.DiploResultPeaceStrong, shell.DiploResultPeaceWeak,
		shell.DiploResultTradeExists, shell.DiploResultTradeStarted, shell.DiploResultResearchExists,
		shell.DiploResultResearchStarted, shell.DiploResultSpecialExists, shell.DiploResultSpecialFoodStarted,
		shell.DiploResultSpecialResStarted, shell.DiploResultFormalConflict, shell.DiploResultNAPStarted,
		shell.DiploResultAllianceStarted, shell.DiploResultTributeExists, shell.DiploResultTribute5Started,
		shell.DiploResultTribute10Started, shell.DiploResultNoGiftTech, shell.DiploResultNoGiftStar,
		shell.DiploResultNoTrade, shell.DiploResultTradeEnded, shell.DiploResultNoResearch,
		shell.DiploResultResearchEnded, shell.DiploResultNoFormal, shell.DiploResultFormalEnded,
		shell.DiploResultNoTribute, shell.DiploResultTributeEnded, shell.DiploResultNoSpecial,
		shell.DiploResultSpecialEnded, shell.DiploResultThreatStrong, shell.DiploResultThreatWeak,
		shell.DiploResultCashInvalid, shell.DiploResultCashInsufficient, shell.DiploResultCashAccepted,
		shell.DiploResultTechUnknown, shell.DiploResultTechKnown, shell.DiploResultTechAccepted,
		shell.DiploResultStarInvalid, shell.DiploResultStarLastColony, shell.DiploResultStarNotOwned,
		shell.DiploResultStarAlreadyOwned, shell.DiploResultStarAccepted,
	}
	for _, code := range codes {
		key := "diplomacy.response." + string(code)
		for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
			template := uiText(lang, key)
			if template == "" || template == key {
				t.Fatalf("缺少外交結果外部文案 %s (%v)", key, lang)
			}
			result := shell.DiplomacyResult{Code: code, Enemy: "ALKARI", Amount: 10, Available: 4, Detail: "TECH"}
			got := diplomacyResultText(lang, result)
			if got == "" || strings.Contains(got, "%!") || strings.Contains(got, key) {
				t.Errorf("外交結果格式化失敗 %s (%v)：%q", code, lang, got)
			}
		}
	}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		got := diplomacyResultText(lang, shell.DiplomacyResult{Code: "not_registered", Enemy: "ALKARI"})
		if got == "" || strings.Contains(got, "not_registered") || strings.Contains(got, "%!") {
			t.Errorf("未知外交結果未安全 fallback (%v)：%q", lang, got)
		}
	}
}

func TestDiplomacyRuleFilesDoNotEmbedPlayerResponses(t *testing.T) {
	for _, path := range []string{"../../internal/shell/session.go", "../../internal/shell/diplomacy_gift.go"} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		source := string(raw)
		for _, phrase := range []string{
			"貿易協定成立", "研究協定成立", "互不侵犯條約成立", "同盟成立",
			"國庫只有 %d BC", "我們接受你贈送的", "至少要保留一座殖民地",
		} {
			if strings.Contains(source, phrase) {
				t.Errorf("%s 仍內嵌玩家外交回應 %q", path, phrase)
			}
		}
	}
}

func TestTreatySummaryUsesExternalBilingualTemplates(t *testing.T) {
	state := shell.TreatyState{
		FormalPolicy: gamedata.DIPLO_ALLIANCE, PlayerTribute: shell.TributeFivePercent,
		TradeActive: true, TradeTurns: 9, PlayerTradeValue: 12,
		ResearchActive: true, ResearchTurns: 8, PlayerResearchValue: -3,
		SpecialTrade: shell.SpecialTradeState{Active: true, Kind: shell.SpecialTradeResearchExchange, Turns: 7},
	}
	keys := []string{
		"diplomacy.summary.unknown", "diplomacy.summary.formal_none",
		"diplomacy.summary.nonaggression", "diplomacy.summary.alliance", "diplomacy.summary.peace",
		"diplomacy.summary.paying_tribute", "diplomacy.summary.receiving_tribute",
		"diplomacy.summary.trade", "diplomacy.summary.research",
		"diplomacy.summary.special_food", "diplomacy.summary.special_research",
		"diplomacy.summary.special_unknown",
	}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		for _, key := range keys {
			if got := uiText(lang, key); got == "" || got == key {
				t.Fatalf("缺少條約摘要外部文案 %s (%v)", key, lang)
			}
		}
		got := treatySummaryText(lang, state)
		if got == "" || strings.Contains(got, "%!") || strings.Contains(got, "diplomacy.summary") {
			t.Errorf("條約摘要格式化失敗 (%v)：%q", lang, got)
		}
		unknown := treatySummaryPartText(lang, shell.TreatySummaryPart{Kind: "not_registered"})
		if unknown == "" || strings.Contains(unknown, "not_registered") {
			t.Errorf("未知條約摘要未安全 fallback (%v)：%q", lang, unknown)
		}
	}
}

func TestTreatyRuleDoesNotEmbedSummaryLabels(t *testing.T) {
	raw, err := os.ReadFile("../../internal/shell/treaty.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, phrase := range []string{
		"Non-Aggression Pact", "無正式條約", "Paying tribute", "進貢%d%%",
		"Trade %d turns", "貿易第%d回合", "食物換現金", "研究交換",
	} {
		if strings.Contains(source, `"`+phrase) {
			t.Errorf("treaty.go 仍內嵌條約摘要文案 %q", phrase)
		}
	}
}
