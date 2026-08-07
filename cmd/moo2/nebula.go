package main

// nebula.go:星圖上的星雲(原版 `Draw_Nebulae_` @ 0x84F8F、`Load_Nebula_Pictures_` @ 0x8DA07)。
//
// 規則面在 `internal/shell/nebula.go`(數量表、判定門檻、戰鬥護盾規則);這一檔只管
// **圖從哪來、畫在哪、以及那個「像素 > 5」的判定要讀哪張圖**——判定要碰 LBX,
// 而規則層是純邏輯不碰資產,所以由這邊算好再灌回去。
//
// ============ 圖在 STARBG.LBX ============
//
// `Load_Nebula_Pictures_` 用的字串就是 `"starbg.lbx"` —— 和星圖背景那三層同一個檔。
// 資產配置(跑 `lbxinfo` 確認過,不是推的):
//
//	0..5   640×480    星空層(remake 目前用 0..2,見 starbg.go)
//	6 起   每 4 張一組  星雲,一組 = 同一團的 4 個縮放等級(大 → 小)
//
// 全檔 54 張 → (54 − 6) / 4 = **12 種**,與 `shell.nebulaTypes` 一致。
// 原版 `Draw_Nebulae_` 取的是 `nebula_pict_seg[星雲][縮放]`,縮放來自星圖目前的倍率;
// remake 的星圖沒有縮放,固定取**最大那張**(組內第 0 張)——它就是 1:1 的那一級。
//
// ============ 判定用的就是同一張圖 ============
//
// `Point_Is_In_Nebula_N_` 拿 `(點 − 星雲原點) / 3` 去索引圖的**調色盤索引值**,> 5 才算在裡面。
// patch 1.5 手冊逐字佐證:「a star is considered "in nebula" if the respective pixel value of
// the nebula picture is greater than 5」。
//
// ⚠ 那個 `/ 3` 是原版銀河座標與圖素的比例。remake 的星圖座標是正規化 0..1、最後映到螢幕,
// 所以這裡**直接在螢幕座標上判定**:星的螢幕位置減去星雲左上角,就是圖上的像素。
// 語意相同(「星壓在圖的亮處」),比例來源不同——因為 remake 的銀河座標系本來就不是原版的。

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

const (
	nebulaLBX = starBGLBX // "starbg.lbx",與星空層同檔
	// nebulaFirstAsset 是第一團星雲的第一張;之前 0..5 是 640×480 的星空層。
	nebulaFirstAsset = 6
	// nebulaZoomLevels 是每團的縮放張數(`Load_Nebula_Pictures_` 內圈跑 4 次)。
	nebulaZoomLevels = 4
	// nebulaZoom 是 remake 取的縮放等級。星圖沒有縮放,固定用最大的那張。
	nebulaZoom = 0
	// nebulaInPixelThreshold 是「在星雲內」的調色盤索引門檻(原版 `cmp …, 5; jbe → 不在`)。
	nebulaInPixelThreshold = 5
)

// nebulaAsset 回傳第 typ 種星雲在 STARBG.LBX 的資產編號。
func nebulaAsset(typ int) int {
	if typ < 0 {
		typ = 0
	}
	return nebulaFirstAsset + typ*nebulaZoomLevels + nebulaZoom
}

// nebulaScreenXY 把星雲的正規化左上角換成星圖上的螢幕座標。
//
// 用的是與星星同一組映射(`starVX0..starVX1` / `starVY0..starVY1`),否則星雲會和星圖錯位。
func nebulaScreenXY(n shell.Nebula) (float64, float64) {
	return starVX0 + n.X*(starVX1-starVX0), starVY0 + n.Y*(starVY1-starVY0)
}

// nebulaPalette 是星雲圖的調色盤鏈。
//
// ⚠ 踩過:第一版沿用殖民地畫面的鏈(`colonyBasePalette`,含殖民地框架盤),整團星雲畫成**鮮紅色**。
// STARBG 沒有內嵌調色盤,顏色完全由鏈決定,而殖民地框架盤會蓋掉星雲用到的那些索引。
// 星圖該用的是星空層那條 —— 見 `starBGImage`,兩邊必須一致(同一個檔、同一個畫面)。
var nebulaPalette = paletteChain{{"buffer0.lbx", 0}}

// nebulaImage 解一張星雲圖並快取。
func (b *sceneBuilder) nebulaImage(typ int) *ebiten.Image {
	return b.colonyScreenImage(nebulaLBX, nebulaAsset(typ), nebulaPalette)
}

// drawNebulae 把星雲畫在星圖上。
//
// 位置在**星空底之後、星星與連線之前**:星雲是背景地形,星星要壓在它上面。
// 原版 `Draw_Main_Main_Screen_` 的圖層序也是先 `Draw_Starfield`/`Draw_Nebulae_`,
// 之後才是蟲洞連線、艦隊、星星。
func (b *sceneBuilder) drawNebulae(dst *ebiten.Image, nebulae []shell.Nebula) {
	for _, n := range nebulae {
		im := b.nebulaImage(n.Type)
		if im == nil {
			continue
		}
		x, y := nebulaScreenXY(n)
		op := &ebiten.DrawImageOptions{}
		// 星雲是稀薄的電離雲,原版畫得很淡;這裡壓透明度免得蓋掉星星。
		// ⚠ 這是 remake 的呈現選擇——原版走的是自己的貼圖模式,沒有解出來。
		op.ColorScale.ScaleAlpha(0.75)
		op.GeoM.Translate(x, y)
		dst.DrawImage(im, op)
	}
}

// nebulaMask 是一團星雲的遮罩(調色盤索引)。
type nebulaMask struct {
	w, h int
	idx  []uint8
}

// nebulaMaskFor 取(並快取)某種星雲的遮罩。
//
// ⚠ **一定要快取**:派遣艦隊時要沿航線取樣上百點(見 shell/route.go 的 `RouteCrossesNebula`),
// 每點都問一次遮罩,而 `decodeAsset` 自己沒有快取。不快取就是每次派遣重解上百次 LBX。
func (b *sceneBuilder) nebulaMaskFor(typ int) *nebulaMask {
	if b.res == nil {
		return nil
	}
	if m, hit := b.nebMaskCache[typ]; hit {
		return m
	}
	if b.nebMaskCache == nil {
		b.nebMaskCache = map[int]*nebulaMask{}
	}
	var m *nebulaMask
	if im, err := decodeAsset(b.res, nebulaLBX, nebulaAsset(typ)); err == nil && len(im.Frames) > 0 {
		m = &nebulaMask{w: im.Width, h: im.Height, idx: im.Frames[0].Index}
	}
	b.nebMaskCache[typ] = m // 取不到也快取,避免每次重試
	return m
}

// nebulaMaskHit 回傳星圖上的螢幕點 (sx, sy) 是否落在第 typ 種星雲(左上角在 nx, ny)的亮處。
//
// 判定就是原版那條:取該點在圖上的調色盤索引,> 5 算在裡面。
// 取不到圖一律回 false —— 沒有圖就不該憑空判定有星雲。
func (b *sceneBuilder) nebulaMaskHit(typ int, nx, ny, sx, sy float64) bool {
	m := b.nebulaMaskFor(typ)
	if m == nil {
		return false
	}
	px, py := int(sx-nx), int(sy-ny)
	if px < 0 || py < 0 || px >= m.w || py >= m.h {
		return false
	}
	if o := py*m.w + px; o >= 0 && o < len(m.idx) {
		return int(m.idx[o]) > nebulaInPixelThreshold
	}
	return false
}

// applyNebulaStarFlags 把「某個正規化座標是否落在星雲內」的判定式裝進對局。
//
// 規則層(`internal/shell`)不碰資產,所以遮罩判定只能從這裡提供。裝上之後 shell 會用它
// 重算每顆星的旗標,並在派遣時沿航線取樣判斷「有沒有穿過星雲」(見 shell/route.go)。
//
// ⚠ **開新局與讀檔後都要呼叫**:那個判定式是未匯出欄位,不進存檔。
func (b *sceneBuilder) applyNebulaStarFlags(sess *shell.GameSession) {
	if sess == nil {
		return
	}
	sess.SetNebulaProbe(func(x, y float64) bool {
		sx := starVX0 + x*(starVX1-starVX0)
		sy := starVY0 + y*(starVY1-starVY0)
		for _, n := range sess.Nebulae {
			nx, ny := nebulaScreenXY(n)
			if b.nebulaMaskHit(n.Type, nx, ny, sx, sy) {
				return true
			}
		}
		return false
	})
}

// nebulaStarTint 是「這顆星在星雲內」在星圖上的標示色。
//
// 原版沒有這個標示(玩家是用肉眼看星壓在雲上)。remake 加它的理由是:星雲圖被壓成半透明
// 之後邊界不明顯,而「在不在星雲內」直接決定打起來有沒有護盾——**看不出來的規則等於沒有規則**。
var nebulaStarTint = color.RGBA{170, 140, 220, 255}
