package shell

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
)

func TestBuildOriginalAITurnDataUsesOneTransientProjection(t *testing.T) {
	s := NewDemoSession()
	s.Difficulty = int(ai.DifficultyHard)
	a := &s.AIPlayers[0]
	base := a.Colonies[0]
	data, ok := s.buildOriginalAITurnData(0)
	if !ok || len(data.TurnColonies) != len(a.Colonies) || len(data.Jobs.ColonyFoodHalf) != len(a.Colonies) {
		t.Fatalf("AI 回合 cache 形狀錯誤：ok=%v data=%+v", ok, data)
	}
	turn := data.TurnColonies[0]
	if turn.GrowthBonusSum != base.GrowthBonusSum+3 || turn.AIDifficultyFoodQuarters != 2 ||
		turn.AIDifficultyIndustryQuarters != 4 || turn.AIDifficultyResearchQuarters != 4 {
		t.Fatalf("cache 未包含本回合難度投影：%+v", turn)
	}
	if stored := a.Colonies[0]; stored.GrowthBonusSum != base.GrowthBonusSum ||
		stored.AIDifficultyFoodQuarters != base.AIDifficultyFoodQuarters {
		t.Fatalf("typed cache 污染持久殖民地：before=%+v after=%+v", base, stored)
	}
}

func TestBuildOriginalAITurnDataCapturesBlockadeMask(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	if len(a.ColonyStars) == 0 || !a.PopulationRaceSlotKnown {
		t.Fatal("demo AI 應有殖民星與已知人口 slot")
	}
	star, slot := a.ColonyStars[0], a.PopulationRaceSlot
	s.Stars[star].BlockadedMask |= 1 << slot
	data, ok := s.buildOriginalAITurnData(0)
	if !ok || len(data.Jobs.ColonyBlockaded) == 0 || !data.Jobs.ColonyBlockaded[0] {
		t.Fatalf("cache 未保存殖民地封鎖輸入：ok=%v blocked=%v", ok, data.Jobs.ColonyBlockaded)
	}
}

func TestBuildOriginalAITurnDataIsNotPersisted(t *testing.T) {
	s := NewDemoSession()
	before, err := json.Marshal(s.snapshot())
	if err != nil {
		t.Fatal(err)
	}
	data, ok := s.buildOriginalAITurnData(0)
	if !ok || len(data.TurnColonies) == 0 {
		t.Fatal("應建立單回合 cache")
	}
	after, err := json.Marshal(s.snapshot())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("建立 Compute_AI_Data_ typed cache 不得改變 session snapshot")
	}
}

func TestOriginalAITurnDataFallbackPreservesTaxRate(t *testing.T) {
	s := NewDemoSession()
	data, ok := s.buildOriginalAITurnData(0)
	if !ok {
		t.Fatal("應建立 AI 回合 cache")
	}
	data.Player.TaxRate = 37
	data.Jobs.ColonyFoodHalfKnown[0] = false
	player, colonies, pressure, exact := data.applyJobs(s.AIPlayers[0].Decider)
	if exact || pressure || len(colonies) != len(data.PersistentColonies) {
		t.Fatalf("缺原版輸入應走有界 fallback：exact=%v pressure=%v colonies=%d", exact, pressure, len(colonies))
	}
	if player.TaxRate != 37 {
		t.Fatalf("fallback 不得用 remake 國庫門檻改寫原版無 writer 的稅率：%d", player.TaxRate)
	}
}

func TestBuildOriginalAITurnDataRejectsInvalidIndex(t *testing.T) {
	s := NewDemoSession()
	for _, index := range []int{-1, len(s.AIPlayers)} {
		if data, ok := s.buildOriginalAITurnData(index); ok || len(data.TurnColonies) != 0 {
			t.Fatalf("無效 AI index %d 應失敗即關閉：%+v", index, data)
		}
	}
}
