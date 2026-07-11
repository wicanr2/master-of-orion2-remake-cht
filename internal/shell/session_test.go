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
	// 玩家研究:母星科學家1 * 每科研3(gamedata.ResearchPerScientistNorm,銀河基準)* 士氣(1+0%)
	// = 3 基礎。2026-07-12 依原版 SAVE10.GAM + 手冊 p.949「usual 3」校正:先前 ResearchPerScientist
	// 硬編 30(約 10x 過高)訂正為 norm 3。
	// 領袖:demoLeaders「馮·諾伊曼(科學家)」套 SKILL_RESEARCHER 固定加成 +25
	// (LeaderSkillBonus(SKILL_RESEARCHER, tier1, exp4)=5*(4+1)),故 3 + 25 = 28(先前 55=30+25)。
	// ⚠ 已知後續議題(見 original-gameplay-reference.md §7.0.1,拆成獨立輪次避免漣漪):
	// ①母星科學家分配對齊原版科2-4(需連同食物/農夫校正,否則工業/稅收漣漪)②領袖 +25 現遠
	// 大於基準 3、且原版開局本無領袖 ③食物每農夫 TERRAN 2 vs 原版存檔 4。
	if s.LastPlayerOutput.TotalResearch != 28 {
		t.Errorf("玩家總研究 = %d,預期 28(3 基礎[科1×norm3] + 25 領袖技能)", s.LastPlayerOutput.TotalResearch)
	}
	if s.Player.ResearchProgress != 28 {
		t.Errorf("玩家研究進度 = %d,預期 28", s.Player.ResearchProgress)
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
