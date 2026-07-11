package shell

import "testing"

// TestAIBuildsAndExpands 驗證 AI 對手主動造艦(FleetStrength 成長)並擴張星圖(佔無主星)。
func TestAIBuildsAndExpands(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	if len(s.AIPlayers) == 0 {
		t.Fatal("需至少一個 AI 對手")
	}
	startFleet := s.AIPlayers[0].FleetStrength
	unownedStart := 0
	for _, st := range s.Stars {
		if st.Owner == 0 {
			unownedStart++
		}
	}

	for i := 0; i < 30; i++ {
		s.EndTurn()
	}

	if s.AIPlayers[0].FleetStrength <= startFleet {
		t.Fatalf("AI 應主動造艦累積軍力:%d → %d", startFleet, s.AIPlayers[0].FleetStrength)
	}
	if s.AIPlayers[0].OwnedStars == 0 {
		t.Fatal("AI 應擴張佔領無主星")
	}
	unownedEnd := 0
	for _, st := range s.Stars {
		if st.Owner == 0 {
			unownedEnd++
		}
	}
	if unownedEnd >= unownedStart {
		t.Fatalf("AI 擴張後無主星應減少:%d → %d", unownedStart, unownedEnd)
	}
	if s.AIPlayers[0].StanceName == "" {
		t.Fatal("AI 應有外交態勢")
	}
	t.Logf("AI 軍力 %d→%d、佔領 %d 星、態勢「%s」", startFleet, s.AIPlayers[0].FleetStrength, s.AIPlayers[0].OwnedStars, s.AIPlayers[0].StanceName)
}

// TestAIExpand_CreatesRealColony 驗證 aiExpand 佔領無主星時會建立真正的 engine.ColonyState
// (而不只是先前的「標旗標、無殖民地模型」簡化),且 Colonies/ColonyStars 兩個平行陣列同步、
// 不越界(見 AIOpponent.ColonyStars 欄位註解、colonization.go newColonyFromStar)。
func TestAIExpand_CreatesRealColony(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	beforeColonies := len(s.AIPlayers[0].Colonies)
	beforeStars := len(s.AIPlayers[0].ColonyStars)

	s.aiExpand(0)

	if len(s.AIPlayers[0].Colonies) != beforeColonies+1 {
		t.Fatalf("aiExpand 後 AI Colonies 應 +1(%d→%d),got %d", beforeColonies, beforeColonies+1, len(s.AIPlayers[0].Colonies))
	}
	if len(s.AIPlayers[0].ColonyStars) != beforeStars+1 {
		t.Fatalf("aiExpand 後 AI ColonyStars 應 +1(%d→%d),got %d", beforeStars, beforeStars+1, len(s.AIPlayers[0].ColonyStars))
	}
	if len(s.AIPlayers[0].Colonies) != len(s.AIPlayers[0].ColonyStars) {
		t.Fatalf("Colonies/ColonyStars 平行陣列長度須一致,got %d vs %d", len(s.AIPlayers[0].Colonies), len(s.AIPlayers[0].ColonyStars))
	}
	newColony := s.AIPlayers[0].Colonies[len(s.AIPlayers[0].Colonies)-1]
	if newColony.Population != colonizeStartPopulation || newColony.PopMax < colonizeStartPopulation {
		t.Fatalf("新 AI 殖民地應有實際人口模型(Population=%d PopMax=%d),不應是零值旗標",
			newColony.Population, newColony.PopMax)
	}
	newStarIdx := s.AIPlayers[0].ColonyStars[len(s.AIPlayers[0].ColonyStars)-1]
	if s.Stars[newStarIdx].Owner != 2 {
		t.Fatalf("ColonyStars 記錄的星索引 %d,其 Star.Owner 應為 2(AI),got %d", newStarIdx, s.Stars[newStarIdx].Owner)
	}
}

// TestAIExpand_EconomyGrowsWithColonyCount 驗證 aiExpand 建立的新殖民地會被下一次
// engine.RunEmpireTurn 算進 AI 經濟——EndTurn 之後 AI 總淨工業/FleetStrength 應該因為擴張而
// 比「維持單一母星」時成長更快(對照修前恆定為初始母星產出、線性軍力成長)。
func TestAIExpand_EconomyGrowsWithColonyCount(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true

	for turn := 0; turn < 10; turn++ {
		s.EndTurn()
	}
	if len(s.AIPlayers[0].Colonies) <= 1 {
		t.Fatalf("10 回合(含 2 次 aiExpand 時機:第5、10回合)後 AI 殖民地數應 >1,got %d", len(s.AIPlayers[0].Colonies))
	}
	fleetAfter10 := s.AIPlayers[0].FleetStrength

	for turn := 0; turn < 10; turn++ {
		s.EndTurn()
	}
	fleetAfter20 := s.AIPlayers[0].FleetStrength
	growthSecondDecade := fleetAfter20 - fleetAfter10
	growthFirstDecade := fleetAfter10 // 起始 FleetStrength=0

	if growthSecondDecade <= growthFirstDecade {
		t.Fatalf("殖民地數增加後,AI 軍力成長速度應加快(第11-20回合成長 %d 應大於第1-10回合成長 %d)",
			growthSecondDecade, growthFirstDecade)
	}
}

// TestAIExpand_NoOpWhenNoUnownedStars 驗證所有星都已有歸屬時,aiExpand 安全 no-op(不 panic、
// 不改動任何陣列長度)。
func TestAIExpand_NoOpWhenNoUnownedStars(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	for i := range s.Stars {
		if s.Stars[i].Owner == 0 {
			s.Stars[i].Owner = 1 // 全部標記已有歸屬,模擬版圖已滿
		}
	}
	beforeColonies := len(s.AIPlayers[0].Colonies)
	beforeStars := len(s.AIPlayers[0].ColonyStars)

	s.aiExpand(0) // 不應 panic

	if len(s.AIPlayers[0].Colonies) != beforeColonies || len(s.AIPlayers[0].ColonyStars) != beforeStars {
		t.Fatalf("無主星用完時 aiExpand 應 no-op,Colonies/ColonyStars 不應變動:%d→%d / %d→%d",
			beforeColonies, len(s.AIPlayers[0].Colonies), beforeStars, len(s.AIPlayers[0].ColonyStars))
	}
}

// TestAIStanceHostileWhenStrong 驗證 AI 遠強於玩家(玩家無艦隊)時,關係轉敵對。
func TestAIStanceHostileWhenStrong(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.Ships = nil    // 玩家無軍力
	s.Difficulty = 3 // 不可能難度(倍率高)
	for i := 0; i < 40; i++ {
		s.EndTurn()
	}
	if s.AIPlayers[0].Relation >= 0 {
		t.Fatalf("AI 遠強於玩家時關係應轉負:%d", s.AIPlayers[0].Relation)
	}
	st := s.AIPlayers[0].StanceName
	if st != "宣戰" && st != "敵視" {
		t.Fatalf("AI 強勢時態勢應敵對(宣戰/敵視),實得「%s」", st)
	}
	t.Logf("AI 關係 %d、態勢「%s」", s.AIPlayers[0].Relation, st)
}
