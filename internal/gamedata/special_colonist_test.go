package gamedata

import "testing"

func TestSpecialColonistProductionOriginalConstants(t *testing.T) {
	tests := []struct {
		slot                         int
		food, industry, research     int
		gravityImmune, profileExists bool
	}{
		{AndroidColonistSlot, 6, 3, 3, true, true},
		{NativeColonistSlot, 4, 0, 0, true, true},
		{0, 0, 0, 0, false, false},
	}
	for _, tt := range tests {
		food, industry, research, immune, ok := SpecialColonistProduction(tt.slot)
		if food != tt.food || industry != tt.industry || research != tt.research ||
			immune != tt.gravityImmune || ok != tt.profileExists {
			t.Fatalf("slot %d profile = %d/%d/%d immune=%v ok=%v，want %d/%d/%d immune=%v ok=%v",
				tt.slot, food, industry, research, immune, ok,
				tt.food, tt.industry, tt.research, tt.gravityImmune, tt.profileExists)
		}
	}
}
