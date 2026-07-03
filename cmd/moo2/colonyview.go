package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// colonyStatRow 是殖民地摘要面板的一列:標籤(英文 key)+ 來源 TSV 名 + 範例數值字串。
type colonyStatRow struct {
	labelEN string
	source  string // i18n.Registry 來源名(哪個 TSV 有該 key)
	value   string
}

// 對照 openorion2 回合結算的殖民經濟摘要(RSTRING0.LBX)與種族/科技分類標籤(ESTRINGS.LBX)。
// 數值為示範用(實際由存檔殖民地資料填)。
var colonyStatRows = []colonyStatRow{
	{"Population", "estrings", "5.2M"},
	{"Food Summary", "rstring", "+12"},
	{"Industry Summary", "rstring", "+34"},
	{"Research", "estrings", "+18"},
	{"Research Summary", "rstring", "+18"},
	{"Morale Summary", "rstring", "普通"},
	{"Pollution Penalty", "rstring", "-2"},
}

type colonyViewGame struct {
	font       *uifont.Font
	title      string // misc: Colony Summary
	colonyName string // 殖民地名
	rows       []colonyStatRow
	rowsZH     []string // 已翻譯標籤
	shotPath   string
	frames     int
	tick       int
	saved      bool
}

func (g *colonyViewGame) Update() error {
	g.tick++
	if g.shotPath != "" && g.saved {
		return ebiten.Termination
	}
	return nil
}

func (g *colonyViewGame) Layout(int, int) (int, int) { return helpScreenW, helpScreenH }

func (g *colonyViewGame) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{16, 20, 40, 255})
	border := color.RGBA{80, 110, 180, 255}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{220, 225, 235, 255}

	g.font.DrawCentered(screen, g.title, helpScreenW/2, 26, helpTitleSz, gold)
	vector.StrokeLine(screen, 16, 46, helpScreenW-16, 46, 1, border, false)

	// 殖民地名
	g.font.Draw(screen, g.colonyName, 40, 60, 18, gold)
	vector.StrokeRect(screen, 32, 54, helpScreenW-64, helpScreenH-100, 1, border, false)

	// 統計列(標籤置左、數值置右)
	y := 90.0
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

// runColonyView 渲染殖民地摘要示範畫面(多來源:misc 標題 + rstring 回合結算摘要 + estrings 分類標籤)。
func runColonyView(lang i18n.Lang, fnt *uifont.Font, reg *i18n.Registry, shot string, frames int) error {
	if fnt == nil {
		return fmt.Errorf("殖民地摘要需以 -font 指定字型")
	}
	rowsZH := make([]string, len(colonyStatRows))
	for i, r := range colonyStatRows {
		rowsZH[i] = reg.Source(r.source).Translate(r.labelEN)
	}
	// 標題:misc.tsv 只有「Colony」單字(無「Colony Summary」複合詞),故取其譯文「殖民」
	// 再接固定字「地摘要」,組成「殖民地摘要」(等同 openorion2 回合結算的殖民經濟摘要畫面)。
	title := reg.Source("misc").Translate("Colony") + "地摘要"
	g := &colonyViewGame{
		font:       fnt,
		title:      title,
		colonyName: "地球 (Earth)",
		rows:       colonyStatRows,
		rowsZH:     rowsZH,
		shotPath:   shot,
		frames:     frames,
	}
	ebiten.SetWindowSize(helpScreenW, helpScreenH)
	ebiten.SetWindowTitle("Master of Orion II — 殖民地摘要 (cht)")
	return ebiten.RunGame(g)
}
