package shell

import "testing"

func TestGameSessionEndTurn(t *testing.T) {
	s := NewDemoSession()
	if s.Turn != 1 {
		t.Fatalf("初始回合 = %d,預期 1", s.Turn)
	}
	aiBefore := s.AIPlayers[0].Player.ResearchProgress
	s.EndTurn()
	if s.Turn != 2 {
		t.Errorf("EndTurn 後回合 = %d,預期 2", s.Turn)
	}
	// 玩家研究:母星科學家2 * 每科研3(gamedata.ResearchPerScientistNorm,銀河基準)* 士氣(1+0%)
	// = 6。沿革:2026-07-12 由科1→科3(SAVE10 oracle 不變式科≥2)+ 每科研基準 3(手冊 p.949
	// 「usual 3」,先前硬編 30 約 10x 過高)+ 開局領袖池清空(原版須雇用);2026-08-06 再由科3→科2
	// ——archive.org 線上原版實測直接讀到 Sol III 母星是 農4/工2/科2(直接觀察原版,oracle 優先序
	// 高於 SAVE10 不變式重建,且仍滿足「工≤2、科≥2」),見 docs/tech/oracle-comparison-20260712.md。
	if s.LastPlayerOutput.TotalResearch != 6 {
		t.Errorf("玩家總研究 = %d,預期 6(科2×norm3,開局無領袖)", s.LastPlayerOutput.TotalResearch)
	}
	if s.Player.ResearchProgress != 6 {
		t.Errorf("玩家研究進度 = %d,預期 6", s.Player.ResearchProgress)
	}
	// AI 也推進了(研究進度增加)
	if s.AIPlayers[0].Player.ResearchProgress <= aiBefore {
		t.Errorf("AI 研究進度未推進:%d → %d", aiBefore, s.AIPlayers[0].Player.ResearchProgress)
	}
	// 連跑第二回合仍正常
	s.EndTurn()
	if s.Turn != 3 {
		t.Errorf("第二次 EndTurn 後回合 = %d,預期 3", s.Turn)
	}
}
