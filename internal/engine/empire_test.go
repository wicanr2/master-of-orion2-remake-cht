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

func TestRunEmpireTurnBC(t *testing.T) {
	// Tolerant 種族免污染清理:淨工業=毛工業。Workers 2*10=20,稅率 50% → 稅收 10。
	colonies := []ColonyState{
		{Population: 5, PopMax: 20, Workers: 2, IndustryPerWorker: 10,
			PlanetSize: gamedata.TINY_PLANET, TolerantRace: true},
	}
	ps := PlayerState{BC: 100, TaxRate: 50, Maintenance: 3,
		ResearchTopic: gamedata.ResearchTopic(1)}
	out := RunEmpireTurn(ps, colonies)

	if out.TaxRevenue != 10 { // 20*50/100
		t.Errorf("稅收 = %d,預期 10", out.TaxRevenue)
	}
	if out.NetBC != 7 { // 10 - 3 維護
		t.Errorf("淨 BC = %d,預期 7", out.NetBC)
	}
	if out.Player.BC != 107 { // 100 + 7
		t.Errorf("國庫 = %d,預期 107", out.Player.BC)
	}
}

// TestRunEmpireTurnTradeGoods 驗證「貿易品」殖民地(cs.TradeGoods=true)的淨工業改以 2:1
// 換算成 BC(gamedata.TradeGoodsIncome,一般種族、非 Fantastic Trader),計入
// EmpireOutput.TradeGoodsRevenue 與 NetBC。
func TestRunEmpireTurnTradeGoods(t *testing.T) {
	// Tolerant 種族免污染清理:淨工業=毛工業=20(Workers 2*10)。未設稅率(0%)、無農夫
	// (食物盈餘為負,不計餘糧收入),隔離出貿易品收入單一變數。
	colonies := []ColonyState{
		{Population: 5, PopMax: 20, Workers: 2, IndustryPerWorker: 10,
			PlanetSize: gamedata.TINY_PLANET, TolerantRace: true, TradeGoods: true},
	}
	ps := PlayerState{BC: 100, Maintenance: 3, ResearchTopic: gamedata.ResearchTopic(1)}
	out := RunEmpireTurn(ps, colonies)

	if out.TradeGoodsRevenue != 10 { // 20 淨工業 * 1/2(一般種族 2:1)
		t.Errorf("貿易品收入 = %d,預期 10", out.TradeGoodsRevenue)
	}
	if out.NetBC != 7 { // 0 稅收 + 0 餘糧收入(負盈餘不計) + 10 貿易品 - 3 維護
		t.Errorf("淨 BC = %d,預期 7", out.NetBC)
	}
	if out.Player.BC != 107 { // 100 + 7
		t.Errorf("國庫 = %d,預期 107", out.Player.BC)
	}
}

// TestRunEmpireTurnTradeGoodsFalseSkipsRevenue 驗證非貿易品殖民地(cs.TradeGoods 預設 false)
// 不計入 TradeGoodsRevenue——確保旗標關閉時行為與加欄位前一致,不會誤觸發轉換。
func TestRunEmpireTurnTradeGoodsFalseSkipsRevenue(t *testing.T) {
	colonies := []ColonyState{
		{Population: 5, PopMax: 20, Workers: 2, IndustryPerWorker: 10,
			PlanetSize: gamedata.TINY_PLANET, TolerantRace: true}, // TradeGoods 預設 false
	}
	ps := PlayerState{BC: 100, Maintenance: 3, ResearchTopic: gamedata.ResearchTopic(1)}
	out := RunEmpireTurn(ps, colonies)

	if out.TradeGoodsRevenue != 0 {
		t.Errorf("非貿易品殖民地不應計入貿易品收入,實得 %d", out.TradeGoodsRevenue)
	}
}
