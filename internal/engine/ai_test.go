package engine

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestApplyAIEconomy(t *testing.T) {
	// 初始工作分配「錯誤」(全農夫),AI(科學傾向)應重分配。
	colonies := []ColonyState{
		{Population: 10, Farmers: 10, FoodPerFarmer: 5, IndustryPerWorker: 10,
			ResearchPerScientist: 5, PlanetSize: gamedata.MEDIUM_PLANET},
	}
	ps := PlayerState{BC: 500} // 充裕 → 稅率應降到 10
	ps2, cs2 := ApplyAIEconomy(ps, colonies, ai.ProfileScientific)
	// DecideColonyJobs(10,5,科學) = 農2 工2 研6
	if cs2[0].Farmers != 2 || cs2[0].Workers != 2 || cs2[0].Scientists != 6 {
		t.Errorf("AI 分配錯誤:%+v", cs2[0])
	}
	if ps2.TaxRate != 10 { // BC 500 > high 300
		t.Errorf("稅率 = %d,預期 10(國庫充裕)", ps2.TaxRate)
	}
}

func TestRunAIEmpireTurn(t *testing.T) {
	colonies := []ColonyState{
		{Population: 10, Farmers: 10, FoodPerFarmer: 5, IndustryPerWorker: 10,
			ResearchPerScientist: 5, PlanetSize: gamedata.MEDIUM_PLANET},
	}
	ps := PlayerState{BC: 100, ResearchTopic: gamedata.ResearchTopic(1)}
	out := RunAIEmpireTurn(ps, colonies, ai.ProfileScientific)
	// 重分配後:農2(食10,消耗10,盈餘0)、研6*5=30
	if out.TotalResearch != 30 {
		t.Errorf("AI 回合研究 = %d,預期 30", out.TotalResearch)
	}
	if out.Colonies[0].FoodSurplus != 0 {
		t.Errorf("食物盈餘 = %d,預期 0(剛好餵飽)", out.Colonies[0].FoodSurplus)
	}
}
