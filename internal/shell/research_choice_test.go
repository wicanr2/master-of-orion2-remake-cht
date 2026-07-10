package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestPendingResearchChoiceFlow:多選主題完成後,shell 能取得待決選項並改選。
func TestPendingResearchChoiceFlow(t *testing.T) {
	s := NewDemoSession()
	topic := gamedata.TOPIC_ADVANCED_CONSTRUCTION
	ch := gamedata.ResearchChoiceFor(topic)
	if len(ch.Choices) < 2 {
		t.Skip("前提:需多選主題")
	}
	s.SetResearchTopic(topic)
	// 直接灌滿成本觸發完成(走引擎研究階段)。
	s.Player.ResearchProgress = ch.Cost
	ps, done := runResearchForTest(s)
	if !done {
		t.Fatalf("應完成主題")
	}
	s.Player = ps
	gotTopic, choices, ok := s.PendingResearchChoice()
	if !ok || gotTopic != topic || len(choices) != len(ch.Choices) {
		t.Fatalf("應有待決抉擇 topic=%v choices=%d,得 ok=%v topic=%v", topic, len(ch.Choices), ok, gotTopic)
	}
	if !s.ChooseResearchTech(ch.Choices[1]) {
		t.Fatalf("改選第二項應成功")
	}
	if got, ok := s.ChosenTechFor(topic); !ok || got != ch.Choices[1] {
		t.Fatalf("選定科技應為 %v,得 %v(ok=%v)", ch.Choices[1], got, ok)
	}
	if _, _, ok := s.PendingResearchChoice(); ok {
		t.Fatalf("改選後不應再有待決")
	}
}
