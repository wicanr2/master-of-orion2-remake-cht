package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func queueRequestForTest(s *GameSession, reason int, action gamedata.OriginalHumanDiplomaticAction) {
	s.AIPlayers[0].WantsAudience = true
	s.AIPlayers[0].AudienceReason = AudienceReasonOriginal
	s.AIPlayers[0].OriginalHumanDiplomaticRequest = &gamedata.OriginalHumanDiplomaticRequest{
		Outcome: 1, ReasonCode: reason, Action: action,
	}
}

func TestAcceptOriginalHumanCreditRequestUsesOriginalClamp(t *testing.T) {
	s := NewDemoSession()
	s.Player.BC, s.AIPlayers[0].Player.BC = 30, 10
	queueRequestForTest(s, 106, gamedata.OriginalHumanDiplomaticAction{
		Kind: gamedata.OriginalHumanDiplomaticActionCredits, Credits: 50,
	})
	if !s.AcceptOriginalAIHumanDiplomaticRequest(0) {
		t.Fatal("合法 BC 要求應可接受")
	}
	if s.Player.BC != 0 || s.AIPlayers[0].Player.BC != 60 {
		t.Fatalf("BC transfer player/AI=%d/%d，預期 0/60", s.Player.BC, s.AIPlayers[0].Player.BC)
	}
	if s.AIPlayers[0].OriginalHumanDiplomaticRequest != nil || s.AIPlayers[0].WantsAudience {
		t.Fatal("接受後必須清掉 request 與會談燈")
	}
}

func TestAcceptOriginalHumanTechnologyAndDirectRequests(t *testing.T) {
	s := NewDemoSession()
	tech := gamedata.TECH_ARMOR_BARRACKS
	topic, ok := gamedata.OrigTechTopic(tech)
	if !ok {
		t.Fatal("測試科技缺 topic")
	}
	grantTechnologyApplication(&s.Player, topic, tech)
	queueRequestForTest(s, 105, gamedata.OriginalHumanDiplomaticAction{
		Kind: gamedata.OriginalHumanDiplomaticActionTechnology, Technology: int(tech),
	})
	if !s.AcceptOriginalAIHumanDiplomaticRequest(0) || !playerStateKnowsTech(s.AIPlayers[0].Player, topic, tech) {
		t.Fatal("接受科技要求後 AI 應取得 application")
	}
	queueRequestForTest(s, 106, gamedata.OriginalHumanDiplomaticAction{
		Kind: gamedata.OriginalHumanDiplomaticActionDirect, DirectTier: 2,
	})
	if !s.AcceptOriginalAIHumanDiplomaticRequest(0) || s.AIPlayers[0].OriginalHumanDirectRequestTier != 2 {
		t.Fatal("direct tier 2 應寫入 sub_52049 對映狀態")
	}
	if got := s.snapshot().restore().AIPlayers[0].OriginalHumanDirectRequestTier; got != 2 {
		t.Fatalf("direct tier snapshot=%d，預期 2", got)
	}
}

func TestAcceptOriginalHumanColonyRequestTransfersTypedState(t *testing.T) {
	s := NewDemoSession()
	colony := s.PlayerColonies[0]
	s.PlayerColonies = append(s.PlayerColonies, colony)
	s.PlayerColonyStars = append(s.PlayerColonyStars, 1)
	s.PlayerColonyPlanets = []int{0, 1}
	s.ColonyBuildings = append(s.ColonyBuildings, map[string]bool{"test": true})
	s.PlayerCapitolPlanet, s.PlayerCapitolPlanetKnown = 0, true
	s.Stars[1].Owner = 1
	before := len(s.AIPlayers[0].Colonies)
	queueRequestForTest(s, 105, gamedata.OriginalHumanDiplomaticAction{
		Kind: gamedata.OriginalHumanDiplomaticActionColony, Colony: 1,
	})
	if !s.AcceptOriginalAIHumanDiplomaticRequest(0) {
		t.Fatal("合法非首都殖民星要求應可接受")
	}
	if len(s.PlayerColonies) != 1 || len(s.AIPlayers[0].Colonies) != before+1 || s.Stars[1].Owner != 2 {
		t.Fatalf("colony transfer player=%d AI=%d owner=%d", len(s.PlayerColonies), len(s.AIPlayers[0].Colonies), s.Stars[1].Owner)
	}
}

func TestOutcomeFourIsNoticeNotAcceptRequest(t *testing.T) {
	s := NewDemoSession()
	queueRequestForTest(s, 124, gamedata.OriginalHumanDiplomaticAction{
		Kind: gamedata.OriginalHumanDiplomaticActionCredits, Credits: 100,
	})
	before := s.Player.BC
	if s.AcceptOriginalAIHumanDiplomaticRequest(0) {
		t.Fatal("reason 124 不得進二選一接受 callback")
	}
	if !s.AcknowledgeOriginalAIHumanDiplomaticNotice(0) || s.Player.BC != before {
		t.Fatal("reason 124 應只清除通知，不套用 payload")
	}
}

func TestRejectOriginalHumanReason105AppliesChangeRelationScore(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Relation, a.OriginalRelationRaw, a.OriginalRelationKnown = 0, 0, true
	queueRequestForTest(s, 105, gamedata.OriginalHumanDiplomaticAction{Kind: gamedata.OriginalHumanDiplomaticActionDirect, DirectTier: 1})
	want, ok := gamedata.OriginalChangeRelationScore(gamedata.OriginalRelationChangeInput{
		CurrentRaw: 0, BaseDelta: -50, ActorGovernment: int(s.Government),
		TargetCharismatic: aiRaceHasTrait(*a, gamedata.TRAIT_CHARISMATIC), Policy: a.Treaty.FormalPolicy,
	})
	if !ok || !s.RejectOriginalAIHumanDiplomaticRequest(0) {
		t.Fatal("reason 105 應可走原版 Change_Relations_ 拒絕 callback")
	}
	if a.OriginalRelationRaw != want || a.Relation != normalizedRelationFromOriginal(want) || a.WantsAudience {
		t.Fatalf("reason105 raw/shown/audience=%d/%d/%v，預期 %d/%d/false", a.OriginalRelationRaw, a.Relation, a.WantsAudience, want, normalizedRelationFromOriginal(want))
	}
}

func TestRejectOriginalHumanReason106RequiresAndCommitsMilitaryCandidate(t *testing.T) {
	s := NewDemoSession()
	queueRequestForTest(s, 106, gamedata.OriginalHumanDiplomaticAction{Kind: gamedata.OriginalHumanDiplomaticActionDirect, DirectTier: 1})
	if s.RejectOriginalAIHumanDiplomaticRequest(0) {
		t.Fatal("+0x837 軍事候選 unknown 時不得冒充 -1 宣戰")
	}
	a := &s.AIPlayers[0]
	a.OriginalHumanMilitaryCandidateStar = 3
	a.OriginalHumanMilitaryCandidateReason = 112
	a.OriginalHumanMilitaryCandidateKnown = true
	if !s.RejectOriginalAIHumanDiplomaticRequest(0) || !a.OriginalHumanMilitaryTargetKnown ||
		a.OriginalHumanMilitaryTargetStar != 3 || a.OriginalHumanMilitaryTargetReason != 112 {
		t.Fatalf("reason106 target=%d/%d known=%v", a.OriginalHumanMilitaryTargetStar, a.OriginalHumanMilitaryTargetReason, a.OriginalHumanMilitaryTargetKnown)
	}
	got := s.snapshot().restore().AIPlayers[0]
	if !got.OriginalHumanMilitaryTargetKnown || got.OriginalHumanMilitaryTargetStar != 3 || got.OriginalHumanMilitaryTargetReason != 112 {
		t.Fatalf("snapshot target=%d/%d known=%v", got.OriginalHumanMilitaryTargetStar, got.OriginalHumanMilitaryTargetReason, got.OriginalHumanMilitaryTargetKnown)
	}
}

func TestRejectOriginalHumanReason106WithoutCandidateDeclaresWar(t *testing.T) {
	s := NewDemoSession()
	queueRequestForTest(s, 106, gamedata.OriginalHumanDiplomaticAction{Kind: gamedata.OriginalHumanDiplomaticActionDirect, DirectTier: 1})
	a := &s.AIPlayers[0]
	a.OriginalHumanMilitaryCandidateStar = -1
	a.OriginalHumanMilitaryCandidateKnown = true
	if !s.RejectOriginalAIHumanDiplomaticRequest(0) {
		t.Fatal("已證實 +0x837=-1 時應走 sub_51078 宣戰")
	}
	if a.Treaty.FormalPolicy < gamedata.DIPLO_LIMITED_WAR || a.WantsAudience || a.OriginalHumanDiplomaticRequest != nil {
		t.Fatalf("war/request=%v/%v/%v", a.Treaty.FormalPolicy, a.WantsAudience, a.OriginalHumanDiplomaticRequest)
	}
}
