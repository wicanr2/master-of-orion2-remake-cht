package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestFindColonyBuildingReturnsFirstInScanOrder:同一種房屋佔好幾格時,取掃到的第一格
// (a 外 b 內)。取錯格會讓抖動搬到別的房子,而畫面上仍然是「房子換了位置」——看不出來。
func TestFindColonyBuildingReturnsFirstInScanOrder(t *testing.T) {
	var grid [36]int
	for i := range grid {
		grid[i] = 99 // 填滿,避免空格抽亂數干擾這支的斷言
	}
	grid[2*colonyGridCells+5] = 3 // (2,5)
	grid[4*colonyGridCells+1] = 3 // (4,1)

	a, b, ok := findColonyBuilding(&grid, 3, gamedata.NewOrigRand(1))
	if !ok {
		t.Fatal("應找到編號 3")
	}
	if a != 2 || b != 5 {
		t.Errorf("應取掃描序在前的 (2,5),實得 (%d,%d)", a, b)
	}
}

// TestFindColonyBuildingConsumesRandomPerEmptyCellBeforeHit 釘住「找建築會吃亂數,
// 而且吃幾次看資料」。這是最容易漏掉的一條:漏了之後道路整串偏掉,而畫面看起來照樣正常。
func TestFindColonyBuildingConsumesRandomPerEmptyCellBeforeHit(t *testing.T) {
	// (0,0)..(0,3) 空、(0,4) 是目標 → 命中前掃到 4 個空格 → 抽 4 次。
	var grid [36]int
	for i := range grid {
		grid[i] = 99
	}
	for b := 0; b < 4; b++ {
		grid[b] = 0
	}
	grid[4] = 3

	rng := gamedata.NewOrigRand(9)
	if a, b, ok := findColonyBuilding(&grid, 3, rng); !ok || a != 0 || b != 4 {
		t.Fatalf("應在 (0,4) 找到,實得 (%d,%d) ok=%v", a, b, ok)
	}
	ref := gamedata.NewOrigRand(9)
	for i := 1; i <= 4; i++ {
		ref.N(i)
	}
	if got, want := rng.N(1000), ref.N(1000); got != want {
		t.Errorf("命中前的 4 個空格應各抽一次亂數(流位置對不上:%d vs %d)", got, want)
	}
}

// TestFindColonyBuildingMinusOneReturnsRandomEmptyCell:id == -1 是原版用來「找空位」的
// 路徑,掃完沒找到反而算成功,回傳蓄水池抽到的空格。
func TestFindColonyBuildingMinusOneReturnsRandomEmptyCell(t *testing.T) {
	var grid [36]int
	for i := range grid {
		grid[i] = 99
	}
	grid[3*colonyGridCells+2] = 0 // 唯一的空格 (3,2)

	a, b, ok := findColonyBuilding(&grid, -1, gamedata.NewOrigRand(4))
	if !ok {
		t.Fatal("id == -1 掃完應算成功")
	}
	if a != 3 || b != 2 {
		t.Errorf("應回傳唯一的空格 (3,2),實得 (%d,%d)", a, b)
	}
}

// TestFindColonyBuildingMissingReturnsFalse:一般編號找不到就是找不到。
func TestFindColonyBuildingMissingReturnsFalse(t *testing.T) {
	var grid [36]int
	for i := range grid {
		grid[i] = 99
	}
	if _, _, ok := findColonyBuilding(&grid, 3, gamedata.NewOrigRand(2)); ok {
		t.Error("格陣裡沒有編號 3,不該回報找到")
	}
}

// TestJitterColonyHousesTargetsDiagonalOnly 釘住原版那個 bug:換位對象永遠落在主對角線上。
//
// 這支存在的理由是**防止有人把它「修好」**。看到 `target := v*6 + v` 很容易覺得是打錯,
// 改成 (a-si, b-di) 之後畫面依然合理——但那就不是原版的畫面了。
func TestJitterColonyHousesTargetsDiagonalOnly(t *testing.T) {
	const start = 1*colonyGridCells + 4 // (1,4),刻意不在對角線上
	moved := 0
	for seed := uint32(0); seed < 60; seed++ {
		var grid [36]int
		grid[start] = colonyHouseStyles[0]
		jitterColonyHouses(&grid, gamedata.NewOrigRand(seed))

		pos := -1
		for i, id := range grid {
			if id == colonyHouseStyles[0] {
				if pos >= 0 {
					t.Fatalf("種子 %d:房屋被複製成兩份", seed)
				}
				pos = i
			}
		}
		if pos < 0 {
			t.Fatalf("種子 %d:房屋不見了", seed)
		}
		if pos == start {
			continue
		}
		moved++
		if a, b := pos/colonyGridCells, pos%colonyGridCells; a != b {
			t.Fatalf("種子 %d:房屋搬到 (%d,%d),不在主對角線上——原版的目標格是 v*6+v", seed, a, b)
		}
	}
	if moved == 0 {
		t.Error("60 顆種子跑下來一次都沒搬動,抖動大概沒接上")
	}
}

// TestJitterColonyHousesPreservesContents:抖動只做交換,不新增也不刪除。
func TestJitterColonyHousesPreservesContents(t *testing.T) {
	base := [36]int{}
	for _, c := range []struct{ cell, id int }{
		{0, 1}, {3, 2}, {7, 3}, {11, 14}, {20, 40}, {29, 41}, {35, 5},
	} {
		base[c.cell] = c.id
	}
	for seed := uint32(0); seed < 40; seed++ {
		grid := base
		jitterColonyHouses(&grid, gamedata.NewOrigRand(seed))

		count := func(g [36]int) map[int]int {
			m := map[int]int{}
			for _, id := range g {
				m[id]++
			}
			return m
		}
		before, after := count(base), count(grid)
		for id, n := range before {
			if after[id] != n {
				t.Fatalf("種子 %d:編號 %d 的數量由 %d 變成 %d", seed, id, n, after[id])
			}
		}
	}
}

// TestJitterColonyHousesDeterministic:同一顆種子 + 同一份格陣結果必須一致。
func TestJitterColonyHousesDeterministic(t *testing.T) {
	base := [36]int{}
	base[4] = 3
	base[9] = 14
	base[22] = 40
	base[31] = 41

	a, b := base, base
	jitterColonyHouses(&a, gamedata.NewOrigRand(77))
	jitterColonyHouses(&b, gamedata.NewOrigRand(77))
	if a != b {
		t.Fatal("同一顆種子應得到同一份格陣")
	}
	c := base
	jitterColonyHouses(&c, gamedata.NewOrigRand(78))
	if a == c {
		t.Error("不同種子結果完全相同,亂數大概沒接上")
	}
}

// TestJitterColonyHousesNoHousesIsNoOp:沒有房屋時格陣不動(但仍會掃描、吃亂數——
// 那是原版行為,道路接在後面所以不能省)。
func TestJitterColonyHousesNoHousesIsNoOp(t *testing.T) {
	var grid [36]int
	grid[0] = 1 // 一棟一般建築,不在 colonyHouseStyles 裡
	before := grid

	rng := gamedata.NewOrigRand(6)
	jitterColonyHouses(&grid, rng)
	if grid != before {
		t.Error("沒有房屋可搬,格陣不該變動")
	}
	// 8 輪各掃過全格陣,35 個空格每格抽一次 → 共 8×35 次。
	ref := gamedata.NewOrigRand(6)
	for r := 0; r < colonyJitterRounds; r++ {
		for i := 1; i <= 35; i++ {
			ref.N(i)
		}
	}
	if got, want := rng.N(1000), ref.N(1000); got != want {
		t.Errorf("掃描消耗的亂數次數對不上(%d vs %d)", got, want)
	}
}
