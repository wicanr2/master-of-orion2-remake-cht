package main

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// combatfx.go：CMBTSFX.LBX 的戰鬥視覺特效播放層。
//
// CMBTSFX 是標準 LBX 影像庫，不是 SOUND.LBX 音效；asset 0 已知是多幀爆炸圖，
// 其餘資產依原版表中的戰鬥特效序列作可選載入。缺少原版資產時保持既有幾何 token，
// 不讓遊戲啟動失敗。

const (
	combatFXExplosion = 0
	combatFXImpact    = 1
)

type combatFXSequence struct {
	frames    []*ebiten.Image
	frameHold int
}

type combatFXInstance struct {
	sequence int
	x, y     int
	start    int
}

// combatFXFrameAt 是可測試的動畫索引規則。CMBTSFX 的 FrameTime 單位在不同資產
// 仍以遊戲 tick 消費；這裡至少保證 delta 幀依序播放、結束後不再取越界索引。
func combatFXFrameAt(tick, start, frameCount, frameHold int) (frame int, active bool) {
	if frameCount <= 0 || tick < start {
		return 0, false
	}
	if frameHold < 1 {
		frameHold = 1
	}
	frame = (tick - start) / frameHold
	if frame >= frameCount {
		return 0, false
	}
	return frame, true
}

// loadCombatFX 只載入已有的、可由 CMBTSFX 解碼的序列。ID 0/1 分別作為爆炸與命中
// 的穩定 fallback；資產缺失或調色盤不完整時略過該序列。
func loadCombatFX(res *assets.Resolver) map[int]combatFXSequence {
	out := make(map[int]combatFXSequence)
	if res == nil {
		return out
	}
	for _, id := range []int{combatFXExplosion, combatFXImpact} {
		im, err := decodeAsset(res, "cmbtsfx.lbx", id)
		if err != nil || len(im.Frames) == 0 {
			continue
		}
		pal := im.Embedded
		if pal == nil {
			pal, err = resolvePalette(res, im, paletteChain{{"combat.lbx", 11}})
			if err != nil {
				continue
			}
		}
		frames := make([]*ebiten.Image, 0, len(im.Frames))
		for i := range im.Frames {
			// CMBTSFX 的多幀圖可能是 LBX delta frame；直接解單幀會
			// 把前一幀未被覆寫的像素清掉，爆炸動畫因此只剩碎片。用
			// 累積解碼還原每一幀實際要畫出的完整畫面。
			frames = append(frames, ebiten.NewImageFromImage(
				im.AccumulatedUpToRGBA(pal, i, im.KeyColor())))
		}
		hold := im.FrameTime
		if hold < 1 {
			hold = 2
		}
		out[id] = combatFXSequence{frames: frames, frameHold: hold}
	}
	return out
}

func (t *tacticalScreen) spawnCombatFX(sequence int, ship shell.CombatShip) {
	if _, ok := t.combatFX[sequence]; !ok {
		return
	}
	cellX, cellY, cellW, cellH := cellRect(ship.Col, ship.Row)
	t.fx = append(t.fx, combatFXInstance{sequence: sequence,
		x: cellX + cellW/2, y: cellY + cellH/2, start: t.tick})
}

func (t *tacticalScreen) pruneCombatFX() {
	active := t.fx[:0]
	for _, fx := range t.fx {
		seq, ok := t.combatFX[fx.sequence]
		if !ok {
			continue
		}
		if _, alive := combatFXFrameAt(t.tick, fx.start, len(seq.frames), seq.frameHold); alive {
			active = append(active, fx)
		}
	}
	t.fx = active
}

func (t *tacticalScreen) drawCombatFX(dst *ebiten.Image) {
	for _, fx := range t.fx {
		seq, ok := t.combatFX[fx.sequence]
		if !ok {
			continue
		}
		frame, active := combatFXFrameAt(t.tick, fx.start, len(seq.frames), seq.frameHold)
		if !active || frame < 0 || frame >= len(seq.frames) || seq.frames[frame] == nil {
			continue
		}
		im := seq.frames[frame]
		bounds := im.Bounds()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(fx.x-bounds.Dx()/2), float64(fx.y-bounds.Dy()/2))
		drawPanelImage(dst, im, op)
	}
}
