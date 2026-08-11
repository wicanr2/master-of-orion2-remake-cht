package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
)

// TestCyberneticBuildKeepsHalfIndustryCarry 驗證半機械族的奇數半單位生產力不會在建造
// 進度轉回整數時遺失；這是舊 Progress 欄位無法表達、ProgressHalf 新欄位負責的狀態。
func TestCyberneticBuildKeepsHalfIndustryCarry(t *testing.T) {
	s := NewDemoSession()
	s.Player.TaxRate = 0
	s.Builds[0] = ColonyBuild{Name: "自動工廠", Cost: 100}
	s.LastPlayerOutput = engine.EmpireOutput{
		Colonies: []engine.ColonyOutput{{Cybernetic: true, NetIndustry: 2, NetIndustryHalf: 5}},
	}

	s.advanceBuilds()
	if got := s.Builds[0].Progress; got != 2 {
		t.Fatalf("第一回合整數建造進度 = %d,預期 2", got)
	}
	if got := s.Builds[0].ProgressHalf; got != 1 {
		t.Fatalf("第一回合半單位餘數 = %d,預期 1", got)
	}
	if got := s.BuildETATurns(0); got != 39 {
		t.Fatalf("半單位建造 ETA = %d,預期 39", got)
	}

	s.advanceBuilds()
	if got := s.Builds[0].Progress; got != 5 {
		t.Fatalf("第二回合應兌現餘數後進度 5,實得 %d", got)
	}
	if got := s.Builds[0].ProgressHalf; got != 0 {
		t.Fatalf("第二回合半單位餘數應歸零,實得 %d", got)
	}
}
