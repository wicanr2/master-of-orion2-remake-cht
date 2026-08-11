package netplay

// protocol.go:訊息型別與鎖步回合表。

// MsgKind 是訊息種類。
type MsgKind string

const (
	// KindHello 是連上之後的第一則:自報玩家編號與名字。
	KindHello MsgKind = "hello"
	// KindHelloChallenge 是主機在 hello 前送出的本次連線 challenge。
	KindHelloChallenge MsgKind = "hello_challenge"
	// KindPing / KindPong 是 Session 內部的心跳，不會交給遊戲規則層。
	KindPing MsgKind = "ping"
	KindPong MsgKind = "pong"
	// KindGameStart 是主機廣播的共同開局快照。收到後客戶端必須以該快照
	// 取代本地暫存對局，不能自行另骰一張星圖。
	KindGameStart MsgKind = "game_start"
	// KindTurnDone 是「我這一回合下完指令了」,帶這一回合的指令與狀態指紋。
	KindTurnDone MsgKind = "turn_done"
	// KindTurnReady 是收到全員指令、重播後的回合狀態指紋。
	// TurnDone 的指紋不能直接比較，因為各玩家送出時只套用了自己的指令；
	// 這一則是所有指令依玩家編號重播完成後才送出。
	KindTurnReady MsgKind = "turn_ready"
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
	// Protocol 只在握手時使用；零值保留給舊測試資料，正式握手會送 ProtocolVersion。
	Protocol int    `json:"protocol,omitempty"`
	Name     string `json:"name,omitempty"`
	// Auth 是對 challenge 的 HMAC proof，不傳送共享密碼本身。
	Auth string `json:"auth,omitempty"`
	// Challenge 只出現在 KindHelloChallenge。
	Challenge string `json:"challenge,omitempty"`
	// ResumeToken 只在 hello 與主機回給該客戶端的 roster 中使用。
	ResumeToken string `json:"resumeToken,omitempty"`
	Turn        int    `json:"turn,omitempty"`
	// StateHash 是**送出這則訊息時**的狀態指紋(shell.GameSession.StateHash)。
	// 鎖步的每一回合都比對一次,分岔才抓得到——等玩家自己發現「怎麼跟你看到的不一樣」
	// 就太晚了,那時候已經沒辦法回推是哪一回合開始歪的。
	StateHash string    `json:"stateHash,omitempty"`
	Commands  []Command `json:"commands,omitempty"`
	// Detail 給 KindDesync 帶說明。
	Detail string `json:"detail,omitempty"`
	// Payload 給 KindGameStart 帶序列化的 GameSession 快照；也可供未來協定
	// 擴充承載版本化資料。它仍受 MaxFrameBytes 保護。
	Payload string `json:"payload,omitempty"`
	// Text 給 KindChat 帶內文(見 chat.go)。
	Text string `json:"text,omitempty"`
}
