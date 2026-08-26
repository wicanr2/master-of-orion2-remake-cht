package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestCustomHomeworldEnvironmentTraitsReachColonyAndPlanet(t *testing.T) {
	s := NewDemoSession()
	s.ApplyCustomRaceBonuses(
		Race{Name: "測試自訂", EnName: "Custom", IndBonus: 0, ResBonus: 0, FoodBonus: 0},
		gamedata.TRAIT_LARGE_HOMEWORLD,
		gamedata.TRAIT_RICH_HOMEWORLD,
		gamedata.TRAIT_ARTIFACTS_HOMEWORLD,
	)
	c := s.PlayerColonies[0]
	if c.PlanetSize != gamedata.LARGE_PLANET || c.PopMax != gamedata.PlanetBasePopMax(gamedata.LARGE_PLANET, gamedata.TERRAN) {
		t.Fatalf("大型母星應改成 Large 的人口上限,得到 size=%v popMax=%d", c.PlanetSize, c.PopMax)
	}
	if c.MineralRichness != gamedata.RICH || c.IndustryPerWorker != gamedata.MineralIndustryPerWorker(gamedata.RICH) {
		t.Fatalf("富礦母星應為每工人 5 產能,得到 mineral=%v industry=%d", c.MineralRichness, c.IndustryPerWorker)
	}
	if c.ResearchPerScientist != gamedata.ResearchPerScientistNorm+2 {
		t.Fatalf("遺物母星應為每科學家基準+2,得到 %d", c.ResearchPerScientist)
	}

	p := &s.Planets[s.PlayerColonyPlanets[0]]
	if p.SizeID != gamedata.LARGE_PLANET || p.MineralID != gamedata.RICH {
		t.Fatalf("母星行星資料未同步:size=%v mineral=%v", p.SizeID, p.MineralID)
	}
}

func TestRaceEnvironmentMappings(t *testing.T) {
	if got := raceFoodClimate(gamedata.TUNDRA, true); got != gamedata.TERRAN {
		t.Errorf("水生 Tundra 應視同 Terran,得到 %v", got)
	}
	if got := raceFoodClimate(gamedata.SWAMP, true); got != gamedata.OCEAN {
		t.Errorf("水生 Swamp 應視同 Ocean,得到 %v", got)
	}
	if got := raceFoodClimate(gamedata.TERRAN, true); got != gamedata.GAIA {
		t.Errorf("水生 Terran 應視同 Gaia,得到 %v", got)
	}

	base := gamedata.PlanetBasePopMax(gamedata.MEDIUM_PLANET, gamedata.BARREN)
	tolerant := gamedata.PlanetBasePopMax(gamedata.MEDIUM_PLANET, gamedata.TERRAN)
	if got := racePopulationMax(gamedata.MEDIUM_PLANET, gamedata.BARREN, false, true, false); got != tolerant {
		t.Errorf("環境耐受族人口上限應把環境視同 Terran:%d vs %d", got, tolerant)
	}
	if got := racePopulationMax(gamedata.MEDIUM_PLANET, gamedata.BARREN, false, false, true); got != base+2*3 {
		t.Errorf("穴居族人口上限應加 2×星球大小:%d vs %d", got, base+2*3)
	}
}

func TestNewColonyUsesRaceEnvironmentRules(t *testing.T) {
	s := NewDemoSession()
	s.ApplyCustomRaceBonuses(
		Race{Name: "測試自訂", EnName: "Custom"},
		gamedata.TRAIT_AQUATIC,
		gamedata.TRAIT_TOLERANT,
		gamedata.TRAIT_SUBTERRANEAN,
	)
	planetIdx := s.PlanetAt(1)
	if planetIdx < 0 {
		t.Fatal("測試星系應有可操作的代表行星")
	}
	s.Planets[planetIdx].Gen = 2
	s.Planets[planetIdx].NoPlanet = false
	s.Planets[planetIdx].TypeID = gamedata.HABITABLE
	s.Planets[planetIdx].ClimateID = gamedata.TUNDRA
	s.Planets[planetIdx].GravityID = gamedata.NORMAL_G
	s.Planets[planetIdx].MineralID = gamedata.ABUNDANT
	s.Planets[planetIdx].SizeID = gamedata.MEDIUM_PLANET

	colony, ok, reason := s.newColonyFromPlanet(planetIdx, gamedata.MoraleGovDictatorship, 0, 0, 0)
	if !ok {
		t.Fatalf("應能建立測試殖民地: %v", reason)
	}
	if colony.FoodPerFarmer != gamedata.ClimateFoodPerFarmer(gamedata.TERRAN) {
		t.Errorf("水生族 Tundra 食物應按 Terran 計算,得到 %d", colony.FoodPerFarmer)
	}
	wantMax := racePopulationMax(gamedata.MEDIUM_PLANET, gamedata.TUNDRA, true, true, true)
	if colony.PopMax != wantMax {
		t.Errorf("水生／環境耐受／穴居人口上限=%d,want %d", colony.PopMax, wantMax)
	}
	if !colony.Aquatic || !colony.TolerantRace || !colony.Subterranean {
		t.Errorf("新殖民地未保存種族環境旗標:%+v", colony)
	}
}
