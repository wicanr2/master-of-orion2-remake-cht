package main

// raceselect.go:原版獨立「種族選擇」畫面(HANDOFF 優先2)。
//
// 依 docs/tech/newgame-flow.md:原版流程是 星系設定 → Accept → **獨立種族選擇畫面**
// (13 族 + Custom,滑過顯示肖像/能力)→ Custom 則進點數畫面 → 命名/旗色 → 進遊戲。
// 目前 remake 把種族擠進設定畫面一格;本畫面把它拆成原版的獨立畫面。
//
// 資產:背景用 RACEOPT#0(螢幕外框,內嵌調色盤);肖像用 RACESEL 15..28(內嵌調色盤)。
// 肖像↔種族對應經讀圖確認為「字母序」(asset 15=Alkari、20=Humans 已渲染核對),
// portrait asset = 15 + 字母序 index。版面為合成近似,**尚未對原版截圖像素對齊**(待優先3)。

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/uifont"
)

// raceEntry 是種族選擇清單一列:中/英名、RACESEL 肖像 asset、對應 shell.Races 索引(-1=Custom)。
type raceEntry struct {
	zh, en   string
	portrait int
	shellIdx int
}

// raceSelectList 依原版字母序排列(對齊 RACESEL 肖像 15..28),shellIdx 指回 shell.Races。
// Custom(28)尚無點數畫面,暫以無加成處理(TODO:接 Race Customization 畫面)。
var raceSelectList = []raceEntry{
	{"阿爾卡里", "Alkari", 15, 6},
	{"布拉西", "Bulrathi", 16, 5},
	{"達洛克", "Darloks", 17, 8},
	{"埃雷里安", "Elerians", 18, 10},
	{"諾蘭姆", "Gnolams", 19, 11},
	{"人類", "Humans", 20, 0},
	{"克拉肯", "Klackons", 21, 3},
	{"梅克拉", "Meklars", 22, 7},
	{"姆瑞森", "Mrrshan", 23, 4},
	{"席隆", "Psilons", 24, 1},
	{"薩克拉", "Sakkra", 25, 2},
	{"矽基", "Silicoids", 26, 12},
	{"崔拉里安", "Trilarians", 27, 9},
	{"自訂種族", "Custom", 28, -1},
}

type raceSelectScreen struct {
	b         *sceneBuilder
	fnt       *uifont.Font
	bg        *ebiten.Image
	portraits map[int]*ebiten.Image // 肖像 asset id → 圖(惰性載入)
	sel       int                   // 目前選定列(raceSelectList 索引)
	hover     int                   // 滑鼠懸停列(-1=無)
}

// 版面常數(合成近似,待對原版截圖校正)。
const (
	rsListX, rsListY = 34, 86 // 左側種族名清單起點
	rsRowH           = 24     // 每列高(14 列須容於底部按鈕之上)
	rsListW          = 150
	rsPortX, rsPortY = 320, 70 // 右側肖像放置點
	rsPortW          = 250     // 肖像顯示寬(等比縮放)
)

// raceSelect 建構種族選擇畫面。預設選「人類」(清單索引 5)。
func (b *sceneBuilder) raceSelect() (origScreen, error) {
	s := &raceSelectScreen{
		b: b, fnt: b.fnt, portraits: map[int]*ebiten.Image{}, sel: 5, hover: -1,
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

func (s *raceSelectScreen) rowRect(i int) (x, y, w, h int) {
	return rsListX, rsListY + i*rsRowH, rsListW, rsRowH - 3
}

// 底部按鈕(取消 / 接受)。
func (s *raceSelectScreen) cancelRect() (int, int, int, int) { return 40, 436, 120, 30 }
func (s *raceSelectScreen) acceptRect() (int, int, int, int) { return 480, 436, 120, 30 }

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
	if !in.ClickReleased {
		return nil
	}
	// 選種族列。
	if s.hover >= 0 {
		s.sel = s.hover
		if clickSound != nil {
			clickSound()
		}
		return nil
	}
	// 取消 → 回星系設定。
	if x, y, w, h := s.cancelRect(); hitBox(in.MouseX, in.MouseY, x, y, w, h) {
		if clickSound != nil {
			clickSound()
		}
		return s.b.goTo(s.b.newGameSetup, "星系設定")
	}
	// 接受 → 若為自訂種族進點數畫面;否則套用種族 + 產生星系 → 星系主畫面。
	if x, y, w, h := s.acceptRect(); hitBox(in.MouseX, in.MouseY, x, y, w, h) {
		if clickSound != nil {
			clickSound()
		}
		if raceSelectList[s.sel].shellIdx < 0 { // 自訂種族 → 點數畫面
			sc, err := s.b.customRace()
			if err != nil {
				return nil
			}
			return &origTransition{next: sc}
		}
		s.applyAndStart()
		return &origTransition{next: s.b.nameFlag(raceSelectList[s.sel].zh + "帝國")}
	}
	return nil
}

// applyAndStart 依設定畫面選的難度/大小 + 本畫面選的種族開新局。
// (真實母星/起始殖民地依 Starting Civilization 設定,屬後續步驟;此處沿用現有 RegenGalaxy。)
func (s *raceSelectScreen) applyAndStart() {
	b := s.b
	if b.session == nil {
		return
	}
	r := raceSelectList[s.sel]
	b.session.Difficulty = b.newGameDiff
	b.newGameSeed++
	b.session.RegenGalaxy(shell.GalaxySizes[b.newGameSize].Stars, int64(b.newGameSeed*7919+42))
	if r.shellIdx >= 0 {
		b.newGameRace = r.shellIdx
		b.session.ApplyRace(r.shellIdx) // Custom(-1)暫不套加成,待點數畫面
	}
}

func (s *raceSelectScreen) draw(dst *ebiten.Image) {
	dst.Fill(color.RGBA{0, 0, 0, 255})
	if s.bg != nil {
		dst.DrawImage(s.bg, nil)
	}
	if s.fnt == nil {
		return
	}
	gold := color.RGBA{240, 220, 120, 255}
	body := color.RGBA{206, 218, 240, 255}
	dim := color.RGBA{150, 160, 180, 255}

	s.fnt.DrawCentered(dst, "選擇你的種族", 320, 54, 18, gold)

	// 左側種族名清單。
	for i, r := range raceSelectList {
		x, y, w, h := s.rowRect(i)
		if i == s.sel {
			vector.DrawFilledRect(dst, float32(x-2), float32(y-1), float32(w+4), float32(h+2),
				color.RGBA{40, 60, 90, 220}, false)
			vector.StrokeRect(dst, float32(x-2), float32(y-1), float32(w+4), float32(h+2), 1.2,
				color.RGBA{120, 170, 230, 255}, false)
		} else if i == s.hover {
			vector.DrawFilledRect(dst, float32(x-2), float32(y-1), float32(w+4), float32(h+2),
				color.RGBA{30, 40, 60, 160}, false)
		}
		col := body
		if i != s.sel && i != s.hover {
			col = dim
		}
		s.fnt.Draw(dst, r.zh, float64(x+6), float64(y)+float64(h)/2-8, 15, col)
	}

	// 右側:選定種族肖像(等比縮放置入)。
	r := raceSelectList[s.sel]
	if p := s.portrait(r.portrait); p != nil {
		pw := p.Bounds().Dx()
		ph := p.Bounds().Dy()
		scale := float64(rsPortW) / float64(pw)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(float64(rsPortX), float64(rsPortY))
		dst.DrawImage(p, op)
		// 肖像下方:種族名 + 描述。
		ty := float64(rsPortY) + float64(ph)*scale + 10
		s.fnt.DrawCentered(dst, r.zh, float64(rsPortX)+float64(rsPortW)/2, ty, 18, gold)
		if r.shellIdx >= 0 {
			s.fnt.DrawCentered(dst, shell.Races[r.shellIdx].Desc,
				float64(rsPortX)+float64(rsPortW)/2, ty+24, 11, body)
		} else {
			s.fnt.DrawCentered(dst, "自訂種族點數(點數畫面待實作)",
				float64(rsPortX)+float64(rsPortW)/2, ty+24, 11, dim)
		}
	}

	// 底部按鈕。
	drawButton := func(rect func() (int, int, int, int), label string, accent color.RGBA) {
		x, y, w, h := rect()
		vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{34, 34, 44, 255}, false)
		vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1.5, accent, false)
		s.fnt.DrawCentered(dst, label, float64(x+w/2), float64(y+h/2), 15, body)
	}
	drawButton(s.cancelRect, "取消", color.RGBA{160, 140, 100, 255})
	drawButton(s.acceptRect, "接受", color.RGBA{120, 200, 130, 255})
}
