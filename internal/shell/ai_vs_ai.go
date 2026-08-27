package shell

// ai_vs_ai.go：可選的 AI 對 AI 外交與抽象戰爭。
//
// 原版的 AI 會在同一個銀河裡彼此評估，但 remake 先前只把關係矩陣餵給
// 議會搖擺票，沒有真正的 AI 對 AI 條約或戰鬥。這裡補一個可保存、可重播的
// remake 模型：關係分數形成正式政策，貿易／研究協定每回合交換少量資源，
// 戰爭則讓唯一的抽象 AI 艦隊飛往敵方殖民地並結算一場確定性戰鬥。
//
// AI 現已有逐艦藍圖與實艦；本檔的 AI 對 AI 戰爭仍採 FleetStrength 比例結算，再把損失
// 回寫為實艦移除。原版精確艦隊戰術與損失選艦仍屬未知，不冒稱逐艦戰鬥 oracle。

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

const (
	aiAIColonyDefensePerPop = 4
	aiAIBattleRandomSpan    = 21
	aiAIAttackerLossDivisor = 6
	aiAIDefenderLossDivisor = 5
)

// AIAIBattleReport 是上一場 AI 對 AI 抽象戰鬥的可見結果。
type AIAIBattleReport struct {
	AttackerIndex int
	DefenderIndex int
	AttackerName  string
	DefenderName  string
	StarIdx       int
	StarName      string
	AttackerWon   bool
	AttackerRoll  int
	DefenderRoll  int
	AttackerLoss  int
	DefenderLoss  int
	Message       string
	MessageEN     string
}

// ensureAIAIState 補齊四個 AI 對 AI 狀態矩陣。舊存檔沒有這些欄位時只在
// EnableAIVsAI 被打開後建立，避免把舊對局悄悄升級成另一套規則。
func (s *GameSession) ensureAIAIState() {
	n := len(s.AIPlayers)
	s.AIWars = resizeBoolMatrix(s.AIWars, n)
	s.AIPolicies = resizePolicyMatrix(s.AIPolicies, n)
	s.AITrade = resizeBoolMatrix(s.AITrade, n)
	s.AIResearch = resizeBoolMatrix(s.AIResearch, n)
	s.AIReputationRaw = resizeIntMatrix(s.AIReputationRaw, n)
	s.AITreatyBiasRaw = resizeIntMatrix(s.AITreatyBiasRaw, n)
	s.AIAgreementBiasRaw = resizeIntMatrix(s.AIAgreementBiasRaw, n)
	s.AITributeModes = resizeIntMatrix(s.AITributeModes, n)
	s.AIIncidentReasonRaw = resizeIntMatrix(s.AIIncidentReasonRaw, n)
	s.AIIncidentMagnitudeRaw = resizeIntMatrix(s.AIIncidentMagnitudeRaw, n)
	s.AIIncidentMemoryRaw = resizeIntMatrix(s.AIIncidentMemoryRaw, n)
	s.AIIncidentBetrayalRaw = resizeBoolMatrix(s.AIIncidentBetrayalRaw, n)
	s.AIWarDurationRaw = resizeIntMatrix(s.AIWarDurationRaw, n)
	s.AIDiplomacyCooldownRaw = resizeIntMatrix(s.AIDiplomacyCooldownRaw, n)
	for i := range s.AIPolicies {
		if i < len(s.AIPolicies[i]) {
			s.AIPolicies[i][i] = gamedata.DIPLO_NONE
		}
	}
	for i := range s.AIPlayers {
		if !s.AIPlayers[i].FleetTargetAISet {
			s.AIPlayers[i].FleetTargetAI = -1
		}
	}
}

func resizeBoolMatrix(old [][]bool, n int) [][]bool {
	out := make([][]bool, n)
	for i := range out {
		out[i] = make([]bool, n)
		if i < len(old) {
			copy(out[i], old[i])
		}
	}
	return out
}

func resizePolicyMatrix(old [][]gamedata.ForeignPolicy, n int) [][]gamedata.ForeignPolicy {
	out := make([][]gamedata.ForeignPolicy, n)
	for i := range out {
		out[i] = make([]gamedata.ForeignPolicy, n)
		if i < len(old) {
			copy(out[i], old[i])
		}
	}
	return out
}

func filterAIBoolMatrix(matrix [][]bool, kept []int) [][]bool {
	if len(kept) == 0 || len(matrix) == 0 {
		return nil
	}
	out := make([][]bool, len(kept))
	for i, oldI := range kept {
		out[i] = make([]bool, len(kept))
		for j, oldJ := range kept {
			if oldI >= 0 && oldI < len(matrix) && oldJ >= 0 && oldJ < len(matrix[oldI]) {
				out[i][j] = matrix[oldI][oldJ]
			}
		}
	}
	return out
}

func filterAIPolicyMatrix(matrix [][]gamedata.ForeignPolicy, kept []int) [][]gamedata.ForeignPolicy {
	if len(kept) == 0 || len(matrix) == 0 {
		return nil
	}
	out := make([][]gamedata.ForeignPolicy, len(kept))
	for i, oldI := range kept {
		out[i] = make([]gamedata.ForeignPolicy, len(kept))
		for j, oldJ := range kept {
			if oldI >= 0 && oldI < len(matrix) && oldJ >= 0 && oldJ < len(matrix[oldI]) {
				out[i][j] = matrix[oldI][oldJ]
			}
		}
	}
	return out
}

// advanceAIAIDiplomacy 先依 sub_2552D 推進 ordered AI pair 的條約、協議與
// 納貢談判，再讓既有可選資源／戰爭 consumer 消費正式狀態。
func (s *GameSession) advanceAIAIDiplomacy() {
	s.ensureOriginalAIAIRelations()
	s.ensureAIAIState()
	s.advanceOriginalAIAIWarTimers()
	s.advanceOriginalAIIncidentMemory(func() int { return s.diplomacyGrowthRandForTurn().Intn(100) + 1 })
	for i := range s.AIPlayers {
		for j := range s.AIPlayers {
			if i == j {
				continue
			}
			s.AITreatyBiasRaw[i][j] = gamedata.OriginalNPCNegotiationBiasRecovery(s.AITreatyBiasRaw[i][j])
			s.AIAgreementBiasRaw[i][j] = gamedata.OriginalNPCNegotiationBiasRecovery(s.AIAgreementBiasRaw[i][j])
		}
	}
	roll := func(n int) int { return s.diplomacyGrowthRandForTurn().Intn(n) + 1 }
	for i := range s.AIPlayers {
		for j := range s.AIPlayers {
			if i == j {
				continue
			}
			s.advanceOriginalAIAINegotiation(i, j, roll)
		}
	}
	s.advanceOriginalAIAIWarPolicy(roll)
	for i := 0; i < len(s.AIPlayers); i++ {
		for j := i + 1; j < len(s.AIPlayers); j++ {
			atWar := s.AIWars[i][j] || s.AIWars[j][i] ||
				s.AIPolicies[i][j] >= gamedata.DIPLO_LIMITED_WAR ||
				s.AIPolicies[j][i] >= gamedata.DIPLO_LIMITED_WAR
			s.AIWars[i][j], s.AIWars[j][i] = atWar, atWar
			if atWar {
				// sub_51078 對 AI↔AI 固定寫 policy 4；5／6 是人類參戰分支。
				s.AIPolicies[i][j], s.AIPolicies[j][i] = gamedata.DIPLO_LIMITED_WAR, gamedata.DIPLO_LIMITED_WAR
				s.AITrade[i][j], s.AITrade[j][i] = false, false
				s.AIResearch[i][j], s.AIResearch[j][i] = false, false
				continue
			}
			trade := s.AITrade[i][j] || s.AITrade[j][i]
			research := s.AIResearch[i][j] || s.AIResearch[j][i]
			s.AITrade[i][j], s.AITrade[j][i] = trade, trade
			s.AIResearch[i][j], s.AIResearch[j][i] = research, research
			if trade || research {
				s.exchangeAIBenefits(i, j, trade, research)
			}
		}
	}
}

// advanceOriginalAIAIWarTimers 對應 sub_5090C：停戰冷卻每回合遞減；正式
// 戰爭的 +0x717 雙向累加，最高 250。policy 3 在冷卻歸零後回到無條約。
func (s *GameSession) advanceOriginalAIAIWarTimers() {
	for i := 0; i < len(s.AIPlayers); i++ {
		for j := i + 1; j < len(s.AIPlayers); j++ {
			for _, pair := range [][2]int{{i, j}, {j, i}} {
				a, b := pair[0], pair[1]
				if s.AIDiplomacyCooldownRaw[a][b] > 0 {
					s.AIDiplomacyCooldownRaw[a][b]--
				}
			}
			if s.AIDiplomacyCooldownRaw[i][j] == 0 && s.AIPolicies[i][j] == gamedata.DIPLO_PEACE {
				s.AIPolicies[i][j], s.AIPolicies[j][i] = gamedata.DIPLO_NONE, gamedata.DIPLO_NONE
			}
			if s.AIPolicies[i][j] >= gamedata.DIPLO_LIMITED_WAR {
				if s.AIWarDurationRaw[i][j] < 250 {
					s.AIWarDurationRaw[i][j]++
				}
				if s.AIWarDurationRaw[j][i] < 250 {
					s.AIWarDurationRaw[j][i]++
				}
			}
		}
	}
}

// advanceOriginalAIAIWarPolicy 接回 sub_25DF1 的 reason 20／68／113 與一般
// reason 23 候選，再依 sub_2670A 處理 AI↔AI 直接停戰。各理由採原版分開掃描
// 全體目標的順序，避免改成逐目標混排後改變亂數序列。
func (s *GameSession) advanceOriginalAIAIWarPolicy(roll func(int) int) {
	count := len(s.AIPlayers)
	if count < 2 {
		return
	}
	power, _ := s.originalAIPowerMatrix()
	for source := range s.AIPlayers {
		wars := 0
		for target := range s.AIPlayers {
			if source != target && s.AIPolicies[source][target] >= gamedata.DIPLO_LIMITED_WAR {
				wars++
			}
		}
		if wars != 0 {
			continue
		}
		candidate := make([]bool, count)
		targetAtWar := make([]bool, count)
		for target := range s.AIPlayers {
			if source == target {
				continue
			}
			for third := range s.AIPlayers {
				if third != source && third != target && s.AIPolicies[target][third] >= gamedata.DIPLO_LIMITED_WAR {
					targetAtWar[target] = true
					break
				}
			}
		}
		government := originalAIRelationGovernment(s.AIPlayers[source])
		// reason 20：raw government 3 的低機率宣戰。
		for target := range s.AIPlayers {
			if source == target {
				continue
			}
			ratio, valid := gamedata.OriginalNPCPowerRatio(power[source][target], power[target][source], wars)
			if !valid {
				continue
			}
			candidate[target], _ = gamedata.OriginalNPCGovernmentWarCandidate(gamedata.OriginalNPCSpecialWarCandidateInput{
				Difficulty: s.Difficulty, Government: government, PowerRatio: ratio,
				Cooldown: s.AIDiplomacyCooldownRaw[source][target],
			}, roll)
		}
		// reason 68：只有 Turn%playerCount 的輪值目標擲敵意門檻。
		for target := range s.AIPlayers {
			if source == target || candidate[target] {
				continue
			}
			ratio, valid := gamedata.OriginalNPCPowerRatio(power[source][target], power[target][source], wars)
			if !valid {
				continue
			}
			candidate[target], _ = gamedata.OriginalNPCHostilityWarCandidate(gamedata.OriginalNPCSpecialWarCandidateInput{
				Difficulty: s.Difficulty, Government: government, PowerRatio: ratio,
				Cooldown: s.AIDiplomacyCooldownRaw[source][target], TargetIsRotating: s.Turn%count == target,
				CurrentRelationRaw: s.originalAIAIRelation(source, target),
			}, roll)
		}
		// reason 113：原版每個 source 無條件先擲一次 Random(100)，再把同一結果
		// 套到所有尚無候選的合格目標；FoodDeficitTurns 為 +0x7EC producer。
		foodRoll := roll
		foodResult := 0
		foodRoll = func(n int) int {
			if foodResult == 0 {
				foodResult = roll(n)
			}
			return foodResult
		}
		for target := range s.AIPlayers {
			if source == target || candidate[target] {
				continue
			}
			ratio, valid := gamedata.OriginalNPCPowerRatio(power[source][target], power[target][source], wars)
			if !valid {
				continue
			}
			candidate[target], _ = gamedata.OriginalNPCFoodDeficitWarCandidate(gamedata.OriginalNPCSpecialWarCandidateInput{
				Difficulty: s.Difficulty, Government: government, PowerRatio: ratio,
				Cooldown:         s.AIDiplomacyCooldownRaw[source][target],
				FoodDeficitTurns: s.AIPlayers[source].OriginalFoodDeficitTurns,
			}, foodRoll)
		}
		if foodResult == 0 {
			// 即使沒有合格目標，原版仍在目標 loop 前消耗這次擲骰。
			foodResult = roll(100)
		}
		// reason 23：一般政策／政府／國力候選。
		for target := range s.AIPlayers {
			if source == target || candidate[target] {
				continue
			}
			got, ok := gamedata.OriginalNPCGenericWarCandidate(gamedata.OriginalNPCWarCandidateInput{
				Difficulty: s.Difficulty, Government: government,
				Policy: s.AIPolicies[source][target], TradeActive: s.AITrade[source][target],
				ResearchActive: s.AIResearch[source][target], TributeMode: s.AITributeModes[source][target],
				SourceStrength: power[source][target], TargetStrength: power[target][source],
				SourceThirdPartyWars: wars, Cooldown: s.AIDiplomacyCooldownRaw[source][target],
				TargetIsRotating: s.Turn%count == target, TargetAtWarWithAI: targetAtWar[target],
			}, roll)
			if ok && got {
				candidate[target] = true
			}
		}
		candidates := make([]int, 0, count-1)
		for target := range candidate {
			if candidate[target] && !(s.Difficulty >= 3 && targetAtWar[target]) {
				candidates = append(candidates, target)
			}
		}
		if len(candidates) > 0 {
			choice := roll(len(candidates))
			if choice >= 1 && choice <= len(candidates) {
				s.declareOriginalAIAIWar(source, candidates[choice-1], roll)
			}
		}
	}
	threshold, ok := gamedata.OriginalNPCCeasefireThreshold(s.Difficulty, 0)
	if !ok {
		return
	}
	for i := 0; i < count; i++ {
		for j := i + 1; j < count; j++ {
			if s.AIPolicies[i][j] >= gamedata.DIPLO_LIMITED_WAR && s.AIWarDurationRaw[i][j] > threshold {
				s.makeOriginalAIAICeasefire(i, j)
			}
		}
	}
}

func (s *GameSession) declareOriginalAIAIWar(source, target int, roll func(int) int) {
	value := roll(25)
	if value < 1 || value > 25 {
		return
	}
	s.markOriginalAITreatyBetrayal(source, target)
	s.AIPolicies[source][target], s.AIPolicies[target][source] = gamedata.DIPLO_LIMITED_WAR, gamedata.DIPLO_LIMITED_WAR
	s.AIWars[source][target], s.AIWars[target][source] = true, true
	s.AITrade[source][target], s.AITrade[target][source] = false, false
	s.AIResearch[source][target], s.AIResearch[target][source] = false, false
	s.AITributeModes[source][target], s.AITributeModes[target][source] = 0, 0
	s.AITreatyBiasRaw[source][target], s.AITreatyBiasRaw[target][source] = -200, -200
	s.AIAgreementBiasRaw[source][target], s.AIAgreementBiasRaw[target][source] = -200, -200
	s.AIWarDurationRaw[source][target], s.AIWarDurationRaw[target][source] = 0, 0
	s.AIDiplomacyCooldownRaw[source][target], s.AIDiplomacyCooldownRaw[target][source] = 0, 0
	s.clearOriginalAIIncidentMemory(source, target)
	s.setOriginalAIAIRelation(source, target, -74-value)
}

func (s *GameSession) makeOriginalAIAICeasefire(a, b int) {
	s.AIPolicies[a][b], s.AIPolicies[b][a] = gamedata.DIPLO_PEACE, gamedata.DIPLO_PEACE
	s.AIWars[a][b], s.AIWars[b][a] = false, false
	s.AITrade[a][b], s.AITrade[b][a] = false, false
	s.AIResearch[a][b], s.AIResearch[b][a] = false, false
	s.AITributeModes[a][b], s.AITributeModes[b][a] = 0, 0
	s.AIDiplomacyCooldownRaw[a][b], s.AIDiplomacyCooldownRaw[b][a] = 30, 30
	s.clearOriginalAIIncidentMemory(a, b)
	next := s.originalAIAIRelation(a, b) + 50
	if next > 0 {
		next = 0
	}
	s.setOriginalAIAIRelation(a, b, next)
}

func (s *GameSession) advanceOriginalAIAINegotiation(outer, inner int, roll func(int) int) {
	power, _ := s.originalAIPowerMatrix()
	thirdPartyBonus, nonHumanWars, outerWars := 0, 0, 0
	for k := range s.AIPlayers {
		if k == outer || k == inner {
			continue
		}
		if s.AIPolicies[outer][k] >= gamedata.DIPLO_LIMITED_WAR {
			thirdPartyBonus += 10
			nonHumanWars++
			outerWars++
		}
		if s.AIPolicies[inner][k] >= gamedata.DIPLO_LIMITED_WAR {
			thirdPartyBonus += 20
			nonHumanWars++
		}
		if inner < len(s.AIIncidentMemoryRaw) && k < len(s.AIIncidentMemoryRaw[inner]) &&
			s.AIIncidentMemoryRaw[inner][k] > 0 {
			thirdPartyBonus += 5
		}
	}
	result, ok := gamedata.OriginalNPCTreatyNegotiation(gamedata.OriginalNPCTreatyInput{
		Difficulty: s.Difficulty, CurrentRaw: s.originalAIAIRelation(outer, inner),
		ReputationRaw: s.AIReputationRaw[outer][inner],
		TreatyBiasRaw: s.AITreatyBiasRaw[outer][inner], AgreementBiasRaw: s.AIAgreementBiasRaw[outer][inner],
		Policy: s.AIPolicies[outer][inner], TradeActive: s.AITrade[outer][inner],
		ResearchActive:   s.AIResearch[outer][inner],
		OuterGovernment:  originalAIRelationGovernment(s.AIPlayers[outer]),
		InnerGovernment:  originalAIRelationGovernment(s.AIPlayers[inner]),
		NonHumanWarCount: nonHumanWars, ThirdPartyBonus: thirdPartyBonus,
		TributeBlocked: s.AITributeModes[outer][inner] < 0,
		OuterStrength:  power[outer][inner], InnerStrength: power[inner][outer],
		OuterThirdPartyWars: outerWars,
	}, roll)
	if !ok {
		return
	}
	s.AITreatyBiasRaw[outer][inner] = result.TreatyBiasRaw
	s.AIAgreementBiasRaw[outer][inner] = result.AgreementBiasRaw
	if result.Policy != s.AIPolicies[outer][inner] {
		s.AIPolicies[outer][inner], s.AIPolicies[inner][outer] = result.Policy, result.Policy
		s.clearOriginalAIIncidentMemory(outer, inner)
		s.addOriginalTreatyCooldown(outer, inner, result.Policy, false)
	}
	if result.TradeActive {
		s.AITrade[outer][inner], s.AITrade[inner][outer] = true, true
	}
	if result.ResearchActive {
		s.AIResearch[outer][inner], s.AIResearch[inner][outer] = true, true
	}
	if result.TributeMode != 0 {
		s.AITributeModes[outer][inner] = result.TributeMode
		s.clearOriginalAIIncidentMemory(outer, inner)
		s.addOriginalTreatyCooldown(outer, inner, gamedata.DIPLO_NONE, true)
	}
	if result.RelationDelta != 0 {
		current := s.originalAIAIRelation(outer, inner)
		next, valid := gamedata.OriginalChangeRelationScore(gamedata.OriginalRelationChangeInput{
			CurrentRaw: current, BaseDelta: result.RelationDelta,
			ActorGovernment:   originalAIRelationGovernment(s.AIPlayers[inner]),
			TargetCharismatic: aiRaceHasTrait(s.AIPlayers[outer], gamedata.TRAIT_CHARISMATIC),
			Policy:            result.Policy, BothAI: true, RelativeTurn: s.Turn - 1, Difficulty: s.Difficulty,
		})
		if valid {
			s.setOriginalAIAIRelation(outer, inner, next)
			s.recordOriginalAIIncident(inner, outer, 14, next-current)
		}
	}
}

func (s *GameSession) exchangeAIBenefits(i, j int, trade, research bool) {
	if trade {
		// 每回合最多移轉 1 BC，避免協定在抽象經濟中取代殖民地收入。
		if s.AIPlayers[i].Player.BC > s.AIPlayers[j].Player.BC && s.AIPlayers[i].Player.BC > 0 {
			s.AIPlayers[i].Player.BC--
			s.AIPlayers[j].Player.BC++
		} else if s.AIPlayers[j].Player.BC > s.AIPlayers[i].Player.BC && s.AIPlayers[j].Player.BC > 0 {
			s.AIPlayers[j].Player.BC--
			s.AIPlayers[i].Player.BC++
		}
	}
	if research {
		// 研究協定分享 1 點目前研究進度；真正的科技選擇仍由各 AI 自己決定。
		if s.AIPlayers[i].Player.ResearchProgress > s.AIPlayers[j].Player.ResearchProgress {
			s.AIPlayers[i].Player.ResearchProgress--
			s.AIPlayers[j].Player.ResearchProgress++
		} else if s.AIPlayers[j].Player.ResearchProgress > s.AIPlayers[i].Player.ResearchProgress {
			s.AIPlayers[j].Player.ResearchProgress--
			s.AIPlayers[i].Player.ResearchProgress++
		}
	}
}

func (s *GameSession) aiWarTarget(i int) (int, int) {
	if i < 0 || i >= len(s.AIPlayers) || i >= len(s.AIWars) {
		return -1, -1
	}
	bestAI, bestStar, bestPop := -1, -1, -1
	for j := range s.AIPlayers {
		if i == j || j >= len(s.AIWars[i]) || !s.AIWars[i][j] {
			continue
		}
		for colonyIdx, star := range s.AIPlayers[j].ColonyStars {
			pop := 0
			if colonyIdx < len(s.AIPlayers[j].Colonies) {
				pop = s.AIPlayers[j].Colonies[colonyIdx].Population
			}
			if pop > bestPop || (pop == bestPop && (bestAI < 0 || j < bestAI)) {
				bestAI, bestStar, bestPop = j, star, pop
			}
		}
	}
	return bestAI, bestStar
}

// aiLaunchAIFleet 讓有正式戰爭目標的 AI 艦隊出發。每個 AI 仍只有一支
// 抽象艦隊；若它正在攻擊玩家，這回合不會再同時發動第二支 AI 戰爭艦隊。
func (s *GameSession) aiLaunchAIFleet(i int) bool {
	if s.DisableEvents || i < 0 || i >= len(s.AIPlayers) {
		return false
	}
	a := &s.AIPlayers[i]
	if a.FleetETA > 0 || a.FleetStrength <= 0 {
		return false
	}
	targetAI, dest := s.aiWarTarget(i)
	if targetAI < 0 || dest < 0 {
		return false
	}
	from := aiFleetStar(*a)
	if from < 0 {
		return false
	}
	if from == dest {
		s.LastAIAIBattle = s.resolveAIAIBattle(i, targetAI, dest)
		return true
	}
	eta := s.aiFleetETATo(*a, from, dest)
	if eta <= 0 {
		return false
	}
	a.FleetDestStar = dest
	a.FleetETA = eta
	a.FleetTargetAI = targetAI
	a.FleetTargetAISet = true
	return true
}

func (s *GameSession) resolveAIAIBattle(attacker, defender, star int) *AIAIBattleReport {
	if attacker < 0 || defender < 0 || attacker >= len(s.AIPlayers) || defender >= len(s.AIPlayers) || attacker == defender {
		return nil
	}
	a, d := &s.AIPlayers[attacker], &s.AIPlayers[defender]
	attackStrength := a.FleetStrength
	defenseStrength := d.FleetStrength + s.aiColonyDefense(defender, star)
	bonus := ((s.Turn*17 + attacker*11 + defender*7 + star*5) % aiAIBattleRandomSpan)
	if bonus < 0 {
		bonus += aiAIBattleRandomSpan
	}
	attackRoll := attackStrength + bonus
	defenseRoll := defenseStrength + (aiAIBattleRandomSpan - 1 - bonus)
	attackerWon := attackRoll >= defenseRoll
	attackerLoss := maxInt(1, defenseStrength/aiAIAttackerLossDivisor)
	defenderLoss := maxInt(1, attackStrength/aiAIDefenderLossDivisor)
	if attackerWon {
		s.reduceAIShipStrength(attacker, attackStrength-attackerLoss)
		s.reduceAIShipStrength(defender, d.FleetStrength-defenderLoss)
		s.transferAIColony(attacker, defender, star)
		a.FleetStar, a.FleetPosSet = star, true
	} else {
		s.reduceAIShipStrength(attacker, attackStrength-attackerLoss)
		s.reduceAIShipStrength(defender, d.FleetStrength-defenderLoss/2)
		if len(a.ColonyStars) > 0 {
			a.FleetStar, a.FleetPosSet = a.ColonyStars[0], true
		}
	}
	report := &AIAIBattleReport{
		AttackerIndex: attacker, DefenderIndex: defender,
		AttackerName: a.Name, DefenderName: d.Name,
		StarIdx: star, StarName: s.starName(star), AttackerWon: attackerWon,
		AttackerRoll: attackRoll, DefenderRoll: defenseRoll,
		AttackerLoss: attackerLoss, DefenderLoss: defenderLoss,
	}
	if attackerWon {
		report.Message = a.Name + " 擊敗 " + d.Name + "，佔領 " + report.StarName
		report.MessageEN = a.Name + " defeated " + d.Name + " and captured " + report.StarName
	} else {
		report.Message = a.Name + " 的艦隊被 " + d.Name + " 擊退"
		report.MessageEN = a.Name + " was repelled by " + d.Name
	}
	return report
}

func (s *GameSession) aiColonyDefense(aiIdx, star int) int {
	a := &s.AIPlayers[aiIdx]
	idx := aiColonyIndexAtStar(*a, star)
	if idx < 0 || idx >= len(a.Colonies) {
		return 0
	}
	return maxInt(1, a.Colonies[idx].Population*aiAIColonyDefensePerPop)
}

func aiColonyIndexAtStar(a AIOpponent, star int) int {
	for i, candidate := range a.ColonyStars {
		if candidate == star {
			return i
		}
	}
	return -1
}

func (s *GameSession) transferAIColony(attacker, defender, star int) {
	if attacker < 0 || defender < 0 || attacker >= len(s.AIPlayers) || defender >= len(s.AIPlayers) {
		return
	}
	a, d := &s.AIPlayers[attacker], &s.AIPlayers[defender]
	idx := aiColonyIndexAtStar(*d, star)
	if idx < 0 || idx >= len(d.Colonies) {
		return
	}
	ensureAIGroundForceSlots(a)
	ensureAIGroundForceSlots(d)
	a.Colonies = append(a.Colonies, d.Colonies[idx])
	a.ColonyStars = append(a.ColonyStars, star)
	if idx < len(d.ColonyPlanets) {
		a.ColonyPlanets = append(a.ColonyPlanets, d.ColonyPlanets[idx])
	} else {
		a.ColonyPlanets = append(a.ColonyPlanets, -1)
	}
	if idx < len(d.ColonyBuildings) {
		a.ColonyBuildings = append(a.ColonyBuildings, cloneBuildings(d.ColonyBuildings[idx]))
	} else {
		a.ColonyBuildings = append(a.ColonyBuildings, nil)
	}
	a.ColonyMarines = append(a.ColonyMarines, d.ColonyMarines[idx])
	a.ColonyTanks = append(a.ColonyTanks, d.ColonyTanks[idx])
	a.MarineBarracksAge = append(a.MarineBarracksAge, d.MarineBarracksAge[idx])
	a.ArmorBarracksAge = append(a.ArmorBarracksAge, d.ArmorBarracksAge[idx])
	d.Colonies = append(d.Colonies[:idx], d.Colonies[idx+1:]...)
	d.ColonyStars = append(d.ColonyStars[:idx], d.ColonyStars[idx+1:]...)
	if idx < len(d.ColonyPlanets) {
		d.ColonyPlanets = append(d.ColonyPlanets[:idx], d.ColonyPlanets[idx+1:]...)
	}
	if idx < len(d.ColonyBuildings) {
		d.ColonyBuildings = append(d.ColonyBuildings[:idx], d.ColonyBuildings[idx+1:]...)
	}
	removeAIGroundForceSlot(d, idx)
	a.OwnedStars = len(a.ColonyStars)
	d.OwnedStars = len(d.ColonyStars)
	if star >= 0 && star < len(s.Stars) {
		s.Stars[star].Owner = 2 // Star.Owner 只有「玩家／AI」粒度，AI 細分由 ColonyStars 決定。
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
