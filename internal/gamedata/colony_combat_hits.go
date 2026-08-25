package gamedata

// OriginalColonyCombatHits 依原版 Get_Colony_Hits_ @ 0x42371 計算快速戰鬥中的
// 殖民地本體耐久。人口、士兵、戰車各加一；每棟一般建築加 40。Battlestation、
// Star Base、Star Fortress 在快速戰鬥中另有戰鬥者，因此 raw ID 8/40/41 不重複計入。
//
// 來源與證據等級見 docs/re/colony-hits-audit-20260824.md。負值、重複 ID 與範圍外 ID
// 是 remake 的型別安全邊界；合法原版資料不會產生這些輸入。
func OriginalColonyCombatHits(population, soldiers, tanks int, rawBuildingIDs []int) int {
	total := nonNegativeColonyHits(population) + nonNegativeColonyHits(soldiers) + nonNegativeColonyHits(tanks)
	seen := make(map[int]struct{}, len(rawBuildingIDs))
	for _, id := range rawBuildingIDs {
		if id < 0 || id >= len(OriginalBuildingTable) || id == 8 || id == 40 || id == 41 {
			continue
		}
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		total += 40
	}
	return total
}

func nonNegativeColonyHits(value int) int {
	if value < 0 {
		return 0
	}
	return value
}
