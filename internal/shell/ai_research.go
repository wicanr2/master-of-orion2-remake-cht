package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// ai_research.go:AI 對手**自己研究科技**。
//
// ============ 先前是壞的,而且壞得很安靜 ============
//
// 探針(NewDemoSession 跑 200 回合)量到的:三個 AI 的 `ResearchTopic` 從第 1 回合到
// 第 200 回合**都是 1**,而同一局的玩家跑過 3 → 31 → 9 → 15。
//
// 原因:`engine.RunEmpireTurn` 對 AI 一樣跑研究階段(`session.go` 有呼叫),主題完成也會
// 標進 `CompletedTopics`——**但沒有任何地方替 AI 選下一個主題**。
// `ai.DecideResearchTopic` 從寫出來到現在**零個呼叫端**。
//
// 所以 AI 每一回合都把研究點投進一個早就完成的主題,無限重複完成同一項。
//
// 那 AI 的科技數為什麼還是會長?**它偷的。** `spy.go` 的間諜每隔一段時間從玩家那裡偷一個
// 主題——AI 對手的整條科技線,先前**完全來自玩家**。玩家不研究,AI 就永遠停在開局。
//
// 這件事在畫面上看不出來(沒有任何 UI 顯示 AI 在研究什麼),但它讓每一個「吃 AI 科技」的
// 系統都凍在第 1 回合:軌道防禦挑得到的最佳武器、AI 艦隊的航速、AI 地面部隊的裝備。
//
// 多選 application 也已從舊版「突破後 PendingChoice」改為研究開始前選定；
// prepareAIResearchApplication 對映原版 sub_DC288 的 application→field 寫入順序。
//
// ============ 挑哪一項：常態回合走原版 application 級抽選 ============
//
// 新局 AI 保存 sub_589D6 的 raw6/raw4/raw7 profile；目前 field 完成時，
// selectOriginalAIResearch 依 sub_DC288 → sub_FD335 對所有可用 application 做一次估值抽選，
// 抽中的 application 同時決定 field。只有舊存檔沒有 raw profile 時，才回退
// `ai.DecideResearchTopic` 的 remake 設計啟發式。
//
// aiPickApplication 只保留給舊 pending-choice 存檔與 profile fallback；正常新局不再用
// category 最大值做第二次 application 選擇。

// aiResearchCandidates 組出這個 AI 現在可以研究的主題(每個研究領域各一個隊首)。
//
// 為什麼不直接用 `gamedata.AvailableTopics`:那支回傳的是扁平清單,**已研究完的領域會被
// 略過**,所以回傳位置對不回領域索引。而 `ai.ResearchCandidate` 需要領域索引來判斷
// 「這是不是軍事領域」。這裡直接走 `TechTree()` 保住索引。
func aiResearchCandidates(ps engine.PlayerState) []ai.ResearchCandidate {
	tree := gamedata.TechTree()
	out := make([]ai.ResearchCandidate, 0, len(tree))
	for area, topics := range tree {
		for _, t := range topics {
			if ps.CompletedTopics[t] && !gamedata.IsHyperAdvancedTopic(t) {
				continue
			}
			cost := gamedata.ResearchChoiceFor(t).Cost
			if gamedata.IsHyperAdvancedTopic(t) {
				base := cost
				if ps.HyperAdvancedResearchCost > 0 {
					base = ps.HyperAdvancedResearchCost
				}
				level := ps.HyperAdvancedLevels[t]
				if ps.HyperAdvancedLevels == nil && ps.CompletedTopics[t] {
					level = 1
				}
				cost = gamedata.HyperAdvancedRepeatedCost(base, level)
			}
			out = append(out, ai.ResearchCandidate{TopicID: int(t), Cost: cost, AreaIndex: area})
			break // 每個領域只有隊首那一個能選(見 AvailableTopics 註解:這是原版規則)
		}
	}
	return out
}

// aiPickApplication 在多選主題裡挑一項,依原版的 category 倍率取最高。
//
// 回傳 (科技, true);choices 為空時回 (0, false)。同分取先出現的那個。
func aiPickApplication(choices []gamedata.Technology) (gamedata.Technology, bool) {
	best, bestW, found := gamedata.Technology(0), -1, false
	for _, t := range choices {
		w := gamedata.TechCategoryWeight(t)
		if w > bestW {
			best, bestW, found = t, w, true
		}
	}
	return best, found
}

// knownTechnologyApplications 把 typed PlayerState 投影成 sub_FC845 使用的 application 已知集合。
// 明確抉擇只加入被選中的 application；ResearchAll／舊存檔未記抉擇時加入該 field 全部應用。
func knownTechnologyApplications(ps engine.PlayerState) map[gamedata.Technology]bool {
	known := map[gamedata.Technology]bool{}
	for tech, granted := range ps.GrantedTechs {
		if granted {
			known[tech] = true
		}
	}
	for topic, completed := range ps.CompletedTopics {
		if !completed {
			continue
		}
		choices := gamedata.ResearchChoicesForTopic(topic)
		if ps.ExplicitChoice != nil && ps.ExplicitChoice[topic] {
			if tech, ok := ps.ChosenTech[topic]; ok {
				known[tech] = true
			}
			continue
		}
		for _, tech := range choices {
			known[tech] = true
		}
	}
	return known
}

// selectOriginalAIResearch 依 sub_DC288 → sub_FD335 的單次 application 級抽選同時決定
// field 與 application。profile 未知時回 false，讓舊存檔走明示的 remake fallback。
func (s *GameSession) selectOriginalAIResearch(i, researchPerTurn int) bool {
	if i < 0 || i >= len(s.AIPlayers) {
		return false
	}
	a := &s.AIPlayers[i]
	if !a.OriginalTechProfileKnown {
		return false
	}
	available := gamedata.AvailableTopics(a.Player.CompletedTopics)
	if len(available) == 0 {
		return false
	}
	opponents := make([]map[gamedata.Technology]bool, 0, len(s.AIPlayers))
	opponents = append(opponents, knownTechnologyApplications(s.Player))
	for j := range s.AIPlayers {
		if j != i {
			opponents = append(opponents, knownTechnologyApplications(s.AIPlayers[j].Player))
		}
	}
	state := gamedata.OriginalStartingValueState{
		Difficulty: s.Difficulty, RelativeTurn: s.Turn,
		AIProfile: a.OriginalTechProfile, AIProfileKnown: true,
		Raw4: a.OriginalTechProfile.Raw4, Raw4Known: true,
		Known: knownTechnologyApplications(a.Player), Opponents: opponents,
	}
	topic, tech, ok := gamedata.StartingOriginalApplicationPick(
		available, researchPerTurn, state, s.researchRandForTurn().Intn)
	if !ok {
		return false
	}
	a.Player.ResearchTopic = topic
	a.Player.ResearchApplication = 0
	a.Player.HasResearchApplication = false
	if !gamedata.ResearchTopicGrantsAll(topic) && !aiRaceHasTrait(*a, gamedata.TRAIT_CREATIVE) {
		if next, selected := engine.SelectResearchApplication(a.Player, topic, tech); selected {
			a.Player = next
		}
	}
	return true
}

// aiResolveResearchChoice 只相容舊存檔中突破後仍掛著的多選主題。
//
// 沒有待決時什麼都不做。走 `engine.ApplyResearchChoice` 而不是自己寫 map——那支會一併
// 設 `ExplicitChoice` 並清掉待決旗標,繞過它就會留下「選了但沒標明確抉擇」的半套狀態。
func aiResolveResearchChoice(ps engine.PlayerState) engine.PlayerState {
	if !ps.HasPendingChoice {
		return ps
	}
	choice := gamedata.ResearchChoiceFor(ps.PendingChoice)
	tech, ok := aiPickApplication(choice.Choices)
	if !ok {
		// 掛著待決卻沒有選項:清掉旗標,否則它會永遠卡在這裡擋住後面每一次抉擇。
		ps.HasPendingChoice = false
		return ps
	}
	next, _ := engine.ApplyResearchChoice(ps, tech)
	return next
}

// prepareAIResearchApplication 對映原版 sub_DC288：AI 在投入研究前先決定 application。
func (s *GameSession) prepareAIResearchApplication(a *AIOpponent) {
	if a == nil {
		return
	}
	ps := &a.Player
	choice := gamedata.ResearchChoiceFor(ps.ResearchTopic)
	if ps.ResearchTopic == 0 || ps.CompletedTopics[ps.ResearchTopic] || len(choice.Choices) == 0 ||
		choice.ResearchAll || aiRaceHasTrait(*a, gamedata.TRAIT_CREATIVE) {
		ps.ResearchApplication = 0
		ps.HasResearchApplication = false
		ps.PendingChoice = 0
		ps.HasPendingChoice = false
		return
	}
	if ps.HasResearchApplication {
		return
	}
	var tech gamedata.Technology
	var ok bool
	if aiRaceHasTrait(*a, gamedata.TRAIT_UNCREATIVE) {
		tech = choice.Choices[s.researchRandForTurn().Intn(len(choice.Choices))]
		ok = true
	} else {
		tech, ok = aiPickApplication(choice.Choices)
	}
	if next, selected := engine.SelectResearchApplication(*ps, ps.ResearchTopic, tech); ok && selected {
		*ps = next
	}
}

// advanceAIResearch 讓第 i 個 AI 對手:先處理待決的科技抉擇,再在目前主題已完成時挑下一個。
//
// **每回合都呼叫**,不是只在「本回合有研究完成」時。理由:主題也可能是被間諜偷來的
// (`spy.go` 直接寫 `CompletedTopics`,不經過研究階段),那時候 `ResearchDone` 是 false
// 但目前主題已經完成了——只看 ResearchDone 會讓 AI 卡在一個偷來的主題上繼續投點。
func (s *GameSession) advanceAIResearch(i int, researchPerTurn ...int) {
	a := &s.AIPlayers[i]
	a.Player = aiResolveResearchChoice(a.Player)

	if a.Player.CompletedTopics != nil && !a.Player.CompletedTopics[a.Player.ResearchTopic] {
		return // 還在研究中
	}
	cands := aiResearchCandidates(a.Player)
	if len(cands) == 0 {
		return // 整棵科技樹研究完了——保持原樣,不要亂設一個已完成的主題
	}
	rp := 1
	if len(researchPerTurn) > 0 && researchPerTurn[0] > 0 {
		rp = researchPerTurn[0]
	}
	if s.selectOriginalAIResearch(i, rp) {
		return
	}
	if id := ai.DecideResearchTopic(cands, aiProfile(*a)); id >= 0 {
		a.Player.ResearchTopic = gamedata.ResearchTopic(id)
		a.Player.ResearchApplication = 0
		a.Player.HasResearchApplication = false
		s.prepareAIResearchApplication(a)
	}
}
