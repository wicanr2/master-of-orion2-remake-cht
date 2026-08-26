package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// earthquakeImpact 是一次原版事件 7 結算後可供 GNN 與測試使用的玩家可見結果。
type earthquakeImpact struct {
	ColonyIndex        int
	ColonyName         string
	PopulationLost     int
	BuildingsDestroyed int
	MarinesLost        int
	TanksLost          int
	BuildProgressLost  int
	ColonyDestroyed    bool
}

// originalEarthquakeDamage 對應 sub_238A8 @ 0x2391A..0x2394C。
// roll3／roll2 是原版 Random(3)／Random(2) 的 1-based 結果。
func originalEarthquakeDamage(population, rawBuildingCount, roll3, roll2 int) int {
	if population < 0 {
		population = 0
	}
	if rawBuildingCount < 0 {
		rawBuildingCount = 0
	}
	if roll3 < 1 {
		roll3 = 1
	} else if roll3 > 3 {
		roll3 = 3
	}
	if roll2 < 1 {
		roll2 = 1
	} else if roll2 > 2 {
		roll2 = 2
	}
	damage := (population + rawBuildingCount) * (roll3 + roll2) / 10
	if damage < 1 {
		return 1
	}
	return damage
}

// pickEarthquakeColony 重製 sub_23DA0 的逐索引 reservoir sampling。remake 的 colony
// slice 本身只保存有效殖民地；raw +0x13F 已證實為 Capitol 建築槽，但此 helper
// 尚未接收建築狀態，因此排除條件列在 WORKLIST 的事件目標 filter 待辦。
func pickEarthquakeColony(colonies []engine.ColonyState, intn func(int) int) (int, bool) {
	pick, seen := -1, 0
	for i := range colonies {
		seen++
		if intn == nil || intn(seen) == 0 {
			pick = i
		}
	}
	return pick, pick >= 0
}

func earthquakeRawBuildingIDs(buildings map[string]bool) []int {
	return originalColonyBuildingIDs(buildings)
}

func removeDestroyedRawBuildings(buildings map[string]bool, ids []int) []string {
	removed := make([]string, 0, len(ids))
	for _, destroyedID := range ids {
		for name, active := range buildings {
			if id, ok := gamedata.OriginalBuildingIDForName(name); active && ok && id == destroyedID {
				delete(buildings, name)
				removed = append(removed, name)
				break
			}
		}
	}
	return removed
}

func resolveEarthquakeDamage(colony *engine.ColonyState, buildings map[string]bool, marines, tanks int,
	player engine.PlayerState, highG bool, rng *randStream) gamedata.StrategicColonyDamageResult {
	ids := earthquakeRawBuildingIDs(buildings)
	damage := originalEarthquakeDamage(colony.Population, len(ids), rng.Intn(3)+1, rng.Intn(2)+1)
	result := gamedata.ResolveStrategicColonyDamage(gamedata.StrategicColonyDamageState{
		Population: colony.Population, LastPopulationPoints: colony.BombardmentLastPopulationPoints,
		Marines: marines, Tanks: tanks, BuildProgress: colony.BombardmentBuildProgress,
		RawBuildingIDs: ids,
		MarineHitCost:  gamedata.GroundMarineHitsToKill(highG, hasPoweredArmorFor(player)),
		TankHitCost:    tankHitsToKillFor(player, highG), BuildingHitCost: 1,
	}, damage, rng.Intn)
	colony.Population = result.State.Population
	colony.BombardmentLastPopulationPoints = result.State.LastPopulationPoints
	colony.BombardmentBuildProgress = result.State.BuildProgress
	normalizeColonyJobsAfterPopulationLoss(colony)
	return result
}

func (s *GameSession) applyPlayerEarthquake() (earthquakeImpact, bool) {
	i, ok := pickEarthquakeColony(s.PlayerColonies, s.eventRand.Intn)
	if !ok {
		return earthquakeImpact{}, false
	}
	name := s.colonyLabel(i)
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
	result := resolveEarthquakeDamage(&s.PlayerColonies[i], buildings, marines, tanks,
		s.Player, s.raceHasTrait(gamedata.TRAIT_HIGH_G), s.eventRand)
	for _, removed := range removeDestroyedRawBuildings(buildings, result.DestroyedBuildingIDs) {
		s.removeBuildingEffect(i, removed)
	}
	if i < len(s.PlayerColonyMarines) {
		s.PlayerColonyMarines[i] = result.State.Marines
	}
	if i < len(s.PlayerColonyTanks) {
		s.PlayerColonyTanks[i] = result.State.Tanks
	}
	impact := earthquakeImpact{ColonyIndex: i, ColonyName: name, PopulationLost: result.PopulationLost,
		BuildingsDestroyed: len(result.DestroyedBuildingIDs), MarinesLost: result.MarinesLost,
		TanksLost: result.TanksLost, BuildProgressLost: result.BuildProgressLost,
		ColonyDestroyed: result.ColonyDestroyed}
	if result.ColonyDestroyed {
		s.removePlayerColony(i)
		if star >= 0 && star < len(s.Stars) && !s.playerHasColonyAtStar(star) {
			s.Stars[star].Owner = 0
		}
	} else {
		s.recalcColonyMorale(i)
	}
	return impact, true
}

func (s *GameSession) applyAIEarthquake(aiIndex int) (earthquakeImpact, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return earthquakeImpact{}, false
	}
	a := &s.AIPlayers[aiIndex]
	i, ok := pickEarthquakeColony(a.Colonies, s.eventRand.Intn)
	if !ok {
		return earthquakeImpact{}, false
	}
	ensureAIGroundForceSlots(a)
	name := stripAILabel(a.Name)
	star := -1
	if i < len(a.ColonyStars) {
		star = a.ColonyStars[i]
	}
	var buildings map[string]bool
	if i < len(a.ColonyBuildings) {
		buildings = a.ColonyBuildings[i]
	}
	result := resolveEarthquakeDamage(&a.Colonies[i], buildings, a.ColonyMarines[i], a.ColonyTanks[i],
		a.Player, aiRaceHasTrait(*a, gamedata.TRAIT_HIGH_G), s.eventRand)
	for _, removed := range removeDestroyedRawBuildings(buildings, result.DestroyedBuildingIDs) {
		removeBuildingEffectFromColony(&a.Colonies[i], removed)
	}
	a.ColonyMarines[i], a.ColonyTanks[i] = result.State.Marines, result.State.Tanks
	impact := earthquakeImpact{ColonyIndex: i, ColonyName: name, PopulationLost: result.PopulationLost,
		BuildingsDestroyed: len(result.DestroyedBuildingIDs), MarinesLost: result.MarinesLost,
		TanksLost: result.TanksLost, BuildProgressLost: result.BuildProgressLost,
		ColonyDestroyed: result.ColonyDestroyed}
	if result.ColonyDestroyed {
		s.removeAIColonyAfterEvent(aiIndex, i)
		if star >= 0 && star < len(s.Stars) && !s.anyAIHasColonyAtStar(star) {
			s.Stars[star].Owner = 0
		}
	}
	return impact, true
}

func (s *GameSession) playerHasColonyAtStar(star int) bool {
	for i := range s.PlayerColonies {
		if s.PlayerColonyStarIndex(i) == star {
			return true
		}
	}
	return false
}

func (s *GameSession) anyAIHasColonyAtStar(star int) bool {
	for i := range s.AIPlayers {
		for _, candidate := range s.AIPlayers[i].ColonyStars {
			if candidate == star {
				return true
			}
		}
	}
	return false
}

func (s *GameSession) removeAIColonyAfterEvent(aiIndex, i int) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return
	}
	a := &s.AIPlayers[aiIndex]
	if i < 0 || i >= len(a.Colonies) {
		return
	}
	a.Colonies = append(a.Colonies[:i], a.Colonies[i+1:]...)
	if i < len(a.ColonyStars) {
		a.ColonyStars = append(a.ColonyStars[:i], a.ColonyStars[i+1:]...)
	}
	if i < len(a.ColonyPlanets) {
		a.ColonyPlanets = append(a.ColonyPlanets[:i], a.ColonyPlanets[i+1:]...)
	}
	if i < len(a.ColonyBuildings) {
		a.ColonyBuildings = append(a.ColonyBuildings[:i], a.ColonyBuildings[i+1:]...)
	}
	if i < len(a.ColonyLeaderNames) {
		a.ColonyLeaderNames = append(a.ColonyLeaderNames[:i], a.ColonyLeaderNames[i+1:]...)
	}
	removeAIGroundForceSlot(a, i)
	a.OwnedStars = len(a.ColonyStars)
}
