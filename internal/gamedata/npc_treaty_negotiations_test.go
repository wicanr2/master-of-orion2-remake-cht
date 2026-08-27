package gamedata

import "testing"

func TestOriginalNPCTreatyNegotiationCreatesAllianceAndResearch(t *testing.T) {
	rolls := []struct{ n, v int }{{250, 1}, {100, 100}, {100, 100}, {100, 1}}
	pos := 0
	got, ok := OriginalNPCTreatyNegotiation(OriginalNPCTreatyInput{
		Difficulty: 0, CurrentRaw: 40, Policy: DIPLO_NON_AGGRESSION,
		TradeActive: true, OuterGovernment: 5, InnerGovernment: 5,
	}, func(n int) int {
		if pos >= len(rolls) || rolls[pos].n != n {
			t.Fatalf("roll %d requested Random(%d)", pos, n)
		}
		v := rolls[pos].v
		pos++
		return v
	})
	if !ok || !got.Processed || got.Policy != DIPLO_ALLIANCE || !got.ResearchActive ||
		got.TreatyBiasRaw != -30 || got.AgreementBiasRaw != -30 || pos != len(rolls) {
		t.Fatalf("unexpected result: %+v ok=%v rolls=%d", got, ok, pos)
	}
}

func TestOriginalNPCTreatyNegotiationCreatesTradeBeforeResearchThreshold(t *testing.T) {
	rolls := []int{1, 1, 81, 1}
	pos := 0
	got, ok := OriginalNPCTreatyNegotiation(OriginalNPCTreatyInput{
		Difficulty: 0, Policy: DIPLO_NONE, OuterGovernment: 3, InnerGovernment: 3,
	}, func(n int) int {
		v := rolls[pos]
		pos++
		return v
	})
	if !ok || !got.TradeActive || got.ResearchActive {
		t.Fatalf("trade threshold mismatch: %+v ok=%v", got, ok)
	}
}

func TestOriginalNPCTreatyNegotiationTributeDemand(t *testing.T) {
	rolls := []struct{ n, v int }{{50, 1}, {100, 1}, {100, 1}, {100, 100}, {20, 1}, {3, 2}}
	pos := 0
	got, ok := OriginalNPCTreatyNegotiation(OriginalNPCTreatyInput{
		Difficulty: 5, CurrentRaw: 60, TreatyBiasRaw: 40,
		OuterGovernment: 5, InnerGovernment: 5, OuterStrength: 10, InnerStrength: 100,
	}, func(n int) int {
		if pos >= len(rolls) || rolls[pos].n != n {
			t.Fatalf("roll %d requested Random(%d)", pos, n)
		}
		v := rolls[pos].v
		pos++
		return v
	})
	if !ok || got.TributeMode != 2 || got.RelationDelta != 5 {
		t.Fatalf("tribute mismatch: %+v ok=%v", got, ok)
	}
}

func TestOriginalNPCTreatyNegotiationGateAndBiasRecovery(t *testing.T) {
	got, ok := OriginalNPCTreatyNegotiation(OriginalNPCTreatyInput{
		Difficulty: 2, TreatyBiasRaw: -20, AgreementBiasRaw: -50,
	}, func(n int) int {
		if n != 170 {
			t.Fatalf("difficulty gate Random(%d), want 170", n)
		}
		return 2
	})
	if !ok || got.Processed || got.TreatyBiasRaw != -20 || got.AgreementBiasRaw != -50 {
		t.Fatalf("failed gate changed state: %+v ok=%v", got, ok)
	}
	if OriginalNPCNegotiationBiasRecovery(-25) != -15 || OriginalNPCNegotiationBiasRecovery(-5) != 0 {
		t.Fatal("bias recovery mismatch")
	}
}
