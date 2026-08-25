package main

// customrace.go:原版「自訂種族(Custom Race)」點數畫面(HANDOFF 優先2 第 2 步)。
//
// 點數值來源:docs/tech/custom-race-picks.md(官方 patch 1.5 config.json 的 race_pick 預設,
// 手冊本身無數字)。起始 10 Picks;負成本=退點。生產/成長/戰鬥類的數值加成會實際套用到
// 開局(session.ApplyCustomRaceBonuses);已有引擎公式的特殊能力會隨選項寫入並生效，
// 尚未建模的能力仍保留在存檔遮罩中。版面為合成近似,尚未對原版截圖像素對齊。

import (
	"fmt"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// pickOpt 是一個可循環選項:點數成本 + 對 Race 數值欄位的增量(未對應者留 0)。
// incPerPop 為「每人每回合 BC」的半單位增量(對應 shell.Race.IncomePerPop / 手冊 Money pick,
// 差 -1(-0.5)、佳 +1(+0.5)、優 +2(+1); combat 欄依類別承載艦攻／艦防／地面戰／諜報值;
// 見 engine.ColonyState.IncomePerPop 半單位註解。
type pickOpt struct {
	label, labelEn                            string
	cost                                      int
	ind, res, food, growth, combat, incPerPop int
}

// pickCat 是一組互斥選項(循環選一),如「人口成長:無/差/佳/優」。
type pickCat struct {
	name   string // 中文名,同時是這一類的穩定 key(政府型態的效果套用靠它比對)
	nameEn string // 原版用語(Population Growth / Farming / …)
	opts   []pickOpt
	sel    int
}

// specialPick 是特殊能力開關(可開關;exclGroup 相同者互斥)。
type specialPick struct {
	label     string
	labelEn   string // 原版能力名(Large Home World / Creative / Fantastic Traders / …)
	cost      int
	on        bool
	exclGroup int // 0=無互斥;同號互斥
	trait     gamedata.RaceTrait
}

// 生產/戰鬥/政府類:循環選一。數值加成僅套用有對應 Race 欄位者;其餘(商業稅賦、
// 艦防/地面/諜報、政府效果)目前只計點數(見 custom-race-picks.md)。
func defaultPickCats() []pickCat {
	return []pickCat{
		{"人口成長", "Population Growth", []pickOpt{{"無", "None", 0, 0, 0, 0, 0, 0, 0}, {"差", "Poor", -4, 0, 0, 0, -50, 0, 0}, {"佳", "Good", 3, 0, 0, 0, 50, 0, 0}, {"優", "Great", 6, 0, 0, 0, 100, 0, 0}}, 0},
		{"農業", "Farming", []pickOpt{{"無", "None", 0, 0, 0, 0, 0, 0, 0}, {"差", "Poor", -3, 0, 0, -1, 0, 0, 0}, {"佳", "Good", 4, 0, 0, 1, 0, 0, 0}, {"優", "Great", 7, 0, 0, 2, 0, 0, 0}}, 0},
		{"工業", "Industry", []pickOpt{{"無", "None", 0, 0, 0, 0, 0, 0, 0}, {"差", "Poor", -3, -1, 0, 0, 0, 0, 0}, {"佳", "Good", 3, 1, 0, 0, 0, 0, 0}, {"優", "Great", 6, 2, 0, 0, 0, 0, 0}}, 0},
		{"研究", "Research", []pickOpt{{"無", "None", 0, 0, 0, 0, 0, 0, 0}, {"差", "Poor", -3, 0, -1, 0, 0, 0, 0}, {"佳", "Good", 3, 0, 1, 0, 0, 0, 0}, {"優", "Great", 6, 0, 2, 0, 0, 0, 0}}, 0},
		{"商業", "Money", []pickOpt{{"無", "None", 0, 0, 0, 0, 0, 0, 0}, {"差", "Poor", -4, 0, 0, 0, 0, -1, 0}, {"佳", "Good", 5, 0, 0, 0, 0, 1, 0}, {"優", "Great", 8, 0, 0, 0, 0, 2, 0}}, 0}, // Money pick 每人每回合半BC:差-0.5/佳+0.5/優+1(手冊 p.16,同諾蘭姆機制)
		{"艦艇攻擊", "Ship Attack", []pickOpt{{"無", "None", 0, 0, 0, 0, 0, 0, 0}, {"差", "Poor", -2, 0, 0, 0, 0, -20, 0}, {"佳", "Good", 2, 0, 0, 0, 0, 20, 0}, {"優", "Great", 4, 0, 0, 0, 0, 50, 0}}, 0},
		{"艦艇防禦", "Ship Defense", []pickOpt{{"無", "None", 0, 0, 0, 0, 0, 0, 0}, {"差", "Poor", -2, 0, 0, 0, 0, -20, 0}, {"佳", "Good", 3, 0, 0, 0, 0, 25, 0}, {"優", "Great", 7, 0, 0, 0, 0, 50, 0}}, 0},
		{"地面戰", "Ground Combat", []pickOpt{{"無", "None", 0, 0, 0, 0, 0, 0, 0}, {"差", "Poor", -2, 0, 0, 0, 0, -10, 0}, {"佳", "Good", 2, 0, 0, 0, 0, 10, 0}, {"優", "Great", 4, 0, 0, 0, 0, 20, 0}}, 0},
		{"諜報", "Spying", []pickOpt{{"無", "None", 0, 0, 0, 0, 0, 0, 0}, {"差", "Poor", -3, 0, 0, 0, 0, -10, 0}, {"佳", "Good", 3, 0, 0, 0, 0, 10, 0}, {"優", "Great", 6, 0, 0, 0, 0, 20, 0}}, 0},
		{"政府型態", "Government", []pickOpt{{"獨裁", "Dictatorship", 0, 0, 0, 0, 0, 0, 0}, {"封建", "Feudal", -4, 0, 0, 0, 0, 0, 0}, {"統一", "Unification", 6, 0, 0, 0, 0, 0, 0}, {"民主", "Democracy", 7, 0, 0, 0, 0, 0, 0}}, 0}, // 政府效果待實作
	}
}

// 特殊能力:開關;選項會寫入客製種族特性遮罩。清單依手冊 p.23–26 與
// custom-race-picks.md 的官方主表列出完整 22 項；已有引擎公式的能力直接生效，
// 尚未建模的深層效果先保留選項語意。互斥成對以 exclGroup 標記。
func defaultSpecials() []specialPick {
	return []specialPick{
		{"低重力世界", "Low-G World", -5, false, 4, gamedata.TRAIT_LOW_G},
		{"高重力世界", "High-G World", 6, false, 4, gamedata.TRAIT_HIGH_G},
		{"水生", "Aquatic", 5, false, 0, gamedata.TRAIT_AQUATIC},
		{"穴居", "Subterranean", 6, false, 0, gamedata.TRAIT_SUBTERRANEAN},
		{"大型母星", "Large Home World", 1, false, 0, gamedata.TRAIT_LARGE_HOMEWORLD},
		{"富礦母星", "Rich Home World", 2, false, 1, gamedata.TRAIT_RICH_HOMEWORLD},
		{"貧礦母星", "Poor Home World", -1, false, 1, gamedata.TRAIT_POOR_HOMEWORLD},
		{"遺物母星", "Artifacts World", 3, false, 0, gamedata.TRAIT_ARTIFACTS_HOMEWORLD},
		{"半機械化", "Cybernetic", 4, false, 5, gamedata.TRAIT_CYBERNETIC},
		{"食岩", "Lithovore", 10, false, 5, gamedata.TRAIT_LITHOVORE},
		{"惹人厭", "Repulsive", -6, false, 3, gamedata.TRAIT_REPULSIVE},
		{"魅力非凡", "Charismatic", 3, false, 3, gamedata.TRAIT_CHARISMATIC},
		{"缺乏創造力", "Uncreative", -4, false, 2, gamedata.TRAIT_UNCREATIVE},
		{"富創造力", "Creative", 8, false, 2, gamedata.TRAIT_CREATIVE},
		{"環境耐受", "Tolerant", 10, false, 0, gamedata.TRAIT_TOLERANT},
		{"貿易奇才", "Fantastic Traders", 4, false, 0, gamedata.TRAIT_FANTASTIC_TRADERS},
		{"心靈感應", "Telepathic", 6, false, 0, gamedata.TRAIT_TELEPATHIC},
		{"幸運", "Lucky", 3, false, 0, gamedata.TRAIT_LUCKY},
		{"全知", "Omniscient", 3, false, 0, gamedata.TRAIT_OMNISCIENCE},
		{"匿蹤艦船", "Stealthy Ships", 4, false, 0, gamedata.TRAIT_STEALTHY_SHIPS},
		{"戰爭領主", "Warlord", 4, false, 0, gamedata.TRAIT_WARLORD},
		{"跨維度", "Trans-Dimensional", 5, false, 0, gamedata.TRAIT_TRANS_DIMENSIONAL},
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
	crSpcH         = 20
	crSpcRows      = 11 // 22 個官方特殊能力分成兩欄
	crSpcColW      = 140
	crSpcW         = 135
)

func (s *customRaceScreen) catRect(i int) (int, int, int, int) {
	return crCatX, crCatY + i*crCatH, crCatW, crCatH - 4
}
func (s *customRaceScreen) spcRect(i int) (int, int, int, int) {
	return crSpcX + (i/crSpcRows)*crSpcColW, crSpcY + (i%crSpcRows)*crSpcH, crSpcW, crSpcH - 4
}
func (s *customRaceScreen) cancelRect() (int, int, int, int) { return 40, 440, 120, 28 }
func (s *customRaceScreen) acceptRect() (int, int, int, int) { return 480, 440, 120, 28 }

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
		return &origTransition{next: s.b.nameFlag(s.b.tr("自訂帝國", "Custom Empire"))}
	}
	return nil
}

// customRaceValues 把客製畫面已選的數值型 picks 聚合成 shell.Race。
// 艦艇攻擊、防禦、地面戰、諜報共用 pickOpt.combat 欄位,但依類別寫入不同的原版特性欄位。
func customRaceValues(cats []pickCat) shell.Race {
	var r shell.Race
	for _, c := range cats {
		o := c.opts[c.sel]
		r.IndBonus += o.ind
		r.ResBonus += o.res
		r.FoodBonus += o.food
		r.GrowthPct += o.growth
		switch c.nameEn {
		case "Ship Attack":
			r.CombatPct += o.combat
		case "Ship Defense":
			r.ShipDefPct += o.combat
		case "Ground Combat":
			r.GroundCombatBonus += o.combat
		case "Spying":
			r.SpyBonus += o.combat
		}
		r.IncomePerPop += o.incPerPop // 商業 pick → 每人每回合半BC(取代先前捏造的一次性 StartBC)
	}
	return r
}

// applyAndStart 聚合已選數值加成成一個 Race,套用並開局。
func (s *customRaceScreen) applyAndStart() {
	b := s.b
	if b.session == nil {
		return
	}
	r := customRaceValues(s.cats)
	r.Name = b.tr("自訂種族", "Custom")
	b.session.Difficulty = b.newGameDiff
	// 五個 NEW GAME 設定要在 SetupNewGame 之前套用(星系年齡會影響星系生成,見該函式註解)。
	b.applyNewGameSettings()
	b.newGameSeed++
	b.session.SetupNewGame(shell.GalaxySizes[b.newGameSize].Stars, int64(b.newGameSeed*7919+42), b.newGameOpponents())
	b.session.SetRuleProfile(profileForVersion(b.gameVersion)) // 主選單選的 1.3/1.5 規則版本
	traits := make([]gamedata.RaceTrait, 0, len(s.specials))
	for _, sp := range s.specials {
		if sp.on {
			traits = append(traits, sp.trait)
		}
	}
	b.session.ApplyCustomRaceBonuses(r, traits...)
	b.session.SetCustomRaceUnusedPicks(s.remaining())
	// 政府型態效果(僅已建模資源乘數;政府型態循環索引即 shell.Governments 索引)。
	for _, c := range s.cats {
		if c.nameEn == "Government" {
			b.session.ApplyGovernment(c.sel)
			break
		}
	}
}

func (s *customRaceScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{0, 0, 0, 255})
	if s.bg != nil {
		drawPanelImage(dst, s.bg, nil)
	}
	if s.fnt == nil {
		return
	}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{206, 218, 240, 255}
	red := color.RGBA{235, 130, 120, 255}
	green := color.RGBA{140, 210, 150, 255}

	s.fnt.DrawCentered(dst, truncateToWidth(s.fnt, s.b.tr("自訂種族", "CUSTOM RACE"), 18, 600), 320, 46, 18, gold)
	rem := s.remaining()
	remCol := gold
	if rem < 0 {
		remCol = red
	}
	s.fnt.DrawCentered(dst, truncateToWidth(s.fnt, fmt.Sprintf(s.b.tr("剩餘點數 %d / %d", "Picks left %d / %d"), rem, startingPicks), 14, 360),
		320, 70, 14, remCol)

	// 左:循環類。
	for i, c := range s.cats {
		x, y, w, h := s.catRect(i)
		bgc := color.RGBA{26, 34, 50, 160}
		if i == s.hoverCat {
			bgc = color.RGBA{40, 56, 84, 210}
		}
		fillPanel(dst, float32(x), float32(y), float32(w), float32(h), bgc, false)
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1, color.RGBA{90, 120, 170, 255}, false)
		o := c.opts[c.sel]
		s.fnt.Draw(dst, truncateToWidth(s.fnt, s.b.tr(c.name, c.nameEn), 13, float64(w-134)), float64(x+8), float64(y)+float64(h)/2-8, 13, body)
		costStr := ""
		if o.cost != 0 {
			costStr = fmt.Sprintf(" (%+d)", -o.cost) // 顯示對剩餘點數的影響:退點=+
		}
		s.fnt.Draw(dst, truncateToWidth(s.fnt, s.b.tr(o.label, o.labelEn)+costStr, 13, float64(w-148)), float64(x+140), float64(y)+float64(h)/2-8, 13, gold)
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
		fillPanel(dst, float32(x), float32(y), float32(w), float32(h), bgc, false)
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
		s.fnt.Draw(dst, truncateToWidth(s.fnt, mark+" "+s.b.tr(sp.label, sp.labelEn), 12, float64(w-50)), float64(x+8), float64(y)+float64(h)/2-8, 12, col)
		s.fnt.Draw(dst, fmt.Sprintf("%+d", -sp.cost), float64(x+w-34), float64(y)+float64(h)/2-8, 12, gold)
	}

	// 底部按鈕。
	drawBtn := func(rect func() (int, int, int, int), label string, accent color.RGBA, enabled bool) {
		x, y, w, h := rect()
		fillPanel(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{34, 34, 44, 255}, false)
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1.5, accent, false)
		lc := body
		if !enabled {
			lc = color.RGBA{110, 110, 120, 255}
		}
		s.fnt.DrawCentered(dst, truncateToWidth(s.fnt, label, 14, float64(w-10)), float64(x+w/2), float64(y+h/2), 14, lc)
	}
	drawBtn(s.cancelRect, s.b.tr("取消", "CANCEL"), color.RGBA{160, 140, 100, 255}, true)
	drawBtn(s.acceptRect, s.b.tr("接受", "ACCEPT"), color.RGBA{120, 200, 130, 255}, rem >= 0)
}
