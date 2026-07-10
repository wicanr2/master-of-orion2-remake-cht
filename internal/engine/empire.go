package engine

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// EmpireOutput 是一個帝國(玩家)一回合的結算結果:各殖民地經濟 + 帝國層級聚合 + 研究推進。
type EmpireOutput struct {
	Colonies           []ColonyOutput // 對應輸入 colonies 順序
	TotalFood          int            // 各殖民地食物盈餘總和(可為負,饑荒殖民地拖累總和)
	TotalNetIndustry   int            // 各殖民地淨工業總和
	TotalResearch      int            // 各殖民地研究總和(投入研究進度)
	TaxRevenue         int            // 各殖民地稅收 BC 總和
	FoodSurplusRevenue int            // 各殖民地「餘糧出售」BC 總和(見下方 RunEmpireTurn 說明)
	NetBC              int            // 本回合國庫淨變化(TaxRevenue + FoodSurplusRevenue - Maintenance)
	Player             PlayerState    // 研究推進 + BC 結算後的玩家狀態
	ResearchDone       bool           // 本回合是否有研究主題完成
}

// RunEmpireTurn 編排一個帝國的一回合:
//  1. 逐殖民地跑經濟結算(RunColonyTurn)。
//  2. 聚合帝國層級的食物盈餘 / 淨工業 / 研究點。
//  3. 用研究總點數推進研究進度(RunResearchPhase)。
//
// 注意:人口成長(各 ColonyOutput.PopGrowth)在本引擎層只輸出、不回寫 Population——MOO2 的
// 成長以分數累積到門檻才 +1 人口單位,該累積門檻/尺度手冊未給、存檔未能乾淨反推(避免臆造)。
// 「累積→回寫 Population」由上層 shell.GameSession.advancePopulation 以 remake 調校門檻處理
// (見該處 provenance 註記),保持本引擎層公式純淨。國庫 BC 結算已於下方以稅收-維護費處理。
func RunEmpireTurn(ps PlayerState, colonies []ColonyState) EmpireOutput {
	out := EmpireOutput{Colonies: make([]ColonyOutput, len(colonies))}
	for i, cs := range colonies {
		co := RunColonyTurn(cs)
		out.Colonies[i] = co
		out.TotalFood += co.FoodSurplus
		out.TotalNetIndustry += co.NetIndustry
		out.TotalResearch += co.Research
		// 稅收:對各殖民地淨工業依帝國稅率抽稅(gamedata.IncomeTaxRevenue,1:1 換 BC)。
		out.TaxRevenue += gamedata.IncomeTaxRevenue(co.NetIndustry, ps.TaxRate)
		// 餘糧收入(GAME_MANUAL.pdf p.25,見 gamedata/income.go IncomeFoodSurplusRevenue
		// provenance):把「賣不完的食物」換成 BC,每單位 0.5 BC(無條件捨去)。只對正盈餘
		// (co.FoodSurplus>0)計入——手冊只描述「出售剩餘糧食」這個收入來源,饑荒(負盈餘)
		// 本身已經由 Starving/colonyGrowth 停擺懲罰,不應該再疊加一筆負 BC(手冊沒有「食物
		// 赤字倒扣 BC」的敘述,IncomeFoodSurplusRevenue 若傳負數字面上會算出負值,故由呼叫端
		// 夾在正盈餘才呼叫,避免雙重懲罰)。fantasticTrader 固定傳 false:本專案的 ColonyState
		// 目前沒有追蹤「Fantastic Trader」這個種族特質的欄位(無可推導模型),TODO 待種族特質
		// 系統補上後再接。
		if co.FoodSurplus > 0 {
			out.FoodSurplusRevenue += gamedata.IncomeFoodSurplusRevenue(co.FoodSurplus, false)
		}
	}
	out.Player, out.ResearchDone = RunResearchPhase(ps, out.TotalResearch)
	// 國庫結算:稅收 + 餘糧收入 - 維護費。
	//
	// TradeGoodsIncome(貿易財收入,gamedata/income.go)刻意未接:手冊定義的貿易財是「玩家把部分
	// 工人產能明確配置去生產 Trade Goods(而非科技/建造)」換來的 BC,但本專案的 ColonyState
	// 只有 Farmers/Workers/Scientists 三種職務,沒有「配置到貿易財」這個第四種產能分配欄位或
	// 對應 UI——沒有可推導的輸入模型,強行套用等於臆造一個不存在的分配比例。TODO:待補「貿易財
	// 配置」這個職務/產能分配機制後再接 TradeGoodsIncome,不在本輪自行捏造模型。
	out.NetBC = out.TaxRevenue + out.FoodSurplusRevenue - ps.Maintenance
	out.Player.BC += out.NetBC
	return out
}
