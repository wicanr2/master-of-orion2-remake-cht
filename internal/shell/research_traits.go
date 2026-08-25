package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// applyResearchRaceTrait 只處理舊存檔／直接 engine caller 未在研究前建立 application 的
// 相容狀態。正常遊戲已由 preparePlayerResearchApplication／prepareAIResearchApplication
// 在投入 RP 前套用種族規則。
//
// 原版手冊的語意是:
//   - Creative:研究一個領域時拿到該領域全部應用。
//   - Uncreative:每個領域只隨機拿一項應用。
//
// 若舊狀態突破後才掛 PendingChoice，這裡仍把它收斂成既有 ChosenTech／ExplicitChoice
// 表示法，避免舊存檔永久卡住。
//
// choose 是外部注入的 [0,n) 選擇器。正常遊戲傳入可存檔的 researchRand.Intn;
// 測試可傳入固定選擇,不在這裡偷偷建立不可重現的全域亂數。
func applyResearchRaceTrait(ps *engine.PlayerState, topic gamedata.ResearchTopic, creative, uncreative bool, choose func(int) int) {
	if ps == nil || !ps.HasPendingChoice {
		return
	}
	choice := gamedata.ResearchChoiceFor(topic)
	if len(choice.Choices) <= 1 {
		return
	}

	switch {
	case creative:
		// 未明確抉擇是本 remake 的「主題層級全解」表示法。若舊狀態留下
		// ExplicitChoice,移除它才能保證 Creative 的全部應用仍然可見。
		if ps.ExplicitChoice != nil {
			delete(ps.ExplicitChoice, topic)
		}
		ps.HasPendingChoice = false
		ps.PendingChoice = 0
	case uncreative:
		if choose == nil {
			return
		}
		idx := choose(len(choice.Choices))
		if idx < 0 || idx >= len(choice.Choices) {
			idx = 0
		}
		next, ok := engine.ApplyResearchChoice(*ps, choice.Choices[idx])
		if ok {
			*ps = next
			ps.PendingChoice = 0
		}
	}
}

// researchRandForTurn 回傳研究專用的長壽命亂數流。
//
// 研究選項不能共用事件／間諜亂數流:否則一次研究完成會改變之後事件與諜報的
// 序列。流的位置由 sessionSnapshot 的 ResearchDraws 保存,存讀檔後仍能重現
// Uncreative 的選擇。
func (s *GameSession) researchRandForTurn() *randStream {
	if s.researchRand == nil {
		s.researchRand = newRandStream(s.EventSeed*2654435761 + 13)
	}
	return s.researchRand
}

func (s *GameSession) researchBreakthroughRoll(max int) int {
	if max <= 0 {
		return 0
	}
	return s.researchRandForTurn().Intn(max) + 1
}

func (s *GameSession) applyPlayerResearchRaceTrait(researchDone bool) {
	if !researchDone {
		return
	}
	applyResearchRaceTrait(&s.Player, s.Player.ResearchTopic,
		s.RaceCreative(), s.RaceUncreative(), s.researchRandForTurn().Intn)
}

func (s *GameSession) applyAIResearchRaceTrait(a *AIOpponent, researchDone bool) {
	if a == nil || !researchDone {
		return
	}
	applyResearchRaceTrait(&a.Player, a.Player.ResearchTopic,
		aiRaceHasTrait(*a, gamedata.TRAIT_CREATIVE),
		aiRaceHasTrait(*a, gamedata.TRAIT_UNCREATIVE),
		s.researchRandForTurn().Intn)
}
