package gamedata

// 光束傷害解算(命中「後」的傷害量):距離衰減(dissipation)、Hv/PD/HEF mount 加成、
// 護盾(shield)、裝甲穿透(Armor Piercing mod)。逐一移植 MOO2 patch 1.5 官方手冊
// (moo2_patch1.5/MANUAL_150.html,"Notes on Beam Weapon Mechanics > Damage Potential" 章節;
// moo2_patch1.5/GAME_MANUAL.pdf 的 Shield/Armor/Weapon Mods 條目補充),以及部分
// "Notes on Spherical Damage > Damage Calculation" 章節(僅移植手冊講清楚數字的片段)。
//
// 本檔專注「命中後的傷害量」,命中率(to-hit)已在 combat.go 移植,不重複定義
// CombatRangeLevel* 系列,而是沿用 combat.go 算出的 range level(0-8)、
// CombatHitThreshold 作為輸入。damage.go 自己的「距離衰減表」數值與 combat.go 的
// 「to-hit 距離懲罰表」不同,不可混用(見 damageDissipationPenaltyTable 註解)。
// 命名一律加 Damage 前綴,不定義通用 helper(如 round/clamp)。

// damageDissipationPenaltyTable 手冊 Damage Potential 表(Reduced by Range 小節)的
// 「Penalty」列,依 range level 0-8:
//
//	Range(sq) 0    1-3  4-6  7-9  10-12 13-15 16-18 19-21 22-24
//	level     0    1    2    3    4     5     6     7     8
//	Penalty   0    0    -10  -20  -30   -40   -50   -60   -65
//	Dmg%      100% 100% 90%  80%  70%   60%   50%   40%   35%
//
// 原始 HTML(<table class="c23">)用 colspan="2" 讓第一格同時覆蓋 range level 0(0 sq)與
// level 1(1-3 sq),純文字擷取後每列只剩 8 個數字,對齊時已核對原始 HTML 的 colspan 屬性
// 確認第一格套用在 level0 與 level1 兩者(而非「Range 9 欄 vs Penalty 8 欄對不齊」的錯位)。
// 本表存正值(懲罰百分點),供 DamageMountAdjustedValue 的 rangePenaltyPoints 參數使用。
//
// 與 combat.go 的 combatRangeLevelPenaltyTable(to-hit 用,0,0,10,20,30,40,55,70,85)是
// 兩張不同的表,不可混用:那張決定「命不命中」,這張決定「命中後傷害打幾折」。
var damageDissipationPenaltyTable = [9]int{0, 0, 10, 20, 30, 40, 50, 60, 65}

// DamageDissipationPenalty 依 range level(0-8,由 combat.go 的 CombatRangeLevel /
// CombatRangeLevelPointDefense / CombatRangeLevelHeavy 算得)查出傷害衰減懲罰值(百分點)。
// level 超出 0-8 一律夾限到最近端點(手冊未列更遠距離)。
func DamageDissipationPenalty(level int) int {
	if level < 0 {
		level = 0
	}
	if level > 8 {
		level = 8
	}
	return damageDissipationPenaltyTable[level]
}

// 手冊武器 mount/mod 對傷害的固定加成(百分點),與距離衰減共用同一套「百分點相加減、非
// 乘法」規則(見 DamageMountAdjustedValue)。出處:GAME_MANUAL.pdf「Weapon Mods」附錄與
// System 條目。
const (
	// DamageMountBonusHeavy 手冊原文:「HV: Heavy Mount beam weapons are large-platform
	// versions that cause 150% of the normal amount of damage.」即 +50 百分點。
	DamageMountBonusHeavy = 50
	// DamageMountPenaltyPointDefense 手冊原文:「PD: Point Defense weapons ... inflict only
	// half the damage of a full-size beam」即 -50 百分點。
	DamageMountPenaltyPointDefense = 50
	// DamageMountBonusHEF 手冊原文(High Energy Focus (System)):「increasing the damage each
	// of these weapons inflicts by 50%.」即 +50 百分點。
	DamageMountBonusHEF = 50
)

// damageRoundDiv100 算 value/100 並四捨五入到最接近整數(手冊原文:「Round to nearest
// applies」)。用整數運算「+50 再整除」避免浮點誤差;此技巧僅對「除以 100」成立,不是通用
// round helper,故不外露、也不用於其他分母。
func damageRoundDiv100(value int) int {
	if value >= 0 {
		return (value + 50) / 100
	}
	return -((-value + 50) / 100)
}

// DamageMountAdjustedValue 實作手冊「Hv, PD, HEF & Ordnance」小節的傷害加成公式。手冊原文:
//
//	「These bonuses are percentages of damage of a regular mount beam. ... Their interaction
//	is not multiplicative but additive and they work the same way as dissipation range
//	penalty, modifying min&max damage, so these bonuses are not subject to dissipation
//	themselves.」
//
// 並以「a 50-100 beam with dissipation Range_penalty of 30」逐一驗算 5 組、共 10 個數字
// (完整複現於 damage_test.go):
//
//	Hv:         50*(100+50-30)%  = 60 ;  100*(100+50-30)%  = 120
//	Hv+HEF:     50*(100+50+50-30)% = 85 ; 100*(100+50+50-30)% = 170
//	PD:         50*(100-50-30)%  = 10 ;  100*(100-50-30)%  = 20
//	PD(2x):     50*(100-50-2*30)% = -5(夾為1); 100*(100-50-2*30)% = -10(夾為1)
//	PD+HEF(2x): 50*(100-50+50-2*30)% = 20 ; 100*(100-50+50-2*30)% = 40
//
// base 為武器基礎 min 或 max 傷害;hvBonus/hefBonus/pdPenalty 用上方 DamageMountBonus* /
// DamageMountPenaltyPointDefense 常數(不適用填 0);rangePenaltyPoints 用
// DamageDissipationPenalty 查出的百分點。手冊 1.50 新增「2x range 衰減」可選規則(某些武器
// mod 讓 to-hit 與 dissipation 的距離懲罰加倍,是否加倍、如何加倍屬於 config 選項而非固定
// 數字)是否加倍由呼叫端決定(對照 combat.go 的 CombatRangeLevelPenaltyDoubled 慣例),若
// 加倍就先乘 2 再傳入 rangePenaltyPoints,本函式不重複那段邏輯。
// 「the minimum damage potential is always 1」故結果一律夾限最小為 1。
func DamageMountAdjustedValue(base, hvBonus, hefBonus, pdPenalty, rangePenaltyPoints int) int {
	pct := 100 + hvBonus + hefBonus - pdPenalty - rangePenaltyPoints
	v := damageRoundDiv100(base * pct)
	if v < 1 {
		v = 1
	}
	return v
}

// DamageApplyDissipation 只套用距離衰減(不含 Hv/PD/HEF),對應手冊「Reduced by Range」小節
// 給出的 Laser Cannon/Phasor/Mauler/Death Ray 範例表。
//
// 注意(誠實記錄一個對不上的儲存格):手冊 Laser Cannon(base 1-4)範例列在 level7
// (19-21 sq,Dmg% 40%)欄印「1-1」,但用本函式公式算 round(4*40/100)=round(1.6)=2,對不上;
// Phasor(base 5-20)、Mauler(base 100-100)、以及手冊「Different Min-Max Damage」小節
// 逐步驗算的 Death Ray(base 50-100)兩個 worked example,則與此公式完全吻合(共 30+ 個數字,
// 見 damage_test.go)。研判 Laser 那一格是手冊排版/校對誤差,不因單一儲存格回頭修改已被其餘
// 範例與 Hv/PD/HEF 公式(DamageMountAdjustedValue)充分驗證的通用公式。
//
// NR(Not Reduced by Range,如 Mass Driver/Gauss/Disrupter)武器手冊原文:「no dissipation
// penalty applies」,呼叫端不應呼叫本函式,直接使用基礎 min/max 傷害即可。
func DamageApplyDissipation(minDmg, maxDmg, level int) (int, int) {
	penalty := DamageDissipationPenalty(level)
	return DamageMountAdjustedValue(minDmg, 0, 0, 0, penalty), DamageMountAdjustedValue(maxDmg, 0, 0, 0, penalty)
}

// DamageForHit 實作手冊「Different Min-Max Damage」公式,算出命中後的實際傷害。手冊原文:
//
//	UNCAPPED DAMAGE IS: min_dmg + (max_dmg-min_dmg+1) * A / B
//	WHERE: A = roll_plus_attack - hit_threshold
//	       B = 100 - hit_threshold
//	       roll_plus_attack = min(random(100) + BA+CO-AF-BD; 100)
//	CAPPED DAMAGE IS: 上式結果再夾限於 max_dmg。
//
// 命中判定沿用手冊 [1][2] 兩個「必為 max dmg」的分支,與 combat.go 的 CombatClassicToHit
// 共用同一組 roll/netAttack/hitThreshold 輸入(呼叫端須先用 CombatClassicToHit /
// CombatAlternativeToHit 確認命中,再呼叫本函式取得傷害量,本函式不重複判斷是否命中):
//
//	[1] random(100) > 95         → max_dmg
//	[2] BA+CO-AF-BD >= 99         → max_dmg
//	[3] 其餘情況用上面的內插公式
//
// 已用手冊 Examples 段落兩組逐步驗算數字核對(damage_test.go,roll=85、netAttack=10 為手冊
// 給定的共同輸入,分別對應 range 23 sq 與 11 sq 兩個 hit_threshold):
//   - Death Ray 23 sq(衰減後 minDmg=18,maxDmg=35;hitThreshold=95)→ 18(= min_dmg)
//   - Death Ray 11 sq(衰減後 minDmg=35,maxDmg=70;hitThreshold=70)→ 65(= min_dmg + ⅚*36)
//
// minDmg/maxDmg 須為「已套用距離衰減與 Hv/PD/HEF 加成之後」的傷害潛力(DamageApplyDissipation
// / DamageMountAdjustedValue 算好);hitThreshold 用 combat.go 的 CombatHitThreshold 算好傳入
// (其值已夾限在 95 以內,故下方 b 恆 >= 5,不會除以 0)。
func DamageForHit(minDmg, maxDmg, roll, netAttack, hitThreshold int) int {
	if roll > 95 || netAttack >= 99 {
		return maxDmg
	}
	rollPlusAttack := roll + netAttack
	if rollPlusAttack > 100 {
		rollPlusAttack = 100
	}
	a := rollPlusAttack - hitThreshold
	b := 100 - hitThreshold
	uncapped := minDmg
	if b > 0 {
		uncapped = minDmg + (maxDmg-minDmg+1)*a/b
	}
	if uncapped > maxDmg {
		uncapped = maxDmg
	}
	if uncapped < minDmg {
		uncapped = minDmg
	}
	return uncapped
}

// ---- 護盾(Shield) ----
//
// 手冊原文(GAME_MANUAL.pdf,Class I/III/V/VII/X Shield (Ship) 各條目)。遊戲只有這 5 級
// 護盾(對照 techtree.go 的 TECH_CLASS_I/III/V/VII/X_SHIELD,沒有 II/IV/VI/VIII/IX 這幾級):
//
//	Class I:   absorbing up to 5  times ship's size in damage per facing, 每次攻擊減傷 1  點
//	Class III: absorbing up to 15 times ship's size in damage per facing, 每次攻擊減傷 3  點
//	Class V:   absorbing up to 25 times ship's size in damage per facing, 每次攻擊減傷 5  點
//	Class VII: absorbing up to 35 times ship's size in damage per facing, 每次攻擊減傷 7  點
//	Class X:   absorbing up to 50 times ship's size in damage per facing, 每次攻擊減傷 10 點
//
// 兩個數字都與護盾等級數字本身成正比(每次攻擊減傷 = 等級數字;總容量倍率 = 5 * 等級數字),
// 故不逐級寫死一張表,改由 DamageShieldCapacity 直接算。「regenerate 30% of the maximum
// strength of a facing ... each combat round」手冊未講清楚無條件捨入規則,不移植(TODO 待查證)。
const (
	DamageShieldReductionClassI   = 1
	DamageShieldReductionClassIII = 3
	DamageShieldReductionClassV   = 5
	DamageShieldReductionClassVII = 7
	DamageShieldReductionClassX   = 10
)

// DamageShieldCapacity 算護盾單一 facing 的總可吸收量。手冊原文(以 Class I 為例):
// 「absorbing up to 5 times the ship's size in damage per facing」;III/V/VII/X 依等級數字
// 等比例放大(見上方常數區塊的表)。shieldReduction 用 DamageShieldReductionClass* 常數,
// shipSize 為船體 size class 對應的數值(手冊講的是「ship's size」比例,呼叫端自行從
// enums.go 的 CombatShipClass 或艦體噸位換算,本函式不假設固定對照)。
func DamageShieldCapacity(shieldReduction, shipSize int) int {
	return 5 * shieldReduction * shipSize
}

// DamageHardShieldBonus 手冊原文(Hard Shields (System)):「This reduces the damage of each
// enemy attack — by 3 points — regardless of whether or not the shield in that quarter has
// collapsed.」與 Class 護盾的減傷值相加(見 DamageAfterShield),且「Hard Shields ... provide
// immunity to shield-piercing weapons」。
const DamageHardShieldBonus = 3

// DamageAfterShield 算一次命中在扣掉護盾後的剩餘傷害(供下一步 DamageApplyArmor 使用)。
// shieldPiercing(SP mod)手冊原文:「Shield Piercing weapons ignore the target's shields
// completely, passing through as if there were no shields. This modification has no effect
// against planetary shields.」(本函式只處理艦對艦,行星護盾情境不適用,呼叫端應另行判斷)。
// Hard Shields 使 SP 失效,手冊原文:「Hard Shields ... provide immunity to shield-piercing
// weapons」,故 hardShield 為真時忽略 shieldPiercing。shieldReduction 用
// DamageShieldReductionClass* 常數(無護盾傳 0)。
func DamageAfterShield(dmg, shieldReduction int, hardShield, shieldPiercing bool) int {
	if shieldPiercing && !hardShield {
		return dmg
	}
	reduction := shieldReduction
	if hardShield {
		reduction += DamageHardShieldBonus
	}
	remaining := dmg - reduction
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

// ---- 裝甲穿透(Armor Piercing, AP mod) ----
//
// 手冊原文(GAME_MANUAL.pdf,Weapon Mods 附錄):「AP: Armor Piercing beam weapons penetrate
// any type of armor except Xentronium and Heavy Armor. All of the damage done passes through
// as if there were no armor at all.」以及:
//   - 「Heavy Armor (System): ... This system also negates the Armor Piercing abilities of
//     enemy weapons that hit the ship.」
//   - 「Xentronium Armor: ... Negates armor piercing effects of enemy weapons.」
//
// DamageApplyArmor 算一次命中(已扣過護盾)在裝甲/結構兩層 HP 池之間如何分配。armorPiercing
// 為攻擊方武器是否帶 AP mod;apNegated 為目標艦是否裝有 Heavy Armor 或 Xentronium Armor
// (兩者皆會使 AP 失效,由呼叫端合併判斷後傳入,本函式不重複那兩個系統各自的判定邏輯)。
// 回傳 (裝甲承受量, 結構承受量, 命中後剩餘裝甲HP)。
func DamageApplyArmor(dmg, armorHP int, armorPiercing, apNegated bool) (dmgToArmor, dmgToStructure, remainingArmorHP int) {
	if armorPiercing && !apNegated {
		return 0, dmg, armorHP
	}
	if dmg <= armorHP {
		return dmg, 0, armorHP - dmg
	}
	return armorHP, dmg - armorHP, 0
}

// ---- 球形武器(Spherical Weapon)傷害,僅移植手冊講清楚數字的片段 ----
//
// 手冊出處:"Notes on Spherical Damage > Damage Calculation"(MANUAL_150.html)。球形武器
// (Pulsar/Plasma Flux/Spatial Compressor/引擎爆炸)是與光束武器不同的武器類型,鐵律要求
// 「沒清楚給數字的不要編」,故這裡只移植手冊明確給數字的部分,「Each roll is re-rolled if
// the outcome is not 1」這個重骰機制手冊沒講清楚終止條件與機率分佈,不猜、不移植。

// DamageSphericalRoll 手冊原文:「For each spherical weapon in a slot a damage value is
// generated: D = (DamageMin + random(max-min)) * ordnance」。roll 為呼叫端已擲好的
// random(max-min)(手冊未講清楚含不含端點,由呼叫端負責產生);ordnancePercent 為 Ordnance
// 加成(100=無加成)。手冊只給出各首領浮動的 Ordnance 數字(如 +5%/+10%/+15%),並非單一固定
// 值,故不內建常數,由呼叫端提供。
func DamageSphericalRoll(damageMin, roll, ordnancePercent int) int {
	base := damageMin + roll
	return damageRoundDiv100(base * ordnancePercent)
}

// DamageSphericalShipRollCount 手冊原文(Damage Calculation > Ships):「the number of rolls
// is determined by the size class of the target ship (0,1,2,3,4,5) + 1. Thus a frigate gets
// one roll, a destroyer two rolls, etc.」sizeClass 對應 enums.go 既有的 CombatShipClass
// (0=SHIP_FRIGATE ... 5=SHIP_DOOMSTAR)。骰出的每個 random(aggD) 該怎麼重骰(「re-rolled if
// the outcome is not 1」)手冊描述不足以還原成確定性演算法,不移植,呼叫端需自行決定或標記
// TODO 待進一步查證(如參考 openorion2 原始碼)。
func DamageSphericalShipRollCount(sizeClass CombatShipClass) int {
	return int(sizeClass) + 1
}

// DamageSphericalFlyerDestroyed 對飛彈/戰機的球形武器毀滅判定。手冊原文:「DESTRUCTION IF:
// aggD * 25 / hit points >= random(100)」。roll 為呼叫端已擲好的 random(100)(1-100 含)。
func DamageSphericalFlyerDestroyed(aggD, hitPoints, roll int) bool {
	if hitPoints <= 0 {
		return true
	}
	return aggD*25/hitPoints >= roll
}

// DamageEngineExplosionPotential 手冊原文(Engine Explosion):「a damage potential of 5 times
// the maximum engine hit points (without Reinforced Hull). This damage value is tripled if the
// ship has a Quantum Detonator onboard.」「Contrary to other spherical weapons, the damage
// caused by a drive explosion dissipates linearly over range」未給出每格衰減率的精確數字,
// 不移植(TODO 待查證)。
func DamageEngineExplosionPotential(maxEngineHP int, quantumDetonator bool) int {
	potential := 5 * maxEngineHP
	if quantumDetonator {
		potential *= 3
	}
	return potential
}
