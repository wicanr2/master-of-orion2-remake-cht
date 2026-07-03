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
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// MOO2 的畫面為 640×480(openorion2 screen.h 亦以此為邏輯座標)。
// 實際邏輯尺寸依載入的背景圖 bounds 決定。

type game struct {
	bg       *ebiten.Image
	logicalW int
	logicalH int
	shotPath string
	frames   int
	tick     int
	saved    bool
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
	fontPath := flag.String("font", "", "CJK 字型檔(.ttf/.otf/.ttc);設定則用它渲染中文")
	menuMode := flag.Bool("menu", false, "主選單模式")
	planetsMode := flag.Bool("planets", false, "行星列表畫面模式")
	helpMode := flag.Bool("help-viewer", false, "百科檢視器模式")
	helpIndex := flag.Int("help-index", 1, "百科條目 index(HELP.LBX asset0)")
	helpTitle := flag.String("help-title", "", "依英文標題選百科條目(優先於 -help-index)")
	helpList := flag.Bool("help-list", false, "列出所有百科條目(headless,不開視窗)")
	infoMode := flag.Bool("info-viewer", false, "科技總覽畫面模式(示範單畫面多 TSV 來源)")
	infoTech := flag.String("info-tech", "Achilles Targeting Unit", "科技總覽右欄範例科技(英文標題)")
	raceMode := flag.Bool("race-viewer", false, "種族統計畫面模式")
	tsvPath := flag.String("tsv", "", "譯表 TSV(留空用該畫面預設)")
	lang := flag.String("lang", "zh", "語言:zh(繁中)或 en(英文)")
	flag.Parse()

	langID := i18n.Traditional
	if *lang == "en" {
		langID = i18n.English
	}

	// 畫面覆蓋(擦底疊字)模式:主選單 / 行星列表。
	if *menuMode || *planetsMode {
		if *dataDirs == "" {
			fmt.Fprintln(os.Stderr, "需指定 -data <遊戲資料夾>")
			os.Exit(2)
		}
		fnt, err := loadFont(*fontPath)
		if err != nil {
			fatal(fmt.Errorf("載入字型: %w", err))
		}
		dirs := strings.Split(*dataDirs, ",")
		if *menuMode {
			tsv := *tsvPath
			if tsv == "" {
				tsv = "assets/i18n/menu.tsv"
			}
			err = runMenu(dirs, langID, fnt, tsv, *shot, *frames)
		} else {
			tsv := *tsvPath
			if tsv == "" {
				tsv = "assets/i18n/planets.tsv"
			}
			err = runPlanets(dirs, langID, fnt, tsv, *shot, *frames)
		}
		if err != nil {
			fatal(err)
		}
		return
	}

	// 百科檢視器模式:HELP.LBX + help.tsv,自繪面板顯示一則條目。
	if *helpMode {
		if *dataDirs == "" {
			fmt.Fprintln(os.Stderr, "需指定 -data <遊戲資料夾>")
			os.Exit(2)
		}
		fnt, err := loadFont(*fontPath)
		if err != nil {
			fatal(fmt.Errorf("載入字型: %w", err))
		}
		reg := i18n.NewRegistry(langID)
		if _, err := reg.LoadFS(os.DirFS("assets/i18n"), "."); err != nil {
			fatal(fmt.Errorf("載入譯表: %w", err))
		}
		dirs := strings.Split(*dataDirs, ",")
		if *helpList {
			if err := runHelpList(dirs, "help.lbx", reg); err != nil {
				fatal(err)
			}
			return
		}
		if err := runHelp(dirs, "help.lbx", *helpIndex, *helpTitle, langID, fnt, reg, *shot, *frames); err != nil {
			fatal(err)
		}
		return
	}

	// 科技總覽模式:示範單畫面多 TSV 來源(misc 標題/分組 + tech 名 + help 本文)。
	if *infoMode {
		if *dataDirs == "" {
			fmt.Fprintln(os.Stderr, "需指定 -data <遊戲資料夾>")
			os.Exit(2)
		}
		fnt, err := loadFont(*fontPath)
		if err != nil {
			fatal(fmt.Errorf("載入字型: %w", err))
		}
		reg := i18n.NewRegistry(langID)
		if _, err := reg.LoadFS(os.DirFS("assets/i18n"), "."); err != nil {
			fatal(fmt.Errorf("載入譯表: %w", err))
		}
		dirs := strings.Split(*dataDirs, ",")
		if err := runInfoReview(dirs, "help.lbx", langID, fnt, reg, *infoTech, *shot, *frames); err != nil {
			fatal(err)
		}
		return
	}

	// 種族統計模式:多來源(misc 標題 + raceinfo 標籤 + estrings 政體),不需遊戲資料。
	if *raceMode {
		fnt, err := loadFont(*fontPath)
		if err != nil {
			fatal(fmt.Errorf("載入字型: %w", err))
		}
		reg := i18n.NewRegistry(langID)
		if _, err := reg.LoadFS(os.DirFS("assets/i18n"), "."); err != nil {
			fatal(fmt.Errorf("載入譯表: %w", err))
		}
		if err := runRaceInfo(langID, fnt, reg, *shot, *frames); err != nil {
			fatal(err)
		}
		return
	}

	// 星圖模式:載入存檔並繪製(資料驅動畫面)。
	if *savePath != "" {
		fnt, err := loadFont(*fontPath)
		if err != nil {
			fatal(fmt.Errorf("載入字型: %w", err))
		}
		if err := runGalaxy(*savePath, fnt, *shot, *frames); err != nil {
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

// loadFont 依副檔名載入 CJK 字型(.ttc 走集合解析取第 0 個)。path 為空回 nil(不畫中文)。
func loadFont(path string) (*uifont.Font, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(filepath.Ext(path), ".ttc") {
		return uifont.LoadCollection(data, 0)
	}
	return uifont.Load(data)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "錯誤:", err)
	os.Exit(1)
}
