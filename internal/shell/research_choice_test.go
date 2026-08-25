package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestPendingResearchChoiceFlow:多選主題在開始研究前選定 application，突破時才解鎖。
func TestPendingResearchChoiceFlow(t *testing.T) {
	s := NewDemoSession()
	topic := gamedata.TOPIC_ADVANCED_CONSTRUCTION
	ch := gamedata.ResearchChoiceFor(topic)
	if len(ch.Choices) < 2 {
		t.Skip("前提:需多選主題")
	}
	s.SetResearchTopic(topic)
	gotTopic, choices, ok := s.PendingResearchChoice()
	if !ok || gotTopic != topic || len(choices) != len(ch.Choices) {
		t.Fatalf("開始研究前應有待決 application topic=%v choices=%d,得 ok=%v topic=%v", topic, len(ch.Choices), ok, gotTopic)
	}
	if !s.ChooseResearchTech(ch.Choices[1]) {
		t.Fatalf("研究前選第二項應成功")
	}
	if _, known := s.ChosenTechFor(topic); known {
		t.Fatal("研究尚未突破，不得提前寫入 ChosenTech")
	}
	// 進度到成本，再投入 1 RP 形成最低突破率；測試 helper 固定 roll=1。
	s.Player.ResearchProgress = ch.Cost
	ps, done := runResearchForTest(s)
	if !done {
		t.Fatalf("應完成主題")
	}
	s.Player = ps
	if got, ok := s.ChosenTechFor(topic); !ok || got != ch.Choices[1] {
		t.Fatalf("選定科技應為 %v,得 %v(ok=%v)", ch.Choices[1], got, ok)
	}
	if _, _, ok := s.PendingResearchChoice(); ok {
		t.Fatalf("改選後不應再有待決")
	}
}

func TestUncreativeSelectsApplicationBeforeResearch(t *testing.T) {
	s := NewDemoSession()
	s.ApplyCustomRaceBonuses(Race{}, gamedata.TRAIT_UNCREATIVE)
	topic := gamedata.TOPIC_ADVANCED_BIOLOGY
	s.SetResearchTopic(topic)
	if s.Player.HasPendingChoice || !s.Player.HasResearchApplication {
		t.Fatal("Uncreative 應在研究開始時自動限縮為一個 application")
	}
	if s.Player.ChosenTech != nil && s.Player.ChosenTech[topic] != 0 {
		t.Fatal("Uncreative 的預選不得提前解鎖")
	}
	if s.researchRand == nil || s.researchRand.Draws() == 0 {
		t.Fatal("Uncreative 預選應使用可存檔研究亂數流")
	}
}

func TestChangingResearchTopicClearsPreviousApplication(t *testing.T) {
	s := NewDemoSession()
	first := gamedata.TOPIC_ADVANCED_CONSTRUCTION
	s.SetResearchTopic(first)
	firstChoice := gamedata.ResearchChoiceFor(first)
	if !s.ChooseResearchTech(firstChoice.Choices[1]) {
		t.Fatal("第一個 topic 應可預選 application")
	}
	old := s.Player.ResearchApplication
	second := gamedata.TOPIC_ADVANCED_BIOLOGY
	s.SetResearchTopic(second)
	if s.Player.HasResearchApplication && s.Player.ResearchApplication == old {
		t.Fatal("切換 topic 不得沿用上一個 topic 的 application")
	}
	if !s.Player.HasPendingChoice || s.Player.PendingChoice != second {
		t.Fatal("新多選 topic 應建立自己的待選 application")
	}
}
