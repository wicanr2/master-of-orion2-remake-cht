package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 三個掃描科技的 parsec 值逐字對手冊——先前是自編近似,而且**順序反了**。
//
// ⚠ 這一組先前**零測試覆蓋**,所以「迅子比中子遠」這個錯誤可以一直躺著。
// 一個沒有任何測試看著的查表函式,寫錯不會有任何徵兆。
func TestScannerRangesAreTheManualNumbers(t *testing.T) {
	cases := []struct {
		name                         string
		space, neutron, tachyon      bool
		wantParsec, wantJamReduction int
	}{
		{"無任何掃描科技", false, false, false, 2, 0},
		{"空間掃描儀", true, false, false, 2, 0},
		{"迅子掃描儀", false, false, true, 4, 20},
		{"中子掃描儀", false, true, false, 6, 40},
		{"迅子 + 中子", false, true, true, 6, 40},
		{"三個都有", true, true, true, 6, 40},
	}
	for _, c := range cases {
		if got := gamedata.ScannerRangeParsec(c.space, c.neutron, c.tachyon); got != c.wantParsec {
			t.Errorf("%s 的偵測範圍應為 %d parsec,得到 %d", c.name, c.wantParsec, got)
		}
		if got := gamedata.ScannerMissileEvasionReduction(c.space, c.neutron, c.tachyon); got != c.wantJamReduction {
			t.Errorf("%s 的飛彈閃避抵銷應為 %d,得到 %d", c.name, c.wantJamReduction, got)
		}
	}
}

// 中子在**兩張表上都**贏過迅子——這正是先前那版按科技樹階序挑所踩到的坑。
//
// 手冊:迅子 4 parsec / −20,中子 6 parsec / −40。先前那版把迅子當最高階,
// 結果「研究出迅子」會把更遠的中子覆蓋掉,能力不進反退。
func TestNeutronBeatsTachyonOnBothTables(t *testing.T) {
	tachyonOnly := gamedata.ScannerRangeParsec(false, false, true)
	neutronOnly := gamedata.ScannerRangeParsec(false, true, false)
	if neutronOnly <= tachyonOnly {
		t.Errorf("中子(%d)應比迅子(%d)遠", neutronOnly, tachyonOnly)
	}
	// 兩個都有時不該退回到比較差的那一個。
	if both := gamedata.ScannerRangeParsec(false, true, true); both != neutronOnly {
		t.Errorf("兩個都有時應取較遠的 %d,得到 %d", neutronOnly, both)
	}
	if both := gamedata.ScannerMissileEvasionReduction(false, true, true); both != 40 {
		t.Errorf("兩個都有時抵銷應取較大的 40,得到 %d", both)
	}
}

// 戰鬥掃描器的**第二個效果**:戰鬥之外掃描範圍 +2 parsec。
//
// ⚠ 第 69 項只接了手冊那段的第一句(命中 +50),還寫了一條測試把「只加命中」釘住。
// 一個元件兩個效果、只接一半——與第 69 項慣性穩定器同一個坑,連續兩項都踩。
func TestBattleScannerExtendsScanningRange(t *testing.T) {
	mk := func(special string) []detectionSource {
		s := NewDemoSession()
		s.Fleet().Ships = append(s.Fleet().Ships, Ship{Name: "偵察艦", Class: "巡防艦", Special: special})
		return s.fleetDetectionSources()
	}
	plain, scan := mk(""), mk(battleScannerName)
	if len(plain) == 0 || len(scan) == 0 {
		t.Fatal("測試前提不成立:艦隊應是偵測源")
	}
	if plain[0].bonusParsec != 0 {
		t.Errorf("沒裝戰鬥掃描器的艦隊不該有加成,得到 %d", plain[0].bonusParsec)
	}
	if got := scan[0].bonusParsec; got != gamedata.ShipBattleScannerScanParsecBonus {
		t.Errorf("戰鬥掃描器應 +%d parsec,得到 %d", gamedata.ShipBattleScannerScanParsecBonus, got)
	}
	// 多台不疊加(手冊那句的主詞是「這艘船」,不是艦隊)。
	s := NewDemoSession()
	for i := 0; i < 3; i++ {
		s.Fleet().Ships = append(s.Fleet().Ships, Ship{Name: "偵察艦", Class: "巡防艦", Special: battleScannerName})
	}
	if got := s.fleetDetectionSources()[0].bonusParsec; got != gamedata.ShipBattleScannerScanParsecBonus {
		t.Errorf("三台戰鬥掃描器不該疊加成 %d", got)
	}
}

// 加成真的變成更大的可見半徑(不是只存在 struct 裡)。
func TestBattleScannerActuallyRevealsMoreStars(t *testing.T) {
	// 手擺星圖:母星在原點,另一顆星恰好落在「基礎範圍外、加了 2 parsec 之後範圍內」。
	base := 2
	inner := gamedata.DetectionRangeNormalized(base, 0, 0)
	outer := gamedata.DetectionRangeNormalized(base, gamedata.ShipBattleScannerScanParsecBonus, 0)
	if outer <= inner {
		t.Fatalf("測試前提不成立:加成應擴大半徑(%v vs %v)", outer, inner)
	}
	mid := (inner + outer) / 2
	stars := []Star{
		{Name: "母星", X: 0, Y: 0, Owner: 1},
		{Name: "遠星", X: mid, Y: 0},
	}
	without := playerDetectionVisible(stars, []int{0}, []detectionSource{{starIdx: 0}},
		[]map[string]bool{{}}, base, 0, nil)
	with := playerDetectionVisible(stars, []int{0},
		[]detectionSource{{starIdx: 0, bonusParsec: gamedata.ShipBattleScannerScanParsecBonus}},
		[]map[string]bool{{}}, base, 0, nil)
	if without[1] {
		t.Error("沒有加成時那顆星應在霧裡")
	}
	if !with[1] {
		t.Error("加了戰鬥掃描器的 2 parsec 之後那顆星應可見")
	}
}

// 每一支艦隊都是偵測源——不是只有「目前選中的那一支」。
//
// ⚠ 這先前是真的 bug:切換選中的艦隊會改變戰爭迷霧。選中哪一支是**選單狀態**,
// 不該影響遊戲規則。多艦隊模型上線時 detection.go 沒跟著改。
func TestEveryFleetIsADetectionSourceNotJustTheSelectedOne(t *testing.T) {
	s := NewDemoSession()
	if len(s.Stars) < 3 {
		t.Fatal("測試前提不成立:需要至少三顆星")
	}
	// 造第二支艦隊,擺到一顆**不是**第一支艦隊所在的星上——否則這條測試會空轉。
	far := len(s.Stars) - 1
	if far == s.Fleets[0].AtStar {
		t.Fatalf("測試前提不成立:兩支艦隊在同一顆星(%d),測不出差別", far)
	}
	s.Fleets = append(s.Fleets, Fleet{AtStar: far, DestStar: -1,
		Ships: []Ship{{Name: "偵察艦", Class: "巡防艦"}}})

	srcs := s.fleetDetectionSources()
	if len(srcs) != len(s.Fleets) {
		t.Fatalf("應有 %d 個艦隊偵測源,得到 %d", len(s.Fleets), len(srcs))
	}
	seen := map[int]bool{}
	for _, src := range srcs {
		seen[src.starIdx] = true
	}
	if !seen[far] {
		t.Errorf("第二艦隊所在的星 %d 應是偵測源", far)
	}

	// 切換選中的艦隊不該改變可見性。
	before := s.VisibleStars()
	s.SelectedFleet = 1
	after := s.VisibleStars()
	for i := range before {
		if before[i] != after[i] {
			t.Fatalf("切換選中的艦隊改變了第 %d 顆星的可見性——選單狀態洩漏進遊戲規則", i)
		}
	}
}

// 掃描科技的抵銷真的接進飛彈干擾判定(帝國級,兩條戰鬥路徑都要有)。
func TestScannerJamReductionReachesBothCombatPaths(t *testing.T) {
	plain := NewDemoSession()
	plain.Fleet().Ships = append(plain.Fleet().Ships, Ship{
		Name: "飛彈艦", Class: "戰艦", Weapon: "核飛彈"})
	if cs, _ := plain.mkPlayerCombatantsIndexed(); cs[len(cs)-1].scannerJamReduction != 0 {
		t.Fatalf("測試前提不成立:開局不該有掃描科技加成,得到 %d",
			cs[len(cs)-1].scannerJamReduction)
	}

	tach := NewDemoSession()
	tach.Player = withTech(t, tach.Player, gamedata.TECH_TACHYON_SCANNER)
	tach.Fleet().Ships = append(tach.Fleet().Ships, Ship{
		Name: "飛彈艦", Class: "戰艦", Weapon: "核飛彈"})
	cs, _ := tach.mkPlayerCombatantsIndexed()
	if got := cs[len(cs)-1].scannerJamReduction; got != 20 {
		t.Errorf("快速結算:迅子掃描儀應抵銷 20,得到 %d", got)
	}
	player, _ := tach.StartCombat("測試敵人")
	if got := player[len(player)-1].ScannerJamReduction; got != 20 {
		t.Errorf("格子戰術:迅子掃描儀應抵銷 20,得到 %d", got)
	}
}

// 抵銷真的改變干擾機率(公式端逐字對手冊 p.123 的範例)。
func TestJamChanceUsesTheManualExample(t *testing.T) {
	// 手冊範例:閃避加總 87、攻方迅子掃描儀 20、飛彈具 ECCM → P = (87−20)/2 = 33%。
	if got := gamedata.MissileJamChance(87, 20, true); got != 33 {
		t.Errorf("手冊範例應算出 33%%,得到 %d%%", got)
	}
	// 沒有掃描器時機率明顯較高——這正是先前恆傳 0 造成的偏差。
	if noScanner := gamedata.MissileJamChance(87, 0, true); noScanner <= 33 {
		t.Errorf("沒有掃描器時應更容易被干擾,得到 %d%%", noScanner)
	}
}
