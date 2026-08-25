package gamedata

// MiniaturizationSpaceCategory 是 sub_6F11C 傳給 sub_6E60E 的 raw category。
type MiniaturizationSpaceCategory uint8

const (
	MiniSpaceGeneral MiniaturizationSpaceCategory = iota
	MiniSpaceTorpedoOrSpecial
	MiniSpaceFixed
)

// MiniaturizationSpaceCategoryForTech 依原版 typed weapon record 與五個船體裝置例外分類。
func MiniaturizationSpaceCategoryForTech(tech Technology) MiniaturizationSpaceCategory {
	if w, ok := OrigWeaponByTech(tech); ok {
		switch w.Cat {
		case WeaponCatBeam, WeaponCatMissile, WeaponCatBomb:
			return MiniSpaceGeneral
		case WeaponCatTorpedo, WeaponCatSpecial:
			return MiniSpaceTorpedoOrSpecial
		case WeaponCatFighterBay:
			return MiniSpaceFixed
		}
	}
	switch tech {
	case TECH_AUGMENTED_ENGINES, TECH_REINFORCED_HULL, TECH_EXTENDED_FUEL_TANKS,
		TECH_HEAVY_ARMOR, TECH_TROOP_PODS:
		return MiniSpaceFixed
	case TECH_NONE:
		return MiniSpaceGeneral
	default:
		// sub_6F11C 對一般船體特殊裝置（runtime type 9）保留初值 1。
		return MiniSpaceTorpedoOrSpecial
	}
}

// WeaponSpaceAtMiniLevelForCategory 對應 sub_6E60E 的三條佔格階梯。
func WeaponSpaceAtMiniLevelForCategory(base, level int, category MiniaturizationSpaceCategory) int {
	var perMille int
	switch category {
	case MiniSpaceGeneral:
		ladder := [...]int{1000, 800, 650, 500, 350}
		perMille = 250
		if level >= 0 && level < len(ladder) {
			perMille = ladder[level]
		}
	case MiniSpaceTorpedoOrSpecial:
		ladder := [...]int{1000, 800, 700, 600, 500}
		perMille = 400
		if level >= 0 && level < len(ladder) {
			perMille = ladder[level]
		}
	default:
		perMille = 1000
	}
	value := (base*perMille + 500) / 1000
	if value < 1 && base > 0 {
		return 1
	}
	return value
}
