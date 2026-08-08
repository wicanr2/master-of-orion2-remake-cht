package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// 事件畫面(原版 `Event_Screen_` / `Draw_Event_Screen_` / `Event_Fade_In_`,module 15)。
//
// 原版把隨機事件做成 **GNN(Galactic News Network)新聞快報**:一個獨立畫面,
// 主播圖 + 快報文字,而不是把事件塞進回合摘要的一行字。
// 證據:EVENTMSG.LBX 的前 8 條就是 GNN 開場白("You are tuned to GNN…"、
// "And now, in a GNN exclusive story,"),其後每個事件 4 條快報文字;
// 畫面函式在 module 15 與事件邏輯同一個模組(見 gamedata/events.go 檔頭)。
//
// remake 先前只有「回合摘要裡一行字」,事件發生跟沒發生一樣沒有存在感。
//
// 版面:remake 沒有原版 EVENTS.LBX 的主播圖版面資料,這裡疊在回合摘要背景上自繪
// 快報面板。對齊的是**流程**(事件觸發 → 專屬播報畫面 → 回到回合摘要),不是像素。

const (
	evPanelX = 60.0
	evPanelY = 110.0
	evPanelW = 520.0
	evPanelH = 250.0
)

var (
	evPanelBg     = color.RGBA{10, 14, 30, 245}
	evGoodEdge    = color.RGBA{90, 190, 120, 255} // 好消息:綠框
	evBadEdge     = color.RGBA{210, 110, 90, 255} // 壞消息:紅框
	evTitleCol    = color.RGBA{240, 220, 120, 255}
	evBodyCol     = color.RGBA{220, 228, 242, 255}
	evBrandCol    = color.RGBA{120, 180, 240, 255}
	evButtonBg    = color.RGBA{30, 40, 70, 255}
	evGNNHeader   = "銀河新聞網 GNN ── 快報"
	evGNNHeaderEn = "GALACTIC NEWS NETWORK ── BULLETIN"
	// 星系發現不是 GNN 新聞,是自家勘查隊的回報,台標列另用一組字。
	evScoutHeader   = "帝國勘查回報"
	evScoutHeaderEn = "IMPERIAL SURVEY REPORT"
)

// reportPanel 是快報面板要畫的內容(隨機事件與星系發現共用同一個版面)。
type reportPanel struct {
	header string // 台標列文字
	title  string // 標題(事件名/特殊物產名)
	tag    string // 標題前的方括號標記
	body   string // 內文
	good   bool   // 好消息(綠框)/壞消息(紅框)
}

// currentReport 依 session 目前的狀態決定要播哪一則快報;兩者皆無回 nil。
// 隨機事件優先——它是本回合結算出來的全銀河新聞,發現則是自家艦隊的回報,兩則同時發生時
// 先播新聞、發現的內容仍留在回合摘要文字裡。
func (b *sceneBuilder) currentReport() *reportPanel {
	if b.session == nil {
		return nil
	}
	if r := b.session.LastEventReport; r != nil {
		tag := b.tr("警訊", "ALERT")
		if r.Good {
			tag = b.tr("喜訊", "GOOD NEWS")
		}
		return &reportPanel{header: b.tr(evGNNHeader, evGNNHeaderEn),
			title: r.Name, tag: tag, body: r.Message, good: r.Good}
	}
	if d := b.session.LastDiscovery; d != nil {
		// 星系發現一律是好消息(原版這五種特殊物產沒有負面的)。
		return &reportPanel{header: b.tr(evScoutHeader, evScoutHeaderEn),
			title: d.Name, tag: b.tr("發現", "DISCOVERY"), body: d.Message, good: true}
	}
	return nil
}

// eventScreen 建快報畫面。內容取自 currentReport();沒有可播的就直接回回合摘要。
func (b *sceneBuilder) eventScreen() (*overlayScreen, error) {
	playSceneBGM(trackEventScreen) // Start_Main_Event_ / Draw_Event_Screen_ → STREAMHD #18
	if b.currentReport() == nil {
		return b.turnSummary()
	}
	hits := []hitRegion{{270, 372, 100, 24, "ok"}}
	onAction := func(a string) *origTransition {
		if a == "ok" {
			return b.goTo(b.turnSummary, "回合摘要")
		}
		return nil
	}
	// turnsum.lbx 資產 0 沒有內嵌調色盤,要跟 buffer0.lbx 借(與 tacticalCombat 同一個做法)。
	// 少了這條 paletteChain 會載入失敗、goTo 回 nil,結果是「按下結束回合後畫面完全不動」——
	// EndTurn 其實跑了(星曆與國庫都變了),只是轉場沒發生,看起來像按鈕壞掉。
	s, err := loadOverlayScreen(b.res, "turnsum.lbx", 0, b.lang, b.fnt, "misc.tsv",
		nil, evBodyCol, 13, hits, onAction, paletteChain{{"buffer0.lbx", 0}})
	if err != nil {
		return nil, err
	}
	s.postDraw = func(dst *ebiten.Image) { b.drawEventReport(dst) }
	// 快報面板畫在 640×480 螢幕座標(postDraw 拿到的是整張畫布),但熱區命中判定用的是
	// **背景圖局部座標**(overlayScreen.update 會先扣掉置中偏移)。turnsum.lbx 的背景小於
	// 全螢幕、會被置中,兩套座標因此差一個偏移量——不補這一段,「繼續」鈕看得到卻永遠點不到。
	// (colsum/buffer0 那些滿版 640×480 的畫面 offset 為 0,才沒踩到這個坑。)
	for i := range s.hits {
		s.hits[i].x -= s.offsetX
		s.hits[i].y -= s.offsetY
	}
	return s, nil
}

// drawEventReport 畫快報面板(隨機事件 / 星系發現共用)。
func (b *sceneBuilder) drawEventReport(dst *ebiten.Image) {
	rep := b.currentReport()
	if b.fnt == nil || rep == nil {
		return
	}

	edge := evBadEdge
	if rep.good {
		edge = evGoodEdge
	}
	// 先鋪一層暗遮罩:快報是疊在回合摘要背景上的彈窗,不遮的話底下的「TURN SUMMARY」
	// 標題與 CLOSE 鈕會從面板外露出來,跟快報的「繼續」鈕互相干擾。
	vector.DrawFilledRect(dst, 0, 0, 640, 480, color.RGBA{0, 0, 0, 205}, false)
	vector.DrawFilledRect(dst, evPanelX, evPanelY, evPanelW, evPanelH, evPanelBg, false)
	vector.StrokeRect(dst, evPanelX, evPanelY, evPanelW, evPanelH, 2, edge, false)

	// 台標列(原版 EVENTMSG 前 8 條就是這種 GNN 開場白)。
	vector.DrawFilledRect(dst, evPanelX, evPanelY, evPanelW, 26, color.RGBA{18, 30, 60, 255}, false)
	b.fnt.Draw(dst, rep.header, evPanelX+12, evPanelY+6, 13, evBrandCol)

	// 事件名 + 好壞標記(對應原版 _event_good_array)。
	b.fnt.Draw(dst, "【"+rep.tag+"】"+rep.title, evPanelX+16, evPanelY+42, 15, evTitleCol)

	// 快報內文(自動換行)。
	for i, ln := range b.fnt.Wrap(rep.body, 13, evPanelW-32) {
		if i >= 8 {
			break
		}
		b.fnt.Draw(dst, ln, evPanelX+16, evPanelY+76+float64(i)*20, 13, evBodyCol)
	}

	// 確認鈕
	vector.DrawFilledRect(dst, 270, 372, 100, 24, evButtonBg, false)
	vector.StrokeRect(dst, 270, 372, 100, 24, 1, edge, false)
	b.fnt.DrawCentered(dst, b.tr("繼續", "CONTINUE"), 320, 384, 13, evBodyCol)
}
