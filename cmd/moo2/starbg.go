package main

// starbg.go:星圖的**視差星空背景**(原版 `Draw_Paralax_` @ 0x8500F)。
//
// remake 先前把星圖區塗成一片 `RGBA{6,6,16}` 的純黑——那是佔位。原版底下鋪的是三層
// 會各自以不同速度捲動的星空圖。
//
// ============ 一手來源(反組譯)============
//
// `Draw_Paralax_` 的結構完全規則,三層一模一樣只差來源與位移:
//
//	sub_8FD71(edx=0x16, ebx=0x20F, ecx=0x1A5)   ; 裁切區:x 22..527、y ..421
//	for layer in 0, 1, 2:
//	    img = STARBG.LBX 資產 layer
//	    Draw(x,       y      )
//	    Draw(x − 640, y      )                   ; 0x280 = 640
//	    Draw(x − 640, y − 480)                   ; 0x1E0 = 480
//	    Draw(x,       y − 480)
//
// 每張圖就是 640×480(實測 STARBG.LBX 0/1/2 皆是),四次貼圖是**環繞平鋪**:捲動時
// 從右邊/下面出去的部分要從左邊/上面接回來。位移取自三對全域變數:
//
//	層 0:x = word_199980、y = word_19997E
//	層 1:x = word_19998A、y = word_19998E
//	層 2:x = word_199984、y = word_199986
//
// 三對各自更新,更新速率不同 —— 這就是「視差」:近的那層捲得快,遠的慢。
//
// ⚠ **remake 的星圖不捲動**(固定全銀河檢視),所以三個位移都是 0,環繞平鋪的另外三次
// 全部落在畫布外,等於三層各貼一次 (0,0)。真的做出捲動時,要把那四次貼圖補回來,
// 而不是改圖或改位移公式。
//
// ⚠ 裁切是必要的:圖是 640×480、星圖框只有 24..523 × 24..418,不裁會蓋掉左右的框架美術。

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	starBGLBX    = "starbg.lbx"
	starBGLayers = 3 // 原版就是三層(Draw_Paralax_ 硬寫的三段)
)

// starBGImage 解出第 n 層星空並快取。STARBG 沒有內嵌調色盤,靠 buffer0 基底。
func (b *sceneBuilder) starBGImage(layer int) *ebiten.Image {
	if b.res == nil {
		return nil // 沒有資產解析器(單元測試等):畫面自己降級,不要 panic
	}
	key := starBGLBX + ":" + string(rune('0'+layer))
	if im, hit := b.colBldgCache[key]; hit {
		return im
	}
	if b.colBldgCache == nil {
		b.colBldgCache = map[string]*ebiten.Image{}
	}
	var img *ebiten.Image
	if im, err := decodeAsset(b.res, starBGLBX, layer); err == nil && len(im.Frames) > 0 {
		if pal, err := resolvePalette(b.res, im, paletteChain{{"buffer0.lbx", 0}}); err == nil {
			img = ebiten.NewImageFromImage(im.Frames[0].ToRGBADropTranslucent(pal, im.KeyColor()))
		}
	}
	b.colBldgCache[key] = img
	return img
}

// starBGFill 是三層星空底下的純黑底。
//
// ⚠ 這一層**不能省**,而且踩過:三層星空全都是**稀疏的點疊在透明上**——層 0 最多的三種
// 索引是 2/3/5,對到 (16,16,24) / (24,24,32) / (44,44,52) 這種極暗藍灰,本來就是要疊在
// 黑底上才看得出是「遠方的星塵」。把底色拿掉改成「有資產就不填」的話,透明處會露出底下
// `buffer0.lbx#0` 的框架美術,整片星圖變成白底黑點。
// 原版也是先 `Fill` 再貼視差層(`Draw_Main_Screen_Filled_`)。
var starBGFill = color.RGBA{6, 6, 16, 255}

// drawStarmapBackground 鋪星圖底:先純黑,再疊三層星空(取不到資產就只有純黑)。
func (b *sceneBuilder) drawStarmapBackground(dst *ebiten.Image) {
	fillPanel(dst, starVX0, starVY0, starVX1-starVX0, starVY1-starVY0, starBGFill, false)
	clip, ok := dst.SubImage(image.Rect(starVX0, starVY0, starVX1, starVY1)).(*ebiten.Image)
	if !ok {
		return
	}
	for layer := 0; layer < starBGLayers; layer++ {
		im := b.starBGImage(layer)
		if im == nil {
			continue
		}
		// 位移固定 0(見檔頭):貼 (0,0) 一次就好,環繞平鋪的另外三次全在畫布外。
		clip.DrawImage(im, &ebiten.DrawImageOptions{})
	}
}
