package shell

import (
	"sort"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

type originalDiplomaticTechnologyCandidate struct {
	tech  gamedata.Technology
	value int
}

type originalDiplomaticColonyCandidate struct {
	star       int
	population int
}

type originalAIHumanDecision struct {
	Outcome   int
	Intensity int
	Reason    int
	Action    gamedata.OriginalHumanDiplomaticAction
}

// originalAIMaintenanceTotal 對映 sub_E2000 寫入、sub_4F93B 讀取的 player+0xB4。
// 納貢成本依本回合毛收入計算；目前沒有持久化其原始 +0xC0 分項，所以存在 AI→玩家
// 納貢時失敗即關閉。
func (s *GameSession) originalAIMaintenanceTotal(aiIndex int) (int, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return 0, false
	}
	a := &s.AIPlayers[aiIndex]
	if a.Treaty.AITribute != TributeNone {
		return 0, false
	}
	ps := a.Player
	commandCostPerPoint := ps.CommandOverflowCostPerPoint
	if commandCostPerPoint <= 0 {
		commandCostPerPoint = gamedata.IncomeCommandOverflowCostPerPoint
	}
	uncovered := ps.UsedCommandPoints - ps.CommandPointsSupply
	if uncovered < 0 {
		uncovered = 0
	}
	total := ps.Maintenance + gamedata.IncomeFreighterMaintenanceCost(ps.ActiveFreighters) +
		uncovered*commandCostPerPoint + ps.SpyMaintenance + ps.OfficerMaintenance
	if total < 0 {
		return 0, false
	}
	return total, true
}

// originalAIHumanTechnologyCandidates 對映 sub_27094：列出真人已知、來源 AI 未知且
// Calc_Tech_Value_ 對來源為正的 application，依 value、raw technology ID 穩定升冪排序。
func (s *GameSession) originalAIHumanTechnologyCandidates(aiIndex int) ([]int, int, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return nil, 0, false
	}
	a := &s.AIPlayers[aiIndex]
	if !a.OriginalTechProfileKnown {
		return nil, 0, false
	}
	sourceKnown := knownTechnologyApplications(a.Player)
	targetKnown := knownTechnologyApplications(s.Player)
	ratio, ok := gamedata.OriginalDiplomaticTechnologyRatioLimit(sourceKnown, targetKnown)
	if !ok {
		return nil, 0, false
	}
	opponents := make([]map[gamedata.Technology]bool, 0, len(s.AIPlayers))
	opponents = append(opponents, targetKnown)
	for i := range s.AIPlayers {
		if i != aiIndex {
			opponents = append(opponents, knownTechnologyApplications(s.AIPlayers[i].Player))
		}
	}
	state := gamedata.OriginalStartingValueState{
		Difficulty: s.Difficulty, RelativeTurn: s.Turn,
		AIProfile: a.OriginalTechProfile, AIProfileKnown: true,
		Raw4: a.OriginalTechProfile.Raw4, Raw4Known: true,
		Known: sourceKnown, Opponents: opponents,
	}
	candidates := make([]originalDiplomaticTechnologyCandidate, 0)
	for tech, targetHas := range targetKnown {
		if !targetHas || sourceKnown[tech] {
			continue
		}
		value := gamedata.OriginalAITechValueKnownSlice(tech, state)
		if value > 0 {
			candidates = append(candidates, originalDiplomaticTechnologyCandidate{tech: tech, value: value})
		}
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].value != candidates[j].value {
			return candidates[i].value < candidates[j].value
		}
		return candidates[i].tech < candidates[j].tech
	})
	out := make([]int, len(candidates))
	for i := range candidates {
		out[i] = int(candidates[i].tech)
	}
	return out, ratio, true
}

// originalAIHumanColonyCandidates 對映 sub_E5CD4／sub_E5BE3／sub_E4A09 的 remake
// 可表示路徑。候選是玩家至少兩座殖民星中的非首都星，必須可由來源 AI／聯盟人口殖民地
// 以 source fuel range 抵達；同星人口加總後依 uint8 分數升冪，初始同分順序沿原版倒掃星號。
// star+0x67 的暫時外交佔用槽在 remake 沒有獨立狀態；正常狀態等價為全 -1。
func (s *GameSession) originalAIHumanColonyCandidates(aiIndex int) ([]int, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) ||
		len(s.PlayerColonies) != len(s.PlayerColonyStars) ||
		len(s.PlayerColonies) != len(s.PlayerColonyPlanets) ||
		len(s.PlayerColonies) != len(s.ColonyBuildings) || !s.PlayerCapitolPlanetKnown {
		return nil, false
	}
	_, bases, rangeParsecs, ok := s.originalAIPopulationReachabilityContext(aiIndex)
	if !ok {
		return nil, false
	}
	populationByStar := map[int]int{}
	capitolStars := map[int]bool{}
	for i, colony := range s.PlayerColonies {
		star := s.PlayerColonyStars[i]
		if star < 0 || star >= len(s.Stars) || colony.Population < 0 || colony.Population > 255 {
			return nil, false
		}
		populationByStar[star] = (populationByStar[star] + colony.Population) & 0xff
		if s.PlayerColonyPlanets[i] == s.PlayerCapitolPlanet || builtMapHasOriginalBuildingID(s.ColonyBuildings[i], 9) {
			capitolStars[star] = true
		}
	}
	if len(populationByStar) <= 1 {
		return []int{}, true
	}
	candidates := make([]originalDiplomaticColonyCandidate, 0, len(populationByStar)-1)
	for star := len(s.Stars) - 1; star >= 0; star-- {
		population, owned := populationByStar[star]
		if !owned || capitolStars[star] {
			continue
		}
		if s.Stars[star].Wormhole >= 0 {
			return nil, false
		}
		if !anyOriginalAIColonyInRange(s, star, bases, float64(rangeParsecs)) {
			continue
		}
		candidates = append(candidates, originalDiplomaticColonyCandidate{star: star, population: population})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].population < candidates[j].population
	})
	out := make([]int, len(candidates))
	for i := range candidates {
		out[i] = candidates[i].star
	}
	return out, true
}

// originalAIHumanDiplomaticAction 對映 sub_544A1 對 sub_4F93B 的四個 enable=1 呼叫。
// 任一 producer 未知時整體失敗即關閉，避免部分候選改變 kind RNG 路徑。
func (s *GameSession) originalAIHumanDiplomaticAction(aiIndex, intensity int,
	roll func(int) int) (gamedata.OriginalHumanDiplomaticAction, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) || intensity < 0 || roll == nil {
		return gamedata.OriginalHumanDiplomaticAction{}, false
	}
	technologies, technologyLimit, ok := s.originalAIHumanTechnologyCandidates(aiIndex)
	if !ok {
		return gamedata.OriginalHumanDiplomaticAction{}, false
	}
	colonies, ok := s.originalAIHumanColonyCandidates(aiIndex)
	if !ok {
		return gamedata.OriginalHumanDiplomaticAction{}, false
	}
	maintenance, ok := s.originalAIMaintenanceTotal(aiIndex)
	if !ok || s.Player.BC < 0 {
		return gamedata.OriginalHumanDiplomaticAction{}, false
	}
	creditLimit := 0
	denominator := maintenance
	if denominator < 10 {
		denominator = 10
	}
	if s.Player.BC != 0 {
		creditLimit = s.Player.BC / denominator / 10
	}
	if creditLimit > 10 {
		creditLimit = 10
	}
	return gamedata.OriginalHumanDiplomaticActionSelect(gamedata.OriginalHumanDiplomaticActionInput{
		Intensity:     intensity,
		DirectEnabled: true, TechnologyEnabled: true, CreditsEnabled: true, ColonyEnabled: true,
		TechnologyCandidates: technologies, TechnologyRatioLimit: technologyLimit,
		SourceMaintenance: maintenance, TargetCredits: s.Player.BC, CreditIntensityLimit: creditLimit,
		ColonyCandidates: colonies,
	}, roll)
}

// queueOriginalAIHumanDiplomaticRequest 對映 sub_53EDB outcome 1／3／4 寫
// +0x657 reason、sub_54CC0 鏡射 payload，以及 Humans_Requesting_Diplomacy_ 的玩家可見閘門。
// 這裡只保存請求；資產、關係與軍事 callback 必須等玩家在正常外交 UI 明確選擇後才套用。
func (s *GameSession) queueOriginalAIHumanDiplomaticRequest(aiIndex, outcome int,
	action gamedata.OriginalHumanDiplomaticAction) bool {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return false
	}
	request, ok := gamedata.OriginalHumanDiplomaticRequestForOutcome(outcome, action)
	if !ok {
		return false
	}
	a := &s.AIPlayers[aiIndex]
	a.OriginalHumanDiplomaticRequest = &request
	a.WantsAudience = true
	a.AudienceReason = AudienceReasonOriginal
	return true
}

func (s *GameSession) originalAISourcePopulationStrongest(aiIndex int) (bool, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) {
		return false, false
	}
	source := colonyPopulationTotal(s.AIPlayers[aiIndex].Colonies)
	if source < 0 || source > 32767 {
		return false, false
	}
	if p := colonyPopulationTotal(s.PlayerColonies); p > source {
		return false, true
	}
	for i := range s.AIPlayers {
		if i != aiIndex && colonyPopulationTotal(s.AIPlayers[i].Colonies) > source {
			return false, true
		}
	}
	return true, true
}

// originalAIHumanDecision 串起 sub_544A1 → sub_4F93B → outcome 尾端，保留同一條
// roll stream。舊 JSON／GAM 若缺 word_19A0E2 typed 狀態，且其值可能影響 type 4，會在
// 擲骰前失敗即關閉；新局與新版 snapshot 直接消費議會生命週期保存的 raw 三態。
func (s *GameSession) originalAIHumanDecision(aiIndex int, roll func(int) int) (originalAIHumanDecision, bool) {
	if aiIndex < 0 || aiIndex >= len(s.AIPlayers) || roll == nil {
		return originalAIHumanDecision{}, false
	}
	strongest, ok := s.originalAISourcePopulationStrongest(aiIndex)
	sourcePower, targetPower, exact := s.originalAIHumanDirectionalFleetPower(aiIndex)
	powerRatio, ratioOK := gamedata.OriginalNPCPowerRatio(sourcePower, targetPower,
		s.originalAIThirdPartyWars(aiIndex))
	if !ok || !exact || !ratioOK || strongest && powerRatio/40+1 > 3 && !s.OriginalCouncilDiplomacyStateKnown {
		return originalAIHumanDecision{}, false
	}
	// sub_4F93B 的 typed 候選／維護資料先做無 RNG preflight；若等 score 擲完才發現
	// producer unknown，再落回 stance fallback 會讓 fallback 平白消耗原版 RNG。
	if _, _, techOK := s.originalAIHumanTechnologyCandidates(aiIndex); !techOK {
		return originalAIHumanDecision{}, false
	}
	if _, colonyOK := s.originalAIHumanColonyCandidates(aiIndex); !colonyOK {
		return originalAIHumanDecision{}, false
	}
	if _, maintenanceOK := s.originalAIMaintenanceTotal(aiIndex); !maintenanceOK || s.Player.BC < 0 {
		return originalAIHumanDecision{}, false
	}
	result, scoreRatio, ok := s.originalAIHumanTargetScore(aiIndex, roll)
	if !ok || scoreRatio != powerRatio {
		return originalAIHumanDecision{}, false
	}
	intensity, ok := gamedata.OriginalHumanTargetActionIntensity(powerRatio, roll(3))
	if !ok {
		return originalAIHumanDecision{}, false
	}
	action, ok := s.originalAIHumanDiplomaticAction(aiIndex, intensity, roll)
	if !ok {
		return originalAIHumanDecision{}, false
	}
	a := s.AIPlayers[aiIndex]
	outcome, ok := gamedata.OriginalHumanTargetOutcomeAfterAction(gamedata.OriginalHumanTargetOutcomeInput{
		Score: result.Score, ContactTurns: a.OriginalHumanContactTurns, Difficulty: s.Difficulty,
		DiplomaticActionAvailable: action.Kind != gamedata.OriginalHumanDiplomaticActionNone,
		ForcedType2:               result.ForcedType2, SourcePopulationStrongest: strongest,
		CouncilStateIs1: s.OriginalCouncilDiplomacyStateKnown && s.OriginalCouncilDiplomacyState == 1,
		SourceRepulsive: aiRaceHasTrait(a, gamedata.TRAIT_REPULSIVE),
		TargetRepulsive: s.RaceRepulsive(),
	}, intensity, roll)
	if !ok {
		return originalAIHumanDecision{}, false
	}
	return originalAIHumanDecision{Outcome: outcome, Intensity: intensity, Reason: result.ReasonCode, Action: action}, true
}
