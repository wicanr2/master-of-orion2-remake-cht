package gamedata

// crew.go:**艦員經驗**——四條加成軌早就在,缺的是「經驗怎麼累積、什麼時候升級」。
//
// ============ remake 先前的狀態 ============
//
// 三張加成表已經分散在兩個檔案裡,而且都對得上手冊:
//
//	formulas.go  shipCrewOffenseBonuses = {0, 15, 30, 50, 75}   ← 手冊 BA 欄
//	formulas.go  shipCrewDefenseBonuses = {0, 15, 30, 50, 75}   ← 手冊 BD 欄
//	missile.go   MissileCrew*           =  0, 7, 15, 25, 37     ← 手冊 ME 欄
//
// 但**沒有任何一艘船有等級**:`shell.Ship` 沒有這個欄位,也沒有東西會讓它上升。
// 三張表都只在「讀存檔的船」那條路徑上有呼叫端(`engine.ShipBeamAttackFromDesign`),
// remake 自己造的船永遠是新兵。
//
// ============ 手冊 p.121 的完整表 ============
//
//	| 等級 | BA | BD | ME | Bo | EP | EP(統帥種族) |
//	|---|---|---|---|---|---|---|
//	| Green 新兵      |  0 |  0 |  0 |  0 |   0 |  — |
//	| Regular 正規    | +15| +15| +7 | +5 |  50 |   0 |
//	| Veteran 老兵    | +30| +30| +15| +10| 150 |  50 |
//	| Elite 精銳      | +50| +50| +25| +15| 500 | 150 |
//	| Ultra-Elite     | +75| +75| +37| +20|  —  | 500 |
//
// 最後一列的星號:「**This level is only attainable by crews in the service of a
// Warlord race.**」——所以統帥種族不是「升級快一階」,而是**整條階梯往上平移一格**:
// 他們的船一出廠就是正規兵,而且多一個一般種族碰不到的頂級。
//
// 這也解釋了手冊前面那句「All ship crews start out as green rookies **unless** they're
// in service with a race that has the Warlord characteristic」——兩句話講的是同一件事。
//
// ============ 經驗從哪來 ============
//
//	Each turn in space counts for **1 experience point**.
//	Each battle won gains the crew … experience points equal to the **halved sum of
//	size classes (1-6) of destroyed (not captured) enemy ships** (rounded down with a
//	**minimum of 1** experience point).
//
// 三個限定詞都有意義:**halved**(所以打小船升得慢)、**destroyed not captured**
// (俘虜不算,鼓勵擊沉)、**minimum 1**(擊沉一艘巡防艦 1/2 = 0,但保底給 1)。
//
// 另外太空學院(p.97)給兩件事:該殖民地造的船**起始等級 +1**,
// 以及該星系內每有一座學院,停泊的船**每回合多 1 EP**。

// 艦員等級(索引與 formulas.go 的三張加成表一致)。
const (
	CrewGreen      = 0
	CrewRegular    = 1
	CrewVeteran    = 2
	CrewElite      = 3
	CrewUltraElite = 4
)

// CrewLevelCount 是等級數(含超級精銳)。
const CrewLevelCount = 5

// crewBoardingBonuses 是手冊 Bo 欄(登艦戰加成)。
//
// 這是四條軌裡 remake 唯一沒有的一條——BA/BD 在 formulas.go、ME 在 missile.go,
// 登艦戰先前沒有建模所以沒人抄。抄進來,即使暫時沒有呼叫端:
// 缺一條軌會讓下次有人接登艦戰時以為手冊沒給數字。
var crewBoardingBonuses = [CrewLevelCount]int{0, 5, 10, 15, 20}

// crewMissileEvasionBonuses 是手冊 ME 欄,值取自 missile.go 既有常數(同一份手冊表)。
var crewMissileEvasionBonuses = [CrewLevelCount]int{
	MissileCrewGreen, MissileCrewRegular, MissileCrewVeteran,
	MissileCrewElite, MissileCrewUltraElite,
}

// crewXPThresholds 是升到該等級所需的累積經驗(手冊 EP 欄,一般種族)。
//
// 索引 4(超級精銳)是 -1 = **一般種族到不了**。用 -1 而不是一個很大的數:
// 「到不了」與「很難到」在規則上是兩件事,寫一個大數字會讓某天有人把它調小就破功。
var crewXPThresholds = [CrewLevelCount]int{0, 50, 150, 500, -1}

// crewXPThresholdsWarlord 是統帥種族的版本:整條階梯往上平移一格。
//
// 索引 0(新兵)是 -1 = **統帥種族沒有新兵**,他們的船出廠就是正規兵。
var crewXPThresholdsWarlord = [CrewLevelCount]int{-1, 0, 50, 150, 500}

// CrewXPPerTurnInSpace 是「每回合在太空中」得到的經驗(手冊:1)。
const CrewXPPerTurnInSpace = 1

// CrewXPTurnMaximum 是 Do_All_Ships_XP_Check_／sub_149D5 @ 0x149D5 在每回合
// CrewExp 寫回前套用的上限；raw word_17D186 @ 0x17D186 = 500。
const CrewXPTurnMaximum = 500

// CrewXPBattleMinimum 是打贏一場至少拿到的經驗(手冊:minimum of 1)。
const CrewXPBattleMinimum = 1

// CrewXPAfterTurnGain 套用原版每回合 XP consumer 的非負邊界與 500 上限。
// 戰後獎勵屬另一條原版鏈，不使用本函式。
func CrewXPAfterTurnGain(current, gain int) int {
	if current < 0 {
		current = 0
	}
	if gain < 0 {
		gain = 0
	}
	current += gain
	if current > CrewXPTurnMaximum {
		return CrewXPTurnMaximum
	}
	return current
}

// SpaceAcademyStartingLevelBonus 是太空學院讓該殖民地造的船起始等級提高幾級(手冊 p.97:1)。
const SpaceAcademyStartingLevelBonus = 1

// SpaceAcademyXPPerTurn 是星系內每座太空學院給停泊艦艇的每回合額外經驗(手冊 p.97:1)。
const SpaceAcademyXPPerTurn = 1

// ShipCrewBoardingBonus 回傳艦員等級的登艦戰加成(手冊 Bo 欄)。
func ShipCrewBoardingBonus(level int) int {
	if level < 0 || level >= CrewLevelCount {
		return 0
	}
	return crewBoardingBonuses[level]
}

// ShipCrewMissileEvasionBonus 回傳艦員等級的飛彈閃避加成(手冊 ME 欄)。
func ShipCrewMissileEvasionBonus(level int) int {
	if level < 0 || level >= CrewLevelCount {
		return 0
	}
	return crewMissileEvasionBonuses[level]
}

// CrewStartingLevel 回傳一艘新船的起始艦員等級。
//
// 統帥種族從正規兵開始(手冊:「unless they're in service with a race that has the
// Warlord characteristic」),其餘從新兵開始。
func CrewStartingLevel(warlord bool) int {
	if warlord {
		return CrewRegular
	}
	return CrewGreen
}

// CrewMaxLevel 回傳這個種族的艦員能達到的最高等級。
//
// 超級精銳只有統帥種族碰得到(手冊表的星號)。
func CrewMaxLevel(warlord bool) int {
	if warlord {
		return CrewUltraElite
	}
	return CrewElite
}

// CrewLevelForXP 回傳累積 xp 對應的艦員等級。
//
// xp 是「這艘船累積的經驗」,不是「距離下一級還差多少」。
func CrewLevelForXP(xp int, warlord bool) int {
	table := &crewXPThresholds
	if warlord {
		table = &crewXPThresholdsWarlord
	}
	level := CrewStartingLevel(warlord)
	for l := 0; l < CrewLevelCount; l++ {
		need := table[l]
		if need < 0 {
			continue // 這個種族到不了這一級(見兩張表的 -1 說明)
		}
		if xp >= need && l > level {
			level = l
		}
	}
	return level
}

// CrewXPToNextLevel 回傳距離下一級還差多少經驗;已經到頂回 0。
//
// 這個函式存在是為了讓 UI 能顯示進度——一個只會默默上升的等級對玩家等於不存在。
func CrewXPToNextLevel(xp int, warlord bool) int {
	table := &crewXPThresholds
	if warlord {
		table = &crewXPThresholdsWarlord
	}
	cur := CrewLevelForXP(xp, warlord)
	if cur >= CrewMaxLevel(warlord) {
		return 0
	}
	need := table[cur+1]
	if need < 0 {
		return 0
	}
	if d := need - xp; d > 0 {
		return d
	}
	return 0
}

// CrewXPForLevel 回傳「剛好升到 level」所需的累積經驗。
//
// 給「一出廠就是某個等級」的情境用(太空學院 +1 起始等級):與其另存一個等級欄位,
// 不如把起始經驗設成那一級的門檻——**等級只有一個真相來源(經驗)**,
// 兩個欄位遲早會不同步。
//
// 該種族到不了的等級回 -1。
func CrewXPForLevel(level int, warlord bool) int {
	if level < 0 || level >= CrewLevelCount {
		return -1
	}
	if warlord {
		return crewXPThresholdsWarlord[level]
	}
	return crewXPThresholds[level]
}

// CrewBattleXPFromDestroyedHullClassSum 回傳打贏一場戰鬥時，每艘合格艦取得的經驗。
//
// sub_4B184 @ 0x4B184 先累加被摧毀敵艦的 1-based hull class，再於
// 0x4B90C..0x4B919 除二並強制最少 1。即使總和是 0，勝方 recipient 仍得到 1；
// caller 必須自行確保「確實有勝方且有 recipient」。這條直接 add Ship+0x72，不共用
// 每回合 consumer 的 500 上限。
func CrewBattleXPFromDestroyedHullClassSum(sum int) int {
	if sum < 0 {
		sum = 0
	}
	if xp := sum / 2; xp >= CrewXPBattleMinimum {
		return xp
	}
	return CrewXPBattleMinimum
}

// CrewBattleXP 回傳打贏一場戰鬥得到的經驗。
//
// sizeClasses 是**被擊沉**(不含被俘)的敵艦艦體等級(1–6)。
// 手冊:總和折半、無條件捨去、最少 1。
//
// 空清單仍回 1；這不是手冊延伸解讀，而是 sub_4B184 的原始 min-1 分支。
func CrewBattleXP(sizeClasses []int) int {
	sum := 0
	for _, c := range sizeClasses {
		if c > 0 {
			sum += c
		}
	}
	return CrewBattleXPFromDestroyedHullClassSum(sum)
}
