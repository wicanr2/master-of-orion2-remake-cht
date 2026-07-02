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

func TestRunEmpireTurnResearchNotComplete(t *testing.T) {
	// 研究總點不足成本 → 累積但不完成。
	colonies := []ColonyState{
		{Population: 5, PopMax: 20, Scientists: 1, ResearchPerScientist: 50,
			PlanetSize: gamedata.SMALL_PLANET},
	}
	ps := PlayerState{ResearchTopic: gamedata.ResearchTopic(1)} // cost 400
	out := RunEmpireTurn(ps, colonies)
	if out.ResearchDone {
		t.Error("研究不應完成(50 < 400)")
	}
	if out.Player.ResearchProgress != 50 {
		t.Errorf("研究進度 = %d,預期 50", out.Player.ResearchProgress)
	}
}

func TestRunEmpireTurnMultiTurnProgression(t *testing.T) {
	// 多回合推進:同一組殖民地連跑數回合,把 output.Player 回饋為下回合輸入,
	// 驗證研究進度跨回合累積,並在累積達成本(400)的那回合完成。
	colonies := []ColonyState{
		{Population: 6, PopMax: 20, Scientists: 3, ResearchPerScientist: 50,
			PlanetSize: gamedata.MEDIUM_PLANET}, // 每回合研究 150
	}
	ps := PlayerState{ResearchTopic: gamedata.ResearchTopic(1)} // cost 400

	var completedTurn int
	for turn := 1; turn <= 3; turn++ {
		out := RunEmpireTurn(ps, colonies)
		ps = out.Player // 狀態帶到下回合
		if out.ResearchDone {
			completedTurn = turn
			break
		}
	}
	// 回合1:150、回合2:300、回合3:450≥400 → 第 3 回合完成,溢出保留 50
	if completedTurn != 3 {
		t.Errorf("研究應於第 3 回合完成,實際第 %d 回合", completedTurn)
	}
	if !ps.CompletedTopics[gamedata.ResearchTopic(1)] {
		t.Error("完成後 topic 1 應標記")
	}
	if ps.ResearchProgress != 50 { // 450-400 溢出
		t.Errorf("完成後溢出進度 = %d,預期 50", ps.ResearchProgress)
	}
}
