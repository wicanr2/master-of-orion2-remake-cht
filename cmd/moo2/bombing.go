package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// bombing.go:軌道轟炸畫面。`sub_B4D02 @ 0xB4D02` 是結果外層入口，
// `sub_B4800 @ 0xB4800` 才是逐幀畫面回呼。
//
// remake 的軌道轟炸解算是完整的(`internal/shell/orbital_bombardment.go`:建築吸收、人口損失、
// 防禦方反擊都算),但結果只是星系主畫面上的一行字。
//
// ============ 反組譯挖到的事實 ============
//
//	`Draw_Colony_Bombing_Screen_` @ 0xB4800
//	  標題:`sprintf` 星球名後 `Print_Centered_(x=0x13F=319, y=0x0A=10)`
//	        ——與 `Draw_Colony_Landing_Screen_` 同一個錨點,兩個畫面共用版面慣例。
//
//	`Do_Bomb_` @ 0xB4606 — 炸彈記錄的結構逐欄位讀得出來,**每筆 15 位元組**(`imul esi, eax, 0Fh`)
//	  於 `dword_19DEC0`:
//	    +0x00 精靈指標   +0x04 X   +0x06 Y   +0x08 目前動畫幀   +0x0E 啟用旗標
//	  每個 tick:已啟用且幀數未到頂 → 幀 +1 並重畫;**未啟用者 `Random_(5) == 1` 才啟用**
//	  ——也就是每 tick 20% 機率開始落下,所以原版的炸彈是零零星星散著炸,不是整排同時落地。
//	  (`Random_` 回 1..n,見 docs/re/01-gap-report.md 的語意訂正。)
//
//	`Add_Bomb_To_End_Of_Queue_` @ 0xB43A6 — 在 **0x31 = 49** 個目標槽上做蓄水池抽樣挑下一個
//	  挨炸的目標(`inc edx / Random_(edx) == 1` 的經典寫法)。49 這個數字對上殖民地建築格數。
//
//	`Load_Bombing_Anims_` @ 0xB435A — 只載 **COLONY.LBX 資產 1**。
//	  ⚠ **這份遊戲資料裡該資產是 0 位元組**(COLONY.LBX 資產 0–4 全部長度為 0,實測)。
//	  所以炸彈動畫的圖抽不出來,remake 用自繪的爆點代替,並在此標明——這是資料層的事實,
//	  不是「懶得做」。
//
// ============ 背景 ============
//
//	COLONY.LBX 資產 8(640×480,6 幀 delta)是殖民地的**建築格地面**——透視格線那張。
//	原版的轟炸畫面就是把炸彈落在這張格子上(呼應上面 49 個目標槽)。
//	調色盤借 COLBLDG.LBX#0,格線解出來是一條紫色階(#6C306C / #8C488C / #B05CB0)。
//
//	這條紫色階在 remake 中作為帝國色佔位色，但 IDA 已證實 `sub_B8EFB`
//	的直接呼叫全在地面戰 loader `sub_B6D51` 內，**不能**證明軌道轟炸也用同一表。
//	因此三階換色只是 remake 視覺近似，並且使用被轟炸的守方色，不冒稱原版精確還原。
//
// ============ 誠實留白 ============
//
//	① remake 的殖民地不是**空間格**模型(只有職務人數 + 建築集合),沒有「第 k 座建築在格子的
//	   哪個位置」這種資料,所以炸彈落點是依結果數量平均散在格面上,不是打在真正的建築格上。
//	   要忠實重現得先讓殖民地變成格子模型,那是另一條線。
//	② 原版是逐幀動畫(20% 機率逐顆啟動);remake 呈現戰後定格。
//	③ 炸彈精靈缺圖(見上),用自繪爆點。

// 原版轟炸畫面的錨點與資產。
const (
	bombTitleSize    = 18
	bombGridLBX      = "colony.lbx"
	bombGridAsset    = 8 // 建築格地面(640×480,6 幀 delta,需累積)
	bombGridPalLBX   = "colbldg.lbx"
	bombGridPalAsset = 0
	// bombTargetSlots 是原版一次轟炸掃過的目標槽數(Add_Bomb_To_End_Of_Queue_ 的 `cmp si, 31h`)。
	// remake 沒有格子模型用不到它,留著記錄這個數字。
	bombTargetSlots = 49
)

// playerColorRamp 是建築格地面上「待換成帝國旗色」的三階佔位色(實測自 COLONY.LBX#8
// 以 COLBLDG#0 上色後的結果;語意見檔頭的 Replace_Colgcbt_Color_With_Player_Colors_)。
var bombingPlaceholderRamp = [3]color.RGBA{
	{0x6C, 0x30, 0x6C, 0xFF}, // 暗
	{0x8C, 0x48, 0x8C, 0xFF}, // 中
	{0xB0, 0x5C, 0xB0, 0xFF}, // 亮
}

// recolorBombingGridApprox 是將背景佔位色映射到守方旗色的 remake 視覺近似。
// 55% / 75% / 100% 三階並非原版已證實表格。
func recolorBombingGridApprox(src *image.RGBA, flag int) *image.RGBA {
	if flag < 0 || flag >= len(shell.FlagColors) {
		flag = 0
	}
	f := shell.FlagColors[flag]
	scale := func(pct int) color.RGBA {
		return color.RGBA{uint8(int(f.R) * pct / 100), uint8(int(f.G) * pct / 100), uint8(int(f.B) * pct / 100), 0xFF}
	}
	want := [3]color.RGBA{scale(55), scale(75), scale(100)}
	out := image.NewRGBA(src.Bounds())
	copy(out.Pix, src.Pix)
	for i := 0; i+3 < len(out.Pix); i += 4 {
		c := color.RGBA{out.Pix[i], out.Pix[i+1], out.Pix[i+2], out.Pix[i+3]}
		for k, p := range bombingPlaceholderRamp {
			if c.R == p.R && c.G == p.G && c.B == p.B {
				out.Pix[i], out.Pix[i+1], out.Pix[i+2] = want[k].R, want[k].G, want[k].B
				break
			}
		}
	}
	return out
}

// bombingScreen 是一次軌道轟炸的戰報畫面。
type bombingScreen struct {
	b    *sceneBuilder
	fnt  *uifont.Font
	res  shell.GroundBombardResult
	grid *ebiten.Image
}

func newBombingScreen(b *sceneBuilder, res shell.GroundBombardResult) *bombingScreen {
	s := &bombingScreen{b: b, fnt: b.fnt, res: res}
	prov, err := decodeAsset(b.res, bombGridPalLBX, bombGridPalAsset)
	if err != nil || prov.Embedded == nil {
		return s
	}
	im, err := decodeAsset(b.res, bombGridLBX, bombGridAsset)
	if err != nil || len(im.Frames) == 0 {
		return s
	}
	s.grid = ebiten.NewImageFromImage(recolorBombingGridApprox(im.AccumulatedRGBA(prov.Embedded), res.DefenderColor))
	return s
}

func (s *bombingScreen) contRect() (int, int, int, int) { return 265, 432, 110, 30 }

func bombingTitleTextRect() textSafeRect {
	// 18px 中文 bitmap 字的實際字墨高可達 32px；安全框保留 34px，
	// 同時將中心維持在 y=19，對應原版 y=10 上緣錨點加半個字高。
	return textSafeRect{x: 10, y: 2, w: 620, h: 34, insetX: 4, insetY: 1, lineH: 32}
}

func bombingReportTextRect(row int) textSafeRect {
	return textSafeRect{x: 12, y: 70 + row*24, w: 616, h: 24, insetX: 4, insetY: 1, lineH: 22}
}

func (s *bombingScreen) continueTextRect() textSafeRect {
	x, y, w, h := s.contRect()
	return textSafeRect{x: x, y: y, w: w, h: h, insetX: 5, insetY: 1, lineH: h - 2}
}

func (s *bombingScreen) update(in shell.InputState) *origTransition {
	if !in.ClickReleased {
		return nil
	}
	if x, y, w, h := s.contRect(); hitRect(in, x, y, w, h) {
		return s.b.goTo(s.b.galaxy, uiText(s.b.lang, "bombing.transition.galaxy"))
	}
	return nil
}

// impactPoint 回傳第 i 個彈著點。原版的落點是真正的建築格座標,remake 沒有格子模型
// (見檔頭留白①),改成沿格面平均散開的可重現位置。
func impactPoint(i int) (float32, float32) {
	// 格面在畫面下半部(見 COLONY.LBX#8 的透視格線),取 y 300..430、x 60..580。
	x := 60 + (i*97)%520
	y := 300 + (i*53)%130
	return float32(x), float32(y)
}

func (s *bombingScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{6, 8, 12, 255})
	if s.grid != nil {
		drawPanelImage(dst, s.grid, nil)
	}
	if s.fnt == nil {
		return
	}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{228, 232, 240, 255}
	warn := color.RGBA{240, 150, 120, 255}

	title := uiText(s.b.lang, "bombing.title.default")
	if s.res.ColonyName != "" {
		title = fmt.Sprintf(uiText(s.b.lang, "bombing.title.colony"), s.res.ColonyName)
	}
	bombingTitleTextRect().drawCentered(dst, s.fnt, title, bombTitleSize, gold)

	// 彈著點:每摧毀一座建築畫一個,再加上人口損失數(炸到居住區)。
	// ⚠ 精靈缺圖(COLONY.LBX#1 在這份資料裡是 0 位元組,見檔頭),用自繪爆點。
	impacts := s.res.BuildingsDestroyed + s.res.PopulationLost
	for i := 0; i < impacts && i < 40; i++ {
		x, y := impactPoint(i)
		vector.DrawFilledCircle(dst, x, y, 9, color.RGBA{255, 170, 60, 190}, true)
		vector.DrawFilledCircle(dst, x, y, 4, color.RGBA{255, 240, 190, 230}, true)
	}

	// 戰報。壓一條半透明深色條,格線底上直接寫字看不清。
	fillPanel(dst, 0, 60, moo2ScreenW, 118, color.RGBA{6, 4, 8, 180}, false)
	row := 0
	line := func(text string, col color.RGBA) {
		bombingReportTextRect(row).drawCentered(dst, s.fnt, text, 14, col)
		row++
	}
	line(fmt.Sprintf(uiText(s.b.lang, "bombing.report.salvo"),
		s.res.TotalDamage, s.res.Hits), body)
	line(fmt.Sprintf(uiText(s.b.lang, "bombing.report.buildings"),
		s.res.BuildingsDestroyed, s.res.BuildingsRemaining), body)
	// 生物武器那一份要看得見,否則玩家研究了死亡孢子也不知道它有沒有在作用
	// ——有生物武器傷亡時取代一般人口列，固定保持四列。
	popLine := fmt.Sprintf(uiText(s.b.lang, "bombing.report.population"),
		s.res.PopulationLost, s.res.PopulationBefore)
	if s.res.BioWeaponKills > 0 {
		popLine = fmt.Sprintf(uiText(s.b.lang, "bombing.report.population_bio"),
			s.res.PopulationLost, s.res.BioWeaponKills, s.res.PopulationBefore)
	}
	line(popLine, warn)
	if s.res.DefenderRetaliated {
		line(fmt.Sprintf(uiText(s.b.lang, "bombing.report.retaliated"),
			s.res.AttackerShipsLost), warn)
	} else {
		line(uiText(s.b.lang, "bombing.report.no_retaliation"),
			color.RGBA{160, 220, 160, 255})
	}

	cx, cy, cw, ch := s.contRect()
	fillPanel(dst, float32(cx), float32(cy), float32(cw), float32(ch), color.RGBA{34, 30, 40, 235}, false)
	vector.StrokeRect(dst, float32(cx), float32(cy), float32(cw), float32(ch), 1.5, color.RGBA{150, 140, 170, 255}, false)
	s.continueTextRect().drawCentered(dst, s.fnt, uiText(s.b.lang, "bombing.button.continue"), 14, body)
}

// bombing 進入軌道轟炸畫面。
func (b *sceneBuilder) bombing(res shell.GroundBombardResult) (origScreen, error) {
	if b.session == nil {
		return nil, fmt.Errorf("%s", uiText(b.lang, "bombing.error.no_session"))
	}
	return newBombingScreen(b, res), nil
}
