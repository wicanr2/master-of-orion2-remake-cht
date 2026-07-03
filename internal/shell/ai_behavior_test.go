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
