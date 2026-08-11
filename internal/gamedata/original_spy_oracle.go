package gamedata

// original_spy_oracle.go 保存同一份 Orion2.exe.i64 由 IDA Pro 9.4 追回的
// SABOTAGE／Spy-vs-Spy raw 低階形狀。這些函式只描述可直接從指令與 operand 證實的
// 位元與算式，不把原版尚未命名的帝國欄位硬翻成遊戲語意。
//
// 證據輸入：Orion2.exe.i64 SHA-256
// 4a01791fcf877ed87a740a54748694ab34a02675e3117dac052aeaa3f883944e，IDA Pro 9.4，
// IDA 線性位址。原始位址與推論等級見 docs/re/oracle-static-ida-20260811.md。

const (
	// OriginalSpyRelationshipByteOffset 是每個 0xEA9 bytes 帝國記錄內的 raw 關係矩陣
	// 起點；sub_1026CF／sub_1026F1／sub_10278D 都以 record + 0xE57 + otherEmpire
	// 讀寫同一個 byte。
	OriginalSpyRelationshipByteOffset      = 0xE57
	OriginalSpyRelationshipCountMask  byte = 0x3F
	OriginalSpyRelationshipModeMask   byte = 0xC0
	OriginalSpyScoreTableAAddress          = 0x1ACE78 // word_1ACE78
	OriginalSpyScoreTableBAddress          = 0x1ACE7A // word_1ACE7A
)

// OriginalSpyRelationshipCount 對應 sub_1026CF @ 0x1026CF 的 `& 0x3F`。
// 這是 raw byte 的低 6 位數量，不替它命名成「Spy」或「Agent」，因為同一 byte
// 會依 row／column 方向被不同任務路徑消費。
func OriginalSpyRelationshipCount(raw byte) int {
	return int(raw & OriginalSpyRelationshipCountMask)
}

// OriginalSpyRelationshipMode 對應 sub_1026F1 @ 0x1026F1 的 `>> 6`。
// 這只保留原版在 sub_1014A4 的 0/1/2 分支選擇；4 種值的完整高階語意仍未知。
func OriginalSpyRelationshipMode(raw byte) int {
	return int((raw & OriginalSpyRelationshipModeMask) >> 6)
}

// PackOriginalSpyRelationship 是 sub_10278D @ 0x10278D 的非破壞性等價 adapter：
// 寫入低 6 位、保留高 2 位的 raw mode。它只供 remake 的 raw fixture／邊界測試使用，
// 不會把舊 JSON 強行改成原版 packed 存檔格式。
func PackOriginalSpyRelationship(count, mode int) byte {
	if count < 0 {
		count = 0
	}
	if count > 63 {
		count = 63
	}
	if mode < 0 {
		mode = 0
	}
	if mode > 3 {
		mode = 3
	}
	return byte(count) | byte(mode<<6)
}

// OriginalSpyScoreHelper 對應 sub_101483 @ 0x101483 的精確三段式 helper：
// n<=5 → 2n；6..10 → n+5；n>10 → floor((n-10)/2)+15。
// 原版呼叫端把它餵給 0..63 的 packed low-six-bit 數量；本函式不偷偷夾限，讓
// 反組譯算式與測試可以逐值對照。
func OriginalSpyScoreHelper(n int) int {
	switch {
	case n <= 5:
		return n * 2
	case n <= 10:
		return n + 5
	default:
		return (n-10)/2 + 15
	}
}
