package shell

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestOriginalAIExactBuildingScores(t *testing.T) {
	colony := engine.ColonyState{Population: 6}
	tests := []struct {
		name      string
		balanced  int
		honorable int
	}{
		{"太空大學", 5, 5},
		{"自動工廠", 19, 21},
		{"深層核心礦場", 18, 22},
		{"機器人工廠", 12, 14},
		{"機器人採礦廠", 11, 13},
	}
	for _, tt := range tests {
		b, ok := gamedata.BuildingByNameZH(tt.name)
		if !ok {
			t.Fatalf("測試建築不存在：%s", tt.name)
		}
		got, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityXenophobic, false)
		if !exact || got != tt.balanced {
			t.Errorf("%s 一般性格分數=(%d,%v)，want (%d,true)", tt.name, got, exact, tt.balanced)
		}
		got, exact = originalAIExactBuildingScore(b, colony, ai.PersonalityHonorable, false)
		if !exact || got != tt.honorable {
			t.Errorf("%s Honorable 分數=(%d,%v)，want (%d,true)", tt.name, got, exact, tt.honorable)
		}
	}
	fallback, ok := gamedata.BuildingByNameZH("研究實驗室")
	if !ok {
		t.Fatal("研究實驗室不存在")
	}
	if score, exact := originalAIExactBuildingScore(fallback, colony, ai.PersonalityHonorable, false); exact || score != 0 {
		t.Fatalf("未閉合 case 不得冒稱 exact：score=%d exact=%v", score, exact)
	}
}

func TestOriginalAILateTechResearchBuildingScoresAreZero(t *testing.T) {
	for _, name := range []string{"自動實驗室", "銀河網路中心", "行星超級電腦", "研究實驗室"} {
		b, ok := gamedata.BuildingByNameZH(name)
		if !ok {
			t.Fatalf("測試建築不存在：%s", name)
		}
		if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{Population: 9}, ai.PersonalityErratic, true); !exact || score != 0 {
			t.Errorf("%s 晚期科技分數=(%d,%v)，want (0,true)", name, score, exact)
		}
		if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{Population: 9}, ai.PersonalityErratic, false); exact || score != 0 {
			t.Errorf("%s 晚期科技前仍受未知 ah gate，不能冒稱 exact：(%d,%v)", name, score, exact)
		}
	}
}

func TestAIOriginalLateTechReachedUsesSelectedOrCompletedHyperField(t *testing.T) {
	ps := engine.PlayerState{ResearchTopic: gamedata.TOPIC_HYPER_PHYSICS}
	if !aiOriginalLateTechReached(ps) {
		t.Fatal("選中 raw 75..82 的 Hyper field 時應立即進入晚期科技")
	}
	ps = engine.PlayerState{CompletedTopics: map[gamedata.ResearchTopic]bool{gamedata.TOPIC_HYPER_BIOLOGY: true}}
	if !aiOriginalLateTechReached(ps) {
		t.Fatal("完成任一 Hyper field 後晚期科技狀態應持續")
	}
	ps = engine.PlayerState{HyperAdvancedLevels: map[gamedata.ResearchTopic]int{gamedata.TOPIC_HYPER_SOCIOLOGY: 2}}
	if !aiOriginalLateTechReached(ps) {
		t.Fatal("舊存檔只保留 Hyper 等級時仍應重建晚期科技狀態")
	}
	if aiOriginalLateTechReached(engine.PlayerState{ResearchTopic: gamedata.TOPIC_ADVANCED_CONSTRUCTION}) {
		t.Fatal("一般研究 field 不得誤判為晚期科技")
	}
}

func TestAIColonyBuildConsumesIndustryBeforeShipPool(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Colonies = []engine.ColonyState{{Population: 4, PopMax: 10, IndustryPerWorker: 2}}
	a.ColonyStars = []int{7}
	a.ColonyBuildings = []map[string]bool{{}}
	a.ColonyBuilds = nil
	a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{
		gamedata.TOPIC_ADVANCED_CONSTRUCTION: true,
	}
	// 此主題同時開多棟建築；把其餘候選標成已完成，讓測試只驗證逐殖民地
	// 產品與產能消費，不把近似分數表的選擇偏好凍成契約。
	for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
		if b.NameZH != "自動工廠" {
			a.ColonyBuildings[0][b.NameZH] = true
		}
	}
	out := engine.EmpireOutput{Colonies: []engine.ColonyOutput{{NetIndustry: 30}}}
	if got := s.advanceAIColonyBuilds(0, out); got != 0 {
		t.Fatalf("建築產品應先消費殖民地工業，交給造艦=%d", got)
	}
	build := a.ColonyBuilds[7]
	if build.Name != "自動工廠" || build.Progress != 30 || build.Cost != 60 {
		t.Fatalf("AI 殖民地產品=%+v，want 自動工廠 30/60", build)
	}
	if got := s.advanceAIColonyBuilds(0, out); got != 0 {
		t.Fatalf("完工回合不應重複把同份產能交給造艦，got %d", got)
	}
	if !a.ColonyBuildings[0]["自動工廠"] || a.Colonies[0].FlatIndustry != 5 || a.Colonies[0].IndustryPerWorker != 3 {
		t.Fatalf("完工建築未接到 typed 效果：building=%v colony=%+v", a.ColonyBuildings[0], a.Colonies[0])
	}
	if _, ok := a.ColonyBuilds[7]; ok {
		t.Fatal("完工後應清除目前產品")
	}
}

func TestAIColonyBuildSurvivesSaveLoad(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers[0].ColonyBuilds = map[int]ColonyBuild{
		5: {Name: "自動工廠", Cost: 60, Progress: 37},
	}
	path := filepath.Join(t.TempDir(), "ai-colony-build.json")
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSession(path)
	if err != nil {
		t.Fatal(err)
	}
	if build := got.AIPlayers[0].ColonyBuilds[5]; build.Name != "自動工廠" || build.Cost != 60 || build.Progress != 37 {
		t.Fatalf("AI 殖民地產品未完整往返：%+v", build)
	}
}

func TestAIColonyWithoutBuildableBuildingFeedsShipProduction(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Colonies = []engine.ColonyState{{Population: 4, PopMax: 10}}
	a.ColonyStars = []int{3}
	a.ColonyBuildings = []map[string]bool{{}}
	a.Player.CompletedTopics = nil
	out := engine.EmpireOutput{Colonies: []engine.ColonyOutput{{NetIndustry: 17}}}
	if got := s.advanceAIColonyBuilds(0, out); got != 17 {
		t.Fatalf("無可建建築時造艦產能=%d，want 17", got)
	}
}
