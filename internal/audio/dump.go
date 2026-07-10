package audio

// dump.go:把 LBX entry 內的原始 WAV bytes 原封抽出(不經 DecodeWAV 轉檔),
// 供人耳試聽、建立曲目/音效對應表(見 docs/tech/ 內的對應表文件)。
//
// 與 bank.go 的 LoadMusic/LoadSoundBank 不同:那兩者回傳「轉成 16-bit stereo
// 給 ebiten 播放」的 Clip;這裡回傳的是 ExtractWAV 切出的原始 RIFF/WAVE bytes,
// 位元組與原版磁片上的資料一致,才適合拿來聽「原始音質」。

import "fmt"

// RawWAV 是一筆抽出的原始 WAV。Name 僅音效(SOUND.LBX)有值,音樂為空字串。
type RawWAV struct {
	Index int    // 來源 LBX entry id
	Name  string // 具名時的音效名稱(SOUND.LBX 名稱表對應),音樂為 ""
	WAV   []byte // ExtractWAV 切出的原始 RIFF/WAVE bytes
}

// RawMusic 掃描一個音樂 LBX(STREAM/STREAMHD)的所有 entry,原封抽出每個
// 能定位到 WAV 的 entry。非 WAV 的 entry(如彩蛋字串槽)略過不算錯誤。
func RawMusic(a archive) ([]RawWAV, error) {
	var out []RawWAV
	for id := 0; id < a.Count(); id++ {
		raw, err := a.Asset(id)
		if err != nil {
			return nil, fmt.Errorf("audio: 讀 entry %d: %w", id, err)
		}
		wav, ok := ExtractWAV(raw)
		if !ok {
			continue
		}
		out = append(out, RawWAV{Index: id, WAV: wav})
	}
	return out, nil
}

// RawSounds 掃描 SOUND.LBX:entry 0 為名稱表,entry 1..N 為音效。原封抽出每個
// 能定位到 WAV 的 entry,並附上名稱表對應到的名稱(缺名則 Name 為空字串)。
func RawSounds(a archive) ([]RawWAV, error) {
	if a.Count() < 1 {
		return nil, fmt.Errorf("audio: SOUND.LBX entry 數過少(%d)", a.Count())
	}
	var names []string
	if a.Count() >= 2 {
		nameTable, err := a.Asset(0)
		if err != nil {
			return nil, fmt.Errorf("audio: 讀名稱表: %w", err)
		}
		names = parseNameTable(nameTable, a.Count()-1)
	}

	var out []RawWAV
	for id := 1; id < a.Count(); id++ {
		raw, err := a.Asset(id)
		if err != nil {
			return nil, fmt.Errorf("audio: 讀 entry %d: %w", id, err)
		}
		wav, ok := ExtractWAV(raw)
		if !ok {
			continue
		}
		name := ""
		if idx := id - 1; idx < len(names) {
			name = names[idx]
		}
		out = append(out, RawWAV{Index: id, Name: name, WAV: wav})
	}
	return out, nil
}
