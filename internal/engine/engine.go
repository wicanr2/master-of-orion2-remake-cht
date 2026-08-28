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

// PopulationGroup 是同一個原版 player slot 的殖民者聚合。RaceSlot 是當局帝國槽，
// 不是十三族的 OrigIdx；8／9 保留給 Android／Natives。
type PopulationGroup struct {
	RaceSlot           int
	RaceSlotKnown      bool
	Farmers            int
	Workers            int
	Scientists         int
	PrisonerFarmers    int
	PrisonerWorkers    int
	PrisonerScientists int
	FoodBonus          int
	IndustryBonus      int
	ResearchBonus      int
	Gravity            gamedata.PlanetGravity
	GravityImmune      bool
	Aquatic            bool
	Cybernetic         bool
	Lithovore          bool
	Tolerant           bool
	Subterranean       bool
	GrowthBonusPercent int
	GrowthPoints       int
	ProfileKnown       bool
}

// ColonyState 是回合引擎操作的乾淨殖民地狀態(人口與產出以「單位」計)。
type ColonyState struct {
	Population int // 該殖民地目前總人口
	// BombardmentLastPopulationPoints 對應 .GAM RacePopulation 的最後一名殖民者
	// 百分之一人口點數；BombardmentBuildProgress 對應 raw +0x125。兩者只供原版
	// sub_DCEBD 戰略轟炸傷亡寫回，零值由 resolver 安全解讀為 100／無進度。
	BombardmentLastPopulationPoints int
	BombardmentBuildProgress        int
	// PopMax 人口上限(gamedata.MaxPopulation=42 為硬上限)。生態圈(Biospheres p.99,「星球
	// 人口上限 +2 單位」)直接對這個欄位 += 2(shell.applyBuildingEffect),不另立
	// PopMaxBonus 影子欄位——PopMax 本身就是 colonyGrowth/shell.advancePopulation 直接讀取
	// 的成長上限,沒有其他公式需要區分「原始值」與「加成後的值」,疊加一個影子欄位只會多一個
	// 「兩處都要記得加總」的錯誤來源,不划算。
	PopMax     int
	Farmers    int // 分配為農夫的人口數
	Workers    int // 分配為工人的人口數
	Scientists int // 分配為科學家的人口數
	// PopulationGroups 完整時是逐 player slot 的人口真相；三職務總和必須等於 Population。
	// 空或不完整只供舊 JSON 回退上方三個總數，不得視為混合種族 parity。
	PopulationGroups []PopulationGroup `json:"populationGroups,omitempty"`

	// 每單位產出率(存檔已依科技/種族/地形算好,引擎直接乘人數)。
	FoodPerFarmer        int
	IndustryPerWorker    int
	ResearchPerScientist int
	// Owner*Bonus 是已烘進上方三個快取率的所有者種族數值 trait。逐群產出先扣 owner，
	// 再加入群組自己的 trait；Known 防止舊 JSON 零值被誤認為完整 profile。
	OwnerFoodBonus        int
	OwnerIndustryBonus    int
	OwnerResearchBonus    int
	OwnerRaceProfileKnown bool
	OwnerRaceSlot         int
	OwnerRaceSlotKnown    bool

	PlanetSize gamedata.PlanetSize // 決定污染容忍值

	// 種族/建築旗標(影響食物與污染)。
	// Lithovore 是食岩特性:每人口不消耗食物,因此殖民地不會因食物赤字饑荒。
	// Cybernetic 是半機械特性:每人口消耗半食物、另消耗半生產力；精確帳本在
	// ColonyOutput 的 *Half 欄位,整數欄位則保留給既有 UI/存檔相容層。
	Lithovore          bool
	Cybernetic         bool
	TolerantRace       bool // Tolerant 特性/矽晶:不需花產能清污染
	Aquatic            bool // 水生:建立／改造殖民地時使用水生氣候對映
	Subterranean       bool // 穴居:建立殖民地時人口上限 +2×星球大小
	PollutionProcessor bool // 污染處理器
	AtmosphericRenewer bool // 大氣更新器
	CoreWasteDump      bool // 核心廢料場(完全消除污染)

	// UnassimilatedPop 是這個殖民地還沒被同化的**征服人口**(GAME_MANUAL.pdf p.21-24)。
	// 攻下一個殖民地時等於當地全部人口,之後每隔幾回合減一(回合數依政體,見
	// gamedata/assimilation.go)。0 = 全部是自己的子民(自己拓殖的殖民地一開始就是 0)。
	//
	// AssimilationProgress 是「朝下一個單位累積的原版 0..239 raw 進度」。兩個欄位都由
	// shell.advanceAssimilation 推進;engine 的經濟結算不直接讀它們——多種族人口的
	// 20% 士氣懲罰走 shell.colonyMoralePercent 折成 MoralePercent(第 42 項(關掉兩條留白)接的)。
	UnassimilatedPop     int
	AssimilationProgress int
	// Unassimilated* 保存未同化人口的實際職務分布。原版 packed Colonist 的
	// PRISONER flag 是逐人口、逐職務消費；三欄總和等於 UnassimilatedPop 時，
	// 產出使用精確分布。舊 JSON 沒有這三欄時，經濟層會回退舊比例模型。
	UnassimilatedFarmers    int
	UnassimilatedWorkers    int
	UnassimilatedScientists int

	// ConqueredFrom 是這個殖民地**被打下來之前**的主人(shell.AIPlayers 的索引);
	// 叛亂成功時殖民地要還給它(手冊 p.165「the colony reverts back」)。
	//
	// ⚠ **不能用 0 當「沒有」的哨兵**——0 是合法的 AI 索引,而舊存檔解出來就是 0。
	// 所以另立 ConqueredFromKnown:false 時 ConqueredFrom 一律無意義。
	ConqueredFrom      int
	ConqueredFromKnown bool

	// FoodReplicators 是食物複製機(GAME_MANUAL.pdf p.85)。true 時,食物赤字會用產能
	// 以 2:1 換成食物補足(每單位再花 1 BC,由 RunEmpireTurn 結算)。
	// **只補缺口、不換出盈餘**——見 gamedata/food_replicators.go 的「as needed」段落。
	FoodReplicators bool

	// Recyclotron 是再生反應爐(GAME_MANUAL.pdf p.81)。
	//
	// 手冊原文兩句缺一不可:「each unit of population generates 1 industrial production,
	// **regardless of its assigned job**」與「This increased production does **not count
	// toward the planetary pollution level**, since all the materials used are recycled.」
	//
	// 所以它**不能**接成 FlatIndustry:那個欄位是在污染縮減之**前**併進 gross 的,
	// 接錯地方會讓這份產能跟著產生污染,正好與手冊那句相反。接法見 colonyOutput。
	Recyclotron bool

	// --- 「成就」科技的全帝國效果(2026-08-08 第 59 項(成就科技效果),見 shell/achievements.go)---
	//
	// 這兩個由 shell 每回合依玩家科技重算(`syncAchievementColonyFields`),不是建築旗標
	// ——成就是研究出來的,而研究結果會變(被偷、被交換),所以不能像建築那樣「完工時設一次」。

	// NanoDisassemblers 是奈米分解者成就:行星的污染容忍值加倍。
	NanoDisassemblers bool
	// IndustryPerWorkerBonus 是微晶構築成就給**每個工業工人**的額外產能。
	// 與 FlatIndustry 不同:那個是殖民地整體的固定值,這個要乘工人數。
	IndustryPerWorkerBonus int
	Housing                bool // 是否處於「住房」產能配置(啟用住房成長獎金 h)
	// TradeGoods 是否處於「貿易品」建造佇列配置(shell.GameSession.syncTradeGoodsFlag 依玩家
	// 建造選單同步)。true 時該殖民地當回合淨工業不蓋建築,改由 RunEmpireTurn 呼叫
	// gamedata.TradeGoodsIncome 以 2:1(一般種族)/1:1(Fantastic Trader)換算成 BC,計入
	// EmpireOutput.TradeGoodsRevenue(GAME_MANUAL.pdf p.70)。「不累積建造進度」由呼叫端
	// (shell.advanceBuilds)依建造項名稱處理,engine 層只負責換算收入。
	TradeGoods bool
	// ResearchDiverted 表示本回合研究產出被事件轉作他用（目前為超新星搶救）。
	// RunColonyTurn 仍輸出完整 Research，只有帝國研究聚合略過；這是 shell 每回合由
	// PersistentEvents 重建的暫態輸入，不寫入 JSON／多人快照。
	ResearchDiverted bool `json:"-"`

	// 成長獎金(百分點)之和:g 一般 + r 種族 + i AI + t 科技 + l + e(住房 h 由引擎計)。
	GrowthBonusSum int
	// AIDifficulty*Quarters 是官方 Generic AI bonuses 表的暫態每職務固定加值，單位 1/4。
	// 只由 shell 的 AI 回合副本填入；玩家與持久殖民地維持零值。原版逐殖民地捨入順序尚未
	// 由 IDA 閉合，因此此層採有文件的向下取整近似。
	AIDifficultyFoodQuarters     int `json:"-"`
	AIDifficultyIndustryQuarters int `json:"-"`
	AIDifficultyResearchQuarters int `json:"-"`

	// MoralePercent 是淨士氣對產出的百分點調整(每格笑臉 +10、哭臉 -10;正負皆可)。
	// 依手冊套用於食物/工業/研究(見 gamedata.MoraleProductionOutput)。
	MoralePercent int
	// Government*BonusPercent 對應原版 sub_DE280 的逐職務政體項。它們每回合由 shell
	// 依目前生效政體重建，不能烘進每人口基礎率，否則改政體／取得進階政體會累乘或殘留。
	GovernmentFoodBonusPercent     int
	GovernmentIndustryBonusPercent int
	GovernmentResearchBonusPercent int

	// --- 殖民地整體「固定加成」欄位(與人數無關,對照 docs/tech/colony-buildings.md 逐項頁碼) ---
	//
	// 這組欄位修正舊版建模誤差:手冊裡明寫「殖民地整體固定 +N」的建築(自動化工廠 p.78、
	// 機器人採礦廠 p.80、深層核心礦場 p.82、研究實驗室 p.94、行星超級電腦 p.95、銀河網路
	// 中心 p.98、水耕農場 p.99、地底農場 p.100),因 engine 舊版沒有固定加成欄位,曾被近似
	// 揉進「每工人/科學家/農夫」的 per-worker 欄位(FoodPerFarmer/IndustryPerWorker/
	// ResearchPerScientist)裡湊數——這會讓小殖民地(人少)吃到過高倍率、大殖民地(人多)吃到
	// 過低倍率,兩頭都偏離原版。加了這組獨立欄位後,per-worker 與固定值分開累加,不再互相污染。
	// --- 分項百分比加成(領袖 admin 技能,GAME_MANUAL.pdf p.137)---
	//
	// 與 MoralePercent 同一把尺(百分點),但**只影響一項**:士氣是三項一起動,
	// 這三個各管各的。合併進同一個 pct 再套一次公式,理由同 colonyFood 註解
	// (避免兩次連續整數除法的複合誤差)。
	//
	// 來源:農業官 / 勞工官 / 科學官各 +10%(gamedata.baseSkillValues[2],單位由
	// openorion2 的 skillFormatStrings[2] 確認是百分比,見 gamedata/leader_skill_apply.go)。
	// 固定加成欄位(FlatFood/FlatIndustry/FlatResearch)不吃這個百分比,與士氣的處理一致。
	FoodBonusPercent     int
	IndustryBonusPercent int
	ResearchBonusPercent int

	// PollutionReductionPercent 是「會產生污染的產能」的減幅百分比(**正值 = 減少**)。
	//
	// 來源:環保官(gamedata.baseSkillValues[2][0] = −10%,格式 "%+d%%")。技能的加成值本身
	// 是負的,存進來時已翻成正的減幅——讓消費端讀起來是「(100 − 減幅)」,公式裡不必處理負號。
	//
	// ⚠ 這**不是**減產能。它降的是 colonyPollution 算出來的「致污染產能」,
	// 殖民地實際工業產出不變,只是清理污染少扣一點。見 gamedata.PollutionReducedByPercent。
	PollutionReductionPercent int

	FlatFood     int // 殖民地食物整體固定加成(水耕農場 p.99 +2、地底農場 p.100 +4)
	FlatIndustry int // 殖民地工業整體固定加成(自動化工廠 p.78 +5、機器人採礦廠 p.80 +10、深層核心礦場 p.82 +15)
	FlatResearch int // 殖民地研究整體固定加成(研究實驗室 p.94 +5、行星超級電腦 p.95 +10、銀河網路中心 p.98 +15)

	// FlatGrowth 是複製中心(p.99)「人口成長 +0.1 單位/回合,直到達星球人口上限為止」的固定
	// 成長點數。官方 patch 1.50 手冊的標準值 +100k 對應 100 點；完整人口單位門檻為 1,000 點。
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

	// SpecialIncome 是行星特殊物產給這個殖民地的每回合固定 BC 收入
	// (手冊:寶石礦 +10 BC/回合、金礦 +5 BC/回合;見 gamedata.SpecialIncomePerTurn)。
	// 與 IncomePerPop 不同,這筆與人口無關——三人小殖民地與四十人大都會拿一樣多。
	// 套用點在 RunEmpireTurn 逐殖民地迴圈,併進該殖民地的收入小計,因此一併受
	// IncomeBonusPercent(太空港/證交所)放大——手冊原文是「該殖民地**所有來源** BC 收入
	// +N%」,寶石礦收入是該殖民地的一個 BC 來源,不排除。
	SpecialIncome int

	// PlanetGravity 該殖民地所在行星的重力等級(LOW_G/NORMAL_G/HEAVY_G,GAME_MANUAL.pdf p.58)。
	// 驅動 colonyFood/RunColonyTurn 對 per-worker 產出套用的重力懲罰(見
	// gamedata.GravityPenaltyPercent)。
	//
	// 種族自身的 Low-G/High-G 由下方 RaceGravity/RaceGravityKnown 建模；未知舊狀態回退
	// Normal-G。完整 typed population groups 會逐 colonist slot 改讀各群 Gravity；此欄位
	// 只保留 owner 快取與舊 JSON fallback。
	//
	// Go 零值陷阱:gamedata.LOW_G 的 ordinal 恰好是 0,與這個欄位「未設定」的零值相同——
	// 任何建構 ColonyState 卻沒有明確設定本欄位的呼叫端,會被視為 Low-G(-25% 懲罰),而非
	// 預期的「無重力資料」。因此所有既有 ColonyState{...} 字面值(engine/shell 測試、
	// cmd/moo2sim)在這次接線時都已明確補上 PlanetGravity(多半是 NORMAL_G),不依賴零值
	// 隱含語意——新增呼叫端請比照辦理,別漏設這個欄位。
	PlanetGravity gamedata.PlanetGravity
	// RaceGravity 是目前所有者種族的重力適性。RaceGravityKnown 防止舊 JSON／
	// 測試字面值的 Go 零值 0 被誤認為 Low-G；未知時按一般種族 Normal-G。
	// 完整 typed population groups 會逐 colonist slot 查值；此欄位只供 owner／舊狀態 fallback。
	RaceGravity      gamedata.PlanetGravity
	RaceGravityKnown bool

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
	// FantasticTrader 是「神級商人」種族特性(諾蘭姆)。影響兩筆帝國收入:
	// 貿易品 1:1 換 BC(一般種族 2:1)、餘糧每單位 1 BC(一般種族 0.5)。
	//
	// ⚠ 2026-08-08(第 65 項(種族特性31格))。RunEmpireTurn 先前對這兩處硬傳 `false`,理由寫在
	// empire.go:「ColonyState 目前沒有追蹤『Fantastic Trader』這個種族特質的欄位
	// (無可推導模型),TODO 待種族特質系統補上後再接。」——第 65 項(種族特性31格)把特質表接進來了。
	//
	// 放在 PlayerState 不是 ColonyState:它是**帝國層**特性,每個殖民地都一樣;
	// 放進殖民地會變成 N 份可以不一致的真相。
	FantasticTrader bool
	// FleetResearch 是艦上偵察實驗室每回合產生的研究點數(手冊:依艦體 1/2/4/8/16/32)。
	//
	// 與殖民地研究併入同一個研究階段——手冊沒有把它們分開處理,而分開會多出一個
	// 「艦隊研究要不要吃士氣/政體加成」的問題,那個問題手冊沒有答案。
	// 由 shell.GameSession.FleetResearchPoints 每回合算好傳入(同 Maintenance 的輸入模式)。
	FleetResearch int
	// Maintenance 是建築維護費；間諜與軍官保留獨立分項，RunEmpireTurn 在同一次
	// 國庫結算中扣除，讓 UI／存檔可追查每筆來源。
	Maintenance        int
	SpyMaintenance     int
	OfficerMaintenance int
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
	// 兩者預設值 0（呼叫端未設值時視為供需皆零）。shell 會在玩家與 AI 的經濟結算前
	// 依殖民地建築與持久實艦重算；engine 只消費正規化後的數值。
	CommandPointsSupply int
	UsedCommandPoints   int
	// CommandOverflowCostPerPoint 是本回合每一點指揮赤字成本。<=0 使用玩家預設 10 BC；
	// shell 對 AI 依官方五級難度表暫態覆寫，且不序列化。
	CommandOverflowCostPerPoint int                    `json:"-"`
	ResearchTopic               gamedata.ResearchTopic // 目前研究中的主題
	ResearchProgress            int                    // 目前主題已累積的研究點(RP)
	// ResearchApplication 是原版 player+0x322 的 typed 對應：開始研究 topic 時已選定、
	// 突破後才真正取得的科技應用。HasResearchApplication 區分合法零值與尚未選定。
	ResearchApplication    gamedata.Technology
	HasResearchApplication bool
	// TreatyIncomeBC / TreatyResearch 是 shell 在回合開始前依正式貿易／研究協議
	// 填入的帝國層級外部輸入。零值代表沒有協議，RunEmpireTurn 會把它們納入
	// 同一個 BC／研究結算；shell 在結算後會清回零，不把一次性回合輸入當成持久收入。
	TreatyIncomeBC int
	TreatyResearch int
	// CompletedTopics 記錄已完成的研究主題；Hyper 的重複次數另存 HyperAdvancedLevels。
	CompletedTopics map[gamedata.ResearchTopic]bool
	// HyperAdvancedLevels 是八個終端研究主題各自的重複完成次數。
	HyperAdvancedLevels map[gamedata.ResearchTopic]int
	// ChosenTech 記錄每個已完成主題「實際取得」的科技。尚在研究中的選項只能放在
	// ResearchApplication，不能提早寫入此 map 而讓消費端誤判已解鎖。
	ChosenTech map[gamedata.ResearchTopic]gamedata.Technology
	// GrantedTechs 保存不能表示成「某主題唯一選擇」的額外科技應用，例如原版在取得
	// Battleoids 時連帶授予 Armor Barracks。舊 JSON 缺欄位時 nil 等同空集合。
	GrantedTechs map[gamedata.Technology]bool `json:"grantedTechs,omitempty"`
	// PendingChoice 是目前 topic 尚待玩家選 application 的 UI 狀態。讀取舊存檔時也可能
	// 暫時代表舊版「突破後待改選」，ChooseResearchTech 會依 CompletedTopics 判別相容路徑。
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

	// HasGalacticCurrencyExchange 是「已研究銀河貨幣交易所」(Achievement 科技)。
	// 手冊:「increases the income generated by all colonies (from all sources) by 50%」。
	// 由 shell 依科技擁有狀況填(見 shell 的 syncEmpireTechFlags);零值 false = 沒有,安全。
	HasGalacticCurrencyExchange bool

	// ActiveFreighters 對映原版 player+0x36 的貨運艦總數，供食物／殖民者運輸與
	// RunEmpireTurn 算維護費。實際使用數量在 SurplusFreighters>0 時為總數減餘量，否則
	// 為總數；每艘 0.5 BC/回合並無條件捨去。本專案的艦種塑模
	// (gamedata.ShipType:COMBAT_SHIP/COLONY_SHIP/TRANSPORT_SHIP/
	// OUTPOST_SHIP,見 enums.go)沒有獨立的「Freighter」艦種概念——TRANSPORT_SHIP 是地面入侵
	// 用的運兵船,不是手冊這裡講的貨運艦隊(Freighter Fleet,一種抽象的貿易/後勤艦隊,不佔
	// Command Rating,見 shipspace.go 註解)。
	//
	// 2026-07-11(#4)追加接線:本欄位不再是零呼叫端的死碼——`internal/shell` 新增「運輸艦隊」
	// (`gamedata.FreighterFleetActionName`)殖民地建造選項(Special 一次性行動,見
	// `gamedata/special_actions.go`),每完工一次由 `shell.GameSession.applySpecialAction` 對本
	// 欄位 `+= gamedata.FreighterFleetShipsPerBuild`(手冊:每次建造 +5 艘)。AI 不走一般建造
	// 佇列；精確職務路徑依 sub_D6AD4 的壓力旗標、難度與 Random(10) 直接增加 5 艘。
	ActiveFreighters int
	// SettlersFreighted 對映原版 player+0x40；每個在途 settler 佔用 5 艘貨運艦。
	// remake 尚未提供殖民者運輸 UI，正常新局為 0；GAM 匯入保留原值。
	SettlersFreighted int `json:"settlersFreighted,omitempty"`
	// FoodFreighted／SurplusFreighters 分別對映 player+0x3E／+0x38，於帝國食物
	// 運輸重算時更新。SurplusFreighters 可為負，表示有可供應食物卻受運力限制。
	FoodFreighted     int `json:"foodFreighted,omitempty"`
	SurplusFreighters int `json:"surplusFreighters,omitempty"`
	// FoodReplicatorBCHalfRemainder 是食物複製機半食物付款留下的半 BC 餘數。
	// 0/1 都是合法值；每兩個半 BC 在下一次帝國結算時合併成 1 BC。
	// 這是 remake 的精確帳本欄位，舊 JSON 缺欄位時零值安全。
	FoodReplicatorBCHalfRemainder int `json:"foodReplicatorBCHalfRemainder,omitempty"`
	// AIDifficultyIncomeQuartersPerPop 是 AI 每人口 BC 難度加值，單位 1/4 BC。
	// shell 僅在本回合 AI PlayerState 副本設定；零值不影響玩家或舊存檔。
	AIDifficultyIncomeQuartersPerPop int `json:"-"`

	// HyperAdvancedResearchCost 是版本規則 profile 對 Hyper-Advanced 第一級研究(8 個共用同一
	// 基礎成本的 TOPIC_HYPER_* 主題,見 gamedata.IsHyperAdvancedTopic)的覆寫值；研究結算再依
	// Player_Research_Cost_ @ 0xE1E96 加上已完成 level×10000。它與
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
	Food         int // 農業總產出
	FoodConsumed int // 整數相容值；精確人口消耗見 FoodConsumedHalf
	FoodSurplus  int // 半單位餘糧轉成的整數相容值(負值=饑荒,見 Starving)
	// 半單位精確帳本。原版 food_consumption_* / industry_consumption_* 以半單位
	// 儲存；整數欄位不能表達 Cybernetic 的人口奇數情況,所以 UI 相容值與公式值分開。
	FoodHalf             int // Food × 2
	FoodConsumedHalf     int // 實際食物消耗(一般=2×人口,Cybernetic=人口,Lithovore=0)
	FoodSurplusHalf      int // FoodHalf - FoodConsumedHalf
	IndustryConsumedHalf int // Cybernetic 每人口半生產力,一般為 0
	NetIndustryHalf      int // 扣污染、再生反應爐與半生產力消耗後的淨工業 × 2
	Starving             bool
	// FoodReplicated 是食物複製機這回合用產能換出來的食物單位數(p.85)。
	// 已計入 Food / FoodSurplus,且對應的產能已從 NetIndustry 扣掉;單獨曝露是因為
	// 帝國層要用它乘 gamedata.FoodReplicatorBCPerFood 算 BC 成本。
	FoodReplicated int
	// FoodReplicatedHalf 是複製機本回合實際換出的半食物數；完整食物時為偶數。
	FoodReplicatedHalf int
	// FoodReplicatorCostHalfBC 是本殖民地本回合應付的半 BC，帝國層會跨回合累積。
	FoodReplicatorCostHalfBC int
	GrossIndustry            int // 工人總工業產出(未扣污染清理)
	PollutingProduction      int // 仍會產生污染的產能
	PollutionCleanupCost     int // 清理污染消耗的產能
	NetIndustry              int // 半單位淨工業轉成的整數相容值
	Research                 int // 科學家總研究產出
	PopGrowth                int // 本回合人口成長(gamedata.ColonyGrowth 結果;饑荒時見備註)
	// PopulationGroupGrowth 與 ColonyState.PopulationGroups 同索引，供 shell 將成長累積到
	// 正確 player slot。原版共有十個人口槽，固定陣列讓 ColonyOutput 保持可比較；Count=0
	// 表示舊 JSON／群組不完整。
	PopulationGroupGrowth      [gamedata.PopulationRaceSlots]int `json:"-"`
	PopulationGroupGrowthCount int                               `json:"-"`
	Cybernetic                 bool                              // 本回合是否使用半單位生產/食物帳本
}
