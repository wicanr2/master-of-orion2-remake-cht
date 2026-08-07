package main

import (
	"fmt"
	"image/color"
	"sort"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
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

// origBuildingID 是 remake 的建築英文名(`gamedata.Building.NameEN`,取自手冊 The Big List)
// → **原版內部建築編號**(1-based,見檔頭的兩個獨立來源)。
//
// 為什麼需要一張手寫對照:手冊的用字和遊戲內部字串不一樣——手冊寫 "Automated Factories"、
// 遊戲內部是 "Automated Factory";手冊 "Alien Management Center"、內部 "Alien Control Center";
// 手冊 "Planetary Stock Exchange"、內部 "Stock Exchange"。照字串比對會漏一半。
// 右欄註記的是原版 TECHNAME.LBX 裡的原文,方便日後核對。
var origBuildingID = map[string]int{
	"Alien Management Center":     1,  // Alien Control Center
	"Armor Barracks":              2,  // Armor Barracks
	"Artemis System Net":          3,  // Artemis System Net
	"Astro University":            4,  // Astro University
	"Atmospheric Renewer":         5,  // Atmosphere Renewer
	"Autolab":                     6,  // Autolab
	"Automated Factories":         7,  // Automated Factory
	"Battlestation":               8,  // Battlestation
	"Cloning Center":              10, // Cloning Center
	"Deep Core Mine":              12, // Deep Core Mine
	"Core Waste Dumps":            13, // Core Waste Dump
	"Dimensional Portal":          14, // Dimensional Portal
	"Biospheres":                  15, // Biospheres
	"Food Replicators":            16, // Food Replicators
	"Galactic Cybernet":           19, // Galactic Cybernet
	"Holo Simulator":              20, // Holo Simulator
	"Hydroponic Farm":             21, // Hydroponic Farm
	"Marine Barracks":             22, // Marine Barracks
	"Planetary Barrier Shield":    23, // Barrier Shield
	"Planetary Flux Shield":       24, // Flux Shield
	"Planetary Gravity Generator": 25, // Gravity Generator
	"Missile Base":                26, // Missile Base
	"Ground Batteries":            27, // Ground Batteries
	"Planetary Radiation Shield":  28, // Radiation Shield
	"Planetary Stock Exchange":    29, // Stock Exchange
	"Planetary Supercomputer":     30, // Supercomputer
	"Pleasure Dome":               31, // Pleasure Dome
	"Pollution Processor":         32, // Pollution Processor
	"Recyclotron":                 33, // Recyclotron
	"Robotic Factory":             34, // Robotic Factory
	"Research Laboratory":         35, // Research Lab
	"Robo Mining Plant":           36, // Robo Miner Plant
	"Space Academy":               38, // Space Academy
	"Spaceport":                   39, // Spaceport
	"Star Base":                   40, // Star Base
	"Star Fortress":               41, // Star Fortress
	"Subterranean Farms":          43, // Subterranean Farms
	"Warp Field Interdictor":      45, // Warp Interdictor
	"Weather Controller":          46, // Weather Controller
	"Fighter Garrison":            47, // Fighter Garrison
}

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

// colonySurfacePlan 決定這個殖民地的 36 格各放什麼(原版建築編號,0 = 空)。
//
// ⚠ **哪一格放哪棟,remake 與原版不同,這裡誠實標明。**
// 原版 `Make_Bldg_Array_For_Colony_` @ 0xBC30B 的作法是:
// `Set_Random_Seed(colonyIdx, 0, 144)` 起手 → 逐棟 `Insert_Bldg_Into_Array_`,
// 而那支對所有空格做**蓄水池抽樣**(`Random(++n) == 1`)→ 再 `Sort_Bldg_Array_Columns_`
// 依建築分類做氣泡排序 → 最後一段隨機微調。也就是說擺法綁死在**原版的 PRNG**
// (`Random_` @ 0x1247A0 / `Set_Random_Seed_` @ 0x124820)上,那還沒實作。
//
// 在 PRNG 到手之前,這裡用「依編號順序從近端往遠端填」——**幾何、圖檔、遮擋順序、
// 半透明處理全是原版真值,只有落在哪一格是 remake 自己的**。不假裝這是原版擺法。
//
// 房屋那一段倒是照原版做的:數量 = 人口/3 + 1,外觀在 3 / 14 / 40 / 41 之間輪
// (那四個是軌道衛星的編號——衛星不畫在地表,所以編號空著被借去當房屋,見檔頭)。
func (b *sceneBuilder) colonySurfacePlan(idx int) map[[2]int]int {
	plan := map[[2]int]int{}
	if b.session == nil || idx < 0 || idx >= len(b.session.PlayerColonies) {
		return plan
	}
	// 近端 → 遠端(colonySlotOrder 是遠→近,倒著走)。
	free := make([][2]int, 0, len(colonySlotOrder))
	for i := len(colonySlotOrder) - 1; i >= 0; i-- {
		free = append(free, colonySlotOrder[i])
	}
	next := 0
	put := func(id int) {
		if next >= len(free) {
			return
		}
		plan[free[next]] = id
		next++
	}

	// 已建建築:依原版編號排序,讓同一組建築每次畫出來一樣(不隨 map 迭代順序跳動)。
	var ids []int
	if idx < len(b.session.ColonyBuildings) {
		for zh := range b.session.ColonyBuildings[idx] {
			bd, ok := gamedata.BuildingByNameZH(zh)
			if !ok {
				continue
			}
			if id, ok := origBuildingID[bd.NameEN]; ok {
				ids = append(ids, id)
			}
		}
	}
	sort.Ints(ids)
	for _, id := range ids {
		if isOrbitalOnly(id) {
			continue // 軌道衛星不畫在地表(原版靠建築表 +14 分類 == 7 篩掉)
		}
		put(id)
	}

	// 房屋:人口/3 + 1(原版 `Make_Bldg_Array_For_Colony_` 的迴圈上限)。
	houses := b.session.PlayerColonies[idx].Population/3 + 1
	for i := 0; i < houses; i++ {
		put(colonyHouseStyles[(idx+i+1)%len(colonyHouseStyles)])
	}
	return plan
}

// colonyHouseStyles 是四種房屋外觀,借用軌道衛星的編號(見 colonySurfacePlan 註解)。
// 對應 Artemis System Net / Dimensional Portal / Star Base / Star Fortress。
var colonyHouseStyles = [4]int{3, 14, 40, 41}

// isOrbitalOnly 回報這個編號是不是只存在於軌道(衛星),不畫在地表。
func isOrbitalOnly(id int) bool {
	for _, h := range colonyHouseStyles {
		if id == h {
			return true
		}
	}
	return id == 8 // Battlestation
}

// drawColonySurface 畫殖民地畫面中段的行星表面:格線 + 建築。
//
// 畫的順序用 `colonySlotOrder`(遠 → 近),那是原版 `Add_Bldg_Fields_` 那張表的順序,
// 也正是重疊建築要的畫家演算法次序。
func (b *sceneBuilder) drawColonySurface(dst *ebiten.Image, idx int) {
	plan := b.colonySurfacePlan(idx)

	// 格線:原版這裡是地表底圖(還沒接,見 gap report 第 27 項),先用淡格線把
	// 透視平面畫出來,免得建築看起來浮在半空。
	grid := color.RGBA{56, 74, 56, 120}
	for a := 0; a < colonyGridCells; a++ {
		for bb := 0; bb < colonyGridCells; bb++ {
			q := colonyCellQuad(a, bb)
			for i := 0; i < 4; i++ {
				j := (i + 1) % 4
				vector.StrokeLine(dst, float32(q[i][0]), float32(q[i][1]),
					float32(q[j][0]), float32(q[j][1]), 1, grid, false)
			}
		}
	}

	for _, cell := range colonySlotOrder {
		id, ok := plan[cell]
		if !ok || id == 0 {
			continue
		}
		if im := b.colonyBuildingImage(id, cell[0], cell[1]); im != nil {
			dst.DrawImage(im, &ebiten.DrawImageOptions{})
		}
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
