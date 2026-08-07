package gamedata

// combat_speed.go:艦艇的**戰鬥速度**與引擎階——一手表,不是估計值。
//
// ============ 表從哪來 ============
//
// 執行檔的 `Current_Design_Min_Combat_Speed_`(0x6B82A)與 `..._Max_..._`(0x6B84A)
// 讀的是同一張二維表:
//
//	mov     dx, [eax+1Bh]              ; 引擎階
//	mov     bx, [eax+0E5h]             ; 艦體等級(0=巡防艦 … 5=末日之星)
//	imul    eax, 2Eh                   ; 每列 46 位元組
//	mov     al, byte_17FE90[edx+eax]   ; 最小速(最大速在 byte_17FE96,即 +6)
//
// 攤開來(檔案位移 0x1FE524 起,每列 46 位元組,前 6 是各艦體最小速、6..11 是最大速):
//
//	引擎階 1: min=[10  8  6  5  4  3]   max=[20 18 16 15 14 13]
//	引擎階 2: min=[12 10  8  7  6  5]   max=[22 20 18 17 16 15]
//	引擎階 3: min=[14 12 10  9  8  7]   max=[24 22 20 19 18 17]
//	引擎階 4: min=[16 14 12 11 10  9]   max=[26 24 22 21 20 19]
//	引擎階 5: min=[18 16 14 13 12 11]   max=[28 26 24 23 22 21]
//	引擎階 6: min=[20 18 16 15 14 13]   max=[30 28 26 25 24 23]
//
// (第 0 列全 0 = 無引擎;第 7 列以後是表外的別的資料。)
//
// ============ 這張表自己驗證自己 ============
//
// 三個規律,每一個都不像巧合:
//
//   - **每升一階引擎,全部 +2**——與手冊那張戰機表的註腳逐字相符:
//     「* base speed is modified by **+2 per drive level**」
//   - **最大速恆等於最小速 +10**,六階六個艦體共 36 組全部成立
//   - 艦體越大越慢(10/8/6/5/4/3),而且**遞減幅度自己也遞減**——
//     大船之間的差距比小船之間小,這不是隨手填的數列
//
// 所以這裡只存**第 1 階的基準列**,其餘用公式算。存整張表反而讓上面那三個規律變成
// 「36 個各自獨立的數字」,改壞一格不會有任何東西發現。
//
// ============ 誠實留白 ============
//
//   - **min/max 之間怎麼取,remake 不模擬。** 反組譯的
//     `Current_Design_Combat_Speed_`(0x6B86A)是
//     `speed = max − (ratio × (max−min) / 100)`,ratio 由設計的兩個欄位(+0xE9/+0xED)算出。
//     那兩個欄位是什麼**沒有查證**(看起來像佔格用量/總格數),所以不猜。
//     remake 用 **max**(未受損/空載的那一端),並在此標明差距最多是 10。
//   - **戰術棋盤只有 8×6**,而速度值是 13..30。直接當「一回合可走幾格」會讓任何船
//     一步橫跨全場。移動限制要先決定一個棋盤比例尺,**那是 remake 的設計決定不是轉寫**,
//     不夾在這一項裡做。這一項先接**主動權排序**——手冊給了確切公式,不需要比例尺。

// CombatSpeedDriveLevels 是引擎階數(1..6,對應六種 FTL 引擎)。
const CombatSpeedDriveLevels = 6

// combatSpeedBaseMin 是引擎階 1 時各艦體等級的最小戰鬥速度(執行檔第一列)。
var combatSpeedBaseMin = [6]int{10, 8, 6, 5, 4, 3}

const (
	// CombatSpeedPerDriveLevel 每升一階引擎的加成。手冊註腳:「base speed is modified by
	// +2 per drive level」。
	CombatSpeedPerDriveLevel = 2
	// CombatSpeedMaxOverMin 是同一格的最大速與最小速之差(表中 36 組全部成立)。
	CombatSpeedMaxOverMin = 10
	// CombatSpeedAugmentedEngines 手冊(Augmented Engines):「increase the combat speed of
	// a ship by +5」。
	CombatSpeedAugmentedEngines = 5
	// CombatSpeedTransDimensional 手冊 p.?(Trans-Dimensional):「4 to their combat speed」。
	// 與 openorion2 `ShipDesign::combatSpeed` 的 `if (transDimensional) ret += 4` 同值。
	CombatSpeedTransDimensional = 4
)

// CombatSpeedRange 回傳某個引擎階、某個艦體等級的戰鬥速度區間;參數超界回 ok=false。
//
// driveLevel 1..6;class 用 CombatShipClass(0=巡防艦 … 5=末日之星)。
func CombatSpeedRange(driveLevel int, class CombatShipClass) (min, max int, ok bool) {
	if driveLevel < 1 || driveLevel > CombatSpeedDriveLevels {
		return 0, 0, false
	}
	if int(class) < 0 || int(class) >= len(combatSpeedBaseMin) {
		return 0, 0, false
	}
	min = combatSpeedBaseMin[class] + CombatSpeedPerDriveLevel*(driveLevel-1)
	return min, min + CombatSpeedMaxOverMin, true
}

// ShipCombatSpeed 回傳一艘船的戰鬥速度(取區間上端,見檔頭的誠實留白)。
//
// augmentedEngines / transDimensional 依手冊各加 +5 / +4。查無回 0。
func ShipCombatSpeed(driveLevel int, class CombatShipClass, augmentedEngines, transDimensional bool) int {
	_, max, ok := CombatSpeedRange(driveLevel, class)
	if !ok {
		return 0
	}
	if augmentedEngines {
		max += CombatSpeedAugmentedEngines
	}
	if transDimensional {
		max += CombatSpeedTransDimensional
	}
	return max
}

// CombatInitiative 回傳一艘船的主動權。
//
// 手冊逐字:「A ship's initiative is equal to its current **Beam Attack plus 10 times its
// current combat speed**. Thus, smaller ships should move before bigger, slower ones.」
// 後面還把公式再寫了一次:`[Beam Attack + 10 * combat speed]`。
//
// 那句「smaller ships should move before bigger」是這條公式的驗收:速度那一項乘 10,
// 就是為了讓體型差距壓過攻擊力差距。
func CombatInitiative(beamAttack, combatSpeed int) int {
	return beamAttack + 10*combatSpeed
}

// driveTechLevels 依引擎階排列的六種 FTL 引擎科技(索引 0 = 第 1 階)。
var driveTechLevels = [CombatSpeedDriveLevels]Technology{
	TECH_NUCLEAR_DRIVE,
	TECH_FUSION_DRIVE,
	TECH_ION_DRIVE,
	TECH_ANTIMATTER_DRIVE,
	TECH_HYPER_DRIVE,
	TECH_INTERPHASED_DRIVE,
}

// DriveTechForLevel 回傳第 level 階(1..6)的引擎科技;超界回 TECH_NONE。
func DriveTechForLevel(level int) Technology {
	if level < 1 || level > CombatSpeedDriveLevels {
		return TECH_NONE
	}
	return driveTechLevels[level-1]
}

// DriveLevelForTech 回傳某個引擎科技的階數(1..6);不是引擎科技回 0。
func DriveLevelForTech(tech Technology) int {
	for i, t := range driveTechLevels {
		if t == tech {
			return i + 1
		}
	}
	return 0
}

// --- 原版戰術棋盤尺寸(第 137 項,一手)---
//
// `Assign_Combat_Grids_`(0x46CC8)開頭把整張格點清成 0xFFFF:
//
//	loc_46CD4:
//	  xor  eax, eax
//	loc_46CD6:
//	  imul esi, ecx, 88h              ; 列距 0x88 = 136 位元組 = 68 格 × 2
//	  mov  word_18C9A8[esi+ecx*2], 0FFFFh
//	  cmp  ax, 44h                    ; 0x44 = 68
//	  jl   short loc_46CD6
//	  inc  ebx
//	  cmp  bx, 51h                    ; 0x51 = 81
//	  jl   short loc_46CD4
//
// **兩個界限與列距互相驗證**:列距 136 位元組正好裝得下 68 個 uint16,而內圈上界就是 68。
// 同一支函式稍後用 `[ship+21h]` 當外圈索引、`[ship+22h]` 當內圈索引寫回艦艇編號,
// 所以那兩個位元組就是艦艇在棋盤上的座標。
//
// 這兩個數字解釋了手冊那些「以格為單位」的射程為什麼可以那麼大:
// 質子魚雷 24 格、傳送器 12 格、投彈 3 格——放在 81×68 的盤面上都是合理的比例。

const (
	// CombatGridColumns 是原版戰術棋盤的外圈格數(`cmp bx, 51h`)。
	CombatGridColumns = 81
	// CombatGridRows 是內圈格數(`cmp ax, 44h`,亦即列距 0x88 / 2)。
	CombatGridRows = 68
)
