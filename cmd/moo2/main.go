// moo2 是遊戲主程式骨架(Phase 2):用 ebiten 開視窗,載入玩家正版 .lbx 的一張背景圖並繪製。
//
// 此階段僅驗證「資料層 → ebiten 畫面」全鏈路可跑;完整 UI/gameplay 為後續 Phase。
//
// 用法:
//
//	moo2 -data <遊戲資料夾>[,<patch 資料夾>...] [-lbx mainmenu.lbx] [-asset 21]
//	     [-shot out.png] [-frames 3]   # headless:跑 N 幀後存截圖並結束
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
)

// MOO2 的畫面為 640×480(openorion2 screen.h 亦以此為邏輯座標)。
// 實際邏輯尺寸依載入的背景圖 bounds 決定。

type game struct {
	bg           *ebiten.Image
	logicalW     int
	logicalH     int
	shotPath     string
	frames       int
	tick         int
	saved        bool
}

func (g *game) Update() error {
	g.tick++
	if g.shotPath != "" && g.saved {
		return ebiten.Termination // 截圖已存,結束
	}
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.bg, nil)
	// headless 截圖:跑滿指定幀數後,讀回畫面存 PNG。
	if g.shotPath != "" && !g.saved && g.tick >= g.frames {
		if err := saveScreenshot(screen, g.shotPath); err != nil {
			fmt.Fprintln(os.Stderr, "截圖失敗:", err)
			os.Exit(1)
		}
		g.saved = true
	}
}

func (g *game) Layout(outsideW, outsideH int) (int, int) { return g.logicalW, g.logicalH }

func saveScreenshot(img *ebiten.Image, path string) error {
	b := img.Bounds()
	rgba := image.NewRGBA(b)
	img.ReadPixels(rgba.Pix)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, rgba)
}

func main() {
	dataDirs := flag.String("data", "", "遊戲資料夾,可用逗號串多個(前者優先,如 patch,base)")
	lbxName := flag.String("lbx", "mainmenu.lbx", "背景所在的 .lbx")
	assetID := flag.Int("asset", 21, "背景資產 index")
	shot := flag.String("shot", "", "headless 截圖輸出路徑(設定則跑 N 幀後結束)")
	frames := flag.Int("frames", 3, "截圖前先跑幾幀")
	savePath := flag.String("save", "", "存檔路徑;設定則以星圖模式繪製該存檔")
	flag.Parse()

	// 星圖模式:載入存檔並繪製(資料驅動畫面)。
	if *savePath != "" {
		if err := runGalaxy(*savePath, *shot, *frames); err != nil {
			fatal(err)
		}
		return
	}

	if *dataDirs == "" {
		fmt.Fprintln(os.Stderr, "需指定 -data <遊戲資料夾>")
		os.Exit(2)
	}
	dirs := strings.Split(*dataDirs, ",")

	res, err := assets.NewResolver(dirs...)
	if err != nil {
		fatal(err)
	}
	arch, err := res.OpenLBX(*lbxName)
	if err != nil {
		fatal(err)
	}
	raw, err := arch.Asset(*assetID)
	if err != nil {
		fatal(err)
	}
	im, err := lbx.DecodeImage(raw)
	if err != nil {
		fatal(fmt.Errorf("解碼資產 %d: %w", *assetID, err))
	}
	if im.Embedded == nil {
		fatal(fmt.Errorf("資產 %d 無內嵌調色盤(此骨架僅示範背景圖)", *assetID))
	}
	rgba := im.Frames[0].ToRGBA(im.Embedded, im.KeyColor())

	g := &game{
		bg:       ebiten.NewImageFromImage(rgba),
		logicalW: im.Width,
		logicalH: im.Height,
		shotPath: *shot,
		frames:   *frames,
	}

	ebiten.SetWindowSize(im.Width, im.Height)
	ebiten.SetWindowTitle("Master of Orion II — remake (cht)")
	if err := ebiten.RunGame(g); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "錯誤:", err)
	os.Exit(1)
}
