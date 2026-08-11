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

// hotseat.go:熱座交接畫面(原版 `Hotseat_Screen_` @ 0x628E2)。
//
// 熱座是多位真人同機輪流下令。換人時要有一個「把鍵盤交出去」的中斷畫面,
// 否則下一位玩家會直接看到上一位的星圖與艦隊配置——那等於白玩。
//
// ============ 版面(反組譯 `Draw_Hotseat_Screen_` @ 0x626D6)============
//
//	視窗**置中**:x = (0x280 − 寬)/2、y = (0x1E0 − 高)/2
//	                (`mov edx, 280h / sub edx, 寬 / sar eax, 1`,0x280=640、0x1E0=480)
//	文字起點:x + 0x0E(+14)、y + 0x46(+70)  (`add edi, 0Eh / add esi, 46h`)
//
// ⚠ remake 沒有這個視窗的美術:原版的底圖是 `dword_19B874` 指向的一張已載入影像,
// 而那個全域是在別處填的,追不到具體的 LBX 資產。所以用自繪面板 + 原版的**尺寸與
// 文字錨點**,不宣稱像素對齊。視窗尺寸取 320×200(能塞下交接文字的保守值)。
//
// 席位模型與交換機制見 internal/shell/hotseat.go。

// 原版錨點(見檔頭)。
const (
	hotseatWinW, hotseatWinH = 320, 200
	hotseatTextDX            = 14 // +0x0E
	hotseatTextDY            = 70 // +0x46
)

// hotseatScreen 是換人時的交接畫面。
type hotseatScreen struct {
	b     *sceneBuilder
	fnt   *uifont.Font
	name  string // 接手的玩家
	seat  int    // 接手的席位(0-based)
	total int    // 這一局共幾席
	note  string // 副標(例如「本回合已全數下令,世界即將推進」)
	// onDone 是玩家按下「接手」之後要做的事。用閉包而不是固定的畫面建構函式,
	// 因為交接之後的去向不只一種——交棒給下一位是回星圖,最後一位交完是推進世界。
	onDone func() *origTransition
}

func newHotseatScreen(b *sceneBuilder, seat int, name, note string, onDone func() *origTransition) *hotseatScreen {
	total := 1
	if b.session != nil {
		total = b.session.SeatCount()
	}
	return &hotseatScreen{b: b, fnt: b.fnt, seat: seat, total: total, name: name, note: note, onDone: onDone}
}

func (h *hotseatScreen) winRect() (int, int, int, int) {
	// 原版:置中。
	return (moo2ScreenW - hotseatWinW) / 2, (moo2ScreenH - hotseatWinH) / 2, hotseatWinW, hotseatWinH
}

func (h *hotseatScreen) okRect() (int, int, int, int) {
	x, y, w, _ := h.winRect()
	return x + w/2 - 55, y + hotseatWinH - 52, 110, 30
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

	cx := float64(x + w/2)
	h.fnt.DrawCentered(dst, truncateToWidth(h.fnt, h.b.tr("換人接手", "PASS THE KEYBOARD"), 18, float64(w-20)), cx, float64(y+30), 18, gold)
	// 原版的文字錨點:視窗左上 +14 / +70。
	tx := float64(x + hotseatTextDX)
	ty := float64(y + hotseatTextDY)
	h.fnt.Draw(dst, truncateToWidth(h.fnt, h.b.tr("下一位:", "Next: ")+hotseatNameLabel(h.b.lang, h.name), 14, float64(w-hotseatTextDX-14)), tx, ty, 14, body)
	h.fnt.Draw(dst, truncateToWidth(h.fnt, fmt.Sprintf(h.b.tr("(席位 %d / %d)", "(seat %d of %d)"), h.seat+1, h.total), 11, float64(w-hotseatTextDX-14)), tx, ty+20, 11, dim)
	h.fnt.Draw(dst, truncateToWidth(h.fnt, h.b.tr("請把鍵盤交給這位玩家,確認之後再按「接手」。",
		"Pass the keyboard, then press TAKE OVER."), 13, float64(w-hotseatTextDX-14)), tx, ty+44, 13, dim)
	if h.note != "" {
		h.fnt.Draw(dst, truncateToWidth(h.fnt, h.note, 12, float64(w-hotseatTextDX-14)), tx, ty+68, 12, color.RGBA{200, 180, 110, 255})
	}

	bx, by, bw, bh := h.okRect()
	fillPanel(dst, float32(bx), float32(by), float32(bw), float32(bh), color.RGBA{40, 48, 66, 255}, false)
	vector.StrokeRect(dst, float32(bx), float32(by), float32(bw), float32(bh), 1.5, color.RGBA{150, 165, 200, 255}, false)
	h.fnt.DrawCentered(dst, truncateToWidth(h.fnt, h.b.tr("接手", "TAKE OVER"), 15, float64(bw-10)), float64(bx+bw/2), float64(by+bh/2), 15, body)
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

// finishResolvedTurn 處理已經完成 EndTurn 的後半段:自動存檔、研究抉擇、
// 結局／事件／回合摘要。網路鎖步在所有玩家重播同一批指令後只呼叫一次
// EndTurn，接著也必須走完全相同的原版畫面順序。
func (b *sceneBuilder) finishResolvedTurn() *origTransition {
	if b.savePath != "" { // 每回合自動存檔(持久化對局)
		if err := b.session.Save(b.savePath); err != nil {
			fmt.Fprintln(os.Stderr, "自動存檔失敗:", err)
		}
	}
	// 若本回合完成的研究主題有多科技可選 → 先進抉擇畫面(MOO2 每主題擇一),
	// 選定後再顯示回合摘要。
	if _, _, pending := b.session.PendingResearchChoice(); pending {
		sc, err := b.researchChoice(b.turnSummary)
		if err == nil {
			return &origTransition{next: sc}
		}
	}
	// 對局已分出勝負 → 先播結局過場,再進最終得分(原版也是這個順序)。
	// 排在事件快報之前:遊戲都結束了,再播一則新聞快報只是擋路。
	// 過場載不動就直接跳到最終得分——結局片不該擋住結算。
	if b.session.Victory.Over {
		if !b.skipCutscenes {
			if sc := b.endingCutsceneFor(); sc != nil {
				return &origTransition{next: sc}
			}
		}
		return b.goTo(b.hiScore, "最終得分")
	}
	// 本回合有隨機事件或星系發現 → 先播快報(原版事件有專屬畫面,不是回合摘要裡一行字,
	// 見 eventscreen.go 檔頭),按「繼續」才進回合摘要。
	if b.currentReport() != nil {
		return b.goTo(b.eventScreen, "事件快報")
	}
	return b.goTo(b.turnSummary, "回合摘要")
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
	if b.networkTurn != nil {
		return b.submitNetworkTurn()
	}
	if !b.session.HotseatEnabled() {
		return b.advanceWorldTurn()
	}
	next, wrapped := b.session.AdvanceSeat()
	if !wrapped {
		return &origTransition{next: newHotseatScreen(b, next, b.session.SeatName(next), "",
			func() *origTransition { return b.goTo(b.galaxy, "星系主畫面") })}
	}
	// 繞回第 0 席:所有真人都下完令 → 交接畫面按下去才推進世界,
	// 這樣第一位玩家一接手就看到本回合的結算,不會錯過事件快報。
	return &origTransition{next: newHotseatScreen(b, next, b.session.SeatName(next),
		b.tr("本回合全員已下令,接手後結算。", "All players have moved; the turn resolves after you take over."),
		b.advanceWorldTurn)}
}
