package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestAIVsAIDiplomacyCreatesWarAndCeasefire(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers[0].FleetStrength = 0
	s.AIPlayers[1].FleetStrength = 0
	s.AIRelations[0][1], s.AIRelations[1][0] = -30, -30
	s.advanceAIDiplomacy()
	if !s.AIWars[0][1] || !s.AIWars[1][0] {
		t.Fatalf("負關係應建立對稱戰爭：wars=%v", s.AIWars)
	}
	if s.AIPolicies[0][1] != gamedata.DIPLO_WAR || s.AITrade[0][1] {
		t.Fatalf("戰爭政策／貿易旗標不對：policy=%v trade=%v", s.AIPolicies, s.AITrade)
	}

	s.AIRelations[0][1], s.AIRelations[1][0] = 20, 20
	s.advanceAIDiplomacy()
	if s.AIWars[0][1] || s.AIWars[1][0] {
		t.Fatalf("恢復關係後應停戰：wars=%v", s.AIWars)
	}
	if s.AIPolicies[0][1] != gamedata.DIPLO_NON_AGGRESSION || !s.AITrade[0][1] {
		t.Fatalf("停戰後應有互不侵犯／貿易：policy=%v trade=%v", s.AIPolicies, s.AITrade)
	}
}

func TestAIVsAIWarFleetTransfersColony(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.EnableAIVsAI = true
	s.ensureAIAIState()
	attacker, defender := 0, 1
	from := s.AIPlayers[attacker].ColonyStars[0]
	target := s.AIPlayers[defender].ColonyStars[0]
	s.AIWars[attacker][defender] = true
	s.AIWars[defender][attacker] = true
	s.AIPlayers[attacker].FleetStrength = 1000
	s.AIPlayers[defender].FleetStrength = 1
	s.AIPlayers[attacker].FleetStar = from
	s.AIPlayers[attacker].FleetPosSet = true
	s.AIPlayers[attacker].FleetDestStar = target
	s.AIPlayers[attacker].FleetETA = 1
	s.AIPlayers[attacker].FleetTargetAI = defender
	s.AIPlayers[attacker].FleetTargetAISet = true

	s.advanceAIFleets()
	if s.LastAIAIBattle == nil || !s.LastAIAIBattle.AttackerWon {
		t.Fatalf("應產生攻方勝利報告：%+v", s.LastAIAIBattle)
	}
	if len(s.AIPlayers[attacker].ColonyStars) != 2 || len(s.AIPlayers[defender].ColonyStars) != 0 {
		t.Fatalf("殖民地應轉移：attacker=%v defender=%v", s.AIPlayers[attacker].ColonyStars, s.AIPlayers[defender].ColonyStars)
	}
	if s.AIPlayers[attacker].FleetTargetAISet || s.AIPlayers[attacker].FleetETA != 0 {
		t.Fatalf("抵達後艦隊目標應清除：%+v", s.AIPlayers[attacker])
	}
}
