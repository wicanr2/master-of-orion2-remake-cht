package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestTotalCommandPointsSupply 驗證「帝國基礎值 + 逐殖民地星基/戰鬥站/星辰要塞供給」的
// 指揮評等總和(GAME_MANUAL.pdf p.79/82/83 為軌道衛星供給出處;基礎值 5 是用真實存檔
// SAVE10.GAM oracle 反推,見 gamedata.CommandPointsBase 註解的完整推導與 flat-vs-per-colony
// 不確定性 TODO)。跨殖民地可疊加(同殖民地內三者取代關係已在 gamedata.CommandPointsFromBuildings
// 驗證過,這裡驗證的是「跨殖民地加總」+ 基礎值只加一次)。
func TestTotalCommandPointsSupply(t *testing.T) {
	s := NewDemoSession()
	s.ColonyBuildings = []map[string]bool{
		{"星基": true},  // +1
		{"戰鬥站": true}, // +2(不同殖民地,各自獨立不受「取代」影響)
		nil,           // 尚無建築,+0
		{"星辰要塞": true, "海軍陸戰隊營": true}, // +3(海軍陸戰隊營與指揮評等無關,不影響)
	}
	// 2026-07-11:基礎值 5(gamedata.CommandPointsBase)只加一次(per-empire,不逐殖民地),
	// 故 want = 5(基礎) + 1+2+0+3(建築) = 11。
	if got := s.totalCommandPointsSupply(); got != 11 {
		t.Errorf("totalCommandPointsSupply=%d, want 11(基礎5+建築1+2+0+3)", got)
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
// (gamedata.IncomeCommandOverflowCost,經 engine.RunEmpireTurn)。
//
// 2026-07-11 改版(修 CommandPointsBase regression 後重新設計):修復前,開局母星只有星基
// (+1 供給)缺基礎值,3 艘開局艦艇(殖民船+2偵察艦,需求 3)就超支 2 點 → 每回合 -20 BC 死亡
// 螺旋(見 docs/HONEST-STATUS.md 與 gamedata.CommandPointsBase 註解的 oracle 反推)。修復後
// 供給 = 基礎 5 + 星基 1 = 6 ≥ 3,開局本身不再超支——這正是本次修復要達成的行為,故第一個
// 子測試改成「驗證開局不再超支」;要驗證超支路徑仍正確運作,後面子測試改用外加艦隊把需求推到
// 明確超過新基準供給的水準。
func TestEndTurnCommandOverflowPenalty(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true // 隔離隨機事件,避免干擾這回合的 BC 精確斷言
	s.EndTurn()
	if s.LastPlayerOutput.CommandOverflowCost != 0 {
		t.Errorf("開局供給=基礎5+星基1=6、需求3,不應超支,CommandOverflowCost=%d, want 0", s.LastPlayerOutput.CommandOverflowCost)
	}

	// s2:外加 6 艘護衛艦(各 1 點),需求從 3 推到 9;供給仍是基礎5+星基1=6 → 缺口 3 → 30 BC。
	s2 := NewDemoSession()
	s2.DisableEvents = true
	for i := 0; i < 6; i++ {
		s2.Ships = append(s2.Ships, Ship{Name: "額外護衛艦", Class: "護衛艦"})
	}
	s2.EndTurn()
	if s2.LastPlayerOutput.CommandOverflowCost != 30 {
		t.Errorf("供給6、需求9(3+6艘護衛艦),CommandOverflowCost=%d, want 30", s2.LastPlayerOutput.CommandOverflowCost)
	}

	// s3:同 s2 的艦隊規模(需求9),但把母星星基升級成戰鬥站(+2,取代星基後供給=基礎5+2=7),
	// 缺口縮小為 2 → 20 BC(仍超支,驗證建築升級確實反映到供給,但沒補到足以歸零)。
	s3 := NewDemoSession()
	s3.DisableEvents = true
	for i := 0; i < 6; i++ {
		s3.Ships = append(s3.Ships, Ship{Name: "額外護衛艦", Class: "護衛艦"})
	}
	s3.ColonyBuildings[0]["星基"] = false
	s3.ColonyBuildings[0]["戰鬥站"] = true
	s3.EndTurn()
	if s3.LastPlayerOutput.CommandOverflowCost != 20 {
		t.Errorf("供給7(基礎5+戰鬥站2)、需求9,CommandOverflowCost=%d, want 20", s3.LastPlayerOutput.CommandOverflowCost)
	}

	// s4:同艦隊規模(需求9),母星升級成星辰要塞(+3,取代戰鬥站/星基後供給=基礎5+3=8),
	// 缺口縮小為 1 → 10 BC。
	s4 := NewDemoSession()
	s4.DisableEvents = true
	for i := 0; i < 6; i++ {
		s4.Ships = append(s4.Ships, Ship{Name: "額外護衛艦", Class: "護衛艦"})
	}
	s4.ColonyBuildings[0]["星基"] = false
	s4.ColonyBuildings[0]["星辰要塞"] = true
	s4.EndTurn()
	if s4.LastPlayerOutput.CommandOverflowCost != 10 {
		t.Errorf("供給8(基礎5+星辰要塞3)、需求9,CommandOverflowCost=%d, want 10", s4.LastPlayerOutput.CommandOverflowCost)
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
