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
	if score, exact := originalAIExactBuildingScore(b, colony, ai.PersonalityPacifist, originalAIBuildScoreContext{empireFoodBalanceHalf: -1}); !exact || score != 0 {
		t.Fatalf("主要人口非 Lithovore 時分數=(%d,%v)，want (0,true)", score, exact)
	}
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
