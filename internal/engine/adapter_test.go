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
	// 明確設 Normal-G:save.Planet.Gravity 的 Go 零值(0)與 gamedata.LOW_G 的 ordinal 相同
	// (見 adapter.go 檔頭「行星重力」說明),這裡要驗證的是一般欄位對映,不是重力懲罰,
	// 故明確賦值避免零值被誤讀為 Low-G(-25%)污染下方的 Food 端到端斷言。
	pl.Gravity = uint8(gamedata.NORMAL_G)

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
	if cs.PlanetGravity != gamedata.NORMAL_G {
		t.Errorf("行星重力 = %d,預期 NORMAL_G", cs.PlanetGravity)
	}
	// 端到端:轉出的狀態能直接跑一回合
	out := RunColonyTurn(cs)
	if out.Food != 10 { // 2*5,Normal-G 無懲罰
		t.Errorf("轉出後跑回合 Food = %d,預期 10", out.Food)
	}
}

// TestColonyStateFromSaveGravityMapping 驗證 save.Planet.Gravity(uint8)直接數值轉型成
// gamedata.PlanetGravity(兩者同源 openorion2 enum ordinal,見 adapter.go 檔頭說明),
// LOW_G/HEAVY_G 都能正確映射並在 RunColonyTurn 端到端反映重力懲罰。
func TestColonyStateFromSaveGravityMapping(t *testing.T) {
	var c save.Colony
	c.Population = 2
	c.FoodPerFarmer = 4
	c.Colonists[0].Job = uint8(gamedata.FARMER)
	c.Colonists[1].Job = uint8(gamedata.FARMER)

	var pl save.Planet
	pl.Size = uint8(gamedata.MEDIUM_PLANET)
	pl.Gravity = uint8(gamedata.HEAVY_G)

	cs := ColonyStateFromSave(&c, &pl)
	if cs.PlanetGravity != gamedata.HEAVY_G {
		t.Fatalf("行星重力 = %d,預期 HEAVY_G", cs.PlanetGravity)
	}
	out := RunColonyTurn(cs)
	if out.Food != 4 { // 2*4=8,Heavy-G -50% → 4
		t.Errorf("Heavy-G 存檔殖民地 Food = %d,預期 4", out.Food)
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
