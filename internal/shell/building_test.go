package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestBuildingLongTermEffect 驗證建築完工後套用長期產出效果:自動工廠提升該殖民地工業/工人,
// 且效果只套一次(重複完工不疊加)。
func TestBuildingLongTermEffect(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true // 隔離隨機事件(富礦脈會改工業,干擾精確斷言)
	if len(s.PlayerColonies) == 0 {
		t.Fatal("需至少一個殖民地")
	}
	startIPW := s.PlayerColonies[0].IndustryPerWorker

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
	if afterIPW != startIPW+2 {
		t.Fatalf("自動工廠應使工業/工人 +2:%d → %d", startIPW, afterIPW)
	}

	// 再建一次自動工廠,不應再疊加效果。
	s.Builds[0] = ColonyBuild{Name: "自動工廠", Progress: 0, Cost: 60}
	for i := 0; i < 20 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}
	if s.PlayerColonies[0].IndustryPerWorker != afterIPW {
		t.Fatalf("重複建造不應再疊加效果:%d → %d", afterIPW, s.PlayerColonies[0].IndustryPerWorker)
	}
	t.Logf("自動工廠:工業/工人 %d→%d(+2,不重複疊加)", startIPW, afterIPW)
}

// TestResearchLabEffect 驗證研究實驗室提升研究/科學家 +5。
func TestResearchLabEffect(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	start := s.PlayerColonies[0].ResearchPerScientist
	s.Builds[0] = ColonyBuild{Name: "研究實驗室", Progress: 0, Cost: 60}
	for i := 0; i < 25 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}
	if got := s.PlayerColonies[0].ResearchPerScientist; got != start+5 {
		t.Fatalf("研究實驗室應使研究/科學家 +5:%d → %d", start, got)
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
