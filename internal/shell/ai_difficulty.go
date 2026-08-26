package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
)

// applyAIDifficultyPlayerInputs 將官方五級 AI 帝國層難度值寫入本回合 PlayerState 副本。
// 無效難度保持零值，讓 engine 安全回退玩家預設；欄位均標 json:"-"，不形成第二份持久真相。
func applyAIDifficultyPlayerInputs(ps engine.PlayerState, difficulty int) engine.PlayerState {
	bonus, ok := ai.AIDifficultyBonus(ai.Difficulty(difficulty))
	if !ok {
		return ps
	}
	ps.AIDifficultyIncomeQuartersPerPop = bonus.BCQuarters
	ps.CommandOverflowCostPerPoint = bonus.CommandDeficitBC
	return ps
}
