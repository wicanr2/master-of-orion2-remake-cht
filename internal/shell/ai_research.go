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
// ============ 挑哪一項:用原版的 category 倍率 ============
//
// 主題**選哪個**用 `ai.DecideResearchTopic`(那是 remake 設計的啟發式,吃性格;
// 原版的 `Calc_Tech_Value_` 估的是**科技應用項**不是主題,而且它的核心那幾段仍卡在
// `word_1AB1xx` 的語意上——見 `docs/re/calc-tech-value.md`)。
//
// 一般 AI **開始研究前選哪一個應用**,使用原版的一手資料:`gamedata.TechCategoryWeight`
// 就是 `Calc_Tech_Value_` 階段 B 給 `ecx` 的初始值(`byte_17D196[category*2]`),
// 而那一段是該文件第 7 節明列的「風險遠低於其他階段、可以先照抄」的部分。
//
// 同分取**選項清單裡先出現的那個**——不擲骰,整條 AI 研究線因此是決定性的
// (`determinism_test.go` 那組閘門守著的前提)。

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
func (s *GameSession) advanceAIResearch(i int) {
	a := &s.AIPlayers[i]
	a.Player = aiResolveResearchChoice(a.Player)

	if a.Player.CompletedTopics != nil && !a.Player.CompletedTopics[a.Player.ResearchTopic] {
		return // 還在研究中
	}
	cands := aiResearchCandidates(a.Player)
	if len(cands) == 0 {
		return // 整棵科技樹研究完了——保持原樣,不要亂設一個已完成的主題
	}
	if id := ai.DecideResearchTopic(cands, aiProfile(*a)); id >= 0 {
		a.Player.ResearchTopic = gamedata.ResearchTopic(id)
		a.Player.ResearchApplication = 0
		a.Player.HasResearchApplication = false
		s.prepareAIResearchApplication(a)
	}
}
