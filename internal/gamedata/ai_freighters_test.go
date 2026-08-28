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

func TestOriginalAIFreighterFleetBuildQuota(t *testing.T) {
	tests := []struct {
		surplus, moving, want int
	}{
		{0, 0, 0}, {1, 1, 0}, {0, 1, 1}, {-1, 0, 1}, {-5, 0, 1}, {-6, 0, 2}, {2, 8, 2},
	}
	for _, tt := range tests {
		got, ok := OriginalAIFreighterFleetBuildQuota(tt.surplus, tt.moving)
		if !ok || got != tt.want {
			t.Errorf("quota(%d,%d)=%d,%v want %d,true", tt.surplus, tt.moving, got, ok, tt.want)
		}
	}
	if _, ok := OriginalAIFreighterFleetBuildQuota(0, -1); ok {
		t.Error("負航行殖民船數應失敗即關閉")
	}
}
