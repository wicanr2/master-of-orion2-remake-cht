package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestChooseMultiNetGamePlayerTextComesFromExternalCatalog(t *testing.T) {
	keys := []string{
		"netgames.title",
		"netgames.empty.title",
		"netgames.empty.hint",
		"netgames.button.direct_address",
		"netgames.button.cancel",
		"netgames.error.selected_game",
		"netgames.error.direct_address",
		"netgames.demo.orion",
		"netgames.demo.sakkra",
		"netgames.demo.antares",
	}
	for _, key := range keys {
		for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
			if got := uiText(lang, key); got == "" || got == key {
				t.Errorf("區網對局選擇缺少外部雙語文案：%s (%v)", key, lang)
			}
		}
	}
}

func TestChooseMultiNetGameStaticTextFitsSafeRects(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	winX, winY := cmngWindow()
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		checkFull := func(name string, rect textSafeRect, value string, size float64) {
			t.Helper()
			checkClippedTextFits(t, fnt, rect, value, size)
			if got := rect.clipped(fnt, value, size); got != value {
				t.Errorf("%s 不得被裁切：%q → %q", name, value, got)
			}
		}
		checkFull("標題", cmngTitleTextRect(winX, winY), uiText(lang, "netgames.title"), 16)
		checkFull("空清單", cmngEmptyTitleTextRect(winX, winY), uiText(lang, "netgames.empty.title"), 13)
		checkFull("空清單提示", cmngEmptyHintTextRect(winX, winY), uiText(lang, "netgames.empty.hint"), 11)
		checkFull("直接位址", cmngDirectTextRect(winX, winY), uiText(lang, "netgames.button.direct_address"), 12)
		checkFull("取消", cmngCancelTextRect(winX, winY), uiText(lang, "netgames.button.cancel"), 13)
		checkFull("已探索對局錯誤", cmngMessageTextRect(winX, winY), uiText(lang, "netgames.error.selected_game"), 12)
		errorText := fmt.Sprintf(uiText(lang, "netgames.error.direct_address"), "192.168.100.200:24501")
		if strings.Contains(errorText, "%!") {
			t.Errorf("直接位址錯誤格式參數不相容：%q", errorText)
		}
		checkFull("直接位址錯誤", cmngMessageTextRect(winX, winY), errorText, 12)
	}
}

func TestChooseMultiNetGameRowTextColumnsAreBounded(t *testing.T) {
	fnt := uifont.LoadBitmapTC()
	winX, winY := cmngWindow()
	for i := 0; i < cmngMaxRows; i++ {
		rx, ry, rw, rh := cmngRowRect(winX, winY, i)
		columns := []textSafeRect{
			cmngRowNameTextRect(winX, winY, i),
			cmngRowAddressTextRect(winX, winY, i),
			cmngRowPlayersTextRect(winX, winY, i),
		}
		for j, rect := range columns {
			if rect.x < rx || rect.y < ry || rect.x+rect.w > rx+rw || rect.y+rect.h > ry+rh {
				t.Errorf("第 %d 列第 %d 欄超出 row：%+v row=(%d,%d,%d,%d)", i, j, rect, rx, ry, rw, rh)
			}
			if j > 0 && columns[j-1].x+columns[j-1].w > rect.x {
				t.Errorf("第 %d 列第 %d、%d 欄重疊", i, j-1, j)
			}
		}
		checkClippedTextFits(t, fnt, columns[0], strings.Repeat("超長對局名稱", 12), 13)
		checkClippedTextFits(t, fnt, columns[1], "example.really-long-host-name.invalid:65535", 11)
		checkClippedTextFits(t, fnt, columns[2], "99/99", 11)
	}
}

func TestChooseMultiNetGameButtonTextSharesHitRectCenter(t *testing.T) {
	winX, winY := cmngWindow()
	dx, dy, dw, dh := cmngDirectRect(winX, winY)
	direct := cmngDirectTextRect(winX, winY)
	if 2*direct.x+direct.w != 2*dx+dw || 2*direct.y+direct.h != 2*dy+dh {
		t.Errorf("直接位址文字框與熱區中心不一致：text=%+v hit=(%d,%d,%d,%d)", direct, dx, dy, dw, dh)
	}
	cx, cy, cw, ch := cmngCancelRect(winX, winY)
	cancel := cmngCancelTextRect(winX, winY)
	if 2*cancel.x+cancel.w != 2*cx+cw || 2*cancel.y+cancel.h != 2*cy+ch {
		t.Errorf("取消文字框與按鈕中心不一致：text=%+v button=(%d,%d,%d,%d)", cancel, cx, cy, cw, ch)
	}
}

func TestChooseMultiNetGameSourceHasNoEmbeddedPlayerSentences(t *testing.T) {
	raw, err := os.ReadFile("choosemultinetgame.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	if strings.Contains(source, ".tr(") {
		t.Fatal("choosemultinetgame.go 不得再以 tr 內嵌雙語玩家文案")
	}
	for _, value := range []string{
		"選擇要加入的對局", "區網上沒有偵測到對局", "直接輸入位址", "No games found on the LAN",
		"Could not connect", "ORION", "SAKKRA", "ANTARES",
	} {
		if strings.Contains(source, `"`+value) {
			t.Errorf("choosemultinetgame.go 仍內嵌玩家文案 %q", value)
		}
	}
}
