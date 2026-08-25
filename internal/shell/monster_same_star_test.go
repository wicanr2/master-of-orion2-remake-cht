package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestMonsterGroupsAtStarPreservesKindsAndNames(t *testing.T) {
	s := NewDemoSession()
	const star = 2
	s.Monsters = []MonsterGuard{
		{StarIndex: star, Kind: gamedata.MonsterAmoeba, Structure: 50},
		{StarIndex: star, Kind: gamedata.MonsterCrystal, Structure: 80},
		{StarIndex: star, Kind: gamedata.MonsterDragon, Structure: 80, TransitETA: 2},
	}
	groups := s.MonsterGroupsAtStar(star)
	if len(groups) != 2 || groups[0].Kind != gamedata.MonsterAmoeba || groups[1].Kind != gamedata.MonsterCrystal {
		t.Fatalf("同星停泊群組列舉錯誤：%+v", groups)
	}
	want := gamedata.MonsterNameZH(gamedata.MonsterAmoeba) + "、" + gamedata.MonsterNameZH(gamedata.MonsterCrystal)
	if got := s.MonsterNameAtStar(star); got != want {
		t.Fatalf("同星名稱未完整顯示：got=%q want=%q", got, want)
	}
}

func TestRemoveMonsterAtOnlyRemovesFirstParkedGroup(t *testing.T) {
	s := NewDemoSession()
	const star = 2
	s.Monsters = []MonsterGuard{
		{StarIndex: star, Kind: gamedata.MonsterDragon, Structure: 80, TransitETA: 2},
		{StarIndex: star, Kind: gamedata.MonsterAmoeba, Structure: 50},
		{StarIndex: star, Kind: gamedata.MonsterCrystal, Structure: 80},
	}
	s.removeMonsterAt(star)
	if len(s.Monsters) != 2 || s.Monsters[0].Kind != gamedata.MonsterDragon || s.Monsters[0].TransitETA != 2 {
		t.Fatalf("不得誤刪同目的星的航行中 record：%+v", s.Monsters)
	}
	if got := s.MonsterAtStar(star); got == nil || got.Kind != gamedata.MonsterCrystal {
		t.Fatalf("清除第一群後第二群應繼續守衛：%+v", got)
	}
	if !s.StarGuardedByMonster(star) {
		t.Fatal("仍有第二群時星系不得解除守衛 gate")
	}
}

func TestSameStarMonsterGroupsSurviveSnapshot(t *testing.T) {
	s := NewDemoSession()
	const star = 2
	s.Monsters = []MonsterGuard{
		{StarIndex: star, Kind: gamedata.MonsterAmoeba, Structure: 17, Armor: 23},
		{StarIndex: star, Kind: gamedata.MonsterCrystal, Structure: 31, Armor: 47},
	}
	restored := s.snapshot().restore()
	groups := restored.MonsterGroupsAtStar(star)
	if len(groups) != 2 || groups[0].Structure != 17 || groups[1].Armor != 47 {
		t.Fatalf("同星群組快照往返遺失：%+v", groups)
	}
}
