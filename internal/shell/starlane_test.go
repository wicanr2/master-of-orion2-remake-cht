package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestGalaxyDimsCrossCheckSizeFactor 是這一組最重要的一支:四檔銀河尺寸的三重交叉驗證。
//
// 反組譯讀到的表(506×400 / 759×600 / 1012×800 / 1518×1200)、同一段碼寫的 SizeFactor
// (10/15/20/30)、以及原版存檔 SAVE10.GAM(759×600、SizeFactor 15)必須同時成立。
// 任何一項抄錯,`寬 = SizeFactor × 50.6`、`高 = SizeFactor × 40` 這兩個恆等式就會破。
func TestGalaxyDimsCrossCheckSizeFactor(t *testing.T) {
	for i, d := range gamedata.GalaxyDims {
		if got := d.SizeFactor * 40; got != d.Height {
			t.Errorf("檔位 %d:高應為 SizeFactor×40 = %d,表上是 %d", i, got, d.Height)
		}
		// 寬 = SizeFactor × 50.6,用整數比對避免浮點:寬×10 == SizeFactor×506
		if got := d.SizeFactor * 506; got != d.Width*10 {
			t.Errorf("檔位 %d:寬應為 SizeFactor×50.6(即 寬×10 = %d),表上寬×10 = %d",
				i, got, d.Width*10)
		}
	}
	// SAVE10.GAM 這個真實存檔的值必須落在檔位 1。
	if d := gamedata.GalaxyDims[1]; d.Width != 759 || d.Height != 600 || d.SizeFactor != 15 {
		t.Errorf("檔位 1 應為 759×600 / SizeFactor 15(SAVE10.GAM 讀出來的值),實得 %+v", d)
	}
}

// TestGalaxySizeFromStarsThresholds 釘住原版 `Galaxy_Size_From_N_Stars_` 的四個門檻。
func TestGalaxySizeFromStarsThresholds(t *testing.T) {
	cases := []struct{ n, want int }{
		{1, 0}, {20, 0}, {21, 1}, {36, 1}, {37, 2}, {54, 2}, {55, 3}, {72, 3}, {200, 3},
	}
	for _, c := range cases {
		if got := gamedata.GalaxySizeFromStars(c.n); got != c.want {
			t.Errorf("%d 星應是檔位 %d,實得 %d", c.n, c.want, got)
		}
	}
}

// TestGalaxySizesMatchOriginalThresholds:遊戲提供的四個星系大小就是原版那四檔。
//
// ⚠ 這支替換掉一個踩過的錯:先前 remake 自訂 12/24/36/48,而星雲數、銀河跨距這些表
// 都是**以檔位為索引**的,對不上就整串偏掉(星雲那次的徵狀是開局常常一團都沒有)。
func TestGalaxySizesMatchOriginalThresholds(t *testing.T) {
	if len(GalaxySizes) != len(gamedata.GalaxyStarCounts) {
		t.Fatalf("星系大小選項應有 %d 檔,實得 %d", len(gamedata.GalaxyStarCounts), len(GalaxySizes))
	}
	for i, g := range GalaxySizes {
		if g.Stars != gamedata.GalaxyStarCounts[i] {
			t.Errorf("%s 應為 %d 星,實得 %d", g.Name, gamedata.GalaxyStarCounts[i], g.Stars)
		}
		if got := gamedata.GalaxySizeFromStars(g.Stars); got != i {
			t.Errorf("%s(%d 星)應對到檔位 %d,實得 %d", g.Name, g.Stars, i, got)
		}
	}
}

// TestParsecsBetweenMatchesOriginalCeil 釘住 `Parsecs_Between_Points_` 的無條件進位語意。
//
// 原版是「最小的 p 使得 p²×900 ≥ d²」,d 以遊戲單位計 —— 換成秒差距就是 ceil(距離)。
func TestParsecsBetweenMatchesOriginalCeil(t *testing.T) {
	cases := []struct {
		dx, dy float64
		want   int
	}{
		{0, 0, 0},
		{1, 0, 1},
		{3, 4, 5},    // 正好 5,不進位
		{3.01, 4, 6}, // 超過一點點就進位
		{0.1, 0, 1},  // 再近都算 1 秒差距
		{0, 12, 12},
	}
	for _, c := range cases {
		if got := gamedata.ParsecsBetween(c.dx, c.dy); got != c.want {
			t.Errorf("(%.2f, %.2f) 應為 %d 秒差距,實得 %d", c.dx, c.dy, c.want, got)
		}
	}
}

// TestDriveSpeedsMatchManual 逐級核對引擎速度(手冊的元件說明各寫死一個數字)。
func TestDriveSpeedsMatchManual(t *testing.T) {
	want := []int{0, 2, 3, 4, 5, 6, 7} // 無 / 核融 / 融合 / 離子 / 反物質 / 超空間 / 相位
	if len(gamedata.DriveSpeeds) != len(want) {
		t.Fatalf("引擎速度表應有 %d 項,實得 %d", len(want), len(gamedata.DriveSpeeds))
	}
	for i, w := range want {
		if gamedata.DriveSpeeds[i] != w {
			t.Errorf("第 %d 階引擎應為 %d 秒差距/回合,實得 %d", i, w, gamedata.DriveSpeeds[i])
		}
	}
	if len(gamedata.DriveTechOrder) != len(want)-1 {
		t.Errorf("引擎科技清單應有 %d 級(速度表扣掉「無」),實得 %d",
			len(want)-1, len(gamedata.DriveTechOrder))
	}
	// 速度必須嚴格遞增——研究了更高階引擎卻變慢是明顯的表抄錯。
	for i := 2; i < len(gamedata.DriveSpeeds); i++ {
		if gamedata.DriveSpeeds[i] <= gamedata.DriveSpeeds[i-1] {
			t.Errorf("第 %d 階(%d)沒有比第 %d 階(%d)快",
				i, gamedata.DriveSpeeds[i], i-1, gamedata.DriveSpeeds[i-1])
		}
	}
}

// TestFleetSpeedFallsBackToNuclearWhenFTL 釘住那個下界:有 FTL 就至少是核融引擎。
//
// ⚠ 這支是為了一個踩過的錯而存在:`FleetHasFTL` 對非曲速前開局直接回 true、不看科技表,
// 所以那些開局的引擎階查出來是 0 → 航速 0 → **ETA 全被夾成 1**,整個秒差距模型形同虛設,
// 而且畫面上完全看不出來(每一趟都是「1 回合到」,看起來只是船很快)。
func TestFleetSpeedFallsBackToNuclearWhenFTL(t *testing.T) {
	s := NewDemoSession()
	if !s.FleetHasFTL() {
		t.Skip("demo 局沒有 FTL")
	}
	if got := s.FleetSpeedParsecs(); got < gamedata.DriveSpeeds[1] {
		t.Errorf("有 FTL 時航速至少應為核融引擎的 %d,實得 %d", gamedata.DriveSpeeds[1], got)
	}
}

// TestFleetETAUsesParsecsNotFixedFactor:ETA 必須隨距離變化,而且遠的比近的久。
func TestFleetETAUsesParsecsNotFixedFactor(t *testing.T) {
	s := NewDemoSession()
	near, far := -1, -1
	nearP, farP := 1<<30, -1
	for i := 1; i < len(s.Stars); i++ {
		p := s.ParsecsBetweenStars(0, i)
		if p < nearP {
			nearP, near = p, i
		}
		if p > farP {
			farP, far = p, i
		}
	}
	if near < 0 || far < 0 || nearP == farP {
		t.Skip("星圖上找不到距離差異夠大的兩顆星")
	}
	if s.FleetETATo(0, far) < s.FleetETATo(0, near) {
		t.Errorf("遠星(%d 秒差距)的 ETA %d 不該小於近星(%d 秒差距)的 %d",
			farP, s.FleetETATo(0, far), nearP, s.FleetETATo(0, near))
	}
}

// TestNebulaSlowsFleetToOneParsec 釘住手冊那條:星雲中航速降為 1 秒差距/回合。
func TestNebulaSlowsFleetToOneParsec(t *testing.T) {
	s := NewDemoSession()
	// 找一顆夠遠的星,確保正常航速下 ETA > 1。
	dest := -1
	for i := 1; i < len(s.Stars); i++ {
		if s.ParsecsBetweenStars(0, i) >= 4 {
			dest = i
			break
		}
	}
	if dest < 0 {
		t.Skip("星圖上找不到 4 秒差距以外的星")
	}
	base := s.FleetETATo(0, dest)
	pc := s.ParsecsBetweenStars(0, dest)

	s.Stars[dest].InNebula = true
	slowed := s.FleetETATo(0, dest)
	if slowed != pc {
		t.Errorf("目的星在星雲內時應以 1 秒差距/回合計(= %d 回合),實得 %d", pc, slowed)
	}
	if slowed <= base {
		t.Errorf("星雲應該讓航程變久:原本 %d、星雲中 %d", base, slowed)
	}
}

// TestNavigatorIgnoresNebula 釘住手冊那條:領航員可無視星雲造成的移動限制。
func TestNavigatorIgnoresNebula(t *testing.T) {
	s := NewDemoSession()
	dest := -1
	for i := 1; i < len(s.Stars); i++ {
		if s.ParsecsBetweenStars(0, i) >= 4 {
			dest = i
			break
		}
	}
	if dest < 0 {
		t.Skip("星圖上找不到 4 秒差距以外的星")
	}
	s.Stars[dest].InNebula = true
	slowed := s.FleetETATo(0, dest)

	s.Leaders = append(s.Leaders, Leader{Name: "測試領航員", Skill: navigatorSkillLabel, Level: 3, Ship: true, Tier: 1})
	if !s.FleetHasNavigator() {
		t.Fatal("加了艦艇領航員之後 FleetHasNavigator 應為 true")
	}
	withNav := s.FleetETATo(0, dest)
	if withNav >= slowed {
		t.Errorf("領航員應可無視星雲降速:無領航 %d 回合、有領航 %d 回合", slowed, withNav)
	}
}

// TestNavigatorMustBeShipOfficer:殖民地領袖不隨艦隊走,不該算數。
func TestNavigatorMustBeShipOfficer(t *testing.T) {
	s := NewDemoSession()
	s.Leaders = append(s.Leaders, Leader{Name: "地面領航員", Skill: navigatorSkillLabel, Ship: false, Tier: 1})
	if s.FleetHasNavigator() {
		t.Error("殖民地領袖(Ship=false)不該讓艦隊取得領航效果")
	}
}
