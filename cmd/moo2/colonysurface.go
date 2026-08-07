package main

import (
	"fmt"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// colonysurface.go:殖民地畫面的**行星表面格點**(原版 module 74 的 `CR_To_XY_` 那一套)。
//
// ============ 這一塊先前被判定為「卡住」,卡的原因是找錯了 ============
//
// `docs/re/01-gap-report.md` 第 26 項原本寫:格點→螢幕座標是「烘在資料段的幾何表,
// 沒有公式可抄」,所以「要做這一塊得先把那幾張表從執行檔的資料段抽出來」,並把它列為
// 獨立工程。**表確實是烘死的,但它就在反組譯的資料段裡,直接讀得到**——下面那 49 個
// 座標點就是抽出來的結果。真正的教訓是:發現「這是查表不是公式」時,下一步是**去讀那張表**,
// 不是把它記成阻塞。
//
// ============ 一手來源(原版執行檔反組譯)============
//
//	`CR_To_XY_` @ 0xBC5D8:
//	    imul ecx, eax, 38h        ; a × 56
//	    movsx eax, dx / shl eax,3 ; b × 8
//	    movsx ecx, bx             ; c
//	    mov ax, word_182C9C[eax+ecx*4]
//	  → 表基底 0x182C9C,索引 = a×56 + b×8 + c×4,取 16-bit。
//	    c=0 取 x、c=1 取 y(見下面 `Get_Bldg_Box_` 的用法)。
//	    表長 392 位元組 = 7×56 → **7×7 個角點**,每格 8 位元組 =(x, y)兩個 32-bit。
//
//	`Get_Bldg_Box_` @ 0xBC5F2:對 (A,B) 連呼 8 次 `CR_To_XY_`,
//	    c=0 那四次寫進 esi[0..3]、c=1 那四次寫進 ecx[0..3],
//	    角點順序 (A,B) → (A+1,B) → (A+1,B+1) → (A,B+1)。
//	  → 一格的四個角;也證實 c=0 是 X 陣列、c=1 是 Y 陣列。
//
//	`Bldg_Coords_To_Centered_Screen_Coord_` @ 0xBC866:
//	    (dword_182E24[a×56+b×8+c×4] + dword_182E2C[(a+1)×56+b×8+c×4]) / 2
//	    dword_182E2C = dword_182E24 + 8(差一個 b 槽),而 dword_182E24 是 word_182C9C 的
//	    **完整副本**(內容逐值相同,執行檔裡放了兩份)。
//	  → 結果 = (角點[a][b] + 角點[a+1][b+1]) / 2 = **該格的中心**。
//
//	`word_182FAC`:第三張表,是同一組角點去掉第一列與第一欄的 6×6 版本(步距 48)。
//	  remake 用不到——有 7×7 就推得出來。
//
// ============ 獨立驗證(這一步別跳過)============
//
// `COLONY.LBX` 資產 8 是**單格的高亮菱形**,而且是**畫好位置的 640×480 圖**。把它渲染出來,
// 菱形的四個角落在 (316,430) 上 / (225,461) 左 / (406,461) 右 / 底邊超出畫布下緣——
// 與下面 `colonyGridCorners` 算出的格 (0,0) 四角**完全吻合**。
// 表是從程式碼抽的、菱形是從美術檔渲染的,兩個獨立來源對上,才敢說這組座標是對的。
//
// ============ 順帶推翻的一個版面斷言 ============
//
// `colonyscreen.go` 檔頭寫「中段(y 159..423)原版畫行星表面」。**行星表面不是被關在中段那個
// 框裡的**:`Draw_Colony_Screen_` 一開場就是 `C_Anims(1, 0, 639, 479)` + `Draw(0,0,…)`,
// 地表是**整個 640×480 的底圖**,資訊面板是疊在它上面(所以 COLPUPS#5 的面板才是不透明的)。
// 下面的格點 y 範圍 288..492、x 範圍 −21..666 也印證這點:菱形比畫面還寬,左右兩角是被裁掉的。
//
// ============ 建築 sprite:不需要這組座標也能畫 ============
//
// `BLDG0..BLDG4.LBX` 每個資產都是 **640×480、已經畫好位置**的稀疏圖(實測 BLDG0 資產 0..35
// 的墨點依序沿著格點推進,資產 36 起重新回到近端 → 每 36 個一循環)。所以
//
//	資產編號 = 建築型別 × 36 + 格號
//
// 畫的時候直接貼 (0,0) 就對位,`Draw_Building_With_Bottom_Centered_` 那套底邊置中是原版
// **產生**這些圖時用的,不是執行期需要的。資產數 360/360/360/360/324 → 10+10+10+10+9 = **49 種**。
//
// ============ 建築編號 → 圖檔:`Cache_Load_Bldg_` @ 0xAF6DC 給出全部算式 ============
//
//	dec ebx                     ; 建築編號是 **1-based**(0 = 空格)
//	mov eax, ebx / idiv 10      ; 檔號 = (id−1) / 10
//	mov eax, 0C9h / call E_Strings_ / sprintf_   ; 格式字串 → "BLDG%d.LBX"
//	mov eax, ebx / idiv 10 / imul ebx, edx, 24h  ; 檔內型別 = (id−1) % 10,再 ×36
//	call sub_BC8A6              ; (a,b) → 格號
//	add edx, ebx                ; 資產 = 檔內型別×36 + 格號
//
// `sub_BC8A6`(格號)也是一張乾淨的算式,不是表:
//
//	slot(a,b) = b×6 + a      (b 偶數)
//	slot(a,b) = b×6 + 5 − a  (b 奇數)
//
// 也就是**蛇行**(一列由左往右、下一列反過來),與實測 BLDG0 資產 0..35 墨點的走法一致。
//
// ============ 建築編號本身:兩個獨立來源對上才敢用 ============
//
//	① openorion2 `src/gamestate.h` 的 `BUILDING_*` 列舉:`BUILDING_NONE = 0` 起,48 棟。
//	② 原版 `TECHNAME.LBX` 資產 0 的第 295 條起:"No Building"、"Alien Control Center"、
//	   "Armor Barracks"…"Artificial Planet" —— **逐條與 ① 同序**
//	   (openorion2 `src/lang.h` 也寫著 `TNAME_BUILDING_NONE 295`)。
//
// 一個是別人重製專案的列舉、一個是原版資料檔的字串順序,對得起來才算數。
// 48 棟 → 型別 0..47,而圖檔共 49 種(360/360/360/360/324),最後一種沒有對應建築。
//
// ============ 軸向:格陣的外層索引就是角點表的第一維 ============
//
// 這一條**極容易弄反而不自覺**——轉置之後畫面仍然是一片合理的透視格,建築也還在格線上,
// 只是整張佈局沿對角線鏡射掉。第一版就是這樣錯的(建築全擠在遠端那幾排)。
// 定案證據在 `Add_Bldg_Fields_` @ 0xBE44A,同一對 (v1, v2) 同時餵給兩邊:
//
//	colony_bldgs[24×v1 + 4×v2]                          ; 格陣元素 → v1 是「列」
//	Bldg_Coords_To_Centered_Screen_Coord(v1, v2, c)     ; 螢幕座標 → 索引 v1×56 + v2×8 + c×4
//
// 所以 **格陣 grid[a×6 + b] ↔ 角點 colonyGridCorners[a][b]**,a 是步距 56 那一維。
// `Insert_Bldg_Into_Array_` 的抽樣迴圈(外 ebx ×0x18、內 edx ×4)與
// `Sort_Bldg_Array_Columns_` 的雙層迴圈也都是「外層 = a」。
//
// ⚠ 格號的算式反過來:`sub_BC8A6(eax=a, edx=b)` → `b×6 + a`(b 為奇數時 `b×6 + 5 − a`)。
// 格陣索引是 a×6+b、格號是 b×6+a,兩者**不是同一個數**,別互相代用。

// colonyGridCorners 是 7×7 角點的螢幕座標,直接抄自 word_182C9C(見檔頭)。
// 第一維是 a(反組譯裡步距 56 的那一維),第二維是 b(步距 8)。
//
// 排出來是個透視菱形:a 增加往右下、b 增加往左上;(0,0) 在近端(畫面下方,底角 y=492
// 已超出 480 被裁掉),(6,6) 在遠端(y=288)。左右兩角 x=−21 / 666 同樣超出畫布——
// 原版本來就是這樣裁的,不是抄錯。
var colonyGridCorners = [7][7][2]int{
	{{316, 492}, {225, 461}, {153, 428}, {96, 401}, {50, 379}, {10, 361}, {-21, 348}},
	{{406, 461}, {316, 430}, {242, 404}, {182, 381}, {130, 363}, {88, 346}, {53, 333}},
	{{480, 430}, {391, 403}, {317, 381}, {255, 362}, {204, 348}, {158, 334}, {121, 322}},
	{{539, 403}, {454, 382}, {380, 363}, {319, 348}, {264, 335}, {219, 322}, {180, 312}},
	{{587, 382}, {505, 363}, {434, 348}, {372, 334}, {318, 323}, {271, 312}, {232, 303}},
	{{629, 364}, {549, 349}, {479, 335}, {419, 323}, {367, 314}, {318, 303}, {277, 295}},
	{{666, 351}, {586, 336}, {517, 324}, {459, 313}, {409, 304}, {360, 295}, {320, 288}},
}

// colonyGridCells 是每邊的格數(7×7 角點 → 6×6 = 36 格)。
const colonyGridCells = 6

// colonySlotOrder 是原版建立建築欄位時走訪 36 格的順序,抄自 `Add_Bldg_Fields_` @ 0xBE44A
// 裡 `qmemcpy(v6, dword_BA784, 72)` 那 72 個位元組(36 組 (a,b))。
//
// 順序是 a+b **由大到小**——也就是從遠端往近端。畫重疊的建築時這正是畫家演算法要的次序
// (遠的先畫、近的後畫才會正確遮擋)。`Add_Bldg_Fields_` 自己是倒著跑
// (`for i = 70; i != -2; i -= 2`),因為熱區要讓近端的優先命中。
var colonySlotOrder = [36][2]int{
	{5, 5}, {5, 4}, {4, 5}, {5, 3}, {3, 5}, {4, 4},
	{5, 2}, {2, 5}, {4, 3}, {3, 4}, {5, 1}, {1, 5},
	{4, 2}, {2, 4}, {3, 3}, {5, 0}, {0, 5}, {4, 1},
	{1, 4}, {2, 3}, {3, 2}, {4, 0}, {0, 4}, {3, 1},
	{1, 3}, {2, 2}, {3, 0}, {0, 3}, {2, 1}, {1, 2},
	{2, 0}, {0, 2}, {1, 1}, {1, 0}, {0, 1}, {0, 0},
}

// colonyCellQuad 回傳格 (a,b) 的四個角,順序同 `Get_Bldg_Box_`:
// (a,b) → (a+1,b) → (a+1,b+1) → (a,b+1)。
func colonyCellQuad(a, b int) [4][2]int {
	return [4][2]int{
		colonyGridCorners[a][b],
		colonyGridCorners[a+1][b],
		colonyGridCorners[a+1][b+1],
		colonyGridCorners[a][b+1],
	}
}

// colonyCellCenter 回傳格 (a,b) 的中心,公式同
// `Bldg_Coords_To_Centered_Screen_Coord_`:對角兩角點取中點。
func colonyCellCenter(a, b int) (int, int) {
	p0 := colonyGridCorners[a][b]
	p1 := colonyGridCorners[a+1][b+1]
	return (p0[0] + p1[0]) / 2, (p0[1] + p1[1]) / 2
}

// colonyCellAt 回傳滑鼠落在哪一格,沒有就回 (-1,-1)。
//
// 由近往遠找(`colonySlotOrder` 倒著走),與原版 `Add_Bldg_Fields_` 建熱區的方向一致——
// 透視格重疊時近端優先,不然點下方那排會被上方的格搶走。
func colonyCellAt(mx, my int) (int, int) {
	for i := len(colonySlotOrder) - 1; i >= 0; i-- {
		a, b := colonySlotOrder[i][0], colonySlotOrder[i][1]
		if pointInQuad(mx, my, colonyCellQuad(a, b)) {
			return a, b
		}
	}
	return -1, -1
}

// pointInQuad 是標準的射線穿越測試(奇數次穿越 = 在內部)。
func pointInQuad(px, py int, q [4][2]int) bool {
	in := false
	for i := 0; i < 4; i++ {
		j := (i + 1) % 4
		xi, yi := q[i][0], q[i][1]
		xj, yj := q[j][0], q[j][1]
		if (yi > py) == (yj > py) {
			continue
		}
		// 用浮點算交點 x:座標是整數但斜率不是,整數除法會在格線上抖動。
		x := float64(xj-xi)*float64(py-yi)/float64(yj-yi) + float64(xi)
		if float64(px) < x {
			in = !in
		}
	}
	return in
}

// colonyBuildingSlot 依 `sub_BC8A6` 把格 (a,b) 換成格號 0..35(蛇行:偶數列左→右、
// 奇數列右→左)。這是 BLDGn.LBX 資產編號的低位部分。
func colonyBuildingSlot(a, b int) int {
	if b%2 == 0 {
		return b*colonyGridCells + a
	}
	return b*colonyGridCells + (colonyGridCells - 1 - a)
}

// origBuildingID 是 remake 的建築英文名 → 原版內部建築編號。
//
// 2026-08-07 搬進 `internal/gamedata`(見該檔說明):開局建築的優先清單也要用同一份對照,
// 兩邊各抄一份遲早會漂開。這裡留一個別名,既有的呼叫端與回歸測試都不必改。
var origBuildingID = gamedata.OrigBuildingID

// colonyBuildingSprite 把「原版建築編號 + 格 (a,b)」換成該畫哪個檔的哪個資產。
// 算式全部來自 `Cache_Load_Bldg_`(見檔頭)。編號不在 1..48 就回 ok=false。
func colonyBuildingSprite(origID, a, b int) (lbx string, asset int, ok bool) {
	if origID < 1 || origID > 48 {
		return "", 0, false
	}
	t := origID - 1
	return fmt.Sprintf("bldg%d.lbx", t/10), (t%10)*36 + colonyBuildingSlot(a, b), true
}

// --- 以下是「畫出來」的部分:把上面那些真值接到殖民地畫面 ---

// origBuildingCategory 是建築表 +14 欄(分類),抄自原版執行檔 `off_17EB3D + 14`
// (`byte_17EB4B`,見 internal/gamedata/buildings.go 檔頭)。索引 = 建築編號 0..48。
//
// 兩處用得到:**7 = 軌道衛星**(`Make_Bldg_Array_For_Colony_` 靠它把衛星排除在地表格點外),
// 以及 `Sort_Bldg_Array_Columns_` 拿它當排序鍵。
var origBuildingCategory = [49]uint8{
	6, 4, 0, 7, 1, 2, 1, 3, 7, 1,
	0, 0, 0, 0, 7, 3, 1, 1, 5, 5,
	3, 0, 0, 2, 1, 1, 2, 2, 0, 1,
	2, 0, 0, 1, 1, 0, 3, 0, 1, 0,
	7, 7, 0, 0, 0, 0, 0, 0, 0,
}

// colonyHouseStyles 是四種房屋外觀,借用軌道衛星的編號。
// 對應 Artemis System Net / Dimensional Portal / Star Base / Star Fortress ——
// 衛星由 `Draw_Colony_Satellites_` 另外畫在軌道上,不進地表格點,編號因此空著被借去當房屋。
var colonyHouseStyles = [4]int{3, 14, 40, 41}

// colonySurfacePlan 重現原版 `Make_Bldg_Array_For_Colony_` @ 0xBC30B 的建築擺放。
//
// ============ 原版流程(逐步照抄)============
//
//	Set_Random_Seed(colonyIdx)          ; 種子就是殖民地索引 → 同一顆星每次進去長得一樣
//	memset(colony_bldgs, 0, 144)        ; 36 格 × 4 bytes
//	for id = 0..48:
//	    if 這個殖民地有這棟:
//	        if 分類 == 7 → 丟去衛星清單(不進地表)
//	        else          → Insert_Bldg_Into_Array(colonyIdx, id, count++)
//	for i < 人口/3 + 1:                  ; 房屋數
//	    Insert_Bldg_Into_Array(colonyIdx, -3, count++)
//	Sort_Bldg_Array_Columns()
//
//	`Insert_Bldg_Into_Array_` @ 0xBC05E 的選格是**蓄水池抽樣**:
//	    n = 0
//	    for a = 0..5: for b = 0..5:
//	        if 格空 { n++; if Random(n) == 1 { 選這格 } }
//	  —— 外層是 `colony_bldgs` 步距 6 的那一軸(即本檔的 a),內層是步距 1 的那一軸(b)。
//	  抽樣呼叫 `Random` 的**次數與順序**都會影響後續(道路就接在這條流上),
//	  所以連「格不空就不抽」這個細節也得照抄。
//
//	房屋外觀:`v5 = (colonyIdx + 目前房屋數) % 4`,再依 `(v5+1) % 4` 選
//	    0 → 3、1 → 14、2 → 40、3 → 41
//
//	排序之後還有 `jitterColonyHouses`(8 輪房屋微調),再 `buildColonyRoads`(道路)——
//	三者串在同一條亂數流上,順序不能動。
func (b *sceneBuilder) colonySurfacePlan(idx int) map[[2]int]int {
	return b.colonySurfaceLayout(idx).plan
}

// colonySurface 是一個殖民地地表的完整內容:建築落點、道路、植被。
//
// 三者**共用同一條亂數流而且有先後**(建築 → 道路 → 植被),拆成三支各自起種就會全錯,
// 所以只有 `colonySurfaceLayout` 這一個入口在推。
type colonySurface struct {
	plan  map[[2]int]int
	roads colonyRoadMap
	veg   colonyVeggieMap
}

// colonySurfaceLayout 推出整片地表。
func (b *sceneBuilder) colonySurfaceLayout(idx int) colonySurface {
	surf := colonySurface{plan: map[[2]int]int{}}
	if b.session == nil || idx < 0 || idx >= len(b.session.PlayerColonies) {
		return surf
	}
	plan := surf.plan
	rng := gamedata.NewOrigRand(uint32(idx)) // Set_Random_Seed(colonyIdx)
	var grid [36]int                         // grid[a*6+b],0 = 空(軸向見下方註解)

	// 蓄水池抽樣選一格空格(= Insert_Bldg_Into_Array_ 的主路徑)。
	// 滿了回 (-1,-1):原版此時改走 Find_Replacement_Slot_For_Building_ 擠掉最低優先的格子,
	// remake 目前直接不放(36 格塞滿的局面要有 30+ 棟建築,現階段到不了)。
	pick := func() (int, int) {
		ga, gb, n := -1, -1, 0
		for a := 0; a < colonyGridCells; a++ {
			for b := 0; b < colonyGridCells; b++ {
				if grid[a*colonyGridCells+b] != 0 {
					continue
				}
				n++
				if rng.N(n) == 1 {
					ga, gb = a, b
				}
			}
		}
		return ga, gb
	}
	put := func(id int) {
		ga, gb := pick()
		if ga < 0 {
			return
		}
		grid[ga*colonyGridCells+gb] = id
	}

	// 已建建築:依**原版編號順序**逐棟插入(原版的迴圈就是 id 0..48,不是玩家建造順序)。
	has := b.colonyOrigBuildingIDs(idx)
	for id := 1; id <= 48; id++ {
		if !has[id] || origBuildingCategory[id] == 7 {
			continue // 分類 7 = 軌道衛星,不畫在地表
		}
		put(id)
	}

	// 房屋:人口/3 + 1。外觀依 (colonyIdx + 已放房屋數 + 1) % 4 輪。
	houses := b.session.PlayerColonies[idx].Population/3 + 1
	for i := 0; i < houses; i++ {
		put(colonyHouseStyles[(idx+i+1)%len(colonyHouseStyles)])
	}

	sortColonyColumns(&grid)
	jitterColonyHouses(&grid, rng)

	// 抖動之後、道路之前,原版還做一次 `Get_Bldg_CR_(9)`(找國會大廈)。
	// 找到之後那段交換由 `dword_182B19` 這個全域 gate 住,它初值 0、唯一的寫入點又在
	// gate 內側,remake 沒有對應概念 —— **交換沒模擬**。
	// 但**掃描本身無條件執行而且會吃亂數**(見 `findColonyBuilding` 註解),而道路接在
	// 同一條流上,所以這一步不能省。
	findColonyBuilding(&grid, origCapitolID, rng)

	surf.roads = buildColonyRoads(&grid, rng)
	// 植被接在道路後面。`N_Bldgs_` 數的是殖民地擁有的建築(含軌道衛星),
	// 即 `has` 的大小——它只影響 `Random(n+2)` 的參數,但那一次抽樣會動到共用的亂數流。
	surf.veg = buildColonyVeggies(&grid, surf.roads, len(has),
		int(b.session.PlayerColonies[idx].Climate), b.colonyVeggieSize, rng)

	for i, id := range grid {
		if id != 0 {
			plan[colonyGridKey(i)] = id
		}
	}
	return surf
}

// origCapitolID 是國會大廈的原版建築編號。
//
// 它**不在** `gamedata.Buildings` 裡,那是對的——建立殖民地時自動給予、不可建造,
// 不該出現在建造選單。但它是一棟**實體建築**:佔一格、有美術(`bldg0.lbx` 資產 8×36+格)、
// 會被畫在地表上。「不在建造表裡」與「不在地表格陣裡」是兩件事,先前混為一談了。
//
// 只有母星有。patch 1.5 手冊三處佐證:
//   - 「最多建築數(**不計 Capitol**)是 ⅔ 人口無條件進位」——它獨立於一般建築計數。
//   - 「**沒有** Capitol 的士氣懲罰」可依政府設定——所以殖民地可能沒有。
//   - 「Colony Base 若加進 `initial_buildings`,會發給**每個玩家的母星**」——
//     母星才是自動給予建築的那一個。
const origCapitolID = 9

// origColonyBaseID 是拓殖基地的原版建築編號,是 `origCapitolID` 的對稱項:
// **母星有國會大廈,其餘殖民地有拓殖基地**,兩者都是拓殖時自動給予、不可建造的實體建築。
//
// 與 Capitol 同一個坑:先前因為「不在建造表裡」就連地表也漏掉了。
// 分類 0(地表建築)、成本 200 PP、維護 0 —— 真值見 `docs/re/01-gap-report.md` 第 36 項那張表,
// 該項對它的註記就是「拓殖時自動」。
const origColonyBaseID = 11

// colonyOrigBuildingIDs 把某殖民地已建的建築換成**原版編號**的集合。
func (b *sceneBuilder) colonyOrigBuildingIDs(idx int) map[int]bool {
	has := map[int]bool{}
	if b.session == nil || idx < 0 || idx >= len(b.session.ColonyBuildings) {
		return has
	}
	if idx == 0 {
		// 殖民地 0 恆為玩家母星(見 `GameSession.PlayerColonyStars` 欄位註解:星 0 恆為母星)。
		has[origCapitolID] = true
	} else {
		// 其餘殖民地拿拓殖基地——與母星的國會大廈對稱,都是拓殖時自動給予的實體建築。
		has[origColonyBaseID] = true
	}
	for zh := range b.session.ColonyBuildings[idx] {
		if bd, ok := gamedata.BuildingByNameZH(zh); ok {
			if id, ok := origBuildingID[bd.NameEN]; ok {
				has[id] = true
			}
		}
	}
	return has
}

// colonyGridKey 把格陣索引換回 (a, b)。原版的元素位址是 `colony_bldgs[24×a + 4×b]`
// (`Add_Bldg_Fields_`,見檔頭「軸向」段),4 個位元組一格 → 索引 = a×6 + b。
//
// 單獨抽成一個函式是因為這裡**反過來就會整張鏡射**,而鏡射後的畫面看起來一樣合理。
// 有它才有地方掛回歸測試。
func colonyGridKey(i int) [2]int {
	return [2]int{i / colonyGridCells, i % colonyGridCells}
}

// sortColonyColumns 是 `Sort_Bldg_Array_Columns_` @ 0xBBDC9:依建築分類做氣泡排序,
// 每輪對每個 (a,b) 比三個方向的鄰格((a+1,b+1)、(a+1,b)、(a,b+1)),
// 分類大的往索引大的方向換。最多 6 輪,某一輪沒換過就提早收工。
// 兩格都非空才換——空格不參與,否則會把建築往空的地方推。
//
// 排序鍵是 `byte_17EB4B[id×19]`(建築表 +14 欄的分類),與 `origBuildingCategory` 同一張表。
// 分類 7 最大,所以四種房屋(借了衛星編號 3/14/40/41,分類都是 7)會被推到索引最大的那一角;
// 這是原版本來的行為,不是 bug——排序的用意就是把同類聚在一起。
//
// 「有沒有換過」原版是用 bx 記(swap 時 `mov bx, word ptr dword_182AFD[eax]` 存被換走的建築編號,
// 而兩格都已檢查非零),每輪開頭 `xor ebx, ebx` 清掉 —— 等價於一個 bool 旗標。
func sortColonyColumns(grid *[36]int) {
	cat := func(id int) uint8 {
		if id < 0 || id >= len(origBuildingCategory) {
			return 0
		}
		return origBuildingCategory[id]
	}
	for pass := 0; pass < colonyGridCells; pass++ {
		swapped := false
		for a := 0; a+1 < colonyGridCells; a++ {
			for b := 0; b+1 < colonyGridCells; b++ {
				cur := a*colonyGridCells + b
				for _, other := range []int{
					(a+1)*colonyGridCells + (b + 1),
					(a+1)*colonyGridCells + b,
					a*colonyGridCells + (b + 1),
				} {
					if grid[cur] == 0 || grid[other] == 0 {
						continue
					}
					if cat(grid[cur]) > cat(grid[other]) {
						grid[cur], grid[other] = grid[other], grid[cur]
						swapped = true
					}
				}
			}
		}
		if !swapped {
			break
		}
	}
}

// colonyJitterRounds 是房屋微調的輪數(原版 `cmp cx, 8`)。
const colonyJitterRounds = 8

// jitterColonyHouses 是 `Sort_Bldg_Array_Columns_` 之後那段房屋微調
// (`Make_Bldg_Array_For_Colony_` @ 0xBC30B 的 `loc_BC441`..`loc_BC560`)。
//
// 排序把同類建築聚成整齊的區塊,這一段再把房屋隨機挪一下,免得城市看起來像棋盤。
// 跑 8 輪,第 n 輪處理 `colonyHouseStyles[n%4]` 那種房屋:找到它的格子,
// 然後在一個 3×3 的偏移迴圈裡試著和另一格互換,**換成一次就結束該輪**。
//
// 機率:目標格非空 → 抽兩次 `Random(3)`,任一次中 1 就換(約 5/9);
// 目標格是空的 → 只抽一次(約 1/3)。
//
// ============ ⚠ 原版這段有兩個 bug,照抄不修 ============
//
// 反組譯(已用 objdump 對原始位元組獨立驗過,IDA 清單無誤):
//
//	eax = var_14 - si ; if (eax < 0) eax = 0        ; ← 第一個座標:有夾到 0
//	edx = var_18 - di ; test edx, edx               ; ← 第二個座標:算了、測了號誌…
//	cmp ax, 5 ; jle → edx = ax ; else edx = 5       ; …然後 edx 被無條件蓋掉
//
// **(1) 第二個座標整個被丟棄。** 對應的 `jge / xor edx, edx`(夾到 0)沒有被編出來,
// 而下一條指令兩條路徑都會覆寫 `edx`。於是換位對象的兩個座標都等於
// `clamp(a - si, 0, 5)` —— 目標格永遠落在 6×6 格陣的**主對角線**上。
//
// **(2) 內圈的 `di` 因此完全沒有作用。** 三次迭代做的是一模一樣的事。
// 這裡仍然保留那個迴圈:它會影響 `Random` 被抽幾次,而後面的道路接在同一條流上。
//
// 換句話說,原版的「隨機微調」實際效果是「把房屋往對角線上搬」。照著原版寫才會得到
// 原版的畫面;把它修成「與鄰格互換」會讓 remake 比原版合理,但與原版不同。
func jitterColonyHouses(grid *[36]int, rng *gamedata.OrigRand) {
	for round := 0; round < colonyJitterRounds; round++ {
		id := colonyHouseStyles[round%len(colonyHouseStyles)]
		a, b, ok := findColonyBuilding(grid, id, rng)
		if !ok {
			continue
		}
		done := false
		for si := 1; si >= -1 && !done; si-- {
			for di := 1; di >= -1 && !done; di-- {
				_ = di // ⚠ 原版沒用到,見上方 bug (2);迴圈本身要留著,它決定抽幾次亂數

				v := a - si
				if v < 0 {
					v = 0
				}
				if v > colonyGridCells-1 {
					v = colonyGridCells - 1
				}
				target := v*colonyGridCells + v // ⚠ 對角格,見上方 bug (1)
				src := a*colonyGridCells + b

				swap := grid[target] != 0 && rng.N(3) == 1
				if !swap {
					swap = rng.N(3) == 1
				}
				if swap {
					grid[target], grid[src] = grid[src], grid[target]
					done = true
				}
			}
		}
	}
}

// findColonyBuilding 是 `Get_Bldg_CR_` @ 0xBBD37:在格陣裡找編號 id 的格子。
//
// 語意比名字繞,三件事要一起看:
//
//   - 掃描序是 a 外 b 內,**命中就立刻回傳第一格**(同一種房屋有好幾格時取掃到的第一格)。
//   - 命中之前每碰到一個**空格**都會抽一次 `Random(n)` 做蓄水池抽樣,記住一個隨機空格。
//     所以「找一棟建築」這個動作**會消耗亂數,而且消耗幾次取決於資料**——
//     後面的道路接在同一條流上,這個細節不能省。
//   - 掃完沒找到時:`id == -1` 回傳那個隨機空格且算成功(原版用這條路徑找空位),
//     其他 id 一律回傳失敗。
func findColonyBuilding(grid *[36]int, id int, rng *gamedata.OrigRand) (int, int, bool) {
	ra, rb, n := 0, 0, 0
	for a := 0; a < colonyGridCells; a++ {
		for b := 0; b < colonyGridCells; b++ {
			switch cell := grid[a*colonyGridCells+b]; cell {
			case id:
				return a, b, true
			case 0:
				n++
				if rng.N(n) == 1 {
					ra, rb = a, b
				}
			}
		}
	}
	if id == -1 {
		return ra, rb, true
	}
	return 0, 0, false
}

// drawColonyBuildings 把建築貼上去。與 `drawColonyTerrain` 分開是因為原版**中間夾了框架**:
// `Draw_Colony_Screen_` 的順序是 地表 → `Draw_Colony_Info_Background`(框架)→
// `Draw_Colony_Bldgs` → 資訊面板。近端那幾格的 y 會壓到 423 以下(框架非透明區),
// 順序併掉的話那排建築會被框架切掉一截。
//
// 貼的次序用 `colonySlotOrder`(遠 → 近),那是原版 `Add_Bldg_Fields_` 那張表的順序,
// 也正是重疊建築要的畫家演算法次序。
//
// **每一格是「先植被再建築」**,不是「先畫完所有植被再畫所有建築」——原版
// `Draw_Colony_Bldgs_` @ 0xBEBDC 在同一個迴圈裡對每格先呼叫 `sub_B6B95`(植被)再畫建築。
// 差別在遮擋:近處的草要蓋住遠處的建築下緣。
func (b *sceneBuilder) drawColonyBuildings(dst *ebiten.Image, surf colonySurface) {
	for _, cell := range colonySlotOrder {
		b.drawColonyVeggiesAt(dst, surf.veg, cell[0], cell[1])
		id, ok := surf.plan[cell]
		if !ok || id == 0 {
			continue
		}
		if im := b.colonyBuildingImage(id, cell[0], cell[1]); im != nil {
			dst.DrawImage(im, &ebiten.DrawImageOptions{})
		}
	}
}

// --- 地表底圖 ---
//
// `Draw_Colony_Screen_` @ 0xBED21 開場疊兩層,都貼在 (0,0) 且都是 640×480:
//
//	v1 = C_Anims(1, 0, 639, 479); Draw(0, 0, v1)   ← 底層
//	v4 = C_Anims(0, …);           Draw(0, 0, v4)   ← 地形,疊在上面
//
// `C_Anims` @ 0xBBA8E 是一張 36 路跳表,這兩路解出來是:
//
//	case 1 → COLONY2.LBX 資產 0x31 = 49(整個檔只有這一張是 640×480)
//	case 0 → PLANETS.LBX,編號現算:
//	    eax = colony_idx×0x169 + colony_base      ; 殖民地結構
//	    edx = word[eax+2] × 0x11                  ; 星球索引 × 17(星球表步距)
//	    ax  = byte[eax+0xE2] × 3                  ; 殖民地 +0xE2 欄 × 3
//	    dx  = byte[edx + star_base + 9]           ; 星球表 +9 欄
//	    → 資產 = byte[colony+0xE2]×3 + star[+9]
//
// PLANETS.LBX 恰好 30 張 640×480 → **10 種氣候 × 3 個變體**,所以 +0xE2 欄是氣候
// (0..9)、星球表 +9 欄是變體(0..2)。渲染出來對得上:資產 27 = 9×3 是整片蔥綠(Gaia)、
// 資產 3 = 1×3 是輻射星的赤紅熔岩裂縫、資產 0..2 是有毒星的三種面貌。remake 的
// `gamedata.PlanetClimate` 列舉(TOXIC=0 … GAIA=9)與這個順序逐項相同。
//
// ⚠ **變體那一欄還沒還原**:它是銀河生成時寫進星球表 +9 的一個存檔欄位,產生規則沒追。
// 這裡改用原版 PRNG 以星球索引起種來取 1..3 —— 保證「同一顆星每次進去長一樣」這個
// 玩家看得到的性質,但**不保證與原版同一局的那顆星相同**。追到規則前不要把它寫成真值。

const (
	colTerrainSkyLBX   = "colony2.lbx"
	colTerrainSkyAsset = 49
	colTerrainLBX      = "planets.lbx"
	colTerrainVariants = 3
)

// colonyTerrainVariant 回傳某顆**行星**的地形變體 0..2(見上方 ⚠)。
//
// 起種用行星索引不是星索引:同一個星系可以有多個殖民地,用星索引的話它們的地表會一模一樣。
func colonyTerrainVariant(planet int) int {
	if planet < 0 {
		return 0
	}
	return gamedata.NewOrigRand(uint32(planet)).N(colTerrainVariants) - 1
}

// drawColonyTerrain 畫地表兩層。取不到資產就整張留空(不再自畫格線佔位——
// 原版地表上**沒有格線**,格點是隱形的)。
func (b *sceneBuilder) drawColonyTerrain(dst *ebiten.Image, idx int) {
	if im := b.colonyScreenImage(colTerrainSkyLBX, colTerrainSkyAsset, colonyBasePalette); im != nil {
		dst.DrawImage(im, &ebiten.DrawImageOptions{})
	}
	sess := b.session
	if sess == nil || idx < 0 || idx >= len(sess.PlayerColonies) {
		return
	}
	climate := int(sess.PlayerColonies[idx].Climate)
	if climate < 0 || climate*colTerrainVariants >= 30 {
		return
	}
	// 變體種子用**行星**索引:同一個星系的兩個殖民地用星索引會長得一模一樣。
	asset := climate*colTerrainVariants + colonyTerrainVariant(sess.ColonyPlanetIndex(idx))
	if im := b.colonyScreenImage(colTerrainLBX, asset, colonyBasePalette); im != nil {
		dst.DrawImage(im, &ebiten.DrawImageOptions{})
	}
}

// colonyScreenImage 解一張殖民地畫面用的圖並快取。
//
// 調色盤走「chain 當基底 + 該圖自己的內嵌範圍」:PLANETS 每張只帶 80 色,缺的要基底補,
// 否則熔岩裂縫會變成一片洋紅(⚠ 與第 29 項那次的洋紅**不同原因**——那次是 index ≥ 0xF0)。
// COLONY.LBX 的衛星圖完全沒有內嵌盤,整組都得靠 chain。
func (b *sceneBuilder) colonyScreenImage(lbxName string, asset int, chain paletteChain) *ebiten.Image {
	if b.res == nil {
		return nil // 沒有資產解析器(單元測試等):畫面自己降級,不要 panic
	}
	key := lbxName + ":" + strconv.Itoa(asset)
	if im, hit := b.colBldgCache[key]; hit {
		return im
	}
	if b.colBldgCache == nil {
		b.colBldgCache = map[string]*ebiten.Image{}
	}
	var img *ebiten.Image
	if im, err := decodeAsset(b.res, lbxName, asset); err == nil && len(im.Frames) > 0 {
		if pal, err := resolvePalette(b.res, im, chain); err == nil {
			img = ebiten.NewImageFromImage(im.Frames[0].ToRGBADropTranslucent(pal, im.KeyColor()))
		}
	}
	b.colBldgCache[key] = img
	return img
}

// colonyBasePalette 是殖民地畫面的調色盤基底:全域基底 + 殖民地美術那組。
var colonyBasePalette = paletteChain{{"buffer0.lbx", 0}, {colChromePal, 0}}

// --- 軌道衛星 ---
//
// 分類 7 的建築**不進地表格點**。`Make_Bldg_Array_For_Colony_` 把它們丟進 `word_19F99C`
// 那份 10 格的清單(依建築編號遞增附加),由 `Draw_Colony_Satellites_` @ 0xBE366 另外畫。
//
// 位置(`Draw_Colony_Satellites_` 尾段,整段沒有查表):
//
//	x = 295 + (i 偶數 ? +1 : −1) × i × 50      ; 295 = 0x127、50 = 0x32
//	y = 162                                    ; 0xA2,固定
//	Draw(x, y, img)                            ; 左上角對位
//
// → 第 0 顆在 295、第 1 顆 245、第 2 顆 395、第 3 顆 145…往兩側交錯散開。
// 第 7 顆起 x 已是負的(−55),原版清單雖有 10 格但實際塞得下約 7 顆。
//
// 圖檔(`sub_BE306` 的比較鏈 + `sub_BBB9F` 的 `loc_BBBAF: add edx, 9`):
//
//	⚠ 那個 **+9 不能漏**。`sub_BE306` 算出來的是 0/1/2/3/7,加 9 之後才是真正的資產編號。
//	漏掉會去讀 COLONY.LBX 資產 0..4 —— 那五格在檔案裡是**零長度**的(offset 表全是 0x800),
//	解出來是空圖,畫面上什麼都不會出現,而且不會報錯。
//
//	| 建築 | 編號 | sub_BE306 | COLONY.LBX 資產 |
//	|---|---|---|---|
//	| 星際要塞 Star Fortress | 41 | 0 | 9 |
//	| 戰鬥站 Battlestation | 8 | 1 | 10 |
//	| 星基 Star Base | 40 | 2 | 11 |
//	| 次元傳送門 Dimensional Portal | 14 | 3 | 12 |
//	| 天網 Artemis System Net | 3 | 7 | 16 |
//
// 資產 9..16 全是 57×70,尺寸自洽。
//
// 抑制規則(`sub_BC21B`,回傳 1 = 這顆不畫)——就是原版的**星基升級鏈**:
//
//	星基(40):有戰鬥站(colony+0x13E)或星際要塞(colony+0x15F)就不畫
//	戰鬥站(8):有星際要塞就不畫
//
// (`0x136 + id` 是「這個殖民地有沒有這棟」的旗標陣列,0x136+8 = 0x13E、0x136+41 = 0x15F,
// 兩個位移都對得起來。)
const (
	colonySatLBX   = "colony.lbx"
	colonySatBaseX = 295
	colonySatStepX = 50
	colonySatY     = 162
	colonySatSlots = 10 // word_19F99C 的長度(memset 20 bytes)
)

// origSatelliteAsset 分類 7 的建築編號 → COLONY.LBX 資產(已含 +9,見上方 ⚠)。
var origSatelliteAsset = map[int]int{
	41: 9,  // 星際要塞
	8:  10, // 戰鬥站
	40: 11, // 星基
	14: 12, // 次元傳送門
	3:  16, // 天網
}

// colonySatellites 回傳要畫的衛星編號,順序 = 建築編號遞增(原版迴圈的順序)。
func (b *sceneBuilder) colonySatellites(idx int) []int {
	return colonySatelliteList(b.colonyOrigBuildingIDs(idx))
}

// colonySatelliteX 回傳第 i 顆衛星的左上角 x(見上方公式)。
func colonySatelliteX(i int) int {
	if i%2 != 0 {
		return colonySatBaseX - i*colonySatStepX
	}
	return colonySatBaseX + i*colonySatStepX
}

// colonySatelliteList 是 colonySatellites 的純函式部分(方便直接測抑制規則)。
func colonySatelliteList(has map[int]bool) []int {
	var out []int
	for id := 1; id <= 48; id++ {
		if !has[id] || origBuildingCategory[id] != 7 {
			continue
		}
		switch id {
		case 40: // 星基:被戰鬥站或星際要塞取代
			if has[8] || has[41] {
				continue
			}
		case 8: // 戰鬥站:被星際要塞取代
			if has[41] {
				continue
			}
		}
		if len(out) >= colonySatSlots {
			break
		}
		out = append(out, id)
	}
	return out
}

// drawColonySatellites 把衛星畫在軌道上。原版的次序是建築之後(`Draw_Colony_Screen_`:
// `Draw_Colony_Bldgs` → `Draw_Colony_Satellites`)。
func (b *sceneBuilder) drawColonySatellites(dst *ebiten.Image, idx int) {
	for i, id := range b.colonySatellites(idx) {
		asset, ok := origSatelliteAsset[id]
		if !ok {
			continue
		}
		x := colonySatelliteX(i)
		if x <= -57 || x >= moo2ScreenW {
			continue // 原版也是這樣被畫布裁掉,不特別處理
		}
		im := b.colonyScreenImage(colonySatLBX, asset, colonyBasePalette)
		if im == nil {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), colonySatY)
		dst.DrawImage(im, op)
	}
}

// colonyBuildingImage 取(建築編號, 格)對應的那張**已畫好位置**的 640×480 稀疏圖,
// 貼 (0,0) 就對位(見檔頭)。半透明標記索引依原版 `Draw_No_Glass_` 丟掉,
// 否則陰影會變成一塊洋紅(見 internal/lbx 的 TranslucentIndexMin)。
func (b *sceneBuilder) colonyBuildingImage(id, a, bb int) *ebiten.Image {
	lbxName, asset, ok := colonyBuildingSprite(id, a, bb)
	if !ok {
		return nil
	}
	key := lbxName + ":" + strconv.Itoa(asset)
	if im, hit := b.colBldgCache[key]; hit {
		return im
	}
	if b.colBldgCache == nil {
		b.colBldgCache = map[string]*ebiten.Image{}
	}
	var img *ebiten.Image
	if pal, err := decodeAsset(b.res, colChromePal, 0); err == nil && pal.Embedded != nil {
		if im, err := decodeAsset(b.res, lbxName, asset); err == nil && len(im.Frames) > 0 {
			img = ebiten.NewImageFromImage(im.Frames[0].ToRGBADropTranslucent(pal.Embedded, true))
		}
	}
	b.colBldgCache[key] = img // 取不到也快取,避免每幀重試
	return img
}
