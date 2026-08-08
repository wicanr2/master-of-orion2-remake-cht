package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// confirmbox.go:原版的**是/否確認框**(`Confirmation_Box_` @ 0x77658)。
//
// remake 先前完全沒有 modal 對話框——所以每一條「原版會先問一句」的規則都只能寫在註解裡
// 說「remake 沒有這個基礎設施,直接允許」。這一檔把那個基礎設施補上。
//
// ============ 版面(全部取自反組譯,不是量圖)============
//
//	資產 `CONFIRM.LBX`:0 = 底框 313×227(自帶調色盤)、1 = Y 鈕 54×24(2 幀)、2 = N 鈕 55×24(2 幀)
//
//	`sub_12B7E1(eax=x, edx=y, ebx=img)` 三個貼圖點:
//	     底框  (0A1h, 75h)   = (161, 117)
//	     Y 鈕  (0EBh, 12Eh)  = (235, 302)
//	     N 鈕  (159h, 12Eh)  = (345, 302)
//
//	`sub_11438B` 建的兩塊熱區(eax=x1, edx=y1, ebx=x2, ecx=y2):
//	     Y  235..286 × 302..323   ← 51×21,比圖(54×24)小一圈
//	     N  345..396 × 302..323   ← 51×21
//
//	文字 `sub_77A74(eax=0CCh, edx=0D0h, ebx=0E0h, ecx=字串)`:
//	     左緣 x=204、垂直置中於 y=208、**折行寬度 224**
//	     (204 + 224/2 = 316 ≈ 底框中心 161 + 313/2 = 317.5,對得上)
//
//	`Draw_Confirm_Box_` @ 0x778E4 每幀把兩顆鈕的幀號歸 0,再把**游標所在**那顆設成 1
//	     → 第 1 幀是 hover 高亮,不是按下。
//
// ⚠ 沒有還原的:原版在文字放不下時會**縮字級**(`sub_103CAF` 量高度,var_C 從 4 遞減到 1,
// 直到高度 ≤ 126)。remake 的字型層沒有那組原版字級,改用固定字級 + 自行折行;
// 文字太長時寧可截斷也不假裝有那條規則。

const (
	confirmLBX = "confirm.lbx"

	confirmBoxAsset = 0
	confirmBoxX     = 161
	confirmBoxY     = 117

	confirmYesAsset = 1
	confirmYesX     = 235
	confirmYesY     = 302
	confirmNoAsset  = 2
	confirmNoX      = 345
	confirmNoY      = 302

	// 熱區比圖小一圈,取反組譯的 x1..x2 / y1..y2。
	confirmBtnW = 51
	confirmBtnH = 21

	// 文字塊:左緣、折行寬度、垂直置中線。
	confirmTextX    = 204
	confirmTextW    = 224
	confirmTextMidY = 208
	confirmTextSize = 12
	confirmLineStep = 16
)

// confirmScreen 是一個是/否確認框,疊在 under 這個畫面上。
//
// onYes/onNo 回傳接下來要去哪個畫面(nil = 留在 under)。
type confirmScreen struct {
	b     *sceneBuilder
	under origScreen // 下層畫面(原版是疊在星圖上,所以要先把它畫出來)
	msg   string
	onYes func() *origTransition
	onNo  func() *origTransition

	bg, yes, no *ebiten.Image
	yesHi, noHi *ebiten.Image
	// yesFace/noFace 是兩顆鈕的面色(採樣自圖本身)。中文模式要先用它把烘死的
	// YES / NO 擦掉再疊中文——不擦的話兩層字會疊在一起,那是先前 loadgame/gamemenu
	// 兩個畫面就處理過的同一件事。
	yesFace, noFace color.RGBA
}

// confirmBtnFace 採樣按鈕圖的面色(取左緣內縮 6px、垂直中線那一點)。
//
// 取內縮而不是角落:角落是邊框的高光/陰影,拿來擦底會留下一圈色差。
func confirmBtnFace(im *ebiten.Image) color.RGBA {
	if im == nil {
		return color.RGBA{72, 76, 84, 255}
	}
	b := im.Bounds()
	if b.Dx() <= 8 || b.Dy() <= 6 {
		return color.RGBA{72, 76, 84, 255}
	}
	r, g, bl, a := im.At(b.Min.X+6, b.Min.Y+b.Dy()/2).RGBA()
	if a == 0 {
		return color.RGBA{72, 76, 84, 255}
	}
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), 255}
}

// confirmImage 取 CONFIRM.LBX 的某資產某幀(調色盤取資產 0 自帶的那份)。
func (b *sceneBuilder) confirmImage(assetID, frame int, keyColor bool) *ebiten.Image {
	prov, err := decodeAsset(b.res, confirmLBX, confirmBoxAsset)
	if err != nil || prov.Embedded == nil {
		return nil
	}
	im, err := decodeAsset(b.res, confirmLBX, assetID)
	if err != nil || frame >= len(im.Frames) {
		return nil
	}
	return ebiten.NewImageFromImage(im.Frames[frame].ToRGBA(prov.Embedded, keyColor))
}

// confirm 疊一個是/否確認框在目前的畫面上。
func (b *sceneBuilder) confirm(under origScreen, msg string, onYes, onNo func() *origTransition) *confirmScreen {
	s := &confirmScreen{b: b, under: under, msg: msg, onYes: onYes, onNo: onNo}
	s.bg = b.confirmImage(confirmBoxAsset, 0, false)
	s.yes = b.confirmImage(confirmYesAsset, 0, true)
	s.yesHi = b.confirmImage(confirmYesAsset, 1, true)
	s.no = b.confirmImage(confirmNoAsset, 0, true)
	s.noHi = b.confirmImage(confirmNoAsset, 1, true)
	s.yesFace, s.noFace = confirmBtnFace(s.yes), confirmBtnFace(s.no)
	return s
}

func (s *confirmScreen) update(in shell.InputState) *origTransition {
	if !in.ClickReleased {
		return nil
	}
	switch {
	case hitBox(in.MouseX, in.MouseY, confirmYesX, confirmYesY, confirmBtnW, confirmBtnH):
		if clickSound != nil {
			clickSound()
		}
		if s.onYes != nil {
			if t := s.onYes(); t != nil {
				return t
			}
		}
		return &origTransition{next: s.under}
	case hitBox(in.MouseX, in.MouseY, confirmNoX, confirmNoY, confirmBtnW, confirmBtnH):
		if clickSound != nil {
			clickSound()
		}
		if s.onNo != nil {
			if t := s.onNo(); t != nil {
				return t
			}
		}
		return &origTransition{next: s.under}
	}
	// 框外點擊什麼都不做——modal 的重點就是它擋住下層。
	return nil
}

func (s *confirmScreen) draw(dst *ebiten.Image) {
	if s.under != nil {
		s.under.draw(dst) // 原版的確認框是疊在星圖上的,下層要看得見
	} else {
		dst.Fill(color.RGBA{8, 10, 16, 255})
	}
	if s.bg != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(confirmBoxX, confirmBoxY)
		drawPanelImage(dst, s.bg, op)
	}
	mx, my := ebiten.CursorPosition()
	drawBtn := func(img, hi *ebiten.Image, x, y int) {
		use := img
		if hitBox(mx, my, x, y, confirmBtnW, confirmBtnH) && hi != nil {
			use = hi // Draw_Confirm_Box_ 每幀把游標所在的那顆設成第 1 幀
		}
		if use == nil {
			return
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		drawPanelImage(dst, use, op)
	}
	drawBtn(s.yes, s.yesHi, confirmYesX, confirmYesY)
	drawBtn(s.no, s.noHi, confirmNoX, confirmNoY)

	if s.b.fnt == nil {
		return
	}
	// 中文模式:先擦掉烘在圖上的 YES / NO 再疊中文(同 loadgame/gamemenu 的做法)。
	// 英文模式讓路,露原版烘的字。
	if s.b.lang == i18n.Traditional {
		ink := color.RGBA{235, 240, 250, 255}
		label := func(txt string, x, y int, face color.RGBA) {
			fillPanel(dst, float32(x+4), float32(y+4),
				float32(confirmBtnW-8), float32(confirmBtnH-8), face, false)
			s.b.fnt.DrawCentered(dst, txt, float64(x+confirmBtnW/2), float64(y+confirmBtnH/2)+4, 12, ink)
		}
		label("是", confirmYesX, confirmYesY, s.yesFace)
		label("否", confirmNoX, confirmNoY, s.noFace)
	}
	lines := wrapToWidth(s.b.fnt, s.msg, confirmTextSize, confirmTextW)
	top := float64(confirmTextMidY) - float64(len(lines)-1)*confirmLineStep/2
	for i, ln := range lines {
		s.b.fnt.DrawCentered(dst, ln, float64(confirmTextX+confirmTextW/2), top+float64(i*confirmLineStep),
			confirmTextSize, color.RGBA{230, 224, 200, 255})
	}
}
