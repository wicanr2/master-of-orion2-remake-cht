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
	// 原版剛好達成本沒有突破率，必須嚴格超過成本。
	ps := PlayerState{ResearchTopic: topicCost400, ResearchProgress: 300}

	got, done := RunResearchPhase(ps, 100)

	if done {
		t.Fatalf("done = true,預期 false(300+100=400 尚無突破率)")
	}
	if got.ResearchProgress != 400 {
		t.Errorf("ResearchProgress = %d,預期保留 400", got.ResearchProgress)
	}
	if got.CompletedTopics[topicCost400] {
		t.Errorf("topicCost400 不應標記完成:%+v", got.CompletedTopics)
	}
}

func TestRunResearchPhaseBreakthroughClearsProgress(t *testing.T) {
	// 超過成本後固定 roll=1 成功；原版完成回寫把累積進度直接清零。
	ps := PlayerState{ResearchTopic: topicCost400, ResearchProgress: 0}

	got, done := RunResearchPhase(ps, 450)

	if !done {
		t.Fatalf("done = false,預期 true(450 >= 400)")
	}
	if got.ResearchProgress != 0 {
		t.Errorf("ResearchProgress = %d,預期突破後清零", got.ResearchProgress)
	}
	if !got.CompletedTopics[topicCost400] {
		t.Errorf("topicCost400 應標記完成:%+v", got.CompletedTopics)
	}
}

func TestRunResearchPhaseFailedBreakthroughKeepsAccumulatedProgress(t *testing.T) {
	ps := PlayerState{ResearchTopic: topicCost400, ResearchProgress: 400}
	got, done := RunResearchPhaseWithRoller(ps, 40, func(max int) int {
		if max != 100 {
			t.Fatalf("突破骰上限 = %d，預期 100", max)
		}
		return 11 // 440/400 = 10% chance，11 失敗
	})
	if done || got.ResearchProgress != 440 || got.CompletedTopics[topicCost400] {
		t.Fatalf("突破失敗應保留全部累積進度：done=%v progress=%d completed=%v",
			done, got.ResearchProgress, got.CompletedTopics)
	}
}

func TestRunResearchPhaseDoesNotConsumeRollAtExactCostOrWithoutNewResearch(t *testing.T) {
	called := 0
	roller := func(int) int { called++; return 1 }
	got, done := RunResearchPhaseWithRoller(PlayerState{ResearchTopic: topicCost400, ResearchProgress: 300}, 100, roller)
	if done || got.ResearchProgress != 400 || called != 0 {
		t.Fatalf("剛好成本不得擲突破骰：done=%v progress=%d calls=%d", done, got.ResearchProgress, called)
	}
	got, done = RunResearchPhaseWithRoller(PlayerState{ResearchTopic: topicCost400, ResearchProgress: 500}, 0, roller)
	if done || got.ResearchProgress != 500 || called != 0 {
		t.Fatalf("本回合無正研究產出不得擲骰：done=%v progress=%d calls=%d", done, got.ResearchProgress, called)
	}
}

func TestRunResearchPhaseZeroCostTopic(t *testing.T) {
	// field 0 是「沒有有效研究主題」；原版直接返回，不由回合鏈補發起始科技。
	ps := PlayerState{ResearchTopic: topicCost0, ResearchProgress: 0}

	got, done := RunResearchPhase(ps, 999)

	if done {
		t.Fatalf("done = true,預期 false(field 0 不進突破鏈)")
	}
	if got.ResearchProgress != 0 {
		t.Errorf("ResearchProgress = %d,預期 0(cost=0 不累加本回合研究點)", got.ResearchProgress)
	}
	if got.CompletedTopics[topicCost0] {
		t.Errorf("field 0 不應由研究回合標記完成:%+v", got.CompletedTopics)
	}
}

func TestRunResearchPhaseNilCompletedTopicsMap(t *testing.T) {
	// CompletedTopics 為 nil 時要安全建 map,不 panic。
	ps := PlayerState{ResearchTopic: topicCost400, ResearchProgress: 300}
	if ps.CompletedTopics != nil {
		t.Fatalf("測試前提錯誤:CompletedTopics 應為 nil")
	}

	got, done := RunResearchPhase(ps, 101)

	if got.CompletedTopics == nil {
		t.Fatalf("CompletedTopics 仍為 nil,應已安全建立")
	}
	if !done || !got.CompletedTopics[topicCost400] {
		t.Errorf("nil map 情境下完成判定錯誤:done=%v, map=%+v", done, got.CompletedTopics)
	}
}

func TestRunResearchPhase_HyperAdvancedCostOverride(t *testing.T) {
	// TOPIC_HYPER_BIOLOGY 是 8 個共用 Hyper-Advanced Lv1 成本的主題之一,套件級硬編值 25000
	// (techtree.go),見 gamedata.IsHyperAdvancedTopic/HyperAdvancedCost。
	hyperTopic := gamedata.TOPIC_HYPER_BIOLOGY

	t.Run("override>0 對 Hyper 主題生效(1.3 profile=15000)", func(t *testing.T) {
		ps := PlayerState{ResearchTopic: hyperTopic, ResearchProgress: 0, HyperAdvancedResearchCost: 15000}

		got, done := RunResearchPhase(ps, 15001)

		if !done {
			t.Fatalf("done = false,預期 true(15001 嚴格超過覆寫成本 15000)")
		}
		if got.ResearchProgress != 0 {
			t.Errorf("ResearchProgress = %d,預期 0(15000-15000)", got.ResearchProgress)
		}
		if !got.CompletedTopics[hyperTopic] {
			t.Errorf("hyperTopic 應標記完成:%+v", got.CompletedTopics)
		}
	})

	t.Run("override>0 但未打到覆寫成本則不完成(驗證真的用 15000 而非套件級 25000)", func(t *testing.T) {
		ps := PlayerState{ResearchTopic: hyperTopic, ResearchProgress: 0, HyperAdvancedResearchCost: 15000}

		got, done := RunResearchPhase(ps, 20000) // >= 套件級 25000?否;>= 覆寫 15000?是——用來反證真的讀覆寫值

		if !done {
			t.Fatalf("done = false,預期 true:20000 >= 覆寫成本 15000,若誤用套件級 25000 則會是 false")
		}
		if got.ResearchProgress != 0 {
			t.Errorf("ResearchProgress = %d,預期突破後清零", got.ResearchProgress)
		}
	})

	t.Run("override=0 時退回 gamedata 套件級預設 25000(Profile15 行為不變)", func(t *testing.T) {
		ps := PlayerState{ResearchTopic: hyperTopic, ResearchProgress: 0, HyperAdvancedResearchCost: 0}

		got, done := RunResearchPhase(ps, 15000)

		if done {
			t.Fatalf("done = true,預期 false:15000 < 套件級預設成本 25000,override=0 不應覆寫")
		}
		if got.ResearchProgress != 15000 {
			t.Errorf("ResearchProgress = %d,預期 15000(未完成,原樣累加)", got.ResearchProgress)
		}

		got2, done2 := RunResearchPhase(got, 10001) // 累加到 25001，產生最低突破率
		if !done2 {
			t.Fatalf("done2 = false,預期 true:15000+10001=25001 應可突破")
		}
		if got2.ResearchProgress != 0 {
			t.Errorf("ResearchProgress = %d,預期 0(25000-25000)", got2.ResearchProgress)
		}
	})

	t.Run("override>0 對非 Hyper 主題不生效", func(t *testing.T) {
		// topicCost400 成本固定 400,即使 override 設為離譜的值也不應套用(只對 Hyper 主題生效)。
		ps := PlayerState{ResearchTopic: topicCost400, ResearchProgress: 0, HyperAdvancedResearchCost: 15000}

		got, done := RunResearchPhase(ps, 401)

		if !done {
			t.Fatalf("done = false,預期 true:非 Hyper 主題仍用成本 400，401 可突破")
		}
		if got.ResearchProgress != 0 {
			t.Errorf("ResearchProgress = %d,預期 0(400-400,非 400-15000)", got.ResearchProgress)
		}
	})
}

func TestRunResearchPhasePreservesExistingCompletedTopics(t *testing.T) {
	// 已完成的舊主題不應被新一輪呼叫覆蓋掉。
	ps := PlayerState{
		ResearchTopic:    topicCost400,
		ResearchProgress: 300,
		CompletedTopics:  map[gamedata.ResearchTopic]bool{topicCost0: true},
	}

	got, done := RunResearchPhase(ps, 101)

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

func TestRunResearchPhaseRepeatsHyperAdvancedLevel(t *testing.T) {
	topic := gamedata.TOPIC_HYPER_PHYSICS
	ps := PlayerState{ResearchTopic: topic, HyperAdvancedResearchCost: 25000}
	ps, done := RunResearchPhase(ps, 25001)
	if !done || ps.HyperAdvancedLevels[topic] != 1 || !ps.CompletedTopics[topic] {
		t.Fatalf("Hyper 第一次完成狀態錯誤: done=%v levels=%v completed=%v",
			done, ps.HyperAdvancedLevels, ps.CompletedTopics)
	}
	ps, done = RunResearchPhase(ps, 34999)
	if done || ps.HyperAdvancedLevels[topic] != 1 {
		t.Fatalf("Hyper 第二級在 34999/35000 不應完成: done=%v levels=%v", done, ps.HyperAdvancedLevels)
	}
	ps, done = RunResearchPhase(ps, 2)
	if !done || ps.HyperAdvancedLevels[topic] != 2 {
		t.Fatalf("Hyper 第二次完成應累積為 2: done=%v levels=%v", done, ps.HyperAdvancedLevels)
	}
}

func TestForceCompleteResearchTopicUsesSelectedApplicationAndClearsField(t *testing.T) {
	topic := gamedata.TOPIC_ADVANCED_BIOLOGY
	choice := gamedata.ResearchChoiceFor(topic)
	ps, ok := SelectResearchApplication(PlayerState{ResearchTopic: topic, ResearchProgress: 17}, topic, choice.Choices[1])
	if !ok {
		t.Fatal("測試前提：應可預選第二個 application")
	}
	got, savedTopic, completed := ForceCompleteResearchTopic(ps)
	if !completed || savedTopic != topic || !got.CompletedTopics[topic] || got.ChosenTech[topic] != choice.Choices[1] {
		t.Fatalf("強制完成結果不符：completed=%v topic=%v state=%+v", completed, savedTopic, got)
	}
	if got.ResearchTopic != 0 || got.ResearchProgress != 0 || got.HasResearchApplication {
		t.Fatalf("原版事件完成後必須清空 field/RP/application：%+v", got)
	}
}

func TestForceCompleteResearchTopicRepeatsHyperAndNoTopicDoesNotInventTech(t *testing.T) {
	hyper := gamedata.TOPIC_HYPER_PHYSICS
	got, _, completed := ForceCompleteResearchTopic(PlayerState{ResearchTopic: hyper,
		HyperAdvancedLevels: map[gamedata.ResearchTopic]int{hyper: 2}})
	if !completed || got.HyperAdvancedLevels[hyper] != 3 {
		t.Fatalf("Hyper-Advanced 應增加一級：completed=%v levels=%v", completed, got.HyperAdvancedLevels)
	}
	empty, topic, completed := ForceCompleteResearchTopic(PlayerState{ResearchProgress: 999, HasResearchApplication: true})
	if completed || topic != 0 || empty.ResearchProgress != 0 || empty.HasResearchApplication || len(empty.CompletedTopics) != 0 {
		t.Fatalf("無 field 不得捏造科技，但要清理暫存狀態：completed=%v topic=%v state=%+v", completed, topic, empty)
	}
}

func TestRunResearchPhaseMigratesLegacyHyperCompletion(t *testing.T) {
	topic := gamedata.TOPIC_HYPER_FIELDS
	ps := PlayerState{
		ResearchTopic: topic, HyperAdvancedResearchCost: 25000,
		CompletedTopics: map[gamedata.ResearchTopic]bool{topic: true},
	}
	ps, done := RunResearchPhase(ps, 35001)
	if !done || ps.HyperAdvancedLevels[topic] != 2 {
		t.Fatalf("舊存檔的一級加本次完成應為二級: done=%v levels=%v", done, ps.HyperAdvancedLevels)
	}
}
