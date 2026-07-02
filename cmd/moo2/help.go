package main

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// 百科檢視器邏輯尺寸(沿用 MOO2 640×480)。
const (
	helpScreenW = 640
	helpScreenH = 480
	helpBodyX   = 24
	helpBodyTop = 60
	helpLineH   = 20
	helpBodySz  = 15
	helpTitleSz = 20
)

// helpGame 以自繪深色面板顯示一則百科條目(標題 + 自動換行本文)。
// 這是第一個實際用到 HELP.LBX 譯文(help.tsv)的畫面。
type helpGame struct {
	font      *uifont.Font
	title     string
	bodyLines []string
	shotPath  string
	frames    int
	tick      int
	saved     bool
}

func (g *helpGame) Update() error {
	g.tick++
	if g.shotPath != "" && g.saved {
		return ebiten.Termination
	}
	return nil
}

func (g *helpGame) Layout(int, int) (int, int) { return helpScreenW, helpScreenH }

func (g *helpGame) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{16, 20, 40, 255})
	border := color.RGBA{80, 110, 180, 255}
	vector.StrokeRect(screen, 8, 8, helpScreenW-16, helpScreenH-16, 2, border, false)
	g.font.Draw(screen, g.title, helpBodyX, 20, helpTitleSz, color.RGBA{240, 220, 120, 255})
	vector.StrokeLine(screen, helpBodyX, 50, helpScreenW-helpBodyX, 50, 1, border, false)

	y := float64(helpBodyTop)
	for _, ln := range g.bodyLines {
		if y > helpScreenH-helpLineH {
			break // 超出面板下緣(MVP 不捲動)
		}
		g.font.Draw(screen, ln, helpBodyX, y, helpBodySz, color.RGBA{220, 225, 235, 255})
		y += helpLineH
	}

	if g.shotPath != "" && !g.saved && g.tick >= g.frames {
		if err := saveScreenshot(screen, g.shotPath); err != nil {
			fmt.Println("截圖失敗:", err)
		}
		g.saved = true
	}
}

// sanitizeHelpText 移除本文的 \x07 欄位定位碼(其後常接 "X<數字>."):MVP 以空白替代,
// 讓表格類條目不出現亂碼;\n 換行保留交給換行器處理。
func sanitizeHelpText(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == 0x07 {
			j := i + 1
			if j < len(s) && (s[j] == 'X' || s[j] == 'x') {
				j++
				for j < len(s) && s[j] >= '0' && s[j] <= '9' {
					j++
				}
				if j < len(s) && s[j] == '.' {
					j++
				}
			}
			b.WriteByte(' ')
			i = j - 1
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

// loadHelpEntries 開 HELP.LBX 取 asset 0 解析為條目清單。
func loadHelpEntries(dirs []string, lbxName string) ([]lbx.HelpEntry, error) {
	res, err := assets.NewResolver(dirs...)
	if err != nil {
		return nil, err
	}
	arch, err := res.OpenLBX(lbxName)
	if err != nil {
		return nil, err
	}
	raw, err := arch.Asset(0)
	if err != nil {
		return nil, err
	}
	return lbx.ParseHelp(raw)
}

// runHelpList 列出所有百科條目(index、英文標題、是否有中文譯文),headless、不開視窗。
// 供瀏覽/驗證 704 條譯文覆蓋。
func runHelpList(dirs []string, lbxName string, reg *i18n.Registry) error {
	entries, err := loadHelpEntries(dirs, lbxName)
	if err != nil {
		return err
	}
	help := reg.Source("help")
	nTrans := 0
	for i, e := range entries {
		mark := "  "
		if e.Text != "" && help.Has(e.Text) {
			mark = "✓ "
			nTrans++
		} else if e.Text == "" {
			mark = "· " // 空本文(佔位條目)
		}
		fmt.Printf("%s%3d  %s\n", mark, i, e.Title)
	}
	fmt.Printf("\n共 %d 則,有中文本文譯 %d 則\n", len(entries), nTrans)
	return nil
}

// findHelpIndex 依英文標題(不分大小寫)找條目 index,找不到回 -1。
func findHelpIndex(entries []lbx.HelpEntry, title string) int {
	for i, e := range entries {
		if strings.EqualFold(strings.TrimSpace(e.Title), strings.TrimSpace(title)) {
			return i
		}
	}
	return -1
}

// runHelp 載入 HELP.LBX,取指定條目(index 或 title),翻譯後以自繪面板渲染(中/英皆需 -font)。
func runHelp(dirs []string, lbxName string, index int, title string, lang i18n.Lang, fnt *uifont.Font,
	reg *i18n.Registry, shot string, frames int) error {

	if fnt == nil {
		return fmt.Errorf("百科檢視器需以 -font 指定字型(自繪文字,中英皆需)")
	}
	entries, err := loadHelpEntries(dirs, lbxName)
	if err != nil {
		return err
	}
	if title != "" {
		if idx := findHelpIndex(entries, title); idx >= 0 {
			index = idx
		} else {
			return fmt.Errorf("找不到標題為 %q 的百科條目", title)
		}
	}
	if index < 0 || index >= len(entries) {
		return fmt.Errorf("help index %d 超出範圍(共 %d 則)", index, len(entries))
	}
	e := entries[index]

	title, body := e.Title, e.Text
	if lang == i18n.Traditional {
		// 標題可能是科技/元件名(在 tech.tsv 等),用 merged 備援;本文 key 在 help 來源。
		title = reg.Translate(e.Title)
		body = reg.Source("help").Translate(e.Text) // 先以 raw key 查譯文
	}
	body = sanitizeHelpText(body) // 再淨化排版碼供顯示

	maxW := float64(helpScreenW - 2*helpBodyX)
	g := &helpGame{
		font:      fnt,
		title:     title,
		bodyLines: fnt.Wrap(body, helpBodySz, maxW),
		shotPath:  shot,
		frames:    frames,
	}
	ebiten.SetWindowSize(helpScreenW, helpScreenH)
	ebiten.SetWindowTitle("Master of Orion II — 百科 (cht)")
	return ebiten.RunGame(g)
}
