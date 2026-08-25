package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

type industrialAccidentImpact struct {
	ColonyIndex, PopulationLost, MarinesLost, TanksLost, BuildingsDestroyed int
	ColonyName                                                              string
	ColonyDestroyed                                                         bool
}

func originalIndustrialAccidentHits(population, rollA, rollB int) int {
	if population < 0 {
		population = 0
	}
	if rollA < 1 {
		rollA = 1
	} else if rollA > 3 {
		rollA = 3
	}
	if rollB < 1 {
		rollB = 1
	} else if rollB > 3 {
		rollB = 3
	}
	return population * (rollA + rollB) / 10
}

// originalIndustrialAccidentColony 對應 sub_231B4：200 次拒絕抽樣後，原版
// fallback 持續覆寫結果，所以取得最高索引合格殖民地。
func originalIndustrialAccidentColony(colonies []engine.ColonyState, rng *randStream) (int, bool) {
	if len(colonies) == 0 || rng == nil {
		return 0, false
	}
	for attempt := 0; attempt < 200; attempt++ {
		i := reservoirEventColony(colonies, rng)
		if i >= 0 && colonies[i].Climate > gamedata.RADIATED {
			return i, true
		}
	}
	pick := -1
	for i := range colonies {
		if colonies[i].Climate > gamedata.RADIATED {
			pick = i
		}
	}
	return pick, pick >= 0
}

// industrialColonistCandidate 保留非 Android 的候選集合與 reservoir 分布。
// typed 群組沒有 packed 順序，固定以群組及 Farmer/Worker/Scientist 順序重播。
func industrialColonistCandidate(c *engine.ColonyState, rng *randStream) (int, mortalityCandidate, bool) {
	if c == nil || rng == nil || !engine.PopulationGroupsComplete(*c) {
		return -1, mortalityCandidate{}, false
	}
	group, chosen, count, ok := -1, mortalityCandidate{}, 0, false
	for gi, g := range c.PopulationGroups {
		if g.RaceSlotKnown && g.RaceSlot == gamedata.AndroidColonistSlot {
			continue
		}
		jobs := []struct {
			job              gamedata.ColonistJob
			total, prisoners int
		}{
			{gamedata.FARMER, g.Farmers, g.PrisonerFarmers},
			{gamedata.WORKER, g.Workers, g.PrisonerWorkers},
			{gamedata.SCIENTIST, g.Scientists, g.PrisonerScientists},
		}
		for _, j := range jobs {
			for n := 0; n < j.total; n++ {
				count++
				candidate := mortalityCandidate{job: j.job, prisoner: n >= j.total-j.prisoners}
				if rng.Intn(count) == 0 {
					group, chosen, ok = gi, candidate, true
				}
			}
		}
	}
	return group, chosen, ok
}

func removeIndustrialAggregateColonist(c *engine.ColonyState, rng *randStream) bool {
	if c == nil || c.Population <= 0 {
		return false
	}
	total := c.Farmers + c.Workers + c.Scientists
	if total <= 0 {
		return false
	}
	pick := rng.Intn(total)
	job := gamedata.SCIENTIST
	if pick < c.Farmers {
		job = gamedata.FARMER
	} else if pick < c.Farmers+c.Workers {
		job = gamedata.WORKER
	}
	engine.RemovePopulationGroupUnit(c, job)
	switch job {
	case gamedata.FARMER:
		c.Farmers--
	case gamedata.WORKER:
		c.Workers--
	default:
		c.Scientists--
	}
	c.Population--
	return true
}

func removeIndustrialSelectedColonist(c *engine.ColonyState, rng *randStream) bool {
	group, candidate, ok := industrialColonistCandidate(c, rng)
	if ok {
		return removePopulationGroupCandidate(c, group, candidate)
	}
	// 完整 typed 群組卻沒有候選，代表全 Android；該特殊命中原版直接浪費。
	if c != nil && engine.PopulationGroupsComplete(*c) {
		return false
	}
	return removeIndustrialAggregateColonist(c, rng)
}

func resolveIndustrialSpecialHits(c *engine.ColonyState, marines, tanks, hits int, rng *randStream) (popLost, marinesLost, tanksLost int) {
	for ; hits > 0; hits-- {
		group, candidate, hasNonAndroid := industrialColonistCandidate(c, rng)
		if !hasNonAndroid && c != nil && engine.PopulationGroupsComplete(*c) {
			continue
		}
		removeSelected := func() bool {
			if hasNonAndroid {
				return removePopulationGroupCandidate(c, group, candidate)
			}
			return removeIndustrialAggregateColonist(c, rng)
		}
		if c.Population >= 2 {
			if rng.Intn(c.Population+marines+tanks)+1 <= c.Population {
				if removeSelected() {
					popLost++
				}
				continue
			}
			if marines > 0 && rng.Intn(marines+tanks)+1 <= marines {
				marines--
				marinesLost++
			} else if tanks > 0 {
				tanks--
				tanksLost++
			}
			continue
		}
		if marines > 0 {
			if rng.Intn(marines+tanks)+1 <= marines {
				marines--
				marinesLost++
			} else if tanks > 0 {
				tanks--
				tanksLost++
			}
			continue
		}
		if tanks > 0 {
			tanks--
			tanksLost++
			continue
		}
		if c.BombardmentLastPopulationPoints > 100 {
			c.BombardmentLastPopulationPoints -= 100
		} else if removeSelected() {
			c.BombardmentLastPopulationPoints = 0
			popLost++
		}
	}
	return
}

func resolveIndustrialAccident(c *engine.ColonyState, buildings map[string]bool, marines, tanks int,
	player engine.PlayerState, highG bool, rng *randStream) (gamedata.StrategicColonyDamageResult, int, int, int) {
	hits := originalIndustrialAccidentHits(c.Population, rng.Intn(3)+1, rng.Intn(3)+1)
	popLost, marineLost, tankLost := resolveIndustrialSpecialHits(c, marines, tanks, hits, rng)
	regular := gamedata.ResolveStrategicColonyDamage(gamedata.StrategicColonyDamageState{
		Population: c.Population, LastPopulationPoints: c.BombardmentLastPopulationPoints,
		Marines: marines - marineLost, Tanks: tanks - tankLost, BuildProgress: c.BombardmentBuildProgress,
		RawBuildingIDs: originalColonyBuildingIDs(buildings),
		MarineHitCost:  gamedata.GroundMarineHitsToKill(highG, hasPoweredArmorFor(player)),
		TankHitCost:    tankHitsToKillFor(player, highG), BuildingHitCost: 1,
	}, 1, rng.Intn)
	c.Population = regular.State.Population
	c.BombardmentLastPopulationPoints = regular.State.LastPopulationPoints
	c.BombardmentBuildProgress = regular.State.BuildProgress
	normalizeColonyJobsAfterPopulationLoss(c)
	return regular, popLost + regular.PopulationLost, marineLost + regular.MarinesLost, tankLost + regular.TanksLost
}

func (s *GameSession) applyPlayerIndustrialAccident() (industrialAccidentImpact, bool) {
	if s.RaceTolerant() {
		return industrialAccidentImpact{}, false
	}
	i, ok := originalIndustrialAccidentColony(s.PlayerColonies, s.eventRand)
	if !ok {
		return industrialAccidentImpact{}, false
	}
	name, star := s.colonyLabel(i), s.PlayerColonyStarIndex(i)
	var buildings map[string]bool
	if i < len(s.ColonyBuildings) {
		buildings = s.ColonyBuildings[i]
	}
	marines, tanks := 0, 0
	if i < len(s.PlayerColonyMarines) {
		marines = s.PlayerColonyMarines[i]
	}
	if i < len(s.PlayerColonyTanks) {
		tanks = s.PlayerColonyTanks[i]
	}
	regular, popLost, marineLost, tankLost := resolveIndustrialAccident(&s.PlayerColonies[i], buildings, marines, tanks, s.Player, s.raceHasTrait(gamedata.TRAIT_HIGH_G), s.eventRand)
	for _, removed := range removeDestroyedRawBuildings(buildings, regular.DestroyedBuildingIDs) {
		s.removeBuildingEffect(i, removed)
	}
	if i < len(s.PlayerColonyMarines) {
		s.PlayerColonyMarines[i] = regular.State.Marines
	}
	if i < len(s.PlayerColonyTanks) {
		s.PlayerColonyTanks[i] = regular.State.Tanks
	}
	destroyed := regular.ColonyDestroyed || (s.PlayerColonies[i].Population == 0 && len(originalColonyBuildingIDs(buildings)) == 0)
	impact := industrialAccidentImpact{i, popLost, marineLost, tankLost, len(regular.DestroyedBuildingIDs), name, destroyed}
	if destroyed {
		s.removePlayerColony(i)
		if star >= 0 && star < len(s.Stars) && !s.playerHasColonyAtStar(star) {
			s.Stars[star].Owner = 0
		}
	} else {
		s.recalcColonyMorale(i)
	}
	return impact, true
}

func (s *GameSession) applyAIIndustrialAccident(aiIndex int) (industrialAccidentImpact, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) || aiRaceHasTrait(s.AIPlayers[aiIndex], gamedata.TRAIT_TOLERANT) {
		return industrialAccidentImpact{}, false
	}
	a := &s.AIPlayers[aiIndex]
	i, ok := originalIndustrialAccidentColony(a.Colonies, s.eventRand)
	if !ok {
		return industrialAccidentImpact{}, false
	}
	ensureAIGroundForceSlots(a)
	star := -1
	if i < len(a.ColonyStars) {
		star = a.ColonyStars[i]
	}
	var buildings map[string]bool
	if i < len(a.ColonyBuildings) {
		buildings = a.ColonyBuildings[i]
	}
	regular, popLost, marineLost, tankLost := resolveIndustrialAccident(&a.Colonies[i], buildings, a.ColonyMarines[i], a.ColonyTanks[i], a.Player, aiRaceHasTrait(*a, gamedata.TRAIT_HIGH_G), s.eventRand)
	for _, removed := range removeDestroyedRawBuildings(buildings, regular.DestroyedBuildingIDs) {
		removeBuildingEffectFromColony(&a.Colonies[i], removed)
	}
	a.ColonyMarines[i], a.ColonyTanks[i] = regular.State.Marines, regular.State.Tanks
	destroyed := regular.ColonyDestroyed || (a.Colonies[i].Population == 0 && len(originalColonyBuildingIDs(buildings)) == 0)
	impact := industrialAccidentImpact{i, popLost, marineLost, tankLost, len(regular.DestroyedBuildingIDs), stripAILabel(a.Name), destroyed}
	if destroyed {
		s.removeAIColonyAfterEvent(aiIndex, i)
		if star >= 0 && star < len(s.Stars) && !s.anyAIHasColonyAtStar(star) {
			s.Stars[star].Owner = 0
		}
	}
	return impact, true
}
