package shell

import (
	"math/rand"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestAdvancedCivilizationNormalNewGameCreatesExtraColonies(t *testing.T) {
	s := NewDemoSession()
	s.TechLevel, s.TechLevelSet = 2, true
	s.SetupNewGame(40, 4242, 3)
	s.ApplyRace(raceIndexByEnglishNameForTest(t, "Humans"))
	if len(s.PlayerColonies) <= 1 {
		t.Fatalf("先進文明正常新遊戲應建立額外玩家殖民地，得到 %d", len(s.PlayerColonies))
	}
	for i := range s.AIPlayers {
		if len(s.AIPlayers[i].Colonies) <= 1 {
			t.Fatalf("先進文明 AI %d 應建立額外殖民地，得到 %d", i, len(s.AIPlayers[i].Colonies))
		}
		if len(s.AIPlayers[i].Colonies) != len(s.AIPlayers[i].ColonyStars) ||
			len(s.AIPlayers[i].Colonies) != len(s.AIPlayers[i].ColonyPlanets) {
			t.Fatalf("AI %d 殖民地平行陣列失同步", i)
		}
	}
	restored := s.snapshot().restore()
	if len(restored.PlayerColonies) != len(s.PlayerColonies) || len(restored.AIPlayers[0].Colonies) != len(s.AIPlayers[0].Colonies) {
		t.Fatal("先進文明額外殖民地必須通過存檔往返")
	}
}

func TestAverageCivilizationStillStartsWithHomeworldOnly(t *testing.T) {
	s := NewDemoSession()
	s.TechLevel, s.TechLevelSet = TechLevelDefault, true
	s.SetupNewGame(40, 4242, 3)
	s.ApplyRace(raceIndexByEnglishNameForTest(t, "Humans"))
	if len(s.PlayerColonies) != 1 {
		t.Fatalf("一般文明應只有母星，得到 %d 殖民地", len(s.PlayerColonies))
	}
}

func TestAdvancedCivilizationBalanceNeverDowngradesPlanets(t *testing.T) {
	s := NewDemoSession()
	s.TechLevel, s.TechLevelSet = 2, true
	s.SetupNewGame(40, 5151, 3)
	before := append([]Planet(nil), s.Planets...)
	s.ApplyRace(raceIndexByEnglishNameForTest(t, "Humans"))
	for i := range s.Planets {
		if s.Planets[i].SizeID < before[i].SizeID || s.Planets[i].ClimateID < before[i].ClimateID ||
			s.Planets[i].MineralID < before[i].MineralID {
			t.Fatalf("平衡不得降級行星 %d：before=%+v after=%+v", i, before[i], s.Planets[i])
		}
		if s.Planets[i].SizeID > 4 || s.Planets[i].ClimateID > 9 || s.Planets[i].MineralID > 4 {
			t.Fatalf("平衡超出 raw 欄位上限：planet=%d %+v", i, s.Planets[i])
		}
	}
}

func TestAdvancedCivilizationSpecialRedistributionUsesAllowedSet(t *testing.T) {
	s := NewDemoSession()
	if len(s.Stars) < 3 {
		t.Fatal("測試星圖至少需要三顆星")
	}
	p0, p1 := s.PlanetAt(1), s.PlanetAt(2)
	if p0 < 0 || p1 < 0 {
		t.Fatal("測試星系需要代表行星")
	}
	s.Planets[p0].SpecialID = gamedata.GemDeposits
	s.Planets[p0].ClimateID = gamedata.GAIA
	s.Planets[p0].SizeID = gamedata.HUGE_PLANET
	s.Planets[p0].MineralID = gamedata.ULTRA_RICH
	for _, planet := range s.PlanetsAt(2) {
		s.Planets[planet].SpecialID = gamedata.NoSpecial
	}
	s.Planets[p1].SpecialID = gamedata.NoSpecial
	s.Planets[p1].ClimateID = gamedata.TOXIC
	s.Planets[p1].SizeID = gamedata.TINY_PLANET
	s.Planets[p1].MineralID = gamedata.ULTRA_POOR
	s.redistributeAdvancedCivilizationSpecials([][]int{{p0}, {p1}}, []int{0, 2}, rand.New(rand.NewSource(1)))
	switch got := s.Planets[p1].SpecialID; got {
	case gamedata.GoldDeposits, gamedata.GemDeposits, gamedata.AncientArtifacts:
	default:
		t.Fatalf("再分配 special=%d，必須是 raw 4／5／10", got)
	}
}

func TestAdvancedCivilizationProximityBands(t *testing.T) {
	for _, tc := range []struct{ distance, want int }{{1, 1200}, {2, 1200}, {3, 1100}, {4, 1100}, {5, 1050}, {6, 1000}} {
		if got := 1000 * gamedata.AdvancedCivilizationProximityPercent(tc.distance) / 100; got != tc.want {
			t.Errorf("距離 %d proximity worth=%d，預期 %d", tc.distance, got, tc.want)
		}
	}
}
