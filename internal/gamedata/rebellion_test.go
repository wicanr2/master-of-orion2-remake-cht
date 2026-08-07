package gamedata

import "testing"

// 基準:每一單位未同化人口 1%(原版 `imul edx, ecx, 0Ah` 配 `rand(1..1000)`)。
// 手冊只說「the more unassimilated aliens, the larger the chance」,沒給數字。
func TestRebellionBaseChanceIsOnePercentPerUnit(t *testing.T) {
	for _, units := range []int{1, 3, 8, 20} {
		got := RebellionChancePermille(units, 2, false, false, false, false)
		if want := units * 10; got != want {
			t.Errorf("%d 單位應是 %d‰,得到 %d‰", units, want, got)
		}
	}
	// 沒有未同化人口就不會叛亂(原版 `test ecx,ecx / jle 結束`)。
	if got := RebellionChancePermille(0, 2, false, false, false, false); got != 0 {
		t.Errorf("沒有未同化人口應是 0‰,得到 %d‰", got)
	}
}

// 手冊 p.165:異族管理中心**減半**、滅絕政策**加倍**。
func TestRebellionModifiersMatchTheManual(t *testing.T) {
	base := RebellionChancePermille(10, 2, false, false, false, false)
	if base != 100 {
		t.Fatalf("基準應是 100‰,得到 %d‰", base)
	}
	if got := RebellionChancePermille(10, 2, false, false, true, false); got != base/2 {
		t.Errorf("異族管理中心應減半到 %d‰,得到 %d‰", base/2, got)
	}
	if got := RebellionChancePermille(10, 2, false, false, false, true); got != base*2 {
		t.Errorf("滅絕政策應加倍到 %d‰,得到 %d‰", base*2, got)
	}
	// 兩個同時存在:原版是先減半再加倍,回到基準。
	if got := RebellionChancePermille(10, 2, false, false, true, true); got != base {
		t.Errorf("減半再加倍應回到 %d‰,得到 %d‰", base, got)
	}
}

// 順序有差:原版是**先減半、後加倍**。奇數上兩種順序結果不同,這一條把順序釘住。
func TestHalveBeforeDoubleNotTheOtherWayRound(t *testing.T) {
	// 3 單位 = 30‰。先減半(15)再加倍 = 30;先加倍(60)再減半 = 30——偶數看不出來。
	// 改用會產生奇數的難度修正:5 單位 = 50‰,難度 3 → +4 = 54;減半 27;加倍 54。
	// 反過來:加倍 108;減半 54。仍相同。真正分岔的是「減半後為奇數」的情形:
	// 難度 4 → 5*10+8 = 58;減半 29;加倍 58。反過來 116/58。整數除法在這裡剛好對稱,
	// 所以改測「只減半」時的向零取整。
	if got := RebellionChancePermille(5, 4, true, true, true, false); got != 29 {
		t.Errorf("5 單位 + 難度 4(+8)= 58‰,減半應是 29‰(向零取整),得到 %d‰", got)
	}
	// 奇數減半要向零取整,不是四捨五入(原版 `sub/sar 1`)。
	// 1 單位 + 難度 3(+4)= 14‰,減半 = 7‰。
	if got := RebellionChancePermille(1, 3, true, true, true, false); got != 7 {
		t.Errorf("14‰ 減半應是 7‰,得到 %d‰", got)
	}
}

// 難度修正只在「殖民地主人是人類、叛軍是 AI」時才加(原版兩個 [player+0x28]==100 的比較)。
func TestDifficultyAdjustOnlyAppliesToHumanOwnedColonies(t *testing.T) {
	base := RebellionChancePermille(10, 4, false, false, false, false)
	human := RebellionChancePermille(10, 4, true, true, false, false)
	if base != 100 {
		t.Fatalf("非人類殖民地不該吃難度修正,應是 100‰,得到 %d‰", base)
	}
	if want := 100 + RebellionDifficultyAdjust(4); human != want {
		t.Errorf("人類殖民地在難度 4 應是 %d‰,得到 %d‰", want, human)
	}
	// 「普通」(2)是基準,修正為 0——與 GroundDifficultyBonus 同一種寫法。
	if got := RebellionDifficultyAdjust(2); got != 0 {
		t.Errorf("難度 2 的修正應是 0,得到 %d", got)
	}
	// 低難度是**負的**:原版讓簡單模式的叛亂更少。
	if got := RebellionDifficultyAdjust(0); got >= 0 {
		t.Errorf("難度 0 的修正應為負,得到 %d", got)
	}
}

// 機率不會算成負數。難度 0 的修正是 −8,所以未同化人口少的時候整式會逼近 0。
func TestChanceNeverGoesNegative(t *testing.T) {
	// 1 單位 + 難度 0:10 − 8 = 2‰(還沒到負的)。
	if got := RebellionChancePermille(1, 0, true, true, false, false); got != 2 {
		t.Errorf("1 單位 + 難度 0 應是 2‰,得到 %d‰", got)
	}
	// 這一組才會讓原式變負:難度 0(−8)遇上 0 單位以外的最小值,再加上異族管理中心。
	// 不論怎麼組合,結果一律夾在 0 以上——負機率會讓 `roll <= chance` 的語意壞掉。
	for units := 1; units <= 3; units++ {
		for _, amc := range []bool{false, true} {
			for _, ext := range []bool{false, true} {
				if got := RebellionChancePermille(units, 0, true, true, amc, ext); got < 0 {
					t.Errorf("units=%d amc=%v ext=%v 算出負機率 %d‰", units, amc, ext, got)
				}
			}
		}
	}
}

// 擲骰是 1..1000,判定是 `roll <= chance`(原版 `cmp eax,edx / jg 不叛亂`——等於也算中)。
func TestRebellionTriggersIsInclusive(t *testing.T) {
	if !RebellionTriggers(100, 100) {
		t.Error("擲出剛好等於機率應觸發(原版用 jg,等於不跳走)")
	}
	if RebellionTriggers(100, 101) {
		t.Error("擲出大於機率不該觸發")
	}
	if !RebellionTriggers(100, 1) {
		t.Error("擲出 1 應觸發")
	}
	// 機率 0 時任何擲骰都不觸發(擲骰下界是 1)。
	if RebellionTriggers(0, 1) {
		t.Error("機率 0 不該觸發")
	}
}

// 起事的不是全部未同化人口,是 rand(1..n)——手冊完全沒提這件事,純粹讀出來的。
func TestRebelUnitsIsARollNotTheWholePopulation(t *testing.T) {
	// 擲出最小值。
	if got := RebellionRebelUnits(8, func(int) int { return 1 }); got != 1 {
		t.Errorf("擲 1 應只有 1 單位起事,得到 %d", got)
	}
	// 擲出最大值 = 全部。
	if got := RebellionRebelUnits(8, func(n int) int { return n }); got != 8 {
		t.Errorf("擲滿應是 8 單位,得到 %d", got)
	}
	// 夾在範圍內:壞掉的 roll 不該讓叛軍多於未同化人口。
	if got := RebellionRebelUnits(3, func(int) int { return 99 }); got != 3 {
		t.Errorf("叛軍不該多於未同化人口,得到 %d", got)
	}
	if got := RebellionRebelUnits(0, func(int) int { return 1 }); got != 0 {
		t.Errorf("沒有未同化人口就沒有叛軍,得到 %d", got)
	}
}

// 叛軍是第四種地面部隊(類型 3),攻擊力 −20,是四種裡最弱的。
//
// 定名依據見 ground_battle_orig.go 檔頭:`Get_Rebellion_Info_` 把守方三種填進
// +0x0A/+0x0C/+0x0E,叛軍數量填進 +0x10——同一陣列的第四格。
func TestRebelsAreTheWeakestGroundUnitType(t *testing.T) {
	if GroundTypeRebels != 3 {
		t.Fatalf("叛軍應是類型 3,得到 %d", GroundTypeRebels)
	}
	rebels := GroundTypeStrengthDelta(GroundTypeRebels)
	for _, other := range []int{GroundTypeArmor, GroundTypeMarines, GroundTypeMilitia} {
		if rebels >= GroundTypeStrengthDelta(other) {
			t.Errorf("叛軍(%d)應弱於類型 %d(%d)", rebels, other, GroundTypeStrengthDelta(other))
		}
	}
}
