package shell

import (
	"strings"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

var originalShieldApplications = []gamedata.Technology{
	gamedata.TECH_NONE,
	gamedata.TECH_CLASS_I_SHIELD,
	gamedata.TECH_CLASS_III_SHIELD,
	gamedata.TECH_CLASS_V_SHIELD,
	gamedata.TECH_CLASS_VII_SHIELD,
	gamedata.TECH_CLASS_X_SHIELD,
}

func bestKnownOriginalBeam(ps engine.PlayerState) gamedata.OrigWeapon {
	best := gamedata.OrigWeaponTable[3]
	for id := 0; id < 40 && id < len(gamedata.OrigWeaponTable); id++ {
		weapon := gamedata.OrigWeaponTable[id]
		if weapon.Cat != gamedata.WeaponCatBeam || weapon.Tech == gamedata.TECH_NONE {
			continue
		}
		topic, ok := gamedata.OrigTechTopic(weapon.Tech)
		if !ok || !playerStateKnowsTech(ps, topic, weapon.Tech) {
			continue
		}
		if weapon.DamageMax > best.DamageMax {
			best = weapon
		}
	}
	return best
}

func bestKnownOriginalShieldIndex(ps engine.PlayerState) int {
	best := 0
	for i := 1; i < len(originalShieldApplications); i++ {
		tech := originalShieldApplications[i]
		topic, ok := gamedata.OrigTechTopic(tech)
		if ok && playerStateKnowsTech(ps, topic, tech) {
			best = i
		}
	}
	return best
}

// originalAncientTechApplications 對應 sub_58853：先找未知的銀河光束基準鄰近科技，
// 再附加銀河最高護盾。輸入順序就是原版 player slot 順序。
func originalAncientTechApplications(target engine.PlayerState, empires []engine.PlayerState) []gamedata.Technology {
	benchmark := gamedata.OrigWeaponTable[3]
	shieldIndex := 0
	for _, empire := range empires {
		beam := bestKnownOriginalBeam(empire)
		if beam.DamageMax > benchmark.DamageMax {
			benchmark = beam
		}
		if idx := bestKnownOriginalShieldIndex(empire); idx > shieldIndex {
			shieldIndex = idx
		}
	}

	result := make([]gamedata.Technology, 0, 2)
	bestDifference := 200
	selected := gamedata.TECH_NONE
	for id := 1; id < 40 && id < len(gamedata.OrigWeaponTable); id++ {
		weapon := gamedata.OrigWeaponTable[id]
		if weapon.Cat != gamedata.WeaponCatBeam || weapon.Tech == gamedata.TECH_NONE {
			continue
		}
		topic, ok := gamedata.OrigTechTopic(weapon.Tech)
		if !ok || topic == gamedata.TOPIC_XENON_TECHNOLOGY {
			continue
		}
		difference := weapon.DamageMax - benchmark.DamageMax
		if difference < 0 || difference > 50 || difference > bestDifference ||
			playerStateKnowsTech(target, topic, weapon.Tech) {
			continue
		}
		selected = weapon.Tech
		bestDifference = difference
		if id == benchmark.ID {
			bestDifference = 1
		}
	}
	if selected != gamedata.TECH_NONE {
		result = append(result, selected)
	}

	if shieldIndex > 0 {
		shield := originalShieldApplications[shieldIndex]
		topic, ok := gamedata.OrigTechTopic(shield)
		if ok && !playerStateKnowsTech(target, topic, shield) && shield != selected {
			result = append(result, shield)
		}
	}
	return result
}

func (s *GameSession) ancientTechEmpireStates() []engine.PlayerState {
	states := make([]engine.PlayerState, 0, len(s.Seats)+len(s.AIPlayers)+1)
	if s.HotseatEnabled() {
		for i := range s.Seats {
			if colonyPopulationTotal(s.Seats[i].PlayerColonies) > 0 {
				states = append(states, s.Seats[i].Player)
			}
		}
	} else if colonyPopulationTotal(s.PlayerColonies) > 0 {
		states = append(states, s.Player)
	}
	for i := range s.AIPlayers {
		if colonyPopulationTotal(s.AIPlayers[i].Colonies) > 0 {
			states = append(states, s.AIPlayers[i].Player)
		}
	}
	return states
}

func grantAncientTechApplications(ps *engine.PlayerState, applications []gamedata.Technology) []string {
	if ps == nil {
		return nil
	}
	names := make([]string, 0, len(applications))
	for _, tech := range applications {
		topic, ok := gamedata.OrigTechTopic(tech)
		if !ok || playerStateKnowsTech(*ps, topic, tech) {
			continue
		}
		grantTechnologyApplication(ps, topic, tech)
		names = append(names, gamedata.TechnologyName(tech))
	}
	return names
}

func ancientTechNames(names []string) string { return strings.Join(names, "、") }
