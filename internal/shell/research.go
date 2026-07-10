package shell

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// research.go:可玩遊戲殼的研究主題選單資料(純邏輯)。
//
// 說明:原版每個研究主題的名稱存在遊戲資料 LBX(執行期才載入),demo 對局未載入 LBX,
// 因此這裡提供一份「精選早期主題 + 專案內建譯名 + 成本」的最小集,讓玩家能實際選擇研究方向。
// 待整合原版 LBX 科技名後,ResearchTopicName 可改接權威來源。成本一律以 gamedata 為準(不硬抄)。

// ResearchOption 是一個可選研究主題。
type ResearchOption struct {
	Topic gamedata.ResearchTopic
	Name  string
	Cost  int
}

// StarterResearchTopics 回傳新手可選的早期研究主題(依成本由低到高)。
// 成本取自 gamedata.ResearchChoiceFor(權威來源),名稱為專案內建譯名。
func StarterResearchTopics() []ResearchOption {
	topics := []struct {
		t    gamedata.ResearchTopic
		name string
	}{
		{gamedata.TOPIC_ADVANCED_ENGINEERING, "進階工程學"},
		{gamedata.TOPIC_ADVANCED_CONSTRUCTION, "進階建築學"},
		{gamedata.TOPIC_MILITARY_TACTICS, "軍事戰術"},
		{gamedata.TOPIC_ADVANCED_FUSION, "進階核融合"},
		{gamedata.TOPIC_ADVANCED_MAGNETISM, "進階磁學"},
		{gamedata.TOPIC_ADVANCED_METALLURGY, "進階冶金學"},
		{gamedata.TOPIC_ADVANCED_BIOLOGY, "進階生物學"},
		{gamedata.TOPIC_ARTIFICIAL_INTELLIGENCE, "人工智慧"},
		{gamedata.TOPIC_ADVANCED_CHEMISTRY, "進階化學"},
	}
	out := make([]ResearchOption, 0, len(topics))
	for _, x := range topics {
		out = append(out, ResearchOption{Topic: x.t, Name: x.name, Cost: ResearchCost(x.t)})
	}
	return out
}

// ResearchCost 回傳主題完成所需研究點(RP),取自 gamedata。
func ResearchCost(t gamedata.ResearchTopic) int {
	return gamedata.ResearchChoiceFor(t).Cost
}

// ResearchTopicName 回傳主題的顯示名;未收錄者回後備字串。
func ResearchTopicName(t gamedata.ResearchTopic) string {
	for _, o := range StarterResearchTopics() {
		if o.Topic == t {
			return o.Name
		}
	}
	if t == gamedata.TOPIC_STARTING_TECH {
		return "起始科技"
	}
	return fmt.Sprintf("研究主題 #%d", int(t))
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
