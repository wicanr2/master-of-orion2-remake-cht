package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestOriginalAIMaintenanceTotalUsesPlayerB4Components(t *testing.T) {
	s := &GameSession{AIPlayers: []AIOpponent{{Player: engine.PlayerState{
		Maintenance: 7, ActiveFreighters: 5,
		CommandPointsSupply: 2, UsedCommandPoints: 5, CommandOverflowCostPerPoint: 9,
		SpyMaintenance: 3, OfficerMaintenance: 4,
	}}}}
	got, ok := s.originalAIMaintenanceTotal(0)
	want := 7 + gamedata.IncomeFreighterMaintenanceCost(5) + 3*9 + 3 + 4
	if !ok || got != want {
		t.Fatalf("maintenance=%d/%v，預期 %d", got, ok, want)
	}
	s.AIPlayers[0].Treaty.AITribute = TributeFivePercent
	if _, ok := s.originalAIMaintenanceTotal(0); ok {
		t.Fatal("缺本回合 +0xC0 納貢分項時必須失敗即關閉")
	}
}

func TestOriginalAIHumanTechnologyCandidatesConsumeTargetKnownApplications(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) == 0 || !s.AIPlayers[0].OriginalTechProfileKnown {
		t.Fatal("demo 應提供已知 AI tech profile")
	}
	a := &s.AIPlayers[0]
	sourceKnown := knownTechnologyApplications(a.Player)
	state := gamedata.OriginalStartingValueState{
		Difficulty: s.Difficulty, RelativeTurn: s.Turn,
		AIProfile: a.OriginalTechProfile, AIProfileKnown: true,
		Raw4: a.OriginalTechProfile.Raw4, Raw4Known: true,
		Known: sourceKnown, Opponents: []map[gamedata.Technology]bool{},
	}
	wanted := gamedata.TECH_NONE
	for raw := 1; raw < len(gamedata.TechItemCategory); raw++ {
		tech := gamedata.Technology(raw)
		if !sourceKnown[tech] && gamedata.OriginalAITechValueKnownSlice(tech, state) > 0 {
			wanted = tech
			break
		}
	}
	if wanted == gamedata.TECH_NONE {
		t.Fatal("測試資料找不到來源未知且估值為正的科技")
	}
	if s.Player.GrantedTechs == nil {
		s.Player.GrantedTechs = map[gamedata.Technology]bool{}
	}
	s.Player.GrantedTechs[wanted] = true
	candidates, ratio, ok := s.originalAIHumanTechnologyCandidates(0)
	if !ok || ratio < 1 || ratio > 10 {
		t.Fatalf("candidates=%v ratio=%d ok=%v", candidates, ratio, ok)
	}
	found := false
	for _, raw := range candidates {
		if gamedata.Technology(raw) == wanted {
			found = true
		}
	}
	if !found {
		t.Fatalf("target 獨有科技 %v 未進候選：%v", wanted, candidates)
	}
}
