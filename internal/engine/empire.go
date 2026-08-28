package engine

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// EmpireOutput 是一個帝國(玩家)一回合的結算結果:各殖民地經濟 + 帝國層級聚合 + 研究推進。
type EmpireOutput struct {
	Colonies             []ColonyOutput // 對應輸入 colonies 順序
	TotalFood            int            // 各殖民地食物盈餘總和(可為負,饑荒殖民地拖累總和)
	TotalNetIndustry     int            // 各殖民地淨工業總和
	TotalFoodHalf        int            // 各殖民地精確半單位食物盈餘總和
	TotalNetIndustryHalf int            // 各殖民地精確半單位淨工業總和
	TotalResearch        int            // 各殖民地、艦隊與研究協議研究總和(投入研究進度)
	TaxRevenue           int            // 各殖民地稅收 BC 總和
	FoodSurplusRevenue   int            // 各殖民地「餘糧出售」BC 總和(見下方 RunEmpireTurn 說明)
	TradeGoodsRevenue    int            // 各「貿易品」殖民地淨工業換算 BC 總和(見下方 RunEmpireTurn 說明)
	NetBC                int            // 本回合國庫淨變化(含 TreatyIncomeBC，再扣維護／超支／運輸艦／複製機成本)
	// TributeCost 是 shell 在雙方帝國經濟結算後依原版 +0x63F 納貢模式
	// 扣除的本回合納貢成本。它不由 engine 自己推導，因為納貢是跨帝國的
	// 關係資料；欄位放在輸出方便回合摘要與測試觀察。
	TributeCost int
	// CommandOverflowCost 指揮評等(Command Rating)供給不足艦艇需求時,每回合從收入額外扣除
	// 的維護費(GAME_MANUAL.pdf p.169,gamedata.IncomeCommandOverflowCost)。已計入 NetBC,
	// 這裡單獨曝露供測試/UI 顯示「這筆錢花在哪」,供≥需時為 0。
	CommandOverflowCost int
	// FreighterMaintenanceCost 使用中運輸艦(Freighter)每回合維護費總和(GAME_MANUAL.pdf
	// p.169,gamedata.IncomeFreighterMaintenanceCost,每艘 0.5 BC)。已計入 NetBC,單獨曝露
	// 供測試／UI 顯示。玩家可建造「運輸艦隊」；AI 精確職務路徑也會在運輸壓力成立且
	// Random(10)<=difficulty 時直接增加 5 艘，因此兩側都可能產生維護費。
	FreighterMaintenanceCost int
	SpyMaintenanceCost       int
	OfficerMaintenanceCost   int
	// FoodReplicatorCost 是各殖民地食物複製機這回合的 BC 成本總和(每單位食物 1 BC,
	// GAME_MANUAL.pdf p.85)。已計入 NetBC,單獨曝露供測試/UI 顯示「這筆錢花在哪」。
	// 沒有殖民地在饑荒(或都沒有這棟)時為 0。
	FoodReplicatorCost int
	// FoodReplicatorCostHalfBC 是本回合所有殖民地複製機的精確半 BC 成本；
	// FoodReplicatorCost 是加上 PlayerState.FoodReplicatorBCHalfRemainder 後
	// 實際從國庫扣除的完整 BC。
	FoodReplicatorCostHalfBC int
	// TreatyIncomeBC / TreatyResearch 是 shell 依玩家與 AI 對手的協議狀態填入的
	// 回合結果，供回合摘要與測試辨識協議收益；零值代表本回合沒有協議收益。
	TreatyIncomeBC int
	TreatyResearch int
	Player         PlayerState // 研究推進 + BC 結算後的玩家狀態
	ResearchDone   bool        // 本回合是否有研究主題完成
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
	return RunEmpireTurnWithResearchRoller(ps, colonies, func(int) int { return 1 })
}

// RunEmpireTurnWithResearchRoller 與 RunEmpireTurn 相同，但把原版研究突破的 1..100 擲骰
// 由上層注入。roller 只會在累積研究嚴格超過成本且本回合有正研究產出時被消費。
func RunEmpireTurnWithResearchRoller(ps PlayerState, colonies []ColonyState, roller func(max int) int) EmpireOutput {
	out := EmpireOutput{Colonies: make([]ColonyOutput, len(colonies))}
	for i, cs := range colonies {
		co := RunColonyTurn(cs)
		out.Colonies[i] = co
		out.TotalFood += co.FoodSurplus
		out.TotalFoodHalf += co.FoodSurplusHalf
		// 食物複製機的 BC 成本(手冊 1 BC per 完整食物)。半機械奇數人口
		// 可能只換出半食物；先以半 BC 聚合，結尾再與跨回合餘數合併。
		out.FoodReplicatorCostHalfBC += co.FoodReplicatorCostHalfBC
		out.TotalNetIndustry += co.NetIndustry
		out.TotalNetIndustryHalf += co.NetIndustryHalf
		if !cs.ResearchDiverted {
			out.TotalResearch += co.Research
		}
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
		tax := gamedata.IncomeTaxRevenueHalf(co.NetIndustryHalf, ps.TaxRate)
		// 餘糧收入(GAME_MANUAL.pdf p.25,見 gamedata/income.go IncomeFoodSurplusRevenue
		// provenance):把「賣不完的食物」換成 BC,每單位 0.5 BC(無條件捨去)。只對正盈餘
		// (co.FoodSurplus>0)計入——手冊只描述「出售剩餘糧食」這個收入來源,饑荒(負盈餘)
		// 本身已經由 Starving/colonyGrowth 停擺懲罰,不應該再疊加一筆負 BC(手冊沒有「食物
		// 赤字倒扣 BC」的敘述,IncomeFoodSurplusRevenue 若傳負數字面上會算出負值,故由呼叫端
		// 夾在正盈餘才呼叫,避免雙重懲罰)。
		//
		// ⚠ 2026-08-08(第 65 項(種族特性31格)):fantasticTrader 先前硬傳 `false`,註解寫著「ColonyState
		// 目前沒有追蹤這個種族特質的欄位(無可推導模型),TODO 待種族特質系統補上後再接」。
		// 特質系統(第 65 項(種族特性31格))補上了,改讀 ps.FantasticTrader(諾蘭姆:每單位 1 BC 而非 0.5)。
		foodRev := 0
		if co.FoodSurplus > 0 {
			foodRev = gamedata.IncomeFoodSurplusRevenueHalf(co.FoodSurplusHalf, ps.FantasticTrader)
		}
		// 貿易品(Trade Goods)收入:貿易品是「建造佇列選項」(與 Housing 同類,見
		// engine.ColonyState.Housing 的先例),不是獨立的產能分配職務——手冊(GAME_MANUAL.pdf
		// p.70)描述的是「殖民地把建造改設為貿易品」,該殖民地當回合的淨工業整包不蓋建築、改以
		// 2:1(一般種族)換算成 BC。cs.TradeGoods 由 shell.GameSession.syncTradeGoodsFlag 依玩家
		// 建造選單同步(見該函式);「不累積建造進度」則由 shell.advanceBuilds 依建造項名稱處理,
		// 兩處合力達成手冊行為,engine 層只負責換算收入。fantasticTrader 同上改讀
		// ps.FantasticTrader(諾蘭姆:1:1 而非 2:1)。
		tradeRev := 0
		if cs.TradeGoods {
			tradeRev = gamedata.TradeGoodsIncomeHalf(co.NetIndustryHalf, ps.FantasticTrader)
		}
		// 已知 remake 差異：原版 E03F1 的 Spaceport、Stock Exchange 與 Financial Leader 都只以
		// B=有機人口基礎BC+Gold/Gems 算獨立加項；現行 IncomeBonusPercent 仍放大稅收、餘糧與
		// 貿易品 subtotal。依 RE-first gate 暫不改行為，待 READY spec 一次修正資料模型與取整。
		// 人頭基礎收入(MOO2 收入模型核心,見 gamedata.BaseIncomePerPopHalfBC):每人口每回合基礎
		// 1 BC(2 半BC),與工業/稅完全分離。種族 Money 特質(cs.IncomePerPop 半BC delta,諾蘭姆
		// +2、自訂 Money picks ±)在此基礎上加減,floor 在 0/人(手冊 p.20「cannot reduce below zero
		// per population unit」)。併入稅收分項,一併受下方 IncomeBonusPercent(太空港/證券交易所)
		// 加成放大,對應手冊「money 收入受建築加成」。
		perCapitaHalf := gamedata.BaseIncomePerPopHalfBC + cs.IncomePerPop
		if perCapitaHalf < 0 {
			perCapitaHalf = 0
		}
		tax += cs.Population * perCapitaHalf / 2
		tax += floorQuarterToWhole(cs.Population * ps.AIDifficultyIncomeQuartersPerPop)
		// 行星特殊物產的固定收入(寶石礦 +10 / 金礦 +5,手冊逐字)。與人口無關,
		// 但同屬「該殖民地的 BC 收入」,故計入小計、受 IncomeBonusPercent 加成。
		if cs.SpecialIncome > 0 {
			tax += cs.SpecialIncome
		}
		if cs.IncomeBonusPercent != 0 {
			subtotal := tax + foodRev + tradeRev
			bonus := subtotal * cs.IncomeBonusPercent / 100
			tax += bonus // 加成金額計入稅收分項,不拆分到三個子項(避免無意義的捨入分配)
		}
		out.TaxRevenue += tax
		out.FoodSurplusRevenue += foodRev
		out.TradeGoodsRevenue += tradeRev
	}
	// 已知 remake 差異：原版政府 BC 加項在 E03F1 逐殖民地以 B 計算；此處仍對帝國總 subtotal
	// 套百分比。依 RE-first gate 只保留可玩現況，不再稱為原版精確順序。
	if ps.GovtBonusMoneyPercent != 0 {
		subtotal := out.TaxRevenue + out.FoodSurplusRevenue + out.TradeGoodsRevenue
		bonused := gamedata.IncomeApplyGovernmentMoneyBonus(subtotal, ps.GovtBonusMoneyPercent)
		out.TaxRevenue += bonused - subtotal
	}
	// 已知 remake 差異：原版銀河貨幣交易所在 E03F1 對每座殖民地只加 floor(B/2)，此處仍對
	// 帝國 subtotal 套乘數；待 RE gate 關閉後由 READY spec 修正。
	if ps.HasGalacticCurrencyExchange {
		subtotal := out.TaxRevenue + out.FoodSurplusRevenue + out.TradeGoodsRevenue
		out.TaxRevenue += gamedata.IncomeApplyGalacticCurrencyExchange(subtotal) - subtotal
	}
	// 偵察實驗室的艦隊研究(第 68 項(元件盤點+飛彈防禦))併進總研究。加在 TotalResearch 上而不是另開一條:
	// 研究階段只有一個投入口,分開會讓「研究完成」的判定要看兩個地方。
	out.TreatyIncomeBC = ps.TreatyIncomeBC
	out.TreatyResearch = ps.TreatyResearch
	out.TotalResearch += ps.FleetResearch + ps.TreatyResearch
	out.Player, out.ResearchDone = RunResearchPhaseWithRoller(ps, out.TotalResearch, roller)
	if transport, ok := OriginalFoodTransport(ps, colonies, nil); ok {
		out.Player.FoodFreighted = transport.FoodFreighted
		out.Player.SurplusFreighters = transport.SurplusFreighters
	}
	// 半 BC 不直接丟失：兩個半 BC 才從國庫扣 1 BC，餘數保存到下一回合。
	pendingReplicatorHalfBC := ps.FoodReplicatorBCHalfRemainder + out.FoodReplicatorCostHalfBC
	out.FoodReplicatorCost = pendingReplicatorHalfBC / 2
	out.Player.FoodReplicatorBCHalfRemainder = pendingReplicatorHalfBC % 2
	// 指揮評等(Command Rating)超支懲罰(GAME_MANUAL.pdf p.169:「For each rating point
	// required by a ship that is not covered, 10 BCs come out of your income every turn.」)。
	// uncovered 為負(供給 > 需求)時夾在 0,IncomeCommandOverflowCost 內部也會再夾一次
	// (雙重保險,不影響結果)。
	uncoveredCommandPoints := ps.UsedCommandPoints - ps.CommandPointsSupply
	if uncoveredCommandPoints < 0 {
		uncoveredCommandPoints = 0
	}
	commandCostPerPoint := ps.CommandOverflowCostPerPoint
	if commandCostPerPoint <= 0 {
		commandCostPerPoint = gamedata.IncomeCommandOverflowCostPerPoint
	}
	out.CommandOverflowCost = uncoveredCommandPoints * commandCostPerPoint
	// 運輸艦(Freighter)維護費(GAME_MANUAL.pdf p.169,gamedata.IncomeFreighterMaintenanceCost)。
	// 獨立於 ps.Maintenance（建築分項）。ActiveFreighters 是總數；原版以本回合重算的
	// SurplusFreighters 推導實際使用數，再按每兩艘 1 BC 收費。
	usedFreighters := ps.ActiveFreighters
	if out.Player.SurplusFreighters > 0 {
		usedFreighters -= out.Player.SurplusFreighters
	}
	if usedFreighters < 0 {
		usedFreighters = 0
	}
	out.FreighterMaintenanceCost = gamedata.IncomeFreighterMaintenanceCost(usedFreighters)
	out.SpyMaintenanceCost = ps.SpyMaintenance
	out.OfficerMaintenanceCost = ps.OfficerMaintenance
	// 國庫結算:稅收 + 餘糧收入 + 貿易品收入 - 維護費 - 指揮評等超支懲罰 - 運輸艦維護費。
	out.NetBC = out.TaxRevenue + out.FoodSurplusRevenue + out.TradeGoodsRevenue + ps.TreatyIncomeBC - ps.Maintenance - out.CommandOverflowCost - out.FreighterMaintenanceCost - out.SpyMaintenanceCost - out.OfficerMaintenanceCost - out.FoodReplicatorCost
	out.Player.BC += out.NetBC
	return out
}
