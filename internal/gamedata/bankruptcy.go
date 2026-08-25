package gamedata

// 原版 off_17EB3D 的 49 筆建築 raw category。索引就是原版建築 ID；
// 來源與篩選證據見 docs/re/player-maintenance-audit-20260825.md。
var originalBuildingRawCategories = [...]uint8{
	6, 4, 0, 7, 1, 2, 1, 3, 7, 1, 0, 0, 0, 0, 7, 3, 1,
	1, 5, 5, 3, 0, 0, 2, 1, 1, 2, 2, 0, 1, 2, 0, 0, 1, 1,
	0, 3, 0, 1, 0, 7, 7, 0, 0, 0, 0, 0, 0, 0,
}

// OriginalBuildingRawCategory 回傳原版建築表的 raw category；它不是 remake UI 的
// BuildingCategory，不能互換。未知 ID 回 ok=false。
func OriginalBuildingRawCategory(id int) (category int, ok bool) {
	if id < 0 || id >= len(originalBuildingRawCategories) {
		return 0, false
	}
	return int(originalBuildingRawCategories[id]), true
}

// BankruptcyBuildingEligible 實作 sub_EDAE2、sub_EDB1D 與 0xEDB2D 的三輪篩選。
func BankruptcyBuildingEligible(stage, id, productionCost int) bool {
	category, ok := OriginalBuildingRawCategory(id)
	if !ok {
		return false
	}
	switch stage {
	case 0:
		return productionCost > 0 && category != 1 && category != 3 && id != 9
	case 1:
		return productionCost > 0
	case 2:
		return id != 9
	default:
		return false
	}
}

// BankruptcyBuildingScore 是 sub_ED9EC 的已閉合部分。值越低越早出售。
// governmentRaw 與原版同序；hasBuildingID 用來處理營房取代鏈。
func BankruptcyBuildingScore(id, productionCost, governmentRaw int, hasBuildingID func(int) bool) int {
	score := productionCost
	switch id {
	case 7, 15:
		score *= 8
	case 12, 20, 31:
		score *= 7
	case 21:
		score *= 6
	case 25:
		score *= 4
	case 2:
		if governmentRaw >= 0 && governmentRaw <= 3 {
			score *= 8
		}
	case 22:
		if governmentRaw >= 0 && governmentRaw <= 3 && (hasBuildingID == nil || !hasBuildingID(2)) {
			score *= 8
		}
	}
	return score
}
