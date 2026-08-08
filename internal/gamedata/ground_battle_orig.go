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
// ============ 四種類型是什麼(2026-08-07 追出來的)============
//
// `Compute_Ground_Combat_Info_` @ 0xEC3CE 逐類型算攻擊力與耐受值,四個 case 的**調整量
// 是純立即數**:
//
//	case 0:  攻擊力 += 10 + 加成塊[+1] ; 耐受 += 1 + 加成塊[+2]
//	case 1:  攻擊力 +=      加成塊[+3] ; 耐受 +=     加成塊[+4]
//	case 2:  攻擊力 -= 10
//	case 3:  攻擊力 -= 20              ; 而且基礎值取自**另一方**的加成塊(edi;為 0 時整格歸零)
//
// `Compute_Colony_Ground_Combat_Info_` @ 0xED713 給殖民地填三格數量
//(`[+0x0A]`、`[+0x0C]`、`[+0x0E]`,即類型 0/1/2),第 4 格留 0(它傳 `ebx = 0`)。
//
// 手冊那一句把三種對上了:
//
//	「Along the bottom left of the view are icons representing all the **Marine and Armor**
//	 units stationed in defense of this planet. … In addition to Marine and Armor units your
//	 **militia** are also shown here.」
//
// 對照調整量:類型 0 最強(+10 攻擊、+1 耐受)= **裝甲**(手冊:tank battalions);
// 類型 1 是基準 = **陸戰隊**;類型 2 −10 = **民兵**(未受訓的平民,最弱,合理)。
//
// ⚠ **類型 3 是叛軍**(2026-08-08 第 46 項(叛亂)定名,先前掛著「未定名」)。
//
// 認出來的是 `Get_Rebellion_Info_` @ 0xEC65A:它把守方三種部隊填進
// `[+0x0A]`(裝甲)、`[+0x0C]`(陸戰隊)、`[+0x0E]`(民兵),而**叛軍的數量填進 `[+0x10]`**
// ——同一個陣列的第四格。所以類型 3 = 叛軍。
//
// 這順帶解釋了原本那句「殖民地防守方根本不填它」:**叛軍永遠是攻方**,守方那格當然是空的。
// 攻擊力 −20(四種裡最弱)也合理——起事的是沒有受過訓練的被征服人口。
//
// ⚠ 那幾個「加成塊」欄位(`Compute_Player_Ground_Combat_Bonuses_` @ 0xEC15C 產的)
// 還沒逐欄對出意義,所以這一檔**只提供類型間的相對差**(那部分是純立即數,不依賴未知欄位),
// side 級的基礎加成仍由呼叫端用手冊的表算(見 `ground_battle.go`)。

// 部隊類型索引(見上方的對應推導)。
const (
	GroundTypeArmor   = 0 // 裝甲 / 戰車營(+10 攻擊、+1 耐受)
	GroundTypeMarines = 1 // 陸戰隊(基準)
	GroundTypeMilitia = 2 // 民兵(−10 攻擊)
	GroundTypeRebels  = 3 // 叛軍(−20,最弱;見上方的定名依據)
)

// GroundTypeStrengthDelta 是各部隊類型相對基準的攻擊力調整(原版四個 case 的立即數)。
//
// 只回**立即數的部分**;`加成塊[+1]` / `加成塊[+3]` 那兩個科技加成欄位還沒對出意義,
// 不含在這裡——回一個「差不多」的值會讓日後追出真值時看不出哪裡被污染過。
func GroundTypeStrengthDelta(unitType int) int {
	switch unitType {
	case GroundTypeArmor:
		return 10
	case GroundTypeMarines:
		return 0
	case GroundTypeMilitia:
		return -10
	case GroundTypeRebels:
		return -20
	}
	return 0
}

// GroundTypeHitsDelta 是各部隊類型相對基準的耐受命中數調整(原版只有類型 0 有 +1)。
func GroundTypeHitsDelta(unitType int) int {
	if unitType == GroundTypeArmor {
		return 1
	}
	return 0
}

// --- 加成塊(`Compute_Player_Ground_Combat_Bonuses_` @ 0xEC15C)---
//
// 那支函式產一個 **19 位元組**的加成塊,`Compute_Ground_Combat_Info_` 再從裡面取值。
// 逐欄追下來,大多是「某個科技已知就給 +10」這種、對應手冊既有的加成表
//(remake 的 `GroundArmorTechBonus` 等已經涵蓋),但有**兩條是手冊沒有、只有反組譯才知道**的:
//
// ============ 一、基礎耐受命中數是 1,某個科技讓它變 2 ============
//
//	[out+0x0C] = 1  if [player+0x8AA] != 0
//	...
//	耐受命中數 = [加成塊+0x0C] + 1        ; Compute_Ground_Combat_Info_
//
// 也就是**預設一下就死一個單位**,那個科技讓所有部隊都變成要兩下。
//
// ============ 二、地面戰的難度加成**不給玩家** ============
//
//	if ([player+0x28] == 100) [out+0x0F] = 0            ; 100 = 人類玩家的標記
//	else                      [out+0x0F] = 難度 − 2
//	; 而玩家編號 >= 8 的那一側(安塔蘭 / 怪獸)走另一條路徑:
//	                          [out+0x0F] = 難度×2 − 4
//
// `[out+0x0F]` 被加進**所有部隊類型**的攻擊力(`var_4`)。所以:
//
//	人類玩家     +0
//	AI 帝國      難度 − 2        (普通 = 0、不可能 = +2、教學 = −2)
//	安塔蘭那側   難度×2 − 4      (**雙倍**)
//
// `[player+0x28] == 100` 這個「人類玩家標記」在 `Init_Player_Tech_` @ 0x5E55F 也出現過
// (`cmp byte ptr [eax+28h], 64h`)——**同一個標記在兩支不相干的函式裡對得上**。
//
// ⚠ 其餘欄位(`[+5]`/`[+7]`/`[+9]` 那幾張表、`[+0x0B]`、`[+0x10]`)還沒逐欄對出意義,
// 但它們對應的是手冊已經列出的那幾類加成,remake 已用手冊的表算過——不重複實作。

// GroundBaseHitsToKill 回傳基礎耐受命中數(原版 `[加成塊+0x0C] + 1`)。
//
// 預設 1(一下死一個);`extraArmour` 是那個把所有部隊變成要兩下的科技
// (`[player+0x8AA]`,還沒對出是哪一項,所以參數用中性的名字)。
func GroundBaseHitsToKill(extraArmour bool) int {
	if extraArmour {
		return 2
	}
	return 1
}

// GroundCombatantKind 是加成塊那三條分支對應的三種參戰者。
type GroundCombatantKind int

const (
	// GroundHumanPlayer 是人類玩家([player+0x28] == 100)。
	GroundHumanPlayer GroundCombatantKind = iota
	// GroundAIEmpire 是 AI 帝國(玩家編號 < 8 但不是人類)。
	GroundAIEmpire
	// GroundAntaranSide 是玩家編號 >= 8 的那一側(安塔蘭 / 怪獸)。
	GroundAntaranSide
)

// GroundDifficultyBonus 回傳難度給的地面戰攻擊力加成(原版 `[加成塊+0x0F]`)。
//
// 難度索引同 `shell.Difficulties`(0 = 教學 … 2 = 普通 … 4 = 不可能),
// 所以「普通」對 AI 是 0 ——**加成是以普通為基準往兩邊偏**,不是一律加成。
func GroundDifficultyBonus(difficulty int, kind GroundCombatantKind) int {
	switch kind {
	case GroundHumanPlayer:
		return 0
	case GroundAntaranSide:
		return difficulty*2 - 4
	}
	return difficulty - 2
}

// MilitiaPerPopulation 是「幾個人口單位出一個民兵」(原版 `Colony_N_Militia_` 的 `idiv 5`)。
const MilitiaPerPopulation = 5

// ColonyMilitiaUnits 回傳一個殖民地的民兵數(原版 `Colony_N_Militia_` @ 0xEC61E)。
//
// 原版是**逐個人口單位掃過去**再除以 5:
//
//	ecx = [colony+0x0A]            ; 人口
//	eax = colony + 0x0C + 人口×4    ; 人口單位陣列的尾端(每個單位 4 位元組)
//	往前掃每一個單位:
//	    if (([eax] & 0x0F) >= 8)      跳過   ; 低 4 bits 是**擁有者編號**
//	    if ([eax+1] & 0x04)           跳過   ; 某個旗標
//	    計數++
//	return 計數 / 5
//
// 兩個跳過條件都是**人口單位上的資料旗標**:
//
//   - `擁有者 >= 8`:玩家編號只有 0..7,8 以上不是正常的帝國人口
//     (`Init_Homeworld_Colony2_` 寫入時就是 `and ebx, 0Fh` 把玩家編號塞進低 4 bits)。
//   - `[+1] & 0x04`:初始化時那個位元是被清掉的(`and byte ptr [eax+0Dh], 0F9h / or 2`
//     只設 bit 1),所以正常人口不會有它。是什麼設的還沒追。
//
// remake 的殖民地人口沒有逐單位的擁有者/旗標模型——每一格都是自己的、都能打,
// 所以兩個條件恆不成立,結果就是 **⌊人口 / 5⌋**。
// 等哪天有了逐單位模型(異族人口、奴隸…),這裡才需要真的去掃。
func ColonyMilitiaUnits(population int) int {
	if population <= 0 {
		return 0
	}
	return population / MilitiaPerPopulation
}

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
