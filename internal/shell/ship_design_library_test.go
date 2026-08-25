package shell

import (
	"reflect"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

func TestShipDesignLibraryHasSixIndependentLegalHulls(t *testing.T) {
	s := NewDemoSession()
	s.EnsureShipDesigns()
	if len(s.ShipDesigns) != PlayerShipDesignCount {
		t.Fatalf("玩家設計庫=%d，want %d", len(s.ShipDesigns), PlayerShipDesignCount)
	}
	for i, design := range s.ShipDesigns {
		if design.Class != playerShipDesignClasses[i] {
			t.Fatalf("設計 %d 艦體=%q，want %q", i, design.Class, playerShipDesignClasses[i])
		}
		if !s.DesignFitsWithLoadout(design.Class, design.Weapon, design.Armor, design.Shield,
			design.Special, design.Mods, design.Arc, design.Ammo) {
			t.Fatalf("設計 %d 不符合艦體空間：%+v", i, design)
		}
	}
	cruiser, _ := s.ShipDesign(2)
	frigate, _ := s.ShipDesign(0)
	cruiser.Weapon = 0
	if !s.SetShipDesign(2, cruiser) {
		t.Fatal("巡洋艦設計應可更新")
	}
	gotFrigate, _ := s.ShipDesign(0)
	if gotFrigate.Weapon != frigate.Weapon {
		t.Fatal("修改巡洋艦不得污染巡防艦設計")
	}
}

func TestShipDesignLoadoutPreservesTailMounts(t *testing.T) {
	s := NewDemoSession()
	s.EnsureShipDesigns()
	design, _ := s.ShipDesign(2)
	design.WeaponMounts = append(design.WeaponMounts, ShipWeaponMount{RawType: 16, Name: "脈衝飛彈", MaxCount: 3})
	design.SpecialIDs = []int{2, 17}
	s.SetShipDesign(2, design)
	if !s.SetShipDesignLoadout(2, AutoDesignLoadout{Weapon: 0, Armor: 0, Shield: 0, Special: 0, RawRole: AutoDesignMixed}) {
		t.Fatal("相容 loadout API 應可寫回")
	}
	got, _ := s.ShipDesign(2)
	if len(got.WeaponMounts) != 2 || got.WeaponMounts[1].RawType != 16 {
		t.Fatalf("相容 loadout API 不得丟棄尾端 raw mount：%+v", got.WeaponMounts)
	}
	if len(got.SpecialIDs) != 2 || got.SpecialIDs[1] != 17 {
		t.Fatalf("相容 loadout API 不得丟棄 special IDs：%v", got.SpecialIDs)
	}
}

func TestShipDesignLibrarySnapshotAndSeatRoundTrip(t *testing.T) {
	s := NewDemoSession()
	s.EnsureShipDesigns()
	design, _ := s.ShipDesign(4)
	design.Ammo = 15
	design.SpecialIDs = []int{7, 31}
	s.SetShipDesign(4, design)

	data, err := s.MarshalSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	got, err := RestoreSnapshot(data)
	if err != nil {
		t.Fatal(err)
	}
	restored, _ := got.ShipDesign(4)
	if restored.Ammo != 15 || len(restored.SpecialIDs) != 2 || restored.SpecialIDs[1] != 31 {
		t.Fatalf("設計庫快照往返失真：%+v", restored)
	}

	seatCopy := s.saveSeat()
	s.ShipDesigns = nil
	s.loadSeat(seatCopy)
	seatDesign, _ := s.ShipDesign(4)
	if seatDesign.Ammo != 15 || len(seatDesign.SpecialIDs) != 2 {
		t.Fatalf("熱座換席後設計庫失真：%+v", seatDesign)
	}
}

func TestBuildShipDesignUsesSelectedHullOnlyWhenCalled(t *testing.T) {
	s := NewDemoSession()
	s.Player.BC = 100000
	s.EnsureShipDesigns()
	beforeShips, beforeBC := len(s.Fleet().Ships), s.Player.BC
	if _, ok := s.ShipDesign(3); !ok {
		t.Fatal("讀取設計不應失敗")
	}
	if len(s.Fleet().Ships) != beforeShips || s.Player.BC != beforeBC {
		t.Fatal("選擇／讀取設計不得造船或扣款")
	}
	if !s.BuildShipDesign(3) {
		t.Fatal("明確建造戰艦設計應成功")
	}
	if len(s.Fleet().Ships) != beforeShips+1 || s.Fleet().Ships[len(s.Fleet().Ships)-1].Class != "戰艦" {
		t.Fatalf("BUILD 應新增一艘戰艦：%+v", s.Fleet().Ships)
	}
}

func TestBuildShipDesignPreservesAllMountsAndSpecialIDs(t *testing.T) {
	s := NewDemoSession()
	s.Player.BC = 100000
	design, _ := s.ShipDesign(2)
	design.WeaponMounts = []ShipWeaponMount{
		{RawType: -1, Name: "雷射", MaxCount: 2, WorkingCount: 2, Arc: gamedata.ARC_FWD, Attack: 5},
		{RawType: 16, Name: "核飛彈", MaxCount: 3, WorkingCount: 2, Arc: gamedata.ARC_360, Ammo: 5, Attack: 8},
	}
	design.SpecialIDs = []int{2, 17}
	if !s.SetShipDesign(2, design) || !s.BuildShipDesign(2) {
		t.Fatal("多槽巡洋艦設計應可建造")
	}
	ship := s.Fleet().Ships[len(s.Fleet().Ships)-1]
	if !reflect.DeepEqual(ship.WeaponMounts, design.WeaponMounts) || !reflect.DeepEqual(ship.SpecialIDs, design.SpecialIDs) {
		t.Fatalf("建造邊界不得丟棄多槽資料：mounts=%+v specials=%v", ship.WeaponMounts, ship.SpecialIDs)
	}
	s.ShipDesigns[2].WeaponMounts[1].Name = "已修改"
	s.ShipDesigns[2].SpecialIDs[1] = 99
	if ship.WeaponMounts[1].Name != "核飛彈" || ship.SpecialIDs[1] != 17 {
		t.Fatal("建成艦的多槽資料必須與 blueprint 深複製隔離")
	}
}

func TestUpdatePlayerShipDesignsAfterTechOnlyUpgradesArmor(t *testing.T) {
	s := NewDemoSession()
	s.EnsureShipDesigns()
	before := make([]ShipBlueprint, PlayerShipDesignCount)
	for i := range before {
		design, _ := s.ShipDesign(i)
		design.RawRole = AutoDesignRole(i % 3)
		design.Weapon = i + 1
		design.Shield = i + 2
		design.Special = i + 3
		design.Mods = []string{"NR", "ECCM"}
		design.Arc = gamedata.ARC_360
		design.Ammo = 7 + i
		design.WeaponMounts = []ShipWeaponMount{{RawType: 16 + i, Name: "保留武器", MaxCount: 3}}
		design.SpecialIDs = []int{2, 17 + i}
		design.Armor = 1
		s.SetShipDesign(i, design)
		before[i], _ = s.ShipDesign(i)
	}

	s.Player.CompletedTopics[gamedata.TOPIC_ADVANCED_METALLURGY] = true
	if s.Player.ChosenTech == nil {
		s.Player.ChosenTech = map[gamedata.ResearchTopic]gamedata.Technology{}
	}
	if s.Player.ExplicitChoice == nil {
		s.Player.ExplicitChoice = map[gamedata.ResearchTopic]bool{}
	}
	s.Player.ChosenTech[gamedata.TOPIC_ADVANCED_METALLURGY] = gamedata.TECH_TRITANIUM_ARMOR
	s.Player.ExplicitChoice[gamedata.TOPIC_ADVANCED_METALLURGY] = true
	s.UpdatePlayerShipDesignsAfterTech()

	for i := range before {
		got, _ := s.ShipDesign(i)
		if got.Armor != 2 {
			t.Fatalf("設計 %d 應自動升級為三鈦裝甲，got %d", i, got.Armor)
		}
		want := before[i]
		want.Armor = 2
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("設計 %d 除裝甲外不得被重建：\ngot  %+v\nwant %+v", i, got, want)
		}
	}
}

func TestChooseResearchTechUpdatesAllPlayerDesignArmor(t *testing.T) {
	s := NewDemoSession()
	s.EnsureShipDesigns()
	s.Player.CompletedTopics[gamedata.TOPIC_ADVANCED_METALLURGY] = true
	s.Player.PendingChoice = gamedata.TOPIC_ADVANCED_METALLURGY
	s.Player.HasPendingChoice = true
	if !s.ChooseResearchTech(gamedata.TECH_TRITANIUM_ARMOR) {
		t.Fatal("三鈦裝甲應是進階冶金的合法研究選擇")
	}
	for i, design := range s.ShipDesigns {
		if design.Armor != 2 {
			t.Fatalf("研究選擇後設計 %d 裝甲=%d，want 2", i, design.Armor)
		}
	}
}
