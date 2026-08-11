package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// energy_absorber.go:能量吸收器——被打時把 1/4 的潛在傷害存起來,之後射回去。
//
// 手冊逐字:
//
//	One-quarter of all the potential damage that reaches a ship is diverted to and stored
//	in the Energy Absorber. That ship can fire this stored energy at an enemy ship —
//	automatically hitting it (unless the target has a Displacement Device). Damage done by
//	an energy absorber is reduced by range like that of a beam weapon. A cloaked ship will
//	not decloak from firing its stored energy.
//	● In games with Ship Initiative turned on, the stored energy accumulates until it is
//	  fired at an enemy ship.
//	● In games with Ship Initiative off, the damage energy is stored for only one combat
//	  round; if the ship doesn't use it, it is lost.
//
// ============ 兩個要講清楚的取捨 ============
//
//  1. **主動權一律當成開著。** remake 的齊射次序就是照手冊的主動權公式排的
//     (gamedata.CombatInitiative,第 69 項(戰鬥速度與引擎階)),沒有「關掉主動權」這個選項。
//     所以走上面那條「accumulates until it is fired」,儲能不會每回合歸零。
//
//  2. **儲能是「下一次開火時自動射出」,不是玩家的另一個按鈕。** 原版讓玩家自己挑時機;
//     remake 的格子戰術一次點擊就是「這艘船對那艘船開火」,沒有第二個動作可以掛。
//     做成自動射出的代價是玩家不能存著等大魚,好處是它真的會生效——**這是建模取捨,
//     不是抄漏**。要改成玩家控制,得先有「選擇要用哪個系統」的動作層。
//
// 手冊那句「A cloaked ship will not decloak from firing its stored energy」在這個模型下
// 仍然要守:射儲能的那一發不呼叫 CloakOnFire(見 cloak.go 的警告)。

// EnergyAbsorberAbsorb 把一次攻擊的潛在傷害轉存進防守方的吸收器。
//
// potentialDamage 取「已命中、尚未扣護盾與裝甲」那一刻的值,理由見
// gamedata.EnergyAbsorberStored 的註解(手冊用的詞是 reaches 不是 penetrates)。
// 沒裝吸收器就什麼都不做。
func EnergyAbsorberAbsorb(sh *CombatShip, potentialDamage int) {
	if !sh.EnergyAbsorber {
		return
	}
	sh.StoredEnergy += gamedata.EnergyAbsorberStored(potentialDamage)
}

// EnergyAbsorberRelease 取出並清空儲能,回傳這一發實際打到目標結構的傷害。
//
// 規則(逐句對應手冊):
//   - 「automatically hitting it」→ 不擲命中骰。
//   - 「unless the target has a Displacement Device」→ 目標有位移裝置就照它的 30% 判定
//     (與飛彈那一側同一個常數,手冊那句寫的是「regardless of any other equipment or
//     situation」,能量吸收器不是例外)。
//   - 「reduced by range like that of a beam weapon」→ 套 DamageDissipationPenalty,
//     與光束同一張表。
//   - 手冊沒說它繞過護盾或裝甲,所以照一般順序扣。
//
// displacementRoll 只在目標有位移裝置時才會被讀(**沒裝就不要擲**,否則整條亂數流位移)。
func EnergyAbsorberRelease(sh *CombatShip, tgt *CombatShip, rangeSquares int, displacementRoll int) ShotResult {
	return energyAbsorberRelease(sh, tgt, rangeSquares, displacementRoll, tgt.ShieldReduction)
}

// EnergyAbsorberReleaseAtFacing 是格子戰術在知道攻擊方向後的入口；保留上面的
// 舊入口供純規則測試與其他呼叫端使用。
func EnergyAbsorberReleaseAtFacing(sh *CombatShip, tgt *CombatShip, rangeSquares, displacementRoll, shieldReduction int) ShotResult {
	return energyAbsorberRelease(sh, tgt, rangeSquares, displacementRoll, shieldReduction)
}

func energyAbsorberRelease(sh *CombatShip, tgt *CombatShip, rangeSquares int, displacementRoll int, shieldReduction int) ShotResult {
	stored := sh.StoredEnergy
	sh.StoredEnergy = 0
	if stored <= 0 {
		return ShotResult{Hit: false, RemainingArmorHP: tgt.ArmorHP}
	}
	if tgt.HasDisplacement && displacementRoll <= gamedata.MissileDisplacementDeviceMissChance {
		return ShotResult{Hit: false, RemainingArmorHP: tgt.ArmorHP}
	}
	level := gamedata.CombatRangeLevel(rangeSquares)
	dmg := stored * (100 - gamedata.DamageDissipationPenalty(level)) / 100
	shieldDamage := gamedata.DamageShieldAbsorbed(dmg, shieldReduction)
	dmg = gamedata.DamageAfterShield(dmg, shieldReduction, tgt.HardShield, false)
	_, toStruct, remArmor := gamedata.DamageApplyArmor(dmg, tgt.ArmorHP, false, tgt.APNegated)
	return ShotResult{Hit: true, DamageToStructure: toStruct, RemainingArmorHP: remArmor, ShieldDamage: shieldDamage}
}
