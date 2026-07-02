package engine

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// combat.go 編排單發光束攻擊的完整解算:命中判定 → 傷害潛力(距離衰減)→ 內插傷害 →
// 護盾吸收 → 裝甲/結構分配。組合 gamedata 的 combat.go(命中)與 damage.go(傷害),
// RNG 擲骰以參數注入(toHitRoll/dmgRoll,1-100),使解算可重現、可單測。

// BeamAttacker 是開火方的單發光束參數。
type BeamAttacker struct {
	BeamAttack           int  // BA(已含電腦/掃描/艦員/首領/種族加成)
	DamageMin, DamageMax int  // 武器原始傷害潛力(未衰減)
	RangeLevel           int  // 距離級(gamedata.CombatRangeLevel 由格數換得)
	DoubledRangePenalty  bool // 內建 2x 距離懲罰改造(Fusion Beam/Plasma Cannon/Mauler)
	PDBonus              int  // 點防禦掛載對命中門檻的加成
}

// BeamTarget 是目標方的防禦參數。
type BeamTarget struct {
	BeamDefense     int  // BD(閃避)
	ShieldReduction int  // 護盾每次攻擊減傷點數(級別評等)
	HardShield      bool // Hard Shields(額外減傷、使 Shield Piercing 失效)
	ArmorHP         int  // 剩餘裝甲點數
}

// BeamShotResult 是一發攻擊的解算結果。
type BeamShotResult struct {
	Hit               bool
	DamageToShield    int // 被護盾吸收
	DamageToArmor     int // 打在裝甲上
	DamageToStructure int // 穿透到結構
}

// ResolveBeamShot 解算一發光束攻擊。toHitRoll/dmgRoll 為注入的 1-100 擲骰。
func ResolveBeamShot(atk BeamAttacker, tgt BeamTarget, toHitRoll, dmgRoll int) BeamShotResult {
	netAttack := atk.BeamAttack - tgt.BeamDefense
	rangePenalty := gamedata.CombatRangeLevelPenaltyDoubled(atk.RangeLevel, atk.DoubledRangePenalty)
	hitThreshold := gamedata.CombatHitThreshold(rangePenalty, atk.PDBonus)

	if !gamedata.CombatClassicToHit(toHitRoll, netAttack, hitThreshold) {
		return BeamShotResult{Hit: false}
	}

	// 命中:傷害潛力先套距離衰減,再依命中裕度內插實際傷害。
	dMin, dMax := gamedata.DamageApplyDissipation(atk.DamageMin, atk.DamageMax, atk.RangeLevel)
	raw := gamedata.DamageForHit(dMin, dMax, dmgRoll, netAttack, hitThreshold)

	// 護盾吸收 → 裝甲/結構分配。
	afterShield := gamedata.DamageAfterShield(raw, tgt.ShieldReduction, tgt.HardShield, false)
	shieldAbsorbed := raw - afterShield
	toArmor, toStructure, _ := gamedata.DamageApplyArmor(afterShield, tgt.ArmorHP, false, false)

	return BeamShotResult{
		Hit:               true,
		DamageToShield:    shieldAbsorbed,
		DamageToArmor:     toArmor,
		DamageToStructure: toStructure,
	}
}
