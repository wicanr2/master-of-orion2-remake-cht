package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
)

func TestOriginalPlagueResearchNeed(t *testing.T) {
	for _, tc := range []struct {
		research, difficulty, roll, want int
	}{
		{10, 0, 1, 10},
		{10, 2, 8, 120},
		{7, 4, 3, 77},
	} {
		got, ok := originalPlagueResearchNeed(tc.research, tc.difficulty, tc.roll)
		if !ok || got != tc.want {
			t.Errorf("need(%d,%d,%d)=%d,%v want %d", tc.research, tc.difficulty, tc.roll, got, ok, tc.want)
		}
	}
	if _, ok := originalPlagueResearchNeed(0, 2, 1); ok {
		t.Fatal("建立時研究產出為 0 的殖民地不能啟動瘟疫 record")
	}
}

func TestStartPlayerPlagueUsesCurrentResearchAndSavedRoll(t *testing.T) {
	s := NewDemoSession()
	s.Difficulty = 2
	s.PlayerColonies = s.PlayerColonies[:1]
	s.PlayerColonyStars = s.PlayerColonyStars[:1]
	s.PlayerColonyPlanets = s.PlayerColonyPlanets[:1]
	s.LastPlayerOutput = engine.EmpireOutput{Colonies: []engine.ColonyOutput{{Research: 10}}}
	seed := int64(71)
	probe := newRandStream(seed)
	_ = reservoirEventColony(s.PlayerColonies, probe)
	wantRoll := probe.Intn(8) + 1
	s.eventRand = newRandStream(seed)
	idx, need, ok := s.startPlayerPlague()
	if !ok || idx != 0 || need != 10*(wantRoll+4) {
		t.Fatalf("建立瘟疫錯誤：idx=%d need=%d ok=%v wantNeed=%d", idx, need, ok, 10*(wantRoll+4))
	}
	if len(s.PersistentEvents) != 1 || s.PersistentEvents[0].Turns != -1 ||
		s.PersistentEvents[0].ResearchDone != 0 || s.PersistentEvents[0].ResearchNeeded != need {
		t.Fatalf("建立 record 錯誤：%+v", s.PersistentEvents)
	}
}

func TestPlagueSubtractsTwoHundredGrowthPercentagePoints(t *testing.T) {
	s := NewDemoSession()
	planet := s.ColonyPlanetIndex(0)
	base := s.PlayerColonies[0].GrowthBonusSum
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentPlague, PlanetIndex: planet, ResearchNeeded: 100}}
	turn := s.coloniesForTurn()
	if turn[0].GrowthBonusSum != base-200 {
		t.Fatalf("瘟疫應扣 200 成長百分點：got %d want %d", turn[0].GrowthBonusSum, base-200)
	}
	if s.PlayerColonies[0].GrowthBonusSum != base {
		t.Fatal("瘟疫不得污染殖民地永久 GrowthBonusSum")
	}
	if got := engine.RunColonyTurn(turn[0]).PopGrowth; got >= 0 {
		t.Fatalf("一般殖民地在 -200 百分點下應形成負成長，got %d", got)
	}
}

func TestPlagueResearchProgressAndCure(t *testing.T) {
	s := NewDemoSession()
	planet := s.ColonyPlanetIndex(0)
	e := PersistentEvent{Kind: PersistentPlague, PlanetIndex: planet, ResearchNeeded: 12}
	s.PersistentEvents = []PersistentEvent{e}
	s.recordPlagueResearch(planet, 5)
	if s.PersistentEvents[0].ResearchDone != 5 {
		t.Fatalf("治療進度應為 5，got %d", s.PersistentEvents[0].ResearchDone)
	}
	done, _, _ := s.stepPersistentEvent(&s.PersistentEvents[0])
	if done {
		t.Fatal("研究進度未達需求時不得解除")
	}
	s.recordPlagueResearch(planet, 7)
	done, _, _ = s.stepPersistentEvent(&s.PersistentEvents[0])
	if !done {
		t.Fatal("研究進度達需求時必須解除")
	}
}

func TestPlagueAndPopulationBoomAreMutuallyExclusive(t *testing.T) {
	s := NewDemoSession()
	s.eventRand = newRandStream(3)
	planet := s.ColonyPlanetIndex(0)
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentPopulationBoom, PlanetIndex: planet}}
	if _, ok := s.plagueTargetEligible(s.PlayerColonies[:1], func(int) int { return planet }); ok {
		t.Fatal("同一行星已有事件 17 時不得再建立事件 16")
	}
}

func TestAIPlagueUsesTransientTurnCopy(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) == 0 || len(s.AIPlayers[0].Colonies) == 0 || len(s.AIPlayers[0].ColonyPlanets) == 0 {
		t.Fatal("demo 應有 AI 殖民地")
	}
	planet := s.AIPlayers[0].ColonyPlanets[0]
	base := s.AIPlayers[0].Colonies[0].GrowthBonusSum
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentPlague, PlanetIndex: planet, ResearchNeeded: 100}}
	turn := s.aiColoniesForTurn(0, s.AIPlayers[0].Colonies)
	if turn[0].GrowthBonusSum != base-200 || s.AIPlayers[0].Colonies[0].GrowthBonusSum != base {
		t.Fatalf("AI 瘟疫副本錯誤：turn=%d stored=%d", turn[0].GrowthBonusSum,
			s.AIPlayers[0].Colonies[0].GrowthBonusSum)
	}
}

func TestPlagueAdvancesThroughNormalEndTurn(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	planet := s.ColonyPlanetIndex(0)
	s.PersistentEvents = []PersistentEvent{{
		Kind: PersistentPlague, PlanetIndex: planet, ResearchNeeded: 10000,
	}}
	s.EndTurn()
	if len(s.PersistentEvents) != 1 || s.PersistentEvents[0].ResearchDone <= 0 {
		t.Fatalf("正常 EndTurn 應推進瘟疫研究：%+v", s.PersistentEvents)
	}
	if len(s.LastPlayerOutput.Colonies) == 0 || s.LastPlayerOutput.Colonies[0].PopGrowth >= 0 {
		t.Fatalf("正常 EndTurn 應套用瘟疫負成長：%+v", s.LastPlayerOutput.Colonies)
	}
}
