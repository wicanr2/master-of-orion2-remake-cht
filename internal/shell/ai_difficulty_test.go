package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
)

func TestApplyAIDifficultyPlayerInputs(t *testing.T) {
	ps := applyAIDifficultyPlayerInputs(engine.PlayerState{}, int(ai.DifficultyHard))
	if ps.AIDifficultyIncomeQuartersPerPop != 2 || ps.CommandOverflowCostPerPoint != 9 {
		t.Fatalf("Hard AI 帝國難度輸入錯誤：%+v", ps)
	}
	invalid := applyAIDifficultyPlayerInputs(engine.PlayerState{}, -1)
	if invalid.AIDifficultyIncomeQuartersPerPop != 0 || invalid.CommandOverflowCostPerPoint != 0 {
		t.Fatalf("非法難度必須保持安全零值：%+v", invalid)
	}
}
