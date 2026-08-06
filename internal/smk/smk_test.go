package smk

import (
	"os"
	"testing"
)

// smk_test.go:Smacker 解碼器。
//
// 純合成資料驗不了什麼——這個格式的錯誤都是「位元流錯位」,合成一個合法的 Smacker 檔
// 本身就等於再寫一次編碼器。所以這裡分兩層:①不需要遊戲資料的邊界/防呆 ②有真檔時
// (環境變數 MOO2_SMK_TEST 指向一個 .smk)跑真正的驗收——解得完、幀邊界對得上、
// 樹在標頭上界內。

func TestParseHeaderRejectsGarbage(t *testing.T) {
	if _, err := ParseHeader(nil); err == nil {
		t.Error("空資料應該回錯誤")
	}
	if _, err := ParseHeader(make([]byte, headerSize)); err == nil {
		t.Error("簽章全零應該回錯誤(判斷是不是 Smacker 只能靠簽章,副檔名在 MOO2 裡不可信)")
	}
	bad := make([]byte, headerSize)
	copy(bad, "SMK2")
	if _, err := ParseHeader(bad); err == nil {
		t.Error("寬高幀數全零應該回錯誤")
	}
}

func TestBitReaderIsLSBFirst(t *testing.T) {
	// 0b1011_0010 = 0xB2。LSB-first 讀出來應該是 0,1,0,0,1,1,0,1。
	br := newBitReader([]byte{0xB2})
	want := []uint32{0, 1, 0, 0, 1, 1, 0, 1}
	for i, w := range want {
		if got := br.bit(); got != w {
			t.Fatalf("第 %d 個位元應為 %d,實得 %d(位元序錯的話整個格式都解不開)", i, w, got)
		}
	}
	// 讀過尾端一律回 0(影像位元流常常只用到最後一個位元組的一部分)。
	if got := br.bit(); got != 0 {
		t.Errorf("讀過尾端應回 0,實得 %d", got)
	}
}

func TestBitReaderMultiBit(t *testing.T) {
	br := newBitReader([]byte{0x34, 0x12})
	if got := br.bits(16); got != 0x1234 {
		t.Errorf("16 位元應讀成 0x1234(小端),實得 %#x", got)
	}
}

// 真檔驗收:MOO2_SMK_TEST=/path/to/INTRO.LBX scripts/test.sh ./internal/smk/
func TestRealFile(t *testing.T) {
	path := os.Getenv("MOO2_SMK_TEST")
	if path == "" {
		t.Skip("未設 MOO2_SMK_TEST,略過真檔驗收")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("讀檔: %v", err)
	}
	d, err := Open(data)
	if err != nil {
		t.Fatalf("解析: %v", err)
	}
	// 幀資料全部吃完之後不該有殘餘——有殘餘就代表幀大小表的解讀錯了,
	// 那會讓每一幀的起點都錯位。
	if tr := d.Trailing(); tr != 0 {
		t.Errorf("幀資料吃完後殘餘 %d bytes(應為 0)", tr)
	}
	if w := d.TreeWarnings(); len(w) > 0 {
		t.Errorf("樹超出標頭上界: %v", w)
	}
	// 全片要解得完而且不報位元流錯位。
	for i := 0; i < d.H.Frames; i++ {
		if _, _, err := d.DecodeNext(); err != nil {
			t.Fatalf("第 %d 幀: %v", i, err)
		}
	}
}
