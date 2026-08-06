package main

import (
	"fmt"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
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

// drawInfoSubscreen 依 b.infoTab 把對應子畫面內容加進 overlayScreen。
func (b *sceneBuilder) drawInfoSubscreen(s *overlayScreen) {
	tab := b.infoTab
	if tab < 0 || tab >= len(infoTabTitles) {
		tab = 0
	}
	gold := infoTitleCol
	s.extras = append(s.extras, extraText{
		x: infoPanelX + infoPanelW/2, y: infoPanelY + 16, size: 15,
		text: infoTabTitles[tab], col: gold, align: 1,
	})
	if b.session == nil {
		s.extras = append(s.extras, extraText{
			x: infoPanelX + infoPanelW/2, y: infoPanelY + 60, size: 12,
			text: "尚無對局資料", col: infoBodyCol, align: 1,
		})
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

func (b *sceneBuilder) infoHistory(s *overlayScreen) {
	body := infoBodyCol
	metric := shell.HistoryMetric(b.infoHistoryMetric)
	series, turns := b.session.HistorySeries(metric)

	s.extras = append(s.extras, extraText{
		x: infoPanelX + infoPanelW/2, y: infoPanelY + 36, size: 11,
		text: "指標:" + shell.HistoryMetricName(metric) + "(點圖表切換)", col: body, align: 1,
	})
	if len(turns) < 2 {
		s.extras = append(s.extras, extraText{
			x: infoPanelX + infoPanelW/2, y: infoPanelY + 90, size: 12,
			text: "尚未累積足夠回合(結束回合後開始記錄)", col: body, align: 1,
		})
		return
	}

	// 圖表區
	gx, gy := infoPanelX + 40, infoPanelY + 56
	gw, gh := infoPanelW - 60, 250.0
	maxV := 1
	for _, ser := range series {
		for _, v := range ser {
			if v > maxV {
				maxV = v
			}
		}
	}
	names := b.session.HistoryEmpireNames()
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
	// 軸標 + 圖例
	s.extras = append(s.extras,
		extraText{x: gx - 34, y: gy + 6, size: 10, text: fmt.Sprintf("%d", maxV), col: body},
		extraText{x: gx - 20, y: gy + gh + 4, size: 10, text: "0", col: body},
		extraText{x: gx, y: gy + gh + 18, size: 10, text: fmt.Sprintf("第 %d 回合", turns[0]), col: body},
		extraText{x: gx + gw - 50, y: gy + gh + 18, size: 10, text: fmt.Sprintf("第 %d 回合", turns[len(turns)-1]), col: body},
	)
	ly := gy + gh + 34
	for i, nm := range names {
		if i >= len(series) {
			break
		}
		s.extras = append(s.extras, extraText{
			x: gx + float64(i%2)*190, y: ly + float64(i/2)*16, size: 11,
			text: "── " + nm, col: empireLineColors[i%len(empireLineColors)],
		})
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

	y := infoPanelY + 44
	s.extras = append(s.extras, extraText{
		x: infoPanelX + 16, y: y, size: 12,
		text: fmt.Sprintf("研究中:%s(%d RP)", topicNameZh(b.lang, p.ResearchTopic), p.ResearchProgress), col: cur,
	})
	y += 22
	s.extras = append(s.extras, extraText{
		x: infoPanelX + 16, y: y, size: 12,
		text: fmt.Sprintf("已完成主題:%d 項", len(completed)), col: body,
	})
	y += 20
	// 兩欄列出已完成主題
	for i, nm := range completed {
		if y+float64(i/2)*15 > infoPanelY+infoPanelH-40 {
			s.extras = append(s.extras, extraText{
				x: infoPanelX + 16, y: infoPanelY + infoPanelH - 30, size: 10,
				text: fmt.Sprintf("…另有 %d 項", len(completed)-i), col: body,
			})
			break
		}
		s.extras = append(s.extras, extraText{
			x: infoPanelX + 16 + float64(i%2)*200, y: y + float64(i/2)*15, size: 10,
			text: "✓ " + nm, col: done,
		})
	}
	if len(completed) == 0 {
		s.extras = append(s.extras, extraText{
			x: infoPanelX + 16, y: y, size: 11, text: "(尚無已完成的研究主題)", col: body,
		})
	}
}

// --- 分頁 2:種族統計(原版 Draw_Race_Stats_Subscreen_)---

func (b *sceneBuilder) infoRaceStats(s *overlayScreen) {
	body := infoBodyCol
	gold := infoTitleCol
	sess := b.session

	y := infoPanelY + 44
	s.extras = append(s.extras,
		extraText{x: infoPanelX + 16, y: y, size: 11, text: "帝國", col: gold},
		extraText{x: infoPanelX + 150, y: y, size: 11, text: "殖民地", col: gold},
		extraText{x: infoPanelX + 215, y: y, size: 11, text: "人口", col: gold},
		extraText{x: infoPanelX + 268, y: y, size: 11, text: "艦隊", col: gold},
		extraText{x: infoPanelX + 325, y: y, size: 11, text: "態勢", col: gold},
	)
	y += 20

	pop := 0
	for _, c := range sess.PlayerColonies {
		pop += c.Population
	}
	s.extras = append(s.extras,
		extraText{x: infoPanelX + 16, y: y, size: 11, text: "你", col: empireLineColors[0]},
		extraText{x: infoPanelX + 150, y: y, size: 11, text: fmt.Sprintf("%d", len(sess.PlayerColonies)), col: body},
		extraText{x: infoPanelX + 215, y: y, size: 11, text: fmt.Sprintf("%d", pop), col: body},
		extraText{x: infoPanelX + 268, y: y, size: 11, text: fmt.Sprintf("%d", sess.PlayerFleetStrengthForUI()), col: body},
		extraText{x: infoPanelX + 325, y: y, size: 11, text: "—", col: body},
	)
	y += 18

	for i := range sess.AIPlayers {
		a := &sess.AIPlayers[i]
		apop := 0
		for _, c := range a.Colonies {
			apop += c.Population
		}
		col := empireLineColors[(i+1)%len(empireLineColors)]
		s.extras = append(s.extras,
			extraText{x: infoPanelX + 16, y: y, size: 11, text: a.Name, col: col},
			extraText{x: infoPanelX + 150, y: y, size: 11, text: fmt.Sprintf("%d", len(a.Colonies)), col: body},
			extraText{x: infoPanelX + 215, y: y, size: 11, text: fmt.Sprintf("%d", apop), col: body},
			extraText{x: infoPanelX + 268, y: y, size: 11, text: fmt.Sprintf("%d", a.FleetStrength), col: body},
			extraText{x: infoPanelX + 325, y: y, size: 11, text: a.StanceName, col: body},
		)
		y += 18
	}

	// AI 彼此關係(本專案 2026-07-12 建的 AIRelations 矩陣,原版對應 module 27 外交關係)
	y += 10
	s.extras = append(s.extras, extraText{x: infoPanelX + 16, y: y, size: 11, text: "AI 之間的關係", col: gold})
	y += 18
	for i := range sess.AIPlayers {
		line := sess.AIPlayers[i].Name + ":"
		for j := range sess.AIPlayers {
			if i == j {
				continue
			}
			line += fmt.Sprintf(" [%s %s]", sess.AIPlayers[j].Name, sess.AIRelationName(i, j))
		}
		s.extras = append(s.extras, extraText{x: infoPanelX + 16, y: y, size: 10, text: line, col: body})
		y += 15
	}
}

// --- 分頁 3:回合摘要(原版 Draw_Turn_Summary_Subscreen_)---

func (b *sceneBuilder) infoTurnSummary(s *overlayScreen) {
	body := infoBodyCol
	gold := infoTitleCol
	out := b.session.LastPlayerOutput

	y := infoPanelY + 44
	rows := [][2]string{
		{"星曆", fmt.Sprintf("%d", 3500+(b.session.Turn-1))},
		{"國庫", fmt.Sprintf("%d BC(本回合 %+d)", b.session.Player.BC, out.NetBC)},
		{"稅收", fmt.Sprintf("%d BC", out.TaxRevenue)},
		{"餘糧收入", fmt.Sprintf("%d BC", out.FoodSurplusRevenue)},
		{"貿易品收入", fmt.Sprintf("%d BC", out.TradeGoodsRevenue)},
		{"維護支出", fmt.Sprintf("%d BC", b.session.Player.Maintenance)},
		{"指揮超支", fmt.Sprintf("%d BC", out.CommandOverflowCost)},
		{"食物盈餘", fmt.Sprintf("%d", out.TotalFood)},
		{"淨工業", fmt.Sprintf("%d", out.TotalNetIndustry)},
		{"研究產出", fmt.Sprintf("%d RP", out.TotalResearch)},
	}
	for _, r := range rows {
		s.extras = append(s.extras,
			extraText{x: infoPanelX + 24, y: y, size: 11, text: r[0], col: gold},
			extraText{x: infoPanelX + 150, y: y, size: 11, text: r[1], col: body},
		)
		y += 19
	}
	// 本回合事件/戰報(原版事件有獨立畫面,remake 暫列於此)
	y += 8
	for _, msg := range []string{b.session.LastEvent, b.session.LastAntares, b.session.LastRaid, b.session.LastCouncil} {
		if msg == "" {
			continue
		}
		s.extras = append(s.extras, extraText{x: infoPanelX + 24, y: y, size: 10, text: "◆ " + msg, col: infoEventCol})
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
	cats := []string{"政府型態", "行星與星系", "太空建設", "艦艇系統", "艦艇引擎", "艦艇防禦",
		"艦艇武器", "飛彈與炸彈", "地面戰", "人口", "生產", "生態", "安全", "成就"}
	hows := []string{"建立新殖民地", "移動艦隊", "調動人口", "指派領袖", "前往安塔蘭",
		"入侵行星", "俘獲艦艇", "調整稅率", "設計艦艇", "取得更多說明"}

	y := infoPanelY + 44
	s.extras = append(s.extras,
		extraText{x: infoPanelX + 24, y: y, size: 12, text: "分類", col: gold},
		extraText{x: infoPanelX + 220, y: y, size: 12, text: "怎麼做?", col: gold},
	)
	y += 20
	for i := 0; i < len(cats) || i < len(hows); i++ {
		if i < len(cats) {
			s.extras = append(s.extras, extraText{x: infoPanelX + 24, y: y + float64(i)*17, size: 10, text: "• " + cats[i], col: body})
		}
		if i < len(hows) {
			s.extras = append(s.extras, extraText{x: infoPanelX + 220, y: y + float64(i)*17, size: 10, text: "• " + hows[i], col: body})
		}
	}
	s.extras = append(s.extras, extraText{
		x: infoPanelX + 24, y: infoPanelY + infoPanelH - 24, size: 10,
		text: "詳細規則見專案文件 docs/knowledge-base/manual-cht/", col: infoDimCol,
	})
}

// 讓 gamedata 匯入不被最佳化掉(topicNameZh 需要 gamedata.ResearchTopic 型別)。
var _ gamedata.ResearchTopic
