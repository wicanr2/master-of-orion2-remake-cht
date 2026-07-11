package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// colonization.go:玩家用殖民船(Colony Ship)在無主適居星建立新殖民地的最小可玩流程
// (ColonizeStar)。這是「能玩完整一局」目前最大的缺口——原本玩家只有母星、無法擴張。
//
// --- 硬門檻依據(GAME_MANUAL.pdf,moo2_patch1.5 隨附完整手冊,pdftotext -layout 直接萃取
// 文字,非 OCR,見 docs/tech/colonization.md §1 完整引文) ---
//
//  1. 適居性(p.55「Planets」小節):"Planets come into two different categories: gas giants
//     and habitable worlds... colonies can only survive on a solid planet." p.61「Creation」
//     小節:"A Colony Ship can establish a colonial foothold on any uncolonized planet in its
//     range, as long as all space monsters and enemy ships have been cleared from that
//     planet's system." —— 換言之,非氣態巨星/小行星帶的「一般行星」(gamedata.HABITABLE)
//     一律可由殖民船直接殖民,不需要額外科技;氣態巨星(gamedata.GAS_GIANT)/小行星帶
//     (gamedata.ASTEROIDS)則只能建軍事前哨(Outpost Ship),要另外的科技才能讓前哨「支援殖民
//     地」(p.50 節錄:「該科技允許同系統內的氣態巨星/小行星帶前哨升級為可住人殖民地,行星固定
//     Barren/Normal-G/Abundant,氣態巨星化為 Huge、小行星帶化為 Large」)。
//
//     本 remake 的星系生成(session.go genGalaxy/genPlanets)目前只有「一般行星」這一種行星
//     資料型別——每顆星固定生成一顆行星,從未產生氣態巨星/小行星帶(gamedata.PlanetType 這個
//     enum 雖然存在,但 genPlanets 完全沒有使用它來標記行星類別)。因此「哪些星需要額外科技」
//     這個問題目前沒有實際案例可套用:climateColonizable 保留為未來擴充掛勾點,現階段恆真。
//
//  2. 新殖民地起始狀態(p.61-62「Creation」小節,直接引文,非猜測):
//       - Colony Base:"the new colony is established with one unit of population."
//       - Colony Ship:"a new unit of population is gathered to board the ship."
//     兩種建立方式起始人口一致 = 1。手冊全文未提及新殖民地會自動附帶任何建築(對照母星
//     homeworldBuildings() 明確列出海軍陸戰隊營+星基是「Pre-warp/Average Tech games only」
//     的特例),故新殖民地起始建築為空(nil map),與手冊沉默一致,不臆造。
//
//     初始工作分配:手冊未提供任何規則。population=1 時選擇「全農」(而非比照
//     session.go advancePopulation「新增人口預設分配為工人」的既有慣例)——後者是「已有farmer
//     且經濟穩定的殖民地,人口成長 +1 時」的慣例,套用在population=1、Farmers=0 的全新殖民地會
//     讓 Food=0(FoodPerFarmer 乘 0 個農夫)但 FoodConsumed=1,首回合即饑荒
//     (session.go recoverFromFamine 下回合才會修正回 farmer)。「全農」是任務指示明列的簡單
//     保守預設之一,避免這個不必要的首回合饑荒瞬間,標記於此供 L.CY 檢視。
//
//  3. PopMax:gamedata.PlanetBasePopMax(size, climate),公式移植自 openorion2
//     gamestate.cpp:2288 GameState::planetMaxPop,已與手冊 p.55-56 各尺寸人口範圍交叉驗證
//     (見該函式註解逐項推導),非本檔案臆造。
//
// TODO(未來擴充,不阻塞本輪):
//   - 若之後補上氣態巨星/小行星帶的星系生成,ColonizeStar 需要另外 gate 這兩類行星,要求先解鎖
//     對應科技(p.50 節錄提及但手冊本節未逐字列出科技名稱,待從 techtree.go/patch 1.5 changelog
//     查證)才能把 Outpost 升級成殖民地——climateColonizable 是這個 gate 的掛勾點。
//   - 目前只殖民「該星生成的那顆行星」(session.go genPlanets 每星恆一顆),不支援「同系統多顆
//     行星選擇殖民哪顆」(手冊原文的 System 視窗選擇畫面)——本輪任務明確排除行星選擇子畫面。

// ColonyShipClass 是殖民船的艦體等級字串(見 session.go homeworldShips()/shipStrength 既有
// 命名慣例:{"拓荒號", "殖民船", ...})。
const ColonyShipClass = "殖民船"

// colonizeStartPopulation 是新殖民地起始人口(見檔頭§2 手冊引文:Colony Base/Colony Ship
// 皆為 1 單位人口),高信心手冊數字,非猜測。
const colonizeStartPopulation = 1

// findColonyShipIndex 回傳玩家艦隊中第一艘殖民船在 s.Ships 的索引;找不到回 -1。
func (s *GameSession) findColonyShipIndex() int {
	for i, sh := range s.Ships {
		if sh.Class == ColonyShipClass {
			return i
		}
	}
	return -1
}

// FleetHasColonyShip 回傳玩家艦隊是否載有至少一艘殖民船(供 UI 判斷是否顯示「拓殖」按鈕,
// 見 cmd/moo2/interactive.go galaxy() 的 "colonize" 熱區判斷)。匯出(大寫),因為 cmd/moo2 是
// 獨立套件,只能呼叫 shell 的匯出方法。
func (s *GameSession) FleetHasColonyShip() bool {
	return s.findColonyShipIndex() >= 0
}

// --- shell.Planet 顯示字串 → gamedata 型別對映 ---
//
// session.go genPlanets 產生的 Planet.Climate/Gravity/Mineral/Size 是「供行星列表顯示」的中文
// 字串(該函式註解:「正式版由存檔/星系生成填真值」),與 gamedata 的 PlanetClimate/Gravity/
// Minerals/Size 型別化 enum 是兩套獨立資料(字串陣列僅由 Star.Spectral/Size 衍生的展示用途,
// 從未 import gamedata)。ColonizeStar 需要把「玩家在行星列表看到的那個值」轉成 engine.ColonyState
// 需要的型別化欄位,兩者必須一致——否則玩家會看到「氣候:海洋」卻套用了完全不同的內部氣候規則。
// 以下四個對映表直接對應 genPlanets 裡各自的字串陣列(climates/gravs/minerals/sizes),逐一
// 手動核對到 gamedata enum 語意最接近的值;找不到對映(理論上不會發生,除非 genPlanets 改字串
// 卻忘了同步這裡)回 ok=false,呼叫端保守拒絕拓殖,不猜測。

// climateDisplayToGamedata 對映 genPlanets 的 climates 陣列(僅 7 種,缺 Swamp/Arid/Terran/
// Gaia——那 4 種是「地形改造終點/母星專屬」氣候,無主隨機星本就不生成,見 genPlanets 註解)。
// 「地獄」無手冊對應氣候名稱(genPlanets 用作 Spectral=6/黑洞星系的敘事填充),保守映射到手冊
// 定性最惡劣的 TOXIC(p.58:"Farming is impossible"),不臆造新的氣候值。
var climateDisplayToGamedata = map[string]gamedata.PlanetClimate{
	"放射": gamedata.RADIATED,
	"貧瘠": gamedata.BARREN,
	"海洋": gamedata.OCEAN,
	"沙漠": gamedata.DESERT,
	"凍原": gamedata.TUNDRA,
	"有毒": gamedata.TOXIC,
	"地獄": gamedata.TOXIC, // 見上方註解:無手冊對應,保守取最惡劣氣候,非新發明的氣候值
}

// gravityDisplayToGamedata 對映 genPlanets 的 gravs 陣列(低/常態/高,恰好三種,與
// gamedata.PlanetGravity 一一對應,無需近似)。
var gravityDisplayToGamedata = map[string]gamedata.PlanetGravity{
	"低":  gamedata.LOW_G,
	"常態": gamedata.NORMAL_G,
	"高":  gamedata.HEAVY_G,
}

// mineralDisplayToGamedata 對映 genPlanets 的 minerals 陣列(4 種:貧瘠/一般/豐富/富饒),
// gamedata.PlanetMinerals 有 5 級(多一級 ULTRA_POOR)。「貧瘠」映射到 POOR 而非 ULTRA_POOR——
// genPlanets 沒有「極度貧瘠」這個顯示詞,POOR 是字面最接近的一級,非隨意選擇。
var mineralDisplayToGamedata = map[string]gamedata.PlanetMinerals{
	"貧瘠": gamedata.POOR,
	"一般": gamedata.ABUNDANT,
	"豐富": gamedata.RICH,
	"富饒": gamedata.ULTRA_RICH,
}

// sizeDisplayToGamedata 對映 genPlanets 的 sizes 陣列(4 種:巨大/大型/中型/小型),
// gamedata.PlanetSize 有 5 級(多一級 TINY_PLANET)。genPlanets 從未生成「小型」以下的行星,
// 故 TINY_PLANET 這裡用不到,不影響 ColonizeStar 的實際案例。
var sizeDisplayToGamedata = map[string]gamedata.PlanetSize{
	"巨大": gamedata.HUGE_PLANET,
	"大型": gamedata.LARGE_PLANET,
	"中型": gamedata.MEDIUM_PLANET,
	"小型": gamedata.SMALL_PLANET,
}

func climateFromDisplay(s string) (gamedata.PlanetClimate, bool) {
	c, ok := climateDisplayToGamedata[s]
	return c, ok
}

func gravityFromDisplay(s string) (gamedata.PlanetGravity, bool) {
	g, ok := gravityDisplayToGamedata[s]
	return g, ok
}

func mineralFromDisplay(s string) (gamedata.PlanetMinerals, bool) {
	m, ok := mineralDisplayToGamedata[s]
	return m, ok
}

func sizeFromDisplay(s string) (gamedata.PlanetSize, bool) {
	sz, ok := sizeDisplayToGamedata[s]
	return sz, ok
}

// climateColonizable 回傳該氣候是否可由殖民船直接殖民,不需額外科技。見檔頭§1:本 remake 的
// 星系生成從不產生氣態巨星/小行星帶,故目前傳入的 climate 恆為 TOXIC..GAIA 範圍、恆為 true——
// 保留這個函式是給未來補上氣態巨星/小行星帶星系生成時的 gate 掛勾點,不是把「一律可殖民」這個
// 目前恰好成立的簡化結論直接寫死散落在 ColonizeStar 內部。
func climateColonizable(c gamedata.PlanetClimate) bool {
	return c >= gamedata.TOXIC && c <= gamedata.GAIA
}

// ColonizationResult 是一次拓殖嘗試的結果(供 UI/測試檢視),命名/欄位風格對稱
// ground_invasion.go 的 GroundInvasionResult。
type ColonizationResult struct {
	Ok              bool   // 是否成功建立殖民地(false = 前置條件不足,未消耗任何狀態)
	Reason          string // Ok=false 時的原因(供 UI 提示;Ok=true 時為空字串)
	ColonyIndex     int    // Ok=true 時,新殖民地在 s.PlayerColonies 的索引
	StartPopulation int    // Ok=true 時,新殖民地起始人口(見 colonizeStartPopulation)
	PopMax          int    // Ok=true 時,新殖民地人口上限(見 gamedata.PlanetBasePopMax)
}

// ColonizeStar 嘗試在 starIdx 這顆星建立新殖民地。前置條件:
//  1. 玩家艦隊已抵達該星(FleetAtStar==starIdx 且 FleetETA==0,航行中不能發動,比照 InvadeColony)。
//  2. 該星目前無主(Owner==0)——已被玩家或 AI 佔領的星不可再拓殖。
//  3. 玩家艦隊載有至少一艘殖民船(findColonyShipIndex 找得到)。
//  4. starIdx 對應的行星資料可辨識、氣候可直接殖民(見 climateColonizable)。
//
// 任一條件不足回傳 Ok=false + Reason,不消耗任何狀態(不扣殖民船、不改 Star.Owner)。
//
// 成功:依 starIdx 對應的 Planets[starIdx](climate/gravity/mineral/size 字串轉 gamedata 型別,
// 見上方對映表)建一筆新 engine.ColonyState——起始人口 colonizeStartPopulation、全農(見檔頭§2
// 理由)、FoodPerFarmer/IndustryPerWorker 依環境查表 + 玩家種族加成(Races[s.RaceIndex],比照
// ApplyRace 對既有殖民地的加成邏輯——ApplyRace 只在新遊戲開局套一次,不會回頭套用到之後才建立
// 的殖民地,故這裡手動疊加一次)、士氣依目前政府 + 無建築(colonyMoralePercent(s.Government,
// nil))。append 進 PlayerColonies + 所有平行陣列(Builds/ColonyBuildings/PlayerColonyMarines/
// MarineBarracksAge/PlayerColonyTanks/ArmorBarracksAge/popAccum/PlayerColonyStars,padding 模式
// 比照 InvadeColony 既有慣例),Star.Owner 轉 1,並從 s.Ships 移除用掉的那艘殖民船。
func (s *GameSession) ColonizeStar(starIdx int) ColonizationResult {
	if starIdx < 0 || starIdx >= len(s.Stars) {
		return ColonizationResult{Reason: "無效的星索引"}
	}
	if s.FleetAtStar != starIdx || s.FleetETA != 0 {
		return ColonizationResult{Reason: "艦隊尚未抵達該星"}
	}
	star := &s.Stars[starIdx]
	if star.Owner != 0 {
		return ColonizationResult{Reason: "該星已有歸屬,不可拓殖"}
	}
	shipIdx := s.findColonyShipIndex()
	if shipIdx < 0 {
		return ColonizationResult{Reason: "艦隊未載運殖民船"}
	}
	if starIdx >= len(s.Planets) {
		return ColonizationResult{Reason: "無行星資料(不應發生)"}
	}
	planet := s.Planets[starIdx]

	climate, ok := climateFromDisplay(planet.Climate)
	if !ok {
		return ColonizationResult{Reason: "行星氣候資料無法辨識(不應發生,見 climateDisplayToGamedata)"}
	}
	if !climateColonizable(climate) {
		return ColonizationResult{Reason: "此類行星需額外科技才能建立殖民地(氣態巨星/小行星帶,尚未支援)"}
	}
	gravity, ok := gravityFromDisplay(planet.Gravity)
	if !ok {
		gravity = gamedata.NORMAL_G // 不應發生的保守預設,見 gravityDisplayToGamedata
	}
	mineral, ok := mineralFromDisplay(planet.Mineral)
	if !ok {
		mineral = gamedata.POOR // 不應發生的保守預設,見 mineralDisplayToGamedata
	}
	size, ok := sizeFromDisplay(planet.Size)
	if !ok {
		size = gamedata.MEDIUM_PLANET // 不應發生的保守預設,見 sizeDisplayToGamedata
	}

	foodPerFarmer := gamedata.ClimateFoodPerFarmer(climate)
	industryPerWorker := gamedata.MineralIndustryPerWorker(mineral)
	researchPerScientist := 30 // 見 playerHomeworldColony 註解:手冊無環境相關公式,remake 沿用同一基準值
	if s.RaceIndex >= 0 && s.RaceIndex < len(Races) {
		r := Races[s.RaceIndex]
		foodPerFarmer += r.FoodBonus
		industryPerWorker += r.IndBonus
		researchPerScientist += r.ResBonus
	}

	popMax := gamedata.PlanetBasePopMax(size, climate)
	if popMax < colonizeStartPopulation {
		popMax = colonizeStartPopulation // 保底:新殖民地的人口上限不能低於起始人口本身
	}

	colony := engine.ColonyState{
		Population:           colonizeStartPopulation,
		PopMax:               popMax,
		Farmers:              colonizeStartPopulation, // 全農,見檔頭§2 理由(避免首回合饑荒)
		FoodPerFarmer:        foodPerFarmer,
		IndustryPerWorker:    industryPerWorker,
		ResearchPerScientist: researchPerScientist,
		PlanetSize:           size,
		PlanetGravity:        gravity,
		MineralRichness:      mineral,
		Climate:              climate,
		MoralePercent:        colonyMoralePercent(s.Government, nil), // 新殖民地無任何建築,見檔頭§2
	}

	s.PlayerColonies = append(s.PlayerColonies, colony)
	idx := len(s.PlayerColonies) - 1
	s.Builds = append(s.Builds, ColonyBuild{})
	for len(s.ColonyBuildings) < len(s.PlayerColonies) {
		s.ColonyBuildings = append(s.ColonyBuildings, nil)
	}
	for len(s.PlayerColonyMarines) < len(s.PlayerColonies) {
		s.PlayerColonyMarines = append(s.PlayerColonyMarines, 0)
	}
	for len(s.MarineBarracksAge) < len(s.PlayerColonies) {
		s.MarineBarracksAge = append(s.MarineBarracksAge, 0)
	}
	for len(s.PlayerColonyTanks) < len(s.PlayerColonies) {
		s.PlayerColonyTanks = append(s.PlayerColonyTanks, 0)
	}
	for len(s.ArmorBarracksAge) < len(s.PlayerColonies) {
		s.ArmorBarracksAge = append(s.ArmorBarracksAge, 0)
	}
	for len(s.popAccum) < len(s.PlayerColonies) {
		s.popAccum = append(s.popAccum, 0)
	}
	for len(s.PlayerColonyStars) < len(s.PlayerColonies)-1 {
		s.PlayerColonyStars = append(s.PlayerColonyStars, -1) // 補齊先前未同步的空缺(語意:星索引未知)
	}
	s.PlayerColonyStars = append(s.PlayerColonyStars, starIdx)

	star.Owner = 1
	s.Ships = append(s.Ships[:shipIdx], s.Ships[shipIdx+1:]...) // 消耗這艘殖民船

	return ColonizationResult{Ok: true, ColonyIndex: idx, StartPopulation: colonizeStartPopulation, PopMax: popMax}
}
