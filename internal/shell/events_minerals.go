package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// originalMineralEventColony 對應 sub_2325E／sub_232BB：最多重試 200 次。
// 枯竭經 sub_23DA0 只在無 Capitol 候選間抽樣；發現經 sub_23D44 在全部殖民地間抽樣。
func originalMineralEventColony(colonies []engine.ColonyState, buildings []map[string]bool,
	planetAt func(int) *Planet, eventID int, rng *randStream) (int, bool) {
	if len(colonies) == 0 || planetAt == nil || rng == nil || (eventID != 11 && eventID != 12) {
		return 0, false
	}
	for attempt := 0; attempt < 200; attempt++ {
		i := -1
		if eventID == 11 {
			var ok bool
			i, ok = pickEarthquakeColony(colonies,
				func(i int) bool { return !colonyHasCapitol(buildings, i) }, rng.Intn)
			if !ok {
				continue
			}
		} else {
			i = reservoirEventColony(colonies, rng)
		}
		if i < 0 {
			continue
		}
		p := planetAt(i)
		if p == nil {
			continue
		}
		if eventID == 11 && p.MineralID == gamedata.ULTRA_RICH {
			return i, true
		}
		if eventID == 12 && p.MineralID < gamedata.ULTRA_RICH {
			return i, true
		}
	}
	return 0, false
}

func applyMineralEventToColony(c *engine.ColonyState, p *Planet, eventID int) (from, to gamedata.PlanetMinerals, ok bool) {
	if c == nil || p == nil {
		return 0, 0, false
	}
	from = p.MineralID
	switch eventID {
	case 11:
		if from != gamedata.ULTRA_RICH {
			return 0, 0, false
		}
		to = from - 1
	case 12:
		if from >= gamedata.ULTRA_RICH {
			return 0, 0, false
		}
		to = from + 2
		if to > gamedata.ULTRA_RICH {
			to = gamedata.ULTRA_RICH
		}
	default:
		return 0, 0, false
	}
	p.MineralID = to
	p.Mineral = mineralDisplayName(to)
	// 原版事件消費端只寫 planet+0x0A（礦產），不重算 planet+9（重力）。
	c.MineralRichness = to
	c.IndustryPerWorker = gamedata.MineralIndustryPerWorker(to)
	return from, to, true
}

func (s *GameSession) applyPlayerMineralEvent(eventID int) (idx int, from, to gamedata.PlanetMinerals, ok bool) {
	i, ok := originalMineralEventColony(s.PlayerColonies, s.ColonyBuildings, s.ColonyPlanet, eventID, s.eventRand)
	if !ok {
		return 0, 0, 0, false
	}
	from, to, ok = applyMineralEventToColony(&s.PlayerColonies[i], s.ColonyPlanet(i), eventID)
	return i, from, to, ok
}

func (s *GameSession) applyAIMineralEvent(aiIndex, eventID int) (idx int, from, to gamedata.PlanetMinerals, ok bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return 0, 0, 0, false
	}
	a := &s.AIPlayers[aiIndex]
	planetAt := func(i int) *Planet { return s.aiColonyPlanet(aiIndex, i) }
	i, ok := originalMineralEventColony(a.Colonies, a.ColonyBuildings, planetAt, eventID, s.eventRand)
	if !ok {
		return 0, 0, 0, false
	}
	from, to, ok = applyMineralEventToColony(&a.Colonies[i], planetAt(i), eventID)
	return i, from, to, ok
}
