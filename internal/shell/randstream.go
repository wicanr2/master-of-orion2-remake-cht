package shell

import "math/rand"

// randstream.go:**可存檔的亂數流**。
//
// remake 的事件、星系發現、間諜、研究、議會、人口死亡與外交協議使用長壽命亂數流；它們與 `ground_invasion` /
// `monster` / `orbital_bombardment` 那種「每次用當下的回合數當種子」的一次性亂數不同——
// 它們從開局一路抽到終局。
//
// 問題是**存檔沒有記它們抽到第幾個數**:`restore()` 只從 `EventSeed` 重建,於是
//
//	存檔 → 讀檔 → 繼續玩   ← 事件序列**從頭開始**
//
// 兩個後果:①讀檔之後會重播同一批事件(存檔洗事件變得毫無成本);
// ②網路對戰時,一台中途讀檔的機器會與其他人分岔。第二點是「決定性化」這一項的核心,
// 而第一點本來就是個 bug。
//
// ============ 為什麼要自己算 Intn / Float64 ============
//
// 直覺解法是「記下抽了幾次,讀檔時重抽幾次跳過去」。但 `math/rand` 的 `Intn` 與
// `Float64` **從底層 source 取走的數量不一樣**(`Float64` 取一個 Int63,`Intn` 視 n 而定
// 可能取多個),所以「重抽 n 次」必須連**抽的種類**都一模一樣才會落在同一格。
//
// 這一檔改成直接騎在 `rand.Source64` 上,**每一次抽取恰好消耗一個 uint64**——
// 於是「跳過 n 次」就只是丟掉 n 個原始值,與當初抽的是 Intn 還是 Float64 無關。
//
// ⚠ `Intn` 用取餘數,有理論上的模偏差。n 在本專案是幾十到幾百的量級,64 位元下的偏差
// 約 2⁻⁵⁶,可以忽略;而且這幾條流**不是在重現原版的 PRNG**(原版那顆在
// `gamedata.NewOrigRand`),沒有逐位元對齊的義務。

// randStream 是一條可存檔的亂數流。
type randStream struct {
	seed int64
	n    int64 // 已抽取次數(存檔記這個)
	src  rand.Source64
}

// newRandStream 開一條流。
func newRandStream(seed int64) *randStream {
	return &randStream{seed: seed, src: rand.NewSource(seed).(rand.Source64)}
}

// restoreRandStream 開一條流並快轉到第 n 次抽取之後(讀檔用)。
func restoreRandStream(seed, n int64) *randStream {
	s := newRandStream(seed)
	for i := int64(0); i < n; i++ {
		s.next()
	}
	return s
}

// next 取一個原始值(**每次抽取恰好消耗一個**,快轉才成立)。
func (s *randStream) next() uint64 {
	s.n++
	return s.src.Uint64()
}

// Draws 回傳已抽取次數(存檔用)。
func (s *randStream) Draws() int64 {
	if s == nil {
		return 0
	}
	return s.n
}

// Intn 回傳 [0, n) 的整數;n <= 0 一律回 0(與 math/rand 會 panic 不同——
// 這幾條流的呼叫端散落在事件/發現/間諜裡,為一個空清單 panic 掉整局不划算)。
func (s *randStream) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	return int(s.next() % uint64(n))
}

// Float64 回傳 [0, 1) 的浮點數(取高 53 位,與 math/rand 同一個做法)。
func (s *randStream) Float64() float64 {
	return float64(s.next()>>11) / (1 << 53)
}
