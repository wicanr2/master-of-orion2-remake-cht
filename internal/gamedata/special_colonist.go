package gamedata

const (
	// PopulationRaceSlots 是原版 colony+0xB4／+0xC8 兩組陣列的槽數：
	// 0..7 為玩家帝國，8/9 為 Android／Natives。
	PopulationRaceSlots = 10
	AndroidColonistSlot = 8
	NativeColonistSlot  = 9
)

// SpecialColonistProduction 回傳原版非玩家人口 helper 的固定加成。
// Android：sub_DE0C6/sub_DED47/sub_DFE77 = +6/+3/+3；Natives = +4/+0/+0。
// 兩者在 sub_DDF2C 都直接回 4，故 gravityImmune=true。
func SpecialColonistProduction(slot int) (food, industry, research int, gravityImmune bool, ok bool) {
	switch slot {
	case AndroidColonistSlot:
		return 6, 3, 3, true, true
	case NativeColonistSlot:
		return 4, 0, 0, true, true
	default:
		return 0, 0, 0, false, false
	}
}

// RacePopulationCapacityDelta 回傳三項種族能力相對一般種族造成的人口上限差。
func RacePopulationCapacityDelta(size PlanetSize, climate PlanetClimate, aquatic, tolerant, subterranean bool) int {
	base := PlanetBasePopMax(size, climate)
	effective := climate
	if tolerant {
		effective = TERRAN
	} else {
		effective = RaceFoodClimate(climate, aquatic)
	}
	capacity := PlanetBasePopMax(size, effective)
	if subterranean {
		capacity += 2 * (int(size) + 1)
	}
	return capacity - base
}
