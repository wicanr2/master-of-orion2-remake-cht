package shell

import (
	"os"
	"path/filepath"
	"testing"
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

// TestGeneratedStarsHaveExactlyOneOccupiedOrbit 釘住**這一階段刻意的限制**。
//
// 產生器目前仍然每顆星只放一顆行星(放在它骰到的軌道上),其餘空著——
// 行為逐位元不變,換的是形狀不是內容。
//
// 把 SystemBodies 升格成真正的行星之後這條會紅,**那時候該改的是測試**。
func TestGeneratedStarsHaveExactlyOneOccupiedOrbit(t *testing.T) {
	s := NewDemoSession()
	for i := range s.Stars {
		n := 0
		for _, p := range s.Stars[i].Orbits {
			if p != OrbitEmpty {
				n++
			}
		}
		if n != 1 {
			t.Fatalf("星 %d 目前應恰好一個軌道有行星,實得 %d", i, n)
		}
	}
}

// TestPlanetAtMatchesLegacyParallelIndexing:`PlanetAt(星)` 要與舊的 `Planets[星]` 同義。
//
// 這是整個遷移的相容性支點——一星一行星時兩者必須逐項相同,否則舊呼叫端換過來就會位移。
func TestPlanetAtMatchesLegacyParallelIndexing(t *testing.T) {
	s := NewDemoSession()
	for i := range s.Stars {
		if got := s.PlanetAt(i); got != i {
			t.Fatalf("星 %d 的 PlanetAt 應等於舊的平行索引 %d,實得 %d", i, i, got)
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
