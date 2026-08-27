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
