package audio

// bank_test.go:用 stub archive(不需真 lbx)驗證名稱表解析與 SoundBank 組裝。

import (
	"fmt"
	"testing"
)

// stubArchive 是 archive 介面的最小測試替身。
type stubArchive struct {
	assets [][]byte
}

func (s *stubArchive) Count() int { return len(s.assets) }

func (s *stubArchive) Asset(id int) ([]byte, error) {
	if id < 0 || id >= len(s.assets) {
		return nil, fmt.Errorf("stub: entry %d 越界", id)
	}
	return s.assets[id], nil
}

var _ archive = (*stubArchive)(nil)

// nameRec 組出與 SOUND.LBX entry 0 相同格式的 20-byte 定長記錄:
// 前 8-byte 為 NUL 補齊的名稱,後 12-byte 為保留欄(真實檔案中恆為 0)。
func nameRec(s string) []byte {
	b := make([]byte, soundNameRecordSize)
	copy(b, s) // 名稱只佔前 8 byte,其餘(含名稱後段與保留欄)維持零值
	return b
}

func TestParseNameTable(t *testing.T) {
	// entry 0:3 個具名記錄 + "NEW....." 終止標記(其後應被忽略)。
	var nameTable []byte
	nameTable = append(nameTable, nameRec("BUTTON1")...)
	nameTable = append(nameTable, nameRec("BUTTON2")...)
	nameTable = append(nameTable, nameRec("CLICK1")...)
	nameTable = append(nameTable, nameRec("NEW.....")...) // 名稱佔滿 8 byte,無 NUL

	wavFor := func(seed byte) []byte {
		return buildWAV(22050, 1, 8, []byte{seed, seed + 1, seed + 2})
	}

	arc := &stubArchive{assets: [][]byte{
		nameTable,  // entry 0:名稱表
		wavFor(10), // entry 1 -> BUTTON1
		wavFor(20), // entry 2 -> BUTTON2
		wavFor(30), // entry 3 -> CLICK1
		wavFor(40), // entry 4:NEW 之後,不應被任何名稱取用
	}}

	sb, err := LoadSoundBank(arc)
	if err != nil {
		t.Fatalf("LoadSoundBank 失敗: %v", err)
	}

	if got := sb.Len(); got != 3 {
		t.Errorf("SoundBank.Len() = %d, want 3(應在 NEW 標記處停止)", got)
	}

	for _, name := range []string{"BUTTON1", "BUTTON2", "CLICK1"} {
		if c := sb.Clip(name); c == nil {
			t.Errorf("sb.Clip(%q) = nil, want 非 nil", name)
		} else if len(c.PCM) == 0 {
			t.Errorf("sb.Clip(%q).PCM 為空", name)
		}
	}

	// "NEW....." 是終止標記,不應被當成音效名稱收進來。
	if c := sb.Clip("NEW....."); c != nil {
		t.Error("sb.Clip(\"NEW.....\") 應為 nil(終止標記不應被收錄),卻取到 Clip")
	}
	// entry 4 在 NEW 標記之後,沒有對應名稱,不應能以任何方式取得。
	if c := sb.Clip(""); c != nil {
		t.Error("空字串不應對應任何 Clip")
	}
}

func TestParseNameTable_StopsAtNEW(t *testing.T) {
	var tbl []byte
	tbl = append(tbl, nameRec("FOO1")...)
	tbl = append(tbl, nameRec("NEW.....")...)
	tbl = append(tbl, nameRec("SHOULDNOTAPPEAR")...) // NEW 之後,不應被解析到

	names := parseNameTable(tbl, 10)
	if len(names) != 1 || names[0] != "FOO1" {
		t.Errorf("parseNameTable = %v, want [FOO1]", names)
	}
}

func TestParseNameTable_RespectsMax(t *testing.T) {
	var tbl []byte
	tbl = append(tbl, nameRec("A")...)
	tbl = append(tbl, nameRec("B")...)
	tbl = append(tbl, nameRec("C")...)

	names := parseNameTable(tbl, 2)
	if len(names) != 2 || names[0] != "A" || names[1] != "B" {
		t.Errorf("parseNameTable(max=2) = %v, want [A B]", names)
	}
}

func TestLoadSoundBank_TooFewEntries(t *testing.T) {
	arc := &stubArchive{assets: [][]byte{{0}}} // Count()==1 < 2
	if _, err := LoadSoundBank(arc); err == nil {
		t.Fatal("entry 數過少應回傳 error,卻成功")
	}
}
