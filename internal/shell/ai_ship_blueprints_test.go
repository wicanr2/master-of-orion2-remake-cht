package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func unlockAllForPlayerState(s *GameSession, aiIndex int) {
	ps := &s.AIPlayers[aiIndex].Player
	if ps.CompletedTopics == nil {
		ps.CompletedTopics = map[gamedata.ResearchTopic]bool{}
	}
	for _, area := range gamedata.TechTree() {
		for _, topic := range area {
			ps.CompletedTopics[topic] = true
		}
	}
}

func TestAIShipDesignRolesFollowIDASequence(t *testing.T) {
	s := NewDemoSession()
	if len(s.AIPlayers[0].ShipDesigns) != PlayerShipDesignCount {
		t.Fatalf("AI 應保存六筆原版形狀藍圖，得到 %d", len(s.AIPlayers[0].ShipDesigns))
	}
	for hull, design := range s.AIPlayers[0].ShipDesigns {
		if design.Class != playerShipDesignClasses[hull] || design.RawRole != AutoDesignRole(hull) {
			t.Fatalf("hull %d 應使用 role %d：%+v", hull, hull, design)
		}
	}
}

func TestAIShipTechUpdateRebuildsZeroThroughFourOnly(t *testing.T) {
	s := NewDemoSession()
	sentinel := s.AIPlayers[0].ShipDesigns[5]
	sentinel.RawRole = AutoDesignMixedTheme
	s.AIPlayers[0].ShipDesigns[5] = sentinel
	unlockAllForPlayerState(s, 0)
	s.updateAIShipDesignsAfterTech(0)
	for hull := 0; hull <= 4; hull++ {
		if s.AIPlayers[0].ShipDesigns[hull].RawRole != AutoDesignRole(hull) {
			t.Fatalf("科技更新後 hull %d role 漂移", hull)
		}
	}
	if got := s.AIPlayers[0].ShipDesigns[5].RawRole; got != AutoDesignMixedTheme {
		t.Fatalf("IDA 鏈排除 hull 5，更新不應覆寫：%d", got)
	}
}

func TestAIProductionBuildsPersistentBlueprintShip(t *testing.T) {
	s := NewDemoSession()
	design := s.AIPlayers[0].ShipDesigns[0]
	view := *s
	view.Player = s.AIPlayers[0].Player
	cost, ok := view.BlueprintDesignCost(design)
	if !ok {
		t.Fatal("開局 AI 巡防艦藍圖成本應可解碼")
	}
	s.advanceAIShipProduction(0, cost)
	if len(s.AIPlayers[0].Ships) != 1 {
		t.Fatalf("足額生產點應交付一艘實艦：%d", len(s.AIPlayers[0].Ships))
	}
	ship := s.AIPlayers[0].Ships[0]
	if ship.Class != design.Class || len(ship.WeaponMounts) != len(design.WeaponMounts) {
		t.Fatalf("交付艦未深複製藍圖：ship=%+v design=%+v", ship, design)
	}
	if s.AIPlayers[0].FleetStrength != shipStrength(ship.Class) {
		t.Fatalf("FleetStrength 應由實艦推導：%d", s.AIPlayers[0].FleetStrength)
	}
}

func TestTacticalCombatUsesNamedAIShipsAndWritesLosses(t *testing.T) {
	s := NewDemoSession()
	design := s.AIPlayers[0].ShipDesigns[0]
	ship := shipFromBlueprint("證據艦", design, BuildWeaponOptions(s.RuleProfile), 3, 2)
	s.AIPlayers[0].Ships = []Ship{ship}
	s.syncAIShipStrength(0)
	_, enemies := s.StartCombat(s.PrimaryEnemyName())
	if len(enemies) != 1 || enemies[0].Name != "證據艦" || len(enemies[0].WeaponMounts) != len(ship.WeaponMounts) {
		t.Fatalf("格子戰術未使用 AI 實艦：%+v", enemies)
	}
	s.ApplyCombatOutcomeWithEnemySurvivors(s.PrimaryEnemyName(), 0, 1, map[string]bool{}, map[string]bool{}, true, 1)
	if len(s.AIPlayers[0].Ships) != 0 || s.AIPlayers[0].FleetStrength != 0 {
		t.Fatalf("戰術擊沉未回寫 AI 艦隊：%+v", s.AIPlayers[0])
	}
}

func TestAIShipsGainTurnAcademyAndInstructorXP(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers[0].Ships = []Ship{{Name: "受訓艦", Class: "巡防艦"}}
	s.AIPlayers[0].FleetStar = s.AIPlayers[0].ColonyStars[0]
	s.AIPlayers[0].FleetPosSet = true
	s.AIPlayers[0].ColonyBuildings[0][spaceAcademyName] = true
	s.AIPlayers[0].Leaders = []Leader{{Name: "AI 教官", Skill: "教官", Level: 2, Tier: 1}}
	want := gamedata.CrewXPPerTurnInSpace + gamedata.SpaceAcademyXPPerTurn + leaderInstructorXPBonus(s.AIPlayers[0].Leaders)
	s.advanceCrewExperience()
	if got := s.AIPlayers[0].Ships[0].CrewXP; got != want {
		t.Fatalf("AI 實艦應走每回合／學院／教官 XP 鏈：got %d want %d", got, want)
	}
}

func TestAICommandPointsComeFromBuildingsAndShips(t *testing.T) {
	s := NewDemoSession()
	s.AIPlayers[0].Ships = []Ship{{Class: "巡防艦"}, {Class: "戰艦"}}
	s.syncAICommandPoints(0)
	wantSupply := gamedata.CommandPointsBase + gamedata.CommandPointsFromBuildings(s.AIPlayers[0].ColonyBuildings[0])
	wantUsed := gamedata.ShipCommandCost(gamedata.SHIP_FRIGATE) + gamedata.ShipCommandCost(gamedata.SHIP_BATTLESHIP)
	if got := s.AIPlayers[0].Player.CommandPointsSupply; got != wantSupply {
		t.Fatalf("AI 指揮點供給未讀殖民地建築：got %d want %d", got, wantSupply)
	}
	if got := s.AIPlayers[0].Player.UsedCommandPoints; got != wantUsed {
		t.Fatalf("AI 指揮點需求未讀實艦：got %d want %d", got, wantUsed)
	}
}
