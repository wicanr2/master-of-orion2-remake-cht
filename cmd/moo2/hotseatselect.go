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
		return nil, fmt.Errorf("%s", uiText(b.lang, "hotseat.empire_select.error.no_session"))
	}
	required := b.pendingHotseat - 1
	if required < 1 {
		return nil, fmt.Errorf("%s", uiText(b.lang, "hotseat.empire_select.error.too_few_seats"))
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

func hotseatEmpireTitleTextRect() textSafeRect {
	return textSafeRect{x: 70, y: 30, w: 500, h: 32, insetX: 6, insetY: 2}
}

func hotseatEmpireInstructionTextRect() textSafeRect {
	return textSafeRect{x: 50, y: 64, w: 540, h: 24, insetX: 6, insetY: 2}
}

func hotseatEmpireCountTextRect() textSafeRect {
	return textSafeRect{x: 70, y: 88, w: 500, h: 20, insetX: 6, insetY: 1}
}

func (s *hotseatEmpireSelectScreen) markTextRect(i int) textSafeRect {
	x, y, _, h := s.rowRect(i)
	return textSafeRect{x: x + 10, y: y + 5, w: 42, h: h - 10, insetX: 3, insetY: 2}
}

func (s *hotseatEmpireSelectScreen) rowTextRect(i int) textSafeRect {
	x, y, w, h := s.rowRect(i)
	return textSafeRect{x: x + 58, y: y + 5, w: w - 70, h: h - 10, insetX: 3, insetY: 2}
}

func hotseatEmpireButtonTextRect(rect func() (int, int, int, int)) textSafeRect {
	x, y, w, h := rect()
	return textSafeRect{x: x, y: y, w: w, h: h, insetX: 5, insetY: 3}
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
		return s.b.goTo(s.b.galaxy, uiText(s.b.lang, "hotseat.empire_select.transition.galaxy"))
	}
	return nil
}

func (s *hotseatEmpireSelectScreen) aiLabel(i int) string {
	if i < 0 || i >= len(s.b.session.AIPlayers) {
		return ""
	}
	ai := s.b.session.AIPlayers[i]
	if ai.RaceIndex >= 0 {
		for _, race := range raceSelectList {
			if race.shellIdx == ai.RaceIndex {
				return fmt.Sprintf(uiText(s.b.lang, "hotseat.empire_select.row"), i+2,
					raceSelectEntryText(s.b.lang, race, "name"))
			}
		}
	}
	return fmt.Sprintf(uiText(s.b.lang, "hotseat.empire_select.row"), i+2, ai.Name)
}

func (s *hotseatEmpireSelectScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{7, 10, 18, 255})
	if s.fnt == nil {
		return
	}

	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{216, 224, 240, 255}
	dim := color.RGBA{150, 160, 180, 255}
	hotseatEmpireTitleTextRect().drawCentered(dst, s.fnt,
		uiText(s.b.lang, "hotseat.empire_select.title"), 17, gold)
	hotseatEmpireInstructionTextRect().drawCentered(dst, s.fnt,
		fmt.Sprintf(uiText(s.b.lang, "hotseat.empire_select.instruction"), s.required), 12, body)
	hotseatEmpireCountTextRect().drawCentered(dst, s.fnt,
		fmt.Sprintf(uiText(s.b.lang, "hotseat.empire_select.selected"), s.selectedCount(), s.required), 11, dim)

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
		mark := uiText(s.b.lang, "hotseat.empire_select.mark.off")
		if s.selected[i] {
			mark = uiText(s.b.lang, "hotseat.empire_select.mark.on")
		}
		s.markTextRect(i).drawCentered(dst, s.fnt, mark, 12, border)
		s.rowTextRect(i).drawLeft(dst, s.fnt, s.aiLabel(i), 12, body)
	}

	drawButton := func(rect func() (int, int, int, int), label string, accent color.RGBA) {
		x, y, w, h := rect()
		fillPanel(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{28, 34, 48, 255}, false)
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1.5, accent, false)
		hotseatEmpireButtonTextRect(rect).drawCentered(dst, s.fnt, label, 12, body)
	}
	drawButton(s.cancelRect, uiText(s.b.lang, "hotseat.empire_select.button.back"), color.RGBA{160, 140, 100, 255})
	accent := color.RGBA{120, 200, 130, 255}
	if s.selectedCount() != s.required {
		accent = color.RGBA{90, 100, 110, 255}
	}
	drawButton(s.acceptRect, uiText(s.b.lang, "hotseat.empire_select.button.start"), accent)
}
