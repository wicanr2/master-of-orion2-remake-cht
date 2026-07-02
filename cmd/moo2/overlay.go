package main

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// labelRect 是一個「烘進背景圖的英文標籤」的位置與其英文 key(座標多取自 openorion2)。
type labelRect struct {
	x, y, w, h int
	enKey      string
	size       float64 // 字級;0 用預設
}

// overlayGame 以擦底疊字把某畫面背景圖上烘進的英文標籤換成中文(mom ChtLabel 手法)。
// 英文模式直接顯示原版背景。
type overlayGame struct {
	bg         *ebiten.Image
	rgba       *image.RGBA
	font       *uifont.Font
	cat        *i18n.Catalog
	overlays   []labelRect
	labelColor color.RGBA
	defSize    float64
	shotPath   string
	frames     int
	tick       int
	saved      bool
}

func (g *overlayGame) Update() error {
	g.tick++
	if g.shotPath != "" && g.saved {
		return ebiten.Termination
	}
	return nil
}

func (g *overlayGame) sampleAt(x, y int) color.RGBA {
	if x < 0 || y < 0 || x >= g.rgba.Bounds().Dx() || y >= g.rgba.Bounds().Dy() {
		return color.RGBA{0, 0, 0, 255}
	}
	i := g.rgba.PixOffset(x, y)
	return color.RGBA{g.rgba.Pix[i], g.rgba.Pix[i+1], g.rgba.Pix[i+2], 255}
}

func (g *overlayGame) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.bg, nil)
	if g.cat.Lang() == i18n.Traditional {
		for _, b := range g.overlays {
			plate := g.sampleAt(b.x+6, b.y+b.h/2)
			vector.DrawFilledRect(screen, float32(b.x+3), float32(b.y+3),
				float32(b.w-6), float32(b.h-6), plate, false)
			size := b.size
			if size == 0 {
				size = g.defSize
			}
			zh := g.cat.Translate(b.enKey)
			g.font.DrawCentered(screen, zh, float64(b.x)+float64(b.w)/2, float64(b.y)+float64(b.h)/2, size, g.labelColor)
		}
	}
	if g.shotPath != "" && !g.saved && g.tick >= g.frames {
		if err := saveScreenshot(screen, g.shotPath); err != nil {
			fmt.Println("截圖失敗:", err)
		}
		g.saved = true
	}
}

func (g *overlayGame) Layout(int, int) (int, int) { return g.bg.Bounds().Dx(), g.bg.Bounds().Dy() }

// runOverlay 載入某畫面背景圖,套用標籤覆蓋表渲染(中/英)。
func runOverlay(dirs []string, lbxName string, assetID int, lang i18n.Lang, fnt *uifont.Font,
	tsvPath string, overlays []labelRect, labelColor color.RGBA, defSize float64,
	title, shot string, frames int) error {

	if lang == i18n.Traditional && fnt == nil {
		return fmt.Errorf("中文模式需以 -font 指定 CJK 字型")
	}
	res, err := assets.NewResolver(dirs...)
	if err != nil {
		return err
	}
	arch, err := res.OpenLBX(lbxName)
	if err != nil {
		return err
	}
	raw, err := arch.Asset(assetID)
	if err != nil {
		return err
	}
	im, err := lbx.DecodeImage(raw)
	if err != nil {
		return err
	}
	if im.Embedded == nil {
		return fmt.Errorf("%s 資產 %d 無內嵌調色盤", lbxName, assetID)
	}
	rgba := im.Frames[0].ToRGBA(im.Embedded, im.KeyColor())

	cat := i18n.New(lang)
	if f, err := os.Open(tsvPath); err == nil {
		defer f.Close()
		if _, err := cat.LoadTSV(f); err != nil {
			return err
		}
	} else if lang == i18n.Traditional {
		return fmt.Errorf("開啟譯表 %s: %w", tsvPath, err)
	}

	g := &overlayGame{
		bg: ebiten.NewImageFromImage(rgba), rgba: rgba, font: fnt, cat: cat,
		overlays: overlays, labelColor: labelColor, defSize: defSize,
		shotPath: shot, frames: frames,
	}
	ebiten.SetWindowSize(im.Width, im.Height)
	ebiten.SetWindowTitle(title)
	return ebiten.RunGame(g)
}
