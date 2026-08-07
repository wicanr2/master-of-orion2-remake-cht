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
//     ⚠ 2026-08-07 訂正:上面這段原本接著寫「本 remake 每顆星固定生成一顆一般行星,從未產生
//     氣態巨星/小行星帶,所以這個問題沒有實際案例」——**兩句都已不成立**。genPlanets 依原版
//     骰表生成 1..5 個天體並標上 TypeID(見 orbit.go),氣態巨星/小行星帶天天出現;而拓殖的
//     對象自 2026-08-07 起是**行星**不是星(ColonizePlanet),同一個星系可以有多個殖民地。
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
// 已了結的兩條 TODO(留著記錄它們的下場,免得再被當成缺口):
//   - 「氣態巨星/小行星帶要另外 gate」→ 做了:newColonyFromPlanet 用 TypeID 擋下來並說明原因,
//     升級成殖民地的科技也認出來了(建築 48 人造行星,見 artificialplanet.go)。
//   - 「不支援同系統選擇殖民哪顆行星」→ 做了:ColonizePlanet 是入口,選行星的畫面就是原版的
//     行星列表(PLNTSUM,右下角 SEND COLONY SHIP / SEND OUTPOST SHIP 那兩顆鈕)。
//     ColonizeStar 保留為「該星系第一顆可殖民行星」的捷徑,星圖面板那顆鈕走它。
//
// 還沒做的:AI 一個星系仍然只會有一個殖民地(aiExpand 只找 Owner==0 的星),見 gap report。

// ColonyShipClass 是殖民船的艦體等級字串(見 session.go homeworldShips()/shipStrength 既有
// 命名慣例:{"拓荒號", "殖民船", ...})。
const ColonyShipClass = "殖民船"

// colonizeStartPopulation 是新殖民地起始人口(見檔頭§2 手冊引文:Colony Base/Colony Ship
// 皆為 1 單位人口),高信心手冊數字,非猜測。
const colonizeStartPopulation = 1

// findColonyShipIndex 回傳玩家艦隊中第一艘殖民船在 s.Fleet().Ships 的索引;找不到回 -1。
func (s *GameSession) findColonyShipIndex() int {
	for i, sh := range s.Fleet().Ships {
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

// 以下四張表是「顯示字串 ↔ gamedata enum」的雙向對照。
//
// 2026-08-06 起 genPlanets 直接把 enum 存進 Planet.*ID,新程式碼應該讀 ID 而不是反解字串;
// 這幾張表現在只有兩個用途:①產生顯示字串 ②把 2026-08-06 之前的舊存檔(Planet.Gen==0,
// 只有字串)回填成 enum。
//
// 涵蓋範圍已補齊到十氣候 / 五大小 / 五礦產 —— 舊版只有 7 氣候 4 大小 4 礦產,
// 因為當時的生成器根本產不出 Swamp/Arid/Terran/Gaia/Tiny/Ultra Poor;
// 換成原版骰表之後這些都會實際出現。

var climateDisplayNames = [10]string{
	"有毒", "放射", "貧瘠", "沙漠", "凍原", "海洋", "沼澤", "乾旱", "類地", "蓋亞",
}

var gravityDisplayNames = [3]string{"低", "常態", "高"}

var mineralDisplayNames = [5]string{"極貧", "貧瘠", "一般", "豐富", "富饒"}

var sizeDisplayNames = [5]string{"微型", "小型", "中型", "大型", "巨大"}

func climateDisplayName(c gamedata.PlanetClimate) string {
	if c < 0 || int(c) >= len(climateDisplayNames) {
		return "未知"
	}
	return climateDisplayNames[c]
}

func gravityDisplayName(g gamedata.PlanetGravity) string {
	if g < 0 || int(g) >= len(gravityDisplayNames) {
		return "常態"
	}
	return gravityDisplayNames[g]
}

func mineralDisplayName(m gamedata.PlanetMinerals) string {
	if m < 0 || int(m) >= len(mineralDisplayNames) {
		return "一般"
	}
	return mineralDisplayNames[m]
}

func sizeDisplayName(s gamedata.PlanetSize) string {
	if s < 0 || int(s) >= len(sizeDisplayNames) {
		return "中型"
	}
	return sizeDisplayNames[s]
}

// climateDisplayToGamedata 反向對照。「地獄」是舊生成器對黑洞星系的敘事填充詞,無手冊對應氣候,
// 保守映射到手冊定性最惡劣的 TOXIC(p.58:"Farming is impossible");保留它純粹為了讀舊存檔。
var climateDisplayToGamedata = map[string]gamedata.PlanetClimate{
	"有毒": gamedata.TOXIC,
	"放射": gamedata.RADIATED,
	"貧瘠": gamedata.BARREN,
	"沙漠": gamedata.DESERT,
	"凍原": gamedata.TUNDRA,
	"海洋": gamedata.OCEAN,
	"沼澤": gamedata.SWAMP,
	"乾旱": gamedata.ARID,
	"類地": gamedata.TERRAN,
	"蓋亞": gamedata.GAIA,
	"地獄": gamedata.TOXIC, // 舊存檔專用,見上方註解
}

var gravityDisplayToGamedata = map[string]gamedata.PlanetGravity{
	"低":  gamedata.LOW_G,
	"常態": gamedata.NORMAL_G,
	"高":  gamedata.HEAVY_G,
}

var mineralDisplayToGamedata = map[string]gamedata.PlanetMinerals{
	"極貧": gamedata.ULTRA_POOR,
	"貧瘠": gamedata.POOR,
	"一般": gamedata.ABUNDANT,
	"豐富": gamedata.RICH,
	"富饒": gamedata.ULTRA_RICH,
}

var sizeDisplayToGamedata = map[string]gamedata.PlanetSize{
	"微型": gamedata.TINY_PLANET,
	"小型": gamedata.SMALL_PLANET,
	"中型": gamedata.MEDIUM_PLANET,
	"大型": gamedata.LARGE_PLANET,
	"巨大": gamedata.HUGE_PLANET,
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

// climateColonizable 回傳該氣候是否可由殖民船直接殖民。
//
// 氣態巨星/小行星帶不是靠氣候擋的——它們有自己的 TypeID(見 newColonyFromPlanet 的類別門檻),
// 這裡只負責「這個氣候值是不是十種合法氣候之一」。之所以還存在,是因為手改過的存檔或
// 未來新增的氣候值不該悄悄通過。
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

// newColonyFromStar 依 starIdx 對應的行星資料(s.Planets[starIdx],genPlanets 產生)建一筆新
// engine.ColonyState,是 ColonizeStar(玩家拓殖,下方)與 session.go aiExpand(AI 擴張)共用的
// 殖民地建法——2026-07-11 前 aiExpand 只標記 Star.Owner=2 旗標、從未建立真正的殖民地模型(見
// AIOpponent.ColonyStars 欄位註解),AI 版圖擴張後經濟不會成長。抽出共用後兩處呼叫端行為一致
// (氣候/重力/礦產/大小解析、PopMax 查表、全農起始、士氣算法皆同一套規則),不會出現「玩家殖民地
// 一套規則、AI 殖民地另一套」的不忠實分裂。
//
// gov 是套用士氣基準的政府型態:玩家傳 s.Government;AI 對手(AIOpponent)沒有 Government 欄位
// ——政府型態未建模,aiExpand 傳 gamedata.MoraleGovDictatorship 當保守預設(與母星
// playerHomeworldColony 的政府基準一致,理由見該函式)。
// foodBonus/indBonus/resBonus 是種族環境加成:玩家傳 Races[s.RaceIndex] 對應值;AI 對手沒有種族
// 加成模型可查(擴張出的新殖民地非母星,無種族資料來源),aiExpand 一律傳 0,不臆造。
//
// 回傳 ok=false 時 reason 說明原因(對映 ColonizeStar 既有的前置條件文字):starIdx 越界(不應
// 發生)、氣候資料無法辨識(不應發生)、或該行星需額外科技才能殖民(氣態巨星/小行星帶,目前星系
// 生成從不產生,見檔頭§1,實務上不會觸發)。呼叫端各自處理:ColonizeStar 直接把 reason 回給
// UI;aiExpand 這類背景擴張沒有 UI 可顯示,只用 ok 決定要不要放棄這顆星、繼續找下一顆。
func (s *GameSession) newColonyFromPlanet(planetIdx int, gov gamedata.MoraleGovernmentType, foodBonus, indBonus, resBonus int) (colony engine.ColonyState, ok bool, reason string) {
	if planetIdx < 0 || planetIdx >= len(s.Planets) {
		return engine.ColonyState{}, false, "無行星資料(不應發生)"
	}
	planet := s.Planets[planetIdx]

	if planet.NoPlanet {
		return engine.ColonyState{}, false, "這顆恆星沒有行星(黑洞)"
	}

	var (
		climate gamedata.PlanetClimate
		gravity gamedata.PlanetGravity
		mineral gamedata.PlanetMinerals
		size    gamedata.PlanetSize
	)
	if planet.Gen >= 1 {
		// 原版骰表生成的行星:直接讀 enum,不從顯示字串反解。
		climate, gravity, mineral, size = planet.ClimateID, planet.GravityID, planet.MineralID, planet.SizeID
	} else {
		// 2026-08-06 之前的存檔只有顯示字串,回填一次(見 Planet.Gen 註解)。
		var cok bool
		climate, cok = climateFromDisplay(planet.Climate)
		if !cok {
			return engine.ColonyState{}, false, "行星氣候資料無法辨識(不應發生,見 climateDisplayToGamedata)"
		}
		var gok, mok, szok bool
		if gravity, gok = gravityFromDisplay(planet.Gravity); !gok {
			gravity = gamedata.NORMAL_G // 不應發生的保守預設,見 gravityDisplayToGamedata
		}
		if mineral, mok = mineralFromDisplay(planet.Mineral); !mok {
			mineral = gamedata.POOR // 不應發生的保守預設,見 mineralDisplayToGamedata
		}
		if size, szok = sizeFromDisplay(planet.Size); !szok {
			size = gamedata.MEDIUM_PLANET // 不應發生的保守預設,見 sizeDisplayToGamedata
		}
	}
	// 行星類別門檻(手冊 p.61:「A Colony Ship can establish a colonial foothold on any
	// uncolonized planet」但 p.55 明講殖民地只能建在 solid planet 上)。氣態巨星/小行星帶
	// 要先蓋前哨站(outpost.go),再由人造行星科技把那個天體改造成可住人的行星
	// (artificialplanet.go,原版建築 48)——兩者都已實作,這裡只負責擋下直接殖民。
	//
	// Gen < 2 的舊存檔沒有 TypeID(零值 0 不是任何合法類別),restorePlanetIDs 已回填
	// HABITABLE;這裡再擋一次零值,免得手改過的存檔繞過去。
	if planet.TypeID != 0 && planet.TypeID != gamedata.HABITABLE {
		return engine.ColonyState{}, false, planetTypeDisplayName(planet.TypeID) + "不能直接殖民,要先建前哨站再改造成人造行星"
	}
	if !climateColonizable(climate) {
		return engine.ColonyState{}, false, "此類行星的氣候無法建立殖民地"
	}

	// 行星特殊物產的效果(手冊定性 + 反組譯定量,見 gamedata/planet_special.go)。
	//
	// 這裡只處理「殖民當下」的兩類:①原住民併入人口 ②持續性產出/收入修正。
	// 太空殘骸/海盜藏寶/失散殖民地/受困英雄/遠古文物那五種是**抵達星系時**觸發的一次性發現
	// (原版 Do_System_Discoveries_At_Star_),在 discovery.go,不在這裡。
	special := planet.SpecialID
	// 原住民:殖民船帶來的 1 個人口單位之外,原版再加 3 個原住民人口單位(全部是農夫)。
	startPop := colonizeStartPopulation + gamedata.SpecialExtraPopulationOnColonize(special)

	foodPerFarmer := gamedata.ClimateFoodPerFarmer(climate) + foodBonus +
		gamedata.SpecialFoodPerFarmerBonus(special) // 原住民:手冊「+2 food production advantage」
	industryPerWorker := gamedata.MineralIndustryPerWorker(mineral) + indBonus
	// 每科學家研究用銀河基準 3(gamedata.ResearchPerScientistNorm)。
	// 先前這裡硬編 30——那是 2026-07-12 母星校正前的舊值,母星改成 3 之後這裡沒跟著改,
	// 造成「拓殖一顆新星,研究產出是母星的十倍」的失衡。手冊無環境相關的研究公式,兩處同一基準。
	researchPerScientist := gamedata.ResearchPerScientistNorm + resBonus
	if n := gamedata.SpecialResearchPerScientist(special); n > 0 {
		// 遠古文物:手冊「produces 5 research points instead of the usual 3」——
		// 是**取代**基準值而不是相加,故種族加成仍疊在上面。
		researchPerScientist = n + resBonus
	}

	popMax := gamedata.PlanetBasePopMax(size, climate)
	if popMax < startPop {
		popMax = startPop // 保底:新殖民地的人口上限不能低於起始人口本身
	}

	colony = engine.ColonyState{
		Population:           startPop,
		PopMax:               popMax,
		Farmers:              startPop, // 全農,見檔頭§2 理由(避免首回合饑荒)
		FoodPerFarmer:        foodPerFarmer,
		IndustryPerWorker:    industryPerWorker,
		ResearchPerScientist: researchPerScientist,
		PlanetSize:           size,
		PlanetGravity:        gravity,
		MineralRichness:      mineral,
		Climate:              climate,
		MoralePercent:        colonyMoralePercent(gov, nil), // 新殖民地無任何建築,見檔頭§2
		// 金礦 +5 / 寶石礦 +10 BC/回合(手冊逐字)。SpecialIncome 是殖民地層的固定收入,
		// 由 engine.RunEmpireTurn 併進帝國總收入。
		SpecialIncome: gamedata.SpecialIncomePerTurn(special),
	}
	return colony, true, ""
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
// 比照 InvadeColony 既有慣例),Star.Owner 轉 1,並從 s.Fleet().Ships 移除用掉的那艘殖民船。
func (s *GameSession) ColonizeStar(starIdx int) ColonizationResult {
	p := s.FirstColonizablePlanet(starIdx)
	if p < 0 && starIdx >= 0 && starIdx < len(s.Stars) {
		// 挑不到可拓殖的行星時,仍走下面的完整檢查,好把**真正的原因**講出來
		// (「艦隊沒到」「有怪獸」「已被 AI 佔領」…),而不是統一回一句「沒有可拓殖的行星」。
		p = s.PlanetAt(starIdx)
	}
	return s.ColonizePlanet(p)
}

// FirstColonizablePlanet 回傳某星系裡第一顆「還沒有殖民地、氣候可直接殖民」的一般行星。
//
// 多天體星系之後這才是拓殖的真正對象——**原版本來就是選行星不是選星**
// (手冊 p.61:「Colonies can only be established on solid planets」,而一個星系有 1..5 個天體)。
func (s *GameSession) FirstColonizablePlanet(star int) int {
	for _, p := range s.PlanetsAt(star) {
		pl := s.Planets[p]
		if pl.NoPlanet || pl.TypeID != gamedata.HABITABLE {
			continue // 氣態巨星/小行星帶要蓋前哨站,不能直接殖民(手冊 p.61)
		}
		if !climateColonizable(pl.ClimateID) || s.ColonyIndexOnPlanet(p) >= 0 {
			continue
		}
		return p
	}
	return -1
}

// ColonyPlanetIndex 回傳玩家第 i 個殖民地座落的行星索引;查不到回 −1。
//
// 舊存檔沒有 PlayerColonyPlanets(或那一格是 −1 padding),退回「該殖民地所在星系的代表行星」
// ——那正是多天體之前的語意,所以舊存檔讀進來的行為不變。
func (s *GameSession) ColonyPlanetIndex(i int) int {
	if i >= 0 && i < len(s.PlayerColonyPlanets) {
		if p := s.PlayerColonyPlanets[i]; p >= 0 && p < len(s.Planets) {
			return p
		}
	}
	if star := s.PlayerColonyStarIndex(i); star >= 0 {
		return s.PlanetAt(star)
	}
	return -1
}

// ColonyPlanet 回傳玩家第 i 個殖民地座落行星的**可寫指標**;查不到回 nil。
func (s *GameSession) ColonyPlanet(i int) *Planet {
	p := s.ColonyPlanetIndex(i)
	if p < 0 || p >= len(s.Planets) {
		return nil
	}
	return &s.Planets[p]
}

// ColonyPlanetData 回傳玩家第 i 個殖民地座落行星的複本(查不到回零值 + false)。
func (s *GameSession) ColonyPlanetData(i int) (Planet, bool) {
	if p := s.ColonyPlanet(i); p != nil {
		return *p, true
	}
	return Planet{}, false
}

// ColonyName 回傳玩家第 i 個殖民地的顯示名(即所在行星名);查不到回空字串。
//
// 同星系可以有多個殖民地,所以名字必須取自**行星**而不是星——否則同系的兩個殖民地同名。
func (s *GameSession) ColonyName(i int) string {
	if p := s.ColonyPlanet(i); p != nil {
		return p.Name
	}
	return ""
}

// ColonyIndexOnPlanet 回傳某顆行星上的玩家殖民地索引(沒有回 −1)。
func (s *GameSession) ColonyIndexOnPlanet(planet int) int {
	for i, p := range s.PlayerColonyPlanets {
		if p == planet && i < len(s.PlayerColonies) {
			return i
		}
	}
	return -1
}

// ColonizePlanet 在某顆**行星**上建立殖民地(拓殖的真正入口)。
//
// ⚠ **同一個星系可以有多個殖民地**——只要還有沒被佔的可殖民行星。先前的閘是
// 「該星已有歸屬就不可拓殖」,那是一星一行星時代的形狀:同一顆星只有一顆行星,
// 有歸屬就等於沒空位。多天體之後那句話會擋掉「在自己的星系裡拓殖第二顆行星」,
// 而那是原版最基本的擴張手段。
//
// 敵方(Owner==2)的星系仍然不能拓殖——那要打下來。
func (s *GameSession) ColonizePlanet(planetIdx int) ColonizationResult {
	if planetIdx < 0 || planetIdx >= len(s.Planets) {
		return ColonizationResult{Reason: "無效的行星索引"}
	}
	starIdx := s.PlanetStar(planetIdx)
	if starIdx < 0 {
		return ColonizationResult{Reason: "無效的星索引"}
	}
	if s.Fleet().AtStar != starIdx || s.Fleet().ETA != 0 {
		return ColonizationResult{Reason: "艦隊尚未抵達該星"}
	}
	star := &s.Stars[starIdx]
	if star.Owner == 2 {
		return ColonizationResult{Reason: "該星已被敵方佔領,不可拓殖"}
	}
	if s.ColonyIndexOnPlanet(planetIdx) >= 0 {
		return ColonizationResult{Reason: "那顆行星上已經有你的殖民地"}
	}
	// 手冊 p.62 逐字:殖民船要「as long as all space monsters and enemy ships have been
	// cleared from that planet's system」。這條 gate 先前寫在檔頭的引文裡卻沒有實作,
	// 因為 remake 根本沒有怪獸(見 monster.go)。
	if reason := s.monsterBlockReason(starIdx); reason != "" {
		return ColonizationResult{Reason: reason}
	}
	shipIdx := s.findColonyShipIndex()
	if shipIdx < 0 {
		return ColonizationResult{Reason: "艦隊未載運殖民船"}
	}

	foodBonus, indBonus, resBonus := 0, 0, 0
	if s.RaceIndex >= 0 && s.RaceIndex < len(Races) {
		r := Races[s.RaceIndex]
		foodBonus, indBonus, resBonus = r.FoodBonus, r.IndBonus, r.ResBonus
	}
	colony, ok, reason := s.newColonyFromPlanet(planetIdx, s.Government, foodBonus, indBonus, resBonus)
	if !ok {
		return ColonizationResult{Reason: reason}
	}

	s.appendPlayerColony(colony, starIdx, planetIdx) // 平行陣列的 padding 慣例見該函式
	idx := len(s.PlayerColonies) - 1

	star.Owner = 1
	s.Fleet().Ships = append(s.Fleet().Ships[:shipIdx], s.Fleet().Ships[shipIdx+1:]...) // 消耗這艘殖民船
	s.consumeSpecialOnColonize(planetIdx)
	// 手冊 p.85:「If a colony is created at an outpost, the building remains and is repurposed
	// as Marine Barracks.」——原本的前哨站不是白蓋的,改建成海軍陸戰隊營留給新殖民地。
	if s.consumeOutpostForColony(starIdx, planetIdx) {
		if s.ColonyBuildings[idx] == nil {
			s.ColonyBuildings[idx] = make(map[string]bool)
		}
		s.ColonyBuildings[idx][OutpostMarineBarracks] = true
		s.applyBuildingEffect(idx, OutpostMarineBarracks)
		s.recalcColonyMorale(idx) // 海軍陸戰隊營會解除獨裁政府的無 Barracks 士氣懲罰
	}

	return ColonizationResult{Ok: true, ColonyIndex: idx, StartPopulation: colony.Population, PopMax: colony.PopMax}
}

// consumeSpecialOnColonize 把「殖民後就消失」的特殊物產從行星上清掉。
// 原版只有原住民這一種(`Make_New_Colony_Or_Outpost_` 在原住民分支寫 `[planet+0x0F] = 0`)——
// 原住民一旦併進殖民地人口就不再是「這顆星上的特殊物產」,再殖民一次不會又冒出 3 個人;
// 金礦/寶石礦/遠古文物則是持續效果,原版不清,這裡也不清。
func (s *GameSession) consumeSpecialOnColonize(planetIdx int) {
	if planetIdx < 0 || planetIdx >= len(s.Planets) {
		return
	}
	if p := &s.Planets[planetIdx]; gamedata.SpecialConsumedOnColonize(p.SpecialID) {
		p.SpecialID = gamedata.NoSpecial
	}
}
