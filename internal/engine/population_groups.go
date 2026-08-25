package engine

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// PopulationGroupsComplete 回報逐 slot 群組是否能作為目前殖民地的人口真相。
func PopulationGroupsComplete(c ColonyState) bool { return populationGroupsValid(c) }

func groupJobCount(g *PopulationGroup, job gamedata.ColonistJob) *int {
	switch job {
	case gamedata.FARMER:
		return &g.Farmers
	case gamedata.WORKER:
		return &g.Workers
	case gamedata.SCIENTIST:
		return &g.Scientists
	default:
		return nil
	}
}

func groupPrisonerCount(g *PopulationGroup, job gamedata.ColonistJob) *int {
	switch job {
	case gamedata.FARMER:
		return &g.PrisonerFarmers
	case gamedata.WORKER:
		return &g.PrisonerWorkers
	case gamedata.SCIENTIST:
		return &g.PrisonerScientists
	default:
		return nil
	}
}

func populationGroupTotals(c *ColonyState) (farmers, workers, scientists int, ok bool) {
	if c == nil || len(c.PopulationGroups) == 0 || !c.OwnerRaceProfileKnown {
		return 0, 0, 0, false
	}
	for _, g := range c.PopulationGroups {
		if !g.RaceSlotKnown || !g.ProfileKnown || g.Farmers < 0 || g.Workers < 0 || g.Scientists < 0 ||
			g.PrisonerFarmers < 0 || g.PrisonerFarmers > g.Farmers ||
			g.PrisonerWorkers < 0 || g.PrisonerWorkers > g.Workers ||
			g.PrisonerScientists < 0 || g.PrisonerScientists > g.Scientists {
			return 0, 0, 0, false
		}
		farmers += g.Farmers
		workers += g.Workers
		scientists += g.Scientists
	}
	return farmers, workers, scientists, true
}

// ShiftPopulationGroupJob 以固定群組順序移動一人；若來源有 prisoner，優先移 prisoner。
func ShiftPopulationGroupJob(c *ColonyState, from, to gamedata.ColonistJob) bool {
	if c == nil || !populationGroupsValid(*c) || from == to {
		return false
	}
	pick := -1
	for i := range c.PopulationGroups {
		if p := groupPrisonerCount(&c.PopulationGroups[i], from); p != nil && *p > 0 {
			pick = i
			break
		}
	}
	if pick < 0 {
		for i := range c.PopulationGroups {
			if n := groupJobCount(&c.PopulationGroups[i], from); n != nil && *n > 0 {
				pick = i
				break
			}
		}
	}
	if pick < 0 {
		return false
	}
	g := &c.PopulationGroups[pick]
	fromN, toN := groupJobCount(g, from), groupJobCount(g, to)
	if fromN == nil || toN == nil || *fromN <= 0 {
		return false
	}
	if p := groupPrisonerCount(g, from); p != nil && *p > 0 {
		*p--
		*groupPrisonerCount(g, to)++
	}
	*fromN--
	*toN++
	return true
}

// MarkPopulationGroupsPrisoners 在征服時保留 race slot，只把所有人口標成 prisoner。
func MarkPopulationGroupsPrisoners(c *ColonyState) {
	if c == nil || !populationGroupsValid(*c) {
		return
	}
	for i := range c.PopulationGroups {
		g := &c.PopulationGroups[i]
		g.PrisonerFarmers, g.PrisonerWorkers, g.PrisonerScientists = g.Farmers, g.Workers, g.Scientists
	}
}

// ClearPopulationGroupPrisoners 用於心靈控制等立即完成同化的路徑。
func ClearPopulationGroupPrisoners(c *ColonyState) {
	if c == nil {
		return
	}
	for i := range c.PopulationGroups {
		c.PopulationGroups[i].PrisonerFarmers = 0
		c.PopulationGroups[i].PrisonerWorkers = 0
		c.PopulationGroups[i].PrisonerScientists = 0
	}
}

// PopulationGroupPrisoners 回傳 typed groups 目前的逐職務 prisoner 總數。
func PopulationGroupPrisoners(c ColonyState) (farmers, workers, scientists int, ok bool) {
	_, _, _, ok = populationGroupTotals(&c)
	if !ok {
		return 0, 0, 0, false
	}
	for _, g := range c.PopulationGroups {
		farmers += g.PrisonerFarmers
		workers += g.PrisonerWorkers
		scientists += g.PrisonerScientists
	}
	return farmers, workers, scientists, true
}

// AssimilateOnePopulationGroup 依職務再依群組的固定順序清除一名 prisoner，不改 race slot。
func AssimilateOnePopulationGroup(c *ColonyState) bool {
	if c == nil || !populationGroupsValid(*c) {
		return false
	}
	for _, job := range []gamedata.ColonistJob{gamedata.FARMER, gamedata.WORKER, gamedata.SCIENTIST} {
		for i := range c.PopulationGroups {
			if p := groupPrisonerCount(&c.PopulationGroups[i], job); p != nil && *p > 0 {
				*p--
				return true
			}
		}
	}
	return false
}

// AddPopulationGroupUnit 把成長人口加入指定群組。
func AddPopulationGroupUnit(c *ColonyState, group int, job gamedata.ColonistJob) bool {
	if c == nil || group < 0 || group >= len(c.PopulationGroups) {
		return false
	}
	f, w, s, ok := populationGroupTotals(c)
	if !ok || f+w+s+1 != c.Population || c.Farmers+c.Workers+c.Scientists != c.Population {
		return false
	}
	if n := groupJobCount(&c.PopulationGroups[group], job); n != nil {
		*n++
		return true
	}
	return false
}

// AddOwnerPopulationGroupUnit 是舊呼叫端的 owner slot 包裝。
func AddOwnerPopulationGroupUnit(c *ColonyState, job gamedata.ColonistJob) bool {
	if c == nil || !c.OwnerRaceSlotKnown {
		return false
	}
	for i := range c.PopulationGroups {
		g := &c.PopulationGroups[i]
		if g.RaceSlotKnown && g.RaceSlot == c.OwnerRaceSlot {
			return AddPopulationGroupUnit(c, i, job)
		}
	}
	return false
}

// PopulationGroupLimit 回傳指定群組在目前星球的 race-specific 人口上限。
func PopulationGroupLimit(c ColonyState, group int) int {
	if group < 0 || group >= len(c.PopulationGroups) || !populationGroupsValid(c) {
		return 0
	}
	return groupPopulationLimit(c, c.PopulationGroups[group])
}

// RemovePopulationGroupUnit 從指定職務同步扣一人；優先扣 prisoner，再按固定群組順序。
func RemovePopulationGroupUnit(c *ColonyState, job gamedata.ColonistJob) bool {
	if c == nil {
		return false
	}
	f, w, s, ok := populationGroupTotals(c)
	if !ok || f != c.Farmers || w != c.Workers || s != c.Scientists {
		return false
	}
	pick := -1
	for i := range c.PopulationGroups {
		if p := groupPrisonerCount(&c.PopulationGroups[i], job); p != nil && *p > 0 {
			pick = i
			break
		}
	}
	if pick < 0 {
		for i := range c.PopulationGroups {
			if n := groupJobCount(&c.PopulationGroups[i], job); n != nil && *n > 0 {
				pick = i
				break
			}
		}
	}
	if pick < 0 {
		return false
	}
	g := &c.PopulationGroups[pick]
	if p := groupPrisonerCount(g, job); p != nil && *p > 0 {
		*p--
	}
	*groupJobCount(g, job)--
	return true
}
