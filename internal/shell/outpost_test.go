package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// outpost_test.go:軍事前哨站。期望值全部對照 GAME_MANUAL.pdf 逐字(見 outpost.go 檔頭引文)。

// newOutpostTestSession 造一個「艦隊帶著前哨船停在一顆無主的氣態巨星星系」的 session。
func newOutpostTestSession(t *testing.T, tp gamedata.PlanetType) (*GameSession, int) {
	t.Helper()
	s := NewDemoSession()
	target := -1
	for i := range s.Stars {
		if s.Stars[i].Owner == 0 && i < len(s.Planets) {
			target = i
			break
		}
	}
	if target < 0 {
		t.Fatal("找不到無主星")
	}
	s.Planets[target] = Planet{
		Name: "測試星 I", Gen: planetGenVersion, TypeID: tp,
		ClimateID: gamedata.TERRAN, GravityID: gamedata.NORMAL_G,
		MineralID: gamedata.ABUNDANT, SizeID: gamedata.MEDIUM_PLANET,
	}
	s.Ships = append(s.Ships, Ship{Name: "前哨船 1 號", Class: OutpostShipClass})
	s.FleetAtStar, s.FleetETA = target, 0
	return s, target
}

// 手冊 p.85:前哨站不需要宜居世界,氣態巨星與小行星帶正是它的用途。
func TestBuildOutpostOnGasGiantAndAsteroids(t *testing.T) {
	for _, tp := range []gamedata.PlanetType{gamedata.GAS_GIANT, gamedata.ASTEROIDS, gamedata.HABITABLE} {
		s, target := newOutpostTestSession(t, tp)
		nColonies := len(s.PlayerColonies)

		res := s.BuildOutpost(target)
		if !res.Ok {
			t.Fatalf("類別 %d 建前哨站失敗:%s", tp, res.Reason)
		}
		if !s.HasOutpostAt(target) {
			t.Errorf("類別 %d:建完卻查不到前哨站", tp)
		}
		// 手冊「produces nothing」——不能變出一個殖民地。
		if len(s.PlayerColonies) != nColonies {
			t.Errorf("類別 %d:前哨站不該產生殖民地(%d → %d)", tp, nColonies, len(s.PlayerColonies))
		}
		if s.Stars[target].Owner != 1 {
			t.Errorf("類別 %d:該星應歸玩家所有", tp)
		}
		// 前哨船應被消耗。
		if s.FleetHasOutpostShip() {
			t.Errorf("類別 %d:建完前哨站卻沒消耗前哨船", tp)
		}
	}
}

func TestBuildOutpostPreconditions(t *testing.T) {
	// 沒有前哨船。
	s, target := newOutpostTestSession(t, gamedata.GAS_GIANT)
	s.Ships = nil
	if res := s.BuildOutpost(target); res.Ok {
		t.Error("沒有前哨船不該能建")
	}

	// 艦隊不在該星。
	s, target = newOutpostTestSession(t, gamedata.GAS_GIANT)
	s.FleetAtStar = (target + 1) % len(s.Stars)
	if res := s.BuildOutpost(target); res.Ok {
		t.Error("艦隊不在該星不該能建")
	}

	// 該星已有歸屬。
	s, target = newOutpostTestSession(t, gamedata.GAS_GIANT)
	s.Stars[target].Owner = 2
	if res := s.BuildOutpost(target); res.Ok {
		t.Error("別人的星不該能建")
	}

	// 黑洞(無天體)。
	s, target = newOutpostTestSession(t, gamedata.GAS_GIANT)
	s.Planets[target].NoPlanet = true
	if res := s.BuildOutpost(target); res.Ok {
		t.Error("沒有天體的星系不該能建")
	}

	// 同一顆星不能建兩次。
	s, target = newOutpostTestSession(t, gamedata.GAS_GIANT)
	s.Ships = append(s.Ships, Ship{Class: OutpostShipClass}) // 兩艘船
	if res := s.BuildOutpost(target); !res.Ok {
		t.Fatalf("第一次應成功:%s", res.Reason)
	}
	if res := s.BuildOutpost(target); res.Ok {
		t.Error("同一顆星不該能建第二座前哨站")
	}
}

// 手冊 p.119「These outposts act as scanning stations」——前哨站要真的推開戰爭迷霧。
func TestOutpostExtendsDetection(t *testing.T) {
	stars := mkStars(
		[4]float64{0.0, 0.0, 1, 0},  // 母星
		[4]float64{0.9, 0.9, 0, 0},  // 遠方的前哨站所在星
		[4]float64{0.93, 0.9, 0, 0}, // 緊鄰前哨站、離母星很遠的星
	)
	without := playerDetectionVisible(stars, []int{0}, -1, []map[string]bool{{}}, 2, 0, nil)
	with := playerDetectionVisible(stars, []int{0}, -1, []map[string]bool{{}}, 2, 0, []int{1})

	if without[2] {
		t.Fatal("測試前提不成立:沒有前哨站時星 2 就已經看得到")
	}
	if !with[2] {
		t.Error("前哨站應把偵測範圍推到星 2(手冊:outposts act as scanning stations)")
	}
}

// 手冊 p.85:「If a colony is created at an outpost, the building remains and is repurposed
// as Marine Barracks.」
func TestColonizingOutpostStarLeavesMarineBarracks(t *testing.T) {
	s, target := newOutpostTestSession(t, gamedata.HABITABLE)
	if res := s.BuildOutpost(target); !res.Ok {
		t.Fatalf("建前哨站失敗:%s", res.Reason)
	}
	// 建完前哨站後該星變成玩家的,拓殖前置要求「無主」——把它放回無主狀態模擬
	// 「同一顆星上前哨站先於殖民地」這個手冊情境(原版的前哨站星本來就允許再殖民)。
	s.Stars[target].Owner = 0
	s.Ships = append(s.Ships, Ship{Class: ColonyShipClass})
	s.FleetAtStar, s.FleetETA = target, 0

	res := s.ColonizeStar(target)
	if !res.Ok {
		t.Fatalf("拓殖失敗:%s", res.Reason)
	}
	if !s.ColonyHasBuilding(res.ColonyIndex, OutpostMarineBarracks) {
		t.Errorf("前哨站應改建為%s,實際建築:%v",
			OutpostMarineBarracks, s.ColonyBuildingNames(res.ColonyIndex))
	}
	// 改建後前哨站本身應消失(它已經變成建築了,不能再算一座前哨站)。
	if s.HasOutpostAt(target) {
		t.Error("改建成殖民地後不該還留著前哨站")
	}
}

// 沒有前哨站的星拓殖時不該憑空多出海軍陸戰隊營。
func TestColonizingPlainStarHasNoMarineBarracks(t *testing.T) {
	s, target := newOutpostTestSession(t, gamedata.HABITABLE)
	s.Ships = []Ship{{Class: ColonyShipClass}}
	s.FleetAtStar, s.FleetETA = target, 0

	res := s.ColonizeStar(target)
	if !res.Ok {
		t.Fatalf("拓殖失敗:%s", res.Reason)
	}
	if s.ColonyHasBuilding(res.ColonyIndex, OutpostMarineBarracks) {
		t.Error("沒有前哨站的星不該送一棟海軍陸戰隊營")
	}
}

// 兩種支援艦都要能造出來,而且造完真的進艦隊(先前殖民船用掉就再也沒有了)。
func TestSupportShipsAreBuildable(t *testing.T) {
	s := NewDemoSession()
	opts := s.AvailableBuildOptions()
	found := map[string]bool{}
	for _, o := range opts {
		found[o.Name] = true
	}
	for _, name := range []string{gamedata.ColonyShipActionName, gamedata.OutpostShipActionName} {
		if !found[name] {
			t.Errorf("建造選單裡找不到「%s」", name)
		}
	}

	// 完工效果:進艦隊,且不會被誤記成殖民地建築。
	for _, tc := range []struct{ action, class string }{
		{gamedata.ColonyShipActionName, ColonyShipClass},
		{gamedata.OutpostShipActionName, OutpostShipClass},
	} {
		s := NewDemoSession()
		before := 0
		for _, sh := range s.Ships {
			if sh.Class == tc.class {
				before++
			}
		}
		s.applySpecialAction(0, tc.action)
		after := 0
		for _, sh := range s.Ships {
			if sh.Class == tc.class {
				after++
			}
		}
		if after != before+1 {
			t.Errorf("%s 完工後 %s 數量 %d → %d,應 +1", tc.action, tc.class, before, after)
		}
		if s.ColonyHasBuilding(0, tc.action) {
			t.Errorf("%s 是艦艇,不該被記成殖民地建築", tc.action)
		}
	}
}

// 存檔往返要保住前哨站(否則讀檔後掃描範圍縮回去、還能在同一顆星再建一座)。
func TestOutpostSurvivesSaveLoad(t *testing.T) {
	s, target := newOutpostTestSession(t, gamedata.GAS_GIANT)
	if res := s.BuildOutpost(target); !res.Ok {
		t.Fatalf("建前哨站失敗:%s", res.Reason)
	}
	snap := s.snapshot()
	restored := snap.restore()
	if !restored.HasOutpostAt(target) {
		t.Errorf("讀檔後前哨站不見了(Outposts=%v)", restored.Outposts)
	}
}
