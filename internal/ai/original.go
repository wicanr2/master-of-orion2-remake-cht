// original.go:原版 MOO2 AI 的【權威資料】移植——難度加成表 + classic 種族性格分布。
//
// 與 economy.go/military.go/research.go/diplomacy.go 的【設計性重建】啟發式不同,本檔的兩張表都有
// 一手官方來源,不是本專案臆造:
//
//  1. DifficultyBonus:逐難度的 AI 加成表,來源 moo2_patch1.5/MANUAL_150.html「AI Opponents →
//     Generic AI bonuses」原文表格(手冊自己標註是「standard AI bonuses per difficulty level」,
//     視為 classic 基準值,1.50 只是新增可調 config 的能力,數字本身未變)。
//  2. classicRacePersonality:13 族開局性格分布,來源 patch1.5 內附的 AIRACES.CFG
//     `race_personality` 表 classic 對照值(`##`/`#` 註解後的原始數字;13 族中僅 Humans、Trilarians
//     的 mod 值與 classic 不同,已改用 classic)。
//
// 完整考據過程、可信度分級與交叉驗證見 docs/tech/original-ai-re.md 第 1-2 節。本檔只有查表資料,
// 沒有「性格分數如何逐回合轉成宣戰/求和」這類決策演算法——那部分官方手冊、社群、openorion2 三方都
// 沒有給出,不在本檔範圍(見 original-ai-re.md §4.2)。
package ai

// --- 難度加成表(docs/tech/original-ai-re.md §2.1,逐一比對 MANUAL_150.html 原始 <table> 核實) ---

// Difficulty 是原版 5 檔難度等級。
type Difficulty int

const (
	DifficultyTutor Difficulty = iota
	DifficultyEasy
	DifficultyAverage
	DifficultyHard
	DifficultyImpossible
)

// DifficultyBonus 是官方手冊「Generic AI bonuses」表的一列。
//
// 手冊原表部分欄位是分數(如 Food -1/4、Prod -1/2、BC 3/4),為避免浮點數失真,
// Food/Prod/Res/BC 四欄一律以「四分之幾」為單位存成整數(FoodQuarters=-1 代表 -1/4,
// ProdQuarters=-2 代表 -2/4=-1/2)。其餘欄位(GrowthPercent、CommandDeficitBC、SpyBonus、
// TroopsMarines、AntaranMarines)手冊原文就是整數,直接存。
type DifficultyBonus struct {
	// GrowthPercent 是人口成長率加成(手冊「Growth Percent」欄,整數,單位:百分點)。
	GrowthPercent int
	// FoodQuarters 是每農夫食物加成,單位 1/4(手冊「Food」欄)。
	FoodQuarters int
	// ProdQuarters 是每工人生產加成,單位 1/4(手冊「Prod」欄)。
	ProdQuarters int
	// ResQuarters 是每科學家研究加成,單位 1/4(手冊「Res」欄)。
	ResQuarters int
	// BCQuarters 是每人口 BC(稅收)加成,單位 1/4(手冊「BC」欄)。
	BCQuarters int
	// CommandDeficitBC 是超出 Command Rating 上限時,每艘超編艦隊的維護赤字 BC(手冊「Command
	// Deficit BC」欄)。
	CommandDeficitBC int
	// SpyBonus 是間諜行動加成(手冊「Spy Bonus」欄)。
	SpyBonus int
	// TroopsMarines 是地面部隊/陸戰隊戰力加成(手冊「Troops & Marines」欄)。
	TroopsMarines int
	// AntaranMarines 是 Antaran 陸戰隊戰力加成(手冊「Antaran Marines」欄)。
	AntaranMarines int
}

// Food 回傳 FoodQuarters 換算後的浮點值(= FoodQuarters/4),方便呼叫端直接套用到人口食物公式。
func (b DifficultyBonus) Food() float64 { return float64(b.FoodQuarters) / 4 }

// Prod 回傳 ProdQuarters 換算後的浮點值(= ProdQuarters/4)。
func (b DifficultyBonus) Prod() float64 { return float64(b.ProdQuarters) / 4 }

// Res 回傳 ResQuarters 換算後的浮點值(= ResQuarters/4)。
func (b DifficultyBonus) Res() float64 { return float64(b.ResQuarters) / 4 }

// BC 回傳 BCQuarters 換算後的浮點值(= BCQuarters/4)。
func (b DifficultyBonus) BC() float64 { return float64(b.BCQuarters) / 4 }

// difficultyBonusTable 逐一比對 scratchpad 內 MANUAL_150.html 解析出的原始 <table>(見
// docs/tech/original-ai-re.md §2.1)核實,5 檔難度全部核對無誤:
//
//	難度   Growth  Food  Prod  Res   BC   CmdDeficit  Spy  Troops  Antaran
//	Tutor    0    -1/4  -1/2  -1/2   0       12       -2    -2      -4
//	Easy    +1     0     0     0     0       11       -1    -1      -2
//	Avg     +2    1/4   1/2   1/2  1/4       10        0     0       0
//	Hard    +3    1/2    1     1   2/4        9        1     1       2
//	Imp     +4     1     2     2   3/4        8        2     2       4
var difficultyBonusTable = [5]DifficultyBonus{
	DifficultyTutor:      {GrowthPercent: 0, FoodQuarters: -1, ProdQuarters: -2, ResQuarters: -2, BCQuarters: 0, CommandDeficitBC: 12, SpyBonus: -2, TroopsMarines: -2, AntaranMarines: -4},
	DifficultyEasy:       {GrowthPercent: 1, FoodQuarters: 0, ProdQuarters: 0, ResQuarters: 0, BCQuarters: 0, CommandDeficitBC: 11, SpyBonus: -1, TroopsMarines: -1, AntaranMarines: -2},
	DifficultyAverage:    {GrowthPercent: 2, FoodQuarters: 1, ProdQuarters: 2, ResQuarters: 2, BCQuarters: 1, CommandDeficitBC: 10, SpyBonus: 0, TroopsMarines: 0, AntaranMarines: 0},
	DifficultyHard:       {GrowthPercent: 3, FoodQuarters: 2, ProdQuarters: 4, ResQuarters: 4, BCQuarters: 2, CommandDeficitBC: 9, SpyBonus: 1, TroopsMarines: 1, AntaranMarines: 2},
	DifficultyImpossible: {GrowthPercent: 4, FoodQuarters: 4, ProdQuarters: 8, ResQuarters: 8, BCQuarters: 3, CommandDeficitBC: 8, SpyBonus: 2, TroopsMarines: 2, AntaranMarines: 4},
}

// AIDifficultyBonus 回傳指定難度的官方加成表列。level 超出 Tutor..Impossible 範圍時,
// ok 回傳 false,並回傳零值(呼叫端應視為「無此難度」,不可當 Tutor 使用)。
func AIDifficultyBonus(level Difficulty) (bonus DifficultyBonus, ok bool) {
	if level < DifficultyTutor || level > DifficultyImpossible {
		return DifficultyBonus{}, false
	}
	return difficultyBonusTable[level], true
}

// --- 種族性格(docs/tech/original-ai-re.md §1.1、§1.3) ---

// Personality 是原版 AI 的性格代碼(AIRACES.CFG `race_personality` 表的 0-6)。
// 對照 estrings.tsv 官方字串:0-5 是「種族特質」,6(Dishonored)是「外交狀態」而非開局性格
// (見 original-ai-re.md §1.1),classic 分布表(§1.3)裡也確實沒有任何種族被指派 6。
type Personality int

const (
	PersonalityXenophobic Personality = iota // 0:排外
	PersonalityRuthless                      // 1:冷酷無情
	PersonalityAggressive                    // 2:好戰
	PersonalityErratic                       // 3:反覆無常
	PersonalityHonorable                     // 4:重信譽
	PersonalityPacifist                      // 5:和平主義
	PersonalityDishonored                    // 6:失信(外交狀態,非開局性格,見上)
)

// String 回傳性格英文名(對齊 estrings.tsv 分類用詞)。
func (p Personality) String() string {
	switch p {
	case PersonalityXenophobic:
		return "Xenophobic"
	case PersonalityRuthless:
		return "Ruthless"
	case PersonalityAggressive:
		return "Aggressive"
	case PersonalityErratic:
		return "Erratic"
	case PersonalityHonorable:
		return "Honorable"
	case PersonalityPacifist:
		return "Pacifist"
	case PersonalityDishonored:
		return "Dishonored"
	default:
		return "Unknown"
	}
}

// classicRacePersonality 是 13 族的 classic 性格分布(AIRACES.CFG `race_personality` 表,
// `##`/`#` 後的 classic 對照值——不是 `=` 後的 1.50 improved mod 值)。每族 10 格
// (random0..random9),每格是一個 Personality(0-6)。開局用公式
// `column := Random(10) + 1 - difficulty_byte(0-4)` 抽出 1-10 的欄位(見
// docs/tech/original-ai-re.md §1.3;clamp 規則未經官方驗證),取 table[race][column-1]。
//
// 13 族中僅 Humans、Trilarians 的 `=` mod 值與 `##` classic 值不同(其餘 11 族 mod 值本身就等於
// classic,CFG 用單一 `#` 標記「沒有變動」);下表全部取 classic 值,已逐行核對
// scratchpad/patch15/AIRACES.CFG 第 11-23 行原始內容,與 mod 值刻意不同的两族已標註。
var classicRacePersonality = map[string][10]Personality{
	// 3 4 4 4 4 4 4 4 5 5 → Erratic 10% / Honorable 70% / Pacifist 20%
	"Alkari": {3, 4, 4, 4, 4, 4, 4, 4, 5, 5},
	// 1 2 2 2 2 2 2 2 3 3 → Ruthless 10% / Aggressive 70% / Erratic 20%
	"Bulrathi": {1, 2, 2, 2, 2, 2, 2, 2, 3, 3},
	// 1 1 0 2 2 2 2 2 2 2 → Xenophobic 10% / Ruthless 20% / Aggressive 70%
	"Darloks": {1, 1, 0, 2, 2, 2, 2, 2, 2, 2},
	// 0 0 1 1 1 1 1 2 2 2 → Xenophobic 20% / Ruthless 50% / Aggressive 30%
	"Elerians": {0, 0, 1, 1, 1, 1, 1, 2, 2, 2},
	// 1 1 2 2 3 3 3 5 5 5 → Ruthless 20% / Aggressive 20% / Erratic 30% / Pacifist 30%
	"Gnolams": {1, 1, 2, 2, 3, 3, 3, 5, 5, 5},
	// classic(CFG `##` 後);mod 值改為 3 3 3 4 4 4 4 4 5 5(mod≠classic 的兩族之一,已用 classic)。
	// → Erratic 10% / Honorable 70% / Pacifist 20%
	"Humans": {3, 4, 4, 4, 4, 4, 4, 4, 5, 5},
	// 0 0 0 0 0 0 0 1 2 2 → Xenophobic 70% / Ruthless 10% / Aggressive 20%
	"Klackons": {0, 0, 0, 0, 0, 0, 0, 1, 2, 2},
	// 0 2 2 3 3 3 3 3 3 3 → Xenophobic 10% / Aggressive 20% / Erratic 70%
	"Meklars": {0, 2, 2, 3, 3, 3, 3, 3, 3, 3},
	// 0 1 1 1 1 1 1 1 2 2 → Xenophobic 10% / Ruthless 70% / Aggressive 20%
	"Mrrshan": {0, 1, 1, 1, 1, 1, 1, 1, 2, 2},
	// 3 3 4 5 5 5 5 5 5 5 → Erratic 20% / Honorable 10% / Pacifist 70%
	"Psilons": {3, 3, 4, 5, 5, 5, 5, 5, 5, 5},
	// 1 2 2 2 2 2 2 2 3 3 → Ruthless 10% / Aggressive 70% / Erratic 20%
	"Sakkra": {1, 2, 2, 2, 2, 2, 2, 2, 3, 3},
	// 0 0 0 0 0 0 0 2 2 3 → Xenophobic 70% / Aggressive 20% / Erratic 10%
	"Silicoids": {0, 0, 0, 0, 0, 0, 0, 2, 2, 3},
	// classic(CFG `##` 後);mod 值改為 3 3 4 4 4 5 5 5 5 5(mod≠classic 的兩族之二,已用 classic)。
	// → Erratic 10% / Honorable 50% / Pacifist 40%
	"Trilarians": {3, 4, 4, 4, 4, 4, 5, 5, 5, 5},
}

// ClassicRacePersonality 回傳指定種族的 classic 性格分布(10 格,random0..random9)。
// race 對照 AIRACES.CFG 的英文種族名(如 "Alkari"、"Humans"),大小寫需完全相符。
// 查無此族時 ok 回傳 false。
func ClassicRacePersonality(race string) (dist [10]Personality, ok bool) {
	dist, ok = classicRacePersonality[race]
	return dist, ok
}
