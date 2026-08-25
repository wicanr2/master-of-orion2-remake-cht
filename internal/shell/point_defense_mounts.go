package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// TacticalPointDefenseMount 是一個可自動迎擊的 typed 武器槽。
// Slot=-1 表示舊單槽相容欄位；Count 是同槽仍在工作的武器門數。
type TacticalPointDefenseMount struct {
	Slot          int
	Count         int
	WeaponName    string
	BeamDamageMax int
	BeamMods      []gamedata.WeaponModCode
}

func pointDefenseMountsFor(
	weaponName string,
	mods []string,
	weaponMax int,
	mounts []ShipWeaponMount,
	spentSlots []bool,
	legacySpent bool,
) []TacticalPointDefenseMount {
	if len(mounts) == 0 {
		beamMods := WeaponModCodesForWeapon(weaponName, mods)
		if legacySpent || !PointDefenseCanFire(weaponName, beamMods) {
			return nil
		}
		return []TacticalPointDefenseMount{{
			Slot: -1, Count: 1, WeaponName: weaponName,
			BeamDamageMax: weaponMax, BeamMods: beamMods,
		}}
	}
	out := make([]TacticalPointDefenseMount, 0, len(mounts))
	for i := range mounts {
		mount := mounts[i]
		if mount.WorkingCount <= 0 || i < len(spentSlots) && spentSlots[i] {
			continue
		}
		beamMods := WeaponModCodesForWeapon(mount.Name, mount.Mods)
		if !PointDefenseCanFire(mount.Name, beamMods) {
			continue
		}
		damage := mount.Attack
		if damage <= 0 {
			damage = weaponMax
		}
		out = append(out, TacticalPointDefenseMount{
			Slot: i, Count: mount.WorkingCount, WeaponName: mount.Name,
			BeamDamageMax: damage, BeamMods: beamMods,
		})
	}
	return out
}

// AvailableTacticalPointDefenseMounts 依原版槽序回傳本回合尚未自動開火的 PD。
// 刻意不讀 WeaponModes：原版說明明定紅色關閉的 PD 仍會迎擊來襲飛彈。
func AvailableTacticalPointDefenseMounts(ship *CombatShip) []TacticalPointDefenseMount {
	if ship == nil {
		return nil
	}
	if len(ship.WeaponMounts) > 0 && len(ship.PointDefenseSpentSlots) < len(ship.WeaponMounts) {
		ship.PointDefenseSpentSlots = append(ship.PointDefenseSpentSlots,
			make([]bool, len(ship.WeaponMounts)-len(ship.PointDefenseSpentSlots))...)
	}
	return pointDefenseMountsFor(ship.WeaponName, ship.Mods, ship.WeaponMax,
		ship.WeaponMounts, ship.PointDefenseSpentSlots, ship.PointDefenseSpent)
}

// MarkTacticalPointDefenseMountSpent 將一個槽標成本回合已自動開火。
func MarkTacticalPointDefenseMountSpent(ship *CombatShip, slot int) {
	if ship == nil {
		return
	}
	ship.PointDefenseSpent = true // 保留舊 UI／測試可觀測的聚合旗標。
	if slot < 0 {
		return
	}
	if len(ship.PointDefenseSpentSlots) < len(ship.WeaponMounts) {
		ship.PointDefenseSpentSlots = append(ship.PointDefenseSpentSlots,
			make([]bool, len(ship.WeaponMounts)-len(ship.PointDefenseSpentSlots))...)
	}
	if slot < len(ship.PointDefenseSpentSlots) {
		ship.PointDefenseSpentSlots[slot] = true
	}
}

// ResetTacticalPointDefenseSpent 在回合交界清除逐槽與相容旗標；攔截傷害餘數另行保留。
func ResetTacticalPointDefenseSpent(ship *CombatShip) {
	if ship == nil {
		return
	}
	ship.PointDefenseSpent = false
	for i := range ship.PointDefenseSpentSlots {
		ship.PointDefenseSpentSlots[i] = false
	}
}
