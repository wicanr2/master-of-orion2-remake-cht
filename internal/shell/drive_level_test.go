package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 戰鬥速度表的三個自我驗證規律——抄歪一格會壞掉其中至少一個。
func TestCombatSpeedTableIsSelfConsistent(t *testing.T) {
	classes := []gamedata.CombatShipClass{0, 1, 2, 3, 4, 5}
	for lvl := 1; lvl <= gamedata.CombatSpeedDriveLevels; lvl++ {
		prev := 1 << 30
		for _, c := range classes {
			min, max, ok := gamedata.CombatSpeedRange(lvl, c)
			if !ok {
				t.Fatalf("引擎階 %d 艦體 %d 應查得到", lvl, c)
			}
			// ① 最大速恆等於最小速 +10
			if max-min != gamedata.CombatSpeedMaxOverMin {
				t.Errorf("階%d 艦體%d:max−min 應為 %d,得到 %d",
					lvl, c, gamedata.CombatSpeedMaxOverMin, max-min)
			}
			// ② 艦體越大越慢
			if min >= prev {
				t.Errorf("階%d 艦體%d 應比前一級慢:%d vs %d", lvl, c, min, prev)
			}
			prev = min
			// ③ 每升一階 +2
			if lvl > 1 {
				pmin, _, _ := gamedata.CombatSpeedRange(lvl-1, c)
				if min-pmin != gamedata.CombatSpeedPerDriveLevel {
					t.Errorf("階%d 艦體%d 應比前一階快 %d:%+d",
						lvl, c, gamedata.CombatSpeedPerDriveLevel, min-pmin)
				}
			}
		}
	}
}

// 表的第一列必須逐格等於執行檔抽出來的那六個數字。
//
// 上面那條測的是「規律成立」,這條測的是「起點正確」——規律對但起點錯的話上面全綠。
func TestCombatSpeedFirstRowMatchesTheExecutable(t *testing.T) {
	want := [6]int{10, 8, 6, 5, 4, 3} // byte_17FE90 第一列(引擎階 1)
	for c, w := range want {
		min, _, ok := gamedata.CombatSpeedRange(1, gamedata.CombatShipClass(c))
		if !ok || min != w {
			t.Errorf("引擎階 1 艦體 %d 的最小速應為 %d,得到 %d(ok=%v)", c, w, min, ok)
		}
	}
}

// 超界一律回 ok=false,不 panic。
func TestCombatSpeedOutOfRangeIsSafe(t *testing.T) {
	for _, c := range []struct {
		lvl   int
		class gamedata.CombatShipClass
	}{{0, 0}, {7, 0}, {1, 6}, {1, -1}, {-1, 0}} {
		if _, _, ok := gamedata.CombatSpeedRange(c.lvl, c.class); ok {
			t.Errorf("階%d 艦體%d 應回 false", c.lvl, c.class)
		}
	}
	if got := gamedata.ShipCombatSpeed(99, 0, false, false); got != 0 {
		t.Errorf("超界應回 0,得到 %d", got)
	}
}

// 主動權用手冊公式,而且「小船先動」這句話真的成立。
func TestInitiativePutsSmallShipsFirst(t *testing.T) {
	if got := gamedata.CombatInitiative(30, 20); got != 30+10*20 {
		t.Errorf("主動權公式應為 BA + 10×速度,得到 %d", got)
	}
	// 手冊:「Thus, smaller ships should move before bigger, slower ones.」
	// 同一階引擎下,巡防艦(速快、攻低)應排在末日之星(速慢、攻高)前面。
	small := gamedata.CombatInitiative(4, gamedata.ShipCombatSpeed(1, 0, false, false))
	big := gamedata.CombatInitiative(64, gamedata.ShipCombatSpeed(1, 5, false, false))
	if small <= big {
		t.Errorf("小船的主動權應較高(手冊明說):巡防艦 %d vs 末日之星 %d", small, big)
	}
}

// 引擎階:沒研究任何進階引擎是 1;研究了就往上跳。
func TestDriveLevelFollowsResearch(t *testing.T) {
	s := NewDemoSession()
	if got := s.driveLevel(); got != 1 {
		t.Errorf("開局應為第 1 階,得到 %d", got)
	}
	for lvl := 2; lvl <= gamedata.CombatSpeedDriveLevels; lvl++ {
		ns := NewDemoSession()
		ns.Player = withTech(t, ns.Player, gamedata.DriveTechForLevel(lvl))
		if got := ns.driveLevel(); got != lvl {
			t.Errorf("研究出 %s 後應為第 %d 階,得到 %d",
				gamedata.TechnologyName(gamedata.DriveTechForLevel(lvl)), lvl, got)
		}
	}
	// 六個引擎科技與階數雙向對得上。
	for lvl := 1; lvl <= gamedata.CombatSpeedDriveLevels; lvl++ {
		tech := gamedata.DriveTechForLevel(lvl)
		if got := gamedata.DriveLevelForTech(tech); got != lvl {
			t.Errorf("%v 的階數應為 %d,得到 %d", tech, lvl, got)
		}
	}
	if gamedata.DriveLevelForTech(gamedata.TECH_LASER_CANNON) != 0 {
		t.Error("不是引擎的科技應回 0")
	}
}

// 增強引擎 +5、跨維度種族 +4,兩者可疊加。
func TestCombatSpeedBonusesStack(t *testing.T) {
	base := gamedata.ShipCombatSpeed(1, 3, false, false)
	if got := gamedata.ShipCombatSpeed(1, 3, true, false); got != base+gamedata.CombatSpeedAugmentedEngines {
		t.Errorf("增強引擎應 +%d,得到 %+d", gamedata.CombatSpeedAugmentedEngines, got-base)
	}
	if got := gamedata.ShipCombatSpeed(1, 3, false, true); got != base+gamedata.CombatSpeedTransDimensional {
		t.Errorf("跨維度應 +%d,得到 %+d", gamedata.CombatSpeedTransDimensional, got-base)
	}
	if got := gamedata.ShipCombatSpeed(1, 3, true, true); got != base+
		gamedata.CombatSpeedAugmentedEngines+gamedata.CombatSpeedTransDimensional {
		t.Errorf("兩者應可疊加,得到 %+d", got-base)
	}
}

// 崔拉里安(跨維度)的船真的比較快——第 129 項挖出的種族特性在這裡有了第二個消費端。
func TestTransDimensionalRaceShipsAreFaster(t *testing.T) {
	human := NewDemoSession()
	human.ApplyRace(raceIndexByEnName(t, "Humans"))
	tri := NewDemoSession()
	tri.ApplyRace(raceIndexByEnName(t, "Trilarians"))
	sh := Ship{Name: "測試艦", Class: "戰艦"}
	if got := tri.shipCombatSpeed(sh) - human.shipCombatSpeed(sh); got != gamedata.CombatSpeedTransDimensional {
		t.Errorf("崔拉里安應快 %d,實快 %d", gamedata.CombatSpeedTransDimensional, got)
	}
}

// 齊射真的照主動權排序(先前是艦隊清單順序 = 先造的先打)。
func TestVolleyOrdersByInitiative(t *testing.T) {
	cs := []combatant{
		{initiative: 10, atk: 1, hp: 1, kind: WeaponKindBeam},
		{initiative: 90, atk: 1, hp: 1, kind: WeaponKindBeam},
		{initiative: 50, atk: 1, hp: 1, kind: WeaponKindBeam},
	}
	sortByInitiative(cs)
	for i := 1; i < len(cs); i++ {
		if cs[i-1].initiative < cs[i].initiative {
			t.Fatalf("應由高到低:%v", []int{cs[0].initiative, cs[1].initiative, cs[2].initiative})
		}
	}
	// 主動權相同要維持原序(穩定排序)——否則同一場戰鬥可能打出不同結果。
	tied := []combatant{{initiative: 5, shipIdx: 7}, {initiative: 5, shipIdx: 3}, {initiative: 5, shipIdx: 9}}
	sortByInitiative(tied)
	if tied[0].shipIdx != 7 || tied[1].shipIdx != 3 || tied[2].shipIdx != 9 {
		t.Errorf("同分時應維持原序,得到 %d/%d/%d", tied[0].shipIdx, tied[1].shipIdx, tied[2].shipIdx)
	}
}

// 裝甲級數:鈦 = 0,逐級 +1(戰機 HP 用)。
func TestArmorLevelAboveTitanium(t *testing.T) {
	want := map[string]int{"鈦裝甲": 0, "三鈦裝甲": 1, "佐特裝甲": 2, "中子素裝甲": 3, "精金裝甲": 4, "氙素裝甲": 5}
	for name, lvl := range want {
		if got := armorLevelAboveTitanium(name); got != lvl {
			t.Errorf("%s 應為第 %d 級,得到 %d", name, lvl, got)
		}
	}
	if got := armorLevelAboveTitanium("無裝甲"); got != 0 {
		t.Errorf("無裝甲應為 0,得到 %d", got)
	}
}

// 增強引擎的主題對得上執行檔(我一開始猜的是進階建設,執行檔說是進階融合)。
func TestAugmentedEnginesHasItsRealTopic(t *testing.T) {
	var found *Component
	for i := range SpecialOptions {
		if SpecialOptions[i].Name == augmentedEnginesName {
			found = &SpecialOptions[i]
		}
	}
	if found == nil {
		t.Fatal("增強引擎應出現在 SpecialOptions")
	}
	topic, ok := gamedata.OrigTechTopic(gamedata.TECH_AUGMENTED_ENGINES)
	if !ok {
		t.Fatal("執行檔應查得到增強引擎的主題")
	}
	if found.Tech != topic {
		t.Errorf("主題應為執行檔的 %v,得到 %v", topic, found.Tech)
	}
}

// 棋盤比例尺:原版 81 欄 vs remake 8 欄,速度 13..30 應換算成 1..3 格。
func TestTacticalMoveSquaresScaleIsSane(t *testing.T) {
	if got := TacticalMoveSquares(0); got != 0 {
		t.Errorf("沒有引擎不該能動,得到 %d", got)
	}
	// 最慢(階1 末日之星 max=13)與最快(階6 巡防艦 max=30 + 增強 + 跨維)
	slow := gamedata.ShipCombatSpeed(1, 5, false, false)
	fast := gamedata.ShipCombatSpeed(gamedata.CombatSpeedDriveLevels, 0, true, true)
	ns, nf := TacticalMoveSquares(slow), TacticalMoveSquares(fast)
	if ns < 1 {
		t.Errorf("最慢的船也要能動至少 1 格,得到 %d(速度 %d)", ns, slow)
	}
	if nf <= ns {
		t.Errorf("最快的船應走得比最慢的遠:%d vs %d", nf, ns)
	}
	// 不該有任何組合可以一步橫跨全場——那正是先前「瞬移」的行為。
	if nf >= TacticalGridColumns {
		t.Errorf("最快的船 %d 格已可橫跨 %d 欄的棋盤", nf, TacticalGridColumns)
	}
	// 比例尺本身:用一手的原版欄數,不是隨手挑的數字。
	if gamedata.CombatGridColumns != 81 || gamedata.CombatGridRows != 68 {
		t.Errorf("原版棋盤應為 81×68(Assign_Combat_Grids_ 的迴圈界限),得到 %d×%d",
			gamedata.CombatGridColumns, gamedata.CombatGridRows)
	}
}
