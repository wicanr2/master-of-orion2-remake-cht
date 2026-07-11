package gamedata

// 武器改造(Weapon Modifications)資料層:逐一移植 MOO2 patch 1.5 官方手冊
// (moo2_patch1.5/GAME_MANUAL.pdf,pdftotext -layout 抽字,p.115-118「Modifications」章節,
// 標題原文「Nearly every weapon could be enhanced in some way ... What follows is an
// introduction to the potential modifications available for the weapons you can install in
// your ships. ... each one adds to the size and cost of the weapon, and some are mutually
// exclusive」)。本檔只收「本輪要接線」的 8 個光束/通用 mod(HV/PD/AF/CO/AP/ENV/NR/SP);
// 飛彈/魚雷專屬 mod(ARM/ECCM/EMG/FST/MV/OVR、以及 NR 的魚雷版本「Not Reduced By Range」)
// 手冊同樣給了精確數字,但 remake 的飛彈解算(missile.go/ResolveMissileShot)目前尚未有
// mod 掛鉤機制,故不在此定義,留待飛彈 mod 任務時再移植(避免定義了卻無人使用的死碼)。
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
	// [誠實記錄] remake 的 shell.ResolveShot 簡化模型目前不對 weaponMin/weaponMax 套用
	// 任何射程衰減(固定近距離解算,dissipation 一直是 0),故 NR 在現行引擎下沒有可觀察的
	// 效果——不是遺漏,是因為「要消除的衰減」本身還沒被模擬。保留此 mod 的佔格/成本數字供
	// 玩家設計與存檔使用,傷害效果待 shell 導入射程衰減後再串接(TODO,見
	// docs/tech/weapon-mods.md)。
	ModNoRangeDissipation WeaponModCode = "NR"
	// ModShieldPiercing 手冊原文(p.118):「SP: Shield Piercing weapons ignore the target's
	// shields completely, passing through as if there were no shields. This modification has no
	// effect against planetary shields. Adding Shield Piercing increases the size and cost of
	// the weapon by 50% and is not applicable until the intended weapon has undergone 1 level
	// of miniaturization.」
	ModShieldPiercing WeaponModCode = "SP"
)

// WeaponModAutoFireFlatSpaceCost 见 ModAutoFire 註解:手冊原文「by 50」非「by 50%」,固定值。
const WeaponModAutoFireFlatSpaceCost = 50

// weaponModSpaceCostPercent 各 mod 對佔格/成本的百分比變動(正值=增加,負值=減少)。
// ModAutoFire 不在此表(固定值,見 WeaponModAutoFireFlatSpaceCost),其餘 7 個逐一對應
// 上方常數註解引用的手冊百分比。
var weaponModSpaceCostPercent = map[WeaponModCode]int{
	ModHeavyMount:         100,
	ModPointDefense:       -50,
	ModContinuousFire:     50,
	ModArmorPiercing:      50,
	ModEnveloping:         100,
	ModNoRangeDissipation: 25,
	ModShieldPiercing:     50,
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
func WeaponModShieldPiercing(mods []WeaponModCode) bool {
	return WeaponModHas(mods, ModShieldPiercing)
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
