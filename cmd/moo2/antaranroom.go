package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
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
//	remake 取「55 幀累積後的最終畫格」當靜態背景,與外交議事廳(DIPLOMAT#29,38 幀)
//	同一個做法,見 internal/lbx.Image.AccumulatedRGBA。
//	⚠ 留白:原版是把這 55 幀當推鏡動畫播出來的,remake 只呈現最終定格。
//
// --- 規則來源 ---
//
//	手冊 p.183:「An alternate method is to seek out and defeat the Antaran home fleet.
//	This involves travelling to the Antaran homeworld, which is not possible until you have
//	the right technology and build a Dimensional Gate. … (This strategy is not available if
//	you disabled Antaran Attacks when setting up your game.)」
//
//	結算沿用既有的 shell.AssaultAntares(不在這一層重算戰鬥)。

// loadAntaranRoom 載入安塔蘭王座廳背景(antaroom.lbx 資產 1,以資產 0 為調色盤,55 幀累積)。
// 任何一步失敗都回 nil——畫面會退回純色底,不因為缺美術而整個進不去。
func loadAntaranRoom(res *assets.Resolver) *ebiten.Image {
	prov, err := decodeAsset(res, "antaroom.lbx", 0)
	if err != nil || prov.Embedded == nil {
		return nil
	}
	room, err := decodeAsset(res, "antaroom.lbx", 1)
	if err != nil || len(room.Frames) == 0 {
		return nil
	}
	return ebiten.NewImageFromImage(room.AccumulatedRGBA(prov.Embedded))
}

// antaranRoomScreen 是安塔蘭王座廳畫面。
type antaranRoomScreen struct {
	b    *sceneBuilder
	fnt  *uifont.Font
	room *ebiten.Image

	blockReason string // 非空 = 現在不能發動(逐條講明卡在哪)
	ourStrength int
	theirCount  int
	theirPower  int
}

func newAntaranRoomScreen(b *sceneBuilder) *antaranRoomScreen {
	s := &antaranRoomScreen{
		b: b, fnt: b.fnt, room: loadAntaranRoom(b.res),
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

func hitRect(in shell.InputState, x, y, w, h int) bool {
	return in.MouseX >= x && in.MouseX < x+w && in.MouseY >= y && in.MouseY < y+h
}

func (a *antaranRoomScreen) update(in shell.InputState) *origTransition {
	if !in.ClickReleased {
		return nil
	}
	if x, y, w, h := a.retreatRect(); hitRect(in, x, y, w, h) {
		return a.b.goTo(a.b.fleet, "艦隊列表")
	}
	if x, y, w, h := a.assaultRect(); hitRect(in, x, y, w, h) {
		if a.blockReason != "" || a.b.session == nil {
			return nil // 擋下的原因已經寫在畫面上,不用再彈訊息
		}
		if _, ok := a.b.session.AssaultAntares(); ok {
			return a.b.goTo(a.b.battleResult, "戰鬥結果")
		}
		// 理論上 blockReason 為空就一定 ok;真的失敗就重算一次理由顯示出來,不靜默。
		a.blockReason = a.b.session.AssaultAntaresBlockReason()
	}
	return nil
}

func (a *antaranRoomScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{10, 6, 4, 255})
	if a.room != nil {
		drawPanelImage(dst, a.room, nil)
	}
	if a.fnt == nil {
		return
	}
	gold := color.RGBA{240, 210, 120, 255}
	body := color.RGBA{238, 230, 214, 255}
	warn := color.RGBA{235, 140, 110, 255}

	// 標題帶:背景是滿版美術,文字直接疊上去會看不清,壓一條半透明深色。
	fillPanel(dst, 0, 24, moo2ScreenW, 74, color.RGBA{6, 4, 2, 175}, false)
	a.fnt.DrawCentered(dst, a.b.tr("安塔蘭王座廳", "THE ANTARAN THRONE ROOM"), 320, 46, 22, gold)
	a.fnt.DrawCentered(dst, a.b.tr("次元傳送門的彼端,安塔蘭統治者在等著。",
		"Beyond the dimensional gate, the Antaran overlords are waiting."), 320, 78, 14, body)

	// 戰力對比:同一套 playerMilitary,與實際結算用的數字一致。
	fillPanel(dst, 0, 300, moo2ScreenW, 76, color.RGBA{6, 4, 2, 175}, false)
	a.fnt.DrawCentered(dst,
		fmt.Sprintf(a.b.tr("安塔蘭母星防禦艦隊:%d 艘,總戰力 %d", "Antaran home fleet: %d ships, %d combat power"),
			a.theirCount, a.theirPower), 320, 320, 14, warn)
	a.fnt.DrawCentered(dst,
		fmt.Sprintf(a.b.tr("我方艦隊總戰力:%d", "Your fleet: %d combat power"), a.ourStrength), 320, 344, 14, body)
	odds, oddsCol := a.b.tr("勝算渺茫——這一戰要求對方全滅,帶不夠戰力等於送死",
		"Long odds — this fight demands total annihilation; arriving under-armed is suicide"), warn
	if a.ourStrength >= a.theirPower {
		odds, oddsCol = a.b.tr("戰力已足以一戰", "Your fleet is strong enough to make the attempt"),
			color.RGBA{150, 220, 150, 255}
	}
	a.fnt.DrawCentered(dst, odds, 320, 366, 12, oddsCol)

	ax, ay, aw, ah := a.assaultRect()
	rx, ry, rw, rh := a.retreatRect()
	enabled := a.blockReason == ""
	face, edge := color.RGBA{58, 20, 16, 235}, color.RGBA{190, 110, 70, 255}
	if !enabled {
		face, edge = color.RGBA{34, 30, 28, 235}, color.RGBA{96, 88, 80, 255}
	}
	fillPanel(dst, float32(ax), float32(ay), float32(aw), float32(ah), face, false)
	vector.StrokeRect(dst, float32(ax), float32(ay), float32(aw), float32(ah), 1.5, edge, false)
	lab, labCol := a.b.tr("發動終局反攻", "LAUNCH THE FINAL ASSAULT"), body
	if !enabled {
		labCol = color.RGBA{150, 142, 132, 255}
	}
	a.fnt.DrawCentered(dst, lab, float64(ax+aw/2), float64(ay+ah/2), 16, labCol)

	fillPanel(dst, float32(rx), float32(ry), float32(rw), float32(rh), color.RGBA{34, 30, 44, 235}, false)
	vector.StrokeRect(dst, float32(rx), float32(ry), float32(rw), float32(rh), 1.5, color.RGBA{140, 130, 170, 255}, false)
	a.fnt.DrawCentered(dst, a.b.tr("撤退", "WITHDRAW"), float64(rx+rw/2), float64(ry+rh/2), 16, body)

	if !enabled {
		fillPanel(dst, 0, 448, moo2ScreenW, 26, color.RGBA{6, 4, 2, 190}, false)
		a.fnt.DrawCentered(dst, a.b.tr("無法發動:", "Cannot launch: ")+a.blockReason, 320, 461, 12, warn)
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
