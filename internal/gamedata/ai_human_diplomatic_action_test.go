package gamedata

import "testing"

func TestOriginalHumanDiplomaticActionCredits(t *testing.T) {
	out, ok := OriginalHumanDiplomaticActionSelect(OriginalHumanDiplomaticActionInput{
		Intensity: 4, CreditsEnabled: true, SourceMaintenance: 12, TargetCredits: 987, CreditIntensityLimit: 10,
	}, func(int) int { return 1 })
	if !ok || out.Kind != OriginalHumanDiplomaticActionCredits || out.Credits != 400 {
		t.Fatalf("credits action=%+v/%v，預期 400 BC", out, ok)
	}
}

func TestOriginalHumanDiplomaticActionSelectionRollOrder(t *testing.T) {
	spans := []int{}
	rolls := []int{2, 1, 2}
	roll := func(n int) int {
		spans = append(spans, n)
		v := rolls[0]
		rolls = rolls[1:]
		return v
	}
	out, ok := OriginalHumanDiplomaticActionSelect(OriginalHumanDiplomaticActionInput{
		Intensity: 6, DirectEnabled: true, TechnologyEnabled: true, CreditsEnabled: true, ColonyEnabled: true,
		TechnologyCandidates: []int{10, 20, 30}, TechnologyRatioLimit: 10,
		SourceMaintenance: 20, TargetCredits: 1000, CreditIntensityLimit: 10, ColonyCandidates: []int{7},
	}, roll)
	if !ok || out.Kind != OriginalHumanDiplomaticActionTechnology || out.Technology != 20 {
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

func TestOriginalHumanDiplomaticActionColonyUsesLowAndHighHalves(t *testing.T) {
	candidates := []int{10, 20, 30, 40, 50}
	low, ok := OriginalHumanDiplomaticActionSelect(OriginalHumanDiplomaticActionInput{
		Intensity: 6, ColonyEnabled: true, ColonyCandidates: candidates,
	}, func(int) int { return 2 })
	if !ok || low.Kind != OriginalHumanDiplomaticActionColony || low.Colony != 20 {
		t.Fatalf("low-half colony=%+v/%v", low, ok)
	}
	high, ok := OriginalHumanDiplomaticActionSelect(OriginalHumanDiplomaticActionInput{
		Intensity: 7, ColonyEnabled: true, ColonyCandidates: candidates,
	}, func(int) int { return 2 })
	if !ok || high.Kind != OriginalHumanDiplomaticActionColony || high.Colony != 40 {
		t.Fatalf("high-half colony=%+v/%v", high, ok)
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

func TestOriginalHumanDiplomaticRequestOutcomeReasons(t *testing.T) {
	action := OriginalHumanDiplomaticAction{Kind: OriginalHumanDiplomaticActionCredits, Credits: 500}
	for outcome, reason := range map[int]int{1: 106, 3: 105, 4: 124} {
		got, ok := OriginalHumanDiplomaticRequestForOutcome(outcome, action)
		if !ok || got.Outcome != outcome || got.ReasonCode != reason || got.Action != action {
			t.Fatalf("outcome %d request=%+v/%v，預期 reason %d", outcome, got, ok, reason)
		}
	}
	if _, ok := OriginalHumanDiplomaticRequestForOutcome(2, action); ok {
		t.Fatal("軍事 outcome 2 不得建立外交會談請求")
	}
	if _, ok := OriginalHumanDiplomaticRequestForOutcome(1, OriginalHumanDiplomaticAction{}); ok {
		t.Fatal("沒有 sub_4F93B action payload 不得建立請求")
	}
}
