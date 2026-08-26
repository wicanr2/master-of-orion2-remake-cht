package main

import (
	"errors"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// refit.go 是殖民地 BUILD QUEUE 的 REFIT 子畫面。原版會進設計庫挑同艦體設計；
// remake 目前尚未有可持久化的設計庫，故此頁明示顯示「自動最佳模板」，並在排入前
// 凍結目標 Ship，確保存檔與鎖步重播不會因日後研究改變而漂移。

const (
	refitListX, refitListY = 28, 76
	refitListW, refitListH = 584, 24
	refitListRows          = 10
	refitQueueX            = 350
	refitQueueY            = 424
	refitButtonW           = 126
	refitCancelX           = 492
)

type refitScreen struct {
	b        *sceneBuilder
	colony   int
	selected int
	msg      string
}

func (b *sceneBuilder) refitPopup(colony int) (origScreen, error) {
	if b.session == nil || colony < 0 || colony >= len(b.session.PlayerColonies) {
		return nil, &shell.RefitError{Code: shell.RefitErrorNoColony}
	}
	return &refitScreen{b: b, colony: colony, selected: -1}, nil
}

func (s *refitScreen) candidates() []shell.RefitCandidate {
	if s.b.session == nil {
		return nil
	}
	return s.b.session.RefitCandidates(s.colony)
}

func (s *refitScreen) selectedPreview() (shell.RefitJob, error) {
	candidates := s.candidates()
	if s.selected < 0 || s.selected >= len(candidates) {
		return shell.RefitJob{}, &shell.RefitError{Code: shell.RefitErrorSelectShip}
	}
	c := candidates[s.selected]
	return s.b.session.PreviewRefit(s.colony, c.FleetIndex, c.ShipIndex)
}

func localizedRefitError(lang i18n.Lang, err error) string {
	var refitErr *shell.RefitError
	if !errors.As(err, &refitErr) || refitErr == nil {
		return uiText(lang, "refit.error.unknown")
	}
	key := "refit.error." + string(refitErr.Code)
	if refitErr.Code == shell.RefitErrorNoUpgrade {
		return fmt.Sprintf(uiText(lang, key), refitErr.ShipName)
	}
	return uiText(lang, key)
}

func refitTitleTextRect() textSafeRect {
	return textSafeRect{x: 24, y: 14, w: 592, h: 32, insetX: 4, lineH: 32}
}

func refitSubtitleTextRect() textSafeRect {
	return textSafeRect{x: 24, y: 46, w: 592, h: 16, insetX: 4, lineH: 16}
}

func refitListTextRect(i int) textSafeRect {
	return textSafeRect{x: refitListX, y: refitListY + i*refitListH, w: refitListW, h: refitListH - 2,
		insetX: 6, insetY: 1, lineH: refitListH - 4}
}

func refitEmptyTextRect() textSafeRect {
	return textSafeRect{x: 32, y: 180, w: 576, h: 32, insetX: 4, insetY: 2, lineH: 28}
}

func refitPreviewSourceTextRect() textSafeRect {
	return textSafeRect{x: 38, y: 330, w: 556, h: 20, insetX: 2, insetY: 1, lineH: 18}
}

func refitPreviewDetailTextRect() textSafeRect {
	return textSafeRect{x: 38, y: 352, w: 556, h: 34, insetX: 2, lineH: 16}
}

func refitPreviewWarningTextRect() textSafeRect {
	return textSafeRect{x: 38, y: 388, w: 556, h: 16, insetX: 2, lineH: 16}
}

func refitPreviewPromptTextRect() textSafeRect {
	return textSafeRect{x: 38, y: 348, w: 556, h: 24, insetX: 2, insetY: 1, lineH: 22}
}

func refitButtonTextRect(x int) textSafeRect {
	return textSafeRect{x: x, y: refitQueueY, w: refitButtonW, h: 28, insetX: 4, insetY: 2, lineH: 24}
}

func refitMessageTextRect() textSafeRect {
	return textSafeRect{x: 30, y: 458, w: 580, h: 20, insetX: 2, insetY: 1, lineH: 18}
}

func (s *refitScreen) update(in shell.InputState) *origTransition {
	if !in.ClickReleased {
		return nil
	}
	candidates := s.candidates()
	for i := 0; i < len(candidates) && i < refitListRows; i++ {
		if hitBox(in.MouseX, in.MouseY, refitListX, refitListY+i*refitListH, refitListW, refitListH-2) {
			s.selected = i
			s.msg = ""
			if clickSound != nil {
				clickSound()
			}
			return nil
		}
	}
	if hitBox(in.MouseX, in.MouseY, refitCancelX, refitQueueY, refitButtonW, 28) {
		if clickSound != nil {
			clickSound()
		}
		sc, err := s.b.buildQueuePopup()
		if err != nil {
			return nil
		}
		return &origTransition{next: sc}
	}
	if hitBox(in.MouseX, in.MouseY, refitQueueX, refitQueueY, refitButtonW, 28) {
		job, err := s.selectedPreview()
		if err != nil {
			s.msg = localizedRefitError(s.b.lang, err)
			return nil
		}
		c := candidates[s.selected]
		if _, err := s.b.session.QueueRefit(s.colony, c.FleetIndex, c.ShipIndex); err != nil {
			s.msg = localizedRefitError(s.b.lang, err)
			return nil
		}
		if clickSound != nil {
			clickSound()
		}
		_ = job
		sc, err := s.b.buildQueuePopup()
		if err != nil {
			return nil
		}
		return &origTransition{next: sc}
	}
	return nil
}

func (s *refitScreen) componentName(name string) string {
	if s.b.lang == i18n.Traditional {
		return name
	}
	for _, opts := range [][]shell.Component{
		shell.WeaponOptions, shell.ArmorOptions, shell.ShieldOptions, shell.SpecialOptions,
	} {
		for _, c := range opts {
			if c.Name == name {
				return componentLabel(s.b.lang, c)
			}
		}
	}
	return name
}

func (s *refitScreen) draw(dst *ebiten.Image) {
	b := s.b
	if b.fnt == nil || b.session == nil {
		return
	}
	bg := color.RGBA{12, 18, 34, 255}
	panel := color.RGBA{30, 42, 72, 255}
	row := color.RGBA{42, 57, 92, 255}
	selected := color.RGBA{62, 96, 122, 255}
	body := color.RGBA{216, 224, 240, 255}
	dim := color.RGBA{150, 168, 198, 255}
	gold := color.RGBA{244, 222, 128, 255}
	ok := color.RGBA{154, 224, 174, 255}
	warn := color.RGBA{240, 168, 116, 255}

	dst.Fill(bg)
	fillPanel(dst, 18, 14, 604, 48, panel, false)
	refitTitleTextRect().drawCentered(dst, b.fnt, uiText(b.lang, "refit.title"), 18, gold)
	refitSubtitleTextRect().drawCentered(dst, b.fnt, uiText(b.lang, "refit.subtitle"), 11, dim)

	candidates := s.candidates()
	for i := 0; i < refitListRows; i++ {
		y := refitListY + i*refitListH
		fillPanel(dst, refitListX, float32(y), refitListW, refitListH-2, row, false)
		if i == s.selected {
			fillPanel(dst, refitListX+2, float32(y+2), refitListW-4, refitListH-6, selected, false)
		}
		if i >= len(candidates) {
			continue
		}
		c := candidates[i]
		label := fmt.Sprintf(uiText(b.lang, "refit.candidate.summary"),
			c.Ship.Name, shipClassLabel(b.lang, c.Ship.Class),
			s.componentName(c.Ship.Weapon), s.componentName(c.Ship.Armor),
			s.componentName(c.Ship.Shield), s.componentName(c.Ship.Special))
		refitListTextRect(i).drawLeft(dst, b.fnt, label, 11, body)
	}
	if len(candidates) == 0 {
		refitEmptyTextRect().drawCentered(dst, b.fnt, uiText(b.lang, "refit.empty"), 13, warn)
	}

	fillPanel(dst, 28, 324, 584, 84, panel, false)
	if job, err := s.selectedPreview(); err == nil {
		cost := b.session.RefitCostPPForPlayer(job.Source, job.Target)
		refitPreviewSourceTextRect().drawLeft(dst, b.fnt,
			fmt.Sprintf(uiText(b.lang, "refit.preview.source_target"), job.Source.Name, job.Target.Name), 12, gold)
		detail := fmt.Sprintf(uiText(b.lang, "refit.preview.detail"),
			s.componentName(job.Target.Weapon), s.componentName(job.Target.Armor),
			s.componentName(job.Target.Shield), s.componentName(job.Target.Special), cost)
		refitPreviewDetailTextRect().drawCenteredLines(dst, b.fnt, detail, 11, body)
		refitPreviewWarningTextRect().drawLeft(dst, b.fnt, uiText(b.lang, "refit.preview.scrap_warning"), 10, dim)
	} else {
		refitPreviewPromptTextRect().drawLeft(dst, b.fnt, uiText(b.lang, "refit.preview.select_prompt"), 12, dim)
	}

	for _, button := range []struct {
		x       int
		textKey string
		col     color.RGBA
	}{
		{refitQueueX, "refit.button.queue", ok},
		{refitCancelX, "refit.button.return", dim},
	} {
		fillPanel(dst, float32(button.x), refitQueueY, refitButtonW, 28, button.col, false)
		refitButtonTextRect(button.x).drawCentered(dst, b.fnt, uiText(b.lang, button.textKey), 12, bg)
	}
	if s.msg != "" {
		refitMessageTextRect().drawLeft(dst, b.fnt, s.msg, 11, warn)
	}
}
