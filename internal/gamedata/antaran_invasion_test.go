package gamedata

import "testing"

func TestOriginalAntaranResourcePulseTechAndDifficulty(t *testing.T) {
	for _, tc := range []struct {
		tech, elapsed, difficulty, want int
	}{
		{0, 225, 2, 1}, {1, 125, 2, 1}, {2, 25, 2, 1},
		{1, 150, 3, 3}, // period=2，ceil(2×150/100)=3
		{1, 150, 4, 4},
	} {
		got, ok := OriginalAntaranResourcePulse(tc.elapsed, tc.tech, tc.difficulty)
		if !ok || got != tc.want {
			t.Errorf("tech=%d elapsed=%d difficulty=%d: got %d,%v want %d,true",
				tc.tech, tc.elapsed, tc.difficulty, got, ok, tc.want)
		}
	}
	if _, ok := OriginalAntaranResourcePulse(124, 1, 2); ok {
		t.Fatal("一般科技第 124 回合尚未到第一個 pulse")
	}
}

func TestOriginalAntaranBuildShipsCostsCapsAndDiscount(t *testing.T) {
	resource := 200
	ships := [5]int{}
	maxima := AntaranDefensiveMax
	costs := AntaranShipCosts
	discounted := false
	built := OriginalAntaranBuildShips(&resource, &ships, &maxima, &costs, false, 250, 1, 4, &discounted)
	if built == 0 || ships[2] != 3 || ships[3] != 2 || ships[4] == 0 {
		t.Fatalf("防禦艦建造未依 0/0/3/2/7 上限順序：ships=%v built=%d", ships, built)
	}
	if !discounted || costs != [5]int{1, 4, 10, 27, 67} {
		t.Fatalf("高難度首艘第五級艦應使成本乘 90%%：discount=%v costs=%v", discounted, costs)
	}
}

func TestOriginalAntaranReadinessAndRollBoundaries(t *testing.T) {
	off := [5]int{4, 0, 0, 0, 0}
	if !OriginalAntaranInvasionReady(off, [5]int{}, [5]int{}, AntaranShipCosts, 0, 0) {
		t.Fatal("有未部署攻擊艦且四倍戰力足夠時應 ready")
	}
	if OriginalAntaranInvasionReady(off, [5]int{}, off, AntaranShipCosts, 0, 0) {
		t.Fatal("全部已部署時不得再出兵")
	}
	if !OriginalAntaranInvasionRollSucceeds(1, 2) || OriginalAntaranInvasionRollSucceeds(1, 3) {
		t.Fatal("readiness=1 應只有 Random(200) 的 1、2 成功")
	}
}

func TestOriginalAntaranTargetWeightsMinimumAndLucky(t *testing.T) {
	plain := OriginalAntaranTargetWeights([]int{10, 30, 50}, []bool{true, true, true}, nil, 2)
	if plain[0] != 0 || plain[1] != 400 || plain[2] != 1600 {
		t.Fatalf("人口差平方權重不符：%v", plain)
	}
	lucky := OriginalAntaranTargetWeights([]int{10, 30, 50}, []bool{true, true, true}, []bool{false, false, true}, 2)
	if lucky[2] >= plain[2] || lucky[2] == 0 {
		t.Fatalf("Lucky 應降低而非清零權重：plain=%v lucky=%v", plain, lucky)
	}
}
