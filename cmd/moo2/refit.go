package main

import (
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
		return nil, fmt.Errorf("無殖民地可改裝")
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
		return shell.RefitJob{}, fmt.Errorf("請先選一艘停泊艦艇")
	}
	c := candidates[s.selected]
	return s.b.session.PreviewRefit(s.colony, c.FleetIndex, c.ShipIndex)
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
			s.msg = err.Error()
			return nil
		}
		c := candidates[s.selected]
		if _, err := s.b.session.QueueRefit(s.colony, c.FleetIndex, c.ShipIndex); err != nil {
			s.msg = err.Error()
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
	fillPanel(dst, 18, 18, 604, 44, panel, false)
	b.fnt.DrawCentered(dst, b.tr("艦艇改裝", "SHIP REFIT"), 320, 30, 18, gold)
	b.fnt.DrawCentered(dst, b.tr("選擇停泊於此殖民地星系的戰鬥艦", "Choose a combat ship parked in this colony system"),
		320, 51, 11, dim)

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
		label := fmt.Sprintf("%s  ·  %s  ·  %s/%s/%s/%s",
			c.Ship.Name, shipClassLabel(b.lang, c.Ship.Class),
			s.componentName(c.Ship.Weapon), s.componentName(c.Ship.Armor),
			s.componentName(c.Ship.Shield), s.componentName(c.Ship.Special))
		b.fnt.Draw(dst, truncateToWidth(b.fnt, label, 11, refitListW-12),
			refitListX+6, float64(y+5), 11, body)
	}
	if len(candidates) == 0 {
		b.fnt.DrawCentered(dst, b.tr("這個星系沒有可改裝的停泊戰鬥艦",
			"No stationary refittable combat ship in this system"), 320, 196, 13, warn)
	}

	fillPanel(dst, 28, 324, 584, 84, panel, false)
	if job, err := s.selectedPreview(); err == nil {
		cost := shell.RefitCostPP(job.Source, job.Target)
		b.fnt.Draw(dst, truncateToWidth(b.fnt, fmt.Sprintf(b.tr("來源：%s → 目標：%s（同艦體、自動最佳模板）",
			"Source: %s → target: %s (same hull, automatic best template)"),
			job.Source.Name, job.Target.Name), 12, 556), 38, 334, 12, gold)
		detail := fmt.Sprintf(b.tr("武器 %s　裝甲 %s　護盾 %s　特殊 %s　改裝成本 %d PP",
			"Weapon %s  Armor %s  Shield %s  Special %s  Refit cost %d PP"),
			s.componentName(job.Target.Weapon), s.componentName(job.Target.Armor),
			s.componentName(job.Target.Shield), s.componentName(job.Target.Special), cost)
		for i, line := range b.fnt.Wrap(detail, 11, 556) {
			b.fnt.Draw(dst, line, 38, float64(355+i*14), 11, body)
		}
		b.fnt.Draw(dst, b.tr("改裝中艦艇不可使用；從建造佇列移除會報廢來源艦。",
			"The ship is unavailable while refitting; removing the job scraps it."),
			38, 391, 10, dim)
	} else {
		b.fnt.Draw(dst, b.tr("請先選擇一艘可改裝戰鬥艦。",
			"Select a refittable combat ship to preview the automatic template."),
			38, 358, 12, dim)
	}

	for _, button := range []struct {
		x    int
		text string
		col  color.RGBA
	}{
		{refitQueueX, b.tr("排入改裝", "QUEUE REFIT"), ok},
		{refitCancelX, b.tr("返回", "RETURN"), dim},
	} {
		fillPanel(dst, float32(button.x), refitQueueY, refitButtonW, 28, button.col, false)
		b.fnt.DrawCentered(dst, button.text, float64(button.x+refitButtonW/2), refitQueueY+6, 12, bg)
	}
	if s.msg != "" {
		b.fnt.Draw(dst, truncateToWidth(b.fnt, s.msg, 11, 580), 30, 466, 11, warn)
	}
}
