package engine

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestRunEmpireTurn(t *testing.T) {
	// 兩個殖民地產出 400，既有進度 1，累積 401 形成最低 1% 突破率；
	// RunEmpireTurn 的固定測試擲骰為 1，因此完成。
	colonies := []ColonyState{
		{Population: 10, PopMax: 20, Farmers: 4, Workers: 4, Scientists: 2,
			FoodPerFarmer: 3, IndustryPerWorker: 5, ResearchPerScientist: 100,
			PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT}, // 研究 200
		{Population: 8, PopMax: 20, Farmers: 3, Workers: 3, Scientists: 2,
			FoodPerFarmer: 3, IndustryPerWorker: 5, ResearchPerScientist: 100,
			PlanetSize: gamedata.SMALL_PLANET, PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT}, // 研究 200
	}
	ps := PlayerState{ResearchTopic: gamedata.ResearchTopic(1), ResearchProgress: 1} // cost 400
	out := RunEmpireTurn(ps, colonies)

	if len(out.Colonies) != 2 {
		t.Fatalf("殖民地輸出數 = %d,預期 2", len(out.Colonies))
	}
	if out.TotalResearch != 400 { // 200+200
		t.Errorf("總研究 = %d,預期 400", out.TotalResearch)
	}
	if !out.ResearchDone {
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
	// 驗證研究進度跨回合累積，並在嚴格超過成本後依固定最低擲骰突破。
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
	// 回合1:150、回合2:300、回合3:450>400 → 12% 突破率，固定 roll=1 成功並清零。
	if completedTurn != 3 {
		t.Errorf("研究應於第 3 回合完成,實際第 %d 回合", completedTurn)
	}
	if !ps.CompletedTopics[gamedata.ResearchTopic(1)] {
		t.Error("完成後 topic 1 應標記")
	}
	if ps.ResearchProgress != 0 {
		t.Errorf("突破後研究進度 = %d,預期清零", ps.ResearchProgress)
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

func TestRunEmpireTurnSpyAndOfficerMaintenance(t *testing.T) {
	ps := PlayerState{BC: 50, SpyMaintenance: 2, OfficerMaintenance: 3}
	out := RunEmpireTurn(ps, nil)
	if out.SpyMaintenanceCost != 2 || out.OfficerMaintenanceCost != 3 {
		t.Fatalf("維護費分項 = spy %d / officer %d，預期 2 / 3",
			out.SpyMaintenanceCost, out.OfficerMaintenanceCost)
	}
	if out.NetBC != -5 || out.Player.BC != 45 {
		t.Fatalf("維護費應在單次帝國結算扣除：NetBC=%d BC=%d，預期 -5 / 45", out.NetBC, out.Player.BC)
	}
}

// TestRunEmpireTurnCyberneticHalfIncome 驗證帝國層收入讀精確半單位欄位，而不是先讀整數
// NetIndustry/FoodSurplus 再把 22.5/7.5 截斷。
func TestRunEmpireTurnCyberneticHalfIncome(t *testing.T) {
	colonies := []ColonyState{{
		Population: 5, PopMax: 20, Farmers: 5, Workers: 5,
		FoodPerFarmer: 2, IndustryPerWorker: 5,
		PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G,
		MineralRichness: gamedata.ABUNDANT, TolerantRace: true,
		Cybernetic: true, TradeGoods: true,
	}}
	out := RunEmpireTurn(PlayerState{TaxRate: 50, FantasticTrader: true}, colonies)
	// 淨工業 45/2: 稅 11 BC、貿易品 22 BC；餘糧 15/2: Fantastic Trader 7 BC，
	// 另有人頭基礎收入 5 BC。
	if out.TotalFoodHalf != 15 || out.TotalNetIndustryHalf != 45 {
		t.Fatalf("帝國半單位聚合錯誤: food=%d industry=%d", out.TotalFoodHalf, out.TotalNetIndustryHalf)
	}
	if out.TaxRevenue != 16 || out.FoodSurplusRevenue != 7 || out.TradeGoodsRevenue != 22 || out.NetBC != 45 {
		t.Errorf("半單位收入錯誤:%+v", out)
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

func TestRunEmpireTurnAICommandDeficitOverride(t *testing.T) {
	ps := PlayerState{BC: 100, CommandPointsSupply: 2, UsedCommandPoints: 5,
		CommandOverflowCostPerPoint: 9}
	out := RunEmpireTurn(ps, nil)
	if out.CommandOverflowCost != 27 || out.NetBC != -27 || out.Player.BC != 73 {
		t.Fatalf("Hard AI 指揮赤字成本應為 3×9：%+v", out)
	}
	ps.CommandPointsSupply = 5
	out = RunEmpireTurn(ps, nil)
	if out.CommandOverflowCost != 0 {
		t.Fatalf("供需相等不得扣指揮赤字：%+v", out)
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
	// 一個在途 settler 佔用五艘，故 +0x38<=0 時五艘全數計維護費。
	ps := PlayerState{BC: 100, TaxRate: 50, Maintenance: 3, ActiveFreighters: 5, SettlersFreighted: 1, ResearchTopic: gamedata.ResearchTopic(1)}
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

func TestRunEmpireTurnIdleFreightersHaveNoMaintenance(t *testing.T) {
	out := RunEmpireTurn(PlayerState{ActiveFreighters: 5}, []ColonyState{{Population: 1, Farmers: 1, FoodPerFarmer: 2}})
	if out.Player.SurplusFreighters != 5 || out.FreighterMaintenanceCost != 0 {
		t.Fatalf("未使用貨運艦不應收維護費：surplus=%d cost=%d", out.Player.SurplusFreighters, out.FreighterMaintenanceCost)
	}
}

// TestRunEmpireTurnFreighterMaintenanceZeroNoOp 驗證 ActiveFreighters=0 時完全不影響 NetBC，
// 確保零值與加欄位前行為一致。
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

// TestRunEmpireTurnSpecialIncome 驗證行星特殊物產的固定收入(寶石礦 +10 / 金礦 +5 BC/回合,
// 手冊逐字)真的進到帝國收入,且與人口無關——同一顆寶石礦星,人口 5 與人口 1 拿到的一樣多。
func TestRunEmpireTurnSpecialIncome(t *testing.T) {
	base := ColonyState{
		Population: 5, PopMax: 20, Workers: 2, IndustryPerWorker: 10,
		PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G,
		MineralRichness: gamedata.ABUNDANT, TolerantRace: true,
	}
	ps := PlayerState{BC: 100, ResearchTopic: gamedata.ResearchTopic(1)}

	plain := RunEmpireTurn(ps, []ColonyState{base})
	gems := base
	gems.SpecialIncome = 10
	withGems := RunEmpireTurn(ps, []ColonyState{gems})
	if got := withGems.TaxRevenue - plain.TaxRevenue; got != 10 {
		t.Errorf("寶石礦收入 = %d,預期 10", got)
	}

	// 與人口無關:人口砍到 1,特殊物產收入不變(對照人頭收入會少 4)。
	small := gems
	small.Population, small.Workers = 1, 1
	plainSmall := base
	plainSmall.Population, plainSmall.Workers = 1, 1
	if got := RunEmpireTurn(ps, []ColonyState{small}).TaxRevenue -
		RunEmpireTurn(ps, []ColonyState{plainSmall}).TaxRevenue; got != 10 {
		t.Errorf("人口 1 時寶石礦收入 = %d,預期仍是 10(固定收入不隨人口縮放)", got)
	}
}

// TestRunEmpireTurnSpecialIncomeTakesBuildingBonus 驗證特殊物產收入一併受
// IncomeBonusPercent(太空港/證交所)放大——手冊寫的是「該殖民地**所有來源** BC 收入 +N%」,
// 寶石礦收入是該殖民地的一個 BC 來源,不排除在外。
func TestRunEmpireTurnSpecialIncomeTakesBuildingBonus(t *testing.T) {
	cs := ColonyState{
		Population: 5, PopMax: 20, Workers: 2, IndustryPerWorker: 10,
		PlanetSize: gamedata.TINY_PLANET, PlanetGravity: gamedata.NORMAL_G,
		MineralRichness: gamedata.ABUNDANT, TolerantRace: true,
		SpecialIncome: 10, IncomeBonusPercent: 100,
	}
	ps := PlayerState{BC: 100, ResearchTopic: gamedata.ResearchTopic(1)}
	out := RunEmpireTurn(ps, []ColonyState{cs})
	// 人頭 5 + 寶石 10 = 15,×200% = 30
	if out.TaxRevenue != 30 {
		t.Errorf("money 收入 = %d,預期 30((人頭5+寶石10)×200%%)", out.TaxRevenue)
	}
}

// TestGalacticCurrencyExchangeBoostsAllIncome 釘住手冊那條:
// 「increases the income generated by all colonies (from all sources) by 50%」。
//
// ⚠ 這棟先前被歸類錯:原版建築表裡它有成本與維護費,而 gap report 第 11 項(48棟建築盤點)用
// 「維護費 0 = 一次性」這條**自訂啟發式**把它判成常駐建築,於是「效果不明」擱置了很久。
// 手冊直接寫明它是 Achievement 科技與確切效果——**一手來源贏自訂推論規則**。
func TestGalacticCurrencyExchangeBoostsAllIncome(t *testing.T) {
	colonies := []ColonyState{{
		Population: 6, Farmers: 2, Workers: 2, Scientists: 2,
	}}
	base := PlayerState{TaxRate: 50}
	withEx := base
	withEx.HasGalacticCurrencyExchange = true

	a := RunEmpireTurn(base, colonies)
	b := RunEmpireTurn(withEx, colonies)

	subA := a.TaxRevenue + a.FoodSurplusRevenue + a.TradeGoodsRevenue
	subB := b.TaxRevenue + b.FoodSurplusRevenue + b.TradeGoodsRevenue
	if subA <= 0 {
		t.Skip("這個殖民地本來就沒有收入,驗不出加成")
	}
	want := subA + subA*50/100
	if subB != want {
		t.Errorf("有銀河貨幣交易所時收入應為 %d(%d 的 +50%%),實得 %d", want, subA, subB)
	}
}

// TestGalacticCurrencyExchangeOffByDefault:沒研究就完全沒有影響(零值 false 安全)。
func TestGalacticCurrencyExchangeOffByDefault(t *testing.T) {
	colonies := []ColonyState{{Population: 5, Workers: 5}}
	ps := PlayerState{TaxRate: 50}
	if ps.HasGalacticCurrencyExchange {
		t.Fatal("零值應為 false")
	}
	out := RunEmpireTurn(ps, colonies)
	ps2 := ps
	out2 := RunEmpireTurn(ps2, colonies)
	if out.TaxRevenue != out2.TaxRevenue {
		t.Error("同樣的輸入應得到同樣的收入")
	}
}

// 兩回合各換出半食物時，第一回合只累積半 BC，第二回合才扣掉 1 BC。
func TestRunEmpireTurnFoodReplicatorHalfBCCarries(t *testing.T) {
	cs := ColonyState{
		Population: 3, PopMax: 10, Farmers: 1, Workers: 1,
		FoodPerFarmer: 1, IndustryPerWorker: 4, PlanetSize: 2,
		PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
		Cybernetic: true, FoodReplicators: true,
	}
	ps := PlayerState{}
	first := RunEmpireTurn(ps, []ColonyState{cs})
	if first.FoodReplicatorCostHalfBC != 1 || first.FoodReplicatorCost != 0 || first.Player.FoodReplicatorBCHalfRemainder != 1 {
		t.Fatalf("第一回合應保存半 BC,got %+v", first)
	}
	second := RunEmpireTurn(first.Player, []ColonyState{cs})
	if second.FoodReplicatorCostHalfBC != 1 || second.FoodReplicatorCost != 1 || second.Player.FoodReplicatorBCHalfRemainder != 0 {
		t.Fatalf("第二回合應合併成 1 BC,got %+v", second)
	}
}

func TestRunEmpireTurnDivertedColonyResearchStaysInOutputButNotEmpireTotal(t *testing.T) {
	base := ColonyState{
		Population: 1, Scientists: 1, ResearchPerScientist: 4,
		PlanetGravity: gamedata.NORMAL_G, MineralRichness: gamedata.ABUNDANT,
	}
	diverted := base
	diverted.ResearchDiverted = true
	ps := PlayerState{ResearchTopic: gamedata.ResearchTopic(1)}

	out := RunEmpireTurnWithResearchRoller(ps, []ColonyState{base, diverted}, nil)
	if out.Colonies[0].Research <= 0 || out.Colonies[1].Research != out.Colonies[0].Research {
		t.Fatalf("兩座殖民地輸出應保留相同 RP：%+v", out.Colonies)
	}
	if got, want := out.TotalResearch, out.Colonies[0].Research; got != want {
		t.Fatalf("帝國一般研究只應收到未轉用殖民地 RP：got %d want %d", got, want)
	}
	if out.Player.ResearchProgress != out.TotalResearch {
		t.Fatalf("研究進度應只增加一般研究總量：progress=%d total=%d", out.Player.ResearchProgress, out.TotalResearch)
	}
}

func TestColonyResearchDivertedIsEphemeralAcrossJSON(t *testing.T) {
	want := ColonyState{ResearchDiverted: true}
	b, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(b, []byte("ResearchDiverted")) || bytes.Contains(b, []byte("researchDiverted")) {
		t.Fatalf("暫態事件輸入不得寫入存檔／多人快照：%s", b)
	}
	var got ColonyState
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.ResearchDiverted {
		t.Fatal("載入後應由 PersistentEvents 重建，不得沿用 stale 暫態旗標")
	}
}

func TestAIDifficultyEconomyQuarterBonuses(t *testing.T) {
	base := ColonyState{
		Population: 4, PopMax: 10, Farmers: 1, Workers: 1, Scientists: 2,
		FoodPerFarmer: 2, IndustryPerWorker: 4, ResearchPerScientist: 3,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G,
		RaceGravity: gamedata.NORMAL_G, RaceGravityKnown: true,
	}
	plain := RunEmpireTurn(PlayerState{}, []ColonyState{base})
	hardColony := base
	hardColony.AIDifficultyFoodQuarters = 2
	hardColony.AIDifficultyIndustryQuarters = 4
	hardColony.AIDifficultyResearchQuarters = 4
	hardColony.GrowthBonusSum = 3
	hard := RunEmpireTurn(PlayerState{AIDifficultyIncomeQuartersPerPop: 2}, []ColonyState{hardColony})
	if hard.Colonies[0].GrossIndustry != plain.Colonies[0].GrossIndustry+1 {
		t.Fatalf("Hard 工業難度加值錯誤：plain=%d hard=%d", plain.Colonies[0].GrossIndustry, hard.Colonies[0].GrossIndustry)
	}
	if hard.Colonies[0].Research != plain.Colonies[0].Research+2 {
		t.Fatalf("Hard 研究難度加值錯誤：plain=%d hard=%d", plain.Colonies[0].Research, hard.Colonies[0].Research)
	}
	if hard.TaxRevenue != plain.TaxRevenue+2 {
		t.Fatalf("Hard BC 難度加值錯誤：plain=%d hard=%d", plain.TaxRevenue, hard.TaxRevenue)
	}

	tutor := base
	tutor.AIDifficultyFoodQuarters = -1
	tutorOut := RunColonyTurn(tutor)
	plainColony := RunColonyTurn(base)
	if tutorOut.Food != plainColony.Food-1 {
		t.Fatalf("Tutor 單一農夫 -1/4 必須向下取整：plain=%d tutor=%d", plainColony.Food, tutorOut.Food)
	}
}
