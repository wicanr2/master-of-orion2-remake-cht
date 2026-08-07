package gamedata

// ground_battle_orig.go:**原版的地面戰解算**(`Ground_Combat_Round_` @ 0xEC4FE +
// `Resolve_Ground_Combat_` @ 0xEC601)。
//
// ============ 為什麼要重做一次 ============
//
// `ground_battle.go` 的檔頭寫著:
//
//	解算結構取自一代(1oom)game_ground.c 的 game_ground_kill
//
// 而 `docs/HONEST-STATUS.md` 把它列成「force 值用 MOO2 手冊表,但**結構本身未對 MOO2 實機核實**」。
// 2026-08-07 反組譯把結構挖出來了,而且與一代那套**有三處實質差異**——不是風格差異,
// 是會改變勝負機率的差異。
//
// ============ 原版的一方是 26 個位元組 ============
//
//	+0x00  word     ?(本函式沒用到)
//	+0x02  word[4]  各部隊類型的**攻擊力**
//	+0x0A  word[4]  各部隊類型的**剩餘單位數**
//	+0x12  byte[4]  各部隊類型的**耐受命中數**(要挨幾下才死一個)
//	+0x16  byte     **當前部隊類型**(0..3;**4 = 全滅**)
//	+0x17  byte     當前類型的累積命中數
//	+0x18  byte     本回合被打中但沒死的類型(0xFF = 無)
//	+0x19  byte     本回合死了一個單位的類型(0xFF = 無)
//
// 欄位剛好排滿 0x00..0x19,沒有空隙——這本身就是「四種部隊類型」這個判讀的佐證。
//
// ============ 一回合(`Ground_Combat_Round_`)============
//
//	strA = A.攻擊力[A.當前類型] + Random(100)
//	strB = B.攻擊力[B.當前類型] + Random(100)
//	if (strA <= strB) A 挨一下
//	if (strA >= strB) B 挨一下          ; ← 注意是兩個獨立的 if
//
// 「挨一下」:累積命中 +1;**等於**該類型的耐受值時死一個單位、累積歸零;
// 該類型歸零時往後找下一個還有兵的類型(找到 4 就是全滅)。
//
// `Resolve_Ground_Combat_` 的迴圈條件是 `A.當前類型 < 4 && B.當前類型 < 4`
// ——**打到任一方的類型索引推到 4(全滅)為止**。
//
// ============ 與 remake 沿用的一代結構的三處差異 ============
//
//	1. **平手時雙方都挨打。** 原版是兩個獨立的 `if`(`<=` 與 `>=`),平手時兩個都成立。
//	   remake 是 `if/else`,平手只有攻方挨打(「平手歸守方」)。
//	   一代確實是那樣,但二代不是。
//	2. **攻擊力依當前部隊類型而不同**,不是整隊共用一個值。
//	3. **累積命中用 `==` 判定**(`cmp cl, [ebx+eax+12h] / jnz`),不是「扣到 <= 0」。
//	   差別在耐受值被改小的邊界上會出現——`==` 錯過就永遠不死。
//	   remake 用遞減到 0,那個版本沒有這個坑,但也不是原版的算法。
//
// 第 1 點是**可觀察的機率差異**:d100 對 d100 平手的機率是 1%,而平手在原版會讓雙方
// 各損失一次命中。守方原本白拿的那 1% 沒有了。
//
// ============ 誠實留白 ============
//
// 四種部隊類型**沒有對出名字**。MOO2 的地面單位有陸戰隊 / 裝甲(坦克)/ 機械戰士,
// 那是三種;第四種是什麼、以及各類型的攻擊力從哪張表來,都還沒追。
// 所以這一檔只提供**解算**,兵力怎麼組還是呼叫端的事(見 `ground_battle.go` 的加成表)。

// GroundUnitTypes 是原版一方的部隊類型數(索引 0..3;4 = 全滅哨兵值)。
//
// 來自 `Ground_Combat_Round_` 的 `cmp byte ptr [ebx+16h], 4 / jnb`
// 與 26 位元組結構剛好排滿(4 個 word 攻擊力 + 4 個 word 數量 + 4 個 byte 耐受值)。
const GroundUnitTypes = 4

// GroundSideExhausted 是「全滅」的類型索引(原版的 4)。
const GroundSideExhausted = GroundUnitTypes

// GroundSide 是原版地面戰的一方。
type GroundSide struct {
	// Strength / Count / HitsToKill 逐部隊類型(原版 +0x02 / +0x0A / +0x12)。
	Strength   [GroundUnitTypes]int
	Count      [GroundUnitTypes]int
	HitsToKill [GroundUnitTypes]int

	// CurType / hits 是原版的 +0x16 / +0x17。
	CurType int
	hits    int

	// HitType / DeadType 是本回合的結果(原版 +0x18 / +0x19;-1 = 無,原版用 0xFF)。
	HitType  int
	DeadType int
}

// GroundNone 是 HitType / DeadType 的「無」(原版 0xFF)。
const GroundNone = -1

// NewGroundSide 建一方,並把當前類型推到第一個有兵的類型。
func NewGroundSide(strength, count, hitsToKill [GroundUnitTypes]int) *GroundSide {
	s := &GroundSide{
		Strength: strength, Count: count, HitsToKill: hitsToKill,
		HitType: GroundNone, DeadType: GroundNone,
	}
	s.advance()
	return s
}

// advance 把當前類型往後推到下一個還有兵的類型(推到 4 就是全滅)。
//
// 原版:`while (count[type] == 0 && type < 4) type++`。
func (s *GroundSide) advance() {
	for s.CurType < GroundSideExhausted && s.Count[s.CurType] == 0 {
		s.CurType++
	}
}

// Exhausted 回報這一方是不是全滅了。
func (s *GroundSide) Exhausted() bool { return s.CurType >= GroundSideExhausted }

// AliveUnits 回傳還活著的單位總數(所有類型加總)。
func (s *GroundSide) AliveUnits() int {
	n := 0
	for _, c := range s.Count {
		n += c
	}
	return n
}

// curStrength 回傳當前類型的攻擊力(全滅時回 0)。
func (s *GroundSide) curStrength() int {
	if s.Exhausted() {
		return 0
	}
	return s.Strength[s.CurType]
}

// takeHit 是原版的「挨一下」。回傳這一下有沒有打死一個單位。
func (s *GroundSide) takeHit() bool {
	if s.Exhausted() {
		return false
	}
	t := s.CurType
	s.hits++
	// ⚠ 原版是 `==` 不是 `>=`(`cmp cl, [ebx+eax+12h] / jnz`)。照抄。
	if s.hits != s.HitsToKill[t] {
		s.HitType = t
		return false
	}
	s.DeadType = t
	s.Count[t]--
	s.hits = 0
	if s.Count[t] == 0 {
		s.advance()
		s.hits = 0
	}
	return true
}

// GroundRoll 是解算要用的擲骰:回傳 [0, n) 的整數(原版 `Random_` @ 0x1247A0)。
//
// 傳進來而不是內建亂數源:同 ground_battle.go 的立場——確定性測試要能重現。
type GroundRoll func(n int) int

// GroundCombatRound 跑一回合(原版 `Ground_Combat_Round_` @ 0xEC4FE)。
//
// ⚠ **兩個獨立的 if,不是 if/else**:平手時雙方都挨打。這是與一代結構最實質的差異。
func GroundCombatRound(a, b *GroundSide, roll GroundRoll) {
	a.HitType, a.DeadType = GroundNone, GroundNone
	b.HitType, b.DeadType = GroundNone, GroundNone

	strA := a.curStrength() + roll(100)
	strB := b.curStrength() + roll(100)

	if strA <= strB {
		a.takeHit()
	}
	if strA >= strB {
		b.takeHit()
	}
}

// GroundOrigResult 是一場原版地面戰的結果。
type GroundOrigResult struct {
	AttackerWon      bool
	AttackerSurvived int
	DefenderSurvived int
	Rounds           int
}

// ResolveGroundCombatOrig 打到一方全滅為止(原版 `Resolve_Ground_Combat_` @ 0xEC601)。
//
// 原版的迴圈條件就是「雙方的類型索引都還 < 4」——**沒有回合上限**。這裡加一個
// maxRounds 純粹是防呆:雙方攻擊力都是 0 且耐受值極大時原版也會跑很久,
// 但那是資料錯誤,不該讓它變成當掉。maxRounds <= 0 時用一個保守的預設值。
func ResolveGroundCombatOrig(atk, def *GroundSide, roll GroundRoll, maxRounds int) GroundOrigResult {
	if maxRounds <= 0 {
		maxRounds = 100000
	}
	rounds := 0
	for !atk.Exhausted() && !def.Exhausted() && rounds < maxRounds {
		rounds++
		GroundCombatRound(atk, def, roll)
	}
	return GroundOrigResult{
		// 攻方勝 = 守方全滅。雙方同時全滅(平手那一下互相打死最後一個)時
		// **判給守方**——攻方沒有兵力可以佔領,同 ResolveGroundBattle 的既有立場。
		AttackerWon:      def.Exhausted() && !atk.Exhausted(),
		AttackerSurvived: atk.AliveUnits(),
		DefenderSurvived: def.AliveUnits(),
		Rounds:           rounds,
	}
}
