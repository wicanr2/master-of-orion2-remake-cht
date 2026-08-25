package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestOriginalHyperspaceFluxLifetimeBoundaries(t *testing.T) {
	if originalHyperspaceFluxEnds(4, 1) {
		t.Fatal("raw age 4 即使 roll=1 也不應解除")
	}
	if !originalHyperspaceFluxEnds(5, 1) {
		t.Fatal("raw age 5 起 roll=1 應解除")
	}
	if originalHyperspaceFluxEnds(20, 2) {
		t.Fatal("raw age 20 且骰失敗不應強制解除")
	}
	if !originalHyperspaceFluxEnds(21, 2) {
		t.Fatal("raw age 21 應強制解除")
	}
}

func TestHyperspaceFluxBlocksNewAndExistingPlayerTravel(t *testing.T) {
	s := NewDemoSession()
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentHyperspaceFlux, StarIndex: -1}}
	f := s.Fleet()
	f.AtStar, f.DestStar, f.ETA = 0, 1, 3
	s.advanceFleet()
	if f.ETA != 3 || f.AtStar != 0 || f.DestStar != 1 {
		t.Fatalf("亂流中既有航程被改寫：%+v", *f)
	}
	f.DestStar, f.ETA = -1, 0
	if s.SendFleet(1) {
		t.Fatal("亂流中非跨維度玩家不應能下達新航程")
	}
	if f.DestStar != -1 || f.ETA != 0 {
		t.Fatalf("失敗命令不應改寫艦隊：%+v", *f)
	}
}

func TestTransDimensionalPlayerIgnoresHyperspaceFlux(t *testing.T) {
	s := NewDemoSession()
	s.RaceIndex = -1
	s.CustomRaceTraits = uint32(1) << uint(gamedata.TRAIT_TRANS_DIMENSIONAL)
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentHyperspaceFlux, StarIndex: -1}}
	f := s.Fleet()
	f.AtStar, f.DestStar, f.ETA = 0, 1, 2
	s.advanceFleet()
	if f.ETA != 1 {
		t.Fatalf("跨維度玩家航程應繼續，ETA=%d", f.ETA)
	}
}

func TestHyperspaceFluxFreezesAIExceptTransDimensional(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) < 2 {
		t.Fatal("測試需要至少兩個 AI")
	}
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentHyperspaceFlux, StarIndex: -1}}
	for i := 0; i < 2; i++ {
		a := &s.AIPlayers[i]
		a.FleetPosSet, a.FleetStar, a.FleetDestStar, a.FleetETA = true, 0, 1, 3
	}
	// 崔拉里安是內建跨維度種族；非跨維度對手取人類。
	s.AIPlayers[0].RaceIndex = 0
	for i := range Races {
		if gamedata.OrigRaceHasTrait(Races[i].OrigIdx, gamedata.TRAIT_TRANS_DIMENSIONAL) {
			s.AIPlayers[1].RaceIndex = i
			break
		}
	}
	s.advanceAIFleets()
	if s.AIPlayers[0].FleetETA != 3 {
		t.Fatalf("非跨維度 AI 應凍結，ETA=%d", s.AIPlayers[0].FleetETA)
	}
	if s.AIPlayers[1].FleetETA != 2 {
		t.Fatalf("跨維度 AI 應繼續航行，ETA=%d", s.AIPlayers[1].FleetETA)
	}
}

func TestHyperspaceFluxPersistsAndDoesNotDuplicate(t *testing.T) {
	s := NewDemoSession()
	if _, ok := s.startHyperspaceFlux(); !ok {
		t.Fatal("第一次建立亂流應成功")
	}
	if _, ok := s.startHyperspaceFlux(); ok {
		t.Fatal("active 亂流不得重複建立")
	}
	s.PersistentEvents[0].Turns = 12
	restored := s.snapshot().restore()
	if len(restored.PersistentEvents) != 1 || restored.PersistentEvents[0].Kind != PersistentHyperspaceFlux || restored.PersistentEvents[0].Turns != 12 {
		t.Fatalf("亂流存檔往返失真：%+v", restored.PersistentEvents)
	}
}

func TestHyperspaceFluxFreezesInactiveHotseatFleet(t *testing.T) {
	s := NewDemoSession()
	if s.SetupHotseat(2) != 2 {
		t.Fatal("測試需要兩席熱座")
	}
	s.PersistentEvents = []PersistentEvent{{Kind: PersistentHyperspaceFlux, StarIndex: -1, Turns: 0}}
	s.Seats[1].RaceIndex = 0
	s.Seats[1].CustomRaceTraits = 0
	s.Seats[1].Fleets = []Fleet{{AtStar: 0, DestStar: 1, ETA: 3}}
	s.advanceIdleSeats()
	if got := s.Seats[1].Fleets[0].ETA; got != 3 {
		t.Fatalf("非目前熱座席位的航程也應凍結，ETA=%d", got)
	}
}
