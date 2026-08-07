package main

// colonyveggie.go:殖民地地表的植被(原版 `sub_B6977` 產生 + `sub_B6B95` 繪製)。
//
// 原版的殖民地地表除了建築與道路,空地上還長著草木。這一層 remake 先前整個沒有,
// 所以畫面比原版空曠——尤其是開局那種只有三五棟建築的殖民地,一大片空地什麼都沒有。
//
// ============ 一、13 個群組 × 8 張 = 104 ============
//
// 植物圖分成群組,**每組固定 8 張**,同組內編號越大代表越大株。
// `Pick_Random_Veggie_Anim_Entry_For_Colony_CR_` @ 0xB6647 的算式:
//
//	資產 = 群組×8 + max(Random(8) − 1 − (a+b)/2, 0)
//
// 後面那個 `− (a+b)/2` 就是**透視**:格點越遠(a+b 越大)越容易被壓到 0,也就是最小的那株。
//
// 群組由氣候決定(`sub_B66D0`,一張 10 路跳表),最大群組編號 12 → 13 個群組。
// **13 × 8 = 104,正好是 COLVEGGI.LBX 的資產數** —— 和道路那個 156 一樣,
// 這個等式就是公式對了的確認,不必去比對圖。
//
// ============ 二、放不放:道路越多反而越容易長 ============
//
// `sub_B6977` 只處理**空**格子,先數該格四條邊上有幾段路(`sub_B67E4`,0..4),然後:
//
//	r = Random(7)
//	if (r − 2) < 道路數  → 長
//	else if 道路數 != 0  → 不長
//	else                  → Random(建築數 + 2),回傳值恆 ≥ 1 所以**一定長**
//
// 於是:0 條路 → 必長;k 條路(k>0)→ 機率 (k+1)/7。
// 「沒路的空地一定長草、有路的地方看路多寡」——0 是特例不是連續的。
//
// ⚠ 最後那個 `Random(建築數 + 2)` 的結果**永遠不會是 0**(原版 `Random(n)` 回 1..n),
// 所以那條判斷等於恆真。看起來像是想寫「建築越多越不長草」卻沒生效。
// 這裡照抄:判斷保留(它會消耗一次亂數,而亂數流是共用的)、結果一樣是「長」。
//
// 每格最多 2 株,每株各抽一次 `Random(3)`,中 1 才真的長出來。
//
// ============ 三、位置:格心 + 抖動 ============
//
//	x = 格心x + Random(寬) − 寬/2
//	y = 格心y + Random(高) − 高/4
//
// 格心用的是 `Bldg_Coords_To_Centered_Screen_Coord_` @ 0xBC866,也就是角落 (a,b) 與
// (a+1,b+1) 取平均 —— remake 的 `colonyCellCenter` 已經是同一個算式,直接用。
//
// y 減的是**高/4** 不是高/2:植物的根部落在格心稍下方,不是整株置中。
//
// ============ 四、繪製:和建築交錯,不是獨立一層 ============
//
// `Draw_Colony_Bldgs_` @ 0xBEBDC 對 36 格的遠→近順序,**每一格先畫植被再畫建築**,
// 而不是「先畫完所有植被再畫所有建築」。差別在遮擋:近處的植物要蓋住遠處的建築。
//
// ⚠ **沒模擬的部分**:原版 `sub_B6B95` 的外圈次數由呼叫端的 `bl` 決定,
// 而 `bl = (沒有格子被選取)`。正常畫面 `bl = 1` → 跑一圈、走 `sub_C5D55` 這個繪製路徑;
// 有格子被選取時 `bl = 0` → **一株都不畫**。remake 沒有「選取格子」這個狀態,
// 固定走正常路徑;另一條 `sub_C5D75` 的差別(推測是不同的貼圖模式)沒有追。

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

const (
	colVeggieLBX       = "colveggi.lbx"
	colVeggieGroupSize = 8
	colVeggieGroups    = 13 // 13×8 = 104 = COLVEGGI.LBX 的資產數
	colVeggiePerCell   = 2
)

// colonyVeggie 是一株植物。
type colonyVeggie struct {
	on    bool
	asset int
	x, y  int
}

// colonyVeggieMap 是 36 格 × 每格 2 株。
type colonyVeggieMap [36][colVeggiePerCell]colonyVeggie

// colonyRoadEdges4 是 `sub_B67E4`:數 (a,b) 這一格在四個指定位置上有幾段路。
//
// ⚠ **原版的表對調了兩筆 Δa/Δb**,照抄不修。要判斷格子 (a,b) 四條邊上有沒有路,
// 該查的是 d0@(a,b)、d1@(a,b)、d0@(a+1,b)、d1@(a,b+1);而原版表 `byte_B4DE9`
// 寫成 (0,0,0)、(0,0,1)、(0,1,0)、(1,0,1),也就是後兩筆變成往外延伸的兩段,不是對邊。
//
// 這支在原版有兩個呼叫點:「空格子被路圍住就拆一條」那條分支(整段是死碼,見 colonyroads.go
// 檔頭四),以及這裡的植被密度。**密度這條是活的**,所以這個表錯不錯會影響畫面——
// 照抄才會得到原版的畫面。
func colonyRoadEdges4(roads colonyRoadMap, a, b int) int {
	n := 0
	for _, t := range [4][3]int{{0, 0, 0}, {0, 0, 1}, {0, 1, 0}, {1, 0, 1}} {
		if roads[colonyRoadFlag(a+t[0], b+t[1], t[2])] {
			n++
		}
	}
	return n
}

// colonyVeggieGroup 是 `sub_B66D0`:氣候 → 植物群組。氣候超出 0..9 回預設群組 4。
//
// 表逐項抄自那張 10 路跳表。`gamedata.PlanetClimate` 的列舉(TOXIC=0 … GAIA=9)
// 與原版氣候欄位同序(見 colonysurface.go 地表底圖那段),所以可以直接拿來當索引。
func colonyVeggieGroup(climate int, rng *gamedata.OrigRand) int {
	const def = 4
	switch climate {
	case 0: // 有毒
		if rng.N(2) == 1 {
			return 5
		}
		return 9
	case 1, 2: // 輻射、荒蕪
		if rng.N(2) == 1 {
			return 6
		}
		return def
	case 3: // 沙漠
		return 8
	case 5: // 海洋
		switch rng.N(4) {
		case 1:
			return 7
		case 2:
			return 9
		case 3:
			return 11
		}
		return def
	case 6: // 沼澤
		return 10
	case 7: // 乾燥
		return 1
	case 8, 9: // 類地、蓋亞
		switch rng.N(4) {
		case 1:
			return 12
		case 2:
			return 3
		case 3:
			return 0
		}
		return 2
	}
	return def // 氣候 4(苔原)與範圍外
}

// colonyVeggieAsset 是 `sub_B6647`:挑一張 COLVEGGI 資產。
//
// ⚠ 亂數次序不能對調:原版**先**抽 `Random(8)`,**再**進 `sub_B66D0`(那支可能再抽一次)。
func colonyVeggieAsset(climate, a, b int, rng *gamedata.OrigRand) int {
	r := rng.N(colVeggieGroupSize)
	group := colonyVeggieGroup(climate, rng)
	idx := (r - 1) - (a+b)/2
	if idx < 0 {
		idx = 0
	}
	return group*colVeggieGroupSize + idx
}

// buildColonyVeggies 是 `sub_B6977`:依格陣與道路產生植被。
//
// nBldgs 是 `N_Bldgs_`(殖民地擁有的建築數,含軌道衛星),只用在那個恆真的判斷上——
// 但它決定 `Random(n+2)` 的參數,所以會影響亂數流,不能亂填。
// size 回傳資產的寬高(原版讀的是 LBX 標頭那兩個 word)。
//
// rng 必須是**接續道路**的同一條流。
func buildColonyVeggies(grid *[36]int, roads colonyRoadMap, nBldgs, climate int,
	size func(asset int) (int, int), rng *gamedata.OrigRand) colonyVeggieMap {

	var veg colonyVeggieMap
	for a := 0; a < colonyGridCells; a++ {
		for b := 0; b < colonyGridCells; b++ {
			if grid[a*colonyGridCells+b] != 0 {
				continue // 只有空格子長草
			}
			nRoads := colonyRoadEdges4(roads, a, b)
			place := rng.N(7)-2 < nRoads
			if !place {
				if nRoads != 0 {
					continue
				}
				// 沒路:這次抽的結果不影響判斷(恆為「長」),但要照抄以維持流。
				rng.N(nBldgs + 2)
				place = true
			}
			if !place {
				continue
			}
			cx, cy := colonyCellCenter(a, b)
			for k := 0; k < colVeggiePerCell; k++ {
				if rng.N(3) != 1 {
					continue
				}
				asset := colonyVeggieAsset(climate, a, b, rng)
				w, h := size(asset)
				if w < 1 {
					w = 1
				}
				if h < 1 {
					h = 1
				}
				veg[a*colonyGridCells+b][k] = colonyVeggie{
					on:    true,
					asset: asset,
					x:     cx + rng.N(w) - w/2,
					y:     cy + rng.N(h) - h/4,
				}
			}
		}
	}
	return veg
}

// colonyVeggieSize 回傳某張 COLVEGGI 資產的寬高;取不到回 (0,0)。
//
// 原版讀的是 LBX 標頭的寬高,不是實際著色範圍——這裡用 `lbx.Image` 的同兩個欄位。
//
// ⚠ **一定要快取**:尺寸是在「產生」階段就要用的(它進位置公式),而地表是每幀重算的,
// 不快取就變成每幀重解最多 72 張 LBX。`decodeAsset` 自己沒有快取。
func (b *sceneBuilder) colonyVeggieSize(asset int) (int, int) {
	if b.res == nil {
		return 0, 0
	}
	if wh, hit := b.colVegSizeCache[asset]; hit {
		return wh[0], wh[1]
	}
	if b.colVegSizeCache == nil {
		b.colVegSizeCache = map[int][2]int{}
	}
	wh := [2]int{}
	if im, err := decodeAsset(b.res, colVeggieLBX, asset); err == nil {
		wh = [2]int{im.Width, im.Height}
	}
	b.colVegSizeCache[asset] = wh // 取不到也快取,避免每幀重試
	return wh[0], wh[1]
}

// drawColonyVeggiesAt 畫某一格的植物。由 `drawColonyBuildings` 在每一格的建築**之前**呼叫,
// 遠→近逐格交錯——原版 `Draw_Colony_Bldgs_` 就是這樣排的(見檔頭四)。
func (b *sceneBuilder) drawColonyVeggiesAt(dst *ebiten.Image, veg colonyVeggieMap, a, bb int) {
	for _, v := range veg[a*colonyGridCells+bb] {
		if !v.on {
			continue
		}
		im := b.colonyScreenImage(colVeggieLBX, v.asset, colonyBasePalette)
		if im == nil {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(v.x), float64(v.y))
		dst.DrawImage(im, op)
	}
}
