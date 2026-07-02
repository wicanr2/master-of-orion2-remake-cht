package engine

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestRunEmpireTurn(t *testing.T) {
	// 兩個殖民地,研究總點推進到剛好完成 topic(1)(成本 400)。
	colonies := []ColonyState{
		{Population: 10, PopMax: 20, Farmers: 4, Workers: 4, Scientists: 2,
			FoodPerFarmer: 3, IndustryPerWorker: 5, ResearchPerScientist: 100,
			PlanetSize: gamedata.MEDIUM_PLANET}, // 研究 200
		{Population: 8, PopMax: 20, Farmers: 3, Workers: 3, Scientists: 2,
			FoodPerFarmer: 3, IndustryPerWorker: 5, ResearchPerScientist: 100,
			PlanetSize: gamedata.SMALL_PLANET}, // 研究 200
	}
	ps := PlayerState{ResearchTopic: gamedata.ResearchTopic(1), ResearchProgress: 0} // cost 400
	out := RunEmpireTurn(ps, colonies)

	if len(out.Colonies) != 2 {
		t.Fatalf("殖民地輸出數 = %d,預期 2", len(out.Colonies))
	}
	if out.TotalResearch != 400 { // 200+200
		t.Errorf("總研究 = %d,預期 400", out.TotalResearch)
	}
	if !out.ResearchDone { // 400>=400 完成
		t.Error("研究應完成")
	}
	if !out.Player.CompletedTopics[gamedata.ResearchTopic(1)] {
		t.Error("topic 1 應標記完成")
	}
	// 食物盈餘聚合:c1 surplus=12-10=2,c2=9-8=1 → 3
	if out.TotalFood != 3 {
		t.Errorf("總食物盈餘 = %d,預期 3", out.TotalFood)
	}
}
