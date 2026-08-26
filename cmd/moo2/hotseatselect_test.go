package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestHotseatEmpireSelectDefaultsAndTargetsExactAIIndices(t *testing.T) {
	b := &sceneBuilder{session: shell.NewDemoSession(), pendingHotseat: 3, lang: i18n.Traditional}
	screen, err := b.hotseatEmpireSelect()
	if err != nil {
		t.Fatalf("建立選帝國畫面失敗: %v", err)
	}
	s := screen.(*hotseatEmpireSelectScreen)
	if got := s.selectedIndices(); len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Fatalf("預設應選前兩個 AI,got %v", got)
	}

	// 取消第 0 個,再選第 2 個;選定數量維持兩席,而不是被迫接管尾端 AI。
	s.update(shell.InputState{MouseX: hseListX + 4, MouseY: hseListY + 4, ClickReleased: true})
	s.update(shell.InputState{MouseX: hseListX + 4, MouseY: hseListY + 2*hseRowH + 4, ClickReleased: true})
	got := s.selectedIndices()
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("指定後應為 AI 索引 [1 2],got %v", got)
	}
}

func TestHotseatEmpireSelectDraws(t *testing.T) {
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		b := &sceneBuilder{session: shell.NewDemoSession(), pendingHotseat: 2, lang: lang}
		screen, err := b.hotseatEmpireSelect()
		if err != nil {
			t.Fatalf("建立選帝國畫面失敗: %v", err)
		}
		dst := ebiten.NewImage(640, 480)
		screen.draw(dst)
	}
}

func TestHotseatEmpireSelectPlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"hotseat.empire_select.title", "hotseat.empire_select.instruction",
		"hotseat.empire_select.selected", "hotseat.empire_select.row",
		"hotseat.empire_select.mark.off", "hotseat.empire_select.mark.on",
		"hotseat.empire_select.button.back", "hotseat.empire_select.button.start",
		"hotseat.empire_select.transition.galaxy", "hotseat.empire_select.error.no_session",
		"hotseat.empire_select.error.too_few_seats",
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("熱座選帝國缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
	for key, want := range map[string]int{
		"hotseat.empire_select.instruction": 1,
		"hotseat.empire_select.selected":    2,
		"hotseat.empire_select.row":         2,
	} {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := strings.Count(uiText(lang, key), "%"); got != want {
				t.Errorf("%s (%v) 有 %d 個格式參數，預期 %d", key, lang, got, want)
			}
		}
	}
}

func TestHotseatEmpireSelectTextFitsSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		checkClippedTextFits(t, fnt, hotseatEmpireTitleTextRect(), uiText(lang, "hotseat.empire_select.title"), 17)
		checkClippedTextFits(t, fnt, hotseatEmpireInstructionTextRect(),
			fmt.Sprintf(uiText(lang, "hotseat.empire_select.instruction"), 7), 12)
		checkClippedTextFits(t, fnt, hotseatEmpireCountTextRect(),
			fmt.Sprintf(uiText(lang, "hotseat.empire_select.selected"), 7, 7), 11)
		b := &sceneBuilder{session: shell.NewDemoSession(), pendingHotseat: 2, lang: lang}
		screen, err := b.hotseatEmpireSelect()
		if err != nil {
			t.Fatal(err)
		}
		s := screen.(*hotseatEmpireSelectScreen)
		for i := range s.selected {
			checkClippedTextFits(t, fnt, s.markTextRect(i), uiText(lang, "hotseat.empire_select.mark.on"), 12)
			label := fmt.Sprintf(uiText(lang, "hotseat.empire_select.row"), i+2,
				"ABCDEFGHIJKLMNOPQRSTUVWXYZ超長自訂帝國名稱")
			checkClippedTextFits(t, fnt, s.rowTextRect(i), label, 12)
		}
		checkClippedTextFits(t, fnt, hotseatEmpireButtonTextRect(s.cancelRect),
			uiText(lang, "hotseat.empire_select.button.back"), 12)
		checkClippedTextFits(t, fnt, hotseatEmpireButtonTextRect(s.acceptRect),
			uiText(lang, "hotseat.empire_select.button.start"), 12)
	}
}

func TestHotseatEmpireSelectSourceHasNoEmbeddedPlayerTextOrDirectDraw(t *testing.T) {
	raw, err := os.ReadFile("hotseatselect.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, forbidden := range []string{".tr(", ".fnt.Draw(", ".fnt.DrawCentered("} {
		if strings.Contains(source, forbidden) {
			t.Errorf("hotseatselect.go 不得出現 %s", forbidden)
		}
	}
	for _, value := range []string{"指定真人帝國", "請選 %d 個帝國", "返回命名", "開始遊戲", "CHOOSE HUMAN EMPIRES"} {
		if strings.Contains(source, `"`+value) {
			t.Errorf("hotseatselect.go 仍內嵌玩家文案 %q", value)
		}
	}
}
