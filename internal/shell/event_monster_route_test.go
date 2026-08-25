package shell

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestInvadingMonsterTravelsBeforeGuardingTarget(t *testing.T) {
	s := NewDemoSession()
	target := s.PlayerColonyStarIndex(0)
	message, ok := s.placeInvadingMonster(gamedata.MonsterAmoeba, target)
	if !ok {
		t.Fatal("事件怪物應能建立 owner 8 航行 record")
	}
	m := &s.Monsters[len(s.Monsters)-1]
	if m.TransitETA < 1 || !strings.Contains(message, "正朝") || !strings.Contains(message, "回合") {
		t.Fatalf("出發訊息或 ETA 錯誤：monster=%+v message=%q", *m, message)
	}
	if got := s.MonsterAtStar(target); got != nil {
		t.Fatalf("航行中不得提前盤據目的星：%+v", got)
	}
	if s.StarGuardedByMonster(target) {
		t.Fatal("航行中不得提前阻擋拓殖")
	}
}

func TestEventMonsterRouteAdvancesOnceAndArrives(t *testing.T) {
	s := NewDemoSession()
	target := s.PlayerColonyStarIndex(0)
	s.Monsters = []MonsterGuard{{StarIndex: target, Kind: gamedata.MonsterHydra,
		Structure: 100, TransitETA: 2}}
	if messages := s.advanceEventMonsterRoutes(); len(messages) != 0 || s.Monsters[0].TransitETA != 1 {
		t.Fatalf("ETA 2 應只遞減一次：messages=%v monster=%+v", messages, s.Monsters[0])
	}
	messages := s.advanceEventMonsterRoutes()
	if len(messages) != 1 || messages[0] == "" || s.Monsters[0].TransitETA != 0 {
		t.Fatalf("ETA 1 應抵達並回報：messages=%v monster=%+v", messages, s.Monsters[0])
	}
	if s.MonsterAtStar(target) == nil || !s.StarGuardedByMonster(target) {
		t.Fatal("抵達後才應成為星系守衛")
	}
}

func TestEventMonsterRouteRunsThroughEndTurnAndPersists(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	target := s.PlayerColonyStarIndex(0)
	s.Monsters = []MonsterGuard{{StarIndex: target, Kind: gamedata.MonsterCrystal,
		Structure: 100, TransitETA: 2}}
	restored := s.snapshot().restore()
	if len(restored.Monsters) != 1 || restored.Monsters[0].TransitETA != 2 {
		t.Fatalf("owner 8 航程未隨快照往返：%+v", restored.Monsters)
	}
	restored.EndTurn()
	if restored.Monsters[0].TransitETA != 1 || restored.MonsterAtStar(target) != nil {
		t.Fatalf("正常 EndTurn 應只推進一次且尚未抵達：%+v", restored.Monsters[0])
	}
	restored.EndTurn()
	if restored.Monsters[0].TransitETA != 0 || restored.MonsterAtStar(target) == nil ||
		restored.LastEvent == "" {
		t.Fatalf("第二回合應抵達並顯示訊息：monster=%+v event=%q", restored.Monsters[0], restored.LastEvent)
	}
}

func TestTransitEelsCountTowardCapButDoNotAge(t *testing.T) {
	base := eelStructure(t)
	s := NewDemoSession()
	s.Monsters = []MonsterGuard{
		{StarIndex: 1, Kind: gamedata.MonsterEel, Structure: base, Count: 1, EelAges: []int{29}},
		{StarIndex: 2, Kind: gamedata.MonsterEel, Structure: base * 3, Count: 3,
			EelAges: []int{4, 8, 12}, TransitETA: 3},
	}
	if births := s.advanceSpaceEelSplits(); births != 0 {
		t.Fatalf("航行中太空鰻仍屬 active ship，四艘上限不得分裂：births=%d", births)
	}
	if s.Monsters[0].EelAges[0] != 0 {
		t.Fatalf("停泊母體達 30 即使被 cap 擋住仍須歸零：%v", s.Monsters[0].EelAges)
	}
	if got := s.Monsters[1].EelAges; got[0] != 4 || got[1] != 8 || got[2] != 12 {
		t.Fatalf("航行中的太空鰻不得推進 age：%v", got)
	}
}

func TestArrivingEelAgeStartsAtZero(t *testing.T) {
	base := eelStructure(t)
	s := NewDemoSession()
	s.DisableEvents = true
	s.Monsters = []MonsterGuard{{StarIndex: 2, Kind: gamedata.MonsterEel, Structure: base,
		Count: 1, EelAges: []int{17}, TransitETA: 1}}
	s.EndTurn()
	if s.Monsters[0].TransitETA != 0 || len(s.Monsters[0].EelAges) != 1 || s.Monsters[0].EelAges[0] != 0 {
		t.Fatalf("抵達同回合 age 必須維持 0：%+v", s.Monsters[0])
	}
}
