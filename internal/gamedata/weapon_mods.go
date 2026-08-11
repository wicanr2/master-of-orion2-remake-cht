package gamedata

// 武器改造(Weapon Modifications)資料層:逐一移植 MOO2 patch 1.5 官方手冊
// (moo2_patch1.5/GAME_MANUAL.pdf,pdftotext -layout 抽字,p.115-118「Modifications」章節,
// 標題原文「Nearly every weapon could be enhanced in some way ... What follows is an
// introduction to the potential modifications available for the weapons you can install in
// your ships. ... each one adds to the size and cost of the weapon, and some are mutually
// exclusive」)。本檔包含已接線的光束/通用與飛彈/魚雷改造。ARM/FST 的完整原版攔截器資料流仍
// 有未解欄位，但本專案已把可由原始碼證實的 raw flag 與 remake 快速戰鬥消費端分開
// 記錄；因此兩項改造可以進入資料層，並不代表所有原版戰術細節已經解完。
//
// 出處逐項核對(GAME_MANUAL.pdf p.115-118 原文摘錄,見各常數註解),配合 docs/tech/weapon-mods.md
// 完整記錄。手冊「Modifications」章節與 shipspace.go 引用的「p.128 Design Dock」章節是同一件事的
// 兩處描述(p.128 只重申「會加大 size/cost」的原則,精確數字在 p.115-118),不衝突。

// WeaponModCode 是武器改造的手冊縮寫代碼(如 "HV"、"PD"),直接對應手冊行文用的縮寫,
// 也是 shell.Ship.Mods / Component 序列化時使用的字串值(存檔相容,故不要改動既有代碼字串)。
type WeaponModCode string

const (
	// ModHeavyMount 手冊原文(p.117):「HV: Heavy Mount beam weapons are large-platform
	// versions that cause 150% of the normal amount of damage. In addition, the increased
	// strength of the beam cuts the range penalties (for accuracy and dissipation of damage)
	// in half. This modification increases the size and cost of the weapon by 100%. The Heavy
	// Mount and Point Defense modifications are mutually exclusive.」
	ModHeavyMount WeaponModCode = "HV"
	// ModPointDefense 手冊原文(p.117):「PD: Point Defense weapons are small, precise copies
	// of a beam weapon used to target missiles and fighter craft. They inflict only half the
	// damage of a full-size beam, but have a 25% greater accuracy. Since these are intended
	// only as short-range defensive batteries, the range penalties to dissipation and accuracy
	// are double. All available (nonfired) Point Defense beams fire automatically at any
	// incoming target in the same square as the ship. This modification decreases the size and
	// cost of the weapon by half (50%). The Point Defense and Heavy Mount modifications are
	// mutually exclusive.」
	ModPointDefense WeaponModCode = "PD"
	// ModAutoFire 手冊原文(p.115):「AF: Auto-Fire allows a beam weapon to fire 3 separate
	// times in rapid succession, each time with a 20% penalty to its accuracy. This
	// modification increases the size and cost of the weapon by 50 and is not applicable until
	// the intended weapon has undergone 2 levels of miniaturization.」
	//
	// 注意「by 50」不是「by 50%」——手冊其餘 mod 條目一律明寫「by X%」,唯獨 AF 這條寫
	// 「by 50」(無 % 符號),與其餘百分比 mod 的措辭明顯不同,故本檔把 AF 佔格處理為
	// 固定值 +50(非百分比),見 WeaponModAutoFireFlatSpaceCost。
	ModAutoFire WeaponModCode = "AF"
	// ModContinuousFire 手冊原文(p.115):「CO: Continuous fire prevents a beam weapon from
	// overheating as quickly, allowing it to fire over a longer duration. This gives the
	// targeting computer time to adjust the aim during fire, increasing the weapon's accuracy
	// by 25. This modification increases the size and cost of the weapon by 50% and is not
	// applicable until the intended weapon has undergone 1 level of miniaturization.」
	ModContinuousFire WeaponModCode = "CO"
	// ModArmorPiercing 手冊原文(p.115):「AP: Armor Piercing beam weapons penetrate any type
	// of armor except Xentronium and Heavy Armor. All of the damage done passes through as if
	// there were no armor at all. AP adds 50% to the space and cost of a weapon and is not
	// applicable until the intended weapon has undergone 1 level of Miniaturization.」
	ModArmorPiercing WeaponModCode = "AP"
	// ModEnveloping 手冊原文(p.116):「ENV: Enveloping weapons, whether beams or torpedoes,
	// surround the target at impact and strike all four shield quarters simultaneously. This
	// effectively quadruples the damage done by the hit. This modification increases the size
	// and cost of the weapon by 100% and is not applicable until the intended weapon has
	// undergone 2 levels of miniaturization.」
	ModEnveloping WeaponModCode = "ENV"
	// ModNoRangeDissipation 手冊原文(p.116,beam 版本):「NR: No Range Dissipation affects
	// those beam weapons that diminish in strength (potential damage) over distance. Using an
	// independent collimation beam and continual chaotic feedback analysis, this device
	// focuses the beam and totally eliminates the decrease in damage. This modification
	// increases the size and cost of the weapon by 25% and is not applicable until the intended
	// weapon has undergone 1 level of miniaturization.」
	//
	// 2026-08-06 已接線:shell.ResolveShotWithMods 會依 range level 套 DamageDissipationPenalty,
	// 帶 NR 時跳過。完整消費端是原版 `Get_Beam_Weapon_Modifiers_`(raw `sub_394F7`)，
	// 其中再呼叫射程／命中輔助 `sub_39434`；對應該輔助所使用的
	// `if (!(mods & 0x20) && !(weaponFlags & 0x04)) *damagePct += ranged_damage_penalty[range]`。
	// (先前這裡記著「NR 沒有可觀察效果」,因為衰減本身還沒模擬——那個 TODO 已解。)
	ModNoRangeDissipation WeaponModCode = "NR"
	// ModShieldPiercing 手冊原文(p.118):「SP: Shield Piercing weapons ignore the target's
	// shields completely, passing through as if there were no shields. This modification has no
	// effect against planetary shields. Adding Shield Piercing increases the size and cost of
	// the weapon by 50% and is not applicable until the intended weapon has undergone 1 level
	// of miniaturization.」
	ModShieldPiercing WeaponModCode = "SP"
	// ModMissileECCM 手冊 p.116：ECCM 使單一彈頭被干擾的機率減半。
	ModMissileECCM WeaponModCode = "ECCM"
	// ModEmissionsGuidance 手冊 p.116：命中護盾後繞過裝甲，直接傷害結構。
	ModEmissionsGuidance WeaponModCode = "EMG"
	// ModMIRV 手冊 p.116：一枚飛彈分成四枚完整彈頭；不適用於行星轟炸。
	ModMIRV WeaponModCode = "MV"
	// ModOverloadedTorpedo 手冊 p.116：魚雷整套過載，彈頭強度增加 50%。
	ModOverloadedTorpedo WeaponModCode = "OVR"
	// ModArmoredMissile 手冊 p.116：重裝飛彈，摧毀所需傷害加倍。
	// 原版 raw flag 的 0x0800 對應是強推論：`Missile_Dcv` @ 0x3E095 的高位元
	// 0x08 使攔截耐久加倍，而 `Weapon_In_Range` @ 0x3A0B9 會用該耐久計算擊落數。
	ModArmoredMissile WeaponModCode = "ARM"
	// ModFastMissile 手冊 p.116：快速飛彈；原版 raw `sub_3CD21` @ 0x3CD21
	//（func_names.txt 原名 `Missile_Facing_`，舊匯出另稱 `Missile_Speed_`）
	// 明確在 raw high-byte bit 0x10 成立時加 4 速度，故其 raw flag 對應已證實。
	ModFastMissile WeaponModCode = "FST"
)

// WeaponModAutoFireFlatSpaceCost 见 ModAutoFire 註解:手冊原文「by 50」非「by 50%」,固定值。
const WeaponModAutoFireFlatSpaceCost = 50

// weaponModSpaceCostPercent 各 mod 對佔格/成本的百分比變動(正值=增加,負值=減少)。
// ModAutoFire 不在此表(固定值,見 WeaponModAutoFireFlatSpaceCost),其餘百分比 mod 逐一對應
// 上方常數註解引用的手冊百分比。
var weaponModSpaceCostPercent = map[WeaponModCode]int{
	ModHeavyMount:         100,
	ModPointDefense:       -50,
	ModContinuousFire:     50,
	ModArmorPiercing:      50,
	ModEnveloping:         100,
	ModNoRangeDissipation: 25,
	ModShieldPiercing:     50,
	ModMissileECCM:        25,
	ModEmissionsGuidance:  300, // ×4 size/cost
	ModMIRV:               100,
	ModOverloadedTorpedo:  50,
	ModArmoredMissile:     25,
	ModFastMissile:        25,
}

// WeaponModSpaceCostPercent 查表回傳 mod 對佔格/成本的百分比變動;ok=false 表示該 mod
// 是固定值(目前只有 ModAutoFire)或不是已知 mod 代碼。
func WeaponModSpaceCostPercent(mod WeaponModCode) (percent int, ok bool) {
	p, ok := weaponModSpaceCostPercent[mod]
	return p, ok
}

// WeaponModHas 回傳 mods 中是否含指定 mod(通用小工具,mods 一律用字串比較,呼叫端可傳
// []WeaponModCode)。
func WeaponModHas(mods []WeaponModCode, target WeaponModCode) bool {
	for _, m := range mods {
		if m == target {
			return true
		}
	}
	return false
}

// WeaponSpaceWithMods 依基礎佔格(如 WeaponSpaceByName 查到的值)套用一組 mod 的佔格變動,
// 回傳套用後的佔格。百分比 mod 採「加總後一次套用」(對照 damage.go 的
// DamageMountAdjustedValue 手冊明載的「Hv/PD/HEF interaction is not multiplicative but
// additive」慣例);手冊本身沒有明講「同一武器同時掛多個 mod 時,佔格百分比是加總一次套用
// 還是逐個連續複利套用」,本函式採加總一次套用是依現有 Hv/PD/HEF 傷害公式的既定慣例類推,
// 不是手冊逐字數字,誠實標註於此(見 docs/tech/weapon-mods.md)。ModAutoFire 的固定 +50
// 在百分比套用「之後」再相加(手冊明寫是固定值,不應被其他 mod 的百分比再放大)。
// 結果最少 1(避免 0 或負值佔格)。
func WeaponSpaceWithMods(baseSpace int, mods []WeaponModCode) int {
	pctSum := 0
	flat := 0
	for _, m := range mods {
		if p, ok := weaponModSpaceCostPercent[m]; ok {
			pctSum += p
		}
		if m == ModAutoFire {
			flat += WeaponModAutoFireFlatSpaceCost
		}
	}
	space := baseSpace + baseSpace*pctSum/100 + flat
	if space < 1 {
		space = 1
	}
	return space
}

// WeaponCostWithMods 手冊原文「adds to the size AND cost」——同一套百分比/固定值同時套用在
// 成本上,故直接重用 WeaponSpaceWithMods 的公式(對 baseCost 而非 baseSpace)。
func WeaponCostWithMods(baseCost int, mods []WeaponModCode) int {
	return WeaponSpaceWithMods(baseCost, mods)
}

// ---- 命中率(to-hit)效果:CO/AF/PD 對 netAttack / hit_threshold 的貢獻 ----
//
// combat.go 的 CombatClassicToHit/CombatAlternativeToHit 沿用社群逆向出的
// 「netAttack = BA + CO - AF - BD」記號(docs/tech/community-mechanics-findings.md 引用
// Olesh 的拆解,交叉核對來源中等可信度),CO/AF 在該記號裡本來就是「加成點數」項,恰好
// 與手冊本節「increasing the weapon's accuracy by 25」「20% penalty to its accuracy」的
// 描述吻合(手冊在 Beam Attack/Defense 這套系統中的其他加成也全是點數制,如電腦每級 +25、
// 船員 Elite +75,而非額外的百分比縮放),故 CO=+25、AF(每次射擊)=-20 採直接點數解讀。
const (
	// WeaponModAccuracyBonusContinuousFire 手冊原文「increasing the weapon's accuracy by 25」。
	WeaponModAccuracyBonusContinuousFire = 25
	// WeaponModAccuracyPenaltyAutoFire 手冊原文「each time with a 20% penalty to its
	// accuracy」,套用在 BA+CO-AF-BD 的 AF 項(點數制,見上方段落說明)。
	WeaponModAccuracyPenaltyAutoFire = 20
	// WeaponModAutoFireShots 手冊原文「fire 3 separate times in rapid succession」。
	WeaponModAutoFireShots = 3
	// WeaponModPointDefenseAccuracyBonus 手冊原文「have a 25% greater accuracy」,套用在
	// combat.go CombatHitThreshold 的 pdBonus 參數(該參數註解原本就標記「手冊未在本節給出
	// 精確數字」,本檔補上)。
	WeaponModPointDefenseAccuracyBonus = 25
	// WeaponModEnvelopingDamageMultiplier 手冊原文「effectively quadruples the damage done
	// by the hit」。
	WeaponModEnvelopingDamageMultiplier = 4
)

// WeaponModNetAttackBonus 回傳 CO/AF 對 netAttack(BA+CO-AF-BD)的加減點數總和。
// AF 為每次射擊固定 -20(手冊描述的是單次開火的懲罰,若要模擬「3 連發」需呼叫端自行
// 對 3 發個別呼叫本函式,本函式不重複「發幾次」的邏輯,那屬於呼叫端的迴圈次數控制,
// 見 WeaponModAutoFireShots)。
func WeaponModNetAttackBonus(mods []WeaponModCode) int {
	bonus := 0
	for _, m := range mods {
		switch m {
		case ModContinuousFire:
			bonus += WeaponModAccuracyBonusContinuousFire
		case ModAutoFire:
			bonus -= WeaponModAccuracyPenaltyAutoFire
		}
	}
	return bonus
}

// WeaponModPDBonus 回傳 PD mod 對 CombatHitThreshold 的 pdBonus 貢獻(未掛 PD 回 0)。
func WeaponModPDBonus(mods []WeaponModCode) int {
	if WeaponModHas(mods, ModPointDefense) {
		return WeaponModPointDefenseAccuracyBonus
	}
	return 0
}

// ---- 傷害:HV/PD/ENV/AP/SP 對傷害解算的貢獻 ----

// WeaponModDamageBonuses 回傳 (hvBonus, pdPenalty),直接餵給 damage.go 既有的
// DamageMountAdjustedValue(base, hvBonus, hefBonus, pdPenalty, rangePenaltyPoints)。
// hvBonus/pdPenalty 沿用該檔已有的 DamageMountBonusHeavy/DamageMountPenaltyPointDefense
// 常數(不在此重複定義數字,避免兩處常數各自維護、日後改一邊漏改另一邊)。
func WeaponModDamageBonuses(mods []WeaponModCode) (hvBonus, pdPenalty int) {
	if WeaponModHas(mods, ModHeavyMount) {
		hvBonus = DamageMountBonusHeavy
	}
	if WeaponModHas(mods, ModPointDefense) {
		pdPenalty = DamageMountPenaltyPointDefense
	}
	return hvBonus, pdPenalty
}

// WeaponModEnvelopingMultiply 若掛 ENV,對命中後的傷害 *4(手冊「quadruples the damage
// done by the hit」);未掛 ENV 原樣回傳,是中性 no-op。
func WeaponModEnvelopingMultiply(dmg int, mods []WeaponModCode) int {
	if WeaponModHas(mods, ModEnveloping) {
		return dmg * WeaponModEnvelopingDamageMultiplier
	}
	return dmg
}

// WeaponModArmorPiercing 回傳是否掛 AP(直接對應 damage.go DamageApplyArmor 的
// armorPiercing 參數)。
func WeaponModArmorPiercing(mods []WeaponModCode) bool {
	return WeaponModHas(mods, ModArmorPiercing)
}

// WeaponModShieldPiercing 回傳是否掛 SP(直接對應 damage.go DamageAfterShield 的
// shieldPiercing 參數)。
// WeaponModNoRangeDissipation 回傳該組改造是否含 NR(No Range Dissipation)——
// 有的話命中後的傷害不吃距離衰減。
//
// 對應原版 `Get_Beam_Weapon_Modifiers_`(raw `sub_394F7`) 呼叫
// `Get_Beam_Range_To_Hit_Bonus_`(raw `sub_39434`) 的這一段:
//
//	if (!(mods & 0x20) && !(weaponFlags & 0x04))
//	    *damagePct += ranged_damage_penalty[range];
//
// 即「除非帶了這個改造、或武器本身天生免疫,否則一律套距離衰減」。
// (weaponFlags & 0x04 那一支是「天生不衰減」的武器如質量投射器/高斯砲,
// remake 的武器資料尚未帶這個旗標,見 damage.go DamageApplyDissipation 註解。)
func WeaponModNoRangeDissipation(mods []WeaponModCode) bool {
	return WeaponModHas(mods, ModNoRangeDissipation)
}

func WeaponModShieldPiercing(mods []WeaponModCode) bool {
	return WeaponModHas(mods, ModShieldPiercing)
}

// WeaponModMissileWarheadCount 回傳這一輪飛彈攻擊要解算的彈頭數。
// MIRV 的「四枚完整彈頭」不是把最後傷害粗暴乘四：每枚彈頭都要各自經過干擾、匿蹤與
// 位移判定，呼叫端因此會逐彈頭提供骰子。沒有 MV 時維持單彈頭。
func WeaponModMissileWarheadCount(mods []WeaponModCode) int {
	if WeaponModHas(mods, ModMIRV) {
		return 4
	}
	return 1
}

// WeaponModMissileDamageMultiplier 回傳單一彈頭的傷害倍率。
// MV 的四枚彈頭由 WeaponModMissileWarheadCount 表達，不在這裡重複乘四；ENV 只在魚雷
// 呼叫端通過適用性篩選後使用，OVR 也只對魚雷生效。
func WeaponModMissileDamageMultiplier(mods []WeaponModCode, torpedo bool) int {
	multiplier := 100
	if torpedo && WeaponModHas(mods, ModOverloadedTorpedo) {
		multiplier = multiplier * 150 / 100
	}
	if torpedo && WeaponModHas(mods, ModEnveloping) {
		multiplier *= WeaponModEnvelopingDamageMultiplier
	}
	return multiplier
}

// WeaponModMissileECCM 回報飛彈是否具備 ECCM。保留成 gamedata helper，讓戰鬥解算端不必
// 直接依賴字串代碼。
func WeaponModMissileECCM(mods []WeaponModCode) bool {
	return WeaponModHas(mods, ModMissileECCM)
}

// WeaponModMissileArmored 回報 ARM；其戰鬥效果由飛彈攔截器的
// MissileInterceptionDurability 消費，不能當成命中後艦體傷害倍率。
func WeaponModMissileArmored(mods []WeaponModCode) bool {
	return WeaponModHas(mods, ModArmoredMissile)
}

// WeaponModMissileFast 回報 FST；其戰鬥效果是飛彈速度／Beam Defense 路徑的 +4，
// 不是命中後傷害加成。
func WeaponModMissileFast(mods []WeaponModCode) bool {
	return WeaponModHas(mods, ModFastMissile)
}

// MissileRawFlagsForMods 將目前已知的飛彈改造對應到原版設計／runtime raw flags。
//
// ARM 的 0x0800 是「強推論」：手冊效果、`Missile_Dcv` 的 doubling branch 與
// `Weapon_In_Range` 的攔截分段鏈相符，但原始程式沒有可讀的 ARM 符號可直接證明。
// FST 的 0x1000 是「已證實」：`Missile_Speed_` 的 high-byte bit 0x10 直接控制 +4。
func MissileRawFlagsForMods(mods []WeaponModCode) uint16 {
	var flags uint16
	if WeaponModMissileArmored(mods) {
		flags |= MissileRawFlagArmored
	}
	if WeaponModMissileFast(mods) {
		flags |= MissileRawFlagFast
	}
	return flags
}

// WeaponModMissileArmorPiercing 回報 EMG 是否應在扣除護盾後繞過裝甲。
// 「直接傷害結構」是手冊對 EMG 的文字；它沒有宣稱穿過護盾，因此仍先走護盾吸收。
func WeaponModMissileArmorPiercing(mods []WeaponModCode) bool {
	return WeaponModHas(mods, ModEmissionsGuidance)
}

// ---- 射程等級(range level)選擇:HV 減半、PD 加倍、其餘用一般表 ----

// CombatRangeLevelForBeamMods 依掛載的 mod 選擇正確的射程等級函式:HV 用
// CombatRangeLevelHeavy(手冊「the actual range is halved」)、PD 用
// CombatRangeLevelPointDefense(手冊「range is as if doubled」)、都沒掛用一般
// CombatRangeLevel。HV/PD 手冊明訂互斥,若呼叫端誤同時傳兩者,以 HV 優先(不應發生,
// shell.ToggleWeaponMod 在 UI 層已擋掉同時掛載)。
func CombatRangeLevelForBeamMods(squares int, mods []WeaponModCode) int {
	switch {
	case WeaponModHas(mods, ModHeavyMount):
		return CombatRangeLevelHeavy(squares)
	case WeaponModHas(mods, ModPointDefense):
		return CombatRangeLevelPointDefense(squares)
	default:
		return CombatRangeLevel(squares)
	}
}
