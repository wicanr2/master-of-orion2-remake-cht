package main

// buildqueue.go:原版的**建造彈出視窗**(`Build_Queue_Popup_` @ 0xB4041)。
//
// 為什麼要獨立一個畫面:remake 先前把建造佇列與可建清單塞在殖民地畫面的中段,而原版
// 那一整塊是**行星表面**(見 colonysurface.go 檔頭)。佇列在原版是另外彈出來的一張
// 全螢幕視窗——把它搬回去,殖民地畫面才騰得出地表的位置。
//
// ============ 版面(全部取自反組譯,不是量圖)============
//
//	框架:`COLBLDG.LBX#0`(640×480,中段透明,烘著 Auto Build / REFIT / DESIGN /
//	     REPEAT BUILD / CANCEL / OK 六顆鈕)。`Draw_Build_Queue_Popup_` @ 0xB3CF7 用
//	     `sub_12A478(0, 0, img)` 貼在 (0,0) —— 所以底下這些座標全是**螢幕絕對座標**。
//
//	可建清單 `Add_Buildings_Fields_` @ 0xB08CA:
//	     mov eax, 0Dh / mov ebx, 0B8h    → x 13..184
//	     y1 = i×19 + 20、y2 = (i+1)×19 + 19   → 列高 19,首列 y=20
//	     (那個 `sub ax, word_182AB9` 是捲動時的列高微調,平時為 0)
//
//	佇列 7 格 `Add_Build_Queue_Fields_` @ 0xB325A:
//	     mov eax, 0CFh / mov ebx, 1CAh   → x 207..458
//	     edi = 149h(329)、var_4 = 15Eh(350),每圈各 +14h(20)
//	     → 第 i 格 y 329+20i .. 350+20i
//
//	六顆鈕(`sub_1151B0(eax=x, edx=y)`;Auto Build 是 toggle `sub_11523B`):
//	     Auto Build   (490, 342)   COLBLDG#2  134×22
//	     REFIT        (492, 379)   COLBLDG#1   62×19
//	     DESIGN       (561, 379)   COLBLDG#3   64×19
//	     REPEAT BUILD (503, 411)   COLBLDG#6  113×21
//	     CANCEL       (493, 447)   COLBLDG#4   61×19
//	     OK           (560, 447)   COLBLDG#5   61×19
//	     (資產尺寸實測 lbxinfo,與座標相加都落在框架上那幾顆鈕的位置)
//
// remake 現已接上 REFIT、DESIGN、REPEAT BUILD 與 Auto Build：前者使用明示的
// 自動最佳模板近似（見 refit.go），後兩者的狀態與指令會保存並可重播。原版未安全
// 取得的「最佳」判斷不假稱精確；實際近似範圍見 docs/tech/colony-production-controls.md。

import (
	"errors"
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

const (
	bqChromeLBX   = "colbldg.lbx"
	bqChromeAsset = 0

	// 可建清單(反組譯真值)。
	bqListX0, bqListX1 = 13, 184
	bqListY0           = 20
	bqListStep         = 19
	bqListRows         = 21 // (423−20)/19 ≈ 21 列塞得進透明區

	// 佇列 7 格(反組譯真值)。
	bqQueueX0, bqQueueX1 = 207, 458
	bqQueueY0            = 329
	bqQueueStep          = 20
	bqQueueH             = 21
)

// bqButton 是框架上那六顆鈕:資產、位置、尺寸與操作。
type bqButton struct {
	asset      int
	x, y, w, h int
	act        string
	textKey    string
}

var bqButtons = []bqButton{
	{2, 490, 342, 134, 22, "auto", "buildqueue.button.auto"},
	{1, 492, 379, 62, 19, "refit", "buildqueue.button.refit"},
	{3, 561, 379, 64, 19, "design", "buildqueue.button.design"},
	{6, 503, 411, 113, 21, "repeat", "buildqueue.button.repeat"},
	{4, 493, 447, 61, 19, "cancel", "buildqueue.button.cancel"},
	{5, 560, 447, 61, 19, "ok", "buildqueue.button.ok"},
}

func buildQueueListTextRect(row int) textSafeRect {
	return textSafeRect{x: bqListX0 + 3, y: bqListY0 + row*bqListStep + 2, w: bqListX1 - bqListX0 - 6, h: 16}
}

func buildQueueSlotTextRect(row int) textSafeRect {
	return textSafeRect{x: bqQueueX0 + 3, y: bqQueueY0 + row*bqQueueStep + 2, w: bqQueueX1 - bqQueueX0 - 6, h: 16}
}

func buildQueueButtonTextRect(bt bqButton) textSafeRect {
	return textSafeRect{x: bt.x + 3, y: bt.y + 1, w: bt.w - 6, h: bt.h - 2}
}

func buildQueueTitleTextRect() textSafeRect {
	return textSafeRect{x: 210, y: 136, w: 250, h: 20}
}

func buildQueueHintTextRect() textSafeRect {
	return textSafeRect{x: 210, y: 160, w: 410, h: 16}
}

func buildQueueStatusTextRect() textSafeRect {
	return textSafeRect{x: 210, y: 180, w: 250, h: 16}
}

func buildQueueMessageTextRect() textSafeRect {
	return textSafeRect{x: 210, y: 204, w: 410, h: 36, lineH: 17}
}

type buildQueueScreen struct {
	b      *sceneBuilder
	idx    int
	chrome *ebiten.Image
	msg    string
	// repeatArmed 是玩家按下 REPEAT BUILD 後的下一次選取；可從左方清單或
	// 下方既有佇列項目選擇，避免把原版的「再點一個目標」縮成隱藏預設。
	repeatArmed bool
}

// buildQueuePopup 建造彈出視窗。操作的殖民地同 b.colonyIdx。
func (b *sceneBuilder) buildQueuePopup() (origScreen, error) {
	if b.session == nil || b.colonyIdx < 0 || b.colonyIdx >= len(b.session.PlayerColonies) {
		return nil, errors.New(uiText(b.lang, "buildqueue.error.no_colony"))
	}
	s := &buildQueueScreen{b: b, idx: b.colonyIdx}
	if im, err := decodeAsset(b.res, bqChromeLBX, bqChromeAsset); err == nil && im.Embedded != nil {
		s.chrome = ebiten.NewImageFromImage(im.Frames[0].ToRGBADropTranslucent(im.Embedded, im.KeyColor()))
	}
	return s, nil
}

func (s *buildQueueScreen) update(in shell.InputState) *origTransition {
	if !in.ClickReleased {
		return nil
	}
	b, sess := s.b, s.b.session

	for _, bt := range bqButtons {
		if !hitBox(in.MouseX, in.MouseY, bt.x, bt.y, bt.w, bt.h) {
			continue
		}
		if clickSound != nil {
			clickSound()
		}
		switch bt.act {
		case "cancel", "ok":
			return b.goTo(b.colonyScreen, uiText(b.lang, "buildqueue.transition.colony"))
		case "design":
			return b.goTo(b.shipDesign, uiText(b.lang, "buildqueue.transition.ship_design"))
		case "auto":
			on := !sess.ColonyAutoBuild(s.idx)
			sess.SetAutoBuild(s.idx, on)
			if on {
				s.msg = uiText(b.lang, "buildqueue.message.auto_enabled")
			} else {
				s.msg = uiText(b.lang, "buildqueue.message.auto_disabled")
			}
		case "refit":
			sc, err := b.refitPopup(s.idx)
			if err != nil {
				s.msg = err.Error()
				return nil
			}
			return &origTransition{next: sc}
		case "repeat":
			if active := sess.RepeatBuildFor(s.idx); active.Name != "" {
				sess.SetRepeatBuild(s.idx, "", 0)
				s.repeatArmed = false
				s.msg = uiText(b.lang, "buildqueue.message.repeat_cancelled")
			} else {
				s.repeatArmed = true
				s.msg = uiText(b.lang, "buildqueue.message.choose_repeat")
			}
		default:
			s.msg = uiText(b.lang, "buildqueue.message.unavailable")
		}
		return nil
	}

	// 佇列 7 格:點一下移除。
	for i := 0; i < shell.BuildQueueTotalSlots && i < 7; i++ {
		if hitBox(in.MouseX, in.MouseY, bqQueueX0, bqQueueY0+i*bqQueueStep, bqQueueX1-bqQueueX0, bqQueueH) {
			if clickSound != nil {
				clickSound()
			}
			if s.repeatArmed {
				q := sess.BuildQueueFor(s.idx)
				if i >= len(q) || q[i].Name == "" || !sess.SetRepeatBuild(s.idx, q[i].Name, q[i].Cost) {
					s.msg = uiText(b.lang, "buildqueue.message.repeat_invalid")
					return nil
				}
				s.repeatArmed = false
				s.msg = fmt.Sprintf(uiText(b.lang, "buildqueue.message.repeat_set"), buildItemLabel(b.lang, q[i].Name))
				return nil
			}
			sess.DequeueBuild(s.idx, i)
			return nil
		}
	}
	// 可建清單:點一下排進佇列。
	opts := b.colonyBuildChoices()
	for i := 0; i < bqListRows; i++ {
		if !hitBox(in.MouseX, in.MouseY, bqListX0, bqListY0+i*bqListStep, bqListX1-bqListX0, bqListStep) {
			continue
		}
		if j := b.colonyListTop + i; j >= 0 && j < len(opts) {
			if clickSound != nil {
				clickSound()
			}
			if s.repeatArmed {
				if sess.SetRepeatBuild(s.idx, opts[j].Name, opts[j].Cost) {
					s.repeatArmed = false
					s.msg = fmt.Sprintf(uiText(b.lang, "buildqueue.message.repeat_set"), buildItemLabel(b.lang, opts[j].Name))
				} else {
					s.msg = uiText(b.lang, "buildqueue.message.repeat_invalid")
				}
			} else {
				sess.EnqueueBuild(s.idx, opts[j].Name, opts[j].Cost)
			}
		}
		return nil
	}
	return nil
}

func (s *buildQueueScreen) draw(dst *ebiten.Image) {
	b, sess := s.b, s.b.session
	if b.fnt == nil || sess == nil {
		return
	}
	// 原版是把這張疊在殖民地畫面上;remake 沒有「保留下層」的機制,先鋪深色底。
	dst.Fill(color.RGBA{10, 12, 20, 255})
	if s.chrome != nil {
		drawPanelImage(dst, s.chrome, &ebiten.DrawImageOptions{})
	}

	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{214, 222, 238, 255}
	dim := color.RGBA{140, 152, 178, 255}

	// --- 左:可建清單 ---
	opts := b.colonyBuildChoices()
	for i := 0; i < bqListRows; i++ {
		j := b.colonyListTop + i
		if j >= len(opts) {
			break
		}
		o := opts[j]
		txt := buildItemLabel(b.lang, o.Name)
		if o.Cost > 0 {
			txt = fmt.Sprintf(uiText(b.lang, "buildqueue.list.cost"), buildItemLabel(b.lang, o.Name), o.Cost)
		}
		buildQueueListTextRect(i).drawLeft(dst, b.fnt, txt, 11, body)
	}
	if len(opts) == 0 {
		buildQueueListTextRect(0).drawLeft(dst, b.fnt, uiText(b.lang, "buildqueue.empty.available"), 11, dim)
	}

	// --- 右下:佇列 7 格 ---
	q := sess.BuildQueueFor(s.idx)
	for i := 0; i < 7; i++ {
		label := uiText(b.lang, "buildqueue.empty.slot")
		col := dim
		if i < len(q) && q[i].Name != "" {
			col = body
			label = buildItemLabel(b.lang, q[i].Name)
			if q[i].Refit != nil {
				label = fmt.Sprintf(uiText(b.lang, "buildqueue.queue.refit"), q[i].Refit.Source.Name)
			}
			if i == 0 && q[i].Cost > 0 {
				label = fmt.Sprintf(uiText(b.lang, "buildqueue.queue.progress"), label, q[i].Progress, q[i].Cost)
				if eta := sess.BuildETATurns(s.idx); eta > 0 {
					label += fmt.Sprintf(uiText(b.lang, "buildqueue.queue.eta"), eta)
				}
			}
		}
		buildQueueSlotTextRect(i).drawLeft(dst, b.fnt, label, 11, col)
	}

	// --- 六顆鈕 ---
	// 框架圖上**已經烘好**這六顆鈕了,所以不再另外貼一次資產(貼了會疊出雙重邊框)。
	// 中文模式擦底疊中文;英文模式讓路,露原版烘的 REFIT / DESIGN / CANCEL / OK。
	for _, bt := range bqButtons {
		if b.lang != i18n.Traditional && s.chrome != nil {
			continue
		}
		face, ink := color.RGBA{86, 88, 94, 255}, body
		if bt.act == "auto" && sess.ColonyAutoBuild(s.idx) {
			face, ink = color.RGBA{62, 108, 78, 255}, color.RGBA{232, 250, 232, 255}
		}
		if bt.act == "repeat" && sess.RepeatBuildFor(s.idx).Name != "" {
			face, ink = color.RGBA{118, 96, 52, 255}, color.RGBA{255, 242, 188, 255}
		}
		fillPanel(dst, float32(bt.x+3), float32(bt.y+3),
			float32(bt.w-6), float32(bt.h-6), face, false)
		buildQueueButtonTextRect(bt).drawCentered(dst, b.fnt, uiText(b.lang, bt.textKey), 11, ink)
	}

	// 標題與提示畫在框架中段的透明區上緣。
	name := fmt.Sprintf(uiText(b.lang, "buildqueue.title.default_colony"), s.idx+1)
	if n := sess.ColonyName(s.idx); n != "" {
		name = n
	}
	// 標題與提示放在框架中段的透明區(y 130..300 那塊空的),不要壓到下方的佇列格。
	buildQueueTitleTextRect().drawLeft(dst, b.fnt, fmt.Sprintf(uiText(b.lang, "buildqueue.title.format"), name), 14, gold)
	buildQueueHintTextRect().drawLeft(dst, b.fnt, uiText(b.lang, "buildqueue.hint.queue_remove"), 11, dim)
	status := ""
	if s.repeatArmed {
		status = uiText(b.lang, "buildqueue.status.choose_repeat")
	} else if repeat := sess.RepeatBuildFor(s.idx); repeat.Name != "" {
		status = fmt.Sprintf(uiText(b.lang, "buildqueue.status.repeat_active"), buildItemLabel(b.lang, repeat.Name))
	} else if sess.ColonyAutoBuild(s.idx) {
		status = uiText(b.lang, "buildqueue.status.auto_on")
	}
	if status != "" {
		buildQueueStatusTextRect().drawLeft(dst, b.fnt, status, 11, gold)
	}
	if s.msg != "" {
		buildQueueMessageTextRect().drawLeft(dst, b.fnt, s.msg, 11, color.RGBA{235, 160, 120, 255})
	}
}
