package main

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestHotseatEmpireSelectDefaultsAndTargetsExactAIIndices(t *testing.T) {
	b := &sceneBuilder{session: shell.NewDemoSession(), pendingHotseat: 3, lang: i18n.Traditional}
	screen, err := b.hotseatEmpireSelect()
	if err != nil {
		t.Fatalf("建立選帝國畫面失敗: %v", err)
	}
	s := screen.(*hotseatEmpireSelectScreen)
	if got := s.selectedIndices(); len(got) != 2 || got[0] != 0 || got[1] != 1 {
		t.Fatalf("預設應選前兩個 AI,got %v", got)
	}

	// 取消第 0 個,再選第 2 個;選定數量維持兩席,而不是被迫接管尾端 AI。
	s.update(shell.InputState{MouseX: hseListX + 4, MouseY: hseListY + 4, ClickReleased: true})
	s.update(shell.InputState{MouseX: hseListX + 4, MouseY: hseListY + 2*hseRowH + 4, ClickReleased: true})
	got := s.selectedIndices()
	if len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("指定後應為 AI 索引 [1 2],got %v", got)
	}
}

func TestHotseatEmpireSelectDraws(t *testing.T) {
	b := &sceneBuilder{session: shell.NewDemoSession(), pendingHotseat: 2, lang: i18n.Traditional}
	screen, err := b.hotseatEmpireSelect()
	if err != nil {
		t.Fatalf("建立選帝國畫面失敗: %v", err)
	}
	dst := ebiten.NewImage(640, 480)
	screen.draw(dst)
}
