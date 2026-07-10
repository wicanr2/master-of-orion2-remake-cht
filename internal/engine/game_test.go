package engine

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/save"
)

func TestRunGameTurn(t *testing.T) {
	var gs save.GameState
	// 兩顆行星:0=MEDIUM、1=SMALL
	gs.Planets = make([]save.Planet, 2)
	gs.Planets[0].Size = uint8(gamedata.MEDIUM_PLANET)
	gs.Planets[0].Gravity = uint8(gamedata.NORMAL_G) // 明確設 Normal-G,避免零值=LOW_G 混入本測試(見 adapter.go 檔頭說明)
	gs.Planets[1].Size = uint8(gamedata.SMALL_PLANET)
	gs.Planets[1].Gravity = uint8(gamedata.NORMAL_G)
	gs.PlanetCount = 2

	// 兩個殖民地:colony0 屬玩家0(行星0),colony1 屬玩家1(行星1)
	gs.Colonies = make([]save.Colony, 2)
	gs.Colonies[0].Owner = 0
	gs.Colonies[0].Planet = 0
	gs.Colonies[0].Population = 5
	gs.Colonies[0].IndustryPerWorker = 10
	gs.Colonies[0].Colonists[0].Job = uint8(gamedata.WORKER)
	gs.Colonies[0].Colonists[1].Job = uint8(gamedata.WORKER) // 2 工人
	gs.Colonies[1].Owner = 1
	gs.Colonies[1].Planet = 1
	gs.Colonies[1].Population = 4
	gs.Colonies[1].IndustryPerWorker = 10
	gs.Colonies[1].Colonists[0].Job = uint8(gamedata.WORKER) // 1 工人
	gs.ColonyCount = 2

	// 兩位玩家
	gs.Players = make([]save.Player, 2)
	gs.Players[0].TaxRate = 50
	gs.Players[0].BC = 100
	gs.Players[0].ResearchTopic = 1
	gs.Players[1].TaxRate = 0
	gs.Players[1].BC = 50
	gs.Players[1].ResearchTopic = 1
	gs.PlayerCount = 2

	res := RunGameTurn(&gs)

	if len(res.PlayerOutputs) != 2 {
		t.Fatalf("玩家結果數 = %d,預期 2", len(res.PlayerOutputs))
	}
	// 玩家0:1 殖民地,毛工業 2*10=20,MEDIUM 容忍6,清理(20-6)/2=7,淨13,稅 13*50/100=6
	p0 := res.PlayerOutputs[0]
	if len(p0.Colonies) != 1 {
		t.Errorf("玩家0 殖民地數 = %d,預期 1", len(p0.Colonies))
	}
	if p0.TotalNetIndustry != 13 {
		t.Errorf("玩家0 淨工業 = %d,預期 13", p0.TotalNetIndustry)
	}
	if p0.TaxRevenue != 6 {
		t.Errorf("玩家0 稅收 = %d,預期 6", p0.TaxRevenue)
	}
	// 玩家1:稅率0 → 稅收0
	p1 := res.PlayerOutputs[1]
	if len(p1.Colonies) != 1 || p1.TaxRevenue != 0 {
		t.Errorf("玩家1 結果錯誤:colonies=%d tax=%d", len(p1.Colonies), p1.TaxRevenue)
	}
}
