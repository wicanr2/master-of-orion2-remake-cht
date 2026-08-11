package gamedata

import "testing"

func TestFighterDamageRangeForKind(t *testing.T) {
	tests := []struct {
		kind    int
		minWant int
		maxWant int
	}{
		{0, 1, 4},
		{1, 4, 16},
		{2, 2, 7},
	}
	for _, tt := range tests {
		got, ok := FighterDamageRangeForKind(tt.kind)
		if !ok || got.Min != tt.minWant || got.Max != tt.maxWant {
			t.Errorf("kind %d: got (%d..%d), ok=%v; want (%d..%d), ok=true",
				tt.kind, got.Min, got.Max, ok, tt.minWant, tt.maxWant)
		}
	}
	if got, ok := FighterDamageRangeForKind(3); ok || got != (FighterDamageRange{}) {
		t.Fatalf("assault shuttle should not have a ship-fire range: %#v, ok=%v", got, ok)
	}
	if got, ok := FighterDamageRangeForKind(99); ok || got != (FighterDamageRange{}) {
		t.Fatalf("unknown fighter kind should fail closed: %#v, ok=%v", got, ok)
	}
}
