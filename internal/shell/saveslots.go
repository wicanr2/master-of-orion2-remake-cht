package shell

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// saveslots.go:存檔槽。
//
// remake 先前只有**一個**存檔檔案(每回合自動覆寫),主選單的 Continue / Load Game 點下去
// 都是「讀那一個檔」——所以「載入遊戲」在畫面上根本不存在,`LOADSAVE.LBX` 全 repo 零引用。
//
// --- 原版的槽位規格(openorion2 mainmenu.cpp,它自己是從原版 RE 出來的)---
//
//	#define SAVEGAME_SLOTS 10                  → 十個槽
//	if (slot == SAVEGAME_SLOTS - 1) → ESTR_SAVESLOT_AUTO
//	                                   → **最後一個槽固定是自動存檔**
//	findSavedGames():檔名格式 "SAVE%u.GAM"(`tmp != 2 || c != 'm' || slot < 1 || slot > 10`
//	                 的解析檢查反推出來的),槽號從 1 起算
//
// remake 沿用同一個規格:10 個槽、第 10 個是自動存檔。檔名用 remake 自己的 JSON 格式
// (`save1.json`..`save9.json` + `save.json`),不是原版 `.GAM`——原版格式由 internal/save
// **唯讀**解析,寫回原版格式不在範圍內。
//
// `save.json` 這個既有檔名刻意保留給自動存檔槽,舊存檔不會因為這次改動而消失。

// StartStardate 是遊戲起始星曆。
//
// 反組譯硬證(`Get_Time_Score_` @ 0x9E993):算已過回合用 `word[0x192FD8] - 0x88B8`,
// 0x88B8 = 35000 = 星曆 **3500.0 × 10**。remake 的 Turn 從 1 起算,故星曆 = 3500 + (Turn−1)。
const StartStardate = 3500

// StardateForTurn 把回合數換成顯示用星曆(原版存檔列表顯示的就是星曆,
// openorion2 `HSTR_SAVESLOT_STARDATE`)。
func StardateForTurn(turn int) string {
	if turn < 1 {
		turn = 1
	}
	return fmt.Sprintf("%d", StartStardate+turn-1)
}

// SaveSlotCount 是存檔槽數(原版 openorion2 `SAVEGAME_SLOTS`)。
const SaveSlotCount = 10

// AutoSaveSlot 是自動存檔所在的槽索引(原版:最後一個槽)。
const AutoSaveSlot = SaveSlotCount - 1

// SaveSlotInfo 是一個槽的摘要(供載入畫面逐列顯示)。
type SaveSlotInfo struct {
	Slot     int // 0-based
	Path     string
	Exists   bool
	Auto     bool   // 是否為自動存檔槽
	Turn     int    // 存檔當時的回合
	Stardate string // 由回合換算的星曆(原版列表顯示星曆,見 HSTR_SAVESLOT_STARDATE)
	Empire   string // 玩家帝國/領袖名
}

// SaveSlotPath 回傳某個槽的存檔路徑。dir 為存檔目錄。
//
// 自動存檔槽沿用既有的 `save.json`(舊存檔不會因為引入槽位而消失);其餘槽為 `saveN.json`,
// N 從 1 起算,與原版 `SAVE%u.GAM` 的槽號慣例一致。
func SaveSlotPath(dir string, slot int) string {
	if slot == AutoSaveSlot {
		return filepath.Join(dir, "save.json")
	}
	return filepath.Join(dir, fmt.Sprintf("save%d.json", slot+1))
}

// saveSlotPeek 只解出顯示需要的欄位,不重建整個 GameSession。
// 存檔格式版本不符時視為「不可讀」——列表寧可顯示空槽,也不要秀出一個點下去會爆的槽。
type saveSlotPeek struct {
	Version    int    `json:"version"`
	Turn       int    `json:"turn"`
	PlayerName string `json:"playerName"`
}

// ReadSaveSlot 讀取單一槽的摘要。檔案不存在或不可讀時回傳 Exists=false(不回錯誤——
// 空槽是正常狀態,不是故障)。
func ReadSaveSlot(dir string, slot int) SaveSlotInfo {
	info := SaveSlotInfo{Slot: slot, Path: SaveSlotPath(dir, slot), Auto: slot == AutoSaveSlot}
	data, err := os.ReadFile(info.Path)
	if err != nil {
		return info
	}
	var p saveSlotPeek
	if err := json.Unmarshal(data, &p); err != nil || p.Version != saveFormatVersion {
		return info
	}
	info.Exists = true
	info.Turn = p.Turn
	info.Stardate = StardateForTurn(p.Turn)
	info.Empire = p.PlayerName
	return info
}

// ReadSaveSlots 讀取全部槽的摘要(索引即槽號)。
func ReadSaveSlots(dir string) []SaveSlotInfo {
	out := make([]SaveSlotInfo, SaveSlotCount)
	for i := range out {
		out[i] = ReadSaveSlot(dir, i)
	}
	return out
}

// AnySaveExists 回傳是否至少有一個槽有存檔。
//
// 主選單的 Continue / Load Game 用它決定要不要停用——**原版無存檔時這兩顆是灰階不可按的**
// (2026-07-12 archive.org oracle 對照 issue #2 的結論,見 docs/tech/oracle-comparison-20260712.md)。
func AnySaveExists(dir string) bool {
	for i := 0; i < SaveSlotCount; i++ {
		if ReadSaveSlot(dir, i).Exists {
			return true
		}
	}
	return false
}

// LatestSaveSlot 回傳回合數最大的已存在槽(供 Continue 用:原版的 Continue 是接續最近的進度)。
// 沒有任何存檔時回 -1。
func LatestSaveSlot(dir string) int {
	best, bestTurn := -1, -1
	for i := 0; i < SaveSlotCount; i++ {
		s := ReadSaveSlot(dir, i)
		if s.Exists && s.Turn > bestTurn {
			best, bestTurn = i, s.Turn
		}
	}
	return best
}
