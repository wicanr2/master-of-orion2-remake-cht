// Package audio 從原版 MOO2 的 .lbx 取出 PCM 音訊(音樂/音效)並交給 ebiten 播放。
//
// 格式事實見 docs/tech/audio-format.md:MOO2 沒有 MIDI 音樂,全部音樂與音效
// 都是未壓縮 PCM WAV(22050 Hz、8-bit),直接存在 LBX entry 內。因此不需任何
// 合成器——原封播放原版 PCM 即與原版 bit-identical。
//
// 本檔為純解碼邏輯(不觸碰音訊裝置),可 headless 單元測試。ebiten 裝置層在 mixer.go。
package audio

import (
	"encoding/binary"
	"fmt"
)

// Clip 是解碼後、可直接餵給 ebiten 的音訊:16-bit little-endian、雙聲道交錯。
// 統一轉成 16-bit stereo 是為了對齊 ebiten audio player 的輸入慣例(避免依賴
// ebiten 內建 wav 解碼器對 8-bit 的支援差異)。
type Clip struct {
	PCM        []byte // 16-bit LE,2 聲道交錯(len 必為 4 的倍數)
	SampleRate int
}

// riffMagic / waveMagic 為 RIFF/WAVE 容器識別。
var riffMagic = []byte("RIFF")
var waveMagic = []byte("WAVE")

// ExtractWAV 在一個 LBX entry 的原始 bytes 中找出完整的 WAV(RIFF/WAVE)片段。
//
// STREAM/STREAMHD 的 WAV 乾淨地以 RIFF 開頭;為穩健起見仍掃描 RIFF 位置,
// 並依 RIFF chunk size 切出精確長度。找不到(彩蛋槽/空槽)回傳 ok=false。
func ExtractWAV(asset []byte) (wav []byte, ok bool) {
	pos := indexOf(asset, riffMagic)
	if pos < 0 || pos+12 > len(asset) {
		return nil, false
	}
	// RIFF header:'RIFF' u32(size) 'WAVE';size 為其後(含 WAVE)位元組數。
	if !equal(asset[pos+8:pos+12], waveMagic) {
		return nil, false
	}
	size := binary.LittleEndian.Uint32(asset[pos+4 : pos+8])
	total := 8 + int(size) // 'RIFF'(4) + size 欄(4) 不計入 size,其餘計入
	if total <= 0 || pos+total > len(asset) {
		total = len(asset) - pos // size 欄不可信時退回到 entry 尾
	}
	return asset[pos : pos+total], true
}

// DecodeWAV 解析 PCM WAV 並轉成 16-bit LE 雙聲道 Clip。
// 只支援 MOO2 實際使用的 PCM(audioFormat=1)、8/16-bit、單/雙聲道。
func DecodeWAV(wav []byte) (*Clip, error) {
	if len(wav) < 12 || !equal(wav[0:4], riffMagic) || !equal(wav[8:12], waveMagic) {
		return nil, fmt.Errorf("audio: 非 RIFF/WAVE")
	}

	var (
		haveFmt              bool
		audioFormat          uint16
		channels             uint16
		sampleRate           uint32
		bits                 uint16
		data                 []byte
	)

	// 逐 sub-chunk 掃描:每個 chunk = 4-byte id + u32 size + payload(size,奇數補齊)。
	off := 12
	for off+8 <= len(wav) {
		id := wav[off : off+4]
		sz := int(binary.LittleEndian.Uint32(wav[off+4 : off+8]))
		body := off + 8
		if sz < 0 || body+sz > len(wav) {
			sz = len(wav) - body // 容忍尾端截斷
		}
		switch {
		case equal(id, []byte("fmt ")):
			if sz >= 16 {
				audioFormat = binary.LittleEndian.Uint16(wav[body : body+2])
				channels = binary.LittleEndian.Uint16(wav[body+2 : body+4])
				sampleRate = binary.LittleEndian.Uint32(wav[body+4 : body+8])
				bits = binary.LittleEndian.Uint16(wav[body+14 : body+16])
				haveFmt = true
			}
		case equal(id, []byte("data")):
			data = wav[body : body+sz]
		}
		off = body + sz
		if sz%2 == 1 {
			off++ // chunk 以偶數位元組對齊
		}
	}

	if !haveFmt {
		return nil, fmt.Errorf("audio: 缺 fmt chunk")
	}
	if audioFormat != 1 {
		return nil, fmt.Errorf("audio: 非 PCM(audioFormat=%d)", audioFormat)
	}
	if channels != 1 && channels != 2 {
		return nil, fmt.Errorf("audio: 不支援 %d 聲道", channels)
	}
	if bits != 8 && bits != 16 {
		return nil, fmt.Errorf("audio: 不支援 %d-bit", bits)
	}

	pcm := toStereo16(data, int(channels), int(bits))
	return &Clip{PCM: pcm, SampleRate: int(sampleRate)}, nil
}

// toStereo16 把來源 PCM 轉成 16-bit LE 雙聲道交錯。
//   - 8-bit 為無號(0..255,中點 128),轉為有號 16-bit:(v-128)<<8。
//   - 16-bit 已為有號 LE,原樣使用。
//   - 單聲道複製為左右兩聲道。
func toStereo16(src []byte, channels, bits int) []byte {
	// 先取出每聲道的 16-bit 樣本序列。
	var frames [][2]int16 // [左, 右]
	switch bits {
	case 8:
		if channels == 1 {
			for _, b := range src {
				v := int16(int(b)-128) << 8
				frames = append(frames, [2]int16{v, v})
			}
		} else { // stereo
			for i := 0; i+1 < len(src); i += 2 {
				l := int16(int(src[i])-128) << 8
				r := int16(int(src[i+1])-128) << 8
				frames = append(frames, [2]int16{l, r})
			}
		}
	case 16:
		if channels == 1 {
			for i := 0; i+1 < len(src); i += 2 {
				v := int16(binary.LittleEndian.Uint16(src[i : i+2]))
				frames = append(frames, [2]int16{v, v})
			}
		} else {
			for i := 0; i+3 < len(src); i += 4 {
				l := int16(binary.LittleEndian.Uint16(src[i : i+2]))
				r := int16(binary.LittleEndian.Uint16(src[i+2 : i+4]))
				frames = append(frames, [2]int16{l, r})
			}
		}
	}

	out := make([]byte, len(frames)*4)
	for i, f := range frames {
		binary.LittleEndian.PutUint16(out[i*4:i*4+2], uint16(f[0]))
		binary.LittleEndian.PutUint16(out[i*4+2:i*4+4], uint16(f[1]))
	}
	return out
}

// indexOf 回傳 needle 在 haystack 中第一次出現的位置,無則 -1。
func indexOf(haystack, needle []byte) int {
	n := len(needle)
	if n == 0 || n > len(haystack) {
		return -1
	}
	for i := 0; i+n <= len(haystack); i++ {
		if equal(haystack[i:i+n], needle) {
			return i
		}
	}
	return -1
}

func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
