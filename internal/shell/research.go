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
// 取該領域第一個尚未完成的主題(整條研究完的領域不出現)。
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
// profile——Hyper-Advanced Lv1 主題(gamedata.IsHyperAdvancedTopic)一律回套件級硬編值
// 25000(= 現行 Profile15 行為)。畫面需要顯示「這局(可能是 1.3)實際要花多少 RP」時改用
// (*GameSession).ResearchCostForDisplay。
func ResearchCost(t gamedata.ResearchTopic) int {
	return gamedata.ResearchChoiceFor(t).Cost
}

// ResearchCostForDisplay 同 ResearchCost,但套用這局遊戲的版本規則 profile(s.RuleProfile)——
// Hyper-Advanced Lv1 主題(8 個共用同一成本的 TOPIC_HYPER_*)改讀
// gamedata.HyperAdvancedCost(s.RuleProfile) 覆寫(1.3=15000/1.5=25000),其餘主題與 ResearchCost
// 相同。是 gamedata.HyperAdvancedCost 註解點名的「顯示層接線」:資料層(ruleprofile.go)已備妥,
// 這裡接進研究選單/帝國概況畫面,讓玩家在 1.3 局裡看到的成本真的是 1.3 值,不是永遠顯示 1.5
// 硬編值再讓實際結算(RunResearchPhase)用不同數字扣款的顯示/結算不一致。
func (s *GameSession) ResearchCostForDisplay(t gamedata.ResearchTopic) int {
	if gamedata.IsHyperAdvancedTopic(t) {
		return gamedata.HyperAdvancedCost(s.RuleProfile)
	}
	return gamedata.ResearchChoiceFor(t).Cost
}

// ResearchTopicName 回傳主題的英文顯示名(83 個 topic 全收錄,= gamedata.TopicEnglishName,
// 也就是 assets/i18n/tech.tsv 的 i18n key)。shell 層不 import i18n,由 cmd/moo2 顯示端
// 經 catalog 翻中文(見 cmd/moo2/topicname.go 的 topicNameZh),與其他畫面字串一致。
func ResearchTopicName(t gamedata.ResearchTopic) string {
	return gamedata.TopicEnglishName(t)
}

// PendingResearchChoice 回傳玩家「剛完成、可改選解鎖科技」的主題與其可選科技清單。
// ok=false 表示目前沒有待決抉擇。供研究抉擇 UI 使用(MOO2 每主題數科技間抉擇)。
func (s *GameSession) PendingResearchChoice() (topic gamedata.ResearchTopic, choices []gamedata.Technology, ok bool) {
	if !s.Player.HasPendingChoice {
		return 0, nil, false
	}
	t := s.Player.PendingChoice
	return t, gamedata.ResearchChoiceFor(t).Choices, true
}

// ChooseResearchTech 把目前待決主題改選為 tech(須為合法選項),回傳是否成功。
func (s *GameSession) ChooseResearchTech(tech gamedata.Technology) bool {
	ps, ok := engine.ApplyResearchChoice(s.Player, tech)
	if ok {
		s.Player = ps
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
	if s.Player.ResearchTopic == t {
		return
	}
	s.Player.ResearchTopic = t
	s.Player.ResearchProgress = 0
}
