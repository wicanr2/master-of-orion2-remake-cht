package gamedata

import "testing"

// 偏好便宜的:分數與主題成本成反比(原版 score = weight × horizon ÷ turns)。
func TestStartingRandomScoresFavourCheapTopics(t *testing.T) {
	// TOPIC_ENGINEERING(29)=50、TOPIC_ADVANCED_GOVERNMENTS(6)=4500。
	avail := []ResearchTopic{TOPIC_ENGINEERING, TOPIC_ADVANCED_GOVERNMENTS}
	sc := StartingRandomTopicScores(avail, 10)
	if len(sc) != 2 {
		t.Fatalf("應回兩個分數,得到 %d", len(sc))
	}
	if sc[0] <= sc[1] {
		t.Errorf("成本 %d 的主題應比成本 %d 的分數高:%d vs %d",
			OrigTopicCost[29], OrigTopicCost[6], sc[0], sc[1])
	}
}

// 視野會放寬:全部都太貴時不該回全 0,而是把視野撐大到至少一個候選進得來。
func TestStartingRandomHorizonExpandsUntilSomethingQualifies(t *testing.T) {
	// 每回合 1 點研究、最貴的主題 → turns 遠超過初始視野 15。
	avail := []ResearchTopic{TOPIC_ADVANCED_GOVERNMENTS} // 4500
	sc := StartingRandomTopicScores(avail, 1)
	if len(sc) != 1 || sc[0] <= 0 {
		t.Fatalf("視野應放寬到這個候選進得來,得到 %v", sc)
	}
	// 視野真的有變大(初始 15,4500 回合需要放寬很多輪)。
	if h := startingRandomHorizonGrow(StartingRandomHorizonInitial); h != 22 {
		t.Errorf("15 放寬一輪應是 22(×3÷2),得到 %d", h)
	}
	// 防呆:視野必須嚴格遞增,否則外層迴圈不會結束。
	for _, h := range []int{0, 1, 2, 15, 1000} {
		if startingRandomHorizonGrow(h) <= h {
			t.Errorf("視野 %d 放寬後沒有變大", h)
		}
	}
}

// 加權隨機:骰值落在誰的區間就挑誰,而且是決定性的。
func TestStartingRandomTopicPickIsWeightedAndDeterministic(t *testing.T) {
	avail := []ResearchTopic{TOPIC_ENGINEERING, TOPIC_ADVANCED_GOVERNMENTS}
	sc := StartingRandomTopicScores(avail, 10)
	// 骰 0 → 第一個。
	got, ok := StartingRandomTopicPick(avail, 10, func(int) int { return 0 })
	if !ok || got != TOPIC_ENGINEERING {
		t.Errorf("骰 0 應挑第一個,得到 %d(ok=%v)", got, ok)
	}
	// 骰到第一個的分數 → 第二個。
	got, ok = StartingRandomTopicPick(avail, 10, func(int) int { return sc[0] })
	if !ok || got != TOPIC_ADVANCED_GOVERNMENTS {
		t.Errorf("骰過第一個的區間應挑第二個,得到 %d(ok=%v)", got, ok)
	}
	// 同樣的骰法回同樣的結果。
	a, _ := StartingRandomTopicPick(avail, 10, func(int) int { return 3 })
	b, _ := StartingRandomTopicPick(avail, 10, func(int) int { return 3 })
	if a != b {
		t.Errorf("同樣的輸入應回同樣的結果:%d vs %d", a, b)
	}
}

// 沒有候選時回 false,不是硬挑一個。
func TestStartingRandomTopicPickHandlesEmpty(t *testing.T) {
	if _, ok := StartingRandomTopicPick(nil, 10, func(int) int { return 0 }); ok {
		t.Error("沒有候選時應回 false")
	}
	if sc := StartingRandomTopicScores(nil, 10); sc != nil {
		t.Errorf("沒有候選時分數應是 nil,得到 %v", sc)
	}
}

// 每回合研究點 0 要當成 1(原版 var_1C == 0 → 1),不是除以零。
func TestStartingRandomHandlesZeroResearchPerTurn(t *testing.T) {
	avail := []ResearchTopic{TOPIC_ENGINEERING}
	sc := StartingRandomTopicScores(avail, 0)
	if len(sc) != 1 || sc[0] <= 0 {
		t.Errorf("每回合 0 點研究應當成 1 而不是崩潰,得到 %v", sc)
	}
}
