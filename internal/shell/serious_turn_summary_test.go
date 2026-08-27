package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
)

func TestHasSeriousTurnSummaryReportIgnoresOrdinaryEconomy(t *testing.T) {
	s := NewDemoSession()
	s.LastPlayerOutput.TotalNetIndustry = 99
	s.LastBuilt = []BuildNotice{{Kind: BuildNoticeCompleted, ColonyIndex: 0, Name: "ordinary"}}
	if s.HasSeriousTurnSummaryReport() {
		t.Fatal("一般經濟與建造完成不應觸發只顯示重要摘要")
	}
}

func TestHasSeriousTurnSummaryReportRecognizesTypedThreats(t *testing.T) {
	tests := []struct {
		name string
		set  func(*GameSession)
	}{
		{"starvation", func(s *GameSession) { s.LastPlayerOutput.Colonies = []engine.ColonyOutput{{Starving: true}} }},
		{"rebellion", func(s *GameSession) { s.LastRebellions = []RebellionResult{{Triggered: true}} }},
		{"bankruptcy", func(s *GameSession) { s.LastBankruptcy = []BankruptcyAction{{Kind: BankruptcyScrapShip}} }},
		{"antaran", func(s *GameSession) { s.LastAntaranNotice = &AntaranNotice{Kind: AntaranNoticeLaunched} }},
		{"raid", func(s *GameSession) { s.LastRaid = "typed" }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := NewDemoSession()
			tc.set(s)
			if !s.HasSeriousTurnSummaryReport() {
				t.Fatal("重要玩家結果未觸發摘要")
			}
		})
	}
}
