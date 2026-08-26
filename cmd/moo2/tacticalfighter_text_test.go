package main

import (
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
)

func TestTacticalFighterFixedTextIsExternalized(t *testing.T) {
	src, err := os.ReadFile("tacticalfighter.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	for _, forbidden := range []string{".tr(", ".DrawCentered(", ".Draw("} {
		if strings.Contains(text, forbidden) {
			t.Errorf("tacticalfighter.go 不得直接內嵌或繪製固定文案：仍含 %q", forbidden)
		}
	}
	for _, key := range []string{
		"tactical.fighter.log.launch", "tactical.fighter.log.recovered",
		"tactical.fighter.glyph.interceptor", "tactical.fighter.glyph.heavy",
		"tactical.fighter.glyph.returning", "tactical.fighter.button.launch",
		"tactical.fighter.status.active", "tactical.fighter.log.boarding_success",
		"tactical.fighter.log.boarding_failed",
	} {
		if got := uiText(i18n.Traditional, key); got == "" || got == key {
			t.Errorf("缺少繁中鍵 %s", key)
		}
		if got := uiText(i18n.English, key); got == "" || got == key {
			t.Errorf("缺少英文鍵 %s", key)
		}
	}
}

func TestTacticalFighterLaunchTextRectMatchesButton(t *testing.T) {
	x, y, w, h := launchRect()
	r := textSafeRect{x: x, y: y, w: w, h: h, insetX: 5, insetY: 3}
	if r.x != x || r.y != y || r.w != w || r.h != h || r.x+r.w > moo2ScreenW || r.y+r.h > moo2ScreenH {
		t.Fatalf("出擊文字框必須與按鈕共用邊界且留在畫面內：%+v", r)
	}
}
