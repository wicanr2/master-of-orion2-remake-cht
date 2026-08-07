package gamedata

import "testing"

// ground_battle_orig_test.go:原版地面戰解算的護欄。
//
// 釘的是**三處與一代結構的差異**——那三處才是這一檔存在的理由。

// 固定骰子:依序回傳給定的值,用完循環。讓「平手」之類的情境可以精準造出來。
func fixedRoll(vals ...int) GroundRoll {
	i := 0
	return func(n int) int {
		v := vals[i%len(vals)]
		i++
		if v >= n {
			v = n - 1
		}
		return v
	}
}

func oneType(strength, count, hits int) *GroundSide {
	var st, ct, hk [GroundUnitTypes]int
	st[0], ct[0], hk[0] = strength, count, hits
	return NewGroundSide(st, ct, hk)
}

// ★ 差異一:**平手時雙方都挨打**(原版是兩個獨立的 if)。
//
// 一代(以及 remake 先前沿用的)是 if/else,平手只有攻方挨打。
// d100 對 d100 平手的機率是 1%,而那 1% 在原版是「雙方各損失一次」,不是守方白拿。
func TestTieDamagesBothSides(t *testing.T) {
	a := oneType(10, 5, 1) // 耐受 1:挨一下就死一個,好觀察
	b := oneType(10, 5, 1)
	GroundCombatRound(a, b, fixedRoll(7, 7)) // 攻擊力相同 + 同一個骰 → 平手
	if a.DeadType != 0 {
		t.Error("平手時攻方應該也損失一個單位")
	}
	if b.DeadType != 0 {
		t.Error("平手時守方應該也損失一個單位")
	}
	if a.AliveUnits() != 4 || b.AliveUnits() != 4 {
		t.Errorf("平手一回合後雙方都該剩 4:攻 %d 守 %d", a.AliveUnits(), b.AliveUnits())
	}
}

// 非平手時只有輸的那一方挨打——正對照,否則上面那條可能只是「兩邊都恆挨打」。
func TestNonTieDamagesOnlyTheLoser(t *testing.T) {
	a := oneType(50, 5, 1)
	b := oneType(10, 5, 1)
	GroundCombatRound(a, b, fixedRoll(0, 0)) // 50 vs 10 → 攻方贏
	if a.DeadType != GroundNone {
		t.Error("贏的一方不該挨打")
	}
	if b.DeadType != 0 {
		t.Error("輸的一方應該挨打")
	}
}

// ★ 差異三:累積命中是 `==` 判定,而且死一個之後歸零。
func TestHitsAccumulateAndResetOnKill(t *testing.T) {
	a := oneType(0, 2, 3) // 每個單位要挨 3 下
	b := oneType(0, 9, 1)
	roll := fixedRoll(0, 50) // 攻方 0+0=0、守方 0+50=50 → 攻方每回合挨打
	for i := 0; i < 2; i++ {
		GroundCombatRound(a, b, roll)
		if a.DeadType != GroundNone {
			t.Fatalf("第 %d 下就死了?耐受值是 3", i+1)
		}
		if a.HitType != 0 {
			t.Errorf("第 %d 下應記成「被打中」", i+1)
		}
	}
	GroundCombatRound(a, b, roll) // 第 3 下
	if a.DeadType != 0 {
		t.Error("第 3 下應該死一個單位")
	}
	if a.AliveUnits() != 1 {
		t.Errorf("死一個後應剩 1,實得 %d", a.AliveUnits())
	}
	// 歸零之後再挨兩下不該死——這條驗的是「死了之後累積歸零」。
	for i := 0; i < 2; i++ {
		GroundCombatRound(a, b, roll)
		if a.DeadType != GroundNone {
			t.Errorf("累積沒歸零:第 %d 下就死了", i+1)
		}
	}
}

// ★ 差異二:攻擊力取**當前部隊類型**的,不是整隊一個值。
//
// 第一種部隊打完換第二種,攻擊力要跟著換。
func TestStrengthFollowsTheCurrentUnitType(t *testing.T) {
	var st, ct, hk [GroundUnitTypes]int
	st[0], ct[0], hk[0] = 0, 1, 1 // 第一種:攻擊力 0,一個單位,一下就死
	st[1], ct[1], hk[1] = 90, 1, 1
	a := NewGroundSide(st, ct, hk)
	if a.CurType != 0 {
		t.Fatalf("一開始應在第 0 種,實得 %d", a.CurType)
	}
	b := oneType(50, 9, 1)
	roll := fixedRoll(0, 0)
	// 第 0 種攻擊力 0 vs 守方 50 → 攻方輸,第一種死光,換到第 1 種。
	GroundCombatRound(a, b, roll)
	if a.CurType != 1 {
		t.Fatalf("第一種打完後應換到第 1 種,實得 %d", a.CurType)
	}
	// 第 1 種攻擊力 90 vs 守方 50 → 這回合攻方該贏。
	GroundCombatRound(a, b, roll)
	if a.DeadType != GroundNone {
		t.Error("換成攻擊力 90 的部隊後不該再挨打——攻擊力沒有跟著類型換")
	}
	if b.DeadType != 0 {
		t.Error("守方應該挨打")
	}
}

// 中間有空類型時要跳過去(原版 `while (count[type]==0 && type<4) type++`)。
func TestAdvanceSkipsEmptyTypes(t *testing.T) {
	var st, ct, hk [GroundUnitTypes]int
	st[2], ct[2], hk[2] = 7, 3, 1 // 只有第 2 種有兵
	s := NewGroundSide(st, ct, hk)
	if s.CurType != 2 {
		t.Errorf("應直接跳到第 2 種,實得 %d", s.CurType)
	}
	if s.Exhausted() {
		t.Error("還有兵不該算全滅")
	}
}

// 完全沒兵 → 當前類型推到 4 = 全滅。
func TestEmptySideIsExhausted(t *testing.T) {
	var zero [GroundUnitTypes]int
	s := NewGroundSide(zero, zero, zero)
	if !s.Exhausted() {
		t.Error("沒有任何兵力應該算全滅")
	}
	if s.CurType != GroundSideExhausted {
		t.Errorf("當前類型應為哨兵值 %d,實得 %d", GroundSideExhausted, s.CurType)
	}
	// 全滅的一方挨打不該 panic,也不該有事發生。
	if s.takeHit() {
		t.Error("全滅的一方不該還能被打死一個")
	}
}

// 一場打到底:攻擊力壓倒性的一方要贏,而且回合數要有限。
func TestResolveEndsWhenOneSideIsExhausted(t *testing.T) {
	atk := oneType(90, 5, 1)
	def := oneType(0, 3, 1)
	res := ResolveGroundCombatOrig(atk, def, fixedRoll(0, 0), 0)
	if !res.AttackerWon {
		t.Errorf("攻擊力 90 vs 0 應該攻方勝:%+v", res)
	}
	if res.DefenderSurvived != 0 {
		t.Errorf("守方應全滅,實得 %d", res.DefenderSurvived)
	}
	if res.Rounds != 3 {
		t.Errorf("守方 3 個單位、每個一下就死 → 應為 3 回合,實得 %d", res.Rounds)
	}
}

// 雙方同時全滅(平手那一下互相打死最後一個)判給守方——攻方沒有兵力可以佔領。
func TestSimultaneousWipeGoesToDefender(t *testing.T) {
	atk := oneType(10, 1, 1)
	def := oneType(10, 1, 1)
	res := ResolveGroundCombatOrig(atk, def, fixedRoll(5, 5), 0) // 平手
	if res.AttackerWon {
		t.Error("雙方同時全滅時不該判攻方勝")
	}
	if res.AttackerSurvived != 0 || res.DefenderSurvived != 0 {
		t.Errorf("雙方都該全滅:攻 %d 守 %d", res.AttackerSurvived, res.DefenderSurvived)
	}
}

// maxRounds 是防呆,不是規則:雙方攻擊力都 0 且永遠打不死時要停得下來。
func TestResolveStopsAtMaxRounds(t *testing.T) {
	var st, ct, hk [GroundUnitTypes]int
	st[0], ct[0], hk[0] = 0, 1, 200 // 耐受 200,而每回合最多挨一下
	atk := NewGroundSide(st, ct, hk)
	def := NewGroundSide(st, ct, hk)
	res := ResolveGroundCombatOrig(atk, def, fixedRoll(0, 0), 10)
	if res.Rounds != 10 {
		t.Errorf("應在第 10 回合停下,實得 %d", res.Rounds)
	}
}

// --- 逐類型的攻擊力/耐受調整(Compute_Ground_Combat_Info_ @ 0xEC3CE 的四個 case)---

// 四個 case 的立即數:+10 / 0 / −10 / −20,耐受只有類型 0 有 +1。
func TestGroundTypeDeltasMatchTheOriginalCases(t *testing.T) {
	for _, c := range []struct{ typ, str, hits int }{
		{GroundTypeArmor, 10, 1},
		{GroundTypeMarines, 0, 0},
		{GroundTypeMilitia, -10, 0},
		{GroundTypeFourth, -20, 0},
	} {
		if got := GroundTypeStrengthDelta(c.typ); got != c.str {
			t.Errorf("類型 %d 的攻擊力調整應為 %d,實得 %d", c.typ, c.str, got)
		}
		if got := GroundTypeHitsDelta(c.typ); got != c.hits {
			t.Errorf("類型 %d 的耐受調整應為 %d,實得 %d", c.typ, c.hits, got)
		}
	}
	// 排序關係才是重點:裝甲 > 陸戰隊 > 民兵 > 第四種。
	if !(GroundTypeStrengthDelta(GroundTypeArmor) >
		GroundTypeStrengthDelta(GroundTypeMarines) &&
		GroundTypeStrengthDelta(GroundTypeMarines) >
			GroundTypeStrengthDelta(GroundTypeMilitia) &&
		GroundTypeStrengthDelta(GroundTypeMilitia) >
			GroundTypeStrengthDelta(GroundTypeFourth)) {
		t.Error("強弱順序應為 裝甲 > 陸戰隊 > 民兵 > 第四種")
	}
	// 類型 0 是索引 0 —— 這條釘住「不要又把陸戰隊排回 0」。
	if GroundTypeArmor != 0 || GroundTypeMarines != 1 {
		t.Error("原版的類型 0 是裝甲(最強的那種),不是陸戰隊")
	}
}

// 裝甲比陸戰隊強 10 點:同基礎值下,裝甲該贏過陸戰隊。
func TestArmorBeatsMarinesAtEqualBase(t *testing.T) {
	const base = 30
	var st, ct, hk [GroundUnitTypes]int
	st[0], ct[0], hk[0] = base+GroundTypeStrengthDelta(GroundTypeArmor), 1, 1
	armor := NewGroundSide(st, ct, hk)

	var st2, ct2, hk2 [GroundUnitTypes]int
	st2[1], ct2[1], hk2[1] = base+GroundTypeStrengthDelta(GroundTypeMarines), 1, 1
	marines := NewGroundSide(st2, ct2, hk2)

	// 同一個骰值 → 只有那 10 點的差異決定勝負。
	GroundCombatRound(armor, marines, func(int) int { return 0 })
	if armor.DeadType != GroundNone {
		t.Error("裝甲不該輸給同基礎值的陸戰隊")
	}
	if marines.DeadType != GroundTypeMarines {
		t.Error("陸戰隊應該挨打")
	}
}

// --- 民兵(Colony_N_Militia_ @ 0xEC61E)---

// 民兵 = ⌊人口 / 5⌋。原版是逐個人口單位掃再除以 5,而 remake 的每一格都算數(見該函式說明)。
func TestColonyMilitiaIsPopulationOverFive(t *testing.T) {
	for _, c := range []struct{ pop, want int }{
		{0, 0}, {1, 0}, {4, 0}, {5, 1}, {9, 1}, {10, 2}, {8, 1}, {25, 5},
		{-3, 0}, // 負數不該回負的民兵
	} {
		if got := ColonyMilitiaUnits(c.pop); got != c.want {
			t.Errorf("人口 %d 應有 %d 個民兵,實得 %d", c.pop, c.want, got)
		}
	}
	if MilitiaPerPopulation != 5 {
		t.Errorf("每 5 個人口一個民兵(原版 idiv 5),實得 %d", MilitiaPerPopulation)
	}
}

// 民兵比陸戰隊弱 10 點——同基礎值時陸戰隊該贏。
func TestMilitiaIsWeakerThanMarines(t *testing.T) {
	const base = 40
	var st, ct, hk [GroundUnitTypes]int
	st[GroundTypeMarines] = base + GroundTypeStrengthDelta(GroundTypeMarines)
	ct[GroundTypeMarines], hk[GroundTypeMarines] = 1, 1
	marines := NewGroundSide(st, ct, hk)

	var st2, ct2, hk2 [GroundUnitTypes]int
	st2[GroundTypeMilitia] = base + GroundTypeStrengthDelta(GroundTypeMilitia)
	ct2[GroundTypeMilitia], hk2[GroundTypeMilitia] = 1, 1
	militia := NewGroundSide(st2, ct2, hk2)

	GroundCombatRound(marines, militia, func(int) int { return 0 })
	if marines.DeadType != GroundNone {
		t.Error("陸戰隊不該輸給同基礎值的民兵")
	}
	if militia.DeadType != GroundTypeMilitia {
		t.Error("民兵應該挨打")
	}
}
