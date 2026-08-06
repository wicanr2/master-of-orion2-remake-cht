package main

import "testing"

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
