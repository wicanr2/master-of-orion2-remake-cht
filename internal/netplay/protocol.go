package netplay

// protocol.go:訊息型別與鎖步回合表。

// MsgKind 是訊息種類。
type MsgKind string

const (
	// KindHello 是連上之後的第一則:自報玩家編號與名字。
	KindHello MsgKind = "hello"
	// KindTurnDone 是「我這一回合下完指令了」,帶這一回合的指令與狀態指紋。
	KindTurnDone MsgKind = "turn_done"
	// KindDesync 是「我算出來的狀態跟你不一樣」——收到就該停下來,不要繼續玩。
	KindDesync MsgKind = "desync"
)

// Command 是一條玩家指令。
//
// 用「名稱 + 參數」而不是替每種指令定一個型別:傳輸層不需要知道指令的語意,
// 它只負責把同樣的序列送到每一台機器上。解釋是 shell 的事。
//
// ⚠ `Args` 用 []int 而不是 map:**map 的 JSON 鍵順序雖然穩定,但參數靠名字對應
// 會讓新增參數變成相容性問題**;位置參數配上「指令名決定參數意義」的約定簡單得多。
type Command struct {
	Name string `json:"name"`
	Args []int  `json:"args,omitempty"`
	// Text 給少數需要字串參數的指令(建造項目名之類)。
	Text string `json:"text,omitempty"`
}

// Message 是走在線上的一則訊息。
type Message struct {
	Kind   MsgKind `json:"kind"`
	Player int     `json:"player"`
	Name   string  `json:"name,omitempty"`
	Turn   int     `json:"turn,omitempty"`
	// StateHash 是**送出這則訊息時**的狀態指紋(shell.GameSession.StateHash)。
	// 鎖步的每一回合都比對一次,分岔才抓得到——等玩家自己發現「怎麼跟你看到的不一樣」
	// 就太晚了,那時候已經沒辦法回推是哪一回合開始歪的。
	StateHash string    `json:"stateHash,omitempty"`
	Commands  []Command `json:"commands,omitempty"`
	// Detail 給 KindDesync 帶說明。
	Detail string `json:"detail,omitempty"`
}
