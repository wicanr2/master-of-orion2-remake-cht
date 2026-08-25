package gamedata

// OriginalStartingValueState 是先進級開局應用估值可由 remake 表示的原版輸入。
// Known 與 Opponents 是已取得的科技應用集合；Raw4Known=false 時刻意略過
// player+0x205 的四個後段分支，不以猜測語意名稱補值。
type OriginalStartingValueState struct {
	Human           bool
	AIProfile       OriginalAITechProfile
	AIProfileKnown  bool
	Difficulty      int
	Raw4            int
	Raw4Known       bool
	InitialSixKnown bool
	Known           map[Technology]bool
	Opponents       []map[Technology]bool
}

// OriginalAITechProfile 保留 sub_589D6 寫入 player+0x28/+0x205/+0x206 的 raw 值。
// 這三欄的玩家可見名稱尚未由 consumer 閉合，因此不用推測名稱當型別。
type OriginalAITechProfile struct {
	Raw6   int
	Raw4   int
	Raw7   int
	Traits [RaceTraitCount]int8
}

func originalWeightedIndex(weights []int, roll func(int) int) int {
	total := 0
	for _, weight := range weights {
		total += weight
	}
	for total > 1000 {
		total = 0
		for i := range weights {
			weights[i] /= 2
			total += weights[i]
		}
	}
	if total <= 0 {
		return 0
	}
	r := roll(total)
	for i, weight := range weights {
		if r < weight {
			return i
		}
		r -= weight
	}
	return len(weights) - 1
}

// RollOriginalAITechProfile 實作 sub_589D6 的三組原始加權表。
var originalRaceRaw27Table = [OrigRaceCount][10]uint8{
	{3, 4, 4, 4, 4, 4, 4, 4, 5, 5}, {1, 2, 2, 2, 2, 2, 2, 2, 3, 3},
	{1, 1, 0, 2, 2, 2, 2, 2, 2, 2}, {0, 0, 1, 1, 1, 1, 1, 2, 2, 2},
	{1, 1, 2, 2, 3, 3, 3, 5, 5, 5}, {3, 4, 4, 4, 4, 4, 4, 4, 5, 5},
	{0, 0, 0, 0, 0, 0, 0, 1, 2, 2}, {0, 0, 2, 2, 3, 3, 3, 3, 3, 3},
	{0, 1, 1, 1, 1, 1, 1, 2, 2, 3}, {3, 4, 5, 5, 5, 5, 5, 5, 5, 5},
	{1, 2, 2, 2, 2, 2, 2, 2, 3, 3}, {0, 0, 0, 0, 0, 0, 0, 0, 2, 2},
	{3, 3, 4, 4, 4, 4, 4, 5, 5, 5},
}

// RollOriginalAIRaw27 實作 sub_589D6 對 byte_181090 的種族十格抽選。
func RollOriginalAIRaw27(origRace, difficulty int, roll func(int) int) int {
	if origRace < 0 || origRace >= OrigRaceCount {
		return 0
	}
	column := roll(10) + 1 - difficulty
	if column < 0 {
		column = 0
	}
	if column > 9 {
		column = 9
	}
	return int(originalRaceRaw27Table[origRace][column])
}

func RollOriginalAITechProfile(traits [RaceTraitCount]int8, difficulty, raw27 int, roll func(int) int) OriginalAITechProfile {
	w6 := []int{1, 2, 1, 2, 2, 1}
	w4 := []int{1, 2, 2, 1}
	w7 := []int{2, 2, 1, 2, 1, 2, 3}
	add := func(w []int, i, n int) { w[i] += n }
	t := func(i int) int { return int(traits[i]) }
	if t(1) == 50 {
		add(w6, 5, 11)
		add(w6, 2, 11)
	}
	if t(1) == 100 {
		add(w6, 5, 100)
		add(w6, 2, 100)
	}
	if t(2) == 1 {
		add(w6, 5, 3)
		add(w6, 4, 3)
		add(w6, 3, 3)
	}
	if t(2) == 2 {
		add(w6, 5, 100)
		add(w6, 4, 100)
		add(w6, 3, 100)
	}
	if t(3) == 1 {
		add(w6, 5, 3)
		add(w6, 4, 3)
		add(w6, 3, 3)
	}
	if t(3) == 2 {
		add(w6, 5, 10)
		add(w6, 4, 10)
		add(w6, 3, 10)
	}
	if t(4) == 1 {
		add(w6, 4, 3)
		add(w6, 3, 10)
	}
	if t(4) == 2 {
		add(w6, 4, 10)
		add(w6, 3, 100)
	}
	if t(6) == 20 {
		add(w6, 1, 10)
		add(w7, 3, 100)
		add(w4, 0, 10)
	}
	if t(6) == 40 {
		add(w6, 1, 100)
		add(w7, 3, 1000)
		add(w4, 0, 100)
	}
	if t(7) == 25 {
		add(w6, 1, 10)
		add(w7, 4, 3)
		add(w4, 1, 100)
	}
	if t(7) == 50 {
		add(w6, 1, 100)
		add(w7, 4, 10)
		add(w4, 1, 1000)
	}
	if t(8) == 10 {
		add(w6, 2, 10)
		add(w7, 1, 100)
	}
	if t(8) == 20 {
		add(w6, 2, 100)
		add(w7, 1, 1000)
	}
	if t(9) == 10 {
		add(w6, 0, 3)
	}
	if t(9) == 20 {
		add(w6, 0, 10)
	}
	if t(10) != 0 {
		add(w6, 4, 10)
		add(w6, 5, 3)
		add(w6, 3, 3)
	}
	if t(11) != 0 {
		add(w6, 2, 10)
		add(w7, 1, 1000)
	}
	if t(12) != 0 {
		add(w6, 4, 3)
		add(w6, 5, 3)
		add(w6, 3, 3)
	}
	if t(13) != 0 {
		add(w6, 4, 10)
		add(w6, 5, 3)
		add(w6, 3, 3)
	}
	if t(14) != 0 {
		add(w6, 4, 3)
		add(w6, 5, 10)
	}
	if t(15) != 0 {
		add(w6, 4, 10)
	}
	if t(17) != 0 {
		add(w6, 4, 3)
		add(w7, 5, 1000)
		add(w4, 1, 10)
	}
	if t(18) != 0 {
		add(w6, 4, 3)
		add(w6, 2, 10)
		add(w6, 3, 3)
	}
	if t(20) != 0 {
		add(w6, 0, 100)
	}
	if t(22) != 0 {
		add(w6, 3, 1000)
	}
	if t(23) != 0 {
		add(w6, 4, 100)
	}
	if t(24) != 0 {
		add(w6, 0, 100)
	}
	if t(28) != 0 {
		add(w6, 1, 100)
	}
	if t(30) != 0 {
		add(w6, 2, 10)
	}
	if t(29) != 0 {
		add(w6, 2, 10)
		add(w6, 1, 13)
		add(w7, 3, 100)
	}
	if difficulty == 3 {
		add(w6, 4, 3)
		add(w6, 2, 3)
		add(w6, 3, 3)
	}
	if difficulty == 4 {
		add(w6, 4, 10)
		add(w6, 2, 10)
		add(w6, 3, 10)
	}
	if raw27 == 0 {
		add(w7, 2, 3)
	}
	return OriginalAITechProfile{
		Raw6: originalWeightedIndex(w6, roll), Raw4: originalWeightedIndex(w4, roll),
		Raw7: originalWeightedIndex(w7, roll), Traits: traits,
	}
}

var humanTechValueQuadCategories = map[int]bool{0x1A: true, 0x1B: true, 0x1D: true, 0x1E: true, 0x21: true, 0x15: true}
var closeKnownCategoryZero = map[int]bool{0x0E: true, 0x10: true, 0x12: true, 0x18: true, 0x19: true, 0x20: true, 0x27: true}

func originalTopicLevel(topic ResearchTopic) int {
	i := int(topic)
	if i < 0 || i >= len(OrigTopicLevel) {
		return 0
	}
	return OrigTopicLevel[i]
}

func originalKnownCategoryLevel(known map[Technology]bool, category, limit int) int {
	last := 0
	if limit > len(TechItemCategory) {
		limit = len(TechItemCategory)
	}
	for i := 0; i < limit; i++ {
		tech := Technology(i)
		if !known[tech] || TechItemCategory[i] != category {
			continue
		}
		topic, ok := OrigTechTopic(tech)
		if ok {
			// 原始迴圈依 tech index 掃描並直接覆寫；表按技術進程排列，最後值即原版結果。
			last = originalTopicLevel(topic)
		}
	}
	return last
}

func originalOpponentBestCategory(state OriginalStartingValueState, category int) int {
	best := 0
	for _, known := range state.Opponents {
		if level := originalKnownCategoryLevel(known, category, 75); level > best {
			best = level
		}
	}
	return best
}

// OriginalHumanTechValueKnownSlice 實作 `Calc_Tech_Value_ @ 0xFC845` 中已由 IDA
// 閉合的人類玩家分支與共用後段。raw4 對應原版 +0x205，已由 sub_589D6 初始化；
// 英文符號列出的 Objective 只能協助導覽，不作欄位語意證據。
func OriginalHumanTechValueKnownSlice(tech Technology, state OriginalStartingValueState) int {
	i := int(tech)
	if i <= 0 || i >= len(TechItemCategory) || state.Known[tech] {
		return 0
	}
	topic, ok := OrigTechTopic(tech)
	if !ok || int(topic) >= 75 { // sub_FD335: topic < 0x4B
		return 0
	}
	category := TechItemCategory[i]
	difficulty := state.Difficulty
	if difficulty < 0 {
		difficulty = 0
	}
	value := difficulty*difficulty*25 + 50
	if humanTechValueQuadCategories[category] {
		value *= 4
	}
	return originalTechValueCommon(tech, state, value)
}

func originalTechValueCommon(tech Technology, state OriginalStartingValueState, value int) int {
	i := int(tech)
	topic, _ := OrigTechTopic(tech)
	category := TechItemCategory[i]
	level := originalTopicLevel(topic)
	if level > 22 {
		level = 22
	}
	base := TechResearchLevelValues[level]

	if category >= 0 && category < len(TechCategoryVar24Flag) && TechCategoryVar24Flag[category] != 0 {
		value *= base
	} else {
		knownLevel := originalKnownCategoryLevel(state.Known, category, 204)
		value *= base
		if knownLevel+3 <= level {
			value = (level - knownLevel) * value / 3
		} else if closeKnownCategoryZero[category] {
			value = 0
		} else {
			denom := knownLevel + 3 - level
			if denom > 0 {
				value = 2 * value / denom
			}
		}
	}

	// FD219/FD199 先掃「其他玩家」tech 0..74，建立交叉類別上限。
	capMultiply := func(triggerCategory, opponentCategory int) {
		if category != triggerCategory {
			return
		}
		cap := originalOpponentBestCategory(state, opponentCategory)
		if base < cap {
			value *= cap - base
		}
	}
	capMultiply(0x19, 0x12)
	capMultiply(0x12, 0x19)
	capMultiply(0x13, 0x0A)
	capMultiply(0x0A, 0x13)
	capMultiply(0x0C, 0x0C)
	capMultiply(0x0F, 0x0F)
	if category == 8 {
		known := originalKnownCategoryLevel(state.Known, 0x0F, 75) +
			originalKnownCategoryLevel(state.Known, 0x10, 75)
		if known < originalOpponentBestCategory(state, 0x0F)*2 {
			value *= 2
		}
	}

	if state.Raw4Known {
		if state.Raw4 == 2 && category == 0x18 {
			cap := originalOpponentBestCategory(state, 0x1C)
			if base < cap {
				value *= cap - base
			}
		}
		capCategory := -1
		switch state.Raw4 {
		case 0:
			if category == 0x1A {
				capCategory = 0x21
			}
		case 1:
			if category == 0x1A || category == 0x13 {
				capCategory = 0x21
			}
		case 2:
			if category == 0x15 {
				capCategory = 0x21
			}
		}
		if capCategory >= 0 {
			cap := originalOpponentBestCategory(state, capCategory)
			if base < cap {
				value *= cap - base
			}
		}
	}

	// 開局 stardate 必在 35000+150 之前。
	if category == 0x12 {
		value *= 2
	}

	// 候選主題尚未完成，因此原版 ×5/4 分支不會在先進級開局觸發。
	if value == 0 {
		pendingOther := false
		for _, other := range researchChoices[int(topic)].Choices {
			if other != tech && !state.Known[other] {
				pendingOther = true
				break
			}
		}
		if !pendingOther {
			value = base * 10
		}
	}
	return value
}

func aiProfileCategoryValue(category int, p OriginalAITechProfile) int {
	value := TechCategoryDefaultMultiplier[category]
	set := func(categories []int, n int) {
		for _, c := range categories {
			if category == c {
				value = n
				return
			}
		}
	}
	switch p.Raw4 {
	case 0:
		set([]int{0x1A, 0x1B, 0x19}, 100)
	case 1:
		set([]int{0x1A}, 50)
		set([]int{0x13, 0x1E}, 100)
		set([]int{0x19}, 20)
	case 2:
		set([]int{0x15, 0x1D, 0x18}, 100)
	case 3:
		set([]int{0x24}, 50)
		set([]int{0x26}, 100)
	}
	switch p.Raw7 {
	case 0:
		set([]int{0x12, 0x17}, 50)
	case 1:
		set([]int{0x0F, 0x10}, 20)
		set([]int{0x20}, 50)
		set([]int{0x23}, 100)
	case 2:
		set([]int{0x12}, 50)
		set([]int{0x14}, 100)
	case 3:
		set([]int{0x1C}, 100)
		set([]int{0x1A}, 20)
	case 4:
		set([]int{0x25}, 100)
	case 5:
		set([]int{0x20, 0x1F}, 100)
	case 6:
		set([]int{0x21, 0x22}, 100)
	}
	switch p.Raw6 {
	case 0:
		set([]int{0x0C, 3, 0x0B}, 100)
	case 1:
		set([]int{0x11, 0x21}, 100)
	case 2:
		set([]int{0x27}, 50)
		set([]int{9, 0x0A}, 100)
	case 3:
		set([]int{2}, 100)
	case 4:
		set([]int{1, 4}, 100)
	case 5:
		set([]int{0, 4}, 100)
	}
	t := func(i int) int { return int(p.Traits[i]) }
	switch category {
	case 0:
		if t(2) < 0 {
			value = 100
		} else if t(2) > 0 {
			value = 10
		}
		if t(18) != 0 {
			value = 1
		} else if t(17) != 0 {
			value = 20
		}
	case 1:
		if t(3) < 0 {
			value = 100
		}
	case 2:
		if t(4) != 0 {
			value = 100
		}
	case 3:
		if t(5) < 0 {
			value = 100
		} else if t(5) > 0 {
			value = 20
		}
	case 4:
		if t(3) > 0 {
			value = 100
		}
		if t(23) != 0 {
			value = 1
		}
	case 6:
		if t(13) != 0 {
			value = 20
		}
		if t(1) < 0 {
			value = 100
		} else if t(1) > 0 {
			value = 5
		}
	case 0x10:
		if t(8) < 0 {
			value = 20
		}
	case 0x0C:
		if t(9) != 0 || t(0)/2 == 2 {
			value = 50
		}
	case 0x12:
		if t(6) < 0 {
			value = 50
		}
	case 0x19:
		if t(7) < 0 {
			value = 100
		}
	case 0x1B:
		if t(7) > 0 {
			value = 100
		}
	case 0x1C:
		if t(6) > 0 {
			value = 100
		}
	case 0x25:
		if t(28) != 0 {
			value = 1
		}
	case 0x28:
		if t(0)/2 == 3 {
			value = 1
		}
	}
	return value
}

// OriginalAITechValueKnownSlice 實作 sub_FC845 的 raw profile／種族類別分支，
// 再進入與人類共用的邊際、對手類別與開局時期修正。
func OriginalAITechValueKnownSlice(tech Technology, state OriginalStartingValueState) int {
	i := int(tech)
	if i <= 0 || i >= len(TechItemCategory) || state.Known[tech] || !state.AIProfileKnown {
		return 0
	}
	topic, ok := OrigTechTopic(tech)
	if !ok || int(topic) >= 75 {
		return 0
	}
	value := aiProfileCategoryValue(TechItemCategory[i], state.AIProfile)
	if i == 5 && state.AIProfile.Traits[25] != 0 {
		value = 1
	}
	if i == 0x83 {
		if state.AIProfile.Traits[11] != 0 {
			value = 1
		} else if state.AIProfile.Traits[10] != 0 {
			value = 50
		}
	}
	return originalTechValueCommon(tech, state, value)
}

// StartingOriginalApplicationPick 依 `Choose_Tech_Application_ @ 0xFD335` 的應用級
// 候選、15 回合視野、成本反比與單次加權抽選，回傳同時選中的主題與科技。
func StartingOriginalApplicationPick(available []ResearchTopic, researchPerTurn int,
	state OriginalStartingValueState, roll func(int) int) (ResearchTopic, Technology, bool) {
	if researchPerTurn < 1 {
		researchPerTurn = 1
	}
	availableSet := make(map[ResearchTopic]bool, len(available))
	for _, topic := range available {
		availableSet[topic] = true
	}
	scores := make([]int, len(TechItemCategory))
	for techIdx := 1; techIdx < len(scores); techIdx++ {
		tech := Technology(techIdx)
		topic, ok := OrigTechTopic(tech)
		if !ok || !availableSet[topic] {
			continue
		}
		value := TechCategoryWeight(tech)
		if state.Human {
			value = OriginalHumanTechValueKnownSlice(tech, state)
		} else if state.AIProfileKnown {
			value = OriginalAITechValueKnownSlice(tech, state)
		}
		if state.InitialSixKnown {
			if topic == 4 || techIdx == 0x72 || topic == 0x49 {
				value *= 2
			}
			if techIdx == 0x33 {
				value *= 5
			}
			if value == 0 {
				value = 1
			}
		}
		scores[techIdx] = value
	}

	for horizon := StartingRandomHorizonInitial; horizon <= 1<<20; horizon = startingRandomHorizonGrow(horizon) {
		total := 0
		weighted := make([]int, len(scores))
		for techIdx, value := range scores {
			if value <= 0 {
				continue
			}
			topic, _ := OrigTechTopic(Technology(techIdx))
			turns := OrigTopicCost[int(topic)] / researchPerTurn
			if turns < 1 {
				turns = 1
			}
			if turns > horizon {
				continue
			}
			weighted[techIdx] = value * horizon / turns
		}
		if !state.Human && state.AIProfileKnown && state.Difficulty > 0 {
			maximum := 0
			for _, score := range weighted {
				if score > maximum {
					maximum = score
				}
			}
			for i, score := range weighted {
				drop := false
				switch state.AIProfile.Raw6 {
				case 1, 2:
					drop = score*6 < maximum
				case 3, 4:
					drop = score*5 < maximum*4
				case 5:
					drop = score*2 < maximum
				}
				if drop {
					weighted[i] = 0
				}
			}
		}
		for _, score := range weighted {
			total += score
		}
		if total <= 0 {
			continue
		}
		r := roll(total)
		for techIdx, score := range weighted {
			if r < score {
				tech := Technology(techIdx)
				topic, _ := OrigTechTopic(tech)
				return topic, tech, true
			}
			r -= score
		}
	}
	return 0, 0, false
}

// ResearchTopicGrantsAll 回報該主題是否為 ResearchAll。
func ResearchTopicGrantsAll(topic ResearchTopic) bool {
	i := int(topic)
	return i >= 0 && i < len(researchChoices) && researchChoices[i].ResearchAll
}
