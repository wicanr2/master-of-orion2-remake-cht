package gamedata

import "testing"

func TestColonizationStartsWithWorker(t *testing.T) {
	cases := []struct {
		name                  string
		naturalFood           int
		lithovore, cybernetic bool
		wantWorker            bool
	}{
		{"可自然耕作", 2, false, false, false},
		{"自然食物為零", 0, false, false, true},
		{"Lithovore", 3, true, false, true},
		{"Cybernetic", 3, false, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ColonizationStartsWithWorker(tc.naturalFood, tc.lithovore, tc.cybernetic); got != tc.wantWorker {
				t.Fatalf("got %v, want %v", got, tc.wantWorker)
			}
		})
	}
}
