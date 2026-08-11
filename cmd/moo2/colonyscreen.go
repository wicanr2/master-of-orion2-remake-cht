package main

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// 單一殖民地管理畫面(原版 `Colony_Screen_`,module 74)。
//
// 為什麼要有:反組譯確認原版的 `Colony` 與 `Colony_Summary` 是**兩個不同畫面**,
// 星圖主畫面直接跳單一殖民地畫面(`Main_Screen_ → Do_Colony_Screen_`)。
// remake 先前只有總覽表格,沒有「進到某顆星球管理它」這一層——玩家要調職業或換建造,
// 只能在總覽的小格子上點,而建造更是一次只能排一項。見 docs/re/01-gap-report.md Part A #1。
//
// 這個畫面承載三件原版在殖民地畫面做的事:
//  1. 職業分配(原版 `Add_Job_Fields_`:農/工/科三欄)
//  2. **7 格建造佇列**(原版 `Add_Build_Queue_Fields_`:queue_fields[7])
//  3. 已建建築與行星環境一覽
//
// ============ 版面(2026-08-07 改用原版框架 + 反組譯座標)============
//
// ⚠ 這裡原本寫「remake 沒有原版 COLONY.LBX 的版面資料(repo 不含版權資產),故疊在
// colsum.lbx 上自繪」——**兩句都不對**,而且第二句是會擋死後續工作的錯:
//   ① 版面資料在**執行檔**裡,不在 LBX 裡。`Add_Job_Field_For_` @ 0xBCB4B 直接給出
//      職業欄的座標(見下)。
//   ② 畫面框架的美術也在,只是不在 COLONY.LBX——是 **COLPUPS.LBX 資產 5**
//      (640×480,中段透明,那裡原版畫行星表面)。COLONY.LBX#6 是道路動畫,不是底圖。
//
// 三個獨立來源互證(與 NEW GAME 那次同一套方法):
//
//	反組譯 `Add_Job_Field_For_` @ 0xBCB4B(由 `Add_Job_Fields_` 迴圈 i=0..2 呼叫):
//	    ecx = 0x136(310)、push 0x1FE(510)  → 欄位 x = 310..510
//	    ebx = i*0x1E(30);ebx += 0x3E(62)  → 第 i 列 y = 62 + 30i(62 / 92 / 122)
//	量 COLPUPS.LBX#5 的深色內凹面板:
//	    左 x 7..115 y 28..148 / 中 x 126..304 y 31..140 / **右 x 308..511 y 31..146**
//	    → 右面板與職業欄的 310..510 逐像素吻合(各留 1–2px 內縮),三列 ×30 正好塞滿。
//
// 其餘量自同一張框架圖:上方資訊列 y 15..158、中段行星表面 y 159..423(透明)、
// 右下 LEADERS y 424..449 / RETURN y 456..479,兩顆都在 x 551..637。
//
// ⚠ 2026-08-07 把上面那段「誠實留白」做掉了。先前中段放的是建造佇列與可建清單,
// 那是 **remake 自己的版面**;原版那一塊是**行星表面 + 建築依格點擺放**,佇列則是另外
// 一張彈出視窗。現在:
//   中段 → `drawColonySurface`(格點與建築圖全是原版真值,見 colonysurface.go)
//   佇列 → `buildqueue.go`(`Build_Queue_Popup_` @ 0xB4041,座標同樣是反組譯真值)
//   入口 → 框架上那顆 CHANGE(原版它就是「換要蓋什麼」),先前畫成灰的沒接功能。
//
// 順帶:行星表面**不是關在中段那個框裡**——`Draw_Colony_Screen_` 一開場就是
// `C_Anims(1, 0, 639, 479)`,地表是整個 640×480 的底圖,資訊面板疊在它上面。
// remake 目前只畫格線 + 建築,地表底圖的來源 LBX 還沒追到。

const (
	colScreenW = 640.0
	colScreenH = 480.0

	// --- 原版框架(COLPUPS.LBX#5)與其內部面板 ---
	colChromeLBX   = "colpups.lbx"
	colChromeAsset = 5
	colChromePal   = "colbldg.lbx" // 框架自己沒有調色盤,借建築圖那組(bombing.go 同一條鏈)

	colTopBarY0, colTopBarY1 = 15, 158  // 上方資訊列
	colFieldY0, colFieldY1   = 159, 423 // 中段:原版畫行星表面,remake 放佇列/清單

	colPanelLX, colPanelLY, colPanelLW, colPanelLH = 7, 28, 109, 121   // 左:行星
	colPanelMX, colPanelMY, colPanelMW, colPanelMH = 126, 31, 179, 110 // 中:產出
	colPanelRX, colPanelRY, colPanelRW, colPanelRH = 308, 31, 204, 116 // 右:職業分配

	// 職業欄:反組譯真值(x 310..510、y 62+30i)。
	colJobX0, colJobX1 = 310, 510
	colJobY0           = 62
	colJobStep         = 30

	colBtnX, colBtnW = 551, 87
	// CHANGE / BUY(量自框架圖,放大 4 倍比對 plate 邊緣)。remake 兩顆都還沒接功能。
	colChangeX, colChangeW   = 519, 61
	colBuyX, colBuyW         = 588, 40
	colChangeY, colChangeH   = 123, 20
	colLeadersY, colLeadersH = 424, 26
	colReturnY, colReturnH   = 456, 24
)

var (
	colPanelBg   = color.RGBA{18, 24, 44, 255}
	colPanelEdge = color.RGBA{92, 120, 190, 255}
	colTitleCol  = color.RGBA{240, 220, 120, 255}
	colBodyCol   = color.RGBA{214, 222, 238, 255}
	colDimCol    = color.RGBA{140, 152, 178, 255}
	colSlotBg    = color.RGBA{30, 40, 70, 255}
	colCurBg     = color.RGBA{40, 70, 120, 255}
	colOkCol     = color.RGBA{140, 220, 150, 255}
	colWarnCol   = color.RGBA{235, 170, 110, 255}
)

// colonyScreen 建單一殖民地畫面。操作的殖民地由 b.colonyIdx 指定。
func (b *sceneBuilder) colonyScreen() (*overlayScreen, error) {
	idx := b.colonyIdx
	if b.session == nil || idx < 0 || idx >= len(b.session.PlayerColonies) {
		return b.colonySummary() // 索引失效(殖民地被打掉等)就退回總覽,不留在壞畫面
	}

	hits := []hitRegion{
		{colBtnX, colReturnY, colBtnW, colReturnH, "return"},
		{colBtnX, colLeadersY, colBtnW, colLeadersH, "leaders"},
	}
	// 職業分配三列:座標是反組譯真值(見檔頭)。點一下該列 +1,從別的職業挪一名過來。
	for i, act := range []string{"job_f", "job_w", "job_s"} {
		hits = append(hits, hitRegion{
			colJobX0, colJobY0 + i*colJobStep, colJobX1 - colJobX0, colJobStep - 2, act,
		})
	}
	// CHANGE:原版就是「換要蓋什麼」——點下去開建造彈出視窗(buildqueue.go)。
	// 先前這顆是畫成灰的沒接功能,佇列被塞在中段;中段還給行星表面之後,它才有事做。
	hits = append(hits, hitRegion{colChangeX, colChangeY, colChangeW, colChangeH, "build"})
	// 中段的行星表面:點格子也開建造視窗(原版點空格是選建築槽,remake 還沒有那一層)。
	hits = append(hits, hitRegion{0, colFieldY0, int(colScreenW), colFieldY1 - colFieldY0, "build"})

	onAction := func(a string) *origTransition {
		s := b.session
		switch {
		case a == "return":
			return b.goTo(b.colonySummary, "殖民地總覽")
		case a == "leaders":
			// 原版這顆直接開殖民地領袖分頁；保留殖民地索引讓下一步指派有明確目標。
			b.officerTab = 0
			b.colonyIdx = idx
			return b.goTo(b.officer, "軍官列表")
		case a == "job_f":
			s.ShiftColonyJob(idx, "w", "f")
		case a == "job_w":
			s.ShiftColonyJob(idx, "f", "w")
		case a == "job_s":
			s.ShiftColonyJob(idx, "w", "s")
		case a == "build":
			sc, err := b.buildQueuePopup()
			if err != nil {
				return nil
			}
			return &origTransition{next: sc}
		default:
			return nil
		}
		return b.goTo(b.colonyScreen, "殖民地") // 重繪反映新狀態
	}

	s, err := loadOverlayScreen(b.res, "colsum.lbx", 0, b.lang, b.fnt, "colony.tsv",
		nil, colBodyCol, 13, hits, onAction, nil)
	if err != nil {
		return nil, err
	}
	s.postDraw = func(dst *ebiten.Image) { b.drawColonyScreen(dst, idx) }
	return s, nil
}

// colonyBuildChoices 回傳可排進佇列的建造選項:玩家科技已解鎖、且該殖民地還沒蓋的。
// 「不建造」(空字串)不列入——要清空用點佇列格移除。
func (b *sceneBuilder) colonyBuildChoices() []shell.ColonyBuild {
	if b.session == nil {
		return nil
	}
	idx := b.colonyIdx
	all := b.session.AvailableBuildOptions()
	queued := map[string]bool{}
	for _, q := range b.session.BuildQueueFor(idx) {
		if q.Name != "" {
			queued[q.Name] = true
		}
	}
	out := make([]shell.ColonyBuild, 0, len(all))
	for _, o := range all {
		if o.Name == "" || queued[o.Name] {
			continue
		}
		if b.session.ColonyHasBuilding(idx, o.Name) {
			continue
		}
		out = append(out, o)
	}
	return out
}

// drawColonyScreen 畫整個殖民地畫面。
//
// ⚠ 順序踩過兩次坑,現在照原版 `Draw_Colony_Screen_` @ 0xBED21 的次序來:
//
//	地表兩層 → 框架(`Draw_Colony_Info_Background`)→ 建築(`Draw_Colony_Bldgs`)→ 資訊面板
//
// ① 框架(COLPUPS.LBX#5)只有**中段**(y 159..423,原版的行星表面)是透明的,上方那三個
// 資訊面板是**不透明的深色星空紋理**。內容先畫、框架後蓋,上半部的字會整片不見。
// ② 反過來把地表也擺在框架後面同樣不行:地表底層是整片 640×480 的星空,會把上方面板
// 蓋成一片雜點。地表必須在**框架之前**,建築則在**框架之後**——原版就是這樣夾的。
func (b *sceneBuilder) drawColonyScreen(dst *ebiten.Image, idx int) {
	if b.fnt == nil || b.session == nil {
		return
	}
	sess := b.session
	c := sess.PlayerColonies[idx]

	fillPanel(dst, 0, 0, colScreenW, colScreenH, colPanelBg, false)
	b.drawColonyTerrain(dst, idx)
	// 道路夾在地形與框架之間,和原版 `Draw_Colony_Screen_` 的呼叫序一致。
	surf := b.colonySurfaceLayout(idx)
	b.drawColonyRoads(dst, surf.roads)
	if im := b.colonyChrome(); im != nil {
		drawPanelImage(dst, im, &ebiten.DrawImageOptions{})
	}
	b.drawColonyBuildings(dst, surf)
	b.drawColonySatellites(dst, idx)
	b.drawColonyTopBar(dst, idx, c)

	// 兩顆鈕的英文(LEADERS / RETURN)烘在框架圖上,照既有做法擦底疊中文。
	drawBtn := func(y, h int, zh string) {
		fillPanel(dst, float32(colBtnX+3), float32(y+3),
			float32(colBtnW-6), float32(h-6), color.RGBA{72, 76, 84, 255}, false)
		b.fnt.DrawCentered(dst, zh, float64(colBtnX+colBtnW/2), float64(y+h/2), 12, colBodyCol)
	}
	drawBtn(colLeadersY, colLeadersH, b.tr("領袖", "LEADERS"))
	drawBtn(colReturnY, colReturnH, b.tr("返回", "RETURN"))

	// CHANGE / BUY 也是烘在框架上的英文。
	// **CHANGE 已接上**(2026-08-07):原版它就是「換要蓋什麼」,中段還給行星表面之後
	// 佇列搬進 buildqueue.go 那張彈出視窗,這顆正好是入口。
	// BUY(花 BC 立即完工)仍未實作,沿用既有做法:不給熱區 + 字畫成灰的,
	// 而不是留英文、也不是給一顆點了沒反應的中文鈕。
	for _, btn := range []struct {
		x, y, w, h int
		zh         string
		on         bool
	}{
		{colChangeX, colChangeY, colChangeW, colChangeH, b.tr("更換", "CHANGE"), true},
		{colBuyX, colChangeY, colBuyW, colChangeH, b.tr("購買", "BUY"), false},
	} {
		face, ink := color.RGBA{112, 116, 120, 255}, color.RGBA{72, 74, 78, 255}
		if btn.on { // CHANGE 接上建造視窗了,不再畫成灰的
			face, ink = color.RGBA{74, 88, 74, 255}, color.RGBA{225, 240, 225, 255}
		}
		fillPanel(dst, float32(btn.x+3), float32(btn.y+3),
			float32(btn.w-6), float32(btn.h-6), face, false)
		b.fnt.DrawCentered(dst, btn.zh, float64(btn.x+btn.w/2), float64(btn.y+btn.h/2), 11, ink)
	}
}

// colonyChrome 解出原版殖民地畫面的框架(COLPUPS.LBX#5)。
// 它自己沒有調色盤,借 COLBLDG.LBX#0 的(同一組殖民地美術;bombing.go 用的是同一條鏈)。
// 載不動就回 nil——畫面會退回純自繪版,不會壞。
func (b *sceneBuilder) colonyChrome() *ebiten.Image {
	if b.colChrome != nil {
		return b.colChrome
	}
	pal, err := decodeAsset(b.res, colChromePal, 0)
	if err != nil || pal.Embedded == nil {
		return nil
	}
	im, err := decodeAsset(b.res, colChromeLBX, colChromeAsset)
	if err != nil || len(im.Frames) == 0 {
		return nil
	}
	b.colChrome = ebiten.NewImageFromImage(im.Frames[0].ToRGBA(pal.Embedded, true))
	return b.colChrome
}

// drawColonyTopBar 畫上方資訊列的三個面板:左=行星、中=產出、右=職業分配。
// 三個面板的座標量自原版框架,職業列的座標是反組譯真值(見檔頭)。
func (b *sceneBuilder) drawColonyTopBar(dst *ebiten.Image, idx int, c engine.ColonyState) {
	sess := b.session

	// --- 左面板:殖民地名 + 行星環境 + 人口 ---
	lx := float64(colPanelLX + 5)
	name := fmt.Sprintf(b.tr("殖民地 %d", "Colony %d"), idx+1)
	// 名字取自**殖民地座落的行星**(同星系可以有多個殖民地,取星的代表行星會撞名)。
	if n := sess.ColonyName(idx); n != "" {
		name = n
	}
	b.fnt.Draw(dst, name, lx, float64(colPanelLY+6), 13, colTitleCol)
	if p, ok := sess.ColonyPlanetData(idx); ok {
		rows := colonyPlanetRows(b.lang, p)
		for i, r := range rows {
			if i >= 5 {
				break
			}
			b.fnt.Draw(dst, r, lx, float64(colPanelLY+26+i*16), 10, colBodyCol)
		}
	}
	// 同化進度只在**真的有未同化人口**時才畫——沒有征服來的人口時這一行沒有意義,
	// 而且那也讓沒被征服過的殖民地畫面保持原樣。
	if turns, ok := sess.AssimilationRemainingTurns(idx); ok {
		b.fnt.Draw(dst, fmt.Sprintf(b.tr("未同化 %d ／還需 %d 回合", "Unassimilated %d / %d turns"),
			c.UnassimilatedPop, turns), lx, float64(colPanelLY+colPanelLH-32), 10, colWarnCol)
	}
	b.fnt.Draw(dst, fmt.Sprintf(b.tr("人口 %d/%d", "Pop %d/%d"), c.Population, c.PopMax),
		lx, float64(colPanelLY+colPanelLH-16), 11, colOkCol)

	// --- 中面板:本回合產出 + 已建建築 ---
	// 產出優先用上一回合的結算(與總覽同一份資料);第 1 回合還沒結算過就即時算一次,
	// 讓玩家一進畫面就看得到數字,而不是「結束回合後才顯示」。
	var co engine.ColonyOutput
	if out := sess.LastPlayerOutput; idx < len(out.Colonies) {
		co = out.Colonies[idx]
	} else {
		co = engine.RunColonyTurn(c)
	}
	mx := float64(colPanelMX + 6)
	foodCol := colOkCol
	if co.FoodSurplus < 0 {
		foodCol = colWarnCol
	}
	b.fnt.Draw(dst, fmt.Sprintf(b.tr("食物 %+d", "Food %+d"), co.FoodSurplus), mx, float64(colPanelMY+6), 11, foodCol)
	b.fnt.Draw(dst, fmt.Sprintf(b.tr("工業 %d", "Ind %d"), co.NetIndustry), mx+88, float64(colPanelMY+6), 11, colBodyCol)
	b.fnt.Draw(dst, fmt.Sprintf(b.tr("研究 %d", "Res %d"), co.Research), mx, float64(colPanelMY+22), 11, colBodyCol)
	// MoralePercent 是「對產出的百分點調整」(每格笑臉 +10、哭臉 -10),0 = 無加成也無懲罰,
	// 不是「士氣只有 0 分」。標成「士氣修正」避免誤讀(見 engine.ColonyState.MoralePercent)。
	b.fnt.Draw(dst, fmt.Sprintf(b.tr("士氣 %+d%%", "Morale %+d%%"), c.MoralePercent), mx+88, float64(colPanelMY+22), 11, colBodyCol)

	b.fnt.Draw(dst, b.tr("已建建築", "Buildings"), mx, float64(colPanelMY+44), 11, colTitleCol)
	names := sess.ColonyBuildingNames(idx)
	if len(names) == 0 {
		b.fnt.Draw(dst, b.tr("(無)", "(none)"), mx, float64(colPanelMY+60), 10, colDimCol)
	} else {
		sort.Strings(names)
		displayNames := names
		separator := "、"
		if b.lang != i18n.Traditional {
			displayNames = make([]string, len(names))
			for i, name := range names {
				displayNames[i] = colonyBuildingLabel(b.lang, name)
			}
			separator = ", "
		}
		for i, ln := range b.fnt.Wrap(strings.Join(displayNames, separator), 10, float64(colPanelMW-14)) {
			if i >= 4 {
				break
			}
			b.fnt.Draw(dst, ln, mx, float64(colPanelMY+60+i*14), 10, colOkCol)
		}
	}

	// --- 右面板:職業分配三列(x/y 為反組譯真值)---
	jobs := []struct {
		label string
		n     int
	}{{b.tr("農夫", "Farmers"), c.Farmers}, {b.tr("工人", "Workers"), c.Workers},
		{b.tr("科學家", "Scientists"), c.Scientists}}
	for i, j := range jobs {
		y := colJobY0 + i*colJobStep
		fillPanel(dst, float32(colJobX0), float32(y),
			float32(colJobX1-colJobX0), float32(colJobStep-4), colSlotBg, false)
		vector.StrokeRect(dst, float32(colJobX0), float32(y),
			float32(colJobX1-colJobX0), float32(colJobStep-4), 1, colPanelEdge, false)
		b.fnt.Draw(dst, j.label, float64(colJobX0+10), float64(y+6), 12, colTitleCol)
		b.fnt.Draw(dst, fmt.Sprintf(b.tr("%d 人", "%d"), j.n), float64(colJobX0+70), float64(y+6), 12, colBodyCol)
		b.fnt.Draw(dst, b.tr("點此 +1", "click +1"), float64(colJobX1-56), float64(y+7), 10, colDimCol)
	}
}
