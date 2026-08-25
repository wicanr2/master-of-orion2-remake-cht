package shell

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

type diplomaticIncidentPartner struct {
	kind   eventEmpireKind
	index  int
	name   string
	policy gamedata.ForeignPolicy
}

func originalRelationFromNormalized(value int) int { return clampRelation(value) * 5 / 2 }

func normalizedRelationFromOriginal(value int) int { return clampRelation(value * 2 / 5) }

func (s *GameSession) aiEventGovernment(index int) int {
	if index < 0 || index >= len(s.AIPlayers) {
		return int(gamedata.MoraleGovDictatorship)
	}
	ai := s.AIPlayers[index]
	race := aiRaceIndex(ai)
	if race < 0 || race >= len(Races) {
		return int(gamedata.MoraleGovDictatorship)
	}
	return gamedata.OrigRaceTrait(Races[race].OrigIdx, gamedata.TRAIT_GOVERNMENT)
}

func (s *GameSession) pickDiplomaticIncidentPartner(target eventEmpireTarget) (diplomaticIncidentPartner, bool) {
	if s.eventRand == nil {
		s.eventRand = newRandStream(s.EventSeed*2654435761 + 1)
	}
	seen := 0
	chosen := diplomaticIncidentPartner{}
	consider := func(candidate diplomaticIncidentPartner, alive bool) {
		if !alive {
			return
		}
		seen++
		if s.eventRand.Intn(seen)+1 == 1 {
			chosen = candidate
		}
	}

	switch target.kind {
	case eventEmpireAI:
		if target.index < 0 || target.index >= len(s.AIPlayers) {
			return diplomaticIncidentPartner{}, false
		}
		// 原版槽序以真人槽在前；remake 沒有熱座真人彼此外交矩陣，故只納入目前可表示的真人邊。
		consider(diplomaticIncidentPartner{kind: eventEmpirePlayer, index: 0,
			name:   s.eventEmpireTargetName(eventEmpireTarget{kind: eventEmpirePlayer}),
			policy: s.AIPlayers[target.index].Treaty.FormalPolicy}, colonyPopulationTotal(s.PlayerColonies) > 0)
		s.ensureAIRelations()
		for i := range s.AIPlayers {
			if i == target.index {
				continue
			}
			policy := gamedata.DIPLO_NONE
			if target.index < len(s.AIPolicies) && i < len(s.AIPolicies[target.index]) {
				policy = s.AIPolicies[target.index][i]
			}
			consider(diplomaticIncidentPartner{kind: eventEmpireAI, index: i,
				name: stripAILabel(s.AIPlayers[i].Name), policy: policy},
				colonyPopulationTotal(s.AIPlayers[i].Colonies) > 0)
		}
	default:
		for i := range s.AIPlayers {
			consider(diplomaticIncidentPartner{kind: eventEmpireAI, index: i,
				name: stripAILabel(s.AIPlayers[i].Name), policy: s.AIPlayers[i].Treaty.FormalPolicy},
				colonyPopulationTotal(s.AIPlayers[i].Colonies) > 0)
		}
	}
	return chosen, seen > 0
}

func (s *GameSession) applyDiplomaticIncident(evID int, target eventEmpireTarget) (eventResult, bool) {
	partner, ok := s.pickDiplomaticIncidentPartner(target)
	if !ok {
		return eventResult{}, false
	}
	actorName := s.eventEmpireTargetName(target)
	actorGovernment := int(s.Government)
	targetCharismatic := false
	current := 0

	switch {
	case target.kind == eventEmpireAI && partner.kind == eventEmpireAI:
		s.ensureAIRelations()
		current = s.AIRelations[target.index][partner.index]
		actorGovernment = s.aiEventGovernment(target.index)
		targetCharismatic = aiRaceHasTrait(s.AIPlayers[partner.index], gamedata.TRAIT_CHARISMATIC)
	case target.kind == eventEmpireAI:
		current = s.AIPlayers[target.index].Relation
		actorGovernment = s.aiEventGovernment(target.index)
		targetCharismatic = s.RaceCharismatic()
	default:
		current = s.AIPlayers[partner.index].Relation
		targetCharismatic = aiRaceHasTrait(s.AIPlayers[partner.index], gamedata.TRAIT_CHARISMATIC)
	}

	raw, ok := gamedata.OriginalDiplomaticIncidentRelation(originalRelationFromNormalized(current),
		evID, actorGovernment, targetCharismatic, partner.policy)
	if !ok {
		return eventResult{}, false
	}
	updated := normalizedRelationFromOriginal(raw)
	if target.kind == eventEmpireAI && partner.kind == eventEmpireAI {
		s.AIRelations[target.index][partner.index] = updated
		s.AIRelations[partner.index][target.index] = updated
	} else if target.kind == eventEmpireAI {
		s.AIPlayers[target.index].Relation = updated
	} else {
		s.AIPlayers[partner.index].Relation = updated
	}

	if evID == 4 {
		return eventResult{
			Message:   fmt.Sprintf("%s 與 %s 之間爆發外交暗殺風波，兩國關係惡化", actorName, partner.name),
			MessageEN: fmt.Sprintf("An assassination plot damaged diplomatic relations between %s and %s.", actorName, partner.name),
		}, true
	}
	return eventResult{
		Message:   fmt.Sprintf("%s 與 %s 締結外交聯姻，兩國關係大幅改善", actorName, partner.name),
		MessageEN: fmt.Sprintf("A diplomatic marriage greatly improved relations between %s and %s.", actorName, partner.name),
	}, true
}
