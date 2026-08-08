package shell

// combat_formula.go:用 gamedata 的真 MOO2 戰鬥公式解算單次射擊,取代原本「攻擊−HP」抽象相減。
// 核心數學(射程懲罰→命中門檻→命中判定→傷害分布→過盾→過甲)全部呼叫 gamedata 真公式
// (逐字轉寫自 openorion2 + 手冊,有測試)。per-ship 的攻防/傷害/盾甲數值為 remake 由艦艇
// 設計推導的近似(見 StartCombat 註記;精確值需艦體空間格 + 元件佔格 + 軍官技能模型,待建)。

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// ShotResult 是一次射擊的解算結果。
type ShotResult struct {
	Hit               bool
	DamageToStructure int // 穿過盾與甲、實際打到艦體結構的傷害
	RemainingArmorHP  int // 射擊後剩餘裝甲 HP
}

// ResolveShot 用 gamedata 真公式解算一次射擊(beam 武器路徑,不含 Point Defense/PD 掛載
// 加成——目前呼叫端皆固定傳 0 PD bonus)。委派 ResolveShotWithMods(mods=nil),行為與加入
// 武器改造(mod)系統前完全相同(回歸安全:DamageMountAdjustedValue 在 hv/pd bonus 皆 0、
// rangePenaltyPoints 不變時,對任一 base>=1 恆等於 base,詳見該函式註解)。
//   - netAttack = 攻方 Beam Attack(含命中加成) − 守方防禦(AF+BD)。
//   - rangeSquares = 曼哈頓/格數距離(→射程等級→命中懲罰)。
//   - roll = 呼叫端擲出的 random(1..100)(由戰鬥 RNG 提供,保持可重現)。
//
// 流程:射程等級→射程懲罰→命中門檻→CombatClassicToHit→DamageForHit→DamageAfterShield→DamageApplyArmor。
//
// 只適用一般光束武器(WeaponKindBeam)。飛彈/球形武器的命中判定機制不同(見
// ResolveMissileShot/ResolveSphericalShot),呼叫端須先用 weaponKindByName 分流,
// 不可對飛彈/球形武器呼叫本函式。
func ResolveShot(netAttack, weaponMin, weaponMax, rangeSquares, shieldReduction, armorHP, roll int, hardShield, armorPiercing bool) ShotResult {
	var mods []gamedata.WeaponModCode
	if armorPiercing {
		// 舊呼叫端(呼叫本函式而非 WithMods 版)仍可用既有的 armorPiercing bool 表達穿甲,
		// 不強制改signature,對映成 AP mod 走同一套（DamageApplyArmor 只認 bool,不認
		// mod 清單本身,這裡只是把 bool 包成 mods 給 ResolveShotWithMods 用同一份實作)。
		mods = []gamedata.WeaponModCode{gamedata.ModArmorPiercing}
	}
	return ResolveShotWithMods(netAttack, weaponMin, weaponMax, rangeSquares, shieldReduction, armorHP, roll, hardShield, mods, 0, false)
}

// ResolveShotWithMods 同 ResolveShot,額外接受一組攻方武器改造(mods),依手冊
// (GAME_MANUAL.pdf p.115-118,見 gamedata/weapon_mods.go)套用其對命中/傷害的效果:
//   - CO(+25)/AF(-20) 加減 netAttack(BA+CO-AF-BD,見 gamedata.WeaponModNetAttackBonus)。
//   - PD 額外提供 +25 命中門檻加成(gamedata.WeaponModPDBonus,對應 CombatHitThreshold 的
//     pdBonus 參數)。
//   - HV/PD 選用對應的射程等級表(halved/doubled,gamedata.CombatRangeLevelForBeamMods),
//     再由 gamedata.WeaponModDamageBonuses 算出 hvBonus/pdPenalty 餵給既有的
//     DamageMountAdjustedValue,調整武器傷害潛力(150%/50%)。
//   - ENV 對命中後的傷害 *4(gamedata.WeaponModEnvelopingMultiply)。
//   - AP/SP 分別對應 DamageApplyArmor/DamageAfterShield 既有的 armorPiercing/
//     shieldPiercing 參數。
//
// mods 為 nil 時(無改造)本函式對任何輸入的行為與加入 mod 系統前的 ResolveShot 完全相同
// (中性回歸,見 combat_formula_test.go 的 no-mod 回歸測試)。
// hefBonus 是高能聚焦(High Energy Focus)給的傷害百分點加成:裝了傳
// gamedata.DamageMountBonusHEF(50),沒裝傳 0。用 hefBonusFor 取值,不要自己寫 50。
//
// apNegated 是**目標**讓穿甲(AP)失效(氙素裝甲 或 重裝甲系統,手冊各一句)。
// 用 shipNegatesArmorPiercing 判斷,它把兩條路併起來。
func ResolveShotWithMods(netAttackBase, weaponMin, weaponMax, rangeSquares, shieldReduction, armorHP, roll int, hardShield bool, mods []gamedata.WeaponModCode, hefBonus int, apNegated bool) ShotResult {
	return ResolveBeamShot(BeamShot{
		NetAttack: netAttackBase, WeaponMin: weaponMin, WeaponMax: weaponMax,
		RangeSquares: rangeSquares, Roll: roll, Mods: mods,
		Attacker: BeamAttackerSystems{HEFBonus: hefBonus},
		Target: BeamTargetSystems{
			ShieldReduction: shieldReduction, ArmorHP: armorHP,
			HardShield: hardShield, APNegated: apNegated,
		},
	})
}

// BeamShot 是一次光束射擊的全部輸入。
//
// ============ 為什麼把參數收成結構 ============
//
// `ResolveShotWithMods` 的位置參數在第 67/68 項之後排到了 **11 個**,而第 69 項盤點出來
// 還有兩個手冊系統要接進同一條鏈(結構分析儀:過盾傷害加倍;阿基里斯瞄準器:光束一律
// 無視裝甲)。第 69 項當時的判斷是:
//
//	再加下去該先把攻方/守方系統各收成一個結構——**那是重構,不該夾在資料項裡做**
//
// 這一項就是那個重構。收完之後:
//
//   - **呼叫端讀得出意思**。`false, nil, 0, false` 這種尾巴誰也看不出哪個是什麼。
//   - **加一個系統只動一個地方**,不必再回頭改每一個呼叫端與每一個測試。
//   - **攻方/守方分開**,不會再出現「把守方的旗標填進攻方那一格」這種安靜的錯。
//
// 舊入口 `ResolveShot` / `ResolveShotWithMods` 保留為薄包裝:既有呼叫端與測試一行都不必改,
// 行為逐位元相同(新欄位留零值 = 原本的固定值)。
type BeamShot struct {
	NetAttack    int // 攻方 Beam Attack(含命中加成)− 守方防禦(AF+BD)
	WeaponMin    int // 單發最小傷害
	WeaponMax    int // 單發最大傷害
	RangeSquares int // 格數距離(→射程等級→命中懲罰)
	Roll         int // 呼叫端擲出的 1..100(由戰鬥 RNG 提供,保持可重現)
	Mods         []gamedata.WeaponModCode
	Attacker     BeamAttackerSystems
	Target       BeamTargetSystems
}

// BeamAttackerSystems 是**攻擊方**艦上會影響這一發的系統。
type BeamAttackerSystems struct {
	// HEFBonus 是高能聚焦的傷害百分點加成(裝了 = gamedata.DamageMountBonusHEF)。
	HEFBonus int
	// StructuralAnalyzer 是結構分析儀。手冊逐字:「the damage done by beam weapons that
	// penetrate an enemy ship's shields is **doubled**.」——**過盾之後**才加倍,
	// 所以它與護盾減傷的先後順序有意義:先扣盾、再加倍。
	StructuralAnalyzer bool
	// AchillesUnit 是阿基里斯瞄準器。手冊逐字:「all beam weapons **ignore the target's
	// armor completely**.」——等於不必掛 AP 改造就有 AP 的效果。
	//
	// ⚠ **它會不會被重裝甲/氙素裝甲抵銷,手冊沒有明說。** 那兩條寫的是
	// 「negates the Armor **Piercing abilities** of enemy weapons」——字面涵蓋
	// 「無視裝甲」這個能力,所以這裡採「會被抵銷」的讀法。標在這裡,不假裝手冊講了。
	AchillesUnit bool
}

// BeamTargetSystems 是**目標方**艦上會影響這一發的系統與狀態。
type BeamTargetSystems struct {
	ShieldReduction int  // 護盾每發減傷
	ArmorHP         int  // 目前裝甲 HP
	HardShield      bool // 硬化護盾(額外減傷)
	// APNegated 是這艘船讓穿甲失效(氙素裝甲 或 重裝甲系統)。用 shipNegatesArmorPiercing
	// 判斷,它把兩條路併起來。
	APNegated bool
}

// ResolveBeamShot 是光束射擊的單一實作入口(其餘 Resolve* 都委派到這裡)。
func ResolveBeamShot(in BeamShot) ShotResult {
	netAttackBase, weaponMin, weaponMax := in.NetAttack, in.WeaponMin, in.WeaponMax
	rangeSquares, roll, mods := in.RangeSquares, in.Roll, in.Mods
	shieldReduction, armorHP := in.Target.ShieldReduction, in.Target.ArmorHP
	hardShield, apNegated := in.Target.HardShield, in.Target.APNegated
	hefBonus := in.Attacker.HEFBonus
	netAttack := netAttackBase + gamedata.WeaponModNetAttackBonus(mods)
	level := gamedata.CombatRangeLevelForBeamMods(rangeSquares, mods)
	penalty := gamedata.CombatRangeLevelPenalty(level)
	pdBonus := gamedata.WeaponModPDBonus(mods)
	threshold := gamedata.CombatHitThreshold(penalty, pdBonus)

	if !gamedata.CombatClassicToHit(roll, netAttack, threshold) {
		return ShotResult{Hit: false, RemainingArmorHP: armorHP}
	}

	// 距離傷害衰減(dissipation):命中之後,傷害還要依 range level 打折。
	//
	// 對應原版 Get_Beam_Weapon_Modifiers_(sub_39434):
	//     if (!(mods & 0x20) && !(weaponFlags & 0x04))
	//         *damagePct += ranged_damage_penalty[range];
	// 即「除非帶 NR 改造、或武器天生免疫,否則一律套」。remake 先前完全沒接這一段
	// (weapon_mods.go 的 ModNoRangeDissipation 註解記著這個 TODO),遠距離射擊的傷害
	// 與貼臉一樣——NR 這個改造因此也一直沒有可觀察的效果。
	//
	// gamedata.damageDissipationPenaltyTable = {0,0,10,20,30,40,50,60,65} 與原版
	// `_ranged_damage_penalty`(0x17D867)逐格相同,表本身早就對,缺的只是這裡的接線。
	dissipation := 0
	if !gamedata.WeaponModNoRangeDissipation(mods) {
		dissipation = gamedata.DamageDissipationPenalty(level)
	}

	// [回歸保護] DamageMountAdjustedValue 對「命中後傷害潛力恆為 1」有夾限(手冊「minimum
	// damage potential is always 1」),對 base=0 的「無武裝」武器會把 0 誤夾成 1。
	// 因此 weaponMax<=0 一律跳過調整,避免「無武裝艦艇突然打出 1 點傷害」這種回歸。
	adjMin, adjMax := weaponMin, weaponMax
	if weaponMax > 0 && (len(mods) > 0 || dissipation > 0 || hefBonus > 0) {
		hvBonus, pdPenalty := gamedata.WeaponModDamageBonuses(mods)
		adjMin = gamedata.DamageMountAdjustedValue(weaponMin, hvBonus, hefBonus, pdPenalty, dissipation)
		adjMax = gamedata.DamageMountAdjustedValue(weaponMax, hvBonus, hefBonus, pdPenalty, dissipation)
	}

	dmg := gamedata.DamageForHit(adjMin, adjMax, roll, netAttack, threshold)
	dmg = gamedata.WeaponModEnvelopingMultiply(dmg, mods)
	dmg = gamedata.DamageAfterShield(dmg, shieldReduction, hardShield, gamedata.WeaponModShieldPiercing(mods))
	// 結構分析儀:手冊「the damage done by beam weapons that **penetrate an enemy ship's
	// shields** is doubled」——「penetrate the shields」是條件也是時機,所以加倍發生在
	// 扣完護盾**之後**。順序寫反的話護盾也會跟著被加倍的傷害穿透,那是另一種規則。
	if in.Attacker.StructuralAnalyzer {
		dmg *= gamedata.ShipStructuralAnalyzerMultiplier
	}
	// 阿基里斯瞄準器:手冊「all beam weapons ignore the target's armor completely」
	// ——與 AP 改造走同一個開關(見 BeamAttackerSystems.AchillesUnit 對「會不會被抵銷」的說明)。
	armorPiercing := gamedata.WeaponModArmorPiercing(mods) || in.Attacker.AchillesUnit
	_, toStruct, remArmor := gamedata.DamageApplyArmor(dmg, armorHP, armorPiercing, apNegated)
	return ShotResult{Hit: true, DamageToStructure: toStruct, RemainingArmorHP: remArmor}
}

// ResolveMissileShot 用 gamedata missile.go 已移植公式,解算一發飛彈/魚雷攻擊,對應手冊
// 「Notes on Missile Defenses > Missile Evasion」(p123)+「Notes on Anti-Missile Rockets」
// (p125)。與光束不同,飛彈不是用 Beam Attack/Beam Defense/Range Penalty 的命中門檻公式
// 判定命中,而是①可能先被 AMR 攔截、②再由 Jam Chance 判定目標是否閃避成功——這是兩個獨立
// 事件,呼叫端須各擲一顆獨立的 1-100(amrRoll/jamRoll),不可共用同一個 roll(beam 那套
// 「同一個 roll 同時決定命中與傷害內插」的手法是 beam 公式本身的設計,不適用於飛彈)。
//
// 參數對照手冊/missile.go 出處:
//   - hasAMR/amrRangeSquares:目標艦是否裝有反飛彈火箭(Anti-Missile Rockets)、與其
//     距離(格,→ gamedata.MissileAMRRangeIndex →命中率)。
//     ⚠ 2026-08-08:這裡原本寫著「現行 remake 的 SpecialOptions 尚未提供『反飛彈火箭』
//     這個可造艦元件,呼叫端目前一律傳 hasAMR=false」——**第 65 項已經把該元件補上了**,
//     呼叫端也改成依目標艦是否裝載決定。註解比程式碼晚了三項才更新。
//   - defenderEvasionBonus:目標的飛彈閃避加成加總(ECM Jammer/Stabilizer/種族/艦員/
//     統帥,各項手冊固定數值見 missile.go 的 MissileJammer*/MissileInertialStabilizer/
//     MissileInertialNullifier/MissileShipDefenseRacialBonus/MissileCrew*/
//     MissileHelmsmanEvasionBonus)。現行 remake 的艦艇設計/軍官系統尚未提供這些元件,
//     呼叫端目前一律傳 0(TODO,待補上後從實際裝載/軍官推導)。
//   - attackerScannerBonus:⚠ 2026-08-08(第 72 項)訂正。這裡原本寫著「現行 remake 未提供
//     攻方掃描器(Scanner)…呼叫端一律傳 0(TODO)」——**那句話從一開始就不對**。手冊指的
//     「best known scanner bonus of the attacker」是**掃描科技**(迅子 −20、中子 −40,各自寫在
//     自己的條目裡),不是某個可造艦元件;而那三個掃描科技從 detection.go 建起來那天就在
//     remake 裡了(bestPlayerScannerParsec 讀的正是它們)。缺的不是前置系統,是沒有人把
//     同一個科技的第二個效果查出來。現在走 bestPlayerScannerJamReduction。
//   - hasECCM:仍為 false。ECCM 是**飛彈的改造**(手冊 Weapon Mods 附錄),而 remake 的
//     mod 層只對光束武器開放(見 weapon_mods.go 檔頭)——這一條是真的擋在前置系統後面,
//     不是漏查。
//     以上四項在「無任何裝備」時退化為手冊「若目標無任何閃避能力,預設100%命中」
//     (gamedata.MissileDefaultHitChance)——這是手冊本身的基準情境,不是臆造值,恰好與
//     現行武器/元件表(尚無任何閃避裝備)的現況一致。
//   - weaponMax:飛彈命中後的傷害。手冊只列固定「listed」傷害值(如「Nuclear Missile
//     Damage lowered from 8 to 6」),沒有給出像 beam 命中裕度那樣的內插公式,故不套用
//     beam 專用的 gamedata.DamageForHit(那需要 net-attack/hit-threshold,是命中判定
//     機制不同的 beam 概念,套用會混淆兩種機制);仍依手冊預設(只有掛 Shield
//     Piercing/Armor Piercing mod 才豁免,本 remake 尚未對飛彈掛任何 mod)穿過護盾/裝甲。
//
// MissileDefenses 是目標艦的「特殊防禦裝置」與它們各自要用的骰(手冊 p.123)。
//
// **裝了才擲骰**:呼叫端在沒裝的情況下不該動 RNG,否則整條亂數流會位移,
// 既有存檔與探針的戰鬥結果全部改變(見 battleVolley 的炸彈分支同款處理)。
// 沒裝時把對應的 Has* 留 false、骰值留 0 即可,本函式不會讀它。
type MissileDefenses struct {
	HasLightningField bool // 閃電場:每一枚來襲飛彈/魚雷/戰機各 50% 直接摧毀
	LightningRoll     int  // 1..100
	HasDisplacement   bool // 位移裝置:一律 30% 完全未命中
	DisplacementRoll  int  // 1..100
}

func ResolveMissileShot(
	hasAMR bool, amrRangeSquares, amrRoll int,
	defenderEvasionBonus, attackerScannerBonus int, hasECCM bool, jamRoll int,
	weaponMax, shieldReduction, armorHP int, hardShield bool,
	def MissileDefenses,
) ShotResult {
	// 閃電場在**最前面**:手冊說它「對每一枚試圖命中的飛彈」判定,而且明寫是在
	// MIRV 分裂彈頭「之前」——那個順序表示它擋的是整枚飛彈,不是彈頭。
	if def.HasLightningField && def.LightningRoll <= gamedata.MissileLightningFieldDestroyChance {
		return ShotResult{Hit: false, RemainingArmorHP: armorHP}
	}
	if hasAMR && amrRangeSquares <= gamedata.MissileAMRMaxRangeSquares {
		if amrRoll <= gamedata.MissileAMRChanceToHit(gamedata.MissileAMRRangeIndex(amrRangeSquares)) {
			return ShotResult{Hit: false, RemainingArmorHP: armorHP} // 被 AMR 擊落
		}
	}

	jamChance := gamedata.MissileJamChance(defenderEvasionBonus, attackerScannerBonus, hasECCM)
	hitChance := gamedata.MissileDefaultHitChance - jamChance
	if hitChance > 100 {
		hitChance = 100
	}
	if hitChance < 0 {
		hitChance = 0
	}
	if jamRoll > hitChance {
		return ShotResult{Hit: false, RemainingArmorHP: armorHP} // 被干擾/閃避
	}
	// 位移裝置在**最後面**:手冊說它「不論其他裝備或情況」一律 30% 完全未命中,
	// 而且與匿蹤同組、明寫是在 MIRV 分裂「之後」判定——那個順序表示它躲的是彈頭,
	// 所以排在閃避判定之後、傷害結算之前。
	if def.HasDisplacement && def.DisplacementRoll <= gamedata.MissileDisplacementDeviceMissChance {
		return ShotResult{Hit: false, RemainingArmorHP: armorHP}
	}

	dmg := gamedata.DamageAfterShield(weaponMax, shieldReduction, hardShield, false)
	_, toStruct, remArmor := gamedata.DamageApplyArmor(dmg, armorHP, false, false)
	return ShotResult{Hit: true, DamageToStructure: toStruct, RemainingArmorHP: remArmor}
}

// ResolveSphericalShot 用 gamedata damage.go 已移植的球形武器(Pulsar/Plasma
// Flux/Spatial Compressor 等)公式解算一次對「艦艇」目標的球形武器齊射。手冊
// (「Notes on Spherical Damage」p126)強調「sphericals always use all weapons from the
// slot」,故 aggD 應是呼叫端已用 gamedata.DamageSphericalRoll 對同一 slot 全部武器逐發
// 算好、加總後的值,不是逐發個別解算(這點與 beam/missile 逐發判定不同)。
//
// 手冊「Damage Calculation > Ships」流程:aggD 算好後,還要再做「the number of rolls is
// determined by size class + 1」次 random(aggD)、「each re-rolled if the outcome is not
// 1」加總才是最終傷害——這個重骰終止條件手冊描述不足以還原成確定性演算法,damage.go 本身
// 已明載不移植(見 gamedata.DamageSphericalShipRollCount 的函式註解)。故本函式保守地
// 直接以 aggD 當作對艦傷害(不臆造重骰後的加總值),之後若要精確還原重骰機制,需先查證
// 終止條件(如比對 openorion2 原始碼或實機錄影),詳見
// docs/tech/tactical-combat-weapon-kinds.md 的 TODO。
//
// 與一般光束相同,穿過護盾/裝甲——手冊只有 Spatial Compressor 明講「does all damage to
// structure only, ignoring shields and armor」,其餘球形武器(Pulsar/Plasma Flux)未講
// 豁免,故用 bypassShieldAndArmor 供 Spatial-Compressor 類武器啟用該豁免。
// 手冊「minimum damage of 1 against ships」,aggD 不足 1 時夾為 1。
//
// 現行 WeaponOptions(session.go)沒有任何武器分類到 WeaponKindSpherical(見
// weapon_kind.go 的核對說明),此函式目前無實際呼叫路徑會用到,只是先備好、有測試的解算
// 函式,供未來新增球形武器元件時串接。
func ResolveSphericalShot(aggD, shieldReduction, armorHP int, hardShield, bypassShieldAndArmor bool) ShotResult {
	if aggD < 1 {
		aggD = 1
	}
	if bypassShieldAndArmor {
		return ShotResult{Hit: true, DamageToStructure: aggD, RemainingArmorHP: armorHP}
	}
	dmg := gamedata.DamageAfterShield(aggD, shieldReduction, hardShield, false)
	_, toStruct, remArmor := gamedata.DamageApplyArmor(dmg, armorHP, false, false)
	return ShotResult{Hit: true, DamageToStructure: toStruct, RemainingArmorHP: remArmor}
}
