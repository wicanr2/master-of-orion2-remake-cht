package engine

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/save"

// game.go 是最頂層的回合編排:吃一個完整 save.GameState,逐玩家跑帝國回合(RunEmpireTurn),
// 回傳各玩家的結算結果。這是 save↔engine 的頂層橋接。
//
// v1 只讀不寫:不回寫 GameState(人口成長累積尺度、BC/研究回寫的存檔相容性尚未驗證,見各處註解),
// 只計算並回傳每位玩家本回合的經濟/研究/國庫結果,供上層決定如何呈現或套用。

// GameTurnResult 是一個完整回合的結果:每位玩家(以其索引為 key)的帝國結算。
type GameTurnResult struct {
	PlayerOutputs map[int]EmpireOutput
}

// RunGameTurn 對整個 GameState 跑一回合:
//  1. 依 Colony.Owner 把所有(非前哨站)殖民地分組到各玩家。
//  2. 逐玩家用其殖民地群 + 玩家狀態跑 RunEmpireTurn。
//
// 前哨站(IsOutpost != 0)不計入殖民地經濟;Planet 索引非法的殖民地略過。
func RunGameTurn(gs *save.GameState) GameTurnResult {
	coloniesByOwner := make(map[int][]ColonyState)
	for i := 0; i < gs.ColonyCount && i < len(gs.Colonies); i++ {
		c := &gs.Colonies[i]
		if c.IsOutpost != 0 {
			continue
		}
		pi := int(c.Planet)
		if pi < 0 || pi >= len(gs.Planets) {
			continue
		}
		owner := int(c.Owner)
		coloniesByOwner[owner] = append(coloniesByOwner[owner], ColonyStateFromSave(c, &gs.Planets[pi]))
	}

	res := GameTurnResult{PlayerOutputs: make(map[int]EmpireOutput)}
	for p := 0; p < gs.PlayerCount && p < len(gs.Players); p++ {
		ps := PlayerStateFromSave(&gs.Players[p])
		res.PlayerOutputs[p] = RunEmpireTurn(ps, coloniesByOwner[p])
	}
	return res
}
