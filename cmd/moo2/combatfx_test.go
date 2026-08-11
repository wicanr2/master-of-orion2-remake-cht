package main

import "testing"

func TestCombatFXFrameAtPlaysAndExpires(t *testing.T) {
	if frame, active := combatFXFrameAt(0, 3, 3, 2); active || frame != 0 {
		t.Fatalf("特效尚未開始時應 inactive，得到 frame=%d active=%v", frame, active)
	}
	cases := []struct {
		tick  int
		frame int
	}{
		{3, 0}, {4, 0}, {5, 1}, {7, 2},
	}
	for _, tc := range cases {
		frame, active := combatFXFrameAt(tc.tick, 3, 3, 2)
		if !active || frame != tc.frame {
			t.Errorf("tick %d 應為 frame %d/active，得到 %d/%v", tc.tick, tc.frame, frame, active)
		}
	}
	if frame, active := combatFXFrameAt(9, 3, 3, 2); active || frame != 0 {
		t.Errorf("特效播完應 inactive，得到 frame=%d active=%v", frame, active)
	}
}
