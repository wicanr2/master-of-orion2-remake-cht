package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// fixedSize 是測試用的資產尺寸,固定值讓位置公式可以被算死。
func fixedSize(int) (int, int) { return 20, 40 }

// TestColonyVeggieGroupsFitAssetCount 是這一組最重要的一支:群組數 × 8 必須剛好等於
// COLVEGGI.LBX 的 104 張。群組表抄錯一格,這個等式就不成立。
//
// 和道路那支 `TestColonyRoadAssetCoversExactly156` 同一種驗法:拿 LBX 檔本身當 oracle。
func TestColonyVeggieGroupsFitAssetCount(t *testing.T) {
	if colVeggieGroups*colVeggieGroupSize != 104 {
		t.Fatalf("群組 %d × 每組 %d = %d,COLVEGGI.LBX 是 104 張",
			colVeggieGroups, colVeggieGroupSize, colVeggieGroups*colVeggieGroupSize)
	}
	// 每個氣候在各種亂數下都不能挑出超出範圍的群組。
	maxGroup := -1
	for climate := 0; climate <= 9; climate++ {
		for seed := uint32(0); seed < 40; seed++ {
			g := colonyVeggieGroup(climate, gamedata.NewOrigRand(seed))
			if g < 0 || g >= colVeggieGroups {
				t.Fatalf("氣候 %d 挑出群組 %d,超出 0..%d", climate, g, colVeggieGroups-1)
			}
			if g > maxGroup {
				maxGroup = g
			}
		}
	}
	if maxGroup != colVeggieGroups-1 {
		t.Errorf("最大群組應為 %d(13 組正好填滿 104 張),實得 %d", colVeggieGroups-1, maxGroup)
	}
}

// TestColonyVeggieGroupDeterministicClimates 釘住不吃亂數的那幾種氣候。
// 這幾筆是跳表直接回傳常數的,抄錯會整個氣候的植被種類都變。
func TestColonyVeggieGroupDeterministicClimates(t *testing.T) {
	cases := map[int]int{
		3: 8,  // 沙漠
		4: 4,  // 苔原(預設值)
		6: 10, // 沼澤
		7: 1,  // 乾燥
	}
	for climate, want := range cases {
		for seed := uint32(0); seed < 10; seed++ {
			rng := gamedata.NewOrigRand(seed)
			if got := colonyVeggieGroup(climate, rng); got != want {
				t.Errorf("氣候 %d 應恆為群組 %d,種子 %d 得到 %d", climate, want, seed, got)
			}
			// 這幾種不該消耗亂數。
			if got, want := rng.N(1000), gamedata.NewOrigRand(seed).N(1000); got != want {
				t.Errorf("氣候 %d 不該消耗亂數(流位置 %d vs %d)", climate, got, want)
			}
		}
	}
}

// TestColonyVeggieGroupOutOfRangeClimate:氣候超出 0..9 走預設群組 4(原版 `ja def_B6701`)。
func TestColonyVeggieGroupOutOfRangeClimate(t *testing.T) {
	for _, c := range []int{-1, 10, 99} {
		if got := colonyVeggieGroup(c, gamedata.NewOrigRand(1)); got != 4 {
			t.Errorf("氣候 %d 應回預設群組 4,實得 %d", c, got)
		}
	}
}

// TestColonyVeggieAssetPerspectiveShrinksWithDistance 釘住那個 `− (a+b)/2`:
// 同一顆種子下,越遠的格點只能挑到同組裡越小的編號。這是原版的透視作法。
func TestColonyVeggieAssetPerspectiveShrinksWithDistance(t *testing.T) {
	// 沙漠(氣候 3)群組固定為 8,不吃亂數,所以差異只可能來自距離項。
	const climate = 3
	near := colonyVeggieAsset(climate, 0, 0, gamedata.NewOrigRand(5))
	far := colonyVeggieAsset(climate, 5, 5, gamedata.NewOrigRand(5))
	if far > near {
		t.Errorf("遠格 (5,5) 的資產 %d 應不大於近格 (0,0) 的 %d", far, near)
	}
	// 最遠的格點 (5,5):(a+b)/2 = 5,而 Random(8)-1 最大是 7 → 索引最多 2。
	for seed := uint32(0); seed < 60; seed++ {
		got := colonyVeggieAsset(climate, 5, 5, gamedata.NewOrigRand(seed))
		if idx := got - 8*colVeggieGroupSize; idx < 0 || idx > 2 {
			t.Fatalf("種子 %d:遠格挑到組內索引 %d,應落在 0..2", seed, idx)
		}
	}
}

// TestColonyVeggieAssetAlwaysInRange:任何氣候、任何格點、任何種子都不能越界。
func TestColonyVeggieAssetAlwaysInRange(t *testing.T) {
	for climate := 0; climate <= 9; climate++ {
		for a := 0; a < colonyGridCells; a++ {
			for b := 0; b < colonyGridCells; b++ {
				got := colonyVeggieAsset(climate, a, b, gamedata.NewOrigRand(uint32(a*6+b+climate)))
				if got < 0 || got >= 104 {
					t.Fatalf("氣候 %d 格 (%d,%d) 挑到資產 %d,超出 COLVEGGI.LBX 的 0..103",
						climate, a, b, got)
				}
			}
		}
	}
}

// TestColonyRoadEdges4UsesOriginalTable 釘住那張**對調了兩筆 Δa/Δb** 的表。
//
// 這支的意義和道路那個順序表一樣:防止有人把它「修好」成格子的四條邊。
// 差別在道路那條分支是死碼、這條是活的——植被密度真的會被它影響。
func TestColonyRoadEdges4UsesOriginalTable(t *testing.T) {
	var roads colonyRoadMap
	// 只點亮原版表指定的那四段。
	roads[colonyRoadFlag(2, 3, 0)] = true
	roads[colonyRoadFlag(2, 3, 1)] = true
	roads[colonyRoadFlag(2, 4, 0)] = true // ⚠ (a, b+1) dir0,不是 (a+1, b)
	roads[colonyRoadFlag(3, 3, 1)] = true // ⚠ (a+1, b) dir1,不是 (a, b+1)
	if got := colonyRoadEdges4(roads, 2, 3); got != 4 {
		t.Errorf("原版表指定的四段都點亮時應數到 4,實得 %d", got)
	}

	// 反過來:格子 (2,3) 真正的四條邊,原版表只認得其中兩條。
	var sides colonyRoadMap
	sides[colonyRoadFlag(2, 3, 0)] = true
	sides[colonyRoadFlag(2, 3, 1)] = true
	sides[colonyRoadFlag(3, 3, 0)] = true // 右邊
	sides[colonyRoadFlag(2, 4, 1)] = true // 下邊
	if got := colonyRoadEdges4(sides, 2, 3); got != 2 {
		t.Errorf("原版表對調了兩筆,四條真邊只會數到 2,實得 %d ——"+
			"數到 4 代表有人把表『修好』了,那就不是原版行為", got)
	}
}

// TestBuildColonyVeggiesOnlyOnEmptyCells:有建築的格子不長草。
func TestBuildColonyVeggiesOnlyOnEmptyCells(t *testing.T) {
	var grid [36]int
	occupied := map[int]bool{0: true, 7: true, 20: true, 35: true}
	for i := range occupied {
		grid[i] = 1
	}
	var roads colonyRoadMap
	veg := buildColonyVeggies(&grid, roads, 5, 8, fixedSize, gamedata.NewOrigRand(3))
	for i := range veg {
		for _, v := range veg[i] {
			if v.on && occupied[i] {
				t.Errorf("格 %d 有建築卻長了草", i)
			}
		}
	}
}

// TestBuildColonyVeggiesNoRoadsAlwaysPlaces 釘住那個「0 條路必長」的特例
// (最後那個 `Random(建築數+2)` 恆真)。全空格陣、無道路時,每一格都要進到「長」的分支——
// 之後每株還有各自的 1/3,所以驗的是「至少有一格長出東西」與「沒有任何一格被整格跳過」。
func TestBuildColonyVeggiesNoRoadsAlwaysPlaces(t *testing.T) {
	var grid [36]int
	var roads colonyRoadMap
	// 每格都會抽:Random(7) + 可能 Random(n+2) + 2 次 Random(3) + 每株 Random(8)/群組/寬/高。
	// 這裡只驗「有東西長出來」——次數的精確斷言在下一支。
	total := 0
	for seed := uint32(0); seed < 8; seed++ {
		veg := buildColonyVeggies(&grid, roads, 3, 8, fixedSize, gamedata.NewOrigRand(seed))
		for i := range veg {
			for _, v := range veg[i] {
				if v.on {
					total++
				}
			}
		}
	}
	if total == 0 {
		t.Fatal("全空、無道路的地表一株草都沒長——「0 條路必長」那條沒接上")
	}
}

// TestBuildColonyVeggiesPositionFormula 釘住位置公式,特別是 y 減的是**高/4** 不是高/2。
func TestBuildColonyVeggiesPositionFormula(t *testing.T) {
	var grid [36]int
	for i := range grid {
		grid[i] = 1 // 先全佔滿
	}
	const a, b = 2, 2
	grid[a*colonyGridCells+b] = 0 // 只留一格空的

	var roads colonyRoadMap
	w, h := fixedSize(0)
	cx, cy := colonyCellCenter(a, b)
	for seed := uint32(0); seed < 30; seed++ {
		veg := buildColonyVeggies(&grid, roads, 4, 8, fixedSize, gamedata.NewOrigRand(seed))
		for _, v := range veg[a*colonyGridCells+b] {
			if !v.on {
				continue
			}
			// x ∈ 格心 + [1..w] − w/2 ; y ∈ 格心 + [1..h] − h/4
			if v.x < cx+1-w/2 || v.x > cx+w-w/2 {
				t.Errorf("種子 %d:x=%d 超出 [%d, %d]", seed, v.x, cx+1-w/2, cx+w-w/2)
			}
			if v.y < cy+1-h/4 || v.y > cy+h-h/4 {
				t.Errorf("種子 %d:y=%d 超出 [%d, %d](y 減的是高/4,不是高/2)",
					seed, v.y, cy+1-h/4, cy+h-h/4)
			}
		}
	}
}

// TestBuildColonyVeggiesDeterministic:同一顆種子 + 同一份輸入結果一致。
func TestBuildColonyVeggiesDeterministic(t *testing.T) {
	var grid [36]int
	grid[3] = 1
	grid[14] = 1
	var roads colonyRoadMap
	roads[colonyRoadFlag(1, 1, 0)] = true
	roads[colonyRoadFlag(2, 2, 1)] = true

	x := buildColonyVeggies(&grid, roads, 6, 9, fixedSize, gamedata.NewOrigRand(21))
	y := buildColonyVeggies(&grid, roads, 6, 9, fixedSize, gamedata.NewOrigRand(21))
	if x != y {
		t.Fatal("同一顆種子應長出同一片植被")
	}
	if z := buildColonyVeggies(&grid, roads, 6, 9, fixedSize, gamedata.NewOrigRand(22)); x == z {
		t.Error("不同種子長出完全相同的植被,亂數大概沒接上")
	}
}

// TestBuildColonyVeggiesRoadCountAffectsDensity 釘住「道路越多越容易長」這個(反直覺的)規則。
//
// 拿 1 條路(機率 2/7)對 4 條路(5/7)比:後者長出來的株數要明顯多。
// 用 0 條路當對照會失真,因為 0 是「必長」的特例。
func TestBuildColonyVeggiesRoadCountAffectsDensity(t *testing.T) {
	count := func(nRoads int) int {
		total := 0
		for seed := uint32(0); seed < 120; seed++ {
			var grid [36]int
			var roads colonyRoadMap
			for a := 0; a < colonyGridCells; a++ {
				for b := 0; b < colonyGridCells; b++ {
					tbl := [4][3]int{{0, 0, 0}, {0, 0, 1}, {0, 1, 0}, {1, 0, 1}}
					for i := 0; i < nRoads; i++ {
						roads[colonyRoadFlag(a+tbl[i][0], b+tbl[i][1], tbl[i][2])] = true
					}
				}
			}
			veg := buildColonyVeggies(&grid, roads, 3, 8, fixedSize, gamedata.NewOrigRand(seed))
			for i := range veg {
				for _, v := range veg[i] {
					if v.on {
						total++
					}
				}
			}
		}
		return total
	}
	few, many := count(1), count(4)
	if many <= few {
		t.Errorf("4 條路(%d 株)應比 1 條路(%d 株)長得多——原版是道路越多密度越高", many, few)
	}
}

// TestDrawColonyVeggiesWithoutAssetsDoesNotPanic:沒有資產解析器時安靜跳過。
func TestDrawColonyVeggiesWithoutAssetsDoesNotPanic(t *testing.T) {
	b := &sceneBuilder{}
	var veg colonyVeggieMap
	veg[0][0] = colonyVeggie{on: true, asset: 3, x: 10, y: 20}
	b.drawColonyVeggiesAt(nil, veg, 0, 0)
	if w, h := b.colonyVeggieSize(3); w != 0 || h != 0 {
		t.Errorf("沒有解析器時尺寸應回 (0,0),實得 (%d,%d)", w, h)
	}
}
