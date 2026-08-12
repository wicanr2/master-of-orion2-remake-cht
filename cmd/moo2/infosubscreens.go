package main

import (
	"fmt"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// INFO 畫面的五個子畫面(對齊原版結構)。
//
// 原版結構由反組譯 Orion2.exe 的除錯符號表確認(2026-08-06):INFO 不是跳板,而是
// 「單一畫面 + 5 個子畫面」,各有獨立繪製函式——
//
//	Draw_History_Subscreen_        歷史圖表(國力折線)
//	Draw_Tech_Review_Subscreen_    科技總覽
//	Draw_Race_Stats_Subscreen_     種族統計
//	Draw_Turn_Summary_Subscreen_   回合摘要
//	Draw_Reference_Main_/_Category_/_How_To_Subscreen_  內建參考
//
// 詳見 docs/re/01-gap-report.md Part A。先前 remake 把「Tech Review」誤接成研究選擇
// 畫面(點一下就退回星系,使用者實測 issue #5-2),歷史圖表則完全沒接(issue #5-1)。
//
// 版面:原版右側面板約 x=210..630、y=40..425。以下各子畫面都畫在這個框內。

const (
	infoPanelX = 214.0
	infoPanelY = 44.0
	infoPanelW = 412.0
	infoPanelH = 380.0
)

// infoTabTitles 是五個分頁的中文標題(順序同 b.infoTab)。
var infoTabTitles = []string{"歷史圖表", "科技總覽", "種族統計", "回合摘要", "參考資料"}

// infoTabTitlesEn 是原版 INFO 畫面那五個分頁的英文名。
var infoTabTitlesEn = []string{"HISTORY", "TECHNOLOGY", "RACES", "TURN SUMMARY", "REFERENCE"}

// drawInfoSubscreen 依 b.infoTab 把對應子畫面內容加進 overlayScreen。
func (b *sceneBuilder) drawInfoSubscreen(s *overlayScreen) {
	tab := b.infoTab
	if tab < 0 || tab >= len(infoTabTitles) {
		tab = 0
	}
	gold := infoTitleCol
	b.appendInfoCentered(s, infoTitleTextRect(), b.tr(infoTabTitles[tab], infoTabTitlesEn[tab]), 15, gold)
	if b.session == nil {
		b.appendInfoCentered(s, infoCenteredTextRect(int(infoPanelY)+48, 24),
			b.tr("尚無對局資料", "No game data yet"), 12, infoBodyCol)
		return
	}
	switch tab {
	case 0:
		b.infoHistory(s)
	case 1:
		b.infoTechReview(s)
	case 2:
		b.infoRaceStats(s)
	case 3:
		b.infoTurnSummary(s)
	default:
		b.infoReference(s)
	}
}

// --- 分頁 0:歷史圖表(原版 Draw_History_Subscreen_)---

// empireLineColors 是折線圖各帝國的顏色(index 0 = 玩家)。
var empireLineColors = []color.RGBA{
	{20, 60, 170, 255},  // 玩家:深藍
	{190, 40, 40, 255},  // AI1:深紅
	{20, 130, 140, 255}, // AI2:青
	{30, 130, 50, 255},  // AI3:深綠
	{140, 50, 160, 255}, // AI4:紫
}

// INFO 右側面板是白底(原版 INFO.LBX 資產),故內文一律深色系,不能沿用其他畫面的淺色。
var (
	infoTitleCol = color.RGBA{20, 40, 110, 255}
	infoBodyCol  = color.RGBA{45, 55, 75, 255}
	infoOkCol    = color.RGBA{20, 110, 50, 255}
	infoCurCol   = color.RGBA{25, 85, 165, 255}
	infoDimCol   = color.RGBA{110, 120, 140, 255}
	infoEventCol = color.RGBA{160, 90, 20, 255}
)

// INFO 的內容是 remake 直接繪製的動態資料，不像 overlay 標籤有原版英文底圖可當欄位。
// 所有文字都必須經這組由右側原生面板量得的安全框；不要再直接 append extraText。
func infoPanelBounds() (x, y, w, h int) {
	return int(infoPanelX), int(infoPanelY), int(infoPanelW), int(infoPanelH)
}

func infoTextRect(x, y, w, h int) textSafeRect {
	return textSafeRect{x: x, y: y, w: w, h: h}
}

func infoCenteredTextRect(y, h int) textSafeRect {
	return infoTextRect(int(infoPanelX)+8, y, int(infoPanelW)-16, h)
}

func infoContentTextRect(y, h int) textSafeRect {
	return infoTextRect(int(infoPanelX)+16, y, int(infoPanelW)-32, h)
}

func infoTitleTextRect() textSafeRect {
	return infoCenteredTextRect(int(infoPanelY)+4, 24)
}

func infoTechTopicTextRect(column, y, h int) textSafeRect {
	if column < 0 || column > 1 {
		return textSafeRect{}
	}
	return infoTextRect(int(infoPanelX)+16+column*200, y, 180, h)
}

func infoRaceStatTextRect(column, y, h int) textSafeRect {
	xs := [...]int{int(infoPanelX) + 16, int(infoPanelX) + 150, int(infoPanelX) + 215, int(infoPanelX) + 268, int(infoPanelX) + 325}
	rights := [...]int{int(infoPanelX) + 142, int(infoPanelX) + 207, int(infoPanelX) + 260, int(infoPanelX) + 317, int(infoPanelX) + int(infoPanelW) - 16}
	if column < 0 || column >= len(xs) {
		return textSafeRect{}
	}
	return infoTextRect(xs[column], y, rights[column]-xs[column], h)
}

func infoRaceRelationTextRect(y int) textSafeRect {
	return infoContentTextRect(y, 12)
}

func infoSummaryLabelTextRect(y int) textSafeRect {
	return infoTextRect(int(infoPanelX)+24, y, 110, 12)
}

func infoSummaryValueTextRect(y int) textSafeRect {
	return infoTextRect(int(infoPanelX)+150, y, int(infoPanelW)-166, 12)
}

func infoReferenceTextRect(column, y int) textSafeRect {
	if column == 0 {
		return infoTextRect(int(infoPanelX)+24, y, 180, 12)
	}
	return infoTextRect(int(infoPanelX)+220, y, int(infoPanelW)-236, 12)
}

func infoHistoryMaxTextRect() textSafeRect {
	return infoTextRect(int(infoPanelX)+6, int(infoPanelY)+62, 30, 12)
}

func infoHistoryZeroTextRect() textSafeRect {
	return infoTextRect(int(infoPanelX)+20, int(infoPanelY)+310, 20, 12)
}

func infoHistoryTurnTextRect(last bool) textSafeRect {
	if last {
		return infoTextRect(int(infoPanelX)+int(infoPanelW)-70, int(infoPanelY)+324, 54, 12)
	}
	return infoTextRect(int(infoPanelX)+40, int(infoPanelY)+324, 120, 12)
}

const infoHistoryLegendSlots = 6

func infoHistoryLegendTextRect(i int) textSafeRect {
	if i < 0 || i >= infoHistoryLegendSlots {
		return textSafeRect{}
	}
	return infoTextRect(int(infoPanelX)+40+(i%2)*190, int(infoPanelY)+int(infoPanelH)-42+(i/2)*13, 170, 12)
}

func (b *sceneBuilder) appendInfoLeft(s *overlayScreen, r textSafeRect, text string, size float64, col color.RGBA) {
	s.extras = append(s.extras, r.leftExtras(b.fnt, text, size, col)...)
}

func (b *sceneBuilder) appendInfoCentered(s *overlayScreen, r textSafeRect, text string, size float64, col color.RGBA) {
	s.extras = append(s.extras, r.centeredExtras(b.fnt, text, size, col)...)
}

func (b *sceneBuilder) infoHistory(s *overlayScreen) {
	body := infoBodyCol
	metric := shell.HistoryMetric(b.infoHistoryMetric)
	series, turns := b.session.HistorySeries(metric)

	b.appendInfoCentered(s, infoCenteredTextRect(int(infoPanelY)+26, 20),
		b.tr("指標:", "Metric: ")+historyMetricLabel(b.lang, metric)+
			b.tr("(點圖表切換)", " (click the chart to switch)"), 11, body)
	if len(turns) < 2 {
		b.appendInfoCentered(s, infoCenteredTextRect(int(infoPanelY)+74, 32),
			b.tr("尚未累積足夠回合(結束回合後開始記錄)",
				"Not enough turns recorded yet (recording starts after you end a turn)"), 12, body)
		return
	}

	// 圖表區
	gx, gy := infoPanelX+40, infoPanelY+56
	gw, gh := infoPanelW-60, 250.0
	maxV := 1
	for _, ser := range series {
		for _, v := range ser {
			if v > maxV {
				maxV = v
			}
		}
	}
	names := historyEmpireLabels(b.lang, b.session)
	s.postDraw = func(dst *ebiten.Image) {
		ox, oy := float64(s.offsetX), float64(s.offsetY)
		// 座標軸
		axis := color.RGBA{120, 130, 150, 255}
		vector.StrokeLine(dst, float32(gx+ox), float32(gy+oy), float32(gx+ox), float32(gy+gh+oy), 1, axis, false)
		vector.StrokeLine(dst, float32(gx+ox), float32(gy+gh+oy), float32(gx+gw+ox), float32(gy+gh+oy), 1, axis, false)
		// 折線
		for i, ser := range series {
			col := empireLineColors[i%len(empireLineColors)]
			for j := 1; j < len(ser); j++ {
				x0 := gx + gw*float64(j-1)/float64(len(ser)-1)
				x1 := gx + gw*float64(j)/float64(len(ser)-1)
				y0 := gy + gh - gh*float64(ser[j-1])/float64(maxV)
				y1 := gy + gh - gh*float64(ser[j])/float64(maxV)
				vector.StrokeLine(dst, float32(x0+ox), float32(y0+oy), float32(x1+ox), float32(y1+oy), 2, col, false)
			}
		}
	}
	// 軸標 + 圖例。圖例保留三列、兩欄；超過時最後一格明示省略，不能壓到面板下緣。
	b.appendInfoLeft(s, infoHistoryMaxTextRect(), fmt.Sprintf("%d", maxV), 10, body)
	b.appendInfoLeft(s, infoHistoryZeroTextRect(), "0", 10, body)
	b.appendInfoLeft(s, infoHistoryTurnTextRect(false), fmt.Sprintf(b.tr("第 %d 回合", "Turn %d"), turns[0]), 10, body)
	b.appendInfoLeft(s, infoHistoryTurnTextRect(true), fmt.Sprintf(b.tr("第 %d 回合", "Turn %d"), turns[len(turns)-1]), 10, body)
	limit := len(names)
	if len(series) < limit {
		limit = len(series)
	}
	for i := 0; i < limit && i < infoHistoryLegendSlots; i++ {
		text := "── " + names[i]
		col := empireLineColors[i%len(empireLineColors)]
		if i == infoHistoryLegendSlots-1 && limit > infoHistoryLegendSlots {
			text = fmt.Sprintf(b.tr("…另有 %d 個帝國", "…%d more empires"), limit-i)
			col = infoDimCol
		}
		b.appendInfoLeft(s, infoHistoryLegendTextRect(i), text, 11, col)
	}
}

// --- 分頁 1:科技總覽(原版 Draw_Tech_Review_Subscreen_)---

func (b *sceneBuilder) infoTechReview(s *overlayScreen) {
	body := infoBodyCol
	done := infoOkCol
	cur := infoCurCol

	p := b.session.Player
	// 已完成主題(依名稱排序,穩定顯示)
	var completed []string
	for t := range p.CompletedTopics {
		completed = append(completed, topicNameZh(b.lang, t))
	}
	sort.Strings(completed)

	y := int(infoPanelY) + 44
	b.appendInfoLeft(s, infoContentTextRect(y, 14),
		fmt.Sprintf(b.tr("研究中:%s(%d RP)", "Researching: %s (%d RP)"),
			topicNameZh(b.lang, p.ResearchTopic), p.ResearchProgress), 12, cur)
	y += 22
	b.appendInfoLeft(s, infoContentTextRect(y, 14),
		fmt.Sprintf(b.tr("已完成主題:%d 項", "Fields completed: %d"), len(completed)), 12, body)
	y += 20
	// 兩欄列出已完成主題
	for i, nm := range completed {
		if y+(i/2)*15 > int(infoPanelY+infoPanelH)-40 {
			b.appendInfoLeft(s, infoContentTextRect(int(infoPanelY+infoPanelH)-30, 14),
				fmt.Sprintf(b.tr("…另有 %d 項", "…and %d more"), len(completed)-i), 10, body)
			break
		}
		b.appendInfoLeft(s, infoTechTopicTextRect(i%2, y+(i/2)*15, 12), "✓ "+nm, 10, done)
	}
	if len(completed) == 0 {
		b.appendInfoLeft(s, infoContentTextRect(y, 14), b.tr("(尚無已完成的研究主題)", "(no fields completed yet)"), 11, body)
	}
}

// --- 分頁 2:種族統計(原版 Draw_Race_Stats_Subscreen_)---

func (b *sceneBuilder) infoRaceStats(s *overlayScreen) {
	body := infoBodyCol
	gold := infoTitleCol
	sess := b.session

	y := int(infoPanelY) + 44
	for column, text := range []string{b.tr("帝國", "EMPIRE"), b.tr("殖民地", "COLONIES"), b.tr("人口", "POP"), b.tr("艦隊", "FLEET"), b.tr("態勢", "STANCE")} {
		b.appendInfoLeft(s, infoRaceStatTextRect(column, y, 12), text, 11, gold)
	}
	y += 20

	pop := 0
	for _, c := range sess.PlayerColonies {
		pop += c.Population
	}
	for column, text := range []string{b.tr("你", "You"), fmt.Sprintf("%d", len(sess.PlayerColonies)), fmt.Sprintf("%d", pop), fmt.Sprintf("%d", sess.PlayerFleetStrengthForUI()), "—"} {
		textCol := body
		if column == 0 {
			textCol = empireLineColors[0]
		}
		b.appendInfoLeft(s, infoRaceStatTextRect(column, y, 12), text, 11, textCol)
	}
	y += 18

	visibleAI := len(sess.AIPlayers)
	if visibleAI > racesMaxRows {
		visibleAI = racesMaxRows
	}
	for i := 0; i < visibleAI; i++ {
		a := &sess.AIPlayers[i]
		apop := 0
		for _, c := range a.Colonies {
			apop += c.Population
		}
		col := empireLineColors[(i+1)%len(empireLineColors)]
		for column, text := range []string{aiEmpireLabel(b.lang, *a), fmt.Sprintf("%d", len(a.Colonies)), fmt.Sprintf("%d", apop), fmt.Sprintf("%d", a.FleetStrength), a.StanceName} {
			textCol := body
			if column == 0 {
				textCol = col
			}
			b.appendInfoLeft(s, infoRaceStatTextRect(column, y, 12), text, 11, textCol)
		}
		y += 18
	}

	// AI 彼此關係(本專案 2026-07-12 建的 AIRelations 矩陣,原版對應 module 27 外交關係)
	y += 10
	b.appendInfoLeft(s, infoRaceRelationTextRect(y), b.tr("AI 之間的關係", "Relations between AI empires"), 11, gold)
	y += 18
	for i := 0; i < visibleAI; i++ {
		line := aiEmpireLabel(b.lang, sess.AIPlayers[i]) + ":"
		for j := 0; j < visibleAI; j++ {
			if i == j {
				continue
			}
			line += fmt.Sprintf(" [%s %s]", aiEmpireLabel(b.lang, sess.AIPlayers[j]), sess.AIRelationName(i, j))
		}
		b.appendInfoLeft(s, infoRaceRelationTextRect(y), line, 10, body)
		y += 15
	}
}

// --- 分頁 3:回合摘要(原版 Draw_Turn_Summary_Subscreen_)---

func (b *sceneBuilder) infoTurnSummary(s *overlayScreen) {
	body := infoBodyCol
	gold := infoTitleCol
	out := b.session.LastPlayerOutput

	y := int(infoPanelY) + 44
	rows := [][2]string{
		{b.tr("星曆", "Stardate"), shell.StardateForTurn(b.session.Turn)}, // 3500 起算,見 shell.StartStardate 的反組譯出處
		{b.tr("國庫", "Treasury"), fmt.Sprintf(b.tr("%d BC(本回合 %+d)", "%d BC (%+d this turn)"),
			b.session.Player.BC, out.NetBC)},
		{b.tr("稅收", "Taxes"), fmt.Sprintf("%d BC", out.TaxRevenue)},
		{b.tr("餘糧收入", "Food surplus"), fmt.Sprintf("%d BC", out.FoodSurplusRevenue)},
		{b.tr("貿易品收入", "Trade goods"), fmt.Sprintf("%d BC", out.TradeGoodsRevenue)},
		{b.tr("維護支出", "Maintenance"), fmt.Sprintf("%d BC", b.session.Player.Maintenance)},
		// 領袖薪餉單獨一列,不併進「維護支出」——那一列是**建築**維護,兩者來源不同,
		// 混在一起玩家就看不出「解雇一個領袖能省多少」。
		{b.tr("領袖薪餉", "Leader upkeep"), fmt.Sprintf("%d BC", b.session.LeaderUpkeepTotal())},
		{b.tr("指揮超支", "Command overrun"), fmt.Sprintf("%d BC", out.CommandOverflowCost)},
		{b.tr("食物盈餘", "Food surplus"), fmt.Sprintf("%d", out.TotalFood)},
		{b.tr("淨工業", "Net industry"), fmt.Sprintf("%d", out.TotalNetIndustry)},
		{b.tr("研究產出", "Research"), fmt.Sprintf("%d RP", out.TotalResearch)},
	}
	for _, r := range rows {
		b.appendInfoLeft(s, infoSummaryLabelTextRect(y), r[0], 11, gold)
		b.appendInfoLeft(s, infoSummaryValueTextRect(y), r[1], 11, body)
		y += 19
	}
	// 本回合事件/戰報(原版事件有獨立畫面,remake 暫列於此)
	y += 8
	eventMsg := b.session.LastEvent
	if b.lang != i18n.Traditional && b.session.LastEventReport != nil && b.session.LastEventReport.MessageEN != "" {
		eventMsg = b.session.LastEventReport.MessageEN
	}
	if b.lang != i18n.Traditional && b.session.LastPersistentEventEN != "" {
		if eventMsg != "" {
			eventMsg += " | "
		}
		eventMsg += b.session.LastPersistentEventEN
	}
	raidMsg := b.session.LastRaid
	if b.lang != i18n.Traditional && b.session.LastRaidReport != nil && b.session.LastRaidReport.MessageEN != "" {
		raidMsg = b.session.LastRaidReport.MessageEN
	}
	antaresMsg := b.session.LastAntares
	if b.lang != i18n.Traditional && b.session.LastAntaresEN != "" {
		antaresMsg = b.session.LastAntaresEN
	}
	for _, msg := range []string{eventMsg, antaresMsg, raidMsg, b.session.LastCouncil} {
		if msg == "" {
			continue
		}
		b.appendInfoLeft(s, infoTextRect(int(infoPanelX)+24, y, int(infoPanelW)-48, 12), "◆ "+msg, 10, infoEventCol)
		y += 16
	}
}

// --- 分頁 4:參考資料(原版 Draw_Reference_Main_/_Category_/_How_To_Subscreen_)---

func (b *sceneBuilder) infoReference(s *overlayScreen) {
	body := infoBodyCol
	gold := infoTitleCol

	// 原版 Reference 是「分類 + 怎麼做」兩欄清單(archive.org oracle 截圖已確認版面)。
	// remake 這裡放同結構的速查:左欄=規則分類、右欄=常用操作,內容取自專案的繁中機制大全
	// (docs/knowledge-base/manual-cht/),不重抄手冊原文。
	// 這兩份清單是原版 INFO → REFERENCE 分頁的目錄,英文取原版用語。
	cats := []string{"政府型態", "行星與星系", "太空建設", "艦艇系統", "艦艇引擎", "艦艇防禦",
		"艦艇武器", "飛彈與炸彈", "地面戰", "人口", "生產", "生態", "安全", "成就"}
	catsEn := []string{"Government", "Planets & Star Systems", "Space Construction", "Ship Systems",
		"Ship Drives", "Ship Defenses", "Ship Weapons", "Missiles & Bombs", "Ground Combat",
		"Population", "Production", "Ecology", "Security", "Achievements"}
	hows := []string{"建立新殖民地", "移動艦隊", "調動人口", "指派領袖", "前往安塔蘭",
		"入侵行星", "俘獲艦艇", "調整稅率", "設計艦艇", "取得更多說明"}
	howsEn := []string{"Create a new colony", "Move a fleet", "Move population", "Assign leaders",
		"Go to Antares", "Invade a planet", "Capture a ship", "Adjust taxes", "Design a ship",
		"Get more help"}
	if b.lang != i18n.Traditional {
		cats, hows = catsEn, howsEn
	}

	y := int(infoPanelY) + 44
	b.appendInfoLeft(s, infoReferenceTextRect(0, y), b.tr("分類", "CATEGORY"), 12, gold)
	b.appendInfoLeft(s, infoReferenceTextRect(1, y), b.tr("怎麼做?", "HOW DO I…?"), 12, gold)
	y += 20
	for i := 0; i < len(cats) || i < len(hows); i++ {
		if i < len(cats) {
			b.appendInfoLeft(s, infoReferenceTextRect(0, y+i*17), "• "+cats[i], 10, body)
		}
		if i < len(hows) {
			b.appendInfoLeft(s, infoReferenceTextRect(1, y+i*17), "• "+hows[i], 10, body)
		}
	}
	b.appendInfoLeft(s, infoTextRect(int(infoPanelX)+24, int(infoPanelY+infoPanelH)-24, int(infoPanelW)-48, 12),
		b.tr("詳細規則見專案文件 docs/knowledge-base/manual-cht/",
			"Full rules: docs/knowledge-base/manual-cht/ in the project repo"), 10, infoDimCol)
}

// 讓 gamedata 匯入不被最佳化掉(topicNameZh 需要 gamedata.ResearchTopic 型別)。
var _ gamedata.ResearchTopic
