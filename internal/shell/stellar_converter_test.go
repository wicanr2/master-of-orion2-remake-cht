package shell

import (
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
)

// stellar_converter_test.go:恆星轉換器(行星版)接進殖民地防禦解算的護欄。
//
// 它是**這批防禦建築裡唯一有固定數字的一棟**(手冊 p.106「400 傷 ×2」),所以只有它接得進來。
// 飛彈基地/地面砲台手冊只給「佔 N 空間、裝當時最佳武器」的規則,傷害隨科技現算,
// 沒有艦艇元件的空間模型算不出來——那兩棟仍是「記錄已建但不影響數值」。

// TestStellarConverterAddsColonyDefense:蓋了之後防禦力要真的變高。
// 沒接線的話它就只是一個貴又要維護費的擺設,而測試綠、畫面也正常。
func TestStellarConverterAddsColonyDefense(t *testing.T) {
	s := NewDemoSession()
	if len(s.PlayerColonies) == 0 {
		t.Skip("沒有殖民地")
	}
	for len(s.ColonyBuildings) < len(s.PlayerColonies) {
		s.ColonyBuildings = append(s.ColonyBuildings, map[string]bool{})
	}
	if s.ColonyBuildings[0] == nil {
		s.ColonyBuildings[0] = map[string]bool{}
	}
	before := s.colonyDefense(0)
	s.ColonyBuildings[0][gamedata.StellarConverterName] = true
	after := s.colonyDefense(0)

	if got := after - before; got != gamedata.StellarConverterDefense {
		t.Errorf("恆星轉換器讓防禦 +%d,want +%d(手冊 p.106:400 傷 ×2)",
			got, gamedata.StellarConverterDefense)
	}
}

// TestStellarConverterIsInBuildingTable:成本/維護/分類要對上兩個來源。
//
//	手冊 p.106      → 維護 6 BC
//	原版建築表第 42 列 → 成本 1000 PP、維護 6 BC
//
// 維護費兩邊逐項相符是當初敢建模的理由,所以這裡把它釘住。
func TestStellarConverterIsInBuildingTable(t *testing.T) {
	b, ok := gamedata.BuildingByNameZH(gamedata.StellarConverterName)
	if !ok {
		t.Fatalf("建築表裡找不到 %q", gamedata.StellarConverterName)
	}
	if b.NameEN != "Stellar Converter" {
		t.Errorf("英文名 = %q,want %q", b.NameEN, "Stellar Converter")
	}
	if b.ProductionCost != 1000 {
		t.Errorf("成本 = %d PP,want 1000(原版建築表第 42 列)", b.ProductionCost)
	}
	if b.MaintenanceBC != 6 {
		t.Errorf("維護 = %d BC,want 6(手冊 p.106 與原版建築表兩邊相符)", b.MaintenanceBC)
	}
	if b.PrereqTopic != gamedata.TOPIC_TEMPORAL_PHYSICS {
		t.Errorf("前置 = %v,want TOPIC_TEMPORAL_PHYSICS(手冊 Temporal Physics 15000)", b.PrereqTopic)
	}
}

// TestColonyDefenceUsesSpaceBudgetModel:殖民地防禦要用**與軌道轟炸反擊同一套**推導,
// 不是自編係數。
//
// ⚠ 這條擋的是一個實際存在過的自相矛盾:`colonyDefense` 先前用
// `CommandPointsFromBuildings × 10`,星基因此值 10——比一艘巡洋艦(shipStrength 8)還強,
// 而 `gamedata/satellite.go` 的校準明講星基 ≈ 驅逐艦 tier(4)。同一個東西在兩個地方
// 值不同的分數,而且兩邊都測得綠。
func TestColonyDefenceUsesSpaceBudgetModel(t *testing.T) {
	s := NewDemoSession()
	if len(s.PlayerColonies) == 0 {
		t.Skip("沒有殖民地")
	}
	for len(s.ColonyBuildings) < len(s.PlayerColonies) {
		s.ColonyBuildings = append(s.ColonyBuildings, map[string]bool{})
	}
	if s.ColonyBuildings[0] == nil {
		s.ColonyBuildings[0] = map[string]bool{}
	}
	// 艦隊不在場,才量得到純建築的貢獻。
	s.FleetAtStar, s.FleetETA = -1, 0

	base := s.colonyDefense(0)
	for _, name := range []string{"星基", "戰鬥站", "星辰要塞", "飛彈基地", "地面砲台"} {
		delete(s.ColonyBuildings[0], name)
	}
	bare := s.colonyDefense(0)

	// 三級軌道基地:取代不疊加,而且火力遞增。
	prev := 0
	for _, name := range []string{"星基", "戰鬥站", "星辰要塞"} {
		for _, n2 := range []string{"星基", "戰鬥站", "星辰要塞"} {
			delete(s.ColonyBuildings[0], n2)
		}
		s.ColonyBuildings[0][name] = true
		got := s.colonyDefense(0) - bare
		if got <= prev {
			t.Errorf("%s 的防禦貢獻 %d 沒有比前一級(%d)高——三級應該遞增", name, got, prev)
		}
		prev = got
	}

	// **飛彈基地與地面砲台先前完全不算**,這是這次接線的重點。
	for _, n2 := range []string{"星基", "戰鬥站", "星辰要塞"} {
		delete(s.ColonyBuildings[0], n2)
	}
	for _, name := range []string{"飛彈基地", "地面砲台"} {
		delete(s.ColonyBuildings[0], name)
		before := s.colonyDefense(0)
		s.ColonyBuildings[0][name] = true
		if after := s.colonyDefense(0); after <= before {
			t.Errorf("%s 對防禦沒有貢獻(%d → %d)——手冊 p.78/p.81 給了確認的 space 預算,不該是擺設",
				name, before, after)
		}
	}
	_ = base
}
