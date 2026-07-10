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
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/audio"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
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

// runAudioDump 開啟音樂/音效 LBX,原封抽出所有 WAV entry 寫到 dir。
// 單一 LBX 開啟失敗(例如該版本資料夾沒有該檔)只印警告續跑,不中止其餘檔案。
func runAudioDump(dirs []string, dir string) error {
	res, err := assets.NewResolver(dirs...)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("建立輸出目錄 %q: %w", dir, err)
	}

	total := 0

	dumpMusic := func(lbxName, prefix string) {
		arch, err := res.OpenLBX(lbxName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告:開啟 %s 失敗,略過: %v\n", lbxName, err)
			return
		}
		clips, err := audio.RawMusic(arch)
		if err != nil {
			fmt.Fprintf(os.Stderr, "警告:%s 抽取失敗,略過: %v\n", lbxName, err)
			return
		}
		for _, c := range clips {
			name := fmt.Sprintf("%s_%02d.wav", prefix, c.Index)
			path := filepath.Join(dir, name)
			if err := os.WriteFile(path, c.WAV, 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "警告:寫入 %s 失敗: %v\n", path, err)
				continue
			}
			fmt.Printf("%s\t%d bytes\t%.2fs\n", name, len(c.WAV), wavSeconds(c.WAV))
			total++
		}
	}

	dumpMusic("streamhd.lbx", "streamhd")
	dumpMusic("stream.lbx", "stream")

	if arch, err := res.OpenLBX("sound.lbx"); err != nil {
		fmt.Fprintf(os.Stderr, "警告:開啟 sound.lbx 失敗,略過: %v\n", err)
	} else if clips, err := audio.RawSounds(arch); err != nil {
		fmt.Fprintf(os.Stderr, "警告:sound.lbx 抽取失敗,略過: %v\n", err)
	} else {
		for _, c := range clips {
			var name string
			if c.Name == "" {
				name = fmt.Sprintf("sound_%03d.wav", c.Index)
			} else {
				name = fmt.Sprintf("sound_%03d_%s.wav", c.Index, sanitizeFilename(c.Name))
			}
			path := filepath.Join(dir, name)
			if err := os.WriteFile(path, c.WAV, 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "警告:寫入 %s 失敗: %v\n", path, err)
				continue
			}
			fmt.Printf("%s\t%d bytes\t%.2fs\n", name, len(c.WAV), wavSeconds(c.WAV))
			total++
		}
	}

	fmt.Printf("共抽出 %d 個 wav 檔到 %s\n", total, dir)
	return nil
}

// sanitizeFilename 把名稱中非 [A-Za-z0-9_-] 的字元換成 '_',供組檔名安全使用。
func sanitizeFilename(name string) string {
	b := []byte(name)
	for i, c := range b {
		switch {
		case c >= 'A' && c <= 'Z', c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '_', c == '-':
			// 保留
		default:
			b[i] = '_'
		}
	}
	return string(b)
}

// wavSeconds 從 RIFF/WAVE bytes 的 fmt+data chunk 概算播放秒數;解析失敗回傳 0。
func wavSeconds(wav []byte) float64 {
	clip, err := audio.DecodeWAV(wav)
	if err != nil || clip.SampleRate <= 0 {
		return 0
	}
	// Clip.PCM 已統一轉為 16-bit 雙聲道交錯,每 frame 固定 4 bytes。
	frames := len(clip.PCM) / 4
	return float64(frames) / float64(clip.SampleRate)
}

func main() {
	dataDirs := flag.String("data", "", "遊戲資料夾,可用逗號串多個(前者優先,如 patch,base)")
	lbxName := flag.String("lbx", "mainmenu.lbx", "背景所在的 .lbx")
	assetID := flag.Int("asset", 21, "背景資產 index")
	palAsset := flag.Int("palasset", -1, "調色盤提供資產 index(該 lbx 內;目標資產無內嵌調色盤時用)")
	accum := flag.Bool("accum", false, "多幀 delta 累積渲染(動畫資產如 DIPLOMAT 使節)")
	shot := flag.String("shot", "", "headless 截圖輸出路徑(設定則跑 N 幀後結束)")
	audioDump := flag.String("audiodump", "", "把原版音樂/音效抽成 wav 到此目錄(headless,需 -data)")
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
	gameMode := flag.Bool("game", false, "還原原版互動遊戲(原版主選單→導覽各原版畫面,全繁中;有 -shot 則腳本驗證)")
	playMode := flag.Bool("play", false, "可玩遊戲殼(互動;有 -shot 則跑腳本驗證並截圖)")
	playRecord := flag.String("play-record", "", "錄製模式:scripted playthrough 逐幀存圖到此目錄(供 gameplay footage)")
	colonyMode := flag.Bool("colony-viewer", false, "殖民地摘要畫面模式")
	diploMode := flag.Bool("diplo-viewer", false, "外交關係畫面模式")
	tsvPath := flag.String("tsv", "", "譯表 TSV(留空用該畫面預設)")
	lang := flag.String("lang", "zh", "語言:zh(繁中)或 en(英文)")
	flag.Parse()

	langID := i18n.Traditional
	if *lang == "en" {
		langID = i18n.English
	}

	// 音訊抽取模式:headless,把原版音樂/音效原封抽成 .wav 到指定目錄,
	// 供人耳試聽、建立曲目/音效對應表(不開視窗,不需字型/i18n)。
	if *audioDump != "" {
		if *dataDirs == "" {
			fmt.Fprintln(os.Stderr, "需指定 -data <遊戲資料夾>")
			os.Exit(2)
		}
		dirs := strings.Split(*dataDirs, ",")
		if err := runAudioDump(dirs, *audioDump); err != nil {
			fatal(err)
		}
		return
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

	// 還原原版互動遊戲:原版主選單(真 LBX 美術)→ 點選單導覽到各原版畫面,全繁中。
	// 這是專案主目標「用 go/ebiten 還原原版 MOO2 + 中文化」的骨幹。headless(-shot)時腳本驗證導覽。
	if *gameMode {
		if *dataDirs == "" {
			fmt.Fprintln(os.Stderr, "需指定 -data <遊戲資料夾>")
			os.Exit(2)
		}
		fnt, err := loadFont(*fontPath)
		if err != nil {
			fatal(fmt.Errorf("載入字型: %w", err))
		}
		dirs := strings.Split(*dataDirs, ",")
		var script []shell.InputState
		if *shot != "" {
			// headless 驗證:主選單新遊戲(491,228)→ 星系設定 → Accept(486,405)
			//   →【獨立種族選擇畫面】→ 截圖(驗證新遊戲流程新增的種族畫面)。
			script = []shell.InputState{
				{MouseX: 491, MouseY: 228, ClickReleased: true},
				{MouseX: 486, MouseY: 405, ClickReleased: true},
			}
		}
		if err := runInteractive(dirs, langID, fnt, script, *shot, *frames); err != nil {
			fatal(err)
		}
		return
	}

	// 可玩遊戲殼:互動主選單→遊戲畫面→結束回合。headless(-shot)時跑內建腳本驗證互動。
	if *playMode {
		fnt, err := loadFont(*fontPath)
		if err != nil {
			fatal(fmt.Errorf("載入字型: %w", err))
		}
		// 錄製模式:跑豐富 playthrough 逐幀存圖(gameplay footage)。
		if *playRecord != "" {
			script := recordPlaythrough()
			if err := runPlay(fnt, "", 0, script, *playRecord, len(script)+2); err != nil {
				fatal(err)
			}
			return
		}
		var script []shell.InputState
		if *shot != "" {
			// 腳本 playthrough:新遊戲→管理殖民地→調工人×2→調科學家 → 截圖驗證互動 gameplay。
			script = []shell.InputState{
				{MouseX: 320, MouseY: 218, ClickReleased: true}, // 新遊戲 → 遊戲畫面
				{MouseX: 315, MouseY: 438, ClickReleased: true}, // 管理殖民地 → 殖民地畫面
				{MouseX: 365, MouseY: 185, ClickReleased: true}, // 工人 ▲
				{MouseX: 365, MouseY: 185, ClickReleased: true}, // 工人 ▲
				{MouseX: 365, MouseY: 225, ClickReleased: true}, // 科學家 ▲
			}
		}
		if err := runPlay(fnt, *shot, *frames, script, "", 0); err != nil {
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

	// 殖民地摘要模式:多來源(misc 標題 + rstring 回合結算摘要 + estrings 分類標籤),不需遊戲資料。
	if *colonyMode {
		fnt, err := loadFont(*fontPath)
		if err != nil {
			fatal(fmt.Errorf("載入字型: %w", err))
		}
		reg := i18n.NewRegistry(langID)
		if _, err := reg.LoadFS(os.DirFS("assets/i18n"), "."); err != nil {
			fatal(fmt.Errorf("載入譯表: %w", err))
		}
		if err := runColonyView(langID, fnt, reg, *shot, *frames); err != nil {
			fatal(err)
		}
		return
	}

	// 外交關係模式:多來源(estrings 種族名 + misc 關係等級/條約狀態),不需遊戲資料。
	if *diploMode {
		fnt, err := loadFont(*fontPath)
		if err != nil {
			fatal(fmt.Errorf("載入字型: %w", err))
		}
		reg := i18n.NewRegistry(langID)
		if _, err := reg.LoadFS(os.DirFS("assets/i18n"), "."); err != nil {
			fatal(fmt.Errorf("載入譯表: %w", err))
		}
		if err := runDiploView(langID, fnt, reg, *shot, *frames); err != nil {
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
	// 調色盤:優先用 -palasset 指定的提供資產(供無內嵌調色盤的資產如 DIPLOMAT 使節)。
	pal := im.Embedded
	if *palAsset >= 0 {
		praw, perr := arch.Asset(*palAsset)
		if perr != nil {
			fatal(fmt.Errorf("讀調色盤資產 %d: %w", *palAsset, perr))
		}
		pim, perr := lbx.DecodeImage(praw)
		if perr != nil || pim.Embedded == nil {
			fatal(fmt.Errorf("調色盤資產 %d 無內嵌調色盤", *palAsset))
		}
		pal = pim.Embedded
	}
	if pal == nil {
		fatal(fmt.Errorf("資產 %d 無內嵌調色盤(用 -palasset 指定提供資產)", *assetID))
	}
	var rgba *image.RGBA
	if *accum {
		rgba = im.AccumulatedRGBA(pal) // 多幀 delta 累積(動畫資產)
	} else {
		rgba = im.Frames[0].ToRGBA(pal, im.KeyColor())
	}

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
