package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestAIRaidNoticeCatalogCoverage(t *testing.T) {
	keys := []string{"raid.notice.repelled", "raid.notice.breakthrough", "raid.notice.building_destroyed",
		"raid.notice.attacker_attrition", "raid.notice.detail_separator"}
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		for _, key := range keys {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("語系 %v 缺少 AI 突襲外部文案 %q", lang, key)
			}
		}
	}
}

func TestAIRaidNoticeTextUsesTypedResults(t *testing.T) {
	repelled := &shell.AIRaidReport{AIName: "AI（席隆人）", AINameEN: "Psilons",
		StarName: "測試星", StarNameEN: "Test", Repelled: true, FleetLost: 17}
	zh, en := aiRaidNoticeText(i18n.Traditional, repelled), aiRaidNoticeText(i18n.English, repelled)
	if !strings.Contains(zh, "17") || !strings.Contains(zh, "擊退") ||
		!strings.Contains(en, "Psilons") || !strings.Contains(en, "Test") || !strings.Contains(en, "17") {
		t.Fatalf("擊退通知沒有使用 typed 結果：zh=%q en=%q", zh, en)
	}
	if strings.Contains(en, "測試星") || strings.Contains(en, "席隆人") {
		t.Errorf("英文擊退通知洩漏中文：%q", en)
	}

	breakthrough := &shell.AIRaidReport{AIName: "AI（席隆人）", AINameEN: "Psilons",
		StarName: "測試星", StarNameEN: "Test", PopLost: 2, BCLost: 40,
		Building: "研究實驗室", FleetLost: 9}
	zh, en = aiRaidNoticeText(i18n.Traditional, breakthrough), aiRaidNoticeText(i18n.English, breakthrough)
	for _, want := range []string{"2", "40", "研究實驗室", "9"} {
		if !strings.Contains(zh, want) {
			t.Errorf("繁中突破通知缺少 %q：%q", want, zh)
		}
	}
	for _, want := range []string{"2", "40", "Research Laboratory", "9"} {
		if !strings.Contains(en, want) {
			t.Errorf("英文突破通知缺少 %q：%q", want, en)
		}
	}
	if strings.Contains(en, "研究實驗室") || strings.Contains(en, "測試星") {
		t.Errorf("英文突破通知洩漏中文規則名稱：%q", en)
	}
}

func TestAIRaidNoticeOptionalDetails(t *testing.T) {
	base := &shell.AIRaidReport{AIName: "敵軍", AINameEN: "Enemy", StarName: "星", StarNameEN: "Star", PopLost: 1, BCLost: 2}
	plain := aiRaidNoticeText(i18n.English, base)
	if strings.Contains(plain, "destroyed") || strings.Contains(plain, "defending force") {
		t.Errorf("沒有可選結果時不應顯示額外片段：%q", plain)
	}
	buildingOnly := *base
	buildingOnly.Building = "研究實驗室"
	buildingText := aiRaidNoticeText(i18n.English, &buildingOnly)
	if !strings.Contains(buildingText, "Research Laboratory") || strings.Contains(buildingText, "defending force") {
		t.Errorf("僅有建築損失時片段組合錯誤：%q", buildingText)
	}
	attritionOnly := *base
	attritionOnly.FleetLost = 7
	attritionText := aiRaidNoticeText(i18n.English, &attritionOnly)
	if strings.Contains(attritionText, "destroyed") || !strings.Contains(attritionText, "7 fleet strength") {
		t.Errorf("僅有攻方折損時片段組合錯誤：%q", attritionText)
	}
	unknown := aiRaidNoticeText(i18n.English, &shell.AIRaidReport{})
	if !strings.Contains(unknown, "Unknown Empire") || !strings.Contains(unknown, "Unknown") {
		t.Errorf("缺名結果沒有外部安全 fallback：%q", unknown)
	}
	if aiRaidNoticeText(i18n.Traditional, nil) != "" {
		t.Error("nil AI 突襲報告應回空字串")
	}
}

func TestAIRaidRuleSourceHasNoEmbeddedPlayerSentences(t *testing.T) {
	raw, err := os.ReadFile("../../internal/shell/ai_attack.go")
	if err != nil {
		t.Fatal(err)
	}
	src := strings.ToLower(string(raw))
	for _, forbidden := range []string{"遭防禦部隊擊退", "人口 -", "被摧毀", "raided %s", "was destroyed"} {
		if strings.Contains(src, `"`+strings.ToLower(forbidden)) {
			t.Errorf("AI 突襲規則層仍內嵌玩家文案或成品欄位 %q", forbidden)
		}
	}
	if strings.Contains(src, "messageen") || strings.Contains(src, "message string") {
		t.Error("AI 突襲規則層仍保存成品訊息欄位")
	}
}
