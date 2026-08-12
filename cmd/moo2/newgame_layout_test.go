package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestNewGameSelectorTextUsesCenteredSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.Traditional, i18n.English} {
		b := &sceneBuilder{lang: lang, newGameDiff: newGameDiffDefault, newGameSize: 1,
			newGameAge: newGameAgeDefault, newGameEmpires: shell.DefaultOpponents + 1, newGameTech: newGameTechDefault}
		for _, st := range ngSettings {
			r := ngStripTextRect(st)
			x, y, w, h := ngStripRect(st)
			if r.x < x || r.y < y || r.x+r.w > x+w || r.y+r.h > y+h {
				t.Fatalf("%s 安全框 (%d,%d,%d,%d) 超出選擇器 (%d,%d,%d,%d)", st.act, r.x, r.y, r.w, r.h, x, y, w, h)
			}
			if r.x+r.w/2 != x+w/2 || r.y+r.h/2 != y+h/2 {
				t.Fatalf("%s 文字中心沒有與選擇器對齊", st.act)
			}
			for i := 0; i < st.n(b); i++ {
				st.set(b, i)
				label := r.clipped(fnt, st.label(b), 12)
				if got, _ := fnt.Measure(label, 12); got > r.contentWidth() {
					t.Fatalf("%s 的 %q 寬 %.0f 超出 %.0f", st.act, label, got, r.contentWidth())
				}
			}
		}
	}
}
