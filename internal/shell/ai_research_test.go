package shell

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestAIOriginalResearchSelectionStoresFieldAndApplicationTogether(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	if !a.OriginalTechProfileKnown {
		t.Fatal("新局 AI 應保存 sub_589D6 raw profile")
	}
	a.Player.CompletedTopics[a.Player.ResearchTopic] = true
	if !s.selectOriginalAIResearch(0, 12) {
		t.Fatal("有可研究 field 時應能完成原版 application 級抽選")
	}
	if a.Player.CompletedTopics[a.Player.ResearchTopic] {
		t.Fatalf("選到已完成 field %v", a.Player.ResearchTopic)
	}
	if !gamedata.ResearchTopicGrantsAll(a.Player.ResearchTopic) &&
		!aiRaceHasTrait(*a, gamedata.TRAIT_CREATIVE) && !a.Player.HasResearchApplication {
		t.Fatalf("非全授予 field %v 應在同一次決策保存 application", a.Player.ResearchTopic)
	}
}

func TestAIOriginalResearchSelectionIsDeterministic(t *testing.T) {
	build := func() *GameSession {
		s := NewDemoSession()
		a := &s.AIPlayers[0]
		a.Player.CompletedTopics[a.Player.ResearchTopic] = true
		s.advanceAIResearch(0, 12)
		return s
	}
	a, b := build(), build()
	pa, pb := a.AIPlayers[0].Player, b.AIPlayers[0].Player
	if pa.ResearchTopic != pb.ResearchTopic || pa.ResearchApplication != pb.ResearchApplication ||
		pa.HasResearchApplication != pb.HasResearchApplication {
		t.Fatalf("同 seed/profile/研究產出應選同一 field/application：%+v / %+v", pa, pb)
	}
}

func TestAIOriginalTechProfileSurvivesSaveLoad(t *testing.T) {
	s := NewDemoSession()
	before := s.AIPlayers[0].OriginalTechProfile
	path := filepath.Join(t.TempDir(), "ai-profile.json")
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSession(path)
	if err != nil {
		t.Fatal(err)
	}
	if !got.AIPlayers[0].OriginalTechProfileKnown || got.AIPlayers[0].OriginalTechProfile != before {
		t.Fatalf("raw profile 未完整往返：before=%+v after=%+v known=%v",
			before, got.AIPlayers[0].OriginalTechProfile, got.AIPlayers[0].OriginalTechProfileKnown)
	}
}

// 目前主題完成之後,AI 要換下一個——這是這一項修的那個洞。
//
// 先前沒有任何地方替 AI 選主題,所以它每回合把研究點投進一個早就完成的主題,
// 無限重複完成同一項。
func TestAIPicksANewTopicOnceTheCurrentOneIsDone(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	before := a.Player.ResearchTopic
	a.Player.CompletedTopics[before] = true

	s.advanceAIResearch(0)

	if a.Player.ResearchTopic == before {
		t.Fatalf("主題已完成卻沒換,還是 %v", before)
	}
	if a.Player.CompletedTopics[a.Player.ResearchTopic] {
		t.Errorf("換到的主題 %v 已經完成過了", a.Player.ResearchTopic)
	}
}

// 正對照:**還在研究中**就不該換主題。
//
// 少了這條,「每回合都重挑一次」的實作也會讓上面那條通過——而那會讓 AI 的研究進度
// 每次換主題就作廢一次,永遠研究不完任何東西。
func TestAIKeepsResearchingAnUnfinishedTopic(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	before := a.Player.ResearchTopic
	delete(a.Player.CompletedTopics, before)

	for i := 0; i < 5; i++ {
		s.advanceAIResearch(0)
	}
	if a.Player.ResearchTopic != before {
		t.Errorf("主題還沒完成不該換:%v → %v", before, a.Player.ResearchTopic)
	}
}

// 偷來的主題也算數。
//
// 間諜是直接寫 `CompletedTopics` 的(spy.go),不經過研究階段——所以那一回合的
// `ResearchDone` 是 false。只在「本回合有研究完成」時才重挑的實作,會讓 AI 卡在一個
// 偷來的主題上繼續投點。
func TestAIRepicksAfterStealingItsCurrentTopic(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	before := a.Player.ResearchTopic
	// 模擬間諜:直接標完成,進度完全沒動。
	a.Player.CompletedTopics[before] = true
	a.Player.ResearchProgress = 0

	s.advanceAIResearch(0)
	if a.Player.ResearchTopic == before {
		t.Errorf("偷來的主題也該觸發重挑,還是 %v", before)
	}
}

// 多選主題完成後,AI 要**明確**選一項。
//
// 先前沒人替 AI 選,`HasPendingChoice` 永遠 true、`ExplicitChoice` 永遠空——
// 而 `groundEquipTechOwned` 那組判定是「主題完成 + 沒有明確抉擇 → 視為擁有」,
// 所以 AI 每完成一個三選一主題實際上**三個都拿到**。
func TestAIResolvesItsPendingResearchChoice(t *testing.T) {
	topic := gamedata.TOPIC_ADVANCED_BIOLOGY // 三選一:複製中心 / 死亡孢子 / 土壤改良
	choice := gamedata.ResearchChoiceFor(topic)
	if len(choice.Choices) < 2 {
		t.Fatalf("測試前提不成立:%v 應該是多選主題,得到 %d 個選項", topic, len(choice.Choices))
	}

	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Player.PendingChoice = topic
	a.Player.HasPendingChoice = true
	a.Player.CompletedTopics[topic] = true

	s.advanceAIResearch(0)

	if a.Player.HasPendingChoice {
		t.Error("待決抉擇應已處理")
	}
	if !a.Player.ExplicitChoice[topic] {
		t.Error("應標成明確抉擇——否則 AI 等於三個選項全拿")
	}
	got := a.Player.ChosenTech[topic]
	found := false
	for _, c := range choice.Choices {
		if c == got {
			found = true
		}
	}
	if !found {
		t.Errorf("選到的 %v 不在該主題的選項裡", got)
	}
}

// 挑哪一項用的是**原版的 category 倍率**(`Calc_Tech_Value_` 階段 B 的 `ecx` 初始值),
// 不是「清單第一項」。
func TestAIApplicationPickFollowsTheOriginalCategoryWeight(t *testing.T) {
	choice := gamedata.ResearchChoiceFor(gamedata.TOPIC_ADVANCED_BIOLOGY)
	got, ok := aiPickApplication(choice.Choices)
	if !ok {
		t.Fatal("應該挑得出一項")
	}
	want, wantW := gamedata.Technology(0), -1
	for _, c := range choice.Choices {
		if w := gamedata.TechCategoryWeight(c); w > wantW {
			want, wantW = c, w
		}
	}
	if got != want {
		t.Errorf("應挑 category 倍率最高的 %s(%d),得到 %s(%d)",
			gamedata.TechnologyName(want), wantW,
			gamedata.TechnologyName(got), gamedata.TechCategoryWeight(got))
	}
	// 同分取先出現的那個——不擲骰,整條 AI 研究線是決定性的。
	a, b := gamedata.Technology(0), gamedata.Technology(0)
	for _, c := range choice.Choices {
		if gamedata.TechCategoryWeight(c) == wantW {
			if a == 0 {
				a = c
			} else if b == 0 {
				b = c
			}
		}
	}
	if b != 0 && got != a {
		t.Errorf("同分應取先出現的 %v,得到 %v", a, got)
	}
}

// 候選清單每個研究領域只給隊首那一個(原版規則),而且領域索引要對得回去
// ——`ai.DecideResearchTopic` 靠它判斷「這是不是軍事領域」。
func TestAIResearchCandidatesAreOnePerAreaWithTheRightIndex(t *testing.T) {
	s := NewDemoSession()
	ps := s.AIPlayers[0].Player
	cands := aiResearchCandidates(ps)
	if len(cands) == 0 {
		t.Fatal("開局應該有可研究的主題")
	}
	tree := gamedata.TechTree()
	seen := map[int]bool{}
	for _, c := range cands {
		if seen[c.AreaIndex] {
			t.Errorf("領域 %d 出現兩次", c.AreaIndex)
		}
		seen[c.AreaIndex] = true
		if c.AreaIndex < 0 || c.AreaIndex >= len(tree) {
			t.Fatalf("領域索引 %d 越界", c.AreaIndex)
		}
		// 必須是該領域**第一個未完成**的主題。
		var want gamedata.ResearchTopic = -1
		for _, topic := range tree[c.AreaIndex] {
			if !ps.CompletedTopics[topic] {
				want = topic
				break
			}
		}
		if gamedata.ResearchTopic(c.TopicID) != want {
			t.Errorf("領域 %d 的隊首應為 %v,候選給的是 %v", c.AreaIndex, want, gamedata.ResearchTopic(c.TopicID))
		}
		if c.Cost <= 0 {
			t.Errorf("主題 %v 的成本應 > 0,得到 %d", gamedata.ResearchTopic(c.TopicID), c.Cost)
		}
	}
}

// 一般科技全完成後，八個 terminal Hyper 仍可重複研究，且成本包含自己的 level byte。
func TestAIWithOrdinaryTreeCompleteChoosesRepeatedHyper(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	for _, area := range gamedata.TechTree() {
		for _, topic := range area {
			a.Player.CompletedTopics[topic] = true
		}
	}
	a.Player.HyperAdvancedResearchCost = 25000
	a.Player.HyperAdvancedLevels = map[gamedata.ResearchTopic]int{gamedata.TOPIC_HYPER_PHYSICS: 2}
	cands := aiResearchCandidates(a.Player)
	if len(cands) != 8 {
		t.Fatalf("一般科技完成後應保留八個 Hyper candidate，得到 %d", len(cands))
	}
	for _, c := range cands {
		topic := gamedata.ResearchTopic(c.TopicID)
		want := 25000
		if topic == gamedata.TOPIC_HYPER_PHYSICS {
			want = 45000
		}
		if c.Cost != want {
			t.Errorf("Hyper %v candidate 成本=%d, want %d", topic, c.Cost, want)
		}
	}
}

// 端到端:跑滿一局的長度,AI 的研究線真的有在前進。
//
// 這一條測的是接線(advanceAI 有沒有真的呼叫到),不是單一函式——先前那個洞正是
// 「函式寫好了但沒有呼叫端」。
func TestAIResearchActuallyProgressesOverAGame(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	start := s.AIPlayers[0].Player.ResearchTopic
	startDone := len(s.AIPlayers[0].Player.CompletedTopics)

	for turn := 0; turn < 200; turn++ {
		s.EndTurn()
	}
	a := s.AIPlayers[0].Player
	if a.ResearchTopic == start {
		t.Errorf("200 回合後 AI 的研究主題還停在 %v", start)
	}
	if len(a.CompletedTopics) <= startDone {
		t.Errorf("200 回合後完成的主題數沒有成長:%d → %d", startDone, len(a.CompletedTopics))
	}
	if a.HasPendingChoice {
		t.Error("回合結算完不該還留著待決的科技抉擇")
	}
	if len(a.ExplicitChoice) == 0 {
		t.Error("跑了 200 回合卻一次明確抉擇都沒有——多選主題的那條路徑沒被走到")
	}
}
