package uifont

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// record.go:文字繪製的「錄→放」機制,供 hi-res(2×)畫布用。
//
// ============ 這東西為什麼存在 ============
//
// remake 的畫布是 640×480(= 原版美術的原生解析),而 CJK 在那個尺度下只有 10–13px,
// 筆畫糊成一團。`rulebook/81` 的處方是**拉高內部畫布**:美術用 nearest 放大保持銳利,
// 文字用正常尺寸畫在放大後的畫布裡。
//
// 問題是 remake 有 **420 個繪製呼叫點**散在十幾個檔案裡,全部改成「座標 ×2」既昂貴又
// 容易漏。這裡換一個做法:
//
//	① 畫面照舊畫進 640×480 的離屏(**一行呼叫端都不用改**)
//	② 過程中的文字繪製被**記錄下來而不畫**
//	③ 離屏用 nearest 放大 2× 貼到真正的畫布
//	④ 記錄下來的文字用 **2× 座標、2× 字級**重畫一次
//
// 結果:美術是銳利的整數倍放大,文字是**在最終解析度上重新柵格化**的——不是把
// 640 的字放大,而是真的用 2× 字級畫。這正是 rulebook/81 要的效果。
//
// ⚠ **代價是 z 序**:錄下來的文字一律最後重播,所以「先畫字、再用圖蓋住字」這種寫法會
// 失效。remake 沒有這種畫法(面板先畫、字後畫),但新增畫面時要記得這條。
//
// ⚠ **量測不參與縮放**:`Measure`/`Wrap` 仍在 logical 尺寸下算,因為排版決策
// (折行、置中)是在 640 空間做的。重播時字級 ×2、座標 ×2,比例維持。

// textOp 是一次被記錄下來的文字繪製。
type textOp struct {
	font     *Font
	s        string
	x, y     float64
	size     float64
	c        color.Color
	centered bool
}

// Recorder 收集一幀之內的文字繪製,供之後在高解析畫布上重播。
type Recorder struct {
	ops []textOp
	// lh 快取「某個 logical 字級」的 1× 與 scale× 行高(見 Replay 的垂直補償)。
	// 行高只跟字級有關、與字串內容無關(單行),所以一個尺寸量一次就夠。
	lh map[float64][2]float64
}

// Reset 清空(每幀開始時呼叫;保留底層陣列避免每幀配置)。
func (r *Recorder) Reset() { r.ops = r.ops[:0] }

// Len 回傳這一幀記錄了幾次文字繪製(供測試/診斷)。
func (r *Recorder) Len() int { return len(r.ops) }

// Replay 把記錄的文字以 scale 倍的座標與字級重畫到 dst。
//
// ⚠ **不能只是把 x/y/size 乘以 scale**,因為 remake 的字型是混合模式:
// `size < 18` 走 12px 點陣、`>= 18` 走向量。1× 的字級全部落在 10–13(點陣),
// 2× 之後全部落在 20–26(**向量**)——換了字型路徑,行框的上緣留白就不一樣了。
//
// 實測(艦艇設計「武器: 雷射」那一行):墨水高度 11px → 22px(**剛好 2 倍,正確**),
// 但墨水上緣從 y+2 變成 y+4(邏輯座標),也就是整體**往下漂 2px**。原版面板的內高
// 只有 96px 要塞四列,漂 2px 就足以讓最後一列被下邊框切掉。
//
// 補償方式是**對齊行框中心**而不是左上角:1× 的行框是 [y, y+h1],2× 的是 [y', y'+h2],
// 讓兩者的中心在放大後重合 → `y' = y*scale + (h1*scale - h2)/2`。
// 置中繪製(DrawCentered)本來就以中心為錨,不需要補。
//
// 水平方向實測是 2× **窄** 8%(向量比點陣緊),不會溢出,所以不補——補了反而會讓
// 依 1× 量測折過行的段落再擠回去。
func (r *Recorder) Replay(dst *ebiten.Image, scale float64) {
	for _, op := range r.ops {
		if op.centered {
			op.font.drawCenteredNow(dst, op.s, op.x*scale, op.y*scale, op.size*scale, op.c)
			continue
		}
		h1, h2 := r.lineHeights(op.font, op.size, scale)
		y := op.y*scale + (h1*scale-h2)/2
		op.font.drawNow(dst, op.s, op.x*scale, y, op.size*scale, op.c)
	}
}

// lineHeights 回傳某字級在 1× 與 scale× 下的行高(帶快取)。
//
// 探針字串固定用「字」:行高由字型 metrics 決定與內容無關,但**要挑一個兩條字型路徑
// 都有字形的字**——用 ASCII 探針量到的向量行高會偏小(沒有 CJK 的 ascent)。
func (r *Recorder) lineHeights(f *Font, size, scale float64) (float64, float64) {
	if v, ok := r.lh[size]; ok {
		return v[0], v[1]
	}
	_, h1 := f.Measure("字", size)
	_, h2 := f.Measure("字", size*scale)
	if r.lh == nil {
		r.lh = map[float64][2]float64{}
	}
	r.lh[size] = [2]float64{h1, h2}
	return h1, h2
}

// activeRecorder 是目前生效的錄製器(nil = 直接畫)。
//
// ⚠ 用套件級變數而不是逐 Font 欄位:一個畫面會用到兩三個 Font 實例(內文、標題、
// 向量字),而錄製是**整幀的模式**不是某個字型的屬性。設定點只有一處
// (`interactiveApp.Draw`),而且是單執行緒的 ebiten 繪製迴圈。
var activeRecorder *Recorder

// StartRecording 讓之後所有 Font.Draw / DrawCentered 改成記錄而不繪製。
func StartRecording(r *Recorder) { activeRecorder = r }

// StopRecording 恢復直接繪製。
func StopRecording() { activeRecorder = nil }

// ============ z 序屏障 ============
//
// 錄→放最大的代價是 **z 序**:錄下來的字一律最後重播,所以「先畫字、再用不透明面板蓋住字」
// 這種寫法會失效——字會浮在面板上面。remake 到處都有這種寫法(星圖畫完星名,再把指揮點數
// 視窗疊上去),不是少數例外。
//
// 屏障就是解法:畫不透明面板**之前**呼叫 Barrier(),合成器會把「目前為止錄到的字」立刻
// 畫出去並清空,之後畫的美術自然蓋在那些字上面。
//
// ⚠ 合成器的 flush 必須「**貼完就把離屏清空**」,不能重貼整張離屏:離屏的背景是不透明的,
// 重貼會把上一輪已經畫好的字整片洗掉。清空之後每一輪只帶著「這一輪新畫的美術」,
// source-over 的結合律保證疊出來的結果與直接畫在同一張圖上一致。
//
// hook 由 hi-res 合成器安裝;uiScale==1 時是 nil,Barrier() 整條零成本。
var flushHook func()

// SetFlushHook 安裝屏障的實作(nil = 停用)。
func SetFlushHook(f func()) { flushHook = f }

// Barrier 把目前錄到的文字立刻沖出去,讓之後畫的美術能蓋住它們。
func Barrier() {
	if flushHook != nil && activeRecorder != nil && len(activeRecorder.ops) > 0 {
		flushHook()
	}
}

// BarrierRect 是**只在真的會蓋到字時才沖**的版本(x,y,w,h 是即將畫的美術範圍)。
//
// 為什麼要這一版:屏障插在每個填色/貼圖之前,而一幀可能有幾十個——每沖一次就是一張
// 1280×960 的貼圖 + 清屏。實際上絕大多數美術根本不落在任何字上(星圖上的星、圖示、
// 擦底板都畫在自己那一小塊),測一下矩形相交就能全部略過。
func BarrierRect(x, y, w, h float64) {
	if flushHook == nil || activeRecorder == nil || len(activeRecorder.ops) == 0 {
		return
	}
	if activeRecorder.overlaps(x, y, x+w, y+h) {
		flushHook()
	}
}

// bounds 保守估計一次文字繪製覆蓋的矩形。
//
// ⚠ 刻意**不呼叫 Measure**:那要做字型排版,而這個函式一幀會被叫上萬次。改用
// 「字數 × 字級」當寬度上界(CJK 剛好滿格,拉丁字會高估)——**高估只會多沖幾次,
// 低估才會漏蓋**,所以往保守的那邊估。
func (op textOp) bounds() (x0, y0, x1, y1 float64) {
	w := float64(len([]rune(op.s))) * op.size
	h := op.size * 1.6
	if op.centered {
		return op.x - w/2, op.y - h/2, op.x + w/2, op.y + h/2
	}
	return op.x, op.y, op.x + w, op.y + h
}

// overlaps 回報已錄的文字裡有沒有任何一段落在 (x0,y0)-(x1,y1) 內。
func (r *Recorder) overlaps(x0, y0, x1, y1 float64) bool {
	for _, op := range r.ops {
		a0, b0, a1, b1 := op.bounds()
		if a0 < x1 && a1 > x0 && b0 < y1 && b1 > y0 {
			return true
		}
	}
	return false
}
