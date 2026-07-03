package shell

import "testing"

func TestButtonHitAndClick(t *testing.T) {
	btns := []Button{
		{X: 10, Y: 10, W: 100, H: 30, ID: "new"},
		{X: 10, Y: 50, W: 100, H: 30, ID: "quit"},
	}
	// Hit 邊界
	if !btns[0].Hit(10, 10) || !btns[0].Hit(109, 39) || btns[0].Hit(110, 40) {
		t.Error("Hit 邊界判定錯誤")
	}
	// 點擊(放開)命中 quit
	if got := ClickedButton(btns, InputState{MouseX: 50, MouseY: 60, ClickReleased: true}); got != "quit" {
		t.Errorf("ClickedButton = %q,預期 quit", got)
	}
	// 沒放開(ClickReleased=false)→ 不觸發
	if got := ClickedButton(btns, InputState{MouseX: 50, MouseY: 60, ClickReleased: false}); got != "" {
		t.Errorf("未放開不應觸發,得 %q", got)
	}
	// 點空白處
	if got := ClickedButton(btns, InputState{MouseX: 500, MouseY: 500, ClickReleased: true}); got != "" {
		t.Errorf("點空白應回空,得 %q", got)
	}
	// Hover
	if got := HoveredButton(btns, InputState{MouseX: 50, MouseY: 20}); got != "new" {
		t.Errorf("HoveredButton = %q,預期 new", got)
	}
}
