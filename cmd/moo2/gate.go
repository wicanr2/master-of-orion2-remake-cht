package main

// gate.go:星圖上的星門標記(原版 `Draw_Gate_Icons_` @ 0x8439A / `Draw_A_Gate_Icon_` @ 0x83741)。
//
// 兩種星門都是 Achievement 科技,研究到就**在自己每個有殖民地的星系各生一個門**
// (手冊逐字:「forms a … wormhole terminus in each system in which you have … a colony」),
// 不必逐星建造。所以標記的規則很簡單:有科技 + 那顆星是自己的殖民地。
//
// 規則面(加速多少、幾回合到)在 `internal/shell/starlane.go`。
//
// ⚠ **這不是原版的畫法。** 原版的 `Draw_A_Gate_Icon_` 是一支 330 行的函式,畫的是逐格動畫,
// 資產來源沒有解出來(不是字串常數,得再追一層)。這裡先用一個雙環標記把**資訊**呈現出來——
// 玩家要看得出「這顆星有門」,否則那兩條速度規則等於隱形。追到資產之後再換成原版動畫。

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	// jumpGateTint 是躍遷門的標記色(暫時性的蟲洞終端,偏冷)。
	jumpGateTint = color.RGBA{120, 200, 235, 255}
	// starGateTint 是星際之門的標記色(穩定的終端,偏金)。
	starGateTint = color.RGBA{245, 210, 120, 255}
)

// drawGateIcons 在玩家的殖民地星上標出星門。
//
// 畫在星星**之後**(蓋在上面):它是狀態標示不是背景。星際之門優先於躍遷門——
// 兩者都有時,能一回合到的那個才是玩家實際用到的。
func (b *sceneBuilder) drawGateIcons(dst *ebiten.Image, visible []bool) {
	sess := b.session
	if sess == nil {
		return
	}
	hasStar, hasJump := sess.PlayerHasStarGate(), sess.PlayerHasJumpGate()
	if !hasStar && !hasJump {
		return
	}
	tint := jumpGateTint
	if hasStar {
		tint = starGateTint
	}
	for _, idx := range sess.PlayerColonyStars {
		if idx < 0 || idx >= len(sess.Stars) {
			continue
		}
		if visible != nil && idx < len(visible) && !visible[idx] {
			continue
		}
		st := sess.Stars[idx]
		x := float32(starVX0) + float32(st.X)*(starVX1-starVX0)
		y := float32(starVY0) + float32(st.Y)*(starVY1-starVY0)
		// 雙環:外環細、內環更細,和「擁有環」「星雲環」在視覺上區隔開。
		vector.StrokeCircle(dst, x, y, 11, 1, tint, true)
		vector.StrokeCircle(dst, x, y, 8, 1, tint, true)
	}
}
