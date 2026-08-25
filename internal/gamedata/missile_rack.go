package gamedata

// MissileRackAmmoOptions 是原版標準飛彈可選彈架容量。
var MissileRackAmmoOptions = [...]int{2, 5, 10, 15, 20}

// MissileRackBaseValue 對應 Weapon_Cost_／Weapon_Space_ 的共同查表值。
func MissileRackBaseValue(ammo int) (int, bool) {
	switch ammo {
	case 2:
		return 10, true
	case 5:
		return 20, true
	case 10:
		return 30, true
	case 15:
		return 35, true
	case 20:
		return 40, true
	default:
		return 0, false
	}
}

// NormalizeMissileRackAmmo 將任意值收斂到合法容量；零值／非法值回原版預設 5。
func NormalizeMissileRackAmmo(ammo int) int {
	if _, ok := MissileRackBaseValue(ammo); ok {
		return ammo
	}
	return 5
}

// CycleMissileRackAmmo 依原版容量順序循環。
func CycleMissileRackAmmo(ammo int) int {
	ammo = NormalizeMissileRackAmmo(ammo)
	for i, value := range MissileRackAmmoOptions {
		if value == ammo {
			return MissileRackAmmoOptions[(i+1)%len(MissileRackAmmoOptions)]
		}
	}
	return 5
}
