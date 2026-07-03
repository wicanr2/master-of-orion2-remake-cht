package main

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
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
	bg               *ebiten.Image
	rgba             *image.RGBA
	font             *uifont.Font
	cat              *i18n.Catalog
	overlays         []labelRect
	labelColor       color.RGBA
	defSize          float64
	hits             []hitRegion
	onAction         func(action string) *origTransition
	hover            string
	offsetX, offsetY int         // 背景圖在 640×480 畫布上的置中偏移(小於全螢幕的視窗畫面用)
	eraseColor       *color.RGBA // 非 nil 時強制用此色擦底(背景均勻的畫面用,勝過採樣猜測)
}

func (s *overlayScreen) update(in shell.InputState) *origTransition {
	// 命中判定在背景圖局部座標(扣掉置中偏移)。
	mx, my := in.MouseX-s.offsetX, in.MouseY-s.offsetY
	s.hover = ""
	for _, h := range s.hits {
		if h.hit(mx, my) {
			s.hover = h.action
			break
		}
	}
	if in.ClickReleased {
		for _, h := range s.hits {
			if h.hit(mx, my) && s.onAction != nil {
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
			plate := samplePlate(s.rgba, b)
			if s.eraseColor != nil {
				plate = *s.eraseColor
			}
			// 擦掉烘進圖的英文(蓋底色),再疊中文(同 overlay.go)。
			vector.DrawFilledRect(dst, float32(float64(b.x+3)+ox), float32(float64(b.y+2)+oy),
				float32(b.w-6), float32(b.h-4), plate, false)
			size := b.size
			if size == 0 {
				size = s.defSize
			}
			zh := s.cat.Translate(b.enKey)
			s.font.DrawCentered(dst, zh, float64(b.x)+float64(b.w)/2+ox, float64(b.y)+float64(b.h)/2+oy, size, s.labelColor)
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
}

// samplePlate 取標籤底板色(用來擦掉烘進圖的英文)。
// 策略:合併兩組採樣取「眾數色」——(1) 標籤左內緣窄直帶(置中文字通常不及此),
// (2) 標籤上緣/下緣橫向帶(避開文字所在的垂直中段)。單靠(1)在按鈕左緣恰是光澤高光帶
// 時會誤取亮色;疊(2)後真正占多數的底板色勝出。
// 註:對「背景均勻但文字靠左/寬粗填滿」的畫面(採樣仍可能誤判),改用 overlayScreen.eraseColor
// 強制指定底色(如 info 面板)。與字底極性無關。
func samplePlate(rgba *image.RGBA, b labelRect) color.RGBA {
	W, H := rgba.Bounds().Dx(), rgba.Bounds().Dy()
	counts := map[color.RGBA]int{}
	best := color.RGBA{0, 0, 0, 255}
	bestN := 0
	add := func(x, y int) {
		if x < 0 || x >= W || y < 0 || y >= H {
			return
		}
		i := rgba.PixOffset(x, y)
		c := color.RGBA{rgba.Pix[i], rgba.Pix[i+1], rgba.Pix[i+2], 255}
		counts[c]++
		if counts[c] > bestN {
			bestN = counts[c]
			best = c
		}
	}
	for _, dx := range []int{3, 5, 7, 9} {
		for y := b.y + 1; y < b.y+b.h-1; y++ {
			add(b.x+dx, y)
		}
	}
	for _, y := range []int{b.y + 1, b.y + b.h - 2} {
		for x := b.x + 4; x < b.x+b.w-4; x += 2 {
			add(x, y)
		}
	}
	return best
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
	res  *assets.Resolver
	fnt  *uifont.Font
	lang i18n.Lang
}

// menu 建原版主選單畫面。按鈕熱區用 menuOverlays 的座標(按鈕即標籤)。
func (b *sceneBuilder) menu() (*overlayScreen, error) {
	hits := make([]hitRegion, 0, len(menuOverlays))
	for _, o := range menuOverlays {
		hits = append(hits, hitRegion{o.x, o.y, o.w, o.h, o.enKey})
	}
	onAction := func(a string) *origTransition {
		switch a {
		case "Quit Game":
			return &origTransition{quit: true}
		case "New Game":
			// 新遊戲:先進原版 NEW GAME 設定畫面(難度/星系/玩家…),ACCEPT 後進星系主畫面。
			return b.goTo(b.newGameSetup, "新遊戲設定")
		case "Continue":
			// 續玩:直接進星系主畫面(後續接讀檔)。
			return b.goTo(b.galaxy, "星系主畫面")
		case "Hall of Fame":
			// 暫借「名人堂」入口示範調色盤鏈解鎖的研究選擇畫面(原本無內嵌調色盤)。
			return b.goTo(b.research, "研究選擇")
		}
		// Load Game / Multi Player:尚未實作,暫不動作。
		return nil
	}
	return loadOverlayScreen(b.res, "mainmenu.lbx", 21, b.lang, b.fnt, "assets/i18n/menu.tsv",
		menuOverlays, color.RGBA{104, 224, 96, 255}, 15, hits, onAction, nil)
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
	hits := []hitRegion{
		{15, 430, 67, 44, "colonies"},
		{90, 430, 67, 44, "planets"},
		{165, 430, 67, 44, "fleets"},
		{310, 430, 70, 44, "leaders"},
		{385, 430, 70, 44, "races"},
		{460, 430, 70, 44, "info"},
		{544, 441, 90, 34, "turn"},
	}
	onAction := func(a string) *origTransition {
		switch a {
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
		}
		// turn:尚未接入,暫不動作。
		return nil
	}
	// 工具列標籤擦底疊字(x 為按鈕中心對齊,y 中心經 PIL 量測:一般列 450、TURN 455)。
	overlays := []labelRect{
		{13, 443, 71, 14, "Colonies", 12},
		{88, 443, 71, 14, "Planets", 12},
		{163, 443, 71, 14, "Fleets", 12},
		{308, 443, 74, 14, "Leaders", 12},
		{383, 443, 74, 14, "Races", 12},
		{458, 443, 74, 14, "Info", 12},
		{544, 448, 90, 15, "Turn", 12},
	}
	return loadOverlayScreen(b.res, "buffer0.lbx", 0, b.lang, b.fnt, "assets/i18n/menu.tsv",
		overlays, color.RGBA{210, 216, 230, 255}, 12, hits, onAction, nil)
}

// colonySummary 建原版殖民地總覽畫面(COLSUM.LBX 資產 0,自帶完整調色盤)。
// openorion2 未實作此 view,背景資產由本專案自 LBX 探測定位。
func (b *sceneBuilder) colonySummary() (*overlayScreen, error) {
	hits, onAction := b.backHit(b.galaxy, "星系主畫面")
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
	return loadOverlayScreen(b.res, "colsum.lbx", 0, b.lang, b.fnt, "assets/i18n/colony.tsv",
		overlays, color.RGBA{210, 216, 230, 255}, 13, hits, onAction, nil)
}

// races 建原版種族關係畫面(RACES.LBX 資產 0,自帶完整調色盤)。RACES 按鈕目標。
func (b *sceneBuilder) races() (*overlayScreen, error) {
	hits, onAction := b.backHit(b.galaxy, "星系主畫面")
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
	return loadOverlayScreen(b.res, "races.lbx", 0, b.lang, b.fnt, "assets/i18n/diplo.tsv",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction, nil)
}

// newGameSetup 建原版新遊戲設定畫面(NEWGAME.LBX 資產 28,調色盤鏈 RACEOPT#4→NEWGAME#1)。
// ACCEPT 進星系主畫面;CANCEL 回主選單。
func (b *sceneBuilder) newGameSetup() (*overlayScreen, error) {
	hits := []hitRegion{
		{92, 392, 108, 30, "cancel"},
		{432, 392, 108, 30, "accept"},
	}
	onAction := func(a string) *origTransition {
		if a == "accept" {
			return b.goTo(b.galaxy, "星系主畫面")
		}
		return b.goTo(b.menu, "主選單")
	}
	// 座標經 PIL 量測(remain-scan/newgame_a28_f00.png);開關標籤移到核取框右側(x430)避免採到藍框。
	overlays := []labelRect{
		{244, 44, 166, 24, "New Game", 0},
		{86, 78, 130, 22, "DIFFICULTY", 0},
		{232, 78, 150, 22, "GALAXY SIZE", 0},
		{398, 78, 150, 22, "GALAXY AGE", 0},
		{86, 222, 130, 22, "PLAYERS", 0},
		{232, 222, 150, 22, "TECH LEVEL", 0},
		{426, 266, 134, 18, "TACTICAL COMBAT", 11},
		{426, 301, 134, 18, "RANDOM EVENTS", 11},
		{426, 334, 134, 18, "ANTARANS ATTACK", 11},
		{100, 388, 96, 24, "CANCEL", 0},
		{440, 388, 96, 24, "ACCEPT", 0},
	}
	return loadOverlayScreen(b.res, "newgame.lbx", 28, b.lang, b.fnt, "assets/i18n/menu.tsv",
		overlays, color.RGBA{210, 216, 230, 255}, 13, hits, onAction,
		paletteChain{{"raceopt.lbx", 4}, {"newgame.lbx", 1}})
}

// fleet 建原版艦隊列表畫面(FLEET.LBX 資產 0,三段調色盤鏈)。座標經 PIL 量測
// (screens-scan/fleetlist.png):標題列 y=27,兩排按鈕列 y=394/443。
func (b *sceneBuilder) fleet() (*overlayScreen, error) {
	hits, onAction := b.backHit(b.galaxy, "星系主畫面")
	overlays := []labelRect{
		{190, 17, 260, 20, "FLEET OPERATIONS", 0},
		{346, 384, 70, 18, "ALL", 0},
		{440, 384, 93, 18, "RELOCATE", 0},
		{552, 384, 64, 18, "SCRAP", 0},
		{342, 436, 76, 18, "LEADERS", 0},
		{425, 436, 60, 18, "Support", 0},
		{482, 436, 62, 18, "Combat", 0},
		{543, 436, 82, 18, "RETURN", 0},
	}
	return loadOverlayScreen(b.res, "fleet.lbx", 0, b.lang, b.fnt, "assets/i18n/menu.tsv",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"buffer0.lbx", 0}, {"fleet.lbx", 111}})
}

// officer 建原版軍官列表畫面(OFFICER.LBX 資產 0)。座標經 PIL 量測
// (screens-scan/officer_leaderlist.png):頁籤列 y=12-32,按鈕列 y=440-462。
func (b *sceneBuilder) officer() (*overlayScreen, error) {
	hits, onAction := b.backHit(b.galaxy, "星系主畫面")
	overlays := []labelRect{
		{20, 11, 133, 20, "Colony Leaders", 0},
		{166, 11, 124, 20, "Ship Officers", 0},
		{310, 440, 68, 20, "HIRE", 0},
		{388, 440, 69, 20, "POOL", 0},
		{462, 440, 74, 20, "DISMISS", 0},
		{540, 440, 80, 20, "RETURN", 0},
	}
	return loadOverlayScreen(b.res, "officer.lbx", 0, b.lang, b.fnt, "assets/i18n/officer.tsv",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"buffer0.lbx", 0}})
}

// info 建原版科技總覽畫面(INFO.LBX 資產 0,基底 INFO.LBX 資產 1)。座標經 PIL 量測
// (screens-scan/info_overview.png):左側選單五列 y=57/79/105/134/154,標題 y=16,RETURN y=436。
func (b *sceneBuilder) info() (*overlayScreen, error) {
	hits, onAction := b.backHit(b.galaxy, "星系主畫面")
	// 選單項原版為靠左文字疊在近黑面板背景上(無實心板);擦底取黑=黑疊黑(正確),
	// rect 寬取足以蓋住最長英文、中文置中於偏左位置貼近原版。y 中心經 PIL 量測:64/88/114/142/162。
	overlays := []labelRect{
		{15, 20, 200, 26, "STAR DATE", 0},
		{15, 56, 182, 18, "History Graph", 0},
		{15, 80, 182, 18, "Tech Review", 0},
		{15, 106, 182, 18, "Race Statistics", 0},
		{15, 134, 182, 18, "Turn Summary", 0},
		{15, 154, 182, 18, "Reference", 0},
		{538, 436, 84, 22, "RETURN", 0},
	}
	s, err := loadOverlayScreen(b.res, "info.lbx", 0, b.lang, b.fnt, "assets/i18n/misc.tsv",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"info.lbx", 1}})
	if err != nil {
		return nil, err
	}
	// info 選單/標題都疊在均勻的近黑面板背景上,強制用該背景色擦底(採樣會因長英文誤取字色)。
	black := color.RGBA{0, 8, 24, 255}
	s.eraseColor = &black
	return s, nil
}

// research 建原版研究選擇畫面(TECHSEL.LBX 資產 0,無內嵌調色盤 → 走調色盤鏈,
// 基底取自 SCIENCE.LBX 資產 0)。點畫面任一處返回主選單。
func (b *sceneBuilder) research() (*overlayScreen, error) {
	hits, onAction := b.backHit(b.menu, "主選單")
	// 研究領域標籤擦底疊字(座標為 bg 局部座標,472×480;draw 時自動加置中偏移)。
	// 座標經 PIL 量測原版標籤中心(左右欄列中心 y=36/140/246/352,標題 18);h=18,y=中心−9。
	overlays := []labelRect{
		{155, 9, 162, 18, "Select New Research", 0},
		{22, 27, 128, 18, "Construction", 0},
		{244, 27, 124, 18, "Power", 0},
		{22, 131, 128, 18, "Chemistry", 0},
		{244, 131, 124, 18, "Sociology", 0},
		{22, 237, 128, 18, "Computers", 0},
		{244, 237, 124, 18, "Biology", 0},
		{22, 343, 128, 18, "Physics", 0},
		{244, 343, 124, 18, "Force Fields", 0},
	}
	return loadOverlayScreen(b.res, "techsel.lbx", 0, b.lang, b.fnt, "assets/i18n/tech.tsv",
		overlays, color.RGBA{210, 216, 230, 255}, 13, hits, onAction,
		paletteChain{{"science.lbx", 0}})
}

// planets 建原版行星列表畫面。「返回」按鈕熱區導回星系主畫面。
func (b *sceneBuilder) planets() (*overlayScreen, error) {
	hits := []hitRegion{{454, 440, 157, 23, "Return"}}
	onAction := func(a string) *origTransition {
		if a == "Return" {
			return b.goTo(b.galaxy, "星系主畫面")
		}
		return nil
	}
	return loadOverlayScreen(b.res, "plntsum.lbx", 0, b.lang, b.fnt, "assets/i18n/planets.tsv",
		planetsOverlays, color.RGBA{206, 218, 240, 255}, 14, hits, onAction, nil)
}

// --- interactiveApp(ebiten.Game;支援 headless 腳本驗證)---

type interactiveApp struct {
	cur origScreen

	// headless 驗證:script 逐幀注入輸入,跑滿 frames 存 shot。
	script   []shell.InputState
	shotPath string
	frames   int
	tick     int
	saved    bool
}

func (a *interactiveApp) pollInput() shell.InputState {
	if a.script != nil {
		if idx := a.tick - 1; idx >= 0 && idx < len(a.script) {
			return a.script[idx]
		}
		return shell.InputState{}
	}
	x, y := ebiten.CursorPosition()
	return shell.InputState{
		MouseX: x, MouseY: y,
		ClickReleased: inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft),
	}
}

func (a *interactiveApp) Update() error {
	a.tick++
	if t := a.cur.update(a.pollInput()); t != nil {
		if t.quit {
			return ebiten.Termination
		}
		if t.next != nil {
			a.cur = t.next
		}
	}
	if a.shotPath != "" && a.saved {
		return ebiten.Termination
	}
	return nil
}

func (a *interactiveApp) Draw(dst *ebiten.Image) {
	a.cur.draw(dst)
	if a.shotPath != "" && !a.saved && a.tick >= a.frames {
		if err := saveScreenshot(dst, a.shotPath); err != nil {
			fmt.Println("截圖失敗:", err)
		}
		a.saved = true
	}
}

func (a *interactiveApp) Layout(int, int) (int, int) { return moo2ScreenW, moo2ScreenH }

// runInteractive 啟動「還原原版」的互動遊戲。script/shot 非空時為 headless 驗證。
func runInteractive(dirs []string, lang i18n.Lang, fnt *uifont.Font,
	script []shell.InputState, shot string, frames int) error {

	if lang == i18n.Traditional && fnt == nil {
		return fmt.Errorf("中文模式需以 -font 指定 CJK 字型")
	}
	res, err := assets.NewResolver(dirs...)
	if err != nil {
		return err
	}
	b := &sceneBuilder{res: res, fnt: fnt, lang: lang}
	menu, err := b.menu()
	if err != nil {
		return err
	}
	app := &interactiveApp{cur: menu, script: script, shotPath: shot, frames: frames}
	ebiten.SetWindowSize(moo2ScreenW, moo2ScreenH)
	ebiten.SetWindowTitle("Master of Orion II — 繁體中文化 (remake)")
	return ebiten.RunGame(app)
}
