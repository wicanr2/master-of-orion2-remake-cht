package gamedata

import "testing"

func TestOriginalEventWeightedChoiceHalvesBeforeSampling(t *testing.T) {
	choice, normalized, ok := OriginalEventWeightedChoice([]int{400, 400, 400}, 0)
	if !ok {
		t.Fatal("正權重候選應可抽樣")
	}
	if choice != 0 || len(normalized) != 3 || normalized[0] != 100 || normalized[1] != 100 || normalized[2] != 100 {
		t.Fatalf("sub_586D4 的 0x200 壓縮不符,choice=%d weights=%v", choice, normalized)
	}
	choice, _, ok = OriginalEventWeightedChoice([]int{400, 400, 400}, 99)
	if !ok || choice != 0 {
		t.Fatalf("第一段邊界應仍落在候選 0,choice=%d ok=%v", choice, ok)
	}
	choice, _, ok = OriginalEventWeightedChoice([]int{400, 400, 400}, 100)
	if !ok || choice != 1 {
		t.Fatalf("跨過第一段後應落在候選 1,choice=%d ok=%v", choice, ok)
	}
}

func TestOriginalEventVictimWeightsUseExtremeDifferenceSquare(t *testing.T) {
	indices, weights := OriginalEventVictimWeights([]int{10, 20, 40}, []bool{true, true, true}, true)
	if len(indices) != 2 || indices[0] != 0 || indices[1] != 1 || weights[0] != 900 || weights[1] != 400 {
		t.Fatalf("好事件應排除最高分並用差平方,indices=%v weights=%v", indices, weights)
	}
	indices, weights = OriginalEventVictimWeights([]int{10, 20, 40}, []bool{true, true, true}, false)
	if len(indices) != 2 || indices[0] != 1 || indices[1] != 2 || weights[0] != 100 || weights[1] != 900 {
		t.Fatalf("壞事件應排除最低分並用差平方,indices=%v weights=%v", indices, weights)
	}
}
