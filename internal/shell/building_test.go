package shell

import "testing"

// TestBuildingLongTermEffect 驗證建築完工後套用長期產出效果:自動工廠提升該殖民地工業/工人,
// 且效果只套一次(重複完工不疊加)。
func TestBuildingLongTermEffect(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true // 隔離隨機事件(富礦脈會改工業,干擾精確斷言)
	if len(s.PlayerColonies) == 0 {
		t.Fatal("需至少一個殖民地")
	}
	startIPW := s.PlayerColonies[0].IndustryPerWorker

	// 在殖民地 0 排自動工廠,給足工業直接完工。
	s.Builds[0] = ColonyBuild{Name: "自動工廠", Progress: 0, Cost: 60}
	// 先跑一回合產生 LastPlayerOutput(advanceBuilds 讀其淨工業)。
	s.EndTurn()
	// 多跑幾回合確保累積完工。
	for i := 0; i < 20 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}

	if !s.ColonyBuildings[0]["自動工廠"] {
		t.Fatal("自動工廠應標記為已建")
	}
	afterIPW := s.PlayerColonies[0].IndustryPerWorker
	if afterIPW != startIPW+2 {
		t.Fatalf("自動工廠應使工業/工人 +2:%d → %d", startIPW, afterIPW)
	}

	// 再建一次自動工廠,不應再疊加效果。
	s.Builds[0] = ColonyBuild{Name: "自動工廠", Progress: 0, Cost: 60}
	for i := 0; i < 20 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}
	if s.PlayerColonies[0].IndustryPerWorker != afterIPW {
		t.Fatalf("重複建造不應再疊加效果:%d → %d", afterIPW, s.PlayerColonies[0].IndustryPerWorker)
	}
	t.Logf("自動工廠:工業/工人 %d→%d(+2,不重複疊加)", startIPW, afterIPW)
}

// TestResearchLabEffect 驗證研究實驗室提升研究/科學家 +5。
func TestResearchLabEffect(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	start := s.PlayerColonies[0].ResearchPerScientist
	s.Builds[0] = ColonyBuild{Name: "研究實驗室", Progress: 0, Cost: 60}
	for i := 0; i < 25 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}
	if got := s.PlayerColonies[0].ResearchPerScientist; got != start+5 {
		t.Fatalf("研究實驗室應使研究/科學家 +5:%d → %d", start, got)
	}
}
