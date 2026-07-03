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
	// 玩家研究:殖民地1 科學家1*30*士氣1.1=33 + 殖民地2 1*20=20 = 53
	if s.LastPlayerOutput.TotalResearch != 53 {
		t.Errorf("玩家總研究 = %d,預期 53", s.LastPlayerOutput.TotalResearch)
	}
	if s.Player.ResearchProgress != 53 {
		t.Errorf("玩家研究進度 = %d,預期 53", s.Player.ResearchProgress)
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
