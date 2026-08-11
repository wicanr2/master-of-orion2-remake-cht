package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// hotseatEmpireSelectScreen 是 remake 在新遊戲生成帝國後加入的熱座接管清單。
// 原版的真人旗標是在同一個新局設定階段逐一標記;remake 的基礎資料模型先生成
// 玩家 + AI,所以把同一個選擇放在命名/旗色之後、進星圖之前，保留「逐一指定帝國」
// 的語意而不再依 AI 陣列尾端猜測。
type hotseatEmpireSelectScreen struct {
	b        *sceneBuilder
	fnt      *uifont.Font
	selected []bool
	hover    int
	required int
}

func (b *sceneBuilder) hotseatEmpireSelect() (origScreen, error) {
	if b.session == nil {
		return nil, fmt.Errorf("熱座選帝國時沒有作用中的對局")
	}
	required := b.pendingHotseat - 1
	if required < 1 {
		return nil, fmt.Errorf("熱座真人席位數不足")
	}
	if required > len(b.session.AIPlayers) {
		// 這只應該發生在舊流程或外部測試手動注入設定時;降到可用席位
		// 並記錄回 builder,不讓 UI 顯示一個永遠無法完成的選擇。
		required = len(b.session.AIPlayers)
		b.pendingHotseat = required + 1
	}
	selected := make([]bool, len(b.session.AIPlayers))
	for i := 0; i < required; i++ {
		selected[i] = true
	}
	return &hotseatEmpireSelectScreen{
		b: b, fnt: b.fnt, selected: selected, hover: -1, required: required,
	}, nil
}

const (
	hseListX = 70
	hseListY = 116
	hseListW = 500
	hseRowH  = 42
)

func (s *hotseatEmpireSelectScreen) rowRect(i int) (int, int, int, int) {
	return hseListX, hseListY + i*hseRowH, hseListW, hseRowH - 5
}

func (s *hotseatEmpireSelectScreen) acceptRect() (int, int, int, int) {
	return 470, 440, 130, 28
}

func (s *hotseatEmpireSelectScreen) cancelRect() (int, int, int, int) {
	return 40, 440, 130, 28
}

func (s *hotseatEmpireSelectScreen) selectedCount() int {
	n := 0
	for _, on := range s.selected {
		if on {
			n++
		}
	}
	return n
}

func (s *hotseatEmpireSelectScreen) selectedIndices() []int {
	indices := make([]int, 0, s.selectedCount())
	for i, on := range s.selected {
		if on {
			indices = append(indices, i)
		}
	}
	return indices
}

func (s *hotseatEmpireSelectScreen) update(in shell.InputState) *origTransition {
	s.hover = -1
	for i := range s.selected {
		x, y, w, h := s.rowRect(i)
		if hitBox(in.MouseX, in.MouseY, x, y, w, h) {
			s.hover = i
			break
		}
	}
	if !in.ClickReleased {
		return nil
	}
	if s.hover >= 0 {
		if s.selected[s.hover] || s.selectedCount() < s.required {
			s.selected[s.hover] = !s.selected[s.hover]
			if clickSound != nil {
				clickSound()
			}
		}
		return nil
	}
	if x, y, w, h := s.cancelRect(); hitBox(in.MouseX, in.MouseY, x, y, w, h) {
		if clickSound != nil {
			clickSound()
		}
		return &origTransition{next: s.b.nameFlag(s.b.session.PlayerName)}
	}
	if x, y, w, h := s.acceptRect(); hitBox(in.MouseX, in.MouseY, x, y, w, h) {
		if s.selectedCount() != s.required {
			return nil
		}
		indices := s.selectedIndices()
		s.b.pendingHotseatAI = indices
		s.b.applyPendingHotseat()
		if clickSound != nil {
			clickSound()
		}
		return s.b.goTo(s.b.galaxy, "星系主畫面")
	}
	return nil
}

func (s *hotseatEmpireSelectScreen) aiLabel(i int) string {
	if i < 0 || i >= len(s.b.session.AIPlayers) {
		return ""
	}
	ai := s.b.session.AIPlayers[i]
	if ai.RaceIndex >= 0 && ai.RaceIndex < len(shell.Races) {
		r := shell.Races[ai.RaceIndex]
		name := r.Name
		if s.b.lang != 0 {
			name = r.EnName
		}
		return fmt.Sprintf(s.b.tr("帝國 %d：%s", "Empire %d: %s"), i+2, name)
	}
	return fmt.Sprintf(s.b.tr("帝國 %d：%s", "Empire %d: %s"), i+2, ai.Name)
}

func (s *hotseatEmpireSelectScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{7, 10, 18, 255})
	if s.fnt == nil {
		return
	}

	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{216, 224, 240, 255}
	dim := color.RGBA{150, 160, 180, 255}
	s.fnt.DrawCentered(dst, s.b.tr("指定真人帝國", "CHOOSE HUMAN EMPIRES"), 320, 48, 19, gold)
	s.fnt.DrawCentered(dst,
		fmt.Sprintf(s.b.tr("請選 %d 個帝國給真人接手（玩家帝國之外）", "%d AI empires will be taken over by human players"), s.required),
		320, 76, 13, body)
	s.fnt.DrawCentered(dst,
		fmt.Sprintf(s.b.tr("已選 %d / %d", "Selected %d / %d"), s.selectedCount(), s.required),
		320, 96, 12, dim)

	for i := range s.selected {
		x, y, w, h := s.rowRect(i)
		bg := color.RGBA{22, 30, 48, 255}
		border := color.RGBA{70, 92, 126, 255}
		if s.selected[i] {
			bg = color.RGBA{42, 60, 78, 255}
			border = color.RGBA{224, 190, 92, 255}
		} else if i == s.hover {
			bg = color.RGBA{34, 48, 70, 255}
			border = color.RGBA{150, 180, 220, 255}
		}
		fillPanel(dst, float32(x), float32(y), float32(w), float32(h), bg, false)
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1.5, border, false)
		mark := "[ ]"
		if s.selected[i] {
			mark = "[X]"
		}
		s.fnt.Draw(dst, mark, float64(x+14), float64(y+h/2-7), 14, border)
		s.fnt.Draw(dst, s.aiLabel(i), float64(x+62), float64(y+h/2-7), 14, body)
	}

	drawButton := func(rect func() (int, int, int, int), label string, accent color.RGBA) {
		x, y, w, h := rect()
		fillPanel(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{28, 34, 48, 255}, false)
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1.5, accent, false)
		s.fnt.DrawCentered(dst, label, float64(x+w/2), float64(y+h/2), 14, body)
	}
	drawButton(s.cancelRect, s.b.tr("返回命名", "BACK TO NAME"), color.RGBA{160, 140, 100, 255})
	accent := color.RGBA{120, 200, 130, 255}
	if s.selectedCount() != s.required {
		accent = color.RGBA{90, 100, 110, 255}
	}
	drawButton(s.acceptRect, s.b.tr("開始遊戲", "START GAME"), accent)
}
