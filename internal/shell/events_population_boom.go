package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"

func (s *GameSession) populationBoomTargetEligible(colonies []engine.ColonyState, planetAt func(int) int) (int, bool) {
	if len(colonies) == 0 || planetAt == nil || s.eventRand == nil {
		return 0, false
	}
	// sub_23D44 先從全部正常殖民地做一次 reservoir sampling；事件 17 不會在
	// 抽到滿人口或已有殖民地事件時重抽，該候選直接失敗。
	i := reservoirEventColony(colonies, s.eventRand)
	if i < 0 || colonies[i].Population >= colonies[i].PopMax {
		return 0, false
	}
	planet := planetAt(i)
	if planet < 0 || s.planetColonyEventActive(planet) {
		return 0, false
	}
	return i, true
}

func (s *GameSession) startPlayerPopulationBoom() (int, bool) {
	i, ok := s.populationBoomTargetEligible(s.PlayerColonies, s.ColonyPlanetIndex)
	if !ok {
		return 0, false
	}
	s.PersistentEvents = append(s.PersistentEvents, PersistentEvent{
		Kind: PersistentPopulationBoom, PlanetIndex: s.ColonyPlanetIndex(i), Turns: -1,
	})
	return i, true
}

func (s *GameSession) startAIPopulationBoom(aiIndex int) (int, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return 0, false
	}
	a := &s.AIPlayers[aiIndex]
	planetAt := func(i int) int {
		if i >= 0 && i < len(a.ColonyPlanets) {
			return a.ColonyPlanets[i]
		}
		return -1
	}
	i, ok := s.populationBoomTargetEligible(a.Colonies, planetAt)
	if !ok {
		return 0, false
	}
	s.PersistentEvents = append(s.PersistentEvents, PersistentEvent{
		Kind: PersistentPopulationBoom, PlanetIndex: planetAt(i), Turns: -1,
	})
	return i, true
}

func (s *GameSession) aiColoniesForTurn(aiIndex int, colonies []engine.ColonyState) []engine.ColonyState {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return colonies
	}
	a := &s.AIPlayers[aiIndex]
	needsCopy := false
	for i := range colonies {
		star := -1
		if i < len(a.ColonyStars) {
			star = a.ColonyStars[i]
		}
		if s.StarInStasis(star) || s.StarUnderSupernova(star) || (i < len(a.ColonyPlanets) && (s.planetPopulationBoomActive(a.ColonyPlanets[i]) ||
			s.planetPlagueActive(a.ColonyPlanets[i]))) {
			needsCopy = true
			break
		}
	}
	if !needsCopy {
		return colonies
	}
	out := append([]engine.ColonyState(nil), colonies...)
	for i := range out {
		if i < len(a.ColonyStars) && s.StarInStasis(a.ColonyStars[i]) {
			freezeColonyForStasis(&out[i])
		}
		if i < len(a.ColonyStars) && s.StarUnderSupernova(a.ColonyStars[i]) {
			out[i].ResearchDiverted = true
		}
		if i < len(a.ColonyPlanets) && s.planetPopulationBoomActive(a.ColonyPlanets[i]) {
			out[i].GrowthBonusSum += 100
		}
		if i < len(a.ColonyPlanets) && s.planetPlagueActive(a.ColonyPlanets[i]) {
			out[i].GrowthBonusSum -= 200
		}
	}
	return out
}
