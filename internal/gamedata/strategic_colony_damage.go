package gamedata

import "sort"

// StrategicColonyDamageState 是原版 sub_DCEBD @ 0xDCEBD 會讀寫的殖民地傷亡欄位。
// LastPopulationPoints 使用 .GAM RacePopulation 的百分之一人口單位；人口只剩一時，
// 零值由 resolver 安全視為完整的 100 點。
type StrategicColonyDamageState struct {
	Population           int
	LastPopulationPoints int
	Marines              int
	Tanks                int
	BuildProgress        int
	RawBuildingIDs       []int
	MarineHitCost        int
	TankHitCost          int
	BuildingHitCost      int
}

type StrategicColonyDamageResult struct {
	State                StrategicColonyDamageState
	DamageSpent          int
	DamageRemaining      int
	PopulationLost       int
	MarinesLost          int
	TanksLost            int
	BuildProgressLost    int
	DestroyedBuildingIDs []int
	ColonyDestroyed      bool
}

var strategicBombardmentExcludedBuildings = map[int]bool{
	8: true, 9: true, 26: true, 27: true, 40: true, 41: true, 42: true, 47: true,
}

// ResolveStrategicColonyDamage 重製 sub_DCEBD 的戰略轟炸候選池與寫回。
// intn 必須回傳 [0,n)；nil 時固定取第一項，僅作測試／安全 fallback。
func ResolveStrategicColonyDamage(state StrategicColonyDamageState, damage int, intn func(int) int) StrategicColonyDamageResult {
	state.Population = nonNegativeColonyHits(state.Population)
	state.Marines = nonNegativeColonyHits(state.Marines)
	state.Tanks = nonNegativeColonyHits(state.Tanks)
	state.BuildProgress = nonNegativeColonyHits(state.BuildProgress)
	state.MarineHitCost = positiveColonyDamageCost(state.MarineHitCost)
	state.TankHitCost = positiveColonyDamageCost(state.TankHitCost)
	state.BuildingHitCost = positiveColonyDamageCost(state.BuildingHitCost)
	if state.Population == 1 && state.LastPopulationPoints <= 0 {
		state.LastPopulationPoints = 100
	}
	if state.LastPopulationPoints < 0 {
		state.LastPopulationPoints = 0
	}
	state.RawBuildingIDs = strategicDamageBuildingIDs(state.RawBuildingIDs)

	result := StrategicColonyDamageResult{State: state, DamageRemaining: nonNegativeColonyHits(damage)}
	for result.DamageRemaining > 0 {
		buildingCount := len(result.State.RawBuildingIDs)
		tailCount := 0
		if result.State.BuildProgress > 0 {
			tailCount++
		}
		if result.State.Population != 1 {
			tailCount += result.State.Population
		}
		candidateCount := buildingCount + result.State.Marines + result.State.Tanks + tailCount
		if candidateCount == 0 {
			if result.State.Population != 1 {
				break
			}
			// 原版最後一名殖民者不進一般候選，而是每點轟炸傷害扣 100 個
			// RacePopulation 百分之一人口點數。
			result.DamageRemaining--
			result.DamageSpent++
			if result.State.LastPopulationPoints <= 100 {
				result.State.LastPopulationPoints = 0
				result.State.Population = 0
				result.PopulationLost++
				result.ColonyDestroyed = true
			} else {
				result.State.LastPopulationPoints -= 100
			}
			continue
		}

		pick := 0
		if intn != nil {
			pick = intn(candidateCount)
		}
		if pick < 0 || pick >= candidateCount {
			pick = 0
		}

		switch {
		case pick < buildingCount:
			if result.DamageRemaining < result.State.BuildingHitCost {
				return result
			}
			id := result.State.RawBuildingIDs[pick]
			result.State.RawBuildingIDs = append(result.State.RawBuildingIDs[:pick], result.State.RawBuildingIDs[pick+1:]...)
			result.DestroyedBuildingIDs = append(result.DestroyedBuildingIDs, id)
			result.DamageRemaining -= result.State.BuildingHitCost
			result.DamageSpent += result.State.BuildingHitCost
		case pick < buildingCount+result.State.Marines:
			if result.DamageRemaining < result.State.MarineHitCost {
				return result
			}
			result.State.Marines--
			result.MarinesLost++
			result.DamageRemaining -= result.State.MarineHitCost
			result.DamageSpent += result.State.MarineHitCost
		case pick < buildingCount+result.State.Marines+result.State.Tanks:
			if result.DamageRemaining < result.State.TankHitCost {
				return result
			}
			result.State.Tanks--
			result.TanksLost++
			result.DamageRemaining -= result.State.TankHitCost
			result.DamageSpent += result.State.TankHitCost
		default:
			// 原版候選池對 BuildProgress 只加一格，這裡卻把扣掉前三段後的
			// tail index 與完整 BuildProgress 比較；保留這個可觀察的原版不對稱。
			tailIndex := pick - buildingCount - result.State.Marines - result.State.Tanks
			if tailIndex < result.State.BuildProgress {
				result.BuildProgressLost += result.State.BuildProgress
				result.State.BuildProgress = 0
			} else if result.State.Population > 0 {
				result.State.Population--
				result.PopulationLost++
			}
			result.DamageRemaining--
			result.DamageSpent++
		}
	}
	return result
}

func strategicDamageBuildingIDs(ids []int) []int {
	seen := make(map[int]bool, len(ids))
	out := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 || id >= len(OriginalBuildingTable) || strategicBombardmentExcludedBuildings[id] || seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, id)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(out)))
	return out
}

func positiveColonyDamageCost(cost int) int {
	if cost < 1 {
		return 1
	}
	return cost
}
