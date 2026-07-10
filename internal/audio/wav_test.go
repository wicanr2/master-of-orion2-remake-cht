package audio

// wav_test.go:純函式測試,不碰音訊裝置。合成最小 PCM WAV(自組 RIFF/WAVE/fmt/data
// header)驗證 DecodeWAV/ExtractWAV,不依賴任何版權測資。

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// buildWAV 依 WAV 容器格式手工組出一個最小合法 PCM WAV(RIFF/WAVE/fmt /data)。
func buildWAV(sampleRate uint32, channels, bits uint16, data []byte) []byte {
	var fmtChunk bytes.Buffer
	binary.Write(&fmtChunk, binary.LittleEndian, uint16(1)) // audioFormat = PCM
	binary.Write(&fmtChunk, binary.LittleEndian, channels)
	binary.Write(&fmtChunk, binary.LittleEndian, sampleRate)
	byteRate := sampleRate * uint32(channels) * uint32(bits) / 8
	binary.Write(&fmtChunk, binary.LittleEndian, byteRate)
	blockAlign := channels * bits / 8
	binary.Write(&fmtChunk, binary.LittleEndian, blockAlign)
	binary.Write(&fmtChunk, binary.LittleEndian, bits)

	var buf bytes.Buffer
	buf.WriteString("RIFF")
	sizePos := buf.Len()
	binary.Write(&buf, binary.LittleEndian, uint32(0)) // 佔位,稍後回填
	buf.WriteString("WAVE")

	buf.WriteString("fmt ")
	binary.Write(&buf, binary.LittleEndian, uint32(fmtChunk.Len()))
	buf.Write(fmtChunk.Bytes())

	buf.WriteString("data")
	binary.Write(&buf, binary.LittleEndian, uint32(len(data)))
	buf.Write(data)
	if len(data)%2 == 1 {
		buf.WriteByte(0) // chunk 偶數對齊
	}

	raw := buf.Bytes()
	size := uint32(len(raw) - 8) // RIFF size 欄:其後(含 WAVE)位元組數
	binary.LittleEndian.PutUint32(raw[sizePos:sizePos+4], size)
	return raw
}

// frame16 讀出 PCM 中第 i 個 frame 的 (左, 右) 16-bit 有號樣本。
func frame16(t *testing.T, pcm []byte, i int) (int16, int16) {
	t.Helper()
	off := i * 4
	if off+4 > len(pcm) {
		t.Fatalf("frame %d 超出 PCM 長度 %d", i, len(pcm))
	}
	l := int16(binary.LittleEndian.Uint16(pcm[off : off+2]))
	r := int16(binary.LittleEndian.Uint16(pcm[off+2 : off+4]))
	return l, r
}

func TestDecodeWAV(t *testing.T) {
	cases := []struct {
		name       string
		sampleRate uint32
		channels   uint16
		bits       uint16
		data       []byte
		wantFrames [][2]int16 // 期望的 (左,右) 16-bit 樣本序列
	}{
		{
			name:       "8bit_mono",
			sampleRate: 22050,
			channels:   1,
			bits:       8,
			data:       []byte{0, 128, 255},
			wantFrames: [][2]int16{
				{-32768, -32768}, // 0   -> (0-128)<<8   = -32768
				{0, 0},           // 128 -> (128-128)<<8 = 0
				{32512, 32512},   // 255 -> (255-128)<<8 = 32512
			},
		},
		{
			name:       "8bit_stereo",
			sampleRate: 11025,
			channels:   2,
			bits:       8,
			data:       []byte{0, 255, 128, 128},
			wantFrames: [][2]int16{
				{-32768, 32512},
				{0, 0},
			},
		},
		{
			name:       "16bit_stereo",
			sampleRate: 22050,
			channels:   2,
			bits:       16,
			data: func() []byte {
				var b bytes.Buffer
				binary.Write(&b, binary.LittleEndian, int16(1000))
				binary.Write(&b, binary.LittleEndian, int16(-1000))
				binary.Write(&b, binary.LittleEndian, int16(32767))
				binary.Write(&b, binary.LittleEndian, int16(-32768))
				return b.Bytes()
			}(),
			wantFrames: [][2]int16{
				{1000, -1000},
				{32767, -32768},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wav := buildWAV(tc.sampleRate, tc.channels, tc.bits, tc.data)
			clip, err := DecodeWAV(wav)
			if err != nil {
				t.Fatalf("DecodeWAV 失敗: %v", err)
			}
			if clip.SampleRate != int(tc.sampleRate) {
				t.Errorf("SampleRate = %d, want %d", clip.SampleRate, tc.sampleRate)
			}
			wantLen := len(tc.wantFrames) * 4
			if len(clip.PCM) != wantLen {
				t.Fatalf("PCM 長度 = %d, want %d", len(clip.PCM), wantLen)
			}
			for i, want := range tc.wantFrames {
				l, r := frame16(t, clip.PCM, i)
				if l != want[0] || r != want[1] {
					t.Errorf("frame %d = (%d,%d), want (%d,%d)", i, l, r, want[0], want[1])
				}
			}
		})
	}
}

func TestDecodeWAV_NotRIFF(t *testing.T) {
	if _, err := DecodeWAV([]byte("not a wav file at all")); err == nil {
		t.Fatal("非 RIFF/WAVE 應回傳 error,卻成功")
	}
}

func TestExtractWAV(t *testing.T) {
	wav := buildWAV(22050, 2, 8, []byte{10, 20, 30, 40, 50, 60})

	t.Run("with junk prefix", func(t *testing.T) {
		junk := []byte{0xDE, 0xAD, 0xBE, 0xEF, 0x01, 0x02, 0x03}
		asset := append(append([]byte{}, junk...), wav...)
		got, ok := ExtractWAV(asset)
		if !ok {
			t.Fatal("ExtractWAV 應成功,卻回傳 ok=false")
		}
		if !bytes.Equal(got, wav) {
			t.Errorf("ExtractWAV 切出的片段與原始 WAV 不符\ngot:  % x\nwant: % x", got, wav)
		}
	})

	t.Run("with junk prefix and trailing bytes", func(t *testing.T) {
		junk := []byte{0x00, 0x00, 0x00}
		trailing := []byte{0xFF, 0xFF, 0xFF, 0xFF}
		asset := append(append(append([]byte{}, junk...), wav...), trailing...)
		got, ok := ExtractWAV(asset)
		if !ok {
			t.Fatal("ExtractWAV 應成功,卻回傳 ok=false")
		}
		if !bytes.Equal(got, wav) {
			t.Errorf("ExtractWAV 未依 RIFF size 精確切出片段(混入了 trailing bytes)\ngot:  % x\nwant: % x", got, wav)
		}
	})

	t.Run("no RIFF", func(t *testing.T) {
		asset := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		_, ok := ExtractWAV(asset)
		if ok {
			t.Fatal("無 RIFF 的 bytes 應回傳 ok=false,卻成功")
		}
	})
}
