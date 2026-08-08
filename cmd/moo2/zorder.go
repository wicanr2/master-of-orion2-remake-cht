package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// zorder.go:hi-res 畫布的 z 序保護。
//
// ============ 這個檔案為什麼存在 ============
//
// hi-res(2×)畫布把文字延到最後才以 2× 字級重播(見 internal/uifont/record.go),
// 代價是**「先畫字、再用面板蓋住字」這種寫法會失效**——字會浮在面板上面。
//
// remake 到處都是這種寫法:星圖畫完所有星名,再把指揮點數視窗疊上去;殖民地畫完清單,
// 再彈確認框。第一次跑 2× 畫廊時,指揮點數面板上就浮著「藤斯塔」「巴哈姆」兩個星名。
//
// 修法**不是逐處手動插屏障**——那要審 70 個呼叫點,而且漏一個就是看不見的回歸,
// 之後每加一個面板都可能再漏。改成把「填一塊不透明面板」這個動作本身變成屏障:
// cmd/moo2 裡**所有** `vector.DrawFilledRect` 都走這裡,自動先把已錄的字沖出去。
//
// uiScale==1 時 uifont 的 hook 是 nil,Barrier() 直接返回,這層完全零成本。

// fillPanel 填一塊矩形,並在填之前插入 z 序屏障(取代直接呼叫 vector.DrawFilledRect)。
func fillPanel(dst *ebiten.Image, x, y, w, h float32, clr color.Color, antialias bool) {
	uifont.BarrierRect(float64(x), float64(y), float64(w), float64(h))
	vector.DrawFilledRect(dst, x, y, w, h, clr, antialias)
}

// drawPanelImage 把一張圖貼上去,並在貼之前插入 z 序屏障(取代直接呼叫 dst.DrawImage)。
//
// 範圍取 src 的四個角經 op.GeoM 變換後的 AABB —— 旋轉/縮放都算得對,而**不需要**呼叫端
// 自己算座標。op 為 nil 時等同貼在原點。
func drawPanelImage(dst, src *ebiten.Image, op *ebiten.DrawImageOptions) {
	b := src.Bounds()
	x0, y0 := float64(b.Min.X), float64(b.Min.Y)
	x1, y1 := float64(b.Max.X), float64(b.Max.Y)
	if op != nil {
		cx := [4]float64{x0, x1, x0, x1}
		cy := [4]float64{y0, y0, y1, y1}
		for i := range cx {
			cx[i], cy[i] = op.GeoM.Apply(cx[i], cy[i])
		}
		x0, x1 = cx[0], cx[0]
		y0, y1 = cy[0], cy[0]
		for i := 1; i < 4; i++ {
			x0, x1 = min(x0, cx[i]), max(x1, cx[i])
			y0, y1 = min(y0, cy[i]), max(y1, cy[i])
		}
	}
	uifont.BarrierRect(x0, y0, x1-x0, y1-y0)
	dst.DrawImage(src, op)
}
