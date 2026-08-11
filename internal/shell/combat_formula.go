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
	ShieldDamage      int // 命中後一般護盾實際吸收的傷害，用來扣除命中分面容量
}

// PointDefenseShot 是一發點防禦光束對來襲飛彈的解算輸入。
//
// 這是 remake 的共用垂直切片：原版 raw `sub_3AD57` @ 0x3AD57 →
// `Weapon_In_Range` @ 0x3A0B9 → `Missile_Dcv` @ 0x3E095 的攔截鏈已被拆出，
// 但原版艦艇 AI、每個 PD 槽的完整佈局仍未完全還原。`BeamAttack`／`BeamDamageMax`
// 沿用目前 shell 的艦艇戰力近似，raw kind／flag 則保留原版定位。
type PointDefenseShot struct {
	BeamWeaponName            string
	BeamAttack                int
	BeamDamageMax             int
	BeamRangeSquares          int
	BeamRoll                  int
	BeamSystems               BeamAttackerSystems
	BeamMods                  []gamedata.WeaponModCode
	MissileWeaponName         string
	MissileFTLLevel           int
	MissileMods               []gamedata.WeaponModCode
	CarriedInterceptionDamage int
}

// PointDefenseResult 是點防禦攔截的可觀測結果。DestroyedWarheads 只表示攔截鏈
// 摧毀的彈頭數，不是直接打到目標艦體的傷害。
type PointDefenseResult struct {
	Fired                       bool
	Hit                         bool
	DestroyedWarheads           int
	RemainingInterceptionDamage int
	MissileBeamDefense          int
	InterceptionDurability      int
}

// PointDefenseFighterShot 是 PD 在戰機接戰前的單發輸入。
//
// 手冊 p.117 的順序是「飛彈命中艦艇或戰機接戰前,PD 若尚未開火就先開火」；
// 戰機的 Beam Defense 則用 p.157 的獨立公式計算。這條路徑不把戰機誤當成
// 飛彈彈頭，因此不套用 Missile_Dcv 的 quotient/remainder 消費；命中後傷害
// 由呼叫端送入 FighterSquadron.TakeHit。
type PointDefenseFighterShot struct {
	BeamWeaponName     string
	BeamAttack         int
	BeamDamageMax      int
	BeamRangeSquares   int
	BeamRoll           int
	BeamSystems        BeamAttackerSystems
	BeamMods           []gamedata.WeaponModCode
	FighterBeamDefense int
}

// PointDefenseFighterResult 是 PD 對戰機的一發結果。
type PointDefenseFighterResult struct {
	Fired              bool
	Hit                bool
	DamageToFighter    int
	FighterBeamDefense int
}

// PointDefenseCanFire 回報一個武器槽是否具備 PD 自動開火資格。
// 與來襲物分開抽出，讓飛彈與戰機共用「尚未開火」的觸發條件。
func PointDefenseCanFire(beamWeaponName string, beamMods []gamedata.WeaponModCode) bool {
	return WeaponIsBeam(beamWeaponName) && gamedata.WeaponModHas(beamMods, gamedata.ModPointDefense)
}

// PointDefenseCanEngage 回報目前資料模型是否能讓 PD 對指定來襲物開火。魚雷與
// 尚未對到標準玩家飛彈 raw kind 的武器刻意排除；手冊明說魚雷不能被一般武器瞄準。
func PointDefenseCanEngage(beamWeaponName, missileWeaponName string, beamMods []gamedata.WeaponModCode) bool {
	if !PointDefenseCanFire(beamWeaponName, beamMods) {
		return false
	}
	if WeaponIsTorpedo(missileWeaponName) {
		return false
	}
	_, ok := MissileRawWeaponKind(missileWeaponName)
	return ok
}

// ResolvePointDefenseFighterShot 解算 PD 在戰機接戰前的自動射擊。
//
// `FighterBeamDefense` 必須由呼叫端提供；目前戰術呼叫端已把參戰艦隊的種族
// Ship Defense 與 Fighter Pilot 加成帶入，Helmsman 仍明示傳 0，因為原版 1.50
// 對戰機接戰的完整呼叫端尚未追回。公式本身由 gamedata.CombatFighterBeamDefense
// 保留，未來補齊 Helmsman 時不需要改動 PD 消費端。
func ResolvePointDefenseFighterShot(in PointDefenseFighterShot) PointDefenseFighterResult {
	if !PointDefenseCanFire(in.BeamWeaponName, in.BeamMods) {
		return PointDefenseFighterResult{}
	}
	result := PointDefenseFighterResult{
		Fired:              true,
		FighterBeamDefense: in.FighterBeamDefense,
	}
	level := gamedata.CombatRangeLevelForBeamMods(in.BeamRangeSquares, in.BeamMods)
	toHitPenalty := gamedata.CombatRangeLevelPenalty(level)
	netAttack := in.BeamAttack + gamedata.WeaponModNetAttackBonus(in.BeamMods) - in.FighterBeamDefense
	hitThreshold := gamedata.CombatHitThreshold(toHitPenalty, gamedata.WeaponModPDBonus(in.BeamMods))
	if !gamedata.CombatClassicToHit(in.BeamRoll, netAttack, hitThreshold) {
		return result
	}

	hvBonus, pdPenalty := gamedata.WeaponModDamageBonuses(in.BeamMods)
	dissipation := 0
	if !gamedata.WeaponModNoRangeDissipation(in.BeamMods) {
		dissipation = gamedata.DamageDissipationPenalty(level)
	}
	damage := in.BeamDamageMax
	if damage > 0 {
		damage = gamedata.DamageMountAdjustedValue(damage, hvBonus,
			in.BeamSystems.HEFBonus, pdPenalty, dissipation)
	}
	result.Hit = true
	result.DamageToFighter = damage
	return result
}

// ResolvePointDefenseIntercept 解算 PD 的「同格自動攔截」：依 Beam Defense 做
// Classic to-hit，再以目前 PD 光束傷害潛力餵入原版 Dcv 的 quotient/remainder 鏈。
// 精確的原版攔截器射擊順序與 AI 仍標為未完成；本函式不把這個垂直切片宣稱成完整
// 戰術等價。
func ResolvePointDefenseIntercept(in PointDefenseShot) PointDefenseResult {
	if !PointDefenseCanEngage(in.BeamWeaponName, in.MissileWeaponName, in.BeamMods) {
		return PointDefenseResult{}
	}
	rawKind, ok := MissileRawWeaponKind(in.MissileWeaponName)
	if !ok {
		return PointDefenseResult{}
	}
	ftlLevel := in.MissileFTLLevel
	if ftlLevel <= 0 {
		ftlLevel = 1
	}
	rawFlags := gamedata.MissileRawFlagsForMods(in.MissileMods)
	missileDefense, ok := gamedata.MissileBeamDefenseOf(rawKind, ftlLevel, rawFlags, false)
	if !ok {
		return PointDefenseResult{}
	}
	durability := gamedata.MissileInterceptionDurabilityForRawFlags(rawKind, 0, rawFlags)
	if durability <= 0 {
		return PointDefenseResult{}
	}
	level := gamedata.CombatRangeLevelForBeamMods(in.BeamRangeSquares, in.BeamMods)
	toHitPenalty := gamedata.CombatRangeLevelPenalty(level)
	hitThreshold := gamedata.CombatHitThreshold(toHitPenalty, gamedata.WeaponModPDBonus(in.BeamMods))
	netAttack := in.BeamAttack + gamedata.WeaponModNetAttackBonus(in.BeamMods) - missileDefense
	result := PointDefenseResult{
		Fired:                  true,
		MissileBeamDefense:     missileDefense,
		InterceptionDurability: durability,
		// Weapon_In_Range 的 remainder 不是本次射擊的暫存值；攔截未命中時
		// 仍要帶回下一次，否則 ARM 飛彈的耐久鏈會被無聲清零。
		RemainingInterceptionDamage: in.CarriedInterceptionDamage,
	}
	if !gamedata.CombatClassicToHit(in.BeamRoll, netAttack, hitThreshold) {
		return result
	}
	hvBonus, pdPenalty := gamedata.WeaponModDamageBonuses(in.BeamMods)
	damageLevel := level
	damage := gamedata.DamageMountAdjustedValue(in.BeamDamageMax, hvBonus,
		in.BeamSystems.HEFBonus, pdPenalty, gamedata.DamageDissipationPenalty(damageLevel))
	destroyed, remainder := gamedata.MissileWarheadsDestroyedByInterception(
		damage, in.CarriedInterceptionDamage, durability)
	result.Hit = true
	result.DestroyedWarheads = destroyed
	result.RemainingInterceptionDamage = remainder
	return result
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
// `ResolveShotWithMods` 的位置參數在第 66/67 項之後排到了 **11 個**,而第 68 項(元件盤點+飛彈防禦)盤點出來
// 還有兩個手冊系統要接進同一條鏈(結構分析儀:過盾傷害加倍;阿基里斯瞄準器:光束一律
// 無視裝甲)。第 68 項(元件盤點+飛彈防禦)當時的判斷是:
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
	// Rangemaster 是測距瞄準器。手冊逐字:「reducing the absolute range (which is used to
	// compute accuracy and to hit penalties) to one-third of the actual range. Note that
	// the **dissipation of damage potential is not affected** by this system.」
	//
	// 那兩句話在程式裡是兩個不同的位置——射程等級要算兩次(見 ResolveBeamShot)。
	Rangemaster bool
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
	// 射程等級算**兩次**:一次給命中判定、一次給傷害衰減。
	// 沒有測距瞄準器時兩者相同(hitSquares == rangeSquares),整段與先前逐位元一致。
	hitSquares := rangeSquares
	if in.Attacker.Rangemaster {
		hitSquares = gamedata.RangemasterRangeSquares(rangeSquares)
	}
	hitLevel := gamedata.CombatRangeLevelForBeamMods(hitSquares, mods)
	level := gamedata.CombatRangeLevelForBeamMods(rangeSquares, mods)
	penalty := gamedata.CombatRangeLevelPenalty(hitLevel)
	pdBonus := gamedata.WeaponModPDBonus(mods)
	threshold := gamedata.CombatHitThreshold(penalty, pdBonus)

	if !gamedata.CombatClassicToHit(roll, netAttack, threshold) {
		return ShotResult{Hit: false, RemainingArmorHP: armorHP}
	}

	// 距離傷害衰減(dissipation):命中之後,傷害還要依 range level 打折。
	//
	// 對應原版 `Get_Beam_Weapon_Modifiers_`(raw `sub_394F7`) 呼叫
	// `Get_Beam_Range_To_Hit_Bonus_`(raw `sub_39434`):
	//     if (!(mods & 0x20) && !(weaponFlags & 0x04))
	//         *damagePct += ranged_damage_penalty[range];
	// 即「除非帶 NR 改造、或武器天生免疫,否則一律套」。這段曾是 remake 的歷史缺口；
	// 現在已接入 beam 路徑，遠距離射擊會依表扣除傷害，而 NR 會跳過該扣除。
	//
	// gamedata.damageDissipationPenaltyTable = {0,0,10,20,30,40,50,60,65} 與原版
	// `_ranged_damage_penalty`(0x17D867)逐格相同；魚雷仍沒有可證實的射程傷害模型，
	// 所以魚雷版 NR 不能由這條 beam 路徑冒充完成。
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
	shieldDamage := gamedata.DamageShieldAbsorbed(dmg, shieldReduction)
	if gamedata.WeaponModShieldPiercing(mods) {
		shieldDamage = 0
	}
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
	return ShotResult{Hit: true, DamageToStructure: toStruct, RemainingArmorHP: remArmor, ShieldDamage: shieldDamage}
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
//     這個可造艦元件,呼叫端目前一律傳 hasAMR=false」——**第 64 項(武器傷害真表)已經把該元件補上了**,
//     呼叫端也改成依目標艦是否裝載決定。註解比程式碼晚了三項才更新。
//   - defenderEvasionBonus:目標的飛彈閃避加成加總(ECM Jammer/Stabilizer/種族/艦員/
//     統帥,各項手冊固定數值見 missile.go 的 MissileJammer*/MissileInertialStabilizer/
//     MissileInertialNullifier/MissileShipDefenseRacialBonus/MissileCrew*/
//     MissileHelmsmanEvasionBonus)。現行 remake 的艦艇設計/軍官系統尚未提供這些元件,
//     呼叫端目前一律傳 0(TODO,待補上後從實際裝載/軍官推導)。
//   - attackerScannerBonus:⚠ 2026-08-08(第 71 項(探針③內部函式))訂正。這裡原本寫著「現行 remake 未提供
//     攻方掃描器(Scanner)…呼叫端一律傳 0(TODO)」——**那句話從一開始就不對**。手冊指的
//     「best known scanner bonus of the attacker」是**掃描科技**(迅子 −20、中子 −40,各自寫在
//     自己的條目裡),不是某個可造艦元件;而那三個掃描科技從 detection.go 建起來那天就在
//     remake 裡了(bestPlayerScannerParsec 讀的正是它們)。缺的不是前置系統,是沒有人把
//     同一個科技的第二個效果查出來。現在走 bestPlayerScannerJamReduction。
//   - hasECCM:舊入口仍可由呼叫端直接傳入；含改造的新入口
//     ResolveMissileShotWithMods 會從 WeaponModCodesForWeapon 自動讀 ECCM。
//     以上四項在「無任何裝備」時退化為手冊「若目標無任何閃避能力,預設100%命中」
//     (gamedata.MissileDefaultHitChance)——這是手冊本身的基準情境,不是臆造值,恰好與
//     現行武器/元件表(尚無任何閃避裝備)的現況一致。
//   - weaponMax:飛彈命中後的傷害。手冊只列固定「listed」傷害值(如「Nuclear Missile
//     Damage lowered from 8 to 6」),沒有給出像 beam 命中裕度那樣的內插公式,故不套用
//     beam 專用的 gamedata.DamageForHit(那需要 net-attack/hit-threshold,是命中判定
//     機制不同的 beam 概念,套用會混淆兩種機制);仍依手冊預設(只有掛 Shield
//     Piercing/Armor Piercing mod 才豁免；新入口會依 EMG 套用「先扣護盾、再繞過裝甲」。
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
	// CloakMissChance 是目標**此刻隱形**造成的未命中機率(隱形裝置 50%,未隱形傳 0)。
	// 手冊把它與 +80 光束防禦並列成兩種武器各自的規則,所以做成獨立的一道,
	// 而不是加進 defenderEvasionBonus——加進去會與干擾器/慣性穩定器那一族互相汙染。
	CloakMissChance int
	CloakRoll       int // 1..100(CloakMissChance > 0 時才擲)
	// MIRV 每枚彈頭各自判定；沒有這些切片時，解算器會回退使用單一舊欄位，保留舊
	// ResolveMissileShot 呼叫端的逐位元行為。
	JamRolls          []int
	CloakRolls        []int
	DisplacementRolls []int
	// InterceptedWarheads 是 PD／攔截機在進入飛彈命中判定前已摧毀的彈頭數。
	// 它由呼叫端用 ResolvePointDefenseIntercept 填入；放在這裡讓 MIRV 與既有
	// AMR／干擾／匿蹤逐彈頭流程共用同一個剩餘彈頭計數。
	InterceptedWarheads int
}

func ResolveMissileShot(
	hasAMR bool, amrRangeSquares, amrRoll int,
	defenderEvasionBonus, attackerScannerBonus int, hasECCM bool, jamRoll int,
	weaponMax, shieldReduction, armorHP int, hardShield bool,
	def MissileDefenses,
) ShotResult {
	return ResolveMissileShotWithMods(hasAMR, amrRangeSquares, amrRoll,
		defenderEvasionBonus, attackerScannerBonus, hasECCM, jamRoll,
		weaponMax, shieldReduction, armorHP, hardShield, def, "", nil)
}

// ResolveMissileShotWithMods 是含飛彈／魚雷改造的解算入口。weaponName 只用來區分魚雷專屬
// ENV/OVR；mods 應由 WeaponModCodesForWeapon 先過濾，但函式內也以類型條件再次保護，避免
// 舊存檔或測試直接傳入不適用的代碼。
//
// 已接線規則：ECCM 的干擾機率減半、EMG 在護盾後繞過裝甲、MIRV 四枚彈頭逐枚判定、
// 魚雷 ENV 四倍傷害與 OVR +50%，以及由 PD 垂直切片填入的 ARM/FST 攔截結果。
func ResolveMissileShotWithMods(
	hasAMR bool, amrRangeSquares, amrRoll int,
	defenderEvasionBonus, attackerScannerBonus int, hasECCM bool, jamRoll int,
	weaponMax, shieldReduction, armorHP int, hardShield bool,
	def MissileDefenses, weaponName string, mods []gamedata.WeaponModCode,
) ShotResult {
	torpedo := WeaponIsTorpedo(weaponName)
	if weaponKindByName(weaponName) != WeaponKindMissile {
		mods = nil
	}
	// 這裡再做一次適用性過濾：ResolveMissileShotWithMods 是 shell 對外的公式 API，不能
	// 假設所有呼叫者都經過設計畫面。
	filtered := make([]gamedata.WeaponModCode, 0, len(mods))
	for _, mod := range mods {
		if WeaponModAppliesToWeapon(weaponName, mod) {
			filtered = append(filtered, mod)
		}
	}
	mods = filtered
	warheads := gamedata.WeaponModMissileWarheadCount(mods)

	// 閃電場在**最前面**:手冊說它「對每一枚試圖命中的飛彈」判定,而且明寫是在
	// MIRV 分裂彈頭「之前」——那個順序表示它擋的是整枚飛彈,不是彈頭。
	if def.HasLightningField && def.LightningRoll <= gamedata.MissileLightningFieldDestroyChance {
		return ShotResult{Hit: false, RemainingArmorHP: armorHP}
	}
	if def.InterceptedWarheads > 0 {
		if def.InterceptedWarheads >= warheads {
			return ShotResult{Hit: false, RemainingArmorHP: armorHP}
		}
		warheads -= def.InterceptedWarheads
	}
	if hasAMR && amrRangeSquares <= gamedata.MissileAMRMaxRangeSquares {
		if amrRoll <= gamedata.MissileAMRChanceToHit(gamedata.MissileAMRRangeIndex(amrRangeSquares)) {
			// 手冊明定 AMR 命中只摧毀彈頭堆疊中的一枚；MIRV 的其餘彈頭仍可繼續。
			warheads--
			if warheads == 0 {
				return ShotResult{Hit: false, RemainingArmorHP: armorHP} // 被 AMR 擊落
			}
		}
	}

	if gamedata.WeaponModMissileECCM(mods) {
		hasECCM = true
	}
	jamChance := gamedata.MissileJamChance(defenderEvasionBonus, attackerScannerBonus, hasECCM)
	hitChance := gamedata.MissileDefaultHitChance - jamChance
	if hitChance > 100 {
		hitChance = 100
	}
	if hitChance < 0 {
		hitChance = 0
	}
	// AMR 已摧毀一枚彈頭時，warheads 是剩餘數；每個剩餘彈頭仍各自經過下列三個判定。
	// 舊的單骰欄位是第 0 枚；MIRV 若呼叫端沒有提供切片，回退值可讓公式仍可測試且不
	// 改變非 MIRV 舊路徑。
	rollAt := func(rolls []int, idx, fallback int) int {
		if idx >= 0 && idx < len(rolls) && rolls[idx] != 0 {
			return rolls[idx]
		}
		return fallback
	}
	remainingArmor := armorHP
	totalStructure, totalShield := 0, 0
	damageMultiplier := gamedata.WeaponModMissileDamageMultiplier(mods, torpedo)
	baseDamage := weaponMax
	if torpedo {
		baseDamage = gamedata.TorpedoDamageAfterRange(weaponName, baseDamage, amrRangeSquares,
			gamedata.WeaponModNoRangeDissipation(mods))
	}
	damage := baseDamage * damageMultiplier / 100
	armorPiercing := gamedata.WeaponModMissileArmorPiercing(mods)
	hit := false
	for i := 0; i < warheads; i++ {
		if rollAt(def.JamRolls, i, jamRoll) > hitChance {
			continue // 被干擾/閃避
		}
		// 匿蹤:手冊「missiles and torpedoes have a 50% chance to miss」。排在干擾之後、
		// 位移裝置之前；MIRV 的每枚彈頭獨立判定。
		if def.CloakMissChance > 0 && rollAt(def.CloakRolls, i, def.CloakRoll) <= def.CloakMissChance {
			continue
		}
		// 位移裝置的「regardless of any other equipment or situation」判定在最後。
		if def.HasDisplacement && rollAt(def.DisplacementRolls, i, def.DisplacementRoll) <= gamedata.MissileDisplacementDeviceMissChance {
			continue
		}
		hit = true
		dmg := gamedata.DamageAfterShield(damage, shieldReduction, hardShield, false)
		totalShield += gamedata.DamageShieldAbsorbed(damage, shieldReduction)
		_, toStruct, nextArmor := gamedata.DamageApplyArmor(dmg, remainingArmor, armorPiercing, false)
		totalStructure += toStruct
		remainingArmor = nextArmor
	}
	if !hit {
		return ShotResult{Hit: false, RemainingArmorHP: armorHP}
	}
	return ShotResult{Hit: true, DamageToStructure: totalStructure, RemainingArmorHP: remainingArmor,
		ShieldDamage: totalShield}
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
	return ShotResult{Hit: true, DamageToStructure: toStruct, RemainingArmorHP: remArmor,
		ShieldDamage: gamedata.DamageShieldAbsorbed(aggD, shieldReduction)}
}
