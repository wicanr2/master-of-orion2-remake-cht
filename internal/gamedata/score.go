package gamedata

// score.go:最終得分(Hi-Score / Hall of Fame)。
//
// 手冊 p.184「Score」列了八條計分因素,但**一個數字都沒給**。原版 module 60 有一整組
// `Get_*_Score_` 函式,每個都短到可以逐指令讀完,八條因素的係數全在裡面:
//
//	Get_Time_Score_               @ 0x9E993
//	Get_Population_Score_         @ 0x9E8C7
//	Get_Captured_Colonist_Score_  @ 0x9E735
//	Get_Technology_Score_         @ 0x9E973(+ 計數輔助 sub_9E90B)
//	Get_Elimination_Score_        @ 0x9E84C
//	Get_Orion_Score_              @ 0x9E8A3
//	Get_Won_Council_Score_        @ 0x9EA17
//	Get_Antares_Score_            @ 0x9E711
//
// **八條全部與手冊的定性描述對得上**,包括相對大小:安塔蘭 250 > 獵戶座 100 = 議會 100 >
// 殲滅一族 50,正好是手冊說的「biggest point bonus of all」>「big chunk」/「substantial
// addition」>「a boost」。俘虜人口的除數是「星圖大小 + 1」,也正好對上手冊
// 「This premium is higher in smaller galaxies」。
//
// 另外兩個順帶的交叉驗證:科技分數掃的是 `player+0xC4` 起、長度 0x53(83)的研究主題陣列
// ——0xC4 這個偏移與 `Do_System_Discoveries_At_Star_` 讀遠古文物時用的是同一個,83 也與
// remake 既有的研究主題數相同;時間分數用 `word[0x192FD8] - 0x88B8` 算已過回合,
// 0x88B8 = 35000 = 星曆 3500.0 ×10,正是遊戲的起始星曆。

// 得分係數(全部來自上列函式的指令,非估值)。
const (
	// ScoreTimeBasePerRace 是每個種族貢獻的基礎分裡的常數項(原版 `add eax, 50h`)。
	ScoreTimeBasePerRace = 80
	// ScoreTimeGalaxySizeFactor 是星圖大小的乘數(原版 `imul eax, 14h`)。
	ScoreTimeGalaxySizeFactor = 20
	// ScorePerPopulation 是每單位人口的分數(原版直接把人口加進總和)。
	ScorePerPopulation = 1
	// ScoreCapturedPopNumerator 是俘虜人口的分子(原版 `add eax, eax` = ×2)。
	ScoreCapturedPopNumerator = 2
	// ScorePerResearchTopic 是每個已完成研究主題的分數(原版 `imul …, 3`)。
	ScorePerResearchTopic = 3
	// ScorePerHyperAdvancedLevel 是每級 Hyper-Advanced 科技的分數(原版 `imul …, 5`)。
	ScorePerHyperAdvancedLevel = 5
	// ScorePerRaceEliminated 是每殲滅一個種族的分數(原版 `add ebx, 32h`)。
	ScorePerRaceEliminated = 50
	// ScoreOrionCaptured 是攻下獵戶座的分數(原版 `mov eax, 64h`)。
	ScoreOrionCaptured = 100
	// ScoreCouncilVictory 是議會勝利的分數(原版 `mov eax, 64h`)。
	ScoreCouncilVictory = 100
	// ScoreAntaranVictory 是擊敗安塔蘭的分數(原版 `mov eax, 0FAh`)。
	ScoreAntaranVictory = 250
)

// ScoreInput 是算一個帝國最終得分所需的全部資料。
type ScoreInput struct {
	// GalaxySizeIndex 是星圖大小的索引(原版 `_galaxy_size` 之類,0 = 最小)。
	// 時間分與俘虜人口分都用「索引 + 1」。
	GalaxySizeIndex int
	// RaceCount 是這局的種族數(不含安塔蘭)。
	RaceCount int
	// TurnsElapsed 是已經過的回合數(原版用星曆差算:目前星曆 − 起始 3500)。
	TurnsElapsed int
	// Population 是自己所有殖民地的人口總和。
	Population int
	// CapturedPopulation 是俘虜來的人口單位數。
	CapturedPopulation int
	// ResearchTopicsKnown 是已完成的研究主題數。
	ResearchTopicsKnown int
	// HyperAdvancedLevels 是已達成的 Hyper-Advanced 等級數(原版掃 8 個研究領域)。
	HyperAdvancedLevels int
	// RacesEliminated 是自己殲滅掉的種族數。
	RacesEliminated int
	// OrionCaptured / CouncilVictory / AntaranVictory 是三項一次性大分。
	OrionCaptured  bool
	CouncilVictory bool
	AntaranVictory bool
	// MultiplierPercent 是 sub_58F4A 依未使用種族 Picks 算出的百分比；零值供舊呼叫端
	// 安全回退 100。Evolutionary Mutation 尚未消費的 4 Picks 由 shell 層加入。
	MultiplierPercent int
}

// ScoreBreakdown 是逐項得分(供 Hi-Score 畫面逐列顯示,原版的 `Draw_*_Score_` 也是逐項畫的)。
type ScoreBreakdown struct {
	Time              int
	Population        int
	Captured          int
	Technology        int
	Elimination       int
	Orion             int
	Council           int
	Antares           int
	RawTotal          int
	MultiplierPercent int
	Total             int
}

// ScoreTime 依原版 `Get_Time_Score_` 算時間分。
//
//	nPlayers × (20 × (星圖大小+1) + 80) − 已過回合數
//
// **人口為 0 時整項歸零**(原版:`cmp word [score+0A0h], 0 / jnz` ——那個欄位正是人口分)。
// 語意是「已經滅亡的帝國拿不到時間分」。
func ScoreTime(in ScoreInput) int {
	if in.Population <= 0 {
		return 0
	}
	perRace := ScoreTimeGalaxySizeFactor*(in.GalaxySizeIndex+1) + ScoreTimeBasePerRace
	return in.RaceCount*perRace - in.TurnsElapsed
}

// ScorePopulation 依原版 `Get_Population_Score_`:自己所有殖民地的人口總和。
func ScorePopulation(in ScoreInput) int { return in.Population * ScorePerPopulation }

// ScoreCaptured 依原版 `Get_Captured_Colonist_Score_`:俘虜人口 × 2 ÷ (星圖大小 + 1)。
// 星圖越小除數越小、分數越高——對上手冊「This premium is higher in smaller galaxies」。
func ScoreCaptured(in ScoreInput) int {
	div := in.GalaxySizeIndex + 1
	if div < 1 {
		div = 1
	}
	return in.CapturedPopulation * ScoreCapturedPopNumerator / div
}

// ScoreTechnology 依原版 `Get_Technology_Score_`:3 × 已知主題 + 5 × Hyper-Advanced 等級。
// 係數 5 > 3 對上手冊「First level Hyper-Advanced technologies are worth more points
// than normal ones」。
func ScoreTechnology(in ScoreInput) int {
	return in.ResearchTopicsKnown*ScorePerResearchTopic + in.HyperAdvancedLevels*ScorePerHyperAdvancedLevel
}

// ScoreElimination 依原版 `Get_Elimination_Score_`:每殲滅一族 50 分。
func ScoreElimination(in ScoreInput) int { return in.RacesEliminated * ScorePerRaceEliminated }

// ComputeScore 算出完整的逐項得分與總分。
func ComputeScore(in ScoreInput) ScoreBreakdown {
	b := ScoreBreakdown{
		Time:        ScoreTime(in),
		Population:  ScorePopulation(in),
		Captured:    ScoreCaptured(in),
		Technology:  ScoreTechnology(in),
		Elimination: ScoreElimination(in),
	}
	if in.OrionCaptured {
		b.Orion = ScoreOrionCaptured
	}
	if in.CouncilVictory {
		b.Council = ScoreCouncilVictory
	}
	if in.AntaranVictory {
		b.Antares = ScoreAntaranVictory
	}
	b.RawTotal = b.Time + b.Population + b.Captured + b.Technology + b.Elimination +
		b.Orion + b.Council + b.Antares
	b.MultiplierPercent = in.MultiplierPercent
	if b.MultiplierPercent <= 0 {
		b.MultiplierPercent = 100
	}
	// Score orchestrator @ 0x9DAF8..0x9DB14：乘百分比、+50、再除 100。
	// 不自行夾成零；原版此路徑沒有 clamp。
	b.Total = (b.RawTotal*b.MultiplierPercent + 50) / 100
	return b
}
