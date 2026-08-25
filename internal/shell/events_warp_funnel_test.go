package shell

import (
	"encoding/json"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func warpFunnelEvent(t *testing.T) gamedata.RandomEvent {
	t.Helper()
	ev := gamedata.RandomEventByID(27)
	if ev == nil || !ev.Implemented {
		t.Fatal("事件 27 應已登記為 implemented")
	}
	return *ev
}

func TestWarpFunnelRequiresShipAndDoesNotFreezeFleet(t *testing.T) {
	s := NewDemoSession()
	s.Fleets = []Fleet{{AtStar: 0, DestStar: 1, ETA: 7}}
	if _, ok := s.applyRandomEvent(warpFunnelEvent(t)); ok {
		t.Fatal("沒有艦艇時曲速漏斗應失敗")
	}
	s.Fleets[0].Ships = []Ship{{Name: "證據號"}}
	before := s.Fleets[0]
	if _, ok := s.applyRandomEvent(warpFunnelEvent(t)); !ok {
		t.Fatal("有艦艇時曲速漏斗應建立")
	}
	if s.Fleets[0].ETA != before.ETA || s.Fleets[0].DestStar != before.DestStar || len(s.Fleets[0].Ships) != 1 {
		t.Fatalf("1.31 報告型事件不得改航行或船數：before=%+v after=%+v", before, s.Fleets[0])
	}
	if len(s.PersistentEvents) != 1 || s.PersistentEvents[0].Kind != PersistentWarpFunnel || s.PersistentEvents[0].Turns != -1 {
		t.Fatalf("persistent record 不符：%+v", s.PersistentEvents)
	}
}

func TestWarpFunnelLifecycleAndForcedRelease(t *testing.T) {
	s := NewDemoSession()
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentWarpFunnel, StarIndex: -1, Turns: -1}}
	for turn := 0; turn < 5; turn++ {
		if msgs := s.advancePersistentEvents(); len(msgs) != 0 || len(s.PersistentEvents) != 1 {
			t.Fatalf("active turn %d 不應解除：msgs=%v events=%v", turn+1, msgs, s.PersistentEvents)
		}
	}
	s.PersistentEvents[0].Turns = 20
	msgs := s.advancePersistentEvents()
	if len(msgs) != 1 || len(s.PersistentEvents) != 0 {
		t.Fatalf("age 21 檢查應強制解除：msgs=%v events=%v", msgs, s.PersistentEvents)
	}
}

func TestWarpFunnelAIAndJSONRoundTrip(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers = []AIOpponent{{Name: "AI 測試", FleetStrength: 25, FleetETA: 6, FleetDestStar: 2}}
	before := s.AIPlayers[0]
	result, ok := s.applyRandomEventLocalizedToAI(warpFunnelEvent(t), 0)
	if !ok || result.Message == "" {
		t.Fatalf("AI 曲速漏斗應成功：result=%+v ok=%v", result, ok)
	}
	if s.AIPlayers[0].FleetETA != before.FleetETA || s.AIPlayers[0].FleetDestStar != before.FleetDestStar || s.AIPlayers[0].FleetStrength != before.FleetStrength {
		t.Fatalf("AI 航行／艦力不得改變：before=%+v after=%+v", before, s.AIPlayers[0])
	}
	b, err := json.Marshal(s.PersistentEvents)
	if err != nil {
		t.Fatal(err)
	}
	var got []PersistentEvent
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Kind != PersistentWarpFunnel || got[0].Turns != -1 {
		t.Fatalf("JSON 往返遺失 record：%+v", got)
	}
}

func TestWarpFunnelInactiveHotseatSeat(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) < 1 || s.SetupHotseat(2) != 2 {
		t.Skip("需要可接管的第二帝國")
	}
	s.Seats[1].Fleets = []Fleet{{Ships: []Ship{{Name: "第二席艦"}}, AtStar: 0, DestStar: 1, ETA: 4}}
	before := s.Seats[1].Fleets[0]
	result, ok := s.applyRandomEventLocalizedToTarget(warpFunnelEvent(t), eventEmpireTarget{
		kind: eventEmpireSeat, index: 1, alive: true,
	})
	if !ok || result.Message == "" {
		t.Fatalf("非目前熱座席位應可建立曲速漏斗：result=%+v ok=%v", result, ok)
	}
	if s.ActiveSeat != 0 {
		t.Fatalf("事件結算後應恢復第 0 席，got %d", s.ActiveSeat)
	}
	after := s.Seats[1].Fleets[0]
	if after.ETA != before.ETA || after.DestStar != before.DestStar || len(after.Ships) != len(before.Ships) {
		t.Fatalf("第二席艦隊不得被凍結或改寫：before=%+v after=%+v", before, after)
	}
	if len(s.PersistentEvents) != 1 || s.PersistentEvents[0].Kind != PersistentWarpFunnel {
		t.Fatalf("全局 persistent record 遺失：%+v", s.PersistentEvents)
	}
}
