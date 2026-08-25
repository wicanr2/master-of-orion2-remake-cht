package shell

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func eelStructure(t *testing.T) int {
	t.Helper()
	stats, ok := gamedata.MonsterStatsFor(gamedata.MonsterEel)
	if !ok {
		t.Fatal("缺太空鰻資料")
	}
	return stats.Structure
}

func TestSpaceEelSplitsAtExactAgeThirty(t *testing.T) {
	base := eelStructure(t)
	s := NewDemoSession()
	s.Monsters = []MonsterGuard{{StarIndex: 2, Kind: gamedata.MonsterEel, Structure: base, Count: 1, EelAges: []int{29}}}
	if births := s.advanceSpaceEelSplits(); births != 1 {
		t.Fatalf("age 29→30 應分裂一艘，got %d", births)
	}
	m := s.Monsters[0]
	if m.Count != 2 || m.Structure != base*2 || len(m.EelAges) != 2 || m.EelAges[0] != 0 || m.EelAges[1] != 0 {
		t.Fatalf("分裂後數量／結構／age 錯誤：%+v", m)
	}
}

func TestSpaceEelKeepsIndependentAgesAndNewbornDoesNotAdvance(t *testing.T) {
	base := eelStructure(t)
	s := NewDemoSession()
	s.Monsters = []MonsterGuard{{StarIndex: 2, Kind: gamedata.MonsterEel, Structure: base * 2,
		Count: 2, EelAges: []int{29, 7}}}
	s.advanceSpaceEelSplits()
	if got := s.Monsters[0].EelAges; len(got) != 3 || got[0] != 0 || got[1] != 8 || got[2] != 0 {
		t.Fatalf("逐體 age 或新生同回合時序錯誤：%v", got)
	}
}

func TestSpaceEelGlobalCapFourStillResetsAge(t *testing.T) {
	base := eelStructure(t)
	s := NewDemoSession()
	s.Monsters = []MonsterGuard{
		{StarIndex: 2, Kind: gamedata.MonsterEel, Structure: base * 2, Count: 2, EelAges: []int{29, 1}},
		{StarIndex: 3, Kind: gamedata.MonsterEel, Structure: base * 2, Count: 2, EelAges: []int{4, 5}},
	}
	if births := s.advanceSpaceEelSplits(); births != 0 {
		t.Fatalf("全銀河已有四艘不得再分裂，got births=%d", births)
	}
	if s.Monsters[0].EelAges[0] != 0 || s.Monsters[0].Count != 2 || s.Monsters[0].Structure != base*2 {
		t.Fatalf("達上限仍須把母體 age 歸零且不加結構：%+v", s.Monsters[0])
	}
}

func TestLegacySpaceEelSaveNormalizesAndRoundTripsAges(t *testing.T) {
	base := eelStructure(t)
	s := NewDemoSession()
	s.Monsters = []MonsterGuard{{StarIndex: 2, Kind: gamedata.MonsterEel, Structure: base}}
	s.advanceSpaceEelSplits()
	if s.Monsters[0].Count != 1 || len(s.Monsters[0].EelAges) != 1 || s.Monsters[0].EelAges[0] != 1 {
		t.Fatalf("舊 JSON 單鰻正規化錯誤：%+v", s.Monsters[0])
	}
	r := s.snapshot().restore()
	if r.Monsters[0].Count != 1 || len(r.Monsters[0].EelAges) != 1 || r.Monsters[0].EelAges[0] != 1 {
		t.Fatalf("太空鰻逐體 age 未隨快照往返：%+v", r.Monsters[0])
	}
}

func TestSpaceEelSplitRunsOnceThroughEndTurn(t *testing.T) {
	base := eelStructure(t)
	s := NewDemoSession()
	s.DisableEvents = true
	s.Monsters = []MonsterGuard{{StarIndex: 2, Kind: gamedata.MonsterEel, Structure: base, Count: 1, EelAges: []int{29}}}
	s.EndTurn()
	if s.Monsters[0].Count != 2 || s.Monsters[0].Structure != base*2 {
		t.Fatalf("正常 EndTurn 未推進一次分裂：%+v", s.Monsters[0])
	}
	if !strings.Contains(s.LastEvent, "太空鰻完成分裂") {
		t.Fatalf("玩家看不到分裂回報：%q", s.LastEvent)
	}
}
