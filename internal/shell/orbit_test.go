package shell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// orbit_test.go:一星多行星資料模型的護欄。

// TestStarOrbitsIsFive 釘住原版真值。
//
// 三個獨立來源(見 orbit.go 檔頭):偏移算術 0x54−0x4A = 10 bytes = 5 words、
// `System_Planet_Scanned_To_Planet_Id_` 的索引式、迴圈上界 `cmp ..., 5; jge`。
func TestStarOrbitsIsFive(t *testing.T) {
	if StarOrbits != 5 {
		t.Fatalf("每個星系應有 5 個軌道(原版真值),實得 %d", StarOrbits)
	}
}

// TestEmptyOrbitsIsMinusOneNotZero 釘住這個模型的零值陷阱。
//
// Go 的零值是 0,而 0 是**行星 0 的索引**。軌道表若用零值當「空」,
// 每顆星都會宣稱軌道 0 上有行星 0——同 Star.Wormhole 那個坑。
func TestEmptyOrbitsIsMinusOneNotZero(t *testing.T) {
	if OrbitEmpty != -1 {
		t.Fatalf("空軌道應是 −1,實得 %d", OrbitEmpty)
	}
	o := emptyOrbits()
	for i, p := range o {
		if p != OrbitEmpty {
			t.Errorf("軌道 %d 預設應為空(−1),實得 %d", i, p)
		}
	}
}

// TestEveryOrbitEntryIsAValidDistinctPlanet:軌道表要指到真的行星,而且不能有兩格指同一顆。
//
// 這條先前叫 `...HasExactlyOneOccupiedOrbit`,釘的是「每顆星只放一顆行星」這個階段性限制,
// 註解裡寫明「升格之後這條會紅,那時候該改的是測試」。升格做完了(第 24 項(軌道資料層)),所以改了。
func TestEveryOrbitEntryIsAValidDistinctPlanet(t *testing.T) {
	s := NewDemoSession()
	seen := map[int]int{} // 行星索引 → 是哪顆星的
	total := 0
	for i := range s.Stars {
		n := 0
		for o, p := range s.Stars[i].Orbits {
			if p == OrbitEmpty {
				continue
			}
			if p < 0 || p >= len(s.Planets) {
				t.Fatalf("星 %d 軌道 %d 指到越界的行星 %d", i, o, p)
			}
			if prev, dup := seen[p]; dup {
				t.Fatalf("行星 %d 同時掛在星 %d 與星 %d 上", p, prev, i)
			}
			seen[p] = i
			n++
			total++
		}
		if n < 1 || n > StarOrbits {
			t.Fatalf("星 %d 的天體數應在 1..%d,實得 %d", i, StarOrbits, n)
		}
	}
	if total != len(s.Planets) {
		t.Errorf("每一顆行星都該掛在某個軌道上:掛了 %d 顆,實有 %d 顆", total, len(s.Planets))
	}
}

// TestGalaxyHasMultiPlanetSystems:升格之後應該真的出現多天體星系。
//
// 沒有這一條的話,「升格」可能只是把資料搬個位置而實際仍然一星一顆——
// 而那從上面那條測試看不出來(1 也在 1..5 的範圍內)。
func TestGalaxyHasMultiPlanetSystems(t *testing.T) {
	s := NewDemoSession()
	if len(s.Planets) <= len(s.Stars) {
		t.Fatalf("行星數應多於星數(每顆星 1..5 個天體),實得 %d 顆行星 / %d 顆星",
			len(s.Planets), len(s.Stars))
	}
	multi := 0
	for i := range s.Stars {
		if len(s.PlanetsAt(i)) > 1 {
			multi++
		}
	}
	if multi == 0 {
		t.Error("整個銀河沒有任何多天體星系——升格沒有真的生效")
	}
}

// TestPlanetAtPrefersHabitable:代表行星的挑法要與產生器逐字相同。
//
// **這是整個遷移的相容性支點**:挑法不一致,所有走 `PlanetAt` 的舊呼叫端就會位移,
// 而位移的徵狀是「殖民地總覽說類地、殖民地畫面說凍原」那一類自打嘴巴,不是崩潰。
func TestPlanetAtPrefersHabitable(t *testing.T) {
	s := NewDemoSession()
	for i := range s.Stars {
		rep := s.PlanetAt(i)
		if rep < 0 {
			t.Fatalf("星 %d 挑不到代表行星", i)
		}
		if s.Planets[rep].TypeID == gamedata.HABITABLE {
			continue
		}
		// 代表不是一般行星 → 整個星系都不能有一般行星,否則挑錯了。
		for _, p := range s.PlanetsAt(i) {
			if s.Planets[p].TypeID == gamedata.HABITABLE {
				t.Fatalf("星 %d 的代表是 %d(非一般行星),但軌道上有一般行星 %d", i, rep, p)
			}
		}
	}
}

// TestPlanetStarAndOrbitRoundTrip:行星 → 星 / 軌道 的反查要對得回去。
func TestPlanetStarAndOrbitRoundTrip(t *testing.T) {
	s := NewDemoSession()
	for i := range s.Stars {
		p := s.PlanetAt(i)
		if p < 0 {
			continue
		}
		if got := s.PlanetStar(p); got != i {
			t.Errorf("行星 %d 應屬於星 %d,實得 %d", p, i, got)
		}
		o := s.PlanetOrbit(p)
		if o < 0 || o >= StarOrbits {
			t.Errorf("行星 %d 的軌道號越界:%d", p, o)
		}
		if s.Stars[i].Orbits[o] != p {
			t.Errorf("軌道表對不回去:星 %d 軌道 %d 是 %d,不是 %d", i, o, s.Stars[i].Orbits[o], p)
		}
	}
}

// TestFreeOrbitFindsSpace:人造行星要用它——沒有空軌道就蓋不了。
func TestFreeOrbitFindsSpace(t *testing.T) {
	s := &GameSession{Stars: make([]Star, 2)}
	s.Stars[0].Orbits = emptyOrbits()
	s.Stars[1].Orbits = emptyOrbits()
	if got := s.FreeOrbit(0); got != 0 {
		t.Errorf("全空的星第一個空軌道應是 0,實得 %d", got)
	}
	for o := range s.Stars[1].Orbits {
		s.Stars[1].Orbits[o] = 7 // 全滿
	}
	if got := s.FreeOrbit(1); got != -1 {
		t.Errorf("全滿的星應回 −1,實得 %d", got)
	}
	if got := s.FreeOrbit(9); got != -1 {
		t.Errorf("越界的星應回 −1,實得 %d", got)
	}
}

// TestLoadLegacySaveRebuildsOrbits:舊存檔沒有 Orbits,要重建成一星一行星的形狀。
//
// ⚠ 不重建的話零值 0 會讓每顆星都指向行星 0 —— 而那**不會報錯**,
// 只會讓每顆星的行星資料看起來都一樣。
func TestLoadLegacySaveRebuildsOrbits(t *testing.T) {
	const legacy = `{
	  "version": 1,
	  "turn": 3,
	  "stars": [{"Name":"甲"},{"Name":"乙"},{"Name":"丙"}],
	  "planets": [{"Name":"甲 I"},{"Name":"乙 I"},{"Name":"丙 I"}],
	  "selectedStar": -1
	}`
	path := filepath.Join(t.TempDir(), "legacy.json")
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := LoadSession(path)
	if err != nil {
		t.Fatalf("舊格式存檔應讀得回來:%v", err)
	}
	for i := range s.Stars {
		if got := s.PlanetAt(i); got != i {
			t.Errorf("舊檔的星 %d 應接回行星 %d,實得 %d", i, i, got)
		}
		for o := 1; o < StarOrbits; o++ {
			if s.Stars[i].Orbits[o] != OrbitEmpty {
				t.Errorf("星 %d 的軌道 %d 應為空,實得 %d", i, o, s.Stars[i].Orbits[o])
			}
		}
	}
}
