package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// WeaponUsesVariableMissileRack 回報是否為原版 raw type 1 的標準飛彈。
func WeaponUsesVariableMissileRack(name string) bool {
	return weaponKindByName(name) == WeaponKindMissile && !WeaponIsTorpedo(name)
}

// WeaponDefaultAmmo 回傳原版武器表的彈藥預設；查無時以 255 表示不受彈藥限制。
func WeaponDefaultAmmo(name string) int {
	// 事件怪物專用武器不在玩家可研究的 WeaponOptions，但仍使用同一原版武器表。
	for id, monsterName := range monsterWeaponNames {
		if monsterName == name {
			if w, ok := gamedata.OrigWeaponByID(id); ok {
				return w.Ammo
			}
		}
	}
	for _, c := range WeaponOptions {
		if c.Name != name {
			continue
		}
		if w, ok := gamedata.OrigWeaponByTech(c.UnlockTech); ok {
			return w.Ammo
		}
		break
	}
	return 255
}

// NormalizeWeaponAmmo 保留合法標準飛彈彈架；其餘武器固定回原版表預設。
func NormalizeWeaponAmmo(name string, ammo int) int {
	if WeaponUsesVariableMissileRack(name) {
		return gamedata.NormalizeMissileRackAmmo(ammo)
	}
	return WeaponDefaultAmmo(name)
}

func CycleWeaponAmmo(name string, ammo int) int {
	if !WeaponUsesVariableMissileRack(name) {
		return NormalizeWeaponAmmo(name, ammo)
	}
	return gamedata.CycleMissileRackAmmo(ammo)
}

func weaponAmmoCanFire(ammo int) bool { return ammo == 255 || ammo > 0 }

// prepareCombatantAmmo 只替沒有新欄位的舊建構端補一次預設值；ammoSet 後的 0
// 是真實耗盡狀態，必須保留到這場快速戰鬥結束。
func prepareCombatantAmmo(c *combatant) {
	if c.ammoSet {
		return
	}
	c.ammo = NormalizeWeaponAmmo(c.weaponName, c.ammo)
	c.ammoSet = true
}

func weaponAmmoSpend(ammo int) int {
	if ammo == 255 {
		return ammo
	}
	if ammo > 0 {
		return ammo - 1
	}
	return 0
}
