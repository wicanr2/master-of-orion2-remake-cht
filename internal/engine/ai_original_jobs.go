package engine

import (
	"slices"
	"sort"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// OriginalAIJobContext 是原版 sub_D652C／sub_D66B3 已閉合輸入。
// ColonyFoodHalf 對應 colony+0xDD；Known=false 時不得把猜測值送進原版路徑。
type OriginalAIJobContext struct {
	Personality         ai.Personality
	LateTech            bool
	ColonyFoodHalf      []int
	ColonyFoodHalfKnown []bool
	// ColonyBlockaded 與 colonies 同長；nil 只供既有未封鎖 API 相容。
	ColonyBlockaded []bool
}

type originalAIColonistCandidate struct {
	group, race, from        int
	prisoner                 bool
	food, industry, research int
}

func originalAIGroupUnitOutput(cs ColonyState, g PopulationGroup, job int, prisoner bool) int {
	perUnit, workerFloor := 0, false
	switch job {
	case int(gamedata.FARMER):
		perUnit = groupFoodPerFarmer(cs, g)
	case int(gamedata.WORKER):
		perUnit = cs.IndustryPerWorker - cs.OwnerIndustryBonus + g.IndustryBonus + cs.IndustryPerWorkerBonus
		workerFloor = true
	case int(gamedata.SCIENTIST):
		perUnit = cs.ResearchPerScientist - cs.OwnerResearchBonus + g.ResearchBonus
	}
	prisoners := 0
	if prisoner {
		prisoners = 1
	}
	base := gamedata.UncooperativeJobOutputExact(1, perUnit, prisoners, workerFloor)
	bonus := cs.MoralePercent + groupGravityPenaltyPercent(cs, g)
	switch job {
	case int(gamedata.FARMER):
		bonus += cs.FoodBonusPercent + cs.GovernmentFoodBonusPercent
	case int(gamedata.WORKER):
		bonus += cs.IndustryBonusPercent + cs.GovernmentIndustryBonusPercent
	case int(gamedata.SCIENTIST):
		bonus += cs.ResearchBonusPercent + cs.GovernmentResearchBonusPercent
	}
	return gamedata.GravityAdjustedProduction(base, bonus)
}

func originalAIColonistCandidates(cs ColonyState) []originalAIColonistCandidate {
	var out []originalAIColonistCandidate
	for gi, g := range cs.PopulationGroups {
		if g.RaceSlot == gamedata.AndroidColonistSlot || g.RaceSlot == gamedata.NativeColonistSlot {
			continue
		}
		counts := [3]int{g.Farmers, g.Workers, g.Scientists}
		prisoners := [3]int{g.PrisonerFarmers, g.PrisonerWorkers, g.PrisonerScientists}
		for job := 0; job < 3; job++ {
			for n := 0; n < counts[job]; n++ {
				p := n < prisoners[job]
				out = append(out, originalAIColonistCandidate{
					group: gi, race: g.RaceSlot, from: job, prisoner: p,
					food:     originalAIGroupUnitOutput(cs, g, int(gamedata.FARMER), p),
					industry: originalAIGroupUnitOutput(cs, g, int(gamedata.WORKER), p),
					research: originalAIGroupUnitOutput(cs, g, int(gamedata.SCIENTIST), p),
				})
			}
		}
	}
	return out
}

func originalAIMoveCandidate(cs *ColonyState, c originalAIColonistCandidate, to int) {
	if c.from == to {
		return
	}
	g := &cs.PopulationGroups[c.group]
	counts := []*int{&g.Farmers, &g.Workers, &g.Scientists}
	prisoners := []*int{&g.PrisonerFarmers, &g.PrisonerWorkers, &g.PrisonerScientists}
	*counts[c.from]--
	*counts[to]++
	if c.prisoner {
		*prisoners[c.from]--
		*prisoners[to]++
	}
	cs.Farmers, cs.Workers, cs.Scientists = 0, 0, 0
	for _, group := range cs.PopulationGroups {
		cs.Farmers += group.Farmers
		cs.Workers += group.Workers
		cs.Scientists += group.Scientists
	}
}

func originalAISortInitial(candidates []originalAIColonistCandidate, hasFood bool) {
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if !hasFood {
			if av, bv := a.research-a.industry, b.research-b.industry; av != bv {
				return av > bv
			}
			if a.industry != b.industry {
				return a.industry < b.industry
			}
			if a.race != b.race {
				return a.race < b.race
			}
			return originalAIInitialEquivalentLess(a, b)
		}
		if av, bv := a.food+a.industry-2*a.research, b.food+b.industry-2*b.research; av != bv {
			return av > bv
		}
		if a.industry != b.industry {
			return a.industry > b.industry
		}
		if a.research != b.research {
			return a.research < b.research
		}
		if a.race != b.race {
			return a.race < b.race
		}
		return originalAIInitialEquivalentLess(a, b)
	})
}

// sub_D63A6／sub_D6315 對同種族、同邊際產出的逐人口記錄回傳相等，最終排列
// 取決於 Watcom qsort 與原始 4-byte 人口陣列順序。ColonyState 只保存群組職務
// 計數，無法重建那份順序；等價類別內把既有工人排到尾端，讓 sub_D652C 從尾端
// 取「最低必要工人」時優先維持原職，避免穩定排序固定犧牲科學家的表示偏差。
// 這是 deterministic reconstruction，不冒稱原版 qsort 的逐位元 parity。
func originalAIInitialEquivalentLess(a, b originalAIColonistCandidate) bool {
	if a.from == b.from {
		return false
	}
	return a.from != int(gamedata.WORKER) && b.from == int(gamedata.WORKER)
}

func originalAISortMarginal(candidates []originalAIColonistCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if av, bv := a.research-a.industry, b.research-b.industry; av != bv {
			return av > bv
		}
		if a.industry != b.industry {
			return a.industry < b.industry
		}
		return a.race < b.race
	})
}

func originalAISortBlockaded(candidates []originalAIColonistCandidate) {
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if av, bv := a.food-a.industry, b.food-b.industry; av != bv {
			return av > bv
		}
		return a.food < b.food
	})
}

func applyOriginalAIBlockadedColony(cs *ColonyState, foodHalf int) {
	items := originalAIColonistCandidates(*cs)
	originalAISortBlockaded(items)
	if foodHalf <= 0 {
		to := int(gamedata.WORKER)
		if cs.ResearchDiverted { // sub_23DFE 命中。
			to = int(gamedata.SCIENTIST)
		}
		for _, item := range items {
			originalAIMoveCandidate(cs, item, to)
		}
		return
	}
	first, end := 0, len(items)
	for first < end {
		co := RunColonyTurn(*cs)
		if 2*co.Food >= co.FoodConsumedHalf {
			end--
			originalAIMoveCandidate(cs, items[end], int(gamedata.WORKER))
			continue
		}
		originalAIMoveCandidate(cs, items[first], int(gamedata.FARMER))
		first++
	}
}

// applyOriginalAIAdditionalFarmers 對映 sub_D6AD4 → sub_D6A00 的玩家可表示切片。
// 原版在工業／研究平衡後仍會逐人補農夫，直到帝國食物與 player+0x38 運輸壓力解除。
func originalAIAdditionalFarmerScore(candidate originalAIColonistCandidate, cleanupBefore, cleanupNow int) int {
	score := candidate.food - candidate.research
	if candidate.from == int(gamedata.WORKER) {
		score = candidate.food - candidate.industry
		if cleanupBefore < cleanupNow {
			score += 1000
		}
	}
	return score
}

func originalAISortAdditionalFarmers(candidates []originalAIColonistCandidate, cleanupBefore, cleanupNow int) {
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		as := originalAIAdditionalFarmerScore(a, cleanupBefore, cleanupNow)
		bs := originalAIAdditionalFarmerScore(b, cleanupBefore, cleanupNow)
		if as != bs {
			return as > bs
		}
		if a.food != b.food {
			return a.food < b.food
		}
		if a.from != b.from {
			return a.from < b.from
		}
		return a.race < b.race
	})
}

func applyOriginalAIAdditionalFarmers(ps PlayerState, colonies []ColonyState, blockaded []bool, cleanupBefore []int) (freighterPressure bool) {
	for {
		empireFoodDeficit := RunEmpireTurn(ps, colonies).TotalFoodHalf < 0
		transport, ok := OriginalFoodTransport(ps, colonies, blockaded)
		if !ok {
			return freighterPressure
		}
		if !empireFoodDeficit && transport.SurplusFreighters >= 0 {
			return freighterPressure
		}
		preferLocal := transport.SurplusFreighters <= 0
		if preferLocal {
			freighterPressure = true
		}

		bestColony, bestGroup, bestFrom := -1, -1, -1
		bestScore := -32768
		// sub_D6AD4 從 AI record 尾端往前掃，跨殖民地只比較已排序首候選的 score；
		// 同分保留較高 colony index，不再套候選的次鍵。
		for ci := len(colonies) - 1; ci >= 0; ci-- {
			if blockaded != nil && blockaded[ci] {
				continue
			}
			before := RunColonyTurn(colonies[ci]).FoodSurplusHalf
			if preferLocal && before >= 0 {
				continue
			}
			candidates := originalAIColonistCandidates(colonies[ci])
			candidates = slices.DeleteFunc(candidates, func(candidate originalAIColonistCandidate) bool {
				return candidate.from == int(gamedata.FARMER)
			})
			if len(candidates) == 0 {
				continue
			}
			cleanupNow := RunColonyTurn(colonies[ci]).PollutionCleanupCost
			originalAISortAdditionalFarmers(candidates, cleanupBefore[ci], cleanupNow)
			candidate := candidates[0]
			score := originalAIAdditionalFarmerScore(candidate, cleanupBefore[ci], cleanupNow)
			if score > bestScore {
				bestColony, bestGroup, bestFrom, bestScore = ci, candidate.group, candidate.from, score
			}
		}
		if bestColony < 0 {
			return freighterPressure
		}
		// 同群／同職人口對已閉合 comparator 等價；重新建立候選可避免先前
		// 改職後保存過期的 from／prisoner 索引。
		for _, candidate := range originalAIColonistCandidates(colonies[bestColony]) {
			if candidate.group == bestGroup && candidate.from == bestFrom {
				originalAIMoveCandidate(&colonies[bestColony], candidate, int(gamedata.FARMER))
				break
			}
		}
	}
}

// preserveAIResearchForCollapsedQsortTies 補償 ColonyState 群組模型遺失原版逐人口
// 4-byte 陣列順序的資訊缺口。當 Watcom qsort 比較器對整個等價類別回傳 0 時，
// 原版結果取決於該陣列與 runtime partition；remake 若因固定群組順序把所有科學家
// 都吃掉，就會永久停在同一研究主題。只在帝國研究為 0 時，把邊際損失最小的一名
// 工人保留為科學家；這是 deterministic reconstruction，不是已證實的原版 tie-break。
func preserveAIResearchForCollapsedQsortTies(ps PlayerState, colonies []ColonyState, blockaded []bool) {
	if RunEmpireTurn(ps, colonies).TotalResearch > 0 {
		return
	}
	bestColony, bestGroup := -1, -1
	bestMargin := -32768
	for ci := range colonies {
		if colonies[ci].Population < 2 || (blockaded != nil && blockaded[ci]) {
			continue
		}
		for _, candidate := range originalAIColonistCandidates(colonies[ci]) {
			if candidate.from != int(gamedata.WORKER) || candidate.research <= 0 {
				continue
			}
			if margin := candidate.research - candidate.industry; margin > bestMargin {
				bestColony, bestGroup, bestMargin = ci, candidate.group, margin
			}
		}
	}
	if bestColony < 0 {
		return
	}
	for _, candidate := range originalAIColonistCandidates(colonies[bestColony]) {
		if candidate.group == bestGroup && candidate.from == int(gamedata.WORKER) {
			originalAIMoveCandidate(&colonies[bestColony], candidate, int(gamedata.SCIENTIST))
			return
		}
	}
}

// ApplyOriginalAIUnblockadedJobs 重建 sub_D652C／sub_D66B3 的未封鎖路徑。
// 它不處理 sub_D61E7；呼叫端若存在封鎖殖民地必須回退並保留未閉合狀態。
func ApplyOriginalAIUnblockadedJobs(ps PlayerState, colonies []ColonyState, ctx OriginalAIJobContext) ([]ColonyState, bool) {
	ctx.ColonyBlockaded = make([]bool, len(colonies))
	return ApplyOriginalAIJobs(ps, colonies, ctx)
}

// ApplyOriginalAIJobs 重建 sub_D61E7／sub_D652C／sub_D66B3 的殖民地職務主鏈。
func ApplyOriginalAIJobs(ps PlayerState, colonies []ColonyState, ctx OriginalAIJobContext) ([]ColonyState, bool) {
	out, _, ok := ApplyOriginalAIJobsWithTransport(ps, colonies, ctx)
	return out, ok
}

// ApplyOriginalAIJobsWithTransport 另回傳 sub_D6AD4 是否曾進入 +0x38<=0 的
// 貨運壓力分支，供 shell 套用原版難度亂數的 +5 freighters writer。
func ApplyOriginalAIJobsWithTransport(ps PlayerState, colonies []ColonyState, ctx OriginalAIJobContext) ([]ColonyState, bool, bool) {
	if len(ctx.ColonyFoodHalf) != len(colonies) || len(ctx.ColonyFoodHalfKnown) != len(colonies) {
		return nil, false, false
	}
	if ctx.ColonyBlockaded != nil && len(ctx.ColonyBlockaded) != len(colonies) {
		return nil, false, false
	}
	out := append([]ColonyState(nil), colonies...)
	for i := range out {
		// ColonyState 的 slice 是參考型別；完整驗證前先深拷貝，確保任何後段
		// fallback 都不會把呼叫端原狀態改到一半。
		out[i].PopulationGroups = append([]PopulationGroup(nil), colonies[i].PopulationGroups...)
	}
	type colonyCandidates struct {
		items      []originalAIColonistCandidate
		first, end int
	}
	work := make([]colonyCandidates, len(out))
	cleanupBefore := make([]int, len(out))
	for i := range out {
		if !populationGroupsValid(out[i]) || !ctx.ColonyFoodHalfKnown[i] {
			return nil, false, false
		}
		if ctx.ColonyBlockaded != nil && ctx.ColonyBlockaded[i] {
			applyOriginalAIBlockadedColony(&out[i], ctx.ColonyFoodHalf[i])
			continue
		}
		items := originalAIColonistCandidates(out[i])
		originalAISortInitial(items, ctx.ColonyFoodHalf[i] > 0)
		requiredWorkers := (out[i].Population + 4) / 5
		if out[i].ResearchDiverted {
			requiredWorkers = 0 // sub_23DFE 命中時把 var_4 清零。
		}
		specialWorkers := 0
		for _, g := range out[i].PopulationGroups {
			if g.RaceSlot == gamedata.AndroidColonistSlot || g.RaceSlot == gamedata.NativeColonistSlot {
				specialWorkers += g.Workers
			}
		}
		end := len(items)
		for end > 0 {
			co := RunColonyTurn(out[i])
			if specialWorkers >= requiredWorkers && 2*co.NetIndustry >= co.IndustryConsumedHalf {
				break
			}
			end--
			originalAIMoveCandidate(&out[i], items[end], int(gamedata.WORKER))
			items[end].from = int(gamedata.WORKER)
			specialWorkers++
		}
		items = items[:end]
		originalAISortMarginal(items)
		work[i] = colonyCandidates{items: items, end: len(items)}
		// sub_D652C @ 0xD660C..0xD660F 只保存 colony+8 的低 byte；後續
		// sub_D66B3 改職重算可能提高完整 signed word 清污成本。
		cleanupBefore[i] = RunColonyTurn(out[i]).PollutionCleanupCost & 0xff
	}

	industryWeight, researchWeight := 10, 18
	probe := RunEmpireTurn(ps, out)
	if probe.NetBC < 0 {
		researchWeight += originalIntegerSqrt(-int(int16(probe.NetBC)))
	}
	if ctx.LateTech {
		industryWeight = 0
	} else if ctx.Personality == ai.PersonalityErratic {
		industryWeight += 2
	} else if ctx.Personality == ai.PersonalityHonorable {
		researchWeight += 2
	}
	for {
		probe = RunEmpireTurn(ps, out)
		// sub_E2710 將工業直接寫成 signed word，研究則先夾至 32767。
		industry := int(int16(probe.TotalNetIndustry))
		research := probe.TotalResearch
		if research > 32767 {
			research = 32767
		}
		wantScientist := research*researchWeight < industry*industryWeight
		bestColony := -1
		if wantScientist {
			bestMargin := -32767
			// 原版從 AI record 尾端往前掃，嚴格較大才替換；同分保留較高 colony index。
			for i := len(work) - 1; i >= 0; i-- {
				if work[i].first >= work[i].end {
					continue
				}
				c := work[i].items[work[i].first]
				if margin := c.research - c.industry; margin > bestMargin {
					bestMargin, bestColony = margin, i
				}
			}
			if bestColony < 0 {
				break
			}
			w := &work[bestColony]
			c := w.items[w.first]
			originalAIMoveCandidate(&out[bestColony], c, int(gamedata.SCIENTIST))
			w.first++
			continue
		}
		bestMargin := 32767
		for i := len(work) - 1; i >= 0; i-- {
			if work[i].first >= work[i].end {
				continue
			}
			c := work[i].items[work[i].end-1]
			if margin := c.research - c.industry; margin < bestMargin {
				bestMargin, bestColony = margin, i
			}
		}
		if bestColony < 0 {
			break
		}
		w := &work[bestColony]
		w.end--
		originalAIMoveCandidate(&out[bestColony], w.items[w.end], int(gamedata.WORKER))
	}
	pressure := applyOriginalAIAdditionalFarmers(ps, out, ctx.ColonyBlockaded, cleanupBefore)
	preserveAIResearchForCollapsedQsortTies(ps, out, ctx.ColonyBlockaded)
	return out, pressure, true
}

func originalIntegerSqrt(n int) int {
	root := 0
	for (root+1)*(root+1) <= n {
		root++
	}
	return root
}
