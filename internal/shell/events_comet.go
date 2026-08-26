package shell

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func originalCometStrength(difficulty, roll1Based int) int {
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
	return 10 * (roll1Based + 10 + difficulty)
}

func originalCometCountdown(difficulty, roll1Based int) int {
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

func originalCometImpactDamage(population, rawBuildingCount, rollA, rollB int) int {
	if population < 0 {
		population = 0
	}
	if rawBuildingCount < 0 {
		rawBuildingCount = 0
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
	damage := (population + rawBuildingCount) * (rollA + rollB) / 10
	if damage < 1 {
		return 1
	}
	return damage
}

func (s *GameSession) planetCometActive(planetIndex int) bool {
	for i := range s.PersistentEvents {
		if s.PersistentEvents[i].Kind == PersistentComet && s.PersistentEvents[i].PlanetIndex == planetIndex {
			return true
		}
	}
	return false
}

func (s *GameSession) cometTargetConflicted(planetIndex, starIndex int) bool {
	for i := range s.PersistentEvents {
		e := &s.PersistentEvents[i]
		if e.PlanetIndex == planetIndex && (e.Kind == PersistentComet || e.Kind == PersistentPlague || e.Kind == PersistentPopulationBoom) {
			return true
		}
		if e.StarIndex == starIndex && (e.Kind == PersistentSupernova || e.Kind == PersistentStasis || e.Kind == PersistentPirateActivity) {
			return true
		}
	}
	return false
}

func (s *GameSession) appendComet(planet, star int) PersistentEvent {
	strength := originalCometStrength(s.Difficulty, s.eventRand.Intn(5)+1)
	e := PersistentEvent{Kind: PersistentComet, PlanetIndex: planet, StarIndex: star,
		Countdown: originalCometCountdown(s.Difficulty, s.eventRand.Intn(5)+1),
		Strength:  strength, InitialStrength: strength}
	s.PersistentEvents = append(s.PersistentEvents, e)
	return e
}

func (s *GameSession) startPlayerComet() (string, bool) {
	i, ok := pickEarthquakeColony(s.PlayerColonies,
		func(i int) bool { return !colonyHasCapitol(s.ColonyBuildings, i) }, s.eventRand.Intn)
	if !ok {
		return "", false
	}
	planet, star := s.ColonyPlanetIndex(i), s.PlayerColonyStarIndex(i)
	if planet < 0 || star < 0 || s.cometTargetConflicted(planet, star) {
		return "", false
	}
	e := s.appendComet(planet, star)
	return fmt.Sprintf("一顆彗星正朝 %s 撞來，預計 %d 回合後抵達；彗星耐久 %d，星系內艦艇已開始攔截",
		s.colonyLabel(i), e.Countdown, e.Strength), true
}

func (s *GameSession) startAIComet(aiIndex int) (string, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return "", false
	}
	a := &s.AIPlayers[aiIndex]
	i, ok := pickEarthquakeColony(a.Colonies,
		func(i int) bool { return !colonyHasCapitol(a.ColonyBuildings, i) }, s.eventRand.Intn)
	if !ok || i >= len(a.ColonyPlanets) || i >= len(a.ColonyStars) {
		return "", false
	}
	planet, star := a.ColonyPlanets[i], a.ColonyStars[i]
	if planet < 0 || star < 0 || s.cometTargetConflicted(planet, star) {
		return "", false
	}
	e := s.appendComet(planet, star)
	return fmt.Sprintf("一顆彗星正朝%s的殖民地撞來，預計 %d 回合後抵達；彗星耐久 %d",
		stripAILabel(a.Name), e.Countdown, e.Strength), true
}

func cometFleetStrength(fleets []Fleet, star int) int {
	total := 0
	for i := range fleets {
		if fleets[i].ETA != 0 || fleets[i].AtStar != star {
			continue
		}
		for j := range fleets[i].Ships {
			total += int(shipSizeClass(fleets[i].Ships[j].Class)) + 1
		}
	}
	return total
}

// cometInterceptionStrength 對應 sub_23B28 的「不檢查 owner」掃描。
func (s *GameSession) cometInterceptionStrength(star int) int {
	total := cometFleetStrength(s.Fleets, star)
	if s.HotseatEnabled() {
		for i := range s.Seats {
			if i != s.ActiveSeat {
				total += cometFleetStrength(s.Seats[i].Fleets, star)
			}
		}
	}
	for i := range s.AIPlayers {
		a := &s.AIPlayers[i]
		if a.FleetETA == 0 && a.FleetStar == star {
			for j := range a.Ships {
				total += int(shipSizeClass(a.Ships[j].Class)) + 1
			}
		}
	}
	return total
}

type cometTargetKind uint8

const (
	cometPlayer cometTargetKind = iota
	cometSeat
	cometAI
)

type cometTarget struct {
	kind           cometTargetKind
	empire, colony int
}

func colonyByPlanet(colonies []engine.ColonyState, planets []int, planet int) int {
	for i := range colonies {
		if i < len(planets) && planets[i] == planet {
			return i
		}
	}
	return -1
}

func (s *GameSession) findCometTarget(planet int) (cometTarget, bool) {
	if i := colonyByPlanet(s.PlayerColonies, s.PlayerColonyPlanets, planet); i >= 0 {
		return cometTarget{kind: cometPlayer, colony: i}, true
	}
	if s.HotseatEnabled() {
		for i := range s.Seats {
			if i == s.ActiveSeat {
				continue
			}
			if c := colonyByPlanet(s.Seats[i].PlayerColonies, s.Seats[i].PlayerColonyPlanets, planet); c >= 0 {
				return cometTarget{kind: cometSeat, empire: i, colony: c}, true
			}
		}
	}
	for i := range s.AIPlayers {
		if c := colonyByPlanet(s.AIPlayers[i].Colonies, s.AIPlayers[i].ColonyPlanets, planet); c >= 0 {
			return cometTarget{kind: cometAI, empire: i, colony: c}, true
		}
	}
	return cometTarget{}, false
}

func resolveCometImpact(c *engine.ColonyState, buildings map[string]bool, marines, tanks int,
	player engine.PlayerState, highG bool, rng *randStream) gamedata.StrategicColonyDamageResult {
	damage := originalCometImpactDamage(c.Population, len(originalColonyBuildingIDs(buildings)), rng.Intn(3)+1, rng.Intn(3)+1)
	result := gamedata.ResolveStrategicColonyDamage(gamedata.StrategicColonyDamageState{
		Population: c.Population, LastPopulationPoints: c.BombardmentLastPopulationPoints,
		Marines: marines, Tanks: tanks, BuildProgress: c.BombardmentBuildProgress,
		RawBuildingIDs: originalColonyBuildingIDs(buildings),
		MarineHitCost:  gamedata.GroundMarineHitsToKill(highG, hasPoweredArmorFor(player)),
		TankHitCost:    tankHitsToKillFor(player, highG), BuildingHitCost: 1,
	}, damage, rng.Intn)
	c.Population, c.BombardmentLastPopulationPoints = result.State.Population, result.State.LastPopulationPoints
	c.BombardmentBuildProgress = result.State.BuildProgress
	normalizeColonyJobsAfterPopulationLoss(c)
	return result
}

func (s *GameSession) applyLoadedPlayerCometImpact(i int) (gamedata.StrategicColonyDamageResult, bool) {
	if i < 0 || i >= len(s.PlayerColonies) {
		return gamedata.StrategicColonyDamageResult{}, false
	}
	star := s.PlayerColonyStarIndex(i)
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
	result := resolveCometImpact(&s.PlayerColonies[i], buildings, marines, tanks, s.Player, s.raceHasTrait(gamedata.TRAIT_HIGH_G), s.eventRand)
	for _, removed := range removeDestroyedRawBuildings(buildings, result.DestroyedBuildingIDs) {
		s.removeBuildingEffect(i, removed)
	}
	if i < len(s.PlayerColonyMarines) {
		s.PlayerColonyMarines[i] = result.State.Marines
	}
	if i < len(s.PlayerColonyTanks) {
		s.PlayerColonyTanks[i] = result.State.Tanks
	}
	if result.ColonyDestroyed {
		s.removePlayerColony(i)
		if star >= 0 && star < len(s.Stars) && !s.playerHasColonyAtStar(star) {
			s.Stars[star].Owner = 0
		}
	} else {
		s.recalcColonyMorale(i)
	}
	return result, true
}

func (s *GameSession) applyCometImpact(target cometTarget) (gamedata.StrategicColonyDamageResult, bool) {
	switch target.kind {
	case cometSeat:
		if target.empire < 0 || target.empire >= len(s.Seats) {
			return gamedata.StrategicColonyDamageResult{}, false
		}
		active := s.ActiveSeat
		if active >= 0 && active < len(s.Seats) {
			s.Seats[active] = s.saveSeat()
		}
		s.loadSeat(s.Seats[target.empire])
		result, ok := s.applyLoadedPlayerCometImpact(target.colony)
		s.Seats[target.empire] = s.saveSeat()
		if active >= 0 && active < len(s.Seats) {
			s.loadSeat(s.Seats[active])
		}
		return result, ok
	case cometAI:
		a := &s.AIPlayers[target.empire]
		ensureAIGroundForceSlots(a)
		if target.colony < 0 || target.colony >= len(a.Colonies) {
			return gamedata.StrategicColonyDamageResult{}, false
		}
		star := -1
		if target.colony < len(a.ColonyStars) {
			star = a.ColonyStars[target.colony]
		}
		var buildings map[string]bool
		if target.colony < len(a.ColonyBuildings) {
			buildings = a.ColonyBuildings[target.colony]
		}
		result := resolveCometImpact(&a.Colonies[target.colony], buildings, a.ColonyMarines[target.colony], a.ColonyTanks[target.colony], a.Player, aiRaceHasTrait(*a, gamedata.TRAIT_HIGH_G), s.eventRand)
		for _, removed := range removeDestroyedRawBuildings(buildings, result.DestroyedBuildingIDs) {
			removeBuildingEffectFromColony(&a.Colonies[target.colony], removed)
		}
		a.ColonyMarines[target.colony], a.ColonyTanks[target.colony] = result.State.Marines, result.State.Tanks
		if result.ColonyDestroyed {
			s.removeAIColonyAfterEvent(target.empire, target.colony)
			if star >= 0 && star < len(s.Stars) && !s.anyAIHasColonyAtStar(star) {
				s.Stars[star].Owner = 0
			}
		}
		return result, true
	default:
		return s.applyLoadedPlayerCometImpact(target.colony)
	}
}

func (s *GameSession) stepComet(e *PersistentEvent) (bool, string, string) {
	target, ok := s.findCometTarget(e.PlanetIndex)
	if !ok {
		return true, "彗星事件因目標殖民地不復存在而結束", "The comet event ended because its target colony no longer exists."
	}
	intercept := s.cometInterceptionStrength(e.StarIndex)
	e.Strength -= intercept
	if e.Strength <= 0 {
		return true, fmt.Sprintf("%s 星系的艦隊已完全摧毀來襲彗星", s.starName(e.StarIndex)),
			fmt.Sprintf("Ships in the %s system completely destroyed the incoming comet.", s.starNameEN(e.StarIndex))
	}
	e.Countdown--
	progressRoll := s.eventRoll(20)
	if e.Countdown <= 0 {
		result, applied := s.applyCometImpact(target)
		if !applied {
			return true, "彗星撞擊前失去有效目標，事件結束", "The comet lost its valid target before impact."
		}
		return true, fmt.Sprintf("彗星撞擊 %s 星系：%d 百萬居民、%d 支陸戰隊、%d 支裝甲部隊傷亡，%d 棟建築毀損",
				s.starName(e.StarIndex), result.PopulationLost, result.MarinesLost, result.TanksLost, len(result.DestroyedBuildingIDs)),
			fmt.Sprintf("The comet struck the %s system: %d million residents, %d marines, and %d armor units were lost; %d buildings were destroyed.",
				s.starNameEN(e.StarIndex), result.PopulationLost, result.MarinesLost, result.TanksLost, len(result.DestroyedBuildingIDs))
	}
	if progressRoll == 1 && e.Strength != e.InitialStrength {
		percent := (e.InitialStrength - e.Strength) * 100 / e.InitialStrength
		return false, fmt.Sprintf("%s 星系彗星攔截進度 %d%%，剩餘 %d 回合", s.starName(e.StarIndex), percent, e.Countdown),
			fmt.Sprintf("Comet interception in the %s system is %d%% complete; %d turns remain.", s.starNameEN(e.StarIndex), percent, e.Countdown)
	}
	return false, "", ""
}
