package gamedata

// armor_tech.go:裝甲科技的**結構倍率階梯**——手冊逐句給的數字,不是估計值。
//
// ============ 一則撤回 ============
//
// 這一輪稍早(第 123 項一帶)曾經寫下並據以停手的一句話:
//
//	裝甲科技的倍率**手冊與 openorion2 都沒有**,所以 armorHPByName 的值不動。
//
// **那句話是錯的。** 手冊 Ship 條目裡逐級寫著,只是我沒讀到那幾頁:
//
//	Tritanium Armor (Ship)   increases the structural integrity of ships and fighters
//	                         by **100%**
//	Zortrium Armor (Ship)    increases the structural integrity ... by **300%**
//	Neutronium Armor (Ship)  boost the structural hits of ships and fighters by **500%**
//	Adamantium Armor (Ship)  increases the structural hits ... by **700%**
//	Xentronium Armor         Ships with this armor have **10 times** the base structure
//	                         and armor points
//
// 鈦裝甲沒有倍率描述——它是基準(「standard armor for FTL ships」),所以是 100%。
// 得到的階梯是 100 / 200 / 400 / 600 / 800 / 1000,乾淨到不像巧合。
//
// 記在這裡而不是默默改掉:當初那句話被寫進註解、又被後續幾項當成既有結論引用,
// 而**推翻它的證據一直就在手冊裡**。這正是 `rules/00-rules-index.md` 的
// 「要推翻既有斷言之前,先找出當初支持它的證據」的反面——當初根本沒有證據,
// 只有「我找不到」,而那兩件事被寫成了同一句話。
//
// ============ 這個階梯**不是**直接的裝甲 HP ============
//
// 手冊講的是「structural integrity / structural hits」——在 MOO2 裡裝甲科技決定的是
// **艦艇本身的結構點數**,沒有獨立的「裝甲池」。remake 用的是兩池抽象
// (`combatant.armor` 先擋、耗盡才傷 `hp`),那是 remake 的模型不是原版的。
//
// 所以這張表提供的是**倍率**,由呼叫端決定乘在哪裡;`shell.armorHPByName` 目前把它
// 乘在一個 remake 自訂的基準單位上(見該函式)。**階梯是一手的,基準單位是 remake 值**
// ——兩者分開記,免得日後有人把整條當成原版真值引用。

// 裝甲科技的結構倍率(百分點,100 = 基準)。
const (
	// ArmorStructurePercentTitanium 鈦裝甲是基準(手冊:「standard armor for FTL ships」,
	// 未給倍率)。
	ArmorStructurePercentTitanium = 100
	// ArmorStructurePercentTritanium 手冊:「increases the structural integrity of ships and
	// fighters by 100%」→ 200%。
	ArmorStructurePercentTritanium = 200
	// ArmorStructurePercentZortrium 手冊:「increases the structural integrity ... by 300%」。
	ArmorStructurePercentZortrium = 400
	// ArmorStructurePercentNeutronium 手冊:「boost the structural hits ... by 500%」。
	ArmorStructurePercentNeutronium = 600
	// ArmorStructurePercentAdamantium 手冊:「increases the structural hits ... by 700%」。
	ArmorStructurePercentAdamantium = 800
	// ArmorStructurePercentXentronium 手冊:「Ships with this armor have 10 times the base
	// structure and armor points」——這一條是**倍數**不是增幅,所以是 1000% 不是 1100%。
	ArmorStructurePercentXentronium = 1000
)

// ArmorStructurePercent 依裝甲科技回傳結構倍率(百分點);查無回 100(視同無加成)。
func ArmorStructurePercent(tech Technology) int {
	switch tech {
	case TECH_TITANIUM_ARMOR:
		return ArmorStructurePercentTitanium
	case TECH_TRITANIUM_ARMOR:
		return ArmorStructurePercentTritanium
	case TECH_ZORTRIUM_ARMOR:
		return ArmorStructurePercentZortrium
	case TECH_NEUTRONIUM_ARMOR:
		return ArmorStructurePercentNeutronium
	case TECH_ADAMANTIUM_ARMOR:
		return ArmorStructurePercentAdamantium
	case TECH_XENTRONIUM_ARMOR:
		return ArmorStructurePercentXentronium
	}
	return 100
}

// ArmorHeavyArmorMultiplier 是重裝甲系統對裝甲耐受量的倍率。
//
// 手冊逐字(Heavy Armor (System)):「Installing Heavy Armor **triples** the amount of damage
// the ship's armor can sustain before damage gets through to the structure and internal
// systems. This system also **negates the Armor Piercing abilities** of enemy weapons that
// hit the ship.」
//
// 兩個效果都在這一條裡:倍率用這個常數,穿甲抵銷見 DamageApplyArmor 的 apNegated 參數。
const ArmorHeavyArmorMultiplier = 3

// ArmorNegatesArmorPiercing 回報這件裝甲科技本身是否使敵方穿甲失效。
//
// 手冊只給氙素裝甲這一條(「Negates armor piercing effects of enemy weapons」);
// 重裝甲是**系統**不是裝甲科技,由呼叫端另外判斷後與本函式的結果取聯集
// (見 DamageApplyArmor 的 apNegated 參數註解)。
func ArmorNegatesArmorPiercing(tech Technology) bool {
	return tech == TECH_XENTRONIUM_ARMOR
}

// --- 艦載系統的固定加成(手冊逐句,第 134 項)---

const (
	// ShipInertialStabilizerBeamDefense 手冊(Inertial Stabilizer):「a +50 addition to the
	// ship's beam defense」。與 BeamDefense(移植自 openorion2 ShipDesign::beamDefense)同值。
	ShipInertialStabilizerBeamDefense = 50
	// ShipInertialNullifierBeamDefense 慣性抵消器的對應值(BeamDefense 的 +100 分支)。
	ShipInertialNullifierBeamDefense = 100
	// ShipBattleScannerBeamOffense 手冊(Battle Scanner):「The scanner increases the ship's
	// chance to hit with beam weapons by 50.」與 BeamOffense 的 battleScanner 分支同值。
	ShipBattleScannerBeamOffense = 50
	// ShipReinforcedHullStructurePercent 手冊(Reinforced Hull):「triples the amount of
	// structural damage a ship can sustain before being destroyed」→ 300%。
	ShipReinforcedHullStructurePercent = 300
	// ShipMultiPhasedShieldPercent 手冊(Multi-Phased Shields):「increasing the maximum
	// amount of damage that they can absorb by 50%」→ 150%。
	ShipMultiPhasedShieldPercent = 150
)

// shipScoutLabResearch 是偵察實驗室依艦體等級每回合產生的研究點數。
//
// 手冊罕見地把整張表列出來了:「Frigate = 1, Destroyer = 2, Cruiser = 4, Battleship = 8,
// Titan = 16, and Doom Star = 32.」索引對應 CombatShipClass(0=巡防艦 … 5=末日之星)。
var shipScoutLabResearch = [6]int{1, 2, 4, 8, 16, 32}

// ShipScoutLabResearch 回傳偵察實驗室在該艦體等級每回合的研究點數;超界回 0。
func ShipScoutLabResearch(class CombatShipClass) int {
	if int(class) < 0 || int(class) >= len(shipScoutLabResearch) {
		return 0
	}
	return shipScoutLabResearch[class]
}

// ShipStructuralAnalyzerMultiplier 是結構分析儀對「已穿過護盾」的光束傷害的倍率。
//
// 手冊逐字(Structural Analyzer):「the damage done by beam weapons that penetrate an
// enemy ship's shields is **doubled**.」
const ShipStructuralAnalyzerMultiplier = 2
