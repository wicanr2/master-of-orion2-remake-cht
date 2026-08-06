package main

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// multiplayer.go:MULTI-PLAYER GAME SET UP(原版 `Multi_Player_Screen_` @ 0xF4D99)。
//
// remake 先前主選單的 MULTI PLAYER 是死的——有標籤、有熱區,點下去什麼都不會發生。
// 這個檔把它接上,並且**整張版面都取自反組譯**,沒有一個估的座標。
//
// ============ 資產與版面(反組譯 `sub_F42CA` 初始化 + `sub_F009A` 建 widget)============
//
//	背景        MULTIGM.LBX 資產 0(640×480,自帶調色盤)  → `[win+0x0F]`
//	面板        MULTIGM.LBX 資產 1(482×335)              → `[win+0x13]`
//	面板定位    x = (0x280 − 面板寬)/2、y = (0x1E0 − 面板高)/2  → (79, 72)
//	            (`mov edi,280h / sub edi,eax / sar eax,1`,對高度同一段用 0x1E0)
//
//	左欄(連線方式)x = 面板 x + 0x3B(+59):
//	  NETWORK     y +0x5B(+91)   資產 2   → 遊戲模式 2
//	  MODEM       y +0x7A(+122)  資產 3   → 遊戲模式 3
//	  NULL MODEM  y +0x9B(+155)  資產 4   → 遊戲模式 3(另一個子型別)
//	  HOTSEAT     y +0xBB(+187)  資產 5   → 遊戲模式 1  ← 唯一 remake 做得到的
//	（模式碼取自 `Set_Multi_Player_Game_Type_` @ 0xF5691:寫進 `byte_199F3A`,
//	  0=Single Player、1=Hot Seat、2=Network、3=Modem,見 `sub_12175` 組存檔名那段。）
//
//	右欄(動作)x = 面板 x + 0x10D(+269),y 與左欄同四列:
//	  START NEW GAME 資產 7 / LOAD GAME 資產 8 / JOIN GAME 資產 9 / COMM INFO 資產 10
//
//	TEN         (+0x6E, +0xDC),資產 256(253×30)= TOTAL ENTERTAINMENT NETWORK
//	            (1990 年代的線上對戰服務,1999 年已停止營運)
//	CANCEL      (面板 x + 0xB0 = +176, 面板 y + 0x11E = +286),資產 6(129×25)
//
// 按鈕上的英文是烘在圖裡的(資產 2 = "NETWORK"、7 = "START NEW GAME"…,逐張 dump 確認),
// 所以照專案既有做法擦底疊中文。
//
// ============ 原版自己就會隱藏的按鈕(照做,不是偷懶)============
//
//	`sub_F009A` 在選了 HOTSEAT(`[win+0x0FD]==1`)時把 JOIN GAME 的 widget id 直接設成
//	0xFC18(無效值)——**原版在熱座模式下就沒有 JOIN GAME 這顆鈕**。COMM INFO 同理,
//	只有 modem / null modem 才建。remake 照這個條件顯示,不自己發明規則。
//
// ============ 誠實留白 ============
//
//	NETWORK / MODEM / NULL MODEM 在 remake 沒有對應能力(沒有 IPX、沒有數據機)。
//	點下去會說明「本版未實作」,而不是假裝可選再在後面某處失敗。
//	熱座的席位模型見 internal/shell/hotseat.go、交接畫面見 cmd/moo2/hotseat.go。

const (
	mpLBX      = "multigm.lbx"
	mpBGAsset  = 0 // 640×480 全螢幕背景(自帶調色盤)
	mpPanAsset = 1 // 482×335 設定面板

	mpColLeftDX  = 59  // +0x3B
	mpColRightDX = 269 // +0x10D
	mpRow0DY     = 91  // +0x5B
	mpRow1DY     = 122 // +0x7A
	mpRow2DY     = 155 // +0x9B
	mpRow3DY     = 187 // +0xBB
	mpTenDX      = 110 // +0x6E   TOTAL ENTERTAINMENT NETWORK(資產 256,253×30)
	mpTenDY      = 220 // +0xDC
	mpCancelDX   = 176 // +0xB0
	mpCancelDY   = 286 // +0x11E
)

// mpMode 是連線方式(對應原版 `byte_199BF2` 的四個 case)。
type mpMode int

const (
	mpNetwork mpMode = iota
	mpModem
	mpNullModem
	mpHotseat
)

// mpButton 是面板上的一顆鈕:資產、相對面板的位置、中文標籤、動作代號。
type mpButton struct {
	asset  int
	dx, dy int
	zh     string
	act    string
}

// mpButtons 是左欄四個連線方式 + 右欄四個動作 + CANCEL,全部座標見檔頭。
var mpButtons = []mpButton{
	{2, mpColLeftDX, mpRow0DY, "區域網路", "network"},
	{3, mpColLeftDX, mpRow1DY, "數據機", "modem"},
	{4, mpColLeftDX, mpRow2DY, "序列埠直連", "nullmodem"},
	{5, mpColLeftDX, mpRow3DY, "熱座", "hotseat"},
	{7, mpColRightDX, mpRow0DY, "開始新遊戲", "start"},
	{8, mpColRightDX, mpRow1DY, "載入遊戲", "load"},
	{9, mpColRightDX, mpRow2DY, "加入遊戲", "join"},
	{10, mpColRightDX, mpRow3DY, "連線資訊", "comm"},
	{256, mpTenDX, mpTenDY, "TEN 連線服務", "ten"},
	{6, mpCancelDX, mpCancelDY, "取消", "cancel"},
}

// multiplayerScreen 是多人遊戲設定畫面。
type multiplayerScreen struct {
	b    *sceneBuilder
	fnt  *uifont.Font
	mode mpMode
	// humans 是熱座的真人席位數(2..maxHotseatSeats)。原版是在玩家設定階段逐個帝國
	// 標成真人(`Get_Multi_Player_N_Humans_` @ 0x121F0 就是去數 `player[i]` 裡控制碼
	// 為 100 的帝國),remake 沒有那個逐帝國的設定畫面,改成在 HOTSEAT 鈕上循環選人數
	// ——語意相同(把 N 個帝國從 AI 換成真人),操作方式是 remake 自己的。
	humans int
	msg    string

	bg, panel *ebiten.Image
	// imgs[資產] 是該按鈕的三個狀態幀(見 mpFrame*);faces 是各幀的面色(擦底用)。
	imgs       map[int][]*ebiten.Image
	faces      map[int][]color.RGBA
	titleFace  color.RGBA // 標題帶底色(從面板採樣,用來擦掉烘上去的英文標題)
	panX, panY int
	panW, panH int
}

// 原版每顆按鈕的三個狀態幀(逐張 dump 確認:HOTSEAT 的三幀分別是淺面深字 / 深面橘字 /
// 深面灰字)。remake 直接用這三幀,不自己畫高亮框——狀態長什麼樣是原版已經決定好的。
const (
	mpFrameNormal   = 0 // 一般
	mpFrameSelected = 1 // 選中 / 游標懸停
	mpFrameDisabled = 2 // 停用
)

// 三種狀態各自的文字顏色,量自對應幀烘上去的英文(逐幀取中央帶的顏色分布)。
var (
	mpLabelNormal   = color.RGBA{16, 12, 12, 255}    // 淺面深字
	mpLabelSelected = color.RGBA{252, 136, 0, 255}   // 深面橘字
	mpLabelDisabled = color.RGBA{152, 148, 140, 255} // 深面灰字
	// ⚠ 停用態的字色**刻意偏離原版**:原版停用幀的字也是 (16,12,12),壓在 (72,68,64) 的
	// 深面上幾乎看不見(那是「刻進去」的視覺)。英文短、輪廓好認還讀得出來,中文筆劃多,
	// 照抄就是一團黑。這裡提亮到能讀,但仍比一般態暗——玩家要看得出「這顆不能用」,
	// 也要看得出它是什麼。
)

// maxHotseatSeats 是這個 UI 能選到的最大真人數:一位玩家 + 每個 AI 對手各一席
// (`SetupHotseat` 就是把 AI 帝國換成真人,見 internal/shell/hotseat.go)。
const maxHotseatSeats = 1 + shell.DefaultOpponents

func newMultiplayerScreen(b *sceneBuilder) *multiplayerScreen {
	s := &multiplayerScreen{b: b, fnt: b.fnt, mode: mpHotseat, humans: 2,
		imgs: map[int][]*ebiten.Image{}, faces: map[int][]color.RGBA{}}
	// 背景自帶調色盤,面板與各按鈕都借它上色。
	prov, err := decodeAsset(b.res, mpLBX, mpBGAsset)
	if err != nil || prov.Embedded == nil || len(prov.Frames) == 0 {
		return s
	}
	s.bg = ebiten.NewImageFromImage(prov.Frames[0].ToRGBA(prov.Embedded, false))
	s.titleFace = color.RGBA{58, 60, 56, 255} // 取不到面板時的保守深灰
	if im, err := decodeAsset(b.res, mpLBX, mpPanAsset); err == nil && len(im.Frames) > 0 {
		rgba := im.Frames[0].ToRGBA(prov.Embedded, false)
		s.panel = ebiten.NewImageFromImage(rgba)
		bb := rgba.Bounds()
		s.panW, s.panH = bb.Dx(), bb.Dy()
		// 標題帶底色:切出標題列那一條(y 16..40、左右各留 30px)再取眾數,理由同 dominantFace。
		if s.panW > 80 && s.panH > 40 {
			strip := rgba.SubImage(image.Rect(bb.Min.X+30, bb.Min.Y+16, bb.Max.X-30, bb.Min.Y+40))
			if sub, ok := strip.(*image.RGBA); ok {
				s.titleFace = dominantFace(sub)
			}
		}
	}
	// 面板置中(原版對面板尺寸算 (640−w)/2、(480−h)/2)。
	if s.panW == 0 {
		s.panW, s.panH = 482, 335 // 取不到面板時的退路尺寸(實測 multigm.lbx#1)
	}
	s.panX, s.panY = (moo2ScreenW-s.panW)/2, (moo2ScreenH-s.panH)/2
	for _, btn := range mpButtons {
		im, err := decodeAsset(b.res, mpLBX, btn.asset)
		if err != nil || len(im.Frames) == 0 {
			continue
		}
		for _, f := range im.Frames {
			rgba := f.ToRGBA(prov.Embedded, true)
			s.imgs[btn.asset] = append(s.imgs[btn.asset], ebiten.NewImageFromImage(rgba))
			s.faces[btn.asset] = append(s.faces[btn.asset], dominantFace(rgba))
		}
	}
	return s
}

// dominantFace 取一張按鈕圖的「面色」——**整個內部**(四邊各扣 3px 浮雕邊框)出現最多次的顏色。
//
// 兩個都踩過的坑:
//   ① 不能像其他視窗那樣「取左緣往內第 4 像素」——那個位置在這批按鈕上踩到的是**浮雕邊框**
//      不是面。停用幀的面是深灰 (72,68,64) 而邊框是淺灰 (116,112,104),取錯就會拿淺色去擦
//      深色的按鈕面,得到一塊突兀的亮斑。
//   ② 也不能只取「中央橫帶」——那正好是烘上去的英文所在。字長一點(START NEW GAME)時
//      文字像素會反過來壓過面色像素,眾數變成字色 (16,12,12),於是用黑色擦掉整顆鈕、
//      再畫上黑字,按鈕看起來就是一片空白。
// 取整個內部就沒有這個問題:面積夠大,文字永遠是少數。
func dominantFace(img *image.RGBA) color.RGBA {
	b := img.Bounds()
	if b.Dx() <= 8 || b.Dy() <= 6 {
		return color.RGBA{60, 64, 72, 255}
	}
	inner := image.Rect(b.Min.X+3, b.Min.Y+3, b.Max.X-3, b.Max.Y-3)
	count := map[color.RGBA]int{}
	for y := inner.Min.Y; y < inner.Max.Y; y++ {
		for x := inner.Min.X; x < inner.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			if a == 0 {
				continue
			}
			count[color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), 255}]++
		}
	}
	// 第二趟照掃描順序挑出票數最高的那個色。**不能直接 range map**——Go 的 map 迭代順序
	// 隨機,同票時每次跑會選到不同的顏色,截圖廊就不可重現了。
	best, bestN := color.RGBA{60, 64, 72, 255}, 0
	for y := inner.Min.Y; y < inner.Max.Y; y++ {
		for x := inner.Min.X; x < inner.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			if a == 0 {
				continue
			}
			c := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), 255}
			if n := count[c]; n > bestN {
				best, bestN = c, n
			}
		}
	}
	return best
}

// frameFor 決定某顆鈕現在該用哪一個狀態幀。
func (s *multiplayerScreen) frameFor(act string) int {
	if act == "hotseat" && s.mode == mpHotseat {
		return mpFrameSelected
	}
	if !s.enabled(act) || !s.implemented(act) {
		return mpFrameDisabled
	}
	return mpFrameNormal
}

// frameImage 取某顆鈕的某一幀;該幀不存在時退回第 0 幀,再不行回 nil。
func (s *multiplayerScreen) frameImage(asset, frame int) *ebiten.Image {
	fs := s.imgs[asset]
	if len(fs) == 0 {
		return nil
	}
	if frame >= len(fs) {
		frame = 0
	}
	return fs[frame]
}

// frameFace 取某顆鈕某一幀的面色(擦底用)。
func (s *multiplayerScreen) frameFace(asset, frame int) (color.RGBA, bool) {
	fs := s.faces[asset]
	if len(fs) == 0 {
		return color.RGBA{}, false
	}
	if frame >= len(fs) {
		frame = 0
	}
	return fs[frame], true
}

// btnRect 回傳某顆鈕的螢幕座標。尺寸取自資產本身(原版也是拿圖的 w/h 當熱區,
// `mov bx,[edx+2] / mov dx,[edx]` 就是讀影像標頭的寬高)。
func (s *multiplayerScreen) btnRect(btn mpButton) (int, int, int, int) {
	w, h := 154, 26 // 取不到圖時沿用左欄按鈕實測尺寸
	if im := s.frameImage(btn.asset, mpFrameNormal); im != nil {
		w, h = im.Bounds().Dx(), im.Bounds().Dy()
	}
	return s.panX + btn.dx, s.panY + btn.dy, w, h
}

// enabled 回傳這顆鈕在目前模式下能不能點。
//
// ⚠ 「不能點」不等於「不畫」:原版的面板美術(資產 1)本來就把八顆鈕都畫死在圖上,
// `sub_F009A` 只是**不建 widget**(把 id 設成 0FC18h)。所以熱座模式下 JOIN GAME
// 那顆在原版畫面上仍然看得到,只是點不動。remake 照這個行為做。
func (s *multiplayerScreen) enabled(act string) bool {
	switch act {
	case "join":
		return s.mode != mpHotseat // 熱座模式:原版不建這顆鈕的 widget
	case "comm":
		return s.mode == mpModem || s.mode == mpNullModem
	case "network", "modem", "nullmodem", "ten":
		return true // 可以點,點了會說明本版沒有網路層
	}
	return true
}

// implemented 回傳這個動作在 remake 有沒有真的實作(沒有的畫成灰的)。
func (s *multiplayerScreen) implemented(act string) bool {
	switch act {
	case "network", "modem", "nullmodem", "join", "comm", "ten":
		return false
	}
	return true
}

func (s *multiplayerScreen) update(in shell.InputState) *origTransition {
	if !in.ClickReleased {
		return nil
	}
	for _, btn := range mpButtons {
		if !s.enabled(btn.act) {
			continue
		}
		x, y, w, h := s.btnRect(btn)
		if !hitRect(in, x, y, w, h) {
			continue
		}
		return s.click(btn.act)
	}
	return nil
}

func (s *multiplayerScreen) click(act string) *origTransition {
	switch act {
	case "network", "modem", "nullmodem":
		// 原版這三個真的能連線;remake 沒有網路層,說清楚而不是假裝可選。
		s.msg = "本版只實作熱座(同機輪流)。連線對戰需要原版的 IPX / 數據機層。"
		return nil
	case "hotseat":
		if s.mode != mpHotseat {
			s.mode, s.msg = mpHotseat, ""
			return nil
		}
		// 已經選中 → 循環真人席位數。
		s.humans++
		if s.humans > maxHotseatSeats {
			s.humans = 2
		}
		return nil
	case "start":
		if s.mode != mpHotseat {
			s.msg = "請先選「熱座」——其餘連線方式本版未實作。"
			return nil
		}
		s.b.pendingHotseat = s.humans
		return s.b.goTo(s.b.newGameSetup, "新遊戲設定")
	case "load":
		if !shell.AnySaveExists(saveDirFor()) {
			s.msg = "還沒有任何存檔。"
			return nil
		}
		sc, err := s.b.loadGame()
		if err != nil {
			s.msg = "存檔視窗開不起來。"
			return nil
		}
		return &origTransition{next: sc}
	case "join", "comm":
		s.msg = "本版只實作熱座(同機輪流)。"
		return nil
	case "ten":
		// TEN(Total Entertainment Network)是 1990 年代的線上對戰服務,1999 年就收了。
		s.msg = "TEN 是原版年代的線上對戰服務,早已停止營運。"
		return nil
	}
	return s.b.goTo(s.b.menu, "主選單")
}

func (s *multiplayerScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{6, 8, 14, 255})
	if s.bg != nil {
		dst.DrawImage(s.bg, &ebiten.DrawImageOptions{})
	}
	if s.panel != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(s.panX), float64(s.panY))
		dst.DrawImage(s.panel, op)
	}
	if s.fnt == nil {
		return
	}
	sel := color.RGBA{250, 226, 130, 255}
	dim := color.RGBA{128, 132, 142, 255}

	for _, btn := range mpButtons {
		// 全部都畫:原版的面板美術本來就把八顆鈕都烘在圖上,不能點的只是沒有 widget
		// (見 enabled 的說明)。狀態幀由 frameFor 決定,顏色跟著幀走。
		x, y, w, h := s.btnRect(btn)
		frame := s.frameFor(btn.act)
		if im := s.frameImage(btn.asset, frame); im != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			dst.DrawImage(im, op)
		}
		// 擦掉烘在圖上的英文再疊中文(上下左右各留 3px 保住浮雕邊框)。
		if face, ok := s.frameFace(btn.asset, frame); ok {
			vector.DrawFilledRect(dst, float32(x+3), float32(y+3), float32(w-6), float32(h-6), face, false)
		}
		label := btn.zh
		col := mpLabelNormal
		switch frame {
		case mpFrameSelected:
			col = mpLabelSelected
		case mpFrameDisabled:
			col = mpLabelDisabled
		}
		if btn.act == "hotseat" && s.mode == mpHotseat {
			label = fmt.Sprintf("熱座 %d 人", s.humans)
		}
		s.fnt.DrawCentered(dst, label, float64(x+w/2), float64(y+h/2), 13, col)
	}

	// 面板標題帶(MULTI-PLAYER GAME SET UP)烘在面板上,同樣擦底疊中文。
	// 底色從標題帶自身採樣(左緣往內 8px),不用猜的常數——面板調色盤換了也不會露餡。
	tx, ty, tw, th := s.panX+30, s.panY+16, s.panW-60, 24
	vector.DrawFilledRect(dst, float32(tx), float32(ty), float32(tw), float32(th), s.titleFace, false)
	s.fnt.DrawCentered(dst, "多人遊戲設定", float64(s.panX+s.panW/2), float64(ty+th/2), 16, sel)

	note := fmt.Sprintf("熱座:%d 位真人輪流下令,其餘帝國仍由 AI 操作。", s.humans)
	if s.mode != mpHotseat {
		note = "本版只實作熱座。"
	}
	s.fnt.DrawCentered(dst, note, 320, float64(s.panY+s.panH+16), 12, dim)
	if s.msg != "" {
		s.fnt.DrawCentered(dst, s.msg, 320, float64(s.panY+s.panH+36), 12,
			color.RGBA{235, 160, 120, 255})
	}
}

// multiPlayer 從主選單進入多人遊戲設定畫面。
func (b *sceneBuilder) multiPlayer() (origScreen, error) { return newMultiplayerScreen(b), nil }
