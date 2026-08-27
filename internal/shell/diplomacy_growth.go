package shell

import "github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"

// originalAIRelationGovernment 回傳 Change_Relations_ 所讀 player+0x27 的 raw
// 政體值。remake 尚未保存 AI 當局政府，故依 READY 規格以該種族的
// 原版開局政體作強推論 fallback；無法定位時才使用 Dictatorship。
func originalAIRelationGovernment(a AIOpponent) int {
	race := aiRaceIndex(a)
	if race >= 0 && race < len(Races) {
		return gamedata.OrigRaceTrait(Races[race].OrigIdx, gamedata.TRAIT_GOVERNMENT)
	}
	return int(gamedata.MoraleGovDictatorship)
}

// originalRelationRaw 維持 -100..100 的 raw 餘數。若既有玩法直接改寫
// normalized Relation，投影值會不一致，此時以新玩家狀態重建 raw，
// 不讓舊餘數在下回合把玩家動作沖掉。
func (a *AIOpponent) originalRelationRaw() int {
	if a == nil {
		return 0
	}
	normalized := clampRelation(a.Relation)
	if !a.OriginalRelationKnown || normalizedRelationFromOriginal(a.OriginalRelationRaw) != normalized {
		a.OriginalRelationRaw = originalRelationFromNormalized(normalized)
		a.OriginalRelationKnown = true
	}
	return a.OriginalRelationRaw
}

func (s *GameSession) diplomacyGrowthRandForTurn() *randStream {
	if s.diplomacyGrowthRand == nil {
		s.diplomacyGrowthRand = newRandStream(s.EventSeed*2654435761 + 37)
	}
	return s.diplomacyGrowthRand
}

func (s *GameSession) originalAIRelationTarget(a *AIOpponent) int {
	if a == nil {
		return 0
	}
	if a.OriginalRelationTargetKnown {
		return a.OriginalRelationTargetRaw
	}
	value := 0
	if aiRace := aiRaceIndex(*a); aiRace >= 0 && aiRace < len(Races) {
		if base, ok := gamedata.OriginalBaseRelation(Races[aiRace].OrigIdx, s.raceOrigIdx()); ok {
			value = base
		}
	}
	// 自訂種族不在原版 14×14 表；零是失敗即關閉的中立投影，不冒稱原版值。
	a.OriginalRelationTargetRaw = value
	a.OriginalRelationTargetKnown = true
	return value
}

// advanceOriginalDiplomacyGrowthForAI 把 Diplomacy_Growth_ @ 0x4DD6B 已閉合的
// 玩家↔AI 條約分支接回每回合正常路徑。原版全域 PRNG 不可在
// remake 中逐位元還原，因此使用可存檔獨立流保證鎖步可重播。
func (s *GameSession) advanceOriginalDiplomacyGrowthForAI(index int) {
	s.advanceOriginalDiplomacyGrowthForAIWithRoller(index,
		func(n int) int { return s.diplomacyGrowthRandForTurn().Intn(n) + 1 })
}

func (s *GameSession) advanceOriginalDiplomacyGrowthForAIWithRoller(index int, roll func(int) int) {
	if index < 0 || index >= len(s.AIPlayers) {
		return
	}
	a := &s.AIPlayers[index]
	raw := a.originalRelationRaw()
	next, ok := gamedata.OriginalDiplomacyGrowthTreatyRelation(
		gamedata.OriginalDiplomacyGrowthTreatyInput{
			CurrentRaw: raw, FormalPolicy: a.Treaty.FormalPolicy,
			TradeActive: a.Treaty.TradeActive, ResearchActive: a.Treaty.ResearchActive,
			TributeMode: int(a.Treaty.PlayerTribute), ActorGovernment: originalAIRelationGovernment(*a),
			TargetCharismatic: s.RaceCharismatic(),
		}, roll,
	)
	if !ok {
		return
	}
	a.OriginalRelationRaw = next
	a.OriginalRelationKnown = true
	a.Relation = normalizedRelationFromOriginal(next)
}

func (s *GameSession) ensureOriginalAIAIRelations() {
	s.ensureAIRelations()
	n := len(s.AIPlayers)
	s.AIRelationsRaw = resizeIntMatrix(s.AIRelationsRaw, n)
	s.AIRelationsRawKnown = resizeBoolMatrix(s.AIRelationsRawKnown, n)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if s.AIRelationsRawKnown[i][j] {
				continue
			}
			raw := 0
			loRace, hiRace := aiRaceIndex(s.AIPlayers[i]), aiRaceIndex(s.AIPlayers[j])
			if loRace >= 0 && loRace < len(Races) && hiRace >= 0 && hiRace < len(Races) {
				// 0x4E282 的逐列轉置使每對最終保留高槽 observer→低槽 target。
				raw, _ = gamedata.OriginalBaseRelation(Races[hiRace].OrigIdx, Races[loRace].OrigIdx)
			}
			// 舊存檔若已有非零顯示關係，優先保留玩家已形成的狀態。
			if shown := (s.AIRelations[i][j] + s.AIRelations[j][i]) / 2; shown != 0 {
				raw = originalRelationFromNormalized(shown)
			}
			s.setOriginalAIAIRelation(i, j, raw)
		}
	}
}

func resizeIntMatrix(old [][]int, n int) [][]int {
	out := make([][]int, n)
	for i := range out {
		out[i] = make([]int, n)
		if i < len(old) {
			copy(out[i], old[i])
		}
	}
	return out
}

func (s *GameSession) setOriginalAIAIRelation(i, j, raw int) {
	s.AIRelationsRaw[i][j], s.AIRelationsRaw[j][i] = raw, raw
	s.AIRelationsRawKnown[i][j], s.AIRelationsRawKnown[j][i] = true, true
	shown := normalizedRelationFromOriginal(raw)
	s.AIRelations[i][j], s.AIRelations[j][i] = shown, shown
}

func (s *GameSession) originalAIAIRelation(i, j int) int {
	current := s.AIRelationsRaw[i][j]
	shown := (s.AIRelations[i][j] + s.AIRelations[j][i]) / 2
	if normalizedRelationFromOriginal(current) != shown {
		current = originalRelationFromNormalized(shown)
		s.setOriginalAIAIRelation(i, j, current)
	}
	return current
}

func (s *GameSession) originalAIAITarget(i, j int) int {
	loRace, hiRace := aiRaceIndex(s.AIPlayers[i]), aiRaceIndex(s.AIPlayers[j])
	if loRace < 0 || loRace >= len(Races) || hiRace < 0 || hiRace >= len(Races) {
		return 0
	}
	value, _ := gamedata.OriginalBaseRelation(Races[hiRace].OrigIdx, Races[loRace].OrigIdx)
	return value
}

func (s *GameSession) originalAIAIPolicy(i, j int) gamedata.ForeignPolicy {
	if i >= 0 && i < len(s.AIPolicies) && j >= 0 && j < len(s.AIPolicies[i]) {
		return s.AIPolicies[i][j]
	}
	return gamedata.DIPLO_NONE
}

// advanceOriginalDiplomacyGrowth 保留原版兩階段順序：全部關係邊先消費條約，
// 再全部消費 +0x61F 目標漂移。AI↔AI 每對依原版鏡射保留高槽→低槽結果。
func (s *GameSession) advanceOriginalDiplomacyGrowth() {
	roll := func(n int) int { return s.diplomacyGrowthRandForTurn().Intn(n) + 1 }
	s.ensureOriginalAIAIRelations()
	// 第一階段：玩家↔AI 條約，再依槽位順序處理 AI↔AI pair。
	for i := range s.AIPlayers {
		s.advanceOriginalDiplomacyGrowthForAIWithRoller(i, roll)
	}
	for i := 0; i < len(s.AIPlayers); i++ {
		for j := i + 1; j < len(s.AIPlayers); j++ {
			policy := s.originalAIAIPolicy(i, j)
			trade := i < len(s.AITrade) && j < len(s.AITrade[i]) && s.AITrade[i][j]
			research := i < len(s.AIResearch) && j < len(s.AIResearch[i]) && s.AIResearch[i][j]
			next, ok := gamedata.OriginalDiplomacyGrowthTreatyRelation(
				gamedata.OriginalDiplomacyGrowthTreatyInput{
					CurrentRaw: s.originalAIAIRelation(i, j), FormalPolicy: policy,
					TradeActive: trade, ResearchActive: research,
					ActorGovernment:   originalAIRelationGovernment(s.AIPlayers[j]),
					TargetCharismatic: aiRaceHasTrait(s.AIPlayers[i], gamedata.TRAIT_CHARISMATIC),
				}, roll)
			if ok {
				s.setOriginalAIAIRelation(i, j, next)
			}
		}
	}
	// 第二階段：玩家↔AI 漂移。
	for i := range s.AIPlayers {
		a := &s.AIPlayers[i]
		next, ok := gamedata.OriginalDiplomacyRelationDrift(
			gamedata.OriginalDiplomacyRelationDriftInput{
				CurrentRaw: a.originalRelationRaw(),
				TargetRaw:  s.originalAIRelationTarget(a),
				Policy:     a.Treaty.FormalPolicy,
				// +0x737 的寫入語意尚未閉合；玩家↔AI 正常邊採未鎖定。
			}, roll)
		if !ok {
			continue
		}
		a.OriginalRelationRaw = next
		a.OriginalRelationKnown = true
		a.Relation = normalizedRelationFromOriginal(next)
	}
	// 第二階段續：AI↔AI 漂移；+0x737 writer 未閉合，正常邊採未鎖定。
	for i := 0; i < len(s.AIPlayers); i++ {
		for j := i + 1; j < len(s.AIPlayers); j++ {
			next, ok := gamedata.OriginalDiplomacyRelationDrift(
				gamedata.OriginalDiplomacyRelationDriftInput{
					CurrentRaw: s.originalAIAIRelation(i, j),
					TargetRaw:  s.originalAIAITarget(i, j),
					Policy:     s.originalAIAIPolicy(i, j),
				}, roll)
			if ok {
				s.setOriginalAIAIRelation(i, j, next)
			}
		}
	}
}
