package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

const (
	gameSettingsBGAsset     = 29
	gameSettingsToggleAsset = 9
	gameSettingsAcceptAsset = 10
	gameSettingsW           = 279
	gameSettingsH           = 378
	gameSettingsRowX        = 22
	gameSettingsRowY        = 34
	gameSettingsRowStep     = 17
)

var gameSettingsKeys = [...]string{
	"gamesettings.option.end_turn_summary",
	"gamesettings.option.end_turn_wait",
	"gamesettings.option.enemy_moves",
	"gamesettings.option.expanding_help",
	"gamesettings.option.auto_select_ships",
	"gamesettings.option.animations",
	"gamesettings.option.auto_select_colony",
	"gamesettings.option.show_relocation_lines",
	"gamesettings.option.show_gnn_report",
	"gamesettings.option.auto_delete_trade_housing",
	"gamesettings.option.auto_save_game",
	"gamesettings.option.serious_turn_summary",
	"gamesettings.option.ship_initiative",
}

type gameSettingsScreen struct {
	b          *sceneBuilder
	f          *uifont.Font
	x, y       int
	bg, accept *ebiten.Image
	toggle     [2]*ebiten.Image
	settings   shell.GameSettings
	helpIndex  int
	helpTick   int
	helpX      int
	helpY      int
}

func newGameSettingsScreen(b *sceneBuilder) *gameSettingsScreen {
	s := &gameSettingsScreen{b: b, f: b.fnt, x: (moo2ScreenW - gameSettingsW) / 2,
		y: (moo2ScreenH - gameSettingsH) / 2, settings: shell.DefaultGameSettings(), helpIndex: -1}
	if b.session != nil {
		s.settings = b.session.EffectiveGameSettings()
	}
	s.bg, _ = b.gameMenuImage(gameSettingsBGAsset, false)
	s.accept, _ = b.gameMenuImage(gameSettingsAcceptAsset, true)
	if prov, err := decodeAsset(b.res, gameMenuPalLBX, gameMenuPalAsst); err == nil && prov.Embedded != nil {
		if im, err := decodeAsset(b.res, gameMenuLBX, gameSettingsToggleAsset); err == nil {
			for i := 0; i < len(s.toggle) && i < len(im.Frames); i++ {
				s.toggle[i] = ebiten.NewImageFromImage(im.Frames[i].ToRGBA(prov.Embedded, true))
			}
		}
	}
	return s
}

func (s *gameSettingsScreen) rowRect(i int) (int, int, int, int) {
	return s.x + gameSettingsRowX, s.y + gameSettingsRowY + i*gameSettingsRowStep, 235, 16
}

func (s *gameSettingsScreen) acceptRect() (int, int, int, int) {
	w, h := 75, 20
	if s.accept != nil {
		w, h = s.accept.Bounds().Dx(), s.accept.Bounds().Dy()
	}
	return s.x + (gameSettingsW-w)/2, s.y + 337, w, h
}

func (s *gameSettingsScreen) option(i int) bool {
	v := &s.settings
	values := [...]bool{v.EndOfTurnSummary, v.EndOfTurnWait, v.EnemyMoves, v.ExpandingHelp,
		v.AutoSelectShips, v.Animations, v.AutoSelectColony, v.ShowRelocationLines,
		v.ShowGNNReport, v.AutoDeleteTradeGoodHousing, v.AutoSaveGame,
		v.ShowOnlySeriousTurnSummary, v.ShipInitiative}
	return i >= 0 && i < len(values) && values[i]
}

func (s *gameSettingsScreen) toggleOption(i int) {
	v := &s.settings
	switch i {
	case 0:
		v.EndOfTurnSummary = !v.EndOfTurnSummary
	case 1:
		v.EndOfTurnWait = !v.EndOfTurnWait
	case 2:
		v.EnemyMoves = !v.EnemyMoves
	case 3:
		v.ExpandingHelp = !v.ExpandingHelp
	case 4:
		v.AutoSelectShips = !v.AutoSelectShips
	case 5:
		v.Animations = !v.Animations
	case 6:
		v.AutoSelectColony = !v.AutoSelectColony
	case 7:
		v.ShowRelocationLines = !v.ShowRelocationLines
	case 8:
		v.ShowGNNReport = !v.ShowGNNReport
	case 9:
		v.AutoDeleteTradeGoodHousing = !v.AutoDeleteTradeGoodHousing
	case 10:
		v.AutoSaveGame = !v.AutoSaveGame
	case 11:
		v.ShowOnlySeriousTurnSummary = !v.ShowOnlySeriousTurnSummary
	case 12:
		v.ShipInitiative = !v.ShipInitiative
	}
}

func (s *gameSettingsScreen) update(in shell.InputState) *origTransition {
	if s.helpIndex >= 0 && s.helpTick < 10 {
		s.helpTick++
	}
	if !in.ClickReleased && !in.RightClickReleased {
		return nil
	}
	for i := range gameSettingsKeys {
		x, y, w, h := s.rowRect(i)
		if hitRect(in, x, y, w, h) {
			if in.RightClickReleased {
				s.helpIndex, s.helpTick = i, 0
				s.helpX, s.helpY = in.MouseX, in.MouseY
				return nil
			}
			s.toggleOption(i)
			return nil
		}
	}
	if s.helpIndex >= 0 {
		s.helpIndex = -1
		return nil
	}
	if in.RightClickReleased {
		return nil
	}
	x, y, w, h := s.acceptRect()
	if hitRect(in, x, y, w, h) {
		if s.b.session != nil {
			s.b.session.ApplyGameSettings(s.settings)
		}
		return s.b.goTo(s.b.galaxy, uiText(s.b.lang, "gamesettings.transition.galaxy"))
	}
	return nil
}

func (s *gameSettingsScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{8, 10, 16, 255})
	if s.bg != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(s.x), float64(s.y))
		drawPanelImage(dst, s.bg, op)
	} else {
		fillPanel(dst, float32(s.x), float32(s.y), gameSettingsW, gameSettingsH, color.RGBA{12, 18, 30, 255}, false)
		vector.StrokeRect(dst, float32(s.x), float32(s.y), gameSettingsW, gameSettingsH, 1, color.RGBA{90, 120, 170, 255}, false)
	}
	if s.f == nil {
		return
	}
	if s.bg == nil || s.b.lang == i18n.Traditional {
		textSafeRect{x: s.x + 16, y: s.y + 6, w: 247, h: 23, insetX: 3, insetY: 2}.drawCentered(dst, s.f, uiText(s.b.lang, "gamesettings.title"), 15, color.RGBA{220, 220, 235, 255})
	}
	for i, key := range gameSettingsKeys {
		x, y, _, _ := s.rowRect(i)
		frame := 0
		if s.option(i) {
			frame = 1
		}
		if s.toggle[frame] != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y+2))
			drawPanelImage(dst, s.toggle[frame], op)
		} else {
			vector.StrokeRect(dst, float32(x), float32(y+2), 20, 11, 1, color.RGBA{120, 160, 200, 255}, false)
			if s.option(i) {
				vector.DrawFilledRect(dst, float32(x+4), float32(y+5), 12, 5, color.RGBA{100, 220, 150, 255}, false)
			}
		}
		textSafeRect{x: x + 26, y: y, w: 207, h: 16, insetX: 2, insetY: 1}.drawLeft(dst, s.f, uiText(s.b.lang, key), 11, color.RGBA{218, 220, 232, 255})
	}
	x, y, w, h := s.acceptRect()
	if s.accept != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		drawPanelImage(dst, s.accept, op)
	} else {
		fillPanel(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{46, 52, 66, 255}, false)
	}
	if s.b.lang == i18n.Traditional || s.accept == nil {
		textSafeRect{x: x, y: y, w: w, h: h, insetX: 4, insetY: 2}.drawCentered(dst, s.f, uiText(s.b.lang, "gamesettings.button.accept"), 11, color.RGBA{232, 232, 238, 255})
	}
	s.drawHelp(dst)
}

type helpRect struct{ x, y, w, h int }

func helpExpandRect(sourceX, sourceY int, target helpRect, tick int, expanding bool) helpRect {
	if !expanding || tick >= 10 {
		return target
	}
	if tick < 0 {
		tick = 0
	}
	step := tick + 1
	return helpRect{
		x: sourceX + (target.x-sourceX)*step/10,
		y: sourceY + (target.y-sourceY)*step/10,
		w: target.w * step / 10,
		h: target.h * step / 10,
	}
}

func (s *gameSettingsScreen) drawHelp(dst *ebiten.Image) {
	if s.helpIndex < 0 || s.helpIndex >= len(gameSettingsKeys) || s.f == nil {
		return
	}
	target := helpRect{x: 76, y: 145, w: 488, h: 170}
	r := helpExpandRect(s.helpX, s.helpY, target, s.helpTick, s.settings.ExpandingHelp)
	fillPanel(dst, float32(r.x), float32(r.y), float32(r.w), float32(r.h), color.RGBA{8, 14, 28, 245}, false)
	vector.StrokeRect(dst, float32(r.x), float32(r.y), float32(r.w), float32(r.h), 2, color.RGBA{112, 152, 210, 255}, false)
	if s.settings.ExpandingHelp && s.helpTick < 10 {
		return
	}
	key := gameSettingsKeys[s.helpIndex]
	textSafeRect{x: target.x + 12, y: target.y + 10, w: target.w - 24, h: 25, insetX: 2, insetY: 2}.drawCentered(dst, s.f, uiText(s.b.lang, key), 15, color.RGBA{238, 218, 130, 255})
	body := uiText(s.b.lang, key+".help")
	lines := s.f.Wrap(body, 13, float64(target.w-36))
	y := float64(target.y + 48)
	for _, line := range lines {
		if y+17 > float64(target.y+target.h-10) {
			break
		}
		s.f.Draw(dst, line, float64(target.x+18), y, 13, color.RGBA{222, 228, 240, 255})
		y += 18
	}
}
