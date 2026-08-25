package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func aiPreferredLeader(id int, ship bool) Leader {
	skill := gamedata.SKILL_FAMOUS
	if ship {
		skill = gamedata.SKILL_HELMSMAN
	}
	return Leader{ID: id, Name: "AI 候選", Level: 2, Ship: ship,
		Skills: []LeaderSkill{{ID: int(skill), Tier: 1}}}
}

func TestAILeaderOfferUsesStrictFiftyBCReserve(t *testing.T) {
	s := NewDemoSession()
	ai := &s.AIPlayers[0]
	ai.Leaders = nil
	leader := aiPreferredLeader(40, false)
	cost := aiLeaderHireCost(ai, leader)

	ai.Player.BC = cost + 50
	ai.LeaderOffer = &leader
	s.processAILeaderOffer(0)
	if len(ai.Leaders) != 0 || s.OfficerCooldowns[leader.ID] != 30 {
		t.Fatalf("等於 cost+50 仍須拒絕並進 30 回合 cooldown：leaders=%+v cooldown=%v", ai.Leaders, s.OfficerCooldowns)
	}

	delete(s.OfficerCooldowns, leader.ID)
	ai.Player.BC = cost + 51
	ai.LeaderOffer = &leader
	s.processAILeaderOffer(0)
	if len(ai.Leaders) != 1 || ai.Player.BC != 51 || ai.LeaderOffer != nil {
		t.Fatalf("高於 reserve 應扣費聘用並清 offer：leaders=%+v BC=%d offer=%+v", ai.Leaders, ai.Player.BC, ai.LeaderOffer)
	}
}

func TestAILeaderSkillGateRejectsUnvaluedCandidate(t *testing.T) {
	s := NewDemoSession()
	ai := &s.AIPlayers[0]
	ai.Leaders = nil
	leader := Leader{ID: 41, Name: "無偏好技能", Level: 1,
		Skills: []LeaderSkill{{ID: int(gamedata.SKILL_FARMING_LEADER), Tier: 1}}}
	ai.Player.BC = 10000
	ai.LeaderOffer = &leader
	s.processAILeaderOffer(0)
	if len(ai.Leaders) != 0 || s.OfficerCooldowns[leader.ID] != 30 {
		t.Fatalf("不在原版 AI 偏好遮罩的技能應拒絕：leaders=%+v cooldown=%v", ai.Leaders, s.OfficerCooldowns)
	}
}

func TestAILeaderOfferIsProcessedOnFollowingPass(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers = s.AIPlayers[:1]
	s.AIPlayers[0].Leaders = nil
	s.AIPlayers[0].Player.BC = 10000
	s.Turn = 200
	s.MercCandidatePool = []Leader{
		aiPreferredLeader(60, false), aiPreferredLeader(61, false), aiPreferredLeader(62, true),
	}
	s.generateAILeaderOffers()
	if s.AIPlayers[0].LeaderOffer == nil || len(s.AIPlayers[0].Leaders) != 0 {
		t.Fatalf("生成 offer 的同一階段不應立即聘用：offer=%+v leaders=%+v",
			s.AIPlayers[0].LeaderOffer, s.AIPlayers[0].Leaders)
	}
	s.processAILeaderOffer(0)
	if s.AIPlayers[0].LeaderOffer != nil || len(s.AIPlayers[0].Leaders) != 1 {
		t.Fatalf("下一次 AI 處理才應聘用並清 offer：offer=%+v leaders=%+v",
			s.AIPlayers[0].LeaderOffer, s.AIPlayers[0].Leaders)
	}
}

func TestAssignAILeadersCoversShipAndColony(t *testing.T) {
	s := NewDemoSession()
	ai := &s.AIPlayers[0]
	ai.Leaders = []Leader{
		aiPreferredLeader(50, true),
		{ID: 51, Name: "AI 管理官", Level: 2, Skills: []LeaderSkill{{ID: int(gamedata.SKILL_RESEARCHER), Tier: 1}}},
	}
	ai.Ships = []Ship{{Name: "小船", Class: "護衛艦", ProductionCost: 40}, {Name: "大船", Class: "戰艦", ProductionCost: 400}}
	ai.Colonies = append(ai.Colonies, ai.Colonies[0])
	ai.Colonies[1].Population = ai.Colonies[0].Population + 5
	s.assignAILeaders(0)
	if ai.Ships[1].OfficerID != 50 || ai.Ships[1].OfficerName == "" {
		t.Fatalf("艦長應任命到最高價值未指派艦艇：%+v", ai.Ships)
	}
	if len(ai.ColonyLeaderNames) != 2 || ai.ColonyLeaderNames[1] != "AI 管理官" {
		t.Fatalf("管理官應任命到最高分殖民地：%+v", ai.ColonyLeaderNames)
	}
	if ai.Leaders[0].RawStatus != 1 || ai.Leaders[1].RawStatus != 1 {
		t.Fatalf("任命後應寫回 active status：%+v", ai.Leaders)
	}
}

func TestOfficerCooldownExpiresAfterThirtyTurns(t *testing.T) {
	s := NewDemoSession()
	s.OfficerCooldowns = map[int]int{9: 30}
	for i := 0; i < 29; i++ {
		s.advanceOfficerCooldowns()
	}
	if s.OfficerCooldowns[9] != 1 {
		t.Fatalf("29 回合後應剩 1，got %v", s.OfficerCooldowns)
	}
	s.advanceOfficerCooldowns()
	if _, ok := s.OfficerCooldowns[9]; ok {
		t.Fatalf("第 30 回合應解除 cooldown：%v", s.OfficerCooldowns)
	}
}

func TestPlayerOnlyMercPassDoesNotAdvanceWorldCooldown(t *testing.T) {
	s := NewDemoSession()
	s.Turn = 4 // 玩家 offer gate 為零，隔離本測試。
	s.OfficerCooldowns = map[int]int{9: 12}
	s.advancePlayerMercOffer()
	if s.OfficerCooldowns[9] != 12 {
		t.Fatalf("熱座玩家側 pass 不得重複推進世界 cooldown：%v", s.OfficerCooldowns)
	}
}
