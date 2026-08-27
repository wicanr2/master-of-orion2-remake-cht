package gamedata

import "testing"

func TestOriginalHumanDiplomaticActionCredits(t *testing.T) {
	out, ok := OriginalHumanDiplomaticActionSelect(OriginalHumanDiplomaticActionInput{
		Intensity: 4, CreditsEnabled: true, SourceIncome: 12, TargetCredits: 987, CreditIntensityLimit: 10,
	}, func(int) int { return 1 })
	if !ok || out.Kind != OriginalHumanDiplomaticActionCredits || out.Credits != 400 {
		t.Fatalf("credits action=%+v/%v，預期 400 BC", out, ok)
	}
}

func TestOriginalHumanDiplomaticActionSelectionRollOrder(t *testing.T) {
	spans := []int{}
	rolls := []int{1, 2, 2}
	roll := func(n int) int {
		spans = append(spans, n)
		v := rolls[0]
		rolls = rolls[1:]
		return v
	}
	out, ok := OriginalHumanDiplomaticActionSelect(OriginalHumanDiplomaticActionInput{
		Intensity: 6, DirectEnabled: true, TechnologyEnabled: true, CreditsEnabled: true, ColonyEnabled: true,
		TechnologyCandidates: []int{10, 20, 30}, TechnologyRatioLimit: 10,
		SourceIncome: 20, TargetCredits: 1000, CreditIntensityLimit: 10, ColonyTarget: 7,
	}, roll)
	if !ok || out.Kind != OriginalHumanDiplomaticActionTechnology || out.Technology != 30 {
		t.Fatalf("technology action=%+v/%v", out, ok)
	}
	want := []int{3, 3, 3}
	if len(spans) != len(want) {
		t.Fatalf("RNG spans=%v，預期 %v", spans, want)
	}
	for i := range want {
		if spans[i] != want[i] {
			t.Fatalf("RNG spans=%v，預期 %v", spans, want)
		}
	}
}

func TestOriginalHumanDiplomaticActionDirectTier(t *testing.T) {
	out, ok := OriginalHumanDiplomaticActionSelect(OriginalHumanDiplomaticActionInput{
		Intensity: 3, DirectEnabled: true,
	}, func(n int) int { return n })
	if !ok || out.Kind != OriginalHumanDiplomaticActionDirect || out.DirectTier != 2 {
		t.Fatalf("direct action=%+v/%v", out, ok)
	}
}
