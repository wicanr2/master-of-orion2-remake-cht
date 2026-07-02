package engine

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// topic 用真實研究表資料(internal/gamedata/techtree.go):
//   - gamedata.ResearchTopic(1) = TOPIC_ADVANCED_BIOLOGY,Cost = 400
//   - gamedata.ResearchTopic(0) = TOPIC_STARTING_TECH,   Cost = 0
const (
	topicCost400 = gamedata.ResearchTopic(1)
	topicCost0   = gamedata.ResearchTopic(0)
)

func TestRunResearchPhaseNotComplete(t *testing.T) {
	// 未達成本:0 + 100 = 100 < 400,不完成、不進 CompletedTopics。
	ps := PlayerState{ResearchTopic: topicCost400, ResearchProgress: 0}

	got, done := RunResearchPhase(ps, 100)

	if done {
		t.Fatalf("done = true,預期 false(100 < 400 尚未達成本)")
	}
	if got.ResearchProgress != 100 {
		t.Errorf("ResearchProgress = %d,預期 100", got.ResearchProgress)
	}
	if got.CompletedTopics[topicCost400] {
		t.Errorf("topicCost400 不應標記完成:%+v", got.CompletedTopics)
	}
}

func TestRunResearchPhaseExactCost(t *testing.T) {
	// 剛好達成本:300 + 100 = 400 == cost,完成、溢出 0。
	ps := PlayerState{ResearchTopic: topicCost400, ResearchProgress: 300}

	got, done := RunResearchPhase(ps, 100)

	if !done {
		t.Fatalf("done = false,預期 true(300+100=400 剛好達成本)")
	}
	if got.ResearchProgress != 0 {
		t.Errorf("ResearchProgress = %d,預期 0(400-400)", got.ResearchProgress)
	}
	if !got.CompletedTopics[topicCost400] {
		t.Errorf("topicCost400 應標記完成:%+v", got.CompletedTopics)
	}
}

func TestRunResearchPhaseOverflowCarriesOver(t *testing.T) {
	// 超過成本:0 + 450 = 450 >= 400,完成,溢出 50 保留到下一主題(不歸零)。
	ps := PlayerState{ResearchTopic: topicCost400, ResearchProgress: 0}

	got, done := RunResearchPhase(ps, 450)

	if !done {
		t.Fatalf("done = false,預期 true(450 >= 400)")
	}
	if got.ResearchProgress != 50 {
		t.Errorf("ResearchProgress = %d,預期 50(450-400 溢出保留)", got.ResearchProgress)
	}
	if !got.CompletedTopics[topicCost400] {
		t.Errorf("topicCost400 應標記完成:%+v", got.CompletedTopics)
	}
}

func TestRunResearchPhaseZeroCostTopic(t *testing.T) {
	// cost == 0(TOPIC_STARTING_TECH):視為已完成/無需研究,progress 不動,回 true。
	ps := PlayerState{ResearchTopic: topicCost0, ResearchProgress: 0}

	got, done := RunResearchPhase(ps, 999)

	if !done {
		t.Fatalf("done = false,預期 true(cost=0 視為已完成)")
	}
	if got.ResearchProgress != 0 {
		t.Errorf("ResearchProgress = %d,預期 0(cost=0 不累加本回合研究點)", got.ResearchProgress)
	}
	if !got.CompletedTopics[topicCost0] {
		t.Errorf("topicCost0 應標記完成:%+v", got.CompletedTopics)
	}
}

func TestRunResearchPhaseNilCompletedTopicsMap(t *testing.T) {
	// CompletedTopics 為 nil 時要安全建 map,不 panic。
	ps := PlayerState{ResearchTopic: topicCost400, ResearchProgress: 300}
	if ps.CompletedTopics != nil {
		t.Fatalf("測試前提錯誤:CompletedTopics 應為 nil")
	}

	got, done := RunResearchPhase(ps, 100)

	if got.CompletedTopics == nil {
		t.Fatalf("CompletedTopics 仍為 nil,應已安全建立")
	}
	if !done || !got.CompletedTopics[topicCost400] {
		t.Errorf("nil map 情境下完成判定錯誤:done=%v, map=%+v", done, got.CompletedTopics)
	}
}

func TestRunResearchPhasePreservesExistingCompletedTopics(t *testing.T) {
	// 已完成的舊主題不應被新一輪呼叫覆蓋掉。
	ps := PlayerState{
		ResearchTopic:    topicCost400,
		ResearchProgress: 300,
		CompletedTopics:  map[gamedata.ResearchTopic]bool{topicCost0: true},
	}

	got, done := RunResearchPhase(ps, 100)

	if !done {
		t.Fatalf("done = false,預期 true")
	}
	if !got.CompletedTopics[topicCost0] {
		t.Errorf("舊主題 topicCost0 的完成標記不應消失:%+v", got.CompletedTopics)
	}
	if !got.CompletedTopics[topicCost400] {
		t.Errorf("新主題 topicCost400 應標記完成:%+v", got.CompletedTopics)
	}
}
