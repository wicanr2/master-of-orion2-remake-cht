package shell

import (
	"reflect"
	"testing"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/save"
)

func TestImportedSpecialMountsPreserveKnownAndUnknownRawIDs(t *testing.T) {
	design := save.ShipDesign{Name: "SPECIALS"}
	report := &GAMImportReport{}
	mounts := importedSpecialMounts(design, []int{13, 255}, report)
	if len(mounts) != 2 || mounts[0].RawID != 13 || mounts[0].Name == "原版特殊#13" {
		t.Fatalf("已知 raw ID 應轉成受控名稱：%+v", mounts)
	}
	if mounts[1] != (ShipSpecialMount{RawID: 255, Name: "原版特殊#255"}) || len(report.Notes) != 1 {
		t.Fatalf("未知 raw ID 應原樣保存並留下警告：mount=%+v notes=%v", mounts[1], report.Notes)
	}
}

func TestMultipleSpecialsReachFastAndTacticalConsumers(t *testing.T) {
	s := NewDemoSession()
	sh := Ship{Name: "多系統艦", Class: "戰艦", Weapon: "雷射砲", Special: "無",
		Specials: []ShipSpecialMount{{RawID: -1, Name: highEnergyFocusName}, {RawID: -1, Name: antiMissileRocketName},
			{RawID: -1, Name: "硬化護盾"}, {RawID: -1, Name: "結構分析儀"},
			{RawID: -1, Name: "戰機庫"}, {RawID: -1, Name: "重戰機庫"}}}
	s.Fleet().Ships = append(s.Fleet().Ships, sh)
	fast, _ := s.mkPlayerCombatantsIndexed()
	fc := fast[len(fast)-1]
	if !fc.hasHEF || !fc.hasAMR || !fc.hardShield || !fc.beamSystems.StructuralAnalyzer {
		t.Fatalf("快速結算遺漏多特殊裝置：%+v", fc)
	}
	tactical, _ := s.StartCombat("測試敵人")
	tc := tactical[len(tactical)-1]
	if !tc.HEF || !tc.HasAMR || !tc.HardShield || !tc.BeamSystems.StructuralAnalyzer || len(tc.Bays) != 2 {
		t.Fatalf("格子戰術遺漏多特殊裝置：%+v", tc)
	}
}

func TestSpecialSlotsCostOnceRejectDuplicatesAndRoundTrip(t *testing.T) {
	s := NewDemoSession()
	design, _ := s.ShipDesign(2)
	design.Special = 0
	design.Specials = []ShipSpecialMount{specialMountFromOption(1), specialMountFromOption(2)}
	base := design
	base.Specials = []ShipSpecialMount{specialMountFromOption(0)}
	baseCost, baseOK := s.BlueprintDesignCost(base)
	got, ok := s.BlueprintDesignCost(design)
	one1 := s.DesignCostWithLoadout(design.Class, 0, 0, 0, 1, nil, design.Arc, 255) - ShipCost(design.Class)
	one2 := s.DesignCostWithLoadout(design.Class, 0, 0, 0, 2, nil, design.Arc, 255) - ShipCost(design.Class)
	if !baseOK || !ok || got != baseCost+one1+one2 {
		t.Fatalf("多特殊裝置成本應各計一次：got=%d ok=%v want=%d", got, ok, baseCost+one1+one2)
	}
	dup := design
	dup.Specials = append(cloneSpecialMounts(design.Specials), design.Specials[0])
	if _, ok := s.BlueprintDesignCost(dup); ok {
		t.Fatal("重複特殊裝置必須失敗即關閉")
	}
	s.Fleet().Ships[0].Specials = cloneSpecialMounts(design.Specials)
	b, err := s.MarshalSnapshot()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreSnapshot(b)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(restored.Fleet().Ships[0].Specials, design.Specials) {
		t.Fatalf("特殊裝置槽快照往返失真：%+v", restored.Fleet().Ships[0].Specials)
	}
}

func TestSpecialSlotEditingBoundsAndDuplicateSkip(t *testing.T) {
	s := NewDemoSession()
	idx, ok := s.AddShipDesignSpecialMount(2)
	if !ok || idx != 1 {
		t.Fatalf("新增特殊槽失敗：idx=%d ok=%v", idx, ok)
	}
	if !s.SetShipDesignSpecialMount(2, 0, 0) {
		t.Fatal("第一槽應可切回無特殊裝置")
	}
	design, _ := s.ShipDesign(2)
	if !s.SetShipDesignSpecialMount(2, 1, 2) {
		t.Fatal("第二槽應可保存不同特殊裝置")
	}
	if !s.SetShipDesignSpecialMount(2, 0, 1) {
		t.Fatal("第一槽應可保存測試特殊裝置")
	}
	if s.SetShipDesignSpecialMount(2, 1, 1) {
		t.Fatal("不得把第二槽改成與第一槽重複")
	}
	for i := 0; i < 10; i++ {
		s.AddShipDesignSpecialMount(2)
	}
	design, _ = s.ShipDesign(2)
	if len(design.Specials) != 8 {
		t.Fatalf("特殊裝置槽上限應為 8，得到 %d", len(design.Specials))
	}
}
