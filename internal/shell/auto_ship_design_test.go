package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func unlockAllAutoDesignTech(s *GameSession) {
	if s.Player.CompletedTopics == nil {
		s.Player.CompletedTopics = map[gamedata.ResearchTopic]bool{}
	}
	for _, area := range gamedata.TechTree() {
		for _, topic := range area {
			s.Player.CompletedTopics[topic] = true
		}
	}
}

func TestAutoDesignShipUsesUnlockedEquipmentAndFits(t *testing.T) {
	s := NewDemoSession()
	got, ok := s.AutoDesignShip("巡洋艦", AutoDesignMixed)
	if !ok {
		t.Fatal("一般自動設計應能產生合法巡洋艦")
	}
	if !s.ComponentUnlocked(WeaponOptions[got.Weapon]) ||
		!s.ComponentUnlocked(ArmorOptions[got.Armor]) ||
		!s.ComponentUnlocked(ShieldOptions[got.Shield]) ||
		!s.ComponentUnlocked(SpecialOptions[got.Special]) {
		t.Fatalf("自動設計使用了未解鎖元件：%+v", got)
	}
	if !s.DesignFitsWithLoadout("巡洋艦", got.Weapon, got.Armor, got.Shield,
		got.Special, got.Mods, got.Arc, got.Ammo) {
		t.Fatalf("自動設計超出艦體空間：%+v", got)
	}
}

func TestAutoDesignShipMissileRoleDoesNotChooseBeam(t *testing.T) {
	s := NewDemoSession()
	unlockAllAutoDesignTech(s)
	got, ok := s.AutoDesignShip("戰艦", AutoDesignMissile)
	if !ok {
		t.Fatal("飛彈角色應能產生合法戰艦")
	}
	if got.Weapon == 0 || weaponKindByName(WeaponOptions[got.Weapon].Name) != WeaponKindMissile {
		t.Fatalf("raw role 2 應保留飛彈家族，得到 %s", WeaponOptions[got.Weapon].Name)
	}
}

func TestAutoDesignShipFighterRoleOnlyUsesBaySpecial(t *testing.T) {
	s := NewDemoSession()
	unlockAllAutoDesignTech(s)
	got, ok := s.AutoDesignShip("戰艦", AutoDesignFighterA)
	if !ok {
		t.Fatal("戰機角色應能產生合法戰艦")
	}
	if got.Special == 0 || !autoSpecialAllowed(AutoDesignFighterA, SpecialOptions[got.Special].Name) {
		t.Fatalf("raw role 1 不應選非機庫特殊裝置：%s", SpecialOptions[got.Special].Name)
	}
}
