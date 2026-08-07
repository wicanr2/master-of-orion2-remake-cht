package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	moo2audio "github.com/wicanr2/master-of-orion2-remake-cht/internal/audio"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// interactive.go:把「還原原版」的各原版畫面串成單一可互動、可導覽的程式(-game)。
//
// 核心設計(忠實原版 MOO2 + 繁中):畫面用真的 LBX 背景圖 + 中文標籤擦底疊字
// (overlayScreen,同 overlay.go 的手法),在原版按鈕位置加點擊熱區,滑鼠點選在
// 畫面間導覽。這是「用新技術還原原版 MOO2 並中文化」的骨幹——與自製簡約殼(play.go)
// 不同,這裡每個畫面都是原版美術。
//
// 目前串接:原版主選單 →(新遊戲/繼續)→ 原版行星列表 →(返回)→ 主選單。
// 後續逐畫面把更多原版畫面(殖民地/研究/星圖/戰鬥…)改為 overlay 真美術並接進導覽。

const moo2ScreenW, moo2ScreenH = 640, 480

// origTransition 是原版畫面切換指令。
type origTransition struct {
	next origScreen
	quit bool
}

// origScreen 是一個可互動的原版畫面。
type origScreen interface {
	update(in shell.InputState) *origTransition
	draw(dst *ebiten.Image)
}

// hitRegion 是畫面上一塊可點區域 + 動作 id(通常等於該按鈕的英文 key)。
type hitRegion struct {
	x, y, w, h int
	action     string
}

func (h hitRegion) hit(x, y int) bool {
	return x >= h.x && x < h.x+h.w && y >= h.y && y < h.y+h.h
}

// --- overlayScreen:原版 LBX 背景 + 中文標籤覆蓋 + 點擊熱區 ---

type overlayScreen struct {
	bg         *ebiten.Image
	rgba       *image.RGBA
	font       *uifont.Font
	cat        *i18n.Catalog
	overlays   []labelRect
	labelColor color.RGBA
	defSize    float64
	hits       []hitRegion
	onAction   func(action string) *origTransition
	// onHotkey 處理這個畫面的鍵盤快捷鍵(nil = 該畫面沒有快捷鍵)。
	// 目前只有星圖接了(見 cmd/moo2/hotkeys.go)。
	onHotkey         func(code string) *origTransition
	hover            string
	offsetX, offsetY int         // 背景圖在 640×480 畫布上的置中偏移(小於全螢幕的視窗畫面用)
	eraseColor       *color.RGBA // 非 nil 時強制用此色擦底(背景均勻的畫面用,勝過採樣猜測)
	eraseInsetX      int         // 擦底框在基準(左右各3px)之外「再往內縮」的水平量(每邊);0=不變
	eraseInsetY      int         // 擦底框在基準(上下各2px)之外「再往內縮」的垂直量(每邊);0=不變
	plateFace        bool        // true=擦底色改採按鈕面色(浮雕按鈕列用,見 samplePlate faceSample)
	// eraseInset 用途:浮雕按鈕的上下/左右斜邊會被擦底塊蓋掉 → 加內縮只擦中間文字帶,保留浮雕框
	// (仍蓋掉烘進的英文,因英文置中於文字帶內);plateFace 則讓擦底色貼合按鈕面,兩者可併用。
	// labelColorFor 以 enKey 覆寫個別標籤的顏色(空 = 全部用 labelColor)。
	// 目前唯一的用途是把「停用」的選項畫成灰的——原版主選單無存檔時 Continue / Load Game
	// 就是灰階不可按的(2026-07-12 archive.org oracle 對照 issue #2)。
	labelColorFor map[string]color.RGBA
	extras        []extraText             // 即時動態文字(星曆、國庫…),疊在背景+overlay 之上
	postDraw      func(dst *ebiten.Image) // 任意額外繪製(如星圖),在最後呼叫
	mx, my        int                     // 最近一次 update() 算出的滑鼠局部座標(扣掉置中偏移),供 postDraw 讀取做懸停偵測(如殖民地總覽 Planetary/Production Info)
}

// extraText 是一段即時繪製的動態文字(非來自譯表的固定標籤)。
type extraText struct {
	x, y  float64
	size  float64
	text  string
	col   color.RGBA
	align int // 0=靠左,1=置中
}

func (s *overlayScreen) update(in shell.InputState) *origTransition {
	// 命中判定在背景圖局部座標(扣掉置中偏移)。
	mx, my := in.MouseX-s.offsetX, in.MouseY-s.offsetY
	s.mx, s.my = mx, my
	s.hover = ""
	for _, h := range s.hits {
		if h.hit(mx, my) {
			s.hover = h.action
			break
		}
	}
	// 快捷鍵先於滑鼠處理:同一幀既按鍵又點擊時,鍵盤那步不該被點擊蓋掉。
	if in.Hotkey != "" && s.onHotkey != nil {
		if tr := s.onHotkey(in.Hotkey); tr != nil {
			return tr
		}
	}
	if in.ClickReleased {
		for _, h := range s.hits {
			if h.hit(mx, my) && s.onAction != nil {
				if clickSound != nil {
					clickSound() // 命中按鈕才播原版點擊音(SOUND.LBX BUTTON1)
				}
				return s.onAction(h.action)
			}
		}
	}
	return nil
}

func (s *overlayScreen) draw(dst *ebiten.Image) {
	if s.offsetX != 0 || s.offsetY != 0 {
		dst.Fill(color.RGBA{0, 0, 0, 255}) // 小於全螢幕的視窗:底填黑再置中
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(s.offsetX), float64(s.offsetY))
	dst.DrawImage(s.bg, op)
	ox, oy := float64(s.offsetX), float64(s.offsetY)
	if s.cat.Lang() == i18n.Traditional {
		for _, b := range s.overlays {
			// 擦掉烘進圖的英文:填單色底(eraseColor 指定則用之,否則取標籤帶的「中位數色」——
			// 代表性中間調,避免誤取過暗陰影形成黑框;單色填充不複製紋理,故不會有錯位歪斜)。
			plate := samplePlate(s.rgba, b, s.plateFace)
			if s.eraseColor != nil {
				plate = *s.eraseColor
			}
			// 基準內縮左右各3、上下各2;eraseInsetX/Y 再各邊多縮,保留浮雕框(見欄位註解)。
			ex := 3 + s.eraseInsetX
			ey := 2 + s.eraseInsetY
			vector.DrawFilledRect(dst, float32(float64(b.x+ex)+ox), float32(float64(b.y+ey)+oy),
				float32(b.w-2*ex), float32(b.h-2*ey), plate, false)
			// 疊中文(同 overlay.go)。
			size := b.size
			if size == 0 {
				size = s.defSize
			}
			zh := s.cat.Translate(b.enKey)
			lc := s.labelColor
			if c, ok := s.labelColorFor[b.enKey]; ok { // 停用的選項畫成灰的
				lc = c
			}
			s.font.DrawCentered(dst, zh, float64(b.x)+float64(b.w)/2+ox, float64(b.y)+float64(b.h)/2+oy, size, lc)
		}
	}
	// hover 熱區以細框提示可點(互動回饋)。
	if s.hover != "" {
		for _, h := range s.hits {
			if h.action == s.hover {
				vector.StrokeRect(dst, float32(float64(h.x)+ox), float32(float64(h.y)+oy),
					float32(h.w), float32(h.h), 1, color.RGBA{255, 240, 120, 200}, false)
			}
		}
	}
	// 即時動態文字(星曆、國庫…)。
	for _, e := range s.extras {
		if e.align == 1 {
			s.font.DrawCentered(dst, e.text, e.x+ox, e.y+oy, e.size, e.col)
		} else {
			s.font.Draw(dst, e.text, e.x+ox, e.y+oy, e.size, e.col)
		}
	}
	if s.postDraw != nil {
		s.postDraw(dst)
	}
}

// samplePlate 取標籤底板色(用來擦掉烘進圖的英文)。
// 策略:在「文字帶的上下緣margin」(置中文字不及此的乾淨底)+ 左內緣採樣一組像素,取
// 「中位數亮度色」——中位數為代表性中間調,對少數的亮字/暗陰影都穩健,不會像眾數那樣
// 誤取到反覆出現的過暗陰影而形成黑框。
// 註:背景均勻但文字靠左/寬粗填滿的畫面(如 info),改用 overlayScreen.eraseColor 強制底色。
//
// faceSample=true(浮雕按鈕列,如殖民地底列):只採「左右內緣的垂直中央帶」當面色,
// 跳過上下緣列——那兩列會落在按鈕的上亮/下暗斜邊或按鈕間隙,把中位數往暗拉,擦出來的
// 底板比按鈕面更暗、像挖了黑洞蓋住浮雕框。改採面色後,擦底與按鈕面同色,浮雕框自然可見。
func samplePlate(rgba *image.RGBA, b labelRect, faceSample bool) color.RGBA {
	W, H := rgba.Bounds().Dx(), rgba.Bounds().Dy()
	var cols []color.RGBA
	add := func(x, y int) {
		if x < 0 || x >= W || y < 0 || y >= H {
			return
		}
		i := rgba.PixOffset(x, y)
		cols = append(cols, color.RGBA{rgba.Pix[i], rgba.Pix[i+1], rgba.Pix[i+2], 255})
	}
	if faceSample {
		// 左右內緣各三行,取垂直中央帶(避開上下斜邊列與置中英文),得按鈕面色。
		for _, dx := range []int{3, 5, 7} {
			for y := b.y + 4; y < b.y+b.h-4; y++ {
				add(b.x+dx, y)
				add(b.x+b.w-1-dx, y)
			}
		}
	} else {
		// 上下緣各兩列(文字上下的乾淨底)橫跨全寬 + 左內緣窄帶。
		for _, y := range []int{b.y + 1, b.y + 2, b.y + b.h - 3, b.y + b.h - 2} {
			for x := b.x + 3; x < b.x+b.w-3; x += 2 {
				add(x, y)
			}
		}
		for _, dx := range []int{3, 5, 7} {
			for y := b.y + 3; y < b.y+b.h-3; y++ {
				add(b.x+dx, y)
			}
		}
	}
	if len(cols) == 0 {
		return color.RGBA{0, 0, 0, 255}
	}
	lum := func(c color.RGBA) int { return 30*int(c.R) + 59*int(c.G) + 11*int(c.B) }
	sort.Slice(cols, func(i, j int) bool { return lum(cols[i]) < lum(cols[j]) })
	return cols[len(cols)/2]
}

// assetRef 指向某 LBX 內一張影像。
type assetRef struct {
	lbxName string
	assetID int
}

// paletteChain 是「調色盤提供圖」的疊加鏈:某些原版畫面背景無完整內嵌調色盤,需借其他
// 帶調色盤的圖當基底(openorion2 的 base_palette 機制)。依序疊加(前者為基底,後者覆蓋
// 其內嵌範圍),最後再疊目標圖自己的內嵌範圍。空鏈表示畫面自帶完整可用調色盤。
// 註:提供圖不必填滿 256 色,只需其內嵌範圍涵蓋目標圖用到的索引即可(見 palette-chain.md)。
type paletteChain []assetRef

// decodeAsset 解一張 LBX 影像。
func decodeAsset(res *assets.Resolver, lbxName string, assetID int) (*lbx.Image, error) {
	arch, err := res.OpenLBX(lbxName)
	if err != nil {
		return nil, err
	}
	raw, err := arch.Asset(assetID)
	if err != nil {
		return nil, err
	}
	return lbx.DecodeImage(raw)
}

// overlayRange 把 src 的內嵌範圍疊寫到 dst。
func overlayRange(dst *lbx.Palette, src *lbx.Image) {
	if src.Embedded == nil {
		return
	}
	for i := src.PalStart; i < src.PalStart+src.PalCount; i++ {
		dst[i] = src.Embedded[i]
	}
}

// resolvePalette 重現 openorion2 Image::load 的調色盤合併:
// 依 chain 順序疊加各提供圖的內嵌範圍當基底,最後疊目標圖自己的內嵌範圍。
func resolvePalette(res *assets.Resolver, im *lbx.Image, chain paletteChain) (*lbx.Palette, error) {
	var merged lbx.Palette
	for _, ref := range chain {
		pim, err := decodeAsset(res, ref.lbxName, ref.assetID)
		if err != nil {
			return nil, fmt.Errorf("載入調色盤提供圖 %s#%d: %w", ref.lbxName, ref.assetID, err)
		}
		if pim.Embedded == nil {
			return nil, fmt.Errorf("調色盤提供圖 %s#%d 無內嵌調色盤", ref.lbxName, ref.assetID)
		}
		overlayRange(&merged, pim)
	}
	if len(chain) == 0 && im.Embedded == nil {
		return nil, fmt.Errorf("畫面圖無內嵌調色盤且未指定提供圖鏈")
	}
	overlayRange(&merged, im)
	return &merged, nil
}

// loadOverlayScreen 載入某原版畫面(LBX 背景 + 譯表),組成可互動的 overlayScreen。
// chain 非空時走調色盤鏈(無內嵌調色盤的畫面借提供圖上色)。
func loadOverlayScreen(res *assets.Resolver, lbxName string, assetID int, lang i18n.Lang,
	fnt *uifont.Font, tsvPath string, overlays []labelRect, labelColor color.RGBA, defSize float64,
	hits []hitRegion, onAction func(string) *origTransition, chain paletteChain) (*overlayScreen, error) {

	im, err := decodeAsset(res, lbxName, assetID)
	if err != nil {
		return nil, err
	}
	pal, err := resolvePalette(res, im, chain)
	if err != nil {
		return nil, fmt.Errorf("%s 資產 %d: %w", lbxName, assetID, err)
	}
	rgba := im.Frames[0].ToRGBA(pal, im.KeyColor())

	cat := i18n.New(lang)
	if f, err := os.Open(tsvPath); err == nil {
		defer f.Close()
		if _, err := cat.LoadTSV(f); err != nil {
			return nil, err
		}
	} else if lang == i18n.Traditional {
		return nil, fmt.Errorf("開啟譯表 %s: %w", tsvPath, err)
	}

	// 小於 640×480 的視窗畫面置中(openorion2:_x=(SCREEN_WIDTH-_width)/2)。
	bounds := rgba.Bounds()
	offX := (moo2ScreenW - bounds.Dx()) / 2
	offY := (moo2ScreenH - bounds.Dy()) / 2
	if offX < 0 {
		offX = 0
	}
	if offY < 0 {
		offY = 0
	}
	return &overlayScreen{
		bg: ebiten.NewImageFromImage(rgba), rgba: rgba, font: fnt, cat: cat,
		overlays: overlays, labelColor: labelColor, defSize: defSize,
		hits: hits, onAction: onAction, offsetX: offX, offsetY: offY,
	}, nil
}

// --- sceneBuilder:依需求建構各原版畫面(共用 resolver/字型/語言)---

type sceneBuilder struct {
	// skipCutscenes:headless 驗證與截圖廊要跳過流程中的過場影片——那些腳本是逐 tick
	// 數出來的,插一段會一直往前播的影片會整串偏掉。截圖廊另外用 tick 注入單獨截過場。
	skipCutscenes   bool
	res             *assets.Resolver
	fnt             *uifont.Font // 內文用字型(zh 為混合:內文點陣、標題向量)
	fntVec          *uifont.Font // 純向量 Noto(供主選單等要平滑的畫面;nil 時退回 fnt)
	lang            i18n.Lang
	session         *shell.GameSession       // 活的對局狀態(TURN 推進、畫面顯示即時資料)
	herodataMercs   []shell.Leader           // HERODATA.LBX 解出的真英雄傭兵候選(快取;讀檔後重注入)
	newGameSize     int                      // NEW GAME 選的星系大小索引(shell.GalaxySizes)
	newGameDiff     int                      // NEW GAME 選的難度索引(shell.Difficulties)
	newGameRace     int                      // NEW GAME 選的種族索引(shell.Races)
	newGameSeed     int                      // 每次新遊戲遞增,讓星系種子變化
	newGameAge      int                      // NEW GAME 選的星系年齡索引(shell.GalaxyAges)
	newGameTech     int                      // NEW GAME 選的起始科技等級索引(shell.TechLevels)
	newGameEmpires  int                      // NEW GAME 選的帝國總數(含玩家,shell.MinEmpires..MaxEmpires)
	colChrome       *ebiten.Image            // 殖民地畫面的原版框架(COLPUPS.LBX#5,惰性解碼快取)
	colBldgCache    map[string]*ebiten.Image // 地表建築圖(BLDGn.LBX,惰性解碼快取)
	colVegSizeCache map[int][2]int           // COLVEGGI 資產的寬高;地表每幀重算,不快取會每幀重解 LBX
	// animTick 是動畫用的重畫計數(由 interactiveApp.Update 每幀同步過來)。
	// 原版的動畫是「每次重畫推進一次計數」,remake 把「一次重畫」對應成一個 ebiten 幀,
	// 見 starsprite.go 黑洞那段的 ⚠。
	animTick int
	// measure 是 F9 測距模式的狀態(見 hotkeys.go)。放這裡不放 GameSession:
	// 它是「看的方式」不是世界狀態,不該進存檔。
	measure measureState
	// relocPick 是「正在挑集結點」的畫面狀態;relocColors 是遷移連線的 8 色漸層快取
	// (見 relocation.go)。兩者都是看的方式,不進存檔。
	relocPick   relocatePickState
	relocColors []color.RGBA
	// shipPick 是艦隊列表上被勾選的船(索引是**目前選中艦隊內**的索引),供拆分用。
	// 換艦隊時要清掉——索引換一支就沒意義了。
	shipPick map[int]bool
	// flashMsg / flashUntil 是星圖底緣的短暫訊息(F10 快速存檔的回報等,見 hotkeys.go)。
	// flashUntil 用 animTick 計時,到了就不畫。
	flashMsg          string
	flashUntil        int
	nebMaskCache      map[int]*nebulaMask  // 星雲遮罩;派遣時沿航線取樣上百次,不快取會重解上百次 LBX
	pendingHotseat    int                  // 多人設定畫面選的真人席位數;0/1 = 單人局(開局後由 applyHotseat 套用)
	savePath          string               // remake 存檔路徑(每回合自動存;主選單 Load/Continue 讀)
	designWeapon      int                  // 艦艇設計選的武器元件索引(shell.WeaponOptions)
	designArmor       int                  // 裝甲元件索引(shell.ArmorOptions)
	designShield      int                  // 護盾元件索引(shell.ShieldOptions)
	designSpecial     int                  // 特殊元件索引(shell.SpecialOptions)
	designMods        []string             // 目前設計勾選的武器改造(gamedata.WeaponModCode 字串;僅 beam 武器生效)
	designMsg         string               // 艦艇設計畫面「空間不足,擋下建造」的提示訊息(切換元件/成功建造時清空)
	lastActionMsg     string               // 星圖畫面「載運陸戰隊/發動地面入侵」的最近一次結果訊息(選新星時清空)
	gameVersion       gamedata.GameVersion // 主選單選的規則版本(1.3/1.5);開局注入 session.RuleProfile
	infoTab           int                  // INFO 畫面目前分頁(0=歷史圖表 1=科技總覽 2=種族統計 3=回合摘要 4=參考),見 infosubscreens.go
	colonyIdx         int                  // 單一殖民地畫面目前管理哪個殖民地(索引 PlayerColonies),見 colonyscreen.go
	colonyListTop     int                  // 單一殖民地畫面「可建項目」清單的捲動起點
	infoHistoryMetric int                  // 歷史圖表目前指標(shell.HistoryMetric)
	// planetPick 是行星列表畫面選中的行星索引(−1 = 沒選)。原版那個畫面的
	// SEND COLONY SHIP / SEND OUTPOST SHIP 兩顆鈕就是對著選中的那一列作用的。
	planetPick int
	// planetListTop 是行星列表的捲動起點(該畫面一次只顯示 8 列)。
	planetListTop int
	// planetListMsg 是行星列表畫面最近一次動作的結果訊息。
	planetListMsg string
}

// profileForVersion 把主選單選的版本轉成對應 RuleProfile(開局注入 session)。
func profileForVersion(v gamedata.GameVersion) gamedata.RuleProfile {
	if v == gamedata.VersionClassic13 {
		return gamedata.Profile13()
	}
	return gamedata.Profile15()
}

// versionShort 版本短名(主選單切換顯示用)。
func versionShort(v gamedata.GameVersion) string {
	if v == gamedata.VersionClassic13 {
		return "1.3"
	}
	return "1.5"
}

// saveDirFor 回傳 remake 存檔目錄(使用者設定目錄下,退回暫存目錄),確保可寫。
// 環境變數 `MOO2_SAVE_DIR` 可覆寫——headless 驗證/截圖廊用它把存檔導到暫存目錄,
// 免得測試把玩家真正的存檔覆蓋掉。
func saveDirFor() string {
	if d := os.Getenv("MOO2_SAVE_DIR"); d != "" {
		if err := os.MkdirAll(d, 0o755); err == nil {
			return d
		}
	}
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		dir = os.TempDir()
	}
	sub := filepath.Join(dir, "moo2-remake-cht")
	if mkErr := os.MkdirAll(sub, 0o755); mkErr != nil {
		return os.TempDir()
	}
	return sub
}

// savePathFor 回傳自動存檔的路徑(存檔槽的最後一格,見 shell/saveslots.go)。
// 從載入視窗讀某一格之後,`b.savePath` 會被改成那一格,後續自動存檔就寫回同一格。
func savePathFor() string { return shell.SaveSlotPath(saveDirFor(), shell.AutoSaveSlot) }

// menu 建原版主選單畫面。按鈕熱區用 menuOverlays 的座標(按鈕即標籤)。
func (b *sceneBuilder) menu() (*overlayScreen, error) {
	playSceneBGM(bgmMenu)
	// 無存檔時 Continue / Load Game **停用**:不給熱區(點了不動作)+ 標籤畫成灰的。
	// 這是 2026-07-12 archive.org 原版 oracle 對照的 issue #2 結論——原版那兩顆本來就是
	// 灰階不可按,remake 先前是「可按但靜默無反應」,玩家會以為壞了。
	hasSave := shell.AnySaveExists(saveDirFor())
	dimmed := map[string]color.RGBA{}
	hits := make([]hitRegion, 0, len(menuOverlays)+1)
	for _, o := range menuOverlays {
		if !hasSave && (o.enKey == "Continue" || o.enKey == "Load Game") {
			dimmed[o.enKey] = color.RGBA{104, 116, 104, 255} // 綠色主選單字的暗化版
			continue                                         // 不加熱區 = 點不到
		}
		hits = append(hits, hitRegion{o.x, o.y, o.w, o.h, o.enKey})
	}
	// 規則版本切換(CLAUDE.md:主選單選 1.3/1.5)——左下角熱區,點擊循環切換,開局注入 RuleProfile。
	hits = append(hits, hitRegion{12, 450, 220, 22, "toggleVersion"})
	// 語言切換(CLAUDE.md:「允許在主選單選擇中文/英文」)——擺在版本切換正上方。
	// 先前語言只有啟動旗標 `-lang`,進了遊戲就換不掉,不符合這條需求。
	hits = append(hits, hitRegion{12, 428, 220, 22, "toggleLang"})
	onAction := func(a string) *origTransition {
		switch a {
		case "toggleLang":
			// 切語言 = 換整個顯示層:overlayScreen 是在建構時把譯表烘進去的,
			// 所以改完 b.lang 要重建畫面才會生效(與版本切換同款做法)。
			// 英文模式下 overlay 機制整段跳過擦底疊字,直接露出原版烘進圖的英文。
			if b.lang == i18n.Traditional {
				b.lang = i18n.English
			} else {
				b.lang = i18n.Traditional
			}
			return b.goTo(b.menu, "主選單")
		case "toggleVersion":
			if b.gameVersion == gamedata.VersionClassic13 {
				b.gameVersion = gamedata.VersionCommunity15
			} else {
				b.gameVersion = gamedata.VersionClassic13
			}
			return b.goTo(b.menu, "主選單") // 重繪以更新版本顯示
		case "Quit Game":
			return &origTransition{quit: true}
		case "New Game":
			// 新遊戲:先進原版 NEW GAME 設定畫面(難度/星系/玩家…),ACCEPT 後進星系主畫面。
			b.pendingHotseat = 0 // 單人局(從多人設定畫面進來的才會帶席位數)
			return b.goTo(b.newGameSetup, "新遊戲設定")
		case "Multi Player":
			// 原版 MULTI-PLAYER GAME SET UP(見 cmd/moo2/multiplayer.go)。
			// remake 只實作其中的 HOTSEAT,其餘連線方式在那個畫面裡明示未實作。
			sc, err := b.multiPlayer()
			if err != nil {
				fmt.Fprintln(os.Stderr, "多人遊戲設定:", err)
				return nil
			}
			return &origTransition{next: sc}
		case "Continue":
			// 續玩:接續**最近的存檔槽**(原版 Continue 就是接最近進度,不彈選單)。
			// 無存檔時這顆鈕是停用的,理論上進不來;真進來了就當作沒事發生。
			dir := saveDirFor()
			slot := shell.LatestSaveSlot(dir)
			if slot < 0 {
				return nil
			}
			path := shell.SaveSlotPath(dir, slot)
			gs, err := shell.LoadSession(path)
			if err != nil {
				fmt.Fprintln(os.Stderr, "讀檔失敗:", err)
				return nil
			}
			b.session = gs
			b.applyNebulaStarFlags(b.session) // 星雲判定式不進存檔,讀檔後要重裝(見 nebula.go)
			if len(b.herodataMercs) > 0 {
				b.session.SetMercCandidates(b.herodataMercs) // 讀檔建的是新 session,重注入真英雄池
			}
			b.savePath = path // 後續自動存檔寫回同一格
			return b.goTo(b.galaxy, "星系主畫面")
		case "Load Game":
			// 開原版的十格存檔選擇視窗(見 cmd/moo2/loadgame.go)。無存檔時這顆鈕停用。
			if !shell.AnySaveExists(saveDirFor()) {
				return nil
			}
			sc, err := b.loadGame()
			if err != nil {
				fmt.Fprintln(os.Stderr, "載入遊戲視窗:", err)
				return nil
			}
			return &origTransition{next: sc}
		case "Hall of Fame":
			// 名人堂 → 最終得分畫面(原版 Hall of Fame / Hi-Score,見 cmd/moo2/hiscore.go)。
			// ⚠ 這裡先前暫借給「研究選擇」畫面當調色盤鏈的示範入口,是接錯的,2026-08-07 改正。
			return b.goTo(b.hiScore, "最終得分")
		}
		return nil
	}
	// 主選單用純向量 Noto(平滑),不走內文點陣(使用者要求主選單維持向量觀感);
	// 無 fntVec(如 zh 未帶 -font)時退回 b.fnt。
	menuFont := b.fntVec
	if menuFont == nil {
		menuFont = b.fnt
	}
	s, err := loadOverlayScreen(b.res, "mainmenu.lbx", 21, b.lang, menuFont, "assets/i18n/menu.tsv",
		menuOverlays, color.RGBA{104, 224, 96, 255}, 15, hits, onAction, nil)
	if err != nil {
		return nil, err
	}
	if len(dimmed) > 0 {
		s.labelColorFor = dimmed // 無存檔 → Continue / Load Game 畫成灰的
	}
	// 左下角版本 / 語言切換標籤(點擊上方熱區循環)。
	// 語言標籤本身要跟著當前語言走,否則英文模式下留一行中文很怪。
	langLabel := b.tr("語言 繁體中文(點此切換)", "Language: English (click to switch)")
	verLabel := fmt.Sprintf(b.tr("規則版本 %s(點此切換)", "Rules %s (click to switch)"),
		versionShort(b.gameVersion))
	s.extras = append(s.extras,
		extraText{x: 16, y: 436, size: 13, text: langLabel, col: color.RGBA{150, 210, 150, 255}},
		extraText{x: 16, y: 458, size: 13, text: verLabel, col: color.RGBA{150, 210, 150, 255}},
	)
	return s, nil
}

// tr 是**自繪畫面**的雙語切換:回傳目前語言該顯示的字串。
//
// 為什麼需要:`overlayScreen` 那套「擦底疊字」在英文模式會整段跳過,直接露出原版烘在
// 美術上的英文——那條路徑本來就雙語正確。但**自繪畫面**(remake 自己畫的面板:多人設定、
// 熱座交接、命名旗色、地面戰、轟炸、遊戲選單、載入視窗、自訂種族…)底下沒有原版英文可露,
// 字是程式直接寫死的中文,於是**英文模式下那些畫面仍然全是中文**。
//
// CLAUDE.md 明列「允許在主選單選擇中文/英文」,所以這是個真的缺口,不是設計取捨。
// 現況與待補清單見 docs/HONEST-STATUS.md 的「英文模式覆蓋率」一節。
func (b *sceneBuilder) tr(zh, en string) string {
	if b.lang == i18n.Traditional {
		return zh
	}
	return en
}

// goTo 建構下一個場景並包成 transition;失敗時記錄錯誤並留在原畫面。
func (b *sceneBuilder) goTo(build func() (*overlayScreen, error), name string) *origTransition {
	s, err := build()
	if err != nil {
		fmt.Fprintf(os.Stderr, "載入%s: %v\n", name, err)
		return nil
	}
	return &origTransition{next: s}
}

// backHit 回傳「點畫面任一處返回」的單一全螢幕熱區 + 導回指定場景的動作(過場/子畫面暫用,
// 待各畫面 RETURN 按鈕座標校對後改為精確熱區)。
func (b *sceneBuilder) backHit(dest func() (*overlayScreen, error), name string) ([]hitRegion, func(string) *origTransition) {
	return []hitRegion{{0, 0, moo2ScreenW, moo2ScreenH, "back"}},
		func(string) *origTransition { return b.goTo(dest, name) }
}

// galaxy 建原版星系主畫面(遊戲主樞紐,BUFFER0.LBX 資產 0)。底部工具列導覽到各畫面
// (座標取自 openorion2 galaxy.cpp GalaxyView::initWidgets)。
func (b *sceneBuilder) galaxy() (*overlayScreen, error) {
	playSceneBGM(bgmGalaxy)
	hits := []hitRegion{
		{15, 430, 67, 44, "colonies"},
		{90, 430, 67, 44, "planets"},
		{165, 430, 67, 44, "fleets"},
		{310, 430, 70, 44, "leaders"},
		{385, 430, 70, 44, "races"},
		{460, 430, 70, 44, "info"},
		{249, 5, 66, 24, "gamemenu"}, // 頂端「遊戲」鈕(openorion2 GalaxyView:createWidget(249,5,…GAME_BUTTON))
		{544, 441, 90, 34, "turn"},
		{547, 52, 65, 67, "taxrate"},   // 國庫框:點擊循環工業稅率(手冊 p.37,0-50%/10%級距)
		{547, 348, 65, 67, "research"}, // 研究框(右側第5格):點擊開研究選擇畫面(對齊原版,取代 info→tech 錯接)
		// 指揮框(右側第2格):點擊開指揮點數視窗(原版 `Show_Command_Points_Screen_`
		// @ 0x8BAB9,見 commandpoints.go)。先前這格只顯示一個淨值數字,點不開。
		{547, 124, 65, 67, "commandpoints"},
	}
	// 會談請求燈(星圖上緣,原版 `Draw_Diplomacy_Request_Lights_`):點一下直接進對談。
	// 放在星球熱區**之前** append —— 燈在 y=5,和星圖區(y≥24)不重疊,順序其實無所謂,
	// 但先放能讓「點得到」這件事不依賴後面的熱區數量。
	if b.session != nil {
		for n := range b.session.AudienceRequests() {
			x, y, w, h := audienceLightRect(n)
			if x < 0 {
				break
			}
			hits = append(hits, hitRegion{x, y, w, h, fmt.Sprintf("audience%d", n)})
		}
	}
	// 星圖各星加點擊熱區(點星 → 顯示該星系行星資訊)。
	if b.session != nil {
		sess := b.session
		for i, st := range sess.Stars {
			sx, sy := starScreenPos(st)
			hits = append(hits, hitRegion{sx - starHitHalf, sy - starHitHalf, 2 * starHitHalf, 2 * starHitHalf, fmt.Sprintf("star%d", i)})
		}
		// 選中星資訊面板內的操作鈕(座標同 postDraw 繪製的按鈕框):依艦隊/選中星狀態擇一或
		// 兩顆並存——派遣艦隊(艦隊不在選中星)、載運陸戰隊(艦隊在玩家母星,唯一已知有
		// Marine Barracks 殖民地模型對映的星,見 shell.AIOpponent.ColonyStars 註解同款限制)、
		// 拓殖(艦隊在無主星且載有殖民船,見 shell.GameSession.ColonizeStar)。敵殖民地星
		// (Owner==2)例外為雙鈕共存:軌道轟炸(shell.GameSession.BombardColony,恆可用,402)
		// + 發動地面入侵(shell.GameSession.InvadeColony,額外要求已載運陸戰隊,424)。
		if sess.SelectedStar >= 0 && sess.SelectedStar < len(sess.Stars) {
			hits = append(hits, hitRegion{216, 330, 20, 16, "closestar"}) // 面板右上 CLOSE(✕),對齊原版 Star System 彈窗
			switch {
			case sess.Fleet().ETA > 0:
				// 航行中,面板只顯示狀態文字,無按鈕。
			case sess.SelectedStar == sess.Fleet().AtStar:
				switch {
				case sess.SelectedStar == 0:
					hits = append(hits, hitRegion{38, 402, 190, 20, "loadmarines"})
				case sess.Stars[sess.SelectedStar].Owner == 2:
					// 敵殖民地:軌道轟炸恆可用(艦隊武器開火,不需陸戰隊);
					// 發動地面入侵額外要求已載運陸戰隊,兩鈕不同列共存(402/424)。
					hits = append(hits, hitRegion{38, 402, 190, 20, "bombard"})
					if sess.Fleet().Marines > 0 {
						hits = append(hits, hitRegion{38, 424, 190, 20, "invade"})
					}
				case sess.StarGuardedByMonster(sess.SelectedStar):
					// 怪獸盤據:唯一能做的是打它(手冊 p.62:清場之後才能進駐)。
					hits = append(hits, hitRegion{38, 402, 190, 20, "attackmonster"})
				default:
					// 無主星或**自己的**星系:拓殖(需殖民船)與建前哨站(需前哨船)。
					// 氣態巨星/小行星帶只有後者可用,一般行星兩者都可以(手冊 p.85:前哨站
					// 不限宜居世界)。
					//
					// ⚠ 自己的星系也算 —— 同一個星系可以有多個殖民地(手冊 p.61 的條件是
					// 「那顆**行星**沒被殖民」)。這裡的鈕是「該星系下一顆可用天體」的捷徑;
					// 要指定是哪一顆,走原版的行星列表畫面(planets(),那才是原版選行星的地方)。
					for _, r := range starPanelColonyRows(sess) {
						hits = append(hits, hitRegion{38, r.y, 190, 20, r.action})
					}
				}
			default:
				hits = append(hits, hitRegion{38, 402, 190, 20, "dispatch"})
			}
			// 集結點:選中的是**自己的殖民地**就可以設(原版 star[+0x54+玩家×2],見
			// internal/shell/relocation.go)。放在第二列——第一列在上面各分支已被佔走,
			// 而「是不是自己的殖民地」與「艦隊在不在這裡」是兩件獨立的事。
			if colonyIndexAtStar(sess, sess.SelectedStar) >= 0 {
				hits = append(hits, hitRegion{38, 424, 190, 20, "relocate"})
			}
		}
	}
	onAction := func(a string) *origTransition {
		if len(a) > 4 && a[:4] == "star" && b.session != nil {
			if idx, err := strconv.Atoi(a[4:]); err == nil {
				// 測距(F9)與挑集結點模式先吃掉點星:那一下是模式的輸入,不是「選這顆星」。
				if b.measureClickedStar(idx) || b.relocatePickClickedStar(idx) {
					return b.goTo(b.galaxy, "星系主畫面")
				}
				if idx == b.session.SelectedStar {
					b.session.SelectedStar = -1 // 再點同一顆星 → 取消選取(關閉資訊面板,issue #6)
				} else {
					b.session.SelectedStar = idx
				}
				b.lastActionMsg = ""             // 換選中星,清掉上一顆星的動作結果訊息
				return b.goTo(b.galaxy, "星系主畫面") // 重繪顯示選中星資訊
			}
		}
		if a == "closestar" && b.session != nil {
			b.session.SelectedStar = -1 // 面板 CLOSE 鈕 → 關閉(對齊原版 Star System 彈窗的 CLOSE,issue #6)
			b.lastActionMsg = ""
			return b.goTo(b.galaxy, "星系主畫面")
		}
		if a == "research" {
			return b.goTo(b.research, "研究選擇") // 星系右側研究框 → 研究選擇(對齊原版)
		}
		if strings.HasPrefix(a, "audience") && b.session != nil {
			// 點會談燈 → 直接進外交對談,對象是那位敲門的對手,並把請求清掉(算談過了)。
			x, y, w, h := 0, 0, 0, 0
			idx := -1
			for n := range b.session.AudienceRequests() {
				if fmt.Sprintf("audience%d", n) == a {
					x, y, w, h = audienceLightRect(n)
					idx = b.audienceLightAt(x+w/2, y+h/2)
					break
				}
			}
			if idx < 0 {
				return nil
			}
			name := audienceEnemyName(b.session, idx)
			b.session.ClearAudienceRequest(idx)
			sc, err := b.diplomacyWith(name)
			if err != nil {
				return nil
			}
			return &origTransition{next: sc}
		}
		if a == "commandpoints" {
			// 星圖右欄第 2 格 → 指揮點數視窗(原版 `Show_Command_Points_Screen_`)。
			// 它是 origScreen 不是 overlayScreen(沒有底圖 LBX),所以自己組 transition。
			sc, err := b.commandPoints()
			if err != nil {
				return nil
			}
			return &origTransition{next: sc}
		}
		if a == "relocate" && b.session != nil {
			ci := colonyIndexAtStar(b.session, b.session.SelectedStar)
			if ci < 0 {
				return nil
			}
			_ = ci
			b.beginRelocatePickFrom(b.session.SelectedStar)
			b.flash(b.tr("點一顆星當集結點(點自己就取消)", "Click a star as rally point (click itself to clear)"))
			return b.goTo(b.galaxy, "星系主畫面")
		}
		if a == "dispatch" && b.session != nil {
			// 派遣艦隊至選中星(航行由 EndTurn 推進)。曲速前開局沒有 FTL、出不了本星系,
			// SendFleet 會直接回 false——要說清楚原因,不然玩家只會看到「點了沒反應」。
			if !b.session.SendFleet(b.session.SelectedStar) {
				if !b.session.FleetHasFTL() {
					b.lastActionMsg = b.tr(
						"沒有 FTL 引擎,艦隊無法離開本星系——先研究「核分裂」"+
							"(曲速前開局的規則,手冊:探索本星系之外在研究出 FTL 之前是不可能的)",
						"No FTL drive — the fleet cannot leave this system. Research Nuclear Fission first "+
							"(pre-warp start; the manual: exploring outside this system is impossible until FTL).")
				}
			}
			return b.goTo(b.galaxy, "星系主畫面")
		}
		if a == "loadmarines" && b.session != nil {
			n := b.session.LoadMarines(0) // 母星是唯一已知殖民地索引對映(見上方熱區註解)
			if n > 0 {
				b.lastActionMsg = fmt.Sprintf(b.tr("已載運 %d 名陸戰隊上艦", "%d marines loaded aboard"), n)
			} else {
				b.lastActionMsg = b.tr("無陸戰隊可載運(駐軍不足或艦隊已滿載)",
					"No marines to load (garrison too small, or the fleet is full)")
			}
			return b.goTo(b.galaxy, "星系主畫面")
		}
		if a == "bombard" && b.session != nil {
			res := b.session.BombardColony(b.session.SelectedStar)
			if !res.Ok { // 前置條件不足:沒開炸,留在星系主畫面說明原因
				b.lastActionMsg = res.Reason
				return b.goTo(b.galaxy, "星系主畫面")
			}
			// 真的炸了 → 進轟炸畫面(原版 Colony_Bombing_Screen_),見 cmd/moo2/bombing.go。
			sc, err := b.bombing(res)
			if err != nil {
				fmt.Fprintln(os.Stderr, "軌道轟炸:", err)
				return b.goTo(b.galaxy, "星系主畫面")
			}
			return &origTransition{next: sc}
		}
		if a == "invade" && b.session != nil {
			res := b.session.InvadeColony(b.session.SelectedStar)
			if !res.Ok { // 前置條件不足:沒開打,留在星系主畫面說明原因
				b.lastActionMsg = res.Reason
				return b.goTo(b.galaxy, "星系主畫面")
			}
			// 真的打了一場 → 進地面戰畫面(原版 Colony_Combat),見 cmd/moo2/groundcombat.go。
			sc, err := b.groundCombat(res)
			if err != nil {
				fmt.Fprintln(os.Stderr, "地面戰:", err)
				return b.goTo(b.galaxy, "星系主畫面")
			}
			return &origTransition{next: sc}
		}
		if a == "colonize" && b.session != nil {
			res := b.session.ColonizeStar(b.session.SelectedStar)
			if !res.Ok {
				b.lastActionMsg = res.Reason
			} else {
				b.lastActionMsg = fmt.Sprintf(b.tr("拓殖成功!新殖民地起始人口 %d(上限 %d)",
					"Colony founded — starting population %d (max %d)"), res.StartPopulation, res.PopMax)
			}
			return b.goTo(b.galaxy, "星系主畫面")
		}
		if a == "attackmonster" && b.session != nil {
			res := b.session.AttackMonster(b.session.SelectedStar)
			if !res.Ok {
				b.lastActionMsg = res.Reason
			} else {
				b.lastActionMsg = res.Message
			}
			return b.goTo(b.galaxy, "星系主畫面")
		}
		if a == "outpost" && b.session != nil {
			res := b.session.BuildOutpost(b.session.SelectedStar)
			if !res.Ok {
				b.lastActionMsg = res.Reason
			} else {
				b.lastActionMsg = b.tr("軍事前哨站建立完成——掃描範圍已往外推(手冊:前哨站沒有產出)",
					"Outpost established — scanning range extended (the manual: outposts produce nothing)")
			}
			return b.goTo(b.galaxy, "星系主畫面")
		}
		switch a {
		case "gamemenu":
			// 原版的存檔/讀檔/開新局/離開都在這個視窗裡(見 cmd/moo2/gamemenu.go)。
			sc, err := b.gameMenu()
			if err != nil {
				fmt.Fprintln(os.Stderr, "遊戲選單:", err)
				return nil
			}
			return &origTransition{next: sc}
		case "colonies":
			return b.goTo(b.colonySummary, "殖民地總覽")
		case "planets":
			return b.goTo(b.planets, "行星列表")
		case "fleets":
			return b.goTo(b.fleet, "艦隊列表")
		case "leaders":
			return b.goTo(b.officer, "軍官列表")
		case "info":
			return b.goTo(b.info, "科技總覽")
		case "races":
			return b.goTo(b.races, "種族關係")
		case "taxrate":
			// 點國庫框循環工業稅率(手冊 p.37,0-50%/10%級距):更多稅=更多 BC 但更慢建造。
			b.session.CycleTaxRate()
			return b.goTo(b.galaxy, "星系主畫面")
		case "turn":
			// 核心迴圈:結算一回合(玩家帝國 + 各 AI 對手決策),再顯示回合摘要(原版流程)。
			// 熱座多人時「結束回合」先交棒,全員下完令才推進世界(見 cmd/moo2/hotseat.go)。
			return b.endTurnPressed()
		}
		return nil
	}
	// 工具列標籤擦底疊字(x 為按鈕中心對齊,y 中心經 PIL 量測:一般列 450、TURN 455)。
	overlays := []labelRect{
		{13, 443, 71, 14, "Colonies", 12},
		{88, 443, 71, 14, "Planets", 12},
		{254, 1, 88, 19, "Game", 13}, // 頂部標題列烘進的 GAME
		{163, 443, 71, 14, "Fleets", 12},
		{235, 443, 74, 14, "Zoom", 12},
		{308, 443, 74, 14, "Leaders", 12},
		{383, 443, 74, 14, "Races", 12},
		{458, 443, 74, 14, "Info", 12},
		{544, 448, 90, 15, "Turn", 12},
	}
	s, err := loadOverlayScreen(b.res, "buffer0.lbx", 0, b.lang, b.fnt, "assets/i18n/menu.tsv",
		overlays, color.RGBA{210, 216, 230, 255}, 12, hits, onAction, nil)
	if err != nil {
		return nil, err
	}
	// 星圖的鍵盤快捷鍵(F1/F2 循環艦隊、F5/F6 切換已殖民星系、F9 測距;見 hotkeys.go)。
	s.onHotkey = func(code string) *origTransition {
		// ALT+F9 是換畫面(載入視窗),和其他幾個「只換選取星」的不同,單獨處理。
		if code == shell.HotkeyLoadGame {
			sc, err := b.loadGameInPlay()
			if err != nil {
				return nil
			}
			return &origTransition{next: sc}
		}
		if b.handleGalaxyHotkey(code) {
			return b.goTo(b.galaxy, "星系主畫面")
		}
		return nil
	}
	// 星圖(中央視窗,openorion2 StarmapWidget 20,20,507,401)+ 即時狀態文字(疊在星圖之上)。
	if b.session != nil {
		sess := b.session
		fnt := s.font
		s.postDraw = func(dst *ebiten.Image) {
			// 每幀算一次可見性(#13 輕量戰爭迷霧),避免 drawStarmap 逐星重算。
			// 星圖底:純黑 + 原版 `Draw_Paralax_` 的三層星空(見 starbg.go)。
			b.drawStarmapBackground(dst)
			b.drawNebulae(dst, sess.Nebulae) // 背景地形,壓在星星之下
			// 遷移連線是原版圖層順序的第 2 層(星星是第 3 層),所以畫在星星之下。
			b.drawRelocationLinks(dst)
			vis := sess.VisibleStars()
			drawStarmap(b, dst, fnt, sess.Stars, sess.SelectedStar, vis)
			b.drawGateIcons(dst, vis)             // 狀態標示,蓋在星星之上
			b.drawAudienceLights(dst)             // 會談請求燈(星圖上緣,原版 y=5、由右往左)
			b.drawMeasureOverlay(dst, s.mx, s.my) // F9 測距(手冊:游標移到哪就顯示到哪)
			b.drawFlash(dst)                      // 短暫訊息(F10 快速存檔的回報等)
			if fnt != nil {
				// 狀態數字畫進原版右側資訊格(openorion2 galaxy.cpp:1552-1588 硬編位置,
				// 對齊 buffer0.lbx#0 背景烘印的圖示格):星曆→頂右薄框(549,27,63,13)、
				// 國庫→硬幣格(547,52,65,67)底部。原版這些格是「圖示+數字」,故只畫數字、
				// 不畫中文標籤(標籤即圖示),不再疊在星圖左上蓋住星點。
				year := 3500 + (sess.Turn - 1)
				fnt.DrawCentered(dst, fmt.Sprintf("%d", year), 580, 34, 11, color.RGBA{240, 220, 120, 255})
				// 右側 5 格全部填數字(openorion2 galaxy.cpp:1552-1585 五格位置):
				// 國庫(547,52)/指揮(547,124)/食物(547,199)/貨運(547,273)/研究燒瓶格留給主題名於左上。
				infoCol := color.RGBA{245, 230, 150, 255}
				fnt.DrawCentered(dst, fmt.Sprintf("%d", sess.Player.BC), 579, 110, 12, infoCol) // 國庫 BC
				// 工業稅率(點國庫框循環,手冊 p.37):畫在國庫格頂端,提示可調的經濟槓桿。
				fnt.DrawCentered(dst, fmt.Sprintf(b.tr("稅%d%%", "TAX %d%%"), sess.Player.TaxRate),
					579, 62, 9, color.RGBA{205, 215, 165, 255})
				// 現算,不用 EndTurn 才更新的快取欄位(見 shell.CommandPointsSupplyNow)。
				netCmd := sess.CommandPointsSupplyNow() - sess.CommandPointsUsedNow()
				fnt.DrawCentered(dst, fmt.Sprintf("%d", netCmd), 579, 182, 12, infoCol) // 指揮評等(供給-需求)
				foodSum, rpSum := 0, 0
				for i := range sess.PlayerColonies {
					out := engine.RunColonyTurn(sess.PlayerColonies[i])
					foodSum += out.FoodSurplus
					rpSum += out.Research
				}
				fnt.DrawCentered(dst, fmt.Sprintf("%d", foodSum), 579, 257, 12, infoCol)                      // 食物盈餘
				fnt.DrawCentered(dst, fmt.Sprintf("%d", sess.Player.ActiveFreighters), 579, 331, 12, infoCol) // 運輸艦數
				// 第 5 格(綠燒瓶)= 每回合研究點數,補齊右欄五格。
				//
				// ⚠ 2026-08-07:先前這裡把「研究:<主題名>」畫在星圖左上 (30,34)。那是 remake 自己
				// 加的一行——**原版星圖沒有這個東西**(研究主題只在研究畫面顯示),而且它會壓到
				// 左上角的星星與艦隊圖示。原版這格放的是數字,跟其他四格一樣,所以改成數字。
				fnt.DrawCentered(dst, fmt.Sprintf("%d", rpSum), 579, 405, 12, infoCol)
				// 艦隊位置標記(青色三角)+ 航行目的連線。
				if sess.Fleet().AtStar >= 0 && sess.Fleet().AtStar < len(sess.Stars) {
					fx, fy := starScreenPos(sess.Stars[sess.Fleet().AtStar])
					if sess.Fleet().DestStar >= 0 && sess.Fleet().DestStar < len(sess.Stars) {
						dx, dy := starScreenPos(sess.Stars[sess.Fleet().DestStar])
						vector.StrokeLine(dst, float32(fx), float32(fy), float32(dx), float32(dy), 1, color.RGBA{80, 220, 220, 180}, false)
					}
					// 艦隊圖示:原版 `Draw_Ship_Icons_` 的帶旗色小艦艇(見 shipicon.go)。
					// 取不到資產才退回舊的青色方塊——那是佔位,不是原版的東西。
					if !b.drawShipIconAt(dst, sess.FlagColor, fx, fy) {
						vector.DrawFilledRect(dst, float32(fx-4), float32(fy-4), 8, 8, color.RGBA{80, 240, 240, 255}, false)
					}
				}
				// 選中星:顯示該星系行星資訊 + 派遣艦隊/載運陸戰隊/軌道轟炸/發動入侵按鈕(左下角面板)。
				if sess.SelectedStar >= 0 && sess.SelectedStar < len(sess.Planets) {
					p, _ := sess.PlanetDataAt(sess.SelectedStar)
					// 面板高度 132(非原版 110):敵殖民地時軌道轟炸(402)/地面入侵(424)雙鈕
					// 共存需多留一列,否則第二顆鈕會露出面板背景框之外。
					vector.DrawFilledRect(dst, 28, 326, 210, 132, color.RGBA{10, 14, 30, 235}, false)
					vector.StrokeRect(dst, 28, 326, 210, 132, 1, color.RGBA{90, 130, 200, 255}, false)
					fnt.Draw(dst, p.Name, 38, 344, 14, color.RGBA{240, 220, 120, 255})
					// 右上 CLOSE 鈕(✕),對齊上方 "closestar" 熱區與原版彈窗 CLOSE(issue #6)。
					fnt.Draw(dst, "✕", 220, 344, 14, color.RGBA{235, 150, 140, 255})
					// 行星特殊物產(金礦/寶石礦/原住民/遠古文物…)接在星名右邊,另用一色標出來——
					// 這是「這顆星值不值得搶」的關鍵資訊,埋在下面兩行環境資料裡會被忽略。
					if mon := sess.MonsterNameAtStar(sess.SelectedStar); mon != "" {
						nameW, _ := fnt.Measure(p.Name, 14)
						fnt.Draw(dst, "☠"+mon, 38+nameW+10, 346, 11, color.RGBA{240, 130, 150, 255})
					} else if sp := gamedata.PlanetSpecialName(p.SpecialID); sp != "" {
						nameW, _ := fnt.Measure(p.Name, 14)
						spW, _ := fnt.Measure("★"+sp, 11)
						// 星名很長時往左推,確保不會壓到右上角的 ✕(x=220)。
						sx := 38 + nameW + 10
						if sx+spW > 214 {
							sx = 214 - spW
						}
						fnt.Draw(dst, "★"+sp, sx, 346, 11, color.RGBA{250, 200, 100, 255})
					}
					fnt.Draw(dst, fmt.Sprintf(b.tr("氣候 %s ／ 大小 %s", "Climate %s / Size %s"), p.Climate, p.Size),
						38, 362, 11, color.RGBA{210, 216, 230, 255})
					fnt.Draw(dst, fmt.Sprintf(b.tr("重力 %s ／ 礦產 %s", "Gravity %s / Minerals %s"), p.Gravity, p.Mineral),
						38, 378, 11, color.RGBA{210, 216, 230, 255})
					// 同系其他天體(氣態巨星/小行星帶)的完整摘要放在行星列表畫面——這個面板
					// 只有 344~400 這四列的空間,402 起是操作鈕,再塞一列會壓到按鈕。
					// 陸戰隊狀態行:艦隊目前載運數,選中母星時另顯示殖民地駐軍池數(唯一已知對映)。
					marineLine := fmt.Sprintf(b.tr("艦隊陸戰隊 %d", "Fleet marines %d"), sess.Fleet().Marines)
					if sess.SelectedStar == 0 && len(sess.PlayerColonyMarines) > 0 {
						marineLine = fmt.Sprintf(b.tr("艦隊陸戰隊 %d／殖民地駐軍 %d", "Fleet marines %d / colony garrison %d"),
							sess.Fleet().Marines, sess.PlayerColonyMarines[0])
					}
					fnt.Draw(dst, marineLine, 38, 394, 11, color.RGBA{200, 220, 170, 255})
					// 操作鈕/狀態(與 galaxy() 建 hits 時的判斷邏輯一致)。
					switch {
					case b.lastActionMsg != "":
						vector.DrawFilledRect(dst, 38, 402, 190, 20, color.RGBA{30, 55, 35, 235}, false)
						vector.StrokeRect(dst, 38, 402, 190, 20, 1, color.RGBA{110, 200, 140, 255}, false)
						fnt.Draw(dst, b.lastActionMsg, 42, 415, 10, color.RGBA{225, 240, 225, 255})
					case sess.Fleet().ETA > 0:
						fnt.Draw(dst, fmt.Sprintf(b.tr("艦隊航行中…剩 %d 回合", "Fleet in transit — %d turns left"), sess.Fleet().ETA),
							38, 415, 11, color.RGBA{120, 200, 240, 255})
					case sess.SelectedStar == sess.Fleet().AtStar && sess.SelectedStar == 0:
						vector.DrawFilledRect(dst, 38, 402, 190, 20, color.RGBA{40, 70, 120, 255}, false)
						vector.StrokeRect(dst, 38, 402, 190, 20, 1, color.RGBA{110, 160, 230, 255}, false)
						fnt.Draw(dst, b.tr("▶ 載運陸戰隊", "▶ Load marines"), 46, 415, 12, color.RGBA{230, 235, 245, 255})
					case sess.SelectedStar == sess.Fleet().AtStar && sess.Stars[sess.SelectedStar].Owner == 2:
						// 軌道轟炸恆可用(艦隊武器開火,不需陸戰隊),畫在 402 這列。
						vector.DrawFilledRect(dst, 38, 402, 190, 20, color.RGBA{90, 60, 130, 255}, false)
						vector.StrokeRect(dst, 38, 402, 190, 20, 1, color.RGBA{170, 140, 230, 255}, false)
						fnt.Draw(dst, b.tr("▶ 軌道轟炸", "▶ Bombard"), 46, 415, 12, color.RGBA{240, 235, 250, 255})
						// 發動地面入侵額外要求已載運陸戰隊,畫在下一列(424),與轟炸鈕並存。
						if sess.Fleet().Marines > 0 {
							vector.DrawFilledRect(dst, 38, 424, 190, 20, color.RGBA{120, 50, 40, 255}, false)
							vector.StrokeRect(dst, 38, 424, 190, 20, 1, color.RGBA{230, 130, 110, 255}, false)
							fnt.Draw(dst, b.tr("▶ 發動地面入侵", "▶ Invade"), 46, 437, 12, color.RGBA{245, 235, 230, 255})
						}
					case sess.SelectedStar == sess.Fleet().AtStar && sess.StarGuardedByMonster(sess.SelectedStar):
						vector.DrawFilledRect(dst, 38, 402, 190, 20, color.RGBA{110, 45, 60, 255}, false)
						vector.StrokeRect(dst, 38, 402, 190, 20, 1, color.RGBA{230, 120, 140, 255}, false)
						fnt.Draw(dst, b.tr("▶ 挑戰", "▶ Attack ")+sess.MonsterNameAtStar(sess.SelectedStar),
							46, 415, 12, color.RGBA{250, 230, 235, 255})
					case sess.SelectedStar == sess.Fleet().AtStar && len(starPanelColonyRows(sess)) > 0:
						// 與 galaxy() 建 hits 的判斷共用同一支 starPanelColonyRows,
						// 免得「畫得出來卻點不到」或反過來。
						for _, r := range starPanelColonyRows(sess) {
							row := float32(r.y)
							face, edge, ink := color.RGBA{40, 110, 60, 255}, color.RGBA{130, 220, 150, 255}, color.RGBA{235, 245, 235, 255}
							label := b.tr("▶ 建立殖民地", "▶ Colonize")
							if r.action == "outpost" {
								face, edge, ink = color.RGBA{45, 80, 110, 255}, color.RGBA{140, 190, 230, 255}, color.RGBA{230, 240, 250, 255}
								label = b.tr("▶ 建立前哨站", "▶ Build outpost")
							}
							// 目標天體寫進鈕裡——同星系有多顆天體時,玩家要看得出這一下會落在哪顆。
							if r.planet >= 0 && r.planet < len(sess.Planets) {
								label += "：" + sess.Planets[r.planet].Name
							}
							vector.DrawFilledRect(dst, 38, row, 190, 20, face, false)
							vector.StrokeRect(dst, 38, row, 190, 20, 1, edge, false)
							fnt.Draw(dst, truncateToWidth(fnt, label, 12, 182), 46, float64(row)+13, 12, ink)
						}
					case sess.SelectedStar == sess.Fleet().AtStar:
						fnt.Draw(dst, b.tr("艦隊已在此星", "Fleet is already here"), 38, 415, 11, color.RGBA{140, 200, 140, 255})
					default:
						vector.DrawFilledRect(dst, 38, 402, 190, 20, color.RGBA{40, 70, 120, 255}, false)
						vector.StrokeRect(dst, 38, 402, 190, 20, 1, color.RGBA{110, 160, 230, 255}, false)
						fnt.Draw(dst, b.tr("▶ 派遣艦隊至此星", "▶ Send fleet here"), 46, 415, 12, color.RGBA{230, 235, 245, 255})
					}
					// 集結點鈕(第二列):選中的是自己的殖民地才有。標題直接寫出目前設到哪,
					// 不然玩家看不出「有沒有設」——那正是這個功能最容易被忽略的地方。
					if ci := colonyIndexAtStar(sess, sess.SelectedStar); ci >= 0 {
						vector.DrawFilledRect(dst, 38, 424, 190, 20, color.RGBA{45, 95, 85, 255}, false)
						vector.StrokeRect(dst, 38, 424, 190, 20, 1, color.RGBA{120, 200, 180, 255}, false)
						label := b.tr("▶ 設定集結點", "▶ Set rally point")
						if to := sess.ColonyRelocation(ci); to >= 0 && to < len(sess.Stars) {
							label = b.tr("▶ 集結點:", "▶ Rally: ") + sess.Stars[to].Name
						}
						fnt.Draw(dst, label, 46, 437, 12, color.RGBA{225, 245, 240, 255})
					}
				}
			}
		}
	}
	return s, nil
}

// 星圖視窗座標(openorion2 StarmapWidget 20,20,507,401)。
const starVX0, starVY0, starVX1, starVY1 = 24, 24, 523, 418

// starScreenPos 把星的正規化座標映射到星圖視窗的螢幕座標(供繪製與點擊命中共用)。
func starScreenPos(st shell.Star) (int, int) {
	return starVX0 + int(st.X*(starVX1-starVX0)), starVY0 + int(st.Y*(starVY1-starVY0))
}

// drawStarmap 在星系主畫面中央視窗繪製星圖(深空底 + 依光譜上色/大小定半徑的星 + 星名 +
// 我方/敵方擁有環 + 選中星高亮環)。
//
// visible 是 shell.GameSession.VisibleStars() 算好的可見性陣列(等長 stars,索引對應;nil
// 表示呼叫端還沒接上可見性資料,視為全部可見,維持舊行為不炸開)——診斷/測試等尚未串 session
// 的呼叫路徑不受影響。
//
// fog 純視覺(diff 全量表 #13):對 !visible[i] 的星,不畫星名(未知)、不畫擁有環(未偵測不知
// 道歸屬)、星點降低亮度變暗淡小點;可見星維持原本全繪。刻意不 gate 任何操作——選星/派艦/殖民
// /轟炸等既有流程完全不受影響,玩家仍可對著霧裡的暗星點擊派艦探索。
// drawStarmap 畫星圖上的星球。**底色與星空背景由 `drawStarmapBackground` 負責**,
// 這裡不再自己塗底——順序寫死在呼叫端,免得兩邊各塗一次互相蓋掉。
//
// b 只用來取原版的星球 sprite(見 starsprite.go);取不到就退回色圓,
// 那是**沒有原版資產時的降級**,不是原版的樣子。
func drawStarmap(b *sceneBuilder, dst *ebiten.Image, fnt *uifont.Font, stars []shell.Star, selected int, visible []bool) {
	const vx0, vy0, vx1, vy1 = starVX0, starVY0, starVX1, starVY1
	drawWormholeLinks(dst, stars, visible)
	for i, st := range stars {
		seen := visible == nil || (i < len(visible) && visible[i])

		x := float32(vx0) + float32(st.X)*(vx1-vx0)
		y := float32(vy0) + float32(st.Y)*(vy1-vy0)
		col, ok := spectralColors[uint8(st.Spectral)]
		if !ok {
			col = color.RGBA{200, 200, 200, 255}
		}
		r := float32(6 - st.Size) // 降級色圓的半徑:大=6 .. 小=3
		if r < 3 {
			r = 3
		}
		if !seen {
			// 未偵測到的霧星:縮成暗灰小點,不畫 sprite / 擁有環 / 星名 / 選中高亮(未知歸屬)。
			vector.DrawFilledCircle(dst, x, y, 2, color.RGBA{60, 64, 76, 255}, true)
			continue
		}
		// 原版 sprite 先畫,拿它的半徑當外圈環的基準;取不到就退回色圓。
		sprite := b.drawStarSpriteAt(dst, st, int(x), int(y))
		if sprite > 0 {
			r = sprite
		}
		// 選中星:黃色高亮環。
		if i == selected {
			vector.StrokeCircle(dst, x, y, r+6, 2, color.RGBA{255, 240, 120, 255}, true)
		}
		// 星雲內:淡紫虛環。原版沒有這個標示(玩家用肉眼看星壓在雲上),但星雲圖被壓成
		// 半透明後邊界不明顯,而「在不在星雲內」直接決定打起來有沒有護盾——
		// 看不出來的規則等於沒有規則。見 cmd/moo2/nebula.go。
		if st.InNebula {
			vector.StrokeCircle(dst, x, y, r+8, 1, nebulaStarTint, true)
		}
		// 擁有環:我方藍綠、敵方紅。
		switch st.Owner {
		case 1:
			vector.StrokeCircle(dst, x, y, r+3, 1.5, color.RGBA{90, 230, 180, 255}, true)
		case 2:
			vector.StrokeCircle(dst, x, y, r+3, 1.5, color.RGBA{235, 90, 80, 255}, true)
		}
		if sprite == 0 {
			vector.DrawFilledCircle(dst, x, y, r, col, true)
		}
		// 星名:原版 `Print_A_Star_Name_` @ 0x87768 是**置中在星球正下方**,而且會夾在
		// 星圖框內(x 22..527)——remake 先前畫在星球右側,長名字會直接壓出框外。
		//
		//	x = 星球中心 − 字寬/2      ; sub_12066F 量寬再減半
		//	y = 星球中心 + sprite 半徑  ; var_14 + 邊長/2 − 大小
		//	夾擠:x >= 22、x + 字寬 <= 527
		//
		// 原版還會依縮放換字型樣式(`Zoom_Level_Font_Style_`)並加描邊
		// (`Set_Outline_Color(2)`),remake 的 CJK 字型沒有對應的樣式表,只做位置。
		if fnt != nil && st.Name != "" {
			nw, _ := fnt.Measure(st.Name, 11)
			nx := float64(x) - nw/2
			if nx < starVX0 {
				nx = starVX0
			}
			if nx+nw > starVX1 {
				nx = starVX1 - nw
			}
			fnt.Draw(dst, st.Name, nx, float64(y)+float64(r), 11, color.RGBA{170, 185, 210, 255})
		}
	}
}

// drawWormholeLinks 畫蟲洞連線(原版 `Draw_Wormhole_Links_` @ 0x85593,星圖的**第 1 層**
// ——在星球之前,所以線會被星球蓋住而不是壓在上面)。
//
//	for i:  other = star[i].+0x29;  if other == -1: continue
//	        兩端都不可見就跳過(sub_79E32 / sub_79E06 的探索狀態檢查)
//	        Line(座標(i), 座標(other), 顏色 4)
//
// 原版兩端各畫一次(同一條線畫兩遍,無害)。remake 只畫 i < other 那一次,結果相同。
//
// ⚠ 顏色「4」是原版調色盤索引,remake 的星圖不是索引色畫布,沒有對應物;
// 用一個低飽和的青紫色,讓它看得出來又不搶星球。
func drawWormholeLinks(dst *ebiten.Image, stars []shell.Star, visible []bool) {
	seen := func(i int) bool { return visible == nil || (i < len(visible) && visible[i]) }
	col := color.RGBA{120, 100, 190, 190}
	for i, st := range stars {
		j := st.Wormhole
		// 只畫一次:i < j。單向/越界的資料在讀檔時就被 normalizeWormholes 清掉了。
		if j <= i || j >= len(stars) {
			continue
		}
		if !seen(i) && !seen(j) {
			continue // 兩端都沒偵測到就不揭露
		}
		x1 := float32(starVX0) + float32(st.X)*(starVX1-starVX0)
		y1 := float32(starVY0) + float32(st.Y)*(starVY1-starVY0)
		x2 := float32(starVX0) + float32(stars[j].X)*(starVX1-starVX0)
		y2 := float32(starVY0) + float32(stars[j].Y)*(starVY1-starVY0)
		vector.StrokeLine(dst, x1, y1, x2, y2, 1, col, true)
	}
}

// colonySummary 建原版殖民地總覽畫面(COLSUM.LBX 資產 0,自帶完整調色盤)。
// openorion2 未實作此 view,背景資產由本專案自 LBX 探測定位。
//
// 2026-07-11 查證下方 4 個預覽面板(game tester 回報「3黑1雜訊,疑似解碼失敗」):
// 對照 GAME_MANUAL.pdf(patch1.5)p.38-40「Colonies [C]」一節,原版明文列出這 4 格由左到右
// 依序是 Planetary Info / Production Info / Mini Map(顯示殖民地在銀河的位置)/ 一個「稍後說明」
// 的方格(後續段落證實是 Empire Summary:國庫/收支/人口/食物/研究)。
// 逐像素比對後確認:第 3 格(x≈380-508)那片「白雜訊」實際是 alpha=0(keyColor 透明)背景上
// 散布著真正不透明的星點像素——即 Mini Map 本該有的星圖縮圖,不是調色盤錯或解碼失敗
// (若是解碼錯誤,黑色像素也會是 alpha=0 或呈現隨機色塊,不會是「透明底 + 乾淨星點」這種
// 有意義的稀疏圖樣)。其餘 3 格(x≈10-92 / 102-371 / 516-628)是 alpha=255 的純黑,同樣是
// 正確解碼的原始美術——原版設計是「靜態黑底 + 執行期依滑鼠懸停的殖民地動態疊圖」。
// 結論:不動這 4 格底圖(它們是正確的原版美術,亂改等於銷毀真實資產,見 rulebook 83 完整性
// 原則),疊字改用 s.postDraw 逐幀繪製。
//
// 2026-07-11 追加(懸停互動):第 4 格 Empire Summary 用既有 session 欄位(LastPlayerOutput.*、
// Player.BC)畫成文字,固定不隨懸停變動。第 1、2 格 Planetary Info / Production Info 現接上
// 「目前懸停哪個殖民地」——overlayScreen 新增 mx/my 欄位在 update() 逐幀記錄局部滑鼠座標,
// s.postDraw 逐幀依 mx/my 落在哪一列殖民地表格列,畫出該殖民地的行星資訊(氣候/重力/礦產/
// 大小/人口上限)與生產資訊(食物/工業/研究/污染);無懸停落在任何列時預設顯示殖民地 0
// (母星)。氣候/重力/礦產的**中文名**沒有官方在地化來源可援引——既有 i18n TSV 與先前的
// 中文化專案(~/master-of-orion)都沒有 MOO2 這幾組列舉的定案譯名,故用簡明直譯頂著顯示,
// 不是官方在地化文本。**英文名**則直接取原版手冊用語,不是回譯。
// 純展示層查表,不影響 engine/gamedata 任何邏輯或數值。
// 四組列舉的中英名。英文是原版手冊的用語;中文是簡明直譯(見上方註解:沒有官方定案譯名
// 可援引)。擺成 zh/en 成對的表而不是兩份 slice——加一種氣候時漏填英文會在這裡看得見。
var (
	climateNames = [...][2]string{
		{"有毒", "Toxic"}, {"輻射", "Radiated"}, {"貧瘠", "Barren"}, {"沙漠", "Desert"},
		{"凍原", "Tundra"}, {"海洋", "Ocean"}, {"沼澤", "Swamp"}, {"乾旱", "Arid"},
		{"類地", "Terran"}, {"蓋亞", "Gaia"},
	}
	mineralNames = [...][2]string{
		{"極貧礦", "Ultra Poor"}, {"貧礦", "Poor"}, {"普通", "Abundant"},
		{"富礦", "Rich"}, {"極富礦", "Ultra Rich"},
	}
	planetSizeNames = [...][2]string{
		{"極小", "Tiny"}, {"小型", "Small"}, {"中型", "Medium"},
		{"大型", "Large"}, {"巨型", "Huge"},
	}
	gravityNames = [...][2]string{
		{"低重力", "Low-G"}, {"標準", "Normal-G"}, {"高重力", "Heavy-G"},
	}
	unknownName = [2]string{"未知", "Unknown"}
)

// pickName 依語言挑一組 zh/en。
func pickName(lang i18n.Lang, pair [2]string) string {
	if lang == i18n.Traditional {
		return pair[0]
	}
	return pair[1]
}

func climateName(lang i18n.Lang, c gamedata.PlanetClimate) string {
	if int(c) >= 0 && int(c) < len(climateNames) {
		return pickName(lang, climateNames[c])
	}
	return pickName(lang, unknownName)
}

func gravityName(lang i18n.Lang, g gamedata.PlanetGravity) string {
	switch g {
	case gamedata.LOW_G:
		return pickName(lang, gravityNames[0])
	case gamedata.NORMAL_G:
		return pickName(lang, gravityNames[1])
	case gamedata.HEAVY_G:
		return pickName(lang, gravityNames[2])
	}
	return pickName(lang, unknownName)
}

func mineralsName(lang i18n.Lang, m gamedata.PlanetMinerals) string {
	if int(m) >= 0 && int(m) < len(mineralNames) {
		return pickName(lang, mineralNames[m])
	}
	return pickName(lang, unknownName)
}

func planetSizeName(lang i18n.Lang, sz gamedata.PlanetSize) string {
	if int(sz) >= 0 && int(sz) < len(planetSizeNames) {
		return pickName(lang, planetSizeNames[sz])
	}
	return pickName(lang, unknownName)
}

// shipClassLabel 把 shell 的艦體 key(中文)換成該語言要顯示的名字。
// 英文用原版艦體名(dsHullOrder,順序一致)。
func shipClassLabel(lang i18n.Lang, zhKey string) string {
	if lang == i18n.Traditional {
		return zhKey
	}
	for i, k := range shipClassZH {
		if k == zhKey && i < len(dsHullOrder) {
			return dsHullOrder[i]
		}
	}
	return zhKey
}

// truncateToWidth 把 s 截到在 fnt/size 下量測寬度不超過 maxW,超過則去尾加「…」。
// 用於欄寬有限的清單文字(如殖民地「已建:…」),避免溢出 cell 框。fnt 為 nil 或本就
// 不超寬時原樣回傳。逐 rune 縮短(CJK 無空白),保證至少留 1 個字 + 省略號。
func truncateToWidth(fnt *uifont.Font, s string, size, maxW float64) string {
	if fnt == nil {
		return s
	}
	if w, _ := fnt.Measure(s, size); w <= maxW {
		return s
	}
	rs := []rune(s)
	for len(rs) > 1 {
		rs = rs[:len(rs)-1]
		cand := string(rs) + "…"
		if w, _ := fnt.Measure(cand, size); w <= maxW {
			return cand
		}
	}
	return string(rs)
}

func (b *sceneBuilder) colonySummary() (*overlayScreen, error) {
	// 點各殖民地的職務欄 → 重分配 1 名人口(農夫欄→多農夫、工人欄→多工人、科學家欄→多科學家);
	// RETURN → 星系主畫面。列中心 y 與欄 x 對齊資料。
	rowY := []float64{47, 78, 109, 140, 171, 202, 233, 264, 295}
	hits := []hitRegion{{582, 452, 52, 20, "return"}}
	if b.session != nil {
		for i := range b.session.PlayerColonies {
			if i >= len(rowY) {
				break
			}
			top := int(rowY[i]) - 15
			hits = append(hits,
				// 殖民地名欄 → 進入單一殖民地畫面(原版 Main_Screen_ → Do_Colony_Screen_,
				// COLONY 與 COLONY SUMMARY 是兩個不同畫面,見 colonyscreen.go 檔頭)。
				hitRegion{18, top, 78, 30, fmt.Sprintf("n%d", i)},
				hitRegion{104, top, 118, 30, fmt.Sprintf("f%d", i)},
				hitRegion{236, top, 128, 30, fmt.Sprintf("w%d", i)},
				hitRegion{376, top, 128, 30, fmt.Sprintf("s%d", i)},
				hitRegion{510, top, 120, 30, fmt.Sprintf("b%d", i)}, // 建造欄
			)
		}
	}
	onAction := func(a string) *origTransition {
		if a == "return" {
			return b.goTo(b.galaxy, "星系主畫面")
		}
		if len(a) == 2 && b.session != nil {
			idx := int(a[1] - '0')
			switch a[0] {
			case 'f':
				b.session.ShiftColonyJob(idx, "w", "f") // 工人→農夫
			case 'w':
				b.session.ShiftColonyJob(idx, "f", "w") // 農夫→工人
			case 's':
				b.session.ShiftColonyJob(idx, "w", "s") // 工人→科學家
			case 'b':
				b.session.CycleColonyBuild(idx) // 循環建造項目
			case 'n':
				b.colonyIdx = idx
				b.colonyListTop = 0
				return b.goTo(b.colonyScreen, "殖民地")
			}
			return b.goTo(b.colonySummary, "殖民地總覽") // 重繪顯示新分配
		}
		return nil
	}
	// 欄位標題(上)+ 排序列(下)擦底疊字。座標經 PIL 量測。
	overlays := []labelRect{
		{18, 10, 78, 20, "NAME", 0},
		{104, 10, 118, 20, "FARMERS", 0},
		{236, 10, 128, 20, "WORKERS", 0},
		{376, 10, 128, 20, "SCIENTISTS", 0},
		{512, 10, 118, 20, "BUILDING", 0},
		{8, 452, 62, 20, "SORT", 0},
		{78, 452, 66, 18, "Name", 0},
		{150, 452, 92, 18, "Population", 0},
		{248, 452, 54, 18, "Food", 0},
		{306, 452, 74, 18, "Industry", 0},
		{384, 452, 74, 18, "Science", 0},
		{462, 452, 88, 18, "Producing", 0},
		{550, 452, 28, 18, "BC", 0},
		{582, 452, 52, 20, "RETURN", 0},
	}
	s, err := loadOverlayScreen(b.res, "colsum.lbx", 0, b.lang, b.fnt, "assets/i18n/colony.tsv",
		overlays, color.RGBA{210, 216, 230, 255}, 13, hits, onAction, nil)
	if err != nil {
		return nil, err
	}
	// 底列/欄名皆浮雕鈕:擦底色改採按鈕面色,避免暗塊蓋掉浮雕框(見 samplePlate faceSample)。
	s.plateFace = true
	// 即時殖民地資料填進表格列(欄位中心 x 對齊標題;列中心 y 經 PIL 量測,每列約 31px)。
	if b.session != nil {
		body := color.RGBA{214, 220, 235, 255}
		rowY := []float64{47, 78, 109, 140, 171, 202, 233, 264, 295}
		colX := struct{ name, far, wrk, sci float64 }{57, 163, 300, 440}
		for i, c := range b.session.PlayerColonies {
			if i >= len(rowY) {
				break
			}
			y := rowY[i]
			s.extras = append(s.extras,
				extraText{x: colX.name, y: y, size: 13, text: fmt.Sprintf(b.tr("殖民地 %d", "Colony %d"), i+1), col: body, align: 1},
				extraText{x: colX.far, y: y, size: 13, text: fmt.Sprintf("%d", c.Farmers), col: body, align: 1},
				extraText{x: colX.wrk, y: y, size: 13, text: fmt.Sprintf("%d", c.Workers), col: body, align: 1},
				extraText{x: colX.sci, y: y, size: 13, text: fmt.Sprintf("%d", c.Scientists), col: body, align: 1},
			)
			// 建造欄:項目名 + 進度(空則顯示「—」提示可點)。
			bt := "—"
			if i < len(b.session.Builds) && b.session.Builds[i].Name != "" {
				bd := b.session.Builds[i]
				bt = fmt.Sprintf("%s %d/%d", bd.Name, bd.Progress, bd.Cost)
			}
			if n := b.session.BuildQueueBacklogLen(i); n > 0 {
				bt += fmt.Sprintf(" +%d", n) // 佇列還排著 n 項(原版 7 格 BUILD QUEUE)
			}
			s.extras = append(s.extras, extraText{x: 571, y: y, size: 12, text: bt, col: body, align: 1})
			// 已建建築(顯示效果來源):在建造欄下方以小字列出。
			if i < len(b.session.ColonyBuildings) && len(b.session.ColonyBuildings[i]) > 0 {
				names := make([]string, 0, len(b.session.ColonyBuildings[i]))
				for n := range b.session.ColonyBuildings[i] {
					names = append(names, n)
				}
				sort.Strings(names)
				// 依「建造」欄寬截斷,避免長建築清單溢出 cell 框、撞出畫面右緣
				// (點陣字最小 12px,小字撐大更易超框;BUILDING 欄 x512 寬118,留邊取 110)。
				lbl := truncateToWidth(b.fnt, b.tr("已建:", "Built: ")+strings.Join(names, b.tr("、", ", ")), 10, 110)
				s.extras = append(s.extras, extraText{x: 571, y: y + 13, size: 10, text: lbl, col: color.RGBA{150, 200, 150, 255}, align: 1})
			}
		}
		// 第 4 格「Empire Summary」面板(x≈516-628,y≈349-438;見上方函式註解的
		// manual 查證):原本是空黑格,補上既有 session 欄位算好的帝國概況文字
		// (國庫/收支/人口/食物/研究),不新增邏輯、只讀既有資料。
		pop := 0
		for _, c := range b.session.PlayerColonies {
			pop += c.Population
		}
		out := b.session.LastPlayerOutput
		es := color.RGBA{205, 218, 235, 255}
		lines := []string{
			fmt.Sprintf(b.tr("國庫 %d BC", "Treasury %d BC"), b.session.Player.BC),
			fmt.Sprintf(b.tr("收支 %+d/回合", "Net %+d/turn"), out.NetBC),
			fmt.Sprintf(b.tr("人口 %d", "Pop %d"), pop),
			fmt.Sprintf(b.tr("食物 %+d", "Food %+d"), out.TotalFood),
			fmt.Sprintf(b.tr("研究 %d/回合", "Res %d/turn"), out.TotalResearch),
		}
		for i, l := range lines {
			s.extras = append(s.extras, extraText{x: 522, y: 360 + float64(i)*16, size: 11, text: l, col: es})
		}

		// --- Planetary Info(第1格,x10-92)+ Production Info(第2格,x102-371) ---
		// 座標為 PIL 逐像素量測 COLSUM.LBX 資產0既有黑格邊界所得(兩格與 Mini Map/Empire
		// Summary 同列,y349-438)。原版是「靜態黑底 + 執行期依滑鼠懸停的殖民地動態疊圖」,
		// 故用 postDraw(在 s.mx/s.my 之上,每幀依當下滑鼠位置重繪)取代固定 extras,才能
		// 隨懸停即時換內容;無懸停落在殖民地列時預設顯示殖民地 0(母星)。
		s.postDraw = func(dst *ebiten.Image) {
			if len(b.session.PlayerColonies) == 0 {
				return
			}
			ox, oy := float64(s.offsetX), float64(s.offsetY)
			idx := 0
			for i, ry := range rowY {
				if i >= len(b.session.PlayerColonies) {
					break
				}
				top, bottom := ry-15, ry+16
				if float64(s.my) >= top && float64(s.my) < bottom && s.mx >= 10 && s.mx < 628 {
					idx = i
					break
				}
			}
			if idx >= len(b.session.PlayerColonies) {
				idx = 0
			}
			c := b.session.PlayerColonies[idx]
			label := color.RGBA{225, 232, 245, 255}

			// Planetary Info:窄格(82px),短標籤 + 值同行。
			piLines := []string{
				fmt.Sprintf(b.tr("殖民地%d", "Colony %d"), idx+1),
				b.tr("氣候", "Cli ") + climateName(b.lang, c.Climate),
				b.tr("重力", "Grav ") + gravityName(b.lang, c.PlanetGravity),
				b.tr("礦產", "Min ") + mineralsName(b.lang, c.MineralRichness),
				b.tr("大小", "Size ") + planetSizeName(b.lang, c.PlanetSize),
				fmt.Sprintf(b.tr("上限%d", "Max %d"), c.PopMax),
			}
			for i, l := range piLines {
				s.font.Draw(dst, l, 14+ox, 358+float64(i)*14+oy, 10, label)
			}

			// Production Info:較寬(269px)。優先用 LastPlayerOutput.Colonies[idx](當回合已
			// 結算的實際產出);取不到(如新遊戲尚未跑過第一回合)時退回用 PlayerColonies
			// 欄位 × 人數的簡化估算,並標「約」避免誤當精確結算值。
			var prodLines []string
			if idx < len(b.session.LastPlayerOutput.Colonies) {
				co := b.session.LastPlayerOutput.Colonies[idx]
				prodLines = []string{
					fmt.Sprintf(b.tr("食物 產%d 耗%d 盈虧%+d", "Food +%d -%d net %+d"), co.Food, co.FoodConsumed, co.FoodSurplus),
					fmt.Sprintf(b.tr("工業 毛%d 淨%d", "Industry gross %d net %d"), co.GrossIndustry, co.NetIndustry),
					fmt.Sprintf(b.tr("研究 %d/回合", "Research %d/turn"), co.Research),
					fmt.Sprintf(b.tr("污染清理耗產能 %d", "Pollution cleanup costs %d"), co.PollutionCleanupCost),
				}
				if co.Starving {
					prodLines = append(prodLines, b.tr("缺糧中(饑荒)", "STARVING"))
				}
			} else {
				prodLines = []string{
					fmt.Sprintf(b.tr("食物(約) %d", "Food (est.) %d"), c.Farmers*c.FoodPerFarmer),
					fmt.Sprintf(b.tr("工業(約) %d", "Industry (est.) %d"), c.Workers*c.IndustryPerWorker),
					fmt.Sprintf(b.tr("研究(約) %d", "Research (est.) %d"), c.Scientists*c.ResearchPerScientist),
				}
			}
			for i, l := range prodLines {
				s.font.Draw(dst, l, 110+ox, 360+float64(i)*16+oy, 11, label)
			}
		}
	}
	return s, nil
}

// races 建原版種族關係畫面(RACES.LBX 資產 0,自帶完整調色盤)。RACES 按鈕目標。
func (b *sceneBuilder) races() (*overlayScreen, error) {
	// 「會晤」→ 銀河議會;「宣戰」→ 解算戰鬥;他處 → 星系主畫面。
	hits := []hitRegion{
		{340, 418, 96, 20, "audience"},
		{340, 438, 96, 20, "declarewar"},
		{438, 418, 90, 20, "report"},
		// 精確 RETURN 熱區(對齊 RETURN overlay {536,432,82,22};取代整畫面返回,僅返回鍵返回,
		// 與 openorion2-backed 畫面一致)。races 在 openorion2 是 STUB 無硬編座標,故用 PIL 量測的
		// overlay 位置當熱區來源(擦底疊字位置≈按鈕位置)。
		{536, 428, 82, 26, "back"},
	}
	onAction := func(a string) *origTransition {
		switch a {
		case "audience":
			return b.goTo(b.council, "銀河議會")
		case "report":
			sc, err := b.diplomacy() // 外交對談
			if err != nil {
				fmt.Fprintln(os.Stderr, "外交:", err)
				return nil
			}
			return &origTransition{next: sc}
		case "declarewar":
			sc, err := b.tacticalCombat() // 進格子戰術戰鬥
			if err != nil {
				fmt.Fprintln(os.Stderr, "進入戰鬥:", err)
				return nil
			}
			return &origTransition{next: sc}
		}
		return b.goTo(b.galaxy, "星系主畫面")
	}
	// 座標經 PIL 量測(remain-scan/races_a0_f00.png)。
	overlays := []labelRect{
		{200, 14, 240, 22, "RACE RELATIONS", 0},
		{338, 401, 104, 18, "BONUSES", 12},
		{340, 424, 96, 18, "AUDIENCE", 11},
		{340, 442, 96, 18, "DECLARE WAR", 10},
		{438, 424, 90, 18, "REPORT", 11},
		{438, 442, 90, 18, "IGNORE", 11},
		{536, 432, 82, 22, "RETURN", 0},
	}
	s, err := loadOverlayScreen(b.res, "races.lbx", 0, b.lang, b.fnt, "assets/i18n/diplo.tsv",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction, nil)
	if err != nil {
		return nil, err
	}
	// AI 對手即時狀態(名/態勢/軍力/佔星),讓 AI 主動行為可見。
	if b.session != nil && b.fnt != nil {
		gold := color.RGBA{240, 220, 120, 255}
		body := color.RGBA{210, 216, 230, 255}
		dim := color.RGBA{170, 178, 195, 255}
		y := 66.0
		for i, a := range b.session.AIPlayers {
			s.extras = append(s.extras,
				extraText{x: 40, y: y, size: 15, text: a.Name, col: gold},
				extraText{x: 40, y: y + 20, size: 12, text: fmt.Sprintf(b.tr("對你:%s ／ 軍力 %d ／ 佔領 %d 星", "Toward you: %s / power %d / %d systems"),
					a.StanceName, a.FleetStrength, a.OwnedStars), col: body},
			)
			// AI 之間的外交關係(活星系;支撐議會第三方搖擺)。
			rel := ""
			for j := range b.session.AIPlayers {
				if j == i {
					continue
				}
				if rel != "" {
					rel += "、"
				}
				rel += fmt.Sprintf("%s:%s", b.session.AIPlayers[j].Name, b.session.AIRelationName(i, j))
			}
			if rel != "" {
				s.extras = append(s.extras, extraText{x: 40, y: y + 38, size: 10, text: b.tr("對他國 ", "Toward others: ") + rel, col: dim})
			}
			y += 62
		}
	}
	return s, nil
}

// --- 外交對談畫面(用原版 DIPLOMAT 使節房 + 逐族使節疊合)---
//
// DIPLOMAT.LBX 佈局(2026-07-10 破解,見 docs/tech/diplomat-lbx-layout.md):
//   asset 0–12    :24×24 內嵌調色盤,13 個(各族專屬 palette)。
//   asset 13+2r   :640×480 使節房背景(種族 r,r=0..12)。
//   asset 14+2r   :480×480 FLAG_JUNCTION 使節動畫(種族 r,含使節像 + 廊柱)。
// 配對律:種族 r 的房/使節/調色盤都用同一個 r。房或使節借錯 palette 才會全畫面雜點。

// diplomatRaceIndex 把敵方種族名對應到 DIPLOMAT.LBX 的種族序 r(0..12)。
// 13 族皆已對 RACESEL 肖像逐一核實對應,見 docs/tech/diplomat-lbx-layout.md。
func diplomatRaceIndex(enemy string) int {
	// raceSelectList 本來就是原版字母序(對齊 RACESEL 肖像 15..28),而 DIPLOMAT 的種族序
	// 也是同一套字母序,所以直接拿它推導,不必再手寫一份 14 case 的對照 switch
	// (兩份表遲早會漂移)。中英名都比對:英文模式下 PrimaryEnemyName 有可能回英文名。
	for i, e := range raceSelectList {
		if e.shellIdx < 0 {
			continue // Custom 沒有使節像
		}
		if enemy == e.zh || enemy == e.en || enemy == e.enAdj {
			return i
		}
	}
	// 認不出來就退薩克拉(索引 10)——舊存檔可能帶著已淘汰的字串(如「賽隆人」)。
	return 10
}

// loadDiplomatScene 疊合種族 r 的使節房(640×480 背景)+ 使節動畫(480×480,置中),
// 兩者都用同一個 palette provider r(配對律)。使節 sprite 的未寫入邊緣為透明,疊上後
// 房間從邊緣透出,中央被使節像覆蓋——即原版外交畫面構圖。
func loadDiplomatScene(res *assets.Resolver, r int) *ebiten.Image {
	prov, err := decodeAsset(res, "diplomat.lbx", r) // 該族專屬調色盤
	if err != nil || prov.Embedded == nil {
		return nil
	}
	room, err := decodeAsset(res, "diplomat.lbx", 13+2*r)
	if err != nil || len(room.Frames) == 0 {
		return nil
	}
	scene := ebiten.NewImageFromImage(room.Frames[0].ToRGBA(prov.Embedded, room.KeyColor()))
	// 使節 sprite 疊上(480 寬置中於 640)。
	if envoy, err := decodeAsset(res, "diplomat.lbx", 14+2*r); err == nil && len(envoy.Frames) > 0 {
		esprite := ebiten.NewImageFromImage(envoy.Frames[0].ToRGBA(prov.Embedded, true))
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64((room.Width-envoy.Width)/2), 0)
		scene.DrawImage(esprite, op)
	}
	return scene
}

type diplomacyScreen struct {
	b        *sceneBuilder
	fnt      *uifont.Font
	enemy    string
	response string
	room     *ebiten.Image // 原版 DIPLOMAT 使節房 + 使節疊合
	opts     []struct {
		label, action string
	}
	backRect [4]int
}

func newDiplomacyScreen(b *sceneBuilder) *diplomacyScreen {
	// 對談對象改用真正的主要 AI 對手(races 畫面第一個),使外交動作實際改變其關係、可見於
	// 態勢/議會;取不到 session 時退回示範名。
	enemy := b.tr("薩克拉", "Sakkra") // 取不到 session 時的示範名
	if b.session != nil {
		enemy = b.session.PrimaryEnemyName()
	}
	return &diplomacyScreen{b: b, fnt: b.fnt, enemy: enemy, room: loadDiplomatScene(b.res, diplomatRaceIndex(enemy)),
		response: enemy + b.tr("使節:人類,你有何提議?", " emissary: Human, what do you propose?"),
		opts: []struct{ label, action string }{
			{"提議和平", "peace"}, {"提議貿易", "trade"}, {"威脅恫嚇", "threat"},
		},
		backRect: [4]int{250, 420, 140, 34}}
}

func (d *diplomacyScreen) optRect(i int) (x, y, w, h int) { return 190, 150 + i*54, 260, 40 }

func (d *diplomacyScreen) update(in shell.InputState) *origTransition {
	if !in.ClickReleased {
		return nil
	}
	for i, o := range d.opts {
		x, y, w, h := d.optRect(i)
		if in.MouseX >= x && in.MouseX < x+w && in.MouseY >= y && in.MouseY < y+h {
			d.response = d.b.session.DiplomacyResponse(o.action, d.enemy)
			return nil
		}
	}
	bx, by, bw, bh := d.backRect[0], d.backRect[1], d.backRect[2], d.backRect[3]
	if in.MouseX >= bx && in.MouseX < bx+bw && in.MouseY >= by && in.MouseY < by+bh {
		return d.b.goTo(d.b.races, "種族關係")
	}
	return nil
}

func (d *diplomacyScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{12, 10, 22, 255})
	if d.room != nil { // 原版議事廳背景
		dst.DrawImage(d.room, nil)
	}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{235, 232, 245, 255}
	if d.fnt == nil {
		return
	}
	// 上方標題 + 使節台詞(疊半透明深色條增可讀性)。
	vector.DrawFilledRect(dst, 0, 44, moo2ScreenW, 92, color.RGBA{8, 6, 14, 180}, false)
	d.fnt.DrawCentered(dst, d.b.tr("外交對談", "AUDIENCE"), 320, 62, 20, gold)
	d.fnt.DrawCentered(dst, d.enemy+d.b.tr(" 使節", " Emissary"), 320, 96, 14, color.RGBA{235, 150, 140, 255})
	d.fnt.DrawCentered(dst, d.response, 320, 124, 14, body)
	for i, o := range d.opts {
		x, y, w, h := d.optRect(i)
		vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{34, 30, 54, 255}, false)
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1.5, color.RGBA{110, 90, 160, 255}, false)
		d.fnt.DrawCentered(dst, o.label, float64(x+w/2), float64(y+h/2), 15, body)
	}
	bx, by, bw, bh := d.backRect[0], d.backRect[1], d.backRect[2], d.backRect[3]
	vector.DrawFilledRect(dst, float32(bx), float32(by), float32(bw), float32(bh), color.RGBA{40, 34, 30, 255}, false)
	vector.StrokeRect(dst, float32(bx), float32(by), float32(bw), float32(bh), 1.5, color.RGBA{160, 140, 100, 255}, false)
	d.fnt.DrawCentered(dst, d.b.tr("結束對談", "END AUDIENCE"), float64(bx+bw/2), float64(by+bh/2), 15, body)
}

// diplomacy 進入外交對談畫面(對象是主要對手)。
func (b *sceneBuilder) diplomacy() (origScreen, error) {
	if b.session == nil {
		return nil, fmt.Errorf("無對局")
	}
	playSceneBGM(bgmDiplo)
	return newDiplomacyScreen(b), nil
}

// diplomacyWith 進入外交對談畫面,對象指定。
//
// 由星圖上緣的會談請求燈用(見 audience.go)——是**那位**對手來敲門,不是主要對手。
func (b *sceneBuilder) diplomacyWith(enemy string) (origScreen, error) {
	if b.session == nil {
		return nil, fmt.Errorf("無對局")
	}
	playSceneBGM(bgmDiplo)
	d := newDiplomacyScreen(b)
	if enemy != "" {
		d.enemy = enemy
		d.room = loadDiplomatScene(b.res, diplomatRaceIndex(enemy))
		d.response = enemy + b.tr("使節:人類,你有何提議?", " emissary: Human, what do you propose?")
	}
	return d, nil
}

// --- 格子戰術戰鬥畫面(自繪 origScreen:星空底 + 格線 + 雙方艦艇 token + HP 條)---

// 戰場格子:8 欄 × 6 列。
const (
	gcX0, gcY0     = 40, 70
	gcCols, gcRows = 8, 6
	gcCW, gcCH     = 70, 55
	fireRange      = 4 // 曼哈頓射程
)

type tacticalScreen struct {
	b              *sceneBuilder
	fnt            *uifont.Font
	player, enemy  []shell.CombatShip
	sel            int // 選中的我方艦索引(-1=無)
	round          int
	log            string
	over, won      bool
	pStart, eStart int
	rng            *rand.Rand // 戰鬥擲骰(依回合數種子,可重現)
	bg             *ebiten.Image
	bar            *ebiten.Image
	res            *assets.Resolver      // 供 shipSprite 延遲載入各艦級 sprite
	shipSprites    map[int]*ebiten.Image // CMBTSHP 資產索引 → 已解碼 sprite(nil=載入失敗,亦快取)
}

// loadCombatBG 載入戰場星空背景(STARBG.LBX#0,640×480),借 COMBAT.LBX#11 調色盤。
// STARBG 是稀疏 RLE(大量未寫入像素),原版設計疊在純黑太空上,故未寫入處回傳透明,
// 由呼叫端鋪在黑底上即為正確畫面(見任務交接的 de-risk 事實)。載入失敗回傳 nil,
// 由 draw() fallback 回原本純色 + 格線。
func loadCombatBG(res *assets.Resolver) *ebiten.Image {
	prov, err := decodeAsset(res, "combat.lbx", 11)
	if err != nil || prov.Embedded == nil {
		return nil
	}
	im, err := decodeAsset(res, "starbg.lbx", 0)
	if err != nil || len(im.Frames) == 0 {
		return nil
	}
	return ebiten.NewImageFromImage(im.Frames[0].ToRGBA(prov.Embedded, im.KeyColor()))
}

// loadCombatBar 載入戰鬥畫面底部控制列(COMBAT.LBX#0,640×129),同借 COMBAT#11 調色盤。
func loadCombatBar(res *assets.Resolver) *ebiten.Image {
	prov, err := decodeAsset(res, "combat.lbx", 11)
	if err != nil || prov.Embedded == nil {
		return nil
	}
	im, err := decodeAsset(res, "combat.lbx", 0)
	if err != nil || len(im.Frames) == 0 {
		return nil
	}
	return ebiten.NewImageFromImage(im.Frames[0].ToRGBA(prov.Embedded, im.KeyColor()))
}

// loadCombatShipByIdx 載入 CMBTSHP.LBX 第 idx 個艦艇 sprite(frame0),用其所屬色塊的
// palette-holder(索引 45*(idx/45)+44,內嵌調色盤)上色。見 docs/tech/cmbtshp-ship-sprites.md。
// keyColor 用資產自身旗標(CMBTSHP flags=0x0000 → false):艦體外圍透明來自未寫入的
// RLE 像素(ToRGBA 一律留透明),而艦體本身含 index-0 深色像素須保留——先前誤設
// keyColor=true 會把 index-0 艦體也判成透明,導致 sprite 幾乎全消失(端到端截圖查出)。
func loadCombatShipByIdx(res *assets.Resolver, idx int) *ebiten.Image {
	palIdx := (idx/45)*45 + 44
	prov, err := decodeAsset(res, "cmbtshp.lbx", palIdx)
	if err != nil || prov.Embedded == nil {
		return nil
	}
	im, err := decodeAsset(res, "cmbtshp.lbx", idx)
	if err != nil || len(im.Frames) == 0 {
		return nil
	}
	return ebiten.NewImageFromImage(im.Frames[0].ToRGBA(prov.Embedded, im.KeyColor()))
}

func newTacticalScreen(b *sceneBuilder) *tacticalScreen {
	p, e := b.session.StartCombat(b.session.PrimaryEnemyName())
	// 戰鬥 RNG 依當前回合數種子:同一局同一回合的戰鬥可重現(不引入 wall-clock 不確定性)。
	seed := int64(b.session.Turn*2654435761 + 1013904223)
	return &tacticalScreen{b: b, fnt: b.fnt, player: p, enemy: e, sel: -1,
		log: b.tr("點我方艦選取→點空格移動;點敵艦→射程內我艦開火",
			"Click your ship to select, an empty cell to move, an enemy to fire"),
		pStart: len(p), eStart: len(e),
		rng: rand.New(rand.NewSource(seed)),
		bg:  loadCombatBG(b.res), bar: loadCombatBar(b.res),
		res: b.res, shipSprites: map[int]*ebiten.Image{}}
}

// shipSprite 依 CMBTSHP 資產索引取(並快取)已解碼 sprite,避免每幀重解。
func (t *tacticalScreen) shipSprite(idx int) *ebiten.Image {
	if im, ok := t.shipSprites[idx]; ok {
		return im
	}
	im := loadCombatShipByIdx(t.res, idx)
	t.shipSprites[idx] = im // 允許 nil(載入失敗),快取避免每幀重試
	return im
}

func cellRect(col, row int) (x, y, w, h int) { return gcX0 + col*gcCW, gcY0 + row*gcCH, gcCW, gcCH }

func cellAt(mx, my int) (col, row int, ok bool) {
	if mx < gcX0 || my < gcY0 || mx >= gcX0+gcCols*gcCW || my >= gcY0+gcRows*gcCH {
		return 0, 0, false
	}
	return (mx - gcX0) / gcCW, (my - gcY0) / gcCH, true
}

func shipAt(list []shell.CombatShip, col, row int) int {
	for i, s := range list {
		if s.Col == col && s.Row == row {
			return i
		}
	}
	return -1
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func (t *tacticalScreen) update(in shell.InputState) *origTransition {
	if !in.ClickReleased {
		return nil
	}
	if t.over { // 戰後點擊 → 套用結果 → 戰鬥結果畫面
		survivors := map[string]bool{}
		for _, s := range t.player {
			survivors[s.Name] = true
		}
		t.b.session.ApplyCombatOutcome(t.b.session.PrimaryEnemyName(), t.pStart, t.eStart, survivors, t.won)
		return t.b.goTo(t.b.battleResult, "戰鬥結果")
	}
	col, row, ok := cellAt(in.MouseX, in.MouseY)
	if !ok {
		return nil
	}
	if pi := shipAt(t.player, col, row); pi >= 0 { // 點我方艦 → 選取
		t.sel = pi
		return nil
	}
	if ei := shipAt(t.enemy, col, row); ei >= 0 { // 點敵艦 → 射程內我艦開火
		t.fireRound(ei)
		return nil
	}
	if t.sel >= 0 && t.sel < len(t.player) { // 點空格 → 移動選中艦
		t.player[t.sel].Col, t.player[t.sel].Row = col, row
		t.log = fmt.Sprintf(t.b.tr("%s 移動到 (%d,%d)", "%s moves to (%d,%d)"), t.player[t.sel].Name, col, row)
	}
	return nil
}

func (t *tacticalScreen) fireRound(target int) {
	tc, tr := t.enemy[target].Col, t.enemy[target].Row
	// 射程內我艦逐一依武器類型分流真戰鬥公式:beam(ResolveShot,不動)/missile
	// (ResolveMissileShot,躲避+AMR 攔截)/spherical(ResolveSphericalShot,現行武器表
	// 暫無掛載,分支保留供未來串接)。見 shell/weapon_kind.go 的分類依據。
	preCount := len(t.enemy) // 用來判斷本回合是否有敵艦被擊毀(播爆炸音效)
	pAtk, firing := 0, 0
	anyHit := false
	firedMissile := false // 首艘開火艦是否為飛彈類(決定開火音效)
	firedAny := false
	for i := range t.player {
		s := &t.player[i]
		dist := abs(s.Col-tc) + abs(s.Row-tr)
		if dist > fireRange {
			continue
		}
		firing++
		if !firedAny {
			firedAny = true
			firedMissile = s.Kind == shell.WeaponKindMissile
		}
		enemy := &t.enemy[target]
		var shot shell.ShotResult
		switch s.Kind {
		case shell.WeaponKindMissile:
			amrRoll := t.rng.Intn(100) + 1
			jamRoll := t.rng.Intn(100) + 1
			// hasAMR/evasion 加成現行皆無對應可造艦元件,保守傳 0/false(見
			// shell.ResolveMissileShot 註解的 TODO);dist 是實際格距離(比 battleVolley
			// 固定 range=2 更忠實)。
			shot = shell.ResolveMissileShot(false, dist, amrRoll, 0, 0, false, jamRoll,
				s.WeaponMax, enemy.ShieldReduction, enemy.ArmorHP, false)
		case shell.WeaponKindSpherical:
			span := s.WeaponMax - s.WeaponMin
			r := 0
			if span > 0 {
				r = t.rng.Intn(span + 1)
			}
			aggD := gamedata.DamageSphericalRoll(s.WeaponMin, r, 100)
			shot = shell.ResolveSphericalShot(aggD, enemy.ShieldReduction, enemy.ArmorHP, false, false)
		default:
			roll := t.rng.Intn(100) + 1
			net := s.Attack - enemy.Defense
			shot = shell.ResolveShotWithMods(net, s.WeaponMin, s.WeaponMax, dist,
				enemy.ShieldReduction, enemy.ArmorHP, roll, false, shell.WeaponModCodesFromStrings(s.Mods))
		}
		if shot.Hit {
			anyHit = true
			enemy.ArmorHP = shot.RemainingArmorHP
			enemy.HP -= shot.DamageToStructure
			pAtk += shot.DamageToStructure
		}
	}
	if firing == 0 {
		t.log = t.b.tr("目標超出射程,移動艦艇靠近再開火", "Target out of range — move closer before firing")
		return
	}
	t.round++
	alive := t.enemy[:0]
	for _, s := range t.enemy {
		if s.HP > 0 {
			alive = append(alive, s)
		}
	}
	t.enemy = alive
	// 戰鬥音效(SOUND.LBX 現成音效,headless / 缺音效時閉包為 nil):開火(依武器類型)→
	// 擊毀播爆炸、否則命中播命中音。見 audiohook.go sfx* 閉包。
	playSFX(fireSFX(firedMissile))
	if len(t.enemy) < preCount {
		playSFX(sfxExplode)
	} else if anyHit {
		playSFX(sfxHit)
	}
	// 敵方還擊我方最脆弱艦(同樣走真戰鬥公式,每艦一發)。
	eAtk := 0
	if len(t.player) > 0 && len(t.enemy) > 0 {
		wi := 0
		for i := range t.player {
			if t.player[i].HP < t.player[wi].HP {
				wi = i
			}
		}
		for i := range t.enemy {
			es := &t.enemy[i]
			dist := abs(es.Col-t.player[wi].Col) + abs(es.Row-t.player[wi].Row)
			if dist > fireRange {
				continue
			}
			// 敵艦(genEnemyFleet)沒有個別武器設計,es.Kind 恆為 WeaponKindBeam(既有
			// 簡化,非本輪引入),故還擊固定走 beam 路徑,不需要分流。
			roll := t.rng.Intn(100) + 1
			net := es.Attack - t.player[wi].Defense
			shot := shell.ResolveShot(net, es.WeaponMin, es.WeaponMax, dist,
				t.player[wi].ShieldReduction, t.player[wi].ArmorHP, roll,
				t.player[wi].HardShield, false)
			if shot.Hit {
				t.player[wi].ArmorHP = shot.RemainingArmorHP
				t.player[wi].HP -= shot.DamageToStructure
				eAtk += shot.DamageToStructure
			}
		}
	}
	palive := t.player[:0]
	for _, s := range t.player {
		if s.HP > 0 {
			palive = append(palive, s)
		}
	}
	t.player = palive
	if t.sel >= len(t.player) {
		t.sel = -1
	}
	t.log = fmt.Sprintf(t.b.tr("第 %d 回合:%d 艦齊射 %d ／ 敵方還擊 %d",
		"Round %d: %d ships deal %d / enemy returns %d"), t.round, firing, pAtk, eAtk)
	if len(t.enemy) == 0 {
		t.over, t.won, t.log = true, true, t.b.tr("★ 敵艦隊全滅,勝利!點擊繼續",
			"★ Enemy fleet destroyed — victory! Click to continue")
	} else if len(t.player) == 0 {
		t.over, t.won, t.log = true, false, t.b.tr("✗ 我方艦隊全滅,敗北。點擊繼續",
			"✗ Your fleet is destroyed — defeat. Click to continue")
	}
}

// drawShip 畫單艘艦:依 s.SpriteIdx 取該艦級的 CMBTSHP sprite 就縮放貼原版艦圖
// (敵方水平翻轉朝左),否則 fallback 回原本的矩形 token 畫法。HP 條、艦名、選中金框
// 一律疊在最上層,不受美術是否載入影響。
//
// 2026-07-11 修疊字 bug:原本圖示等比縮放頂滿整格高度,艦名疊在圖示正中央(y+13 恰好落在
// 圖示範圍內),兩者互相蓋字難辨(端到端截圖查出)。改成上→下三段式版面:艦名帶(固定於格
// 頂、半透明黑底墊字)→ 圖示(縮小置中,讓開文字帶與血條)→ HP 條(格底),彼此不重疊。
func (t *tacticalScreen) drawShip(dst *ebiten.Image, s shell.CombatShip, base color.RGBA, selected bool, enemy bool) {
	x, y, w, h := cellRect(s.Col, s.Row)
	x, y, w, h = x+4, y+6, w-8, h-12
	const labelH = 13 // 艦名帶高度(固定在格頂)
	const hpH = 8     // 血條預留高度(固定在格底)
	iconTop := y + labelH
	iconH := h - labelH - hpH
	if iconH < 4 {
		iconH = 4
	}
	if sprite := t.shipSprite(s.SpriteIdx); sprite != nil {
		sb := sprite.Bounds()
		sw0, sh0 := float64(sb.Dx()), float64(sb.Dy())
		sc := float64(iconH) / sh0 // 依縮小後的圖示高度等比縮放(不再頂滿整格)
		iconW := sw0 * sc
		iconX := float64(x) + (float64(w)-iconW)/2 // 水平置中於格內
		op := &ebiten.DrawImageOptions{}
		if enemy {
			op.GeoM.Scale(-sc, sc)
			op.GeoM.Translate(iconX+iconW, float64(iconTop))
		} else {
			op.GeoM.Scale(sc, sc)
			op.GeoM.Translate(iconX, float64(iconTop))
		}
		dst.DrawImage(sprite, op)
	} else {
		vector.DrawFilledRect(dst, float32(x), float32(iconTop), float32(w), float32(iconH), color.RGBA{base.R / 3, base.G / 3, base.B / 3, 255}, false)
	}
	sw := float32(1.5)
	sc := base
	if selected {
		sw, sc = 3, color.RGBA{255, 240, 120, 255}
	}
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), sw, sc, false)
	// 艦名帶:半透明黑底 + 文字,固定在格頂、圖示上方,不與圖示重疊。
	// 注意 uifont.Font.Draw 的 (x,y) 是文字「左上角」基準(非 baseline),故這裡從帶子頂端
	// 往下留 1px 起畫,讓字身整個落在 labelH 高度內,不溢出到下方圖示區(先前 y+labelH-2
	// 誤當 baseline 用,實際會把字往下推到圖示範圍,疊字 bug 未修好,端到端截圖二次查出)。
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(labelH), color.RGBA{0, 0, 0, 150}, false)
	if t.fnt != nil {
		t.fnt.Draw(dst, s.Name, float64(x)+3, float64(y)+1, 10, color.RGBA{235, 240, 250, 255})
	}
	frac := float32(s.HP) / float32(s.MaxHP)
	if frac < 0 {
		frac = 0
	}
	vector.DrawFilledRect(dst, float32(x)+5, float32(y)+float32(h)-8, float32(w-10), 4, color.RGBA{40, 40, 40, 255}, false)
	vector.DrawFilledRect(dst, float32(x)+5, float32(y)+float32(h)-8, (float32(w-10))*frac, 4, base, false)
}

func (t *tacticalScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{0, 0, 0, 255}) // 純黑太空底;STARBG 未寫入處透明,疊上後黑底透出即原版構圖
	if t.bg != nil {
		dst.DrawImage(t.bg, nil)
	} else {
		dst.Fill(color.RGBA{6, 6, 16, 255}) // fallback:原本深藍純色底
	}
	// 格線很淡地疊在星空上,保留移動格線功能但不搶戲。
	grid := color.RGBA{60, 80, 120, 40}
	for gx := 0; gx <= gcCols; gx++ {
		x := float32(gcX0 + gx*gcCW)
		vector.StrokeLine(dst, x, gcY0, x, float32(gcY0+gcRows*gcCH), 1, grid, false)
	}
	for gy := 0; gy <= gcRows; gy++ {
		y := float32(gcY0 + gy*gcCH)
		vector.StrokeLine(dst, gcX0, y, float32(gcX0+gcCols*gcCW), y, 1, grid, false)
	}
	gold := color.RGBA{240, 220, 120, 255}
	if t.fnt != nil {
		t.fnt.DrawCentered(dst, t.b.tr("戰術戰鬥", "TACTICAL COMBAT"), 320, 34, 20, gold)
	}
	for i, s := range t.player {
		t.drawShip(dst, s, color.RGBA{90, 220, 170, 255}, i == t.sel, false)
	}
	for _, s := range t.enemy {
		t.drawShip(dst, s, color.RGBA{235, 110, 100, 255}, false, true)
	}
	logY := 452.0
	if t.bar != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, float64(moo2ScreenH-129))
		dst.DrawImage(t.bar, op)
		// 控制列烘進的英文按鈕疊中文(CLAUDE.md:button 也要中文化)。
		// 英文模式跳過:COMBAT.LBX#0 上本來就是 AUTO / SCAN / BOARD / RETREAT / …。
		if t.b.lang == i18n.Traditional {
			t.drawBarLabelsCHT(dst)
		}
		logY = 343 // log 移到控制列上方星空,不壓按鈕
	}
	if t.fnt != nil {
		t.fnt.DrawCentered(dst, t.log, 320, logY, 14, color.RGBA{214, 220, 235, 255})
	}
}

// barButtonsCHT 是 COMBAT.LBX#0 控制列上各英文按鈕的螢幕中心座標 + 中文標籤。
// 座標於實際戰鬥截圖(gallery)量測;控制列貼在 y=moo2ScreenH-129=351。
// WEAPONS/SPECIALS 兩個欄位標頭在 remake 未用的清單面板內,略過。
//
// `orig` 是原版烘在鈕上的英文。放成欄位而不是行末註解,是因為它有用途:英文模式整段
// 不畫(讓原版美術露出來),而這一欄就是「露出來會是什麼字」的記錄,對得起來才敢讓路。
var barButtonsCHT = []struct {
	cx, cy int
	label  string
	orig   string
}{
	{302, 378, "自動", "AUTO"}, {373, 378, "掃描", "SCAN"},
	{302, 402, "登船", "BOARD"}, {373, 402, "撤退", "RETREAT"},
	{302, 433, "等待", "WAIT"}, {373, 433, "完成", "DONE"},
	{337, 461, "選項", "OPTIONS"},
}

// drawBarLabelsCHT 在原版控制列的英文按鈕上疊深色底 + 中文字,蓋掉烘進的英文。
func (t *tacticalScreen) drawBarLabelsCHT(dst *ebiten.Image) {
	if t.fnt == nil {
		return
	}
	for _, b := range barButtonsCHT {
		x, y := float32(b.cx-27), float32(b.cy-10)
		vector.DrawFilledRect(dst, x, y, 54, 20, color.RGBA{40, 44, 54, 255}, false)
		vector.StrokeRect(dst, x, y, 54, 20, 1, color.RGBA{120, 130, 150, 255}, false)
		t.fnt.DrawCentered(dst, b.label, float64(b.cx), float64(b.cy), 13, color.RGBA{225, 230, 240, 255})
	}
}

// tacticalCombat 進入格子戰術戰鬥畫面。
func (b *sceneBuilder) tacticalCombat() (origScreen, error) {
	playSceneBGM(bgmCombat)
	if b.session == nil {
		return nil, fmt.Errorf("無對局")
	}
	return newTacticalScreen(b), nil
}

// battleResult 顯示上一場戰鬥結果(重用 TURNSUM.LBX#0 視窗當通用面板)。點畫面返回種族關係。
func (b *sceneBuilder) battleResult() (*overlayScreen, error) {
	hits, onAction := b.backHit(b.races, "種族關係")
	// 標題以中文直接當 enKey(misc.tsv 查無 → fallback 回傳自身),擦底覆蓋烘進的 TURN SUMMARY。
	overlays := []labelRect{
		{88, 14, 204, 22, "戰鬥結果", 0},
		{158, 324, 64, 18, "CLOSE", 0},
	}
	s, err := loadOverlayScreen(b.res, "turnsum.lbx", 0, b.lang, b.fnt, "assets/i18n/misc.tsv",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"buffer0.lbx", 0}})
	if err != nil {
		return nil, err
	}
	if b.session != nil && b.session.LastBattle != nil {
		bt := b.session.LastBattle
		gold := color.RGBA{240, 220, 120, 255}
		body := color.RGBA{214, 220, 235, 255}
		win := color.RGBA{120, 220, 140, 255}
		lose := color.RGBA{235, 120, 110, 255}
		outcome, oc := b.tr("✗ 敗北", "✗ DEFEAT"), lose
		if bt.PlayerWon {
			outcome, oc = b.tr("★ 勝利!", "★ VICTORY!"), win
		}
		s.extras = []extraText{
			{x: 40, y: 56, size: 15, text: fmt.Sprintf(b.tr("對「%s」開戰", "Battle against %s"), bt.Enemy), col: gold},
			{x: 40, y: 84, size: 16, text: outcome, col: oc},
			{x: 40, y: 110, size: 12, text: fmt.Sprintf(b.tr("我方 %d 艦 ／ 敵方 %d 艦", "You %d ships / enemy %d ships"),
				bt.PlayerStart, bt.EnemyStart), col: body},
		}
		yy := 134.0
		for _, line := range bt.Log { // 逐回合戰報
			s.extras = append(s.extras, extraText{x: 40, y: yy, size: 12, text: line, col: body})
			yy += 20
		}
		s.extras = append(s.extras, extraText{x: 40, y: yy + 4, size: 13,
			text: fmt.Sprintf(b.tr("損失:我方 %d 艦 ／ 敵方 %d 艦", "Losses: you %d ships / enemy %d ships"),
				bt.PlayerLosses, bt.EnemyLosses), col: gold})
	}
	return s, nil
}

// council 建原版銀河議會畫面(COUNCIL.LBX 資產 1,調色盤鏈 COUNCIL#0)。3D 議事廳,
// 無烘字,疊「銀河議會」標題;點畫面返回種族關係。
func (b *sceneBuilder) council() (*overlayScreen, error) {
	// 有待回應選舉(AI 當選)時,改用「接受/拒絕」熱區——手冊:議會無法強迫玩家接受決議
	// (RespondToCouncilElection)。其餘狀態下整頁點擊返回種族關係(backHit)。原版議會是 3D
	// 議事廳、無內建 accept/reject 按鈕藝術,故此處以可點擊文字提示補上互動,不偽造浮雕按鈕框
	// (尊重「用原版 LBX、不自創按鈕藝術」;仍疊在原版 council.lbx 底圖上)。
	pending := b.session != nil && b.session.CouncilStatus().Pending != nil
	hits, onAction := b.backHit(b.races, "種族關係")
	if pending {
		hits = []hitRegion{
			{120, 402, 400, 26, "accept"},
			{120, 432, 400, 26, "reject"},
			{0, 0, moo2ScreenW, moo2ScreenH, "back"}, // 其餘處:不回應直接離開(pending 保留,可再進來)
		}
		onAction = func(a string) *origTransition {
			switch a {
			case "accept":
				b.session.RespondToCouncilElection(true) // 接受落敗 → 遊戲結束
				return b.goTo(b.galaxy, "星系主畫面")
			case "reject":
				b.session.RespondToCouncilElection(false) // 拒絕 → 清空待決,繼續遊戲
				return b.goTo(b.council, "銀河議會")          // 重繪議會,反映已回應
			}
			return b.goTo(b.races, "種族關係")
		}
	}
	s, err := loadOverlayScreen(b.res, "council.lbx", 1, b.lang, b.fnt, "assets/i18n/misc.tsv",
		nil, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"council.lbx", 0}})
	if err != nil {
		return nil, err
	}
	if b.fnt != nil {
		gold := color.RGBA{240, 220, 120, 255}
		s.extras = []extraText{
			{x: moo2ScreenW / 2, y: 30, size: 22, text: b.tr("銀河議會", "GALACTIC COUNCIL"), col: gold, align: 1},
		}
		if b.session != nil {
			// 在原版 council.lbx 底圖上,誠實呈現 shell.GameSession 已算好的議會狀態(gamedata/
			// council.go + shell/council.go,依 GAME_MANUAL.pdf p.183):是否已成立、逐帝國票數與
			// 搖擺票去向、是否已分出勝負/待回應。逐帝國明細取代舊的單行合計摘要,讓手冊的搖擺票
			// 機制(候選人+第三方依外交關係投票/棄權)在畫面上看得見。
			v := b.session.CouncilStatus()
			win := color.RGBA{120, 220, 140, 255}
			lose := color.RGBA{235, 120, 110, 255}
			neutral := color.RGBA{214, 220, 235, 255}
			var line1, line2 string
			var oc color.RGBA
			switch {
			case v.Victory.Over && v.Victory.Reason == engine.VictoryHighCouncil:
				line1 = fmt.Sprintf(b.tr("已於第 %d 回合分出勝負(共召開 %d 屆選舉)",
					"Decided on turn %d (after %d elections)"), v.Victory.Turn, v.Meetings)
				if v.Victory.Winner == "player" {
					line2, oc = b.tr("★ 你已當選銀河領袖!", "★ You have been elected Galactic Leader!"), win
				} else {
					line2, oc = v.Victory.Winner+b.tr(" 已當選銀河領袖", " has been elected Galactic Leader"), lose
				}
			case v.Pending != nil:
				// 待回應選舉:顯示當選 AI + 兩個可點擊選項(接受落敗 / 拒絕接受繼續遊戲),
				// 對應上方 pending 分支設定的 accept/reject 熱區。
				s.extras = append(s.extras,
					extraText{x: moo2ScreenW / 2, y: 300, size: 16, text: fmt.Sprintf(b.tr("第 %d 屆選舉:%s 以 %d/%d 票達2/3多數當選銀河領袖",
						"Election %d: %s takes the 2/3 majority with %d/%d votes"),
						v.Meetings, v.Pending.EnemyName, v.Pending.EnemyVotes, v.Pending.TotalVotes), col: lose, align: 1},
					extraText{x: moo2ScreenW / 2, y: 330, size: 14, text: b.tr("議會無法強迫你接受不同意的決議(手冊 p.183)——請抉擇:",
						"The council cannot force a decision on you (manual p.183) — choose:"), col: neutral, align: 1},
					extraText{x: moo2ScreenW / 2, y: 410, size: 18, text: b.tr("▶  接受落敗結果(遊戲結束)", "▶  Accept the outcome (game over)"), col: lose, align: 1},
					extraText{x: moo2ScreenW / 2, y: 440, size: 18, text: b.tr("▶  拒絕接受(繼續遊戲,下屆再選)",
						"▶  Refuse (play on; a new election follows)"), col: win, align: 1},
				)
				line1, line2 = "", ""
			case !v.Eligible:
				line1 = b.tr("銀河議會尚未成立", "The Galactic Council has not convened")
				line2, oc = b.tr("需半數銀河星系已殖民 + ≥2個存續帝國",
					"Requires half the galaxy colonized and 2+ surviving empires"), neutral
			default:
				line1 = fmt.Sprintf(b.tr("議會已成立(第 %d 屆待開)", "Council convened (election %d pending)"), v.Meetings+1)
				line2, oc = b.tr("尚無一方達2/3多數", "No one holds a 2/3 majority yet"), neutral
			}
			// 逐帝國投票明細(僅在議會已成立且尚未分出勝負時攤開;其餘狀態沿用 line1/line2 摘要)。
			if v.Eligible && !v.Victory.Over && v.Pending == nil {
				bd := b.session.CouncilBreakdown()
				if bd.Valid {
					gold := color.RGBA{240, 220, 120, 255}
					line1 = fmt.Sprintf(b.tr("第 %d 屆待開  候選人:%s／%s  達2/3需 %d／%d 票",
						"Election %d pending  Candidates: %s / %s  2/3 needs %d of %d votes"),
						v.Meetings+1, bd.Candidates[0], bd.Candidates[1], bd.Threshold, bd.Total)
					s.extras = append(s.extras,
						extraText{x: moo2ScreenW / 2, y: 96, size: 13, text: line1, col: gold, align: 1})
					y := 128.0
					for _, r := range bd.Rows {
						var suffix string
						rc := neutral
						switch {
						case r.IsCandidate:
							suffix, rc = b.tr("(候選人)", "(candidate)"), gold
						case r.Abstained:
							suffix, rc = b.tr("→ 棄權", "→ abstains"), lose
						default:
							suffix = b.tr("→ 投給 ", "→ votes for ") + r.VotedFor
						}
						txt := fmt.Sprintf(b.tr("%s  %d 票  %s", "%s  %d votes  %s"), r.Name, r.BaseVotes, suffix)
						s.extras = append(s.extras,
							extraText{x: moo2ScreenW / 2, y: y, size: 14, text: txt, col: rc, align: 1})
						y += 24
					}
					line1 = ""
					line2, oc = b.tr("第三方帝國依外交關係投票或棄權(手冊 p.183)",
						"Third-party empires vote or abstain by diplomatic standing (manual p.183)"), neutral
				}
			}
			if line1 != "" {
				s.extras = append(s.extras,
					extraText{x: moo2ScreenW / 2, y: 418, size: 15, text: line1, col: neutral, align: 1})
			}
			if line2 != "" {
				s.extras = append(s.extras,
					extraText{x: moo2ScreenW / 2, y: 444, size: 17, text: line2, col: oc, align: 1})
			}
		}
	}
	return s, nil
}

// ============ NEW GAME 設定畫面(原版 `Newgame_Screen_` @ 0xCD435)============
//
// 版面**全部取自反組譯**,先前是 PIL 量測的估計值。三個獨立來源互相印證:
//
//	① 建 widget:`sub_CCE2E`。畫面原點 `word_1831D4`=X=0x0F(15)、`word_1831D6`=Y=0x05(5)。
//	   五個設定框(`sub_11438B` 熱區)與五條數值列(`sub_C6A43` 選擇器)、三個開關
//	   (`sub_11523B`)、兩顆鈕(`sub_1151B0`)的座標全是立即數。
//	② 畫值圖:`sub_CCC3D` 是一個 3 欄 × 2 列的迴圈,起點 (X+0x79, Y+0x77)、
//	   欄距 0x9B(155)、列距 0x8C(140),右下那格跳過。算出來的欄 x = 121/276/431、
//	   列 y = 119/259,**與 ① 的熱區 x1/y1 逐一相同**。
//	③ 美術:`NEWGAME.LBX` 資產 1–22 剛好 22 張 65×65,而五個選擇器的選項數
//	   3+5+4+7+3 = 22。逐張看內容,分組是自證的:
//	     1–3   熔岩 / 沙漠 / 綠地湖泊      → 星系年齡(年輕→成熟)
//	     4–8   伸手扶持 → … → 雙拳相抵    → 難度(五級)
//	     9–12  螺旋星系由小到大            → 星系大小
//	     13–19 數字 2 3 4 5 6 7 8          → 帝國總數
//	     20–22 城市由樸素到未來            → 起始科技等級
//
// ⚠ **修正一個真的還原錯誤**:左下那個框在原版是 **PLAYERS**(帝國總數 2–8,變數
// `word_1A1366` 由 `byte_199CB1 − 2` 得來,配 13–19 的數字圖),remake 先前把它當成
// RACE 用。種族在原版是 ACCEPT 之後的**獨立畫面**(remake 本來就是這樣走的),
// 所以這裡改回 PLAYERS,不影響流程。
//
// ⚠ 誠實留白:TECH LEVEL 目前**只存設定、不影響 gameplay**(原因與手冊給的硬證見
// `shell.TechLevels` 註解)。patch 1.5 的第四級 Post-warp 也未做——1.5 的 CHANGELOG
// 明寫那一級的圖是 1.5 才加進 newgame.lbx 的,1.3 的 LBX 裡沒有。

// 原版新遊戲畫面的座標(全部相對螢幕;已含原點 (15,5))。
const (
	ngOriginX, ngOriginY = 15, 5 // word_1831D4 / word_1831D6

	ngBoxW, ngBoxH = 67, 65 // 熱區 0x79..0xBC / 0x77..0xB8
	ngColStep      = 155    // 0x9B
	ngRowStep      = 140    // 0x8C
	ngBoxX0        = 121    // 0x79
	ngBoxY0        = 119    // 0x77
	ngPicW, ngPicH = 65, 65 // NEWGAME.LBX 值圖尺寸

	ngStripX0 = 105 // 0x69   數值列(選擇器)左緣,相對原點
	ngStripY0 = 204 // 0xCC
	ngStripW  = 100 // 0xCD−0x69
	ngStripH  = 20  // 0xE0−0xCC
	// 數值列的欄距與列距同值圖格(0x105−0x69 = 0x9C ≈ 欄距;0x15D−0xCC = 0x91 ≈ 列距)。
	// 這兩個差值與值圖格的 0x9B/0x8C 差 1，故各自另記,不共用常數。
	ngStripColStep = 156 // 0x105−0x69
	ngStripRowStep = 145 // 0x15D−0xCC

	ngToggleX  = 380 // 0x17C,三個開關同一個 x
	ngToggleY0 = 259 // 0x103
	ngToggleY1 = 295 // 0x127
	ngToggleY2 = 330 // 0x14A

	ngCancelX, ngCancelY = 100, 386 // 0x64, 0x182
	ngAcceptX, ngAcceptY = 418, 387 // 0x1A2, 0x183
)

// 三個設定的預設索引。用具名常數而非字面數字,因為這些清單會增刪
// (2026-08-07 補「教學」之後難度索引就整體位移過一次,見 shell.Difficulties)。
const (
	newGameDiffDefault = 2 // 普通
	newGameAgeDefault  = 1 // 普通(gamedata.GalaxyAverage)
	newGameTechDefault = 1 // 一般
)

// ngSetting 是新遊戲畫面上的一個設定欄。
type ngSetting struct {
	act      string
	col, row int // 值圖格的欄/列(0-based)
	asset0   int // 該設定第一個選項在 NEWGAME.LBX 的資產索引
	n        func(b *sceneBuilder) int
	idx      func(b *sceneBuilder) int
	set      func(b *sceneBuilder, i int)
	label    func(b *sceneBuilder) string
}

// ngSettings 是五個設定欄。欄/列與資產起點見檔頭的三來源對照。
var ngSettings = []ngSetting{
	{
		act: "diff", col: 0, row: 0, asset0: 4,
		n:     func(b *sceneBuilder) int { return len(shell.Difficulties) },
		idx:   func(b *sceneBuilder) int { return b.newGameDiff },
		set:   func(b *sceneBuilder, i int) { b.newGameDiff = i },
		label: func(b *sceneBuilder) string { return shell.Difficulties[b.newGameDiff].Name },
	},
	{
		act: "size", col: 1, row: 0, asset0: 9,
		n:   func(b *sceneBuilder) int { return len(shell.GalaxySizes) },
		idx: func(b *sceneBuilder) int { return b.newGameSize },
		set: func(b *sceneBuilder, i int) { b.newGameSize = i },
		label: func(b *sceneBuilder) string {
			gs := shell.GalaxySizes[b.newGameSize]
			return fmt.Sprintf(b.tr("%s %d 星", "%s  %d stars"), gs.Name, gs.Stars)
		},
	},
	{
		act: "age", col: 2, row: 0, asset0: 1,
		n:     func(b *sceneBuilder) int { return len(shell.GalaxyAges) },
		idx:   func(b *sceneBuilder) int { return b.newGameAge },
		set:   func(b *sceneBuilder, i int) { b.newGameAge = i },
		label: func(b *sceneBuilder) string { return shell.GalaxyAges[b.newGameAge].Name },
	},
	{
		act: "players", col: 0, row: 1, asset0: 13,
		n:   func(b *sceneBuilder) int { return shell.MaxEmpires - shell.MinEmpires + 1 },
		idx: func(b *sceneBuilder) int { return b.newGameEmpires - shell.MinEmpires },
		set: func(b *sceneBuilder, i int) { b.newGameEmpires = shell.MinEmpires + i },
		label: func(b *sceneBuilder) string {
			return fmt.Sprintf(b.tr("%d 個帝國", "%d empires"), b.newGameEmpires)
		},
	},
	{
		act: "tech", col: 1, row: 1, asset0: 20,
		n:     func(b *sceneBuilder) int { return len(shell.TechLevels) },
		idx:   func(b *sceneBuilder) int { return b.newGameTech },
		set:   func(b *sceneBuilder, i int) { b.newGameTech = i },
		label: func(b *sceneBuilder) string { return shell.TechLevels[b.newGameTech].Name },
	},
}

// ngBoxRect 回傳某設定的值圖格(螢幕座標)。
func ngBoxRect(s ngSetting) (int, int, int, int) {
	return ngOriginX + ngBoxX0 + s.col*ngColStep, ngOriginY + ngBoxY0 + s.row*ngRowStep, ngBoxW, ngBoxH
}

// ngStripRect 回傳某設定的數值列(螢幕座標)。
func ngStripRect(s ngSetting) (int, int, int, int) {
	return ngOriginX + ngStripX0 + s.col*ngStripColStep, ngOriginY + ngStripY0 + s.row*ngStripRowStep,
		ngStripW, ngStripH
}

// newGameSetup 建原版新遊戲設定畫面(NEWGAME.LBX 資產 28,調色盤鏈 RACEOPT#4→NEWGAME#1)。
// ACCEPT 進種族選擇;CANCEL 回主選單。版面來源見上方檔頭區塊。
func (b *sceneBuilder) newGameSetup() (*overlayScreen, error) {
	hits := make([]hitRegion, 0, len(ngSettings)*2+2)
	for _, st := range ngSettings {
		x, y, w, h := ngBoxRect(st)
		hits = append(hits, hitRegion{x, y, w, h, st.act})
		sx, sy, sw, sh := ngStripRect(st)
		hits = append(hits, hitRegion{sx, sy, sw, sh, st.act}) // 數值列也可點,與原版一致
	}
	hits = append(hits,
		hitRegion{ngOriginX + ngCancelX, ngOriginY + ngCancelY, 108, 30, "cancel"},
		hitRegion{ngOriginX + ngAcceptX, ngOriginY + ngAcceptY, 108, 30, "accept"})

	onAction := func(a string) *origTransition {
		for _, st := range ngSettings {
			if st.act != a {
				continue
			}
			st.set(b, (st.idx(b)+1)%st.n(b)) // 點一下換下一個選項(原版是左右箭頭,remake 循環)
			return b.goTo(b.newGameSetup, "新遊戲設定")
		}
		if a == "accept" {
			// 原版流程:星系設定 → Accept →【獨立種族選擇畫面】(不在此直接開局)。
			sc, err := b.raceSelect()
			if err != nil {
				fmt.Fprintf(os.Stderr, "載入種族選擇: %v\n", err)
				return nil
			}
			return &origTransition{next: sc}
		}
		return b.goTo(b.menu, "主選單")
	}
	// 標籤位置是**對美術量測**的:這些字烘在 newgame.lbx#28 背景圖裡,不是程式畫的,
	// 反組譯裡沒有它們的座標——量圖就是這一項的一手來源。做法是掃出每個標籤的亮像素
	// 外接矩形(2026-08-07 重量,見下),不是目測估的。
	//
	// 重量之前的舊值有兩個看得見的錯:
	//   ① PLAYERS / TECH LEVEL 的 y 是 222,而數值列佔 209..229 —— 標籤壓在數值列上。
	//      實測標籤在 y 229..251,要往下移。
	//   ② 三個開關的標籤 x 從 422 起,但烘上去的英文從 416(ANTARANS)/418(TACTICAL)開始,
	//      左邊那一兩個字母擦不掉,畫面上會留「AI」「T.」這種殘字。
	overlays := []labelRect{
		{107, 84, 105, 22, "DIFFICULTY", 0},  // 實測 107..208 × 85..104
		{252, 84, 123, 22, "GALAXY SIZE", 0}, // 253..372
		{408, 84, 118, 22, "GALAXY AGE", 0},  // 409..523
		{108, 228, 95, 24, "PLAYERS", 0},     // 109..200 × 229..251
		{257, 228, 104, 24, "TECH LEVEL", 0}, // 258..358 × 230..251
		{412, 268, 112, 15, "TACTICAL COMBAT", 11},
		{412, 303, 112, 15, "RANDOM EVENTS", 11},
		{412, 338, 112, 15, "ANTARANS ATTACK", 11},
		{100, 385, 95, 20, "CANCEL", 0}, // 101..192 × 386..401
		{419, 386, 94, 22, "ACCEPT", 0}, // 420..510 × 388..407
	}
	s, err := loadOverlayScreen(b.res, "newgame.lbx", 28, b.lang, b.fnt, "assets/i18n/menu.tsv",
		overlays, color.RGBA{210, 216, 230, 255}, 13, hits, onAction,
		paletteChain{{"raceopt.lbx", 4}, {"newgame.lbx", 1}})
	if err != nil {
		return nil, err
	}
	// 五個設定各畫一張原版值圖 + 中文值名(原版就是「格內圖 + 格下文字」兩層,見 sub_CCC3D)。
	pics := b.newGamePics()
	s.postDraw = func(dst *ebiten.Image) {
		gold := color.RGBA{240, 220, 120, 255}
		for _, st := range ngSettings {
			x, y, _, _ := ngBoxRect(st)
			if im := pics[st.asset0+st.idx(b)]; im != nil {
				op := &ebiten.DrawImageOptions{}
				op.GeoM.Translate(float64(x+(ngBoxW-ngPicW)/2), float64(y))
				dst.DrawImage(im, op)
			}
			if b.fnt == nil {
				continue
			}
			sx, sy, sw, sh := ngStripRect(st)
			b.fnt.DrawCentered(dst, st.label(b), float64(sx+sw/2), float64(sy+sh/2), 12, gold)
		}
	}
	return s, nil
}

// newGamePics 解出 NEWGAME.LBX 的 22 張設定值圖(資產 1–22),索引即資產編號。
// 資產 1–3 自帶調色盤,其餘沒有,一律借資產 1 的上色(它們同屬一組美術)。
// 載不動就回空 map——值名文字仍會畫,畫面不會壞。
func (b *sceneBuilder) newGamePics() map[int]*ebiten.Image {
	out := map[int]*ebiten.Image{}
	prov, err := decodeAsset(b.res, "newgame.lbx", 1)
	if err != nil || prov.Embedded == nil {
		return out
	}
	for i := 1; i <= 22; i++ {
		im, err := decodeAsset(b.res, "newgame.lbx", i)
		if err != nil || len(im.Frames) == 0 {
			continue
		}
		out[i] = ebiten.NewImageFromImage(im.Frames[0].ToRGBA(prov.Embedded, false))
	}
	return out
}

// fleet 建原版艦隊列表畫面(FLEET.LBX 資產 0,三段調色盤鏈)。座標經 PIL 量測
// (screens-scan/fleetlist.png):標題列 y=27,兩排按鈕列 y=394/443。
func (b *sceneBuilder) fleet() (*overlayScreen, error) {
	// 點右側艦艇格 → 艦艇設計;右下 RETURN → 星系主畫面(精確熱區)。
	// 左下空白區(x<338, y 388-408)加一個「攻打安塔蘭」熱區(手冊三條勝利路徑之二,見
	// internal/shell/antaran_victory.go)。點擊一律進安塔蘭王座廳(原版 Main_Antaran_Room),
	// 發動與否在那裡決定;提示文字仍只在條件滿足時顯示,免得沒有傳送門就先誘導玩家點進去。
	// fleetHits 由下面的名冊繪製迴圈填入(每個艦隊標頭一個),點下去切換操作中的艦隊。
	var fleetHits []hitRegion
	hits := []hitRegion{
		{338, 50, 288, 300, "design"},
		{20, 388, 260, 20, "assault"},
		// RELOCATE(remake 譯「調動」)——手冊逐字:「You set up your Relocation orders on the
		// Fleet Operations console.」**這才是原版的入口**,座標同下面 overlays 的標籤框。
		{440, 384, 93, 18, "relocate"},
		// ALL(remake 譯「全部」)—— 推測對應 Set_All_Star_Relocations_(第 59 項)。
		// ⚠ 那支只改**已經有集結點**的殖民地,不是「全部設成這顆」。
		{346, 384, 70, 18, "relocateall"},
		// RETURN 真值座標取自 openorion2 ships.cpp:718 FleetListView
		// RETURN createWidget(556, 430, ...)(原估計 543,432)。
		{556, 430, 84, 28, "return"},
	}
	onAction := func(a string) *origTransition {
		switch a {
		case "design":
			return b.goTo(b.shipDesign, "艦艇設計")
		}
		if strings.HasPrefix(a, "selfleet") && b.session != nil {
			if n, err := strconv.Atoi(a[len("selfleet"):]); err == nil {
				b.session.SelectFleet(n)
				b.shipPick = map[int]bool{} // 換艦隊 → 清掉選船(索引是艦隊內的,換一支就沒意義)
				return b.goTo(b.fleet, "艦隊列表")
			}
			return nil
		}
		if strings.HasPrefix(a, "pickship") && b.session != nil {
			if n, err := strconv.Atoi(a[len("pickship"):]); err == nil {
				if b.shipPick == nil {
					b.shipPick = map[int]bool{}
				}
				b.shipPick[n] = !b.shipPick[n]
				return b.goTo(b.fleet, "艦隊列表")
			}
			return nil
		}
		if a == "splitfleet" && b.session != nil {
			var picked []int
			for si, on := range b.shipPick {
				if on {
					picked = append(picked, si)
				}
			}
			if _, ok := b.session.SplitFleet(b.session.SelectedFleet, picked); ok {
				b.shipPick = map[int]bool{}
			}
			return b.goTo(b.fleet, "艦隊列表")
		}
		switch a {
		case "relocateall":
			b.beginRelocateAll()
			b.flash(b.tr("全部:點一顆星,把已經設過集結點的殖民地全部改送過去",
				"ALL: click a star to retarget every existing rally point"))
			return b.goTo(b.galaxy, "星系主畫面")
		case "relocate":
			// 原版 `Star_Relocation_` 是兩段點選:先起點星(自己的殖民地)、再終點星。
			// 回到星圖進第一段。
			b.beginRelocatePick()
			b.flash(b.tr("調動:先點一顆自己的殖民星當起點", "Relocate: click one of your colony stars first"))
			return b.goTo(b.galaxy, "星系主畫面")
		case "assault":
			// 進安塔蘭王座廳(原版 Main_Antaran_Room),由那個畫面確認後才發動。
			// 前置條件不滿足時照樣進得去——王座廳會逐條講明卡在哪,比「點了沒反應」清楚。
			sc, err := b.antaranRoom()
			if err != nil {
				fmt.Fprintln(os.Stderr, "安塔蘭王座廳:", err)
				return b.goTo(b.fleet, "艦隊列表")
			}
			return &origTransition{next: sc}
		}
		return b.goTo(b.galaxy, "星系主畫面")
	}
	overlays := []labelRect{
		{190, 17, 260, 20, "FLEET OPERATIONS", 0},
		{346, 384, 70, 18, "ALL", 0},
		{440, 384, 93, 18, "RELOCATE", 0},
		{549, 384, 64, 18, "SCRAP", 0},
		{342, 436, 76, 18, "LEADERS", 0},
		{425, 436, 60, 18, "Support", 0},
		// Combat/RETURN/SCRAP 標籤由 PIL 校正為 openorion2 ships.cpp 真值(Combat 按鈕 487,435,60,19;
		// RETURN 對齊按鈕框 556,430,84,28;SCRAP 549)。
		{487, 435, 60, 19, "Combat", 0},
		{556, 430, 84, 28, "RETURN", 0},
	}
	s, err := loadOverlayScreen(b.res, "fleet.lbx", 0, b.lang, b.fnt, "assets/i18n/menu.tsv",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"buffer0.lbx", 0}, {"fleet.lbx", 111}})
	if err != nil {
		return nil, err
	}
	// 艦隊名冊填進左下暗面板。
	//
	// ⚠ **這裡列的是「艦隊」不是「船」。** 先前攤平成一長串船名,那是單艦隊時代的殘留——
	// 全帝國只有一支艦隊時,「列船」與「列艦隊」看起來一樣。多艦隊之後就不一樣了:
	// 玩家需要看到哪幾艘在一起、停在哪、有沒有在航行,才能選要操作哪一支。
	// (畫面標題是 FLEET OPERATIONS,不是 SHIP LIST。)
	if b.session != nil {
		gold := color.RGBA{240, 220, 120, 255}
		body := color.RGBA{206, 214, 232, 255}
		head := color.RGBA{150, 220, 200, 255}
		sel := color.RGBA{250, 230, 140, 255}
		y := 312.0
		for fi := range b.session.Fleets {
			f := &b.session.Fleets[fi]
			// 艦隊標頭:選中標記 + 所在地 + 航行狀態。
			mark := "  "
			hc := head
			if fi == b.session.SelectedFleet {
				mark, hc = "▶ ", sel
			}
			loc := b.tr("未知", "unknown")
			if f.AtStar >= 0 && f.AtStar < len(b.session.Stars) {
				loc = b.session.Stars[f.AtStar].Name
			}
			title := fmt.Sprintf(b.tr("%s第 %d 艦隊 — %s(%d 艘)", "%sFleet %d — %s (%d ships)"),
				mark, fi+1, loc, len(f.Ships))
			if f.DestStar >= 0 && f.DestStar < len(b.session.Stars) {
				title += fmt.Sprintf(b.tr(" → %s,%d 回合", " → %s, %d turns"),
					b.session.Stars[f.DestStar].Name, f.ETA)
			}
			s.extras = append(s.extras, extraText{x: 24, y: y, size: 12, text: title, col: hc})
			fleetHits = append(fleetHits, hitRegion{20, int(y) - 12, 300, 16, fmt.Sprintf("selfleet%d", fi)})
			y += 20
			// 拆分入口:選了至少一艘、又不是全部時才出現。
			//
			// ⚠ 原版這個畫面的美術上**沒有 SPLIT 鈕**(烘著的是 ALL / RELOCATE / SCRAP /
			// LEADERS / Support / Combat / RETURN)——原版是在右側艦艇格選船再下令。
			// remake 的右側格還沒接上選取,所以先用左側名冊選 + 這一行當入口。
			// **這是 remake 自己加的控制項**,追到原版怎麼下拆分令之後要換掉。
			//
			// 放在**標頭底下**而不是船清單之後:名冊往下長,放在後面會撞到 y=402 那行
			// 「攻打安塔蘭母星」(第一版就是這樣疊在一起的)。
			if fi == b.session.SelectedFleet {
				n := 0
				for si := range f.Ships {
					if b.shipPick[si] {
						n++
					}
				}
				if n > 0 && n < len(f.Ships) {
					s.extras = append(s.extras, extraText{x: 34, y: y, size: 11,
						text: fmt.Sprintf(b.tr("▶ 把選中的 %d 艘拆成新艦隊", "▶ Split %d selected into a new fleet"), n),
						col:  color.RGBA{150, 230, 180, 255}})
					fleetHits = append(fleetHits, hitRegion{28, int(y) - 11, 296, 15, "splitfleet"})
					y += 18
				}
			}
			for si, sh := range f.Ships {
				// 選船(供拆分用):只有目前操作中的艦隊可以選——拆分是對「這一支」做的。
				mk := ""
				nameCol := gold
				if fi == b.session.SelectedFleet {
					if b.shipPick[si] {
						mk, nameCol = "✔ ", sel
					} else {
						mk = "· "
					}
					fleetHits = append(fleetHits, hitRegion{34, int(y) - 11, 290, 15, fmt.Sprintf("pickship%d", si)})
				}
				s.extras = append(s.extras,
					extraText{x: 40, y: y, size: 12, text: mk + sh.Name, col: nameCol},
					extraText{x: 140, y: y, size: 11, text: sh.Class, col: body},
				)
				// 結構損傷(見 internal/shell/repair.go)。原版是在艦艇資訊面板用損壞色標示,
				// remake 只有結構這一份損傷值,直接寫百分比;完好的船不畫,免得整排都是「損傷 0%」。
				if d := shell.ShipDamagePercent(sh); d > 0 {
					col := color.RGBA{235, 190, 90, 255} // 輕傷:琥珀
					if d >= 50 {
						col = color.RGBA{230, 110, 90, 255} // 重傷:紅
					}
					s.extras = append(s.extras,
						extraText{x: 246, y: y, size: 11, text: fmt.Sprintf(b.tr("損傷 %d%%", "%d%% damaged"), d), col: col})
				}
				y += 18
			}
			y += 6
		}
		// 「攻打安塔蘭」提示(手冊三條勝利路徑之二):只在已建次元傳送門 + 艦隊非空時顯示,
		// 對應上面 hits 的 "assault" 熱區(見 CanAssaultAntares)。點下去進王座廳,
		// 發動與否在那裡決定——文案要講清楚是「進去看」而不是「按了就打」。
		if b.session.CanAssaultAntares() {
			warn := color.RGBA{235, 160, 90, 255}
			s.extras = append(s.extras, extraText{x: 28, y: 402, size: 13, text: b.tr("攻打安塔蘭母星(點此進入王座廳)", "Assault the Antaran homeworld (click for the throne room)"), col: warn})
		}
	}
	// 艦隊標頭的熱區要等名冊畫完才知道有幾個,所以最後補進去
	// (loadOverlayScreen 已經把 hits 複製走了,直接接在 s.hits 後面)。
	s.hits = append(s.hits, fleetHits...)
	return s, nil
}

// 艦艇設計畫面的原版座標(全部是 sub_6C8F9 / Add_Design_Buttons_ 的立即數,見下方檔頭)。
var (
	dsHullOrder = []string{"Frigate", "Destroyer", "Cruiser", "Battleship", "Titan", "Doom Star"}
	// shipClassZH 是 shell 那邊的艦體 key(中文),順序同 dsHullOrder / gamedata.CombatShipClass。
	// **這是 key 不是顯示字**——`shell.ShipCost` / `DesignCostWithMods` 都拿它查表,
	// 換成英文會直接查不到。要顯示英文請走 shipClassLabel。
	shipClassZH = []string{"巡防艦", "驅逐艦", "巡洋艦", "戰艦", "泰坦", "末日之星"}
	// dsHullY[i] = {y1, y2}。原版六格**不等距**(高 15/14/17/14/14/16),
	// 不是等距 17px——先前 remake 就是照等距排的,越往下偏得越多。
	dsHullY = [6][2]int{{54, 69}, {70, 84}, {85, 102}, {103, 117}, {118, 132}, {133, 149}}
	// 底部三顆鈕(CLEAR / CANCEL / BUILD)。
	dsBtnX = [3]int{374, 461, 547}
)

const (
	dsHullX0, dsHullX1 = 118, 227 // 0x76..0xE3,六格共用
	dsBtnY             = 443
	dsBtnW, dsBtnH     = 80, 22
)

// ============ 艦艇設計畫面(原版 `Design_Screen_` @ 0x6B9B2)============
//
// 版面 2026-08-07 改用反組譯真值(先前是估計座標,x 差 7px、底部三鈕差 6–11px,
// 而且**六列艦體是照 17px 等距排的,原版根本不等距**)。
//
//	`sub_6C8F9`(由 `Add_Design_Buttons_` @ 0x69E62 對 i=0..5 呼叫)是一張乾淨的
//	switch 表,直接給出六個艦體槽的矩形:
//	    x 一律 `0x76..0xE3` = **118..227**(共用,寫在 default 分支)
//	    y1 = 0x36 / 0x46 / 0x55 / 0x67 / 0x76 / 0x85 = 54 / 70 / 85 / 103 / 118 / 133
//	    y2 = 0x45 / 0x54 / 0x66 / 0x75 / 0x84 / 0x95 = 69 / 84 / 102 / 117 / 132 / 149
//	    → 列高 15/14/17/14/14/16,**不等距**(那是原版美術上那六格的實際高度)
//
//	底部三顆鈕(`sub_1151B0`,引數是三個熱鍵字串 `aLb` / `+2` / `+4`):
//	    (374, 443) / (461, 443) / (547, 443)
//
// ⚠ 尚未套用、但座標已到手(記在 docs/re/01-gap-report.md 第 24 項):
//   - 已裝元件清單列:x 55..68、y = 169 + 13i(`imul eax, esi, 0Dh` / `add eax, 0A9h`)
//   - 右上兩個資訊面板:(437..627, 56..95) 與 (437..627, 97..123)
//     remake 現在把元件選擇列排在 x 300..600,與原版這兩個面板的位置不同;
//     要對齊得先確認那兩格在原版顯示什麼(尚未追到繪製端)。
//
// 點艦體等級 → 建造該艦加入艦隊 → 回艦隊;點他處 → 返回艦隊。
func (b *sceneBuilder) shipDesign() (*overlayScreen, error) {
	// 原版艦體名 → shell 的中文 key。由既有的兩份表建,不再手寫第三份對照
	// (三份表遲早會漂移;順序本來就一致,見 shipClassZH 註解)。
	hullZH := make(map[string]string, len(dsHullOrder))
	for i, en := range dsHullOrder {
		hullZH[en] = shipClassZH[i]
	}
	hits := make([]hitRegion, 0, 20)
	// 六個艦體槽:座標為反組譯真值(見檔頭的 sub_6C8F9 switch 表),不等距。
	for i, name := range dsHullOrder {
		y0, y1 := dsHullY[i][0], dsHullY[i][1]
		hits = append(hits, hitRegion{dsHullX0, y0, dsHullX1 - dsHullX0 + 1, y1 - y0 + 1, name})
	}
	hits = append(hits,
		hitRegion{300, 58, 300, 22, "weapon"}, // 元件選擇(點擊各列循環)
		hitRegion{300, 82, 300, 22, "armor"},
		hitRegion{300, 106, 300, 22, "shield"},
		hitRegion{300, 130, 300, 22, "special"},
		// 武器改造(mod)勾選:8 個 chip,兩排各 4 個,順序對齊 shell.WeaponModOptions
		// (HV/PD/AF/CO 第一排,AP/ENV/NR/SP 第二排)。只對 beam 武器生效(見 onAction
		// 的 WeaponIsBeam 判斷),非 beam 武器仍顯示熱區但點擊不生效(避免熱區位移)。
		hitRegion{305, 368, 76, 18, "mod:0"}, hitRegion{385, 368, 76, 18, "mod:1"},
		hitRegion{465, 368, 76, 18, "mod:2"}, hitRegion{545, 368, 76, 18, "mod:3"},
		hitRegion{305, 390, 76, 18, "mod:4"}, hitRegion{385, 390, 76, 18, "mod:5"},
		hitRegion{465, 390, 76, 18, "mod:6"}, hitRegion{545, 390, 76, 18, "mod:7"},
		hitRegion{0, 0, moo2ScreenW, moo2ScreenH, "back"},
	)
	onAction := func(a string) *origTransition {
		switch a { // 循環只跳到「已研究解鎖」的元件
		case "weapon":
			b.designWeapon = b.session.NextUnlockedComponent(shell.WeaponOptions, b.designWeapon)
			b.designMsg = "" // 換元件可能改變空間是否超格,清掉舊的建造提示避免誤導
			return b.goTo(b.shipDesign, "艦艇設計")
		case "armor":
			b.designArmor = b.session.NextUnlockedComponent(shell.ArmorOptions, b.designArmor)
			b.designMsg = ""
			return b.goTo(b.shipDesign, "艦艇設計")
		case "shield":
			b.designShield = b.session.NextUnlockedComponent(shell.ShieldOptions, b.designShield)
			b.designMsg = ""
			return b.goTo(b.shipDesign, "艦艇設計")
		case "special":
			b.designSpecial = b.session.NextUnlockedComponent(shell.SpecialOptions, b.designSpecial)
			b.designMsg = ""
			return b.goTo(b.shipDesign, "艦艇設計")
		}
		if strings.HasPrefix(a, "mod:") {
			// mods 只對 beam 武器有意義(手冊 HV/PD/AF/CO 明文只講 beam,見
			// shell.WeaponIsBeam);非 beam 武器(核飛彈/麥克萊特飛彈)點擊不生效,
			// 避免玩家對飛彈掛上不會被套用的改造造成誤導。
			w := shell.WeaponOptions[b.designWeapon]
			if shell.WeaponIsBeam(w.Name) {
				var idx int
				fmt.Sscanf(a, "mod:%d", &idx)
				if idx >= 0 && idx < len(shell.WeaponModOptions) {
					b.designMods = shell.ToggleWeaponMod(b.designMods, shell.WeaponModOptions[idx])
					b.designMsg = ""
				}
			}
			return b.goTo(b.shipDesign, "艦艇設計")
		}
		if zh, ok := hullZH[a]; ok && b.session != nil {
			// 建造前驗證空間:超出艦體空間上限(shell.ShipDesignFitsWithMods)就擋下,留在設計畫面提示,不扣款不造艦。
			if !shell.ShipDesignFitsWithMods(zh, b.designWeapon, b.designArmor, b.designShield, b.designSpecial, b.designMods) {
				b.designMsg = fmt.Sprintf(b.tr("空間不足,無法建造%s(目前元件+改造超出艦體空間上限)",
					"%s does not fit — components plus mods exceed the hull space limit"),
					shipClassLabel(b.lang, zh))
				return b.goTo(b.shipDesign, "艦艇設計")
			}
			b.designMsg = ""
			b.session.BuildShipWithMods(zh, b.designWeapon, b.designArmor, b.designShield, b.designSpecial, b.designMods)
			return b.goTo(b.fleet, "艦隊列表")
		}
		return b.goTo(b.fleet, "艦隊列表")
	}
	overlays := []labelRect{{255, 12, 320, 24, "Ship Design", 0}}
	// 六列艦體名的擦底帶跟著槽走(各留 2px 邊,不吃到浮雕框)。
	for i, name := range dsHullOrder {
		y0, y1 := dsHullY[i][0], dsHullY[i][1]
		overlays = append(overlays, labelRect{
			dsHullX0 + 2, y0 + 2, dsHullX1 - dsHullX0 - 3, y1 - y0 - 3, name, 12})
	}
	// 底部三顆鈕:反組譯真值 (374/461/547, 443)。
	for i, name := range []string{"Clear", "Cancel", "Build"} {
		overlays = append(overlays, labelRect{dsBtnX[i], dsBtnY, dsBtnW, dsBtnH, name, 0})
	}
	s, err := loadOverlayScreen(b.res, "design.lbx", 0, b.lang, b.fnt, "assets/i18n/tech.tsv",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"buffer0.lbx", 0}})
	if err != nil {
		return nil, err
	}
	// 各艦體成本(對齊 MOO2 空殼生產成本)+ 目前國庫,顯示在艦體清單右方。
	if b.fnt != nil && b.session != nil {
		body := color.RGBA{210, 216, 230, 255}
		classes := shipClassZH
		// 價格跟著艦體槽的真實 y 走(dsHullY 不等距;先前照 60+17i 等距排,越往下偏越多,
		// 最後一格會壓到下面的總價那一行)。
		for i, cl := range classes {
			y0, y1 := dsHullY[i][0], dsHullY[i][1]
			s.extras = append(s.extras, extraText{
				x: float64(dsHullX1 + 16), y: float64(y0+y1)/2 - 6, size: 11,
				text: fmt.Sprintf("%d BC", shell.ShipCost(cl)), col: body, align: 0})
		}
		// 四類元件(點擊各列循環選擇),顯示名稱 + 效果 + 成本。
		w := shell.WeaponOptions[b.designWeapon]
		ar := shell.ArmorOptions[b.designArmor]
		sd := shell.ShieldOptions[b.designShield]
		sp := shell.SpecialOptions[b.designSpecial]
		gold := color.RGBA{240, 220, 120, 255}
		rows := []struct {
			label string
			c     shell.Component
			eff   string
		}{
			{b.tr("武器", "Weapon"), w, fmt.Sprintf(b.tr("+%d攻", "+%d atk"), w.Value)},
			{b.tr("裝甲", "Armor"), ar, fmt.Sprintf("+%dHP", ar.Value)},
			{b.tr("護盾", "Shield"), sd, fmt.Sprintf("+%dHP", sd.Value)},
			{b.tr("特殊", "Special"), sp, ""},
		}
		for i, r := range rows {
			y := float64(69 + i*24)
			s.extras = append(s.extras,
				extraText{x: 305, y: y, size: 12, text: r.label + " ▸ " + r.c.Name, col: gold},
				extraText{x: 470, y: y, size: 11, text: fmt.Sprintf("%s %dBC", r.eff, r.c.Cost), col: color.RGBA{200, 208, 225, 255}})
		}
		const designHull = "巡洋艦" // shell 的 key(見 shipClassZH 註解)
		total := shell.DesignCostWithMods(designHull, b.designWeapon, b.designArmor, b.designShield, b.designSpecial, b.designMods)
		// 各類已解鎖元件數(需研究對應科技解鎖進階元件)。
		cnt := func(opts []shell.Component) int {
			n := 0
			for _, c := range opts {
				if b.session.ComponentUnlocked(c) {
					n++
				}
			}
			return n
		}
		s.extras = append(s.extras,
			extraText{x: 305, y: 168, size: 12, text: fmt.Sprintf(b.tr("%s總價 %d BC", "%s total %d BC"),
				shipClassLabel(b.lang, designHull), total), col: color.RGBA{170, 220, 180, 255}},
			extraText{x: 305, y: 190, size: 11, text: fmt.Sprintf(b.tr(
				"已解鎖 武器%d/%d 裝甲%d/%d 護盾%d/%d 特殊%d/%d(研究科技解鎖進階元件)",
				"Unlocked: weapons %d/%d armor %d/%d shields %d/%d special %d/%d (research unlocks more)"),
				cnt(shell.WeaponOptions), len(shell.WeaponOptions), cnt(shell.ArmorOptions), len(shell.ArmorOptions),
				cnt(shell.ShieldOptions), len(shell.ShieldOptions), cnt(shell.SpecialOptions), len(shell.SpecialOptions)),
				col: color.RGBA{170, 200, 240, 255}},
			extraText{x: 12, y: 460, size: 12,
				text: fmt.Sprintf(b.tr("國庫 %d BC", "Treasury %d BC"), b.session.Player.BC), col: gold})

		// 空間預算/已用(依目前選定元件即時計算):逐艦體列出「空間:已用／總」,超格轉紅並標
		// 「空間不足」。點艦體列建造時,onAction 用同一份 shell.ShipDesignFits 判斷擋下建造
		// (不扣款、不入艦隊),designMsg 顯示擋下提示——顯示與建造驗證共用同一份判斷,不會不一致。
		spaceHeaderY := 208.0
		s.extras = append(s.extras, extraText{x: 305, y: spaceHeaderY, size: 12,
			text: b.tr("各艦體空間(依目前元件):", "Space per hull (with current components):"), col: gold})
		okCol := color.RGBA{170, 220, 180, 255}
		badCol := color.RGBA{230, 90, 90, 255}
		for i, cl := range classes {
			used := shell.ShipDesignSpaceUsedWithMods(cl, b.designWeapon, b.designArmor, b.designShield, b.designSpecial, b.designMods)
			totalSp := gamedata.ShipHullSpace(gamedata.CombatShipClass(i))
			fits := used <= totalSp
			txt := fmt.Sprintf(b.tr("%s 空間:%d／%d", "%s space %d/%d"), shipClassLabel(b.lang, cl), used, totalSp)
			col := okCol
			if !fits {
				txt += b.tr("(空間不足)", " (over capacity)")
				col = badCol
			}
			s.extras = append(s.extras, extraText{x: 305, y: spaceHeaderY + 17 + float64(i*17), size: 11, text: txt, col: col})
		}
		if b.designMsg != "" {
			s.extras = append(s.extras, extraText{x: 305, y: spaceHeaderY + 17 + float64(len(classes)*17) + 8, size: 12,
				text: b.designMsg, col: badCol})
		}

		// 武器改造(mod)勾選 chip:8 個,順序對齊 shell.WeaponModOptions 與上方 onAction 的
		// mod:0..7 熱區。已勾選轉金色高亮,未勾選灰色;非 beam 武器(核飛彈/麥克萊特飛彈)
		// 整排標「(僅光束武器適用)」提示,不隱藏熱區(避免版面跳動),點擊也不會生效
		// (onAction 已用 WeaponIsBeam 擋掉)。
		modHeaderY := 352.0
		isBeam := shell.WeaponIsBeam(w.Name)
		modHeaderTxt := b.tr("武器改造(點擊切換,HV/PD 互斥):", "Weapon mods (click to toggle; HV/PD exclusive):")
		if !isBeam {
			modHeaderTxt = b.tr("武器改造(僅光束武器適用,此武器不支援):",
				"Weapon mods (beam weapons only; this weapon does not support them):")
		}
		s.extras = append(s.extras, extraText{x: 305, y: modHeaderY, size: 11, text: modHeaderTxt, col: gold})
		activeCol := color.RGBA{240, 220, 120, 255}
		inactiveCol := color.RGBA{150, 155, 165, 255}
		modChipX := []float64{305, 385, 465, 545}
		for i, mod := range shell.WeaponModOptions {
			row := i / 4
			chipX := modChipX[i%4]
			y := modHeaderY + 16 + float64(row*22)
			chipCol := inactiveCol
			if shell.HasWeaponMod(b.designMods, mod) {
				chipCol = activeCol
			}
			s.extras = append(s.extras, extraText{x: chipX, y: y, size: 10, text: shell.WeaponModLabelZH(mod), col: chipCol})
		}
	}
	return s, nil
}

// officer 建原版軍官列表畫面(OFFICER.LBX 資產 0)。座標經 PIL 量測
// (screens-scan/officer_leaderlist.png):頁籤列 y=12-32,按鈕列 y=440-462。
func (b *sceneBuilder) officer() (*overlayScreen, error) {
	// 精確返回鍵熱區(RETURN 按鈕真值座標取自 openorion2 officer.cpp:418
	// LeaderListView RETURN createWidget(538, 441, ...);取代整畫面返回,僅返回鍵返回)。
	// HIRE 熱區對齊原版 OFFICER.LBX 的 HIRE 按鈕(overlay 標於 310,440):雇用 MercPool 首名傭兵。
	hits := []hitRegion{
		{538, 441, 80, 20, "Return"},
		{313, 440, 68, 20, "hire"}, // HIRE 按鈕真值 x=313(openorion2 officer.cpp;原 PIL 310)
	}
	onAction := func(a string) *origTransition {
		switch a {
		case "hire":
			if b.session != nil {
				b.session.HireMerc() // 雇用池首名傭兵(BC不足/滿員則無作用),手冊 p.134
			}
			return b.goTo(b.officer, "軍官列表")
		case "Return":
			return b.goTo(b.galaxy, "星系主畫面")
		}
		return nil
	}
	overlays := []labelRect{
		{20, 11, 133, 20, "Colony Leaders", 0},
		{166, 11, 124, 20, "Ship Officers", 0},
		{313, 440, 68, 20, "HIRE", 0},
		{388, 440, 69, 20, "POOL", 0},
		{462, 440, 74, 20, "DISMISS", 0},
		{540, 440, 80, 20, "RETURN", 0},
	}
	s, err := loadOverlayScreen(b.res, "officer.lbx", 0, b.lang, b.fnt, "assets/i18n/officer.tsv",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"buffer0.lbx", 0}})
	if err != nil {
		return nil, err
	}
	// 領袖名單填進左側槽位:槽中心 y = openorion2 officer.cpp LeaderListView 真值
	// FIRST_ROW 38 + SLOT_HEIGHT 105/2 + i*ROW_DIST 109 = 90.5+i*109 → 取整 90/199/308/417
	// (原 PIL 87/198/307/415 每列高約 1.5-3px、列距不均;現還原 109px 等距)。
	if b.session != nil {
		gold := color.RGBA{240, 220, 120, 255}
		body := color.RGBA{206, 214, 232, 255}
		hireCol := color.RGBA{150, 220, 160, 255} // 可雇用傭兵用綠色標示
		rowY := []float64{90, 199, 308, 417}
		row := 0
		// 已雇用領袖填前幾個槽位。
		for _, ld := range b.session.Leaders {
			if row >= len(rowY) {
				break
			}
			y := rowY[row]
			s.extras = append(s.extras,
				extraText{x: 95, y: y - 12, size: 15, text: ld.Name, col: gold},
				extraText{x: 95, y: y + 12, size: 12, text: fmt.Sprintf(b.tr("%s ｜ Lv %d", "%s | Lv %d"), ld.Skill, ld.Level), col: body},
			)
			row++
		}
		// 剩餘槽位顯示上門待雇的傭兵(綠色 + 雇用費;點 HIRE 鈕雇用池首名)。
		for _, ld := range b.session.MercPool {
			if row >= len(rowY) {
				break
			}
			y := rowY[row]
			s.extras = append(s.extras,
				extraText{x: 95, y: y - 12, size: 15, text: "◆ " + ld.Name, col: hireCol},
				extraText{x: 95, y: y + 12, size: 12, text: fmt.Sprintf(b.tr("%s ｜ Lv %d ｜ 雇用費 %d BC", "%s | Lv %d | hire %d BC"),
					ld.Skill, ld.Level, b.session.MercHireCost(ld)), col: hireCol},
			)
			row++
		}
		// 池空且無領袖時,提示傭兵會不定期上門(手冊 p.134)。
		if len(b.session.Leaders) == 0 && len(b.session.MercPool) == 0 {
			s.extras = append(s.extras, extraText{x: 95, y: rowY[0], size: 13, text: b.tr("(傭兵領袖會不定期上門,屆時按 HIRE 雇用)",
				"(mercenary leaders turn up from time to time; press HIRE when they do)"), col: body})
		}
	}
	return s, nil
}

// info 建原版科技總覽畫面(INFO.LBX 資產 0,基底 INFO.LBX 資產 1)。座標經 PIL 量測
// (screens-scan/info_overview.png):左側選單五列 y=57/79/105/134/154,標題 y=16,RETURN y=436。
func (b *sceneBuilder) info() (*overlayScreen, error) {
	// 「科技總覽」列 → 研究選擇畫面;RETURN → 星系主畫面。
	// RETURN 真值座標取自 openorion2 info.cpp:1028 InfoView
	// RETURN createWidget(535, 434, ...);取代整畫面返回,僅返回鍵返回。
	// 五個分頁各自的熱區(y 對齊下方 overlays 的五列選單;高 24 為列距內的可點帶)。
	// 原版結構由反組譯確認:INFO 是「單一畫面 + 5 個子畫面」,各有獨立的
	// Draw_History_Subscreen_ / Draw_Tech_Review_Subscreen_ / Draw_Race_Stats_Subscreen_ /
	// Draw_Turn_Summary_Subscreen_ / Draw_Reference_*_Subscreen_(見 docs/re/01-gap-report.md)。
	hits := []hitRegion{
		{21, 52, 164, 24, "tab0"},         // History Graph
		{21, 76, 164, 24, "tab1"},         // Tech Review
		{21, 102, 164, 24, "tab2"},        // Race Statistics
		{21, 130, 164, 22, "tab3"},        // Turn Summary
		{21, 152, 164, 22, "tab4"},        // Reference
		{214, 96, 412, 268, "histmetric"}, // 歷史圖表區:點擊循環指標(人口/國庫/艦隊)
		{535, 434, 84, 22, "back"},
	}
	onAction := func(a string) *origTransition {
		switch a {
		case "back":
			return b.goTo(b.galaxy, "星系主畫面")
		case "tab0", "tab1", "tab2", "tab3", "tab4":
			b.infoTab = int(a[3] - '0')
			return b.goTo(b.info, "情報") // 重繪切換子畫面
		case "histmetric":
			if b.infoTab == 0 { // 只有歷史圖表分頁才有意義
				b.infoHistoryMetric = (b.infoHistoryMetric + 1) % 3
				return b.goTo(b.info, "情報")
			}
		}
		return nil
	}
	// 選單項原版為靠左文字疊在近黑面板背景上(無實心板);擦底取黑=黑疊黑(正確),
	// rect 寬取足以蓋住最長英文、中文置中於偏左位置貼近原版。y 中心經 PIL 量測:64/88/114/142/162。
	// 五列選單 x/w 由 PIL 校正為 openorion2 info.cpp ChoiceWidget(21,50,164,131) 真值(x15→21、w182→164);
	// y 為 PIL(openorion2 按鈕 y 由精靈高 runtime 累加、無字面,均距推導與現值 ≤5px 一致,保留)。
	// STAR DATE 是烘進背景的字幕、openorion2 無座標,保留 PIL。RETURN 標籤對齊按鈕真值 535,434。
	overlays := []labelRect{
		{15, 20, 200, 26, "STAR DATE", 0},
		{21, 56, 164, 18, "History Graph", 0},
		{21, 80, 164, 18, "Tech Review", 0},
		{21, 106, 164, 18, "Race Statistics", 0},
		{21, 134, 164, 18, "Turn Summary", 0},
		{21, 154, 164, 18, "Reference", 0},
		{535, 434, 84, 22, "RETURN", 0},
	}
	s, err := loadOverlayScreen(b.res, "info.lbx", 0, b.lang, b.fnt, "assets/i18n/misc.tsv",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"info.lbx", 1}})
	if err != nil {
		return nil, err
	}
	// 右側面板依 b.infoTab 繪出對應子畫面(對齊原版 5 個 Draw_*_Subscreen_)。
	if b.fnt != nil {
		b.drawInfoSubscreen(s)
	}
	// info 選單/標題都疊在均勻的近黑面板背景上,強制用該背景色擦底(採樣會因長英文誤取字色)。
	black := color.RGBA{0, 8, 24, 255}
	s.eraseColor = &black
	return s, nil
}

// turnSummary 建原版回合摘要畫面(TURNSUM.LBX 資產 0,調色盤鏈 buffer0#0,置中視窗)。
// 原版流程:結束回合後顯示本回合結算;點 CLOSE 回星系主畫面。
func (b *sceneBuilder) turnSummary() (*overlayScreen, error) {
	hits, onAction := b.backHit(b.galaxy, "星系主畫面")
	overlays := []labelRect{
		{88, 14, 204, 22, "TURN SUMMARY", 0},
		{158, 324, 64, 18, "CLOSE", 0},
	}
	s, err := loadOverlayScreen(b.res, "turnsum.lbx", 0, b.lang, b.fnt, "assets/i18n/misc.tsv",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"buffer0.lbx", 0}})
	if err != nil {
		return nil, err
	}
	// 事件區(深色空面板)填本回合結算(座標為 bg 局部,draw 自動加置中偏移)。
	if b.session != nil {
		out := b.session.LastPlayerOutput
		year := 3500 + (b.session.Turn - 1)
		gold := color.RGBA{240, 220, 120, 255}
		body := color.RGBA{214, 220, 235, 255}
		s.extras = []extraText{
			{x: 40, y: 62, size: 15, text: fmt.Sprintf(b.tr("星曆 %d 結算", "Stardate %d report"), year), col: gold},
			{x: 40, y: 92, size: 13, text: fmt.Sprintf(b.tr("淨工業 %d ／ 研究 %d", "Net industry %d / research %d"),
				out.TotalNetIndustry, out.TotalResearch), col: body},
			{x: 40, y: 116, size: 13, text: fmt.Sprintf(b.tr("食物盈餘 %d ／ 稅收 %d BC", "Food surplus %d / taxes %d BC"),
				out.TotalFood, out.TaxRevenue), col: body},
			{x: 40, y: 140, size: 13, text: fmt.Sprintf(b.tr("國庫 %d BC(本回合 %+d)", "Treasury %d BC (%+d this turn)"),
				b.session.Player.BC, out.NetBC), col: body},
		}
		yy := 168.0
		if out.ResearchDone {
			s.extras = append(s.extras, extraText{x: 40, y: yy, size: 14, text: b.tr("★ 完成一項研究!", "★ A research field is complete!"), col: color.RGBA{120, 220, 140, 255}})
			yy += 24
		}
		for _, msg := range b.session.LastBuilt {
			s.extras = append(s.extras, extraText{x: 40, y: yy, size: 13, text: "★ " + msg, col: color.RGBA{120, 220, 140, 255}})
			yy += 22
		}
		// 隨機事件(繁榮/瘟疫/海盜…)。
		if b.session.LastEvent != "" {
			s.extras = append(s.extras, extraText{x: 40, y: yy, size: 14, text: "◆ " + b.session.LastEvent, col: color.RGBA{240, 190, 110, 255}})
			yy += 24
		}
		// 安塔蘭人入侵警報(紅色醒目)。
		if b.session.LastAntares != "" {
			s.extras = append(s.extras, extraText{x: 40, y: yy, size: 14, text: b.session.LastAntares, col: color.RGBA{240, 110, 90, 255}})
			yy += 24
		}
		// AI 對手突襲警報(見 shell/ai_attack.go)。擊退用綠字、被打用紅字,
		// 讓玩家一眼看出「這回合的軍備夠不夠」。
		if b.session.LastRaid != "" {
			col := color.RGBA{240, 110, 90, 255}
			if b.session.LastRaidReport != nil && b.session.LastRaidReport.Repelled {
				col = color.RGBA{130, 220, 150, 255}
			}
			s.extras = append(s.extras, extraText{x: 40, y: yy, size: 14, text: b.session.LastRaid, col: col})
			yy += 24
		}
	}
	return s, nil
}

// researchAreaOrder 把畫面 8 個領域熱區名對應到 gamedata.TechTree() 的領域索引(見
// internal/gamedata/techtree.go 陣列註解:0=Biology…7=Sociology)。
var researchAreaOrder = map[string]int{
	"Construction": 3, "Power": 1, "Chemistry": 5, "Sociology": 7,
	"Computers": 6, "Biology": 0, "Physics": 2, "Force Fields": 4,
}

// currentAreaTopic 回傳某研究領域「目前應研究的主題」:MOO2 原版機制是玩家選定領域、
// 該領域依 techtree 固定順序逐一解鎖(非玩家自由挑選領域內個別主題,完成一項後才跳下一項,
// 期間若有多科技可選走 researchChoiceScreen 另外決定),故此處回傳該領域第一個尚未完成的
// 主題 + 其 RP 成本(gamedata.researchChoices 為權威來源)。done=true 表示整領域已研究完畢。
func currentAreaTopic(session *shell.GameSession, areaIdx int) (topic gamedata.ResearchTopic, cost int, done bool) {
	topics := gamedata.TechTree()[areaIdx]
	completed := session.Player.CompletedTopics
	for _, t := range topics {
		if completed == nil || !completed[t] {
			return t, session.ResearchCostForDisplay(t), false
		}
	}
	if len(topics) > 0 {
		last := topics[len(topics)-1]
		return last, session.ResearchCostForDisplay(last), true
	}
	return 0, 0, true
}

// research 建原版研究選擇畫面(TECHSEL.LBX 資產 0,無內嵌調色盤 → 走調色盤鏈,
// 基底取自 SCIENCE.LBX 資產 0)。點畫面任一處返回主選單。
//
// 2026-07-11 修盲選 bug:原本 8 領域框各自綁死一個寫死的代表主題(如 Chemistry 恆選
// TOPIC_ADVANCED_CHEMISTRY),玩家看不到實際會研究哪個主題、要花多少 RP 就得盲點。
// 改為即時算出該領域「目前應研究的主題」(currentAreaTopic,依 techtree 固定順序取第一個
// 未完成主題)並把中文名 + RP 成本疊字顯示在領域框內,點擊即設定為該真主題(而非寫死值)。
func (b *sceneBuilder) research() (*overlayScreen, error) {
	// 8 個研究領域為點擊熱區(bg 局部座標;涵蓋整塊面板)→ 設定該領域目前主題 → 回星系。
	hits := []hitRegion{
		{16, 32, 208, 98, "Construction"}, {242, 32, 214, 98, "Power"},
		{16, 137, 208, 98, "Chemistry"}, {242, 137, 214, 98, "Sociology"},
		{16, 243, 208, 98, "Computers"}, {242, 243, 214, 98, "Biology"},
		{16, 348, 208, 98, "Physics"}, {242, 348, 214, 98, "Force Fields"},
	}
	onAction := func(a string) *origTransition {
		if idx, ok := researchAreaOrder[a]; ok && b.session != nil {
			if t, _, done := currentAreaTopic(b.session, idx); !done {
				b.session.SetResearchTopic(t) // 實際設定研究主題,結束回合朝此累積
			}
		}
		return b.goTo(b.galaxy, "星系主畫面")
	}
	// 研究領域標籤擦底疊字(座標為 bg 局部座標,472×480;draw 時自動加置中偏移)。
	// y=27/131/237/343 為 PIL 量測(openorion2 無按鈕 y 字面,均距推導一致,保留)。
	// 右欄 x 由 PIL 244 校正為 openorion2 tech.cpp 按鈕圖真值 248(左欄 22 與真值 21 僅差 1px,保留)。
	overlays := []labelRect{
		{155, 9, 162, 18, "Select New Research", 0},
		{22, 27, 128, 18, "Construction", 0},
		{248, 27, 124, 18, "Power", 0},
		{22, 131, 128, 18, "Chemistry", 0},
		{248, 131, 124, 18, "Sociology", 0},
		{22, 237, 128, 18, "Computers", 0},
		{248, 237, 124, 18, "Biology", 0},
		{22, 343, 128, 18, "Physics", 0},
		{248, 343, 124, 18, "Force Fields", 0},
	}
	s, err := loadOverlayScreen(b.res, "techsel.lbx", 0, b.lang, b.fnt, "assets/i18n/tech.tsv",
		overlays, color.RGBA{210, 216, 230, 255}, 13, hits, onAction,
		paletteChain{{"science.lbx", 0}})
	if err != nil {
		return nil, err
	}
	// 領域框內疊「目前主題 ・ RP 成本」(取代原本盲選):文字置中在標題帶下方的空白區。
	if b.session != nil {
		gold := color.RGBA{235, 210, 130, 255}
		body := color.RGBA{200, 214, 232, 255}
		for _, h := range hits {
			idx, ok := researchAreaOrder[h.action]
			if !ok {
				continue
			}
			t, cost, done := currentAreaTopic(b.session, idx)
			label := fmt.Sprintf("%s ・ %d RP", topicNameZh(b.lang, t), cost)
			col := body
			if done {
				label, col = b.tr("已完成本領域全部科技", "All technologies in this field are complete"), gold
			}
			cx := float64(h.x) + float64(h.w)/2
			cy := float64(h.y) + 40 // 標題帶(高18)下方留白處置中
			s.extras = append(s.extras, extraText{x: cx, y: cy, size: 12, text: label, col: col, align: 1})
		}
	}
	return s, nil
}

// planetListRows 是行星列表一次顯示幾列(openorion2 PLANET_LIST_ROWS)。
const planetListRows = 8

// planetListRowY 是各列的中心 y(openorion2 galaxy.cpp PlanetsListView 真值:
// FIRST_ROW 36 + ROW_HEIGHT 50/2 + i×ROW_DIST 55 = 61 + i×55)。
var planetListRowY = [planetListRows]float64{61, 116, 171, 226, 281, 336, 391, 446}

// visiblePlanets 回傳行星列表要列出的行星索引:**目前看得見的星系**裡的所有天體,
// 依星、再依軌道排序。
//
// 為什麼要過濾:那個畫面在原版是玩家的行星總覽,列出看不見的星系的行星等於免費送情報。
// 用的是與星圖同一套可見性(shell.VisibleStars:已探索/自己的/在偵測範圍內)——
// 星圖上點得到的星,它的行星就該出現在這裡,兩邊不該各有一套規則。
func (b *sceneBuilder) visiblePlanets() []int {
	if b.session == nil {
		return nil
	}
	vis := b.session.VisibleStars()
	out := make([]int, 0, len(b.session.Planets))
	for i := range b.session.Stars {
		if i < len(vis) && !vis[i] {
			continue
		}
		out = append(out, b.session.PlanetsAt(i)...)
	}
	return out
}

// planetListPage 把捲動起點夾在合法範圍內,並讓「選中的那顆行星」落在可見的 8 列裡——
// 從星圖選了一顆星再進這個畫面時,要直接看得到那個星系,而不是永遠停在第一頁。
func (b *sceneBuilder) planetListPage(list []int) int {
	top := b.planetListTop
	if sel := b.planetPick; sel >= 0 {
		at := -1
		for i, p := range list {
			if p == sel {
				at = i
				break
			}
		}
		if at >= 0 && (at < top || at >= top+planetListRows) {
			top = at - at%planetListRows
		}
	}
	if max := len(list) - planetListRows; top > max {
		top = max
	}
	if top < 0 {
		top = 0
	}
	return top
}

// planets 建原版行星列表畫面。
//
// 這是原版**選行星**的地方:畫面右下角那兩顆 SEND COLONY SHIP / SEND OUTPOST SHIP
// 就是對著選中的那一列作用的(星圖上的按鈕只是「這個星系的第一顆」捷徑)。
// 同一個星系可以有多個殖民地之後,這條路徑才是完整的——星圖面板擠不下逐行星的選擇。
func (b *sceneBuilder) planets() (*overlayScreen, error) {
	list := b.visiblePlanets()
	// 進畫面時預選星圖上選中的那顆星的代表行星,省得每次都要先點一列。
	if b.session != nil && b.planetPick < 0 && b.session.SelectedStar >= 0 {
		b.planetPick = b.session.PlanetAt(b.session.SelectedStar)
	}
	top := b.planetListPage(list)
	b.planetListTop = top

	hits := []hitRegion{
		{454, 440, 157, 23, "Return"},
		{454, 386, 156, 23, "sendcolony"},
		{454, 413, 157, 25, "sendoutpost"},
	}
	// 每一列一個熱區(列高 50,中心 y 見 planetListRowY)。
	for i := 0; i < planetListRows && top+i < len(list); i++ {
		hits = append(hits, hitRegion{16, int(planetListRowY[i]) - 25, 398, 50, fmt.Sprintf("prow%d", i)})
	}
	onAction := func(a string) *origTransition {
		if a == "Return" {
			return b.goTo(b.galaxy, "星系主畫面")
		}
		if strings.HasPrefix(a, "prow") {
			if i, err := strconv.Atoi(a[4:]); err == nil && top+i < len(list) {
				b.planetPick = list[top+i]
				b.planetListMsg = ""
			}
			return b.goTo(b.planets, "行星列表")
		}
		if (a == "sendcolony" || a == "sendoutpost") && b.session != nil {
			b.planetListMsg = b.planetListAction(a)
			return b.goTo(b.planets, "行星列表")
		}
		return nil
	}
	s, err := loadOverlayScreen(b.res, "plntsum.lbx", 0, b.lang, b.fnt, "assets/i18n/planets.tsv",
		planetsOverlays, color.RGBA{206, 218, 240, 255}, 14, hits, onAction, nil)
	if err != nil {
		return nil, err
	}
	// 即時行星資料填進表格列(欄位中心 x 對齊標題;列中心 y 見 planetListRowY)。
	if b.session != nil {
		body := color.RGBA{206, 218, 240, 255}
		cx := struct{ name, cli, grv, min, siz float64 }{57, 136, 218, 303, 382}
		for i := 0; i < planetListRows && top+i < len(list); i++ {
			pi := list[top+i]
			p := b.session.Planets[pi]
			y := planetListRowY[i]
			col := body
			if pi == b.planetPick {
				col = color.RGBA{250, 230, 140, 255} // 選中的那一列換色(這個畫面沒有可畫的選取框)
			}
			s.extras = append(s.extras,
				// 行星名欄寬 78(標題列 {18,14,78,18});長名字要截,否則會溢出欄框。
				extraText{x: cx.name, y: y, size: 12, text: truncateToWidth(b.fnt, p.Name, 12, 74), col: col, align: 1},
				extraText{x: cx.cli, y: y, size: 12, text: p.Climate, col: col, align: 1},
				extraText{x: cx.grv, y: y, size: 12, text: p.Gravity, col: col, align: 1},
				extraText{x: cx.min, y: y, size: 12, text: p.Mineral, col: col, align: 1},
				extraText{x: cx.siz, y: y, size: 12, text: p.Size, col: col, align: 1},
			)
			// 兩項原版有、remake 先前完全沒顯示的資訊,各自塞進對應欄位格子的第二行
			// (格高 50,主文字置中,y+11 這一行還在格內)。
			//
			// ⚠ 不要往 x>410 放:那裡是原版的排序/篩選面板,不是空白區——2026-08-06 試過一次,
			// 文字直接疊在面板按鈕上。
			sub := color.RGBA{150, 172, 205, 255}
			star := b.session.PlanetStar(pi)
			// 星名 + 該星系天體數:多天體之後「這顆行星屬於哪個星系」不再能從行星名推出來。
			if star >= 0 {
				line := b.session.Stars[star].Name
				if n := b.session.SystemBodyCountText(star); n != "" {
					line += " " + n
				}
				// 欄寬 78(標題列 {18,14,78,18}),超出會往左溢到欄框外——截掉。
				s.extras = append(s.extras, extraText{x: cx.name, y: y + 11, size: 9,
					text: truncateToWidth(b.fnt, line, 9, 74), col: sub, align: 1})
			}
			// 這顆行星目前的狀態(自己的殖民地/前哨站),那是「還能不能派船過去」的關鍵資訊。
			if ci := b.session.ColonyIndexOnPlanet(pi); ci >= 0 {
				s.extras = append(s.extras, extraText{x: cx.siz, y: y + 11, size: 9,
					text: b.tr("● 已殖民", "● colony"), col: color.RGBA{150, 225, 165, 255}, align: 1})
			} else if b.session.HasOutpostOnPlanet(pi) {
				s.extras = append(s.extras, extraText{x: cx.siz, y: y + 11, size: 9,
					text: b.tr("● 前哨站", "● outpost"), col: color.RGBA{150, 195, 235, 255}, align: 1})
			}
			if sp := gamedata.PlanetSpecialName(p.SpecialID); sp != "" {
				s.extras = append(s.extras, extraText{x: cx.min, y: y + 11, size: 9, text: "★" + sp, col: sub, align: 1})
			}
		}
		if len(list) == 0 {
			s.extras = append(s.extras, extraText{x: 210, y: 61, size: 12,
				text: b.tr("尚未探索任何星系", "No systems explored yet"), col: body, align: 1})
		}
		// 動作結果訊息:壓在兩顆動作鈕上方那條空白帶。
		if b.planetListMsg != "" {
			s.extras = append(s.extras, extraText{x: 532, y: 376, size: 9, text: b.planetListMsg,
				col: color.RGBA{240, 215, 150, 255}, align: 1})
		}
	}
	return s, nil
}

// planetListAction 執行行星列表右下角那兩顆鈕,回傳要顯示給玩家的結果訊息。
//
// 艦隊不在該星系時**先把艦隊派過去**——原版那兩顆鈕的字面意思就是「派船過去」,
// 一步到位;要玩家自己先回星圖點派遣、下回合再回來按一次是 remake 的贅步。
func (b *sceneBuilder) planetListAction(action string) string {
	sess := b.session
	if b.planetPick < 0 || b.planetPick >= len(sess.Planets) {
		return b.tr("先點一列選一顆行星", "Pick a planet row first")
	}
	star := sess.PlanetStar(b.planetPick)
	if star < 0 {
		return b.tr("那顆天體不屬於任何星系", "That body belongs to no system")
	}
	colony := action == "sendcolony"
	if colony && !sess.FleetHasColonyShip() {
		return b.tr("艦隊未載運殖民船", "No colony ship in the fleet")
	}
	if !colony && !sess.FleetHasOutpostShip() {
		return b.tr("艦隊未載運前哨船", "No outpost ship in the fleet")
	}
	if sess.Fleet().AtStar != star || sess.Fleet().ETA != 0 {
		if sess.Fleet().ETA > 0 {
			return fmt.Sprintf(b.tr("艦隊航行中…剩 %d 回合", "Fleet in transit — %d turns left"), sess.Fleet().ETA)
		}
		if !sess.SendFleet(star) {
			return b.tr("艦隊無法前往該星系", "The fleet cannot reach that system")
		}
		return fmt.Sprintf(b.tr("艦隊已出發前往 %s(%d 回合後抵達)", "Fleet en route to %s (%d turns)"),
			sess.Stars[star].Name, sess.Fleet().ETA)
	}
	if colony {
		res := sess.ColonizePlanet(b.planetPick)
		if !res.Ok {
			return res.Reason
		}
		return fmt.Sprintf(b.tr("%s 拓殖成功——起始人口 %d(上限 %d)", "Colonized %s — pop %d (max %d)"),
			sess.Planets[b.planetPick].Name, res.StartPopulation, res.PopMax)
	}
	res := sess.BuildOutpostOnPlanet(b.planetPick)
	if !res.Ok {
		return res.Reason
	}
	return fmt.Sprintf(b.tr("%s 前哨站建立完成(掃描站,無產出)", "Outpost established on %s (scanner, no output)"),
		sess.Planets[b.planetPick].Name)
}

// --- interactiveApp(ebiten.Game;支援 headless 腳本驗證)---

type interactiveApp struct {
	cur origScreen
	b   *sceneBuilder // 每幀把 tick 同步給它當動畫計數(見 sceneBuilder.animTick)

	// headless 驗證:script 逐幀注入輸入,跑滿 frames 存 shot。
	script   []shell.InputState
	shotPath string
	frames   int
	tick     int
	saved    bool
	scale    int // 目前視窗放大倍率(1~4)

	audio *moo2audio.Mixer // 持有音訊 Mixer,避免 player 被 GC(headless 為 nil)

	// 過場截圖廊(-gamegallery):script 為導覽腳本,galleryShots 指定在哪個絕對 tick
	// 存哪張圖(可多張,依序達成)。與單張 shotPath 模式互斥。
	galleryDir   string
	galleryShots []galleryShot
	galleryDone  int
	// galleryVictoryTick 是截圖廊專用:在這個 tick 把對局設成「已分出勝負」,好讓導覽腳本
	// 走得到最終得分畫面。與上面 EventSeed=1 同一個理由——那條路徑靠正常遊玩要好幾百回合,
	// 截圖驗證等不起。只在截圖廊模式下設值,正常遊戲恆為 0(不觸發)。
	galleryVictoryTick int
	// galleryFleetTick 是截圖廊在哪個 tick 給艦隊注入結構損傷 + 次元傳送門
	// (見該常數的說明:必須晚於最後一次結束回合,否則損傷會被停靠母星的完全修復清掉)。
	galleryFleetTick int
	// galleryGroundTick 是截圖廊在哪個 tick 直接把畫面換成地面戰戰報(見該常數說明)。
	galleryGroundTick int
	// galleryLoadWinTick 是截圖廊在哪個 tick 切到載入遊戲視窗(見該常數說明)。
	galleryLoadWinTick int
	// galleryGameMenuTick 是截圖廊在哪個 tick 切到遊戲選單視窗(見該常數說明)。
	galleryGameMenuTick int
	// galleryBombTick 是截圖廊在哪個 tick 切到軌道轟炸畫面(見該常數說明)。
	galleryBombTick int
	// galleryIntroTick 是截圖廊在哪個 tick 切到片頭過場(見該常數說明)。
	galleryIntroTick int
	// galleryEndingTick 是截圖廊在哪個 tick 切到結局過場(見該常數說明)。
	galleryEndingTick int
	// galleryMultiTick / galleryHotseatTick 是截圖廊切到多人設定畫面 / 熱座交接畫面的 tick。
	galleryMultiTick   int
	galleryHotseatTick int
	// galleryDesignTick 是截圖廊切到艦艇設計畫面的 tick。
	galleryDesignTick int
	// galleryBuildPopupTick 是截圖廊切到建造彈出視窗的 tick。
	galleryBuildPopupTick int
	// galleryCommandPointsTick 是截圖廊切到指揮點數視窗的 tick。
	galleryCommandPointsTick int
	// galleryMeasureTick 是截圖廊切回星圖並打開 F9 測距的 tick。
	galleryMeasureTick int
	galleryBuilder     *sceneBuilder
	gallerySession     *shell.GameSession
}

// galleryVictoryTick 是截圖廊在哪個 tick 把對局設成「已分出勝負」——必須早於腳本裡
// 「按 TURN 進最終得分」那一拍(t29),取它的前一拍。
const galleryVictoryTick = 38

// galleryFleetTick 是截圖廊在哪個 tick 給艦隊注入結構損傷 + 次元傳送門——取「進艦隊列表」
// 那一拍(t19)的前一拍。
//
// ⚠ 這一拍必須晚於腳本裡最後一次「結束回合」(t17 之前那次),否則 EndTurn 的
// advanceShipRepair 會把傷清光:艦隊開局就停在母星,照原版 Repair_Ships_At_Colonies_
// 的規則會被**完全修復**。先前注入在 t28 而 t29 按了結束回合,截出來一艘傷都沒有——
// 那不是顯示壞了,是修復規則正常運作。
const galleryFleetTick = 18

// galleryGroundTick 是截圖廊在哪個 tick 把畫面換成地面戰戰報——取截圖那一拍(t68)的前一拍。
const galleryGroundTick = 67

// galleryLoadWinTick 是截圖廊在哪個 tick 寫兩格示範存檔並切到載入視窗——取截圖(t70)的前一拍。
// 存檔寫到 `MOO2_SAVE_DIR` 指的暫存目錄(見 saveDirFor),不碰玩家真正的存檔。
const galleryLoadWinTick = 69

// galleryGameMenuTick 是截圖廊在哪個 tick 切到遊戲選單視窗——取截圖(t72)的前一拍。
const galleryGameMenuTick = 71

// galleryBombTick 是截圖廊在哪個 tick 切到軌道轟炸畫面——取截圖(t74)的前一拍。
const galleryBombTick = 73

// galleryIntroTick 是截圖廊在哪個 tick 切到片頭過場——留三拍讓它播出幾幀再截圖
// (第一幀是黑的,截了看不出解碼有沒有成功)。
const galleryIntroTick = 75

// galleryIntroSeekFrames 是截圖廊要快轉幾幀才截圖(取一個已確認有內容的畫面)。
const galleryIntroSeekFrames = 350

// galleryEndingTick 是截圖廊在哪個 tick 切到結局過場;同樣要快轉(第一幀是黑的)。
const galleryEndingTick = 79

// galleryEndingSeekFrames 是結局過場要快轉幾幀(ANWINFIN 共 323 幀,取中後段)。
const galleryEndingSeekFrames = 220

// galleryMultiTick 是截圖廊在哪個 tick 切到多人遊戲設定畫面——取截圖(t82)的前一拍。
// 走正常路徑是「主選單 → MULTI PLAYER」,但截圖廊此刻停在結局過場,直接推上來比繞回去可靠。
const galleryMultiTick = 81

// galleryHotseatTick 是截圖廊在哪個 tick 切到熱座交接畫面——取截圖(t84)的前一拍。
// 這一拍會順手把對局設成兩席熱座,交接畫面才有真的席位名可顯示。
const galleryHotseatTick = 83

// galleryDesignTick 是截圖廊在哪個 tick 切到艦艇設計畫面——取截圖(t86)的前一拍。
// 這個畫面先前**從沒被截圖廊拍過**(要從艦隊列表點進去,而腳本沒走那一步),
// 所以版面錯了也不會被發現——與 NEW GAME 同一個盲點。
const galleryDesignTick = 85

// galleryBuildPopupTick 是截圖廊在哪個 tick 切到建造彈出視窗——取截圖(t88)的前一拍。
const galleryBuildPopupTick = 87

// galleryCommandPointsTick 是截圖廊在哪個 tick 切到指揮點數視窗——取截圖(t90)的前一拍。
// 走正常路徑是星圖點右欄第 2 格,但腳本此刻停在建造視窗,直接推上來比重新導覽回星圖可靠。
const galleryCommandPointsTick = 89

// galleryMeasureTick 是截圖廊在哪個 tick 切回星圖並打開 F9 測距——取截圖(t92)的前一拍。
// 走正常路徑要從指揮點數視窗關回星圖再按 F9,直接推上來比重新導覽可靠(同上面幾個)。
const galleryMeasureTick = 91

// galleryRelocTick 是截圖廊在哪個 tick 設好一個集結點(好讓遷移連線真的畫得出來)。
// 排在測距那一拍之前:兩者都在星圖上,連線畫在星星之下,不會互相蓋掉。
const galleryRelocTick = 91

// galleryRelocTo 是示範用的集結點星索引;和測距一樣不能寫死,要挑一顆**可見的**
// (星圖有戰爭迷霧,固定索引很可能落在畫不出來的星上)。見 galleryMeasureTarget 同款理由。

// galleryMeasureFrom 是測距的起點星(0 = 玩家母星,每個 seed 都有)。
// 終點**不能寫死索引**——要驗的是「游標移到某顆星上就顯示距離」,而星圖有戰爭迷霧,
// 隨便一個索引很可能是還沒探索、根本畫不出來的星(第一版寫死 1,截圖就停在
// 「移到另一顆星」的提示上,什麼距離都沒畫)。改成執行時挑一顆**可見的**。
const galleryMeasureFrom = 0

// galleryShot 是「端到端過場截圖廊」腳本中,在某個絕對 tick 存一張圖的指令。
type galleryShot struct {
	tick int
	name string
}

// buildGalleryScript 產生「主選單→新遊戲流程→星系主畫面→殖民地/研究/外交/戰鬥」的
// headless 導覽腳本,並標出各到達畫面該存圖的 tick。
//
// 座標換算依各畫面實作:
//   - overlayScreen 系(menu/newGameSetup/galaxy/colonySummary/info/research/races):
//     hitRegion 座標為背景局部座標,實際點擊座標 = 局部座標 + offsetX/offsetY
//     (offsetX=(640-bg寬)/2,小於整版寬時置中;見 loadOverlayScreen)。
//     menu/newGameSetup/galaxy/colonySummary/info/races 背景皆滿版 640×480(offset=0),
//     直接沿用 hitRegion 座標;research(techsel.lbx)背景 472×480(見該函式註解),
//     offsetX=84,座標需加上此偏移。
//   - raceSelectScreen/nameFlagScreen 為自繪滿版畫面(dst.DrawImage 無置中位移),
//     其 Rect 座標即為絕對螢幕座標,直接使用。
//   - diplomacyScreen/tacticalScreen 亦為自繪滿版畫面,同上。
func buildGalleryScript() ([]shell.InputState, []galleryShot) {
	click := func(x, y int) shell.InputState { return shell.InputState{MouseX: x, MouseY: y, ClickReleased: true} }
	idle := shell.InputState{}
	// 每次導覽點擊後補 idle「settle 幀」再截圖:星系生星、殖民地載面板、研究列表
	// 都需要至少一個 tick 才渲染完成,截圖 tick 落在轉場後 2 幀,避免抓到半載/前一畫面。
	// (截圖已改在 Update 精確 tick 用 offscreen 渲染,不再受 ebiten 跳 Draw 影響,
	//  但 settle 幀仍保留以確保畫面內容本身已載入。)
	script := []shell.InputState{
		idle,            // t1: 主選單(未點擊)→ 截圖 menu
		click(491, 228), // t2: 主選單「新遊戲」→ 新遊戲設定
		idle,            // t3: settle
		click(486, 405), // t4: 新遊戲設定「Accept」→ 種族選擇
		idle,            // t5: settle
		idle,            // t6: settle → 截圖 raceselect
		// t7: 種族選擇——**點種族鈕即確認**(原版沒有 ACCEPT,見 raceselect.go 檔頭)。
		// 人類是清單第 5 項 → 第 0 欄第 5 列 → x 351..474、y 330..375。
		// (先前這裡點的是 (540,451) 的「接受」鈕,2026-08-07 版面改成原版的 2×7 網格後
		//  那個座標什麼都不會命中,腳本會卡在種族畫面、後面每一張都截錯。)
		click(410, 350), // t7: 種族選擇「人類」→ 命名/旗色
		idle,            // t8: settle → 截圖 nameflag
		click(540, 454), // t9: 命名/旗色「接受」→ 星系主畫面
		idle,            // t10: settle
		idle,            // t11: settle → 截圖 galaxy

		// 事件快報排在最前面:「星系主畫面按 TURN」是最短、最不依賴前置狀態的路徑,
		// 截圖廊固定了會觸發事件的 seed(見 runInteractive),所以這一步必定走到快報畫面。
		// (先前把它排在戰鬥之後,只要中間任何一步的座標過期就整串偏掉——2026-08-06 踩過。)
		click(589, 458), // t12: 「結束回合」→ 事件快報
		idle,            // t13: settle
		idle,            // t14: settle → 截圖 event
		click(320, 384), // t15: 事件快報「繼續」→ 回合摘要
		idle,            // t16: settle → 截圖 turnsummary
		click(320, 393), // t17: 回合摘要「關閉」→ 星系主畫面
		idle,            // t18: settle

		// 艦隊列表 + 安塔蘭王座廳。排在最終得分**之前**,王座廳才照得到「按鈕可用」的樣子
		// ——勝負一旦定了,前置條件第一條就先擋掉(見 AssaultAntaresBlockReason)。
		// 艦艇損傷與次元傳送門由 galleryFleetTick 在這一段前注入(見 Update)。
		click(198, 452), // t19: 工具列「FLEETS」→ 艦隊列表
		idle,            // t20: settle
		idle,            // t21: settle → 截圖 fleet(含艦艇損傷顯示)
		click(150, 398), // t22: 艦隊列表左下「攻打安塔蘭」→ 安塔蘭王座廳
		idle,            // t23: settle
		idle,            // t24: settle → 截圖 antaranroom
		click(449, 418), // t25: 王座廳「撤退」→ 艦隊列表
		idle,            // t26: settle
		click(598, 444), // t27: 艦隊列表「RETURN」→ 星系主畫面
		idle,            // t28: settle

		click(48, 452),  // t29: 星系主畫面工具列「殖民地」→ 殖民地總覽
		idle,            // t30: settle
		idle,            // t31: settle → 截圖 colonysummary
		click(50, 47),   // t32: 總覽第一列「殖民地名」欄 → 單一殖民地畫面
		idle,            // t33: settle
		idle,            // t34: settle → 截圖 colonyscreen
		click(590, 459), // t35: 殖民地畫面「返回」→ 殖民地總覽
		idle,            // t36: settle
		click(608, 462), // t37: 殖民地總覽「RETURN」→ 星系主畫面
		idle,            // t38: settle

		// 最終得分畫面:對局分出勝負後按 TURN 就會進去(見 interactive.go 的 turn 分支)。
		// 勝負由 galleryVictoryTick 在下面第一拍前設好——那條路徑正常玩要幾百回合才走得到,
		// 截圖驗證等不起,與上面固定 EventSeed 是同一個理由。
		click(589, 458), // t39: 「結束回合」→ 最終得分
		idle,            // t40: settle
		idle,            // t41: settle → 截圖 hiscore
		click(320, 400), // t42: 最終得分「繼續」→ 回合摘要
		idle,            // t43: settle
		click(320, 393), // t44: 回合摘要「關閉」→ 星系主畫面
		idle,            // t45: settle

		click(123, 450), // t46: 工具列「PLANETS」→ 行星列表
		idle,            // t47: settle
		idle,            // t48: settle → 截圖 planets
		click(532, 451), // t49: 行星列表「返回」→ 星系主畫面
		idle,            // t50: settle

		click(495, 452), // t51: 工具列「INFO」→ 情報畫面
		idle,            // t52: settle
		idle,            // t53: settle → 截圖 info(預設分頁:歷史圖表)
		click(21, 80),   // t54: INFO 左欄「科技總覽」分頁
		idle,            // t55: settle → 截圖 info_tech
		click(535, 434), // t56: INFO「RETURN」→ 星系主畫面
		idle,            // t57: settle

		click(420, 452), // t58: 工具列「RACES」→ 種族關係
		idle,            // t59: settle
		click(483, 428), // t60: 種族關係「REPORT」→ 外交對談
		idle,            // t61: settle
		idle,            // t62: settle → 截圖 diplomacy
		click(320, 437), // t63: 外交對談「結束對談」→ 種族關係
		click(388, 448), // t64: 種族關係「DECLARE WAR」→ 戰術戰鬥
		idle,            // t65: settle
		idle,            // t66: settle → 截圖 tactical

		// 地面戰畫面(原版 Colony_Combat)。走正常玩家路徑要「艦隊飛到敵殖民地星 + 載陸戰隊
		// + 在星資訊面板點入侵」,而敵星在星圖上的位置隨 seed 變動,腳本點不準;
		// 故由 galleryGroundTick 直接把畫面推上來(見 Update)。
		idle, // t67: 由 galleryGroundTick 換成地面戰畫面
		idle, // t68: settle → 截圖 groundcombat

		// 載入遊戲視窗(原版 Mainmenu_Load_Game_Popup_)。同樣由 tick 直接推上來——
		// 走正常路徑要「先有存檔、回主選單、點 Load Game」,而截圖廊是一路往前走的單向腳本。
		idle, // t69: 由 galleryLoadWinTick 寫示範存檔並換成載入視窗
		idle, // t70: settle → 截圖 loadgame

		// 遊戲選單視窗(原版 GameMenuWindow)。從星系主畫面點頂端「遊戲」鈕進得去,
		// 但截圖廊此刻停在載入視窗,直接推上來比重新導覽回星系可靠。
		idle, // t71: 由 galleryGameMenuTick 換成遊戲選單
		idle, // t72: settle → 截圖 gamemenu

		// 軌道轟炸畫面(原版 Colony_Bombing_Screen_)。與地面戰同理:走正常路徑要艦隊飛到
		// 敵殖民地星、在星資訊面板點轟炸,敵星位置隨 seed 變動,腳本點不準。
		idle, // t73: 由 galleryBombTick 換成轟炸畫面
		idle, // t74: settle → 截圖 bombing

		// 片頭過場(原版 Smack)。放最後:它會一直往前播,插在中間會吃掉後續的 tick。
		idle, // t75: 由 galleryIntroTick 換成片頭
		idle, // t76: settle
		idle, // t77: settle
		idle, // t78: settle → 截圖 intro(已播幾幀,不是全黑的第一幀)

		// 結局過場(勝負分出後、進最終得分前播的那一段)。同樣由 tick 注入 + 快轉。
		idle, // t79: 由 galleryEndingTick 換成結局過場
		idle, // t80: settle → 截圖 ending

		// 多人遊戲設定(原版 Multi_Player_Screen_)與熱座交接(Hotseat_Screen_)。
		idle, // t81: 由 galleryMultiTick 換成多人設定畫面
		idle, // t82: settle → 截圖 multiplayer
		idle, // t83: 由 galleryHotseatTick 設成兩席熱座並換成交接畫面
		idle, // t84: settle → 截圖 hotseat

		idle, // t85: 由 galleryDesignTick 換成艦艇設計畫面
		idle, // t86: settle → 截圖 shipdesign

		// 建造彈出視窗(原版 Build_Queue_Popup_)。走正常路徑是殖民地畫面點 CHANGE,
		// 但截圖廊此刻停在艦艇設計,直接推上來比重新導覽回殖民地可靠。
		idle, // t87: 由 galleryBuildPopupTick 換成建造視窗
		idle, // t88: settle → 截圖 buildqueue

		// 指揮點數視窗(原版 Show_Command_Points_Screen_)。同上,直接推上來。
		idle, // t89: 由 galleryCommandPointsTick 換成指揮點數視窗
		idle, // t90: settle → 截圖 commandpoints

		// F9 測距(手冊:點第一顆星,游標移到另一顆就顯示秒差距)。畫面由
		// galleryMeasureTick 推回星圖並設好起點;這裡只負責把游標放到第二顆星上,
		// 因為「移到哪就顯示到哪」正是要驗的行為——沒有游標位置就什麼都不會畫。
		idle, // t91: 由 galleryMeasureTick 切回星圖 + 打開測距
		idle, // t92: settle(游標由 galleryMeasureHover 每幀注入)→ 截圖 measure
	}
	shots := []galleryShot{
		{1, "01_menu.png"},
		// NEW GAME 設定畫面先前從沒被截圖廊拍過(腳本從主選單直接點到種族選擇),
		// 所以那個畫面的版面錯了也不會被發現。名字用 01b 而不是重編後面 23 張的號。
		{3, "01b_newgame.png"},
		{6, "02_raceselect.png"},
		{8, "03_nameflag.png"},
		{11, "04_galaxy.png"},
		{14, "05_event.png"},
		{16, "06_turnsummary.png"},
		{21, "07_fleet.png"},
		{24, "08_antaranroom.png"},
		{31, "09_colonysummary.png"},
		{34, "10_colonyscreen.png"},
		{41, "11_hiscore.png"},
		{48, "12_planets.png"},
		{53, "13_info.png"},
		{55, "14_info_tech.png"},
		{62, "15_diplomacy.png"},
		{66, "16_tactical.png"},
		{68, "17_groundcombat.png"},
		{70, "18_loadgame.png"},
		{72, "19_gamemenu.png"},
		{74, "20_bombing.png"},
		{78, "21_intro.png"},
		{80, "22_ending.png"},
		{82, "23_multiplayer.png"},
		{84, "24_hotseat.png"},
		{86, "25_shipdesign.png"},
		{88, "26_buildqueue.png"},
		{90, "27_commandpoints.png"},
		{92, "28_measure.png"},
	}
	return script, shots
}

// handleWindowKeys 處理縮放/全螢幕快捷鍵:+/- 調整放大倍率(1~4)、F11 或 F 切換全螢幕。
func (a *interactiveApp) handleWindowKeys() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) || inpututil.IsKeyJustPressed(ebiten.KeyF) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
		return
	}
	if ebiten.IsFullscreen() {
		return // 全螢幕時 +/- 不改視窗大小
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEqual) || inpututil.IsKeyJustPressed(ebiten.KeyKPAdd) {
		a.setScale(a.scale + 1)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyMinus) || inpututil.IsKeyJustPressed(ebiten.KeyKPSubtract) {
		a.setScale(a.scale - 1)
	}
}

// setScale 設定視窗放大倍率(夾在 1~4),依邏輯 640×480 等比放大。
func (a *interactiveApp) setScale(s int) {
	if s < 1 {
		s = 1
	}
	if s > 4 {
		s = 4
	}
	if s == a.scale {
		return
	}
	a.scale = s
	ebiten.SetWindowSize(moo2ScreenW*s, moo2ScreenH*s)
}

func (a *interactiveApp) pollInput() shell.InputState {
	if a.script != nil {
		in := shell.InputState{}
		if idx := a.tick - 1; idx >= 0 && idx < len(a.script) {
			in = a.script[idx]
		}
		// 截圖廊的 F9 測距:游標位置得指到第二顆星上,而星的螢幕座標隨 seed 變動,
		// 靜態腳本寫不死。測距打開之後就每幀算一次注入進去。
		if a.galleryMeasureTick > 0 && a.tick >= a.galleryMeasureTick && a.galleryBuilder != nil {
			if idx := a.galleryBuilder.galleryMeasureTarget(); idx >= 0 {
				in.MouseX, in.MouseY = starScreenPos(a.galleryBuilder.session.Stars[idx])
			}
		}
		return in
	}
	x, y := ebiten.CursorPosition()
	return shell.InputState{
		MouseX: x, MouseY: y,
		ClickReleased: inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft),
		Hotkey:        pollHotkey(),
	}
}

func (a *interactiveApp) Update() error {
	a.tick++
	if a.b != nil {
		a.b.animTick = a.tick // 動畫計數(黑洞旋渦等),見 starsprite.go
	}
	if a.script == nil { // 互動模式才處理視窗快捷鍵(headless 略過)
		a.handleWindowKeys()
	}
	// 截圖廊專用:到了指定 tick 把對局設成已分出勝負,好讓導覽腳本走得到最終得分畫面
	// (見 galleryVictoryTick 欄位註解)。
	if a.galleryVictoryTick > 0 && a.tick == a.galleryVictoryTick && a.gallerySession != nil {
		a.gallerySession.Victory = shell.VictoryState{
			Over: true, Reason: engine.VictoryAntaran, Winner: "player", Turn: a.gallerySession.Turn,
		}
		a.gallerySession.AntaranHomeworldConquered = true
	}
	// 截圖廊專用:艦隊列表勾兩艘船,讓「拆成新艦隊」那一行出現在截圖裡
	// (不勾就永遠看不到,那一層等於沒被驗到)。
	if a.galleryFleetTick > 0 && a.tick == a.galleryFleetTick && a.galleryBuilder != nil {
		a.galleryBuilder.shipPick = map[int]bool{0: true}
	}
	// 截圖廊專用:艦隊列表與安塔蘭王座廳的前置狀態(見 galleryFleetTick 的說明)。
	// 與上面同一個理由——截圖驗證等不起真的打一場、也等不起研究到多維物理再蓋傳送門。
	if a.galleryFleetTick > 0 && a.tick == a.galleryFleetTick && a.gallerySession != nil {
		// ① 艦艇損傷:輕傷/完好/重傷各一,琥珀與紅兩個顏色分支都要真的畫出來才算驗到。
		// 開局艦隊都是 strength 1 的船(偵察艦/殖民船,最大血量 3),損傷刻度本來就只有
		// 0/33/66% 三檔——這是艦級抽象戰鬥模型的必然,不是顯示精度問題。
		i := 0
		a.gallerySession.EachShip(func(sh *shell.Ship) {
			switch i % 3 {
			case 0:
				sh.Damage = 1 // 輕傷 → 琥珀
			case 2:
				sh.Damage = 99 // 重傷 → 紅(ShipDamage 會夾到「最大血量−1」)
			}
			i++
		})
		// ② 次元傳送門:安塔蘭王座廳的「發動終局反攻」要它才會亮起來(手冊 p.183)。
		a.gallerySession.GrantDimensionalPortalForGallery()
	}
	// 截圖廊專用:把畫面直接換成地面戰戰報。走正常路徑要艦隊飛到敵殖民地星、載滿陸戰隊、
	// 再在星資訊面板點入侵,而敵星在星圖上的位置隨 seed 變動,點擊腳本對不準。
	// 這裡餵的是一組固定的示範戰果(不呼叫 InvadeColony,不動遊戲狀態)。
	if a.galleryGroundTick > 0 && a.tick == a.galleryGroundTick && a.galleryBuilder != nil {
		demo := shell.GroundInvasionResult{
			Ok: true, AttackerWon: true, StarCaptured: true,
			ColonyName:           a.galleryBuilder.tr("示範殖民地", "Demo Colony"),
			AttackerMarinesStart: 6, AttackerTanksStart: 2, DefenderStart: 5,
			AttackerSurvived:        5,
			AttackerMarinesSurvived: 3, AttackerTanksSurvived: 2, DefenderSurvived: 0,
			Rounds: 4,
		}
		if sc, err := a.galleryBuilder.groundCombat(demo); err == nil {
			a.cur = sc
		}
	}
	// 截圖廊專用:載入遊戲視窗。先寫幾格示範存檔(空白的十格截起來看不出什麼),再切過去。
	if a.galleryLoadWinTick > 0 && a.tick == a.galleryLoadWinTick && a.galleryBuilder != nil && a.gallerySession != nil {
		dir := saveDirFor()
		for _, slot := range []int{1, 4, shell.AutoSaveSlot} {
			if err := a.gallerySession.Save(shell.SaveSlotPath(dir, slot)); err != nil {
				fmt.Fprintln(os.Stderr, "截圖廊寫示範存檔:", err)
			}
		}
		// 其中一格改寫成熱座局,好讓列表右側的兩種對局圖示都露出來。
		// 用「讀回來的獨立 session」設熱座,不動正在跑的 gallerySession——
		// SetupHotseat 會接管 AI 對手,直接對活的對局做會影響後面幾拍的畫面。
		if hs, err := shell.LoadSession(shell.SaveSlotPath(dir, 4)); err == nil {
			hs.SetupHotseat(2)
			if err := hs.Save(shell.SaveSlotPath(dir, 4)); err != nil {
				fmt.Fprintln(os.Stderr, "截圖廊寫熱座示範存檔:", err)
			}
		}
		if sc, err := a.galleryBuilder.loadGame(); err == nil {
			a.cur = sc
		}
	}
	// 截圖廊專用:遊戲選單視窗。
	if a.galleryGameMenuTick > 0 && a.tick == a.galleryGameMenuTick && a.galleryBuilder != nil {
		if sc, err := a.galleryBuilder.gameMenu(); err == nil {
			a.cur = sc
		}
	}
	// 截圖廊專用:軌道轟炸畫面。餵一組固定的示範戰果(不呼叫 BombardColony,不動遊戲狀態)。
	if a.galleryBombTick > 0 && a.tick == a.galleryBombTick && a.galleryBuilder != nil {
		demo := shell.GroundBombardResult{
			Ok: true, TotalDamage: 148, Hits: 7,
			ColonyName:         a.galleryBuilder.tr("示範殖民地", "Demo Colony"),
			BuildingsDestroyed: 3, BuildingsRemaining: 5,
			PopulationLost: 4, PopulationBefore: 12,
			DefenderRetaliated: true, AttackerShipsLost: 1,
		}
		if sc, err := a.galleryBuilder.bombing(demo); err == nil {
			a.cur = sc
		}
	}
	// 截圖廊專用:片頭過場。正常模式是開場就播,截圖廊為了不打亂 tick 才挪到最後。
	if a.galleryIntroTick > 0 && a.tick == a.galleryIntroTick && a.galleryBuilder != nil {
		if sc := a.galleryBuilder.intro(); sc != nil {
			// 快轉到有內容的一幀:片頭 76ms/幀而截圖廊一拍 16.7ms,照正常速度要等上百拍;
			// 第 0 幀又是全黑的(片頭從黑淡入),截了看不出解碼成功與否。
			if cs, ok := sc.(*cutsceneScreen); ok {
				cs.seekForGallery(galleryIntroSeekFrames)
			}
			a.cur = sc
		}
	}
	// 截圖廊專用:結局過場。正常模式是勝負分出後自動播,這裡直接推上來 + 快轉。
	if a.galleryEndingTick > 0 && a.tick == a.galleryEndingTick && a.galleryBuilder != nil {
		if sc := a.galleryBuilder.endingCutsceneFor(); sc != nil {
			if cs, ok := sc.(*cutsceneScreen); ok {
				cs.seekForGallery(galleryEndingSeekFrames)
			}
			a.cur = sc
		}
	}
	// 截圖廊專用:多人遊戲設定畫面(走正常路徑是主選單 → MULTI PLAYER)。
	if a.galleryMultiTick > 0 && a.tick == a.galleryMultiTick && a.galleryBuilder != nil {
		if sc, err := a.galleryBuilder.multiPlayer(); err == nil {
			a.cur = sc
		}
	}
	// 截圖廊專用:熱座交接畫面。先把對局設成兩席熱座,交接畫面才有真的席位名可顯示。
	if a.galleryHotseatTick > 0 && a.tick == a.galleryHotseatTick && a.galleryBuilder != nil && a.gallerySession != nil {
		a.gallerySession.SetupHotseat(2)
		if a.gallerySession.HotseatEnabled() {
			if t := a.galleryBuilder.endTurnPressed(); t != nil && t.next != nil {
				a.cur = t.next
			}
		}
	}
	// 截圖廊專用:艦艇設計畫面(走正常路徑是艦隊列表 → 點右側艦艇格)。
	if a.galleryDesignTick > 0 && a.tick == a.galleryDesignTick && a.galleryBuilder != nil {
		if sc, err := a.galleryBuilder.shipDesign(); err == nil {
			a.cur = sc
		}
	}
	// 截圖廊專用:建造彈出視窗(走正常路徑是殖民地畫面點 CHANGE)。
	if a.galleryBuildPopupTick > 0 && a.tick == a.galleryBuildPopupTick && a.galleryBuilder != nil {
		a.galleryBuilder.colonyIdx = 0
		if sc, err := a.galleryBuilder.buildQueuePopup(); err == nil {
			a.cur = sc
		}
	}
	// 截圖廊專用:指揮點數視窗(走正常路徑是星圖點右欄第 2 格)。
	if a.galleryCommandPointsTick > 0 && a.tick == a.galleryCommandPointsTick && a.galleryBuilder != nil {
		if sc, err := a.galleryBuilder.commandPoints(); err == nil {
			a.cur = sc
		}
	}
	// 截圖廊專用:切回星圖並打開 F9 測距,起點設在母星;順便設一個集結點,
	// 好讓遷移連線那一層真的畫得出來(不設就永遠是空的,截圖驗不到)。
	if a.galleryMeasureTick > 0 && a.tick == a.galleryMeasureTick && a.galleryBuilder != nil {
		b := a.galleryBuilder
		if sess := b.session; sess != nil {
			if to := b.galleryRelocTarget(); to >= 0 {
				// 玩家的第一個殖民地 → 第二顆看得到的星(第一顆被 F9 測距的示範用掉了)。
				// 起點寫 colonyStar(0) 而不是「母星」:截圖廊跑完幾回合後殖民地清單會變,
				// 第一個殖民地不保證還在星 0。
				sess.SetColonyRelocation(0, to)
			}
		}
		b.measure.on, b.measure.from = true, galleryMeasureFrom
		if sc, err := b.galaxy(); err == nil {
			a.cur = sc
		}
	}
	if t := a.cur.update(a.pollInput()); t != nil {
		if t.quit {
			return ebiten.Termination
		}
		if t.next != nil {
			a.cur = t.next
		}
	}
	if a.galleryDir != "" {
		// 截圖在 Update 的精確 tick 用 offscreen image 渲染,與 ebiten 的 Draw
		// 排程完全解耦:避免負載下 ebiten 跳 Draw、把多張錯過的 shot 批次補在
		// 同一幀(舊寫法會讓 04_galaxy 抓到 colony 內容 → 重複幀 bug)。
		for a.galleryDone < len(a.galleryShots) && a.tick >= a.galleryShots[a.galleryDone].tick {
			off := ebiten.NewImage(moo2ScreenW, moo2ScreenH)
			a.cur.draw(off)
			path := filepath.Join(a.galleryDir, a.galleryShots[a.galleryDone].name)
			if err := saveScreenshot(off, path); err != nil {
				fmt.Println("截圖失敗:", path, err)
			} else {
				fmt.Println("已存:", path)
			}
			off.Deallocate()
			a.galleryDone++
		}
		if a.galleryDone >= len(a.galleryShots) {
			return ebiten.Termination
		}
		// 硬性終止保護:即使某些圖因導覽失敗而存不到,超過最後一張的
		// 目標 tick(+緩衝)也一定結束,絕不留無限 render loop 空轉燒 CPU。
		if n := len(a.galleryShots); n > 0 && a.tick > a.galleryShots[n-1].tick+3 {
			return ebiten.Termination
		}
		return nil
	}
	if a.shotPath != "" && a.saved {
		return ebiten.Termination
	}
	return nil
}

func (a *interactiveApp) Draw(dst *ebiten.Image) {
	a.cur.draw(dst)
	if a.galleryDir != "" {
		for a.galleryDone < len(a.galleryShots) && a.tick >= a.galleryShots[a.galleryDone].tick {
			path := filepath.Join(a.galleryDir, a.galleryShots[a.galleryDone].name)
			if err := saveScreenshot(dst, path); err != nil {
				fmt.Println("截圖失敗:", path, err)
			} else {
				fmt.Println("已存:", path)
			}
			a.galleryDone++
		}
		return
	}
	if a.shotPath != "" && !a.saved && a.tick >= a.frames {
		if err := saveScreenshot(dst, a.shotPath); err != nil {
			fmt.Println("截圖失敗:", err)
		}
		a.saved = true
	}
}

func (a *interactiveApp) Layout(int, int) (int, int) { return moo2ScreenW, moo2ScreenH }

// runInteractive 啟動「還原原版」的互動遊戲。script/shot 非空時為 headless 驗證;
// galleryDir 非空時為「端到端過場截圖廊」模式(見 buildGalleryScript),優先於 script/shot。
func runInteractive(dirs []string, lang i18n.Lang, fnt, fntVec *uifont.Font,
	script []shell.InputState, shot string, frames int, galleryDir string) error {

	if lang == i18n.Traditional && fnt == nil {
		return fmt.Errorf("中文模式需以 -font 指定 CJK 字型")
	}
	res, err := assets.NewResolver(dirs...)
	if err != nil {
		return err
	}
	b := &sceneBuilder{res: res, fnt: fnt, fntVec: fntVec, lang: lang, session: shell.NewDemoSession(), newGameSize: 1, newGameDiff: newGameDiffDefault,
		newGameAge: newGameAgeDefault, newGameTech: newGameTechDefault, newGameEmpires: 1 + shell.DefaultOpponents, designWeapon: 1, savePath: savePathFor(), gameVersion: gamedata.VersionCommunity15,
		planetPick: -1} // −1 = 行星列表還沒選任何一列(0 是行星 0 的索引,不能當「沒選」)
	b.skipCutscenes = shot != "" || galleryDir != "" // 見該欄位註解
	// 傭兵候選池改用原版 HERODATA.LBX 真英雄(解析失敗自動退回內建策展名單,不擋遊戲);快取一份
	// 供新局/讀檔後重新注入(SetupNewGame 保留注入池,LoadSession 建新 session 需重注)。
	b.applyNebulaStarFlags(b.session) // demo 局也要有星雲旗標,見 cmd/moo2/nebula.go
	b.herodataMercs = loadHerodataMercs(b, res)
	if len(b.herodataMercs) > 0 {
		b.session.SetMercCandidates(b.herodataMercs)
	}
	menu, err := b.menu()
	if err != nil {
		return err
	}

	var shots []galleryShot
	if galleryDir != "" {
		if err := os.MkdirAll(galleryDir, 0o755); err != nil {
			return fmt.Errorf("建立過場截圖目錄 %q: %w", galleryDir, err)
		}
		script, shots = buildGalleryScript()
		// 截圖廊要驗證「事件快報畫面」,但隨機事件每回合只有 30% 機率,靠連按碰運氣既慢又不穩。
		// 固定成一個已知「第一次 EndTurn 就觸發」的種子(見 events.go;seed 1 → 古代遺骸科技),
		// 讓這條驗收路徑每次都走得到事件畫面。只影響截圖廊模式,不影響正常遊戲。
		b.session.EventSeed = 1
	}

	// 預設放大 2 倍(headless 驗證/截圖廊維持 1 倍);視窗可自由拉伸,內容等比縮放置中。
	scale := 2
	if shot != "" || galleryDir != "" {
		scale = 1
	}
	// 正常互動模式先播片頭(原版 Smack 過場,見 cmd/moo2/cutscene.go);headless 驗證
	// 與截圖廊直接進主選單——那些腳本是從主選單第一拍開始數 tick 的,插一段影片會整串偏掉。
	start := origScreen(menu)
	if !b.skipCutscenes {
		if sc := b.intro(); sc != nil {
			start = sc
		}
	}
	app := &interactiveApp{cur: start, script: script, shotPath: shot, frames: frames, scale: scale,
		galleryDir: galleryDir, galleryShots: shots, b: b}
	if galleryDir != "" {
		app.gallerySession = b.session
		// t29 是腳本裡「按 TURN 進最終得分」那一拍;勝負必須在它之前設好,故取 t28。
		app.galleryVictoryTick = galleryVictoryTick
		app.galleryFleetTick = galleryFleetTick
		app.galleryGroundTick = galleryGroundTick
		app.galleryLoadWinTick = galleryLoadWinTick
		app.galleryGameMenuTick = galleryGameMenuTick
		app.galleryBombTick = galleryBombTick
		app.galleryIntroTick = galleryIntroTick
		app.galleryEndingTick = galleryEndingTick
		app.galleryMultiTick = galleryMultiTick
		app.galleryHotseatTick = galleryHotseatTick
		app.galleryDesignTick = galleryDesignTick
		app.galleryBuildPopupTick = galleryBuildPopupTick
		app.galleryCommandPointsTick = galleryCommandPointsTick
		app.galleryMeasureTick = galleryMeasureTick
		app.galleryBuilder = b
	}
	// 只有真正互動(非 headless 截圖/腳本/截圖廊)才啟用音訊:headless 環境常無音效卡,
	// 且截圖驗證不需要聲音。音訊初始化失敗不致命。
	if shot == "" && script == nil {
		app.audio = initAudio(res)
	}
	ebiten.SetWindowSize(moo2ScreenW*scale, moo2ScreenH*scale)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled) // 允許拖曳邊框縮放
	ebiten.SetWindowTitle("Master of Orion II — 繁體中文化 (remake)｜+/- 縮放  F11 全螢幕")
	return ebiten.RunGame(app)
}

// applyNewGameSettings 把 NEW GAME 畫面的設定寫進對局。
//
// **必須在 SetupNewGame 之前呼叫**:星系年齡決定光譜與氣候骰表,星系與行星就是在那裡生成的。
func (b *sceneBuilder) applyNewGameSettings() {
	if b.session == nil {
		return
	}
	if b.newGameAge >= 0 && b.newGameAge < len(shell.GalaxyAges) {
		b.session.GalaxyAge = shell.GalaxyAges[b.newGameAge].Age
		b.session.GalaxyAgeSet = true
	}
	b.session.TechLevel, b.session.TechLevelSet = b.newGameTech, true
}

// newGameOpponents 把「帝國總數」換成 SetupNewGame 要的「AI 對手數」。
// 熱座會再從這些 AI 帝國裡接管席位(見 internal/shell/hotseat.go SetupHotseat),
// 所以帝國總數同時決定了熱座最多能開幾席。
func (b *sceneBuilder) newGameOpponents() int {
	n := b.newGameEmpires - 1
	if n < shell.MinEmpires-1 {
		n = shell.MinEmpires - 1
	}
	if n > shell.MaxEmpires-1 {
		n = shell.MaxEmpires - 1
	}
	return n
}
