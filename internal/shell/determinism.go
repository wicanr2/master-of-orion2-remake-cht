package shell

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

// determinism.go:**狀態指紋**——網路多人的地基。
//
// 網路對戰要能成立,前提是「同樣的起始種子 + 同樣的指令序列 → 每一台機器算出同樣的狀態」。
// 原版把這件事做在 `Net_Next_Turn` 那條路徑上;remake 連傳輸層都還沒有,但**決定性這件事
// 不需要等傳輸層**——它是規則層自己的性質,而且現在就可以測。
//
// 這一檔提供兩樣東西:
//
//	StateHash()      整局狀態的 SHA-256 指紋(用存檔快照當正規形式)
//	StateFingerprint 人看得懂的短指紋(前 8 個十六進位字元),放進偵錯輸出用
//
// ============ 為什麼用存檔快照當正規形式 ============
//
// `snapshot()` 已經是「這局遊戲的完整可序列化狀態」——存讀檔靠它、熱座交接靠它。
// 拿它當指紋來源有三個好處:
//
//  1. **不必另外維護一份欄位清單。** 新增欄位只要進得了存檔就自動進得了指紋;
//     反過來說,**進不了存檔的欄位本來就不該影響對局結果**(那是 UI 狀態)。
//  2. `encoding/json` 對 map 的鍵**保證排序**,所以 `ColonyBuildings` 這類 map
//     不會因為 Go 的隨機迭代順序讓兩台機器算出不同指紋。
//  3. 指紋不合時可以直接 diff 兩邊的存檔 JSON,看得出是哪個欄位分岔。
//
// 七條長壽命亂數流（事件／發現／間諜／研究／議會／人口／外交協議）的**位置也在快照裡**：
// 它們記的是「已經抽了幾次」,讀檔時快轉回原位(見 randstream.go)。
//
// 這一點 2026-08-07 之前不成立——那時候讀檔會讓整條流從頭開始,於是
// 「存檔 → 讀檔 → 繼續玩」會重播同一批事件(存檔洗事件毫無成本),
// 而網路對戰時中途讀檔的那台會與其他人分岔。`determinism_test.go` 的
// `TestLoadedGameContinuesTheSameRandomStreams` 釘住修好後的行為。

// StateHash 回傳整局狀態的 SHA-256 指紋(十六進位小寫)。
//
// 序列化失敗時回空字串——呼叫端要把空字串當成「算不出指紋」,不要當成「兩邊一樣」。
func (s *GameSession) StateHash() string {
	if s == nil {
		return ""
	}
	b, err := json.Marshal(s.snapshot())
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// StateFingerprint 是 StateHash 的前 8 個字元,給人看的短版本(log / 畫面角落)。
func (s *GameSession) StateFingerprint() string {
	h := s.StateHash()
	if len(h) < 8 {
		return h
	}
	return h[:8]
}
