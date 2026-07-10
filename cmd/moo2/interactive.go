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
	res           *assets.Resolver
	fnt           *uifont.Font
	lang          i18n.Lang
	session       *shell.GameSession // 活的對局狀態(TURN 推進、畫面顯示即時資料)
	newGameSize   int                // NEW GAME 選的星系大小索引(shell.GalaxySizes)
	newGameDiff   int                // NEW GAME 選的難度索引(shell.Difficulties)
	newGameRace   int                // NEW GAME 選的種族索引(shell.Races)
	newGameSeed   int                // 每次新遊戲遞增,讓星系種子變化
	savePath      string             // remake 存檔路徑(每回合自動存;主選單 Load/Continue 讀)
	designWeapon  int                // 艦艇設計選的武器元件索引(shell.WeaponOptions)
	designArmor   int                // 裝甲元件索引(shell.ArmorOptions)
	designShield  int                // 護盾元件索引(shell.ShieldOptions)
	designSpecial int                // 特殊元件索引(shell.SpecialOptions)
	designMsg     string             // 艦艇設計畫面「空間不足,擋下建造」的提示訊息(切換元件/成功建造時清空)
	lastActionMsg string             // 星圖畫面「載運陸戰隊/發動地面入侵」的最近一次結果訊息(選新星時清空)
}

// savePathFor 回傳 remake 存檔路徑(使用者設定目錄下,退回暫存目錄),確保可寫。
func savePathFor() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		dir = os.TempDir()
	}
	sub := filepath.Join(dir, "moo2-remake-cht")
	if mkErr := os.MkdirAll(sub, 0o755); mkErr != nil {
		return filepath.Join(os.TempDir(), "moo2-remake-save.json")
	}
	return filepath.Join(sub, "save.json")
}

// menu 建原版主選單畫面。按鈕熱區用 menuOverlays 的座標(按鈕即標籤)。
func (b *sceneBuilder) menu() (*overlayScreen, error) {
	playSceneBGM(bgmMenu)
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
			// 續玩:若有存檔先讀回,否則沿用目前對局,進星系主畫面。
			if b.savePath != "" && shell.SaveExists(b.savePath) {
				if gs, err := shell.LoadSession(b.savePath); err == nil {
					b.session = gs
				} else {
					fmt.Fprintln(os.Stderr, "讀檔失敗:", err)
				}
			}
			return b.goTo(b.galaxy, "星系主畫面")
		case "Load Game":
			// 讀取存檔並進星系主畫面(無存檔則不動作)。
			if b.savePath != "" && shell.SaveExists(b.savePath) {
				if gs, err := shell.LoadSession(b.savePath); err == nil {
					b.session = gs
					return b.goTo(b.galaxy, "星系主畫面")
				} else {
					fmt.Fprintln(os.Stderr, "讀檔失敗:", err)
				}
			}
			return nil
		case "Hall of Fame":
			// 暫借「名人堂」入口示範調色盤鏈解鎖的研究選擇畫面(原本無內嵌調色盤)。
			return b.goTo(b.research, "研究選擇")
		}
		// Multi Player:尚未實作,暫不動作。
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
	playSceneBGM(bgmGalaxy)
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
		sess := b.session
		for i, st := range sess.Stars {
			sx, sy := starScreenPos(st)
			hits = append(hits, hitRegion{sx - 11, sy - 11, 22, 22, fmt.Sprintf("star%d", i)})
		}
		// 選中星資訊面板內的操作鈕(座標同 postDraw 繪製的按鈕框):三種互斥,依艦隊/選中星
		// 狀態擇一顯示——派遣艦隊(艦隊不在選中星)、載運陸戰隊(艦隊在玩家母星,唯一已知
		// 有 Marine Barracks 殖民地模型對映的星,見 shell.AIOpponent.ColonyStars 註解同款限制)、
		// 發動地面入侵(艦隊在敵方殖民地星且已載運陸戰隊)。
		if sess.SelectedStar >= 0 && sess.SelectedStar < len(sess.Stars) {
			switch {
			case sess.FleetETA > 0:
				// 航行中,面板只顯示狀態文字,無按鈕。
			case sess.SelectedStar == sess.FleetAtStar:
				switch {
				case sess.SelectedStar == 0:
					hits = append(hits, hitRegion{38, 402, 190, 20, "loadmarines"})
				case sess.Stars[sess.SelectedStar].Owner == 2 && sess.FleetMarines > 0:
					hits = append(hits, hitRegion{38, 402, 190, 20, "invade"})
				}
			default:
				hits = append(hits, hitRegion{38, 402, 190, 20, "dispatch"})
			}
		}
	}
	onAction := func(a string) *origTransition {
		if len(a) > 4 && a[:4] == "star" && b.session != nil {
			if idx, err := strconv.Atoi(a[4:]); err == nil {
				b.session.SelectedStar = idx
				b.lastActionMsg = ""             // 換選中星,清掉上一顆星的動作結果訊息
				return b.goTo(b.galaxy, "星系主畫面") // 重繪顯示選中星資訊
			}
		}
		if a == "dispatch" && b.session != nil {
			b.session.SendFleet(b.session.SelectedStar) // 派遣艦隊至選中星(航行由 EndTurn 推進)
			return b.goTo(b.galaxy, "星系主畫面")
		}
		if a == "loadmarines" && b.session != nil {
			n := b.session.LoadMarines(0) // 母星是唯一已知殖民地索引對映(見上方熱區註解)
			if n > 0 {
				b.lastActionMsg = fmt.Sprintf("已載運 %d 名陸戰隊上艦", n)
			} else {
				b.lastActionMsg = "無陸戰隊可載運(駐軍不足或艦隊已滿載)"
			}
			return b.goTo(b.galaxy, "星系主畫面")
		}
		if a == "invade" && b.session != nil {
			res := b.session.InvadeColony(b.session.SelectedStar)
			switch {
			case !res.Ok:
				b.lastActionMsg = res.Reason
			case res.AttackerWon:
				b.lastActionMsg = fmt.Sprintf("入侵勝利!佔領此星(存活 %d／敵剩 %d)", res.AttackerSurvived, res.DefenderSurvived)
			default:
				b.lastActionMsg = fmt.Sprintf("入侵失敗(我方存活 %d／敵剩 %d)", res.AttackerSurvived, res.DefenderSurvived)
			}
			return b.goTo(b.galaxy, "星系主畫面")
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
			if b.savePath != "" { // 每回合自動存檔(持久化對局)
				if err := b.session.Save(b.savePath); err != nil {
					fmt.Fprintln(os.Stderr, "自動存檔失敗:", err)
				}
			}
			// 若本回合完成的研究主題有多科技可選 → 先進抉擇畫面(MOO2 每主題擇一),
			// 選定後再顯示回合摘要。
			if _, _, pending := b.session.PendingResearchChoice(); pending {
				sc, err := b.researchChoice(b.turnSummary)
				if err == nil {
					return &origTransition{next: sc}
				}
			}
			return b.goTo(b.turnSummary, "回合摘要")
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
				fnt.Draw(dst, fmt.Sprintf("研究:%s", topicNameZh(b.lang, sess.Player.ResearchTopic)), 30, 82, 13, color.RGBA{160, 210, 240, 255})
				// 艦隊位置標記(青色三角)+ 航行目的連線。
				if sess.FleetAtStar >= 0 && sess.FleetAtStar < len(sess.Stars) {
					fx, fy := starScreenPos(sess.Stars[sess.FleetAtStar])
					if sess.FleetDestStar >= 0 && sess.FleetDestStar < len(sess.Stars) {
						dx, dy := starScreenPos(sess.Stars[sess.FleetDestStar])
						vector.StrokeLine(dst, float32(fx), float32(fy), float32(dx), float32(dy), 1, color.RGBA{80, 220, 220, 180}, false)
					}
					vector.DrawFilledRect(dst, float32(fx-4), float32(fy-4), 8, 8, color.RGBA{80, 240, 240, 255}, false)
				}
				// 選中星:顯示該星系行星資訊 + 派遣艦隊/載運陸戰隊/發動入侵按鈕(左下角面板)。
				if sess.SelectedStar >= 0 && sess.SelectedStar < len(sess.Planets) {
					p := sess.Planets[sess.SelectedStar]
					vector.DrawFilledRect(dst, 28, 326, 210, 110, color.RGBA{10, 14, 30, 235}, false)
					vector.StrokeRect(dst, 28, 326, 210, 110, 1, color.RGBA{90, 130, 200, 255}, false)
					fnt.Draw(dst, p.Name, 38, 344, 14, color.RGBA{240, 220, 120, 255})
					fnt.Draw(dst, fmt.Sprintf("氣候 %s ／ 大小 %s", p.Climate, p.Size), 38, 362, 11, color.RGBA{210, 216, 230, 255})
					fnt.Draw(dst, fmt.Sprintf("重力 %s ／ 礦產 %s", p.Gravity, p.Mineral), 38, 378, 11, color.RGBA{210, 216, 230, 255})
					// 陸戰隊狀態行:艦隊目前載運數,選中母星時另顯示殖民地駐軍池數(唯一已知對映)。
					marineLine := fmt.Sprintf("艦隊陸戰隊 %d", sess.FleetMarines)
					if sess.SelectedStar == 0 && len(sess.PlayerColonyMarines) > 0 {
						marineLine = fmt.Sprintf("艦隊陸戰隊 %d／殖民地駐軍 %d", sess.FleetMarines, sess.PlayerColonyMarines[0])
					}
					fnt.Draw(dst, marineLine, 38, 394, 11, color.RGBA{200, 220, 170, 255})
					// 操作鈕/狀態(與 galaxy() 建 hits 時的判斷邏輯一致)。
					switch {
					case b.lastActionMsg != "":
						vector.DrawFilledRect(dst, 38, 402, 190, 20, color.RGBA{30, 55, 35, 235}, false)
						vector.StrokeRect(dst, 38, 402, 190, 20, 1, color.RGBA{110, 200, 140, 255}, false)
						fnt.Draw(dst, b.lastActionMsg, 42, 415, 10, color.RGBA{225, 240, 225, 255})
					case sess.FleetETA > 0:
						fnt.Draw(dst, fmt.Sprintf("艦隊航行中…剩 %d 回合", sess.FleetETA), 38, 415, 11, color.RGBA{120, 200, 240, 255})
					case sess.SelectedStar == sess.FleetAtStar && sess.SelectedStar == 0:
						vector.DrawFilledRect(dst, 38, 402, 190, 20, color.RGBA{40, 70, 120, 255}, false)
						vector.StrokeRect(dst, 38, 402, 190, 20, 1, color.RGBA{110, 160, 230, 255}, false)
						fnt.Draw(dst, "▶ 載運陸戰隊", 46, 415, 12, color.RGBA{230, 235, 245, 255})
					case sess.SelectedStar == sess.FleetAtStar && sess.Stars[sess.SelectedStar].Owner == 2 && sess.FleetMarines > 0:
						vector.DrawFilledRect(dst, 38, 402, 190, 20, color.RGBA{120, 50, 40, 255}, false)
						vector.StrokeRect(dst, 38, 402, 190, 20, 1, color.RGBA{230, 130, 110, 255}, false)
						fnt.Draw(dst, "▶ 發動地面入侵", 46, 415, 12, color.RGBA{245, 235, 230, 255})
					case sess.SelectedStar == sess.FleetAtStar:
						fnt.Draw(dst, "艦隊已在此星", 38, 415, 11, color.RGBA{140, 200, 140, 255})
					default:
						vector.DrawFilledRect(dst, 38, 402, 190, 20, color.RGBA{40, 70, 120, 255}, false)
						vector.StrokeRect(dst, 38, 402, 190, 20, 1, color.RGBA{110, 160, 230, 255}, false)
						fnt.Draw(dst, "▶ 派遣艦隊至此星", 46, 415, 12, color.RGBA{230, 235, 245, 255})
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
			// 建造欄:項目名 + 進度(空則顯示「—」提示可點)。
			bt := "—"
			if i < len(b.session.Builds) && b.session.Builds[i].Name != "" {
				bd := b.session.Builds[i]
				bt = fmt.Sprintf("%s %d/%d", bd.Name, bd.Progress, bd.Cost)
			}
			s.extras = append(s.extras, extraText{x: 571, y: y, size: 12, text: bt, col: body, align: 1})
			// 已建建築(顯示效果來源):在建造欄下方以小字列出。
			if i < len(b.session.ColonyBuildings) && len(b.session.ColonyBuildings[i]) > 0 {
				names := make([]string, 0, len(b.session.ColonyBuildings[i]))
				for n := range b.session.ColonyBuildings[i] {
					names = append(names, n)
				}
				sort.Strings(names)
				lbl := "已建:" + strings.Join(names, "、")
				s.extras = append(s.extras, extraText{x: 571, y: y + 13, size: 10, text: lbl, col: color.RGBA{150, 200, 150, 255}, align: 1})
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
		{0, 0, moo2ScreenW, moo2ScreenH, "back"},
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
		y := 70.0
		for _, a := range b.session.AIPlayers {
			s.extras = append(s.extras,
				extraText{x: 40, y: y, size: 15, text: a.Name, col: gold},
				extraText{x: 40, y: y + 22, size: 12, text: fmt.Sprintf("態勢:%s ／ 軍力 %d ／ 佔領 %d 星", a.StanceName, a.FleetStrength, a.OwnedStars), col: body},
			)
			y += 56
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
	switch enemy {
	case "阿爾卡里":
		return 0
	case "布拉西":
		return 1
	case "達洛克":
		return 2
	case "埃雷里安":
		return 3
	case "諾蘭姆":
		return 4
	case "人類":
		return 5
	case "克拉肯":
		return 6
	case "梅克拉":
		return 7
	case "姆瑞森":
		return 8
	case "席隆":
		return 9
	case "薩克拉":
		return 10
	case "矽基":
		return 11
	case "崔拉里安":
		return 12
	case "賽隆人": // 舊字串相容(現行單一對手預設種族)
		return 10
	default:
		return 10
	}
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
	enemy := "薩克拉"
	return &diplomacyScreen{b: b, fnt: b.fnt, enemy: enemy, room: loadDiplomatScene(b.res, diplomatRaceIndex(enemy)),
		response: "薩克拉使節:人類,你有何提議?",
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
	d.fnt.DrawCentered(dst, "外交對談", 320, 62, 20, gold)
	d.fnt.DrawCentered(dst, d.enemy+" 使節", 320, 96, 14, color.RGBA{235, 150, 140, 255})
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
	d.fnt.DrawCentered(dst, "結束對談", float64(bx+bw/2), float64(by+bh/2), 15, body)
}

// diplomacy 進入外交對談畫面。
func (b *sceneBuilder) diplomacy() (origScreen, error) {
	if b.session == nil {
		return nil, fmt.Errorf("無對局")
	}
	playSceneBGM(bgmDiplo)
	return newDiplomacyScreen(b), nil
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
	ship           *ebiten.Image
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

// loadCombatShip 載入艦艇 sprite(CMBTSHP.LBX#0 frame0,59×60),借 COMBAT#11 調色盤。
// keyColor 用資產自身旗標(CMBTSHP flags=0x0000 → false):艦體外圍透明來自未寫入的
// RLE 像素(ToRGBA 一律留透明),而艦體本身含 index-0 深色像素須保留——先前誤設
// keyColor=true 會把 index-0 艦體也判成透明,導致 sprite 幾乎全消失(端到端截圖查出)。
// Phase 1 佔位:所有艦共用同一張圖,之後再依艦型/朝向擴充成完整對照表。
func loadCombatShip(res *assets.Resolver) *ebiten.Image {
	prov, err := decodeAsset(res, "combat.lbx", 11)
	if err != nil || prov.Embedded == nil {
		return nil
	}
	// 佔位艦型:CMBTSHP#30(大型艦,亮部結構在黑星空上清晰可讀;#0 是最小最暗的
	// 戰機,疊黑星空幾乎隱形——經黑底對比挑出 #30)。Phase 2 再按實際艦型/尺寸對照。
	im, err := decodeAsset(res, "cmbtshp.lbx", 30)
	if err != nil || len(im.Frames) == 0 {
		return nil
	}
	return ebiten.NewImageFromImage(im.Frames[0].ToRGBA(prov.Embedded, im.KeyColor()))
}

func newTacticalScreen(b *sceneBuilder) *tacticalScreen {
	p, e := b.session.StartCombat("賽隆人")
	// 戰鬥 RNG 依當前回合數種子:同一局同一回合的戰鬥可重現(不引入 wall-clock 不確定性)。
	seed := int64(b.session.Turn*2654435761 + 1013904223)
	return &tacticalScreen{b: b, fnt: b.fnt, player: p, enemy: e, sel: -1,
		log: "點我方艦選取→點空格移動;點敵艦→射程內我艦開火", pStart: len(p), eStart: len(e),
		rng: rand.New(rand.NewSource(seed)),
		bg:  loadCombatBG(b.res), bar: loadCombatBar(b.res), ship: loadCombatShip(b.res)}
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
		t.b.session.ApplyCombatOutcome("賽隆人", t.pStart, t.eStart, survivors, t.won)
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
		t.log = fmt.Sprintf("%s 移動到 (%d,%d)", t.player[t.sel].Name, col, row)
	}
	return nil
}

func (t *tacticalScreen) fireRound(target int) {
	tc, tr := t.enemy[target].Col, t.enemy[target].Row
	// 射程內我艦逐一依武器類型分流真戰鬥公式:beam(ResolveShot,不動)/missile
	// (ResolveMissileShot,躲避+AMR 攔截)/spherical(ResolveSphericalShot,現行武器表
	// 暫無掛載,分支保留供未來串接)。見 shell/weapon_kind.go 的分類依據。
	pAtk, firing := 0, 0
	for i := range t.player {
		s := &t.player[i]
		dist := abs(s.Col-tc) + abs(s.Row-tr)
		if dist > fireRange {
			continue
		}
		firing++
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
			shot = shell.ResolveShot(net, s.WeaponMin, s.WeaponMax, dist,
				enemy.ShieldReduction, enemy.ArmorHP, roll, false, false)
		}
		if shot.Hit {
			enemy.ArmorHP = shot.RemainingArmorHP
			enemy.HP -= shot.DamageToStructure
			pAtk += shot.DamageToStructure
		}
	}
	if firing == 0 {
		t.log = "目標超出射程,移動艦艇靠近再開火"
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
				t.player[wi].ShieldReduction, t.player[wi].ArmorHP, roll, false, false)
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
	t.log = fmt.Sprintf("第 %d 回合:%d 艦齊射 %d ／ 敵方還擊 %d", t.round, firing, pAtk, eAtk)
	if len(t.enemy) == 0 {
		t.over, t.won, t.log = true, true, "★ 敵艦隊全滅,勝利!點擊繼續"
	} else if len(t.player) == 0 {
		t.over, t.won, t.log = true, false, "✗ 我方艦隊全滅,敗北。點擊繼續"
	}
}

// drawShip 畫單艘艦:有 t.ship sprite 就縮放貼原版艦圖(敵方水平翻轉朝左),
// 否則 fallback 回原本的矩形 token 畫法。HP 條、艦名、選中金框一律疊在最上層,
// 不受美術是否載入影響。
func (t *tacticalScreen) drawShip(dst *ebiten.Image, s shell.CombatShip, base color.RGBA, selected bool, enemy bool) {
	x, y, w, h := cellRect(s.Col, s.Row)
	x, y, w, h = x+4, y+6, w-8, h-12
	if t.ship != nil {
		sb := t.ship.Bounds()
		sw0, sh0 := float64(sb.Dx()), float64(sb.Dy())
		sc := float64(h) / sh0 // 依格高等比縮放,寬度可能超出格寬少許,可接受
		op := &ebiten.DrawImageOptions{}
		if enemy {
			op.GeoM.Scale(-sc, sc)
			op.GeoM.Translate(float64(x)+sw0*sc, float64(y))
		} else {
			op.GeoM.Scale(sc, sc)
			op.GeoM.Translate(float64(x), float64(y))
		}
		dst.DrawImage(t.ship, op)
	} else {
		vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{base.R / 3, base.G / 3, base.B / 3, 255}, false)
	}
	sw := float32(1.5)
	sc := base
	if selected {
		sw, sc = 3, color.RGBA{255, 240, 120, 255}
	}
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), sw, sc, false)
	if t.fnt != nil {
		t.fnt.Draw(dst, s.Name, float64(x)+5, float64(y)+13, 11, color.RGBA{230, 235, 245, 255})
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
		t.fnt.DrawCentered(dst, "戰術戰鬥", 320, 34, 20, gold)
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
		t.drawBarLabelsCHT(dst) // 控制列烘進的英文按鈕疊中文(CLAUDE.md:button 也要中文化)
		logY = 343              // log 移到控制列上方星空,不壓按鈕
	}
	if t.fnt != nil {
		t.fnt.DrawCentered(dst, t.log, 320, logY, 14, color.RGBA{214, 220, 235, 255})
	}
}

// barButtonsCHT 是 COMBAT.LBX#0 控制列上各英文按鈕的螢幕中心座標 + 中文標籤。
// 座標於實際戰鬥截圖(gallery)量測;控制列貼在 y=moo2ScreenH-129=351。
// WEAPONS/SPECIALS 兩個欄位標頭在 remake 未用的清單面板內,略過。
var barButtonsCHT = []struct {
	cx, cy int
	label  string
}{
	{302, 378, "自動"}, {373, 378, "掃描"}, // AUTO / SCAN
	{302, 402, "登船"}, {373, 402, "撤退"}, // BOARD / RETREAT
	{302, 433, "等待"}, {373, 433, "完成"}, // WAIT / DONE
	{337, 461, "選項"}, // OPTIONS
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
		outcome, oc := "✗ 敗北", lose
		if bt.PlayerWon {
			outcome, oc = "★ 勝利!", win
		}
		s.extras = []extraText{
			{x: 40, y: 56, size: 15, text: fmt.Sprintf("對「%s」開戰", bt.Enemy), col: gold},
			{x: 40, y: 84, size: 16, text: outcome, col: oc},
			{x: 40, y: 110, size: 12, text: fmt.Sprintf("我方 %d 艦 ／ 敵方 %d 艦", bt.PlayerStart, bt.EnemyStart), col: body},
		}
		yy := 134.0
		for _, line := range bt.Log { // 逐回合戰報
			s.extras = append(s.extras, extraText{x: 40, y: yy, size: 12, text: line, col: body})
			yy += 20
		}
		s.extras = append(s.extras, extraText{x: 40, y: yy + 4, size: 13,
			text: fmt.Sprintf("損失:我方 %d 艦 ／ 敵方 %d 艦", bt.PlayerLosses, bt.EnemyLosses), col: gold})
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
		gold := color.RGBA{240, 220, 120, 255}
		s.extras = []extraText{
			{x: moo2ScreenW / 2, y: 30, size: 22, text: "銀河議會", col: gold, align: 1},
		}
		if b.session != nil {
			// 這裡刻意不重建原版議會投票 UI(見 docs/HONEST-STATUS.md「UI 未做」標注),只誠實
			// 呈現 shell.GameSession 已算好的議會狀態(gamedata/council.go + shell/council.go,
			// 依 GAME_MANUAL.pdf p.183 手冊原文):是否已成立、目前票數、是否已分出勝負/待回應。
			v := b.session.CouncilStatus()
			win := color.RGBA{120, 220, 140, 255}
			lose := color.RGBA{235, 120, 110, 255}
			neutral := color.RGBA{214, 220, 235, 255}
			var line1, line2 string
			var oc color.RGBA
			switch {
			case v.Victory.Over && v.Victory.Reason == engine.VictoryHighCouncil:
				line1 = fmt.Sprintf("已於第 %d 回合分出勝負(共召開 %d 屆選舉)", v.Victory.Turn, v.Meetings)
				if v.Victory.Winner == "player" {
					line2, oc = "★ 你已當選銀河領袖!", win
				} else {
					line2, oc = v.Victory.Winner+" 已當選銀河領袖", lose
				}
			case v.Pending != nil:
				line1 = fmt.Sprintf("第 %d 屆選舉:%s 以 %d/%d 票達2/3多數當選", v.Meetings, v.Pending.EnemyName, v.Pending.EnemyVotes, v.Pending.TotalVotes)
				line2, oc = "等待你回應是否接受(尚無互動介面,見 HONEST-STATUS)", neutral
			case !v.Eligible:
				line1 = "銀河議會尚未成立"
				line2, oc = "需半數銀河星系已殖民 + ≥2個存續帝國", neutral
			default:
				line1 = fmt.Sprintf("議會已成立(第 %d 屆待開)  我方 %d 票 ／ %s %d 票", v.Meetings+1, v.PlayerVotes, v.EnemyName, v.EnemyVotes)
				line2, oc = "尚無一方達2/3多數", neutral
			}
			s.extras = append(s.extras,
				extraText{x: moo2ScreenW / 2, y: 418, size: 15, text: line1, col: neutral, align: 1},
				extraText{x: moo2ScreenW / 2, y: 444, size: 17, text: line2, col: oc, align: 1},
			)
		}
	}
	return s, nil
}

// newGameSetup 建原版新遊戲設定畫面(NEWGAME.LBX 資產 28,調色盤鏈 RACEOPT#4→NEWGAME#1)。
// ACCEPT 進星系主畫面;CANCEL 回主選單。
func (b *sceneBuilder) newGameSetup() (*overlayScreen, error) {
	hits := []hitRegion{
		{86, 100, 130, 108, "diff"},  // 難度選擇框
		{232, 100, 150, 108, "size"}, // 星系大小選擇框
		{86, 244, 130, 108, "race"},  // 種族選擇框(PLAYERS 位置)
		{92, 392, 108, 30, "cancel"},
		{432, 392, 108, 30, "accept"},
	}
	onAction := func(a string) *origTransition {
		switch a {
		case "diff":
			b.newGameDiff = (b.newGameDiff + 1) % len(shell.Difficulties)
			return b.goTo(b.newGameSetup, "新遊戲設定")
		case "size":
			b.newGameSize = (b.newGameSize + 1) % len(shell.GalaxySizes)
			return b.goTo(b.newGameSetup, "新遊戲設定")
		case "race", "accept":
			// 原版流程:星系設定 → Accept →【獨立種族選擇畫面】(不在此直接開局)。
			// 點種族框或按 Accept 都進種族選擇;開局的 RegenGalaxy/ApplyRace 移到該畫面。
			sc, err := b.raceSelect()
			if err != nil {
				fmt.Fprintf(os.Stderr, "載入種族選擇: %v\n", err)
				return nil
			}
			return &origTransition{next: sc}
		}
		return b.goTo(b.menu, "主選單")
	}
	// 座標經 PIL 量測(remain-scan/newgame_a28_f00.png);開關標籤移到核取框右側(x430)避免採到藍框。
	overlays := []labelRect{
		{86, 78, 130, 22, "DIFFICULTY", 0},
		{232, 78, 150, 22, "GALAXY SIZE", 0},
		{398, 78, 150, 22, "GALAXY AGE", 0},
		{86, 222, 130, 22, "RACE", 0},
		{232, 222, 150, 22, "TECH LEVEL", 0},
		{422, 266, 138, 18, "TACTICAL COMBAT", 11},
		{422, 301, 138, 18, "RANDOM EVENTS", 11},
		{422, 334, 138, 18, "ANTARANS ATTACK", 11},
		{100, 388, 96, 24, "CANCEL", 0},
		{440, 388, 96, 24, "ACCEPT", 0},
	}
	s, err := loadOverlayScreen(b.res, "newgame.lbx", 28, b.lang, b.fnt, "assets/i18n/menu.tsv",
		overlays, color.RGBA{210, 216, 230, 255}, 13, hits, onAction,
		paletteChain{{"raceopt.lbx", 4}, {"newgame.lbx", 1}})
	if err != nil {
		return nil, err
	}
	// 選定的難度 + 星系大小顯示在各自選擇框內。
	if b.fnt != nil {
		gs := shell.GalaxySizes[b.newGameSize]
		df := shell.Difficulties[b.newGameDiff]
		gold := color.RGBA{240, 220, 120, 255}
		body := color.RGBA{210, 216, 230, 255}
		rc := shell.Races[b.newGameRace]
		s.extras = []extraText{
			{x: 151, y: 150, size: 16, text: df.Name, col: gold, align: 1},
			{x: 307, y: 150, size: 16, text: fmt.Sprintf("%s (%d 星)", gs.Name, gs.Stars), col: gold, align: 1},
			{x: 151, y: 286, size: 16, text: rc.Name, col: gold, align: 1},
			{x: 151, y: 312, size: 10, text: rc.Desc, col: body, align: 1},
		}
	}
	return s, nil
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
		{300, 58, 300, 22, "weapon"}, // 元件選擇(點擊各列循環)
		{300, 82, 300, 22, "armor"},
		{300, 106, 300, 22, "shield"},
		{300, 130, 300, 22, "special"},
		{0, 0, moo2ScreenW, moo2ScreenH, "back"},
	}
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
		if zh, ok := hullZH[a]; ok && b.session != nil {
			// 建造前驗證空間:超出艦體空間上限(shell.ShipDesignFits)就擋下,留在設計畫面提示,不扣款不造艦。
			if !shell.ShipDesignFits(zh, b.designWeapon, b.designArmor, b.designShield, b.designSpecial) {
				b.designMsg = fmt.Sprintf("空間不足,無法建造%s(目前元件超出艦體空間上限)", zh)
				return b.goTo(b.shipDesign, "艦艇設計")
			}
			b.designMsg = ""
			b.session.BuildShip(zh, b.designWeapon, b.designArmor, b.designShield, b.designSpecial)
			return b.goTo(b.fleet, "艦隊列表")
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
	s, err := loadOverlayScreen(b.res, "design.lbx", 0, b.lang, b.fnt, "assets/i18n/tech.tsv",
		overlays, color.RGBA{206, 214, 232, 255}, 13, hits, onAction,
		paletteChain{{"buffer0.lbx", 0}})
	if err != nil {
		return nil, err
	}
	// 各艦體成本(對齊 MOO2 空殼生產成本)+ 目前國庫,顯示在艦體清單右方。
	if b.fnt != nil && b.session != nil {
		body := color.RGBA{210, 216, 230, 255}
		classes := []string{"巡防艦", "驅逐艦", "巡洋艦", "戰艦", "泰坦", "末日之星"}
		for i, cl := range classes {
			s.extras = append(s.extras, extraText{x: 250, y: float64(60 + i*17), size: 11,
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
			{"武器", w, fmt.Sprintf("+%d攻", w.Value)},
			{"裝甲", ar, fmt.Sprintf("+%dHP", ar.Value)},
			{"護盾", sd, fmt.Sprintf("+%dHP", sd.Value)},
			{"特殊", sp, ""},
		}
		for i, r := range rows {
			y := float64(69 + i*24)
			s.extras = append(s.extras,
				extraText{x: 305, y: y, size: 12, text: r.label + " ▸ " + r.c.Name, col: gold},
				extraText{x: 470, y: y, size: 11, text: fmt.Sprintf("%s %dBC", r.eff, r.c.Cost), col: color.RGBA{200, 208, 225, 255}})
		}
		total := shell.DesignCost("巡洋艦", b.designWeapon, b.designArmor, b.designShield, b.designSpecial)
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
			extraText{x: 305, y: 168, size: 12, text: fmt.Sprintf("巡洋艦總價 %d BC", total), col: color.RGBA{170, 220, 180, 255}},
			extraText{x: 305, y: 190, size: 11, text: fmt.Sprintf("已解鎖 武器%d/%d 裝甲%d/%d 護盾%d/%d 特殊%d/%d(研究科技解鎖進階元件)",
				cnt(shell.WeaponOptions), len(shell.WeaponOptions), cnt(shell.ArmorOptions), len(shell.ArmorOptions),
				cnt(shell.ShieldOptions), len(shell.ShieldOptions), cnt(shell.SpecialOptions), len(shell.SpecialOptions)),
				col: color.RGBA{170, 200, 240, 255}},
			extraText{x: 12, y: 460, size: 12, text: fmt.Sprintf("國庫 %d BC", b.session.Player.BC), col: gold})

		// 空間預算/已用(依目前選定元件即時計算):逐艦體列出「空間:已用／總」,超格轉紅並標
		// 「空間不足」。點艦體列建造時,onAction 用同一份 shell.ShipDesignFits 判斷擋下建造
		// (不扣款、不入艦隊),designMsg 顯示擋下提示——顯示與建造驗證共用同一份判斷,不會不一致。
		spaceHeaderY := 208.0
		s.extras = append(s.extras, extraText{x: 305, y: spaceHeaderY, size: 12,
			text: "各艦體空間(依目前元件):", col: gold})
		okCol := color.RGBA{170, 220, 180, 255}
		badCol := color.RGBA{230, 90, 90, 255}
		for i, cl := range classes {
			used := shell.ShipDesignSpaceUsed(cl, b.designWeapon, b.designArmor, b.designShield, b.designSpecial)
			totalSp := gamedata.ShipHullSpace(gamedata.CombatShipClass(i))
			fits := used <= totalSp
			txt := fmt.Sprintf("%s 空間:%d／%d", cl, used, totalSp)
			col := okCol
			if !fits {
				txt += "(空間不足)"
				col = badCol
			}
			s.extras = append(s.extras, extraText{x: 305, y: spaceHeaderY + 17 + float64(i*17), size: 11, text: txt, col: col})
		}
		if b.designMsg != "" {
			s.extras = append(s.extras, extraText{x: 305, y: spaceHeaderY + 17 + float64(len(classes)*17) + 8, size: 12,
				text: b.designMsg, col: badCol})
		}
	}
	return s, nil
}

// officer 建原版軍官列表畫面(OFFICER.LBX 資產 0)。座標經 PIL 量測
// (screens-scan/officer_leaderlist.png):頁籤列 y=12-32,按鈕列 y=440-462。
func (b *sceneBuilder) officer() (*overlayScreen, error) {
	// 精確返回鍵熱區(用已量測的 RETURN 按鈕座標,取代整畫面返回;更忠實:僅返回鍵返回)。
	hits := []hitRegion{{540, 440, 80, 20, "Return"}}
	onAction := func(a string) *origTransition {
		if a == "Return" {
			return b.goTo(b.galaxy, "星系主畫面")
		}
		return nil
	}
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
		yy := 168.0
		if out.ResearchDone {
			s.extras = append(s.extras, extraText{x: 40, y: yy, size: 14, text: "★ 完成一項研究!", col: color.RGBA{120, 220, 140, 255}})
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
	scale    int // 目前視窗放大倍率(1~4)

	audio *moo2audio.Mixer // 持有音訊 Mixer,避免 player 被 GC(headless 為 nil)

	// 過場截圖廊(-gamegallery):script 為導覽腳本,galleryShots 指定在哪個絕對 tick
	// 存哪張圖(可多張,依序達成)。與單張 shotPath 模式互斥。
	galleryDir   string
	galleryShots []galleryShot
	galleryDone  int
}

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
	script := []shell.InputState{
		idle,            // t1: 主選單(未點擊)
		click(491, 228), // t2: 主選單「新遊戲」→ 新遊戲設定
		click(486, 405), // t3: 新遊戲設定「Accept」→ 種族選擇
		click(540, 451), // t4: 種族選擇「接受」→ 命名/旗色
		click(540, 454), // t5: 命名/旗色「接受」→ 星系主畫面
		click(48, 452),  // t6: 星系主畫面工具列「殖民地」→ 殖民地總覽
		click(608, 462), // t7: 殖民地總覽「RETURN」→ 星系主畫面
		click(495, 452), // t8: 星系主畫面工具列「INFO」→ 科技總覽
		click(113, 89),  // t9: 科技總覽「Tech Review」→ 研究選擇
		click(204, 186), // t10: 研究選擇(任一領域,如 Chemistry)→ 星系主畫面
		click(420, 452), // t11: 星系主畫面工具列「RACES」→ 種族關係
		click(483, 428), // t12: 種族關係「REPORT」→ 外交對談
		click(320, 437), // t13: 外交對談「結束對談」→ 種族關係
		click(388, 448), // t14: 種族關係「DECLARE WAR」→ 戰術戰鬥
	}
	shots := []galleryShot{
		{1, "01_menu.png"},
		{3, "02_raceselect.png"},
		{4, "03_nameflag.png"},
		{5, "04_galaxy.png"},
		{6, "05_colony.png"},
		{9, "06_research.png"},
		{12, "07_diplomacy.png"},
		{14, "08_tactical.png"},
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
	if a.script == nil { // 互動模式才處理視窗快捷鍵(headless 略過)
		a.handleWindowKeys()
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
		if a.galleryDone >= len(a.galleryShots) {
			return ebiten.Termination
		}
		// 硬性終止保護:即使某些圖因導覽失敗/Draw 跳幀而存不到,超過最後一張的
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
func runInteractive(dirs []string, lang i18n.Lang, fnt *uifont.Font,
	script []shell.InputState, shot string, frames int, galleryDir string) error {

	if lang == i18n.Traditional && fnt == nil {
		return fmt.Errorf("中文模式需以 -font 指定 CJK 字型")
	}
	res, err := assets.NewResolver(dirs...)
	if err != nil {
		return err
	}
	b := &sceneBuilder{res: res, fnt: fnt, lang: lang, session: shell.NewDemoSession(), newGameSize: 1, newGameDiff: 1, designWeapon: 1, savePath: savePathFor()}
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
	}

	// 預設放大 2 倍(headless 驗證/截圖廊維持 1 倍);視窗可自由拉伸,內容等比縮放置中。
	scale := 2
	if shot != "" || galleryDir != "" {
		scale = 1
	}
	app := &interactiveApp{cur: menu, script: script, shotPath: shot, frames: frames, scale: scale,
		galleryDir: galleryDir, galleryShots: shots}
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
