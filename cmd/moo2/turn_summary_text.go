package main

import (
	"fmt"
	"image/color"

	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

func turnSummaryText(lang i18n.Lang, key string, args ...any) string {
	template := uiText(lang, key)
	if len(args) == 0 {
		return template
	}
	return fmt.Sprintf(template, args...)
}

func turnSummaryBuildNoticeText(lang i18n.Lang, notice shell.BuildNotice) string {
	switch notice.Kind {
	case shell.BuildNoticeCompleted:
		return turnSummaryText(lang, "turnsummary.build.completed", notice.ColonyIndex+1,
			buildItemLabel(lang, notice.Name))
	case shell.BuildNoticeRefitCompleted:
		return turnSummaryText(lang, "turnsummary.build.refit_completed", notice.ColonyIndex+1, notice.Name)
	case shell.BuildNoticeRefitCancelled:
		return turnSummaryText(lang, "turnsummary.build.refit_cancelled", notice.ColonyIndex+1, notice.Name)
	case shell.BuildNoticeArtificialPlanet:
		return turnSummaryText(lang, "turnsummary.build.artificial_planet", notice.Name)
	default:
		return turnSummaryText(lang, "turnsummary.build.unknown")
	}
}

func turnSummaryBaseRect(row int) textSafeRect {
	return textSafeRect{x: 40, y: 62 + row*24, w: 320, h: 19, insetX: 2}
}

var turnSummaryDynamicRect = textSafeRect{x: 40, y: 168, w: 320, h: 138, insetX: 2, lineH: 19}

type turnSummaryMessage struct {
	text string
	size float64
	col  color.RGBA
}

func turnSummaryDynamicExtras(fnt *uifont.Font, messages []turnSummaryMessage) []extraText {
	if fnt == nil {
		return nil
	}
	type renderedLine struct {
		text string
		size float64
		col  color.RGBA
	}
	lines := make([]renderedLine, 0, turnSummaryDynamicRect.maxLines())
	overflow := false
	for _, message := range messages {
		wrapped := wrapToWidth(fnt, message.text, message.size, turnSummaryDynamicRect.contentWidth())
		for _, line := range wrapped {
			if len(lines) >= turnSummaryDynamicRect.maxLines() {
				overflow = true
				break
			}
			lines = append(lines, renderedLine{text: line, size: message.size, col: message.col})
		}
		if overflow {
			break
		}
	}
	if overflow && len(lines) > 0 {
		last := &lines[len(lines)-1]
		last.text = truncateToWidth(fnt, last.text+"…", last.size, turnSummaryDynamicRect.contentWidth())
	}
	out := make([]extraText, 0, len(lines))
	for i, line := range lines {
		out = append(out, extraText{
			x:    float64(turnSummaryDynamicRect.contentX()),
			y:    float64(turnSummaryDynamicRect.contentY() + i*turnSummaryDynamicRect.lineH),
			size: line.size, text: line.text, col: line.col,
			maxW: turnSummaryDynamicRect.contentWidth(),
		})
	}
	return out
}
