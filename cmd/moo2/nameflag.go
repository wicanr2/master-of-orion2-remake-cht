package main

// nameflag.go:原版新遊戲流程最後一步「命名 + 選旗色」(手冊 p.14)。
// 種族(或自訂種族)確定後 → 此畫面命名帝國、選旗幟顏色 → 進入遊戲。
// 版面為合成近似,尚未對原版截圖像素對齊。

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

type nameFlagScreen struct {
	b       *sceneBuilder
	fnt     *uifont.Font
	bg      *ebiten.Image
	name    []rune
	flagSel int
	hoverF  int
}

// nameFlagScreen 建命名/旗色畫面;suggested 為預填帝國名。
func (b *sceneBuilder) nameFlag(suggested string) origScreen {
	s := &nameFlagScreen{b: b, fnt: b.fnt, name: []rune(suggested), flagSel: 0, hoverF: -1}
	if im, err := decodeAsset(b.res, "raceopt.lbx", 0); err == nil && im.Embedded != nil {
		s.bg = ebiten.NewImageFromImage(im.Frames[0].ToRGBA(im.Embedded, im.KeyColor()))
	}
	return s
}

const nameMaxRunes = 16

// 旗色色塊版面。
func (s *nameFlagScreen) flagRect(i int) (int, int, int, int) {
	return 200 + i*50, 250, 40, 40
}
func (s *nameFlagScreen) cancelRect() (int, int, int, int) { return 40, 440, 120, 28 }
func (s *nameFlagScreen) acceptRect() (int, int, int, int) { return 480, 440, 120, 28 }

func (s *nameFlagScreen) update(in shell.InputState) *origTransition {
	// 鍵盤編輯名稱(headless 無鍵盤輸入,AppendInputChars 回空,維持預填名)。
	for _, r := range ebiten.AppendInputChars(nil) {
		if r >= 0x20 && len(s.name) < nameMaxRunes {
			s.name = append(s.name, r)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(s.name) > 0 {
		s.name = s.name[:len(s.name)-1]
	}

	s.hoverF = -1
	for i := range shell.FlagColors {
		if x, y, w, h := s.flagRect(i); hitBox(in.MouseX, in.MouseY, x, y, w, h) {
			s.hoverF = i
		}
	}
	if !in.ClickReleased {
		return nil
	}
	if s.hoverF >= 0 {
		s.flagSel = s.hoverF
		if clickSound != nil {
			clickSound()
		}
		return nil
	}
	if x, y, w, h := s.cancelRect(); hitBox(in.MouseX, in.MouseY, x, y, w, h) {
		if clickSound != nil {
			clickSound()
		}
		sc, err := s.b.raceSelect()
		if err != nil {
			return nil
		}
		return &origTransition{next: sc}
	}
	if x, y, w, h := s.acceptRect(); hitBox(in.MouseX, in.MouseY, x, y, w, h) {
		if clickSound != nil {
			clickSound()
		}
		if s.b.session != nil {
			name := string(s.name)
			if name == "" {
				name = "銀河帝國"
			}
			s.b.session.PlayerName = name
			s.b.session.FlagColor = s.flagSel
		}
		return s.b.goTo(s.b.galaxy, "星系主畫面")
	}
	return nil
}

func (s *nameFlagScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{0, 0, 0, 255})
	if s.bg != nil {
		dst.DrawImage(s.bg, nil)
	}
	if s.fnt == nil {
		return
	}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{206, 218, 240, 255}

	s.fnt.DrawCentered(dst, "為你的帝國命名", 320, 70, 18, gold)

	// 名稱輸入框。
	bx, by, bw, bh := 170, 140, 300, 40
	vector.DrawFilledRect(dst, float32(bx), float32(by), float32(bw), float32(bh), color.RGBA{20, 26, 40, 220}, false)
	vector.StrokeRect(dst, float32(bx), float32(by), float32(bw), float32(bh), 1.5, color.RGBA{110, 150, 210, 255}, false)
	name := string(s.name) + "_" // 尾端游標
	s.fnt.DrawCentered(dst, name, float64(bx+bw/2), float64(by+bh/2), 18, body)
	s.fnt.DrawCentered(dst, "(輸入名稱;可用鍵盤編輯)", 320, 200, 11, color.RGBA{150, 160, 180, 255})

	// 旗幟顏色。
	s.fnt.DrawCentered(dst, "選擇旗幟顏色", 320, 232, 13, gold)
	for i, fc := range shell.FlagColors {
		x, y, w, h := s.flagRect(i)
		vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{fc.R, fc.G, fc.B, 255}, false)
		bw2 := float32(1.5)
		bord := color.RGBA{80, 90, 110, 255}
		if i == s.flagSel {
			bord = gold
			bw2 = 3
		} else if i == s.hoverF {
			bord = color.RGBA{200, 210, 230, 255}
		}
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), bw2, bord, false)
	}
	if s.flagSel >= 0 && s.flagSel < len(shell.FlagColors) {
		s.fnt.DrawCentered(dst, "旗色:"+shell.FlagColors[s.flagSel].Name, 320, 312, 13, body)
	}

	drawBtn := func(rect func() (int, int, int, int), label string, accent color.RGBA) {
		x, y, w, h := rect()
		vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{34, 34, 44, 255}, false)
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1.5, accent, false)
		s.fnt.DrawCentered(dst, label, float64(x+w/2), float64(y+h/2), 14, body)
	}
	drawBtn(s.cancelRect, "返回", color.RGBA{160, 140, 100, 255})
	drawBtn(s.acceptRect, "開始遊戲", color.RGBA{120, 200, 130, 255})
}
