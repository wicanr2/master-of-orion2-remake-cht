package main

import (
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// inputbox.go:原版的**文字輸入彈窗**(`Remapped_Input_Box_Popup_` @ 0x91BB4 →
// `sub_91F14` 設版面 → `sub_91BD4` 繪製)。
//
// 這是網路多人最後缺的那一塊基礎設施:先前每一處要打字的地方都只能寫
// 「remake 沒有輸入框」——加入指定位址、改對局名稱、聊天列,全部卡在同一件事上。
//
// ============ 這不是自己設計的彈窗 ============
//
// remake 先前的判斷是「原版的輸入是內嵌欄位」,那是錯的:原版有一個**獨立的 modal 彈窗**,
// 連自己的 LBX 都有(`INBOX.LBX`,只有兩個資產)。找到它的路徑值得記:
// `Change_MP_Game_Name_` @ 0xF5777 呼叫 `sub_91BB4`,而符號表寫著
// `Remapped_Input_Box_Popup_` ——**先查被呼叫的函式叫什麼,再決定要做什麼**。
//
// ============ 版面(全部是立即數,相對於彈窗左上角)============
//
// `sub_91F14` 把呼叫端給的 (x, y) 攤成一組欄位:
//
//	word_19C391 = x                  ; 彈窗左上
//	word_19C393 = y
//	word_19C387 = y + 3              ; 標題帶的 y
//	word_19C384 = 0x36 = 54          ; 標題帶高(字在這 54 px 裡垂直置中)
//	word_19C38D = x + 0x22 = x + 34  ; 輸入欄 x
//	word_19C38F = y + 0x36 = y + 54  ; 輸入欄 y
//	word_19C39B = 0x1A = 26          ; 輸入欄高
//	word_19C389 = x + 0x60 = x + 96  ; OK 鈕 x
//	word_19C38B = y + 0x64 = y + 100 ; OK 鈕 y
//	word_19C399 = min(呼叫端的上限, 0xCD = 205)
//
// `sub_91BD4` 再給:輸入欄寬 = 資產0.寬 − 0x36 = 288 − 54 = **234**,
// 而標題文字**水平置中於彈窗寬**、垂直置中於那 54 px 的標題帶。
//
// ⚠ 輸入欄的左邊距 34、右邊距 288−34−234 = 20 —— **不對稱**。這是反組譯算出來的,
// 不是抄錯:0x22 與 0x36 是兩個獨立的立即數,沒有理由相等。照抄。
//
// 資產(lbxinfo 量的):`INBOX.LBX` 0 = 288×151(2 幀,自帶 32 色調色盤)、1 = 98×28(2 幀,OK 鈕)。
//
// 位置:`Change_MP_Game_Name_` 用 `Star_Name_Popup_Screen_Center_X_` @ 0x923BE = 0xB1 = 177、
// `_Y_` @ 0x923C4 = 0x7D = 125。x 幾乎正好是水平置中((640−288)/2 = 176),
// y=125 則明顯**高於**垂直置中(164)——那是原版選的位置,不是置中算式,所以照抄數字。
//
// ============ 誠實留白 ============
//
// 原版的輸入處理在 `sub_91B89` 那條 callback 上,逐鍵掃描碼處理(含 IME 之前的年代)。
// remake 用 ebiten 的 `AppendInputChars`——**這是移植決策**:掃描碼那一套在現代平台上
// 拿不到,而且會擋掉輸入法。代價是原版的某些鍵行為(如插入模式)沒有還原。
//
// 長度上限照原版的 205 上限夾;個別呼叫端另有更小的上限
//(對局名稱 8,真值取自 `Change_MP_Game_Name_` 的 `edx`,見 `netplay.GameNameMax`)。

const (
	inboxLBX      = "inbox.lbx"
	inboxBoxAsset = 0 // 288×151
	inboxOKAsset  = 1 // 98×28
	inboxBoxW     = 288
	inboxBoxH     = 151

	// 相對彈窗左上角的位移,全部取自 sub_91F14 的立即數。
	inboxTitleDY   = 3
	inboxTitleH    = 0x36             // 54
	inboxFieldDX   = 0x22             // 34
	inboxFieldDY   = 0x36             // 54
	inboxFieldH    = 0x1A             // 26
	inboxOKDX      = 0x60             // 96
	inboxOKDY      = 0x64             // 100
	inboxMaxLenCap = 0xCD             // 205
	inboxFieldW    = inboxBoxW - 0x36 // 234(sub_91BD4:資產0.寬 − 0x36)

	// OK 鈕的尺寸就是資產 1 的尺寸。
	inboxOKW, inboxOKH = 98, 28

	// 原版彈窗的預設位置(Star_Name_Popup_Screen_Center_X_/_Y_)。
	inboxDefaultX = 0xB1 // 177
	inboxDefaultY = 0x7D // 125

	// 游標閃爍週期(幀)。原版的閃法沒有抽出來,這是自己訂的。
	inboxCaretPeriod = 30
)

// inboxFieldRect 回傳輸入欄的螢幕矩形。
func inboxFieldRect(x, y int) (fx, fy, fw, fh int) {
	return x + inboxFieldDX, y + inboxFieldDY, inboxFieldW, inboxFieldH
}

// inboxInputTextRect 是輸入文字的實際安全欄。保留原本 x+6 與 fw-18 的
// 內縮，避免把游標或省略號畫到輸入欄右框；游標寬度由
// inboxVisibleInputText 額外從這個欄位保留。
func inboxInputTextRect(fx, fy, fw, fh int) textSafeRect {
	return textSafeRect{x: fx + 6, y: fy, w: fw - 18, h: fh, insetY: 1, lineH: fh - 2}
}

// inboxVisibleInputText 量測「內容 + 游標」的完整可見寬度。不能先把內容
// 截斷到整個欄寬再附加底線，因為底線會成為最後一個越框像素。
func inboxVisibleInputText(fnt *uifont.Font, text string, size float64, r textSafeRect) string {
	if fnt == nil || r.contentWidth() <= 0 {
		return ""
	}
	caretW, _ := fnt.Measure("_", size)
	contentW := r.contentWidth() - caretW
	if contentW <= 0 {
		return "_"
	}
	return truncateToWidth(fnt, text, size, contentW) + "_"
}

// inboxOKRect 回傳 OK 鈕的螢幕矩形。
func inboxOKRect(x, y int) (bx, by, bw, bh int) {
	return x + inboxOKDX, y + inboxOKDY, inboxOKW, inboxOKH
}

// inboxClampMaxLen 依原版把上限夾在 1..205。
func inboxClampMaxLen(n int) int {
	if n <= 0 || n > inboxMaxLenCap {
		return inboxMaxLenCap
	}
	return n
}

// inputBoxScreen 是文字輸入彈窗。
type inputBoxScreen struct {
	b        *sceneBuilder
	under    origScreen
	titleKey string
	text     []rune
	max      int
	// onOK 收使用者輸入的字串;回傳的轉場非 nil 就走它,nil 則回到下層。
	onOK func(string) *origTransition
	// onCancel 為 nil 時,取消就回下層。
	onCancel func() *origTransition

	x, y     int
	box, ok  *ebiten.Image
	okHi     *ebiten.Image
	tick     int
	hoverOK  bool
	scriptOK bool // 截圖廊/測試用:不吃鍵盤
}

// inputBox 疊一個文字輸入彈窗在目前的畫面上。titleKey 必須指向 ui.json，
// 不接受已翻譯玩家句子，避免共用 modal 的 caller 各自內嵌文案。
func (b *sceneBuilder) inputBox(under origScreen, titleKey, initial string, max int,
	onOK func(string) *origTransition) *inputBoxScreen {
	s := &inputBoxScreen{
		b: b, under: under, titleKey: titleKey, text: []rune(initial),
		max: inboxClampMaxLen(max), onOK: onOK,
		x: inboxDefaultX, y: inboxDefaultY,
	}
	if len(s.text) > s.max {
		s.text = s.text[:s.max]
	}
	s.box = b.inboxImage(inboxBoxAsset, 0, false)
	s.ok = b.inboxImage(inboxOKAsset, 0, true)
	s.okHi = b.inboxImage(inboxOKAsset, 1, true)
	return s
}

func inboxTitleTextRect(x, y int) textSafeRect {
	return textSafeRect{x: x + 10, y: y + inboxTitleDY, w: inboxBoxW - 20, h: inboxTitleH,
		insetX: 2, insetY: 2, lineH: inboxTitleH - 4}
}

func inboxOKTextRect(x, y int) textSafeRect {
	bx, by, bw, bh := inboxOKRect(x, y)
	return textSafeRect{x: bx, y: by, w: bw, h: bh, insetX: 5, insetY: 5, lineH: bh - 10}
}

func inboxHintTextRect(x, y int) textSafeRect {
	_, fy, _, fh := inboxFieldRect(x, y)
	_, by, _, _ := inboxOKRect(x, y)
	top := fy + fh
	return textSafeRect{x: x + 10, y: top, w: inboxBoxW - 20, h: by - top,
		insetX: 2, insetY: 1, lineH: by - top - 2}
}

// inboxImage 取 INBOX.LBX 的某資產某幀(調色盤取資產 0 自帶的那份)。
func (b *sceneBuilder) inboxImage(assetID, frame int, keyColor bool) *ebiten.Image {
	// 沒有資產解析器就沒有圖(單元測試、以及還沒指定 -data 的路徑)。
	// `decodeAsset` 對 nil resolver 是 panic 而不是回錯,所以擋在這裡。
	if b == nil || b.res == nil {
		return nil
	}
	prov, err := decodeAsset(b.res, inboxLBX, inboxBoxAsset)
	if err != nil || prov.Embedded == nil {
		return nil
	}
	im, err := decodeAsset(b.res, inboxLBX, assetID)
	if err != nil || frame >= len(im.Frames) {
		return nil
	}
	pal := prov.Embedded
	if im.Embedded != nil {
		pal = im.Embedded
	}
	return ebiten.NewImageFromImage(im.AccumulatedUpToRGBA(pal, frame, keyColor))
}

// Text 回傳目前的輸入內容(供測試與呼叫端)。
func (s *inputBoxScreen) Text() string { return string(s.text) }

// typeRunes 把字元加進去,超過上限就丟掉。抽成方法是為了讓測試不必模擬鍵盤。
func (s *inputBoxScreen) typeRunes(rs []rune) {
	for _, r := range rs {
		if r < 0x20 || r == 0x7f { // 控制字元不進緩衝區
			continue
		}
		if len(s.text) >= s.max {
			return
		}
		s.text = append(s.text, r)
	}
}

// backspace 刪一個字元(空字串時什麼都不做,不是 panic)。
func (s *inputBoxScreen) backspace() {
	if len(s.text) > 0 {
		s.text = s.text[:len(s.text)-1]
	}
}

// accept 是「按下 OK」——抽出來讓測試不必模擬點擊。
func (s *inputBoxScreen) accept() *origTransition {
	if s.onOK != nil {
		if t := s.onOK(strings.TrimSpace(string(s.text))); t != nil {
			return t
		}
	}
	return &origTransition{next: s.under}
}

func (s *inputBoxScreen) cancel() *origTransition {
	if s.onCancel != nil {
		if t := s.onCancel(); t != nil {
			return t
		}
	}
	return &origTransition{next: s.under}
}

func (s *inputBoxScreen) update(in shell.InputState) *origTransition {
	s.tick++
	bx, by, bw, bh := inboxOKRect(s.x, s.y)
	s.hoverOK = hitBox(in.MouseX, in.MouseY, bx, by, bw, bh)

	if !s.scriptOK {
		s.typeRunes(ebiten.AppendInputChars(nil))
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			s.backspace()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
			inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter) {
			return s.accept()
		}
		// ESC = 取消,不是離開遊戲——同 esc-cancel/f10-quit 的一貫語意。
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return s.cancel()
		}
	}
	if !in.ClickReleased {
		return nil
	}
	if s.hoverOK {
		if clickSound != nil {
			clickSound()
		}
		return s.accept()
	}
	// 框外點擊什麼都不做——modal 的重點就是它擋住下層(同 confirmbox)。
	return nil
}

func (s *inputBoxScreen) draw(dst *ebiten.Image) {
	if s.under != nil {
		s.under.draw(dst)
	}
	blit := func(im *ebiten.Image, x, y int) {
		if im == nil {
			return
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		drawPanelImage(dst, im, op)
	}
	blit(s.box, s.x, s.y)
	okImg := s.ok
	if s.hoverOK && s.okHi != nil {
		okImg = s.okHi // 第 1 幀是 hover 高亮,同 confirmbox 的兩顆鈕
	}
	bx, by, bw, _ := inboxOKRect(s.x, s.y)
	blit(okImg, bx, by)

	if s.b == nil || s.b.fnt == nil {
		return
	}
	// 標題:水平置中於彈窗寬、垂直置中於 54 px 的標題帶(原版就是這樣算的)。
	const titleSize = 15
	inboxTitleTextRect(s.x, s.y).drawCentered(dst, s.b.fnt, uiText(s.b.lang, s.titleKey), titleSize,
		color.RGBA{240, 220, 120, 255})

	fx, fy, fw, fh := inboxFieldRect(s.x, s.y)
	fillPanel(dst, float32(fx), float32(fy), float32(fw), float32(fh),
		color.RGBA{12, 14, 20, 255}, false)
	const textSize = 14
	inputRect := inboxInputTextRect(fx, fy, fw, fh)
	txt := inputRect.clipped(s.b.fnt, string(s.text), textSize)
	// 游標:半個週期顯示。畫在字尾——量字寬要走字型層,這裡用 DrawCentered 的
	// 對稱作法拿不到寬度,所以改用「整串字 + 一個底線」的畫法。可見時必須只畫
	// 截短後的那一串；若先畫完整文字再疊縮短文字，被切掉的最後字不會自己消失。
	if (s.tick/inboxCaretPeriod)%2 == 0 {
		txt = inboxVisibleInputText(s.b.fnt, string(s.text), textSize, inputRect)
	}
	inputRect.drawLeft(dst, s.b.fnt, txt, textSize, color.RGBA{225, 230, 244, 255})

	// OK 鈕的中文字要先把烘死的 "ACCEPT" 擦掉再疊——不擦就是兩層字疊在一起
	// (同 confirmbox 的 YES/NO,`confirmBtnFace` 那支取面色的 helper 通用,這裡直接用)。
	if s.b.lang == i18n.Traditional {
		f := confirmBtnFace(okImg)
		fillPanel(dst, float32(bx+5), float32(by+5),
			float32(bw-10), float32(inboxOKH-10), f, false)
	}
	inboxOKTextRect(s.x, s.y).drawCentered(dst, s.b.fnt, uiText(s.b.lang, "inputbox.button.accept"), 12,
		color.RGBA{28, 28, 24, 255})
	// 提示行放在輸入欄與 OK 鈕之間的 20px 空帶；放在按鈕下方雖仍在 288×151
	// 邏輯框內，實際字墨會壓到 INBOX 美術的下邊框。
	inboxHintTextRect(s.x, s.y).drawCentered(dst, s.b.fnt, uiText(s.b.lang, "inputbox.hint.accept_cancel"), 11,
		color.RGBA{150, 162, 185, 255})
}
