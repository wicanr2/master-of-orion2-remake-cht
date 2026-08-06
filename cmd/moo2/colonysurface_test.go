package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// colonysurface_test.go:行星表面格點的護欄。
//
// 這組座標是從執行檔資料段抄出來的 49 個點,人工轉錄容易錯一格而不自覺——中文模式的
// 畫面看起來還是「有個菱形」,只是位置偏掉。所以這裡驗的是**結構性質**(單調、對稱、
// 格數),以及那個能對到原版美術的定錨點。

// TestColonyGridMatchesOriginalHighlight:格 (0,0) 的四角要對上 COLONY.LBX 資產 8。
//
// 那張是原版**畫好位置**的單格高亮菱形(640×480 稀疏圖),渲染出來量到的角是
// 上 (316,430) / 左 (225,461) / 右 (406,461) / 下 (316,492,超出畫布下緣被裁)。
// 表是從程式碼抽的、菱形是從美術檔渲染的——兩個獨立來源對得上,這組座標才算驗過。
func TestColonyGridMatchesOriginalHighlight(t *testing.T) {
	q := colonyCellQuad(0, 0)
	want := [4][2]int{{316, 492}, {406, 461}, {316, 430}, {225, 461}}
	if q != want {
		t.Errorf("格 (0,0) 四角 = %v,want %v(對 COLONY.LBX#8 量到的菱形)", q, want)
	}
}

// TestColonyGridIsPerspectiveDiamond:b 增加要往遠端(y 變小)、a 增加要往近端右側。
// 轉錄時把某一列貼錯位置,最容易在這裡露餡。
func TestColonyGridIsPerspectiveDiamond(t *testing.T) {
	for a := 0; a < 7; a++ {
		for b := 1; b < 7; b++ {
			if colonyGridCorners[a][b][1] >= colonyGridCorners[a][b-1][1] {
				t.Errorf("a=%d:b=%d 的 y(%d)沒有比 b=%d(%d)更遠", a, b,
					colonyGridCorners[a][b][1], b-1, colonyGridCorners[a][b-1][1])
			}
			if colonyGridCorners[a][b][0] >= colonyGridCorners[a][b-1][0] {
				t.Errorf("a=%d:b=%d 的 x(%d)沒有比 b=%d(%d)更左", a, b,
					colonyGridCorners[a][b][0], b-1, colonyGridCorners[a][b-1][0])
			}
		}
	}
	for b := 0; b < 7; b++ {
		for a := 1; a < 7; a++ {
			if colonyGridCorners[a][b][0] <= colonyGridCorners[a-1][b][0] {
				t.Errorf("b=%d:a=%d 的 x(%d)沒有比 a=%d(%d)更右", b, a,
					colonyGridCorners[a][b][0], a-1, colonyGridCorners[a-1][b][0])
			}
		}
	}
}

// TestColonySlotOrderCoversEveryCell:36 格不重不漏,而且是遠→近(a+b 遞減)。
// 遠→近是畫家演算法要的次序,順序錯了建築就會互相穿透。
func TestColonySlotOrderCoversEveryCell(t *testing.T) {
	seen := map[[2]int]bool{}
	for i, s := range colonySlotOrder {
		if s[0] < 0 || s[0] >= colonyGridCells || s[1] < 0 || s[1] >= colonyGridCells {
			t.Fatalf("第 %d 個格號 %v 超出 %d×%d", i, s, colonyGridCells, colonyGridCells)
		}
		if seen[s] {
			t.Errorf("格 %v 出現兩次(第 %d 個)", s, i)
		}
		seen[s] = true
		if i > 0 {
			prev := colonySlotOrder[i-1]
			if s[0]+s[1] > prev[0]+prev[1] {
				t.Errorf("第 %d 個 %v 比前一個 %v 更遠——順序應該是遠→近(a+b 遞減)", i, s, prev)
			}
		}
	}
	if len(seen) != colonyGridCells*colonyGridCells {
		t.Errorf("只覆蓋 %d 格,應為 %d 格", len(seen), colonyGridCells*colonyGridCells)
	}
}

// TestColonyCellAtRoundTrips:每一格的中心點打回去要打到同一格。
func TestColonyCellAtRoundTrips(t *testing.T) {
	for a := 0; a < colonyGridCells; a++ {
		for b := 0; b < colonyGridCells; b++ {
			cx, cy := colonyCellCenter(a, b)
			if cy >= moo2ScreenH {
				continue // (0,0) 那格的中心已經在畫布外,原版也是裁掉的
			}
			ga, gb := colonyCellAt(cx, cy)
			if ga != a || gb != b {
				t.Errorf("格 (%d,%d) 中心 (%d,%d) 打回 (%d,%d)", a, b, cx, cy, ga, gb)
			}
		}
	}
}

// TestColonyCellAtRejectsOutside:畫面上半部(資訊面板那一帶)不該命中任何格。
func TestColonyCellAtRejectsOutside(t *testing.T) {
	for _, p := range [][2]int{{320, 40}, {320, 200}, {10, 10}, {630, 30}} {
		if a, b := colonyCellAt(p[0], p[1]); a >= 0 {
			t.Errorf("(%d,%d) 不該落在格子上,卻回了 (%d,%d)", p[0], p[1], a, b)
		}
	}
}

// TestColonyBuildingSlotCoversEveryIndex:36 格的格號要恰好覆蓋 0..35。
// 蛇行寫錯(忘了奇數列反向)會讓某些格號重複、某些沒人用,建築就會擺到別格去。
func TestColonyBuildingSlotCoversEveryIndex(t *testing.T) {
	seen := map[int]bool{}
	for b := 0; b < colonyGridCells; b++ {
		for a := 0; a < colonyGridCells; a++ {
			s := colonyBuildingSlot(a, b)
			if s < 0 || s >= 36 {
				t.Fatalf("格 (%d,%d) 的格號 %d 超出 0..35", a, b, s)
			}
			if seen[s] {
				t.Errorf("格號 %d 重複(格 (%d,%d))", s, a, b)
			}
			seen[s] = true
		}
	}
	if len(seen) != 36 {
		t.Errorf("只用到 %d 個格號,應為 36", len(seen))
	}
	// 蛇行的定錨:第 0 列由左往右,第 1 列反過來。
	if got := colonyBuildingSlot(0, 0); got != 0 {
		t.Errorf("格 (0,0) 應為格號 0,實得 %d", got)
	}
	if got := colonyBuildingSlot(5, 0); got != 5 {
		t.Errorf("格 (5,0) 應為格號 5,實得 %d", got)
	}
	if got := colonyBuildingSlot(0, 1); got != 11 {
		t.Errorf("格 (0,1) 應為格號 11(奇數列反向),實得 %d", got)
	}
	if got := colonyBuildingSlot(5, 1); got != 6 {
		t.Errorf("格 (5,1) 應為格號 6,實得 %d", got)
	}
}

// TestEveryBuildingHasOrigID:remake 每一棟建築都要對到原版編號。
// 漏一棟不會編譯失敗,只會在畫面上少一棟房子——而那要跑起來才看得到。
func TestEveryBuildingHasOrigID(t *testing.T) {
	byID := map[int]string{}
	for _, b := range gamedata.Buildings {
		id, ok := origBuildingID[b.NameEN]
		if !ok {
			t.Errorf("%q(%s)沒有對到原版建築編號", b.NameEN, b.NameZH)
			continue
		}
		if id < 1 || id > 48 {
			t.Errorf("%q 的編號 %d 超出 1..48", b.NameEN, id)
		}
		if prev, dup := byID[id]; dup {
			t.Errorf("編號 %d 被 %q 與 %q 同時使用", id, prev, b.NameEN)
		}
		byID[id] = b.NameEN
	}
}

// TestOrigBuildingIDHasNoStrays:對照表裡不該有 remake 建築表沒有的名字
// (打錯字會變成一筆永遠用不到的項目,而測試若只檢查「每棟都有對到」是抓不出來的)。
func TestOrigBuildingIDHasNoStrays(t *testing.T) {
	known := map[string]bool{}
	for _, b := range gamedata.Buildings {
		known[b.NameEN] = true
	}
	for name := range origBuildingID {
		if !known[name] {
			t.Errorf("對照表有 %q,但 gamedata.Buildings 裡沒有這個名字", name)
		}
	}
}

// TestColonyBuildingSpriteWithinAssetCounts:算出來的 (檔, 資產) 要落在該檔真實存在的
// 資產範圍內。BLDG0..BLDG3 各 360、BLDG4 是 324(實測 lbxinfo)。
// 算式抄錯(例如 ×36 寫成 ×35)會在這裡越界。
func TestColonyBuildingSpriteWithinAssetCounts(t *testing.T) {
	counts := map[string]int{
		"bldg0.lbx": 360, "bldg1.lbx": 360, "bldg2.lbx": 360,
		"bldg3.lbx": 360, "bldg4.lbx": 324,
	}
	for _, id := range origBuildingID {
		for b := 0; b < colonyGridCells; b++ {
			for a := 0; a < colonyGridCells; a++ {
				lbxName, asset, ok := colonyBuildingSprite(id, a, b)
				if !ok {
					t.Fatalf("編號 %d 應該有圖", id)
				}
				n, known := counts[lbxName]
				if !known {
					t.Fatalf("編號 %d 指到不存在的檔 %s", id, lbxName)
				}
				if asset < 0 || asset >= n {
					t.Errorf("編號 %d 格 (%d,%d):%s 資產 %d 超出 0..%d", id, a, b, lbxName, asset, n-1)
				}
			}
		}
	}
	if _, _, ok := colonyBuildingSprite(0, 0, 0); ok {
		t.Error("編號 0 是空格,不該有圖")
	}
	if _, _, ok := colonyBuildingSprite(49, 0, 0); ok {
		t.Error("編號 49 超出 48 棟,不該有圖")
	}
}
