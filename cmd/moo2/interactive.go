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
	bg               *ebiten.Image
	rgba             *image.RGBA
	font             *uifont.Font
	cat              *i18n.Catalog
	overlays         []labelRect
	labelColor       color.RGBA
	defSize          float64
	hits             []hitRegion
	onAction         func(action string) *origTransition
	hover            string
	offsetX, offsetY int // 背景圖在 640×480 畫布上的置中偏移(小於全螢幕的視窗畫面用)
}

func (s *overlayScreen) update(in shell.InputState) *origTransition {
	// 命中判定在背景圖局部座標(扣掉置中偏移)。
	mx, my := in.MouseX-s.offsetX, in.MouseY-s.offsetY
	s.hover = ""
	for _, h := range s.hits {
		if h.hit(mx, my) {
			s.hover = h.action
			break
		}
	}
	if in.ClickReleased {
		for _, h := range s.hits {
			if h.hit(mx, my) && s.onAction != nil {
				return s.onAction(h.action)
			}
		}
	}
	return nil
}

func (s *overlayScreen) draw(dst *ebiten.Image) {
	if s.offsetX != 0 || s.offsetY != 0 {
		dst.Fill(color.RGBA{0, 0, 0, 255}) // 小於全螢幕的視窗:底填黑再置中
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(s.offsetX), float64(s.offsetY))
	dst.DrawImage(s.bg, op)
	ox, oy := float64(s.offsetX), float64(s.offsetY)
	if s.cat.Lang() == i18n.Traditional {
		for _, b := range s.overlays {
			plate := samplePlate(s.rgba, b)
			// 擦掉烘進圖的英文(蓋底色),再疊中文(同 overlay.go)。
			vector.DrawFilledRect(dst, float32(float64(b.x+3)+ox), float32(float64(b.y+2)+oy),
				float32(b.w-6), float32(b.h-4), plate, false)
			size := b.size
			if size == 0 {
				size = s.defSize
			}
			zh := s.cat.Translate(b.enKey)
			s.font.DrawCentered(dst, zh, float64(b.x)+float64(b.w)/2+ox, float64(b.y)+float64(b.h)/2+oy, size, s.labelColor)
		}
	}
	// hover 熱區以細框提示可點(互動回饋)。
	if s.hover != "" {
		for _, h := range s.hits {
			if h.action == s.hover {
				vector.StrokeRect(dst, float32(float64(h.x)+ox), float32(float64(h.y)+oy),
					float32(h.w), float32(h.h), 1, color.RGBA{255, 240, 120, 200}, false)
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

// paletteProvider 指定「調色盤提供圖」:某些原版畫面背景無完整內嵌調色盤,需借另一張
// 帶調色盤的圖當基底(openorion2 的 base_palette 機制)。lbxName 為空表示不需要。
type paletteProvider struct {
	lbxName string
	assetID int
}

// decodeAsset 解一張 LBX 影像。
func decodeAsset(res *assets.Resolver, lbxName string, assetID int) (*lbx.Image, error) {
	arch, err := res.OpenLBX(lbxName)
	if err != nil {
		return nil, err
	}
	raw, err := arch.Asset(assetID)
	if err != nil {
		return nil, err
	}
	return lbx.DecodeImage(raw)
}

// resolvePalette 重現 openorion2 Image::load 的調色盤合併:
// 最終 = 基底(提供圖的完整調色盤)+ 目標圖自己的部分內嵌範圍疊上去。
// 無提供圖時直接用目標圖自己的內嵌調色盤。
func resolvePalette(res *assets.Resolver, im *lbx.Image, prov paletteProvider) (*lbx.Palette, error) {
	var merged lbx.Palette
	if prov.lbxName != "" {
		base, err := decodeAsset(res, prov.lbxName, prov.assetID)
		if err != nil {
			return nil, fmt.Errorf("載入調色盤提供圖 %s#%d: %w", prov.lbxName, prov.assetID, err)
		}
		if base.Embedded == nil {
			return nil, fmt.Errorf("調色盤提供圖 %s#%d 無內嵌調色盤", prov.lbxName, prov.assetID)
		}
		merged = *base.Embedded
	} else if im.Embedded == nil {
		return nil, fmt.Errorf("畫面圖無內嵌調色盤且未指定提供圖")
	}
	// 疊上目標圖自己的內嵌範圍(部分覆蓋)。
	if im.Embedded != nil {
		for i := im.PalStart; i < im.PalStart+im.PalCount; i++ {
			merged[i] = im.Embedded[i]
		}
	}
	return &merged, nil
}

// loadOverlayScreen 載入某原版畫面(LBX 背景 + 譯表),組成可互動的 overlayScreen。
// prov 非空時走調色盤鏈(無內嵌調色盤的畫面借提供圖上色)。
func loadOverlayScreen(res *assets.Resolver, lbxName string, assetID int, lang i18n.Lang,
	fnt *uifont.Font, tsvPath string, overlays []labelRect, labelColor color.RGBA, defSize float64,
	hits []hitRegion, onAction func(string) *origTransition, prov paletteProvider) (*overlayScreen, error) {

	im, err := decodeAsset(res, lbxName, assetID)
	if err != nil {
		return nil, err
	}
	pal, err := resolvePalette(res, im, prov)
	if err != nil {
		return nil, fmt.Errorf("%s 資產 %d: %w", lbxName, assetID, err)
	}
	rgba := im.Frames[0].ToRGBA(pal, im.KeyColor())

	cat := i18n.New(lang)
	if f, err := os.Open(tsvPath); err == nil {
		defer f.Close()
		if _, err := cat.LoadTSV(f); err != nil {
			return nil, err
		}
	} else if lang == i18n.Traditional {
		return nil, fmt.Errorf("開啟譯表 %s: %w", tsvPath, err)
	}

	// 小於 640×480 的視窗畫面置中(openorion2:_x=(SCREEN_WIDTH-_width)/2)。
	bounds := rgba.Bounds()
	offX := (moo2ScreenW - bounds.Dx()) / 2
	offY := (moo2ScreenH - bounds.Dy()) / 2
	if offX < 0 {
		offX = 0
	}
	if offY < 0 {
		offY = 0
	}
	return &overlayScreen{
		bg: ebiten.NewImageFromImage(rgba), rgba: rgba, font: fnt, cat: cat,
		overlays: overlays, labelColor: labelColor, defSize: defSize,
		hits: hits, onAction: onAction, offsetX: offX, offsetY: offY,
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
		case "Hall of Fame":
			// 暫借「名人堂」入口示範調色盤鏈解鎖的研究選擇畫面(原本無內嵌調色盤)。
			s, err := b.research()
			if err != nil {
				fmt.Fprintln(os.Stderr, "載入研究選擇:", err)
				return nil
			}
			return &origTransition{next: s}
		}
		// Load Game / Multi Player:尚未實作,暫不動作。
		return nil
	}
	return loadOverlayScreen(b.res, "mainmenu.lbx", 21, b.lang, b.fnt, "assets/i18n/menu.tsv",
		menuOverlays, color.RGBA{104, 224, 96, 255}, 15, hits, onAction, paletteProvider{})
}

// research 建原版研究選擇畫面(TECHSEL.LBX 資產 0,無內嵌調色盤 → 走調色盤鏈,
// 基底取自 SCIENCE.LBX 資產 0)。點畫面任一處返回主選單。
func (b *sceneBuilder) research() (*overlayScreen, error) {
	hits := []hitRegion{{0, 0, moo2ScreenW, moo2ScreenH, "back"}}
	onAction := func(a string) *origTransition {
		s, err := b.menu()
		if err != nil {
			fmt.Fprintln(os.Stderr, "載入主選單:", err)
			return nil
		}
		return &origTransition{next: s}
	}
	// 先以忠實原版(不疊字、置中)呈現,證明調色盤鏈解鎖此畫面;
	// 領域名擦底疊字(建設/動力/化學/社會學/電腦/生物學/物理/力場)+ 座標校對列為下一輪。
	return loadOverlayScreen(b.res, "techsel.lbx", 0, b.lang, b.fnt, "assets/i18n/tech.tsv",
		nil, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteProvider{"science.lbx", 0})
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
		planetsOverlays, color.RGBA{206, 218, 240, 255}, 14, hits, onAction, paletteProvider{})
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
