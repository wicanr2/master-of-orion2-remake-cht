package gamedata

import "sort"

const (
	AdvancedCivilizationCandidateLimit = 360
	AdvancedCivilizationBalanceLimit   = 20
)

// AdvancedCivilizationCandidate 是 Choose_Adv_Civ_Planets_ 所需的 typed 候選投影。
// Planet 是全局行星索引；Distance／Worth 已由星圖與行星估值 producer 算好。
type AdvancedCivilizationCandidate struct {
	Planet   int
	Distance int
	Worth    int
}

// AdvancedCivilizationChoose 依原版的隨機化玩家順序做 round-robin 分配。
// 候選先按 worth 降冪、同分按原輸入順序；同一行星只能被一名玩家取得。
func AdvancedCivilizationChoose(candidates [][]AdvancedCivilizationCandidate, playerOrder []int,
	quota, difficulty, maxDistance int) [][]int {
	out := make([][]int, len(candidates))
	if quota <= 0 {
		return out
	}
	limit := difficulty + 9
	if d := maxDistance / 10; d > limit {
		limit = d
	}
	lists := make([][]AdvancedCivilizationCandidate, len(candidates))
	for player := range candidates {
		lists[player] = append([]AdvancedCivilizationCandidate(nil), candidates[player]...)
		if len(lists[player]) > AdvancedCivilizationCandidateLimit {
			lists[player] = lists[player][:AdvancedCivilizationCandidateLimit]
		}
		sort.SliceStable(lists[player], func(i, j int) bool { return lists[player][i].Worth > lists[player][j].Worth })
	}
	used := map[int]bool{}
	for {
		progress := false
		for _, player := range playerOrder {
			if player < 0 || player >= len(lists) || len(out[player]) >= quota || len(out[player]) >= AdvancedCivilizationBalanceLimit {
				continue
			}
			for len(lists[player]) > 0 {
				c := lists[player][0]
				lists[player] = lists[player][1:]
				if c.Planet < 0 || c.Distance > limit || used[c.Planet] {
					continue
				}
				out[player] = append(out[player], c.Planet)
				used[c.Planet] = true
				progress = true
				break
			}
		}
		if !progress {
			break
		}
	}
	return out
}

// AdvancedCivilizationExtraPlanetQuota 對應 Num_Adv_Civ_Planets_ @ 0x62BB7。
// 全程使用整數除法；最後減一是排除每位玩家已存在的母星。
func AdvancedCivilizationExtraPlanetQuota(starCount, playerCount int) int {
	if starCount < 0 || playerCount <= 0 {
		return 0
	}
	quota := ((starCount / 2) * 10) / playerCount
	quota = (quota+9)/10 - 1
	if quota < 0 {
		return 0
	}
	return quota
}

// AdvancedCivilizationStartingBC 對應 Orion2.exe 1.31 的 sub_E5832：
// Advanced Civilization 開局依 TRAIT_MONEY raw 值設定國庫，而不是沿用一般開局 50 BC。
// 原版公式是 (raw+2)*100；合法標準值 -1／0／+1 分別得到 100／200／300 BC。
func AdvancedCivilizationStartingBC(moneyRaw int) int {
	return (moneyRaw + 2) * 100
}
