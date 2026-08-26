package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

var refitErrorCodes = []shell.RefitErrorCode{
	shell.RefitErrorNoColony, shell.RefitErrorSelectShip, shell.RefitErrorColonyMissing,
	shell.RefitErrorFleetMissing, shell.RefitErrorFleetNotParked, shell.RefitErrorShipMissing,
	shell.RefitErrorFacility, shell.RefitErrorNoUpgrade, shell.RefitErrorQueueFull,
	shell.RefitErrorEnqueueFailed,
}

func TestRefitPlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"refit.title", "refit.subtitle", "refit.empty", "refit.candidate.summary",
		"refit.preview.source_target", "refit.preview.detail", "refit.preview.scrap_warning",
		"refit.preview.select_prompt", "refit.button.queue", "refit.button.return",
		"refit.error.unknown",
	}
	for _, code := range refitErrorCodes {
		keys = append(keys, "refit.error."+string(code))
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("REFIT 缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestRefitFormatContractsMatch(t *testing.T) {
	wants := map[string]int{
		"refit.candidate.summary":     6,
		"refit.preview.source_target": 2,
		"refit.preview.detail":        5,
		"refit.error.no_upgrade":      1,
	}
	for key, want := range wants {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := strings.Count(uiText(lang, key), "%"); got != want {
				t.Errorf("%s (%v) 有 %d 個格式參數，預期 %d", key, lang, got, want)
			}
		}
	}
}

func TestRefitTypedErrorsAreLocalized(t *testing.T) {
	for _, code := range refitErrorCodes {
		err := &shell.RefitError{Code: code, ShipName: "長程測試艦"}
		var typed *shell.RefitError
		if !errors.As(err, &typed) || typed.Code != code {
			t.Fatalf("REFIT typed error 無法保留 code %s", code)
		}
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			got := localizedRefitError(lang, err)
			if got == "" || strings.HasPrefix(got, "refit.error.") || strings.HasPrefix(got, "refit:") {
				t.Errorf("REFIT code %s (%v) 未轉成玩家文案：%q", code, lang, got)
			}
		}
	}
	if got := localizedRefitError(i18n.Traditional, errors.New("opaque")); got != uiText(i18n.Traditional, "refit.error.unknown") {
		t.Errorf("未知錯誤 fallback=%q", got)
	}
}

func TestRefitTextFitsSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		checkClippedTextFits(t, fnt, refitTitleTextRect(), uiText(lang, "refit.title"), 18)
		checkClippedTextFits(t, fnt, refitSubtitleTextRect(), uiText(lang, "refit.subtitle"), 11)
		checkClippedTextFits(t, fnt, refitEmptyTextRect(), uiText(lang, "refit.empty"), 13)
		candidate := fmt.Sprintf(uiText(lang, "refit.candidate.summary"),
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ長程測試戰艦", "Doom Star", "Mauler Device",
			"Xentronium Armor", "Hard Shields", "Sub-Space Teleporter")
		for i := 0; i < refitListRows; i++ {
			checkClippedTextFits(t, fnt, refitListTextRect(i), candidate, 11)
		}
		source := fmt.Sprintf(uiText(lang, "refit.preview.source_target"),
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ來源艦", "ABCDEFGHIJKLMNOPQRSTUVWXYZ目標艦")
		checkClippedTextFits(t, fnt, refitPreviewSourceTextRect(), source, 12)
		detail := fmt.Sprintf(uiText(lang, "refit.preview.detail"),
			"Mauler Device", "Xentronium Armor", "Hard Shields", "Sub-Space Teleporter", 99999)
		assertBuildQueueMultilineFits(t, fnt, refitPreviewDetailTextRect(), detail, 11)
		checkClippedTextFits(t, fnt, refitPreviewWarningTextRect(), uiText(lang, "refit.preview.scrap_warning"), 10)
		checkClippedTextFits(t, fnt, refitPreviewPromptTextRect(), uiText(lang, "refit.preview.select_prompt"), 12)
		checkClippedTextFits(t, fnt, refitButtonTextRect(refitQueueX), uiText(lang, "refit.button.queue"), 12)
		checkClippedTextFits(t, fnt, refitButtonTextRect(refitCancelX), uiText(lang, "refit.button.return"), 12)
		checkClippedTextFits(t, fnt, refitMessageTextRect(), localizedRefitError(lang,
			&shell.RefitError{Code: shell.RefitErrorNoUpgrade, ShipName: "ABCDEFGHIJKLMNOPQRSTUVWXYZ長程測試戰艦"}), 11)
	}
	for _, x := range []int{refitQueueX, refitCancelX} {
		r := refitButtonTextRect(x)
		if 2*r.x+r.w != 2*x+refitButtonW || 2*r.y+r.h != 2*refitQueueY+28 {
			t.Errorf("REFIT 按鈕文字框與熱區中心不一致：x=%d", x)
		}
	}
}

func TestRefitSourcesHaveNoEmbeddedPlayerSentencesOrDirectDraw(t *testing.T) {
	uiRaw, err := os.ReadFile("refit.go")
	if err != nil {
		t.Fatal(err)
	}
	uiSource := string(uiRaw)
	for _, forbidden := range []string{".tr(", ".fnt.Draw(", ".fnt.DrawCentered("} {
		if strings.Contains(uiSource, forbidden) {
			t.Errorf("refit.go 不得出現 %s", forbidden)
		}
	}
	for _, value := range []string{"SHIP REFIT", "艦艇改裝", "QUEUE REFIT", "排入改裝"} {
		if strings.Contains(uiSource, `"`+value) {
			t.Errorf("refit.go 仍內嵌玩家文案 %q", value)
		}
	}
	ruleRaw, err := os.ReadFile("../../internal/shell/production_controls.go")
	if err != nil {
		t.Fatal(err)
	}
	ruleSource := string(ruleRaw)
	for _, value := range []string{"殖民地不存在", "艦隊不存在", "建造佇列已滿", "沒有可套用的同艦體升級"} {
		if strings.Contains(ruleSource, `"`+value) {
			t.Errorf("production_controls.go 的 REFIT 分支仍內嵌玩家文案 %q", value)
		}
	}
}
