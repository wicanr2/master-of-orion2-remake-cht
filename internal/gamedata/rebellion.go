package gamedata

// rebellion.go:**被征服人口的叛亂**——手冊只說「越多越可能」,機率是從執行檔讀出來的。
//
// ============ 手冊給了規則,沒給數字 ============
//
// GAME_MANUAL.pdf p.165:
//
//	There is a chance each turn that the conquered aliens will attempt to revolt.
//	**The more unassimilated aliens, the larger the chance.** This chance is **doubled**
//	if the captured population is being **exterminated** and **halved** if there is an
//	**Alien Management Center** in the colony. If they do revolt, they fight it out with
//	your ground troops for control of the planet. A loss for you is a gain for the world's
//	old ruler — **the colony reverts back**.
//
// 還有一句限定範圍(p.184):
//
//	no colony (**except captured ones**) ever revolts.
//
// 「越多越可能」是什麼函數、基準是多少,手冊沒寫。這些數字來自 `Check_Rebellion_` @ 0xED260。
//
// ============ 一手推導 ============
//
//	cmp  byte ptr [colony+12Fh], 4
//	jz   結束                       ; ★ 4 = 未被征服(初始化時設的預設值)→ 根本不檢定
//	...                             ; 掃人口單位,ecx = 未同化且「原主帝國還在」的單位數
//	test ecx, ecx
//	jle  結束                       ; 沒有這種人口 → 不叛亂
//	imul edx, ecx, 0Ah              ; ★ 機率 = 單位數 × 10
//	...
//	    movzx eax, byte_199CB0      ;   難度(0 教學 … 4 不可能)
//	    shl   eax, 2
//	    sub   eax, 8
//	    add   edx, eax              ; ★ 只在「殖民地主人是人類、叛軍是 AI」時:+ 難度×4 − 8
//	cmp  byte ptr [colony+137h], 0
//	jz   略過
//	    edx = edx / 2               ; ★ 異族管理中心 → 減半
//	cmp  byte ptr [colony+12Fh], 0
//	jnz  略過
//	    add edx, edx                ; ★ 滅絕政策 → 加倍
//	mov  eax, 3E8h
//	call sub_1247A0                 ; rand(1..1000)
//	cmp  eax, edx
//	jg   結束                       ; ★ 擲骰 > 機率 → 不叛亂
//
// 所以基準是**每一單位未同化人口 1%**(10/1000),而順序是:
// 基準 → 難度 → 減半 → 加倍。順序有差:先減半再加倍與先加倍再減半,在奇數上結果不同。
//
// `sub_1247A0` 是原版的 `Random_`,尾巴有 `inc eax`,回傳 **1..n**(見 ground_battle_orig.go
// 的同款註記),所以判定是 `roll <= chance`,不是 `<`。
//
// ============ [colony+0x137] 是異族管理中心 ============
//
// 這個欄位在三個地方被讀,每一個都對上手冊 AMC 描述的**不同一句**:
//
//	| 呼叫端 | 條件 | 手冊 |
//	|---|---|---|
//	| `Check_Rebellion_` | `!= 0` → 機率減半 | 「halving the chance of revolt」 |
//	| `Apply_Assimilation_` | `!= 0` → 速率固定 120 | 「1 per 2 turns, **regardless of government**」 |
//	| `sub_DDAD4` | `== 0` 才算多種族 | 「removes the 20% morale penalty from multi-racial colonies」 |
//
// 而 `OrigBuildingID["Alien Management Center"] = 1`。三句對三個呼叫端,不是猜的。
//
// ============ 誠實留白 ============
//
//   - **`[colony+0x12F]` 的完整列舉沒查**。已知 4 = 未征服(初始值)、0 = 滅絕
//     (由「×2」對上手冊那句反推),1/2/3 在別處有比對但語意未定。
//   - remake **沒有滅絕政策**這個選項,所以「×2」那一路目前不會發生。函式仍收這個參數
//     ——留著介面,等 UI 有那個選項時直接接上,不必回頭改公式。

// RebellionChancePerUnit 是每一單位未同化人口貢獻的叛亂機率,單位是**千分之一**
// (原版 `imul edx, ecx, 0Ah` 搭配 `rand(1..1000)`,即每單位 1%)。
const RebellionChancePerUnit = 10

// RebellionRollMax 是叛亂檢定的擲骰上界(原版 `mov eax, 3E8h` → `rand(1..1000)`)。
const RebellionRollMax = 1000

// RebellionUnconqueredPolicy 是「這個殖民地從來沒被征服過」的政策值
// (原版初始化 `mov byte ptr [colony+12Fh], 4`)。`Check_Rebellion_` 一看到它就直接返回
// ——對上手冊「no colony (except captured ones) ever revolts」。
const RebellionUnconqueredPolicy = 4

// RebellionExterminatePolicy 是「正在滅絕被征服人口」的政策值(原版 `[colony+12Fh] == 0`
// → 機率加倍,對上手冊「doubled if the captured population is being exterminated」)。
const RebellionExterminatePolicy = 0

// RebellionDifficultyAdjust 回傳難度給叛亂機率的調整(千分之一,原版 `難度×4 − 8`)。
//
// 難度索引同 `shell.Difficulties`(0 教學 … 2 普通 … 4 不可能),所以**普通是 0**
// ——與 `GroundDifficultyBonus` 同一種「以普通為基準往兩邊偏」的寫法。
//
// ⚠ 原版只在「殖民地主人是**人類**且叛軍屬於 **AI** 帝國」時才加這一項
// (`[player+0x28] == 100` 的雙重比較,見 GroundHumanPlayer 的同款判定)。
func RebellionDifficultyAdjust(difficulty int) int {
	return difficulty*4 - 8
}

// RebellionChancePermille 回傳這個殖民地本回合的叛亂機率(千分之一)。
//
// unassimilated 只算「原主帝國仍然存在」的未同化人口單位——原版在計數迴圈裡查了
// `[player+0x24] == 0` 才 `inc ecx`。這條有規則上的道理:叛亂成功要有可以「還政」的舊主。
//
// 運算順序照原版:基準 → 難度 → 減半 → 加倍。
func RebellionChancePermille(unassimilated, difficulty int, ownerIsHuman, rebelIsAI, hasAlienManagementCenter, exterminating bool) int {
	if unassimilated <= 0 {
		return 0
	}
	chance := unassimilated * RebellionChancePerUnit
	if ownerIsHuman && rebelIsAI {
		chance += RebellionDifficultyAdjust(difficulty)
	}
	if hasAlienManagementCenter {
		chance /= 2 // 原版是 `sub/sar 1`(向零取整),對正數等同 Go 的整數除法
	}
	if exterminating {
		chance *= 2
	}
	if chance < 0 {
		chance = 0 // 低難度 + 只有一單位未同化人口時算得出負數,擲骰恆不成立,夾成 0 讓語意直白
	}
	return chance
}

// RebellionTriggers 回傳這一擲是否觸發叛亂。roll 由 rand(1..RebellionRollMax) 產生。
//
// 原版是 `cmp eax, edx / jg 結束`——擲骰**大於**機率才不叛亂,所以等於也算中。
func RebellionTriggers(chancePermille, roll int) bool {
	return roll <= chancePermille
}

// RebellionRebelUnits 回傳實際起事的叛軍單位數。
//
// **不是全部未同化人口**——原版擲了第二個骰:`mov eax, ecx / call sub_1247A0`,
// 結果直接當成叛軍的部隊數傳給 `Get_Rebellion_Info_`([叛軍結構+0x10])。
// 所以是 **rand(1..未同化人口數)**。這一條沒有手冊來源,純粹是讀出來的。
//
// roll 傳 `func(n int) int` 回 1..n(呼叫端負責語意,見 shell 的 eventRoll)。
func RebellionRebelUnits(unassimilated int, roll func(n int) int) int {
	if unassimilated <= 0 {
		return 0
	}
	n := roll(unassimilated)
	if n < 1 {
		n = 1
	}
	if n > unassimilated {
		n = unassimilated
	}
	return n
}
