package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// CapitolBuildName 是原版 tech.json 的穩定查詢鍵；顯示層必須經外部譯表轉成目前語言。
const CapitolBuildName = "Capitol"

// CapitolProductionCost 對應原版 raw building 9 的 200 生產點。
const CapitolProductionCost = 200

func isUnifiedGovernment(gov gamedata.MoraleGovernmentType) bool {
	return gov == gamedata.MoraleGovUnification || gov == gamedata.MoraleGovGalacticUnification
}

func colonyHasCapitol(buildings []map[string]bool, colony int) bool {
	return colony >= 0 && colony < len(buildings) && buildings[colony] != nil &&
		buildings[colony][CapitolBuildName]
}

func colonyIndexForPlanet(planets []int, planet int) int {
	for i, candidate := range planets {
		if candidate == planet {
			return i
		}
	}
	return -1
}

// replacementCapitolPlanet 對映 sub_ECB65：排除失去的殖民地，選人口最高者；
// 同人口時保留較低索引。沒有候選回 -1。
func replacementCapitolPlanet(colonies []engine.ColonyState, planets []int, lost int) int {
	best, bestPop := -1, -1
	for i := range colonies {
		if i == lost || i >= len(planets) || planets[i] < 0 {
			continue
		}
		if colonies[i].Population > bestPop {
			best, bestPop = planets[i], colonies[i].Population
		}
	}
	return best
}

func (s *GameSession) ensureCapitolState() {
	if !s.PlayerCapitolPlanetKnown {
		s.PlayerCapitolPlanet = -1
		if len(s.PlayerColonyPlanets) > 0 {
			s.PlayerCapitolPlanet = s.PlayerColonyPlanets[0]
		}
		s.PlayerCapitolPlanetKnown = true
		if len(s.PlayerColonies) > 0 {
			for len(s.ColonyBuildings) < len(s.PlayerColonies) {
				s.ColonyBuildings = append(s.ColonyBuildings, nil)
			}
			if s.ColonyBuildings[0] == nil {
				s.ColonyBuildings[0] = make(map[string]bool)
			}
			if !isUnifiedGovernment(s.effectiveGovernment()) {
				s.ColonyBuildings[0][CapitolBuildName] = true
			}
		}
	}
	for i := range s.AIPlayers {
		a := &s.AIPlayers[i]
		if a.CapitolPlanetKnown {
			continue
		}
		a.CapitolPlanet = -1
		if len(a.ColonyPlanets) > 0 {
			a.CapitolPlanet = a.ColonyPlanets[0]
		}
		a.CapitolPlanetKnown = true
		if len(a.Colonies) > 0 {
			for len(a.ColonyBuildings) < len(a.Colonies) {
				a.ColonyBuildings = append(a.ColonyBuildings, nil)
			}
			if a.ColonyBuildings[0] == nil {
				a.ColonyBuildings[0] = make(map[string]bool)
			}
			if !isUnifiedGovernment(effectiveAIGovernment(a)) {
				a.ColonyBuildings[0][CapitolBuildName] = true
			}
		}
	}
}

func (s *GameSession) playerCapitolMissing() bool {
	if !s.PlayerCapitolPlanetKnown {
		return false
	}
	if s.PlayerCapitolPlanet < 0 || isUnifiedGovernment(s.effectiveGovernment()) {
		return false
	}
	return s.PlayerCapitolRebuildRequired
}

func (s *GameSession) playerCanBuildCapitol(colony int) bool {
	if !s.PlayerCapitolPlanetKnown {
		return false
	}
	return colony >= 0 && colony < len(s.PlayerColonies) && colony < len(s.PlayerColonyPlanets) &&
		s.PlayerColonyPlanets[colony] == s.PlayerCapitolPlanet && s.playerCapitolMissing() &&
		!colonyHasCapitol(s.ColonyBuildings, colony)
}

func (s *GameSession) aiCanBuildCapitol(aiIndex, colony int) bool {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return false
	}
	a := &s.AIPlayers[aiIndex]
	if !a.CapitolPlanetKnown {
		return false
	}
	return !isUnifiedGovernment(effectiveAIGovernment(a)) && colony >= 0 && colony < len(a.Colonies) &&
		colony < len(a.ColonyPlanets) && a.ColonyPlanets[colony] == a.CapitolPlanet &&
		!colonyHasCapitol(a.ColonyBuildings, colony)
}

func (s *GameSession) syncStartingCapitolForGovernment() {
	s.ensureCapitolState()
	idx := colonyIndexForPlanet(s.PlayerColonyPlanets, s.PlayerCapitolPlanet)
	if idx < 0 || idx >= len(s.ColonyBuildings) {
		return
	}
	if s.ColonyBuildings[idx] == nil {
		s.ColonyBuildings[idx] = make(map[string]bool)
	}
	if isUnifiedGovernment(s.effectiveGovernment()) {
		delete(s.ColonyBuildings[idx], CapitolBuildName)
		s.PlayerCapitolRebuildRequired = false
	} else if s.Turn <= 1 {
		s.ColonyBuildings[idx][CapitolBuildName] = true
		s.PlayerCapitolRebuildRequired = false
	}
}

func (s *GameSession) recalcAIColonyMorale(aiIndex int) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return
	}
	a := &s.AIPlayers[aiIndex]
	gov := effectiveAIGovernment(a)
	missing := !isUnifiedGovernment(gov) && a.CapitolPlanet >= 0 && a.CapitolRebuildRequired
	for i := range a.Colonies {
		buildings := map[string]bool(nil)
		if i < len(a.ColonyBuildings) {
			buildings = a.ColonyBuildings[i]
		}
		a.Colonies[i].MoralePercent = colonyMoralePercent(gov, buildings,
			a.Colonies[i].UnassimilatedPop > 0, achievementMoralePercent(a.Player, gov))
		if missing {
			a.Colonies[i].MoralePercent += gamedata.MoraleCapitalCapturedPenalty(gov)
		}
		setColonyGovernmentOutput(&a.Colonies[i], gov)
	}
}

// prepareCapturedAIColony 對映 sub_ECBF7／sub_ECB65，須在 AI 平行陣列移除前呼叫。
// 回傳可轉給新擁有者的建築集合；Capitol 一律不隨殖民地過戶。
func (s *GameSession) prepareCapturedAIColony(aiIndex, colony, planet int) map[string]bool {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return nil
	}
	s.ensureCapitolState()
	a := &s.AIPlayers[aiIndex]
	var transferred map[string]bool
	if colony >= 0 && colony < len(a.ColonyBuildings) && a.ColonyBuildings[colony] != nil {
		transferred = cloneBuildings(a.ColonyBuildings[colony])
		delete(transferred, CapitolBuildName)
	}
	if a.CapitolPlanet == planet {
		a.CapitolPlanet = replacementCapitolPlanet(a.Colonies, a.ColonyPlanets, colony)
		a.CapitolPlanetKnown = true
		a.CapitolRebuildRequired = a.CapitolPlanet >= 0
	}
	if s.PlayerCapitolPlanet < 0 {
		s.PlayerCapitolPlanet = planet
		s.PlayerCapitolPlanetKnown = true
		s.PlayerCapitolRebuildRequired = true
	}
	return transferred
}
