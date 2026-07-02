package gamedata

// 地面戰鬥(Ground Combat / Invasion)公式,移植自:
//   - moo2_patch1.5/GAME_MANUAL.pdf(pdftotext -layout):
//       p.15-16  Race Picks「Ground Combat」種族加成(Bulrathi / Gnolam)
//       p.21     Combat Modifiers 段落對「Ground Combat modifiers」的定義
//       p.24     Special Abilities:Low-G World / High-G World / Subterranean
//       p.27     Special Abilities:Warlord(barracks 容量加倍)
//       p.77     Marine Barracks(Building)
//       p.79     Troop Pods(System)/ Armor Barracks(Building)
//       p.80     Powered Armor(Equipment)
//       p.81     Battleoids(Equipment)
//       p.85     Transport Ship(每艘建成配 4 個 Marine 單位)
//       p.90     Tritanium Armor(Advanced Metallurgy)對地面部隊戰力加成
//       p.91     Zortrium Armor(Nano Technology)/ Neutronium Armor(Molecular Manipulation)
//       p.92     Adamantium Armor(Molecular Control)
//       p.108    Anti-Grav Harness(Gravitic Fields)
//       p.109    Personal Shield(Electromagnetic Refraction)
//       p.114    Xentronium Armor(Armor technology)
//       p.162-164 Invading a Colony / Ground Combat(流程敘述,無額外數字公式)
//   - moo2_patch1.5/MANUAL_150.html(python 去標籤,依其自身內部頁碼):
//       p.129    Notes on Orbital Assault > Orbital Bombardment(Estimated Bomb Hits / Planet Hits 表)
//
// openorion2 未實作地面戰鬥判定邏輯(只有存檔欄位與 tech/building 名稱),本檔為手冊到程式碼
// 的首次移植。只搬手冊明確列出、附精確數字的公式/表;沒有精確數字的一律標 `TODO 手冊未明列`,
// 不臆測填數字(見檔尾)。
//
// 命名一律加 Ground 前綴,避免與其他檔案的通用 helper 撞名。

// --- Marine / Armor Barracks 建造與人口上限(手冊 p.77, p.79) ---

const (
	// GroundMarineBarracksInitialUnits 手冊 p.77:「When first built, a Marine Barracks
	// immediately generates up to 4 Marine units.」
	GroundMarineBarracksInitialUnits = 4
	// GroundMarineBarracksTurnsPerUnit 手冊 p.77:「The barracks train 1 new unit of ground
	// troops every 5 turns」。
	GroundMarineBarracksTurnsPerUnit = 5

	// GroundArmorBarracksInitialUnits 手冊 p.79:「When first built, an Armor Barracks
	// immediately produces up to 2 armor battalions」。
	GroundArmorBarracksInitialUnits = 2
	// GroundArmorBarracksTurnsPerUnit 手冊 p.79:「then another tank battalion every 5 turns」。
	GroundArmorBarracksTurnsPerUnit = 5

	// GroundWarlordBarracksMultiplier 手冊 p.27(Warlord):「Warlord barracks — Marines and
	// Armor — can support twice the usual number of ground troops.」
	GroundWarlordBarracksMultiplier = 2

	// GroundTransportShipMarineCapacity 手冊 p.85(Transport Ship):「As a Transport Ship is
	// built, 4 new Marine units are created to fill it.」
	GroundTransportShipMarineCapacity = 4

	// GroundTroopPodsMultiplier 手冊 p.79(Troop Pods):「doubling the number of Marines on
	// board a ship」。
	GroundTroopPodsMultiplier = 2
)

// GroundMarineBarracksCap Marine Barracks 可維持的部隊上限。
// 手冊原文(p.77):「up to a maximum equal to half the current population of the colony or
// half the base maximum population of that size planet, whichever is less」。currentPop /
// planetMaxPop 皆為人口單位數(整數),「half」採整數除法(向下取整,手冊未進一步說明取捨方向)。
// warlord=true 時依 p.27 Warlord 特性加倍。
func GroundMarineBarracksCap(currentPop, planetMaxPop int, warlord bool) int {
	a := currentPop / 2
	b := planetMaxPop / 2
	cap := a
	if b < a {
		cap = b
	}
	if warlord {
		cap *= GroundWarlordBarracksMultiplier
	}
	return cap
}

// GroundArmorBarracksCap Armor Barracks 可維持的戰車營上限。
// 手冊原文(p.79):「up to a maximum equal to one-quarter the current population of the
// colony or a quarter of the base maximum population of that size planet, whichever is
// less」。同樣採整數除法。warlord=true 時依 p.27 Warlord 特性加倍。
func GroundArmorBarracksCap(currentPop, planetMaxPop int, warlord bool) int {
	a := currentPop / 4
	b := planetMaxPop / 4
	cap := a
	if b < a {
		cap = b
	}
	if warlord {
		cap *= GroundWarlordBarracksMultiplier
	}
	return cap
}

// GroundMarineBarracksUnits 已運作 turnsSinceBuilt 回合的 Marine Barracks 現有部隊數,
// 已套用 GroundMarineBarracksCap 上限。turnsSinceBuilt 為負數視為 0(尚未建成不會被呼叫,
// 這裡只防呆)。
func GroundMarineBarracksUnits(turnsSinceBuilt, currentPop, planetMaxPop int, warlord bool) int {
	if turnsSinceBuilt < 0 {
		turnsSinceBuilt = 0
	}
	n := GroundMarineBarracksInitialUnits + turnsSinceBuilt/GroundMarineBarracksTurnsPerUnit
	cap := GroundMarineBarracksCap(currentPop, planetMaxPop, warlord)
	if n > cap {
		return cap
	}
	return n
}

// GroundArmorBarracksUnits 已運作 turnsSinceBuilt 回合的 Armor Barracks 現有戰車營數,
// 已套用 GroundArmorBarracksCap 上限。
func GroundArmorBarracksUnits(turnsSinceBuilt, currentPop, planetMaxPop int, warlord bool) int {
	if turnsSinceBuilt < 0 {
		turnsSinceBuilt = 0
	}
	n := GroundArmorBarracksInitialUnits + turnsSinceBuilt/GroundArmorBarracksTurnsPerUnit
	cap := GroundArmorBarracksCap(currentPop, planetMaxPop, warlord)
	if n > cap {
		return cap
	}
	return n
}

// --- 部隊每次被擊中致死所需的「命中(hit)」數(手冊 p.129 Planet Hits 表 + p.24/p.80/p.81) ---

const (
	// GroundMarineBaseHitsToKill 手冊 p.129(Planet Hits):「Each Marine  1 hit (modified by
	// Heavy-G, Powered Armor)」。
	GroundMarineBaseHitsToKill = 1
	// GroundTankBaseHitsToKill 手冊 p.129(Planet Hits):「Each Tank  2 hits (modified by
	// Heavy-G, Battleoids)」。
	GroundTankBaseHitsToKill = 2
	// GroundBattleoidHitsToKill 手冊 p.81(Battleoids):「Battleoids have a ground combat
	// rating 10 higher than a tank and take 3 hits to destroy.」Battleoid 取代 Tank,非疊加。
	GroundBattleoidHitsToKill = 3

	// GroundHighGRaceExtraHit 手冊 p.24(High-G World):「High-G ground troops can sustain
	// substantially more physical damage than other troops; they take 1 hit more than normal
	// troops before being slain in ground combat.」對應 p.129 Planet Hits 表註記的
	// 「modified by Heavy-G」。
	GroundHighGRaceExtraHit = 1
	// GroundPoweredArmorExtraHit 手冊 p.80(Powered Armor):「Troops equipped with powered
	// armor have a bonus of 10 added to their combat rating and take 1 extra hit to kill.」
	// 只見於 Planet Hits 表 Marine 那列的修飾詞,Tank 列未列 Powered Armor,故不套用於 Tank。
	GroundPoweredArmorExtraHit = 1
)

// GroundMarineHitsToKill 單一 Marine 單位需要被命中幾次才會陣亡。
func GroundMarineHitsToKill(highGRace, poweredArmor bool) int {
	hits := GroundMarineBaseHitsToKill
	if highGRace {
		hits += GroundHighGRaceExtraHit
	}
	if poweredArmor {
		hits += GroundPoweredArmorExtraHit
	}
	return hits
}

// GroundTankHitsToKill 單一 Tank(戰車營)單位需要被命中幾次才會陣亡。
// 手冊 p.129 Planet Hits 表只列 Tank 受 Heavy-G(對應 p.24 種族 High-G 特性)與 Battleoids 修飾,
// 不含 Powered Armor,故本函式不接收 poweredArmor 參數。研究出 Battleoids 後,Tank 單位整批換成
// Battleoid,固定 3 hits(見 GroundBattleoidHitsToKill),不再套用本函式。
func GroundTankHitsToKill(highGRace bool) int {
	hits := GroundTankBaseHitsToKill
	if highGRace {
		hits += GroundHighGRaceExtraHit
	}
	return hits
}

// --- 地面部隊戰力(combat strength / combat rating)加成表 ---

// GroundArmorTechBonus 依裝甲科技(TECH_TRITANIUM_ARMOR 等,enums.go)回傳其對「所有地面部隊
// 戰力」的加成。手冊原文逐條:
//
//	p.90 Tritanium Armor  :「Tritanium alloy used in other equipment adds 10 to all
//	                          ground troop combat strengths.」
//	p.91 Zortrium Armor   :「Zortrium body armor adds 15 to the combat strength of all
//	                          ground troops.」
//	p.91 Neutronium Armor :「Neutronium-laced armor adds 20 to all ground troop combat
//	                          strengths.」
//	p.92 Adamantium Armor :「Adamantium-based systems add 25 to the combat strength of
//	                          all ground troops.」
//	p.114 Xentronium Armor:「Adds +30 to ground troop combat strengths.」
//
// 基礎的 Titanium Armor(TECH_TITANIUM_ARMOR)手冊未提供地面戰力加成,回 0。傳入其他未列出
// 的科技一律回 0(不臆測)。同一艘殖民地應只套用「目前已知最佳」的一項,不與較低階者疊加
// (手冊逐條都是各自獨立描述最佳裝甲的加成,未提及疊加規則)。
func GroundArmorTechBonus(tech Technology) int {
	switch tech {
	case TECH_TRITANIUM_ARMOR:
		return 10
	case TECH_ZORTRIUM_ARMOR:
		return 15
	case TECH_NEUTRONIUM_ARMOR:
		return 20
	case TECH_ADAMANTIUM_ARMOR:
		return 25
	case TECH_XENTRONIUM_ARMOR:
		return 30
	default:
		return 0
	}
}

// GroundEquipmentTechBonus 依地面裝備科技(TECH_ANTIGRAV_HARNESS 等,enums.go)回傳其對地面
// 部隊戰力/戰鬥評等的加成。手冊原文逐條:
//
//	p.80  Powered Armor    :「a bonus of 10 added to their combat rating」(另見
//	                          GroundPoweredArmorExtraHit 的 +1 hit)。
//	p.108 Anti-Grav Harness:「adding 10 to their ground combat rating.」
//	p.109 Personal Shield  :「increasing the combat rating of Marines and armor by 20.」
//
// 傳入其他未列出的科技一律回 0(不臆測)。
func GroundEquipmentTechBonus(tech Technology) int {
	switch tech {
	case TECH_POWERED_ARMOR:
		return 10
	case TECH_ANTIGRAV_HARNESS:
		return 10
	case TECH_PERSONAL_SHIELD:
		return 20
	default:
		return 0
	}
}

// GroundBattleoidCombatBonus 手冊 p.81(Battleoids):「Battleoids have a ground combat
// rating 10 higher than a tank」。此為相對 Tank 的加成,非獨立疊加項。
const GroundBattleoidCombatBonus = 10

// GroundRace 手冊種族創建畫面中,對「Ground Combat」有明確數字加成/懲罰的種族。
// enums.go 的 RaceTrait.TRAIT_GROUND_COMBAT 只標記「此特性存在」,不含正負與數值,
// 兩個種族的加成方向相反,故另建此列舉對應手冊數字(同 spy.go SpyGovernmentType 的作法)。
type GroundRace int

const (
	GroundRaceOther    GroundRace = iota // 手冊未列明確數字的其他種族
	GroundRaceBulrathi                   // 手冊 p.15:「Bulrathi enjoy a +10 bonus in ground combat」
	GroundRaceGnolam                     // 手冊 p.16:「Gnolams' Low-G roots put them at a -10 disadvantage in ground combat」
)

// GroundRaceCombatBonus 種族對地面戰力的固定加成(手冊 p.15-16)。
func GroundRaceCombatBonus(race GroundRace) int {
	switch race {
	case GroundRaceBulrathi:
		return 10
	case GroundRaceGnolam:
		return -10
	default:
		return 0
	}
}

// GroundLowGPenaltyPercent 手冊 p.24(Low-G World):「Low-G troops suffer a 10% penalty
// during ground combat.」(此為與 GroundRaceGnolam 的種族 -10 分開的獨立效果——Gnolam 剛好
// 兩者皆適用,但 Low-G 懲罰是任何 Low-G 種族共通的百分比懲罰,不限 Gnolam。)
const GroundLowGPenaltyPercent = 10

// GroundApplyLowGPenalty 對地面戰力套用 Low-G 種族的 10% 懲罰。手冊只給了百分比數字本身,
// 未列出「10% 套用在哪個基準值、如何捨入」的計算細節;此處採最直接的讀法——以整數戰力乘 10%
// 後捨去,交由呼叫端在需要不同捨入方式時自行調整。
func GroundApplyLowGPenalty(strength int) int {
	return strength - strength*GroundLowGPenaltyPercent/100
}

// GroundSubterraneanDefenseBonus 手冊 p.24(Subterranean):「subterranean troops receive a
// +10 ground combat bonus when defending their colonies.」僅在防守己方殖民地時生效,
// 進攻時不適用(手冊未提供攻擊情境下的數字)。
const GroundSubterraneanDefenseBonus = 10

// GroundSubterraneanBonus 依是否為防守方回傳 Subterranean 種族的地面戰力加成。
func GroundSubterraneanBonus(defending bool) int {
	if defending {
		return GroundSubterraneanDefenseBonus
	}
	return 0
}

// --- Notes on Orbital Assault > Orbital Bombardment(MANUAL_150.html p.129) ---

const (
	// GroundMaxBombHitsPerFleet 手冊原文:「The maximum number of bomb hits for the fleet in
	// orbit is 320.」
	GroundMaxBombHitsPerFleet = 320

	// GroundPlanetMissileEvasionPercent 手冊原文:「The planet has 7% missile evasion,
	// affecting missiles and torp hit chances.」
	GroundPlanetMissileEvasionPercent = 7

	// Planet Hits 表(手冊原文逐列,每項對地面設施/人口造成 1 個「hit」需求,Marine/Tank 見
	// 上方 GroundMarineBaseHitsToKill / GroundTankBaseHitsToKill):
	//   Each building                1 hit
	//   Stored Production (if >0)    1 hit (larger stored prod increases its hit chance——
	//                                 手冊未給「增加多少」的精確數字,僅保留「觸發條件」本身)
	//   Each full population         1 hit
	//   Each fraction of pop (100k)  1 hit
	GroundPlanetHitsPerBuilding         = 1
	GroundPlanetHitsPerStoredProduction = 1
	GroundPlanetHitsPerFullPop          = 1
	GroundPlanetHitsPerPopFraction      = 1
)

// GroundBombHitsFromDamage 依「模擬 10 輪齊射後的總傷害」換算成 Orbital Combat Selection
// 視窗顯示的轟炸命中(hit)數。手冊原文(Estimated Bomb Hits):「All remaining ships fire all
// weapons 10 times, or as many times as there is ammo in 10 turns... and total damage is
// calculated from it. This damage is divided by 100 to get the displayed number... The
// maximum number of bomb hits for the fleet in orbit is 320.」除以 100 用整數除法(捨去);
// 傷害總和本身(含光束/魚雷減半、電腦加成、飛彈命中率等)不在本函式範圍內,由呼叫端算好
// totalDamage 後傳入。
func GroundBombHitsFromDamage(totalDamage int) int {
	if totalDamage < 0 {
		return 0
	}
	hits := totalDamage / 100
	if hits > GroundMaxBombHitsPerFleet {
		return GroundMaxBombHitsPerFleet
	}
	return hits
}

// GroundPlanetTotalHits 依 Planet Hits 表加總防守方(建築/儲存生產/人口/部隊)需要承受的
// 總 hit 數,供轟炸模擬使用。marineHitsEach / tankHitsEach 由呼叫端依 GroundMarineHitsToKill /
// GroundTankHitsToKill(或 GroundBattleoidHitsToKill)先算好再傳入,本函式只做手冊表格定義的
// 加總(每個建築/整數人口/人口零頭各算 1 個「hit 需求單位」,Marine/Tank 各以其自身 hits-to-kill
// 計)。
func GroundPlanetTotalHits(buildings int, storedProductionPositive bool, fullPop, popFraction int, marines, marineHitsEach, tanks, tankHitsEach int) int {
	total := buildings*GroundPlanetHitsPerBuilding + fullPop*GroundPlanetHitsPerFullPop + popFraction*GroundPlanetHitsPerPopFraction
	if storedProductionPositive {
		total += GroundPlanetHitsPerStoredProduction
	}
	total += marines * marineHitsEach
	total += tanks * tankHitsEach
	return total
}

// --- TODO 手冊未明列精確數字,不臆測 ---
//
// - Commando Leader(手冊 p.135「Commando: Increases the ground combat strength of all
//   troops in the same system as the Colony leader or the strength of all marines in the
//   same fleet as the Ship Officer.」)未給基礎加成數字;MANUAL_150.html 只補充「A defending
//   commando gives 2.5x the regular commando bonus to ground troops, just like an attacking
//   commando already gives in classic.」——2.5x 是相對倍率,但「regular commando bonus」本身
//   的基準值兩份手冊都沒有給出精確數字(僅在 ORION2.CFG 找到技能訓練花費 10/30 BC 的資料,
//   與戰力加成無關)。
// - AI Ground Troops Bonus(MANUAL_150.html:「During ground invasion, the AI troops
//   bonus/penalty was listed but not added to the sum... this bonus/penalty did already
//   apply to the actual combat resolve.」)只確認該加成存在且已生效,未列出依難度分級的精確
//   數字。
// - Orbital Bombardment 的「A better computer helps for beams here too」與「Damage of beams
//   and torpedoes is halved just like in tactical combat」沒有給出轟炸情境下的獨立數字,兩者
//   應沿用一般戰術戰鬥的電腦加成表(見 formulas.go computerBonusTable)與既有光束/魚雷減傷
//   規則,故不在本檔重複定義。
// - Stored Production 越高 hit 機率越高的精確曲線(「larger stored prod increases its hit
//   chance」)手冊沒有給出公式,僅以 GroundPlanetHitsPerStoredProduction 保留「有觸發此條件」
//   這件事,不臆測遞增規則。
