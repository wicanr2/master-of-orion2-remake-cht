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
	aiWarRelationThreshold  = -25
	aiCeasefireRelation     = 12
	aiAllianceRelation      = 25
	aiTradeRelation         = 8
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

// advanceAIAIDiplomacy 把 AIRelations 轉成正式政策，並讓貿易／研究協定產生
// 可觀察但保守的資源交換。政策是雙方視角分開保存，戰爭旗標則對稱。
func (s *GameSession) advanceAIAIDiplomacy() {
	s.ensureAIAIState()
	for i := 0; i < len(s.AIPlayers); i++ {
		for j := i + 1; j < len(s.AIPlayers); j++ {
			r := (s.AIRelations[i][j] + s.AIRelations[j][i]) / 2
			atWar := s.AIWars[i][j] || s.AIWars[j][i]
			if atWar && r >= aiCeasefireRelation {
				atWar = false
			} else if !atWar && r <= aiWarRelationThreshold {
				atWar = true
			}
			s.AIWars[i][j], s.AIWars[j][i] = atWar, atWar
			if atWar {
				s.AIPolicies[i][j], s.AIPolicies[j][i] = gamedata.DIPLO_WAR, gamedata.DIPLO_WAR
				s.AITrade[i][j], s.AITrade[j][i] = false, false
				s.AIResearch[i][j], s.AIResearch[j][i] = false, false
				continue
			}

			policy := gamedata.DIPLO_PEACE
			switch {
			case r >= aiAllianceRelation:
				policy = gamedata.DIPLO_ALLIANCE
			case r >= aiTradeRelation:
				policy = gamedata.DIPLO_NON_AGGRESSION
			case r < 0:
				policy = gamedata.DIPLO_LIMITED_WAR
			}
			s.AIPolicies[i][j], s.AIPolicies[j][i] = policy, policy
			trade := r >= aiTradeRelation
			research := r >= aiAllianceRelation
			s.AITrade[i][j], s.AITrade[j][i] = trade, trade
			s.AIResearch[i][j], s.AIResearch[j][i] = research, research
			if trade || research {
				s.exchangeAIBenefits(i, j, trade, research)
			}
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
