package lbx

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

// buildImageAsset 組出一個影像資產:header + (1 幀的) offset 表 + 選用內嵌調色盤 + frame 資料。
func buildImageAsset(w, h, frameTime int, flags uint16, palBlock, frameData []byte) []byte {
	var buf bytes.Buffer
	le := binary.LittleEndian
	wr16 := func(v uint16) { binary.Write(&buf, le, v) }
	wr32 := func(v uint32) { binary.Write(&buf, le, v) }

	wr16(uint16(w))
	wr16(uint16(h))
	wr16(0) // 未知
	wr16(1) // frameCount = 1
	wr16(uint16(frameTime))
	wr16(flags)

	// offset 表:off0 = header(12) + 表(2*4) + palBlock;off1 = off0 + len(frameData)。
	off0 := 12 + 8 + len(palBlock)
	wr32(uint32(off0))
	wr32(uint32(off0 + len(frameData)))
	buf.Write(palBlock)
	buf.Write(frameData)
	return buf.Bytes()
}

func TestDecodeImageNoCompress(t *testing.T) {
	// 4×2,未壓縮,index = 0..7。
	frame := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	raw := buildImageAsset(4, 2, 0, flagNoCompress, nil, frame)

	im, err := DecodeImage(raw)
	if err != nil {
		t.Fatalf("DecodeImage 失敗: %v", err)
	}
	if im.Width != 4 || im.Height != 2 || len(im.Frames) != 1 {
		t.Fatalf("header 錯:w=%d h=%d frames=%d", im.Width, im.Height, len(im.Frames))
	}
	fr := im.Frames[0]
	for i, want := range frame {
		if fr.Index[i] != want || !fr.Written[i] {
			t.Errorf("像素 %d:index=%d written=%v,預期 index=%d written=true", i, fr.Index[i], fr.Written[i], want)
		}
	}
}

func TestDecodeImageRLE(t *testing.T) {
	// 4×2 壓縮幀,手工編碼:
	// row0: skip1 + 寫 [0xAA,0xBB] 於 x=1,2;row1: 寫 [0xCC] 於 x=0。
	frame := []byte{
		0x01, 0x00, 0x00, 0x00, // 起始標記 size=1, y=0
		0x02, 0x00, 0x01, 0x00, 0xAA, 0xBB, // row0: size=2 skip=1 pixels
		0x00, 0x00, 0x01, 0x00, // row0 終止:size=0 skip=1 → y+=1
		0x01, 0x00, 0x00, 0x00, 0xCC, 0x00, // row1: size=1 skip=0 pixel + 奇數補 1
		0x00, 0x00, 0x01, 0x00, // row1 終止:y+=1 → y=2 結束
	}
	raw := buildImageAsset(4, 2, 0, 0, nil, frame)

	im, err := DecodeImage(raw)
	if err != nil {
		t.Fatalf("DecodeImage 失敗: %v", err)
	}
	fr := im.Frames[0]

	type px struct {
		idx     uint8
		written bool
	}
	want := []px{
		{0, false}, {0xAA, true}, {0xBB, true}, {0, false}, // row0
		{0xCC, true}, {0, false}, {0, false}, {0, false}, // row1
	}
	for i, wp := range want {
		if fr.Index[i] != wp.idx || fr.Written[i] != wp.written {
			t.Errorf("像素 %d:index=%#x written=%v,預期 index=%#x written=%v",
				i, fr.Index[i], fr.Written[i], wp.idx, wp.written)
		}
	}
}

func TestDecodeImageEmbeddedPalette(t *testing.T) {
	// FLAG_PALETTE:palStart=1, palCount=2,兩色(6-bit)。
	palBlock := []byte{
		0x01, 0x00, 0x02, 0x00, // palStart=1, palCount=2
		0x00, 0x3f, 0x00, 0x00, // 色1:R=0x3f(63)
		0x00, 0x00, 0x20, 0x10, // 色2:G=0x20 B=0x10
	}
	frame := []byte{1, 2, 1, 2} // 2×2 未壓縮
	raw := buildImageAsset(2, 2, 0, flagNoCompress|flagPalette, palBlock, frame)

	im, err := DecodeImage(raw)
	if err != nil {
		t.Fatalf("DecodeImage 失敗: %v", err)
	}
	if im.Embedded == nil || im.PalStart != 1 || im.PalCount != 2 {
		t.Fatalf("調色盤 meta 錯:embedded=%v start=%d count=%d", im.Embedded != nil, im.PalStart, im.PalCount)
	}
	c1 := im.Embedded[1]
	if c1.R != 0x3f<<2 || c1.G != 0 || c1.B != 0 || c1.A != 0xff {
		t.Errorf("色1 = %+v,預期 R=%d", c1, uint8(0x3f<<2))
	}
	c2 := im.Embedded[2]
	if c2.G != 0x20<<2 || c2.B != 0x10<<2 {
		t.Errorf("色2 = %+v,預期 G=%d B=%d", c2, uint8(0x20<<2), uint8(0x10<<2))
	}

	// 上色:keyColor=false,全部像素應不透明。
	img := im.Frames[0].ToRGBA(im.Embedded, false)
	if img.Bounds().Dx() != 2 || img.Bounds().Dy() != 2 {
		t.Fatalf("RGBA 尺寸錯:%v", img.Bounds())
	}
	if img.Pix[3] != 0xff {
		t.Errorf("像素0 alpha = %d,預期 255", img.Pix[3])
	}
}

func TestToRGBAKeyColorTransparent(t *testing.T) {
	frame := []byte{0, 1, 0, 1}
	raw := buildImageAsset(2, 2, 0, flagNoCompress|flagKeyColor|flagPalette, []byte{
		0x00, 0x00, 0x02, 0x00, // palStart=0 count=2
		0x00, 0x10, 0x10, 0x10, // 色0
		0x00, 0x3f, 0x3f, 0x3f, // 色1
	}, frame)
	im, err := DecodeImage(raw)
	if err != nil {
		t.Fatal(err)
	}
	img := im.Frames[0].ToRGBA(im.Embedded, im.KeyColor())
	// index 0 + keycolor → 透明(alpha 0);index 1 → 不透明。
	if img.Pix[3] != 0 {
		t.Errorf("index0 keycolor 應透明,alpha=%d", img.Pix[3])
	}
	if img.Pix[4+3] != 0xff {
		t.Errorf("index1 應不透明,alpha=%d", img.Pix[4+3])
	}
}

// TestDecodeRealImages 在 MOO2_LBX_TEST 指向真實影像 .lbx 時,嘗試解碼每個資產,
// 統計成功數並驗證成功者的幀資料長度一致(非影像資產解碼失敗屬正常,不算錯)。
func TestDecodeRealImages(t *testing.T) {
	path := os.Getenv("MOO2_LBX_TEST")
	if path == "" {
		t.Skip("未設 MOO2_LBX_TEST")
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	fi, _ := f.Stat()
	a, err := Open(f, fi.Size())
	if err != nil {
		t.Fatal(err)
	}
	ok := 0
	for i := 0; i < a.Count(); i++ {
		raw, err := a.Asset(i)
		if err != nil {
			t.Fatalf("讀資產 %d: %v", i, err)
		}
		im, err := DecodeImage(raw)
		if err != nil {
			continue // 非影像資產
		}
		for fi, fr := range im.Frames {
			if len(fr.Index) != im.Width*im.Height {
				t.Errorf("資產 %d 幀 %d 尺寸不一致", i, fi)
			}
		}
		ok++
	}
	t.Logf("%s:%d/%d 個資產解碼為影像", path, ok, a.Count())
}
