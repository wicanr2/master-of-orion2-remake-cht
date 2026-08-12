package main

import (
	"testing"
	"time"
)

func TestPromoDemoStepsFollowNormalPlayableRoute(t *testing.T) {
	steps := buildPromoDemoSteps()
	if len(steps) != 23 {
		t.Fatalf("導覽步數 = %d, want 23", len(steps))
	}
	if first := steps[0]; first.input.ClickReleased || first.hold != 2*time.Second {
		t.Fatalf("第一步 = %+v, want 2 秒主選單停留", first)
	}
	if accept := steps[2]; accept.cursorHidden != 1500*time.Millisecond {
		t.Fatalf("新局 ACCEPT 的游標隱藏時間 = %s, want 1.5s", accept.cursorHidden)
	}

	want := []struct{ x, y int }{
		{491, 228}, // NEW GAME
		{526, 398}, // 設定 ACCEPT：按鈕右側空白處
		{464, 370}, // 人類種族：右下留白，游標移往下一頁不橫越鄰鈕文字
		{578, 448}, // 命名／旗色 ACCEPT：按鈕右側空白處
		{48, 452},  // COLONIES
		{50, 47},   // 第一個殖民地
		{410, 107}, // 農夫 → 工人
		{590, 459}, // 殖民地 RETURN
		{608, 462}, // 殖民地總覽 RETURN
		{420, 452}, // RACES
		{130, 135}, // RACES：增派間諜
		{204, 135}, // RACES：循環任務
		{276, 135}, // RACES：隱匿
		{483, 428}, // REPORT
		{320, 437}, // 結束對談
		{388, 448}, // DECLARE WAR
		{215, 97},  // 第一艘艦移動
		{495, 97},  // 第一艘敵艦
		{145, 152}, // 第二艘我方艦
		{215, 152}, // 第二艘艦移動
		{495, 152}, // 第二艘敵艦
		{300, 374}, // AUTO
	}
	var total time.Duration
	for i, step := range steps {
		total += step.hold
		if i == 0 {
			continue
		}
		got := step.input
		if !got.ClickReleased || got.MouseX != want[i-1].x || got.MouseY != want[i-1].y {
			t.Fatalf("第 %d 個操作 = %+v, want click(%d,%d)", i, got, want[i-1].x, want[i-1].y)
		}
	}
	if total != 60*time.Second {
		t.Fatalf("導覽總停留時間 = %s, want 60s", total)
	}
}

func TestPromoCursorInterpolatesDuringTheHold(t *testing.T) {
	start := time.Unix(1_700_000_000, 0)
	a := interactiveApp{
		promoCursorReady:  true,
		promoCursorX:      100,
		promoCursorY:      80,
		promoCursorFromX:  100,
		promoCursorFromY:  80,
		promoCursorToX:    200,
		promoCursorToY:    120,
		promoCursorMoveAt: start,
		promoCursorUntil:  start.Add(2 * time.Second),
	}
	a.updatePromoCursor(start.Add(time.Second))
	if a.promoCursorX != 150 || a.promoCursorY != 100 {
		t.Fatalf("游標半途位置 = (%.1f,%.1f), want (150,100)", a.promoCursorX, a.promoCursorY)
	}
	a.updatePromoCursor(start.Add(3 * time.Second))
	if a.promoCursorX != 200 || a.promoCursorY != 120 {
		t.Fatalf("游標終點 = (%.1f,%.1f), want (200,120)", a.promoCursorX, a.promoCursorY)
	}
}

func TestTacticalMessageBandDoesNotOverlapTheGridOrControlDeck(t *testing.T) {
	if tacticalMessageY+tacticalMessageH > gcY0 {
		t.Fatalf("訊息帶底部 %d 跨入格線起點 %d", tacticalMessageY+tacticalMessageH, gcY0)
	}
	if tacticalMessageY+tacticalMessageH > combatControlDeckY {
		t.Fatalf("訊息帶底部 %d 跨入控制列 %d", tacticalMessageY+tacticalMessageH, combatControlDeckY)
	}
}
