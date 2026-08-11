package main

import (
	"fmt"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// loadgame.go:載入 / 儲存遊戲視窗(原版 `Mainmenu_Load_Game_Popup_` @ 0x804B7、
// `Do_Save_Game_Popup_` @ 0x7E154)。
//
// **兩者是同一個視窗**:原版的繪製函式只有一個,叫 `_Draw_Load_Save_Game_Popup_` @ 0x7F206
// ——名字自己就把這件事說完了。差別只在說明列(`Set_Load_Game_Screen_Help_List_` @ 0x6F850
// vs `Set_Save_Game_Screen_Help_List_` @ 0x6F865)與動作鈕的字。remake 照這個結構做:
// 同一個 `loadGameScreen`,`mode` 決定是讀還是存。
//
// remake 先前根本沒有這個畫面:主選單的 Continue 與 Load Game 都是「讀那唯一一個存檔檔案」,
// 沒有存檔時點下去靜默無反應。2026-07-12 用 archive.org 線上原版對照時,使用者回報的
// issue #2 就是這件事——原版**無存檔時 Continue / Load Game 是灰階不可按的**,有存檔時
// Load Game 會開一個十格的存檔選擇視窗。
//
// ============ 版面:openorion2 `LoadGameWindow::initWidgets`(它自己是從原版 RE 出來的)============
//
//	背景        game.lbx 資產 20,調色盤取自 mainmenu.lbx 資產 21
//	視窗定位    _x = (640 − 寬)/2、_y = (480 − 高)/2   ← 置中,不是固定座標
//	LOAD 鈕     (37, 337),圖 game.lbx 資產 21
//	CANCEL 鈕   (171, 338),68×22,圖 game.lbx 資產 22
//	存檔槽      10 格,第 i 格 (22, 22 + 31×i),232×27
//	槽內第一行  存檔名 (_x + 32, _y + 24 + 31×i)
//	槽內第二行  星曆 (_x + 32, +14)、存檔時間 (_x + 122, +14),皆用更小一級的字
//	對局類型    圖示 (_x + 206, +12)
//	最後一格    固定是自動存檔(`if (slot == SAVEGAME_SLOTS - 1)` → ESTR_SAVESLOT_AUTO)
//
// 上面除了「置中」是公式,其餘都是硬編立即數,一個都沒有估。槽位規格(10 格、末格自動存檔)
// 見 internal/shell/saveslots.go 檔頭。
//
// 對局類型圖示(2026-08-07 補):原版右側依存檔類型畫單人/熱座/網路/數據機四種
// (`ASSET_LOAD_SINGLE`=23 / `HOTSEAT`=24 / `NETWORK`=25 / `MODEM`=26,位置
// `_x + 206, y + 12`)。熱座落地後 remake 有單人與熱座兩種對局,兩種圖示都畫得出來,
// 故接上;網路/數據機 remake 沒有,那兩張永遠不會被選到。
//
// ⚠ 誠實留白:remake 存檔用自己的 JSON 格式,不是原版 `.GAM`
// (原版格式由 `internal/save` **唯讀**解析,寫回不在範圍內)。

// 原版載入視窗的資產與座標(openorion2 mainmenu.cpp)。
const (
	loadWinLBX      = "game.lbx"
	loadWinBGAsset  = 20 // ASSET_LOAD_BACKGROUND
	loadWinBtnAsset = 21 // ASSET_LOAD_LOADBUTTON
	loadWinCanAsset = 22 // ASSET_LOAD_CANCEL
	// 上面三個索引與反組譯對得上:`Load_Mainmenu_Load_Game_Popup_` @ 0x803D9 依序載
	// GAME.LBX 資產 0x14/0x15/0x16/0x17/0x18/0x19/0x1A(20–26),正是 openorion2 的
	// ASSET_LOAD_BACKGROUND…MODEM 那一串。兩個獨立來源互相印證。

	loadWinPalLBX   = "mainmenu.lbx"
	loadWinPalAsset = 21 // ASSET_MENU_BACKGROUND(主選單背景提供調色盤)

	loadBtnX, loadBtnY       = 37, 337
	loadCanX, loadCanY       = 171, 338
	loadCanW, loadCanH       = 68, 22
	loadSlotX, loadSlotY0    = 22, 22
	loadSlotW, loadSlotH     = 232, 27
	loadSlotStep             = 31
	loadSlotTextX, loadTextY = 32, 24 // 槽內第一行(存檔名)相對視窗左上的偏移

	// 第二行:星曆在名稱正下方 14px,存檔時間在 x+122(openorion2 drawSlot:
	// `smallfnt->renderText(_x + 32, y + 14, …)` 與 `(_x + 122, y + 14, …)`)。
	loadSlotSubDY   = 14
	loadSlotTimeX   = 122
	loadSlotIconX   = 206 // 對局類型圖示 `_singleIcon->draw(_x + 206, y + 12)`
	loadSlotIconDY  = 12
	loadWinIconSing = 23 // ASSET_LOAD_SINGLE
	loadWinIconHot  = 24 // ASSET_LOAD_HOTSEAT
)

// saveLoadMode 決定這個視窗是「讀」還是「存」(原版共用同一個視窗,見檔頭)。
type saveLoadMode int

const (
	modeLoad saveLoadMode = iota
	modeSave
)

// loadGameScreen 是載入 / 儲存遊戲視窗。
type loadGameScreen struct {
	b    *sceneBuilder
	fnt  *uifont.Font
	mode saveLoadMode
	// back 是關閉視窗後要回哪裡——從主選單開的回主選單,從遊戲選單開的回星系主畫面。
	back     func() (*overlayScreen, error)
	backName string

	bg, loadBtn, cancelBtn *ebiten.Image
	// singleIcon / hotseatIcon 是槽右側的對局類型圖示(見檔頭)。
	singleIcon, hotseatIcon *ebiten.Image
	// loadFace / cancelFace 是兩顆鈕的「面色」,用來擦掉烘在圖上的英文(LOAD / CANCEL)
	// 再疊中文——與 overlayScreen 的擦底疊字同一套手法,只是這裡是自繪畫面,得自己採樣。
	loadFace, cancelFace   color.RGBA
	winX, winY, winW, winH int

	slots    []shell.SaveSlotInfo
	selected int // -1 = 未選
	msg      string
}

// loadWinImage 取 game.lbx 的某資產(調色盤借主選單背景,見檔頭)。
// 一併回傳「面色」:從按鈕左緣往內數第 4 個像素、垂直置中處採樣,那裡是純按鈕面、
// 不會踩到烘上去的英文,拿來當擦底色。失敗回 nil + 零值。
func loadWinImage(res *assets.Resolver, assetID int, keyColor bool) (*ebiten.Image, color.RGBA) {
	prov, err := decodeAsset(res, loadWinPalLBX, loadWinPalAsset)
	if err != nil || prov.Embedded == nil {
		return nil, color.RGBA{}
	}
	im, err := decodeAsset(res, loadWinLBX, assetID)
	if err != nil || len(im.Frames) == 0 {
		return nil, color.RGBA{}
	}
	rgba := im.Frames[0].ToRGBA(prov.Embedded, keyColor)
	b := rgba.Bounds()
	face := color.RGBA{60, 64, 72, 255} // 採樣失敗時的保守深灰
	if b.Dx() > 8 && b.Dy() > 4 {
		r, g, bl, a := rgba.At(b.Min.X+4, b.Min.Y+b.Dy()/2).RGBA()
		if a > 0 {
			face = color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), 255}
		}
	}
	return ebiten.NewImageFromImage(rgba), face
}

func newLoadGameScreen(b *sceneBuilder, mode saveLoadMode, back func() (*overlayScreen, error), backName string) *loadGameScreen {
	s := &loadGameScreen{b: b, fnt: b.fnt, mode: mode, back: back, backName: backName, selected: -1}
	s.bg, _ = loadWinImage(b.res, loadWinBGAsset, false)
	s.loadBtn, s.loadFace = loadWinImage(b.res, loadWinBtnAsset, true)
	s.cancelBtn, s.cancelFace = loadWinImage(b.res, loadWinCanAsset, true)
	s.singleIcon, _ = loadWinImage(b.res, loadWinIconSing, true)
	s.hotseatIcon, _ = loadWinImage(b.res, loadWinIconHot, true)
	// 視窗置中(openorion2:_x = (SCREEN_WIDTH − _width) / 2)。
	s.winW, s.winH = 276, 376 // 取不到背景時的退路尺寸(實測 game.lbx#20 為此尺寸)
	if s.bg != nil {
		s.winW, s.winH = s.bg.Bounds().Dx(), s.bg.Bounds().Dy()
	}
	s.winX, s.winY = (moo2ScreenW-s.winW)/2, (moo2ScreenH-s.winH)/2
	s.slots = shell.ReadSaveSlots(saveDirFor())
	// 讀檔預選最近的存檔(十格都要玩家自己找太煩,而 Continue 本來就是接續最近進度);
	// 存檔預選第一個空格,免得手一滑蓋掉舊進度。
	if mode == modeLoad {
		if latest := shell.LatestSaveSlot(saveDirFor()); latest >= 0 {
			s.selected = latest
		}
		return s
	}
	for i, sl := range s.slots {
		if !sl.Exists && i != shell.AutoSaveSlot { // 自動存檔格不給手動存
			s.selected = i
			break
		}
	}
	return s
}

// slotRect 回傳第 i 格的螢幕座標(openorion2:createWidget(22, 22 + 31*i, 232, 27),
// 座標相對視窗左上)。
func (s *loadGameScreen) slotRect(i int) (int, int, int, int) {
	return s.winX + loadSlotX, s.winY + loadSlotY0 + loadSlotStep*i, loadSlotW, loadSlotH
}

func (s *loadGameScreen) loadRect() (int, int, int, int) {
	w, h := loadCanW, loadCanH // LOAD 鈕沒有明寫尺寸(用圖的),取不到圖時沿用 CANCEL 的
	if s.loadBtn != nil {
		w, h = s.loadBtn.Bounds().Dx(), s.loadBtn.Bounds().Dy()
	}
	return s.winX + loadBtnX, s.winY + loadBtnY, w, h
}

func (s *loadGameScreen) cancelRect() (int, int, int, int) {
	return s.winX + loadCanX, s.winY + loadCanY, loadCanW, loadCanH
}

func (s *loadGameScreen) update(in shell.InputState) *origTransition {
	if !in.ClickReleased {
		return nil
	}
	if x, y, w, h := s.cancelRect(); hitRect(in, x, y, w, h) {
		return s.b.goTo(s.back, s.backName)
	}
	if x, y, w, h := s.loadRect(); hitRect(in, x, y, w, h) {
		if s.mode == modeSave {
			return s.doSave()
		}
		return s.doLoad()
	}
	for i := range s.slots {
		if x, y, w, h := s.slotRect(i); hitRect(in, x, y, w, h) {
			if s.mode == modeLoad && !s.slots[i].Exists {
				s.msg = s.b.tr("這一格是空的", "That slot is empty.")
				return nil
			}
			if s.mode == modeSave && i == shell.AutoSaveSlot {
				s.msg = s.b.tr("最後一格是自動存檔,不能手動覆寫",
					"The last slot is the autosave; it cannot be overwritten by hand.")
				return nil
			}
			s.selected, s.msg = i, ""
			return nil
		}
	}
	return nil
}

// doSave 把目前對局寫進選定的槽。
func (s *loadGameScreen) doSave() *origTransition {
	if s.selected < 0 || s.selected >= len(s.slots) || s.selected == shell.AutoSaveSlot {
		s.msg = s.b.tr("請先選一格(自動存檔格除外)", "Pick a slot first (not the autosave).")
		return nil
	}
	if s.b.session == nil {
		s.msg = s.b.tr("目前沒有進行中的對局", "There is no game in progress.")
		return nil
	}
	path := s.slots[s.selected].Path
	if s.slots[s.selected].NativeGAM {
		// 原版 `.GAM` 只可作匯入來源；覆寫同一槽時另建 remake JSON。
		path = shell.SaveSlotPath(saveDirFor(), s.selected)
	}
	if err := s.b.session.Save(path); err != nil {
		fmt.Fprintln(os.Stderr, "存檔失敗:", err)
		s.msg = s.b.tr("寫不進去(磁碟或權限問題)", "Could not write the file (disk or permissions).")
		return nil
	}
	// 之後的自動存檔跟著寫到這一格,與讀檔後的行為一致。
	s.b.savePath = path
	return s.b.goTo(s.back, s.backName)
}

// doLoad 讀取選定的槽並進星系主畫面。
func (s *loadGameScreen) doLoad() *origTransition {
	if s.selected < 0 || s.selected >= len(s.slots) || !s.slots[s.selected].Exists {
		s.msg = s.b.tr("請先選一格有存檔的格子", "Pick a slot that has a saved game.")
		return nil
	}
	gs, err := shell.LoadSession(s.slots[s.selected].Path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "讀檔失敗:", err)
		s.msg = s.b.tr("這個存檔讀不起來(格式可能來自舊版本)",
			"That save would not load (it may come from an older build).")
		return nil
	}
	s.b.session = gs
	s.b.applyNebulaStarFlags(s.b.session) // 星雲判定式不進存檔,讀檔後要重裝(見 nebula.go)
	if len(s.b.herodataMercs) > 0 {
		s.b.session.SetMercCandidates(s.b.herodataMercs) // 讀檔建的是新 session,重注入真英雄池
	}
	// 之後的自動存檔要寫回同一格,不然玩家讀了第 3 格、存檔卻跑去覆蓋自動存檔格。
	// 原版 `.GAM` 是唯讀匯入來源；讀入後的自動／手動存檔要落到同槽的
	// remake JSON，不能沿用 `.GAM` 路徑覆寫原始存檔。
	if s.slots[s.selected].NativeGAM {
		s.b.savePath = shell.SaveSlotPath(saveDirFor(), s.selected)
	} else {
		s.b.savePath = s.slots[s.selected].Path
	}
	return s.b.goTo(s.b.galaxy, "星系主畫面")
}

func (s *loadGameScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{6, 8, 14, 255})
	if s.bg != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(s.winX), float64(s.winY))
		drawPanelImage(dst, s.bg, op)
	}
	if s.fnt == nil {
		return
	}
	body := color.RGBA{206, 214, 232, 255}
	sel := color.RGBA{250, 226, 130, 255}
	dim := color.RGBA{120, 126, 140, 255}

	for i, sl := range s.slots {
		col := body
		if i == s.selected {
			col = sel
			x, ry, w, h := s.slotRect(i)
			vector.StrokeRect(dst, float32(x), float32(ry), float32(w), float32(h), 1, sel, false)
		}
		// 文字基準取自 openorion2 drawSlot:_x + 32、_y + 24 + 31*slot。
		tx := float64(s.winX + loadSlotTextX)
		ty := float64(s.winY + loadTextY + loadSlotStep*i)
		if !sl.Exists {
			label := s.b.tr("（空）", "(empty)")
			if sl.Auto {
				label = s.b.tr("（自動存檔：尚無）", "(autosave: none yet)")
			}
			c := dim
			if s.mode == modeSave && i == s.selected {
				c = sel // 存檔模式下空格是可選的
			}
			s.fnt.Draw(dst, label, tx, ty, 12, c)
			continue
		}
		name := sl.Empire
		if name == "" {
			name = s.b.tr("無名帝國", "Unnamed Empire") // 舊存檔沒存帝國名(見 persist.go 該欄位註解)
		}
		title := fmt.Sprintf("%d. %s", i+1, name)
		if sl.Auto {
			title = fmt.Sprintf(s.b.tr("自動  %s", "AUTO  %s"), name)
		}
		s.fnt.Draw(dst, title, tx, ty, 12, col)
		// 第二行:星曆在名稱正下方、存檔時間在 +122(原版 drawSlot 的兩個 renderText)。
		s.fnt.Draw(dst, fmt.Sprintf(s.b.tr("星曆 %s", "Stardate %s"), sl.Stardate), tx, ty+loadSlotSubDY, 10, col)
		if !sl.Modified.IsZero() {
			// 日期用兩位數年份(26/08/06):這一欄從 x+122 起,x+206 就是對局類型圖示,
			// 只有 84px 可用。原版在這裡用的是 C 的 `%x`,在 C locale 同樣是兩位數年份的
			// 八字元日期——寫完整年月日會壓到圖示上,那是版面本來就沒有的空間。
			s.fnt.Draw(dst, sl.Modified.Format("06/01/02 15:04"),
				float64(s.winX+loadSlotTimeX), ty+loadSlotSubDY, 10, col)
		}
		// 對局類型圖示:remake 只會是單人或熱座(網路/數據機沒做)。
		icon := s.singleIcon
		if sl.Hotseat {
			icon = s.hotseatIcon
		}
		if icon != nil {
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(s.winX+loadSlotIconX), ty+loadSlotIconDY)
			drawPanelImage(dst, icon, op)
		}
	}

	if s.loadBtn != nil {
		x, y, _, _ := s.loadRect()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		drawPanelImage(dst, s.loadBtn, op)
	}
	if s.cancelBtn != nil {
		x, y, _, _ := s.cancelRect()
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(x), float64(y))
		drawPanelImage(dst, s.cancelBtn, op)
	}
	// 兩顆鈕的英文(LOAD / CANCEL)烘在圖上,照既有做法擦底疊中文:先用採樣到的面色
	// 填掉文字帶(上下左右各留 3px 保住浮雕邊框),再疊中文。
	drawBtnLabel := func(x, y, w, h int, face color.RGBA, zh string) {
		if s.b.lang != i18n.Traditional {
			return // 英文模式露原版鈕上烘的 LOAD / SAVE / CANCEL
		}
		if face.A > 0 {
			fillPanel(dst, float32(x+3), float32(y+3), float32(w-6), float32(h-6), face, false)
		}
		s.fnt.DrawCentered(dst, zh, float64(x+w/2), float64(y+h/2), 13, body)
	}
	action := s.b.tr("載入", "LOAD")
	if s.mode == modeSave {
		action = s.b.tr("儲存", "SAVE")
	}
	lx, ly, lw, lh := s.loadRect()
	drawBtnLabel(lx, ly, lw, lh, s.loadFace, action)
	cx, cy, cw, ch := s.cancelRect()
	drawBtnLabel(cx, cy, cw, ch, s.cancelFace, s.b.tr("取消", "CANCEL"))

	if s.msg != "" {
		s.fnt.DrawCentered(dst, s.msg, 320, float64(s.winY+s.winH+14), 12,
			color.RGBA{235, 160, 120, 255})
	}
}

// loadGame 從主選單進入載入遊戲視窗(取消回主選單)。
func (b *sceneBuilder) loadGame() (origScreen, error) {
	return newLoadGameScreen(b, modeLoad, b.menu, "主選單"), nil
}

// loadGameInPlay 從遊戲中的「遊戲選單」進入載入視窗(取消回星系主畫面)。
func (b *sceneBuilder) loadGameInPlay() (origScreen, error) {
	return newLoadGameScreen(b, modeLoad, b.galaxy, "星系主畫面"), nil
}

// saveGameInPlay 從遊戲中的「遊戲選單」進入儲存視窗(同一個視窗,見檔頭)。
func (b *sceneBuilder) saveGameInPlay() (origScreen, error) {
	return newLoadGameScreen(b, modeSave, b.galaxy, "星系主畫面"), nil
}
