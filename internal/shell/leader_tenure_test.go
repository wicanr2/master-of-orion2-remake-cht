package shell

import "testing"

func TestAdvanceLeaderLimboUsesRawStatusAndThirtyTurnThreshold(t *testing.T) {
	s := NewDemoSession()
	s.Leaders = []Leader{
		{Name: "待清除", RawStatus: originalLeaderLimboStatus, RawETA: originalLeaderLimboThreshold - 1},
		{Name: "仍在閒置", RawStatus: originalLeaderLimboStatus, RawETA: originalLeaderLimboThreshold - 2},
		{Name: "活躍任命", RawStatus: 1, RawETA: originalLeaderLimboThreshold - 1},
	}

	if got := s.advanceLeaderLimbo(); got != 1 {
		t.Fatalf("達到 raw status=4 門檻者應清除 1 人,got %d", got)
	}
	if len(s.Leaders) != 2 {
		t.Fatalf("清除後應剩 2 位領袖,got %+v", s.Leaders)
	}
	if s.Leaders[0].Name != "仍在閒置" || s.Leaders[0].RawETA != originalLeaderLimboThreshold-1 {
		t.Fatalf("未達門檻者應只遞增一次,got %+v", s.Leaders[0])
	}
	if s.Leaders[1].Name != "活躍任命" || s.Leaders[1].RawETA != originalLeaderLimboThreshold-2 {
		t.Fatalf("status=1 應走 active ETA 遞減而非 status=4 閒置門檻,got %+v", s.Leaders[1])
	}
}

func TestAdvanceLeaderLimboWithdrawsColonyBonusAndShipAssignment(t *testing.T) {
	s := NewDemoSession()
	s.Leaders = []Leader{{Name: "解除官", Skill: "科學家", Tier: 1, RawStatus: originalLeaderLimboStatus, RawETA: originalLeaderLimboThreshold - 1}}
	s.ensureColonyLeaderSlots()
	s.ColonyLeaderNames[0] = "解除官"
	s.PlayerColonies[0].FlatResearch = 7
	if len(s.Fleets) == 0 {
		s.Fleets = []Fleet{{Ships: []Ship{{}}}}
	} else if len(s.Fleets[0].Ships) == 0 {
		s.Fleets[0].Ships = []Ship{{}}
	}
	s.Fleets[0].Ships[0].OfficerName = "解除官"
	s.Fleets[0].Ships[0].OfficerID = 0

	if got := s.advanceLeaderLimbo(); got != 1 {
		t.Fatalf("應清除解除官,got %d", got)
	}
	if len(s.ColonyLeaderNames) == 0 || s.ColonyLeaderNames[0] != "" {
		t.Fatalf("殖民地領袖指派應撤回,got %q", s.ColonyLeaderNames[0])
	}
	if s.Fleets[0].Ships[0].OfficerName != "" || s.Fleets[0].Ships[0].OfficerID != 0 {
		t.Fatal("艦艇軍官指派應一併清除")
	}
}

func TestAdvanceActiveLeaderETADecrementsAndPreservesLeader(t *testing.T) {
	s := NewDemoSession()
	s.Leaders = []Leader{{Name: "殖民地軍官", RawStatus: originalLeaderActiveStatus, RawETA: 2, RawLocation: 1}}
	if got := s.advanceLeaderLimbo(); got != 0 {
		t.Fatalf("active ETA 不應走 status=4 清除路徑,got released=%d", got)
	}
	if len(s.Leaders) != 1 || s.Leaders[0].RawETA != 1 {
		t.Fatalf("active ETA 應由 2 遞減為 1 且保留領袖,got %+v", s.Leaders)
	}
	if got := s.advanceLeaderLimbo(); got != 0 {
		t.Fatalf("active ETA 歸零也不應直接清除領袖,got released=%d", got)
	}
	if len(s.Leaders) != 1 || s.Leaders[0].RawETA != 0 {
		t.Fatalf("active ETA=1 應遞減至 0 且保留領袖,got %+v", s.Leaders)
	}
}

func TestAdvanceActiveLeaderETAAlsoCoversAILeaders(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers[0].Leaders = []Leader{{Name: "AI軍官", RawStatus: originalLeaderActiveStatus, RawETA: 1, RawLocation: 1}}
	s.advanceLeaderLimbo()
	if got := s.AIPlayers[0].Leaders[0].RawETA; got != 0 {
		t.Fatalf("AI active ETA 應同步遞減,got %d", got)
	}
}

func TestAdvanceActiveLeaderETACallsColonyCalculationApproximation(t *testing.T) {
	s := NewDemoSession()
	leader := Leader{
		Name: "到期殖民官", Skill: "農業官", Level: 3, Tier: 1,
		RawStatus: originalLeaderActiveStatus, RawETA: 1, RawLocation: 1,
	}
	s.Leaders = []Leader{leader}
	s.ensureColonyLeaderSlots()
	s.ColonyLeaderNames[0] = leader.Name
	applyColonyLeaderBonusDelta(&s.PlayerColonies[0], colonyLeaderBonusFor(leader), 1)
	wantMorale := colonyMoralePercent(s.effectiveGovernment(), s.buildingsFor(0),
		s.PlayerColonies[0].UnassimilatedPop > 0, achievementMoralePercent(s.Player, s.effectiveGovernment()))
	s.PlayerColonies[0].MoralePercent = 123 // 故意放入 stale derived value，驗 callback 真的重算

	s.advanceLeaderLimbo()
	if len(s.Leaders) != 1 || s.Leaders[0].RawETA != 0 {
		t.Fatalf("ETA=1 應歸零但保留 active leader,got %+v", s.Leaders)
	}
	if s.ColonyLeaderNames[0] != leader.Name {
		t.Fatalf("callback 不應把 ETA=0 誤當解雇,got %q", s.ColonyLeaderNames[0])
	}
	if got := s.PlayerColonies[0].MoralePercent; got != wantMorale {
		t.Fatalf("Colony_Calculation 近似 callback 應重算士氣=%d,got %d", wantMorale, got)
	}
}
