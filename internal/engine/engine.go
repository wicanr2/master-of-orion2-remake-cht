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
	PopMax     int // 人口上限(gamedata.MaxPopulation=42 為硬上限)
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
}

// PlayerState 是回合引擎操作的乾淨玩家(帝國)狀態。
type PlayerState struct {
	BC               int                    // 國庫(Billion Credits)
	TaxRate          int                    // 稅率(百分比)
	// Maintenance 每回合總維護費,BC 結算時扣除。目前呼叫端(shell.GameSession.EndTurn)只
	// 依實際已建成建築(gamedata.BuiltMaintenanceBC)加總計入;艦隊/間諜/軍官維護費本專案尚無
	// 可推導的模型(未追蹤運輸艦數量等),未計入——TODO 待補,見 session.go
	// totalBuildingMaintenance/newHomeworldPlayerState 的同款註記。此欄位本身仍是純粹輸入,
	// 引擎層不關心維護費怎麼算出來。
	Maintenance int
	ResearchTopic    gamedata.ResearchTopic // 目前研究中的主題
	ResearchProgress int                    // 目前主題已累積的研究點(RP)
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
