package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// TestTerraformAdvancesClimateFoodAndPopMax 驗證地形改造(Terraforming)完工後:氣候沿階梯推進
// 一級、FoodPerFarmer 依手冊絕對值差值疊加、PopMax 依 pop_climate 係數等比例縮放。
func TestTerraformAdvancesClimateFoodAndPopMax(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	if len(s.PlayerColonies) == 0 {
		t.Fatal("需至少一個殖民地")
	}
	// 手動把母星氣候設成 Arid(手冊:Arid 地形改造 → Terran),對照組:FoodPerFarmer/PopMax
	// 同步改成與 Arid 相符的起始值,才能驗證「差值疊加」而非整批覆寫。
	c := &s.PlayerColonies[0]
	c.Climate = gamedata.ARID
	c.FoodPerFarmer = gamedata.ClimateFoodPerFarmer(gamedata.ARID) // = 1
	c.PopMax = 20                                                  // Arid 係數 60%

	// 解鎖地形改造前置科技(TOPIC_GENETIC_MUTATIONS)。
	if s.Player.CompletedTopics == nil {
		s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{}
	}
	s.Player.CompletedTopics[gamedata.TOPIC_GENETIC_MUTATIONS] = true

	action, ok := gamedata.SpecialActionByNameZH(gamedata.TerraformActionName)
	if !ok {
		t.Fatal("找不到地形改造的 SpecialAction 資料")
	}
	s.Builds[0] = ColonyBuild{Name: gamedata.TerraformActionName, Progress: 0, Cost: action.ProductionCost}
	s.EndTurn()
	for i := 0; i < 500 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}

	if s.PlayerColonies[0].Climate != gamedata.TERRAN {
		t.Fatalf("地形改造後氣候應推進到 Terran(手冊:Arid → Terran),got %v", s.PlayerColonies[0].Climate)
	}
	// FoodPerFarmer:1(Arid)+ (2-1) = 2(Terran)。
	if got, want := s.PlayerColonies[0].FoodPerFarmer, 2; got != want {
		t.Errorf("FoodPerFarmer 應為 %d,got %d", want, got)
	}
	// PopMax:20 * 80/60 = 26(向下取整)。
	if got, want := s.PlayerColonies[0].PopMax, 26; got != want {
		t.Errorf("PopMax 應為 %d(20*80/60 向下取整),got %d", want, got)
	}
	// Special 行動不記入 ColonyBuildings(可重複套用,見 advanceBuilds 註解)。
	if s.ColonyBuildings[0][gamedata.TerraformActionName] {
		t.Error("地形改造不應記入 ColonyBuildings(需可重複套用)")
	}
}

// TestTerraformNoOpWhenNoNextClimate 驗證地形改造在沒有下一級可推進的氣候(如已是 Terran)
// 套用時安全無效果,不 panic、不誤改狀態。
func TestTerraformNoOpWhenNoNextClimate(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	c := &s.PlayerColonies[0]
	c.Climate = gamedata.TERRAN // 母星預設本來就是 Terran
	startFood := c.FoodPerFarmer
	startPopMax := c.PopMax

	if s.Player.CompletedTopics == nil {
		s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{}
	}
	s.Player.CompletedTopics[gamedata.TOPIC_GENETIC_MUTATIONS] = true

	action, _ := gamedata.SpecialActionByNameZH(gamedata.TerraformActionName)
	s.Builds[0] = ColonyBuild{Name: gamedata.TerraformActionName, Progress: 0, Cost: action.ProductionCost}
	s.EndTurn()
	for i := 0; i < 500 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}

	if s.PlayerColonies[0].Climate != gamedata.TERRAN {
		t.Fatalf("Terran 上地形改造應無效果,氣候應維持 Terran,got %v", s.PlayerColonies[0].Climate)
	}
	if s.PlayerColonies[0].FoodPerFarmer != startFood || s.PlayerColonies[0].PopMax != startPopMax {
		t.Errorf("Terran 上地形改造應無效果:FoodPerFarmer %d→%d,PopMax %d→%d",
			startFood, s.PlayerColonies[0].FoodPerFarmer, startPopMax, s.PlayerColonies[0].PopMax)
	}
}

// TestGaiaTransformationRequiresTerran 驗證蓋亞轉化只在 Terran 星球生效,非 Terran 套用無效果。
func TestGaiaTransformationRequiresTerran(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	c := &s.PlayerColonies[0]
	c.Climate = gamedata.TERRAN
	c.FoodPerFarmer = gamedata.ClimateFoodPerFarmer(gamedata.TERRAN)
	c.PopMax = 40 // Terran 係數 80%

	if s.Player.CompletedTopics == nil {
		s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{}
	}
	s.Player.CompletedTopics[gamedata.TOPIC_TRANS_GENETICS] = true

	action, ok := gamedata.SpecialActionByNameZH(gamedata.GaiaTransformationActionName)
	if !ok {
		t.Fatal("找不到蓋亞轉化的 SpecialAction 資料")
	}
	s.Builds[0] = ColonyBuild{Name: gamedata.GaiaTransformationActionName, Progress: 0, Cost: action.ProductionCost}
	s.EndTurn()
	for i := 0; i < 500 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}

	if s.PlayerColonies[0].Climate != gamedata.GAIA {
		t.Fatalf("Terran 星球套用蓋亞轉化後應變成 Gaia,got %v", s.PlayerColonies[0].Climate)
	}
	if got, want := s.PlayerColonies[0].FoodPerFarmer, 3; got != want {
		t.Errorf("FoodPerFarmer 應為 %d(Gaia,手冊 p.59),got %d", want, got)
	}
	if got, want := s.PlayerColonies[0].PopMax, 50; got != want {
		t.Errorf("PopMax 應為 %d(40*100/80),got %d", want, got)
	}
}

// TestSoilEnrichmentBlockedOnHostileClimate 驗證土壤改良在 Barren/Radiated/Toxic 星球套用時
// 手冊規定的化學反應會抵銷肥沃化效果(誠實模擬「做了但沒用」,不是擋下建造選項)。
func TestSoilEnrichmentBlockedOnHostileClimate(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	c := &s.PlayerColonies[0]
	c.Climate = gamedata.BARREN
	startFood := c.FoodPerFarmer

	if s.Player.CompletedTopics == nil {
		s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{}
	}
	s.Player.CompletedTopics[gamedata.TOPIC_ADVANCED_BIOLOGY] = true

	action, _ := gamedata.SpecialActionByNameZH(gamedata.SoilEnrichmentActionName)
	s.Builds[0] = ColonyBuild{Name: gamedata.SoilEnrichmentActionName, Progress: 0, Cost: action.ProductionCost}
	s.EndTurn()
	for i := 0; i < 500 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}

	if s.PlayerColonies[0].FoodPerFarmer != startFood {
		t.Errorf("Barren 星球土壤改良應無效果,FoodPerFarmer 不應變動:%d→%d", startFood, s.PlayerColonies[0].FoodPerFarmer)
	}
}

// TestSoilEnrichmentWorksOnHospitableClimate 驗證土壤改良在一般氣候(非 Barren/Radiated/Toxic)
// 套用時,每個農夫食物產出 +1(手冊 p.99)。
func TestSoilEnrichmentWorksOnHospitableClimate(t *testing.T) {
	s := NewDemoSession()
	s.DisableEvents = true
	c := &s.PlayerColonies[0]
	c.Climate = gamedata.TERRAN
	startFood := c.FoodPerFarmer

	if s.Player.CompletedTopics == nil {
		s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{}
	}
	s.Player.CompletedTopics[gamedata.TOPIC_ADVANCED_BIOLOGY] = true

	action, _ := gamedata.SpecialActionByNameZH(gamedata.SoilEnrichmentActionName)
	s.Builds[0] = ColonyBuild{Name: gamedata.SoilEnrichmentActionName, Progress: 0, Cost: action.ProductionCost}
	s.EndTurn()
	for i := 0; i < 500 && s.Builds[0].Name != ""; i++ {
		s.EndTurn()
	}

	if got, want := s.PlayerColonies[0].FoodPerFarmer, startFood+gamedata.TerraformSoilEnrichmentFoodBonusPerFarmer; got != want {
		t.Errorf("Terran 星球土壤改良應使 FoodPerFarmer +%d:%d→%d(got %d)",
			gamedata.TerraformSoilEnrichmentFoodBonusPerFarmer, startFood, want, got)
	}
}
