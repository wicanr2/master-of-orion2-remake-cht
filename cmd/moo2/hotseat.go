package main

import (
	"fmt"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// hotseat.go:熱座逐席隱私交接畫面。
//
// 熱座是多位真人同機輪流下令。換人時要有一個「把鍵盤交出去」的中斷畫面,
// 否則下一位玩家會直接看到上一位的星圖與艦隊配置——那等於白玩。
//
// IDA 證實 `sub_628E2 @ 0x628E2..0x62BB7` 是原版熱座相關互動流程，但不足以證明
// 本畫面的逐回合 privacy gate、尺寸或文字錨點。這裡採自繪 adapter，不宣稱像素對齊；
// 證據邊界見 docs/re/hotseat-handoff-ui-audit-20260827.md。
//
// 席位模型與交換機制見 internal/shell/hotseat.go。

const (
	hotseatWinW, hotseatWinH = 360, 230
)

// hotseatScreen 是換人時的交接畫面。
type hotseatScreen struct {
	b       *sceneBuilder
	fnt     *uifont.Font
	name    string // 接手的玩家
	seat    int    // 接手的席位(0-based)
	total   int    // 這一局共幾席
	noteKey string // ui.json 副標鍵；空字串表示不顯示
	// onDone 是玩家按下「接手」之後要做的事。用閉包而不是固定的畫面建構函式,
	// 因為交接之後的去向不只一種——交棒給下一位是回星圖,最後一位交完是推進世界。
	onDone func() *origTransition
}

func newHotseatScreen(b *sceneBuilder, seat int, name, noteKey string, onDone func() *origTransition) *hotseatScreen {
	total := 1
	if b.session != nil {
		total = b.session.SeatCount()
	}
	return &hotseatScreen{b: b, fnt: b.fnt, seat: seat, total: total, name: name, noteKey: noteKey, onDone: onDone}
}

func (h *hotseatScreen) winRect() (int, int, int, int) {
	// 原版:置中。
	return (moo2ScreenW - hotseatWinW) / 2, (moo2ScreenH - hotseatWinH) / 2, hotseatWinW, hotseatWinH
}

func (h *hotseatScreen) okRect() (int, int, int, int) {
	x, y, w, _ := h.winRect()
	return x + w/2 - 55, y + hotseatWinH - 45, 110, 30
}

func (h *hotseatScreen) titleTextRect() textSafeRect {
	x, y, w, _ := h.winRect()
	return textSafeRect{x: x + 10, y: y + 10, w: w - 20, h: 40, insetX: 4, insetY: 3}
}

func (h *hotseatScreen) nextPlayerTextRect() textSafeRect {
	x, y, w, _ := h.winRect()
	return textSafeRect{x: x + 14, y: y + 62, w: w - 28, h: 22, insetX: 2, insetY: 2}
}

func (h *hotseatScreen) seatTextRect() textSafeRect {
	x, y, w, _ := h.winRect()
	return textSafeRect{x: x + 14, y: y + 85, w: w - 28, h: 18, insetX: 2, insetY: 1}
}

func (h *hotseatScreen) instructionTextRect() textSafeRect {
	x, y, w, _ := h.winRect()
	return textSafeRect{x: x + 14, y: y + 106, w: w - 28, h: 36, insetX: 2, insetY: 1, lineH: 17}
}

func (h *hotseatScreen) noteTextRect() textSafeRect {
	x, y, w, _ := h.winRect()
	return textSafeRect{x: x + 14, y: y + 145, w: w - 28, h: 34, insetX: 2, insetY: 1, lineH: 16}
}

func (h *hotseatScreen) okTextRect() textSafeRect {
	x, y, w, hh := h.okRect()
	return textSafeRect{x: x, y: y, w: w, h: hh, insetX: 5, insetY: 3}
}

func (h *hotseatScreen) update(in shell.InputState) *origTransition {
	if !in.ClickReleased {
		return nil
	}
	if x, y, w, hh := h.okRect(); hitRect(in, x, y, w, hh) {
		return h.onDone()
	}
	return nil
}

func (h *hotseatScreen) draw(dst *ebiten.Image) {
	// 整片遮黑是這個畫面的**功能**,不是偷懶:交接時上一位玩家的星圖絕對不能露出來。
	dst.Fill(color.RGBA{0, 0, 0, 255})
	if h.fnt == nil {
		return
	}
	x, y, w, hh := h.winRect()
	fillPanel(dst, float32(x), float32(y), float32(w), float32(hh), color.RGBA{22, 26, 38, 255}, false)
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(hh), 2, color.RGBA{150, 165, 200, 255}, false)

	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{216, 224, 240, 255}
	dim := color.RGBA{140, 150, 170, 255}

	h.titleTextRect().drawCentered(dst, h.fnt, uiText(h.b.lang, "hotseat.handoff.title"), 18, gold)
	nextPlayer := fmt.Sprintf(uiText(h.b.lang, "hotseat.handoff.next_player"), hotseatNameLabel(h.b.lang, h.name))
	h.nextPlayerTextRect().drawLeft(dst, h.fnt, nextPlayer, 14, body)
	seat := fmt.Sprintf(uiText(h.b.lang, "hotseat.handoff.seat_position"), h.seat+1, h.total)
	h.seatTextRect().drawLeft(dst, h.fnt, seat, 11, dim)
	h.instructionTextRect().drawLeft(dst, h.fnt, uiText(h.b.lang, "hotseat.handoff.instruction"), 13, dim)
	if h.noteKey != "" {
		h.noteTextRect().drawLeft(dst, h.fnt, uiText(h.b.lang, h.noteKey), 12, color.RGBA{200, 180, 110, 255})
	}

	bx, by, bw, bh := h.okRect()
	fillPanel(dst, float32(bx), float32(by), float32(bw), float32(bh), color.RGBA{40, 48, 66, 255}, false)
	vector.StrokeRect(dst, float32(bx), float32(by), float32(bw), float32(bh), 1.5, color.RGBA{150, 165, 200, 255}, false)
	h.okTextRect().drawCentered(dst, h.fnt, uiText(h.b.lang, "hotseat.handoff.button.take_over"), 15, body)
}

// applyPendingHotseat 在新遊戲流程走完後,把多人設定畫面選的席位數與
// 選帝國畫面指定的 AI 索引套用到這一局。沒選過(0/1)就什麼都不做。
func (b *sceneBuilder) applyPendingHotseat() {
	if b.session == nil || b.pendingHotseat <= 1 {
		return
	}
	got := 0
	if b.pendingHotseatAI != nil {
		got = b.session.SetupHotseatWithAIIndices(b.pendingHotseatAI)
	} else {
		got = b.session.SetupHotseat(b.pendingHotseat)
	}
	if got < b.pendingHotseat {
		// 席位被可接管的 AI 對手數壓下來了,說一聲而不是默默少開一席。
		fmt.Fprintf(os.Stderr, "熱座:要求 %d 席,實際只開得出 %d 席(受 AI 對手數限制)\n",
			b.pendingHotseat, got)
	}
	b.pendingHotseat = 0
	b.pendingHotseatAI = nil
}

// advanceWorldTurn 真的推進一回合:先結算，再共用後續畫面流程。
func (b *sceneBuilder) advanceWorldTurn() *origTransition {
	b.session.EndTurn()
	return b.finishResolvedTurn()
}

// finishResolvedTurn 處理已經完成 EndTurn 的後半段:自動存檔、舊存檔研究抉擇相容、
// 結局／事件／回合摘要。網路鎖步在所有玩家重播同一批指令後只呼叫一次
// EndTurn，接著也必須走完全相同的原版畫面順序。
func (b *sceneBuilder) finishResolvedTurn() *origTransition {
	settings := b.session.EffectiveGameSettings()
	if settings.AutoSaveGame && b.savePath != "" { // SETTINGS 的 Auto Save Game 消費端
		if err := b.session.Save(b.savePath); err != nil {
			fmt.Fprintln(os.Stderr, "自動存檔失敗:", err)
		}
	}
	// 新流程在投入 RP 前已選 application；這裡只接住舊存檔可能留下的突破後待決狀態。
	if _, _, pending := b.session.PendingResearchChoice(); pending {
		b.stopContinuousTurns()
		sc, err := b.researchChoice(b.turnSummary)
		if err == nil {
			return &origTransition{next: sc}
		}
	}
	// 對局已分出勝負 → 先播結局過場,再進最終得分(原版也是這個順序)。
	// 排在事件快報之前:遊戲都結束了,再播一則新聞快報只是擋路。
	// 過場載不動就直接跳到最終得分——結局片不該擋住結算。
	if b.session.Victory.Over {
		b.stopContinuousTurns()
		if settings.Animations && !b.skipCutscenes {
			if sc := b.endingCutsceneFor(); sc != nil {
				return &origTransition{next: sc}
			}
		}
		return b.goTo(b.hiScore, uiText(b.lang, "hiscore.transition.screen"))
	}
	// 本回合有隨機事件或星系發現 → 先播快報(原版事件有專屬畫面,不是回合摘要裡一行字,
	// 見 eventscreen.go 檔頭),按「繼續」才進回合摘要。
	if b.shouldOpenReportScreen() {
		b.stopContinuousTurns()
		return b.goTo(b.eventScreen, uiText(b.lang, "event.transition.report"))
	}
	if !b.shouldShowTurnSummary() {
		return b.goTo(b.galaxy, uiText(b.lang, "gamesettings.transition.galaxy"))
	}
	b.stopContinuousTurns()
	return b.goTo(b.turnSummary, uiText(b.lang, "event.transition.summary"))
}

func (b *sceneBuilder) canRunContinuousTurns() bool {
	return b != nil && b.session != nil && !b.session.HotseatEnabled() &&
		b.networkTurn == nil && !b.networkPending
}

func (b *sceneBuilder) startContinuousTurns() {
	if !b.canRunContinuousTurns() || b.session.EffectiveGameSettings().EndOfTurnWait {
		b.stopContinuousTurns()
		return
	}
	b.continuousTurns = true
	b.continuousTurnAt = b.animTick + continuousTurnInterval
}

func (b *sceneBuilder) stopContinuousTurns() {
	if b == nil {
		return
	}
	b.continuousTurns = false
	b.continuousTurnAt = 0
}

// endTurnPressed 是「結束回合」鈕的完整流程。
//
// 單人局就是直接推進世界。熱座則是原版的節奏:每個人按一次只代表**自己**下完令了,
// 鍵盤交給下一位;等最後一位交完(繞回第 0 席),世界才推進一回合。
// 不這樣做的話,四個人的熱座局會在同一個星曆年裡跑掉四回合。
func (b *sceneBuilder) endTurnPressed() *origTransition {
	if b.session == nil {
		return nil
	}
	if b.session.EnsurePlayerResearchApplication() {
		b.stopContinuousTurns()
		if sc, err := b.researchChoice(b.galaxy); err == nil {
			return &origTransition{next: sc}
		}
	}
	if b.networkTurn != nil {
		b.stopContinuousTurns()
		return b.submitNetworkTurn()
	}
	if !b.session.HotseatEnabled() {
		b.startContinuousTurns()
		return b.advanceWorldTurn()
	}
	b.stopContinuousTurns()
	next, wrapped := b.session.AdvanceSeat()
	if !wrapped {
		return &origTransition{next: newHotseatScreen(b, next, b.session.SeatName(next), "",
			func() *origTransition { return b.goTo(b.galaxy, uiText(b.lang, "hotseat.transition.galaxy")) })}
	}
	// 繞回第 0 席:所有真人都下完令 → 交接畫面按下去才推進世界,
	// 這樣第一位玩家一接手就看到本回合的結算,不會錯過事件快報。
	return &origTransition{next: newHotseatScreen(b, next, b.session.SeatName(next),
		"hotseat.handoff.note.resolve",
		b.advanceWorldTurn)}
}
