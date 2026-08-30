package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func raceIndexByEnglishNameForTest(t *testing.T, name string) int {
	t.Helper()
	for i, race := range Races {
		if race.EnName == name {
			return i
		}
	}
	t.Fatalf("找不到種族 %s", name)
	return -1
}

func TestFormalNewGameFinalizesStartingTechAfterBuiltInRace(t *testing.T) {
	s := NewDemoSession()
	s.TechLevel, s.TechLevelSet = 2, true
	s.SetupNewGame(20, 4242, 3)
	if !s.newGameRacePending {
		t.Fatal("SetupNewGame 後應等待種族／政府完成")
	}
	s.ApplyRace(raceIndexByEnglishNameForTest(t, "Humans"))
	if s.newGameRacePending {
		t.Fatal("內建種族套用後應已完成開局科技")
	}
	if s.Government != gamedata.MoraleGovDemocracy {
		t.Fatalf("人類原版政體應為民主，得到 %d", s.Government)
	}
	got := 0
	for topic := range s.Player.CompletedTopics {
		if topic != gamedata.TOPIC_STARTING_TECH {
			got++
		}
	}
	if got != 25 {
		t.Fatalf("種族完成後應重建為 25 個開局主題，得到 %d", got)
	}
	if s.Player.CompletedTopics[s.Player.ResearchTopic] {
		t.Fatalf("開局重建後目前研究不得停在已完成主題 %d", s.Player.ResearchTopic)
	}
}

func TestGovernmentChangeAfterOpeningDoesNotResetResearch(t *testing.T) {
	s := NewDemoSession()
	s.TechLevel, s.TechLevelSet = 2, true
	s.SetupNewGame(20, 4242, 3)
	s.ApplyRace(raceIndexByEnglishNameForTest(t, "Humans"))
	sentinel := gamedata.TOPIC_HYPER_BIOLOGY
	s.Player.CompletedTopics[sentinel] = true
	s.Player.ResearchProgress = 777
	s.ApplyGovernment(0)
	if !s.Player.CompletedTopics[sentinel] || s.Player.ResearchProgress != 777 {
		t.Fatal("開局 pending 結束後換政府不得重置玩家研究")
	}
}

func TestAdvancedCivilizationStartingBCAppliesAfterRaceFinalization(t *testing.T) {
	s := NewDemoSession()
	s.TechLevel, s.TechLevelSet = 2, true
	s.SetupNewGame(20, 4242, 3)
	s.ApplyRace(raceIndexByEnglishNameForTest(t, "Gnolams"))
	if got, want := s.Player.BC, gamedata.AdvancedCivilizationStartingBC(2); got != want {
		t.Fatalf("先進諾蘭姆開局 BC=%d，預期 %d", got, want)
	}
	for i := range s.AIPlayers {
		raw := 0
		if raceIdx := aiRaceIndex(s.AIPlayers[i]); raceIdx >= 0 && raceIdx < len(Races) {
			if traits, ok := gamedata.OrigRaceTraits(Races[raceIdx].OrigIdx); ok {
				raw = int(traits[gamedata.TRAIT_MONEY])
			}
		}
		if got, want := s.AIPlayers[i].Player.BC, gamedata.AdvancedCivilizationStartingBC(raw); got != want {
			t.Errorf("AI %d 先進開局 BC=%d，預期 %d", i, got, want)
		}
	}

	standard := NewDemoSession()
	standard.TechLevel, standard.TechLevelSet = TechLevelDefault, true
	standard.SetupNewGame(20, 4242, 2)
	standard.ApplyRace(raceIndexByEnglishNameForTest(t, "Gnolams"))
	if standard.Player.BC != 50 {
		t.Fatalf("一般文明開局不得套先進國庫，得到 %d BC", standard.Player.BC)
	}
}

func TestCustomRaceBuildsCompleteRuntimeTraitArray(t *testing.T) {
	s := NewDemoSession()
	r := Race{OrigIdx: -1, GrowthPct: 50, FoodBonus: 2, IndBonus: 1, ResBonus: -1,
		IncomePerPop: 2, CombatPct: 20, ShipDefPct: 25, GroundCombatBonus: 10, SpyBonus: -10}
	s.ApplyCustomRaceBonuses(r, gamedata.TRAIT_CREATIVE, gamedata.TRAIT_TELEPATHIC)
	traits := s.CustomRaceRuntimeTraits
	want := map[gamedata.RaceTrait]int8{
		gamedata.TRAIT_POPULATION: 50, gamedata.TRAIT_FARMING: 2,
		gamedata.TRAIT_INDUSTRY: 1, gamedata.TRAIT_SCIENCE: -1,
		gamedata.TRAIT_MONEY: 2, gamedata.TRAIT_SHIP_ATTACK: 20,
		gamedata.TRAIT_SHIP_DEFENSE: 25, gamedata.TRAIT_GROUND_COMBAT: 10,
		gamedata.TRAIT_SPYING: -10, gamedata.TRAIT_CREATIVE: 1,
		gamedata.TRAIT_TELEPATHIC: 1,
	}
	for trait, value := range want {
		if traits[trait] != value {
			t.Errorf("客製 runtime 特性 %d=%d，期望 %d", trait, traits[trait], value)
		}
	}
	restored := s.snapshot().restore()
	if restored.CustomRaceRuntimeTraits != s.CustomRaceRuntimeTraits {
		t.Fatal("客製 31 格 runtime 特性必須通過存檔往返")
	}
	seatCopy := s.saveSeat()
	s.CustomRaceRuntimeTraits = [gamedata.RaceTraitCount]int8{}
	s.loadSeat(seatCopy)
	if s.CustomRaceRuntimeTraits != restored.CustomRaceRuntimeTraits {
		t.Fatal("客製 31 格 runtime 特性必須隨熱座席位切換")
	}
}
