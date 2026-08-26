package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestHotseatPlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"hotseat.handoff.title",
		"hotseat.handoff.next_player",
		"hotseat.handoff.seat_position",
		"hotseat.handoff.instruction",
		"hotseat.handoff.note.resolve",
		"hotseat.handoff.button.take_over",
		"hotseat.transition.galaxy",
		"hiscore.transition.screen",
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("熱座交接缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestHotseatTextFitsSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	h := &hotseatScreen{fnt: fnt, seat: 7, total: 8}
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		checkClippedTextFits(t, fnt, h.titleTextRect(), uiText(lang, "hotseat.handoff.title"), 18)
		checkClippedTextFits(t, fnt, h.nextPlayerTextRect(),
			fmt.Sprintf(uiText(lang, "hotseat.handoff.next_player"), "ABCDEFGHIJKLMNOPQRSTUVWXYZ 帝國司令官"), 14)
		checkClippedTextFits(t, fnt, h.seatTextRect(),
			fmt.Sprintf(uiText(lang, "hotseat.handoff.seat_position"), 8, 8), 11)
		checkHotseatWrappedTextFits(t, fnt, h.instructionTextRect(), uiText(lang, "hotseat.handoff.instruction"), 13)
		checkHotseatWrappedTextFits(t, fnt, h.noteTextRect(), uiText(lang, "hotseat.handoff.note.resolve"), 12)
		checkClippedTextFits(t, fnt, h.okTextRect(), uiText(lang, "hotseat.handoff.button.take_over"), 15)
	}

	bx, by, bw, bh := h.okRect()
	buttonText := h.okTextRect()
	if 2*buttonText.x+buttonText.w != 2*bx+bw || 2*buttonText.y+buttonText.h != 2*by+bh {
		t.Error("接手按鈕文字框與點擊熱區中心不一致")
	}
	if h.noteTextRect().y+h.noteTextRect().h > by {
		t.Errorf("結算提示底緣 %d 侵入按鈕頂緣 %d", h.noteTextRect().y+h.noteTextRect().h, by)
	}
}

func checkHotseatWrappedTextFits(t *testing.T, fnt *uifont.Font, r textSafeRect, value string, size float64) {
	t.Helper()
	lines := r.lines(fnt, value, size)
	if len(lines) > r.maxLines() {
		t.Fatalf("%q 產生 %d 行，超過安全框上限 %d", value, len(lines), r.maxLines())
	}
	for _, line := range lines {
		w, h := fnt.Measure(line, size)
		if w > r.contentWidth() || h > float64(r.lineH) {
			t.Fatalf("%q 的行 %q 尺寸 %.1fx%.1f 超出安全框 %.1fx%d", value, line, w, h, r.contentWidth(), r.lineH)
		}
	}
}

func TestHotseatSourceHasNoEmbeddedPlayerSentencesOrDirectDraw(t *testing.T) {
	raw, err := os.ReadFile("hotseat.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, forbidden := range []string{".tr(", ".fnt.Draw(", ".fnt.DrawCentered("} {
		if strings.Contains(src, forbidden) {
			t.Errorf("hotseat.go 不得出現 %s", forbidden)
		}
	}
	for _, value := range []string{
		"PASS THE KEYBOARD", "換人接手", "Next:", "下一位：",
		"Pass the keyboard", "請把鍵盤交給", "TAKE OVER", "接手後進行結算",
	} {
		if strings.Contains(src, `"`+value) {
			t.Errorf("hotseat.go 仍內嵌玩家文案 %q", value)
		}
	}
}
