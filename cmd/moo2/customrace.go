package main

// customrace.go:原版「自訂種族(Custom Race)」點數畫面(HANDOFF 優先2 第 2 步)。
//
// 點數值來源:docs/tech/custom-race-picks.md(官方 patch 1.5 config.json 的 race_pick 預設,
// 手冊本身無數字)。起始 10 Picks;負成本=退點。生產/成長/戰鬥類的數值加成會實際套用到
// 開局(session.ApplyCustomRaceBonuses);政府型態與特殊能力的深層效果尚未模擬(只計點數/記錄,
// 見該文件)。版面為合成近似,尚未對原版截圖像素對齊。

import (
	"fmt"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// pickOpt 是一個可循環選項:點數成本 + 對 Race 數值欄位的增量(未對應者留 0)。
type pickOpt struct {
	label                              string
	cost                               int
	ind, res, food, growth, combat, bc int
}

// pickCat 是一組互斥選項(循環選一),如「人口成長:無/差/佳/優」。
type pickCat struct {
	name string
	opts []pickOpt
	sel  int
}

// specialPick 是特殊能力開關(可開關;exclGroup 相同者互斥)。
type specialPick struct {
	label     string
	cost      int
	on        bool
	exclGroup int // 0=無互斥;同號互斥
}

// 生產/戰鬥/政府類:循環選一。數值加成僅套用有對應 Race 欄位者;其餘(商業稅賦、
// 艦防/地面/諜報、政府效果)目前只計點數(見 custom-race-picks.md)。
func defaultPickCats() []pickCat {
	return []pickCat{
		{"人口成長", []pickOpt{{"無", 0, 0, 0, 0, 0, 0, 0}, {"差", -4, 0, 0, 0, -50, 0, 0}, {"佳", 3, 0, 0, 0, 50, 0, 0}, {"優", 6, 0, 0, 0, 100, 0, 0}}, 0},
		{"農業", []pickOpt{{"無", 0, 0, 0, 0, 0, 0, 0}, {"差", -3, 0, 0, -1, 0, 0, 0}, {"佳", 4, 0, 0, 1, 0, 0, 0}, {"優", 7, 0, 0, 2, 0, 0, 0}}, 0},
		{"工業", []pickOpt{{"無", 0, 0, 0, 0, 0, 0, 0}, {"差", -3, -1, 0, 0, 0, 0, 0}, {"佳", 3, 1, 0, 0, 0, 0, 0}, {"優", 6, 2, 0, 0, 0, 0, 0}}, 0},
		{"研究", []pickOpt{{"無", 0, 0, 0, 0, 0, 0, 0}, {"差", -3, 0, -1, 0, 0, 0, 0}, {"佳", 3, 0, 1, 0, 0, 0, 0}, {"優", 6, 0, 2, 0, 0, 0, 0}}, 0},
		{"商業", []pickOpt{{"無", 0, 0, 0, 0, 0, 0, 0}, {"差", -4, 0, 0, 0, 0, 0, 0}, {"佳", 5, 0, 0, 0, 0, 0, 0}, {"優", 8, 0, 0, 0, 0, 0, 0}}, 0}, // 稅賦效果待實作
		{"艦艇攻擊", []pickOpt{{"無", 0, 0, 0, 0, 0, 0, 0}, {"差", -2, 0, 0, 0, 0, -20, 0}, {"佳", 2, 0, 0, 0, 0, 20, 0}, {"優", 4, 0, 0, 0, 0, 50, 0}}, 0},
		{"政府型態", []pickOpt{{"獨裁", 0, 0, 0, 0, 0, 0, 0}, {"封建", -4, 0, 0, 0, 0, 0, 0}, {"統一", 6, 0, 0, 0, 0, 0, 0}, {"民主", 7, 0, 0, 0, 0, 0, 0}}, 0}, // 政府效果待實作
	}
}

// 特殊能力:開關;數值加成多屬深層效果,目前只計點數(記錄待實作)。互斥成對以 exclGroup 標記。
func defaultSpecials() []specialPick {
	return []specialPick{
		{"大型母星", 1, false, 0},
		{"富礦母星", 2, false, 1}, {"貧礦母星", -1, false, 1},
		{"富創造力", 8, false, 2}, {"缺乏創造力", -4, false, 2},
		{"魅力非凡", 3, false, 3}, {"惹人厭", -6, false, 3},
		{"環境耐受", 10, false, 0},
		{"水生", 5, false, 0},
		{"幸運", 3, false, 0},
		{"貿易奇才", 4, false, 0},
	}
}

type customRaceScreen struct {
	b        *sceneBuilder
	fnt      *uifont.Font
	bg       *ebiten.Image
	cats     []pickCat
	specials []specialPick
	hoverCat int
	hoverSpc int
}

const startingPicks = 10

func (b *sceneBuilder) customRace() (origScreen, error) {
	s := &customRaceScreen{
		b: b, fnt: b.fnt, cats: defaultPickCats(), specials: defaultSpecials(),
		hoverCat: -1, hoverSpc: -1,
	}
	if im, err := decodeAsset(b.res, "raceopt.lbx", 0); err == nil && im.Embedded != nil {
		s.bg = ebiten.NewImageFromImage(im.Frames[0].ToRGBA(im.Embedded, im.KeyColor()))
	}
	return s, nil
}

// spent 回傳已花點數(負選項退點,故總和可為負)。
func (s *customRaceScreen) spent() int {
	total := 0
	for _, c := range s.cats {
		total += c.opts[c.sel].cost
	}
	for _, sp := range s.specials {
		if sp.on {
			total += sp.cost
		}
	}
	return total
}

func (s *customRaceScreen) remaining() int { return startingPicks - s.spent() }

// 版面。
const (
	crCatX, crCatY = 30, 92
	crCatH         = 30
	crCatW         = 250
	crSpcX, crSpcY = 330, 92
	crSpcH         = 26
	crSpcW         = 280
)

func (s *customRaceScreen) catRect(i int) (int, int, int, int) { return crCatX, crCatY + i*crCatH, crCatW, crCatH - 4 }
func (s *customRaceScreen) spcRect(i int) (int, int, int, int) { return crSpcX, crSpcY + i*crSpcH, crSpcW, crSpcH - 4 }
func (s *customRaceScreen) cancelRect() (int, int, int, int)   { return 40, 440, 120, 28 }
func (s *customRaceScreen) acceptRect() (int, int, int, int)   { return 480, 440, 120, 28 }

func (s *customRaceScreen) update(in shell.InputState) *origTransition {
	s.hoverCat, s.hoverSpc = -1, -1
	for i := range s.cats {
		if x, y, w, h := s.catRect(i); hitBox(in.MouseX, in.MouseY, x, y, w, h) {
			s.hoverCat = i
		}
	}
	for i := range s.specials {
		if x, y, w, h := s.spcRect(i); hitBox(in.MouseX, in.MouseY, x, y, w, h) {
			s.hoverSpc = i
		}
	}
	if !in.ClickReleased {
		return nil
	}
	if s.hoverCat >= 0 { // 循環該類選項
		c := &s.cats[s.hoverCat]
		c.sel = (c.sel + 1) % len(c.opts)
		if clickSound != nil {
			clickSound()
		}
		return nil
	}
	if s.hoverSpc >= 0 { // 開關特殊能力(開啟時關掉同互斥組其他項)
		sp := &s.specials[s.hoverSpc]
		sp.on = !sp.on
		if sp.on && sp.exclGroup != 0 {
			for j := range s.specials {
				if j != s.hoverSpc && s.specials[j].exclGroup == sp.exclGroup {
					s.specials[j].on = false
				}
			}
		}
		if clickSound != nil {
			clickSound()
		}
		return nil
	}
	if x, y, w, h := s.cancelRect(); hitBox(in.MouseX, in.MouseY, x, y, w, h) {
		if clickSound != nil {
			clickSound()
		}
		sc, err := s.b.raceSelect()
		if err != nil {
			fmt.Fprintf(os.Stderr, "返回種族選擇: %v\n", err)
			return nil
		}
		return &origTransition{next: sc}
	}
	if x, y, w, h := s.acceptRect(); hitBox(in.MouseX, in.MouseY, x, y, w, h) {
		if s.remaining() < 0 { // 點數超支不可接受
			return nil
		}
		if clickSound != nil {
			clickSound()
		}
		s.applyAndStart()
		return &origTransition{next: s.b.nameFlag("自訂帝國")}
	}
	return nil
}

// applyAndStart 聚合已選數值加成成一個 Race,套用並開局。
func (s *customRaceScreen) applyAndStart() {
	b := s.b
	if b.session == nil {
		return
	}
	var r shell.Race
	r.Name = "自訂種族"
	for _, c := range s.cats {
		o := c.opts[c.sel]
		r.IndBonus += o.ind
		r.ResBonus += o.res
		r.FoodBonus += o.food
		r.GrowthPct += o.growth
		r.CombatPct += o.combat
		r.StartBC += o.bc
	}
	b.session.Difficulty = b.newGameDiff
	b.newGameSeed++
	b.session.SetupNewGame(shell.GalaxySizes[b.newGameSize].Stars, int64(b.newGameSeed*7919+42), 3)
	b.session.ApplyCustomRaceBonuses(r)
	// 政府型態效果(僅已建模資源乘數;政府型態循環索引即 shell.Governments 索引)。
	for _, c := range s.cats {
		if c.name == "政府型態" {
			b.session.ApplyGovernment(c.sel)
			break
		}
	}
}

func (s *customRaceScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{0, 0, 0, 255})
	if s.bg != nil {
		dst.DrawImage(s.bg, nil)
	}
	if s.fnt == nil {
		return
	}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{206, 218, 240, 255}
	red := color.RGBA{235, 130, 120, 255}
	green := color.RGBA{140, 210, 150, 255}

	s.fnt.DrawCentered(dst, "自訂種族", 320, 46, 18, gold)
	rem := s.remaining()
	remCol := gold
	if rem < 0 {
		remCol = red
	}
	s.fnt.DrawCentered(dst, fmt.Sprintf("剩餘點數 %d / %d", rem, startingPicks), 320, 70, 14, remCol)

	// 左:循環類。
	for i, c := range s.cats {
		x, y, w, h := s.catRect(i)
		bgc := color.RGBA{26, 34, 50, 160}
		if i == s.hoverCat {
			bgc = color.RGBA{40, 56, 84, 210}
		}
		vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), bgc, false)
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1, color.RGBA{90, 120, 170, 255}, false)
		o := c.opts[c.sel]
		s.fnt.Draw(dst, c.name, float64(x+8), float64(y)+float64(h)/2-8, 13, body)
		costStr := ""
		if o.cost != 0 {
			costStr = fmt.Sprintf(" (%+d)", -o.cost) // 顯示對剩餘點數的影響:退點=+
		}
		s.fnt.Draw(dst, o.label+costStr, float64(x+140), float64(y)+float64(h)/2-8, 13, gold)
	}

	// 右:特殊能力開關。
	for i, sp := range s.specials {
		x, y, w, h := s.spcRect(i)
		bgc := color.RGBA{26, 34, 50, 140}
		if sp.on {
			bgc = color.RGBA{30, 60, 44, 210}
		}
		if i == s.hoverSpc {
			bgc = color.RGBA{48, 60, 84, 210}
		}
		vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), bgc, false)
		bord := color.RGBA{90, 120, 170, 255}
		if sp.on {
			bord = green
		}
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1, bord, false)
		mark := "○"
		if sp.on {
			mark = "●"
		}
		col := body
		if sp.on {
			col = green
		}
		s.fnt.Draw(dst, mark+" "+sp.label, float64(x+8), float64(y)+float64(h)/2-8, 12, col)
		s.fnt.Draw(dst, fmt.Sprintf("%+d", -sp.cost), float64(x+w-34), float64(y)+float64(h)/2-8, 12, gold)
	}

	// 底部按鈕。
	drawBtn := func(rect func() (int, int, int, int), label string, accent color.RGBA, enabled bool) {
		x, y, w, h := rect()
		vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{34, 34, 44, 255}, false)
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1.5, accent, false)
		lc := body
		if !enabled {
			lc = color.RGBA{110, 110, 120, 255}
		}
		s.fnt.DrawCentered(dst, label, float64(x+w/2), float64(y+h/2), 14, lc)
	}
	drawBtn(s.cancelRect, "取消", color.RGBA{160, 140, 100, 255}, true)
	drawBtn(s.acceptRect, "接受", color.RGBA{120, 200, 130, 255}, rem >= 0)
}
