package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// commandpoints_live_test.go:指揮點數「現算 vs 快取」的護欄。
//
// ⚠ 這條擋的是一個**畫面上自相矛盾但不會報錯**的狀態:`Player.CommandPointsSupply` 是
// 只在 `EndTurn` 更新的快取欄位,開局或剛蓋好星基還沒結算時是舊值。指揮點數視窗第一版
// 直接讀它,結果畫出「起始 5 + 軌道基地 0,總計卻是 1」——三個數字擺在一起自打嘴巴。

// TestCommandPointsSupplyNowIsLive:現算值不依賴 EndTurn 跑過。
func TestCommandPointsSupplyNowIsLive(t *testing.T) {
	s := NewDemoSession()
	// 刻意把快取欄位設成明顯錯誤的值,證明現算不看它。
	s.Player.CommandPointsSupply = -999
	s.Player.UsedCommandPoints = -999

	want := gamedata.CommandPointsBase
	for _, built := range s.ColonyBuildings {
		want += gamedata.CommandPointsFromBuildings(built)
	}
	if got := s.CommandPointsSupplyNow(); got != want {
		t.Errorf("CommandPointsSupplyNow = %d,want %d(基礎 %d + 建築)",
			got, want, gamedata.CommandPointsBase)
	}
	if got := s.CommandPointsUsedNow(); got < 0 {
		t.Errorf("CommandPointsUsedNow = %d,不該是負的(讀到快取的 -999 了)", got)
	}
}

// TestCommandPointsSupplyNowFollowsNewOrbitalBase:剛蓋好星基**當下**就要反映,
// 不必等回合結束。玩家蓋完看到數字沒動會以為建築沒生效。
func TestCommandPointsSupplyNowFollowsNewOrbitalBase(t *testing.T) {
	s := NewDemoSession()
	if len(s.ColonyBuildings) == 0 {
		s.ColonyBuildings = append(s.ColonyBuildings, map[string]bool{})
	}
	if s.ColonyBuildings[0] == nil {
		s.ColonyBuildings[0] = map[string]bool{}
	}
	for _, n := range []string{"星基", "戰鬥站", "星辰要塞"} {
		delete(s.ColonyBuildings[0], n)
	}
	before := s.CommandPointsSupplyNow()
	s.ColonyBuildings[0]["星基"] = true
	after := s.CommandPointsSupplyNow()
	if after <= before {
		t.Errorf("蓋好星基後供給沒變(%d → %d)——現算值沒吃到新建築", before, after)
	}
}
