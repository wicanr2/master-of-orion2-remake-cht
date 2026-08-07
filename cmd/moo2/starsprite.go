package main

// starsprite.go:星圖上的**星球 sprite**(原版 `Draw_A_Star_` @ 0x83B02 →
// `Get_Star_Picture_Seg_` @ 0x81CD3)。
//
// remake 先前用 `vector.DrawFilledCircle` 畫色圓——那是佔位。原版是 BUFFER0.LBX 裡
// 一組會閃爍的星球圖。
//
// ============ 資產編號(兩個獨立來源逐項相同)============
//
//	① 反組譯 `Get_Star_Picture_Seg_`(eax = 光譜、ebx = 大小):
//	     edx = al × 6
//	     if al == 6:  eax = zoom              ; 6 = 黑洞,沒有大小變體
//	     else:        eax = zoom + bl
//	     資產 = 0x94(148) + eax + edx
//	② openorion2 `src/galaxy.cpp:1664`:
//	     `_starimg[s->spectralClass][_zoom + s->size]`,`ASSET_GALAXY_STAR_IMAGES 148`,
//	     `STAR_TYPE_COUNT 6`、`GALAXY_STAR_SIZES 6`
//
//	→ **資產 = 148 + 光譜×6 + (縮放 + 大小)**
//
// 實測 BUFFER0.LBX 148..183 正好是 6 組 × 6 張,每組尺寸 33/29/25/23/21/17(遞減),
// 每張 5 幀(閃爍動畫)。索引 0 最大、5 最小。
//
// 光譜與大小的列舉 remake 早就對上了:
//
//	`shell.Star.Spectral` 0=藍 1=白 2=黃 3=橙 4=紅 5=棕 6=黑洞
//	  = openorion2 `SpectralClass{Blue, White, Yellow, Orange, Red, Brown, ...}`
//	`shell.Star.Size` 0=大 .. 3=小
//	  = openorion2 `StarSize{Large=0, Medium=1, Small=2, Tiny=3}`
//
// ============ ⚠ 縮放:原版那個值 remake 沒有對應物 ============
//
// `Map_Scale_To_Zoom_Level_` @ 0x79917 把地圖比例(`word_199992` = 10/15/20/30,
// 即 openorion2 的 `galaxySizeFactors`)換成縮放 0..3。而 openorion2 `galaxy.cpp:1385`
// 揭露關鍵:**縮放是銀河尺寸的函式,不是玩家控制的**——
// 銀河越大 → sizeFactor 越大 → 星球畫得越小,因為要塞進同一個視窗。
//
// 螢幕座標是 `21 + 10×(x − 捲動)/sizeFactor`(openorion2 `transformX`),
// 所以 sizeFactor 越大距離越短 → **0 = 最放大、3 = 最縮小**。
//
// remake 這裡有個**模型落差**:它把星球座標正規化成 0..1 再攤滿視窗,銀河多大都一樣,
// 也沒有捲動。所以原版那個「銀河尺寸 → 縮放」的對應**接不過來**,不存在忠實值。
// 這裡固定用 3(最縮小)——理由是 remake 永遠一次顯示整個銀河,語意上就是最縮小那段。
// **這是 remake 自己的選擇,不是原版真值**,做出捲動/縮放之前不要把它寫成真值。
//
// 每張 5 幀的閃爍動畫沒做,只取第 0 幀(同艦隊圖示)。
//
// ============ 黑洞:公式自己證明了自己 ============
//
// 光譜 6 那條分支(`al == 6` 不加大小)算出來是 `148 + 6×6 + 縮放` = **184 + 縮放**。
// 而 openorion2 `galaxy.cpp:45` 另外命名了一個常數:
//
//	#define ASSET_GALAXY_BHOLE_IMAGES 184
//	_bholeimg[i] = getImage(GALAXY_ARCHIVE, ASSET_GALAXY_BHOLE_IMAGES + i, …)   // i = 縮放
//
// **兩邊落在同一個數字上**——一邊是從組語推出來的算式、一邊是別人專案裡獨立命名的常數。
// 這同時證實了整條公式(基底 148、每組 6 張、光譜×6)與「光譜 6 = 黑洞」這個對應。
//
// ⚠ 不過原版畫黑洞走的是 `Draw_Black_Holes_` @ 0x83BF9 那條**獨立的迴圈**(黑洞在原版
// 是星球以外的地圖物件,不在星球陣列裡)。remake 目前把黑洞當成光譜 6 的星球一起畫,
// 圖是對的,但阻擋航線那套機制(`Star.blackHoleBlocks`)沒有。

import (
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

const (
	starSpriteLBX  = "buffer0.lbx"
	starSpriteBase = 148 // ASSET_GALAXY_STAR_IMAGES
	starSpriteSize = 6   // GALAXY_STAR_SIZES:每個光譜 6 張(索引 0 最大、5 最小)
	starSpriteZoom = 3   // remake 固定最縮小(見檔頭 ⚠,這是選擇不是真值)
)

// starSpriteAsset 回傳某顆星的 BUFFER0.LBX 資產編號。
// 光譜 6(黑洞)照原版不加大小;其餘 `縮放 + 大小` 夾在 0..5(原版沒有夾,
// 但它的縮放與大小組合不會同時取到最大——remake 的縮放固定 3,不夾會溢位到下一個光譜)。
func starSpriteAsset(spectral, size int) int {
	if spectral < 0 {
		spectral = 0
	}
	if spectral > 6 {
		spectral = 6
	}
	idx := starSpriteZoom
	if spectral != 6 {
		if size < 0 {
			size = 0
		}
		idx += size
	}
	if idx > starSpriteSize-1 {
		idx = starSpriteSize - 1
	}
	return starSpriteBase + spectral*starSpriteSize + idx
}

// starSpriteImage 解出星球圖並快取(取不到回 nil,呼叫端自己退回色圓)。
func (b *sceneBuilder) starSpriteImage(spectral, size int) *ebiten.Image {
	if b.res == nil {
		return nil // 沒有資產解析器(單元測試等):畫面自己降級,不要 panic
	}
	asset := starSpriteAsset(spectral, size)
	key := starSpriteLBX + ":" + strconv.Itoa(asset)
	if im, hit := b.colBldgCache[key]; hit {
		return im
	}
	if b.colBldgCache == nil {
		b.colBldgCache = map[string]*ebiten.Image{}
	}
	var img *ebiten.Image
	if im, err := decodeAsset(b.res, starSpriteLBX, asset); err == nil && len(im.Frames) > 0 {
		if pal, err := resolvePalette(b.res, im, paletteChain{{"buffer0.lbx", 0}}); err == nil {
			img = ebiten.NewImageFromImage(im.Frames[0].ToRGBADropTranslucent(pal, im.KeyColor()))
		}
	}
	b.colBldgCache[key] = img
	return img
}

// drawStarSpriteAt 把星球圖畫在 (cx, cy) 置中,回傳半徑(供外圈的擁有環/選取環用)。
// 沒有原版資產時回 0,呼叫端要自己退回色圓。
//
// 置中的依據是原版 `Draw_A_Star_`:它取 `word_1931AC[縮放+大小]`(就是這 6 張圖的邊長表)
// 除以 2 再從座標扣掉——也就是**以圖的中心對準星球座標**。
func (b *sceneBuilder) drawStarSpriteAt(dst *ebiten.Image, st shell.Star, cx, cy int) float32 {
	im := b.starSpriteImage(st.Spectral, st.Size)
	if im == nil {
		return 0
	}
	w, h := im.Bounds().Dx(), im.Bounds().Dy()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(cx-w/2), float64(cy-h/2))
	dst.DrawImage(im, op)
	return float32(w) / 2
}
