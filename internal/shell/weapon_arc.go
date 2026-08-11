package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// weapon_arc.go:把原版 WeaponArc 值接到重製的單武器艦艇設計模型。
//
// 原版每艘船最多保存八個 ShipWeapon，每個武器各自有 arc；重製目前仍是
// 每艘 Ship 一個 Weapon 欄位，所以本輪把單一掛載的 arc 接完整，並保留
// 未來擴成武器槽時可直接搬用的 gamedata 原始值。

var weaponArcOptions = []gamedata.WeaponArc{
	gamedata.ARC_FWD,
	gamedata.ARC_FWD_EXT,
	gamedata.ARC_BACK_EXT,
	gamedata.ARC_BACK,
	gamedata.ARC_360,
}

// DefaultWeaponArc 回傳新設計的預設火線角。原版資料／百科把飛彈與魚雷標成
// 360 度全向；其他目前可設計的武器先以前向基準弧開始。
func DefaultWeaponArc(weaponName string) gamedata.WeaponArc {
	if weaponKindByName(weaponName) == WeaponKindMissile {
		return gamedata.ARC_360
	}
	return gamedata.ARC_FWD
}

// WeaponArcOptionsForWeapon 回傳設計畫面可選的火線角。
// 飛彈／魚雷沒有四向選單，固定保存 360 度；其餘目前可設計武器可選五種弧。
func WeaponArcOptionsForWeapon(weaponName string) []gamedata.WeaponArc {
	if weaponKindByName(weaponName) == WeaponKindMissile {
		return []gamedata.WeaponArc{gamedata.ARC_360}
	}
	out := make([]gamedata.WeaponArc, len(weaponArcOptions))
	copy(out, weaponArcOptions)
	return out
}

// NormalizeWeaponArc 把舊 JSON 的零值、無效值或換武器後不適用的弧修正成合法值。
// 這只負責資料邊界，不等同於戰術射擊方向判定。
func NormalizeWeaponArc(weaponName string, arc gamedata.WeaponArc) gamedata.WeaponArc {
	options := WeaponArcOptionsForWeapon(weaponName)
	for _, allowed := range options {
		if arc == allowed {
			return arc
		}
	}
	return DefaultWeaponArc(weaponName)
}

// WeaponArcLabelZH／EN 是艦艇設計畫面的火線角標籤。
func WeaponArcLabelZH(arc gamedata.WeaponArc) string {
	switch arc {
	case gamedata.ARC_FWD:
		return "前向(FWD)"
	case gamedata.ARC_FWD_EXT:
		return "前向延伸(FWD EXT)"
	case gamedata.ARC_BACK_EXT:
		return "後向延伸(BACK EXT)"
	case gamedata.ARC_BACK:
		return "後向(BACK)"
	case gamedata.ARC_360, gamedata.ARC_MONSTER_360:
		return "全向(360)"
	default:
		return "前向(FWD)"
	}
}

func WeaponArcLabelEN(arc gamedata.WeaponArc) string {
	switch arc {
	case gamedata.ARC_FWD:
		return "Forward (FWD)"
	case gamedata.ARC_FWD_EXT:
		return "Forward Extended (FWD EXT)"
	case gamedata.ARC_BACK_EXT:
		return "Back Extended (BACK EXT)"
	case gamedata.ARC_BACK:
		return "Back (BACK)"
	case gamedata.ARC_360, gamedata.ARC_MONSTER_360:
		return "360 Degree (FULL)"
	default:
		return "Forward (FWD)"
	}
}

func cycleWeaponArc(weaponName string, current gamedata.WeaponArc) gamedata.WeaponArc {
	options := WeaponArcOptionsForWeapon(weaponName)
	if len(options) == 0 {
		return DefaultWeaponArc(weaponName)
	}
	current = NormalizeWeaponArc(weaponName, current)
	for i, arc := range options {
		if arc == current {
			return options[(i+1)%len(options)]
		}
	}
	return options[0]
}

// CycleWeaponArc 供設計畫面循環切換目前武器可用的火線角。
func CycleWeaponArc(weaponName string, current gamedata.WeaponArc) gamedata.WeaponArc {
	return cycleWeaponArc(weaponName, current)
}
