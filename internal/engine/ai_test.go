package engine

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestApplyAIEconomy(t *testing.T) {
	colonies := []ColonyState{
		{Population: 10, Farmers: 10, FoodPerFarmer: 5, IndustryPerWorker: 10,
			ResearchPerScientist: 5, PlanetSize: gamedata.MEDIUM_PLANET},
	}
	ps := PlayerState{BC: 500}
	d := ai.NewRemakeDecider(ai.ProfileScientific)
	ps2, cs2 := ApplyAIEconomy(ps, colonies, d)
	if cs2[0].Farmers != 2 || cs2[0].Workers != 2 || cs2[0].Scientists != 6 {
		t.Errorf("AI 分配錯誤:%+v", cs2[0])
	}
	if ps2.TaxRate != 10 {
		t.Errorf("稅率 = %d,預期 10", ps2.TaxRate)
	}
}

func TestRunAIEmpireTurn(t *testing.T) {
	colonies := []ColonyState{
		{Population: 10, Farmers: 10, FoodPerFarmer: 5, IndustryPerWorker: 10,
			ResearchPerScientist: 5, PlanetSize: gamedata.MEDIUM_PLANET},
	}
	ps := PlayerState{BC: 100, ResearchTopic: gamedata.ResearchTopic(1)}
	out := RunAIEmpireTurn(ps, colonies, ai.NewRemakeDecider(ai.ProfileScientific))
	if out.TotalResearch != 30 {
		t.Errorf("AI 回合研究 = %d,預期 30", out.TotalResearch)
	}
}
