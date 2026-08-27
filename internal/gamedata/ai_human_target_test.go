package gamedata

import "testing"

func TestOriginalHumanTargetPersonalityScoreTable(t *testing.T) {
	want := [...]int{-10, -5, -3, 0, 20, 20, -10}
	if OriginalHumanTargetPersonalityScore != want {
		t.Fatalf("word_181080=%v，預期 %v", OriginalHumanTargetPersonalityScore, want)
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
		Score: -10, ContactTurns: 20, PowerRatio: 100, Difficulty: 1, HasMilitaryTarget: true,
	}, roll)
	if !ok || got != 2 || len(rolls) != 0 {
		t.Fatalf("outcome=%d/%v rollsLeft=%v，預期 type 2 且三次 RNG", got, ok, rolls)
	}
}

func TestOriginalHumanTargetOutcomeContactGateStillConsumesEarlyRolls(t *testing.T) {
	spans := []int{}
	got, ok := OriginalHumanTargetOutcome(OriginalHumanTargetOutcomeInput{
		Score: -10, ContactTurns: 9, PowerRatio: 100, Difficulty: 2, HasMilitaryTarget: true,
	}, func(n int) int { spans = append(spans, n); return 1 })
	if !ok || got != 0 || len(spans) != 2 || spans[0] != 3 || spans[1] != 100 {
		t.Fatalf("contact gate outcome=%d/%v spans=%v", got, ok, spans)
	}
}
