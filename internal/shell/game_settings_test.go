package shell

import (
	"path/filepath"
	"testing"
)

func TestDefaultGameSettingsMatchesOriginalInitializer(t *testing.T) {
	g := DefaultGameSettings()
	if !g.EndOfTurnSummary || !g.EndOfTurnWait || g.EnemyMoves || g.ExpandingHelp ||
		!g.AutoSelectShips || !g.Animations || g.AutoSelectColony || !g.ShowRelocationLines ||
		!g.ShowGNNReport || g.AutoDeleteTradeGoodHousing || !g.AutoSaveGame ||
		g.ShowOnlySeriousTurnSummary || g.ShipInitiative {
		t.Fatalf("SETTINGS 預設值偏離 sub_127E1：%+v", g)
	}
}

func TestGameSettingsSaveRoundTrip(t *testing.T) {
	s := NewDemoSession()
	g := s.EffectiveGameSettings()
	g.EnemyMoves, g.AutoSaveGame, g.ShipInitiative = true, false, true
	s.ApplyGameSettings(g)
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSession(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.EffectiveGameSettings() != g {
		t.Fatalf("設定往返不符：got=%+v want=%+v", got.EffectiveGameSettings(), g)
	}
}
