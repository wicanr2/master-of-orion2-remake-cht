package lbx

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

// buildSyntheticLBX 依 lbx 格式手工組出一個含指定資產的 .lbx 位元組串,
// 用來在無版權測資的情況下驗證解析器。
func buildSyntheticLBX(assets [][]byte) []byte {
	n := len(assets)
	headerSize := 8 + 4*(n+1) // assetCount+magic+version(8) + (n+1) 個 u32 offset
	var buf bytes.Buffer

	binary.Write(&buf, binary.LittleEndian, uint16(n))
	binary.Write(&buf, binary.LittleEndian, uint16(lbxMagic))
	binary.Write(&buf, binary.LittleEndian, uint32(0)) // version/旗標

	// offset 表:off0 = headerSize,之後每筆累加前一個資產大小。
	off := uint32(headerSize)
	binary.Write(&buf, binary.LittleEndian, off)
	for _, a := range assets {
		off += uint32(len(a))
		binary.Write(&buf, binary.LittleEndian, off)
	}
	// 資產資料。
	for _, a := range assets {
		buf.Write(a)
	}
	return buf.Bytes()
}

func TestOpenSynthetic(t *testing.T) {
	assets := [][]byte{
		[]byte("HELLO"),
		[]byte("WORLD!!"),
		{0x00, 0x01, 0x02, 0x03},
	}
	raw := buildSyntheticLBX(assets)

	a, err := Open(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatalf("Open 失敗: %v", err)
	}
	if a.Count() != len(assets) {
		t.Fatalf("Count = %d,預期 %d", a.Count(), len(assets))
	}
	for i, want := range assets {
		got, err := a.Asset(i)
		if err != nil {
			t.Fatalf("Asset(%d) 失敗: %v", i, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("Asset(%d) = %q,預期 %q", i, got, want)
		}
	}
}

func TestOpenRejectsBadMagic(t *testing.T) {
	raw := buildSyntheticLBX([][]byte{[]byte("x")})
	raw[2] = 0x00 // 破壞 magic
	raw[3] = 0x00
	if _, err := Open(bytes.NewReader(raw), int64(len(raw))); err == nil {
		t.Fatal("magic 錯誤時應回傳 error,卻成功")
	}
}

func TestOpenRejectsZeroAssets(t *testing.T) {
	raw := make([]byte, 12)
	binary.LittleEndian.PutUint16(raw[2:4], lbxMagic) // assetCount 仍為 0
	if _, err := Open(bytes.NewReader(raw), int64(len(raw))); err == nil {
		t.Fatal("assetCount=0 時應回傳 error,卻成功")
	}
}

func TestAssetOutOfRange(t *testing.T) {
	raw := buildSyntheticLBX([][]byte{[]byte("only")})
	a, err := Open(bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Asset(5); err == nil {
		t.Fatal("越界 id 應回傳 error,卻成功")
	}
}

// TestOpenRealLBX 在環境變數 MOO2_LBX_TEST 指向真實 .lbx 時才跑,
// 驗證解析器能吃下原版檔案(測資為版權物,不入 repo)。
func TestOpenRealLBX(t *testing.T) {
	path := os.Getenv("MOO2_LBX_TEST")
	if path == "" {
		t.Skip("未設 MOO2_LBX_TEST,跳過真實 .lbx 測試")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("開檔失敗: %v", err)
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	a, err := Open(f, fi.Size())
	if err != nil {
		t.Fatalf("解析真實 .lbx 失敗: %v", err)
	}
	t.Logf("%s: %d 個資產", path, a.Count())
	// 逐一讀出確認 offset/size 一致、不越界。
	for i := 0; i < a.Count(); i++ {
		if _, err := a.Asset(i); err != nil {
			t.Errorf("讀取資產 %d 失敗: %v", i, err)
		}
	}
}
