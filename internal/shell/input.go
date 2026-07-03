package shell

// input.go:輸入狀態與按鈕命中判定(純邏輯,可單測;ebiten 端把滑鼠/鍵盤填成 InputState)。

// InputState 是某一幀的輸入快照。ClickReleased 表示這一幀有一次「放開左鍵」的點擊(邊緣觸發,
// 避免按住連觸);由 ebiten 端用 inpututil.IsMouseButtonJustReleased 之類填入。
type InputState struct {
	MouseX, MouseY int
	ClickReleased  bool
}

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
