package shell

import (
	"reflect"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func cloneShipsForMutinyTest(ships []Ship) []Ship {
	if ships == nil {
		return nil
	}
	out := make([]Ship, len(ships))
	copy(out, ships)
	return out
}

func TestShipMutinyPlayerReportDoesNotTransferShip(t *testing.T) {
	s := NewDemoSession()
	playerBefore := cloneShipsForMutinyTest(s.Fleet().Ships)
	aiBefore := make([][]Ship, len(s.AIPlayers))
	for i := range s.AIPlayers {
		aiBefore[i] = cloneShipsForMutinyTest(s.AIPlayers[i].Ships)
	}
	result, ok := s.applyRandomEventLocalized(*gamedata.RandomEventByID(13))
	if !ok || result.Message == "" || result.MessageEN == "" {
		t.Fatalf("事件 13 應成功產生雙語通報：ok=%v result=%+v", ok, result)
	}
	if !reflect.DeepEqual(playerBefore, s.Fleet().Ships) {
		t.Fatal("1.31 事件 13 consumer 不得移除或改寫玩家艦艇")
	}
	for i := range s.AIPlayers {
		if !reflect.DeepEqual(aiBefore[i], s.AIPlayers[i].Ships) {
			t.Fatalf("1.31 事件 13 consumer 不得把艦艇移交給 AI %d", i)
		}
	}
}

func TestShipMutinyAIAndHotseatTargetsAreReportsOnly(t *testing.T) {
	ev := *gamedata.RandomEventByID(13)
	s := NewDemoSession()
	aiBefore := cloneShipsForMutinyTest(s.AIPlayers[0].Ships)
	if result, ok := s.applyRandomEventLocalizedToTarget(ev,
		eventEmpireTarget{kind: eventEmpireAI, index: 0}); !ok || result.Message == "" {
		t.Fatalf("AI 目標應成功產生通報：ok=%v result=%+v", ok, result)
	}
	if !reflect.DeepEqual(aiBefore, s.AIPlayers[0].Ships) {
		t.Fatal("AI 目標事件 13 不得改寫 AI 艦艇")
	}

	hotseat := NewDemoSession()
	if hotseat.SetupHotseatWithAIIndices([]int{0}) != 2 {
		t.Fatal("測試需要兩個熱座席位")
	}
	seatBefore := cloneShipsForMutinyTest(hotseat.Seats[1].Fleets[0].Ships)
	if result, ok := hotseat.applyRandomEventLocalizedToTarget(ev,
		eventEmpireTarget{kind: eventEmpireSeat, index: 1}); !ok || result.Message == "" {
		t.Fatalf("非目前熱座席位應成功產生通報：ok=%v result=%+v", ok, result)
	}
	if !reflect.DeepEqual(seatBefore, hotseat.Seats[1].Fleets[0].Ships) {
		t.Fatal("非目前熱座席位的事件 13 不得改寫艦艇")
	}
}
