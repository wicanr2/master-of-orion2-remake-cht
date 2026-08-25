package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func (s *GameSession) populationRandForTurn() *randStream {
	if s.populationRand == nil {
		s.populationRand = newRandStream(s.EventSeed*2654435761 + 23)
	}
	return s.populationRand
}

func populationGroupUnits(g engine.PopulationGroup) int { return g.Farmers + g.Workers + g.Scientists }

func populationGroupIndexBySlot(c engine.ColonyState, slot int) int {
	for i, g := range c.PopulationGroups {
		if g.RaceSlotKnown && g.RaceSlot == slot {
			return i
		}
	}
	return -1
}

func (s *GameSession) activePopulationRaceSlots() int {
	maxSlot := 0
	for _, c := range s.PlayerColonies {
		if c.OwnerRaceSlotKnown && c.OwnerRaceSlot >= 0 && c.OwnerRaceSlot < gamedata.AndroidColonistSlot && c.OwnerRaceSlot > maxSlot {
			maxSlot = c.OwnerRaceSlot
		}
	}
	for _, a := range s.AIPlayers {
		if a.PopulationRaceSlotKnown && a.PopulationRaceSlot >= 0 &&
			a.PopulationRaceSlot < gamedata.AndroidColonistSlot && a.PopulationRaceSlot > maxSlot {
			maxSlot = a.PopulationRaceSlot
		}
	}
	return maxSlot + 1
}

// positivePopulationSlotOrder 對應 sub_E2DCA → sub_FE9F5：owner 第一，其餘做
// 1-based Random 的 Fisher–Yates；最後一格 Random(1) 仍消耗一次亂數。
func positivePopulationSlotOrder(owner, activeSlots int, rng *randStream) []int {
	if activeSlots < 1 {
		activeSlots = 1
	}
	if activeSlots > gamedata.AndroidColonistSlot {
		activeSlots = gamedata.AndroidColonistSlot
	}
	order := make([]int, activeSlots)
	for i := range order {
		order[i] = i
	}
	if owner >= 0 && owner < activeSlots {
		order[0], order[owner] = order[owner], order[0]
	}
	for i := 1; i < len(order); i++ {
		pick := i
		if rng != nil {
			pick += rng.Intn(len(order) - i)
		}
		order[i], order[pick] = order[pick], order[i]
	}
	return order
}

// protectedPopulationGroup 對應 sub_E2DCA @ 0xE2F24..0xE2F7F：殖民地永遠留一人。
func protectedPopulationGroup(c engine.ColonyState) int {
	if c.OwnerRaceSlotKnown {
		if i := populationGroupIndexBySlot(c, c.OwnerRaceSlot); i >= 0 && populationGroupUnits(c.PopulationGroups[i]) > 0 {
			return i
		}
	}
	pick, most := -1, 0
	for slot := gamedata.AndroidColonistSlot - 1; slot >= 0; slot-- { // 下降掃描使同數保留較高 slot。
		if i := populationGroupIndexBySlot(c, slot); i >= 0 {
			if n := populationGroupUnits(c.PopulationGroups[i]); n > most {
				pick, most = i, n
			}
		}
	}
	if pick >= 0 {
		return pick
	}
	if i := populationGroupIndexBySlot(c, gamedata.AndroidColonistSlot); i >= 0 && populationGroupUnits(c.PopulationGroups[i]) > 0 {
		return i
	}
	if i := populationGroupIndexBySlot(c, gamedata.NativeColonistSlot); i >= 0 && populationGroupUnits(c.PopulationGroups[i]) > 0 {
		return i
	}
	return -1
}

type mortalityCandidate struct {
	job      gamedata.ColonistJob
	prisoner bool
}

// mortalityReservoirCandidate 保留原版「非農夫優先 + 每候選 Random(count)==1」契約。
// typed groups 不保存 packed 陣列順序，這裡以職務、合作／prisoner 的固定順序重播同一分布。
func mortalityReservoirCandidate(g engine.PopulationGroup, rng *randStream) (mortalityCandidate, bool) {
	jobs := []gamedata.ColonistJob{gamedata.WORKER, gamedata.SCIENTIST}
	if g.Workers+g.Scientists == 0 {
		jobs = []gamedata.ColonistJob{gamedata.FARMER}
	}
	chosen, count, ok := mortalityCandidate{}, 0, false
	for _, job := range jobs {
		total, prisoners := 0, 0
		switch job {
		case gamedata.FARMER:
			total, prisoners = g.Farmers, g.PrisonerFarmers
		case gamedata.WORKER:
			total, prisoners = g.Workers, g.PrisonerWorkers
		case gamedata.SCIENTIST:
			total, prisoners = g.Scientists, g.PrisonerScientists
		}
		for i := 0; i < total; i++ {
			count++
			candidate := mortalityCandidate{job: job, prisoner: i >= total-prisoners}
			if rng == nil || rng.Intn(count) == 0 {
				chosen, ok = candidate, true
			}
		}
	}
	return chosen, ok
}

func removeNegativeGrowthColonist(c *engine.ColonyState, group int, rng *randStream) bool {
	if c == nil || group < 0 || group >= len(c.PopulationGroups) || c.Population <= 0 {
		return false
	}
	pick, ok := mortalityReservoirCandidate(c.PopulationGroups[group], rng)
	if !ok {
		return false
	}
	return removePopulationGroupCandidate(c, group, pick)
}

// removePopulationGroupCandidate 同步逐 race 群組、總職務與 prisoner 計數。
// 事件與負成長共用此最小 mutation，避免只改 aggregate 而遺失混居人口真相。
func removePopulationGroupCandidate(c *engine.ColonyState, group int, pick mortalityCandidate) bool {
	if c == nil || group < 0 || group >= len(c.PopulationGroups) || c.Population <= 0 {
		return false
	}
	g := &c.PopulationGroups[group]
	var groupJob, groupPrisoners, colonyJob, colonyPrisoners *int
	switch pick.job {
	case gamedata.FARMER:
		groupJob, groupPrisoners = &g.Farmers, &g.PrisonerFarmers
		colonyJob, colonyPrisoners = &c.Farmers, &c.UnassimilatedFarmers
	case gamedata.WORKER:
		groupJob, groupPrisoners = &g.Workers, &g.PrisonerWorkers
		colonyJob, colonyPrisoners = &c.Workers, &c.UnassimilatedWorkers
	case gamedata.SCIENTIST:
		groupJob, groupPrisoners = &g.Scientists, &g.PrisonerScientists
		colonyJob, colonyPrisoners = &c.Scientists, &c.UnassimilatedScientists
	default:
		return false
	}
	if *groupJob <= 0 || *colonyJob <= 0 {
		return false
	}
	*groupJob--
	*colonyJob--
	c.Population--
	if pick.prisoner && *groupPrisoners > 0 {
		*groupPrisoners--
		if *colonyPrisoners > 0 {
			*colonyPrisoners--
		}
		if c.UnassimilatedPop > 0 {
			c.UnassimilatedPop--
		}
	}
	return true
}
