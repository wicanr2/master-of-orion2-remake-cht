package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestPopulationBoomTargetDoesNotRetryRejectedColony(t *testing.T) {
	s := NewDemoSession()
	colonies := []engine.ColonyState{{Population: 10, PopMax: 10}, {Population: 1, PopMax: 10}}
	planetAt := func(i int) int { return i + 100 }
	seed := int64(-1)
	for n := int64(0); n < 1000; n++ {
		if reservoirEventColony(colonies, newRandStream(n)) == 0 {
			seed = n
			break
		}
	}
	if seed < 0 {
		t.Fatal("找不到抽中滿人口殖民地的固定 seed")
	}
	s.eventRand = newRandStream(seed)
	if _, ok := s.populationBoomTargetEligible(colonies, planetAt); ok {
		t.Fatal("原版只抽一次；抽到滿人口殖民地時不得改選另一座")
	}
}

func TestPopulationBoomAddsOneHundredGrowthPercentagePoints(t *testing.T) {
	s := NewDemoSession()
	idx := 0
	planet := s.ColonyPlanetIndex(idx)
	baseBonus := s.PlayerColonies[idx].GrowthBonusSum
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentPopulationBoom, PlanetIndex: planet}}
	turnColonies := s.coloniesForTurn()
	if got := turnColonies[idx].GrowthBonusSum; got != baseBonus+100 {
		t.Fatalf("人口暴增應增加 100 成長百分點：got %d want %d", got, baseBonus+100)
	}
	if s.PlayerColonies[idx].GrowthBonusSum != baseBonus {
		t.Fatal("事件加成只能存在於回合輸入副本，不得永久污染殖民地基礎值")
	}

	plain := s.PlayerColonies[idx]
	boosted := turnColonies[idx]
	plainOut := engine.RunColonyTurn(plain)
	boostedOut := engine.RunColonyTurn(boosted)
	base := gamedata.ColonyBaseGrowth(plain.Population, plain.Population, plain.PopMax)
	wantDelta := gamedata.ColonyGrowth(base, baseBonus+100) - gamedata.ColonyGrowth(base, baseBonus)
	if got := boostedOut.PopGrowth - plainOut.PopGrowth; got != wantDelta {
		t.Fatalf("成長輸出未按同一公式增加 100 百分點：got %d want %d", got, wantDelta)
	}
}

func TestPopulationBoomDurationBoundaries(t *testing.T) {
	s := NewDemoSession()
	planet := s.ColonyPlanetIndex(0)
	s.eventRand = newRandStream(19)
	e := PersistentEvent{Kind: PersistentPopulationBoom, PlanetIndex: planet, Turns: 5}
	done, _, _ := s.stepPersistentEvent(&e)
	if done {
		t.Fatal("前五個 active turn 結束後仍不得提前終止")
	}
	e.Turns = 22
	done, _, _ = s.stepPersistentEvent(&e)
	if !done {
		t.Fatal("age > 20 的強制終止分支必須生效")
	}
}

func TestPopulationBoomAIUsesTransientTurnCopy(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) == 0 || len(s.AIPlayers[0].Colonies) == 0 || len(s.AIPlayers[0].ColonyPlanets) == 0 {
		t.Fatal("demo 應有具行星對映的 AI 殖民地")
	}
	base := s.AIPlayers[0].Colonies[0].GrowthBonusSum
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentPopulationBoom, PlanetIndex: s.AIPlayers[0].ColonyPlanets[0]}}
	got := s.aiColoniesForTurn(0, s.AIPlayers[0].Colonies)
	// Demo 難度 1 另有官方 AI 成長 +1；事件本身再加 100，兩者同在暫態副本疊加。
	if got[0].GrowthBonusSum != base+101 || s.AIPlayers[0].Colonies[0].GrowthBonusSum != base {
		t.Fatalf("AI 回合副本錯誤：turn=%d stored=%d base=%d", got[0].GrowthBonusSum,
			s.AIPlayers[0].Colonies[0].GrowthBonusSum, base)
	}
}

func TestAIColoniesForTurnAddsDifficultyWithoutPersisting(t *testing.T) {
	s := NewDemoSession()
	s.Difficulty = int(ai.DifficultyHard)
	base := s.AIPlayers[0].Colonies[0]
	turn := s.aiColoniesForTurn(0, s.AIPlayers[0].Colonies)
	if len(turn) == 0 {
		t.Fatal("AI 殖民地副本為空")
	}
	if turn[0].GrowthBonusSum != base.GrowthBonusSum+3 ||
		turn[0].AIDifficultyFoodQuarters != 2 ||
		turn[0].AIDifficultyIndustryQuarters != 4 ||
		turn[0].AIDifficultyResearchQuarters != 4 {
		t.Fatalf("Hard AI 暫態加值錯誤：%+v", turn[0])
	}
	stored := s.AIPlayers[0].Colonies[0]
	if stored.GrowthBonusSum != base.GrowthBonusSum || stored.AIDifficultyFoodQuarters != 0 ||
		stored.AIDifficultyIndustryQuarters != 0 || stored.AIDifficultyResearchQuarters != 0 {
		t.Fatalf("難度加值污染持久殖民地：before=%+v after=%+v", base, stored)
	}
}
