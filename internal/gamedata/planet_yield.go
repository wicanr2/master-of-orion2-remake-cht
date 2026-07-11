package gamedata

// 殖民地基礎產出:由行星屬性(氣候 climate / 礦產豐度 mineral / 重力 gravity)驅動,取代
// remake 先前寫死的 FoodPerFarmer/IndustryPerWorker 固定值。
//
// 來源:`GAME_MANUAL.pdf`(patch 1.5 隨附完整手冊)「Your Home World」章節下的 Mineral
// Richness(p.56-57)、Gravity(p.58)、Climate(p.58-59)三小節,逐字附頁碼於下方各表註解;
// 「Food per farmer」「Industry per worker」「Worker penalty」三詞的官方定義引自同章節
// New Colony 對話框說明(p.63)。查不到手冊數字的一律標「待查」,不編造(專案鐵律,見
// docs/tech/homeworld-init.md 的「誠實待確認清單」慣例)。
//
// 與既有檔案的分工:
//   - `terraform.go` 的 `terraformClimatePopFactor` 是「氣候 → 人口容量係數」,回答「這星球最多
//     住幾人」;本檔的 `climateFoodPerFarmer` 是「氣候 → 每個農業人口的食物產出」,回答「每個農夫
//     種多少糧」——兩者手冊來源不同小節、驅動 ColonyState 不同欄位,不可混用。
//   - `formulas.go` 的 `mineralProductionTable`/`PlanetBaseProduction` 已經是本檔案要的「礦產豐度
//     → 基礎工業產出」表(來源同一手冊 p.56-57,數值 1/2/3/5/8 與 openorion2
//     gamestate.cpp:31 的 mineralProductionTable 逐項相符,交叉驗證高信心)。本檔 `MineralIndustryPerWorker`
//     直接包一層 `PlanetBaseProduction`,不重複造表,避免兩處常數漂移。

// climateFoodPerFarmer 每種氣候(PlanetClimate 0-based:TOXIC..GAIA)每個農業人口單位的基礎
// 食物產出。GAME_MANUAL.pdf p.58-59,「Climate」小節逐條目「Base Food per Unit: N」原文:
//
//	Toxic:0(p.58,"Farming is impossible.")
//	Radiated:0(p.58,"Natural farming is impossible")
//	Barren:0(p.59,"no potential for natural farming")
//	Desert:1(p.59)
//	Tundra:1(p.59)
//	Ocean:2(p.59)
//	Swamp:2(p.59)
//	Arid:1(p.59)
//	Terran:2(p.59)
//	Gaia:3(p.59)
//
// 順序採本專案既有 enums.go 的 PlanetClimate 0-based 順序,與 terraform.go 的
// terraformClimatePopFactor 表一致(同一份手冊、同一氣候排序,已交叉驗證高信心)。
var climateFoodPerFarmer = [10]int{
	0, // TOXIC    (p.58)
	0, // RADIATED (p.58)
	0, // BARREN   (p.59)
	1, // DESERT   (p.59)
	1, // TUNDRA   (p.59)
	2, // OCEAN    (p.59)
	2, // SWAMP    (p.59)
	1, // ARID     (p.59)
	2, // TERRAN   (p.59)
	3, // GAIA     (p.59)
}

// ClimateFoodPerFarmer 回傳指定氣候下,每個農業人口單位的基礎食物產出(GAME_MANUAL.pdf p.58-59,
// "Base Food per Unit")。climate 超出範圍回 0。
//
// 手冊 p.63 對「Food per farmer」的定義:"the base amount of food that each unit of your race's
// population you assign to farming would produce on this planet. This is a function of the
// environment."——本函式即該定義的直接移植,尚未套用種族天賦/建築/士氣加成(那些由呼叫端疊加,
// 與既有 applyBuildingEffect/ApplyRace 的疊加慣例一致)。
func ClimateFoodPerFarmer(climate PlanetClimate) int {
	if climate < 0 || int(climate) >= len(climateFoodPerFarmer) {
		return 0
	}
	return climateFoodPerFarmer[climate]
}

// MineralIndustryPerWorker 回傳指定礦產豐度下,每個工業人口單位的基礎產能(GAME_MANUAL.pdf
// p.56-57,"Base Industry per Unit")。直接包一層既有 `PlanetBaseProduction`(formulas.go),
// 不重複造表——兩者是同一份手冊小節、同一組數字(1/2/3/5/8),已與 openorion2
// gamestate.cpp:31 mineralProductionTable 交叉驗證。
//
// 手冊 p.63 對「Industry per worker」的定義:"the base amount of production that each unit of
// your race's population assigned to work would produce on this world. This is primarily a
// function of mineral abundance."
func MineralIndustryPerWorker(mineral PlanetMinerals) int {
	return PlanetBaseProduction(int(mineral))
}

// ResearchPerScientistNorm 是「銀河基準」每個科學家人口單位的基礎研究產出(GAME_MANUAL.pdf
// 三重確認:①p.949「produces 5 research points instead of the usual 3」明列 usual=3;②p.614
// 「Each Psilon scientist produces 2 more research than the galactic norm」(Psilon=norm+2);
// ③原版存檔 SAVE10.GAM 普通種族母星 ResearchPerScientist=3、Psilon 母星=5)。與食物(氣候表)/
// 工業(礦產表)不同,研究基準不依氣候/礦產,是全域固定值。Psilon 等創造性種族的 +2 由
// Race.ResBonus 疊加(見 shell.Races)。
//
// 注意:先前 playerHomeworldColony 硬編 ResearchPerScientist=30(約 10x 過高,無手冊出處),
// 2026-07-12 依上述三重來源校正為 norm=3——這是開局經濟平衡的載重改動,已探針驗證研究節奏
// 忠實變慢但遊戲仍可進展。
const ResearchPerScientistNorm = 3

// PlanetBasePopMax 回傳指定行星大小(size,0-based TINY..HUGE)、氣候(climate)下的「基礎人口
// 容量」——新殖民地建立當下的 PopMax(不含 Biospheres/+2、Advanced City Planning/+5、
// Subterranean/+2*(size+1)、Tolerant/+25%、Aquatic 等加成,那些由呼叫端疊加,與既有
// ColonyState.PopMax 疊加慣例一致,例如 session.go applyBuildingEffect 的生態圈 +2)。
//
// 公式移植自 openorion2/src/gamestate.cpp:2288 GameState::planetMaxPop 的核心運算式(無條件捨去
// 前的一般種族基準值,忽略該函式裡的 Aquatic/Tolerant/Subterranean/AdvancedCityPlanning 修飾項):
//
//	ret = ((size + 1) * 5 * climateFactor + 50) / 100
//
// climateFactor = TerraformClimatePopFactorPercent(climate)(0-100,terraform.go 既有表,與
// MANUAL_150.html modding 附錄 pop_climate 參數同一份數字,已交叉驗證)。
//
// 已用 GAME_MANUAL.pdf p.55-56「Size」小節逐段給出的人口容量範圍交叉驗證(climateFactor 代入
// 25 與 100 兩端,結果與手冊原文逐字相符,高信心,非臆造):
//
//	Tiny(size=0):  (1*5*25+50)/100=1 .. (1*5*100+50)/100=5   → 手冊「1–5」
//	Small(size=1): (2*5*25+50)/100=3 .. (2*5*100+50)/100=10  → 手冊「3–10」
//	Medium(size=2):(3*5*25+50)/100=4 .. (3*5*100+50)/100=15  → 手冊「4–15」
//	Large(size=3): (4*5*25+50)/100=5 .. (4*5*100+50)/100=20  → 手冊「5–20」
//	Huge(size=4):  (5*5*25+50)/100=6 .. (5*5*100+50)/100=25  → 手冊「6–25」
//
// ⚠ 與既有 playerHomeworldColony()(session.go)的母星 PopMax=20(Large/Terran)不完全相符:
// 代入 size=LARGE_PLANET(3)、climate=TERRAN(climateFactor=80)得 (4*5*80+50)/100=16,非 20。
// 母星的 20 是 docs/tech/homeworld-init.md 既有慣例值(可能含未拆解的起始文明加成),本函式
// 不回頭套用去改動母星既有數字(避免既有經濟平衡 regression),只用於新建殖民地
// (shell.GameSession.ColonizeStar)。size/climate 超出合法範圍時 climateFactor 依
// TerraformClimatePopFactorPercent 的既有邊界規則回 0,本函式不重複做邊界檢查。
func PlanetBasePopMax(size PlanetSize, climate PlanetClimate) int {
	factor := TerraformClimatePopFactorPercent(climate)
	return ((int(size)+1)*5*factor + 50) / 100
}

// gravityPenaltyTable[raceGravity][planetGravity] = 生產產出的百分比懲罰(0 或負值)。
// GAME_MANUAL.pdf p.58「Gravity」小節原文(以一般種族、無重力天賦為基準):
//
//	"Low-G planets have a gravitational pull less than half that of the Earth (1 G). The
//	disorientation and increased number of accidents this causes decrease the output of
//	farmers, scientists, and workers by 25%."
//	"Normal gravity worlds have gravity very close to 1 G. Production rates on these planets
//	are unaffected by gravity."
//	"Heavy-G planets put more than 1.5 G on their inhabitants. All three types of production
//	are reduced by 50%."
//
// 手冊本節文字只描述「行星重力本身」對production的影響,未提及種族自身的重力天賦(Low-G
// World / High-G World,見 docs/tech/custom-race-picks.md)如何平移這個基準。種族天賦的
// 相對關係查表來自 openorion2/src/gamestate.cpp:62-66 的 gravityPenalties[home][dest]
// 陣列(NORMAL_G 種族那一列 {-25, 0, -50} 與手冊原文三句逐項相符,交叉驗證高信心;
// LOW_G/HEAVY_G 種族兩列是 openorion2 對「天賦平移基準點」的具體實作,手冊本節文字沒有
// 直接覆蓋這兩列,視為次一手來源)。
//
// ⚠ 已知版本落差(待查,不裁決):moo2_patch1.5/MANUAL_150.html changelog 記載 1.50 修正一項
// 「Planets Screen Gravity Penalty」顯示 bug——"Low-G planets now show -25% prod for a
// Heavy-G race, instead of -50% prod"。這與 openorion2 表中 [HEAVY_G][LOW_G] = -50 的值
// 字面衝突(該 changelog 暗示修正後應為 -25)。無法判斷 openorion2 反組譯的是修正前還是修正後
// 的版本,也無法排除該 changelog 描述的僅是「顯示文字」bug、與底層產出計算無關的可能性。
// 本表照抄 openorion2 的原始表格值(-50),並在此明確標註這一格存在待查的版本落差,
// 不擅自改動成 changelog 暗示的 -25——需要 DOSBox 兩版本存檔實測才能裁決。
var gravityPenaltyTable = [3][3]int{
	{0, -25, -50}, // LOW_G 種族天賦 (home=LOW_G)
	{-25, 0, -50}, // NORMAL_G 種族(無重力天賦,手冊本節原文對應此列)
	{-50, 0, 0},   // HEAVY_G 種族天賦(home=HEAVY_G;[HEAVY_G][LOW_G]=-50 見上方待查註記)
}

// GravityPenaltyPercent 回傳指定「種族重力天賦」在指定「行星重力」上的生產百分比懲罰
// (0 或負值,套用方式見 GravityAdjustedProduction)。raceGravity 預設種族(無 Low-G/High-G
// 天賦選項)應傳 NORMAL_G。任一參數超出範圍回 0(視同無懲罰,保守預設)。
func GravityPenaltyPercent(planetGravity, raceGravity PlanetGravity) int {
	if raceGravity < 0 || int(raceGravity) >= len(gravityPenaltyTable) {
		return 0
	}
	if planetGravity < 0 || int(planetGravity) >= len(gravityPenaltyTable[0]) {
		return 0
	}
	return gravityPenaltyTable[raceGravity][planetGravity]
}

// GravityAdjustedProduction 套用重力懲罰百分比到基礎產出上(GAME_MANUAL.pdf p.58:懲罰同時
// 套用在 farmers/scientists/workers 三種產出)。算法與既有 `MoraleProductionOutput`(morale.go)
// 同型:`base*(100+penaltyPercent)/100`,無條件捨去。呼叫端目前多半傳 penaltyPercent=0
// (預設種族 + Normal-G 母星,見 session.go averageHomeworldColony 的 TODO 接線點說明)。
func GravityAdjustedProduction(base, penaltyPercent int) int {
	return base * (100 + penaltyPercent) / 100
}
