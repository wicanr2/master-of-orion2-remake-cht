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
	evPanelBg   = color.RGBA{10, 14, 30, 245}
	evGoodEdge  = color.RGBA{90, 190, 120, 255} // 好消息:綠框
	evBadEdge   = color.RGBA{210, 110, 90, 255} // 壞消息:紅框
	evTitleCol  = color.RGBA{240, 220, 120, 255}
	evBodyCol   = color.RGBA{220, 228, 242, 255}
	evBrandCol  = color.RGBA{120, 180, 240, 255}
	evButtonBg  = color.RGBA{30, 40, 70, 255}
	evGNNHeader = "銀河新聞網 GNN ── 快報"
)

// eventScreen 建事件快報畫面。內容取自 session.LastEventReport;沒有事件就直接回回合摘要。
func (b *sceneBuilder) eventScreen() (*overlayScreen, error) {
	if b.session == nil || b.session.LastEventReport == nil {
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
	s, err := loadOverlayScreen(b.res, "turnsum.lbx", 0, b.lang, b.fnt, "assets/i18n/misc.tsv",
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

// drawEventReport 畫 GNN 快報面板。
func (b *sceneBuilder) drawEventReport(dst *ebiten.Image) {
	if b.fnt == nil || b.session == nil || b.session.LastEventReport == nil {
		return
	}
	rep := b.session.LastEventReport

	edge := evBadEdge
	if rep.Good {
		edge = evGoodEdge
	}
	// 先鋪一層暗遮罩:快報是疊在回合摘要背景上的彈窗,不遮的話底下的「TURN SUMMARY」
	// 標題與 CLOSE 鈕會從面板外露出來,跟快報的「繼續」鈕互相干擾。
	vector.DrawFilledRect(dst, 0, 0, 640, 480, color.RGBA{0, 0, 0, 205}, false)
	vector.DrawFilledRect(dst, evPanelX, evPanelY, evPanelW, evPanelH, evPanelBg, false)
	vector.StrokeRect(dst, evPanelX, evPanelY, evPanelW, evPanelH, 2, edge, false)

	// 台標列(原版 EVENTMSG 前 8 條就是這種 GNN 開場白)。
	vector.DrawFilledRect(dst, evPanelX, evPanelY, evPanelW, 26, color.RGBA{18, 30, 60, 255}, false)
	b.fnt.Draw(dst, evGNNHeader, evPanelX+12, evPanelY+6, 13, evBrandCol)

	// 事件名 + 好壞標記(對應原版 _event_good_array)。
	tag := "警訊"
	if rep.Good {
		tag = "喜訊"
	}
	b.fnt.Draw(dst, "【"+tag+"】"+rep.Name, evPanelX+16, evPanelY+42, 15, evTitleCol)

	// 快報內文(自動換行)。
	for i, ln := range b.fnt.Wrap(rep.Message, 13, evPanelW-32) {
		if i >= 8 {
			break
		}
		b.fnt.Draw(dst, ln, evPanelX+16, evPanelY+76+float64(i)*20, 13, evBodyCol)
	}

	// 確認鈕
	vector.DrawFilledRect(dst, 270, 372, 100, 24, evButtonBg, false)
	vector.StrokeRect(dst, 270, 372, 100, 24, 1, edge, false)
	b.fnt.DrawCentered(dst, "繼續", 320, 384, 13, evBodyCol)
}
