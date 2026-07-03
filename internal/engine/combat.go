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

// ShipCombatState 是戰鬥中一艘艦的可變狀態(裝甲/結構會被逐發削減)。
type ShipCombatState struct {
	StructureHP     int  // 結構點(歸零即摧毀)
	ArmorHP         int  // 裝甲點(先於結構被消耗)
	ShieldReduction int  // 護盾每次攻擊減傷(級別評等;本模型不衰減容量,採每擊固定減傷)
	HardShield      bool // Hard Shields
	BeamDefense     int  // 閃避(BD)
	Destroyed       bool
}

// ApplyBeamShot 對 ship 施加一發光束攻擊:呼叫 ResolveBeamShot 取傷害分配,削減 ship 的
// 裝甲/結構,結構歸零則標記 Destroyed。回傳該發解算結果。已摧毀的艦不再受擊(回 Hit=false)。
func ApplyBeamShot(atk BeamAttacker, ship *ShipCombatState, toHitRoll, dmgRoll int) BeamShotResult {
	if ship.Destroyed {
		return BeamShotResult{Hit: false}
	}
	tgt := BeamTarget{
		BeamDefense:     ship.BeamDefense,
		ShieldReduction: ship.ShieldReduction,
		HardShield:      ship.HardShield,
		ArmorHP:         ship.ArmorHP,
	}
	res := ResolveBeamShot(atk, tgt, toHitRoll, dmgRoll)
	if res.Hit {
		ship.ArmorHP -= res.DamageToArmor
		if ship.ArmorHP < 0 {
			ship.ArmorHP = 0
		}
		ship.StructureHP -= res.DamageToStructure
		if ship.StructureHP <= 0 {
			ship.StructureHP = 0
			ship.Destroyed = true
		}
	}
	return res
}

// ResolveVolley 讓開火方對同一目標連發 len(rolls) 發(每發一組 toHit/dmg 擲骰),
// 逐發套用並累積削減。回傳實際命中發數與是否擊毀目標。
// rolls 每列 [2]int{toHitRoll, dmgRoll};以外部注入 RNG 保持可重現。
func ResolveVolley(atk BeamAttacker, ship *ShipCombatState, rolls [][2]int) (hits int, destroyed bool) {
	for _, r := range rolls {
		res := ApplyBeamShot(atk, ship, r[0], r[1])
		if res.Hit {
			hits++
		}
		if ship.Destroyed {
			break // 已摧毀,停止
		}
	}
	return hits, ship.Destroyed
}
