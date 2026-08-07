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

// relocatePickClickedStar 在挑集結點模式下吃掉一次點星,回傳是否已消化。
func (b *sceneBuilder) relocatePickClickedStar(star int) bool {
	if !b.relocPick.on || b.session == nil {
		return false
	}
	sess := b.session
	if b.relocPick.all { // ALL:一次把已經有集結點的全部改送到這顆
		if r := sess.CanRelocateTo(star); r != "" {
			b.flash(b.tr(string(r), string(r)))
			return true
		}
		n := sess.SetAllStarRelocations(star)
		b.relocPick.on, b.relocPick.all = false, false
		if n == 0 {
			b.flash(b.tr("沒有任何殖民地設過集結點——ALL 只改已經設過的",
				"No colony has a rally point — ALL only retargets existing ones"))
		} else {
			b.flash(fmt.Sprintf(b.tr("已把 %d 個集結點改到這裡", "Retargeted %d rally points here"), n))
		}
		return true
	}
	if b.relocPick.from < 0 { // 第一段:選起點
		if r := sess.CanRelocateFrom(star); r != "" {
			b.flash(b.tr(string(r), string(r)))
			return true // 仍算消化掉:這一下是模式的輸入,不該同時去選星
		}
		b.relocPick.from = star
		b.flash(b.tr("起點已選——再點一顆星當集結點(點回自己就取消)",
			"Origin set — now click the rally star (click it again to clear)"))
		return true
	}
	// 第二段:選終點。
	from := b.relocPick.from
	// 原版對「終點被怪獸盤據」是**問一句**不是拒絕(`Okay_To_Set_Relocate_Star_` 走
	// `User_Box_(kind=1)`);玩家說是就照設。⚠ 上面的 ALL 分支沒有這一問——
	// 原版的 `Set_All_Star_Relocations_` 是一支沒有任何驗證的迴圈,不替它加規則。
	if msg := sess.RelocateToNeedsConfirm(star); msg != "" {
		b.relocPick.on = false
		b.pendingConfirm = &pendingConfirm{
			msg: b.tr(msg, msg),
			onYes: func() *origTransition {
				b.applyRelocation(from, star)
				return nil // 回到下層的星圖
			},
		}
		return true
	}
	if r := sess.SetStarRelocation(from, star); r != "" {
		b.flash(b.tr(string(r), string(r)))
		return true
	}
	b.relocPick.on = false
	ci := colonyIndexAtStar(sess, from)
	if to := sess.ColonyRelocation(ci); to == shell.ColonyRelocationNone {
		b.flash(b.tr("已取消集結點", "Relocation cleared"))
	} else {
		b.flash(b.tr("集結點已設定——新造的艦會自動送過去",
			"Relocation set — new ships will travel there"))
	}
	return true
}

// applyRelocation 落地一次集結點設定並回報結果(確認框按「是」之後走這支)。
func (b *sceneBuilder) applyRelocation(from, to int) {
	sess := b.session
	if sess == nil {
		return
	}
	if r := sess.SetStarRelocation(from, to); r != "" {
		b.flash(b.tr(string(r), string(r)))
		return
	}
	ci := colonyIndexAtStar(sess, from)
	if t := sess.ColonyRelocation(ci); t == shell.ColonyRelocationNone {
		b.flash(b.tr("已取消集結點", "Relocation cleared"))
		return
	}
	b.flash(b.tr("集結點已設定——新造的艦會自動送過去",
		"Relocation set — new ships will travel there"))
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
			if msg := sess.RelocateToNeedsConfirm(i); msg != "" {
				return msg
			}
		}
	}
	return b.tr("這個星系被太空怪獸盤據,送過去的艦艇會遭到攻擊。仍要把集結點設在那裡嗎?",
		"That system is guarded by a space monster and ships sent there will be attacked. Set the rally point anyway?")
}
