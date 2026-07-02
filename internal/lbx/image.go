package lbx

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
)

// 影像 header 旗標,對照 openorion2 gfx.cpp:27-31。
const (
	flagJunction   = 0x2000
	flagPalette    = 0x1000 // 內嵌調色盤
	flagKeyColor   = 0x0800 // index 0 視為透明
	flagFillBg     = 0x0400
	flagNoCompress = 0x0100 // 未壓縮,逐位元組即 palette index
)

// Palette 是 256 色的調色盤。原版每色以 6-bit VGA 儲存,載入時 <<2 放大成 8-bit
// (對照 openorion2 loadPalette,gfx.cpp:290-305)。
type Palette [256]color.RGBA

// Frame 是一張已解碼、但**尚未套用調色盤**的影像。Index 是每像素的 palette 索引,
// Written 標記該像素在 RLE 中是否被寫入(未寫入 = 透明,與 keycolor 無關)。
// 解碼與上色分離,方便同一幀套不同調色盤(原版的 palette variant)。
type Frame struct {
	W, H    int
	Index   []uint8 // len == W*H
	Written []bool  // len == W*H
}

// Image 是一個 .lbx 影像資產(可含多幀動畫)。
type Image struct {
	Width, Height int
	FrameTime     int
	Flags         uint16
	Frames        []*Frame
	// Embedded 為內嵌調色盤(若 FLAG_PALETTE);僅填入 [PalStart, PalStart+PalCount)
	// 區段,其餘需由 base palette 補齊。無內嵌時為 nil。
	Embedded *Palette
	PalStart int
	PalCount int
}

// KeyColor 回傳此影像是否把 index 0 當透明色。
func (im *Image) KeyColor() bool { return im.Flags&flagKeyColor != 0 }

// DecodeImage 解析一個影像資產的原始位元組。對照 openorion2 Image::load /
// decodeFrame(gfx.cpp:326-531)。回傳的 Frame 尚未上色,用 Frame.ToRGBA 套調色盤。
func DecodeImage(data []byte) (*Image, error) {
	if len(data) < 12 {
		return nil, fmt.Errorf("lbx image: 資料太短(%d bytes)", len(data))
	}
	le := binary.LittleEndian
	width := int(le.Uint16(data[0:2]))
	height := int(le.Uint16(data[2:4]))
	// data[4:6] 未知,略過。
	frameCount := int(le.Uint16(data[6:8]))
	frameTime := int(le.Uint16(data[8:10]))
	flags := le.Uint16(data[10:12])

	if width == 0 || height == 0 || frameCount == 0 {
		return nil, fmt.Errorf("lbx image: 非法 header(w=%d h=%d frames=%d)", width, height, frameCount)
	}

	im := &Image{Width: width, Height: height, FrameTime: frameTime, Flags: flags}

	// frame offset 表:(frameCount+1) 個 u32,絕對 offset,最後一個須等於資料長度。
	pos := 12
	need := pos + 4*(frameCount+1)
	if need > len(data) {
		return nil, fmt.Errorf("lbx image: frame offset 表超出資料")
	}
	offsets := make([]uint32, frameCount+1)
	for i := range offsets {
		offsets[i] = le.Uint32(data[pos : pos+4])
		pos += 4
		if i > 0 && offsets[i] <= offsets[i-1] {
			return nil, fmt.Errorf("lbx image: frame offset 非遞增(第 %d 個)", i)
		}
	}
	if int(offsets[frameCount]) != len(data) {
		return nil, fmt.Errorf("lbx image: 尾端 offset %d 與資料長度 %d 不符", offsets[frameCount], len(data))
	}

	// 內嵌調色盤(緊接 offset 表之後)。
	if flags&flagPalette != 0 {
		if pos+4 > len(data) {
			return nil, fmt.Errorf("lbx image: 調色盤 header 超出資料")
		}
		palStart := int(le.Uint16(data[pos : pos+2]))
		palCount := int(le.Uint16(data[pos+2 : pos+4]))
		pos += 4
		if palStart+palCount > 256 {
			return nil, fmt.Errorf("lbx image: 調色盤越界(start=%d count=%d)", palStart, palCount)
		}
		if pos+palCount*4 > len(data) {
			return nil, fmt.Errorf("lbx image: 調色盤資料超出")
		}
		pal := &Palette{}
		for i := 0; i < palCount; i++ {
			// 每色 4 bytes:byte0 未用,byte1=R6 byte2=G6 byte3=B6。
			r := data[pos+4*i+1]
			g := data[pos+4*i+2]
			b := data[pos+4*i+3]
			pal[palStart+i] = color.RGBA{R: r << 2, G: g << 2, B: b << 2, A: 0xff}
		}
		im.Embedded = pal
		im.PalStart = palStart
		im.PalCount = palCount
		// 注意:pos 之後不再依序讀取,frame 資料以 offsets 絕對定位。
	}

	// 逐幀解碼。
	im.Frames = make([]*Frame, frameCount)
	for i := 0; i < frameCount; i++ {
		fdata := data[offsets[i]:offsets[i+1]]
		fr, err := im.decodeFrame(fdata)
		if err != nil {
			return nil, fmt.Errorf("lbx image: 第 %d 幀解碼失敗: %w", i, err)
		}
		im.Frames[i] = fr
	}
	return im, nil
}

// decodeFrame 解碼單幀。對照 openorion2 Image::decodeFrame(gfx.cpp:476-531),
// 但輸出 palette index + 寫入遮罩,不在此上色。
func (im *Image) decodeFrame(d []byte) (*Frame, error) {
	w, h := im.Width, im.Height
	fr := &Frame{W: w, H: h, Index: make([]uint8, w*h), Written: make([]bool, w*h)}
	le := binary.LittleEndian

	// 未壓縮:逐位元組即 index,全部視為已寫入。
	if im.Flags&flagNoCompress != 0 {
		if len(d) < w*h {
			return nil, fmt.Errorf("未壓縮資料不足(%d < %d)", len(d), w*h)
		}
		copy(fr.Index, d[:w*h])
		for i := range fr.Written {
			fr.Written[i] = true
		}
		return fr, nil
	}

	// 壓縮:scan-line RLE。
	p := 0
	readU16 := func() (uint16, bool) {
		if p+2 > len(d) {
			return 0, false
		}
		v := le.Uint16(d[p : p+2])
		p += 2
		return v, true
	}

	size, ok := readU16()
	if !ok {
		return nil, fmt.Errorf("讀取起始標記失敗")
	}
	y32, ok := readU16()
	if !ok {
		return nil, fmt.Errorf("讀取起始 y 失敗")
	}
	if size != 1 {
		return nil, fmt.Errorf("首列標記 != 1(得 %d)", size)
	}
	y := int(y32)

	for y < h {
		rowBase := y * w
		x := 0
		for x < w {
			sz, ok := readU16()
			if !ok {
				return nil, fmt.Errorf("讀取 run size 失敗")
			}
			skip, ok := readU16()
			if !ok {
				return nil, fmt.Errorf("讀取 run skip 失敗")
			}
			if sz == 0 {
				y += int(skip)
				break
			}
			if x+int(skip)+int(sz) > w {
				return nil, fmt.Errorf("掃描線溢出(x=%d skip=%d size=%d w=%d)", x, skip, sz, w)
			}
			x += int(skip) + int(sz)
			idx := rowBase + (x - int(sz)) // 起點 = 跳過 skip 後的位置
			if p+int(sz) > len(d) {
				return nil, fmt.Errorf("像素資料不足")
			}
			for i := 0; i < int(sz); i++ {
				fr.Index[idx+i] = d[p+i]
				fr.Written[idx+i] = true
			}
			p += int(sz)
			if sz%2 == 1 {
				p++ // 奇數 size 補 1 byte 對齊
			}
		}
	}
	return fr, nil
}

// ToRGBA 用給定調色盤把幀上色成 *image.RGBA。keyColor 為 true 時,index 0 視為透明;
// 未寫入的像素(RLE 跳過)一律透明。
func (fr *Frame) ToRGBA(pal *Palette, keyColor bool) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, fr.W, fr.H))
	for i := 0; i < fr.W*fr.H; i++ {
		if !fr.Written[i] || (keyColor && fr.Index[i] == 0) {
			continue // 透明(image.NewRGBA 預設全 0)
		}
		img.Pix[4*i+0] = pal[fr.Index[i]].R
		img.Pix[4*i+1] = pal[fr.Index[i]].G
		img.Pix[4*i+2] = pal[fr.Index[i]].B
		img.Pix[4*i+3] = 0xff
	}
	return img
}
