package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestOriginalMineralEventEligibility(t *testing.T) {
	colonies := make([]engine.ColonyState, 3)
	planets := []Planet{{MineralID: gamedata.ULTRA_POOR}, {MineralID: gamedata.RICH}, {MineralID: gamedata.ULTRA_RICH}}
	planetAt := func(i int) *Planet { return &planets[i] }

	idx, ok := originalMineralEventColony(colonies, planetAt, 11, newRandStream(11))
	if !ok || idx != 2 {
		t.Fatalf("枯竭事件只能選 Ultra Rich，得到 idx=%d ok=%v", idx, ok)
	}
	for seed := int64(1); seed <= 20; seed++ {
		idx, ok = originalMineralEventColony(colonies, planetAt, 12, newRandStream(seed))
		if !ok || idx == 2 {
			t.Fatalf("發現事件不可選已達上限的 Ultra Rich，seed=%d idx=%d ok=%v", seed, idx, ok)
		}
	}
}

func TestMineralEventExactDeltasAndGravity(t *testing.T) {
	for _, tc := range []struct {
		name    string
		eventID int
		from    gamedata.PlanetMinerals
		want    gamedata.PlanetMinerals
	}{
		{"枯竭四降三", 11, gamedata.ULTRA_RICH, gamedata.RICH},
		{"發現零升二", 12, gamedata.ULTRA_POOR, gamedata.ABUNDANT},
		{"發現三封頂四", 12, gamedata.RICH, gamedata.ULTRA_RICH},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := engine.ColonyState{MineralRichness: tc.from}
			p := Planet{MineralID: tc.from, GravityID: gamedata.HEAVY_G, Gravity: "高重力"}
			from, to, ok := applyMineralEventToColony(&c, &p, tc.eventID)
			if !ok || from != tc.from || to != tc.want {
				t.Fatalf("效果錯誤：from=%v to=%v ok=%v", from, to, ok)
			}
			if p.MineralID != tc.want || p.Mineral != mineralDisplayName(tc.want) || c.MineralRichness != tc.want {
				t.Fatalf("行星／殖民地礦產未同步：planet=%v display=%q colony=%v", p.MineralID, p.Mineral, c.MineralRichness)
			}
			if c.IndustryPerWorker != gamedata.MineralIndustryPerWorker(tc.want) {
				t.Fatalf("每工人產能未同步：got %d", c.IndustryPerWorker)
			}
			if p.GravityID != gamedata.HEAVY_G || p.Gravity != "高重力" {
				t.Fatalf("原版事件不應重算重力：%v %q", p.GravityID, p.Gravity)
			}
		})
	}
}

func TestMineralEventUpdatesPlayerAndAIWorldState(t *testing.T) {
	s := NewDemoSession()
	s.eventRand = newRandStream(7)
	for i := range s.PlayerColonies {
		if p := s.ColonyPlanet(i); p != nil {
			p.MineralID = gamedata.ULTRA_RICH
			s.PlayerColonies[i].MineralRichness = gamedata.ULTRA_RICH
		}
	}
	idx, from, to, ok := s.applyPlayerMineralEvent(11)
	if !ok || from != gamedata.ULTRA_RICH || to != gamedata.RICH {
		t.Fatalf("玩家枯竭事件錯誤：idx=%d from=%v to=%v ok=%v", idx, from, to, ok)
	}
	if p := s.ColonyPlanet(idx); p == nil || p.MineralID != gamedata.RICH || s.PlayerColonies[idx].MineralRichness != gamedata.RICH {
		t.Fatal("玩家行星與殖民地未同步")
	}

	if len(s.AIPlayers) == 0 || len(s.AIPlayers[0].Colonies) == 0 {
		t.Fatal("demo 應有 AI 殖民地")
	}
	playerMineral := s.ColonyPlanet(0).MineralID
	for i := range s.AIPlayers[0].Colonies {
		if p := s.aiColonyPlanet(0, i); p != nil {
			p.MineralID = gamedata.ULTRA_POOR
			s.AIPlayers[0].Colonies[i].MineralRichness = gamedata.ULTRA_POOR
		}
	}
	idx, from, to, ok = s.applyAIMineralEvent(0, 12)
	if !ok || from != gamedata.ULTRA_POOR || to != gamedata.ABUNDANT {
		t.Fatalf("AI 發現事件錯誤：idx=%d from=%v to=%v ok=%v", idx, from, to, ok)
	}
	if p := s.aiColonyPlanet(0, idx); p == nil || p.MineralID != gamedata.ABUNDANT || s.AIPlayers[0].Colonies[idx].MineralRichness != gamedata.ABUNDANT {
		t.Fatal("AI 行星與殖民地未同步")
	}
	if s.ColonyPlanet(0).MineralID != playerMineral {
		t.Fatal("AI 事件不應改動玩家殖民地")
	}
}
