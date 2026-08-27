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

// advanceOriginalDiplomacyGrowth 保留原版兩階段順序：先讓全部邊消費條約
// 成長，再讓全部邊消費 +0x61F 目標漂移。remake 目前只有玩家↔AI 關係邊。
func (s *GameSession) advanceOriginalDiplomacyGrowth() {
	roll := func(n int) int { return s.diplomacyGrowthRandForTurn().Intn(n) + 1 }
	for i := range s.AIPlayers {
		s.advanceOriginalDiplomacyGrowthForAIWithRoller(i, roll)
	}
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
}
