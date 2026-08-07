package shell

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// monster_test.go:守衛星系的太空怪獸。
// 種類清單有硬證(執行檔字串表)、傷害有手冊(p.114)、生成規則有手冊(p.60);
// 結構值與挑選機率是 remake 估值,測的是**行為與界限**不是「數字等於原版」。

// newMonsterTestSession 造一個「艦隊停在被怪獸把守的無主星」的 session。
func newMonsterTestSession(t *testing.T, kind gamedata.SpaceMonster) (*GameSession, int) {
	t.Helper()
	s := NewDemoSession()
	target := -1
	for i := range s.Stars {
		if s.Stars[i].Owner == 0 && i < len(s.Planets) && !s.StarGuardedByMonster(i) {
			target = i
			break
		}
	}
	if target < 0 {
		t.Fatal("找不到無主且無怪獸的星")
	}
	s.Planets[target] = Planet{
		Name: "測試星 I", Gen: planetGenVersion, TypeID: gamedata.HABITABLE,
		ClimateID: gamedata.TERRAN, GravityID: gamedata.NORMAL_G,
		MineralID: gamedata.ABUNDANT, SizeID: gamedata.MEDIUM_PLANET,
	}
	st, _ := gamedata.MonsterStatsFor(kind)
	s.Monsters = append(s.Monsters, MonsterGuard{StarIndex: target, Kind: kind, Structure: st.Structure})
	s.Fleet().AtStar, s.Fleet().ETA = target, 0
	return s, target
}

// 手冊 p.62:殖民船要「as long as all space monsters … have been cleared from that planet's
// system」。這條 gate 先前只寫在檔頭引文裡,沒有東西可擋。
func TestMonsterBlocksColonizeAndOutpost(t *testing.T) {
	s, target := newMonsterTestSession(t, gamedata.MonsterCrystal)
	s.Fleet().Ships = append(s.Fleet().Ships, Ship{Class: ColonyShipClass}, Ship{Class: OutpostShipClass})

	res := s.ColonizeStar(target)
	if res.Ok {
		t.Error("怪獸盤據時不該能拓殖")
	}
	if !strings.Contains(res.Reason, gamedata.MonsterNameZH(gamedata.MonsterCrystal)) {
		t.Errorf("拒絕理由應點名怪獸,實為 %q", res.Reason)
	}
	if op := s.BuildOutpost(target); op.Ok {
		t.Error("怪獸盤據時不該能建前哨站")
	}
	// 被擋下時不該消耗船。
	if !s.FleetHasColonyShip() || !s.FleetHasOutpostShip() {
		t.Error("被怪獸擋下卻消耗了船")
	}
}

// 清除怪獸之後就能拓殖了(gate 是可解除的,不是死路)。
func TestClearingMonsterUnblocksColonize(t *testing.T) {
	s, target := newMonsterTestSession(t, gamedata.MonsterAmoeba)
	// 給一支足以打贏的艦隊(變形蟲結構 60)。
	s.Fleet().Ships = []Ship{
		{Name: "戰艦一", Class: "戰艦", Weapon: "死光", Armor: "無裝甲", Shield: "無護盾", Special: "無"},
		{Name: "戰艦二", Class: "戰艦", Weapon: "死光", Armor: "無裝甲", Shield: "無護盾", Special: "無"},
		{Name: "戰艦三", Class: "戰艦", Weapon: "死光", Armor: "無裝甲", Shield: "無護盾", Special: "無"},
	}
	res := s.AttackMonster(target)
	if !res.Ok {
		t.Fatalf("挑戰失敗:%s", res.Reason)
	}
	if !res.Won {
		t.Skipf("這次沒打贏(剩餘 %d),戰力/亂數組合下可接受;本測試只在打贏時繼續", res.Remaining)
	}
	if s.StarGuardedByMonster(target) {
		t.Fatal("打贏了怪獸卻還在")
	}
	// 清場之後才把殖民船送進來——這正是手冊 p.62 描述的順序。
	s.Fleet().Ships = append(s.Fleet().Ships, Ship{Class: ColonyShipClass})
	if c := s.ColonizeStar(target); !c.Ok {
		t.Errorf("清除怪獸後應可拓殖,卻回 %q", c.Reason)
	}
}

// 打不贏時怪獸要留著剩餘結構,下次接續(不是每次都回滿血)。
func TestMonsterKeepsDamageBetweenAttacks(t *testing.T) {
	s, target := newMonsterTestSession(t, gamedata.MonsterGuardian) // 結構 300,一次打不完
	s.Fleet().Ships = []Ship{{Name: "偵察", Class: "偵察艦", Weapon: "無武裝", Armor: "無裝甲", Shield: "無護盾", Special: "無"}}

	before := s.MonsterAtStar(target).Structure
	res := s.AttackMonster(target)
	if !res.Ok {
		t.Fatalf("挑戰失敗:%s", res.Reason)
	}
	m := s.MonsterAtStar(target)
	if m == nil {
		t.Skip("偵察艦竟然打死了守護者——戰力表若調整過,這條測試的前提要重設")
	}
	if m.Structure >= before {
		t.Errorf("怪獸結構應下降:%d → %d", before, m.Structure)
	}
	if res.Damage != before-m.Structure {
		t.Errorf("報告傷害 %d 與實際 %d 不符", res.Damage, before-m.Structure)
	}
}

func TestAttackMonsterPreconditions(t *testing.T) {
	s, target := newMonsterTestSession(t, gamedata.MonsterHydra)
	// 艦隊不在該星。
	s.Fleet().AtStar = (target + 1) % len(s.Stars)
	if res := s.AttackMonster(target); res.Ok {
		t.Error("艦隊不在該星不該能挑戰")
	}
	// 沒有怪獸的星。
	s2 := NewDemoSession()
	if res := s2.AttackMonster(s2.Fleet().AtStar); res.Ok {
		t.Error("沒有怪獸的星不該能挑戰")
	}
}

// 生成規則:手冊 p.60「a system with a monster will always have another special」。
func TestGenMonstersAlwaysGiveStarASpecial(t *testing.T) {
	for seed := int64(0); seed < 25; seed++ {
		galaxy, aiHomes := genGalaxy(24, seed, 3, galaxyAgeSetting)
		homes := demoHomeStarSet(aiHomes)
		planets := genPlanets(galaxy, rand.New(rand.NewSource(seed+1)), rand.New(rand.NewSource(seed+5)), galaxyAgeSetting, homes)
		monsters := genMonsters(galaxy, planets, rand.New(rand.NewSource(seed+2)), homes)
		if len(monsters) == 0 {
			t.Fatalf("seed %d:24 星的星圖應該要有怪獸(密度 %d)", seed, gamedata.DefaultGuardMonsterCount)
		}
		seen := map[int]bool{}
		for _, m := range monsters {
			if homes[m.StarIndex] {
				t.Errorf("seed %d:怪獸擺在母星 %d 上", seed, m.StarIndex)
			}
			if seen[m.StarIndex] {
				t.Errorf("seed %d:星 %d 有兩隻怪獸", seed, m.StarIndex)
			}
			seen[m.StarIndex] = true
			// ⚠ 不能寫 `planets[m.StarIndex]` —— Planets 不再與 Stars 平行。
			pi := representativePlanet(galaxy, planets, m.StarIndex)
			if pi < 0 {
				t.Errorf("seed %d:星系 %d 挑不到代表行星", seed, m.StarIndex)
				continue
			}
			if planets[pi].NoPlanet {
				t.Errorf("seed %d:怪獸擺在沒有行星的星系 %d", seed, m.StarIndex)
			}
			// 手冊 p.60 的規則。
			if planets[pi].SpecialID == gamedata.NoSpecial {
				t.Errorf("seed %d:有怪獸的星系 %d 沒有特殊物產(手冊 p.60 說一定有)", seed, m.StarIndex)
			}
			if m.Structure <= 0 {
				t.Errorf("seed %d:怪獸結構 %d 應 > 0", seed, m.Structure)
			}
		}
	}
}

// 存檔往返要保住怪獸(否則讀檔後怪獸消失,星圖難度整個變了)。
func TestMonstersSurviveSaveLoad(t *testing.T) {
	s, target := newMonsterTestSession(t, gamedata.MonsterDragon)
	s.MonsterAtStar(target).Structure = 42 // 打過一半的狀態也要留住
	restored := s.snapshot().restore()
	m := restored.MonsterAtStar(target)
	if m == nil {
		t.Fatal("讀檔後怪獸不見了")
	}
	if m.Structure != 42 {
		t.Errorf("讀檔後剩餘結構 %d,want 42", m.Structure)
	}
}

// 五種怪獸的資料表:名字有硬證(執行檔字串),傷害有手冊(p.114)。
func TestMonsterStatsTable(t *testing.T) {
	all := []gamedata.SpaceMonster{
		gamedata.MonsterGuardian, gamedata.MonsterAmoeba, gamedata.MonsterDragon,
		gamedata.MonsterHydra, gamedata.MonsterCrystal,
	}
	for _, m := range all {
		st, ok := gamedata.MonsterStatsFor(m)
		if !ok {
			t.Fatalf("怪獸 %d 查不到資料", m)
		}
		if st.NameZH == "" || st.NameEN == "" {
			t.Errorf("怪獸 %d 缺名字", m)
		}
		if st.DamageMin <= 0 || st.DamageMax < st.DamageMin {
			t.Errorf("%s 傷害範圍不合理:%d-%d", st.NameEN, st.DamageMin, st.DamageMax)
		}
		if st.Structure <= 0 {
			t.Errorf("%s 結構值 %d 應 > 0", st.NameEN, st.Structure)
		}
	}
	// 手冊 p.114 逐字的兩個數字,直接釘住。
	if st, _ := gamedata.MonsterStatsFor(gamedata.MonsterCrystal); st.DamageMin != 40 || st.DamageMax != 80 {
		t.Errorf("水晶射線應為 40-80(手冊 p.114),實為 %d-%d", st.DamageMin, st.DamageMax)
	}
	// 必中:海德拉的電漿吐息與巨龍的龍焰(手冊「always strikes」/「always hits」)。
	for _, m := range []gamedata.SpaceMonster{gamedata.MonsterHydra, gamedata.MonsterDragon} {
		if st, _ := gamedata.MonsterStatsFor(m); !st.AlwaysHits {
			t.Errorf("%s 的攻擊在手冊裡是必中的", st.NameEN)
		}
	}
	// 守衛星系的清單不含守護者(它只守獵戶座)。
	for _, m := range gamedata.GuardStarMonsters {
		if m == gamedata.MonsterGuardian {
			t.Error("守護者不該出現在一般星系的守衛清單裡")
		}
	}
}
