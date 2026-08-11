package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func completedResearchState(t *testing.T, topic gamedata.ResearchTopic) engine.PlayerState {
	t.Helper()
	rc := gamedata.ResearchChoiceFor(topic)
	ps := engine.PlayerState{ResearchTopic: topic}
	got, done := engine.RunResearchPhase(ps, rc.Cost)
	if !done {
		t.Fatalf("研究主題 %v 應完成", topic)
	}
	return got
}

func TestCreativeResearchUnlocksEveryApplicationWithoutPendingChoice(t *testing.T) {
	topic := gamedata.TOPIC_ADVANCED_BIOLOGY
	rc := gamedata.ResearchChoiceFor(topic)
	if len(rc.Choices) < 2 {
		t.Fatalf("測試前提不成立:應為多選主題,得到 %d 項", len(rc.Choices))
	}

	ps := completedResearchState(t, topic)
	applyResearchRaceTrait(&ps, topic, true, false, nil)

	if ps.HasPendingChoice || ps.PendingChoice != 0 {
		t.Fatalf("Creative 不應留下待決選項:pending=%v topic=%v", ps.HasPendingChoice, ps.PendingChoice)
	}
	if ps.ExplicitChoice != nil && ps.ExplicitChoice[topic] {
		t.Fatal("Creative 不應把領域標成明確擇一")
	}
	for _, tech := range rc.Choices {
		if !psKnowsTech(ps, topic, tech) {
			t.Errorf("Creative 應知道同領域全部科技,但缺少 %v", tech)
		}
	}
}

func TestUncreativeResearchRandomlyChoosesOneApplication(t *testing.T) {
	topic := gamedata.TOPIC_ADVANCED_BIOLOGY
	rc := gamedata.ResearchChoiceFor(topic)
	if len(rc.Choices) < 2 {
		t.Fatalf("測試前提不成立:應為多選主題,得到 %d 項", len(rc.Choices))
	}

	ps := completedResearchState(t, topic)
	applyResearchRaceTrait(&ps, topic, false, true, func(n int) int { return n - 1 })

	if ps.HasPendingChoice || ps.PendingChoice != 0 {
		t.Fatalf("Uncreative 自動選擇後不應留下待決:pending=%v topic=%v", ps.HasPendingChoice, ps.PendingChoice)
	}
	if !ps.ExplicitChoice[topic] {
		t.Fatal("Uncreative 應標成明確擇一")
	}
	if got := ps.ChosenTech[topic]; got != rc.Choices[len(rc.Choices)-1] {
		t.Fatalf("固定選擇器應選最後一項:got=%v want=%v", got, rc.Choices[len(rc.Choices)-1])
	}
	if !psKnowsTech(ps, topic, rc.Choices[len(rc.Choices)-1]) {
		t.Fatal("Uncreative 應知道選中的科技")
	}
	if psKnowsTech(ps, topic, rc.Choices[0]) {
		t.Fatal("Uncreative 不應知道未選中的科技")
	}
}

func TestResearchTraitFlagsReachBuiltInAndCustomRaces(t *testing.T) {
	s := NewDemoSession()
	s.ApplyRace(1) // Psilons
	if !s.RaceCreative() || s.RaceUncreative() {
		t.Fatal("Psilons 應只有 Creative")
	}
	s.ApplyCustomRaceBonuses(Race{}, gamedata.TRAIT_UNCREATIVE)
	if !s.RaceUncreative() || s.RaceCreative() {
		t.Fatal("客製 Uncreative 應與 Creative 互斥")
	}
}

func TestEndTurnAppliesResearchTraitRules(t *testing.T) {
	topic := gamedata.TOPIC_ADVANCED_BIOLOGY
	rc := gamedata.ResearchChoiceFor(topic)

	creative := NewDemoSession()
	creative.ApplyCustomRaceBonuses(Race{}, gamedata.TRAIT_CREATIVE)
	creative.Player.ResearchTopic = topic
	creative.Player.ResearchProgress = rc.Cost
	creative.EndTurn()
	if creative.Player.HasPendingChoice {
		t.Fatal("Creative 經 EndTurn 後不應進入一般擇一畫面")
	}
	for _, tech := range rc.Choices {
		if !psKnowsTech(creative.Player, topic, tech) {
			t.Errorf("Creative EndTurn 後應知道 %v", tech)
		}
	}

	uncreative := NewDemoSession()
	uncreative.ApplyCustomRaceBonuses(Race{}, gamedata.TRAIT_UNCREATIVE)
	uncreative.Player.ResearchTopic = topic
	uncreative.Player.ResearchProgress = rc.Cost
	uncreative.EndTurn()
	if uncreative.Player.HasPendingChoice || !uncreative.Player.ExplicitChoice[topic] {
		t.Fatal("Uncreative 經 EndTurn 後應自動完成且標記明確擇一")
	}
	if uncreative.researchRand == nil || uncreative.researchRand.Draws() == 0 {
		t.Fatal("Uncreative 應使用研究專用亂數流")
	}
}
