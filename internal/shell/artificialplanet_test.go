package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// artificialplanet_test.go:人造行星的護欄。手冊逐字 + 反組譯兩邊都對得上才算。

// apSession 造一個星系:星 0 有玩家殖民地,軌道上放指定的天體。
func apSession(types ...gamedata.PlanetType) *GameSession {
	s := &GameSession{Stars: make([]Star, 1)}
	s.Stars[0].Name = "測試"
	s.Stars[0].Orbits = emptyOrbits()
	for o, t := range types {
		s.Planets = append(s.Planets, Planet{TypeID: t, Orbit: o, Name: "測試 " + string(rune('A'+o))})
		s.Stars[0].Orbits[o] = len(s.Planets) - 1
	}
	s.PlayerColonies = []engine.ColonyState{{}}
	s.PlayerColonyStars = []int{0}
	return s
}

// TestArtificialPlanetPrefersGasGiant 釘住原版的**兩趟掃描順序**。
//
// 原版先掃一遍找氣態巨星(型別 2),沒有才再掃一遍找小行星帶(型別 1)。
// 不能合成一趟——那會讓軌道較內的小行星帶搶先,而原版不是那樣。
func TestArtificialPlanetPrefersGasGiant(t *testing.T) {
	// 小行星帶在**較內**的軌道,氣態巨星在外——合成一趟的話會挑錯。
	s := apSession(gamedata.ASTEROIDS, gamedata.HABITABLE, gamedata.GAS_GIANT)
	got, ok := s.FindArtificialPlanetTarget(0)
	if !ok {
		t.Fatal("應該找得到材料")
	}
	if s.Planets[got.Planet].TypeID != gamedata.GAS_GIANT {
		t.Errorf("應優先挑氣態巨星,實得型別 %d", s.Planets[got.Planet].TypeID)
	}
}

// TestArtificialPlanetSizeFromMaterial:手冊逐字——氣態巨星做出 Huge、小行星帶做出 Large。
//
// 反組譯那邊是 `var_1C = 4` / `= 3`,而 4/3 在尺寸列舉裡正是 HUGE / LARGE。兩邊逐字對上。
func TestArtificialPlanetSizeFromMaterial(t *testing.T) {
	for _, c := range []struct {
		mat  gamedata.PlanetType
		want gamedata.PlanetSize
	}{
		{gamedata.GAS_GIANT, gamedata.HUGE_PLANET},
		{gamedata.ASTEROIDS, gamedata.LARGE_PLANET},
	} {
		s := apSession(c.mat)
		p, ok := s.BuildArtificialPlanet(0)
		if !ok {
			t.Fatalf("材料 %d:應該蓋得成", c.mat)
		}
		if s.Planets[p].SizeID != c.want {
			t.Errorf("材料 %d 應做出大小 %d,實得 %d", c.mat, c.want, s.Planets[p].SizeID)
		}
	}
}

// TestArtificialPlanetResultIsFixed:手冊逐字——Barren、Normal G、Abundant。
func TestArtificialPlanetResultIsFixed(t *testing.T) {
	s := apSession(gamedata.GAS_GIANT)
	p, ok := s.BuildArtificialPlanet(0)
	if !ok {
		t.Fatal("應該蓋得成")
	}
	got := s.Planets[p]
	if got.TypeID != gamedata.HABITABLE {
		t.Errorf("結果應是一般行星(可殖民),實得型別 %d", got.TypeID)
	}
	if got.ClimateID != gamedata.BARREN {
		t.Errorf("氣候應是 Barren,實得 %d", got.ClimateID)
	}
	if got.GravityID != gamedata.NORMAL_G {
		t.Errorf("重力應是 Normal G,實得 %d", got.GravityID)
	}
	if got.MineralID != gamedata.ABUNDANT {
		t.Errorf("礦產應是 Abundant,實得 %d", got.MineralID)
	}
	if got.NoPlanet {
		t.Error("結果不該還帶著「沒有行星」旗標")
	}
}

// TestArtificialPlanetNeedsMaterial 釘住手冊的前置條件。
//
// ⚠ 這一條同時訂正了 remake 先前的假設:gap report 第 61 項寫著「人造行星按定義是在既有
// 星系裡**再多**一顆世界」,於是把它列為「要有空軌道才蓋得了」。手冊說的不是那樣——
// 它是把**既有的**氣態巨星/小行星帶組裝成行星,那顆天體本來就佔著軌道。
// 所以「五個軌道全滿但有氣態巨星」是**可以蓋**的。
func TestArtificialPlanetNeedsMaterialNotFreeOrbit(t *testing.T) {
	// 五個軌道全滿,但其中一個是氣態巨星 → 應該蓋得成。
	full := apSession(gamedata.HABITABLE, gamedata.HABITABLE, gamedata.GAS_GIANT,
		gamedata.HABITABLE, gamedata.HABITABLE)
	if full.FreeOrbit(0) != -1 {
		t.Fatal("前置條件:這個星系應該沒有空軌道")
	}
	if !full.CanBuildArtificialPlanet(0) {
		t.Error("沒有空軌道但有氣態巨星,應該蓋得成——蓋的是既有天體,不是新增軌道")
	}

	// 有空軌道但沒有材料 → 蓋不成。
	bare := apSession(gamedata.HABITABLE)
	if bare.FreeOrbit(0) < 0 {
		t.Fatal("前置條件:這個星系應該有空軌道")
	}
	if bare.CanBuildArtificialPlanet(0) {
		t.Error("沒有氣態巨星/小行星帶就不該蓋得成,空軌道不是條件")
	}
}

// TestArtificialPlanetConsumesTheSpecial:材料被組裝掉了,原本的礦藏不留。
func TestArtificialPlanetConsumesTheSpecial(t *testing.T) {
	s := apSession(gamedata.ASTEROIDS)
	s.Planets[0].SpecialID = gamedata.GoldDeposits
	if _, ok := s.BuildArtificialPlanet(0); !ok {
		t.Fatal("應該蓋得成")
	}
	if s.Planets[0].SpecialID != gamedata.NoSpecial {
		t.Errorf("材料的特殊物產應被消耗,實得 %d", s.Planets[0].SpecialID)
	}
}
