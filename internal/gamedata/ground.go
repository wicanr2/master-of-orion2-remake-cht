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

// ✅ **2026-08-07:整張表從原版讀出來了,而且手冊的五項逐項吻合**(gap report 第 35 項(三張查表))。
//
// `Player_Best_Armor_` @ 0xDC323 走訪的表在 `word_17F63E`(每列 15 位元組:+0 科技 id、
// +3 加成),六列的科技 id 與 remake 的 `Technology` 列舉**完全對得上**:
//
//	187 TECH_TITANIUM_ARMOR    +5     ← 手冊沒列
//	191 TECH_TRITANIUM_ARMOR   +10    ← 手冊:adds 10
//	203 TECH_ZORTRIUM_ARMOR    +15    ← 手冊:adds 15
//	117 TECH_NEUTRONIUM_ARMOR  +20    ← 手冊:adds 20
//	  2 TECH_ADAMANTIUM_ARMOR  +25    ← 手冊:adds 25
//	201 TECH_XENTRONIUM_ARMOR  +30    ← 手冊:+30
//
// 上面五項與手冊逐字相同——**這正是「這張表就是它」的證明**,而第六項(鈦裝甲 +5)
// 是手冊沒寫的。鈦裝甲是開局就有的基礎裝甲,所以那 5 點是每個帝國、每一場地面戰都少的。
//
// 原版取的是「已知的裡面最高階那一項」(從表尾往前找第一個已知的),不是加總。
//
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
	case TECH_TITANIUM_ARMOR:
		// ⚠ 手冊**沒有**列這一項,是 2026-08-07 從原版的表讀出來的(見上方說明)。
		// 鈦裝甲是開局就有的基礎裝甲,所以先前少的這 5 點是**每個帝國、每一場地面戰**都少的。
		return 5
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

// --- 步槍科技(`Player_Best_Rifle_` @ 0xDC416)---
//
// ⚠ **這一整條通道 remake 先前完全沒有**(gap report 第 35 項(三張查表))。
// 表在 `word_14A88`(每列 3 位元組:+0 科技 id、+2 加成),五列的科技 id 同樣完全對得上:
//
//	145 TECH_PULSE_RIFLE    +0     ← 開局就有的基礎步槍
//	101 TECH_LASER_RIFLE    +5
//	 73 TECH_FUSION_RIFLE   +10
//	128 TECH_PHASOR_RIFLE   +20
//	138 TECH_PLASMA_RIFLE   +30
//
// 十二個科技 id(裝甲 6 + 步槍 5 + 個人護盾 1)**全部**對上 remake 的列舉,
// 而且裝甲那六項的加成與手冊逐字相同——不是巧合。
//
// 上限差 **30 點**:後期科技全開的帝國,remake 先前的地面部隊比原版弱整整 30。

// GroundRifleTechBonus 依步槍科技回傳地面戰力加成(原版 `word_14A88` 的表)。
//
// 同裝甲:原版取「已知的裡面最高階那一項」,不是加總。
func GroundRifleTechBonus(tech Technology) int {
	switch tech {
	case TECH_PULSE_RIFLE:
		return 0
	case TECH_LASER_RIFLE:
		return 5
	case TECH_FUSION_RIFLE:
		return 10
	case TECH_PHASOR_RIFLE:
		return 20
	case TECH_PLASMA_RIFLE:
		return 30
	}
	return 0
}

// GroundRifleLadder 是步槍科技由低到高的順序(供「取最高階已知」用)。
func GroundRifleLadder() []Technology {
	return []Technology{
		TECH_PULSE_RIFLE, TECH_LASER_RIFLE, TECH_FUSION_RIFLE,
		TECH_PHASOR_RIFLE, TECH_PLASMA_RIFLE,
	}
}

// GroundBattleoidCombatBonus 手冊 p.81(Battleoids):「Battleoids have a ground combat
// rating 10 higher than a tank」。此為相對 Tank 的加成,非獨立疊加項。
//
// ✅ 2026-08-07 反組譯確認,並補上手冊沒說的部分:`Compute_Player_Ground_Combat_Bonuses_`
// 對 `TECH_BATTLEOIDS`(id 24,`[player+0x12F]`)寫 `[out+1] = 10` 與 `[out+2] = 1`,
// 而那兩欄只被**類型 0(裝甲)**的 case 讀走。也就是 Battleoids 是
// **裝甲專屬的 +10 攻擊 + 1 耐受**,不是給整支部隊的。
const GroundBattleoidCombatBonus = 10

// GroundBattleoidExtraHits 是 Battleoids 給裝甲部隊的額外耐受命中數(原版 `[out+2] = 1`)。
// 手冊只提了戰力 +10,沒提這一項。
const GroundBattleoidExtraHits = 1

// GroundPoweredArmorAppliesTo 記錄一件手冊沒說清楚的事:
// `TECH_POWERED_ARMOR`(id 144,`[player+0x1A7]`)寫的是 `[out+3] = 10` / `[out+4] = 1`,
// 而那兩欄只被**類型 1(陸戰隊)**的 case 讀走——動力裝甲是**陸戰隊專屬**的。
// 對照之下 `TECH_ANTIGRAV_HARNESS`(id 9,`[player+0x120]`)寫的是 `[out+0]`,
// 那一欄進的是**所有類型共用**的基礎(`var_C`)。
const GroundPoweredArmorAppliesTo = GroundTypeMarines

// --- 種族地面戰加成:改由 gamedata.OrigRaceTrait(race, TRAIT_GROUND_COMBAT) 提供 ---
//
// 這裡原本有一個 `GroundRace` 列舉與 `GroundRaceCombatBonus`,把「手冊有明確數字的種族」
// 硬編成三格(Bulrathi +10 / Gnolam −10 / 其他 0)。2026-08-08(第 65 項(種族特性31格))拆掉,原因有二:
//
//	① **一手表出現了。** `RACESTUF.LBX` asset 7 有全部 13 族的特性陣列,
//	   不必再靠「手冊有沒有寫數字」來決定哪幾族查得到。
//	② **Gnolam 的 −10 是重複扣的。** 它的 `TRAIT_GROUND_COMBAT` 是 **0**;
//	   那 −10 完全來自 `TRAIT_LOW_G`,而呼叫端本來就會再套一次 GroundApplyLowGPenalty。
//	   兩條路各扣一次 → 諾蘭姆的地面戰力被多扣了 10 點。
//
// **不要把那個列舉加回來。** 需要某族的地面戰加成就查特性表。

// --- 重力種族特性(手冊 p.24 + `Compute_Player_Ground_Combat_Bonuses_` @ 0xEC15C)---
//
// 手冊明說這兩個**互斥**(「High-G World and Low-G World are mutually exclusive」),
// 而反組譯把互斥直接寫成 `if / else if`:
//
//	cmp byte ptr [player+8AAh], 0     ; High-G
//	jz  short loc_EC227
//	mov byte ptr [out+0Ch], 1         ; → 耐受命中數 +1
//	jmp short loc_EC234
//	loc_EC227:
//	cmp byte ptr [player+8A9h], 0     ; Low-G(只有在不是 High-G 時才看)
//	jz  short loc_EC234
//	mov byte ptr [out+0Dh], 0F6h      ; → 攻擊力 −10(0xF6 = −10 的有號位元組)
//
// **那個 else 本身就是互斥的證據**,而互斥又是手冊明寫的——兩邊互證。

// GroundHighGExtraHits 是 High-G 種族的地面部隊多挨幾下才死。
//
// 手冊逐字:「High-G ground troops can sustain substantially more physical damage than
// other troops; **they take 1 hit more than normal troops before being slain in ground
// combat.**」——與反組譯的 `mov byte ptr [out+0Ch], 1` 逐字對上
// (基礎耐受是 `[out+0x0C] + 1`,見 ground_battle_orig.go 的 `GroundBaseHitsToKill`)。
const GroundHighGExtraHits = 1

// GroundLowGCombatPenalty 是 Low-G 種族的地面戰力懲罰。
//
// ⚠ **這裡與手冊的字面讀法不同,而且是刻意的。** 手冊寫「Low-G troops suffer a **10%**
// penalty during ground combat」,remake 先前照字面做成乘法(戰力 × 90%)並在註解裡寫著
// 「手冊未列出 10% 套用在哪個基準值、如何捨入」。
//
// 2026-08-07 反組譯給了答案:`mov byte ptr [ecx+0Dh], 0F6h` —— **0xF6 是有號位元組的 −10,
// 是個定值,不是百分比**。它與其他所有加成一起加進攻擊力(`Compute_Ground_Combat_Info_`
// 的 `var_4`),而那些加成本身也都是 +10/+15/+20 這種定值。
//
// 手冊那個「%」多半是行文上的隨手寫法(其他加成手冊也是寫「adds 10 to…」而不是 10%)。
// 一手事實優先。
const GroundLowGCombatPenalty = -10

// GroundApplyLowGPenalty 對地面戰力套用 Low-G 種族的懲罰。
//
// ⚠ 2026-08-07 由「×90%」改成 **−10 定值**(見 GroundLowGCombatPenalty)。
// 差異在典型戰力(30..60)下是 3..6 點 vs 固定 10 點,不是可以忽略的捨入差。
func GroundApplyLowGPenalty(strength int) int {
	return strength + GroundLowGCombatPenalty
}

// GroundSubterraneanDefenseBonus 手冊 p.24(Subterranean):「subterranean troops receive a
// +10 ground combat bonus when defending their colonies.」僅在防守己方殖民地時生效,
// 進攻時不適用(手冊未提供攻擊情境下的數字)。
//
// ✅ 2026-08-07 反組譯**獨立確認**:`Compute_Player_Ground_Combat_Bonuses_` @ 0xEC15C
//
//	cmp byte ptr [player+8ACh], 0     ; Subterranean
//	jz  short loc_EC247
//	cmp [ebp+var_4], 0                ; ← 呼叫端傳的旗標;殖民地(守方)傳 1
//	jz  short loc_EC247
//	mov byte ptr [out+0Eh], 0Ah       ; → +10
//
// 「只有守方才給」在原版是那個由呼叫端傳進來的旗標,而 `Compute_Colony_Ground_Combat_Info_`
// (殖民地 = 守方)傳的正是 1。數字與條件兩邊都對上——這一條從「手冊單一來源」升級為雙來源。
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
//   same fleet as the Ship Officer.」):2026-07-11 已用 PARAMETERS.CFG:2745-2753 逐字數字
//   (攻方 5x/7.5x、守方 2x/3x,守方 1.5 起再追平攻方套用 2.5x)近似補實作,見
//   ground_version_diff.go(GroundCommandoAttackerForceBonus/GroundCommandoDefenderForceBonus)
//   + ruleprofile.go(RuleProfile.DefenderCommandoBonus)。⚠ 誠實範圍:「regular commando
//   bonus」基準值本身(2/3)手冊只以相對倍率描述,本專案直接當成最終加成點數,屬近似而非手冊
//   獨立驗證值;shell 層(internal/shell/ground_invasion.go)只接了攻方(玩家 Leaders 有資料),
//   守方(AI 無 Leaders 模型)仍是掛鉤未接,詳見該檔案與 RuleProfile 欄位註解。
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
