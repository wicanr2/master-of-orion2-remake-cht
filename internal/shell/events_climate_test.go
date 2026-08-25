package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestClimateEventAcceptsAllPreTerranClimates(t *testing.T) {
	for climate := gamedata.TOXIC; climate < gamedata.TERRAN; climate++ {
		colonies := []engine.ColonyState{{Climate: climate}}
		i, ok := originalClimateEventColony(colonies, newRandStream(int64(climate)+1))
		if !ok || i != 0 {
			t.Errorf("climate %v 應可成為事件目標：i=%d ok=%v", climate, i, ok)
		}
	}
	for _, climate := range []gamedata.PlanetClimate{gamedata.TERRAN, gamedata.GAIA} {
		if _, ok := originalClimateEventColony([]engine.ColonyState{{Climate: climate}}, newRandStream(1)); ok {
			t.Errorf("climate %v 不可成為事件目標", climate)
		}
	}
}

func TestClimateEventOnlyChangesEligiblePlayerColonyAndPlanet(t *testing.T) {
	s := NewDemoSession()
	s.eventRand = newRandStream(41)
	s.PlayerColonies[0].Climate = gamedata.RADIATED
	s.PlayerColonies[0].FoodPerFarmer = gamedata.ClimateFoodPerFarmer(gamedata.RADIATED)
	s.PlayerColonies[0].PopMax = 10
	p := s.ColonyPlanet(0)
	if p == nil {
		t.Fatal("demo 母星必須有行星映射")
	}
	p.ClimateID, p.Climate = gamedata.RADIATED, climateDisplayName(gamedata.RADIATED)

	ev := *gamedata.RandomEventByID(1)
	result, ok := s.applyRandomEventLocalized(ev)
	if !ok || result.Message == "" || result.MessageEN == "" {
		t.Fatalf("氣候事件應可結算：result=%+v ok=%v", result, ok)
	}
	if s.PlayerColonies[0].Climate != gamedata.TERRAN || p.ClimateID != gamedata.TERRAN ||
		p.Climate != climateDisplayName(gamedata.TERRAN) {
		t.Fatalf("colony／planet 未同步到 Terran：colony=%v planet=%+v", s.PlayerColonies[0].Climate, *p)
	}
	if s.PlayerColonies[0].FoodPerFarmer != gamedata.ClimateFoodPerFarmer(gamedata.TERRAN) {
		t.Fatalf("食物基值未重算：got %d", s.PlayerColonies[0].FoodPerFarmer)
	}
}

func TestAIClimateEventChangesAIColonyAndPlanet(t *testing.T) {
	s := NewDemoSession()
	s.eventRand = newRandStream(43)
	a := &s.AIPlayers[0]
	if len(a.Colonies) == 0 {
		t.Fatal("demo AI 必須有殖民地")
	}
	a.Colonies[0].Climate = gamedata.TOXIC
	a.Colonies[0].FoodPerFarmer = gamedata.ClimateFoodPerFarmer(gamedata.TOXIC)
	p := s.aiColonyPlanet(0, 0)
	if p == nil {
		t.Fatal("demo AI 殖民地必須有行星映射")
	}
	p.ClimateID, p.Climate = gamedata.TOXIC, climateDisplayName(gamedata.TOXIC)
	playerClimate := s.PlayerColonies[0].Climate

	ev := *gamedata.RandomEventByID(1)
	result, ok := s.applyRandomEventLocalizedToAI(ev, 0)
	if !ok || result.Message == "" || result.MessageEN == "" {
		t.Fatalf("AI 氣候事件應可結算：result=%+v ok=%v", result, ok)
	}
	if a.Colonies[0].Climate != gamedata.TERRAN || p.ClimateID != gamedata.TERRAN {
		t.Fatalf("AI colony／planet 未同步：colony=%v planet=%v", a.Colonies[0].Climate, p.ClimateID)
	}
	if s.PlayerColonies[0].Climate != playerClimate {
		t.Fatal("AI 事件不得修改玩家殖民地")
	}
}
