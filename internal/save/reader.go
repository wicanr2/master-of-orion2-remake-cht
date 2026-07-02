// Package save 讀取原版 Master of Orion II 的存檔(save?.gam)。
//
// 存檔格式對照 openorion2 的 src/gamestate.cpp(GPL v2, Martin Doucha)逐欄位移植。
// 存檔非純序列:GameConfig 從 0 起,galaxy 存在檔尾(offset 0x31be4),
// colonyCount 在 offset 0x25b,其後 colonies→planets→stars→leaders→players→ships 接續。
package save

import (
	"encoding/binary"
	"fmt"
)

// reader 是存檔位元組上的游標,提供小端讀取與 seek,對應 openorion2 的 SeekableReadStream。
type reader struct {
	data []byte
	pos  int
}

func newReader(data []byte) *reader { return &reader{data: data} }

func (r *reader) seek(off int) error {
	if off < 0 || off > len(r.data) {
		return fmt.Errorf("save: seek 越界(%d / %d)", off, len(r.data))
	}
	r.pos = off
	return nil
}

func (r *reader) remaining() int { return len(r.data) - r.pos }

func (r *reader) need(n int) error {
	if r.pos+n > len(r.data) {
		return fmt.Errorf("save: 讀取越界(pos=%d need=%d len=%d)", r.pos, n, len(r.data))
	}
	return nil
}

func (r *reader) u8() uint8 {
	v := r.data[r.pos]
	r.pos++
	return v
}

func (r *reader) u16() uint16 {
	v := binary.LittleEndian.Uint16(r.data[r.pos : r.pos+2])
	r.pos += 2
	return v
}

func (r *reader) u32() uint32 {
	v := binary.LittleEndian.Uint32(r.data[r.pos : r.pos+4])
	r.pos += 4
	return v
}

func (r *reader) i8() int8   { return int8(r.u8()) }
func (r *reader) i16() int16 { return int16(r.u16()) }
func (r *reader) i32() int32 { return int32(r.u32()) }

// pos 回傳目前游標位置。
func (r *reader) at() int { return r.pos }

// bytesN 讀 n 個位元組(複製一份)。
func (r *reader) bytesN(n int) []byte {
	b := make([]byte, n)
	copy(b, r.data[r.pos:r.pos+n])
	r.pos += n
	return b
}

// skip 跳過 n 個位元組。
func (r *reader) skip(n int) { r.pos += n }

// cstr 讀固定長度欄位並轉成字串(截到第一個 NUL)。
func (r *reader) cstr(n int) string {
	b := r.bytesN(n)
	for i, c := range b {
		if c == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}
