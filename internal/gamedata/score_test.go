package gamedata

import "testing"

// score_test.go:最終得分。八個係數全部來自反組譯的 `Get_*_Score_` 函式(見 score.go 檔頭),
// 每條測試都標明它釘的是哪一條指令,以及手冊 p.184 的哪一句定性描述。

// 原版 `Get_Time_Score_`:nPlayers × (20 × (星圖大小+1) + 80) − 已過回合數。
// 手冊三句話全在這一條公式裡:越快贏分越高、星圖越大分越高、種族越多分越高。
func TestScoreTimeFormula(t *testing.T) {
	base := ScoreInput{GalaxySizeIndex: 1, RaceCount: 4, TurnsElapsed: 100, Population: 10}
	// 4 × (20×2 + 80) − 100 = 4×120 − 100 = 380
	if got := ScoreTime(base); got != 380 {
		t.Errorf("時間分 = %d,want 380", got)
	}
	// 越快贏分越高。
	faster := base
	faster.TurnsElapsed = 50
	if ScoreTime(faster) <= ScoreTime(base) {
		t.Error("回合數越少分數應越高(手冊:the quicker you win, the higher your score)")
	}
	// 星圖越大分越高。
	bigger := base
	bigger.GalaxySizeIndex = 3
	if ScoreTime(bigger) <= ScoreTime(base) {
		t.Error("星圖越大分數應越高(手冊:playing in a larger galaxy results in a score increase)")
	}
	// 種族越多分越高。
	crowded := base
	crowded.RaceCount = 8
	if ScoreTime(crowded) <= ScoreTime(base) {
		t.Error("種族越多分數應越高(手冊:as the number of races … goes up, so does your score)")
	}
	// 人口 0 → 整項歸零(原版 `cmp word [score+0A0h], 0 / jnz`)。
	dead := base
	dead.Population = 0
	if got := ScoreTime(dead); got != 0 {
		t.Errorf("人口 0 時時間分應為 0,實得 %d", got)
	}
}

// 原版 `Get_Captured_Colonist_Score_`:俘虜人口 × 2 ÷ (星圖大小 + 1)。
// 手冊:「This premium is higher in smaller galaxies」——除數變小、分數變高。
func TestScoreCapturedHigherInSmallerGalaxies(t *testing.T) {
	small := ScoreInput{GalaxySizeIndex: 0, CapturedPopulation: 12}
	big := ScoreInput{GalaxySizeIndex: 3, CapturedPopulation: 12}
	if got := ScoreCaptured(small); got != 24 { // 12×2÷1
		t.Errorf("小星圖俘虜分 = %d,want 24", got)
	}
	if got := ScoreCaptured(big); got != 6 { // 12×2÷4
		t.Errorf("大星圖俘虜分 = %d,want 6", got)
	}
	if ScoreCaptured(small) <= ScoreCaptured(big) {
		t.Error("小星圖的俘虜加分應較高(手冊逐字)")
	}
}

// 原版 `Get_Technology_Score_`:3 × 已知主題 + 5 × Hyper-Advanced 等級。
// 手冊:「First level Hyper-Advanced technologies are worth more points than normal ones」
// ——5 > 3,係數本身就是那句話。
func TestScoreTechnologyHyperWorthMore(t *testing.T) {
	in := ScoreInput{ResearchTopicsKnown: 10, HyperAdvancedLevels: 4}
	if got := ScoreTechnology(in); got != 10*3+4*5 {
		t.Errorf("科技分 = %d,want %d", got, 10*3+4*5)
	}
	if ScorePerHyperAdvancedLevel <= ScorePerResearchTopic {
		t.Errorf("Hyper-Advanced 每級 %d 應高於一般主題每項 %d",
			ScorePerHyperAdvancedLevel, ScorePerResearchTopic)
	}
}

// 三項一次性大分的**相對大小**是手冊定性描述的直接對照:
// 安塔蘭「biggest point bonus of all」> 獵戶座「big chunk」= 議會「substantial addition」
// > 殲滅一族「a boost」。
func TestScoreBonusOrderingMatchesManual(t *testing.T) {
	if !(ScoreAntaranVictory > ScoreOrionCaptured) {
		t.Errorf("安塔蘭 %d 應大於獵戶座 %d", ScoreAntaranVictory, ScoreOrionCaptured)
	}
	if ScoreOrionCaptured != ScoreCouncilVictory {
		t.Errorf("獵戶座 %d 與議會 %d 在原版是同一個數字(都是 0x64)",
			ScoreOrionCaptured, ScoreCouncilVictory)
	}
	if !(ScoreCouncilVictory > ScorePerRaceEliminated) {
		t.Errorf("議會 %d 應大於殲滅一族 %d", ScoreCouncilVictory, ScorePerRaceEliminated)
	}
	// 原版硬編值,直接釘住。
	if ScoreAntaranVictory != 250 || ScoreOrionCaptured != 100 ||
		ScoreCouncilVictory != 100 || ScorePerRaceEliminated != 50 {
		t.Errorf("一次性加分被改動了:安塔蘭%d 獵戶座%d 議會%d 殲滅%d",
			ScoreAntaranVictory, ScoreOrionCaptured, ScoreCouncilVictory, ScorePerRaceEliminated)
	}
}

// 原版先加總八項，再套種族倍率並以 +50/100 四捨五入；沒有自訂負分夾限。
func TestComputeScoreTotalAndMultiplier(t *testing.T) {
	in := ScoreInput{
		GalaxySizeIndex: 1, RaceCount: 4, TurnsElapsed: 100,
		Population: 30, CapturedPopulation: 6, ResearchTopicsKnown: 20,
		HyperAdvancedLevels: 2, RacesEliminated: 1,
		OrionCaptured: true, CouncilVictory: true, AntaranVictory: true,
	}
	b := ComputeScore(in)
	want := b.Time + b.Population + b.Captured + b.Technology + b.Elimination + b.Orion + b.Council + b.Antares
	if b.RawTotal != want || b.Total != want || b.MultiplierPercent != 100 {
		t.Errorf("100%% 總分分層錯誤:%+v, raw want %d", b, want)
	}
	if b.Orion != 100 || b.Council != 100 || b.Antares != 250 {
		t.Errorf("一次性加分沒接上:%+v", b)
	}

	scaled := ComputeScore(ScoreInput{Population: 101, MultiplierPercent: 150})
	if scaled.RawTotal != 101 || scaled.Total != 152 {
		t.Errorf("(101*150+50)/100 應四捨五入為 152,got %+v", scaled)
	}
	slow := ScoreInput{GalaxySizeIndex: 0, RaceCount: 2, TurnsElapsed: 100000, Population: 1}
	if got := ComputeScore(slow).Total; got >= 0 {
		t.Errorf("原版沒有負分 clamp,實得 %d", got)
	}
}
