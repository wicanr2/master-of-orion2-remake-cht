package shell

import "testing"

func TestOriginalAIStartsAtZeroTax(t *testing.T) {
	aiPlayers := buildDemoAIOpponents([]int{1}, 2, 1, 42)
	if len(aiPlayers) != 1 {
		t.Fatalf("AI 數量 = %d，預期 1", len(aiPlayers))
	}
	if got := aiPlayers[0].Player.TaxRate; got != 0 {
		t.Fatalf("原版 AI 開局稅率 = %d，預期清零初始化的 0", got)
	}
}

func TestOriginalAITurnPreservesImportedTaxRate(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) == 0 {
		t.Fatal("示範局沒有 AI")
	}
	s.AIPlayers[0].Player.TaxRate = 40
	s.EndTurn()
	if got := s.AIPlayers[0].Player.TaxRate; got != 40 {
		t.Fatalf("AI 回合覆蓋既有稅率：實得 %d，預期保持 40", got)
	}
}
