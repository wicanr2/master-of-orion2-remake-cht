package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestOriginalBestComputerUsesKnownApplications(t *testing.T) {
	ps := engine.PlayerState{GrantedTechs: map[gamedata.Technology]bool{
		gamedata.TECH_ELECTRONIC_COMPUTER: true,
		gamedata.TECH_POSITRONIC_COMPUTER: true,
	}}
	if got := originalBestComputer(ps); got != 3 {
		t.Fatalf("最佳電腦 raw=%d，預期 Positronic 的 3", got)
	}
}

func TestOriginalWeaponMountPreservesKnownRawShape(t *testing.T) {
	component := Component{Name: "雷射", Value: 4, UnlockTech: gamedata.TECH_LASER_CANNON}
	mount := originalWeaponMount(component,
		[]string{string(gamedata.ModHeavyMount), string(gamedata.ModContinuousFire)},
		gamedata.ARC_FWD, 255, 3)
	if mount.RawType != 3 || mount.RawMods != 0x0012 || mount.WorkingCount != 3 {
		t.Fatalf("新造雷射 raw shape 錯誤：%+v", mount)
	}
	unknown := originalWeaponMount(component, []string{string(gamedata.ModArmorPiercing)}, gamedata.ARC_FWD, 255, 1)
	if unknown.RawType != -1 {
		t.Fatalf("未閉合的 AP raw bit 必須阻止國力 producer 消費：%+v", unknown)
	}
}

func TestShipFromBlueprintSetsOriginalPowerInputs(t *testing.T) {
	s := NewDemoSession()
	design := s.AIPlayers[0].ShipDesigns[0]
	ship := shipFromBlueprint("raw input", design, BuildWeaponOptions(s.RuleProfile), 4, 3)
	if !ship.ComputerRawKnown || ship.ComputerRaw != 4 || !ship.DesignSizeRawKnown ||
		!ship.BaseCombatSpeedKnown || !ship.ArmorRawKnown || !ship.ShieldRawKnown {
		t.Fatalf("新造艦未建立完整國力 raw 輸入：%+v", ship)
	}
	if want := gamedata.ShipCombatSpeed(3, gamedata.CombatShipClass(ship.DesignSizeRaw), false, false); int(ship.BaseCombatSpeedRaw) != want {
		t.Fatalf("base combat speed=%d，預期 %d", ship.BaseCombatSpeedRaw, want)
	}
}
