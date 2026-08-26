package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
)

func (s *GameSession) plagueTargetEligible(colonies []engine.ColonyState, planetAt func(int) int) (int, bool) {
	if len(colonies) == 0 || planetAt == nil || s.eventRand == nil {
		return 0, false
	}
	// sub_23DA0 做一次 reservoir sampling；它另要求 colony+0x13F==0（無 Capitol），該
	// 欄位的玩家語意仍未知，remake 只保留已能表示的殖民地事件互斥條件。
	i := reservoirEventColony(colonies, s.eventRand)
	if i < 0 {
		return 0, false
	}
	planet := planetAt(i)
	if planet < 0 || s.planetColonyEventActive(planet) {
		return 0, false
	}
	return i, true
}

func originalPlagueResearchNeed(research, difficulty, roll1To8 int) (int, bool) {
	if research < 1 || difficulty < 0 || difficulty > 4 || roll1To8 < 1 || roll1To8 > 8 {
		return 0, false
	}
	return research * (roll1To8 + 2*difficulty), true
}

func (s *GameSession) startPlayerPlague() (idx, need int, ok bool) {
	i, ok := s.plagueTargetEligible(s.PlayerColonies, s.ColonyPlanetIndex)
	if !ok || i >= len(s.LastPlayerOutput.Colonies) {
		return 0, 0, false
	}
	need, ok = originalPlagueResearchNeed(s.LastPlayerOutput.Colonies[i].Research, s.Difficulty, s.eventRoll(8))
	if !ok {
		return 0, 0, false
	}
	s.PersistentEvents = append(s.PersistentEvents, PersistentEvent{
		Kind: PersistentPlague, PlanetIndex: s.ColonyPlanetIndex(i), Turns: -1, ResearchNeeded: need,
	})
	return i, need, true
}

func (s *GameSession) startAIPlague(aiIndex int) (idx, need int, ok bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return 0, 0, false
	}
	a := &s.AIPlayers[aiIndex]
	planetAt := func(i int) int {
		if i >= 0 && i < len(a.ColonyPlanets) {
			return a.ColonyPlanets[i]
		}
		return -1
	}
	i, ok := s.plagueTargetEligible(a.Colonies, planetAt)
	if !ok {
		return 0, 0, false
	}
	research := engine.RunColonyTurn(a.Colonies[i]).Research
	need, ok = originalPlagueResearchNeed(research, s.Difficulty, s.eventRoll(8))
	if !ok {
		return 0, 0, false
	}
	s.PersistentEvents = append(s.PersistentEvents, PersistentEvent{
		Kind: PersistentPlague, PlanetIndex: planetAt(i), Turns: -1, ResearchNeeded: need,
	})
	return i, need, true
}

func (s *GameSession) recordPlagueResearch(planetIndex, research int) {
	if research < 0 {
		return
	}
	for i := range s.PersistentEvents {
		e := &s.PersistentEvents[i]
		if e.Kind == PersistentPlague && e.PlanetIndex == planetIndex {
			e.ResearchDone += research
		}
	}
}

func (s *GameSession) recordPlayerPlagueResearch(out engine.EmpireOutput) {
	for i := range s.PlayerColonies {
		if i < len(out.Colonies) {
			s.recordPlagueResearch(s.ColonyPlanetIndex(i), out.Colonies[i].Research)
		}
	}
}

func (s *GameSession) recordAIPlagueResearch(aiIndex int, out engine.EmpireOutput) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return
	}
	a := &s.AIPlayers[aiIndex]
	for i := range a.Colonies {
		if i < len(out.Colonies) && i < len(a.ColonyPlanets) {
			s.recordPlagueResearch(a.ColonyPlanets[i], out.Colonies[i].Research)
		}
	}
}
