package shell

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func officerAssignmentSession() *GameSession {
	s := NewDemoSession()
	s.Fleets = []Fleet{{
		Ships: []Ship{
			{Name: "甲艦", Class: "巡洋艦", Weapon: "雷射", WeaponAttack: 10},
			{Name: "乙艦", Class: "巡洋艦", Weapon: "雷射", WeaponAttack: 10},
		},
		AtStar: 0, DestStar: -1,
	}}
	s.SelectedFleet = 0
	s.Leaders = []Leader{{
		ID:    41,
		Name:  "武器官",
		Skill: "武器官",
		Level: 5,
		Ship:  true,
		Tier:  2,
		Skills: []LeaderSkill{
			{ID: int(gamedata.SKILL_WEAPONRY), Tier: 2},
			{ID: int(gamedata.SKILL_HELMSMAN), Tier: 2},
		},
	}}
	return s
}

func TestOfficerAssignmentIsPerShipAndReassigns(t *testing.T) {
	s := officerAssignmentSession()
	if !s.AssignOfficerToShip(0, 0, 0) {
		t.Fatal("指派艦艇軍官應成功")
	}
	if got := s.Fleets[0].Ships[0].OfficerName; got != "武器官" {
		t.Fatalf("船 0 軍官 = %q, want 武器官", got)
	}
	if got := s.Fleets[0].Ships[0].OfficerID; got != 41 {
		t.Fatalf("船 0 軍官來源 ID = %d, want 41", got)
	}
	if got := s.Fleets[0].Ships[1].OfficerName; got != "" {
		t.Fatalf("船 1 不應被同時指派, got %q", got)
	}

	player, _ := s.StartCombat("測試敵人")
	if player[0].Attack <= player[1].Attack {
		t.Fatalf("軍官只在被指派的船上加攻擊: ship0=%d ship1=%d", player[0].Attack, player[1].Attack)
	}
	if player[0].Defense <= player[1].Defense {
		t.Fatalf("舵手只在被指派的船上加防禦: ship0=%d ship1=%d", player[0].Defense, player[1].Defense)
	}

	if !s.AssignOfficerToShip(0, 1, 0) {
		t.Fatal("改派艦艇軍官應成功")
	}
	if s.Fleets[0].Ships[0].OfficerName != "" || s.Fleets[0].Ships[1].OfficerName != "武器官" {
		t.Fatalf("改派後的逐艦欄位錯誤: %+v", s.Fleets[0].Ships)
	}
	player, _ = s.StartCombat("測試敵人")
	if player[0].Attack >= player[1].Attack {
		t.Fatalf("改派後舊船不應保留武器官加成: ship0=%d ship1=%d", player[0].Attack, player[1].Attack)
	}

	if !s.UnassignOfficerFromShip(0, 1) || s.Fleets[0].Ships[1].OfficerName != "" {
		t.Fatal("再次點擊解除指派應清空船欄位")
	}
}

func TestOfficerAssignmentRoundTripsThroughJSON(t *testing.T) {
	s := officerAssignmentSession()
	if !s.AssignOfficerToShip(0, 1, 0) {
		t.Fatal("指派艦艇軍官應成功")
	}
	path := filepath.Join(t.TempDir(), "officer.json")
	if err := s.Save(path); err != nil {
		t.Fatalf("保存軍官指派: %v", err)
	}
	loaded, err := LoadSession(path)
	if err != nil {
		t.Fatalf("讀回軍官指派: %v", err)
	}
	if got := loaded.Fleets[0].Ships[1].OfficerName; got != "武器官" {
		t.Fatalf("讀回的軍官 = %q, want 武器官", got)
	}
	if got := loaded.Fleets[0].Ships[1].OfficerID; got != 41 {
		t.Fatalf("讀回的軍官來源 ID = %d, want 41", got)
	}
	if got, ok := loaded.OfficerForShip(loaded.Fleets[0].Ships[1]); !ok || got.Name != "武器官" {
		t.Fatalf("讀回後應能解析船上軍官: %+v, %v", got, ok)
	}
}

func TestOfficerAssignmentFallsBackToNameForOldJSON(t *testing.T) {
	s := officerAssignmentSession()
	s.Fleets[0].Ships[0].OfficerName = "武器官"
	// 舊 JSON 沒有 officer_id；名稱仍可解析，且不會把零值 ID 當成唯一依據。
	if got, ok := s.OfficerForShip(s.Fleets[0].Ships[0]); !ok || got.Name != "武器官" {
		t.Fatalf("舊名稱格式應能回退解析軍官: %+v, %v", got, ok)
	}
}

func TestColonyLeaderCannotBeAssignedToShip(t *testing.T) {
	s := officerAssignmentSession()
	s.Leaders[0].Ship = false
	if s.AssignOfficerToShip(0, 0, 0) {
		t.Fatal("殖民地領袖不應能指派到艦艇")
	}
}

// 戰機 Beam Defense 的種族／Fighter Pilot 加成應在參戰資料建立時保留，
// 而且 Fighter Pilot 取目前參戰艦隊中最高的逐艦軍官值。
func TestFighterPilotBonusReachesCombatFighterData(t *testing.T) {
	s := officerAssignmentSession()
	s.ApplyRace(6) // Alkari: Ship Defense +50。
	s.Leaders[0].Skills = append(s.Leaders[0].Skills,
		LeaderSkill{ID: int(gamedata.SKILL_FIGHTER_PILOT), Tier: 2})
	if !s.AssignOfficerToShip(0, 0, 0) {
		t.Fatal("帶 Fighter Pilot 的艦艇軍官應能指派")
	}
	wantPilot := s.shipOfficerSkillBonus(s.Fleets[0].Ships[0], gamedata.SKILL_FIGHTER_PILOT)
	player, _ := s.StartCombat("測試敵人")
	if len(player) < 2 {
		t.Fatalf("測試艦隊應有兩艘參戰艦，實得 %d", len(player))
	}
	for i, ship := range player {
		if ship.FighterRacialDefenseBonus != 50 {
			t.Errorf("第 %d 艘艦的戰機種族防禦應為 +50，實得 %d", i, ship.FighterRacialDefenseBonus)
		}
		if ship.FighterPilotBonus != wantPilot {
			t.Errorf("第 %d 艘艦的 Fighter Pilot 應為 %d，實得 %d", i, wantPilot, ship.FighterPilotBonus)
		}
	}
}
