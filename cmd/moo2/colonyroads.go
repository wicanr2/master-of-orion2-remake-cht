package main

// colonyroads.go:殖民地地表的道路(原版 `Build_Road_List_Based_On_Bldg_List_` @ 0xB6099
// + `Draw_Road_List_` @ 0xB6255 + `Load_Road_List_Anims_` @ 0xB5FBE)。
//
// 原版的殖民地地表,建築之間有一張道路網把房子串起來。remake 先前只畫地形 + 建築,
// 少了這一層,畫面看起來像「房子撒在草地上」而不是聚落。
//
// ============ 一、道路長在格點上,不是格子裡 ============
//
// 建築佔的是 6×6 的**格子**;道路走的是 7×7 的**格點**,格點之間的連線就是路段。
// 每個格點掛 4 個方向,記錄用的是同一組陣列(步距 `0x1D` = 29 一個格點、`0xCB` = 203 = 7×29 一列):
//
//	資產編號 word_19E574[a×203 + b×29 + dir×2]     ; 0xFFFF = 這個方向不存在
//	是否畫   byte_19E57C[a×203 + b×29 + dir]       ; 旗標,0/1
//
// 四個方向的幾何由 `Load_Road_List_Anims_` 的**合法範圍**決定(超出範圍就寫 0xFFFF):
//
//	dir 0:a 0..6、b 0..5 → 格點 (a,b) 到 (a,b+1) 的邊    42 段
//	dir 1:a 0..5、b 0..6 → 格點 (a,b) 到 (a+1,b) 的邊    42 段
//	dir 2:a,b 0..5       → 格子 (a,b) 的一條對角線        36 段
//	dir 3:a,b 0..5       → 格子 (a,b) 的另一條對角線      36 段
//
// 加起來 42+42+36+36 = **156**,與 COLROADS.LBX 的資產數一模一樣——這就是幾何解對了的確認。
//
// ⚠ 這裡的 a 是 **203 步距**那一軸(= 建築格陣 `colony_bldgs[24×a + 4×b]` 的 a),
// b 是 29 步距那一軸。兩軸對調畫面會整張鏡射而且**看起來一樣合理**,和建築格陣同一個坑。
//
// ============ 二、資產編號是連號分配的 ============
//
// `Load_Road_List_Anims_` 用一個從 0 開始的計數器,依 dir → a → b 的巢狀順序,
// 每碰到一個合法方向就發下一號。所以編號是四段連續區間:
//
//	dir 0 → 0   + a×6 + b
//	dir 1 → 42  + a×7 + b
//	dir 2 → 84  + a×6 + b
//	dir 3 → 120 + a×6 + b
//
// 每張都是 640×480 的**預先定位稀疏圖**(和 BLDGn / PLANETS 同一種),貼在 (0,0) 就位,
// 不需要自己算座標。
//
// ============ 三、產生規則:接在建築擺放的同一條亂數流上 ============
//
// `Build_Road_List_Based_On_Bldg_List_` 由 `Make_Bldg_Array_For_Colony_` 在 `loc_BC5B4` 呼叫,
// 也就是**建築擺放 → 排序 → 抖動之後**,共用 `Set_Random_Seed(colonyIdx)` 起的那條流。
// 對每個有建築的格子 (a,b),抽三次 `Random(2)`:
//
//	r1:0 → 設 dir0@(a,b)      否則 → 設 dir1@(a,b)      ; 左邊或上邊,二選一
//	r2:非 0 → 設 dir0@(a+1,b)                          ; 右邊,一半機率
//	r3:非 0 → 設 dir1@(a,b+1)                          ; 下邊,一半機率
//
// 這四段正好是格子 (a,b) 的四條邊,所以視覺上是「房子外圍畫框,框缺一到三邊」。
//
// ============ 四、兩件從原版讀出來的事實,不要當成 bug 修掉 ============
//
// **(1) dir 2 / dir 3 永遠不會被畫。** 全執行檔對 `byte_19E57E` / `byte_19E57F`
// (dir 2、dir 3 的旗標)只有**寫 0**,沒有任何一處寫 1(IDA 的 DATA XREF 也只記錄
// `sub_B6099` 裡那兩個 `mov …, 0`)。也就是說 COLROADS.LBX 裡那 72 張對角線圖
// 在出貨版是**沒被用到的美術**。連帶讓產生器裡「空格子」那一整條分支變成死碼:
// 它先要求 dir2/dir3 非 0 才往下走,而那永遠不成立。remake 因此只實作有建築那條分支——
// 這不是簡化,是把死碼認出來之後不抄。
//
// **(2) 繪製順序表少一個格點。** 見 `colonyRoadOrder`。
//
// ============ 五、⚠ 位置不會與原版同一局逐格相同 ============
//
// 道路吃的是建築擺放之後的亂數流位置,而 remake **還沒實作** `Sort_Bldg_Array_Columns_`
// 之後那段房屋抖動(8 輪、約 1/3 機率換鄰格,見 `colonySurfacePlan` 檔頭)。
// 那段會消耗亂數,所以到了道路這一步,兩邊的流早就錯開了。
//
// 於是:**美術、格點、方向規則、密度都是原版真值,但同一顆星的道路走法會與原版不同。**
// 要逐格對上,前置是把抖動那段補完——不是這一檔的問題。

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

const (
	colRoadsLBX = "colroads.lbx"
	// colonyRoadPoints 是格點數(7×7),比建築格子多一排。
	colonyRoadPoints = colonyGridCells + 1
	colonyRoadDirs   = 4
)

// colonyRoadMap 是 49 個格點 × 4 個方向的旗標,對應原版 `byte_19E57C`。
type colonyRoadMap [colonyRoadPoints * colonyRoadPoints * colonyRoadDirs]bool

// colonyRoadFlag 算旗標索引。原版是 `a×203 + b×29 + dir`,29 是格點的位元組步距;
// 這裡只存 bool,所以壓成 `(a×7 + b)×4 + dir`——同構,少掉那 25 個位元組的洞。
func colonyRoadFlag(a, b, dir int) int {
	return (a*colonyRoadPoints+b)*colonyRoadDirs + dir
}

// colonyRoadInRange 回傳 (a,b,dir) 是不是合法路段(見檔頭一)。
func colonyRoadInRange(a, b, dir int) bool {
	if a < 0 || b < 0 || dir < 0 || dir >= colonyRoadDirs {
		return false
	}
	switch dir {
	case 0:
		return a < colonyRoadPoints && b < colonyGridCells
	case 1:
		return a < colonyGridCells && b < colonyRoadPoints
	default:
		return a < colonyGridCells && b < colonyGridCells
	}
}

// colonyRoadAsset 回傳 COLROADS.LBX 的資產編號,不合法回 -1(= 原版的 0xFFFF)。
func colonyRoadAsset(a, b, dir int) int {
	if !colonyRoadInRange(a, b, dir) {
		return -1
	}
	switch dir {
	case 0:
		return a*colonyGridCells + b
	case 1:
		return 42 + a*colonyRoadPoints + b
	case 2:
		return 84 + a*colonyGridCells + b
	default:
		return 120 + a*colonyGridCells + b
	}
}

// colonyRoadOrder 是 49 個格點的**遠 → 近**繪製順序,逐位元組抄自原版 `byte_B4D5B`
// (`Road_Back_To_Front_` @ 0xB6227 從這裡取 (b, a),⚠ 表裡是先 b 後 a)。
//
// ⚠ **原版這張表有一個位元組是錯的,這裡照抄不修。** 排序原則是 a+b 由大到小
// (遠的先畫),但索引 7 是 (3,4)——a+b=7,夾在兩個 9 中間;而 (5,4) 整張表都沒出現,
// (3,4) 反而出現兩次(索引 7 和 18)。把那一個位元組從 0x03 改成 0x05,49 格就剛好
// 不重不漏、a+b 完全遞減。這是原版建表時手滑,不是這裡解錯:
//
//   - 表的起點與長度都釘死了——緊接在後面的 4 個位元組就是 `dword_B4DBD`(繪製方向順序),
//     值 0x01000203 與反組譯讀到的完全一致。
//   - 用「49 格不重不漏 + a+b 遞減」當指紋掃整個執行檔,**零命中**;改用前 14 個位元組
//     當錨點才掃到,而且全檔唯一。
//
// 後果:格點 (5,4) 的路段永遠不會被畫,(3,4) 被畫兩次(同一張圖疊兩次,看不出差別)。
// 也就是原版畫面上那個位置本來就有個缺口。修掉它會讓 remake 比原版「正確」而與原版不同。
var colonyRoadOrder = [colonyRoadPoints * colonyRoadPoints][2]int{
	{6, 6}, {6, 5}, {5, 6}, {6, 4}, {5, 5}, {4, 6}, {6, 3},
	{3, 4}, {4, 5}, {3, 6}, {6, 2}, {5, 3}, {4, 4}, {3, 5},
	{2, 6}, {6, 1}, {5, 2}, {4, 3}, {3, 4}, {2, 5}, {1, 6},
	{6, 0}, {5, 1}, {4, 2}, {3, 3}, {2, 4}, {1, 5}, {0, 6},
	{5, 0}, {4, 1}, {3, 2}, {2, 3}, {1, 4}, {0, 5}, {4, 0},
	{3, 1}, {2, 2}, {1, 3}, {0, 4}, {3, 0}, {2, 1}, {1, 2},
	{0, 3}, {2, 0}, {1, 1}, {0, 2}, {1, 0}, {0, 1}, {0, 0},
}

// colonyRoadDrawDirs 是四個方向的繪製順序,來自 `dword_B4DBD` = 0x01000203
// (小端 → 位元組 03 02 00 01)。dir 2/3 從來沒被設起來(見檔頭四),所以實際只有 0、1 會畫;
// 順序照抄是為了將來若補上對角線時不必再回頭查。
var colonyRoadDrawDirs = [colonyRoadDirs]int{3, 2, 0, 1}

// buildColonyRoads 由建築格陣推出道路,消耗與原版相同的亂數次數與順序。
//
// grid 是 `colonySurfacePlan` 排序後的 36 格(0 = 空),rng 必須是**接續建築擺放**的同一條流。
func buildColonyRoads(grid *[36]int, rng *gamedata.OrigRand) colonyRoadMap {
	var roads colonyRoadMap
	set := func(a, b, dir int) {
		if colonyRoadInRange(a, b, dir) {
			roads[colonyRoadFlag(a, b, dir)] = true
		}
	}
	// 迴圈巢狀與原版一致:外層是 203 步距那一軸(edi),內層是 29 步距那一軸(esi)。
	for a := 0; a < colonyGridCells; a++ {
		for b := 0; b < colonyGridCells; b++ {
			if grid[a*colonyGridCells+b] == 0 {
				continue // 空格子那條分支是死碼,見檔頭四(1)。也不吃亂數。
			}
			if rng.N(2)-1 == 0 {
				set(a, b, 0)
			} else {
				set(a, b, 1)
			}
			if rng.N(2)-1 != 0 {
				set(a+1, b, 0)
			}
			if rng.N(2)-1 != 0 {
				set(a, b+1, 1)
			}
		}
	}
	return roads
}

// drawColonyRoads 依原版的遠→近格點順序把路段貼上去。
//
// 位置在**地形之後、框架之前**:`Draw_Colony_Screen_` @ 0xBED21 的呼叫序是
// 填色 → C_Anims(1) 天空 → C_Anims(0) 地形 → `Draw_Road_List_` → …框架… → 建築。
func (b *sceneBuilder) drawColonyRoads(dst *ebiten.Image, roads colonyRoadMap) {
	for _, p := range colonyRoadOrder {
		for _, dir := range colonyRoadDrawDirs {
			if !roads[colonyRoadFlag(p[0], p[1], dir)] {
				continue
			}
			asset := colonyRoadAsset(p[0], p[1], dir)
			if asset < 0 {
				continue
			}
			if im := b.colonyScreenImage(colRoadsLBX, asset, colonyBasePalette); im != nil {
				dst.DrawImage(im, &ebiten.DrawImageOptions{})
			}
		}
	}
}
