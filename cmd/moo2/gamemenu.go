package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// gamemenu.go:遊戲中的「遊戲」選單視窗(原版 `GameMenuWindow`)。
//
// 星系主畫面頂端那條「遊戲」按鈕先前是**死的**——畫得出來但點了沒事。它是原版通往
// 存檔/讀檔/開新局/離開的唯一入口,少了它,玩家在遊戲中沒有任何辦法主動存檔。
//
// ============ 版面:openorion2 `GameMenuWindow`(它自己是從原版 RE 出來的)============
//
//	背景        game.lbx 資產 0,調色盤取 buffer0.lbx 資產 0(GALAXY_ARCHIVE / ASSET_GALAXY_GUI)
//	視窗位置    (144, 25)  ← **硬編,不是置中**
//	SAVE GAME   (40, 43)    圖 game.lbx 資產 1
//	LOAD GAME   (147, 43)   圖 game.lbx 資產 2
//	NEW GAME    (40, 88)    圖 game.lbx 資產 3
//	QUIT GAME   (147, 88)   圖 game.lbx 資產 4
//	SETTINGS    (40, 307)   圖 game.lbx 資產 5
//	RETURN      (151, 307)  圖 game.lbx 資產 6
//	星系畫面上的「遊戲」鈕本身 (249, 5),圖 buffer0.lbx 資產 1
//
// 按鈕座標相對視窗左上,與 `LoadGameWindow` 同一套慣例。
//
// ⚠ 誠實留白:原版這個視窗中段有 Music / Sound Fx 兩條音量滑桿(背景圖上畫得很清楚),
// remake 的音訊層目前沒有音量控制介面,滑桿不畫也不接——畫一條拖不動的滑桿比沒有更糟。
// SETTINGS 同理:原版另有一整個設定畫面,remake 尚無對應內容,按鈕保留但不接。

// 原版遊戲選單視窗的資產與座標(openorion2 galaxy.cpp GameMenuWindow)。
const (
	gameMenuLBX     = "game.lbx"
	gameMenuPalLBX  = "buffer0.lbx"
	gameMenuPalAsst = 0

	gameMenuBGAsset = 0
	gameMenuWinX    = 144
	gameMenuWinY    = 25
)

// gameMenuButton 是視窗上的一顆鈕:相對視窗左上的座標 + 精靈資產 + 中文字 + 動作。
type gameMenuButton struct {
	x, y   int
	asset  int
	label  string
	action string
}

// 六顆鈕的座標與資產全部來自 openorion2 `GameMenuWindow::initWidgets`。
var gameMenuButtons = []gameMenuButton{
	{40, 43, 1, "儲存遊戲", "save"},
	{147, 43, 2, "載入遊戲", "load"},
	{40, 88, 3, "開新遊戲", "new"},
	{147, 88, 4, "離開遊戲", "quit"},
	{40, 307, 5, "設定", "settings"},
	{151, 307, 6, "返回", "return"},
}

// gameMenuScreen 是遊戲中的「遊戲」選單視窗。
type gameMenuScreen struct {
	b   *sceneBuilder
	fnt *uifont.Font

	bg                     *ebiten.Image
	btnImg                 []*ebiten.Image
	btnFace                []color.RGBA
	winX, winY, winW, winH int
	msg                    string
}

// gameMenuImage 取 game.lbx 的某資產(調色盤借 buffer0#0,同星系主畫面那條鏈),
// 一併回傳按鈕面色供擦底疊字用(採樣點見 loadgame.go 的 loadWinImage)。
func (b *sceneBuilder) gameMenuImage(assetID int, keyColor bool) (*ebiten.Image, color.RGBA) {
	prov, err := decodeAsset(b.res, gameMenuPalLBX, gameMenuPalAsst)
	if err != nil || prov.Embedded == nil {
		return nil, color.RGBA{}
	}
	im, err := decodeAsset(b.res, gameMenuLBX, assetID)
	if err != nil || len(im.Frames) == 0 {
		return nil, color.RGBA{}
	}
	rgba := im.Frames[0].ToRGBA(prov.Embedded, keyColor)
	bounds := rgba.Bounds()
	face := color.RGBA{60, 64, 72, 255}
	if bounds.Dx() > 8 && bounds.Dy() > 4 {
		r, g, bl, a := rgba.At(bounds.Min.X+4, bounds.Min.Y+bounds.Dy()/2).RGBA()
		if a > 0 {
			face = color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), 255}
		}
	}
	return ebiten.NewImageFromImage(rgba), face
}

func newGameMenuScreen(b *sceneBuilder) *gameMenuScreen {
	s := &gameMenuScreen{b: b, fnt: b.fnt}
	s.bg, _ = b.gameMenuImage(gameMenuBGAsset, false)
	s.winX, s.winY = gameMenuWinX, gameMenuWinY
	s.winW, s.winH = 276, 376 // 取不到背景時的退路
	if s.bg != nil {
		s.winW, s.winH = s.bg.Bounds().Dx(), s.bg.Bounds().Dy()
	}
	s.btnImg = make([]*ebiten.Image, len(gameMenuButtons))
	s.btnFace = make([]color.RGBA, len(gameMenuButtons))
	for i, btn := range gameMenuButtons {
		s.btnImg[i], s.btnFace[i] = b.gameMenuImage(btn.asset, true)
	}
	return s
}

// btnRect 回傳第 i 顆鈕的螢幕座標。尺寸取自精靈圖;取不到就用一個保守值,
// 免得整排鈕變成點不到的零寬熱區。
func (s *gameMenuScreen) btnRect(i int) (int, int, int, int) {
	w, h := 88, 26
	if s.btnImg[i] != nil {
		w, h = s.btnImg[i].Bounds().Dx(), s.btnImg[i].Bounds().Dy()
	}
	return s.winX + gameMenuButtons[i].x, s.winY + gameMenuButtons[i].y, w, h
}

func (s *gameMenuScreen) update(in shell.InputState) *origTransition {
	if !in.ClickReleased {
		return nil
	}
	for i, btn := range gameMenuButtons {
		x, y, w, h := s.btnRect(i)
		if !hitRect(in, x, y, w, h) {
			continue
		}
		switch btn.action {
		case "return":
			return s.b.goTo(s.b.galaxy, "星系主畫面")
		case "save":
			sc, err := s.b.saveGameInPlay()
			if err == nil {
				return &origTransition{next: sc}
			}
		case "load":
			if !shell.AnySaveExists(saveDirFor()) {
				s.msg = "還沒有任何存檔"
				return nil
			}
			sc, err := s.b.loadGameInPlay()
			if err == nil {
				return &origTransition{next: sc}
			}
		case "new":
			return s.b.goTo(s.b.newGameSetup, "新遊戲設定")
		case "quit":
			return s.b.goTo(s.b.menu, "主選單") // 原版是回主選單,不是直接關程式
		case "settings":
			// 原版另有一整個設定畫面,remake 尚無對應內容(見檔頭留白)。
			s.msg = "設定畫面尚未建置"
		}
		return nil
	}
	return nil
}

func (s *gameMenuScreen) draw(dst *ebiten.Image) {
	// 這是疊在星系主畫面上的視窗;remake 沒有「保留下層畫面」的機制,用深色底代替。
	dst.Fill(color.RGBA{8, 10, 16, 255})
	if s.bg != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(s.winX), float64(s.winY))
		dst.DrawImage(s.bg, op)
	}
	if s.fnt == nil {
		return
	}
	body := color.RGBA{212, 220, 236, 255}
	for i, btn := range gameMenuButtons {
		x, y, w, h := s.btnRect(i)
		if s.btnImg[i] != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			dst.DrawImage(s.btnImg[i], op)
		}
		// 英文烘在鈕上,擦底疊中文(同 loadgame.go 的做法)。
		// 英文模式跳過:GAME.LBX 那六顆鈕上本來就是 SAVE GAME / LOAD GAME / …。
		if s.b.lang != i18n.Traditional {
			continue
		}
		if s.btnFace[i].A > 0 {
			vector.DrawFilledRect(dst, float32(x+3), float32(y+3), float32(w-6), float32(h-6),
				s.btnFace[i], false)
		}
		s.fnt.DrawCentered(dst, btn.label, float64(x+w/2), float64(y+h/2), 12, body)
	}
	if s.msg != "" {
		s.fnt.DrawCentered(dst, s.msg, 320, float64(s.winY+s.winH+14), 12,
			color.RGBA{235, 160, 120, 255})
	}
}

// gameMenu 進入遊戲選單視窗。
func (b *sceneBuilder) gameMenu() (origScreen, error) { return newGameMenuScreen(b), nil }
