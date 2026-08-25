package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func TestColonyTopTextSafeRectsStayInsideNativePanels(t *testing.T) {
	inside := func(name string, r textSafeRect, x, y, w, h int) {
		t.Helper()
		if r.x < x || r.y < y || r.x+r.w > x+w || r.y+r.h > y+h {
			t.Fatalf("%s 安全框 (%d,%d,%d,%d) 超出原生面板 (%d,%d,%d,%d)",
				name, r.x, r.y, r.w, r.h, x, y, w, h)
		}
	}

	inside("左欄標題", colonyLeftTitleTextRect(), colPanelLX, colPanelLY, colPanelLW, colPanelLH)
	inside("左欄一般資料", colonyLeftRowsTextRect(false), colPanelLX, colPanelLY, colPanelLW, colPanelLH)
	inside("左欄同化資料", colonyLeftRowsTextRect(true), colPanelLX, colPanelLY, colPanelLW, colPanelLH)
	inside("未同化", colonyAssimilationTextRect(), colPanelLX, colPanelLY, colPanelLW, colPanelLH)
	inside("人口", colonyPopulationTextRect(), colPanelLX, colPanelLY, colPanelLW, colPanelLH)
	if got, want := colonyLeftRowsTextRect(true).y+colonyLeftRowsTextRect(true).h, colonyAssimilationTextRect().y; got > want {
		t.Fatalf("左欄資料底部 %d 壓到未同化列 %d", got, want)
	}
	if got, want := colonyAssimilationTextRect().y+colonyAssimilationTextRect().h, colonyPopulationTextRect().y; got > want {
		t.Fatalf("未同化列底部 %d 壓到人口列 %d", got, want)
	}

	for row := 0; row < 2; row++ {
		for column := 0; column < 2; column++ {
			inside("產出", colonyOutputTextRect(column, row), colPanelMX, colPanelMY, colPanelMW, colPanelMH)
		}
	}
	if got, want := colonyOutputTextRect(0, 0).y, 37; got != want {
		t.Fatalf("產出 row0 起點 y=%d，want 可視 cell y=%d", got, want)
	}
	if got, want := colonyOutputTextRect(0, 1).y, 59; got != want {
		t.Fatalf("產出 row1 起點 y=%d，want 分隔線下方 cell y=%d", got, want)
	}
	for row := 0; row < 2; row++ {
		if got, want := colonyOutputTextRect(0, row).h, 16; got != want {
			t.Fatalf("產出 row%d 高度=%d，want 能容納 runtime bitmap glyph 的 16px", row, got)
		}
	}
	if got := colonyOutputTextRect(0, 0).y + colonyOutputTextRect(0, 0).h; got > 58 {
		t.Fatalf("產出 row0 底端 y=%d 侵入水平分隔線 y=58", got)
	}
	if got := colonyOutputTextRect(0, 1).y; got <= 58 {
		t.Fatalf("產出 row1 起點 y=%d 未留在水平分隔線下方", got)
	}
	if got, want := colonyOutputTextRect(1, 1).y+colonyOutputTextRect(1, 1).h, colonyBuildingsTitleTextRect().y; got > want {
		t.Fatalf("產出 row1 底端 y=%d 壓到 Buildings 標題 y=%d", got, want)
	}

	// 以實際 bitmap 字型量測繁中與英文長值；寬度與高度都必須在各自
	// 可視列矩形內，避免只以 maxWidth 讓字仍跨進分隔線或標題列。
	fnt := uifont.LoadBitmapTC()
	for _, tc := range []struct {
		name string
		r    textSafeRect
		text string
	}{
		{"繁中 row0", colonyOutputTextRect(0, 0), "食物盈虧 %+d／每回合產出"},
		{"英文 row1", colonyOutputTextRect(1, 1), "Research output +999"},
	} {
		for _, line := range tc.r.lines(fnt, tc.text, 11) {
			w, h := fnt.Measure(line, 11)
			if w > tc.r.contentWidth() || h > float64(tc.r.h-2*tc.r.insetY) {
				t.Fatalf("%s 字墨尺寸 %.0fx%.0f 超出安全列 %dx%d：%q", tc.name, w, h,
					tc.r.w-2*tc.r.insetX, tc.r.h-2*tc.r.insetY, line)
			}
		}
	}
	inside("建築標題", colonyBuildingsTitleTextRect(), colPanelMX, colPanelMY, colPanelMW, colPanelMH)
	inside("建築清單", colonyBuildingsListTextRect(), colPanelMX, colPanelMY, colPanelMW, colPanelMH)

	for row := 0; row < 3; row++ {
		for field := 0; field < 3; field++ {
			inside("職業列", colonyJobTextRect(row, field), colJobX0, colJobY0+row*colJobStep, colJobX1-colJobX0, colJobStep-4)
		}
	}
}
