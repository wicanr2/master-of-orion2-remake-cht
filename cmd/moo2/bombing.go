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

// bombing.go:軌道轟炸畫面(原版 `Colony_Bombing_Screen_` @ 0xB4D02)。
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
//	那條紫色階正是**帝國旗色的佔位色**:原版有 `Replace_Colgcbt_Color_With_Player_Colors_`
//	@ 0xB8EFB 這個函式,在載入殖民地/地面戰美術後把保留色換成該帝國的旗色。remake 照做——
//	把這三階換成玩家旗色的三階(見 recolorPlayerRamp)。
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
	bombTitleX, bombTitleY = 319, 10 // Print_Centered_(0x13F, 0x0A)
	// ⚠ 原版那個 y 是文字的**上緣**不是中心:當中心的話 y=10 配一般字高會被畫面上緣切掉,
	//   原版不可能這樣排。remake 的 DrawCentered 是以中心對齊,所以畫的時候要加半個字高。
	//   (同一個結論也套用到地面戰畫面 groundcombat.go 的標題。)
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
var playerColorRamp = [3]color.RGBA{
	{0x6C, 0x30, 0x6C, 0xFF}, // 暗
	{0x8C, 0x48, 0x8C, 0xFF}, // 中
	{0xB0, 0x5C, 0xB0, 0xFF}, // 亮
}

// recolorPlayerRamp 把佔位色的三階換成該旗色的三階(暗 55% / 中 75% / 亮 100%),
// 對應原版 `Replace_Colgcbt_Color_With_Player_Colors_` 的語意。
func recolorPlayerRamp(src *image.RGBA, flag int) *image.RGBA {
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
		for k, p := range playerColorRamp {
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
	flag := 0
	if b.session != nil {
		flag = b.session.FlagColor
	}
	s.grid = ebiten.NewImageFromImage(recolorPlayerRamp(im.AccumulatedRGBA(prov.Embedded), flag))
	return s
}

func (s *bombingScreen) contRect() (int, int, int, int) { return 265, 432, 110, 30 }

func (s *bombingScreen) update(in shell.InputState) *origTransition {
	if !in.ClickReleased {
		return nil
	}
	if x, y, w, h := s.contRect(); hitRect(in, x, y, w, h) {
		return s.b.goTo(s.b.galaxy, "星系主畫面")
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
		dst.DrawImage(s.grid, nil)
	}
	if s.fnt == nil {
		return
	}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{228, 232, 240, 255}
	warn := color.RGBA{240, 150, 120, 255}

	title := s.b.tr("軌道轟炸", "ORBITAL BOMBARDMENT")
	if s.res.ColonyName != "" {
		title += " — " + s.res.ColonyName
	}
	s.fnt.DrawCentered(dst, title, bombTitleX, bombTitleY+bombTitleSize/2, bombTitleSize, gold)

	// 彈著點:每摧毀一座建築畫一個,再加上人口損失數(炸到居住區)。
	// ⚠ 精靈缺圖(COLONY.LBX#1 在這份資料裡是 0 位元組,見檔頭),用自繪爆點。
	impacts := s.res.BuildingsDestroyed + s.res.PopulationLost
	for i := 0; i < impacts && i < 40; i++ {
		x, y := impactPoint(i)
		vector.DrawFilledCircle(dst, x, y, 9, color.RGBA{255, 170, 60, 190}, true)
		vector.DrawFilledCircle(dst, x, y, 4, color.RGBA{255, 240, 190, 230}, true)
	}

	// 戰報。壓一條半透明深色條,格線底上直接寫字看不清。
	vector.DrawFilledRect(dst, 0, 60, moo2ScreenW, 118, color.RGBA{6, 4, 8, 180}, false)
	y := 82.0
	line := func(text string, col color.RGBA) {
		s.fnt.DrawCentered(dst, text, 320, y, 14, col)
		y += 24
	}
	line(fmt.Sprintf(s.b.tr("齊射總傷害 %d,估計命中 %d 發", "Salvo damage %d, est. %d hits"),
		s.res.TotalDamage, s.res.Hits), body)
	line(fmt.Sprintf(s.b.tr("摧毀建築 %d 座(該殖民地尚存 %d 座)", "%d buildings destroyed (%d still standing)"),
		s.res.BuildingsDestroyed, s.res.BuildingsRemaining), body)
	// 生物武器那一份要看得見,否則玩家研究了死亡孢子也不知道它有沒有在作用
	// ——但只在真的殺到人時才多印一段,沒有生物武器的一般轟炸維持原本的四行版面
	// (這個報告框高 118px、行距 24,第五行會掉出框外)。
	popLine := fmt.Sprintf(s.b.tr("人口損失 %d(轟炸前 %d)", "%d population lost (was %d)"),
		s.res.PopulationLost, s.res.PopulationBefore)
	if s.res.BioWeaponKills > 0 {
		popLine = fmt.Sprintf(s.b.tr("人口損失 %d(生物武器 %d;轟炸前 %d)", "%d population lost (%d to bioweapons; was %d)"),
			s.res.PopulationLost, s.res.BioWeaponKills, s.res.PopulationBefore)
	}
	line(popLine, warn)
	if s.res.DefenderRetaliated {
		line(fmt.Sprintf(s.b.tr("敵方軌道防禦反擊,我方損失 %d 艘", "Orbital defenses returned fire; %d ships lost"),
			s.res.AttackerShipsLost), warn)
	} else {
		line(s.b.tr("敵方無存活的軌道防禦,未遭反擊", "No orbital defenses survived; no return fire"),
			color.RGBA{160, 220, 160, 255})
	}

	cx, cy, cw, ch := s.contRect()
	vector.DrawFilledRect(dst, float32(cx), float32(cy), float32(cw), float32(ch), color.RGBA{34, 30, 40, 235}, false)
	vector.StrokeRect(dst, float32(cx), float32(cy), float32(cw), float32(ch), 1.5, color.RGBA{150, 140, 170, 255}, false)
	s.fnt.DrawCentered(dst, s.b.tr("繼續", "CONTINUE"), float64(cx+cw/2), float64(cy+ch/2), 14, body)
}

// bombing 進入軌道轟炸畫面。
func (b *sceneBuilder) bombing(res shell.GroundBombardResult) (origScreen, error) {
	if b.session == nil {
		return nil, fmt.Errorf("無對局")
	}
	return newBombingScreen(b, res), nil
}
