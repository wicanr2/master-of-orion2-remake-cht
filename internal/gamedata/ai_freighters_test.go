package gamedata

import "testing"

func TestOriginalAIFreighterFleetGain(t *testing.T) {
	if got, ok := OriginalAIFreighterFleetGain(true, 2, 2); !ok || got != 5 {
		t.Fatalf("pressure/difficulty/roll=1/2/2 got=%d ok=%v", got, ok)
	}
	if got, ok := OriginalAIFreighterFleetGain(true, 2, 3); !ok || got != 0 {
		t.Fatalf("嚴格門檻外 got=%d ok=%v", got, ok)
	}
	if got, ok := OriginalAIFreighterFleetGain(false, 4, 1); !ok || got != 0 {
		t.Fatalf("無壓力不得增建 got=%d ok=%v", got, ok)
	}
}
