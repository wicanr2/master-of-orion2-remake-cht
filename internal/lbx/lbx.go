// Package lbx 讀取原版 Master of Orion II 的 .lbx 資產封存檔。
//
// 格式對照 openorion2 的 src/lbx.cpp:169-223(GPL v2, Martin Doucha)逐位元組移植:
//
//	offset  型別    內容
//	0       u16 LE  assetCount(資產數,不得為 0)
//	2       u16 LE  magic,必為 0xfead
//	4       u32 LE  版本/旗標(openorion2 讀後丟棄)
//	8       u32 LE  第 0 個資產的起始 offset
//	12..    u32 LE  之後共有 assetCount 個 offset;
//	                第 i 個資產 = [offset[i], offset[i+1]),size 由相鄰相減得出。
//
// 即 header 從 byte 8 起有 (assetCount+1) 個 u32 offset。
package lbx

import (
	"encoding/binary"
	"fmt"
	"io"
)

// lbxMagic 是 .lbx 檔第 2-3 位元組的識別碼。
const lbxMagic = 0xfead

// Entry 描述封存檔內單一資產在檔案中的位置。
type Entry struct {
	Offset uint32
	Size   uint32
}

// Archive 是一個已解析的 .lbx 封存檔。資產以 ReaderAt 惰性讀取,
// 不在開檔時把整包載進記憶體。
type Archive struct {
	r     io.ReaderAt
	index []Entry
}

// Open 解析 r 的 header 並回傳可用的 Archive。r 必須支援隨機存取
// (通常是 *os.File 或 *bytes.Reader)。size 是整個 .lbx 的位元組數,
// 用來驗證最後一個資產的 offset 不超出檔案。
func Open(r io.ReaderAt, size int64) (*Archive, error) {
	var head [12]byte
	if _, err := r.ReadAt(head[:], 0); err != nil {
		return nil, fmt.Errorf("lbx: 讀取 header 失敗: %w", err)
	}

	assetCount := binary.LittleEndian.Uint16(head[0:2])
	magic := binary.LittleEndian.Uint16(head[2:4])
	if assetCount == 0 || magic != lbxMagic {
		return nil, fmt.Errorf("lbx: 非合法 .lbx(assetCount=%d, magic=%#x)", assetCount, magic)
	}
	// head[4:8] 是版本/旗標,略過。
	firstOffset := binary.LittleEndian.Uint32(head[8:12])

	// 讀取剩餘的 assetCount 個 offset(header 從 byte 8 起共 assetCount+1 個 u32)。
	tableBytes := make([]byte, 4*int(assetCount))
	if _, err := r.ReadAt(tableBytes, 12); err != nil {
		return nil, fmt.Errorf("lbx: 讀取 offset 表失敗: %w", err)
	}

	index := make([]Entry, assetCount)
	cur := firstOffset
	for i := 0; i < int(assetCount); i++ {
		next := binary.LittleEndian.Uint32(tableBytes[4*i : 4*i+4])
		if next <= cur {
			return nil, fmt.Errorf("lbx: 資產 %d 的 offset 非遞增(%d <= %d)", i, next, cur)
		}
		index[i] = Entry{Offset: cur, Size: next - cur}
		cur = next
	}
	if size >= 0 && int64(cur) > size {
		return nil, fmt.Errorf("lbx: 資產表超出檔案大小(結尾 offset %d > 檔案 %d)", cur, size)
	}

	return &Archive{r: r, index: index}, nil
}

// Count 回傳封存檔內的資產數。
func (a *Archive) Count() int {
	return len(a.index)
}

// Entry 回傳第 id 個資產的位置資訊。
func (a *Archive) Entry(id int) (Entry, error) {
	if id < 0 || id >= len(a.index) {
		return Entry{}, fmt.Errorf("lbx: 資產 id %d 超出範圍(共 %d 個)", id, len(a.index))
	}
	return a.index[id], nil
}

// Asset 讀出第 id 個資產的原始位元組。
func (a *Archive) Asset(id int) ([]byte, error) {
	e, err := a.Entry(id)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, e.Size)
	if _, err := a.r.ReadAt(buf, int64(e.Offset)); err != nil {
		return nil, fmt.Errorf("lbx: 讀取資產 %d 失敗: %w", id, err)
	}
	return buf, nil
}
