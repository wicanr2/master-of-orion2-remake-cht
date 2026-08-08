package main

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/smk"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// cutscene.go:Smacker 過場播放(原版 `Smack` 畫面)。
//
// MOO2 的片頭與各結局過場是**裸的 Smacker 檔**,只是沿用了 `.LBX` 副檔名
// (`INTRO.LBX` 開頭四個位元組就是 `SMK2`)。解碼器在 `internal/smk`。
//
// 播放策略:
//   - 影片原尺寸(片頭 480×160、獵戶座結局 288×208)在 640×480 上**置中、整數倍放大**。
//     整數倍是刻意的——1996 年的點陣素材做非整數縮放會糊掉,寧可留黑邊。
//   - 幀率由檔案的 rate 欄位決定(片頭 76 ms/幀 ≈ 13 fps),與 ebiten 的 60 tick/s
//     用累積毫秒對齊,不假設兩者同步。
//   - 點任意處或播完 → 離開。原版的過場也是可以按鍵跳過的。
//
// ⚠ 誠實留白:**只有畫面沒有聲音**。Smacker 的音軌是壓縮的(片頭第 0 軌 23748 bytes、
// 11025 Hz),解碼器目前跳過音訊區塊——那需要再實作一組 Smacker 音訊 Huffman,
// 與畫面是兩套獨立的編碼。

// cutsceneScreen 播放一段 Smacker 影片。
type cutsceneScreen struct {
	b    *sceneBuilder
	fnt  *uifont.Font
	dec  *smk.Decoder
	name string

	canvas *ebiten.Image // 影片原尺寸的畫布
	rgba   *image.RGBA
	scale  int
	offX   int
	offY   int

	accumMS  float64 // 距離下一幀還差多少毫秒
	done     bool
	next     func() (*overlayScreen, error)
	nextName string
}

// tickMS 是 ebiten 一個 tick 的毫秒數(固定 60 TPS)。
const tickMS = 1000.0 / 60.0

// newCutsceneScreen 開一段過場。載不動就回 nil,呼叫端直接跳過這段過場
// ——過場失敗不該擋住玩家進遊戲。
func newCutsceneScreen(b *sceneBuilder, lbxName string, next func() (*overlayScreen, error), nextName string) *cutsceneScreen {
	raw, err := b.res.Read(lbxName)
	if err != nil {
		return nil
	}
	dec, err := smk.Open(raw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "過場 %s 不是 Smacker:%v\n", lbxName, err)
		return nil
	}
	s := &cutsceneScreen{b: b, fnt: b.fnt, dec: dec, name: lbxName, next: next, nextName: nextName}
	s.rgba = image.NewRGBA(image.Rect(0, 0, dec.H.Width, dec.H.Height))
	s.canvas = ebiten.NewImage(dec.H.Width, dec.H.Height)
	// 整數倍放大到塞得下 640×480 的最大倍率(至少 1 倍)。
	s.scale = moo2ScreenW / dec.H.Width
	if v := moo2ScreenH / dec.H.Height; v < s.scale {
		s.scale = v
	}
	if s.scale < 1 {
		s.scale = 1
	}
	s.offX = (moo2ScreenW - dec.H.Width*s.scale) / 2
	s.offY = (moo2ScreenH - dec.H.Height*s.scale) / 2
	s.advance() // 先解第一幀,免得第一個畫面是空的
	return s
}

// advance 解下一幀並更新畫布。解到底或出錯就標記結束。
func (s *cutsceneScreen) advance() {
	if s.done {
		return
	}
	pix, pal, err := s.dec.DecodeNext()
	if err != nil {
		s.done = true
		return
	}
	for i, idx := range pix {
		p := int(idx) * 3
		s.rgba.Pix[i*4+0] = pal[p+0]
		s.rgba.Pix[i*4+1] = pal[p+1]
		s.rgba.Pix[i*4+2] = pal[p+2]
		s.rgba.Pix[i*4+3] = 0xFF
	}
	s.canvas.WritePixels(s.rgba.Pix)
	if s.dec.Position() >= s.dec.H.Frames {
		s.done = true
	}
}

func (s *cutsceneScreen) update(in shell.InputState) *origTransition {
	if in.ClickReleased { // 點任意處跳過(原版的過場也能按鍵跳過)
		return s.b.goTo(s.next, s.nextName)
	}
	if s.done {
		return s.b.goTo(s.next, s.nextName)
	}
	// 影片幀率(片頭 76 ms/幀)與 ebiten 的 60 TPS 不同步,用累積毫秒對齊。
	s.accumMS += tickMS
	for s.accumMS >= float64(s.dec.H.FrameRateMS) && !s.done {
		s.accumMS -= float64(s.dec.H.FrameRateMS)
		s.advance()
	}
	return nil
}

func (s *cutsceneScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.Black) // 影片不滿版時的黑邊
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(s.scale), float64(s.scale))
	op.GeoM.Translate(float64(s.offX), float64(s.offY))
	drawPanelImage(dst, s.canvas, op)
	if s.fnt != nil {
		s.fnt.DrawCentered(dst, s.b.tr("點擊跳過", "click to skip"), 320, 468, 11, color.RGBA{130, 140, 150, 255})
	}
}

// seekForGallery 讓截圖廊快轉到第 n 幀。
//
// 為什麼需要:片頭是 76 ms 一幀,而截圖廊一拍只有 16.7 ms——要等到有內容的畫面
// 得花上百拍。第 0 幀是全黑的(片頭從黑淡入),截了看不出解碼有沒有成功。
// **只給截圖廊用**,正常播放靠 update 裡的累積毫秒推進。
func (s *cutsceneScreen) seekForGallery(n int) {
	for i := 0; i < n && !s.done; i++ {
		s.advance()
	}
}

// intro 播放片頭,播完或點擊後進主選單。
// 影片載不動時直接回主選單——過場不該擋住玩家進遊戲。
func (b *sceneBuilder) intro() origScreen {
	if sc := newCutsceneScreen(b, gamedata.CutsceneFileFor(gamedata.CutsceneIntro), b.menu, "主選單"); sc != nil {
		return sc
	}
	s, err := b.menu()
	if err != nil {
		return nil
	}
	return s
}

// endingCutsceneFor 依這局的勝負挑一段結局過場(對映依據見 gamedata/cutscene.go 檔頭)。
// 沒有對應的過場、或影片載不動時回 nil,呼叫端直接跳到最終得分——結局片不該擋住結算。
func (b *sceneBuilder) endingCutsceneFor() origScreen {
	if b.session == nil || !b.session.Victory.Over {
		return nil
	}
	var kind gamedata.CutsceneKind
	switch {
	case b.session.Victory.Winner != "player":
		kind = gamedata.CutsceneDefeat
	case b.session.Victory.Reason == engine.VictoryAntaran:
		kind = gamedata.CutsceneAntaranWin
	default:
		kind = gamedata.CutsceneWin
	}
	name := gamedata.CutsceneFileFor(kind)
	if name == "" {
		return nil
	}
	// 結局片播完 → 最終得分(原版也是結局片接 Hall of Fame)。
	if sc := newCutsceneScreen(b, name, b.hiScore, "最終得分"); sc != nil {
		return sc
	}
	return nil
}
