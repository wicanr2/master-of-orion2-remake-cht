package main

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// interactive.go:把「還原原版」的各原版畫面串成單一可互動、可導覽的程式(-game)。
//
// 核心設計(忠實原版 MOO2 + 繁中):畫面用真的 LBX 背景圖 + 中文標籤擦底疊字
// (overlayScreen,同 overlay.go 的手法),在原版按鈕位置加點擊熱區,滑鼠點選在
// 畫面間導覽。這是「用新技術還原原版 MOO2 並中文化」的骨幹——與自製簡約殼(play.go)
// 不同,這裡每個畫面都是原版美術。
//
// 目前串接:原版主選單 →(新遊戲/繼續)→ 原版行星列表 →(返回)→ 主選單。
// 後續逐畫面把更多原版畫面(殖民地/研究/星圖/戰鬥…)改為 overlay 真美術並接進導覽。

const moo2ScreenW, moo2ScreenH = 640, 480

// origTransition 是原版畫面切換指令。
type origTransition struct {
	next origScreen
	quit bool
}

// origScreen 是一個可互動的原版畫面。
type origScreen interface {
	update(in shell.InputState) *origTransition
	draw(dst *ebiten.Image)
}

// hitRegion 是畫面上一塊可點區域 + 動作 id(通常等於該按鈕的英文 key)。
type hitRegion struct {
	x, y, w, h int
	action     string
}

func (h hitRegion) hit(x, y int) bool {
	return x >= h.x && x < h.x+h.w && y >= h.y && y < h.y+h.h
}

// --- overlayScreen:原版 LBX 背景 + 中文標籤覆蓋 + 點擊熱區 ---

type overlayScreen struct {
	bg         *ebiten.Image
	rgba       *image.RGBA
	font       *uifont.Font
	cat        *i18n.Catalog
	overlays   []labelRect
	labelColor color.RGBA
	defSize    float64
	hits       []hitRegion
	onAction   func(action string) *origTransition
	hover      string
}

func (s *overlayScreen) update(in shell.InputState) *origTransition {
	s.hover = ""
	for _, h := range s.hits {
		if h.hit(in.MouseX, in.MouseY) {
			s.hover = h.action
			break
		}
	}
	if in.ClickReleased {
		for _, h := range s.hits {
			if h.hit(in.MouseX, in.MouseY) && s.onAction != nil {
				return s.onAction(h.action)
			}
		}
	}
	return nil
}

func (s *overlayScreen) draw(dst *ebiten.Image) {
	dst.DrawImage(s.bg, nil)
	if s.cat.Lang() == i18n.Traditional {
		for _, b := range s.overlays {
			plate := samplePlate(s.rgba, b)
			// 擦掉烘進圖的英文(蓋底色),再疊中文(同 overlay.go)。
			vector.DrawFilledRect(dst, float32(b.x+3), float32(b.y+2),
				float32(b.w-6), float32(b.h-4), plate, false)
			size := b.size
			if size == 0 {
				size = s.defSize
			}
			zh := s.cat.Translate(b.enKey)
			s.font.DrawCentered(dst, zh, float64(b.x)+float64(b.w)/2, float64(b.y)+float64(b.h)/2, size, s.labelColor)
		}
	}
	// hover 熱區以細框提示可點(互動回饋)。
	if s.hover != "" {
		for _, h := range s.hits {
			if h.action == s.hover {
				vector.StrokeRect(dst, float32(h.x), float32(h.y), float32(h.w), float32(h.h),
					1, color.RGBA{255, 240, 120, 200}, false)
			}
		}
	}
}

// samplePlate 取標籤左內緣中線的底色(用來擦掉英文;置中文字不在此,採到的是乾淨底板)。
func samplePlate(rgba *image.RGBA, b labelRect) color.RGBA {
	x, y := b.x+6, b.y+b.h/2
	if x < 0 || y < 0 || x >= rgba.Bounds().Dx() || y >= rgba.Bounds().Dy() {
		return color.RGBA{0, 0, 0, 255}
	}
	i := rgba.PixOffset(x, y)
	return color.RGBA{rgba.Pix[i], rgba.Pix[i+1], rgba.Pix[i+2], 255}
}

// loadOverlayScreen 載入某原版畫面(LBX 背景 + 譯表),組成可互動的 overlayScreen。
func loadOverlayScreen(res *assets.Resolver, lbxName string, assetID int, lang i18n.Lang,
	fnt *uifont.Font, tsvPath string, overlays []labelRect, labelColor color.RGBA, defSize float64,
	hits []hitRegion, onAction func(string) *origTransition) (*overlayScreen, error) {

	arch, err := res.OpenLBX(lbxName)
	if err != nil {
		return nil, err
	}
	raw, err := arch.Asset(assetID)
	if err != nil {
		return nil, err
	}
	im, err := lbx.DecodeImage(raw)
	if err != nil {
		return nil, err
	}
	if im.Embedded == nil {
		return nil, fmt.Errorf("%s 資產 %d 無內嵌調色盤", lbxName, assetID)
	}
	rgba := im.Frames[0].ToRGBA(im.Embedded, im.KeyColor())

	cat := i18n.New(lang)
	if f, err := os.Open(tsvPath); err == nil {
		defer f.Close()
		if _, err := cat.LoadTSV(f); err != nil {
			return nil, err
		}
	} else if lang == i18n.Traditional {
		return nil, fmt.Errorf("開啟譯表 %s: %w", tsvPath, err)
	}

	return &overlayScreen{
		bg: ebiten.NewImageFromImage(rgba), rgba: rgba, font: fnt, cat: cat,
		overlays: overlays, labelColor: labelColor, defSize: defSize,
		hits: hits, onAction: onAction,
	}, nil
}

// --- sceneBuilder:依需求建構各原版畫面(共用 resolver/字型/語言)---

type sceneBuilder struct {
	res  *assets.Resolver
	fnt  *uifont.Font
	lang i18n.Lang
}

// menu 建原版主選單畫面。按鈕熱區用 menuOverlays 的座標(按鈕即標籤)。
func (b *sceneBuilder) menu() (*overlayScreen, error) {
	hits := make([]hitRegion, 0, len(menuOverlays))
	for _, o := range menuOverlays {
		hits = append(hits, hitRegion{o.x, o.y, o.w, o.h, o.enKey})
	}
	onAction := func(a string) *origTransition {
		switch a {
		case "Quit Game":
			return &origTransition{quit: true}
		case "New Game", "Continue":
			// 進入遊戲:先示範導覽到真原版行星列表畫面(後續接真正的新遊戲流程)。
			s, err := b.planets()
			if err != nil {
				fmt.Fprintln(os.Stderr, "載入行星列表:", err)
				return nil
			}
			return &origTransition{next: s}
		}
		// Load Game / Multi Player / Hall of Fame:尚未實作,暫不動作。
		return nil
	}
	return loadOverlayScreen(b.res, "mainmenu.lbx", 21, b.lang, b.fnt, "assets/i18n/menu.tsv",
		menuOverlays, color.RGBA{104, 224, 96, 255}, 15, hits, onAction)
}

// planets 建原版行星列表畫面。「返回」按鈕熱區導回主選單。
func (b *sceneBuilder) planets() (*overlayScreen, error) {
	hits := []hitRegion{{454, 440, 157, 23, "Return"}}
	onAction := func(a string) *origTransition {
		if a == "Return" {
			s, err := b.menu()
			if err != nil {
				fmt.Fprintln(os.Stderr, "載入主選單:", err)
				return nil
			}
			return &origTransition{next: s}
		}
		return nil
	}
	return loadOverlayScreen(b.res, "plntsum.lbx", 0, b.lang, b.fnt, "assets/i18n/planets.tsv",
		planetsOverlays, color.RGBA{206, 218, 240, 255}, 14, hits, onAction)
}

// --- interactiveApp(ebiten.Game;支援 headless 腳本驗證)---

type interactiveApp struct {
	cur origScreen

	// headless 驗證:script 逐幀注入輸入,跑滿 frames 存 shot。
	script   []shell.InputState
	shotPath string
	frames   int
	tick     int
	saved    bool
}

func (a *interactiveApp) pollInput() shell.InputState {
	if a.script != nil {
		if idx := a.tick - 1; idx >= 0 && idx < len(a.script) {
			return a.script[idx]
		}
		return shell.InputState{}
	}
	x, y := ebiten.CursorPosition()
	return shell.InputState{
		MouseX: x, MouseY: y,
		ClickReleased: inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft),
	}
}

func (a *interactiveApp) Update() error {
	a.tick++
	if t := a.cur.update(a.pollInput()); t != nil {
		if t.quit {
			return ebiten.Termination
		}
		if t.next != nil {
			a.cur = t.next
		}
	}
	if a.shotPath != "" && a.saved {
		return ebiten.Termination
	}
	return nil
}

func (a *interactiveApp) Draw(dst *ebiten.Image) {
	a.cur.draw(dst)
	if a.shotPath != "" && !a.saved && a.tick >= a.frames {
		if err := saveScreenshot(dst, a.shotPath); err != nil {
			fmt.Println("截圖失敗:", err)
		}
		a.saved = true
	}
}

func (a *interactiveApp) Layout(int, int) (int, int) { return moo2ScreenW, moo2ScreenH }

// runInteractive 啟動「還原原版」的互動遊戲。script/shot 非空時為 headless 驗證。
func runInteractive(dirs []string, lang i18n.Lang, fnt *uifont.Font,
	script []shell.InputState, shot string, frames int) error {

	if lang == i18n.Traditional && fnt == nil {
		return fmt.Errorf("中文模式需以 -font 指定 CJK 字型")
	}
	res, err := assets.NewResolver(dirs...)
	if err != nil {
		return err
	}
	b := &sceneBuilder{res: res, fnt: fnt, lang: lang}
	menu, err := b.menu()
	if err != nil {
		return err
	}
	app := &interactiveApp{cur: menu, script: script, shotPath: shot, frames: frames}
	ebiten.SetWindowSize(moo2ScreenW, moo2ScreenH)
	ebiten.SetWindowTitle("Master of Orion II — 繁體中文化 (remake)")
	return ebiten.RunGame(app)
}
