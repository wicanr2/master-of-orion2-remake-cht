package gamedata

import "testing"

func TestOriginalNPCPowerRatioCapsAndHalves(t *testing.T) {
	if got, ok := OriginalNPCPowerRatio(1000, 1, 2); !ok || got != 200 {
		t.Fatalf("ratio=%d ok=%v，應先夾 800 再折半兩次", got, ok)
	}
}

func TestOriginalNPCGenericWarCandidateUsesPolicyAndGovernmentScore(t *testing.T) {
	in := OriginalNPCWarCandidateInput{
		Difficulty: 2, Government: 0, Policy: DIPLO_NONE,
		SourceStrength: 300, TargetStrength: 100, TargetIsRotating: true,
	}
	got, ok := OriginalNPCGenericWarCandidate(in, func(n int) int { return 1 })
	if !ok || !got {
		t.Fatalf("低門檻政府的強國應成為候選：got=%v ok=%v", got, ok)
	}
	in.Policy = DIPLO_PEACE
	got, ok = OriginalNPCGenericWarCandidate(in, func(n int) int { return 200 })
	if !ok || got {
		t.Fatalf("和平政策與高亂數應阻止候選：got=%v ok=%v", got, ok)
	}
}

func TestOriginalNPCGenericWarCandidateHonorsCooldownAndHardDifficultyVeto(t *testing.T) {
	base := OriginalNPCWarCandidateInput{
		Difficulty: 3, Government: 3, SourceStrength: 1000, TargetStrength: 1,
		TargetIsRotating: true,
	}
	base.Cooldown = 1
	if got, ok := OriginalNPCGenericWarCandidate(base, func(n int) int { t.Fatal("冷卻時不應取亂數"); return 1 }); !ok || got {
		t.Fatalf("冷卻 veto 失敗：got=%v ok=%v", got, ok)
	}
	base.Cooldown = 0
	base.TargetAtWarWithAI = true
	if got, ok := OriginalNPCGenericWarCandidate(base, func(n int) int { t.Fatal("高難度第三方戰爭時不應取亂數"); return 1 }); !ok || got {
		t.Fatalf("第三方戰爭 veto 失敗：got=%v ok=%v", got, ok)
	}
}

func TestOriginalNPCCeasefireThreshold(t *testing.T) {
	if got, ok := OriginalNPCCeasefireThreshold(4, 1); !ok || got != 10 {
		t.Fatalf("threshold=%d ok=%v", got, ok)
	}
}

func TestOriginalNPCGovernmentWarCandidate(t *testing.T) {
	in := OriginalNPCSpecialWarCandidateInput{Difficulty: 2, Government: 3, PowerRatio: 100}
	got, ok := OriginalNPCGovernmentWarCandidate(in, func(n int) int {
		if n != 150 {
			t.Fatalf("roll span=%d，want 150", n)
		}
		return 1
	})
	if !ok || !got {
		t.Fatalf("reason 20 應成立：got=%v ok=%v", got, ok)
	}
	in.Government = 2
	if got, ok = OriginalNPCGovernmentWarCandidate(in, func(int) int { t.Fatal("非 government 3 不應取亂數"); return 1 }); !ok || got {
		t.Fatalf("政府 gate 失敗：got=%v ok=%v", got, ok)
	}
}

func TestOriginalNPCHostilityWarCandidateUsesSignedRelationThreshold(t *testing.T) {
	in := OriginalNPCSpecialWarCandidateInput{
		Difficulty: 2, Government: 1, PowerRatio: 100, TargetIsRotating: true, CurrentRelationRaw: -80,
	}
	// (-(-80)-5)/(2*2+1) = 15。
	if got, ok := OriginalNPCHostilityWarCandidate(in, func(n int) int { return 15 }); !ok || !got {
		t.Fatalf("門檻內應成立：got=%v ok=%v", got, ok)
	}
	if got, ok := OriginalNPCHostilityWarCandidate(in, func(n int) int { return 16 }); !ok || got {
		t.Fatalf("門檻外不應成立：got=%v ok=%v", got, ok)
	}
}

func TestOriginalNPCFoodDeficitWarCandidateConsumesRollAtZero(t *testing.T) {
	calls := 0
	in := OriginalNPCSpecialWarCandidateInput{Difficulty: 2, Government: 1, PowerRatio: 100}
	if got, ok := OriginalNPCFoodDeficitWarCandidate(in, func(n int) int { calls++; return 1 }); !ok || got || calls != 1 {
		t.Fatalf("零赤字 streak 仍應擲一次且不成立：got=%v ok=%v calls=%d", got, ok, calls)
	}
	in.FoodDeficitTurns = 3
	if got, ok := OriginalNPCFoodDeficitWarCandidate(in, func(n int) int { return 3 }); !ok || !got {
		t.Fatalf("streak 內應成立：got=%v ok=%v", got, ok)
	}
}

func TestOriginalNPCFoodDeficitTurnsAccumulatesAndResets(t *testing.T) {
	if got, ok := OriginalNPCFoodDeficitTurns(4, -1); !ok || got != 5 {
		t.Fatalf("赤字 streak=%d ok=%v，want 5,true", got, ok)
	}
	if got, ok := OriginalNPCFoodDeficitTurns(5, 0); !ok || got != 0 {
		t.Fatalf("非赤字應歸零：got=%d ok=%v", got, ok)
	}
	if got, ok := OriginalNPCFoodDeficitTurns(32767, -1); !ok || got != 32767 {
		t.Fatalf("signed word 上限應飽和：got=%d ok=%v", got, ok)
	}
}
