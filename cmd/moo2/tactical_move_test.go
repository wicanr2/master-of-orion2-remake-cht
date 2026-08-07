package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// 棋盤尺寸常數兩邊必須一致。
//
// `shell.TacticalMoveSquares` 用 shell 這一側的常數算比例尺,而畫面用 cmd 這一側的
// `gcCols`/`gcRows`。兩邊各寫一份,不釘住的話改了其中一個會讓移動格數安靜地算錯。
func TestGridConstantsAgreeAcrossPackages(t *testing.T) {
	if gcCols != shell.TacticalGridColumns {
		t.Errorf("欄數對不上:cmd=%d shell=%d", gcCols, shell.TacticalGridColumns)
	}
	if gcRows != shell.TacticalGridRows {
		t.Errorf("列數對不上:cmd=%d shell=%d", gcRows, shell.TacticalGridRows)
	}
}

// 移動力預算逐艦算,而且與艦的戰鬥速度同向。
func TestFreshMoveBudgetsFollowCombatSpeed(t *testing.T) {
	ships := []shell.CombatShip{
		{Name: "快", CombatSpeed: 30},
		{Name: "慢", CombatSpeed: 13},
		{Name: "沒引擎", CombatSpeed: 0},
	}
	got := freshMoveBudgets(ships)
	if len(got) != len(ships) {
		t.Fatalf("長度應為 %d,得到 %d", len(ships), len(got))
	}
	if got[0] <= got[1] {
		t.Errorf("速度高的應走得遠:%d vs %d", got[0], got[1])
	}
	if got[1] < 1 {
		t.Errorf("再慢的船也要能動至少 1 格,得到 %d", got[1])
	}
	if got[2] != 0 {
		t.Errorf("沒引擎的船不該能動,得到 %d", got[2])
	}
	// 不該有任何船一步橫跨全場——那正是先前「瞬移」的行為。
	for i, n := range got {
		if n >= gcCols {
			t.Errorf("%s 的移動力 %d 已可橫跨整個棋盤(%d 欄)", ships[i].Name, n, gcCols)
		}
	}
}
