package gamedata

import "testing"

func TestOriginalHumanTargetPersonalityScoreTable(t *testing.T) {
	want := [...]int{-10, -5, -3, 0, 20, 20, -10}
	if OriginalHumanTargetPersonalityScore != want {
		t.Fatalf("word_181080=%v，預期 %v", OriginalHumanTargetPersonalityScore, want)
	}
}

func TestOriginalHumanTargetIncidentScore(t *testing.T) {
	for personality, want := range []int{-30, -15, -10, -10, -7, -6, -15} {
		got, reason, ok := OriginalHumanTargetIncidentScore(3, 7, personality)
		if !ok || got != want || reason != 77 {
			t.Fatalf("personality=%d: got score=%d reason=%d ok=%v，預期 %d/77/true",
				personality, got, reason, ok, want)
		}
	}
	if got, reason, ok := OriginalHumanTargetIncidentScore(0, 0, 4); !ok || got != 0 || reason != 0 {
		t.Fatalf("空記憶應為中性：score=%d reason=%d ok=%v", got, reason, ok)
	}
	if _, _, ok := OriginalHumanTargetIncidentScore(1, 10, 0); ok {
		t.Fatal("+0x6CF 不可接受原版 writer 範圍外的 reason")
	}
}

func TestOriginalHumanTargetPopulationDominance(t *testing.T) {
	for _, tc := range []struct {
		populations []int
		source      int
		want        int
	}{
		{[]int{40}, 0, -10},
		{[]int{40, 30}, 0, -10},
		{[]int{40, 40}, 0, 0},
		{[]int{40, 30, 20}, 0, 0},
	} {
		got, ok := OriginalHumanTargetPopulationDominance(tc.populations, tc.source)
		if !ok || got != tc.want {
			t.Fatalf("populations=%v source=%d: got %d/%v，預期 %d/true",
				tc.populations, tc.source, got, ok, tc.want)
		}
	}
}

func TestOriginalHumanTargetPopulationTrend(t *testing.T) {
	if got, ok := OriginalHumanTargetPopulationTrend(99, 100, 80, 120, 80); !ok || got != 0 {
		t.Fatalf("100 回合前不得套趨勢：%d/%v", got, ok)
	}
	if got, ok := OriginalHumanTargetPopulationTrend(100, 100, 80, 121, 80); !ok || got != -10 {
		t.Fatalf("AI +20、真人 +41 應得 (20-41)/2=-10：%d/%v", got, ok)
	}
	if got, ok := OriginalHumanTargetPopulationTrend(100, 120, 80, 100, 80); !ok || got != 0 {
		t.Fatalf("真人成長未領先應為 0：%d/%v", got, ok)
	}
}

func TestOriginalHumanTargetThreshold(t *testing.T) {
	if got, ok := OriginalHumanTargetThreshold(-5, 20); !ok || got != 100 {
		t.Fatalf("負分 threshold=%d/%v，預期 100/true", got, ok)
	}
	if got, ok := OriginalHumanTargetThreshold(3, 64); !ok || got != 12 { // isqrt(32)=5
		t.Fatalf("正分 threshold=%d/%v，預期 12/true", got, ok)
	}
}

func TestOriginalHumanTargetHonorableBetrayalUsesDishonoredScore(t *testing.T) {
	loyal, ok1 := OriginalHumanTargetBaseScore(0, 4, 0, 2, 0, 0, false, false)
	betrayed, ok2 := OriginalHumanTargetBaseScore(0, 4, 0, 2, 0, 0, false, true)
	if !ok1 || !ok2 || loyal-betrayed != 30 {
		t.Fatalf("Honorable/Dishonored score=%d/%d ok=%v/%v，預期差 30", loyal, betrayed, ok1, ok2)
	}
}

func TestOriginalHumanTargetOutcomeConsumesRollsInOriginalOrder(t *testing.T) {
	rolls := []int{2, 1, 16}
	roll := func(n int) int {
		if len(rolls) == 0 {
			t.Fatalf("多消耗 Random_(%d)", n)
		}
		v := rolls[0]
		rolls = rolls[1:]
		return v
	}
	got, ok := OriginalHumanTargetOutcome(OriginalHumanTargetOutcomeInput{
		Score: -10, ContactTurns: 20, PowerRatio: 100, Difficulty: 1, DiplomaticActionAvailable: true,
	}, roll)
	if !ok || got != 2 || len(rolls) != 0 {
		t.Fatalf("outcome=%d/%v rollsLeft=%v，預期 type 2 且三次 RNG", got, ok, rolls)
	}
}

func TestOriginalHumanTargetOutcomeContactGateStillConsumesEarlyRolls(t *testing.T) {
	spans := []int{}
	got, ok := OriginalHumanTargetOutcome(OriginalHumanTargetOutcomeInput{
		Score: -10, ContactTurns: 9, PowerRatio: 100, Difficulty: 2, DiplomaticActionAvailable: true,
	}, func(n int) int { spans = append(spans, n); return 1 })
	if !ok || got != 0 || len(spans) != 2 || spans[0] != 3 || spans[1] != 100 {
		t.Fatalf("contact gate outcome=%d/%v spans=%v", got, ok, spans)
	}
}
