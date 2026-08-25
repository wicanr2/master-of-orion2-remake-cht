package gamedata

// 殖民地生產/污染唯讀公式,移植自 GAME_MANUAL.pdf(moo2_patch1.5 隨附的完整遊戲手冊,
// 「System Overview / Yield / Population」章節,約 p.64-66)與各建築說明章節(p.80-90 附近)。
//
// moo2_patch1.5/MANUAL_150.html(1.50 patch 說明書)本身只有 1 處提到 pollution、且是 UI 顯示
// bugfix 敘述(時空異象不再顯示汙染桶圖示),沒有生產/污染的數值公式或 Notes 附錄(該手冊唯一的
// 數值公式附錄是「Notes on Population Growth」,已在 colony.go 移植)。其 Modding 章節的
// Productivity 小節(p.77)也只列 weather_control_climate / fungi_climate / terraform_toxic_building
// 三個參數,無汙染或工廠加成的可調參數。因此本檔案數值全部來自 GAME_MANUAL.pdf。
// openorion2 的 gamestate.h/.cpp 只有 Colony.pollution 等唯讀存檔欄位與 Planet::baseProduction
// (已在 formulas.go 的 PlanetBaseProduction 移植),並未實作生產/污染的計算公式。
//
// 命名前綴:Prod = 生產(工人/工廠加成),Pollution = 污染(容忍值/清理成本/建築減免)。

const (
	// ProdWorkerMinimum 每個工人單位至少產出 1 產能,不論星球狀況如何
	// (GAME_MANUAL.pdf:"Each unit produces at least 1 production, no matter what the situation.")。
	ProdWorkerMinimum = 1

	// ProdAutomatedFactoryPerWorkerBonus 自動化工廠(Automated Factory)使每個工業人口單位
	// 額外 +1 產能(GAME_MANUAL.pdf:"increasing the output of each industrial unit of population
	// by +1 production each turn")。
	ProdAutomatedFactoryPerWorkerBonus = 1
	// ProdAutomatedFactoryFlatBonus 自動化工廠額外給殖民地 +5 產能(同段:"giving the colony
	// +5 production")。
	ProdAutomatedFactoryFlatBonus = 5

	// ProdDeepCoreMinePerWorkerBonus 深層核心礦(Deep Core Mine)使每個工人單位額外 +3 產能
	// (GAME_MANUAL.pdf:"increases the productivity of each worker unit by 3 production")。
	ProdDeepCoreMinePerWorkerBonus = 3
	// ProdDeepCoreMineFlatBonus 深層核心礦額外給殖民地 +15 產能(同段:"and the colony by 15")。
	ProdDeepCoreMineFlatBonus = 15

	// ProdRecyclotronPerPopulation 再生反應爐(Recyclotron)使每個人口單位(不論職業)產出
	// 1 點工業產能,且此額外產能不計入污染(GAME_MANUAL.pdf:"each unit of population generates
	// 1 industrial production, regardless of its assigned job. This increased production does
	// not count toward the planetary pollution level")。
	ProdRecyclotronPerPopulation = 1

	// ProdMicroliteConstructionPerWorkerBonus 微晶構築(Microlite Construction)成就使全帝國
	// 每個工業工人額外 +1 產能(GAME_MANUAL.pdf:"increases the output of all your empire's
	// industrial workers by 1 production per turn each")。
	ProdMicroliteConstructionPerWorkerBonus = 1

	// ProdAlienUncooperativeNumerator / ProdAlienUncooperativeDenominator:被征服殖民地上未整合
	// 的外星人口(Aliens),在整合前每單位只產出正常值的 3/4(GAME_MANUAL.pdf:"each alien unit
	// produces only three quarters what it normally would")。
	ProdAlienUncooperativeNumerator   = 3
	ProdAlienUncooperativeDenominator = 4

	// PollutionNanoDisassemblersMultiplier 奈米分解者(Nano Disassemblers)成就使星球的污染
	// 容忍值加倍(GAME_MANUAL.pdf:"doubles the planet's inherent tolerance to pollution")。
	PollutionNanoDisassemblersMultiplier = 2
)

// roboticFactoryBonusTable 機器人工廠(Robotic Factory)依礦產豐度給殖民地的產能加成
// (GAME_MANUAL.pdf:"+5 on Ultra Poor worlds, +8 for Poor, +10 on Abundant planets, +15 for
// Rich, and +20 on Ultra Rich worlds")。索引順序與 formulas.go 的 mineralProductionTable
// 一致(PlanetMinerals 0-4:ULTRA_POOR, POOR, ABUNDANT, RICH, ULTRA_RICH)。
var roboticFactoryBonusTable = [5]int{5, 8, 10, 15, 20}

// ProdWorkerOutput 套用工人單位最低產出下限(ProdWorkerMinimum)。base 為未套用下限前的
// 每工人產能(例如礦產基礎值 + 自動化工廠/深層核心礦等加成),回傳值保證 >= 1。
func ProdWorkerOutput(base int) int {
	if base < ProdWorkerMinimum {
		return ProdWorkerMinimum
	}
	return base
}

// ProdRoboticFactoryBonus 回傳機器人工廠依礦產豐度(PlanetMinerals 0-4)給殖民地的產能加成。
// minerals 超出範圍回 0。
func ProdRoboticFactoryBonus(minerals int) int {
	if minerals < 0 || minerals >= len(roboticFactoryBonusTable) {
		return 0
	}
	return roboticFactoryBonusTable[minerals]
}

// ProdAlienWorkerOutput 未整合外星人口的實際產出:base * 3/4(向下取整)。
func ProdAlienWorkerOutput(base int) int {
	return base * ProdAlienUncooperativeNumerator / ProdAlienUncooperativeDenominator
}

// PollutionTolerance 星球污染容忍值 = 2 * 星球尺寸等級(1-based:Tiny=1...Huge=5)
// (GAME_MANUAL.pdf:"determined as twice the planet's size class. For example, a medium
// planet (size class 3) has a pollution tolerance of 6 production.")。
// size 為 enums.go 的 PlanetSize(0-based:TINY_PLANET=0...HUGE_PLANET=4),故容忍值 = 2*(size+1);
// MEDIUM_PLANET(=2)代入得 2*(2+1)=6,與手冊範例相符。
func PollutionTolerance(size PlanetSize) int {
	return 2 * (int(size) + 1)
}

// PollutionToleranceWithNanoDisassemblers 已研究奈米分解者成就後的污染容忍值(加倍)。
func PollutionToleranceWithNanoDisassemblers(size PlanetSize) int {
	return PollutionTolerance(size) * PollutionNanoDisassemblersMultiplier
}

// PollutionCleanupCost 回傳清理污染所需消耗的產能。
// 基礎規則(GAME_MANUAL.pdf):產能超出污染容忍值(tolerance)的部分,其中一半用於清理污染
// ("Half of the production exceeding the planet's pollution tolerance is used to clean up
// pollution.")。超出部分為奇數時,採向下取整(手冊未逐字標明本規則的捨入方向,但同份手冊在
// 貨運船維護費敘述中使用「0.5 BC each, rounded down」的慣例,此處沿用向下取整)。
// tolerantRace:Tolerant 特性種族(含矽晶生物 Silicoids)不受污染影響、不需花費產能清理
// (GAME_MANUAL.pdf:"Tolerant races also suffer no harm from pollution and need not spend
// production resources cleaning it up." / Silicoids:"spend no effort to clean up pollution")。
func PollutionCleanupCost(production, tolerance int, tolerantRace bool) int {
	if tolerantRace {
		return 0
	}
	excess := production - tolerance
	if excess <= 0 {
		return 0
	}
	return excess / 2
}

// PollutionEighths 回傳「仍會產生污染」的產能比例,以 8 分之幾表示(0-8)。
// 依手冊給出的精確分數組合(GAME_MANUAL.pdf):
//   - 無建築:8/8(全部產能都計入污染)。
//   - 只有污染處理器(Pollution Processor):"process the waste from fully half of the colony's
//     production" → 4/8(1/2)。
//   - 只有大氣更新器(Atmospheric Renewer):"cuts out the pollution produced by three-quarters
//     of the industry" → 剩 1/4 → 2/8。
//   - 兩者皆有:"only one-eighth of the industry produces pollution" → 1/8(與前兩者相乘
//     1/2 * 1/4 = 1/8 一致,非額外假設)。
//   - 核心廢料場(Core Waste Dump):取代污染處理器與大氣更新器,"eliminates all pollution on
//     the planet" → 0/8,不論另兩者是否已建。
//
// 注意:手冊只描述「污染處理器/大氣更新器如何折算仍會致污染的產能比例」,並未給出這個比例
// 與 PollutionTolerance/PollutionCleanupCost 兩段式規則合併運算的逐字範例;若要合併使用,
// 呼叫端需自行決定套用順序(例如：先用本函式縮減產能,再套用 PollutionCleanupCost)。
func PollutionEighths(pollutionProcessor, atmosphericRenewer, coreWasteDump bool) int {
	if coreWasteDump {
		return 0
	}
	e := 8
	if pollutionProcessor {
		e /= 2
	}
	if atmosphericRenewer {
		e /= 4
	}
	return e
}

// PollutionPollutingProduction 回傳「仍會產生污染」的實際產能(production * eighths / 8,
// 向下取整),eighths 建議由 PollutionEighths 取得。
func PollutionPollutingProduction(production, eighths int) int {
	return production * eighths / 8
}

// PollutionReducedByPercent 把「仍會產生污染的產能」再打掉一個百分比(環保官領袖技能)。
//
// ============ 為什麼是接在這裡 ============
//
// 手冊 p.137 那條逐字是:
//
//	Environmentalist: Reduces **the amount of production that causes pollution**
//	on the colonies in the system.
//
// 「the amount of production that causes pollution」就是 `PollutionPollutingProduction`
// 回傳的那個量——這支函式的名字本身就是那句話。所以接的位置是**八分之幾查表之後、
// 容忍值扣除之前**:eighths → 環保官 → tolerance → 一半。
//
// ⚠ 這不是「減產能」。環保官降的是**會致污染的那部分產能**,殖民地實際的工業產出不變
// (它只讓 `PollutionCleanupCost` 少扣一點)。接到 gross 那一行會變成「請一個環保官等於少一個工人」,
// 那是手冊那句話的反面。
//
// ============ 為什麼是相乘,不是相加 ============
//
// 這條規則手冊沒有直接寫,但它給了同一類效果怎麼疊的**算術**(p.90 大氣更新器):
//
//	This effect is cumulative with that of the Pollution Processor; if both are in place,
//	**only one-eighth** of the industry produces pollution.
//
// 污染處理器留下 1/2、大氣更新器留下 1/4,兩者並存留下 **1/8 = 1/2 × 1/4**。
// (手冊用的字是 "cumulative",但它自己算出來的數字是相乘——**數字贏字面**:
// 相加會是 1/2 + 1/4 或 3/4 之類的值,得不到 1/8。)
//
// 所以環保官的百分比也當成同一條鏈上的另一個乘數,套在查表結果之上。這是手冊給過的
// 唯一一條「污染削減怎麼疊」的規則,不是另創的模型。
//
// reductionPercent 是**正的減幅**(環保官的 skillBonus 是負值,呼叫端負負得正);
// <=0 原封不動,>=100 全部歸零。
func PollutionReducedByPercent(pollutingProd, reductionPercent int) int {
	if reductionPercent <= 0 || pollutingProd <= 0 {
		return pollutingProd
	}
	if reductionPercent >= 100 {
		return 0
	}
	return pollutingProd * (100 - reductionPercent) / 100
}

// ============ 未整合外星人的 3/4 產出(2026-08-08 第 57 項(征服人口產出)接上)============
//
// 手冊(GAME_MANUAL.pdf,Population 一節)講被征服殖民地上的「Aliens」圖示:
//
//	Aliens appear in an enemy's colony that you've conquered, representing the population
//	left there by the former owner. At first, all aliens are **uncooperative**. Until they
//	are integrated into your empire, **each alien unit produces only three quarters what it
//	normally would**.
//
// 注意手冊寫的是「**produces**」——不限工業。所以食物、工業、研究三項都套。
//
// ColonyState 現已保存逐職務 prisoner 數；本檔的比例 helper 只供沒有新欄位的舊 JSON
// 遷移 fallback。原版逐人口證據見 docs/re/colonist-prisoner-production-audit-20260825.md。
//
// 這一段同時把 `ProdWorkerOutput` 那條「每工人至少 1 產能」的下限接上了:
// 先前礦產表最低就是 1(`mineralProductionTable = {1,2,3,5,8}`),下限**永遠不會生效**,
// 那支函式因此零覆蓋。有了 3/4 之後 `1 × 3/4 = 0`,下限這才真的擋到東西。

// UncooperativeAlienUnits 回傳某職業裡有幾個單位是未整合的外星人。
//
// 按人口比例向下取整(見上方 ⚠)。population <= 0 或沒有未整合人口時回 0;
// unassimilated 超過 population 時夾住——征服/同化那條路徑若算出不一致的數,
// 這裡不該把它放大成「外星人比總人口還多」。
func UncooperativeAlienUnits(jobUnits, unassimilated, population int) int {
	if jobUnits <= 0 || unassimilated <= 0 || population <= 0 {
		return 0
	}
	if unassimilated > population {
		unassimilated = population
	}
	n := jobUnits * unassimilated / population
	if n > jobUnits {
		n = jobUnits
	}
	return n
}

// UncooperativeJobOutput 回傳一個職業的總產出,其中未整合外星人那一份只有 3/4。
//
// applyWorkerFloor 只有**工業**要傳 true:`ProdWorkerMinimum` 的手冊依據講的是
// 「每個**工人**單位至少產出 1 產能」,沒有講農夫與科學家,所以不擅自套到那兩項。
func UncooperativeJobOutput(jobUnits, perUnit, unassimilated, population int, applyWorkerFloor bool) int {
	return UncooperativeJobOutputExact(jobUnits, perUnit,
		UncooperativeAlienUnits(jobUnits, unassimilated, population), applyWorkerFloor)
}

// UncooperativeJobOutputExact 使用該職務內實際帶 PRISONER flag 的人口數計算。
// 原版 sub_DE280 @ 0xDE280 先依 job bits 篩選，再對 [colonist+1]&4 的人口
// 扣除 base*5/20；所以 prisoner 的每人產出正好是 3/4。
func UncooperativeJobOutputExact(jobUnits, perUnit, prisoners int, applyWorkerFloor bool) int {
	if jobUnits <= 0 || perUnit <= 0 {
		return 0
	}
	if prisoners < 0 {
		prisoners = 0
	}
	if prisoners > jobUnits {
		prisoners = jobUnits
	}
	if prisoners == 0 {
		return jobUnits * perUnit
	}
	alienPerUnit := ProdAlienWorkerOutput(perUnit)
	if applyWorkerFloor {
		alienPerUnit = ProdWorkerOutput(alienPerUnit)
	}
	return (jobUnits-prisoners)*perUnit + prisoners*alienPerUnit
}
