package main

// hotkeys.go:星圖的鍵盤快捷鍵(手冊行文寫死的那幾個)。
//
// 規則面(循環順序、清單怎麼取)在 `internal/shell/starnav.go`;這一檔只管
// 「哪個實體按鍵對應哪個代碼」與「F9 測距的畫面呈現」。
//
// ============ 為什麼 F9 值得先做 ============
//
// 秒差距模型在 gap report 第 45 項就建好了(`ParsecsBetweenStars`、1 秒差距 = 30 遊戲單位、
// 距離取整),而且引擎速度、星雲減速、干擾器範圍全都掛在它上面。
// 但玩家在畫面上**看不到任何秒差距數字**——整套模型是隱形的。
//
// 手冊逐字:「To quickly find the distance between two star systems, use the keyboard
// shortcut [F9]. You'll need to click on the first star, then move the mouse cursor over
// any other star to see the distance (in parsecs) between them.」
//
// 也就是**兩段式**:按 F9 進測距模式 → 點第一顆 → 游標移到第二顆就顯示。
// 不是「點兩顆看結果」,而是**跟著游標即時更新**。

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// starHitHalf 是星球點擊/懸停熱區的半邊長(22×22 方框)。
// 點選熱區與測距的懸停判定共用它——兩邊各寫一份遲早會差幾個像素。
const starHitHalf = 11

// hotkeyBindings 是實體按鍵 → 代碼的對照。
//
// 只放手冊行文中直接寫死的(見 shell 的 Hotkey* 常數註解);ALT+Fn 那組刻意不放。
var hotkeyBindings = []struct {
	key  ebiten.Key
	code string
}{
	{ebiten.KeyF1, shell.HotkeyNextFleet},
	{ebiten.KeyF2, shell.HotkeyPrevFleet},
	{ebiten.KeyF5, shell.HotkeyNextColony},
	{ebiten.KeyF6, shell.HotkeyPrevColony},
	{ebiten.KeyF9, shell.HotkeyMeasureDist},
}

// pollHotkey 回傳這一幀剛按下的快捷鍵代碼(沒有回空字串)。
//
// 邊緣觸發:按住不連發,和原版一樣是一次一步。
func pollHotkey() string {
	for _, b := range hotkeyBindings {
		if inpututil.IsKeyJustPressed(b.key) {
			return b.code
		}
	}
	return ""
}

// ---- F9 測距模式 ----

// measureState 是測距模式的畫面狀態。
//
// 放在 sceneBuilder 而不是 GameSession:它是**看的方式**不是世界狀態,不該進存檔
// (同 `SetNebulaProbe` 那個判斷)。
type measureState struct {
	on   bool // 已按 F9,等著點第一顆星
	from int  // 已點的第一顆星索引;−1 = 還沒點
}

// measureReset 關掉測距模式。
func (m *measureState) reset() {
	m.on, m.from = false, -1
}

// handleGalaxyHotkey 處理星圖上的快捷鍵,回傳是否要重繪。
//
// F1/F2/F5/F6 都只是換選取星——把視角帶過去,不動遊戲狀態。
func (b *sceneBuilder) handleGalaxyHotkey(code string) bool {
	sess := b.session
	if sess == nil || code == "" {
		return false
	}
	switch code {
	case shell.HotkeyMeasureDist:
		// 再按一次 F9 = 取消(原版沒有明說怎麼離開,但「同一個鍵切換」是這類模式的常規,
		// 而且不給出口會把玩家鎖在測距模式裡)。
		if b.measure.on {
			b.measure.reset()
		} else {
			b.measure.on, b.measure.from = true, -1
		}
		return true
	case shell.HotkeyNextFleet, shell.HotkeyPrevFleet:
		idx := sess.CycleFleetStar(sess.SelectedStar, code == shell.HotkeyNextFleet)
		if idx < 0 {
			return false
		}
		sess.SelectedStar = idx
		b.lastActionMsg = ""
		return true
	case shell.HotkeyNextColony, shell.HotkeyPrevColony:
		idx := sess.CycleColonizedStar(sess.SelectedStar, code == shell.HotkeyNextColony)
		if idx < 0 {
			return false
		}
		sess.SelectedStar = idx
		b.lastActionMsg = ""
		return true
	}
	return false
}

// measureClickedStar 在測距模式下吃掉一次點星:第一次記起點,之後換起點。
// 回傳 true 表示這次點擊已被測距模式消化,不要再當成「選取星球」。
func (b *sceneBuilder) measureClickedStar(idx int) bool {
	if !b.measure.on {
		return false
	}
	b.measure.from = idx
	return true
}

// drawMeasureOverlay 畫測距模式的提示與結果。
//
// 起點已定時,從起點拉一條線到游標所在的星,並在中點標出秒差距——
// 「移到哪就看到哪」是手冊描述的行為,不是點完兩顆才出數字。
func (b *sceneBuilder) drawMeasureOverlay(dst *ebiten.Image, mx, my int) {
	if !b.measure.on || b.session == nil || b.fnt == nil {
		return
	}
	sess := b.session
	tint := color.RGBA{120, 235, 200, 255}
	// 提示字**跟著游標走**,不釘在畫面角落——星圖每個角落都可能有星星和星名,
	// 固定位置一定會壓到某顆星(第一版釘在 (30,34),結果蓋掉左上角那顆星的名字)。
	// (參數收已翻好的字而不是 zh/en 一對:英文缺口棘輪只認得直接寫在呼叫點的 `tr(...)`,
	//  包一層就漏掉了——那支測試的價值正是「別讓中文字面值繞過翻譯」。)
	hint := func(s string) {
		b.fnt.Draw(dst, s, float64(mx+12), float64(my-6), 11, tint)
	}
	if b.measure.from < 0 || b.measure.from >= len(sess.Stars) {
		hint(b.tr("測距:點選第一顆星", "MEASURE: click first star"))
		return
	}
	fx, fy := starScreenPos(sess.Stars[b.measure.from])
	vector.StrokeCircle(dst, float32(fx), float32(fy), 9, 1, tint, true)

	to := b.starAtScreen(mx, my)
	if to < 0 || to == b.measure.from {
		hint(b.tr("測距:移到另一顆星", "MEASURE: hover another star"))
		return
	}
	tx, ty := starScreenPos(sess.Stars[to])
	vector.StrokeLine(dst, float32(fx), float32(fy), float32(tx), float32(ty), 1, tint, true)
	pc := sess.ParsecsBetweenStars(b.measure.from, to)
	b.fnt.DrawCentered(dst, fmt.Sprintf(b.tr("%d 秒差距", "%d PC"), pc),
		float64(fx+tx)/2, float64(fy+ty)/2-8, 11, tint)
}

// starAtScreen 回傳螢幕座標 (mx,my) 落在哪顆星上(沒有回 −1)。
//
// 用的是和點選熱區**同一個** 22×22 方框(`starHitHalf`),所以「點得到的星就懸停得到」——
// 兩邊各寫一份判定,遲早會差幾個像素。
func (b *sceneBuilder) starAtScreen(mx, my int) int {
	if b.session == nil {
		return -1
	}
	vis := b.session.VisibleStars()
	for i := range b.session.Stars {
		if vis != nil && i < len(vis) && !vis[i] {
			continue
		}
		x, y := starScreenPos(b.session.Stars[i])
		box := hitRegion{x - starHitHalf, y - starHitHalf, 2 * starHitHalf, 2 * starHitHalf, ""}
		if box.hit(mx, my) {
			return i
		}
	}
	return -1
}

// galleryMeasureTarget 是截圖廊要把游標放到哪一顆星上(沒有合適的回 −1)。
//
// 挑法:第一顆**可見且不是起點**的星。不能寫死索引——星圖有戰爭迷霧,固定索引很可能
// 落在還沒探索、畫都畫不出來的星上,截圖就只會停在「移到另一顆星」的提示。
func (b *sceneBuilder) galleryMeasureTarget() int {
	if b.session == nil {
		return -1
	}
	vis := b.session.VisibleStars()
	for i := range b.session.Stars {
		if i == galleryMeasureFrom {
			continue
		}
		if vis != nil && i < len(vis) && !vis[i] {
			continue
		}
		return i
	}
	return -1
}
