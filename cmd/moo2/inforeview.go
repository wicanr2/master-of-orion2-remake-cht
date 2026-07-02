package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// 科技分組(BILLTEX2 26 條;Defense 出現兩次=殖民防禦/艦艇防禦,原文同名)。
// 對照 openorion2 info.cpp TechReviewWidget → misctext(BILLTEX2, ...)。
var techGroups = []string{
	"Miscellaneous", "New Construction Types", "Spies", "Colony", "Ground Combat",
	"Ship Equipment", "Food", "Pollution", "Morale", "Production", "Defense",
	"Research", "Money", "Beams", "Missiles/Torpedoes", "Bombs/Biological",
	"Fighters", "Special", "Shields", "Armor", "Drives", "Computers",
	"Fuels/Range", "Scanners", "Offense",
}

// infoReviewGame 自繪「科技總覽」畫面,示範單畫面多 TSV 來源:
// 標題/分組列 = misc,詳情科技名 = tech,詳情本文 = help(對照 SA2 規格 §2 TechReviewWidget)。
type infoReviewGame struct {
	font       *uifont.Font
	title      string
	groups     []string // 已翻譯的分組名
	detailName string   // 範例科技名(tech 來源)
	detailBody []string // 範例 help 本文(help 來源,已換行)
	shotPath   string
	frames     int
	tick       int
	saved      bool
}

func (g *infoReviewGame) Update() error {
	g.tick++
	if g.shotPath != "" && g.saved {
		return ebiten.Termination
	}
	return nil
}

func (g *infoReviewGame) Layout(int, int) (int, int) { return helpScreenW, helpScreenH }

func (g *infoReviewGame) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{16, 20, 40, 255})
	border := color.RGBA{80, 110, 180, 255}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{220, 225, 235, 255}

	// 標題列
	g.font.DrawCentered(screen, g.title, helpScreenW/2, 26, helpTitleSz, gold)
	vector.StrokeLine(screen, 16, 46, helpScreenW-16, 46, 1, border, false)

	// 左欄:26 科技分組
	vector.StrokeRect(screen, 12, 54, 210, helpScreenH-66, 1, border, false)
	y := 62.0
	for _, gname := range g.groups {
		g.font.Draw(screen, gname, 22, y, 14, body)
		y += 16
	}

	// 右欄:詳情(科技名 + help 本文)
	rx := float32(232)
	vector.StrokeRect(screen, rx, 54, helpScreenW-16-rx, helpScreenH-66, 1, border, false)
	g.font.Draw(screen, g.detailName, float64(rx)+12, 64, 17, gold)
	vector.StrokeLine(screen, rx+12, 90, helpScreenW-28, 90, 1, border, false)
	dy := 100.0
	for _, ln := range g.detailBody {
		if dy > helpScreenH-24 {
			break
		}
		g.font.Draw(screen, ln, float64(rx)+12, dy, helpBodySz, body)
		dy += helpLineH
	}

	if g.shotPath != "" && !g.saved && g.tick >= g.frames {
		if err := saveScreenshot(screen, g.shotPath); err != nil {
			fmt.Println("截圖失敗:", err)
		}
		g.saved = true
	}
}

// runInfoReview 渲染「科技總覽」示範畫面。detailTech 為右欄範例科技(以英文標題查 HELP.LBX 取本文,
// 需同時在 tech.tsv 與 help.tsv)。
func runInfoReview(dirs []string, lbxName string, lang i18n.Lang, fnt *uifont.Font, reg *i18n.Registry,
	detailTech, shot string, frames int) error {
	if fnt == nil {
		return fmt.Errorf("科技總覽需以 -font 指定字型")
	}
	entries, err := loadHelpEntries(dirs, lbxName)
	if err != nil {
		return err
	}
	helpBody := ""
	if idx := findHelpIndex(entries, detailTech); idx >= 0 {
		helpBody = reg.Source("help").Translate(entries[idx].Text)
	}
	misc := reg.Source("misc")
	groups := make([]string, len(techGroups))
	for i, gname := range techGroups {
		groups[i] = misc.Translate(gname)
	}
	maxW := float64(helpScreenW - 16 - 232 - 24)
	g := &infoReviewGame{
		font:       fnt,
		title:      misc.Translate("TECH REVIEW"),
		groups:     groups,
		detailName: reg.Translate(detailTech), // tech 來源(merged)
		detailBody: fnt.Wrap(helpBody, helpBodySz, maxW),
		shotPath:   shot,
		frames:     frames,
	}
	ebiten.SetWindowSize(helpScreenW, helpScreenH)
	ebiten.SetWindowTitle("Master of Orion II — 科技總覽 (cht)")
	return ebiten.RunGame(g)
}
