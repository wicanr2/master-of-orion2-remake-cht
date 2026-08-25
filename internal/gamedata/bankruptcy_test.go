package gamedata

import "testing"

func TestBankruptcyBuildingFilters(t *testing.T) {
	// ID 7 是 raw category 3：第一輪保留，第二輪成本正值即可出售。
	if BankruptcyBuildingEligible(0, 7, 60) || !BankruptcyBuildingEligible(1, 7, 60) {
		t.Fatal("ID 7 的三輪篩選不符 sub_EDAE2/sub_EDB1D")
	}
	// ID 9 在第一與第三輪排除；第二輪只檢查成本，原始指令沒有 ID 特例。
	if BankruptcyBuildingEligible(0, 9, 200) || !BankruptcyBuildingEligible(1, 9, 200) || BankruptcyBuildingEligible(2, 9, 200) {
		t.Fatal("ID 9 的三輪邊界不符原始篩選器")
	}
	if BankruptcyBuildingEligible(1, 10, 0) || !BankruptcyBuildingEligible(2, 10, 0) {
		t.Fatal("零成本建築只能進第三輪")
	}
}

func TestBankruptcyBuildingScore(t *testing.T) {
	if got := BankruptcyBuildingScore(7, 60, 4, nil); got != 480 {
		t.Fatalf("ID 7 score=%d, want 480", got)
	}
	if got := BankruptcyBuildingScore(12, 250, 4, nil); got != 1750 {
		t.Fatalf("ID 12 score=%d, want 1750", got)
	}
	if got := BankruptcyBuildingScore(22, 60, 2, func(id int) bool { return id == 2 }); got != 60 {
		t.Fatalf("已有 ID 2 時 ID 22 不應乘 8，got %d", got)
	}
}
