package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// hiscore.go:最終得分畫面(原版 Hi-Score / Hall of Fame,module 60 的
// `Draw_Hi_Score_Screen_` @ 0x9E27A 與一整組 `Draw_*_Score_` 分項)。
//
// 分項與係數全部來自反組譯(見 internal/gamedata/score.go 檔頭的八個 `Get_*_Score_` 位址),
// 這裡只負責把 shell 算好的逐項分數畫出來。原版是逐項固定顯示、不是有分才畫,照做。
//
// 版面:remake 沒有原版 Hi-Score 底圖的版面資料,這裡自繪。對齊的是**內容與分項**,不是像素。

const (
	hsPanelX = 90.0
	hsPanelY = 60.0
	hsPanelW = 460.0
	hsPanelH = 360.0

	hsTitleY      = 68
	hsSummaryY    = 96
	hsSummaryH    = 32
	hsScoreTop    = 148
	hsScoreBottom = 376 // 繼續按鈕 y=388，上方保留 12 px 緩衝。
	hsScoreLabelX = 124
	hsScoreLabelW = 280
	hsScoreValueX = 458
	hsScoreValueW = 58
	hsContinueX   = 270
	hsContinueY   = 388
	hsContinueW   = 100
	hsContinueH   = 24
)

func hiScoreTitleTextRect() textSafeRect {
	return textSafeRect{x: int(hsPanelX), y: hsTitleY, w: int(hsPanelW), h: 32, insetX: 10, insetY: 1, lineH: 30}
}

func hiScoreSummaryTextRect() textSafeRect {
	return textSafeRect{x: int(hsPanelX) + 20, y: hsSummaryY, w: int(hsPanelW) - 40, h: hsSummaryH, insetX: 4, insetY: 1, lineH: 30}
}

func hiScoreLabelTextRect(y, h int) textSafeRect {
	return textSafeRect{x: hsScoreLabelX, y: y, w: hsScoreLabelW, h: h, insetX: 2, insetY: 1, lineH: h - 2}
}

func hiScoreValueTextRect(y, h int) textSafeRect {
	return textSafeRect{x: hsScoreValueX, y: y, w: hsScoreValueW, h: h, insetX: 2, insetY: 1, lineH: h - 2}
}

// hiScoreRowY 以繼續按鈕前的明確底界限制列數。呼叫端可用它判斷
// ScoreLines 是否超量，不能讓新增分項默默畫進按鈕或面板外。
func hiScoreRowY(index int, total bool) (float64, bool) {
	y := float64(hsScoreTop + index*24)
	if total {
		y += 10
	}
	rowH := 18
	if total {
		rowH = 24
	}
	return y, y+float64(rowH) <= float64(hsScoreBottom)
}

func hiScoreRightAlignedValue(fnt *uifont.Font, dst *ebiten.Image, r textSafeRect, value string, size float64, col color.Color) {
	if fnt == nil {
		return
	}
	value = r.clipped(fnt, value, size)
	w, _ := fnt.Measure(value, size)
	fnt.Draw(dst, value, float64(r.x+r.w-r.insetX)-w, float64(r.contentY()), size, col)
}

// hiScore 建最終得分畫面。對局尚未結束時直接回星系主畫面(不該從那裡進來)。
func (b *sceneBuilder) hiScore() (*overlayScreen, error) {
	if b.session == nil || !b.session.Victory.Over {
		return b.galaxy()
	}
	hits := []hitRegion{{270, 388, 100, 24, "ok"}}
	onAction := func(a string) *origTransition {
		if a == "ok" {
			return b.goTo(b.turnSummary, "回合摘要")
		}
		return nil
	}
	// 沿用回合摘要底圖(turnsum.lbx 資產 0 沒有內嵌調色盤,要跟 buffer0.lbx 借,
	// 與 eventScreen 同一個做法——少了 paletteChain 會載入失敗、轉場靜默失效)。
	s, err := loadOverlayScreen(b.res, "turnsum.lbx", 0, b.lang, b.fnt, "misc.tsv",
		nil, color.RGBA{220, 228, 242, 255}, 13, hits, onAction, paletteChain{{"buffer0.lbx", 0}})
	if err != nil {
		return nil, err
	}
	s.postDraw = func(dst *ebiten.Image) { b.drawHiScore(dst) }
	// 熱區用背景圖局部座標,面板畫在全螢幕座標——與 eventScreen 同一個偏移修正。
	for i := range s.hits {
		s.hits[i].x -= s.offsetX
		s.hits[i].y -= s.offsetY
	}
	return s, nil
}

// drawHiScore 畫最終得分面板。
func (b *sceneBuilder) drawHiScore(dst *ebiten.Image) {
	if b.fnt == nil || b.session == nil {
		return
	}
	v := b.session.Victory
	won := v.Winner == "player"

	edge := color.RGBA{210, 110, 90, 255}
	if won {
		edge = color.RGBA{240, 220, 120, 255}
	}
	fillPanel(dst, 0, 0, 640, 480, color.RGBA{0, 0, 0, 210}, false)
	fillPanel(dst, hsPanelX, hsPanelY, hsPanelW, hsPanelH, color.RGBA{10, 14, 30, 245}, false)
	vector.StrokeRect(dst, hsPanelX, hsPanelY, hsPanelW, hsPanelH, 2, edge, false)

	title := b.tr("帝國殞落", "YOUR EMPIRE HAS FALLEN")
	if won {
		title = b.tr("銀河霸主", "MASTER OF ORION")
	}
	hiScoreTitleTextRect().drawCentered(dst, b.fnt, title, 22, edge)

	// 勝負與方式。
	reason := map[engine.VictoryCondition]string{
		engine.VictoryExtermination: b.tr("殲滅所有對手", "all rivals exterminated"),
		engine.VictoryHighCouncil:   b.tr("銀河議會選舉", "elected by the Galactic Council"),
		engine.VictoryAntaran:       b.tr("攻陷安塔蘭母星", "the Antaran homeworld taken"),
	}[v.Reason]
	if reason == "" {
		reason = b.tr("對局結束", "game over")
	}
	line := fmt.Sprintf(b.tr("第 %d 回合・%s", "Turn %d — %s"), v.Turn, reason)
	if !won {
		line = fmt.Sprintf(b.tr("第 %d 回合・%s 取得勝利", "Turn %d — %s wins"), v.Turn, v.Winner)
	}
	hiScoreSummaryTextRect().drawCentered(dst, b.fnt, line, 13, color.RGBA{200, 214, 232, 255})

	// 逐項得分(原版 Draw_*_Score_ 的分項順序)。
	lines := b.session.ScoreLines()
	rowIndex := 0
	for i, ln := range lines {
		col := color.RGBA{206, 218, 240, 255}
		size := 14.0
		isTotal := i == len(lines)-1
		y, ok := hiScoreRowY(rowIndex, isTotal)
		if !ok {
			// ScoreLines 若日後擴充，明確停止在面板底界；不能撞繼續按鈕。
			break
		}
		if isTotal { // 總分
			vector.StrokeRect(dst, hsPanelX+28, float32(y)-6, hsPanelW-56, 1, 1, color.RGBA{90, 110, 150, 255}, false)
			col, size = edge, 18
		}
		labelRect := hiScoreLabelTextRect(int(y), int(size)+4)
		b.fnt.Draw(dst, labelRect.clipped(b.fnt, ln.Label, size), float64(labelRect.contentX()), float64(labelRect.contentY()), size, col)
		// 分數靠右對齊:uifont 沒有 DrawRight,但仍受獨立數值安全欄限制。
		val := fmt.Sprintf("%d", ln.Value)
		hiScoreRightAlignedValue(b.fnt, dst, hiScoreValueTextRect(int(y), int(size)+4), val, size, col)
		rowIndex++
	}

	fillPanel(dst, hsContinueX, hsContinueY, hsContinueW, hsContinueH, color.RGBA{30, 40, 70, 255}, false)
	vector.StrokeRect(dst, hsContinueX, hsContinueY, hsContinueW, hsContinueH, 1, edge, false)
	textSafeRect{x: hsContinueX, y: hsContinueY, w: hsContinueW, h: hsContinueH, insetX: 5, insetY: 1, lineH: 20}.drawCentered(dst, b.fnt, b.tr("繼續", "CONTINUE"), 13, color.RGBA{220, 228, 242, 255})
}
