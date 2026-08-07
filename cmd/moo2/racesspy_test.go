package main

import "testing"

// 種族關係畫面的間諜區塊座標來自**執行檔的 .data 靜態表**,不是量圖。
//
// 這張畫面在 openorion2 是 STUB,所以 remake 先前只能量圖(自編的「y 從 66 起、每列 +62」)。
// 現在有更上游的來源,而來源優先序是專案硬規則——這一支擋住「改回自編排版」。
func TestRacesSpyAnchorsComeFromTheExecutable(t *testing.T) {
	want := [racesMaxRows][2]int{
		{120, 126}, {120, 233}, {120, 338}, {120, 445},
		{332, 126}, {332, 233}, {332, 338},
	}
	if racesSpyAnchors != want {
		t.Errorf("間諜鈕錨點應是 %v,得到 %v", want, racesSpyAnchors)
	}
}

// 版面的形狀:左欄 4 列 + 右欄 3 列,兩欄 y 完全對齊。
//
// 這是那張表最顯眼的結構,抄錯任何一筆都會破——比逐筆比對更能抓到「順序弄反」這種錯。
func TestRacesLayoutIsFourLeftThreeRightWithSharedYs(t *testing.T) {
	const leftX, rightX = 120, 332
	for i := 0; i < 4; i++ {
		if racesSpyAnchors[i][0] != leftX {
			t.Errorf("第 %d 列應在左欄 x=%d,得到 %d", i, leftX, racesSpyAnchors[i][0])
		}
	}
	for i := 4; i < racesMaxRows; i++ {
		if racesSpyAnchors[i][0] != rightX {
			t.Errorf("第 %d 列應在右欄 x=%d,得到 %d", i, rightX, racesSpyAnchors[i][0])
		}
	}
	// 右欄三列的 y 與左欄前三列相同。
	for i := 0; i < 3; i++ {
		if racesSpyAnchors[i][1] != racesSpyAnchors[i+4][1] {
			t.Errorf("左欄第 %d 列 y=%d 應與右欄第 %d 列 y=%d 對齊",
				i, racesSpyAnchors[i][1], i+4, racesSpyAnchors[i+4][1])
		}
	}
}

// 關係滑桿也是同一種兩欄版面(x=105 / 528)。
func TestRacesRelationBarsAreTwoColumns(t *testing.T) {
	for i := 0; i < 4; i++ {
		if racesRelationBars[i][0] != 105 {
			t.Errorf("第 %d 條滑桿應在 x=105,得到 %d", i, racesRelationBars[i][0])
		}
	}
	for i := 4; i < racesMaxRows; i++ {
		if racesRelationBars[i][0] != 528 {
			t.Errorf("第 %d 條滑桿應在 x=528,得到 %d", i, racesRelationBars[i][0])
		}
	}
}

// 熱區數量跟著 AI 數走,並夾在版面上限(原版最多同時顯示 7 個已接觸種族)。
func TestRacesSpyHitRegionsClampToTheLayout(t *testing.T) {
	for _, n := range []int{0, 1, 3, 7} {
		if got := len(racesSpyHitRegions(n)); got != n {
			t.Errorf("%d 個 AI 應有 %d 個熱區,得到 %d", n, n, got)
		}
	}
	// 超過版面上限要夾住,不能越界索引那張 7 筆的表。
	if got := len(racesSpyHitRegions(99)); got != racesMaxRows {
		t.Errorf("超過上限應夾在 %d,得到 %d", racesMaxRows, got)
	}
}

// 動作字串 round-trip,而且不會把別的動作誤判成間諜。
func TestRacesSpyActionRoundTrips(t *testing.T) {
	for i := 0; i < racesMaxRows; i++ {
		got, ok := racesSpyActionIndex(racesSpyAction(i))
		if !ok || got != i {
			t.Errorf("第 %d 個動作 round-trip 失敗:得到 (%d,%v)", i, got, ok)
		}
	}
	// 這張畫面上其他動作不能被當成間諜——誤判會讓「宣戰」變成派間諜。
	for _, a := range []string{"audience", "declarewar", "report", "back", "spy", "spyX", "spy99"} {
		if _, ok := racesSpyActionIndex(a); ok {
			t.Errorf("%q 不該被判成間諜動作", a)
		}
	}
}
