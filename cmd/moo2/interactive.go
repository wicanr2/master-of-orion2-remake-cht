package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
	moo2audio "github.com/wicanr2/master-of-orion2-remake-cht/internal/audio"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/engine"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/gamedata"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/i18n"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/lbx"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/netplay"
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

const continuousTurnInterval = 15

// uiScale 是**內部畫布**相對於 640×480 邏輯座標的倍率(第 86 項(hi-res 畫布))。
//
// `rulebook/81`:老遊戲做 CJK 中文化不要縮字,要拉高內部畫布——美術用 nearest 整數倍
// 放大保持銳利,文字用足夠的字級畫在放大後的畫布上。
//
// remake 的畫布本來就是 640×480(原版美術的原生解析),所以「拉高」是指再往上 2×
// 到 1280×960:CJK 從 10–13px 變成 20–26px,而美術一個像素都不糊。
//
// ⚠ **所有畫面的座標仍然是 640×480 邏輯座標,一行都沒改。** 縮放發生在
// `interactiveApp.drawScene`:畫面照舊畫進 640 離屏,文字被錄下來(uifont/record.go),
// 離屏 nearest 放大 2× 之後再用 2× 字級重播文字。
//
// `-uiscale 1` 回到舊路徑(完全不走離屏與錄製),與這一輪之前**逐位元相同**——
// 畫廊的回歸驗證就是這樣做的。
var uiScale = 2.0

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
	// clickAnywhereAction 讓原版以整張畫面作輸入欄的場景接受任意點擊，但不把
	// 640×480 畫成 hover 外框。可見按鈕仍可留在 hits 作現代操作提示。
	clickAnywhereAction string
	// onHotkey 處理這個畫面的鍵盤快捷鍵(nil = 該畫面沒有快捷鍵)。
	// 目前只有星圖接了(見 cmd/moo2/hotkeys.go)。
	onHotkey         func(code string) *origTransition
	hover            string
	offsetX, offsetY int // 背景圖在 640×480 畫布上的置中偏移(小於全螢幕的視窗畫面用)
	// animFrames 是少數原版 overlay 的 delta 動畫(目前議會)；nil 代表靜態背景。
	// animTick 由 sceneBuilder 提供，使用函式避免把場景計數複製成不會更新的值。
	animFrames         []*ebiten.Image
	animTick           func() int
	animationStartTick int
	eraseColor         *color.RGBA // 非 nil 時強制用此色擦底(背景均勻的畫面用,勝過採樣猜測)
	eraseInsetX        int         // 擦底框在基準(左右各3px)之外「再往內縮」的水平量(每邊);0=不變
	eraseInsetY        int         // 擦底框在基準(上下各2px)之外「再往內縮」的垂直量(每邊);0=不變
	plateFace          bool        // true=擦底色改採按鈕面色(浮雕按鈕列用,見 samplePlate faceSample)
	// eraseInset 用途:浮雕按鈕的上下/左右斜邊會被擦底塊蓋掉 → 加內縮只擦中間文字帶,保留浮雕框
	// (仍蓋掉烘進的英文,因英文置中於文字帶內);plateFace 則讓擦底色貼合按鈕面,兩者可併用。
	// labelColorFor 以 enKey 覆寫個別標籤的顏色(空 = 全部用 labelColor)。
	// 目前唯一的用途是把「停用」的選項畫成灰的——原版主選單無存檔時 Continue / Load Game
	// 就是灰階不可按的(2026-07-12 archive.org oracle 對照 issue #2)。
	labelColorFor map[string]color.RGBA
	extraPanels   []extraPanel            // remake 動態控制的底板；先畫，避免譯文和底圖英文疊字
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
	// maxW > 0 時，繪製前依實際欄寬省略，避免動態文字穿出面板。
	maxW float64
}

// extraPanel 是與動態控制文字成對的可見內框。它只用於 remake 明確新增的控制，
// 讓中文字有自己的可見邊界，且能在底圖烘有英文時先乾淨覆蓋。
type extraPanel struct {
	x, y, w, h int
	fill       color.RGBA
	border     color.RGBA
}

// centeredExtraTextInRect 建立必須完整落在可點擊按鈕內的動態文字。extraText 的一般
// 座標是左上角，若直接把按鈕的 x+2／y+14 塞進去，CJK 字身會向下越過 20px 高的
// 按鈕；所有按鈕內的動態文字都應改以中心座標繪製。
func centeredExtraTextInRect(x, y, w, h int, size float64, text string, col color.RGBA) extraText {
	return centeredExtraTextInSafeRect(textSafeRect{x: x, y: y, w: w, h: h, insetX: 3}, size, text, col)
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
		if s.clickAnywhereAction != "" && s.onAction != nil {
			if clickSound != nil {
				clickSound()
			}
			return s.onAction(s.clickAnywhereAction)
		}
	}
	return nil
}

func (s *overlayScreen) draw(dst *ebiten.Image) {
	if s.offsetX != 0 || s.offsetY != 0 {
		dst.Fill(color.RGBA{0, 0, 0, 255}) // 小於全螢幕的視窗:底填黑再置中
	}
	bg := s.bg
	if len(s.animFrames) > 0 && s.animTick != nil {
		frame := (s.animTick() - s.animationStartTick) / 3
		if frame < 0 {
			frame = 0
		}
		if frame >= len(s.animFrames) {
			frame = len(s.animFrames) - 1
		}
		bg = s.animFrames[frame]
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(s.offsetX), float64(s.offsetY))
	drawPanelImage(dst, bg, op)
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
			fillPanel(dst, float32(float64(b.x+ex)+ox), float32(float64(b.y+ey)+oy),
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
			zh = truncateToWidth(s.font, zh, size, float64(b.w-8))
			s.font.DrawCentered(dst, zh, float64(b.x)+float64(b.w)/2+ox, float64(b.y)+float64(b.h)/2+oy, size, lc)
		}
	}
	// 動態控制的底板必須在 hover 與文字之前畫；否則既遮不掉 LBX 內烘的英文，
	// 也會把剛置中的文字再次蓋住。
	for _, p := range s.extraPanels {
		fillPanel(dst, float32(float64(p.x)+ox), float32(float64(p.y)+oy), float32(p.w), float32(p.h), p.fill, false)
		if p.border.A != 0 {
			vector.StrokeRect(dst, float32(float64(p.x)+ox), float32(float64(p.y)+oy), float32(p.w), float32(p.h), 1, p.border, false)
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
		text := e.text
		maxW := e.maxW
		if maxW <= 0 && e.align == 1 {
			// 置中文字沒有自己的面板欄位時，至少不能穿出 640px 畫布。
			maxW = moo2ScreenW - 16
		}
		if maxW > 0 {
			text = truncateToWidth(s.font, text, e.size, maxW)
		}
		if e.align == 1 {
			s.font.DrawCentered(dst, text, e.x+ox, e.y+oy, e.size, e.col)
		} else {
			s.font.Draw(dst, text, e.x+ox, e.y+oy, e.size, e.col)
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
	if f, err := OpenI18NJSON(tsvPath); err == nil {
		defer f.Close()
		if _, err := cat.LoadJSON(f); err != nil {
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

// loadOverlayAnimationFrames 載入少數 overlay 的原版 delta 動畫。與靜態 overlay
// 共用調色盤鏈，但不把多幀畫面全部疊成一張：每個可見 frame 都是從 0 累積到 n，
// 才能保留 delta 幀沒有重寫的像素。呼叫端只對已確認有動畫的畫面使用，避免把
// 不明旗標的所有 LBX 背景一律改變語意。
func loadOverlayAnimationFrames(res *assets.Resolver, lbxName string, assetID int,
	chain paletteChain) ([]*ebiten.Image, error) {
	im, err := decodeAsset(res, lbxName, assetID)
	if err != nil {
		return nil, err
	}
	pal, err := resolvePalette(res, im, chain)
	if err != nil {
		return nil, err
	}
	frames := make([]*ebiten.Image, len(im.Frames))
	for i := range im.Frames {
		frames[i] = ebiten.NewImageFromImage(im.AccumulatedUpToRGBA(pal, i, im.KeyColor()))
	}
	return frames, nil
}

// --- sceneBuilder:依需求建構各原版畫面(共用 resolver/字型/語言)---

type sceneBuilder struct {
	// skipCutscenes:headless 驗證與截圖廊要跳過流程中的過場影片——那些腳本是逐 tick
	// 數出來的,插一段會一直往前播的影片會整串偏掉。截圖廊另外用 tick 注入單獨截過場。
	skipCutscenes bool
	// officerScroll 是軍官清單的捲動位移(第 50 項(軍官畫面座標)加的上下箭頭)。
	// 原版那兩顆鈕的座標一直都在(`_officer_up_button_seg` / `_officer_dn_button_seg`),
	// 只是 remake 先前沒接——所以清單超過四列就看不到後面的人。
	officerTab         int // 0=殖民地領袖、1=艦艇軍官
	officerScroll      int
	officerSelected    int
	officerSelectedSet bool
	officerHireMode    bool
	officerMsg         string // 軍官管理操作的回饋,隨畫面切換保留到下一次操作
	res                *assets.Resolver
	versionAssets      versionAssetDirs // 主選單選版本時切換的兩套 LBX 搜尋路徑
	fnt                *uifont.Font     // 內文用字型(zh 為混合:內文點陣、標題向量)
	fntVec             *uifont.Font     // 純向量 Noto(供主選單等要平滑的畫面;nil 時退回 fnt)
	lang               i18n.Lang
	session            *shell.GameSession       // 活的對局狀態(TURN 推進、畫面顯示即時資料)
	herodataMercs      []shell.Leader           // HERODATA.LBX 解出的真英雄傭兵候選(快取;讀檔後重注入)
	newGameSize        int                      // NEW GAME 選的星系大小索引(shell.GalaxySizes)
	newGameDiff        int                      // NEW GAME 選的難度索引(shell.Difficulties)
	newGameRace        int                      // NEW GAME 選的種族索引(shell.Races)
	newGameSeed        int                      // 每次新遊戲遞增,讓星系種子變化
	newGameAge         int                      // NEW GAME 選的星系年齡索引(shell.GalaxyAges)
	newGameTech        int                      // NEW GAME 選的起始科技等級索引(shell.TechLevels)
	newGameEmpires     int                      // NEW GAME 選的帝國總數(含玩家,shell.MinEmpires..MaxEmpires)
	colChrome          *ebiten.Image            // 殖民地畫面的原版框架(COLPUPS.LBX#5,惰性解碼快取)
	colBldgCache       map[string]*ebiten.Image // 地表建築圖(BLDGn.LBX,惰性解碼快取)
	colVegSizeCache    map[int][2]int           // COLVEGGI 資產的寬高;地表每幀重算,不快取會每幀重解 LBX
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
	flashMsg   string
	flashUntil int
	// continuousTurns 是 End Of Turn Wait 關閉後的可中斷連續回合 UI 狀態；不進存檔。
	continuousTurns    bool
	continuousTurnAt   int
	nebMaskCache       map[int]*nebulaMask  // 星雲遮罩;派遣時沿航線取樣上百次,不快取會重解上百次 LBX
	pendingHotseat     int                  // 多人設定畫面選的真人席位數;0/1 = 單人局
	pendingHotseatAI   []int                // 新局生成後由選帝國畫面指定的 AIPlayers 索引
	savePath           string               // remake 存檔路徑(每回合自動存;主選單 Load/Continue 讀)
	designWeapon       int                  // 艦艇設計選的武器元件索引(shell.WeaponOptions)
	designArmor        int                  // 裝甲元件索引(shell.ArmorOptions)
	designShield       int                  // 護盾元件索引(shell.ShieldOptions)
	designSpecial      int                  // 特殊元件索引(shell.SpecialOptions)
	designMods         []string             // 目前設計勾選的武器改造(gamedata.WeaponModCode 字串;僅 beam 武器生效)
	designArc          gamedata.WeaponArc   // 目前設計武器火線角(原版 WeaponArc)
	designAmmo         int                  // 標準飛彈彈架容量；其他武器由原版表固定
	designMsg          string               // 艦艇設計畫面「空間不足,擋下建造」的提示訊息(切換元件/成功建造時清空)
	designHull         int                  // 目前編輯的六艦體 blueprint 索引(0..5)
	designMount        int                  // 目前編輯的武器槽索引(0..7)
	designSpecialMount int                  // 目前編輯的特殊裝置槽索引(0..7)
	designLoaded       bool                 // UI 是否已從 GameSession 的持久設計庫載入
	lastActionMsg      string               // 星圖畫面「載運陸戰隊/發動地面入侵」的最近一次結果訊息(選新星時清空)
	gameVersion        gamedata.GameVersion // 主選單選的規則版本(1.3/1.5);開局注入 session.RuleProfile
	infoTab            int                  // INFO 畫面目前分頁(0=歷史圖表 1=科技總覽 2=種族統計 3=回合摘要 4=參考),見 infosubscreens.go
	colonyIdx          int                  // 單一殖民地畫面目前管理哪個殖民地(索引 PlayerColonies),見 colonyscreen.go
	colonyListTop      int                  // 單一殖民地畫面「可建項目」清單的捲動起點
	infoHistoryMetric  int                  // 歷史圖表目前指標(shell.HistoryMetric)
	// planetPick 是行星列表畫面選中的行星索引(−1 = 沒選)。原版那個畫面的
	// SEND COLONY SHIP / SEND OUTPOST SHIP 兩顆鈕就是對著選中的那一列作用的。
	planetPick int
	// planetListTop 是行星列表的捲動起點(該畫面一次只顯示 8 列)。
	planetListTop int
	// planetListMsg 是行星列表畫面最近一次動作的結果訊息。
	planetListMsg string
	// netLobby / netConn 是網路對戰的大廳與連線(見 choosenetplyrs.go)。
	// 放在 sceneBuilder 上是因為它們要活過畫面切換——連線不能隨畫面被 GC 掉。
	netLobby *netplay.Lobby
	netConn  net.Conn
	// netAddr / netPlayerName / netJoinOptions 讓客戶端在 socket 斷線後用同一
	// 個身份重連；resume token 只存在記憶體，不寫入遊戲存檔。
	netAddr        string
	netPlayerName  string
	netJoinOptions netplay.JoinOptions
	netLobbyOpts   netplay.LobbyOptions
	// netMe 是本方在名冊裡的玩家編號(主機恆為 0,客戶端由主機指派)。
	netMe int
	// netSess 是對局期間的訊息幫浦(見 internal/netplay/session.go)。
	// 由 netSession() 在第一次需要時才建——大廳階段還在收人,那時候建會漏掉後來的人。
	netSess *netplay.Session
	// networkHost / networkPending 表示這一局正在走 TCP 網路共同開局流程。
	// 它們與 pendingHotseat 分開，避免主機在新遊戲設定中被誤當成本機熱座。
	networkHost    bool
	networkPending bool
	networkRoster  netplay.Roster
	networkTurn    *networkTurnState
	networkError   string
	// netAnnouncer / netBrowser 是區網探索的兩端(見 internal/netplay/discovery.go)。
	// 同樣要活過畫面切換:廣播停了別人就看不到這場對局。
	netAnnouncer *netplay.Announcer
	netBrowser   *netplay.Browser
	// pendingConfirm 是「這一下要先問過玩家」的是/否確認框(原版 `User_Box_(kind=1)`,
	// 見 cmd/moo2/confirmbox.go)。處理器只負責記下來,由呼叫端 takePendingConfirm 換成畫面
	// ——處理器手上沒有「下層畫面」,而確認框要疊在它上面。
	pendingConfirm *pendingConfirm
}

// pendingConfirm 是一個等著被換成畫面的確認框。
type pendingConfirm struct {
	msg   string
	onYes func() *origTransition
}

// takePendingConfirm 取走這一幀待跳的確認框並疊到星圖上;沒有就回 nil。
func (b *sceneBuilder) takePendingConfirm() origScreen {
	p := b.pendingConfirm
	if p == nil {
		return nil
	}
	b.pendingConfirm = nil
	under, _ := b.galaxy() // 取不到就讓確認框自己鋪深色底(見 confirmScreen.draw)
	return b.confirm(under, p.msg, p.onYes, nil)
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
	playBackgroundMusic() // 原版主選單走 Play_Background_Music_:STREAM 1/2/3 每次重擲
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
	hits = append(hits, hitRegion{menuToggleHitX, menuVersionY, menuToggleHitW, menuToggleHitH, "toggleVersion"})
	// 語言切換(CLAUDE.md:「允許在主選單選擇中文/英文」)——擺在版本切換正上方。
	// 先前語言只有啟動旗標 `-lang`,進了遊戲就換不掉,不符合這條需求。
	hits = append(hits, hitRegion{menuToggleHitX, menuLanguageY, menuToggleHitW, menuToggleHitH, "toggleLang"})
	onAction := func(a string) *origTransition {
		switch a {
		case "toggleLang":
			// 切語言 = 換整個顯示層：overlayScreen 是在建構時載入外部文案的，
			// 所以改完 b.lang 要重建畫面才會生效(與版本切換同款做法)。
			// 英文模式下 overlay 機制整段跳過擦底疊字,直接露出原版烘進圖的英文。
			if b.lang == i18n.Traditional {
				b.lang = i18n.English
			} else {
				b.lang = i18n.Traditional
			}
			return b.goTo(b.menu, uiText(b.lang, "mainmenu.transition.main_menu"))
		case "toggleVersion":
			next := gamedata.VersionCommunity15
			if b.gameVersion == gamedata.VersionCommunity15 {
				next = gamedata.VersionClassic13
			}
			if err := b.selectGameVersion(next); err != nil {
				// 沒有該版正版資料時留在原畫面，不讓規則標籤與實際資產版本分離。
				fmt.Fprintln(os.Stderr, "切換遊戲版本:", err)
				return nil
			}
			return b.goTo(b.menu, uiText(b.lang, "mainmenu.transition.main_menu")) // 重繪以更新版本顯示
		case "Quit Game":
			return &origTransition{quit: true}
		case "New Game":
			// 新遊戲:先進原版 NEW GAME 設定畫面(難度/星系/玩家…),ACCEPT 後進星系主畫面。
			b.pendingHotseat = 0 // 單人局(從多人設定畫面進來的才會帶席位數)
			b.pendingHotseatAI = nil
			return b.goTo(b.newGameSetup, uiText(b.lang, "mainmenu.transition.new_game"))
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
			return b.goTo(b.galaxy, uiText(b.lang, "mainmenu.transition.galaxy"))
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
			return b.goTo(b.hiScore, uiText(b.lang, "mainmenu.transition.hiscore"))
		}
		return nil
	}
	// 主選單用純向量 Noto(平滑),不走內文點陣(使用者要求主選單維持向量觀感);
	// 無 fntVec(如 zh 未帶 -font)時退回 b.fnt。
	menuFont := b.fntVec
	if menuFont == nil {
		menuFont = b.fnt
	}
	s, err := loadOverlayScreen(b.res, "mainmenu.lbx", 21, b.lang, menuFont, "menu.json",
		menuOverlays, color.RGBA{104, 224, 96, 255}, 15, hits, onAction, nil)
	if err != nil {
		return nil, err
	}
	if len(dimmed) > 0 {
		s.labelColorFor = dimmed // 無存檔 → Continue / Load Game 畫成灰的
	}
	// 左下角版本 / 語言切換標籤(點擊上方熱區循環)。
	// 語言標籤本身要跟著當前語言走,否則英文模式下留一行中文很怪。
	langLabel := uiText(b.lang, "mainmenu.toggle.language")
	verLabel := mainMenuText(b.lang, "mainmenu.toggle.rules", versionShort(b.gameVersion))
	s.extras = append(s.extras,
		mainMenuToggleTextRect(menuLanguageY).leftExtras(menuFont, langLabel, 11, color.RGBA{150, 210, 150, 255})...,
	)
	s.extras = append(s.extras,
		mainMenuToggleTextRect(menuVersionY).leftExtras(menuFont, verLabel, 11, color.RGBA{150, 210, 150, 255})...,
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

// 星圖選星資訊面板的文字安全框。面板與熱區都固定在 640×480 邏輯座標；文字框
// 另外保留內縮，避免點陣字墨碰到右上 CLOSE 或下方操作鈕。
const (
	starPanelX, starPanelY = 28, 326
	starPanelW, starPanelH = 210, 140
	starPanelButtonX       = 38
	starPanelButtonW       = 190
)

func starPanelNameTextRect() textSafeRect {
	return textSafeRect{x: 38, y: 337, w: 120, h: 16, insetX: 0, insetY: 0}
}

func starPanelSpecialTextRect() textSafeRect {
	return textSafeRect{x: 158, y: 337, w: 56, h: 16, insetX: 0, insetY: 0}
}

func starPanelCloseTextRect() textSafeRect {
	return textSafeRect{x: 216, y: 337, w: 18, h: 17, insetX: 0, insetY: 1}
}

func starPanelEnvironmentTextRect(y int) textSafeRect {
	return textSafeRect{x: 38, y: y, w: 190, h: 16, insetX: 0, insetY: 0}
}

func starPanelMarineTextRect() textSafeRect {
	return textSafeRect{x: 38, y: 385, w: 190, h: 16, insetX: 0, insetY: 0}
}

func starPanelButtonTextRect(y int) textSafeRect {
	return textSafeRect{x: starPanelButtonX, y: y, w: starPanelButtonW, h: 20, insetX: 8, insetY: 2}
}

func drawStarPanelText(dst *ebiten.Image, fnt *uifont.Font, r textSafeRect, text string, size float64, col color.Color) {
	r.drawLeft(dst, fnt, text, size, col)
}

// galaxy 建原版星系主畫面(遊戲主樞紐,BUFFER0.LBX 資產 0)。底部工具列導覽到各畫面
// (座標取自 openorion2 galaxy.cpp GalaxyView::initWidgets)。
func (b *sceneBuilder) galaxy() (*overlayScreen, error) {
	playBackgroundMusic() // 星圖同樣走 Play_Background_Music_(第 73 項(音樂場景表))
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
					if sess.RaceTelepathic() {
						hits = append(hits, hitRegion{38, 446, 190, 20, "mindcontrol"})
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
					if c := b.takePendingConfirm(); c != nil {
						return &origTransition{next: c} // 原版對這一下要先問一句(見 confirmbox.go)
					}
					return b.goTo(b.galaxy, "星系主畫面")
				}
				if idx == b.session.SelectedStar {
					b.session.SelectedStar = -1 // 再點同一顆星 → 取消選取(關閉資訊面板,issue #6)
				} else {
					b.session.SelectedStar = idx
					if colony := autoSelectedColonyIndex(b.session, idx); colony >= 0 {
						b.colonyIdx = colony
						b.colonyListTop = 0
						return b.goTo(b.colonyScreen, uiText(b.lang, "colony.transition.colony"))
					}
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
			b.flash(uiText(b.lang, "relocation.prompt.star_panel_target"))
			return b.goTo(b.galaxy, uiText(b.lang, "relocation.transition.galaxy"))
		}
		if a == "dispatch" && b.session != nil {
			// 派遣艦隊至選中星(航行由 EndTurn 推進)。曲速前開局沒有 FTL、出不了本星系,
			// SendFleet 會直接回 false——要說清楚原因,不然玩家只會看到「點了沒反應」。
			if !b.session.SendFleet(b.session.SelectedStar) {
				if !b.session.FleetHasFTL() {
					b.lastActionMsg = uiText(b.lang, "galaxy.star_panel.error.no_ftl")
				}
			}
			return b.goTo(b.galaxy, "星系主畫面")
		}
		if a == "loadmarines" && b.session != nil {
			n := b.session.LoadMarines(0) // 母星是唯一已知殖民地索引對映(見上方熱區註解)
			if n > 0 {
				b.lastActionMsg = fmt.Sprintf(uiText(b.lang, "galaxy.star_panel.result.marines_loaded"), n)
			} else {
				b.lastActionMsg = uiText(b.lang, "galaxy.star_panel.error.no_marines")
			}
			return b.goTo(b.galaxy, "星系主畫面")
		}
		if a == "bombard" && b.session != nil {
			res := b.session.BombardColony(b.session.SelectedStar)
			if !res.Ok { // 前置條件不足:沒開炸,留在星系主畫面說明原因
				b.lastActionMsg = bombardmentRefusalText(b.lang, res)
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
				b.lastActionMsg = invasionRefusalText(b.lang, res)
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
		if a == "mindcontrol" && b.session != nil {
			res := b.session.MindControlColony(b.session.SelectedStar)
			if !res.Ok {
				b.lastActionMsg = mindControlRefusalText(b.lang, res)
				return b.goTo(b.galaxy, "星系主畫面")
			}
			b.lastActionMsg = uiText(b.lang, "galaxy.star_panel.result.mind_control")
			return b.goTo(b.galaxy, "星系主畫面")
		}
		if a == "colonize" && b.session != nil {
			res := b.session.ColonizeStar(b.session.SelectedStar)
			if !res.Ok {
				b.lastActionMsg = colonizationRefusalText(b.lang, res)
			} else {
				b.lastActionMsg = fmt.Sprintf(uiText(b.lang, "galaxy.star_panel.result.colonized"), res.StartPopulation, res.PopMax)
			}
			return b.goTo(b.galaxy, "星系主畫面")
		}
		if a == "attackmonster" && b.session != nil {
			star := b.session.SelectedStar
			p, e, reason := b.session.StartMonsterCombat(star)
			if reason != "" {
				b.lastActionMsg = monsterCombatRefusalText(b.lang, reason)
				return b.goTo(b.galaxy, "星系主畫面")
			}
			playCombatMusic()
			return &origTransition{next: newTacticalScreenForShips(b, p, e, star)}
		}
		if a == "outpost" && b.session != nil {
			res := b.session.BuildOutpost(b.session.SelectedStar)
			if !res.Ok {
				b.lastActionMsg = outpostRefusalText(b.lang, res)
			} else {
				b.lastActionMsg = uiText(b.lang, "galaxy.star_panel.result.outpost")
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
			// 原版 Auto Select Ships 只在新進入艦隊操作時初始化；nil 與玩家手動
			// 全不選的空 map 必須分開，否則畫面重建會把取消選取吃掉。
			b.shipPick = nil
			return b.goTo(b.fleet, uiText(b.lang, "fleet.transition.fleet"))
		case "leaders":
			b.officerTab = 0
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
	// ⚠ 底部工具列七格的 y/h 是**擦底板**的範圍,不是按鈕範圍(按鈕熱區在上面的 hits)。
	// 原本 y=443/h=14 → 擦到 445..455,而烘進美術的 `COLONIES`/`PLANETS`/… 上緣在 440,
	// 所以**每一格的英文都露出上面 5 列**。1× 就有這個瑕疵,只是 2× 之後一眼就看得到。
	// 改成 438/19 → 擦 440..455,上不吃到按鈕的浮雕上緣(在 435)。
	overlays := []labelRect{
		{13, 438, 71, 19, "Colonies", 12},
		{88, 438, 71, 19, "Planets", 12},
		{254, 1, 88, 19, "Game", 13}, // 頂部標題列烘進的 GAME
		{163, 438, 71, 19, "Fleets", 12},
		{235, 438, 74, 19, "Zoom", 12},
		{308, 438, 74, 19, "Leaders", 12},
		{383, 438, 74, 19, "Races", 12},
		{458, 438, 74, 19, "Info", 12},
		{544, 448, 90, 15, "Turn", 12},
	}
	s, err := loadOverlayScreen(b.res, "buffer0.lbx", 0, b.lang, b.fnt, "menu.json",
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
			vis := sess.StarChartVisible()
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
				fnt.DrawCentered(dst, fmt.Sprintf(uiText(b.lang, "galaxy.status.tax"), sess.Player.TaxRate),
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
						fillPanel(dst, float32(fx-4), float32(fy-4), 8, 8, color.RGBA{80, 240, 240, 255}, false)
					}
				}
				// 選中星:顯示該星系行星資訊 + 派遣艦隊/載運陸戰隊/軌道轟炸/發動入侵按鈕(左下角面板)。
				// starPanelTextLayoutStart
				if sess.SelectedStar >= 0 && sess.SelectedStar < len(sess.Planets) {
					p, _ := sess.PlanetDataAt(sess.SelectedStar)
					// 面板高度 140 是 remake 轉接：敵殖民地可能同時顯示轟炸(402)、入侵(424)
					// 與心靈控制(446)，底緣需到 466，否則最下列會露出背景框之外。
					fillPanel(dst, starPanelX, starPanelY, starPanelW, starPanelH, color.RGBA{10, 14, 30, 235}, false)
					vector.StrokeRect(dst, starPanelX, starPanelY, starPanelW, starPanelH, 1, color.RGBA{90, 130, 200, 255}, false)
					drawStarPanelText(dst, fnt, starPanelNameTextRect(), p.Name, 14, color.RGBA{240, 220, 120, 255})
					// 右上 CLOSE 鈕(✕),對齊上方 "closestar" 熱區與原版彈窗 CLOSE(issue #6)。
					drawStarPanelText(dst, fnt, starPanelCloseTextRect(), uiText(b.lang, "galaxy.star_panel.button.close"), 14, color.RGBA{235, 150, 140, 255})
					// 行星特殊物產(金礦/寶石礦/原住民/遠古文物…)接在星名右邊,另用一色標出來——
					// 這是「這顆星值不值得搶」的關鍵資訊。標題列左右分欄，
					// 不讓長星名把它推過 CLOSE，也不讓兩者互相蓋字。
					specialLabel := ""
					specialColor := color.RGBA{250, 200, 100, 255}
					if mon := sess.MonsterNameAtStar(sess.SelectedStar); mon != "" {
						specialLabel = fmt.Sprintf(uiText(b.lang, "galaxy.star_panel.monster"), mon)
						specialColor = color.RGBA{240, 130, 150, 255}
					} else if sp := planetSpecialLabel(b.lang, p.SpecialID); sp != "" {
						specialLabel = fmt.Sprintf(uiText(b.lang, "galaxy.star_panel.special"), sp)
					}
					drawStarPanelText(dst, fnt, starPanelSpecialTextRect(), specialLabel, 10, specialColor)
					climate, gravity, minerals, size := planetEnvironmentLabels(b.lang, p)
					drawStarPanelText(dst, fnt, starPanelEnvironmentTextRect(353), fmt.Sprintf(uiText(b.lang, "galaxy.star_panel.environment.climate_size"), climate, size), 11, color.RGBA{210, 216, 230, 255})
					drawStarPanelText(dst, fnt, starPanelEnvironmentTextRect(369), fmt.Sprintf(uiText(b.lang, "galaxy.star_panel.environment.gravity_minerals"), gravity, minerals), 11, color.RGBA{210, 216, 230, 255})
					// 同系其他天體(氣態巨星/小行星帶)的完整摘要放在行星列表畫面——這個面板
					// 只有 337~401 這四列的空間,402 起是操作鈕,再塞一列會壓到按鈕。
					// 陸戰隊狀態行:艦隊目前載運數,選中母星時另顯示殖民地駐軍池數(唯一已知對映)。
					marineLine := fmt.Sprintf(uiText(b.lang, "galaxy.star_panel.marines.fleet"), sess.Fleet().Marines)
					if sess.SelectedStar == 0 && len(sess.PlayerColonyMarines) > 0 {
						marineLine = fmt.Sprintf(uiText(b.lang, "galaxy.star_panel.marines.garrison"),
							sess.Fleet().Marines, sess.PlayerColonyMarines[0])
					}
					// 艦員等級接在同一行:這是玩家唯一看得到艦員經驗的地方(remake 沒有逐艦
					// 資訊面板),而那個等級直接影響命中、防禦與飛彈閃避(見第 60 項(打得準也閃得掉))。
					// 取艦隊裡**最低**的那一艘——戰力由最弱的那條線決定。
					if lv, toNext, ok := sess.FleetCrewSummary(); ok {
						marineLine += fmt.Sprintf(uiText(b.lang, "galaxy.star_panel.crew"), shell.ShipCrewLevelName(lv))
						if toNext > 0 {
							marineLine += fmt.Sprintf(uiText(b.lang, "galaxy.star_panel.crew_next"), toNext)
						}
					}
					drawStarPanelText(dst, fnt, starPanelMarineTextRect(), marineLine, 11, color.RGBA{200, 220, 170, 255})
					// 操作鈕/狀態(與 galaxy() 建 hits 時的判斷邏輯一致)。
					switch {
					case b.lastActionMsg != "":
						fillPanel(dst, 38, 402, 190, 20, color.RGBA{30, 55, 35, 235}, false)
						vector.StrokeRect(dst, 38, 402, 190, 20, 1, color.RGBA{110, 200, 140, 255}, false)
						drawStarPanelText(dst, fnt, starPanelButtonTextRect(402), b.lastActionMsg, 10, color.RGBA{225, 240, 225, 255})
					case sess.Fleet().ETA > 0:
						drawStarPanelText(dst, fnt, starPanelButtonTextRect(402), fmt.Sprintf(uiText(b.lang, "galaxy.star_panel.status.transit"), sess.Fleet().ETA), 11, color.RGBA{120, 200, 240, 255})
					case sess.SelectedStar == sess.Fleet().AtStar && sess.SelectedStar == 0:
						fillPanel(dst, 38, 402, 190, 20, color.RGBA{40, 70, 120, 255}, false)
						vector.StrokeRect(dst, 38, 402, 190, 20, 1, color.RGBA{110, 160, 230, 255}, false)
						drawStarPanelText(dst, fnt, starPanelButtonTextRect(402), uiText(b.lang, "galaxy.star_panel.button.load_marines"), 12, color.RGBA{230, 235, 245, 255})
					case sess.SelectedStar == sess.Fleet().AtStar && sess.Stars[sess.SelectedStar].Owner == 2:
						// 軌道轟炸恆可用(艦隊武器開火,不需陸戰隊),畫在 402 這列。
						fillPanel(dst, 38, 402, 190, 20, color.RGBA{90, 60, 130, 255}, false)
						vector.StrokeRect(dst, 38, 402, 190, 20, 1, color.RGBA{170, 140, 230, 255}, false)
						drawStarPanelText(dst, fnt, starPanelButtonTextRect(402), uiText(b.lang, "galaxy.star_panel.button.bombard"), 12, color.RGBA{240, 235, 250, 255})
						// 發動地面入侵額外要求已載運陸戰隊,畫在下一列(424),與轟炸鈕並存。
						if sess.Fleet().Marines > 0 {
							fillPanel(dst, 38, 424, 190, 20, color.RGBA{120, 50, 40, 255}, false)
							vector.StrokeRect(dst, 38, 424, 190, 20, 1, color.RGBA{230, 130, 110, 255}, false)
							drawStarPanelText(dst, fnt, starPanelButtonTextRect(424), uiText(b.lang, "galaxy.star_panel.button.invade"), 12, color.RGBA{245, 235, 230, 255})
						}
						if sess.RaceTelepathic() {
							fillPanel(dst, 38, 446, 190, 20, color.RGBA{55, 95, 125, 255}, false)
							vector.StrokeRect(dst, 38, 446, 190, 20, 1, color.RGBA{120, 210, 235, 255}, false)
							drawStarPanelText(dst, fnt, starPanelButtonTextRect(446), uiText(b.lang, "galaxy.star_panel.button.mind_control"), 12, color.RGBA{230, 250, 255, 255})
						}
					case sess.SelectedStar == sess.Fleet().AtStar && sess.StarGuardedByMonster(sess.SelectedStar):
						fillPanel(dst, 38, 402, 190, 20, color.RGBA{110, 45, 60, 255}, false)
						vector.StrokeRect(dst, 38, 402, 190, 20, 1, color.RGBA{230, 120, 140, 255}, false)
						drawStarPanelText(dst, fnt, starPanelButtonTextRect(402), fmt.Sprintf(uiText(b.lang, "galaxy.star_panel.button.attack"), sess.MonsterNameAtStar(sess.SelectedStar)), 12, color.RGBA{250, 230, 235, 255})
					case sess.SelectedStar == sess.Fleet().AtStar && len(starPanelColonyRows(sess)) > 0:
						// 與 galaxy() 建 hits 的判斷共用同一支 starPanelColonyRows,
						// 免得「畫得出來卻點不到」或反過來。
						for _, r := range starPanelColonyRows(sess) {
							row := float32(r.y)
							face, edge, ink := color.RGBA{40, 110, 60, 255}, color.RGBA{130, 220, 150, 255}, color.RGBA{235, 245, 235, 255}
							labelKey := "galaxy.star_panel.button.colonize"
							if r.action == "outpost" {
								face, edge, ink = color.RGBA{45, 80, 110, 255}, color.RGBA{140, 190, 230, 255}, color.RGBA{230, 240, 250, 255}
								labelKey = "galaxy.star_panel.button.outpost"
							}
							label := uiText(b.lang, labelKey)
							// 目標天體寫進鈕裡——同星系有多顆天體時,玩家要看得出這一下會落在哪顆。
							if r.planet >= 0 && r.planet < len(sess.Planets) {
								label = fmt.Sprintf(uiText(b.lang, "galaxy.star_panel.button.target"), label, sess.Planets[r.planet].Name)
							}
							fillPanel(dst, 38, row, 190, 20, face, false)
							vector.StrokeRect(dst, 38, row, 190, 20, 1, edge, false)
							drawStarPanelText(dst, fnt, starPanelButtonTextRect(r.y), label, 12, ink)
						}
					case sess.SelectedStar == sess.Fleet().AtStar:
						drawStarPanelText(dst, fnt, starPanelButtonTextRect(402), uiText(b.lang, "galaxy.star_panel.status.arrived"), 11, color.RGBA{140, 200, 140, 255})
					default:
						fillPanel(dst, 38, 402, 190, 20, color.RGBA{40, 70, 120, 255}, false)
						vector.StrokeRect(dst, 38, 402, 190, 20, 1, color.RGBA{110, 160, 230, 255}, false)
						drawStarPanelText(dst, fnt, starPanelButtonTextRect(402), uiText(b.lang, "galaxy.star_panel.button.dispatch"), 12, color.RGBA{230, 235, 245, 255})
					}
					// 集結點鈕(第二列):選中的是自己的殖民地才有。標題直接寫出目前設到哪,
					// 不然玩家看不出「有沒有設」——那正是這個功能最容易被忽略的地方。
					if ci := colonyIndexAtStar(sess, sess.SelectedStar); ci >= 0 {
						fillPanel(dst, 38, 424, 190, 20, color.RGBA{45, 95, 85, 255}, false)
						vector.StrokeRect(dst, 38, 424, 190, 20, 1, color.RGBA{120, 200, 180, 255}, false)
						label := uiText(b.lang, "relocation.button.set")
						if to := sess.ColonyRelocation(ci); to >= 0 && to < len(sess.Stars) {
							label = fmt.Sprintf(uiText(b.lang, "relocation.button.current"), sess.Stars[to].Name)
						}
						drawStarPanelText(dst, fnt, starPanelButtonTextRect(424), label, 12, color.RGBA{225, 245, 240, 255})
					}
				}
				// starPanelTextLayoutEnd
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
	if b != nil && b.session != nil {
		drawEnemyFleetMoves(dst, stars, b.session.VisibleEnemyFleetMoves(), b.animTick)
	}
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

// drawEnemyFleetMoves 把規則層已通過設定與霧區 gate 的敵方航線畫在星球下方。
// 線色與 marker timing 是 remake 視覺近似；原版 byte_199BDF 的精確動畫 consumer 尚未知。
func drawEnemyFleetMoves(dst *ebiten.Image, stars []shell.Star, moves []shell.EnemyFleetMove, tick int) {
	if dst == nil {
		return
	}
	lineColor := color.RGBA{235, 72, 64, 180}
	markerColor := color.RGBA{255, 196, 96, 240}
	for _, move := range moves {
		x1, y1, x2, y2, mx, my, ok := enemyMoveGeometry(stars, move, tick)
		if !ok {
			continue
		}
		vector.StrokeLine(dst, x1, y1, x2, y2, 1, lineColor, true)
		vector.DrawFilledCircle(dst, mx, my, 2, markerColor, true)
	}
}

func enemyMoveGeometry(stars []shell.Star, move shell.EnemyFleetMove, tick int) (x1, y1, x2, y2, mx, my float32, ok bool) {
	if move.FromStar < 0 || move.FromStar >= len(stars) || move.ToStar < 0 || move.ToStar >= len(stars) ||
		move.FromStar == move.ToStar {
		return 0, 0, 0, 0, 0, 0, false
	}
	from, to := stars[move.FromStar], stars[move.ToStar]
	x1 = float32(starVX0) + float32(from.X)*(starVX1-starVX0)
	y1 = float32(starVY0) + float32(from.Y)*(starVY1-starVY0)
	x2 = float32(starVX0) + float32(to.X)*(starVX1-starVX0)
	y2 = float32(starVY0) + float32(to.Y)*(starVY1-starVY0)
	phase := tick + move.AIIndex*17
	if phase < 0 {
		phase = -phase
	}
	fraction := float32(phase%90) / 89
	mx = x1 + (x2-x1)*fraction
	my = y1 + (y2-y1)*fraction
	return x1, y1, x2, y2, mx, my, true
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
// (母星)。氣候/重力/礦產的**中文名**沒有官方在地化來源可援引——既有 i18n JSON 與先前的
// 中文化專案(~/master-of-orion)都沒有 MOO2 這幾組列舉的定案譯名,故用簡明直譯頂著顯示,
// 不是官方在地化文本。**英文名**則直接取原版手冊用語,不是回譯。
// 純展示層查表,不影響 engine/gamedata 任何邏輯或數值。
// shipClassLabel 把 shell 的艦體規則鍵換成外部 ui.json 的玩家顯示名稱。
func shipClassLabel(lang i18n.Lang, zhKey string) string {
	for i, k := range shipClassZH {
		if k == zhKey && i < len(shipClassTextKeys) {
			return uiText(lang, shipClassTextKeys[i])
		}
	}
	if key, ok := supportClassTextKeys[zhKey]; ok {
		return uiText(lang, key)
	}
	return uiText(lang, "ship.class.unknown")
}

// supportClassTextKeys 只把支援艦規則鍵路由到語意鍵；顯示文字不留在 Go。
var supportClassTextKeys = map[string]string{
	"殖民船": "ship.class.colony_ship",
	"偵察艦": "ship.class.scout",
	"前哨船": "ship.class.outpost_ship",
	"運輸艦": "ship.class.freighter",
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

// wrapToWidth 把 s 折成每行寬度不超過 maxW 的多行。
//
// 中文沒有空白可斷,所以逐 rune 累積;遇到 '\n' 強制換行。西文詞會被硬切——
// remake 的 UI 文字以中文為主,為了一句英文提示引進斷詞規則不划算,
// 真的需要時在原文裡自己放 '\n'。
func wrapToWidth(fnt *uifont.Font, s string, size, maxW float64) []string {
	if fnt == nil || s == "" {
		return []string{s}
	}
	var out []string
	for _, para := range strings.Split(s, "\n") {
		line := make([]rune, 0, 32)
		for _, r := range para {
			cand := append(append([]rune(nil), line...), r)
			if w, _ := fnt.Measure(string(cand), size); w > maxW && len(line) > 0 {
				out = append(out, string(line))
				line = []rune{r}
				continue
			}
			line = cand
		}
		out = append(out, string(line))
	}
	return out
}

func (b *sceneBuilder) colonySummary() (*overlayScreen, error) {
	// 點各殖民地的職務欄 → 重分配 1 名人口(農夫欄→多農夫、工人欄→多工人、科學家欄→多科學家);
	// RETURN → 星系主畫面。列中心 y 與欄 x 對齊資料。
	hits := []hitRegion{{582, 452, 52, 20, "return"}}
	if b.session != nil {
		for i := range b.session.PlayerColonies {
			if i >= len(colonySummaryRowY) {
				break
			}
			top := colonySummaryRowY[i] - 15
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
			return b.goTo(b.galaxy, uiText(b.lang, "colony.summary.transition.galaxy"))
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
				return b.goTo(b.colonyScreen, uiText(b.lang, "colony.summary.transition.colony"))
			}
			return b.goTo(b.colonySummary, uiText(b.lang, "colony.summary.transition.summary")) // 重繪顯示新分配
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
		// ⚠ 底部排序列這一整排原本**整體偏右**:每一格的擦底板都蓋在英文的右半邊,
		// 於是每個中文標籤左邊都掛著一小截沒擦掉的英文(`P`、`IOC`、`R`、`e`…),
		// 而「返回」更誇張——板在 585..631,`RETURN` 其實在 552..594,等於完全沒蓋到。
		//
		// 下面的值是**在英文模式跑同一張畫廊圖**(擦底整段不畫)之後掃亮字得到的:
		// SORT 23..74、Name 99..130、Population 148..208、Food 228..253、Industry 269..316、
		// Science 334..382、Producing 402..467、BC 489..504、RETURN 552..594(暗紅字,另量)。
		// 每一格的板都涵蓋對應區段,且板中心對齊英文中心(中文才不會左右歪)。
		{20, 450, 58, 22, "SORT", 0},
		// 七格**刻意互相重疊**(擦底板 91..514 連續無縫):留縫的話英文抗鋸齒的邊緣像素
		// 會從縫裡露出來,變成每個標籤旁邊的小雜點。板色都採自同一條深藍長條,重疊無害。
		{88, 450, 55, 22, "Name", 0},
		{136, 450, 84, 22, "Population", 0},
		{214, 450, 54, 22, "Food", 0},
		{258, 450, 70, 22, "Industry", 0},
		{322, 450, 72, 22, "Science", 0},
		{388, 450, 94, 22, "Producing", 0},
		{475, 450, 42, 22, "BC", 0},
		{538, 447, 76, 22, "RETURN", 0},
	}
	s, err := loadOverlayScreen(b.res, "colsum.lbx", 0, b.lang, b.fnt, "colony.json",
		overlays, color.RGBA{210, 216, 230, 255}, 13, hits, onAction, nil)
	if err != nil {
		return nil, err
	}
	// 底列/欄名皆浮雕鈕:擦底色改採按鈕面色,避免暗塊蓋掉浮雕框(見 samplePlate faceSample)。
	s.plateFace = true
	// 即時殖民地資料填進表格列(欄位中心 x 對齊標題;列中心 y 經 PIL 量測,每列約 31px)。
	if b.session != nil {
		body := color.RGBA{214, 220, 235, 255}
		for i, c := range b.session.PlayerColonies {
			if i >= len(colonySummaryRowY) {
				break
			}
			s.extras = append(s.extras, colonySummaryColumnRect(i, 0).centeredExtras(b.fnt,
				colonySummaryText(b.lang, "colony.summary.row.name", i+1), 11, body)...)
			for column, value := range []int{c.Farmers, c.Workers, c.Scientists} {
				s.extras = append(s.extras, colonySummaryColumnRect(i, column+1).centeredExtras(b.fnt,
					fmt.Sprintf("%d", value), 11, body)...)
			}
			// 建造欄:項目名 + 進度(空則顯示「—」提示可點)。
			bt := uiText(b.lang, "colony.summary.build.empty")
			if i < len(b.session.Builds) && b.session.Builds[i].Name != "" {
				bd := b.session.Builds[i]
				bt = colonySummaryText(b.lang, "colony.summary.build.progress",
					buildItemLabel(b.lang, bd.Name), bd.Progress, bd.Cost)
			}
			if n := b.session.BuildQueueBacklogLen(i); n > 0 {
				bt += colonySummaryText(b.lang, "colony.summary.build.backlog", n)
			}
			builtLabel := ""
			// 已建建築(顯示效果來源):與目前建造合併為一列，避免 30px 列內兩行 CJK 字墨重疊。
			if i < len(b.session.ColonyBuildings) && len(b.session.ColonyBuildings[i]) > 0 {
				names := make([]string, 0, len(b.session.ColonyBuildings[i]))
				for n := range b.session.ColonyBuildings[i] {
					names = append(names, n)
				}
				sort.Strings(names)
				// 依「建造」欄寬截斷,避免長建築清單溢出 cell 框、撞出畫面右緣
				// (點陣字最小 12px,小字撐大更易超框;BUILDING 欄 x512 寬118,留邊取 110)。
				displayNames := names
				if b.lang != i18n.Traditional {
					displayNames = make([]string, len(names))
					for j, name := range names {
						displayNames[j] = colonyBuildingLabel(b.lang, name)
					}
				}
				builtLabel = colonySummaryText(b.lang, "colony.summary.build.built",
					strings.Join(displayNames, uiText(b.lang, "list.separator")))
			}
			if builtLabel != "" {
				bt = colonySummaryText(b.lang, "colony.summary.build.with_built", bt, builtLabel)
			}
			s.extras = append(s.extras, colonySummaryBuildRect(i).centeredExtras(b.fnt, bt, 8, body)...)
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
			colonySummaryText(b.lang, "colony.summary.empire.treasury", b.session.Player.BC),
			colonySummaryText(b.lang, "colony.summary.empire.net", out.NetBC),
			colonySummaryText(b.lang, "colony.summary.empire.population", pop),
			colonySummaryText(b.lang, "colony.summary.empire.food", out.TotalFood),
			colonySummaryText(b.lang, "colony.summary.empire.research", out.TotalResearch),
		}
		for i, l := range lines {
			s.extras = append(s.extras, colonySummaryEmpireRect(i).leftExtras(b.fnt, l, 8, es)...)
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
			for i, ry := range colonySummaryRowY {
				if i >= len(b.session.PlayerColonies) {
					break
				}
				top, bottom := ry-15, ry+16
				if s.my >= top && s.my < bottom && s.mx >= 10 && s.mx < 628 {
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
				colonySummaryText(b.lang, "colony.summary.planet.climate", climateName(b.lang, c.Climate)),
				colonySummaryText(b.lang, "colony.summary.planet.gravity", gravityName(b.lang, c.PlanetGravity)),
				colonySummaryText(b.lang, "colony.summary.planet.minerals", mineralsName(b.lang, c.MineralRichness)),
				colonySummaryText(b.lang, "colony.summary.planet.size", planetSizeName(b.lang, c.PlanetSize)),
				colonySummaryText(b.lang, "colony.summary.planet.max_population", c.PopMax),
			}
			for i, l := range piLines {
				r := colonySummaryPlanetInfoRect(i)
				r.x, r.y = r.x+int(ox), r.y+int(oy)
				r.drawLeft(dst, s.font, l, 7, label)
			}

			// Production Info:較寬(269px)。優先用 LastPlayerOutput.Colonies[idx](當回合已
			// 結算的實際產出);取不到(如新遊戲尚未跑過第一回合)時退回用 PlayerColonies
			// 欄位 × 人數的簡化估算,並標「約」避免誤當精確結算值。
			var prodLines []string
			if idx < len(b.session.LastPlayerOutput.Colonies) {
				co := b.session.LastPlayerOutput.Colonies[idx]
				prodLines = []string{
					colonySummaryText(b.lang, "colony.summary.production.food", co.Food, co.FoodConsumed, co.FoodSurplus),
					colonySummaryText(b.lang, "colony.summary.production.industry", co.GrossIndustry, co.NetIndustry),
					colonySummaryText(b.lang, "colony.summary.production.research", co.Research),
					colonySummaryText(b.lang, "colony.summary.production.pollution", co.PollutionCleanupCost),
				}
				if co.Starving {
					prodLines = append(prodLines, uiText(b.lang, "colony.summary.production.starving"))
				}
			} else {
				prodLines = []string{
					colonySummaryText(b.lang, "colony.summary.production.food_estimate", c.Farmers*c.FoodPerFarmer),
					colonySummaryText(b.lang, "colony.summary.production.industry_estimate", c.Workers*c.IndustryPerWorker),
					colonySummaryText(b.lang, "colony.summary.production.research_estimate", c.Scientists*c.ResearchPerScientist),
				}
			}
			for i, l := range prodLines {
				r := colonySummaryProductionRect(i)
				r.x, r.y = r.x+int(ox), r.y+int(oy)
				r.drawLeft(dst, s.font, l, 9, label)
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
	// 每個已接觸種族一顆「派間諜」鈕,座標是**執行檔立即數**(見 racesSpyAnchors)。
	hits = append(hits, racesSpyHitRegions(b.aiCount())...)
	// Sabotage 左右順序尚未由原版反組譯證實；這個熱區是 remake 明確標籤的
	// STEAL/SABOTAGE/HIDE 循環，不冒充原版三顆鈕的未解語意。
	hits = append(hits, racesSpyMissionHitRegions(b.aiCount())...)
	// 第三個原生槽是 remake 的明確「隱匿」快捷鍵；不再留下英文 HIDE 或塞進下一列。
	hits = append(hits, racesSpyHideHitRegions(b.aiCount())...)
	// 每個已顯示 AI 列都可直接進入該對手的外交對談；這是 remake 的明確逐對手
	// 入口，原版單列 REPORT 的選取語意仍保持未知。
	hits = append(hits, racesDiplomacyHitRegions(b.aiCount())...)
	// 防守 Agent 與進攻間諜分開管理；這兩個控制項讓已接上的 Agent slot
	// 真正能由玩家調整，而不是只存在於資料層。
	hits = append(hits, racesAgentHitRegions()...)
	onAction := func(a string) *origTransition {
		switch a {
		case "audience":
			return b.goTo(b.council, uiText(b.lang, "races.transition.council"))
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
		case "trainagent":
			if b.session != nil {
				b.session.TrainDefensiveAgent()
			}
			return b.goTo(b.races, uiText(b.lang, "races.transition.screen"))
		case "dismissagent":
			if b.session != nil {
				b.session.DismissDefensiveAgent()
			}
			return b.goTo(b.races, uiText(b.lang, "races.transition.screen"))
		}
		if idx, ok := racesSpyActionIndex(a); ok {
			if b.session != nil {
				b.session.TrainSpy(idx) // BC 不足時無作用(見 shell.TrainSpy)
			}
			return b.goTo(b.races, uiText(b.lang, "races.transition.screen"))
		}
		if idx, ok := racesSpyMissionActionIndex(a); ok {
			if b.session != nil {
				b.session.CycleSpyMission(idx)
			}
			return b.goTo(b.races, uiText(b.lang, "races.transition.screen"))
		}
		if idx, ok := racesSpyHideActionIndex(a); ok {
			if b.session != nil {
				b.session.SetSpyMission(idx, shell.SpyMissionHide)
			}
			return b.goTo(b.races, uiText(b.lang, "races.transition.screen"))
		}
		if idx, ok := racesDiplomacyActionIndex(a); ok {
			if b.session == nil || idx < 0 || idx >= len(b.session.AIPlayers) {
				return nil
			}
			sc, err := b.diplomacyWith(b.session.AIPlayers[idx].Name)
			if err != nil {
				fmt.Fprintln(os.Stderr, "外交:", err)
				return nil
			}
			return &origTransition{next: sc}
		}
		return b.goTo(b.galaxy, uiText(b.lang, "races.transition.galaxy"))
	}
	// 座標經 PIL 量測(remain-scan/races_a0_f00.png)。
	overlays := []labelRect{
		{200, 14, 240, 22, "RACE RELATIONS", 0},
		// BONUSES 的原始標題帶在 y=365..385；舊的 y=401 是下方空白欄，
		// 會同時留下英文標題又多畫一個漂浮的「加成」。
		{338, 365, 104, 20, "BONUSES", 12},
		{340, 424, 96, 18, "AUDIENCE", 11},
		{340, 442, 96, 18, "DECLARE WAR", 10},
		{438, 424, 90, 18, "REPORT", 11},
		{438, 442, 90, 18, "IGNORE", 11},
		{536, 432, 82, 22, "RETURN", 0},
	}
	s, err := loadOverlayScreen(b.res, "races.lbx", 0, b.lang, b.fnt, "diplo.json",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction, nil)
	if err != nil {
		return nil, err
	}
	// AI 對手即時狀態(名/態勢/軍力/佔星),讓 AI 主動行為可見。
	if b.session != nil && b.fnt != nil {
		gold := color.RGBA{240, 220, 120, 255}
		body := color.RGBA{210, 216, 230, 255}
		dim := color.RGBA{170, 178, 195, 255}
		controlFace := color.RGBA{22, 29, 42, 255}
		dormantControlFace := color.RGBA{17, 22, 31, 255}
		controlBorder := color.RGBA{92, 118, 150, 255}
		// RACES.LBX 在所有七列都烘了英文按鈕字，即使該列尚無已接觸的帝國。
		// 因此先為七列全數畫出對齊原生三槽的底板；未啟用列顏色較暗且不給熱區，
		// 既不留下英文，也不誤導玩家可以點擊。
		for i := 0; i < racesMaxRows; i++ {
			face := dormantControlFace
			if i < len(b.session.AIPlayers) {
				face = controlFace
			}
			for slot := 0; slot < racesSpyButtonSlots; slot++ {
				x, y, w, h := racesSpySlotRect(i, slot)
				s.extraPanels = append(s.extraPanels, extraPanel{x: x + 1, y: y + 1, w: w - 2, h: h - 2,
					fill: face, border: controlBorder})
			}
		}
		for i, a := range b.session.AIPlayers {
			if i >= racesMaxRows {
				break // 版面只放得下 7 個(原版就是這個上限,見 racesSpyAnchors)
			}
			// 資訊文字與對談熱區共用 racesInfoRect：左、右欄都在關係滑桿及間諜鈕前
			// 收束。過去左欄只寬 90px 卻給了 250px maxW，文字必然跨欄。
			s.extras = append(s.extras, racesInfoLineRect(i, 0).leftExtras(b.fnt, aiEmpireLabel(b.lang, a), 11, gold)...)
			s.extras = append(s.extras, racesInfoLineRect(i, 1).leftExtras(b.fnt,
				racesText(b.lang, "races.info.you", infoStanceLabel(b.lang, a.StanceName)), 10, body)...)
			s.extras = append(s.extras, racesInfoLineRect(i, 2).leftExtras(b.fnt,
				racesText(b.lang, "races.info.power_stars", a.FleetStrength, a.OwnedStars), 10, body)...)
			// AI 之間的外交關係(活星系;支撐議會第三方搖擺)。
			rel := ""
			for j := range b.session.AIPlayers {
				if j == i {
					continue
				}
				if rel != "" {
					rel += uiText(b.lang, "races.info.relation_separator")
				}
				rel += racesText(b.lang, "races.info.relation_entry", aiEmpireLabel(b.lang, b.session.AIPlayers[j]),
					infoAIRelationLabel(b.lang, b.session, i, j))
			}
			if rel == "" {
				rel = uiText(b.lang, "races.info.none")
			}
			s.extras = append(s.extras, racesInfoLineRect(i, 3).leftExtras(b.fnt, racesText(b.lang, "races.info.others", rel), 9, dim)...)
			// 三個原生間諜槽各自有置中文字。底板已先覆蓋七列，過去會落入
			// 下一列的 y+23 假按鈕已移除；繪製與命中共用 racesSpySlotRect。
			spies := 0
			if i < len(b.session.PlayerSpies) {
				spies = b.session.PlayerSpies[i]
			}
			spyRect := racesSpySlotTextRect(i, 0)
			missionRect := racesSpySlotTextRect(i, 1)
			hideRect := racesSpySlotTextRect(i, 2)
			s.extras = append(s.extras, centeredExtraTextInSafeRect(spyRect, 10,
				racesText(b.lang, "races.spy.add", spies), color.RGBA{150, 220, 160, 255}))
			mission := b.session.SpyMissionFor(i)
			s.extras = append(s.extras,
				centeredExtraTextInSafeRect(missionRect, 10,
					uiText(b.lang, racesSpyMissionTextKey(mission)), color.RGBA{220, 190, 120, 255}),
				centeredExtraTextInSafeRect(hideRect, 10,
					uiText(b.lang, "races.spy.mission.hide"), color.RGBA{180, 200, 220, 255}),
			)
		}
		agents := b.session.DefensiveAgents
		s.extraPanels = append(s.extraPanels,
			extraPanel{x: racesAgentStatusX, y: racesAgentStatusY, w: racesAgentStatusW, h: racesAgentStatusH,
				fill: controlFace, border: controlBorder},
			extraPanel{x: racesAgentTrainX + 1, y: racesAgentY + 1, w: racesAgentW - 2, h: racesAgentH - 2,
				fill: controlFace, border: controlBorder},
			extraPanel{x: racesAgentDismissX + 1, y: racesAgentY + 1, w: racesAgentW - 2, h: racesAgentH - 2,
				fill: controlFace, border: controlBorder},
		)
		s.extras = append(s.extras,
			centeredExtraTextInSafeRect(racesAgentStatusTextRect(), 10,
				racesText(b.lang, "races.agent.status", agents, b.session.Player.BC),
				color.RGBA{235, 215, 145, 255}),
			centeredExtraTextInSafeRect(racesAgentTrainTextRect(), 10,
				uiText(b.lang, "races.agent.train"), color.RGBA{150, 220, 160, 255}),
			centeredExtraTextInSafeRect(racesAgentDismissTextRect(), 10,
				uiText(b.lang, "races.agent.dismiss"), color.RGBA{220, 180, 160, 255}),
		)
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
		ruleRace := shell.Races[e.shellIdx]
		if enemy == ruleRace.Name || enemy == ruleRace.EnName ||
			enemy == raceSelectEntryText(i18n.Traditional, e, "name") ||
			enemy == raceSelectEntryText(i18n.English, e, "name") ||
			enemy == raceSelectEntryText(i18n.Traditional, e, "adjective") ||
			enemy == raceSelectEntryText(i18n.English, e, "adjective") {
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

type diplomacyOption struct {
	textKey, action string
}

type diplomacyScreen struct {
	b              *sceneBuilder
	fnt            *uifont.Font
	enemy          string
	response       string
	room           *ebiten.Image // 原版 DIPLOMAT 使節房 + 使節疊合
	opts           []diplomacyOption
	backRect       [4]int
	hoverX, hoverY int
}

func newDiplomacyScreen(b *sceneBuilder) *diplomacyScreen {
	// 對談對象改用真正的主要 AI 對手(races 畫面第一個),使外交動作實際改變其關係、可見於
	// 態勢/議會;取不到 session 時退回示範名。
	enemy := uiText(b.lang, "diplomacy.audience.fallback_enemy") // 取不到 session 時的示範名
	if b.session != nil {
		enemy = enemyDisplayName(b.lang, b.session, b.session.PrimaryEnemyName())
	}
	return &diplomacyScreen{b: b, fnt: b.fnt, enemy: enemy, room: loadDiplomatScene(b.res, diplomatRaceIndex(enemy)),
		response: fmt.Sprintf(uiText(b.lang, "diplomacy.audience.opening"), enemy),
		opts: []diplomacyOption{
			{"diplomacy.audience.option.peace", "peace"},
			{"diplomacy.audience.option.trade", "trade"},
			{"diplomacy.audience.option.research", "research"},
			{"diplomacy.audience.option.nonaggression", "nonaggression"},
			{"diplomacy.audience.option.alliance", "alliance"},
			{"diplomacy.audience.option.threat", "threat"},
			{"diplomacy.audience.option.tribute_5", "tribute_5"},
			{"diplomacy.audience.option.tribute_10", "tribute_10"},
			{"diplomacy.audience.option.gift_cash", "gift_cash"},
			{"diplomacy.audience.option.special_food", "special_food"},
			{"diplomacy.audience.option.special_research", "special_research"},
			{"diplomacy.audience.option.gift_tech", "gift_tech"},
			{"diplomacy.audience.option.gift_star", "gift_star"},
		},
		backRect: [4]int{250, 430, 140, 34}}
}

func (d *diplomacyScreen) optRect(i int) (x, y, w, h int) {
	return 16 + (i%3)*208, 190 + (i/3)*42, 192, 34
}

func (d *diplomacyScreen) optTextRect(i int) textSafeRect {
	x, y, w, h := d.optRect(i)
	return textSafeRect{x: x, y: y, w: w, h: h, insetX: 6, insetY: 3}
}

func (d *diplomacyScreen) breakOptions() []diplomacyOption {
	state := d.b.session.TreatyFor(d.enemy)
	options := make([]diplomacyOption, 0, 3)
	if state.TradeActive {
		options = append(options, diplomacyOption{"diplomacy.audience.break.trade", "break_trade"})
	}
	if state.ResearchActive {
		options = append(options, diplomacyOption{"diplomacy.audience.break.research", "break_research"})
	}
	if state.FormalPolicy != gamedata.DIPLO_NONE {
		options = append(options, diplomacyOption{"diplomacy.audience.break.formal", "break_formal"})
	}
	if state.PlayerTribute != shell.TributeNone || state.AITribute != shell.TributeNone {
		options = append(options, diplomacyOption{"diplomacy.audience.break.tribute", "break_tribute"})
	}
	if state.SpecialTrade.Active {
		options = append(options, diplomacyOption{"diplomacy.audience.break.special", "break_special"})
	}
	return options
}

func (d *diplomacyScreen) breakRect(i int) (x, y, w, h int) {
	return 16 + i*124, 398, 116, 24
}

func (d *diplomacyScreen) breakTextRect(i int) textSafeRect {
	x, y, w, h := d.breakRect(i)
	return textSafeRect{x: x, y: y, w: w, h: h, insetX: 6, insetY: 3}
}

func (d *diplomacyScreen) update(in shell.InputState) *origTransition {
	d.hoverX, d.hoverY = in.MouseX, in.MouseY
	if !in.ClickReleased && !in.RightClickReleased {
		return nil
	}
	for i, o := range d.opts {
		x, y, w, h := d.optRect(i)
		if in.MouseX >= x && in.MouseX < x+w && in.MouseY >= y && in.MouseY < y+h {
			d.response = diplomacyResultText(d.b.lang, d.b.session.DiplomacyResponse(o.action, d.enemy))
			return nil
		}
	}
	for i, o := range d.breakOptions() {
		x, y, w, h := d.breakRect(i)
		if in.MouseX >= x && in.MouseX < x+w && in.MouseY >= y && in.MouseY < y+h {
			d.response = diplomacyResultText(d.b.lang, d.b.session.DiplomacyResponse(o.action, d.enemy))
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
		drawPanelImage(dst, d.room, nil)
	}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{235, 232, 245, 255}
	if d.fnt == nil {
		return
	}
	// 上方標題 + 使節台詞(疊半透明深色條增可讀性)。
	fillPanel(dst, 0, 44, moo2ScreenW, 92, color.RGBA{8, 6, 14, 180}, false)
	textSafeRect{x: 32, y: 48, w: 576, h: 28, insetX: 8, insetY: 2}.drawCentered(dst, d.fnt, uiText(d.b.lang, "diplomacy.audience.title"), 20, gold)
	textSafeRect{x: 40, y: 82, w: 560, h: 24, insetX: 4, insetY: 2}.drawCentered(dst, d.fnt,
		fmt.Sprintf(uiText(d.b.lang, "diplomacy.audience.emissary"), d.enemy), 14, color.RGBA{235, 150, 140, 255})
	responseLines := d.fnt.Wrap(d.response, 14, 560)
	if len(responseLines) > 2 {
		responseLines = responseLines[:2]
		responseLines[1] = truncateToWidth(d.fnt, responseLines[1], 14, 560)
	}
	for i, line := range responseLines {
		y := 124.0
		if len(responseLines) > 1 {
			y = 116 + float64(i)*16
		}
		d.fnt.DrawCentered(dst, line, 320, y, 14, body)
	}
	state := d.b.session.TreatyFor(d.enemy)
	textSafeRect{x: 40, y: 151, w: 560, h: 24, insetX: 4, insetY: 2}.drawCentered(dst, d.fnt,
		fmt.Sprintf(uiText(d.b.lang, "diplomacy.audience.agreements"), treatySummaryText(d.b.lang, state)), 13, gold)
	for i, o := range d.opts {
		x, y, w, h := d.optRect(i)
		fillPanel(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{34, 30, 54, 255}, false)
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1.5, color.RGBA{110, 90, 160, 255}, false)
		drawHoverBorder(dst, float32(x), float32(y), float32(w), float32(h), pointInRect(d.hoverX, d.hoverY, x, y, w, h))
		d.optTextRect(i).drawCentered(dst, d.fnt, uiText(d.b.lang, o.textKey), 15, body)
	}
	for i, o := range d.breakOptions() {
		x, y, w, h := d.breakRect(i)
		fillPanel(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{52, 30, 30, 255}, false)
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1.5, color.RGBA{170, 100, 90, 255}, false)
		drawHoverBorder(dst, float32(x), float32(y), float32(w), float32(h), pointInRect(d.hoverX, d.hoverY, x, y, w, h))
		d.breakTextRect(i).drawCentered(dst, d.fnt, uiText(d.b.lang, o.textKey), 11, body)
	}
	bx, by, bw, bh := d.backRect[0], d.backRect[1], d.backRect[2], d.backRect[3]
	fillPanel(dst, float32(bx), float32(by), float32(bw), float32(bh), color.RGBA{40, 34, 30, 255}, false)
	vector.StrokeRect(dst, float32(bx), float32(by), float32(bw), float32(bh), 1.5, color.RGBA{160, 140, 100, 255}, false)
	drawHoverBorder(dst, float32(bx), float32(by), float32(bw), float32(bh), pointInRect(d.hoverX, d.hoverY, bx, by, bw, bh))
	textSafeRect{x: bx, y: by, w: bw, h: bh, insetX: 6, insetY: 3}.drawCentered(dst, d.fnt,
		uiText(d.b.lang, "diplomacy.audience.button.end"), 15, body)
}

// diplomacy 進入外交對談畫面(對象是主要對手)。
func (b *sceneBuilder) diplomacy() (origScreen, error) {
	if b.session == nil {
		return nil, fmt.Errorf("無對局")
	}
	primary := b.session.PrimaryEnemyName()
	playDiplomacyMusic(diplomatRaceIndex(primary), b.session.RelationToPlayer(primary))
	return newDiplomacyScreen(b), nil
}

// diplomacyWith 進入外交對談畫面,對象指定。
//
// 由星圖上緣的會談請求燈用(見 audience.go)——是**那位**對手來敲門,不是主要對手。
func (b *sceneBuilder) diplomacyWith(enemy string) (origScreen, error) {
	if b.session == nil {
		return nil, fmt.Errorf("無對局")
	}
	// 是**那位**對手來敲門,所以曲子也要是那一族的(不是主要對手的)。
	who := enemy
	if who == "" {
		who = b.session.PrimaryEnemyName()
	}
	playDiplomacyMusic(diplomatRaceIndex(who), b.session.RelationToPlayer(who))
	d := newDiplomacyScreen(b)
	if enemy != "" {
		d.enemy = enemyDisplayName(b.lang, b.session, enemy)
		d.room = loadDiplomatScene(b.res, diplomatRaceIndex(enemy))
		d.response = fmt.Sprintf(uiText(b.lang, "diplomacy.audience.opening"), d.enemy)
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

	// 標題和格線之間原本有一條 26px 留白；戰術訊息固定放在這裡，不能再壓到
	// y=351 開始的 COMBAT 控制列。這組座標是 remake 的安全訊息帶，非原版宣稱。
	tacticalMessageX, tacticalMessageY = 58, 48
	tacticalMessageW, tacticalMessageH = 524, 18
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
	// shipMotionStart 記錄本次移動開始的 remake tick；key 包含敵我側與艦名，
	// 讓同名艦也不會共用動畫狀態。原版 timer 未由靜態碼追回，這是固定、可重播
	// 的 CMBTSHP 顯示 adapter。
	shipMotionStart map[string]int
	// squads 是場上的戰機中隊(見 tacticalfighter.go / internal/shell/fighter.go)。
	squads []shell.FighterSquadron
	// combatFX 是 CMBTSFX.LBX 的可選多幀視覺特效；fx 是場上尚未播完的實例。
	combatFX map[int]combatFXSequence
	fx       []combatFXInstance
	tick     int
	// moveLeft 是各我方艦這一回合剩餘的移動格數(第 69 項(戰鬥速度與引擎階))。每回合重置為
	// shell.TacticalMoveSquares(艦的戰鬥速度)。
	moveLeft []int
	// acted / waited 是逐艦行動佇列。原版控制列的 WAIT/DONE 不是裝飾：WAIT 把本艦
	// 放到本回合尚未行動艦之後，DONE 則結束本艦行動；所有艦完成後才結算敵方回擊。
	// 這兩張表只屬於戰術畫面，不進存檔；艦艇陣列被戰損壓縮時會一併重建。
	acted  []bool
	waited []bool
	// initiativeQueue 只在 Ship Initiative 開啟時使用；它保存單場暫態 ID，不能保存
	// 會因戰損壓縮而失效的 player/enemy 切片索引。
	initiativeQueue       []tacticalTurnAction
	initiativePos         int
	initiativeEnemyDamage int
	// mode 是控制列切出來的點擊模式(掃描 / 登艦);見 tacticalbar.go。
	mode           tacticalMode
	hoverX, hoverY int
	// destroyedEnemyHullClassSum 只累加 HP 歸零且未 Captured 的敵艦 SizeClass+1；
	// 戰後交給 shell 套 sub_4B184 的 XP consumer。
	destroyedEnemyHullClassSum int
	// mountDispatch 防止多槽派送遞迴再次展開；只存在本次同步呼叫堆疊。
	mountDispatch bool
	// monsterStar >=0 表示本場敵方是該星的 MonsterGuard；戰後走怪物雙血池回寫。
	monsterStar int
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

// loadCombatShipByIdx 載入 CMBTSHP.LBX 第 idx 個艦艇 sprite 的 frame0,用其所屬色塊的
// palette-holder(索引 45*(idx/45)+44,內嵌調色盤)上色。見 docs/tech/cmbtshp-ship-sprites.md。
// keyColor 用資產自身旗標(CMBTSHP flags=0x0000 → false):艦體外圍透明來自未寫入的
// RLE 像素(ToRGBA 一律留透明),而艦體本身含 index-0 深色像素須保留——先前誤設
// keyColor=true 會把 index-0 艦體也判成透明,導致 sprite 幾乎全消失(端到端截圖查出)。
func loadCombatShipByIdx(res *assets.Resolver, idx int) *ebiten.Image {
	return loadCombatShipByIdxFrame(res, idx, 0)
}

// loadCombatShipByIdxFrame 取 CMBTSHP 每個資產的方向幀。LBX 解碼抽樣確認
// CMBTSHP 每個資產有 20 幀；shell.CMBTSHPFrameForHeading 將 raw 16 向 heading
// 換成最近角度。幀停留的原版 timer 尚未知，tactical 使用 shell 的固定 tick
// adapter：只在艦艇剛移動後播放短掃掠，停止後固定在朝向幀。
func loadCombatShipByIdxFrame(res *assets.Resolver, idx, frame int) *ebiten.Image {
	palIdx := (idx/45)*45 + 44
	prov, err := decodeAsset(res, "cmbtshp.lbx", palIdx)
	if err != nil || prov.Embedded == nil {
		return nil
	}
	im, err := decodeAsset(res, "cmbtshp.lbx", idx)
	if err != nil || len(im.Frames) == 0 {
		return nil
	}
	if frame < 0 || frame >= len(im.Frames) {
		frame = 0
	}
	return ebiten.NewImageFromImage(im.Frames[frame].ToRGBA(prov.Embedded, im.KeyColor()))
}

func newTacticalScreen(b *sceneBuilder) *tacticalScreen {
	p, e := b.session.StartCombat(b.session.PrimaryEnemyName())
	return newTacticalScreenForShips(b, p, e, -1)
}

func newTacticalScreenForShips(b *sceneBuilder, p, e []shell.CombatShip, monsterStar int) *tacticalScreen {
	// 戰鬥 RNG 依當前回合數種子:同一局同一回合的戰鬥可重現(不引入 wall-clock 不確定性)。
	seed := int64(b.session.Turn*2654435761 + 1013904223)
	// 開場先算一次狀態效果,否則第一回合的移動力會用未受牽引的速度(第 69 項(戰鬥速度與引擎階))。
	shell.ApplyTacticalStatusEffects(p, e)
	t := &tacticalScreen{b: b, fnt: b.fnt, player: p, enemy: e, sel: firstReadyShip(p),
		log:    tacticalText(b.lang, "tactical.log.initial"),
		pStart: len(p), eStart: len(e),
		rng: rand.New(rand.NewSource(seed)),
		bg:  loadCombatBG(b.res), bar: loadCombatBar(b.res),
		res: b.res, shipSprites: map[int]*ebiten.Image{}, shipMotionStart: map[string]int{}, combatFX: loadCombatFX(b.res),
		moveLeft: freshMoveBudgets(p), acted: make([]bool, len(p)), waited: make([]bool, len(p)),
		monsterStar: monsterStar}
	t.launchEnemySquadrons()
	if b.session.EffectiveGameSettings().ShipInitiative {
		t.resetInitiativeQueue()
		t.advanceInitiativeQueue()
	}
	return t
}

func firstReadyShip(ships []shell.CombatShip) int {
	if len(ships) == 0 {
		return -1
	}
	return 0
}

// ensureActionQueue 兼容舊測試與未經 newTacticalScreen 建構的戰術狀態。
func (t *tacticalScreen) ensureActionQueue() {
	if len(t.acted) != len(t.player) {
		t.acted = make([]bool, len(t.player))
	}
	if len(t.waited) != len(t.player) {
		t.waited = make([]bool, len(t.player))
	}
}

// nextActionableShip 優先找未等待的艦；若本輪其餘艦都已 WAIT，才讓等待中的艦
// 重新成為目前行動者。這正是 WAIT 的「移到佇列末端」，而不是跳過整回合。
func (t *tacticalScreen) nextActionableShip() int {
	t.ensureActionQueue()
	if t.shipInitiativeEnabled() {
		return t.currentInitiativePlayerIndex()
	}
	for pass := 0; pass < 2; pass++ {
		for i := range t.player {
			if t.acted[i] || (pass == 0 && t.waited[i]) {
				continue
			}
			return i
		}
	}
	return -1
}

func (t *tacticalScreen) allShipsActed() bool {
	t.ensureActionQueue()
	for i := range t.player {
		if !t.acted[i] {
			return false
		}
	}
	return true
}

// finishSelectedAction 將目前艦標成 DONE，若它是本回合最後一艘，便進入回合交界。
func (t *tacticalScreen) finishSelectedAction() {
	t.ensureActionQueue()
	if t.sel < 0 || t.sel >= len(t.player) || t.acted[t.sel] {
		return
	}
	name := t.player[t.sel].Name
	if t.shipInitiativeEnabled() {
		actor := t.sel
		t.completeInitiativePlayerAction(actor)
		return
	}
	t.acted[t.sel] = true
	t.waited[t.sel] = false
	if t.allShipsActed() {
		t.finishRound(0, 0, false, false, 0)
		return
	}
	t.sel = t.nextActionableShip()
	t.log = tacticalText(t.b.lang, "tactical.log.ship_done", name)
}

func (t *tacticalScreen) waitSelectedAction() {
	t.ensureActionQueue()
	if t.sel < 0 || t.sel >= len(t.player) || t.acted[t.sel] {
		return
	}
	name := t.player[t.sel].Name
	if t.shipInitiativeEnabled() {
		t.waitInitiativePlayerAction(t.sel)
		return
	}
	t.waited[t.sel] = true
	next := t.nextActionableShip()
	if next == t.sel {
		t.log = tacticalText(t.b.lang, "tactical.log.no_other_ready")
		return
	}
	t.sel = next
	t.log = tacticalText(t.b.lang, "tactical.log.ship_waits", name)
}

// freshMoveBudgets 依各艦的戰鬥速度算出這一回合的移動格數(第 69 項(戰鬥速度與引擎階))。
func freshMoveBudgets(ships []shell.CombatShip) []int {
	out := make([]int, len(ships))
	for i, sh := range ships {
		// 用**實際**速度:被牽引光束拖慢或被停滯力場定住的船走不了那麼遠(第 69 項(戰鬥速度與引擎階))。
		out[i] = shell.TacticalMoveSquares(shell.TacticalEffectiveSpeed(sh))
	}
	return out
}

func tacticalShipMotionKey(s shell.CombatShip, enemy bool) string {
	return fmt.Sprintf("%t:%s:%d", enemy, s.Name, s.SpriteIdx)
}

func (t *tacticalScreen) markShipMotion(s shell.CombatShip, enemy bool) {
	if t.shipMotionStart == nil {
		t.shipMotionStart = map[string]int{}
	}
	t.shipMotionStart[tacticalShipMotionKey(s, enemy)] = t.tick
}

func (t *tacticalScreen) shipMotionElapsed(s shell.CombatShip, enemy bool) (int, bool) {
	if t.shipMotionStart == nil {
		return 0, false
	}
	start, ok := t.shipMotionStart[tacticalShipMotionKey(s, enemy)]
	if !ok {
		return 0, false
	}
	elapsed := t.tick - start
	if elapsed < 0 || elapsed >= shell.CMBTSHPMotionDurationTicks {
		return 0, false
	}
	return elapsed, true
}

// shipSprite 依 CMBTSHP 資產索引、raw heading 與移動 timer 取(並快取)已解碼
// sprite，避免每幀重解。moving=true 時 elapsed 是本次移動開始後的 tick。
func (t *tacticalScreen) shipSprite(idx, heading, elapsed int, moving bool) *ebiten.Image {
	frame := shell.CMBTSHPFrameAtTick(heading, elapsed, moving)
	cacheKey := idx*shell.CMBTSHPFrameCount + frame
	if im, ok := t.shipSprites[cacheKey]; ok {
		return im
	}
	im := loadCombatShipByIdxFrame(t.res, idx, frame)
	t.shipSprites[cacheKey] = im // 允許 nil(載入失敗),快取避免每幀重試
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
	t.tick++
	t.pruneCombatFX()
	t.hoverX, t.hoverY = in.MouseX, in.MouseY
	if !in.ClickReleased && !in.RightClickReleased {
		return nil
	}
	t.ensureActionQueue()
	if t.over { // 戰後點擊 → 套用結果 → 戰鬥結果畫面
		survivors := map[string]bool{}
		for _, s := range t.player {
			survivors[s.Name] = true
		}
		enemySurvivors := map[string]bool{}
		for _, s := range t.enemy {
			enemySurvivors[s.Name] = true
		}
		if t.monsterStar >= 0 {
			t.b.session.ApplyMonsterTacticalOutcome(t.monsterStar, t.pStart, t.eStart, survivors, t.enemy, t.won)
		} else {
			t.b.session.ApplyCombatOutcomeWithEnemySurvivors(t.b.session.PrimaryEnemyName(), t.pStart, t.eStart, survivors, enemySurvivors,
				t.won, t.destroyedEnemyHullClassSum)
		}
		return t.b.goTo(t.b.battleResult, tacticalText(t.b.lang, "tactical.transition.result"))
	}
	if slot := t.tacticalWeaponSlotAt(in.MouseX, in.MouseY); slot >= 0 {
		if in.RightClickReleased {
			t.describeSelectedWeapon(slot)
		} else {
			t.cycleSelectedWeaponMode(slot)
		}
		return nil
	}
	// 底部控制列(自動/掃描/登船/撤退/等待/完成/選項)。**放在棋盤判定之前**:控制列在
	// y≥365,與棋盤不重疊,但先判定可以讓「按鈕能按」這件事不依賴棋盤範圍算得對不對。
	// COMBAT.LBX 缺席時會改畫本地 fallback 控制列；那不是裝飾，所以不可再把可點擊性
	// 綁在原版美術是否載入成功。
	if i := barButtonHit(in.MouseX, in.MouseY); i >= 0 {
		t.handleBarButton(i)
		return nil
	}
	// 出擊鈕在格線右側(⚠ 不是原版版面,見 tacticalfighter.go launchRect 註解)。
	if lx, ly, lw, lh := launchRect(); t.canLaunchFrom(t.sel) && hitBox(in.MouseX, in.MouseY, lx, ly, lw, lh) {
		if clickSound != nil {
			clickSound()
		}
		t.launchFrom(t.sel)
		return nil
	}
	col, row, ok := cellAt(in.MouseX, in.MouseY)
	if !ok {
		return nil
	}
	if pi := shipAt(t.player, col, row); pi >= 0 { // 點我方艦 → 選取
		if t.shipInitiativeEnabled() && pi != t.currentInitiativePlayerIndex() {
			return nil
		}
		if pi < len(t.acted) && t.acted[pi] {
			t.log = tacticalText(t.b.lang, "tactical.log.already_finished", t.player[pi].Name)
			return nil
		}
		if pi < len(t.waited) && t.waited[pi] && t.nextActionableShip() != pi {
			t.log = tacticalText(t.b.lang, "tactical.log.waiting_ship")
			return nil
		}
		t.sel = pi
		return nil
	}
	if ei := shipAt(t.enemy, col, row); ei >= 0 { // 點敵艦 → 依模式:開火 / 掃描 / 登艦
		switch t.mode {
		case tacticalModeScan:
			t.log = t.scanEnemy(ei)
		case tacticalModeBoard:
			t.boardEnemy(ei)
		default:
			t.fireSelectedShip(ei)
		}
		return nil
	}
	if t.sel >= 0 && t.sel < len(t.player) && !t.acted[t.sel] { // 點空格 → 移動選中艦(受戰鬥速度限制)
		sh := &t.player[t.sel]
		// ⚠ 2026-08-08(第 69 項(戰鬥速度與引擎階)):先前這裡是**瞬移**——點任何空格都能到,沒有距離限制。
		// 現在照原版走戰鬥速度:一回合能走幾格見 shell.TacticalMoveSquares
		// (原版棋盤 81×68 vs remake 8×6,比例尺約 1:10)。
		//
		// 距離用曼哈頓距離,與同一畫面的射程判定(fireRound 的 `abs(dc)+abs(dr)`)同一個度量
		// ——**兩處用不同度量會讓「走得到卻打不到」變成一種很難解釋的行為**。
		budget := 0
		if t.sel < len(t.moveLeft) {
			budget = t.moveLeft[t.sel]
		}
		need := abs(sh.Col-col) + abs(sh.Row-row)
		if need > budget {
			t.log = tacticalText(t.b.lang, "tactical.log.move_insufficient", sh.Name, budget, need)
			return nil
		}
		sh.Facing = gamedata.CombatFacingForVector(col-sh.Col, row-sh.Row)
		sh.Col, sh.Row = col, row
		if t.sel < len(t.moveLeft) {
			t.moveLeft[t.sel] = budget - need
		}
		t.markShipMotion(*sh, false)
		t.log = tacticalText(t.b.lang, "tactical.log.moved", sh.Name, col, row, t.moveLeft[t.sel])
	}
	return nil
}

// fireRound 保留給「自動」路徑：一次讓所有尚未摧毀的我方艦開火，再完成回合。
// 手動點敵艦走 fireSelectedShip，讓 WAIT/DONE 有真正的逐艦落點。
func (t *tacticalScreen) fireRound(target int) {
	actors := make([]int, len(t.player))
	for i := range actors {
		actors[i] = i
	}
	t.fireRoundForActors(target, actors, true)
}

func (t *tacticalScreen) fireSelectedShip(target int) {
	t.ensureActionQueue()
	if t.shipInitiativeEnabled() {
		t.sel = t.currentInitiativePlayerIndex()
		if t.sel < 0 || t.sel >= len(t.player) {
			return
		}
		actor := t.sel
		if !t.fireRoundForActors(target, []int{actor}, false) {
			return
		}
		t.completeInitiativePlayerAction(actor)
		return
	}
	if t.sel < 0 || t.sel >= len(t.player) || t.acted[t.sel] || (t.waited[t.sel] && t.nextActionableShip() != t.sel) {
		t.sel = t.nextActionableShip()
	}
	if t.sel < 0 || t.sel >= len(t.player) {
		t.log = tacticalText(t.b.lang, "tactical.log.no_player_ready")
		return
	}
	actor := t.sel
	last := true
	for i := range t.player {
		if i != actor && !t.acted[i] {
			last = false
			break
		}
	}
	if !t.fireRoundForActors(target, []int{actor}, last) {
		return
	}
	if last {
		// fireRoundForActors 已完成回合交界並重建佇列。
		return
	}
	t.acted[actor] = true
	t.waited[actor] = false
	if !last && !t.over {
		t.sel = t.nextActionableShip()
	}
}

// fireRoundForActors 讓指定艦開火。endRound=false 時只結算這艦的攻擊，敵方與戰機
// 留到本回合最後一艦完成時才處理；endRound=true 則接上完整回合交界。
func (t *tacticalScreen) fireRoundForActors(target int, actors []int, endRound bool) bool {
	if target < 0 || target >= len(t.enemy) {
		return false
	}
	if !t.mountDispatch && t.hasMultiMountActor(actors) {
		return t.fireMultiMountActors(target, actors, endRound)
	}
	if !t.mountDispatch {
		t.ensureWeaponModes()
		ready := actors[:0]
		for _, actor := range actors {
			if actor >= 0 && actor < len(t.player) && len(t.player[actor].WeaponModes) > 0 &&
				t.player[actor].WeaponModes[0] == shell.TacticalWeaponReady {
				ready = append(ready, actor)
			}
		}
		actors = ready
		if len(actors) == 0 {
			t.log = tacticalText(t.b.lang, "tactical.log.no_active_mount")
			return false
		}
	}
	tc, tr := t.enemy[target].Col, t.enemy[target].Row
	// 射程內我艦逐一依武器類型分流真戰鬥公式:beam(ResolveShot,不動)/missile
	// (ResolveMissileShot,躲避+AMR 攔截)/spherical(ResolveSphericalShot,現行武器表
	// 暫無掛載,分支保留供未來串接)。見 shell/weapon_kind.go 的分類依據。
	preCount := len(t.enemy) // 用來判斷本回合是否有敵艦被擊毀(播爆炸音效)
	pAtk, firing := 0, 0
	anyHit := false
	inRange := false
	arcBlocked := false
	ammoBlocked := false
	firedMissile := false // 首艘開火艦是否為飛彈類(決定開火音效)
	firedAny := false
	// 相位匿蹤:手冊「While cloaked, the ship **cannot be attacked**」。與停滯力場同一個形狀
	// (兩者都是「這一發根本沒有目標」),但原因相反——那個是被定住,這個是自己躲起來。
	// 回合數用 t.round+1:t.round 要到本次開火結算完才會 ++。
	if shell.CloakUntargetable(t.enemy[target], t.round+1) {
		t.log = tacticalText(t.b.lang, "tactical.log.phase_cloaked",
			combatShipLabel(t.b.lang, t.b.session, t.enemy[target].Name))
		return false
	}
	if t.enemy[target].InStasis {
		// 手冊(Stasis Field):「cannot … **or be affected by any weapon**. It is
		// effectively removed from battle entirely.」——只做「不能動」會讓它變成活靶,
		// 那是相反的效果。
		t.log = tacticalText(t.b.lang, "tactical.log.in_stasis",
			combatShipLabel(t.b.lang, t.b.session, t.enemy[target].Name))
		return false
	}
	for _, i := range actors {
		if i < 0 || i >= len(t.player) {
			continue
		}
		s := &t.player[i]
		if s.InStasis {
			continue // 被定住的船不能開火
		}
		enemy := &t.enemy[target]
		dist := abs(s.Col-tc) + abs(s.Row-tr)
		if dist > fireRange {
			continue
		}
		inRange = true
		facing := shell.ShieldFacingForShot(*s, *enemy)
		enemy.EnsureShieldFacings()
		shieldReduction := enemy.ShieldReductionForFacing(facing)
		// 行動次數(第 70 項(陀螺去穩器)):超載電容/快速飛彈架/時間扭曲加速器可以再打一次。
		// 沒有這些系統的船 shots==1,整段行為與先前逐位元相同。
		// 能量吸收器:先把儲能射出去(自動命中)。手冊「A cloaked ship will **not** decloak
		// from firing its stored energy」——所以這一發在 CloakOnFire 之前,而且不設 Fired。
		if s.StoredEnergy > 0 {
			dispRoll := 0
			if t.enemy[target].HasDisplacement {
				dispRoll = t.rng.Intn(100) + 1
			}
			er := shell.EnergyAbsorberReleaseAtFacing(s, enemy, dist, dispRoll, shieldReduction)
			if er.Hit {
				anyHit = true
				enemy.ApplyShieldDamage(facing, er.ShieldDamage)
				enemy.ArmorHP = er.RemainingArmorHP
				enemy.HP -= er.DamageToStructure
				pAtk += er.DamageToStructure
			}
		}
		// 原版在每個武器槽的攻擊端以 Relative_Bearing 遮罩檢查射界。
		// 儲能屬於前置特殊武器，已依原版呼叫順序先釋放；正常武器則必須
		// 同時滿足射程與目前 Facing 的方向弧，且被射界擋住時不消費骰子／彈藥。
		if !shell.WeaponArcAllowsCombatShot(*s, *enemy) {
			arcBlocked = true
			continue
		}
		shots := shell.TacticalShotsThisRound(*s)
		if s.Kind == shell.WeaponKindMissile {
			if s.WeaponAmmo <= 0 {
				ammoBlocked = true
				continue
			}
			if shots > s.WeaponAmmo {
				shots = s.WeaponAmmo
			}
		}
		s.Fired = true       // 開過火 → 這一回合結束時不會充能(手冊的 unused 是「完全沒開火」)
		shell.CloakOnFire(s) // 隱形在開火**當下**失效,不是下一回合
		for shot := 0; shot < shots; shot++ {
			firing++
			if !firedAny {
				firedAny = true
				firedMissile = s.Kind == shell.WeaponKindMissile
			}
			facing = shell.ShieldFacingForShot(*s, *enemy)
			shieldReduction = enemy.ShieldReductionForFacing(facing)
			var shot shell.ShotResult
			switch s.Kind {
			case shell.WeaponKindMissile:
				s.WeaponAmmo--
				missileMods := shell.WeaponModCodesForWeapon(s.WeaponName, s.Mods)
				warheads := gamedata.WeaponModMissileWarheadCount(missileMods)
				var mdef shell.MissileDefenses
				mdef.InterceptedWarheads = t.resolveTacticalMissilePointDefense(
					enemy, s.WeaponName, s.DriveLevel, missileMods)
				amrRoll := t.rng.Intn(100) + 1
				jamRoll := t.rng.Intn(100) + 1
				// ⚠ 2026-08-08:上一版寫著「hasAMR/evasion 加成現行皆無對應可造艦元件,
				// 保守傳 0/false」——第 64 項(武器傷害真表)補了反飛彈火箭、第 68 項(元件盤點+飛彈防禦)補了干擾器/慣性穩定器/
				// 閃電場/位移裝置,四項現在都查得到。dist 是實際格距離(比 battleVolley
				// 固定 range=2 更忠實)。
				//
				// 特殊防禦**裝了才擲骰**:沒裝就不動 t.rng,既有戰鬥逐位元不變。
				if enemy.HasLightningField {
					mdef.HasLightningField, mdef.LightningRoll = true, t.rng.Intn(100)+1
				}
				if enemy.HasDisplacement {
					mdef.HasDisplacement, mdef.DisplacementRoll = true, t.rng.Intn(100)+1
				}
				// 匿蹤:手冊「missiles and torpedoes have a 50% chance to miss」。
				if c := shell.CloakMissileMissChance(*enemy, t.round+1); c > 0 {
					mdef.CloakMissChance, mdef.CloakRoll = c, t.rng.Intn(100)+1
				}
				if warheads > 1 {
					mdef.JamRolls = []int{jamRoll}
					if mdef.CloakMissChance > 0 {
						mdef.CloakRolls = []int{mdef.CloakRoll}
					}
					if mdef.HasDisplacement {
						mdef.DisplacementRolls = []int{mdef.DisplacementRoll}
					}
					for i := 1; i < warheads; i++ {
						mdef.JamRolls = append(mdef.JamRolls, t.rng.Intn(100)+1)
						if mdef.CloakMissChance > 0 {
							mdef.CloakRolls = append(mdef.CloakRolls, t.rng.Intn(100)+1)
						}
						if mdef.HasDisplacement {
							mdef.DisplacementRolls = append(mdef.DisplacementRolls, t.rng.Intn(100)+1)
						}
					}
				}
				// ⚠ 第 71 項:第五個引數(攻方掃描器抵銷)先前恆為 0、倒數第二個
				// (目標的硬化護盾)恆為 false。兩個都有真值,只是沒人回頭填。
				shot = shell.ResolveMissileShotWithMods(enemy.HasAMR, dist, amrRoll, enemy.MissileEvasion,
					s.ScannerJamReduction, false, jamRoll,
					s.WeaponMax, shieldReduction, enemy.ArmorHP, enemy.HardShield, mdef,
					s.WeaponName, missileMods)
			case shell.WeaponKindSpherical:
				span := s.WeaponMax - s.WeaponMin
				r := 0
				if span > 0 {
					r = t.rng.Intn(span + 1)
				}
				aggD := gamedata.DamageSphericalRoll(s.WeaponMin, r, 100)
				shot = shell.ResolveSphericalShot(aggD, shieldReduction, enemy.ArmorHP,
					enemy.HardShield, false)
			default:
				roll := t.rng.Intn(100) + 1
				net := s.Attack - shell.TacticalEffectiveDefenseAtRound(*enemy, t.round+1)
				shot = shell.ResolveBeamShot(shell.BeamShot{
					NetAttack: net, WeaponMin: s.WeaponMin, WeaponMax: s.WeaponMax,
					RangeSquares: dist, Roll: roll,
					Mods:     shell.WeaponModCodesForWeapon(s.WeaponName, s.Mods),
					Attacker: s.BeamSystems,
					Target: shell.BeamTargetSystems{
						ShieldReduction: shieldReduction, ArmorHP: enemy.ArmorHP,
						APNegated: enemy.APNegated,
						// ⚠ 2026-08-08:HardShield 先前**沒有填**——快速結算那一側同一個漏。
						// 「Resolve* 有這個參數」不等於「呼叫端有填」。
						HardShield: enemy.HardShield,
					},
				})
			}
			if shot.Hit {
				anyHit = true
				t.spawnCombatFX(combatFXImpact, *enemy)
				enemy.ApplyShieldDamage(facing, shot.ShieldDamage)
				// 能量吸收器:轉存 1/4「抵達這艘船」的潛在傷害(見 gamedata.EnergyAbsorberStored)。
				shell.EnergyAbsorberAbsorb(enemy, s.WeaponMax)
				enemy.ArmorHP = shot.RemainingArmorHP
				enemy.HP -= shot.DamageToStructure
				pAtk += shot.DamageToStructure
			}
		}
	}
	if target >= 0 && target < len(t.enemy) && t.enemy[target].HP <= 0 {
		t.spawnCombatFX(combatFXExplosion, t.enemy[target])
	}
	if firing == 0 {
		if ammoBlocked {
			t.log = tacticalText(t.b.lang, "tactical.log.ammo_depleted")
		} else if inRange && arcBlocked {
			t.log = tacticalText(t.b.lang, "tactical.log.outside_arc")
		} else {
			t.log = tacticalText(t.b.lang, "tactical.log.out_of_range")
		}
		return false
	}
	if !endRound {
		t.compactEnemyCasualties()
		t.refreshSquadronCarriers()
		if !t.mountDispatch {
			playSFX(fireSFX(firedMissile))
			if len(t.enemy) < preCount {
				playSFX(sfxExplode)
			} else if anyHit {
				playSFX(sfxHit)
			}
		}
		name := tacticalText(t.b.lang, "tactical.label.player_ship")
		if len(actors) > 0 && actors[0] >= 0 && actors[0] < len(t.player) {
			name = t.player[actors[0]].Name
		}
		t.log = tacticalText(t.b.lang, "tactical.log.action_damage", name, pAtk)
		if len(t.enemy) == 0 {
			t.over, t.won, t.log = true, true, tacticalText(t.b.lang, "tactical.log.victory")
		}
		return true
	}
	return t.finishRound(preCount, pAtk, firedMissile, anyHit, firing)
	/* unreachable legacy full-fleet finalizer retained only as audit reference
	t.round++
	// 戰機中隊與艦砲同一回合行動:先飛+開火,再讓貼身的敵艦還擊它們。
	fDmg := t.advanceSquadrons()
	pAtk += fDmg
	fLost := t.enemyFiresAtSquadrons()
	t.dropDeadSquadrons()
	t.compactEnemyCasualties()
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
			if es.InStasis || t.player[wi].InStasis {
				continue // 被定住的不能打,也不能被打(第 69 項(戰鬥速度與引擎階))
			}
			dist := abs(es.Col-t.player[wi].Col) + abs(es.Row-t.player[wi].Row)
			if dist > fireRange {
				continue
			}
			// 敵艦(genEnemyFleet)沒有個別武器設計,es.Kind 恆為 WeaponKindBeam(既有
			// 簡化,非本輪引入),故還擊固定走 beam 路徑,不需要分流。
			roll := t.rng.Intn(100) + 1
			// 防禦用**實際**值:完全被定住的船有 −20 防禦(手冊 Tractor Beam)。
			net := es.Attack - shell.TacticalEffectiveDefense(t.player[wi])
			shot := shell.ResolveShot(net, es.WeaponMin, es.WeaponMax, dist,
				t.player[wi].ShieldReduction, t.player[wi].ArmorHP, roll,
				t.player[wi].HardShield, false)
			if shot.Hit {
				t.spawnCombatFX(combatFXImpact, t.player[wi])
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
	// 充能推進(第 70 項(陀螺去穩器)):依這一回合有沒有開過火,決定下一回合能不能連射。
	// 放在狀態重算旁邊——兩者都是「回合交界」的處理。
	shell.TacticalAdvanceCharge(t.player)
	shell.TacticalAdvanceCharge(t.enemy)
	// 狀態效果每回合重算(第 69 項(戰鬥速度與引擎階)):產生源被打掉、或目標飛出射程,效果就該消失。
	// 必須在**移動力重置之前**——移動力是依實際速度算的,而實際速度吃這些狀態。
	shell.ApplyTacticalStatusEffects(t.player, t.enemy)
	// ⚠ 移動力重置**必須在戰損壓縮之後**(第 69 項(戰鬥速度與引擎階))。放在 round++ 那裡的話,
	// 下面這個 palive 壓縮會把 t.player 縮短並讓索引往前移,而 moveLeft 還停在舊長度
	// ——選中第 3 艘卻讀到第 5 艘的移動力,而且陣列還會越界。
	t.moveLeft = freshMoveBudgets(t.player)
	t.acted = make([]bool, len(t.player))
	t.waited = make([]bool, len(t.player))
	if t.sel >= len(t.player) {
		t.sel = firstReadyShip(t.player)
	}
	t.log = tacticalText(t.b.lang, "tactical.log.round_volley", t.round, firing, pAtk, eAtk)
	if fDmg > 0 || fLost > 0 {
		t.log += tacticalText(t.b.lang, "tactical.log.fighter_suffix", fDmg, fLost)
	}
	if len(t.enemy) == 0 {
		t.over, t.won, t.log = true, true, tacticalText(t.b.lang, "tactical.log.victory")
	} else if len(t.player) == 0 {
		t.over, t.won, t.log = true, false, tacticalText(t.b.lang, "tactical.log.defeat")
	}
	return true
	*/
}

func (t *tacticalScreen) hasMultiMountActor(actors []int) bool {
	for _, i := range actors {
		if i < 0 || i >= len(t.player) {
			continue
		}
		mounts := t.player[i].WeaponMounts
		if len(mounts) > 1 {
			return true
		}
		if len(mounts) == 1 && mounts[0].WorkingCount > 1 {
			return true
		}
	}
	return false
}

// fireMultiMountActors 以既有單槽戰術公式逐門派送，避免複製命中／護盾／點防禦公式。
// 只有武器欄位暫時替換；Fired、Cloaked、StoredEnergy、目標戰損等艦級狀態原地保留。
func (t *tacticalScreen) fireMultiMountActors(target int, actors []int, endRound bool) bool {
	t.ensureWeaponModes()
	preCount := len(t.enemy)
	targetName := t.enemy[target].Name
	beforeHP := 0
	for i := range t.enemy {
		beforeHP += t.enemy[i].HP
	}
	fired, firedMissile, enabled := 0, false, 0
	t.mountDispatch = true
	defer func() { t.mountDispatch = false }()

	for _, actor := range actors {
		if actor < 0 || actor >= len(t.player) {
			continue
		}
		s := &t.player[actor]
		if len(s.WeaponMounts) == 0 {
			s.WeaponMounts = []shell.ShipWeaponMount{{RawType: -1, Name: s.WeaponName,
				MaxCount: 1, WorkingCount: 1, Arc: s.WeaponArc, Mods: append([]string(nil), s.Mods...),
				Ammo: s.WeaponAmmo, Attack: s.WeaponMax}}
		}
		actorFired := fired
		origKind, origName := s.Kind, s.WeaponName
		origMin, origMax := s.WeaponMin, s.WeaponMax
		origArc, origAmmo, origMods := s.WeaponArc, s.WeaponAmmo, s.Mods
		for mi := range s.WeaponMounts {
			mount := &s.WeaponMounts[mi]
			if mount.Name == "" || mount.WorkingCount <= 0 {
				continue
			}
			if mi >= len(s.WeaponModes) || s.WeaponModes[mi] != shell.TacticalWeaponReady {
				continue
			}
			enabled++
			s.Kind = shell.WeaponKindForName(mount.Name)
			s.WeaponName = mount.Name
			s.WeaponArc = shell.NormalizeWeaponArc(mount.Name, mount.Arc)
			s.WeaponAmmo = shell.NormalizeWeaponAmmo(mount.Name, mount.Ammo)
			if mi == 0 {
				s.WeaponMin, s.WeaponMax = origMin, origMax
			} else {
				s.WeaponMax = mount.Attack
				if s.WeaponMax < 1 {
					s.WeaponMax = origMax
				}
				s.WeaponMin = s.WeaponMax / 2
			}
			s.Mods = append([]string(nil), mount.Mods...)
			if mi == 0 && len(s.Mods) == 0 {
				s.Mods = origMods
			}
			for n := 0; n < mount.WorkingCount && target < len(t.enemy) && t.enemy[target].Name == targetName; n++ {
				if t.fireRoundForActors(target, []int{actor}, false) {
					if fired == 0 {
						firedMissile = s.Kind == shell.WeaponKindMissile
					}
					fired++
				}
			}
			mount.Ammo = s.WeaponAmmo
		}
		if fired > actorFired {
			for mi := range s.WeaponModes {
				if s.WeaponModes[mi] == shell.TacticalWeaponStandby {
					s.WeaponModes[mi] = shell.TacticalWeaponReady
				}
			}
		}
		s.Kind, s.WeaponName = origKind, origName
		s.WeaponMin, s.WeaponMax = origMin, origMax
		s.WeaponArc, s.WeaponAmmo, s.Mods = origArc, origAmmo, origMods
	}
	if fired == 0 {
		if enabled == 0 {
			t.log = tacticalText(t.b.lang, "tactical.log.no_active_mount")
		}
		return false
	}
	afterHP := 0
	for i := range t.enemy {
		afterHP += t.enemy[i].HP
	}
	damage := beforeHP - afterHP
	if damage < 0 {
		damage = 0
	}
	if !endRound {
		playSFX(fireSFX(firedMissile))
		if len(t.enemy) < preCount {
			playSFX(sfxExplode)
		}
		t.log = tacticalText(t.b.lang, "tactical.log.multi_mount_damage", damage)
		return true
	}
	return t.finishRound(preCount, damage, firedMissile, damage > 0, fired)
}

// compactEnemyCasualties 移除無戰力敵艦並只累加真正擊沉者；登艦俘獲的艦不屬於
// sub_4B184 的 destroyed hull-class accumulator。
func (t *tacticalScreen) compactEnemyCasualties() {
	alive := t.enemy[:0]
	for _, ship := range t.enemy {
		if ship.HP > 0 {
			alive = append(alive, ship)
			continue
		}
		if !ship.Captured {
			t.destroyedEnemyHullClassSum += int(ship.SizeClass) + 1
		}
	}
	t.enemy = alive
}

// resolveTacticalMissilePointDefense 讓防守艦所有本回合尚未使用的 typed PD 逐門
// 迎擊同一批來襲彈頭。WeaponModes 刻意不在參數中：紅色關閉不影響自動 PD。
func (t *tacticalScreen) resolveTacticalMissilePointDefense(
	defender *shell.CombatShip,
	missileName string,
	missileFTLLevel int,
	missileMods []gamedata.WeaponModCode,
) int {
	destroyed := 0
	for _, mount := range shell.AvailableTacticalPointDefenseMounts(defender) {
		shell.MarkTacticalPointDefenseMountSpent(defender, mount.Slot)
		for n := 0; n < mount.Count; n++ {
			if !shell.PointDefenseCanEngage(mount.WeaponName, missileName, mount.BeamMods) {
				continue
			}
			pd := shell.ResolvePointDefenseIntercept(shell.PointDefenseShot{
				BeamWeaponName:            mount.WeaponName,
				BeamAttack:                defender.Attack,
				BeamDamageMax:             mount.BeamDamageMax,
				BeamRangeSquares:          0,
				BeamRoll:                  t.rng.Intn(100) + 1,
				BeamSystems:               defender.BeamSystems,
				BeamMods:                  mount.BeamMods,
				MissileWeaponName:         missileName,
				MissileFTLLevel:           missileFTLLevel,
				MissileMods:               missileMods,
				CarriedInterceptionDamage: defender.PointDefenseInterceptionDamage,
			})
			if pd.Fired {
				destroyed += pd.DestroyedWarheads
				defender.PointDefenseInterceptionDamage = pd.RemainingInterceptionDamage
			}
		}
	}
	return destroyed
}

// enemyRetaliationDamage 讓一艘敵艦依自己的 typed 武器槽還擊。一般 genEnemyFleet
// 仍是舊單槽光束代理，因此該路徑的骰數與公式不變；有逐槽資料的敵艦則能發射飛彈，
// 使玩家艦的紅色 PD 例外有真正的來襲飛彈消費端。
func (t *tacticalScreen) enemyRetaliationDamage(enemyIndex, playerIndex int) int {
	if enemyIndex < 0 || enemyIndex >= len(t.enemy) ||
		playerIndex < 0 || playerIndex >= len(t.player) {
		return 0
	}
	attacker := &t.enemy[enemyIndex]
	target := &t.player[playerIndex]
	if attacker.InStasis || target.InStasis || target.HP <= 0 {
		return 0
	}
	if len(attacker.WeaponMounts) == 0 {
		if shell.IsPlasmaFluxName(attacker.WeaponName) {
			if !shell.PlasmaFluxInRange(attacker.Col-target.Col, attacker.Row-target.Row) {
				return 0
			}
			damage, fired := t.enemyPlasmaFluxShot(enemyIndex, attacker.WeaponMin, attacker.WeaponMax, 1)
			attacker.Fired = attacker.Fired || fired
			return damage
		}
		ammo := attacker.WeaponAmmo
		damage, fired := t.enemyWeaponShot(attacker, target, attacker.WeaponName,
			attacker.Mods, attacker.WeaponArc, attacker.WeaponMin, attacker.WeaponMax,
			&ammo)
		if fired && attacker.Kind == shell.WeaponKindMissile {
			attacker.WeaponAmmo = ammo
		}
		attacker.Fired = attacker.Fired || fired
		return damage
	}
	total := 0
	firedAny := false
	for i := range attacker.WeaponMounts {
		mount := &attacker.WeaponMounts[i]
		if mount.Name == "" || mount.WorkingCount <= 0 || target.HP <= 0 {
			continue
		}
		minDamage, maxDamage := mount.Attack/2, mount.Attack
		if i == 0 {
			minDamage, maxDamage = attacker.WeaponMin, attacker.WeaponMax
		}
		if maxDamage <= 0 {
			maxDamage = attacker.WeaponMax
			minDamage = maxDamage / 2
		}
		ammo := shell.NormalizeWeaponAmmo(mount.Name, mount.Ammo)
		if shell.IsPlasmaFluxName(mount.Name) {
			if shell.PlasmaFluxInRange(attacker.Col-target.Col, attacker.Row-target.Row) {
				damage, fired := t.enemyPlasmaFluxShot(enemyIndex, minDamage, maxDamage, mount.WorkingCount)
				total += damage
				firedAny = firedAny || fired
			}
			continue
		}
		for n := 0; n < mount.WorkingCount && target.HP > 0; n++ {
			damage, fired := t.enemyWeaponShot(attacker, target, mount.Name,
				mount.Mods, mount.Arc, minDamage, maxDamage, &ammo)
			total += damage
			firedAny = firedAny || fired
			if !fired && shell.WeaponKindForName(mount.Name) == shell.WeaponKindMissile {
				break
			}
		}
		mount.Ammo = ammo
	}
	attacker.Fired = attacker.Fired || firedAny
	return total
}

// enemyPlasmaFluxShot 對應 sub_ADE18 effect type 2：以射手為中心，同時傷害半徑內雙方艦艇。
func (t *tacticalScreen) enemyPlasmaFluxShot(enemyIndex, weaponMin, weaponMax, count int) (damage int, fired bool) {
	if enemyIndex < 0 || enemyIndex >= len(t.enemy) || count <= 0 {
		return 0, false
	}
	attacker := &t.enemy[enemyIndex]
	if attacker.HP <= 0 || attacker.InStasis {
		return 0, false
	}
	base := 0
	span := weaponMax - weaponMin
	for i := 0; i < count; i++ {
		rolled := weaponMin
		if span > 0 {
			rolled += t.rng.Intn(span + 1)
		}
		if rolled > 0 {
			base += rolled
		}
	}
	if base < 1 {
		base = 1
	}
	apply := func(target *shell.CombatShip) {
		if target == nil || target.HP <= 0 || target.InStasis {
			return
		}
		dx, dy := target.Col-attacker.Col, target.Row-attacker.Row
		attenuated := shell.PlasmaFluxAttenuatedDamage(base, dx, dy)
		if attenuated == 0 {
			return
		}
		rolled := shell.PlasmaFluxSizeDamage(attenuated, target.SizeClass, func(limit int) int {
			return t.rng.Intn(limit) + 1
		})
		facing := shell.ShieldFacingForShot(*attacker, *target)
		target.EnsureShieldFacings()
		shot := shell.ResolveSphericalShot(rolled, target.ShieldReductionForFacing(facing),
			target.ArmorHP, target.HardShield, false)
		if !shot.Hit {
			return
		}
		target.ApplyShieldDamage(facing, shot.ShieldDamage)
		target.ArmorHP = shot.RemainingArmorHP
		target.HP -= shot.DamageToStructure
		damage += shot.DamageToStructure
		t.spawnCombatFX(combatFXImpact, *target)
	}
	for i := range t.player {
		apply(&t.player[i])
	}
	for i := range t.enemy {
		if i != enemyIndex {
			apply(&t.enemy[i])
		}
	}
	for i := range t.squads {
		squad := &t.squads[i]
		if squad.Dead() {
			continue
		}
		attenuated := shell.PlasmaFluxAttenuatedDamage(base,
			squad.Col-attacker.Col, squad.Row-attacker.Row)
		if attenuated == 0 {
			continue
		}
		avoidRoll := t.rng.Intn(2) + 1
		killed := shell.PlasmaFluxFighterCasualties(squad.Alive, squad.HPEach, attenuated, avoidRoll,
			func(limit int) int { return t.rng.Intn(limit) + 1 })
		if killed > squad.Alive {
			killed = squad.Alive
		}
		squad.Alive -= killed
		if squad.Alive <= 0 {
			squad.Alive, squad.HPEach = 0, 0
		}
	}
	return damage, true
}

func (t *tacticalScreen) enemyWeaponShot(
	attacker, target *shell.CombatShip,
	weaponName string,
	rawMods []string,
	arc gamedata.WeaponArc,
	weaponMin, weaponMax int,
	ammo *int,
) (damage int, fired bool) {
	dist := abs(attacker.Col-target.Col) + abs(attacker.Row-target.Row)
	if dist > fireRange {
		return 0, false
	}
	view := *attacker
	view.WeaponName = weaponName
	view.Kind = shell.WeaponKindForName(weaponName)
	view.WeaponArc = shell.NormalizeWeaponArc(weaponName, arc)
	if !shell.WeaponArcAllowsCombatShot(view, *target) {
		return 0, false
	}
	kind := shell.WeaponKindForName(weaponName)
	if kind == shell.WeaponKindBomb {
		return 0, false
	}
	mods := shell.WeaponModCodesForWeapon(weaponName, rawMods)
	facing := shell.ShieldFacingForShot(*attacker, *target)
	target.EnsureShieldFacings()
	shieldReduction := target.ShieldReductionForFacing(facing)
	var shot shell.ShotResult
	switch kind {
	case shell.WeaponKindMissile:
		if ammo == nil || *ammo <= 0 {
			return 0, false
		}
		(*ammo)--
		var defenses shell.MissileDefenses
		defenses.InterceptedWarheads = t.resolveTacticalMissilePointDefense(
			target, weaponName, attacker.DriveLevel, mods)
		amrRoll := t.rng.Intn(100) + 1
		jamRoll := t.rng.Intn(100) + 1
		if target.HasLightningField {
			defenses.HasLightningField = true
			defenses.LightningRoll = t.rng.Intn(100) + 1
		}
		if target.HasDisplacement {
			defenses.HasDisplacement = true
			defenses.DisplacementRoll = t.rng.Intn(100) + 1
		}
		if chance := shell.CloakMissileMissChance(*target, t.round+1); chance > 0 {
			defenses.CloakMissChance = chance
			defenses.CloakRoll = t.rng.Intn(100) + 1
		}
		warheads := gamedata.WeaponModMissileWarheadCount(mods)
		if warheads > 1 {
			defenses.JamRolls = []int{jamRoll}
			if defenses.CloakMissChance > 0 {
				defenses.CloakRolls = []int{defenses.CloakRoll}
			}
			if defenses.HasDisplacement {
				defenses.DisplacementRolls = []int{defenses.DisplacementRoll}
			}
			for i := 1; i < warheads; i++ {
				defenses.JamRolls = append(defenses.JamRolls, t.rng.Intn(100)+1)
				if defenses.CloakMissChance > 0 {
					defenses.CloakRolls = append(defenses.CloakRolls, t.rng.Intn(100)+1)
				}
				if defenses.HasDisplacement {
					defenses.DisplacementRolls = append(defenses.DisplacementRolls, t.rng.Intn(100)+1)
				}
			}
		}
		shot = shell.ResolveMissileShotWithMods(target.HasAMR, dist, amrRoll,
			target.MissileEvasion, attacker.ScannerJamReduction, false, jamRoll,
			weaponMax, shieldReduction, target.ArmorHP, target.HardShield,
			defenses, weaponName, mods)
	case shell.WeaponKindSpherical:
		span := weaponMax - weaponMin
		roll := 0
		if span > 0 {
			roll = t.rng.Intn(span + 1)
		}
		sphericalDamage := gamedata.DamageSphericalRoll(weaponMin, roll, 100)
		shot = shell.ResolveSphericalShot(
			sphericalDamage,
			shieldReduction, target.ArmorHP, target.HardShield, false)
	default:
		if shell.IsCausticSlimeName(weaponName) {
			span := weaponMax - weaponMin
			strength := weaponMin
			if span > 0 {
				strength += t.rng.Intn(span + 1)
			}
			shell.AddCausticSlimeStrength(target, strength)
			t.spawnCombatFX(combatFXImpact, *target)
			return 0, true
		}
		shot = shell.ResolveShotWithMods(
			attacker.Attack-shell.TacticalEffectiveDefense(*target),
			weaponMin, weaponMax, dist, shieldReduction, target.ArmorHP,
			t.rng.Intn(100)+1, target.HardShield, mods,
			attacker.BeamSystems.HEFBonus, target.APNegated)
	}
	if shot.Hit {
		t.spawnCombatFX(combatFXImpact, *target)
		target.ApplyShieldDamage(facing, shot.ShieldDamage)
		target.ArmorHP = shot.RemainingArmorHP
		target.HP -= shot.DamageToStructure
	}
	return shot.DamageToStructure, true
}

// finishRound 結算回合交界：戰機、敵方還擊、充能、狀態與下一回合行動佇列。
func (t *tacticalScreen) finishRound(preCount, pAtk int, firedMissile, anyHit bool, firing int) bool {
	initiative := t.shipInitiativeEnabled()
	t.round++
	fDmg := t.advanceSquadrons()
	pAtk += fDmg
	fLost := t.enemyFiresAtSquadrons()
	t.dropDeadSquadrons()
	t.compactEnemyCasualties()
	if firing > 0 {
		playSFX(fireSFX(firedMissile))
		if len(t.enemy) < preCount {
			playSFX(sfxExplode)
		} else if anyHit {
			playSFX(sfxHit)
		}
	}

	// 敵方還擊我方最脆弱艦(同樣走真戰鬥公式,每艦一發)。
	eAtk := t.initiativeEnemyDamage
	if !initiative && len(t.player) > 0 && len(t.enemy) > 0 {
		wi := 0
		for i := range t.player {
			if t.player[i].HP < t.player[wi].HP {
				wi = i
			}
		}
		enemyOrder := make([]int, len(t.enemy))
		for i := range enemyOrder {
			enemyOrder[i] = i
		}
		for _, i := range enemyOrder {
			eAtk += t.enemyRetaliationDamage(i, wi)
			if t.player[wi].HP <= 0 {
				break
			}
		}
	}
	for i := range t.player {
		eAtk += shell.TickCausticSlime(&t.player[i])
	}
	for i := range t.enemy {
		pAtk += shell.TickCausticSlime(&t.enemy[i])
	}
	t.compactEnemyCasualties()

	palive := t.player[:0]
	for _, s := range t.player {
		if s.HP > 0 {
			palive = append(palive, s)
		}
	}
	t.player = palive
	t.refreshSquadronCarriers()
	if !initiative {
		shell.ExpireTacticalStoredEnergy(t.player)
		shell.ExpireTacticalStoredEnergy(t.enemy)
	}
	shell.TacticalAdvanceCharge(t.player)
	shell.TacticalAdvanceCharge(t.enemy)
	for i := range t.player {
		shell.ResetTacticalPointDefenseSpent(&t.player[i])
	}
	for i := range t.enemy {
		shell.ResetTacticalPointDefenseSpent(&t.enemy[i])
	}
	shell.ApplyTacticalStatusEffects(t.player, t.enemy)
	t.moveLeft = freshMoveBudgets(t.player)
	t.acted = make([]bool, len(t.player))
	t.waited = make([]bool, len(t.player))
	t.sel = firstReadyShip(t.player)
	if firing > 0 {
		t.log = tacticalText(t.b.lang, "tactical.log.round_volley", t.round, firing, pAtk, eAtk)
	} else {
		t.log = tacticalText(t.b.lang, "tactical.log.round_complete", t.round, eAtk)
	}
	if fDmg > 0 || fLost > 0 {
		t.log += tacticalText(t.b.lang, "tactical.log.fighter_suffix", fDmg, fLost)
	}
	if len(t.enemy) == 0 {
		t.over, t.won, t.log = true, true, tacticalText(t.b.lang, "tactical.log.victory")
	} else if len(t.player) == 0 {
		t.over, t.won, t.log = true, false, tacticalText(t.b.lang, "tactical.log.defeat")
	}
	t.initiativeEnemyDamage = 0
	if initiative && !t.over {
		t.resetInitiativeQueue()
		t.advanceInitiativeQueue()
	}
	return true
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
	elapsed, moving := t.shipMotionElapsed(s, enemy)
	if sprite := t.shipSprite(s.SpriteIdx, s.Facing, elapsed, moving); sprite != nil {
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
		drawPanelImage(dst, sprite, op)
	} else {
		fillPanel(dst, float32(x), float32(iconTop), float32(w), float32(iconH), color.RGBA{base.R / 3, base.G / 3, base.B / 3, 255}, false)
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
	fillPanel(dst, float32(x), float32(y), float32(w), float32(labelH), color.RGBA{0, 0, 0, 150}, false)
	if t.fnt != nil {
		name := s.Name
		if enemy {
			name = combatShipLabel(t.b.lang, t.b.session, name)
		}
		t.fnt.Draw(dst, truncateToWidth(t.fnt, name, 10, float64(w-6)), float64(x)+3, float64(y)+1, 10, color.RGBA{235, 240, 250, 255})
	}
	frac := float32(s.HP) / float32(s.MaxHP)
	if frac < 0 {
		frac = 0
	}
	fillPanel(dst, float32(x)+5, float32(y)+float32(h)-8, float32(w-10), 4, color.RGBA{40, 40, 40, 255}, false)
	fillPanel(dst, float32(x)+5, float32(y)+float32(h)-8, (float32(w-10))*frac, 4, base, false)
}

func (t *tacticalScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{0, 0, 0, 255}) // 純黑太空底;STARBG 未寫入處透明,疊上後黑底透出即原版構圖
	if t.bg != nil {
		drawPanelImage(dst, t.bg, nil)
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
		t.fnt.DrawCentered(dst, truncateToWidth(t.fnt, tacticalText(t.b.lang, "tactical.title"), 20, 600), 320, 34, 20, gold)
	}
	for i, s := range t.player {
		t.drawShip(dst, s, color.RGBA{90, 220, 170, 255}, i == t.sel, false)
	}
	for _, s := range t.enemy {
		t.drawShip(dst, s, color.RGBA{235, 110, 100, 255}, false, true)
	}
	t.drawCombatFX(dst)
	t.drawSquadrons(dst) // 戰機畫在艦艇之上(它們是繞著目標飛的)
	t.drawTacticalMessage(dst)
	t.drawLaunchButton(dst)
	t.drawCombatControlDeck(dst)
}

// drawTacticalMessage 把戰鬥提示限制在標題與格線中間的安全帶。舊版把文字中心放在
// y=343，14px 字身會跨進 y=351 的控制列，造成訊息與按鈕重疊；這裡以固定面板與
// 實際欄寬截斷保證不會越過格線或控制列。
func (t *tacticalScreen) drawTacticalMessage(dst *ebiten.Image) {
	if t.fnt == nil {
		return
	}
	message := t.log
	if hint := t.modeHint(); hint != "" {
		message = hint + " · " + message
	}
	fillPanel(dst, tacticalMessageX, tacticalMessageY, tacticalMessageW, tacticalMessageH,
		color.RGBA{5, 10, 20, 205}, false)
	vector.StrokeRect(dst, tacticalMessageX, tacticalMessageY, tacticalMessageW, tacticalMessageH, 1,
		color.RGBA{86, 106, 148, 210}, false)
	t.fnt.DrawCentered(dst, truncateToWidth(t.fnt, message, 12, tacticalMessageW-12),
		tacticalMessageX+tacticalMessageW/2, tacticalMessageY+tacticalMessageH/2, 12,
		color.RGBA{214, 220, 235, 255})
}

// barButtonsCHT 是 COMBAT.LBX#0 控制列按鈕的螢幕中心座標、動作識別字與外部文案鍵。
// 座標於實際戰鬥截圖(gallery)量測;控制列貼在 y=moo2ScreenH-129=351。
// WEAPONS/SPECIALS 兩個欄位標頭在 remake 未用的清單面板內,略過。
var barButtonsCHT = []struct {
	cx, cy  int
	action  string
	textKey string
}{
	{300, 374, "auto", "tactical.button.auto"}, {365, 374, "scan", "tactical.button.scan"},
	{300, 401, "board", "tactical.button.board"}, {365, 401, "retreat", "tactical.button.retreat"},
	{300, 428, "wait", "tactical.button.wait"}, {365, 428, "done", "tactical.button.done"},
	{334, 455, "options", "tactical.button.options"},
}

// 座標來源:**英文模式跑同一張畫廊圖**(擦底整段不畫,COMBAT.LBX#0 的按鈕直接露出來),
// 再掃浮雕亮邊定出六顆鈕的矩形:左欄 x 274..327、右欄 x 338..391(各寬 54),
// 三列 y 365..383 / 392..410 / 419..438,OPTIONS 在 x 300..367 / y 447..462。
//
// 先前的值是照中文截圖目測的,右欄整欄偏右 9px、列也偏下 3~5px —— 擦底板一邊露出按鈕的
// 灰邊、另一邊蓋到旁邊的星空。**用蓋著英文的截圖去量英文的位置**本來就量不準,
// 英文模式那張才是原版美術本身。

// barButtonPlate 是擦底板相對按鈕中心的尺寸。按鈕本體 54×18,板取 52×16 留住浮雕邊框。
const barButtonPlateW, barButtonPlateH = 52, 16

// combatControlDeckY 是原版 COMBAT.LBX#0 控制列貼到 640×480 畫布時的頂緣。
const combatControlDeckY = moo2ScreenH - 129

// drawCombatControlDeck 畫出可操作的控制列。正版資料帶有 COMBAT.LBX 時保留原版美術；
// 完整包若因資料裁切未帶此檔，則使用同座標的本地控制列，避免畫面與操作一起消失。
func (t *tacticalScreen) drawCombatControlDeck(dst *ebiten.Image) {
	if t.bar != nil {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, float64(combatControlDeckY))
		drawPanelImage(dst, t.bar, op)
		// 控制列烘進的英文按鈕疊中文(CLAUDE.md:button 也要中文化)。
		// 英文模式跳過:COMBAT.LBX#0 上本來就是 AUTO / SCAN / BOARD / RETREAT / …。
		if t.b.lang == i18n.Traditional {
			t.drawBarLabelsCHT(dst)
		}
	} else {
		t.drawFallbackCombatBar(dst)
	}
	for _, b := range barButtonsCHT {
		drawHoverBorder(dst, float32(b.cx-barButtonPlateW/2), float32(b.cy-barButtonPlateH/2),
			barButtonPlateW, barButtonPlateH, pointInRect(t.hoverX, t.hoverY, b.cx-27, b.cy-9, 54, 18))
	}
	t.drawTacticalWeaponPanel(dst)
}

const (
	tacticalWeaponPanelX   = 12
	tacticalWeaponPanelY   = 360
	tacticalWeaponSlotW    = 124
	tacticalWeaponSlotH    = 23
	tacticalWeaponSlotGapX = 4
)

func tacticalWeaponSlotRect(i int) [4]int {
	if i < 0 || i > 7 {
		return [4]int{}
	}
	return [4]int{tacticalWeaponPanelX + (i/4)*(tacticalWeaponSlotW+tacticalWeaponSlotGapX),
		tacticalWeaponPanelY + (i%4)*tacticalWeaponSlotH, tacticalWeaponSlotW, tacticalWeaponSlotH - 2}
}

func (t *tacticalScreen) ensureWeaponModes() {
	for i := range t.player {
		n := len(t.player[i].WeaponMounts)
		if n == 0 {
			n = 1
		}
		if len(t.player[i].WeaponModes) < n {
			t.player[i].WeaponModes = append(t.player[i].WeaponModes, make([]shell.TacticalWeaponMode, n-len(t.player[i].WeaponModes))...)
		} else if len(t.player[i].WeaponModes) > n {
			t.player[i].WeaponModes = t.player[i].WeaponModes[:n]
		}
	}
}

func (t *tacticalScreen) tacticalWeaponSlotAt(x, y int) int {
	t.ensureWeaponModes()
	if t.sel < 0 || t.sel >= len(t.player) {
		return -1
	}
	for i := range t.player[t.sel].WeaponModes {
		r := tacticalWeaponSlotRect(i)
		if hitBox(x, y, r[0], r[1], r[2], r[3]) {
			return i
		}
	}
	return -1
}

func (t *tacticalScreen) cycleSelectedWeaponMode(slot int) {
	t.ensureWeaponModes()
	if t.sel < 0 || t.sel >= len(t.player) || slot < 0 || slot >= len(t.player[t.sel].WeaponModes) {
		return
	}
	mode := (t.player[t.sel].WeaponModes[slot] + 1) % 3
	t.player[t.sel].WeaponModes[slot] = mode
	labels := []string{
		tacticalText(t.b.lang, "tactical.weapon.mode.ready_log"),
		tacticalText(t.b.lang, "tactical.weapon.mode.standby_log"),
		tacticalText(t.b.lang, "tactical.weapon.mode.off_log"),
	}
	label := labels[mode]
	t.log = tacticalText(t.b.lang, "tactical.weapon.mode_changed", slot+1, label)
}

func (t *tacticalScreen) describeSelectedWeapon(slot int) {
	if t.sel < 0 || t.sel >= len(t.player) {
		return
	}
	ship := t.player[t.sel]
	name, count, attack, ammo, arc := ship.WeaponName, 1, ship.WeaponMax, ship.WeaponAmmo, ship.WeaponArc
	mods := ship.Mods
	if slot >= 0 && slot < len(ship.WeaponMounts) {
		mount := ship.WeaponMounts[slot]
		name, count, attack, ammo, arc, mods = mount.Name, mount.WorkingCount, mount.Attack, mount.Ammo, mount.Arc, mount.Mods
	}
	arcLabel := shell.WeaponArcLabelZH(arc)
	if t.b.lang == i18n.English {
		arcLabel = shell.WeaponArcLabelEN(arc)
	}
	ammoLabel := tacticalText(t.b.lang, "tactical.weapon.ammo_unlimited")
	if ammo != 255 && ammo >= 0 {
		ammoLabel = fmt.Sprintf("%d", ammo)
	}
	modLabel := "-"
	if len(mods) > 0 {
		modLabel = strings.Join(mods, "/")
	}
	t.log = tacticalText(t.b.lang, "tactical.weapon.description",
		slot+1, tacticalWeaponDisplayName(t.b, name), count, attack, arcLabel, ammoLabel, modLabel)
}

func tacticalWeaponDisplayName(b *sceneBuilder, name string) string {
	for _, c := range shell.WeaponOptions {
		if c.Name == name {
			return componentLabel(b.lang, c)
		}
	}
	if name == "" {
		return tacticalText(b.lang, "tactical.weapon.fallback_name")
	}
	return name
}

func (t *tacticalScreen) drawTacticalWeaponPanel(dst *ebiten.Image) {
	if t.fnt == nil {
		return
	}
	t.ensureWeaponModes()
	if t.sel < 0 || t.sel >= len(t.player) {
		return
	}
	ship := t.player[t.sel]
	for i, mode := range ship.WeaponModes {
		r := tacticalWeaponSlotRect(i)
		name, count, ammo := ship.WeaponName, 1, ship.WeaponAmmo
		if i < len(ship.WeaponMounts) {
			mount := ship.WeaponMounts[i]
			name, count, ammo = mount.Name, mount.WorkingCount, mount.Ammo
		}
		col := color.RGBA{100, 220, 130, 255}
		status := tacticalText(t.b.lang, "tactical.weapon.mode.ready")
		if mode == shell.TacticalWeaponStandby {
			col, status = color.RGBA{235, 170, 70, 255}, tacticalText(t.b.lang, "tactical.weapon.mode.standby")
		} else if mode == shell.TacticalWeaponOff {
			col, status = color.RGBA{225, 90, 85, 255}, tacticalText(t.b.lang, "tactical.weapon.mode.off")
		}
		fillPanel(dst, float32(r[0]), float32(r[1]), float32(r[2]), float32(r[3]), color.RGBA{8, 13, 24, 225}, false)
		vector.StrokeRect(dst, float32(r[0]), float32(r[1]), float32(r[2]), float32(r[3]), 1, col, false)
		label := fmt.Sprintf("%d %s ×%d", i+1, tacticalWeaponDisplayName(t.b, name), count)
		if ammo != 255 && ammo >= 0 {
			label += fmt.Sprintf(" [%d]", ammo)
		}
		t.fnt.Draw(dst, truncateToWidth(t.fnt, label, 9, float64(r[2]-6)), float64(r[0]+3), float64(r[1]+2), 9, col)
		t.fnt.Draw(dst, status, float64(r[0]+3), float64(r[1]+11), 8, col)
		drawHoverBorder(dst, float32(r[0]), float32(r[1]), float32(r[2]), float32(r[3]),
			pointInRect(t.hoverX, t.hoverY, r[0], r[1], r[2], r[3]))
	}
}

// drawFallbackCombatBar 是 COMBAT.LBX 未提供時的可用控制列。按鈕座標、熱區與原版
// 資產路徑共用 barButtonsCHT，故日後補回資產也不會讓滑鼠位置或行為漂移。
func (t *tacticalScreen) drawFallbackCombatBar(dst *ebiten.Image) {
	fillPanel(dst, 0, combatControlDeckY, moo2ScreenW, 129, color.RGBA{12, 16, 28, 255}, false)
	vector.StrokeRect(dst, 0, combatControlDeckY, moo2ScreenW, 129, 1, color.RGBA{91, 110, 146, 255}, false)
	vector.StrokeLine(dst, 8, combatControlDeckY+5, moo2ScreenW-8, combatControlDeckY+5, 1, color.RGBA{44, 65, 99, 255}, false)
	if t.fnt == nil {
		return
	}
	for _, b := range barButtonsCHT {
		x, y := float32(b.cx-27), float32(b.cy-9)
		fillPanel(dst, x, y, 54, 18, color.RGBA{42, 48, 63, 255}, false)
		vector.StrokeRect(dst, x, y, 54, 18, 1, color.RGBA{126, 141, 169, 255}, false)
		label := uiText(t.b.lang, b.textKey)
		t.fnt.DrawCentered(dst, label, float64(b.cx), float64(b.cy), 12, color.RGBA{225, 230, 240, 255})
	}
}

// drawBarLabelsCHT 在原版控制列的英文按鈕上疊深色底 + 中文字,蓋掉烘進的英文。
func (t *tacticalScreen) drawBarLabelsCHT(dst *ebiten.Image) {
	if t.fnt == nil {
		return
	}
	for _, b := range barButtonsCHT {
		x, y := float32(b.cx-barButtonPlateW/2), float32(b.cy-barButtonPlateH/2)
		fillPanel(dst, x, y, barButtonPlateW, barButtonPlateH, color.RGBA{40, 44, 54, 255}, false)
		vector.StrokeRect(dst, x, y, barButtonPlateW, barButtonPlateH, 1, color.RGBA{120, 130, 150, 255}, false)
		t.fnt.DrawCentered(dst, uiText(i18n.Traditional, b.textKey), float64(b.cx), float64(b.cy), 13, color.RGBA{225, 230, 240, 255})
	}
}

// tacticalCombat 進入格子戰術戰鬥畫面。
func (b *sceneBuilder) tacticalCombat() (origScreen, error) {
	playCombatMusic() // Tactical_Combat_ 是唯一呼叫 Play_Combat_Music_ 的地方
	if b.session == nil {
		return nil, fmt.Errorf("無對局")
	}
	return newTacticalScreen(b), nil
}

// battleResult 顯示上一場戰鬥結果(重用 TURNSUM.LBX#0 視窗當通用面板)。點畫面返回種族關係。
func (b *sceneBuilder) battleResult() (*overlayScreen, error) {
	hits, onAction := b.backHit(b.races, "種族關係")
	// 標題以中文直接當 enKey(misc.json 查無 → fallback 回傳自身),擦底覆蓋烘進的 TURN SUMMARY。
	overlays := []labelRect{
		{88, 14, 204, 22, "BATTLE RESULT", 0},
		{158, 324, 64, 18, "CLOSE", 0},
	}
	s, err := loadOverlayScreen(b.res, "turnsum.lbx", 0, b.lang, b.fnt, "misc.json",
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
		outcome, oc := uiText(b.lang, "battle.result.defeat"), lose
		if bt.PlayerWon {
			outcome, oc = uiText(b.lang, "battle.result.victory"), win
		}
		s.extras = battleResultExtras(b.fnt, battleResultTextRect(0),
			fmt.Sprintf(uiText(b.lang, "battle.result.against"), battleResultEnemyText(b.lang, bt)), 15, gold)
		s.extras = append(s.extras, battleResultExtras(b.fnt, battleResultTextRect(1), outcome, 16, oc)...)
		s.extras = append(s.extras, battleResultExtras(b.fnt, battleResultTextRect(2),
			fmt.Sprintf(uiText(b.lang, "battle.result.start"), bt.PlayerStart, bt.EnemyStart), 12, body)...)
		for i, round := range bt.Log { // 逐回合 typed 戰報
			if i >= 6 {
				break
			}
			s.extras = append(s.extras, battleResultExtras(b.fnt, battleResultLogTextRect(i), battleRoundText(b.lang, round), 12, body)...)
		}
		s.extras = append(s.extras, battleResultExtras(b.fnt, battleResultLossTextRect(len(bt.Log)),
			fmt.Sprintf(uiText(b.lang, "battle.result.losses"), bt.PlayerLosses, bt.EnemyLosses), 13, gold)...)
	}
	return s, nil
}

// council 建原版銀河議會畫面(COUNCIL.LBX 資產 1,調色盤鏈 COUNCIL#0)。3D 議事廳,
// 無烘字,疊「銀河議會」標題;點畫面返回種族關係。
func (b *sceneBuilder) council() (*overlayScreen, error) {
	playSceneBGM(trackCouncil) // Main_Council_Screen_ → STREAMHD #19(第 73 項(音樂場景表))
	// 有待回應選舉(AI 當選)時,改用「接受/拒絕」熱區——手冊:議會無法強迫玩家接受決議
	// (RespondToCouncilElection)。其餘狀態下整頁點擊返回種族關係(backHit)。原版議會是 3D
	// 議事廳、無內建 accept/reject 按鈕藝術,故此處以可點擊文字提示補上互動,不偽造浮雕按鈕框
	// (尊重「用原版 LBX、不自創按鈕藝術」;仍疊在原版 council.lbx 底圖上)。
	pending := b.session != nil && b.session.CouncilStatus().Pending != nil
	pendingVote := b.session != nil && b.session.CouncilStatus().PendingVote != nil
	hits, onAction := b.backHit(b.races, "種族關係")
	if pendingVote {
		hits = []hitRegion{{80, 370, 150, 40, "vote0"}, {245, 370, 150, 40, "vote1"}, {410, 370, 150, 40, "abstain"}, {0, 0, moo2ScreenW, moo2ScreenH, "back"}}
		onAction = func(a string) *origTransition {
			switch a {
			case "vote0":
				b.session.RespondToCouncilVote(0)
			case "vote1":
				b.session.RespondToCouncilVote(1)
			case "abstain":
				b.session.RespondToCouncilVote(2)
			default:
				return b.goTo(b.races, "種族關係")
			}
			return b.goTo(b.council, "銀河議會")
		}
	} else if pending {
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
	s, err := loadOverlayScreen(b.res, "council.lbx", 1, b.lang, b.fnt, "misc.json",
		nil, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"council.lbx", 0}})
	if err != nil {
		return nil, err
	}
	// COUNCIL.LBX#1 是 640×480、10 幀的 delta 動畫；逐幀累積播放後停在最後一幀。
	// 失敗時保留 loadOverlayScreen 的 frame 0 靜態 fallback，不讓美術缺失阻塞議會玩法。
	if frames, ferr := loadOverlayAnimationFrames(b.res, "council.lbx", 1,
		paletteChain{{"council.lbx", 0}}); ferr == nil && len(frames) > 1 {
		s.animFrames = frames
		s.animTick = func() int { return b.animTick }
		s.animationStartTick = b.animTick
	}
	if b.fnt != nil {
		gold := color.RGBA{240, 220, 120, 255}
		s.extras = councilCentered(b.fnt, councilTitleTextRect(), uiText(b.lang, "council.title"), 22, gold)
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
				line1 = fmt.Sprintf(uiText(b.lang, "council.victory.decided"), v.Victory.Turn, v.Meetings)
				if v.Victory.Winner == "player" {
					line2, oc = uiText(b.lang, "council.victory.player"), win
				} else {
					line2, oc = fmt.Sprintf(uiText(b.lang, "council.victory.enemy"), v.Victory.Winner), lose
				}
			case v.Pending != nil:
				// 待回應選舉:顯示當選 AI + 兩個可點擊選項(接受落敗 / 拒絕接受繼續遊戲),
				// 對應上方 pending 分支設定的 accept/reject 熱區。
				s.extras = append(s.extras, councilCentered(b.fnt, councilSummaryTextRect(286, 34),
					fmt.Sprintf(uiText(b.lang, "council.pending.winner"), v.Meetings, v.Pending.EnemyName, v.Pending.EnemyVotes, v.Pending.TotalVotes), 16, lose)...)
				s.extras = append(s.extras, councilCentered(b.fnt, councilSummaryTextRect(320, 46),
					uiText(b.lang, "council.pending.explanation"), 14, neutral)...)
				s.extras = append(s.extras, councilCentered(b.fnt, councilDecisionTextRect(0), uiText(b.lang, "council.pending.accept"), 12, lose)...)
				s.extras = append(s.extras, councilCentered(b.fnt, councilDecisionTextRect(1), uiText(b.lang, "council.pending.reject"), 12, win)...)
				line1, line2 = "", ""
			case v.PendingVote != nil:
				p := v.PendingVote
				s.extras = append(s.extras, councilCentered(b.fnt, councilSummaryTextRect(286, 34),
					fmt.Sprintf(uiText(b.lang, "council.vote.prompt"), v.Meetings, p.PlayerBaseVotes), 16, neutral)...)
				s.extras = append(s.extras, councilCentered(b.fnt, councilVoteTextRect(0), p.CandidateName[0], 16, gold)...)
				s.extras = append(s.extras, councilCentered(b.fnt, councilVoteTextRect(1), p.CandidateName[1], 16, gold)...)
				s.extras = append(s.extras, councilCentered(b.fnt, councilVoteTextRect(2), uiText(b.lang, "council.vote.abstain"), 16, neutral)...)
				line1, line2 = "", ""
			case !v.Eligible:
				line1 = uiText(b.lang, "council.status.not_convened")
				line2, oc = uiText(b.lang, "council.status.requirements"), neutral
			default:
				line1 = fmt.Sprintf(uiText(b.lang, "council.status.convened"), v.Meetings+1)
				line2, oc = uiText(b.lang, "council.status.no_majority"), neutral
			}
			// 逐帝國投票明細(僅在議會已成立且尚未分出勝負時攤開;其餘狀態沿用 line1/line2 摘要)。
			if v.Eligible && !v.Victory.Over && v.Pending == nil {
				bd := b.session.CouncilBreakdown()
				if bd.Valid {
					gold := color.RGBA{240, 220, 120, 255}
					line1 = fmt.Sprintf(uiText(b.lang, "council.breakdown.header"),
						v.Meetings+1, bd.Candidates[0], bd.Candidates[1], bd.Threshold, bd.Total)
					s.extras = append(s.extras, councilCentered(b.fnt, councilSummaryTextRect(82, 30), line1, 13, gold)...)
					y := 128.0
					for _, r := range bd.Rows {
						var suffix string
						rc := neutral
						switch {
						case r.IsCandidate:
							suffix, rc = uiText(b.lang, "council.breakdown.candidate"), gold
						case r.Abstained:
							suffix, rc = uiText(b.lang, "council.breakdown.abstains"), lose
						default:
							suffix = fmt.Sprintf(uiText(b.lang, "council.breakdown.votes_for"), r.VotedFor)
						}
						txt := fmt.Sprintf(uiText(b.lang, "council.breakdown.row"), r.Name, r.BaseVotes, suffix)
						s.extras = append(s.extras, councilCentered(b.fnt, councilRowTextRect(int((y-128)/24)), txt, 14, rc)...)
						y += 24
					}
					line1 = ""
					line2, oc = uiText(b.lang, "council.breakdown.explanation"), neutral
				}
			}
			if line1 != "" {
				s.extras = append(s.extras, councilCentered(b.fnt, councilSummaryTextRect(404, 28), line1, 15, neutral)...)
			}
			if line2 != "" {
				s.extras = append(s.extras, councilCentered(b.fnt, councilSummaryTextRect(432, 32), line2, 17, oc)...)
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
		label: func(b *sceneBuilder) string { return newGameDifficultyLabel(b.lang, b.newGameDiff) },
	},
	{
		act: "size", col: 1, row: 0, asset0: 9,
		n:   func(b *sceneBuilder) int { return len(shell.GalaxySizes) },
		idx: func(b *sceneBuilder) int { return b.newGameSize },
		set: func(b *sceneBuilder, i int) { b.newGameSize = i },
		label: func(b *sceneBuilder) string {
			gs := shell.GalaxySizes[b.newGameSize]
			return fmt.Sprintf(uiText(b.lang, "newgame.setting.galaxy_size"),
				newGameGalaxySizeLabel(b.lang, b.newGameSize), gs.Stars)
		},
	},
	{
		act: "age", col: 2, row: 0, asset0: 1,
		n:     func(b *sceneBuilder) int { return len(shell.GalaxyAges) },
		idx:   func(b *sceneBuilder) int { return b.newGameAge },
		set:   func(b *sceneBuilder, i int) { b.newGameAge = i },
		label: func(b *sceneBuilder) string { return newGameGalaxyAgeLabel(b.lang, b.newGameAge) },
	},
	{
		act: "players", col: 0, row: 1, asset0: 13,
		n:   func(b *sceneBuilder) int { return shell.MaxEmpires - shell.MinEmpires + 1 },
		idx: func(b *sceneBuilder) int { return b.newGameEmpires - shell.MinEmpires },
		set: func(b *sceneBuilder, i int) { b.newGameEmpires = shell.MinEmpires + i },
		label: func(b *sceneBuilder) string {
			return fmt.Sprintf(uiText(b.lang, "newgame.setting.empires"), b.newGameEmpires)
		},
	},
	{
		act: "tech", col: 1, row: 1, asset0: 20,
		n:     func(b *sceneBuilder) int { return len(shell.TechLevels) },
		idx:   func(b *sceneBuilder) int { return b.newGameTech },
		set:   func(b *sceneBuilder, i int) { b.newGameTech = i },
		label: func(b *sceneBuilder) string { return newGameTechLevelLabel(b.lang, b.newGameTech) },
	},
}

// ngBoxRect 回傳某設定的值圖格(螢幕座標)。
func ngBoxRect(s ngSetting) (int, int, int, int) {
	return ngOriginX + ngBoxX0 + s.col*ngColStep, ngOriginY + ngBoxY0 + s.row*ngRowStep, ngBoxW, ngBoxH
}

// ngStripRect 回傳某設定的數值列(640×480 畫布絕對座標)。
//
// 完整 NEWGAME 背景量得的可視數值列與對應 selector caller 坐標已是畫布座標；
// 它和上方值圖的相對原點路徑不同，不能再加 ngOriginX/Y。先前把兩套座標系
// 混在一起，讓數值列熱區與文字同步偏到右下，看起來像按鈕文字沒有置中。
func ngStripRect(s ngSetting) (int, int, int, int) {
	return ngStripX0 + s.col*ngStripColStep, ngStripY0 + s.row*ngStripRowStep,
		ngStripW, ngStripH
}

// ngStripTextRect 是 NEW GAME 每條選擇器內的文字安全框。數值列與熱區相同，
// 但文字保留浮雕邊框的安全帶；中心仍與原生選擇器完全相同。
func ngStripTextRect(s ngSetting) textSafeRect {
	x, y, w, h := ngStripRect(s)
	return textSafeRect{x: x, y: y, w: w, h: h, insetX: 5, insetY: 2}
}

// drawNewGameSettingLabel 是 NEW GAME 數值列唯一的文字繪製入口。
// textSafeRect 已經限制寬度；這裡再檢查實際字型高度，避免日後換字型或字級
// 後，置中文字伸出浮雕列。高度不符時採固定較小字級，再不適合就不畫，
// 不以繪圖裁切掩蓋版面錯誤。
func drawNewGameSettingLabel(dst *ebiten.Image, fnt *uifont.Font, r textSafeRect, label string, col color.Color) {
	if fnt == nil || r.maxLines() == 0 {
		return
	}
	for _, size := range []float64{12, 11, 10} {
		clipped := r.clipped(fnt, label, size)
		if clipped == "" {
			return
		}
		_, measuredH := fnt.Measure(clipped, size)
		if measuredH <= float64(r.h-2*r.insetY) {
			r.drawCentered(dst, fnt, clipped, size, col)
			return
		}
	}
}

// newGameBackgroundAsset 回傳兩版 NEWGAME.LBX 的滿版設定背景資產。
// 1.31 只有 30 張資產,滿版背景在 #28；1.50.26 多了三張設定圖,背景順延到 #31。
func newGameBackgroundAsset(v gamedata.GameVersion) int {
	if v == gamedata.VersionCommunity15 {
		return 31
	}
	return 28
}

// newGameBackgroundAssetForCount 以實際封存檔資產數修正版本標籤與資料目錄
// 不一致的情況。使用者可能以 -version 1.5 指向仍只有 1.31 資產的共用目錄；
// 這時 #31 不存在，但兩版都共有的 #28 仍是正確且可顯示的設定背景。
// 若連 #28 都不存在，保留原候選索引，讓呼叫端回報真正的資料錯誤。
func newGameBackgroundAssetForCount(v gamedata.GameVersion, assetCount int) int {
	candidate := newGameBackgroundAsset(v)
	if candidate >= 0 && candidate < assetCount {
		return candidate
	}
	if 28 >= 0 && 28 < assetCount {
		return 28
	}
	return candidate
}

// newGameBackgroundAssetForResolver 只用封存檔 header 的資產數做防呆，不會
// 預先解碼整張背景。真正 1.5 資料仍使用 #31；只有 #31 不存在時才回退 #28。
func newGameBackgroundAssetForResolver(res *assets.Resolver, v gamedata.GameVersion) int {
	candidate := newGameBackgroundAsset(v)
	if res == nil {
		return candidate
	}
	arch, err := res.OpenLBX("newgame.lbx")
	if err != nil {
		return candidate
	}
	return newGameBackgroundAssetForCount(v, arch.Count())
}

// newGameSetup 建原版新遊戲設定畫面(兩版 NEWGAME.LBX 的滿版背景,調色盤鏈
// RACEOPT#4→NEWGAME#1)。ACCEPT 進種族選擇;CANCEL 回主選單。版面來源見上方檔頭區塊。
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
			return b.goTo(b.newGameSetup, uiText(b.lang, "newgame.transition.setup"))
		}
		if a == "accept" {
			// 熱座每一席都要接管一個帝國。若玩家在多人畫面選的真人數
			// 超過 NEW GAME 目前的帝國數,自動把帝國數補到足夠,避免之後
			// 選帝國畫面只能悄悄少開席位。
			if b.pendingHotseat > b.newGameEmpires {
				b.newGameEmpires = b.pendingHotseat
			}
			if b.networkPending && len(b.networkRoster.Players) > b.newGameEmpires {
				b.newGameEmpires = len(b.networkRoster.Players)
			}
			// 原版流程:星系設定 → Accept →【獨立種族選擇畫面】(不在此直接開局)。
			sc, err := b.raceSelect()
			if err != nil {
				fmt.Fprintf(os.Stderr, "載入種族選擇: %v\n", err)
				return nil
			}
			return &origTransition{next: sc}
		}
		return b.goTo(b.menu, uiText(b.lang, "newgame.transition.main_menu"))
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
	s, err := loadOverlayScreen(b.res, "newgame.lbx", newGameBackgroundAssetForResolver(b.res, b.gameVersion), b.lang, b.fnt, "menu.json",
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
				drawPanelImage(dst, im, op)
			}
			if b.fnt == nil {
				continue
			}
			drawNewGameSettingLabel(dst, b.fnt, ngStripTextRect(st), st.label(b), gold)
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
	b.initializeFleetShipSelection()
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
		// LEADERS 以目前艦隊列表中勾選的船作為指派目標;未勾選時由軍官畫面取第一艘。
		{342, 430, 78, 28, "leaders"},
		// ALL(remake 譯「全部」)—— **全選/全不選這支艦隊的艦艇**,不是集結點。
		//
		// ⚠ 2026-08-07 訂正:先前把它接成 `Set_All_Star_Relocations_`,那是**推測**且推錯了。
		// 手冊在兩個地方各講了一次同一件事:
		//   p.32「To select or deselect all of the ships in the window, you can use the All button.」
		//   p.47「All: Selects all of the ships in the fleet to prepare to receive orders.
		//        (If all the ships are already selected, this deselects them instead.)」
		// 括號那句(已全選就變成全不選)是 toggle 語意,照做。
		// Set_All / Clear_All 的真正入口見 relocateall / relocateclear 兩個熱區的註解。
		{346, 384, 70, 18, "selectall"},
		// RETURN 真值座標取自 openorion2 ships.cpp:718 FleetListView
		// RETURN createWidget(556, 430, ...)(原估計 543,432)。
		{556, 430, 84, 28, "return"},
		// ⚠ **以下兩顆不是原版版面。**
		//
		// 原版的 `Set_All_Star_Relocations_` / `Clear_All_Star_Relocations_` 都是從**星圖**的
		// 輸入處理器 `sub_73980` 進來的,而且是**鍵盤事件**不是按鈕:
		//
		//	cmp eax, 0FFFFFBAFh   ; −1105 → Clear_All_Star_Relocations_(玩家)+ 訊息 0x76
		//	cmp eax, 0FFFFFC13h   ; −1005 → 切換 byte_19BED0(「下一次點星要 Set_All」模式)
		//	                      ;         之後點星才呼叫 Set_All_Star_Relocations_ + 訊息 0x77
		//
		// 那組負數 id 是鍵盤來的(同一支函式裡 −1002/−1001 是被拿來與滑鼠 widget id 併列判斷的
		// 替代鍵),而且兩者**差 100**——看起來是「某鍵」與「ALT+同一鍵」。
		// **是哪一顆鍵沒有確認**(id → 鍵碼的對照表還沒追),所以不綁快捷鍵、不猜。
		// 先放兩顆明確標示為 remake 自加的鈕在名冊下方,追出鍵碼之後改成快捷鍵。
		{20, 412, 140, 18, "relocateall"},
		{168, 412, 140, 18, "relocateclear"},
	}
	onAction := func(a string) *origTransition {
		switch a {
		case "design":
			return b.goTo(b.shipDesign, uiText(b.lang, "fleet.transition.design"))
		case "leaders":
			b.officerMsg = ""
			b.officerTab = 1
			return b.goTo(b.officer, uiText(b.lang, "fleet.transition.officers"))
		}
		if strings.HasPrefix(a, "selfleet") && b.session != nil {
			if n, err := strconv.Atoi(a[len("selfleet"):]); err == nil {
				b.session.SelectFleet(n)
				b.shipPick = nil // 新艦隊用自己的索引，並由 Auto Select Ships 決定初始集合。
				return b.goTo(b.fleet, uiText(b.lang, "fleet.transition.fleet"))
			}
			return nil
		}
		if strings.HasPrefix(a, "pickship") && b.session != nil {
			if n, err := strconv.Atoi(a[len("pickship"):]); err == nil {
				if b.shipPick == nil {
					b.shipPick = map[int]bool{}
				}
				b.shipPick[n] = !b.shipPick[n]
				return b.goTo(b.fleet, uiText(b.lang, "fleet.transition.fleet"))
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
				b.shipPick = nil // 艦數與索引已改變，依設定重建目前艦隊的初始集合。
			}
			return b.goTo(b.fleet, uiText(b.lang, "fleet.transition.fleet"))
		}
		switch a {
		case "selectall":
			// 手冊 p.47:已全選就變成全不選。
			b.toggleSelectAllShips()
			return b.goTo(b.fleet, uiText(b.lang, "fleet.transition.fleet"))
		case "relocateall":
			b.beginRelocateAll()
			b.flash(uiText(b.lang, "relocation.prompt.retarget_all"))
			return b.goTo(b.galaxy, uiText(b.lang, "relocation.transition.galaxy"))
		case "relocateclear":
			n := b.session.ClearAllStarRelocations()
			if n == 0 {
				b.flash(uiText(b.lang, "relocation.result.none_to_clear"))
			} else {
				b.flash(fmt.Sprintf(uiText(b.lang, "relocation.result.cleared_count"), n))
			}
			return b.goTo(b.fleet, uiText(b.lang, "relocation.transition.fleet"))
		case "relocate":
			// 原版 `Star_Relocation_` 是兩段點選:先起點星(自己的殖民地)、再終點星。
			// 回到星圖進第一段。
			b.beginRelocatePick()
			b.flash(uiText(b.lang, "relocation.prompt.choose_origin"))
			return b.goTo(b.galaxy, uiText(b.lang, "relocation.transition.galaxy"))
		case "assault":
			// 進安塔蘭王座廳(原版 Main_Antaran_Room),由那個畫面確認後才發動。
			// 前置條件不滿足時照樣進得去——王座廳會逐條講明卡在哪,比「點了沒反應」清楚。
			sc, err := b.antaranRoom()
			if err != nil {
				fmt.Fprintln(os.Stderr, "安塔蘭王座廳:", err)
				return b.goTo(b.fleet, uiText(b.lang, "fleet.transition.fleet"))
			}
			return &origTransition{next: sc}
		}
		return b.goTo(b.galaxy, uiText(b.lang, "fleet.transition.galaxy"))
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
	s, err := loadOverlayScreen(b.res, "fleet.lbx", 0, b.lang, b.fnt, "menu.json",
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
		y := float64(fleetRosterStart + 12)
		for fi := range b.session.Fleets {
			f := &b.session.Fleets[fi]
			// 艦隊標頭:選中標記 + 所在地 + 航行狀態。
			headerKey := "fleet.roster.header.inactive"
			hc := head
			if fi == b.session.SelectedFleet {
				headerKey, hc = "fleet.roster.header.active", sel
			}
			loc := uiText(b.lang, "fleet.roster.unknown_location")
			if f.AtStar >= 0 && f.AtStar < len(b.session.Stars) {
				loc = b.session.Stars[f.AtStar].Name
			}
			title := fmt.Sprintf(uiText(b.lang, headerKey), fi+1, loc, len(f.Ships))
			if f.DestStar >= 0 && f.DestStar < len(b.session.Stars) {
				title += fmt.Sprintf(uiText(b.lang, "fleet.roster.transit"),
					b.session.Stars[f.DestStar].Name, f.ETA)
			}
			s.extras = append(s.extras, fleetHeaderTextRect(int(y)).leftExtras(b.fnt, title, fleetRosterFont, hc)...)
			fleetHits = append(fleetHits, hitRegion{20, int(y) - 9, 304, fleetRosterRowH, fmt.Sprintf("selfleet%d", fi)})
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
					s.extras = append(s.extras, fleetSplitTextRect(int(y)).leftExtras(b.fnt,
						fmt.Sprintf(uiText(b.lang, "fleet.roster.split"), n), fleetRosterFont,
						color.RGBA{150, 230, 180, 255})...)
					fleetHits = append(fleetHits, hitRegion{28, int(y) - 9, 296, fleetRosterRowH, "splitfleet"})
					y += 18
				}
			}
			for si, sh := range f.Ships {
				// 選船(供拆分用):只有目前操作中的艦隊可以選——拆分是對「這一支」做的。
				mk := ""
				nameCol := gold
				if fi == b.session.SelectedFleet {
					if b.shipPick[si] {
						mk, nameCol = uiText(b.lang, "fleet.selection.selected"), sel
					} else {
						mk = uiText(b.lang, "fleet.selection.unselected")
					}
					fleetHits = append(fleetHits, hitRegion{34, int(y) - 9, 290, fleetRosterRowH, fmt.Sprintf("pickship%d", si)})
				}
				s.extras = append(s.extras,
					fleetShipNameTextRect(int(y)).leftExtras(b.fnt, mk+sh.Name, fleetRosterFont, nameCol)...)
				s.extras = append(s.extras,
					fleetShipClassTextRect(int(y)).leftExtras(b.fnt, shipClassLabel(b.lang, sh.Class), fleetRosterFont, body)...)
				// 結構損傷(見 internal/shell/repair.go)。原版是在艦艇資訊面板用損壞色標示,
				// remake 只有結構這一份損傷值,直接寫百分比;完好的船不畫,免得整排都是「損傷 0%」。
				if d := shell.ShipDamagePercent(sh); d > 0 {
					col := color.RGBA{235, 190, 90, 255} // 輕傷:琥珀
					if d >= 50 {
						col = color.RGBA{230, 110, 90, 255} // 重傷:紅
					}
					s.extras = append(s.extras, fleetShipDamageTextRect(int(y)).leftExtras(b.fnt,
						fmt.Sprintf(uiText(b.lang, "fleet.roster.damage"), d), fleetRosterFont, col)...)
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
			s.extras = append(s.extras, centeredExtraTextInSafeRect(
				fleetAntaranEntryTextRect(), 13, uiText(b.lang, "fleet.antares.entry"), warn))
		}
		// ⚠ 兩個 remake 自加的集結點入口(原版是星圖上的鍵盤指令,見上方 hits 的註解)。
		// 字前加「＋」與原版烘在美術上的鈕區隔開來——這個畫面沒有自繪框的機制,
		// 標記只能靠文字本身。
		mark := color.RGBA{190, 225, 215, 255}
		s.extras = append(s.extras,
			centeredExtraTextInSafeRect(relocationRetargetAllTextRect(),
				11, uiText(b.lang, "relocation.button.retarget_all"), mark),
			centeredExtraTextInSafeRect(relocationClearAllTextRect(),
				11, uiText(b.lang, "relocation.button.clear_all"), mark))
	}
	// 艦隊標頭的熱區要等名冊畫完才知道有幾個,所以最後補進去
	// (loadOverlayScreen 已經把 hits 複製走了,直接接在 s.hits 後面)。
	s.hits = append(s.hits, fleetHits...)
	return s, nil
}

// 艦艇設計畫面的原版座標(全部是 sub_6C8F9 / Add_Design_Buttons_ 的立即數,見下方檔頭)。
var (
	shipClassTextKeys = []string{"ship.class.frigate", "ship.class.destroyer", "ship.class.cruiser", "ship.class.battleship", "ship.class.titan", "ship.class.doom_star"}
	// shipClassZH 是 shell 那邊的艦體 key(中文),順序同 shipClassTextKeys / gamedata.CombatShipClass。
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
// ⚠ 尚未套用、但座標已到手(記在 docs/re/01-gap-report.md 第 5 項(NEW GAME 設定畫面)):
//   - 已裝元件清單列:x 55..68、y = 169 + 13i(`imul eax, esi, 0Dh` / `add eax, 0A9h`)
//   - 右上兩個資訊面板:(437..627, 56..95) 與 (437..627, 97..123)
//     remake 現在把元件選擇列排在 x 300..600,與原版這兩個面板的位置不同;
//     要對齊得先確認那兩格在原版顯示什麼(尚未追到繪製端)。
func (b *sceneBuilder) loadShipDesign(hull int) bool {
	if b.session == nil {
		return false
	}
	design, ok := b.session.ShipDesign(hull)
	if !ok {
		return false
	}
	b.designHull = hull
	if b.designMount < 0 || b.designMount >= len(design.WeaponMounts) {
		b.designMount = 0
	}
	mount := design.WeaponMounts[b.designMount]
	weapon, found := 0, false
	for i, c := range shell.WeaponOptions {
		if c.Name == mount.Name {
			weapon, found = i, true
			break
		}
	}
	if !found {
		weapon = design.Weapon // 未知 raw 武器只顯示相容槽；BUILD 仍由 shell fail-closed。
	}
	b.designWeapon, b.designArmor = weapon, design.Armor
	b.designShield, b.designSpecial = design.Shield, design.Special
	if len(design.Specials) > 0 {
		if b.designSpecialMount < 0 || b.designSpecialMount >= len(design.Specials) {
			b.designSpecialMount = 0
		}
		if idx, ok := shell.SpecialOptionIndex(design.Specials[b.designSpecialMount].Name); ok {
			b.designSpecial = idx
		}
	}
	b.designMods = append([]string(nil), mount.Mods...)
	if b.designMount == 0 && len(b.designMods) == 0 {
		b.designMods = append([]string(nil), design.Mods...)
	}
	b.designArc, b.designAmmo = mount.Arc, mount.Ammo
	b.designLoaded = true
	return true
}

func (b *sceneBuilder) saveShipDesign() bool {
	if b.session == nil || !b.designLoaded {
		return false
	}
	design, ok := b.session.ShipDesign(b.designHull)
	if !ok {
		return false
	}
	if !b.session.SetShipDesignMountLoadout(b.designHull, b.designMount, shell.AutoDesignLoadout{
		Weapon: b.designWeapon, Armor: b.designArmor, Shield: b.designShield, Special: b.designSpecial,
		Mods: append([]string(nil), b.designMods...), Arc: b.designArc, Ammo: b.designAmmo,
		RawRole: design.RawRole,
	}) {
		return false
	}
	return b.session.SetShipDesignSpecialMount(b.designHull, b.designSpecialMount, b.designSpecial)
}

// 點艦體等級只切換持久設計；只有底部 BUILD 會依目前 blueprint 建造。
func (b *sceneBuilder) shipDesign() (*overlayScreen, error) {
	playSceneBGM(trackShipDesign) // Design_Screen_ → STREAM #8
	// 舊行為只保留一份巡洋艦暫態選擇；現在第一次進入時從 session 的六筆設計庫載入。
	if b.session != nil && !b.designLoaded {
		b.loadShipDesign(2) // 原本畫面預設巡洋艦；索引與 shipClassZH 一致。
	}
	hits := make([]hitRegion, 0, 20)
	// 六個艦體槽:座標為反組譯真值(見檔頭的 sub_6C8F9 switch 表),不等距。
	for i := range shipClassZH {
		y0, y1 := dsHullY[i][0], dsHullY[i][1]
		hits = append(hits, hitRegion{dsHullX0, y0, dsHullX1 - dsHullX0 + 1, y1 - y0 + 1, fmt.Sprintf("hull:%d", i)})
	}
	for _, action := range []string{"specialprev", "specialnext", "specialadd", "specialdel"} {
		r := designSpecialControlRect(action)
		hits = append(hits, hitRegion{r[0], r[1], r[2], r[3], action})
	}
	hits = append(hits,
		hitRegion{300, 58, 300, 22, "weapon"}, // 元件選擇(點擊各列循環)
		hitRegion{300, 82, 300, 22, "armor"},
		hitRegion{300, 106, 300, 22, "shield"},
		hitRegion{300, 130, 300, 22, "special"},
		hitRegion{300, 151, 300, 17, "arc"},  // 火線角(點擊循環；飛彈固定 360)
		hitRegion{300, 168, 300, 17, "ammo"}, // 標準飛彈彈架 2/5/10/15/20
		hitRegion{dsBtnX[0], dsBtnY, dsBtnW, dsBtnH, "clear"},
		hitRegion{dsBtnX[1], dsBtnY, dsBtnW, dsBtnH, "cancel"},
		hitRegion{dsBtnX[2], dsBtnY, dsBtnW, dsBtnH, "build"},
		hitRegion{0, 0, moo2ScreenW, moo2ScreenH, "back"},
	)
	for _, action := range []string{"mountadd", "mountdel", "mountdec", "mountinc"} {
		r := designMountControlRect(action)
		hits = append(hits, hitRegion{r[0], r[1], r[2], r[3], action})
	}
	for i := 0; i < 8; i++ {
		r := designMountSlotRect(i)
		hits = append(hits, hitRegion{r[0], r[1], r[2], r[3], fmt.Sprintf("mount:%d", i)})
	}
	// 武器改造(mod)勾選:依目前武器顯示已接線且適用的 chip。飛彈／魚雷不再顯示光束專用
	// 改造，避免玩家勾選一個只會增加成本卻不會進入戰鬥公式的選項。
	//
	// ⚠ 熱區與繪製**共用 designModChipRect**。先前是兩份寫死的座標(這裡 8 列、繪製那邊
	// 一列 modChipX),兩份遲早漂移;而且都是 4 欄 × 76px —— **英文標籤塞不下**
	// (`No Range Dissipation (NR)` 在 size 10 就超過 76px),畫出來疊在一起。
	// 改成 2 欄 × 4 列,欄寬從畫布右緣算出來。
	//
	// ⚠ **熱區要插在整頁 back 之前**——命中判定取第一個中的,back 蓋住整個畫面。
	weaponName := shell.WeaponOptions[b.designWeapon].Name
	b.designArc = shell.NormalizeWeaponArc(weaponName, b.designArc)
	designArc := b.designArc
	modOptions := shell.WeaponModOptionsForWeapon(weaponName)
	if b.session != nil {
		modOptions = b.session.WeaponModOptionsForPlayer(b.designWeapon)
	}
	modHits := make([]hitRegion, len(modOptions))
	for i := range modOptions {
		r := designModChipRect(i)
		// 高度與 16px 列距一致，讓點擊區和實際中文字形共用同一邊界。
		modHits[i] = hitRegion{int(r.x), int(r.y), int(r.w), 16, fmt.Sprintf("mod:%d", i)}
	}
	hits = append(modHits, hits...)
	onAction := func(a string) *origTransition {
		switch a { // 循環只跳到「已研究解鎖」的元件
		case "weapon":
			b.designWeapon = b.session.NextUnlockedComponent(shell.WeaponOptions, b.designWeapon)
			b.designMods = shell.FilterWeaponModsForWeapon(shell.WeaponOptions[b.designWeapon].Name, b.designMods)
			b.designArc = shell.DefaultWeaponArc(shell.WeaponOptions[b.designWeapon].Name)
			b.designAmmo = shell.NormalizeWeaponAmmo(shell.WeaponOptions[b.designWeapon].Name, 0)
			b.designMsg = "" // 換元件可能改變空間是否超格,清掉舊的建造提示避免誤導
			b.saveShipDesign()
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		case "armor":
			b.designArmor = b.session.NextUnlockedComponent(shell.ArmorOptions, b.designArmor)
			b.designMsg = ""
			b.saveShipDesign()
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		case "shield":
			b.designShield = b.session.NextUnlockedComponent(shell.ShieldOptions, b.designShield)
			b.designMsg = ""
			b.saveShipDesign()
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		case "special":
			b.designSpecial = b.session.NextUnlockedSpecialForDesign(b.designHull, b.designSpecialMount, b.designSpecial)
			b.designMsg = ""
			b.saveShipDesign()
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		case "specialprev":
			b.saveShipDesign()
			design, _ := b.session.ShipDesign(b.designHull)
			if b.designSpecialMount > 0 && b.designSpecialMount < len(design.Specials) {
				b.designSpecialMount--
				b.loadShipDesign(b.designHull)
			}
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		case "specialnext":
			b.saveShipDesign()
			design, _ := b.session.ShipDesign(b.designHull)
			if b.designSpecialMount+1 < len(design.Specials) {
				b.designSpecialMount++
				b.loadShipDesign(b.designHull)
			}
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		case "specialadd":
			b.saveShipDesign()
			if idx, ok := b.session.AddShipDesignSpecialMount(b.designHull); ok {
				b.designSpecialMount = idx
				b.loadShipDesign(b.designHull)
			}
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		case "specialdel":
			b.saveShipDesign()
			if b.session.RemoveShipDesignSpecialMount(b.designHull, b.designSpecialMount) {
				if b.designSpecialMount > 0 {
					b.designSpecialMount--
				}
				b.loadShipDesign(b.designHull)
			}
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		case "arc":
			b.designArc = shell.CycleWeaponArc(weaponName, b.designArc)
			b.designMsg = ""
			b.saveShipDesign()
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		case "ammo":
			b.designAmmo = shell.CycleWeaponAmmo(weaponName, b.designAmmo)
			b.designMsg = ""
			b.saveShipDesign()
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		case "mountadd":
			b.saveShipDesign()
			if idx, ok := b.session.AddShipDesignMount(b.designHull, b.designMount); ok {
				b.designMount = idx
				b.loadShipDesign(b.designHull)
			}
			b.designMsg = ""
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		case "mountdel":
			b.saveShipDesign()
			if b.session.RemoveShipDesignMount(b.designHull, b.designMount) {
				if b.designMount > 0 {
					b.designMount--
				}
				b.loadShipDesign(b.designHull)
			}
			b.designMsg = ""
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		case "mountdec":
			b.saveShipDesign()
			b.session.AdjustShipDesignMountCount(b.designHull, b.designMount, -1)
			b.loadShipDesign(b.designHull)
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		case "mountinc":
			b.saveShipDesign()
			b.session.AdjustShipDesignMountCount(b.designHull, b.designMount, 1)
			b.loadShipDesign(b.designHull)
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		case "clear":
			if b.session.ResetShipDesign(b.designHull) {
				b.loadShipDesign(b.designHull)
			}
			b.designMsg = ""
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		case "cancel":
			b.saveShipDesign()
			return b.goTo(b.fleet, uiText(b.lang, "shipdesign.transition.fleet"))
		case "build":
			b.saveShipDesign()
			design, _ := b.session.ShipDesign(b.designHull)
			if !b.session.BlueprintDesignFits(design) {
				b.designMsg = shipDesignText(b.lang, "shipdesign.message.no_space", shipClassLabel(b.lang, design.Class))
				return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
			}
			if !b.session.BuildShipDesign(b.designHull) {
				b.designMsg = uiText(b.lang, "shipdesign.message.no_treasury")
				return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
			}
			b.designMsg = ""
			return b.goTo(b.fleet, uiText(b.lang, "shipdesign.transition.fleet"))
		}
		if strings.HasPrefix(a, "mod:") {
			var idx int
			fmt.Sscanf(a, "mod:%d", &idx)
			if idx >= 0 && idx < len(modOptions) {
				b.designMods = shell.ToggleWeaponMod(b.designMods, modOptions[idx])
				b.designMsg = ""
				b.saveShipDesign()
			}
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		}
		if strings.HasPrefix(a, "mount:") {
			var idx int
			fmt.Sscanf(a, "mount:%d", &idx)
			b.saveShipDesign()
			design, _ := b.session.ShipDesign(b.designHull)
			if idx >= 0 && idx < len(design.WeaponMounts) {
				b.designMount = idx
				b.loadShipDesign(b.designHull)
			}
			b.designMsg = ""
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		}
		if strings.HasPrefix(a, "hull:") && b.session != nil {
			var hull int
			if _, err := fmt.Sscanf(a, "hull:%d", &hull); err != nil || hull < 0 || hull >= len(shipClassZH) {
				return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
			}
			b.saveShipDesign()
			b.designMount = 0
			b.loadShipDesign(hull)
			b.designMsg = ""
			return b.goTo(b.shipDesign, uiText(b.lang, "shipdesign.transition.screen"))
		}
		b.saveShipDesign()
		return b.goTo(b.fleet, uiText(b.lang, "shipdesign.transition.fleet"))
	}
	overlays := []labelRect{{255, 12, 320, 24, uiText(b.lang, "shipdesign.title"), 0}}
	// 六列艦體名的擦底帶跟著槽走(各留 2px 邊,不吃到浮雕框)。
	for i, key := range shipClassTextKeys {
		y0, y1 := dsHullY[i][0], dsHullY[i][1]
		overlays = append(overlays, labelRect{
			dsHullX0 + 2, y0 + 2, dsHullX1 - dsHullX0 - 3, y1 - y0 - 3, uiText(b.lang, key), 12})
	}
	// 底部三顆鈕:反組譯真值 (374/461/547, 443)。
	for i, key := range []string{"shipdesign.button.clear", "shipdesign.button.cancel", "shipdesign.button.build"} {
		overlays = append(overlays, labelRect{dsBtnX[i], dsBtnY, dsBtnW, dsBtnH, uiText(b.lang, key), 0})
	}
	s, err := loadOverlayScreen(b.res, "design.lbx", 0, b.lang, b.fnt, "tech.json",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"buffer0.lbx", 0}})
	if err != nil {
		return nil, err
	}
	// 各艦體成本(對齊 MOO2 空殼生產成本)+ 目前國庫,顯示在艦體清單右方。
	if b.fnt != nil && b.session != nil {
		body := color.RGBA{210, 216, 230, 255}
		classes := shipClassZH
		// 原版六筆設計是可選列；用兩條字型路徑都有的 ASCII `>` 標出目前 blueprint，
		// 避免玩家把「點列只切換設計」誤認成仍會立即造船。
		if b.designHull >= 0 && b.designHull < len(dsHullY) {
			y0, y1 := dsHullY[b.designHull][0], dsHullY[b.designHull][1]
			s.extras = append(s.extras, extraText{x: dsHullX0 - 9, y: float64(y0+y1)/2 - 6,
				size: 12, text: uiText(b.lang, "shipdesign.marker.current"), col: color.RGBA{240, 220, 120, 255}})
		}
		// 價格欄保留原槽順序，但以六條連續 16px 安全列容納 runtime 中文字形；
		// 不改原版不等高的艦體點擊槽，兩者中心誤差最多 1px。
		for i, cl := range classes {
			s.extras = append(s.extras, shipDesignHullCostRect(i).leftExtras(b.fnt,
				shipDesignText(b.lang, "shipdesign.hull.cost", shell.ShipCost(cl)), 8, body)...)
		}
		// 四類元件(點擊各列循環選擇),顯示名稱 + 效果 + 成本。
		w := shell.WeaponOptions[b.designWeapon]
		ar := shell.ArmorOptions[b.designArmor]
		sd := shell.ShieldOptions[b.designShield]
		sp := shell.SpecialOptions[b.designSpecial]
		blueprint, _ := b.session.ShipDesign(b.designHull)
		gold := color.RGBA{240, 220, 120, 255}
		rows := []struct {
			label string
			c     shell.Component
			eff   string
		}{
			{uiText(b.lang, "shipdesign.component.weapon"), w, shipDesignText(b.lang, "shipdesign.component.attack", w.Value)},
			{uiText(b.lang, "shipdesign.component.armor"), ar, shipDesignText(b.lang, "shipdesign.component.hp", ar.Value)},
			{uiText(b.lang, "shipdesign.component.shield"), sd, shipDesignText(b.lang, "shipdesign.component.hp", sd.Value)},
			{uiText(b.lang, "shipdesign.component.special"), sp, ""},
		}
		for i, r := range rows {
			// 面板內高只有 55..149(量自背景圖:上邊框 51/53、下邊框 151/153),四列
			// 用 69+24i 排的話末列(「特殊」)畫到 153,**剛好被下邊框切掉一半**。
			// 改成 60+22i → 60/82/104/126,末列墨水收在 140,離下邊框還有 9px。
			s.extras = append(s.extras, shipDesignComponentNameRect(i).leftExtras(b.fnt,
				shipDesignText(b.lang, "shipdesign.component.line", r.label, componentLabel(b.lang, r.c)), 9, gold)...)
			if i < 3 {
				s.extras = append(s.extras, shipDesignComponentEffectRect(i).leftExtras(b.fnt,
					shipDesignText(b.lang, "shipdesign.component.effect_cost", r.eff, r.c.Cost), 8,
					color.RGBA{200, 208, 225, 255})...)
			}
		}
		for _, ctl := range []struct{ action, label string }{
			{"specialprev", uiText(b.lang, "shipdesign.control.previous")},
			{"specialnext", uiText(b.lang, "shipdesign.control.next")},
			{"specialadd", uiText(b.lang, "shipdesign.control.add")},
			{"specialdel", uiText(b.lang, "shipdesign.control.remove")},
		} {
			r := designSpecialControlRect(ctl.action)
			s.extraPanels = append(s.extraPanels, extraPanel{x: r[0], y: r[1], w: r[2], h: r[3],
				fill: color.RGBA{25, 31, 45, 255}, border: color.RGBA{105, 120, 145, 255}})
			s.extras = append(s.extras, centeredExtraTextInSafeRect(shipDesignControlTextRect(r), 8, ctl.label, body))
		}
		s.extras = append(s.extras, shipDesignComponentEffectRect(3).leftExtras(b.fnt,
			shipDesignText(b.lang, "shipdesign.special.slot", b.designSpecialMount+1, len(blueprint.Specials)), 8, body)...)
		arcPercent := gamedata.WeaponArcCostPercent(designArc)
		arcLabel := shell.WeaponArcLabelZH(designArc)
		if b.lang == i18n.English {
			arcLabel = shell.WeaponArcLabelEN(designArc)
		}
		s.extras = append(s.extras, shipDesignArcTextRect().leftExtras(b.fnt,
			shipDesignText(b.lang, "shipdesign.arc", arcLabel, arcPercent), 9, color.RGBA{190, 205, 235, 255})...)
		ammo := shell.NormalizeWeaponAmmo(w.Name, b.designAmmo)
		ammoText := uiText(b.lang, "shipdesign.ammo.fixed")
		if shell.WeaponUsesVariableMissileRack(w.Name) {
			ammoText = shipDesignText(b.lang, "shipdesign.ammo.variable", ammo)
		}
		s.extras = append(s.extras, shipDesignAmmoTextRect().leftExtras(b.fnt,
			ammoText, 9, color.RGBA{190, 205, 235, 255})...)
		designHull := shipClassZH[b.designHull] // shell 的 key(見 shipClassZH 註解)
		total, totalKnown := b.session.BlueprintDesignCost(blueprint)
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
		mountCount := 1
		if b.designMount >= 0 && b.designMount < len(blueprint.WeaponMounts) {
			mountCount = blueprint.WeaponMounts[b.designMount].MaxCount
		}
		for i := 0; i < 8; i++ {
			r := designMountSlotRect(i)
			x, y, w, h := r[0], r[1], r[2], r[3]
			face := color.RGBA{25, 31, 45, 255}
			if i == b.designMount {
				face = color.RGBA{70, 58, 32, 255}
			}
			if i < len(blueprint.WeaponMounts) {
				s.extraPanels = append(s.extraPanels, extraPanel{x: x, y: y, w: w, h: h, fill: face, border: color.RGBA{105, 120, 145, 255}})
				s.extras = append(s.extras, centeredExtraTextInSafeRect(
					shipDesignControlTextRect(r), 8, fmt.Sprintf("%d", i+1), body))
			}
		}
		for _, ctl := range []struct {
			action string
			label  string
		}{
			{"mountadd", uiText(b.lang, "shipdesign.control.mount_add")},
			{"mountdel", uiText(b.lang, "shipdesign.control.mount_delete")},
			{"mountdec", uiText(b.lang, "shipdesign.control.remove")},
			{"mountinc", uiText(b.lang, "shipdesign.control.add")},
		} {
			r := designMountControlRect(ctl.action)
			s.extraPanels = append(s.extraPanels, extraPanel{x: r[0], y: r[1], w: r[2], h: r[3],
				fill: color.RGBA{25, 31, 45, 255}, border: color.RGBA{105, 120, 145, 255}})
			s.extras = append(s.extras, centeredExtraTextInSafeRect(shipDesignControlTextRect(r), 8, ctl.label, body))
		}
		totalText := shipDesignText(b.lang, "shipdesign.total",
			shipClassLabel(b.lang, designHull), total, b.designMount+1, len(blueprint.WeaponMounts), mountCount)
		if !totalKnown {
			totalText = uiText(b.lang, "shipdesign.total.unknown")
		}
		s.extras = append(s.extras, shipDesignTotalTextRect().leftExtras(b.fnt,
			totalText, 9, color.RGBA{170, 220, 180, 255})...)
		s.extras = append(s.extras, shipDesignTreasuryTextRect().leftExtras(b.fnt,
			shipDesignText(b.lang, "shipdesign.treasury", b.session.Player.BC), 9, gold)...)

		// 「已解鎖」或建造錯誤使用固定兩行安全框，不再依折行數推動下游區塊。
		//
		// ⚠ 先前是單行寫死在 x=305:中文版尾巴「(研究科技解鎖進階元件)」被畫布右緣切掉,
		// 英文版更長。**這種缺口截圖看得到、測試看不到**——沒有任何測試在量文字寬度。
		// 這一段之後每加一個元件字串都會更長(特殊系統這一輪就從 32 個變成 38 個),
		// 所以修法不能是「把字改短」,要真的折行。
		unlockText := shipDesignText(b.lang, "shipdesign.unlocked",
			cnt(shell.WeaponOptions), len(shell.WeaponOptions), cnt(shell.ArmorOptions), len(shell.ArmorOptions),
			cnt(shell.ShieldOptions), len(shell.ShieldOptions), cnt(shell.SpecialOptions), len(shell.SpecialOptions))
		if b.designMsg != "" {
			unlockText = b.designMsg
		}
		s.extras = append(s.extras, shipDesignStatusTextRect().leftExtras(b.fnt,
			unlockText, 9, color.RGBA{170, 200, 240, 255})...)

		// 空間預算/已用(依目前選定元件即時計算):逐艦體列出「空間:已用／總」,超格轉紅並標
		// 「空間不足」。底部 BUILD 用同一份 session 判斷擋下建造(不扣款、不入艦隊)，
		// designMsg 顯示擋下提示——顯示與建造驗證共用同一份判斷，不會不一致。
		// ⚠ 六列單欄 17px 間距會**壓到面板下緣的分隔線**(末日之星那一列直接掉進下一格)。
		// 改成 **3 列 × 2 欄**:同樣六筆,高度從 102px 降到 45px,整塊留在面板內。
		// 欄寬 166px 由固定右側內容區 300..632 對半分配。
		s.extras = append(s.extras, shipDesignSpaceHeaderRect().leftExtras(b.fnt,
			uiText(b.lang, "shipdesign.space.header"), 9, gold)...)
		okCol := color.RGBA{170, 220, 180, 255}
		badCol := color.RGBA{230, 90, 90, 255}
		for i, cl := range classes {
			candidate := blueprint
			candidate.Class = cl
			used, known := b.session.BlueprintDesignSpaceUsed(candidate)
			// 總空間同樣要含巨型通量器加成,否則顯示的「已用／總」會與 onAction 的建造判斷不一致
			// ——兩邊本來就共用同一份判斷,這裡改一邊就要改另一邊。
			totalSp := gamedata.ShipHullSpace(gamedata.CombatShipClass(i))
			if b.session != nil {
				totalSp = b.session.HullSpaceFor(cl)
			}
			fits := known && used <= totalSp
			txt := shipDesignText(b.lang, "shipdesign.space.row", shipClassLabel(b.lang, cl), used, totalSp)
			col := okCol
			if !known {
				txt = shipDesignText(b.lang, "shipdesign.space.unknown", shipClassLabel(b.lang, cl))
				col = badCol
			} else if !fits {
				txt += uiText(b.lang, "shipdesign.space.over")
				col = badCol
			}
			s.extras = append(s.extras, shipDesignSpaceRowRect(i).leftExtras(b.fnt, txt, 8, col)...)
		}

		// 武器改造(mod)勾選 chip:順序對齊上方 modOptions 與 mod:0..N 熱區。已勾選
		// 轉金色高亮,未勾選灰色；只有目前武器適用的改造會出現在這裡。
		modHeaderTxt := uiText(b.lang, "shipdesign.mods.available")
		if len(modOptions) == 0 {
			modHeaderTxt = uiText(b.lang, "shipdesign.mods.none")
		}
		s.extras = append(s.extras, shipDesignModHeaderRect().leftExtras(b.fnt, modHeaderTxt, 8, gold)...)
		activeCol := color.RGBA{240, 220, 120, 255}
		inactiveCol := color.RGBA{150, 155, 165, 255}
		for i, mod := range modOptions {
			chipCol := inactiveCol
			if shell.HasWeaponMod(b.designMods, mod) {
				chipCol = activeCol
			}
			modLabel := shell.WeaponModLabelZH(mod)
			if b.lang == i18n.English {
				modLabel = shell.WeaponModLabelEN(mod)
			}
			s.extras = append(s.extras, shipDesignModTextRect(i).leftExtras(b.fnt, modLabel, 8, chipCol)...)
		}
	}
	return s, nil
}

// officer 建原版軍官列表畫面(OFFICER.LBX 資產 0)。座標經 PIL 量測
// (screens-scan/officer_leaderlist.png):頁籤列 y=12-32,按鈕列 y=440-462。
func (b *sceneBuilder) officer() (*overlayScreen, error) {
	// ⚠ 2026-08-08(第 50 項(軍官畫面座標))座標來源升級:openorion2 → **原版執行檔的立即數**。
	//
	// `Add_Officer_Screen_Fields_` @ 0x9264E 逐欄位讀出來的值,與先前照 openorion2
	// `officer.cpp` 抄的差了幾像素——依專案的來源優先序(**反組譯立即數 > openorion2**),
	// 以執行檔為準:
	//
	//	| 元素 | 先前(openorion2/PIL) | 執行檔立即數 |
	//	|---|---|---|
	//	| Colony Leaders 分頁 | x=20 | **x=9** |
	//	| Ship Leaders 分頁 | x=166 | **x=156** |
	//	| HIRE | (313, 440) | (313, **441**) |
	//	| POOL | (388, 440) | (388, **441**) |
	//	| DISMISS | (462, 440) | (**463**, **441**) |
	//	| RETURN | (540, 440) | (**538**, **441**) |
	//	| 清單列中心 | 90/199/308/417 | **88/197/306/415**(列起點 34、高 108、列距 109)|
	//	| 上下捲鈕 | **沒有** | (613, 22) / (613, 170) |
	//
	// 寬高**維持原樣**:執行檔那邊的寬高是 LBX 資產控制碼,不是字面尺寸(見
	// `docs/re/screen-coords-spy-leader.md` §2.3),**沒查到的就不動**。
	//
	// RETURN 那一格先前還自相矛盾:熱區在 538(對的)、疊字標籤在 540。一併對齊。
	hits := officerHitRegions()
	// 清單列是可操作熱區:從艦隊畫面先勾選目標船,再點這裡的艦艇軍官列即可指派／改派。
	for row, y := range officerRowCenters() {
		hits = append(hits, hitRegion{20, int(y) - 54, 280, 108, fmt.Sprintf("officerRow%d", row)})
	}
	onAction := func(a string) *origTransition {
		switch a {
		case "colonyTab":
			b.officerTab, b.officerScroll, b.officerSelectedSet = 0, 0, false
			return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
		case "shipTab":
			b.officerTab, b.officerScroll, b.officerSelectedSet = 1, 0, false
			return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
		}
		if strings.HasPrefix(a, "officerRow") && b.session != nil {
			row, err := strconv.Atoi(a[len("officerRow"):])
			if err != nil || row < 0 || row >= len(officerRowCenters()) {
				return nil
			}
			rosterCount := len(b.session.Leaders) + len(b.session.MercPool)
			entry := b.officerScroll + row
			if entry < 0 || entry >= rosterCount {
				return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
			}
			b.officerSelected = entry
			b.officerSelectedSet = true
			if b.officerHireMode && entry < len(b.session.Leaders) {
				b.officerMsg = uiText(b.lang, "officer.message.hire_existing")
				return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
			}
			// 待雇傭兵尚未進 Leader Pool,不能提前指派。
			if entry >= len(b.session.Leaders) {
				if b.officerHireMode {
					mercIndex := entry - len(b.session.Leaders)
					ld := b.session.MercPool[mercIndex]
					if b.session.HireMercAt(mercIndex) {
						b.officerHireMode = false
						b.officerSelected = len(b.session.Leaders) - 1
						b.officerSelectedSet = true
						b.officerMsg = officerText(b.lang, "officer.message.hired", ld.Name)
					} else {
						b.officerMsg = uiText(b.lang, "officer.message.hire_failed")
					}
					return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
				}
				b.officerMsg = uiText(b.lang, "officer.message.hire_first")
				return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
			}
			ld := b.session.Leaders[entry]
			if b.officerTab == 0 {
				if ld.Ship {
					b.officerMsg = uiText(b.lang, "officer.message.ship_wrong_tab")
					return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
				}
				colonyIndex, ok := b.officerTargetColony()
				if !ok {
					b.officerMsg = uiText(b.lang, "officer.message.no_colony")
					return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
				}
				if current, assigned := b.session.ColonyLeaderFor(colonyIndex); assigned && current.Name == ld.Name {
					b.session.UnassignLeaderFromColony(colonyIndex)
					b.officerMsg = officerText(b.lang, "officer.message.colony_unassigned", ld.Name)
				} else if b.session.AssignLeaderToColony(colonyIndex, entry) {
					b.officerMsg = officerText(b.lang, "officer.message.colony_assigned", ld.Name, colonyIndex+1)
				} else {
					b.officerMsg = uiText(b.lang, "officer.message.colony_assign_failed")
				}
				return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
			}
			if !ld.Ship {
				b.officerMsg = uiText(b.lang, "officer.message.colony_to_ship")
				return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
			}
			fleetIndex, shipIndex, ok := b.officerTargetShip()
			if !ok {
				b.officerMsg = uiText(b.lang, "officer.message.no_ship")
				return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
			}
			ship := &b.session.Fleets[fleetIndex].Ships[shipIndex]
			if ship.OfficerName == ld.Name {
				b.session.UnassignOfficerFromShip(fleetIndex, shipIndex)
				b.officerMsg = officerText(b.lang, "officer.message.ship_unassigned", ld.Name, shipIndex+1)
			} else if b.session.AssignOfficerToShip(fleetIndex, shipIndex, entry) {
				b.officerMsg = officerText(b.lang, "officer.message.ship_assigned", ld.Name, shipIndex+1)
			} else {
				b.officerMsg = uiText(b.lang, "officer.message.ship_assign_failed")
			}
			return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
		}
		switch a {
		case "hire":
			b.officerHireMode = !b.officerHireMode
			if b.officerHireMode {
				b.officerMsg = uiText(b.lang, "officer.message.hire_mode_on")
			} else {
				b.officerMsg = uiText(b.lang, "officer.message.hire_mode_off")
			}
			return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
		case "pool":
			if b.session == nil || !b.officerSelectedSet || b.officerSelected < 0 || b.officerSelected >= len(b.session.Leaders) {
				b.officerMsg = uiText(b.lang, "officer.message.select_hired")
				return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
			}
			ld := b.session.Leaders[b.officerSelected]
			if b.officerTab == 0 {
				if colonyIndex, ok := b.session.AssignedColonyForLeader(ld.Name); ok {
					b.session.UnassignLeaderFromColony(colonyIndex)
					b.officerMsg = officerText(b.lang, "officer.message.returned_colony_pool", ld.Name)
				} else {
					b.officerMsg = uiText(b.lang, "officer.message.colony_not_assigned")
				}
				return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
			}
			if b.session.ReturnShipOfficerToPool(ld.Name) {
				b.officerMsg = officerText(b.lang, "officer.message.returned_ship_pool", ld.Name)
			} else {
				b.officerMsg = uiText(b.lang, "officer.message.only_ship_pool")
			}
			return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
		case "dismiss":
			if b.session == nil || !b.officerSelectedSet || b.officerSelected < 0 || b.officerSelected >= len(b.session.Leaders) {
				b.officerMsg = uiText(b.lang, "officer.message.select_dismiss")
				return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
			}
			ld := b.session.Leaders[b.officerSelected]
			if b.officerTab == 0 {
				if b.session.DismissColonyLeader(ld.Name) {
					b.officerSelectedSet = false
					b.officerMsg = officerText(b.lang, "officer.message.dismissed_colony", ld.Name)
				} else {
					b.officerMsg = uiText(b.lang, "officer.message.only_colony_dismiss")
				}
				return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
			}
			if b.session.DismissShipOfficer(ld.Name) {
				b.officerSelectedSet = false
				b.officerMsg = officerText(b.lang, "officer.message.dismissed_ship", ld.Name)
			} else {
				b.officerMsg = uiText(b.lang, "officer.message.colony_dismiss_ship_tab")
			}
			return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
		case "scrollUp":
			if b.officerScroll > 0 {
				b.officerScroll--
			}
			return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
		case "scrollDown":
			b.officerScroll++
			return b.goTo(b.officer, uiText(b.lang, "officer.transition.screen"))
		case "Return":
			b.officerHireMode = false
			return b.goTo(b.galaxy, uiText(b.lang, "officer.transition.galaxy"))
		}
		return nil
	}
	overlays := officerOverlays()
	s, err := loadOverlayScreen(b.res, "officer.lbx", 0, b.lang, b.fnt, "officer.json",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"buffer0.lbx", 0}})
	if err != nil {
		return nil, err
	}
	// 領袖名單填進左側槽位。槽中心來自**執行檔立即數**(第 50 項(軍官畫面座標)):
	// `Add_Officer_Screen_Fields_` 的清單迴圈建立四列熱區,y 範圍 34–142 / 143–251 /
	// 252–360 / 361–469——列起點 34、高 108、**列距 109**,中心即 88/197/306/415。
	//
	// 先前的 90/199/308/417 是照 openorion2 的 `FIRST_ROW 38 + SLOT_HEIGHT 105/2` 推的,
	// 列距 109 對上了(那一半 openorion2 是對的),起點差 2px。
	//
	// ⚠ 那四列熱區的**語意**沒有 100% 確認(沒讀 `Check_Officer_Fields_`),
	// 但座標本身是立即數,而且四列高度精確一致——手算若有錯不會這麼齊。
	if b.session != nil {
		gold := color.RGBA{240, 220, 120, 255}
		body := color.RGBA{206, 214, 232, 255}
		hireCol := color.RGBA{150, 220, 160, 255} // 可雇用傭兵用綠色標示
		rowY := officerRowCenters()
		if b.officerTab == 0 {
			if colonyIndex, ok := b.officerTargetColony(); ok {
				s.extras = append(s.extras, officerTargetTextRect().leftExtras(b.fnt,
					officerText(b.lang, "officer.target.colony", colonyIndex+1), officerDynamicFont, body)...)
			} else {
				s.extras = append(s.extras, officerTargetTextRect().leftExtras(b.fnt,
					uiText(b.lang, "officer.target.no_colony"), officerDynamicFont, body)...)
			}
		} else if fleetIndex, shipIndex, ok := b.officerTargetShip(); ok {
			sh := b.session.Fleets[fleetIndex].Ships[shipIndex]
			s.extras = append(s.extras, officerTargetTextRect().leftExtras(b.fnt,
				officerText(b.lang, "officer.target.ship", shipIndex+1, sh.Name), officerDynamicFont, body)...)
		} else {
			s.extras = append(s.extras, officerTargetTextRect().leftExtras(b.fnt,
				uiText(b.lang, "officer.target.no_ship"), officerDynamicFont, body)...)
		}
		if b.officerMsg != "" {
			s.extras = append(s.extras, officerMessageTextRect().leftExtras(b.fnt,
				b.officerMsg, officerDynamicFont, gold)...)
		}
		// 捲動:把「已雇用領袖 + 待雇傭兵」當成一份連續清單,捲動位移就是跳過前 n 筆。
		// 夾在 [0, 總筆數-1]——捲過頭會變成一片空白,那看起來像壞掉而不是捲到底。
		roster := make([]shell.Leader, 0, len(b.session.Leaders)+len(b.session.MercPool))
		roster = append(roster, b.session.Leaders...)
		mercFrom := len(roster)
		roster = append(roster, b.session.MercPool...)
		if b.officerScroll > len(roster)-1 {
			b.officerScroll = len(roster) - 1
		}
		if b.officerScroll < 0 {
			b.officerScroll = 0
		}
		// 從捲動位移開始填,一次填滿四列。已雇用的是金色,待雇傭兵是綠色 + 雇用費。
		for row, k := 0, b.officerScroll; row < len(rowY) && k < len(roster); row, k = row+1, k+1 {
			ld := roster[k]
			marker := ""
			if b.officerSelectedSet && b.officerSelected == k {
				marker = uiText(b.lang, "officer.marker.selected")
			}
			if k >= mercFrom {
				if marker == "" {
					marker = uiText(b.lang, "officer.marker.candidate")
				}
				s.extras = append(s.extras, officerNameTextRect(row).leftExtras(b.fnt,
					marker+ld.Name, 12, hireCol)...)
				s.extras = append(s.extras, officerSkillTextRect(row).leftExtras(b.fnt,
					officerText(b.lang, "officer.roster.mercenary", ld.Skill, ld.Level,
						b.session.MercHireCost(ld)), officerDynamicFont, hireCol)...)
				continue
			}
			s.extras = append(s.extras, officerNameTextRect(row).leftExtras(b.fnt, marker+ld.Name, 12, gold)...)
			s.extras = append(s.extras, officerSkillTextRect(row).leftExtras(b.fnt,
				officerText(b.lang, "officer.roster.hired", ld.Skill, ld.Level), officerDynamicFont, body)...)
			if b.officerTab == 0 {
				if ci, ok := b.session.AssignedColonyForLeader(ld.Name); ok {
					s.extras = append(s.extras, officerAssignmentTextRect(row).leftExtras(b.fnt,
						officerText(b.lang, "officer.roster.assigned_colony", ci+1), officerDynamicFont,
						color.RGBA{170, 220, 190, 255})...)
				} else if !ld.Ship {
					s.extras = append(s.extras, officerAssignmentTextRect(row).leftExtras(b.fnt,
						uiText(b.lang, "officer.roster.unassigned"), officerDynamicFont, body)...)
				}
			} else if fi, si, ok := b.session.AssignedShipForOfficer(ld.Name); ok {
				s.extras = append(s.extras, officerAssignmentTextRect(row).leftExtras(b.fnt,
					officerText(b.lang, "officer.roster.assigned_ship", si+1), officerDynamicFont,
					color.RGBA{170, 220, 190, 255})...)
				_ = fi // fleet index is used by the assignment query, ship label is enough here
			} else if ld.Ship {
				s.extras = append(s.extras, officerAssignmentTextRect(row).leftExtras(b.fnt,
					uiText(b.lang, "officer.roster.unassigned"), officerDynamicFont, body)...)
			}
		}
		// 池空且無領袖時,提示傭兵會不定期上門(手冊 p.134)。
		if len(b.session.Leaders) == 0 && len(b.session.MercPool) == 0 {
			s.extras = append(s.extras, officerEmptyTextRect().centeredExtras(b.fnt,
				uiText(b.lang, "officer.roster.empty"), 11, body)...)
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
		{214, 96, 412, 268, "histmetric"}, // 歷史圖表區：點擊循環艦隊／科技／人口／建築
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
				b.infoHistoryMetric = (b.infoHistoryMetric + 1) % 4
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
	s, err := loadOverlayScreen(b.res, "info.lbx", 0, b.lang, b.fnt, "misc.json",
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
	hits, onAction := b.backHit(b.galaxy, turnSummaryText(b.lang, "turnsummary.transition.galaxy"))
	overlays := []labelRect{
		{88, 14, 204, 22, "TURN SUMMARY", 0},
		{158, 324, 64, 18, "CLOSE", 0},
	}
	s, err := loadOverlayScreen(b.res, "turnsum.lbx", 0, b.lang, b.fnt, "misc.json",
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
		s.extras = append(s.extras, turnSummaryBaseRect(0).leftExtras(b.fnt,
			turnSummaryText(b.lang, "turnsummary.report.stardate", year), 15, gold)...)
		s.extras = append(s.extras, turnSummaryBaseRect(1).leftExtras(b.fnt,
			turnSummaryText(b.lang, "turnsummary.report.industry_research", out.TotalNetIndustry, out.TotalResearch), 13, body)...)
		s.extras = append(s.extras, turnSummaryBaseRect(2).leftExtras(b.fnt,
			turnSummaryText(b.lang, "turnsummary.report.food_tax", out.TotalFood, out.TaxRevenue), 13, body)...)
		s.extras = append(s.extras, turnSummaryBaseRect(3).leftExtras(b.fnt,
			turnSummaryText(b.lang, "turnsummary.report.treasury", b.session.Player.BC, out.NetBC), 13, body)...)
		messages := make([]turnSummaryMessage, 0, 12)
		if len(b.session.LastBankruptcy) > 0 {
			buildings, spies, leaders, recovered := 0, 0, 0, 0
			for _, action := range b.session.LastBankruptcy {
				recovered += action.RecoveredBC
				switch action.Kind {
				case shell.BankruptcySellBuilding:
					buildings++
				case shell.BankruptcyDismissSpy:
					spies++
				case shell.BankruptcyDismissLeader:
					leaders++
				}
			}
			messages = append(messages, turnSummaryMessage{
				text: turnSummaryText(b.lang, "turnsummary.report.fiscal_crisis", buildings, spies, leaders, recovered),
				size: 13, col: color.RGBA{240, 150, 100, 255},
			})
		}
		starving := 0
		for _, colony := range out.Colonies {
			if colony.Starving {
				starving++
			}
		}
		if starving > 0 {
			messages = append(messages, turnSummaryMessage{
				text: turnSummaryText(b.lang, "turnsummary.serious.starvation_count", starving),
				size: 13, col: color.RGBA{240, 130, 100, 255},
			})
		}
		if rebellions := len(b.session.LastRebellions); rebellions > 0 {
			messages = append(messages, turnSummaryMessage{
				text: turnSummaryText(b.lang, "turnsummary.serious.rebellion_count", rebellions),
				size: 13, col: color.RGBA{240, 130, 100, 255},
			})
		}
		if out.ResearchDone {
			messages = append(messages, turnSummaryMessage{
				text: turnSummaryText(b.lang, "turnsummary.report.research_complete"),
				size: 14, col: color.RGBA{120, 220, 140, 255},
			})
		}
		for _, notice := range b.session.LastBuilt {
			messages = append(messages, turnSummaryMessage{
				text: turnSummaryBuildNoticeText(b.lang, notice),
				size: 13, col: color.RGBA{120, 220, 140, 255},
			})
		}
		// 隨機事件(繁榮/瘟疫/海盜…)。英文模式讀結構化雙語報告，
		// 不直接把引擎保留給繁中回合摘要的 LastEvent 當成英文顯示字串。
		eventMsg := b.session.LastEvent
		if b.lang != i18n.Traditional && b.session.LastEventReport != nil && b.session.LastEventReport.MessageEN != "" {
			eventMsg = b.session.LastEventReport.MessageEN
		}
		if b.lang != i18n.Traditional && b.session.LastPersistentEventEN != "" {
			if eventMsg != "" {
				eventMsg += " | "
			}
			eventMsg += b.session.LastPersistentEventEN
		}
		if eventMsg != "" {
			messages = append(messages, turnSummaryMessage{
				text: turnSummaryText(b.lang, "turnsummary.report.event", eventMsg),
				size: 13, col: color.RGBA{240, 190, 110, 255},
			})
		}
		// 安塔蘭人入侵警報(紅色醒目)。
		if b.session.LastAntares != "" {
			antaresMsg := b.session.LastAntares
			if b.lang != i18n.Traditional && b.session.LastAntaresEN != "" {
				antaresMsg = b.session.LastAntaresEN
			}
			messages = append(messages, turnSummaryMessage{
				text: antaresMsg, size: 14, col: color.RGBA{240, 110, 90, 255},
			})
		}
		// AI 對手突襲警報(見 shell/ai_attack.go)。擊退用綠字、被打用紅字,
		// 讓玩家一眼看出「這回合的軍備夠不夠」。
		if b.session.LastRaid != "" {
			col := color.RGBA{240, 110, 90, 255}
			if b.session.LastRaidReport != nil && b.session.LastRaidReport.Repelled {
				col = color.RGBA{130, 220, 150, 255}
			}
			raidMsg := b.session.LastRaid
			if b.lang != i18n.Traditional && b.session.LastRaidReport != nil && b.session.LastRaidReport.MessageEN != "" {
				raidMsg = b.session.LastRaidReport.MessageEN
			}
			messages = append(messages, turnSummaryMessage{text: raidMsg, size: 14, col: col})
		}
		s.extras = append(s.extras, turnSummaryDynamicExtras(b.fnt, messages)...)
	}
	return s, nil
}

// researchAreaOrder 把畫面 8 個領域熱區名對應到 gamedata.TechTree() 的領域索引(見
// internal/gamedata/techtree.go 陣列註解:0=Biology…7=Sociology)。
var researchAreaOrder = map[string]int{
	"Construction": 3, "Power": 1, "Chemistry": 5, "Sociology": 7,
	"Computers": 6, "Biology": 0, "Physics": 2, "Force Fields": 4,
}

var researchAreaHits = []hitRegion{
	{16, 32, 208, 98, "Construction"}, {242, 32, 214, 98, "Power"},
	{16, 137, 208, 98, "Chemistry"}, {242, 137, 214, 98, "Sociology"},
	{16, 243, 208, 98, "Computers"}, {242, 243, 214, 98, "Biology"},
	{16, 348, 208, 98, "Physics"}, {242, 348, 214, 98, "Force Fields"},
}

// currentAreaTopic 回傳某研究領域「目前應研究的主題」:MOO2 原版機制是玩家選定領域、
// 該領域依 techtree 固定順序逐一解鎖(非玩家自由挑選領域內個別主題,完成一項後才跳下一項,
// 期間若有多科技可選走 researchChoiceScreen 另外決定),故此處回傳該領域第一個尚未完成的
// 主題 + 其 RP 成本。Hyper-Advanced 是可重複研究的終端主題，因此整條完成後仍回最後一項，
// done=false；done 只保留給異常空領域。
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
		if gamedata.IsHyperAdvancedTopic(last) {
			return last, session.ResearchCostForDisplay(last), false
		}
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
	// Science_Room_ / _Tech_Select_ → STREAMHD #17,**播完接隨機 STREAM 1..3**
	// (Play_Streaming_Music_ 的 edx = −2 哨兵)。接的那一步在 tickBGM。
	playSceneBGMOnce(trackScienceRoom)
	// 8 個研究領域為點擊熱區(bg 局部座標;涵蓋整塊面板)→ 設定該領域目前主題 → 回星系。
	hits := researchAreaHits
	onAction := func(a string) *origTransition {
		if idx, ok := researchAreaOrder[a]; ok && b.session != nil {
			if t, _, done := currentAreaTopic(b.session, idx); !done {
				b.session.SetResearchTopic(t)
				if _, _, pending := b.session.PendingResearchChoice(); pending {
					if sc, err := b.researchChoice(b.galaxy); err == nil {
						return &origTransition{next: sc}
					}
				}
			}
		}
		return b.goTo(b.galaxy, researchAreaText(b.lang, "research.area.transition.galaxy"))
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
	s, err := loadOverlayScreen(b.res, "techsel.lbx", 0, b.lang, b.fnt, "tech.json",
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
			label := researchAreaText(b.lang, "research.area.topic_cost", topicNameZh(b.lang, t), cost)
			if gamedata.IsHyperAdvancedTopic(t) {
				level := b.session.Player.HyperAdvancedLevels[t]
				if level == 0 && b.session.Player.CompletedTopics[t] {
					level = 1 // 舊存檔尚未經下一次研究結算遷移
				}
				label = researchAreaText(b.lang, "research.area.hyper_level_cost",
					topicNameZh(b.lang, t), level+1, cost)
			}
			col := body
			if done {
				label, col = researchAreaText(b.lang, "research.area.complete"), gold
			}
			s.extras = append(s.extras,
				centeredExtraTextInSafeRect(researchAreaTopicTextRect(h), 12, label, col))
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
	vis := b.session.StarChartVisible()
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
	s, err := loadOverlayScreen(b.res, "plntsum.lbx", 0, b.lang, b.fnt, "planets.json",
		planetsOverlays, color.RGBA{206, 218, 240, 255}, 14, hits, onAction, nil)
	if err != nil {
		return nil, err
	}
	// 即時行星資料填進表格列(欄位中心 x 對齊標題;列中心 y 見 planetListRowY)。
	if b.session != nil {
		body := color.RGBA{206, 218, 240, 255}
		for i := 0; i < planetListRows && top+i < len(list); i++ {
			pi := list[top+i]
			p := b.session.Planets[pi]
			climate, gravity, minerals, size := planetEnvironmentLabels(b.lang, p)
			col := body
			if pi == b.planetPick {
				col = color.RGBA{250, 230, 140, 255} // 選中的那一列換色(這個畫面沒有可畫的選取框)
			}
			for column, text := range []string{p.Name, climate, gravity, minerals, size} {
				s.extras = append(s.extras, planetListColumnRect(i, column, false).centeredExtras(b.fnt, text, 12, col)...)
			}
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
				if n := systemBodyCountLabel(b.lang, b.session, star); n != "" {
					line += " " + n
				}
				// 欄寬 78(標題列 {18,14,78,18}),超出會往左溢到欄框外——截掉。
				s.extras = append(s.extras, planetListColumnRect(i, 0, true).centeredExtras(b.fnt, line, 9, sub)...)
			}
			// 這顆行星目前的狀態(自己的殖民地/前哨站),那是「還能不能派船過去」的關鍵資訊。
			if ci := b.session.ColonyIndexOnPlanet(pi); ci >= 0 {
				s.extras = append(s.extras, planetListColumnRect(i, 4, true).centeredExtras(b.fnt,
					uiText(b.lang, "planet.list.status.colony"), 9, color.RGBA{150, 225, 165, 255})...)
			} else if b.session.HasOutpostOnPlanet(pi) {
				s.extras = append(s.extras, planetListColumnRect(i, 4, true).centeredExtras(b.fnt,
					uiText(b.lang, "planet.list.status.outpost"), 9, color.RGBA{150, 195, 235, 255})...)
			}
			if sp := planetSpecialLabel(b.lang, p.SpecialID); sp != "" {
				s.extras = append(s.extras, planetListColumnRect(i, 3, true).centeredExtras(b.fnt, "★"+sp, 9, sub)...)
			}
		}
		if len(list) == 0 {
			s.extras = append(s.extras, planetListEmptyTextRect().centeredExtras(b.fnt,
				uiText(b.lang, "planet.list.empty"), 12, body)...)
		}
		// 動作結果訊息:壓在兩顆動作鈕上方那條空白帶。
		if b.planetListMsg != "" {
			s.extras = append(s.extras, planetListMessageTextRect().centeredExtras(b.fnt,
				b.planetListMsg, 9, color.RGBA{240, 215, 150, 255})...)
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
		return uiText(b.lang, "planet.action.pick_first")
	}
	star := sess.PlanetStar(b.planetPick)
	if star < 0 {
		return uiText(b.lang, "planet.action.no_system")
	}
	colony := action == "sendcolony"
	if colony && !sess.FleetHasColonyShip() {
		return uiText(b.lang, "galaxy.colonization.refusal.no_colony_ship")
	}
	if !colony && !sess.FleetHasOutpostShip() {
		return uiText(b.lang, "galaxy.outpost.refusal.no_outpost_ship")
	}
	if sess.Fleet().AtStar != star || sess.Fleet().ETA != 0 {
		if sess.Fleet().ETA > 0 {
			return fmt.Sprintf(uiText(b.lang, "galaxy.star_panel.status.transit"), sess.Fleet().ETA)
		}
		if !sess.SendFleet(star) {
			return uiText(b.lang, "planet.action.unreachable")
		}
		return fmt.Sprintf(uiText(b.lang, "planet.action.en_route"),
			sess.Stars[star].Name, sess.Fleet().ETA)
	}
	if colony {
		res := sess.ColonizePlanet(b.planetPick)
		if !res.Ok {
			return colonizationRefusalText(b.lang, res)
		}
		return fmt.Sprintf(uiText(b.lang, "planet.action.colonized"),
			sess.Planets[b.planetPick].Name, res.StartPopulation, res.PopMax)
	}
	res := sess.BuildOutpostOnPlanet(b.planetPick)
	if !res.Ok {
		return outpostRefusalText(b.lang, res)
	}
	return fmt.Sprintf(uiText(b.lang, "planet.action.outpost_established"),
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

	// promoSteps 是實機推廣錄影專用的「真實時間」導覽。不能重用 script：script
	// 每個 Update 都前進一格，在 Xvfb 的實際 TPS 低於 60 時會把幾秒停留拉成數十秒。
	promoSteps       []promoDemoStep
	promoStepIndex   int
	promoStepAt      time.Time
	promoStepStarted bool
	// promoCursor 是導覽預覽可選的可見游標。它沿著下一個正常 UI 點擊平滑移動；正式錄影
	// 可用 -promo-hide-cursor 關閉，避免準星壓在 CJK 文字或按鈕上造成錯誤的跑版印象。
	// 遊戲端仍只走原本的 hitRegion／戰術點擊處理，沒有切入畫廊專用展示狀態。
	promoCursorVisible                  bool
	promoCursorReady                    bool
	promoCursorX, promoCursorY          float64
	promoCursorFromX, promoCursorFromY  float64
	promoCursorToX, promoCursorToY      float64
	promoCursorMoveAt, promoCursorUntil time.Time
	promoClickUntil                     time.Time
	promoCursorHiddenUntil              time.Time
	// promoCompletionLogged 防止流程末端每個 Update 都重複寫 log。錄製腳本以這一行
	// 作為停止擷取的 checkpoint；只有戰鬥結果已實際寫回 session 才算完成。
	promoCompletionLogged bool

	// hi-res 畫布(第 86 項(hi-res 畫布)):off 是 640×480 的離屏,rec 收集這一幀的文字繪製。
	// uiScale==1 時兩者都不建立,整條路徑不走。
	off *ebiten.Image
	rec *uifont.Recorder

	audio *moo2audio.Mixer // 持有音訊 Mixer,避免 player 被 GC(headless 為 nil)

	// 過場截圖廊(-gamegallery):script 為導覽腳本,galleryShots 指定在哪個絕對 tick
	// 存哪張圖(可多張,依序達成)。與單張 shotPath 模式互斥。
	galleryDir   string
	galleryShots []galleryShot
	galleryDone  int
	// galleryEventTick 是截圖廊顯示固定事件戰報的 tick；只驗證版面，不注入玩法狀態。
	galleryEventTick int
	// galleryVictoryTick 是截圖廊專用:在這個 tick 把對局設成「已分出勝負」,好讓導覽腳本
	// 走得到最終得分畫面。那條路徑靠正常遊玩要好幾百回合,
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
	// galleryConfirmTick 是截圖廊把畫面換成是/否確認框的 tick。
	galleryConfirmTick int
	// galleryFighterTick 是截圖廊在戰術戰鬥裡派出一隊戰機的 tick。
	galleryFighterTick int
	// galleryNetWaitTick 是截圖廊把畫面換成網路等待畫面的 tick。
	galleryNetWaitTick int
	// galleryNetRosterTick 是截圖廊把畫面換成連線玩家名冊的 tick。
	galleryNetRosterTick int
	// galleryNetInfoTick 是截圖廊把畫面換成連線狀態面板的 tick。
	galleryNetInfoTick int
	// galleryNetGamesTick 是截圖廊把畫面換成區網對局清單的 tick。
	galleryNetGamesTick int
	// galleryInputBoxTick 是截圖廊把畫面換成文字輸入彈窗的 tick。
	galleryInputBoxTick int
	// galleryResearchTick 是截圖廊把畫面換成研究領域畫面的 tick。
	galleryResearchTick int
	galleryBuilder      *sceneBuilder
	gallerySession      *shell.GameSession
}

// promoDemoStep 的 hold 是送出 input 之後、下一個 input 前要保留的實際牆鐘時間。
// 這讓錄影節奏不依賴 Xvfb／軟體 OpenGL 的即時 TPS。
type promoDemoStep struct {
	input        shell.InputState
	hold         time.Duration
	cursorHidden time.Duration // 轉場後游標會穿過文字時，先隱藏至安全位置
}

func (a *interactiveApp) updatePromoCursor(now time.Time) {
	if !a.promoCursorReady {
		a.promoCursorReady = true
		a.promoCursorX, a.promoCursorY = moo2ScreenW/2, moo2ScreenH/2
		a.promoCursorFromX, a.promoCursorFromY = a.promoCursorX, a.promoCursorY
		a.promoCursorToX, a.promoCursorToY = a.promoCursorX, a.promoCursorY
		return
	}
	span := a.promoCursorUntil.Sub(a.promoCursorMoveAt)
	if span <= 0 || !now.Before(a.promoCursorUntil) {
		a.promoCursorX, a.promoCursorY = a.promoCursorToX, a.promoCursorToY
		return
	}
	p := float64(now.Sub(a.promoCursorMoveAt)) / float64(span)
	if p < 0 {
		p = 0
	}
	// smoothstep：停在按鈕上的片刻比線性等速更接近人手移動，也能讓 hover 邊框可見。
	p = p * p * (3 - 2*p)
	a.promoCursorX = a.promoCursorFromX + (a.promoCursorToX-a.promoCursorFromX)*p
	a.promoCursorY = a.promoCursorFromY + (a.promoCursorToY-a.promoCursorFromY)*p
}

func (a *interactiveApp) planPromoCursor(now, until time.Time) {
	for i := a.promoStepIndex; i < len(a.promoSteps); i++ {
		in := a.promoSteps[i].input
		if !in.ClickReleased {
			continue
		}
		a.updatePromoCursor(now)
		a.promoCursorFromX, a.promoCursorFromY = a.promoCursorX, a.promoCursorY
		a.promoCursorToX, a.promoCursorToY = float64(in.MouseX), float64(in.MouseY)
		a.promoCursorMoveAt, a.promoCursorUntil = now, until
		return
	}
}

func (a *interactiveApp) drawPromoCursor(dst *ebiten.Image) {
	if !a.promoCursorVisible || a.promoSteps == nil || !a.promoCursorReady || time.Now().Before(a.promoCursorHiddenUntil) {
		return
	}
	scale := float32(uiScale)
	x, y := float32(a.promoCursorX)*scale, float32(a.promoCursorY)*scale
	ink := color.RGBA{250, 242, 190, 255}
	shadow := color.RGBA{4, 8, 16, 220}
	vector.DrawFilledCircle(dst, x, y, 3*scale, shadow, true)
	vector.StrokeCircle(dst, x, y, 7*scale, 1.5*scale, ink, true)
	vector.StrokeLine(dst, x-11*scale, y, x+11*scale, y, 1*scale, ink, true)
	vector.StrokeLine(dst, x, y-11*scale, x, y+11*scale, 1*scale, ink, true)
	if time.Now().Before(a.promoClickUntil) {
		vector.StrokeCircle(dst, x, y, 13*scale, 2*scale, color.RGBA{255, 190, 90, 230}, true)
	}
}

// galleryVictoryTick 是截圖廊在哪個 tick 把對局設成「已分出勝負」——必須早於腳本裡
// 「按 TURN 進最終得分」那一拍(t29),取它的前一拍。
const galleryVictoryTick = 38

// 原版一般事件至少要到第 50 回合；短畫廊在此顯示外部文案組成的固定戰報，
// 僅驗證事件畫面，不宣稱這是第一回合的玩法結果。
const galleryEventTick = 16

// galleryFleetTick 是截圖廊在哪個 tick 給艦隊注入結構損傷 + 次元傳送門——取「進艦隊列表」
// 那一拍(t19)的前一拍。
//
// ⚠ 這一拍必須晚於腳本裡最後一次「結束回合」(t17 之前那次),否則 EndTurn 的
// advanceShipRepair 會把傷清光:艦隊開局就停在母星,照原版 Repair_Ships_At_Colonies_
// 的規則會被**完全修復**。先前注入在 t28 而 t29 按了結束回合,截出來一艘傷都沒有——
// 那不是顯示壞了,是修復規則正常運作。
const galleryFleetTick = 18

// galleryNetWaitTick 是截圖廊在哪個 tick 換成網路等待畫面——取截圖(t96)的前一拍。
const galleryNetWaitTick = 95

// galleryNetRosterTick 是截圖廊在哪個 tick 換成連線玩家名冊——取截圖(t98)的前一拍。
const galleryNetRosterTick = 97

// galleryNetInfoTick 是截圖廊在哪個 tick 換成連線狀態面板。
//
// ⚠ 這一拍**不能與前一張的截圖同一拍**:換畫面的處理跑在截圖之前,設成 98 會把
// 名冊那張(t98)換掉——截出來的 31_netroster 變成狀態面板。第一次就是這樣寫錯的,
// 截圖廊的逐位元組比對當場抓到。所以是 99,截圖留到 t102 讓面板動畫推幾幀。
const galleryNetInfoTick = 99

// galleryNetGamesTick 是截圖廊在哪個 tick 換成區網對局清單——取截圖(t104)的前一拍。
const galleryNetGamesTick = 103

// galleryInputBoxTick 是截圖廊在哪個 tick 疊上文字輸入彈窗——取截圖(t106)的前一拍。
//
// 這一張是**疊在對局清單上**的(modal 要看得見下層,同確認框),所以刻意接在
// galleryNetGamesTick 之後而不是自己推一張底。
const galleryInputBoxTick = 105

// galleryResearchTick 是截圖廊在哪個 tick 換成研究領域畫面——取截圖(t108)的前一拍。
// 此注入只驗證正式 renderer 與真實 session 資料；研究選題的正常玩家 gate 另由 t12~t13 驗證。
const galleryResearchTick = 107

// galleryFighterTick 是截圖廊在哪個 tick 於戰術戰鬥裡派出一隊戰機——取截圖(t66)的前一拍。
//
// ⚠ 這一拍會**給第一艘我方艦裝上戰機庫**。開局那三艘船(拓荒號/先驅一二號)沒有戰機庫,
// 而戰機庫要靠艦艇設計 + 生產才拿得到,截圖廊跑不到那一步。同 galleryFleetTick 注入
// 結構損傷的立場:為了讓那一層**被看見**而擺出狀態,不是改規則。
const galleryFighterTick = 65

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

// galleryConfirmTick 是截圖廊在哪個 tick 換成是/否確認框——取截圖(t94)的前一拍。
//
// 走正常路徑要「艦隊列表按 RELOCATE → 點自己的殖民地 → 點一顆被怪獸盤據的星」,
// 那要求截圖廊那一局剛好有怪獸而且看得見;直接推上來可靠得多(同建造視窗/指揮點數的處理)。
const galleryConfirmTick = 93

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
		// 人類是清單第 2 項(一星多行星缺口) → 第 0 欄第 5 列 → x 351..474、y 330..375。
		// (先前這裡點的是 (540,451) 的「接受」鈕,2026-08-07 版面改成原版的 2×7 網格後
		//  那個座標什麼都不會命中,腳本會卡在種族畫面、後面每一張都截錯。)
		click(410, 350), // t7: 種族選擇「人類」→ 命名/旗色
		idle,            // t8: settle → 截圖 nameflag
		click(540, 454), // t9: 命名/旗色「接受」→ 星系主畫面
		idle,            // t10: settle
		idle,            // t11: settle → 截圖 galaxy

		// 第一個正常世界回合只會進回合摘要；原版一般事件至少要到第 50 回合，
		// 因此事件版面另由 galleryEventTick 的固定戰報驗證。
		// 新對局尚未選研究 application；第一次按 TURN 會先進入 researchChoice，
		// 不會結算世界。舊腳本漏掉這個正常玩家 gate，導致 t14..t74 全拍成同一張研究畫面。
		click(589, 458), // t12: 第一次「結束回合」→ 選擇研究 application
		click(320, 173), // t13: 選第一項 application → 回星系
		click(589, 458), // t14: 再按「結束回合」→ 回合摘要並截圖
		click(320, 393), // t15: 回合摘要「關閉」→ 星系主畫面
		idle,            // t16: galleryEventTick 顯示固定事件戰報並截圖
		click(320, 384), // t17: 事件快報「繼續」→ 回合摘要
		click(320, 393), // t18: 回合摘要「關閉」→ 星系主畫面

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
		// 截圖驗證等不起。
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

		// 遊戲選單視窗(原版 Do_Main_Game_Popup_ @ 0x7DD41)。從星系主畫面點頂端「遊戲」鈕進得去,
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

		// 是/否確認框(原版 Confirmation_Box_)。同上,直接推上來。
		idle, // t93: 由 galleryConfirmTick 換成確認框
		idle, // t94: settle → 截圖 confirm

		// 網路等待畫面(原版 Net_Next_Turn)。同上,直接推上來。
		idle, // t95: 由 galleryNetWaitTick 換成等待畫面
		idle, // t96: settle → 截圖 netwait

		// 連線玩家名冊(原版 Choose_Network_Plyrs_Screen_)。同上,直接推上來。
		idle, // t97: 由 galleryNetRosterTick 換成名冊畫面
		idle, // t98: settle → 截圖 netroster

		// 連線狀態面板(原版 Generic_Net_Info / Join_Net / SendGet_Net_Info 共用的那一張)。
		idle, // t99:  由 galleryNetInfoTick 換成狀態面板
		idle, // t100: 推動畫
		idle, // t101: 推動畫
		idle, // t102: settle → 截圖 netinfo

		// 區網對局清單(原版 Choose_Multi_Network_Game_Screen_)。同上,直接推上來。
		idle, // t103: 由 galleryNetGamesTick 換成對局清單
		idle, // t104: settle → 截圖 netgames

		// 文字輸入彈窗(原版 Remapped_Input_Box_Popup_),疊在對局清單上。
		idle, // t105: 由 galleryInputBoxTick 疊上輸入彈窗
		idle, // t106: settle → 截圖 inputbox

		idle, // t107: 由 galleryResearchTick 換成研究領域畫面
		idle, // t108: settle → 截圖 research
	}
	shots := []galleryShot{
		{1, "01_menu.png"},
		// NEW GAME 設定畫面先前從沒被截圖廊拍過(腳本從主選單直接點到種族選擇),
		// 所以那個畫面的版面錯了也不會被發現。名字用 01b 而不是重編後面 23 張的號。
		{3, "01b_newgame.png"},
		{6, "02_raceselect.png"},
		{8, "03_nameflag.png"},
		{11, "04_galaxy.png"},
		{14, "06_turnsummary.png"},
		{16, "05_event.png"},
		{21, "07_fleet.png"},
		{24, "08_antaranroom.png"},
		{31, "09_colonysummary.png"},
		{34, "10_colonyscreen.png"},
		{41, "11_hiscore.png"},
		{48, "12_planets.png"},
		{53, "13_info.png"},
		{55, "14_info_tech.png"},
		// RACES 原本正常路徑有走到，卻在下一拍就進外交而沒有截圖；間諜列與 Agent
		// 版面因此長期不在逐位元驗收範圍。t59 是進畫面後的 settle 幀。
		{59, "15a_races.png"},
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
		{94, "29_confirm.png"},
		{96, "30_netwait.png"},
		{98, "31_netroster.png"},
		{102, "32_netinfo.png"},
		{104, "33_netgames.png"},
		{106, "34_inputbox.png"},
		{108, "35_research.png"},
	}
	return script, shots
}

// buildPromoDemoSteps 是給實機推廣錄影用的可重播導覽。它只使用正常玩家介面可達的
// 點擊：主選單、新局、種族、星圖、殖民地人口調配、RACES 間諜、外交與戰術。不同於 buildGalleryScript，
// 它不建立示範存檔、不推入地面戰／多人等展示畫面，也不輸出 PNG。
//
// 每段停留時間以牆鐘時間計算，避免不同 Docker/Xvfb 更新率改變導覽節奏。
// interactiveApp 仍是正常的 ebiten.Game，因此錄影捕捉的是即時渲染、轉場與實際 UI
// 狀態，而不是投影片或預先輸出的圖檔。
func buildPromoDemoSteps() []promoDemoStep {
	idle := shell.InputState{}
	click := func(x, y int) shell.InputState {
		return shell.InputState{MouseX: x, MouseY: y, ClickReleased: true}
	}
	var steps []promoDemoStep
	hold := func(seconds int) {
		steps = append(steps, promoDemoStep{input: idle, hold: time.Duration(seconds) * time.Second})
	}
	action := func(x, y, secondsAfter int) {
		steps = append(steps, promoDemoStep{input: click(x, y), hold: time.Duration(secondsAfter) * time.Second})
	}
	actionWithCursorHide := func(x, y, secondsAfter int, hide time.Duration) {
		steps = append(steps, promoDemoStep{
			input: click(x, y), hold: time.Duration(secondsAfter) * time.Second, cursorHidden: hide,
		})
	}

	// 主選單 → 新局 → 種族 → 命名 → 星圖。每一段都預留給操作結果顯示，
	// 而非讓單一畫面長時間像投影片般停住。
	hold(2)             // 等 Xvfb 錄影器附著後再點 NEW GAME，避免片頭被截斷。
	action(491, 228, 2) // NEW GAME
	// 設定接受與開始遊戲都點在按鈕右側留白，導覽游標不會蓋住置中的中文標籤。
	// 從設定頁的 ACCEPT 直接移往人類鈕會斜穿自訂種族；轉場後先隱藏游標，
	// 直到它抵達人類鈕右下留白，避免把游標線誤看成文字偏移。
	actionWithCursorHide(526, 398, 2, 1500*time.Millisecond) // NEW GAME ACCEPT
	// 人類按鈕右下角的留白；後續移往「開始遊戲」時不會斜穿右欄的崔拉里安文字。
	action(464, 370, 2) // 人類種族
	action(578, 448, 3) // 命名／旗色 ACCEPT

	// 殖民地不只停留在畫面：實際把一名農夫調成工人，再返回星圖。
	action(48, 452, 2)  // COLONIES
	action(50, 47, 3)   // 第一個殖民地
	action(410, 107, 2) // 工人欄：農夫 → 工人
	action(590, 459, 1) // 殖民地 RETURN
	action(608, 462, 2) // 殖民地總覽 RETURN

	// 正常種族關係 → 實際調整間諜 → 外交對談 → 宣戰。這三次操作使影片
	// 呈現的不是靜態 RACES 畫面，而是可見的數值／任務狀態變化。
	action(420, 452, 2) // RACES
	action(130, 135, 2) // 第一列：增派間諜（避開置中文字）
	action(204, 135, 2) // 第一列：循環任務
	action(276, 135, 2) // 第一列：設為隱匿
	action(483, 428, 3) // REPORT
	action(320, 437, 1) // 結束對談
	action(388, 448, 2) // DECLARE WAR

	// 戰術段落實際移動、開火、撤離與結算；每次都經由戰術棋盤的 hitRegion／距離／
	// 射程／命中／特效與戰後回寫路徑，不是狀態注入或預先輸出的畫格。
	action(215, 97, 3) // 拓荒號前進一格
	action(495, 97, 7) // 對第一艘敵艦開火，保留命中／特效時間
	// 開局小艦隊的先驅艦在本回合沒有移動力，敵方回合會在拓荒號失去掩護時立即還擊。
	// 導覽因此展示玩家可隨時使用的撤離，而非假裝能以 AUTO 無限推進；下一步仍經過
	// tacticalScreen 的 ApplyCombatOutcome 寫回 session，確保片尾是完整的正常玩家路徑。
	action(365, 401, 3) // RETREAT
	action(320, 245, 3) // 點擊套用戰果並進入戰鬥結果
	action(190, 334, 9) // CLOSE：返回 RACES，片尾不停在負面戰報
	return steps
}

// reportPromoCompletion 在所有導覽操作與最後一段停留均結束後寫出單一 checkpoint。
// LastBattle 只能由 tacticalScreen 在戰鬥結束、玩家點擊離開後透過 ApplyCombatOutcome 建立；
// 因此它證明錄影沒有在 AUTO 無射程、戰術中途或轉場途中被固定秒數截斷。
func promoDemoResultReached(b *sceneBuilder) bool {
	return b != nil && b.session != nil && b.session.LastBattle != nil
}

func (a *interactiveApp) reportPromoCompletion(now time.Time) {
	if a.promoSteps == nil || a.promoCompletionLogged || a.promoStepIndex < len(a.promoSteps) || now.Before(a.promoStepAt) {
		return
	}
	a.promoCompletionLogged = true
	if !promoDemoResultReached(a.b) {
		fmt.Fprintln(os.Stderr, "promo-demo: failed: tactical battle did not reach result")
		return
	}
	fmt.Fprintln(os.Stderr, "promo-demo: complete")
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
	// 視窗尺寸以**邏輯**尺寸為單位(+/- 縮放的體感不變);內部畫布是 uiScale 倍,
	// 所以視窗至少要有 uiScale 倍大,否則 1280×960 的內容縮進小視窗會**比放大前更糊**
	// (縮小取樣),等於白做。
	w, h := moo2ScreenW*s, moo2ScreenH*s
	if cw, ch := canvasSize(); w < cw {
		w, h = cw, ch
	}
	ebiten.SetWindowSize(w, h)
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
	if a.promoSteps != nil {
		now := time.Now()
		if !a.promoStepStarted {
			a.promoStepStarted = true
			a.promoStepAt = now
		}
		a.updatePromoCursor(now)
		if a.promoStepIndex >= len(a.promoSteps) {
			a.reportPromoCompletion(now)
			return shell.InputState{}
		}
		if now.Before(a.promoStepAt) {
			// 導覽游標只負責可視化，不在移動途中餵給畫面 hover 判定。尤其種族選擇
			// 是「滑過即預覽」，若把途中座標當真實滑鼠輸入，游標經過別顆族鈕時會把
			// 已選定的人類文字換成別族的高亮，畫面看起來就像文字或按鈕跑位。
			// 真正的 click 仍在下一個 step 經正常 hitRegion 傳入。
			return shell.InputState{}
		}
		step := a.promoSteps[a.promoStepIndex]
		a.promoStepIndex++
		a.promoStepAt = now.Add(step.hold)
		if step.cursorHidden > 0 {
			a.promoCursorHiddenUntil = now.Add(step.cursorHidden)
		}
		if !step.input.ClickReleased {
			a.planPromoCursor(now, a.promoStepAt)
			return shell.InputState{}
		}
		a.promoCursorX, a.promoCursorY = float64(step.input.MouseX), float64(step.input.MouseY)
		a.promoClickUntil = now.Add(180 * time.Millisecond)
		a.planPromoCursor(now, a.promoStepAt)
		return step.input
	}
	// ⚠ `CursorPosition` 回的是 **Layout 空間**(hi-res 時是 1280×960),而所有命中區都是
	// 640×480 邏輯座標——除不回去的話滑鼠會偏一倍。這是 rulebook/81 明列的踩雷之一。
	x, y := ebiten.CursorPosition()
	x, y = int(float64(x)/uiScale), int(float64(y)/uiScale)
	return shell.InputState{
		MouseX: x, MouseY: y,
		ClickReleased:      inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft),
		RightClickReleased: inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight),
		MouseDown:          ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft),
		AnyKeyPressed:      len(inpututil.AppendJustPressedKeys(nil)) > 0,
		Hotkey:             pollHotkey(),
	}
}

func (a *interactiveApp) Update() error {
	a.tick++
	// 單次播放的曲子(科學室)播完要接隨機 STREAM 1..3——原版 Play_Streaming_Music_ 的
	// −2 哨兵。放在最前面:它與畫面狀態無關,而且 headless 時是 no-op。
	tickBGM()
	if a.b != nil {
		a.b.animTick = a.tick // 動畫計數(黑洞旋渦等),見 starsprite.go
	}
	if a.script == nil && a.promoSteps == nil { // 互動模式才處理視窗快捷鍵(headless／推廣導覽略過)
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
	// 截圖廊專用：一般事件最早第 50 回合才可能發生，短畫廊只用固定雙語戰報
	// 驗證事件視窗版面；不呼叫事件規則，也不修改對局經濟或殖民地狀態。
	if a.galleryEventTick > 0 && a.tick == a.galleryEventTick && a.galleryBuilder != nil && a.gallerySession != nil {
		a.gallerySession.LastEventReport = &shell.EventReport{
			EventID:   0,
			Good:      true,
			Name:      uiText(i18n.Traditional, "gallery.event.name"),
			NameEN:    uiText(i18n.English, "gallery.event.name"),
			Message:   uiText(i18n.Traditional, "gallery.event.message"),
			MessageEN: uiText(i18n.English, "gallery.event.message"),
		}
		if sc, err := a.galleryBuilder.eventScreen(); err == nil {
			a.cur = sc
		}
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
			ColonyName:           uiText(a.galleryBuilder.lang, "gallery.colony.demo"),
			AttackerMarinesStart: 6, AttackerTanksStart: 2, DefenderStart: 5,
			AttackerSurvived:        5,
			AttackerMarinesSurvived: 3, AttackerTanksSurvived: 2, DefenderSurvived: 0,
			Rounds: 4, AttackerColor: 6, DefenderColor: 3,
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
			ColonyName:         uiText(a.galleryBuilder.lang, "gallery.colony.demo"),
			DefenderColor:      3,
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
	// 截圖廊專用:網路等待畫面(原版 Net_Next_Turn)。
	if a.galleryNetWaitTick > 0 && a.tick == a.galleryNetWaitTick && a.galleryBuilder != nil {
		a.cur = a.galleryBuilder.netNextTurnDemo()
	}
	// 截圖廊專用:連線玩家名冊(原版 Choose_Network_Plyrs_Screen_)。
	if a.galleryNetRosterTick > 0 && a.tick == a.galleryNetRosterTick && a.galleryBuilder != nil {
		a.cur = a.galleryBuilder.chooseNetPlayersDemo()
	}
	// 截圖廊專用:連線狀態面板(原版 Generic_Net_Info 家族)。
	if a.galleryNetInfoTick > 0 && a.tick == a.galleryNetInfoTick && a.galleryBuilder != nil {
		a.cur = a.galleryBuilder.netInfoDemo()
	}
	// 截圖廊專用:區網對局清單(原版 Choose_Multi_Network_Game_Screen_)。
	if a.galleryNetGamesTick > 0 && a.tick == a.galleryNetGamesTick && a.galleryBuilder != nil {
		a.cur = a.galleryBuilder.chooseMultiNetGameDemo()
	}
	// 截圖廊專用:文字輸入彈窗(原版 Remapped_Input_Box_Popup_),疊在目前畫面上。
	if a.galleryInputBoxTick > 0 && a.tick == a.galleryInputBoxTick && a.galleryBuilder != nil {
		ib := a.galleryBuilder.inputBox(a.cur,
			"inputbox.title.host_address",
			"192.168.1.20:24501", 45, nil)
		ib.scriptOK = true // 截圖廊不吃鍵盤,否則腳本的按鍵會被它收走
		a.cur = ib
	}
	// 截圖廊專用:研究領域畫面。使用同一場對局的 session，只省略重新導覽回星圖的步驟。
	if a.galleryResearchTick > 0 && a.tick == a.galleryResearchTick && a.galleryBuilder != nil {
		if sc, err := a.galleryBuilder.research(); err == nil {
			a.cur = sc
		}
	}
	// 截圖廊專用:戰術戰鬥裡派一隊戰機出擊(見 galleryFighterTick 的說明)。
	if a.galleryFighterTick > 0 && a.tick == a.galleryFighterTick {
		if ts, ok := a.cur.(*tacticalScreen); ok && len(ts.player) > 1 {
			// 用第 2 艘(偵察艦)而不是第 1 艘:第 1 艘是殖民船,讓殖民船派戰機出擊
			// 在截圖裡看起來像個 bug。
			const carrier = 1
			ts.player[carrier].Bay, ts.player[carrier].BayKind = true, shell.FighterInterceptor
			ts.sel = carrier
			ts.launchFrom(carrier)
			ts.advanceSquadrons() // 推一回合,好讓中隊離開母艦格、看得出它在飛
		}
	}
	// 截圖廊專用:是/否確認框,疊在星圖上(原版就是這樣疊的)。
	if a.galleryConfirmTick > 0 && a.tick == a.galleryConfirmTick && a.galleryBuilder != nil {
		b := a.galleryBuilder
		b.measure.on = false // 測距的提示線會蓋在框上,先關掉
		under, _ := b.galaxy()
		a.cur = b.confirm(under, b.galleryConfirmMessage(), nil, nil)
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
	in := a.pollInput()
	if a.b != nil && a.b.continuousTurns {
		if in.ClickReleased || in.RightClickReleased {
			// 原版用全畫面輸入欄中止連續回合；同一個 click 不再傳給星圖熱區。
			a.b.stopContinuousTurns()
			in = shell.InputState{}
		} else if a.tick >= a.b.continuousTurnAt {
			a.b.continuousTurnAt = a.tick + continuousTurnInterval
			if t := a.b.advanceWorldTurn(); t != nil {
				if t.quit {
					return ebiten.Termination
				}
				if t.next != nil {
					a.cur = t.next
				}
			}
			return nil
		}
	}
	if t := a.cur.update(in); t != nil {
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
			// ⚠ 這裡要走 drawScene 而不是 cur.draw:**截圖是 remake 唯一的驗收管道**,
			// 繞過合成的話畫廊拍到的永遠是 640 的舊畫面,hi-res 改了也看不出來
			// ——這正是第 86 項(hi-res 畫布)第一次以為「Layout 沒生效」的真正原因。
			off := ebiten.NewImage(canvasSize())
			a.drawScene(off)
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
	a.drawScene(dst)
	// 推廣導覽的游標必須在 hi-res 畫布完成後才畫，避免被離屏縮放或文字錄製流程蓋掉。
	a.drawPromoCursor(dst)
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

// drawScene 把當前畫面畫到 dst。
//
// uiScale==1:直接畫(與 hi-res 畫布之前完全相同的路徑)。
// uiScale>1:畫面進 640 離屏、文字改用錄製,離屏 nearest 放大之後把文字以 2× 重播
// ——**美術是銳利的整數倍放大,文字是在最終解析度重新柵格化的**(見 uifont/record.go)。
func (a *interactiveApp) drawScene(dst *ebiten.Image) {
	if uiScale <= 1 {
		a.cur.draw(dst)
		return
	}
	if a.off == nil {
		a.off = ebiten.NewImage(moo2ScreenW, moo2ScreenH)
		a.rec = &uifont.Recorder{}
	}
	a.off.Clear()
	a.rec.Reset()
	// flush:把離屏這一段美術貼上去 → 重播這一段的文字 → **清空離屏**,讓下一段只帶新畫的東西。
	// 清空是關鍵:離屏背景不透明,重貼整張會把剛畫好的字洗掉(見 uifont/record.go 的屏障說明)。
	flush := func() {
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterNearest} // pixel art 不能雙線性
		op.GeoM.Scale(uiScale, uiScale)
		dst.DrawImage(a.off, op)
		a.rec.Replay(dst, uiScale)
		a.rec.Reset()
		a.off.Clear()
	}
	uifont.SetFlushHook(flush)
	uifont.StartRecording(a.rec)
	a.cur.draw(a.off)
	uifont.StopRecording()
	uifont.SetFlushHook(nil)
	flush()
}

// canvasSize 是內部畫布(= Layout 空間、也是截圖的實際尺寸)。
func canvasSize() (int, int) { return int(moo2ScreenW * uiScale), int(moo2ScreenH * uiScale) }

func (a *interactiveApp) Layout(int, int) (int, int) { return canvasSize() }

// runInteractive 啟動「還原原版」的互動遊戲。script/shot 非空時為 headless 驗證;
// galleryDir 非空時為「端到端過場截圖廊」模式(見 buildGalleryScript),優先於 script/shot。
func runInteractive(versionAssets versionAssetDirs, initial gamedata.GameVersion, lang i18n.Lang, fnt, fntVec *uifont.Font,
	script []shell.InputState, shot string, frames int, galleryDir string, noAudio, promoDemo, promoHideCursor bool) error {

	if lang == i18n.Traditional && fnt == nil {
		return fmt.Errorf("中文模式需以 -font 指定 CJK 字型")
	}
	dirs, ok := versionAssets.forVersion(initial)
	if !ok {
		return fmt.Errorf("未提供初始版本 %s 的遊戲資料", versionShort(initial))
	}
	res, err := assets.NewResolver(dirs...)
	if err != nil {
		return err
	}
	b := &sceneBuilder{res: res, versionAssets: versionAssets, fnt: fnt, fntVec: fntVec, lang: lang, session: shell.NewDemoSession(), newGameSize: 1, newGameDiff: newGameDiffDefault,
		newGameAge: newGameAgeDefault, newGameTech: newGameTechDefault, newGameEmpires: 1 + shell.DefaultOpponents, designWeapon: 1, designAmmo: 5, savePath: savePathFor(), gameVersion: initial,
		planetPick: -1} // −1 = 行星列表還沒選任何一列(0 是行星 0 的索引,不能當「沒選」)
	b.skipCutscenes = shot != "" || galleryDir != "" || promoDemo // 見該欄位註解
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
	var promoSteps []promoDemoStep
	if promoDemo {
		promoSteps = buildPromoDemoSteps()
	}
	if galleryDir != "" {
		if err := os.MkdirAll(galleryDir, 0o755); err != nil {
			return fmt.Errorf("建立過場截圖目錄 %q: %w", galleryDir, err)
		}
		script, shots = buildGalleryScript()
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
	app := &interactiveApp{cur: start, script: script, promoSteps: promoSteps, promoCursorVisible: promoDemo && !promoHideCursor, shotPath: shot, frames: frames, scale: scale,
		galleryDir: galleryDir, galleryShots: shots, b: b}
	if galleryDir != "" {
		app.gallerySession = b.session
		app.galleryEventTick = galleryEventTick
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
		app.galleryConfirmTick = galleryConfirmTick
		app.galleryFighterTick = galleryFighterTick
		app.galleryNetWaitTick = galleryNetWaitTick
		app.galleryNetRosterTick = galleryNetRosterTick
		app.galleryNetInfoTick = galleryNetInfoTick
		app.galleryNetGamesTick = galleryNetGamesTick
		app.galleryInputBoxTick = galleryInputBoxTick
		app.galleryResearchTick = galleryResearchTick
		app.galleryBuilder = b
	}
	// 只有真正互動(非 headless 截圖/腳本/截圖廊)才啟用音訊:headless 環境常無音效卡,
	// 且截圖驗證不需要聲音。-noaudio 讓實機畫面錄製能在沒有 ALSA 裝置的 Docker
	// 裡執行；它只停用 runtime mixer，不改變遊戲規則或素材解碼。
	if shot == "" && script == nil && !noAudio {
		app.audio = initAudio(res)
	}
	winW, winH := moo2ScreenW*scale, moo2ScreenH*scale
	if cw, ch := canvasSize(); winW < cw {
		winW, winH = cw, ch
	}
	ebiten.SetWindowSize(winW, winH)
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

// designModChipRect 回傳艦艇設計畫面第 i 個武器改造晶片的位置與可用寬度。
//
// **熱區與繪製共用這一份**——先前是兩份寫死的座標表,那是會漂移的重複。
//
// 版面:2 欄 × 4 列。欄數不是 4 而是 2,因為**英文標籤比中文長得多**
// (`No Range Dissipation (NR)` vs `無射程衰減(NR)`),4 欄放不下會疊字。
// 2 欄各 163px,在 size 10 下容得下最長的那個。
func designModChipRect(i int) struct{ x, y, w float64 } {
	const (
		x0      = 305.0 // 與同一面板其他文字的左緣一致
		right   = 632.0 // 畫布右緣留 8px
		rows    = 4
		rowStep = 16.0
		// 面板下邊框量在 y=431。四列使用 16px 字形高度與列距，位置為
		// 366/382/398/414，末列止於 430；標題止於 366，兩者只共用邊界。
		// 先前 4 欄 × 22px 兩列的版面在英文下會疊字,改 2 欄 4 列之後行高才收得緊。
		y0 = 366.0
	)
	colW := (right - x0) / 2
	return struct{ x, y, w float64 }{
		x: x0 + float64(i/rows)*colW,
		y: y0 + float64(i%rows)*rowStep,
		w: colW,
	}
}

// designMountSlotRect 與 designMountControlRect 是多槽 UI 的單一幾何來源；繪製、文字安全框與
// hit region 都必須呼叫它們，避免再次出現按鈕文字與點擊區中心漂移。
func designMountSlotRect(i int) [4]int {
	if i < 0 {
		i = 0
	}
	if i > 7 {
		i = 7
	}
	return [4]int{305 + i*30, 180, 28, 18}
}

func designMountControlRect(action string) [4]int {
	switch action {
	case "mountadd":
		return [4]int{548, 180, 40, 18}
	case "mountdel":
		return [4]int{590, 180, 40, 18}
	case "mountdec":
		return [4]int{548, 201, 40, 18}
	case "mountinc":
		return [4]int{590, 201, 40, 18}
	default:
		return [4]int{}
	}
}

func designSpecialControlRect(action string) [4]int {
	switch action {
	case "specialprev":
		return [4]int{520, 130, 25, 18}
	case "specialnext":
		return [4]int{547, 130, 25, 18}
	case "specialadd":
		return [4]int{574, 130, 25, 18}
	case "specialdel":
		return [4]int{601, 130, 25, 18}
	default:
		return [4]int{}
	}
}
