package main

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
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
// 版面說明:remake 沒有原版 COLONY.LBX 的版面資料(repo 不含版權資產),故這個畫面
// 疊在既有的 colsum.lbx 背景上、以自繪面板覆蓋,不宣稱像素對齊原版。
// 對齊的是**結構與流程**(哪些操作在同一個畫面上、佇列幾格、完工怎麼接),
// 那部分有反組譯硬證;版面留待取得 COLONY.LBX 版面座標後再對。

const (
	colScreenW = 640.0
	colScreenH = 480.0

	colQueueX = 330.0 // 右欄:建造佇列 + 可建清單
	colQueueY = 84.0
	colRowH   = 22.0

	// 佇列 7 格佔 y=84..238;其下依序是清單標題、翻頁鈕、清單本體(到返回鈕之上)。
	colListTitleY = 246.0
	colListPageY  = 264.0
	colListY      = 292.0 // 可建清單起始 y
	colListRows   = 7     // 可建清單一頁顯示幾列(7×22=154,底在 446,不撞返回鈕)
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
		{552, 448, 76, 22, "return"},
		// 職業分配:三個「+1」按鈕,各自從工人挪一名過來(工人自己那顆從農夫挪)。
		{28, 150, 90, 20, "job_f"},
		{28, 174, 90, 20, "job_w"},
		{28, 198, 90, 20, "job_s"},
	}
	// 佇列 7 格:點一下移除該格。
	for i := 0; i < shell.BuildQueueTotalSlots; i++ {
		hits = append(hits, hitRegion{
			int(colQueueX), int(colQueueY + float64(i)*colRowH), 286, int(colRowH) - 2,
			fmt.Sprintf("qdel%d", i),
		})
	}
	// 可建清單:點一下排進佇列。
	for i := 0; i < colListRows; i++ {
		hits = append(hits, hitRegion{
			int(colQueueX), int(colListY + float64(i)*colRowH), 286, int(colRowH) - 2,
			fmt.Sprintf("qadd%d", i),
		})
	}
	hits = append(hits,
		hitRegion{int(colQueueX), int(colListPageY), 60, 20, "listup"},
		hitRegion{int(colQueueX) + 70, int(colListPageY), 60, 20, "listdown"},
	)

	onAction := func(a string) *origTransition {
		s := b.session
		switch {
		case a == "return":
			return b.goTo(b.colonySummary, "殖民地總覽")
		case a == "job_f":
			s.ShiftColonyJob(idx, "w", "f")
		case a == "job_w":
			s.ShiftColonyJob(idx, "f", "w")
		case a == "job_s":
			s.ShiftColonyJob(idx, "w", "s")
		case a == "listup":
			b.colonyListTop -= colListRows
			if b.colonyListTop < 0 {
				b.colonyListTop = 0
			}
		case a == "listdown":
			opts := b.colonyBuildChoices()
			if b.colonyListTop+colListRows < len(opts) {
				b.colonyListTop += colListRows
			}
		case strings.HasPrefix(a, "qdel"):
			var pos int
			fmt.Sscanf(a, "qdel%d", &pos)
			s.DequeueBuild(idx, pos)
		case strings.HasPrefix(a, "qadd"):
			var row int
			fmt.Sscanf(a, "qadd%d", &row)
			opts := b.colonyBuildChoices()
			if i := b.colonyListTop + row; i >= 0 && i < len(opts) {
				s.EnqueueBuild(idx, opts[i].Name, opts[i].Cost)
			}
		default:
			return nil
		}
		return b.goTo(b.colonyScreen, "殖民地") // 重繪反映新狀態
	}

	s, err := loadOverlayScreen(b.res, "colsum.lbx", 0, b.lang, b.fnt, "assets/i18n/colony.tsv",
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

// drawColonyScreen 畫整個殖民地畫面(自繪面板蓋掉底下的總覽表格)。
func (b *sceneBuilder) drawColonyScreen(dst *ebiten.Image, idx int) {
	if b.fnt == nil || b.session == nil {
		return
	}
	sess := b.session
	c := sess.PlayerColonies[idx]

	vector.DrawFilledRect(dst, 0, 0, colScreenW, colScreenH, colPanelBg, false)
	vector.StrokeRect(dst, 4, 4, colScreenW-8, colScreenH-8, 2, colPanelEdge, false)

	name := fmt.Sprintf("殖民地 %d", idx+1)
	if star := sess.PlayerColonyStarIndex(idx); star >= 0 && star < len(sess.Planets) {
		if pn := sess.Planets[star].Name; pn != "" {
			name = pn
		}
	}
	b.fnt.DrawCentered(dst, name, colScreenW/2, 30, 17, colTitleCol)

	b.drawColonyLeftPanel(dst, idx, c)
	b.drawColonyQueue(dst, idx)
	b.drawColonyBuildList(dst)

	// RETURN
	vector.DrawFilledRect(dst, 552, 448, 76, 22, colSlotBg, false)
	vector.StrokeRect(dst, 552, 448, 76, 22, 1, colPanelEdge, false)
	b.fnt.DrawCentered(dst, "返回", 590, 459, 13, colBodyCol)
}

// drawColonyLeftPanel 畫左欄:行星環境、人口與職業分配、產出、已建建築。
func (b *sceneBuilder) drawColonyLeftPanel(dst *ebiten.Image, idx int, c engine.ColonyState) {
	sess := b.session
	x := 24.0
	b.fnt.Draw(dst, "行星", x, 62, 14, colTitleCol)

	env := "—"
	if star := sess.PlayerColonyStarIndex(idx); star >= 0 && star < len(sess.Planets) {
		p := sess.Planets[star]
		env = fmt.Sprintf("%s / %s / 礦產%s / 重力%s", p.Climate, p.Size, p.Mineral, p.Gravity)
		if sp := gamedata.PlanetSpecialName(p.SpecialID); sp != "" {
			env += " / ★" + sp // 特殊物產:金礦/寶石礦的收入、遠古文物的研究都靠它
		}
	}
	b.fnt.Draw(dst, env, x, 82, 11, colBodyCol)
	b.fnt.Draw(dst, fmt.Sprintf("人口 %d / %d", c.Population, c.PopMax), x, 100, 12, colBodyCol)

	// 職業分配:三列,每列一個可點的「+1」。
	b.fnt.Draw(dst, "職業分配(點擊 +1)", x, 128, 13, colTitleCol)
	jobs := []struct {
		label string
		n     int
		y     float64
	}{
		{"農夫", c.Farmers, 150},
		{"工人", c.Workers, 174},
		{"科學家", c.Scientists, 198},
	}
	for _, j := range jobs {
		vector.DrawFilledRect(dst, float32(x+4), float32(j.y), 90, 20, colSlotBg, false)
		vector.StrokeRect(dst, float32(x+4), float32(j.y), 90, 20, 1, colPanelEdge, false)
		b.fnt.Draw(dst, fmt.Sprintf("%s %d", j.label, j.n), x+12, j.y+4, 12, colBodyCol)
	}

	// 產出:優先用上一回合的結算(與總覽同一份資料);第 1 回合還沒結算過就即時算一次,
	// 讓玩家一進畫面就看得到數字,而不是「結束回合後才顯示」。
	b.fnt.Draw(dst, "本回合產出", x, 244, 13, colTitleCol)
	var co engine.ColonyOutput
	if out := sess.LastPlayerOutput; idx < len(out.Colonies) {
		co = out.Colonies[idx]
	} else {
		co = engine.RunColonyTurn(c)
	}
	foodCol := colOkCol
	if co.FoodSurplus < 0 {
		foodCol = colWarnCol
	}
	b.fnt.Draw(dst, fmt.Sprintf("食物 %+d", co.FoodSurplus), x+4, 264, 12, foodCol)
	for i, l := range []string{
		fmt.Sprintf("工業 %d", co.NetIndustry),
		fmt.Sprintf("研究 %d", co.Research),
		// MoralePercent 是「對產出的百分點調整」(每格笑臉 +10、哭臉 -10),0 = 無加成也無懲罰,
		// 不是「士氣只有 0 分」。標成「士氣修正」避免誤讀(見 engine.ColonyState.MoralePercent)。
		fmt.Sprintf("士氣修正 %+d%%", c.MoralePercent),
	} {
		b.fnt.Draw(dst, l, x+4, 282+float64(i)*18, 12, colBodyCol)
	}

	// 已建建築。
	b.fnt.Draw(dst, "已建建築", x, 352, 13, colTitleCol)
	names := sess.ColonyBuildingNames(idx)
	if len(names) == 0 {
		b.fnt.Draw(dst, "(無)", x+4, 372, 11, colDimCol)
		return
	}
	sort.Strings(names)
	for i, ln := range b.fnt.Wrap(strings.Join(names, "、"), 11, 270) {
		if i >= 5 {
			break
		}
		b.fnt.Draw(dst, ln, x+4, 372+float64(i)*16, 11, colOkCol)
	}
}

// drawColonyQueue 畫右欄上半:7 格建造佇列(第 0 格是建造中,有進度條)。
func (b *sceneBuilder) drawColonyQueue(dst *ebiten.Image, idx int) {
	b.fnt.Draw(dst, fmt.Sprintf("建造佇列(%d 格,點擊移除)", shell.BuildQueueTotalSlots),
		colQueueX, colQueueY-18, 13, colTitleCol)

	q := b.session.BuildQueueFor(idx)
	for i := 0; i < shell.BuildQueueTotalSlots; i++ {
		y := colQueueY + float64(i)*colRowH
		bg := colSlotBg
		if i == 0 {
			bg = colCurBg // 建造中那格高亮
		}
		vector.DrawFilledRect(dst, float32(colQueueX), float32(y), 286, float32(colRowH)-2, bg, false)

		if i >= len(q) || q[i].Name == "" {
			b.fnt.Draw(dst, "—", colQueueX+8, y+4, 11, colDimCol)
			continue
		}
		item := q[i]
		label := item.Name
		if i == 0 && item.Cost > 0 {
			// 建造中:顯示進度與預估剩餘回合(用上一回合的實際投入速率推,不另立公式)。
			label = fmt.Sprintf("%s  %d/%d", item.Name, item.Progress, item.Cost)
			if eta := b.session.BuildETATurns(idx); eta > 0 {
				label += fmt.Sprintf("  約 %d 回合", eta)
			}
			// 進度條
			w := float32(282 * item.Progress / item.Cost)
			if w > 282 {
				w = 282
			}
			vector.DrawFilledRect(dst, float32(colQueueX)+2, float32(y+colRowH)-6, w, 3, colOkCol, false)
		} else if item.Cost > 0 {
			label = fmt.Sprintf("%s  (%d PP)", item.Name, item.Cost)
		}
		b.fnt.Draw(dst, label, colQueueX+8, y+4, 11, colBodyCol)
	}
}

// drawColonyBuildList 畫右欄下半:可建清單(分頁捲動),點一下排進佇列。
func (b *sceneBuilder) drawColonyBuildList(dst *ebiten.Image) {
	opts := b.colonyBuildChoices()
	total := len(opts)
	if b.colonyListTop >= total {
		b.colonyListTop = 0
	}
	page := b.colonyListTop/colListRows + 1
	pages := (total + colListRows - 1) / colListRows
	if pages < 1 {
		pages = 1
	}

	b.fnt.Draw(dst, fmt.Sprintf("可建項目(%d/%d 頁,點擊排入)", page, pages),
		colQueueX, colListTitleY, 13, colTitleCol)
	for i, lbl := range []string{"上一頁", "下一頁"} {
		bx := colQueueX + float64(i)*70
		vector.DrawFilledRect(dst, float32(bx), float32(colListPageY), 60, 20, colSlotBg, false)
		vector.StrokeRect(dst, float32(bx), float32(colListPageY), 60, 20, 1, colPanelEdge, false)
		b.fnt.DrawCentered(dst, lbl, bx+30, colListPageY+10, 11, colBodyCol)
	}
	if total == 0 {
		// 開局科技少、可蓋的又都蓋了或已排進佇列時清單就是空的——說清楚,不要留一片空白讓人以為畫面壞了。
		b.fnt.Draw(dst, "目前沒有可排入的項目", colQueueX, colListY+4, 12, colDimCol)
		b.fnt.Draw(dst, "(已建或已在佇列中的不重複列出;完成研究會解鎖更多)",
			colQueueX, colListY+24, 11, colDimCol)
		return
	}

	for i := 0; i < colListRows; i++ {
		y := colListY + float64(i)*colRowH
		j := b.colonyListTop + i
		if j >= total {
			break
		}
		vector.DrawFilledRect(dst, float32(colQueueX), float32(y), 286, float32(colRowH)-2, colSlotBg, false)
		o := opts[j]
		txt := o.Name
		if o.Cost > 0 {
			txt = fmt.Sprintf("%s  (%d PP)", o.Name, o.Cost)
		} else if o.Name == shell.TradeGoodsBuildName {
			txt = o.Name + "  (工業轉現金)"
		}
		if _, isSpecial := gamedata.SpecialActionByNameZH(o.Name); isSpecial {
			txt += "  ※可重複"
		}
		b.fnt.Draw(dst, txt, colQueueX+8, y+4, 11, colBodyCol)
	}
}
