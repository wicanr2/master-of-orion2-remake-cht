package main

import (
	"fmt"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/save"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// galaxyGame 從解析後的存檔繪製星圖(Phase 2 資料驅動畫面驗證)。
// 目前用向量色點 + 英文標籤呈現真實星系資料;真實 sprite 美術與 CJK 字型為後續 Phase。
type galaxyGame struct {
	gs       *save.GameState
	font     *uifont.Font // CJK 字型(nil 則標題退回英文 debug 字)
	shotPath string
	frames   int
	tick     int
	saved    bool
}

const (
	galViewW = 640
	galViewH = 480
	plotX0   = 30
	plotY0   = 40
	plotX1   = 610
	plotY1   = 462
)

// 光譜類(SpectralClass 0-6)對應顏色。
var spectralColors = map[uint8]color.RGBA{
	0: {120, 160, 255, 255}, // Blue
	1: {235, 235, 235, 255}, // White
	2: {255, 230, 120, 255}, // Yellow
	3: {255, 170, 80, 255},  // Orange
	4: {230, 90, 80, 255},   // Red
	5: {150, 110, 80, 255},  // Brown
	6: {70, 70, 90, 255},    // BlackHole
}

func (g *galaxyGame) Update() error {
	g.tick++
	if g.shotPath != "" && g.saved {
		return ebiten.Termination
	}
	return nil
}

// mapX/mapY 把星系座標線性映射到繪圖區。
func (g *galaxyGame) mapX(x uint16) float32 {
	gw := g.gs.Galaxy.Width
	if gw == 0 {
		gw = 1
	}
	return plotX0 + float32(int(x))*(plotX1-plotX0)/float32(gw)
}

func (g *galaxyGame) mapY(y uint16) float32 {
	gh := g.gs.Galaxy.Height
	if gh == 0 {
		gh = 1
	}
	return plotY0 + float32(int(y))*(plotY1-plotY0)/float32(gh)
}

func (g *galaxyGame) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{6, 6, 14, 255}) // 深空底

	// 星雲(畫在星星之後,暗紫色暈)。
	for i := 0; i < int(g.gs.Galaxy.NebulaCount) && i < len(g.gs.Galaxy.Nebulas); i++ {
		n := g.gs.Galaxy.Nebulas[i]
		vector.DrawFilledCircle(screen, g.mapX(n.X), g.mapY(n.Y), 46, color.RGBA{40, 16, 48, 255}, true)
	}

	// 星星:依光譜上色、依大小定半徑,標星名。
	for i := 0; i < g.gs.StarCount && i < len(g.gs.Stars); i++ {
		s := g.gs.Stars[i]
		x, y := g.mapX(s.X), g.mapY(s.Y)
		col, ok := spectralColors[s.SpectralClass]
		if !ok {
			col = color.RGBA{200, 200, 200, 255}
		}
		r := float32(5 - int(s.Size)) // StarSize Large=0..Tiny=3
		if r < 2 {
			r = 2
		}
		vector.DrawFilledCircle(screen, x, y, r, col, true)
		if s.Name != "" {
			ebitenutil.DebugPrintAt(screen, s.Name, int(x)+5, int(y)-4)
		}
	}

	// 標題列:有 CJK 字型則畫繁體中文,否則退回英文 debug 字。
	if g.font != nil {
		title := fmt.Sprintf("銀河霸主 II — 星系圖    星數:%d   殖民地:%d   玩家:%d   [%s]",
			g.gs.StarCount, g.gs.ColonyCount, g.gs.PlayerCount, g.gs.Config.SaveGameName)
		g.font.Draw(screen, title, 8, 6, 16, color.RGBA{180, 220, 140, 255})
	} else {
		title := fmt.Sprintf("MASTER OF ORION II  —  Galaxy %dx%d   stars:%d  colonies:%d  players:%d   [%s]",
			g.gs.Galaxy.Width, g.gs.Galaxy.Height, g.gs.StarCount, g.gs.ColonyCount, g.gs.PlayerCount, g.gs.Config.SaveGameName)
		ebitenutil.DebugPrintAt(screen, title, 6, 6)
	}

	if g.shotPath != "" && !g.saved && g.tick >= g.frames {
		if err := saveScreenshot(screen, g.shotPath); err != nil {
			fmt.Println("截圖失敗:", err)
		}
		g.saved = true
	}
}

func (g *galaxyGame) Layout(int, int) (int, int) { return galViewW, galViewH }

// runGalaxy 載入存檔並以星圖模式執行。fnt 為 nil 時標題退回英文。
func runGalaxy(savePath string, fnt *uifont.Font, shot string, frames int) error {
	data, err := os.ReadFile(savePath)
	if err != nil {
		return err
	}
	gs, err := save.Load(data)
	if err != nil {
		return fmt.Errorf("解析存檔: %w", err)
	}
	g := &galaxyGame{gs: gs, font: fnt, shotPath: shot, frames: frames}
	ebiten.SetWindowSize(galViewW, galViewH)
	ebiten.SetWindowTitle("Master of Orion II — Galaxy (cht)")
	return ebiten.RunGame(g)
}
