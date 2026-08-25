package shell

import (
	"reflect"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func twoLaserBlueprint(s *GameSession) ShipBlueprint {
	design, _ := s.ShipDesign(2)
	design.Weapon = 1
	design.Mods = nil
	design.Arc = gamedata.ARC_FWD
	design.WeaponMounts = []ShipWeaponMount{
		{RawType: -1, Name: "雷射", MaxCount: 2, WorkingCount: 2, Arc: gamedata.ARC_FWD, Ammo: 255, Attack: 5},
		{RawType: -1, Name: "雷射", MaxCount: 3, WorkingCount: 3, Arc: gamedata.ARC_FWD, Ammo: 255, Attack: 5},
	}
	return design
}

func TestBlueprintCostAndSpaceSumEveryMountCount(t *testing.T) {
	s := NewDemoSession()
	design := twoLaserBlueprint(s)
	baseCost := s.DesignCostWithLoadout(design.Class, 0, design.Armor, design.Shield, design.Special,
		nil, gamedata.ARC_FWD, 255)
	oneCost := s.DesignCostWithLoadout(design.Class, 1, 0, 0, 0, nil,
		gamedata.ARC_FWD, 255) - ShipCost(design.Class)
	gotCost, ok := s.BlueprintDesignCost(design)
	if !ok || gotCost != baseCost+oneCost*5 {
		t.Fatalf("多槽成本應為共用設備一次 + 各槽單門×數量：got=%d ok=%v want=%d", gotCost, ok, baseCost+oneCost*5)
	}
	baseSpace := s.DesignSpaceUsedWithLoadout(design.Class, 0, design.Armor, design.Shield, design.Special,
		nil, gamedata.ARC_FWD, 255)
	oneSpace := s.DesignSpaceUsedWithLoadout(design.Class, 1, 0, 0, 0, nil, gamedata.ARC_FWD, 255)
	gotSpace, ok := s.BlueprintDesignSpaceUsed(design)
	if !ok || gotSpace != baseSpace+oneSpace*5 {
		t.Fatalf("多槽空間應為特殊一次 + 各槽單門×數量：got=%d ok=%v want=%d", gotSpace, ok, baseSpace+oneSpace*5)
	}
}

func TestBuildShipDesignChargesBlueprintTotalAndReplays(t *testing.T) {
	s := NewDemoSession()
	s.Player.BC = 100000
	design := twoLaserBlueprint(s)
	if !s.SetShipDesign(2, design) {
		t.Fatal("應可保存測試 blueprint")
	}
	total, ok := s.BlueprintDesignCost(design)
	if !ok {
		t.Fatal("已知武器 blueprint 應可計價")
	}
	var commands []PlayerCommand
	s.SetCommandRecorder(func(c PlayerCommand) { commands = append(commands, c) })
	before := s.Player.BC
	if !s.BuildShipDesign(2) {
		t.Fatal("多槽設計應可建造")
	}
	if got := before - s.Player.BC; got != total {
		t.Fatalf("BUILD 應扣完整 blueprint 成本：got=%d want=%d", got, total)
	}
	if len(commands) != 1 || commands[0].Name != CmdBuildShipDesign {
		t.Fatalf("多槽 BUILD 應只送完整 blueprint 命令：%+v", commands)
	}

	replay := NewDemoSession()
	replay.Player.BC = before
	if err := replay.ApplyPlayerCommand(commands[0]); err != nil {
		t.Fatal(err)
	}
	gotShip := replay.Fleet().Ships[len(replay.Fleet().Ships)-1]
	wantShip := s.Fleet().Ships[len(s.Fleet().Ships)-1]
	if replay.Player.BC != s.Player.BC || !reflect.DeepEqual(gotShip.WeaponMounts, wantShip.WeaponMounts) {
		t.Fatalf("多槽命令重播分岔：BC=%d/%d mounts=%+v/%+v", replay.Player.BC, s.Player.BC,
			gotShip.WeaponMounts, wantShip.WeaponMounts)
	}
}

func TestBlueprintUnknownRawEquipmentFailsClosed(t *testing.T) {
	s := NewDemoSession()
	s.Player.BC = 100000
	design := twoLaserBlueprint(s)
	design.WeaponMounts[1].RawType = 999
	design.WeaponMounts[1].Name = "未知原版武器"
	beforeBC, beforeShips := s.Player.BC, len(s.Fleet().Ships)
	if _, ok := s.BlueprintDesignCost(design); ok || s.buildShipBlueprint(design) {
		t.Fatal("未知 raw 武器不得以零成本通過 BUILD")
	}
	if s.Player.BC != beforeBC || len(s.Fleet().Ships) != beforeShips {
		t.Fatal("失敗即關閉不得扣款或新增艦艇")
	}
}

func TestShipDesignMountEditingBoundsAndCompatibility(t *testing.T) {
	s := NewDemoSession()
	idx, ok := s.AddShipDesignMount(2, 0)
	if !ok || idx != 1 {
		t.Fatalf("新增第二槽失敗：idx=%d ok=%v", idx, ok)
	}
	for i := 0; i < 120; i++ {
		s.AdjustShipDesignMountCount(2, idx, 1)
	}
	design, _ := s.ShipDesign(2)
	if design.WeaponMounts[idx].MaxCount != 99 || design.WeaponMounts[idx].WorkingCount != 99 {
		t.Fatalf("槽數量上限應為 99：%+v", design.WeaponMounts[idx])
	}
	for i := 0; i < 120; i++ {
		s.AdjustShipDesignMountCount(2, idx, -1)
	}
	design, _ = s.ShipDesign(2)
	if design.WeaponMounts[idx].MaxCount != 1 {
		t.Fatalf("槽數量下限應為 1：%+v", design.WeaponMounts[idx])
	}
	if !s.RemoveShipDesignMount(2, 0) {
		t.Fatal("兩槽時應可刪除第一槽")
	}
	design, _ = s.ShipDesign(2)
	if len(design.WeaponMounts) != 1 || design.Weapon != 1 || design.WeaponMounts[0].Name != "雷射" {
		t.Fatalf("刪除第一槽後相容欄位應同步新第一槽：%+v", design)
	}
	if s.RemoveShipDesignMount(2, 0) {
		t.Fatal("設計至少必須保留一個武器槽")
	}
}
