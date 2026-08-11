package audio

import "testing"

func TestClampVolume(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   float64
		want float64
	}{
		{name: "below", in: -0.25, want: 0},
		{name: "inside", in: 0.375, want: 0.375},
		{name: "above", in: 1.25, want: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClampVolume(tc.in); got != tc.want {
				t.Fatalf("ClampVolume(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
