package main

import "testing"

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
	inside("建築標題", colonyBuildingsTitleTextRect(), colPanelMX, colPanelMY, colPanelMW, colPanelMH)
	inside("建築清單", colonyBuildingsListTextRect(), colPanelMX, colPanelMY, colPanelMW, colPanelMH)

	for row := 0; row < 3; row++ {
		for field := 0; field < 3; field++ {
			inside("職業列", colonyJobTextRect(row, field), colJobX0, colJobY0+row*colJobStep, colJobX1-colJobX0, colJobStep-4)
		}
	}
}
