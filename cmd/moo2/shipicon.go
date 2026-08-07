package main

// shipicon.go:星圖上的**艦隊圖示**(原版 `Draw_Ship_Icons_` @ 0xA070F)。
//
// remake 先前在艦隊所在星畫一個 8×8 的青色方塊——那是佔位,不是原版的東西。
// 原版是一組帶旗色的小艦艇 sprite。
//
// ============ 星圖的圖層順序(`Draw_Main_Main_Screen_` @ 0x8440E)============
//
// 順帶把整張星圖的圖層抄下來,免得日後又要重找:
//
//	1. `Draw_Wormhole_Links_`   @ 0x85593
//	2. `Draw_Relocation_Links_` @ 0x85320
//	3. `Draw_Stars_`            @ 0x85550   → 逐星 `Draw_A_Star_` @ 0x83B02
//	4. `Draw_A_Gate_Icon_`      @ 0x83741(迴圈)
//	5. `Print_Star_Names_`      @ 0x88CB7
//	6. `Draw_Black_Holes_`      @ 0x83BF9
//	7. `Draw_Ship_Icons_`       @ 0xA070F   ← 這一支
//	8. `Print_Main_Screen_Data_`@ 0x87BAE
//	9. `Draw_Diplomacy_Request_Lights_` @ 0x83D06
//
// (外層 `Draw_Main_Screen_` 另有 `Draw_Nebulae_` @ 0x84F8F 與 `Draw_Paralax_` @ 0x8500F。)
// remake 目前只做了 3 與 7 的簡化版,其餘都還沒有——記在 gap report。
//
// ============ 圖檔:`Get_Ship_Icon_Pict_Seg_` @ 0xA0D78 ============
//
//	帝國艦隊(id 0..7):BUFFER0.LBX 資產 = 0xCD(205) + 旗色×4 + 縮放
//	第 9 類(id 8)  :              = 0xED(237) + 縮放
//	id 9..14        :              = 0xF1(241) + (id−9)×4 + 縮放
//
// 旗色取自玩家結構 +0x26(結構步距 0xEA9)。「×4」是因為原版星圖有**四段縮放**,
// 每段一張圖:實測 205..208 是 11×11 / 12×11 / 12×10 / 16×12,由小到大。
// 縮放值由 `sub_79917` 給,原版還把它反過來映射(3→0、2→1、1→2、0→3)——
// 也就是**縮得最遠 → 用最小的圖**。
//
// ⚠ remake 的星圖沒有縮放(固定全銀河檢視,對應原版縮得最遠那段),所以固定用縮放 0。
// 真的做出縮放時,這裡要跟著換,不是改圖。
//
// 每張圖有 **8 幀**(原版 `Cycle_Ship_Icons_` @ 0x82DFF 在跑動畫)。remake 目前只取第 0 幀,
// 靜態的——動畫沒做,寫在這裡免得被誤認為已完成。
//
// ⚠ 位置用 remake 自己算的星座標,不是原版的 `Get_Ship_Icon_Coords_` @ 0xA0A5C。
// 原版那支要有「艦隊在兩星之間航行到第幾格」的模型才有意義,remake 的艦隊移動是整段跳的。

import (
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

const (
	shipIconLBX = "buffer0.lbx"
	// 原版三段的資產基底(見檔頭)。remake 目前只用得到第一段。
	shipIconEmpireBase = 205
	shipIconPerFlag    = 4
	// remake 星圖固定全銀河檢視 = 原版縮得最遠那一段 → 縮放 0(最小的那張圖)。
	shipIconZoom = 0
)

// shipIconAsset 回傳某旗色的艦隊圖示資產編號。旗色超出 0..7 就夾回去——
// 原版的旗色本來就只有 8 種,越界代表上游有 bug,但畫面不該因此整個不見。
func shipIconAsset(flag int) int {
	if flag < 0 || flag >= len(shell.FlagColors) {
		flag = 0
	}
	return shipIconEmpireBase + flag*shipIconPerFlag + shipIconZoom
}

// shipIconImage 解出艦隊圖示並快取(取不到回 nil,呼叫端自己退回方塊)。
func (b *sceneBuilder) shipIconImage(flag int) *ebiten.Image {
	asset := shipIconAsset(flag)
	key := shipIconLBX + ":" + strconv.Itoa(asset)
	if im, hit := b.colBldgCache[key]; hit {
		return im
	}
	if b.colBldgCache == nil {
		b.colBldgCache = map[string]*ebiten.Image{}
	}
	var img *ebiten.Image
	if im, err := decodeAsset(b.res, shipIconLBX, asset); err == nil && len(im.Frames) > 0 {
		if pal, err := resolvePalette(b.res, im, paletteChain{{"buffer0.lbx", 0}}); err == nil {
			img = ebiten.NewImageFromImage(im.Frames[0].ToRGBADropTranslucent(pal, im.KeyColor()))
		}
	}
	b.colBldgCache[key] = img
	return img
}

// drawShipIconAt 把艦隊圖示畫在 (cx, cy) 置中。回傳有沒有畫成功——
// 沒有原版資產時(例如單元測試或資料夾不完整)呼叫端要自己退回舊的方塊標記。
func (b *sceneBuilder) drawShipIconAt(dst *ebiten.Image, flag, cx, cy int) bool {
	im := b.shipIconImage(flag)
	if im == nil {
		return false
	}
	w, h := im.Bounds().Dx(), im.Bounds().Dy()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(cx-w/2), float64(cy-h/2))
	dst.DrawImage(im, op)
	return true
}
