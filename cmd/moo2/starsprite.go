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

// starSpriteImage 解出星球圖(第 0 幀)並快取(取不到回 nil,呼叫端自己退回色圓)。
func (b *sceneBuilder) starSpriteImage(spectral, size int) *ebiten.Image {
	return b.starSpriteFrame(starSpriteAsset(spectral, size), 0)
}

// starSpriteFrameCount 回傳某張星圖有幾幀(取不到回 0)。
//
// BUFFER0 的星球圖每張 5 幀(閃爍)、黑洞那張 16 幀(旋渦)——跑 `lbxinfo` 確認過。
func (b *sceneBuilder) starSpriteFrameCount(asset int) int {
	if b.res == nil {
		return 0
	}
	im, err := decodeAsset(b.res, starSpriteLBX, asset)
	if err != nil {
		return 0
	}
	return len(im.Frames)
}

// starSpriteFrame 解出星球圖的第 frame 幀並快取。
//
// 快取的 key 帶幀號 —— 先前只解第 0 幀,所以 key 不必帶;加上動畫之後不帶就會整個星圖
// 都用同一幀。
func (b *sceneBuilder) starSpriteFrame(asset, frame int) *ebiten.Image {
	if b.res == nil {
		return nil // 沒有資產解析器(單元測試等):畫面自己降級,不要 panic
	}
	key := starSpriteLBX + ":" + strconv.Itoa(asset) + ":" + strconv.Itoa(frame)
	if im, hit := b.colBldgCache[key]; hit {
		return im
	}
	if b.colBldgCache == nil {
		b.colBldgCache = map[string]*ebiten.Image{}
	}
	var img *ebiten.Image
	if im, err := decodeAsset(b.res, starSpriteLBX, asset); err == nil && frame >= 0 && frame < len(im.Frames) {
		if pal, err := resolvePalette(b.res, im, paletteChain{{"buffer0.lbx", 0}}); err == nil {
			img = ebiten.NewImageFromImage(im.Frames[frame].ToRGBADropTranslucent(pal, im.KeyColor()))
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
	im := b.starSpriteFrame(starSpriteAsset(st.Spectral, st.Size), b.starSpriteFrameFor(st))
	if im == nil {
		return 0
	}
	w, h := im.Bounds().Dx(), im.Bounds().Dy()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(cx-w/2), float64(cy-h/2))
	dst.DrawImage(im, op)
	return float32(w) / 2
}

// ---- 原版的動畫模型(`sub_12A478` 這支通用貼圖器)----
//
// 查黑洞動畫時順帶把引擎層的規則挖出來了,記在這裡免得下次重找:
//
//	word[圖+4] = 目前幀號        word[圖+6] = 幀數
//	sub_12B726(圖)      → 幀號 ← 0
//	sub_12B753(圖, 幀號) → 幀號 ← min(幀號, 幀數−1)
//	sub_12A478(x, y, 圖) → **畫完自動把幀號 +1**([圖+0Bh] 的 0x20 位元控制要不要循環)
//
// 「畫一次進一幀」是**貼圖器的預設行為**,所以呼叫端要靜止就得每次先 `sub_12B726` 歸零
// (艦隊圖示就是這樣)、要自訂節奏就每次 `sub_12B753` 寫死幀號(黑洞就是這樣)。
//
// ---- 黑洞的旋渦動畫(原版 `Draw_Black_Holes_` @ 0x83BF9)----
//
// 星圖上的黑洞不是靜止的圖。原版那支的推進規則整段可讀:
//
//	計數 = (_black_hole_anim_count[黑洞序號] + 1) % (幀數 × 2)
//	幀號 = 計數 / 2      → 再 `sub_12B753` 寫進圖裡
//
// (每個黑洞各有獨立計數器,序號夾在 0..10;但它們都從 0 起、每次重畫一起 +1,
//  所以幀號恆同——remake 用單一 `animTick` 是等價的。)
//
// 也就是**每一幀停留 2 次重畫**,而且**每個黑洞各有獨立的計數器**(序號夾在 0..10)。
// 那個「除以 2」在一般星球的 `Draw_A_Star_` 裡也出現(`sar eax, 1`)——同一個比例被兩處
// 獨立證實,不是單點讀出來的。
//
// ⚠ **只做黑洞,一般星球維持靜止——而且查證後確認那是原版行為,不是留白。**
//
// `Draw_A_Star_` 裡確實有一段閃爍碼:`star[+0x64]` 是計數器(−1 = 不閃),
// `幀號 = (計數 / 2) % 幀數`(同一個「除以 2」),計數達到 `star[+0x65]` 這個爆發長度就把
// 兩個欄位都寫回 −1,並把全域 `word_19C164` 減 1。
//
// 問題是**沒有任何一處程式碼會啟動它**(2026-08-07 全檔查證):
//
//	· `star[+0x64]` / `star[+0x65]` 的位元組寫入只有四處 reset 迴圈,全部寫 0xFF(= −1),
//	  外加一處星系生成時的暫存搬運。**沒有「歸零 + 寫入爆發長度」那組設定**。
//	· `word_19C164` **只有歸零與遞減,全檔沒有任何遞增**。一個只減不加的預算,
//	  不可能是還在運作的機制的閘門。
//
// 所以出貨版的一般星球**不會閃**,那段是死碼。remake 固定第 0 幀與原版一致。
// (正對照:同樣的搜法找得到星球結構 `+0x16` 光譜欄位的寫入端,方法沒有假陰性。)
//
// 黑洞不一樣——`Draw_Black_Holes_` 每次繪製都主動算幀號並呼叫 `sub_12B753` 寫進去,
// 動畫無條件連續。
//
// ⚠ **一次重畫 = 一個 ebiten 幀** 是 remake 的對應:原版的主畫面重畫由它自己的迴圈驅動,
// 頻率沒有解出來,所以**動畫的絕對速度是 remake 的選擇**,只有「2 次重畫換 1 幀」這個**比例**
// 是原版真值。

// blackHoleHoldRedraws 是每一幀停留幾次重畫(原版 `幀數 × 2` 再除以 2 的那個 2)。
const blackHoleHoldRedraws = 2

// blackHoleSpectralClass 是黑洞的光譜值(與 shell 的 blackHoleSpectral 一致)。
const blackHoleSpectralClass = 6

// blackHoleFrameAt 是原版那兩行的直譯:計數 %(幀數×2),再 /2。
//
// 抽成純函式是為了能單獨驗——`starSpriteFrameFor` 要有資產解析器才跑得動,
// 而這條算式本身跟資產無關。
func blackHoleFrameAt(tick, frames int) int {
	if frames <= 1 {
		return 0
	}
	if tick < 0 {
		tick = 0
	}
	return (tick % (frames * blackHoleHoldRedraws)) / blackHoleHoldRedraws
}

// starSpriteFrameFor 回傳這顆星目前該畫第幾幀。
//
// 黑洞:依 `b.animTick` 連續循環。其餘:固定第 0 幀(見上方 ⚠)。
func (b *sceneBuilder) starSpriteFrameFor(st shell.Star) int {
	if st.Spectral != blackHoleSpectralClass {
		return 0
	}
	return blackHoleFrameAt(b.animTick, b.starSpriteFrameCount(starSpriteAsset(st.Spectral, st.Size)))
}
