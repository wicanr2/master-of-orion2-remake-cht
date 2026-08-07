package gamedata

// origrand.go:**原版的亂數產生器**(`Random_` @ 0x1247A0 / `Set_Random_Seed_` @ 0x124820 /
// `Get_Random_Seed_` @ 0x12484C)。
//
// 為什麼要照抄一個 1996 年的 LCG:MOO2 有好幾處**靠特定種子重現特定結果**——殖民地的建築
// 擺法就是 `Set_Random_Seed(colonyIdx)` 起手,同一個殖民地每次進去長得一樣。用 Go 的
// `math/rand` 換不出同一組數,那些地方就永遠對不上原版。
//
// ============ 一手來源(反組譯,三支加起來不到 40 行組語)============
//
//	`Set_Random_Seed_`:state = eax。就這樣,沒有 scramble。
//	`Get_Random_Seed_`:回傳 state(原版會先存起來、用完再還原,見
//	                    `Make_Bldg_Array_For_Colony_` 頭尾那對呼叫)。
//	`Random_`(eax = n,回傳 **1..n**):
//	    xor edx, edx / mov eax, 0FFFFFFFFh / div n   → bucket = 0xFFFFFFFF / n
//	    imul eax, n                                  → limit  = bucket × n
//	    cmp n, 0 / jnz …                             → n == 0 直接回 1
//	  loop:
//	    imul eax, state, 41C64E6Dh / add eax, 3039h / mov state, eax
//	    cmp limit, state / jbe loop                  → state >= limit 就重抽
//	    div bucket / inc eax                         → 回 state/bucket + 1
//
// 乘數 0x41C64E6D = 1103515245、增量 0x3039 = 12345 —— 就是 ANSI C `rand()` 那組常數,
// 但**保留完整 32-bit state**(沒有 `>>16`、沒有 `& 0x7FFFFFFF`)。
//
// 那個 while 迴圈是**拒絕取樣**:把 [0, limit) 切成 n 個等寬桶,落在 limit 之外就重抽,
// 所以 1..n 是嚴格等機率的(不是常見的 `% n` 偏差版)。照抄很重要——少了它,
// 序列會從第一次「本該重抽」的地方起整條偏掉。
type OrigRand struct {
	state uint32
}

// 原版的 LCG 常數(見檔頭)。
const (
	origRandMul = 0x41C64E6D
	origRandInc = 0x3039
)

// NewOrigRand 建一個以 seed 起始的產生器。原版 `Set_Random_Seed_` 就是直接指派,
// 所以 seed 0 是合法的(殖民地 0 的建築擺法用的就是它)。
func NewOrigRand(seed uint32) *OrigRand { return &OrigRand{state: seed} }

// Seed 重設狀態(= `Set_Random_Seed_`)。
func (r *OrigRand) Seed(v uint32) { r.state = v }

// State 取出目前狀態(= `Get_Random_Seed_`)。原版會在需要固定序列的地方先存後還原。
func (r *OrigRand) State() uint32 { return r.state }

// N 回傳 **1..n**(= `Random_(n)`)。n <= 0 時回 1,與原版 `cmp n, 0` 那條分支一致。
//
// ⚠ 回傳是 **1-based**,不是 0..n-1。原版到處都寫 `Random(3) == 1` 這種判斷,
// 換成 0-based 會把機率整個挪一格。
func (r *OrigRand) N(n int) int {
	if n <= 0 {
		return 1
	}
	un := uint32(n)
	bucket := uint32(0xFFFFFFFF) / un
	limit := bucket * un
	for {
		r.state = r.state*origRandMul + origRandInc
		if r.state < limit {
			break
		}
	}
	return int(r.state/bucket) + 1
}
