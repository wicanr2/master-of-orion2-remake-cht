package shell

// TacticalWeaponMode 是戰術戰鬥期間的逐槽開火狀態；不寫回持久艦艇設計。
type TacticalWeaponMode uint8

const (
	TacticalWeaponReady TacticalWeaponMode = iota
	TacticalWeaponStandby
	TacticalWeaponOff
)

func NewTacticalWeaponModes(mounts []ShipWeaponMount) []TacticalWeaponMode {
	n := len(mounts)
	if n == 0 {
		n = 1 // 舊單槽艦相容模式
	}
	return make([]TacticalWeaponMode, n)
}
