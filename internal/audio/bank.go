package audio

// bank.go:從 LBX 封存檔取出音樂/音效 Clip,並解析 SOUND.LBX 的名稱表。
// 純資料層(不觸碰音訊裝置),可用合成/真檔測試。

import (
	"fmt"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
)

// archive 是本套件對 lbx.Archive 的最小依賴(便於測試以 stub 取代)。
type archive interface {
	Count() int
	Asset(id int) ([]byte, error)
}

// LoadMusic 取出一個音樂 LBX(STREAM/STREAMHD)內所有 WAV entry。
// 略過非 WAV 的 entry(如 entry 0 的彩蛋字串槽),回傳的 index 與原 entry id
// 不保證連續,故另回傳每條的來源 entry id 供對應表使用。
func LoadMusic(a archive) (clips []*Clip, entryIDs []int, err error) {
	for id := 0; id < a.Count(); id++ {
		raw, e := a.Asset(id)
		if e != nil {
			return nil, nil, fmt.Errorf("audio: 讀 entry %d: %w", id, e)
		}
		wav, ok := ExtractWAV(raw)
		if !ok {
			continue
		}
		c, e := DecodeWAV(wav)
		if e != nil {
			// 單條解碼失敗不致命(容忍偶發異常槽),記錄後略過。
			continue
		}
		clips = append(clips, c)
		entryIDs = append(entryIDs, id)
	}
	if len(clips) == 0 {
		return nil, nil, fmt.Errorf("audio: 此 LBX 無可用 WAV")
	}
	return clips, entryIDs, nil
}

// SoundBank 是 SOUND.LBX 的具名音效庫:名稱(如 "BUTTON1")→ 解碼後 Clip。
type SoundBank struct {
	byName map[string]*Clip
}

// LoadSoundBank 解析 SOUND.LBX:entry 0 為 8-byte 定長名稱表,順序對應
// entry 1..N 的 WAV。回傳可用名稱查詢的音效庫。
func LoadSoundBank(a archive) (*SoundBank, error) {
	if a.Count() < 2 {
		return nil, fmt.Errorf("audio: SOUND.LBX entry 數過少(%d)", a.Count())
	}
	nameTable, err := a.Asset(0)
	if err != nil {
		return nil, fmt.Errorf("audio: 讀名稱表: %w", err)
	}
	names := parseNameTable(nameTable, a.Count()-1)

	sb := &SoundBank{byName: make(map[string]*Clip)}
	for i, name := range names {
		id := i + 1 // 名稱 i 對應 sound entry i+1
		if id >= a.Count() {
			break
		}
		raw, e := a.Asset(id)
		if e != nil {
			continue
		}
		wav, ok := ExtractWAV(raw)
		if !ok {
			continue
		}
		c, e := DecodeWAV(wav)
		if e != nil {
			continue
		}
		if _, exists := sb.byName[name]; !exists { // 保留首次出現
			sb.byName[name] = c
		}
	}
	if len(sb.byName) == 0 {
		return nil, fmt.Errorf("audio: SOUND.LBX 未解出任何具名音效")
	}
	return sb, nil
}

// Clip 依名稱取音效;找不到回傳 nil。
func (sb *SoundBank) Clip(name string) *Clip { return sb.byName[name] }

// Names 回傳已載入的音效名稱數(供除錯/測試)。
func (sb *SoundBank) Len() int { return len(sb.byName) }

// soundNameRecordSize 是 SOUND.LBX entry 0 每筆記錄的實際跨距:對真實檔案
// 逐 byte 核對後確認為 8-byte 名稱(NUL 補齊)+ 12-byte 保留欄(用途未知,
// 樣本中恆為 0)。誤用 8-byte 跨距會在讀完第一筆後,把下一筆的保留欄當成
// 名稱(全 0 → cutName 回傳空字串 → 誤判為結尾),導致整張表只解出 1 筆。
const soundNameRecordSize = 20

// parseNameTable 從 SOUND.LBX entry 0 解析定長 20-byte 名稱記錄(名稱佔前
// 8-byte,NUL 補齊),最多 max 個。名稱為大寫英數/`-`/`_`,遇 "NEW" 標記
// (名稱表尾端的別名段起點)或非法記錄即停止。
func parseNameTable(tbl []byte, max int) []string {
	var names []string
	for i := 0; i+8 <= len(tbl) && len(names) < max; i += soundNameRecordSize {
		rec := tbl[i : i+8]
		name := cutName(rec)
		if name == "" || name == "NEW" || (len(name) >= 3 && name[:3] == "NEW") {
			break
		}
		names = append(names, name)
	}
	return names
}

// cutName 取 8-byte 記錄中到第一個 NUL 為止的可列印名稱;含非法字元回傳 ""。
func cutName(rec []byte) string {
	end := len(rec)
	for i, b := range rec {
		if b == 0 {
			end = i
			break
		}
	}
	if end == 0 {
		return ""
	}
	for _, b := range rec[:end] {
		if !isNameByte(b) {
			return ""
		}
	}
	return string(rec[:end])
}

func isNameByte(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') ||
		(b >= '0' && b <= '9') || b == '-' || b == '_' || b == '.'
}

// 確保 lbx.Archive 滿足 archive 介面(編譯期檢查)。
var _ archive = (*lbx.Archive)(nil)
