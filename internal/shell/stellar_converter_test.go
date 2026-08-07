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
