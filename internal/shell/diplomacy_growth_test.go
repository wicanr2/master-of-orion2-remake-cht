package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestOriginalDiplomacyGrowthPlayerTreatyUsesRawRemainder(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Relation = -8
	a.OriginalRelationRaw = -20
	a.OriginalRelationKnown = true
	a.Treaty.FormalPolicy = gamedata.DIPLO_NON_AGGRESSION
	rolls := []struct{ n, value int }{{100, 1}, {3, 1}}
	pos := 0
	s.advanceOriginalDiplomacyGrowthForAIWithRoller(0, func(n int) int {
		if pos >= len(rolls) || rolls[pos].n != n {
			t.Fatalf("roll %d requested n=%d", pos, n)
		}
		value := rolls[pos].value
		pos++
		return value
	})
	if pos != 2 || a.OriginalRelationRaw <= -20 {
		t.Fatalf("raw treaty growth missing: raw=%d rolls=%d", a.OriginalRelationRaw, pos)
	}
	if a.Relation != normalizedRelationFromOriginal(a.OriginalRelationRaw) {
		t.Fatalf("normalized projection mismatch: relation=%d raw=%d", a.Relation, a.OriginalRelationRaw)
	}
}

func TestOriginalDiplomacyGrowthNoTreatyDoesNotUseInventedStrengthDrift(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Relation = 7
	a.OriginalRelationRaw = 17
	a.OriginalRelationKnown = true
	a.Treaty = TreatyState{}
	calls := 0
	s.advanceOriginalDiplomacyGrowthForAIWithRoller(0, func(int) int {
		calls++
		return 1
	})
	if calls != 0 || a.OriginalRelationRaw != 18 || a.Relation != 7 {
		t.Fatalf("no-treaty slice must preserve relation: calls=%d raw=%d relation=%d",
			calls, a.OriginalRelationRaw, a.Relation)
	}
}

func TestOriginalDiplomacyGrowthRebasesAfterExternalRelationChange(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.OriginalRelationRaw = 21
	a.OriginalRelationKnown = true
	a.Relation = -12 // 模擬外交行動直接改寫 normalized 值。
	if got := a.originalRelationRaw(); got != originalRelationFromNormalized(-12) {
		t.Fatalf("external relation change was overwritten by stale raw remainder: %d", got)
	}
}

func TestDiplomacyGrowthRawAndRandPersist(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers[0].OriginalRelationRaw = -17
	s.AIPlayers[0].OriginalRelationKnown = true
	s.AIPlayers[0].OriginalRelationTargetRaw = -24
	s.AIPlayers[0].OriginalRelationTargetKnown = true
	s.ensureOriginalAIAIRelations()
	s.setOriginalAIAIRelation(0, 1, -13)
	s.ensureAIAIState()
	s.AIReputationRaw[0][1] = -7
	s.AITreatyBiasRaw[0][1] = -20
	s.AIAgreementBiasRaw[0][1] = -30
	s.AITributeModes[0][1] = 2
	s.diplomacyGrowthRandForTurn().Intn(100)
	s.diplomacyGrowthRandForTurn().Intn(3)
	snap := s.snapshot()
	got := snap.restore()
	if got.AIPlayers[0].OriginalRelationRaw != -17 || !got.AIPlayers[0].OriginalRelationKnown ||
		got.AIPlayers[0].OriginalRelationTargetRaw != -24 || !got.AIPlayers[0].OriginalRelationTargetKnown {
		t.Fatalf("raw relation did not round-trip: %+v", got.AIPlayers[0])
	}
	if got.diplomacyGrowthRand == nil || got.diplomacyGrowthRand.Draws() != 2 {
		t.Fatalf("diplomacy growth RNG draws did not round-trip: %+v", got.diplomacyGrowthRand)
	}
	if got.AIRelationsRaw[0][1] != -13 || !got.AIRelationsRawKnown[1][0] {
		t.Fatalf("AI↔AI raw relation did not round-trip: raw=%v known=%v",
			got.AIRelationsRaw, got.AIRelationsRawKnown)
	}
	if got.AIReputationRaw[0][1] != -7 || got.AITreatyBiasRaw[0][1] != -20 ||
		got.AIAgreementBiasRaw[0][1] != -30 || got.AITributeModes[0][1] != 2 {
		t.Fatalf("NPC negotiation raw state did not round-trip: rep=%v treaty=%v agreement=%v tribute=%v",
			got.AIReputationRaw, got.AITreatyBiasRaw, got.AIAgreementBiasRaw, got.AITributeModes)
	}
}

func TestOriginalAIAIRelationsInitializeFromHighToLowRaceTarget(t *testing.T) {
	s := NewDemoSession()
	s.AIRelations, s.AIRelationsRaw, s.AIRelationsRawKnown = nil, nil, nil
	s.ensureOriginalAIAIRelations()
	lo, hi := aiRaceIndex(s.AIPlayers[0]), aiRaceIndex(s.AIPlayers[1])
	want, ok := gamedata.OriginalBaseRelation(Races[hi].OrigIdx, Races[lo].OrigIdx)
	if !ok || s.AIRelationsRaw[0][1] != want || s.AIRelationsRaw[1][0] != want {
		t.Fatalf("high→low target missing: raw=%v want=%d", s.AIRelationsRaw, want)
	}
}

func TestOriginalAIAIRelationWarCapUsesFormalPolicy(t *testing.T) {
	s := NewDemoSession()
	s.ensureAIAIState()
	s.ensureOriginalAIAIRelations()
	s.setOriginalAIAIRelation(0, 1, 20)
	s.AIPolicies[0][1], s.AIPolicies[1][0] = gamedata.DIPLO_WAR, gamedata.DIPLO_WAR
	s.advanceOriginalDiplomacyGrowth()
	if got := s.AIRelationsRaw[0][1]; got != -90 || s.AIRelationsRaw[1][0] != -90 {
		t.Fatalf("formal war must mirror -90 raw, got %v", s.AIRelationsRaw)
	}
	if s.AIRelations[0][1] != normalizedRelationFromOriginal(-90) {
		t.Fatalf("war projection mismatch: %v", s.AIRelations)
	}
}
