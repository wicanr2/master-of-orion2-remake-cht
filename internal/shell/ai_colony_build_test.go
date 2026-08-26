package shell

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestOriginalAIExactBuildingScores(t *testing.T) {
	colony := engine.ColonyState{Population: 6}
	tests := []struct {
		name      string
		balanced  int
		honorable int
	}{
		{"太空大學", 5, 5},
		{"自動工廠", 19, 21},
		{"深層核心礦場", 18, 22},
		{"機器人工廠", 12, 14},
		{"機器人採礦廠", 11, 13},
	}
	for _, tt := range tests {
		b, ok := gamedata.BuildingByNameZH(tt.name)
		if !ok {
			t.Fatalf("測試建築不存在：%s", tt.name)
		}
		got, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityXenophobic, originalAIBuildScoreContext{})
		if !exact || got != tt.balanced {
			t.Errorf("%s 一般性格分數=(%d,%v)，want (%d,true)", tt.name, got, exact, tt.balanced)
		}
		got, exact = originalAIExactBuildingScore(b, colony, ai.PersonalityHonorable, originalAIBuildScoreContext{})
		if !exact || got != tt.honorable {
			t.Errorf("%s Honorable 分數=(%d,%v)，want (%d,true)", tt.name, got, exact, tt.honorable)
		}
	}
	fallback, ok := gamedata.BuildingByNameZH("研究實驗室")
	if !ok {
		t.Fatal("研究實驗室不存在")
	}
	if score, exact := originalAIExactBuildingScore(fallback, colony, ai.PersonalityHonorable, originalAIBuildScoreContext{}); !exact || score != 5 {
		t.Fatalf("研究實驗室完整公式：score=%d exact=%v，want 5,true", score, exact)
	}
}

func TestOriginalAIResearchBuildingScores(t *testing.T) {
	tests := []struct {
		name    string
		normal  int
		erratic int
	}{
		{"自動實驗室", 11, 15},
		{"銀河網路中心", 11, 11},
		{"行星超級電腦", 8, 11},
		{"研究實驗室", 5, 7},
	}
	for _, tt := range tests {
		name := tt.name
		b, ok := gamedata.BuildingByNameZH(name)
		if !ok {
			t.Fatalf("測試建築不存在：%s", name)
		}
		if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{Population: 9}, ai.PersonalityXenophobic, originalAIBuildScoreContext{}); !exact || score != tt.normal {
			t.Errorf("%s 一般分數=(%d,%v)，want (%d,true)", name, score, exact, tt.normal)
		}
		if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{Population: 9}, ai.PersonalityErratic, originalAIBuildScoreContext{}); !exact || score != tt.erratic {
			t.Errorf("%s Erratic 分數=(%d,%v)，want (%d,true)", name, score, exact, tt.erratic)
		}
		if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{Population: 9}, ai.PersonalityErratic, originalAIBuildScoreContext{lateTech: true}); !exact || score != 0 {
			t.Errorf("%s 晚期科技分數=(%d,%v)，want (0,true)", name, score, exact)
		}
		if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{Population: 9}, ai.PersonalityErratic, originalAIBuildScoreContext{priorityGate: true}); !exact || score != 0 {
			t.Errorf("%s 優先建築 gate 分數=(%d,%v)，want (0,true)", name, score, exact)
		}
	}
}

func TestOriginalAIBiospheresScore(t *testing.T) {
	b, ok := gamedata.BuildingByNameZH("生態圈")
	if !ok {
		t.Fatal("測試建築不存在：生態圈")
	}
	colony := engine.ColonyState{Population: 9}
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityXenophobic, originalAIBuildScoreContext{}); !exact || score != 18 {
		t.Fatalf("Biospheres 一般分數=(%d,%v)，want (18,true)", score, exact)
	}
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, originalAIBuildScoreContext{}); !exact || score != 19 {
		t.Fatalf("Biospheres Pacifist 分數=(%d,%v)，want (19,true)", score, exact)
	}
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, originalAIBuildScoreContext{priorityGate: true}); !exact || score != 0 {
		t.Fatalf("Biospheres priority gate 分數=(%d,%v)，want (0,true)", score, exact)
	}
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, originalAIBuildScoreContext{lateTech: true}); !exact || score != 19 {
		t.Fatalf("Biospheres 不讀 late-tech，分數=(%d,%v)，want (19,true)", score, exact)
	}
}

func TestOriginalAIFoodReplicatorsScore(t *testing.T) {
	b, ok := gamedata.BuildingByNameZH("食物複製機")
	if !ok {
		t.Fatal("測試建築不存在：食物複製機")
	}
	colony := engine.ColonyState{
		Population: 6, Workers: 6, OwnerRaceSlot: 1, OwnerRaceSlotKnown: true, OwnerRaceProfileKnown: true,
		PopulationGroups: []engine.PopulationGroup{
			{RaceSlot: 1, RaceSlotKnown: true, ProfileKnown: true, Workers: 2},
			{RaceSlot: 2, RaceSlotKnown: true, ProfileKnown: true, Workers: 4, Lithovore: true},
		},
	}
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityXenophobic, originalAIBuildScoreContext{}); !exact || score != 4 {
		t.Fatalf("非赤字、外來食岩主要人口分數=(%d,%v)，want (4,true)", score, exact)
	}
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, originalAIBuildScoreContext{empireFoodBalanceHalf: -1}); !exact || score != 9 {
		t.Fatalf("赤字、Pacifist 分數=(%d,%v)，want (9,true)", score, exact)
	}
	colony.Lithovore = true
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, originalAIBuildScoreContext{empireFoodBalanceHalf: -1}); !exact || score != 0 {
		t.Fatalf("owner 已是 Lithovore 時分數=(%d,%v)，want (0,true)", score, exact)
	}
	colony.Lithovore = false
	colony.PopulationGroups[1].Lithovore = false
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, originalAIBuildScoreContext{empireFoodBalanceHalf: -1}); !exact || score != 9 {
		t.Fatalf("主要人口與 owner 均非 Lithovore 時分數=(%d,%v)，want (9,true)", score, exact)
	}
	colony.Lithovore = true
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, originalAIBuildScoreContext{empireFoodBalanceHalf: -1}); !exact || score != 9 {
		t.Fatalf("owner Lithovore、主要人口非 Lithovore 時分數=(%d,%v)，want (9,true)", score, exact)
	}
	colony.Lithovore = false
	colony.PopulationGroups = nil
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, originalAIBuildScoreContext{empireFoodBalanceHalf: -1}); exact || score != 0 {
		t.Fatalf("人口 profile 不完整時不應冒稱 exact：score=%d exact=%v", score, exact)
	}
}

func TestOriginalAIPrimaryPopulationSlotOwnerFallback(t *testing.T) {
	colony := engine.ColonyState{
		Population: 6, Workers: 6, OwnerRaceSlot: 1, OwnerRaceSlotKnown: true, OwnerRaceProfileKnown: true,
		PopulationGroups: []engine.PopulationGroup{
			{RaceSlot: 0, RaceSlotKnown: true, ProfileKnown: true, Workers: 2},
			{RaceSlot: 1, RaceSlotKnown: true, ProfileKnown: true, Workers: 1},
			{RaceSlot: 2, RaceSlotKnown: true, ProfileKnown: true, Workers: 3},
		},
	}
	if slot, known := originalAIPrimaryPopulationSlot(colony); !known || slot != 0 {
		t.Fatalf("無嚴格多數且 owner 不逾三分之一時應走 raw slot-0 fallback：slot=%d known=%v", slot, known)
	}
	colony.PopulationGroups[0].Workers = 1
	colony.PopulationGroups[1].Workers = 3
	colony.PopulationGroups[2].Workers = 2
	if slot, known := originalAIPrimaryPopulationSlot(colony); !known || slot != 1 {
		t.Fatalf("owner 嚴格逾三分之一時應採 owner：slot=%d known=%v", slot, known)
	}
}

func TestOriginalAIBudgetFactorBoundaries(t *testing.T) {
	tests := []struct {
		treasury int
		netBC    int
		want     int
	}{
		{1499, 6400, 0},
		{1500, 63, 0},
		{1500, 64, 1},
		{1500, 255, 1},
		{1500, 256, 2},
		{1500, 6399, 9},
		{1500, 6400, 10},
		{1500, 7744, 10},
		{1500, -63, 0},
		{1500, -64, 10},
	}
	for _, tt := range tests {
		if got := originalAIBudgetFactor(tt.treasury, tt.netBC); got != tt.want {
			t.Errorf("budgetFactor(%d,%d)=%d，want %d", tt.treasury, tt.netBC, got, tt.want)
		}
	}
}

func TestOriginalAICloningCenterScore(t *testing.T) {
	b, ok := gamedata.BuildingByNameZH("複製中心")
	if !ok {
		t.Fatal("測試建築不存在：複製中心")
	}
	ctx := originalAIBuildScoreContext{treasuryBefore: 1500, netBC: 6400}
	if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{}, ai.PersonalityPacifist, ctx); !exact || score != 5 {
		t.Fatalf("非負成長的 Pacifist 不加性格分：score=%d exact=%v，want 5,true", score, exact)
	}
	ctx.raceGrowthPercent = -50
	if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{}, ai.PersonalityPacifist, ctx); !exact || score != 6 {
		t.Fatalf("負成長 Pacifist 分數：score=%d exact=%v，want 6,true", score, exact)
	}
	ctx.treasuryBefore = 1499
	if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{}, ai.PersonalityPacifist, ctx); !exact || score != 1 {
		t.Fatalf("低國庫負成長 Pacifist 分數：score=%d exact=%v，want 1,true", score, exact)
	}
	ctx.priorityGate = true
	if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{}, ai.PersonalityPacifist, ctx); !exact || score != 0 {
		t.Fatalf("priority gate 應先歸零：score=%d exact=%v，want 0,true", score, exact)
	}
}

func TestChooseAICloningCenterUsesPreSettlementTreasury(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Personality = ai.PersonalityXenophobic
	a.Colonies = []engine.ColonyState{{Population: 4, Workers: 4}}
	a.ColonyStars = []int{3}
	a.ColonyBuildings = []map[string]bool{{}}
	a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{gamedata.TOPIC_ADVANCED_BIOLOGY: true}
	for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
		if b.NameZH != "複製中心" {
			a.ColonyBuildings[0][b.NameZH] = true
		}
	}
	out := engine.EmpireOutput{
		Colonies: []engine.ColonyOutput{{NetIndustry: 1}},
		NetBC:    6400,
		Player:   engine.PlayerState{BC: 7899}, // 結算前 1499：factor=0，唯一候選分數為 0。
	}
	if build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1); ok {
		t.Fatalf("結算前國庫 1499 不應選到零分 Cloning Center：%+v", build)
	}
	out.Player.BC = 7900 // 結算前 1500：factor=10，Cloning Center 分數為 5。
	if build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1); !ok || build.Name != "複製中心" {
		t.Fatalf("結算前國庫 1500 應選到 Cloning Center：build=%+v ok=%v", build, ok)
	}
}

func TestOriginalAIMoraleBuildingScores(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{"全息模擬艙", 10},
		{"歡樂穹頂", 16},
	}
	for _, tt := range tests {
		b, ok := gamedata.BuildingByNameZH(tt.name)
		if !ok {
			t.Fatalf("測試建築不存在：%s", tt.name)
		}
		ctx := originalAIBuildScoreContext{government: gamedata.MoraleGovFeudalism}
		if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{Population: 3}, ai.PersonalityPacifist, ctx); !exact || score != tt.want {
			t.Errorf("%s 人口 3 固定分=(%d,%v)，want (%d,true)", tt.name, score, exact, tt.want)
		}
		ctx.priorityGate = true
		if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{Population: 3}, ai.PersonalityErratic, ctx); !exact || score != tt.want {
			t.Errorf("%s 不讀 priority/personality：score=%d exact=%v，want %d,true", tt.name, score, exact, tt.want)
		}
		ctx = originalAIBuildScoreContext{government: gamedata.MoraleGovFeudalism, treasuryBefore: 1499, netBC: 6400}
		if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{Population: 2}, ai.PersonalityXenophobic, ctx); !exact || score != 0 {
			t.Errorf("%s 人口 2、factor 0 應歸零：score=%d exact=%v", tt.name, score, exact)
		}
		ctx.treasuryBefore = 1500
		if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{Population: 2}, ai.PersonalityXenophobic, ctx); !exact || score != tt.want {
			t.Errorf("%s 人口 2、factor>0 分數=(%d,%v)，want (%d,true)", tt.name, score, exact, tt.want)
		}
		if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{Population: 1}, ai.PersonalityXenophobic, ctx); !exact || score != 0 {
			t.Errorf("%s 人口 1 即使 factor>0 仍應歸零：score=%d exact=%v", tt.name, score, exact)
		}
		for gov := gamedata.MoraleGovFeudalism; gov <= gamedata.MoraleGovGalacticUnification; gov++ {
			ctx.government = gov
			want := tt.want
			if int(gov)/2 == 3 {
				want = 0
			}
			if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{Population: 9}, ai.PersonalityXenophobic, ctx); !exact || score != want {
				t.Errorf("%s 政府 %d 分數=(%d,%v)，want (%d,true)", tt.name, gov, score, exact, want)
			}
		}
	}
}

func TestOriginalAIGaiaTransformationScore(t *testing.T) {
	b := gamedata.Building{NameZH: gamedata.GaiaTransformationActionName, NameEN: "Gaia Transformation"}
	ctx := originalAIBuildScoreContext{treasuryBefore: 1499, netBC: 6400, priorityGate: true, lateTech: true}
	if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{}, ai.PersonalityXenophobic, ctx); !exact || score != 0 {
		t.Fatalf("低國庫一般性格分數=(%d,%v)，want (0,true)", score, exact)
	}
	if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{}, ai.PersonalityPacifist, ctx); !exact || score != 1 {
		t.Fatalf("Gaia 不讀 priority/late-tech，Pacifist 分數=(%d,%v)，want (1,true)", score, exact)
	}
	ctx.treasuryBefore = 1500
	if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{}, ai.PersonalityPacifist, ctx); !exact || score != 11 {
		t.Fatalf("國庫門檻後 Pacifist 分數=(%d,%v)，want (11,true)", score, exact)
	}
}

func TestAICompletesGaiaTransformationAsSpecial(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{gamedata.TOPIC_TRANS_GENETICS: true}
	a.Colonies[0].Climate = gamedata.TERRAN
	a.Colonies[0].PopMax = 10
	planet := s.aiColonyPlanet(0, 0)
	if planet == nil {
		t.Fatal("測試 AI 殖民地缺少對應全局行星")
	}
	syncPlanetClimate(planet, gamedata.TERRAN)
	key := aiColonyBuildKey(a, 0)
	a.ColonyBuilds = map[int]ColonyBuild{key: {
		Name: gamedata.GaiaTransformationActionName, Progress: 499, Cost: 500,
	}}
	out := engine.EmpireOutput{Colonies: []engine.ColonyOutput{{NetIndustry: 1}}}
	s.advanceAIColonyBuilds(0, out)
	if a.Colonies[0].Climate != gamedata.GAIA || planet.ClimateID != gamedata.GAIA {
		t.Fatalf("Gaia 完工未同步 colony／planet：%v／%v", a.Colonies[0].Climate, planet.ClimateID)
	}
	if len(a.ColonyBuildings) > 0 && a.ColonyBuildings[0][gamedata.GaiaTransformationActionName] {
		t.Fatal("Gaia Transformation 是 Special，不得殘留於 ColonyBuildings")
	}
	if _, ok := a.ColonyBuilds[key]; ok {
		t.Fatal("Gaia 完工後目前產品應清空")
	}
}

func TestAIGaiaTransformationCandidateRequiresTerran(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Personality = ai.PersonalityPacifist
	a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{gamedata.TOPIC_TRANS_GENETICS: true}
	a.ColonyBuildings[0] = make(map[string]bool)
	for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
		a.ColonyBuildings[0][b.NameZH] = true
	}
	out := engine.EmpireOutput{
		Colonies: []engine.ColonyOutput{{NetIndustry: 1}},
		Player:   engine.PlayerState{BC: 1499},
	}
	a.Colonies[0].Climate = gamedata.DESERT
	if build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1); ok {
		t.Fatalf("非 Terran 不得選 Gaia Transformation：%+v", build)
	}
	a.Colonies[0].Climate = gamedata.TERRAN
	if build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1); !ok || build.Name != gamedata.GaiaTransformationActionName {
		t.Fatalf("Terran 的唯一正分候選應是 Gaia Transformation：build=%+v ok=%v", build, ok)
	}
}

func TestOriginalAITerraformingScore(t *testing.T) {
	b := gamedata.Building{NameZH: gamedata.TerraformActionName, NameEN: "Terraforming"}
	tests := []struct {
		climate             gamedata.PlanetClimate
		nonAquatic, aquatic int
	}{
		{gamedata.BARREN, 2, 2},
		{gamedata.DESERT, 1, 1},
		{gamedata.TUNDRA, 0, 1},
		{gamedata.OCEAN, 4, 0},
		{gamedata.SWAMP, 6, 0},
		{gamedata.ARID, 1, 1},
	}
	for _, tt := range tests {
		colony := engine.ColonyState{Climate: tt.climate}
		if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityXenophobic,
			originalAIBuildScoreContext{treasuryBefore: 1499}); !exact || score != tt.nonAquatic {
			t.Errorf("氣候 %d 非 Aquatic 分數=(%d,%v)，want (%d,true)", tt.climate, score, exact, tt.nonAquatic)
		}
		if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityXenophobic,
			originalAIBuildScoreContext{treasuryBefore: 1499, aquatic: true}); !exact || score != tt.aquatic {
			t.Errorf("氣候 %d Aquatic 分數=(%d,%v)，want (%d,true)", tt.climate, score, exact, tt.aquatic)
		}
		want := tt.nonAquatic + 3 + 10
		if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist,
			originalAIBuildScoreContext{treasuryBefore: 1500, netBC: 6400}); !exact || score != want {
			t.Errorf("氣候 %d Pacifist+budget 分數=(%d,%v)，want (%d,true)", tt.climate, score, exact, want)
		}
		if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist,
			originalAIBuildScoreContext{priorityGate: true, treasuryBefore: 1500, netBC: 6400}); !exact || score != 0 {
			t.Errorf("氣候 %d priority gate 分數=(%d,%v)，want (0,true)", tt.climate, score, exact)
		}
	}
}

func TestAICompletesTerraformingAsSpecial(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Colonies[0].Climate = gamedata.BARREN
	a.Colonies[0].PopMax = 4
	planet := s.aiColonyPlanet(0, 0)
	if planet == nil {
		t.Fatal("測試 AI 殖民地缺少對應全局行星")
	}
	syncPlanetClimate(planet, gamedata.BARREN)
	key := aiColonyBuildKey(a, 0)
	a.ColonyBuilds = map[int]ColonyBuild{key: {
		Name: gamedata.TerraformActionName, Progress: 249, Cost: 250,
	}}
	s.advanceAIColonyBuilds(0, engine.EmpireOutput{Colonies: []engine.ColonyOutput{{NetIndustry: 1}}})
	want := gamedata.TerraformNextClimateOptions(gamedata.BARREN)[0]
	if a.Colonies[0].Climate != want || planet.ClimateID != want {
		t.Fatalf("Terraforming 完工未同步 colony／planet：%v／%v，want %v", a.Colonies[0].Climate, planet.ClimateID, want)
	}
	if len(a.ColonyBuildings) > 0 && a.ColonyBuildings[0][gamedata.TerraformActionName] {
		t.Fatal("Terraforming 是 Special，不得殘留於 ColonyBuildings")
	}
}

func TestAITerraformingCandidateRequiresNextClimate(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Personality = ai.PersonalityPacifist
	a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{gamedata.TOPIC_GENETIC_MUTATIONS: true}
	a.ColonyBuildings[0] = make(map[string]bool)
	for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
		a.ColonyBuildings[0][b.NameZH] = true
	}
	out := engine.EmpireOutput{Colonies: []engine.ColonyOutput{{NetIndustry: 1}}, Player: engine.PlayerState{BC: 1499}}
	a.Colonies[0].Climate = gamedata.TOXIC
	if build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1); ok {
		t.Fatalf("無下一級的 Toxic 不得選 Terraforming：%+v", build)
	}
	a.Colonies[0].Climate = gamedata.BARREN
	if build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1); !ok || build.Name != gamedata.TerraformActionName {
		t.Fatalf("Barren 的唯一 Special 候選應是 Terraforming：build=%+v ok=%v", build, ok)
	}
}

func TestOriginalAISoilEnrichmentScore(t *testing.T) {
	b := gamedata.Building{NameZH: gamedata.SoilEnrichmentActionName, NameEN: "Soil Enrichment"}
	colony := engine.ColonyState{
		Population: 4, Workers: 4, FoodPerFarmer: 1,
		OwnerRaceSlot: 1, OwnerRaceSlotKnown: true, OwnerRaceProfileKnown: true,
		PopulationGroups: []engine.PopulationGroup{{
			RaceSlot: 1, RaceSlotKnown: true, ProfileKnown: true, Workers: 4,
		}},
	}
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityXenophobic, originalAIBuildScoreContext{}); !exact || score != 3 {
		t.Fatalf("一般、非赤字分數=(%d,%v)，want (3,true)", score, exact)
	}
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist,
		originalAIBuildScoreContext{empireFoodBalanceHalf: -1}); !exact || score != 7 {
		t.Fatalf("Pacifist、赤字分數=(%d,%v)，want (7,true)", score, exact)
	}
	colony.Lithovore = true
	colony.PopulationGroups[0].Lithovore = true
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist,
		originalAIBuildScoreContext{empireFoodBalanceHalf: -1}); !exact || score != 0 {
		t.Fatalf("主要人口與 owner 同為 Lithovore 應歸零：score=%d exact=%v", score, exact)
	}
	colony.PopulationGroups[0].Lithovore = false
	colony.PopulationGroups = []engine.PopulationGroup{
		{RaceSlot: 1, RaceSlotKnown: true, ProfileKnown: true, Workers: 1, Lithovore: true},
		{RaceSlot: 2, RaceSlotKnown: true, ProfileKnown: true, Workers: 3},
	}
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityXenophobic, originalAIBuildScoreContext{}); !exact || score != 3 {
		t.Fatalf("只有 owner Lithovore 仍可建：score=%d exact=%v", score, exact)
	}
	colony.FoodPerFarmer = 0
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist,
		originalAIBuildScoreContext{empireFoodBalanceHalf: -1}); !exact || score != 0 {
		t.Fatalf("FoodPerFarmer=0 應歸零：score=%d exact=%v", score, exact)
	}
	colony.FoodPerFarmer = 1
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist,
		originalAIBuildScoreContext{priorityGate: true, empireFoodBalanceHalf: -1}); !exact || score != 0 {
		t.Fatalf("priority gate 應歸零：score=%d exact=%v", score, exact)
	}
	colony.PopulationGroups = nil
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, originalAIBuildScoreContext{}); exact || score != 0 {
		t.Fatalf("人口 profile 不完整不得冒稱 exact：score=%d exact=%v", score, exact)
	}
}

func completeAIFoodScoreColony() engine.ColonyState {
	return engine.ColonyState{
		Population: 4, Workers: 4, Climate: gamedata.TERRAN,
		OwnerRaceSlot: 1, OwnerRaceSlotKnown: true, OwnerRaceProfileKnown: true,
		PopulationGroups: []engine.PopulationGroup{{
			RaceSlot: 1, RaceSlotKnown: true, ProfileKnown: true, Workers: 4,
		}},
	}
}

func TestOriginalAIFoodBuildingScoreTables(t *testing.T) {
	colony := completeAIFoodScoreColony()
	tests := []struct {
		name  string
		raw   int
		bases [4]int
		pac   int
	}{
		{"水耕農場", 21, [4]int{12, 11, 10, 6}, 4},
		{"地底農場", 43, [4]int{13, 12, 10, 7}, 3},
	}
	for _, tt := range tests {
		b, ok := gamedata.BuildingByNameZH(tt.name)
		if !ok {
			t.Fatalf("測試建築不存在：%s", tt.name)
		}
		for half, want := range tt.bases {
			ctx := originalAIBuildScoreContext{colonyFoodHalf: half, colonyFoodHalfKnown: true}
			if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityXenophobic, ctx); !exact || score != want {
				t.Errorf("%s half=%d 分數=(%d,%v)，want (%d,true)", tt.name, half, score, exact, want)
			}
			ctx.empireFoodBalanceHalf = -3
			if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, ctx); !exact || score != want+3+tt.pac {
				t.Errorf("%s half=%d Pacifist 赤字分數=(%d,%v)，want (%d,true)", tt.name, half, score, exact, want+3+tt.pac)
			}
		}
		ctx := originalAIBuildScoreContext{priorityGate: true, colonyFoodHalf: 2, colonyFoodHalfKnown: true}
		score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityXenophobic, ctx)
		want := 10
		if tt.raw == 43 {
			want = 0
		}
		if !exact || score != want {
			t.Errorf("%s priority gate 分數=(%d,%v)，want (%d,true)", tt.name, score, exact, want)
		}
	}
}

func TestOriginalAIWeatherControllerScore(t *testing.T) {
	b, ok := gamedata.BuildingByNameZH("氣候控制器")
	if !ok {
		t.Fatal("氣候控制器不存在")
	}
	colony := completeAIFoodScoreColony()
	ctx := originalAIBuildScoreContext{colonyFoodHalfKnown: true}
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, ctx); !exact || score != 0 {
		t.Fatalf("零食物快取分數=(%d,%v)，want (0,true)", score, exact)
	}
	ctx.colonyFoodHalf = 2
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityXenophobic, ctx); !exact || score != 5 {
		t.Fatalf("非赤字分數=(%d,%v)，want (5,true)", score, exact)
	}
	ctx.empireFoodBalanceHalf = -3
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, ctx); !exact || score != 12 {
		t.Fatalf("Pacifist 赤字分數=(%d,%v)，want (12,true)", score, exact)
	}
	ctx.priorityGate = true
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, ctx); !exact || score != 0 {
		t.Fatalf("priority gate 分數=(%d,%v)，want (0,true)", score, exact)
	}
}

func TestOriginalAIFoodBuildingProfileAndCacheGates(t *testing.T) {
	b, _ := gamedata.BuildingByNameZH("水耕農場")
	colony := completeAIFoodScoreColony()
	ctx := originalAIBuildScoreContext{colonyFoodHalf: 2, colonyFoodHalfKnown: true}
	colony.Lithovore = true
	colony.PopulationGroups[0].Lithovore = true
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, ctx); !exact || score != 0 {
		t.Fatalf("主要人口與 owner 同為 Lithovore 應歸零：score=%d exact=%v", score, exact)
	}
	colony.PopulationGroups = nil
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, ctx); exact || score != 0 {
		t.Fatalf("人口 profile 不完整不得冒稱 exact：score=%d exact=%v", score, exact)
	}
	colony = completeAIFoodScoreColony()
	ctx.colonyFoodHalfKnown = false
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, ctx); exact || score != 0 {
		t.Fatalf("食物快取未知不得冒稱 exact：score=%d exact=%v", score, exact)
	}
	if got, known := originalAIColonyFoodHalf(engine.ColonyState{Climate: gamedata.TERRAN}, map[string]bool{
		"氣候控制器": true, "太空大學": true,
	}, nil); !known || got != 10 {
		t.Fatalf("Terran+Weather+Astro raw half=(%d,%v)，want (10,true)", got, known)
	}
	if got, known := originalAIColonyFoodHalf(engine.ColonyState{Climate: gamedata.TOXIC}, nil,
		map[gamedata.Technology]bool{gamedata.TECH_BIOMORPHIC_FUNGI: true}); !known || got != 2 {
		t.Fatalf("Toxic+Biomorphic Fungi raw half=(%d,%v)，want (2,true)", got, known)
	}
}

func TestAIFoodBuildingsCandidateAndCompletion(t *testing.T) {
	tests := []struct {
		name      string
		topic     gamedata.ResearchTopic
		startFood int
		startFlat int
		wantFood  int
		wantFlat  int
	}{
		{"水耕農場", gamedata.TOPIC_ASTRO_BIOLOGY, 2, 0, 2, 2},
		{"地底農場", gamedata.TOPIC_MACRO_GENETICS, 2, 0, 2, 4},
		{"氣候控制器", gamedata.TOPIC_MACRO_GENETICS, 2, 0, 4, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDemoSession()
			a := &s.AIPlayers[0]
			a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{tt.topic: true}
			a.Colonies[0] = completeAIFoodScoreColony()
			a.Colonies[0].FoodPerFarmer = tt.startFood
			a.Colonies[0].FlatFood = tt.startFlat
			a.ColonyBuildings[0] = make(map[string]bool)
			for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
				a.ColonyBuildings[0][b.NameZH] = true
			}
			delete(a.ColonyBuildings[0], tt.name)
			if tt.name == gamedata.BuildingPlanetaryFluxShield {
				a.ColonyBuildings[0][gamedata.BuildingPlanetaryRadiationShield] = true
			}
			if tt.name == gamedata.BuildingPlanetaryBarrierShield {
				a.ColonyBuildings[0][gamedata.BuildingPlanetaryRadiationShield] = true
				a.ColonyBuildings[0][gamedata.BuildingPlanetaryFluxShield] = true
			}
			out := engine.EmpireOutput{Colonies: []engine.ColonyOutput{{NetIndustry: 1}}, Player: engine.PlayerState{BC: 100}}
			build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1)
			if !ok || build.Name != tt.name {
				t.Fatalf("唯一候選錯誤：build=%+v ok=%v", build, ok)
			}
			key := aiColonyBuildKey(a, 0)
			a.ColonyBuilds = map[int]ColonyBuild{key: {Name: tt.name, Cost: 1}}
			s.advanceAIColonyBuilds(0, out)
			if !a.ColonyBuildings[0][tt.name] || a.Colonies[0].FoodPerFarmer != tt.wantFood || a.Colonies[0].FlatFood != tt.wantFlat {
				t.Fatalf("完工未垂直接線：built=%v food=%d flat=%d，want true/%d/%d",
					a.ColonyBuildings[0][tt.name], a.Colonies[0].FoodPerFarmer, a.Colonies[0].FlatFood, tt.wantFood, tt.wantFlat)
			}
		})
	}
}

func TestOriginalAIPollutionBuildingScores(t *testing.T) {
	colony := completeAIFoodScoreColony()
	tests := []struct {
		name string
		raw  int
	}{
		{"大氣更新器", 5},
		{"核心廢料場", 13},
		{"污染處理器", 32},
	}
	for _, tt := range tests {
		b, ok := gamedata.BuildingByNameZH(tt.name)
		if !ok {
			t.Fatalf("測試建築不存在：%s", tt.name)
		}
		for _, tc := range []struct {
			cleanup int
			want    int
		}{{5, 0}, {6, 0}, {10, 0}, {11, 3}, {16, 4}} {
			ctx := originalAIBuildScoreContext{pollutionCleanupCost: tc.cleanup}
			if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityXenophobic, ctx); !exact || score != tc.want {
				t.Errorf("%s cleanup=%d 分數=(%d,%v)，want (%d,true)", tt.name, tc.cleanup, score, exact, tc.want)
			}
			if tc.cleanup > 5 {
				if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, ctx); !exact || score != tc.want+1 {
					t.Errorf("%s cleanup=%d Pacifist 分數=(%d,%v)，want (%d,true)", tt.name, tc.cleanup, score, exact, tc.want+1)
				}
			}
		}
		ctx := originalAIBuildScoreContext{priorityGate: true, pollutionCleanupCost: 16}
		score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityXenophobic, ctx)
		want := 4
		if tt.raw == 13 {
			want = 0
		}
		if !exact || score != want {
			t.Errorf("%s priority gate 分數=(%d,%v)，want (%d,true)", tt.name, score, exact, want)
		}
	}
}

func TestOriginalAIPollutionBuildingPrimaryTolerantGate(t *testing.T) {
	b, _ := gamedata.BuildingByNameZH("污染處理器")
	colony := completeAIFoodScoreColony()
	ctx := originalAIBuildScoreContext{pollutionCleanupCost: 16}
	colony.TolerantRace = true
	colony.PopulationGroups[0].Tolerant = true
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, ctx); !exact || score != 0 {
		t.Fatalf("主要人口 Tolerant 應歸零：score=%d exact=%v", score, exact)
	}
	colony.TolerantRace = false
	colony.PopulationGroups = []engine.PopulationGroup{
		{RaceSlot: 1, RaceSlotKnown: true, ProfileKnown: true, Workers: 1},
		{RaceSlot: 2, RaceSlotKnown: true, ProfileKnown: true, Workers: 3, Tolerant: true},
	}
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, ctx); !exact || score != 0 {
		t.Fatalf("主要外族人口 Tolerant 應歸零：score=%d exact=%v", score, exact)
	}
	colony.PopulationGroups = nil
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, ctx); exact || score != 0 {
		t.Fatalf("人口 profile 不完整不得冒稱 exact：score=%d exact=%v", score, exact)
	}
}

func TestAIPollutionBuildingsCandidateAndCompletion(t *testing.T) {
	tests := []struct {
		name  string
		topic gamedata.ResearchTopic
		check func(engine.ColonyState) bool
	}{
		{"大氣更新器", gamedata.TOPIC_MOLECULAR_COMPRESSION, func(c engine.ColonyState) bool { return c.AtmosphericRenewer }},
		{"核心廢料場", gamedata.TOPIC_TECTONIC_ENGINEERING, func(c engine.ColonyState) bool { return c.CoreWasteDump }},
		{"污染處理器", gamedata.TOPIC_ADVANCED_CHEMISTRY, func(c engine.ColonyState) bool { return c.PollutionProcessor }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDemoSession()
			a := &s.AIPlayers[0]
			a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{tt.topic: true}
			a.Colonies[0] = completeAIFoodScoreColony()
			a.ColonyBuildings[0] = make(map[string]bool)
			for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
				a.ColonyBuildings[0][b.NameZH] = true
			}
			delete(a.ColonyBuildings[0], tt.name)
			out := engine.EmpireOutput{Colonies: []engine.ColonyOutput{{NetIndustry: 20, PollutionCleanupCost: 16}}}
			build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1)
			if !ok || build.Name != tt.name {
				t.Fatalf("唯一候選錯誤：build=%+v ok=%v", build, ok)
			}
			key := aiColonyBuildKey(a, 0)
			a.ColonyBuilds = map[int]ColonyBuild{key: {Name: tt.name, Cost: 1}}
			s.advanceAIColonyBuilds(0, out)
			if !a.ColonyBuildings[0][tt.name] || !tt.check(a.Colonies[0]) {
				t.Fatalf("完工未垂直接線：built=%v colony=%+v", a.ColonyBuildings[0][tt.name], a.Colonies[0])
			}
		})
	}
}

func TestOriginalAIPlanetaryGravityGeneratorScore(t *testing.T) {
	b, ok := gamedata.BuildingByNameZH("行星重力產生器")
	if !ok {
		t.Fatal("行星重力產生器不存在")
	}
	tests := []struct {
		name     string
		low      bool
		high     bool
		gravity  gamedata.PlanetGravity
		wantBase int
	}{
		{"一般-LowG星", false, false, gamedata.LOW_G, 3},
		{"一般-NormalG星", false, false, gamedata.NORMAL_G, 0},
		{"一般-HeavyG星", false, false, gamedata.HEAVY_G, 6},
		{"LowG族-LowG星", true, false, gamedata.LOW_G, 0},
		{"LowG族-NormalG星", true, false, gamedata.NORMAL_G, 3},
		{"LowG族-HeavyG星", true, false, gamedata.HEAVY_G, 6},
		{"HighG族-LowG星", false, true, gamedata.LOW_G, 3},
		{"HighG族-NormalG星", false, true, gamedata.NORMAL_G, 0},
		{"HighG族-HeavyG星", false, true, gamedata.HEAVY_G, 0},
		{"雙trait採HighG-LowG星", true, true, gamedata.LOW_G, 3},
		{"雙trait採HighG-HeavyG星", true, true, gamedata.HEAVY_G, 0},
	}
	for _, tt := range tests {
		colony := engine.ColonyState{PlanetGravity: tt.gravity}
		ctx := originalAIBuildScoreContext{
			ownerLowGravity: tt.low, ownerHighGravity: tt.high, priorityGate: true,
		}
		if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityXenophobic, ctx); !exact || score != tt.wantBase {
			t.Errorf("%s 一般分數=(%d,%v)，want (%d,true)", tt.name, score, exact, tt.wantBase)
		}
		wantPacifist := tt.wantBase
		if wantPacifist > 0 {
			wantPacifist++
		}
		if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, ctx); !exact || score != wantPacifist {
			t.Errorf("%s Pacifist 分數=(%d,%v)，want (%d,true)", tt.name, score, exact, wantPacifist)
		}
	}
}

func TestOriginalAIPrimaryPopulationCapacity(t *testing.T) {
	colony := engine.ColonyState{
		Population: 5, Workers: 5, PlanetSize: gamedata.MEDIUM_PLANET, Climate: gamedata.TERRAN,
		OwnerRaceSlot: 1, OwnerRaceSlotKnown: true, OwnerRaceProfileKnown: true,
		PopulationGroups: []engine.PopulationGroup{
			{RaceSlot: 1, RaceSlotKnown: true, ProfileKnown: true, Workers: 1},
			{RaceSlot: 2, RaceSlotKnown: true, ProfileKnown: true, Workers: 4, Subterranean: true},
		},
	}
	known := map[gamedata.Technology]bool{gamedata.TECH_ADVANCED_CITY_PLANNING: true}
	capacity, ok := originalAIPrimaryPopulationCapacity(colony, map[string]bool{"生態圈": true}, known)
	// Medium／Terran 基礎 12，主要外族 Subterranean +6，ACP +5，Biospheres +2。
	if !ok || capacity != 25 {
		t.Fatalf("主要外族人口容量=(%d,%v)，want (25,true)", capacity, ok)
	}
	colony.PopulationGroups = nil
	if capacity, ok := originalAIPrimaryPopulationCapacity(colony, nil, nil); ok || capacity != 0 {
		t.Fatalf("人口 profile 不完整不得冒稱容量 exact：capacity=%d known=%v", capacity, ok)
	}
}

func TestOriginalAICommerceAndRecyclotronScores(t *testing.T) {
	colony := engine.ColonyState{
		Population: 5, Workers: 5, OwnerRaceSlot: 1, OwnerRaceSlotKnown: true, OwnerRaceProfileKnown: true,
		PopulationGroups: []engine.PopulationGroup{{RaceSlot: 1, RaceSlotKnown: true, ProfileKnown: true, Workers: 5}},
	}
	stock, _ := gamedata.BuildingByNameZH("行星證券交易所")
	spaceport, _ := gamedata.BuildingByNameZH("太空港")
	recyclotron, _ := gamedata.BuildingByNameZH("再生反應爐")

	ctx := originalAIBuildScoreContext{primaryPopCapacity: 21, primaryPopCapKnown: true}
	if score, exact := originalAIExactBuildingScore(stock, colony, ai.PersonalityXenophobic, ctx); !exact || score != 8 {
		t.Fatalf("證交所一般分數=(%d,%v)，want (8,true)", score, exact)
	}
	if score, exact := originalAIExactBuildingScore(stock, colony, ai.PersonalityHonorable, ctx); !exact || score != 9 {
		t.Fatalf("證交所 Honorable 分數=(%d,%v)，want (9,true)", score, exact)
	}
	colony.Population = 4
	if score, exact := originalAIExactBuildingScore(stock, colony, ai.PersonalityHonorable, ctx); !exact || score != 0 {
		t.Fatalf("證交所人口 4 應被門檻阻擋：score=%d exact=%v", score, exact)
	}
	colony.Population = 3
	ctx.primaryPopCapacity = 20
	if score, exact := originalAIExactBuildingScore(spaceport, colony, ai.PersonalityXenophobic, ctx); !exact || score != 7 {
		t.Fatalf("太空港一般分數=(%d,%v)，want (7,true)", score, exact)
	}
	if score, exact := originalAIExactBuildingScore(spaceport, colony, ai.PersonalityHonorable, ctx); !exact || score != 8 {
		t.Fatalf("太空港 Honorable 分數=(%d,%v)，want (8,true)", score, exact)
	}
	ctx.priorityGate = true
	if score, exact := originalAIExactBuildingScore(spaceport, colony, ai.PersonalityHonorable, ctx); !exact || score != 0 {
		t.Fatalf("太空港 priority gate 應歸零：score=%d exact=%v", score, exact)
	}

	colony.Population = 6
	colony.Workers = 6
	colony.PopulationGroups[0].Workers = 6
	ctx = originalAIBuildScoreContext{primaryPopCapacity: 18, primaryPopCapKnown: true, priorityGate: true}
	if score, exact := originalAIExactBuildingScore(recyclotron, colony, ai.PersonalityXenophobic, ctx); !exact || score != 12 {
		t.Fatalf("再生反應爐一般分數=(%d,%v)，want (12,true)", score, exact)
	}
	if score, exact := originalAIExactBuildingScore(recyclotron, colony, ai.PersonalityPacifist, ctx); !exact || score != 14 {
		t.Fatalf("再生反應爐 Pacifist 應 +2：score=%d exact=%v", score, exact)
	}
	if score, exact := originalAIExactBuildingScore(recyclotron, colony, ai.PersonalityHonorable, ctx); !exact || score != 14 {
		t.Fatalf("再生反應爐 Honorable 應 +2：score=%d exact=%v", score, exact)
	}
	colony.TolerantRace = true
	colony.PopulationGroups[0].Tolerant = true
	if score, exact := originalAIExactBuildingScore(recyclotron, colony, ai.PersonalityXenophobic, ctx); !exact || score != 10 {
		t.Fatalf("主要人口 Tolerant 應少 2：score=%d exact=%v", score, exact)
	}
	ctx.primaryPopCapKnown = false
	if score, exact := originalAIExactBuildingScore(recyclotron, colony, ai.PersonalityXenophobic, ctx); exact || score != 0 {
		t.Fatalf("容量未知不得冒稱 exact：score=%d exact=%v", score, exact)
	}
}

func TestAICommerceAndRecyclotronCandidateCompletionConsumers(t *testing.T) {
	tests := []struct {
		name  string
		topic gamedata.ResearchTopic
	}{
		{"行星證券交易所", gamedata.TOPIC_MACRO_ECONOMICS},
		{"太空港", gamedata.TOPIC_ASTRO_ENGINEERING},
		{"再生反應爐", gamedata.TOPIC_ADVANCED_MANUFACTURING},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDemoSession()
			a := &s.AIPlayers[0]
			a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{tt.topic: true}
			a.Colonies[0] = engine.ColonyState{
				Population: 6, PopMax: 12, Workers: 6, IndustryPerWorker: 4,
				PlanetSize: gamedata.MEDIUM_PLANET, PlanetGravity: gamedata.NORMAL_G,
				MineralRichness: gamedata.ABUNDANT, Climate: gamedata.TERRAN,
				OwnerRaceSlot: 1, OwnerRaceSlotKnown: true, OwnerRaceProfileKnown: true,
				PopulationGroups: []engine.PopulationGroup{{RaceSlot: 1, RaceSlotKnown: true, ProfileKnown: true, Workers: 6}},
			}
			a.ColonyBuildings[0] = make(map[string]bool)
			for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
				a.ColonyBuildings[0][b.NameZH] = true
			}
			delete(a.ColonyBuildings[0], tt.name)
			capacity, capacityKnown := originalAIPrimaryPopulationCapacity(a.Colonies[0], a.ColonyBuildings[0], knownTechnologyApplications(a.Player))
			target, found := gamedata.BuildingByNameZH(tt.name)
			if !found {
				t.Fatalf("測試建築不存在：%s", tt.name)
			}
			ctx := originalAIBuildScoreContext{primaryPopCapacity: capacity, primaryPopCapKnown: capacityKnown}
			if score, exact := originalAIExactBuildingScore(target, a.Colonies[0], a.Personality, ctx); !exact || score <= 0 {
				t.Fatalf("唯一候選前未走精確正分：score=%d exact=%v capacity=%d known=%v", score, exact, capacity, capacityKnown)
			}
			out := engine.EmpireOutput{Colonies: []engine.ColonyOutput{{NetIndustry: 20}}}
			build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1)
			if !ok || build.Name != tt.name {
				t.Fatalf("唯一候選錯誤：build=%+v ok=%v", build, ok)
			}
			beforeColony := engine.RunColonyTurn(a.Colonies[0])
			beforeEmpire := engine.RunEmpireTurn(engine.PlayerState{TaxRate: 50}, []engine.ColonyState{a.Colonies[0]})
			key := aiColonyBuildKey(a, 0)
			a.ColonyBuilds = map[int]ColonyBuild{key: {Name: tt.name, Cost: 1}}
			s.advanceAIColonyBuilds(0, out)
			if !a.ColonyBuildings[0][tt.name] {
				t.Fatalf("完工未寫入建築旗標：%s", tt.name)
			}
			if tt.name == "再生反應爐" {
				after := engine.RunColonyTurn(a.Colonies[0])
				if after.NetIndustry-beforeColony.NetIndustry != a.Colonies[0].Population ||
					after.PollutingProduction != beforeColony.PollutingProduction {
					t.Fatalf("再生反應爐消費端錯誤：before=%+v after=%+v", beforeColony, after)
				}
				return
			}
			after := engine.RunEmpireTurn(engine.PlayerState{TaxRate: 50}, []engine.ColonyState{a.Colonies[0]})
			if after.TaxRevenue <= beforeEmpire.TaxRevenue {
				t.Fatalf("%s 完工後收入未增加：%d → %d", tt.name, beforeEmpire.TaxRevenue, after.TaxRevenue)
			}
		})
	}
}

func TestOriginalAISpaceAcademyScore(t *testing.T) {
	academy, ok := gamedata.BuildingByNameZH("太空學院")
	if !ok {
		t.Fatal("太空學院不存在")
	}
	tests := []struct {
		name       string
		population int
		net        int
		treasury   int
		netBC      int
		priority   bool
		want       int
	}{
		{"低人口低產能無預算", 4, 14, 0, 0, false, 0},
		{"低人口低產能有預算", 4, 14, 1500, 64, false, 1000},
		{"人口門通過負差值", 5, 14, 0, 0, false, 1000},
		{"差值零", 5, 15, 0, 0, false, 0},
		{"差值一", 5, 16, 0, 0, false, 1},
		{"產能門直接通過", 4, 17, 0, 0, false, 1},
		{"priority gate", 5, 100, 1500, 6400, true, 0},
	}
	for _, tt := range tests {
		ctx := originalAIBuildScoreContext{
			netIndustry: tt.net, treasuryBefore: tt.treasury, netBC: tt.netBC, priorityGate: tt.priority,
		}
		score, exact := originalAIExactBuildingScore(academy, engine.ColonyState{Population: tt.population}, ai.PersonalityErratic, ctx)
		if !exact || score != tt.want {
			t.Errorf("%s 分數=(%d,%v)，want (%d,true)", tt.name, score, exact, tt.want)
		}
	}
}

func TestAISpaceAcademyCandidateCompletionAndCrewConsumer(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{gamedata.TOPIC_MILITARY_TACTICS: true}
	a.Colonies[0].Population = 5
	a.ColonyBuildings[0] = make(map[string]bool)
	for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
		a.ColonyBuildings[0][b.NameZH] = true
	}
	delete(a.ColonyBuildings[0], spaceAcademyName)
	out := engine.EmpireOutput{Colonies: []engine.ColonyOutput{{NetIndustry: 17}}}
	build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1)
	if !ok || build.Name != spaceAcademyName {
		t.Fatalf("唯一候選錯誤：build=%+v ok=%v", build, ok)
	}
	a.FleetStar = a.ColonyStars[0]
	a.FleetPosSet = true
	a.Ships = []Ship{{Name: "學院驗證艦", Class: "巡防艦"}}
	key := aiColonyBuildKey(a, 0)
	a.ColonyBuilds = map[int]ColonyBuild{key: {Name: spaceAcademyName, Cost: 1}}
	s.advanceAIColonyBuilds(0, out)
	if !a.ColonyBuildings[0][spaceAcademyName] {
		t.Fatal("太空學院完工未寫入 AI 殖民地建築 map")
	}
	s.advanceCrewExperience()
	want := gamedata.CrewXPPerTurnInSpace + gamedata.SpaceAcademyXPPerTurn
	if got := a.Ships[0].CrewXP; got != want {
		t.Fatalf("同星系 AI 艦艇經驗=%d，want %d", got, want)
	}
}

func TestAIOriginalBarracksScores(t *testing.T) {
	armor, ok := gamedata.BuildingByNameZH(armorBarracksBuildingName)
	if !ok {
		t.Fatal("找不到裝甲營房")
	}
	marine, ok := gamedata.BuildingByNameZH(marineBarracksBuildingName)
	if !ok {
		t.Fatal("找不到海軍陸戰隊營")
	}
	colony := engine.ColonyState{Population: 3}
	ctx := originalAIBuildScoreContext{
		strategicPressureContextKnown: true,
		reachTreatyNear:               1,
		reachNoPolicyNear:             2,
		reachWarNear:                  3,
		reachExtended:                 4,
		incomingOtherFleetETA9:        true,
		hostileAlienPopulation:        true,
		armorBarracksBuilt:            true,
		marineBarracksBuilt:           true,
		government:                    gamedata.MoraleGovDemocracy,
	}
	if score, exact := originalAIExactBuildingScore(armor, colony, ai.PersonalityRuthless, ctx); !exact || score != 20 {
		t.Fatalf("Armor Barracks 分數=(%d,%v)，want (20,true)", score, exact)
	}
	if score, exact := originalAIExactBuildingScore(marine, colony, ai.PersonalityRuthless, ctx); !exact || score != 42 {
		t.Fatalf("Marine Barracks 分數=(%d,%v)，want (42,true)", score, exact)
	}

	ctx = originalAIBuildScoreContext{strategicPressureContextKnown: true, government: gamedata.MoraleGovDictatorship}
	if score, exact := originalAIExactBuildingScore(armor, engine.ColonyState{Population: 2}, ai.PersonalityAggressive, ctx); !exact || score != 0 {
		t.Fatalf("人口 2、無 budget 的 Armor gate=(%d,%v)，want (0,true)", score, exact)
	}
	if score, exact := originalAIExactBuildingScore(marine, engine.ColonyState{Population: 2}, ai.PersonalityAggressive, ctx); !exact || score != 12 {
		t.Fatalf("Marine 交叉 Armor／政府加分=(%d,%v)，want (12,true)", score, exact)
	}
	ctx.treasuryBefore = 1500
	ctx.netBC = 64
	if score, exact := originalAIExactBuildingScore(armor, engine.ColonyState{Population: 2}, ai.PersonalityAggressive, ctx); !exact || score != 6 {
		t.Fatalf("正 budget 應通過 Armor 人口 gate 並取政府加分：(%d,%v)", score, exact)
	}
	ctx.strategicPressureContextKnown = false
	if _, exact := originalAIExactBuildingScore(armor, colony, ai.PersonalityRuthless, ctx); exact {
		t.Fatal("缺 session-wide context 時不得冒稱 Armor Barracks exact")
	}
}

func TestAIOriginalFuelRangeTable(t *testing.T) {
	tests := []struct {
		tech gamedata.Technology
		want int
	}{
		{gamedata.TECH_STANDARD_FUEL_CELLS, 4},
		{gamedata.TECH_DEUTERIUM_FUEL_CELLS, 6},
		{gamedata.TECH_IRIDIUM_FUEL_CELLS, 9},
		{gamedata.TECH_URRIDIUM_FUEL_CELLS, 12},
		{gamedata.TECH_THORIUM_FUEL_CELLS, 255},
	}
	for _, tt := range tests {
		got, ok := originalAIFuelRangeParsecs(engine.PlayerState{
			GrantedTechs: map[gamedata.Technology]bool{tt.tech: true},
		})
		if !ok || got != tt.want {
			t.Errorf("fuel tech %d range=(%d,%v)，want (%d,true)", tt.tech, got, ok, tt.want)
		}
	}
	if _, ok := originalAIFuelRangeParsecs(engine.PlayerState{}); ok {
		t.Fatal("沒有已知 fuel application 時不得擅自補 Standard Fuel Cells")
	}
}

func TestAIOriginalPlanetaryShieldScores(t *testing.T) {
	barrier, _ := gamedata.BuildingByNameZH(gamedata.BuildingPlanetaryBarrierShield)
	flux, _ := gamedata.BuildingByNameZH(gamedata.BuildingPlanetaryFluxShield)
	radiation, _ := gamedata.BuildingByNameZH(gamedata.BuildingPlanetaryRadiationShield)
	ctx := originalAIBuildScoreContext{
		strategicPressureContextKnown: true,
		reachTreatyNear:               1,
		reachNoPolicyNear:             2,
		reachWarNear:                  3,
		reachExtended:                 4,
		incomingOtherFleetETA9:        true,
	}
	colony := engine.ColonyState{Climate: gamedata.RADIATED}
	for _, b := range []gamedata.Building{barrier, flux} {
		if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityRuthless, ctx); !exact || score != 64 {
			t.Errorf("%s score=(%d,%v)，want (64,true)", b.NameZH, score, exact)
		}
	}
	if score, exact := originalAIExactBuildingScore(radiation, colony, ai.PersonalityRuthless, ctx); !exact || score != 83 {
		t.Fatalf("Radiation Shield pressure score=(%d,%v)，want (83,true)", score, exact)
	}
	ctx = originalAIBuildScoreContext{strategicPressureContextKnown: true, priorityGate: true}
	if score, exact := originalAIExactBuildingScore(radiation, colony, ai.PersonalityPacifist, ctx); !exact || score != 0 {
		t.Fatalf("priority gate 無 ETA9 應先歸零：(%d,%v)", score, exact)
	}
	ctx.priorityGate = false
	if score, exact := originalAIExactBuildingScore(radiation, colony, ai.PersonalityPacifist, ctx); !exact || score != 2 {
		t.Fatalf("Radiated Pacifist bonus=(%d,%v)，want (2,true)", score, exact)
	}
	ctx.treasuryBefore, ctx.netBC = 1500, 64
	if score, exact := originalAIExactBuildingScore(barrier, colony, ai.PersonalityPacifist, ctx); !exact || score != 1 {
		t.Fatalf("零 pressure 的 Barrier 應只取 budget factor：(%d,%v)", score, exact)
	}
}

func TestAIPlanetaryShieldCandidateCompletionAndConsumers(t *testing.T) {
	tests := []struct {
		name      string
		topic     gamedata.ResearchTopic
		reduction int
	}{
		{gamedata.BuildingPlanetaryRadiationShield, gamedata.TOPIC_MAGNETO_GRAVITICS, 5},
		{gamedata.BuildingPlanetaryFluxShield, gamedata.TOPIC_QUANTUM_FIELDS, 10},
		{gamedata.BuildingPlanetaryBarrierShield, gamedata.TOPIC_TEMPORAL_FIELDS, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDemoSession()
			s.ensureAIAIState()
			a := &s.AIPlayers[0]
			a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{tt.topic: true}
			a.Colonies[0].Climate = gamedata.RADIATED
			if p := s.aiColonyPlanet(0, 0); p != nil {
				syncPlanetClimate(p, gamedata.RADIATED)
			}
			a.ColonyBuildings[0] = make(map[string]bool)
			for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
				a.ColonyBuildings[0][b.NameZH] = true
			}
			delete(a.ColonyBuildings[0], tt.name)
			out := engine.EmpireOutput{
				Colonies: []engine.ColonyOutput{{NetIndustry: 600}},
				Player:   engine.PlayerState{BC: 2001}, NetBC: 1,
			}
			build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1, s.originalAIStrategicPressureContext(0, 0))
			if !ok || build.Name != tt.name {
				t.Fatalf("唯一候選錯誤：build=%+v ok=%v", build, ok)
			}
			a.ColonyBuilds = map[int]ColonyBuild{aiColonyBuildKey(a, 0): {Name: tt.name, Cost: 1}}
			s.advanceAIColonyBuilds(0, out)
			if got := gamedata.PlanetaryShieldReduction(a.ColonyBuildings[0]); got != tt.reduction {
				t.Fatalf("逐發轟炸減傷=%d，want %d；built=%v", got, tt.reduction, a.ColonyBuildings[0])
			}
			if tt.name != gamedata.BuildingPlanetaryRadiationShield &&
				a.ColonyBuildings[0][gamedata.BuildingPlanetaryRadiationShield] {
				t.Fatal("較高階護盾完工後仍殘留 Radiation Shield")
			}
			if tt.name == gamedata.BuildingPlanetaryBarrierShield &&
				a.ColonyBuildings[0][gamedata.BuildingPlanetaryFluxShield] {
				t.Fatal("Barrier Shield 完工後仍殘留 Flux Shield")
			}
			if a.Colonies[0].Climate != gamedata.BARREN {
				t.Fatalf("AI colony 氣候=%v，want Barren", a.Colonies[0].Climate)
			}
			if p := s.aiColonyPlanet(0, 0); p != nil && p.ClimateID != gamedata.BARREN {
				t.Fatalf("全局 planet 氣候=%v，want Barren", p.ClimateID)
			}
		})
	}
}

func TestAIOriginalOrbitalBaseScores(t *testing.T) {
	starBase, _ := gamedata.BuildingByNameZH("星基")
	battlestation, _ := gamedata.BuildingByNameZH("戰鬥站")
	starFortress, _ := gamedata.BuildingByNameZH("星辰要塞")
	ctx := originalAIBuildScoreContext{
		strategicPressureContextKnown: true,
		reachTreatyNear:               1,
		reachNoPolicyNear:             2,
		reachWarNear:                  3,
		reachExtended:                 4,
		incomingOtherFleetETA9:        true,
		commandPointsSupply:           5,
		usedCommandPoints:             7,
	}
	if score, exact := originalAIExactBuildingScore(starBase, engine.ColonyState{}, ai.PersonalityRuthless, ctx); !exact || score != 94 {
		t.Fatalf("Star Base score=(%d,%v)，want (94,true)", score, exact)
	}
	for _, b := range []gamedata.Building{battlestation, starFortress} {
		if score, exact := originalAIExactBuildingScore(b, engine.ColonyState{}, ai.PersonalityRuthless, ctx); !exact || score != 77 {
			t.Errorf("%s score=(%d,%v)，want (77,true)", b.NameZH, score, exact)
		}
	}

	ctx = originalAIBuildScoreContext{
		strategicPressureContextKnown: true,
		priorityGate:                  true,
		commandPointsSupply:           5,
		usedCommandPoints:             99,
	}
	if score, exact := originalAIExactBuildingScore(starBase, engine.ColonyState{}, ai.PersonalityRuthless, ctx); !exact || score != 0 {
		t.Fatalf("priority gate 無 ETA9 必須先歸零，不得吃指揮赤字：(%d,%v)", score, exact)
	}
	ctx.priorityGate = false
	ctx.usedCommandPoints = 5
	if score, exact := originalAIExactBuildingScore(starBase, engine.ColonyState{}, ai.PersonalityRuthless, ctx); !exact || score != 2 {
		t.Fatalf("恰好用滿供給時 command deficit 1 再加 Ruthless：(%d,%v)，want 2", score, exact)
	}
	ctx.commandPointsSupply = 6
	if score, exact := originalAIExactBuildingScore(starBase, engine.ColonyState{}, ai.PersonalityRuthless, ctx); !exact || score != 0 {
		t.Fatalf("尚餘一點供給時不得加 deficit 或 Ruthless：(%d,%v)", score, exact)
	}
}

func TestAIOrbitalBaseCandidateCompletionReplacementAndConsumers(t *testing.T) {
	tests := []struct {
		name        string
		topic       gamedata.ResearchTopic
		lower       []string
		wantCommand int
		wantScan    int
	}{
		{"星基", gamedata.TOPIC_ENGINEERING, nil, 1, 2},
		{"戰鬥站", gamedata.TOPIC_ROBOTICS, []string{"星基"}, 2, 4},
		{"星辰要塞", gamedata.TOPIC_SUPERSCALAR_CONSTRUCTION, []string{"星基", "戰鬥站"}, 3, 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDemoSession()
			a := &s.AIPlayers[0]
			a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{tt.topic: true}
			a.ColonyBuildings[0] = make(map[string]bool)
			for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
				if b.NameZH != tt.name {
					a.ColonyBuildings[0][b.NameZH] = true
				}
			}
			for _, name := range tt.lower {
				a.ColonyBuildings[0][name] = true
			}
			out := engine.EmpireOutput{
				Colonies: []engine.ColonyOutput{{NetIndustry: 3000}},
				Player:   engine.PlayerState{CommandPointsSupply: 5, UsedCommandPoints: 5},
			}
			build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1,
				originalAIStrategicPressureContext{known: true})
			if !ok || build.Name != tt.name {
				t.Fatalf("唯一軌道基地候選錯誤：build=%+v ok=%v built=%v", build, ok, a.ColonyBuildings[0])
			}
			a.ColonyBuilds = map[int]ColonyBuild{aiColonyBuildKey(a, 0): {Name: tt.name, Cost: 1}}
			s.advanceAIColonyBuilds(0, out)
			if !a.ColonyBuildings[0][tt.name] {
				t.Fatalf("%s 完工未寫入建築 map", tt.name)
			}
			for _, name := range tt.lower {
				if a.ColonyBuildings[0][name] {
					t.Fatalf("%s 完工後仍殘留低階基地 %s", tt.name, name)
				}
			}
			if got := gamedata.CommandPointsFromBuildings(a.ColonyBuildings[0]); got != tt.wantCommand {
				t.Fatalf("指揮評等 consumer=%d，want %d", got, tt.wantCommand)
			}
			if got := gamedata.OrbitalScannerBonusParsec(a.ColonyBuildings[0]); got != tt.wantScan {
				t.Fatalf("掃描範圍 consumer=%d，want %d", got, tt.wantScan)
			}
		})
	}
}

func TestAIOrbitalBaseDoesNotDowngrade(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{
		gamedata.TOPIC_ENGINEERING:              true,
		gamedata.TOPIC_ROBOTICS:                 true,
		gamedata.TOPIC_SUPERSCALAR_CONSTRUCTION: true,
	}
	a.ColonyBuildings[0] = map[string]bool{"星辰要塞": true}
	for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
		if b.NameZH != "星基" && b.NameZH != "戰鬥站" {
			a.ColonyBuildings[0][b.NameZH] = true
		}
	}
	out := engine.EmpireOutput{
		Colonies: []engine.ColonyOutput{{NetIndustry: 100}},
		Player:   engine.PlayerState{CommandPointsSupply: 5, UsedCommandPoints: 99},
	}
	if build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1,
		originalAIStrategicPressureContext{known: true}); ok {
		t.Fatalf("已有 Star Fortress 時不得降級候選：%+v", build)
	}
	if !aiOrbitalBaseSuperseded("星基", a.ColonyBuildings[0]) ||
		!aiOrbitalBaseSuperseded("戰鬥站", a.ColonyBuildings[0]) {
		t.Fatal("高階基地的 typed 取代 gate 未覆蓋兩個低階產品")
	}
}

func TestAIBarracksCandidateCompletionAndGroundForceConsumer(t *testing.T) {
	tests := []struct {
		name       string
		building   string
		topic      gamedata.ResearchTopic
		wantMarine bool
	}{
		{"Marine Barracks", marineBarracksBuildingName, gamedata.TOPIC_ENGINEERING, true},
		{"Armor Barracks", armorBarracksBuildingName, gamedata.TOPIC_ASTRO_ENGINEERING, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewDemoSession()
			s.ensureAIAIState()
			a := &s.AIPlayers[0]
			a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{tt.topic: true}
			a.Colonies[0].Population = 8
			a.Colonies[0].PopMax = 12
			a.Colonies[0].OwnerRaceSlot = a.PopulationRaceSlot
			a.Colonies[0].OwnerRaceSlotKnown = true
			a.Colonies[0].OwnerRaceProfileKnown = true
			a.Colonies[0].PopulationGroups = []engine.PopulationGroup{{
				RaceSlot: a.PopulationRaceSlot, RaceSlotKnown: true, ProfileKnown: true, Workers: 8,
			}}
			a.ColonyBuildings[0] = make(map[string]bool)
			for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
				a.ColonyBuildings[0][b.NameZH] = true
			}
			delete(a.ColonyBuildings[0], tt.building)
			out := engine.EmpireOutput{Colonies: []engine.ColonyOutput{{NetIndustry: 200}}, Player: engine.PlayerState{BC: 2000}}
			build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1, s.originalAIStrategicPressureContext(0, 0))
			if !ok || build.Name != tt.building {
				t.Fatalf("唯一候選錯誤：build=%+v ok=%v", build, ok)
			}
			a.ColonyMarines, a.ColonyTanks = nil, nil
			a.MarineBarracksAge, a.ArmorBarracksAge = nil, nil
			a.ColonyBuilds = map[int]ColonyBuild{aiColonyBuildKey(a, 0): {Name: tt.building, Cost: 1}}
			s.advanceAIColonyBuilds(0, out)
			advanceAIGroundForces(a)
			if !a.ColonyBuildings[0][tt.building] {
				t.Fatalf("%s 完工未寫入建築 map", tt.building)
			}
			if tt.wantMarine && a.ColonyMarines[0] <= 0 {
				t.Fatalf("Marine Barracks 完工後駐軍 consumer 未產生陸戰隊：%v", a.ColonyMarines)
			}
			if !tt.wantMarine && a.ColonyTanks[0] <= 0 {
				t.Fatalf("Armor Barracks 完工後駐軍 consumer 未產生戰車營：%v", a.ColonyTanks)
			}
		})
	}
}

func TestAIPlanetaryGravityGeneratorCandidateCompletionAndConsumer(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{gamedata.TOPIC_ARTIFICIAL_GRAVITY: true}
	a.Colonies[0] = completeAIFoodScoreColony()
	a.Colonies[0].PlanetGravity = gamedata.NORMAL_G
	a.Colonies[0].IndustryPerWorker = 4
	a.Colonies[0].PopulationGroups[0].Gravity = gamedata.LOW_G
	a.ColonyBuildings[0] = make(map[string]bool)
	for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
		a.ColonyBuildings[0][b.NameZH] = true
	}
	delete(a.ColonyBuildings[0], "行星重力產生器")
	gravityBuilding, found := gamedata.BuildingByNameZH("行星重力產生器")
	if !found {
		t.Fatal("行星重力產生器不存在")
	}
	ctx := originalAIBuildScoreContext{
		ownerLowGravity:  aiRaceHasTrait(*a, gamedata.TRAIT_LOW_G),
		ownerHighGravity: aiRaceHasTrait(*a, gamedata.TRAIT_HIGH_G),
	}
	if score, exact := originalAIExactBuildingScore(gravityBuilding, a.Colonies[0], a.Personality, ctx); !exact || score <= 0 {
		t.Fatalf("測試前提：Low-G owner 在 Normal-G 星球的重力產生器應有正分：score=%d exact=%v low=%v high=%v",
			score, exact, ctx.ownerLowGravity, ctx.ownerHighGravity)
	}
	before := engine.RunColonyTurn(a.Colonies[0])
	out := engine.EmpireOutput{Colonies: []engine.ColonyOutput{{NetIndustry: 20}}}
	build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1)
	if !ok || build.Name != "行星重力產生器" {
		t.Fatalf("唯一候選錯誤：build=%+v ok=%v", build, ok)
	}
	key := aiColonyBuildKey(a, 0)
	a.ColonyBuilds = map[int]ColonyBuild{key: {Name: "行星重力產生器", Cost: 1}}
	s.advanceAIColonyBuilds(0, out)
	after := engine.RunColonyTurn(a.Colonies[0])
	if !a.ColonyBuildings[0]["行星重力產生器"] || !a.Colonies[0].NormalizeGravity {
		t.Fatalf("完工未寫入建築／NormalizeGravity：built=%v normalize=%v",
			a.ColonyBuildings[0]["行星重力產生器"], a.Colonies[0].NormalizeGravity)
	}
	if after.GrossIndustry <= before.GrossIndustry {
		t.Fatalf("完工後重力 consumer 未改善工業：before=%d after=%d", before.GrossIndustry, after.GrossIndustry)
	}
}

func TestAISoilEnrichmentCandidateAndCompletion(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Personality = ai.PersonalityPacifist
	a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{gamedata.TOPIC_ADVANCED_BIOLOGY: true}
	a.ColonyBuildings[0] = make(map[string]bool)
	for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
		a.ColonyBuildings[0][b.NameZH] = true
	}
	a.Colonies[0].Population = 4
	a.Colonies[0].Farmers = 0
	a.Colonies[0].Workers = 4
	a.Colonies[0].Scientists = 0
	a.Colonies[0].FoodPerFarmer = 1
	a.Colonies[0].OwnerRaceSlot = 1
	a.Colonies[0].OwnerRaceSlotKnown = true
	a.Colonies[0].OwnerRaceProfileKnown = true
	a.Colonies[0].PopulationGroups = []engine.PopulationGroup{{
		RaceSlot: 1, RaceSlotKnown: true, ProfileKnown: true, Workers: 4,
	}}
	out := engine.EmpireOutput{Colonies: []engine.ColonyOutput{{NetIndustry: 1}}, Player: engine.PlayerState{BC: 1499}}
	a.Colonies[0].Climate = gamedata.BARREN
	if build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1); ok {
		t.Fatalf("Barren 不得選 Soil Enrichment：%+v", build)
	}
	a.Colonies[0].Climate = gamedata.TERRAN
	actions := gamedata.AvailableSpecialActions(a.Player.CompletedTopics)
	foundSoil := false
	for _, action := range actions {
		foundSoil = foundSoil || action.NameZH == gamedata.SoilEnrichmentActionName
	}
	if !foundSoil {
		t.Fatal("完成 Advanced Biology 後 Soil Enrichment 未進 AvailableSpecialActions")
	}
	proxy := gamedata.Building{NameZH: gamedata.SoilEnrichmentActionName, NameEN: "Soil Enrichment"}
	ctx := originalAIBuildScoreContext{empireFoodBalanceHalf: out.TotalFoodHalf}
	if score, exact := originalAIExactBuildingScore(proxy, a.Colonies[0], a.Personality, ctx); !exact || score != 5 {
		t.Fatalf("候選前 Soil Enrichment 分數=(%d,%v)，want (5,true)", score, exact)
	}
	build, ok := chooseAIColonyBuilding(a, 0, out, 2, 1)
	if !ok || build.Name != gamedata.SoilEnrichmentActionName {
		t.Fatalf("Terran 的唯一 Special 候選應是 Soil Enrichment：build=%+v ok=%v", build, ok)
	}
	key := aiColonyBuildKey(a, 0)
	a.ColonyBuilds = map[int]ColonyBuild{key: {
		Name: gamedata.SoilEnrichmentActionName, Progress: 119, Cost: 120,
	}}
	s.advanceAIColonyBuilds(0, out)
	if a.Colonies[0].FoodPerFarmer != 2 {
		t.Fatalf("Soil Enrichment 完工 FoodPerFarmer=%d，want 2", a.Colonies[0].FoodPerFarmer)
	}
	if a.ColonyBuildings[0][gamedata.SoilEnrichmentActionName] {
		t.Fatal("Soil Enrichment 是 Special，不得殘留於 ColonyBuildings")
	}
}

func TestAIOriginalPriorityBuildingGate(t *testing.T) {
	known := map[gamedata.Technology]bool{gamedata.TECH_AUTOMATED_FACTORIES: true}
	if !aiOriginalPriorityBuildingGate(engine.ColonyState{MineralRichness: gamedata.ABUNDANT}, nil, known, gamedata.MoraleGovDemocracy) {
		t.Fatal("Abundant 且已知但未建 Automated Factory 應觸發優先 gate")
	}
	if aiOriginalPriorityBuildingGate(engine.ColonyState{MineralRichness: gamedata.RICH}, nil, known, gamedata.MoraleGovDictatorship) {
		t.Fatal("Rich 超過原版 <=2 礦產邊界，不應觸發工廠 gate")
	}
	if aiOriginalPriorityBuildingGate(engine.ColonyState{MineralRichness: gamedata.POOR}, map[string]bool{"自動工廠": true}, known, gamedata.MoraleGovDictatorship) {
		t.Fatal("已建 Automated Factory 不應再次觸發 gate")
	}

	known = map[gamedata.Technology]bool{gamedata.TECH_MARINE_BARRACKS: true}
	if !aiOriginalPriorityBuildingGate(engine.ColonyState{MineralRichness: gamedata.RICH}, nil, known, gamedata.MoraleGovImperium) {
		t.Fatal("Imperium 已知但未建 Marine Barracks 應觸發 gate")
	}
	if aiOriginalPriorityBuildingGate(engine.ColonyState{MineralRichness: gamedata.RICH}, nil, known, gamedata.MoraleGovDemocracy) {
		t.Fatal("Democracy 不在 raw government/2 <=1 gate")
	}
	if aiOriginalPriorityBuildingGate(engine.ColonyState{MineralRichness: gamedata.RICH}, map[string]bool{"海軍陸戰隊營": true}, known, gamedata.MoraleGovFeudalism) {
		t.Fatal("已建 Marine Barracks 不應再次觸發 gate")
	}

	ps := engine.PlayerState{
		CompletedTopics: map[gamedata.ResearchTopic]bool{gamedata.TOPIC_ADVANCED_CONSTRUCTION: true},
		ChosenTech:      map[gamedata.ResearchTopic]gamedata.Technology{gamedata.TOPIC_ADVANCED_CONSTRUCTION: gamedata.TECH_HEAVY_ARMOR},
		ExplicitChoice:  map[gamedata.ResearchTopic]bool{gamedata.TOPIC_ADVANCED_CONSTRUCTION: true},
	}
	if aiOriginalPriorityBuildingGate(engine.ColonyState{MineralRichness: gamedata.POOR}, nil,
		knownTechnologyApplications(ps), gamedata.MoraleGovDictatorship) {
		t.Fatal("完成主題但明確選 Heavy Armor，不得誤認已知 Automated Factories")
	}
	ps.ChosenTech[gamedata.TOPIC_ADVANCED_CONSTRUCTION] = gamedata.TECH_AUTOMATED_FACTORIES
	if !aiOriginalPriorityBuildingGate(engine.ColonyState{MineralRichness: gamedata.POOR}, nil,
		knownTechnologyApplications(ps), gamedata.MoraleGovDictatorship) {
		t.Fatal("明確選 Automated Factories 後應觸發工廠 gate")
	}
}

func TestAIOriginalLateTechReachedUsesSelectedOrCompletedHyperField(t *testing.T) {
	ps := engine.PlayerState{ResearchTopic: gamedata.TOPIC_HYPER_PHYSICS}
	if !aiOriginalLateTechReached(ps) {
		t.Fatal("選中 raw 75..82 的 Hyper field 時應立即進入晚期科技")
	}
	ps = engine.PlayerState{CompletedTopics: map[gamedata.ResearchTopic]bool{gamedata.TOPIC_HYPER_BIOLOGY: true}}
	if !aiOriginalLateTechReached(ps) {
		t.Fatal("完成任一 Hyper field 後晚期科技狀態應持續")
	}
	ps = engine.PlayerState{HyperAdvancedLevels: map[gamedata.ResearchTopic]int{gamedata.TOPIC_HYPER_SOCIOLOGY: 2}}
	if !aiOriginalLateTechReached(ps) {
		t.Fatal("舊存檔只保留 Hyper 等級時仍應重建晚期科技狀態")
	}
	if aiOriginalLateTechReached(engine.PlayerState{ResearchTopic: gamedata.TOPIC_ADVANCED_CONSTRUCTION}) {
		t.Fatal("一般研究 field 不得誤判為晚期科技")
	}
}

func TestAIColonyBuildConsumesIndustryBeforeShipPool(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Colonies = []engine.ColonyState{{Population: 4, PopMax: 10, IndustryPerWorker: 2}}
	a.ColonyStars = []int{7}
	a.ColonyBuildings = []map[string]bool{{}}
	a.ColonyBuilds = nil
	a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{
		gamedata.TOPIC_ADVANCED_CONSTRUCTION: true,
	}
	// 此主題同時開多棟建築；把其餘候選標成已完成，讓測試只驗證逐殖民地
	// 產品與產能消費，不把近似分數表的選擇偏好凍成契約。
	for _, b := range gamedata.AvailableBuildings(a.Player.CompletedTopics) {
		if b.NameZH != "自動工廠" {
			a.ColonyBuildings[0][b.NameZH] = true
		}
	}
	out := engine.EmpireOutput{Colonies: []engine.ColonyOutput{{NetIndustry: 30}}}
	if got := s.advanceAIColonyBuilds(0, out); got != 0 {
		t.Fatalf("建築產品應先消費殖民地工業，交給造艦=%d", got)
	}
	build := a.ColonyBuilds[7]
	if build.Name != "自動工廠" || build.Progress != 30 || build.Cost != 60 {
		t.Fatalf("AI 殖民地產品=%+v，want 自動工廠 30/60", build)
	}
	if got := s.advanceAIColonyBuilds(0, out); got != 0 {
		t.Fatalf("完工回合不應重複把同份產能交給造艦，got %d", got)
	}
	if !a.ColonyBuildings[0]["自動工廠"] || a.Colonies[0].FlatIndustry != 5 || a.Colonies[0].IndustryPerWorker != 3 {
		t.Fatalf("完工建築未接到 typed 效果：building=%v colony=%+v", a.ColonyBuildings[0], a.Colonies[0])
	}
	if _, ok := a.ColonyBuilds[7]; ok {
		t.Fatal("完工後應清除目前產品")
	}
}

func TestAIColonyBuildSurvivesSaveLoad(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers[0].ColonyBuilds = map[int]ColonyBuild{
		5: {Name: "自動工廠", Cost: 60, Progress: 37},
	}
	path := filepath.Join(t.TempDir(), "ai-colony-build.json")
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSession(path)
	if err != nil {
		t.Fatal(err)
	}
	if build := got.AIPlayers[0].ColonyBuilds[5]; build.Name != "自動工廠" || build.Cost != 60 || build.Progress != 37 {
		t.Fatalf("AI 殖民地產品未完整往返：%+v", build)
	}
}

func TestAIColonyWithoutBuildableBuildingFeedsShipProduction(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Colonies = []engine.ColonyState{{Population: 4, PopMax: 10}}
	a.ColonyStars = []int{3}
	a.ColonyBuildings = []map[string]bool{{}}
	a.Player.CompletedTopics = nil
	out := engine.EmpireOutput{Colonies: []engine.ColonyOutput{{NetIndustry: 17}}}
	if got := s.advanceAIColonyBuilds(0, out); got != 17 {
		t.Fatalf("無可建建築時造艦產能=%d，want 17", got)
	}
}
