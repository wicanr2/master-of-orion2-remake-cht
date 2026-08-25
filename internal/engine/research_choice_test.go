package engine

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestResearchChoiceMultiOption 保留舊存檔沒有預選 application 時的相容分支。
func TestResearchChoiceMultiOption(t *testing.T) {
	// TOPIC_ADVANCED_CONSTRUCTION 成本 150,3 選項(工廠/重甲/行星飛彈基地)。
	topic := gamedata.TOPIC_ADVANCED_CONSTRUCTION
	ch := gamedata.ResearchChoiceFor(topic)
	if len(ch.Choices) < 2 {
		t.Fatalf("測試前提失效:%v 應為多選主題", topic)
	}
	ps := PlayerState{ResearchTopic: topic}
	ps, done := RunResearchPhase(ps, ch.Cost+1) // 超過成本且固定 roll=1，保證突破
	if !done {
		t.Fatalf("灌滿成本應完成主題")
	}
	if !ps.HasPendingChoice || ps.PendingChoice != topic {
		t.Fatalf("多選主題完成應開啟 PendingChoice=%v,得 pending=%v topic=%v", topic, ps.HasPendingChoice, ps.PendingChoice)
	}
	if ps.ChosenTech[topic] != ch.Choices[0] {
		t.Fatalf("預設應選第一項 %v,得 %v", ch.Choices[0], ps.ChosenTech[topic])
	}
	// 改選第二項(合法)。
	ps2, ok := ApplyResearchChoice(ps, ch.Choices[1])
	if !ok {
		t.Fatalf("改選合法項應成功")
	}
	if ps2.ChosenTech[topic] != ch.Choices[1] || ps2.HasPendingChoice {
		t.Fatalf("改選後應記第二項且清除待決:tech=%v pending=%v", ps2.ChosenTech[topic], ps2.HasPendingChoice)
	}
	// 改選非法項應失敗。
	if _, ok := ApplyResearchChoice(ps, gamedata.TECH_DEATH_RAY); ok {
		t.Fatalf("非該主題的科技不應可選")
	}
}

func TestResearchApplicationSelectedBeforeCompletion(t *testing.T) {
	topic := gamedata.TOPIC_ADVANCED_CONSTRUCTION
	choice := gamedata.ResearchChoiceFor(topic)
	ps := PlayerState{ResearchTopic: topic}
	selected, ok := SelectResearchApplication(ps, topic, choice.Choices[1])
	if !ok {
		t.Fatal("研究開始前選擇合法 application 應成功")
	}
	if selected.ChosenTech != nil || selected.CompletedTopics[topic] {
		t.Fatal("尚未突破時不得提前解鎖所選 application")
	}
	completed, done := RunResearchPhase(selected, choice.Cost+1)
	if !done {
		t.Fatal("超過成本且 roll=1 應完成")
	}
	if got := completed.ChosenTech[topic]; got != choice.Choices[1] {
		t.Fatalf("突破應授予研究前選定項：got=%v want=%v", got, choice.Choices[1])
	}
	if completed.HasPendingChoice || completed.HasResearchApplication {
		t.Fatal("授予後不得留下目前 application 或待決旗標")
	}
}

// TestResearchChoiceResearchAll:ResearchAll 主題完成不開待決(全解語意)。
func TestResearchChoiceResearchAll(t *testing.T) {
	// TOPIC_CHEMISTRY 成本 50,ResearchAll=true。
	topic := gamedata.TOPIC_CHEMISTRY
	ch := gamedata.ResearchChoiceFor(topic)
	if !ch.ResearchAll {
		t.Fatalf("測試前提失效:%v 應 ResearchAll", topic)
	}
	ps := PlayerState{ResearchTopic: topic}
	ps, done := RunResearchPhase(ps, ch.Cost+1)
	if !done {
		t.Fatalf("應完成")
	}
	if ps.HasPendingChoice {
		t.Fatalf("ResearchAll 主題不應開待決選擇")
	}
}
