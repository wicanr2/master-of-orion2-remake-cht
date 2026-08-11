package main

// officerscreen.go:軍官畫面的**版面座標**。
//
// 座標一律取自原版執行檔 `Add_Officer_Screen_Fields_` @ 0x9264E 的立即數
// (`docs/re/screen-coords-spy-leader.md`),不是 openorion2、不是量圖。
//
// 抽出來當套件層級的資料而不是內嵌在 `officer()` 裡,理由有二:
// 一是**可測**——建畫面要真 LBX 資產,測試環境沒有;二是讓「這些數字是查來的」
// 這件事在程式碼結構上看得見,不會被當成隨手填的版面微調。
//
// ============ 與先前那組的差異 ============
//
//	| 元素 | 先前(openorion2 / PIL 量測)| 執行檔立即數 |
//	|---|---|---|
//	| Colony Leaders 分頁 | x=20 | **x=9** |
//	| Ship Leaders 分頁 | x=166 | **x=156** |
//	| HIRE | (313, 440) | (313, **441**) |
//	| POOL | (388, 440) | (388, **441**) |
//	| DISMISS | (462, 440) | (**463**, **441**) |
//	| RETURN | (540, 440) | (**538**, **441**) |
//	| 清單列中心 | 90/199/308/417 | **88/197/306/415** |
//	| 上下捲鈕 | **沒有** | (613, 22) / (613, 170) |
//
// 依專案的來源優先序(**反組譯立即數 > 手冊 > openorion2 > LBX 尺寸 > 量圖**),
// 以執行檔為準。
//
// ⚠ **寬高全部維持原樣。** 執行檔那邊的寬高欄位是 LBX 資產控制碼、不是字面尺寸,
// 所以沒查到——**沒查到的就不動**,不拿「看起來差不多」的數字去覆蓋既有值。

// officerButtonY 是底部那排按鈕的共同 y(執行檔立即數 `mov edx, 1B9h`)。
const officerButtonY = 441

// officerButtonSpacing 是 HIRE→POOL→DISMISS→RETURN 的等距 x 間距。
// 313 / 388 / 463 / 538——這個規律本身是解碼正確的內部佐證。
const officerButtonSpacing = 75

// officerOverlays 回傳擦底疊字的中文標籤框。
func officerOverlays() []labelRect {
	return []labelRect{
		{9, 11, 133, 20, "Colony Leaders", 0},
		{156, 11, 124, 20, "Ship Officers", 0},
		{313, officerButtonY, 68, 20, "HIRE", 0},
		{313 + officerButtonSpacing, officerButtonY, 69, 20, "POOL", 0},
		{313 + 2*officerButtonSpacing, officerButtonY, 74, 20, "DISMISS", 0},
		{313 + 3*officerButtonSpacing, officerButtonY, 80, 20, "RETURN", 0},
	}
}

// officerHitRegions 回傳可點擊熱區。
//
// 上下捲鈕(`_officer_up_button_seg` / `_officer_dn_button_seg`)先前 remake
// **完全沒有**——所以清單超過四列就看不到後面的人。寬高未查,取與其他按鈕同量級的 20×20。
func officerHitRegions() []hitRegion {
	return []hitRegion{
		{9, 11, 133, 20, "colonyTab"},
		{156, 11, 124, 20, "shipTab"},
		{313 + 3*officerButtonSpacing, officerButtonY, 80, 20, "Return"},
		{313, officerButtonY, 68, 20, "hire"},
		{313 + officerButtonSpacing, officerButtonY, 69, 20, "pool"},
		{313 + 2*officerButtonSpacing, officerButtonY, 74, 20, "dismiss"},
		{613, 22, 20, 20, "scrollUp"},
		{613, 170, 20, 20, "scrollDown"},
	}
}

// officerRowCenters 回傳軍官清單四列的中心 y。
//
// 執行檔那個迴圈建立的熱區 y 範圍是 34–142 / 143–251 / 252–360 / 361–469
// ——列起點 34、高 108、**列距 109**,中心即 88/197/306/415。
//
// 先前的 90/199/308/417 照 openorion2 的 `FIRST_ROW 38 + SLOT_HEIGHT 105/2` 推,
// 列距 109 對上了(那一半 openorion2 是對的),起點差 2px。
func officerRowCenters() []float64 {
	const (
		first = 88
		pitch = 109
	)
	out := make([]float64, 4)
	for i := range out {
		out[i] = float64(first + i*pitch)
	}
	return out
}

// officerRowPrefix 只用文字標出目前的管理選取項。
// 原版亮框的繪製控制碼尚未從資產／執行檔資料中解出，先保留可見且不改底圖的提示。
func officerRowPrefix(selected bool, unselected string) string {
	if selected {
		return "▶ "
	}
	return unselected
}

// officerTargetShip 取得軍官畫面的指派目標。
//
// 艦隊畫面點擊艦艇列會留下 shipPick;選了多艘時取最低索引,保證結果可重現。
// 沒有勾選時退回目前艦隊第一艘,讓「艦隊 → LEADERS → 點軍官」仍是可完成的正常路徑。
func (b *sceneBuilder) officerTargetShip() (fleetIndex, shipIndex int, ok bool) {
	if b.session == nil || b.session.SelectedFleet < 0 || b.session.SelectedFleet >= len(b.session.Fleets) {
		return -1, -1, false
	}
	fleetIndex = b.session.SelectedFleet
	f := &b.session.Fleets[fleetIndex]
	for i := range f.Ships {
		if b.shipPick != nil && b.shipPick[i] {
			return fleetIndex, i, true
		}
	}
	if len(f.Ships) == 0 {
		return -1, -1, false
	}
	return fleetIndex, 0, true
}

// officerTargetColony 取得殖民地領袖畫面的指派目標。從殖民地畫面進入時保留
// b.colonyIdx；從星圖／全域軍官畫面進入時退回選中星，再退回母星。
func (b *sceneBuilder) officerTargetColony() (int, bool) {
	if b.session == nil {
		return -1, false
	}
	if b.colonyIdx >= 0 && b.colonyIdx < len(b.session.PlayerColonies) {
		return b.colonyIdx, true
	}
	if idx := colonyIndexAtStar(b.session, b.session.SelectedStar); idx >= 0 {
		return idx, true
	}
	if len(b.session.PlayerColonies) > 0 {
		return 0, true
	}
	return -1, false
}
