package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestAIVsAIOriginalNegotiationCreatesAllianceAndResearch(t *testing.T) {
	s := NewDemoSession()
	s.ensureOriginalAIAIRelations()
	s.ensureAIAIState()
	s.setOriginalAIAIRelation(0, 1, 100)
	s.AIReputationRaw[0][1] = 100
	s.AITreatyBiasRaw[0][1] = 100
	s.AIAgreementBiasRaw[0][1] = 100
	s.AIPolicies[0][1], s.AIPolicies[1][0] = gamedata.DIPLO_NON_AGGRESSION, gamedata.DIPLO_NON_AGGRESSION
	s.AITrade[0][1], s.AITrade[1][0] = true, true
	s.AIPlayers[0].FleetStrength, s.AIPlayers[1].FleetStrength = 100, 100
	s.advanceOriginalAIAINegotiation(0, 1, func(n int) int {
		if n == 250-40*s.Difficulty {
			return 1
		}
		return n
	})
	if s.AIPolicies[0][1] != gamedata.DIPLO_ALLIANCE || !s.AIResearch[0][1] {
		t.Fatalf("原版分數應升同盟並建立研究協議：policy=%v research=%v", s.AIPolicies, s.AIResearch)
	}
	if s.AITreatyBiasRaw[0][1] != 70 || s.AIAgreementBiasRaw[0][1] != 70 {
		t.Fatalf("談判後 raw 記憶應各扣 30：treaty=%v agreement=%v",
			s.AITreatyBiasRaw, s.AIAgreementBiasRaw)
	}
}

func TestAIVsAIWarDoesNotUseInventedRelationCeasefire(t *testing.T) {
	s := NewDemoSession()
	s.ensureAIAIState()
	s.AIWars[0][1], s.AIWars[1][0] = true, true
	s.AIPolicies[0][1], s.AIPolicies[1][0] = gamedata.DIPLO_WAR, gamedata.DIPLO_WAR
	s.AIRelations[0][1], s.AIRelations[1][0] = 40, 40
	s.advanceAIDiplomacy()
	if !s.AIWars[0][1] || s.AIPolicies[0][1] != gamedata.DIPLO_WAR {
		t.Fatalf("高顯示關係不得以自編 +12 門檻自動停戰：wars=%v policy=%v",
			s.AIWars, s.AIPolicies)
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
