package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// score_test.go:計分公式接到實際對局狀態上的那一層(係數本身由 gamedata/score_test.go 釘住)。

func TestFinalScoreReflectsGameState(t *testing.T) {
	s := NewDemoSession()
	b := s.FinalScore()

	// 開局:有人口、有起始科技,總分應為正。
	if b.Population <= 0 {
		t.Errorf("開局人口分應 > 0,實得 %d", b.Population)
	}
	if b.Total <= 0 {
		t.Errorf("開局總分應 > 0,實得 %d", b.Total)
	}
	// 三項一次性大分開局都還沒拿到。
	if b.Orion != 0 || b.Council != 0 || b.Antares != 0 {
		t.Errorf("開局不該有一次性加分:%+v", b)
	}

	// 人口增加 → 人口分增加。
	before := s.FinalScore().Population
	s.PlayerColonies[0].Population += 5
	if got := s.FinalScore().Population; got != before+5 {
		t.Errorf("人口 +5 後人口分應 %d,實得 %d", before+5, got)
	}

	// 回合數增加 → 時間分下降(手冊:每回合扣分)。
	tBefore := s.FinalScore().Time
	s.Turn += 30
	if got := s.FinalScore().Time; got != tBefore-30 {
		t.Errorf("30 回合後時間分應 %d,實得 %d", tBefore-30, got)
	}
}

// 擊敗安塔蘭母星要拿到 250 分那一項(remake 既有的 AntaranHomeworldConquered 旗標)。
func TestFinalScoreAntaranBonus(t *testing.T) {
	s := NewDemoSession()
	before := s.FinalScore().Total
	s.AntaranHomeworldConquered = true
	got := s.FinalScore()
	if got.Antares != gamedata.ScoreAntaranVictory {
		t.Errorf("安塔蘭加分 = %d,want %d", got.Antares, gamedata.ScoreAntaranVictory)
	}
	if got.Total != before+gamedata.ScoreAntaranVictory {
		t.Errorf("總分應增加 %d:%d → %d", gamedata.ScoreAntaranVictory, before, got.Total)
	}
}

// 俘虜人口:地面入侵佔領敵殖民地時要累計(手冊 p.184「premium for captured population units」)。
func TestCapturedPopulationCountedOnInvasion(t *testing.T) {
	s := NewDemoSession()
	if s.CapturedPop != 0 {
		t.Fatalf("開局俘虜人口應為 0,實為 %d", s.CapturedPop)
	}
	// 直接驗接線點:CapturedPop 有進計分輸入。
	s.CapturedPop = 8
	want := gamedata.ScoreCaptured(gamedata.ScoreInput{
		GalaxySizeIndex: s.galaxySizeIndex(), CapturedPopulation: 8,
	})
	if got := s.FinalScore().Captured; got != want {
		t.Errorf("俘虜人口分 = %d,want %d", got, want)
	}
	if want == 0 {
		t.Error("測試前提不成立:俘虜 8 人口應該算得出分數")
	}
}

// ScoreLines 是畫面用的逐列資料:最後一列必須是總分,且與 FinalScore 一致。
func TestScoreLinesLastIsTotal(t *testing.T) {
	s := NewDemoSession()
	lines := s.ScoreLines()
	if len(lines) != 9 {
		t.Fatalf("得分列數 = %d,want 9(八個分項 + 總分)", len(lines))
	}
	last := lines[len(lines)-1]
	if last.Label != "總分" {
		t.Errorf("最後一列應是總分,實為 %q", last.Label)
	}
	if last.Value != s.FinalScore().Total {
		t.Errorf("總分列 %d 與 FinalScore %d 不符", last.Value, s.FinalScore().Total)
	}
}

func TestFinalScoreSumsRepeatedHyperAdvancedLevels(t *testing.T) {
	s := NewDemoSession()
	s.Player.HyperAdvancedLevels = map[gamedata.ResearchTopic]int{
		gamedata.TOPIC_HYPER_PHYSICS: 3,
		gamedata.TOPIC_HYPER_FIELDS:  2,
	}
	if got := s.hyperAdvancedLevels(); got != 5 {
		t.Fatalf("Hyper 分數 consumer 應加總八個 level byte，得到 %d want 5", got)
	}
}

func TestFinalScoreUsesUnusedRacePicksAndEvolutionaryMutation(t *testing.T) {
	s := NewDemoSession()
	s.SetCustomRaceUnusedPicks(5)
	base := s.FinalScore()
	if base.MultiplierPercent != 150 || base.Total != (base.RawTotal*150+50)/100 {
		t.Fatalf("剩餘 5 Picks 應套 150%%:%+v", base)
	}
	grantTechnologyApplication(&s.Player, gamedata.TOPIC_TRANS_GENETICS, gamedata.TECH_EVOLUTIONARY_MUTATION)
	evolved := s.FinalScore()
	if evolved.MultiplierPercent != 190 || evolved.Total != (evolved.RawTotal*190+50)/100 {
		t.Fatalf("Evolutionary Mutation 未消費 4 Picks 應再加 40%%:%+v", evolved)
	}
}

// 星圖大小索引要跟著實際星數走(時間分與俘虜分都用它)。
func TestGalaxySizeIndexTracksStarCount(t *testing.T) {
	s := NewDemoSession()
	for i, g := range GalaxySizes {
		s.SetupNewGame(g.Stars, int64(100+i), 2)
		if got := s.galaxySizeIndex(); got != i {
			t.Errorf("%d 星的星圖大小索引 = %d,want %d(%s)", g.Stars, got, i, g.Name)
		}
	}
}
