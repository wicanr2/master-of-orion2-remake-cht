package gamedata

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestOriginalChangeRelationScore(t *testing.T) {
	tests := []struct {
		name string
		in   OriginalRelationChangeInput
		want int
	}{
		{"negative doubles and stops at ten", OriginalRelationChangeInput{CurrentRaw: -20, BaseDelta: 30, ActorGovernment: 2, Policy: DIPLO_NONE}, 10},
		{"positive relation damps gain", OriginalRelationChangeInput{CurrentRaw: 50, BaseDelta: 9, ActorGovernment: 2, Policy: DIPLO_ALLIANCE}, 53},
		{"negative delta doubles across positive", OriginalRelationChangeInput{CurrentRaw: 50, BaseDelta: -10, ActorGovernment: 2, Policy: DIPLO_NONE}, 30},
		{"feudal negative three halves", OriginalRelationChangeInput{CurrentRaw: 0, BaseDelta: -20, ActorGovernment: 0, Policy: DIPLO_NONE}, -30},
		{"democracy negative doubles", OriginalRelationChangeInput{CurrentRaw: 0, BaseDelta: -20, ActorGovernment: 4, Policy: DIPLO_NONE}, -40},
		{"charismatic doubles positive", OriginalRelationChangeInput{CurrentRaw: 20, BaseDelta: 5, ActorGovernment: 2, TargetCharismatic: true, Policy: DIPLO_ALLIANCE}, 30},
		{"non alliance positive cap", OriginalRelationChangeInput{CurrentRaw: 64, BaseDelta: 20, ActorGovernment: 2, Policy: DIPLO_NON_AGGRESSION}, 65},
		{"alliance may exceed cap", OriginalRelationChangeInput{CurrentRaw: 64, BaseDelta: 20, ActorGovernment: 2, Policy: DIPLO_ALLIANCE}, 70},
		{"late AI positive doubles", OriginalRelationChangeInput{CurrentRaw: 0, BaseDelta: 5, ActorGovernment: 2, Policy: DIPLO_NONE, BothAI: true, RelativeTurn: 101, Difficulty: 4}, 10},
		{"late AI negative difficulty divisor", OriginalRelationChangeInput{CurrentRaw: 0, BaseDelta: -12, ActorGovernment: 2, Policy: DIPLO_NONE, BothAI: true, RelativeTurn: 101, Difficulty: 4}, -4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := OriginalChangeRelationScore(tt.in)
			if !ok || got != tt.want {
				t.Fatalf("got=%d ok=%v want=%d", got, ok, tt.want)
			}
		})
	}
}

func TestOriginalWarBlockadeGrievance(t *testing.T) {
	got, ok := OriginalWarBlockadeGrievance(OriginalRelationChangeInput{
		CurrentRaw: -25, BaseDelta: -5, ActorGovernment: 2, Policy: DIPLO_WAR,
	})
	if !ok || got != -1 {
		t.Fatalf("-5 經 -25 關係縮減後再向下除四應為 -1：got=%d ok=%v", got, ok)
	}
	got, ok = OriginalWarBlockadeGrievance(OriginalRelationChangeInput{
		CurrentRaw: 0, BaseDelta: -5, ActorGovernment: 4, TargetCharismatic: true, Policy: DIPLO_LIMITED_WAR,
	})
	if !ok || got != -2 {
		t.Fatalf("民主負向倍增、Charismatic 減半後 -5/4 向下應為 -2：got=%d ok=%v", got, ok)
	}
}

func TestOriginalDiplomacyGrowthTreatyRelationOrder(t *testing.T) {
	rolls := []struct{ n, value int }{
		{100, 1}, {3, 1}, // NAP: -20 -> -18
		{100, 1}, {3, 2}, // trade: -18 -> -14
		{100, 1}, {3, 3}, // research: -14 -> -8
		{100, 1}, {3, 2}, // tribute: -8 -> -4
	}
	pos := 0
	got, ok := OriginalDiplomacyGrowthTreatyRelation(OriginalDiplomacyGrowthTreatyInput{
		CurrentRaw: -20, FormalPolicy: DIPLO_NON_AGGRESSION, TradeActive: true,
		ResearchActive: true, TributeMode: 1, ActorGovernment: 2,
	}, func(n int) int {
		if pos >= len(rolls) || rolls[pos].n != n {
			t.Fatalf("roll %d requested n=%d", pos, n)
		}
		value := rolls[pos].value
		pos++
		return value
	})
	if !ok || got != -4 || pos != len(rolls) {
		t.Fatalf("got=%d ok=%v rolls=%d", got, ok, pos)
	}
}

func TestOriginalDiplomacyGrowthAllianceIsUnconditional(t *testing.T) {
	calls := 0
	got, ok := OriginalDiplomacyGrowthTreatyRelation(OriginalDiplomacyGrowthTreatyInput{
		CurrentRaw: 90, FormalPolicy: DIPLO_ALLIANCE, ActorGovernment: 2,
	}, func(n int) int {
		calls++
		if n != 5 {
			t.Fatalf("alliance requested Random(%d), want 5", n)
		}
		return 5
	})
	if !ok || got != 91 || calls != 1 {
		t.Fatalf("got=%d ok=%v calls=%d", got, ok, calls)
	}
}

func TestOriginalDiplomacyGrowthFailedGateConsumesNoDeltaRoll(t *testing.T) {
	calls := 0
	got, ok := OriginalDiplomacyGrowthTreatyRelation(OriginalDiplomacyGrowthTreatyInput{
		CurrentRaw: 100, FormalPolicy: DIPLO_NON_AGGRESSION, ActorGovernment: 2,
	}, func(n int) int {
		calls++
		if n != 100 {
			t.Fatalf("failed gate must not request delta roll: n=%d", n)
		}
		return 1
	})
	if !ok || got != 100 || calls != 1 {
		t.Fatalf("got=%d ok=%v calls=%d", got, ok, calls)
	}
}

func TestOriginalDiplomacyGrowthRejectsInvalidRoll(t *testing.T) {
	got, ok := OriginalDiplomacyGrowthTreatyRelation(OriginalDiplomacyGrowthTreatyInput{
		CurrentRaw: 0, FormalPolicy: DIPLO_ALLIANCE,
	}, func(int) int { return 0 })
	if ok || got != 0 {
		t.Fatalf("invalid roller must fail closed: got=%d ok=%v", got, ok)
	}
}

func TestOriginalBaseRelationTableDirections(t *testing.T) {
	if got, ok := OriginalBaseRelation(0, 2); !ok || got != -24 {
		t.Fatalf("row 0 col 2 = %d, ok=%v; want -24", got, ok)
	}
	if got, ok := OriginalBaseRelation(2, 1); !ok || got != -12 {
		t.Fatalf("row 2 col 1 = %d, ok=%v; want -12", got, ok)
	}
	if _, ok := OriginalBaseRelation(-1, 0); ok {
		t.Fatal("custom race must not index the original table")
	}
}

func TestOriginalBaseRelationTableMatchesIDABytes(t *testing.T) {
	raw := make([]byte, 0, 14*14*2)
	for _, row := range originalBaseRelationLowBytes {
		for _, value := range row {
			raw = append(raw, byte(value))
			if value < 0 {
				raw = append(raw, 0xff)
			} else {
				raw = append(raw, 0)
			}
		}
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != "05e582491a6173319e8d57d0751d24dd5629f73f00dc3d7c2a5b979e39efa831" {
		t.Fatalf("transcribed relation table hash = %s", got)
	}
}

func TestOriginalDiplomacyRelationDrift(t *testing.T) {
	rolls := []struct{ n, value int }{{105, 105}, {4, 1}, {2, 2}}
	pos := 0
	got, ok := OriginalDiplomacyRelationDrift(OriginalDiplomacyRelationDriftInput{
		CurrentRaw: 5, TargetRaw: -24, Policy: DIPLO_NONE,
	}, func(n int) int {
		if pos >= len(rolls) || rolls[pos].n != n {
			t.Fatalf("roll %d requested n=%d", pos, n)
		}
		v := rolls[pos].value
		pos++
		return v
	})
	if !ok || got != 4 || pos != 3 {
		t.Fatalf("got=%d ok=%v rolls=%d", got, ok, pos)
	}
}

func TestOriginalDiplomacyRelationDriftWarCapAndLock(t *testing.T) {
	got, ok := OriginalDiplomacyRelationDrift(OriginalDiplomacyRelationDriftInput{
		CurrentRaw: 40, TargetRaw: 24, Policy: DIPLO_WAR, Locked: true,
	}, func(n int) int {
		if n != 105 {
			t.Fatalf("unexpected Random(%d)", n)
		}
		return 1
	})
	if !ok || got != -90 {
		t.Fatalf("war cap got=%d ok=%v", got, ok)
	}
}
