package main

// raceselect.go:原版獨立「種族選擇」畫面(HANDOFF 優先2)。
//
// 依 docs/tech/newgame-flow.md:原版流程是 星系設定 → Accept → **獨立種族選擇畫面**
// (13 族 + Custom,滑過顯示肖像/能力)→ Custom 則進點數畫面 → 命名/旗色 → 進遊戲。
// 目前 remake 把種族擠進設定畫面一格;本畫面把它拆成原版的獨立畫面。
//
// ============ 版面(2026-08-07 改用反組譯真值)============
//
// ⚠ 這裡原本寫「版面為合成近似,尚未對原版截圖像素對齊」,而且擺法是**左文字清單 +
// 右肖像**——**跟原版左右相反**。原版是**左肖像 + 右邊 2 欄 × 7 列的種族按鈕網格**。
//
// 三個獨立來源互證(與 NEW GAME / 殖民地畫面同一套方法):
//
//	`Race_Selection_Screen_` @ 0x5C510 的建鈕迴圈(i = 0..13):
//	    y = 0x5A + 0x30 × (i mod 7)   → 90 + 48i',列距 48
//	    x = 0x15F + 0x7E × (i div 7)  → 351 / 477,欄距 126
//	`Draw_Race_Selection_Screen_` @ 0x5BD97:
//	    肖像 = RACESEL 資產 **15 + 種族索引**,畫在 (0x36, 0x3F) = **(54, 63)**
//	    標題橫幅 = 資產 0x21(33)畫在 (0x16E, 0x34) = (366, 52)
//	`RACESEL.LBX` 資產表:
//	    1–14  每張 **123×45**、2 幀(一般/選中)→ 與步距 126×48 差 3px 間隙,正是按鈕
//	    15–28 每張 **290×322** → 畫在 (54,63) 佔 54..344,與按鈕欄 x=351 剛好不重疊
//
// 順帶:remake 註解原本寫「肖像↔種族對應**經讀圖確認**為字母序」——那個結論是對的,
// 現在有反組譯硬證(`lea edx, [eax+0Fh]`,即 15 + 索引),不必再靠讀圖。
//
// ⚠ 原版這個畫面**沒有 ACCEPT 鈕**:建的 widget 只有 14 顆種族鈕 + 一個 ESC(取消)
// 處理器(`sub_114C72("\x1B")`)。點種族即確認。remake 照做:滑過看肖像、點下去確認。

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// raceEntry 是種族選擇清單一列。玩家名稱／形容詞／摘要全部由 ui.json 的穩定 key 提供。
type raceEntry struct {
	key      string
	portrait int
	shellIdx int
}

// raceSelectList 依原版字母序排列(對齊 RACESEL 肖像 15..28),shellIdx 指回 shell.Races。
// Custom(28)尚無點數畫面,暫以無加成處理(TODO:接 Race Customization 畫面)。
var raceSelectList = []raceEntry{
	{"alkari", 15, 6},
	{"bulrathi", 16, 5},
	{"darlok", 17, 8},
	{"elerian", 18, 10},
	{"gnolam", 19, 11},
	{"human", 20, 0},
	{"klackon", 21, 3},
	{"meklar", 22, 7},
	{"mrrshan", 23, 4},
	{"psilon", 24, 1},
	{"sakkra", 25, 2},
	{"silicoid", 26, 12},
	{"trilarian", 27, 9},
	{"custom", 28, -1},
}

func raceSelectEntryText(lang i18n.Lang, entry raceEntry, field string) string {
	return uiText(lang, "race.select.race."+entry.key+"."+field)
}

type raceSelectScreen struct {
	b         *sceneBuilder
	fnt       *uifont.Font
	bg        *ebiten.Image
	portraits map[int]*ebiten.Image // 肖像 asset id → 圖(惰性載入)
	buttons   map[int]*ebiten.Image // 種族鈕:asset*10+frame → 圖(惰性載入)
	sel       int                   // 目前選定列(raceSelectList 索引)
	hover     int                   // 滑鼠懸停列(-1=無)
}

// 版面常數:全部是反組譯真值 / 資產實測尺寸(見檔頭)。
const (
	rsBtnX0   = 351 // 0x15F,第一欄
	rsBtnColW = 126 // 0x7E,欄距
	rsBtnY0   = 90  // 0x5A,第一列
	rsBtnRowH = 48  // 0x30,列距
	rsBtnRows = 7   // 每欄 7 列(14 族 = 2 欄 × 7 列)
	rsBtnW    = 123 // RACESEL 資產 1–14 實測寬
	rsBtnH    = 45  // 同上,高

	rsPortX, rsPortY = 54, 63   // Draw_Race_Selection_Screen_ 的肖像位置
	rsPortW, rsPortH = 290, 322 // RACESEL 資產 15–28 實測尺寸

	rsTitleX, rsTitleY = 366, 52 // 標題橫幅(資產 33,219×29)
	rsTitleW, rsTitleH = 219, 29
	rsTitleAsset       = 33 // 0x21,Draw_Race_Selection_Screen_ 畫的標題橫幅

	// 肖像下右側是 remake 補的種族能力資訊框。它不能跨到最下方機械飾板，
	// 否則第二行描述會看起來像文字跑出對話框。
	rsInfoX, rsInfoY = 180, 395
	rsInfoW, rsInfoH = 164, 40

	// 種族鈕的浮雕外框不能被中文字碰到：面板本身內縮 4px，文字安全帶再多留
	// 2px。這個矩形是文字置中與截斷的唯一來源，避免日後只改一邊而把字移出按鈕。
	rsButtonFaceInset = 4
	rsButtonTextInset = 6
)

// raceSelect 建構種族選擇畫面。預設選「人類」(清單索引 5)。
func (b *sceneBuilder) raceSelect() (origScreen, error) {
	s := &raceSelectScreen{
		b: b, fnt: b.fnt, portraits: map[int]*ebiten.Image{},
		buttons: map[int]*ebiten.Image{}, sel: 5, hover: -1,
	}
	if im, err := decodeAsset(b.res, "raceopt.lbx", 0); err == nil && im.Embedded != nil {
		s.bg = ebiten.NewImageFromImage(im.Frames[0].ToRGBA(im.Embedded, im.KeyColor()))
	}
	return s, nil
}

// portrait 惰性載入某族肖像(RACESEL,內嵌調色盤)。
func (s *raceSelectScreen) portrait(assetID int) *ebiten.Image {
	if img, ok := s.portraits[assetID]; ok {
		return img
	}
	var img *ebiten.Image
	if im, err := decodeAsset(s.b.res, "racesel.lbx", assetID); err == nil && im.Embedded != nil {
		img = ebiten.NewImageFromImage(im.Frames[0].ToRGBA(im.Embedded, im.KeyColor()))
	}
	s.portraits[assetID] = img // 即使 nil 也快取,避免重複嘗試
	return img
}

// rowRect 回傳第 i 顆種族鈕的矩形:2 欄 × 7 列,座標為反組譯真值(見檔頭)。
func (s *raceSelectScreen) rowRect(i int) (x, y, w, h int) {
	return rsBtnX0 + (i/rsBtnRows)*rsBtnColW, rsBtnY0 + (i%rsBtnRows)*rsBtnRowH, rsBtnW, rsBtnH
}

// rowTextRect 回傳種族按鈕內可安全放置譯文的內框。它與原版資產的按鈕外框完全分離，
// 卻保持同一個整數像素中心，使點陣／向量字都不會因半像素位移而顯得模糊或偏移。
func (s *raceSelectScreen) rowTextRect(i int) textSafeRect {
	bx, by, bw, bh := s.rowRect(i)
	return textSafeRect{x: bx, y: by, w: bw, h: bh, insetX: rsButtonTextInset, insetY: rsButtonTextInset}
}

func raceSelectTitleTextRect() textSafeRect {
	return textSafeRect{x: rsTitleX, y: rsTitleY, w: rsTitleW, h: rsTitleH, insetX: 4}
}

func raceSelectInfoNameTextRect() textSafeRect {
	return textSafeRect{x: rsInfoX + 6, y: rsInfoY, w: rsInfoW - 12, h: 18, insetX: 2}
}

func raceSelectInfoDescriptionTextRect() textSafeRect {
	return textSafeRect{x: rsInfoX + 6, y: rsInfoY + 18, w: rsInfoW - 12, h: 20, insetX: 2, insetY: 1}
}

func (s *raceSelectScreen) cancelTextRect() textSafeRect {
	x, y, w, h := s.cancelRect()
	return textSafeRect{x: x, y: y, w: w, h: h, insetX: 4, insetY: 2}
}

// cancelRect 是「取消」。原版這個畫面沒有可見的取消鈕,只綁 ESC;remake 沒有鍵盤路徑,
// 借左下角肖像下方的空白區放一顆——**這是 remake 自己加的**,不是原版版面。
func (s *raceSelectScreen) cancelRect() (int, int, int, int) {
	return rsPortX, rsPortY + rsPortH + 10, 120, 28
}

func hitBox(mx, my, x, y, w, h int) bool {
	return mx >= x && mx < x+w && my >= y && my < y+h
}

func (s *raceSelectScreen) update(in shell.InputState) *origTransition {
	s.hover = -1
	for i := range raceSelectList {
		x, y, w, h := s.rowRect(i)
		if hitBox(in.MouseX, in.MouseY, x, y, w, h) {
			s.hover = i
			break
		}
	}
	if s.hover >= 0 {
		s.sel = s.hover // 滑過即預覽該族肖像(原版 word_19B858 = 懸停中的族)
	}
	if !in.ClickReleased {
		return nil
	}
	// 取消 → 回星系設定(原版是 ESC;remake 的鈕見 cancelRect 註解)。
	if x, y, w, h := s.cancelRect(); hitBox(in.MouseX, in.MouseY, x, y, w, h) {
		if clickSound != nil {
			clickSound()
		}
		return s.b.goTo(s.b.newGameSetup, uiText(s.b.lang, "race.select.transition.new_game"))
	}
	// 點種族鈕即確認(原版沒有 ACCEPT,見檔頭)。自訂種族先進點數畫面。
	if s.hover >= 0 {
		if clickSound != nil {
			clickSound()
		}
		s.sel = s.hover
		if raceSelectList[s.sel].shellIdx < 0 {
			sc, err := s.b.customRace()
			if err != nil {
				return nil
			}
			return &origTransition{next: sc}
		}
		s.applyAndStart()
		e := raceSelectList[s.sel]
		return &origTransition{next: s.b.nameFlag(fmt.Sprintf(uiText(s.b.lang, "race.select.empire_name"),
			raceSelectEntryText(s.b.lang, e, "adjective")))}
	}
	return nil
}

// applyAndStart 依設定畫面選的五個設定 + 本畫面選的種族開新局。
//
// 2026-08-07:「numAI 先固定 3、玩家可設定 AI 數的 UI 尚未接上」已不成立——NEW GAME 畫面的
// PLAYERS 欄接上了(那本來就是原版該欄的用途,見 interactive.go 的 ngSettings 檔頭),
// 對手數 = 帝國總數 − 1。星系年齡同樣接上了(先前被一個常數寫死成「普通」)。
//
// ⚠ 順序有講究:`GalaxyAge` 必須在 `SetupNewGame` **之前**設好——星系與行星就是在那裡
// 依年齡骰出來的,設晚了這一局的年齡設定等於沒選。
func (s *raceSelectScreen) applyAndStart() {
	b := s.b
	if b.session == nil {
		return
	}
	r := raceSelectList[s.sel]
	b.session.Difficulty = b.newGameDiff
	b.applyNewGameSettings()
	b.newGameSeed++
	b.session.SetupNewGame(shell.GalaxySizes[b.newGameSize].Stars, int64(b.newGameSeed*7919+42), b.newGameOpponents())
	b.session.SetRuleProfile(profileForVersion(b.gameVersion)) // 主選單選的 1.3/1.5 規則版本
	// 星雲的「在星雲內」旗標要讀星雲圖的遮罩,規則層碰不到資產,所以在這裡算好灌回去。
	// 星雲與星的位置之後都不會變,算一次就夠(見 cmd/moo2/nebula.go)。
	b.applyNebulaStarFlags(b.session)
	if r.shellIdx >= 0 {
		b.newGameRace = r.shellIdx
		b.session.ApplyRace(r.shellIdx) // Custom(-1)暫不套加成,待點數畫面
	}
}

func (s *raceSelectScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{0, 0, 0, 255})
	if s.bg != nil {
		drawPanelImage(dst, s.bg, nil)
	}
	if s.fnt == nil {
		return
	}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{206, 218, 240, 255}
	dim := color.RGBA{150, 160, 180, 255}

	// 左側:目前(懸停或選定)那一族的肖像,原尺寸畫在反組譯給的 (54,63)。
	// 資產本來就是 290×322,不縮放——縮放會讓 1996 年的點陣肖像糊掉。
	r := raceSelectList[s.sel]
	if p := s.portrait(r.portrait); p != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(rsPortX), float64(rsPortY))
		drawPanelImage(dst, p, op)
	}

	// 標題橫幅(資產 33 @ 366,52):**先畫原版那張**。
	// 先前只在中文模式擦掉這塊再疊中文,從沒真的畫過它——中文模式看不出來(反正要蓋掉),
	// 英文模式一跳過覆蓋就露出一片空白。畫了之後兩種語言都對:英文露原版標題,中文照蓋。
	if im := s.raceButton(rsTitleAsset, 0); im != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(rsTitleX, rsTitleY)
		drawPanelImage(dst, im, op)
	}
	// 上疊中文。**英文模式不擦不疊**:原版美術上本來就是英文標題。
	if s.b.lang == i18n.Traditional {
		fillPanel(dst, rsTitleX, rsTitleY, rsTitleW, rsTitleH,
			color.RGBA{26, 28, 34, 255}, false)
		raceSelectTitleTextRect().drawCentered(dst, s.fnt, uiText(s.b.lang, "race.select.title"), 16, gold)
	}

	// 右側:14 顆種族鈕,2 欄 × 7 列(座標為反組譯真值)。
	// 用原版的按鈕圖:第 0 幀=一般、第 1 幀=選中,再擦底疊中文族名。
	for i, e := range raceSelectList {
		x, y, w, h := s.rowRect(i)
		frame := 0
		if i == s.sel {
			frame = 1
		}
		if im := s.raceButton(e.portrait-14, frame); im != nil { // 肖像 15+i ↔ 按鈕 1+i
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			drawPanelImage(dst, im, op)
		} else {
			// 取不到按鈕圖時的退路:自繪一格,畫面仍可用。
			fillPanel(dst, float32(x), float32(y), float32(w), float32(h),
				color.RGBA{34, 38, 48, 255}, false)
		}
		// 擦掉烘在按鈕上的英文族名再疊中文(上下左右各留 4px 保住浮雕邊框)。
		// 英文模式跳過整段:原版按鈕上就是英文族名,擦掉再畫一次只會更醜。
		if s.b.lang != i18n.Traditional {
			continue
		}
		face := color.RGBA{92, 88, 78, 255}
		if i == s.sel {
			face = color.RGBA{70, 66, 58, 255}
		}
		fillPanel(dst, float32(x+rsButtonFaceInset), float32(y+rsButtonFaceInset),
			float32(w-2*rsButtonFaceInset), float32(h-2*rsButtonFaceInset), face, false)
		col := body
		if i == s.sel {
			col = gold
		}
		s.rowTextRect(i).drawCentered(dst, s.fnt, raceSelectEntryText(s.b.lang, e, "name"), 14, col)
	}

	// 肖像下方那一條(y 395..438)放:左邊取消鈕、右邊族名 + 能力描述。
	// ⚠ 這一區是 remake 自己的排版——原版這個畫面上沒有能力說明文字
	// (`sub_103915(60,175,寬280)` 那個文字帶只在「該族已被別的玩家選走」的多人分支才畫)。
	cx, cy, cw, ch := s.cancelRect()
	fillPanel(dst, rsInfoX, rsInfoY, rsInfoW, rsInfoH, color.RGBA{17, 29, 35, 255}, false)
	vector.StrokeRect(dst, rsInfoX, rsInfoY, rsInfoW, rsInfoH, 1, color.RGBA{108, 130, 150, 255}, false)
	raceSelectInfoNameTextRect().drawCentered(dst, s.fnt, raceSelectEntryText(s.b.lang, r, "name"), 12, gold)
	descCol := body
	if r.shellIdx < 0 {
		descCol = dim
	}
	raceSelectInfoDescriptionTextRect().drawCentered(dst, s.fnt,
		raceSelectEntryText(s.b.lang, r, "description"), 8, descCol)

	// 取消(remake 自己加的,原版只綁 ESC,見 cancelRect 註解)。
	fillPanel(dst, float32(cx), float32(cy), float32(cw), float32(ch),
		color.RGBA{34, 34, 44, 255}, false)
	vector.StrokeRect(dst, float32(cx), float32(cy), float32(cw), float32(ch), 1.5,
		color.RGBA{160, 140, 100, 255}, false)
	s.cancelTextRect().drawCentered(dst, s.fnt, uiText(s.b.lang, "race.select.cancel"), 14, body)
}

// raceButton 惰性載入某族的選擇鈕(RACESEL 資產 1–14,兩幀:一般 / 選中)。
// 這些資產沒有內嵌調色盤,借背景 RACEOPT#0 的。
func (s *raceSelectScreen) raceButton(assetID, frame int) *ebiten.Image {
	key := assetID*10 + frame
	if img, ok := s.buttons[key]; ok {
		return img
	}
	var img *ebiten.Image
	if pal, err := decodeAsset(s.b.res, "raceopt.lbx", 0); err == nil && pal.Embedded != nil {
		if im, err := decodeAsset(s.b.res, "racesel.lbx", assetID); err == nil && len(im.Frames) > frame {
			img = ebiten.NewImageFromImage(im.Frames[frame].ToRGBA(pal.Embedded, true))
		}
	}
	s.buttons[key] = img // 即使 nil 也快取,避免每幀重試
	return img
}
