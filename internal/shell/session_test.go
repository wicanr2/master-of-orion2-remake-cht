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
	// 玩家研究:母星(單一殖民地,docs/tech/homeworld-init.md)科學家1*30*士氣(1+0%)=30。
	// (原本 53 含已移除的假殖民地2 貢獻的 20,母星初始狀態改為忠實單一母星後不再適用。)
	// 2026-07-11 士氣接線訂正:MoralePercent 先前硬編 +10(無手冊依據),現改用
	// colonyMoralePercent(獨裁 + homeworldBuildings 已建海軍陸戰隊營)算出忠實值 0——獨裁「無
	// Barracks -20%」懲罰因已建 Marine Barracks 被解除,且無士氣類建築帶來額外加成,淨額為 0,
	// 不再有 +10% 產出灌水。見 playerHomeworldColony 註解。
	// 2026-07-11 領袖技能接線:demoLeaders 的「馮·諾伊曼(科學家)」現套用 SKILL_RESEARCHER 固定
	// 研究加成到母星(applyLeaderColonyBonuses,見 leader_test.go),30 → 55:
	// bonus = gamedata.LeaderSkillBonus(SKILL_RESEARCHER, tier=1, expLevel=4) = 5*(4+1) = 25。
	if s.LastPlayerOutput.TotalResearch != 55 {
		t.Errorf("玩家總研究 = %d,預期 55(30 基礎 + 25 領袖技能)", s.LastPlayerOutput.TotalResearch)
	}
	if s.Player.ResearchProgress != 55 {
		t.Errorf("玩家研究進度 = %d,預期 55", s.Player.ResearchProgress)
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
