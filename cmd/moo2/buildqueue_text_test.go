package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestBuildQueuePlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"buildqueue.error.no_colony", "buildqueue.transition.colony", "buildqueue.transition.ship_design",
		"buildqueue.message.auto_enabled", "buildqueue.message.auto_disabled",
		"buildqueue.message.repeat_cancelled", "buildqueue.message.choose_repeat",
		"buildqueue.message.unavailable", "buildqueue.message.repeat_invalid", "buildqueue.message.repeat_set",
		"buildqueue.empty.available", "buildqueue.empty.slot", "buildqueue.queue.refit",
		"buildqueue.queue.eta", "buildqueue.queue.progress", "buildqueue.list.cost",
		"buildqueue.title.default_colony", "buildqueue.title.format", "buildqueue.hint.queue_remove",
		"buildqueue.status.choose_repeat", "buildqueue.status.repeat_active", "buildqueue.status.auto_on",
		"buildqueue.item.trade_goods", "buildqueue.item.housing",
	}
	for _, button := range bqButtons {
		keys = append(keys, button.textKey)
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("建造佇列缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestBuildQueueTextFitsSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		for _, button := range bqButtons {
			r := buildQueueButtonTextRect(button)
			checkClippedTextFits(t, fnt, r, uiText(lang, button.textKey), 11)
			if 2*r.x+r.w != 2*button.x+button.w || 2*r.y+r.h != 2*button.y+button.h {
				t.Errorf("%s 文字框與熱區中心不同", button.act)
			}
		}
		checkClippedTextFits(t, fnt, buildQueueTitleTextRect(),
			fmt.Sprintf(uiText(lang, "buildqueue.title.format"), "ABCDEFGHIJKLMNOPQRSTUVWXYZ"), 14)
		checkClippedTextFits(t, fnt, buildQueueHintTextRect(), uiText(lang, "buildqueue.hint.queue_remove"), 11)
		checkClippedTextFits(t, fnt, buildQueueStatusTextRect(),
			fmt.Sprintf(uiText(lang, "buildqueue.status.repeat_active"), "Trans Dimensional Colony Ship"), 11)
		assertBuildQueueMultilineFits(t, fnt, buildQueueMessageTextRect(), uiText(lang, "buildqueue.message.repeat_invalid"), 11)
	}
	msg := buildQueueMessageTextRect()
	if msg.y+msg.h > bqQueueY0 || msg.y < 0 || msg.y+msg.h > 480 {
		t.Fatalf("狀態訊息框超出畫布或侵入七格佇列：%+v", msg)
	}
}

func assertBuildQueueMultilineFits(t *testing.T, fnt *uifont.Font, r textSafeRect, value string, size float64) {
	t.Helper()
	lines := r.lines(fnt, value, size)
	if len(lines) == 0 || len(lines) > r.maxLines() {
		t.Fatalf("%q 產生 %d 行，安全框最多 %d 行", value, len(lines), r.maxLines())
	}
	for _, line := range lines {
		w, h := fnt.Measure(line, size)
		if w > r.contentWidth() || h > float64(r.lineH) {
			t.Fatalf("%q 的收束行 %q 超出安全框：%.1fx%.1f", value, line, w, h)
		}
	}
}

func TestBuildQueueSourceHasNoEmbeddedPlayerSentencesOrDirectDraw(t *testing.T) {
	raw, err := os.ReadFile("buildqueue.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, forbidden := range []string{".tr(", ".fnt.Draw(", ".fnt.DrawCentered("} {
		if strings.Contains(src, forbidden) {
			t.Errorf("buildqueue.go 不得出現 %s", forbidden)
		}
	}
	for _, value := range []string{"AUTO BUILD", "自動建造", "Nothing available to queue", "目前沒有可排入的項目"} {
		if strings.Contains(src, `"`+value+`"`) {
			t.Errorf("buildqueue.go 仍內嵌玩家文案 %q", value)
		}
	}
}
