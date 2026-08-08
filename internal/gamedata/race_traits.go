package gamedata

// race_traits.go:原版 13 個內建種族的**特性陣列**——一手資料,不是估計值。
//
// ============ 為什麼要有這一層 ============
//
// remake 先前把種族壓成 7 個自編數字(`shell.Race` 的 IndBonus/ResBonus/FoodBonus/
// GrowthPct/StartBC/IncomePerPop/CombatPct)。原版不是這樣:每位玩家帶一個
// **31 格特性陣列**,存在玩家結構 `+0x89F`——`internal/save/entities.go` 的
// `Traits [31]int8` 讀的就是它,只是遊戲層從來沒有用過。
//
// 壓成 7 個數字有兩個後果:一是**數字本身錯了**(克拉肯是農業+2/工業+1,remake 寫工業+2;
// 阿爾卡里是艦艇**防禦**+50,remake 寫通用戰鬥+15),二是**布林特性整批消失**
// ——水棲、地底、食岩、寬容、戰帥、神級商人、高/低重力……全都無處可查,於是
// `GroundSubterraneanBonus`、`crewXPThresholdsWarlord`、`TradeGoodsIncome` 的
// fantasticTrader 參數、`AIPlanetValueInput.RaceLowG/RaceHeavyG` 這些**寫好的公式
// 一個呼叫端都沒有**。
//
// ============ 資料怎麼來的(三個獨立來源交叉驗證) ============
//
// ① **選項等級**:`RACESTUF.LBX` asset 7。表頭 4 位元組(13 筆、每筆 31 格),
//    其後 13×31。`sub_12983`(開局玩家設定)就是這樣讀的:
//
//	push    1Fh                        ; 31 格
//	mov     edx, 7                     ; RACESTUF.LBX 的 asset 7
//	mov     eax, offset aRacestufLbx
//	mov     edx, dword_19B7DC[ecx*4]   ; 第 ecx 族的陣列
//	add     eax, 89Fh                  ; → player+0x89F
//	call    sub_12779E                 ; memcpy
//
// ② **等級 → 真值的換算表**:執行檔 `byte_17D1F9`(檔案位移 0x1FB88D),10 列 × 3 級。
//    轉換迴圈只跑索引 **1..9**(`inc edx; cmp dx, 0Ah; jl`)——所以特性 0(政體)
//    與 10..30(布林)**存原值不換算**。這一點很重要:照 `[idx*3+level]` 一路換到 30
//    會把「水棲=1」變成「水棲=20」。
//
// ③ **展開後的真值**:`SAVE10.GAM` 裡五個種族(Alkari / Klackon / Mrrshan / Sakkra /
//    Trilarian)的 `Traits[31]`,由原版自己寫出來。①②推出來的結果與這五族**逐格相同**。
//
// 手冊行文再獨立對第四次:布拉西「+20 to Ship Attack」「+10 bonus in ground combat」、
// 埃雷里安「+25 defensive and +20 offensive」、達洛克「+20 more likely」、
// 阿爾卡里「+50 defensive」、姆瑞森 +50 攻擊——全部對上。
//
// ============ 誠實留白 ============
//
//   - **只有 13 個內建種族。** 自訂種族(Custom)走點數畫面自己組,不在這張表上。
//   - **政體(特性 0)的編號**與 `MoraleGovernmentType` / `AssimilationGovernment` /
//     `SpyGovernmentType` 同一套(第 54 項:原版只有一個 `[player+0x89F]`)。
//   - 特性 31(貧瘠母星)在列舉裡有,但陣列只有 31 格(0..30),原版放不下它。
//     不替它捏一格。

// RaceTraitCount 是原版特性陣列的長度(玩家結構 +0x89F 起 31 位元組)。
const RaceTraitCount = 31

// OrigRaceCount 是內建種族數(RACESTUF.LBX asset 7 表頭的第一個欄位)。
const OrigRaceCount = 13

// OrigRaceEnglishNames 依原版索引(= 字母序,也是 RACESEL 肖像順序)排列。
var OrigRaceEnglishNames = [OrigRaceCount]string{
	"Alkari",
	"Bulrathi",
	"Darlok",
	"Elerian",
	"Gnolam",
	"Human",
	"Klackon",
	"Meklar",
	"Mrrshan",
	"Psilon",
	"Sakkra",
	"Silicoid",
	"Trilarian"}

// OrigRaceChineseNames 與上表同序。
var OrigRaceChineseNames = [OrigRaceCount]string{
	"阿爾卡里",
	"布拉西",
	"達洛克",
	"埃雷里安",
	"諾蘭姆",
	"人類",
	"克拉肯",
	"梅克拉",
	"姆瑞森",
	"席隆",
	"薩克拉",
	"矽基",
	"崔拉里安"}

// origRacePickLevels 是 RACESTUF.LBX asset 7 的 13×31 選項等級(來源 ①)。
//
// 值域 0..3:0 = 沒選這一項;1..3 = 手冊 Race Picks 的三個檔位(1 檔通常是**扣分**選項)。
var origRacePickLevels = [OrigRaceCount][RaceTraitCount]int8{
	{2, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, //  0 Alkari 阿爾卡里
	{2, 0, 0, 0, 0, 0, 0, 2, 2, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, //  1 Bulrathi 布拉西
	{2, 0, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0}, //  2 Darlok 達洛克
	{0, 0, 0, 0, 0, 0, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0}, //  3 Elerian 埃雷里安
	{2, 0, 0, 0, 0, 3, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0}, //  4 Gnolam 諾蘭姆
	{4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, //  5 Human 人類
	{6, 0, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0}, //  6 Klackon 克拉肯
	{2, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, //  7 Meklar 梅克拉
	{2, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}, //  8 Mrrshan 姆瑞森
	{2, 0, 0, 0, 3, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0}, //  9 Psilon 席隆
	{0, 3, 2, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // 10 Sakkra 薩克拉
	{2, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0}, // 11 Silicoid 矽基
	{2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0}, // 12 Trilarian 崔拉里安
}

// raceTraitLevelValue 是執行檔 `byte_17D1F9` 的等級→真值換算表(來源 ②)。
//
// 只涵蓋特性 1..9(數值型);列 0 留空是因為特性 0 是政體,存原值。
// 每列四格對應等級 0..3,等級 0 恆為 0。
var raceTraitLevelValue = [10][4]int8{
	{0, 0, 0, 0},      // 特性 0(政體,存原值不換算)
	{0, -50, 50, 100}, // 特性 1 人口成長
	{0, -1, 2, 4},     // 特性 2 農業
	{0, -1, 1, 2},     // 特性 3 工業
	{0, -1, 1, 2},     // 特性 4 科研
	{0, -1, 1, 2},     // 特性 5 金錢
	{0, -20, 25, 50},  // 特性 6 艦艇防禦
	{0, -20, 20, 50},  // 特性 7 艦艇攻擊
	{0, -10, 10, 20},  // 特性 8 地面戰
	{0, -10, 10, 20},  // 特性 9 間諜
}

// origRaceTraits 是展開後的特性陣列,init 時由 origRacePickLevels + raceTraitLevelValue 算出。
//
// **刻意不寫死展開結果**:展開規則(哪些索引要換算、哪些存原值)本身就是這一項要釘的東西,
// 寫死等於把規則藏起來,`race_traits_test.go` 拿 SAVE10 的五族核對時就核不到規則了。
var origRaceTraits [OrigRaceCount][RaceTraitCount]int8

func init() {
	for r := range origRacePickLevels {
		origRaceTraits[r] = origRacePickLevels[r]
		// 只換算 1..9;特性 0(政體)與 10..30(布林)存原值,見檔頭 ②。
		for t := 1; t < len(raceTraitLevelValue); t++ {
			lv := origRacePickLevels[r][t]
			if lv <= 0 || int(lv) >= len(raceTraitLevelValue[t]) {
				origRaceTraits[r][t] = 0
				continue
			}
			origRaceTraits[r][t] = raceTraitLevelValue[t][lv]
		}
	}
}

// OrigRaceTraits 回傳某個內建種族展開後的 31 格特性陣列。
//
// 回傳的是複本,呼叫端改了不會污染表。
func OrigRaceTraits(race int) ([RaceTraitCount]int8, bool) {
	if race < 0 || race >= OrigRaceCount {
		return [RaceTraitCount]int8{}, false
	}
	return origRaceTraits[race], true
}

// OrigRaceTrait 回傳某族某項特性的值(查無回 0)。
//
// 數值型特性(1..9)回傳的是**已換算的真值**(如姆瑞森的艦艇攻擊 = 50);
// 布林特性(10..30)回傳 0 或 1;特性 0 是政體編號。
func OrigRaceTrait(race int, t RaceTrait) int {
	if race < 0 || race >= OrigRaceCount || t < 0 || int(t) >= RaceTraitCount {
		return 0
	}
	return int(origRaceTraits[race][t])
}

// OrigRaceHasTrait 回報某族有沒有某項布林特性。
//
// 只對布林特性(10..30)有意義;數值型特性請用 OrigRaceTrait 取值。
func OrigRaceHasTrait(race int, t RaceTrait) bool {
	return OrigRaceTrait(race, t) != 0
}

// OrigRacePickLevel 回傳某族某項特性的**選項等級**(0..3),供自訂種族畫面與文件對照。
func OrigRacePickLevel(race int, t RaceTrait) int {
	if race < 0 || race >= OrigRaceCount || t < 0 || int(t) >= RaceTraitCount {
		return 0
	}
	return int(origRacePickLevels[race][t])
}

// RaceTraitPickValue 回傳「某個數值型特性選到第 level 檔時的值」(手冊 Race Picks 三檔)。
//
// level 超出 0..3 或特性不是數值型時回 0。
func RaceTraitPickValue(t RaceTrait, level int) int {
	if t < 0 || int(t) >= len(raceTraitLevelValue) || level < 0 || level >= 4 {
		return 0
	}
	return int(raceTraitLevelValue[t][level])
}
