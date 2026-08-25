package gamedata

// AntaranShipCosts 是 1.31 byte_181734 的五個非零艦級成本。
var AntaranShipCosts = [5]int{2, 5, 12, 30, 75}

// AntaranOffensiveMax／AntaranDefensiveMax 分別來自 byte_18173D／byte_181746。
var (
	AntaranOffensiveMax = [5]int{4, 4, 3, 2, 2}
	AntaranDefensiveMax = [5]int{0, 0, 3, 2, 7}
)

// OriginalAntaranTechDelay 對應 sub_63E4C：曲速前／一般／先進延遲 200／100／0 回合。
func OriginalAntaranTechDelay(techLevel int) int {
	switch techLevel {
	case 0:
		return 200
	case 1:
		return 100
	default:
		return 0
	}
}

// OriginalAntaranResourcePulse 對應 sub_645EC。elapsed 是從 3500.0 起算的完整回合數；
// 回傳本回合每一份資源的數量與是否有 pulse。
func OriginalAntaranResourcePulse(elapsed, techLevel, difficulty int) (int, bool) {
	delta := elapsed - OriginalAntaranTechDelay(techLevel)
	if delta <= 0 || delta%25 != 0 {
		return 0, false
	}
	period := delta / 25
	scale := 100
	if difficulty == 3 {
		scale = 150
	} else if difficulty >= 4 {
		scale = 200
	}
	return (period*scale + 99) / 100, true
}

// OriginalAntaranDefenseComplete 對應 sub_646BD：只檢查原始上限非零的艦級。
func OriginalAntaranDefenseComplete(ships [5]int) bool {
	for i, max := range AntaranDefensiveMax {
		if max > 0 && ships[i] < max {
			return false
		}
	}
	return true
}

func antaranEarlyClassLimit(class, difficulty int) int {
	if class == 0 {
		if difficulty == 3 {
			return 12500 / 150
		}
		if difficulty >= 4 {
			return 12500 / 200
		}
		return 100
	}
	if class == 1 {
		if difficulty == 3 {
			return 20000 / 150
		}
		if difficulty >= 4 {
			return 20000 / 200
		}
		return 199
	}
	return 1 << 30
}

// OriginalAntaranBuildShips 對應 sub_63FF0 可表示的五級建艦迴圈。
// maxima 與 costs 會像原版靜態表一樣被本局狀態修改；回傳本回合建成艦數。
func OriginalAntaranBuildShips(resource *int, ships, maxima, costs *[5]int,
	offensive bool, elapsed, techLevel, difficulty int, discounted *bool,
) int {
	if resource == nil || ships == nil || maxima == nil || costs == nil || *resource < costs[0] {
		return 0
	}
	start := 0
	if techLevel >= 2 {
		start = 1
	}
	built := 0
	for {
		pick := -1
		for class := start; class < len(costs); class++ {
			if class <= 1 && elapsed > antaranEarlyClassLimit(class, difficulty) {
				continue
			}
			if costs[class] <= 0 || *resource < costs[class] || maxima[class] <= ships[class] {
				continue
			}
			pick = class
			break
		}
		if pick < 0 {
			return built
		}
		ships[pick]++
		*resource -= costs[pick]
		built++
		if offensive && pick <= 2 && ships[pick] == maxima[pick] {
			maxima[pick] = 0
		}
		if pick == 4 && difficulty > 2 && discounted != nil && !*discounted {
			*discounted = true
			for i := range costs {
				costs[i] = costs[i] * 90 / 100
			}
		}
	}
}

// OriginalAntaranWeightedStrength 是 sub_63DDA／sub_63E73 使用的成本加權總和。
func OriginalAntaranWeightedStrength(ships [5]int, costs [5]int) int {
	total := 0
	for i := range ships {
		total += ships[i] * costs[i]
	}
	return total
}

// OriginalAntaranInvasionReady 對應 sub_63EDD 的戰力與未部署艦閘門。
func OriginalAntaranInvasionReady(offensive, defensive, deployed [5]int, costs [5]int,
	offensiveResource, defensiveResource int,
) bool {
	offensiveStrength := OriginalAntaranWeightedStrength(offensive, costs)
	total := offensiveStrength + OriginalAntaranWeightedStrength(defensive, costs) + offensiveResource + defensiveResource
	if offensiveStrength*4 < total {
		return false
	}
	for i := range offensive {
		if offensive[i] > deployed[i] {
			return true
		}
	}
	return false
}

// OriginalAntaranInvasionRollSucceeds 對應 Random(200)-1 < readiness+1。
func OriginalAntaranInvasionRollSucceeds(readiness, roll1Based int) bool {
	if roll1Based < 1 || roll1Based > 200 || readiness <= 0 {
		return false
	}
	return roll1Based-1 < readiness+1
}

// OriginalAntaranTargetWeights 對應 sub_22F5C 的帝國目標表。回傳與輸入等長的權重；
// 最低 raw 人口帝國被排除，Lucky 只把難度調整後分數除以三，不是完全免疫。
func OriginalAntaranTargetWeights(populations []int, eligible, lucky []bool, difficulty int) []int {
	out := make([]int, len(populations))
	minimum, haveMinimum := 0, false
	for i, pop := range populations {
		if i >= len(eligible) || !eligible[i] {
			continue
		}
		if !haveMinimum || pop < minimum {
			minimum, haveMinimum = pop, true
		}
	}
	if !haveMinimum {
		return out
	}
	for i, pop := range populations {
		if i >= len(eligible) || !eligible[i] || pop == minimum {
			continue
		}
		score := pop
		switch difficulty {
		case 0:
			score /= 10
		case 1:
			score /= 5
		case 3:
			score = score * 3 / 2
		case 4:
			score *= 3
		}
		if i < len(lucky) && lucky[i] {
			score /= 3
		}
		delta := score - minimum
		out[i] = delta * delta
	}
	return out
}
