package main

// relocation.go:星圖的**遷移連線**(原版 `Draw_Relocation_Links_` @ 0x85320)。
//
// 這是主畫面圖層順序裡的第 2 層(見 shipicon.go 檔頭),也是 remake 4 層裡最後補上的一層。
// 規則面在 `internal/shell/relocation.go`。
//
// ============ 顏色是真值,而且它是一條會流動的線 ============
//
// 原版把一個 **8 位元組的調色盤索引表**丟給畫線函式 `sub_A11C0`:
//
//	dword_81C80  dd 2 dup(70706F6Eh)     ; = 6E 6F 70 70 6E 6F 70 70
//
// 也就是四個索引 110/111/112/112 重複兩次。`sub_A11C0` 另外收一個相位參數,
// 開頭就是 `edi = 7 - 相位`,並依線的方向把表反轉——**那是讓漸層沿著線跑**,
// 玩家因此看得出方向(新艦是從殖民地往集結點送)。
//
// ⚠ **相位的來源沒有追死**:那個全域(`word_19995E`)在星圖這一段只被讀、沒被寫,
// 寫入端落在另一個模組。所以 remake 用自己的 `animTick` 驅動相位,
// 步進沿用已經被證實三次的「**2 次重畫換 1 步**」比例(黑洞、一般星球、這裡各一次)。
// **比例是真值,絕對速度是 remake 的選擇。**

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/master-of-orion2-remake-cht/internal/shell"
)

// relocationRamp 是原版 `dword_81C80` 那 8 個調色盤索引。
//
// 直接寫索引而不是 RGB:它們要用星圖自己的調色盤(buffer0.lbx#0)查,
// 換調色盤時顏色會跟著對。
var relocationRamp = [8]uint8{0x6E, 0x6F, 0x70, 0x70, 0x6E, 0x6F, 0x70, 0x70}

// relocationDashLen 是漸層每一段的長度(像素)。
//
// ⚠ 原版是逐像素走 Bresenham 取表,remake 用向量畫線,沒有逐像素的鉤子。
// 用「每 4 px 換一段顏色」近似同一個視覺:**這是 remake 的選擇,不是原版真值**。
const relocationDashLen = 4

// relocationHoldRedraws 是相位每一步停留幾次重畫(同黑洞那個 2,見 starsprite.go)。
const relocationHoldRedraws = 2

// relocationPalette 是星圖用的調色盤鏈(同星球 sprite / 星雲)。
var relocationPalette = paletteChain{{"buffer0.lbx", 0}}

// relocationColors 解出 8 個漸層色(解不到回 nil,呼叫端退回單色)。
func (b *sceneBuilder) relocationColors() []color.RGBA {
	if b.relocColors != nil {
		return b.relocColors
	}
	if b.res == nil {
		return nil
	}
	im, err := decodeAsset(b.res, "buffer0.lbx", 0)
	if err != nil {
		return nil
	}
	pal, err := resolvePalette(b.res, im, relocationPalette)
	if err != nil || pal == nil {
		return nil
	}
	out := make([]color.RGBA, len(relocationRamp))
	for i, idx := range relocationRamp {
		if int(idx) >= len(*pal) {
			return nil
		}
		c := (*pal)[idx]
		out[i] = color.RGBA{c.R, c.G, c.B, 255}
	}
	b.relocColors = out
	return out
}

// relocationFallback 是解不到調色盤時的單色(偏青,和艦隊航行線區隔開)。
var relocationFallback = color.RGBA{140, 210, 190, 220}

// ⚠ **這條線本來就很暗,那是原版的樣子,不是畫錯。**
//
// 那三個索引在 BUFFER0.LBX#0 的調色盤裡解出來是 (0,20,0) / (4,56,4) / (0,76,0)
// ——深綠,壓在黑色星空上很低調。手冊自己也是這樣描述的:
// 「If you'd rather not clutter up the galaxy with them, turn this option off.」
// 它的定位就是**可以關掉的雜訊**,不是要搶眼的指示線。
//
// 所以這裡**不把顏色調亮**。要驗證它有沒有畫出來,用截圖加亮(`-evaluate multiply 4`)去看,
// 不要為了「看得清楚」去改一個已經是真值的數字。

// drawRelocationLinks 畫出所有遷移連線。
//
// 畫在星星**之下**(和原版的圖層順序一致:遷移連線是第 2 層、星星是第 3 層),
// 所以它不會蓋住星點。
func (b *sceneBuilder) drawRelocationLinks(dst *ebiten.Image) {
	sess := b.session
	if sess == nil {
		return
	}
	links := sess.RelocationLinks()
	if len(links) == 0 {
		return
	}
	ramp := b.relocationColors()
	phase := b.animTick / relocationHoldRedraws
	for _, l := range links {
		x1, y1 := starScreenPos(sess.Stars[l.From])
		x2, y2 := starScreenPos(sess.Stars[l.To])
		drawRampLine(dst, float32(x1), float32(y1), float32(x2), float32(y2), ramp, phase)
	}
}

// drawRampLine 把一條線切成 relocationDashLen 長的小段,逐段換 ramp 的下一個顏色。
//
// 相位往前走 = 顏色往前移 = **漸層沿著線流動**,方向就看得出來了。
func drawRampLine(dst *ebiten.Image, x1, y1, x2, y2 float32, ramp []color.RGBA, phase int) {
	if len(ramp) == 0 {
		vector.StrokeLine(dst, x1, y1, x2, y2, 1, relocationFallback, true)
		return
	}
	dx, dy := x2-x1, y2-y1
	length := float32(dx*dx + dy*dy)
	if length <= 0 {
		return
	}
	// 段數:用曼哈頓距離近似長度就夠了(只是切段,不需要開根號的精度)。
	approx := absF(dx) + absF(dy)
	n := int(approx / relocationDashLen)
	if n < 1 {
		n = 1
	}
	for i := 0; i < n; i++ {
		t0 := float32(i) / float32(n)
		t1 := float32(i+1) / float32(n)
		// 相位**減**索引:段號往前 = 顏色往後,看起來就是顏色朝終點流。
		c := ramp[((phase-i)%len(ramp)+len(ramp))%len(ramp)]
		// ⚠ **不開反鋸齒**:原版是逐像素寫調色盤索引,畫出來是硬邊的實心像素。
		// 開了反鋸齒會把本來就很暗的綠(見下方註解)再和黑底混一次,線幾乎就看不見了。
		vector.StrokeLine(dst,
			x1+dx*t0, y1+dy*t0, x1+dx*t1, y1+dy*t1, 1, c, false)
	}
}

func absF(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// ---- 設定集結點 ----

// relocatePickState 是「正在挑集結點」的畫面狀態(同 F9 測距,是**看的方式**不進存檔)。
//
// **兩段式**,對齊原版 `Star_Relocation_` @ 0x75180:先點起點星(必須是自己的殖民地),
// 再點終點星;點到同一顆 = 取消。from = −1 表示還在等起點。
type relocatePickState struct {
	on   bool
	from int
	// all = 這一輪是「ALL」:點到的星會把**已經有集結點**的殖民地全部改送過去
	// (原版 Set_All_Star_Relocations_,見 shell 那支的 ⚠——沒設過的不會被順便設上)。
	all bool
}

func relocationRetargetAllTextRect() textSafeRect {
	return textSafeRect{x: 20, y: 412, w: 140, h: 18, insetX: 4, insetY: 1}
}

func relocationClearAllTextRect() textSafeRect {
	return textSafeRect{x: 168, y: 412, w: 140, h: 18, insetX: 4, insetY: 1}
}

func fleetAntaranEntryTextRect() textSafeRect {
	return textSafeRect{x: 20, y: 392, w: 288, h: 18, insetX: 4, insetY: 1}
}

// beginRelocatePick 從**起點**開始挑(艦隊列表的 RELOCATE 鈕走這條,同原版)。
func (b *sceneBuilder) beginRelocatePick() {
	b.relocPick.on, b.relocPick.from, b.relocPick.all = true, -1, false
}

// beginRelocateAll 進「一次改全部」模式(艦隊列表的 ALL 鈕)。
func (b *sceneBuilder) beginRelocateAll() {
	b.relocPick.on, b.relocPick.from, b.relocPick.all = true, -1, true
}

// beginRelocatePickFrom 起點已知時直接進第二段(星圖面板的「設定集結點」鈕走這條)。
//
// ⚠ 這是 remake 多出來的捷徑:玩家已經在星資訊面板上選著那顆殖民星了,
// 再叫他點一次自己是多餘的。規則面走的仍是同一支 `SetStarRelocation`。
func (b *sceneBuilder) beginRelocatePickFrom(star int) {
	b.relocPick.on, b.relocPick.from = true, star
}

func relocationMonsterName(b *sceneBuilder, sess *shell.GameSession, star int) string {
	if b == nil || sess == nil {
		return ""
	}
	seen := map[int]bool{}
	names := make([]string, 0, 1)
	for _, monster := range sess.MonsterGroupsAtStar(star) {
		kind := int(monster.Kind)
		if seen[kind] {
			continue
		}
		seen[kind] = true
		names = append(names, uiText(b.lang, fmt.Sprintf("monster.name.%d", kind)))
	}
	return strings.Join(names, uiText(b.lang, "list.separator"))
}

func relocationRefusalText(b *sceneBuilder, sess *shell.GameSession, star int, refusal shell.RelocateRefusal) string {
	if b == nil {
		return ""
	}
	switch refusal {
	case shell.RelocateInvalidStar:
		return uiText(b.lang, "relocation.refusal.invalid_star")
	case shell.RelocateBlackHoleOrigin:
		return uiText(b.lang, "relocation.refusal.black_hole_origin")
	case shell.RelocateBlackHoleTarget:
		return uiText(b.lang, "relocation.refusal.black_hole_target")
	case shell.RelocateUnexploredOrigin:
		return uiText(b.lang, "relocation.refusal.unexplored_origin")
	case shell.RelocateUnexploredTarget:
		return uiText(b.lang, "relocation.refusal.unexplored_target")
	case shell.RelocateMonsterOrigin:
		return fmt.Sprintf(uiText(b.lang, "relocation.refusal.monster_origin"), relocationMonsterName(b, sess, star))
	case shell.RelocateNoColony:
		name := ""
		if sess != nil && star >= 0 && star < len(sess.Stars) {
			name = sess.Stars[star].Name
		}
		return fmt.Sprintf(uiText(b.lang, "relocation.refusal.no_colony"), name)
	case shell.RelocateWriteFailed:
		return uiText(b.lang, "relocation.refusal.write_failed")
	default:
		return ""
	}
}

func relocationConfirmText(b *sceneBuilder, sess *shell.GameSession, star int) string {
	if b == nil || sess == nil || star < 0 || star >= len(sess.Stars) {
		return ""
	}
	return fmt.Sprintf(uiText(b.lang, "relocation.confirm.monster"),
		sess.Stars[star].Name, relocationMonsterName(b, sess, star))
}

// relocatePickClickedStar 在挑集結點模式下吃掉一次點星,回傳是否已消化。
func (b *sceneBuilder) relocatePickClickedStar(star int) bool {
	if !b.relocPick.on || b.session == nil {
		return false
	}
	sess := b.session
	if b.relocPick.all { // ALL:一次把已經有集結點的全部改送到這顆
		if r := sess.CanRelocateTo(star); r != shell.RelocateAllowed {
			b.flash(relocationRefusalText(b, sess, star, r))
			return true
		}
		n := sess.SetAllStarRelocations(star)
		b.relocPick.on, b.relocPick.all = false, false
		if n == 0 {
			b.flash(uiText(b.lang, "relocation.result.no_existing"))
		} else {
			b.flash(fmt.Sprintf(uiText(b.lang, "relocation.result.retargeted_count"), n))
		}
		return true
	}
	if b.relocPick.from < 0 { // 第一段:選起點
		if r := sess.CanRelocateFrom(star); r != shell.RelocateAllowed {
			b.flash(relocationRefusalText(b, sess, star, r))
			return true // 仍算消化掉:這一下是模式的輸入,不該同時去選星
		}
		b.relocPick.from = star
		b.flash(uiText(b.lang, "relocation.prompt.origin_set"))
		return true
	}
	// 第二段:選終點。
	from := b.relocPick.from
	// 原版對「終點被怪獸盤據」是**問一句**不是拒絕(`Okay_To_Set_Relocate_Star_` 走
	// `User_Box_(kind=1)`);玩家說是就照設。⚠ 上面的 ALL 分支沒有這一問——
	// 原版的 `Set_All_Star_Relocations_` 是一支沒有任何驗證的迴圈,不替它加規則。
	if sess.RelocateToNeedsConfirm(star) {
		b.relocPick.on = false
		b.pendingConfirm = &pendingConfirm{
			msg: relocationConfirmText(b, sess, star),
			onYes: func() *origTransition {
				b.applyRelocation(from, star)
				return nil // 回到下層的星圖
			},
		}
		return true
	}
	if r := sess.SetStarRelocation(from, star); r != shell.RelocateAllowed {
		b.flash(relocationRefusalText(b, sess, star, r))
		return true
	}
	b.relocPick.on = false
	ci := colonyIndexAtStar(sess, from)
	if to := sess.ColonyRelocation(ci); to == shell.ColonyRelocationNone {
		b.flash(uiText(b.lang, "relocation.result.cleared"))
	} else {
		b.flash(uiText(b.lang, "relocation.result.set"))
	}
	return true
}

// applyRelocation 落地一次集結點設定並回報結果(確認框按「是」之後走這支)。
func (b *sceneBuilder) applyRelocation(from, to int) {
	sess := b.session
	if sess == nil {
		return
	}
	if r := sess.SetStarRelocation(from, to); r != shell.RelocateAllowed {
		b.flash(relocationRefusalText(b, sess, to, r))
		return
	}
	ci := colonyIndexAtStar(sess, from)
	if t := sess.ColonyRelocation(ci); t == shell.ColonyRelocationNone {
		b.flash(uiText(b.lang, "relocation.result.cleared"))
		return
	}
	b.flash(uiText(b.lang, "relocation.result.set"))
}

// colonyIndexAtStar 回傳玩家在某顆星的殖民地索引(沒有回 −1)。
func colonyIndexAtStar(sess *shell.GameSession, star int) int {
	if sess == nil {
		return -1
	}
	for i, st := range sess.PlayerColonyStars {
		if st == star {
			return i
		}
	}
	return -1
}

// galleryRelocTarget 是截圖廊示範用的集結點星。
//
// ⚠ 刻意挑**第二顆**可見的星:第一顆已經被 F9 測距的示範用掉了,兩條線重疊的話
// 遷移連線會整條藏在測距線底下——截圖看起來像沒做(第一版就是這樣)。
func (b *sceneBuilder) galleryRelocTarget() int {
	if b.session == nil {
		return -1
	}
	vis := b.session.VisibleStars()
	seen := 0
	for i := range b.session.Stars {
		if i == galleryMeasureFrom {
			continue
		}
		if vis != nil && i < len(vis) && !vis[i] {
			continue
		}
		seen++
		if seen == 2 {
			return i
		}
	}
	return -1
}

// --- 星圖星系面板的拓殖/前哨站鈕(hits 與繪製共用同一份佈局)---

// starPanelRow 是面板上的一列動作鈕:螢幕 y、動作碼、以及這一下會落在哪顆天體(−1=不適用)。
type starPanelRow struct {
	y      int
	action string
	planet int
}

// starPanelColonyRows 算出選中星要畫哪幾顆拓殖/前哨站鈕。
//
// 為什麼要抽出來:先前 hits 與繪製各寫了一份「有殖民船 → 402,有前哨船 → 下一列」的判斷,
// 兩邊只要有一邊改了就會出現「畫得出來卻點不到」。
//
// 版面限制(面板 326..458 高 132):只有 402 與 424 兩列。而**自己的殖民地星系** 424 那列
// 已經被集結點鈕佔走(見繪製處),所以那種情況只畫得下一顆——依原版的擴張順序,
// 殖民地優先於前哨站。
func starPanelColonyRows(sess *shell.GameSession) []starPanelRow {
	if sess == nil || sess.SelectedStar < 0 || sess.SelectedStar >= len(sess.Stars) {
		return nil
	}
	star := sess.SelectedStar
	if sess.Stars[star].Owner == 2 || sess.Fleet().AtStar != star || sess.Fleet().ETA != 0 {
		return nil
	}
	if sess.StarGuardedByMonster(star) {
		return nil // 怪獸盤據:那一列是「挑戰」鈕
	}
	var out []starPanelRow
	if sess.FleetHasColonyShip() {
		if p := sess.FirstColonizablePlanet(star); p >= 0 {
			out = append(out, starPanelRow{y: 402, action: "colonize", planet: p})
		}
	}
	if sess.FleetHasOutpostShip() {
		if p := sess.OutpostTargetPlanet(star); p >= 0 {
			out = append(out, starPanelRow{y: 402, action: "outpost", planet: p})
		}
	}
	if len(out) == 2 {
		out[1].y = 424
	}
	// 集結點鈕佔著 424:自己的殖民地星系只留第一顆。
	if colonyIndexAtStar(sess, star) >= 0 && len(out) > 1 {
		out = out[:1]
	}
	return out
}

// galleryConfirmMessage 是截圖廊要顯示在確認框裡的那句話。
//
// 優先用真的規則產出的訊息(找一顆被怪獸盤據的星,走 RelocateToNeedsConfirm)——
// 截圖要驗的是「這句話長什麼樣、放不放得下」,拿一句寫死的假字驗不到折行。
// 這一局沒有怪獸時退回一句等長的示意文字。
func (b *sceneBuilder) galleryConfirmMessage() string {
	if sess := b.session; sess != nil {
		for i := range sess.Stars {
			if sess.RelocateToNeedsConfirm(i) {
				return relocationConfirmText(b, sess, i)
			}
		}
	}
	return fmt.Sprintf(uiText(b.lang, "relocation.confirm.monster"),
		uiText(b.lang, "relocation.fallback.system"), uiText(b.lang, "relocation.fallback.monster"))
}

// --- 艦隊列表的 ALL 鈕:全選 / 全不選(手冊 p.32 + p.47)---

// initializeFleetShipSelection 依 Auto Select Ships 設定建立「新進入／新切換艦隊」的
// 初始選取集合。呼叫端只可在 shipPick 尚未初始化時使用；空但非 nil 的 map 代表玩家已
// 手動全不選，重繪畫面時不得再替玩家選回來。
func (b *sceneBuilder) initializeFleetShipSelection() {
	if b == nil || b.session == nil || b.shipPick != nil {
		return
	}
	b.shipPick = map[int]bool{}
	if !b.session.EffectiveGameSettings().AutoSelectShips {
		return
	}
	f := b.session.Fleet()
	if f == nil {
		return
	}
	for i := range f.Ships {
		b.shipPick[i] = true
	}
}

// toggleSelectAllShips 全選這支艦隊的艦艇;已經全選就變成全不選。
//
// 手冊 p.47 逐字:「All: Selects all of the ships in the fleet to prepare to receive orders.
// (If all the ships are already selected, this deselects them instead.)」
// 括號那句是 toggle 語意,不是「按一次全選、再按一次還是全選」。
//
// ⚠ 2026-08-07 訂正:這顆鈕先前被接成 `Set_All_Star_Relocations_`,那是**推測**且推錯了
// (見 interactive.go 那塊 hits 的註解)。選取狀態本來就有——分艦隊用的就是它
// (`sceneBuilder.shipPick`),所以這顆鈕接上去之後,「全選 → 分艦隊」變成兩下就做得完。
func (b *sceneBuilder) toggleSelectAllShips() {
	if b.session == nil {
		return
	}
	f := b.session.Fleet()
	if f == nil {
		return
	}
	if b.shipPick == nil {
		b.shipPick = map[int]bool{}
	}
	allOn := len(f.Ships) > 0
	for i := range f.Ships {
		if !b.shipPick[i] {
			allOn = false
			break
		}
	}
	b.shipPick = map[int]bool{}
	if !allOn {
		for i := range f.Ships {
			b.shipPick[i] = true
		}
	}
}
