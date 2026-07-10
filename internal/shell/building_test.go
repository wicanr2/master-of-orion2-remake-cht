package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestBuildingLongTermEffect 驗證建築完工後套用長期產出效果:自動工廠提升該殖民地工業/工人,
// 且效果只套一次(重複完工不疊加)。
//
// 2026-07-11 訂正:自動工廠(Automated Factories,GAME_MANUAL.pdf p.78)手冊原文是「每工業
// 人口 +1 產能,殖民地整體另固定 +5 產能」,舊測試斷言 IndustryPerWorker +2 是把手冊的固定
// +5 揉進 per-worker 值湊出來的近似(小殖民地過度受益、大殖民地受益不足)。現在 engine 有
// FlatIndustry 欄位可以忠實分開建模,per-worker 訂正回手冊的 +1,固定值另計入 FlatIndustry。
func TestBuildingLongTermEffect(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true // 隔離隨機事件(富礦脈會改工業,干擾精確斷言)
	if len(s.PlayerColonies) == 0 {
		t.Fatal("需至少一個殖民地")
	}
	startIPW := s.PlayerColonies[0].IndustryPerWorker
	startFlat := s.PlayerColonies[0].FlatIndustry

	// 在殖民地 0 排自動工廠,給足工業直接完工。
	s.Builds[0] = ColonyBuild{Name: "自動工廠", Progress: 0, Cost: 60}
	// 先跑一回合產生 LastPlayerOutput(advanceBuilds 讀其淨工業)。
	s.EndTurn()
	// 多跑幾回合確保累積完工。
	for i := 0; i < 20 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}

	if !s.ColonyBuildings[0]["自動工廠"] {
		t.Fatal("自動工廠應標記為已建")
	}
	afterIPW := s.PlayerColonies[0].IndustryPerWorker
	afterFlat := s.PlayerColonies[0].FlatIndustry
	if afterIPW != startIPW+1 {
		t.Fatalf("自動工廠應使工業/工人 +1(p.78):%d → %d", startIPW, afterIPW)
	}
	if afterFlat != startFlat+5 {
		t.Fatalf("自動工廠應使殖民地固定工業 +5(p.78):%d → %d", startFlat, afterFlat)
	}

	// 再建一次自動工廠,不應再疊加效果。
	s.Builds[0] = ColonyBuild{Name: "自動工廠", Progress: 0, Cost: 60}
	for i := 0; i < 20 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}
	if s.PlayerColonies[0].IndustryPerWorker != afterIPW {
		t.Fatalf("重複建造不應再疊加效果:%d → %d", afterIPW, s.PlayerColonies[0].IndustryPerWorker)
	}
	if s.PlayerColonies[0].FlatIndustry != afterFlat {
		t.Fatalf("重複建造不應再疊加固定值:%d → %d", afterFlat, s.PlayerColonies[0].FlatIndustry)
	}
	t.Logf("自動工廠:工業/工人 %d→%d(+1),固定工業 %d→%d(+5),不重複疊加", startIPW, afterIPW, startFlat, afterFlat)
}

// TestResearchLabEffect 驗證研究實驗室提升研究/科學家 +1,並使殖民地固定研究 +5。
//
// 2026-07-11 訂正:研究實驗室(Research Laboratory,GAME_MANUAL.pdf p.94)手冊原文是「每科學家
// +1 研究點,另自動產生 5 研究點」,舊測試斷言 ResearchPerScientist +5 是把手冊的固定值錯當成
// per-worker 值。現在分開建模:per-worker 訂正回 +1,固定值計入 FlatResearch。
func TestResearchLabEffect(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	start := s.PlayerColonies[0].ResearchPerScientist
	startFlat := s.PlayerColonies[0].FlatResearch
	s.Builds[0] = ColonyBuild{Name: "研究實驗室", Progress: 0, Cost: 60}
	for i := 0; i < 25 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}
	if got := s.PlayerColonies[0].ResearchPerScientist; got != start+1 {
		t.Fatalf("研究實驗室應使研究/科學家 +1(p.94):%d → %d", start, got)
	}
	if got := s.PlayerColonies[0].FlatResearch; got != startFlat+5 {
		t.Fatalf("研究實驗室應使殖民地固定研究 +5(p.94):%d → %d", startFlat, got)
	}
}

// TestSpaceportIncomeBonusPercent 驗證太空港(Spaceport p.79「該殖民地所有來源 BC 收入 +50%」)
// 改用 IncomeBonusPercent 建模,不再誤動 IndustryPerWorker。
func TestSpaceportIncomeBonusPercent(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	startIPW := s.PlayerColonies[0].IndustryPerWorker
	startBonus := s.PlayerColonies[0].IncomeBonusPercent
	s.Builds[0] = ColonyBuild{Name: "太空港", Progress: 0, Cost: 60}
	for i := 0; i < 20 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}
	if got := s.PlayerColonies[0].IncomeBonusPercent; got != startBonus+50 {
		t.Fatalf("太空港應使 IncomeBonusPercent +50(p.79):%d → %d", startBonus, got)
	}
	if got := s.PlayerColonies[0].IndustryPerWorker; got != startIPW {
		t.Fatalf("太空港不應再誤動 IndustryPerWorker:%d → %d", startIPW, got)
	}
}

// TestPlanetaryStockExchangeIncomeBonusPercent 驗證行星證券交易所(p.93「該殖民地收入 +100%」)
// 與太空港的 IncomeBonusPercent 可疊加。
func TestPlanetaryStockExchangeIncomeBonusPercent(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.Builds[0] = ColonyBuild{Name: "太空港", Progress: 0, Cost: 60}
	for i := 0; i < 20 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}
	afterSpaceport := s.PlayerColonies[0].IncomeBonusPercent

	s.Builds[0] = ColonyBuild{Name: "行星證券交易所", Progress: 0, Cost: 60}
	for i := 0; i < 20 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}
	if got := s.PlayerColonies[0].IncomeBonusPercent; got != afterSpaceport+100 {
		t.Fatalf("太空港+證券交易所應疊加至 %d,實得 %d", afterSpaceport+100, got)
	}
}

// TestHydroponicFarmFlatFood 驗證水耕農場(p.99「殖民地食物產出 +2」)為固定值,與農夫數無關;
// 不再誤動 FoodPerFarmer。
func TestHydroponicFarmFlatFood(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	startFPF := s.PlayerColonies[0].FoodPerFarmer
	startFlat := s.PlayerColonies[0].FlatFood
	s.Builds[0] = ColonyBuild{Name: "水耕農場", Progress: 0, Cost: 60}
	for i := 0; i < 20 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}
	if got := s.PlayerColonies[0].FlatFood; got != startFlat+2 {
		t.Fatalf("水耕農場應使 FlatFood +2(p.99):%d → %d", startFlat, got)
	}
	if got := s.PlayerColonies[0].FoodPerFarmer; got != startFPF {
		t.Fatalf("水耕農場不應再誤動 FoodPerFarmer:%d → %d", startFPF, got)
	}
}

// TestBiospheresRaisesPopMax 驗證生態圈(p.99「星球人口上限 +2 單位」)直接疊加到 PopMax。
func TestBiospheresRaisesPopMax(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	startPopMax := s.PlayerColonies[0].PopMax
	s.Builds[0] = ColonyBuild{Name: "生態圈", Progress: 0, Cost: 60}
	for i := 0; i < 20 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}
	if got := s.PlayerColonies[0].PopMax; got != startPopMax+2 {
		t.Fatalf("生態圈應使 PopMax +2(p.99):%d → %d", startPopMax, got)
	}
}

// TestCloningCenterFlatGrowth 驗證複製中心(p.99「人口成長 +0.1 單位/回合」)換算成
// FlatGrowth = popGrowthThreshold/10。
func TestCloningCenterFlatGrowth(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.Builds[0] = ColonyBuild{Name: "複製中心", Progress: 0, Cost: 60}
	for i := 0; i < 20 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}
	if got := s.PlayerColonies[0].FlatGrowth; got != popGrowthThreshold/10 {
		t.Fatalf("複製中心應使 FlatGrowth = %d,實得 %d", popGrowthThreshold/10, got)
	}
}

// TestPlanetaryGravityGeneratorRecordedOnly 驗證行星重力產生器(p.104)目前只記錄
// NormalizeGravity 旗標,誠實反映「重力懲罰系統尚未接進生產管線」的現況(見
// engine.ColonyState.NormalizeGravity 欄位註解)。
func TestPlanetaryGravityGeneratorRecordedOnly(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.Builds[0] = ColonyBuild{Name: "行星重力產生器", Progress: 0, Cost: 60}
	for i := 0; i < 20 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}
	if !s.PlayerColonies[0].NormalizeGravity {
		t.Fatal("行星重力產生器應標記 NormalizeGravity=true")
	}
	if !s.ColonyBuildings[0]["行星重力產生器"] {
		t.Fatal("行星重力產生器應標記為已建")
	}
}

// TestTradeGoodsBuildOption 驗證「貿易品」建造佇列選項(見 session.go TradeGoodsBuildName):
// 設為貿易品的殖民地淨工業不累積建造進度,改以 gamedata.TradeGoodsIncome(一般種族 2:1)
// 換算成 BC 計入國庫。
func TestTradeGoodsBuildOption(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.Builds[0] = ColonyBuild{Name: TradeGoodsBuildName, Progress: 0, Cost: 0}

	beforeBC := s.Player.BC
	s.EndTurn()

	netIndustry := s.LastPlayerOutput.Colonies[0].NetIndustry
	if netIndustry <= 0 {
		t.Fatalf("測試前提:母星初始淨工業應 > 0 才能驗證貿易品收入,實得 %d", netIndustry)
	}
	wantRevenue := gamedata.TradeGoodsIncome(netIndustry, false)
	if s.LastPlayerOutput.TradeGoodsRevenue != wantRevenue {
		t.Fatalf("貿易品收入 = %d,預期 %d(NetIndustry=%d)",
			s.LastPlayerOutput.TradeGoodsRevenue, wantRevenue, netIndustry)
	}
	if s.Builds[0].Progress != 0 {
		t.Fatalf("貿易品不應累積建造進度,實得 %d", s.Builds[0].Progress)
	}
	if s.ColonyBuildings[0][TradeGoodsBuildName] {
		t.Fatal("貿易品不應被記為已建成建築")
	}
	wantBC := beforeBC + s.LastPlayerOutput.NetBC
	if s.Player.BC != wantBC {
		t.Fatalf("國庫 = %d,預期 %d(含貿易品收入 %d)", s.Player.BC, wantBC, wantRevenue)
	}
	t.Logf("貿易品:淨工業 %d → BC +%d(2:1),國庫 %d→%d", netIndustry, wantRevenue, beforeBC, s.Player.BC)
}

// TestTradeGoodsBuildOptionNonTradeColonyUnaffected 驗證非貿易品殖民地不受影響:建造項為一般
// 建築時,照常累積建造進度(貿易品旗標只在該殖民地建造項確實是貿易品時才生效)。
func TestTradeGoodsBuildOptionNonTradeColonyUnaffected(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.Builds[0] = ColonyBuild{Name: "自動工廠", Progress: 0, Cost: 60}
	s.EndTurn()

	if s.PlayerColonies[0].TradeGoods {
		t.Fatal("建造項為一般建築時,TradeGoods 旗標應同步為 false")
	}
	if s.Builds[0].Progress <= 0 {
		t.Fatalf("非貿易品殖民地應正常累積建造進度,實得 %d", s.Builds[0].Progress)
	}
	if s.LastPlayerOutput.TradeGoodsRevenue != 0 {
		t.Fatalf("非貿易品殖民地不應計入貿易品收入,實得 %d", s.LastPlayerOutput.TradeGoodsRevenue)
	}
}
