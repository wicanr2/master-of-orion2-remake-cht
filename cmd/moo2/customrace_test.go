package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestCustomRaceSpecialsCarryTraitsAndFitLogicalCanvas(t *testing.T) {
	specials := defaultSpecials()
	if len(specials) != 22 {
		t.Fatalf("客製種族特殊能力應有官方 22 項,得到 %d", len(specials))
	}
	for i, sp := range specials {
		if sp.trait == 0 {
			t.Errorf("第 %d 項特殊能力沒有特性編號", i)
		}
	}
	for i := range specials {
		x, y, w, h := (&customRaceScreen{}).spcRect(i)
		if x < crSpcX || x+w > 640 || y < crSpcY || y+h > 440 {
			t.Fatalf("第 %d 項特殊能力矩形超出邏輯畫布或撞到底部按鈕: (%d,%d,%d,%d)", i, x, y, w, h)
		}
	}
}

func TestCustomRaceLabelsMatchRACESTUFAndExternalJSON(t *testing.T) {
	cases := []struct{ key, en, zh string }{
		{"customrace.category.population", "Population", "人口"},
		{"customrace.option.growth.minus50", "-50% Growth", "-50% 成長"},
		{"customrace.category.ship_defense", "Ship Defense", "艦船防禦"},
		{"customrace.option.defense.plus25", "+25", "+25"},
		{"customrace.special.trans_dimensional", "Trans Dimensional", "跨次元"},
	}
	for _, tc := range cases {
		if got := uiText(i18n.English, tc.key); got != tc.en {
			t.Errorf("%s 英文=%q，預期 RACESTUF %q", tc.key, got, tc.en)
		}
		if got := uiText(i18n.Traditional, tc.key); got != tc.zh {
			t.Errorf("%s 繁中=%q，預期 %q", tc.key, got, tc.zh)
		}
	}
	for _, cat := range defaultPickCats() {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, cat.textKey); got == cat.textKey {
				t.Errorf("類別缺少外部文案：%s", cat.textKey)
			}
		}
		for i, opt := range cat.opts {
			if i == 0 && cat.id != pickCatGovernment {
				if opt.textKey != "" {
					t.Errorf("%s 未選狀態不得自創文字鍵：%s", cat.id, opt.textKey)
				}
				continue
			}
			for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
				if got := uiText(lang, opt.textKey); got == opt.textKey {
					t.Errorf("選項缺少外部文案：%s", opt.textKey)
				}
			}
		}
	}
	for _, sp := range defaultSpecials() {
		if uiText(i18n.English, sp.textKey) == sp.textKey || uiText(i18n.Traditional, sp.textKey) == sp.textKey {
			t.Errorf("特殊能力缺少雙語外部文案：%s", sp.textKey)
		}
	}
}

func TestCustomRaceTextSafeRectsContainBothLanguages(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	s := &customRaceScreen{cats: defaultPickCats(), specials: defaultSpecials()}
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		for i, cat := range s.cats {
			checkClippedTextFits(t, fnt, s.catNameTextRect(i), uiText(lang, cat.textKey), 13)
			for _, opt := range cat.opts {
				label := ""
				if opt.textKey != "" {
					label = uiText(lang, opt.textKey)
				}
				if opt.cost != 0 {
					label += fmt.Sprintf(" (%+d)", -opt.cost)
				}
				checkClippedTextFits(t, fnt, s.catOptionTextRect(i), label, 13)
			}
		}
		for i, sp := range s.specials {
			checkClippedTextFits(t, fnt, s.specialLabelTextRect(i), "● "+uiText(lang, sp.textKey), 12)
			checkClippedTextFits(t, fnt, s.specialCostTextRect(i), fmt.Sprintf("%+d", -sp.cost), 12)
		}
		checkClippedTextFits(t, fnt, customRaceTitleTextRect(), uiText(lang, "customrace.title"), 18)
		checkClippedTextFits(t, fnt, customRacePicksTextRect(),
			fmt.Sprintf(uiText(lang, "customrace.picks_remaining"), -10, startingPicks), 14)
	}
}

func checkClippedTextFits(t *testing.T, fnt *uifont.Font, r textSafeRect, value string, size float64) {
	t.Helper()
	clipped := r.clipped(fnt, value, size)
	w, h := fnt.Measure(clipped, size)
	if w > r.contentWidth() || h > float64(r.h-2*r.insetY) {
		t.Fatalf("%q 裁切後 %.1fx%.1f 超出安全框 %.1fx%d", value, w, h, r.contentWidth(), r.h-2*r.insetY)
	}
}

func TestCustomRaceSourceHasNoEmbeddedPlayerSentences(t *testing.T) {
	raw, err := os.ReadFile("customrace.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	if strings.Contains(src, ".tr(") {
		t.Fatal("customrace.go 不得再用 tr 內嵌中英文玩家文案")
	}
	for _, value := range []string{"Population Growth", "人口成長", "CUSTOM RACE", "自訂種族", "Fantastic Traders", "貿易奇才"} {
		if strings.Contains(src, `"`+value+`"`) {
			t.Errorf("customrace.go 仍內嵌玩家文案 %q", value)
		}
	}
}

func TestCustomRaceNumericCombatPickCategoriesReachSeparateRaceFields(t *testing.T) {
	cats := defaultPickCats()
	for i := range cats {
		switch cats[i].id {
		case pickCatShipAttack, pickCatShipDefense, pickCatGroundCombat, pickCatSpying:
			cats[i].sel = 2 // 各自選官方表的中間正向檔
		}
	}
	r := customRaceValues(cats)
	if r.CombatPct != 20 || r.ShipDefPct != 25 || r.GroundCombatBonus != 10 || r.SpyBonus != 10 {
		t.Fatalf("四類戰鬥／諜報 picks 應分別寫入艦攻／艦防／地面／諜報: %+v", r)
	}
}
