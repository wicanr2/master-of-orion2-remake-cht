package smk

import "fmt"

// huffman.go:Smacker 的兩層 Huffman。
//
// Smacker 不用一般的「一棵樹配一組碼長」,而是套了兩層:
//
//	① **8 位元小樹**:位元流上直接寫樹的形狀——讀到 0 是葉(接著八個位元是值),
//	   讀到 1 是節點(遞迴讀左右子樹)。這種寫法不需要另外傳碼長表。
//	② **16 位元大樹**:形狀同上,但每個葉的值是「用小樹 A 解出低位元組、
//	   用小樹 B 解出高位元組」組出來的。
//
// 大樹另外帶三個 **escape 值**。建樹時葉子若命中 escape,該葉存 0 並把它的索引記進
// last[0..2];解碼時每取一個值,就把 last 三格做 MRU 輪替(取到的值塞進 last[0],
// 原本的往後推)。效果是「最近用過的三個值」永遠掛在極短的碼上——這是 Smacker
// 壓縮率的主要來源之一,也是為什麼樹的葉子必須可寫、不能用不可變的指標樹。
//
// 因此這裡用**扁平陣列**表示大樹:節點帶 nodeFlag 與「跳過幾格到右分支」的位移,
// 葉直接存值。MRU 輪替就是改陣列裡三個位置的值。

// nodeFlag 標記扁平陣列裡的某格是節點而不是葉。
const nodeFlag = 0x80000000

// maxTreeDepth 是遞迴建樹的深度上限。正常檔案遠達不到;
// 壞掉/被截斷的檔案會讓遞迴無止境下去,這條是防呆不是格式限制。
const maxTreeDepth = 512

// bitReader 是 LSB-first 的位元讀取器(Smacker 的位元流是每個位元組從最低位讀起)。
type bitReader struct {
	data     []byte
	pos      int // 位元位置
	overread bool
}

func newBitReader(data []byte) *bitReader { return &bitReader{data: data} }

// bit 讀一個位元。讀過尾端一律回 0——影像位元流常常只用到最後一個位元組的一部分,
// 把「讀完了」當錯誤會讓正常的檔案解不完。
func (r *bitReader) bit() uint32 {
	i := r.pos >> 3
	if i >= len(r.data) {
		r.pos++
		r.overread = true
		return 0
	}
	v := uint32(r.data[i]>>(uint(r.pos)&7)) & 1
	r.pos++
	return v
}

// bits 讀 n 個位元(先讀到的是低位)。
func (r *bitReader) bits(n int) uint32 {
	var v uint32
	for i := 0; i < n; i++ {
		v |= r.bit() << uint(i)
	}
	return v
}

// tree8 是 8 位元小樹。只在建大樹時用到,所以用指標樹就夠,不需要扁平化。
type tree8 struct {
	leaf      bool
	val       uint8
	zero, one *tree8
}

// readTree8 從位元流讀出一棵 8 位元小樹。
func readTree8(br *bitReader, depth int) (*tree8, error) {
	if depth > maxTreeDepth {
		return nil, fmt.Errorf("8 位元樹過深(>%d),資料可能損壞", maxTreeDepth)
	}
	if br.bit() == 0 {
		return &tree8{leaf: true, val: uint8(br.bits(8))}, nil
	}
	n := &tree8{}
	var err error
	if n.zero, err = readTree8(br, depth+1); err != nil {
		return nil, err
	}
	if n.one, err = readTree8(br, depth+1); err != nil {
		return nil, err
	}
	return n, nil
}

// decode 用小樹解出一個位元組。
func (t *tree8) decode(br *bitReader) uint8 {
	if t == nil {
		return 0 // 該棵樹不存在(位元流明示略過)→ 該位元組固定為 0
	}
	n := t
	for !n.leaf {
		if br.bit() != 0 {
			n = n.one
		} else {
			n = n.zero
		}
		if n == nil {
			return 0
		}
	}
	return n.val
}

// bigTree 是 16 位元大樹的扁平表示 + MRU 三格。
type bigTree struct {
	vals []uint32 // 節點:nodeFlag | 右分支位移;葉:值
	last [3]int   // MRU 三格在 vals 裡的索引
}

// readBigTree 從位元流讀出一棵大樹。sizeHint 是標頭給的該樹大小,用來預配空間;
// 它是位元組數不是節點數,所以只當容量提示,不當上限。
func readBigTree(br *bitReader, sizeHint uint32) (*bigTree, error) {
	bt := &bigTree{}
	if br.bit() == 0 {
		// 位元流明示「沒有這棵樹」:退化成單一個 0。
		bt.vals = []uint32{0, 0, 0, 0}
		bt.last = [3]int{1, 2, 3}
		return bt, nil
	}
	// 兩棵小樹各有一個「有沒有」的前置位元。
	var lowT, highT *tree8
	var err error
	if br.bit() != 0 {
		if lowT, err = readTree8(br, 0); err != nil {
			return nil, fmt.Errorf("低位元組樹: %w", err)
		}
		br.bit() // 樹尾的一個結束位元
	}
	if br.bit() != 0 {
		if highT, err = readTree8(br, 0); err != nil {
			return nil, fmt.Errorf("高位元組樹: %w", err)
		}
		br.bit()
	}
	var esc [3]uint32
	for i := range esc {
		esc[i] = br.bits(16)
	}
	bt.last = [3]int{-1, -1, -1}
	bt.vals = make([]uint32, 0, int(sizeHint/4)+8)

	var rec func(depth int) (int, error)
	rec = func(depth int) (int, error) {
		if depth > maxTreeDepth {
			return 0, fmt.Errorf("16 位元樹過深(>%d),資料可能損壞", maxTreeDepth)
		}
		if br.bit() == 0 { // 葉
			v := uint32(lowT.decode(br)) | uint32(highT.decode(br))<<8
			i := len(bt.vals)
			// escape 比對是**逐一往下、先中先算**,而且不看這一格先前有沒有被設過——
			// 同一個 escape 值出現在多個葉時,last[k] 會指到**最後**那個。
			// (先前這裡多加了「只認第一個命中」的守衛,解出來的畫面會帶區塊雜訊。)
			for k := 0; k < 3; k++ {
				if esc[k] == v {
					bt.last[k] = i
					v = 0 // escape 葉先存 0,之後由 MRU 輪替填入
					break
				}
			}
			bt.vals = append(bt.vals, v)
			return 1, nil
		}
		// 節點:先佔位,解完 0 分支才知道右分支的位移。
		t := len(bt.vals)
		bt.vals = append(bt.vals, 0)
		n0, err := rec(depth + 1)
		if err != nil {
			return 0, err
		}
		bt.vals[t] = nodeFlag | uint32(n0)
		n1, err := rec(depth + 1)
		if err != nil {
			return 0, err
		}
		return n0 + n1 + 1, nil
	}
	if _, err := rec(0); err != nil {
		return nil, err
	}
	br.bit() // 樹尾的結束位元

	// 沒被任何葉命中的 escape:補三格空位給 MRU 用,免得索引指到樹裡的真葉。
	for k := 0; k < 3; k++ {
		if bt.last[k] == -1 {
			bt.last[k] = len(bt.vals)
			bt.vals = append(bt.vals, 0)
		}
	}
	return bt, nil
}

// get 從位元流解出一個 16 位元值,並更新 MRU 三格。
func (bt *bigTree) get(br *bitReader) (uint32, error) {
	i := 0
	for steps := 0; ; steps++ {
		if i >= len(bt.vals) {
			return 0, fmt.Errorf("走出樹外(索引 %d / 共 %d),位元流可能已錯位", i, len(bt.vals))
		}
		v := bt.vals[i]
		if v&nodeFlag == 0 {
			break
		}
		if br.bit() != 0 {
			i += int(v &^ nodeFlag)
		}
		i++
		if steps > maxTreeDepth {
			return 0, fmt.Errorf("解碼走太深(>%d),位元流可能已錯位", maxTreeDepth)
		}
	}
	v := bt.vals[i]
	// MRU:剛用到的值推到 last[0],原本的往後擠。與 last[0] 相同就不動
	// (已經在最前面,再輪替只是白做)。
	if v != bt.vals[bt.last[0]] {
		bt.vals[bt.last[2]] = bt.vals[bt.last[1]]
		bt.vals[bt.last[1]] = bt.vals[bt.last[0]]
		bt.vals[bt.last[0]] = v
	}
	return v, nil
}
