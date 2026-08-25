package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
)

func TestHistoryUsesOriginalFourMetrics(t *testing.T) {
	want := []string{"艦隊", "科技", "人口", "建築"}
	for i, name := range want {
		if got := HistoryMetricName(HistoryMetric(i)); got != name {
			t.Fatalf("metric %d = %q, want %q", i, got, name)
		}
	}
}

func TestHistoryDivisorRescalesExistingSamples(t *testing.T) {
	s := &GameSession{HistoryScales: [4]int{1, 1, 1, 1}, History: []HistoryTurn{{
		Turn: 1, Empires: []HistorySample{{Population: 200}},
	}}, PlayerColonies: []engine.ColonyState{{Population: 251}}}
	s.recordHistory()
	if s.HistoryScales[HistoryPopulation] != 2 {
		t.Fatalf("population divisor = %d, want 2", s.HistoryScales[HistoryPopulation])
	}
	if got := s.History[0].Empires[0].Population; got != 100 {
		t.Fatalf("old sample after rescale = %d, want 100", got)
	}
	if got := s.History[1].Empires[0].Population; got != 125 {
		t.Fatalf("new sample = %d, want 125", got)
	}
}

func TestHistoryKeepsOriginalRingLength(t *testing.T) {
	s := &GameSession{}
	for turn := 1; turn <= historyMaxTurns+1; turn++ {
		s.Turn = turn
		s.recordHistory()
	}
	if len(s.History) != historyMaxTurns || s.History[0].Turn != 2 {
		t.Fatalf("history len/first = %d/%d, want %d/2", len(s.History), s.History[0].Turn, historyMaxTurns)
	}
}
