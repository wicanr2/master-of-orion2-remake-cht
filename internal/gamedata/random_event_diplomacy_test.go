package gamedata

import "testing"

func TestOriginalDiplomaticIncidentRelation(t *testing.T) {
	tests := []struct {
		name                      string
		current, event, gov, want int
		charismatic               bool
		policy                    ForeignPolicy
	}{
		{"和平不適用", -20, 4, 2, -20, false, DIPLO_PEACE},
		{"聯姻仍受戰爭上限", -100, 5, 2, -25, false, DIPLO_WAR},
		{"暗殺由中度敵對降至極端", -25, 4, 2, -75, false, DIPLO_WAR},
		{"封建負面加重", -25, 4, 0, -100, false, DIPLO_WAR},
		{"魅力目標減半負面", -25, 4, 2, -50, true, DIPLO_LIMITED_WAR},
		{"民主負面加倍", -25, 4, 4, -100, false, DIPLO_WAR},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := OriginalDiplomaticIncidentRelation(tt.current, tt.event, tt.gov,
				tt.charismatic, tt.policy)
			if tt.policy < DIPLO_LIMITED_WAR {
				if ok || got != tt.want {
					t.Fatalf("非戰爭配對應不適用：got=%d ok=%v", got, ok)
				}
				return
			}
			if !ok || got != tt.want {
				t.Fatalf("got=%d ok=%v want=%d", got, ok, tt.want)
			}
		})
	}
}
