package shell

import (
	"path/filepath"
	"testing"
)

// TestSaveLoadRoundTrip 驗證存檔→讀檔後對局狀態一致,且讀回的 AI 可續跑、事件/成長系統續行。
func TestSaveLoadRoundTrip(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	// 推進若干回合並選種族,製造非初始狀態。
	s.ApplyRace(3) // 克拉肯
	for i := 0; i < 12; i++ {
		s.EndTurn()
	}
	s.SendFleet(5)
	s.Builds[0] = ColonyBuild{Name: "自動工廠", Progress: 20, Cost: 60}

	path := filepath.Join(t.TempDir(), "save.json")
	if err := s.Save(path); err != nil {
		t.Fatalf("存檔失敗: %v", err)
	}
	if !SaveExists(path) {
		t.Fatal("存檔應存在")
	}

	got, err := LoadSession(path)
	if err != nil {
		t.Fatalf("讀檔失敗: %v", err)
	}

	// 逐項核對關鍵狀態。
	if got.Turn != s.Turn {
		t.Errorf("Turn 不符:%d vs %d", got.Turn, s.Turn)
	}
	if got.Player.BC != s.Player.BC {
		t.Errorf("BC 不符:%d vs %d", got.Player.BC, s.Player.BC)
	}
	if got.RaceIndex != s.RaceIndex {
		t.Errorf("種族不符:%d vs %d", got.RaceIndex, s.RaceIndex)
	}
	if len(got.Stars) != len(s.Stars) {
		t.Errorf("星數不符:%d vs %d", len(got.Stars), len(s.Stars))
	}
	if len(got.PlayerColonies) != len(s.PlayerColonies) ||
		got.PlayerColonies[0].Population != s.PlayerColonies[0].Population {
		t.Errorf("殖民地人口不符:%d vs %d", got.PlayerColonies[0].Population, s.PlayerColonies[0].Population)
	}
	if got.FleetDestStar != s.FleetDestStar || got.FleetETA != s.FleetETA {
		t.Errorf("艦隊航行狀態不符")
	}
	if got.Builds[0].Name != s.Builds[0].Name || got.Builds[0].Progress != s.Builds[0].Progress {
		t.Errorf("建造佇列不符")
	}
	if len(got.AIPlayers) != len(s.AIPlayers) {
		t.Fatalf("AI 數不符:%d vs %d", len(got.AIPlayers), len(s.AIPlayers))
	}
	if got.AIPlayers[0].Decider == nil {
		t.Fatal("讀回的 AI Decider 應重建,不為 nil")
	}

	// 讀回的對局應可續跑一回合而不 panic。
	got.EndTurn()
	if got.Turn != s.Turn+1 {
		t.Errorf("讀回對局續跑後 Turn 應 +1:%d", got.Turn)
	}
	t.Logf("存讀檔往返一致(Turn %d、BC %d、種族 %s、%d 星)", got.Turn-1, got.Player.BC, Races[got.RaceIndex].Name, len(got.Stars))
}

// TestLoadRejectsMissing 驗證讀取不存在的檔回傳錯誤。
func TestLoadRejectsMissing(t *testing.T) {
	if _, err := LoadSession(filepath.Join(t.TempDir(), "nope.json")); err == nil {
		t.Fatal("讀取不存在的存檔應回傳錯誤")
	}
}
