package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"sort"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/assets"
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
	offsetX, offsetY int                     // 背景圖在 640×480 畫布上的置中偏移(小於全螢幕的視窗畫面用)
	eraseColor       *color.RGBA             // 非 nil 時強制用此色擦底(背景均勻的畫面用,勝過採樣猜測)
	extras           []extraText             // 即時動態文字(星曆、國庫…),疊在背景+overlay 之上
	postDraw         func(dst *ebiten.Image) // 任意額外繪製(如星圖),在最後呼叫
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
			// 擦掉烘進圖的英文:填單色底(eraseColor 指定則用之,否則取標籤帶的「中位數色」——
			// 代表性中間調,避免誤取過暗陰影形成黑框;單色填充不複製紋理,故不會有錯位歪斜)。
			plate := samplePlate(s.rgba, b)
			if s.eraseColor != nil {
				plate = *s.eraseColor
			}
			vector.DrawFilledRect(dst, float32(float64(b.x+3)+ox), float32(float64(b.y+2)+oy),
				float32(b.w-6), float32(b.h-4), plate, false)
			// 疊中文(同 overlay.go)。
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
func samplePlate(rgba *image.RGBA, b labelRect) color.RGBA {
	W, H := rgba.Bounds().Dx(), rgba.Bounds().Dy()
	var cols []color.RGBA
	add := func(x, y int) {
		if x < 0 || x >= W || y < 0 || y >= H {
			return
		}
		i := rgba.PixOffset(x, y)
		cols = append(cols, color.RGBA{rgba.Pix[i], rgba.Pix[i+1], rgba.Pix[i+2], 255})
	}
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
	res     *assets.Resolver
	fnt     *uifont.Font
	lang    i18n.Lang
	session *shell.GameSession // 活的對局狀態(TURN 推進、畫面顯示即時資料)
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
	// 星圖各星加點擊熱區(點星 → 顯示該星系行星資訊)。
	if b.session != nil {
		for i, st := range b.session.Stars {
			sx, sy := starScreenPos(st)
			hits = append(hits, hitRegion{sx - 11, sy - 11, 22, 22, fmt.Sprintf("star%d", i)})
		}
	}
	onAction := func(a string) *origTransition {
		if len(a) > 4 && a[:4] == "star" && b.session != nil {
			if idx, err := strconv.Atoi(a[4:]); err == nil {
				b.session.SelectedStar = idx
				return b.goTo(b.galaxy, "星系主畫面") // 重繪顯示選中星資訊
			}
		}
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
		case "turn":
			// 核心迴圈:結算一回合(玩家帝國 + 各 AI 對手決策),再顯示回合摘要(原版流程)。
			b.session.EndTurn()
			return b.goTo(b.turnSummary, "回合摘要")
		}
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
	s, err := loadOverlayScreen(b.res, "buffer0.lbx", 0, b.lang, b.fnt, "assets/i18n/menu.tsv",
		overlays, color.RGBA{210, 216, 230, 255}, 12, hits, onAction, nil)
	if err != nil {
		return nil, err
	}
	// 星圖(中央視窗,openorion2 StarmapWidget 20,20,507,401)+ 即時狀態文字(疊在星圖之上)。
	if b.session != nil {
		sess := b.session
		fnt := s.font
		s.postDraw = func(dst *ebiten.Image) {
			drawStarmap(dst, fnt, sess.Stars, sess.SelectedStar)
			if fnt != nil {
				year := 3500 + (sess.Turn - 1)
				fnt.Draw(dst, fmt.Sprintf("星曆 %d", year), 30, 40, 16, color.RGBA{240, 220, 120, 255})
				fnt.Draw(dst, fmt.Sprintf("國庫 %d BC", sess.Player.BC), 30, 62, 13, color.RGBA{210, 216, 230, 255})
				fnt.Draw(dst, fmt.Sprintf("研究:%s", shell.ResearchTopicName(sess.Player.ResearchTopic)), 30, 82, 13, color.RGBA{160, 210, 240, 255})
				// 選中星:顯示該星系行星資訊(左下角面板)。
				if sess.SelectedStar >= 0 && sess.SelectedStar < len(sess.Planets) {
					p := sess.Planets[sess.SelectedStar]
					vector.DrawFilledRect(dst, 28, 330, 210, 82, color.RGBA{10, 14, 30, 230}, false)
					vector.StrokeRect(dst, 28, 330, 210, 82, 1, color.RGBA{90, 130, 200, 255}, false)
					fnt.Draw(dst, p.Name, 38, 352, 15, color.RGBA{240, 220, 120, 255})
					fnt.Draw(dst, fmt.Sprintf("氣候 %s ／ 大小 %s", p.Climate, p.Size), 38, 376, 12, color.RGBA{210, 216, 230, 255})
					fnt.Draw(dst, fmt.Sprintf("重力 %s ／ 礦產 %s", p.Gravity, p.Mineral), 38, 398, 12, color.RGBA{210, 216, 230, 255})
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
func drawStarmap(dst *ebiten.Image, fnt *uifont.Font, stars []shell.Star, selected int) {
	const vx0, vy0, vx1, vy1 = starVX0, starVY0, starVX1, starVY1
	vector.DrawFilledRect(dst, vx0, vy0, vx1-vx0, vy1-vy0, color.RGBA{6, 6, 16, 255}, false)
	for i, st := range stars {
		x := float32(vx0) + float32(st.X)*(vx1-vx0)
		y := float32(vy0) + float32(st.Y)*(vy1-vy0)
		col, ok := spectralColors[uint8(st.Spectral)]
		if !ok {
			col = color.RGBA{200, 200, 200, 255}
		}
		r := float32(6 - st.Size) // 大=6 .. 小=3
		if r < 3 {
			r = 3
		}
		// 選中星:黃色高亮環。
		if i == selected {
			vector.StrokeCircle(dst, x, y, r+6, 2, color.RGBA{255, 240, 120, 255}, true)
		}
		// 擁有環:我方藍綠、敵方紅。
		switch st.Owner {
		case 1:
			vector.StrokeCircle(dst, x, y, r+3, 1.5, color.RGBA{90, 230, 180, 255}, true)
		case 2:
			vector.StrokeCircle(dst, x, y, r+3, 1.5, color.RGBA{235, 90, 80, 255}, true)
		}
		vector.DrawFilledCircle(dst, x, y, r, col, true)
		if fnt != nil && st.Name != "" {
			fnt.Draw(dst, st.Name, float64(x)+float64(r)+3, float64(y)-2, 11, color.RGBA{170, 185, 210, 255})
		}
	}
}

// colonySummary 建原版殖民地總覽畫面(COLSUM.LBX 資產 0,自帶完整調色盤)。
// openorion2 未實作此 view,背景資產由本專案自 LBX 探測定位。
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
				hitRegion{104, top, 118, 30, fmt.Sprintf("f%d", i)},
				hitRegion{236, top, 128, 30, fmt.Sprintf("w%d", i)},
				hitRegion{376, top, 128, 30, fmt.Sprintf("s%d", i)},
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
				extraText{x: colX.name, y: y, size: 13, text: fmt.Sprintf("殖民地 %d", i+1), col: body, align: 1},
				extraText{x: colX.far, y: y, size: 13, text: fmt.Sprintf("%d", c.Farmers), col: body, align: 1},
				extraText{x: colX.wrk, y: y, size: 13, text: fmt.Sprintf("%d", c.Workers), col: body, align: 1},
				extraText{x: colX.sci, y: y, size: 13, text: fmt.Sprintf("%d", c.Scientists), col: body, align: 1},
			)
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
		{0, 0, moo2ScreenW, moo2ScreenH, "back"},
	}
	onAction := func(a string) *origTransition {
		switch a {
		case "audience":
			return b.goTo(b.council, "銀河議會")
		case "declarewar":
			if b.session != nil {
				b.session.ResolveBattle("賽隆人")
			}
			return b.goTo(b.battleResult, "戰鬥結果")
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
	return loadOverlayScreen(b.res, "races.lbx", 0, b.lang, b.fnt, "assets/i18n/diplo.tsv",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction, nil)
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
		outcome, oc := "✗ 敗北", lose
		if bt.PlayerWon {
			outcome, oc = "★ 勝利!", win
		}
		s.extras = []extraText{
			{x: 40, y: 60, size: 15, text: fmt.Sprintf("對「%s」開戰", bt.Enemy), col: gold},
			{x: 40, y: 92, size: 16, text: outcome, col: oc},
			{x: 40, y: 122, size: 13, text: fmt.Sprintf("我方戰力 %d ／ 敵方戰力 %d", bt.PlayerStrength, bt.EnemyStrength), col: body},
			{x: 40, y: 146, size: 13, text: fmt.Sprintf("損失:我方 %d 艘", bt.PlayerLosses), col: body},
		}
	}
	return s, nil
}

// council 建原版銀河議會畫面(COUNCIL.LBX 資產 1,調色盤鏈 COUNCIL#0)。3D 議事廳,
// 無烘字,疊「銀河議會」標題;點畫面返回種族關係。
func (b *sceneBuilder) council() (*overlayScreen, error) {
	hits, onAction := b.backHit(b.races, "種族關係")
	s, err := loadOverlayScreen(b.res, "council.lbx", 1, b.lang, b.fnt, "assets/i18n/misc.tsv",
		nil, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"council.lbx", 0}})
	if err != nil {
		return nil, err
	}
	if b.fnt != nil {
		s.extras = []extraText{
			{x: moo2ScreenW / 2, y: 30, size: 22, text: "銀河議會", col: color.RGBA{240, 220, 120, 255}, align: 1},
		}
	}
	return s, nil
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
		{86, 78, 130, 22, "DIFFICULTY", 0},
		{232, 78, 150, 22, "GALAXY SIZE", 0},
		{398, 78, 150, 22, "GALAXY AGE", 0},
		{86, 222, 130, 22, "PLAYERS", 0},
		{232, 222, 150, 22, "TECH LEVEL", 0},
		{422, 266, 138, 18, "TACTICAL COMBAT", 11},
		{422, 301, 138, 18, "RANDOM EVENTS", 11},
		{422, 334, 138, 18, "ANTARANS ATTACK", 11},
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
	// 點右側艦艇格 → 艦艇設計;右下 RETURN → 星系主畫面(精確熱區)。
	hits := []hitRegion{
		{338, 50, 288, 300, "design"},
		{543, 432, 84, 28, "return"},
	}
	onAction := func(a string) *origTransition {
		if a == "design" {
			return b.goTo(b.shipDesign, "艦艇設計")
		}
		return b.goTo(b.galaxy, "星系主畫面")
	}
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
	s, err := loadOverlayScreen(b.res, "fleet.lbx", 0, b.lang, b.fnt, "assets/i18n/menu.tsv",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"buffer0.lbx", 0}, {"fleet.lbx", 111}})
	if err != nil {
		return nil, err
	}
	// 艦隊名冊填進左下暗面板(艦名 + 艦體等級)。
	if b.session != nil {
		gold := color.RGBA{240, 220, 120, 255}
		body := color.RGBA{206, 214, 232, 255}
		y := 312.0
		for _, sh := range b.session.Ships {
			s.extras = append(s.extras,
				extraText{x: 28, y: y, size: 13, text: sh.Name, col: gold},
				extraText{x: 130, y: y, size: 12, text: sh.Class, col: body},
			)
			y += 28
		}
	}
	return s, nil
}

// shipDesign 建原版艦艇設計畫面(DESIGN.LBX 資產 0,調色盤鏈 buffer0#0)。
// 點艦體等級 → 建造該艦加入艦隊 → 回艦隊;點他處 → 返回艦隊。
func (b *sceneBuilder) shipDesign() (*overlayScreen, error) {
	hullZH := map[string]string{
		"Frigate": "巡防艦", "Destroyer": "驅逐艦", "Cruiser": "巡洋艦",
		"Battleship": "戰艦", "Titan": "泰坦", "Doom Star": "末日之星",
	}
	hits := []hitRegion{
		{125, 50, 118, 16, "Frigate"}, {125, 67, 118, 16, "Destroyer"},
		{125, 84, 118, 16, "Cruiser"}, {125, 101, 118, 16, "Battleship"},
		{125, 118, 118, 16, "Titan"}, {125, 135, 118, 16, "Doom Star"},
		{0, 0, moo2ScreenW, moo2ScreenH, "back"},
	}
	onAction := func(a string) *origTransition {
		if zh, ok := hullZH[a]; ok && b.session != nil {
			b.session.BuildShip(zh) // 造艦加入艦隊
		}
		return b.goTo(b.fleet, "艦隊列表")
	}
	overlays := []labelRect{
		{255, 12, 320, 24, "Ship Design", 0},
		{130, 52, 105, 16, "Frigate", 12},
		{130, 69, 105, 16, "Destroyer", 12},
		{130, 86, 105, 16, "Cruiser", 12},
		{130, 103, 105, 16, "Battleship", 12},
		{130, 120, 105, 16, "Titan", 12},
		{130, 137, 105, 16, "Doom Star", 12},
		{380, 440, 80, 20, "Clear", 0},
		{470, 440, 80, 20, "Cancel", 0},
		{558, 440, 72, 20, "Build", 0},
	}
	return loadOverlayScreen(b.res, "design.lbx", 0, b.lang, b.fnt, "assets/i18n/tech.tsv",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"buffer0.lbx", 0}})
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
	s, err := loadOverlayScreen(b.res, "officer.lbx", 0, b.lang, b.fnt, "assets/i18n/officer.tsv",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"buffer0.lbx", 0}})
	if err != nil {
		return nil, err
	}
	// 領袖名單填進左側槽位(肖像右方名字區;槽中心 y 經 PIL 量測:31/144/253/362 分隔)。
	if b.session != nil {
		gold := color.RGBA{240, 220, 120, 255}
		body := color.RGBA{206, 214, 232, 255}
		rowY := []float64{87, 198, 307, 415}
		for i, ld := range b.session.Leaders {
			if i >= len(rowY) {
				break
			}
			y := rowY[i]
			s.extras = append(s.extras,
				extraText{x: 95, y: y - 12, size: 15, text: ld.Name, col: gold},
				extraText{x: 95, y: y + 12, size: 12, text: fmt.Sprintf("%s ｜ Lv %d", ld.Skill, ld.Level), col: body},
			)
		}
	}
	return s, nil
}

// info 建原版科技總覽畫面(INFO.LBX 資產 0,基底 INFO.LBX 資產 1)。座標經 PIL 量測
// (screens-scan/info_overview.png):左側選單五列 y=57/79/105/134/154,標題 y=16,RETURN y=436。
func (b *sceneBuilder) info() (*overlayScreen, error) {
	// 「科技總覽」列 → 研究選擇畫面;RETURN 或他處 → 星系主畫面。
	hits := []hitRegion{
		{15, 78, 197, 22, "tech"},
		{0, 0, moo2ScreenW, moo2ScreenH, "back"},
	}
	onAction := func(a string) *origTransition {
		if a == "tech" {
			return b.goTo(b.research, "研究選擇")
		}
		return b.goTo(b.galaxy, "星系主畫面")
	}
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
			{x: 40, y: 62, size: 15, text: fmt.Sprintf("星曆 %d 結算", year), col: gold},
			{x: 40, y: 92, size: 13, text: fmt.Sprintf("淨工業 %d ／ 研究 %d", out.TotalNetIndustry, out.TotalResearch), col: body},
			{x: 40, y: 116, size: 13, text: fmt.Sprintf("食物盈餘 %d ／ 稅收 %d BC", out.TotalFood, out.TaxRevenue), col: body},
			{x: 40, y: 140, size: 13, text: fmt.Sprintf("國庫 %d BC(本回合 %+d)", b.session.Player.BC, out.NetBC), col: body},
		}
		if out.ResearchDone {
			s.extras = append(s.extras, extraText{x: 40, y: 168, size: 14, text: "★ 完成一項研究!", col: color.RGBA{120, 220, 140, 255}})
		}
	}
	return s, nil
}

// research 建原版研究選擇畫面(TECHSEL.LBX 資產 0,無內嵌調色盤 → 走調色盤鏈,
// 基底取自 SCIENCE.LBX 資產 0)。點畫面任一處返回主選單。
func (b *sceneBuilder) research() (*overlayScreen, error) {
	// 8 個研究領域為點擊熱區(bg 局部座標;涵蓋整塊面板)→ 設定該領域代表研究主題 → 回星系。
	areaTopic := map[string]gamedata.ResearchTopic{
		"Construction": gamedata.TOPIC_ADVANCED_CONSTRUCTION,
		"Power":        gamedata.TOPIC_ADVANCED_FUSION,
		"Chemistry":    gamedata.TOPIC_ADVANCED_CHEMISTRY,
		"Sociology":    gamedata.TOPIC_ADVANCED_GOVERNMENTS,
		"Computers":    gamedata.TOPIC_ARTIFICIAL_INTELLIGENCE,
		"Biology":      gamedata.TOPIC_ADVANCED_BIOLOGY,
		"Physics":      gamedata.TOPIC_ADVANCED_MAGNETISM,
		"Force Fields": gamedata.TOPIC_ADVANCED_ENGINEERING,
	}
	hits := []hitRegion{
		{16, 32, 208, 98, "Construction"}, {242, 32, 214, 98, "Power"},
		{16, 137, 208, 98, "Chemistry"}, {242, 137, 214, 98, "Sociology"},
		{16, 243, 208, 98, "Computers"}, {242, 243, 214, 98, "Biology"},
		{16, 348, 208, 98, "Physics"}, {242, 348, 214, 98, "Force Fields"},
	}
	onAction := func(a string) *origTransition {
		if t, ok := areaTopic[a]; ok && b.session != nil {
			b.session.SetResearchTopic(t) // 實際設定研究主題,結束回合朝此累積
		}
		return b.goTo(b.galaxy, "星系主畫面")
	}
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
	s, err := loadOverlayScreen(b.res, "plntsum.lbx", 0, b.lang, b.fnt, "assets/i18n/planets.tsv",
		planetsOverlays, color.RGBA{206, 218, 240, 255}, 14, hits, onAction, nil)
	if err != nil {
		return nil, err
	}
	// 即時行星資料填進表格列(欄位中心 x 對齊標題;列中心 y 經量測估計)。
	if b.session != nil {
		body := color.RGBA{206, 218, 240, 255}
		rowY := []float64{61, 116, 170, 225, 280, 335, 390} // 格中心,PIL 量測(格線 34/89/143/198/253/308/363/418)
		cx := struct{ name, cli, grv, min, siz float64 }{57, 136, 218, 303, 382}
		for i, p := range b.session.Planets {
			if i >= len(rowY) {
				break
			}
			y := rowY[i]
			s.extras = append(s.extras,
				extraText{x: cx.name, y: y, size: 12, text: p.Name, col: body, align: 1},
				extraText{x: cx.cli, y: y, size: 12, text: p.Climate, col: body, align: 1},
				extraText{x: cx.grv, y: y, size: 12, text: p.Gravity, col: body, align: 1},
				extraText{x: cx.min, y: y, size: 12, text: p.Mineral, col: body, align: 1},
				extraText{x: cx.siz, y: y, size: 12, text: p.Size, col: body, align: 1},
			)
		}
	}
	return s, nil
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
	b := &sceneBuilder{res: res, fnt: fnt, lang: lang, session: shell.NewDemoSession()}
	menu, err := b.menu()
	if err != nil {
		return err
	}
	app := &interactiveApp{cur: menu, script: script, shotPath: shot, frames: frames}
	ebiten.SetWindowSize(moo2ScreenW, moo2ScreenH)
	ebiten.SetWindowTitle("Master of Orion II — 繁體中文化 (remake)")
	return ebiten.RunGame(app)
}
