package shell

import (
	"sort"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

type originalDiplomaticTechnologyCandidate struct {
	tech  gamedata.Technology
	value int
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
