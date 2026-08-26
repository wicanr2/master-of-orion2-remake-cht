package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// antaranroom.go:安塔蘭王座廳(原版 `Main_Antaran_Room`)。
//
// 這是三條勝利路徑之二的入口畫面。remake 先前把它做成「艦隊列表左下角一行文字」——
// 點下去直接跳戰鬥結果,中間沒有任何確認、沒有戰力對比、也看不出為什麼有時點了沒反應
// (前置條件不滿足時 CanAssaultAntares 回 false,而畫面上完全沒有跡象)。
//
// --- 美術來源 ---
//
//	反組譯 sub_14C83:`mov edx, 0 / mov eax, offset aAntaroomLbx / call sub_126B42`
//	——載入 antaroom.LBX **資產 0**;資產 1 是 640×480、55 幀的 delta 動畫(鏡頭推進到
//	安塔蘭統治者面前),資產 0 是它的調色盤提供者(自身無法單獨成畫)。
//
//	remake 依原版順序把 55 幀 delta 畫格逐幀累積播放；最後一幀仍保留作動畫
//	結束後的靜態 fallback。累積方法見 internal/lbx.Image.AccumulatedUpToRGBA。
//
// --- 規則來源 ---
//
//	手冊 p.183:「An alternate method is to seek out and defeat the Antaran home fleet.
//	This involves travelling to the Antaran homeworld, which is not possible until you have
//	the right technology and build a Dimensional Gate. … (This strategy is not available if
//	you disabled Antaran Attacks when setting up your game.)」
//
//	結算沿用既有的 shell.AssaultAntares(不在這一層重算戰鬥)。

// loadAntaranRoomFrames 載入安塔蘭王座廳背景(antaroom.lbx 資產 1,以資產 0 為調色盤)
// 的逐幀累積結果。任何一步失敗都回 nil——畫面會退回純色底,不因為缺美術而整個進不去。
func loadAntaranRoomFrames(res *assets.Resolver) []*ebiten.Image {
	prov, err := decodeAsset(res, "antaroom.lbx", 0)
	if err != nil || prov.Embedded == nil {
		return nil
	}
	room, err := decodeAsset(res, "antaroom.lbx", 1)
	if err != nil || len(room.Frames) == 0 {
		return nil
	}
	frames := make([]*ebiten.Image, len(room.Frames))
	for i := range room.Frames {
		frames[i] = ebiten.NewImageFromImage(room.AccumulatedUpToRGBA(prov.Embedded, i, room.KeyColor()))
	}
	return frames
}

// antaranRoomScreen 是安塔蘭王座廳畫面。
type antaranRoomScreen struct {
	b    *sceneBuilder
	fnt  *uifont.Font
	room *ebiten.Image
	// roomFrames 是 55 幀 delta 動畫的累積畫面；room 保留最後一幀作 fallback。
	roomFrames         []*ebiten.Image
	animationStartTick int

	blockReason shell.AntaranAssaultBlockReason
	ourStrength int
	theirCount  int
	theirPower  int
}

func newAntaranRoomScreen(b *sceneBuilder) *antaranRoomScreen {
	frames := loadAntaranRoomFrames(b.res)
	var room *ebiten.Image
	if len(frames) > 0 {
		room = frames[len(frames)-1]
	}
	s := &antaranRoomScreen{
		b: b, fnt: b.fnt, room: room, roomFrames: frames, animationStartTick: b.animTick,
		theirCount: shell.AntaranDefenseShipCount(),
		theirPower: shell.AntaranDefenseStrength(),
	}
	if b.session != nil {
		s.blockReason = b.session.AssaultAntaresBlockReason()
		s.ourStrength = b.session.PlayerFleetStrength()
	}
	return s
}

// 兩顆按鈕:發動反攻 / 撤退。座標無原版參照(openorion2 沒有這個畫面),
// 取畫面下緣兩側對稱擺放。
func (a *antaranRoomScreen) assaultRect() (int, int, int, int) { return 96, 396, 190, 44 }
func (a *antaranRoomScreen) retreatRect() (int, int, int, int) { return 354, 396, 190, 44 }

func antaranRoomTitleTextRect() textSafeRect {
	return textSafeRect{x: 20, y: 24, w: 600, h: 36, insetX: 4}
}

func antaranRoomSubtitleTextRect() textSafeRect {
	return textSafeRect{x: 20, y: 62, w: 600, h: 30, insetX: 4, insetY: 1}
}

func antaranRoomDefenseTextRect() textSafeRect {
	return textSafeRect{x: 20, y: 304, w: 600, h: 20, insetX: 4, insetY: 1}
}

func antaranRoomPlayerPowerTextRect() textSafeRect {
	return textSafeRect{x: 20, y: 328, w: 600, h: 20, insetX: 4, insetY: 1}
}

func antaranRoomOddsTextRect() textSafeRect {
	return textSafeRect{x: 20, y: 352, w: 600, h: 20, insetX: 4, insetY: 1}
}

func antaranRoomButtonTextRect(x, y, w, h int) textSafeRect {
	return textSafeRect{x: x, y: y, w: w, h: h, insetX: 5, insetY: 2}
}

func antaranRoomBlockTextRect() textSafeRect {
	return textSafeRect{x: 20, y: 450, w: 600, h: 22, insetX: 4, insetY: 1}
}

func antaranAssaultBlockText(lang i18n.Lang, reason shell.AntaranAssaultBlockReason) string {
	switch reason {
	case shell.AntaranAssaultGameOver:
		return uiText(lang, "antaran.room.block.game_over")
	case shell.AntaranAssaultEventsDisabled:
		return uiText(lang, "antaran.room.block.events_disabled")
	case shell.AntaranAssaultNoPortal:
		return uiText(lang, "antaran.room.block.no_portal")
	case shell.AntaranAssaultNoFleet:
		return uiText(lang, "antaran.room.block.no_fleet")
	default:
		return ""
	}
}

func hitRect(in shell.InputState, x, y, w, h int) bool {
	return in.MouseX >= x && in.MouseX < x+w && in.MouseY >= y && in.MouseY < y+h
}

func (a *antaranRoomScreen) update(in shell.InputState) *origTransition {
	if !in.ClickReleased {
		return nil
	}
	if x, y, w, h := a.retreatRect(); hitRect(in, x, y, w, h) {
		return a.b.goTo(a.b.fleet, uiText(a.b.lang, "antaran.room.transition.fleet"))
	}
	if x, y, w, h := a.assaultRect(); hitRect(in, x, y, w, h) {
		if a.blockReason != shell.AntaranAssaultAllowed || a.b.session == nil {
			return nil // 擋下的原因已經寫在畫面上,不用再彈訊息
		}
		if _, ok := a.b.session.AssaultAntares(); ok {
			return a.b.goTo(a.b.battleResult, uiText(a.b.lang, "antaran.room.transition.battle"))
		}
		// 理論上 blockReason 為空就一定 ok;真的失敗就重算一次理由顯示出來,不靜默。
		a.blockReason = a.b.session.AssaultAntaresBlockReason()
	}
	return nil
}

func (a *antaranRoomScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{10, 6, 4, 255})
	room := a.room
	if len(a.roomFrames) > 0 {
		frame := (a.b.animTick - a.animationStartTick) / 3
		if frame < 0 {
			frame = 0
		}
		if frame >= len(a.roomFrames) {
			frame = len(a.roomFrames) - 1
		}
		room = a.roomFrames[frame]
	}
	if room != nil {
		drawPanelImage(dst, room, nil)
	}
	if a.fnt == nil {
		return
	}
	gold := color.RGBA{240, 210, 120, 255}
	body := color.RGBA{238, 230, 214, 255}
	warn := color.RGBA{235, 140, 110, 255}

	// 標題帶:背景是滿版美術,文字直接疊上去會看不清,壓一條半透明深色。
	fillPanel(dst, 0, 24, moo2ScreenW, 74, color.RGBA{6, 4, 2, 175}, false)
	antaranRoomTitleTextRect().drawCentered(dst, a.fnt, uiText(a.b.lang, "antaran.room.title"), 22, gold)
	antaranRoomSubtitleTextRect().drawCentered(dst, a.fnt, uiText(a.b.lang, "antaran.room.subtitle"), 14, body)

	// 戰力對比:同一套 playerMilitary,與實際結算用的數字一致。
	fillPanel(dst, 0, 300, moo2ScreenW, 76, color.RGBA{6, 4, 2, 175}, false)
	antaranRoomDefenseTextRect().drawCentered(dst, a.fnt,
		fmt.Sprintf(uiText(a.b.lang, "antaran.room.defense"), a.theirCount, a.theirPower), 14, warn)
	antaranRoomPlayerPowerTextRect().drawCentered(dst, a.fnt,
		fmt.Sprintf(uiText(a.b.lang, "antaran.room.player_power"), a.ourStrength), 14, body)
	odds, oddsCol := uiText(a.b.lang, "antaran.room.odds.low"), warn
	if a.ourStrength >= a.theirPower {
		odds, oddsCol = uiText(a.b.lang, "antaran.room.odds.ready"),
			color.RGBA{150, 220, 150, 255}
	}
	antaranRoomOddsTextRect().drawCentered(dst, a.fnt, odds, 12, oddsCol)

	ax, ay, aw, ah := a.assaultRect()
	rx, ry, rw, rh := a.retreatRect()
	enabled := a.blockReason == shell.AntaranAssaultAllowed
	face, edge := color.RGBA{58, 20, 16, 235}, color.RGBA{190, 110, 70, 255}
	if !enabled {
		face, edge = color.RGBA{34, 30, 28, 235}, color.RGBA{96, 88, 80, 255}
	}
	fillPanel(dst, float32(ax), float32(ay), float32(aw), float32(ah), face, false)
	vector.StrokeRect(dst, float32(ax), float32(ay), float32(aw), float32(ah), 1.5, edge, false)
	lab, labCol := uiText(a.b.lang, "antaran.room.button.assault"), body
	if !enabled {
		labCol = color.RGBA{150, 142, 132, 255}
	}
	antaranRoomButtonTextRect(ax, ay, aw, ah).drawCentered(dst, a.fnt, lab, 16, labCol)

	fillPanel(dst, float32(rx), float32(ry), float32(rw), float32(rh), color.RGBA{34, 30, 44, 235}, false)
	vector.StrokeRect(dst, float32(rx), float32(ry), float32(rw), float32(rh), 1.5, color.RGBA{140, 130, 170, 255}, false)
	antaranRoomButtonTextRect(rx, ry, rw, rh).drawCentered(dst, a.fnt,
		uiText(a.b.lang, "antaran.room.button.retreat"), 16, body)

	if !enabled {
		fillPanel(dst, 0, 448, moo2ScreenW, 26, color.RGBA{6, 4, 2, 190}, false)
		antaranRoomBlockTextRect().drawCentered(dst, a.fnt,
			uiText(a.b.lang, "antaran.room.block.prefix")+antaranAssaultBlockText(a.b.lang, a.blockReason), 12, warn)
	}
}

// antaranRoom 進入安塔蘭王座廳。
func (b *sceneBuilder) antaranRoom() (origScreen, error) {
	playSceneBGM(trackAntaranRoom) // Main_Antaran_Room_Screen_ → STREAMHD #20
	if b.session == nil {
		return nil, fmt.Errorf("無對局")
	}
	return newAntaranRoomScreen(b), nil
}
