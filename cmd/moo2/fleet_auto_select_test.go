package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func autoSelectTestBuilder(on bool, shipCounts ...int) *sceneBuilder {
	s := shell.NewDemoSession()
	s.Fleets = make([]shell.Fleet, len(shipCounts))
	for fi, count := range shipCounts {
		s.Fleets[fi] = shell.NewFleet(0)
		for si := 0; si < count; si++ {
			s.Fleets[fi].Ships = append(s.Fleets[fi].Ships, shell.Ship{Name: "test"})
		}
	}
	s.SelectedFleet = 0
	settings := s.EffectiveGameSettings()
	settings.AutoSelectShips = on
	s.ApplyGameSettings(settings)
	return &sceneBuilder{session: s}
}

func TestAutoSelectShipsInitializesCurrentFleet(t *testing.T) {
	b := autoSelectTestBuilder(true, 3)
	b.initializeFleetShipSelection()
	if len(b.shipPick) != 3 {
		t.Fatalf("開啟設定應選取目前艦隊全部 3 艘，得到 %#v", b.shipPick)
	}
	for i := 0; i < 3; i++ {
		if !b.shipPick[i] {
			t.Errorf("艦艇 %d 未被初始化選取", i)
		}
	}
}

func TestAutoSelectShipsDisabledStartsEmpty(t *testing.T) {
	b := autoSelectTestBuilder(false, 3)
	b.initializeFleetShipSelection()
	if b.shipPick == nil || len(b.shipPick) != 0 {
		t.Fatalf("關閉設定應建立空但非 nil 的集合，得到 %#v", b.shipPick)
	}
}

func TestAutoSelectShipsDoesNotOverrideManualDeselect(t *testing.T) {
	b := autoSelectTestBuilder(true, 2)
	b.initializeFleetShipSelection()
	b.toggleSelectAllShips() // 已全選時 ALL 會全不選。
	if b.shipPick == nil || len(b.shipPick) != 0 {
		t.Fatalf("ALL 應留下玩家明確的空集合，得到 %#v", b.shipPick)
	}
	b.initializeFleetShipSelection() // 模擬同一畫面的重建。
	if len(b.shipPick) != 0 {
		t.Fatalf("重建畫面不得把玩家取消的選取加回來，得到 %#v", b.shipPick)
	}
}

func TestAutoSelectShipsRebuildsForNewFleet(t *testing.T) {
	b := autoSelectTestBuilder(true, 4, 1)
	b.initializeFleetShipSelection()
	b.session.SelectFleet(1)
	b.shipPick = nil // selfleet action 的契約。
	b.initializeFleetShipSelection()
	if len(b.shipPick) != 1 || !b.shipPick[0] {
		t.Fatalf("切到一艘船的新艦隊應只選 index 0，得到 %#v", b.shipPick)
	}
	for i := 1; i < 4; i++ {
		if b.shipPick[i] {
			t.Errorf("上一支艦隊的 index %d 殘留", i)
		}
	}
}

func TestAutoSelectShipsRebuildsAfterSplit(t *testing.T) {
	b := autoSelectTestBuilder(true, 3)
	b.initializeFleetShipSelection()
	if _, ok := b.session.SplitFleet(0, []int{1}); !ok {
		t.Fatal("測試前置：單獨拆出 index 1 應成功")
	}
	b.shipPick = nil // splitfleet action 在索引改變後的契約。
	b.initializeFleetShipSelection()
	if len(b.shipPick) != 2 || !b.shipPick[0] || !b.shipPick[1] {
		t.Fatalf("拆分後目前艦隊剩兩艘，應只重建 index 0..1，得到 %#v", b.shipPick)
	}
	if b.shipPick[2] {
		t.Error("拆分前的 index 2 不得殘留")
	}
}

func TestFleetSelectionMarksAreExternalAndFixedWidth(t *testing.T) {
	for _, lang := range []i18n.Lang{i18n.English, i18n.Traditional} {
		selected := uiText(lang, "fleet.selection.selected")
		unselected := uiText(lang, "fleet.selection.unselected")
		if selected == "fleet.selection.selected" || unselected == "fleet.selection.unselected" {
			t.Fatalf("語言 %d 的艦艇選取標記不得裸露 JSON key", lang)
		}
		if len(selected) != len(unselected) {
			t.Errorf("語言 %d 的選取兩態應同寬：%q / %q", lang, selected, unselected)
		}
	}
}
