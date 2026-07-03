package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// raceStatRow 是種族統計面板的一列:標籤(raceinfo 來源的英文 key)+ 範例數值字串。
type raceStatRow struct {
	labelEN string
	value   string
}

// 對照 openorion2 info.cpp RaceInfoWidget:標籤來自 RACESTUF(raceinfo.tsv),逐列顯示。
// 數值為示範用(實際由存檔種族資料填)。
var raceStatRows = []raceStatRow{
	{"Population Growth:", "+25%"},
	{"Food Production:", "+0"},
	{"Industrial Production", "+0"},
	{"Scientific Research:", "+0"},
	{"Tax:", "+0"},
	{"Ship Defense:", "+0"},
	{"Ship Offense:", "+0"},
	{"Ground Combat:", "+0"},
	{"Espionage:", "+0"},
}

type raceInfoGame struct {
	font       *uifont.Font
	title      string // misc: RACE STATISTICS
	raceName   string // 種族名
	government string // estrings: 政體
	rows       []raceStatRow
	rowsZH     []string // 已翻譯標籤
	shotPath   string
	frames     int
	tick       int
	saved      bool
}

func (g *raceInfoGame) Update() error {
	g.tick++
	if g.shotPath != "" && g.saved {
		return ebiten.Termination
	}
	return nil
}

func (g *raceInfoGame) Layout(int, int) (int, int) { return helpScreenW, helpScreenH }

func (g *raceInfoGame) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{16, 20, 40, 255})
	border := color.RGBA{80, 110, 180, 255}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{220, 225, 235, 255}

	g.font.DrawCentered(screen, g.title, helpScreenW/2, 26, helpTitleSz, gold)
	vector.StrokeLine(screen, 16, 46, helpScreenW-16, 46, 1, border, false)

	// 種族名 + 政體
	g.font.Draw(screen, g.raceName, 40, 60, 18, gold)
	g.font.Draw(screen, g.government, 40, 88, 15, body)
	vector.StrokeRect(screen, 32, 54, helpScreenW-64, helpScreenH-100, 1, border, false)

	// 統計列(標籤置左、數值置右)
	y := 120.0
	for i, zh := range g.rowsZH {
		g.font.Draw(screen, zh, 48, y, 15, body)
		g.font.Draw(screen, g.rows[i].value, helpScreenW-120, y, 15, body)
		y += 26
	}

	if g.shotPath != "" && !g.saved && g.tick >= g.frames {
		if err := saveScreenshot(screen, g.shotPath); err != nil {
			fmt.Println("截圖失敗:", err)
		}
		g.saved = true
	}
}

// runRaceInfo 渲染種族統計示範畫面(多來源:misc 標題 + raceinfo 標籤 + estrings 政體)。
func runRaceInfo(lang i18n.Lang, fnt *uifont.Font, reg *i18n.Registry, shot string, frames int) error {
	if fnt == nil {
		return fmt.Errorf("種族統計需以 -font 指定字型")
	}
	rowsZH := make([]string, len(raceStatRows))
	for i, r := range raceStatRows {
		rowsZH[i] = reg.Source("raceinfo").Translate(r.labelEN)
	}
	g := &raceInfoGame{
		font:       fnt,
		title:      reg.Source("misc").Translate("RACE STATISTICS"),
		raceName:   "人類 (Human)",
		government: reg.Source("estrings").Translate("Democracy"),
		rows:       raceStatRows,
		rowsZH:     rowsZH,
		shotPath:   shot,
		frames:     frames,
	}
	ebiten.SetWindowSize(helpScreenW, helpScreenH)
	ebiten.SetWindowTitle("Master of Orion II — 種族統計 (cht)")
	return ebiten.RunGame(g)
}
