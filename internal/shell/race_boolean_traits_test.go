package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// 五個布林特性各自落在正確的種族身上(對一手表,不是對記憶)。
func TestBooleanTraitsLandOnTheRightRaces(t *testing.T) {
	cases := []struct {
		en   string
		has  func(*GameSession) bool
		name string
	}{
		{"Mrrshan", (*GameSession).RaceWarlord, "統帥"},
		{"Silicoids", (*GameSession).RaceRepulsive, "惹人厭"},
		{"Silicoids", (*GameSession).RaceTolerant, "寬容"},
		{"Humans", (*GameSession).RaceCharismatic, "魅力"},
		{"Gnolams", (*GameSession).RaceFantasticTrader, "神級商人"},
	}
	for _, c := range cases {
		s := NewDemoSession()
		s.ApplyRace(raceIndexByEnName(t, c.en))
		if !c.has(s) {
			t.Errorf("%s 應有「%s」特性", c.en, c.name)
		}
		// 反面:換一族就不該有(否則「回傳 true」也會通過)。
		other := NewDemoSession()
		other.ApplyRace(raceIndexByEnName(t, "Psilons")) // 席隆這五項都沒有
		if c.has(other) {
			t.Errorf("席隆不該有「%s」特性", c.name)
		}
	}
}

// 寬容(矽基)真的省下清污染的產能——`TolerantRace` 先前**沒有任何寫入端**。
func TestTolerantRaceSkipsPollutionCleanup(t *testing.T) {
	base := engine.ColonyState{
		Population: 10, PopMax: 14, Farmers: 2, Workers: 8,
		FoodPerFarmer: 4, IndustryPerWorker: 5, ResearchPerScientist: 2,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G,
		MineralRichness: gamedata.ABUNDANT,
	}
	dirty := engine.RunColonyTurn(base)
	if dirty.PollutionCleanupCost == 0 {
		t.Fatalf("測試前提不成立:一般種族應有清理成本,得到 %d", dirty.PollutionCleanupCost)
	}
	tol := base
	tol.TolerantRace = true
	if clean := engine.RunColonyTurn(tol); clean.PollutionCleanupCost != 0 {
		t.Errorf("寬容種族不該花產能清污染,得到 %d", clean.PollutionCleanupCost)
	}
}

// 回合結算真的把寬容同步進殖民地(先前這個欄位是死的)。
func TestEndTurnSyncsTolerantIntoColonies(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.ApplyRace(raceIndexByEnName(t, "Silicoids"))
	s.EndTurn()
	if !s.PlayerColonies[0].TolerantRace {
		t.Error("回合結算應把寬容同步進殖民地")
	}
	// 反面:一般種族不該被設起來。
	h := NewDemoSession()
	h.DisableEvents = true
	h.ApplyRace(raceIndexByEnName(t, "Humans"))
	h.EndTurn()
	if h.PlayerColonies[0].TolerantRace {
		t.Error("人類不該被標成寬容種族")
	}
}

// 神級商人(諾蘭姆)讓貿易品與餘糧收入翻倍。
//
// 兩處先前都硬傳 `false`,註解寫著「ColonyState 目前沒有追蹤這個種族特質的欄位」。
func TestFantasticTraderDoublesTradeAndFoodRevenue(t *testing.T) {
	cs := engine.ColonyState{
		Population: 10, PopMax: 14, Farmers: 6, Workers: 4,
		FoodPerFarmer: 5, IndustryPerWorker: 4, ResearchPerScientist: 2,
		PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G,
		MineralRichness: gamedata.ABUNDANT,
		TradeGoods:      true,
	}
	probe := engine.RunColonyTurn(cs)
	if probe.FoodSurplus <= 0 || probe.NetIndustry <= 0 {
		t.Fatalf("測試前提不成立:需要正的餘糧(%d)與淨工業(%d)", probe.FoodSurplus, probe.NetIndustry)
	}

	plain := engine.RunEmpireTurn(engine.PlayerState{TaxRate: 50}, []engine.ColonyState{cs})
	rich := engine.RunEmpireTurn(engine.PlayerState{TaxRate: 50, FantasticTrader: true}, []engine.ColonyState{cs})
	if rich.NetBC <= plain.NetBC {
		t.Errorf("神級商人的帝國淨收入應較高:%d vs %d", rich.NetBC, plain.NetBC)
	}

	// 逐項確認,免得只是某一項變了而另一項沒接。
	if a, b := gamedata.TradeGoodsIncome(probe.NetIndustry, false),
		gamedata.TradeGoodsIncome(probe.NetIndustry, true); b <= a {
		t.Errorf("貿易品應由 2:1 變 1:1:%d → %d", a, b)
	}
	if a, b := gamedata.IncomeFoodSurplusRevenue(probe.FoodSurplus, false),
		gamedata.IncomeFoodSurplusRevenue(probe.FoodSurplus, true); b <= a {
		t.Errorf("餘糧收入應由 0.5 變 1 BC/單位:%d → %d", a, b)
	}
}

// 回合結算把神級商人同步進 PlayerState(跨層那一步)。
func TestEndTurnSyncsFantasticTraderIntoPlayerState(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.ApplyRace(raceIndexByEnName(t, "Gnolams"))
	s.EndTurn()
	if !s.Player.FantasticTrader {
		t.Error("回合結算應把神級商人同步進 PlayerState")
	}
}

// 統帥(姆瑞森)讓營房容量加倍。
func TestWarlordDoublesBarracksCapacity(t *testing.T) {
	const pop, popMax = 12, 16
	plain := gamedata.GroundMarineBarracksCap(pop, popMax, false)
	warlord := gamedata.GroundMarineBarracksCap(pop, popMax, true)
	if warlord != plain*gamedata.GroundWarlordBarracksMultiplier {
		t.Fatalf("統帥營房容量應為 %d 倍:%d vs %d",
			gamedata.GroundWarlordBarracksMultiplier, warlord, plain)
	}
	// 而且 shell 這一側真的傳得進去。
	s := NewDemoSession()
	s.ApplyRace(raceIndexByEnName(t, "Mrrshan"))
	if !s.RaceWarlord() {
		t.Fatal("姆瑞森應是統帥種族")
	}
}

// 種族布林特性是**算出來的**,不存檔——存讀往返不該改變任何一項。
//
// ⚠ 這一條是踩過才加的:早期版本把種族編號與五個旗標存進 GameSession 與存檔,
// 結果舊存檔解出零值(編號 0 = 阿爾卡里)整個查錯族,存讀指紋也對不上。
func TestBooleanTraitsSurviveSaveLoad(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	s.ApplyRace(raceIndexByEnName(t, "Silicoids"))
	s.EndTurn()

	got := s.snapshot().restore()
	if !got.RaceTolerant() || !got.RaceRepulsive() {
		t.Errorf("讀檔後矽基應仍是寬容 + 惹人厭,得到 寬容=%v 惹人厭=%v",
			got.RaceTolerant(), got.RaceRepulsive())
	}
	if got.raceOrigIdx() != s.raceOrigIdx() {
		t.Errorf("種族編號應一致:%d vs %d", got.raceOrigIdx(), s.raceOrigIdx())
	}
}
