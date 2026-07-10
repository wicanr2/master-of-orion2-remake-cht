package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestTotalCommandPointsSupply 驗證逐殖民地加總星基/戰鬥站/星辰要塞供給的指揮評等
// (GAME_MANUAL.pdf p.79/82/83),跨殖民地可疊加(同殖民地內三者取代關係已在
// gamedata.CommandPointsFromBuildings 驗證過,這裡驗證的是「跨殖民地加總」)。
func TestTotalCommandPointsSupply(t *testing.T) {
	s := NewDemoSession()
	s.ColonyBuildings = []map[string]bool{
		{"星基": true},  // +1
		{"戰鬥站": true}, // +2(不同殖民地,各自獨立不受「取代」影響)
		nil,           // 尚無建築,+0
		{"星辰要塞": true, "海軍陸戰隊營": true}, // +3(海軍陸戰隊營與指揮評等無關,不影響)
	}
	if got := s.totalCommandPointsSupply(); got != 6 { // 1+2+0+3
		t.Errorf("totalCommandPointsSupply=%d, want 6", got)
	}
}

// TestUsedCommandPoints 驗證逐艦加總指揮評等需求(GAME_MANUAL.pdf p.169 size class 公式)。
func TestUsedCommandPoints(t *testing.T) {
	s := NewDemoSession()
	s.Ships = []Ship{
		{Name: "A", Class: "殖民船"}, // 手冊 p.84:Colony Ship 明文 1 點
		{Name: "B", Class: "護衛艦"}, // Frigate = 1
		{Name: "C", Class: "驅逐艦"}, // Destroyer = 2
		{Name: "D", Class: "泰坦"},  // Titan = 5(p.83 具體數字交叉驗證)
	}
	if got := s.usedCommandPoints(); got != 9 { // 1+1+2+5
		t.Errorf("usedCommandPoints=%d, want 9", got)
	}
}

// TestUsedCommandPointsEmptyFleet 邊界:無艦隊時需求為 0。
func TestUsedCommandPointsEmptyFleet(t *testing.T) {
	s := NewDemoSession()
	s.Ships = nil
	if got := s.usedCommandPoints(); got != 0 {
		t.Errorf("usedCommandPoints(無艦隊)=%d, want 0", got)
	}
}

// TestEndTurnCommandOverflowPenalty 驗證 EndTurn 實際把指揮評等超支懲罰接進國庫結算
// (gamedata.IncomeCommandOverflowCost,經 engine.RunEmpireTurn)。開局母星只有 1 座星基
// (+1 供給),但有 3 艘開局艦艇(殖民船+2 偵察艦,各 1 點=3 點需求),缺口 2 點 → 20 BC/回合。
// 補上戰鬥站(+2,取代星基後仍是 2,supply=need=3)後懲罰應歸零。
func TestEndTurnCommandOverflowPenalty(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true // 隔離隨機事件,避免干擾這回合的 BC 精確斷言
	s.EndTurn()
	if s.LastPlayerOutput.CommandOverflowCost != 20 {
		t.Errorf("開局供給 1、需求 3,CommandOverflowCost=%d, want 20", s.LastPlayerOutput.CommandOverflowCost)
	}

	s2 := NewDemoSession()
	s2.DisableEvents = true
	s2.ColonyBuildings[0]["戰鬥站"] = true // 供給補到 2,需求仍 3(尚未 3=3,verify 未歸零)
	s2.EndTurn()
	if s2.LastPlayerOutput.CommandOverflowCost != 10 {
		t.Errorf("供給補到 2、需求 3,CommandOverflowCost=%d, want 10", s2.LastPlayerOutput.CommandOverflowCost)
	}

	s3 := NewDemoSession()
	s3.DisableEvents = true
	s3.ColonyBuildings[0]["星辰要塞"] = true // 供給補到 3(取代戰鬥站/星基),等於需求 3
	s3.EndTurn()
	if s3.LastPlayerOutput.CommandOverflowCost != 0 {
		t.Errorf("供給=需求時 CommandOverflowCost=%d, want 0", s3.LastPlayerOutput.CommandOverflowCost)
	}
}

// TestUsedCommandPointsUsesGamedataTable 驗證 usedCommandPoints 實際查的是
// gamedata.ShipCommandCost(而非本檔另建一份數字),避免兩處數字未來各自漂移。
func TestUsedCommandPointsUsesGamedataTable(t *testing.T) {
	s := NewDemoSession()
	s.Ships = []Ship{{Name: "X", Class: "末日之星"}}
	want := gamedata.ShipCommandCost(gamedata.SHIP_DOOMSTAR)
	if got := s.usedCommandPoints(); got != want {
		t.Errorf("usedCommandPoints(末日之星)=%d, want %d(gamedata.ShipCommandCost)", got, want)
	}
}
