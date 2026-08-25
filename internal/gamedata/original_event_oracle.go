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

// OriginalEventVictimWeights 對應 sub_22D57 @ 0x22D57 的帝國目標候選。
// populations 是 sub_E2710 @ 0xE2710 寫入 player+0xA6 的殖民地總人口，並非
// remake 自訂國力分數。eligible=false 的帝國不進池；good=true 先排除人口最高者，
// 否則先排除人口最低者，剩餘候選權重是與該極值的差平方。它只回傳 raw
// 目標權重，不把「誰是好／壞事件」偷換成 remake 的事件 ID。
func OriginalEventVictimWeights(populations []int, eligible []bool, good bool) (indices, weights []int) {
	limit := len(populations)
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
		if (good && populations[idx] > populations[extreme]) || (!good && populations[idx] < populations[extreme]) {
			extreme = idx
		}
	}
	indices = make([]int, 0, len(candidates)-1)
	weights = make([]int, 0, len(candidates)-1)
	for _, idx := range candidates {
		if idx == extreme {
			continue
		}
		delta := populations[idx] - populations[extreme]
		if delta < 0 {
			delta = -delta
		}
		indices = append(indices, idx)
		weights = append(weights, delta*delta)
	}
	return indices, weights
}

// OriginalLuckyEventDivisor 對應 sub_24511 @ 0x24511。兩個設定旗標分別是
// Random Events 與 Antaran Attacks；回傳值是 Lucky 累積計數器的除數。
func OriginalLuckyEventDivisor(randomEventsEnabled, antaranAttacksEnabled bool) int {
	if randomEventsEnabled {
		if antaranAttacksEnabled {
			return 8
		}
		return 6
	}
	if antaranAttacksEnabled {
		return 12
	}
	return 10
}

// OriginalLuckyEventRollSucceeds 對應 sub_24511 的整數除法與 <= 比較。
// roll1Based 是原版 Random(1000) 的 1..1000 結果；非法輸入安全視為失敗。
func OriginalLuckyEventRollSucceeds(counter, divisor, roll1Based int) bool {
	if counter < 0 || divisor <= 0 || roll1Based < 1 || roll1Based > 1000 {
		return false
	}
	return roll1Based <= counter/divisor
}

// OriginalEventScheduleThreshold 對應 sub_2230A @ 0x2230A 的一般事件門檻。
// 前五次檢查只增加 attemptCounter，當次門檻固定為 0；之後才按難度套用
// delta 的 1/2、2/3、3/4、4/5、5/6。difficulty 超界時 ok=false。
func OriginalEventScheduleThreshold(delta, attemptCounter, difficulty int) (threshold, nextAttempts int, ok bool) {
	if delta < 0 || difficulty < 0 || difficulty > 4 {
		return 0, attemptCounter, false
	}
	if attemptCounter < 0 {
		attemptCounter = 0
	}
	if attemptCounter < 5 {
		return 0, attemptCounter + 1, true
	}
	nextAttempts = 5
	switch difficulty {
	case 0:
		threshold = delta / 2
	case 1:
		threshold = (2 * delta) / 3
	case 2:
		threshold = (3 * delta) / 4
	case 3:
		threshold = (4 * delta) / 5
	case 4:
		threshold = (5 * delta) / 6
	}
	return threshold, nextAttempts, true
}

// OriginalEventScheduleRollSucceeds 對應 Random(512) <= threshold。
func OriginalEventScheduleRollSucceeds(threshold, roll1Based int) bool {
	return threshold >= 0 && roll1Based >= 1 && roll1Based <= 512 && roll1Based <= threshold
}

// OriginalMerchantDonation 對應 sub_2230A 建立事件 6 的 record+3 金額。
// elapsed 是相對開局星曆（remake 的 Turn-1），每滿 20 回合增加 100 BC。
func OriginalMerchantDonation(elapsed int) (amount int, ok bool) {
	if elapsed < 0 {
		return 0, false
	}
	return (elapsed/20)*100 + 100, true
}

// OriginalPirateRaidLoss 對應 sub_2230A 建立事件 15 的 record+5 金額。
// 原版只接受國庫至少 100 BC 的帝國；Random(21) 是 1..21，映射成 30..50%。
func OriginalPirateRaidLoss(treasury, roll1Based int) (amount int, ok bool) {
	if treasury < 100 || roll1Based < 1 || roll1Based > 21 {
		return 0, false
	}
	return treasury * (roll1Based + 29) / 100, true
}

// OriginalComputerVirusLoss 對應 sub_2230A 的事件 3 適用門檻，以及
// sub_206A2 case 3 的即時研究進度扣除。Random(50) 為 1..50，原版再加 50；
// 若結果高於目前進度，就只扣到零。
func OriginalComputerVirusLoss(progress, roll1Based int) (amount int, ok bool) {
	if progress < 10 || roll1Based < 1 || roll1Based > 50 {
		return 0, false
	}
	amount = roll1Based + 50
	if amount > progress {
		amount = progress
	}
	return amount, true
}

// OriginalEventMinimumTurn 回傳 sub_2230A 對特定事件 ID 的相對星曆限制。
// 沒有額外限制的事件回傳 0。
func OriginalEventMinimumTurn(eventID int) int {
	switch eventID {
	case 19:
		return 100
	case 22:
		return 150
	case 2, 20, 24:
		return 200
	case 23:
		return 250
	case 21:
		return 300
	default:
		return 0
	}
}

// OriginalSupernovaCountdown 對應 sub_2230A 事件 24 建立端：Random(5)+10-difficulty。
func OriginalSupernovaCountdown(difficulty, roll1Based int) int {
	if difficulty < 0 {
		difficulty = 0
	} else if difficulty > 4 {
		difficulty = 4
	}
	if roll1Based < 1 {
		roll1Based = 1
	} else if roll1Based > 5 {
		roll1Based = 5
	}
	return roll1Based + 10 - difficulty
}

// OriginalSupernovaResearchNeed 對應建立端把目標星五個殖民槽 +0xEB 加總後乘倒數。
func OriginalSupernovaResearchNeed(systemResearch, countdown int) int {
	if systemResearch < 0 {
		systemResearch = 0
	}
	if countdown < 0 {
		countdown = 0
	}
	return systemResearch * countdown
}

// OriginalStasisEnds 對應 sub_206A2 case 25：age>4 才擲 Random(20)==1，age>20 強制結束。
func OriginalStasisEnds(age, roll1Based int) bool {
	if age > 20 {
		return true
	}
	return age > 4 && roll1Based == 1
}
