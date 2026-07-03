package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// diploRelationRow 是外交關係面板的一列:對手種族(estrings 種族名稱 key)+
// 對我方關係等級(misc.tsv 17 級關係字,如 FEUD/HATE/…/NEUTRAL/…/HARMONY)+
// 條約狀態(misc.tsv 已譯 key,可留空表示無條約)。
type diploRelationRow struct {
	raceEN     string // estrings: 種族名稱
	relationEN string // misc: 外交關係等級(17 級)
	treatyEN   string // misc: 條約狀態(RESEARCH TREATY: / TRADE TREATY: 等),空字串表無
}

// 對照 openorion2 外交畫面(DIPLOMSE.LBX 對白 + RACES 畫面關係顯示)。
// 範例對手為示範用(實際由存檔外交狀態填)。
var diploRelationRows = []diploRelationRow{
	{"Klackon", "HARMONY", "RESEARCH TREATY: "},
	{"Psilon", "NEUTRAL", "TRADE TREATY: "},
	{"Silicoid", "HATE", ""},
}

// diploRow 是已翻譯的一列顯示資料。
type diploRow struct {
	raceZH     string
	relationZH string
	treatyZH   string // 空字串表示不顯示條約行
}

type diploGame struct {
	font     *uifont.Font
	title    string // 外交關係(畫面標題,見 runDiploView 註解)
	rows     []diploRow
	shotPath string
	frames   int
	tick     int
	saved    bool
}

func (g *diploGame) Update() error {
	g.tick++
	if g.shotPath != "" && g.saved {
		return ebiten.Termination
	}
	return nil
}

func (g *diploGame) Layout(int, int) (int, int) { return helpScreenW, helpScreenH }

func (g *diploGame) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{16, 20, 40, 255})
	border := color.RGBA{80, 110, 180, 255}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{220, 225, 235, 255}
	dim := color.RGBA{160, 170, 190, 255}

	g.font.DrawCentered(screen, g.title, helpScreenW/2, 26, helpTitleSz, gold)
	vector.StrokeLine(screen, 16, 46, helpScreenW-16, 46, 1, border, false)
	vector.StrokeRect(screen, 32, 54, helpScreenW-64, helpScreenH-100, 1, border, false)

	// 每個對手帝國一列:種族名置左、關係等級置右;條約狀態(若有)另起一行縮排顯示。
	y := 90.0
	for _, r := range g.rows {
		g.font.Draw(screen, r.raceZH, 48, y, 16, gold)
		g.font.Draw(screen, r.relationZH, helpScreenW-160, y, 15, body)
		y += 26
		if r.treatyZH != "" {
			g.font.Draw(screen, r.treatyZH, 64, y, 13, dim)
			y += 22
		}
		y += 8
	}

	if g.shotPath != "" && !g.saved && g.tick >= g.frames {
		if err := saveScreenshot(screen, g.shotPath); err != nil {
			fmt.Println("截圖失敗:", err)
		}
		g.saved = true
	}
}

// runDiploView 渲染外交關係示範畫面(多來源:estrings 種族名 + misc 關係等級/條約狀態)。
//
// misc.tsv / estrings.tsv 皆未收錄「外交」畫面標題本身的 key(已 grep 確認:
// 兩檔都沒有 "Diplomacy" 相關字串,help.tsv 內的 "diplomacy screen" 僅出現在句子翻譯中,
// 並非可獨立取用的 UI key)。因此標題比照 raceinfo.go 的 raceName、colonyview.go 的
// colonyName,採固定字串,不強行套用不相關的 TSV key。
func runDiploView(lang i18n.Lang, fnt *uifont.Font, reg *i18n.Registry, shot string, frames int) error {
	if fnt == nil {
		return fmt.Errorf("外交關係需以 -font 指定字型")
	}
	rows := make([]diploRow, len(diploRelationRows))
	for i, r := range diploRelationRows {
		treatyZH := ""
		if r.treatyEN != "" {
			treatyZH = reg.Source("misc").Translate(r.treatyEN)
		}
		rows[i] = diploRow{
			raceZH:     reg.Source("estrings").Translate(r.raceEN),
			relationZH: reg.Source("misc").Translate(r.relationEN),
			treatyZH:   treatyZH,
		}
	}
	g := &diploGame{
		font:     fnt,
		title:    "外交關係",
		rows:     rows,
		shotPath: shot,
		frames:   frames,
	}
	ebiten.SetWindowSize(helpScreenW, helpScreenH)
	ebiten.SetWindowTitle("Master of Orion II — 外交關係 (cht)")
	return ebiten.RunGame(g)
}
