package main

import "testing"

func TestSliderVolumeAt(t *testing.T) {
	const x, w = 100, 155
	for _, tc := range []struct {
		name  string
		mouse int
		want  float64
	}{
		{name: "left edge mutes", mouse: 100, want: 0},
		{name: "right edge full", mouse: 254, want: 1},
		{name: "middle", mouse: 177, want: 77.0 / 154.0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := sliderVolumeAt(tc.mouse, x, w); got != tc.want {
				t.Fatalf("sliderVolumeAt(%d, %d, %d) = %v, want %v", tc.mouse, x, w, got, tc.want)
			}
		})
	}
}
