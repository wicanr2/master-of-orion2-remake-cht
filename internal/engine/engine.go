// Package engine 是回合引擎:把 internal/gamedata 已驗證的公式編排成一個回合的狀態推進。
//
// 設計:
//   - 與存檔二進位格式(internal/save)解耦——引擎操作乾淨的 int 欄位狀態,save↔engine 的轉接
//     另立(未來 adapter)。這讓回合邏輯可獨立單測、不被 save 的 Unknown 欄位污染。
//   - 每個「回合階段」是一個純函式:讀狀態、算輸出,不做 I/O、不含隨機(RNG 擲骰由上層注入)。
//   - 編排器(RunColonyTurn 等)依 MOO2 回合順序串接各階段。
//
// 目前涵蓋:殖民地經濟(食物/工業/污染/研究/人口成長)。研究進度、國庫、戰鬥解算為後續階段。
package engine

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// ColonyState 是回合引擎操作的乾淨殖民地狀態(人口與產出以「單位」計)。
type ColonyState struct {
	Population int // 該殖民地目前總人口
	// PopMax 人口上限(gamedata.MaxPopulation=42 為硬上限)。生態圈(Biospheres p.99,「星球
	// 人口上限 +2 單位」)直接對這個欄位 += 2(shell.applyBuildingEffect),不另立
	// PopMaxBonus 影子欄位——PopMax 本身就是 colonyGrowth/shell.advancePopulation 直接讀取
	// 的成長上限,沒有其他公式需要區分「原始值」與「加成後的值」,疊加一個影子欄位只會多一個
	// 「兩處都要記得加總」的錯誤來源,不划算。
	PopMax     int
	Farmers    int // 分配為農夫的人口數
	Workers    int // 分配為工人的人口數
	Scientists int // 分配為科學家的人口數

	// 每單位產出率(存檔已依科技/種族/地形算好,引擎直接乘人數)。
	FoodPerFarmer        int
	IndustryPerWorker    int
	ResearchPerScientist int

	PlanetSize gamedata.PlanetSize // 決定污染容忍值

	// 種族/建築旗標(影響污染與成長)。
	TolerantRace       bool // Tolerant 特性/矽晶:不需花產能清污染
	PollutionProcessor bool // 污染處理器
	AtmosphericRenewer bool // 大氣更新器
	CoreWasteDump      bool // 核心廢料場(完全消除污染)
	Housing            bool // 是否處於「住房」產能配置(啟用住房成長獎金 h)
	// TradeGoods 是否處於「貿易品」建造佇列配置(shell.GameSession.syncTradeGoodsFlag 依玩家
	// 建造選單同步)。true 時該殖民地當回合淨工業不蓋建築,改由 RunEmpireTurn 呼叫
	// gamedata.TradeGoodsIncome 以 2:1(一般種族)/1:1(Fantastic Trader)換算成 BC,計入
	// EmpireOutput.TradeGoodsRevenue(GAME_MANUAL.pdf p.70)。「不累積建造進度」由呼叫端
	// (shell.advanceBuilds)依建造項名稱處理,engine 層只負責換算收入。
	TradeGoods bool

	// 成長獎金(百分點)之和:g 一般 + r 種族 + i AI + t 科技 + l + e(住房 h 由引擎計)。
	GrowthBonusSum int

	// MoralePercent 是淨士氣對產出的百分點調整(每格笑臉 +10、哭臉 -10;正負皆可)。
	// 依手冊套用於食物/工業/研究(見 gamedata.MoraleProductionOutput)。
	MoralePercent int

	// --- 殖民地整體「固定加成」欄位(與人數無關,對照 docs/tech/colony-buildings.md 逐項頁碼) ---
	//
	// 這組欄位修正舊版建模誤差:手冊裡明寫「殖民地整體固定 +N」的建築(自動化工廠 p.78、
	// 機器人採礦廠 p.80、深層核心礦場 p.82、研究實驗室 p.94、行星超級電腦 p.95、銀河網路
	// 中心 p.98、水耕農場 p.99、地底農場 p.100),因 engine 舊版沒有固定加成欄位,曾被近似
	// 揉進「每工人/科學家/農夫」的 per-worker 欄位(FoodPerFarmer/IndustryPerWorker/
	// ResearchPerScientist)裡湊數——這會讓小殖民地(人少)吃到過高倍率、大殖民地(人多)吃到
	// 過低倍率,兩頭都偏離原版。加了這組獨立欄位後,per-worker 與固定值分開累加,不再互相污染。
	FlatFood     int // 殖民地食物整體固定加成(水耕農場 p.99 +2、地底農場 p.100 +4)
	FlatIndustry int // 殖民地工業整體固定加成(自動化工廠 p.78 +5、機器人採礦廠 p.80 +10、深層核心礦場 p.82 +15)
	FlatResearch int // 殖民地研究整體固定加成(研究實驗室 p.94 +5、行星超級電腦 p.95 +10、銀河網路中心 p.98 +15)

	// FlatGrowth 是複製中心(p.99)「人口成長 +0.1 單位/回合,直到達星球人口上限為止」的固定
	// 成長點數。本 remake 的成長累加尺度(shell.popGrowthThreshold=300 代表 1 人口單位)本身
	// 是調校值、非官方 1 人口=100,000 的精確换算(見 session.go popGrowthThreshold 註記),故
	// 0.1 官方人口單位無法精確转成這個尺度的點數——呼叫端(shell.applyBuildingEffect)以
	// 「popGrowthThreshold 的 1/10」設定本欄位,維持與既有成長門檻同一把尺,但仍是近似值,
	// 非官方精確數字。
	FlatGrowth int

	// IncomeBonusPercent 是該殖民地「所有來源 BC 收入」加成百分比,可累加(太空港 p.79 +50、
	// 行星證券交易所 p.93 +100 → 兩者皆建則 +150)。套用點在 RunEmpireTurn(逐殖民地迴圈內,
	// 對「這個殖民地」當回合稅收+餘糧收入+貿易品收入的小計乘上 (100+bonus)/100,再併入帝國
	// 總額)——可精確做到手冊原文「該殖民地」的範圍,不是帝國整體近似。不含維護費(手冊只講
	// 收入加成,未講維護費打折)。
	IncomeBonusPercent int

	// IncomePerPop 是種族「錢」特質對人頭收入的**增量(delta)**,以半 BC 為單位(與 food_per_farmer
	// 同款半單位慣例,因手冊 Money pick 有 0.5 粒度)。一般種族為 0(仍享 gamedata.BaseIncomePerPopHalfBC
	// 的每人基礎 1 BC);諾蘭姆 +2(手冊 p.16「additional 1 BC per turn」→基礎1+額外1=2 BC/人);
	// 自訂種族 Money pick:差 -1(-0.5)、佳 +1(+0.5)、優 +2(+1)。套用點在 RunEmpireTurn 逐殖民地
	// 迴圈:perCapita = (BaseIncomePerPopHalfBC + IncomePerPop),floor 0,income = perCapita*Pop/2,
	// 併入殖民地稅收+餘糧+貿易品小計(故建築 % 一併放大,對應手冊「money 收入受太空港/證交所加成」)。
	IncomePerPop int

	// PlanetGravity 該殖民地所在行星的重力等級(LOW_G/NORMAL_G/HEAVY_G,GAME_MANUAL.pdf p.58)。
	// 驅動 colonyFood/RunColonyTurn 對 per-worker 產出套用的重力懲罰(見
	// gamedata.GravityPenaltyPercent)。
	//
	// 種族自身的 Low-G/High-G 重力天賦(TRAIT_LOW_G 等,見 docs/tech/custom-race-picks.md)
	// 尚未在 ColonyState 建模——RunColonyTurn 呼叫 GravityPenaltyPercent 時固定傳
	// gamedata.NORMAL_G 當「種族重力」基準,懲罰值只反映「行星重力」單一因子,不含種族天賦
	// 平移。這是 remake 建模簡化,非手冊本節文字直接依據(手冊本節只講行星重力對一般種族的
	// 影響,見 planet_yield.go 檔頭大段註解的版本落差說明)。
	//
	// Go 零值陷阱:gamedata.LOW_G 的 ordinal 恰好是 0,與這個欄位「未設定」的零值相同——
	// 任何建構 ColonyState 卻沒有明確設定本欄位的呼叫端,會被視為 Low-G(-25% 懲罰),而非
	// 預期的「無重力資料」。因此所有既有 ColonyState{...} 字面值(engine/shell 測試、
	// cmd/moo2sim)在這次接線時都已明確補上 PlanetGravity(多半是 NORMAL_G),不依賴零值
	// 隱含語意——新增呼叫端請比照辦理,別漏設這個欄位。
	PlanetGravity gamedata.PlanetGravity

	// NormalizeGravity 對應行星重力產生器(p.104,手冊:「將星球重力正常化至 Normal-G,消除
	// Low-G/Heavy-G 負面效果」)。true 時 RunColonyTurn 強制把重力懲罰歸零,即使
	// PlanetGravity 是 LOW_G/HEAVY_G。
	//
	// 2026-07-11 已接線:gamedata.GravityPenaltyPercent/GravityAdjustedProduction 現由
	// colony.go 的 colonyGravityPenaltyPercent 呼叫,套用在 colonyFood/RunColonyTurn 的
	// per-worker 產出上(食物/工業/研究三者皆套,固定加成 FlatFood/FlatIndustry/
	// FlatResearch 不套,理由見 colony.go 註解)。此旗標現在會真正讓 GravityPenaltyPercent
	// 歸零,行星重力產生器不再是無效旗標。
	NormalizeGravity bool

	// MineralRichness 該殖民地所在行星的礦產豐度分級(ULTRA_POOR..ULTRA_RICH,GAME_MANUAL.pdf
	// p.56-57)。ColonyState 建立當下已經把這個分級「烘」進 IndustryPerWorker(見
	// gamedata.MineralIndustryPerWorker)算出每工人的基礎產能——本欄位是額外保留的原始分類,
	// 供 applyBuildingEffect 的機器人工廠(Robotic Factory p.82)查表取得依豐度分級的
	// 固定加成(gamedata.ProdRoboticFactoryBonus),因為那筆固定加成無法從已經算好的
	// per-worker 費率反推回原始豐度分類。
	//
	// Go 零值陷阱:gamedata.ULTRA_POOR 的 ordinal 恰好是 0,與本欄位「未設定」的零值相同——
	// 比照 PlanetGravity 的既有慣例(見該欄位註解),任何建構 ColonyState 卻未明確設定本欄位的
	// 呼叫端,會被靜默當成 Ultra Poor(機器人工廠只 +5,而非實際豐度應有的加成)。因此所有既有
	// ColonyState{...} 字面值(engine/shell 測試、cmd/moo2sim)在這次接線時都已明確補上
	// MineralRichness——新增呼叫端請比照辦理,別漏設這個欄位。
	MineralRichness gamedata.PlanetMinerals

	// Climate 該殖民地所在行星目前的氣候階梯(TOXIC..GAIA,GAME_MANUAL.pdf p.58-59)。地形改造
	// (Terraforming)/蓋亞轉化(Gaia Transformation)兩個一次性「Special」行動(見
	// internal/gamedata/terraform.go,不是常駐建築,docs/tech/colony-buildings.md 已註明其型別
	// 排除在 40 項建築表之外)靠這個欄位判斷「目前在哪一階、下一階是什麼」;套用完成時直接推進
	// 本欄位,並連帶重算 FoodPerFarmer(gamedata.ClimateFoodPerFarmer 前後差值疊加,保留既有建築
	// 加成不被覆蓋)與 PopMax(gamedata.TerraformPopMaxAfterClimateChange 等比例縮放,近似值,
	// 理由見該函式註解)。實際套用邏輯在 shell.GameSession.applyClimateChange。
	//
	// 與 PlanetGravity/MineralRichness 不同:Climate 不會被 RunColonyTurn 每回合讀取——它是被動
	// 儲存的「目前狀態」,只在地形改造/蓋亞轉化套用的那個瞬間被讀寫一次,平常回合結算仍完全依賴
	// FoodPerFarmer 這個已烘進的費率欄位,不會每回合重新查表。
	//
	// Go 零值陷阱:gamedata.TOXIC 的 ordinal 恰好是 0,與本欄位「未設定」的零值相同——比照
	// PlanetGravity/MineralRichness 的既有慣例,任何會被玩家實際操作地形改造/蓋亞轉化的
	// ColonyState 建構點(shell.playerHomeworldColony、engine.ColonyStateFromSave)都已明確補上
	// Climate,不依賴零值隱含語意。既有 engine/shell/cmd 單元測試(不牽涉地形改造機制)維持零值
	// 不受影響——因為本欄位不像 PlanetGravity 那樣被每回合的核心公式讀取,零值對那些測試無副作用。
	Climate gamedata.PlanetClimate
}

// PlayerState 是回合引擎操作的乾淨玩家(帝國)狀態。
type PlayerState struct {
	BC      int // 國庫(Billion Credits)
	TaxRate int // 稅率(百分比)
	// Maintenance 每回合總維護費,BC 結算時扣除。目前呼叫端(shell.GameSession.EndTurn)只
	// 依實際已建成建築(gamedata.BuiltMaintenanceBC)加總計入;艦隊/間諜/軍官維護費本專案尚無
	// 可推導的模型(未追蹤運輸艦數量等),未計入——TODO 待補,見 session.go
	// totalBuildingMaintenance/newHomeworldPlayerState 的同款註記。此欄位本身仍是純粹輸入,
	// 引擎層不關心維護費怎麼算出來。
	Maintenance int
	// CommandPointsSupply / UsedCommandPoints 指揮評等(Command Rating)供需(GAME_MANUAL.pdf
	// p.169)。與 Maintenance 同款輸入模式:引擎層不關心怎麼算出來,純粹接收呼叫端(通常是
	// shell.GameSession.EndTurn,依實際已建成的軌道衛星 gamedata.CommandPointsFromBuildings +
	// 玩家艦艇清單 gamedata.ShipCommandCost 加總)算好的數字,由 RunEmpireTurn 算超支懲罰。
	//
	// CommandPointsSupply:玩家目前所有殖民地的星基/戰鬥站/星辰要塞供給的指揮評等點數總和
	// (三者取代關係不疊加,見 gamedata.CommandPointsFromBuildings)。手冊同段還提到「通訊科技
	// (Tachyon/Subspace/Hyperspace Communications)與具備 Operations 技能的軍官也會增加此
	// 評等」,但通訊科技只有「每軌道衛星 +1(Tachyon)/+3(Hyperspace,取代前者)」的定性數字、
	// Operations 軍官技能手冊完全沒給數字——兩者都不計入本欄位,TODO 待補(不臆造)。
	// 殖民地本身(不含建築)是否提供基礎指揮評等:手冊全文未提及,故亦不計入,同上 TODO。
	//
	// UsedCommandPoints:玩家目前所有艦艇(不含貨運艦隊 Freighter Fleet,手冊 p.168 明文排除)
	// 依艦體等級加總的指揮評等需求(gamedata.ShipCommandCost)。
	//
	// 兩者預設值 0(呼叫端未設值時視為「供給/需求皆零」,即無超支懲罰)——AI 對手目前用
	// FleetStrength 抽象戰力值,無逐艦清單可推導 UsedCommandPoints,暫不計算,是誠實的
	// 「架構未跟上」而非漏算,見 RunEmpireTurn 註解。
	CommandPointsSupply int
	UsedCommandPoints   int
	ResearchTopic       gamedata.ResearchTopic // 目前研究中的主題
	ResearchProgress    int                    // 目前主題已累積的研究點(RP)
	// CompletedTopics 記錄已完成的研究主題(避免重複)。
	CompletedTopics map[gamedata.ResearchTopic]bool
	// ChosenTech 記錄每個已完成主題「實際選定解鎖」的那一項科技(MOO2 每主題數科技抉擇)。
	// ResearchAll 主題會把全部 Choices 記入。多選主題完成時預設記第一項,玩家可經 UI 改選。
	ChosenTech map[gamedata.ResearchTopic]gamedata.Technology
	// PendingChoice 為「剛完成、玩家可改選解鎖科技」的主題;HasPendingChoice 標記其有效
	// (因 ResearchTopic 0 = 起始科技為合法值,不能用零值判斷)。
	PendingChoice    gamedata.ResearchTopic
	HasPendingChoice bool
	// ExplicitChoice 記錄哪些主題是玩家「明確抉擇」過的(非預設)。用於元件解鎖:
	// 明確抉擇過的主題只解鎖所選科技對應元件;未明確抉擇(AI/預設)維持主題層級(不回歸)。
	ExplicitChoice map[gamedata.ResearchTopic]bool

	// GovtBonusMoneyPercent 政府形式對「money」(BC/稅收)收入的加成百分比(MANUAL_150.html
	// govt_bonus democracy_money/federation_money,見 gamedata.IncomeGovtMoneyBonusPercent /
	// IncomeApplyGovernmentMoneyBonus)。與 Maintenance/CommandPointsSupply 同款輸入模式:
	// 呼叫端(shell.GameSession.EndTurn)依 s.Government 算好傳入,引擎層只套用公式,不關心
	// 政府型態本身如何對應到百分比(不需要 import gamedata 的 MoraleGovernmentType 判斷邏輯)。
	// 0 = 無加成——手冊只列出 Democracy(+50%)/Federation(+75%)兩種政府有此加成,其餘政府
	// (含 demo 用的 Dictatorship)呼叫端應傳 0,RunEmpireTurn 對 0 是 no-op。
	GovtBonusMoneyPercent int

	// ActiveFreighters 玩家目前「使用中」的運輸艦(Freighter)數量,供 RunEmpireTurn 算維護費
	// (gamedata.IncomeFreighterMaintenanceCost,GAME_MANUAL.pdf p.169:每艘 0.5 BC/回合,無條件
	// 捨去)。本專案的艦種塑模(gamedata.ShipType:COMBAT_SHIP/COLONY_SHIP/TRANSPORT_SHIP/
	// OUTPOST_SHIP,見 enums.go)沒有獨立的「Freighter」艦種概念——TRANSPORT_SHIP 是地面入侵
	// 用的運兵船,不是手冊這裡講的貨運艦隊(Freighter Fleet,一種抽象的貿易/後勤艦隊,不佔
	// Command Rating,見 shipspace.go 註解)。
	//
	// 2026-07-11(#4)追加接線:本欄位不再是零呼叫端的死碼——`internal/shell` 新增「運輸艦隊」
	// (`gamedata.FreighterFleetActionName`)殖民地建造選項(Special 一次性行動,見
	// `gamedata/special_actions.go`),每完工一次由 `shell.GameSession.applySpecialAction` 對本
	// 欄位 `+= gamedata.FreighterFleetShipsPerBuild`(手冊:每次建造 +5 艘)。engine 層本身不變,
	// 仍只是「吃呼叫端算好的數字」,呼叫端從恆傳 0 變成會隨玩家建造累加——這就是本欄位原本設計
	// 「接線先備妥、待補艦種追蹤即生效」的那個生效時刻。AI 對手(`AIOpponent`)未接同一個建造
	// 佇列流程,本欄位對 AI 仍恆為 0,見該呼叫端(`shell.EndTurn`)AI 迴圈註解。
	ActiveFreighters int

	// HyperAdvancedResearchCost 是版本規則 profile 對 Hyper-Advanced Lv1 研究(8 個共用同一
	// 成本的 TOPIC_HYPER_* 主題,見 gamedata.IsHyperAdvancedTopic)的成本覆寫值,與
	// CommandPointsSupply/GovtBonusMoneyPercent 同款輸入模式:引擎層不關心版本 profile 本身
	// (不 import gamedata.RuleProfile 判斷邏輯),只接收呼叫端(shell.GameSession.EndTurn,依
	// gamedata.HyperAdvancedCost(s.RuleProfile) 算好)傳入的數字。
	//
	// 0 = 用 gamedata.ResearchChoiceFor(topic).Cost 的套件級預設值(techtree.go 硬編 25000,
	// 即現行 Profile15 行為);非 0 = 覆寫(見 internal/gamedata/ruleprofile.go RuleProfile)。
	// 呼叫端未設值時 Go 零值剛好是「用預設」,無零值陷阱。
	HyperAdvancedResearchCost int
}

// ColonyOutput 是一回合殖民地經濟結算結果。
type ColonyOutput struct {
	Food                 int // 農業總產出
	FoodConsumed         int // 人口消耗(每人口單位 1)
	FoodSurplus          int // Food - FoodConsumed(負值=饑荒,見 Starving)
	Starving             bool
	GrossIndustry        int // 工人總工業產出(未扣污染清理)
	PollutingProduction  int // 仍會產生污染的產能
	PollutionCleanupCost int // 清理污染消耗的產能
	NetIndustry          int // GrossIndustry - PollutionCleanupCost
	Research             int // 科學家總研究產出
	PopGrowth            int // 本回合人口成長(gamedata.ColonyGrowth 結果;饑荒時見備註)
}
