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
	if !s.AIWars[0][1] || s.AIPolicies[0][1] != gamedata.DIPLO_LIMITED_WAR {
		t.Fatalf("高顯示關係不得以自編 +12 門檻自動停戰：wars=%v policy=%v",
			s.AIWars, s.AIPolicies)
	}
}

func TestOriginalAIAIWarDeclarationWritesRawState(t *testing.T) {
	s := NewDemoSession()
	s.ensureOriginalAIAIRelations()
	s.ensureAIAIState()
	s.AITrade[0][1], s.AITrade[1][0] = true, true
	s.AIResearch[0][1], s.AIResearch[1][0] = true, true
	s.AITributeModes[0][1], s.AITributeModes[1][0] = 2, 1
	s.declareOriginalAIAIWar(0, 1, func(n int) int {
		if n != 25 {
			t.Fatalf("宣戰關係亂數 Random(%d)，應為 25", n)
		}
		return 25
	})
	if !s.AIWars[0][1] || s.AIPolicies[0][1] != gamedata.DIPLO_LIMITED_WAR ||
		s.AITrade[0][1] || s.AIResearch[0][1] || s.AITributeModes[0][1] != 0 {
		t.Fatalf("宣戰正式狀態錯誤：wars=%v policy=%v trade=%v research=%v tribute=%v",
			s.AIWars, s.AIPolicies, s.AITrade, s.AIResearch, s.AITributeModes)
	}
	if got := s.originalAIAIRelation(0, 1); got != -99 {
		t.Fatalf("宣戰關係=%d，應為 -99", got)
	}
	if s.AITreatyBiasRaw[0][1] != -200 || s.AIAgreementBiasRaw[1][0] != -200 ||
		s.AIWarDurationRaw[0][1] != 0 || s.AIDiplomacyCooldownRaw[1][0] != 0 {
		t.Fatalf("宣戰 raw 狀態錯誤：treaty=%v agreement=%v duration=%v cooldown=%v",
			s.AITreatyBiasRaw, s.AIAgreementBiasRaw, s.AIWarDurationRaw, s.AIDiplomacyCooldownRaw)
	}
}

func TestOriginalAIAIHostilityWarCandidateReachesDeclaration(t *testing.T) {
	s := NewDemoSession()
	s.ensureOriginalAIAIRelations()
	s.ensureAIAIState()
	s.Difficulty = 2
	s.Turn = 1 // 三位 AI 時 target 1 是 reason 68 的輪值目標。
	s.setOriginalAIAIRelation(0, 1, -100)
	s.AIPlayers[0].FleetStrength = 10000
	s.AIPlayers[1].FleetStrength = 1
	s.AIPlayers[2].FleetStrength = 1
	s.advanceOriginalAIAIWarPolicy(func(n int) int {
		switch n {
		case 100, 1, 25:
			return 1
		case 200:
			return 200
		default:
			return n
		}
	})
	if !s.AIWars[0][1] || s.AIPolicies[0][1] != gamedata.DIPLO_LIMITED_WAR {
		t.Fatalf("reason 68 應抵達正式宣戰 consumer：wars=%v policy=%v", s.AIWars, s.AIPolicies)
	}
}

func TestOriginalAIAIWarPolicyHonorsHyperspaceFluxTransDimensionalGate(t *testing.T) {
	newCandidate := func(transDimensional bool) *GameSession {
		s := NewDemoSession()
		s.ensureOriginalAIAIRelations()
		s.ensureAIAIState()
		s.Difficulty, s.Turn = 2, 1
		s.PersistentEvents = []PersistentEvent{{Kind: PersistentHyperspaceFlux, StarIndex: -1}}
		s.AIPlayers[0].RaceIndex = 0
		if transDimensional {
			for i := range Races {
				if gamedata.OrigRaceHasTrait(Races[i].OrigIdx, gamedata.TRAIT_TRANS_DIMENSIONAL) {
					s.AIPlayers[0].RaceIndex = i
					break
				}
			}
		}
		s.setOriginalAIAIRelation(0, 1, -100)
		s.AIPlayers[0].FleetStrength = 10000
		s.AIPlayers[1].FleetStrength = 1
		s.AIPlayers[2].FleetStrength = 1
		return s
	}
	roll := func(n int) int {
		switch n {
		case 100, 1, 25:
			return 1
		case 200:
			return 200
		default:
			return n
		}
	}
	blocked := newCandidate(false)
	blocked.advanceOriginalAIAIWarPolicy(roll)
	if blocked.AIWars[0][1] {
		t.Fatal("亂流中非跨維度 AI 不得進入宣戰候選鏈")
	}
	immune := newCandidate(true)
	immune.advanceOriginalAIAIWarPolicy(roll)
	if !immune.AIWars[0][1] {
		t.Fatal("亂流中跨維度 AI 應繞過 gate 並可正常宣戰")
	}
}

func TestOriginalAIAICeasefireUsesDurationAndCooldown(t *testing.T) {
	s := NewDemoSession()
	s.ensureOriginalAIAIRelations()
	s.ensureAIAIState()
	s.Difficulty = 4 // 門檻 30。
	s.setOriginalAIAIRelation(0, 1, -80)
	s.AIWars[0][1], s.AIWars[1][0] = true, true
	s.AIPolicies[0][1], s.AIPolicies[1][0] = gamedata.DIPLO_LIMITED_WAR, gamedata.DIPLO_LIMITED_WAR
	s.AIWarDurationRaw[0][1], s.AIWarDurationRaw[1][0] = 30, 30
	s.advanceAIAIDiplomacy()
	if s.AIWars[0][1] || s.AIPolicies[0][1] != gamedata.DIPLO_PEACE ||
		s.AIDiplomacyCooldownRaw[0][1] != 30 {
		t.Fatalf("停戰狀態錯誤：wars=%v policy=%v cooldown=%v", s.AIWars, s.AIPolicies, s.AIDiplomacyCooldownRaw)
	}
	if got := s.originalAIAIRelation(0, 1); got != -30 {
		t.Fatalf("停戰關係=%d，應為 -30", got)
	}
	for i := 0; i < 30; i++ {
		s.advanceOriginalAIAIWarTimers()
	}
	if s.AIDiplomacyCooldownRaw[0][1] != 0 || s.AIPolicies[0][1] != gamedata.DIPLO_NONE {
		t.Fatalf("30 回合後應解除暫時和平：policy=%v cooldown=%v", s.AIPolicies, s.AIDiplomacyCooldownRaw)
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
