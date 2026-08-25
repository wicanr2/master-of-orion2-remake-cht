package shell

import (
	"strings"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestSecretExperimentCompletesPlayerResearchInsteadOfAddingRP(t *testing.T) {
	s := NewDemoSession()
	topic := gamedata.TOPIC_ADVANCED_BIOLOGY
	choice := gamedata.ResearchChoiceFor(topic)
	ps, ok := engine.SelectResearchApplication(engine.PlayerState{ResearchTopic: topic, ResearchProgress: 23}, topic, choice.Choices[1])
	if !ok {
		t.Fatal("測試前提：應可預選 application")
	}
	s.Player = ps
	result, ok := s.applyRandomEventLocalized(*gamedata.RandomEventByID(18))
	if !ok || !strings.Contains(result.MessageEN, ResearchTopicName(topic)) {
		t.Fatalf("事件應以完成前 field 顯示雙語結果：result=%+v ok=%v", result, ok)
	}
	if !s.Player.CompletedTopics[topic] || s.Player.ChosenTech[topic] != choice.Choices[1] {
		t.Fatalf("事件應立即授予已選 application：%+v", s.Player)
	}
	if s.Player.ResearchProgress != 0 || s.Player.ResearchTopic != 0 {
		t.Fatalf("事件不得沿用舊 RP 加值公式：topic=%v RP=%d", s.Player.ResearchTopic, s.Player.ResearchProgress)
	}
}

func TestSecretExperimentCreativeAndAITarget(t *testing.T) {
	creative := NewDemoSession()
	creative.ApplyCustomRaceBonuses(Race{}, gamedata.TRAIT_CREATIVE)
	topic := gamedata.TOPIC_ADVANCED_BIOLOGY
	creative.Player.ResearchTopic = topic
	if _, ok := creative.applyRandomEventLocalized(*gamedata.RandomEventByID(18)); !ok {
		t.Fatal("Creative 玩家秘密實驗應成功")
	}
	for _, tech := range gamedata.ResearchChoicesForTopic(topic) {
		if !playerStateKnowsTech(creative.Player, topic, tech) {
			t.Fatalf("Creative 應取得 field 全部 application，缺少 %v", tech)
		}
	}

	s := NewDemoSession()
	aiTopic := gamedata.TOPIC_ADVANCED_CONSTRUCTION
	choice := gamedata.ResearchChoiceFor(aiTopic)
	selected, ok := engine.SelectResearchApplication(engine.PlayerState{ResearchTopic: aiTopic, ResearchProgress: 31}, aiTopic, choice.Choices[0])
	if !ok {
		t.Fatal("測試前提：AI 應可預選 application")
	}
	s.AIPlayers[0].Player = selected
	playerBefore := s.Player.ResearchProgress
	if _, ok := s.applyRandomEventLocalizedToTarget(*gamedata.RandomEventByID(18), eventEmpireTarget{kind: eventEmpireAI, index: 0}); !ok {
		t.Fatal("AI 秘密實驗應成功")
	}
	if !s.AIPlayers[0].Player.CompletedTopics[aiTopic] || s.AIPlayers[0].Player.ResearchTopic != 0 ||
		s.AIPlayers[0].Player.ResearchProgress != 0 || s.Player.ResearchProgress != playerBefore {
		t.Fatalf("AI 結果不得誤寫玩家或保留 field/RP：ai=%+v playerRP=%d", s.AIPlayers[0].Player, s.Player.ResearchProgress)
	}
}

func TestSecretExperimentWritesBackInactiveHotseat(t *testing.T) {
	s := NewDemoSession()
	if s.SetupHotseatWithAIIndices([]int{0}) != 2 {
		t.Fatal("測試需要兩個熱座席位")
	}
	topic := gamedata.TOPIC_ADVANCED_BIOLOGY
	s.Seats[1].Player.ResearchTopic = topic
	s.Seats[1].Player.ResearchProgress = 44
	activeTopic := s.Player.ResearchTopic
	if _, ok := s.applyRandomEventLocalizedToTarget(*gamedata.RandomEventByID(18), eventEmpireTarget{kind: eventEmpireSeat, index: 1}); !ok {
		t.Fatal("非目前席位秘密實驗應成功")
	}
	if !s.Seats[1].Player.CompletedTopics[topic] || s.Seats[1].Player.ResearchTopic != 0 || s.Seats[1].Player.ResearchProgress != 0 {
		t.Fatalf("結果未回寫非目前席位：%+v", s.Seats[1].Player)
	}
	if s.ActiveSeat != 0 || s.Player.ResearchTopic != activeTopic {
		t.Fatalf("事件後應恢復原席位：seat=%d topic=%v", s.ActiveSeat, s.Player.ResearchTopic)
	}
}
