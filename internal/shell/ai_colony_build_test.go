package shell

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

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
