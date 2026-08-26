package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/netplay"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// choosemultinetgame.go:**選擇要加入哪一場區網對局**
//(原版 `Choose_Multi_Network_Game_Screen_` @ 0xF0C8E,版面在
// `Load_Choose_Multi_Net_Game_Screen_` @ 0xF40D3 + `Add_Choose_Multi_Net_Game_Fields_`
// @ 0xEFF87,繪製在 `Draw_Choose_Multi_Net_Game_Screen_` @ 0xF1AF4)。
//
// ============ 這張畫面的資料從哪來,原版沒有回答 ============
//
// 原版走 IPX,而 IPX **自帶**廣播式的服務公告——「列出區網上有哪些對局」是協定給的能力,
// 不是遊戲自己做的。remake 走 TCP,TCP 沒有這個能力:照抄畫面而不補那一層,
// 會得到一張永遠空的清單。所以先補 `internal/netplay/discovery.go`(UDP 廣播),
// 這一檔才有東西可列。**那一層是移植決策,不是還原**,標在該檔檔頭。
//
// ============ 版面 ============
//
// `Load_Choose_Multi_Net_Game_Screen_`:
//
//	[win+0x0F] = MULTIGM 資產 0        ; 背景
//	[win+0x1B] = MULTIGM 資產 0x29=41  ; 主面板(479×384)
//	[win+0x73] = MULTIGM 字型 0x101
//	x = (0x280 − 資產41.寬) / 2                    = (640−479)/2 = 80
//	y = ((0x1E0 − 資產41.高) − 0x51) / 2 + 0x25    = ((480−384)−81)/2 + 37 = 44
//
// ⚠ 那個 `−0x51`(81)剛好是標題帶(資產 27)的高,但這張畫面**沒有畫標題帶**
//(`Draw_Choose_Multi_Net_Game_Screen_` 只 blit 資產 41)。也就是它是**版面上的讓位**,
// 不是「上面還有一塊」。照抄數字,不照抄我對數字的解讀。
//
// `Add_Choose_Multi_Net_Game_Fields_` 給每一列的點擊區(`sub_11438B` = `Add_Hidden_Field_`,
// 隱形熱區,美術已經畫好格子了):
//
//	x1 = winX + 0x26  (+38)     x2 = winX + 0x190 (+400)
//	y1 = winY + edi,edi 起始 0x40 (64),每列 += 0x1B (27)
//	y2 = y1 + 0x16    (+22)
//
// 也就是每列 **362×22**,列距 27,最多 **10** 場(`[win + i*2 + 0xA7]` 那個陣列的長度,
// 初值 0xFC18)。底下另有一顆按鈕(`sub_1151B0` = `Add_Button_Field_`)在
// (winX + 0xBF, winY + 0x158) = (+191, +344)。
//
// `Draw_Choose_Multi_Net_Game_Screen_` 再給文字的擺法:
//
//	文字 x = winX + 0x26 + 9 = winX + 0x2F (+47)
//	文字 y = winY + var_C + (0x16 − 字高)/2      ; var_C 起始 0x43 (67),每列 += 0x1B
//
// 也就是**字在 22 px 的列裡垂直置中**——不是靠上也不是靠下。67 = 64 + 3,
// 與熱區的起點差 3 px。
//
// 被選中的那一列另有一個**脈動亮度**:`[win+0x1E8]` 在 −3 與 +4 之間來回、
// `[win+0x1E9]` 是方向(±1),而配色從 (0x95,0x97,0x91) 換成 (0x97,0x99,0x91)
// ——選中就是整組色往上推兩階。remake 用一條亮邊 + 較亮的字色表達同一件事。
//
// ============ 誠實留白 ============
//
// 原版可以在這張畫面**改對局名稱**(`Change_MP_Game_Name_` @ 0xF5777,長度上限 8、
// 且要與既有對局不同名)。remake 還沒有文字輸入框,所以名稱目前取玩家名的前 8 字元。
// 上限與唯一性的規則已經記在 `netplay.GameNameMax`,做輸入框時直接套。

const (
	cmngPanelAsset = 41 // 479×384
	cmngPanelW     = 479
	cmngPanelH     = 384

	// 版面讓位值:`((0x1E0 − 高) − 0x51)/2 + 0x25` 裡的那兩個數。
	cmngYBias   = 0x51 // 81
	cmngYOffset = 0x25 // 37

	// 每列熱區(相對視窗左上,取自 Add_Choose_Multi_Net_Game_Fields_)。
	cmngRowX1    = 0x26  // +38
	cmngRowX2    = 0x190 // +400
	cmngRowFirst = 0x40  // +64
	cmngRowStep  = 0x1B  // 27
	cmngRowH     = 0x16  // 22

	// 文字:x 再 +9,y 在列內垂直置中(起點比熱區低 3 px)。
	cmngTextDX     = 9
	cmngTextFirstY = 0x43 // 67

	cmngMaxRows = 10 // [win + i*2 + 0xA7] 那個陣列的長度

	// 底下那顆鈕(Add_Button_Field_)。
	cmngBtnX = 0xBF  // +191
	cmngBtnY = 0x158 // +344
	// 按鈕尺寸量自資產 41(反組譯沒有給,標成量的)。
	cmngBtnW, cmngBtnH = 86, 22

	// 「直接輸入位址」的熱區。**原版沒有這顆鈕**——擺在最後一列下方的空白帶,
	// 不佔用任何一個原版座標。
	cmngDirectDX = 0x26
	cmngDirectDY = 0x40 + 10*0x1B + 8 // 最後一列之後
	cmngDirectW  = 130
	cmngDirectH  = 20
)

// cmngWindow 回傳主面板左上角(見檔頭的算式)。
func cmngWindow() (x, y int) {
	return (moo2ScreenW - cmngPanelW) / 2,
		((moo2ScreenH-cmngPanelH)-cmngYBias)/2 + cmngYOffset
}

// cmngDirectRect 回傳「直接輸入位址」的熱區(remake 自加,不是原版座標)。
func cmngDirectRect(winX, winY int) (x, y, w, h int) {
	return winX + cmngDirectDX, winY + cmngDirectDY, cmngDirectW, cmngDirectH
}

// cmngRowRect 回傳第 i 列的點擊區(螢幕座標)。
func cmngRowRect(winX, winY, i int) (x, y, w, h int) {
	x1 := winX + cmngRowX1
	y1 := winY + cmngRowFirst + i*cmngRowStep
	return x1, y1, (winX + cmngRowX2) - x1, cmngRowH
}

// cmngTextTop 回傳第 i 列文字的**上緣** y(字在 22 px 的列裡垂直置中)。
//
// glyphH 是實際字高;原版拿 `sub_122259` 問字型要,remake 傳進來。
//
// ⚠ 回傳的是上緣不是基線:`uifont.Draw` 底層是 ebiten text/v2,`GeoM.Translate(x,y)`
// 的 y 是**行框上緣**(v1 才是基線)。第一版多加了一個字高當基線,結果整欄字掉到下一列
// ——截圖上選取框在第一列、字在第二列,一眼就看得出來。
// 原版的算式本來就是上緣(`var_14 + (0x16 − 字高)/2`),多加那一下是自己加的。
func cmngTextTop(winY, i, glyphH int) int {
	return winY + cmngTextFirstY + i*cmngRowStep + (cmngRowH-glyphH)/2
}

// chooseMultiNetGameScreen 是對局清單畫面。
type chooseMultiNetGameScreen struct {
	b       *sceneBuilder
	browser *netplay.Browser
	games   []netplay.Game
	sel     int
	msg     string

	bg, panel *ebiten.Image
	tick      int
}

func (b *sceneBuilder) chooseMultiNetGame(browser *netplay.Browser) *chooseMultiNetGameScreen {
	s := &chooseMultiNetGameScreen{b: b, browser: browser, sel: -1}
	s.bg = b.multigmImage(mpBGAsset, false)
	s.panel = b.multigmImage(cmngPanelAsset, true)
	return s
}

// chooseMultiNetGameDemo 是截圖廊用的清單(不開 socket,同 chooseNetPlayersDemo)。
func (b *sceneBuilder) chooseMultiNetGameDemo() *chooseMultiNetGameScreen {
	s := b.chooseMultiNetGame(nil)
	s.games = []netplay.Game{
		{Name: "ORION", Addr: "192.168.1.20:24501", Players: 2, Max: 8},
		{Name: "SAKKRA", Addr: "192.168.1.31:24501", Players: 1, Max: 4},
		{Name: "ANTARES", Addr: "192.168.1.44:24501", Players: 5, Max: 6},
	}
	s.sel = 0
	return s
}

func (s *chooseMultiNetGameScreen) update(in shell.InputState) *origTransition {
	s.tick++
	if s.browser != nil {
		s.games = s.browser.Games()
	}
	if !in.ClickReleased {
		return nil
	}
	winX, winY := cmngWindow()
	for i := 0; i < len(s.games) && i < cmngMaxRows; i++ {
		x, y, w, h := cmngRowRect(winX, winY, i)
		if in.MouseX >= x && in.MouseX < x+w && in.MouseY >= y && in.MouseY < y+h {
			s.sel = i
			if sc, err := s.b.joinNetGame(s.games[i]); err == nil {
				s.closeBrowser()
				return &origTransition{next: sc}
			}
			s.msg = s.b.tr("連不上這場對局(主機可能已關)", "Could not connect (host may be gone)")
			return nil
		}
	}
	dx, dy, dw, dh := cmngDirectRect(winX, winY)
	if in.MouseX >= dx && in.MouseX < dx+dw && in.MouseY >= dy && in.MouseY < dy+dh {
		return &origTransition{next: s.b.inputBox(s,
			"inputbox.title.host_address", netLobbyDialAddr, 45,
			func(addr string) *origTransition {
				if addr == "" {
					return nil
				}
				sc, err := s.b.joinNetGame(netplay.Game{Name: "direct", Addr: addr})
				if err != nil {
					s.msg = s.b.tr("連不上 ", "Could not connect to ") + addr
					return nil
				}
				s.closeBrowser()
				return &origTransition{next: sc}
			})}
	}
	// 點清單以外的地方 = 離開。
	s.closeBrowser()
	sc, err := s.b.multiPlayer()
	if err != nil {
		return nil
	}
	return &origTransition{next: sc}
}

// closeBrowser 收掉背景的 UDP 監聽——離開畫面還開著 socket 是洩漏,
// 而且下一次進來會開第二個聽同一個埠的 socket 而失敗。
func (s *chooseMultiNetGameScreen) closeBrowser() {
	if s.browser != nil {
		s.browser.Close()
		s.browser = nil
		s.b.netBrowser = nil
	}
}

func (s *chooseMultiNetGameScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{6, 8, 14, 255})
	blit := func(im *ebiten.Image, x, y int) {
		if im == nil {
			return
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		drawPanelImage(dst, im, op)
	}
	blit(s.bg, 0, 0)
	winX, winY := cmngWindow()
	blit(s.panel, winX, winY)

	if s.b.fnt == nil {
		return
	}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{205, 212, 228, 255}
	hot := color.RGBA{245, 238, 200, 255}
	dim := color.RGBA{140, 150, 172, 255}

	// 標題:原版烘的是 "JOIN NETWORK GAME SETUP",中文要擦底疊上去。
	fillPanel(dst, float32(winX+86), float32(winY+26), 308, 26,
		color.RGBA{24, 27, 34, 255}, false)
	s.b.fnt.DrawCentered(dst, truncateToWidth(s.b.fnt, s.b.tr("選擇要加入的對局", "Choose a network game"), 16, 380),
		float64(winX+cmngPanelW/2), float64(winY)+30, 16, gold)

	const glyphH = 13
	for i := 0; i < cmngMaxRows && i < len(s.games); i++ {
		g := s.games[i]
		x, y, w, h := cmngRowRect(winX, winY, i)
		col := body
		if i == s.sel {
			// 原版的「選中」是整組配色往上推兩階 + 脈動亮度;這裡用一條亮邊表達。
			vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1, gold, false)
			col = hot
		}
		s.b.fnt.Draw(dst, truncateToWidth(s.b.fnt, g.Name, glyphH, 110), float64(x+cmngTextDX),
			float64(cmngTextTop(winY, i, glyphH)), glyphH, col)
		s.b.fnt.Draw(dst, truncateToWidth(s.b.fnt, g.Addr, 11, float64(w-cmngTextDX-162)), float64(x+cmngTextDX+120),
			float64(cmngTextTop(winY, i, 11)), 11, dim)
		s.b.fnt.Draw(dst, fmt.Sprintf("%d/%d", g.Players, g.Max),
			float64(x+w-42), float64(cmngTextTop(winY, i, 11)), 11, dim)
	}

	if len(s.games) == 0 {
		s.b.fnt.DrawCentered(dst, truncateToWidth(s.b.fnt, s.b.tr("區網上沒有偵測到對局", "No games found on the LAN"), 13, 380),
			float64(winX+cmngPanelW/2), float64(winY+cmngRowFirst)+40, 13, dim)
		s.b.fnt.DrawCentered(dst,
			truncateToWidth(s.b.fnt, s.b.tr("主機端要先在多人設定按「開始新遊戲」", "The host must start a game first"), 11, 380),
			float64(winX+cmngPanelW/2), float64(winY+cmngRowFirst)+62, 11, dim)
	}

	// 「直接輸入位址」:原版**沒有**這顆鈕(IPX 自己找得到,不需要打位址)。
	// remake 的 UDP 廣播只跨得過同一個區網,跨網段就得能打位址——這是補的,不是還原,
	// 所以擺在清單外面而不是塞進原版的版面裡。
	dx, dy, dw, dh := cmngDirectRect(winX, winY)
	vector.StrokeRect(dst, float32(dx), float32(dy), float32(dw), float32(dh), 1, dim, false)
	s.b.fnt.DrawCentered(dst, truncateToWidth(s.b.fnt, s.b.tr("直接輸入位址", "Enter address"), 12, float64(dw-10)),
		float64(dx+dw/2), float64(dy)+5, 12, body)

	// 底下那顆鈕的位置是反組譯真值(欄位左上角);寬高是量的,原版沒有給。
	bx, by := winX+cmngBtnX, winY+cmngBtnY
	fillPanel(dst, float32(bx-2), float32(by-4),
		float32(cmngBtnW), float32(cmngBtnH), color.RGBA{150, 148, 138, 255}, false)
	s.b.fnt.DrawCentered(dst, truncateToWidth(s.b.fnt, s.b.tr("返回", "CANCEL"), 13, float64(cmngBtnW-10)),
		float64(bx-2+cmngBtnW/2), float64(by)-1, 13, color.RGBA{28, 28, 24, 255})
	if s.msg != "" {
		s.b.fnt.DrawCentered(dst, truncateToWidth(s.b.fnt, s.msg, 12, float64(cmngPanelW-24)), float64(winX+cmngPanelW/2),
			float64(by)-26, 12, color.RGBA{240, 170, 140, 255})
	}
}
