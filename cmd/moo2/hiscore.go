package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
)

// hiscore.go:最終得分畫面(原版 Hi-Score / Hall of Fame,module 60 的
// `Draw_Hi_Score_Screen_` @ 0x9E27A 與一整組 `Draw_*_Score_` 分項)。
//
// 分項與係數全部來自反組譯(見 internal/gamedata/score.go 檔頭的八個 `Get_*_Score_` 位址),
// 這裡只負責把 shell 算好的逐項分數畫出來。原版是逐項固定顯示、不是有分才畫,照做。
//
// 原版主背景是 SCORE.LBX#0；缺資產時才退回自繪面板。

const (
	hsPanelX = 90.0
	hsPanelY = 60.0
	hsPanelW = 460.0
	hsPanelH = 360.0

	hsTitleY      = 122
	hsSummaryY    = 154
	hsSummaryH    = 24
	hsScoreTop    = 184
	hsScoreBottom = 390
	hsScoreLabelX = 140
	hsScoreLabelW = 280
	hsScoreValueX = 440
	hsScoreValueW = 60
	hsContinueX   = 220
	hsContinueY   = 394
	hsContinueW   = 200
	hsContinueH   = 18
)

func hiScoreTitleTextRect() textSafeRect {
	return textSafeRect{x: 140, y: hsTitleY, w: 360, h: 32, insetX: 4, insetY: 0, lineH: 32}
}

func hiScoreSummaryTextRect() textSafeRect {
	return textSafeRect{x: 140, y: hsSummaryY, w: 360, h: hsSummaryH, insetX: 4, insetY: 1, lineH: 22}
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
	y := float64(hsScoreTop + index*22)
	if total {
		y += 6
	}
	rowH := 18
	if total {
		rowH = 24
	}
	return y, y+float64(rowH) <= float64(hsScoreBottom)
}

func hiScoreBackgroundAvailable(res *assets.Resolver) bool {
	im, err := decodeAsset(res, "score.lbx", 0)
	return err == nil && im.Width == 640 && im.Height == 480 && len(im.Frames) > 0 && im.Embedded != nil
}

// hiScore 建最終得分畫面。對局尚未結束時直接回星系主畫面(不該從那裡進來)。
func (b *sceneBuilder) hiScore() (*overlayScreen, error) {
	if b.session == nil || !b.session.Victory.Over {
		return b.galaxy()
	}
	onAction := func(a string) *origTransition {
		if a == "ok" {
			return b.goTo(b.turnSummary, uiText(b.lang, "hiscore.transition.summary"))
		}
		return nil
	}
	original := hiScoreBackgroundAvailable(b.res)
	lbxName, chain := "score.lbx", paletteChain(nil)
	if !original {
		lbxName, chain = "turnsum.lbx", paletteChain{{"buffer0.lbx", 0}}
	}
	s, err := loadOverlayScreen(b.res, lbxName, 0, b.lang, b.fnt, "misc.json",
		nil, color.RGBA{220, 228, 242, 255}, 13, nil, onAction, chain)
	if err != nil {
		return nil, err
	}
	s.postDraw = func(dst *ebiten.Image) { b.drawHiScore(dst, original) }
	s.clickAnywhereAction = "ok"
	return s, nil
}

// drawHiScore 畫最終得分面板。
func (b *sceneBuilder) drawHiScore(dst *ebiten.Image, original bool) {
	if b.fnt == nil || b.session == nil {
		return
	}
	v := b.session.Victory
	won := v.Winner == "player"

	edge := color.RGBA{210, 110, 90, 255}
	if won {
		edge = color.RGBA{240, 220, 120, 255}
	}
	if original {
		fillPanel(dst, 136, 122, 368, 292, color.RGBA{0, 0, 0, 235}, false)
		if b.lang != i18n.English {
			fillPanel(dst, 165, 43, 310, 28, color.RGBA{8, 8, 12, 255}, false)
			textSafeRect{x: 165, y: 43, w: 310, h: 28, insetX: 5, insetY: 1, lineH: 26}.drawCentered(dst, b.fnt, uiText(b.lang, "hiscore.header.hall_of_fame"), 18, edge)
		}
	} else {
		fillPanel(dst, 0, 0, 640, 480, color.RGBA{0, 0, 0, 210}, false)
		fillPanel(dst, hsPanelX, hsPanelY, hsPanelW, hsPanelH, color.RGBA{10, 14, 30, 245}, false)
		vector.StrokeRect(dst, hsPanelX, hsPanelY, hsPanelW, hsPanelH, 2, edge, false)
	}

	title := uiText(b.lang, "hiscore.title.lost")
	if won {
		title = uiText(b.lang, "hiscore.title.won")
	}
	hiScoreTitleTextRect().drawCentered(dst, b.fnt, title, 22, edge)

	// 勝負與方式。
	reason := map[engine.VictoryCondition]string{
		engine.VictoryExtermination: uiText(b.lang, "hiscore.reason.extermination"),
		engine.VictoryHighCouncil:   uiText(b.lang, "hiscore.reason.council"),
		engine.VictoryAntaran:       uiText(b.lang, "hiscore.reason.antaran"),
	}[v.Reason]
	if reason == "" {
		reason = uiText(b.lang, "hiscore.reason.game_over")
	}
	line := fmt.Sprintf(uiText(b.lang, "hiscore.summary.won"), v.Turn, reason)
	if !won {
		line = fmt.Sprintf(uiText(b.lang, "hiscore.summary.lost"), v.Turn, v.Winner)
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
			vector.StrokeRect(dst, 140, float32(y)-6, 360, 1, 1, color.RGBA{90, 110, 150, 255}, false)
			col = edge
		}
		labelRect := hiScoreLabelTextRect(int(y), int(size)+4)
		labelRect.drawLeft(dst, b.fnt, uiText(b.lang, ln.TextKey), size, col)
		val := fmt.Sprintf("%d", ln.Value)
		hiScoreValueTextRect(int(y), int(size)+4).drawRight(dst, b.fnt, val, size, col)
		rowIndex++
	}

	textSafeRect{x: hsContinueX, y: hsContinueY, w: hsContinueW, h: hsContinueH, insetX: 5, insetY: 1, lineH: 16}.drawCentered(dst, b.fnt, uiText(b.lang, "hiscore.button.continue"), 12, color.RGBA{220, 228, 242, 255})
}
