package lbx

import (
	"encoding/binary"
	"os"
	"testing"
)

func TestParseFixedStrings(t *testing.T) {
	// count=3, bufsize=8
	buf := make([]byte, 4+3*8)
	binary.LittleEndian.PutUint16(buf[0:], 3)
	binary.LittleEndian.PutUint16(buf[2:], 8)
	copy(buf[4:], "Ion\x00")            // 第 0 條
	copy(buf[12:], "Fusion\x00")        // 第 1 條
	copy(buf[20:], "Antimat\x00")       // 第 2 條(7 字元 + NUL)
	got, err := ParseFixedStrings(buf)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Ion", "Fusion", "Antimat"}
	if len(got) != 3 {
		t.Fatalf("條數 = %d", len(got))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] = %q,預期 %q", i, got[i], want[i])
		}
	}
}

func TestParseFixedStringsRejectsGarbage(t *testing.T) {
	// count/bufsize 過大 → 資料不足
	buf := []byte{0xff, 0xff, 0xff, 0xff, 0x00}
	if _, err := ParseFixedStrings(buf); err == nil {
		t.Fatal("超大 count/bufsize 應回 error")
	}
}

func TestParseCStrings(t *testing.T) {
	data := []byte("Starting Tech\x00Advanced Biology\x00Military Tactics\x00\x00\x00")
	got := ParseCStrings(data, 0)
	want := []string{"Starting Tech", "Advanced Biology", "Military Tactics"}
	if len(got) != len(want) {
		t.Fatalf("條數 = %d,預期 %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("[%d] = %q,預期 %q", i, got[i], want[i])
		}
	}
}

func TestParseHelp(t *testing.T) {
	const size = helpEntrySize
	// count=2,兩則記錄。
	buf := make([]byte, 4+2*size)
	binary.LittleEndian.PutUint16(buf[0:], 2)
	binary.LittleEndian.PutUint16(buf[2:], size)

	writeRec := func(base int, title, archive string, asset, frame uint16, section uint8, text string) {
		rec := buf[base : base+size]
		copy(rec[0:], title+"\x00")
		copy(rec[80:], archive+"\x00")
		binary.LittleEndian.PutUint16(rec[94:], asset)
		binary.LittleEndian.PutUint16(rec[96:], frame)
		rec[98] = section
		// rec[99:103] nextParagraph 留 0
		copy(rec[103:], text+"\x00")
	}
	writeRec(4, "Adamantium Armor", "vfx.lbx", 42, 3, 1,
		"Ships with this armor have 8 times the base structure.")
	writeRec(4+size, "No Tech", "", 0, 0, 0, "No description")

	got, err := ParseHelp(buf)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("條目數 = %d,預期 2", len(got))
	}
	e := got[0]
	if e.Title != "Adamantium Armor" || e.Archive != "vfx.lbx" ||
		e.AssetID != 42 || e.Frame != 3 || e.Section != 1 ||
		e.Text != "Ships with this armor have 8 times the base structure." {
		t.Errorf("記錄0 解析錯誤: %+v", e)
	}
	if got[1].Title != "No Tech" || got[1].Archive != "" || got[1].Text != "No description" {
		t.Errorf("記錄1 解析錯誤: %+v", got[1])
	}
}

func TestParseHelpRejectsBadSize(t *testing.T) {
	buf := make([]byte, 4+100)
	binary.LittleEndian.PutUint16(buf[0:], 1)
	binary.LittleEndian.PutUint16(buf[2:], 100) // < helpEntrySize
	if _, err := ParseHelp(buf); err == nil {
		t.Error("記錄寬過小應回錯誤")
	}
}

// TestParseHelpRealFile 對真實 HELP.LBX 驗證(設 MOO2_HELP_LBX 指向玩家自備檔才跑;
// 版權檔不入 repo)。確認能解析且首則標題如預期。
func TestParseHelpRealFile(t *testing.T) {
	path := os.Getenv("MOO2_HELP_LBX")
	if path == "" {
		t.Skip("未設 MOO2_HELP_LBX")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fi, _ := f.Stat()
	arch, err := Open(f, fi.Size())
	if err != nil {
		t.Fatal(err)
	}
	raw, err := arch.Asset(0)
	if err != nil {
		t.Fatal(err)
	}
	entries, err := ParseHelp(raw)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("HELP 條目數 = %d", len(entries))
	if len(entries) < 100 {
		t.Errorf("條目數 %d 過少,疑似解析錯位", len(entries))
	}
	if entries[1].Title != "Achilles Targeting Unit" {
		t.Errorf("entries[1].Title = %q,預期 Achilles Targeting Unit", entries[1].Title)
	}
}
