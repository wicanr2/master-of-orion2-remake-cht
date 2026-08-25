package shell

import (
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// research.go:可玩遊戲殼的研究主題選單資料(純邏輯)。
//
// 說明:原版每個研究主題的名稱存在遊戲資料 LBX(執行期才載入)。83 個 topic 的英文顯示名
// 現由 gamedata.TopicEnglishName 提供(權威來源,= tech.tsv 的 i18n key);shell 層不 import
// i18n,故 ResearchTopicName 回英文名,由 cmd/moo2 顯示端經 i18n catalog 翻中文
// (與其他畫面字串一致的顯示層翻譯做法)。成本一律以 gamedata 為準(不硬抄)。

// ResearchOption 是一個可選研究主題。
type ResearchOption struct {
	Topic gamedata.ResearchTopic
	Name  string
	Cost  int
}

// AvailableResearchTopics 回傳這一刻真的可以選的研究主題:**8 個領域各一個**,
// 取該領域第一個尚未完成的主題；走到底後仍提供可重複的 Hyper-Advanced。
//
// ============ 這裡先前是一份手挑的清單 ============
//
// 舊的 `StarterResearchTopics` 是寫死的 9 個「新手可選的早期主題」。拿原版的樹一比,
// 那份清單在開局那一刻是錯的:
//
//	不該出現(領域裡前面還有沒研究完的):進階建築學、進階生物學、人工智慧、進階化學
//	漏掉了(該領域的隊首):太空生物學、核融合物理學、光電子學
//
// 原因是原版的研究**每個領域是一條線**,同時只有隊首那一個能選——主題表
// `word_17D90C` 每筆只有一個後繼,完成一個才解鎖下一個(見
// gamedata/orig_research_table.go)。手挑清單看起來合理,但它跳過了順序。
//
// s 為 nil 時回傳「什麼都沒完成」的狀態,可當純資料查詢用。
func AvailableResearchTopics(s *GameSession) []ResearchOption {
	var completed map[gamedata.ResearchTopic]bool
	if s != nil {
		completed = s.Player.CompletedTopics
	}
	topics := gamedata.AvailableTopics(completed)
	out := make([]ResearchOption, 0, len(topics))
	for _, t := range topics {
		cost := ResearchCost(t)
		if s != nil {
			cost = s.ResearchCostForDisplay(t)
		}
		out = append(out, ResearchOption{Topic: t, Name: ResearchTopicName(t), Cost: cost})
	}
	return out
}

// ResearchCost 回傳主題完成所需研究點(RP),取自 gamedata。套件級純函式,不吃版本規則
// profile——Hyper-Advanced 第一級主題(gamedata.IsHyperAdvancedTopic)一律回套件級硬編值
// 25000(= 現行 Profile15 行為)。畫面需要顯示「這局(可能是 1.3)實際要花多少 RP」時改用
// (*GameSession).ResearchCostForDisplay。
func ResearchCost(t gamedata.ResearchTopic) int {
	return gamedata.ResearchChoiceFor(t).Cost
}

// ResearchCostForDisplay 同 ResearchCost,但套用這局遊戲的版本規則 profile(s.RuleProfile)——
// Hyper-Advanced 主題改讀 profile 基礎值，再依原版 Player_Research_Cost_ @ 0xE1E96
// 加上已完成 level×10000。gamedata.HyperAdvancedCost(s.RuleProfile) 提供
// 1.3=15000／1.5=25000 的基礎值，其餘主題與 ResearchCost
// 相同。是 gamedata.HyperAdvancedCost 註解點名的「顯示層接線」:資料層(ruleprofile.go)已備妥,
// 這裡接進研究選單/帝國概況畫面,讓玩家在 1.3 局裡看到的成本真的是 1.3 值,不是永遠顯示 1.5
// 硬編值再讓實際結算(RunResearchPhase)用不同數字扣款的顯示/結算不一致。
func (s *GameSession) ResearchCostForDisplay(t gamedata.ResearchTopic) int {
	if gamedata.IsHyperAdvancedTopic(t) {
		level := 0
		if s.Player.HyperAdvancedLevels != nil {
			level = s.Player.HyperAdvancedLevels[t]
		} else if s.Player.CompletedTopics[t] {
			level = 1 // 尚未經研究結算遷移的舊 JSON
		}
		return gamedata.HyperAdvancedRepeatedCost(gamedata.HyperAdvancedCost(s.RuleProfile), level)
	}
	return gamedata.ResearchChoiceFor(t).Cost
}

// ResearchTopicName 回傳主題的英文顯示名(83 個 topic 全收錄,= gamedata.TopicEnglishName,
// 也就是 assets/i18n/tech.tsv 的 i18n key)。shell 層不 import i18n,由 cmd/moo2 顯示端
// 經 catalog 翻中文(見 cmd/moo2/topicname.go 的 topicNameZh),與其他畫面字串一致。
func ResearchTopicName(t gamedata.ResearchTopic) string {
	return gamedata.TopicEnglishName(t)
}

// PendingResearchChoice 回傳目前 topic 在投入研究前尚待選定的 application。
// 舊存檔也可能帶有一次「突破後待改選」狀態，ChooseResearchTech 會相容處理。
func (s *GameSession) PendingResearchChoice() (topic gamedata.ResearchTopic, choices []gamedata.Technology, ok bool) {
	if !s.Player.HasPendingChoice {
		return 0, nil, false
	}
	t := s.Player.PendingChoice
	return t, gamedata.ResearchChoiceFor(t).Choices, true
}

// ChooseResearchTech 選定目前正在研究的 application；尚未突破時不得提前解鎖。
// 若讀到舊存檔的已完成 PendingChoice，才走突破後改選相容路徑。
func (s *GameSession) ChooseResearchTech(tech gamedata.Technology) bool {
	s.recordPlayerCommand(PlayerCommand{Name: CmdChooseResearch, Args: []int{int(tech)}})
	legacyCompletedChoice := s.Player.HasPendingChoice && s.Player.CompletedTopics[s.Player.PendingChoice]
	var (
		ps engine.PlayerState
		ok bool
	)
	if s.Player.HasPendingChoice && !s.Player.CompletedTopics[s.Player.PendingChoice] {
		ps, ok = engine.SelectResearchApplication(s.Player, s.Player.PendingChoice, tech)
	} else {
		ps, ok = engine.ApplyResearchChoice(s.Player, tech)
	}
	if ok {
		s.Player = ps
		if legacyCompletedChoice {
			s.UpdatePlayerShipDesignsAfterTech()
		}
	}
	return ok
}

// ChosenTechFor 回傳某已完成主題實際選定解鎖的科技(未完成/未記錄回 false)。
func (s *GameSession) ChosenTechFor(topic gamedata.ResearchTopic) (gamedata.Technology, bool) {
	if s.Player.ChosenTech == nil {
		return 0, false
	}
	t, ok := s.Player.ChosenTech[topic]
	return t, ok
}

// SetResearchTopic 切換玩家目前研究主題;若切到不同主題則歸零進度(換題重來)。
func (s *GameSession) SetResearchTopic(t gamedata.ResearchTopic) {
	s.recordPlayerCommand(PlayerCommand{Name: CmdSetResearch, Args: []int{int(t)}})
	if !gamedata.IsResearchableTopic(t) {
		return
	}
	if s.Player.ResearchTopic == t {
		s.preparePlayerResearchApplication()
		return
	}
	s.Player.ResearchTopic = t
	s.Player.ResearchProgress = 0
	s.Player.ResearchApplication = 0
	s.Player.HasResearchApplication = false
	s.Player.PendingChoice = 0
	s.Player.HasPendingChoice = false
	s.preparePlayerResearchApplication()
}

// EnsurePlayerResearchApplication 補齊新局／舊存檔可能尚未建立的目前 application。
// 回傳 true 代表一般玩家必須先經 UI 選擇，呼叫端不可先推進世界回合。
func (s *GameSession) EnsurePlayerResearchApplication() bool {
	s.preparePlayerResearchApplication()
	return s.Player.HasPendingChoice && !s.Player.CompletedTopics[s.Player.PendingChoice]
}

func (s *GameSession) preparePlayerResearchApplication() {
	ps := &s.Player
	choice := gamedata.ResearchChoiceFor(ps.ResearchTopic)
	if ps.ResearchTopic == 0 || ps.CompletedTopics[ps.ResearchTopic] || len(choice.Choices) == 0 ||
		choice.ResearchAll || s.RaceCreative() {
		ps.ResearchApplication = 0
		ps.HasResearchApplication = false
		ps.PendingChoice = 0
		ps.HasPendingChoice = false
		return
	}
	if ps.HasResearchApplication {
		return
	}
	if len(choice.Choices) == 1 {
		next, ok := engine.SelectResearchApplication(*ps, ps.ResearchTopic, choice.Choices[0])
		if ok {
			*ps = next
		}
		return
	}
	if s.RaceUncreative() {
		idx := s.researchRandForTurn().Intn(len(choice.Choices))
		next, ok := engine.SelectResearchApplication(*ps, ps.ResearchTopic, choice.Choices[idx])
		if ok {
			*ps = next
		}
		return
	}
	ps.PendingChoice = ps.ResearchTopic
	ps.HasPendingChoice = true
}
