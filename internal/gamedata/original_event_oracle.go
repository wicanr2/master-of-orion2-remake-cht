package gamedata

// original_event_oracle.go 收納由原版 raw 位址直接證實、但尚未能安全掛到
// remake 36-event record 的純函式。它們不自行捏造事件名稱、觸發率或目標欄位。

// OriginalEventWeightedChoice 對應 Ken_Init / sub_586D4 @ 0x586D4。
// 原版先把總和壓到小於 0x200：每輪將所有候選權重整除 2，再以
// random(total) 做累減抽樣。roll 應是 [0,total)；為了測試可重現，函式會
// 將外部 roll 夾回這個範圍。回傳 normalized 是實際抽樣前的權重快照。
// 權重全為零時 ok=false，choice 保持 0。
func OriginalEventWeightedChoice(weights []int, roll int) (choice int, normalized []int, ok bool) {
	normalized = append([]int(nil), weights...)
	if len(normalized) == 0 {
		return 0, normalized, false
	}
	for {
		total := 0
		for _, weight := range normalized {
			if weight > 0 {
				total += weight
			}
		}
		if total < 0x200 {
			break
		}
		for i, weight := range normalized {
			if weight > 0 {
				normalized[i] = weight / 2
			}
		}
	}
	total := 0
	for _, weight := range normalized {
		if weight > 0 {
			total += weight
		}
	}
	if total == 0 {
		return 0, normalized, false
	}
	if roll < 0 {
		roll = 0
	}
	roll %= total
	for choice, weight := range normalized {
		if weight <= 0 {
			continue
		}
		if roll < weight {
			return choice, normalized, true
		}
		roll -= weight
	}
	// total 是正權重總和，理論上迴圈必定在上面返回；保留最後一個
	// 正權重作為整數邊界的防禦性回退，不把零權重誤選成候選。
	for i := len(normalized) - 1; i >= 0; i-- {
		if normalized[i] > 0 {
			return i, normalized, true
		}
	}
	return 0, normalized, false
}

// OriginalEventVictimWeights 對應 Determine_Event @ 0x22D57 的玩家漂移候選。
// eligible=false 的玩家不進池；good=true 先排除分數最高者，否則先排除分數
// 最低者，剩餘候選權重是與該極值的差平方。它只回傳 raw 漂移權重，不把
// 「誰是好／壞事件」偷換成 remake 的事件 ID。
func OriginalEventVictimWeights(scores []int, eligible []bool, good bool) (indices, weights []int) {
	limit := len(scores)
	if len(eligible) < limit {
		limit = len(eligible)
	}
	if limit == 0 {
		return nil, nil
	}
	candidates := make([]int, 0, limit)
	for i := 0; i < limit; i++ {
		if eligible[i] {
			candidates = append(candidates, i)
		}
	}
	if len(candidates) <= 1 {
		return nil, nil
	}
	extreme := candidates[0]
	for _, idx := range candidates[1:] {
		if (good && scores[idx] > scores[extreme]) || (!good && scores[idx] < scores[extreme]) {
			extreme = idx
		}
	}
	indices = make([]int, 0, len(candidates)-1)
	weights = make([]int, 0, len(candidates)-1)
	for _, idx := range candidates {
		if idx == extreme {
			continue
		}
		delta := scores[idx] - scores[extreme]
		if delta < 0 {
			delta = -delta
		}
		indices = append(indices, idx)
		weights = append(weights, delta*delta)
	}
	return indices, weights
}
