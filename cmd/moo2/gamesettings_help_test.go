package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

func TestHelpExpandRectImmediateAndTenStepEndpoint(t *testing.T) {
	target := helpRect{x: 76, y: 145, w: 488, h: 170}
	if got := helpExpandRect(20, 30, target, 0, false); got != target {
		t.Fatalf("停用展開說明應立即顯示完整框：got=%+v want=%+v", got, target)
	}
	first := helpExpandRect(20, 30, target, 0, true)
	if first.w != 48 || first.h != 17 || first.x <= 20 || first.y <= 30 {
		t.Fatalf("第一步應由來源朝目的框展開十分之一：%+v", first)
	}
	previous := first
	for tick := 1; tick < 10; tick++ {
		got := helpExpandRect(20, 30, target, tick, true)
		if got.w < previous.w || got.h < previous.h {
			t.Fatalf("展開框不可縮回：tick=%d previous=%+v got=%+v", tick, previous, got)
		}
		previous = got
	}
	if previous != target {
		t.Fatalf("第十步應抵達完整目的框：got=%+v want=%+v", previous, target)
	}
}

func TestGameSettingsRightClickOpensHelpWithoutToggling(t *testing.T) {
	s := &gameSettingsScreen{settings: shell.DefaultGameSettings(), helpIndex: -1}
	x, y, _, _ := s.rowRect(3)
	before := s.settings.ExpandingHelp
	s.update(shell.InputState{MouseX: x + 2, MouseY: y + 2, RightClickReleased: true})
	if s.helpIndex != 3 || s.settings.ExpandingHelp != before {
		t.Fatalf("右鍵應開啟第 3 列說明且不切換設定：index=%d before=%v after=%v", s.helpIndex, before, s.settings.ExpandingHelp)
	}
}
