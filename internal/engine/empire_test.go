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
			PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT}, // 研究 200
		{Population: 8, PopMax: 20, Farmers: 3, Workers: 3, Scientists: 2,
			FoodPerFarmer: 3, IndustryPerWorker: 5, ResearchPerScientist: 100,
			PlanetSize: gamedata.SMALL_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT}, // 研究 200
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
			PlanetSize: gamedata.SMALL_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT},
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
			PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT}, // 每回合研究 150
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
			PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, TolerantRace: true},
	}
	ps := PlayerState{BC: 100, TaxRate: 50, Maintenance: 3,
		ResearchTopic: gamedata.ResearchTopic(1)}
	out := RunEmpireTurn(ps, colonies)

	// 2026-07-12 收入模型:併入人頭基礎收入(Pop5 × 1 BC = 5,見 gamedata.BaseIncomePerPopHalfBC)。
	if out.TaxRevenue != 15 { // 工業稅 10(20*50/100)+ 人頭 5
		t.Errorf("money 收入 = %d,預期 15", out.TaxRevenue)
	}
	if out.NetBC != 12 { // 15 - 3 維護
		t.Errorf("淨 BC = %d,預期 12", out.NetBC)
	}
	if out.Player.BC != 112 { // 100 + 12
		t.Errorf("國庫 = %d,預期 112", out.Player.BC)
	}
}

// TestRunEmpireTurnCommandOverflow 驗證指揮評等(Command Rating)供給不足艦艇需求時,
// 每回合每未覆蓋點扣 10 BC(GAME_MANUAL.pdf p.169,gamedata.IncomeCommandOverflowCost),
// 並正確併入 NetBC/Player.BC,曝露在 EmpireOutput.CommandOverflowCost。
//
// 2026-07-11 附註:本測試直接手寫 PlayerState.CommandPointsSupply=1(任意取值,測引擎公式本身
// 的 uncovered/overflow 算術),不是透過 shell.totalCommandPointsSupply() 算出來的實際供給——
// 帝國基礎值 gamedata.CommandPointsBase(=5,見該常數 oracle 反推註解)是 shell 層
// totalCommandPointsSupply 才會加的東西,RunEmpireTurn 本身只認呼叫端傳進來的
// CommandPointsSupply/UsedCommandPoints 兩個數字,不知道、也不需要知道基礎值怎麼來的。
// 因此這裡刻意不跟著 CommandPointsBase 修復「+5」,改成加 5 反而會讓 uncovered 從 2 變 -3
// (夾到 0),整個測試失去驗證「超支路徑」的意義。真正會受 CommandPointsBase 影響的整合測試
// 在 internal/shell/command_points_test.go(TestEndTurnCommandOverflowPenalty)與
// internal/shell/events_test.go(bcCrashFloor300Turns),已個別更新。
func TestRunEmpireTurnCommandOverflow(t *testing.T) {
	colonies := []ColonyState{
		{Population: 5, PopMax: 20, Workers: 2, IndustryPerWorker: 10,
			PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, TolerantRace: true},
	}
	// 稅收 10(20*50/100),維護費 3;供給 1(僅星基)、需求 3(2 艘 Frigate+1 艘 Destroyer=1+1+2)
	// → uncovered=2 → 超支懲罰 2*10=20 BC。
	ps := PlayerState{BC: 100, TaxRate: 50, Maintenance: 3, CommandPointsSupply: 1, UsedCommandPoints: 3,
		ResearchTopic: gamedata.ResearchTopic(1)}
	out := RunEmpireTurn(ps, colonies)

	if out.CommandOverflowCost != 20 {
		t.Errorf("CommandOverflowCost = %d,預期 20", out.CommandOverflowCost)
	}
	if out.NetBC != -8 { // 15(工業稅10+人頭5) - 3(維護) - 20(指揮評等超支) = -8
		t.Errorf("淨 BC = %d,預期 -8", out.NetBC)
	}
	if out.Player.BC != 92 { // 100 - 8
		t.Errorf("國庫 = %d,預期 92", out.Player.BC)
	}
}

// TestRunEmpireTurnCommandSupplyCoversDemand 驗證供給 >= 需求時無懲罰(NetBC 不含超支扣款)。
func TestRunEmpireTurnCommandSupplyCoversDemand(t *testing.T) {
	colonies := []ColonyState{
		{Population: 5, PopMax: 20, Workers: 2, IndustryPerWorker: 10,
			PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, TolerantRace: true},
	}
	ps := PlayerState{BC: 100, TaxRate: 50, Maintenance: 3, CommandPointsSupply: 3, UsedCommandPoints: 3,
		ResearchTopic: gamedata.ResearchTopic(1)}
	out := RunEmpireTurn(ps, colonies)

	if out.CommandOverflowCost != 0 {
		t.Errorf("供給=需求時 CommandOverflowCost 應為 0,got %d", out.CommandOverflowCost)
	}
	if out.NetBC != 12 { // 15(工業稅10+人頭5) - 3,同 TestRunEmpireTurnBC
		t.Errorf("淨 BC = %d,預期 12", out.NetBC)
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
			PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, TolerantRace: true, TradeGoods: true},
	}
	ps := PlayerState{BC: 100, Maintenance: 3, ResearchTopic: gamedata.ResearchTopic(1)}
	out := RunEmpireTurn(ps, colonies)

	if out.TradeGoodsRevenue != 10 { // 20 淨工業 * 1/2(一般種族 2:1)
		t.Errorf("貿易品收入 = %d,預期 10", out.TradeGoodsRevenue)
	}
	if out.NetBC != 12 { // 人頭5 + 0 稅收 + 0 餘糧收入(負盈餘不計) + 10 貿易品 - 3 維護
		t.Errorf("淨 BC = %d,預期 12", out.NetBC)
	}
	if out.Player.BC != 112 { // 100 + 12
		t.Errorf("國庫 = %d,預期 112", out.Player.BC)
	}
}

// TestRunEmpireTurnIncomeBonusPercent 驗證 IncomeBonusPercent(太空港 p.79 +50%、行星證券
// 交易所 p.93 +100%,可疊加)精確套用在「該殖民地」當回合的收入小計上,而非帝國整體近似:
// Tolerant 種族免污染清理,淨工業=毛工業=20(Workers 2*10),稅率 50% → 稅收基數 10;
// IncomeBonusPercent=150(太空港50+證券100)→ 稅收 = 10*250/100 = 25。
func TestRunEmpireTurnIncomeBonusPercent(t *testing.T) {
	colonies := []ColonyState{
		{Population: 5, PopMax: 20, Workers: 2, IndustryPerWorker: 10,
			PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, TolerantRace: true, IncomeBonusPercent: 150},
	}
	ps := PlayerState{BC: 100, TaxRate: 50, Maintenance: 3, ResearchTopic: gamedata.ResearchTopic(1)}
	out := RunEmpireTurn(ps, colonies)

	if out.TaxRevenue != 37 {
		t.Errorf("加成後 money 收入 = %d,預期 37((工業稅10+人頭5)×250%%)", out.TaxRevenue)
	}
	if out.NetBC != 34 { // 37 - 3 維護
		t.Errorf("淨 BC = %d,預期 34", out.NetBC)
	}
}

// TestRunEmpireTurnIncomeBonusPercentPerColony 驗證 IncomeBonusPercent 只影響「有該旗標」
// 的殖民地,不會誤把加成套到帝國內其他殖民地的收入上(逐殖民地套用,非先加總帝國收入再打折)。
func TestRunEmpireTurnIncomeBonusPercentPerColony(t *testing.T) {
	colonies := []ColonyState{
		{Population: 5, PopMax: 20, Workers: 2, IndustryPerWorker: 10,
			PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, TolerantRace: true, IncomeBonusPercent: 50}, // 稅收 10→15
		{Population: 5, PopMax: 20, Workers: 2, IndustryPerWorker: 10,
			PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, TolerantRace: true}, // 無加成,稅收 10
	}
	ps := PlayerState{TaxRate: 50, ResearchTopic: gamedata.ResearchTopic(1)}
	out := RunEmpireTurn(ps, colonies)

	if out.TaxRevenue != 37 { // 殖1(工業稅10+人頭5)×150%=22 + 殖2(10+5)=15
		t.Errorf("兩殖民地合計 money 收入 = %d,預期 37(僅第一個殖民地吃到 +50%%)", out.TaxRevenue)
	}
}

// TestRunEmpireTurnTradeGoodsFalseSkipsRevenue 驗證非貿易品殖民地(cs.TradeGoods 預設 false)
// 不計入 TradeGoodsRevenue——確保旗標關閉時行為與加欄位前一致,不會誤觸發轉換。
func TestRunEmpireTurnTradeGoodsFalseSkipsRevenue(t *testing.T) {
	colonies := []ColonyState{
		{Population: 5, PopMax: 20, Workers: 2, IndustryPerWorker: 10,
			PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, TolerantRace: true}, // TradeGoods 預設 false
	}
	ps := PlayerState{BC: 100, Maintenance: 3, ResearchTopic: gamedata.ResearchTopic(1)}
	out := RunEmpireTurn(ps, colonies)

	if out.TradeGoodsRevenue != 0 {
		t.Errorf("非貿易品殖民地不應計入貿易品收入,實得 %d", out.TradeGoodsRevenue)
	}
}

// TestRunEmpireTurnGovtBonusMoneyPercent 驗證政府 money 加成(Democracy +50%,見
// gamedata.IncomeGovtBonusDemocracyMoneyPercent)套用在帝國「已加總」的稅收+餘糧收入+貿易品
// 收入上,差額計入 TaxRevenue、併入 NetBC。Tolerant 種族免污染清理,淨工業=毛工業=20
// (Workers 2*10),稅率 50% → 稅收基數 10;+50% → 15。
func TestRunEmpireTurnGovtBonusMoneyPercent(t *testing.T) {
	colonies := []ColonyState{
		{Population: 5, PopMax: 20, Workers: 2, IndustryPerWorker: 10,
			PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, TolerantRace: true},
	}
	ps := PlayerState{BC: 100, TaxRate: 50, Maintenance: 3, ResearchTopic: gamedata.ResearchTopic(1),
		GovtBonusMoneyPercent: gamedata.IncomeGovtBonusDemocracyMoneyPercent}
	out := RunEmpireTurn(ps, colonies)

	if out.TaxRevenue != 22 { // (工業稅10+人頭5)=15 * 150/100
		t.Errorf("加成後 money 收入 = %d,預期 22((10+5)×150%%)", out.TaxRevenue)
	}
	if out.NetBC != 19 { // 22 - 3 維護
		t.Errorf("淨 BC = %d,預期 19", out.NetBC)
	}
}

// TestRunEmpireTurnGovtBonusMoneyPercentZeroNoOp 驗證 GovtBonusMoneyPercent=0(手冊未列出加成
// 的政府,如 demo 用的 Dictatorship)時完全不影響稅收——確保「無加成政府」與「加欄位前」行為
// 一致,對齊 demo 開局經濟軌跡不因本次接線變化。
func TestRunEmpireTurnGovtBonusMoneyPercentZeroNoOp(t *testing.T) {
	colonies := []ColonyState{
		{Population: 5, PopMax: 20, Workers: 2, IndustryPerWorker: 10,
			PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, TolerantRace: true},
	}
	ps := PlayerState{BC: 100, TaxRate: 50, Maintenance: 3, ResearchTopic: gamedata.ResearchTopic(1)} // GovtBonusMoneyPercent 零值
	out := RunEmpireTurn(ps, colonies)

	if out.TaxRevenue != 15 { // 工業稅10 + 人頭5(無政府加成)
		t.Errorf("無政府加成時 money 收入 = %d,預期 15", out.TaxRevenue)
	}
	if out.NetBC != 12 {
		t.Errorf("淨 BC = %d,預期 12", out.NetBC)
	}
}

// TestRunEmpireTurnFreighterMaintenance 驗證 ps.ActiveFreighters>0 時每艘 0.5 BC 維護費從 NetBC
// 扣除(GAME_MANUAL.pdf p.169,gamedata.IncomeFreighterMaintenanceCost),並曝露在
// EmpireOutput.FreighterMaintenanceCost。
func TestRunEmpireTurnFreighterMaintenance(t *testing.T) {
	colonies := []ColonyState{
		{Population: 5, PopMax: 20, Workers: 2, IndustryPerWorker: 10,
			PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, TolerantRace: true},
	}
	// 稅收 10(20*50/100),維護費 3,運輸艦 5 艘 → 5*0.5=2.5 無條件捨去 → 2。
	ps := PlayerState{BC: 100, TaxRate: 50, Maintenance: 3, ActiveFreighters: 5, ResearchTopic: gamedata.ResearchTopic(1)}
	out := RunEmpireTurn(ps, colonies)

	if out.FreighterMaintenanceCost != 2 {
		t.Errorf("FreighterMaintenanceCost = %d,預期 2", out.FreighterMaintenanceCost)
	}
	if out.NetBC != 10 { // 15(工業稅10+人頭5) - 3(建築維護) - 2(運輸艦維護)
		t.Errorf("淨 BC = %d,預期 10", out.NetBC)
	}
	if out.Player.BC != 110 {
		t.Errorf("國庫 = %d,預期 110", out.Player.BC)
	}
}

// TestRunEmpireTurnFreighterMaintenanceZeroNoOp 驗證 ActiveFreighters=0(demo/remake 目前恆定
// 狀態,見該欄位註解)時完全不影響 NetBC——確保零值(no Freighter 塑模)與加欄位前行為一致。
func TestRunEmpireTurnFreighterMaintenanceZeroNoOp(t *testing.T) {
	colonies := []ColonyState{
		{Population: 5, PopMax: 20, Workers: 2, IndustryPerWorker: 10,
			PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT, TolerantRace: true},
	}
	ps := PlayerState{BC: 100, TaxRate: 50, Maintenance: 3, ResearchTopic: gamedata.ResearchTopic(1)} // ActiveFreighters 零值
	out := RunEmpireTurn(ps, colonies)

	if out.FreighterMaintenanceCost != 0 {
		t.Errorf("FreighterMaintenanceCost = %d,預期 0(無運輸艦塑模)", out.FreighterMaintenanceCost)
	}
	if out.NetBC != 12 { // 同 TestRunEmpireTurnBC(工業稅10+人頭5-3),確認接線未改變既有 no-op 行為
		t.Errorf("淨 BC = %d,預期 12", out.NetBC)
	}
}
