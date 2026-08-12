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

// 每列有三個原版按鈕槽；錨點是第一槽，後兩槽的 x 偏移來自同一份執行檔靜態表。
//
// ⚠ 寬高不是反組譯真值：執行檔欄位是 LBX 控制碼，不能當成字面尺寸。寬度依
// RACES.LBX 畫面的可見槽位量測，最後一槽在左欄必須在中央立柱前收束。把控制
// 文字放回這三槽，不能再往下加一列，否則會壓到下一個帝國資料列。
const (
	racesSpyButtonW, racesSpyButtonH = 68, 20
	racesSpyButtonSlots              = 3
)

var (
	racesSpySlotOffsets = [racesSpyButtonSlots]int{0, 76, 149}
	racesSpySlotWidths  = [racesSpyButtonSlots]int{68, 68, 48}
)

// racesSpySlotRect 回傳第 i 個帝國列、第 slot 個原生按鈕槽的矩形。繪製與命中
// 共用它，避免譯文位置再漂進下一列。
func racesSpySlotRect(i, slot int) (x, y, w, h int) {
	anchor := racesSpyAnchors[i]
	return anchor[0] + racesSpySlotOffsets[slot], anchor[1], racesSpySlotWidths[slot], racesSpyButtonH
}

// 防守 Agent 是 RACES 畫面上進攻間諜之外的第二條管理線。座標放在左下方
// 的空白區，不覆蓋原版外交按鈕或七列種族資料；尺寸是 remake 控制項尺寸。
const (
	racesAgentTrainX, racesAgentDismissX  = 20, 128
	racesAgentY, racesAgentW, racesAgentH = 414, 100, 20
	racesAgentStatusX, racesAgentStatusY  = 20, 396
	racesAgentStatusW, racesAgentStatusH  = 208, 16
)

// racesDiplomacyRowW / racesDiplomacyRowH 是 remake 在每個已顯示 AI 名稱列上的
// 「對談」熱區。原版 RACES 的列資料與關係滑桿座標已知，但單列 REPORT 選取語意
// 未由反組譯完整追出；因此這是明確標籤的 remake 入口，不冒充原版欄位。
const racesDiplomacyRowW, racesDiplomacyRowH = 90, 38
const racesDiplomacyRightRowW = 180

// ============ 只做「說得出意思」的明確控制 ============
//
// ⚠ 原版每個種族有**三顆**任務鈕(錨點 x 偏移 0 / +76 / +149)。反組譯只看得出
// 「同一列建三顆」,**沒有查到哪一顆對應哪個任務**——`Adjust_Spy_Mission_Data_` 只看到
// 把任務欄位從 0 設成預設值 3,三顆各自的目標值沒追到。
//
// 手冊的書寫順序是 Espionage→Sabotage→Hide,但**書寫順序不保證等於 UI 由左到右的順序**。
//
// 所以這裡不把三個未解座標硬對應到任務，而是在原版任務鈕列下提供一個明確標籤的
// remake 控制；三種任務的最小結算規則見 internal/shell/spy.go。原版左右順序仍記在
// docs/re 研究紀錄裡等待新證據，不影響這個可讀的 remake 入口。

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
		x, y, w, h := racesSpySlotRect(i, 0)
		out = append(out, hitRegion{x, y, w, h, racesSpyAction(i)})
	}
	return out
}

// racesSpyMissionAction 是種族關係畫面中循環切換 STEAL/SABOTAGE/HIDE 的 remake 操作。
// 它使用原生列的第二槽，而不是過去會壓到下一列的 y+23 假按鈕。
func racesSpyMissionAction(i int) string { return "spymission" + strconv.Itoa(i) }

// racesSpyMissionActionIndex 把切換任務動作解回種族索引。
func racesSpyMissionActionIndex(action string) (int, bool) {
	const prefix = "spymission"
	if !strings.HasPrefix(action, prefix) {
		return 0, false
	}
	i, err := strconv.Atoi(strings.TrimPrefix(action, prefix))
	if err != nil || i < 0 || i >= racesMaxRows {
		return 0, false
	}
	return i, true
}

// racesSpyMissionHitRegions 回傳前 n 個種族的 remake 任務切換熱區(第二槽)。
func racesSpyMissionHitRegions(n int) []hitRegion {
	if n > racesMaxRows {
		n = racesMaxRows
	}
	out := make([]hitRegion, 0, n)
	for i := 0; i < n; i++ {
		x, y, w, h := racesSpySlotRect(i, 1)
		out = append(out, hitRegion{x, y, w, h, racesSpyMissionAction(i)})
	}
	return out
}

// racesSpyHideAction 是第三槽的明確「隱匿」快捷鍵。這是 remake 的可讀控制，
// 不把未完整追出的原版 callback 細節冒充為已證實行為。
func racesSpyHideAction(i int) string { return "spyhide" + strconv.Itoa(i) }

func racesSpyHideActionIndex(action string) (int, bool) {
	const prefix = "spyhide"
	if !strings.HasPrefix(action, prefix) {
		return 0, false
	}
	i, err := strconv.Atoi(strings.TrimPrefix(action, prefix))
	if err != nil || i < 0 || i >= racesMaxRows {
		return 0, false
	}
	return i, true
}

// racesSpyHideHitRegions 回傳前 n 個種族的「隱匿」快捷鍵(第三槽)。
func racesSpyHideHitRegions(n int) []hitRegion {
	if n > racesMaxRows {
		n = racesMaxRows
	}
	out := make([]hitRegion, 0, n)
	for i := 0; i < n; i++ {
		x, y, w, h := racesSpySlotRect(i, 2)
		out = append(out, hitRegion{x, y, w, h, racesSpyHideAction(i)})
	}
	return out
}

func racesAgentTrainAction() string   { return "trainagent" }
func racesAgentDismissAction() string { return "dismissagent" }

func racesAgentHitRegions() []hitRegion {
	return []hitRegion{
		{racesAgentTrainX, racesAgentY, racesAgentW, racesAgentH, racesAgentTrainAction()},
		{racesAgentDismissX, racesAgentY, racesAgentW, racesAgentH, racesAgentDismissAction()},
	}
}

// racesDiplomacyAction 是 RACES 畫面逐對手進入外交對談的 remake 操作。
func racesDiplomacyAction(i int) string { return "racediplomacy" + strconv.Itoa(i) }

// racesDiplomacyActionIndex 把逐對手外交動作解回種族索引。
func racesDiplomacyActionIndex(action string) (int, bool) {
	const prefix = "racediplomacy"
	if !strings.HasPrefix(action, prefix) {
		return 0, false
	}
	i, err := strconv.Atoi(strings.TrimPrefix(action, prefix))
	if err != nil || i < 0 || i >= racesMaxRows {
		return 0, false
	}
	return i, true
}

// racesDiplomacyHitRegions 回傳前 n 個種族列的逐對手外交熱區。
func racesDiplomacyHitRegions(n int) []hitRegion {
	if n > racesMaxRows {
		n = racesMaxRows
	}
	out := make([]hitRegion, 0, n)
	for i := 0; i < n; i++ {
		bar := racesRelationBars[i]
		x := 20
		w := racesDiplomacyRowW
		if i >= 4 {
			x = 440
			w = racesDiplomacyRightRowW
		}
		out = append(out, hitRegion{x, bar[1], w, racesDiplomacyRowH, racesDiplomacyAction(i)})
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
