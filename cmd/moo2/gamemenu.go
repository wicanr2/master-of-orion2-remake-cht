package main

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	moo2audio "github.com/wicanr2/master-of-orion2-remake-cht/internal/audio"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// gamemenu.go:遊戲中的「遊戲」選單視窗(原版 `Do_Main_Game_Popup_`
// @ 0x7DD41 / `_Draw_Main_Game_Popup_` @ 0x7F701)。
//
// 星系主畫面頂端那條「遊戲」按鈕先前是**死的**——畫得出來但點了沒事。它是原版通往
// 存檔/讀檔/開新局/離開的唯一入口,少了它,玩家在遊戲中沒有任何辦法主動存檔。
//
// ============ 版面:openorion2 `MainMenuWindow`(它自己是從原版 RE 出來的)============
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
// 音量條:原始 GAME.LBX 資產 7 是 155×12 的音量條貼圖;手冊說明按下後可拖曳,
// 且靠左會關閉音量。remake 以同一條貼圖做動態裁切,把音訊層的即時音量接回這個視窗。
//
// SETTINGS 進入 gamesettings.go 的原版 13 列分頁；偏好由 shell.GameSettings 保存。

// 原版遊戲選單視窗的資產與座標(openorion2 galaxy.cpp MainMenuWindow)。
const (
	gameMenuLBX     = "game.lbx"
	gameMenuPalLBX  = "buffer0.lbx"
	gameMenuPalAsst = 0

	gameMenuBGAsset     = 0
	gameMenuSliderAsset = 7
	gameMenuWinX        = 144
	gameMenuWinY        = 25
	gameMenuSliderX     = 61
	gameMenuSliderW     = 155
	gameMenuSliderH     = 12
	gameMenuMusicY      = 170
	gameMenuSFXY        = 195
)

// gameMenuButton 是視窗上的一顆鈕:相對視窗左上的座標 + 精靈資產 + 外部文案鍵 + 動作。
type gameMenuButton struct {
	x, y    int
	asset   int
	textKey string
	action  string
}

// 六顆鈕的座標與資產全部來自 openorion2 `MainMenuWindow::initWidgets`。
var gameMenuButtons = []gameMenuButton{
	{40, 43, 1, "gamemenu.button.save", "save"},
	{147, 43, 2, "gamemenu.button.load", "load"},
	{40, 88, 3, "gamemenu.button.new", "new"},
	{147, 88, 4, "gamemenu.button.quit", "quit"},
	{40, 307, 5, "gamemenu.button.settings", "settings"},
	{151, 307, 6, "gamemenu.button.return", "return"},
}

// gameMenuScreen 是遊戲中的「遊戲」選單視窗。
type gameMenuScreen struct {
	b   *sceneBuilder
	fnt *uifont.Font

	bg                     *ebiten.Image
	slider                 *ebiten.Image
	btnImg                 []*ebiten.Image
	btnFace                []color.RGBA
	winX, winY, winW, winH int
	msg                    string
	musicVolume            float64
	sfxVolume              float64
	dragSlider             int // 0=Music, 1=Sound Fx, -1=沒有拖曳
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
	musicVolume, sfxVolume := moo2audio.DefaultBGMVolume, moo2audio.DefaultSFXVolume
	if theMixer != nil {
		musicVolume, sfxVolume = theMixer.Volumes()
	}
	s := &gameMenuScreen{b: b, fnt: b.fnt, musicVolume: musicVolume, sfxVolume: sfxVolume, dragSlider: -1}
	s.bg, _ = b.gameMenuImage(gameMenuBGAsset, false)
	s.slider, _ = b.gameMenuImage(gameMenuSliderAsset, true)
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

func (s *gameMenuScreen) buttonTextRect(i int) textSafeRect {
	x, y, w, h := s.btnRect(i)
	return textSafeRect{x: x, y: y, w: w, h: h, insetX: 4, insetY: 3}
}

func (s *gameMenuScreen) volumeLabelTextRect(which int) textSafeRect {
	y := s.winY + 128
	if which == 1 {
		y = s.winY + 205
	}
	return textSafeRect{x: s.winX + gameMenuSliderX - 2, y: y, w: 170, h: 31, insetX: 2, insetY: 2}
}

func (s *gameMenuScreen) messageTextRect() textSafeRect {
	y := s.winY + s.winH
	h := moo2ScreenH - y
	if h > 30 {
		h = 30
	}
	return textSafeRect{x: s.winX, y: y, w: s.winW, h: h, insetX: 4, insetY: 2}
}

func (s *gameMenuScreen) sliderRect(which int) (int, int, int, int) {
	y := gameMenuMusicY
	if which == 1 {
		y = gameMenuSFXY
	}
	return s.winX + gameMenuSliderX, s.winY + y, gameMenuSliderW, gameMenuSliderH
}

func sliderVolumeAt(mouseX, x, w int) float64 {
	if mouseX <= x {
		return 0
	}
	if mouseX >= x+w-1 {
		return 1
	}
	return moo2audio.ClampVolume(float64(mouseX-x) / float64(w-1))
}

func (s *gameMenuScreen) setSlider(which, mouseX int) {
	x, _, w, _ := s.sliderRect(which)
	v := sliderVolumeAt(mouseX, x, w)
	if which == 0 {
		s.musicVolume = v
	} else {
		s.sfxVolume = v
	}
	if theMixer != nil {
		theMixer.SetVolumes(s.musicVolume, s.sfxVolume)
	}
}

func (s *gameMenuScreen) update(in shell.InputState) *origTransition {
	// 先處理音量條。實際滑鼠按住時每幀更新;測試/截圖腳本只送 ClickReleased
	// 也能以單次點擊設定位置。
	if s.dragSlider >= 0 {
		if in.MouseDown || in.ClickReleased {
			s.setSlider(s.dragSlider, in.MouseX)
		}
		if in.ClickReleased {
			s.dragSlider = -1
		}
		return nil
	}
	for i := 0; i < 2; i++ {
		x, y, w, h := s.sliderRect(i)
		if !hitRect(in, x, y, w, h) || (!in.MouseDown && !in.ClickReleased) {
			continue
		}
		s.setSlider(i, in.MouseX)
		if in.MouseDown {
			s.dragSlider = i
		}
		return nil
	}
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
			return s.b.goTo(s.b.galaxy, uiText(s.b.lang, "gamemenu.transition.galaxy"))
		case "save":
			sc, err := s.b.saveGameInPlay()
			if err == nil {
				return &origTransition{next: sc}
			}
		case "load":
			if !shell.AnySaveExists(saveDirFor()) {
				s.msg = uiText(s.b.lang, "gamemenu.message.no_saves")
				return nil
			}
			sc, err := s.b.loadGameInPlay()
			if err == nil {
				return &origTransition{next: sc}
			}
		case "new":
			return s.b.goTo(s.b.newGameSetup, uiText(s.b.lang, "gamemenu.transition.new_game"))
		case "quit":
			return s.b.goTo(s.b.menu, uiText(s.b.lang, "gamemenu.transition.main_menu")) // 原版是回主選單,不是直接關程式
		case "settings":
			return &origTransition{next: newGameSettingsScreen(s.b)}
		}
		return nil
	}
	return nil
}

// drawVolumeSlider 擦掉背景圖中固定的音量條,再以原始資產 7 裁出目前音量。
// 這樣英文模式保留原版標籤,中文模式則可同時覆蓋烘進去的英文標籤。
func (s *gameMenuScreen) drawVolumeSlider(dst *ebiten.Image, which int, volume float64) {
	x, y, w, h := s.sliderRect(which)
	fillPanel(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{0, 0, 0, 255}, false)
	filled := int(moo2audio.ClampVolume(volume) * float64(w))
	if filled > 0 && s.slider != nil {
		if filled > w {
			filled = w
		}
		part := s.slider.SubImage(image.Rect(0, 0, filled, h)).(*ebiten.Image)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		drawPanelImage(dst, part, op)
		return
	}
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1,
		color.RGBA{100, 130, 180, 255}, false)
}

func (s *gameMenuScreen) draw(dst *ebiten.Image) {
	// 這是疊在星系主畫面上的視窗;remake 沒有「保留下層畫面」的機制,用深色底代替。
	dst.Fill(color.RGBA{8, 10, 16, 255})
	if s.bg != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(s.winX), float64(s.winY))
		drawPanelImage(dst, s.bg, op)
	} else {
		fillPanel(dst, float32(s.winX), float32(s.winY), float32(s.winW), float32(s.winH),
			color.RGBA{12, 18, 30, 250}, false)
		vector.StrokeRect(dst, float32(s.winX), float32(s.winY), float32(s.winW), float32(s.winH), 1,
			color.RGBA{90, 120, 170, 255}, false)
	}
	if s.fnt == nil {
		return
	}
	// 音量標籤改由 JSON 統一供應。原圖的 Sound Fx 是烘字,Music 則無法在現行圖上
	// 穩定讀取;兩種語言都先擦底再疊字,避免一個欄位來自圖、另一個欄位消失。
	if s.bg != nil {
		// 黑底正是原始音量面板的底色。
		fillPanel(dst, float32(s.winX+gameMenuSliderX-2), float32(s.winY+132), 170, 31,
			color.RGBA{0, 0, 0, 255}, false)
		// Sound Fx 的原版字樣位在第二條滑桿下方;多擦一點垂直範圍，避免英文殘影
		// 從點陣字的下緣漏出來，同時仍留在中央黑色面板內。
		fillPanel(dst, float32(s.winX+gameMenuSliderX-2), float32(s.winY+209), 170, 58,
			color.RGBA{0, 0, 0, 255}, false)
	}
	labelColor := color.RGBA{190, 198, 214, 255}
	s.volumeLabelTextRect(0).drawLeft(dst, s.fnt, uiText(s.b.lang, "gamemenu.label.music"), 18, labelColor)
	s.volumeLabelTextRect(1).drawLeft(dst, s.fnt, uiText(s.b.lang, "gamemenu.label.sound_fx"), 18, labelColor)
	s.drawVolumeSlider(dst, 0, s.musicVolume)
	s.drawVolumeSlider(dst, 1, s.sfxVolume)
	body := color.RGBA{212, 220, 236, 255}
	for i, btn := range gameMenuButtons {
		x, y, w, h := s.btnRect(i)
		if s.btnImg[i] != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			drawPanelImage(dst, s.btnImg[i], op)
		} else {
			fillPanel(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{46, 52, 66, 255}, false)
			vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1,
				color.RGBA{110, 130, 170, 255}, false)
		}
		// 英文烘在鈕上,擦底疊中文(同 loadgame.go 的做法)。
		// 英文模式跳過:GAME.LBX 那六顆鈕上本來就是 SAVE GAME / LOAD GAME / …。
		if s.b.lang != i18n.Traditional && s.btnImg[i] != nil {
			continue
		}
		if s.btnFace[i].A > 0 {
			fillPanel(dst, float32(x+3), float32(y+3), float32(w-6), float32(h-6),
				s.btnFace[i], false)
		}
		s.buttonTextRect(i).drawCentered(dst, s.fnt, uiText(s.b.lang, btn.textKey), 12, body)
	}
	if s.msg != "" {
		s.messageTextRect().drawCentered(dst, s.fnt, s.msg, 12, color.RGBA{235, 160, 120, 255})
	}
}

// gameMenu 進入遊戲選單視窗。
func (b *sceneBuilder) gameMenu() (origScreen, error) { return newGameMenuScreen(b), nil }
