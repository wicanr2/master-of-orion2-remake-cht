package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestMissileRackChangesDesignCostAndSpace(t *testing.T) {
	s := NewDemoSession()
	weapon := componentIndexByName(WeaponOptions, "核飛彈")
	space2 := s.DesignSpaceUsedWithLoadout("巡洋艦", weapon, 0, 0, 0, nil, gamedata.ARC_360, 2)
	space20 := s.DesignSpaceUsedWithLoadout("巡洋艦", weapon, 0, 0, 0, nil, gamedata.ARC_360, 20)
	cost2 := s.DesignCostWithLoadout("巡洋艦", weapon, 0, 0, 0, nil, gamedata.ARC_360, 2)
	cost20 := s.DesignCostWithLoadout("巡洋艦", weapon, 0, 0, 0, nil, gamedata.ARC_360, 20)
	if space2 >= space20 || cost2 >= cost20 {
		t.Fatalf("大彈架應增加成本與佔格: space %d/%d cost %d/%d", space2, space20, cost2, cost20)
	}
}

func TestBuildShipAmmoPersistsAndReplays(t *testing.T) {
	base := NewDemoSession()
	base.Player.BC = 9999
	weapon := componentIndexByName(WeaponOptions, "核飛彈")
	var commands []PlayerCommand
	base.SetCommandRecorder(func(c PlayerCommand) { commands = append(commands, c) })
	if !base.BuildShipWithLoadout("巡洋艦", weapon, 0, 0, 0, nil, gamedata.ARC_360, 15) {
		t.Fatal("15 發飛彈設計應能建造")
	}
	ship := base.Fleet().Ships[len(base.Fleet().Ships)-1]
	if ship.WeaponAmmo != 15 {
		t.Fatalf("Ship Ammo=%d want 15", ship.WeaponAmmo)
	}
	if len(commands) != 1 || len(commands[0].Args) < 6 || commands[0].Args[4] != 15 {
		t.Fatalf("CmdBuildShip 應保存 Ammo=15: %+v", commands)
	}

	replayed := NewDemoSession()
	replayed.Player.BC = 9999
	if err := replayed.ApplyPlayerCommands(commands); err != nil {
		t.Fatalf("重播造艦失敗: %v", err)
	}
	replayedShip := replayed.Fleet().Ships[len(replayed.Fleet().Ships)-1]
	if replayedShip.WeaponAmmo != 15 {
		t.Fatalf("重播 Ship Ammo=%d want 15", replayedShip.WeaponAmmo)
	}
}

func TestLegacyBuildShipCommandDefaultsMissileRack(t *testing.T) {
	s := NewDemoSession()
	s.Player.BC = 9999
	weapon := componentIndexByName(WeaponOptions, "核飛彈")
	legacy := PlayerCommand{Name: CmdBuildShip, Args: []int{weapon, 0, 0, 0}, Text: "巡洋艦"}
	if err := s.ApplyPlayerCommand(legacy); err != nil {
		t.Fatalf("舊造艦命令重播失敗: %v", err)
	}
	ship := s.Fleet().Ships[len(s.Fleet().Ships)-1]
	if ship.WeaponAmmo != 5 {
		t.Fatalf("舊造艦命令應回退標準 5 發彈架，得到 %d", ship.WeaponAmmo)
	}
}

func TestStartCombatUsesShipAmmoAndLegacyDefault(t *testing.T) {
	s := NewDemoSession()
	s.Fleet().Ships = []Ship{
		{Name: "二發艦", Class: "巡洋艦", Weapon: "核飛彈", WeaponAmmo: 2},
		{Name: "舊存檔艦", Class: "巡洋艦", Weapon: "核飛彈"},
	}
	ships, _ := s.StartCombat("測試敵")
	if ships[0].WeaponAmmo != 2 || ships[1].WeaponAmmo != 5 {
		t.Fatalf("Combat Ammo=(%d,%d) want (2,5)", ships[0].WeaponAmmo, ships[1].WeaponAmmo)
	}
}
