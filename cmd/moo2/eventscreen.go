package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
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
// 有 EVENTS.LBX 時使用原版 31 幀主播背景與 eventID+2 插圖；缺檔、解碼失敗、
// 非原版事件 ID 或勘查回報則安全退回回合摘要背景上的自繪面板。

const (
	evPanelX = 60.0
	evPanelY = 110.0
	evPanelW = 520.0
	evPanelH = 250.0

	evArtworkX = 320.0
	evArtworkY = 14.0
)

var (
	evPanelBg  = color.RGBA{10, 14, 30, 245}
	evGoodEdge = color.RGBA{90, 190, 120, 255} // 好消息:綠框
	evBadEdge  = color.RGBA{210, 110, 90, 255} // 壞消息:紅框
	evTitleCol = color.RGBA{240, 220, 120, 255}
	evBodyCol  = color.RGBA{220, 228, 242, 255}
	evBrandCol = color.RGBA{120, 180, 240, 255}
	evButtonBg = color.RGBA{30, 40, 70, 255}
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
		tag := uiText(b.lang, "event.tag.alert")
		if r.Good {
			tag = uiText(b.lang, "event.tag.good_news")
		}
		title, body := r.Name, r.Message
		if b.lang != i18n.Traditional {
			if r.NameEN != "" {
				title = r.NameEN
			}
			if r.MessageEN != "" {
				body = r.MessageEN
			}
		}
		return &reportPanel{header: uiText(b.lang, "event.header.gnn"),
			title: title, tag: tag, body: body, good: r.Good}
	}
	if d := b.session.LastDiscovery; d != nil {
		// 星系發現一律是好消息(原版這五種特殊物產沒有負面的)。
		title, body := d.Name, d.Message
		if b.lang != i18n.Traditional {
			if d.NameEN != "" {
				title = d.NameEN
			}
			if d.MessageEN != "" {
				body = d.MessageEN
			}
		}
		return &reportPanel{header: uiText(b.lang, "event.header.survey"),
			title: title, tag: uiText(b.lang, "event.tag.discovery"), body: body, good: true}
	}
	return nil
}

func eventHeaderTextRect() textSafeRect {
	return textSafeRect{x: int(evPanelX), y: int(evPanelY), w: int(evPanelW), h: 26, insetX: 12, insetY: 2}
}

func eventTitleTextRect(original bool) textSafeRect {
	if original {
		return textSafeRect{x: 76, y: 258, w: 488, h: 24, insetX: 2, insetY: 1}
	}
	return textSafeRect{x: 76, y: 148, w: 488, h: 24, insetX: 2, insetY: 1}
}

func eventBodyTextRect(original bool) textSafeRect {
	if original {
		return textSafeRect{x: 76, y: 288, w: 488, h: 70, insetX: 2, insetY: 1, lineH: 17}
	}
	return textSafeRect{x: 76, y: 184, w: 488, h: 156, insetX: 2, insetY: 1, lineH: 20}
}

func eventButtonTextRect() textSafeRect {
	return textSafeRect{x: 270, y: 372, w: 100, h: 24, insetX: 6, insetY: 1}
}

func eventArtworkAssetID(eventID, archiveCount int) (int, bool) {
	assetID := eventID + 2
	return assetID, eventID >= 0 && eventID < 36 && assetID >= 2 && assetID < archiveCount
}

func loadEventArtwork(res *assets.Resolver, eventID int) *ebiten.Image {
	archive, err := res.OpenLBX("events.lbx")
	if err != nil {
		return nil
	}
	assetID, ok := eventArtworkAssetID(eventID, archive.Count())
	if !ok {
		return nil
	}
	im, err := decodeAsset(res, "events.lbx", assetID)
	if err != nil || im.Embedded == nil || im.Width != 157 || im.Height != 125 || len(im.Frames) == 0 {
		return nil
	}
	return ebiten.NewImageFromImage(im.Frames[0].ToRGBA(im.Embedded, im.KeyColor()))
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
			return b.goTo(b.turnSummary, uiText(b.lang, "event.transition.summary"))
		}
		return nil
	}
	// turnsum.lbx 資產 0 沒有內嵌調色盤,要跟 buffer0.lbx 借(與 tacticalCombat 同一個做法)。
	// 少了這條 paletteChain 會載入失敗、goTo 回 nil,結果是「按下結束回合後畫面完全不動」——
	// EndTurn 其實跑了(星曆與國庫都變了),只是轉場沒發生,看起來像按鈕壞掉。
	original := b.session != nil && b.session.LastEventReport != nil
	var s *overlayScreen
	var err error
	if original {
		s, err = loadOverlayScreen(b.res, "events.lbx", 1, b.lang, b.fnt, "misc.json",
			nil, evBodyCol, 13, hits, onAction, paletteChain{{"events.lbx", 0}})
	}
	if s == nil || err != nil {
		original = false
		s, err = loadOverlayScreen(b.res, "turnsum.lbx", 0, b.lang, b.fnt, "misc.json",
			nil, evBodyCol, 13, hits, onAction, paletteChain{{"buffer0.lbx", 0}})
	}
	if err != nil {
		return nil, err
	}
	if original {
		if frames, frameErr := loadOverlayAnimationFrames(b.res, "events.lbx", 1,
			paletteChain{{"events.lbx", 0}}); frameErr == nil && len(frames) > 0 {
			s.animFrames = frames
			s.animTick = func() int { return b.animTick }
			s.animationStartTick = b.animTick
		}
	}
	var artwork *ebiten.Image
	if original && b.session != nil && b.session.LastEventReport != nil {
		artwork = loadEventArtwork(b.res, b.session.LastEventReport.EventID)
	}
	s.postDraw = func(dst *ebiten.Image) { b.drawEventReport(dst, original, artwork) }
	s.clickAnywhereAction = "ok"
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
func (b *sceneBuilder) drawEventReport(dst *ebiten.Image, original bool, artwork *ebiten.Image) {
	rep := b.currentReport()
	if b.fnt == nil || rep == nil {
		return
	}

	edge := evBadEdge
	if rep.good {
		edge = evGoodEdge
	}
	if original && artwork != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(evArtworkX, evArtworkY)
		drawPanelImage(dst, artwork, op)
	}

	// fallback 先鋪一層暗遮罩:快報是疊在回合摘要背景上的彈窗,不遮的話底下的「TURN SUMMARY」
	// 標題與 CLOSE 鈕會從面板外露出來,跟快報的「繼續」鈕互相干擾。
	// ⚠ 遮罩必須**完全不透明**。快報面板只蓋住畫面中段,而底下那張 TURNSUM 視窗的標題列
	// 烘著 `TURN SUMMARY`、底部烘著 `CLOSE` ——半透明遮罩(原本 205)會讓這兩個英文字樣
	// 從面板上下方透出來。這是**英文外洩**,不是美觀取捨。
	//
	// 為什麼不改成疊中文標題:overlayScreen 的擦底疊字跑在 postDraw **之前**,疊上去的中文
	// 一樣會被這層遮罩蓋掉。而這個畫面本來就不是像素對齊原版的(見檔頭),所以直接讓
	// 快報面板獨自浮在黑底上,反而是一致的做法。
	//
	// 底圖仍然載 turnsum.lbx:它提供這個場景的熱區/轉場骨架(見下方 offset 修正那段),
	// 只是視覺上被完全遮住。
	if original {
		fillPanel(dst, 60, 248, 520, 120, color.RGBA{10, 14, 30, 218}, false)
		vector.StrokeRect(dst, 60, 248, 520, 120, 2, edge, false)
	} else {
		fillPanel(dst, 0, 0, 640, 480, color.RGBA{0, 0, 0, 255}, false)
		fillPanel(dst, evPanelX, evPanelY, evPanelW, evPanelH, evPanelBg, false)
		vector.StrokeRect(dst, evPanelX, evPanelY, evPanelW, evPanelH, 2, edge, false)
		fillPanel(dst, evPanelX, evPanelY, evPanelW, 26, color.RGBA{18, 30, 60, 255}, false)
		eventHeaderTextRect().drawLeft(dst, b.fnt, rep.header, 13, evBrandCol)
	}

	title := fmt.Sprintf(uiText(b.lang, "event.title.format"), rep.tag, rep.title)
	eventTitleTextRect(original).drawLeft(dst, b.fnt, title, 15, evTitleCol)
	eventBodyTextRect(original).drawLeft(dst, b.fnt, rep.body, 13, evBodyCol)

	// 確認鈕
	fillPanel(dst, 270, 372, 100, 24, evButtonBg, false)
	vector.StrokeRect(dst, 270, 372, 100, 24, 1, edge, false)
	eventButtonTextRect().drawCentered(dst, b.fnt, uiText(b.lang, "event.button.continue"), 13, evBodyCol)
}
