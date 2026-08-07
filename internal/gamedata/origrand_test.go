package gamedata

import "testing"

// origrand_test.go:原版 LCG 的護欄。
//
// 這支東西的價值全在「跟原版一模一樣」——差一個常數、少一個拒絕取樣分支,序列就整條偏掉,
// 而那種偏掉**不會有任何症狀**(數字看起來還是很隨機),只有拿去重現原版佈局時才發現對不上。
// 所以這裡驗的是可以手算的性質,不是「看起來夠亂」。

// TestOrigRandLCGStep:單步遞推要等於 state×0x41C64E6D + 0x3039(32-bit 迴繞)。
// 用 N(1) 觸發一次遞推——n=1 時 bucket = 0xFFFFFFFF、limit = 0xFFFFFFFF,
// 除非抽到 0xFFFFFFFF 否則一次就過,所以剛好可以當「單步」用。
func TestOrigRandLCGStep(t *testing.T) {
	for _, seed := range []uint32{0, 1, 12345, 0xDEADBEEF} {
		r := NewOrigRand(seed)
		want := seed*origRandMul + origRandInc
		r.N(1)
		if r.State() != want {
			t.Errorf("seed %d:遞推後 state = %d,want %d", seed, r.State(), want)
		}
	}
}

// TestOrigRandSeedZeroFirstDraw:seed 0 的第一次遞推是 0×mul + 12345 = 12345。
// 殖民地 0 的建築擺法就是從這個數開始的,釘住它等於釘住整條序列的起點。
func TestOrigRandSeedZeroFirstDraw(t *testing.T) {
	r := NewOrigRand(0)
	r.N(1)
	if r.State() != origRandInc {
		t.Errorf("seed 0 第一次遞推應為 %d,實得 %d", origRandInc, r.State())
	}
}

// TestOrigRandIsOneBased:回傳範圍是 1..n,不是 0..n-1。
// 原版滿地都是 `Random(3) == 1` 這種寫法,換成 0-based 會把機率挪一格。
func TestOrigRandIsOneBased(t *testing.T) {
	r := NewOrigRand(7)
	seen := map[int]int{}
	for i := 0; i < 20000; i++ {
		v := r.N(6)
		if v < 1 || v > 6 {
			t.Fatalf("N(6) 回傳 %d,超出 1..6", v)
		}
		seen[v]++
	}
	if len(seen) != 6 {
		t.Errorf("N(6) 只出現 %d 種值,應為 6 種", len(seen))
	}
	// 拒絕取樣的重點:每個值都該接近 1/6 ≈ 0.1667。抽 20000 次,任一面偏離 ±0.01 就有問題。
	for v, n := range seen {
		if p := float64(n) / 20000; p < 0.157 || p > 0.177 {
			t.Errorf("N(6) = %d 的頻率 %.4f 偏離 1/6 太多(拒絕取樣可能沒照抄)", v, p)
		}
	}
}

// TestOrigRandZeroReturnsOne:n = 0 走原版那條 `cmp n, 0` 分支,直接回 1 且**不遞推**。
func TestOrigRandZeroReturnsOne(t *testing.T) {
	r := NewOrigRand(99)
	if got := r.N(0); got != 1 {
		t.Errorf("N(0) 應回 1,實得 %d", got)
	}
	if r.State() != 99 {
		t.Errorf("N(0) 不該動到 state,實得 %d", r.State())
	}
}

// TestOrigRandDeterministic:同一個種子跑出同一串。這正是原版拿它來做
// 「同一個殖民地每次進去長得一樣」的依據。
func TestOrigRandDeterministic(t *testing.T) {
	draw := func() []int {
		r := NewOrigRand(3)
		out := make([]int, 50)
		for i := range out {
			out[i] = r.N(36)
		}
		return out
	}
	a, b := draw(), draw()
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("同種子第 %d 抽不同:%d vs %d", i, a[i], b[i])
		}
	}
	// 不同種子要走出不同序列(否則等於 seed 沒接上)。
	r := NewOrigRand(4)
	same := true
	for i := range a {
		if r.N(36) != a[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("seed 3 與 seed 4 的序列相同——seed 沒真的接上")
	}
}
