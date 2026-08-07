package main

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
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

// TestColonyGridKeyMatchesOriginalAddressing:格陣索引 → (a,b) 的**軸向**。
//
// 這一條是花最多時間才抓到的錯,而且它**不會有任何症狀**:軸向反過來以後,畫面上照樣是
// 一片合理的透視格、建築也照樣落在格線上,只是整張佈局沿對角線鏡射掉。第一版就是這樣錯的
// ——當時的徵狀只有「建築全擠在遠端那幾排」,而那太容易被當成隨機。
//
// 定案證據是 `Add_Bldg_Fields_` @ 0xBE44A 把同一對 (v1, v2) 同時餵給兩邊:
// 元素位址 `colony_bldgs[24×v1 + 4×v2]`、螢幕座標
// `Bldg_Coords_To_Centered_Screen_Coord(v1, v2, c)`(索引 v1×56 + v2×8 + c×4)。
func TestColonyGridKeyMatchesOriginalAddressing(t *testing.T) {
	for a := 0; a < colonyGridCells; a++ {
		for b := 0; b < colonyGridCells; b++ {
			i := (24*a + 4*b) / 4 // 原版位址 → 索引(一格 4 bytes)
			if got := colonyGridKey(i); got != [2]int{a, b} {
				t.Errorf("位址 24×%d+4×%d(索引 %d)應回 (%d,%d),實得 %v", a, b, i, a, b, got)
			}
		}
	}
	if k := colonyGridKey(1); k != [2]int{0, 1} {
		t.Fatalf("索引 1 應為 (0,1),實得 %v ——軸向反了,整張佈局會鏡射", k)
	}
	if k := colonyGridKey(6); k != [2]int{1, 0} {
		t.Fatalf("索引 6 應為 (1,0),實得 %v ——軸向反了", k)
	}

	// 把索引接回螢幕:索引 +1 是沿 b 走(往遠端左上,y 變小、x 變小),
	// 索引 +6 是沿 a 走(往近端右下,y 變小但 x 變大)。兩者方向不同,軸向反了就對不上。
	x0, y0 := colonyCellCenter(0, 0)
	xb, yb := colonyCellCenter(colonyGridKey(1)[0], colonyGridKey(1)[1])
	xa, ya := colonyCellCenter(colonyGridKey(6)[0], colonyGridKey(6)[1])
	if !(xb < x0 && yb < y0) {
		t.Errorf("索引 1 應在 (0,0) 的左上方:(%d,%d) → (%d,%d)", x0, y0, xb, yb)
	}
	if !(xa > x0 && ya < y0) {
		t.Errorf("索引 6 應在 (0,0) 的右上方:(%d,%d) → (%d,%d)", x0, y0, xa, ya)
	}
}

// TestColonySurfacePlanIsDeterministic:同一個殖民地每次算出來要一樣。
// 原版 `Set_Random_Seed(colonyIdx)` 的用意就是這個——玩家每次進去看到同一片城市。
// 若不小心改用 map 迭代或 Go 的全域 rand,這裡會炸。
func TestColonySurfacePlanIsDeterministic(t *testing.T) {
	b := newSurfaceTestBuilder(t)
	a1 := b.colonySurfacePlan(0)
	a2 := b.colonySurfacePlan(0)
	if len(a1) != len(a2) {
		t.Fatalf("兩次算出的棟數不同:%d vs %d", len(a1), len(a2))
	}
	for k, v := range a1 {
		if a2[k] != v {
			t.Errorf("格 %v:第一次 %d、第二次 %d", k, v, a2[k])
		}
	}
}

// TestColonySurfacePlanDiffersByColony:種子是殖民地索引,所以**輸入完全一樣的兩個殖民地
// 也要長得不一樣**。這裡複製兩份殖民地 0(建築、人口全同)當殖民地 1 與 2 再比——
// 這樣任何差異都只可能來自種子,不會被「兩顆星本來就不同」蓋過去。
//
// ⚠ 不拿殖民地 0 來比:0 號是母星,會多一棟國會大廈(見 `origCapitolID`),
// 那是**索引本身帶進來的輸入差異**,不是種子造成的。
func TestColonySurfacePlanDiffersByColony(t *testing.T) {
	b := newSurfaceTestBuilder(t)
	sess := b.session
	for i := 0; i < 2; i++ {
		sess.PlayerColonies = append(sess.PlayerColonies, sess.PlayerColonies[0])
		dup := map[string]bool{}
		for k, v := range sess.ColonyBuildings[0] {
			dup[k] = v
		}
		sess.ColonyBuildings = append(sess.ColonyBuildings, dup)
	}

	p1, p2 := b.colonySurfacePlan(1), b.colonySurfacePlan(2)
	if len(p1) == 0 {
		t.Skip("殖民地沒有可畫的東西")
	}
	if len(p1) != len(p2) {
		t.Fatalf("輸入相同,棟數卻不同:%d vs %d", len(p1), len(p2))
	}
	same := true
	for k, v := range p1 {
		if p2[k] != v {
			same = false
			break
		}
	}
	if same {
		t.Error("兩個輸入相同的殖民地擺法完全一樣——Set_Random_Seed(colonyIdx) 沒接上")
	}
}

// TestColonySurfacePlanHomeworldHasCapitol:母星要有國會大廈,其他殖民地沒有。
//
// 它不在 `gamedata.Buildings` 裡(不可建造,對建造選單而言正確),先前因此連地表也漏掉了——
// 「不在建造表」與「不在地表」是兩件事。
func TestColonySurfacePlanHomeworldHasCapitol(t *testing.T) {
	b := newSurfaceTestBuilder(t)
	sess := b.session
	sess.PlayerColonies = append(sess.PlayerColonies, sess.PlayerColonies[0])
	dup := map[string]bool{}
	for k, v := range sess.ColonyBuildings[0] {
		dup[k] = v
	}
	sess.ColonyBuildings = append(sess.ColonyBuildings, dup)

	count := func(idx int) int {
		n := 0
		for _, id := range b.colonySurfacePlan(idx) {
			if id == origCapitolID {
				n++
			}
		}
		return n
	}
	if got := count(0); got != 1 {
		t.Errorf("母星應有 1 棟國會大廈,實得 %d", got)
	}
	if got := count(1); got != 0 {
		t.Errorf("非母星不該有國會大廈,實得 %d", got)
	}
}

// TestColonySurfacePlanExcludesOrbitals:軌道衛星(分類 7)不該出現在地表格點上,
// 除非它是被借去當房屋的那四個編號。原版靠建築表 +14 欄篩掉。
func TestColonySurfacePlanExcludesOrbitals(t *testing.T) {
	b := newSurfaceTestBuilder(t)
	house := map[int]bool{}
	for _, h := range colonyHouseStyles {
		house[h] = true
	}
	for i := range b.session.PlayerColonies {
		for cell, id := range b.colonySurfacePlan(i) {
			if origBuildingCategory[id] == 7 && !house[id] {
				t.Errorf("殖民地 %d 格 %v 放了軌道衛星編號 %d", i, cell, id)
			}
		}
	}
}

// TestColonySurfacePlanFitsGrid:所有格號都在 6×6 之內,且不重疊。
func TestColonySurfacePlanFitsGrid(t *testing.T) {
	b := newSurfaceTestBuilder(t)
	for i := range b.session.PlayerColonies {
		plan := b.colonySurfacePlan(i)
		if len(plan) > colonyGridCells*colonyGridCells {
			t.Errorf("殖民地 %d 放了 %d 個東西,超過 %d 格", i, len(plan), colonyGridCells*colonyGridCells)
		}
		for cell := range plan {
			if cell[0] < 0 || cell[0] >= colonyGridCells || cell[1] < 0 || cell[1] >= colonyGridCells {
				t.Errorf("殖民地 %d 的格 %v 超出範圍", i, cell)
			}
		}
	}
}

// TestSortColonyColumnsOrdersByCategory:分類大的要被換到索引大的那一端,
// 而且空格不參與(否則建築會被往空的地方推)。
func TestSortColonyColumnsOrdersByCategory(t *testing.T) {
	var g [36]int
	g[0] = 3  // 分類 7(房屋)
	g[7] = 35 // 分類 0(研究實驗室)—— 在 (a=1,b=1)
	sortColonyColumns(&g)
	if origBuildingCategory[g[0]] > origBuildingCategory[g[7]] {
		t.Errorf("分類大的沒被換到後面:g[0]=%d(分類 %d) g[7]=%d(分類 %d)",
			g[0], origBuildingCategory[g[0]], g[7], origBuildingCategory[g[7]])
	}

	// 空格不動:只有一棟建築時排序後那棟還在原地。
	var h [36]int
	h[14] = 22
	sortColonyColumns(&h)
	if h[14] != 22 {
		t.Errorf("只有一棟建築時不該被搬走,實得 g[14]=%d", h[14])
	}
}

// newSurfaceTestBuilder 建一個有真實對局的 sceneBuilder(不載入任何版權資產)。
func newSurfaceTestBuilder(t *testing.T) *sceneBuilder {
	t.Helper()
	sess := shell.NewDemoSession()
	sess.SetupNewGame(24, 4242, shell.DefaultOpponents)
	for i := 0; i < 12; i++ {
		sess.EndTurn() // 跑幾回合讓殖民地長出建築與人口
	}
	return &sceneBuilder{session: sess, lang: i18n.Traditional}
}

// --- 軌道衛星 ---

// TestOrigSatelliteAssetsCoverEveryCategory7:分類 7 的建築編號與衛星圖對照要**互相涵蓋**。
// 少一筆 → 那顆衛星在畫面上悄悄消失(不會報錯);多一筆 → 會去畫一棟其實在地表的建築。
func TestOrigSatelliteAssetsCoverEveryCategory7(t *testing.T) {
	for id := 1; id <= 48; id++ {
		_, hasSprite := origSatelliteAsset[id]
		isSat := origBuildingCategory[id] == 7
		if isSat && !hasSprite {
			t.Errorf("建築 %d 是分類 7 卻沒有衛星圖", id)
		}
		if hasSprite && !isSat {
			t.Errorf("建築 %d 有衛星圖卻不是分類 7", id)
		}
	}
	// ⚠ 資產編號含 `sub_BBB9F` 的 +9。少了它會去讀 COLONY.LBX 資產 0..4——
	// 那五格在檔案裡是零長度的,解出來是空圖,畫面什麼都不會出現而且不會報錯。
	for id, asset := range origSatelliteAsset {
		if asset < 9 || asset > 16 {
			t.Errorf("建築 %d 的衛星圖資產 %d 不在 9..16(是不是漏了 +9?)", id, asset)
		}
	}
}

// TestColonySatelliteUpgradeChain:星基 → 戰鬥站 → 星際要塞 的升級鏈,
// 抄自 `sub_BC21B`(回傳 1 = 這顆不畫)。三顆同時存在時只該看到最高階那顆。
func TestColonySatelliteUpgradeChain(t *testing.T) {
	set := func(ids ...int) map[int]bool {
		m := map[int]bool{}
		for _, id := range ids {
			m[id] = true
		}
		return m
	}
	eq := func(got []int, want ...int) bool {
		if len(got) != len(want) {
			return false
		}
		for i := range got {
			if got[i] != want[i] {
				return false
			}
		}
		return true
	}
	cases := []struct {
		name string
		have map[int]bool
		want []int
	}{
		{"只有星基", set(40), []int{40}},
		{"星基+戰鬥站 → 只剩戰鬥站", set(40, 8), []int{8}},
		{"星基+戰鬥站+要塞 → 只剩要塞", set(40, 8, 41), []int{41}},
		{"星基+要塞(跳過戰鬥站)→ 只剩要塞", set(40, 41), []int{41}},
		{"天網與傳送門不受升級鏈影響", set(3, 14, 40, 41), []int{3, 14, 41}},
		{"地表建築不進清單", set(22, 35), nil},
	}
	for _, c := range cases {
		if got := colonySatelliteList(c.have); !eq(got, c.want...) {
			t.Errorf("%s:得到 %v,want %v", c.name, got, c.want)
		}
	}
}

// TestColonySatellitePositionsMatchOriginal:x = 295 + (i 偶數 ? +1 : −1) × i × 50。
// 釘住前幾顆的絕對座標——寫錯正負或步距,衛星會擠成一團或飛出畫面。
func TestColonySatellitePositionsMatchOriginal(t *testing.T) {
	want := []int{295, 245, 395, 145, 495, 45, 595, -55}
	for i, w := range want {
		if got := colonySatelliteX(i); got != w {
			t.Errorf("第 %d 顆 x = %d,want %d", i, got, w)
		}
	}
	// 第 7 顆在 x = −55,57 寬的圖只剩右緣 2px 在畫面內;第 8 顆 x = 695 整個在右外。
	// 原版清單雖有 10 格,實際看得到的就是前 7 顆——這是原版本來的裁切,不特別處理。
	if x := colonySatelliteX(7); x+57 > 5 {
		t.Errorf("第 7 顆右緣 %d 應該幾乎全在畫布左外", x+57)
	}
	if x := colonySatelliteX(8); x < moo2ScreenW {
		t.Errorf("第 8 顆 x = %d 應該整個在畫布右外", x)
	}
}
