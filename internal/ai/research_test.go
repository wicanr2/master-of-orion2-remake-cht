package ai

import "testing"

// TestDecideResearchTopicAggressivePrefersMilitary 驗證好戰型在有軍事領域候選時,
// 優先選軍事領域中成本最低者(即使全體候選中有更便宜的非軍事主題)。
func TestDecideResearchTopicAggressivePrefersMilitary(t *testing.T) {
	candidates := []ResearchCandidate{
		{TopicID: 1, Cost: 20, AreaIndex: 0}, // Biology,非軍事,最便宜但不該被選
		{TopicID: 2, Cost: 80, AreaIndex: 2}, // Physics,軍事
		{TopicID: 3, Cost: 60, AreaIndex: 3}, // Construction,軍事,軍事領域中最便宜
	}
	got := DecideResearchTopic(candidates, ProfileAggressive)
	if got != 3 {
		t.Errorf("好戰型應選軍事領域最便宜主題 TopicID=3,得到 %d", got)
	}
}

// TestDecideResearchTopicAggressiveFallsBackWithoutMilitary 驗證好戰型在候選中完全沒有
// 軍事領域主題時,退回選全體候選中成本最低者。
func TestDecideResearchTopicAggressiveFallsBackWithoutMilitary(t *testing.T) {
	candidates := []ResearchCandidate{
		{TopicID: 1, Cost: 50, AreaIndex: 0}, // Biology
		{TopicID: 2, Cost: 30, AreaIndex: 7}, // Sociology
		{TopicID: 3, Cost: 90, AreaIndex: 6}, // Computers
	}
	got := DecideResearchTopic(candidates, ProfileAggressive)
	if got != 2 {
		t.Errorf("無軍事候選時應退回全體最便宜 TopicID=2,得到 %d", got)
	}
}

// TestDecideResearchTopicExpansionistPrefersMilitary 驗證擴張型(IndustryWeight>ResearchWeight,
// 與好戰型同一分支)也套用「軍事領域優先、選最便宜」策略。
func TestDecideResearchTopicExpansionistPrefersMilitary(t *testing.T) {
	candidates := []ResearchCandidate{
		{TopicID: 1, Cost: 40, AreaIndex: 1}, // Power,軍事,最便宜
		{TopicID: 2, Cost: 45, AreaIndex: 5}, // Chemistry,軍事
		{TopicID: 3, Cost: 10, AreaIndex: 0}, // Biology,非軍事,更便宜但不該被選
	}
	got := DecideResearchTopic(candidates, ProfileExpansionist)
	if got != 1 {
		t.Errorf("擴張型應選軍事領域最便宜主題 TopicID=1,得到 %d", got)
	}
}

// TestDecideResearchTopicScientificPicksCostliest 驗證科學型選成本最高者(長線投資)。
func TestDecideResearchTopicScientificPicksCostliest(t *testing.T) {
	candidates := []ResearchCandidate{
		{TopicID: 1, Cost: 50, AreaIndex: 4},
		{TopicID: 2, Cost: 120, AreaIndex: 0}, // 最貴
		{TopicID: 3, Cost: 90, AreaIndex: 2},
	}
	got := DecideResearchTopic(candidates, ProfileScientific)
	if got != 2 {
		t.Errorf("科學型應選成本最高 TopicID=2,得到 %d", got)
	}
}

// TestDecideResearchTopicBalancedPicksCheapest 驗證平衡型選成本最低者(穩定推進)。
func TestDecideResearchTopicBalancedPicksCheapest(t *testing.T) {
	candidates := []ResearchCandidate{
		{TopicID: 1, Cost: 70, AreaIndex: 3},
		{TopicID: 2, Cost: 40, AreaIndex: 0}, // 最便宜
		{TopicID: 3, Cost: 55, AreaIndex: 6},
	}
	got := DecideResearchTopic(candidates, ProfileBalanced)
	if got != 2 {
		t.Errorf("平衡型應選成本最低 TopicID=2,得到 %d", got)
	}
}

// TestDecideResearchTopicTieBreakKeepsFirst 驗證成本並列時保留候選清單中先出現者。
func TestDecideResearchTopicTieBreakKeepsFirst(t *testing.T) {
	candidates := []ResearchCandidate{
		{TopicID: 1, Cost: 50, AreaIndex: 0},
		{TopicID: 2, Cost: 50, AreaIndex: 7}, // 同成本,較晚出現,不該被選
	}
	got := DecideResearchTopic(candidates, ProfileBalanced)
	if got != 1 {
		t.Errorf("同成本應保留先出現者 TopicID=1,得到 %d", got)
	}
}

// TestDecideResearchTopicEmptyReturnsNegativeOne 驗證無候選時任何 Profile 都回傳 -1。
func TestDecideResearchTopicEmptyReturnsNegativeOne(t *testing.T) {
	for _, p := range []Profile{ProfileAggressive, ProfileScientific, ProfileBalanced, ProfileExpansionist} {
		if got := DecideResearchTopic(nil, p); got != -1 {
			t.Errorf("%s: 無候選應回傳 -1,得到 %d", p.Name, got)
		}
	}
}
