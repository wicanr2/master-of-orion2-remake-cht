package lbx

import (
	"encoding/binary"
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
