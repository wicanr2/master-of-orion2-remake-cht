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
	// = 9 基礎。2026-07-12 依原版 SAVE10.GAM 校正母星分配 科1→科3(oracle 不變式科≥2,先前科1
	// 違反),每科研基準 3(手冊 p.949「usual 3」,先前硬編 30 約 10x 過高已訂正)。
	// 領袖:demoLeaders「馮·諾伊曼(科學家)」套 SKILL_RESEARCHER 固定加成 +25
	// (LeaderSkillBonus(SKILL_RESEARCHER, tier1, exp4)=5*(4+1)),故 9 + 25 = 34(先前 28=3+25)。
	// ⚠ 已知後續議題(見 original-gameplay-reference.md §7.0.1):領袖 +25 現仍大於基礎 9、
	// 且原版開局本無領袖,待後續獨立輪次處理。
	if s.LastPlayerOutput.TotalResearch != 34 {
		t.Errorf("玩家總研究 = %d,預期 34(9 基礎[科3×norm3] + 25 領袖技能)", s.LastPlayerOutput.TotalResearch)
	}
	if s.Player.ResearchProgress != 34 {
		t.Errorf("玩家研究進度 = %d,預期 34", s.Player.ResearchProgress)
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
