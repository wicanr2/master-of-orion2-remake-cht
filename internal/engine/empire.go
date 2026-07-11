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
	TradeGoodsRevenue  int            // 各「貿易品」殖民地淨工業換算 BC 總和(見下方 RunEmpireTurn 說明)
	NetBC              int            // 本回合國庫淨變化(TaxRevenue + FoodSurplusRevenue + TradeGoodsRevenue - Maintenance - CommandOverflowCost)
	// CommandOverflowCost 指揮評等(Command Rating)供給不足艦艇需求時,每回合從收入額外扣除
	// 的維護費(GAME_MANUAL.pdf p.169,gamedata.IncomeCommandOverflowCost)。已計入 NetBC,
	// 這裡單獨曝露供測試/UI 顯示「這筆錢花在哪」,供≥需時為 0。
	CommandOverflowCost int
	// FreighterMaintenanceCost 使用中運輸艦(Freighter)每回合維護費總和(GAME_MANUAL.pdf
	// p.169,gamedata.IncomeFreighterMaintenanceCost,每艘 0.5 BC)。已計入 NetBC,單獨曝露
	// 供測試/UI 顯示。ps.ActiveFreighters 玩家側已可透過建造「運輸艦隊」變非 0(見該欄位註解
	// 2026-07-11(#4)追加接線段落)——本欄位隨之反映真實維護費;AI 對手未接該建造流程,
	// ActiveFreighters 對 AI 仍恆為 0,本欄位對 AI 側仍是 no-op。
	FreighterMaintenanceCost int
	Player                   PlayerState // 研究推進 + BC 結算後的玩家狀態
	ResearchDone             bool        // 本回合是否有研究主題完成
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
		//
		// 2026-07-11 決定不接 gamedata.IncomeMoraleAdjustedProduction 到這裡(或任何收入項目),
		// 誠實記錄判定依據,避免之後有人「順手」補上造成雙重計算:
		// 手冊(GAME_MANUAL.pdf p.170)講「士氣調整 food/industry/research/income 四項總產出」,
		// 但 income(稅收/餘糧收入/貿易品收入)在本 remake 全部是從已經套過士氣的產出「再換算」
		// 出來的——co.NetIndustry 由 colony.go RunColonyTurn 用
		// `pct := cs.MoralePercent + colonyGravityPenaltyPercent(cs)` 套過 GravityAdjustedProduction
		// 才算出(工業/研究皆同一 pct);co.FoodSurplus 同樣經 colonyFood 用同一 pct 算出。
		// 上面這行 tax 直接讀 co.NetIndustry、下面的 foodRev 讀 co.FoodSurplus、tradeRev 也讀
		// co.NetIndustry——三者都已經隱含士氣調整過一次。若在這裡對 tax/foodRev/tradeRev(或加總
		// 後的 NetBC)再套一次 IncomeMoraleAdjustedProduction,士氣就對同一筆錢生效兩次
		// (一次在「產出」、一次在「產出換算成的收入」),與手冊「每格士氣=10% 總產出變化」
		// 的單一調整量不符。故本檔刻意不呼叫 IncomeMoraleAdjustedProduction;該函式與其單元測試
		// (income_test.go TestIncomeMoraleAdjustedProduction)保留,是驗證公式本身正確、供未來
		// 若改為「income 獨立於已調整產出計算」的架構時使用,不是死碼。demo 母星 morale=0 時
		// 這個決定本來就是 no-op,不影響探針驗證的 BC 軌跡。
		tax := gamedata.IncomeTaxRevenue(co.NetIndustry, ps.TaxRate)
		// 餘糧收入(GAME_MANUAL.pdf p.25,見 gamedata/income.go IncomeFoodSurplusRevenue
		// provenance):把「賣不完的食物」換成 BC,每單位 0.5 BC(無條件捨去)。只對正盈餘
		// (co.FoodSurplus>0)計入——手冊只描述「出售剩餘糧食」這個收入來源,饑荒(負盈餘)
		// 本身已經由 Starving/colonyGrowth 停擺懲罰,不應該再疊加一筆負 BC(手冊沒有「食物
		// 赤字倒扣 BC」的敘述,IncomeFoodSurplusRevenue 若傳負數字面上會算出負值,故由呼叫端
		// 夾在正盈餘才呼叫,避免雙重懲罰)。fantasticTrader 固定傳 false:本專案的 ColonyState
		// 目前沒有追蹤「Fantastic Trader」這個種族特質的欄位(無可推導模型),TODO 待種族特質
		// 系統補上後再接。
		foodRev := 0
		if co.FoodSurplus > 0 {
			foodRev = gamedata.IncomeFoodSurplusRevenue(co.FoodSurplus, false)
		}
		// 貿易品(Trade Goods)收入:貿易品是「建造佇列選項」(與 Housing 同類,見
		// engine.ColonyState.Housing 的先例),不是獨立的產能分配職務——手冊(GAME_MANUAL.pdf
		// p.70)描述的是「殖民地把建造改設為貿易品」,該殖民地當回合的淨工業整包不蓋建築、改以
		// 2:1(一般種族)換算成 BC。cs.TradeGoods 由 shell.GameSession.syncTradeGoodsFlag 依玩家
		// 建造選單同步(見該函式);「不累積建造進度」則由 shell.advanceBuilds 依建造項名稱處理,
		// 兩處合力達成手冊行為,engine 層只負責換算收入。fantasticTrader 固定傳 false:同上
		// FoodSurplusRevenue 呼叫的理由,ColonyState 目前無「Fantastic Trader」種族特質欄位,
		// TODO 待種族特質系統補上後再接。
		tradeRev := 0
		if cs.TradeGoods {
			tradeRev = gamedata.TradeGoodsIncome(co.NetIndustry, false)
		}
		// IncomeBonusPercent(太空港 p.79 +50、行星證券交易所 p.93 +100,可疊加):手冊原文是
		// 「該殖民地所有來源 BC 收入 +N%」——這裡在「逐殖民地」這層迴圈內,對這個殖民地當回合
		// 的稅收+餘糧收入+貿易品收入小計套用加成,再併入帝國總額,精確對應手冊「該殖民地」的
		// 範圍(不是先加總帝國全部收入再打折的近似做法)。不含維護費(手冊只講收入加成,沒講
		// 維護費打折)。
		if cs.IncomeBonusPercent != 0 {
			subtotal := tax + foodRev + tradeRev
			bonus := subtotal * cs.IncomeBonusPercent / 100
			tax += bonus // 加成金額計入稅收分項,不拆分到三個子項(避免無意義的捨入分配)
		}
		out.TaxRevenue += tax
		out.FoodSurplusRevenue += foodRev
		out.TradeGoodsRevenue += tradeRev
	}
	// 政府「money」加成(GAME_MANUAL.pdf 引用 MANUAL_150.html govt_bonus democracy_money/
	// federation_money,gamedata.IncomeApplyGovernmentMoneyBonus)。與上面 cs.IncomeBonusPercent
	// (太空港/證券交易所)不同,政府是帝國層級屬性、不是逐殖民地建築,故在迴圈外對「本回合已
	// 加總的帝國 money 收入」(稅收+餘糧收入+貿易品收入,此時已含各殖民地 IncomeBonusPercent)
	// 套一次,而非逐殖民地套用。加成後差額計入 TaxRevenue(與上方同款作法:不拆分到三個子項,
	// 避免無意義的捨入分配)。ps.GovtBonusMoneyPercent=0(手冊未列出加成的政府,含 demo 用的
	// Dictatorship)時 no-op。
	if ps.GovtBonusMoneyPercent != 0 {
		subtotal := out.TaxRevenue + out.FoodSurplusRevenue + out.TradeGoodsRevenue
		bonused := gamedata.IncomeApplyGovernmentMoneyBonus(subtotal, ps.GovtBonusMoneyPercent)
		out.TaxRevenue += bonused - subtotal
	}
	out.Player, out.ResearchDone = RunResearchPhase(ps, out.TotalResearch)
	// 指揮評等(Command Rating)超支懲罰(GAME_MANUAL.pdf p.169:「For each rating point
	// required by a ship that is not covered, 10 BCs come out of your income every turn.」)。
	// uncovered 為負(供給 > 需求)時夾在 0,IncomeCommandOverflowCost 內部也會再夾一次
	// (雙重保險,不影響結果)。
	uncoveredCommandPoints := ps.UsedCommandPoints - ps.CommandPointsSupply
	if uncoveredCommandPoints < 0 {
		uncoveredCommandPoints = 0
	}
	out.CommandOverflowCost = gamedata.IncomeCommandOverflowCost(uncoveredCommandPoints)
	// 運輸艦(Freighter)維護費(GAME_MANUAL.pdf p.169,gamedata.IncomeFreighterMaintenanceCost)。
	// 獨立於 ps.Maintenance(只含已建建築維護費,見該欄位註解)。ps.ActiveFreighters 玩家側
	// 建造「運輸艦隊」後會非 0(見該欄位註解 2026-07-11(#4)追加接線段落),AI 對手仍恆 0。
	out.FreighterMaintenanceCost = gamedata.IncomeFreighterMaintenanceCost(ps.ActiveFreighters)
	// 國庫結算:稅收 + 餘糧收入 + 貿易品收入 - 維護費 - 指揮評等超支懲罰 - 運輸艦維護費。
	out.NetBC = out.TaxRevenue + out.FoodSurplusRevenue + out.TradeGoodsRevenue - ps.Maintenance - out.CommandOverflowCost - out.FreighterMaintenanceCost
	out.Player.BC += out.NetBC
	return out
}
