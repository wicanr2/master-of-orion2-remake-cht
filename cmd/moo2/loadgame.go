package main

import (
	"fmt"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// loadgame.go:載入遊戲視窗(原版 `Mainmenu_Load_Game_Popup_` @ 0x804B7)。
//
// remake 先前根本沒有這個畫面:主選單的 Continue 與 Load Game 都是「讀那唯一一個存檔檔案」,
// 沒有存檔時點下去靜默無反應。2026-07-12 用 archive.org 線上原版對照時,使用者回報的
// issue #2 就是這件事——原版**無存檔時 Continue / Load Game 是灰階不可按的**,有存檔時
// Load Game 會開一個十格的存檔選擇視窗。
//
// ============ 版面:openorion2 `LoadGameWindow::initWidgets`(它自己是從原版 RE 出來的)============
//
//	背景        game.lbx 資產 20,調色盤取自 mainmenu.lbx 資產 21
//	視窗定位    _x = (640 − 寬)/2、_y = (480 − 高)/2   ← 置中,不是固定座標
//	LOAD 鈕     (37, 337),圖 game.lbx 資產 21
//	CANCEL 鈕   (171, 338),68×22,圖 game.lbx 資產 22
//	存檔槽      10 格,第 i 格 (22, 22 + 31×i),232×27
//	槽內文字    (_x + 32, _y + 24 + 31×i)
//	最後一格    固定是自動存檔(`if (slot == SAVEGAME_SLOTS - 1)` → ESTR_SAVESLOT_AUTO)
//
// 上面除了「置中」是公式,其餘都是硬編立即數,一個都沒有估。槽位規格(10 格、末格自動存檔)
// 見 internal/shell/saveslots.go 檔頭。
//
// ⚠ 誠實留白:原版這個視窗右側還會依存檔類型畫單人/熱座/網路/數據機四種圖示
// (`ASSET_LOAD_SINGLE`=23 / `HOTSEAT`=24 / `NETWORK`=25 / `MODEM`=26)。remake 只有單人對局,
// 四種圖示只會有一種,畫上去等於裝飾,故不畫——等多人連線子專案落地再補。

// 原版載入視窗的資產與座標(openorion2 mainmenu.cpp)。
const (
	loadWinLBX      = "game.lbx"
	loadWinBGAsset  = 20 // ASSET_LOAD_BACKGROUND
	loadWinBtnAsset = 21 // ASSET_LOAD_LOADBUTTON
	loadWinCanAsset = 22 // ASSET_LOAD_CANCEL

	loadWinPalLBX   = "mainmenu.lbx"
	loadWinPalAsset = 21 // ASSET_MENU_BACKGROUND(主選單背景提供調色盤)

	loadBtnX, loadBtnY       = 37, 337
	loadCanX, loadCanY       = 171, 338
	loadCanW, loadCanH       = 68, 22
	loadSlotX, loadSlotY0    = 22, 22
	loadSlotW, loadSlotH     = 232, 27
	loadSlotStep             = 31
	loadSlotTextX, loadTextY = 32, 24 // 槽內文字相對視窗左上的偏移
)

// loadGameScreen 是載入遊戲視窗。
type loadGameScreen struct {
	b   *sceneBuilder
	fnt *uifont.Font

	bg, loadBtn, cancelBtn *ebiten.Image
	// loadFace / cancelFace 是兩顆鈕的「面色」,用來擦掉烘在圖上的英文(LOAD / CANCEL)
	// 再疊中文——與 overlayScreen 的擦底疊字同一套手法,只是這裡是自繪畫面,得自己採樣。
	loadFace, cancelFace   color.RGBA
	winX, winY, winW, winH int

	slots    []shell.SaveSlotInfo
	selected int // -1 = 未選
	msg      string
}

// loadWinImage 取 game.lbx 的某資產(調色盤借主選單背景,見檔頭)。
// 一併回傳「面色」:從按鈕左緣往內數第 4 個像素、垂直置中處採樣,那裡是純按鈕面、
// 不會踩到烘上去的英文,拿來當擦底色。失敗回 nil + 零值。
func loadWinImage(res *assets.Resolver, assetID int, keyColor bool) (*ebiten.Image, color.RGBA) {
	prov, err := decodeAsset(res, loadWinPalLBX, loadWinPalAsset)
	if err != nil || prov.Embedded == nil {
		return nil, color.RGBA{}
	}
	im, err := decodeAsset(res, loadWinLBX, assetID)
	if err != nil || len(im.Frames) == 0 {
		return nil, color.RGBA{}
	}
	rgba := im.Frames[0].ToRGBA(prov.Embedded, keyColor)
	b := rgba.Bounds()
	face := color.RGBA{60, 64, 72, 255} // 採樣失敗時的保守深灰
	if b.Dx() > 8 && b.Dy() > 4 {
		r, g, bl, a := rgba.At(b.Min.X+4, b.Min.Y+b.Dy()/2).RGBA()
		if a > 0 {
			face = color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), 255}
		}
	}
	return ebiten.NewImageFromImage(rgba), face
}

func newLoadGameScreen(b *sceneBuilder) *loadGameScreen {
	s := &loadGameScreen{b: b, fnt: b.fnt, selected: -1}
	s.bg, _ = loadWinImage(b.res, loadWinBGAsset, false)
	s.loadBtn, s.loadFace = loadWinImage(b.res, loadWinBtnAsset, true)
	s.cancelBtn, s.cancelFace = loadWinImage(b.res, loadWinCanAsset, true)
	// 視窗置中(openorion2:_x = (SCREEN_WIDTH − _width) / 2)。
	s.winW, s.winH = 276, 376 // 取不到背景時的退路尺寸(實測 game.lbx#20 為此尺寸)
	if s.bg != nil {
		s.winW, s.winH = s.bg.Bounds().Dx(), s.bg.Bounds().Dy()
	}
	s.winX, s.winY = (moo2ScreenW-s.winW)/2, (moo2ScreenH-s.winH)/2
	s.slots = shell.ReadSaveSlots(saveDirFor())
	// 預選最近的存檔——十格都要玩家自己找太煩,而 Continue 本來就是接續最近進度。
	if latest := shell.LatestSaveSlot(saveDirFor()); latest >= 0 {
		s.selected = latest
	}
	return s
}

// slotRect 回傳第 i 格的螢幕座標(openorion2:createWidget(22, 22 + 31*i, 232, 27),
// 座標相對視窗左上)。
func (s *loadGameScreen) slotRect(i int) (int, int, int, int) {
	return s.winX + loadSlotX, s.winY + loadSlotY0 + loadSlotStep*i, loadSlotW, loadSlotH
}

func (s *loadGameScreen) loadRect() (int, int, int, int) {
	w, h := loadCanW, loadCanH // LOAD 鈕沒有明寫尺寸(用圖的),取不到圖時沿用 CANCEL 的
	if s.loadBtn != nil {
		w, h = s.loadBtn.Bounds().Dx(), s.loadBtn.Bounds().Dy()
	}
	return s.winX + loadBtnX, s.winY + loadBtnY, w, h
}

func (s *loadGameScreen) cancelRect() (int, int, int, int) {
	return s.winX + loadCanX, s.winY + loadCanY, loadCanW, loadCanH
}

func (s *loadGameScreen) update(in shell.InputState) *origTransition {
	if !in.ClickReleased {
		return nil
	}
	if x, y, w, h := s.cancelRect(); hitRect(in, x, y, w, h) {
		return s.b.goTo(s.b.menu, "主選單")
	}
	if x, y, w, h := s.loadRect(); hitRect(in, x, y, w, h) {
		return s.doLoad()
	}
	for i := range s.slots {
		if x, y, w, h := s.slotRect(i); hitRect(in, x, y, w, h) {
			if !s.slots[i].Exists {
				s.msg = "這一格是空的"
				return nil
			}
			s.selected, s.msg = i, ""
			return nil
		}
	}
	return nil
}

// doLoad 讀取選定的槽並進星系主畫面。
func (s *loadGameScreen) doLoad() *origTransition {
	if s.selected < 0 || s.selected >= len(s.slots) || !s.slots[s.selected].Exists {
		s.msg = "請先選一格有存檔的格子"
		return nil
	}
	gs, err := shell.LoadSession(s.slots[s.selected].Path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "讀檔失敗:", err)
		s.msg = "這個存檔讀不起來(格式可能來自舊版本)"
		return nil
	}
	s.b.session = gs
	if len(s.b.herodataMercs) > 0 {
		s.b.session.SetMercCandidates(s.b.herodataMercs) // 讀檔建的是新 session,重注入真英雄池
	}
	// 之後的自動存檔要寫回同一格,不然玩家讀了第 3 格、存檔卻跑去覆蓋自動存檔格。
	s.b.savePath = s.slots[s.selected].Path
	return s.b.goTo(s.b.galaxy, "星系主畫面")
}

func (s *loadGameScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{6, 8, 14, 255})
	if s.bg != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(s.winX), float64(s.winY))
		dst.DrawImage(s.bg, op)
	}
	if s.fnt == nil {
		return
	}
	body := color.RGBA{206, 214, 232, 255}
	sel := color.RGBA{250, 226, 130, 255}
	dim := color.RGBA{120, 126, 140, 255}

	for i, sl := range s.slots {
		col := body
		if i == s.selected {
			col = sel
			x, ry, w, h := s.slotRect(i)
			vector.StrokeRect(dst, float32(x), float32(ry), float32(w), float32(h), 1, sel, false)
		}
		// 文字基準取自 openorion2 drawSlot:_x + 32、_y + 24 + 31*slot。
		tx := float64(s.winX + loadSlotTextX)
		ty := float64(s.winY + loadTextY + loadSlotStep*i)
		if !sl.Exists {
			label := "（空）"
			if sl.Auto {
				label = "（自動存檔：尚無）"
			}
			s.fnt.Draw(dst, label, tx, ty, 12, dim)
			continue
		}
		name := sl.Empire
		if name == "" {
			name = "無名帝國" // 舊存檔沒存帝國名(見 persist.go 該欄位註解)
		}
		title := fmt.Sprintf("%d. %s", i+1, name)
		if sl.Auto {
			title = fmt.Sprintf("自動  %s", name)
		}
		s.fnt.Draw(dst, title, tx, ty, 12, col)
		s.fnt.Draw(dst, fmt.Sprintf("星曆 %s", sl.Stardate), tx+140, ty, 11, col)
	}

	if s.loadBtn != nil {
		x, y, _, _ := s.loadRect()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		dst.DrawImage(s.loadBtn, op)
	}
	if s.cancelBtn != nil {
		x, y, _, _ := s.cancelRect()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		dst.DrawImage(s.cancelBtn, op)
	}
	// 兩顆鈕的英文(LOAD / CANCEL)烘在圖上,照既有做法擦底疊中文:先用採樣到的面色
	// 填掉文字帶(上下左右各留 3px 保住浮雕邊框),再疊中文。
	drawBtnLabel := func(x, y, w, h int, face color.RGBA, zh string) {
		if face.A > 0 {
			vector.DrawFilledRect(dst, float32(x+3), float32(y+3), float32(w-6), float32(h-6), face, false)
		}
		s.fnt.DrawCentered(dst, zh, float64(x+w/2), float64(y+h/2), 13, body)
	}
	lx, ly, lw, lh := s.loadRect()
	drawBtnLabel(lx, ly, lw, lh, s.loadFace, "載入")
	cx, cy, cw, ch := s.cancelRect()
	drawBtnLabel(cx, cy, cw, ch, s.cancelFace, "取消")

	if s.msg != "" {
		s.fnt.DrawCentered(dst, s.msg, 320, float64(s.winY+s.winH+14), 12,
			color.RGBA{235, 160, 120, 255})
	}
}

// loadGame 進入載入遊戲視窗。
func (b *sceneBuilder) loadGame() (origScreen, error) { return newLoadGameScreen(b), nil }
