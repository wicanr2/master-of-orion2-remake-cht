package gamedata

import "testing"

func TestOriginalBuildingTableKeepsRaw49Records(t *testing.T) {
	if OriginalBuildingTableAddress != 0x17EB3D || OriginalBuildingRecordSize != 0x13 {
		t.Fatalf("建築表 raw 定位錯誤: address=%#x stride=%#x", OriginalBuildingTableAddress, OriginalBuildingRecordSize)
	}
	if len(OriginalBuildingTable) != 49 {
		t.Fatalf("原版建築表筆數=%d,want 49", len(OriginalBuildingTable))
	}
	for id, record := range OriginalBuildingTable {
		if record.ID != id {
			t.Fatalf("第 %d 筆 raw ID=%d", id, record.ID)
		}
	}
}

func TestOriginalBuildingProductionCostMatchesSabotageRawWeights(t *testing.T) {
	cases := map[int]int{
		9:  200, // 原版保留槽，Steal_App 會跳過；表本身仍保留
		18: 250,
		35: 60,
		41: 2500,
		48: 800,
	}
	for id, want := range cases {
		got, ok := OriginalBuildingProductionCost(id)
		if !ok || got != want {
			t.Errorf("raw building id=%d cost=(%d,%v),want (%d,true)", id, got, ok, want)
		}
	}
	if _, ok := OriginalBuildingProductionCost(0); ok {
		t.Fatal("ID 0 保留列不可作 SABOTAGE 權重")
	}
	if _, ok := OriginalBuildingProductionCost(49); ok {
		t.Fatal("越界 raw building ID 不應有權重")
	}
}
