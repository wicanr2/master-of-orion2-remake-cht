package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// applyResearchRaceTrait 把研究完成後的種族規則接在引擎完成邊界。
//
// 原版手冊的語意是:
//   - Creative:研究一個領域時拿到該領域全部應用。
//   - Uncreative:每個領域只隨機拿一項應用。
//
// engine.RunResearchPhase 仍保留一般種族的純研究結果(多選主題先記第一項並掛
// PendingChoice),所以這個函式只負責把種族規則轉成既有的 ChosenTech/
// ExplicitChoice 表示法。未明確抉擇的完成主題在 componentUnlockedFor/psKnowsTech
// 會視為該領域全解,正好是 Creative 所需的既有語意。
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
