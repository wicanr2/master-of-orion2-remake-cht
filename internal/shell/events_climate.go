package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func reservoirEventColony(colonies []engine.ColonyState, rng *randStream) int {
	selected, count := -1, 0
	for i := range colonies {
		count++
		if rng.Intn(count)+1 == 1 {
			selected = i
		}
	}
	return selected
}

// originalClimateEventColony 對應 sub_2310C：每次先由 sub_23D44 對全部殖民地做
// reservoir sampling，再拒絕 Terran／Gaia；最多 200 次，最後以低索引 fallback。
func originalClimateEventColony(colonies []engine.ColonyState, rng *randStream) (int, bool) {
	if len(colonies) == 0 || rng == nil {
		return 0, false
	}
	for attempt := 0; attempt < 200; attempt++ {
		i := reservoirEventColony(colonies, rng)
		if i >= 0 && colonies[i].Climate < gamedata.TERRAN {
			return i, true
		}
	}
	for i := range colonies {
		if colonies[i].Climate < gamedata.TERRAN {
			return i, true
		}
	}
	return 0, false
}

func syncPlanetClimate(p *Planet, climate gamedata.PlanetClimate) {
	if p == nil {
		return
	}
	p.ClimateID = climate
	p.Climate = climateDisplayName(climate)
}

func (s *GameSession) applyPlayerClimateEvent() (idx int, before gamedata.PlanetClimate, ok bool) {
	i, ok := originalClimateEventColony(s.PlayerColonies, s.eventRand)
	if !ok {
		return 0, 0, false
	}
	before = s.PlayerColonies[i].Climate
	s.applyClimateChange(i, gamedata.TERRAN)
	syncPlanetClimate(s.ColonyPlanet(i), gamedata.TERRAN)
	return i, before, true
}

func (s *GameSession) aiColonyPlanet(aiIndex, colonyIndex int) *Planet {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return nil
	}
	a := &s.AIPlayers[aiIndex]
	if colonyIndex >= 0 && colonyIndex < len(a.ColonyPlanets) {
		if p := a.ColonyPlanets[colonyIndex]; p >= 0 && p < len(s.Planets) {
			return &s.Planets[p]
		}
	}
	if colonyIndex >= 0 && colonyIndex < len(a.ColonyStars) {
		if p := s.PlanetAt(a.ColonyStars[colonyIndex]); p >= 0 && p < len(s.Planets) {
			return &s.Planets[p]
		}
	}
	return nil
}

func (s *GameSession) applyAIClimateEvent(aiIndex int) (idx int, before gamedata.PlanetClimate, ok bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return 0, 0, false
	}
	a := &s.AIPlayers[aiIndex]
	i, ok := originalClimateEventColony(a.Colonies, s.eventRand)
	if !ok {
		return 0, 0, false
	}
	before = a.Colonies[i].Climate
	applyClimateChangeToColony(&a.Colonies[i], gamedata.TERRAN,
		aiRaceHasTrait(*a, gamedata.TRAIT_AQUATIC), aiRaceHasTrait(*a, gamedata.TRAIT_TOLERANT))
	syncPlanetClimate(s.aiColonyPlanet(aiIndex, i), gamedata.TERRAN)
	return i, before, true
}
