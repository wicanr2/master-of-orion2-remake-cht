package gamedata

import "testing"

// 手冊 p.121 那張表的四條軌逐格釘住。
// BA/BD 兩欄是既有的 formulas.go(從 openorion2 轉寫),這裡順帶用手冊再對一次
// ——兩個獨立來源給同一組數字。
func TestCrewBonusTracksMatchTheManualTable(t *testing.T) {
	for _, tc := range []struct{ level, ba, bd, me, bo int }{
		{CrewGreen, 0, 0, 0, 0},
		{CrewRegular, 15, 15, 7, 5},
		{CrewVeteran, 30, 30, 15, 10},
		{CrewElite, 50, 50, 25, 15},
		{CrewUltraElite, 75, 75, 37, 20},
	} {
		if got := ShipCrewOffenseBonus(tc.level); got != tc.ba {
			t.Errorf("等級 %d 的 BA 應 +%d,得到 +%d", tc.level, tc.ba, got)
		}
		if got := ShipCrewDefenseBonus(tc.level); got != tc.bd {
			t.Errorf("等級 %d 的 BD 應 +%d,得到 +%d", tc.level, tc.bd, got)
		}
		if got := ShipCrewMissileEvasionBonus(tc.level); got != tc.me {
			t.Errorf("等級 %d 的 ME 應 +%d,得到 +%d", tc.level, tc.me, got)
		}
		if got := ShipCrewBoardingBonus(tc.level); got != tc.bo {
			t.Errorf("等級 %d 的 Bo 應 +%d,得到 +%d", tc.level, tc.bo, got)
		}
	}
	// 超出範圍一律 0,不是回最後一格——那會讓資料錯誤變成免費的頂級加成。
	for _, l := range []int{-1, CrewLevelCount, 99} {
		if ShipCrewBoardingBonus(l) != 0 || ShipCrewMissileEvasionBonus(l) != 0 {
			t.Errorf("等級 %d 超出範圍應回 0", l)
		}
	}
}

// 統帥種族不是「升級快」,是**整條階梯往上平移一格**:出廠就是正規兵,
// 而且多一個一般種族碰不到的超級精銳。
func TestWarlordShiftsTheWholeLadderNotJustTheSpeed(t *testing.T) {
	if CrewStartingLevel(false) != CrewGreen {
		t.Error("一般種族應從新兵開始")
	}
	if CrewStartingLevel(true) != CrewRegular {
		t.Error("統帥種族應從正規兵開始(手冊 unless they're … Warlord)")
	}
	if CrewMaxLevel(false) != CrewElite {
		t.Error("一般種族的頂是精銳——超級精銳有星號註明只有統帥種族到得了")
	}
	if CrewMaxLevel(true) != CrewUltraElite {
		t.Error("統帥種族應到得了超級精銳")
	}
	// 同樣的經驗,統帥種族高一級。
	for _, xp := range []int{0, 50, 150, 500} {
		n, w := CrewLevelForXP(xp, false), CrewLevelForXP(xp, true)
		if w != n+1 {
			t.Errorf("經驗 %d:一般 %d、統帥 %d ——應正好差一級", xp, n, w)
		}
	}
}

// EP 門檻:0 / 50 / 150 / 500。
func TestCrewLevelForXPMatchesTheThresholds(t *testing.T) {
	for _, tc := range []struct{ xp, want int }{
		{0, CrewGreen}, {49, CrewGreen},
		{50, CrewRegular}, {149, CrewRegular},
		{150, CrewVeteran}, {499, CrewVeteran},
		{500, CrewElite}, {100000, CrewElite}, // 一般種族封頂在精銳
	} {
		if got := CrewLevelForXP(tc.xp, false); got != tc.want {
			t.Errorf("經驗 %d 應是等級 %d,得到 %d", tc.xp, tc.want, got)
		}
	}
	// 統帥種族在 500 才碰得到超級精銳。
	if got := CrewLevelForXP(499, true); got != CrewElite {
		t.Errorf("統帥種族 499 經驗應是精銳,得到 %d", got)
	}
	if got := CrewLevelForXP(500, true); got != CrewUltraElite {
		t.Errorf("統帥種族 500 經驗應是超級精銳,得到 %d", got)
	}
}

func TestCrewXPAfterTurnGainUsesOriginalFiveHundredCap(t *testing.T) {
	for _, tc := range []struct{ current, gain, want int }{
		{499, 1, 500},
		{500, 1, 500},
		{490, 25, 500},
		{-10, 1, 1},
		{100, -3, 100},
	} {
		if got := CrewXPAfterTurnGain(tc.current, tc.gain); got != tc.want {
			t.Errorf("CrewXPAfterTurnGain(%d,%d)=%d，want %d", tc.current, tc.gain, got, tc.want)
		}
	}
}

// 距離下一級的進度,給 UI 用。
func TestCrewXPToNextLevel(t *testing.T) {
	if got := CrewXPToNextLevel(0, false); got != 50 {
		t.Errorf("新兵距正規兵應差 50,得到 %d", got)
	}
	if got := CrewXPToNextLevel(120, false); got != 30 {
		t.Errorf("120 經驗距老兵(150)應差 30,得到 %d", got)
	}
	if got := CrewXPToNextLevel(500, false); got != 0 {
		t.Errorf("一般種族到頂應回 0,得到 %d", got)
	}
	if got := CrewXPToNextLevel(500, true); got != 0 {
		t.Errorf("統帥種族到頂應回 0,得到 %d", got)
	}
}

// 戰鬥經驗:折半、捨去、最少 1。
func TestCrewBattleXPHalvesTheSumOfDestroyedSizeClasses(t *testing.T) {
	// 擊沉一艘巡防艦(1):1/2 = 0 → 保底 1。
	if got := CrewBattleXP([]int{1}); got != 1 {
		t.Errorf("擊沉一艘巡防艦應是保底的 1,得到 %d", got)
	}
	// 一艘泰坦(5)+ 一艘巡洋艦(3)= 8,折半 4。
	if got := CrewBattleXP([]int{5, 3}); got != 4 {
		t.Errorf("5+3 折半應是 4,得到 %d", got)
	}
	// 奇數要捨去:3+3+1 = 7,折半 3。
	if got := CrewBattleXP([]int{3, 3, 1}); got != 3 {
		t.Errorf("7 折半捨去應是 3,得到 %d", got)
	}
	// 六艘末日之星(6)= 36,折半 18。
	if got := CrewBattleXP([]int{6, 6, 6, 6, 6, 6}); got != 18 {
		t.Errorf("36 折半應是 18,得到 %d", got)
	}
}

// sub_4B184 先除二再強制最少 1，因此勝方即使零擊沉仍取得 1。
func TestCrewBattleXPUsesOriginalMinimumWhenNothingWasDestroyed(t *testing.T) {
	if got := CrewBattleXP(nil); got != 1 {
		t.Errorf("沒擊沉任何船的勝方 recipient 應得 1,得到 %d", got)
	}
	if got := CrewBattleXP([]int{}); got != 1 {
		t.Errorf("空清單應套原版最少 1,得到 %d", got)
	}
	if got := CrewBattleXP([]int{0, 0}); got != 1 {
		t.Errorf("全是 0 的艦體等級應套原版最少 1,得到 %d", got)
	}
	for _, tc := range []struct{ sum, want int }{{0, 1}, {1, 1}, {2, 1}, {3, 1}, {4, 2}, {12, 6}} {
		if got := CrewBattleXPFromDestroyedHullClassSum(tc.sum); got != tc.want {
			t.Errorf("raw hull class sum %d: got %d, want %d", tc.sum, got, tc.want)
		}
	}
}
