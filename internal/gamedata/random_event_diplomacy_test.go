package gamedata

import "testing"

func TestOriginalDiplomaticIncidentRelation(t *testing.T) {
	tests := []struct {
		name                 string
		current, event, want int
		policy               ForeignPolicy
	}{
		{"和平不適用", -20, 4, -20, DIPLO_PEACE},
		{"戰爭中聯姻不寫回關係", -100, 5, -100, DIPLO_WAR},
		{"戰爭中暗殺不寫回關係", -25, 4, -25, DIPLO_WAR},
		{"正關係也不會繞過戰爭早退", 20, 4, 20, DIPLO_WAR},
		{"有限戰爭同樣早退", -25, 5, -25, DIPLO_LIMITED_WAR},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := OriginalDiplomaticIncidentRelation(tt.current, tt.event, tt.policy)
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
