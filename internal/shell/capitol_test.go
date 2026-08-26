package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/ai"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestNewGameCapitolStateUsesPlanetIndex(t *testing.T) {
	s := NewDemoSession()
	if !s.PlayerCapitolPlanetKnown || s.PlayerCapitolPlanet != s.PlayerColonyPlanets[0] ||
		!s.ColonyBuildings[0][CapitolBuildName] || s.PlayerCapitolRebuildRequired {
		t.Fatalf("玩家開局 Capitol 狀態不一致：planet=%d known=%v rebuild=%v buildings=%v",
			s.PlayerCapitolPlanet, s.PlayerCapitolPlanetKnown, s.PlayerCapitolRebuildRequired, s.ColonyBuildings[0])
	}
	for i := range s.AIPlayers {
		a := &s.AIPlayers[i]
		if !a.CapitolPlanetKnown || a.CapitolPlanet != a.ColonyPlanets[0] ||
			!a.ColonyBuildings[0][CapitolBuildName] || a.CapitolRebuildRequired {
			t.Fatalf("AI %d 開局 Capitol 狀態不一致：%+v buildings=%v", i, *a, a.ColonyBuildings[0])
		}
	}
}

func TestReplacementCapitolPlanetUsesPopulationAndLowerIndexTieBreak(t *testing.T) {
	colonies := []engine.ColonyState{{Population: 4}, {Population: 9}, {Population: 9}, {Population: 12}}
	planets := []int{10, 11, 12, 13}
	if got := replacementCapitolPlanet(colonies, planets, 3); got != 11 {
		t.Fatalf("最高人口同分應選較低殖民地索引的行星 11，got %d", got)
	}
	if got := replacementCapitolPlanet(colonies[:1], planets[:1], 0); got != -1 {
		t.Fatalf("沒有剩餘殖民地應回 -1，got %d", got)
	}
}

func TestCapturedCapitolReassignsOldOwnerAndDoesNotTransferBuilding(t *testing.T) {
	s := NewDemoSession()
	s.PlayerCapitolPlanet, s.PlayerCapitolPlanetKnown = 5, true
	a := &s.AIPlayers[0]
	a.Colonies = []engine.ColonyState{{Population: 12}, {Population: 7}, {Population: 7}}
	a.ColonyPlanets = []int{20, 21, 22}
	a.ColonyStars = []int{2, 3, 4}
	a.ColonyBuildings = []map[string]bool{
		{CapitolBuildName: true, "星基": true}, {"自動工廠": true}, {"研究實驗室": true},
	}
	a.CapitolPlanet, a.CapitolPlanetKnown, a.CapitolRebuildRequired = 20, true, false

	transferred := s.prepareCapturedAIColony(0, 0, 20)
	if transferred[CapitolBuildName] || !transferred["星基"] {
		t.Fatalf("Capitol 不得過戶，其他建築應保留：%v", transferred)
	}
	if a.CapitolPlanet != 21 || !a.CapitolRebuildRequired {
		t.Fatalf("舊擁有者應指定人口同分中較低索引行星 21 並進入重建：planet=%d rebuild=%v",
			a.CapitolPlanet, a.CapitolRebuildRequired)
	}
}

func TestCapitolBuildOptionAndMoralePenaltyOnlyAtDesignatedPlanet(t *testing.T) {
	s := NewDemoSession()
	s.PlayerColonies = append(s.PlayerColonies, s.PlayerColonies[0])
	s.PlayerColonyPlanets = append(s.PlayerColonyPlanets, s.PlayerCapitolPlanet+1)
	s.ColonyBuildings = append(s.ColonyBuildings, map[string]bool{})
	delete(s.ColonyBuildings[0], CapitolBuildName)
	s.PlayerCapitolRebuildRequired = true
	s.recalcAllColonyMorale()

	wantPenalty := gamedata.MoraleCapitalCapturedPenalty(s.effectiveGovernment())
	base := colonyMoralePercent(s.effectiveGovernment(), s.ColonyBuildings[0], false,
		achievementMoralePercent(s.Player, s.effectiveGovernment()))
	if got := s.PlayerColonies[0].MoralePercent; got != base+wantPenalty {
		t.Fatalf("失都士氣=%d，want %d", got, base+wantPenalty)
	}
	if !hasBuildOption(s.AvailableBuildOptionsForColony(0), CapitolBuildName) ||
		hasBuildOption(s.AvailableBuildOptionsForColony(1), CapitolBuildName) {
		t.Fatal("Capitol 只能出現在指定重建行星")
	}

	s.Builds[0] = ColonyBuild{Name: CapitolBuildName, Cost: CapitolProductionCost, Progress: CapitolProductionCost}
	s.completeColonyBuild(0)
	if s.PlayerCapitolRebuildRequired || !s.ColonyBuildings[0][CapitolBuildName] {
		t.Fatalf("完工應清除重建狀態並寫入建築：rebuild=%v buildings=%v",
			s.PlayerCapitolRebuildRequired, s.ColonyBuildings[0])
	}
}

func TestAIRaw9ScoreCandidateCompletionAndSnapshot(t *testing.T) {
	s := NewDemoSession()
	a := &s.AIPlayers[0]
	a.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{}
	delete(a.ColonyBuildings[0], CapitolBuildName)
	a.CapitolRebuildRequired = true
	ctx := originalAIBuildScoreContext{government: effectiveAIGovernment(a), capitolStateKnown: true,
		capitolPlanetMatches: true}
	proxy := gamedata.Building{NameZH: CapitolBuildName, NameEN: CapitolBuildName}
	if score, exact := originalAIExactBuildingScore(proxy, a.Colonies[0], ai.PersonalityPacifist, ctx); !exact || score != 100 {
		t.Fatalf("raw 9 score=(%d,%v)，want (100,true)", score, exact)
	}
	out := engine.EmpireOutput{Colonies: []engine.ColonyOutput{{NetIndustry: CapitolProductionCost}}, Player: a.Player}
	s.advanceAIColonyBuilds(0, out)
	if a.CapitolRebuildRequired || !a.ColonyBuildings[0][CapitolBuildName] {
		t.Fatalf("AI Capitol 應完成：rebuild=%v buildings=%v", a.CapitolRebuildRequired, a.ColonyBuildings[0])
	}

	got := s.snapshot().restore()
	if got.AIPlayers[0].CapitolPlanet != a.CapitolPlanet || !got.AIPlayers[0].CapitolPlanetKnown ||
		got.AIPlayers[0].CapitolRebuildRequired {
		t.Fatalf("存讀檔未保留 AI Capitol 狀態：%+v", got.AIPlayers[0])
	}
}

func hasBuildOption(options []ColonyBuild, name string) bool {
	for _, option := range options {
		if option.Name == name {
			return true
		}
	}
	return false
}
