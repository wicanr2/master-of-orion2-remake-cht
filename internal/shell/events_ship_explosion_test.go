package shell

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestShipExplosionUsesOriginalReservoirSamplingAndOnlyRemovesOneShip(t *testing.T) {
	ships := []Ship{{Name: "甲", Class: "巡防艦", Damage: 1}, {Name: "乙", Class: "巡洋艦", Damage: 2},
		{Name: "丙", Class: "戰艦", Damage: 3}, {Name: "丁", Class: "泰坦", Damage: 4}}
	const seed = 73
	wantIndex := -1
	wantRand := newRandStream(seed)
	for i := range ships {
		if wantRand.Intn(i+1) == 0 {
			wantIndex = i
		}
	}
	s := NewDemoSession()
	s.Fleets = []Fleet{{Ships: append([]Ship(nil), ships...)}}
	s.eventRand = newRandStream(seed)
	result, ok := s.resolvePlayerShipExplosion()
	if !ok || result.Lost.Name != ships[wantIndex].Name || s.ShipCount() != len(ships)-1 {
		t.Fatalf("reservoir 選艦／單艦移除不符：result=%+v ships=%d want=%s", result, s.ShipCount(), ships[wantIndex].Name)
	}
	for _, survivor := range s.AllShips() {
		for _, original := range ships {
			if survivor.Name == original.Name && survivor.Damage != original.Damage {
				t.Fatalf("事件不得對倖存艦套連鎖傷害：got=%+v want=%+v", survivor, original)
			}
		}
	}
}

func TestShipExplosionCanDestroyOnlyShipAndKillsAssignedOfficer(t *testing.T) {
	s := NewDemoSession()
	s.Leaders = []Leader{{ID: 41, Name: "艦長", Ship: true, RawStatus: 1}, {ID: 42, Name: "留任者", Ship: true}}
	s.Fleets = []Fleet{{Ships: []Ship{{Name: "唯一艦", Class: "巡防艦", OfficerID: 41, OfficerName: "艦長"}}}}
	s.Player.OfficerMaintenance = eventLeaderUpkeepTotal(s.Leaders)
	s.Player.UsedCommandPoints = 99
	result, ok := s.resolvePlayerShipExplosion()
	if !ok || s.ShipCount() != 0 || result.OfficerName != "艦長" {
		t.Fatalf("唯一艦及軍官應一併移除：result=%+v ships=%d", result, s.ShipCount())
	}
	if len(s.Leaders) != 1 || s.Leaders[0].Name != "留任者" {
		t.Fatalf("死亡軍官不得回人才庫，其他領袖須保留：%+v", s.Leaders)
	}
	if s.Player.OfficerMaintenance != eventLeaderUpkeepTotal(s.Leaders) || s.Player.UsedCommandPoints != 0 {
		t.Fatalf("軍官維護與指揮點衍生值應立即刷新：maintenance=%d command=%d",
			s.Player.OfficerMaintenance, s.Player.UsedCommandPoints)
	}
	message, ok := s.applyRandomEventLocalized(*gamedata.RandomEventByID(8))
	if ok || message.Message != "" {
		t.Fatal("無艦艇時事件應不成立")
	}
}

func TestShipExplosionWritesAIAndInactiveHotseatTargets(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers[0].Ships = []Ship{{Name: "AI 唯一艦", OfficerID: 51, OfficerName: "AI 艦長"}}
	s.AIPlayers[0].Leaders = []Leader{{ID: 51, Name: "AI 艦長", Ship: true}}
	result, ok := s.applyRandomEventLocalizedToTarget(*gamedata.RandomEventByID(8),
		eventEmpireTarget{kind: eventEmpireAI, index: 0})
	if !ok || len(s.AIPlayers[0].Ships) != 0 || len(s.AIPlayers[0].Leaders) != 0 ||
		!strings.Contains(result.Message, "AI 艦長") {
		t.Fatalf("AI 艦艇／軍官回寫不符：result=%+v ai=%+v", result, s.AIPlayers[0])
	}

	hotseat := NewDemoSession()
	if hotseat.SetupHotseatWithAIIndices([]int{0}) != 2 {
		t.Fatal("測試需要兩個熱座席位")
	}
	hotseat.Seats[1].Fleets = []Fleet{{Ships: []Ship{{Name: "第二席唯一艦"}}}}
	if _, ok := hotseat.applyRandomEventLocalizedToTarget(*gamedata.RandomEventByID(8),
		eventEmpireTarget{kind: eventEmpireSeat, index: 1}); !ok {
		t.Fatal("非目前熱座席位事件應成立")
	}
	if len(hotseat.Seats[1].Fleets[0].Ships) != 0 || hotseat.ActiveSeat != 0 {
		t.Fatalf("事件應回寫第二席並恢復目前席：ships=%v active=%d", hotseat.Seats[1].Fleets[0].Ships, hotseat.ActiveSeat)
	}
}
