package shell

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
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
// remake 沿用同一個規格:10 個槽、第 10 個是自動存檔。remake 自己的 JSON 檔名為
// (`save1.json`..`save9.json` + `save.json`)；若槽位沒有 JSON，載入畫面也會以同槽的
// `SAVE1.GAM`..`SAVE10.GAM` 作為唯讀原版匯入來源。原版 `.GAM` 不會被 remake 寫回。
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
	Slot      int // 0-based
	Path      string
	Exists    bool
	Auto      bool   // 是否為自動存檔槽
	NativeGAM bool   // 是否由原版 .GAM 匯入；讀取後另存會轉成 remake JSON
	Turn      int    // 存檔當時的回合
	Stardate  string // 由回合換算的星曆(原版列表顯示星曆,見 HSTR_SAVESLOT_STARDATE)
	Empire    string // 玩家帝國/領袖名
	// Hotseat 是這個存檔是不是熱座局(原版存檔header的 `multiplayer` 欄位,
	// openorion2 `MultiplayerType`;載入視窗右側依此畫不同圖示)。
	Hotseat bool
	// Modified 是存檔檔案的修改時間(原版列表在星曆右邊顯示存檔日期時間,
	// openorion2 drawSlot:`buf.ftime("%x   %H:%M", ltime)` 畫在 `_x+122, y+14`)。
	Modified time.Time
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
	// Seats 只需要知道「有幾席」,所以解成 json.RawMessage 的陣列就夠——
	// 不必為了列表顯示把每一席的完整帝國狀態都解出來。
	Seats []json.RawMessage `json:"seats"`
}

// ReadSaveSlot 讀取單一槽的摘要。檔案不存在或不可讀時回傳 Exists=false(不回錯誤——
// 空槽是正常狀態,不是故障)。
func ReadSaveSlot(dir string, slot int) SaveSlotInfo {
	info := SaveSlotInfo{Slot: slot, Path: SaveSlotPath(dir, slot), Auto: slot == AutoSaveSlot}
	data, err := os.ReadFile(info.Path)
	if err == nil {
		var p saveSlotPeek
		if json.Unmarshal(data, &p) == nil && p.Version == saveFormatVersion {
			info.Exists = true
			info.Turn = p.Turn
			info.Stardate = StardateForTurn(p.Turn)
			info.Empire = p.PlayerName
			info.Hotseat = len(p.Seats) > 1
			if st, err := os.Stat(info.Path); err == nil {
				info.Modified = st.ModTime()
			}
			return info
		}
	}

	// JSON 優先；只有 JSON 不存在／版本不符時才探測同槽的原版檔，避免同一槽
	// 中已有 remake 存檔卻被舊資料覆蓋顯示。GAM 只匯入，不在這裡寫回。
	for _, path := range originalGAMSlotPaths(dir, slot) {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		session, report, err := LoadGAMSession(path)
		if err != nil || session == nil {
			continue
		}
		info.Path = path
		info.Exists = true
		info.NativeGAM = true
		info.Turn = report.Turn
		info.Stardate = fmt.Sprintf("%d", report.Stardate/10)
		info.Empire = session.PlayerName
		if st, err := os.Stat(path); err == nil {
			info.Modified = st.ModTime()
		}
		return info
	}
	return info
}

// originalGAMSlotPaths 回傳常見的 Windows／DOS 檔名大小寫變體。Linux 版若使用者
// 把原始檔直接放進 saves 目錄，也不應因副檔名大小寫而失去匯入入口。
func originalGAMSlotPaths(dir string, slot int) []string {
	n := slot + 1
	return []string{
		filepath.Join(dir, fmt.Sprintf("SAVE%d.GAM", n)),
		filepath.Join(dir, fmt.Sprintf("save%d.GAM", n)),
		filepath.Join(dir, fmt.Sprintf("SAVE%d.gam", n)),
		filepath.Join(dir, fmt.Sprintf("save%d.gam", n)),
	}
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
