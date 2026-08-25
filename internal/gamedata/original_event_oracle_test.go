package gamedata

import "testing"

func TestOriginalEventVictimWeightsUsePopulationExtremes(t *testing.T) {
	populations := []int{10, 20, 40, 80}
	eligible := []bool{true, true, true, true}

	indices, weights := OriginalEventVictimWeights(populations, eligible, false)
	assertIntSlice(t, "壞事件索引", indices, []int{1, 2, 3})
	assertIntSlice(t, "壞事件權重", weights, []int{100, 900, 4900})

	indices, weights = OriginalEventVictimWeights(populations, eligible, true)
	assertIntSlice(t, "好事件索引", indices, []int{0, 1, 2})
	assertIntSlice(t, "好事件權重", weights, []int{4900, 3600, 1600})
}

func TestOriginalSupernovaCountdown(t *testing.T) {
	want := [5][2]int{{11, 15}, {10, 14}, {9, 13}, {8, 12}, {7, 11}}
	for difficulty := 0; difficulty < 5; difficulty++ {
		if got := OriginalSupernovaCountdown(difficulty, 1); got != want[difficulty][0] {
			t.Errorf("difficulty %d roll1: got %d want %d", difficulty, got, want[difficulty][0])
		}
		if got := OriginalSupernovaCountdown(difficulty, 5); got != want[difficulty][1] {
			t.Errorf("difficulty %d roll5: got %d want %d", difficulty, got, want[difficulty][1])
		}
	}
}

func TestOriginalSupernovaResearchNeed(t *testing.T) {
	if got := OriginalSupernovaResearchNeed(17, 9); got != 153 {
		t.Fatalf("需求應為 system RP×countdown，got %d", got)
	}
}

func TestOriginalStasisEndsBoundaries(t *testing.T) {
	cases := []struct {
		age, roll int
		want      bool
	}{{4, 1, false}, {5, 1, true}, {5, 2, false}, {20, 2, false}, {21, 2, true}}
	for _, tc := range cases {
		if got := OriginalStasisEnds(tc.age, tc.roll); got != tc.want {
			t.Errorf("age=%d roll=%d: got %v want %v", tc.age, tc.roll, got, tc.want)
		}
	}
}

func TestOriginalEventVictimWeightsRespectEligibilityAndDegeneratePools(t *testing.T) {
	indices, weights := OriginalEventVictimWeights(
		[]int{10, 20, 40, 80}, []bool{true, false, true, true}, false)
	assertIntSlice(t, "排除不適用帝國後索引", indices, []int{2, 3})
	assertIntSlice(t, "排除不適用帝國後權重", weights, []int{900, 4900})

	indices, weights = OriginalEventVictimWeights([]int{10, 20}, []bool{false, true}, true)
	if len(indices) != 0 || len(weights) != 0 {
		t.Fatalf("只剩一個候選時原版移除極值後應為空池：indices=%v weights=%v", indices, weights)
	}
}

func assertIntSlice(t *testing.T, label string, got, want []int) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s 長度=%d，want %d；got=%v", label, len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s[%d]=%d，want %d；got=%v", label, i, got[i], want[i], got)
		}
	}
}

func TestOriginalLuckyEventDivisor(t *testing.T) {
	cases := []struct {
		randomEvents, antaran bool
		want                  int
	}{{false, false, 10}, {false, true, 12}, {true, false, 6}, {true, true, 8}}
	for _, tc := range cases {
		if got := OriginalLuckyEventDivisor(tc.randomEvents, tc.antaran); got != tc.want {
			t.Errorf("settings (%v,%v): got %d want %d", tc.randomEvents, tc.antaran, got, tc.want)
		}
	}
}

func TestOriginalLuckyEventRollBoundary(t *testing.T) {
	if !OriginalLuckyEventRollSucceeds(80, 8, 10) {
		t.Fatal("80/8=10 時 roll 10 應成功")
	}
	if OriginalLuckyEventRollSucceeds(80, 8, 11) {
		t.Fatal("80/8=10 時 roll 11 不應成功")
	}
	if OriginalLuckyEventRollSucceeds(7, 8, 1) {
		t.Fatal("商數為零時不應成功")
	}
	if OriginalLuckyEventRollSucceeds(80, 0, 1) || OriginalLuckyEventRollSucceeds(80, 8, 0) {
		t.Fatal("非法除數或擲骰應安全失敗")
	}
}

func TestOriginalEventScheduleThreshold(t *testing.T) {
	attempts := 0
	for i := 0; i < 5; i++ {
		threshold, next, ok := OriginalEventScheduleThreshold(100, attempts, 2)
		if !ok || threshold != 0 || next != attempts+1 {
			t.Fatalf("保護檢查 %d：threshold=%d next=%d ok=%v", i, threshold, next, ok)
		}
		attempts = next
	}
	want := []int{50, 66, 75, 80, 83}
	for difficulty, expected := range want {
		threshold, next, ok := OriginalEventScheduleThreshold(100, 5, difficulty)
		if !ok || threshold != expected || next != 5 {
			t.Errorf("難度 %d：threshold=%d next=%d ok=%v，want %d/5/true",
				difficulty, threshold, next, ok, expected)
		}
	}
	if _, _, ok := OriginalEventScheduleThreshold(10, 5, 5); ok {
		t.Fatal("非法難度不可冒充原版公式")
	}
}

func TestOriginalEventScheduleRollAndMinimumTurns(t *testing.T) {
	if !OriginalEventScheduleRollSucceeds(40, 40) || OriginalEventScheduleRollSucceeds(40, 41) {
		t.Fatal("排程擲骰必須使用 <= 邊界")
	}
	if OriginalEventScheduleRollSucceeds(600, 513) || OriginalEventScheduleRollSucceeds(1, 0) {
		t.Fatal("Random(512) 範圍外必須安全失敗")
	}
	want := map[int]int{2: 200, 19: 100, 20: 200, 21: 300, 22: 150, 23: 250, 24: 200}
	for id, turn := range want {
		if got := OriginalEventMinimumTurn(id); got != turn {
			t.Errorf("事件 %d 最早回合=%d，want %d", id, got, turn)
		}
	}
	if OriginalEventMinimumTurn(0) != 0 {
		t.Fatal("沒有額外限制的事件應回傳 0")
	}
}

func TestOriginalEventBCEffects(t *testing.T) {
	for _, tc := range []struct{ elapsed, want int }{{0, 100}, {19, 100}, {20, 200}, {99, 500}} {
		got, ok := OriginalMerchantDonation(tc.elapsed)
		if !ok || got != tc.want {
			t.Errorf("富商捐獻 elapsed=%d：got %d/%v，want %d/true", tc.elapsed, got, ok, tc.want)
		}
	}
	if _, ok := OriginalMerchantDonation(-1); ok {
		t.Fatal("負星曆不可冒充原版金額")
	}
	for _, tc := range []struct{ treasury, roll, want int }{{100, 1, 30}, {100, 21, 50}, {101, 21, 50}, {999, 1, 299}} {
		got, ok := OriginalPirateRaidLoss(tc.treasury, tc.roll)
		if !ok || got != tc.want {
			t.Errorf("海盜劫掠 treasury=%d roll=%d：got %d/%v，want %d/true", tc.treasury, tc.roll, got, ok, tc.want)
		}
	}
	if _, ok := OriginalPirateRaidLoss(99, 1); ok {
		t.Fatal("原版國庫不足 100 BC 時不可建立海盜劫掠")
	}
	if _, ok := OriginalPirateRaidLoss(100, 0); ok {
		t.Fatal("非法 Random(21) 結果不可接受")
	}
}

func TestOriginalComputerVirusLoss(t *testing.T) {
	for _, tc := range []struct{ progress, roll, want int }{
		{10, 1, 10}, {51, 1, 51}, {75, 50, 75}, {100, 50, 100}, {200, 1, 51},
	} {
		got, ok := OriginalComputerVirusLoss(tc.progress, tc.roll)
		if !ok || got != tc.want {
			t.Errorf("病毒 progress=%d roll=%d：got %d/%v，want %d/true", tc.progress, tc.roll, got, ok, tc.want)
		}
	}
	for _, tc := range []struct{ progress, roll int }{{9, 1}, {100, 0}, {100, 51}} {
		if _, ok := OriginalComputerVirusLoss(tc.progress, tc.roll); ok {
			t.Errorf("非法病毒輸入不應接受：progress=%d roll=%d", tc.progress, tc.roll)
		}
	}
}
