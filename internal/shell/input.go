package shell

// input.go:輸入狀態與按鈕命中判定(純邏輯,可單測;ebiten 端把滑鼠/鍵盤填成 InputState)。

// InputState 是某一幀的輸入快照。ClickReleased 表示這一幀有一次「放開左鍵」的點擊(邊緣觸發,
// 避免按住連觸);由 ebiten 端用 inpututil.IsMouseButtonJustReleased 之類填入。
type InputState struct {
	MouseX, MouseY int
	ClickReleased  bool
	// Hotkey 是這一幀剛按下的快捷鍵名(如 "F1"、"F9"),沒有按就是空字串。
	// 同樣是邊緣觸發:按住不會連發。
	//
	// 用字串不用 ebiten 的 key 型別,是為了讓規則層不相依 ebiten——
	// headless 腳本(截圖廊、測試)因此也能直接注入按鍵,不必真的有鍵盤。
	Hotkey string
}

// 原版的星圖快捷鍵(手冊行文中直接寫死的那幾個;見 starnav.go 檔頭)。
//
// ⚠ 手冊另有一組 ALT+F1..F8 的設定開關,但那些鍵在 PDF 裡是**右側邊欄標籤**,
// 抽出來的文字流會排到前一個選項的描述尾巴,對應關係有 off-by-one 的風險(中間還缺 F4)。
// **沒有對原版逐項確認之前不要加**。
const (
	HotkeyNextFleet   = "F1"  // 循環到下一支已知艦隊
	HotkeyPrevFleet   = "F2"  // 循環到上一支
	HotkeyNextColony  = "F5"  // 下一個已殖民星系
	HotkeyPrevColony  = "F6"  // 上一個已殖民星系
	HotkeyMeasureDist = "F9"  // 測距:點第一顆星,游標移到另一顆看秒差距
	HotkeyQuickSave   = "F10" // 快速存檔(沿用上次的存檔名)
)

// Button 是一個矩形按鈕(左上 X,Y + 寬高),ID 供邏輯辨識、Label 供顯示。
type Button struct {
	X, Y, W, H int
	ID         string
	Label      string
}

// Hit 回傳座標 (x,y) 是否落在按鈕內。
func (b Button) Hit(x, y int) bool {
	return x >= b.X && x < b.X+b.W && y >= b.Y && y < b.Y+b.H
}

// ClickedButton 回傳這一幀被點擊(放開)的按鈕 ID;沒有點擊或沒命中回空字串。
func ClickedButton(btns []Button, in InputState) string {
	if !in.ClickReleased {
		return ""
	}
	for _, b := range btns {
		if b.Hit(in.MouseX, in.MouseY) {
			return b.ID
		}
	}
	return ""
}

// HoveredButton 回傳滑鼠目前懸停的按鈕 ID(供 highlight);沒命中回空字串。
func HoveredButton(btns []Button, in InputState) string {
	for _, b := range btns {
		if b.Hit(in.MouseX, in.MouseY) {
			return b.ID
		}
	}
	return ""
}
