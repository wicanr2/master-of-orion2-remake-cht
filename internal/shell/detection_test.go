package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// detection_test.go:diff 全量表 #13(掃描/偵測距離)輕量戰爭迷霧的單元測試。全部針對純函式
// playerDetectionVisible(不依賴 *GameSession),用手擺的合成星圖精準控制距離,避免依賴程序化
// 星系(genGalaxy)的隨機分布導致測試脆弱。

// mkStars 是測試用的合成星圖建構器:每個參數是 (x, y, owner, explored) 四元組,座標與
// shell.Star.X/Y 同規格(0..1 正規化),owner 與 Star.Owner 同語意(0=無主/1=玩家/2=AI)。
func mkStars(specs ...[4]float64) []Star {
	stars := make([]Star, len(specs))
	for i, sp := range specs {
		stars[i] = Star{X: sp[0], Y: sp[1], Owner: int(sp[2]), Explored: sp[3] != 0}
	}
	return stars
}

// TestPlayerDetectionVisible_HomeworldAndRange 驗證開局母星可見、範圍外遠星不可見、範圍內
// 未探索星可見(對應設計驗證項 ①②)。
func TestPlayerDetectionVisible_HomeworldAndRange(t *testing.T) {
	stars := mkStars(
		[4]float64{0.5, 0.5, 1, 0},  // 星 0:玩家母星(Owner=1,未 Explored 也應可見)
		[4]float64{0.55, 0.5, 0, 0}, // 星 1:距母星 0.05,在基礎偵測範圍(2 parsec*0.1=0.20)內、無主、未探索
		[4]float64{0.9, 0.9, 0, 0},  // 星 2:距母星甚遠(~0.57),遠超基礎偵測範圍,無主、未探索
	)
	playerColonyStars := []int{0}
	colonyBuildings := []map[string]bool{{}} // 母星無任何軌道基地
	const scannerParsec = 2                  // 未解鎖任何掃描科技(基礎值)
	const versionBonus = 0

	visible := playerDetectionVisible(stars, playerColonyStars, 0, colonyBuildings, scannerParsec, versionBonus)

	if !visible[0] {
		t.Error("星 0(母星,Owner=1)應可見")
	}
	if !visible[1] {
		t.Error("星 1(偵測範圍內、未探索)應可見——這是設計要「揭示鄰近星」的核心效果")
	}
	if visible[2] {
		t.Error("星 2(偵測範圍外、未探索、無主)不應可見——應被 fog 遮蔽")
	}
}

// TestPlayerDetectionVisible_ExploredAlwaysVisible 驗證已探索星恆可見,即使遠在偵測範圍外
// (對應設計驗證項 ④)——艦隊探索過的星不會因為離開該處後又被重新蓋上迷霧。
func TestPlayerDetectionVisible_ExploredAlwaysVisible(t *testing.T) {
	stars := mkStars(
		[4]float64{0.5, 0.5, 1, 0},   // 母星
		[4]float64{0.95, 0.95, 0, 1}, // 遠星,無主但已探索(Explored=true)
	)
	visible := playerDetectionVisible(stars, []int{0}, 0, []map[string]bool{{}}, 2, 0)

	if !visible[1] {
		t.Error("已探索星(Explored=true)無論距離都應恆可見")
	}
}

// TestPlayerDetectionVisible_VersionDiff 驗證版本差異:同盤面下 Profile15(+1 parsec)偵測範圍
// 比 Profile13 大,能看到 Profile13 看不到的邊界星(對應設計驗證項 ③)。
func TestPlayerDetectionVisible_VersionDiff(t *testing.T) {
	// 母星在原點,基礎掃描 2 parsec + 母星星基 2 parsec(ParsecToNormalized=1/10):
	//   Profile13(versionBonus=0):range = (2+2+0)*0.1 = 0.40
	//   Profile15(versionBonus=1):range = (2+2+1)*0.1 = 0.50
	// 邊界星距離設在 0.45,恰好落在兩者之間——Profile13 看不到、Profile15 看得到。
	stars := mkStars(
		[4]float64{0.0, 0.0, 1, 0},  // 母星(有星基)
		[4]float64{0.45, 0.0, 0, 0}, // 邊界星:介於兩版偵測範圍之間
	)
	colonyBuildings := []map[string]bool{{"星基": true}}
	const scannerParsec = 2

	p13 := gamedata.Profile13()
	p15 := gamedata.Profile15()

	visible13 := playerDetectionVisible(stars, []int{0}, 0, colonyBuildings, scannerParsec, p13.SensorRangeVersionBonusParsec)
	visible15 := playerDetectionVisible(stars, []int{0}, 0, colonyBuildings, scannerParsec, p15.SensorRangeVersionBonusParsec)

	if visible13[1] {
		t.Error("Profile13(無 #13 修正)在此距離不應看到邊界星")
	}
	if !visible15[1] {
		t.Error("Profile15(#13 +1 parsec 修正)在此距離應看到邊界星——版本差異的核心效果")
	}

	count13, count15 := 0, 0
	for _, v := range visible13 {
		if v {
			count13++
		}
	}
	for _, v := range visible15 {
		if v {
			count15++
		}
	}
	if count15 < count13 {
		t.Errorf("Profile15 可見星數(%d)不應少於 Profile13(%d)——1.5 多揭示一圈星是設計預期", count15, count13)
	}
	if count15 <= count13 {
		t.Errorf("本測試盤面設計為讓 Profile15 嚴格多看到 1 顆星,got count13=%d count15=%d(未觀察到差異,盤面設計可能需要調整)", count13, count15)
	}
}

// TestPlayerDetectionVisible_OrbitalBonus 驗證軌道基地加成:有星辰要塞的殖民地偵測範圍比只有
// 星基的殖民地大(對應設計驗證項 ⑤)。
func TestPlayerDetectionVisible_OrbitalBonus(t *testing.T) {
	// 星基加成 2 parsec:range=(2+2+0)*0.1=0.40;星辰要塞加成 6 parsec:range=(2+6+0)*0.1=0.80。
	// 邊界星距離設在 0.60,星基版看不到、星辰要塞版看得到。
	stars := mkStars(
		[4]float64{0.0, 0.0, 1, 0},
		[4]float64{0.60, 0.0, 0, 0},
	)
	const scannerParsec = 2
	const versionBonus = 0

	visibleStarBase := playerDetectionVisible(stars, []int{0}, 0, []map[string]bool{{"星基": true}}, scannerParsec, versionBonus)
	visibleFortress := playerDetectionVisible(stars, []int{0}, 0, []map[string]bool{{"星辰要塞": true}}, scannerParsec, versionBonus)

	if visibleStarBase[1] {
		t.Error("只有星基的殖民地,在此距離不應看到邊界星")
	}
	if !visibleFortress[1] {
		t.Error("有星辰要塞的殖民地,在此距離應看到邊界星——軌道基地加成應比星基大")
	}
}

// TestPlayerDetectionVisible_FleetSource 驗證艦隊所在星也是偵測源(無軌道加成),對應設計
// 「偵測源含艦隊所在星(FleetAtStar)」的描述。
func TestPlayerDetectionVisible_FleetSource(t *testing.T) {
	stars := mkStars(
		[4]float64{0.0, 0.0, 1, 0},  // 母星(玩家殖民地,遠離下面的艦隊)
		[4]float64{0.9, 0.9, 0, 0},  // 艦隊目前所在的無主星
		[4]float64{0.95, 0.9, 0, 0}, // 距艦隊很近、距母星很遠的星
	)
	// 母星附近沒有任何殖民地能看到星 2,只有艦隊移動過去才能揭示它。
	visible := playerDetectionVisible(stars, []int{0}, 1, []map[string]bool{{}}, 2, 0)

	if !visible[1] {
		t.Error("艦隊當前所在星本身應可見(星圖上正在待著的星)")
	}
	if !visible[2] {
		t.Error("艦隊所在星的偵測範圍應能揭示鄰近星,即使該處沒有玩家殖民地")
	}
}

// TestGameSession_VisibleStars_Homeworld 用真實 NewDemoSession 驗證 GameSession.VisibleStars/
// starVisible 有正確接上 Player/PlayerColonyStars/FleetAtStar/ColonyBuildings/RuleProfile,
// 母星(星 0)可見。純粹的偵測範圍數學已由上面的合成星圖測試覆蓋,這裡只驗證接線正確。
func TestGameSession_VisibleStars_Homeworld(t *testing.T) {
	s := NewDemoSession()
	visible := s.VisibleStars()
	if len(visible) != len(s.Stars) {
		t.Fatalf("VisibleStars() 長度 = %d,want %d(等長 s.Stars)", len(visible), len(s.Stars))
	}
	if !visible[0] {
		t.Error("母星(星 0)應可見")
	}
	if !s.starVisible(0) {
		t.Error("starVisible(0) 應與 VisibleStars()[0] 一致,回傳 true")
	}
	if s.starVisible(-1) || s.starVisible(len(s.Stars)) {
		t.Error("starVisible 對越界索引應回傳 false,不應 panic")
	}

	// 版本差異(真實星系):Profile15 可見星數不應少於 Profile13(#13 核心效果)。
	s.RuleProfile = gamedata.Profile13()
	count13 := 0
	for _, v := range s.VisibleStars() {
		if v {
			count13++
		}
	}
	s.RuleProfile = gamedata.Profile15()
	count15 := 0
	for _, v := range s.VisibleStars() {
		if v {
			count15++
		}
	}
	t.Logf("NewDemoSession() 24 星星系:Profile13 可見 %d 顆,Profile15 可見 %d 顆", count13, count15)
	if count15 < count13 {
		t.Errorf("Profile15 可見星數(%d)不應少於 Profile13(%d)", count15, count13)
	}
}
