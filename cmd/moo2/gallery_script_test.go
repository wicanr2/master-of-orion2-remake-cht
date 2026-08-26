package main

import "testing"

func TestGalleryScriptResolvesInitialResearchBeforeWorldTurn(t *testing.T) {
	script, shots := buildGalleryScript()
	if len(script) < 14 {
		t.Fatalf("畫廊腳本太短：%d", len(script))
	}
	firstTurn, chooseApplication, worldTurn := script[11], script[12], script[13]
	if !firstTurn.ClickReleased || firstTurn.MouseX != 589 || firstTurn.MouseY != 458 {
		t.Fatalf("t12 必須先按 TURN，得到 %+v", firstTurn)
	}
	if !chooseApplication.ClickReleased || chooseApplication.MouseX != 320 || chooseApplication.MouseY != 173 {
		t.Fatalf("t13 必須解除新局研究 application gate，得到 %+v", chooseApplication)
	}
	if !worldTurn.ClickReleased || worldTurn.MouseX != 589 || worldTurn.MouseY != 458 {
		t.Fatalf("t14 必須再次按 TURN 才會真正結算世界，得到 %+v", worldTurn)
	}
	foundEvent, foundSummary := false, false
	for _, shot := range shots {
		if shot.name == "05_event.png" {
			foundEvent = shot.tick == galleryEventTick
		}
		if shot.name == "06_turnsummary.png" {
			foundSummary = shot.tick == 14
		}
	}
	if !foundEvent {
		t.Fatal("事件快報須由明示的畫廊事件 tick 截圖")
	}
	if !foundSummary {
		t.Fatal("正常第一個世界回合須在 t14 截取回合摘要")
	}
}
