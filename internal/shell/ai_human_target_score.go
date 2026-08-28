package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

type originalPopulationColony struct {
	colony engine.ColonyState
	star   int
}

func colonyHasOriginalRaceSlot(colony engine.ColonyState, slot int) (bool, bool) {
	if slot < 0 || !engine.PopulationGroupsComplete(colony) {
		return false, false
	}
	for _, group := range colony.PopulationGroups {
		if !group.RaceSlotKnown {
			return false, false
		}
		if group.RaceSlot == slot && populationGroupUnits(group) > 0 {
			return true, true
		}
	}
	return false, true
}

func (s *GameSession) originalAIPopulationReachabilityContext(aiIndex int) ([]originalPopulationColony, []int, int, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return nil, nil, 0, false
	}
	source := &s.AIPlayers[aiIndex]
	if !source.PopulationRaceSlotKnown {
		return nil, nil, 0, false
	}
	rangeParsecs, ok := originalAIFuelRangeParsecs(source.Player)
	if !ok {
		return nil, nil, 0, false
	}
	colonies := make([]originalPopulationColony, 0, len(s.PlayerColonies)+len(source.Colonies))
	if len(s.PlayerColonies) != len(s.PlayerColonyStars) {
		return nil, nil, 0, false
	}
	for i := range s.PlayerColonies {
		colonies = append(colonies, originalPopulationColony{s.PlayerColonies[i], s.PlayerColonyStars[i]})
	}
	for i := range s.AIPlayers {
		owner := &s.AIPlayers[i]
		if len(owner.Colonies) != len(owner.ColonyStars) {
			return nil, nil, 0, false
		}
		for j := range owner.Colonies {
			colonies = append(colonies, originalPopulationColony{owner.Colonies[j], owner.ColonyStars[j]})
		}
	}

	alliedSlots := map[int]bool{source.PopulationRaceSlot: true}
	if source.Treaty.FormalPolicy == gamedata.DIPLO_ALLIANCE {
		alliedSlots[0] = true
	}
	for i := range s.AIPlayers {
		if i == aiIndex || !s.AIPlayers[i].PopulationRaceSlotKnown || aiIndex >= len(s.AIPolicies) ||
			i >= len(s.AIPolicies[aiIndex]) {
			continue
		}
		if s.AIPolicies[aiIndex][i] == gamedata.DIPLO_ALLIANCE {
			alliedSlots[s.AIPlayers[i].PopulationRaceSlot] = true
		}
	}

	bases := make([]int, 0, len(colonies))
	for _, item := range colonies {
		for slot := range alliedSlots {
			has, known := colonyHasOriginalRaceSlot(item.colony, slot)
			if !known {
				return nil, nil, 0, false
			}
			if has {
				bases = append(bases, item.star)
				break
			}
		}
	}
	return colonies, bases, rangeParsecs, true
}

// originalAIHumanGovernmentZeroReachability 對映 sub_DCB47 → sub_FF666 →
// sub_FF5F8/sub_FF593/sub_FF4E9。每個含 target 人口的殖民地若也含 source 人口加 5；
// 否則只要本星在航程內，或蟲洞另一端已被 source 造訪且另一端在航程內，便加 1。
func (s *GameSession) originalAIHumanGovernmentZeroReachability(aiIndex int) (int, bool) {
	colonies, bases, rangeParsecs, ok := s.originalAIPopulationReachabilityContext(aiIndex)
	if !ok {
		return 0, false
	}
	source := &s.AIPlayers[aiIndex]

	score := 0
	for _, item := range colonies {
		hasTarget, known := colonyHasOriginalRaceSlot(item.colony, 0)
		if !known {
			return 0, false
		}
		if !hasTarget {
			continue
		}
		hasSource, known := colonyHasOriginalRaceSlot(item.colony, source.PopulationRaceSlot)
		if !known {
			return 0, false
		}
		if hasSource {
			score += 5
			continue
		}
		if item.star < 0 || item.star >= len(s.Stars) {
			return 0, false
		}
		if anyOriginalAIColonyInRange(s, item.star, bases, float64(rangeParsecs)) {
			score++
			continue
		}
		partner := s.Stars[item.star].Wormhole
		if partner < 0 {
			continue
		}
		if partner >= len(s.Stars) || !source.ExploredStarsKnown || len(source.ExploredStars) != len(s.Stars) {
			return 0, false
		}
		if source.ExploredStars[partner] && anyOriginalAIColonyInRange(s, partner, bases, float64(rangeParsecs)) {
			score++
		}
	}
	return score, true
}

// originalAIHumanPopulationTrend 從 remake 的時間排序 350 格環取得原版目前／40 格前人口值。
func (s *GameSession) originalAIHumanPopulationTrend(aiIndex int) (int, bool) {
	if s.Turn < 100 {
		return 0, true
	}
	if aiIndex < 0 || len(s.History) < 41 {
		return 0, false
	}
	now, old := s.History[len(s.History)-1], s.History[len(s.History)-41]
	empire := aiIndex + 1
	if empire >= len(now.Empires) || empire >= len(old.Empires) || len(now.Empires) == 0 || len(old.Empires) == 0 {
		return 0, false
	}
	return gamedata.OriginalHumanTargetPopulationTrend(s.Turn,
		now.Empires[empire].Population, old.Empires[empire].Population,
		now.Empires[0].Population, old.Empires[0].Population)
}

func (s *GameSession) originalAIHumanPopulationDominance(aiIndex int) (int, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return 0, false
	}
	populations := make([]int, 0, len(s.AIPlayers)+1)
	source := -1
	if player := colonyPopulationTotal(s.PlayerColonies); player > 0 {
		populations = append(populations, player)
	}
	for index := range s.AIPlayers {
		population := colonyPopulationTotal(s.AIPlayers[index].Colonies)
		if population <= 0 {
			continue
		}
		if index == aiIndex {
			source = len(populations)
		}
		populations = append(populations, population)
	}
	return gamedata.OriginalHumanTargetPopulationDominance(populations, source)
}

func (s *GameSession) originalAIThirdPartyWars(aiIndex int) int {
	wars := 0
	for target := range s.AIPlayers {
		if target != aiIndex && aiIndex < len(s.AIPolicies) && target < len(s.AIPolicies[aiIndex]) &&
			s.AIPolicies[aiIndex][target] >= gamedata.DIPLO_LIMITED_WAR {
			wars++
		}
	}
	return wars
}

// originalAIHumanTargetScore 組合 sub_544A1 已閉合輸入。ok=false 表示至少一個原版欄位
// 尚無 typed producer；呼叫端不得用部分分數冒充完整決策。
func (s *GameSession) originalAIHumanTargetScore(aiIndex int,
	roll func(int) int) (gamedata.OriginalHumanTargetScoreResult, int, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) || roll == nil {
		return gamedata.OriginalHumanTargetScoreResult{}, 0, false
	}
	a := &s.AIPlayers[aiIndex]
	if !a.OriginalRaw28Known || !a.OriginalHumanIncidentKnown || !a.PopulationRaceSlotKnown {
		return gamedata.OriginalHumanTargetScoreResult{}, 0, false
	}
	government := originalAIRelationGovernment(*a)
	reachability, reachabilityKnown := 0, false
	if government == 0 {
		reachability, reachabilityKnown = s.originalAIHumanGovernmentZeroReachability(aiIndex)
		if !reachabilityKnown {
			return gamedata.OriginalHumanTargetScoreResult{}, 0, false
		}
	}
	sourcePower, targetPower, exact := s.originalAIHumanDirectionalFleetPower(aiIndex)
	if !exact {
		return gamedata.OriginalHumanTargetScoreResult{}, 0, false
	}
	powerRatio, valid := gamedata.OriginalNPCPowerRatio(sourcePower, targetPower,
		s.originalAIThirdPartyWars(aiIndex))
	if !valid {
		return gamedata.OriginalHumanTargetScoreResult{}, 0, false
	}
	dominance, valid := s.originalAIHumanPopulationDominance(aiIndex)
	if !valid {
		return gamedata.OriginalHumanTargetScoreResult{}, 0, false
	}
	trend, valid := s.originalAIHumanPopulationTrend(aiIndex)
	if !valid {
		return gamedata.OriginalHumanTargetScoreResult{}, 0, false
	}
	targetCapacity := 0
	for _, colony := range s.PlayerColonies {
		targetCapacity += colony.PopMax
	}
	_, targetValue := s.aiRaidTargetWithValue(aiIndex)
	input := gamedata.OriginalHumanTargetScoreInput{
		RelationRaw: a.originalRelationRaw(), Personality: int(a.Personality),
		DiplomatBonus: leaderEmpireSkillBonus(s.Leaders, gamedata.SKILL_DIPLOMAT),
		Difficulty:    s.Difficulty, FormalPolicy: int(a.Treaty.FormalPolicy), Government: government,
		Raw28: a.OriginalRaw28, SourcePopulation: colonyPopulationTotal(a.Colonies),
		TargetPopulationCapacity: targetCapacity, ForceWarRaw: a.OriginalWarFlag60ERaw,
		FoodDeficitTurns: a.OriginalFoodDeficitTurns, PowerRatio: powerRatio,
		GovernmentOneTargetValue: targetValue, GovernmentOneTargetExists: targetValue > 0,
		GovernmentZeroReachability: reachability, GovernmentZeroReachabilityKnown: reachabilityKnown,
		IncidentMemory:  a.OriginalHumanIncidentMemoryRaw,
		IncidentReason:  a.OriginalHumanIncidentReasonRaw,
		TreatyGrievance: a.OriginalHumanTreatyGrievanceRaw,
		TreatyVictimRaw: a.OriginalHumanTreatyVictimRaw, TreatyVictimKnown: a.OriginalHumanTreatyVictimKnown,
		SourceRaw: a.PopulationRaceSlot, PopulationDominance: dominance, PopulationTrend: trend,
		TargetRaw1DFIs3: int(s.Government) == 3, TargetCharismatic: s.RaceCharismatic(),
		TargetBetrayedHonorable: a.OriginalHumanBetrayalRaw,
	}
	result, valid := gamedata.OriginalHumanTargetScore(input, roll)
	return result, powerRatio, valid
}
