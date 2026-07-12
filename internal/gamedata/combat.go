package gamedata

// 光束武器命中(to-hit)相關公式,逐一移植 MOO2 patch 1.5 官方手冊
// (moo2_patch1.5/MANUAL_150.html,"Notes on Beam Weapon Mechanics" 章節)。
// 命名一律加 Combat/Beam 前綴,避免與 formulas.go 既有的 BeamOffense/BeamDefense/
// ShipCrewOffenseBonus 撞名;也不定義通用 helper(如 clamp),以免與其他並行檔衝突。
//
// 手冊原文引用見各函式註解。手冊沒有清楚給出精確數字的部分(如 Regular 之外,
// 若干中間推導細節)一律不推測、不外插,只依表格/公式明載的數字實作。

// combatRegularRangeLevelRaw 依「1 range unit = 3 squares」換算的未夾限 range level。
// 手冊原文:「One range unit equals 3 squares. At range 0 two opponents share the
// same square... At range 1 opponents are adjacent to each other or have 1-2
// squares between them.」對照 Regular (sq) 表:0→0、1-3→1、4-6→2...22-24→8,
// 公式為 sq<=0 時 0,否則 ceil(sq/3)。此處不夾限在 8,供 CombatRangeLevelHeavy
// 內部换算大距離時使用(Heavy 換算需要參考「若未夾限時」的原始 level)。
func combatRegularRangeLevelRaw(squares int) int {
	if squares <= 0 {
		return 0
	}
	return (squares + 2) / 3
}

// CombatRangeLevel 依 Regular 掛載的距離(格數)查出 range level(0-8),
// 對照手冊 Range Penalty 表的「Regular (sq)」列:
//
//	Range   0  1  2  3  4   5   6   7   8
//	Regular 0 1-3 4-6 7-9 10-12 13-15 16-18 19-21 22-24
//
// 超過 22-24(level 8)一律夾在 8(手冊未列更遠距離,以最高懲罰處理)。
func CombatRangeLevel(squares int) int {
	lv := combatRegularRangeLevelRaw(squares)
	if lv > 8 {
		return 8
	}
	return lv
}

// CombatRangeLevelPointDefense 依 Point Defense 掛載的距離(格數)查出 range level。
//
// 手冊原文:「Point Defense weapons get a penalty as if range is doubled」,
// 且手冊隨附的 PD (sq) 列只在偶數 level 有值、奇數 level 全空:
//
//	Range 0  1   2   3   4   5   6   7   8
//	PD    0  -  1-3  -  4-6  -  7-9  -  10-12
//
// 對照 Regular (sq) 列可證實「doubled」指的是 range level 加倍(PD level = 2 *
// Regular level),而非距離格數加倍再查表(格數加倍再查表會得到不同結果,
// 與手冊 PD 列的空格位置對不上)。level 上限同樣夾在 8。
func CombatRangeLevelPointDefense(squares int) int {
	lv := combatRegularRangeLevelRaw(squares) * 2
	if lv > 8 {
		return 8
	}
	return lv
}

// CombatRangeLevelHeavy 依 Heavy 掛載的距離(格數)查出 range level。
//
// 手冊原文:「for Heavy mount weapons the actual range is halved (and rounded
// down)」,對照手冊隨附的 Hv (sq) 列:
//
//	Range 0   1   2    3    4    5    6    7    8
//	Hv   0-3 4-9 10-15 16-21 22-27 28-33 34-39 40-45 46-51
//
// 逐一代入驗證後,對應的是 Heavy level = floor(RegularLevelRaw(sq) / 2)
// (未夾限的 Regular level 先算出、再整除 2),而非距離格數先減半再查 Regular
// 表(該讀法在多處邊界對不上手冊列出的區間)。level 上限同樣夾在 8。
func CombatRangeLevelHeavy(squares int) int {
	lv := combatRegularRangeLevelRaw(squares) / 2
	if lv > 8 {
		return 8
	}
	return lv
}

// combatRangeLevelPenaltyTable 手冊 Range Penalty 表(Range level 0-8 → Penalty)。
var combatRangeLevelPenaltyTable = [9]int{0, 0, 10, 20, 30, 40, 55, 70, 85}

// CombatRangeLevelPenalty 依 range level(0-8,由 CombatRangeLevel /
// CombatRangeLevelPointDefense / CombatRangeLevelHeavy 算得)查出 to-hit 懲罰值。
// 手冊 Range Penalty 表:
//
//	Range   0 1  2  3  4  5  6  7  8
//	Penalty 0 0 10 20 30 40 55 70 85
//
// level 超出 0-8 一律夾限到最近端點。
func CombatRangeLevelPenalty(level int) int {
	if level < 0 {
		level = 0
	}
	if level > 8 {
		level = 8
	}
	return combatRangeLevelPenaltyTable[level]
}

// CombatRangeLevelPenaltyDoubled 套用「Classic Fusion Beam、Plasma Cannon、
// Mauler Device 有內建 2x range to-hit penalty(weapon flag #1)」的加倍規則。
// 手冊原文:「This mod doubles the calculated penalty. So after all Hv/Reg/PD
// calculations are done ... the result is doubled.」
func CombatRangeLevelPenaltyDoubled(level int, doubled bool) int {
	p := CombatRangeLevelPenalty(level)
	if doubled {
		p *= 2
	}
	return p
}

// CombatHitThreshold 算出 Classic / Alternative to-hit 公式共用的 hit_threshold。
// 手冊原文(Classic Chance to Hit Formula):
//
//	[3b] hit_threshold = min(40 + range_penalty* - PD_bonus; 95)
//	     *doubled for inherent 2x range to-hit penalty mod
//
// rangePenalty 已包含(或不包含)2x mod 加倍,由呼叫端先用
// CombatRangeLevelPenaltyDoubled 算好再傳入。pdBonus 為 Point Defense 加成
// (weapon flag PD_bonus,手冊未在本節給出精確數字,由呼叫端提供)。
//
// 已用手冊 Damage Potential 範例驗證:
//   - range 23 sq(level 8,penalty 85,pdBonus 0)→ min(40+85-0,95) = 95
//   - range 11 sq(level 4,penalty 30,pdBonus 0)→ min(40+30-0,95) = 70
//
// 兩者皆與手冊 Examples 段落算出的 hit_threshold 一致。
func CombatHitThreshold(rangePenalty, pdBonus int) int {
	threshold := 40 + rangePenalty - pdBonus
	if threshold > 95 {
		return 95
	}
	return threshold
}

// CombatClassicToHit 實作手冊「Classic Chance to Hit Formula」:
//
//	HIT IF: (with max dmg)
//	    [1] random(100) > 95
//	ELSE HIT IF: (with max dmg)
//	    [2] BA+CO-AF-BD >= 99
//	ELSE HIT IF: (real penalty affects dmg distribution)
//	    [3a] random(100) + BA+CO-AF-BD >= hit_threshold
//
// roll 為呼叫端已擲出的 random(100)(值域 1-100,含);netAttack =
// BA+CO-AF-BD(Beam Attack + Continuous Fire 命中加成 - Auto-Fire 命中懲罰 -
// 目標 Beam Defense,由呼叫端算好);hitThreshold 由 CombatHitThreshold 算得。
//
// [2026-07-11 訂正] 先前這裡把 AF 誤寫成「Point Defense 目標的 AF」(暗示與目標的
// Point Defense 有關)。移植武器改造(mod)系統時查手冊(GAME_MANUAL.pdf p.115)+
// docs/tech/community-mechanics-findings.md 引用的社群拆解交叉核對後確認:AF 是
// Auto-Fire mod 本身的自我懲罰(攻方武器掛 Auto-Fire,每次射擊 -20 命中),與 Point
// Defense、與「目標」都無關,是攻方 netAttack 裡的一項,不是目標的屬性。CO 同理是
// Continuous Fire mod 給攻方自己的 +25 命中加成。實作見
// gamedata.WeaponModNetAttackBonus(weapon_mods.go)。
func CombatClassicToHit(roll, netAttack, hitThreshold int) bool {
	if roll > 95 {
		return true
	}
	if netAttack >= 99 {
		return true
	}
	return roll+netAttack >= hitThreshold
}

// CombatAlternativeToHit 實作手冊「1.50 Alternative Chance to Hit Formula
// (Optional)」(simplified_beam_formula = 1 時啟用):
//
//	HIT IF: (with max dmg)
//	    [1] random(100) > 95
//	ELSE HIT IF: (with max dmg)
//	    [2] BA+CO-AF-BD - range_penalty* + PD_bonus >= 99
//	ELSE HIT IF: (real penalty affects dmg distribution)
//	    [3] BA+CO-AF-BD - range_penalty* + PD_bonus + random(100) >= 40
//
// roll 為 random(100);netAttack = BA+CO-AF-BD;rangePenalty 已包含(或不含)
// 2x mod 加倍(呼叫端用 CombatRangeLevelPenaltyDoubled 算好);pdBonus 同
// CombatHitThreshold。與 Classic 公式不同,此式讓距離懲罰「一致地」影響命中率,
// 不會被 [2] 的高 BA/低 BD 情境完全蓋掉。
func CombatAlternativeToHit(roll, netAttack, rangePenalty, pdBonus int) bool {
	adjusted := netAttack - rangePenalty + pdBonus
	if roll > 95 {
		return true
	}
	if adjusted >= 99 {
		return true
	}
	return adjusted+roll >= 40
}

// 飛彈 Beam Defense(Speed = BaseSpeed + 2*(FTLlevel-1) + FastBonus、
// MissileBonus 表、MissileBeamDefense)已在 missile.go 移植過(MissileSpeed /
// MissileWarheadBonus / MissileBeamDefense),此處不重複定義。

// combatFighterTDBonus 手冊原文:「TransDimensionalBonus is 4 for all fighter
// types.」(戰機版的 FastBonus 對應項)。
const combatFighterTDBonus = 4

// 手冊 Fighter 表格的 BaseSpeed 欄位。
const (
	CombatFighterBaseSpeedInterceptor    = 10
	CombatFighterBaseSpeedAssaultShuttle = 6
	CombatFighterBaseSpeedBomber         = 8
	CombatFighterBaseSpeedHeavyFighter   = 8
)

// CombatFighterSpeed 手冊原文:「Speed = BaseSpeed of Fighter + 2 *
// (FTLlevel - 1) + TDBonus」。baseSpeed 用上列 CombatFighterBaseSpeed* 常數。
func CombatFighterSpeed(baseSpeed, ftlLevel int) int {
	return baseSpeed + 2*(ftlLevel-1) + combatFighterTDBonus
}

// CombatFighterBeamDefense 手冊原文:「Fighter Beam Defense is calculated as
// follows: 5 * Speed + RacialShipDefenseBonus + FighterPilotBonus +
// HelmsmanBonus (in 1.50 only)」。speed 由 CombatFighterSpeed 算得;三項加成
// 手冊未在本節給出精確數字(來自種族/飛行員/艦隊領袖等其他系統),由呼叫端
// 提供,此處不臆造。
func CombatFighterBeamDefense(speed, racialShipDefenseBonus, fighterPilotBonus, helmsmanBonus int) int {
	return 5*speed + racialShipDefenseBonus + fighterPilotBonus + helmsmanBonus
}

// --- 戰機庫(Fighter Bay)對抽象快速戰鬥的貢獻 ---

// FighterInterceptorSquadron 是一個攔截機戰機庫每次出擊的戰機數(手冊 GM p.127「出擊數」欄:
// 攔截機 4)。first-version 戰機庫以最常見的攔截機隊建模。此值為手冊硬數字。
const FighterInterceptorSquadron = 4

// 每攔截機對抽象戰力的攻擊/HP 貢獻。⚠ remake 近似,非手冊定值:手冊給攔截機「1 光束」武裝
// (實際傷害隨當局光束科技,非固定數)與血量 2-20(隨裝甲級),此處取低階代表值。中隊規模(4)
// 本身是手冊硬數字(見上)。remake 的 ResolveBattle 是艦級抽象結算、不逐戰機模擬,故以「母艦
// 戰力加成」承接一整隊戰機的火力,而非獨立 combatant。
const (
	fighterInterceptorAttackApprox = 3
	fighterInterceptorHPApprox     = 4
)

// FighterBayCombatContribution 回傳一個攔截機戰機庫在 ResolveBattle 快速結算中對母艦戰力的
// 加成(攻擊, HP):中隊 4 架(手冊 GM p.127),每架近似攻 3 / HP 4 → 母艦 +12 攻、+16 HP。
func FighterBayCombatContribution() (atk, hp int) {
	return FighterInterceptorSquadron * fighterInterceptorAttackApprox,
		FighterInterceptorSquadron * fighterInterceptorHPApprox
}

// FighterHeavySquadron 是一個重戰機庫每次出擊的重戰機數(手冊 GM p.127「出擊數」欄:重戰機 2)。
// 手冊硬數字。重戰機武裝為「1 光束 + 1 炸彈」,對艦火力較攔截機強、數量較少。
const FighterHeavySquadron = 2

// 每重戰機對抽象戰力的攻擊/HP 貢獻。⚠ remake 近似,非手冊定值(同攔截機的理由:手冊給武裝類型
// 與血量範圍 5-50 隨裝甲級,非固定戰力值),取較攔截機高的低階代表值以反映重戰機較強。
const (
	fighterHeavyAttackApprox = 8
	fighterHeavyHPApprox     = 8
)

// FighterHeavyBayCombatContribution 回傳一個重戰機庫對母艦戰力的加成(攻擊, HP):中隊 2 架
// (手冊 GM p.127),每架近似攻 8 / HP 8 → 母艦 +16 攻、+16 HP(較攔截機庫 +12 攻略強)。
func FighterHeavyBayCombatContribution() (atk, hp int) {
	return FighterHeavySquadron * fighterHeavyAttackApprox,
		FighterHeavySquadron * fighterHeavyHPApprox
}
