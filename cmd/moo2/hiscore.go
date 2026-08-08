package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
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
)

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
	b.fnt.DrawCentered(dst, title, 320, 84, 22, edge)

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
	b.fnt.DrawCentered(dst, line, 320, 112, 13, color.RGBA{200, 214, 232, 255})

	// 逐項得分(原版 Draw_*_Score_ 的分項順序)。
	lines := b.session.ScoreLines()
	y := 148.0
	for i, ln := range lines {
		col := color.RGBA{206, 218, 240, 255}
		size := 14.0
		if i == len(lines)-1 { // 總分
			vector.StrokeRect(dst, hsPanelX+28, float32(y)-6, hsPanelW-56, 1, 1, color.RGBA{90, 110, 150, 255}, false)
			y += 10
			col, size = edge, 18
		}
		b.fnt.Draw(dst, ln.Label, hsPanelX+34, y, size, col)
		// 分數靠右對齊:uifont 沒有 DrawRight,量出寬度自己減(數字欄位對齊比左對齊好讀太多)。
		val := fmt.Sprintf("%d", ln.Value)
		w, _ := b.fnt.Measure(val, size)
		b.fnt.Draw(dst, val, hsPanelX+hsPanelW-34-w, y, size, col)
		y += size + 10
	}

	fillPanel(dst, 270, 388, 100, 24, color.RGBA{30, 40, 70, 255}, false)
	vector.StrokeRect(dst, 270, 388, 100, 24, 1, edge, false)
	b.fnt.DrawCentered(dst, b.tr("繼續", "CONTINUE"), 320, 400, 13, color.RGBA{220, 228, 242, 255})
}
