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

// ResolveShot 用 gamedata 真公式解算一次射擊。
//   - netAttack = 攻方 Beam Attack(含命中加成) − 守方防禦(AF+BD)。
//   - rangeSquares = 曼哈頓/格數距離(→射程等級→命中懲罰)。
//   - roll = 呼叫端擲出的 random(1..100)(由戰鬥 RNG 提供,保持可重現)。
//
// 流程:射程等級→射程懲罰→命中門檻→CombatClassicToHit→DamageForHit→DamageAfterShield→DamageApplyArmor。
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
