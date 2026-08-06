// Package smk 解碼 RAD Game Tools 的 Smacker 影片(SMK2 / SMK4)。
//
// 為什麼專案裡會有這個:MOO2 的片頭與各結局過場**不是 LBX**,是裸的 Smacker 檔,
// 只是沿用了 `.LBX` 副檔名。實測 `INTRO.LBX` 開頭四個位元組就是 `SMK2`,
// 480×160、1407 幀、rate=-7692(≈13 fps);`WININFIN.LBX` / `LOSERFIN.LBX` /
// `ORIONFIN.LBX` / `ANATKFIN.LBX` / `AMEBAFIN.LBX` / `PLNTDFIN.LBX` / `DIMTVFIN.LBX` /
// `ANWINFIN.LBX` 同理。所以「Smacker 過場」這一項的前置不是找檔案,是寫一個解碼器。
//
// 這份實作是照 Smacker 的位元流結構自己寫的,不是移植誰的程式碼。驗收方式是拿真檔解。
//
// ============ 驗收現況 ============
//
// **已驗證**(INTRO / ORIONFIN / WININFIN 三個檔都成立):
//   - 標頭、幀大小表、幀旗標表、樹區的解析:所有幀資料吃完後檔案**殘餘 0 位元組**
//     (`Trailing()`)。幀邊界只要錯一個位元組,這個數字就不會是 0。
//   - 四棵 Huffman 樹的節點數都在標頭上界 `((size+3)>>2)+4` 之內(`TreeWarnings()`)。
//   - 全片 1407 幀解得完,**沒有任何一次位元流走出樹外**。
//   - **位元預算幾乎完全吻合**:1407 幀合計只多讀 721 bits(平均每幀 0.5 bit,
//     就是最後一個 code 跨過位元組邊界的填補)。樹的形狀是固定的,消耗的位元數
//     只由「讀了哪些 code」決定——位元預算對得上,就代表 code 讀對了。
//   - 第 1050 幀(經過一千多幀差分累積)畫面依然乾淨連貫。值解錯的解碼器會愈解愈爛。
//
// ⚠ **一個曾經誤判、寫下來免得重犯**:中段幾幀(118–124,全片動作最劇烈的段落)畫面有明顯
// 的方塊感,一開始被當成解碼錯誤查了很久。那是**原始素材本身**——1996 年 480×160、
// 256 色、低位元率的 Smacker,高動態段落本來就這樣。真正的 bug 是另一件事(見下)。
//
// **修掉的 bug**:編碼器在「剩餘區塊全部與上一幀相同」時就**停筆**,不會把 SKIP 一路寫到
// 最後一格。原本的實作照著「解滿 blocks 個區塊」跑,位元流用完之後繼續從補零的位元讀,
// 於是在畫面底部畫出垃圾。改成位元流用完就停、剩餘區塊維持上一幀內容之後,
// 超讀從 **1,289,836 bits(162 幀)降到 721 bits(38 幀)**。
//
// **已測過、確定不是原因的**(負面結果,別再重試):escape 值的位元組序反過來
// (超讀 162→701 幀,更糟)、MRU 拿掉「與 last[0] 相同就跳過」的判斷(162→210,更糟)、
// 對 SMK2 也讀 SMK4 的全彩模式位元(162→517,更糟)。
//
// 診斷工具在 `cmd/smkdump`:`SMKDUMP_USAGE=1` 印每幀位元消耗、超讀起點與消耗曲線;
// `SMKDUMP_DIAG=1` 另外輸出索引灰階圖與調色盤色卡(索引圖乾淨而彩色圖髒 = 調色盤問題,
// 反之 = 影像解碼問題)。真檔驗收:
// `MOO2_SMK_TEST=<某個 .smk> scripts/test.sh ./internal/smk/`。
//
// ============ 格式重點(全部由實作驗證)============
//
//	標頭 104 位元組:
//	  +0x00 "SMK2"/"SMK4"  +0x04 寬  +0x08 高  +0x0C 幀數
//	  +0x10 rate(>0 = 每幀毫秒;<0 = 每幀 -rate/100 毫秒;0 = 10fps)
//	  +0x14 flags  +0x18 七軌音訊大小  +0x34 樹區大小
//	  +0x38 MMAP  +0x3C MCLR  +0x40 FULL  +0x44 TYPE(四棵樹解出來的大小)
//	  +0x48 七軌音訊取樣率  +0x64 保留
//	接著:幀大小表(4×幀數)、幀旗標表(1×幀數)、樹區(樹區大小 位元組)。
//
//	位元流是 **LSB-first**(每個位元組從最低位讀起)。
//
//	四棵樹都是「16 位元大樹」:先各讀一棵 8 位元小樹當低位元組/高位元組的解碼器,
//	再用它們逐葉組出 16 位元值。大樹另外帶三個 **escape 值**——葉子若命中 escape,
//	該葉存 0 並把索引記進 last[0..2];解碼時每取一個值就把 last 三格做 MRU 輪替
//	(這是 Smacker 省位元的手法:最近用過的三個值可以用極短碼再取一次)。
//
//	影像以 **4×4 區塊**逐列解碼,區塊型別由 TYPE 樹給:
//	  低 2 位 = 型別(0 單色 / 1 全彩 / 2 略過 / 3 填色),
//	  bit2..7 = 連續幾個區塊(查 blockRuns 表,末五格是 128/256/512/1024/2048),
//	  bit8.. = 填色型別的顏色索引。
package smk

import (
	"encoding/binary"
	"fmt"
)

// Header 是 Smacker 檔的標頭欄位(只留解碼與播放用得到的)。
type Header struct {
	Signature     string // "SMK2" 或 "SMK4"
	Width, Height int
	Frames        int
	FrameRateMS   int // 每幀毫秒(已由 rate 欄位換算)
	Flags         uint32

	TreesSize                              uint32
	MMapSize, MClrSize, FullSize, TypeSize uint32
}

// headerSize 是 Smacker 標頭的固定長度。
const headerSize = 104

// smkVersion4 代表 SMK4;它的全彩區塊多兩種壓縮模式(見 decodeFrame)。
const smkVersion4 = '4'

// ParseHeader 解出標頭。資料太短或簽章不對會回錯誤——這是判斷「這個檔到底是不是
// Smacker」的唯一可靠方法(副檔名在 MOO2 裡完全不能信)。
func ParseHeader(data []byte) (*Header, error) {
	if len(data) < headerSize {
		return nil, fmt.Errorf("smk: 資料太短(%d 位元組,標頭需 %d)", len(data), headerSize)
	}
	sig := string(data[0:4])
	if sig != "SMK2" && sig != "SMK4" {
		return nil, fmt.Errorf("smk: 簽章不是 SMK2/SMK4(%q)", sig)
	}
	le := binary.LittleEndian
	h := &Header{
		Signature: sig,
		Width:     int(le.Uint32(data[4:])),
		Height:    int(le.Uint32(data[8:])),
		Frames:    int(le.Uint32(data[12:])),
		Flags:     le.Uint32(data[20:]),
		TreesSize: le.Uint32(data[52:]),
		MMapSize:  le.Uint32(data[56:]),
		MClrSize:  le.Uint32(data[60:]),
		FullSize:  le.Uint32(data[64:]),
		TypeSize:  le.Uint32(data[68:]),
	}
	// rate:>0 每幀毫秒;<0 每幀 -rate/100 毫秒;0 視為 10fps(格式慣例)。
	rate := int32(le.Uint32(data[16:]))
	switch {
	case rate > 0:
		h.FrameRateMS = int(rate)
	case rate < 0:
		h.FrameRateMS = int(-rate) / 100
	default:
		h.FrameRateMS = 100
	}
	if h.FrameRateMS <= 0 {
		h.FrameRateMS = 100
	}
	if h.Width <= 0 || h.Height <= 0 || h.Frames <= 0 {
		return nil, fmt.Errorf("smk: 標頭尺寸不合理(%dx%d,%d 幀)", h.Width, h.Height, h.Frames)
	}
	return h, nil
}

// 幀旗標的位元意義(旗標表每幀一位元組)。
const (
	frameHasPalette = 1 << 0 // bit0:本幀帶調色盤記錄
	// bit1..bit7:七條音軌各一位;本解碼器略過音訊區塊(只取畫面)。
	frameKeyframe = 1 // 幀大小表的 bit0 兼作關鍵幀旗標(見 Open)
)

// Decoder 是一份已載入的 Smacker 影片。畫面以 8-bit 調色盤索引保存,
// 逐幀解碼時就地更新(Smacker 是差分格式,不能跳幀解)。
type Decoder struct {
	H *Header

	frameSizes []uint32
	frameFlags []byte
	frameData  [][]byte // 各幀的原始資料(調色盤 + 音訊 + 影像位元流)

	mmap, mclr, full, typ *bigTree

	pal    [768]byte // 目前調色盤(RGB,已由 6-bit 展開成 8-bit)
	pix    []byte    // 目前畫面(寬×高 的調色盤索引)
	cur    int       // 下一個要解的幀
	stride int

	// 診斷用:上一幀影像資料的位元組數與實際用掉的位元數(見 LastVideoUsage)。
	lastVideoBytes int
	lastVideoBits  int
	// trailing 是所有幀資料吃完之後檔案還剩幾個位元組。應該是 0——不是 0 就代表
	// 幀大小表的解讀(尤其低位旗標要不要遮)有問題,那會讓每一幀的起點都錯位。
	trailing int
	// treeWarn 記錄「樹的格數超過標頭上界」的自我檢查結果(空 = 四棵樹都在界內)。
	treeWarn []string

	// 診斷用:上一幀的區塊統計(見 LastBlockStats)。
	blkTotal    int
	blkOverAt   int    // 位元位置首次超出資料尾端時的區塊索引(-1 = 沒超出)
	blkTypeUsed [4]int // 各型別各用了幾次
	blkCurve    [][2]int
}

// LastBlockCurve 回傳上一幀的「區塊進度% vs 位元消耗%」取樣(八個點)。
func (d *Decoder) LastBlockCurve() [][2]int { return d.blkCurve }

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// LastBlockStats 回傳上一幀的區塊統計:總區塊數、超讀起點的區塊索引(-1 = 沒超讀)、
// 四種區塊型別各用了幾次。超讀起點落在尾端 = 系統性少算;落在中段 = 解碼中途漂移。
func (d *Decoder) LastBlockStats() (total, overAt int, byType [4]int) {
	return d.blkTotal, d.blkOverAt, d.blkTypeUsed
}

// TreeWarnings 回傳建樹時的自我檢查警告(空 = 正常)。
func (d *Decoder) TreeWarnings() []string { return d.treeWarn }

// Trailing 回傳檔案在所有幀資料之後的殘餘位元組數(健康的檔案應為 0)。
func (d *Decoder) Trailing() int { return d.trailing }

// LastVideoUsage 回傳上一幀「影像資料位元組數、實際消耗的位元數」。
// 兩者應該幾乎相等(消耗位元數 ≈ 位元組數×8);差很多就代表幀資料起始位移或
// 區塊數算錯——這是判斷解碼器有沒有對齊的最直接量測。
func (d *Decoder) LastVideoUsage() (bytes, bits int) { return d.lastVideoBytes, d.lastVideoBits }

// Open 解析整個 Smacker 檔並建好四棵 Huffman 樹,回傳可逐幀解碼的 Decoder。
func Open(data []byte) (*Decoder, error) {
	h, err := ParseHeader(data)
	if err != nil {
		return nil, err
	}
	need := headerSize + h.Frames*5 + int(h.TreesSize)
	if len(data) < need {
		return nil, fmt.Errorf("smk: 資料不足(需至少 %d 位元組,實得 %d)", need, len(data))
	}
	d := &Decoder{H: h, stride: h.Width}
	off := headerSize
	d.frameSizes = make([]uint32, h.Frames)
	for i := range d.frameSizes {
		d.frameSizes[i] = binary.LittleEndian.Uint32(data[off:])
		off += 4
	}
	d.frameFlags = make([]byte, h.Frames)
	copy(d.frameFlags, data[off:off+h.Frames])
	off += h.Frames

	// 樹區:四棵大樹接連放在同一條位元流上。
	br := newBitReader(data[off : off+int(h.TreesSize)])
	off += int(h.TreesSize)
	for _, t := range []struct {
		dst  **bigTree
		size uint32
		name string
	}{
		{&d.mmap, h.MMapSize, "MMAP"},
		{&d.mclr, h.MClrSize, "MCLR"},
		{&d.full, h.FullSize, "FULL"},
		{&d.typ, h.TypeSize, "TYPE"},
	} {
		bt, err := readBigTree(br, t.size)
		if err != nil {
			return nil, fmt.Errorf("smk: 解 %s 樹: %w", t.name, err)
		}
		// 標頭給的樹大小換算成「最多幾格」:((size+3)>>2)+4。實際格數超過它就代表
		// 這棵樹讀多了(位元流會從那裡開始錯位),是最直接的自我檢查。
		if bound := int((t.size+3)>>2) + 4; len(bt.vals) > bound {
			d.treeWarn = append(d.treeWarn,
				fmt.Sprintf("%s 樹 %d 格 > 上界 %d", t.name, len(bt.vals), bound))
		}
		*t.dst = bt
	}

	// 各幀資料。幀大小表的低兩位是旗標(bit0 關鍵幀、bit1 未使用),要遮掉才是長度。
	d.frameData = make([][]byte, h.Frames)
	for i := 0; i < h.Frames; i++ {
		size := int(d.frameSizes[i] &^ 3)
		if off+size > len(data) {
			return nil, fmt.Errorf("smk: 第 %d 幀超出檔案範圍(需 %d,剩 %d)", i, size, len(data)-off)
		}
		d.frameData[i] = data[off : off+size]
		off += size
	}
	d.pix = make([]byte, h.Width*h.Height)
	d.trailing = len(data) - off
	return d, nil
}

// smkPalRamp 是 Smacker 的 6-bit → 8-bit 調色盤展開表(格式固定值)。
var smkPalRamp = [64]byte{
	0x00, 0x04, 0x08, 0x0C, 0x10, 0x14, 0x18, 0x1C,
	0x20, 0x24, 0x28, 0x2C, 0x30, 0x34, 0x38, 0x3C,
	0x41, 0x45, 0x49, 0x4D, 0x51, 0x55, 0x59, 0x5D,
	0x61, 0x65, 0x69, 0x6D, 0x71, 0x75, 0x79, 0x7D,
	0x82, 0x86, 0x8A, 0x8E, 0x92, 0x96, 0x9A, 0x9E,
	0xA2, 0xA6, 0xAA, 0xAE, 0xB2, 0xB6, 0xBA, 0xBE,
	0xC3, 0xC7, 0xCB, 0xCF, 0xD3, 0xD7, 0xDB, 0xDF,
	0xE3, 0xE7, 0xEB, 0xEF, 0xF3, 0xF7, 0xFB, 0xFF,
}

// blockRuns 把 TYPE 值的 bit2..7 換成「連續幾個區塊」。前 59 格是 1..59,
// 末五格跳成 128/256/512/1024/2048——這是格式定義的,不是等差。
var blockRuns = [64]int{
	1, 2, 3, 4, 5, 6, 7, 8,
	9, 10, 11, 12, 13, 14, 15, 16,
	17, 18, 19, 20, 21, 22, 23, 24,
	25, 26, 27, 28, 29, 30, 31, 32,
	33, 34, 35, 36, 37, 38, 39, 40,
	41, 42, 43, 44, 45, 46, 47, 48,
	49, 50, 51, 52, 53, 54, 55, 56,
	57, 58, 59, 128, 256, 512, 1024, 2048,
}

// 區塊型別(TYPE 值的低 2 位)。
const (
	blkMono = 0 // 單色:兩色 + 16 位遮罩
	blkFull = 1 // 全彩:逐像素
	blkSkip = 2 // 略過:沿用上一幀
	blkFill = 3 // 填色:整塊同一色
)

// decodePalette 解一幀的調色盤記錄(就地更新 d.pal)。
// 三種記錄:bit7 = 沿用舊值 N 格、bit6 = 從舊調色盤某位移複製 N 格、其餘 = 三個新的 6-bit 分量。
func (d *Decoder) decodePalette(chunk []byte) {
	var old [768]byte
	copy(old[:], d.pal[:])
	i, p := 0, 0
	for i < 256 && p < len(chunk) {
		t := chunk[p]
		p++
		switch {
		case t&0x80 != 0: // 沿用舊值
			n := int(t&0x7F) + 1
			i += n
		case t&0x40 != 0: // 從舊調色盤某位移複製
			if p >= len(chunk) {
				return
			}
			n := int(t&0x3F) + 1
			off := int(chunk[p]) * 3
			p++
			for ; n > 0 && i < 256; n-- {
				if off+2 < len(old) {
					d.pal[i*3+0] = old[off+0]
					d.pal[i*3+1] = old[off+1]
					d.pal[i*3+2] = old[off+2]
				}
				off += 3
				i++
			}
		default: // 三個新的 6-bit 分量
			if p+1 >= len(chunk) {
				return
			}
			d.pal[i*3+0] = smkPalRamp[t&0x3F]
			d.pal[i*3+1] = smkPalRamp[chunk[p]&0x3F]
			d.pal[i*3+2] = smkPalRamp[chunk[p+1]&0x3F]
			p += 2
			i++
		}
	}
}

// DecodeNext 解下一幀。Smacker 是差分格式,必須從第 0 幀依序解,不能跳。
// 回傳目前畫面(調色盤索引,長度 寬×高)與調色盤(768 位元組 RGB)。
// 回傳的切片是內部緩衝,下一次呼叫會被覆寫;要留著請自行複製。
func (d *Decoder) DecodeNext() (pix []byte, pal []byte, err error) {
	if d.cur >= d.H.Frames {
		return nil, nil, fmt.Errorf("smk: 已無下一幀(共 %d 幀)", d.H.Frames)
	}
	data := d.frameData[d.cur]
	flags := d.frameFlags[d.cur]
	d.cur++

	off := 0
	if flags&frameHasPalette != 0 {
		if off >= len(data) {
			return nil, nil, fmt.Errorf("smk: 第 %d 幀調色盤記錄超出範圍", d.cur-1)
		}
		size := int(data[off]) * 4
		if size < 1 || off+size > len(data) {
			return nil, nil, fmt.Errorf("smk: 第 %d 幀調色盤長度不合理(%d)", d.cur-1, size)
		}
		d.decodePalette(data[off+1 : off+size])
		off += size
	}
	// 七條音軌:每條前四個位元組是解壓後大小,接著是區塊本體。本解碼器只取畫面,略過。
	af := flags >> 1
	for i := 0; i < 7; i++ {
		if af&1 != 0 {
			if off+4 > len(data) {
				return nil, nil, fmt.Errorf("smk: 第 %d 幀音軌 %d 超出範圍", d.cur-1, i)
			}
			size := int(binary.LittleEndian.Uint32(data[off:]))
			if size < 4 || off+size > len(data) {
				return nil, nil, fmt.Errorf("smk: 第 %d 幀音軌 %d 長度不合理(%d)", d.cur-1, i, size)
			}
			off += size
		}
		af >>= 1
	}
	d.lastVideoBytes = len(data) - off
	if err := d.decodeVideo(data[off:]); err != nil {
		return nil, nil, fmt.Errorf("smk: 第 %d 幀影像: %w", d.cur-1, err)
	}
	return d.pix, d.pal[:], nil
}

// decodeVideo 解一幀的影像位元流(4×4 區塊逐列)。
func (d *Decoder) decodeVideo(data []byte) error {
	br := newBitReader(data)
	defer func() { d.lastVideoBits = br.pos }()
	bw := (d.H.Width + 3) / 4
	bh := (d.H.Height + 3) / 4
	blocks := bw * bh
	v4 := d.H.Signature[3] == smkVersion4

	d.blkTotal, d.blkOverAt, d.blkTypeUsed = blocks, -1, [4]int{}
	limit := len(data) * 8

	d.blkCurve = d.blkCurve[:0]
	nextMark := blocks / 8

	blk := 0
	for blk < blocks {
		if br.pos > limit {
			// 位元流用完了:剩下的區塊維持上一幀的內容。
			// 編碼器在「剩餘區塊全部不變」時就停筆,不會把 SKIP 一路寫到最後一格;
			// 繼續解下去只會從補零的位元讀出垃圾,畫在畫面底部。
			if d.blkOverAt < 0 {
				d.blkOverAt = blk
			}
			break
		}
		for blocks > 0 && blk >= nextMark && len(d.blkCurve) < 8 {
			// 記「解到 X% 的區塊時,位元用掉幾 %」。兩個百分比同步 = 均勻;
			// 位元跑在前面 = 每個 code 都多吃一點;末段才暴衝 = 收尾有問題。
			d.blkCurve = append(d.blkCurve, [2]int{blk * 100 / blocks, br.pos * 100 / max(limit, 1)})
			nextMark += blocks / 8
		}
		t, err := d.typ.get(br)
		if err != nil {
			return err
		}
		d.blkTypeUsed[t&3]++
		run := blockRuns[(t>>2)&0x3F]
		switch t & 3 {
		case blkMono:
			for ; run > 0 && blk < blocks; run-- {
				clr, err := d.mclr.get(br)
				if err != nil {
					return err
				}
				m, err := d.mmap.get(br)
				if err != nil {
					return err
				}
				hi, lo := byte(clr>>8), byte(clr&0xFF)
				d.eachPixel(blk, bw, func(row, col int) byte {
					bit := m >> (uint(row)*4 + uint(col)) & 1
					if bit != 0 {
						return hi
					}
					return lo
				})
				blk++
			}
		case blkFull:
			// SMK4 的全彩區塊多兩種壓縮模式:1 = 每列一個 16 位值攤成 4 像素、
			// 2 = 每兩列共用同一對值。SMK2 只有模式 0。
			mode := 0
			if v4 {
				if br.bit() != 0 {
					mode = 1
				} else if br.bit() != 0 {
					mode = 2
				}
			}
			for ; run > 0 && blk < blocks; run-- {
				if err := d.fullBlock(br, blk, bw, mode); err != nil {
					return err
				}
				blk++
			}
		case blkSkip:
			for ; run > 0 && blk < blocks; run-- {
				blk++ // 沿用上一幀,什麼都不寫
			}
		case blkFill:
			col := byte(t >> 8)
			for ; run > 0 && blk < blocks; run-- {
				d.eachPixel(blk, bw, func(int, int) byte { return col })
				blk++
			}
		}
	}
	return nil
}

// eachPixel 把第 blk 個 4×4 區塊的每個像素設成 f(row, col) 的結果。
// 超出畫面邊界的像素直接跳過(寬高不是 4 的倍數時會發生)。
func (d *Decoder) eachPixel(blk, bw int, f func(row, col int) byte) {
	x0 := (blk % bw) * 4
	y0 := (blk / bw) * 4
	for r := 0; r < 4; r++ {
		y := y0 + r
		if y >= d.H.Height {
			break
		}
		base := y*d.stride + x0
		for c := 0; c < 4; c++ {
			if x0+c >= d.H.Width {
				break
			}
			d.pix[base+c] = f(r, c)
		}
	}
}

// fullBlock 解一個全彩區塊。
func (d *Decoder) fullBlock(br *bitReader, blk, bw, mode int) error {
	var px [16]byte
	switch mode {
	case 0: // 逐像素:每列兩個 16 位值,先右半再左半
		for r := 0; r < 4; r++ {
			right, err := d.full.get(br)
			if err != nil {
				return err
			}
			left, err := d.full.get(br)
			if err != nil {
				return err
			}
			px[r*4+0] = byte(left & 0xFF)
			px[r*4+1] = byte(left >> 8)
			px[r*4+2] = byte(right & 0xFF)
			px[r*4+3] = byte(right >> 8)
		}
	case 1: // 每兩列一個值:低位元組給左半、高位元組給右半
		for r := 0; r < 4; r += 2 {
			v, err := d.full.get(br)
			if err != nil {
				return err
			}
			lo, hi := byte(v&0xFF), byte(v>>8)
			for k := 0; k < 2; k++ {
				px[(r+k)*4+0], px[(r+k)*4+1] = lo, lo
				px[(r+k)*4+2], px[(r+k)*4+3] = hi, hi
			}
		}
	case 2: // 每兩列共用同一對 16 位值
		for r := 0; r < 4; r += 2 {
			right, err := d.full.get(br)
			if err != nil {
				return err
			}
			left, err := d.full.get(br)
			if err != nil {
				return err
			}
			for k := 0; k < 2; k++ {
				px[(r+k)*4+0] = byte(left & 0xFF)
				px[(r+k)*4+1] = byte(left >> 8)
				px[(r+k)*4+2] = byte(right & 0xFF)
				px[(r+k)*4+3] = byte(right >> 8)
			}
		}
	}
	d.eachPixel(blk, bw, func(row, col int) byte { return px[row*4+col] })
	return nil
}

// Reset 把解碼器倒回第 0 幀(畫面與調色盤一併清空)。
func (d *Decoder) Reset() {
	d.cur = 0
	for i := range d.pix {
		d.pix[i] = 0
	}
	for i := range d.pal {
		d.pal[i] = 0
	}
}

// Position 回傳下一個要解的幀序號。
func (d *Decoder) Position() int { return d.cur }
