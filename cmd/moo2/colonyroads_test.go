package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestColonyRoadAssetCoversExactly156 是這一組最重要的一支:資產編號的封閉式若推錯,
// 合法路段的總數就不會剛好等於 COLROADS.LBX 的 156 張,而且編號會撞號或跳號。
//
// 「156」不是我算出來的目標,是 LBX 檔本身的資產數——所以這支測試等於拿檔案當 oracle。
func TestColonyRoadAssetCoversExactly156(t *testing.T) {
	seen := map[int]int{} // asset -> 出現次數
	for dir := 0; dir < colonyRoadDirs; dir++ {
		for a := 0; a < colonyRoadPoints; a++ {
			for b := 0; b < colonyRoadPoints; b++ {
				got := colonyRoadAsset(a, b, dir)
				if !colonyRoadInRange(a, b, dir) {
					if got != -1 {
						t.Fatalf("(%d,%d,dir%d) 不合法卻回傳資產 %d", a, b, dir, got)
					}
					continue
				}
				if got < 0 || got >= 156 {
					t.Fatalf("(%d,%d,dir%d) 資產 %d 超出 COLROADS.LBX 的 0..155", a, b, dir, got)
				}
				seen[got]++
			}
		}
	}
	if len(seen) != 156 {
		t.Fatalf("合法路段應剛好用滿 156 張資產,實得 %d 張", len(seen))
	}
	for asset, n := range seen {
		if n != 1 {
			t.Fatalf("資產 %d 被 %d 個路段共用,編號應一對一", asset, n)
		}
	}
}

// TestColonyRoadDirRanges 釘住四個方向的合法範圍。這四組數字是從
// `Load_Road_List_Anims_` 的跳過條件解出來的;範圍抄錯會讓某些路段畫不出來或畫到界外。
func TestColonyRoadDirRanges(t *testing.T) {
	want := []struct {
		dir        int
		maxA, maxB int // 含
		count      int
	}{
		{0, 6, 5, 42},
		{1, 5, 6, 42},
		{2, 5, 5, 36},
		{3, 5, 5, 36},
	}
	for _, w := range want {
		n := 0
		for a := 0; a < colonyRoadPoints; a++ {
			for b := 0; b < colonyRoadPoints; b++ {
				ok := colonyRoadInRange(a, b, w.dir)
				if ok != (a <= w.maxA && b <= w.maxB) {
					t.Fatalf("dir%d (%d,%d) 合法性 = %v,應為 %v", w.dir, a, b, ok, a <= w.maxA && b <= w.maxB)
				}
				if ok {
					n++
				}
			}
		}
		if n != w.count {
			t.Errorf("dir%d 應有 %d 段,實得 %d", w.dir, w.count, n)
		}
	}
}

// TestColonyRoadOrderMatchesOriginalTable 逐格核對繪製順序表,**包含原版那個位元組的錯誤**。
//
// 這支不只驗表沒抄錯,更是把「原版有 bug」這個結論釘在程式碼裡:
// 哪天有人覺得 (3,4) 重複很礙眼順手改成 (5,4),這支會立刻炸,並在訊息裡說明為什麼不能改。
func TestColonyRoadOrderMatchesOriginalTable(t *testing.T) {
	if len(colonyRoadOrder) != 49 {
		t.Fatalf("順序表應有 49 個格點,實得 %d", len(colonyRoadOrder))
	}
	seen := map[[2]int]int{}
	for i, p := range colonyRoadOrder {
		if p[0] < 0 || p[0] > 6 || p[1] < 0 || p[1] > 6 {
			t.Fatalf("索引 %d 的格點 (%d,%d) 超出 7×7", i, p[0], p[1])
		}
		seen[p]++
	}
	// 原版真值:少 (5,4)、(3,4) 出現兩次。改掉任何一項都代表偏離原版。
	if seen[[2]int{5, 4}] != 0 {
		t.Errorf("原版表裡沒有格點 (5,4)——出現了代表有人把原版的資料錯誤『修好』了,那會與原版畫面不同")
	}
	if seen[[2]int{3, 4}] != 2 {
		t.Errorf("原版表裡 (3,4) 出現 2 次(索引 7 與 18),實得 %d 次", seen[[2]int{3, 4}])
	}
	if len(seen) != 48 {
		t.Errorf("扣掉那一個重複,相異格點應為 48,實得 %d", len(seen))
	}

	// 除了索引 7 那個手滑,其餘必須 a+b 遞減(遠 → 近的畫家演算法次序)。
	// 跳過 i==7 與 i==8:這兩次比較都以索引 7 當其中一端。
	for i := 1; i < len(colonyRoadOrder); i++ {
		if i == 7 || i == 8 {
			continue
		}
		prev := colonyRoadOrder[i-1][0] + colonyRoadOrder[i-1][1]
		cur := colonyRoadOrder[i][0] + colonyRoadOrder[i][1]
		if cur > prev {
			t.Errorf("索引 %d 的 a+b=%d 大於前一格的 %d,遠近順序壞了", i, cur, prev)
		}
	}
}

// TestColonyRoadDrawDirsFromOriginalDword 釘住 dword_B4DBD = 0x01000203 的小端解讀。
func TestColonyRoadDrawDirsFromOriginalDword(t *testing.T) {
	const orig = 0x01000203
	for i, want := range colonyRoadDrawDirs {
		got := (orig >> (8 * i)) & 0xFF
		if want != got {
			t.Errorf("繪製方向第 %d 個應為 %d(dword_B4DBD 的第 %d 個位元組),實得 %d", i, got, i, want)
		}
	}
}

// TestBuildColonyRoadsBoxesOccupiedCell 驗證「有建築的格子只會在自己的四條邊上長路」。
//
// 這是軸向抄反時最會露餡的地方:對調 a/b 之後路段會落到轉置的位置,
// 而畫面上仍然是「一堆路」——看起來一樣合理。
func TestBuildColonyRoadsBoxesOccupiedCell(t *testing.T) {
	for _, cell := range [][2]int{{0, 0}, {2, 3}, {5, 5}, {5, 0}, {0, 5}} {
		var grid [36]int
		grid[cell[0]*colonyGridCells+cell[1]] = 1
		roads := buildColonyRoads(&grid, gamedata.NewOrigRand(7))

		a, b := cell[0], cell[1]
		allowed := map[int]bool{
			colonyRoadFlag(a, b, 0):   true, // 左
			colonyRoadFlag(a, b, 1):   true, // 上
			colonyRoadFlag(a+1, b, 0): true, // 右
			colonyRoadFlag(a, b+1, 1): true, // 下
		}
		n := 0
		for i, on := range roads {
			if !on {
				continue
			}
			n++
			if !allowed[i] {
				t.Fatalf("格子 (%d,%d) 長出不屬於自己四條邊的路段(旗標索引 %d)", a, b, i)
			}
		}
		// 第一次抽必中一邊,另外兩次各半機率 → 至少 1 段、至多 3 段。
		if n < 1 || n > 3 {
			t.Errorf("格子 (%d,%d) 應長 1..3 段路,實得 %d", a, b, n)
		}
	}
}

// TestBuildColonyRoadsEmptyGridDrawsNothingAndBurnsNoRandom 驗證空格子那條死碼分支
// 既不長路、也不吃亂數——後者才是關鍵:多吃一個亂數,後面每一格都會錯開。
func TestBuildColonyRoadsEmptyGridDrawsNothingAndBurnsNoRandom(t *testing.T) {
	var empty [36]int
	rng := gamedata.NewOrigRand(3)
	roads := buildColonyRoads(&empty, rng)
	for i, on := range roads {
		if on {
			t.Fatalf("空格陣不應有任何路段,旗標 %d 卻是 true", i)
		}
	}
	// 走完 36 個空格後,亂數流必須停在原地。
	if got, want := rng.N(1000), gamedata.NewOrigRand(3).N(1000); got != want {
		t.Errorf("空格陣消耗了亂數(下一個值 %d,未動用應為 %d)", got, want)
	}
}

// TestBuildColonyRoadsConsumesThreeRandomsPerBuilding 釘住「每棟建築剛好抽三次」。
// 次數錯了,道路本身看不太出來,但**後續共用同一條流的東西會整串偏掉**。
func TestBuildColonyRoadsConsumesThreeRandomsPerBuilding(t *testing.T) {
	for _, n := range []int{1, 5, 36} {
		var grid [36]int
		for i := 0; i < n; i++ {
			grid[i] = 1
		}
		rng := gamedata.NewOrigRand(11)
		buildColonyRoads(&grid, rng)

		ref := gamedata.NewOrigRand(11)
		for i := 0; i < 3*n; i++ {
			ref.N(2)
		}
		if got, want := rng.N(1000), ref.N(1000); got != want {
			t.Errorf("%d 棟建築應消耗 %d 次亂數(流位置對不上:%d vs %d)", n, 3*n, got, want)
		}
	}
}

// TestBuildColonyRoadsDeterministic 同一顆種子 + 同一份格陣必須長出同一張路網
// (玩家反覆進出同一個殖民地畫面看到的東西要一樣)。
func TestBuildColonyRoadsDeterministic(t *testing.T) {
	var grid [36]int
	for _, i := range []int{0, 7, 8, 14, 20, 21, 35} {
		grid[i] = 1
	}
	a := buildColonyRoads(&grid, gamedata.NewOrigRand(42))
	b := buildColonyRoads(&grid, gamedata.NewOrigRand(42))
	if a != b {
		t.Fatal("同一顆種子應長出同一張路網")
	}
	if c := buildColonyRoads(&grid, gamedata.NewOrigRand(43)); a == c {
		t.Error("不同種子長出完全相同的路網,亂數大概沒接上")
	}
}

// TestBuildColonyRoadsNeverSetsDiagonals 釘住檔頭四(1):dir 2/3 在原版永遠不會被設起來。
// 哪天有人「補上對角線」讓畫面更豐富,這支會提醒他那不是原版行為。
func TestBuildColonyRoadsNeverSetsDiagonals(t *testing.T) {
	var grid [36]int
	for i := range grid {
		grid[i] = 1
	}
	roads := buildColonyRoads(&grid, gamedata.NewOrigRand(5))
	for a := 0; a < colonyRoadPoints; a++ {
		for b := 0; b < colonyRoadPoints; b++ {
			for _, dir := range []int{2, 3} {
				if roads[colonyRoadFlag(a, b, dir)] {
					t.Fatalf("(%d,%d) 的 dir%d 被設起來了;原版只寫 0,對角線那 72 張是未使用美術", a, b, dir)
				}
			}
		}
	}
}

// TestDrawColonyRoadsWithoutAssetsDoesNotPanic:沒有資產解析器時整層安靜跳過。
// (`Image.At()` 不能在遊戲迴圈外呼叫,所以像素層級只能靠截圖驗,這裡只保證不炸。)
func TestDrawColonyRoadsWithoutAssetsDoesNotPanic(t *testing.T) {
	b := &sceneBuilder{}
	var roads colonyRoadMap
	roads[colonyRoadFlag(0, 0, 0)] = true
	roads[colonyRoadFlag(6, 5, 0)] = true
	b.drawColonyRoads(nil, roads)
}
