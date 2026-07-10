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
// 加成——目前呼叫端皆固定傳 0 PD bonus)。
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
	level := gamedata.CombatRangeLevel(rangeSquares)
	penalty := gamedata.CombatRangeLevelPenalty(level)
	threshold := gamedata.CombatHitThreshold(penalty, 0)

	if !gamedata.CombatClassicToHit(roll, netAttack, threshold) {
		return ShotResult{Hit: false, RemainingArmorHP: armorHP}
	}
	dmg := gamedata.DamageForHit(weaponMin, weaponMax, roll, netAttack, threshold)
	dmg = gamedata.DamageAfterShield(dmg, shieldReduction, hardShield, false)
	_, toStruct, remArmor := gamedata.DamageApplyArmor(dmg, armorHP, armorPiercing, false)
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
//     距離(格,→ gamedata.MissileAMRRangeIndex →命中率)。現行 remake 的 SpecialOptions
//     尚未提供「反飛彈火箭」這個可造艦元件,呼叫端目前一律傳 hasAMR=false(TODO:待新增
//     該元件後,依目標艦是否裝載決定,不在此臆造裝載狀態)。
//   - defenderEvasionBonus:目標的飛彈閃避加成加總(ECM Jammer/Stabilizer/種族/艦員/
//     統帥,各項手冊固定數值見 missile.go 的 MissileJammer*/MissileInertialStabilizer/
//     MissileInertialNullifier/MissileShipDefenseRacialBonus/MissileCrew*/
//     MissileHelmsmanEvasionBonus)。現行 remake 的艦艇設計/軍官系統尚未提供這些元件,
//     呼叫端目前一律傳 0(TODO,待補上後從實際裝載/軍官推導)。
//   - attackerScannerBonus/hasECCM:同理,現行 remake 未提供攻方掃描器(Scanner)、飛彈
//     ECCM 元件,呼叫端一律傳 0/false(TODO)。
//     以上四項在「無任何裝備」時退化為手冊「若目標無任何閃避能力,預設100%命中」
//     (gamedata.MissileDefaultHitChance)——這是手冊本身的基準情境,不是臆造值,恰好與
//     現行武器/元件表(尚無任何閃避裝備)的現況一致。
//   - weaponMax:飛彈命中後的傷害。手冊只列固定「listed」傷害值(如「Nuclear Missile
//     Damage lowered from 8 to 6」),沒有給出像 beam 命中裕度那樣的內插公式,故不套用
//     beam 專用的 gamedata.DamageForHit(那需要 net-attack/hit-threshold,是命中判定
//     機制不同的 beam 概念,套用會混淆兩種機制);仍依手冊預設(只有掛 Shield
//     Piercing/Armor Piercing mod 才豁免,本 remake 尚未對飛彈掛任何 mod)穿過護盾/裝甲。
func ResolveMissileShot(
	hasAMR bool, amrRangeSquares, amrRoll int,
	defenderEvasionBonus, attackerScannerBonus int, hasECCM bool, jamRoll int,
	weaponMax, shieldReduction, armorHP int, hardShield bool,
) ShotResult {
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
		return ShotResult{Hit: false, RemainingArmorHP: armorHP} // 被幹擾/閃避
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
