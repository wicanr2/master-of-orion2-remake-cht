package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func clearPlayerMonsterAttackTargets(s *GameSession, colony int) {
	if colony < len(s.ColonyBuildings) {
		s.ColonyBuildings[colony] = map[string]bool{}
	}
	if colony < len(s.PlayerColonyMarines) {
		s.PlayerColonyMarines[colony] = 0
	}
	if colony < len(s.PlayerColonyTanks) {
		s.PlayerColonyTanks[colony] = 0
	}
	s.PlayerColonies[colony].BombardmentBuildProgress = 0
}

func TestMonsterStrategicBombardmentUsesThreeRoundsAndDivideForty(t *testing.T) {
	s := NewDemoSession()
	m := &MonsterGuard{Kind: gamedata.MonsterAmoeba, Count: 1, Structure: 60}
	hits := s.monsterStrategicBombardHits(m, 0)
	// 變形蟲每輪 25..50，三輪總和 75..150，原版 /40 向下取整。
	if hits < 1 || hits > 3 {
		t.Fatalf("固定三輪與 /40 邊界錯誤：hits=%d", hits)
	}

	plain, shielded := NewDemoSession(), NewDemoSession()
	plain.EventSeed, shielded.EventSeed = 991, 991
	plainHits := plain.monsterStrategicBombardHits(m, 0)
	shieldHits := shielded.monsterStrategicBombardHits(m, 20)
	if shieldHits > plainHits {
		t.Fatalf("逐發行星護盾不得增加傷害：plain=%d shield=%d", plainHits, shieldHits)
	}
}

func TestSpaceEelNeverBombardsColony(t *testing.T) {
	s := NewDemoSession()
	m := &MonsterGuard{StarIndex: s.PlayerColonyStarIndex(0), Kind: gamedata.MonsterEel,
		Structure: eelStructure(t), Count: 1, EelAges: []int{0}}
	before := s.PlayerColonies[0].Population
	if impact, attacked := s.resolveEventMonsterColonyAttack(m); attacked || impact.Hits != 0 {
		t.Fatalf("sub_E8029 對 type 13 應立即排除：attacked=%v impact=%+v", attacked, impact)
	}
	if s.PlayerColonies[0].Population != before {
		t.Fatal("太空鰻不得傷害殖民地")
	}
}

func TestArrivingMonsterBombardsPlayerColonyAndStays(t *testing.T) {
	s := NewDemoSession()
	clearPlayerMonsterAttackTargets(s, 0)
	s.PlayerColonies[0].Population = 5
	m := &MonsterGuard{StarIndex: s.PlayerColonyStarIndex(0), Kind: gamedata.MonsterAmoeba, Structure: 60}
	impact, attacked := s.resolveEventMonsterColonyAttack(m)
	if !attacked || impact.Hits < 1 || impact.PopulationLost < 1 || impact.MonsterDestroyed {
		t.Fatalf("無固定防禦時應三輪轟炸且怪物留在星系：attacked=%v impact=%+v", attacked, impact)
	}
	if s.PlayerColonies[0].Population >= 5 || m.Structure <= 0 {
		t.Fatalf("殖民地／怪物回寫錯誤：population=%d monster=%+v", s.PlayerColonies[0].Population, *m)
	}
}

func TestFixedDefenseCanDestroyMonsterBeforeBombardment(t *testing.T) {
	s := NewDemoSession()
	buildings := map[string]bool{gamedata.StellarConverterName: true}
	m := &MonsterGuard{StarIndex: s.PlayerColonyStarIndex(0), Kind: gamedata.MonsterAmoeba, Structure: 1}
	if !s.monsterDefenseBattle(m, buildings, s.Player) || m.Structure != 0 {
		t.Fatalf("固定防禦應能在轟炸前擊毀低結構怪物：%+v", *m)
	}
}

func TestMonsterAttackWritesAIColony(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers) == 0 || len(s.AIPlayers[0].Colonies) == 0 {
		t.Skip("需要 AI 殖民地")
	}
	a := &s.AIPlayers[0]
	ensureAIGroundForceSlots(a)
	a.ColonyBuildings[0] = map[string]bool{}
	a.ColonyMarines[0], a.ColonyTanks[0] = 0, 0
	a.Colonies[0].Population, a.Colonies[0].BombardmentBuildProgress = 5, 0
	m := &MonsterGuard{StarIndex: a.ColonyStars[0], Kind: gamedata.MonsterCrystal, Structure: 90}
	impact, attacked := s.resolveEventMonsterColonyAttack(m)
	if !attacked || impact.Hits < 1 || a.Colonies[0].Population >= 5 {
		t.Fatalf("AI 殖民地未走共用戰略傷亡池：attacked=%v impact=%+v pop=%d", attacked, impact, a.Colonies[0].Population)
	}
}

func TestMonsterAttackWritesInactiveHotseatSeat(t *testing.T) {
	s := NewDemoSession()
	if s.SetupHotseat(2) != 2 || len(s.Seats) < 2 || len(s.Seats[1].PlayerColonies) == 0 {
		t.Skip("需要第二熱座席位")
	}
	seat := &s.Seats[1]
	seat.ColonyBuildings[0] = map[string]bool{}
	seat.PlayerColonyMarines[0], seat.PlayerColonyTanks[0] = 0, 0
	seat.PlayerColonies[0].Population, seat.PlayerColonies[0].BombardmentBuildProgress = 5, 0
	star := seat.PlayerColonyStars[0]
	m := &MonsterGuard{StarIndex: star, Kind: gamedata.MonsterHydra, Structure: 80}
	impact, attacked := s.resolveEventMonsterColonyAttack(m)
	if !attacked || impact.Hits < 1 || s.Seats[1].PlayerColonies[0].Population >= 5 {
		t.Fatalf("非目前熱座席位未正確寫回：attacked=%v impact=%+v pop=%d",
			attacked, impact, s.Seats[1].PlayerColonies[0].Population)
	}
}

func TestMonsterGroupCanDestroyAndRemoveColony(t *testing.T) {
	s := NewDemoSession()
	clearPlayerMonsterAttackTargets(s, 0)
	star := s.PlayerColonyStarIndex(0)
	s.PlayerColonies[0].Population = 1
	s.PlayerColonies[0].BombardmentLastPopulationPoints = 100
	before := len(s.PlayerColonies)
	// 原版 sub_4267B 會收集同星同 type 的全部 ship；Count 50 表示同星聚合群。
	m := &MonsterGuard{StarIndex: star, Kind: gamedata.MonsterDragon, Structure: 120, Count: 50}
	impact, attacked := s.resolveEventMonsterColonyAttack(m)
	if !attacked || !impact.ColonyDestroyed || len(s.PlayerColonies) != before-1 {
		t.Fatalf("人口耗盡後必須移除殖民地：attacked=%v impact=%+v colonies=%d→%d",
			attacked, impact, before, len(s.PlayerColonies))
	}
}
