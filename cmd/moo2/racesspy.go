package main

import (
	"strconv"
	"strings"
)

// racesspy.go:種族關係畫面上**間諜派遣區塊**的版面座標。
//
// ============ 為什麼沒有「間諜畫面」 ============
//
// WORKLIST 上「spy UI」那一項的原始規劃是「做一張間諜畫面」。**原版沒有那張畫面。**
//
// 反組譯搜遍 `Spy_Screen` / `Espionage_Screen` 等關鍵字**零命中**——間諜的任務指派內嵌在
// `Race_Screen_` @ 0x10ACBA 這張種族關係畫面裡:每個已接觸種族一列,列上有關係滑桿,
// 列旁邊有派間諜的按鈕。照舊規劃去做,方向從一開始就錯。
//
// 這是負面結論的價值:它省掉的是一整張不存在的畫面。
// (詳見 `docs/re/screen-coords-spy-leader.md` §3.1。)
//
// ============ 座標是執行檔立即數,不是量圖 ============
//
// 這張畫面在 openorion2 是 STUB(沒有硬編座標),所以 remake 先前只能量圖。
// 現在有更上游的來源:`.data` 段的兩張靜態陣列(`Orion2.exe.asm` 554330–554472),
// 逐位元組解出來的。
//
// 版面是 **左欄 4 列 + 右欄 3 列 = 7 個已接觸種族**,兩欄 y 完全對齊——
// 那正是 MOO2 最多同時顯示 7 個對手的版面。remake 先前是自編的「y 從 66 起、每列 +62」。

// racesMaxRows 是這張畫面放得下的種族數(版面決定的,不是遊戲規則)。
const racesMaxRows = 7

// racesSpyAnchors 是每個已接觸種族那一列的任務鈕錨點
// (`_race_spy_btns` @ 0x18406F,逐位元組解出)。左欄 i0-3 x=120、右欄 i4-6 x=332。
var racesSpyAnchors = [racesMaxRows][2]int{
	{120, 126}, {120, 233}, {120, 338}, {120, 445},
	{332, 126}, {332, 233}, {332, 338},
}

// racesRelationBars 是關係滑桿軌道的貼圖位置(`_race_bar` @ 0x18400D)。
var racesRelationBars = [racesMaxRows][2]int{
	{105, 49}, {105, 155}, {105, 262}, {105, 370},
	{528, 49}, {528, 156}, {528, 261},
}

// racesSpyButtonW / racesSpyButtonH 是「派間諜」鈕的點擊範圍。
//
// ⚠ **寬高不是查來的。** 執行檔那邊的寬高欄位是 LBX 資產控制碼、不是字面尺寸。
// 已知的約束是三顆鈕的 x 偏移為 0 / +76 / +149,相鄰間距 76,所以單顆不會寬過那個數。
// 取 68×20(與軍官畫面 HIRE 鈕同量級)是 **remake 的選擇,不是真值**。
const racesSpyButtonW, racesSpyButtonH = 68, 20

// ============ 只做「說得出意思」的那一顆 ============
//
// ⚠ 原版每個種族有**三顆**任務鈕(錨點 x 偏移 0 / +76 / +149)。反組譯只看得出
// 「同一列建三顆」,**沒有查到哪一顆對應哪個任務**——`Adjust_Spy_Mission_Data_` 只看到
// 把任務欄位從 0 設成預設值 3,三顆各自的目標值沒追到。
//
// 手冊的書寫順序是 Espionage→Sabotage→Hide,但**書寫順序不保證等於 UI 由左到右的順序**。
//
// 所以這裡**只做最左邊那一顆**,而且只掛 remake 真的有模型的那一件事(派間諜偷科技,
// 見 internal/shell/spy.go——破壞與隱匿都沒有規則可依)。另外兩顆的座標記在 docs/re 裡
// 等順序查明。**畫一顆意思說不出來的按鈕比不畫更糟**:玩家點下去不知道會發生什麼。

// racesSpyAction 是第 i 個種族那顆「派間諜」鈕的動作字串。
func racesSpyAction(i int) string { return "spy" + strconv.Itoa(i) }

// racesSpyActionIndex 把動作字串解回種族索引;不是間諜動作時回 (0, false)。
func racesSpyActionIndex(action string) (int, bool) {
	if !strings.HasPrefix(action, "spy") {
		return 0, false
	}
	i, err := strconv.Atoi(strings.TrimPrefix(action, "spy"))
	if err != nil || i < 0 || i >= racesMaxRows {
		return 0, false
	}
	return i, true
}

// racesSpyHitRegions 回傳前 n 個種族的「派間諜」鈕熱區(n 夾在 racesMaxRows)。
func racesSpyHitRegions(n int) []hitRegion {
	if n > racesMaxRows {
		n = racesMaxRows
	}
	out := make([]hitRegion, 0, n)
	for i := 0; i < n; i++ {
		a := racesSpyAnchors[i]
		out = append(out, hitRegion{a[0], a[1], racesSpyButtonW, racesSpyButtonH, racesSpyAction(i)})
	}
	return out
}

// aiCount 回傳目前的 AI 對手數(無對局時 0)。
func (b *sceneBuilder) aiCount() int {
	if b.session == nil {
		return 0
	}
	return len(b.session.AIPlayers)
}
