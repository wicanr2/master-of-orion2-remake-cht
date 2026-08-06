package gamedata

import "math/rand"

// 星系/行星生成:原版 Orion2.exe 的權威骰表與生成規則。
//
// 來源:反組譯 Orion2.exe(patch 1.31)內建的 Watcom 除錯符號表 + Hex-Rays/組語,
// 見 docs/re/00-orion2-symbols.md。每張表都經「內容 + 讀它的函式」雙重確認:
// 表的語意不是靠名字猜的,是靠「原版哪個函式用什麼索引讀它」釘死的。
//
// 為什麼要換掉 remake 原本的生成:remake 先前用 rand.Intn(7) 均勻擲光譜、
// rand.Intn(4) 擲行星大小、氣候直接拿光譜當索引,每顆星一顆行星。
// 對照 archive.org 原版實測(docs/tech/oracle-comparison-20260712.md)星圖明顯不同——
// 原版紅矮星佔 47%、宜居行星集中在黃/橙星的中段軌道,而不是均勻散布。
//
// 原版生成鏈(函式名為符號表原名):
//
//	Generate_Spectral_Class_  → 恆星光譜(依星系年齡加權)
//	Generate_Size_            → 行星大小(d10 累計骰表)
//	Generate_Mineral_Class_   → 礦產豐度(d10 × 光譜)
//	Generate_Gravity_Class_   → 重力(礦產 × 大小,即「密度 × 體積」)
//	Get_Planet_Group_         → [光譜][軌道] → 溫度帶 0..3
//	Generate_Climate_         → 溫度帶 → 十氣候加權骰
//	Generate_Max_Farms_       → 農場上限(依大小)

// GalaxyAge 是新遊戲的星系年齡設定(原版 Generate_Spectral_Class_ 的參數,
// 以及 Generate_Climate_ 讀的全域 byte_199F2F)。
type GalaxyAge int

const (
	GalaxyYouthful GalaxyAge = 0 // 年輕:藍白星多,宜居星少
	GalaxyAverage  GalaxyAge = 1
	GalaxyMature   GalaxyAge = 2 // 成熟:黃橙星多,宜居星多
)

// StarClassWeights 是各光譜類別的出現權重,依星系年齡分三欄。
// 原版 `_star_class_table` @ 0x17D83E(21 bytes),
// Generate_Spectral_Class_(sub_8C807)以 `table[spectral*3 + age]` 取 7 個權重後
// 交給 Get_Weighted_Choice_Char_ 加權隨機。
//
// 三欄各自總和 = 100(20+25+10+10+32+1+2 / 10+15+16+16+37+2+4 / 5+5+30+21+30+3+6),
// 這是表結構正確的獨立驗證。
var StarClassWeights = [7][3]int{
	{20, 10, 5},  // Blue
	{25, 15, 5},  // White
	{10, 16, 30}, // Yellow
	{10, 16, 21}, // Orange
	{32, 37, 30}, // Red — 一般星系近半數是紅矮星
	{1, 2, 3},    // Brown dwarf
	{2, 4, 6},    // Black hole
}

// PlanetSizeRollTable 是行星大小的 **d10 累計** 骰表(不是權重)。
// 原版 `_planet_size_table` @ 0x17D7F7 = {1,3,7,9,10};
// Generate_Size_(sub_8C7D1)骰 1..10,由小到大找第一個 `roll <= table[i]` 的 i。
//
//	roll 1     → Tiny   (10%)
//	roll 2–3   → Small  (20%)
//	roll 4–7   → Medium (40%)
//	roll 8–9   → Large  (20%)
//	roll 10    → Huge   (10%)
var PlanetSizeRollTable = [5]int{1, 3, 7, 9, 10}

// ClassToGroup 把 [光譜][軌道] 對應到「溫度帶」0..3,再由溫度帶決定氣候骰表那一欄。
// 原版 `_class_to_group` @ 0x17D761(30 bytes = 6×5),
// Get_Planet_Group_(sub_8CD55)以 `table[spectral*5 + orbit]` 讀。
// 黑洞(spectral 6)沒有行星,故只有 6 列。
//
// 0 = 最靠恆星的高溫帶、2 = 宜居帶、3 = 遠離恆星的低溫帶。
// 藍星整條軌道幾乎都是 0(輻射過強),紅矮星從軌道 3 起就是 3(太冷)——
// 這就是原版「黃/橙星才有好行星」的來源。
var ClassToGroup = [6][5]int{
	{0, 0, 0, 0, 1}, // Blue
	{0, 0, 1, 2, 3}, // White
	{0, 1, 2, 2, 3}, // Yellow
	{1, 2, 2, 2, 3}, // Orange
	{1, 2, 3, 3, 3}, // Red
	{0, 0, 1, 2, 3}, // Brown dwarf
}

// NormalGalaxyClimateWeights 是氣候加權骰表 [氣候][溫度帶]。
// 原版 `_normal_gal_climate_roll_table` @ 0x17D7A7(40 bytes = 10×4),
// Generate_Climate_(sub_8BEAB)以 `table[climate*4 + group]` 取 10 個權重加權骰。
// 星系年齡 Youthful/Average 用這張。
//
// 四欄權重和 = 100 / 105 / 100 / 100(欄1 的 105 是原版原值,加權骰會自行正規化)。
var NormalGalaxyClimateWeights = [10][4]int{
	{15, 15, 10, 20}, // Toxic
	{55, 50, 15, 0},  // Radiated
	{25, 25, 10, 70}, // Barren
	{5, 10, 10, 0},   // Desert
	{0, 5, 10, 8},    // Tundra
	{0, 0, 10, 2},    // Ocean
	{0, 0, 11, 0},    // Swamp
	{0, 0, 11, 0},    // Arid
	{0, 0, 11, 0},    // Terran
	{0, 0, 2, 0},     // Gaia
}

// OldGalaxyClimateWeights 是成熟星系(GalaxyMature)用的氣候骰表。
// 原版 `_old_gal_climate_roll_table` @ 0x17D7CF(40 bytes = 10×4),四欄總和皆為 100。
// 與 Normal 相比:宜居帶(欄2)的可農作氣候合計 65 → 79,Gaia 從 2% 升到 4%。
var OldGalaxyClimateWeights = [10][4]int{
	{15, 5, 5, 20},  // Toxic
	{40, 30, 8, 0},  // Radiated
	{20, 20, 8, 50}, // Barren
	{25, 25, 13, 0}, // Desert
	{0, 20, 13, 30}, // Tundra
	{0, 0, 13, 0},   // Ocean
	{0, 0, 13, 0},   // Swamp
	{0, 0, 13, 0},   // Arid
	{0, 0, 10, 0},   // Terran
	{0, 0, 4, 0},    // Gaia
}

// GravityTable 是 [礦產豐度][行星大小] → 重力等級(0=Low, 1=Normal, 2=Heavy)。
// 原版 `_gravity_table` @ 0x17D72A(25 bytes = 5×5),
// Generate_Gravity_Class_(sub_8BFE0)以 `table[mineral*5 + size]` 讀。
//
// 物理上自洽:礦產豐度代表密度、大小代表體積,兩者都往上就重力大。
var GravityTable = [5][5]int{
	{0, 0, 0, 1, 1}, // Ultra Poor
	{0, 0, 1, 1, 1}, // Poor
	{0, 1, 1, 1, 2}, // Abundant
	{1, 1, 1, 2, 2}, // Rich
	{1, 1, 2, 2, 2}, // Ultra Rich
}

// ClassToMineral 是 [d10 骰值-1][光譜] → 礦產豐度(0..4)。
// 原版 `_class_to_mineral` @ 0x17D6EE(60 bytes = 10×6),
// Generate_Mineral_Class_(sub_8C05B)以 `table[(Random(10)-1)*6 + spectral]` 讀。
// 藍星最富、紅矮星最貧。
var ClassToMineral = [10][6]int{
	{2, 1, 1, 0, 0, 1},
	{2, 1, 1, 1, 0, 2},
	{2, 2, 1, 1, 1, 2},
	{2, 2, 2, 1, 1, 2},
	{3, 2, 2, 1, 1, 2},
	{3, 2, 2, 2, 1, 2},
	{3, 3, 2, 2, 2, 3},
	{3, 3, 3, 2, 2, 3},
	{4, 3, 3, 2, 2, 3},
	{4, 4, 4, 3, 2, 4},
}

// ClassToNumSatellites 是 [d10 骰值-1][光譜] → 該星系的行星/衛星數量。
// 原版 `_class_to_num_satellites` @ 0x17D680(60 bytes = 10×6),
// Generate_Number_Of_Satellites_(sub_8C527)以 `table[(Random(10)-1)*6 + spectral]` 讀,
// 光譜 >= 6(黑洞)直接回 0。
var ClassToNumSatellites = [10][6]int{
	{0, 0, 1, 2, 0, 0},
	{1, 1, 2, 2, 1, 0},
	{1, 1, 2, 2, 1, 0},
	{2, 1, 2, 3, 1, 0},
	{3, 2, 3, 3, 2, 0},
	{3, 2, 3, 4, 2, 0},
	{4, 3, 4, 4, 2, 0},
	{4, 3, 4, 5, 3, 1},
	{5, 4, 5, 5, 3, 1},
	{5, 4, 5, 5, 4, 1},
}

// PlanetMaxFarms / PlanetMaxMines 是各行星大小可蓋的農場/礦場上限。
// 原版 `_planet_max_farms` @ 0x17D7FC、`_planet_max_mines` @ 0x17D801,
// Generate_Max_Farms_(sub_8C01E)以 `table[size]` 讀。
// remake 目前沒有這層上限概念,先建表待接。
var (
	PlanetMaxFarms = [5]int{2, 4, 5, 7, 10}
	PlanetMaxMines = [5]int{2, 4, 6, 9, 12}
)

// PlanetMaxPopulationBase 是各行星大小的基礎人口上限(未乘氣候係數)。
// 原版 `_planet_max_population` @ 0x17D806 = {5,10,15,20,25}。
// 這**獨立證實**了 remake 既有的 PlanetBasePopMax 用的 `(size+1)*5`。
var PlanetMaxPopulationBase = [5]int{5, 10, 15, 20, 25}

// RollSpectralClass 依星系年齡加權擲一個恆星光譜類別(原版 Generate_Spectral_Class_)。
func RollSpectralClass(r *rand.Rand, age GalaxyAge) SpectralClass {
	a := int(age)
	if a < 0 || a > 2 {
		a = int(GalaxyAverage)
	}
	w := make([]int, 7)
	for i := range w {
		w[i] = StarClassWeights[i][a]
	}
	return SpectralClass(WeightedChoice(r, w))
}

// RollPlanetSize 把 1..10 的骰值換成行星大小(原版 Generate_Size_ 的累計骰表查法)。
// 骰值超出範圍時夾回 1..10,不回傳原版的 -1(remake 沒有「無行星」這個回傳路徑)。
func RollPlanetSize(roll int) PlanetSize {
	if roll < 1 {
		roll = 1
	}
	if roll > 10 {
		roll = 10
	}
	for i := 0; i < 5; i++ {
		if roll <= PlanetSizeRollTable[i] {
			return PlanetSize(i)
		}
	}
	return HUGE_PLANET
}

// PlanetOrbitGroup 回傳 [光譜][軌道] 對應的溫度帶 0..3(原版 Get_Planet_Group_)。
// 黑洞或軌道越界時回 0(同原版的邊界處理)。orbit 以 0 起算。
func PlanetOrbitGroup(sc SpectralClass, orbit int) int {
	if int(sc) < 0 || int(sc) >= 6 || orbit < 0 || orbit >= 5 {
		return 0
	}
	return ClassToGroup[int(sc)][orbit]
}

// RollClimate 依溫度帶與星系年齡擲氣候(原版 Generate_Climate_)。
//
// requireHabitable 對應原版的 byte_19C31A 分支:開啟時
// ①若該溫度帶只剩一個「有食物產出」的氣候就直接選它 ②否則重骰直到骰到可農作的氣候。
// 原版用它保證母星那類「必須宜居」的星球不會骰出 Toxic。
func RollClimate(r *rand.Rand, group int, age GalaxyAge, requireHabitable bool) PlanetClimate {
	if group < 0 || group >= 4 {
		group = 0
	}
	tbl := NormalGalaxyClimateWeights
	if age == GalaxyMature {
		tbl = OldGalaxyClimateWeights
	}
	w := make([]int, 10)
	for i := range w {
		w[i] = tbl[i][group]
	}

	if requireHabitable {
		// 先數「權重非 0 且該氣候有食物產出」的候選;只剩一個就直接給,不必重骰。
		cnt, last := 0, 0
		for i := 0; i < 10; i++ {
			if w[i] != 0 && ClimateFoodPerFarmer(PlanetClimate(i)) > 0 {
				cnt++
				last = i
			}
		}
		if cnt == 1 {
			return PlanetClimate(last)
		}
		if cnt == 0 {
			// 該溫度帶完全沒有可農作氣候(如藍星內側軌道)。原版在這裡會無窮重骰,
			// remake 不能掛住:退回不強制,由加權骰決定。
			return PlanetClimate(WeightedChoice(r, w))
		}
		for {
			c := WeightedChoice(r, w)
			if ClimateFoodPerFarmer(PlanetClimate(c)) > 0 {
				return PlanetClimate(c)
			}
		}
	}
	return PlanetClimate(WeightedChoice(r, w))
}

// RollMineralClass 依骰值與光譜回傳礦產豐度(原版 Generate_Mineral_Class_)。roll 為 1..10。
func RollMineralClass(roll int, sc SpectralClass) PlanetMinerals {
	if int(sc) < 0 || int(sc) >= 6 {
		return ABUNDANT
	}
	i := clampRoll10(roll) - 1
	return PlanetMinerals(ClassToMineral[i][int(sc)])
}

// RollNumSatellites 依骰值與光譜回傳該恆星的行星數(原版 Generate_Number_Of_Satellites_)。
// 黑洞回 0。roll 為 1..10。
func RollNumSatellites(roll int, sc SpectralClass) int {
	if int(sc) < 0 || int(sc) >= 6 {
		return 0
	}
	return ClassToNumSatellites[clampRoll10(roll)-1][int(sc)]
}

// PlanetGravityFor 由礦產豐度(密度)與行星大小(體積)決定重力(原版 Generate_Gravity_Class_)。
func PlanetGravityFor(mineral PlanetMinerals, size PlanetSize) PlanetGravity {
	m, s := int(mineral), int(size)
	if m < 0 || m >= 5 {
		m = int(ABUNDANT)
	}
	if s < 0 || s >= 5 {
		s = int(MEDIUM_PLANET)
	}
	return PlanetGravity(GravityTable[m][s])
}

// WeightedChoice 依權重挑一個索引(原版 Get_Weighted_Choice_Char_ @ 0xFE8DA 的對應物)。
// 權重全 0 或空陣列時回 0,避免呼叫端要處理 -1。
func WeightedChoice(r *rand.Rand, w []int) int {
	sum := 0
	for _, v := range w {
		if v > 0 {
			sum += v
		}
	}
	if sum <= 0 || len(w) == 0 {
		return 0
	}
	n := r.Intn(sum)
	for i, v := range w {
		if v <= 0 {
			continue
		}
		if n < v {
			return i
		}
		n -= v
	}
	return len(w) - 1
}

func clampRoll10(roll int) int {
	if roll < 1 {
		return 1
	}
	if roll > 10 {
		return 10
	}
	return roll
}

// --- 行星類別:氣態巨星 / 小行星帶 / 一般行星 ---
//
// 原版 `Generate_Satellite_Type_` @ 0x8C6FE 決定每條軌道上的天體是什麼:
//
//	roll = Random_(10) - 1                       // 0..9(Random_ 回 1..n,見下方註記)
//	t    = _orbit_to_satellite_type[roll*5 + orbit]
//	if t == 4 { …特例,見 OrbitSatelliteSpecialRoll… }
//	if t == GAS_GIANT && orbit == 0 { 整個重骰 }  // 最內圈不放氣態巨星
//
// 索引方式怎麼確定的:表是 50 bytes,而 `_class_to_num_satellites` 之後緊接著它,
// 大小正好 10 列 × 5 軌道。決定性的佐證是那個唯一的 `4`——它只出現在 (roll 1, orbit 0),
// 而原版處理 4 的特例分支寫死了 `bl == 1 && orbit == 0`。兩邊完全咬合,索引不可能是別的擺法。
//
// ⚠ `Random_` 的語意(0x1247A0):LCG 取樣 + 拒絕超界值,最後 `div bucket` 再 **`inc eax`**,
// 所以回的是 **1..n**,不是 C 慣例的 0..n-1。`roll = Random_(10) - 1` 因此落在 0..9,
// 剛好對上 10 列——這也是「表是 10×5 不是 5×10」的另一個佐證。

// planetTypeSpecialMarker 是表裡的 4:不是行星類別,是「特例處理」的標記(見
// OrbitSatelliteSpecialRoll)。真正的行星類別只有 1/2/3(對齊 openorion2 `enum PlanetType`)。
const planetTypeSpecialMarker = 4

// orbitToSatelliteType 是原版 `_orbit_to_satellite_type` @ 0x17D6BC(50 bytes)。
// 列 = Random_(10)-1(0..9),欄 = 軌道(0..4,由內而外)。
// 值:1 小行星帶、2 氣態巨星、3 一般行星、4 特例標記。
//
// 讀出來的分布本身就有物理直覺:內圈多岩石/小行星、外圈多氣態巨星,
// 而 roll 值越大整個系統越「宜居」(roll 6-9 五條軌道全是一般行星)。
var orbitToSatelliteType = [10][5]int{
	{1, 1, 1, 1, 1}, // roll 0:整組小行星帶
	{4, 1, 1, 1, 2}, // roll 1:內圈特例,外圈氣態巨星
	{3, 2, 1, 2, 2},
	{3, 3, 2, 2, 2},
	{3, 3, 2, 2, 2},
	{3, 3, 3, 3, 2},
	{3, 3, 3, 3, 3},
	{3, 3, 3, 3, 3},
	{3, 3, 3, 3, 3},
	{3, 3, 3, 3, 3}, // roll 9
}

// OrbitSatelliteSpecialRoll 是 4 這個標記的特例機率(原版 `Random_(100)`,命中 1..10)。
// 也就是 **10%**;沒命中就退回小行星帶。
const OrbitSatelliteSpecialRoll = 10

// RollSatelliteType 依原版表決定某條軌道上的天體類別。
//
//	roll10  = Random_(10) 的結果(1..10)
//	orbit   = 軌道 0..4
//	special = Random_(100) 的結果(1..100),只在命中表裡的 4 時才會用到
//
// 回傳 (類別, 是否為特例天體)。特例天體在原版會依恆星光譜寫進一個 4 以上的類別碼
// (`spectral == 0 ? 5 : spectral + 4`),remake 沒有那些類別的任何行為或美術,
// 一律當成小行星帶處理、另外用 bool 標出來——**不臆造原版那些碼的語意**。
func RollSatelliteType(roll10, orbit, special int) (t PlanetType, isSpecial bool) {
	r := roll10 - 1
	if r < 0 {
		r = 0
	}
	if r > 9 {
		r = 9
	}
	if orbit < 0 {
		orbit = 0
	}
	if orbit > 4 {
		orbit = 4
	}
	v := orbitToSatelliteType[r][orbit]
	if v == planetTypeSpecialMarker {
		if special >= 1 && special <= OrbitSatelliteSpecialRoll {
			return ASTEROIDS, true
		}
		return ASTEROIDS, false
	}
	if PlanetType(v) == GAS_GIANT && orbit == 0 {
		// 原版在這裡整個重骰;呼叫端拿不到新的骰子,退回一般行星(表裡 orbit 0 從來不是 2,
		// 這條路實際走不到,留著是為了不讓「最內圈出現氣態巨星」這個原版明確排除的情況溜過去)。
		return HABITABLE, false
	}
	return PlanetType(v), false
}
