package main

import (
	"testing"
	"time"
)

func TestPromoDemoStepsFollowNormalPlayableRoute(t *testing.T) {
	steps := buildPromoDemoSteps()
	if len(steps) != 16 {
		t.Fatalf("導覽步數 = %d, want 16", len(steps))
	}
	if first := steps[0]; first.input.ClickReleased || first.hold != 2*time.Second {
		t.Fatalf("第一步 = %+v, want 2 秒主選單停留", first)
	}

	want := []struct{ x, y int }{
		{491, 228}, // NEW GAME
		{486, 405}, // 設定 ACCEPT
		{410, 350}, // 人類種族
		{540, 454}, // 命名／旗色 ACCEPT
		{48, 452},  // COLONIES
		{50, 47},   // 第一個殖民地
		{590, 459}, // 殖民地 RETURN
		{608, 462}, // 殖民地總覽 RETURN
		{495, 452}, // INFO
		{21, 80},   // 科技頁
		{535, 434}, // INFO RETURN
		{420, 452}, // RACES
		{483, 428}, // REPORT
		{320, 437}, // 結束對談
		{388, 448}, // DECLARE WAR
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
		t.Fatalf("導覽總停留時間 = %s, want 1m0s", total)
	}
}
