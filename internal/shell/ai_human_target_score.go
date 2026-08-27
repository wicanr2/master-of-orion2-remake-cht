package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

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
	// government raw 0 需要 sub_DCB47 的殖民地 player-mask／可達計數；未 typed 前失敗即關閉。
	if government == 0 {
		return gamedata.OriginalHumanTargetScoreResult{}, 0, false
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
		GovernmentZeroReachabilityKnown: false,
		IncidentMemory:                  a.OriginalHumanIncidentMemoryRaw,
		IncidentReason:                  a.OriginalHumanIncidentReasonRaw,
		TreatyGrievance:                 a.OriginalHumanTreatyGrievanceRaw,
		TreatyVictimRaw:                 a.OriginalHumanTreatyVictimRaw, TreatyVictimKnown: a.OriginalHumanTreatyVictimKnown,
		SourceRaw: a.PopulationRaceSlot, PopulationDominance: dominance, PopulationTrend: trend,
		TargetRaw1DFIs3: int(s.Government) == 3, TargetCharismatic: s.RaceCharismatic(),
		TargetBetrayedHonorable: a.OriginalHumanBetrayalRaw,
	}
	result, valid := gamedata.OriginalHumanTargetScore(input, roll)
	return result, powerRatio, valid
}
