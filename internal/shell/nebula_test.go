package shell

import (
	"math/rand"
	"testing"
)

// seqRandN 回傳固定序列的 rnd(n)(語意同原版 Random:1..n)。
func seqRandN(vals ...int) func(int) int {
	i := 0
	return func(n int) int {
		v := vals[i%len(vals)]
		i++
		if v > n {
			v = n
		}
		if v < 1 {
			v = 1
		}
		return v
	}
}

// TestNebulaCountMatchesOriginalTable 逐檔核對星雲數的四路跳表
// (`Generate_Number_Of_Nebulas_` @ 0x8C4D3)。
//
// 表抄錯的徵狀很隱晦:星雲數只是「少一團」,畫面照樣合理。這裡把每檔的**上下界**都釘住。
func TestNebulaCountMatchesOriginalTable(t *testing.T) {
	cases := []struct {
		size           int
		wantMin, wantMax int
	}{
		{0, 0, 1}, // Random(2) − 1
		{1, 1, 2}, // Random(2)
		{2, 1, 3}, // Random(3)
		{3, 2, 4}, // Random(3) + 1
	}
	for _, c := range cases {
		lo := nebulaCount(c.size, seqRandN(1))
		hi := nebulaCount(c.size, seqRandN(9)) // 夾到 n
		if lo != c.wantMin {
			t.Errorf("大小 %d 的最小值應為 %d,實得 %d", c.size, c.wantMin, lo)
		}
		if hi != c.wantMax {
			t.Errorf("大小 %d 的最大值應為 %d,實得 %d", c.size, c.wantMax, hi)
		}
	}
	// 上限與 internal/save 反推的 maxNebulas 一致——兩邊獨立對上,別讓其中一邊漂走。
	if got := nebulaCount(3, seqRandN(9)); got != maxNebulae {
		t.Errorf("最大檔位的上限應等於 maxNebulae=%d,實得 %d", maxNebulae, got)
	}
	// 超出 0..3 一律 0(原版 `cmp ax, 3; ja default` 直接 return)。
	for _, bad := range []int{-1, 4, 99} {
		if got := nebulaCount(bad, seqRandN(3)); got != 0 {
			t.Errorf("大小 %d 超出範圍應回 0,實得 %d", bad, got)
		}
	}
}

// TestGalaxySizeClassMapsGameOptions 釘住「四個星系大小選項 = 四個檔位」。
//
// ⚠ 這支是為了一個踩過的錯而存在:第一版自己拿原版四檔的星數取中點當界線,
// 結果「中型」被判成檔位 0(星雲數有一半機率是 0),開局星圖常常一團星雲都沒有——
// 而那看起來完全合理。
func TestGalaxySizeClassMapsGameOptions(t *testing.T) {
	for i, g := range GalaxySizes {
		if got := galaxySizeClass(g.Stars); got != i {
			t.Errorf("%s(%d 星)應是檔位 %d,實得 %d", g.Name, g.Stars, i, got)
		}
	}
	// 最小的檔位保證至少能長出星雲的可能;最大的檔位下限必須 ≥ 1(原版 `Random(3) + 1`)。
	if lo := nebulaCount(galaxySizeClass(GalaxySizes[len(GalaxySizes)-1].Stars), seqRandN(1)); lo < 1 {
		t.Errorf("最大星系至少要有 1 團星雲,實得 %d", lo)
	}
	// 單調:星愈多檔位不能倒退。
	prev := -1
	for n := 1; n <= 120; n++ {
		c := galaxySizeClass(n)
		if c < prev {
			t.Fatalf("星數 %d 的檔位 %d 比前一個 %d 小,換算不是單調的", n, c, prev)
		}
		if c < 0 || c >= len(GalaxySizes) {
			t.Fatalf("星數 %d 換出檔位 %d,超出範圍", n, c)
		}
		prev = c
	}
}

// TestGenNebulaeRespectsCountAndTypes:數量不超上限、種類落在 STARBG 有圖的範圍內。
func TestGenNebulaeRespectsCountAndTypes(t *testing.T) {
	stars := make([]Star, 24)
	for i := range stars {
		stars[i] = Star{X: float64(i%6) / 6, Y: float64(i/6) / 4}
	}
	for seed := int64(0); seed < 40; seed++ {
		neb := genNebulae(3, map[int]bool{0: true}, stars, rand.New(rand.NewSource(seed)))
		if len(neb) > maxNebulae {
			t.Fatalf("種子 %d 產生 %d 團星雲,超過上限 %d", seed, len(neb), maxNebulae)
		}
		for _, n := range neb {
			if n.Type < 0 || n.Type >= nebulaTypes {
				t.Fatalf("種子 %d:星雲種類 %d 超出 0..%d(STARBG.LBX 只有這麼多張)",
					seed, n.Type, nebulaTypes-1)
			}
			if n.X < 0 || n.X > 1 || n.Y < 0 || n.Y > 1 {
				t.Fatalf("種子 %d:星雲位置 (%.3f, %.3f) 跑出正規化範圍", seed, n.X, n.Y)
			}
		}
	}
}

// TestGenNebulaeAvoidsHomeStars:母星附近不放星雲——開局就被扣護盾不合理。
func TestGenNebulaeAvoidsHomeStars(t *testing.T) {
	stars := []Star{{X: 0.5, Y: 0.5}, {X: 0.1, Y: 0.9}}
	homes := map[int]bool{0: true}
	for seed := int64(0); seed < 60; seed++ {
		for _, n := range genNebulae(3, homes, stars, rand.New(rand.NewSource(seed))) {
			dx, dy := stars[0].X-(n.X+0.1), stars[0].Y-(n.Y+0.1)
			if dx*dx+dy*dy < 0.04 {
				t.Fatalf("種子 %d:星雲中心 (%.3f, %.3f) 離母星太近", seed, n.X+0.1, n.Y+0.1)
			}
		}
	}
}

// TestSetStarNebulaFlagsSkipsBlackHoles 釘住手冊那條:黑洞不會出現在星雲裡
// (patch 1.5「Mapgen prevents Black Holes from appearing in Nebulas」)。
func TestSetStarNebulaFlagsSkipsBlackHoles(t *testing.T) {
	s := &GameSession{
		Stars:   []Star{{Spectral: 2}, {Spectral: blackHoleSpectral}, {Spectral: 4}},
		Nebulae: []Nebula{{X: 0, Y: 0, Type: 0}},
	}
	s.SetStarNebulaFlags(func(nebIdx, starIdx int) bool { return true }) // 全都「在裡面」
	if !s.Stars[0].InNebula || !s.Stars[2].InNebula {
		t.Error("一般星應被標成在星雲內")
	}
	if s.Stars[1].InNebula {
		t.Error("黑洞不該被標成在星雲內")
	}
}

// TestSetStarNebulaFlagsNilClears:傳 nil 判定式要把旗標全清掉(headless 模擬即這條路徑)。
func TestSetStarNebulaFlagsNilClears(t *testing.T) {
	s := &GameSession{Stars: []Star{{InNebula: true}, {InNebula: true}}}
	s.SetStarNebulaFlags(nil)
	for i, st := range s.Stars {
		if st.InNebula {
			t.Errorf("星 %d 的旗標沒被清掉", i)
		}
	}
}

// TestNebulaShieldDisablesUnlessHardShield 是這一組最重要的一支:手冊 p.158
// 「if combat takes place in a nebula, all shields become inoperative,
//
//	except for those on ships equipped with Hard Shields.」
func TestNebulaShieldDisablesUnlessHardShield(t *testing.T) {
	s := &GameSession{
		Stars:       []Star{{}, {InNebula: true}},
		FleetAtStar: 0,
	}
	// 不在星雲:護盾照常。
	if got := s.nebulaShield(6, false); got != 6 {
		t.Errorf("不在星雲時護盾應維持 6,實得 %d", got)
	}
	// 在星雲:沒有硬化護盾就歸零。
	s.FleetAtStar = 1
	if got := s.nebulaShield(6, false); got != 0 {
		t.Errorf("星雲中沒有硬化護盾,護盾應歸零,實得 %d", got)
	}
	// 在星雲 + 硬化護盾:照常。
	if got := s.nebulaShield(6, true); got != 6 {
		t.Errorf("星雲中有硬化護盾,護盾應維持 6,實得 %d", got)
	}
}

// TestShipHasHardShield:元件名比對(與 shipHasAutoRepair 同作法)。
func TestShipHasHardShield(t *testing.T) {
	if !shipHasHardShield(Ship{Special: "硬化護盾"}) {
		t.Error("裝了硬化護盾應回 true")
	}
	if shipHasHardShield(Ship{Special: "自動修復"}) {
		t.Error("其他特殊元件不該回 true")
	}
	if shipHasHardShield(Ship{}) {
		t.Error("空特殊欄不該回 true")
	}
}

// TestHardShieldIsASelectableComponent 驗證硬化護盾真的進了可選元件表,
// 而且掛在與隱形裝置相同的研究主題上(techtree.go 的三選一)。
//
// 沒有這個元件的話,`gamedata.DamageHardShieldBonus` 與星雲的例外條款**兩者都無法觸及**。
func TestHardShieldIsASelectableComponent(t *testing.T) {
	var found *Component
	for i := range SpecialOptions {
		if SpecialOptions[i].Name == "硬化護盾" {
			found = &SpecialOptions[i]
		}
	}
	if found == nil {
		t.Fatal("SpecialOptions 裡找不到硬化護盾")
	}
	var cloak *Component
	for i := range SpecialOptions {
		if SpecialOptions[i].Name == "隱形裝置" {
			cloak = &SpecialOptions[i]
		}
	}
	if cloak == nil {
		t.Fatal("SpecialOptions 裡找不到隱形裝置(對照組)")
	}
	if found.Tech != cloak.Tech {
		t.Errorf("硬化護盾與隱形裝置應同屬一個研究主題,實得 %v vs %v", found.Tech, cloak.Tech)
	}
	if found.UnlockTech == 0 {
		t.Error("硬化護盾應對應到具體科技 TECH_HARD_SHIELDS,不是走主題層級")
	}
}

// TestStarInNebulaBounds:索引越界不 panic,一律 false。
func TestStarInNebulaBounds(t *testing.T) {
	s := &GameSession{Stars: []Star{{InNebula: true}}}
	if !s.StarInNebula(0) {
		t.Error("星 0 應在星雲內")
	}
	for _, bad := range []int{-1, 1, 99} {
		if s.StarInNebula(bad) {
			t.Errorf("索引 %d 越界應回 false", bad)
		}
	}
}
