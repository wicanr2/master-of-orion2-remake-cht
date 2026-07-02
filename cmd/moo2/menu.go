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

// 主選單六按鈕:座標取自 openorion2 mainmenu.cpp(x=415, w=153),英文文字烘在背景圖。
// enKey 對應 assets/i18n/menu.tsv 的英文 key。
var menuButtons = []struct {
	x, y, w, h int
	enKey      string
}{
	{415, 172, 153, 23, "Continue"},
	{415, 195, 153, 22, "Load Game"},
	{415, 217, 153, 23, "New Game"},
	{415, 240, 153, 22, "Multi Player"},
	{415, 262, 153, 23, "Hall of Fame"},
	{415, 285, 153, 22, "Quit Game"},
}

// menuGame 以擦底疊字把主選單按鈕中文化(mom ChtLabel 手法)。
type menuGame struct {
	bg       *ebiten.Image
	rgba     *image.RGBA // 供採樣底色
	font     *uifont.Font
	cat      *i18n.Catalog
	shotPath string
	frames   int
	tick     int
	saved    bool
}

func (g *menuGame) Update() error {
	g.tick++
	if g.shotPath != "" && g.saved {
		return ebiten.Termination
	}
	return nil
}

// sampleAt 讀背景圖某點的顏色(採樣按鈕底色用)。
func (g *menuGame) sampleAt(x, y int) color.RGBA {
	if x < 0 || y < 0 || x >= g.rgba.Bounds().Dx() || y >= g.rgba.Bounds().Dy() {
		return color.RGBA{0, 0, 0, 255}
	}
	i := g.rgba.PixOffset(x, y)
	return color.RGBA{g.rgba.Pix[i], g.rgba.Pix[i+1], g.rgba.Pix[i+2], 255}
}

func (g *menuGame) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.bg, nil)

	labelColor := color.RGBA{104, 224, 96, 255} // 選單亮綠(近似原版按鈕字色)
	for _, b := range menuButtons {
		// 採樣按鈕左內緣的底色(避開置中文字),當擦底色。
		plate := g.sampleAt(b.x+6, b.y+b.h/2)
		// 擦掉烘進圖的英文:蓋一塊底色(內縮保留邊框凹凸)。
		vector.DrawFilledRect(screen, float32(b.x+4), float32(b.y+3),
			float32(b.w-8), float32(b.h-6), plate, false)
		// 疊置中中文。
		zh := g.cat.Translate(b.enKey)
		tw, th := g.font.Measure(zh, 15)
		tx := float64(b.x) + (float64(b.w)-tw)/2
		ty := float64(b.y) + (float64(b.h)-th)/2
		g.font.Draw(screen, zh, tx, ty, 15, labelColor)
	}

	if g.shotPath != "" && !g.saved && g.tick >= g.frames {
		if err := saveScreenshot(screen, g.shotPath); err != nil {
			fmt.Println("截圖失敗:", err)
		}
		g.saved = true
	}
}

func (g *menuGame) Layout(int, int) (int, int) { return g.bg.Bounds().Dx(), g.bg.Bounds().Dy() }

// runMenu 渲染中文化主選單。
func runMenu(dirs []string, fnt *uifont.Font, tsvPath, shot string, frames int) error {
	if fnt == nil {
		return fmt.Errorf("中文選單需以 -font 指定 CJK 字型")
	}
	res, err := assets.NewResolver(dirs...)
	if err != nil {
		return err
	}
	arch, err := res.OpenLBX("mainmenu.lbx")
	if err != nil {
		return err
	}
	raw, err := arch.Asset(21) // ASSET_MENU_BACKGROUND
	if err != nil {
		return err
	}
	im, err := lbx.DecodeImage(raw)
	if err != nil {
		return err
	}
	if im.Embedded == nil {
		return fmt.Errorf("主選單背景無內嵌調色盤")
	}
	rgba := im.Frames[0].ToRGBA(im.Embedded, im.KeyColor())

	cat := i18n.New(i18n.Traditional)
	f, err := os.Open(tsvPath)
	if err != nil {
		return fmt.Errorf("開啟譯表 %s: %w", tsvPath, err)
	}
	defer f.Close()
	if _, err := cat.LoadTSV(f); err != nil {
		return err
	}

	g := &menuGame{
		bg: ebiten.NewImageFromImage(rgba), rgba: rgba, font: fnt, cat: cat,
		shotPath: shot, frames: frames,
	}
	ebiten.SetWindowSize(im.Width, im.Height)
	ebiten.SetWindowTitle("Master of Orion II — 主選單 (cht)")
	return ebiten.RunGame(g)
}
