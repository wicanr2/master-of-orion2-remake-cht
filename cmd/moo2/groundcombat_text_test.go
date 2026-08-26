package main

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func groundCombatTextKeys() []string {
	return []string{
		"gallery.colony.demo",
		"groundcombat.transition.galaxy", "groundcombat.error.no_session",
		"groundcombat.title.default", "groundcombat.title.colony",
		"groundcombat.side.attacker", "groundcombat.side.defender",
		"groundcombat.attacker.marines", "groundcombat.attacker.armor", "groundcombat.attacker.survivors",
		"groundcombat.defender.garrison", "groundcombat.defender.rounds",
		"groundcombat.outcome.repelled", "groundcombat.outcome.success", "groundcombat.outcome.captured",
		"groundcombat.button.continue",
	}
}

func TestGroundCombatPlayerTextComesFromExternalCatalog(t *testing.T) {
	for _, key := range groundCombatTextKeys() {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("地面戰缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestGroundCombatExternalTextFitsSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		title := fmt.Sprintf(uiText(lang, "groundcombat.title.colony"), strings.Repeat("ABCDEFGHIJKLMNOPQRSTUVWXYZ銀河殖民地", 6))
		checkClippedTextFits(t, fnt, groundCombatTitleTextRect(), title, 16)
		attack := []string{
			uiText(lang, "groundcombat.side.attacker"),
			fmt.Sprintf(uiText(lang, "groundcombat.attacker.marines"), 2147483647, 2147483647),
			fmt.Sprintf(uiText(lang, "groundcombat.attacker.armor"), 2147483647, 2147483647),
			fmt.Sprintf(uiText(lang, "groundcombat.attacker.survivors"), 2147483647),
		}
		defense := []string{
			uiText(lang, "groundcombat.side.defender"),
			fmt.Sprintf(uiText(lang, "groundcombat.defender.garrison"), 2147483647, 2147483647),
			fmt.Sprintf(uiText(lang, "groundcombat.defender.rounds"), 2147483647),
		}
		for i, value := range attack {
			checkClippedTextFits(t, fnt, groundCombatSideTextRect(gcAtkTextX, i), value, 10)
		}
		for i, value := range defense {
			checkClippedTextFits(t, fnt, groundCombatSideTextRect(gcDefTextX, i), value, 10)
		}
		for _, key := range []string{"groundcombat.outcome.repelled", "groundcombat.outcome.success", "groundcombat.outcome.captured"} {
			checkClippedTextFits(t, fnt, groundCombatOutcomeTextRect(), uiText(lang, key), 15)
		}
		g := &groundCombatScreen{}
		checkClippedTextFits(t, fnt, g.continueTextRect(), uiText(lang, "groundcombat.button.continue"), 14)
	}
}

func TestGroundCombatSafeRectsStayInsideOwners(t *testing.T) {
	for _, tc := range []struct {
		panelX int
		textX  int
		rows   int
	}{{gcAtkPanelX, gcAtkTextX, 4}, {gcDefPanelX, gcDefTextX, 3}} {
		lastBottom := -1
		for i := 0; i < tc.rows; i++ {
			r := groundCombatSideTextRect(tc.textX, i)
			if r.x < tc.panelX || r.x+r.w > tc.panelX+gcPanelW || r.y < gcPanelY || r.y+r.h > gcPanelY+gcPanelH {
				t.Fatalf("地面戰第 %d 列超出兵力面板：%+v", i, r)
			}
			if r.y < lastBottom {
				t.Fatalf("地面戰第 %d 列與上一列重疊：%+v", i, r)
			}
			lastBottom = r.y + r.h
		}
	}
	g := &groundCombatScreen{}
	x, y, w, h := g.contRect()
	r := g.continueTextRect()
	if 2*r.x+r.w != 2*x+w || 2*r.y+r.h != 2*y+h {
		t.Fatal("地面戰按鈕文字框與可見外框／熱區中心不同")
	}
}

func TestGroundCombatRealAssets(t *testing.T) {
	dir := os.Getenv("MOO2_GROUND_TEST")
	if dir == "" {
		t.Skip("未設 MOO2_GROUND_TEST，跳過正版 COLGCBT.LBX 測試")
	}
	res, err := assets.NewResolver(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []int{gcbtFrameAsset, gcbtMarineAsset, gcbtDefMarineAsset, gcbtTankAsset, gcbtBattleoidAsset} {
		if got := loadGroundSprite(res, id, 2, id != gcbtFrameAsset); got == nil {
			t.Errorf("COLGCBT.LBX#%d 未能以殖民地調色盤解碼", id)
		}
	}
}

func TestGroundCombatPlayerColorIndexMapsMatchOriginalTables(t *testing.T) {
	src := &lbx.Frame{W: 4, H: 1, Index: []uint8{0xC0, 0xC7, 0xE8, 0xEB}, Written: []bool{true, true, true, true}}
	got := remapGroundPlayerIndexes(src, 3)
	want := []uint8{0x59, 0x6C, 0x56, 0x6B}
	if !slices.Equal(got.Index, want) {
		t.Fatalf("raw color 3 index map = % X，want % X", got.Index, want)
	}
	if !slices.Equal(src.Index, []uint8{0xC0, 0xC7, 0xE8, 0xEB}) {
		t.Fatal("地面戰換色不得修改原始 frame")
	}
	if same := remapGroundPlayerIndexes(src, 2); same != src {
		t.Fatal("raw color 2 應依原版直接保留 COLGCBT 原生色")
	}
}

func TestGroundCombatSourceHasNoEmbeddedPlayerTextOrDirectDraw(t *testing.T) {
	raw, err := os.ReadFile("groundcombat.go")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	for _, forbidden := range []string{".tr(", ".fnt.Draw(", ".fnt.DrawCentered("} {
		if strings.Contains(src, forbidden) {
			t.Errorf("groundcombat.go 不得出現 %s", forbidden)
		}
	}
	for _, value := range []string{"GROUND COMBAT", "ATTACKER", "DEFENDER", "入侵成功", "繼續", "星系主畫面"} {
		if strings.Contains(src, `"`+value+`"`) {
			t.Errorf("groundcombat.go 仍內嵌玩家文案 %q", value)
		}
	}
	galleryRaw, err := os.ReadFile("interactive.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"Demo Colony", "示範殖民地"} {
		if strings.Contains(string(galleryRaw), `"`+value+`"`) {
			t.Errorf("interactive.go 仍內嵌地面戰畫廊文案 %q", value)
		}
	}
}
