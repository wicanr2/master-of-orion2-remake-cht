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
	// 玩家研究:母星科學家3 * 每科研3(gamedata.ResearchPerScientistNorm,銀河基準)* 士氣(1+0%)
	// = 9。2026-07-12 校正:①母星分配 科1→科3(SAVE10 oracle 不變式科≥2)②每科研基準 3(手冊
	// p.949「usual 3」,先前硬編 30 約 10x 過高)③開局領袖池改為空(手冊 p.47/134 原版開局無領袖、
	// 須雇用,先前 demoLeaders 自帶「馮·諾伊曼」+25 是機制錯誤)。故總研究 = 9 純基礎,不再 +25。
	if s.LastPlayerOutput.TotalResearch != 9 {
		t.Errorf("玩家總研究 = %d,預期 9(科3×norm3,開局無領袖)", s.LastPlayerOutput.TotalResearch)
	}
	if s.Player.ResearchProgress != 9 {
		t.Errorf("玩家研究進度 = %d,預期 9", s.Player.ResearchProgress)
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
