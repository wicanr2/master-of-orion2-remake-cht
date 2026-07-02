package engine

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/save"
)

func TestColonyStateFromSave(t *testing.T) {
	var c save.Colony
	c.Population = 4
	c.MaxPopulation = 20
	c.FoodPerFarmer = 5
	c.IndustryPerWorker = 6
	c.ResearchPerScientist = 3
	// 工作分配:2 農夫、1 工人、1 科學家
	c.Colonists[0].Job = uint8(gamedata.FARMER)
	c.Colonists[1].Job = uint8(gamedata.FARMER)
	c.Colonists[2].Job = uint8(gamedata.WORKER)
	c.Colonists[3].Job = uint8(gamedata.SCIENTIST)

	var pl save.Planet
	pl.Size = uint8(gamedata.MEDIUM_PLANET)

	cs := ColonyStateFromSave(&c, &pl)
	if cs.Farmers != 2 || cs.Workers != 1 || cs.Scientists != 1 {
		t.Errorf("工作分配錯誤:%+v", cs)
	}
	if cs.Population != 4 || cs.PopMax != 20 || cs.FoodPerFarmer != 5 ||
		cs.IndustryPerWorker != 6 || cs.ResearchPerScientist != 3 {
		t.Errorf("欄位對映錯誤:%+v", cs)
	}
	if cs.PlanetSize != gamedata.MEDIUM_PLANET {
		t.Errorf("行星尺寸 = %d,預期 MEDIUM", cs.PlanetSize)
	}
	// 端到端:轉出的狀態能直接跑一回合
	out := RunColonyTurn(cs)
	if out.Food != 10 { // 2*5
		t.Errorf("轉出後跑回合 Food = %d,預期 10", out.Food)
	}
}

func TestPlayerStateFromSave(t *testing.T) {
	var p save.Player
	p.BC = 500
	p.TaxRate = 45
	p.ResearchTopic = 1
	p.ResearchProgress = 200
	ps := PlayerStateFromSave(&p)
	if ps.BC != 500 || ps.TaxRate != 45 || ps.ResearchTopic != gamedata.ResearchTopic(1) ||
		ps.ResearchProgress != 200 {
		t.Errorf("Player 對映錯誤:%+v", ps)
	}
}
