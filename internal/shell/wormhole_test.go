package shell

import "testing"

// wormhole_test.go:蟲洞的護欄。
//
// 最重要的一條是**雙向**:openorion2 `gamestate.cpp:1946` 對單向蟲洞直接丟例外
// ("One-way wormholes not allowed"),而單向在畫面上看起來完全正常(線照樣畫一條),
// 只有從另一端走的時候才會發現走不回來。

// mkWormholeStars 建 n 顆等距排開的星(X 從 0 到 1),方便控制距離。
func mkWormholeStars(n int) []Star {
	out := make([]Star, n)
	for i := range out {
		out[i] = Star{X: float64(i) / float64(n-1), Y: 0.5, Wormhole: -1}
	}
	return out
}

// seqRand 是決定性的假亂數:依序回傳 0,1,2,… 取模,讓測試不依賴 math/rand 的實作。
func seqRand() func(int) int {
	i := 0
	return func(n int) int {
		if n <= 0 {
			return 0
		}
		v := i % n
		i++
		return v
	}
}

// TestGenWormholesIsMutual:配出來的每一對都必須雙向。
func TestGenWormholesIsMutual(t *testing.T) {
	stars := mkWormholeStars(24)
	genWormholes(stars, map[int]bool{0: true}, seqRand())
	pairs := 0
	for i, st := range stars {
		if st.Wormhole < 0 {
			continue
		}
		pairs++
		j := st.Wormhole
		if j < 0 || j >= len(stars) {
			t.Fatalf("星 %d 的蟲洞指向越界的 %d", i, j)
		}
		if stars[j].Wormhole != i {
			t.Errorf("單向蟲洞:星 %d → %d,但 %d → %d", i, j, j, stars[j].Wormhole)
		}
	}
	if pairs == 0 {
		t.Error("24 顆星應該至少配出一對蟲洞")
	}
	if pairs%2 != 0 {
		t.Errorf("有蟲洞的星數 %d 是奇數——一定有一端沒配到", pairs)
	}
}

// TestGenWormholesSkipsHomeAndBlackHoles:母星與黑洞不可當端點(原版兩者都排除)。
func TestGenWormholesSkipsHomeAndBlackHoles(t *testing.T) {
	stars := mkWormholeStars(24)
	stars[5].Spectral = blackHoleSpectral
	stars[9].Spectral = blackHoleSpectral
	homes := map[int]bool{0: true, 1: true, 2: true}

	genWormholes(stars, homes, seqRand())
	for i, st := range stars {
		if st.Wormhole < 0 {
			continue
		}
		if homes[i] {
			t.Errorf("母星 %d 被當成蟲洞端點", i)
		}
		if st.Spectral == blackHoleSpectral {
			t.Errorf("黑洞 %d 被當成蟲洞端點", i)
		}
	}
}

// TestGenWormholesRespectsMinDistance:原版有最短距離門檻,蟲洞不連鄰星
// (不然它就只是一條沒意義的捷徑)。
func TestGenWormholesRespectsMinDistance(t *testing.T) {
	stars := mkWormholeStars(24)
	genWormholes(stars, map[int]bool{0: true}, seqRand())
	for i, st := range stars {
		j := st.Wormhole
		if j <= i {
			continue
		}
		dx, dy := stars[i].X-stars[j].X, stars[i].Y-stars[j].Y
		if d2 := dx*dx + dy*dy; d2 <= wormholeMinDist*wormholeMinDist {
			t.Errorf("星 %d↔%d 的距離平方 %.4f 沒過門檻 %.4f", i, j, d2, wormholeMinDist*wormholeMinDist)
		}
	}
}

// TestNormalizeWormholesFixesLegacySaves:**舊存檔沒有這個欄位,零值是 0**——
// 那會讓每顆星都宣稱與星 0 有蟲洞(星圖畫滿放射狀連線、艦隊到處一回合直達)。
// 這條是這個功能最容易靜默壞掉的地方。
func TestNormalizeWormholesFixesLegacySaves(t *testing.T) {
	legacy := make([]Star, 6) // 全部零值:Wormhole == 0
	normalizeWormholes(legacy)
	for i, st := range legacy {
		if st.Wormhole != -1 {
			t.Errorf("舊存檔的星 %d 應正規化成 -1,實得 %d", i, st.Wormhole)
		}
	}

	// 單向、越界、自己連自己都要清掉。
	stars := mkWormholeStars(6)
	stars[1].Wormhole = 4 // 單向
	stars[2].Wormhole = 99
	stars[3].Wormhole = 3
	stars[5].Wormhole = 0 // 指向星 0,但星 0 沒回指
	normalizeWormholes(stars)
	for i, st := range stars {
		if st.Wormhole != -1 {
			t.Errorf("星 %d 的不合法蟲洞沒被清掉,實得 %d", i, st.Wormhole)
		}
	}

	// 合法的雙向對不能被誤殺。
	ok := mkWormholeStars(6)
	ok[1].Wormhole, ok[4].Wormhole = 4, 1
	normalizeWormholes(ok)
	if ok[1].Wormhole != 4 || ok[4].Wormhole != 1 {
		t.Errorf("合法的雙向對被清掉了:%d / %d", ok[1].Wormhole, ok[4].Wormhole)
	}
}

// TestWormholeMakesTravelInstant:蟲洞的**遊戲機制**價值——走它一回合就到,
// 不管兩端在星圖上隔多遠。沒接上的話它就只是一條裝飾線。
func TestWormholeMakesTravelInstant(t *testing.T) {
	s := NewDemoSession()
	// 找一對真的很遠的星,手動接上蟲洞,比較有/沒有蟲洞的 ETA。
	normalizeWormholes(s.Stars)
	far, farD := -1, 0.0
	for i := range s.Stars {
		dx := s.Stars[i].X - s.Stars[0].X
		dy := s.Stars[i].Y - s.Stars[0].Y
		if d := dx*dx + dy*dy; d > farD {
			far, farD = i, d
		}
	}
	if far < 0 {
		t.Fatal("找不到目標星")
	}
	s.Player.CompletedTopics = nil // 讓 FTL 檢查走既有路徑;下面直接比 ETA

	// 先量沒有蟲洞時的 ETA。
	for i := range s.Stars {
		s.Stars[i].Wormhole = -1
	}
	s.FleetAtStar, s.FleetDestStar, s.FleetETA = 0, -1, 0
	if !s.SendFleet(far) {
		t.Skip("這局的艦隊派不出去(FTL 未解鎖),ETA 比較不成立")
	}
	plain := s.FleetETA
	if plain <= 1 {
		t.Skipf("最遠的星只要 %d 回合,測不出差異", plain)
	}

	// 接上蟲洞再量一次。
	s.Stars[0].Wormhole, s.Stars[far].Wormhole = far, 0
	s.FleetAtStar, s.FleetDestStar, s.FleetETA = 0, -1, 0
	if !s.SendFleet(far) {
		t.Fatal("接上蟲洞之後反而派不出去")
	}
	if s.FleetETA != 1 {
		t.Errorf("走蟲洞的 ETA = %d,應為 1(原本要 %d 回合)", s.FleetETA, plain)
	}
	if !s.WormholeBetween(0, far) {
		t.Error("WormholeBetween 沒認出這一對")
	}
	if s.WormholeBetween(0, 0) {
		t.Error("WormholeBetween 不該認自己連自己")
	}
}
