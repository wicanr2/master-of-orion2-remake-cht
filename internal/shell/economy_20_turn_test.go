package shell

import "testing"

// TestOpeningEconomy20TurnProbe 是一次固定開局的 20 回合體感探針：只記錄國庫、
// 人口、士氣、食物、工業與研究軌跡，不在這裡偷偷調公式。它的目的，是把「開局
// 是否出現明顯饑荒／負士氣／國庫失控」變成可重跑的抽樣證據，平衡調整仍需玩家
// 實際操作後另立決策。
func TestOpeningEconomy20TurnProbe(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	t.Log("turn BC deltaBC population morale food industry research")
	for turn := 1; turn <= 20; turn++ {
		beforeBC := s.Player.BC
		s.EndTurn()
		if len(s.PlayerColonies) == 0 || len(s.LastPlayerOutput.Colonies) == 0 {
			t.Fatalf("第 %d 回合失去玩家殖民地輸出", turn)
		}
		colony := s.PlayerColonies[0]
		out := s.LastPlayerOutput.Colonies[0]
		if colony.Population < 1 {
			t.Fatalf("第 %d 回合人口低於 1: %d", turn, colony.Population)
		}
		if colony.MoralePercent < -100 || colony.MoralePercent > 100 {
			t.Fatalf("第 %d 回合士氣超出安全範圍: %d%%", turn, colony.MoralePercent)
		}
		t.Logf("%02d %d %+d %d %d %d %d %d", turn, s.Player.BC, s.Player.BC-beforeBC,
			colony.Population, colony.MoralePercent, out.FoodSurplus, out.NetIndustry, out.Research)
	}
}
