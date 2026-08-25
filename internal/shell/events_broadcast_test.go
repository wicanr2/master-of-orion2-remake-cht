package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestOriginalExpansionBroadcastBoundaries(t *testing.T) {
	for _, c := range []struct {
		stars, stage, colonies int
		council                bool
		want                   bool
	}{
		{40, 0, 9, false, false}, {40, 0, 10, false, true},
		{40, 1, 14, false, false}, {40, 1, 15, false, true},
		{40, 2, 17, false, false}, {40, 2, 18, false, true},
		{40, 2, 18, true, false}, {40, 2, 20, true, true},
		{40, 3, 40, false, false},
	} {
		if got := originalExpansionBroadcastDue(c.stars, c.stage, c.colonies, c.council); got != c.want {
			t.Errorf("stars=%d stage=%d colonies=%d council=%v: got %v want %v",
				c.stars, c.stage, c.colonies, c.council, got, c.want)
		}
	}
}

func TestExpansionBroadcastCrossesOriginalThreshold(t *testing.T) {
	s := NewDemoSession()
	s.Stars = make([]Star, 40)
	s.PlayerColonies = make([]engine.ColonyState, 10)
	s.PlayerColonyStars = make([]int, 10)
	for i := range s.PlayerColonies {
		s.PlayerColonies[i] = engine.ColonyState{Population: 1}
		s.PlayerColonyStars[i] = i
	}
	for i := range s.AIPlayers {
		s.AIPlayers[i].Colonies = nil
		s.AIPlayers[i].ColonyStars = nil
	}
	s.advanceStatusGrowthAndRanking()
	if s.StatusBroadcast.GrowthStage != 1 || len(s.StatusBroadcast.Queue) == 0 ||
		s.StatusBroadcast.Queue[0].EventID != 30 ||
		s.StatusBroadcast.Queue[0].TargetKind != eventEmpirePlayer.String() {
		t.Fatalf("跨過第一階段門檻應建立玩家擴張新聞：stage=%d queue=%+v",
			s.StatusBroadcast.GrowthStage, s.StatusBroadcast.Queue)
	}
}

func TestEmpireEliminationBroadcastUsesAliveTransition(t *testing.T) {
	s := NewDemoSession()
	s.ensureStatusEmpireBaseline()
	if len(s.AIPlayers) == 0 {
		t.Fatal("需要 AI 帝國")
	}
	s.AIPlayers[0].Colonies = nil
	s.AIPlayers[0].ColonyStars = nil
	s.detectEmpireEliminationBroadcasts()
	if len(s.StatusBroadcast.Queue) != 1 || s.StatusBroadcast.Queue[0].EventID != 29 ||
		s.StatusBroadcast.Queue[0].TargetKind != eventEmpireAI.String() || s.StatusBroadcast.Queue[0].TargetIndex != 0 {
		t.Fatalf("帝國滅亡新聞錯誤：%+v", s.StatusBroadcast.Queue)
	}
	s.detectEmpireEliminationBroadcasts()
	if len(s.StatusBroadcast.Queue) != 1 {
		t.Fatal("同一個已滅亡帝國不得每回合重播")
	}
}

func TestStatusBroadcastQueueSurvivesSnapshotAndPreservesOrder(t *testing.T) {
	s := NewDemoSession()
	s.StatusBroadcast = StatusBroadcastState{GrowthStage: 2, EmpireAlive: []bool{true, false}, OrionStars: []int{7},
		Queue: []EventReport{{EventID: 29, Message: "第一則"}, {EventID: 35, Message: "第二則"}}}
	r := s.snapshot().restore()
	if r.StatusBroadcast.GrowthStage != 2 || len(r.StatusBroadcast.Queue) != 2 ||
		r.StatusBroadcast.Queue[0].EventID != 29 || r.StatusBroadcast.Queue[1].EventID != 35 ||
		len(r.StatusBroadcast.OrionStars) != 1 || r.StatusBroadcast.OrionStars[0] != 7 {
		t.Fatalf("狀態播報存檔往返失真：%+v", r.StatusBroadcast)
	}
	r.publishNextStatusBroadcast()
	if r.LastEventReport == nil || r.LastEventReport.EventID != 29 || len(r.StatusBroadcast.Queue) != 1 {
		t.Fatalf("第一則應先播且第二則保留：report=%+v queue=%+v", r.LastEventReport, r.StatusBroadcast.Queue)
	}
}

func TestOrionDiscoveryBroadcastComesFromFleetArrival(t *testing.T) {
	s := NewDemoSession()
	star := 4
	s.Monsters = append(s.Monsters, MonsterGuard{StarIndex: star, Kind: gamedata.MonsterGuardian, Structure: 300})
	s.Fleet().DestStar, s.Fleet().ETA = star, 1
	s.advanceFleet()
	if len(s.StatusBroadcast.Queue) != 1 || s.StatusBroadcast.Queue[0].EventID != 32 ||
		s.StatusBroadcast.Queue[0].TargetKind != eventEmpirePlayer.String() {
		t.Fatalf("Orion 發現新聞錯誤：%+v", s.StatusBroadcast.Queue)
	}
	// 同一星的第二支艦隊抵達不得重播。
	s.Fleets = append(s.Fleets, Fleet{AtStar: 0, DestStar: star, ETA: 1, Ships: []Ship{{Name: "後續艦"}}})
	s.advanceFleet()
	if len(s.StatusBroadcast.Queue) != 1 {
		t.Fatal("Orion 發現是全局一次性新聞")
	}
}

func TestAntaranAndRebellionBroadcastHooks(t *testing.T) {
	s := NewDemoSession()
	s.AntaranHomeworldConquered = true
	s.advanceAntaranVictory()
	if len(s.StatusBroadcast.Queue) != 1 || s.StatusBroadcast.Queue[0].EventID != 33 {
		t.Fatalf("安塔蘭勝利新聞錯誤：%+v", s.StatusBroadcast.Queue)
	}
	s.advanceAntaranVictory()
	if len(s.StatusBroadcast.Queue) != 1 {
		t.Fatal("安塔蘭勝利不得重播")
	}
	s.queueRebellionBroadcasts([]RebellionResult{{ColonyName: "叛亂星", ColonyLost: true, RevertedToAI: 0}})
	if len(s.StatusBroadcast.Queue) != 2 || s.StatusBroadcast.Queue[1].EventID != 35 ||
		s.StatusBroadcast.Queue[1].TargetIndex != 0 {
		t.Fatalf("叛亂易手新聞錯誤：%+v", s.StatusBroadcast.Queue)
	}
}

func TestRankingBroadcastOriginalGates(t *testing.T) {
	s := NewDemoSession()
	s.StatusBroadcast.GrowthStage = 3 // 本測試只隔離事件 31。
	s.Turn = 60
	for seed := int64(0); ; seed++ {
		probe := newRandStream(seed)
		if probe.Intn(40)+1 == 1 {
			s.eventRand = newRandStream(seed)
			break
		}
	}
	s.advanceStatusGrowthAndRanking()
	if len(s.StatusBroadcast.Queue) != 1 || s.StatusBroadcast.Queue[0].EventID != 31 {
		t.Fatalf("elapsed>50 且 1/40 命中應建立排行榜：%+v", s.StatusBroadcast.Queue)
	}

	blocked := NewDemoSession()
	blocked.StatusBroadcast.GrowthStage = 3
	blocked.Turn, blocked.CouncilMeetings = 60, 1
	blocked.eventRand = newRandStream(0)
	blocked.advanceStatusGrowthAndRanking()
	if len(blocked.StatusBroadcast.Queue) != 0 {
		t.Fatal("議會已成立後不得建立排行榜播報")
	}
}

func TestStatusBroadcastPublishesToAllHotseatSeats(t *testing.T) {
	s := NewDemoSession()
	if s.SetupHotseat(2) != 2 {
		t.Fatal("需要兩席熱座")
	}
	s.queueStatusBroadcast(EventReport{EventID: 31, Message: "排行榜", MessageEN: "Rankings"})
	s.publishNextStatusBroadcast()
	for i := range s.Seats {
		if s.Seats[i].LastEventReport == nil || s.Seats[i].LastEventReport.EventID != 31 {
			t.Fatalf("席位 %d 未收到全銀河新聞：%+v", i, s.Seats[i].LastEventReport)
		}
	}
}
