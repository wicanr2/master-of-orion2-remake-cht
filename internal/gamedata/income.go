package gamedata

// 殖民地/帝國收入(Treasury/Income)唯讀公式,移植自 GAME_MANUAL.pdf(moo2_patch1.5 隨附的完整
// 遊戲手冊)與 MANUAL_150.html(1.50 patch 說明書,latin1 HTML,已去標籤讀取)。
//
// 手冊「Taxes」章節(GAME_MANUAL.pdf p.168)只有敘述性文字與一個 50% 稅率的例子,精確的稅率
// 範圍/級距/轉換比例其實寫在更前面的「Management Buttons/指標說明」章節(GAME_MANUAL.pdf
// p.37,Treasury indicator):「values range from 0% to 50% with increments of 10%. The set
// percentage will be deducted from the industry produced on all your colonies and converted
// into money 1:1.」p.168 補充了與 Trade Goods 轉換比(2:1)的對照,兩處數字一致,互相印證。
//
// MANUAL_150.html 本身沒有稅率/貿易財/殖民地收入的獨立章節,但「Modding with Config →
// Additional Settings」小節列出政府加成的預設數值與換算公式(value * 5 = 加成百分比),其中
// democracy_money / federation_money 兩個政府對「money」(BC/稅收)有明確加成,故一併移植。
//
// 命名前綴:Income = 帝國收入/國庫(泛用),Tax = 稅率,Trade = 貿易財。
// 刻意不用 clamp/round 等通用 helper 名稱,避免與其他檔案(formulas.go 等)撞名。

const (
	// TaxRateMinPercent / TaxRateMaxPercent / TaxRateStepPercent 帝國稅率的可選範圍與級距
	// (GAME_MANUAL.pdf p.37:"values range from 0% to 50% with increments of 10%")。
	TaxRateMinPercent  = 0
	TaxRateMaxPercent  = 50
	TaxRateStepPercent = 10

	// TaxConversionNumerator / TaxConversionDenominator 稅收轉換比例 1:1——稅率扣掉的產能
	// 直接 1:1 換算成 BC(GAME_MANUAL.pdf p.37:"converted into money 1:1"；p.168 再次確認
	// "Taxes also have a better conversion rate (1:1) than Trade Goods (2:1)")。
	TaxConversionNumerator   = 1
	TaxConversionDenominator = 1

	// TradeGoodsConversionNumerator / TradeGoodsConversionDenominator 貿易財轉換比例——每 2 點
	// 產能換 1 BC(GAME_MANUAL.pdf p.70:"Every 2 industry converts to 1 BC")。
	TradeGoodsConversionNumerator   = 1
	TradeGoodsConversionDenominator = 2

	// TradeGoodsFantasticTraderConversionNumerator / ...Denominator Fantastic Trader 種族特質的
	// 貿易財轉換比例——每 1 點產能換 1 BC(GAME_MANUAL.pdf p.70:"unless you are a Fantastic
	// Trader in which case every 1 industry converts to 1 BC")。
	//
	// 注意:手冊 p.25(種族特質「Fantastic Traders」說明)另外寫「on top of that, traders get a
	// 50% bonus to all income derived from producing trade goods」,若照字面計算是 2 倍(2:1 的兩
	// 倍轉換率,而非 1.5 倍),與 p.70 這句"every 1 industry converts to 1 BC"（同樣是 2 倍)數字
	// 上一致但敘述方式矛盾(一個說「1:1」一個說「+50%」)。本檔採用 p.70 這個與一般種族描述並列、
	// 直接給轉換比的句子(TODO 手冊此處敘述前後不一致,+50% 說法未在此另外實作,待查證原版程式行為)。
	TradeGoodsFantasticTraderConversionNumerator   = 1
	TradeGoodsFantasticTraderConversionDenominator = 1

	// IncomeFoodSurplusPerUnitHalfBC 一般種族:每單位剩餘(可外銷)糧食換 0.5 BC(GAME_MANUAL.pdf
	// p.25,描述 Fantastic Traders 特質時提及對照組:"1 BC (instead of the usual half) for every
	// surplus unit of food generated"，故「usual」= 0.5 BC/單位)。
	// 因為 0.5 是半 BC,程式用「分子/分母」表示,実際換算見 IncomeFoodSurplusRevenue。
	IncomeFoodSurplusNumerator   = 1
	IncomeFoodSurplusDenominator = 2

	// IncomeFoodSurplusFantasticTraderPerUnitBC Fantastic Trader 種族:每單位剩餘糧食換 1 BC
	// (GAME_MANUAL.pdf p.25:"receive... plus 1 BC (instead of the usual half) for every surplus
	// unit of food generated")。
	IncomeFoodSurplusFantasticTraderPerUnitBC = 1

	// IncomeCommandOverflowCostPerPoint 指揮評等(Command Rating)不足時,每一點未被覆蓋的評等
	// 需求,每回合從收入扣 10 BC(GAME_MANUAL.pdf p.169:"For each rating point required by a ship
	// that is not covered, 10 BCs come out of your income every turn")。
	IncomeCommandOverflowCostPerPoint = 10

	// CommandPointsBase 帝國「基礎」指揮評等供給——手冊 p.169 只寫軌道衛星(星基/戰鬥站/
	// 星辰要塞)的加成數字,沒有明講不建任何軌道衛星時的評等下限是不是 0,先前 remake 誤植為 0
	// (只算建築供給),導致開局殖民船+2 偵察艦=3 點需求、供給只有母星星基 1 點,缺口 2 點
	// 每回合 -20 BC 死亡螺旋(2026-07-11 回合探針實測:BC 第 7 回合轉負、第 21 回合 -255,
	// 人口第 20 回合起餓死)。
	//
	// 這個常數是用真實存檔反推(oracle,見 rulebook 62/64「靜態溯源+已知輸出反推」)校正回來的:
	// /home/anr2/moo2-private-build/gamedata/mastori2/SAVE10.GAM 有 5 個活躍玩家(不同種族)
	// 各持 1 個殖民地,CommandPoints 欄位分別讀到 6(其中 1 名玩家=8)、UsedCommandPoints=3。
	// 逐一比對該玩家殖民地已建成的軌道衛星:
	//   - CommandPoints=6 的玩家:殖民地只建了 BUILDING_STAR_BASE(存檔 Buildings 索引 40,
	//     星基,+1)。6 - 1 = 5。
	//   - CommandPoints=8 的玩家:殖民地建了星辰要塞(索引 41,+3)。8 - 3 = 5。
	// 5 個不同種族玩家(含以上兩名)反推出的基礎值一致都是 5,與種族/政府無關,故訂為通用常數。
	//
	// 已知限制(TODO,誠實標記不假裝確定):SAVE10.GAM 裡每個玩家都只有 1 個殖民地,無法從單一
	// 存檔分辨這 5 點是「每帝國(per-empire)一次性」還是「每殖民地(per-colony)各自 +5」——
	// 兩種假設在單殖民地情境下算出來的總供給完全相同,無法用這份存檔區分。本專案暫採
	// per-empire flat(較保守、較貼近手冊敘述「你的」指揮評等而非「每個殖民地」的指揮評等),
	// 待找到多殖民地存檔驗證後再確認或修正。
	CommandPointsBase = 5

	// IncomeFreighterMaintenanceHalfBCNumerator / ...Denominator 每艘「使用中」的運輸艦
	// (Freighter)每回合維護費 0.5 BC(GAME_MANUAL.pdf p.169:"each freighter that is in use costs
	// 1/2 BC per turn for maintenance")。未使用中的運輸艦不計費(手冊同段:閒置的運輸艦可以直接
	// 報廢,未提及維護費)。
	IncomeFreighterMaintenanceNumerator   = 1
	IncomeFreighterMaintenanceDenominator = 2

	// FreighterFleetShipsPerBuild 每建成一次「運輸艦隊」(Freighter Fleet)取得的運輸艦艘數
	// (GAME_MANUAL.pdf p.168:"Every time you build a Freighter Fleet, you gain a group of 5
	// ships"；MANUAL_150.html 同段重申 "Every time you build one of these, you get 5 support
	// ships")。兩份手冊皆明確給 5,無版本差異,故不進 RuleProfile。
	FreighterFleetShipsPerBuild = 5

	// IncomeMoraleProductionPercentPerIcon 殖民地畫面上每一格士氣圖示(笑臉/哭臉),代表該殖民地
	// 「總產出」(食物、工業、科研、收入)變化 10%(GAME_MANUAL.pdf p.170:"Every morale icon on
	// the Colony screen represents a change of 10% in the total production output of the colony.
	// Populations with high morale work harder, adding to the food, industry, science, and income
	// of a world.")。手冊未列出圖示數量上限,故本檔不假設上限,由呼叫端傳入淨圖示數(笑臉為正、
	// 哭臉為負)。
	IncomeMoraleProductionPercentPerIcon = 10

	// IncomeGovtBonusDemocracyMoneyPercent / IncomeGovtBonusFederationMoneyPercent 政府形式對
	// 「money」(BC/稅收)的加成百分比,移植自 MANUAL_150.html「Modding with Config → Additional
	// Settings」:
	//   govt_bonus democracy_money  = 10;
	//   govt_bonus federation_money = 15;
	// 該節明文給出換算公式:"value * 5 is the percent bonus for the item, for example democracy
	// has a 10 * 5 = 50% bonus to research."——故 democracy_money=10 → 10*5=50%,
	// federation_money=15 → 15*5=75%。手冊只列出這兩種政府對 money 的加成(其餘政府
	// govt_bonus 只列出 science/food/production,未列 money),因此本檔只移植這兩個常數。
	IncomeGovtBonusDemocracyMoneyPercent  = 50
	IncomeGovtBonusFederationMoneyPercent = 75

	// incomeGovtBonusRawToPercentMultiplier MANUAL_150.html 給的 govt_bonus 原始值換算百分比的
	// 乘數(同上引文:"value * 5 is the percent bonus")。保留供測試驗證換算公式本身,不對外
	// 匯出(未使用通用 helper 命名,避免撞名)。
	incomeGovtBonusRawToPercentMultiplier = 5
	incomeGovtBonusDemocracyMoneyRaw      = 10
	incomeGovtBonusFederationMoneyRaw     = 15
)

// IncomeTaxRateIsValid 檢查稅率是否為手冊允許的合法值:0%~50%,10% 為一級距
// (GAME_MANUAL.pdf p.37)。
func IncomeTaxRateIsValid(taxRatePercent int) bool {
	if taxRatePercent < TaxRateMinPercent || taxRatePercent > TaxRateMaxPercent {
		return false
	}
	return (taxRatePercent-TaxRateMinPercent)%TaxRateStepPercent == 0
}

// IncomeTaxRevenue 依帝國稅率,從殖民地總產能中換算出的稅收 BC。稅率所扣的產能以 1:1 比例
// 轉成 BC(GAME_MANUAL.pdf p.37 / p.168)。taxRatePercent 不做合法性檢查(呼叫端應先用
// IncomeTaxRateIsValid 檢查),此處僅計算,採整數無條件捨去(手冊未提及進位規則,採本專案其他
// 換算式一致的無條件捨去慣例,見 IncomeFoodSurplusRevenue 註解)。
func IncomeTaxRevenue(totalIndustry, taxRatePercent int) int {
	return totalIndustry * taxRatePercent * TaxConversionNumerator / (100 * TaxConversionDenominator)
}

// IncomeTaxRemainingIndustry 稅率扣除後,實際留給殖民地建造用的產能
// (GAME_MANUAL.pdf p.168:"if your tax rate is an astronomical 50%, fully half your production
// potential goes toward taxes. Only the remaining half is available for building.")。
func IncomeTaxRemainingIndustry(totalIndustry, taxRatePercent int) int {
	return totalIndustry - IncomeTaxRevenue(totalIndustry, taxRatePercent)
}

// TradeGoodsIncome 貿易財(Trade Goods)產出換算成的 BC。一般種族每 2 點分配到貿易財的產能
// 換 1 BC;Fantastic Trader 種族每 1 點產能換 1 BC(GAME_MANUAL.pdf p.70)。換算採整數無條件
// 捨去(與手冊其他半 BC 換算的慣例一致,見 IncomeFoodSurplusRevenue 註解)。
func TradeGoodsIncome(industryAllocated int, fantasticTrader bool) int {
	if fantasticTrader {
		return industryAllocated * TradeGoodsFantasticTraderConversionNumerator / TradeGoodsFantasticTraderConversionDenominator
	}
	return industryAllocated * TradeGoodsConversionNumerator / TradeGoodsConversionDenominator
}

// IncomeFoodSurplusRevenue 出售剩餘糧食換得的 BC。一般種族每單位換 0.5 BC,Fantastic Trader
// 每單位換 1 BC(GAME_MANUAL.pdf p.25)。手冊沒有明講 0.5 BC 是否無條件捨去,但同一份文件
// (MANUAL_150.html 1.31/1.40/1.50 行為比較表)在描述類似的半 BC 換算(運輸艦短缺維護費)時
// 明文寫「0.5 BC each, rounded down」,本檔沿用相同的無條件捨去慣例以求可驗證、可重現。
func IncomeFoodSurplusRevenue(surplusFoodUnits int, fantasticTrader bool) int {
	if fantasticTrader {
		return surplusFoodUnits * IncomeFoodSurplusFantasticTraderPerUnitBC
	}
	return surplusFoodUnits * IncomeFoodSurplusNumerator / IncomeFoodSurplusDenominator
}

// IncomeCommandOverflowCost 指揮評等不足時,每回合從收入扣除的維護費
// (GAME_MANUAL.pdf p.169)。uncoveredCommandPoints 為負值或 0 時回傳 0(代表沒有超支)。
func IncomeCommandOverflowCost(uncoveredCommandPoints int) int {
	if uncoveredCommandPoints <= 0 {
		return 0
	}
	return uncoveredCommandPoints * IncomeCommandOverflowCostPerPoint
}

// IncomeFreighterMaintenanceCost 使用中運輸艦每回合的維護費總和,每艘 0.5 BC,無條件捨去
// (GAME_MANUAL.pdf p.169:"each freighter that is in use costs 1/2 BC per turn for maintenance")。
func IncomeFreighterMaintenanceCost(activeFreighters int) int {
	if activeFreighters <= 0 {
		return 0
	}
	return activeFreighters * IncomeFreighterMaintenanceNumerator / IncomeFreighterMaintenanceDenominator
}

// IncomeMoraleAdjustedProduction 依士氣圖示淨值(笑臉為正、哭臉為負),調整殖民地總產出
// (食物/工業/科研/收入皆適用同一比例,GAME_MANUAL.pdf p.170)。每一格圖示 = ±10%。
// 例如 netMoraleIcons = 2 時,產出變成 baseProduction * 120 / 100。
func IncomeMoraleAdjustedProduction(baseProduction, netMoraleIcons int) int {
	percent := 100 + netMoraleIcons*IncomeMoraleProductionPercentPerIcon
	return baseProduction * percent / 100
}

// IncomeApplyGovernmentMoneyBonus 套用政府形式對 BC 收入的加成百分比(如
// IncomeGovtBonusDemocracyMoneyPercent / IncomeGovtBonusFederationMoneyPercent),回傳加成後的
// BC(無條件捨去)。bonusPercent 為 0 時原值傳回(手冊未列出加成的政府,呼叫端應傳入 0)。
func IncomeApplyGovernmentMoneyBonus(baseBC, bonusPercent int) int {
	return baseBC * (100 + bonusPercent) / 100
}

// IncomeGovtMoneyBonusPercent 依政府型態回傳「money」(BC/稅收)加成百分比,供
// IncomeApplyGovernmentMoneyBonus 使用。MANUAL_150.html「Modding with Config → Additional
// Settings」只列出 Democracy(govt_bonus democracy_money=10 → 10*5=50%)與 Federation
// (federation_money=15 → 15*5=75%)這兩種政府對 money 有加成,其餘政府(含 Feudalism/
// Confederation/Dictatorship/Imperium/Unification/Galactic Unification)手冊未列出對應項目,
// 回 0(非漏列,是手冊該節本來就只給這兩項)。
//
// gov 用 morale.go 既有的 MoraleGovernmentType(政府列舉本已在該檔定義,見其註解:enums.go
// 目前無通用 Government 型別,故沿用同一個列舉,不另建第二份撞名的政府型別)。
func IncomeGovtMoneyBonusPercent(gov MoraleGovernmentType) int {
	switch gov {
	case MoraleGovDemocracy:
		return IncomeGovtBonusDemocracyMoneyPercent
	case MoraleGovFederation:
		return IncomeGovtBonusFederationMoneyPercent
	default:
		return 0
	}
}
