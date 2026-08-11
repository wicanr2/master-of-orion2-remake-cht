package gamedata

// original_building_table.go 保存 IDA 從原版建築表直接讀出的 49 筆 raw record。
//
// 證據契約：輸入為 `Orion2.exe` 的 IDA Pro 9.4 資料庫，SHA-256
// `4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e`；位址基準為
// IDA 線性位址。表頭 `0x17EB3D`，步距 `0x13`，欄位為 +4 ID、+6 前置科技、+8
// 建造成本、+12 維護費、+14 分類。這裡是可回查的資料轉錄，不取代 raw 位址。
//
// `Steal_App @ 0x10130A` 讀每列 +8 作 SABOTAGE 權重，因此該消費端必須使用本表，
// 不能再讀 remake 建築描述裡可能不同的估計值。欄位語意已證實；未映射到 UI 名稱的
// raw 槽位仍保留在表內，不因 remake 沒有對應名稱而刪除。

const (
	OriginalBuildingTableAddress = 0x17EB3D
	OriginalBuildingRecordSize   = 0x13
)

// OriginalBuildingRecord 是原版 0x17EB3D + id*0x13 的已解出欄位。
type OriginalBuildingRecord struct {
	ID             int
	Prereq         int
	ProductionCost int
	MaintenanceBC  int
	Category       int
}

// OriginalBuildingTable 是 raw 49 筆表（含 ID 0 的保留列）。
var OriginalBuildingTable = [...]OriginalBuildingRecord{
	{0, 0, 0, 0, 6},
	{1, 5, 60, 1, 4},
	{2, 14, 150, 2, 0},
	{3, 15, 1000, 5, 7},
	{4, 18, 200, 4, 1},
	{5, 19, 150, 3, 2},
	{6, 21, 200, 3, 1},
	{7, 22, 60, 1, 3},
	{8, 27, 1000, 3, 7},
	{9, 32, 200, 0, 1},
	{10, 39, 100, 2, 0},
	{11, 40, 200, 0, 0},
	{12, 49, 250, 3, 0},
	{13, 50, 200, 8, 0},
	{14, 52, 500, 2, 7},
	{15, 61, 60, 1, 3},
	{16, 68, 200, 10, 1},
	{17, 74, 500, 0, 1},
	{18, 75, 250, 3, 5},
	{19, 76, 250, 3, 5},
	{20, 86, 120, 1, 3},
	{21, 87, 60, 2, 0},
	{22, 103, 60, 1, 0},
	{23, 129, 500, 5, 2},
	{24, 130, 200, 3, 1},
	{25, 131, 120, 2, 1},
	{26, 132, 120, 2, 2},
	{27, 133, 200, 2, 2},
	{28, 134, 80, 1, 0},
	{29, 135, 150, 2, 1},
	{30, 136, 150, 2, 2},
	{31, 141, 250, 3, 0},
	{32, 142, 80, 1, 0},
	{33, 152, 200, 3, 1},
	{34, 154, 200, 3, 1},
	{35, 155, 60, 1, 0},
	{36, 156, 150, 2, 3},
	{37, 162, 120, 0, 0},
	{38, 163, 100, 2, 1},
	{39, 164, 80, 1, 0},
	{40, 168, 400, 2, 7},
	{41, 169, 2500, 4, 7},
	{42, 174, 1000, 6, 0},
	{43, 178, 150, 4, 0},
	{44, 183, 250, 0, 0},
	{45, 197, 300, 3, 0},
	{46, 198, 200, 3, 0},
	{47, 67, 150, 2, 0},
	{48, 16, 800, 0, 0},
}

// OriginalBuildingRecordByID 依 raw ID 取記錄；ID 0 也保留為有效保留列。
func OriginalBuildingRecordByID(id int) (OriginalBuildingRecord, bool) {
	if id < 0 || id >= len(OriginalBuildingTable) {
		return OriginalBuildingRecord{}, false
	}
	record := OriginalBuildingTable[id]
	if record.ID != id {
		return OriginalBuildingRecord{}, false
	}
	return record, true
}

// OriginalBuildingProductionCost 回傳 Steal_App 使用的 raw +8 權重。
func OriginalBuildingProductionCost(id int) (int, bool) {
	record, ok := OriginalBuildingRecordByID(id)
	if !ok || record.ProductionCost <= 0 {
		return 0, false
	}
	return record.ProductionCost, true
}
